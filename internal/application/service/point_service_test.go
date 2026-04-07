package service

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPointRepositoryForPointService 采集点仓储Mock
type MockPointRepositoryForPointService struct {
	mock.Mock
}

func (m *MockPointRepositoryForPointService) Create(ctx context.Context, point *entity.Point) error {
	args := m.Called(ctx, point)
	return args.Error(0)
}

func (m *MockPointRepositoryForPointService) BatchCreate(ctx context.Context, points []*entity.Point) error {
	args := m.Called(ctx, points)
	return args.Error(0)
}

func (m *MockPointRepositoryForPointService) Update(ctx context.Context, point *entity.Point) error {
	args := m.Called(ctx, point)
	return args.Error(0)
}

func (m *MockPointRepositoryForPointService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPointRepositoryForPointService) GetByID(ctx context.Context, id string) (*entity.Point, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Point), args.Error(1)
}

func (m *MockPointRepositoryForPointService) GetByCode(ctx context.Context, code string) (*entity.Point, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Point), args.Error(1)
}

func (m *MockPointRepositoryForPointService) List(ctx context.Context, deviceID *string, pointType *entity.PointType) ([]*entity.Point, error) {
	args := m.Called(ctx, deviceID, pointType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}

func (m *MockPointRepositoryForPointService) GetByStationID(ctx context.Context, stationID string) ([]*entity.Point, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}

func (m *MockPointRepositoryForPointService) GetByProtocol(ctx context.Context, protocol string) ([]*entity.Point, error) {
	args := m.Called(ctx, protocol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}

func TestPointService_CreatePoint_Success(t *testing.T) {
	ctx := context.Background()

	mockPointRepo := new(MockPointRepositoryForPointService)
	service := NewPointService(mockPointRepo)

	req := &CreatePointRequest{
		Code:         "P001",
		Name:         "电压采集点",
		Type:         entity.PointTypeYaoCe,
		DeviceID:      "device001",
		StationID:     "station001",
		Unit:          "V",
		Precision:     2,
		MinValue:      0,
		MaxValue:      500,
		Protocol:      "modbus",
		Address:       100,
		DataFormat:    "int16",
		ScanInterval:  1000,
		Deadband:      0.1,
		IsAlarm:       true,
		AlarmHigh:     450,
		AlarmLow:      350,
	}

	mockPointRepo.On("GetByCode", ctx, "P001").Return(nil, errors.New("not found"))
	mockPointRepo.On("Create", ctx, mock.AnythingOfType("*entity.Point")).Return(nil)

	point, err := service.CreatePoint(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, point)
	assert.Equal(t, "P001", point.Code)
	assert.Equal(t, "电压采集点", point.Name)
	assert.Equal(t, entity.PointTypeYaoCe, point.Type)
	mockPointRepo.AssertExpectations(t)
}

func TestPointService_CreatePoint_AlreadyExists(t *testing.T) {
	ctx := context.Background()

	mockPointRepo := new(MockPointRepositoryForPointService)
	service := NewPointService(mockPointRepo)

	existingPoint := entity.NewPoint("P001", "已存在采集点", entity.PointTypeYaoCe)
	req := &CreatePointRequest{
		Code: "P001",
		Name: "电压采集点",
		Type: entity.PointTypeYaoCe,
	}

	mockPointRepo.On("GetByCode", ctx, "P001").Return(existingPoint, nil)

	point, err := service.CreatePoint(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.Nil(t, point)
	mockPointRepo.AssertExpectations(t)
}

func TestPointService_BatchCreatePoints_Success(t *testing.T) {
	ctx := context.Background()

	mockPointRepo := new(MockPointRepositoryForPointService)
	service := NewPointService(mockPointRepo)

	reqs := []*CreatePointRequest{
		{
			Code:      "P001",
			Name:      "电压采集点",
			Type:      entity.PointTypeYaoCe,
			DeviceID:  "device001",
			StationID: "station001",
			Unit:      "V",
		},
		{
			Code:      "P002",
			Name:      "电流采集点",
			Type:      entity.PointTypeYaoCe,
			DeviceID:  "device001",
			StationID: "station001",
			Unit:      "A",
		},
	}

	mockPointRepo.On("BatchCreate", ctx, mock.AnythingOfType("[]*entity.Point")).Return(nil)

	err := service.BatchCreatePoints(ctx, reqs)

	assert.NoError(t, err)
	mockPointRepo.AssertExpectations(t)
}

func TestPointService_UpdatePoint_Success(t *testing.T) {
	ctx := context.Background()

	mockPointRepo := new(MockPointRepositoryForPointService)
	service := NewPointService(mockPointRepo)

	existingPoint := entity.NewPoint("P001", "旧名称", entity.PointTypeYaoCe)
	req := &UpdatePointRequest{
		Name:         "新名称",
		Unit:         "V",
		Precision:    2,
		MinValue:     0,
		MaxValue:     500,
		ScanInterval: 1000,
		Deadband:     0.1,
	}

	mockPointRepo.On("GetByID", ctx, "point001").Return(existingPoint, nil)
	mockPointRepo.On("Update", ctx, mock.AnythingOfType("*entity.Point")).Return(nil)

	point, err := service.UpdatePoint(ctx, "point001", req)

	assert.NoError(t, err)
	assert.NotNil(t, point)
	assert.Equal(t, "新名称", point.Name)
	mockPointRepo.AssertExpectations(t)
}

func TestPointService_UpdatePoint_NotFound(t *testing.T) {
	ctx := context.Background()

	mockPointRepo := new(MockPointRepositoryForPointService)
	service := NewPointService(mockPointRepo)

	req := &UpdatePointRequest{
		Name: "新名称",
	}

	mockPointRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	point, err := service.UpdatePoint(ctx, "nonexistent", req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Nil(t, point)
	mockPointRepo.AssertExpectations(t)
}

func TestPointService_DeletePoint_Success(t *testing.T) {
	ctx := context.Background()

	mockPointRepo := new(MockPointRepositoryForPointService)
	service := NewPointService(mockPointRepo)

	mockPointRepo.On("Delete", ctx, "point001").Return(nil)

	err := service.DeletePoint(ctx, "point001")

	assert.NoError(t, err)
	mockPointRepo.AssertExpectations(t)
}

func TestPointService_GetPoint_Success(t *testing.T) {
	ctx := context.Background()

	mockPointRepo := new(MockPointRepositoryForPointService)
	service := NewPointService(mockPointRepo)

	expectedPoint := entity.NewPoint("P001", "电压采集点", entity.PointTypeYaoCe)
	mockPointRepo.On("GetByID", ctx, "point001").Return(expectedPoint, nil)

	point, err := service.GetPoint(ctx, "point001")

	assert.NoError(t, err)
	assert.NotNil(t, point)
	assert.Equal(t, "P001", point.Code)
	mockPointRepo.AssertExpectations(t)
}

func TestPointService_GetPoint_NotFound(t *testing.T) {
	ctx := context.Background()

	mockPointRepo := new(MockPointRepositoryForPointService)
	service := NewPointService(mockPointRepo)

	mockPointRepo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	point, err := service.GetPoint(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, point)
	mockPointRepo.AssertExpectations(t)
}

func TestPointService_ListPoints(t *testing.T) {
	ctx := context.Background()

	mockPointRepo := new(MockPointRepositoryForPointService)
	service := NewPointService(mockPointRepo)

	expectedPoints := []*entity.Point{
		entity.NewPoint("P001", "采集点1", entity.PointTypeYaoCe),
		entity.NewPoint("P002", "采集点2", entity.PointTypeYaoXin),
	}
	deviceID := "device001"
	pointType := entity.PointTypeYaoCe

	mockPointRepo.On("List", ctx, &deviceID, &pointType).Return(expectedPoints, nil)

	points, err := service.ListPoints(ctx, &deviceID, &pointType)

	assert.NoError(t, err)
	assert.Len(t, points, 2)
	mockPointRepo.AssertExpectations(t)
}

func TestPointService_GetPointsByStation(t *testing.T) {
	ctx := context.Background()

	mockPointRepo := new(MockPointRepositoryForPointService)
	service := NewPointService(mockPointRepo)

	expectedPoints := []*entity.Point{
		entity.NewPoint("P001", "采集点1", entity.PointTypeYaoCe),
		entity.NewPoint("P002", "采集点2", entity.PointTypeYaoXin),
	}

	mockPointRepo.On("GetByStationID", ctx, "station001").Return(expectedPoints, nil)

	points, err := service.GetPointsByStation(ctx, "station001")

	assert.NoError(t, err)
	assert.Len(t, points, 2)
	mockPointRepo.AssertExpectations(t)
}

func TestPointService_GetPointsByProtocol(t *testing.T) {
	ctx := context.Background()

	mockPointRepo := new(MockPointRepositoryForPointService)
	service := NewPointService(mockPointRepo)

	expectedPoints := []*entity.Point{
		entity.NewPoint("P001", "采集点1", entity.PointTypeYaoCe),
		entity.NewPoint("P002", "采集点2", entity.PointTypeYaoCe),
	}

	mockPointRepo.On("GetByProtocol", ctx, "modbus").Return(expectedPoints, nil)

	points, err := service.GetPointsByProtocol(ctx, "modbus")

	assert.NoError(t, err)
	assert.Len(t, points, 2)
	mockPointRepo.AssertExpectations(t)
}
