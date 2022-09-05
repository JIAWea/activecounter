package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"activecounter/internal/config"
	"activecounter/internal/model"
	"activecounter/internal/utils/middleware"
	"activecounter/pkg/log"

	"github.com/gin-gonic/gin"
)

func main() {
	err := log.NewLogger(&log.Config{
		Writers: "stdout",
	}, log.InstanceZapLogger)
	if err != nil {
		fmt.Printf("err: %v", err)
		os.Exit(1)
	}

	err = config.InitConfig()
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	err = model.InitModel()
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	gin.SetMode(config.Server.RunMode)

	engine := gin.New()
	engine.Use(gin.Recovery(), middleware.Logging(), middleware.Cors())

	apiGroup := engine.Group("/api")
	group := reflect.ValueOf(apiGroup)
	for _, v := range CmdFunMap {
		fun := v
		method := group.MethodByName(fun.Method)
		method.Call([]reflect.Value{
			reflect.ValueOf(fun.Path),
			reflect.ValueOf(func(ctx *gin.Context) {
				log.Infof("path: %v", fun.Path)

				// TODO 调用处理
				_, err := fun.Func(ctx, nil)
				if err != nil {
				}
			}),
		})
	}

	ginSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Server.Port),
		Handler: engine,
		// ReadTimeout:    time.Second * 20,
		// WriteTimeout:   time.Second * 20,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Info("Server is listening on: ", config.Server.Port)
		if err = ginSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("start server err: ", err)
		}
	}()

	// Wait for interrupt signal to gracefully to shutdown the server with
	// a timeout of 5 seconds
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	signalValue := <-quit
	switch signalValue {
	case syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := ginSrv.Shutdown(ctx); err != nil {
			log.Fatal("shutdown err:", err)
		}
		defer cancel()
		log.Info("server shutdown...")
	case syscall.SIGHUP:
		// reload
	default:
	}
}
