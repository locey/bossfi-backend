package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"bossfi-blockchain-backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware 自定义恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := debug.Stack()

				// 记录panic日志
				logger.GetLogger().Error("Panic recovered",
					zap.String("error", fmt.Sprintf("%v", err)),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.String("user_agent", c.Request.UserAgent()),
					zap.String("stack", string(stack)),
				)

				// 返回500错误
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": "Internal server error",
				})

				// 终止请求处理
				c.Abort()
			}
		}()

		c.Next()
	}
}
