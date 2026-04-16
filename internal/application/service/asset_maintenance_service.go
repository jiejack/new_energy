package service

import (
	"context"
	"errors"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// AssetMaintenanceService 资产维护服务
type AssetMaintenanceService struct {
	assetMaintenanceRepo repository.AssetMaintenanceRepository
	assetRepo           repository.AssetRepository
}

// NewAssetMaintenanceService 创建资产维护服务实例
func NewAssetMaintenanceService(
	assetMaintenanceRepo repository.AssetMaintenanceRepository,
	assetRepo repository.AssetRepository,
) *AssetMaintenanceService {
	return &AssetMaintenanceService{
		assetMaintenanceRepo: assetMaintenanceRepo,
		assetRepo:           assetRepo,
	}
}

// CreateMaintenanceRequest 创建资产维护记录请求
type CreateMaintenanceRequest struct {
	AssetID          string  `json:"asset_id" binding:"required"`
	MaintenanceType  string  `json:"maintenance_type" binding:"required"`
	Description      string  `json:"description"`
	Cost             float64 `json:"cost"`
	MaintenanceDate  string  `json:"maintenance_date" binding:"required"`
	Technician       string  `json:"technician"`
	Status           string  `json:"status"`
}

// UpdateMaintenanceRequest 更新资产维护记录请求
type UpdateMaintenanceRequest struct {
	AssetID          string  `json:"asset_id"`
	MaintenanceType  string  `json:"maintenance_type"`
	Description      string  `json:"description"`
	Cost             float64 `json:"cost"`
	MaintenanceDate  string  `json:"maintenance_date"`
	Technician       string  `json:"technician"`
	Status           string  `json:"status"`
}

// CreateMaintenanceRecord 创建资产维护记录
func (s *AssetMaintenanceService) CreateMaintenanceRecord(ctx context.Context, req *CreateMaintenanceRequest) (*entity.AssetMaintenanceRecord, error) {
	// 验证资产是否存在
	_, err := s.assetRepo.GetByID(ctx, req.AssetID)
	if err != nil {
		return nil, errors.New("资产不存在")
	}

	// 创建维护记录实体
	record := &entity.AssetMaintenanceRecord{
		AssetID:           req.AssetID,
		MaintenanceType:   req.MaintenanceType,
		Description:       req.Description,
		Cost:              req.Cost,
		Technician:        req.Technician,
		Status:            req.Status,
	}

	// 解析维护日期
	if req.MaintenanceDate != "" {
		maintenanceDate, err := time.Parse("2006-01-02", req.MaintenanceDate)
		if err == nil {
			record.MaintenanceDate = &maintenanceDate
		}
	}

	// 设置默认值
	if record.Status == "" {
		record.Status = "pending"
	}

	record.CreatedAt = time.Now()
	record.UpdatedAt = time.Now()

	if err := s.assetMaintenanceRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// GetMaintenanceRecord 根据ID获取资产维护记录
func (s *AssetMaintenanceService) GetMaintenanceRecord(ctx context.Context, id string) (*entity.AssetMaintenanceRecord, error) {
	return s.assetMaintenanceRepo.GetByID(ctx, id)
}

// UpdateMaintenanceRecord 更新资产维护记录
func (s *AssetMaintenanceService) UpdateMaintenanceRecord(ctx context.Context, id string, req *UpdateMaintenanceRequest) (*entity.AssetMaintenanceRecord, error) {
	// 验证维护记录是否存在
	existing, err := s.assetMaintenanceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("维护记录不存在")
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
	if req.MaintenanceType != "" {
		existing.MaintenanceType = req.MaintenanceType
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Cost > 0 {
		existing.Cost = req.Cost
	}
	if req.Technician != "" {
		existing.Technician = req.Technician
	}
	if req.Status != "" {
		existing.Status = req.Status
		// 如果状态改为完成，设置结束时间
		if req.Status == "completed" && existing.EndDate == nil {
			now := time.Now()
			existing.EndDate = &now
		}
	}

	// 解析维护日期
	if req.MaintenanceDate != "" {
		maintenanceDate, err := time.Parse("2006-01-02", req.MaintenanceDate)
		if err == nil {
			existing.MaintenanceDate = &maintenanceDate
		}
	}

	existing.UpdatedAt = time.Now()

	if err := s.assetMaintenanceRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// DeleteMaintenanceRecord 删除资产维护记录
func (s *AssetMaintenanceService) DeleteMaintenanceRecord(ctx context.Context, id string) error {
	// 验证维护记录是否存在
	_, err := s.assetMaintenanceRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("维护记录不存在")
	}

	return s.assetMaintenanceRepo.Delete(ctx, id)
}

// ListMaintenanceRecords 列出资产维护记录
func (s *AssetMaintenanceService) ListMaintenanceRecords(ctx context.Context, assetID, maintenanceType, status string, page, pageSize int) ([]*entity.AssetMaintenanceRecord, int64, error) {
	var statusPtr *string
	var typePtr *string

	if status != "" {
		statusPtr = &status
	}
	if maintenanceType != "" {
		typePtr = &maintenanceType
	}

	offset := (page - 1) * pageSize
	return s.assetMaintenanceRepo.ListByAssetID(ctx, assetID, statusPtr, typePtr, offset, pageSize)
}

// GetMaintenanceCosts 获取资产维护成本
func (s *AssetMaintenanceService) GetMaintenanceCosts(ctx context.Context, assetID, startDate, endDate string) (float64, error) {
	// 验证资产是否存在
	_, err := s.assetRepo.GetByID(ctx, assetID)
	if err != nil {
		return 0, errors.New("资产不存在")
	}

	var startDatePtr *time.Time
	var endDatePtr *time.Time

	// 解析开始日期
	if startDate != "" {
		parsedStartDate, err := time.Parse("2006-01-02", startDate)
		if err == nil {
			startDatePtr = &parsedStartDate
		}
	}

	// 解析结束日期
	if endDate != "" {
		parsedEndDate, err := time.Parse("2006-01-02", endDate)
		if err == nil {
			endDatePtr = &parsedEndDate
		}
	}

	return s.assetMaintenanceRepo.GetMaintenanceCostByAsset(ctx, assetID, startDatePtr, endDatePtr)
}
