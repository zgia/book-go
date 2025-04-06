package route

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"zgia.net/book/internal/db"
	log "zgia.net/book/internal/logger"
	models "zgia.net/book/internal/models"
	"zgia.net/book/internal/util"
)

func ListAuthors(c *gin.Context) {
	page := util.PageNum(c.Query("p"))
	words := c.Query("q")

	authors, err := models.ListAuthors(page, words)

	log.Infof("ListAuthors: %#v", authors)

	if err != nil {
		Json500(c, "List authors: "+err.Error())
		return
	}

	Json200(c, authors)
}

func AuthorExists(c *gin.Context, id int64) *db.Author {
	author, err := db.QueryAuthor(id)

	if err != nil {
		panic(err.Error())
	}

	if author == nil {
		Json404(c, fmt.Sprintf("Author(%d) not found", id))
		return nil
	}

	return author
}

func GetAuthor(c *gin.Context) {
	authorid := util.ParamInt64(c.Param("authorid"))

	author := AuthorExists(c, authorid)
	if author == nil {
		return
	}

	Json200(c, models.GetAuthor(author))
}

func UpdateAuthor(c *gin.Context) {
	authorid := util.ParamInt64(c.Param("authorid"))

	author := &db.Author{
		Name:       c.PostForm("name"),
		FormerName: c.PostForm("former_name"),
	}

	authorid, err := db.UpdateAuthor(author, authorid)
	if err != nil {
		Json500(c, fmt.Sprintf("Update author(%v): %s", author.Name, err.Error()))
		return
	}
	author.Id = authorid

	Json200(c, models.GetAuthor(author))
}

func DeleteAuthor(c *gin.Context) {
	authorid := util.ParamInt64(c.Param("authorid"))
	_, err := db.DeleteAuthor(authorid)

	if err != nil {
		Json500(c, fmt.Sprintf("Delete Author(%d): %s", authorid, err.Error()))
		return
	}

	Json200(c, map[string]int64{"Id": authorid})
}
