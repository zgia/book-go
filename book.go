package main

import (
	"flag"
	"fmt"

	"zgia.net/book/internal/initial"
	"zgia.net/book/internal/conf"
	"zgia.net/book/internal/modules/translation"
	"zgia.net/book/internal/route"
)

func init() {
	conf.App.Version = "0.0.1"
}

func main() {
	var customConf string
	flag.StringVar(&customConf, "c", "", "Custom config path")
	flag.Parse()

	if err := initial.Initialize(customConf); err != nil {
		panic(fmt.Sprintf("Failed to initialize application: %v", err))
	}

	// I18n
	translation.InitLocales()
	translation.Lang = translation.NewLocale(conf.I18n.Langs[0])

	route.GoHttp()
}
