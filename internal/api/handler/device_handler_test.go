package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDeviceRepoForHandler struct {
	mock.Mock
}

func (m *mockDeviceRepoForHandler) Create(ctx context.Context, device *entity.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}
func (m *mockDeviceRepoForHandler) Update(ctx context.Context, device *entity.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}
func (m *mockDeviceRepoForHandler) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockDeviceRepoForHandler) GetByID(ctx context.Context, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}
func (m *mockDeviceRepoForHandler) GetByCode(ctx context.Context, code string) (*entity.Device, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}
func (m *mockDeviceRepoForHandler) List(ctx context.Context, stationID *string, deviceType *entity.DeviceType) ([]*entity.Device, error) {
	args := m.Called(ctx, stationID, deviceType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Device), args.Error(1)
}
func (m *mockDeviceRepoForHandler) GetWithPoints(ctx context.Context, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}
func (m *mockDeviceRepoForHandler) GetOnlineDevices(ctx context.Context, stationID string) ([]*entity.Device, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Device), args.Error(1)
}

type mockPointRepoForHandler struct {
	mock.Mock
}

func (m *mockPointRepoForHandler) Create(ctx context.Context, point *entity.Point) error {
	args := m.Called(ctx, point)
	return args.Error(0)
}
func (m *mockPointRepoForHandler) BatchCreate(ctx context.Context, points []*entity.Point) error {
	args := m.Called(ctx, points)
	return args.Error(0)
}
func (m *mockPointRepoForHandler) Update(ctx context.Context, point *entity.Point) error {
	args := m.Called(ctx, point)
	return args.Error(0)
}
func (m *mockPointRepoForHandler) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockPointRepoForHandler) GetByID(ctx context.Context, id string) (*entity.Point, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Point), args.Error(1)
}
func (m *mockPointRepoForHandler) GetByCode(ctx context.Context, code string) (*entity.Point, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Point), args.Error(1)
}
func (m *mockPointRepoForHandler) List(ctx context.Context, deviceID *string, pointType *entity.PointType) ([]*entity.Point, error) {
	args := m.Called(ctx, deviceID, pointType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}
func (m *mockPointRepoForHandler) GetByStationID(ctx context.Context, stationID string) ([]*entity.Point, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}
func (m *mockPointRepoForHandler) GetByProtocol(ctx context.Context, protocol string) ([]*entity.Point, error) {
	args := m.Called(ctx, protocol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Point), args.Error(1)
}

func setupDeviceHandler(deviceRepo *mockDeviceRepoForHandler, pointRepo *mockPointRepoForHandler) (*DeviceHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewDeviceService(deviceRepo, pointRepo)
	handler := NewDeviceHandler(svc)
	r := gin.New()
	return handler, r
}

func TestDeviceHandler_CreateDevice_Success(t *testing.T) {
	deviceRepo := new(mockDeviceRepoForHandler)
	pointRepo := new(mockPointRepoForHandler)
	handler, r := setupDeviceHandler(deviceRepo, pointRepo)

	r.POST("/devices", handler.CreateDevice)

	deviceRepo.On("GetByCode", mock.Anything, "INV_001").Return(nil, assert.AnError)
	deviceRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Device")).Return(nil)

	body := service.CreateDeviceRequest{
		Code: "INV_001", Name: "Inverter", Type: entity.DeviceTypeInverter, StationID: "station-001",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/devices", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestDeviceHandler_CreateDevice_InvalidJSON(t *testing.T) {
	deviceRepo := new(mockDeviceRepoForHandler)
	pointRepo := new(mockPointRepoForHandler)
	handler, r := setupDeviceHandler(deviceRepo, pointRepo)

	r.POST("/devices", handler.CreateDevice)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/devices", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeviceHandler_CreateDevice_Error(t *testing.T) {
	deviceRepo := new(mockDeviceRepoForHandler)
	pointRepo := new(mockPointRepoForHandler)
	handler, r := setupDeviceHandler(deviceRepo, pointRepo)

	r.POST("/devices", handler.CreateDevice)

	existing := entity.NewDevice("INV_001", "Inverter", entity.DeviceTypeInverter, "station-001")
	deviceRepo.On("GetByCode", mock.Anything, "INV_001").Return(existing, nil)

	body := service.CreateDeviceRequest{
		Code: "INV_001", Name: "Inverter", Type: entity.DeviceTypeInverter, StationID: "station-001",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/devices", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeviceHandler_GetDevice_Success(t *testing.T) {
	deviceRepo := new(mockDeviceRepoForHandler)
	pointRepo := new(mockPointRepoForHandler)
	handler, r := setupDeviceHandler(deviceRepo, pointRepo)

	r.GET("/devices/:id", handler.GetDevice)

	device := entity.NewDevice("INV_001", "Inverter", entity.DeviceTypeInverter, "station-001")
	deviceRepo.On("GetByID", mock.Anything, "device-001").Return(device, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/devices/device-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeviceHandler_GetDevice_NotFound(t *testing.T) {
	deviceRepo := new(mockDeviceRepoForHandler)
	pointRepo := new(mockPointRepoForHandler)
	handler, r := setupDeviceHandler(deviceRepo, pointRepo)

	r.GET("/devices/:id", handler.GetDevice)

	deviceRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/devices/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeviceHandler_ListDevices_Success(t *testing.T) {
	deviceRepo := new(mockDeviceRepoForHandler)
	pointRepo := new(mockPointRepoForHandler)
	handler, r := setupDeviceHandler(deviceRepo, pointRepo)

	r.GET("/devices", handler.ListDevices)

	deviceRepo.On("List", mock.Anything, (*string)(nil), (*entity.DeviceType)(nil)).Return([]*entity.Device{}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/devices", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeviceHandler_ListDevices_Error(t *testing.T) {
	deviceRepo := new(mockDeviceRepoForHandler)
	pointRepo := new(mockPointRepoForHandler)
	handler, r := setupDeviceHandler(deviceRepo, pointRepo)

	r.GET("/devices", handler.ListDevices)

	deviceRepo.On("List", mock.Anything, (*string)(nil), (*entity.DeviceType)(nil)).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/devices", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeviceHandler_UpdateDevice_Success(t *testing.T) {
	deviceRepo := new(mockDeviceRepoForHandler)
	pointRepo := new(mockPointRepoForHandler)
	handler, r := setupDeviceHandler(deviceRepo, pointRepo)

	r.PUT("/devices/:id", handler.UpdateDevice)

	device := entity.NewDevice("INV_001", "Inverter", entity.DeviceTypeInverter, "station-001")
	deviceRepo.On("GetByID", mock.Anything, "device-001").Return(device, nil)
	deviceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Device")).Return(nil)

	body := service.UpdateDeviceRequest{Name: "Updated Inverter"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/devices/device-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeviceHandler_UpdateDevice_InvalidJSON(t *testing.T) {
	deviceRepo := new(mockDeviceRepoForHandler)
	pointRepo := new(mockPointRepoForHandler)
	handler, r := setupDeviceHandler(deviceRepo, pointRepo)

	r.PUT("/devices/:id", handler.UpdateDevice)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/devices/device-001", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeviceHandler_DeleteDevice_Success(t *testing.T) {
	deviceRepo := new(mockDeviceRepoForHandler)
	pointRepo := new(mockPointRepoForHandler)
	handler, r := setupDeviceHandler(deviceRepo, pointRepo)

	r.DELETE("/devices/:id", handler.DeleteDevice)

	device := entity.NewDevice("INV_001", "Inverter", entity.DeviceTypeInverter, "station-001")
	deviceRepo.On("GetByID", mock.Anything, "device-001").Return(device, nil)
	pointRepo.On("List", mock.Anything, &device.ID, (*entity.PointType)(nil)).Return([]*entity.Point{}, nil)
	deviceRepo.On("Delete", mock.Anything, "device-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/devices/device-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeviceHandler_DeleteDevice_NotFound(t *testing.T) {
	deviceRepo := new(mockDeviceRepoForHandler)
	pointRepo := new(mockPointRepoForHandler)
	handler, r := setupDeviceHandler(deviceRepo, pointRepo)

	r.DELETE("/devices/:id", handler.DeleteDevice)

	deviceRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/devices/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeviceHandler_NewDeviceHandler(t *testing.T) {
	deviceRepo := new(mockDeviceRepoForHandler)
	pointRepo := new(mockPointRepoForHandler)
	svc := service.NewDeviceService(deviceRepo, pointRepo)
	handler := NewDeviceHandler(svc)
	assert.NotNil(t, handler)
}
