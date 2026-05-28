package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAssetRepo struct {
	mock.Mock
}

func (m *mockAssetRepo) Create(ctx context.Context, asset *entity.Asset) error {
	args := m.Called(ctx, asset)
	return args.Error(0)
}

func (m *mockAssetRepo) Update(ctx context.Context, asset *entity.Asset) error {
	args := m.Called(ctx, asset)
	return args.Error(0)
}

func (m *mockAssetRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockAssetRepo) GetByID(ctx context.Context, id string) (*entity.Asset, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Asset), args.Error(1)
}

func (m *mockAssetRepo) GetByCode(ctx context.Context, code string) (*entity.Asset, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Asset), args.Error(1)
}

func (m *mockAssetRepo) List(ctx context.Context, assetType *string, status *string, category *string, offset, limit int) ([]*entity.Asset, int64, error) {
	args := m.Called(ctx, assetType, status, category, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Asset), args.Get(1).(int64), args.Error(2)
}

func (m *mockAssetRepo) GetByLocation(ctx context.Context, location string) ([]*entity.Asset, error) {
	args := m.Called(ctx, location)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Asset), args.Error(1)
}

func (m *mockAssetRepo) GetByDepartment(ctx context.Context, departmentID string) ([]*entity.Asset, error) {
	args := m.Called(ctx, departmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Asset), args.Error(1)
}

func (m *mockAssetRepo) GetByResponsiblePerson(ctx context.Context, person string) ([]*entity.Asset, error) {
	args := m.Called(ctx, person)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Asset), args.Error(1)
}

func (m *mockAssetRepo) GetDepreciatingAssets(ctx context.Context) ([]*entity.Asset, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Asset), args.Error(1)
}

func (m *mockAssetRepo) GetAssetsNearWarrantyEnd(ctx context.Context, days int) ([]*entity.Asset, error) {
	args := m.Called(ctx, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Asset), args.Error(1)
}

type mockAssetMaintenanceRepo struct {
	mock.Mock
}

func (m *mockAssetMaintenanceRepo) Create(ctx context.Context, record *entity.AssetMaintenanceRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *mockAssetMaintenanceRepo) Update(ctx context.Context, record *entity.AssetMaintenanceRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *mockAssetMaintenanceRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockAssetMaintenanceRepo) GetByID(ctx context.Context, id string) (*entity.AssetMaintenanceRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AssetMaintenanceRecord), args.Error(1)
}

func (m *mockAssetMaintenanceRepo) ListByAssetID(ctx context.Context, assetID string, status *string, maintenanceType *string, offset, limit int) ([]*entity.AssetMaintenanceRecord, int64, error) {
	args := m.Called(ctx, assetID, status, maintenanceType, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.AssetMaintenanceRecord), args.Get(1).(int64), args.Error(2)
}

func (m *mockAssetMaintenanceRepo) ListByStatus(ctx context.Context, status string, offset, limit int) ([]*entity.AssetMaintenanceRecord, int64, error) {
	args := m.Called(ctx, status, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.AssetMaintenanceRecord), args.Get(1).(int64), args.Error(2)
}

func (m *mockAssetMaintenanceRepo) GetMaintenanceCostByAsset(ctx context.Context, assetID string, startDate, endDate *time.Time) (float64, error) {
	args := m.Called(ctx, assetID, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

type mockAssetDepreciationRepo struct {
	mock.Mock
}

func (m *mockAssetDepreciationRepo) Create(ctx context.Context, record *entity.AssetDepreciationRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *mockAssetDepreciationRepo) Update(ctx context.Context, record *entity.AssetDepreciationRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *mockAssetDepreciationRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockAssetDepreciationRepo) GetByID(ctx context.Context, id string) (*entity.AssetDepreciationRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AssetDepreciationRecord), args.Error(1)
}

func (m *mockAssetDepreciationRepo) ListByAssetID(ctx context.Context, assetID string, period *string, offset, limit int) ([]*entity.AssetDepreciationRecord, int64, error) {
	args := m.Called(ctx, assetID, period, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.AssetDepreciationRecord), args.Get(1).(int64), args.Error(2)
}

func (m *mockAssetDepreciationRepo) GetLatestByAssetID(ctx context.Context, assetID string) (*entity.AssetDepreciationRecord, error) {
	args := m.Called(ctx, assetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AssetDepreciationRecord), args.Error(1)
}

func (m *mockAssetDepreciationRepo) GetDepreciationSummaryByPeriod(ctx context.Context, period string, startDate, endDate *time.Time) (float64, error) {
	args := m.Called(ctx, period, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

type mockAssetDocumentRepo struct {
	mock.Mock
}

func (m *mockAssetDocumentRepo) Create(ctx context.Context, document *entity.AssetDocument) error {
	args := m.Called(ctx, document)
	return args.Error(0)
}

func (m *mockAssetDocumentRepo) Update(ctx context.Context, document *entity.AssetDocument) error {
	args := m.Called(ctx, document)
	return args.Error(0)
}

func (m *mockAssetDocumentRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockAssetDocumentRepo) GetByID(ctx context.Context, id string) (*entity.AssetDocument, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AssetDocument), args.Error(1)
}

func (m *mockAssetDocumentRepo) ListByAssetID(ctx context.Context, assetID string, documentType *string, offset, limit int) ([]*entity.AssetDocument, int64, error) {
	args := m.Called(ctx, assetID, documentType, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.AssetDocument), args.Get(1).(int64), args.Error(2)
}

func (m *mockAssetDocumentRepo) GetByType(ctx context.Context, documentType string, offset, limit int) ([]*entity.AssetDocument, int64, error) {
	args := m.Called(ctx, documentType, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.AssetDocument), args.Get(1).(int64), args.Error(2)
}

func setupAssetHandler(assetRepo *mockAssetRepo, maintRepo *mockAssetMaintenanceRepo, depRepo *mockAssetDepreciationRepo, docRepo *mockAssetDocumentRepo) (*AssetHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewAssetService(assetRepo, maintRepo, depRepo, docRepo)
	handler := NewAssetHandler(svc)
	r := gin.New()
	return handler, r
}

func TestAssetHandler_CreateAsset_Success(t *testing.T) {
	assetRepo := new(mockAssetRepo)
	maintRepo := new(mockAssetMaintenanceRepo)
	depRepo := new(mockAssetDepreciationRepo)
	docRepo := new(mockAssetDocumentRepo)
	handler, r := setupAssetHandler(assetRepo, maintRepo, depRepo, docRepo)

	r.POST("/assets", handler.CreateAsset)

	assetRepo.On("GetByCode", mock.Anything, "ASSET001").Return(nil, assert.AnError)
	assetRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Asset")).Return(nil)

	body := map[string]interface{}{
		"code": "ASSET001", "name": "Test Asset", "category": "equipment",
		"purchase_price": 100000.0, "purchase_date": "2024-01-01", "expected_life": 10,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/assets", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAssetHandler_CreateAsset_BadRequest(t *testing.T) {
	assetRepo := new(mockAssetRepo)
	maintRepo := new(mockAssetMaintenanceRepo)
	depRepo := new(mockAssetDepreciationRepo)
	docRepo := new(mockAssetDocumentRepo)
	handler, r := setupAssetHandler(assetRepo, maintRepo, depRepo, docRepo)

	r.POST("/assets", handler.CreateAsset)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/assets", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssetHandler_GetAsset_Success(t *testing.T) {
	assetRepo := new(mockAssetRepo)
	maintRepo := new(mockAssetMaintenanceRepo)
	depRepo := new(mockAssetDepreciationRepo)
	docRepo := new(mockAssetDocumentRepo)
	handler, r := setupAssetHandler(assetRepo, maintRepo, depRepo, docRepo)

	r.GET("/assets/:id", handler.GetAsset)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/asset-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssetHandler_GetAsset_NotFound(t *testing.T) {
	assetRepo := new(mockAssetRepo)
	maintRepo := new(mockAssetMaintenanceRepo)
	depRepo := new(mockAssetDepreciationRepo)
	docRepo := new(mockAssetDocumentRepo)
	handler, r := setupAssetHandler(assetRepo, maintRepo, depRepo, docRepo)

	r.GET("/assets/:id", handler.GetAsset)

	assetRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAssetHandler_UpdateAsset_Success(t *testing.T) {
	assetRepo := new(mockAssetRepo)
	maintRepo := new(mockAssetMaintenanceRepo)
	depRepo := new(mockAssetDepreciationRepo)
	docRepo := new(mockAssetDocumentRepo)
	handler, r := setupAssetHandler(assetRepo, maintRepo, depRepo, docRepo)

	r.PUT("/assets/:id", handler.UpdateAsset)

	existing := &entity.Asset{ID: "asset-001", Code: "ASSET001"}
	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(existing, nil)
	assetRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Asset")).Return(nil)

	body := map[string]interface{}{
		"code": "ASSET001", "name": "Updated Asset", "category": "equipment",
		"purchase_price": 120000.0, "purchase_date": "2024-01-01", "expected_life": 10,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/assets/asset-001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssetHandler_UpdateAsset_BadRequest(t *testing.T) {
	assetRepo := new(mockAssetRepo)
	maintRepo := new(mockAssetMaintenanceRepo)
	depRepo := new(mockAssetDepreciationRepo)
	docRepo := new(mockAssetDocumentRepo)
	handler, r := setupAssetHandler(assetRepo, maintRepo, depRepo, docRepo)

	r.PUT("/assets/:id", handler.UpdateAsset)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/assets/asset-001", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssetHandler_DeleteAsset_Success(t *testing.T) {
	assetRepo := new(mockAssetRepo)
	maintRepo := new(mockAssetMaintenanceRepo)
	depRepo := new(mockAssetDepreciationRepo)
	docRepo := new(mockAssetDocumentRepo)
	handler, r := setupAssetHandler(assetRepo, maintRepo, depRepo, docRepo)

	r.DELETE("/assets/:id", handler.DeleteAsset)

	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(&entity.Asset{ID: "asset-001"}, nil)
	maintRepo.On("ListByAssetID", mock.Anything, "asset-001", (*string)(nil), (*string)(nil), 0, 1).Return([]*entity.AssetMaintenanceRecord{}, int64(0), nil)
	depRepo.On("ListByAssetID", mock.Anything, "asset-001", (*string)(nil), 0, 1).Return([]*entity.AssetDepreciationRecord{}, int64(0), nil)
	docRepo.On("ListByAssetID", mock.Anything, "asset-001", (*string)(nil), 0, 1).Return([]*entity.AssetDocument{}, int64(0), nil)
	assetRepo.On("Delete", mock.Anything, "asset-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/assets/asset-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAssetHandler_DeleteAsset_NotFound(t *testing.T) {
	assetRepo := new(mockAssetRepo)
	maintRepo := new(mockAssetMaintenanceRepo)
	depRepo := new(mockAssetDepreciationRepo)
	docRepo := new(mockAssetDocumentRepo)
	handler, r := setupAssetHandler(assetRepo, maintRepo, depRepo, docRepo)

	r.DELETE("/assets/:id", handler.DeleteAsset)

	assetRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/assets/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAssetHandler_ListAssets_Success(t *testing.T) {
	assetRepo := new(mockAssetRepo)
	maintRepo := new(mockAssetMaintenanceRepo)
	depRepo := new(mockAssetDepreciationRepo)
	docRepo := new(mockAssetDocumentRepo)
	handler, r := setupAssetHandler(assetRepo, maintRepo, depRepo, docRepo)

	r.GET("/assets", handler.ListAssets)

	assetRepo.On("List", mock.Anything, (*string)(nil), (*string)(nil), (*string)(nil), 0, 10).Return([]*entity.Asset{{ID: "asset-001"}}, int64(1), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssetHandler_ListAssets_WithFilters(t *testing.T) {
	assetRepo := new(mockAssetRepo)
	maintRepo := new(mockAssetMaintenanceRepo)
	depRepo := new(mockAssetDepreciationRepo)
	docRepo := new(mockAssetDocumentRepo)
	handler, r := setupAssetHandler(assetRepo, maintRepo, depRepo, docRepo)

	r.GET("/assets", handler.ListAssets)

	category := "equipment"
	status := "active"
	assetRepo.On("List", mock.Anything, (*string)(nil), &status, &category, 0, 10).Return([]*entity.Asset{{ID: "asset-001"}}, int64(1), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets?category=equipment&status=active", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssetHandler_CalculateDepreciation_Success(t *testing.T) {
	assetRepo := new(mockAssetRepo)
	maintRepo := new(mockAssetMaintenanceRepo)
	depRepo := new(mockAssetDepreciationRepo)
	docRepo := new(mockAssetDocumentRepo)
	handler, r := setupAssetHandler(assetRepo, maintRepo, depRepo, docRepo)

	r.GET("/assets/:id/depreciation", handler.CalculateDepreciation)

	asset := &entity.Asset{ID: "asset-001", Cost: 100000, UsefulLife: 10, SalvageValue: 10000}
	assetRepo.On("GetByID", mock.Anything, "asset-001").Return(asset, nil)
	depRepo.On("GetLatestByAssetID", mock.Anything, "asset-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/asset-001/depreciation?method=straight-line", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssetHandler_CalculateDepreciation_MissingMethod(t *testing.T) {
	assetRepo := new(mockAssetRepo)
	maintRepo := new(mockAssetMaintenanceRepo)
	depRepo := new(mockAssetDepreciationRepo)
	docRepo := new(mockAssetDocumentRepo)
	handler, r := setupAssetHandler(assetRepo, maintRepo, depRepo, docRepo)

	r.GET("/assets/:id/depreciation", handler.CalculateDepreciation)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/asset-001/depreciation", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssetHandler_CalculateDepreciation_AssetNotFound(t *testing.T) {
	assetRepo := new(mockAssetRepo)
	maintRepo := new(mockAssetMaintenanceRepo)
	depRepo := new(mockAssetDepreciationRepo)
	docRepo := new(mockAssetDocumentRepo)
	handler, r := setupAssetHandler(assetRepo, maintRepo, depRepo, docRepo)

	r.GET("/assets/:id/depreciation", handler.CalculateDepreciation)

	assetRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/nonexistent/depreciation?method=straight-line", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNewAssetHandler(t *testing.T) {
	assetRepo := new(mockAssetRepo)
	maintRepo := new(mockAssetMaintenanceRepo)
	depRepo := new(mockAssetDepreciationRepo)
	docRepo := new(mockAssetDocumentRepo)
	svc := service.NewAssetService(assetRepo, maintRepo, depRepo, docRepo)
	handler := NewAssetHandler(svc)
	assert.NotNil(t, handler)
}

var _ repository.AssetRepository = (*mockAssetRepo)(nil)
var _ repository.AssetMaintenanceRepository = (*mockAssetMaintenanceRepo)(nil)
var _ repository.AssetDepreciationRepository = (*mockAssetDepreciationRepo)(nil)
var _ repository.AssetDocumentRepository = (*mockAssetDocumentRepo)(nil)
