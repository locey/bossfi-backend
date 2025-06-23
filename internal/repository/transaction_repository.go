package repository

import (
	"context"

	"bossfi-blockchain-backend/internal/domain/transaction"
	"bossfi-blockchain-backend/pkg/database"

	"gorm.io/gorm"
)

// TransactionRepository 交易仓储接口
type TransactionRepository interface {
	Create(ctx context.Context, tx *transaction.Transaction) error
	GetByID(ctx context.Context, id string) (*transaction.Transaction, error)
	GetByHash(ctx context.Context, hash string) (*transaction.Transaction, error)
	GetByAddress(ctx context.Context, address string, offset, limit int) ([]*transaction.Transaction, int64, error)
	GetByAddressAndStatus(ctx context.Context, address string, status transaction.Status) ([]*transaction.Transaction, error)
	Update(ctx context.Context, tx *transaction.Transaction) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*transaction.Transaction, int64, error)
	GetPendingTransactions(ctx context.Context) ([]*transaction.Transaction, error)
	UpdateStatus(ctx context.Context, id string, status transaction.Status) error
	UpdateConfirmations(ctx context.Context, id string, confirmations uint64) error
	ExistsByHash(ctx context.Context, hash string) (bool, error)
}

// transactionRepository 交易仓储实现
type transactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository 创建交易仓储实例
func NewTransactionRepository() TransactionRepository {
	return &transactionRepository{
		db: database.GetDB(),
	}
}

func (r *transactionRepository) Create(ctx context.Context, tx *transaction.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *transactionRepository) GetByID(ctx context.Context, id string) (*transaction.Transaction, error) {
	var tx transaction.Transaction
	err := r.db.WithContext(ctx).First(&tx, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, transaction.ErrTransactionNotFound
		}
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) GetByHash(ctx context.Context, hash string) (*transaction.Transaction, error) {
	var tx transaction.Transaction
	err := r.db.WithContext(ctx).First(&tx, "hash = ?", hash).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, transaction.ErrTransactionNotFound
		}
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) GetByAddress(ctx context.Context, address string, offset, limit int) ([]*transaction.Transaction, int64, error) {
	var transactions []*transaction.Transaction
	var total int64

	// 获取总数
	query := r.db.WithContext(ctx).Model(&transaction.Transaction{}).
		Where("from_address = ? OR to_address = ?", address, address)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := query.Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (r *transactionRepository) GetByAddressAndStatus(ctx context.Context, address string, status transaction.Status) ([]*transaction.Transaction, error) {
	var transactions []*transaction.Transaction
	err := r.db.WithContext(ctx).
		Where("(from_address = ? OR to_address = ?) AND status = ?", address, address, status).
		Order("created_at DESC").
		Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) Update(ctx context.Context, tx *transaction.Transaction) error {
	return r.db.WithContext(ctx).Save(tx).Error
}

func (r *transactionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&transaction.Transaction{}, "id = ?", id).Error
}

func (r *transactionRepository) List(ctx context.Context, offset, limit int) ([]*transaction.Transaction, int64, error) {
	var transactions []*transaction.Transaction
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&transaction.Transaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := r.db.WithContext(ctx).Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (r *transactionRepository) GetPendingTransactions(ctx context.Context) ([]*transaction.Transaction, error) {
	var transactions []*transaction.Transaction
	err := r.db.WithContext(ctx).
		Where("status = ?", transaction.StatusPending).
		Order("created_at ASC").
		Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) UpdateStatus(ctx context.Context, id string, status transaction.Status) error {
	return r.db.WithContext(ctx).Model(&transaction.Transaction{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *transactionRepository) UpdateConfirmations(ctx context.Context, id string, confirmations uint64) error {
	updates := map[string]interface{}{
		"confirmations": confirmations,
	}

	// 如果确认数达到6个，更新状态为已确认
	if confirmations >= 6 {
		updates["status"] = transaction.StatusConfirmed
	}

	return r.db.WithContext(ctx).Model(&transaction.Transaction{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *transactionRepository) ExistsByHash(ctx context.Context, hash string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&transaction.Transaction{}).Where("hash = ?", hash).Count(&count).Error
	return count > 0, err
}
