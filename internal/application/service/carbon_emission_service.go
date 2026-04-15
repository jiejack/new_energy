package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type CarbonEmissionService struct {
	ceRepo repository.CarbonEmissionRepository
}

func NewCarbonEmissionService(ceRepo repository.CarbonEmissionRepository) *CarbonEmissionService {
	return &CarbonEmissionService{ceRepo: ceRepo}
}

type CreateCarbonEmissionFactorRequest struct {
	Name        string               `json:"name" binding:"required"`
	Code        string               `json:"code" binding:"required"`
	Scope       entity.CarbonEmissionScope `json:"scope" binding:"required"`
	Source      string               `json:"source" binding:"required"`
	Value       float64              `json:"value" binding:"required,min=0"`
	Unit        string               `json:"unit" binding:"required"`
	Description string               `json:"description"`
	Version     string               `json:"version" binding:"required"`
	EffectiveAt time.Time            `json:"effective_at" binding:"required"`
	ExpiresAt   *time.Time           `json:"expires_at"`
}

type UpdateCarbonEmissionFactorRequest struct {
	Name        *string              `json:"name"`
	Scope       *entity.CarbonEmissionScope `json:"scope"`
	Source      *string              `json:"source"`
	Value       *float64             `json:"value"`
	Unit        *string              `json:"unit"`
	Description *string              `json:"description"`
	Version     *string              `json:"version"`
	EffectiveAt *time.Time           `json:"effective_at"`
	ExpiresAt   *time.Time           `json:"expires_at"`
	IsActive    *bool                `json:"is_active"`
}

type CreateCarbonEmissionRecordRequest struct {
	RecordTime   time.Time              `json:"record_time" binding:"required"`
	Scope        entity.CarbonEmissionScope `json:"scope" binding:"required"`
	TargetID     string                 `json:"target_id" binding:"required"`
	TargetName   string                 `json:"target_name" binding:"required"`
	FactorID     string                 `json:"factor_id"`
	FactorCode   string                 `json:"factor_code" binding:"required"`
	FactorValue  float64                `json:"factor_value" binding:"required,min=0"`
	ActivityData float64                `json:"activity_data" binding:"required,min=0"`
	ActivityUnit string                 `json:"activity_unit" binding:"required"`
	Period       string                 `json:"period" binding:"required"`
	Remark       string                 `json:"remark"`
}

type QueryCarbonEmissionFactorsRequest struct {
	Page     int                       `form:"page,default=1"`
	PageSize int                       `form:"page_size,default=10"`
	Scope    *entity.CarbonEmissionScope `form:"scope"`
	Source   *string                   `form:"source"`
	IsActive *bool                     `form:"is_active"`
}

type QueryCarbonEmissionRecordsRequest struct {
	Page      int                       `form:"page,default=1"`
	PageSize  int                       `form:"page_size,default=10"`
	Scope     *entity.CarbonEmissionScope `form:"scope"`
	TargetID  *string                   `form:"target_id"`
	Period    *string                   `form:"period"`
	Status    *entity.CarbonEmissionStatus `form:"status"`
	StartTime *time.Time                `form:"start_time"`
	EndTime   *time.Time                `form:"end_time"`
}

type QueryCarbonEmissionSummariesRequest struct {
	Page      int                       `form:"page,default=1"`
	PageSize  int                       `form:"page_size,default=10"`
	TargetID  *string                   `form:"target_id"`
	Period    *string                   `form:"period"`
	Status    *entity.CarbonEmissionStatus `form:"status"`
	StartTime *time.Time                `form:"start_time"`
	EndTime   *time.Time                `form:"end_time"`
}

type QueryCarbonReductionTargetsRequest struct {
	Page      int     `form:"page,default=1"`
	PageSize  int     `form:"page_size,default=10"`
	TargetID  *string `form:"target_id"`
	Status    *string `form:"status"`
	StartYear *int    `form:"start_year"`
	EndYear   *int    `form:"end_year"`
}

type CreateCarbonReductionTargetRequest struct {
	Name             string    `json:"name" binding:"required"`
	Description      string    `json:"description"`
	TargetID         string    `json:"target_id" binding:"required"`
	TargetName       string    `json:"target_name" binding:"required"`
	BaseYear         int       `json:"base_year" binding:"required"`
	BaseEmission     float64   `json:"base_emission" binding:"required,min=0"`
	TargetYear       int       `json:"target_year" binding:"required"`
	TargetReduction  float64   `json:"target_reduction" binding:"required,min=0"`
	StartDate        time.Time `json:"start_date" binding:"required"`
	EndDate          time.Time `json:"end_date" binding:"required"`
}

type UpdateCarbonReductionTargetRequest struct {
	Name             *string  `json:"name"`
	Description      *string  `json:"description"`
	TargetEmission   *float64 `json:"target_emission"`
	CurrentProgress  *float64 `json:"current_progress"`
	CurrentEmission  *float64 `json:"current_emission"`
	StartDate        *time.Time `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	Status           *string  `json:"status"`
}

type CarbonEmissionTrendData struct {
	Time           time.Time `json:"time"`
	Scope1Emission float64   `json:"scope1_emission"`
	Scope2Emission float64   `json:"scope2_emission"`
	Scope3Emission float64   `json:"scope3_emission"`
	TotalEmission  float64   `json:"total_emission"`
}

func (s *CarbonEmissionService) CreateFactor(ctx context.Context, req *CreateCarbonEmissionFactorRequest, createdBy string) (*entity.CarbonEmissionFactor, error) {
	existing, _ := s.ceRepo.GetFactorByCode(ctx, req.Code)
	if existing != nil {
		return nil, fmt.Errorf("carbon emission factor with code %s already exists", req.Code)
	}

	factor := entity.NewCarbonEmissionFactor(
		req.Name,
		req.Code,
		req.Scope,
		req.Source,
		req.Value,
		req.Unit,
		req.Version,
		req.EffectiveAt,
	)
	factor.ID = uuid.New().String()
	factor.Description = req.Description
	factor.ExpiresAt = req.ExpiresAt

	if err := s.ceRepo.CreateFactor(ctx, factor); err != nil {
		return nil, fmt.Errorf("failed to create carbon emission factor: %w", err)
	}

	return factor, nil
}

func (s *CarbonEmissionService) UpdateFactor(ctx context.Context, id string, req *UpdateCarbonEmissionFactorRequest, updatedBy string) (*entity.CarbonEmissionFactor, error) {
	factor, err := s.ceRepo.GetFactorByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("carbon emission factor not found: %w", err)
	}

	if req.Name != nil {
		factor.Name = *req.Name
	}
	if req.Scope != nil {
		factor.Scope = *req.Scope
	}
	if req.Source != nil {
		factor.Source = *req.Source
	}
	if req.Value != nil {
		factor.Value = *req.Value
	}
	if req.Unit != nil {
		factor.Unit = *req.Unit
	}
	if req.Description != nil {
		factor.Description = *req.Description
	}
	if req.Version != nil {
		factor.Version = *req.Version
	}
	if req.EffectiveAt != nil {
		factor.EffectiveAt = *req.EffectiveAt
	}
	if req.ExpiresAt != nil {
		factor.ExpiresAt = req.ExpiresAt
	}
	if req.IsActive != nil {
		factor.IsActive = *req.IsActive
	}

	factor.UpdatedAt = time.Now()

	if err := s.ceRepo.UpdateFactor(ctx, factor); err != nil {
		return nil, fmt.Errorf("failed to update carbon emission factor: %w", err)
	}

	return factor, nil
}

func (s *CarbonEmissionService) GetFactor(ctx context.Context, id string) (*entity.CarbonEmissionFactor, error) {
	return s.ceRepo.GetFactorByID(ctx, id)
}

func (s *CarbonEmissionService) ListFactors(ctx context.Context, req *QueryCarbonEmissionFactorsRequest) ([]*entity.CarbonEmissionFactor, int64, error) {
	query := &repository.CarbonEmissionFactorQuery{
		Page:     req.Page,
		PageSize: req.PageSize,
		Scope:    req.Scope,
		Source:   req.Source,
		IsActive: req.IsActive,
	}
	return s.ceRepo.ListFactors(ctx, query)
}

func (s *CarbonEmissionService) GetActiveFactors(ctx context.Context, scope *entity.CarbonEmissionScope) ([]*entity.CarbonEmissionFactor, error) {
	return s.ceRepo.GetActiveFactors(ctx, scope)
}

func (s *CarbonEmissionService) CreateRecord(ctx context.Context, req *CreateCarbonEmissionRecordRequest, createdBy string) (*entity.CarbonEmissionRecord, error) {
	record := entity.NewCarbonEmissionRecord(
		req.RecordTime,
		req.Scope,
		req.TargetID,
		req.TargetName,
		req.FactorID,
		req.FactorCode,
		req.FactorValue,
		req.ActivityData,
		req.ActivityUnit,
		req.Period,
	)
	record.ID = uuid.New().String()
	record.Remark = req.Remark
	record.CreatedBy = createdBy
	record.UpdatedBy = createdBy

	if err := s.ceRepo.CreateRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create carbon emission record: %w", err)
	}

	return record, nil
}

func (s *CarbonEmissionService) BatchCreateRecords(ctx context.Context, reqs []*CreateCarbonEmissionRecordRequest, createdBy string) ([]*entity.CarbonEmissionRecord, error) {
	records := make([]*entity.CarbonEmissionRecord, 0, len(reqs))
	
	for _, req := range reqs {
		record := entity.NewCarbonEmissionRecord(
			req.RecordTime,
			req.Scope,
			req.TargetID,
			req.TargetName,
			req.FactorID,
			req.FactorCode,
			req.FactorValue,
			req.ActivityData,
			req.ActivityUnit,
			req.Period,
		)
		record.ID = uuid.New().String()
		record.Remark = req.Remark
		record.CreatedBy = createdBy
		record.UpdatedBy = createdBy
		records = append(records, record)
	}

	if err := s.ceRepo.BatchCreateRecords(ctx, records); err != nil {
		return nil, fmt.Errorf("failed to batch create carbon emission records: %w", err)
	}

	return records, nil
}

func (s *CarbonEmissionService) GetRecord(ctx context.Context, id string) (*entity.CarbonEmissionRecord, error) {
	return s.ceRepo.GetRecordByID(ctx, id)
}

func (s *CarbonEmissionService) ListRecords(ctx context.Context, req *QueryCarbonEmissionRecordsRequest) ([]*entity.CarbonEmissionRecord, int64, error) {
	query := &repository.CarbonEmissionRecordQuery{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Scope:     req.Scope,
		TargetID:  req.TargetID,
		Period:    req.Period,
		Status:    req.Status,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
	return s.ceRepo.ListRecords(ctx, query)
}

func (s *CarbonEmissionService) GetTrendData(ctx context.Context, targetID string, period string, startTime, endTime time.Time) ([]CarbonEmissionTrendData, error) {
	records, err := s.ceRepo.GetRecordsByTimeRange(ctx, targetID, nil, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get records by time range: %w", err)
	}

	grouped := make(map[time.Time]*CarbonEmissionTrendData)
	
	for _, record := range records {
		date := record.RecordTime.Truncate(24 * time.Hour)
		if _, ok := grouped[date]; !ok {
			grouped[date] = &CarbonEmissionTrendData{
				Time: date,
			}
		}
		
		data := grouped[date]
		switch record.Scope {
		case entity.CarbonEmissionScope1:
			data.Scope1Emission += record.EmissionValue
		case entity.CarbonEmissionScope2:
			data.Scope2Emission += record.EmissionValue
		case entity.CarbonEmissionScope3:
			data.Scope3Emission += record.EmissionValue
		}
		data.TotalEmission += record.EmissionValue
	}

	result := make([]CarbonEmissionTrendData, 0, len(grouped))
	for _, data := range grouped {
		result = append(result, *data)
	}

	return result, nil
}

func (s *CarbonEmissionService) CreateSummary(ctx context.Context, targetID, targetName, period string, periodStart, periodEnd time.Time, createdBy string) (*entity.CarbonEmissionSummary, error) {
	scope1, _ := s.ceRepo.GetTotalEmissionByScope(ctx, targetID, entity.CarbonEmissionScope1, period, periodStart, periodEnd)
	scope2, _ := s.ceRepo.GetTotalEmissionByScope(ctx, targetID, entity.CarbonEmissionScope2, period, periodStart, periodEnd)
	scope3, _ := s.ceRepo.GetTotalEmissionByScope(ctx, targetID, entity.CarbonEmissionScope3, period, periodStart, periodEnd)
	total := scope1 + scope2 + scope3

	summary := &entity.CarbonEmissionSummary{
		ID:              uuid.New().String(),
		SummaryTime:     time.Now(),
		TargetID:        targetID,
		TargetName:      targetName,
		Period:          period,
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		Scope1Emission:  scope1,
		Scope2Emission:  scope2,
		Scope3Emission:  scope3,
		TotalEmission:   total,
		Status:          entity.CarbonEmissionStatusDraft,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	trend := "stable"
	lastSummary, _ := s.ceRepo.GetLatestSummary(ctx, targetID, period)
	if lastSummary != nil {
		if total > lastSummary.TotalEmission*1.05 {
			trend = "increasing"
		} else if total < lastSummary.TotalEmission*0.95 {
			trend = "decreasing"
		}
		summary.Trend = trend
		
		if lastSummary.TotalEmission > 0 {
			summary.MoMChange = (total - lastSummary.TotalEmission) / lastSummary.TotalEmission * 100
		}
	}
	summary.Trend = trend

	if err := s.ceRepo.CreateSummary(ctx, summary); err != nil {
		return nil, fmt.Errorf("failed to create carbon emission summary: %w", err)
	}

	return summary, nil
}

func (s *CarbonEmissionService) GetSummary(ctx context.Context, id string) (*entity.CarbonEmissionSummary, error) {
	return s.ceRepo.GetSummaryByID(ctx, id)
}

func (s *CarbonEmissionService) ListSummaries(ctx context.Context, req *QueryCarbonEmissionSummariesRequest) ([]*entity.CarbonEmissionSummary, int64, error) {
	query := &repository.CarbonEmissionSummaryQuery{
		Page:      req.Page,
		PageSize:  req.PageSize,
		TargetID:  req.TargetID,
		Period:    req.Period,
		Status:    req.Status,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
	return s.ceRepo.ListSummaries(ctx, query)
}

func (s *CarbonEmissionService) CreateTarget(ctx context.Context, req *CreateCarbonReductionTargetRequest, createdBy string) (*entity.CarbonReductionTarget, error) {
	target := &entity.CarbonReductionTarget{
		ID:              uuid.New().String(),
		Name:            req.Name,
		Description:     req.Description,
		TargetID:        req.TargetID,
		TargetName:      req.TargetName,
		BaseYear:        req.BaseYear,
		BaseEmission:    req.BaseEmission,
		TargetYear:      req.TargetYear,
		TargetReduction: req.TargetReduction,
		TargetEmission:  req.BaseEmission * (1 - req.TargetReduction/100),
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		Status:          "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		CreatedBy:       createdBy,
		UpdatedBy:       createdBy,
	}

	if err := s.ceRepo.CreateTarget(ctx, target); err != nil {
		return nil, fmt.Errorf("failed to create carbon reduction target: %w", err)
	}

	return target, nil
}

func (s *CarbonEmissionService) UpdateTarget(ctx context.Context, id string, req *UpdateCarbonReductionTargetRequest, updatedBy string) (*entity.CarbonReductionTarget, error) {
	target, err := s.ceRepo.GetTargetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("carbon reduction target not found: %w", err)
	}

	if req.Name != nil {
		target.Name = *req.Name
	}
	if req.Description != nil {
		target.Description = *req.Description
	}
	if req.TargetEmission != nil {
		target.TargetEmission = *req.TargetEmission
	}
	if req.CurrentProgress != nil {
		target.CurrentProgress = *req.CurrentProgress
	}
	if req.CurrentEmission != nil {
		target.CurrentEmission = *req.CurrentEmission
	}
	if req.StartDate != nil {
		target.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		target.EndDate = *req.EndDate
	}
	if req.Status != nil {
		target.Status = *req.Status
	}

	target.UpdatedAt = time.Now()
	target.UpdatedBy = updatedBy

	if err := s.ceRepo.UpdateTarget(ctx, target); err != nil {
		return nil, fmt.Errorf("failed to update carbon reduction target: %w", err)
	}

	return target, nil
}

func (s *CarbonEmissionService) GetTarget(ctx context.Context, id string) (*entity.CarbonReductionTarget, error) {
	return s.ceRepo.GetTargetByID(ctx, id)
}

func (s *CarbonEmissionService) ListTargets(ctx context.Context, req *QueryCarbonReductionTargetsRequest) ([]*entity.CarbonReductionTarget, int64, error) {
	query := &repository.CarbonReductionTargetQuery{
		Page:      req.Page,
		PageSize:  req.PageSize,
		TargetID:  req.TargetID,
		Status:    req.Status,
		StartYear: req.StartYear,
		EndYear:   req.EndYear,
	}
	return s.ceRepo.ListTargets(ctx, query)
}

func (s *CarbonEmissionService) GetActiveTargets(ctx context.Context, targetID string) ([]*entity.CarbonReductionTarget, error) {
	return s.ceRepo.GetActiveTargets(ctx, targetID)
}
