package service

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAlarmRepositoryForExport 告警仓储Mock
type MockAlarmRepositoryForExport struct {
	mock.Mock
}

func (m *MockAlarmRepositoryForExport) Create(ctx context.Context, alarm *entity.Alarm) error {
	args := m.Called(ctx, alarm)
	return args.Error(0)
}

func (m *MockAlarmRepositoryForExport) Update(ctx context.Context, alarm *entity.Alarm) error {
	args := m.Called(ctx, alarm)
	return args.Error(0)
}

func (m *MockAlarmRepositoryForExport) GetByID(ctx context.Context, id string) (*entity.Alarm, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Alarm), args.Error(1)
}

func (m *MockAlarmRepositoryForExport) GetActiveAlarms(ctx context.Context, stationID *string, level *entity.AlarmLevel) ([]*entity.Alarm, error) {
	args := m.Called(ctx, stationID, level)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Alarm), args.Error(1)
}

func (m *MockAlarmRepositoryForExport) GetHistoryAlarms(ctx context.Context, stationID *string, startTime, endTime int64) ([]*entity.Alarm, error) {
	args := m.Called(ctx, stationID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Alarm), args.Error(1)
}

func (m *MockAlarmRepositoryForExport) Acknowledge(ctx context.Context, id, by string) error {
	args := m.Called(ctx, id, by)
	return args.Error(0)
}

func (m *MockAlarmRepositoryForExport) Clear(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlarmRepositoryForExport) CountByLevel(ctx context.Context, stationID *string) (map[entity.AlarmLevel]int64, error) {
	args := m.Called(ctx, stationID)
	return args.Get(0).(map[entity.AlarmLevel]int64), args.Error(1)
}

// MockDeviceRepositoryForExport 设备仓储Mock
type MockDeviceRepositoryForExport struct {
	mock.Mock
}

func (m *MockDeviceRepositoryForExport) Create(ctx context.Context, device *entity.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepositoryForExport) Update(ctx context.Context, device *entity.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepositoryForExport) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDeviceRepositoryForExport) GetByID(ctx context.Context, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *MockDeviceRepositoryForExport) GetByCode(ctx context.Context, code string) (*entity.Device, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *MockDeviceRepositoryForExport) List(ctx context.Context, stationID *string, deviceType *entity.DeviceType) ([]*entity.Device, error) {
	args := m.Called(ctx, stationID, deviceType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Device), args.Error(1)
}

func (m *MockDeviceRepositoryForExport) GetWithPoints(ctx context.Context, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *MockDeviceRepositoryForExport) GetOnlineDevices(ctx context.Context, stationID string) ([]*entity.Device, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Device), args.Error(1)
}

// MockStationRepositoryForExport 厂站仓储Mock
type MockStationRepositoryForExport struct {
	mock.Mock
}

func (m *MockStationRepositoryForExport) Create(ctx context.Context, station *entity.Station) error {
	args := m.Called(ctx, station)
	return args.Error(0)
}

func (m *MockStationRepositoryForExport) Update(ctx context.Context, station *entity.Station) error {
	args := m.Called(ctx, station)
	return args.Error(0)
}

func (m *MockStationRepositoryForExport) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStationRepositoryForExport) GetByID(ctx context.Context, id string) (*entity.Station, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Station), args.Error(1)
}

func (m *MockStationRepositoryForExport) GetByCode(ctx context.Context, code string) (*entity.Station, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Station), args.Error(1)
}

func (m *MockStationRepositoryForExport) List(ctx context.Context, subRegionID *string, stationType *entity.StationType) ([]*entity.Station, error) {
	args := m.Called(ctx, subRegionID, stationType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Station), args.Error(1)
}

func (m *MockStationRepositoryForExport) GetWithDevices(ctx context.Context, id string) (*entity.Station, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Station), args.Error(1)
}

func TestNewExportService(t *testing.T) {
	mockAlarmRepo := new(MockAlarmRepositoryForExport)
	mockDeviceRepo := new(MockDeviceRepositoryForExport)
	mockStationRepo := new(MockStationRepositoryForExport)

	service := NewExportService(mockAlarmRepo, mockDeviceRepo, mockStationRepo)
	assert.NotNil(t, service)
	assert.Equal(t, mockAlarmRepo, service.alarmRepo)
	assert.Equal(t, mockDeviceRepo, service.deviceRepo)
	assert.Equal(t, mockStationRepo, service.stationRepo)
}

func TestExportService_Export_UnsupportedType(t *testing.T) {
	service := NewExportService(nil, nil, nil)

	req := &ExportRequest{
		Type:   ExportType("unsupported"),
		Format: ExportFormatExcel,
	}

	result, err := service.Export(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported export type")
}

func TestExportService_ExportAlarms_Active(t *testing.T) {
	mockAlarmRepo := new(MockAlarmRepositoryForExport)
	service := NewExportService(mockAlarmRepo, nil, nil)

	alarms := []*entity.Alarm{
		{
			ID:         "alarm-1",
			DeviceID:   "device-1",
			StationID:  "station-1",
			Type:       entity.AlarmTypeDevice,
			Level:      entity.AlarmLevelMajor,
			Title:      "Test Alarm",
			Message:    "Test message",
			Value:      100.0,
			Threshold:  80.0,
			Status:     entity.AlarmStatusActive,
		},
	}

	stationID := "station-1"
	mockAlarmRepo.On("GetActiveAlarms", mock.Anything, &stationID, (*entity.AlarmLevel)(nil)).
		Return(alarms, nil)

	req := &ExportRequest{
		Type:    ExportTypeAlarm,
		Format:  ExportFormatExcel,
		Filters: map[string]interface{}{"station_id": stationID},
	}

	result, err := service.Export(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Buffer)
	assert.Contains(t, result.Filename, "alarms_")
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", result.ContentType)
	mockAlarmRepo.AssertExpectations(t)
}

func TestExportService_ExportAlarms_History(t *testing.T) {
	mockAlarmRepo := new(MockAlarmRepositoryForExport)
	service := NewExportService(mockAlarmRepo, nil, nil)

	alarms := []*entity.Alarm{
		{
			ID:         "alarm-1",
			DeviceID:   "device-1",
			StationID:  "station-1",
			Type:       entity.AlarmTypeDevice,
			Level:      entity.AlarmLevelMajor,
			Title:      "Test Alarm",
			Message:    "Test message",
			Value:      100.0,
			Threshold:  80.0,
			Status:     entity.AlarmStatusCleared,
		},
	}

	stationID := "station-1"
	mockAlarmRepo.On("GetHistoryAlarms", mock.Anything, &stationID, int64(1000000), int64(2000000)).
		Return(alarms, nil)

	req := &ExportRequest{
		Type:      ExportTypeAlarm,
		Format:    ExportFormatExcel,
		StartTime: 1000000,
		EndTime:   2000000,
		Filters:   map[string]interface{}{"station_id": stationID},
	}

	result, err := service.Export(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Buffer)
	mockAlarmRepo.AssertExpectations(t)
}

func TestExportService_ExportAlarms_CSV(t *testing.T) {
	mockAlarmRepo := new(MockAlarmRepositoryForExport)
	service := NewExportService(mockAlarmRepo, nil, nil)

	alarms := []*entity.Alarm{
		{ID: "alarm-1", DeviceID: "device-1"},
	}

	mockAlarmRepo.On("GetActiveAlarms", mock.Anything, (*string)(nil), (*entity.AlarmLevel)(nil)).
		Return(alarms, nil)

	req := &ExportRequest{
		Type:   ExportTypeAlarm,
		Format: ExportFormatCSV,
	}

	result, err := service.Export(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Filename, ".csv")
	assert.Equal(t, "text/csv", result.ContentType)
	mockAlarmRepo.AssertExpectations(t)
}

func TestExportService_ExportAlarms_Error(t *testing.T) {
	mockAlarmRepo := new(MockAlarmRepositoryForExport)
	service := NewExportService(mockAlarmRepo, nil, nil)

	mockAlarmRepo.On("GetActiveAlarms", mock.Anything, (*string)(nil), (*entity.AlarmLevel)(nil)).
		Return(nil, errors.New("database error"))

	req := &ExportRequest{
		Type:   ExportTypeAlarm,
		Format: ExportFormatExcel,
	}

	result, err := service.Export(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get alarms")
	mockAlarmRepo.AssertExpectations(t)
}

func TestExportService_ExportDevices(t *testing.T) {
	mockDeviceRepo := new(MockDeviceRepositoryForExport)
	service := NewExportService(nil, mockDeviceRepo, nil)

	devices := []*entity.Device{
		{
			ID:           "device-1",
			Code:         "DEV001",
			Name:         "Test Device",
			Type:         entity.DeviceTypeInverter,
			StationID:    "station-1",
			Manufacturer: "Test Manufacturer",
			Model:        "Model X",
		},
	}

	stationID := "station-1"
	deviceType := entity.DeviceTypeInverter
	mockDeviceRepo.On("List", mock.Anything, &stationID, &deviceType).
		Return(devices, nil)

	req := &ExportRequest{
		Type:   ExportTypeDevice,
		Format: ExportFormatExcel,
		Filters: map[string]interface{}{
			"station_id": stationID,
			"type":       string(deviceType),
		},
	}

	result, err := service.Export(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Buffer)
	assert.Contains(t, result.Filename, "devices_")
	mockDeviceRepo.AssertExpectations(t)
}

func TestExportService_ExportDevices_CSV(t *testing.T) {
	mockDeviceRepo := new(MockDeviceRepositoryForExport)
	service := NewExportService(nil, mockDeviceRepo, nil)

	devices := []*entity.Device{
		{ID: "device-1", Code: "DEV001"},
	}

	mockDeviceRepo.On("List", mock.Anything, (*string)(nil), (*entity.DeviceType)(nil)).
		Return(devices, nil)

	req := &ExportRequest{
		Type:   ExportTypeDevice,
		Format: ExportFormatCSV,
	}

	result, err := service.Export(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Filename, ".csv")
	mockDeviceRepo.AssertExpectations(t)
}

func TestExportService_ExportDevices_Error(t *testing.T) {
	mockDeviceRepo := new(MockDeviceRepositoryForExport)
	service := NewExportService(nil, mockDeviceRepo, nil)

	mockDeviceRepo.On("List", mock.Anything, (*string)(nil), (*entity.DeviceType)(nil)).
		Return(nil, errors.New("database error"))

	req := &ExportRequest{
		Type:   ExportTypeDevice,
		Format: ExportFormatExcel,
	}

	result, err := service.Export(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get devices")
	mockDeviceRepo.AssertExpectations(t)
}

func TestExportService_ExportStations(t *testing.T) {
	mockStationRepo := new(MockStationRepositoryForExport)
	service := NewExportService(nil, nil, mockStationRepo)

	stations := []*entity.Station{
		{
			ID:           "station-1",
			Code:         "ST001",
			Name:         "Test Station",
			Type:         entity.StationTypePV,
			SubRegionID:  "region-1",
			Capacity:     100.0,
			VoltageLevel: "220V",
		},
	}

	subRegionID := "region-1"
	stationType := entity.StationTypePV
	mockStationRepo.On("List", mock.Anything, &subRegionID, &stationType).
		Return(stations, nil)

	req := &ExportRequest{
		Type:   ExportTypeStation,
		Format: ExportFormatExcel,
		Filters: map[string]interface{}{
			"sub_region_id": subRegionID,
			"type":          string(stationType),
		},
	}

	result, err := service.Export(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Buffer)
	assert.Contains(t, result.Filename, "stations_")
	mockStationRepo.AssertExpectations(t)
}

func TestExportService_ExportStations_CSV(t *testing.T) {
	mockStationRepo := new(MockStationRepositoryForExport)
	service := NewExportService(nil, nil, mockStationRepo)

	stations := []*entity.Station{
		{ID: "station-1", Code: "ST001"},
	}

	mockStationRepo.On("List", mock.Anything, (*string)(nil), (*entity.StationType)(nil)).
		Return(stations, nil)

	req := &ExportRequest{
		Type:   ExportTypeStation,
		Format: ExportFormatCSV,
	}

	result, err := service.Export(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Filename, ".csv")
	mockStationRepo.AssertExpectations(t)
}

func TestExportService_ExportStations_Error(t *testing.T) {
	mockStationRepo := new(MockStationRepositoryForExport)
	service := NewExportService(nil, nil, mockStationRepo)

	mockStationRepo.On("List", mock.Anything, (*string)(nil), (*entity.StationType)(nil)).
		Return(nil, errors.New("database error"))

	req := &ExportRequest{
		Type:   ExportTypeStation,
		Format: ExportFormatExcel,
	}

	result, err := service.Export(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get stations")
	mockStationRepo.AssertExpectations(t)
}

func TestExportService_StreamExportAlarms_Excel(t *testing.T) {
	mockAlarmRepo := new(MockAlarmRepositoryForExport)
	service := NewExportService(mockAlarmRepo, nil, nil)

	alarms := []*entity.Alarm{
		{ID: "alarm-1", DeviceID: "device-1"},
		{ID: "alarm-2", DeviceID: "device-2"},
	}

	stationID := "station-1"
	mockAlarmRepo.On("GetHistoryAlarms", mock.Anything, &stationID, int64(1000000), int64(2000000)).
		Return(alarms, nil)

	req := &ExportRequest{
		Type:      ExportTypeAlarm,
		Format:    ExportFormatExcel,
		StartTime: 1000000,
		EndTime:   2000000,
		Filters:   map[string]interface{}{"station_id": stationID},
	}

	result, err := service.StreamExportAlarms(context.Background(), req, 100)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Buffer)
	mockAlarmRepo.AssertExpectations(t)
}

func TestExportService_StreamExportAlarms_CSV(t *testing.T) {
	mockAlarmRepo := new(MockAlarmRepositoryForExport)
	service := NewExportService(mockAlarmRepo, nil, nil)

	alarms := []*entity.Alarm{
		{ID: "alarm-1", DeviceID: "device-1"},
	}

	stationID := "station-1"
	mockAlarmRepo.On("GetHistoryAlarms", mock.Anything, &stationID, int64(1000000), int64(2000000)).
		Return(alarms, nil)

	req := &ExportRequest{
		Type:      ExportTypeAlarm,
		Format:    ExportFormatCSV,
		StartTime: 1000000,
		EndTime:   2000000,
		Filters:   map[string]interface{}{"station_id": stationID},
	}

	result, err := service.StreamExportAlarms(context.Background(), req, 100)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Buffer)
	assert.Contains(t, result.Filename, ".csv")
	mockAlarmRepo.AssertExpectations(t)
}

func TestExportService_StreamExportAlarms_UnsupportedFormat(t *testing.T) {
	service := NewExportService(nil, nil, nil)

	req := &ExportRequest{
		Type:   ExportTypeAlarm,
		Format: ExportFormat("unsupported"),
	}

	result, err := service.StreamExportAlarms(context.Background(), req, 100)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported export format")
}

// 确保ExportService实现了接口
func TestExportService_ImplementsInterface(t *testing.T) {
	service := NewExportService(nil, nil, nil)
	var _ ExportServiceInterface = service
}
