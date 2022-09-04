package app

import (
	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	gin.SetMode(config.ServerSetting.RunMode)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.Logging(), middleware.Cors())

	v1Group := r.Group("api/v1")
	v1Group.Use(middleware.Auth()).Use(middleware.RbacCheck())
	v1Group.POST("/1", user.RegisterApi)
	v1Group.GET("/2", user.AllUserInfoApi)

	return r
}
