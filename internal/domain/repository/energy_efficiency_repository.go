package repository

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
)

type EnergyEfficiencyRepository interface {
	CreateRecord(ctx context.Context, record *entity.EnergyEfficiencyRecord) error
	BatchCreateRecords(ctx context.Context, records []*entity.EnergyEfficiencyRecord) error
	GetRecordByID(ctx context.Context, id string) (*entity.EnergyEfficiencyRecord, error)
	ListRecords(ctx context.Context, query *EnergyEfficiencyQuery) ([]*entity.EnergyEfficiencyRecord, int64, error)
	GetRecordsByTimeRange(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType, startTime, endTime time.Time) ([]*entity.EnergyEfficiencyRecord, error)
	
	CreateAnalysis(ctx context.Context, analysis *entity.EnergyEfficiencyAnalysis) error
	GetAnalysisByID(ctx context.Context, id string) (*entity.EnergyEfficiencyAnalysis, error)
	ListAnalyses(ctx context.Context, query *EnergyEfficiencyAnalysisQuery) ([]*entity.EnergyEfficiencyAnalysis, int64, error)
	GetLatestAnalysis(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType) (*entity.EnergyEfficiencyAnalysis, error)
	
	GetStatistics(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType, period string, startTime, endTime time.Time) (*EnergyEfficiencyStatistics, error)
	GetBenchmark(ctx context.Context, eeType entity.EnergyEfficiencyType, targetID string) (float64, error)
}

type EnergyEfficiencyQuery struct {
	Page           int
	PageSize       int
	Type           *entity.EnergyEfficiencyType
	Level          *entity.EnergyEfficiencyLevel
	TargetID       *string
	Period         *string
	StartTime      *time.Time
	EndTime        *time.Time
}

type EnergyEfficiencyAnalysisQuery struct {
	Page           int
	PageSize       int
	Type           *entity.EnergyEfficiencyType
	TargetID       *string
	StartTime      *time.Time
	EndTime        *time.Time
}

type EnergyEfficiencyStatistics struct {
	TotalRecords     int64
	AvgEfficiency    float64
	MaxEfficiency    float64
	MinEfficiency    float64
	StdDevEfficiency float64
	TotalInputEnergy float64
	TotalOutputEnergy float64
	ExcellentCount   int64
	GoodCount        int64
	NormalCount      int64
	PoorCount        int64
}
