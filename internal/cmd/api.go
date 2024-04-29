package cmd

import (
	"github.com/urfave/cli/v2"

	"zgia.net/book/internal/conf"
	"zgia.net/book/internal/modules/translation"
	"zgia.net/book/internal/route"
)

var Api = &cli.Command{
	Name:   "api",
	Usage:  "Runs as API Server",
	Action: runApi,
}

func runApi(c *cli.Context) error {
	// I18n
	translation.InitLocales()
	translation.Lang = translation.NewLocale(conf.I18n.Langs[0])

	return route.GoHttp()
}
