package service

import (
	"context"
	"errors"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// AssetDocumentService 资产文档服务
type AssetDocumentService struct {
	assetDocumentRepo repository.AssetDocumentRepository
	assetRepo         repository.AssetRepository
}

// NewAssetDocumentService 创建资产文档服务实例
func NewAssetDocumentService(
	assetDocumentRepo repository.AssetDocumentRepository,
	assetRepo repository.AssetRepository,
) *AssetDocumentService {
	return &AssetDocumentService{
		assetDocumentRepo: assetDocumentRepo,
		assetRepo:         assetRepo,
	}
}

// CreateDocumentRequest 创建资产文档请求
type CreateDocumentRequest struct {
	AssetID      string `json:"asset_id" binding:"required"`
	DocumentType string `json:"document_type" binding:"required"`
	Title        string `json:"title" binding:"required"`
	FilePath     string `json:"file_path" binding:"required"`
	Description  string `json:"description"`
	UploadDate   string `json:"upload_date"`
}

// UpdateDocumentRequest 更新资产文档请求
type UpdateDocumentRequest struct {
	AssetID      string `json:"asset_id"`
	DocumentType string `json:"document_type"`
	Title        string `json:"title"`
	FilePath     string `json:"file_path"`
	Description  string `json:"description"`
	UploadDate   string `json:"upload_date"`
}

// CreateDocument 创建资产文档
func (s *AssetDocumentService) CreateDocument(ctx context.Context, req *CreateDocumentRequest) (*entity.AssetDocument, error) {
	// 验证资产是否存在
	_, err := s.assetRepo.GetByID(ctx, req.AssetID)
	if err != nil {
		return nil, errors.New("资产不存在")
	}

	// 创建文档实体
	document := &entity.AssetDocument{
		AssetID:      req.AssetID,
		DocumentType: req.DocumentType,
		Title:        req.Title,
		FilePath:     req.FilePath,
		Description:  req.Description,
	}

	// 解析上传日期
	if req.UploadDate != "" {
		uploadDate, err := time.Parse("2006-01-02", req.UploadDate)
		if err == nil {
			document.UploadDate = uploadDate
		}
	}

	// 设置默认值
	if document.UploadDate.IsZero() {
		document.UploadDate = time.Now()
	}

	document.CreatedAt = time.Now()
	document.UpdatedAt = time.Now()

	if err := s.assetDocumentRepo.Create(ctx, document); err != nil {
		return nil, err
	}

	return document, nil
}

// GetDocument 根据ID获取资产文档
func (s *AssetDocumentService) GetDocument(ctx context.Context, id string) (*entity.AssetDocument, error) {
	return s.assetDocumentRepo.GetByID(ctx, id)
}

// UpdateDocument 更新资产文档
func (s *AssetDocumentService) UpdateDocument(ctx context.Context, id string, req *UpdateDocumentRequest) (*entity.AssetDocument, error) {
	// 验证文档是否存在
	existing, err := s.assetDocumentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("文档不存在")
	}

	// 验证资产是否存在
	if req.AssetID != "" {
		_, err = s.assetRepo.GetByID(ctx, req.AssetID)
		if err != nil {
			return nil, errors.New("资产不存在")
		}
		existing.AssetID = req.AssetID
	}

	// 更新其他字段
	if req.DocumentType != "" {
		existing.DocumentType = req.DocumentType
	}
	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.FilePath != "" {
		existing.FilePath = req.FilePath
	}
	if req.Description != "" {
		existing.Description = req.Description
	}

	// 解析上传日期
	if req.UploadDate != "" {
		uploadDate, err := time.Parse("2006-01-02", req.UploadDate)
		if err == nil {
			existing.UploadDate = uploadDate
		}
	}

	existing.UpdatedAt = time.Now()

	if err := s.assetDocumentRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// DeleteDocument 删除资产文档
func (s *AssetDocumentService) DeleteDocument(ctx context.Context, id string) error {
	// 验证文档是否存在
	_, err := s.assetDocumentRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("文档不存在")
	}

	return s.assetDocumentRepo.Delete(ctx, id)
}

// ListDocuments 列出资产文档
func (s *AssetDocumentService) ListDocuments(ctx context.Context, assetID, documentType string, page, pageSize int) ([]*entity.AssetDocument, int64, error) {
	var typePtr *string

	if documentType != "" {
		typePtr = &documentType
	}

	offset := (page - 1) * pageSize
	return s.assetDocumentRepo.ListByAssetID(ctx, assetID, typePtr, offset, pageSize)
}
