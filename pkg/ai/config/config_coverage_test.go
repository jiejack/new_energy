package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultAssistantConfig(t *testing.T) {
	cfg := DefaultAssistantConfig()
	assert.Equal(t, 0.6, cfg.MinConfidence)
	assert.Equal(t, 10, cfg.MaxSuggestions)
	assert.True(t, cfg.EnableAutoSuggest)
	assert.True(t, cfg.EnableValidation)
	assert.True(t, cfg.EnableOptimization)
}

func TestNewConfigAssistant_NilConfig(t *testing.T) {
	a := NewConfigAssistant(nil)
	require.NotNil(t, a)
}

func TestConfigAssistant_ParseConfig_EmptyText(t *testing.T) {
	a := NewConfigAssistant(nil)
	_, err := a.ParseConfig(context.Background(), "")
	assert.Error(t, err)
}

func TestConfigAssistant_ParseConfig_DeviceType(t *testing.T) {
	a := NewConfigAssistant(nil)
	parsed, err := a.ParseConfig(context.Background(), "设备配置，逆变器类型为modbus协议")
	require.NoError(t, err)
	assert.Equal(t, ConfigDevice, parsed.Type)
}

func TestConfigAssistant_ParseConfig_AlarmType(t *testing.T) {
	a := NewConfigAssistant(nil)
	parsed, err := a.ParseConfig(context.Background(), "告警配置，阈值设置为100")
	require.NoError(t, err)
	assert.Equal(t, ConfigAlarm, parsed.Type)
}

func TestConfigAssistant_ParseConfig_CollectorType(t *testing.T) {
	a := NewConfigAssistant(nil)
	parsed, err := a.ParseConfig(context.Background(), "采集配置，批量大小为200")
	require.NoError(t, err)
	assert.Equal(t, ConfigCollector, parsed.Type)
}

func TestConfigAssistant_ParseConfig_StorageType(t *testing.T) {
	a := NewConfigAssistant(nil)
	parsed, err := a.ParseConfig(context.Background(), "存储配置，保留365天")
	require.NoError(t, err)
	assert.Equal(t, ConfigStorage, parsed.Type)
}

func TestConfigAssistant_ParseConfig_SecurityType(t *testing.T) {
	a := NewConfigAssistant(nil)
	parsed, err := a.ParseConfig(context.Background(), "安全配置，启用加密")
	require.NoError(t, err)
	assert.Equal(t, ConfigSecurity, parsed.Type)
}

func TestConfigAssistant_ParseConfig_NetworkType(t *testing.T) {
	a := NewConfigAssistant(nil)
	parsed, err := a.ParseConfig(context.Background(), "网络配置，通信协议mqtt")
	require.NoError(t, err)
	assert.Equal(t, ConfigNetwork, parsed.Type)
}

func TestConfigAssistant_ParseConfig_ComputeType(t *testing.T) {
	a := NewConfigAssistant(nil)
	parsed, err := a.ParseConfig(context.Background(), "计算配置，统计分析公式")
	require.NoError(t, err)
	assert.Equal(t, ConfigCompute, parsed.Type)
}

func TestConfigAssistant_GenerateSuggestions_PerformanceRule(t *testing.T) {
	a := NewConfigAssistant(nil)
	parsed := &ParsedConfig{
		Type:    ConfigSystem,
		RawText: "高性能配置",
	}
	suggestions, err := a.GenerateSuggestions(context.Background(), parsed, map[string]interface{}{})
	require.NoError(t, err)
	_ = suggestions
}

func TestConfigAssistant_GenerateSuggestions_PowerRule(t *testing.T) {
	a := NewConfigAssistant(nil)
	parsed := &ParsedConfig{
		Type:    ConfigSystem,
		RawText: "低功耗节能配置",
	}
	suggestions, err := a.GenerateSuggestions(context.Background(), parsed, map[string]interface{}{})
	require.NoError(t, err)
	assert.NotEmpty(t, suggestions)
}

func TestConfigAssistant_GenerateSuggestions_SecurityRule(t *testing.T) {
	a := NewConfigAssistant(nil)
	parsed := &ParsedConfig{
		Type:    ConfigSecurity,
		RawText: "安全加密配置",
	}
	suggestions, err := a.GenerateSuggestions(context.Background(), parsed, map[string]interface{}{})
	require.NoError(t, err)
	assert.NotEmpty(t, suggestions)
}

func TestConfigAssistant_GenerateSuggestions_BestPractice(t *testing.T) {
	a := NewConfigAssistant(nil)
	parsed := &ParsedConfig{Type: ConfigDevice}
	suggestions, err := a.GenerateSuggestions(context.Background(), parsed, map[string]interface{}{
		"device.poll_interval": 500,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, suggestions)
}

func TestConfigAssistant_GenerateSuggestions_AlarmBestPractice(t *testing.T) {
	a := NewConfigAssistant(nil)
	parsed := &ParsedConfig{Type: ConfigAlarm}
	suggestions, err := a.GenerateSuggestions(context.Background(), parsed, map[string]interface{}{
		"alarm.threshold_high": 50.0,
		"alarm.threshold_low":  80.0,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, suggestions)
}

func TestConfigAssistant_GenerateSuggestions_CollectorBestPractice(t *testing.T) {
	a := NewConfigAssistant(nil)
	parsed := &ParsedConfig{Type: ConfigCollector}
	suggestions, err := a.GenerateSuggestions(context.Background(), parsed, map[string]interface{}{
		"collector.batch_size": 600,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, suggestions)
}

func TestConfigAssistant_GenerateSuggestions_StorageBestPractice(t *testing.T) {
	a := NewConfigAssistant(nil)
	parsed := &ParsedConfig{Type: ConfigStorage}
	suggestions, err := a.GenerateSuggestions(context.Background(), parsed, map[string]interface{}{
		"storage.retention_days": 30,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, suggestions)
}

func TestConfigAssistant_ValidateConfig_Valid(t *testing.T) {
	a := NewConfigAssistant(nil)
	result, err := a.ValidateConfig(context.Background(), map[string]interface{}{
		"ip_address": "192.168.1.1",
		"port":       8080,
	})
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_IPInvalid(t *testing.T) {
	a := NewConfigAssistant(nil)
	result, err := a.ValidateConfig(context.Background(), map[string]interface{}{
		"ip_address": "invalid-ip",
	})
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
}

func TestConfigAssistant_ValidateConfig_PortInvalid(t *testing.T) {
	a := NewConfigAssistant(nil)
	result, err := a.ValidateConfig(context.Background(), map[string]interface{}{
		"port": 99999,
	})
	require.NoError(t, err)
	assert.False(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_URLInvalid(t *testing.T) {
	a := NewConfigAssistant(nil)
	result, err := a.ValidateConfig(context.Background(), map[string]interface{}{
		"url": "not-a-url",
	})
	require.NoError(t, err)
	assert.False(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_EmailInvalid(t *testing.T) {
	a := NewConfigAssistant(nil)
	result, err := a.ValidateConfig(context.Background(), map[string]interface{}{
		"email": "not-an-email",
	})
	require.NoError(t, err)
	assert.False(t, result.Valid)
}

func TestConfigAssistant_ValidateConfig_Warnings(t *testing.T) {
	a := NewConfigAssistant(nil)
	result, err := a.ValidateConfig(context.Background(), map[string]interface{}{
		"some_key": "",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, result.Warnings)
}

func TestConfigAssistant_OptimizeConfig_Performance(t *testing.T) {
	a := NewConfigAssistant(nil)
	suggestions, err := a.OptimizeConfig(context.Background(), map[string]interface{}{
		"collector.batch_size":  100,
		"collector.buffer_size": 10000,
	}, "performance")
	require.NoError(t, err)
	assert.NotEmpty(t, suggestions)
}

func TestConfigAssistant_OptimizeConfig_Reliability(t *testing.T) {
	a := NewConfigAssistant(nil)
	suggestions, err := a.OptimizeConfig(context.Background(), map[string]interface{}{
		"collector.retry_count": 1,
		"device.timeout":         2000,
	}, "reliability")
	require.NoError(t, err)
	assert.NotEmpty(t, suggestions)
}

func TestConfigAssistant_OptimizeConfig_Cost(t *testing.T) {
	a := NewConfigAssistant(nil)
	suggestions, err := a.OptimizeConfig(context.Background(), map[string]interface{}{
		"storage.retention_days": 365,
		"device.poll_interval":   1000,
	}, "cost")
	require.NoError(t, err)
	assert.NotEmpty(t, suggestions)
}

func TestConfigAssistant_OptimizeConfig_Default(t *testing.T) {
	a := NewConfigAssistant(nil)
	suggestions, err := a.OptimizeConfig(context.Background(), map[string]interface{}{
		"collector.batch_size":   100,
		"collector.retry_count":  1,
		"storage.retention_days": 365,
	}, "")
	require.NoError(t, err)
	assert.NotEmpty(t, suggestions)
}

func TestConfigAssistant_AddTemplate_Nil(t *testing.T) {
	a := NewConfigAssistant(nil)
	err := a.AddTemplate(nil)
	assert.Error(t, err)
}

func TestConfigAssistant_AddTemplate_Success(t *testing.T) {
	a := NewConfigAssistant(nil)
	tmpl := &ConfigTemplate{
		TemplateID: "tpl_custom",
		Name:       "Custom Template",
		Type:       ConfigSystem,
		Items:      map[string]*ConfigItem{"custom.key": {Key: "custom.key", Type: "string"}},
		Priority:   1,
	}
	err := a.AddTemplate(tmpl)
	require.NoError(t, err)
	templates := a.GetTemplates(ConfigSystem)
	assert.GreaterOrEqual(t, len(templates), 1)
}

func TestConfigAssistant_AddValidator_Nil(t *testing.T) {
	a := NewConfigAssistant(nil)
	err := a.AddValidator("test", nil)
	assert.Error(t, err)
}

func TestConfigAssistant_AddValidator_Success(t *testing.T) {
	a := NewConfigAssistant(nil)
	rule := &ValidationRule{MinValue: 1, MaxValue: 100}
	err := a.AddValidator("custom_validator", rule)
	require.NoError(t, err)
	validators := a.GetValidators()
	assert.Contains(t, validators, "custom_validator")
}

func TestConfigAssistant_AddRule_Nil(t *testing.T) {
	a := NewConfigAssistant(nil)
	err := a.AddRule(nil)
	assert.Error(t, err)
}

func TestConfigAssistant_AddRule_Success(t *testing.T) {
	a := NewConfigAssistant(nil)
	rule := &ConfigRule{
		RuleID: "rule_test",
		Name:   "Test Rule",
		Type:   ConfigSystem,
		Priority: 1,
	}
	err := a.AddRule(rule)
	require.NoError(t, err)
}

func TestConfigAssistant_GetTemplates_NotFound(t *testing.T) {
	a := NewConfigAssistant(nil)
	templates := a.GetTemplates("nonexistent_type")
	assert.Nil(t, templates)
}

func TestConfigAssistant_GetValidators(t *testing.T) {
	a := NewConfigAssistant(nil)
	validators := a.GetValidators()
	assert.Contains(t, validators, "ip_address")
	assert.Contains(t, validators, "port")
	assert.Contains(t, validators, "url")
	assert.Contains(t, validators, "email")
	assert.Contains(t, validators, "interval")
}

func TestConfigAssistant_GenerateConfigFromTemplate_Found(t *testing.T) {
	a := NewConfigAssistant(nil)
	config, err := a.GenerateConfigFromTemplate("tpl_system_001")
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Contains(t, config, "system.name")
}

func TestConfigAssistant_GenerateConfigFromTemplate_NotFound(t *testing.T) {
	a := NewConfigAssistant(nil)
	_, err := a.GenerateConfigFromTemplate("nonexistent_template")
	assert.Error(t, err)
}

func TestConfigAssistant_ExplainConfig_Found(t *testing.T) {
	a := NewConfigAssistant(nil)
	explanation, err := a.ExplainConfig("system.name")
	require.NoError(t, err)
	assert.Contains(t, explanation, "系统名称")
}

func TestConfigAssistant_ExplainConfig_NotFound(t *testing.T) {
	a := NewConfigAssistant(nil)
	_, err := a.ExplainConfig("nonexistent.key")
	assert.Error(t, err)
}

func TestConfigAssistant_ValidateValue_IntMin(t *testing.T) {
	a := NewConfigAssistant(nil)
	rule := &ValidationRule{MinValue: 10}
	err := a.validateValue("test", 5, rule)
	assert.Error(t, err)
}

func TestConfigAssistant_ValidateValue_IntMax(t *testing.T) {
	a := NewConfigAssistant(nil)
	rule := &ValidationRule{MaxValue: 100}
	err := a.validateValue("test", 150, rule)
	assert.Error(t, err)
}

func TestConfigAssistant_ValidateValue_FloatMin(t *testing.T) {
	a := NewConfigAssistant(nil)
	rule := &ValidationRule{MinValue: 10.0}
	err := a.validateValue("test", 5.0, rule)
	assert.Error(t, err)
}

func TestConfigAssistant_ValidateValue_FloatMax(t *testing.T) {
	a := NewConfigAssistant(nil)
	rule := &ValidationRule{MaxValue: 100.0}
	err := a.validateValue("test", 150.0, rule)
	assert.Error(t, err)
}

func TestConfigAssistant_ValidateValue_Pattern(t *testing.T) {
	a := NewConfigAssistant(nil)
	rule := &ValidationRule{Pattern: `^\d+$`}
	err := a.validateValue("test", "abc", rule)
	assert.Error(t, err)
}

func TestConfigAssistant_ValidateValue_Enum(t *testing.T) {
	a := NewConfigAssistant(nil)
	rule := &ValidationRule{EnumValues: []interface{}{"a", "b", "c"}}
	err := a.validateValue("test", "d", rule)
	assert.Error(t, err)
}

func TestConfigAssistant_ValidateValue_CustomFunc(t *testing.T) {
	a := NewConfigAssistant(nil)
	called := false
	rule := &ValidationRule{CustomFunc: func(v interface{}) error {
		called = true
		return assert.AnError
	}}
	err := a.validateValue("test", "value", rule)
	assert.True(t, called)
	assert.Error(t, err)
}

func TestConfigAssistant_ConvertValue_BoolTrue(t *testing.T) {
	a := NewConfigAssistant(nil)
	result := a.convertValue("true", "bool")
	assert.True(t, result.(bool))
}

func TestConfigAssistant_ConvertValue_BoolFalse(t *testing.T) {
	a := NewConfigAssistant(nil)
	result := a.convertValue("false", "bool")
	assert.False(t, result.(bool))
}

func TestConfigAssistant_ConvertValue_BoolOne(t *testing.T) {
	a := NewConfigAssistant(nil)
	result := a.convertValue("1", "auto")
	_ = result
}

func TestConfigAssistant_ConvertValue_BoolChineseYes(t *testing.T) {
	a := NewConfigAssistant(nil)
	result := a.convertValue("是", "auto")
	_ = result
}

func TestConfigAssistant_ConvertValue_AutoFloat(t *testing.T) {
	a := NewConfigAssistant(nil)
	result := a.convertValue("3.14", "auto")
	assert.InDelta(t, 3.14, result.(float64), 0.001)
}

func TestConfigAssistant_ConvertValue_AutoInt(t *testing.T) {
	a := NewConfigAssistant(nil)
	result := a.convertValue("42", "auto")
	assert.Equal(t, 42, result.(int))
}

func TestConfigAssistant_InferType_AllTypes(t *testing.T) {
	a := NewConfigAssistant(nil)
	assert.Equal(t, "int", a.inferType(42))
	assert.Equal(t, "float", a.inferType(3.14))
	assert.Equal(t, "bool", a.inferType(true))
	assert.Equal(t, "string", a.inferType("hello"))
	assert.Equal(t, "unknown", a.inferType([]int{}))
}

func TestConfigAssistant_MatchRuleConditions_Pattern(t *testing.T) {
	a := NewConfigAssistant(nil)
	rule := &ConfigRule{
		Conditions: []RuleCondition{{Type: "pattern", Value: `^test.*`, Operator: "regex"}},
	}
	assert.True(t, a.matchRuleConditions(rule, "test value"))
	assert.False(t, a.matchRuleConditions(rule, "other value"))
}

func TestConfigAssistant_DeduplicateSuggestions(t *testing.T) {
	a := NewConfigAssistant(nil)
	suggestions := []*ConfigSuggestion{
		{Key: "key1", SuggestedValue: "val1"},
		{Key: "key2", SuggestedValue: "val2"},
		{Key: "key1", SuggestedValue: "val3"},
	}
	deduped := a.deduplicateSuggestions(suggestions)
	assert.Len(t, deduped, 2)
}

func TestCov_ConfigItem_Struct(t *testing.T) {
	item := &ConfigItem{
		Key:          "test.key",
		Value:        "value",
		Type:         "string",
		DefaultValue: "default",
		Description:  "desc",
		Required:     true,
		Metadata:     map[string]interface{}{"meta": "data"},
	}
	assert.Equal(t, "test.key", item.Key)
	assert.True(t, item.Required)
}

func TestCov_ConfigSuggestion_Struct(t *testing.T) {
	s := &ConfigSuggestion{
		Key:            "k",
		CurrentValue:   "old",
		SuggestedValue: "new",
		Reason:         "reason",
		Confidence:     0.9,
		Impact:         "high",
		Category:       "cat",
	}
	assert.Equal(t, 0.9, s.Confidence)
}

func TestCov_ConfigValidationResult_Struct(t *testing.T) {
	r := &ConfigValidationResult{
		Valid: true,
		Errors: []ConfigError{{Key: "e", Message: "err"}},
		Warnings: []ConfigWarning{{Key: "w", Message: "warn"}},
	}
	assert.True(t, r.Valid)
	assert.Len(t, r.Errors, 1)
	assert.Len(t, r.Warnings, 1)
}

func TestParsedConfig_Struct(t *testing.T) {
	p := &ParsedConfig{
		Type:      ConfigSystem,
		Items:     map[string]*ConfigItem{"k": {Key: "k"}},
		Intent:    "intent text",
		Confidence: 0.85,
		RawText:   "raw",
	}
	assert.Equal(t, ConfigSystem, p.Type)
	assert.InDelta(t, 0.85, p.Confidence, 0.01)
}

func TestConfigTemplate_Struct(t *testing.T) {
	tmpl := &ConfigTemplate{
		TemplateID: "id",
		Name:       "name",
		Type:       ConfigDevice,
		Items:      map[string]*ConfigItem{},
		Conditions: []TemplateCondition{},
		Priority:   5,
	}
	assert.Equal(t, "id", tmpl.TemplateID)
}

func TestConfigRule_Struct(t *testing.T) {
	rule := &ConfigRule{
		RuleID:   "r1",
		Name:     "name",
		Type:     ConfigAlarm,
		Priority: 10,
	}
	assert.Equal(t, "r1", rule.RuleID)
}

func TestConfigConstants(t *testing.T) {
	assert.Equal(t, ConfigType("system"), ConfigSystem)
	assert.Equal(t, ConfigType("device"), ConfigDevice)
	assert.Equal(t, ConfigType("alarm"), ConfigAlarm)
	assert.Equal(t, ConfigType("collector"), ConfigCollector)
	assert.Equal(t, ConfigType("compute"), ConfigCompute)
	assert.Equal(t, ConfigType("storage"), ConfigStorage)
	assert.Equal(t, ConfigType("network"), ConfigNetwork)
	assert.Equal(t, ConfigType("security"), ConfigSecurity)
}
