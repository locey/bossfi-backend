package transaction

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Transaction 交易实体
type Transaction struct {
	ID            string           `gorm:"type:varchar(36);primaryKey" json:"id"`
	Hash          string           `gorm:"type:varchar(100);uniqueIndex" json:"hash"`
	FromAddress   string           `gorm:"type:varchar(100);not null;index" json:"from_address"`
	ToAddress     string           `gorm:"type:varchar(100);not null;index" json:"to_address"`
	Amount        decimal.Decimal  `gorm:"type:decimal(36,18);not null" json:"amount"`
	Fee           decimal.Decimal  `gorm:"type:decimal(36,18);default:0" json:"fee"`
	Network       string           `gorm:"type:varchar(20);not null" json:"network"`
	Type          Type             `gorm:"type:varchar(20);not null" json:"type"`
	Status        Status           `gorm:"type:tinyint;default:1" json:"status"`
	BlockNumber   *uint64          `gorm:"index" json:"block_number"`
	BlockHash     string           `gorm:"type:varchar(100);index" json:"block_hash"`
	Confirmations uint64           `gorm:"default:0" json:"confirmations"`
	GasUsed       *uint64          `json:"gas_used"`
	GasPrice      *decimal.Decimal `gorm:"type:decimal(36,18)" json:"gas_price"`
	Nonce         *uint64          `json:"nonce"`
	Memo          string           `gorm:"type:text" json:"memo"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	DeletedAt     gorm.DeletedAt   `gorm:"index" json:"-"`
}

// Type 交易类型
type Type string

const (
	TypeDeposit   Type = "deposit"   // 充值
	TypeWithdraw  Type = "withdraw"  // 提现
	TypeTransfer  Type = "transfer"  // 转账
	TypeTrade     Type = "trade"     // 交易
	TypeStaking   Type = "staking"   // 质押
	TypeUnstaking Type = "unstaking" // 解押
	TypeReward    Type = "reward"    // 奖励
)

// Status 交易状态
type Status int

const (
	StatusPending   Status = 1 // 待确认
	StatusConfirmed Status = 2 // 已确认
	StatusFailed    Status = 3 // 失败
	StatusCancelled Status = 4 // 已取消
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusConfirmed:
		return "confirmed"
	case StatusFailed:
		return "failed"
	case StatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// BeforeCreate GORM钩子
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (Transaction) TableName() string {
	return "transactions"
}

// IsConfirmed 判断交易是否已确认
func (t *Transaction) IsConfirmed() bool {
	return t.Status == StatusConfirmed
}

// IsPending 判断交易是否待确认
func (t *Transaction) IsPending() bool {
	return t.Status == StatusPending
}

// UpdateConfirmations 更新确认数
func (t *Transaction) UpdateConfirmations(confirmations uint64) {
	t.Confirmations = confirmations
	if confirmations >= 6 && t.Status == StatusPending {
		t.Status = StatusConfirmed
	}
}
