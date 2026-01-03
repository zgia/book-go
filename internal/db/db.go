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

	connStr := ""
	switch conf.Database.Type {
	case "mysql":
		conf.UseMySQL = true
		conf.UseSQLite = false
		if conf.Database.Host[0] == '/' { // looks like a unix socket
			connStr = fmt.Sprintf("%s:%s@unix(%s)/%s%scharset=utf8mb4&parseTime=true",
				conf.Database.User, conf.Database.Password, conf.Database.Host, conf.Database.Name, Param)
		} else {
			connStr = fmt.Sprintf("%s:%s@tcp(%s)/%s%scharset=utf8mb4&parseTime=true",
				conf.Database.User, conf.Database.Password, conf.Database.Host, conf.Database.Name, Param)
		}
		var engineParams = map[string]string{"rowFormat": "DYNAMIC"}

		return xorm.NewEngineWithParams(conf.Database.Type, connStr, engineParams)

	case "sqlite", "sqlite3":
		conf.UseMySQL = false
		conf.UseSQLite = true
		// For SQLite, Host field contains the file path
		// Format: file:path/to/database.db?cache=shared&mode=rwc
		sqlitePath := conf.Database.Host
		if sqlitePath == "" {
			sqlitePath = conf.Database.Name // fallback to Name field
		}
		// Ensure file: prefix for XORM SQLite driver
		if !strings.HasPrefix(sqlitePath, "file:") {
			sqlitePath = "file:" + sqlitePath
		}
		// Add cache=shared for better concurrency handling
		// Add _foreign_keys=1 to enable foreign key constraints
		if !strings.Contains(sqlitePath, "?") {
			sqlitePath += "?cache=shared&_foreign_keys=1"
		} else {
			sqlitePath += "&cache=shared&_foreign_keys=1"
		}
		// SQLite doesn't need user/password, but keep the connection string simple
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

	// Adjust connection pool settings based on database type
	if conf.UseSQLite {
		// SQLite has limited concurrency, use smaller connection pool
		x.SetMaxOpenConns(1)
		x.SetMaxIdleConns(1)
		// SQLite connections can be kept longer
		x.SetConnMaxLifetime(time.Hour)
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
