package service

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStationRepository 厂站仓储Mock
type MockStationRepository struct {
	mock.Mock
}

func (m *MockStationRepository) Create(ctx context.Context, station *entity.Station) error {
	args := m.Called(ctx, station)
	return args.Error(0)
}

func (m *MockStationRepository) Update(ctx context.Context, station *entity.Station) error {
	args := m.Called(ctx, station)
	return args.Error(0)
}

func (m *MockStationRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStationRepository) GetByID(ctx context.Context, id string) (*entity.Station, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Station), args.Error(1)
}

func (m *MockStationRepository) GetByCode(ctx context.Context, code string) (*entity.Station, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Station), args.Error(1)
}

func (m *MockStationRepository) List(ctx context.Context, subRegionID *string, stationType *entity.StationType) ([]*entity.Station, error) {
	args := m.Called(ctx, subRegionID, stationType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Station), args.Error(1)
}

func (m *MockStationRepository) GetWithDevices(ctx context.Context, id string) (*entity.Station, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Station), args.Error(1)
}

// MockDeviceRepository 设备仓储Mock
type MockDeviceRepositoryForStationService struct {
	mock.Mock
}

func (m *MockDeviceRepositoryForStationService) Create(ctx context.Context, device *entity.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepositoryForStationService) Update(ctx context.Context, device *entity.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepositoryForStationService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDeviceRepositoryForStationService) GetByID(ctx context.Context, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *MockDeviceRepositoryForStationService) GetByCode(ctx context.Context, code string) (*entity.Device, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *MockDeviceRepositoryForStationService) List(ctx context.Context, stationID *string, deviceType *entity.DeviceType) ([]*entity.Device, error) {
	args := m.Called(ctx, stationID, deviceType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Device), args.Error(1)
}

func (m *MockDeviceRepositoryForStationService) GetWithPoints(ctx context.Context, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *MockDeviceRepositoryForStationService) GetOnlineDevices(ctx context.Context, stationID string) ([]*entity.Device, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Device), args.Error(1)
}

// MockPointRepository 采集点仓储Mock
type MockPointRepositoryForStationService struct {
	mock.Mock
}

func (m *MockPointRepositoryForStationService) Create(ctx context.Context, point *entity.Point) error {
	args := m.Called(ctx, point)
	return args.Error(0)
}

func (m *MockPointRepositoryForStationService) BatchCreate(ctx context.Context, points []*entity.Point) error {
	args := m.Called(ctx, points)
	return args.Error(0)
}

func (m *MockPointRepositoryForStationService) Update(ctx context.Context, point *entity.Point) error {
	args := m.Called(ctx, point)
	return args.Error(0)
}

func (m *MockPointRepositoryForStationService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPointRepositoryForStationService) GetByID(ctx context.Context, id string) (*entity.Point, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Point), args.Error(1)
}

func (m *MockPointRepositoryForStationService) GetByCode(ctx context.Context, code string) (*entity.Point, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
	 return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Point), args.Error(1)
}

func (m *MockPointRepositoryForStationService) List(ctx context.Context, deviceID *string, pointType *entity.PointType) ([]*entity.Point, error) {
	args := m.Called(ctx, deviceID, pointType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}

func (m *MockPointRepositoryForStationService) GetByStationID(ctx context.Context, stationID string) ([]*entity.Point, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}

func (m *MockPointRepositoryForStationService) GetByProtocol(ctx context.Context, protocol string) ([]*entity.Point, error) {
	args := m.Called(ctx, protocol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}

func TestStationService_CreateStation_Success(t *testing.T) {
	ctx := context.Background()

	mockStationRepo := new(MockStationRepository)
	mockDeviceRepo := new(MockDeviceRepositoryForStationService)
	mockPointRepo := new(MockPointRepositoryForStationService)
	service := NewStationService(mockStationRepo, mockDeviceRepo, mockPointRepo)

	req := &CreateStationRequest{
		Code:        "ST001",
		Name:        "测试电站",
		Type:        entity.StationTypePV,
		SubRegionID:  "region001",
		Capacity:     100.0,
		VoltageLevel: "110kV",
		Longitude:   116.4074,
		Latitude:    39.9042,
		Address:     "北京市朝阳区",
	}

	mockStationRepo.On("GetByCode", ctx, "ST001").Return(nil, errors.New("not found"))
	mockStationRepo.On("Create", ctx, mock.AnythingOfType("*entity.Station")).Return(nil)

	station, err := service.CreateStation(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, station)
	assert.Equal(t, "ST001", station.Code)
	assert.Equal(t, "测试电站", station.Name)
	assert.Equal(t, entity.StationTypePV, station.Type)
	mockStationRepo.AssertExpectations(t)
}

func TestStationService_CreateStation_AlreadyExists(t *testing.T) {
	ctx := context.Background()

	mockStationRepo := new(MockStationRepository)
	mockDeviceRepo := new(MockDeviceRepositoryForStationService)
	mockPointRepo := new(MockPointRepositoryForStationService)
	service := NewStationService(mockStationRepo, mockDeviceRepo, mockPointRepo)

	existingStation := entity.NewStation("ST001", "已存在电站", entity.StationTypePV, "region001")
	req := &CreateStationRequest{
		Code:       "ST001",
		Name:       "测试电站",
		Type:       entity.StationTypePV,
		SubRegionID: "region001",
	}

	mockStationRepo.On("GetByCode", ctx, "ST001").Return(existingStation, nil)

	station, err := service.CreateStation(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.Nil(t, station)
	mockStationRepo.AssertExpectations(t)
}

func TestStationService_GetStation_Success(t *testing.T) {
	ctx := context.Background()

	mockStationRepo := new(MockStationRepository)
	mockDeviceRepo := new(MockDeviceRepositoryForStationService)
	mockPointRepo := new(MockPointRepositoryForStationService)
	service := NewStationService(mockStationRepo, mockDeviceRepo, mockPointRepo)

	expectedStation := entity.NewStation("ST001", "测试电站", entity.StationTypePV, "region001")
	mockStationRepo.On("GetByID", ctx, "station001").Return(expectedStation, nil)

	station, err := service.GetStation(ctx, "station001")

	assert.NoError(t, err)
	assert.NotNil(t, station)
	assert.Equal(t, "ST001", station.Code)
	mockStationRepo.AssertExpectations(t)
}

func TestStationService_GetStation_NotFound(t *testing.T) {
	ctx := context.Background()

	mockStationRepo := new(MockStationRepository)
	mockDeviceRepo := new(MockDeviceRepositoryForStationService)
	mockPointRepo := new(MockPointRepositoryForStationService)
	service := NewStationService(mockStationRepo, mockDeviceRepo, mockPointRepo)

	mockStationRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	station, err := service.GetStation(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, station)
	mockStationRepo.AssertExpectations(t)
}

func TestStationService_UpdateStation_Success(t *testing.T) {
	ctx := context.Background()

	mockStationRepo := new(MockStationRepository)
	mockDeviceRepo := new(MockDeviceRepositoryForStationService)
	mockPointRepo := new(MockPointRepositoryForStationService)
	service := NewStationService(mockStationRepo, mockDeviceRepo, mockPointRepo)

	existingStation := entity.NewStation("ST001", "旧名称", entity.StationTypePV, "region001")
	req := &UpdateStationRequest{
		Name:         "新名称",
		Capacity:     200.0,
		VoltageLevel: "220kV",
		Address:      "新地址",
	}

	mockStationRepo.On("GetByID", ctx, "station001").Return(existingStation, nil)
	mockStationRepo.On("Update", ctx, mock.AnythingOfType("*entity.Station")).Return(nil)

	station, err := service.UpdateStation(ctx, "station001", req)

	assert.NoError(t, err)
	assert.NotNil(t, station)
	assert.Equal(t, "新名称", station.Name)
	mockStationRepo.AssertExpectations(t)
}

func TestStationService_UpdateStation_NotFound(t *testing.T) {
	ctx := context.Background()

	mockStationRepo := new(MockStationRepository)
	mockDeviceRepo := new(MockDeviceRepositoryForStationService)
	mockPointRepo := new(MockPointRepositoryForStationService)
	service := NewStationService(mockStationRepo, mockDeviceRepo, mockPointRepo)

	req := &UpdateStationRequest{
		Name: "新名称",
	}

	mockStationRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	station, err := service.UpdateStation(ctx, "nonexistent", req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Nil(t, station)
	mockStationRepo.AssertExpectations(t)
}

func TestStationService_DeleteStation_Success(t *testing.T) {
	ctx := context.Background()

	mockStationRepo := new(MockStationRepository)
	mockDeviceRepo := new(MockDeviceRepositoryForStationService)
	mockPointRepo := new(MockPointRepositoryForStationService)
	service := NewStationService(mockStationRepo, mockDeviceRepo, mockPointRepo)

	existingStation := entity.NewStation("ST001", "测试电站", entity.StationTypePV, "region001")
	existingStation.ID = "station001"

	mockStationRepo.On("GetByID", ctx, "station001").Return(existingStation, nil)
	mockDeviceRepo.On("List", ctx, &existingStation.ID, (*entity.DeviceType)(nil)).Return([]*entity.Device{}, nil)
	mockStationRepo.On("Delete", ctx, "station001").Return(nil)

	err := service.DeleteStation(ctx, "station001")

	assert.NoError(t, err)
	mockStationRepo.AssertExpectations(t)
	mockDeviceRepo.AssertExpectations(t)
}

func TestStationService_DeleteStation_WithDevices(t *testing.T) {
	ctx := context.Background()

	mockStationRepo := new(MockStationRepository)
	mockDeviceRepo := new(MockDeviceRepositoryForStationService)
	mockPointRepo := new(MockPointRepositoryForStationService)
	service := NewStationService(mockStationRepo, mockDeviceRepo, mockPointRepo)

	existingStation := entity.NewStation("ST001", "测试电站", entity.StationTypePV, "region001")
	existingStation.ID = "station001"

	devices := []*entity.Device{
		entity.NewDevice("DEV001", "设备1", entity.DeviceTypeInverter, "station001"),
	}

	mockStationRepo.On("GetByID", ctx, "station001").Return(existingStation, nil)
	mockDeviceRepo.On("List", ctx, &existingStation.ID, (*entity.DeviceType)(nil)).Return(devices, nil)

	err := service.DeleteStation(ctx, "station001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete station with devices")
	mockStationRepo.AssertExpectations(t)
	mockDeviceRepo.AssertExpectations(t)
}

func TestStationService_ListStations(t *testing.T) {
	ctx := context.Background()

	mockStationRepo := new(MockStationRepository)
	mockDeviceRepo := new(MockDeviceRepositoryForStationService)
	mockPointRepo := new(MockPointRepositoryForStationService)
	service := NewStationService(mockStationRepo, mockDeviceRepo, mockPointRepo)

	expectedStations := []*entity.Station{
		entity.NewStation("ST001", "电站1", entity.StationTypePV, "region001"),
		entity.NewStation("ST002", "电站2", entity.StationTypeWind, "region001"),
	}
	subRegionID := "region001"
	stationType := entity.StationTypePV

	mockStationRepo.On("List", ctx, &subRegionID, &stationType).Return(expectedStations, nil)

	stations, err := service.ListStations(ctx, &subRegionID, &stationType)

	assert.NoError(t, err)
	assert.Len(t, stations, 2)
	mockStationRepo.AssertExpectations(t)
}

func TestStationService_GetStationWithDevices(t *testing.T) {
	ctx := context.Background()

	mockStationRepo := new(MockStationRepository)
	mockDeviceRepo := new(MockDeviceRepositoryForStationService)
	mockPointRepo := new(MockPointRepositoryForStationService)
	service := NewStationService(mockStationRepo, mockDeviceRepo, mockPointRepo)

	expectedStation := entity.NewStation("ST001", "测试电站", entity.StationTypePV, "region001")
	expectedStation.Devices = []*entity.Device{
		entity.NewDevice("DEV001", "设备1", entity.DeviceTypeInverter, "station001"),
	}

	mockStationRepo.On("GetWithDevices", ctx, "station001").Return(expectedStation, nil)

	station, err := service.GetStationWithDevices(ctx, "station001")

	assert.NoError(t, err)
	assert.NotNil(t, station)
	assert.Len(t, station.Devices, 1)
	mockStationRepo.AssertExpectations(t)
}

func TestStationService_GetStationPoints(t *testing.T) {
	ctx := context.Background()

	mockStationRepo := new(MockStationRepository)
	mockDeviceRepo := new(MockDeviceRepositoryForStationService)
	mockPointRepo := new(MockPointRepositoryForStationService)
	service := NewStationService(mockStationRepo, mockDeviceRepo, mockPointRepo)

	expectedPoints := []*entity.Point{
		entity.NewPoint("P001", "采集点1", entity.PointTypeYaoCe),
		entity.NewPoint("P002", "采集点2", entity.PointTypeYaoXin),
	}

	mockPointRepo.On("GetByStationID", ctx, "station001").Return(expectedPoints, nil)

	points, err := service.GetStationPoints(ctx, "station001")

	assert.NoError(t, err)
	assert.Len(t, points, 2)
	mockPointRepo.AssertExpectations(t)
}
