package testify

import (
	"os"
	"strings"

	"zgia.net/book/internal/util"
)

const PROJECT_SRC_NAME = "book-go"

func init() {
	os.Setenv("BOOK_WORK_DIR", AppPath())
}

func AppPath() string {
	return util.Substring(util.PWD(), 0, strings.Index(util.PWD(), PROJECT_SRC_NAME)+len(PROJECT_SRC_NAME))
}
