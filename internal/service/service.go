package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func SayHello(ctx *gin.Context, req interface{}) (interface{}, error) {
	fmt.Printf("ctx: %v, req: %v", ctx, req)

	ctx.JSON(200, "success")
	return nil, nil
}

func SayHelloV2(ctx *gin.Context, req interface{}) (interface{}, error) {
	return nil, nil
}
