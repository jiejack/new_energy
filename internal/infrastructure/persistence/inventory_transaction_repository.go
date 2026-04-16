package persistence

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// InventoryTransactionRepository 库存交易仓储实现
type InventoryTransactionRepository struct {
	db *Database
}

// NewInventoryTransactionRepository 创建库存交易仓储
func NewInventoryTransactionRepository(db *Database) repository.InventoryTransactionRepository {
	return &InventoryTransactionRepository{db: db}
}

// Create 创建库存交易
func (r *InventoryTransactionRepository) Create(ctx context.Context, transaction *entity.InventoryTransaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

// GetByID 根据ID获取库存交易
func (r *InventoryTransactionRepository) GetByID(ctx context.Context, id string) (*entity.InventoryTransaction, error) {
	var transaction entity.InventoryTransaction
	err := r.db.WithContext(ctx).First(&transaction, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// ListByInventoryID 根据库存ID获取交易列表
func (r *InventoryTransactionRepository) ListByInventoryID(ctx context.Context, inventoryID string) ([]*entity.InventoryTransaction, error) {
	var transactions []*entity.InventoryTransaction
	err := r.db.WithContext(ctx).Where("inventory_id = ?", inventoryID).Order("created_at DESC").Find(&transactions).Error
	return transactions, err
}

// ListByReference 根据参考ID和类型获取交易列表
func (r *InventoryTransactionRepository) ListByReference(ctx context.Context, referenceID string, referenceType string) ([]*entity.InventoryTransaction, error) {
	var transactions []*entity.InventoryTransaction
	query := r.db.WithContext(ctx).Where("reference_id = ?", referenceID)
	if referenceType != "" {
		query = query.Where("reference_type = ?", referenceType)
	}
	err := query.Order("created_at DESC").Find(&transactions).Error
	return transactions, err
}

// GetTransactionHistory 获取库存交易历史
func (r *InventoryTransactionRepository) GetTransactionHistory(ctx context.Context, inventoryID string, limit int) ([]*entity.InventoryTransaction, error) {
	var transactions []*entity.InventoryTransaction
	query := r.db.WithContext(ctx).Where("inventory_id = ?", inventoryID)
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("created_at DESC").Find(&transactions).Error
	return transactions, err
}
