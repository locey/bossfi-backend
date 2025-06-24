package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// User 用户实体
type User struct {
	ID                string          `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	WalletAddress     string          `gorm:"type:varchar(42);uniqueIndex;not null" json:"wallet_address"`
	Username          *string         `gorm:"type:varchar(50);uniqueIndex" json:"username"`
	Email             *string         `gorm:"type:varchar(255);uniqueIndex" json:"email"`
	Avatar            *string         `gorm:"type:varchar(500)" json:"avatar"`
	Bio               *string         `gorm:"type:text" json:"bio"`
	BossBalance       decimal.Decimal `gorm:"type:decimal(36,18);default:0" json:"boss_balance"`
	StakedAmount      decimal.Decimal `gorm:"type:decimal(36,18);default:0" json:"staked_amount"`
	RewardBalance     decimal.Decimal `gorm:"type:decimal(36,18);default:0" json:"reward_balance"`
	IsProfileComplete bool            `gorm:"type:boolean;default:false" json:"is_profile_complete"`
	LastLoginAt       *time.Time      `gorm:"type:timestamptz" json:"last_login_at"`
	CreatedAt         time.Time       `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time       `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt         gorm.DeletedAt  `gorm:"index" json:"-"`
}

// BeforeCreate GORM钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsProfileCompleted 检查资料是否完善
func (u *User) IsProfileCompleted() bool {
	return u.Username != nil && *u.Username != "" &&
		u.Email != nil && *u.Email != ""
}

// AddBossBalance 增加BOSS币余额
func (u *User) AddBossBalance(amount decimal.Decimal) {
	u.BossBalance = u.BossBalance.Add(amount)
}

// SubBossBalance 减少BOSS币余额
func (u *User) SubBossBalance(amount decimal.Decimal) error {
	if u.BossBalance.LessThan(amount) {
		return ErrInsufficientBalance
	}
	u.BossBalance = u.BossBalance.Sub(amount)
	return nil
}

// HasSufficientBalance 检查BOSS币余额是否充足
func (u *User) HasSufficientBalance(amount decimal.Decimal) bool {
	return u.BossBalance.GreaterThanOrEqual(amount)
}

// AddStakedAmount 增加质押金额
func (u *User) AddStakedAmount(amount decimal.Decimal) {
	u.StakedAmount = u.StakedAmount.Add(amount)
}

// SubStakedAmount 减少质押金额
func (u *User) SubStakedAmount(amount decimal.Decimal) error {
	if u.StakedAmount.LessThan(amount) {
		return ErrInsufficientStakedAmount
	}
	u.StakedAmount = u.StakedAmount.Sub(amount)
	return nil
}

// AddRewardBalance 增加奖励余额
func (u *User) AddRewardBalance(amount decimal.Decimal) {
	u.RewardBalance = u.RewardBalance.Add(amount)
}

// UpdateLastLogin 更新最后登录时间
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}
