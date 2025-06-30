package routes

import (
	"bossfi-backend/api/controllers"
	"bossfi-backend/middleware"

	"github.com/gin-gonic/gin"
)

func LoadRoutes(v1 *gin.RouterGroup) {
	authController := controllers.NewAuthController()
	articleController := controllers.NewArticleController()
	articleCommentController := controllers.NewArticleCommentController()
	// =============================================================================
	// 认证相关路由
	// =============================================================================
	auth := v1.Group("/auth")
	{
		// 公开端点 - 不需要认证
		auth.POST("/nonce", authController.GetNonce)          // 获取签名消息和 nonce
		auth.POST("/login", authController.Login)             // 钱包签名登录
		auth.POST("/test-token", authController.GetTestToken) // 获取测试用token（仅开发测试）

		// 需要认证的端点
		auth.Use(middleware.AuthMiddleware())
		auth.GET("/profile", authController.GetProfile) // 获取用户信息
		auth.POST("/logout", authController.Logout)     // 用户登出
	}

	// =============================================================================
	// 文章相关路由
	// =============================================================================
	articles := v1.Group("/articles")
	{
		// 公开端点 - 不需要认证
		articles.GET("", articleController.GetArticles)    // 获取文章列表
		articles.GET("/:id", articleController.GetArticle) // 获取文章详情

		// 需要认证的端点
		articles.Use(middleware.AuthMiddleware())
		articles.POST("", articleController.CreateArticle)              // 创建文章
		articles.PUT("/:id", articleController.UpdateArticle)           // 更新文章
		articles.DELETE("/:id", articleController.DeleteArticle)        // 删除文章
		articles.POST("/:id/like", articleController.LikeArticle)       // 点赞文章
		articles.DELETE("/:id/unlike", articleController.UnlikeArticle) // 取消点赞文章
	}

	comments := v1.Group("/comments")
	{
		comments.GET("", articleCommentController.GetComments) // 获取评论列表
		comments.Use(middleware.AuthMiddleware())
		comments.POST("", articleCommentController.CreateComment)        // 创建评论
		comments.POST("/:id/like", articleCommentController.LikeComment) // 点赞评论

	}

}
