package persistence

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

var _ repository.AssetRepository = (*AssetRepository)(nil)

// AssetRepository 资产仓储实现
type AssetRepository struct {
	db *Database
}

// NewAssetRepository 创建资产仓储实例
func NewAssetRepository(db *Database) repository.AssetRepository {
	return &AssetRepository{db: db}
}

// Create 创建资产
func (r *AssetRepository) Create(ctx context.Context, asset *entity.Asset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

// Update 更新资产
func (r *AssetRepository) Update(ctx context.Context, asset *entity.Asset) error {
	return r.db.WithContext(ctx).Save(asset).Error
}

// Delete 删除资产
func (r *AssetRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Asset{}, "id = ?", id).Error
}

// GetByID 根据ID获取资产
func (r *AssetRepository) GetByID(ctx context.Context, id string) (*entity.Asset, error) {
	var asset entity.Asset
	err := r.db.WithContext(ctx).Preload("MaintenanceRecords").Preload("DepreciationRecords").Preload("Documents").First(&asset, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// GetByCode 根据编码获取资产
func (r *AssetRepository) GetByCode(ctx context.Context, code string) (*entity.Asset, error) {
	var asset entity.Asset
	err := r.db.WithContext(ctx).Preload("MaintenanceRecords").Preload("DepreciationRecords").Preload("Documents").First(&asset, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// List 列出资产
func (r *AssetRepository) List(ctx context.Context, assetType *string, status *string, category *string, offset, limit int) ([]*entity.Asset, int64, error) {
	var assets []*entity.Asset
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.Asset{})

	if assetType != nil {
		query = query.Where("asset_type = ?", *assetType)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if category != nil {
		query = query.Where("category = ?", *category)
	}

	// 计算总数
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	err := query.Preload("MaintenanceRecords").Preload("DepreciationRecords").Preload("Documents").Offset(offset).Limit(limit).Order("created_at DESC").Find(&assets).Error
	if err != nil {
		return nil, 0, err
	}

	return assets, count, nil
}

// GetByLocation 根据位置获取资产
func (r *AssetRepository) GetByLocation(ctx context.Context, location string) ([]*entity.Asset, error) {
	var assets []*entity.Asset
	err := r.db.WithContext(ctx).Where("location = ?", location).Find(&assets).Error
	return assets, err
}

// GetByDepartment 根据部门获取资产
func (r *AssetRepository) GetByDepartment(ctx context.Context, departmentID string) ([]*entity.Asset, error) {
	var assets []*entity.Asset
	err := r.db.WithContext(ctx).Where("department_id = ?", departmentID).Find(&assets).Error
	return assets, err
}

// GetByResponsiblePerson 根据负责人获取资产
func (r *AssetRepository) GetByResponsiblePerson(ctx context.Context, person string) ([]*entity.Asset, error) {
	var assets []*entity.Asset
	err := r.db.WithContext(ctx).Where("responsible_person = ?", person).Find(&assets).Error
	return assets, err
}

// GetDepreciatingAssets 获取正在折旧的资产
func (r *AssetRepository) GetDepreciatingAssets(ctx context.Context) ([]*entity.Asset, error) {
	var assets []*entity.Asset
	err := r.db.WithContext(ctx).Where("status = ? AND depreciation_method IS NOT NULL", "active").Find(&assets).Error
	return assets, err
}

// GetAssetsNearWarrantyEnd 获取 warranty 即将到期的资产
func (r *AssetRepository) GetAssetsNearWarrantyEnd(ctx context.Context, days int) ([]*entity.Asset, error) {
	var assets []*entity.Asset
	endDate := time.Now().AddDate(0, 0, days)
	err := r.db.WithContext(ctx).Where("warranty_end_date IS NOT NULL AND warranty_end_date BETWEEN ? AND ?", time.Now(), endDate).Find(&assets).Error
	return assets, err
}
