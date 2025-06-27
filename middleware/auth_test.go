package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"bossfi-backend/config"
	"bossfi-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 设置测试配置
	os.Setenv("JWT_SECRET", "test-secret-key")
	os.Setenv("JWT_EXPIRE_HOURS", "1")

	// 初始化配置
	config.Init()

	// 运行测试
	code := m.Run()
	os.Exit(code)
}

func TestAuthMiddleware(t *testing.T) {
	// 创建测试用户数据
	userID := uuid.New().String()
	walletAddress := "0x1234567890123456789012345678901234567890"

	// 生成有效的 JWT token
	validToken, err := utils.GenerateJWT(userID, walletAddress)
	require.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectUserID   bool
		expectWallet   bool
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectUserID:   true,
			expectWallet:   true,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
			expectWallet:   false,
		},
		{
			name:           "invalid authorization format - no Bearer",
			authHeader:     validToken,
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
			expectWallet:   false,
		},
		{
			name:           "invalid authorization format - wrong prefix",
			authHeader:     "Basic " + validToken,
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
			expectWallet:   false,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
			expectWallet:   false,
		},
		{
			name:           "malformed token",
			authHeader:     "Bearer malformed",
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
			expectWallet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试路由
			router := gin.New()
			router.Use(AuthMiddleware())

			// 添加测试端点
			router.GET("/test", func(c *gin.Context) {
				userID, userExists := GetUserIDFromContext(c)
				walletAddr, walletExists := GetWalletAddressFromContext(c)

				c.JSON(http.StatusOK, gin.H{
					"user_id":       userID,
					"user_exists":   userExists,
					"wallet":        walletAddr,
					"wallet_exists": walletExists,
				})
			})

			// 创建请求
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// 发送请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证响应状态
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// 解析响应
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// 验证用户信息是否正确设置
				if tt.expectUserID {
					assert.True(t, response["user_exists"].(bool))
					assert.Equal(t, userID, response["user_id"].(string))
				} else {
					assert.False(t, response["user_exists"].(bool))
				}

				if tt.expectWallet {
					assert.True(t, response["wallet_exists"].(bool))
					assert.Equal(t, walletAddress, response["wallet"].(string))
				} else {
					assert.False(t, response["wallet_exists"].(bool))
				}
			}
		})
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	// 创建过期的 token
	userID := uuid.New().String()
	walletAddress := "0x1234567890123456789012345678901234567890"

	// 临时修改配置以创建即时过期的 token
	originalExpireHours := config.AppConfig.JWT.ExpireHours
	config.AppConfig.JWT.ExpireHours = -1 * time.Hour // 设置为负值使其立即过期

	expiredToken, err := utils.GenerateJWT(userID, walletAddress)
	require.NoError(t, err)

	// 恢复原始配置
	config.AppConfig.JWT.ExpireHours = originalExpireHours

	// 创建测试路由
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 创建请求
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)

	// 发送请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证过期 token 被拒绝
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetUserIDFromContext(t *testing.T) {
	// 创建 Gin 上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	tests := []struct {
		name         string
		setValue     interface{}
		expectID     string
		expectExists bool
	}{
		{
			name:         "valid user ID",
			setValue:     "test-user-id",
			expectID:     "test-user-id",
			expectExists: true,
		},
		{
			name:         "invalid type",
			setValue:     123,
			expectID:     "",
			expectExists: false,
		},
		{
			name:         "no value set",
			setValue:     nil,
			expectID:     "",
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置上下文
			c.Keys = make(map[string]interface{})

			// 设置值（如果有）
			if tt.setValue != nil {
				c.Set("user_id", tt.setValue)
			}

			// 测试函数
			userID, exists := GetUserIDFromContext(c)

			assert.Equal(t, tt.expectExists, exists)
			assert.Equal(t, tt.expectID, userID)
		})
	}
}

func TestGetWalletAddressFromContext(t *testing.T) {
	// 创建 Gin 上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	tests := []struct {
		name         string
		setValue     interface{}
		expectAddr   string
		expectExists bool
	}{
		{
			name:         "valid wallet address",
			setValue:     "0x1234567890123456789012345678901234567890",
			expectAddr:   "0x1234567890123456789012345678901234567890",
			expectExists: true,
		},
		{
			name:         "invalid type",
			setValue:     123,
			expectAddr:   "",
			expectExists: false,
		},
		{
			name:         "no value set",
			setValue:     nil,
			expectAddr:   "",
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置上下文
			c.Keys = make(map[string]interface{})

			// 设置值（如果有）
			if tt.setValue != nil {
				c.Set("wallet_address", tt.setValue)
			}

			// 测试函数
			walletAddr, exists := GetWalletAddressFromContext(c)

			assert.Equal(t, tt.expectExists, exists)
			assert.Equal(t, tt.expectAddr, walletAddr)
		})
	}
}
