package service

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockOperationLogRepo struct {
	mock.Mock
}

func (m *mockOperationLogRepo) Create(ctx context.Context, log *entity.OperationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *mockOperationLogRepo) GetByID(ctx context.Context, id string) (*entity.OperationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OperationLog), args.Error(1)
}

func (m *mockOperationLogRepo) List(ctx context.Context, query *repository.OperationLogQuery) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func (m *mockOperationLogRepo) DeleteBefore(ctx context.Context, before int64) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

func TestOperationLogService_CreateLog(t *testing.T) {
	repo := new(mockOperationLogRepo)
	svc := NewOperationLogService(repo)
	ctx := context.Background()

	repo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	req := &CreateOperationLogRequest{
		UserID:       "user-001",
		Username:     "admin",
		Action:       entity.ActionLogin,
		ResourceType: "system",
		ResourceID:   "sys-001",
		Details:      map[string]interface{}{"ip": "192.168.1.1"},
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
	}

	resp, err := svc.CreateLog(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "user-001", resp.UserID)
	assert.Equal(t, "admin", resp.Username)
	assert.Equal(t, entity.ActionLogin, resp.Action)
	assert.Equal(t, "system", resp.ResourceType)
	assert.Equal(t, "sys-001", resp.ResourceID)
	assert.Equal(t, "192.168.1.1", resp.IPAddress)
	assert.Equal(t, "Mozilla/5.0", resp.UserAgent)
	repo.AssertExpectations(t)
}

func TestOperationLogService_CreateLog_WithoutDetails(t *testing.T) {
	repo := new(mockOperationLogRepo)
	svc := NewOperationLogService(repo)
	ctx := context.Background()

	repo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	req := &CreateOperationLogRequest{
		UserID:   "user-001",
		Username: "admin",
		Action:   entity.ActionCreateUser,
	}

	resp, err := svc.CreateLog(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, entity.ActionCreateUser, resp.Action)
}

func TestOperationLogService_CreateLog_RepoError(t *testing.T) {
	repo := new(mockOperationLogRepo)
	svc := NewOperationLogService(repo)
	ctx := context.Background()

	repo.On("Create", ctx, mock.AnythingOfType("*entity.OperationLog")).Return(errors.New("db error"))

	req := &CreateOperationLogRequest{
		UserID:   "user-001",
		Username: "admin",
		Action:   entity.ActionLogin,
	}

	_, err := svc.CreateLog(ctx, req)
	assert.Error(t, err)
}

func TestOperationLogService_GetLog(t *testing.T) {
	repo := new(mockOperationLogRepo)
	svc := NewOperationLogService(repo)
	ctx := context.Background()

	log := entity.NewOperationLog("user-001", "admin", entity.ActionLogin)
	log.SetRequestInfo("192.168.1.1", "Mozilla/5.0")
	repo.On("GetByID", ctx, "log-001").Return(log, nil)

	resp, err := svc.GetLog(ctx, "log-001")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "user-001", resp.UserID)
	assert.Equal(t, entity.ActionLogin, resp.Action)
}

func TestOperationLogService_GetLog_NotFound(t *testing.T) {
	repo := new(mockOperationLogRepo)
	svc := NewOperationLogService(repo)
	ctx := context.Background()

	repo.On("GetByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	_, err := svc.GetLog(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestOperationLogService_ListLogs(t *testing.T) {
	repo := new(mockOperationLogRepo)
	svc := NewOperationLogService(repo)
	ctx := context.Background()

	logs := []*entity.OperationLog{
		entity.NewOperationLog("user-001", "admin", entity.ActionLogin),
		entity.NewOperationLog("user-002", "operator", entity.ActionCreateUser),
	}
	query := &repository.OperationLogQuery{Page: 1, PageSize: 10}
	repo.On("List", ctx, query).Return(logs, int64(2), nil)

	resp, err := svc.ListLogs(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)
	assert.Len(t, resp.List, 2)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 10, resp.PageSize)
}

func TestOperationLogService_ListLogs_RepoError(t *testing.T) {
	repo := new(mockOperationLogRepo)
	svc := NewOperationLogService(repo)
	ctx := context.Background()

	query := &repository.OperationLogQuery{Page: 1, PageSize: 10}
	repo.On("List", ctx, query).Return(nil, int64(0), errors.New("db error"))

	_, err := svc.ListLogs(ctx, query)
	assert.Error(t, err)
}

func TestOperationLogService_DeleteOldLogs(t *testing.T) {
	repo := new(mockOperationLogRepo)
	svc := NewOperationLogService(repo)
	ctx := context.Background()

	repo.On("DeleteBefore", ctx, mock.AnythingOfType("int64")).Return(int64(5), nil)

	count, err := svc.DeleteOldLogs(ctx, 30)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestOperationLogService_DeleteOldLogs_Error(t *testing.T) {
	repo := new(mockOperationLogRepo)
	svc := NewOperationLogService(repo)
	ctx := context.Background()

	repo.On("DeleteBefore", ctx, mock.AnythingOfType("int64")).Return(int64(0), errors.New("db error"))

	_, err := svc.DeleteOldLogs(ctx, 30)
	assert.Error(t, err)
}
