// @title BossFi Backend API
// @version 1.0.0
// @description BossFi区块链后端API服务，支持钱包登录、用户管理、区块链数据同步等功能
// @termsOfService https://www.bossfi.com/terms
// @contact.name BossFi Team
// @contact.url https://www.bossfi.com
// @contact.email support@bossfi.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description 输入Bearer {token}
package main

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"bossfi-backend/api/routes"
	"bossfi-backend/config"
	"bossfi-backend/db/database"
	"bossfi-backend/db/redis"
	cron "bossfi-backend/schedule"

	// 导入docs包以初始化swagger
	"bossfi-backend/docs"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 初始化日志
	initLogger()

	logrus.Info("Starting BossFi Backend...")

	// 初始化配置
	config.Init()
	logrus.Info("Configuration loaded successfully")

	// 设置 Swagger 信息
	swaggerURL := os.Getenv("SWAGGER_URL")
	if swaggerURL == "" {
		swaggerURL = "http://localhost:" + config.AppConfig.Server.Port
	}
	// 移除可能的协议前缀
	swaggerURL = strings.TrimPrefix(strings.TrimPrefix(swaggerURL, "http://"), "https://")
	docs.SwaggerInfo.Host = swaggerURL

	// 生成 Swagger 文档
	generateSwaggerDocs()

	// 设置 Gin 模式
	gin.SetMode(config.AppConfig.Server.GinMode)

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化 Redis
	if err := redis.InitRedis(); err != nil {
		logrus.Fatalf("Failed to initialize Redis: %v", err)
	}

	// 初始化定时任务调度器
	scheduler := cron.NewScheduler()
	if err := scheduler.Start(); err != nil {
		logrus.Errorf("Failed to start scheduler: %v", err)
	}

	// 设置路由
	router := routes.SetupRoutes()

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    "0.0.0.0:" + config.AppConfig.Server.Port,
		Handler: router,
	}

	// 启动服务器
	go func() {
		logrus.Infof("Server starting on port %s", config.AppConfig.Server.Port)
		logrus.Infof("Swagger UI available at: http://localhost:%s/swagger/index.html", config.AppConfig.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 优雅关闭
	gracefulShutdown(srv, scheduler)
}

// initLogger 初始化日志配置
func initLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFunc:  "caller",
		},
	})

	// 设置日志级别
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	// 设置报告调用者信息
	logrus.SetReportCaller(true)
}

// generateSwaggerDocs 生成 Swagger 文档
func generateSwaggerDocs() {
	logrus.Info("Generating Swagger documentation...")

	// 检查是否安装了 swag 命令
	swagPath := "swag"
	if _, err := exec.LookPath("swag"); err != nil {
		// 尝试从 GOPATH 查找
		goPath := os.Getenv("GOPATH")
		if goPath != "" {
			swagPath = goPath + "/bin/swag.exe"
			if _, err := os.Stat(swagPath); os.IsNotExist(err) {
				swagPath = goPath + "/bin/swag"
			}
		}

		// 如果还是找不到，尝试安装
		if _, err := os.Stat(swagPath); os.IsNotExist(err) {
			logrus.Warn("swag command not found, attempting to install...")
			installCmd := exec.Command("go", "install", "github.com/swaggo/swag/cmd/swag@latest")
			if err := installCmd.Run(); err != nil {
				logrus.Errorf("Failed to install swag: %v", err)
				logrus.Warn("Please install swag manually: go install github.com/swaggo/swag/cmd/swag@latest")
				return
			}
			logrus.Info("swag installed successfully")
			// 重新设置路径
			if goPath != "" {
				swagPath = goPath + "/bin/swag.exe"
				if _, err := os.Stat(swagPath); os.IsNotExist(err) {
					swagPath = goPath + "/bin/swag"
				}
			}
		}
	}

	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		logrus.Errorf("Failed to get working directory: %v", err)
		return
	}

	// 生成 Swagger 文档
	cmd := exec.Command(swagPath, "init", "--dir", "./api", "--output", "./docs", "--generalInfo", "main.go")

	// 设置工作目录为项目根目录
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("Failed to generate swagger docs: %v", err)
		logrus.Errorf("Command output: %s", string(output))
		logrus.Errorf("Working directory: %s", workDir)
		logrus.Errorf("Swag path: %s", swagPath)
		return
	}

	logrus.Info("Swagger documentation generated successfully")
	logrus.Debugf("Swagger generation output: %s", string(output))
}

// gracefulShutdown 优雅关闭
func gracefulShutdown(srv *http.Server, scheduler *cron.Scheduler) {
	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal, 1)
	// kill (无参数) 默认发送 syscall.SIGTERM
	// kill -2 发送 syscall.SIGINT
	// kill -9 发送 syscall.SIGKILL 但不能捕获，所以不需要添加
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down server...")

	// 5 秒的超时时间用于完成剩余的请求处理
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭定时任务调度器
	if scheduler != nil {
		scheduler.Stop()
	}

	// 关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
	}

	// 关闭数据库连接
	if database.DB != nil {
		if sqlDB, err := database.DB.DB(); err == nil {
			sqlDB.Close()
			logrus.Info("Database connection closed")
		}
	}

	// 关闭 Redis 连接
	if redis.RedisClient != nil {
		redis.RedisClient.Close()
		logrus.Info("Redis connection closed")
	}

	logrus.Info("Server exited")
}
