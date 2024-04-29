package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"zgia.net/book/internal/cmd"
	"zgia.net/book/internal/conf"
	"zgia.net/book/internal/initial"
	"zgia.net/book/internal/util"
)

func init() {
	conf.App.Version = "0.0.1"
	conf.App.BrandName = "zGia! Book Library"
}

func main() {
	app := &cli.App{
		Name:      conf.App.BrandName,
		Version:   conf.App.Version,
		Usage:     "Book library api server",
		UsageText: "book [global options] command [command options]",
		Before: func(ctx *cli.Context) error {

			// 检查command
			if ctx.NArg() == 0 {
				return nil
			}
			cmd := ctx.Args().Get(0)
			y := false
			for _, command := range ctx.App.Commands {
				if command.Name == cmd {
					y = true
					break
				}
			}
			if !y {
				return nil
			}

			if err := initial.Initialize(ctx.String("config")); err != nil {
				return errors.Wrap(err, "Failed to initialize application")
			}

			return nil
		},
		ExtraInfo: func() map[string]string {
			return map[string]string{
				"App Name":      conf.App.BrandName + " " + conf.App.Version,
				"App Path":      util.AppPath(),
				"Custom Config": conf.CustomConf,
				"Current Dir":   util.PWD(),
				"Custom Dir":    util.CustomDir(),
				"Home Dir":      util.HomeDir(),
				"Logs Dir":      conf.LogDir(),
				"Work Dir":      util.WorkDir(),
			}
		},
		OnUsageError: func(ctx *cli.Context, err error, isSubcommand bool) error {
			if isSubcommand {
				return err
			}

			fmt.Fprintf(ctx.App.Writer, "WRONG: %s\n", err.Error())
			return nil
		},
		CommandNotFound: func(ctx *cli.Context, command string) {
			fmt.Fprintf(ctx.App.Writer, "\nCannot find command %q here.\n", command)
		},
		Commands: []*cli.Command{
			cmd.Api,
			cmd.Export,
			cmd.Extra,
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "",
				Usage:   "custom config path: /path/to/app.ini",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("\n\nFailed to start application: %v\n", err)
	}
}
