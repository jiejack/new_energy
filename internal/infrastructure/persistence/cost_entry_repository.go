package persistence

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

var _ repository.CostEntryRepository = (*CostEntryRepository)(nil)

// CostEntryRepository 成本条目仓储实现
type CostEntryRepository struct {
	db *Database
}

// NewCostEntryRepository 创建成本条目仓储实例
func NewCostEntryRepository(db *Database) repository.CostEntryRepository {
	return &CostEntryRepository{db: db}
}

// Create 创建成本条目
func (r *CostEntryRepository) Create(ctx context.Context, entry *entity.CostEntry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

// Update 更新成本条目
func (r *CostEntryRepository) Update(ctx context.Context, entry *entity.CostEntry) error {
	return r.db.WithContext(ctx).Save(entry).Error
}

// Delete 删除成本条目
func (r *CostEntryRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.CostEntry{}, "id = ?", id).Error
}

// GetByID 根据ID获取成本条目
func (r *CostEntryRepository) GetByID(ctx context.Context, id string) (*entity.CostEntry, error) {
	var entry entity.CostEntry
	err := r.db.WithContext(ctx).Preload("CostCategory").First(&entry, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// GetByCode 根据编码获取成本条目
func (r *CostEntryRepository) GetByCode(ctx context.Context, code string) (*entity.CostEntry, error) {
	var entry entity.CostEntry
	err := r.db.WithContext(ctx).Preload("CostCategory").First(&entry, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// List 列出成本条目
func (r *CostEntryRepository) List(ctx context.Context, categoryID *string, startDate, endDate *time.Time, status *string, offset, limit int) ([]*entity.CostEntry, int64, error) {
	var entries []*entity.CostEntry
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.CostEntry{})

	if categoryID != nil {
		query = query.Where("cost_category_id = ?", *categoryID)
	}

	if startDate != nil {
		query = query.Where("date >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("date <= ?", *endDate)
	}

	if status != nil {
		query = query.Where("approval_status = ?", *status)
	}

	// 计算总数
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	err := query.Preload("CostCategory").Offset(offset).Limit(limit).Order("date DESC").Find(&entries).Error
	if err != nil {
		return nil, 0, err
	}

	return entries, count, nil
}

// GetTotalByCategory 根据成本类别获取总金额
func (r *CostEntryRepository) GetTotalByCategory(ctx context.Context, categoryID string, startDate, endDate *time.Time) (float64, error) {
	var total float64
	query := r.db.WithContext(ctx).Model(&entity.CostEntry{}).Where("cost_category_id = ?", categoryID)

	if startDate != nil {
		query = query.Where("date >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("date <= ?", *endDate)
	}

	err := query.Select("COALESCE(SUM(amount), 0) as total").Scan(&total).Error
	return total, err
}

// GetTotalByPeriod 根据时间段获取总金额
func (r *CostEntryRepository) GetTotalByPeriod(ctx context.Context, startDate, endDate *time.Time) (float64, error) {
	var total float64
	query := r.db.WithContext(ctx).Model(&entity.CostEntry{})

	if startDate != nil {
		query = query.Where("date >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("date <= ?", *endDate)
	}

	err := query.Select("COALESCE(SUM(amount), 0) as total").Scan(&total).Error
	return total, err
}
