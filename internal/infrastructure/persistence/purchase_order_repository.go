package persistence

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

var _ repository.PurchaseOrderRepository = (*PurchaseOrderRepository)(nil)

// PurchaseOrderRepository 采购订单仓储实现
type PurchaseOrderRepository struct {
	db *Database
}

// NewPurchaseOrderRepository 创建采购订单仓储实例
func NewPurchaseOrderRepository(db *Database) repository.PurchaseOrderRepository {
	return &PurchaseOrderRepository{db: db}
}

// Create 创建采购订单
func (r *PurchaseOrderRepository) Create(ctx context.Context, order *entity.PurchaseOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// Update 更新采购订单
func (r *PurchaseOrderRepository) Update(ctx context.Context, order *entity.PurchaseOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// Delete 删除采购订单
func (r *PurchaseOrderRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.PurchaseOrder{}, "id = ?", id).Error
}

// GetByID 根据ID获取采购订单
func (r *PurchaseOrderRepository) GetByID(ctx context.Context, id string) (*entity.PurchaseOrder, error) {
	var order entity.PurchaseOrder
	err := r.db.WithContext(ctx).Preload("Items").Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// List 列出采购订单
func (r *PurchaseOrderRepository) List(ctx context.Context, supplierID *string, status *string, startDate, endDate *time.Time, offset, limit int) ([]*entity.PurchaseOrder, int64, error) {
	var orders []*entity.PurchaseOrder
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.PurchaseOrder{})

	if supplierID != nil {
		query = query.Where("supplier_id = ?", *supplierID)
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
	err := query.Preload("Items").Offset(offset).Limit(limit).Order("created_at DESC").Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, count, nil
}

// GetByCode 根据订单编号获取采购订单
func (r *PurchaseOrderRepository) GetByCode(ctx context.Context, code string) (*entity.PurchaseOrder, error) {
	var order entity.PurchaseOrder
	err := r.db.WithContext(ctx).Preload("Items").Where("code = ?", code).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}
