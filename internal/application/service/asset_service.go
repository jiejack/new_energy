package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// AssetService 资产服务
type AssetService struct {
	assetRepo           repository.AssetRepository
	assetMaintenanceRepo repository.AssetMaintenanceRepository
	assetDepreciationRepo repository.AssetDepreciationRepository
	assetDocumentRepo   repository.AssetDocumentRepository
}

// NewAssetService 创建资产服务实例
func NewAssetService(
	assetRepo repository.AssetRepository,
	assetMaintenanceRepo repository.AssetMaintenanceRepository,
	assetDepreciationRepo repository.AssetDepreciationRepository,
	assetDocumentRepo repository.AssetDocumentRepository,
) *AssetService {
	return &AssetService{
		assetRepo:           assetRepo,
		assetMaintenanceRepo: assetMaintenanceRepo,
		assetDepreciationRepo: assetDepreciationRepo,
		assetDocumentRepo:   assetDocumentRepo,
	}
}

// CreateAssetRequest 创建资产请求
type CreateAssetRequest struct {
	Code          string  `json:"code" binding:"required"`
	Name          string  `json:"name" binding:"required"`
	Category      string  `json:"category" binding:"required"`
	AssetType     string  `json:"asset_type"`
	Manufacturer  string  `json:"manufacturer"`
	Model         string  `json:"model"`
	SerialNumber  string  `json:"serial_number"`
	PurchasePrice float64 `json:"purchase_price" binding:"required"`
	PurchaseDate  string  `json:"purchase_date" binding:"required"`
	ExpectedLife  int     `json:"expected_life" binding:"required"`
	ResidualValue float64 `json:"residual_value"`
	Location      string  `json:"location"`
	Status        string  `json:"status"`
	Description   string  `json:"description"`
}

// UpdateAssetRequest 更新资产请求
type UpdateAssetRequest struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Category      string  `json:"category"`
	AssetType     string  `json:"asset_type"`
	Manufacturer  string  `json:"manufacturer"`
	Model         string  `json:"model"`
	SerialNumber  string  `json:"serial_number"`
	PurchasePrice float64 `json:"purchase_price"`
	PurchaseDate  string  `json:"purchase_date"`
	ExpectedLife  int     `json:"expected_life"`
	ResidualValue float64 `json:"residual_value"`
	Location      string  `json:"location"`
	Status        string  `json:"status"`
	Description   string  `json:"description"`
}

// CreateAsset 创建资产
func (s *AssetService) CreateAsset(ctx context.Context, req *CreateAssetRequest) (*entity.Asset, error) {
	// 验证资产编码是否已存在
	existing, err := s.assetRepo.GetByCode(ctx, req.Code)
	if err == nil && existing != nil {
		return nil, errors.New("资产编码已存在")
	}

	// 生成资产编码
	code := req.Code
	if code == "" {
		code = fmt.Sprintf("ASSET%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
	}

	// 创建资产实体
	asset := &entity.Asset{
		Code:          code,
		Name:          req.Name,
		Category:      req.Category,
		AssetType:     req.AssetType,
		Manufacturer:  req.Manufacturer,
		Model:         req.Model,
		SerialNumber:  req.SerialNumber,
		PurchasePrice: req.PurchasePrice,
		ExpectedLife:  req.ExpectedLife,
		ResidualValue: req.ResidualValue,
		Location:      req.Location,
		Status:        req.Status,
		Description:   req.Description,
	}

	// 解析采购日期
	if req.PurchaseDate != "" {
		purchaseDate, err := time.Parse("2006-01-02", req.PurchaseDate)
		if err == nil {
			asset.PurchaseDate = &purchaseDate
		}
	}

	// 设置默认值
	if asset.Status == "" {
		asset.Status = "active"
	}
	if asset.UsageStatus == "" {
		asset.UsageStatus = "in_use"
	}
	if asset.CurrentValue == 0 {
		asset.CurrentValue = asset.PurchasePrice
	}

	asset.CreatedAt = time.Now()
	asset.UpdatedAt = time.Now()

	if err := s.assetRepo.Create(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

// GetAsset 根据ID获取资产
func (s *AssetService) GetAsset(ctx context.Context, id string) (*entity.Asset, error) {
	return s.assetRepo.GetByID(ctx, id)
}

// UpdateAsset 更新资产
func (s *AssetService) UpdateAsset(ctx context.Context, id string, req *UpdateAssetRequest) (*entity.Asset, error) {
	// 验证资产是否存在
	existing, err := s.assetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("资产不存在")
	}

	// 验证编码是否已被其他资产使用
	if req.Code != "" && req.Code != existing.Code {
		existingByCode, err := s.assetRepo.GetByCode(ctx, req.Code)
		if err == nil && existingByCode != nil && existingByCode.ID != id {
			return nil, errors.New("资产编码已被使用")
		}
		existing.Code = req.Code
	}

	// 更新其他字段
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Category != "" {
		existing.Category = req.Category
	}
	if req.AssetType != "" {
		existing.AssetType = req.AssetType
	}
	if req.Manufacturer != "" {
		existing.Manufacturer = req.Manufacturer
	}
	if req.Model != "" {
		existing.Model = req.Model
	}
	if req.SerialNumber != "" {
		existing.SerialNumber = req.SerialNumber
	}
	if req.PurchasePrice > 0 {
		existing.PurchasePrice = req.PurchasePrice
	}
	if req.ExpectedLife > 0 {
		existing.ExpectedLife = req.ExpectedLife
	}
	if req.ResidualValue > 0 {
		existing.ResidualValue = req.ResidualValue
	}
	if req.Location != "" {
		existing.Location = req.Location
	}
	if req.Status != "" {
		existing.Status = req.Status
	}
	if req.Description != "" {
		existing.Description = req.Description
	}

	// 解析采购日期
	if req.PurchaseDate != "" {
		purchaseDate, err := time.Parse("2006-01-02", req.PurchaseDate)
		if err == nil {
			existing.PurchaseDate = &purchaseDate
		}
	}

	existing.UpdatedAt = time.Now()

	if err := s.assetRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// DeleteAsset 删除资产
func (s *AssetService) DeleteAsset(ctx context.Context, id string) error {
	// 验证资产是否存在
	_, err := s.assetRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("资产不存在")
	}

	// 检查是否有维护记录
	maintenanceRecords, _, err := s.assetMaintenanceRepo.ListByAssetID(ctx, id, nil, nil, 0, 1)
	if err != nil {
		return err
	}
	if len(maintenanceRecords) > 0 {
		return errors.New("该资产有维护记录，无法删除")
	}

	// 检查是否有折旧记录
	depreciationRecords, _, err := s.assetDepreciationRepo.ListByAssetID(ctx, id, nil, 0, 1)
	if err != nil {
		return err
	}
	if len(depreciationRecords) > 0 {
		return errors.New("该资产有折旧记录，无法删除")
	}

	// 检查是否有文档
	documents, _, err := s.assetDocumentRepo.ListByAssetID(ctx, id, nil, 0, 1)
	if err != nil {
		return err
	}
	if len(documents) > 0 {
		return errors.New("该资产有文档，无法删除")
	}

	return s.assetRepo.Delete(ctx, id)
}

// ListAssets 列出资产
func (s *AssetService) ListAssets(ctx context.Context, name, category, status string, page, pageSize int) ([]*entity.Asset, int64, error) {
	var assetType *string
	var statusPtr *string
	var categoryPtr *string

	if status != "" {
		statusPtr = &status
	}
	if category != "" {
		categoryPtr = &category
	}

	offset := (page - 1) * pageSize
	return s.assetRepo.List(ctx, assetType, statusPtr, categoryPtr, offset, pageSize)
}

// CalculateDepreciation 计算资产折旧
func (s *AssetService) CalculateDepreciation(ctx context.Context, id, method string) (map[string]interface{}, error) {
	// 验证资产是否存在
	asset, err := s.assetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("资产不存在")
	}

	// 验证资产是否可折旧
	if asset.ExpectedLife <= 0 {
		return nil, errors.New("资产不可折旧")
	}

	// 计算折旧金额
	var annualDepreciation float64
	var monthlyDepreciation float64
	var accumulatedDepreciation float64
	var bookValue float64

	// 获取最新折旧记录
	latestRecord, err := s.assetDepreciationRepo.GetLatestByAssetID(ctx, id)

	if err != nil {
		// 首次折旧
		bookValue = asset.PurchasePrice
	} else {
		bookValue = latestRecord.BookValue
		accumulatedDepreciation = latestRecord.AccumulatedDepreciation
	}

	// 根据折旧方法计算折旧金额
	switch method {
	case "straight-line":
		// 直线法
		annualDepreciation = (asset.PurchasePrice - asset.ResidualValue) / float64(asset.ExpectedLife)
		monthlyDepreciation = annualDepreciation / 12
	case "declining-balance":
		// 双倍余额递减法
		if bookValue <= asset.ResidualValue {
			return nil, errors.New("资产已折旧完毕")
		}
		depreciationRate := 2.0 / float64(asset.ExpectedLife)
		annualDepreciation = bookValue * depreciationRate
		monthlyDepreciation = annualDepreciation / 12
		// 确保账面价值不低于残值
		if bookValue-annualDepreciation < asset.ResidualValue {
			annualDepreciation = bookValue - asset.ResidualValue
			monthlyDepreciation = annualDepreciation / 12
		}
	default:
		return nil, errors.New("不支持的折旧方法")
	}

	// 计算累计折旧和账面价值
	accumulatedDepreciation += annualDepreciation
	bookValue -= annualDepreciation

	return map[string]interface{}{
		"asset_id":                  id,
		"method":                   method,
		"annual_depreciation":      annualDepreciation,
		"monthly_depreciation":     monthlyDepreciation,
		"accumulated_depreciation": accumulatedDepreciation,
		"book_value":               bookValue,
	}, nil
}
