package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockModelService struct {
	mock.Mock
}

func (m *mockModelService) RegisterModel(ctx context.Context, modelName, version, artifactPath string, accuracy float64) (*entity.ModelVersion, error) {
	args := m.Called(ctx, modelName, version, artifactPath, accuracy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ModelVersion), args.Error(1)
}
func (m *mockModelService) GetModel(ctx context.Context, id string) (*entity.ModelVersion, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ModelVersion), args.Error(1)
}
func (m *mockModelService) GetProductionModel(ctx context.Context, modelName string) (*entity.ModelVersion, error) {
	args := m.Called(ctx, modelName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ModelVersion), args.Error(1)
}
func (m *mockModelService) ListModels(ctx context.Context, modelName string) ([]*entity.ModelVersion, error) {
	args := m.Called(ctx, modelName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ModelVersion), args.Error(1)
}
func (m *mockModelService) PromoteToProduction(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockModelService) RetireModel(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupModelHandler(svc *mockModelService) (*ModelHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	handler := NewModelHandler(svc)
	r := gin.New()
	return handler, r
}

func TestModelHandler_RegisterModel_Success(t *testing.T) {
	mockSvc := new(mockModelService)
	handler, r := setupModelHandler(mockSvc)

	r.POST("/models", handler.RegisterModel)

	accuracy := 0.95
	model := &entity.ModelVersion{ID: "model-001", ModelName: "solar_v2", Version: "1.0.0", Accuracy: &accuracy}
	mockSvc.On("RegisterModel", mock.Anything, "solar_v2", "1.0.0", "/path/to/model", 0.95).Return(model, nil)

	body := map[string]interface{}{
		"model_name":    "solar_v2",
		"version":       "1.0.0",
		"artifact_path": "/path/to/model",
		"accuracy":      0.95,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/models", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestModelHandler_RegisterModel_InvalidJSON(t *testing.T) {
	mockSvc := new(mockModelService)
	handler, r := setupModelHandler(mockSvc)

	r.POST("/models", handler.RegisterModel)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/models", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestModelHandler_RegisterModel_Error(t *testing.T) {
	mockSvc := new(mockModelService)
	handler, r := setupModelHandler(mockSvc)

	r.POST("/models", handler.RegisterModel)

	mockSvc.On("RegisterModel", mock.Anything, "solar_v2", "1.0.0", "/path/to/model", 0.95).Return(nil, assert.AnError)

	body := map[string]interface{}{
		"model_name":    "solar_v2",
		"version":       "1.0.0",
		"artifact_path": "/path/to/model",
		"accuracy":      0.95,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/models", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestModelHandler_GetModel_Success(t *testing.T) {
	mockSvc := new(mockModelService)
	handler, r := setupModelHandler(mockSvc)

	r.GET("/models/:id", handler.GetModel)

	accuracy := 0.95
	model := &entity.ModelVersion{ID: "model-001", ModelName: "solar_v2", Accuracy: &accuracy}
	mockSvc.On("GetModel", mock.Anything, "model-001").Return(model, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/models/model-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestModelHandler_GetModel_NotFound(t *testing.T) {
	mockSvc := new(mockModelService)
	handler, r := setupModelHandler(mockSvc)

	r.GET("/models/:id", handler.GetModel)

	mockSvc.On("GetModel", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/models/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestModelHandler_ListModels_Success(t *testing.T) {
	mockSvc := new(mockModelService)
	handler, r := setupModelHandler(mockSvc)

	r.GET("/models", handler.ListModels)

	mockSvc.On("ListModels", mock.Anything, "").Return([]*entity.ModelVersion{}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/models", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestModelHandler_ListModels_Error(t *testing.T) {
	mockSvc := new(mockModelService)
	handler, r := setupModelHandler(mockSvc)

	r.GET("/models", handler.ListModels)

	mockSvc.On("ListModels", mock.Anything, "").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/models", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestModelHandler_GetProductionModel_Success(t *testing.T) {
	mockSvc := new(mockModelService)
	handler, r := setupModelHandler(mockSvc)

	r.GET("/models/production/:model_name", handler.GetProductionModel)

	accuracy := 0.95
	model := &entity.ModelVersion{ID: "model-001", ModelName: "solar_v2", Status: entity.ModelStatusProd, Accuracy: &accuracy}
	mockSvc.On("GetProductionModel", mock.Anything, "solar_v2").Return(model, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/models/production/solar_v2", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestModelHandler_GetProductionModel_NotFound(t *testing.T) {
	mockSvc := new(mockModelService)
	handler, r := setupModelHandler(mockSvc)

	r.GET("/models/production/:model_name", handler.GetProductionModel)

	mockSvc.On("GetProductionModel", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/models/production/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestModelHandler_PromoteToProduction_Success(t *testing.T) {
	mockSvc := new(mockModelService)
	handler, r := setupModelHandler(mockSvc)

	r.PUT("/models/:id/promote", handler.PromoteToProduction)

	mockSvc.On("PromoteToProduction", mock.Anything, "model-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/models/model-001/promote", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestModelHandler_PromoteToProduction_Error(t *testing.T) {
	mockSvc := new(mockModelService)
	handler, r := setupModelHandler(mockSvc)

	r.PUT("/models/:id/promote", handler.PromoteToProduction)

	mockSvc.On("PromoteToProduction", mock.Anything, "model-001").Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/models/model-001/promote", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestModelHandler_RetireModel_Success(t *testing.T) {
	mockSvc := new(mockModelService)
	handler, r := setupModelHandler(mockSvc)

	r.PUT("/models/:id/retire", handler.RetireModel)

	mockSvc.On("RetireModel", mock.Anything, "model-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/models/model-001/retire", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestModelHandler_RetireModel_Error(t *testing.T) {
	mockSvc := new(mockModelService)
	handler, r := setupModelHandler(mockSvc)

	r.PUT("/models/:id/retire", handler.RetireModel)

	mockSvc.On("RetireModel", mock.Anything, "model-001").Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/models/model-001/retire", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestModelHandler_NewModelHandler(t *testing.T) {
	mockSvc := new(mockModelService)
	handler := NewModelHandler(mockSvc)
	assert.NotNil(t, handler)
}
