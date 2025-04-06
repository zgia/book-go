package route

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"zgia.net/book/internal/db"
	models "zgia.net/book/internal/models"
	"zgia.net/book/internal/util"
)

// ListChapters gets all chapters of volume
func GetVolumeChapters(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))
	volumeid := util.ParamInt64(c.Param("volumeid"))

	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	chapters := db.QueryVolumeChapters(bookid, volumeid)

	content := models.GetVolumeChapters(book, chapters)

	data := map[string]string{"content": content}
	Json200(c, data)
}

// ListChapters gets all chapters
func ListChapters(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))

	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	dt, err := models.ListChapters(book)

	if err != nil {
		Json500(c, fmt.Sprintf("List <%d,%v> chapters: %s", bookid, book.Title, err.Error()))
		return
	}

	Json200(c, dt)
}

// GetChapter gets a chapter
func GetChapter(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))

	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	next := util.ParamInt(c.Query("next"))
	chapterid := util.ParamInt64(c.Param("chapterid"))

	chapter := map[string]any{}

	if chapterid != 0 {
		chap, err := db.QueryChapter(bookid, chapterid, next)

		if err != nil {
			Json500(c, fmt.Sprintf("Get book(%d) chapter(%d): %s", bookid, chapterid, err.Error()))
			return

		} else if chap == nil {
			Json404(c, fmt.Sprintf("Get book(%d) chapter(%d): %s", bookid, chapterid, "not found"))
			return
		}

		chapter = chap
	}

	chapter["book"] = models.GetBook(book)

	Json200(c, chapter)
}

// UpdateChapter updates/creates a chapter
func UpdateChapter(c *gin.Context) {
	chapterid := util.ParamInt64(c.Param("chapterid"))
	bookid := util.ParamInt64(c.Param("bookid"))

	postCont := c.PostForm("content")
	content := &db.Content{Txt: postCont}

	chapter := &db.Chapter{
		Title:     c.PostForm("title"),
		Bookid:    bookid,
		Volumeid:  util.ParamInt64(c.PostForm("volumeid")),
		Wordcount: int64(len([]rune(postCont))),
	}

	chapterid, err := db.UpdateChapter(chapterid, chapter, content)
	if err != nil {
		Json500(c, fmt.Sprintf("Update chapter(%v): %s", chapter.Title, err.Error()))
		return
	}
	chapter.Id = chapterid

	Json200(c, chapter)
}

// DeleteChapter deletes a chapter
func DeleteChapter(c *gin.Context) {
	chapterid := util.ParamInt64(c.Param("chapterid"))

	if _, err := db.DeleteChapter(chapterid); err != nil {
		Json500(c, fmt.Sprintf("Delete chapter(%d): %s", chapterid, err.Error()))
		return
	}

	data := map[string]int64{"Id": chapterid}
	Json200(c, data)
}
