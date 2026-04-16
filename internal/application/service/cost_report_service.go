package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// CostReportService 成本报表服务
type CostReportService struct {
	costReportRepo repository.CostReportRepository
	costEntryRepo  repository.CostEntryRepository
}

// NewCostReportService 创建成本报表服务实例
func NewCostReportService(
	costReportRepo repository.CostReportRepository,
	costEntryRepo repository.CostEntryRepository,
) *CostReportService {
	return &CostReportService{
		costReportRepo: costReportRepo,
		costEntryRepo:  costEntryRepo,
	}
}

// CreateCostReport 创建成本报表
func (s *CostReportService) CreateCostReport(ctx context.Context, report *entity.CostReport) error {
	// 验证报表时间段是否合法
	if report.PeriodEnd.Before(report.PeriodStart) {
		return errors.New("报表结束时间不能早于开始时间")
	}

	// 检查是否已存在同类型同时间段的报表
	existing, err := s.costReportRepo.GetByPeriod(ctx, report.ReportType, report.PeriodStart, report.PeriodEnd)
	if err == nil && existing != nil {
		return errors.New("该类型该时间段的报表已存在")
	}

	// 生成报表编码
	report.Code = fmt.Sprintf("CR%s%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
	report.Status = "draft"
	report.CreatedAt = time.Now()
	report.UpdatedAt = time.Now()

	return s.costReportRepo.Create(ctx, report)
}

// UpdateCostReport 更新成本报表
func (s *CostReportService) UpdateCostReport(ctx context.Context, report *entity.CostReport) error {
	// 验证报表是否存在
	existing, err := s.costReportRepo.GetByID(ctx, report.ID)
	if err != nil {
		return errors.New("成本报表不存在")
	}

	// 验证报表时间段是否合法
	if report.PeriodEnd.Before(report.PeriodStart) {
		return errors.New("报表结束时间不能早于开始时间")
	}

	// 检查是否已存在同类型同时间段的报表
	if existing.ReportType != report.ReportType || !existing.PeriodStart.Equal(report.PeriodStart) || !existing.PeriodEnd.Equal(report.PeriodEnd) {
		existingByPeriod, err := s.costReportRepo.GetByPeriod(ctx, report.ReportType, report.PeriodStart, report.PeriodEnd)
		if err == nil && existingByPeriod != nil && existingByPeriod.ID != report.ID {
			return errors.New("该类型该时间段的报表已存在")
		}
	}

	// 已审批的报表不能修改
	if existing.Status == "approved" {
		return errors.New("已审批的报表不能修改")
	}

	report.UpdatedAt = time.Now()

	return s.costReportRepo.Update(ctx, report)
}

// DeleteCostReport 删除成本报表
func (s *CostReportService) DeleteCostReport(ctx context.Context, id string) error {
	// 验证报表是否存在
	existing, err := s.costReportRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("成本报表不存在")
	}

	// 已审批的报表不能删除
	if existing.Status == "approved" {
		return errors.New("已审批的报表不能删除")
	}

	return s.costReportRepo.Delete(ctx, id)
}

// GetCostReportByID 根据ID获取成本报表
func (s *CostReportService) GetCostReportByID(ctx context.Context, id string) (*entity.CostReport, error) {
	return s.costReportRepo.GetByID(ctx, id)
}

// GetCostReportByCode 根据编码获取成本报表
func (s *CostReportService) GetCostReportByCode(ctx context.Context, code string) (*entity.CostReport, error) {
	return s.costReportRepo.GetByCode(ctx, code)
}

// ListCostReports 列出成本报表
func (s *CostReportService) ListCostReports(ctx context.Context, reportType *string, status *string, startDate, endDate *time.Time, page, pageSize int) ([]*entity.CostReport, int64, error) {
	offset := (page - 1) * pageSize
	return s.costReportRepo.List(ctx, reportType, status, startDate, endDate, offset, pageSize)
}

// GenerateCostReport 生成成本报表
func (s *CostReportService) GenerateCostReport(ctx context.Context, id, generatedBy string) error {
	// 验证报表是否存在
	report, err := s.costReportRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("成本报表不存在")
	}

	// 验证报表状态
	if report.Status == "approved" {
		return errors.New("已审批的报表不能重新生成")
	}

	// 计算报表期间的成本总额
	totalCost, err := s.costEntryRepo.GetTotalByPeriod(ctx, &report.PeriodStart, &report.PeriodEnd)
	if err != nil {
		return err
	}

	// 更新报表
	report.TotalCost = totalCost
	report.Status = "generated"
	report.GeneratedBy = generatedBy
	report.UpdatedAt = time.Now()

	return s.costReportRepo.Update(ctx, report)
}

// ApproveCostReport 审批成本报表
func (s *CostReportService) ApproveCostReport(ctx context.Context, id, approvedBy string) error {
	// 验证报表是否存在
	report, err := s.costReportRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("成本报表不存在")
	}

	// 验证报表状态
	if report.Status != "generated" {
		return errors.New("报表状态不是已生成")
	}

	// 更新审批状态
	report.Status = "approved"
	report.ApprovedBy = &approvedBy
	now := time.Now()
	report.ApprovedAt = &now
	report.UpdatedAt = now

	return s.costReportRepo.Update(ctx, report)
}

// RejectCostReport 拒绝成本报表
func (s *CostReportService) RejectCostReport(ctx context.Context, id, approvedBy string) error {
	// 验证报表是否存在
	report, err := s.costReportRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("成本报表不存在")
	}

	// 验证报表状态
	if report.Status != "generated" {
		return errors.New("报表状态不是已生成")
	}

	// 更新审批状态
	report.Status = "rejected"
	report.ApprovedBy = &approvedBy
	now := time.Now()
	report.ApprovedAt = &now
	report.UpdatedAt = now

	return s.costReportRepo.Update(ctx, report)
}
