package route

import (
	"fmt"
	"strings"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"

	"zgia.net/book/internal/conf"
	"zgia.net/book/internal/db"
	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/middleware"
	"zgia.net/book/internal/models"
)

// GetCurrentUser extracts current user from JWT context
func GetCurrentUser(c *gin.Context) (*middleware.UserResult, error) {
	value, exists := c.Get(middleware.JwtIdentityKey)
	if !exists {
		return nil, fmt.Errorf("user not found in context")
	}

	user, ok := value.(*middleware.UserResult)
	if !ok {
		return nil, fmt.Errorf("invalid user type in context")
	}

	if user.Id == 0 || user.Username == "" {
		return nil, fmt.Errorf("invalid user data")
	}

	return user, nil
}

func HelloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	user, err := GetCurrentUser(c)
	if err != nil {
		Json401(c, "Unauthorized")
		return
	}

	log.Infof("HelloHandler\n\nclaims: %v\nuser: %v", claims, user)

	Json200(c, map[string]any{
		"userID":   claims[middleware.JwtIdentityKey],
		"userName": user.Username,
	})
}

func UserInfo(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		Json401(c, "Unauthorized")
		return
	}

	data := map[string]*middleware.UserResult{
		"user": user,
	}

	Json200(c, data)
}

func ChangePassword(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		Json401(c, "Unauthorized")
		return
	}

	password := strings.TrimSpace(c.PostForm("password"))
	password1 := strings.TrimSpace(c.PostForm("password1"))
	password2 := strings.TrimSpace(c.PostForm("password2"))

	if password == "" || password1 == "" || password2 == "" {
		Json500(c, "All password fields are required")
		return
	}

	if len(password1) < 6 {
		Json500(c, "New password must be at least 6 characters")
		return
	}

	if password1 != password2 {
		Json500(c, "New passwords do not match")
		return
	}

	u, err := db.GetUser(user.Username, password)
	if err != nil {
		log.Errorf("Change password authentication failed for user %s: %v", user.Username, err)
		Json500(c, "Invalid credentials")
		return
	}

	err = db.ChangeUserPassword(u, password1)
	if err != nil {
		log.Errorf("Change password failed for user %s: %v", user.Username, err)
		Json500(c, "Password change failed")
		return
	}

	Json200(c, make(map[string]interface{}))
}

func GlobalOptions(c *gin.Context) {
	categories, err := models.ListCategories()
	if err != nil {
		log.Errorf("GlobalOptions: failed to list categories: %v", err)
		Json500(c, "Failed to load categories")
		return
	}

	data := map[string]any{
		"pagesize":   conf.PageSize(0),
		"categories": categories,
	}

	Json200(c, data)
}
