package service

import (
	"context"
	"errors"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// AssetDepreciationService 资产折旧服务
type AssetDepreciationService struct {
	assetDepreciationRepo repository.AssetDepreciationRepository
	assetRepo           repository.AssetRepository
}

// NewAssetDepreciationService 创建资产折旧服务实例
func NewAssetDepreciationService(
	assetDepreciationRepo repository.AssetDepreciationRepository,
	assetRepo repository.AssetRepository,
) *AssetDepreciationService {
	return &AssetDepreciationService{
		assetDepreciationRepo: assetDepreciationRepo,
		assetRepo:           assetRepo,
	}
}

// CreateDepreciationRequest 创建资产折旧记录请求
type CreateDepreciationRequest struct {
	AssetID            string  `json:"asset_id" binding:"required"`
	DepreciationMethod string  `json:"depreciation_method" binding:"required"`
	Year               int     `json:"year" binding:"required"`
	Amount             float64 `json:"amount" binding:"required"`
	AccumulatedAmount  float64 `json:"accumulated_amount" binding:"required"`
	BookValue          float64 `json:"book_value" binding:"required"`
}

// UpdateDepreciationRequest 更新资产折旧记录请求
type UpdateDepreciationRequest struct {
	AssetID            string  `json:"asset_id"`
	DepreciationMethod string  `json:"depreciation_method"`
	Year               int     `json:"year"`
	Amount             float64 `json:"amount"`
	AccumulatedAmount  float64 `json:"accumulated_amount"`
	BookValue          float64 `json:"book_value"`
}

// CreateDepreciationRecord 创建资产折旧记录
func (s *AssetDepreciationService) CreateDepreciationRecord(ctx context.Context, req *CreateDepreciationRequest) (*entity.AssetDepreciationRecord, error) {
	// 验证资产是否存在
	_, err := s.assetRepo.GetByID(ctx, req.AssetID)
	if err != nil {
		return nil, errors.New("资产不存在")
	}

	// 创建折旧记录实体
	record := &entity.AssetDepreciationRecord{
		AssetID:                 req.AssetID,
		DepreciationMethod:      req.DepreciationMethod,
		Year:                    req.Year,
		DepreciationAmount:      req.Amount,
		AccumulatedDepreciation: req.AccumulatedAmount,
		BookValue:               req.BookValue,
		Period:                  "annual", // 默认年度折旧
	}

	// 验证折旧期间是否有效
	validPeriods := map[string]bool{
		"monthly":   true,
		"quarterly": true,
		"annual":    true,
	}
	if !validPeriods[record.Period] {
		return nil, errors.New("无效的折旧期间")
	}

	record.CreatedAt = time.Now()
	record.UpdatedAt = time.Now()

	if err := s.assetDepreciationRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// GetDepreciationRecord 根据ID获取资产折旧记录
func (s *AssetDepreciationService) GetDepreciationRecord(ctx context.Context, id string) (*entity.AssetDepreciationRecord, error) {
	return s.assetDepreciationRepo.GetByID(ctx, id)
}

// UpdateDepreciationRecord 更新资产折旧记录
func (s *AssetDepreciationService) UpdateDepreciationRecord(ctx context.Context, id string, req *UpdateDepreciationRequest) (*entity.AssetDepreciationRecord, error) {
	// 验证折旧记录是否存在
	existing, err := s.assetDepreciationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("折旧记录不存在")
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
	if req.DepreciationMethod != "" {
		existing.DepreciationMethod = req.DepreciationMethod
	}
	if req.Year > 0 {
		existing.Year = req.Year
	}
	if req.Amount > 0 {
		existing.DepreciationAmount = req.Amount
	}
	if req.AccumulatedAmount > 0 {
		existing.AccumulatedDepreciation = req.AccumulatedAmount
	}
	if req.BookValue > 0 {
		existing.BookValue = req.BookValue
	}

	// 验证折旧期间是否有效
	validPeriods := map[string]bool{
		"monthly":   true,
		"quarterly": true,
		"annual":    true,
	}
	if !validPeriods[existing.Period] {
		return nil, errors.New("无效的折旧期间")
	}

	existing.UpdatedAt = time.Now()

	if err := s.assetDepreciationRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// DeleteDepreciationRecord 删除资产折旧记录
func (s *AssetDepreciationService) DeleteDepreciationRecord(ctx context.Context, id string) error {
	// 验证折旧记录是否存在
	_, err := s.assetDepreciationRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("折旧记录不存在")
	}

	return s.assetDepreciationRepo.Delete(ctx, id)
}

// ListDepreciationRecords 列出资产折旧记录
func (s *AssetDepreciationService) ListDepreciationRecords(ctx context.Context, assetID, method string, year int, page, pageSize int) ([]*entity.AssetDepreciationRecord, int64, error) {
	var periodPtr *string

	offset := (page - 1) * pageSize
	return s.assetDepreciationRepo.ListByAssetID(ctx, assetID, periodPtr, offset, pageSize)
}

// GetDepreciationSummary 获取资产折旧汇总
func (s *AssetDepreciationService) GetDepreciationSummary(ctx context.Context, assetID, upToDate string) (float64, error) {
	// 验证资产是否存在
	_, err := s.assetRepo.GetByID(ctx, assetID)
	if err != nil {
		return 0, errors.New("资产不存在")
	}

	var startDatePtr *time.Time
	var endDatePtr *time.Time

	// 解析截止日期
	if upToDate != "" {
		parsedEndDate, err := time.Parse("2006-01-02", upToDate)
		if err == nil {
			endDatePtr = &parsedEndDate
		}
	}

	return s.assetDepreciationRepo.GetDepreciationSummaryByPeriod(ctx, "annual", startDatePtr, endDatePtr)
}
