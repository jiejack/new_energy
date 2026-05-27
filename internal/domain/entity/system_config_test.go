package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSystemConfig(t *testing.T) {
	config := NewSystemConfig("basic", "system_name", "NEM System", SystemConfigValueTypeString, "System name")
	assert.NotEmpty(t, config.ID)
	assert.Equal(t, "basic", config.Category)
	assert.Equal(t, "system_name", config.Key)
	assert.Equal(t, "NEM System", config.Value)
	assert.Equal(t, SystemConfigValueTypeString, config.ValueType)
	assert.Equal(t, "System name", config.Description)
	assert.NotZero(t, config.CreatedAt)
	assert.NotZero(t, config.UpdatedAt)
}

func TestSystemConfig_SetValue(t *testing.T) {
	config := NewSystemConfig("basic", "key1", "val1", SystemConfigValueTypeString, "")
	config.SetValue("val2", SystemConfigValueTypeInt)
	assert.Equal(t, "val2", config.Value)
	assert.Equal(t, SystemConfigValueTypeInt, config.ValueType)
}

func TestSystemConfig_SetDescription(t *testing.T) {
	config := NewSystemConfig("basic", "key1", "val1", SystemConfigValueTypeString, "")
	config.SetDescription("New description")
	assert.Equal(t, "New description", config.Description)
}

func TestSystemConfig_IsString(t *testing.T) {
	config := NewSystemConfig("basic", "key1", "val1", SystemConfigValueTypeString, "")
	assert.True(t, config.IsString())
	assert.False(t, config.IsInt())
	assert.False(t, config.IsBool())
	assert.False(t, config.IsJSON())
}

func TestSystemConfig_IsInt(t *testing.T) {
	config := NewSystemConfig("basic", "key1", "100", SystemConfigValueTypeInt, "")
	assert.True(t, config.IsInt())
	assert.False(t, config.IsString())
}

func TestSystemConfig_IsBool(t *testing.T) {
	config := NewSystemConfig("basic", "key1", "true", SystemConfigValueTypeBool, "")
	assert.True(t, config.IsBool())
}

func TestSystemConfig_IsJSON(t *testing.T) {
	config := NewSystemConfig("basic", "key1", "{}", SystemConfigValueTypeJSON, "")
	assert.True(t, config.IsJSON())
}

func TestSystemConfig_TableName(t *testing.T) {
	config := SystemConfig{}
	assert.Equal(t, "system_configs", config.TableName())
}

func TestSystemConfigValueType_Constants(t *testing.T) {
	assert.Equal(t, SystemConfigValueType("string"), SystemConfigValueTypeString)
	assert.Equal(t, SystemConfigValueType("int"), SystemConfigValueTypeInt)
	assert.Equal(t, SystemConfigValueType("bool"), SystemConfigValueTypeBool)
	assert.Equal(t, SystemConfigValueType("json"), SystemConfigValueTypeJSON)
}

func TestSystemConfigFilter_Fields(t *testing.T) {
	cat := "basic"
	key := "system_name"
	filter := SystemConfigFilter{
		Category: &cat,
		Key:      &key,
		Page:     1,
		PageSize: 10,
	}
	assert.Equal(t, "basic", *filter.Category)
	assert.Equal(t, "system_name", *filter.Key)
	assert.Equal(t, 1, filter.Page)
	assert.Equal(t, 10, filter.PageSize)
}
