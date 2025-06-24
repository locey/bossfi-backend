package stake

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// StakeStatus 质押状态
type StakeStatus string

const (
	StakeStatusActive    StakeStatus = "active"    // 活跃
	StakeStatusUnstaking StakeStatus = "unstaking" // 解质押中
	StakeStatusCompleted StakeStatus = "completed" // 已完成
)

// Stake 质押实体
type Stake struct {
	ID               string          `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID           string          `gorm:"type:uuid;not null;index" json:"user_id"`
	Amount           decimal.Decimal `gorm:"type:decimal(36,18);not null" json:"amount"`
	RewardEarned     decimal.Decimal `gorm:"type:decimal(36,18);default:0" json:"reward_earned"`
	Status           StakeStatus     `gorm:"type:stake_status_enum;default:'active'" json:"status"`
	StakedAt         time.Time       `gorm:"type:timestamptz;not null" json:"staked_at"`
	UnstakeRequestAt *time.Time      `gorm:"type:timestamptz" json:"unstake_request_at"`
	UnstakedAt       *time.Time      `gorm:"type:timestamptz" json:"unstaked_at"`
	LastRewardAt     time.Time       `gorm:"type:timestamptz;not null" json:"last_reward_at"`
	CreatedAt        time.Time       `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time       `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt        gorm.DeletedAt  `gorm:"index" json:"-"`
}

// BeforeCreate GORM钩子
func (s *Stake) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	now := time.Now()
	if s.StakedAt.IsZero() {
		s.StakedAt = now
	}
	if s.LastRewardAt.IsZero() {
		s.LastRewardAt = now
	}

	return nil
}

// TableName 指定表名
func (Stake) TableName() string {
	return "stakes"
}

// IsActive 检查是否活跃
func (s *Stake) IsActive() bool {
	return s.Status == StakeStatusActive
}

// IsUnstaking 检查是否正在解质押
func (s *Stake) IsUnstaking() bool {
	return s.Status == StakeStatusUnstaking
}

// IsCompleted 检查是否已完成
func (s *Stake) IsCompleted() bool {
	return s.Status == StakeStatusCompleted
}

// CanUnstake 检查是否可以解质押
func (s *Stake) CanUnstake() bool {
	return s.Status == StakeStatusActive
}

// RequestUnstake 请求解质押
func (s *Stake) RequestUnstake() error {
	if !s.CanUnstake() {
		return ErrCannotUnstake
	}

	now := time.Now()
	s.Status = StakeStatusUnstaking
	s.UnstakeRequestAt = &now
	return nil
}

// CompleteUnstake 完成解质押
func (s *Stake) CompleteUnstake() error {
	if s.Status != StakeStatusUnstaking {
		return ErrCannotCompleteUnstake
	}

	now := time.Now()
	s.Status = StakeStatusCompleted
	s.UnstakedAt = &now
	return nil
}

// AddReward 添加奖励
func (s *Stake) AddReward(amount decimal.Decimal) {
	s.RewardEarned = s.RewardEarned.Add(amount)
	s.LastRewardAt = time.Now()
}

// CalculateDaysSinceStaked 计算质押天数
func (s *Stake) CalculateDaysSinceStaked() int {
	return int(time.Since(s.StakedAt).Hours() / 24)
}

// CalculateDaysSinceLastReward 计算距离上次奖励的天数
func (s *Stake) CalculateDaysSinceLastReward() int {
	return int(time.Since(s.LastRewardAt).Hours() / 24)
}

// CanReceiveReward 检查是否可以获得奖励
func (s *Stake) CanReceiveReward() bool {
	return s.IsActive() && s.CalculateDaysSinceLastReward() >= 1
}

// GetUnstakeDelay 获取解质押延迟时间
func (s *Stake) GetUnstakeDelay(delayDays int) time.Duration {
	if s.UnstakeRequestAt == nil {
		return 0
	}

	targetTime := s.UnstakeRequestAt.AddDate(0, 0, delayDays)
	now := time.Now()

	if now.After(targetTime) {
		return 0
	}

	return targetTime.Sub(now)
}

// CanCompleteUnstake 检查是否可以完成解质押
func (s *Stake) CanCompleteUnstake(delayDays int) bool {
	return s.IsUnstaking() && s.GetUnstakeDelay(delayDays) == 0
}
