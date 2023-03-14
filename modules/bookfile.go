package modules

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math"

	"zgia.net/book/db"
	log "zgia.net/book/logger"
	"zgia.net/book/util"
)

func WriteToFile(book *db.Book, chapters []map[string]string) (string, int) {

	log.Infof("%d chapters in book <%s>", len(chapters), book.Title)

	dir, err := ioutil.TempDir("", "books-")
	if err != nil {
		panic(fmt.Sprintf("Cannot create tmp dir for book(%d), %s", book.Id, err.Error()))
	}

	file, err := ioutil.TempFile(dir, fmt.Sprintf("book-%d-*.txt", book.Id))
	if err != nil {
		panic(fmt.Sprintf("Cannot create tmp file for book(%d), %s", book.Id, err.Error()))
	}

	log.Infof("tmpFile is %s", file.Name())

	defer func() {
		if err := file.Close(); err != nil {
			panic(fmt.Sprintf("Close file(%s) for book(%d) error, %s", file.Name(), book.Id, err.Error()))
		}
	}()

	write := bufio.NewWriter(file)

	var wr = func(str string) {
		if _, err := write.WriteString(str); err != nil {
			panic(fmt.Sprintf("Write book(%d) to file(%s) error, %s", book.Id, file.Name(), err.Error()))
		}
	}
	var flush = func() {
		if err := write.Flush(); err != nil {
			panic(fmt.Sprintf("Write flush book(%d) to file(%s) error, %s", book.Id, file.Name(), err.Error()))
		}
	}

	// 书名
	wr(fmt.Sprintf("%s%s%s%s%s%s", book.Title, util.Eol(2), book.Author, util.Eol(2), book.Summary, util.Eol(3)))

	volTitle := ""
	for i, v := range chapters {

		if volTitle != v["volTitle"] {
			volTitle = v["volTitle"]

			// 卷名
			if v["volTitle"] != book.Title {
				wr(fmt.Sprintf("%s%s", v["volTitle"], util.Eol(3)))
			}
		}

		// 章节
		wr(fmt.Sprintf("%s%s%s%s", v["title"], util.Eol(2), v["content"], util.Eol(2)))

		if math.Mod(float64(i), 5) == 0 {
			flush()
		}
	}

	flush()

	filesize := 0
	if fs, err := file.Stat(); err == nil {
		filesize = int(fs.Size())
	}

	return file.Name(), filesize
}
