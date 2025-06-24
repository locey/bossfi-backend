package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"bossfi-blockchain-backend/internal/service"
	"bossfi-blockchain-backend/pkg/config"
	"bossfi-blockchain-backend/pkg/logger"
	"bossfi-blockchain-backend/pkg/middleware"
)

// SetupRoutes 设置路由
func SetupRoutes(
	r *gin.Engine,
	cfg *config.Config,
	logger *logger.Logger,
	userService service.UserService,
	postService service.PostService,
	stakeService service.StakeService,
) {
	// 全局中间件
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimit(cfg))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "bossfi-backend",
			"version": "1.0.0",
		})
	})

	// Swagger文档
	if cfg.App.Debug {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// API路由组
	api := r.Group("/api")

	// 注册v1版本的API路由
	RegisterV1Routes(api, userService, postService, stakeService, logger)
}
