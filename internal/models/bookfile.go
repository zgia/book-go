package models

import (
	"bufio"
	"fmt"
	"math"
	"os"

	"zgia.net/book/internal/db"
	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/util"
)

func WriteToFile(book *db.Book, chapters []map[string]string) string {

	log.Infof("%d chapters in book <%s>", len(chapters), book.Title)

	dir, err := os.MkdirTemp("", "books-")
	if err != nil {
		panic(fmt.Sprintf("Cannot create tmp dir for book(%d), %s", book.Id, err.Error()))
	}

	file, err := os.Create(dir + fmt.Sprintf("/%s.txt", book.Title))
	if err != nil {
		panic(fmt.Sprintf("Cannot create tmp text file for book(%d), %s", book.Id, err.Error()))
	}

	log.Infof("Tmp text file is %s", file.Name())

	defer func() {
		if err := file.Close(); err != nil {
			panic(fmt.Sprintf("Close file(%s) for book(%d) error, %s", file.Name(), book.Id, err.Error()))
		}
	}()

	writer := bufio.NewWriter(file)

	writeStr := func(str string) {
		if _, err := writer.WriteString(str); err != nil {
			panic(fmt.Sprintf("Write book(%d) to file(%s) error, %s", book.Id, file.Name(), err.Error()))
		}
	}
	flush := func() {
		if err := writer.Flush(); err != nil {
			panic(fmt.Sprintf("Write flush book(%d) to file(%s) error, %s", book.Id, file.Name(), err.Error()))
		}
	}

	// 书名
	writeStr(fmt.Sprintf("%s%s%s%s%s%s", book.Title, util.Eol(2), book.Author, util.Eol(2), book.Summary, util.Eol(3)))

	volTitle := ""
	for i, v := range chapters {

		if volTitle != v["volTitle"] {
			volTitle = v["volTitle"]

			// 卷名
			if v["volTitle"] != book.Title {
				writeStr(fmt.Sprintf("%s%s", v["volTitle"], util.Eol(3)))
			}
		}

		// 章节
		writeStr(fmt.Sprintf("%s%s%s%s", v["title"], util.Eol(2), v["content"], util.Eol(2)))

		if math.Mod(float64(i), 5) == 0 {
			flush()
		}
	}

	flush()

	src := file.Name()
	dst := src + ".zip"

	if err := util.ZipFile(src, dst); err != nil {
		panic(fmt.Sprintf("Cannot zip file , %s", err.Error()))
	}

	return dst
}
