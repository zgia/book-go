package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/models"
	"zgia.net/book/internal/util"
)

var Export = &cli.Command{
	Name:      "export",
	HelpName:  "Export",
	Usage:     "Exports books to text file",
	UsageText: "book -c xxx export [command options] [arguments...]",
	Action:    runText,
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:     "bookid",
			Aliases:  []string{"b"},
			Value:    0,
			Usage:    "Export books after the book id",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Value:   "/tmp",
			Usage:   "Export books to path, such as /tmp",
		},
	},
}

func runText(c *cli.Context) error {
	path := c.String("path")

	if !(util.IsDir(path) && util.IsWritable(path)) {
		msg := fmt.Sprintf("%s is not writable\n", path)
		log.Fatal(msg)

		return errors.New(msg)
	}

	models.SaveAllBooksToTxt(c.Int64("bookid"), strings.TrimSuffix(path, "/"))

	return nil
}
