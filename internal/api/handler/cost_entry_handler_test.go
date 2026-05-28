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

func setupCostEntryHandler(ceRepo *mockCostEntryRepo, ccRepo *mockCostCategoryRepo) (*CostEntryHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewCostEntryService(ceRepo, ccRepo)
	handler := NewCostEntryHandler(svc)
	r := gin.New()
	return handler, r
}

func TestCostEntryHandler_CreateCostEntry_Success(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.POST("/cost-entries", handler.CreateCostEntry)

	ccRepo.On("GetByID", mock.Anything, "cc-001").Return(&entity.CostCategory{ID: "cc-001"}, nil)
	ceRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.CostEntry")).Return(nil)

	body := map[string]interface{}{
		"cost_category_id": "cc-001", "amount": 1000.0, "date": "2024-01-01T00:00:00Z",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cost-entries", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCostEntryHandler_CreateCostEntry_BadRequest(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.POST("/cost-entries", handler.CreateCostEntry)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cost-entries", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCostEntryHandler_GetCostEntryByID_Success(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.GET("/cost-entries/:id", handler.GetCostEntryByID)

	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(&entity.CostEntry{ID: "ce-001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-entries/ce-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostEntryHandler_GetCostEntryByID_Error(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.GET("/cost-entries/:id", handler.GetCostEntryByID)

	ceRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-entries/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostEntryHandler_GetCostEntryByCode_Success(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.GET("/cost-entries/code/:code", handler.GetCostEntryByCode)

	ceRepo.On("GetByCode", mock.Anything, "CE001").Return(&entity.CostEntry{ID: "ce-001", Code: "CE001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-entries/code/CE001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostEntryHandler_GetCostEntryByCode_Error(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.GET("/cost-entries/code/:code", handler.GetCostEntryByCode)

	ceRepo.On("GetByCode", mock.Anything, "INVALID").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-entries/code/INVALID", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostEntryHandler_UpdateCostEntry_Success(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.PUT("/cost-entries/:id", handler.UpdateCostEntry)

	existing := &entity.CostEntry{ID: "ce-001", CostCategoryID: "cc-001", ApprovalStatus: "pending"}
	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(existing, nil)
	ceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostEntry")).Return(nil)

	body := map[string]interface{}{
		"cost_category_id": "cc-001", "amount": 2000.0, "date": "2024-01-01T00:00:00Z",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cost-entries/ce-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostEntryHandler_UpdateCostEntry_BadRequest(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.PUT("/cost-entries/:id", handler.UpdateCostEntry)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cost-entries/ce-001", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCostEntryHandler_DeleteCostEntry_Success(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.DELETE("/cost-entries/:id", handler.DeleteCostEntry)

	existing := &entity.CostEntry{ID: "ce-001", ApprovalStatus: "pending"}
	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(existing, nil)
	ceRepo.On("Delete", mock.Anything, "ce-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/cost-entries/ce-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostEntryHandler_DeleteCostEntry_Error(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.DELETE("/cost-entries/:id", handler.DeleteCostEntry)

	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/cost-entries/ce-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostEntryHandler_ListCostEntries_Success(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.GET("/cost-entries", handler.ListCostEntries)

	ceRepo.On("List", mock.Anything, (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), (*string)(nil), 0, 10).Return([]*entity.CostEntry{{ID: "ce-001"}}, int64(1), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-entries", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostEntryHandler_ListCostEntries_Error(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.GET("/cost-entries", handler.ListCostEntries)

	ceRepo.On("List", mock.Anything, (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), (*string)(nil), 0, 10).Return(([]*entity.CostEntry)(nil), int64(0), assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-entries", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostEntryHandler_ApproveCostEntry_Success(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.PUT("/cost-entries/:id/approve", handler.ApproveCostEntry)

	existing := &entity.CostEntry{ID: "ce-001", ApprovalStatus: "pending"}
	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(existing, nil)
	ceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostEntry")).Return(nil)

	body := map[string]interface{}{"approved_by": "admin"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cost-entries/ce-001/approve", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostEntryHandler_ApproveCostEntry_BadRequest(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.PUT("/cost-entries/:id/approve", handler.ApproveCostEntry)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cost-entries/ce-001/approve", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCostEntryHandler_RejectCostEntry_Success(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.PUT("/cost-entries/:id/reject", handler.RejectCostEntry)

	existing := &entity.CostEntry{ID: "ce-001", ApprovalStatus: "pending"}
	ceRepo.On("GetByID", mock.Anything, "ce-001").Return(existing, nil)
	ceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostEntry")).Return(nil)

	body := map[string]interface{}{"approved_by": "admin"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cost-entries/ce-001/reject", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostEntryHandler_GetTotalByCategory_Success(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.GET("/cost-entries/category/:category_id/total", handler.GetTotalByCategory)

	ccRepo.On("GetByID", mock.Anything, "cc-001").Return(&entity.CostCategory{ID: "cc-001"}, nil)
	ceRepo.On("GetTotalByCategory", mock.Anything, "cc-001", (*time.Time)(nil), (*time.Time)(nil)).Return(1000.0, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-entries/category/cc-001/total", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostEntryHandler_GetTotalByPeriod_Success(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	handler, r := setupCostEntryHandler(ceRepo, ccRepo)

	r.GET("/cost-entries/period/total", handler.GetTotalByPeriod)

	ceRepo.On("GetTotalByPeriod", mock.Anything, (*time.Time)(nil), (*time.Time)(nil)).Return(5000.0, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-entries/period/total", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewCostEntryHandler(t *testing.T) {
	ceRepo := new(mockCostEntryRepo)
	ccRepo := new(mockCostCategoryRepo)
	svc := service.NewCostEntryService(ceRepo, ccRepo)
	handler := NewCostEntryHandler(svc)
	assert.NotNil(t, handler)
}
