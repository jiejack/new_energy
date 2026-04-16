package persistence

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

var _ repository.CostAllocationRepository = (*CostAllocationRepository)(nil)

// CostAllocationRepository 成本分配仓储实现
type CostAllocationRepository struct {
	db *Database
}

// NewCostAllocationRepository 创建成本分配仓储实例
func NewCostAllocationRepository(db *Database) repository.CostAllocationRepository {
	return &CostAllocationRepository{db: db}
}

// Create 创建成本分配
func (r *CostAllocationRepository) Create(ctx context.Context, allocation *entity.CostAllocation) error {
	return r.db.WithContext(ctx).Create(allocation).Error
}

// Update 更新成本分配
func (r *CostAllocationRepository) Update(ctx context.Context, allocation *entity.CostAllocation) error {
	return r.db.WithContext(ctx).Save(allocation).Error
}

// Delete 删除成本分配
func (r *CostAllocationRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.CostAllocation{}, "id = ?", id).Error
}

// GetByID 根据ID获取成本分配
func (r *CostAllocationRepository) GetByID(ctx context.Context, id string) (*entity.CostAllocation, error) {
	var allocation entity.CostAllocation
	err := r.db.WithContext(ctx).Preload("CostEntry").First(&allocation, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &allocation, nil
}

// ListByCostEntryID 根据成本条目ID列出成本分配
func (r *CostAllocationRepository) ListByCostEntryID(ctx context.Context, costEntryID string) ([]*entity.CostAllocation, error) {
	var allocations []*entity.CostAllocation
	err := r.db.WithContext(ctx).Preload("CostEntry").Where("cost_entry_id = ?", costEntryID).Find(&allocations).Error
	return allocations, err
}

// ListByAllocated 根据分配对象列出成本分配
func (r *CostAllocationRepository) ListByAllocated(ctx context.Context, allocatedTo, allocatedID string) ([]*entity.CostAllocation, error) {
	var allocations []*entity.CostAllocation
	err := r.db.WithContext(ctx).Preload("CostEntry").Where("allocated_to = ? AND allocated_id = ?", allocatedTo, allocatedID).Find(&allocations).Error
	return allocations, err
}

// GetTotalByAllocated 根据分配对象获取总金额
func (r *CostAllocationRepository) GetTotalByAllocated(ctx context.Context, allocatedTo, allocatedID string, startDate, endDate *time.Time) (float64, error) {
	var total float64
	query := r.db.WithContext(ctx).Model(&entity.CostAllocation{}).Where("allocated_to = ? AND allocated_id = ?", allocatedTo, allocatedID)

	if startDate != nil || endDate != nil {
		query = query.Joins("JOIN cost_entries ON cost_allocations.cost_entry_id = cost_entries.id")
		if startDate != nil {
			query = query.Where("cost_entries.date >= ?", *startDate)
		}
		if endDate != nil {
			query = query.Where("cost_entries.date <= ?", *endDate)
		}
	}

	err := query.Select("COALESCE(SUM(amount), 0) as total").Scan(&total).Error
	return total, err
}
