package db

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"xorm.io/builder"
	"xorm.io/xorm"

	"zgia.net/book/internal/conf"
	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/util"
)

type Book struct {
	Id               int64 `xorm:"pk autoincr"`
	CategoryID       int64 `xorm:"notnull default 0"`
	Title            string
	Alias            string
	Author           string `xorm:"<-"`
	AuthorID         int64
	AuthorFormerName string `xorm:"<-"`
	Summary          string
	Source           string
	Cover            string
	Latest           string
	Rate             int64     `xorm:"notnull default 0"`
	WordCount        int64     `xorm:"notnull default 0"`
	IsFinished       int64     `xorm:"notnull default 0"`
	CreatedAt        time.Time `xorm:"created notnull"`
	UpdatedAt        time.Time `xorm:"updated notnull"`
	DeletedAt        time.Time `xorm:"deleted null"`
}

func (b *Book) String() string {
	return fmt.Sprintf("Book: %s(%s-%d), %s", b.Title, b.Author, b.AuthorID, b.Summary)
}

func condition(session *xorm.Session, words, searchMode, rate string) {

	if words != "" {
		if searchMode == "categ" {
			session.Where("book.category_id=?", words)
		} else if searchMode == "author" {
			session.Where("au.name LIKE ? OR au.former_name LIKE ?", "%"+words+"%", "%"+words+"%")
		} else {
			session.Where("title LIKE ? OR alias LIKE ?", "%"+words+"%", "%"+words+"%")
		}
	}

	rates := util.ParseRates(rate)
	if len(rates) > 0 {
		session.In("rate", rates)
	}
}

// CountBooks returns number of books.
func CountBooks(words, searchMode, rate string) (int64, error) {
	session := x.Table("book").Join("LEFT", "author AS au", "book.author_id = au.id")
	condition(session, words, searchMode, rate)

	return session.Count(new(Book))

}

// ListBooks returns number of books in given page.
func QueryBooks(page int, words, searchMode, orderby, direction, rate string) ([]*Book, error) {
	// Validate page number
	if page < 1 {
		page = 1
	}
	pageSize := conf.PageSize(0)
	books := make([]*Book, 0, pageSize)

	session := x.Table("book").Join("LEFT", "author AS au", "book.author_id = au.id").Select("book.*,au.name AS author,au.former_name AS author_former_name")
	condition(session, words, searchMode, rate)

	// 排序
	orders := map[string]int{"id": 1, "word_count": 1, "updated_at": 1, "rate": 1}
	if _, ok := orders[orderby]; !ok {
		orderby = "updated_at"
	}
	orderby = fmt.Sprintf("book.%s", orderby)

	if direction != "ascending" {
		session.Desc(orderby)
	} else {
		session.Asc(orderby)
	}

	return books, session.Limit(pageSize, (page-1)*pageSize).Find(&books)
}

// QueryBooksByAuthorIds returns books by author IDs.
func QueryBooksByAuthorIds(authorIds []int64) ([]*Book, error) {
	books := make([]*Book, 0, len(authorIds))

	return books, x.Table("book").Select("id, title,author_id").In("author_id", authorIds).Find(&books)
}

func QueryAllBookIds() ([]int64, error) {
	var ids []int64
	err := x.Table("book").Cols("id").Find(&ids)

	return ids, err
}

type categoryCount struct {
	Total      int64
	CategoryID int64
}

// QueryCountsByCategory returns books count group by categoryid.
func QueryCountsByCategory() (map[int64]int64, error) {

	res := []categoryCount{}

	err := x.Table("book").Select("category_id, COUNT(*) AS total").Where("deleted_at IS NULL").GroupBy("category_id").Find(&res)

	count := map[int64]int64{}
	for _, v := range res {
		count[v.CategoryID] = v.Total
	}

	return count, err
}

// QueryLatestChapters returns latest chapters of given books.
func QueryLatestChapters(bookids []int64) (map[int64]string, error) {
	var results []map[string]string

	subQuery := builder.Select("MAX(id) AS id").
		From("chapter").
		Where(builder.In("book_id", bookids)).
		GroupBy("book_id")

	sql, args, err := builder.Select("book_id, title").
		From("chapter").
		Where(builder.In("id", subQuery)).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("cannot build SQL for latest chapters: %w", err)
	}

	if err := x.SQL(sql, args...).Find(&results); err != nil {
		return nil, fmt.Errorf("cannot query books latest chapters, sql: %s, error: %w", sql, err)
	}

	chaps := make(map[int64]string, len(results))
	for _, v := range results {
		chaps[util.ParamInt64(v["book_id"])] = v["title"]
	}

	return chaps, nil
}

func QueryVolumeChapters(bookid, volumeid int64) ([]map[string]string, error) {
	volume_chapters_sql := "SELECT txt AS content,cha.title " +
		"FROM chapter AS cha " +
		"INNER JOIN content AS content ON cha.id=content.chapter_id " +
		"WHERE cha.book_id=? AND cha.volume_id=? AND cha.deleted_at IS NULL ORDER BY cha.id"

	var results []map[string]string

	if err := x.SQL(volume_chapters_sql, bookid, volumeid).Find(&results); err != nil {
		return nil, fmt.Errorf("cannot query book(%d)-vol(%d) chapters: %w", bookid, volumeid, err)
	}

	return results, nil
}

func QueryAllChapters(bookid int64) ([]map[string]string, error) {
	chapters_sql := "SELECT txt AS content,cha.title,vol.title AS volume_title " +
		"FROM chapter AS cha " +
		"INNER JOIN content AS content ON cha.id=content.chapter_id " +
		"INNER JOIN volume AS vol ON vol.id=cha.volume_id " +
		"WHERE cha.book_id=? AND cha.deleted_at IS NULL ORDER BY cha.id"

	var results []map[string]string

	if err := x.SQL(chapters_sql, bookid).Find(&results); err != nil {
		return nil, fmt.Errorf("cannot query book(%d) chapters: %w", bookid, err)
	}

	return results, nil
}

func QueryBooksByKeywords(words string, bookid int64) ([]map[string]string, error) {
	// 有 2-gram 索引且查询词适用时走倒排，否则保持原来的全表扫描。
	// 索引是加速层，出错时静默退回 LIKE，保证搜索始终可用。
	if canSearchByFTS(words) {
		results, err := queryBooksByFTS(words, bookid)
		if err == nil {
			return results, nil
		}
		log.Warnf("FTS 搜索失败，回退 LIKE：%v", err)
	}

	return queryBooksByLike(words, bookid)
}

// queryBooksByFTS 用 2-gram 索引先定位章节，再回表复核并取正文。
// 子查询里的 ORDER BY rowid 是为了保持 “按 id 顺序取前 N” 的既有语义，
// 避免索引返回顺序随意导致结果抖动。
func queryBooksByFTS(words string, bookid int64) ([]map[string]string, error) {
	match := MatchExpr(words)
	if match == "" {
		return nil, fmt.Errorf("无法为 %q 构造 FTS 表达式", words)
	}

	search_books_sql := "SELECT content.txt AS content,cha.id AS chapter_id,cha.title AS chapter_title,vol.title AS volume_title " +
		",volume_id AS volume_id,book.id AS book_id,book.title AS book_title,au.name AS author " +
		"FROM chapter AS cha " +
		"INNER JOIN content AS content ON cha.id=content.chapter_id " +
		"INNER JOIN volume AS vol ON vol.id=cha.volume_id " +
		"INNER JOIN book AS book ON cha.book_id=book.id " +
		"LEFT JOIN author AS au ON au.id=book.author_id " +
		"WHERE cha.id IN (SELECT rowid FROM " + ftsTableName + " WHERE " + ftsTableName + " MATCH ? " +
		"ORDER BY rowid LIMIT " + strconv.Itoa(ftsCandidateLimit) + ") " +
		"AND _BOOKID_CONDITION_ content.txt LIKE ? AND cha.deleted_at IS NULL ORDER BY cha.id LIMIT 100"

	var results []map[string]string

	var bookidCond = ""
	args := []any{match}
	if bookid != 0 {
		bookidCond = " cha.book_id = ? AND "
		args = append(args, bookid)
	}
	args = append(args, "%"+words+"%")
	search_books_sql = strings.Replace(search_books_sql, "_BOOKID_CONDITION_", bookidCond, 1)

	if err := x.SQL(search_books_sql, args...).Find(&results); err != nil {
		return nil, fmt.Errorf("cannot query book(%d) by keywords(%s) with fts: %w", bookid, words, err)
	}

	return results, nil
}

// queryBooksByLike 是未建索引（或查询词不适用）时的兜底路径。
func queryBooksByLike(words string, bookid int64) ([]map[string]string, error) {
	search_books_sql := "SELECT txt AS content,cha.id AS chapter_id,cha.title AS chapter_title,vol.title AS volume_title " +
		",volume_id AS volume_id,book.id AS book_id,book.title AS book_title,au.name AS author " +
		"FROM chapter AS cha " +
		"INNER JOIN content AS content ON cha.id=content.chapter_id " +
		"INNER JOIN volume AS vol ON vol.id=cha.volume_id " +
		"INNER JOIN book AS book ON cha.book_id=book.id " +
		"LEFT JOIN author AS au ON au.id=book.author_id " +
		"WHERE _BOOKID_CONDITION_ content.txt LIKE ? AND cha.deleted_at IS NULL ORDER BY cha.id LIMIT 100"

	var results []map[string]string

	var bookidCond = ""
	var args []any
	if bookid != 0 {
		bookidCond = " cha.book_id = ? AND "
		args = append(args, bookid)
	}
	args = append(args, "%"+words+"%")
	search_books_sql = strings.Replace(search_books_sql, "_BOOKID_CONDITION_", bookidCond, 1)

	if err := x.SQL(search_books_sql, args...).Find(&results); err != nil {
		return nil, fmt.Errorf("cannot query book(%d) by keywords(%s) chapters: %w", bookid, words, err)
	}

	return results, nil
}

func QueryBooksSize() (map[string]string, error) {
	if conf.UseMySQL {
		sql := "SELECT CEILING(sum(data_length)/1024/1024) AS data_size, CEILING(sum(index_length)/1024/1024) AS index_size FROM information_schema.tables WHERE TABLE_SCHEMA = ?"
		var results []map[string]string
		if err := x.SQL(sql, conf.Database.Name).Find(&results); err != nil {
			return nil, fmt.Errorf("cannot query books size: %w", err)
		}
		if len(results) == 0 {
			return map[string]string{"data_size": "0", "index_size": "0"}, nil
		}
		return results[0], nil
	} else if conf.UseSQLite {
		// Get page count and page size via PRAGMA
		var pageSize, pageCount int64
		// Query page_size
		var sizeResults []map[string]string
		if err := x.SQL("PRAGMA page_size;").Find(&sizeResults); err != nil {
			return nil, fmt.Errorf("cannot query SQLite page_size: %w", err)
		}
		if len(sizeResults) > 0 {
			for _, v := range sizeResults[0] {
				pageSize, _ = strconv.ParseInt(v, 10, 64)
				break
			}
		}
		// Query page_count
		var countResults []map[string]string
		if err := x.SQL("PRAGMA page_count;").Find(&countResults); err != nil {
			return nil, fmt.Errorf("cannot query SQLite page_count: %w", err)
		}
		if len(countResults) > 0 {
			for _, v := range countResults[0] {
				pageCount, _ = strconv.ParseInt(v, 10, 64)
				break
			}
		}
		// Calculate total size in bytes
		totalBytes := pageSize * pageCount
		// Convert to MB (ceiling)
		totalMB := (totalBytes + 1024*1024 - 1) / (1024 * 1024)
		if totalMB == 0 {
			totalMB = 1 // at least 1 MB
		}
		return map[string]string{
			"data_size":  strconv.FormatInt(totalMB, 10),
			"index_size": "0",
		}, nil
	}
	return nil, fmt.Errorf("unknown database type")
}

// QueryBook gets a book info
func QueryBook(bookid int64) (*Book, error) {
	book := &Book{
		Id: bookid,
	}
	has, err := x.Table("book").Join("LEFT", "author AS au", "book.author_id = au.id").Select("book.*,au.name AS author,au.former_name AS author_former_name").Get(book)

	if err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return book, nil
}

// UpdateBook updates/creates a book
func UpdateBook(book *Book, bookid int64) (int64, error) {
	var err error

	// 处理作者
	book.AuthorID, err = CheckAuthor(book.Author, "")
	if err != nil {
		return 0, fmt.Errorf("check author %q: %w", book.Author, err)
	}
	log.Infof("%v", book)

	sess := x.NewSession()
	defer sess.Close()

	if err = sess.Begin(); err != nil {
		return 0, err
	}

	if bookid == 0 {
		_, err = sess.Insert(book)
		if err != nil {
			return 0, err
		}

		volume := &Volume{
			Title:  book.Title,
			BookID: book.Id,
		}
		_, err = sess.Insert(volume)
		if err != nil {
			return 0, err
		}

		bookid = book.Id
	} else {
		if _, err = sess.ID(bookid).Cols("category_id", "title", "author_id", "alias", "summary", "source", "word_count", "is_finished", "rate").Update(book); err != nil {
			return 0, err
		}
	}

	if err = sess.Commit(); err != nil {
		return 0, err
	}

	return bookid, nil

}

// DeleteBook deletes a book
func DeleteBook(bookid int64) (int64, error) {
	return x.Delete(Book{Id: bookid})
}
