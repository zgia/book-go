package route

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"zgia.net/book/db"
	log "zgia.net/book/logger"
	"zgia.net/book/util"
)

// 接口返回
type ChapterResult struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
}

// ListChapters gets all chapters
func ListChapters(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))

	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	chapters, err := db.QueryChapters(bookid)

	msg := fmt.Sprintf("List <%d,%v> chapters: ", bookid, book.Title)

	if err != nil {
		JSON500(c, msg+err.Error())

		return
	}

	volumes, _ := db.QueryVolumes(bookid)
	if len(volumes) == 0 {
		volumes = make([]*db.Volume, 1)
		volumes[0] = &db.Volume{Id: 0, Title: book.Title, Summary: ""}
	}

	data := make(map[int64][]*ChapterResult)
	for _, v := range chapters {
		data[v.Volumeid] = append(data[v.Volumeid], &ChapterResult{
			Id:    v.Id,
			Title: v.Title,
		})
	}

	if len(data) == 0 {
		var empty []*ChapterResult
		for _, v := range volumes {
			data[v.Id] = empty
		}
	}

	vols := make([]*VolumeResult, len(volumes))
	for i, v := range volumes {
		vols[i] = &VolumeResult{
			Id:      v.Id,
			Title:   v.Title,
			Summary: v.Summary,
		}
	}

	dt := map[string]any{
		"items":   data,
		"volumes": vols,
	}
	JSON200(c, msg, dt)
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

	msg := fmt.Sprintf("Get chapter(%d): ", chapterid)

	chapter := map[string]any{}

	if chapterid != 0 {
		chap, err := db.QueryChapter(chapterid, next)

		if err != nil {
			JSON500(c, msg+err.Error())

			return

		} else if chap == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"code": http.StatusNotFound,
				"msg":  msg + "not found",
			})

			return
		}

		chapter = chap
	}

	chapter["book"] = map[string]string{"title": book.Title}

	JSON200(c, msg, chapter)
}

// UpdateChapter updates/creates a chapter
func UpdateChapter(c *gin.Context) {
	chapterid := util.ParamInt64(c.Param("chapterid"))
	bookid := util.ParamInt64(c.Param("bookid"))

	chapter := &db.Chapter{
		Title:    c.PostForm("title"),
		Bookid:   bookid,
		Volumeid: util.ParamInt64(c.PostForm("volumeid")),
	}
	content := &db.Content{Txt: c.PostForm("content")}

	log.Infof("UpdateChapter chapter: %#v", chapter)

	chapterid, err := db.UpdateChapter(chapterid, chapter, content)
	chapter.Id = chapterid

	msg := fmt.Sprintf("Update chapter(%v): ", chapter.Title)

	if err != nil {
		JSON500(c, msg+err.Error())

		return
	}

	JSON200(c, msg, chapter)
}

// DeleteChapter deletes a chapter
func DeleteChapter(c *gin.Context) {
	chapterid := util.ParamInt64(c.Param("chapterid"))
	_, err := db.DeleteChapter(chapterid)

	msg := fmt.Sprintf("Delete chapter(%d): ", chapterid)

	if err != nil {
		JSON500(c, msg+err.Error())

		return
	}

	data := map[string]int64{"Id": chapterid}
	JSON200(c, msg, data)
}
