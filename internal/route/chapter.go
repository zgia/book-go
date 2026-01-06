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

	chapters, err := db.QueryVolumeChapters(bookid, volumeid)
	if err != nil {
		Json500(c, fmt.Sprintf("Get volume chapters(%d,%d): %s", bookid, volumeid, err.Error()))
		return
	}

	content := models.GetVolumeChapters(book, chapters)

	data := map[string]string{"content": content}
	Json200(c, data)
}

func ChapterExists(c *gin.Context, bookid, chapterid int64) map[string]any {
	chapter, err := db.QueryChapter(bookid, chapterid, 0)

	if err != nil {
		Json500(c, fmt.Sprintf("Query chapter(%d,%d): %s", bookid, chapterid, err.Error()))
		return nil
	}

	if chapter == nil {
		Json404(c, fmt.Sprintf("Chapter(%d,%d) not found", bookid, chapterid))
		return nil
	}

	return chapter
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

	// 检查书籍是否存在
	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	title := c.PostForm("title")
	postCont := c.PostForm("content")
	if title == "" || postCont == "" {
		Json500(c, "Chapter title and content cannot be empty")
		return
	}

	// 检查章节是否存在（更新时）
	if chapterid > 0 {
		chapter := ChapterExists(c, bookid, chapterid)
		if chapter == nil {
			return
		}
	}

	content := &db.Content{Txt: postCont}
	chapter := &db.Chapter{
		Title:     title,
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
	bookid := util.ParamInt64(c.Param("bookid"))

	if chapterid <= 0 {
		Json404(c, fmt.Sprintf("Invalid chapter id: %d", chapterid))
		return
	}

	// 检查书籍是否存在
	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	// 检查章节是否存在
	chapter := ChapterExists(c, bookid, chapterid)
	if chapter == nil {
		return
	}

	if _, err := db.DeleteChapter(chapterid); err != nil {
		Json500(c, fmt.Sprintf("Delete chapter(%d): %s", chapterid, err.Error()))
		return
	}

	data := map[string]int64{"Id": chapterid}
	Json200(c, data)
}
