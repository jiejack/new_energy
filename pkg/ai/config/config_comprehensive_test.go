package config

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigAssistant_WithConfig(t *testing.T) {
	cfg := &AssistantConfig{
		MinConfidence:     0.8,
		MaxSuggestions:    5,
		EnableAutoSuggest: true,
		EnableValidation:  true,
		EnableOptimization: true,
	}
	assistant := NewConfigAssistant(cfg)
	require.NotNil(t, assistant)
}

func TestNewConfigAssistant_NilConfig(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	require.NotNil(t, assistant)
	assert.Equal(t, 0.6, assistant.config.MinConfidence)
	assert.Equal(t, 10, assistant.config.MaxSuggestions)
}

func TestConfigAssistant_ParseConfig_Empty(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	_, err := assistant.ParseConfig(context.Background(), "")
	assert.Error(t, err)
}

func TestConfigAssistant_ParseConfig_System(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	parsed, err := assistant.ParseConfig(context.Background(), "配置系统参数，系统名称为监控平台")
	require.NoError(t, err)
	assert.Equal(t, ConfigSystem, parsed.Type)
	assert.Greater(t, parsed.Confidence, 0.0)
}

func TestConfigAssistant_ParseConfig_Device(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	parsed, err := assistant.ParseConfig(context.Background(), "配置设备参数，逆变器采集间隔为5000毫秒")
	require.NoError(t, err)
	assert.Contains(t, []ConfigType{ConfigDevice, ConfigCollector}, parsed.Type)
}

func TestConfigAssistant_ParseConfig_Alarm(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	parsed, err := assistant.ParseConfig(context.Background(), "配置告警阈值，高限阈值为100")
	require.NoError(t, err)
	assert.Equal(t, ConfigAlarm, parsed.Type)
}

func TestConfigAssistant_ParseConfig_Collector(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	parsed, err := assistant.ParseConfig(context.Background(), "配置数据采集频率和批量大小")
	require.NoError(t, err)
	assert.Equal(t, ConfigCollector, parsed.Type)
}

func TestConfigAssistant_ParseConfig_Storage(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	parsed, err := assistant.ParseConfig(context.Background(), "配置存储数据库保留天数")
	require.NoError(t, err)
	assert.Equal(t, ConfigStorage, parsed.Type)
}

func TestConfigAssistant_ParseConfig_Security(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	parsed, err := assistant.ParseConfig(context.Background(), "配置安全认证和加密权限")
	require.NoError(t, err)
	assert.Equal(t, ConfigSecurity, parsed.Type)
}

func TestConfigAssistant_ParseConfig_Network(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	parsed, err := assistant.ParseConfig(context.Background(), "配置网络通信协议和连接参数")
	require.NoError(t, err)
	assert.Equal(t, ConfigNetwork, parsed.Type)
}

func TestConfigAssistant_ParseConfig_Compute(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	parsed, err := assistant.ParseConfig(context.Background(), "配置计算公式和统计分析参数")
	require.NoError(t, err)
	assert.Equal(t, ConfigCompute, parsed.Type)
}

func TestConfigAssistant_ParseConfig_KeyValuePairs(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	parsed, err := assistant.ParseConfig(context.Background(), "device.poll_interval=5000 collector.batch_size=200")
	require.NoError(t, err)
	assert.NotNil(t, parsed.Items)
}

func TestConfigAssistant_GenerateSuggestions_HighPerformance(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	parsed, _ := assistant.ParseConfig(ctx, "高性能高吞吐配置")
	suggestions, err := assistant.GenerateSuggestions(ctx, parsed, map[string]interface{}{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(suggestions), 0)
}

func TestConfigAssistant_GenerateSuggestions_LowPower(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	parsed, _ := assistant.ParseConfig(ctx, "低功耗节能配置")
	suggestions, err := assistant.GenerateSuggestions(ctx, parsed, map[string]interface{}{})
	require.NoError(t, err)
	assert.Greater(t, len(suggestions), 0)
}

func TestConfigAssistant_GenerateSuggestions_Security(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	parsed, _ := assistant.ParseConfig(ctx, "安全加密配置")
	suggestions, err := assistant.GenerateSuggestions(ctx, parsed, map[string]interface{}{})
	require.NoError(t, err)
	assert.Greater(t, len(suggestions), 0)
}

func TestConfigAssistant_GenerateSuggestions_BestPractice_Device(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	parsed, _ := assistant.ParseConfig(ctx, "设备配置")
	currentConfig := map[string]interface{}{
		"device.poll_interval": 500,
	}
	suggestions, err := assistant.GenerateSuggestions(ctx, parsed, currentConfig)
	require.NoError(t, err)
	for _, s := range suggestions {
		if s.Key == "device.poll_interval" {
			assert.Equal(t, 1000, s.SuggestedValue)
		}
	}
}

func TestConfigAssistant_GenerateSuggestions_BestPractice_Alarm(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	parsed, _ := assistant.ParseConfig(ctx, "告警配置")
	currentConfig := map[string]interface{}{
		"alarm.threshold_high": 10.0,
		"alarm.threshold_low":  100.0,
	}
	_, err := assistant.GenerateSuggestions(ctx, parsed, currentConfig)
	require.NoError(t, err)
}

func TestConfigAssistant_GenerateSuggestions_BestPractice_Collector(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	parsed, _ := assistant.ParseConfig(ctx, "采集配置")
	currentConfig := map[string]interface{}{
		"collector.batch_size": 600,
	}
	_, err := assistant.GenerateSuggestions(ctx, parsed, currentConfig)
	require.NoError(t, err)
}

func TestConfigAssistant_GenerateSuggestions_BestPractice_Storage(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	parsed, _ := assistant.ParseConfig(ctx, "存储配置")
	currentConfig := map[string]interface{}{
		"storage.retention_days": 30,
	}
	_, err := assistant.GenerateSuggestions(ctx, parsed, currentConfig)
	require.NoError(t, err)
}

func TestConfigAssistant_GenerateSuggestions_MaxLimit(t *testing.T) {
	cfg := &AssistantConfig{MaxSuggestions: 1}
	assistant := NewConfigAssistant(cfg)
	ctx := context.Background()

	parsed, _ := assistant.ParseConfig(ctx, "高性能高吞吐配置")
	suggestions, err := assistant.GenerateSuggestions(ctx, parsed, map[string]interface{}{})
	require.NoError(t, err)
	assert.LessOrEqual(t, len(suggestions), 1)
}

func TestConfigAssistant_ValidateConfig_Valid(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"ip_address": "192.168.1.1",
		"port":       8080,
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_InvalidPort(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"port": 0,
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.Greater(t, len(result.Errors), 0)
}

func TestConfigAssistant_ValidateConfig_InvalidIP(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"ip_address": "invalid-ip",
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.False(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_InvalidURL(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"url": "not-a-url",
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.False(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_ValidURL(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"url": "https://example.com",
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_InvalidEmail(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"email": "not-an-email",
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.False(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_ValidEmail(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"email": "user@example.com",
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_InvalidInterval(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"interval": 50,
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.False(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_EmptyValue(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"some_key": "",
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.Greater(t, len(result.Warnings), 0)
}

func TestConfigAssistant_ValidateConfig_NilValue(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"some_key": nil,
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.Greater(t, len(result.Warnings), 0)
}

func TestConfigAssistant_ValidateConfig_EnumValues(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	assistant.AddValidator("log_level", &ValidationRule{
		EnumValues: []interface{}{"debug", "info", "warn", "error"},
	})

	config := map[string]interface{}{
		"log_level": "invalid",
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.False(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_CustomFunc(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	assistant.AddValidator("custom_key", &ValidationRule{
		CustomFunc: func(value interface{}) error {
			return fmt.Errorf("custom validation failed")
		},
	})

	config := map[string]interface{}{
		"custom_key": "test",
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.False(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_FloatMinMax(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	assistant.AddValidator("threshold", &ValidationRule{
		MinValue: 0.0,
		MaxValue: 100.0,
	})

	config := map[string]interface{}{
		"threshold": 150.0,
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.False(t, result.Valid)

	config2 := map[string]interface{}{
		"threshold": -5.0,
	}
	result2, _ := assistant.ValidateConfig(ctx, config2)
	assert.False(t, result2.Valid)

	config3 := map[string]interface{}{
		"threshold": 50.0,
	}
	result3, _ := assistant.ValidateConfig(ctx, config3)
	assert.True(t, result3.Valid)
}

func TestConfigAssistant_OptimizeConfig_Performance(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"collector.batch_size":  100,
		"collector.buffer_size": 10000,
	}
	suggestions, err := assistant.OptimizeConfig(ctx, config, "performance")
	require.NoError(t, err)
	assert.Greater(t, len(suggestions), 0)
}

func TestConfigAssistant_OptimizeConfig_Reliability(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"collector.retry_count": 1,
		"device.timeout":        2000,
	}
	suggestions, err := assistant.OptimizeConfig(ctx, config, "reliability")
	require.NoError(t, err)
	assert.Greater(t, len(suggestions), 0)
}

func TestConfigAssistant_OptimizeConfig_Cost(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"storage.retention_days": 365,
		"device.poll_interval":   1000,
	}
	suggestions, err := assistant.OptimizeConfig(ctx, config, "cost")
	require.NoError(t, err)
	assert.Greater(t, len(suggestions), 0)
}

func TestConfigAssistant_OptimizeConfig_Default(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"collector.batch_size":  100,
		"collector.retry_count": 1,
	}
	suggestions, err := assistant.OptimizeConfig(ctx, config, "default")
	require.NoError(t, err)
	assert.NotNil(t, suggestions)
}

func TestConfigAssistant_AddTemplate(t *testing.T) {
	assistant := NewConfigAssistant(nil)

	template := &ConfigTemplate{
		TemplateID: "tpl_custom_001",
		Name:       "Custom Template",
		Type:       ConfigSystem,
		Items: map[string]*ConfigItem{
			"custom.key": {
				Key:          "custom.key",
				Type:         "string",
				DefaultValue: "default",
				Description:  "Custom key",
			},
		},
	}

	err := assistant.AddTemplate(template)
	require.NoError(t, err)

	templates := assistant.GetTemplates(ConfigSystem)
	assert.Greater(t, len(templates), 0)
}

func TestConfigAssistant_AddTemplate_Nil(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	err := assistant.AddTemplate(nil)
	assert.Error(t, err)
}

func TestConfigAssistant_AddValidator(t *testing.T) {
	assistant := NewConfigAssistant(nil)

	rule := &ValidationRule{
		MinValue: 0,
		MaxValue: 100,
	}
	err := assistant.AddValidator("custom_validator", rule)
	require.NoError(t, err)

	validators := assistant.GetValidators()
	assert.Contains(t, validators, "custom_validator")
}

func TestConfigAssistant_AddValidator_Nil(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	err := assistant.AddValidator("key", nil)
	assert.Error(t, err)
}

func TestConfigAssistant_AddRule(t *testing.T) {
	assistant := NewConfigAssistant(nil)

	rule := &ConfigRule{
		RuleID: "rule_custom_001",
		Name:   "Custom Rule",
		Type:   ConfigSystem,
		Conditions: []RuleCondition{
			{Type: "keyword", Value: "自定义", Operator: "contains"},
		},
		Actions: []RuleAction{
			{Type: "suggest", ConfigKey: "custom.key", ConfigValue: "value", Message: "Custom suggestion"},
		},
	}

	err := assistant.AddRule(rule)
	require.NoError(t, err)
}

func TestConfigAssistant_AddRule_Nil(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	err := assistant.AddRule(nil)
	assert.Error(t, err)
}

func TestConfigAssistant_GetTemplates_NotFound(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	templates := assistant.GetTemplates(ConfigType("nonexistent"))
	assert.Nil(t, templates)
}

func TestConfigAssistant_GenerateConfigFromTemplate(t *testing.T) {
	assistant := NewConfigAssistant(nil)

	config, err := assistant.GenerateConfigFromTemplate("tpl_system_001")
	require.NoError(t, err)
	assert.Contains(t, config, "system.name")
	assert.Contains(t, config, "system.log_level")
	assert.Contains(t, config, "system.timezone")
}

func TestConfigAssistant_GenerateConfigFromTemplate_Device(t *testing.T) {
	assistant := NewConfigAssistant(nil)

	config, err := assistant.GenerateConfigFromTemplate("tpl_device_001")
	require.NoError(t, err)
	assert.Contains(t, config, "device.type")
	assert.Contains(t, config, "device.protocol")
}

func TestConfigAssistant_GenerateConfigFromTemplate_NotFound(t *testing.T) {
	assistant := NewConfigAssistant(nil)

	_, err := assistant.GenerateConfigFromTemplate("nonexistent")
	assert.Error(t, err)
}

func TestConfigAssistant_ExplainConfig(t *testing.T) {
	assistant := NewConfigAssistant(nil)

	explanation, err := assistant.ExplainConfig("system.name")
	require.NoError(t, err)
	assert.Contains(t, explanation, "系统名称")
}

func TestConfigAssistant_ExplainConfig_NotFound(t *testing.T) {
	assistant := NewConfigAssistant(nil)

	_, err := assistant.ExplainConfig("nonexistent.key")
	assert.Error(t, err)
}

func TestConfigAssistant_MatchRuleConditions_Pattern(t *testing.T) {
	assistant := NewConfigAssistant(nil)

	rule := &ConfigRule{
		Conditions: []RuleCondition{
			{Type: "pattern", Value: `高性能|高吞吐`},
		},
	}
	assert.True(t, assistant.matchRuleConditions(rule, "高性能配置"))
	assert.False(t, assistant.matchRuleConditions(rule, "普通配置"))
}

func TestConfigAssistant_MatchRuleConditions_Keyword(t *testing.T) {
	assistant := NewConfigAssistant(nil)

	rule := &ConfigRule{
		Conditions: []RuleCondition{
			{Type: "keyword", Value: "高性能", Operator: "contains"},
		},
	}
	assert.True(t, assistant.matchRuleConditions(rule, "需要高性能配置"))
	assert.False(t, assistant.matchRuleConditions(rule, "需要低功耗配置"))
}

func TestConfigAssistant_ConvertValue(t *testing.T) {
	assistant := NewConfigAssistant(nil)

	assert.Equal(t, 42, assistant.convertValue("42", "int"))
	assert.Equal(t, 3.14, assistant.convertValue("3.14", "float"))
	assert.True(t, assistant.convertValue("true", "bool").(bool))
	assert.True(t, assistant.convertValue("1", "bool").(bool))
	assert.True(t, assistant.convertValue("是", "bool").(bool))
	assert.True(t, assistant.convertValue("启用", "bool").(bool))
	assert.False(t, assistant.convertValue("false", "bool").(bool))
	assert.Equal(t, "hello", assistant.convertValue("hello", "string"))
	assert.Equal(t, "test", assistant.convertValue("test", "unknown"))
}

func TestConfigAssistant_ConvertValue_Auto(t *testing.T) {
	assistant := NewConfigAssistant(nil)

	assert.Equal(t, true, assistant.convertValue("true", "auto"))
	assert.Equal(t, false, assistant.convertValue("false", "auto"))
	assert.Equal(t, 3.14, assistant.convertValue("3.14", "auto"))
	assert.Equal(t, 42, assistant.convertValue("42", "auto"))
	assert.Equal(t, "hello", assistant.convertValue("hello", "auto"))
}

func TestConfigAssistant_InferType(t *testing.T) {
	assistant := NewConfigAssistant(nil)

	assert.Equal(t, "int", assistant.inferType(42))
	assert.Equal(t, "float", assistant.inferType(3.14))
	assert.Equal(t, "bool", assistant.inferType(true))
	assert.Equal(t, "string", assistant.inferType("hello"))
	assert.Equal(t, "unknown", assistant.inferType([]int{1, 2}))
}

func TestDefaultAssistantConfig(t *testing.T) {
	cfg := DefaultAssistantConfig()
	assert.Equal(t, 0.6, cfg.MinConfidence)
	assert.Equal(t, 10, cfg.MaxSuggestions)
	assert.True(t, cfg.EnableAutoSuggest)
	assert.True(t, cfg.EnableValidation)
	assert.True(t, cfg.EnableOptimization)
}

func TestConfigAssistant_ParseConfig_WithKeyValuePairs(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	parsed, err := assistant.ParseConfig(context.Background(), "device.poll_interval=10000 alarm.enabled=true")
	require.NoError(t, err)
	assert.NotNil(t, parsed.Items)
}

func TestConfigAssistant_ParseConfig_WithChineseKeyValue(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	parsed, err := assistant.ParseConfig(context.Background(), "系统名称为监控平台 device.type=inverter")
	require.NoError(t, err)
	assert.NotNil(t, parsed.Items)
}

func TestConfigAssistant_ValidateConfig_PortOverMax(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"port": 70000,
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.False(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_IntervalOverMax(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"interval": 99999999,
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.False(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_ValidInterval(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"interval": 5000,
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_NoValidator(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"unknown_key": "some_value",
	}
	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestConfigAssistant_OptimizeConfig_PerformanceAlreadyOptimal(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"collector.batch_size":  500,
		"collector.buffer_size": 50000,
	}
	suggestions, err := assistant.OptimizeConfig(ctx, config, "performance")
	require.NoError(t, err)
	assert.Equal(t, 0, len(suggestions))
}

func TestConfigAssistant_OptimizeConfig_ReliabilityAlreadyOptimal(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"collector.retry_count": 5,
		"device.timeout":        10000,
	}
	suggestions, err := assistant.OptimizeConfig(ctx, config, "reliability")
	require.NoError(t, err)
	assert.Equal(t, 0, len(suggestions))
}

func TestConfigAssistant_OptimizeConfig_CostAlreadyOptimal(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"storage.retention_days": 90,
		"device.poll_interval":   10000,
	}
	suggestions, err := assistant.OptimizeConfig(ctx, config, "cost")
	require.NoError(t, err)
	assert.Equal(t, 0, len(suggestions))
}

func TestConfigAssistant_DeduplicateSuggestions(t *testing.T) {
	assistant := NewConfigAssistant(nil)

	suggestions := []*ConfigSuggestion{
		{Key: "key1", SuggestedValue: 1},
		{Key: "key1", SuggestedValue: 2},
		{Key: "key2", SuggestedValue: 3},
	}

	result := assistant.deduplicateSuggestions(suggestions)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 1, result[0].SuggestedValue)
}
