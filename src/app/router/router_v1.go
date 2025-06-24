package router

import (
	"bossfi-backend/src/app/controller"
	"bossfi-backend/src/core/config"
	"bossfi-backend/src/core/ctx"
	"github.com/gin-gonic/gin"
)

func Bind(r *gin.Engine, ctx *ctx.Context) {
	api := r.Group("/api/" + config.Conf.App.Version)
	{
		api.GET("/demo/page", controller.PageDemo)
		api.POST("/demo", controller.CreateDemo)
		api.GET("/demo/:id", controller.GetDemoByID)
		api.PUT("/demo/:id", controller.UpdateDemo)
		api.DELETE("/demo/:id", controller.DeleteDemo)
		api.GET("/demo/list", controller.ListDemo)

		api.GET("/evm/get_block_by_num/:block_num", controller.GetBlockByNum)

	}
}
