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

func CategoryExists(c *gin.Context, id int64) *db.Category {
	category, err := db.QueryCategory(id)

	if err != nil {
		Json500(c, fmt.Sprintf("Query category(%d): %s", id, err.Error()))
		return nil
	}

	if category == nil {
		Json404(c, fmt.Sprintf("Category(%d) not found", id))
		return nil
	}

	return category
}

func UpdateCategory(c *gin.Context) {
	catid := util.ParamInt64(c.Param("catid"))

	title := c.PostForm("title")
	if title == "" {
		Json500(c, "Category title cannot be empty")
		return
	}

	// 检查类别是否存在（更新时）
	if catid > 0 {
		category := CategoryExists(c, catid)
		if category == nil {
			return
		}
	}

	cat := &db.Category{
		Id:    catid,
		Title: title,
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
	if catid <= 0 {
		Json404(c, fmt.Sprintf("Invalid category id: %d", catid))
		return
	}

	// 检查类别是否存在
	category := CategoryExists(c, catid)
	if category == nil {
		return
	}

	_, err := db.DeleteCategory(catid)

	if err != nil {
		Json500(c, fmt.Sprintf("Delete category(%d): %s", catid, err.Error()))
		return
	}

	Json200(c, map[string]int64{"Id": catid})
}
