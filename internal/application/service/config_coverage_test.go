package service

import (
	"context"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfigService_UpdateConfig_WithDescription2(t *testing.T) {
	configRepo := NewMockSystemConfigRepository()
	opRepo := new(MockOperationLogRepositoryForConfigService)
	svc := NewConfigService(configRepo, opRepo)

	existing := &entity.SystemConfig{ID: "c1", Key: "test.key", Value: "old", Description: "old desc", Category: "system", ValueType: entity.SystemConfigValueTypeString}
	configRepo.On("GetByKey", mock.Anything, "system", "test.key").Return(existing, nil)
	configRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.SystemConfig")).Return(nil)
	opRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	req := &UpdateConfigRequest{Value: "new", Description: "new desc"}
	config, err := svc.UpdateConfig(context.Background(), "system", "test.key", req, "admin")
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "new", config.Value)
}

func TestConfigService_UpdateConfig_NotFound3(t *testing.T) {
	configRepo := NewMockSystemConfigRepository()
	opRepo := new(MockOperationLogRepositoryForConfigService)
	svc := NewConfigService(configRepo, opRepo)

	configRepo.On("GetByKey", mock.Anything, "system", "nonexistent").Return(nil, assert.AnError)

	req := &UpdateConfigRequest{Value: "new"}
	config, err := svc.UpdateConfig(context.Background(), "system", "nonexistent", req, "admin")
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestConfigService_GetConfigsByCategory2(t *testing.T) {
	configRepo := NewMockSystemConfigRepository()
	opRepo := new(MockOperationLogRepositoryForConfigService)
	svc := NewConfigService(configRepo, opRepo)

	configRepo.On("GetByCategory", mock.Anything, "system").Return([]*entity.SystemConfig{{ID: "c1"}}, nil)

	result, err := svc.GetConfigsByCategory(context.Background(), "system")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "system", result.Category)
}

func TestConfigService_ListConfigs2(t *testing.T) {
	configRepo := NewMockSystemConfigRepository()
	opRepo := new(MockOperationLogRepositoryForConfigService)
	svc := NewConfigService(configRepo, opRepo)

	configRepo.On("List", mock.Anything, mock.AnythingOfType("*entity.SystemConfigFilter")).Return([]*entity.SystemConfig{{ID: "c1"}}, int64(1), nil)

	result, err := svc.ListConfigs(context.Background(), &entity.SystemConfigFilter{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestConfigService_BatchUpdateConfigs2(t *testing.T) {
	configRepo := NewMockSystemConfigRepository()
	opRepo := new(MockOperationLogRepositoryForConfigService)
	svc := NewConfigService(configRepo, opRepo)

	configRepo.On("GetByKey", mock.Anything, "system", "key1").Return(&entity.SystemConfig{ID: "c1", Key: "key1", Category: "system", Value: "old", ValueType: entity.SystemConfigValueTypeString}, nil)
	configRepo.On("BatchUpdate", mock.Anything, mock.Anything).Return(nil)
	opRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.OperationLog")).Return(nil)

	req := &BatchUpdateConfigRequest{
		Configs: []ConfigUpdateItem{
			{Category: "system", Key: "key1", Value: "new1"},
		},
	}
	err := svc.BatchUpdateConfigs(context.Background(), req, "admin")
	assert.NoError(t, err)
}

func TestConfigService_GetConfigAsInt_Found2(t *testing.T) {
	configRepo := NewMockSystemConfigRepository()
	opRepo := new(MockOperationLogRepositoryForConfigService)
	svc := NewConfigService(configRepo, opRepo)

	configRepo.On("GetByKey", mock.Anything, "system", "int_key").Return(&entity.SystemConfig{Value: "42"}, nil)

	result, err := svc.GetConfigAsInt(context.Background(), "system", "int_key")
	assert.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestConfigService_GetConfigAsBool_Found2(t *testing.T) {
	configRepo := NewMockSystemConfigRepository()
	opRepo := new(MockOperationLogRepositoryForConfigService)
	svc := NewConfigService(configRepo, opRepo)

	configRepo.On("GetByKey", mock.Anything, "system", "bool_key").Return(&entity.SystemConfig{Value: "true"}, nil)

	result, err := svc.GetConfigAsBool(context.Background(), "system", "bool_key")
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestConfigService_GetConfigAsJSON4(t *testing.T) {
	configRepo := NewMockSystemConfigRepository()
	opRepo := new(MockOperationLogRepositoryForConfigService)
	svc := NewConfigService(configRepo, opRepo)

	configRepo.On("GetByKey", mock.Anything, "system", "json_key").Return(&entity.SystemConfig{Value: `{"a":1}`}, nil)

	result := svc.GetConfigAsJSON(context.Background(), "system", "json_key", map[string]interface{}{})
	assert.NotNil(t, result)
}

func TestConfigService_BatchUpdateConfigs_ConfigNotFound2(t *testing.T) {
	configRepo := NewMockSystemConfigRepository()
	opRepo := new(MockOperationLogRepositoryForConfigService)
	svc := NewConfigService(configRepo, opRepo)

	configRepo.On("GetByKey", mock.Anything, "system", "nonexistent").Return(nil, assert.AnError)

	req := &BatchUpdateConfigRequest{
		Configs: []ConfigUpdateItem{
			{Category: "system", Key: "nonexistent", Value: "new"},
		},
	}
	err := svc.BatchUpdateConfigs(context.Background(), req, "admin")
	assert.Error(t, err)
}

func TestConfigService_GetConfigAsInt_NotFound2(t *testing.T) {
	configRepo := NewMockSystemConfigRepository()
	opRepo := new(MockOperationLogRepositoryForConfigService)
	svc := NewConfigService(configRepo, opRepo)

	configRepo.On("GetByKey", mock.Anything, "system", "nonexistent").Return(nil, assert.AnError)

	result, err := svc.GetConfigAsInt(context.Background(), "system", "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, 0, result)
}

func TestConfigService_GetConfigAsBool_NotFound2(t *testing.T) {
	configRepo := NewMockSystemConfigRepository()
	opRepo := new(MockOperationLogRepositoryForConfigService)
	svc := NewConfigService(configRepo, opRepo)

	configRepo.On("GetByKey", mock.Anything, "system", "nonexistent").Return(nil, assert.AnError)

	result, err := svc.GetConfigAsBool(context.Background(), "system", "nonexistent")
	assert.Error(t, err)
	assert.False(t, result)
}
