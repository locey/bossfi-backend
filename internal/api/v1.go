package api

import (
	v1 "bossfi-blockchain-backend/internal/api/v1"
	"bossfi-blockchain-backend/internal/service"
	"bossfi-blockchain-backend/pkg/logger"
	"bossfi-blockchain-backend/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterV1Routes 注册v1版本的路由
func RegisterV1Routes(
	router *gin.RouterGroup,
	userService service.UserService,
	postService service.PostService,
	stakeService service.StakeService,
	logger *logger.Logger,
) {
	v1Group := router.Group("/v1")

	// 初始化处理器
	userHandler := v1.NewUserHandler(userService, logger)
	postHandler := v1.NewPostHandler(postService, logger)
	stakeHandler := v1.NewStakeHandler(stakeService, logger)

	// 认证相关路由
	auth := v1Group.Group("/auth")
	{
		auth.POST("/nonce", userHandler.GenerateNonce)
		auth.POST("/login", userHandler.Login)
	}

	// 用户相关路由
	users := v1Group.Group("/users")
	{

		// 公开路由
		users.GET("/search", userHandler.SearchUsers)
		users.GET("/:user_id/posts", postHandler.GetUserPosts)

		// 需要认证的路由
		authenticated := users.Group("")
		authenticated.Use(middleware.AuthMiddleware())
		{
			authenticated.GET("/profile", userHandler.GetProfile)
			authenticated.PUT("/profile", userHandler.UpdateProfile)
			authenticated.GET("/stats", userHandler.GetUserStats)
		}
	}

	// 帖子相关路由
	posts := v1Group.Group("/posts")
	{
		// 公开路由
		posts.GET("", postHandler.ListPosts)
		posts.GET("/search", postHandler.SearchPosts)
		posts.GET("/popular", postHandler.GetPopularPosts)
		posts.GET("/trending", postHandler.GetTrendingPosts)
		posts.GET("/:id", postHandler.GetPost)

		// 需要认证的路由
		authenticated := posts.Group("")
		authenticated.Use(middleware.AuthMiddleware())
		{
			authenticated.POST("", postHandler.CreatePost)
			authenticated.PUT("/:id", postHandler.UpdatePost)
			authenticated.DELETE("/:id", postHandler.DeletePost)
			authenticated.POST("/:id/publish", postHandler.PublishPost)
			authenticated.POST("/:id/close", postHandler.ClosePost)
			authenticated.POST("/:id/like", postHandler.LikePost)
			authenticated.POST("/:id/unlike", postHandler.UnlikePost)
		}
	}

	// 质押相关路由
	stakes := v1Group.Group("/stakes")
	stakes.Use(middleware.AuthMiddleware())
	{
		stakes.POST("", stakeHandler.CreateStake)
		stakes.GET("/:id", stakeHandler.GetStake)
		stakes.GET("/user/:user_id", stakeHandler.GetUserStakes)
		stakes.POST("/:id/unstake", stakeHandler.RequestUnstake)
		stakes.POST("/:id/complete", stakeHandler.CompleteUnstake)
		stakes.POST("/rewards/claim", stakeHandler.ClaimRewards)
		stakes.GET("/user/:user_id/stats", stakeHandler.GetUserStakeStats)
	}

	// 管理员路由
	admin := v1Group.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.AdminMiddleware())
	{
		// 用户管理
		adminUsers := admin.Group("/users")
		{
			adminUsers.GET("", userHandler.ListUsers)
			adminUsers.GET("/:id", userHandler.GetUserByID)
			adminUsers.DELETE("/:id", userHandler.DeleteUser)
		}

		// 质押管理
		adminStakes := admin.Group("/stakes")
		{
			adminStakes.POST("/distribute", stakeHandler.DistributeRewards)
			adminStakes.GET("/stats", stakeHandler.GetStakeStats)
		}
	}

}
