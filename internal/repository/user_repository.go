package repository

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"bossfi-blockchain-backend/internal/domain/user"
)

// UserStats 用户统计信息
type UserStats struct {
	PostCount    int64           `json:"post_count"`
	ReplyCount   int64           `json:"reply_count"`
	LikeCount    int64           `json:"like_count"`
	StakeCount   int64           `json:"stake_count"`
	TotalStaked  decimal.Decimal `json:"total_staked"`
	TotalRewards decimal.Decimal `json:"total_rewards"`
}

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(ctx context.Context, user *user.User) error
	GetByID(ctx context.Context, id string) (*user.User, error)
	GetByWalletAddress(ctx context.Context, walletAddress string) (*user.User, error)
	GetByUsername(ctx context.Context, username string) (*user.User, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
	Update(ctx context.Context, user *user.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*user.User, int64, error)
	Search(ctx context.Context, keyword string, offset, limit int) ([]*user.User, int64, error)
	GetUserStats(ctx context.Context, userID string) (*UserStats, error)
}

// userRepository 用户仓储实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

// GetByWalletAddress 根据钱包地址获取用户
func (r *userRepository) GetByWalletAddress(ctx context.Context, walletAddress string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("wallet_address = ?", walletAddress).First(&u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

// Delete 删除用户
func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&user.User{}).Error
}

// List 获取用户列表
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*user.User, int64, error) {
	var users []*user.User
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&user.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取用户列表
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error

	return users, total, err
}

// Search 搜索用户
func (r *userRepository) Search(ctx context.Context, keyword string, offset, limit int) ([]*user.User, int64, error) {
	var users []*user.User
	var total int64

	query := r.db.WithContext(ctx).Model(&user.User{})

	// 构建搜索条件
	searchCondition := fmt.Sprintf("%%%s%%", keyword)
	query = query.Where("username LIKE ? OR email LIKE ? OR wallet_address LIKE ?",
		searchCondition, searchCondition, searchCondition)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取用户列表
	err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error

	return users, total, err
}

// GetUserStats 获取用户统计信息
func (r *userRepository) GetUserStats(ctx context.Context, userID string) (*UserStats, error) {
	stats := &UserStats{}

	// 获取帖子数量
	r.db.WithContext(ctx).Model(&struct {
		UserID string `gorm:"column:user_id"`
	}{}).
		Table("posts").
		Where("user_id = ?", userID).
		Count(&stats.PostCount)

	// 获取回复数量
	r.db.WithContext(ctx).Model(&struct {
		UserID string `gorm:"column:user_id"`
	}{}).
		Table("post_replies").
		Where("user_id = ?", userID).
		Count(&stats.ReplyCount)

	// 获取质押统计
	r.db.WithContext(ctx).Model(&struct {
		UserID string          `gorm:"column:user_id"`
		Amount decimal.Decimal `gorm:"column:amount"`
		Reward decimal.Decimal `gorm:"column:reward"`
	}{}).
		Table("stakes").
		Where("user_id = ?", userID).
		Select("COUNT(*) as stake_count, COALESCE(SUM(amount), 0) as total_staked, COALESCE(SUM(reward), 0) as total_rewards").
		Scan(&struct {
			StakeCount   int64           `json:"stake_count"`
			TotalStaked  decimal.Decimal `json:"total_staked"`
			TotalRewards decimal.Decimal `json:"total_rewards"`
		}{
			StakeCount:   stats.StakeCount,
			TotalStaked:  stats.TotalStaked,
			TotalRewards: stats.TotalRewards,
		})

	return stats, nil
}
