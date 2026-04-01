package service

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDeviceRepository 设备仓储Mock
type MockDeviceRepository struct {
	mock.Mock
}

func (m *MockDeviceRepository) Create(ctx context.Context, device *entity.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepository) Update(ctx context.Context, device *entity.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDeviceRepository) GetByID(ctx context.Context, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *MockDeviceRepository) GetByCode(ctx context.Context, code string) (*entity.Device, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *MockDeviceRepository) List(ctx context.Context, stationID *string, deviceType *entity.DeviceType) ([]*entity.Device, error) {
	args := m.Called(ctx, stationID, deviceType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Device), args.Error(1)
}

func (m *MockDeviceRepository) GetWithPoints(ctx context.Context, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *MockDeviceRepository) GetOnlineDevices(ctx context.Context, stationID string) ([]*entity.Device, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Device), args.Error(1)
}

// MockPointRepository 测点仓储Mock
type MockPointRepository struct {
	mock.Mock
}

func (m *MockPointRepository) Create(ctx context.Context, point *entity.Point) error {
	args := m.Called(ctx, point)
	return args.Error(0)
}

func (m *MockPointRepository) BatchCreate(ctx context.Context, points []*entity.Point) error {
	args := m.Called(ctx, points)
	return args.Error(0)
}

func (m *MockPointRepository) Update(ctx context.Context, point *entity.Point) error {
	args := m.Called(ctx, point)
	return args.Error(0)
}

func (m *MockPointRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPointRepository) GetByID(ctx context.Context, id string) (*entity.Point, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Point), args.Error(1)
}

func (m *MockPointRepository) GetByCode(ctx context.Context, code string) (*entity.Point, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Point), args.Error(1)
}

func (m *MockPointRepository) List(ctx context.Context, deviceID *string, pointType *entity.PointType) ([]*entity.Point, error) {
	args := m.Called(ctx, deviceID, pointType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}

func (m *MockPointRepository) GetByStationID(ctx context.Context, stationID string) ([]*entity.Point, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}

func (m *MockPointRepository) GetByProtocol(ctx context.Context, protocol string) ([]*entity.Point, error) {
	args := m.Called(ctx, protocol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}

func TestDeviceService_CreateDevice(t *testing.T) {
	ctx := context.Background()

	t.Run("成功创建设备", func(t *testing.T) {
		mockDeviceRepo := new(MockDeviceRepository)
		mockPointRepo := new(MockPointRepository)
		service := NewDeviceService(mockDeviceRepo, mockPointRepo)

		req := &CreateDeviceRequest{
			Code:          "INV001",
			Name:          "逆变器1",
			Type:          entity.DeviceTypeInverter,
			StationID:     "station001",
			Manufacturer:  "华为",
			Model:         "SUN2000",
			RatedPower:    100.0,
			Protocol:      "modbus",
			IPAddress:     "192.168.1.100",
			Port:          502,
			SlaveID:       1,
		}

		mockDeviceRepo.On("GetByCode", ctx, "INV001").Return(nil, errors.New("not found"))
		mockDeviceRepo.On("Create", ctx, mock.AnythingOfType("*entity.Device")).Return(nil)

		device, err := service.CreateDevice(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, device)
		assert.Equal(t, "INV001", device.Code)
		assert.Equal(t, "逆变器1", device.Name)
		assert.Equal(t, entity.DeviceTypeInverter, device.Type)
		mockDeviceRepo.AssertExpectations(t)
	})

	t.Run("设备编码已存在", func(t *testing.T) {
		mockDeviceRepo := new(MockDeviceRepository)
		mockPointRepo := new(MockPointRepository)
		service := NewDeviceService(mockDeviceRepo, mockPointRepo)

		req := &CreateDeviceRequest{
			Code:      "INV001",
			Name:      "逆变器1",
			Type:      entity.DeviceTypeInverter,
			StationID: "station001",
		}

		existingDevice := entity.NewDevice("INV001", "已存在设备", entity.DeviceTypeInverter, "station001")
		mockDeviceRepo.On("GetByCode", ctx, "INV001").Return(existingDevice, nil)

		device, err := service.CreateDevice(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, device)
		assert.Contains(t, err.Error(), "already exists")
		mockDeviceRepo.AssertExpectations(t)
	})
}

func TestDeviceService_UpdateDevice(t *testing.T) {
	ctx := context.Background()

	t.Run("成功更新设备", func(t *testing.T) {
		mockDeviceRepo := new(MockDeviceRepository)
		mockPointRepo := new(MockPointRepository)
		service := NewDeviceService(mockDeviceRepo, mockPointRepo)

		existingDevice := entity.NewDevice("INV001", "逆变器1", entity.DeviceTypeInverter, "station001")
		req := &UpdateDeviceRequest{
			Name:         "逆变器1-更新",
			Manufacturer: "华为",
			Model:        "SUN2000",
		}

		mockDeviceRepo.On("GetByID", ctx, "device001").Return(existingDevice, nil)
		mockDeviceRepo.On("Update", ctx, existingDevice).Return(nil)

		device, err := service.UpdateDevice(ctx, "device001", req)

		assert.NoError(t, err)
		assert.NotNil(t, device)
		assert.Equal(t, "逆变器1-更新", device.Name)
		mockDeviceRepo.AssertExpectations(t)
	})

	t.Run("设备不存在", func(t *testing.T) {
		mockDeviceRepo := new(MockDeviceRepository)
		mockPointRepo := new(MockPointRepository)
		service := NewDeviceService(mockDeviceRepo, mockPointRepo)

		req := &UpdateDeviceRequest{
			Name: "逆变器1-更新",
		}

		mockDeviceRepo.On("GetByID", ctx, "device001").Return(nil, errors.New("not found"))

		device, err := service.UpdateDevice(ctx, "device001", req)

		assert.Error(t, err)
		assert.Nil(t, device)
		mockDeviceRepo.AssertExpectations(t)
	})
}

func TestDeviceService_DeleteDevice(t *testing.T) {
	ctx := context.Background()

	t.Run("成功删除设备", func(t *testing.T) {
		mockDeviceRepo := new(MockDeviceRepository)
		mockPointRepo := new(MockPointRepository)
		service := NewDeviceService(mockDeviceRepo, mockPointRepo)

		device := entity.NewDevice("INV001", "逆变器1", entity.DeviceTypeInverter, "station001")
		deviceID := "device001"

		mockDeviceRepo.On("GetByID", ctx, deviceID).Return(device, nil)
		mockPointRepo.On("List", ctx, &deviceID, (*entity.PointType)(nil)).Return([]*entity.Point{}, nil)
		mockDeviceRepo.On("Delete", ctx, deviceID).Return(nil)

		err := service.DeleteDevice(ctx, deviceID)

		assert.NoError(t, err)
		mockDeviceRepo.AssertExpectations(t)
		mockPointRepo.AssertExpectations(t)
	})

	t.Run("设备存在测点无法删除", func(t *testing.T) {
		mockDeviceRepo := new(MockDeviceRepository)
		mockPointRepo := new(MockPointRepository)
		service := NewDeviceService(mockDeviceRepo, mockPointRepo)

		device := entity.NewDevice("INV001", "逆变器1", entity.DeviceTypeInverter, "station001")
		deviceID := "device001"
		points := []*entity.Point{
			entity.NewPoint("P001", "测点1", entity.PointTypeYaoCe),
		}

		mockDeviceRepo.On("GetByID", ctx, deviceID).Return(device, nil)
		mockPointRepo.On("List", ctx, &deviceID, (*entity.PointType)(nil)).Return(points, nil)

		err := service.DeleteDevice(ctx, deviceID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete device with points")
		mockDeviceRepo.AssertExpectations(t)
		mockPointRepo.AssertExpectations(t)
	})
}

func TestDeviceService_GetDevice(t *testing.T) {
	ctx := context.Background()

	mockDeviceRepo := new(MockDeviceRepository)
	mockPointRepo := new(MockPointRepository)
	service := NewDeviceService(mockDeviceRepo, mockPointRepo)

	expectedDevice := entity.NewDevice("INV001", "逆变器1", entity.DeviceTypeInverter, "station001")
	mockDeviceRepo.On("GetByID", ctx, "device001").Return(expectedDevice, nil)

	device, err := service.GetDevice(ctx, "device001")

	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, "INV001", device.Code)
	mockDeviceRepo.AssertExpectations(t)
}

func TestDeviceService_ListDevices(t *testing.T) {
	ctx := context.Background()

	mockDeviceRepo := new(MockDeviceRepository)
	mockPointRepo := new(MockPointRepository)
	service := NewDeviceService(mockDeviceRepo, mockPointRepo)

	expectedDevices := []*entity.Device{
		entity.NewDevice("INV001", "逆变器1", entity.DeviceTypeInverter, "station001"),
		entity.NewDevice("INV002", "逆变器2", entity.DeviceTypeInverter, "station001"),
	}
	stationID := "station001"
	deviceType := entity.DeviceTypeInverter

	mockDeviceRepo.On("List", ctx, &stationID, &deviceType).Return(expectedDevices, nil)

	devices, err := service.ListDevices(ctx, &stationID, &deviceType)

	assert.NoError(t, err)
	assert.Len(t, devices, 2)
	mockDeviceRepo.AssertExpectations(t)
}

func TestDeviceService_SetDeviceOnline(t *testing.T) {
	ctx := context.Background()

	mockDeviceRepo := new(MockDeviceRepository)
	mockPointRepo := new(MockPointRepository)
	service := NewDeviceService(mockDeviceRepo, mockPointRepo)

	device := entity.NewDevice("INV001", "逆变器1", entity.DeviceTypeInverter, "station001")
	assert.Equal(t, entity.DeviceStatusOffline, device.Status)

	mockDeviceRepo.On("GetByID", ctx, "device001").Return(device, nil)
	mockDeviceRepo.On("Update", ctx, device).Return(nil)

	err := service.SetDeviceOnline(ctx, "device001")

	assert.NoError(t, err)
	assert.Equal(t, entity.DeviceStatusOnline, device.Status)
	mockDeviceRepo.AssertExpectations(t)
}

func TestDeviceService_SetDeviceOffline(t *testing.T) {
	ctx := context.Background()

	mockDeviceRepo := new(MockDeviceRepository)
	mockPointRepo := new(MockPointRepository)
	service := NewDeviceService(mockDeviceRepo, mockPointRepo)

	device := entity.NewDevice("INV001", "逆变器1", entity.DeviceTypeInverter, "station001")
	device.SetOnline()

	mockDeviceRepo.On("GetByID", ctx, "device001").Return(device, nil)
	mockDeviceRepo.On("Update", ctx, device).Return(nil)

	err := service.SetDeviceOffline(ctx, "device001")

	assert.NoError(t, err)
	assert.Equal(t, entity.DeviceStatusOffline, device.Status)
	mockDeviceRepo.AssertExpectations(t)
}
