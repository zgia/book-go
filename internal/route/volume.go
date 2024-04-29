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

func GetVolume(c *gin.Context) {
	volumeid := util.ParamInt64(c.Param("volumeid"))

	volume, err := db.QueryVolume(volumeid)
	if err != nil {
		Json500(c, fmt.Sprintf("Get volume(%d): %s", volumeid, err.Error()))
		return
	}

	Json200(c, models.GetVolume(volume))
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

	_, err := db.DeleteVolume(volumeid)
	if err != nil {
		Json500(c, fmt.Sprintf("Delete volume(%d): %s", volumeid, err.Error()))
		return
	}

	data := map[string]int64{"Id": volumeid}

	Json200(c, data)
}
