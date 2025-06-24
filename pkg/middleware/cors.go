package middleware

import (
	"time"

	"bossfi-blockchain-backend/pkg/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	cfg := config.Get()

	corsConfig := cors.Config{
		AllowOrigins: cfg.Security.CorsOrigins,
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-Trace-ID",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"Cache-Control",
			"Connection",
			"Host",
			"Pragma",
			"Referer",
			"Sec-Fetch-Dest",
			"Sec-Fetch-Mode",
			"Sec-Fetch-Site",
			"User-Agent",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Trace-ID",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	// 如果是开发环境且配置为通配符，添加更宽松的设置
	if cfg.Server.Mode == "debug" && len(cfg.Security.CorsOrigins) == 1 && cfg.Security.CorsOrigins[0] == "*" {
		corsConfig.AllowOriginFunc = func(origin string) bool {
			// 开发环境允许localhost和本地IP
			return true
		}
	}

	return cors.New(corsConfig)
}
