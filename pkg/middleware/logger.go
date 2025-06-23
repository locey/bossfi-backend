package middleware

import (
	"time"

	"bossfi-blockchain-backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerMiddleware 自定义日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 使用Zap记录请求日志
		fields := []zap.Field{
			zap.String("client_ip", param.ClientIP),
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.String("protocol", param.Request.Proto),
			zap.Int("status_code", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.Int("body_size", param.BodySize),
		}

		// 根据状态码选择日志级别
		switch {
		case param.StatusCode >= 400 && param.StatusCode < 500:
			// 4xx 客户端错误
			logger.GetLogger().Warn("HTTP Request - Client Error", fields...)
		case param.StatusCode >= 500:
			// 5xx 服务器错误
			logger.GetLogger().Error("HTTP Request - Server Error", fields...)
		case param.StatusCode >= 300 && param.StatusCode < 400:
			// 3xx 重定向
			logger.GetLogger().Info("HTTP Request - Redirect", fields...)
		default:
			// 2xx 成功
			logger.GetLogger().Info("HTTP Request - Success", fields...)
		}

		// 返回空字符串，因为我们已经用Zap记录了
		return ""
	})
}

// AccessLogMiddleware 访问日志中间件（更详细的版本）
func AccessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 记录请求信息
		param := gin.LogFormatterParams{
			Request:      c.Request,
			TimeStamp:    time.Now(),
			Latency:      time.Since(start),
			ClientIP:     c.ClientIP(),
			Method:       c.Request.Method,
			StatusCode:   c.Writer.Status(),
			ErrorMessage: c.Errors.ByType(gin.ErrorTypePrivate).String(),
			BodySize:     c.Writer.Size(),
		}

		if raw != "" {
			path = path + "?" + raw
		}
		param.Path = path

		// 构建日志字段
		fields := []zap.Field{
			zap.String("client_ip", param.ClientIP),
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.String("protocol", c.Request.Proto),
			zap.Int("status_code", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("referer", c.Request.Referer()),
			zap.Int("body_size", param.BodySize),
		}

		// 添加错误信息（如果有）
		if param.ErrorMessage != "" {
			fields = append(fields, zap.String("error", param.ErrorMessage))
		}

		// 添加请求头信息（可选）
		if c.Request.Header.Get("Authorization") != "" {
			fields = append(fields, zap.Bool("authenticated", true))
		}

		// 根据状态码和路径选择日志级别和消息
		switch {
		case param.StatusCode >= 500:
			logger.GetLogger().Error("HTTP Request Failed", fields...)
		case param.StatusCode >= 400:
			if param.StatusCode == 404 {
				logger.GetLogger().Warn("HTTP Request Not Found", fields...)
			} else {
				logger.GetLogger().Warn("HTTP Request Client Error", fields...)
			}
		case param.StatusCode >= 300:
			logger.GetLogger().Info("HTTP Request Redirect", fields...)
		default:
			logger.GetLogger().Info("HTTP Request Success", fields...)
		}
	}
}
