package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockReportRepo struct {
	mock.Mock
}

func (m *mockReportRepo) GetStationPowerStats(ctx context.Context, stationID string, startTime, endTime time.Time) (*repository.StationPowerStats, error) {
	args := m.Called(ctx, stationID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.StationPowerStats), args.Error(1)
}

func (m *mockReportRepo) GetStationAlarmStats(ctx context.Context, stationID string, startTime, endTime time.Time) (*repository.StationAlarmStats, error) {
	args := m.Called(ctx, stationID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.StationAlarmStats), args.Error(1)
}

func (m *mockReportRepo) GetStationOnlineStats(ctx context.Context, stationID string, startTime, endTime time.Time) (*repository.StationOnlineStats, error) {
	args := m.Called(ctx, stationID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.StationOnlineStats), args.Error(1)
}

func (m *mockReportRepo) GetAllStationPowerStats(ctx context.Context, startTime, endTime time.Time) ([]*repository.StationPowerStats, error) {
	args := m.Called(ctx, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.StationPowerStats), args.Error(1)
}

func (m *mockReportRepo) GetAllStationAlarmStats(ctx context.Context, startTime, endTime time.Time) ([]*repository.StationAlarmStats, error) {
	args := m.Called(ctx, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.StationAlarmStats), args.Error(1)
}

func (m *mockReportRepo) GetAllStationOnlineStats(ctx context.Context, startTime, endTime time.Time) ([]*repository.StationOnlineStats, error) {
	args := m.Called(ctx, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.StationOnlineStats), args.Error(1)
}

func TestReportService_GenerateSingleStationReport(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	powerStats := &repository.StationPowerStats{
		StationID:   "station-001",
		StationName: "PV Station 1",
		TotalPower:  50000.0,
		YoYChange:   5.2,
		MoMChange:   3.1,
	}
	alarmStats := &repository.StationAlarmStats{
		StationID:  "station-001",
		AlarmCount: 10,
	}
	onlineStats := &repository.StationOnlineStats{
		StationID:  "station-001",
		OnlineRate: 98.5,
	}

	repo.On("GetStationPowerStats", ctx, "station-001", startTime, endTime).Return(powerStats, nil)
	repo.On("GetStationAlarmStats", ctx, "station-001", startTime, endTime).Return(alarmStats, nil)
	repo.On("GetStationOnlineStats", ctx, "station-001", startTime, endTime).Return(onlineStats, nil)

	req := &ReportRequest{
		Type:      ReportTypeDaily,
		StartTime: startTime,
		EndTime:   endTime,
		StationID: "station-001",
	}

	report, err := svc.GenerateStationReport(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "2024-01-01", report.StartTime)
	assert.Equal(t, "2024-01-31", report.EndTime)
	assert.Len(t, report.Stations, 1)
	assert.Equal(t, "station-001", report.Stations[0].StationID)
	assert.Equal(t, 50000.0, report.Stations[0].TotalPower)
	assert.Equal(t, 10, report.Stations[0].AlarmCount)
	assert.Equal(t, 98.5, report.Stations[0].OnlineRate)
	assert.Equal(t, 50000.0, report.Summary.TotalPower)
	assert.Equal(t, 10, report.Summary.TotalAlarms)
	assert.Equal(t, 98.5, report.Summary.AvgOnlineRate)
	repo.AssertExpectations(t)
}

func TestReportService_GenerateSingleStationReport_PowerError(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	repo.On("GetStationPowerStats", ctx, "station-001", startTime, endTime).Return(nil, errors.New("db error"))

	req := &ReportRequest{
		Type:      ReportTypeDaily,
		StartTime: startTime,
		EndTime:   endTime,
		StationID: "station-001",
	}

	_, err := svc.GenerateStationReport(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get power stats")
}

func TestReportService_GenerateSingleStationReport_AlarmError(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	powerStats := &repository.StationPowerStats{StationID: "station-001"}
	repo.On("GetStationPowerStats", ctx, "station-001", startTime, endTime).Return(powerStats, nil)
	repo.On("GetStationAlarmStats", ctx, "station-001", startTime, endTime).Return(nil, errors.New("db error"))

	req := &ReportRequest{
		Type:      ReportTypeDaily,
		StartTime: startTime,
		EndTime:   endTime,
		StationID: "station-001",
	}

	_, err := svc.GenerateStationReport(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get alarm stats")
}

func TestReportService_GenerateSingleStationReport_OnlineError(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	powerStats := &repository.StationPowerStats{StationID: "station-001"}
	alarmStats := &repository.StationAlarmStats{StationID: "station-001"}
	repo.On("GetStationPowerStats", ctx, "station-001", startTime, endTime).Return(powerStats, nil)
	repo.On("GetStationAlarmStats", ctx, "station-001", startTime, endTime).Return(alarmStats, nil)
	repo.On("GetStationOnlineStats", ctx, "station-001", startTime, endTime).Return(nil, errors.New("db error"))

	req := &ReportRequest{
		Type:      ReportTypeDaily,
		StartTime: startTime,
		EndTime:   endTime,
		StationID: "station-001",
	}

	_, err := svc.GenerateStationReport(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get online stats")
}

func TestReportService_GenerateAllStationsReport(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	powerStats := []*repository.StationPowerStats{
		{StationID: "s1", StationName: "Station 1", TotalPower: 30000.0, YoYChange: 5.0, MoMChange: 2.0},
		{StationID: "s2", StationName: "Station 2", TotalPower: 20000.0, YoYChange: -1.0, MoMChange: 4.0},
	}
	alarmStats := []*repository.StationAlarmStats{
		{StationID: "s1", AlarmCount: 5},
		{StationID: "s2", AlarmCount: 3},
	}
	onlineStats := []*repository.StationOnlineStats{
		{StationID: "s1", OnlineRate: 99.0},
		{StationID: "s2", OnlineRate: 95.0},
	}

	repo.On("GetAllStationPowerStats", ctx, startTime, endTime).Return(powerStats, nil)
	repo.On("GetAllStationAlarmStats", ctx, startTime, endTime).Return(alarmStats, nil)
	repo.On("GetAllStationOnlineStats", ctx, startTime, endTime).Return(onlineStats, nil)

	req := &ReportRequest{
		Type:      ReportTypeMonthly,
		StartTime: startTime,
		EndTime:   endTime,
	}

	report, err := svc.GenerateStationReport(ctx, req)
	assert.NoError(t, err)
	assert.Len(t, report.Stations, 2)
	assert.Equal(t, 50000.0, report.Summary.TotalPower)
	assert.Equal(t, 8, report.Summary.TotalAlarms)
	assert.InDelta(t, 97.0, report.Summary.AvgOnlineRate, 0.01)
}

func TestReportService_GenerateAllStationsReport_Empty(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	repo.On("GetAllStationPowerStats", ctx, startTime, endTime).Return([]*repository.StationPowerStats{}, nil)
	repo.On("GetAllStationAlarmStats", ctx, startTime, endTime).Return([]*repository.StationAlarmStats{}, nil)
	repo.On("GetAllStationOnlineStats", ctx, startTime, endTime).Return([]*repository.StationOnlineStats{}, nil)

	req := &ReportRequest{
		Type:      ReportTypeMonthly,
		StartTime: startTime,
		EndTime:   endTime,
	}

	report, err := svc.GenerateStationReport(ctx, req)
	assert.NoError(t, err)
	assert.Len(t, report.Stations, 0)
	assert.Equal(t, 0.0, report.Summary.AvgOnlineRate)
}

func TestReportService_GenerateAllStationsReport_PowerError(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	repo.On("GetAllStationPowerStats", ctx, startTime, endTime).Return(nil, errors.New("db error"))

	req := &ReportRequest{
		Type:      ReportTypeMonthly,
		StartTime: startTime,
		EndTime:   endTime,
	}

	_, err := svc.GenerateStationReport(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get power stats")
}

func TestReportService_GenerateAllStationsReport_AlarmError(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	repo.On("GetAllStationPowerStats", ctx, startTime, endTime).Return([]*repository.StationPowerStats{}, nil)
	repo.On("GetAllStationAlarmStats", ctx, startTime, endTime).Return(nil, errors.New("db error"))

	req := &ReportRequest{
		Type:      ReportTypeMonthly,
		StartTime: startTime,
		EndTime:   endTime,
	}

	_, err := svc.GenerateStationReport(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get alarm stats")
}

func TestReportService_GenerateAllStationsReport_OnlineError(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	repo.On("GetAllStationPowerStats", ctx, startTime, endTime).Return([]*repository.StationPowerStats{}, nil)
	repo.On("GetAllStationAlarmStats", ctx, startTime, endTime).Return([]*repository.StationAlarmStats{}, nil)
	repo.On("GetAllStationOnlineStats", ctx, startTime, endTime).Return(nil, errors.New("db error"))

	req := &ReportRequest{
		Type:      ReportTypeMonthly,
		StartTime: startTime,
		EndTime:   endTime,
	}

	_, err := svc.GenerateStationReport(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get online stats")
}

func TestReportService_ExportReport_Excel(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	powerStats := &repository.StationPowerStats{StationID: "s1", StationName: "Station 1", TotalPower: 1000.0}
	alarmStats := &repository.StationAlarmStats{StationID: "s1", AlarmCount: 5}
	onlineStats := &repository.StationOnlineStats{StationID: "s1", OnlineRate: 99.0}

	repo.On("GetStationPowerStats", ctx, "s1", startTime, endTime).Return(powerStats, nil)
	repo.On("GetStationAlarmStats", ctx, "s1", startTime, endTime).Return(alarmStats, nil)
	repo.On("GetStationOnlineStats", ctx, "s1", startTime, endTime).Return(onlineStats, nil)

	req := &ReportRequest{
		Type:      ReportTypeDaily,
		StartTime: startTime,
		EndTime:   endTime,
		StationID: "s1",
	}

	data, filename, err := svc.ExportReport(ctx, req, "excel")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, ".xlsx")
}

func TestReportService_ExportReport_CSV(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	powerStats := &repository.StationPowerStats{StationID: "s1", StationName: "Station 1", TotalPower: 1000.0}
	alarmStats := &repository.StationAlarmStats{StationID: "s1", AlarmCount: 5}
	onlineStats := &repository.StationOnlineStats{StationID: "s1", OnlineRate: 99.0}

	repo.On("GetStationPowerStats", ctx, "s1", startTime, endTime).Return(powerStats, nil)
	repo.On("GetStationAlarmStats", ctx, "s1", startTime, endTime).Return(alarmStats, nil)
	repo.On("GetStationOnlineStats", ctx, "s1", startTime, endTime).Return(onlineStats, nil)

	req := &ReportRequest{
		Type:      ReportTypeDaily,
		StartTime: startTime,
		EndTime:   endTime,
		StationID: "s1",
	}

	data, filename, err := svc.ExportReport(ctx, req, "csv")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, filename, ".csv")
}

func TestReportService_ExportReport_UnsupportedFormat(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	powerStats := &repository.StationPowerStats{StationID: "s1", StationName: "Station 1"}
	alarmStats := &repository.StationAlarmStats{StationID: "s1"}
	onlineStats := &repository.StationOnlineStats{StationID: "s1"}

	repo.On("GetStationPowerStats", ctx, "s1", startTime, endTime).Return(powerStats, nil)
	repo.On("GetStationAlarmStats", ctx, "s1", startTime, endTime).Return(alarmStats, nil)
	repo.On("GetStationOnlineStats", ctx, "s1", startTime, endTime).Return(onlineStats, nil)

	req := &ReportRequest{
		Type:      ReportTypeDaily,
		StartTime: startTime,
		EndTime:   endTime,
		StationID: "s1",
	}

	_, _, err := svc.ExportReport(ctx, req, "pdf")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestReportService_ExportReport_GenerationError(t *testing.T) {
	repo := new(mockReportRepo)
	svc := NewReportService(repo)
	ctx := context.Background()

	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	repo.On("GetStationPowerStats", ctx, "s1", startTime, endTime).Return(nil, errors.New("db error"))

	req := &ReportRequest{
		Type:      ReportTypeDaily,
		StartTime: startTime,
		EndTime:   endTime,
		StationID: "s1",
	}

	_, _, err := svc.ExportReport(ctx, req, "excel")
	assert.Error(t, err)
}
