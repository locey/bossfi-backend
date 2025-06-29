package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"bossfi-backend/config"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateNonce 生成随机 nonce
func GenerateNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSignMessage 创建签名消息
func CreateSignMessage(walletAddress, nonce string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("Welcome to BossFi!\n\nClick to sign in and accept the BossFi Terms of Service.\n\nThis request will not trigger a blockchain transaction or cost any gas fees.\n\nWallet address:\n%s\n\nNonce:\n%s\n\nTimestamp:\n%d",
		walletAddress, nonce, timestamp)
}

// VerifySignature 验证钱包签名
func VerifySignature(message, signature, walletAddress string) (bool, error) {
	// 移除 0x 前缀
	signature = strings.TrimPrefix(signature, "0x")
	walletAddress = strings.TrimPrefix(walletAddress, "0x")

	// 解码签名
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %v", err)
	}

	if len(sigBytes) != 65 {
		return false, fmt.Errorf("signature length should be 65 bytes")
	}

	// 调整 v 值
	if sigBytes[64] == 27 || sigBytes[64] == 28 {
		sigBytes[64] -= 27
	}

	// 计算消息哈希
	hash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))

	// 恢复公钥
	pubKey, err := crypto.SigToPub(hash.Bytes(), sigBytes)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %v", err)
	}

	// 计算地址
	recoveredAddress := crypto.PubkeyToAddress(*pubKey).Hex()
	expectedAddress := "0x" + walletAddress

	return strings.EqualFold(recoveredAddress, expectedAddress), nil
}

// JWTClaims JWT 声明结构
type JWTClaims struct {
	UserID        string `json:"user_id"`
	WalletAddress string `json:"wallet_address"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成 JWT token
func GenerateJWT(userID, walletAddress string) (string, error) {
	// 生成随机JWT ID
	jwtID, err := GenerateNonce()
	if err != nil {
		jwtID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	claims := JWTClaims{
		UserID:        userID,
		WalletAddress: walletAddress,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.AppConfig.JWT.ExpireHours)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "bossfi-backend",
			Subject:   userID,
			ID:        jwtID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWT.Secret))
}

// ParseJWT 解析 JWT token
func ParseJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.AppConfig.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateWalletAddress 验证钱包地址格式
func ValidateWalletAddress(address string) bool {
	if !strings.HasPrefix(address, "0x") {
		return false
	}

	if len(address) != 42 {
		return false
	}

	_, err := hexutil.Decode(address)
	return err == nil
}
