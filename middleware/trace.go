package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	TraceIDHeader = "X-Trace-ID"
	TraceIDKey    = "trace_id"
)

// generateTraceID 生成随机的TraceID
func generateTraceID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机生成失败，使用时间戳作为备选
		return fmt.Sprintf("trace_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// TraceIDMiddleware 提取或生成 TraceID，并设置到请求上下文中
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取 TraceID
		traceID := c.GetHeader(TraceIDHeader)

		// 如果没有提供 TraceID，则生成一个新的
		if traceID == "" {
			traceID = generateTraceID()
		}

		// 设置到 Gin 上下文
		c.Set(TraceIDKey, traceID)

		// 设置到标准 context
		ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		// 添加到响应头
		c.Header(TraceIDHeader, traceID)

		// 为当前请求设置日志字段
		entry := logrus.WithField("trace_id", traceID)
		c.Set("logger", entry)

		c.Next()
	}
}

// GetTraceIDFromContext 从 Gin 上下文中获取 TraceID
func GetTraceIDFromContext(c *gin.Context) (string, bool) {
	traceID, exists := c.Get(TraceIDKey)
	if !exists {
		return "", false
	}

	if id, ok := traceID.(string); ok {
		return id, true
	}

	return "", false
}

// GetLoggerFromContext 从 Gin 上下文中获取带 TraceID 的 logger
func GetLoggerFromContext(c *gin.Context) *logrus.Entry {
	if logger, exists := c.Get("logger"); exists {
		if entry, ok := logger.(*logrus.Entry); ok {
			return entry
		}
	}

	// 如果没有找到，创建一个默认的
	if traceID, exists := GetTraceIDFromContext(c); exists {
		return logrus.WithField("trace_id", traceID)
	}

	return logrus.NewEntry(logrus.StandardLogger())
}

// GetTraceIDFromStandardContext 从标准 context 中获取 TraceID
func GetTraceIDFromStandardContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}
