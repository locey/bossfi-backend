package middleware

import (
	"fmt"
	"net/http"
	"time"

	"bossfi-blockchain-backend/pkg/config"
	"bossfi-blockchain-backend/pkg/redis"

	"github.com/gin-gonic/gin"
	redisV8 "github.com/go-redis/redis/v8"
)

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get()
		client := redis.GetClient()

		// 获取客户端IP
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		// 检查当前请求数
		ctx := c.Request.Context()
		current, err := client.Get(ctx, key).Int()
		if err != nil && err != redisV8.Nil {
			// Redis错误，允许请求通过
			c.Next()
			return
		}

		if current >= cfg.Security.RateLimit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		// 增加计数器
		pipe := client.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Minute)
		_, err = pipe.Exec(ctx)
		if err != nil {
			// Redis错误，记录日志但允许请求通过
			// logger.Error("Failed to update rate limit", zap.Error(err))
		}

		c.Next()
	}
}
