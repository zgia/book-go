package main

import (
	"flag"
	"fmt"

	"go.uber.org/zap"
	"zgia.net/book/conf"
	"zgia.net/book/db"
	log "zgia.net/book/logger"
	"zgia.net/book/route"
	"zgia.net/book/util"
)

func init() {
	conf.App.Version = "0.0.1"
}

func main() {
	var customConf string
	flag.StringVar(&customConf, "c", "", "Custom config path")
	flag.Parse()

	if err := initialize(customConf); err != nil {
		panic(fmt.Sprintf("Failed to initialize application: %v", err))
	}

	route.GoHttp()
}

// Get global configuration.
func initialize(customConf string) error {
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
