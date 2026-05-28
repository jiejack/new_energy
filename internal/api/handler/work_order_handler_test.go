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
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockWorkOrderRepo struct {
	mock.Mock
}

func (m *mockWorkOrderRepo) Create(ctx context.Context, workOrder *entity.WorkOrder) error {
	args := m.Called(ctx, workOrder)
	return args.Error(0)
}

func (m *mockWorkOrderRepo) Update(ctx context.Context, workOrder *entity.WorkOrder) error {
	args := m.Called(ctx, workOrder)
	return args.Error(0)
}

func (m *mockWorkOrderRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockWorkOrderRepo) GetByID(ctx context.Context, id string) (*entity.WorkOrder, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkOrder), args.Error(1)
}

func (m *mockWorkOrderRepo) List(ctx context.Context, filter interface{}) ([]*entity.WorkOrder, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkOrder), args.Error(1)
}

func (m *mockWorkOrderRepo) Count(ctx context.Context, filter interface{}) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

type mockDeviceRepoForWO struct {
	mock.Mock
}

func (m *mockDeviceRepoForWO) Create(ctx context.Context, device *entity.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *mockDeviceRepoForWO) Update(ctx context.Context, device *entity.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *mockDeviceRepoForWO) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockDeviceRepoForWO) GetByID(ctx context.Context, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *mockDeviceRepoForWO) GetByCode(ctx context.Context, code string) (*entity.Device, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *mockDeviceRepoForWO) List(ctx context.Context, stationID *string, deviceType *entity.DeviceType) ([]*entity.Device, error) {
	args := m.Called(ctx, stationID, deviceType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Device), args.Error(1)
}

func (m *mockDeviceRepoForWO) GetWithPoints(ctx context.Context, id string) (*entity.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Device), args.Error(1)
}

func (m *mockDeviceRepoForWO) GetOnlineDevices(ctx context.Context, stationID string) ([]*entity.Device, error) {
	args := m.Called(ctx, stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Device), args.Error(1)
}

func setupWorkOrderHandler(woRepo *mockWorkOrderRepo, deviceRepo *mockDeviceRepoForWO) (*WorkOrderHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewWorkOrderService(woRepo, deviceRepo)
	handler := NewWorkOrderHandler(svc)
	r := gin.New()
	return handler, r
}

func TestWorkOrderHandler_CreateWorkOrder_Success(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	deviceRepo := new(mockDeviceRepoForWO)
	handler, r := setupWorkOrderHandler(woRepo, deviceRepo)

	r.POST("/work-orders", handler.CreateWorkOrder)

	woRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.WorkOrder")).Return(nil)

	body := map[string]interface{}{
		"type": "maintenance", "title": "Fix inverter", "priority": "high",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/work-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestWorkOrderHandler_CreateWorkOrder_BadRequest(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	deviceRepo := new(mockDeviceRepoForWO)
	handler, r := setupWorkOrderHandler(woRepo, deviceRepo)

	r.POST("/work-orders", handler.CreateWorkOrder)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/work-orders", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWorkOrderHandler_CreateWorkOrder_ServiceError(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	deviceRepo := new(mockDeviceRepoForWO)
	handler, r := setupWorkOrderHandler(woRepo, deviceRepo)

	r.POST("/work-orders", handler.CreateWorkOrder)

	woRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.WorkOrder")).Return(assert.AnError)

	body := map[string]interface{}{
		"type": "maintenance", "title": "Fix inverter", "priority": "high",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/work-orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWorkOrderHandler_GetWorkOrder_Success(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	deviceRepo := new(mockDeviceRepoForWO)
	handler, r := setupWorkOrderHandler(woRepo, deviceRepo)

	r.GET("/work-orders/:id", handler.GetWorkOrder)

	woRepo.On("GetByID", mock.Anything, "wo-001").Return(&entity.WorkOrder{ID: "wo-001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/work-orders/wo-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWorkOrderHandler_GetWorkOrder_NotFound(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	deviceRepo := new(mockDeviceRepoForWO)
	handler, r := setupWorkOrderHandler(woRepo, deviceRepo)

	r.GET("/work-orders/:id", handler.GetWorkOrder)

	woRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/work-orders/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWorkOrderHandler_ListWorkOrders_Success(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	deviceRepo := new(mockDeviceRepoForWO)
	handler, r := setupWorkOrderHandler(woRepo, deviceRepo)

	r.GET("/work-orders", handler.ListWorkOrders)

	woRepo.On("List", mock.Anything, mock.AnythingOfType("*service.WorkOrderFilter")).Return([]*entity.WorkOrder{{ID: "wo-001"}}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/work-orders", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWorkOrderHandler_ListWorkOrders_Error(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	deviceRepo := new(mockDeviceRepoForWO)
	handler, r := setupWorkOrderHandler(woRepo, deviceRepo)

	r.GET("/work-orders", handler.ListWorkOrders)

	woRepo.On("List", mock.Anything, mock.AnythingOfType("*service.WorkOrderFilter")).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/work-orders", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWorkOrderHandler_UpdateWorkOrder_Success(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	deviceRepo := new(mockDeviceRepoForWO)
	handler, r := setupWorkOrderHandler(woRepo, deviceRepo)

	r.PUT("/work-orders/:id", handler.UpdateWorkOrder)

	existing := &entity.WorkOrder{ID: "wo-001", Title: "Fix inverter"}
	woRepo.On("GetByID", mock.Anything, "wo-001").Return(existing, nil)
	woRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WorkOrder")).Return(nil)

	body := map[string]interface{}{
		"title": "Fix inverter urgently", "priority": "urgent",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/work-orders/wo-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWorkOrderHandler_UpdateWorkOrder_BadRequest(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	deviceRepo := new(mockDeviceRepoForWO)
	handler, r := setupWorkOrderHandler(woRepo, deviceRepo)

	r.PUT("/work-orders/:id", handler.UpdateWorkOrder)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/work-orders/wo-001", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWorkOrderHandler_UpdateWorkOrder_NotFound(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	deviceRepo := new(mockDeviceRepoForWO)
	handler, r := setupWorkOrderHandler(woRepo, deviceRepo)

	r.PUT("/work-orders/:id", handler.UpdateWorkOrder)

	woRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	body := map[string]interface{}{"title": "Updated"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/work-orders/nonexistent", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWorkOrderHandler_DeleteWorkOrder_Success(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	deviceRepo := new(mockDeviceRepoForWO)
	handler, r := setupWorkOrderHandler(woRepo, deviceRepo)

	r.DELETE("/work-orders/:id", handler.DeleteWorkOrder)

	woRepo.On("GetByID", mock.Anything, "wo-001").Return(&entity.WorkOrder{ID: "wo-001"}, nil)
	woRepo.On("Delete", mock.Anything, "wo-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/work-orders/wo-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestWorkOrderHandler_DeleteWorkOrder_NotFound(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	deviceRepo := new(mockDeviceRepoForWO)
	handler, r := setupWorkOrderHandler(woRepo, deviceRepo)

	r.DELETE("/work-orders/:id", handler.DeleteWorkOrder)

	woRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/work-orders/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWorkOrderHandler_GetWorkOrderStats_Success(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	deviceRepo := new(mockDeviceRepoForWO)
	handler, r := setupWorkOrderHandler(woRepo, deviceRepo)

	r.GET("/work-orders/stats", handler.GetWorkOrderStats)

	woRepo.On("Count", mock.Anything, mock.Anything).Return(int64(10), nil)
	woRepo.On("Count", mock.Anything, &service.WorkOrderFilter{Status: "open"}).Return(int64(3), nil)
	woRepo.On("Count", mock.Anything, &service.WorkOrderFilter{Status: "in_progress"}).Return(int64(2), nil)
	woRepo.On("Count", mock.Anything, &service.WorkOrderFilter{Status: "completed"}).Return(int64(5), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/work-orders/stats", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewWorkOrderHandler(t *testing.T) {
	woRepo := new(mockWorkOrderRepo)
	deviceRepo := new(mockDeviceRepoForWO)
	svc := service.NewWorkOrderService(woRepo, deviceRepo)
	handler := NewWorkOrderHandler(svc)
	assert.NotNil(t, handler)
}

var _ repository.WorkOrderRepository = (*mockWorkOrderRepo)(nil)
var _ repository.DeviceRepository = (*mockDeviceRepoForWO)(nil)
