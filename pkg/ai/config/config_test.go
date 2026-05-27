package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigType_Constants(t *testing.T) {
	assert.Equal(t, ConfigType("system"), ConfigSystem)
	assert.Equal(t, ConfigType("device"), ConfigDevice)
	assert.Equal(t, ConfigType("alarm"), ConfigAlarm)
	assert.Equal(t, ConfigType("collector"), ConfigCollector)
	assert.Equal(t, ConfigType("compute"), ConfigCompute)
	assert.Equal(t, ConfigType("storage"), ConfigStorage)
	assert.Equal(t, ConfigType("network"), ConfigNetwork)
	assert.Equal(t, ConfigType("security"), ConfigSecurity)
}

func TestConfigItem_Struct(t *testing.T) {
	item := ConfigItem{
		Key:          "max_devices",
		Value:        1000,
		Type:         "int",
		DefaultValue: 500,
		Description:  "Maximum number of devices",
		Required:     true,
		Validation: &ValidationRule{
			MinValue: 1,
			MaxValue: 10000,
		},
	}
	assert.Equal(t, "max_devices", item.Key)
	assert.Equal(t, 1000, item.Value)
	assert.True(t, item.Required)
}

func TestValidationRule_Struct(t *testing.T) {
	rule := &ValidationRule{
		MinValue:   0,
		MaxValue:   100,
		MinLength:  1,
		MaxLength:  255,
		Pattern:    "^[a-zA-Z0-9]+$",
		EnumValues: []interface{}{"active", "inactive"},
	}
	assert.Equal(t, 0, rule.MinValue)
	assert.Equal(t, 255, rule.MaxLength)
}

func TestConfigSuggestion_Struct(t *testing.T) {
	suggestion := ConfigSuggestion{
		Key:            "batch_size",
		CurrentValue:   100,
		SuggestedValue: 500,
		Reason:         "Increasing batch size improves throughput",
		Confidence:     0.85,
		Impact:         "medium",
		Category:       "performance",
	}
	assert.Equal(t, "batch_size", suggestion.Key)
	assert.Equal(t, 0.85, suggestion.Confidence)
}

func TestConfigValidationResult_Struct(t *testing.T) {
	result := ConfigValidationResult{
		Valid: true,
		Errors:   []ConfigError{},
		Warnings: []ConfigWarning{{Key: "timeout", Message: "Value is high", Value: 300}},
	}
	assert.True(t, result.Valid)
	assert.Equal(t, 1, len(result.Warnings))
}

func TestConfigError_Struct(t *testing.T) {
	err := ConfigError{
		Key:     "port",
		Message: "Port must be between 1 and 65535",
		Value:   0,
	}
	assert.Equal(t, "port", err.Key)
}

func TestConfigWarning_Struct(t *testing.T) {
	warn := ConfigWarning{
		Key:     "timeout",
		Message: "Timeout is very high",
		Value:   3600,
	}
	assert.Equal(t, "timeout", warn.Key)
}

func TestNewConfigAssistant(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	require.NotNil(t, assistant)
}

func TestConfigAssistant_ParseConfig(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	parsed, err := assistant.ParseConfig(ctx, "配置系统参数")
	require.NoError(t, err)
	assert.NotNil(t, parsed)
}

func TestConfigAssistant_GenerateSuggestions(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	parsed, err := assistant.ParseConfig(ctx, "高性能配置")
	require.NoError(t, err)

	suggestions, err := assistant.GenerateSuggestions(ctx, parsed, map[string]interface{}{})
	require.NoError(t, err)
	assert.NotNil(t, suggestions)
}

func TestConfigAssistant_ValidateConfig(t *testing.T) {
	assistant := NewConfigAssistant(nil)
	ctx := context.Background()

	config := map[string]interface{}{
		"port":    8080,
		"host":    "localhost",
		"enabled": true,
	}

	result, err := assistant.ValidateConfig(ctx, config)
	require.NoError(t, err)
	assert.NotNil(t, result)
}
