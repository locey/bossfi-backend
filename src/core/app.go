package core

import (
	appRouter "bossfi-backend/src/app/router"
	"bossfi-backend/src/common"
	"bossfi-backend/src/core/config"
	"bossfi-backend/src/core/db"
	"bossfi-backend/src/core/gin/router"
	"bossfi-backend/src/core/log"
)

func Start(configPath string) {
	// 初始化配置信息
	initConfig(configPath)
	// 初始化日志组件
	initLog()
	// 初始化数据库/Redis
	initDB()
	// 初始化Gin
	initGin()
}

func initConfig(configPath string) {
	common.Ctx.Config = config.InitConfig(configPath)
}

func initLog() {
	common.Ctx.Log = log.InitLog()
}

func initDB() {
	common.Ctx.DB = db.InitPgsql()
	common.Ctx.Redis = db.InitRedis()
}

func initGin() {
	r := router.InitRouter()
	appRouter.Bind(r, &common.Ctx)
	err := r.Run(":" + common.Ctx.Config.App.Port)
	if err != nil {
		panic(err)
	}
}
