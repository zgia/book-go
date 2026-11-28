package cmd

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"

	"zgia.net/book/internal/db"
)

var FTS = &cli.Command{
	Name:      "fts",
	HelpName:  "FTS",
	Usage:     "Build or inspect the full-text search index (SQLite only)",
	UsageText: "book -c xxx fts <rebuild|stat>",
	Subcommands: []*cli.Command{
		{
			Name:   "rebuild",
			Usage:  "Rebuild the 2-gram index for all chapter contents",
			Action: runFTSRebuild,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:    "batch",
					Aliases: []string{"b"},
					Value:   500,
					Usage:   "Rows per transaction",
				},
			},
		},
		{
			Name:   "stat",
			Usage:  "Show index coverage",
			Action: runFTSStat,
		},
	},
}

func runFTSRebuild(c *cli.Context) error {
	batch := c.Int("batch")

	fmt.Printf("开始重建正文索引（每批 %d 行），可能需要数分钟…\n", batch)

	start := time.Now()
	last := int64(0)
	done, err := db.RebuildFTS(batch, func(done, total int64) {
		if total <= 0 {
			return
		}
		if done-last >= 5000 || done >= total {
			last = done
			fmt.Printf("  %d/%d 章 (%.1f%%)\n", done, total, float64(done)*100/float64(total))
		}
	})
	if err != nil {
		return fmt.Errorf("重建索引失败: %w", err)
	}

	indexed, _ := db.FTSCount()
	fmt.Printf("完成：处理 %d 行正文，实际入索引 %d 章（正文为空或仅含标点的章节不收录），耗时 %v\n",
		done, indexed, time.Since(start).Round(time.Second))

	return nil
}

func runFTSStat(c *cli.Context) error {
	indexed, err := db.FTSCount()
	if err != nil {
		return fmt.Errorf("索引表不可用（%v），先执行 `book fts rebuild`", err)
	}

	total, err := db.CountContents()
	if err != nil {
		return err
	}

	const probeLimit = 10000
	gaps, err := db.FTSGaps(probeLimit)
	if err != nil {
		return err
	}

	missing, empty := 0, 0
	for _, g := range gaps {
		if g.Indexable {
			missing++
		} else {
			empty++
		}
	}

	fmt.Printf("索引表  : %s\n已索引  : %d 章\n正文总数: %d 章\n", "content_fts", indexed, total)

	if missing == 0 && empty == 0 {
		fmt.Println("状态    : 全部已索引")
		return nil
	}

	if missing == 0 {
		fmt.Printf("状态    : 已全部索引（另有 %d 章正文为空或仅含标点，2-gram 无法索引，重建也补不上）\n", empty)
		return nil
	}

	fmt.Printf("待补索引: %d 章可索引但未进索引%s\n", missing, gapNote(empty))
	fmt.Println("处理    : 执行 `book fts rebuild`，章节写入时也会自动同步")
	if len(gaps) >= probeLimit {
		fmt.Println("          差异超过探测上限，建议直接全量 rebuild")
	}

	return nil
}

// gapNote 把“空正文章节”这种正常差异拼进提示，避免误判成索引没建好。
func gapNote(empty int) string {
	if empty == 0 {
		return ""
	}

	return fmt.Sprintf("；另有 %d 章正文为空或仅含标点，不参与索引", empty)
}
