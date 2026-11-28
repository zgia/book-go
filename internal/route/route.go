package route

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"zgia.net/book/internal/conf"
	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/middleware"
	models "zgia.net/book/internal/models"
)

var g *gin.Engine

func callerFuncName() string {
	counter, _, _, success := runtime.Caller(2)

	if !success {
		return "success"
	}

	return strings.Replace(runtime.FuncForPC(counter).Name(), "zgia.net/book/", "", 1)
}

func Json401(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code": http.StatusUnauthorized,
		"msg":  msg,
	})
}

func Json404(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, gin.H{
		"code": http.StatusNotFound,
		"msg":  msg,
	})
}

func Json500(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code": http.StatusInternalServerError,
		"msg":  msg,
	})
}

func Json200(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  callerFuncName(),
		"data": data,
	})
}

func auth() *jwt.GinJWTMiddleware {
	authMiddleware, err := middleware.JwtMiddleware()
	if err != nil {
		panic("JWT new instance error: " + err.Error())
	}

	return authMiddleware
}

func routerAuth(authMiddleware *jwt.GinJWTMiddleware) {

	auth := g.Group("/auth")
	{
		// cin to get Token
		auth.POST("/login", authMiddleware.LoginHandler)

		// Refresh time can be longer than token timeout
		auth.GET("/refresh_token", authMiddleware.RefreshHandler)

		auth.Use(authMiddleware.MiddlewareFunc())
		{
			auth.GET("/user", UserInfo)
			auth.POST("/changepassword", ChangePassword)
			// test
			auth.GET("/hello", HelloHandler)
		}
	}
}

func routerConfig(authMiddleware *jwt.GinJWTMiddleware) {
	config := g.Group("/cnf")
	{
		config.Use(authMiddleware.MiddlewareFunc())
		{
			config.GET("/categories", ListCategories)
			config.POST("/category/:category_id", UpdateCategory)
			config.DELETE("/category/:category_id", DeleteCategory)
		}
	}
}

func routerBook(authMiddleware *jwt.GinJWTMiddleware) {
	book := g.Group("/book")
	{
		book.Use(authMiddleware.MiddlewareFunc())
		{
			book.GET("/books", ListBooks)
			book.GET("/books/size", GetBooksSize)
			book.GET("/:book_id", GetBook)
			book.POST("/:book_id", UpdateBook)
			book.DELETE("/:book_id", DeleteBook)

			book.POST("/:book_id/download", DownloadBook)
			book.POST("/:book_id/finish", FinishBook)

			book.GET("/:book_id/chapters", ListChapters)
			book.GET("/:book_id/chapter/:chapter_id", GetChapter)
			book.POST("/:book_id/chapter/:chapter_id", UpdateChapter)
			book.DELETE("/:book_id/chapter/:chapter_id", DeleteChapter)
			book.GET("/:book_id/volchapters/:volume_id", GetVolumeChapters)

			book.GET("/:book_id/volumes", ListVolumes)
			book.GET("/:book_id/volume/:volume_id", GetVolume)
			book.POST("/:book_id/volume/:volume_id", UpdateVolume)
			book.DELETE("/:book_id/volume/:volume_id", DeleteVolume)
		}
	}
}

func routerAuthor(authMiddleware *jwt.GinJWTMiddleware) {
	author := g.Group("/author")
	{
		author.Use(authMiddleware.MiddlewareFunc())
		{
			author.GET("/index", ListAuthors)
			author.GET("/:author_id", GetAuthor)
			author.POST("/:author_id", UpdateAuthor)
			author.DELETE("/:author_id", DeleteAuthor)
		}
	}
}

func routerApi(authMiddleware *jwt.GinJWTMiddleware) {
	g.GET("/", func(c *gin.Context) {
		Json200(c, nil)
	})

	g.GET("/search", SearchBooks)

	g.GET("/options", GlobalOptions)

	g.GET("/panic", func(c *gin.Context) {
		panic("It is a custom recovery from any panics.")
	})
}

// listen and serve on 0.0.0.0:6767 (for windows "localhost:6767")
func GoHttp() error {
	log.Debug("Route init")

	newGin()

	startRoutes()

	return models.StartServer(g)
}

func newGin() {
	if conf.IsProdMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	g = gin.New()

	// pprof
	if conf.Server.EnablePprof {
		pprof.Register(g)
	}

	// GZIP
	if conf.Server.EnableGzip {
		g.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	// CORS
	corsCfg := cors.DefaultConfig()
	if len(conf.HTTP.AccessControlAllowOrigin) > 0 {
		corsCfg.AllowOrigins = strings.Split(conf.HTTP.AccessControlAllowOrigin, ",")
	} else {
		corsCfg.AllowOrigins = []string{"*"}
	}
	log.Debugf("CORS config: %#v", corsCfg)
	g.Use(cors.New(corsCfg))

	g.Use(log.ZapGinLogger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	g.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			Json500(c, fmt.Sprintf("Server error: %s", err))
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
}

func startRoutes() {
	g.NoRoute(func(c *gin.Context) {
		Json404(c, fmt.Sprintf("Route(%s) Not Found", c.Request.URL.Path))
	})

	authMiddleware := auth()
	routerAuth(authMiddleware)
	routerApi(authMiddleware)
	routerBook(authMiddleware)
	routerAuthor(authMiddleware)
	routerConfig(authMiddleware)
}
