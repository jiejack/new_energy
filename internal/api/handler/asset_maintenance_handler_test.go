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

func setupAssetMaintenanceHandler(maintRepo *mockAssetMaintenanceRepo, assetRepo *mockAssetRepo) (*AssetMaintenanceHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewAssetMaintenanceService(maintRepo, assetRepo)
	handler := NewAssetMaintenanceHandler(svc)
	r := gin.New()
	return handler, r
}

func TestAssetMaintenanceHandler_CreateMaintenanceRecord_Success(t *testing.T) {
	maintRepo := new(mockAssetMaintenanceRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetMaintenanceHandler(maintRepo, assetRepo)

	r.POST("/assets/maintenance", handler.CreateMaintenanceRecord)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	maintRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AssetMaintenanceRecord")).Return(nil)

	body := map[string]interface{}{
		"asset_id": "asset-001", "maintenance_type": "preventive",
		"maintenance_date": "2024-01-15", "cost": 5000.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/assets/maintenance", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAssetMaintenanceHandler_CreateMaintenanceRecord_BadRequest(t *testing.T) {
	maintRepo := new(mockAssetMaintenanceRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetMaintenanceHandler(maintRepo, assetRepo)

	r.POST("/assets/maintenance", handler.CreateMaintenanceRecord)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/assets/maintenance", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssetMaintenanceHandler_GetMaintenanceRecord_Success(t *testing.T) {
	maintRepo := new(mockAssetMaintenanceRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetMaintenanceHandler(maintRepo, assetRepo)

	r.GET("/assets/maintenance/:id", handler.GetMaintenanceRecord)

	maintRepo.On("GetByID", mock.Anything, "maint-001").Return(&entity.AssetMaintenanceRecord{ID: "maint-001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/maintenance/maint-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssetMaintenanceHandler_GetMaintenanceRecord_NotFound(t *testing.T) {
	maintRepo := new(mockAssetMaintenanceRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetMaintenanceHandler(maintRepo, assetRepo)

	r.GET("/assets/maintenance/:id", handler.GetMaintenanceRecord)

	maintRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/maintenance/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAssetMaintenanceHandler_UpdateMaintenanceRecord_Success(t *testing.T) {
	maintRepo := new(mockAssetMaintenanceRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetMaintenanceHandler(maintRepo, assetRepo)

	r.PUT("/assets/maintenance/:id", handler.UpdateMaintenanceRecord)

	existing := &entity.AssetMaintenanceRecord{ID: "maint-001", AssetID: "asset-001"}
	maintRepo.On("GetByID", mock.Anything, "maint-001").Return(existing, nil)
	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	maintRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AssetMaintenanceRecord")).Return(nil)

	body := map[string]interface{}{
		"asset_id": "asset-001", "maintenance_type": "corrective", "status": "completed",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/assets/maintenance/maint-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssetMaintenanceHandler_UpdateMaintenanceRecord_BadRequest(t *testing.T) {
	maintRepo := new(mockAssetMaintenanceRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetMaintenanceHandler(maintRepo, assetRepo)

	r.PUT("/assets/maintenance/:id", handler.UpdateMaintenanceRecord)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/assets/maintenance/maint-001", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssetMaintenanceHandler_DeleteMaintenanceRecord_Success(t *testing.T) {
	maintRepo := new(mockAssetMaintenanceRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetMaintenanceHandler(maintRepo, assetRepo)

	r.DELETE("/assets/maintenance/:id", handler.DeleteMaintenanceRecord)

	maintRepo.On("GetByID", mock.Anything, "maint-001").Return(&entity.AssetMaintenanceRecord{ID: "maint-001"}, nil)
	maintRepo.On("Delete", mock.Anything, "maint-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/assets/maintenance/maint-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAssetMaintenanceHandler_DeleteMaintenanceRecord_Error(t *testing.T) {
	maintRepo := new(mockAssetMaintenanceRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetMaintenanceHandler(maintRepo, assetRepo)

	r.DELETE("/assets/maintenance/:id", handler.DeleteMaintenanceRecord)

	maintRepo.On("GetByID", mock.Anything, "maint-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/assets/maintenance/maint-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAssetMaintenanceHandler_ListMaintenanceRecords_Success(t *testing.T) {
	maintRepo := new(mockAssetMaintenanceRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetMaintenanceHandler(maintRepo, assetRepo)

	r.GET("/assets/maintenance", handler.ListMaintenanceRecords)

	maintRepo.On("ListByAssetID", mock.Anything, "", (*string)(nil), (*string)(nil), 0, 10).Return([]*entity.AssetMaintenanceRecord{{ID: "maint-001"}}, int64(1), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/maintenance", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssetMaintenanceHandler_ListMaintenanceRecords_Error(t *testing.T) {
	maintRepo := new(mockAssetMaintenanceRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetMaintenanceHandler(maintRepo, assetRepo)

	r.GET("/assets/maintenance", handler.ListMaintenanceRecords)

	maintRepo.On("ListByAssetID", mock.Anything, "", (*string)(nil), (*string)(nil), 0, 10).Return(([]*entity.AssetMaintenanceRecord)(nil), int64(0), assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/maintenance", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAssetMaintenanceHandler_GetMaintenanceCosts_Success(t *testing.T) {
	maintRepo := new(mockAssetMaintenanceRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetMaintenanceHandler(maintRepo, assetRepo)

	r.GET("/assets/:asset_id/maintenance/costs", handler.GetMaintenanceCosts)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	maintRepo.On("GetMaintenanceCostByAsset", mock.Anything, "asset-001", (*time.Time)(nil), (*time.Time)(nil)).Return(5000.0, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/asset-001/maintenance/costs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewAssetMaintenanceHandler(t *testing.T) {
	maintRepo := new(mockAssetMaintenanceRepo)
	assetRepo := new(mockAssetRepo)
	svc := service.NewAssetMaintenanceService(maintRepo, assetRepo)
	handler := NewAssetMaintenanceHandler(svc)
	assert.NotNil(t, handler)
}
