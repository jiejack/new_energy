package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type EnergyEfficiencyService struct {
	eeRepo repository.EnergyEfficiencyRepository
}

func NewEnergyEfficiencyService(eeRepo repository.EnergyEfficiencyRepository) *EnergyEfficiencyService {
	return &EnergyEfficiencyService{eeRepo: eeRepo}
}

type CreateEnergyEfficiencyRecordRequest struct {
	RecordTime   time.Time                `json:"record_time" binding:"required"`
	Type         entity.EnergyEfficiencyType `json:"type" binding:"required"`
	TargetID     string                   `json:"target_id" binding:"required"`
	TargetName   string                   `json:"target_name" binding:"required"`
	InputEnergy  float64                  `json:"input_energy" binding:"required,min=0"`
	OutputEnergy float64                  `json:"output_energy" binding:"required,min=0"`
	Period       string                   `json:"period" binding:"required"`
	Unit         string                   `json:"unit"`
}

type QueryEnergyEfficiencyRecordsRequest struct {
	Page      int                       `form:"page,default=1"`
	PageSize  int                       `form:"page_size,default=10"`
	Type      *entity.EnergyEfficiencyType `form:"type"`
	Level     *entity.EnergyEfficiencyLevel `form:"level"`
	TargetID  *string                   `form:"target_id"`
	Period    *string                   `form:"period"`
	StartTime *time.Time                `form:"start_time"`
	EndTime   *time.Time                `form:"end_time"`
}

type QueryEnergyEfficiencyAnalysesRequest struct {
	Page      int                       `form:"page,default=1"`
	PageSize  int                       `form:"page_size,default=10"`
	Type      *entity.EnergyEfficiencyType `form:"type"`
	TargetID  *string                   `form:"target_id"`
	StartTime *time.Time                `form:"start_time"`
	EndTime   *time.Time                `form:"end_time"`
}

type EnergyEfficiencyTrendData struct {
	Time         time.Time `json:"time"`
	Efficiency   float64   `json:"efficiency"`
	InputEnergy  float64   `json:"input_energy"`
	OutputEnergy float64   `json:"output_energy"`
}

type EnergyEfficiencyComparisonData struct {
	CurrentPeriod  repository.EnergyEfficiencyStatistics `json:"current_period"`
	PreviousPeriod repository.EnergyEfficiencyStatistics `json:"previous_period"`
	YoYChange      float64                               `json:"yoy_change"`
	MoMChange      float64                               `json:"mom_change"`
}

func (s *EnergyEfficiencyService) CreateRecord(ctx context.Context, req *CreateEnergyEfficiencyRecordRequest) (*entity.EnergyEfficiencyRecord, error) {
	record := entity.NewEnergyEfficiencyRecord(
		req.RecordTime,
		req.Type,
		req.TargetID,
		req.TargetName,
		req.InputEnergy,
		req.OutputEnergy,
		req.Period,
	)
	record.ID = uuid.New().String()
	
	if req.Unit != "" {
		record.Unit = req.Unit
	}
	
	benchmark, err := s.eeRepo.GetBenchmark(ctx, req.Type, req.TargetID)
	if err == nil && benchmark > 0 {
		record.BenchmarkEfficiency = benchmark
		if benchmark > 0 {
			record.ImprovementRate = (record.Efficiency - benchmark) / benchmark * 100
		}
	}
	
	if err := s.eeRepo.CreateRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create energy efficiency record: %w", err)
	}
	
	return record, nil
}

func (s *EnergyEfficiencyService) BatchCreateRecords(ctx context.Context, reqs []*CreateEnergyEfficiencyRecordRequest) ([]*entity.EnergyEfficiencyRecord, error) {
	records := make([]*entity.EnergyEfficiencyRecord, 0, len(reqs))
	
	for _, req := range reqs {
		record := entity.NewEnergyEfficiencyRecord(
			req.RecordTime,
			req.Type,
			req.TargetID,
			req.TargetName,
			req.InputEnergy,
			req.OutputEnergy,
			req.Period,
		)
		record.ID = uuid.New().String()
		
		if req.Unit != "" {
			record.Unit = req.Unit
		}
		
		records = append(records, record)
	}
	
	if err := s.eeRepo.BatchCreateRecords(ctx, records); err != nil {
		return nil, fmt.Errorf("failed to batch create energy efficiency records: %w", err)
	}
	
	return records, nil
}

func (s *EnergyEfficiencyService) GetRecord(ctx context.Context, id string) (*entity.EnergyEfficiencyRecord, error) {
	return s.eeRepo.GetRecordByID(ctx, id)
}

func (s *EnergyEfficiencyService) ListRecords(ctx context.Context, req *QueryEnergyEfficiencyRecordsRequest) ([]*entity.EnergyEfficiencyRecord, int64, error) {
	query := &repository.EnergyEfficiencyQuery{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Type:      req.Type,
		Level:     req.Level,
		TargetID:  req.TargetID,
		Period:    req.Period,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
	return s.eeRepo.ListRecords(ctx, query)
}

func (s *EnergyEfficiencyService) GetTrendData(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType, startTime, endTime time.Time) ([]EnergyEfficiencyTrendData, error) {
	records, err := s.eeRepo.GetRecordsByTimeRange(ctx, targetID, eeType, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get records by time range: %w", err)
	}
	
	trendData := make([]EnergyEfficiencyTrendData, 0, len(records))
	for _, record := range records {
		trendData = append(trendData, EnergyEfficiencyTrendData{
			Time:         record.RecordTime,
			Efficiency:   record.Efficiency,
			InputEnergy:  record.InputEnergy,
			OutputEnergy: record.OutputEnergy,
		})
	}
	
	return trendData, nil
}

func (s *EnergyEfficiencyService) GetStatistics(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType, period string, startTime, endTime time.Time) (*repository.EnergyEfficiencyStatistics, error) {
	return s.eeRepo.GetStatistics(ctx, targetID, eeType, period, startTime, endTime)
}

func (s *EnergyEfficiencyService) GetComparisonData(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType, period string, currentStart, currentEnd time.Time) (*EnergyEfficiencyComparisonData, error) {
	duration := currentEnd.Sub(currentStart)
	previousStart := currentStart.Add(-duration)
	previousEnd := currentStart
	lastYearStart := currentStart.AddDate(-1, 0, 0)
	lastYearEnd := currentEnd.AddDate(-1, 0, 0)
	
	currentStats, err := s.eeRepo.GetStatistics(ctx, targetID, eeType, period, currentStart, currentEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get current statistics: %w", err)
	}
	
	previousStats, err := s.eeRepo.GetStatistics(ctx, targetID, eeType, period, previousStart, previousEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous statistics: %w", err)
	}
	
	lastYearStats, err := s.eeRepo.GetStatistics(ctx, targetID, eeType, period, lastYearStart, lastYearEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get last year statistics: %w", err)
	}
	
	comparison := &EnergyEfficiencyComparisonData{
		CurrentPeriod:  *currentStats,
		PreviousPeriod: *previousStats,
	}
	
	if lastYearStats.AvgEfficiency > 0 {
		comparison.YoYChange = (currentStats.AvgEfficiency - lastYearStats.AvgEfficiency) / lastYearStats.AvgEfficiency * 100
	}
	if previousStats.AvgEfficiency > 0 {
		comparison.MoMChange = (currentStats.AvgEfficiency - previousStats.AvgEfficiency) / previousStats.AvgEfficiency * 100
	}
	
	return comparison, nil
}

func (s *EnergyEfficiencyService) CreateAnalysis(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType, targetName string, timeRangeStart, timeRangeEnd time.Time) (*entity.EnergyEfficiencyAnalysis, error) {
	period := "month"
	stats, err := s.eeRepo.GetStatistics(ctx, targetID, eeType, period, timeRangeStart, timeRangeEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics for analysis: %w", err)
	}
	
	analysis := &entity.EnergyEfficiencyAnalysis{
		ID:               uuid.New().String(),
		AnalysisTime:     time.Now(),
		Type:             eeType,
		TargetID:         targetID,
		TargetName:       targetName,
		TimeRangeStart:   timeRangeStart,
		TimeRangeEnd:     timeRangeEnd,
		AvgEfficiency:    stats.AvgEfficiency,
		MaxEfficiency:    stats.MaxEfficiency,
		MinEfficiency:    stats.MinEfficiency,
		StdDevEfficiency: stats.StdDevEfficiency,
		CreatedAt:        time.Now(),
	}
	
	trend := "stable"
	if stats.TotalRecords >= 3 {
		records, err := s.eeRepo.GetRecordsByTimeRange(ctx, targetID, eeType, timeRangeStart, timeRangeEnd)
		if err == nil && len(records) >= 3 {
			firstThird := records[:len(records)/3]
			lastThird := records[len(records)*2/3:]
			
			var firstAvg, lastAvg float64
			for _, r := range firstThird {
				firstAvg += r.Efficiency
			}
			firstAvg /= float64(len(firstThird))
			
			for _, r := range lastThird {
				lastAvg += r.Efficiency
			}
			lastAvg /= float64(len(lastThird))
			
			change := (lastAvg - firstAvg) / firstAvg
			if change > 0.05 {
				trend = "improving"
			} else if change < -0.05 {
				trend = "declining"
			}
		}
	}
	analysis.Trend = trend
	
	suggestions := s.generateOptimizationSuggestions(stats)
	analysis.OptimizationSuggestions = suggestions
	
	if stats.AvgEfficiency < 0.9 {
		benchmark, _ := s.eeRepo.GetBenchmark(ctx, eeType, targetID)
		if benchmark > stats.AvgEfficiency {
			potentialInput := stats.TotalInputEnergy
			analysis.SavingPotential = potentialInput * (benchmark - stats.AvgEfficiency)
		}
	}
	
	duration := timeRangeEnd.Sub(timeRangeStart)
	lastYearStart := timeRangeStart.AddDate(-1, 0, 0)
	lastYearEnd := timeRangeEnd.AddDate(-1, 0, 0)
	lastYearStats, _ := s.eeRepo.GetStatistics(ctx, targetID, eeType, period, lastYearStart, lastYearEnd)
	if lastYearStats.AvgEfficiency > 0 {
		analysis.YoYChange = (stats.AvgEfficiency - lastYearStats.AvgEfficiency) / lastYearStats.AvgEfficiency * 100
	}
	
	previousStart := timeRangeStart.Add(-duration)
	previousEnd := timeRangeStart
	previousStats, _ := s.eeRepo.GetStatistics(ctx, targetID, eeType, period, previousStart, previousEnd)
	if previousStats.AvgEfficiency > 0 {
		analysis.MoMChange = (stats.AvgEfficiency - previousStats.AvgEfficiency) / previousStats.AvgEfficiency * 100
	}
	
	if err := s.eeRepo.CreateAnalysis(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to create analysis: %w", err)
	}
	
	return analysis, nil
}

func (s *EnergyEfficiencyService) GetAnalysis(ctx context.Context, id string) (*entity.EnergyEfficiencyAnalysis, error) {
	return s.eeRepo.GetAnalysisByID(ctx, id)
}

func (s *EnergyEfficiencyService) ListAnalyses(ctx context.Context, req *QueryEnergyEfficiencyAnalysesRequest) ([]*entity.EnergyEfficiencyAnalysis, int64, error) {
	query := &repository.EnergyEfficiencyAnalysisQuery{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Type:      req.Type,
		TargetID:  req.TargetID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
	return s.eeRepo.ListAnalyses(ctx, query)
}

func (s *EnergyEfficiencyService) GetLatestAnalysis(ctx context.Context, targetID string, eeType entity.EnergyEfficiencyType) (*entity.EnergyEfficiencyAnalysis, error) {
	return s.eeRepo.GetLatestAnalysis(ctx, targetID, eeType)
}

func (s *EnergyEfficiencyService) generateOptimizationSuggestions(stats *repository.EnergyEfficiencyStatistics) []string {
	suggestions := make([]string, 0)
	
	if stats.AvgEfficiency < 0.8 {
		suggestions = append(suggestions, "建议开展系统能效审计，识别主要损耗环节")
	}
	
	if stats.PoorCount > 0 {
		poorRatio := float64(stats.PoorCount) / float64(stats.TotalRecords)
		if poorRatio > 0.2 {
			suggestions = append(suggestions, "能效较差时段占比较高，建议优化运行策略")
		}
	}
	
	if stats.StdDevEfficiency > 0.1 {
		suggestions = append(suggestions, "能效波动较大，建议加强设备维护和运行管理")
	}
	
	if stats.MaxEfficiency-stats.MinEfficiency > 0.2 {
		suggestions = append(suggestions, "能效差异较大，建议分析最佳实践并推广")
	}
	
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "当前能效表现良好，建议继续保持并持续监测")
	}
	
	return suggestions
}
