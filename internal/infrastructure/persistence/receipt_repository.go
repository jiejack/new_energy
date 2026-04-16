package persistence

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

var _ repository.ReceiptRepository = (*ReceiptRepository)(nil)

// ReceiptRepository 收货单仓储实现
type ReceiptRepository struct {
	db *Database
}

// NewReceiptRepository 创建收货单仓储实例
func NewReceiptRepository(db *Database) repository.ReceiptRepository {
	return &ReceiptRepository{db: db}
}

// Create 创建收货单
func (r *ReceiptRepository) Create(ctx context.Context, receipt *entity.Receipt) error {
	return r.db.WithContext(ctx).Create(receipt).Error
}

// Update 更新收货单
func (r *ReceiptRepository) Update(ctx context.Context, receipt *entity.Receipt) error {
	return r.db.WithContext(ctx).Save(receipt).Error
}

// Delete 删除收货单
func (r *ReceiptRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Receipt{}, "id = ?", id).Error
}

// GetByID 根据ID获取收货单
func (r *ReceiptRepository) GetByID(ctx context.Context, id string) (*entity.Receipt, error) {
	var receipt entity.Receipt
	err := r.db.WithContext(ctx).Preload("Items").Where("id = ?", id).First(&receipt).Error
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

// List 列出收货单
func (r *ReceiptRepository) List(ctx context.Context, purchaseOrderID *string, status *string, startDate, endDate *time.Time, offset, limit int) ([]*entity.Receipt, int64, error) {
	var receipts []*entity.Receipt
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.Receipt{})

	if purchaseOrderID != nil {
		query = query.Where("purchase_order_id = ?", *purchaseOrderID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	// 计算总数
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	err := query.Preload("Items").Offset(offset).Limit(limit).Order("created_at DESC").Find(&receipts).Error
	if err != nil {
		return nil, 0, err
	}

	return receipts, count, nil
}

// GetByCode 根据收货单编号获取收货单
func (r *ReceiptRepository) GetByCode(ctx context.Context, code string) (*entity.Receipt, error) {
	var receipt entity.Receipt
	err := r.db.WithContext(ctx).Preload("Items").Where("code = ?", code).First(&receipt).Error
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

// GetByPurchaseOrderID 根据采购订单ID获取收货单
func (r *ReceiptRepository) GetByPurchaseOrderID(ctx context.Context, purchaseOrderID string) ([]*entity.Receipt, error) {
	var receipts []*entity.Receipt
	err := r.db.WithContext(ctx).Preload("Items").Where("purchase_order_id = ?", purchaseOrderID).Find(&receipts).Error
	if err != nil {
		return nil, err
	}
	return receipts, nil
}
