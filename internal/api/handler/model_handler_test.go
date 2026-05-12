package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

type mockModelService struct {
	registerModelFunc      func(ctx context.Context, modelName, version, artifactPath string, accuracy float64) (*entity.ModelVersion, error)
	getModelFunc           func(ctx context.Context, id string) (*entity.ModelVersion, error)
	getProductionModelFunc func(ctx context.Context, modelName string) (*entity.ModelVersion, error)
	listModelsFunc         func(ctx context.Context, modelName string) ([]*entity.ModelVersion, error)
	promoteToProductionFunc func(ctx context.Context, id string) error
	retireModelFunc        func(ctx context.Context, id string) error
}

func (m *mockModelService) RegisterModel(ctx context.Context, modelName, version, artifactPath string, accuracy float64) (*entity.ModelVersion, error) {
	if m.registerModelFunc != nil {
		return m.registerModelFunc(ctx, modelName, version, artifactPath, accuracy)
	}
	return &entity.ModelVersion{
		ID:           "model-001",
		ModelName:    modelName,
		Version:      version,
		Status:       entity.ModelStatusStaging,
		Accuracy:     &accuracy,
		ArtifactPath: artifactPath,
	}, nil
}

func (m *mockModelService) GetModel(ctx context.Context, id string) (*entity.ModelVersion, error) {
	if m.getModelFunc != nil {
		return m.getModelFunc(ctx, id)
	}
	if id == "not-found" {
		return nil, fmt.Errorf("not found")
	}
	accuracy := 0.92
	return &entity.ModelVersion{
		ID:        id,
		ModelName: "solar-forecast",
		Version:   "1.0.0",
		Status:    entity.ModelStatusProd,
		Accuracy:  &accuracy,
	}, nil
}

func (m *mockModelService) GetProductionModel(ctx context.Context, modelName string) (*entity.ModelVersion, error) {
	if m.getProductionModelFunc != nil {
		return m.getProductionModelFunc(ctx, modelName)
	}
	if modelName == "not-found" {
		return nil, fmt.Errorf("not found")
	}
	accuracy := 0.92
	return &entity.ModelVersion{
		ID:        "model-prod-001",
		ModelName: modelName,
		Version:   "1.0.0",
		Status:    entity.ModelStatusProd,
		Accuracy:  &accuracy,
	}, nil
}

func (m *mockModelService) ListModels(ctx context.Context, modelName string) ([]*entity.ModelVersion, error) {
	if m.listModelsFunc != nil {
		return m.listModelsFunc(ctx, modelName)
	}
	accuracy := 0.92
	return []*entity.ModelVersion{
		{
			ID:        "model-001",
			ModelName: "solar-forecast",
			Version:   "1.0.0",
			Status:    entity.ModelStatusProd,
			Accuracy:  &accuracy,
		},
	}, nil
}

func (m *mockModelService) PromoteToProduction(ctx context.Context, id string) error {
	if m.promoteToProductionFunc != nil {
		return m.promoteToProductionFunc(ctx, id)
	}
	return nil
}

func (m *mockModelService) RetireModel(ctx context.Context, id string) error {
	if m.retireModelFunc != nil {
		return m.retireModelFunc(ctx, id)
	}
	return nil
}

func setupModelRouter(h *ModelHandler) *gin.Engine {
	r := gin.New()
	r.POST("/api/v1/models", h.RegisterModel)
	r.GET("/api/v1/models/:id", h.GetModel)
	r.GET("/api/v1/models", h.ListModels)
	r.GET("/api/v1/models/production/:model_name", h.GetProductionModel)
	r.PUT("/api/v1/models/:id/promote", h.PromoteToProduction)
	r.PUT("/api/v1/models/:id/retire", h.RetireModel)
	return r
}

func TestModelHandler_RegisterModel_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockModelService{}
	handler := NewModelHandler(mockSvc)
	router := setupModelRouter(handler)

	body := `{"model_name":"solar-forecast","version":"1.0.0","artifact_path":"/models/solar-v1","accuracy":0.92}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/models", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp dto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
}

func TestModelHandler_RegisterModel_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockModelService{}
	handler := NewModelHandler(mockSvc)
	router := setupModelRouter(handler)

	body := `{"model_name":"solar-forecast"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/models", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != 400 {
		t.Fatalf("expected error code 400, got %d", resp.Code)
	}
}

func TestModelHandler_GetModel_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockModelService{}
	handler := NewModelHandler(mockSvc)
	router := setupModelRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/models/model-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp dto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
}

func TestModelHandler_GetModel_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockModelService{}
	handler := NewModelHandler(mockSvc)
	router := setupModelRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/models/not-found", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestModelHandler_ListModels_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockModelService{}
	handler := NewModelHandler(mockSvc)
	router := setupModelRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/models?model_name=solar-forecast", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp dto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
}

func TestModelHandler_GetProductionModel_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockModelService{}
	handler := NewModelHandler(mockSvc)
	router := setupModelRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/models/production/solar-forecast", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp dto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
}

func TestModelHandler_GetProductionModel_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockModelService{}
	handler := NewModelHandler(mockSvc)
	router := setupModelRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/models/production/not-found", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestModelHandler_PromoteToProduction_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockModelService{}
	handler := NewModelHandler(mockSvc)
	router := setupModelRouter(handler)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/models/model-001/promote", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestModelHandler_PromoteToProduction_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockModelService{
		promoteToProductionFunc: func(ctx context.Context, id string) error {
			return fmt.Errorf("promote failed")
		},
	}
	handler := NewModelHandler(mockSvc)
	router := setupModelRouter(handler)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/models/model-001/promote", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestModelHandler_RetireModel_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockModelService{}
	handler := NewModelHandler(mockSvc)
	router := setupModelRouter(handler)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/models/model-001/retire", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestModelHandler_RetireModel_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockModelService{
		retireModelFunc: func(ctx context.Context, id string) error {
			return fmt.Errorf("retire failed")
		},
	}
	handler := NewModelHandler(mockSvc)
	router := setupModelRouter(handler)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/models/model-001/retire", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func init() {
	_ = time.Now
}
