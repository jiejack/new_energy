package repository

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
)

type CarbonEmissionRepository interface {
	CreateFactor(ctx context.Context, factor *entity.CarbonEmissionFactor) error
	UpdateFactor(ctx context.Context, factor *entity.CarbonEmissionFactor) error
	GetFactorByID(ctx context.Context, id string) (*entity.CarbonEmissionFactor, error)
	GetFactorByCode(ctx context.Context, code string) (*entity.CarbonEmissionFactor, error)
	ListFactors(ctx context.Context, query *CarbonEmissionFactorQuery) ([]*entity.CarbonEmissionFactor, int64, error)
	GetActiveFactors(ctx context.Context, scope *entity.CarbonEmissionScope) ([]*entity.CarbonEmissionFactor, error)
	
	CreateRecord(ctx context.Context, record *entity.CarbonEmissionRecord) error
	BatchCreateRecords(ctx context.Context, records []*entity.CarbonEmissionRecord) error
	UpdateRecord(ctx context.Context, record *entity.CarbonEmissionRecord) error
	GetRecordByID(ctx context.Context, id string) (*entity.CarbonEmissionRecord, error)
	ListRecords(ctx context.Context, query *CarbonEmissionRecordQuery) ([]*entity.CarbonEmissionRecord, int64, error)
	GetRecordsByTimeRange(ctx context.Context, targetID string, scope *entity.CarbonEmissionScope, startTime, endTime time.Time) ([]*entity.CarbonEmissionRecord, error)
	GetTotalEmissionByScope(ctx context.Context, targetID string, scope entity.CarbonEmissionScope, period string, startTime, endTime time.Time) (float64, error)
	
	CreateSummary(ctx context.Context, summary *entity.CarbonEmissionSummary) error
	UpdateSummary(ctx context.Context, summary *entity.CarbonEmissionSummary) error
	GetSummaryByID(ctx context.Context, id string) (*entity.CarbonEmissionSummary, error)
	GetLatestSummary(ctx context.Context, targetID string, period string) (*entity.CarbonEmissionSummary, error)
	ListSummaries(ctx context.Context, query *CarbonEmissionSummaryQuery) ([]*entity.CarbonEmissionSummary, int64, error)
	
	CreateTarget(ctx context.Context, target *entity.CarbonReductionTarget) error
	UpdateTarget(ctx context.Context, target *entity.CarbonReductionTarget) error
	GetTargetByID(ctx context.Context, id string) (*entity.CarbonReductionTarget, error)
	ListTargets(ctx context.Context, query *CarbonReductionTargetQuery) ([]*entity.CarbonReductionTarget, int64, error)
	GetActiveTargets(ctx context.Context, targetID string) ([]*entity.CarbonReductionTarget, error)
}

type CarbonEmissionFactorQuery struct {
	Page      int
	PageSize  int
	Scope     *entity.CarbonEmissionScope
	Source    *string
	IsActive  *bool
}

type CarbonEmissionRecordQuery struct {
	Page      int
	PageSize  int
	Scope     *entity.CarbonEmissionScope
	TargetID  *string
	Period    *string
	Status    *entity.CarbonEmissionStatus
	StartTime *time.Time
	EndTime   *time.Time
}

type CarbonEmissionSummaryQuery struct {
	Page      int
	PageSize  int
	TargetID  *string
	Period    *string
	Status    *entity.CarbonEmissionStatus
	StartTime *time.Time
	EndTime   *time.Time
}

type CarbonReductionTargetQuery struct {
	Page      int
	PageSize  int
	TargetID  *string
	Status    *string
	StartYear *int
	EndYear   *int
}
