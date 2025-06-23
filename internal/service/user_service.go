package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"bossfi-blockchain-backend/internal/domain/user"
	"bossfi-blockchain-backend/internal/repository"
	"bossfi-blockchain-backend/pkg/config"
	"bossfi-blockchain-backend/pkg/logger"
)

// UserService 用户服务接口
type UserService interface {
	// 认证相关
	GenerateNonce(ctx context.Context, walletAddress string) (string, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)

	// 用户管理
	GetProfile(ctx context.Context, userID string) (*user.User, error)
	UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*user.User, error)
	GetUserStats(ctx context.Context, userID string) (*repository.UserStats, error)
	SearchUsers(ctx context.Context, keyword string, page, pageSize int) (*SearchUsersResponse, error)

	// 余额管理
	GetBalance(ctx context.Context, userID string) (*BalanceResponse, error)
	UpdateBalance(ctx context.Context, userID string, bossBalance, stakedAmount, rewardBalance decimal.Decimal) error

	// 管理员功能
	ListUsers(ctx context.Context, page, pageSize int) (*ListUsersResponse, error)
	GetUserByID(ctx context.Context, userID string) (*user.User, error)
	DeleteUser(ctx context.Context, userID string) error
}

// 请求和响应结构
type LoginRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	Signature     string `json:"signature" binding:"required"`
	Message       string `json:"message" binding:"required"`
}

type LoginResponse struct {
	Token string     `json:"token"`
	User  *user.User `json:"user"`
}

type UpdateProfileRequest struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Avatar   *string `json:"avatar"`
	Bio      *string `json:"bio"`
}

type BalanceResponse struct {
	BossBalance   decimal.Decimal `json:"boss_balance"`
	StakedAmount  decimal.Decimal `json:"staked_amount"`
	RewardBalance decimal.Decimal `json:"reward_balance"`
}

type SearchUsersResponse struct {
	Users      []*user.User `json:"users"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}

type ListUsersResponse struct {
	Users      []*user.User `json:"users"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}

// userService 用户服务实现
type userService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
	logger   *logger.Logger
	redis    *redis.Client
}

// NewUserService 创建用户服务
func NewUserService(userRepo repository.UserRepository, cfg *config.Config, logger *logger.Logger, redisClient *redis.Client) UserService {
	return &userService{
		userRepo: userRepo,
		cfg:      cfg,
		logger:   logger,
		redis:    redisClient,
	}
}

// GenerateNonce 生成登录随机数
func (s *userService) GenerateNonce(ctx context.Context, walletAddress string) (string, error) {
	// 生成随机nonce
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random nonce: %w", err)
	}

	nonce := fmt.Sprintf("Please sign this message to authenticate with BossFi:\n\nWallet: %s\nNonce: %s\nTimestamp: %d",
		walletAddress, hex.EncodeToString(bytes), time.Now().Unix())

	// 存储到Redis，5分钟过期
	key := fmt.Sprintf("login_nonce:%s", strings.ToLower(walletAddress))
	err := s.redis.Set(ctx, key, nonce, 5*time.Minute).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store nonce: %w", err)
	}

	return nonce, nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	walletAddress := strings.ToLower(req.WalletAddress)

	// 从Redis获取nonce
	key := fmt.Sprintf("login_nonce:%s", walletAddress)
	storedMessage, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("nonce not found or expired")
		}
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// 验证消息是否匹配
	if storedMessage != req.Message {
		return nil, fmt.Errorf("message mismatch")
	}

	// 验证签名（这里需要实现具体的签名验证逻辑）
	if !s.verifySignature(req.WalletAddress, req.Message, req.Signature) {
		return nil, fmt.Errorf("invalid signature")
	}

	// 删除已使用的nonce
	s.redis.Del(ctx, key)

	// 查找或创建用户
	u, err := s.userRepo.GetByWalletAddress(ctx, walletAddress)
	if err != nil {
		if err == user.ErrUserNotFound {
			// 创建新用户
			u = &user.User{
				WalletAddress: walletAddress,
				BossBalance:   decimal.Zero,
				StakedAmount:  decimal.Zero,
				RewardBalance: decimal.Zero,
			}
			if err := s.userRepo.Create(ctx, u); err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
	}

	// 生成JWT token
	token, err := s.generateJWT(u.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token: token,
		User:  u,
	}, nil
}

// verifySignature 验证以太坊钱包签名
func (s *userService) verifySignature(walletAddress, message, signature string) bool {
	// 验证钱包地址格式
	if !common.IsHexAddress(walletAddress) {
		s.logger.Warn("Invalid wallet address format",
			zap.String("wallet_address", walletAddress))
		return false
	}

	// 解析签名
	sigBytes, err := hexutil.Decode(signature)
	if err != nil {
		s.logger.Warn("Failed to decode signature",
			zap.String("signature", signature),
			zap.Error(err))
		return false
	}

	// 签名长度必须是65字节
	if len(sigBytes) != 65 {
		s.logger.Warn("Invalid signature length",
			zap.Int("length", len(sigBytes)))
		return false
	}

	// 以太坊签名的v值需要调整
	if sigBytes[64] == 27 || sigBytes[64] == 28 {
		sigBytes[64] -= 27
	}

	// 创建消息哈希（以太坊使用特定的前缀）
	messageHash := accounts.TextHash([]byte(message))

	// 从签名中恢复公钥
	publicKey, err := crypto.SigToPub(messageHash, sigBytes)
	if err != nil {
		s.logger.Warn("Failed to recover public key from signature",
			zap.Error(err))
		return false
	}

	// 从公钥生成地址
	recoveredAddress := crypto.PubkeyToAddress(*publicKey)

	// 比较地址（忽略大小写）
	expectedAddress := common.HexToAddress(walletAddress)
	isValid := strings.EqualFold(recoveredAddress.Hex(), expectedAddress.Hex())

	if !isValid {
		s.logger.Warn("Signature verification failed",
			zap.String("expected_address", expectedAddress.Hex()),
			zap.String("recovered_address", recoveredAddress.Hex()))
	}

	return isValid
}

// generateJWT 生成JWT token
func (s *userService) generateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.AccessSecret))
}

// GetProfile 获取用户资料
func (s *userService) GetProfile(ctx context.Context, userID string) (*user.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// UpdateProfile 更新用户资料
func (s *userService) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*user.User, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Username != nil && *req.Username != "" {
		if u.Username == nil || *u.Username != *req.Username {
			existing, err := s.userRepo.GetByUsername(ctx, *req.Username)
			if err != nil && err != user.ErrUserNotFound {
				return nil, err
			}
			if existing != nil {
				return nil, user.ErrUsernameAlreadyExists
			}
		}
		u.Username = req.Username
	}

	if req.Email != nil && *req.Email != "" {
		if u.Email == nil || *u.Email != *req.Email {
			existing, err := s.userRepo.GetByEmail(ctx, *req.Email)
			if err != nil && err != user.ErrUserNotFound {
				return nil, err
			}
			if existing != nil {
				return nil, user.ErrEmailAlreadyExists
			}
		}
		u.Email = req.Email
	}

	if req.Avatar != nil {
		u.Avatar = req.Avatar
	}
	if req.Bio != nil {
		u.Bio = req.Bio
	}

	u.IsProfileComplete = u.IsProfileCompleted()

	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

// GetUserStats 获取用户统计信息
func (s *userService) GetUserStats(ctx context.Context, userID string) (*repository.UserStats, error) {
	return s.userRepo.GetUserStats(ctx, userID)
}

// SearchUsers 搜索用户
func (s *userService) SearchUsers(ctx context.Context, keyword string, page, pageSize int) (*SearchUsersResponse, error) {
	offset := (page - 1) * pageSize
	users, total, err := s.userRepo.Search(ctx, keyword, offset, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &SearchUsersResponse{
		Users:      users,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetBalance 获取用户余额
func (s *userService) GetBalance(ctx context.Context, userID string) (*BalanceResponse, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &BalanceResponse{
		BossBalance:   u.BossBalance,
		StakedAmount:  u.StakedAmount,
		RewardBalance: u.RewardBalance,
	}, nil
}

// UpdateBalance 更新用户余额
func (s *userService) UpdateBalance(ctx context.Context, userID string, bossBalance, stakedAmount, rewardBalance decimal.Decimal) error {
	return s.userRepo.UpdateBalance(ctx, userID, bossBalance, stakedAmount, rewardBalance)
}

// ListUsers 获取用户列表（管理员功能）
func (s *userService) ListUsers(ctx context.Context, page, pageSize int) (*ListUsersResponse, error) {
	offset := (page - 1) * pageSize
	users, total, err := s.userRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &ListUsersResponse{
		Users:      users,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUserByID 根据ID获取用户（管理员功能）
func (s *userService) GetUserByID(ctx context.Context, userID string) (*user.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// DeleteUser 删除用户（管理员功能）
func (s *userService) DeleteUser(ctx context.Context, userID string) error {
	return s.userRepo.Delete(ctx, userID)
}
