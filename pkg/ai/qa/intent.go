package qa

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// IntentType 意图类型
type IntentType string

const (
	IntentQuery     IntentType = "query"     // 查询意图
	IntentControl   IntentType = "control"   // 控制意图
	IntentConfig    IntentType = "config"    // 配置意图
	IntentDiagnose  IntentType = "diagnose"  // 诊断意图
	IntentUnknown   IntentType = "unknown"   // 未知意图
)

// EntityType 实体类型
type EntityType string

const (
	EntityDevice    EntityType = "device"    // 设备实体
	EntityPoint     EntityType = "point"     // 测点实体
	EntityStation   EntityType = "station"   // 电站实体
	EntityTime      EntityType = "time"      // 时间实体
	EntityMetric    EntityType = "metric"    // 指标实体
	EntityThreshold EntityType = "threshold" // 阈值实体
	EntityStatus    EntityType = "status"    // 状态实体
)

// Entity 识别的实体
type Entity struct {
	Type       EntityType
	Value      string
	Normalized string      // 标准化后的值
	Position   Position    // 在文本中的位置
	Metadata   interface{} // 附加元数据
}

// Position 位置信息
type Position struct {
	Start int
	End   int
}

// Slot 槽位
type Slot struct {
	Name      string
	Value     interface{}
	Required  bool
	Filled    bool
	Entities  []Entity // 关联的实体
	Validator SlotValidator
}

// SlotValidator 槽位验证器
type SlotValidator interface {
	Validate(value interface{}) error
}

// Intent 意图识别结果
type Intent struct {
	Type        IntentType
	Name        string
	Confidence  float64
	Entities    []Entity
	Slots       map[string]*Slot
	SubIntents  []*Intent // 子意图
	RawText     string
	Timestamp   time.Time
}

// IntentPattern 意图模式
type IntentPattern struct {
	IntentType   IntentType
	IntentName   string
	Patterns     []*regexp.Regexp
	Keywords     []string
	EntityTypes  []EntityType
	SlotDefs     map[string]*SlotDefinition
	Priority     int
}

// SlotDefinition 槽位定义
type SlotDefinition struct {
	Name       string
	Type       string
	Required   bool
	Prompts    []string // 填充提示语
	Default    interface{}
	EntityType EntityType // 关联的实体类型
}

// IntentRecognizer 意图识别器
type IntentRecognizer struct {
	patterns    map[IntentType][]*IntentPattern
	entityRules map[EntityType][]*EntityRule
	config      *RecognizerConfig
	mu          sync.RWMutex
}

// EntityRule 实体识别规则
type EntityRule struct {
	Type       EntityType
	Pattern    *regexp.Regexp
	Dictionary []string
	Normalizer func(string) string
}

// RecognizerConfig 识别器配置
type RecognizerConfig struct {
	MinConfidence     float64
	MaxEntities       int
	EnableSubIntents  bool
	CacheEnabled      bool
	CacheTTL          time.Duration
}

// DefaultRecognizerConfig 默认识别器配置
func DefaultRecognizerConfig() *RecognizerConfig {
	return &RecognizerConfig{
		MinConfidence:    0.6,
		MaxEntities:      20,
		EnableSubIntents: true,
		CacheEnabled:     true,
		CacheTTL:         5 * time.Minute,
	}
}

// NewIntentRecognizer 创建意图识别器
func NewIntentRecognizer(config *RecognizerConfig) *IntentRecognizer {
	if config == nil {
		config = DefaultRecognizerConfig()
	}

	recognizer := &IntentRecognizer{
		patterns:    make(map[IntentType][]*IntentPattern),
		entityRules: make(map[EntityType][]*EntityRule),
		config:      config,
	}

	// 初始化默认规则
	recognizer.initDefaultPatterns()
	recognizer.initDefaultEntityRules()

	return recognizer
}

// initDefaultPatterns 初始化默认意图模式
func (r *IntentRecognizer) initDefaultPatterns() {
	// 查询意图模式
	queryPatterns := []*IntentPattern{
		{
			IntentType: IntentQuery,
			IntentName: "query_realtime",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)(查询|查看|获取|显示).*(实时|当前|最新).*(数据|状态|值)`),
				regexp.MustCompile(`(?i)(实时|当前).*(数据|状态|值).*(查询|查看|获取)`),
			},
			Keywords:    []string{"查询", "查看", "获取", "显示", "实时", "当前", "最新"},
			EntityTypes: []EntityType{EntityDevice, EntityPoint, EntityStation},
			SlotDefs: map[string]*SlotDefinition{
				"target": {Name: "target", Type: "string", Required: true, EntityType: EntityDevice},
				"metric": {Name: "metric", Type: "string", Required: false, EntityType: EntityMetric},
			},
			Priority: 10,
		},
		{
			IntentType: IntentQuery,
			IntentName: "query_history",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)(查询|查看|获取).*(历史|过去|昨天|上周|上月).*(数据|记录)`),
				regexp.MustCompile(`(?i)(历史|过去).*(数据|记录).*(查询|查看)`),
			},
			Keywords:    []string{"历史", "过去", "昨天", "上周", "上月", "历史数据"},
			EntityTypes: []EntityType{EntityDevice, EntityPoint, EntityTime},
			SlotDefs: map[string]*SlotDefinition{
				"target":    {Name: "target", Type: "string", Required: true, EntityType: EntityDevice},
				"startTime": {Name: "startTime", Type: "time", Required: true, EntityType: EntityTime},
				"endTime":   {Name: "endTime", Type: "time", Required: false, EntityType: EntityTime},
			},
			Priority: 9,
		},
		{
			IntentType: IntentQuery,
			IntentName: "query_statistics",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)(统计|分析|计算).*(数据|指标|结果)`),
				regexp.MustCompile(`(?i)(平均值|最大值|最小值|总和).*(查询|计算)`),
			},
			Keywords:    []string{"统计", "分析", "平均值", "最大值", "最小值", "总和"},
			EntityTypes: []EntityType{EntityDevice, EntityMetric, EntityTime},
			SlotDefs: map[string]*SlotDefinition{
				"target":     {Name: "target", Type: "string", Required: true, EntityType: EntityDevice},
				"metric":     {Name: "metric", Type: "string", Required: true, EntityType: EntityMetric},
				"aggregation": {Name: "aggregation", Type: "string", Required: true},
			},
			Priority: 8,
		},
	}

	// 控制意图模式
	controlPatterns := []*IntentPattern{
		{
			IntentType: IntentControl,
			IntentName: "control_device",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)(启动|开启|打开|关闭|停止|重启).*(设备|机器|系统)`),
				regexp.MustCompile(`(?i)(设置|调整|修改).*(参数|配置|阈值)`),
			},
			Keywords:    []string{"启动", "开启", "打开", "关闭", "停止", "重启", "设置", "调整", "修改"},
			EntityTypes: []EntityType{EntityDevice, EntityStatus},
			SlotDefs: map[string]*SlotDefinition{
				"device": {Name: "device", Type: "string", Required: true, EntityType: EntityDevice},
				"action": {Name: "action", Type: "string", Required: true},
				"params": {Name: "params", Type: "map", Required: false},
			},
			Priority: 10,
		},
		{
			IntentType: IntentControl,
			IntentName: "control_threshold",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)(设置|修改|调整).*(阈值|告警阈值|上限|下限)`),
				regexp.MustCompile(`(?i)(阈值|告警阈值).*(设置|修改|调整)`),
			},
			Keywords:    []string{"阈值", "告警阈值", "上限", "下限", "设置", "修改", "调整"},
			EntityTypes: []EntityType{EntityDevice, EntityThreshold},
			SlotDefs: map[string]*SlotDefinition{
				"target":    {Name: "target", Type: "string", Required: true, EntityType: EntityDevice},
				"threshold": {Name: "threshold", Type: "float", Required: true, EntityType: EntityThreshold},
				"type":      {Name: "type", Type: "string", Required: true},
			},
			Priority: 9,
		},
	}

	// 配置意图模式
	configPatterns := []*IntentPattern{
		{
			IntentType: IntentConfig,
			IntentName: "config_system",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)(配置|设置).*(系统|参数|选项)`),
				regexp.MustCompile(`(?i)(系统|参数|选项).*(配置|设置)`),
			},
			Keywords:    []string{"配置", "设置", "系统配置", "参数配置"},
			EntityTypes: []EntityType{},
			SlotDefs: map[string]*SlotDefinition{
				"configType": {Name: "configType", Type: "string", Required: true},
				"configValue": {Name: "configValue", Type: "interface", Required: true},
			},
			Priority: 8,
		},
		{
			IntentType: IntentConfig,
			IntentName: "config_alarm",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)(配置|设置).*(告警|报警|预警).*(规则|策略)`),
				regexp.MustCompile(`(?i)(告警|报警|预警).*(规则|策略).*(配置|设置)`),
			},
			Keywords:    []string{"告警", "报警", "预警", "告警规则", "告警策略"},
			EntityTypes: []EntityType{EntityDevice, EntityThreshold},
			SlotDefs: map[string]*SlotDefinition{
				"alarmType": {Name: "alarmType", Type: "string", Required: true},
				"rules":     {Name: "rules", Type: "array", Required: true},
			},
			Priority: 9,
		},
	}

	// 诊断意图模式
	diagnosePatterns := []*IntentPattern{
		{
			IntentType: IntentDiagnose,
			IntentName: "diagnose_fault",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)(诊断|分析|排查).*(故障|异常|问题)`),
				regexp.MustCompile(`(?i)(故障|异常|问题).*(诊断|分析|排查|原因)`),
			},
			Keywords:    []string{"诊断", "故障", "异常", "问题", "排查", "原因分析"},
			EntityTypes: []EntityType{EntityDevice, EntityStatus},
			SlotDefs: map[string]*SlotDefinition{
				"target": {Name: "target", Type: "string", Required: true, EntityType: EntityDevice},
				"faultType": {Name: "faultType", Type: "string", Required: false},
			},
			Priority: 10,
		},
		{
			IntentType: IntentDiagnose,
			IntentName: "diagnose_performance",
			Patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)(分析|评估|检查).*(性能|效率|运行状态)`),
				regexp.MustCompile(`(?i)(性能|效率).*(分析|评估|问题)`),
			},
			Keywords:    []string{"性能", "效率", "运行状态", "性能分析"},
			EntityTypes: []EntityType{EntityDevice, EntityMetric},
			SlotDefs: map[string]*SlotDefinition{
				"target": {Name: "target", Type: "string", Required: true, EntityType: EntityDevice},
				"metrics": {Name: "metrics", Type: "array", Required: false, EntityType: EntityMetric},
			},
			Priority: 9,
		},
	}

	r.patterns[IntentQuery] = queryPatterns
	r.patterns[IntentControl] = controlPatterns
	r.patterns[IntentConfig] = configPatterns
	r.patterns[IntentDiagnose] = diagnosePatterns
}

// initDefaultEntityRules 初始化默认实体识别规则
func (r *IntentRecognizer) initDefaultEntityRules() {
	// 设备实体规则
	r.entityRules[EntityDevice] = []*EntityRule{
		{
			Type: EntityDevice,
			Pattern: regexp.MustCompile(`(?i)(逆变器|汇流箱|变压器|配电柜|组件|风机|储能|电池)`),
			Dictionary: []string{"逆变器", "汇流箱", "变压器", "配电柜", "组件", "风机", "储能", "电池"},
			Normalizer: func(s string) string {
				return strings.TrimSpace(s)
			},
		},
		{
			Type:    EntityDevice,
			Pattern: regexp.MustCompile(`(?i)([A-Z]{2,3}-\d{3,4})`),
			Normalizer: func(s string) string {
				return strings.ToUpper(s)
			},
		},
	}

	// 测点实体规则
	r.entityRules[EntityPoint] = []*EntityRule{
		{
			Type: EntityPoint,
			Pattern: regexp.MustCompile(`(?i)(电压|电流|功率|温度|辐照|风速|发电量|频率)`),
			Dictionary: []string{"电压", "电流", "功率", "温度", "辐照", "风速", "发电量", "频率"},
			Normalizer: func(s string) string {
				return strings.TrimSpace(s)
			},
		},
		{
			Type:    EntityPoint,
			Pattern: regexp.MustCompile(`(?i)(点[0-9]+|测点[0-9]+|POINT_[0-9]+)`),
			Normalizer: func(s string) string {
				return strings.ToUpper(s)
			},
		},
	}

	// 电站实体规则
	r.entityRules[EntityStation] = []*EntityRule{
		{
			Type: EntityStation,
			Pattern: regexp.MustCompile(`(?i)(电站|光伏电站|风电场|储能站)`),
			Dictionary: []string{"电站", "光伏电站", "风电场", "储能站"},
			Normalizer: func(s string) string {
				return strings.TrimSpace(s)
			},
		},
	}

	// 时间实体规则
	r.entityRules[EntityTime] = []*EntityRule{
		{
			Type:    EntityTime,
			Pattern: regexp.MustCompile(`(?i)(今天|昨天|前天|本周|上周|本月|上月|最近\d+天|最近\d+小时)`),
			Normalizer: func(s string) string {
				return normalizeTime(s)
			},
		},
		{
			Type:    EntityTime,
			Pattern: regexp.MustCompile(`(?i)(\d{4}-\d{2}-\d{2})`),
			Normalizer: func(s string) string {
				return s
			},
		},
		{
			Type:    EntityTime,
			Pattern: regexp.MustCompile(`(?i)(\d{2}:\d{2}|\d{1,2}点\d{1,2}分)`),
			Normalizer: func(s string) string {
				return s
			},
		},
	}

	// 指标实体规则
	r.entityRules[EntityMetric] = []*EntityRule{
		{
			Type: EntityMetric,
			Pattern: regexp.MustCompile(`(?i)(发电量|上网电量|自用电量|等效利用小时|PR值|系统效率|转换效率)`),
			Dictionary: []string{"发电量", "上网电量", "自用电量", "等效利用小时", "PR值", "系统效率", "转换效率"},
			Normalizer: func(s string) string {
				return strings.TrimSpace(s)
			},
		},
	}

	// 阈值实体规则
	r.entityRules[EntityThreshold] = []*EntityRule{
		{
			Type:    EntityThreshold,
			Pattern: regexp.MustCompile(`(?i)(\d+\.?\d*)\s*(kw|mw|v|a|℃|%|hz)`),
			Normalizer: func(s string) string {
				return s
			},
		},
	}

	// 状态实体规则
	r.entityRules[EntityStatus] = []*EntityRule{
		{
			Type: EntityStatus,
			Pattern: regexp.MustCompile(`(?i)(运行|停止|故障|告警|正常|异常|在线|离线)`),
			Dictionary: []string{"运行", "停止", "故障", "告警", "正常", "异常", "在线", "离线"},
			Normalizer: func(s string) string {
				return strings.TrimSpace(s)
			},
		},
	}
}

// Recognize 识别意图
func (r *IntentRecognizer) Recognize(ctx context.Context, text string) (*Intent, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// 提取实体
	entities := r.extractEntities(text)

	// 匹配意图模式
	intent := r.matchIntent(text, entities)

	// 填充槽位
	r.fillSlots(intent, entities)

	intent.RawText = text
	intent.Timestamp = time.Now()

	return intent, nil
}

// extractEntities 提取实体
func (r *IntentRecognizer) extractEntities(text string) []Entity {
	entities := make([]Entity, 0)

	r.mu.RLock()
	defer r.mu.RUnlock()

	for entityType, rules := range r.entityRules {
		for _, rule := range rules {
			// 模式匹配
			if rule.Pattern != nil {
				matches := rule.Pattern.FindAllStringIndex(text, -1)
				for _, match := range matches {
					if len(entities) >= r.config.MaxEntities {
						return entities
					}

					value := text[match[0]:match[1]]
					normalized := value
					if rule.Normalizer != nil {
						normalized = rule.Normalizer(value)
					}

					entities = append(entities, Entity{
						Type:       entityType,
						Value:      value,
						Normalized: normalized,
						Position: Position{
							Start: match[0],
							End:   match[1],
						},
					})
				}
			}

			// 字典匹配
			if len(rule.Dictionary) > 0 {
				for _, dictWord := range rule.Dictionary {
					index := strings.Index(text, dictWord)
					if index != -1 {
						if len(entities) >= r.config.MaxEntities {
							return entities
						}

						normalized := dictWord
						if rule.Normalizer != nil {
							normalized = rule.Normalizer(dictWord)
						}

						entities = append(entities, Entity{
							Type:       entityType,
							Value:      dictWord,
							Normalized: normalized,
							Position: Position{
								Start: index,
								End:   index + len(dictWord),
							},
						})
					}
				}
			}
		}
	}

	// 去重
	entities = deduplicateEntities(entities)

	return entities
}

// matchIntent 匹配意图
func (r *IntentRecognizer) matchIntent(text string, entities []Entity) *Intent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bestIntent := &Intent{
		Type:       IntentUnknown,
		Confidence: 0,
		Entities:   entities,
		Slots:      make(map[string]*Slot),
	}

	// 遍历所有意图类型
	for intentType, patterns := range r.patterns {
		for _, pattern := range patterns {
			confidence := r.calculateConfidence(text, pattern, entities)

			if confidence > bestIntent.Confidence && confidence >= r.config.MinConfidence {
				bestIntent.Type = intentType
				bestIntent.Name = pattern.IntentName
				bestIntent.Confidence = confidence

				// 初始化槽位
				for slotName, slotDef := range pattern.SlotDefs {
					bestIntent.Slots[slotName] = &Slot{
						Name:     slotDef.Name,
						Required: slotDef.Required,
						Filled:   false,
					}
				}
			}
		}
	}

	return bestIntent
}

// calculateConfidence 计算置信度
func (r *IntentRecognizer) calculateConfidence(text string, pattern *IntentPattern, entities []Entity) float64 {
	confidence := 0.0

	// 模式匹配得分
	patternMatchCount := 0
	for _, p := range pattern.Patterns {
		if p.MatchString(text) {
			patternMatchCount++
		}
	}
	if len(pattern.Patterns) > 0 {
		confidence += float64(patternMatchCount) / float64(len(pattern.Patterns)) * 0.5
	}

	// 关键词匹配得分
	keywordMatchCount := 0
	lowerText := strings.ToLower(text)
	for _, keyword := range pattern.Keywords {
		if strings.Contains(lowerText, strings.ToLower(keyword)) {
			keywordMatchCount++
		}
	}
	if len(pattern.Keywords) > 0 {
		confidence += float64(keywordMatchCount) / float64(len(pattern.Keywords)) * 0.3
	}

	// 实体匹配得分
	entityMatchCount := 0
	entityTypes := make(map[EntityType]bool)
	for _, et := range pattern.EntityTypes {
		entityTypes[et] = true
	}
	for _, entity := range entities {
		if entityTypes[entity.Type] {
			entityMatchCount++
		}
	}
	if len(pattern.EntityTypes) > 0 {
		confidence += float64(entityMatchCount) / float64(len(pattern.EntityTypes)) * 0.2
	}

	// 优先级加成
	confidence += float64(pattern.Priority) * 0.01

	// 限制在0-1之间
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// fillSlots 填充槽位
func (r *IntentRecognizer) fillSlots(intent *Intent, entities []Entity) {
	if intent.Type == IntentUnknown {
		return
	}

	// 根据实体类型填充槽位
	for slotName, slot := range intent.Slots {
		for _, entity := range entities {
			// 查找匹配的槽位定义
			if pattern := r.findPattern(intent.Type, intent.Name); pattern != nil {
				if slotDef, exists := pattern.SlotDefs[slotName]; exists {
					if slotDef.EntityType == entity.Type {
						slot.Value = entity.Normalized
						slot.Filled = true
						slot.Entities = append(slot.Entities, entity)
						break
					}
				}
			}
		}
	}
}

// findPattern 查找意图模式
func (r *IntentRecognizer) findPattern(intentType IntentType, intentName string) *IntentPattern {
	r.mu.RLock()
	defer r.mu.RUnlock()

	patterns, exists := r.patterns[intentType]
	if !exists {
		return nil
	}

	for _, pattern := range patterns {
		if pattern.IntentName == intentName {
			return pattern
		}
	}

	return nil
}

// AddPattern 添加意图模式
func (r *IntentRecognizer) AddPattern(pattern *IntentPattern) error {
	if pattern == nil {
		return fmt.Errorf("pattern cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.patterns[pattern.IntentType] = append(r.patterns[pattern.IntentType], pattern)
	return nil
}

// AddEntityRule 添加实体识别规则
func (r *IntentRecognizer) AddEntityRule(entityType EntityType, rule *EntityRule) error {
	if rule == nil {
		return fmt.Errorf("rule cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.entityRules[entityType] = append(r.entityRules[entityType], rule)
	return nil
}

// GetSlotPrompt 获取槽位填充提示
func (r *IntentRecognizer) GetSlotPrompt(intent *Intent, slotName string) string {
	if intent == nil {
		return ""
	}

	pattern := r.findPattern(intent.Type, intent.Name)
	if pattern == nil {
		return ""
	}

	if slotDef, exists := pattern.SlotDefs[slotName]; exists {
		if len(slotDef.Prompts) > 0 {
			return slotDef.Prompts[0]
		}
		return fmt.Sprintf("请提供%s信息", slotName)
	}

	return ""
}

// ValidateIntent 验证意图完整性
func (r *IntentRecognizer) ValidateIntent(intent *Intent) error {
	if intent == nil {
		return fmt.Errorf("intent cannot be nil")
	}

	// 检查必填槽位
	for slotName, slot := range intent.Slots {
		if slot.Required && !slot.Filled {
			return fmt.Errorf("required slot '%s' is not filled", slotName)
		}
	}

	return nil
}

// GetMissingSlots 获取未填充的必填槽位
func (r *IntentRecognizer) GetMissingSlots(intent *Intent) []string {
	missing := make([]string, 0)

	if intent == nil {
		return missing
	}

	for slotName, slot := range intent.Slots {
		if slot.Required && !slot.Filled {
			missing = append(missing, slotName)
		}
	}

	return missing
}

// deduplicateEntities 去重实体
func deduplicateEntities(entities []Entity) []Entity {
	seen := make(map[string]bool)
	result := make([]Entity, 0)

	for _, entity := range entities {
		key := fmt.Sprintf("%s-%s-%d-%d", entity.Type, entity.Value, entity.Position.Start, entity.Position.End)
		if !seen[key] {
			seen[key] = true
			result = append(result, entity)
		}
	}

	return result
}

// normalizeTime 标准化时间表达式
func normalizeTime(timeExpr string) string {
	now := time.Now()

	switch strings.ToLower(timeExpr) {
	case "今天":
		return now.Format("2006-01-02")
	case "昨天":
		return now.AddDate(0, 0, -1).Format("2006-01-02")
	case "前天":
		return now.AddDate(0, 0, -2).Format("2006-01-02")
	case "本周":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		startOfWeek := now.AddDate(0, 0, -weekday+1)
		return startOfWeek.Format("2006-01-02")
	case "上周":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		startOfLastWeek := now.AddDate(0, 0, -weekday-6)
		return startOfLastWeek.Format("2006-01-02")
	case "本月":
		return now.Format("2006-01")
	case "上月":
		return now.AddDate(0, -1, 0).Format("2006-01")
	default:
		// 处理"最近N天"、"最近N小时"等
		if strings.Contains(timeExpr, "最近") {
			// 简化处理，返回原始表达式
			return timeExpr
		}
		return timeExpr
	}
}

// BatchRecognize 批量识别意图
func (r *IntentRecognizer) BatchRecognize(ctx context.Context, texts []string) ([]*Intent, error) {
	results := make([]*Intent, len(texts))
	errors := make([]error, len(texts))

	var wg sync.WaitGroup

	for i, text := range texts {
		wg.Add(1)
		go func(index int, t string) {
			defer wg.Done()
			intent, err := r.Recognize(ctx, t)
			results[index] = intent
			errors[index] = err
		}(i, text)
	}

	wg.Wait()

	// 检查是否有错误
	for _, err := range errors {
		if err != nil {
			return results, fmt.Errorf("batch recognition failed: %w", err)
		}
	}

	return results, nil
}

// GetIntentPatterns 获取意图模式
func (r *IntentRecognizer) GetIntentPatterns(intentType IntentType) []*IntentPattern {
	r.mu.RLock()
	defer r.mu.RUnlock()

	patterns, exists := r.patterns[intentType]
	if !exists {
		return nil
	}

	// 返回副本
	result := make([]*IntentPattern, len(patterns))
	copy(result, patterns)
	return result
}

// GetEntityRules 获取实体规则
func (r *IntentRecognizer) GetEntityRules(entityType EntityType) []*EntityRule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rules, exists := r.entityRules[entityType]
	if !exists {
		return nil
	}

	// 返回副本
	result := make([]*EntityRule, len(rules))
	copy(result, rules)
	return result
}
