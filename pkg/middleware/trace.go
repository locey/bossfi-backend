package middleware

import (
	"bossfi-blockchain-backend/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	TraceIDKey    = "trace_id"
	TraceIDHeader = "X-Trace-ID"
)

// TraceMiddleware 请求追踪中间件
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 获取或生成TraceID
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			// 如果前端没有提供TraceID，后端生成一个
			traceID = "server_" + uuid.New().String()
		}

		// 设置到上下文中
		c.Set(TraceIDKey, traceID)

		// 设置响应头
		c.Header(TraceIDHeader, traceID)

		// 记录请求开始日志
		logger.GetLogger().Info("Request started",
			zap.String("trace_id", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("raw_query", c.Request.URL.RawQuery),
		)

		// 处理请求
		c.Next()

		// 请求结束，记录响应日志
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 构建日志字段
		fields := []zap.Field{
			zap.String("trace_id", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("status_code", statusCode),
			zap.Duration("latency", latency),
			zap.Int("response_size", c.Writer.Size()),
		}

		// 如果有用户ID，添加到日志中
		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.String("user_id", userID.(string)))
		}

		// 根据状态码选择日志级别
		switch {
		case statusCode >= 500:
			logger.GetLogger().Error("Request completed with server error", fields...)
		case statusCode >= 400:
			logger.GetLogger().Warn("Request completed with client error", fields...)
		default:
			logger.GetLogger().Info("Request completed successfully", fields...)
		}
	}
}

// GetTraceID 从gin.Context中获取TraceID
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(TraceIDKey); exists {
		return traceID.(string)
	}
	return ""
}
