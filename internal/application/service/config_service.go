package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

var (
	ErrConfigNotFound     = errors.New("config not found")
	ErrConfigKeyExists    = errors.New("config key already exists")
	ErrInvalidValueType   = errors.New("invalid value type")
	ErrValueConversion    = errors.New("value conversion failed")
)

// ConfigService 系统配置服务
type ConfigService struct {
	configRepo repository.SystemConfigRepository
	logRepo    repository.OperationLogRepository
}

// NewConfigService 创建系统配置服务
func NewConfigService(
	configRepo repository.SystemConfigRepository,
	logRepo repository.OperationLogRepository,
) *ConfigService {
	return &ConfigService{
		configRepo: configRepo,
		logRepo:    logRepo,
	}
}

// CreateConfigRequest 创建配置请求
type CreateConfigRequest struct {
	Category    string `json:"category" binding:"required"`
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	ValueType   string `json:"value_type" binding:"required"`
	Description string `json:"description"`
}

// UpdateConfigRequest 更新配置请求
type UpdateConfigRequest struct {
	Value       string `json:"value" binding:"required"`
	ValueType   string `json:"value_type"`
	Description string `json:"description"`
}

// BatchUpdateConfigRequest 批量更新配置请求
type BatchUpdateConfigRequest struct {
	Configs []ConfigUpdateItem `json:"configs" binding:"required"`
}

// ConfigUpdateItem 配置更新项
type ConfigUpdateItem struct {
	Category  string `json:"category" binding:"required"`
	Key       string `json:"key" binding:"required"`
	Value     string `json:"value" binding:"required"`
	ValueType string `json:"value_type"`
}

// ConfigListResponse 配置列表响应
type ConfigListResponse struct {
	Configs  []*entity.SystemConfig `json:"configs"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// ConfigCategoryResponse 配置分类响应
type ConfigCategoryResponse struct {
	Category string                 `json:"category"`
	Configs  []*entity.SystemConfig `json:"configs"`
}

// CreateConfig 创建配置
func (s *ConfigService) CreateConfig(ctx context.Context, req *CreateConfigRequest, operatorID string) (*entity.SystemConfig, error) {
	// 检查键是否已存在
	exists, err := s.configRepo.ExistsByKey(ctx, req.Category, req.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to check config key: %w", err)
	}
	if exists {
		return nil, ErrConfigKeyExists
	}

	// 验证值类型
	valueType := entity.SystemConfigValueType(req.ValueType)
	if !isValidValueType(valueType) {
		return nil, ErrInvalidValueType
	}

	// 验证值是否能正确转换
	if err := validateValue(req.Value, valueType); err != nil {
		return nil, err
	}

	config := entity.NewSystemConfig(req.Category, req.Key, req.Value, valueType, req.Description)

	if err := s.configRepo.Create(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	s.logOperation(ctx, operatorID, entity.ActionCreate, entity.ResourceSystemConfig, config.ID, entity.Details{
		"category":    config.Category,
		"key":         config.Key,
		"value":       config.Value,
		"value_type":  config.ValueType,
		"description": config.Description,
	})

	return config, nil
}

// UpdateConfig 更新配置
func (s *ConfigService) UpdateConfig(ctx context.Context, category, key string, req *UpdateConfigRequest, operatorID string) (*entity.SystemConfig, error) {
	config, err := s.configRepo.GetByKey(ctx, category, key)
	if err != nil {
		return nil, ErrConfigNotFound
	}

	// 如果提供了新的值类型，验证并更新
	if req.ValueType != "" {
		valueType := entity.SystemConfigValueType(req.ValueType)
		if !isValidValueType(valueType) {
			return nil, ErrInvalidValueType
		}
		if err := validateValue(req.Value, valueType); err != nil {
			return nil, err
		}
		config.SetValue(req.Value, valueType)
	} else {
		// 使用原有的值类型验证
		if err := validateValue(req.Value, config.ValueType); err != nil {
			return nil, err
		}
		config.SetValue(req.Value, config.ValueType)
	}

	// 更新描述
	if req.Description != "" {
		config.SetDescription(req.Description)
	}

	if err := s.configRepo.Update(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update config: %w", err)
	}

	s.logOperation(ctx, operatorID, entity.ActionUpdate, entity.ResourceSystemConfig, config.ID, entity.Details{
		"category":    config.Category,
		"key":         config.Key,
		"value":       config.Value,
		"value_type":  config.ValueType,
		"description": config.Description,
	})

	return config, nil
}

// DeleteConfig 删除配置
func (s *ConfigService) DeleteConfig(ctx context.Context, category, key string, operatorID string) error {
	config, err := s.configRepo.GetByKey(ctx, category, key)
	if err != nil {
		return ErrConfigNotFound
	}

	if err := s.configRepo.Delete(ctx, config.ID); err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	s.logOperation(ctx, operatorID, entity.ActionDelete, entity.ResourceSystemConfig, config.ID, entity.Details{
		"category": config.Category,
		"key":      config.Key,
	})

	return nil
}

// GetConfig 获取单个配置
func (s *ConfigService) GetConfig(ctx context.Context, category, key string) (*entity.SystemConfig, error) {
	return s.configRepo.GetByKey(ctx, category, key)
}

// GetConfigByID 根据ID获取配置
func (s *ConfigService) GetConfigByID(ctx context.Context, id string) (*entity.SystemConfig, error) {
	return s.configRepo.GetByID(ctx, id)
}

// GetConfigsByCategory 获取指定分类的配置
func (s *ConfigService) GetConfigsByCategory(ctx context.Context, category string) (*ConfigCategoryResponse, error) {
	configs, err := s.configRepo.GetByCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get configs by category: %w", err)
	}

	return &ConfigCategoryResponse{
		Category: category,
		Configs:  configs,
	}, nil
}

// GetAllConfigs 获取所有配置
func (s *ConfigService) GetAllConfigs(ctx context.Context) ([]*entity.SystemConfig, error) {
	return s.configRepo.GetAll(ctx)
}

// ListConfigs 分页查询配置列表
func (s *ConfigService) ListConfigs(ctx context.Context, filter *entity.SystemConfigFilter) (*ConfigListResponse, error) {
	configs, total, err := s.configRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}

	return &ConfigListResponse{
		Configs:  configs,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// BatchUpdateConfigs 批量更新配置
func (s *ConfigService) BatchUpdateConfigs(ctx context.Context, req *BatchUpdateConfigRequest, operatorID string) error {
	configs := make([]*entity.SystemConfig, 0, len(req.Configs))

	for _, item := range req.Configs {
		config, err := s.configRepo.GetByKey(ctx, item.Category, item.Key)
		if err != nil {
			return fmt.Errorf("config not found: %s/%s", item.Category, item.Key)
		}

		// 确定值类型
		valueType := config.ValueType
		if item.ValueType != "" {
			valueType = entity.SystemConfigValueType(item.ValueType)
			if !isValidValueType(valueType) {
				return fmt.Errorf("invalid value type for %s/%s: %s", item.Category, item.Key, item.ValueType)
			}
		}

		// 验证值
		if err := validateValue(item.Value, valueType); err != nil {
			return fmt.Errorf("invalid value for %s/%s: %w", item.Category, item.Key, err)
		}

		config.SetValue(item.Value, valueType)
		configs = append(configs, config)
	}

	if err := s.configRepo.BatchUpdate(ctx, configs); err != nil {
		return fmt.Errorf("failed to batch update configs: %w", err)
	}

	s.logOperation(ctx, operatorID, entity.ActionUpdate, entity.ResourceSystemConfig, "", entity.Details{
		"count": len(configs),
	})

	return nil
}

// GetConfigValue 获取配置值（带类型转换）
func (s *ConfigService) GetConfigValue(ctx context.Context, category, key string) (interface{}, error) {
	config, err := s.configRepo.GetByKey(ctx, category, key)
	if err != nil {
		return nil, ErrConfigNotFound
	}

	return convertValue(config.Value, config.ValueType)
}

// GetConfigAsString 获取配置值作为字符串
func (s *ConfigService) GetConfigAsString(ctx context.Context, category, key string) (string, error) {
	config, err := s.configRepo.GetByKey(ctx, category, key)
	if err != nil {
		return "", ErrConfigNotFound
	}
	return config.Value, nil
}

// GetConfigAsInt 获取配置值作为整数
func (s *ConfigService) GetConfigAsInt(ctx context.Context, category, key string) (int, error) {
	config, err := s.configRepo.GetByKey(ctx, category, key)
	if err != nil {
		return 0, ErrConfigNotFound
	}

	val, err := strconv.Atoi(config.Value)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrValueConversion, err)
	}
	return val, nil
}

// GetConfigAsBool 获取配置值作为布尔值
func (s *ConfigService) GetConfigAsBool(ctx context.Context, category, key string) (bool, error) {
	config, err := s.configRepo.GetByKey(ctx, category, key)
	if err != nil {
		return false, ErrConfigNotFound
	}

	val, err := strconv.ParseBool(config.Value)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrValueConversion, err)
	}
	return val, nil
}

// GetConfigAsJSON 获取配置值作为JSON对象
func (s *ConfigService) GetConfigAsJSON(ctx context.Context, category, key string, v interface{}) error {
	config, err := s.configRepo.GetByKey(ctx, category, key)
	if err != nil {
		return ErrConfigNotFound
	}

	if err := json.Unmarshal([]byte(config.Value), v); err != nil {
		return fmt.Errorf("%w: %v", ErrValueConversion, err)
	}
	return nil
}

// 辅助函数

// isValidValueType 验证值类型是否有效
func isValidValueType(valueType entity.SystemConfigValueType) bool {
	switch valueType {
	case entity.SystemConfigValueTypeString,
		entity.SystemConfigValueTypeInt,
		entity.SystemConfigValueTypeBool,
		entity.SystemConfigValueTypeJSON:
		return true
	default:
		return false
	}
}

// validateValue 验证值是否能正确转换
func validateValue(value string, valueType entity.SystemConfigValueType) error {
	switch valueType {
	case entity.SystemConfigValueTypeInt:
		_, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("%w: invalid int value", ErrValueConversion)
		}
	case entity.SystemConfigValueTypeBool:
		_, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("%w: invalid bool value", ErrValueConversion)
		}
	case entity.SystemConfigValueTypeJSON:
		if !json.Valid([]byte(value)) {
			return fmt.Errorf("%w: invalid JSON value", ErrValueConversion)
		}
	}
	return nil
}

// convertValue 将字符串值转换为对应类型
func convertValue(value string, valueType entity.SystemConfigValueType) (interface{}, error) {
	switch valueType {
	case entity.SystemConfigValueTypeString:
		return value, nil
	case entity.SystemConfigValueTypeInt:
		return strconv.Atoi(value)
	case entity.SystemConfigValueTypeBool:
		return strconv.ParseBool(value)
	case entity.SystemConfigValueTypeJSON:
		var result interface{}
		if err := json.Unmarshal([]byte(value), &result); err != nil {
			return nil, err
		}
		return result, nil
	default:
		return value, nil
	}
}

// logOperation 记录操作日志
func (s *ConfigService) logOperation(ctx context.Context, operatorID, action, resourceType, resourceID string, details entity.Details) {
	log := entity.NewOperationLog(operatorID, "", action)
	log.SetResource(resourceType, resourceID)
	if details != nil {
		log.SetDetails(details)
	}
	_ = s.logRepo.Create(ctx, log)
}
