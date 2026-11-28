package models

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"zgia.net/book/internal/conf"
	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/modules/translation"
)

var srv *http.Server

func StartServer(g *gin.Engine) error {
	ctx, cancel := context.WithCancel(context.Background())
	go handleSignals(ctx)
	defer cancel()

	var listenAddr string
	if conf.Server.Protocol == "unix" {
		listenAddr = conf.Server.HTTPAddr
	} else {
		listenAddr = fmt.Sprintf("%s:%s", conf.Server.HTTPAddr, conf.Server.HTTPPort)
	}
	log.Infof(translation.Lang.Tr("available_server", conf.Server.ExternalURL))

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
		return err
	}

	log.Infof("HTTP Listener: %s Closed", listenAddr)
	return nil
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
