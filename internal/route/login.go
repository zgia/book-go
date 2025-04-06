package route

import (
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"

	"zgia.net/book/internal/conf"
	"zgia.net/book/internal/db"
	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/middleware"
	"zgia.net/book/internal/models"
)

func HelloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	user, _ := c.Get(middleware.JwtIdentityKey)
	log.Infof("HelloHandler\n\nclaims: %v\nuser: %v", claims, user)

	Json200(c, map[string]any{
		"userID":   claims[middleware.JwtIdentityKey],
		"userName": user.(*middleware.UserResult).Username,
	})
}

func UserInfo(c *gin.Context) {
	value, exists := c.Get(middleware.JwtIdentityKey)

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": http.StatusUnauthorized,
			"msg":  "Unauthorized",
		})

		return
	}

	user := value.(*middleware.UserResult)

	if user.Id == 0 || user.Username == "" {
		Json404(c, "User not found")
		return
	}

	data := map[string]*middleware.UserResult{
		"user": user,
	}

	Json200(c, data)
}

func ChangePassword(c *gin.Context) {
	user, _ := c.Get(middleware.JwtIdentityKey)
	username := user.(*middleware.UserResult).Username

	password := c.PostForm("password")
	password1 := c.PostForm("password1")
	password2 := c.PostForm("password2")

	if password1 != password2 {
		Json500(c, "New Passwords are not matched.")

		return
	}

	u, err := db.GetUser(username, password)
	if err != nil {
		Json500(c, err.Error())

		return
	}

	err = db.ChangeUserPassword(u, password1)
	if err != nil {
		Json500(c, err.Error())
	} else {
		Json200(c, make(map[string]interface{}))
	}
}

func GlobalOptions(c *gin.Context) {

	categories, _ := models.ListCategories()

	data := map[string]any{
		"pagesize":   conf.PageSize(0),
		"categories": categories,
	}

	Json200(c, data)
}
