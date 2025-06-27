package services

import (
	"os"
	"testing"
	"time"

	"bossfi-backend/api/models"
	"bossfi-backend/config"
	"bossfi-backend/db/database"
	"bossfi-backend/db/redis"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UserServiceTestSuite struct {
	suite.Suite
	userService *UserService
}

func (suite *UserServiceTestSuite) SetupSuite() {
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

	// 创建用户服务实例
	suite.userService = NewUserService()
}

func (suite *UserServiceTestSuite) SetupTest() {
	// 每个测试前清理数据
	database.DB.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	redis.RedisClient.FlushDB(redis.RedisClient.Context())
}

func (suite *UserServiceTestSuite) TearDownSuite() {
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

func (suite *UserServiceTestSuite) TestGetNonceMessage() {
	tests := []struct {
		name          string
		walletAddress string
		expectError   bool
	}{
		{
			name:          "valid wallet address",
			walletAddress: "0x1234567890123456789012345678901234567890",
			expectError:   false,
		},
		{
			name:          "invalid wallet address - no 0x prefix",
			walletAddress: "1234567890123456789012345678901234567890",
			expectError:   true,
		},
		{
			name:          "invalid wallet address - too short",
			walletAddress: "0x123",
			expectError:   true,
		},
		{
			name:          "empty wallet address",
			walletAddress: "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			message, nonce, err := suite.userService.GetNonceMessage(tt.walletAddress)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, message)
				assert.Empty(t, nonce)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, message)
				assert.NotEmpty(t, nonce)
				assert.Contains(t, message, "Welcome to BossFi!")
				assert.Contains(t, message, tt.walletAddress)
				assert.Contains(t, message, nonce)

				// 验证 nonce 是否存储到 Redis
				storedNonce, err := redis.GetNonce(tt.walletAddress)
				assert.NoError(t, err)
				assert.Equal(t, nonce, storedNonce)
			}
		})
	}
}

func (suite *UserServiceTestSuite) TestFindOrCreateUser() {
	walletAddress := "0x1234567890123456789012345678901234567890"

	// 第一次调用应该创建新用户
	user1, err := suite.userService.findOrCreateUser(walletAddress)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user1)
	assert.Equal(suite.T(), walletAddress, user1.WalletAddress)
	assert.NotEqual(suite.T(), uuid.Nil, user1.ID)

	// 第二次调用应该返回相同的用户
	user2, err := suite.userService.findOrCreateUser(walletAddress)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user2)
	assert.Equal(suite.T(), user1.ID, user2.ID)
	assert.Equal(suite.T(), user1.WalletAddress, user2.WalletAddress)
}

func (suite *UserServiceTestSuite) TestGetUserByID() {
	// 创建测试用户
	testUser := models.User{
		ID:            uuid.New(),
		WalletAddress: "0x1234567890123456789012345678901234567890",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err := database.DB.Create(&testUser).Error
	require.NoError(suite.T(), err)

	tests := []struct {
		name        string
		userID      string
		expectError bool
	}{
		{
			name:        "valid user ID",
			userID:      testUser.ID.String(),
			expectError: false,
		},
		{
			name:        "invalid user ID format",
			userID:      "invalid-uuid",
			expectError: true,
		},
		{
			name:        "non-existent user ID",
			userID:      uuid.New().String(),
			expectError: true,
		},
		{
			name:        "empty user ID",
			userID:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			user, err := suite.userService.GetUserByID(tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, testUser.ID, user.ID)
				assert.Equal(t, testUser.WalletAddress, user.WalletAddress)
			}
		})
	}
}

func (suite *UserServiceTestSuite) TestUpdateUser() {
	// 创建测试用户
	testUser := models.User{
		ID:            uuid.New(),
		WalletAddress: "0x1234567890123456789012345678901234567890",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err := database.DB.Create(&testUser).Error
	require.NoError(suite.T(), err)

	// 更新用户信息
	updateData := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}

	updatedUser, err := suite.userService.UpdateUser(testUser.ID.String(), updateData)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedUser)
	assert.Equal(suite.T(), testUser.ID, updatedUser.ID)
	assert.Equal(suite.T(), "testuser", *updatedUser.Username)
	assert.Equal(suite.T(), "test@example.com", *updatedUser.Email)
}

func (suite *UserServiceTestSuite) TestLogout() {
	userID := uuid.New().String()
	token := "test-token"

	// 首先设置用户会话
	err := redis.SetUserSession(userID, token)
	require.NoError(suite.T(), err)

	// 验证会话存在
	storedToken, err := redis.GetUserSession(userID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), token, storedToken)

	// 登出用户
	err = suite.userService.Logout(userID)
	assert.NoError(suite.T(), err)

	// 验证会话已删除
	_, err = redis.GetUserSession(userID)
	assert.Error(suite.T(), err) // 应该返回 redis.Nil 错误
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
