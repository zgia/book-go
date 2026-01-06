package route

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"zgia.net/book/internal/db"
	models "zgia.net/book/internal/models"
	"zgia.net/book/internal/util"
)

func ListVolumes(c *gin.Context) {
	bookid := util.ParamInt64(c.Param("bookid"))

	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	volumes, err := models.ListVolumes(book)
	if err != nil {
		Json500(c, fmt.Sprintf("List <%d,%v> volumes: %s", bookid, book.Title, err.Error()))
		return
	}

	Json200(c, volumes)
}

func VolumeExists(c *gin.Context, bookid, volumeid int64) *db.Volume {
	volume, err := db.QueryVolume(volumeid)

	if err != nil {
		Json500(c, fmt.Sprintf("Query volume(%d): %s", volumeid, err.Error()))
		return nil
	}

	if volume == nil {
		Json404(c, fmt.Sprintf("Volume(%d) not found", volumeid))
		return nil
	}

	// 检查volume是否属于指定的书籍
	if volume.Bookid != bookid {
		Json404(c, fmt.Sprintf("Volume(%d) does not belong to book(%d)", volumeid, bookid))
		return nil
	}

	return volume
}

func GetVolume(c *gin.Context) {
	volumeid := util.ParamInt64(c.Param("volumeid"))
	bookid := util.ParamInt64(c.Param("bookid"))

	if volumeid <= 0 {
		Json404(c, fmt.Sprintf("Invalid volume id: %d", volumeid))
		return
	}

	// 检查书籍是否存在
	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	volume := VolumeExists(c, bookid, volumeid)
	if volume == nil {
		return
	}

	Json200(c, models.GetVolume(volume))
}

// UpdateVolume updates/creates a volume
func UpdateVolume(c *gin.Context) {
	volumeid := util.ParamInt64(c.Param("volumeid"))
	bookid := util.ParamInt64(c.Param("bookid"))

	// 检查书籍是否存在
	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	title := c.PostForm("title")
	if title == "" {
		Json500(c, "Volume title cannot be empty")
		return
	}

	// 检查volume是否存在（更新时）
	if volumeid > 0 {
		volume := VolumeExists(c, bookid, volumeid)
		if volume == nil {
			return
		}
	}

	volume := &db.Volume{
		Title:   title,
		Bookid:  bookid,
		Summary: c.PostForm("summary"),
	}

	volumeid, err := db.UpdateVolume(volumeid, volume)
	if err != nil {
		Json500(c, fmt.Sprintf("Update volume(%v): %s", volume.Title, err.Error()))
		return
	}
	volume.Id = volumeid

	Json200(c, map[string]int64{"id": volume.Id})
}

// DeleteVolume deletes a volume
func DeleteVolume(c *gin.Context) {
	volumeid := util.ParamInt64(c.Param("volumeid"))
	bookid := util.ParamInt64(c.Param("bookid"))

	if volumeid <= 0 {
		Json404(c, fmt.Sprintf("Invalid volume id: %d", volumeid))
		return
	}

	// 检查书籍是否存在
	book := BookExists(c, bookid)
	if book == nil {
		return
	}

	// 检查volume是否存在
	volume := VolumeExists(c, bookid, volumeid)
	if volume == nil {
		return
	}

	_, err := db.DeleteVolume(volumeid)
	if err != nil {
		Json500(c, fmt.Sprintf("Delete volume(%d): %s", volumeid, err.Error()))
		return
	}

	data := map[string]int64{"Id": volumeid}

	Json200(c, data)
}
