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

func setupCostReportHandler(crRepo *mockCostReportRepo, ceRepo *mockCostEntryRepo) (*CostReportHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewCostReportService(crRepo, ceRepo)
	handler := NewCostReportHandler(svc)
	r := gin.New()
	return handler, r
}

func TestCostReportHandler_CreateCostReport_Success(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.POST("/cost-reports", handler.CreateCostReport)

	crRepo.On("GetByPeriod", mock.Anything, "monthly", mock.Anything, mock.Anything).Return(nil, assert.AnError)
	crRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.CostReport")).Return(nil)

	body := map[string]interface{}{
		"name": "Monthly Report", "report_type": "monthly",
		"period_start": "2024-01-01T00:00:00Z", "period_end": "2024-01-31T00:00:00Z",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cost-reports", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCostReportHandler_CreateCostReport_BadRequest(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.POST("/cost-reports", handler.CreateCostReport)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cost-reports", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCostReportHandler_GetCostReportByID_Success(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.GET("/cost-reports/:id", handler.GetCostReportByID)

	crRepo.On("GetByID", mock.Anything, "cr-001").Return(&entity.CostReport{ID: "cr-001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-reports/cr-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostReportHandler_GetCostReportByID_Error(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.GET("/cost-reports/:id", handler.GetCostReportByID)

	crRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-reports/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostReportHandler_GetCostReportByCode_Success(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.GET("/cost-reports/code/:code", handler.GetCostReportByCode)

	crRepo.On("GetByCode", mock.Anything, "CR001").Return(&entity.CostReport{ID: "cr-001", Code: "CR001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-reports/code/CR001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostReportHandler_GetCostReportByCode_Error(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.GET("/cost-reports/code/:code", handler.GetCostReportByCode)

	crRepo.On("GetByCode", mock.Anything, "INVALID").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-reports/code/INVALID", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostReportHandler_UpdateCostReport_Success(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.PUT("/cost-reports/:id", handler.UpdateCostReport)

	existing := &entity.CostReport{ID: "cr-001", ReportType: "monthly", Status: "draft",
		PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)}
	crRepo.On("GetByID", mock.Anything, "cr-001").Return(existing, nil)
	crRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostReport")).Return(nil)

	body := map[string]interface{}{
		"name": "Updated Report", "report_type": "monthly",
		"period_start": "2024-01-01T00:00:00Z", "period_end": "2024-01-31T00:00:00Z",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cost-reports/cr-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostReportHandler_UpdateCostReport_BadRequest(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.PUT("/cost-reports/:id", handler.UpdateCostReport)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cost-reports/cr-001", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCostReportHandler_DeleteCostReport_Success(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.DELETE("/cost-reports/:id", handler.DeleteCostReport)

	existing := &entity.CostReport{ID: "cr-001", Status: "draft"}
	crRepo.On("GetByID", mock.Anything, "cr-001").Return(existing, nil)
	crRepo.On("Delete", mock.Anything, "cr-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/cost-reports/cr-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostReportHandler_DeleteCostReport_Error(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.DELETE("/cost-reports/:id", handler.DeleteCostReport)

	crRepo.On("GetByID", mock.Anything, "cr-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/cost-reports/cr-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostReportHandler_ListCostReports_Success(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.GET("/cost-reports", handler.ListCostReports)

	crRepo.On("List", mock.Anything, (*string)(nil), (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), 0, 10).Return([]*entity.CostReport{{ID: "cr-001"}}, int64(1), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-reports", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostReportHandler_ListCostReports_Error(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.GET("/cost-reports", handler.ListCostReports)

	crRepo.On("List", mock.Anything, (*string)(nil), (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), 0, 10).Return(([]*entity.CostReport)(nil), int64(0), assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cost-reports", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCostReportHandler_GenerateCostReport_Success(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.POST("/cost-reports/:id/generate", handler.GenerateCostReport)

	existing := &entity.CostReport{ID: "cr-001", Status: "draft",
		PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)}
	crRepo.On("GetByID", mock.Anything, "cr-001").Return(existing, nil)
	ceRepo.On("GetTotalByPeriod", mock.Anything, mock.Anything, mock.Anything).Return(5000.0, nil)
	crRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostReport")).Return(nil)

	body := map[string]interface{}{"generated_by": "admin"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cost-reports/cr-001/generate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostReportHandler_GenerateCostReport_BadRequest(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.POST("/cost-reports/:id/generate", handler.GenerateCostReport)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cost-reports/cr-001/generate", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCostReportHandler_ApproveCostReport_Success(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.PUT("/cost-reports/:id/approve", handler.ApproveCostReport)

	existing := &entity.CostReport{ID: "cr-001", Status: "generated"}
	crRepo.On("GetByID", mock.Anything, "cr-001").Return(existing, nil)
	crRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostReport")).Return(nil)

	body := map[string]interface{}{"approved_by": "admin"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cost-reports/cr-001/approve", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCostReportHandler_ApproveCostReport_BadRequest(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.PUT("/cost-reports/:id/approve", handler.ApproveCostReport)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cost-reports/cr-001/approve", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCostReportHandler_RejectCostReport_Success(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	handler, r := setupCostReportHandler(crRepo, ceRepo)

	r.PUT("/cost-reports/:id/reject", handler.RejectCostReport)

	existing := &entity.CostReport{ID: "cr-001", Status: "generated"}
	crRepo.On("GetByID", mock.Anything, "cr-001").Return(existing, nil)
	crRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.CostReport")).Return(nil)

	body := map[string]interface{}{"approved_by": "admin"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cost-reports/cr-001/reject", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewCostReportHandler(t *testing.T) {
	crRepo := new(mockCostReportRepo)
	ceRepo := new(mockCostEntryRepo)
	svc := service.NewCostReportService(crRepo, ceRepo)
	handler := NewCostReportHandler(svc)
	assert.NotNil(t, handler)
}
