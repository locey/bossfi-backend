package services

import (
	"errors"
	"strings"
	"time"

	"bossfi-backend/api/models"
	"bossfi-backend/db/database"
	"bossfi-backend/db/redis"
	"bossfi-backend/utils"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

// GetNonceMessage 获取用于签名的消息和 nonce
func (s *UserService) GetNonceMessage(walletAddress string) (string, string, error) {
	// 验证钱包地址格式
	if !utils.ValidateWalletAddress(walletAddress) {
		return "", "", errors.New("invalid wallet address format")
	}

	// 统一地址格式
	walletAddress = strings.ToLower(walletAddress)

	// 生成 nonce
	nonce, err := utils.GenerateNonce()
	if err != nil {
		logrus.Errorf("Failed to generate nonce: %v", err)
		return "", "", errors.New("failed to generate nonce")
	}

	// 存储 nonce 到 Redis
	if err := redis.SetNonce(walletAddress, nonce); err != nil {
		logrus.Errorf("Failed to store nonce: %v", err)
		return "", "", errors.New("failed to store nonce")
	}

	// 创建签名消息
	message := utils.CreateSignMessage(walletAddress, nonce)

	return message, nonce, nil
}

// VerifyAndLogin 验证签名并登录
func (s *UserService) VerifyAndLogin(walletAddress, signature, message string) (*models.User, string, error) {
	// 验证钱包地址格式
	if !utils.ValidateWalletAddress(walletAddress) {
		return nil, "", errors.New("invalid wallet address format")
	}

	// 统一地址格式
	walletAddress = strings.ToLower(walletAddress)

	// 验证签名
	isValid, err := utils.VerifySignature(message, signature, walletAddress)
	if err != nil {
		logrus.Errorf("Failed to verify signature: %v", err)
		return nil, "", errors.New("failed to verify signature")
	}

	if !isValid {
		return nil, "", errors.New("invalid signature")
	}

	// 查找或创建用户
	user, err := s.findOrCreateUser(walletAddress)
	if err != nil {
		logrus.Errorf("Failed to find or create user: %v", err)
		return nil, "", errors.New("failed to process user")
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	if err := database.DB.Save(user).Error; err != nil {
		logrus.Errorf("Failed to update last login time: %v", err)
	}

	// 生成 JWT token
	token, err := utils.GenerateJWT(user.ID.String(), user.WalletAddress)
	if err != nil {
		logrus.Errorf("Failed to generate JWT: %v", err)
		return nil, "", errors.New("failed to generate token")
	}

	// 存储用户会话到 Redis
	if err := redis.SetUserSession(user.ID.String(), token); err != nil {
		logrus.Warnf("Failed to store user session: %v", err)
	}

	// 删除使用过的 nonce
	if err := redis.DeleteNonce(walletAddress); err != nil {
		logrus.Warnf("Failed to delete nonce: %v", err)
	}

	return user, token, nil
}

// findOrCreateUser 查找或创建用户
func (s *UserService) findOrCreateUser(walletAddress string) (*models.User, error) {
	var user models.User

	err := database.DB.Where("wallet_address = ?", walletAddress).First(&user).Error
	if err == nil {
		return &user, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建新用户
		newUser := models.User{
			ID:            uuid.New(),
			WalletAddress: walletAddress,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := database.DB.Create(&newUser).Error; err != nil {
			return nil, err
		}

		return &newUser, nil
	}

	return nil, err
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserByID(userID string) (*models.User, error) {
	var user models.User

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	err = database.DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(userID string, updateData map[string]interface{}) (*models.User, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Model(user).Updates(updateData).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// Logout 用户登出
func (s *UserService) Logout(userID string) error {
	// 删除 Redis 中的用户会话
	if err := redis.DeleteUserSession(userID); err != nil {
		logrus.Warnf("Failed to delete user session: %v", err)
	}

	return nil
}
