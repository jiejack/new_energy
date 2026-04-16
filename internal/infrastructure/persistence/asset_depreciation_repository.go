package persistence

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

var _ repository.AssetDepreciationRepository = (*AssetDepreciationRepository)(nil)

// AssetDepreciationRepository 资产折旧记录仓储实现
type AssetDepreciationRepository struct {
	db *Database
}

// NewAssetDepreciationRepository 创建资产折旧记录仓储实例
func NewAssetDepreciationRepository(db *Database) repository.AssetDepreciationRepository {
	return &AssetDepreciationRepository{db: db}
}

// Create 创建资产折旧记录
func (r *AssetDepreciationRepository) Create(ctx context.Context, record *entity.AssetDepreciationRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// Update 更新资产折旧记录
func (r *AssetDepreciationRepository) Update(ctx context.Context, record *entity.AssetDepreciationRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

// Delete 删除资产折旧记录
func (r *AssetDepreciationRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.AssetDepreciationRecord{}, "id = ?", id).Error
}

// GetByID 根据ID获取资产折旧记录
func (r *AssetDepreciationRepository) GetByID(ctx context.Context, id string) (*entity.AssetDepreciationRecord, error) {
	var record entity.AssetDepreciationRecord
	err := r.db.WithContext(ctx).Preload("Asset").First(&record, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListByAssetID 根据资产ID列出折旧记录
func (r *AssetDepreciationRepository) ListByAssetID(ctx context.Context, assetID string, period *string, offset, limit int) ([]*entity.AssetDepreciationRecord, int64, error) {
	var records []*entity.AssetDepreciationRecord
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.AssetDepreciationRecord{}).Where("asset_id = ?", assetID)

	if period != nil {
		query = query.Where("period = ?", *period)
	}

	// 计算总数
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	err := query.Preload("Asset").Offset(offset).Limit(limit).Order("depreciation_date DESC").Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, count, nil
}

// GetLatestByAssetID 根据资产ID获取最新折旧记录
func (r *AssetDepreciationRepository) GetLatestByAssetID(ctx context.Context, assetID string) (*entity.AssetDepreciationRecord, error) {
	var record entity.AssetDepreciationRecord
	err := r.db.WithContext(ctx).Preload("Asset").Where("asset_id = ?", assetID).Order("depreciation_date DESC").First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetDepreciationSummaryByPeriod 根据时间段获取折旧汇总
func (r *AssetDepreciationRepository) GetDepreciationSummaryByPeriod(ctx context.Context, period string, startDate, endDate *time.Time) (float64, error) {
	var total float64
	query := r.db.WithContext(ctx).Model(&entity.AssetDepreciationRecord{}).Where("period = ?", period)

	if startDate != nil {
		query = query.Where("depreciation_date >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("depreciation_date <= ?", *endDate)
	}

	err := query.Select("COALESCE(SUM(depreciation_amount), 0) as total").Scan(&total).Error
	return total, err
}
