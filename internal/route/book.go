package route

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"zgia.net/book/internal/db"
	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/util"
	models "zgia.net/book/internal/models"
)

func SearchBooks(c *gin.Context) {
	words := c.Query("words")
	bookid := util.ParamInt64(c.Query("bookid"))

	books := models.SearchBooks(words, bookid)

	Json200(c, books)
}

func ListBooks(c *gin.Context) {
	page := util.PageNum(c.Query("p"))
	title := c.Query("title")
	searchMode := c.Query("smode")

	books, err := models.ListBooks(page, title, searchMode)

	if err != nil {
		Json500(c, "List books: "+err.Error())
		return
	}

	Json200(c, books)
}

func BookExists(c *gin.Context, bookid int64) *db.Book {
	book, err := db.QueryBook(bookid)

	if err != nil {
		panic(err.Error())
	}

	if book == nil {
		Json404(c, fmt.Sprintf("Book(%d) not found", bookid))
		return nil
	}

	return book
}

func GetBook(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))

	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	Json200(c, models.GetBook(book))
}

func SaveAllBooksToTxt() (int64, int64) {
	count, _ := db.CountBooks("", "")

	ids, err := db.QueryAllBookIds()
	if err != nil {
		log.Errorf("%v", err)
		return 0, 0
	}

	i := len(ids)

	// i := 0
	// for _, id := range ids {
	// 	book, err := db.QueryBook(id)
	// 	if err != nil {
	// 		log.Errorf("%v", err)
	// 		break
	// 	}
	// 	if book == nil {
	// 		log.Errorf("book %d is not exist", id)
	// 		break
	// 	}

	// 	i++
	// 	chapters := db.QueryAllChapters(id)
	// 	fpath := models.WriteToFile(book, chapters)

	// 	srcInfo, _ := os.Lstat(fpath)
	// 	util.MoveFile(fpath, "/Users/zgia/Desktop/books/"+srcInfo.Name())

	// 	log.Infof("book %d save to %s...", id, fpath)
	// }

	log.Infof("books saved")

	return count, int64(i)
}

// DownloadBook downloads book to plain text file
func DownloadBook(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))

	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	chapters := db.QueryAllChapters(bookid)

	fpath := models.WriteToFile(book, chapters)

	filesize := 0
	if fs, err := os.Stat(fpath); err == nil {
		filesize = int(fs.Size())
	}

	// 下载后，删除临时文件
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

func SetBookDone(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))

	book := BookExists(c, bookid)
	if book == nil {
		Json500(c, fmt.Sprintf("Set book(%d) done: not exist.", bookid))
		return
	}

	if book.Done == 1 {
		Json200(c, models.GetBook(book))
		return
	}

	book.Done = 1
	_, err := db.UpdateBook(book, bookid)
	if err != nil {
		Json500(c, fmt.Sprintf("Set book(%v) done: %s", book.Title, err.Error()))
		return
	}

	Json200(c, models.GetBook(book))
}

func UpdateBook(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))

	book := &db.Book{
		Categoryid: util.ParamInt64(c.PostForm("categoryid")),
		Title:      c.PostForm("title"),
		Author:     c.PostForm("author"),
		Summary:    c.PostForm("summary"),
		Source:     c.PostForm("source"),
	}

	bookid, err := db.UpdateBook(book, bookid)
	if err != nil {
		Json500(c, fmt.Sprintf("Update book(%v): %s", book.Title, err.Error()))
		return
	}
	book.Id = bookid

	Json200(c, models.GetBook(book))
}

func DeleteBook(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))
	_, err := db.DeleteBook(bookid)

	if err != nil {
		Json500(c, fmt.Sprintf("Delete Book(%d): %s", bookid, err.Error()))
		return
	}

	Json200(c, map[string]int64{"Id": bookid})
}
