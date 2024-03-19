package initial

import (
	"fmt"

	"go.uber.org/zap"
	"zgia.net/book/internal/conf"
	"zgia.net/book/internal/db"
	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/util"
)

// Get global configuration.
func Initialize(customConf string) error {
	if err := conf.Init(customConf); err != nil {
		return err
	}

	log.InitLogger(conf.File, conf.LogDir(), conf.Time.FormatLayout)

	log.Infof("%s %s", conf.App.BrandName, conf.App.Version)
	log.Debug("App dir config",
		zap.String("HomeDir", util.HomeDir()),
		zap.String("AppPath", util.AppPath()),
		zap.String("WorkDir", util.WorkDir()),
		zap.String("PWD", util.PWD()),
		zap.String("CustomDir", util.CustomDir()),
		zap.String("CustomConfig", conf.CustomConf),
		zap.String("LogDir", conf.LogDir()),
	)

	if err := db.NewEngine(); err != nil {
		return fmt.Errorf("init XORM: %v", err)
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping db: %v", err)
	}

	return nil
}
