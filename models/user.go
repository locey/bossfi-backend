package models

import (
	"time"
)

type User struct {
	ID                uint       `json:"id" gorm:"primarykey"`
	WalletAddress     string     `json:"wallet_address" gorm:"type:varchar(42);unique;not null"`
	Username          *string    `json:"username" gorm:"type:varchar(50);unique"`
	Email             *string    `json:"email" gorm:"type:varchar(255);unique"`
	Avatar            *string    `json:"avatar" gorm:"type:varchar(500)"`
	Bio               *string    `json:"bio" gorm:"type:text"`
	BossBalance       float64    `json:"boss_balance" gorm:"type:decimal(36,18);default:0"`
	StakedAmount      float64    `json:"staked_amount" gorm:"type:decimal(36,18);default:0"`
	RewardBalance     float64    `json:"reward_balance" gorm:"type:decimal(36,18);default:0"`
	IsProfileComplete bool       `json:"is_profile_complete" gorm:"default:false"`
	LastLoginAt       *time.Time `json:"last_login_at" gorm:"type:timestamptz"`
	CreatedAt         time.Time  `json:"created_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
}

func (u *User) TableName() string {
	return "users"
}
