package routes

import (
	"bossfi-backend/api/controllers"
	"bossfi-backend/middleware"

	"github.com/gin-gonic/gin"
)

func LoadRoutes(v1 *gin.RouterGroup) {
	authController := controllers.NewAuthController()

	auth := v1.Group("/auth")
	{
		// 公开端点 - 不需要认证
		auth.POST("/nonce", authController.GetNonce) // 获取签名消息和 nonce
		auth.POST("/login", authController.Login)    // 钱包签名登录

		// 需要认证的端点
		auth.Use(middleware.AuthMiddleware())
		auth.GET("/profile", authController.GetProfile) // 获取用户信息
		auth.POST("/logout", authController.Logout)     // 用户登出
	}

}
