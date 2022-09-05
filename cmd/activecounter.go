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
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/gin-gonic/gin"
)

func initState() error {
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

	return nil
}

func main() {
	err := initState()
	if err != nil {
		log.Fatalf("err: %v", err)
		os.Exit(1)
	}

	RunServer("", CmdFunMap)
}

func RunServer(svr string, cmdMap []ServerCmdMap) {
	gin.SetMode(config.Server.RunMode)
	engine := gin.New()
	engine.Use(gin.Recovery(), middleware.Logging(), middleware.Cors())
	apiGroup := engine.Group("/api")
	group := reflect.ValueOf(apiGroup)
	for _, fun := range cmdMap {
		RegisterWithCmd(group, fun)
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
		if err := ginSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

func RegisterWithCmd(group reflect.Value, cmd ServerCmdMap) {
	method := group.MethodByName(cmd.Method)
	method.Call([]reflect.Value{
		reflect.ValueOf(cmd.Path),
		reflect.ValueOf(func(ctx *gin.Context) {
			log.Infof("path: %v", cmd.Path)
			var (
				errCode                int
				errMsg, internalErrMsg string
			)

			// TODO 调用处理
			v := reflect.ValueOf(cmd.Handler)
			t := v.Type()
			if t.In(0).Elem().String() != "rpc.Context" {
				panic("XX(*rpc.Context, proto.Message)(proto.Message, error): first in arg must be *rpc.Context")
			}
			// isRawReq := t.In(1).Elem().String() == "ext.RawReq"
			// if !isRawReq {
			// 	if !t.In(1).Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {
			// 		panic("XX(*rpc.Context, proto.Message)(proto.Message, error): second in arg must be proto.Message")
			// 	}
			// }
			// isRawRsp := t.Out(0).Elem().String() == "ext.RawRsp"
			// if !t.Out(0).Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {
			// 	panic("XX(*rpc.Context, proto.Message)(proto.Message, error): first out arg must be proto.Message")
			// }
			// if t.Out(1).String() != "error" {
			// 	panic("XX(*rpc.Context, proto.Message)(proto.Message, error): second out arg must be error")
			// }

			reqT := t.In(1).Elem()
			reqV := reflect.New(reqT)
			handlerRet := v.Call([]reflect.Value{reflect.ValueOf(ctx), reqV})
			if handlerRet[1].IsValid() && !handlerRet[1].IsNil() {
				if x, ok := handlerRet[1].Interface().(*ErrMsg); ok {
					errCode = int(x.ErrCode)
					errMsg = x.ErrMsg
				} else {
					err := handlerRet[1].Interface().(error)
					errCode = IErrSystem
					internalErrMsg = err.Error()
				}
			}
			if handlerRet[0].IsValid() && !handlerRet[0].IsNil() {
				m := jsonpb.Marshaler{
					EmitDefaults: true,
					OrigName:     true,
				}
				tmp, err := m.MarshalToString(handlerRet[0].Interface().(proto.Message))
				if err != nil {
					log.Errorf("MarshalToString err %v", err)
					errCode = IErrResponseMarshalFail
				}
				if tmp == "" {
					tmp = "{}"
				}
				// TODO handle tmp rsp
			}
		}),
	})
}

const (
	IErrSystem = -1
	// IErrRequestBodyReadFail 服务端读取请求数据异常
	IErrRequestBodyReadFail = -2002
	// IErrResponseMarshalFail 服务返回数据序列化失败
	IErrResponseMarshalFail = -2003
	// IPanicProcess 业务处理异常
	IPanicProcess       = -2004
	IExceedMaxCallDepth = -2005
)

type ErrMsg struct {
	ErrCode int32  `protobuf:"varint,1,opt,name=err_code,json=errcode" json:"err_code,omitempty"`
	ErrMsg  string `protobuf:"bytes,2,opt,name=err_msg,json=errmsg" json:"err_msg,omitempty"`
	Hint    string `protobuf:"bytes,3,opt,name=hint" json:"hint,omitempty"`
}
