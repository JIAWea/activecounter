package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"activecounter/internal/app"
	"activecounter/pkg/log"
)

func main() {
	err := log.NewLogger(&log.Config{}, log.InstanceZapLogger)
	if err != nil {
		os.Exit(1)
	}

	err = app.InitConfig()
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	err = app.InitModel()
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	engine := app.InitRoutes()

	addr := config.GetHostPort()
	log.Info("Server is listening on", addr)

	ginSrv := &http.Server{
		Addr:           addr,
		Handler:        engine,
		ReadTimeout:    config.ServerSetting.ReadTimeout,
		WriteTimeout:   config.ServerSetting.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// ginSrv.ListenAndServe()
	go func() {
		// 服务启动 server connections
		if err = ginSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("start server err: ", err)
		}
	}()

	// Wait for interrupt signal to gracefully to shutdown the server with
	// a timeout of 5 seconds
	quit := make(chan os.Signal)
	// kill (no param) default send syscal.SIGTERM <- kill无参数信号默认为SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	for {
		signalVaule := <-quit
		switch signalVaule {
		case syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM:
			// Shutdown server
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := ginSrv.Shutdown(ctx); err != nil {
				log.Fatal("服务关闭错误, err:", err)
			}
			log.Info("Server shutdown...")
			return
		case syscall.SIGHUP:
			// reload
		default:
			return
		}
	}
}
