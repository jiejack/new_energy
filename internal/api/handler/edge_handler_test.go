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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockEdgeService struct {
	mock.Mock
}

func (m *mockEdgeService) RegisterNode(ctx context.Context, name, stationID, ipAddress string) (*entity.EdgeNode, error) {
	args := m.Called(ctx, name, stationID, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.EdgeNode), args.Error(1)
}
func (m *mockEdgeService) GetNode(ctx context.Context, id string) (*entity.EdgeNode, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.EdgeNode), args.Error(1)
}
func (m *mockEdgeService) ListNodes(ctx context.Context, stationID *string, status *entity.EdgeNodeStatus) ([]*entity.EdgeNode, error) {
	args := m.Called(ctx, stationID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.EdgeNode), args.Error(1)
}
func (m *mockEdgeService) UpdateNodeConfig(ctx context.Context, id string, configVersion string) error {
	args := m.Called(ctx, id, configVersion)
	return args.Error(0)
}
func (m *mockEdgeService) DeployModel(ctx context.Context, nodeID string, modelName, version string) error {
	args := m.Called(ctx, nodeID, modelName, version)
	return args.Error(0)
}
func (m *mockEdgeService) GetNodeStatus(ctx context.Context, id string) (*entity.EdgeNode, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.EdgeNode), args.Error(1)
}
func (m *mockEdgeService) TriggerSync(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupEdgeHandler(svc service.EdgeService) (*EdgeHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	handler := NewEdgeHandler(svc)
	r := gin.New()
	return handler, r
}

func TestEdgeHandler_ListNodes_Success(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.GET("/edge/nodes", handler.ListNodes)

	mockSvc.On("ListNodes", mock.Anything, (*string)(nil), (*entity.EdgeNodeStatus)(nil)).Return([]*entity.EdgeNode{}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/edge/nodes", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEdgeHandler_ListNodes_WithFilters(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.GET("/edge/nodes", handler.ListNodes)

	stationID := "station-001"
	status := entity.EdgeNodeOnline
	mockSvc.On("ListNodes", mock.Anything, &stationID, &status).Return([]*entity.EdgeNode{}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/edge/nodes?station_id=station-001&status=online", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEdgeHandler_ListNodes_Error(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.GET("/edge/nodes", handler.ListNodes)

	mockSvc.On("ListNodes", mock.Anything, (*string)(nil), (*entity.EdgeNodeStatus)(nil)).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/edge/nodes", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEdgeHandler_GetNode_Success(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.GET("/edge/nodes/:id", handler.GetNode)

	node := &entity.EdgeNode{ID: "node-001", Name: "Edge1"}
	mockSvc.On("GetNode", mock.Anything, "node-001").Return(node, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/edge/nodes/node-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEdgeHandler_GetNode_NotFound(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.GET("/edge/nodes/:id", handler.GetNode)

	mockSvc.On("GetNode", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/edge/nodes/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestEdgeHandler_RegisterNode_Success(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.POST("/edge/nodes", handler.RegisterNode)

	node := &entity.EdgeNode{ID: "node-001", Name: "Edge1", StationID: "station-001", IPAddress: "192.168.1.1"}
	mockSvc.On("RegisterNode", mock.Anything, "Edge1", "station-001", "192.168.1.1").Return(node, nil)

	body := map[string]string{"name": "Edge1", "station_id": "station-001", "ip_address": "192.168.1.1"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/edge/nodes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestEdgeHandler_RegisterNode_InvalidJSON(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.POST("/edge/nodes", handler.RegisterNode)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/edge/nodes", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEdgeHandler_RegisterNode_Error(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.POST("/edge/nodes", handler.RegisterNode)

	mockSvc.On("RegisterNode", mock.Anything, "Edge1", "station-001", "192.168.1.1").Return(nil, assert.AnError)

	body := map[string]string{"name": "Edge1", "station_id": "station-001", "ip_address": "192.168.1.1"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/edge/nodes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEdgeHandler_UpdateNodeConfig_Success(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.PUT("/edge/nodes/:id/config", handler.UpdateNodeConfig)

	mockSvc.On("UpdateNodeConfig", mock.Anything, "node-001", "v2.0").Return(nil)

	body := map[string]string{"config_version": "v2.0"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/edge/nodes/node-001/config", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEdgeHandler_UpdateNodeConfig_InvalidJSON(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.PUT("/edge/nodes/:id/config", handler.UpdateNodeConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/edge/nodes/node-001/config", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEdgeHandler_DeployModel_Success(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.POST("/edge/nodes/:id/deploy", handler.DeployModel)

	mockSvc.On("DeployModel", mock.Anything, "node-001", "solar_v2", "1.0.0").Return(nil)

	body := map[string]string{"model_name": "solar_v2", "version": "1.0.0"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/edge/nodes/node-001/deploy", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEdgeHandler_DeployModel_InvalidJSON(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.POST("/edge/nodes/:id/deploy", handler.DeployModel)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/edge/nodes/node-001/deploy", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEdgeHandler_GetNodeStatus_Success(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.GET("/edge/nodes/:id/status", handler.GetNodeStatus)

	node := &entity.EdgeNode{ID: "node-001", Status: entity.EdgeNodeOnline}
	mockSvc.On("GetNodeStatus", mock.Anything, "node-001").Return(node, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/edge/nodes/node-001/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEdgeHandler_GetNodeStatus_NotFound(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.GET("/edge/nodes/:id/status", handler.GetNodeStatus)

	mockSvc.On("GetNodeStatus", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/edge/nodes/nonexistent/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestEdgeHandler_TriggerSync_Success(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.POST("/edge/nodes/:id/sync", handler.TriggerSync)

	mockSvc.On("TriggerSync", mock.Anything, "node-001").Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/edge/nodes/node-001/sync", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEdgeHandler_TriggerSync_Error(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler, r := setupEdgeHandler(mockSvc)

	r.POST("/edge/nodes/:id/sync", handler.TriggerSync)

	mockSvc.On("TriggerSync", mock.Anything, "node-001").Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/edge/nodes/node-001/sync", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEdgeHandler_NewEdgeHandler(t *testing.T) {
	mockSvc := new(mockEdgeService)
	handler := NewEdgeHandler(mockSvc)
	assert.NotNil(t, handler)
}
