package route

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"zgia.net/book/db"
	"zgia.net/book/util"
)

// 接口返回
type VolumeResult struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

func ListVolumes(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))

	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	volumes, err := db.QueryVolumes(bookid)

	msg := fmt.Sprintf("List <%d,%v> volumes: ", bookid, book.Title)

	if err != nil {
		JSON500(c, msg+err.Error())

		return
	}

	vr := make([]*VolumeResult, len(volumes))

	for i, v := range volumes {
		vr[i] = &VolumeResult{
			Id:      v.Id,
			Title:   v.Title,
			Summary: v.Summary,
		}
	}

	JSON200(c, msg, map[string]any{"items": vr, "book": map[string]string{"title": book.Title}})
}

// UpdateVolume updates/creates a volume
func UpdateVolume(c *gin.Context) {
	volumeid := util.ParamInt64(c.Param("volumeid"))
	bookid := util.ParamInt64(c.Param("bookid"))

	volume := &db.Volume{
		Title:   c.PostForm("title"),
		Bookid:  bookid,
		Summary: c.PostForm("summary"),
	}

	volumeid, err := db.UpdateVolume(volumeid, volume)
	volume.Id = volumeid

	msg := fmt.Sprintf("Update volume(%v): ", volume.Title)

	if err != nil {
		JSON500(c, msg+err.Error())

		return
	}

	JSON200(c, msg, map[string]int64{"id": volume.Id})
}

// DeleteVolume deletes a volume
func DeleteVolume(c *gin.Context) {
	volumeid := util.ParamInt64(c.Param("volumeid"))
	_, err := db.DeleteVolume(volumeid)

	msg := fmt.Sprintf("Delete volume(%d): ", volumeid)

	if err != nil {
		JSON500(c, msg+err.Error())

		return
	}

	data := map[string]int64{"Id": volumeid}
	JSON200(c, msg, data)
}
