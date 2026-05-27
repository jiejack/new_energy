package service

import (
	"context"
	"testing"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockEdgeNodeRepo struct {
	mock.Mock
}

func (m *mockEdgeNodeRepo) Create(ctx context.Context, node *entity.EdgeNode) error {
	args := m.Called(ctx, node)
	return args.Error(0)
}
func (m *mockEdgeNodeRepo) Update(ctx context.Context, node *entity.EdgeNode) error {
	args := m.Called(ctx, node)
	return args.Error(0)
}
func (m *mockEdgeNodeRepo) UpdateStatus(ctx context.Context, id string, status entity.EdgeNodeStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}
func (m *mockEdgeNodeRepo) UpdateHeartbeat(ctx context.Context, id string, cpuUsage, memoryUsage float64) error {
	args := m.Called(ctx, id, cpuUsage, memoryUsage)
	return args.Error(0)
}
func (m *mockEdgeNodeRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockEdgeNodeRepo) GetByID(ctx context.Context, id string) (*entity.EdgeNode, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.EdgeNode), args.Error(1)
}
func (m *mockEdgeNodeRepo) List(ctx context.Context, stationID *string, status *entity.EdgeNodeStatus) ([]*entity.EdgeNode, error) {
	args := m.Called(ctx, stationID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.EdgeNode), args.Error(1)
}

func TestEdgeService_RegisterNode_Success(t *testing.T) {
	repo := new(mockEdgeNodeRepo)
	svc := NewEdgeService(repo)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.EdgeNode")).Return(nil)
	node, err := svc.RegisterNode(context.Background(), "Edge1", "station-001", "192.168.1.1")
	assert.NoError(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, "Edge1", node.Name)
	assert.Equal(t, entity.EdgeNodeOffline, node.Status)
}

func TestEdgeService_RegisterNode_Error(t *testing.T) {
	repo := new(mockEdgeNodeRepo)
	svc := NewEdgeService(repo)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.EdgeNode")).Return(assert.AnError)
	node, err := svc.RegisterNode(context.Background(), "Edge1", "station-001", "192.168.1.1")
	assert.Error(t, err)
	assert.Nil(t, node)
}

func TestEdgeService_GetNode_Success(t *testing.T) {
	repo := new(mockEdgeNodeRepo)
	svc := NewEdgeService(repo)
	node := &entity.EdgeNode{ID: "node-001", Name: "Edge1"}
	repo.On("GetByID", mock.Anything, "node-001").Return(node, nil)
	result, err := svc.GetNode(context.Background(), "node-001")
	assert.NoError(t, err)
	assert.Equal(t, "Edge1", result.Name)
}

func TestEdgeService_ListNodes_Success(t *testing.T) {
	repo := new(mockEdgeNodeRepo)
	svc := NewEdgeService(repo)
	nodes := []*entity.EdgeNode{{ID: "node-001", Name: "Edge1"}}
	repo.On("List", mock.Anything, (*string)(nil), (*entity.EdgeNodeStatus)(nil)).Return(nodes, nil)
	result, err := svc.ListNodes(context.Background(), nil, nil)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestEdgeService_UpdateNodeConfig_Success(t *testing.T) {
	repo := new(mockEdgeNodeRepo)
	svc := NewEdgeService(repo)
	node := &entity.EdgeNode{ID: "node-001", Name: "Edge1"}
	repo.On("GetByID", mock.Anything, "node-001").Return(node, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*entity.EdgeNode")).Return(nil)
	err := svc.UpdateNodeConfig(context.Background(), "node-001", "v2.0")
	assert.NoError(t, err)
}

func TestEdgeService_UpdateNodeConfig_NotFound(t *testing.T) {
	repo := new(mockEdgeNodeRepo)
	svc := NewEdgeService(repo)
	repo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)
	err := svc.UpdateNodeConfig(context.Background(), "nonexistent", "v2.0")
	assert.Error(t, err)
}

func TestEdgeService_DeployModel_Success(t *testing.T) {
	repo := new(mockEdgeNodeRepo)
	svc := NewEdgeService(repo)
	node := &entity.EdgeNode{ID: "node-001", Name: "Edge1", ModelVersions: make(entity.MapJSON)}
	repo.On("GetByID", mock.Anything, "node-001").Return(node, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*entity.EdgeNode")).Return(nil)
	err := svc.DeployModel(context.Background(), "node-001", "solar_v2", "1.0.0")
	assert.NoError(t, err)
}

func TestEdgeService_DeployModel_NilMap(t *testing.T) {
	repo := new(mockEdgeNodeRepo)
	svc := NewEdgeService(repo)
	node := &entity.EdgeNode{ID: "node-001", Name: "Edge1", ModelVersions: nil}
	repo.On("GetByID", mock.Anything, "node-001").Return(node, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*entity.EdgeNode")).Return(nil)
	err := svc.DeployModel(context.Background(), "node-001", "solar_v2", "1.0.0")
	assert.NoError(t, err)
}

func TestEdgeService_GetNodeStatus_Success(t *testing.T) {
	repo := new(mockEdgeNodeRepo)
	svc := NewEdgeService(repo)
	node := &entity.EdgeNode{ID: "node-001", Status: entity.EdgeNodeOnline}
	repo.On("GetByID", mock.Anything, "node-001").Return(node, nil)
	result, err := svc.GetNodeStatus(context.Background(), "node-001")
	assert.NoError(t, err)
	assert.Equal(t, entity.EdgeNodeOnline, result.Status)
}

func TestEdgeService_TriggerSync_Success(t *testing.T) {
	repo := new(mockEdgeNodeRepo)
	svc := NewEdgeService(repo)
	node := &entity.EdgeNode{ID: "node-001"}
	repo.On("GetByID", mock.Anything, "node-001").Return(node, nil)
	err := svc.TriggerSync(context.Background(), "node-001")
	assert.NoError(t, err)
}

func TestCheckNodeHealth_NilHeartbeat(t *testing.T) {
	status := CheckNodeHealth(nil, 5*time.Minute)
	assert.Equal(t, entity.EdgeNodeOffline, status)
}

func TestCheckNodeHealth_Online(t *testing.T) {
	now := time.Now()
	status := CheckNodeHealth(&now, 5*time.Minute)
	assert.Equal(t, entity.EdgeNodeOnline, status)
}

func TestCheckNodeHealth_Offline(t *testing.T) {
	old := time.Now().Add(-10 * time.Minute)
	status := CheckNodeHealth(&old, 5*time.Minute)
	assert.Equal(t, entity.EdgeNodeOffline, status)
}

var _ repository.EdgeNodeRepository = (*mockEdgeNodeRepo)(nil)
