package route

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"zgia.net/book/conf"
	log "zgia.net/book/logger"
	"zgia.net/book/middleware"
)

var (
	g               *gin.Engine
	srv             *http.Server
	serverWaitGroup sync.WaitGroup
)

func JSON500(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code": http.StatusInternalServerError,
		"msg":  msg,
	})
}

func JSON200(c *gin.Context, msg string, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  msg + " success",
		"data": data,
	})
}

func auth() *jwt.GinJWTMiddleware {
	authMiddleware, err := middleware.JwtMiddleware()
	if err != nil {
		panic("JWT new instance error: " + err.Error())
	}

	auth := g.Group("/auth")
	{
		// cin to get Token
		auth.POST("/login", authMiddleware.LoginHandler)

		// Refresh time can be longer than token timeout
		auth.GET("/refresh_token", authMiddleware.RefreshHandler)

		auth.Use(authMiddleware.MiddlewareFunc())
		{
			auth.GET("/islogined", IsLogined)
			// test
			auth.GET("/hello", HelloHandler)
		}
	}

	return authMiddleware
}

func book(authMiddleware *jwt.GinJWTMiddleware) {
	book := g.Group("/book")
	{
		book.Use(authMiddleware.MiddlewareFunc())
		{
			book.GET("/books", ListBooks)
			book.GET("/:bookid", GetBook)
			book.POST("/:bookid", UpdateBook)
			book.DELETE("/:bookid", DeleteBook)

			book.POST("/:bookid/download", DownloadBook)

			book.GET("/:bookid/chapters", ListChapters)
			book.GET("/:bookid/chapter/:chapterid", GetChapter)
			book.POST("/:bookid/chapter/:chapterid", UpdateChapter)
			book.DELETE("/:bookid/chapter/:chapterid", DeleteChapter)

			book.GET("/:bookid/volumes", ListVolumes)
			book.POST("/:bookid/volume/:volumeid", UpdateVolume)
			book.DELETE("/:bookid/volume/:volumeid", DeleteVolume)
		}
	}
}

func api(authMiddleware *jwt.GinJWTMiddleware) {
	api := g.Group("")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "hello world",
			})
		})

		api.GET("/globaloptions", GlobalOptions)
	}
}

// listen and serve on 0.0.0.0:6767 (for windows "localhost:6767")
func GoHttp() {
	log.Debug("Route init")

	newGin()

	startRoutes()

	startServer()
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

	ginLogCfg := gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s\" %s \"%s\"\n",
				param.ClientIP,
				param.TimeStamp.Format(time.RFC1123),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Latency,
				param.Request.UserAgent(),
				param.ErrorMessage,
			)
		},
		SkipPaths: []string{},
	}

	g.Use(zaplog(log.Log.Logger, &ginLogCfg))

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	g.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  fmt.Sprintf("Server error: %s", err),
			})
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
}

func startRoutes() {
	g.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code": http.StatusNotFound,
			"msg":  fmt.Sprintf("Route(%s) Not Found", c.Request.URL.Path),
		})
	})

	authMiddleware := auth()
	api(authMiddleware)
	book(authMiddleware)

	// test CustomRecovery
	g.GET("/panic", func(c *gin.Context) {
		panic("It is a custom recovery from any panics.")
	})
}

// zaplog for access log
func zaplog(logger *zap.Logger, conf *gin.LoggerConfig) gin.HandlerFunc {
	formatter := conf.Formatter
	notlogged := conf.SkipPaths

	var skip map[string]struct{}

	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			param := gin.LogFormatterParams{
				Request: c.Request,
				Keys:    c.Keys,
			}

			// Stop timer
			param.TimeStamp = time.Now()
			param.Latency = param.TimeStamp.Sub(start)

			param.ClientIP = c.ClientIP()
			param.Method = c.Request.Method
			param.StatusCode = c.Writer.Status()
			param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()

			param.BodySize = c.Writer.Size()

			if raw != "" {
				path = path + "?" + raw
			}

			param.Path = path

			logger.Info(formatter(param))
		}
	}
}

func startServer() {
	ctx, cancel := context.WithCancel(context.Background())
	go handleSignals(ctx)
	defer cancel()

	var listenAddr string
	if conf.Server.Protocol == "unix" {
		listenAddr = conf.Server.HTTPAddr
	} else {
		listenAddr = fmt.Sprintf("%s:%s", conf.Server.HTTPAddr, conf.Server.HTTPPort)
	}
	log.Infof("Available on %s", conf.Server.ExternalURL)

	srv = &http.Server{
		Addr:              listenAddr,
		Handler:           g,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	// 服务连接
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}

	log.Infof("HTTP Listener: %s Closed", listenAddr)
}

func handleSignals(ctx context.Context) {
	signalChannel := make(chan os.Signal, 1)

	signal.Notify(
		signalChannel,
		syscall.SIGHUP,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGTSTP,
	)

	pid := syscall.Getpid()
	for {
		select {
		case sig := <-signalChannel:
			switch sig {
			case syscall.SIGHUP:
				log.Infof("PID: %d. Received SIGHUP.", pid)
			case syscall.SIGUSR1:
				log.Warnf("PID %d. Received SIGUSR1.", pid)
			case syscall.SIGUSR2:
				log.Warnf("PID %d. Received SIGUSR1.", pid)
			case syscall.SIGINT:
				log.Warnf("PID %d. Received SIGINT. Shutting down...", pid)
				doShutdown(ctx)
			case syscall.SIGTERM:
				log.Warnf("PID %d. Received SIGTERM. Shutting down...", pid)
				doShutdown(ctx)
			case syscall.SIGTSTP:
				log.Infof("PID %d. Received SIGTSTP.", pid)
			default:
				log.Infof("PID %d. Received %v.", pid, sig)
			}
		case <-ctx.Done():
			log.Warnf("PID: %d. Context closed %v. Shutting down...", pid, ctx.Err())
			doShutdown(ctx)
			return
		}
	}
}

func doShutdown(ctx context.Context) {
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Error: %s\n", err)
	}
	log.Debug("Server exiting")
}
