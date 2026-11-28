package initial

import (
	"fmt"

	"zgia.net/book/internal/conf"
	"zgia.net/book/internal/db"
	log "zgia.net/book/internal/logger"
)

// Get global configuration.
func Initialize(customConf string) error {
	if err := conf.Init(customConf); err != nil {
		return err
	}

	log.InitLogger(conf.File, conf.LogDir(), conf.Time.FormatLayout)

	if err := db.NewEngine(); err != nil {
		return fmt.Errorf("init XORM: %v", err)
	}

	// 探测正文全文索引；不可用时搜索自动回退 LIKE，不影响启动。
	db.InitFTS()

	return nil
}
