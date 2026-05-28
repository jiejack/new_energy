package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupCostAllocationHandler(caRepo *mockCostAllocationRepo, ceRepo *mockCostEntryRepo) (*CostAllocationHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewCostAllocationService(caRepo, ceRepo)
	handler := NewCostAllocationHandler(svc)
	r := gin.New()
	return handler, r
}

func TestCostAllocationHandler_CreateCostAllocation_Success(t *testing.T) {
	caRepo := new(mockCostAllocationRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostAllocationHandler(caRepo, ceRepo)

	r.POST("/cost-allocations", handler.CreateCostAllocation)

	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(&entity.CostEntry{ID: "ce-001", ApprovalStatus: "approved", Amount: 10000}, nil)
	caRepo.On("ListByCostEntryID", mock.Anything, "ce-001").Return([]*entity.CostAllocation{}, nil)
	caRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.CostAllocation")).Return(nil)

	body := map[string]interface{}{
		"cost_entry_id": "ce-001", "allocated_to": "project", "allocated_id": "proj-001",
		"amount": 5000.0, "percentage": 50.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cost-allocations", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCostAllocationHandler_CreateCostAllocation_BadRequest(t *testing.T) {
	caRepo := new(mockCostAllocationRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostAllocationHandler(caRepo, ceRepo)

	r.POST("/cost-allocations", handler.CreateCostAllocation)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cost-allocations", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCostAllocationHandler_GetCostAllocationByID_Success(t *testing.T) {
	caRepo := new(mockCostAllocationRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostAllocationHandler(caRepo, ceRepo)

	r.GET("/cost-allocations/:id", handler.GetCostAllocationByID)

	caRepo.On("GetByID", mock.Anything, "ca-001").Return(&entity.CostAllocation{ID: "ca-001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-allocations/ca-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostAllocationHandler_GetCostAllocationByID_Error(t *testing.T) {
	caRepo := new(mockCostAllocationRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostAllocationHandler(caRepo, ceRepo)

	r.GET("/cost-allocations/:id", handler.GetCostAllocationByID)

	caRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-allocations/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostAllocationHandler_UpdateCostAllocation_Success(t *testing.T) {
	caRepo := new(mockCostAllocationRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostAllocationHandler(caRepo, ceRepo)

	r.PUT("/cost-allocations/:id", handler.UpdateCostAllocation)

	caRepo.On("GetByID", mock.Anything, "ca-001").Return(&entity.CostAllocation{ID: "ca-001", CostEntryID: "ce-001"}, nil)
	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(&entity.CostEntry{ID: "ce-001", ApprovalStatus: "approved", Amount: 10000}, nil)
	caRepo.On("ListByCostEntryID", mock.Anything, "ce-001").Return([]*entity.CostAllocation{}, nil)
	caRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostAllocation")).Return(nil)

	body := map[string]interface{}{
		"cost_entry_id": "ce-001", "allocated_to": "project", "allocated_id": "proj-001",
		"amount": 6000.0, "percentage": 60.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cost-allocations/ca-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostAllocationHandler_UpdateCostAllocation_BadRequest(t *testing.T) {
	caRepo := new(mockCostAllocationRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostAllocationHandler(caRepo, ceRepo)

	r.PUT("/cost-allocations/:id", handler.UpdateCostAllocation)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cost-allocations/ca-001", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCostAllocationHandler_DeleteCostAllocation_Success(t *testing.T) {
	caRepo := new(mockCostAllocationRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostAllocationHandler(caRepo, ceRepo)

	r.DELETE("/cost-allocations/:id", handler.DeleteCostAllocation)

	caRepo.On("GetByID", mock.Anything, "ca-001").Return(&entity.CostAllocation{ID: "ca-001"}, nil)
	caRepo.On("Delete", mock.Anything, "ca-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/cost-allocations/ca-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostAllocationHandler_DeleteCostAllocation_Error(t *testing.T) {
	caRepo := new(mockCostAllocationRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostAllocationHandler(caRepo, ceRepo)

	r.DELETE("/cost-allocations/:id", handler.DeleteCostAllocation)

	caRepo.On("GetByID", mock.Anything, "ca-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/cost-allocations/ca-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostAllocationHandler_ListCostAllocationsByCostEntryID_Success(t *testing.T) {
	caRepo := new(mockCostAllocationRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostAllocationHandler(caRepo, ceRepo)

	r.GET("/cost-entries/:cost_entry_id/allocations", handler.ListCostAllocationsByCostEntryID)

	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(&entity.CostEntry{ID: "ce-001"}, nil)
	caRepo.On("ListByCostEntryID", mock.Anything, "ce-001").Return([]*entity.CostAllocation{{ID: "ca-001"}}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-entries/ce-001/allocations", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostAllocationHandler_ListCostAllocationsByAllocated_MissingParams(t *testing.T) {
	caRepo := new(mockCostAllocationRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostAllocationHandler(caRepo, ceRepo)

	r.GET("/cost-allocations/allocated", handler.ListCostAllocationsByAllocated)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-allocations/allocated", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCostAllocationHandler_ListCostAllocationsByAllocated_Success(t *testing.T) {
	caRepo := new(mockCostAllocationRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostAllocationHandler(caRepo, ceRepo)

	r.GET("/cost-allocations/allocated", handler.ListCostAllocationsByAllocated)

	caRepo.On("ListByAllocated", mock.Anything, "project", "proj-001").Return([]*entity.CostAllocation{{ID: "ca-001"}}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-allocations/allocated?allocated_to=project&allocated_id=proj-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostAllocationHandler_GetTotalByAllocated_MissingParams(t *testing.T) {
	caRepo := new(mockCostAllocationRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostAllocationHandler(caRepo, ceRepo)

	r.GET("/cost-allocations/total", handler.GetTotalByAllocated)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-allocations/total", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCostAllocationHandler_GetTotalByAllocated_Success(t *testing.T) {
	caRepo := new(mockCostAllocationRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostAllocationHandler(caRepo, ceRepo)

	r.GET("/cost-allocations/total", handler.GetTotalByAllocated)

	caRepo.On("GetTotalByAllocated", mock.Anything, "project", "proj-001", (*time.Time)(nil), (*time.Time)(nil)).Return(5000.0, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-allocations/total?allocated_to=project&allocated_id=proj-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewCostAllocationHandler(t *testing.T) {
	caRepo := new(mockCostAllocationRepo)
	ceRepo := new(mockCostEntryRepo)
	svc := service.NewCostAllocationService(caRepo, ceRepo)
	handler := NewCostAllocationHandler(svc)
	assert.NotNil(t, handler)
}
