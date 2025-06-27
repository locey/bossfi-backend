package config

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 测试辅助函数
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := os.Getenv(name)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := os.Getenv(name)
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	// 支持更多的布尔值表示
	switch strings.ToLower(valStr) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	}
	return defaultVal
}

func getEnvAsDuration(name string, defaultVal time.Duration) time.Duration {
	valueStr := os.Getenv(name)
	if hours, err := strconv.Atoi(valueStr); err == nil {
		return time.Duration(hours) * time.Hour
	}
	return defaultVal
}

func TestInit(t *testing.T) {
	// 备份原始环境变量
	originalVars := make(map[string]string)
	testVars := map[string]string{
		"DB_HOST":          "test-host",
		"DB_PORT":          "5433",
		"DB_USER":          "test-user",
		"DB_PASSWORD":      "test-password",
		"DB_NAME":          "test-db",
		"REDIS_HOST":       "test-redis",
		"REDIS_PORT":       "6380",
		"REDIS_DB":         "2",
		"JWT_SECRET":       "test-jwt-secret",
		"JWT_EXPIRE_HOURS": "24",
		"PORT":             "8081",
		"GIN_MODE":         "debug",
	}

	// 备份和设置测试环境变量
	for key, value := range testVars {
		originalVars[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	// 测试后恢复原始环境变量
	defer func() {
		for key, originalValue := range originalVars {
			if originalValue == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, originalValue)
			}
		}
	}()

	// 初始化配置
	Init()

	// 验证数据库配置
	assert.Equal(t, "test-host", AppConfig.Database.Host)
	assert.Equal(t, "5433", AppConfig.Database.Port)
	assert.Equal(t, "test-user", AppConfig.Database.User)
	assert.Equal(t, "test-password", AppConfig.Database.Password)
	assert.Equal(t, "test-db", AppConfig.Database.DBName)

	// 验证 Redis 配置
	assert.Equal(t, "test-redis", AppConfig.Redis.Host)
	assert.Equal(t, "6380", AppConfig.Redis.Port)
	assert.Equal(t, 2, AppConfig.Redis.DB)

	// 验证 JWT 配置
	assert.Equal(t, "test-jwt-secret", AppConfig.JWT.Secret)
	assert.Equal(t, 24*time.Hour, AppConfig.JWT.ExpireHours)

	// 验证服务器配置
	assert.Equal(t, "8081", AppConfig.Server.Port)
	assert.Equal(t, "debug", AppConfig.Server.GinMode)
}

func TestInitWithDefaults(t *testing.T) {
	// 清除所有相关环境变量
	envVars := []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
		"REDIS_HOST", "REDIS_PORT", "REDIS_DB",
		"JWT_SECRET", "JWT_EXPIRE_HOURS",
		"PORT", "GIN_MODE",
		"BLOCKCHAIN_RPC_URL",
		"CRON_ENABLED", "BLOCKCHAIN_SYNC_INTERVAL",
	}

	originalVars := make(map[string]string)
	for _, key := range envVars {
		originalVars[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	// 测试后恢复原始环境变量
	defer func() {
		for key, originalValue := range originalVars {
			if originalValue == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, originalValue)
			}
		}
	}()

	// 初始化配置
	Init()

	// 验证默认值
	assert.Equal(t, "localhost", AppConfig.Database.Host)
	assert.Equal(t, "5432", AppConfig.Database.Port)
	assert.Equal(t, "postgres", AppConfig.Database.User)
	assert.Equal(t, "bossfi", AppConfig.Database.DBName)

	assert.Equal(t, "localhost", AppConfig.Redis.Host)
	assert.Equal(t, "6379", AppConfig.Redis.Port)
	assert.Equal(t, 0, AppConfig.Redis.DB)

	assert.Equal(t, 24*time.Hour, AppConfig.JWT.ExpireHours)

	assert.Equal(t, "8080", AppConfig.Server.Port)
	assert.Equal(t, "debug", AppConfig.Server.GinMode)

	assert.Equal(t, "", AppConfig.Blockchain.RPCURL)

	assert.True(t, AppConfig.Cron.Enabled)
	assert.Equal(t, "*/5 * * * *", AppConfig.Cron.BlockchainSyncInterval)
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue int
		expected     int
	}{
		{
			name:         "valid integer",
			envValue:     "123",
			defaultValue: 456,
			expected:     123,
		},
		{
			name:         "invalid integer",
			envValue:     "abc",
			defaultValue: 456,
			expected:     456,
		},
		{
			name:         "empty value",
			envValue:     "",
			defaultValue: 456,
			expected:     456,
		},
		{
			name:         "negative integer",
			envValue:     "-123",
			defaultValue: 456,
			expected:     -123,
		},
		{
			name:         "zero",
			envValue:     "0",
			defaultValue: 456,
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_INT_VAR"

			// 设置环境变量
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			} else {
				os.Unsetenv(key)
			}

			// 测试后清理
			defer os.Unsetenv(key)

			result := getEnvAsInt(key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvAsBool(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "true string",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "True string",
			envValue:     "True",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "TRUE string",
			envValue:     "TRUE",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "1 string",
			envValue:     "1",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "false string",
			envValue:     "false",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "0 string",
			envValue:     "0",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "invalid string",
			envValue:     "invalid",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "empty value",
			envValue:     "",
			defaultValue: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_BOOL_VAR"

			// 设置环境变量
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			} else {
				os.Unsetenv(key)
			}

			// 测试后清理
			defer os.Unsetenv(key)

			result := getEnvAsBool(key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{
			name:         "valid hours",
			envValue:     "2",
			defaultValue: 1 * time.Hour,
			expected:     2 * time.Hour,
		},
		{
			name:         "zero hours",
			envValue:     "0",
			defaultValue: 1 * time.Hour,
			expected:     0,
		},
		{
			name:         "invalid value",
			envValue:     "abc",
			defaultValue: 1 * time.Hour,
			expected:     1 * time.Hour,
		},
		{
			name:         "empty value",
			envValue:     "",
			defaultValue: 1 * time.Hour,
			expected:     1 * time.Hour,
		},
		{
			name:         "negative hours",
			envValue:     "-5",
			defaultValue: 1 * time.Hour,
			expected:     -5 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_DURATION_VAR"

			// 设置环境变量
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			} else {
				os.Unsetenv(key)
			}

			// 测试后清理
			defer os.Unsetenv(key)

			result := getEnvAsDuration(key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigSingleton(t *testing.T) {
	// 初始化配置
	Init()

	// 获取配置引用
	config1 := AppConfig

	// 再次初始化
	Init()

	// 获取另一个配置引用
	config2 := AppConfig

	// 验证是同一个实例（指针相等）
	assert.Equal(t, config1, config2)
}

func TestValidateConfig(t *testing.T) {
	// 设置有效配置
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("JWT_SECRET", "test-secret-key")

	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("JWT_SECRET")
	}()

	// 初始化配置不应该 panic
	require.NotPanics(t, func() {
		Init()
	})

	// 验证配置被正确设置
	assert.NotNil(t, AppConfig)
	assert.NotEmpty(t, AppConfig.Database.Host)
	assert.NotEmpty(t, AppConfig.Database.User)
	assert.NotEmpty(t, AppConfig.Database.DBName)
	assert.NotEmpty(t, AppConfig.JWT.Secret)
}
