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

// MockAuditOperationLogRepository 操作日志仓储Mock
type MockAuditOperationLogRepository struct {
	mock.Mock
}

func (m *MockAuditOperationLogRepository) Create(ctx context.Context, log *entity.OperationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditOperationLogRepository) GetByID(ctx context.Context, id string) (*entity.OperationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OperationLog), args.Error(1)
}

func (m *MockAuditOperationLogRepository) List(ctx context.Context, query *repository.OperationLogQuery) ([]*entity.OperationLog, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.OperationLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockAuditOperationLogRepository) DeleteBefore(ctx context.Context, before int64) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

func TestNewAuditService(t *testing.T) {
	mockRepo := new(MockAuditOperationLogRepository)
	service := NewAuditService(mockRepo)
	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.logRepo)
}

func TestAuditService_LogOperation(t *testing.T) {
	tests := []struct {
		name    string
		req     *LogOperationRequest
		mockErr error
		wantErr bool
	}{
		{
			name: "成功记录操作日志",
			req: &LogOperationRequest{
				UserID:       "user-1",
				Username:     "testuser",
				Action:       entity.ActionLogin,
				ResourceType: "",
				ResourceID:   "",
				Details:      entity.Details{"success": true},
				IPAddress:    "192.168.1.1",
				UserAgent:    "Mozilla/5.0",
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "记录带资源的操作日志",
			req: &LogOperationRequest{
				UserID:       "user-1",
				Username:     "testuser",
				Action:       entity.ActionCreateUser,
				ResourceType: entity.ResourceUser,
				ResourceID:   "user-2",
				Details:      entity.Details{"new_user": "test"},
				IPAddress:    "192.168.1.1",
				UserAgent:    "Mozilla/5.0",
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "仓储错误",
			req: &LogOperationRequest{
				UserID:   "user-1",
				Username: "testuser",
				Action:   entity.ActionLogin,
			},
			mockErr: errors.New("database error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuditOperationLogRepository)
			service := NewAuditService(mockRepo)

			mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).
				Return(tt.mockErr)

			err := service.LogOperation(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.mockErr, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuditService_GetLog(t *testing.T) {
	tests := []struct {
		name    string
		logID   string
		mockLog *entity.OperationLog
		mockErr error
		wantErr bool
	}{
		{
			name:  "成功获取日志",
			logID: "log-1",
			mockLog: &entity.OperationLog{
				ID:       "log-1",
				UserID:   "user-1",
				Username: "testuser",
				Action:   entity.ActionLogin,
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "日志不存在",
			logID:   "non-existent",
			mockLog: nil,
			mockErr: errors.New("not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuditOperationLogRepository)
			service := NewAuditService(mockRepo)

			mockRepo.On("GetByID", mock.Anything, tt.logID).
				Return(tt.mockLog, tt.mockErr)

			log, err := service.GetLog(context.Background(), tt.logID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, log)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, log)
				assert.Equal(t, tt.logID, log.ID)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuditService_ListLogs(t *testing.T) {
	tests := []struct {
		name      string
		userID    *string
		action    *string
		startTime int64
		endTime   int64
		page      int
		pageSize  int
		mockLogs  []*entity.OperationLog
		mockTotal int64
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "成功获取日志列表",
			userID:    strPtr("user-1"),
			action:    strPtr(entity.ActionLogin),
			startTime: 1000000,
			endTime:   2000000,
			page:      1,
			pageSize:  10,
			mockLogs: []*entity.OperationLog{
				{ID: "log-1", UserID: "user-1", Action: entity.ActionLogin},
				{ID: "log-2", UserID: "user-1", Action: entity.ActionLogin},
			},
			mockTotal: 2,
			mockErr:   nil,
			wantErr:   false,
		},
		{
			name:      "空列表",
			userID:    nil,
			action:    nil,
			startTime: 0,
			endTime:   0,
			page:      1,
			pageSize:  10,
			mockLogs:  []*entity.OperationLog{},
			mockTotal: 0,
			mockErr:   nil,
			wantErr:   false,
		},
		{
			name:      "仓储错误",
			userID:    nil,
			action:    nil,
			startTime: 0,
			endTime:   0,
			page:      1,
			pageSize:  10,
			mockLogs:  nil,
			mockTotal: 0,
			mockErr:   errors.New("database error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuditOperationLogRepository)
			service := NewAuditService(mockRepo)

			expectedQuery := &repository.OperationLogQuery{
				Page:      tt.page,
				PageSize:  tt.pageSize,
				StartTime: tt.startTime,
				EndTime:   tt.endTime,
			}
			if tt.userID != nil {
				expectedQuery.UserID = *tt.userID
			}
			if tt.action != nil {
				expectedQuery.Action = *tt.action
			}

			mockRepo.On("List", mock.Anything, mock.MatchedBy(func(q *repository.OperationLogQuery) bool {
				return q.Page == expectedQuery.Page &&
					q.PageSize == expectedQuery.PageSize &&
					q.UserID == expectedQuery.UserID &&
					q.Action == expectedQuery.Action &&
					q.StartTime == expectedQuery.StartTime &&
					q.EndTime == expectedQuery.EndTime
			})).Return(tt.mockLogs, tt.mockTotal, tt.mockErr)

			resp, err := service.ListLogs(context.Background(), tt.userID, tt.action, tt.startTime, tt.endTime, tt.page, tt.pageSize)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp.List, len(tt.mockLogs))
				assert.Equal(t, tt.mockTotal, resp.Total)
				assert.Equal(t, tt.page, resp.Page)
				assert.Equal(t, tt.pageSize, resp.PageSize)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuditService_GetUserLogs(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		page      int
		pageSize  int
		mockLogs  []*entity.OperationLog
		mockTotal int64
		mockErr   error
		wantErr   bool
	}{
		{
			name:     "成功获取用户日志",
			userID:   "user-1",
			page:     1,
			pageSize: 10,
			mockLogs: []*entity.OperationLog{
				{ID: "log-1", UserID: "user-1"},
				{ID: "log-2", UserID: "user-1"},
			},
			mockTotal: 2,
			mockErr:   nil,
			wantErr:   false,
		},
		{
			name:      "用户无日志",
			userID:    "user-2",
			page:      1,
			pageSize:  10,
			mockLogs:  []*entity.OperationLog{},
			mockTotal: 0,
			mockErr:   nil,
			wantErr:   false,
		},
		{
			name:      "仓储错误",
			userID:    "user-1",
			page:      1,
			pageSize:  10,
			mockLogs:  nil,
			mockTotal: 0,
			mockErr:   errors.New("database error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuditOperationLogRepository)
			service := NewAuditService(mockRepo)

			expectedQuery := &repository.OperationLogQuery{
				Page:     tt.page,
				PageSize: tt.pageSize,
				UserID:   tt.userID,
			}

			mockRepo.On("List", mock.Anything, mock.MatchedBy(func(q *repository.OperationLogQuery) bool {
				return q.Page == expectedQuery.Page &&
					q.PageSize == expectedQuery.PageSize &&
					q.UserID == expectedQuery.UserID
			})).Return(tt.mockLogs, tt.mockTotal, tt.mockErr)

			resp, err := service.GetUserLogs(context.Background(), tt.userID, tt.page, tt.pageSize)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp.List, len(tt.mockLogs))
				assert.Equal(t, tt.mockTotal, resp.Total)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuditService_LogLogin(t *testing.T) {
	mockRepo := new(MockAuditOperationLogRepository)
	service := NewAuditService(mockRepo)

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).
		Return(nil)

	err := service.LogLogin(context.Background(), "user-1", "testuser", true, "192.168.1.1", "Mozilla/5.0")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogLogout(t *testing.T) {
	mockRepo := new(MockAuditOperationLogRepository)
	service := NewAuditService(mockRepo)

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).
		Return(nil)

	err := service.LogLogout(context.Background(), "user-1", "testuser", "192.168.1.1", "Mozilla/5.0")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogCreateUser(t *testing.T) {
	mockRepo := new(MockAuditOperationLogRepository)
	service := NewAuditService(mockRepo)

	newUser := &entity.User{
		ID:       "user-2",
		Username: "newuser",
		Email:    "new@example.com",
		RealName: "New User",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).
		Return(nil)

	err := service.LogCreateUser(context.Background(), "admin", "adminuser", newUser, "192.168.1.1", "Mozilla/5.0")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogUpdateUser(t *testing.T) {
	mockRepo := new(MockAuditOperationLogRepository)
	service := NewAuditService(mockRepo)

	updatedUser := &entity.User{
		ID:       "user-2",
		Username: "updateduser",
	}
	changes := map[string]interface{}{"email": "new@example.com"}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).
		Return(nil)

	err := service.LogUpdateUser(context.Background(), "admin", "adminuser", updatedUser, changes, "192.168.1.1", "Mozilla/5.0")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogDeleteUser(t *testing.T) {
	mockRepo := new(MockAuditOperationLogRepository)
	service := NewAuditService(mockRepo)

	deletedUser := &entity.User{
		ID:       "user-2",
		Username: "deleteduser",
		Email:    "deleted@example.com",
		RealName: "Deleted User",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).
		Return(nil)

	err := service.LogDeleteUser(context.Background(), "admin", "adminuser", deletedUser, "192.168.1.1", "Mozilla/5.0")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogChangePassword(t *testing.T) {
	mockRepo := new(MockAuditOperationLogRepository)
	service := NewAuditService(mockRepo)

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).
		Return(nil)

	err := service.LogChangePassword(context.Background(), "admin", "adminuser", "user-2", "192.168.1.1", "Mozilla/5.0")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogAssignRole(t *testing.T) {
	mockRepo := new(MockAuditOperationLogRepository)
	service := NewAuditService(mockRepo)

	roleIDs := []string{"role-1", "role-2"}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).
		Return(nil)

	err := service.LogAssignRole(context.Background(), "admin", "adminuser", "user-2", roleIDs, "192.168.1.1", "Mozilla/5.0")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogCreateRole(t *testing.T) {
	mockRepo := new(MockAuditOperationLogRepository)
	service := NewAuditService(mockRepo)

	newRole := &entity.Role{
		ID:          "role-1",
		Code:        "admin",
		Name:        "Administrator",
		Description: "Admin role",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).
		Return(nil)

	err := service.LogCreateRole(context.Background(), "admin", "adminuser", newRole, "192.168.1.1", "Mozilla/5.0")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuditService_LogAssignPermission(t *testing.T) {
	mockRepo := new(MockAuditOperationLogRepository)
	service := NewAuditService(mockRepo)

	permissionIDs := []string{"perm-1", "perm-2"}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).
		Return(nil)

	err := service.LogAssignPermission(context.Background(), "admin", "adminuser", "role-1", permissionIDs, "192.168.1.1", "Mozilla/5.0")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// 辅助函数
func strPtr(s string) *string {
	return &s
}
