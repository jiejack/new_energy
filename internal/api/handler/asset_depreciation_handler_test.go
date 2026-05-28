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

func setupAssetDepreciationHandler(depRepo *mockAssetDepreciationRepo, assetRepo *mockAssetRepo) (*AssetDepreciationHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewAssetDepreciationService(depRepo, assetRepo)
	handler := NewAssetDepreciationHandler(svc)
	r := gin.New()
	return handler, r
}

func TestAssetDepreciationHandler_CreateDepreciationRecord_Success(t *testing.T) {
	depRepo := new(mockAssetDepreciationRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDepreciationHandler(depRepo, assetRepo)

	r.POST("/assets/depreciation", handler.CreateDepreciationRecord)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	depRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AssetDepreciationRecord")).Return(nil)

	body := map[string]interface{}{
		"asset_id": "asset-001", "depreciation_method": "straight-line",
		"year": 2024, "amount": 9000.0, "accumulated_amount": 9000.0, "book_value": 91000.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/assets/depreciation", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAssetDepreciationHandler_CreateDepreciationRecord_BadRequest(t *testing.T) {
	depRepo := new(mockAssetDepreciationRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDepreciationHandler(depRepo, assetRepo)

	r.POST("/assets/depreciation", handler.CreateDepreciationRecord)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/assets/depreciation", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssetDepreciationHandler_GetDepreciationRecord_Success(t *testing.T) {
	depRepo := new(mockAssetDepreciationRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDepreciationHandler(depRepo, assetRepo)

	r.GET("/assets/depreciation/:id", handler.GetDepreciationRecord)

	depRepo.On("GetByID", mock.Anything, "dep-001").Return(&entity.AssetDepreciationRecord{ID: "dep-001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/depreciation/dep-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssetDepreciationHandler_GetDepreciationRecord_NotFound(t *testing.T) {
	depRepo := new(mockAssetDepreciationRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDepreciationHandler(depRepo, assetRepo)

	r.GET("/assets/depreciation/:id", handler.GetDepreciationRecord)

	depRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/depreciation/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAssetDepreciationHandler_UpdateDepreciationRecord_Success(t *testing.T) {
	depRepo := new(mockAssetDepreciationRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDepreciationHandler(depRepo, assetRepo)

	r.PUT("/assets/depreciation/:id", handler.UpdateDepreciationRecord)

	existing := &entity.AssetDepreciationRecord{ID: "dep-001", AssetID: "asset-001", Period: "annual"}
	depRepo.On("GetByID", mock.Anything, "dep-001").Return(existing, nil)
	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	depRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AssetDepreciationRecord")).Return(nil)

	body := map[string]interface{}{
		"asset_id": "asset-001", "depreciation_method": "straight-line",
		"year": 2024, "amount": 9500.0, "accumulated_amount": 9500.0, "book_value": 90500.0,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/assets/depreciation/dep-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssetDepreciationHandler_UpdateDepreciationRecord_BadRequest(t *testing.T) {
	depRepo := new(mockAssetDepreciationRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDepreciationHandler(depRepo, assetRepo)

	r.PUT("/assets/depreciation/:id", handler.UpdateDepreciationRecord)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/assets/depreciation/dep-001", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssetDepreciationHandler_DeleteDepreciationRecord_Success(t *testing.T) {
	depRepo := new(mockAssetDepreciationRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDepreciationHandler(depRepo, assetRepo)

	r.DELETE("/assets/depreciation/:id", handler.DeleteDepreciationRecord)

	depRepo.On("GetByID", mock.Anything, "dep-001").Return(&entity.AssetDepreciationRecord{ID: "dep-001"}, nil)
	depRepo.On("Delete", mock.Anything, "dep-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/assets/depreciation/dep-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAssetDepreciationHandler_DeleteDepreciationRecord_Error(t *testing.T) {
	depRepo := new(mockAssetDepreciationRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDepreciationHandler(depRepo, assetRepo)

	r.DELETE("/assets/depreciation/:id", handler.DeleteDepreciationRecord)

	depRepo.On("GetByID", mock.Anything, "dep-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/assets/depreciation/dep-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAssetDepreciationHandler_ListDepreciationRecords_Success(t *testing.T) {
	depRepo := new(mockAssetDepreciationRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDepreciationHandler(depRepo, assetRepo)

	r.GET("/assets/depreciation", handler.ListDepreciationRecords)

	depRepo.On("ListByAssetID", mock.Anything, "", (*string)(nil), 0, 10).Return([]*entity.AssetDepreciationRecord{{ID: "dep-001"}}, int64(1), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/depreciation", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssetDepreciationHandler_GetDepreciationSummary_Success(t *testing.T) {
	depRepo := new(mockAssetDepreciationRepo)
	assetRepo := new(mockAssetRepo)
	handler, r := setupAssetDepreciationHandler(depRepo, assetRepo)

	r.GET("/assets/:asset_id/depreciation/summary", handler.GetDepreciationSummary)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	depRepo.On("GetDepreciationSummaryByPeriod", mock.Anything, "annual", (*time.Time)(nil), mock.Anything).Return(9000.0, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/asset-001/depreciation/summary", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewAssetDepreciationHandler(t *testing.T) {
	depRepo := new(mockAssetDepreciationRepo)
	assetRepo := new(mockAssetRepo)
	svc := service.NewAssetDepreciationService(depRepo, assetRepo)
	handler := NewAssetDepreciationHandler(svc)
	assert.NotNil(t, handler)
}
