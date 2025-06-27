package routes

import (
	"runtime"
	"time"

	"bossfi-backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// HealthResponse 健康检查响应结构
type HealthResponse struct {
	Status    string            `json:"status" example:"ok"`
	Message   string            `json:"message" example:"BossFi Backend is running"`
	Timestamp string            `json:"timestamp" example:"2025-06-26T15:50:00Z"`
	Version   string            `json:"version" example:"1.0.0"`
	Uptime    string            `json:"uptime" example:"2h30m45s"`
	System    map[string]string `json:"system"`
	TraceID   string            `json:"trace_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

var startTime = time.Now()

// SetupRoutes 设置主路由
func SetupRoutes() *gin.Engine {
	r := gin.New()

	// 添加基础中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 添加 TraceID 中间件 - 必须在所有其他中间件之前
	r.Use(middleware.TraceIDMiddleware())

	// CORS 配置
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // 生产环境应该设置具体的域名
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Trace-ID"}
	config.ExposeHeaders = []string{"X-Trace-ID"} // 允许前端获取 TraceID
	r.Use(cors.New(config))

	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查端点
	r.GET("/health", HealthCheck)

	// API 版本分组
	v1 := r.Group("/api/v1")
	{
		// 设置各模块路由
		LoadRoutes(v1) // 认证相关路由
	}

	return r
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务器是否正常运行，返回系统状态信息和TraceID
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "服务正常"
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	logger := middleware.GetLoggerFromContext(c)
	traceID, _ := middleware.GetTraceIDFromContext(c)

	logger.Info("Health check requested")

	// 计算运行时间
	uptime := time.Since(startTime)

	// 系统信息
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	systemInfo := map[string]string{
		"go_version":    runtime.Version(),
		"arch":          runtime.GOARCH,
		"os":            runtime.GOOS,
		"num_cpu":       string(rune(runtime.NumCPU() + '0')),
		"num_goroutine": string(rune(runtime.NumGoroutine() + '0')),
		"memory_alloc":  formatBytes(memStats.Alloc),
		"memory_sys":    formatBytes(memStats.Sys),
	}

	response := HealthResponse{
		Status:    "ok",
		Message:   "BossFi Backend is running",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0", // 可以从环境变量或构建信息中获取
		Uptime:    uptime.String(),
		System:    systemInfo,
		TraceID:   traceID,
	}

	logger.WithFields(map[string]interface{}{
		"uptime":          uptime.String(),
		"num_goroutine":   runtime.NumGoroutine(),
		"memory_alloc":    formatBytes(memStats.Alloc),
		"response_status": "ok",
	}).Info("Health check completed")

	c.JSON(200, response)
}

// formatBytes 格式化字节数
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return string(rune(bytes)) + " B"
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return string(rune(bytes/div)) + " " + "KMGTPE"[exp:exp+1] + "B"
}
