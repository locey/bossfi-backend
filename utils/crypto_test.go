package utils

import (
	"os"
	"testing"

	"bossfi-backend/config"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// 设置测试配置
	os.Setenv("JWT_SECRET", "test-secret-key")
	os.Setenv("JWT_EXPIRE_HOURS", "1")

	// 初始化配置
	config.Init()

	// 运行测试
	code := m.Run()
	os.Exit(code)
}

func TestGenerateNonce(t *testing.T) {
	t.Run("should generate valid nonce", func(t *testing.T) {
		nonce, err := GenerateNonce()

		assert.NoError(t, err)
		assert.NotEmpty(t, nonce)
		assert.Equal(t, 32, len(nonce)) // 16 bytes = 32 hex characters
	})

	t.Run("should generate different nonces", func(t *testing.T) {
		nonce1, err1 := GenerateNonce()
		nonce2, err2 := GenerateNonce()

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, nonce1, nonce2)
	})
}

func TestCreateSignMessage(t *testing.T) {
	walletAddress := "0x1234567890123456789012345678901234567890"
	nonce := "testnonce123"

	message := CreateSignMessage(walletAddress, nonce)

	assert.Contains(t, message, "Welcome to BossFi!")
	assert.Contains(t, message, walletAddress)
	assert.Contains(t, message, nonce)
	assert.Contains(t, message, "Timestamp:")
}

func TestValidateWalletAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{
			name:     "valid address",
			address:  "0x1234567890123456789012345678901234567890",
			expected: true,
		},
		{
			name:     "valid address with uppercase",
			address:  "0x1234567890ABCDEF1234567890ABCDEF12345678",
			expected: true,
		},
		{
			name:     "invalid - no 0x prefix",
			address:  "1234567890123456789012345678901234567890",
			expected: false,
		},
		{
			name:     "invalid - too short",
			address:  "0x123456789012345678901234567890123456789",
			expected: false,
		},
		{
			name:     "invalid - too long",
			address:  "0x12345678901234567890123456789012345678901",
			expected: false,
		},
		{
			name:     "invalid - non-hex characters",
			address:  "0x123456789012345678901234567890123456789G",
			expected: false,
		},
		{
			name:     "empty address",
			address:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateWalletAddress(tt.address)
			assert.Equal(t, tt.expected, result)
		})
	}
}
