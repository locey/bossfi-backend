package repository

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"bossfi-blockchain-backend/internal/domain/stake"
)

// StakeRepository 质押仓储接口
type StakeRepository interface {
	Create(ctx context.Context, s *stake.Stake) error
	GetByID(ctx context.Context, id string) (*stake.Stake, error)
	Update(ctx context.Context, s *stake.Stake) error
	Delete(ctx context.Context, id string) error
	GetByUserID(ctx context.Context, userID string, offset, limit int) ([]*stake.Stake, int64, error)
	GetActiveStakes(ctx context.Context, userID string) ([]*stake.Stake, error)
	GetUnstakingStakes(ctx context.Context, userID string) ([]*stake.Stake, error)
	GetStakesReadyForReward(ctx context.Context) ([]*stake.Stake, error)
	GetStakesReadyForUnstake(ctx context.Context, delayDays int) ([]*stake.Stake, error)
	GetTotalStaked(ctx context.Context, userID string) (decimal.Decimal, error)
	GetTotalRewards(ctx context.Context, userID string) (decimal.Decimal, error)
	GetStakeStats(ctx context.Context) (*StakeStats, error)
	GetUserStakeStats(ctx context.Context, userID string) (*UserStakeStats, error)
}

// StakeStats 质押统计信息
type StakeStats struct {
	TotalStakes     int64           `json:"total_stakes"`
	ActiveStakes    int64           `json:"active_stakes"`
	UnstakingStakes int64           `json:"unstaking_stakes"`
	CompletedStakes int64           `json:"completed_stakes"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	TotalRewards    decimal.Decimal `json:"total_rewards"`
	AverageAmount   decimal.Decimal `json:"average_amount"`
}

// UserStakeStats 用户质押统计信息
type UserStakeStats struct {
	UserID          string          `json:"user_id"`
	TotalStakes     int64           `json:"total_stakes"`
	ActiveStakes    int64           `json:"active_stakes"`
	UnstakingStakes int64           `json:"unstaking_stakes"`
	CompletedStakes int64           `json:"completed_stakes"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	ActiveAmount    decimal.Decimal `json:"active_amount"`
	TotalRewards    decimal.Decimal `json:"total_rewards"`
	PendingRewards  decimal.Decimal `json:"pending_rewards"`
}

// stakeRepository 质押仓储实现
type stakeRepository struct {
	db *gorm.DB
}

// NewStakeRepository 创建质押仓储
func NewStakeRepository(db *gorm.DB) StakeRepository {
	return &stakeRepository{db: db}
}

// Create 创建质押
func (r *stakeRepository) Create(ctx context.Context, s *stake.Stake) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// GetByID 根据ID获取质押
func (r *stakeRepository) GetByID(ctx context.Context, id string) (*stake.Stake, error) {
	var s stake.Stake
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, stake.ErrStakeNotFound
		}
		return nil, err
	}
	return &s, nil
}

// Update 更新质押
func (r *stakeRepository) Update(ctx context.Context, s *stake.Stake) error {
	return r.db.WithContext(ctx).Save(s).Error
}

// Delete 删除质押
func (r *stakeRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&stake.Stake{}, "id = ?", id).Error
}

// GetByUserID 获取用户的质押记录
func (r *stakeRepository) GetByUserID(ctx context.Context, userID string, offset, limit int) ([]*stake.Stake, int64, error) {
	var stakes []*stake.Stake
	var total int64

	query := r.db.WithContext(ctx).Model(&stake.Stake{}).Where("user_id = ?", userID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&stakes).Error

	return stakes, total, err
}

// GetActiveStakes 获取用户的活跃质押
func (r *stakeRepository) GetActiveStakes(ctx context.Context, userID string) ([]*stake.Stake, error) {
	var stakes []*stake.Stake
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, stake.StakeStatusActive).
		Order("created_at DESC").
		Find(&stakes).Error
	return stakes, err
}

// GetUnstakingStakes 获取用户正在解质押的记录
func (r *stakeRepository) GetUnstakingStakes(ctx context.Context, userID string) ([]*stake.Stake, error) {
	var stakes []*stake.Stake
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, stake.StakeStatusUnstaking).
		Order("unstake_request_at DESC").
		Find(&stakes).Error
	return stakes, err
}

// GetStakesReadyForReward 获取可以获得奖励的质押记录
func (r *stakeRepository) GetStakesReadyForReward(ctx context.Context) ([]*stake.Stake, error) {
	var stakes []*stake.Stake

	// 查找活跃状态且距离上次奖励超过1天的质押记录
	yesterday := time.Now().AddDate(0, 0, -1)

	err := r.db.WithContext(ctx).
		Where("status = ? AND last_reward_at <= ?", stake.StakeStatusActive, yesterday).
		Find(&stakes).Error

	return stakes, err
}

// GetStakesReadyForUnstake 获取可以完成解质押的记录
func (r *stakeRepository) GetStakesReadyForUnstake(ctx context.Context, delayDays int) ([]*stake.Stake, error) {
	var stakes []*stake.Stake

	// 计算延迟时间
	delayTime := time.Now().AddDate(0, 0, -delayDays)

	err := r.db.WithContext(ctx).
		Where("status = ? AND unstake_request_at <= ?", stake.StakeStatusUnstaking, delayTime).
		Find(&stakes).Error

	return stakes, err
}

// GetTotalStaked 获取用户总质押金额
func (r *stakeRepository) GetTotalStaked(ctx context.Context, userID string) (decimal.Decimal, error) {
	var result struct {
		Total decimal.Decimal `gorm:"column:total"`
	}

	err := r.db.WithContext(ctx).
		Model(&stake.Stake{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND status = ?", userID, stake.StakeStatusActive).
		Scan(&result).Error

	return result.Total, err
}

// GetTotalRewards 获取用户总奖励
func (r *stakeRepository) GetTotalRewards(ctx context.Context, userID string) (decimal.Decimal, error) {
	var result struct {
		Total decimal.Decimal `gorm:"column:total"`
	}

	err := r.db.WithContext(ctx).
		Model(&stake.Stake{}).
		Select("COALESCE(SUM(reward_earned), 0) as total").
		Where("user_id = ?", userID).
		Scan(&result).Error

	return result.Total, err
}

// GetStakeStats 获取质押统计信息
func (r *stakeRepository) GetStakeStats(ctx context.Context) (*StakeStats, error) {
	var stats StakeStats

	// 获取基本统计
	err := r.db.WithContext(ctx).
		Model(&stake.Stake{}).
		Select(`
			COUNT(*) as total_stakes,
			SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) as active_stakes,
			SUM(CASE WHEN status = 'unstaking' THEN 1 ELSE 0 END) as unstaking_stakes,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_stakes,
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(SUM(reward_earned), 0) as total_rewards,
			COALESCE(AVG(amount), 0) as average_amount
		`).
		Scan(&stats).Error

	return &stats, err
}

// GetUserStakeStats 获取用户质押统计信息
func (r *stakeRepository) GetUserStakeStats(ctx context.Context, userID string) (*UserStakeStats, error) {
	var stats UserStakeStats
	stats.UserID = userID

	// 获取用户质押统计
	err := r.db.WithContext(ctx).
		Model(&stake.Stake{}).
		Select(`
			COUNT(*) as total_stakes,
			SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) as active_stakes,
			SUM(CASE WHEN status = 'unstaking' THEN 1 ELSE 0 END) as unstaking_stakes,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_stakes,
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(SUM(CASE WHEN status = 'active' THEN amount ELSE 0 END), 0) as active_amount,
			COALESCE(SUM(reward_earned), 0) as total_rewards
		`).
		Where("user_id = ?", userID).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	// 计算待领取奖励（这里简化处理，实际应该根据时间和利率计算）
	stats.PendingRewards = decimal.Zero

	return &stats, nil
}
