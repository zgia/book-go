package middleware

import (
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"

	"zgia.net/book/internal/conf"
	"zgia.net/book/internal/db"
	log "zgia.net/book/internal/logger"
)

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

const JwtIdentityKey = "book_zgia_net_user"

type UserResult struct {
	Id          int64          `json:"id"`
	Username    string         `json:"username"`
	Realname    string         `json:"realname"`
	Mobile      string         `json:"mobile"`
	Permissions map[string]any `json:"permissions"`
}

var user *UserResult

// JwtMiddleware returns a new jwt instance
func JwtMiddleware() (*jwt.GinJWTMiddleware, error) {
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "book.neo.zgia.net",
		Key:         []byte(conf.HTTP.JwtSecretKey),
		IdentityKey: JwtIdentityKey,

		PayloadFunc:     payloadFunc,
		IdentityHandler: identityHandler,
		Authenticator:   authenticator,
		Authorizator:    authorizator,
		Unauthorized:    unauthorized,
		LoginResponse:   loginResponse,

		Timeout: 365 * 24 * time.Hour,
	})
}

// 检查前端传来的用户数据是否正确，验证正确后，再调用 @see payloadFunc
func authenticator(c *gin.Context) (interface{}, error) {
	var loginVals login
	if err := c.ShouldBind(&loginVals); err != nil {
		return "", jwt.ErrMissingLoginValues
	}
	username := loginVals.Username
	// md5 format
	password := loginVals.Password

	// Get user from DB
	if u, uerr := db.GetUser(username, password); uerr == nil {
		user = &UserResult{
			Id:          u.Id,
			Username:    u.Username,
			Realname:    u.Realname,
			Mobile:      u.Mobile,
			Permissions: map[string]any{"books": 1, "chapters": 1, "content": 1},
		}
		return user, nil
	}

	return nil, jwt.ErrFailedAuthentication
}

// 用户登陆成功后，可以把更多信息写到Token里
// @see authMiddleware.LoginHandler
func payloadFunc(data interface{}) jwt.MapClaims {
	if v, ok := data.(*UserResult); ok {

		return jwt.MapClaims{
			JwtIdentityKey: v,
		}
	}
	return jwt.MapClaims{}
}

// 用于访问某个URL时的身份校验
// @see authorizator
func identityHandler(c *gin.Context) interface{} {
	claims := jwt.ExtractClaims(c)

	user := claims[JwtIdentityKey].(map[string]interface{})

	return &UserResult{
		Id:          int64(user["id"].(float64)),
		Username:    user["username"].(string),
		Realname:    user["realname"].(string),
		Mobile:      user["mobile"].(string),
		Permissions: user["permissions"].(map[string]any),
	}
}

// 用于访问某个URL时的身份校验
// @see identityHandler
func authorizator(data interface{}, c *gin.Context) bool {
	v, ok := data.(*UserResult)

	return ok && v.Id != 0 && v.Username != ""
}

func unauthorized(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"code": code, "msg": message})
}

func loginResponse(c *gin.Context, code int, token string, expire time.Time) {
	claims := jwt.ExtractClaims(c)
	log.Infof("claims: %v, user: %#v", claims, user)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "jwt.loginResponse",
		"data": map[string]any{
			"token":  token,
			"expire": expire.Format(time.RFC3339),
			"user":   user,
		},
	})
}
