package persistence

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

var _ repository.CostReportRepository = (*CostReportRepository)(nil)

// CostReportRepository 成本报表仓储实现
type CostReportRepository struct {
	db *Database
}

// NewCostReportRepository 创建成本报表仓储实例
func NewCostReportRepository(db *Database) repository.CostReportRepository {
	return &CostReportRepository{db: db}
}

// Create 创建成本报表
func (r *CostReportRepository) Create(ctx context.Context, report *entity.CostReport) error {
	return r.db.WithContext(ctx).Create(report).Error
}

// Update 更新成本报表
func (r *CostReportRepository) Update(ctx context.Context, report *entity.CostReport) error {
	return r.db.WithContext(ctx).Save(report).Error
}

// Delete 删除成本报表
func (r *CostReportRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.CostReport{}, "id = ?", id).Error
}

// GetByID 根据ID获取成本报表
func (r *CostReportRepository) GetByID(ctx context.Context, id string) (*entity.CostReport, error) {
	var report entity.CostReport
	err := r.db.WithContext(ctx).First(&report, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// GetByCode 根据编码获取成本报表
func (r *CostReportRepository) GetByCode(ctx context.Context, code string) (*entity.CostReport, error) {
	var report entity.CostReport
	err := r.db.WithContext(ctx).First(&report, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// List 列出成本报表
func (r *CostReportRepository) List(ctx context.Context, reportType *string, status *string, startDate, endDate *time.Time, offset, limit int) ([]*entity.CostReport, int64, error) {
	var reports []*entity.CostReport
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.CostReport{})

	if reportType != nil {
		query = query.Where("report_type = ?", *reportType)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if startDate != nil {
		query = query.Where("period_start >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("period_end <= ?", *endDate)
	}

	// 计算总数
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	err := query.Offset(offset).Limit(limit).Order("period_start DESC").Find(&reports).Error
	if err != nil {
		return nil, 0, err
	}

	return reports, count, nil
}

// GetByPeriod 根据时间段获取成本报表
func (r *CostReportRepository) GetByPeriod(ctx context.Context, reportType string, periodStart, periodEnd time.Time) (*entity.CostReport, error) {
	var report entity.CostReport
	err := r.db.WithContext(ctx).Where("report_type = ? AND period_start = ? AND period_end = ?", reportType, periodStart, periodEnd).First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}
