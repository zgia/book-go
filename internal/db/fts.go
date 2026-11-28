package db

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"

	"xorm.io/xorm"
	xormlog "xorm.io/xorm/log"

	"zgia.net/book/internal/conf"
	log "zgia.net/book/internal/logger"
)

// 中文没有词边界，FTS5 内置的 unicode61 会把一整句话当成一个 token，
// 导致 “张三” 搜不到 “张三丰”。这里的做法是：入库时把正文切成相邻两字的
// 窗口（2-gram）用空格连接，检索时把查询词同样切分，用 AND 组合 ——
// 语义等价于 LIKE '%查询词%'，但走的是倒排索引。
const (
	// ftsTableName 是正文 2-gram 索引表。
	ftsTableName = "content_fts"

	// createFTSTableSQL 建索引表：
	//   content=''   —— contentless，索引不保存正文副本（正文仍在 content 表），
	//                   950MB 的正文不会因此翻倍；
	//   detail=none  —— 不存 token 位置。2-gram 本身已表达“相邻”，查询只需
	//                   gram 是否存在，不需要短语匹配，省下大量位置信息；
	//   unicode61    —— gram 之间用空格分隔，按空格切词即可，且自带英文大小写
	//                   折叠，与 SQLite LIKE 的 ASCII 不敏感行为一致。
	createFTSTableSQL = `CREATE VIRTUAL TABLE ` + ftsTableName +
		` USING fts5(g, content='', detail=none, tokenize='unicode61')`

	// ftsRebuildBatch 全量重建时每个事务处理的行数。
	ftsRebuildBatch = 500
	// ftsCandidateLimit FTS 命中后回表复核的候选上限。
	ftsCandidateLimit = 5000
)

var (
	ftsReady atomic.Bool
	// ftsLogMu 串行化 “临时压制 SQL 日志” 这一段，避免并发写章节时
	// 级别来回切换互相踩踏。
	ftsLogMu sync.Mutex
)

// withQuietSQL 临时把 xorm 日志压到 WARN：索引 SQL 的参数是一整章
// 2-gram 文本，按 INFO 打印会把日志撑爆（一次 rebuild 能写出几百兆）。
func withQuietSQL(fn func() error) error {
	ftsLogMu.Lock()
	defer ftsLogMu.Unlock()

	x.SetLogLevel(xormlog.LOG_WARNING)
	defer x.SetLogLevel(logLevel())

	return fn()
}

// IsFTSReady 报告当前进程是否可以使用 FTS 索引检索。
func IsFTSReady() bool {
	return ftsReady.Load()
}

// InitFTS 在连库后探测索引是否可用，并把状态写进日志。
// 表不存在、或二进制没编译 FTS5 时，搜索会自动回退到 LIKE。
func InitFTS() {
	ftsReady.Store(false)
	if !conf.UseSQLite {
		return
	}

	if _, err := x.SQL("SELECT rowid FROM " + ftsTableName + " LIMIT 1").QueryString(); err != nil {
		log.Infof("FTS 索引不可用（%v）；正文搜索回退 LIKE，需要时执行 `book fts rebuild`", err)
		return
	}

	ftsReady.Store(true)
	n, _ := FTSCount()
	log.Infof("FTS 索引已就绪：%s，已索引 %d 章", ftsTableName, n)
}

// FTSCount 返回索引里已收录的章节数。
func FTSCount() (int64, error) {
	rows, err := x.SQL("SELECT count(*) AS n FROM " + ftsTableName).QueryString()
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return strconv.ParseInt(rows[0]["n"], 10, 64)
}

// CountContents 返回正文总行数，用于比对索引覆盖率。
func CountContents() (int64, error) {
	rows, err := x.SQL("SELECT count(*) AS n FROM content").QueryString()
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return strconv.ParseInt(rows[0]["n"], 10, 64)
}

// GramText 把正文切成相邻两字窗口，用空格连接，供 FTS5 建索引。
// 只保留字母/数字，标点与空白不参与窗口，因此 “张，三” 不会产生 “张三”
// 这个 gram —— 与 LIKE '%张三%' 的语义保持一致。
func GramText(s string) string {
	r := []rune(s)
	if len(r) < 2 {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s) * 2)
	for i := 0; i+1 < len(r); i++ {
		if !unicode.IsLetter(r[i]) || !unicode.IsLetter(r[i+1]) {
			continue
		}
		b.WriteRune(r[i])
		b.WriteRune(r[i+1])
		b.WriteByte(' ')
	}

	return b.String()
}

// MatchExpr 把查询词切成 2-gram 并构造 FTS5 的 MATCH 表达式。
// detail=none 的索引不支持 phrase 查询，所以这里输出裸 token（空格即 AND），
// 不能加双引号。gram 只含字母/数字，不含 FTS5 查询语法的特殊字符。
// 单字查询返回空串（无法用 2-gram 表达）。
func MatchExpr(words string) string {
	r := []rune(words)
	if len(r) < 2 {
		return ""
	}

	grams := make([]string, 0, len(r)-1)
	for i := 0; i+1 < len(r); i++ {
		if !unicode.IsLetter(r[i]) || !unicode.IsLetter(r[i+1]) {
			return ""
		}
		grams = append(grams, string(r[i])+string(r[i+1]))
	}

	return strings.Join(grams, " ")
}

// canSearchByFTS 判断一次搜索是否适合走 2-gram 索引。
// 单字查询，或含标点/空白的查询词都会回退 LIKE：2-gram 无法表达
// “这些字必须连续出现” 之外的语义，强行走索引会给出错误结果。
func canSearchByFTS(words string) bool {
	if !conf.UseSQLite || !ftsReady.Load() {
		return false
	}

	r := []rune(words)
	if len(r) < 2 {
		return false
	}
	for _, c := range r {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			return false
		}
	}

	return true
}

// RebuildFTS 全量重建正文索引，耗时随正文总量线性增长。
// progress 为 nil 时不回调；batch <= 0 时使用默认批量。
func RebuildFTS(batchSize int, progress func(done, total int64)) (int64, error) {
	if !conf.UseSQLite {
		return 0, fmt.Errorf("FTS 索引只支持 SQLite")
	}
	if batchSize <= 0 {
		batchSize = ftsRebuildBatch
	}

	// rebuild 会执行数万条带长参数的 INSERT，期间压掉 INFO 级 SQL 日志。
	x.SetLogLevel(xormlog.LOG_WARNING)
	defer x.SetLogLevel(logLevel())

	if _, err := x.Exec("DROP TABLE IF EXISTS " + ftsTableName); err != nil {
		return 0, fmt.Errorf("删除旧索引表: %w", err)
	}
	if _, err := x.Exec(createFTSTableSQL); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such module") ||
			strings.Contains(strings.ToLower(err.Error()), "fts5") {
			return 0, fmt.Errorf("当前二进制未编译 FTS5，请用 `go build -tags sqlite_fts5` 重新构建: %w", err)
		}
		return 0, fmt.Errorf("创建索引表: %w", err)
	}

	total, _ := CountContents()

	var lastID, done int64
	for {
		rows, err := x.SQL(
			"SELECT chapter_id, txt FROM content WHERE chapter_id > ? ORDER BY chapter_id LIMIT ?",
			lastID, batchSize,
		).QueryString()
		if err != nil {
			return done, fmt.Errorf("读取正文（chapter_id > %d）: %w", lastID, err)
		}
		if len(rows) == 0 {
			break
		}

		sess := x.NewSession()
		if err := sess.Begin(); err != nil {
			sess.Close()
			return done, err
		}
		for _, r := range rows {
			id, perr := strconv.ParseInt(r["chapter_id"], 10, 64)
			if perr != nil {
				continue
			}
			lastID = id

			g := GramText(r["txt"])
			if g == "" {
				continue
			}
			if _, err := sess.Exec("INSERT INTO "+ftsTableName+"(rowid, g) VALUES(?, ?)", id, g); err != nil {
				sess.Rollback()
				sess.Close()
				return done, fmt.Errorf("索引第 %d 章: %w", id, err)
			}
		}
		if err := sess.Commit(); err != nil {
			sess.Rollback()
			sess.Close()
			return done, err
		}
		sess.Close()

		done += int64(len(rows))
		if progress != nil {
			progress(done, total)
		}
	}

	// 合并段：contentless 表同样支持 optimize，能显著压缩索引并加快查询。
	if _, err := x.Exec("INSERT INTO " + ftsTableName + "(" + ftsTableName + ") VALUES('optimize')"); err != nil {
		log.Warnf("FTS optimize 失败（不影响查询）: %v", err)
	}

	ftsReady.Store(true)
	return done, nil
}

// ftsUpsert 在调用方事务内写入一章的 2-gram 索引。
// 索引属于加速层，失败时只告警、不让正文章节保存失败。
func ftsUpsert(sess *xorm.Session, chapterid int64, txt string) error {
	if !ftsReady.Load() {
		return nil
	}
	g := GramText(txt)
	if g == "" {
		return nil
	}
	return withQuietSQL(func() error {
		_, err := sess.Exec("INSERT INTO "+ftsTableName+"(rowid, g) VALUES(?, ?)", chapterid, g)
		return err
	})
}

// ftsRemove 删除一章的索引。contentless 表不带原文，删除时必须
// 提供当初写进去的 gram 文本，所以调用方要先把旧正文读出来。
func ftsRemove(sess *xorm.Session, chapterid int64, oldTxt string) error {
	if !ftsReady.Load() {
		return nil
	}
	g := GramText(oldTxt)
	if g == "" {
		return nil
	}
	return withQuietSQL(func() error {
		_, err := sess.Exec(
			"INSERT INTO "+ftsTableName+"("+ftsTableName+", rowid, g) VALUES('delete', ?, ?)",
			chapterid, g,
		)
		return err
	})
}

// FTSGap 描述一个还没进索引的章节。
type FTSGap struct {
	ID int64
	// Indexable 为 false 表示正文为空、或只有标点/单字，2-gram 无从生成，
	// 重建索引也不会收录。这类“差异”是正常的，不是索引没建好。
	Indexable bool
}

// FTSGaps 返回最多 limit 个还没进索引的章节，并标出哪些本身没有任何可索引内容。
// limit 是兜底：索引整体缺失时，避免把 95 万行正文一次性拉进内存。
func FTSGaps(limit int) ([]FTSGap, error) {
	if limit <= 0 {
		limit = 1000
	}

	rows, err := x.SQL(
		"SELECT chapter_id, txt FROM content WHERE chapter_id NOT IN "+
			"(SELECT rowid FROM "+ftsTableName+") ORDER BY chapter_id LIMIT ?",
		limit,
	).QueryString()
	if err != nil {
		return nil, err
	}

	gaps := make([]FTSGap, 0, len(rows))
	for _, r := range rows {
		id, perr := strconv.ParseInt(r["chapter_id"], 10, 64)
		if perr != nil {
			continue
		}
		gaps = append(gaps, FTSGap{ID: id, Indexable: GramText(r["txt"]) != ""})
	}

	return gaps, nil
}
