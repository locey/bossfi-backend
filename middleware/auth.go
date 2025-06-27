package middleware

import (
	"net/http"
	"strings"

	"bossfi-backend/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// 解析 JWT
		claims, err := utils.ParseJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("wallet_address", claims.WalletAddress)

		c.Next()
	}
}

// GetUserIDFromContext 从上下文中获取用户 ID
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	if id, ok := userID.(string); ok {
		return id, true
	}

	return "", false
}

// GetWalletAddressFromContext 从上下文中获取钱包地址
func GetWalletAddressFromContext(c *gin.Context) (string, bool) {
	walletAddress, exists := c.Get("wallet_address")
	if !exists {
		return "", false
	}

	if address, ok := walletAddress.(string); ok {
		return address, true
	}

	return "", false
}
