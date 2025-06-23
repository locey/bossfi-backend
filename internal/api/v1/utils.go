package v1

import "github.com/gin-gonic/gin"

// getUserID 从gin.Context中获取用户ID
func getUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	return userID.(string)
}
