package db

import (
	"fmt"
	"strings"
	"time"

	// https://go.dev/ref/spec#Import_declarations
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/core"
	"xorm.io/xorm"
	xormlog "xorm.io/xorm/log"

	"zgia.net/book/internal/conf"
	log "zgia.net/book/internal/logger"
)

var (
	x *xorm.Engine
)

const (
	// SQLite 默认连接数。WAL 下读可并发、写串行，4 是经验上的甜点值。
	sqliteDefaultMaxConns = 4
	// SQLite 连接数上限：再多只会加剧写锁竞争，收益为负。
	sqliteMaxConnsLimit = 8
)

func init() {
	gonicNames := []string{"SSL"}
	for _, name := range gonicNames {
		core.LintGonicMapper[name] = true
	}
}

// 如果出现错误：commands out of sync. Did you run multiple statements at once
// 可将 Host 改成 socket 的链接方式: Host = /usr/local/mysql/mysql.socket
// SHOW VARIABLES LIKE "socket"
func getEngine() (*xorm.Engine, error) {
	Param := "?"
	if strings.Contains(conf.Database.Name, Param) {
		Param = "&"
	}

	// MySQL 会话时区与 [time] ZONE 保持一致，否则 DEFAULT CURRENT_TIMESTAMP / NOW()
	// 会按服务器时区写入，与 xorm 写入的本地时间相差一个时区偏移。
	// 形如 time_zone=%27%2B08%3A00%27，驱动解码后执行 SET time_zone='+08:00'。
	mysqlTZ := ""
	{
		_, offset := time.Now().Zone()
		// go-sql-driver 会对 DSN 参数值做一次 URL 解码，'+' 会被解成空格，
		// 因此正偏移必须写成 %2B，否则 SET time_zone 收到的是 " 08:00"。
		sign := "%2B"
		if offset < 0 {
			sign = "-"
			offset = -offset
		}
		mysqlTZ = fmt.Sprintf("&time_zone=%%27%s%02d%%3A%02d%%27", sign, offset/3600, (offset%3600)/60)
	}

	connStr := ""
	switch conf.Database.Type {
	case "mysql":
		conf.UseMySQL = true
		conf.UseSQLite = false
		// loc=Local: 让驱动按本地时区（[time] ZONE）解析 DATETIME，
		// 否则默认按 UTC 解析，写入东八区后读回会差 8 小时。
		if conf.Database.Host[0] == '/' { // looks like a unix socket
			connStr = fmt.Sprintf("%s:%s@unix(%s)/%s%scharset=utf8mb4&parseTime=true&loc=Local%s",
				conf.Database.User, conf.Database.Password, conf.Database.Host, conf.Database.Name, Param, mysqlTZ)
		} else {
			connStr = fmt.Sprintf("%s:%s@tcp(%s)/%s%scharset=utf8mb4&parseTime=true&loc=Local%s",
				conf.Database.User, conf.Database.Password, conf.Database.Host, conf.Database.Name, Param, mysqlTZ)
		}
		var engineParams = map[string]string{"rowFormat": "DYNAMIC"}

		return xorm.NewEngineWithParams(conf.Database.Type, connStr, engineParams)

	case "sqlite", "sqlite3":
		conf.UseMySQL = false
		conf.UseSQLite = true
		// For SQLite, Host field contains the file path
		sqlitePath := conf.Database.Host
		if sqlitePath == "" {
			sqlitePath = conf.Database.Name // fallback to Name field
		}
		// 必须是 file: 形式，mattn/go-sqlite3 只在 URI 模式下才会把
		// 非下划线开头的参数（如 mode=rwc）当 SQLite 自身的 URI 参数处理。
		if !strings.HasPrefix(sqlitePath, "file:") {
			sqlitePath = "file:" + sqlitePath
		}
		sep := "?"
		if strings.Contains(sqlitePath, "?") {
			sep = "&"
		}
		// 调用方在 HOST 里显式配置过的同名参数不覆盖，避免把只读挂载的
		// mode=ro 之类强行改写成 rwc。
		hasParam := func(key string) bool {
			return strings.Contains(sqlitePath, "?"+key+"=") || strings.Contains(sqlitePath, "&"+key+"=")
		}
		// WAL：读不阻塞写、写不阻塞读，这是几百 MB 级单文件库可用性的前提。
		// 驱动在 journal_mode=wal 时会自动把 synchronous 调成 NORMAL，
		// 这里仍显式写出，避免将来换驱动或调参时语义悄悄漂移。
		// busy_timeout：写锁冲突时等待 5 秒再报 SQLITE_BUSY，而不是立刻失败。
		// foreign_keys：开启外键约束。mode=rwc：库文件不存在时自动创建。
		// 不再使用 cache=shared：共享缓存在多连接下会放大锁竞争，官方自
		// SQLite 3.7.0 起推荐 WAL + 每连接独立页缓存。
		params := []struct{ key, value string }{
			{"mode", "rwc"},
			{"_journal_mode", "wal"},
			{"_busy_timeout", "5000"},
			{"_synchronous", "normal"},
			{"_foreign_keys", "1"},
		}
		for _, p := range params {
			if !hasParam(p.key) {
				sqlitePath += sep + p.key + "=" + p.value
				sep = "&"
			}
		}
		return xorm.NewEngine("sqlite3", sqlitePath)

	default:
		return nil, fmt.Errorf("unknown database type: %s", conf.Database.Type)
	}
}

func NewEngine() (err error) {

	x, err = getEngine()
	if err != nil {
		return fmt.Errorf("connect to database: %v", err)
	}

	x.SetMapper(core.GonicMapper{})

	// 统一数据库读写时区为 [time] ZONE（conf.Init 已把它赋给 time.Local）。
	// xorm 对 SQLite 默认 DatabaseTZ = UTC，不设置的话 DATETIME 会按 UTC 写入/读回，
	// 与东八区的时间差 8 小时。
	x.SetTZDatabase(time.Local)
	x.SetTZLocation(time.Local)

	// Adjust connection pool settings based on database type
	if conf.UseSQLite {
		// WAL 模式下读连接可以并发，写由 busy_timeout 排队。
		// 池子取 [database] MAX_OPEN_CONNS，但夹在 1..8 之间：
		// 1 个连接会让一次长查询（例如正文 LIKE 扫描）堵住整个服务，
		// 连接数过高则加剧写锁竞争，反而拉高尾部延迟。
		conns := conf.Database.MaxOpenConns
		if conns <= 0 {
			conns = sqliteDefaultMaxConns
		}
		if conns > sqliteMaxConnsLimit {
			conns = sqliteMaxConnsLimit
		}
		x.SetMaxOpenConns(conns)
		x.SetMaxIdleConns(conns)
		// 连接长期存活有利于复用页缓存；WAL 的检查点由写入连接自动完成。
		x.SetConnMaxLifetime(time.Hour)

		// DSN 里的 PRAGMA 在文件系统不支持 WAL 时会被静默降级，
		// 启动时回读一次并落日志，避免事后靠猜。
		logSQLitePragmas(conns)
	} else {
		// MySQL connection pool settings
		x.SetMaxOpenConns(conf.Database.MaxOpenConns)
		x.SetMaxIdleConns(conf.Database.MaxIdleConns)
		x.SetConnMaxLifetime(time.Second)
	}

	zapLogger := &log.ZapXormLogger{}
	zapLogger.SetLevel(logLevel())
	zapLogger.ShowSQL(true)
	x.SetLogger(zapLogger)

	return nil
}

func Ping() error {
	return x.Ping()
}

func logLevel() xormlog.LogLevel {
	zapLevel := log.Log.Level
	if sec, err := conf.File.GetSection("log.xorm"); err == nil {
		zapLevel = log.GetLevel(sec.Key("LEVEL").MustString(zapLevel.String()))
	}

	return log.ZapLevelToXorm(zapLevel)
}

// sqlitePragma 读取单值 PRAGMA 的首行首列，读不到时返回 false。
func sqlitePragma(name string) (string, bool) {
	rows, err := x.SQL("PRAGMA " + name).QueryString()
	if err != nil || len(rows) == 0 {
		return "", false
	}
	for _, v := range rows[0] {
		return v, true
	}
	return "", false
}

// logSQLitePragmas 在启动时回读关键 PRAGMA 并落日志。
// 库文件若位于不支持 WAL 的文件系统（部分网络盘、云同步目录），SQLite 会
// 静默回退到 delete 日志模式——把回退暴露在启动日志里，免得等到搜索变慢、
// 写入卡顿才去猜原因。
func logSQLitePragmas(maxOpenConns int) {
	mode, _ := sqlitePragma("journal_mode")
	sync, _ := sqlitePragma("synchronous")
	busy, _ := sqlitePragma("busy_timeout")
	if !strings.EqualFold(mode, "wal") {
		log.Warnf("SQLite journal_mode 未生效（当前 %q）：库文件所在文件系统可能不支持 WAL", mode)
	}
	log.Infof("SQLite: journal_mode=%s synchronous=%s busy_timeout=%s max_open_conns=%d",
		mode, sync, busy, maxOpenConns)
}
