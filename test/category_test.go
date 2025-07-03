package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"bossfi-backend/api/routes"
	"bossfi-backend/db/database"
	"bossfi-backend/db/redis"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CategoryTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *CategoryTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	// 设置测试环境
	err := setupTestEnv()
	require.NoError(suite.T(), err)

	// 设置路由
	suite.router = routes.SetupRoutes()
}

func (suite *CategoryTestSuite) SetupTest() {
	// 每个测试前清理数据
	database.DB.Exec("TRUNCATE TABLE article_categories RESTART IDENTITY CASCADE")
	database.DB.Exec("TRUNCATE TABLE articles RESTART IDENTITY CASCADE")
	database.DB.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	redis.RedisClient.FlushDB(redis.RedisClient.Context())
}

func (suite *CategoryTestSuite) TearDownSuite() {
	cleanupTestEnv()
}

func (suite *CategoryTestSuite) TestCreateCategory() {
	// 先登录获取token
	token := suite.getTestToken()

	requestBody := map[string]interface{}{
		"name":        "测试分类",
		"description": "这是一个测试分类",
		"icon":        "test-icon",
		"color":       "#FF5733",
		"sort_order":  1,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "测试分类", response["name"])
	assert.Equal(suite.T(), "这是一个测试分类", response["description"])
	assert.Equal(suite.T(), "test-icon", response["icon"])
	assert.Equal(suite.T(), "#FF5733", response["color"])
}

func (suite *CategoryTestSuite) TestGetCategories() {
	// 先创建一些测试分类
	suite.createTestCategories()

	req := httptest.NewRequest("GET", "/api/v1/categories", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	categories := response["categories"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(categories), 2)
}

func (suite *CategoryTestSuite) TestGetActiveCategories() {
	// 先创建一些测试分类
	suite.createTestCategories()

	req := httptest.NewRequest("GET", "/api/v1/categories/active", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	categories := response["categories"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(categories), 2)
}

func (suite *CategoryTestSuite) TestGetCategory() {
	// 先创建测试分类
	categoryID := suite.createTestCategory()

	req := httptest.NewRequest("GET", "/api/v1/categories/"+categoryID, nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "测试分类", response["name"])
}

func (suite *CategoryTestSuite) TestUpdateCategory() {
	// 先创建测试分类
	categoryID := suite.createTestCategory()
	token := suite.getTestToken()

	requestBody := map[string]interface{}{
		"name":        "更新后的分类",
		"description": "这是更新后的描述",
		"icon":        "updated-icon",
		"color":       "#33FF57",
		"sort_order":  2,
		"is_active":   true,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/api/v1/categories/"+categoryID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "更新后的分类", response["name"])
	assert.Equal(suite.T(), "这是更新后的描述", response["description"])
}

func (suite *CategoryTestSuite) TestDeleteCategory() {
	// 先创建测试分类
	categoryID := suite.createTestCategory()
	token := suite.getTestToken()

	req := httptest.NewRequest("DELETE", "/api/v1/categories/"+categoryID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// 验证分类已被删除
	req2 := httptest.NewRequest("GET", "/api/v1/categories/"+categoryID, nil)
	w2 := httptest.NewRecorder()

	suite.router.ServeHTTP(w2, req2)

	assert.Equal(suite.T(), http.StatusNotFound, w2.Code)
}

// 辅助方法
func (suite *CategoryTestSuite) createTestCategory() string {
	token := suite.getTestToken()

	requestBody := map[string]interface{}{
		"name":        "测试分类",
		"description": "这是一个测试分类",
		"icon":        "test-icon",
		"color":       "#FF5733",
		"sort_order":  1,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	return fmt.Sprintf("%.0f", response["id"])
}

func (suite *CategoryTestSuite) createTestCategories() {
	suite.createTestCategory()

	// 创建第二个分类
	token := suite.getTestToken()
	requestBody := map[string]interface{}{
		"name":        "测试分类2",
		"description": "这是第二个测试分类",
		"icon":        "test-icon-2",
		"color":       "#33FF57",
		"sort_order":  2,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
}

func (suite *CategoryTestSuite) getTestToken() string {
	// 创建测试用户并获取token
	walletAddress := GetTestWalletAddress()

	// 获取nonce
	nonceReq := map[string]interface{}{
		"wallet_address": walletAddress,
	}
	nonceBody, _ := json.Marshal(nonceReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/nonce", bytes.NewBuffer(nonceBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	var nonceResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &nonceResp)

	// 获取测试token
	tokenReq := map[string]interface{}{
		"wallet_address": walletAddress,
	}
	tokenBody, _ := json.Marshal(tokenReq)
	req2 := httptest.NewRequest("POST", "/api/v1/auth/test-token", bytes.NewBuffer(tokenBody))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, req2)

	var tokenResp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &tokenResp)

	return tokenResp["token"].(string)
}

func TestCategoryTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryTestSuite))
}
