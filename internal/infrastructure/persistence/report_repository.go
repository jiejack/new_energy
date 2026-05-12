package persistence

import (
	"context"
	"time"

	"github.com/new-energy-monitoring/internal/domain/repository"
)

type reportRepository struct {
	db *Database
}

func NewReportRepository(db *Database) repository.ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) GetStationPowerStats(ctx context.Context, stationID string, startTime, endTime time.Time) (*repository.StationPowerStats, error) {
	var stats repository.StationPowerStats
	stats.StationID = stationID
	stats.TotalPower = 0
	stats.YoYChange = 0
	stats.MoMChange = 0
	return &stats, nil
}

func (r *reportRepository) GetStationAlarmStats(ctx context.Context, stationID string, startTime, endTime time.Time) (*repository.StationAlarmStats, error) {
	var stats repository.StationAlarmStats
	stats.StationID = stationID
	stats.AlarmCount = 0
	return &stats, nil
}

func (r *reportRepository) GetStationOnlineStats(ctx context.Context, stationID string, startTime, endTime time.Time) (*repository.StationOnlineStats, error) {
	var stats repository.StationOnlineStats
	stats.StationID = stationID
	stats.OnlineRate = 0
	return &stats, nil
}

func (r *reportRepository) GetAllStationPowerStats(ctx context.Context, startTime, endTime time.Time) ([]*repository.StationPowerStats, error) {
	return []*repository.StationPowerStats{}, nil
}

func (r *reportRepository) GetAllStationAlarmStats(ctx context.Context, startTime, endTime time.Time) ([]*repository.StationAlarmStats, error) {
	return []*repository.StationAlarmStats{}, nil
}

func (r *reportRepository) GetAllStationOnlineStats(ctx context.Context, startTime, endTime time.Time) ([]*repository.StationOnlineStats, error) {
	return []*repository.StationOnlineStats{}, nil
}
