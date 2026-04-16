package persistence

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

var _ repository.AssetMaintenanceRepository = (*AssetMaintenanceRepository)(nil)

// AssetMaintenanceRepository 资产维护记录仓储实现
type AssetMaintenanceRepository struct {
	db *Database
}

// NewAssetMaintenanceRepository 创建资产维护记录仓储实例
func NewAssetMaintenanceRepository(db *Database) repository.AssetMaintenanceRepository {
	return &AssetMaintenanceRepository{db: db}
}

// Create 创建资产维护记录
func (r *AssetMaintenanceRepository) Create(ctx context.Context, record *entity.AssetMaintenanceRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// Update 更新资产维护记录
func (r *AssetMaintenanceRepository) Update(ctx context.Context, record *entity.AssetMaintenanceRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

// Delete 删除资产维护记录
func (r *AssetMaintenanceRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.AssetMaintenanceRecord{}, "id = ?", id).Error
}

// GetByID 根据ID获取资产维护记录
func (r *AssetMaintenanceRepository) GetByID(ctx context.Context, id string) (*entity.AssetMaintenanceRecord, error) {
	var record entity.AssetMaintenanceRecord
	err := r.db.WithContext(ctx).Preload("Asset").First(&record, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListByAssetID 根据资产ID列出维护记录
func (r *AssetMaintenanceRepository) ListByAssetID(ctx context.Context, assetID string, status *string, maintenanceType *string, offset, limit int) ([]*entity.AssetMaintenanceRecord, int64, error) {
	var records []*entity.AssetMaintenanceRecord
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.AssetMaintenanceRecord{}).Where("asset_id = ?", assetID)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if maintenanceType != nil {
		query = query.Where("maintenance_type = ?", *maintenanceType)
	}

	// 计算总数
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	err := query.Preload("Asset").Offset(offset).Limit(limit).Order("start_date DESC").Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, count, nil
}

// ListByStatus 根据状态列出维护记录
func (r *AssetMaintenanceRepository) ListByStatus(ctx context.Context, status string, offset, limit int) ([]*entity.AssetMaintenanceRecord, int64, error) {
	var records []*entity.AssetMaintenanceRecord
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.AssetMaintenanceRecord{}).Where("status = ?", status)

	// 计算总数
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	err := query.Preload("Asset").Offset(offset).Limit(limit).Order("start_date DESC").Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, count, nil
}

// GetMaintenanceCostByAsset 根据资产ID获取维护成本
func (r *AssetMaintenanceRepository) GetMaintenanceCostByAsset(ctx context.Context, assetID string, startDate, endDate *time.Time) (float64, error) {
	var total float64
	query := r.db.WithContext(ctx).Model(&entity.AssetMaintenanceRecord{}).Where("asset_id = ?", assetID)

	if startDate != nil {
		query = query.Where("start_date >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("start_date <= ?", *endDate)
	}

	err := query.Select("COALESCE(SUM(cost), 0) as total").Scan(&total).Error
	return total, err
}
