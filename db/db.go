package db

import (
	"fmt"
	"strings"
	"time"

	// https://go.dev/ref/spec#Import_declarations
	_ "github.com/go-sql-driver/mysql"
	"xorm.io/core"
	"xorm.io/xorm"
	xormlog "xorm.io/xorm/log"

	"zgia.net/book/conf"
	log "zgia.net/book/logger"
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

func getEngine() (*xorm.Engine, error) {
	Param := "?"
	if strings.Contains(conf.Database.Name, Param) {
		Param = "&"
	}

	connStr := ""
	switch conf.Database.Type {
	case "mysql":
		conf.UseMySQL = true
		if conf.Database.Host[0] == '/' { // looks like a unix socket
			connStr = fmt.Sprintf("%s:%s@unix(%s)/%s%scharset=utf8mb4&parseTime=true",
				conf.Database.User, conf.Database.Password, conf.Database.Host, conf.Database.Name, Param)
		} else {
			connStr = fmt.Sprintf("%s:%s@tcp(%s)/%s%scharset=utf8mb4&parseTime=true",
				conf.Database.User, conf.Database.Password, conf.Database.Host, conf.Database.Name, Param)
		}
		var engineParams = map[string]string{"rowFormat": "DYNAMIC"}

		//log.Debugf("ConnStr is %v", connStr)

		return xorm.NewEngineWithParams(conf.Database.Type, connStr, engineParams)

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
	x.SetMaxOpenConns(conf.Database.MaxOpenConns)
	x.SetMaxIdleConns(conf.Database.MaxIdleConns)
	x.SetConnMaxLifetime(time.Second)

	zapLogger := &log.XormZapLogger{}
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
