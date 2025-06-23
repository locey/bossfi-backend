package router

import (
	"bossfi-backend/src/app/controller"
	"bossfi-backend/src/common"
	"bossfi-backend/src/core/config"
	"github.com/gin-gonic/gin"
)

func Bind(r *gin.Engine, ctx *common.Context) {
	api := r.Group("/api/" + config.Conf.App.Version)
	{
		api.GET("/test", controller.Test)
		api.GET("/demo/page", controller.PageDemo)
		api.POST("/demo", controller.CreateDemo)
		api.GET("/demo/:id", controller.GetDemoByID)
		api.PUT("/demo/:id", controller.UpdateDemo)
		api.DELETE("/demo/:id", controller.DeleteDemo)
		api.GET("/demo/list", controller.ListDemo)
	}
}
