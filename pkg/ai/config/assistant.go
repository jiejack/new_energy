package config

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// ConfigType 配置类型
type ConfigType string

const (
	ConfigSystem      ConfigType = "system"      // 系统配置
	ConfigDevice      ConfigType = "device"      // 设备配置
	ConfigAlarm       ConfigType = "alarm"       // 告警配置
	ConfigCollector   ConfigType = "collector"   // 采集配置
	ConfigCompute     ConfigType = "compute"     // 计算配置
	ConfigStorage     ConfigType = "storage"     // 存储配置
	ConfigNetwork     ConfigType = "network"     // 网络配置
	ConfigSecurity    ConfigType = "security"    // 安全配置
)

// ConfigItem 配置项
type ConfigItem struct {
	Key          string
	Value        interface{}
	Type         string // "string", "int", "float", "bool", "array", "map"
	DefaultValue interface{}
	Description  string
	Required     bool
	Validation   *ValidationRule
	Metadata     map[string]interface{}
}

// ValidationRule 验证规则
type ValidationRule struct {
	MinValue   interface{}
	MaxValue   interface{}
	MinLength  int
	MaxLength  int
	Pattern    string
	EnumValues []interface{}
	CustomFunc func(interface{}) error
}

// ConfigSuggestion 配置建议
type ConfigSuggestion struct {
	Key          string
	CurrentValue interface{}
	SuggestedValue interface{}
	Reason       string
	Confidence   float64
	Impact       string // "high", "medium", "low"
	Category     string
}

// ConfigValidationResult 配置验证结果
type ConfigValidationResult struct {
	Valid    bool
	Errors   []ConfigError
	Warnings []ConfigWarning
}

// ConfigError 配置错误
type ConfigError struct {
	Key     string
	Message string
	Value   interface{}
}

// ConfigWarning 配置警告
type ConfigWarning struct {
	Key     string
	Message string
	Value   interface{}
}

// ParsedConfig 解析后的配置
type ParsedConfig struct {
	Type      ConfigType
	Items     map[string]*ConfigItem
	Intent    string
	Confidence float64
	RawText   string
}

// ConfigTemplate 配置模板
type ConfigTemplate struct {
	TemplateID string
	Name       string
	Type       ConfigType
	Items      map[string]*ConfigItem
	Conditions []TemplateCondition
	Priority   int
}

// TemplateCondition 模板条件
type TemplateCondition struct {
	Key      string
	Operator string
	Value    interface{}
}

// ConfigAssistant 配置助手
type ConfigAssistant struct {
	templates    map[ConfigType][]*ConfigTemplate
	validators   map[string]*ValidationRule
	configRules  []*ConfigRule
	config       *AssistantConfig
	mu           sync.RWMutex
}

// ConfigRule 配置规则
type ConfigRule struct {
	RuleID     string
	Name       string
	Type       ConfigType
	Conditions []RuleCondition
	Actions    []RuleAction
	Priority   int
}

// RuleCondition 规则条件
type RuleCondition struct {
	Type     string // "text", "keyword", "pattern"
	Value    string
	Operator string
}

// RuleAction 规则动作
type RuleAction struct {
	Type       string // "suggest", "validate", "optimize"
	ConfigKey  string
	ConfigValue interface{}
	Message    string
}

// AssistantConfig 助手配置
type AssistantConfig struct {
	MinConfidence     float64
	MaxSuggestions    int
	EnableAutoSuggest bool
	EnableValidation  bool
	EnableOptimization bool
}

// DefaultAssistantConfig 默认助手配置
func DefaultAssistantConfig() *AssistantConfig {
	return &AssistantConfig{
		MinConfidence:      0.6,
		MaxSuggestions:     10,
		EnableAutoSuggest:  true,
		EnableValidation:   true,
		EnableOptimization: true,
	}
}

// NewConfigAssistant 创建配置助手
func NewConfigAssistant(config *AssistantConfig) *ConfigAssistant {
	if config == nil {
		config = DefaultAssistantConfig()
	}

	assistant := &ConfigAssistant{
		templates:   make(map[ConfigType][]*ConfigTemplate),
		validators:  make(map[string]*ValidationRule),
		configRules: make([]*ConfigRule, 0),
		config:      config,
	}

	// 初始化默认模板和规则
	assistant.initDefaultTemplates()
	assistant.initDefaultValidators()
	assistant.initDefaultRules()

	return assistant
}

// initDefaultTemplates 初始化默认配置模板
func (a *ConfigAssistant) initDefaultTemplates() {
	// 系统配置模板
	systemTemplates := []*ConfigTemplate{
		{
			TemplateID: "tpl_system_001",
			Name:       "基础系统配置",
			Type:       ConfigSystem,
			Items: map[string]*ConfigItem{
				"system.name": {
					Key:          "system.name",
					Type:         "string",
					DefaultValue: "新能源监控系统",
					Description:  "系统名称",
					Required:     true,
				},
				"system.log_level": {
					Key:          "system.log_level",
					Type:         "string",
					DefaultValue: "info",
					Description:  "日志级别",
					Required:     true,
					Validation: &ValidationRule{
						EnumValues: []interface{}{"debug", "info", "warn", "error"},
					},
				},
				"system.timezone": {
					Key:          "system.timezone",
					Type:         "string",
					DefaultValue: "Asia/Shanghai",
					Description:  "系统时区",
					Required:     true,
				},
			},
			Priority: 10,
		},
	}

	// 设备配置模板
	deviceTemplates := []*ConfigTemplate{
		{
			TemplateID: "tpl_device_001",
			Name:       "逆变器配置",
			Type:       ConfigDevice,
			Items: map[string]*ConfigItem{
				"device.type": {
					Key:          "device.type",
					Type:         "string",
					DefaultValue: "inverter",
					Description:  "设备类型",
					Required:     true,
				},
				"device.protocol": {
					Key:          "device.protocol",
					Type:         "string",
					DefaultValue: "modbus",
					Description:  "通信协议",
					Required:     true,
					Validation: &ValidationRule{
						EnumValues: []interface{}{"modbus", "iec104", "mqtt", "http"},
					},
				},
				"device.poll_interval": {
					Key:          "device.poll_interval",
					Type:         "int",
					DefaultValue: 5000,
					Description:  "采集间隔(毫秒)",
					Required:     true,
					Validation: &ValidationRule{
						MinValue: 1000,
						MaxValue: 60000,
					},
				},
				"device.timeout": {
					Key:          "device.timeout",
					Type:         "int",
					DefaultValue: 3000,
					Description:  "超时时间(毫秒)",
					Required:     true,
					Validation: &ValidationRule{
						MinValue: 1000,
						MaxValue: 30000,
					},
				},
			},
			Priority: 10,
		},
		{
			TemplateID: "tpl_device_002",
			Name:       "汇流箱配置",
			Type:       ConfigDevice,
			Items: map[string]*ConfigItem{
				"device.type": {
					Key:          "device.type",
					Type:         "string",
					DefaultValue: "combiner_box",
					Description:  "设备类型",
					Required:     true,
				},
				"device.channel_count": {
					Key:          "device.channel_count",
					Type:         "int",
					DefaultValue: 16,
					Description:  "通道数量",
					Required:     true,
					Validation: &ValidationRule{
						MinValue: 1,
						MaxValue: 64,
					},
				},
			},
			Priority: 9,
		},
	}

	// 告警配置模板
	alarmTemplates := []*ConfigTemplate{
		{
			TemplateID: "tpl_alarm_001",
			Name:       "告警规则配置",
			Type:       ConfigAlarm,
			Items: map[string]*ConfigItem{
				"alarm.enabled": {
					Key:          "alarm.enabled",
					Type:         "bool",
					DefaultValue: true,
					Description:  "是否启用告警",
					Required:     true,
				},
				"alarm.threshold_high": {
					Key:          "alarm.threshold_high",
					Type:         "float",
					DefaultValue: 100.0,
					Description:  "高限阈值",
					Required:     true,
					Validation: &ValidationRule{
						MinValue: 0.0,
						MaxValue: 1000.0,
					},
				},
				"alarm.threshold_low": {
					Key:          "alarm.threshold_low",
					Type:         "float",
					DefaultValue: 10.0,
					Description:  "低限阈值",
					Required:     true,
					Validation: &ValidationRule{
						MinValue: 0.0,
						MaxValue: 1000.0,
					},
				},
				"alarm.notification.email": {
					Key:          "alarm.notification.email",
					Type:         "bool",
					DefaultValue: true,
					Description:  "邮件通知",
					Required:     false,
				},
				"alarm.notification.sms": {
					Key:          "alarm.notification.sms",
					Type:         "bool",
					DefaultValue: false,
					Description:  "短信通知",
					Required:     false,
				},
			},
			Priority: 10,
		},
		{
			TemplateID: "tpl_alarm_002",
			Name:       "告警级别配置",
			Type:       ConfigAlarm,
			Items: map[string]*ConfigItem{
				"alarm.level.critical": {
					Key:          "alarm.level.critical",
					Type:         "string",
					DefaultValue: "紧急",
					Description:  "紧急告警级别",
					Required:     true,
				},
				"alarm.level.major": {
					Key:          "alarm.level.major",
					Type:         "string",
					DefaultValue: "重要",
					Description:  "重要告警级别",
					Required:     true,
				},
				"alarm.level.minor": {
					Key:          "alarm.level.minor",
					Type:         "string",
					DefaultValue: "次要",
					Description:  "次要告警级别",
					Required:     true,
				},
			},
			Priority: 9,
		},
	}

	// 采集配置模板
	collectorTemplates := []*ConfigTemplate{
		{
			TemplateID: "tpl_collector_001",
			Name:       "数据采集配置",
			Type:       ConfigCollector,
			Items: map[string]*ConfigItem{
				"collector.batch_size": {
					Key:          "collector.batch_size",
					Type:         "int",
					DefaultValue: 100,
					Description:  "批量采集大小",
					Required:     true,
					Validation: &ValidationRule{
						MinValue: 10,
						MaxValue: 1000,
					},
				},
				"collector.buffer_size": {
					Key:          "collector.buffer_size",
					Type:         "int",
					DefaultValue: 10000,
					Description:  "缓冲区大小",
					Required:     true,
					Validation: &ValidationRule{
						MinValue: 1000,
						MaxValue: 100000,
					},
				},
				"collector.retry_count": {
					Key:          "collector.retry_count",
					Type:         "int",
					DefaultValue: 3,
					Description:  "重试次数",
					Required:     true,
					Validation: &ValidationRule{
						MinValue: 0,
						MaxValue: 10,
					},
				},
			},
			Priority: 10,
		},
	}

	// 存储配置模板
	storageTemplates := []*ConfigTemplate{
		{
			TemplateID: "tpl_storage_001",
			Name:       "数据存储配置",
			Type:       ConfigStorage,
			Items: map[string]*ConfigItem{
				"storage.retention_days": {
					Key:          "storage.retention_days",
					Type:         "int",
					DefaultValue: 365,
					Description:  "数据保留天数",
					Required:     true,
					Validation: &ValidationRule{
						MinValue: 30,
						MaxValue: 3650,
					},
				},
				"storage.compression": {
					Key:          "storage.compression",
					Type:         "bool",
					DefaultValue: true,
					Description:  "是否启用压缩",
					Required:     false,
				},
				"storage.partition_size": {
					Key:          "storage.partition_size",
					Type:         "string",
					DefaultValue: "1D",
					Description:  "分区大小",
					Required:     true,
					Validation: &ValidationRule{
						EnumValues: []interface{}{"1H", "1D", "1W", "1M"},
					},
				},
			},
			Priority: 10,
		},
	}

	a.templates[ConfigSystem] = systemTemplates
	a.templates[ConfigDevice] = deviceTemplates
	a.templates[ConfigAlarm] = alarmTemplates
	a.templates[ConfigCollector] = collectorTemplates
	a.templates[ConfigStorage] = storageTemplates
}

// initDefaultValidators 初始化默认验证器
func (a *ConfigAssistant) initDefaultValidators() {
	// IP地址验证
	a.validators["ip_address"] = &ValidationRule{
		Pattern: `^(\d{1,3}\.){3}\d{1,3}$`,
	}

	// 端口号验证
	a.validators["port"] = &ValidationRule{
		MinValue: 1,
		MaxValue: 65535,
	}

	// URL验证
	a.validators["url"] = &ValidationRule{
		Pattern: `^https?://[^\s]+$`,
	}

	// 邮箱验证
	a.validators["email"] = &ValidationRule{
		Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
	}

	// 时间间隔验证
	a.validators["interval"] = &ValidationRule{
		MinValue: 100,
		MaxValue: 86400000,
	}
}

// initDefaultRules 初始化默认配置规则
func (a *ConfigAssistant) initDefaultRules() {
	// 高性能配置建议规则
	a.configRules = append(a.configRules, &ConfigRule{
		RuleID: "rule_perf_001",
		Name:   "高性能配置建议",
		Type:   ConfigSystem,
		Conditions: []RuleCondition{
			{Type: "keyword", Value: "高性能", Operator: "contains"},
			{Type: "keyword", Value: "高吞吐", Operator: "contains"},
		},
		Actions: []RuleAction{
			{Type: "suggest", ConfigKey: "collector.batch_size", ConfigValue: 500, Message: "建议增加批量采集大小以提高吞吐量"},
			{Type: "suggest", ConfigKey: "collector.buffer_size", ConfigValue: 50000, Message: "建议增加缓冲区大小以处理高并发"},
		},
		Priority: 10,
	})

	// 低功耗配置建议规则
	a.configRules = append(a.configRules, &ConfigRule{
		RuleID: "rule_power_001",
		Name:   "低功耗配置建议",
		Type:   ConfigSystem,
		Conditions: []RuleCondition{
			{Type: "keyword", Value: "低功耗", Operator: "contains"},
			{Type: "keyword", Value: "节能", Operator: "contains"},
		},
		Actions: []RuleAction{
			{Type: "suggest", ConfigKey: "device.poll_interval", ConfigValue: 10000, Message: "建议增加采集间隔以降低功耗"},
			{Type: "suggest", ConfigKey: "collector.batch_size", ConfigValue: 50, Message: "建议减少批量大小以降低内存占用"},
		},
		Priority: 10,
	})

	// 安全配置建议规则
	a.configRules = append(a.configRules, &ConfigRule{
		RuleID: "rule_security_001",
		Name:   "安全配置建议",
		Type:   ConfigSecurity,
		Conditions: []RuleCondition{
			{Type: "keyword", Value: "安全", Operator: "contains"},
			{Type: "keyword", Value: "加密", Operator: "contains"},
		},
		Actions: []RuleAction{
			{Type: "suggest", ConfigKey: "security.encryption", ConfigValue: true, Message: "建议启用数据加密"},
			{Type: "suggest", ConfigKey: "security.auth_enabled", ConfigValue: true, Message: "建议启用身份认证"},
		},
		Priority: 10,
	})
}

// ParseConfig 解析自然语言配置
func (a *ConfigAssistant) ParseConfig(ctx context.Context, text string) (*ParsedConfig, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// 识别配置类型
	configType := a.recognizeConfigType(text)

	// 提取配置项
	items := a.extractConfigItems(text, configType)

	// 计算置信度
	confidence := a.calculateConfidence(text, configType, items)

	return &ParsedConfig{
		Type:       configType,
		Items:      items,
		Intent:     text,
		Confidence: confidence,
		RawText:    text,
	}, nil
}

// recognizeConfigType 识别配置类型
func (a *ConfigAssistant) recognizeConfigType(text string) ConfigType {
	keywords := map[ConfigType][]string{
		ConfigSystem:    {"系统", "全局", "基础配置"},
		ConfigDevice:    {"设备", "逆变器", "汇流箱", "变压器", "采集器"},
		ConfigAlarm:     {"告警", "报警", "预警", "阈值", "通知"},
		ConfigCollector: {"采集", "数据采集", "采集频率", "采集间隔"},
		ConfigCompute:   {"计算", "公式", "统计", "分析"},
		ConfigStorage:   {"存储", "数据库", "保留", "归档"},
		ConfigNetwork:   {"网络", "通信", "连接", "协议"},
		ConfigSecurity:  {"安全", "认证", "加密", "权限"},
	}

	scores := make(map[ConfigType]int)
	lowerText := strings.ToLower(text)

	for configType, words := range keywords {
		for _, word := range words {
			if strings.Contains(lowerText, strings.ToLower(word)) {
				scores[configType]++
			}
		}
	}

	// 找出得分最高的类型
	var maxScore int
	var result ConfigType = ConfigSystem

	for configType, score := range scores {
		if score > maxScore {
			maxScore = score
			result = configType
		}
	}

	return result
}

// extractConfigItems 提取配置项
func (a *ConfigAssistant) extractConfigItems(text string, configType ConfigType) map[string]*ConfigItem {
	items := make(map[string]*ConfigItem)

	// 获取匹配的模板
	templates := a.getMatchingTemplates(configType, text)

	// 从模板中提取配置项
	for _, template := range templates {
		for key, item := range template.Items {
			// 尝试从文本中提取值
			value := a.extractValue(text, key, item)
			if value != nil {
				newItem := *item
				newItem.Value = value
				items[key] = &newItem
			}
		}
	}

	// 提取键值对
	kvPairs := a.extractKeyValuePairs(text)
	for key, value := range kvPairs {
		if _, exists := items[key]; !exists {
			items[key] = &ConfigItem{
				Key:   key,
				Value: value,
				Type:  a.inferType(value),
			}
		}
	}

	return items
}

// getMatchingTemplates 获取匹配的模板
func (a *ConfigAssistant) getMatchingTemplates(configType ConfigType, text string) []*ConfigTemplate {
	a.mu.RLock()
	defer a.mu.RUnlock()

	templates, exists := a.templates[configType]
	if !exists {
		return nil
	}

	// 简单匹配：返回所有该类型的模板
	result := make([]*ConfigTemplate, len(templates))
	copy(result, templates)
	return result
}

// extractValue 提取配置值
func (a *ConfigAssistant) extractValue(text string, key string, item *ConfigItem) interface{} {
	// 构建匹配模式
	patterns := []string{
		fmt.Sprintf(`%s[为是:：]\s*(\S+)`, regexp.QuoteMeta(key)),
		fmt.Sprintf(`%s\s*[=：]\s*(\S+)`, regexp.QuoteMeta(key)),
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return a.convertValue(matches[1], item.Type)
		}
	}

	return nil
}

// extractKeyValuePairs 提取键值对
func (a *ConfigAssistant) extractKeyValuePairs(text string) map[string]interface{} {
	pairs := make(map[string]interface{})

	// 匹配 "key=value" 或 "key:value" 格式
	patterns := []string{
		`(\w+(?:\.\w+)*)\s*[=：]\s*(\S+)`,
		`(\w+(?:\.\w+)*)\s*[为是]\s*(\S+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				key := match[1]
				value := match[2]
				pairs[key] = a.convertValue(value, "auto")
			}
		}
	}

	return pairs
}

// convertValue 转换值类型
func (a *ConfigAssistant) convertValue(value string, targetType string) interface{} {
	switch targetType {
	case "int":
		var result int
		fmt.Sscanf(value, "%d", &result)
		return result
	case "float":
		var result float64
		fmt.Sscanf(value, "%f", &result)
		return result
	case "bool":
		return strings.ToLower(value) == "true" || value == "1" || strings.ToLower(value) == "是" || strings.ToLower(value) == "启用"
	case "string":
		return value
	case "auto":
		// 自动推断类型
		if strings.ToLower(value) == "true" || strings.ToLower(value) == "false" {
			return strings.ToLower(value) == "true"
		}
		if strings.Contains(value, ".") {
			var result float64
			if _, err := fmt.Sscanf(value, "%f", &result); err == nil {
				return result
			}
		}
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
		return value
	default:
		return value
	}
}

// inferType 推断类型
func (a *ConfigAssistant) inferType(value interface{}) string {
	switch value.(type) {
	case int:
		return "int"
	case float64:
		return "float"
	case bool:
		return "bool"
	case string:
		return "string"
	default:
		return "unknown"
	}
}

// calculateConfidence 计算置信度
func (a *ConfigAssistant) calculateConfidence(text string, configType ConfigType, items map[string]*ConfigItem) float64 {
	confidence := 0.5

	// 根据配置项数量调整
	if len(items) > 0 {
		confidence += 0.1 * float64(len(items))
		if confidence > 0.9 {
			confidence = 0.9
		}
	}

	// 根据关键词匹配调整
	keywords := map[ConfigType][]string{
		ConfigSystem:    {"配置", "设置"},
		ConfigDevice:    {"设备", "配置"},
		ConfigAlarm:     {"告警", "阈值"},
		ConfigCollector: {"采集", "频率"},
	}

	if words, exists := keywords[configType]; exists {
		for _, word := range words {
			if strings.Contains(text, word) {
				confidence += 0.05
			}
		}
	}

	// 限制在0-1之间
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// GenerateSuggestions 生成配置建议
func (a *ConfigAssistant) GenerateSuggestions(ctx context.Context, parsed *ParsedConfig, currentConfig map[string]interface{}) ([]*ConfigSuggestion, error) {
	suggestions := make([]*ConfigSuggestion, 0)

	// 应用配置规则
	for _, rule := range a.configRules {
		if rule.Type == parsed.Type {
			if a.matchRuleConditions(rule, parsed.RawText) {
				for _, action := range rule.Actions {
					if action.Type == "suggest" {
						currentValue := currentConfig[action.ConfigKey]
						suggestion := &ConfigSuggestion{
							Key:            action.ConfigKey,
							CurrentValue:   currentValue,
							SuggestedValue: action.ConfigValue,
							Reason:         action.Message,
							Confidence:     0.8,
							Impact:         "medium",
							Category:       string(parsed.Type),
						}
						suggestions = append(suggestions, suggestion)
					}
				}
			}
		}
	}

	// 基于最佳实践生成建议
	bestPracticeSuggestions := a.generateBestPracticeSuggestions(parsed, currentConfig)
	suggestions = append(suggestions, bestPracticeSuggestions...)

	// 限制建议数量
	if len(suggestions) > a.config.MaxSuggestions {
		suggestions = suggestions[:a.config.MaxSuggestions]
	}

	return suggestions, nil
}

// matchRuleConditions 匹配规则条件
func (a *ConfigAssistant) matchRuleConditions(rule *ConfigRule, text string) bool {
	lowerText := strings.ToLower(text)

	for _, condition := range rule.Conditions {
		switch condition.Type {
		case "keyword":
			if condition.Operator == "contains" {
				if !strings.Contains(lowerText, strings.ToLower(condition.Value)) {
					return false
				}
			}
		case "pattern":
			matched, _ := regexp.MatchString(condition.Value, text)
			if !matched {
				return false
			}
		}
	}

	return true
}

// generateBestPracticeSuggestions 生成最佳实践建议
func (a *ConfigAssistant) generateBestPracticeSuggestions(parsed *ParsedConfig, currentConfig map[string]interface{}) []*ConfigSuggestion {
	suggestions := make([]*ConfigSuggestion, 0)

	// 根据配置类型生成不同的最佳实践建议
	switch parsed.Type {
	case ConfigDevice:
		// 设备配置最佳实践
		if pollInterval, exists := currentConfig["device.poll_interval"]; exists {
			if interval, ok := pollInterval.(int); ok && interval < 1000 {
				suggestions = append(suggestions, &ConfigSuggestion{
					Key:            "device.poll_interval",
					CurrentValue:   pollInterval,
					SuggestedValue: 1000,
					Reason:         "采集间隔过小可能导致设备负载过高，建议最小设置为1000ms",
					Confidence:     0.9,
					Impact:         "high",
					Category:       "performance",
				})
			}
		}

	case ConfigAlarm:
		// 告警配置最佳实践
		if highThreshold, exists := currentConfig["alarm.threshold_high"]; exists {
			if lowThreshold, exists := currentConfig["alarm.threshold_low"]; exists {
				if high, ok := highThreshold.(float64); ok {
					if low, ok := lowThreshold.(float64); ok {
						if high <= low {
							suggestions = append(suggestions, &ConfigSuggestion{
								Key:            "alarm.threshold_high",
								CurrentValue:   highThreshold,
								SuggestedValue: low + 10,
								Reason:         "高限阈值应大于低限阈值",
								Confidence:     1.0,
								Impact:         "high",
								Category:       "validation",
							})
						}
					}
				}
			}
		}

	case ConfigCollector:
		// 采集配置最佳实践
		if batchSize, exists := currentConfig["collector.batch_size"]; exists {
			if size, ok := batchSize.(int); ok && size > 500 {
				suggestions = append(suggestions, &ConfigSuggestion{
					Key:            "collector.batch_size",
					CurrentValue:   batchSize,
					SuggestedValue: 500,
					Reason:         "批量大小过大可能导致内存占用过高，建议不超过500",
					Confidence:     0.85,
					Impact:         "medium",
					Category:       "performance",
				})
			}
		}

	case ConfigStorage:
		// 存储配置最佳实践
		if retention, exists := currentConfig["storage.retention_days"]; exists {
			if days, ok := retention.(int); ok && days < 90 {
				suggestions = append(suggestions, &ConfigSuggestion{
					Key:            "storage.retention_days",
					CurrentValue:   retention,
					SuggestedValue: 90,
					Reason:         "数据保留时间过短可能影响历史数据分析，建议至少保留90天",
					Confidence:     0.8,
					Impact:         "medium",
					Category:       "data_integrity",
				})
			}
		}
	}

	return suggestions
}

// ValidateConfig 验证配置
func (a *ConfigAssistant) ValidateConfig(ctx context.Context, config map[string]interface{}) (*ConfigValidationResult, error) {
	result := &ConfigValidationResult{
		Valid:    true,
		Errors:   make([]ConfigError, 0),
		Warnings: make([]ConfigWarning, 0),
	}

	// 验证每个配置项
	for key, value := range config {
		// 查找对应的验证规则
		if validator, exists := a.validators[key]; exists {
			if err := a.validateValue(key, value, validator); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, ConfigError{
					Key:     key,
					Message: err.Error(),
					Value:   value,
				})
			}
		}

		// 检查是否为空值
		if value == nil || value == "" {
			result.Warnings = append(result.Warnings, ConfigWarning{
				Key:     key,
				Message: "配置值为空",
				Value:   value,
			})
		}
	}

	return result, nil
}

// validateValue 验证值
func (a *ConfigAssistant) validateValue(key string, value interface{}, rule *ValidationRule) error {
	// 最小值验证
	if rule.MinValue != nil {
		if min, ok := rule.MinValue.(int); ok {
			if v, ok := value.(int); ok && v < min {
				return fmt.Errorf("value %d is less than minimum %d", v, min)
			}
		}
		if min, ok := rule.MinValue.(float64); ok {
			if v, ok := value.(float64); ok && v < min {
				return fmt.Errorf("value %f is less than minimum %f", v, min)
			}
		}
	}

	// 最大值验证
	if rule.MaxValue != nil {
		if max, ok := rule.MaxValue.(int); ok {
			if v, ok := value.(int); ok && v > max {
				return fmt.Errorf("value %d is greater than maximum %d", v, max)
			}
		}
		if max, ok := rule.MaxValue.(float64); ok {
			if v, ok := value.(float64); ok && v > max {
				return fmt.Errorf("value %f is greater than maximum %f", v, max)
			}
		}
	}

	// 正则表达式验证
	if rule.Pattern != "" {
		if str, ok := value.(string); ok {
			matched, err := regexp.MatchString(rule.Pattern, str)
			if err != nil {
				return fmt.Errorf("pattern match error: %w", err)
			}
			if !matched {
				return fmt.Errorf("value '%s' does not match pattern '%s'", str, rule.Pattern)
			}
		}
	}

	// 枚举值验证
	if len(rule.EnumValues) > 0 {
		found := false
		for _, enumValue := range rule.EnumValues {
			if value == enumValue {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value %v is not in enum values %v", value, rule.EnumValues)
		}
	}

	// 自定义验证
	if rule.CustomFunc != nil {
		return rule.CustomFunc(value)
	}

	return nil
}

// OptimizeConfig 优化配置
func (a *ConfigAssistant) OptimizeConfig(ctx context.Context, config map[string]interface{}, optimizationType string) ([]*ConfigSuggestion, error) {
	suggestions := make([]*ConfigSuggestion, 0)

	switch optimizationType {
	case "performance":
		suggestions = a.optimizeForPerformance(config)
	case "reliability":
		suggestions = a.optimizeForReliability(config)
	case "cost":
		suggestions = a.optimizeForCost(config)
	default:
		// 综合优化
		suggestions = append(suggestions, a.optimizeForPerformance(config)...)
		suggestions = append(suggestions, a.optimizeForReliability(config)...)
	}

	// 去重
	suggestions = a.deduplicateSuggestions(suggestions)

	// 限制数量
	if len(suggestions) > a.config.MaxSuggestions {
		suggestions = suggestions[:a.config.MaxSuggestions]
	}

	return suggestions, nil
}

// optimizeForPerformance 性能优化
func (a *ConfigAssistant) optimizeForPerformance(config map[string]interface{}) []*ConfigSuggestion {
	suggestions := make([]*ConfigSuggestion, 0)

	// 采集性能优化
	if batchSize, exists := config["collector.batch_size"]; exists {
		if size, ok := batchSize.(int); ok && size < 200 {
			suggestions = append(suggestions, &ConfigSuggestion{
				Key:            "collector.batch_size",
				CurrentValue:   batchSize,
				SuggestedValue: 200,
				Reason:         "增加批量大小可以提高数据吞吐量",
				Confidence:     0.85,
				Impact:         "high",
				Category:       "performance",
			})
		}
	}

	// 缓冲区优化
	if bufferSize, exists := config["collector.buffer_size"]; exists {
		if size, ok := bufferSize.(int); ok && size < 20000 {
			suggestions = append(suggestions, &ConfigSuggestion{
				Key:            "collector.buffer_size",
				CurrentValue:   bufferSize,
				SuggestedValue: 20000,
				Reason:         "增加缓冲区大小可以处理更高的并发",
				Confidence:     0.8,
				Impact:         "medium",
				Category:       "performance",
			})
		}
	}

	return suggestions
}

// optimizeForReliability 可靠性优化
func (a *ConfigAssistant) optimizeForReliability(config map[string]interface{}) []*ConfigSuggestion {
	suggestions := make([]*ConfigSuggestion, 0)

	// 重试次数优化
	if retryCount, exists := config["collector.retry_count"]; exists {
		if count, ok := retryCount.(int); ok && count < 3 {
			suggestions = append(suggestions, &ConfigSuggestion{
				Key:            "collector.retry_count",
				CurrentValue:   retryCount,
				SuggestedValue: 3,
				Reason:         "增加重试次数可以提高数据采集的可靠性",
				Confidence:     0.9,
				Impact:         "high",
				Category:       "reliability",
			})
		}
	}

	// 超时时间优化
	if timeout, exists := config["device.timeout"]; exists {
		if t, ok := timeout.(int); ok && t < 5000 {
			suggestions = append(suggestions, &ConfigSuggestion{
				Key:            "device.timeout",
				CurrentValue:   timeout,
				SuggestedValue: 5000,
				Reason:         "增加超时时间可以减少因网络波动导致的失败",
				Confidence:     0.85,
				Impact:         "medium",
				Category:       "reliability",
			})
		}
	}

	return suggestions
}

// optimizeForCost 成本优化
func (a *ConfigAssistant) optimizeForCost(config map[string]interface{}) []*ConfigSuggestion {
	suggestions := make([]*ConfigSuggestion, 0)

	// 存储成本优化
	if retention, exists := config["storage.retention_days"]; exists {
		if days, ok := retention.(int); ok && days > 180 {
			suggestions = append(suggestions, &ConfigSuggestion{
				Key:            "storage.retention_days",
				CurrentValue:   retention,
				SuggestedValue: 180,
				Reason:         "减少数据保留时间可以降低存储成本",
				Confidence:     0.75,
				Impact:         "medium",
				Category:       "cost",
			})
		}
	}

	// 采集频率优化
	if pollInterval, exists := config["device.poll_interval"]; exists {
		if interval, ok := pollInterval.(int); ok && interval < 3000 {
			suggestions = append(suggestions, &ConfigSuggestion{
				Key:            "device.poll_interval",
				CurrentValue:   pollInterval,
				SuggestedValue: 3000,
				Reason:         "降低采集频率可以减少计算和网络成本",
				Confidence:     0.7,
				Impact:         "low",
				Category:       "cost",
			})
		}
	}

	return suggestions
}

// deduplicateSuggestions 去重建议
func (a *ConfigAssistant) deduplicateSuggestions(suggestions []*ConfigSuggestion) []*ConfigSuggestion {
	seen := make(map[string]bool)
	result := make([]*ConfigSuggestion, 0)

	for _, suggestion := range suggestions {
		key := suggestion.Key
		if !seen[key] {
			seen[key] = true
			result = append(result, suggestion)
		}
	}

	return result
}

// AddTemplate 添加配置模板
func (a *ConfigAssistant) AddTemplate(template *ConfigTemplate) error {
	if template == nil {
		return fmt.Errorf("template cannot be nil")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.templates[template.Type] = append(a.templates[template.Type], template)
	return nil
}

// AddValidator 添加验证器
func (a *ConfigAssistant) AddValidator(key string, rule *ValidationRule) error {
	if rule == nil {
		return fmt.Errorf("rule cannot be nil")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.validators[key] = rule
	return nil
}

// AddRule 添加配置规则
func (a *ConfigAssistant) AddRule(rule *ConfigRule) error {
	if rule == nil {
		return fmt.Errorf("rule cannot be nil")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.configRules = append(a.configRules, rule)
	return nil
}

// GetTemplates 获取模板
func (a *ConfigAssistant) GetTemplates(configType ConfigType) []*ConfigTemplate {
	a.mu.RLock()
	defer a.mu.RUnlock()

	templates, exists := a.templates[configType]
	if !exists {
		return nil
	}

	result := make([]*ConfigTemplate, len(templates))
	copy(result, templates)
	return result
}

// GetValidators 获取验证器
func (a *ConfigAssistant) GetValidators() map[string]*ValidationRule {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]*ValidationRule)
	for k, v := range a.validators {
		result[k] = v
	}
	return result
}

// GenerateConfigFromTemplate 从模板生成配置
func (a *ConfigAssistant) GenerateConfigFromTemplate(templateID string) (map[string]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 查找模板
	for _, templates := range a.templates {
		for _, template := range templates {
			if template.TemplateID == templateID {
				config := make(map[string]interface{})
				for key, item := range template.Items {
					config[key] = item.DefaultValue
				}
				return config, nil
			}
		}
	}

	return nil, fmt.Errorf("template not found: %s", templateID)
}

// ExplainConfig 解释配置
func (a *ConfigAssistant) ExplainConfig(key string) (string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 在模板中查找配置项
	for _, templates := range a.templates {
		for _, template := range templates {
			if item, exists := template.Items[key]; exists {
				return fmt.Sprintf("%s\n类型: %s\n默认值: %v\n是否必填: %v",
					item.Description, item.Type, item.DefaultValue, item.Required), nil
			}
		}
	}

	return "", fmt.Errorf("config key not found: %s", key)
}
