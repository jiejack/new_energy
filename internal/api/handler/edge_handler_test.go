package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

type mockEdgeService struct {
	registerNodeFunc    func(ctx context.Context, name, stationID, ipAddress string) (*entity.EdgeNode, error)
	getNodeFunc         func(ctx context.Context, id string) (*entity.EdgeNode, error)
	listNodesFunc       func(ctx context.Context, stationID *string, status *entity.EdgeNodeStatus) ([]*entity.EdgeNode, error)
	updateNodeConfigFunc func(ctx context.Context, id string, configVersion string) error
	deployModelFunc     func(ctx context.Context, nodeID string, modelName, version string) error
	getNodeStatusFunc   func(ctx context.Context, id string) (*entity.EdgeNode, error)
	triggerSyncFunc     func(ctx context.Context, id string) error
}

func (m *mockEdgeService) RegisterNode(ctx context.Context, name, stationID, ipAddress string) (*entity.EdgeNode, error) {
	if m.registerNodeFunc != nil {
		return m.registerNodeFunc(ctx, name, stationID, ipAddress)
	}
	return &entity.EdgeNode{
		ID:        "edge-001",
		Name:      name,
		StationID: stationID,
		IPAddress: ipAddress,
		Status:    entity.EdgeNodeOffline,
	}, nil
}

func (m *mockEdgeService) GetNode(ctx context.Context, id string) (*entity.EdgeNode, error) {
	if m.getNodeFunc != nil {
		return m.getNodeFunc(ctx, id)
	}
	if id == "not-found" {
		return nil, fmt.Errorf("not found")
	}
	return &entity.EdgeNode{
		ID:        id,
		Name:      "test-node",
		StationID: "station-001",
		IPAddress: "192.168.1.100",
		Status:    entity.EdgeNodeOnline,
	}, nil
}

func (m *mockEdgeService) ListNodes(ctx context.Context, stationID *string, status *entity.EdgeNodeStatus) ([]*entity.EdgeNode, error) {
	if m.listNodesFunc != nil {
		return m.listNodesFunc(ctx, stationID, status)
	}
	return []*entity.EdgeNode{
		{
			ID:        "edge-001",
			Name:      "node-1",
			StationID: "station-001",
			IPAddress: "192.168.1.100",
			Status:    entity.EdgeNodeOnline,
		},
	}, nil
}

func (m *mockEdgeService) UpdateNodeConfig(ctx context.Context, id string, configVersion string) error {
	if m.updateNodeConfigFunc != nil {
		return m.updateNodeConfigFunc(ctx, id, configVersion)
	}
	return nil
}

func (m *mockEdgeService) DeployModel(ctx context.Context, nodeID string, modelName, version string) error {
	if m.deployModelFunc != nil {
		return m.deployModelFunc(ctx, nodeID, modelName, version)
	}
	return nil
}

func (m *mockEdgeService) GetNodeStatus(ctx context.Context, id string) (*entity.EdgeNode, error) {
	if m.getNodeStatusFunc != nil {
		return m.getNodeStatusFunc(ctx, id)
	}
	if id == "not-found" {
		return nil, fmt.Errorf("not found")
	}
	return &entity.EdgeNode{
		ID:        id,
		Name:      "test-node",
		StationID: "station-001",
		IPAddress: "192.168.1.100",
		Status:    entity.EdgeNodeOnline,
	}, nil
}

func (m *mockEdgeService) TriggerSync(ctx context.Context, id string) error {
	if m.triggerSyncFunc != nil {
		return m.triggerSyncFunc(ctx, id)
	}
	return nil
}

func setupEdgeRouter(h *EdgeHandler) *gin.Engine {
	r := gin.New()
	r.GET("/api/v1/edge/nodes", h.ListNodes)
	r.GET("/api/v1/edge/nodes/:id", h.GetNode)
	r.POST("/api/v1/edge/nodes", h.RegisterNode)
	r.PUT("/api/v1/edge/nodes/:id/config", h.UpdateNodeConfig)
	r.POST("/api/v1/edge/nodes/:id/deploy", h.DeployModel)
	r.GET("/api/v1/edge/nodes/:id/status", h.GetNodeStatus)
	r.POST("/api/v1/edge/nodes/:id/sync", h.TriggerSync)
	return r
}

func TestEdgeHandler_RegisterNode_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockEdgeService{}
	handler := NewEdgeHandler(mockSvc)
	router := setupEdgeRouter(handler)

	body := `{"name":"test-node","station_id":"station-001","ip_address":"192.168.1.100"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/edge/nodes", strings.NewReader(body))
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

func TestEdgeHandler_RegisterNode_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockEdgeService{}
	handler := NewEdgeHandler(mockSvc)
	router := setupEdgeRouter(handler)

	body := `{"name":"test-node"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/edge/nodes", strings.NewReader(body))
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

func TestEdgeHandler_ListNodes_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockEdgeService{}
	handler := NewEdgeHandler(mockSvc)
	router := setupEdgeRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/edge/nodes?station_id=station-001", nil)
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

func TestEdgeHandler_GetNode_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockEdgeService{}
	handler := NewEdgeHandler(mockSvc)
	router := setupEdgeRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/edge/nodes/edge-001", nil)
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

func TestEdgeHandler_GetNode_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockEdgeService{}
	handler := NewEdgeHandler(mockSvc)
	router := setupEdgeRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/edge/nodes/not-found", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestEdgeHandler_UpdateNodeConfig_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockEdgeService{}
	handler := NewEdgeHandler(mockSvc)
	router := setupEdgeRouter(handler)

	body := `{"config_version":"2.0.0"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/edge/nodes/edge-001/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestEdgeHandler_UpdateNodeConfig_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockEdgeService{}
	handler := NewEdgeHandler(mockSvc)
	router := setupEdgeRouter(handler)

	body := `{}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/edge/nodes/edge-001/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestEdgeHandler_DeployModel_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockEdgeService{}
	handler := NewEdgeHandler(mockSvc)
	router := setupEdgeRouter(handler)

	body := `{"model_name":"solar-forecast","version":"1.0.0"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/edge/nodes/edge-001/deploy", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestEdgeHandler_DeployModel_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockEdgeService{}
	handler := NewEdgeHandler(mockSvc)
	router := setupEdgeRouter(handler)

	body := `{"model_name":"solar-forecast"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/edge/nodes/edge-001/deploy", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestEdgeHandler_GetNodeStatus_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockEdgeService{}
	handler := NewEdgeHandler(mockSvc)
	router := setupEdgeRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/edge/nodes/edge-001/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestEdgeHandler_GetNodeStatus_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockEdgeService{}
	handler := NewEdgeHandler(mockSvc)
	router := setupEdgeRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/edge/nodes/not-found/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestEdgeHandler_TriggerSync_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockEdgeService{}
	handler := NewEdgeHandler(mockSvc)
	router := setupEdgeRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/edge/nodes/edge-001/sync", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestEdgeHandler_TriggerSync_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockEdgeService{
		triggerSyncFunc: func(ctx context.Context, id string) error {
			return fmt.Errorf("sync failed")
		},
	}
	handler := NewEdgeHandler(mockSvc)
	router := setupEdgeRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/edge/nodes/edge-001/sync", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
