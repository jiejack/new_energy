package persistence

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

var _ repository.AssetDocumentRepository = (*AssetDocumentRepository)(nil)

// AssetDocumentRepository 资产文档仓储实现
type AssetDocumentRepository struct {
	db *Database
}

// NewAssetDocumentRepository 创建资产文档仓储实例
func NewAssetDocumentRepository(db *Database) repository.AssetDocumentRepository {
	return &AssetDocumentRepository{db: db}
}

// Create 创建资产文档
func (r *AssetDocumentRepository) Create(ctx context.Context, document *entity.AssetDocument) error {
	return r.db.WithContext(ctx).Create(document).Error
}

// Update 更新资产文档
func (r *AssetDocumentRepository) Update(ctx context.Context, document *entity.AssetDocument) error {
	return r.db.WithContext(ctx).Save(document).Error
}

// Delete 删除资产文档
func (r *AssetDocumentRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.AssetDocument{}, "id = ?", id).Error
}

// GetByID 根据ID获取资产文档
func (r *AssetDocumentRepository) GetByID(ctx context.Context, id string) (*entity.AssetDocument, error) {
	var document entity.AssetDocument
	err := r.db.WithContext(ctx).Preload("Asset").First(&document, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

// ListByAssetID 根据资产ID列出文档
func (r *AssetDocumentRepository) ListByAssetID(ctx context.Context, assetID string, documentType *string, offset, limit int) ([]*entity.AssetDocument, int64, error) {
	var documents []*entity.AssetDocument
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.AssetDocument{}).Where("asset_id = ?", assetID)

	if documentType != nil {
		query = query.Where("type = ?", *documentType)
	}

	// 计算总数
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	err := query.Preload("Asset").Offset(offset).Limit(limit).Order("upload_date DESC").Find(&documents).Error
	if err != nil {
		return nil, 0, err
	}

	return documents, count, nil
}

// GetByType 根据文档类型列出文档
func (r *AssetDocumentRepository) GetByType(ctx context.Context, documentType string, offset, limit int) ([]*entity.AssetDocument, int64, error) {
	var documents []*entity.AssetDocument
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.AssetDocument{}).Where("type = ?", documentType)

	// 计算总数
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	err := query.Preload("Asset").Offset(offset).Limit(limit).Order("upload_date DESC").Find(&documents).Error
	if err != nil {
		return nil, 0, err
	}

	return documents, count, nil
}
