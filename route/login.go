package route

import (
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"

	"zgia.net/book/db"
	log "zgia.net/book/logger"
	"zgia.net/book/middleware"
)

func HelloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	user, _ := c.Get(middleware.JwtIdentityKey)
	log.Infof("HelloHandler\n\nclaims: %v\nuser: %v", claims, user)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "Hello World.",
		"data": map[string]any{
			"userID":   claims[middleware.JwtIdentityKey],
			"userName": user.(*db.User).Username,
		},
	})
}

func IsLogined(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{
			"code": http.StatusNotFound,
			"msg":  "User not found",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "ok",
		"data": map[string]*middleware.UserResult{"user": user},
	})
}

func GlobalOptions(c *gin.Context) {
	data := map[string]any{
		"pwdminlen": 6,
		"perpage":   20,
	}

	JSON200(c, "category", data)
}
