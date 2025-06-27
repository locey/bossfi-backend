package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bossfi-backend/api/routes"
	"bossfi-backend/config"
	"bossfi-backend/db/database"
	"bossfi-backend/db/redis"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *IntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	// 设置测试环境
	err := setupTestEnv()
	require.NoError(suite.T(), err)

	// 设置路由
	suite.router = routes.SetupRoutes()
}

func (suite *IntegrationTestSuite) SetupTest() {
	// 每个测试前清理数据
	database.DB.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	redis.RedisClient.FlushDB(redis.RedisClient.Context())
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	cleanupTestEnv()
}

func (suite *IntegrationTestSuite) TestCompleteAuthFlow() {
	walletAddress := GetTestWalletAddress()

	// Step 1: 获取 nonce 和签名消息
	suite.T().Run("Step1_GetNonce", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"wallet_address": walletAddress,
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/api/v1/auth/nonce", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "message")
		assert.Contains(t, response, "nonce")
		assert.Contains(t, response["message"].(string), "Welcome to BossFi!")

		// 存储 nonce 和 message 供后续使用
		suite.T().Setenv("TEST_NONCE", response["nonce"].(string))
		suite.T().Setenv("TEST_MESSAGE", response["message"].(string))
	})

	// Step 2: 尝试使用无效签名登录（应该失败）
	suite.T().Run("Step2_LoginWithInvalidSignature", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"wallet_address": walletAddress,
			"signature":      "invalid-signature",
			"message":        GetTestMessage(),
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Step 3: 测试受保护的路由（未认证，应该失败）
	suite.T().Run("Step3_AccessProtectedRouteWithoutAuth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/auth/profile", nil)

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func (suite *IntegrationTestSuite) TestHealthCheck() {
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

func (suite *IntegrationTestSuite) TestAPIRoutes() {
	routes := []struct {
		method     string
		path       string
		expectCode int
	}{
		{"GET", "/health", http.StatusOK},
		{"POST", "/api/v1/auth/nonce", http.StatusBadRequest},    // 缺少请求体
		{"POST", "/api/v1/auth/login", http.StatusBadRequest},    // 缺少请求体
		{"GET", "/api/v1/auth/profile", http.StatusUnauthorized}, // 未认证
		{"POST", "/api/v1/auth/logout", http.StatusUnauthorized}, // 未认证
		{"GET", "/api/v1/auth/invalid", http.StatusNotFound},     // 不存在的路由
	}

	for _, route := range routes {
		suite.T().Run(fmt.Sprintf("%s_%s", route.method, route.path), func(t *testing.T) {
			var req *http.Request
			if route.method == "POST" {
				req = httptest.NewRequest(route.method, route.path, bytes.NewBuffer([]byte("{}")))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(route.method, route.path, nil)
			}

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, route.expectCode, w.Code)
		})
	}
}

func (suite *IntegrationTestSuite) TestRateLimiting() {
	// 测试是否正确处理大量请求
	walletAddress := GetTestWalletAddress()

	requestBody := map[string]interface{}{
		"wallet_address": walletAddress,
	}

	body, _ := json.Marshal(requestBody)

	// 发送多个请求
	successCount := 0
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/nonce", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			successCount++
		}

		// 短暂延迟避免太快的请求
		time.Sleep(10 * time.Millisecond)
	}

	// 验证至少有一些请求成功（具体的速率限制取决于配置）
	assert.Greater(suite.T(), successCount, 0)
}

func (suite *IntegrationTestSuite) TestConcurrentRequests() {
	walletAddress := GetTestWalletAddress()

	requestBody := map[string]interface{}{
		"wallet_address": walletAddress,
	}

	body, _ := json.Marshal(requestBody)

	// 并发发送请求
	ch := make(chan int, 10)
	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest("POST", "/api/v1/auth/nonce", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			ch <- w.Code
		}()
	}

	// 收集结果
	successCount := 0
	for i := 0; i < 10; i++ {
		code := <-ch
		if code == http.StatusOK {
			successCount++
		}
	}

	// 验证大部分请求成功
	assert.Greater(suite.T(), successCount, 5)
}

func (suite *IntegrationTestSuite) TestDatabaseConnection() {
	// 测试数据库连接是否正常
	err := database.DB.Raw("SELECT 1").Error
	assert.NoError(suite.T(), err)
}

func (suite *IntegrationTestSuite) TestRedisConnection() {
	// 测试 Redis 连接是否正常
	err := redis.RedisClient.Ping(redis.RedisClient.Context()).Err()
	assert.NoError(suite.T(), err)

	// 测试基本的读写操作
	testKey := "test:integration"
	testValue := "test-value"

	err = redis.RedisClient.Set(redis.RedisClient.Context(), testKey, testValue, time.Minute).Err()
	assert.NoError(suite.T(), err)

	result, err := redis.RedisClient.Get(redis.RedisClient.Context(), testKey).Result()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testValue, result)

	// 清理测试数据
	redis.RedisClient.Del(redis.RedisClient.Context(), testKey)
}

func (suite *IntegrationTestSuite) TestConfigurationLoading() {
	// 验证配置是否正确加载
	assert.NotNil(suite.T(), config.AppConfig)
	assert.Equal(suite.T(), TestDBName, config.AppConfig.Database.DBName)
	assert.Equal(suite.T(), TestRedisDB, config.AppConfig.Redis.DB)
	assert.Equal(suite.T(), TestJWTSecret, config.AppConfig.JWT.Secret)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
