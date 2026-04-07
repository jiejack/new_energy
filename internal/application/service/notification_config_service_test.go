package service

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNotificationConfigRepository 通知配置仓储Mock
type MockNotificationConfigRepository struct {
	mock.Mock
}

func (m *MockNotificationConfigRepository) Create(ctx context.Context, config *entity.NotificationConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockNotificationConfigRepository) Update(ctx context.Context, config *entity.NotificationConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockNotificationConfigRepository) GetByType(ctx context.Context, notifType entity.NotificationType) (*entity.NotificationConfig, error) {
	args := m.Called(ctx, notifType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.NotificationConfig), args.Error(1)
}

func (m *MockNotificationConfigRepository) GetAll(ctx context.Context) ([]*entity.NotificationConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.NotificationConfig), args.Error(1)
}

func TestNewNotificationConfigService(t *testing.T) {
	mockRepo := new(MockNotificationConfigRepository)
	service := NewNotificationConfigService(mockRepo)
	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.configRepo)
}

func TestNotificationConfigService_GetByType(t *testing.T) {
	tests := []struct {
		name       string
		notifType  entity.NotificationType
		mockConfig *entity.NotificationConfig
		mockErr    error
		wantErr    bool
	}{
		{
			name:      "成功获取邮件配置",
			notifType: entity.NotificationTypeEmail,
			mockConfig: &entity.NotificationConfig{
				ID:      "config-1",
				Type:    entity.NotificationTypeEmail,
				Name:    "Email Config",
				Enabled: true,
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:       "配置不存在",
			notifType:  entity.NotificationTypeSMS,
			mockConfig: nil,
			mockErr:    errors.New("not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockNotificationConfigRepository)
			service := NewNotificationConfigService(mockRepo)

			mockRepo.On("GetByType", mock.Anything, tt.notifType).
				Return(tt.mockConfig, tt.mockErr)

			config, err := service.GetByType(context.Background(), tt.notifType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, tt.notifType, config.Type)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNotificationConfigService_GetAll(t *testing.T) {
	tests := []struct {
		name        string
		mockConfigs []*entity.NotificationConfig
		mockErr     error
		wantErr     bool
	}{
		{
			name: "成功获取所有配置",
			mockConfigs: []*entity.NotificationConfig{
				{ID: "config-1", Type: entity.NotificationTypeEmail, Name: "Email"},
				{ID: "config-2", Type: entity.NotificationTypeSMS, Name: "SMS"},
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:        "空列表",
			mockConfigs: []*entity.NotificationConfig{},
			mockErr:     nil,
			wantErr:     false,
		},
		{
			name:        "仓储错误",
			mockConfigs: nil,
			mockErr:     errors.New("database error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockNotificationConfigRepository)
			service := NewNotificationConfigService(mockRepo)

			mockRepo.On("GetAll", mock.Anything).Return(tt.mockConfigs, tt.mockErr)

			configs, err := service.GetAll(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, configs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockConfigs, configs)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNotificationConfigService_UpdateConfig_Create(t *testing.T) {
	mockRepo := new(MockNotificationConfigRepository)
	service := NewNotificationConfigService(mockRepo)

	req := &UpdateNotificationConfigRequest{
		Name:   "Email Config",
		Config: map[string]interface{}{"smtp_host": "smtp.example.com"},
	}

	// 配置不存在，创建新配置
	mockRepo.On("GetByType", mock.Anything, entity.NotificationTypeEmail).
		Return(nil, errors.New("not found"))
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.NotificationConfig")).
		Return(nil)

	config, err := service.UpdateConfig(context.Background(), entity.NotificationTypeEmail, req)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, entity.NotificationTypeEmail, config.Type)
	assert.Equal(t, req.Name, config.Name)
	assert.Equal(t, entity.JSONMap(req.Config), config.Config)
	assert.True(t, config.Enabled)
	mockRepo.AssertExpectations(t)
}

func TestNotificationConfigService_UpdateConfig_Update(t *testing.T) {
	mockRepo := new(MockNotificationConfigRepository)
	service := NewNotificationConfigService(mockRepo)

	existingConfig := &entity.NotificationConfig{
		ID:      "config-1",
		Type:    entity.NotificationTypeEmail,
		Name:    "Old Name",
		Config:  map[string]interface{}{"smtp_host": "old.smtp.com"},
		Enabled: true,
	}

	req := &UpdateNotificationConfigRequest{
		Name:   "New Name",
		Config: map[string]interface{}{"smtp_host": "new.smtp.com"},
	}

	mockRepo.On("GetByType", mock.Anything, entity.NotificationTypeEmail).
		Return(existingConfig, nil)
	mockRepo.On("Update", mock.Anything, existingConfig).Return(nil)

	config, err := service.UpdateConfig(context.Background(), entity.NotificationTypeEmail, req)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "New Name", config.Name)
	assert.Equal(t, entity.JSONMap(req.Config), config.Config)
	mockRepo.AssertExpectations(t)
}

func TestNotificationConfigService_EnableConfig(t *testing.T) {
	tests := []struct {
		name       string
		notifType  entity.NotificationType
		mockConfig *entity.NotificationConfig
		mockErr    error
		wantErr    bool
	}{
		{
			name:      "成功启用配置",
			notifType: entity.NotificationTypeEmail,
			mockConfig: &entity.NotificationConfig{
				ID:      "config-1",
				Type:    entity.NotificationTypeEmail,
				Enabled: false,
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:       "配置不存在",
			notifType:  entity.NotificationTypeSMS,
			mockConfig: nil,
			mockErr:    errors.New("not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockNotificationConfigRepository)
			service := NewNotificationConfigService(mockRepo)

			mockRepo.On("GetByType", mock.Anything, tt.notifType).
				Return(tt.mockConfig, tt.mockErr)

			if !tt.wantErr {
				mockRepo.On("Update", mock.Anything, tt.mockConfig).Return(nil)
			}

			err := service.EnableConfig(context.Background(), tt.notifType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "config not found")
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.mockConfig.Enabled)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNotificationConfigService_DisableConfig(t *testing.T) {
	tests := []struct {
		name       string
		notifType  entity.NotificationType
		mockConfig *entity.NotificationConfig
		mockErr    error
		wantErr    bool
	}{
		{
			name:      "成功禁用配置",
			notifType: entity.NotificationTypeEmail,
			mockConfig: &entity.NotificationConfig{
				ID:      "config-1",
				Type:    entity.NotificationTypeEmail,
				Enabled: true,
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:       "配置不存在",
			notifType:  entity.NotificationTypeSMS,
			mockConfig: nil,
			mockErr:    errors.New("not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockNotificationConfigRepository)
			service := NewNotificationConfigService(mockRepo)

			mockRepo.On("GetByType", mock.Anything, tt.notifType).
				Return(tt.mockConfig, tt.mockErr)

			if !tt.wantErr {
				mockRepo.On("Update", mock.Anything, tt.mockConfig).Return(nil)
			}

			err := service.DisableConfig(context.Background(), tt.notifType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "config not found")
			} else {
				assert.NoError(t, err)
				assert.False(t, tt.mockConfig.Enabled)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNotificationConfigService_TestConfig_Email(t *testing.T) {
	tests := []struct {
		name       string
		notifType  entity.NotificationType
		mockConfig *entity.NotificationConfig
		mockErr    error
		wantErr    bool
		errMsg     string
	}{
		{
			name:      "邮件配置测试成功",
			notifType: entity.NotificationTypeEmail,
			mockConfig: &entity.NotificationConfig{
				Type: entity.NotificationTypeEmail,
				Config: map[string]interface{}{
					"smtp_host": "smtp.example.com",
					"smtp_port": 587,
				},
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:      "邮件配置缺少smtp_host",
			notifType: entity.NotificationTypeEmail,
			mockConfig: &entity.NotificationConfig{
				Type: entity.NotificationTypeEmail,
				Config: map[string]interface{}{
					"smtp_port": 587,
				},
			},
			mockErr: nil,
			wantErr: true,
			errMsg:  "smtp_host is required",
		},
		{
			name:       "配置不存在",
			notifType:  entity.NotificationTypeEmail,
			mockConfig: nil,
			mockErr:    errors.New("not found"),
			wantErr:    true,
			errMsg:     "config not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockNotificationConfigRepository)
			service := NewNotificationConfigService(mockRepo)

			mockRepo.On("GetByType", mock.Anything, tt.notifType).
				Return(tt.mockConfig, tt.mockErr)

			err := service.TestConfig(context.Background(), tt.notifType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNotificationConfigService_TestConfig_SMS(t *testing.T) {
	tests := []struct {
		name       string
		notifType  entity.NotificationType
		mockConfig *entity.NotificationConfig
		mockErr    error
		wantErr    bool
		errMsg     string
	}{
		{
			name:      "短信配置测试成功",
			notifType: entity.NotificationTypeSMS,
			mockConfig: &entity.NotificationConfig{
				Type: entity.NotificationTypeSMS,
				Config: map[string]interface{}{
					"access_key": "test-key",
					"secret_key": "test-secret",
				},
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:      "短信配置缺少access_key",
			notifType: entity.NotificationTypeSMS,
			mockConfig: &entity.NotificationConfig{
				Type: entity.NotificationTypeSMS,
				Config: map[string]interface{}{
					"secret_key": "test-secret",
				},
			},
			mockErr: nil,
			wantErr: true,
			errMsg:  "access_key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockNotificationConfigRepository)
			service := NewNotificationConfigService(mockRepo)

			mockRepo.On("GetByType", mock.Anything, tt.notifType).
				Return(tt.mockConfig, tt.mockErr)

			err := service.TestConfig(context.Background(), tt.notifType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNotificationConfigService_TestConfig_Webhook(t *testing.T) {
	tests := []struct {
		name       string
		notifType  entity.NotificationType
		mockConfig *entity.NotificationConfig
		mockErr    error
		wantErr    bool
		errMsg     string
	}{
		{
			name:      "Webhook配置测试成功",
			notifType: entity.NotificationTypeWebhook,
			mockConfig: &entity.NotificationConfig{
				Type: entity.NotificationTypeWebhook,
				Config: map[string]interface{}{
					"url": "https://example.com/webhook",
				},
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:      "Webhook配置缺少url",
			notifType: entity.NotificationTypeWebhook,
			mockConfig: &entity.NotificationConfig{
				Type: entity.NotificationTypeWebhook,
				Config: map[string]interface{}{
					"method": "POST",
				},
			},
			mockErr: nil,
			wantErr: true,
			errMsg:  "url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockNotificationConfigRepository)
			service := NewNotificationConfigService(mockRepo)

			mockRepo.On("GetByType", mock.Anything, tt.notifType).
				Return(tt.mockConfig, tt.mockErr)

			err := service.TestConfig(context.Background(), tt.notifType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNotificationConfigService_TestConfig_WeChat(t *testing.T) {
	tests := []struct {
		name       string
		notifType  entity.NotificationType
		mockConfig *entity.NotificationConfig
		mockErr    error
		wantErr    bool
		errMsg     string
	}{
		{
			name:      "微信配置测试成功",
			notifType: entity.NotificationTypeWeChat,
			mockConfig: &entity.NotificationConfig{
				Type: entity.NotificationTypeWeChat,
				Config: map[string]interface{}{
					"corp_id":  "test-corp",
					"agent_id": "test-agent",
					"secret":   "test-secret",
				},
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:      "微信配置缺少corp_id",
			notifType: entity.NotificationTypeWeChat,
			mockConfig: &entity.NotificationConfig{
				Type: entity.NotificationTypeWeChat,
				Config: map[string]interface{}{
					"agent_id": "test-agent",
				},
			},
			mockErr: nil,
			wantErr: true,
			errMsg:  "corp_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockNotificationConfigRepository)
			service := NewNotificationConfigService(mockRepo)

			mockRepo.On("GetByType", mock.Anything, tt.notifType).
				Return(tt.mockConfig, tt.mockErr)

			err := service.TestConfig(context.Background(), tt.notifType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNotificationConfigService_TestConfig_UnsupportedType(t *testing.T) {
	mockRepo := new(MockNotificationConfigRepository)
	service := NewNotificationConfigService(mockRepo)

	mockConfig := &entity.NotificationConfig{
		Type:   "unsupported",
		Config: map[string]interface{}{},
	}

	mockRepo.On("GetByType", mock.Anything, entity.NotificationType("unsupported")).
		Return(mockConfig, nil)

	err := service.TestConfig(context.Background(), entity.NotificationType("unsupported"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported notification type")
	mockRepo.AssertExpectations(t)
}
