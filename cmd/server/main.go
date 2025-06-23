// @title BossFi Backend API
// @version 1.0
// @description BossFi区块链招聘论坛后端API
// @termsOfService https://bossfi.io/terms

// @contact.name BossFi Team
// @contact.url https://bossfi.io
// @contact.email support@bossfi.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

//go:generate swag init -g main.go --output ../../docs

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	_ "bossfi-blockchain-backend/docs"
	"bossfi-blockchain-backend/internal/api"
	"bossfi-blockchain-backend/internal/domain/post"
	"bossfi-blockchain-backend/internal/domain/reply"
	"bossfi-blockchain-backend/internal/domain/stake"
	"bossfi-blockchain-backend/internal/domain/user"
	"bossfi-blockchain-backend/internal/repository"
	"bossfi-blockchain-backend/internal/service"
	"bossfi-blockchain-backend/pkg/config"
	"bossfi-blockchain-backend/pkg/database"
	"bossfi-blockchain-backend/pkg/logger"
	"bossfi-blockchain-backend/pkg/redis"
)

// generateSwaggerDocs 自动生成 Swagger 文档
func generateSwaggerDocs() {
	log.Println("Checking Swagger documentation...")

	// 尝试自动生成 Swagger 文档
	cmd := exec.Command("swag", "init", "-g", "cmd/server/main.go", "--output", "./docs")

	// 执行命令
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to generate Swagger docs automatically: %v", err)
		log.Println("Please run manually: swag init -g cmd/server/main.go --output ./docs")

		// 检查是否存在已生成的文档
		if _, err := os.Stat("docs/swagger.json"); os.IsNotExist(err) {
			log.Println("Warning: No Swagger documentation found!")
		} else {
			log.Println("Using existing Swagger documentation")
		}
	} else {
		log.Println("Swagger documentation generated successfully")
	}
}

func main() {
	// 自动生成 Swagger 文档
	generateSwaggerDocs()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化日志
	appLogger, err := logger.New(cfg)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer appLogger.Sync()

	// 连接数据库
	db, err := database.Connect(cfg)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// 自动迁移数据库
	if err := db.AutoMigrate(
		&user.User{},
		&post.Post{},
		&reply.PostReply{},
		&stake.Stake{},
	); err != nil {
		appLogger.Fatal("Failed to migrate database", zap.Error(err))
	}
	appLogger.Info("Database migrated successfully")

	// 连接Redis
	redisClient, err := redis.Connect(cfg)
	if err != nil {
		appLogger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	appLogger.Info("Connected to Redis successfully")

	// 初始化仓储层
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	stakeRepo := repository.NewStakeRepository(db)

	// 初始化服务层
	userService := service.NewUserService(userRepo, cfg, appLogger, redisClient)
	postService := service.NewPostService(postRepo, userRepo, cfg, appLogger)
	stakeService := service.NewStakeService(stakeRepo, userRepo, cfg, appLogger)

	// 创建Gin实例
	r := gin.New()

	// 设置路由
	api.SetupRoutes(r, cfg, appLogger, userService, postService, stakeService)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	// 启动服务器
	go func() {
		appLogger.Info("Starting server", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLogger.Info("Shutting down server...")

	// 5秒超时的优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	appLogger.Info("Server exited")
}
