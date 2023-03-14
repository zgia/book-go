package route

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"zgia.net/book/db"
	log "zgia.net/book/logger"
	"zgia.net/book/modules"
	"zgia.net/book/util"
)

// 接口返回
type BookResult struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Summary string `json:"summary"`
	Source  string `json:"source"`
}

func ListBooks(c *gin.Context) {
	page := util.PageNum(c.Query("p"))
	title := c.Query("title")

	books, err := db.QueryBooks(page, title)

	msg := "List books: "

	if err != nil {
		JSON500(c, msg+err.Error())

		return
	}

	bs := make([]*BookResult, len(books))

	for i, v := range books {
		bs[i] = &BookResult{
			Id:      v.Id,
			Title:   v.Title,
			Author:  v.Author,
			Summary: v.Summary,
			Source:  v.Source,
		}
	}

	data := map[string]any{
		"total": db.CountBooks(title),
		"items": bs,
	}
	JSON200(c, msg, data)
}

func BookExists(c *gin.Context, bookid int64) *db.Book {
	book, err := db.QueryBook(bookid)

	if err != nil {
		panic(err.Error())
	}

	if book == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": http.StatusNotFound,
			"msg":  fmt.Sprintf("Book(%d) not found", bookid),
		})
	}

	return book
}

func GetBook(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))

	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	msg := fmt.Sprintf("Get book(%d): success", bookid)
	data := &BookResult{
		Id:      book.Id,
		Title:   book.Title,
		Author:  book.Author,
		Summary: book.Summary,
		Source:  book.Source,
	}

	JSON200(c, msg, data)
}

// TODO...
func DownloadBook(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))

	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	chapters := db.QueryAllChapters(bookid)

	fpath, filesize := modules.WriteToFile(book, chapters)

	defer func() {
		if err := os.RemoveAll(filepath.Dir(fpath)); err != nil {
			log.Errorf(fmt.Sprintf("Remove file(%s) for book(%d) error, %s", fpath, book.Id, err.Error()))
		}
	}()

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(fpath)))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")
	c.Header("Expires", "0")
	c.Header("Pragma", "public")
	c.Header("Content-Length", fmt.Sprintf("%d", filesize))

	c.File(fpath)
}

func UpdateBook(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))

	book := &db.Book{
		Title:   c.PostForm("title"),
		Author:  c.PostForm("author"),
		Summary: c.PostForm("summary"),
		Source:  c.PostForm("source"),
	}

	log.Infof("UpdateBook book: %#v", book)

	bookid, err := db.UpdateBook(book, bookid)
	book.Id = bookid

	msg := fmt.Sprintf("Update book(%v): ", book.Title)

	if err != nil {
		JSON500(c, msg+err.Error())

		return
	}

	JSON200(c, msg, book)
}

func DeleteBook(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))
	_, err := db.DeleteBook(bookid)

	msg := fmt.Sprintf("Delete Book(%d): ", bookid)

	if err != nil {
		JSON500(c, msg+err.Error())

		return
	}

	JSON200(c, msg, map[string]int64{"Id": bookid})
}
