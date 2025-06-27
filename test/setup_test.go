package test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"bossfi-backend/config"
	"bossfi-backend/db/database"
	"bossfi-backend/db/redis"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var (
	TestDBName    = "bossfi_test"
	TestRedisDB   = 1
	TestJWTSecret = "test-secret-key-for-testing-only"
)

// TestMain 设置和清理测试环境
func TestMain(m *testing.M) {
	// 设置测试环境
	if err := setupTestEnv(); err != nil {
		log.Fatalf("Failed to setup test environment: %v", err)
	}

	// 运行测试
	code := m.Run()

	// 清理测试环境
	cleanupTestEnv()

	os.Exit(code)
}

// setupTestEnv 设置测试环境
func setupTestEnv() error {
	// 设置日志级别为 Error 以减少测试输出
	logrus.SetLevel(logrus.ErrorLevel)

	// 加载环境变量
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using default test config")
	}

	// 设置测试配置
	setTestConfig()

	// 初始化配置
	config.Init()

	// 覆盖配置为测试配置
	config.AppConfig.Database.DBName = TestDBName
	config.AppConfig.Redis.DB = TestRedisDB
	config.AppConfig.JWT.Secret = TestJWTSecret
	config.AppConfig.Server.GinMode = "test"

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		return fmt.Errorf("failed to init test database: %v", err)
	}

	// 初始化 Redis
	if err := redis.InitRedis(); err != nil {
		return fmt.Errorf("failed to init test redis: %v", err)
	}

	return nil
}

// setTestConfig 设置测试环境变量
func setTestConfig() {
	os.Setenv("DB_NAME", TestDBName)
	os.Setenv("REDIS_DB", "1")
	os.Setenv("JWT_SECRET", TestJWTSecret)
	os.Setenv("GIN_MODE", "test")
	os.Setenv("LOG_LEVEL", "error")
	os.Setenv("CRON_ENABLED", "false")
}

// cleanupTestEnv 清理测试环境
func cleanupTestEnv() {
	// 清理 Redis 测试数据
	if redis.RedisClient != nil {
		redis.RedisClient.FlushDB(redis.RedisClient.Context())
		redis.RedisClient.Close()
	}

	// 关闭数据库连接
	if database.DB != nil {
		if sqlDB, err := database.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}
}

// CreateTestUser 创建测试用户
func CreateTestUser() map[string]interface{} {
	return map[string]interface{}{
		"wallet_address": "0x1234567890123456789012345678901234567890",
		"username":       "testuser",
		"email":          "test@example.com",
	}
}

// GetTestWalletAddress 获取测试钱包地址
func GetTestWalletAddress() string {
	return "0x1234567890123456789012345678901234567890"
}

// GetTestSignature 获取测试签名（模拟）
func GetTestSignature() string {
	return "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12345678901b"
}

// GetTestMessage 获取测试消息
func GetTestMessage() string {
	return "Welcome to BossFi!\n\nClick to sign in and accept the BossFi Terms of Service.\n\nThis request will not trigger a blockchain transaction or cost any gas fees.\n\nWallet address:\n0x1234567890123456789012345678901234567890\n\nNonce:\ntestnonce123\n\nTimestamp:\n1640995200"
}
