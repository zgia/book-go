package cmd

import (
	"fmt"
	"sort"

	"github.com/urfave/cli/v2"
)

var Extra = &cli.Command{
	Name:      "extra",
	HelpName:  "Extra",
	Usage:     "Print app extra info",
	UsageText: "book -c xxx extra",
	Action:    runExtra,
}

func runExtra(c *cli.Context) error {
	ex := c.App.ExtraInfo()

	keys := make([]string, len(ex))

	i := 0
	for k := range ex {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("%s\t=> %s\n", k, ex[k])
	}

	return nil
}
