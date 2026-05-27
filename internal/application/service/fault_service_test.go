package service

import (
	"context"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFaultRepo struct {
	mock.Mock
}

func (m *mockFaultRepo) Create(ctx context.Context, result *entity.FaultDetectionResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}
func (m *mockFaultRepo) GetByID(ctx context.Context, id string) (*entity.FaultDetectionResult, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.FaultDetectionResult), args.Error(1)
}
func (m *mockFaultRepo) ListByDevice(ctx context.Context, deviceID string, severity *entity.FaultSeverity, status *entity.FaultDetectionStatus, offset, limit int) ([]*entity.FaultDetectionResult, int64, error) {
	args := m.Called(ctx, deviceID, severity, status, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.FaultDetectionResult), args.Get(1).(int64), args.Error(2)
}
func (m *mockFaultRepo) UpdateStatus(ctx context.Context, id string, status entity.FaultDetectionStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}
func (m *mockFaultRepo) CountBySeverity(ctx context.Context, deviceID *string) (map[entity.FaultSeverity]int64, error) {
	args := m.Called(ctx, deviceID)
	return args.Get(0).(map[entity.FaultSeverity]int64), args.Error(1)
}
func (m *mockFaultRepo) ListByStation(ctx context.Context, stationID string, severity *entity.FaultSeverity, offset, limit int) ([]*entity.FaultDetectionResult, int64, error) {
	args := m.Called(ctx, stationID, severity, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.FaultDetectionResult), args.Get(1).(int64), args.Error(2)
}
func (m *mockFaultRepo) UpdateRootCause(ctx context.Context, id, rootCause string) error {
	args := m.Called(ctx, id, rootCause)
	return args.Error(0)
}
func (m *mockFaultRepo) LinkWorkOrder(ctx context.Context, detectionID, workOrderID string) error {
	args := m.Called(ctx, detectionID, workOrderID)
	return args.Error(0)
}

type mockWorkOrderCreator struct {
	mock.Mock
}

func (m *mockWorkOrderCreator) CreateWorkOrder(ctx context.Context, req *CreateWorkOrderRequest) (*entity.WorkOrder, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkOrder), args.Error(1)
}

func TestFaultService_GetDetections_Success(t *testing.T) {
	faultRepo := new(mockFaultRepo)
	woCreator := new(mockWorkOrderCreator)
	svc := NewFaultService(faultRepo, woCreator)
	faultRepo.On("ListByDevice", mock.Anything, "device-001", (*entity.FaultSeverity)(nil), (*entity.FaultDetectionStatus)(nil), 0, 20).Return([]*entity.FaultDetectionResult{}, int64(0), nil)
	results, total, err := svc.GetDetections(context.Background(), "device-001", nil, nil, 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, results, 0)
}

func TestFaultService_GetDetectionByID_Success(t *testing.T) {
	faultRepo := new(mockFaultRepo)
	woCreator := new(mockWorkOrderCreator)
	svc := NewFaultService(faultRepo, woCreator)
	result := entity.NewFaultDetectionResult("device-001", "temp", entity.FaultSeverityWarning, 0.9, "test", "1.0")
	faultRepo.On("GetByID", mock.Anything, "det-001").Return(result, nil)
	det, err := svc.GetDetectionByID(context.Background(), "det-001")
	assert.NoError(t, err)
	assert.NotNil(t, det)
}

func TestFaultService_UpdateDetectionStatus_Success(t *testing.T) {
	faultRepo := new(mockFaultRepo)
	woCreator := new(mockWorkOrderCreator)
	svc := NewFaultService(faultRepo, woCreator)
	faultRepo.On("UpdateStatus", mock.Anything, "det-001", entity.FaultDetectionStatus(2)).Return(nil)
	err := svc.UpdateDetectionStatus(context.Background(), "det-001", entity.FaultDetectionStatus(2))
	assert.NoError(t, err)
}

func TestFaultService_GetDeviceHealth_Success(t *testing.T) {
	faultRepo := new(mockFaultRepo)
	woCreator := new(mockWorkOrderCreator)
	svc := NewFaultService(faultRepo, woCreator)
	counts := map[entity.FaultSeverity]int64{
		entity.FaultSeverityCritical: 2,
		entity.FaultSeverityWarning:  3,
		entity.FaultSeverityInfo:     1,
	}
	faultRepo.On("CountBySeverity", mock.Anything, mock.AnythingOfType("*string")).Return(counts, nil)
	score, err := svc.GetDeviceHealth(context.Background(), "device-001")
	assert.NoError(t, err)
	assert.Equal(t, 100-2*20-3*5-1, score)
}

func TestFaultService_GetDeviceHealth_ZeroScore(t *testing.T) {
	faultRepo := new(mockFaultRepo)
	woCreator := new(mockWorkOrderCreator)
	svc := NewFaultService(faultRepo, woCreator)
	counts := map[entity.FaultSeverity]int64{
		entity.FaultSeverityCritical: 10,
		entity.FaultSeverityWarning:  20,
	}
	faultRepo.On("CountBySeverity", mock.Anything, mock.AnythingOfType("*string")).Return(counts, nil)
	score, err := svc.GetDeviceHealth(context.Background(), "device-001")
	assert.NoError(t, err)
	assert.Equal(t, 0, score)
}

func TestFaultService_AnalyzeRootCause_Success(t *testing.T) {
	faultRepo := new(mockFaultRepo)
	woCreator := new(mockWorkOrderCreator)
	svc := NewFaultService(faultRepo, woCreator)
	result := entity.NewFaultDetectionResult("device-001", "temp", entity.FaultSeverityWarning, 0.9, "test", "1.0")
	faultRepo.On("GetByID", mock.Anything, "det-001").Return(result, nil)
	det, err := svc.AnalyzeRootCause(context.Background(), "det-001")
	assert.NoError(t, err)
	assert.NotNil(t, det)
}

func TestFaultService_CreateWorkOrderFromDetection_Success(t *testing.T) {
	faultRepo := new(mockFaultRepo)
	woCreator := new(mockWorkOrderCreator)
	svc := NewFaultService(faultRepo, woCreator)
	result := entity.NewFaultDetectionResult("device-001", "temp", entity.FaultSeverityCritical, 0.9, "test", "1.0")
	faultRepo.On("GetByID", mock.Anything, "det-001").Return(result, nil)
	woCreator.On("CreateWorkOrder", mock.Anything, mock.AnythingOfType("*service.CreateWorkOrderRequest")).Return(&entity.WorkOrder{ID: "wo-001"}, nil)
	faultRepo.On("LinkWorkOrder", mock.Anything, "det-001", "wo-001").Return(nil)
	woID, err := svc.CreateWorkOrderFromDetection(context.Background(), "det-001")
	assert.NoError(t, err)
	assert.Equal(t, "wo-001", woID)
}

func TestFaultService_CreateWorkOrderFromDetection_InfoLevel(t *testing.T) {
	faultRepo := new(mockFaultRepo)
	woCreator := new(mockWorkOrderCreator)
	svc := NewFaultService(faultRepo, woCreator)
	result := entity.NewFaultDetectionResult("device-001", "temp", entity.FaultSeverityInfo, 0.9, "test", "1.0")
	faultRepo.On("GetByID", mock.Anything, "det-001").Return(result, nil)
	woID, err := svc.CreateWorkOrderFromDetection(context.Background(), "det-001")
	assert.Error(t, err)
	assert.Equal(t, "", woID)
}

func TestFaultService_CreateWorkOrderFromDetection_NotFound(t *testing.T) {
	faultRepo := new(mockFaultRepo)
	woCreator := new(mockWorkOrderCreator)
	svc := NewFaultService(faultRepo, woCreator)
	faultRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, assert.AnError)
	woID, err := svc.CreateWorkOrderFromDetection(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, "", woID)
}

var _ repository.FaultDetectionResultRepository = (*mockFaultRepo)(nil)
