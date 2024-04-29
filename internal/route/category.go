package route

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"zgia.net/book/internal/db"
	models "zgia.net/book/internal/models"
	"zgia.net/book/internal/util"
)

func ListCategories(c *gin.Context) {

	categories, err := models.ListCategories()

	if err != nil {
		Json500(c, "List categories: "+err.Error())
		return
	}

	Json200(c, categories)
}

func UpdateCategory(c *gin.Context) {
	catid := util.ParamInt64(c.Param("catid"))

	cat := &db.Category{
		Id:    catid,
		Title: c.PostForm("title"),
	}

	catid, err := db.UpdateCategory(cat, catid)
	if err != nil {
		Json500(c, fmt.Sprintf("Update category(%v): %s", cat.Title, err.Error()))
		return
	}
	cat.Id = catid

	Json200(c, cat)
}

func DeleteCategory(c *gin.Context) {
	catid := util.ParamInt64(c.Param("catid"))
	_, err := db.DeleteCategory(catid)

	if err != nil {
		Json500(c, fmt.Sprintf("Delete category(%d): %s", catid, err.Error()))
		return
	}

	Json200(c, map[string]int64{"Id": catid})
}
