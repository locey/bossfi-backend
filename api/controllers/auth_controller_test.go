package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"bossfi-backend/api/models"
	"bossfi-backend/api/routes"
	"bossfi-backend/config"
	"bossfi-backend/db/database"
	"bossfi-backend/db/redis"
	"bossfi-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AuthControllerTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *AuthControllerTestSuite) SetupSuite() {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 设置测试环境变量
	os.Setenv("DB_NAME", "bossfi_test")
	os.Setenv("REDIS_DB", "1")
	os.Setenv("JWT_SECRET", "test-secret-key")
	os.Setenv("GIN_MODE", "test")

	// 初始化配置
	config.Init()

	// 初始化数据库
	err := database.InitDB()
	require.NoError(suite.T(), err)

	// 初始化 Redis
	err = redis.InitRedis()
	require.NoError(suite.T(), err)

	// 设置路由
	suite.router = routes.SetupRoutes()
}

func (suite *AuthControllerTestSuite) SetupTest() {
	// 每个测试前清理数据
	database.DB.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	redis.RedisClient.FlushDB(redis.RedisClient.Context())
}

func (suite *AuthControllerTestSuite) TearDownSuite() {
	// 清理测试环境
	if database.DB != nil {
		if sqlDB, err := database.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}
	if redis.RedisClient != nil {
		redis.RedisClient.Close()
	}
}

func (suite *AuthControllerTestSuite) TestGetNonce() {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectMessage  bool
		expectNonce    bool
	}{
		{
			name: "valid wallet address",
			requestBody: map[string]interface{}{
				"wallet_address": "0x1234567890123456789012345678901234567890",
			},
			expectedStatus: http.StatusOK,
			expectMessage:  true,
			expectNonce:    true,
		},
		{
			name: "invalid wallet address",
			requestBody: map[string]interface{}{
				"wallet_address": "invalid-address",
			},
			expectedStatus: http.StatusInternalServerError,
			expectMessage:  false,
			expectNonce:    false,
		},
		{
			name:           "missing wallet address",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectMessage:  false,
			expectNonce:    false,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			// 准备请求
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/auth/nonce", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 发送请求
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response GetNonceResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				if tt.expectMessage {
					assert.NotEmpty(t, response.Message)
					assert.Contains(t, response.Message, "Welcome to BossFi!")
				}

				if tt.expectNonce {
					assert.NotEmpty(t, response.Nonce)
				}
			}
		})
	}
}

func (suite *AuthControllerTestSuite) TestLogin() {
	// 注意：这里只测试请求结构验证，实际的签名验证需要真实的签名数据
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "missing wallet address",
			requestBody: map[string]interface{}{
				"signature": "0x123456",
				"message":   "test message",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing signature",
			requestBody: map[string]interface{}{
				"wallet_address": "0x1234567890123456789012345678901234567890",
				"message":        "test message",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing message",
			requestBody: map[string]interface{}{
				"wallet_address": "0x1234567890123456789012345678901234567890",
				"signature":      "0x123456",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid signature format",
			requestBody: map[string]interface{}{
				"wallet_address": "0x1234567890123456789012345678901234567890",
				"signature":      "invalid-signature",
				"message":        "test message",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			// 准备请求
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 发送请求
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func (suite *AuthControllerTestSuite) TestGetProfile() {
	// 创建测试用户
	testUser := models.User{
		ID:            uuid.New(),
		WalletAddress: "0x1234567890123456789012345678901234567890",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err := database.DB.Create(&testUser).Error
	require.NoError(suite.T(), err)

	// 生成测试 token
	token, err := utils.GenerateJWT(testUser.ID.String(), testUser.WalletAddress)
	require.NoError(suite.T(), err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token format",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			// 准备请求
			req := httptest.NewRequest("GET", "/api/v1/auth/profile", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// 发送请求
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				user, exists := response["user"]
				assert.True(t, exists)

				userMap := user.(map[string]interface{})
				assert.Equal(t, testUser.ID.String(), userMap["id"])
				assert.Equal(t, testUser.WalletAddress, userMap["wallet_address"])
			}
		})
	}
}

func (suite *AuthControllerTestSuite) TestLogout() {
	// 创建测试用户
	testUser := models.User{
		ID:            uuid.New(),
		WalletAddress: "0x1234567890123456789012345678901234567890",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err := database.DB.Create(&testUser).Error
	require.NoError(suite.T(), err)

	// 生成测试 token
	token, err := utils.GenerateJWT(testUser.ID.String(), testUser.WalletAddress)
	require.NoError(suite.T(), err)

	// 设置用户会话
	err = redis.SetUserSession(testUser.ID.String(), token)
	require.NoError(suite.T(), err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			// 准备请求
			req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// 发送请求
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				message, exists := response["message"]
				assert.True(t, exists)
				assert.Equal(t, "Logged out successfully", message)
			}
		})
	}
}

func (suite *AuthControllerTestSuite) TestHealthCheck() {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "ok", response["status"])
	assert.Equal(suite.T(), "BossFi Backend is running", response["message"])
}

func TestAuthControllerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthControllerTestSuite))
}
