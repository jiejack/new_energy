package qa

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Answer 答案
type Answer struct {
	Content      string
	Confidence   float64
	References   []*Reference
	Sources      []*KnowledgeSource
	Metadata     map[string]interface{}
	GeneratedAt  time.Time
}

// Reference 引用
type Reference struct {
	SourceID    string
	SourceType  string // "knowledge_base", "device_data", "rule", "document"
	Title       string
	Content     string
	URL         string
	Relevance   float64
	Position    int // 在答案中的位置
}

// KnowledgeSource 知识源
type KnowledgeSource struct {
	SourceID   string
	Name       string
	Type       string
	Weight     float64
	Enabled    bool
	LastUpdate time.Time
}

// AnswerTemplate 答案模板
type AnswerTemplate struct {
	TemplateID  string
	Name        string
	IntentType  IntentType
	IntentName  string
	Template    string
	Variables   []string
	Conditions  []TemplateCondition
	Priority    int
}

// TemplateCondition 模板条件
type TemplateCondition struct {
	Variable string
	Operator string
	Value    interface{}
}

// KnowledgeProvider 知识提供者接口
type KnowledgeProvider interface {
	// Query 查询知识
	Query(ctx context.Context, query string, limit int) ([]*KnowledgeItem, error)
	// GetByID 根据ID获取知识
	GetByID(ctx context.Context, id string) (*KnowledgeItem, error)
	// GetRelated 获取相关知识
	GetRelated(ctx context.Context, id string, limit int) ([]*KnowledgeItem, error)
}

// KnowledgeItem 知识项
type KnowledgeItem struct {
	ID          string
	Title       string
	Content     string
	Category    string
	Tags        []string
	Relevance   float64
	Source      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AnswerGenerator 答案生成器
type AnswerGenerator struct {
	knowledgeProviders map[string]KnowledgeProvider
	templates          map[string][]*AnswerTemplate
	config             *GeneratorConfig
	mu                 sync.RWMutex
}

// GeneratorConfig 生成器配置
type GeneratorConfig struct {
	MinConfidence     float64
	MaxReferences     int
	MaxSources        int
	EnableCache       bool
	CacheTTL          time.Duration
	TemplatePriority  bool // 是否优先使用模板
}

// DefaultGeneratorConfig 默认生成器配置
func DefaultGeneratorConfig() *GeneratorConfig {
	return &GeneratorConfig{
		MinConfidence:    0.5,
		MaxReferences:    5,
		MaxSources:       3,
		EnableCache:      true,
		CacheTTL:         10 * time.Minute,
		TemplatePriority: true,
	}
}

// NewAnswerGenerator 创建答案生成器
func NewAnswerGenerator(config *GeneratorConfig) *AnswerGenerator {
	if config == nil {
		config = DefaultGeneratorConfig()
	}

	generator := &AnswerGenerator{
		knowledgeProviders: make(map[string]KnowledgeProvider),
		templates:          make(map[string][]*AnswerTemplate),
		config:             config,
	}

	// 初始化默认模板
	generator.initDefaultTemplates()

	return generator
}

// initDefaultTemplates 初始化默认答案模板
func (g *AnswerGenerator) initDefaultTemplates() {
	// 查询实时数据模板
	queryRealtimeTemplates := []*AnswerTemplate{
		{
			TemplateID: "tpl_query_realtime_001",
			Name:       "实时数据查询结果",
			IntentType: IntentQuery,
			IntentName: "query_realtime",
			Template:   "{{.target}}的实时数据如下：\n\n{{.data}}\n\n数据更新时间：{{.updateTime}}",
			Variables:  []string{"target", "data", "updateTime"},
			Priority:   10,
		},
		{
			TemplateID: "tpl_query_realtime_002",
			Name:       "实时数据查询结果（简洁）",
			IntentType: IntentQuery,
			IntentName: "query_realtime",
			Template:   "{{.target}}当前{{.metric}}为{{.value}}{{.unit}}，状态{{.status}}",
			Variables:  []string{"target", "metric", "value", "unit", "status"},
			Priority:   9,
		},
	}

	// 查询历史数据模板
	queryHistoryTemplates := []*AnswerTemplate{
		{
			TemplateID: "tpl_query_history_001",
			Name:       "历史数据查询结果",
			IntentType: IntentQuery,
			IntentName: "query_history",
			Template:   "{{.target}}在{{.startTime}}至{{.endTime}}期间的数据统计：\n\n{{.statistics}}\n\n详细数据请查看附件。",
			Variables:  []string{"target", "startTime", "endTime", "statistics"},
			Priority:   10,
		},
	}

	// 查询统计数据模板
	queryStatisticsTemplates := []*AnswerTemplate{
		{
			TemplateID: "tpl_query_statistics_001",
			Name:       "统计数据查询结果",
			IntentType: IntentQuery,
			IntentName: "query_statistics",
			Template:   "{{.target}}的{{.aggregation}}统计结果：\n\n{{.result}}\n\n统计周期：{{.period}}",
			Variables:  []string{"target", "aggregation", "result", "period"},
			Priority:   10,
		},
	}

	// 控制设备模板
	controlDeviceTemplates := []*AnswerTemplate{
		{
			TemplateID: "tpl_control_device_001",
			Name:       "设备控制确认",
			IntentType: IntentControl,
			IntentName: "control_device",
			Template:   "已成功{{.action}}{{.device}}。\n\n当前状态：{{.status}}\n执行时间：{{.executeTime}}",
			Variables:  []string{"action", "device", "status", "executeTime"},
			Priority:   10,
		},
		{
			TemplateID: "tpl_control_device_002",
			Name:       "设备控制失败",
			IntentType: IntentControl,
			IntentName: "control_device",
			Template:   "{{.device}}{{.action}}失败。\n\n失败原因：{{.reason}}\n\n建议：{{.suggestion}}",
			Variables:  []string{"device", "action", "reason", "suggestion"},
			Priority:   10,
			Conditions: []TemplateCondition{
				{Variable: "success", Operator: "eq", Value: false},
			},
		},
	}

	// 控制阈值模板
	controlThresholdTemplates := []*AnswerTemplate{
		{
			TemplateID: "tpl_control_threshold_001",
			Name:       "阈值设置确认",
			IntentType: IntentControl,
			IntentName: "control_threshold",
			Template:   "已成功将{{.target}}的{{.thresholdType}}阈值设置为{{.threshold}}。\n\n新的告警规则已生效。",
			Variables:  []string{"target", "thresholdType", "threshold"},
			Priority:   10,
		},
	}

	// 配置系统模板
	configSystemTemplates := []*AnswerTemplate{
		{
			TemplateID: "tpl_config_system_001",
			Name:       "系统配置确认",
			IntentType: IntentConfig,
			IntentName: "config_system",
			Template:   "系统配置已更新：\n\n配置项：{{.configType}}\n配置值：{{.configValue}}\n\n配置已生效。",
			Variables:  []string{"configType", "configValue"},
			Priority:   10,
		},
	}

	// 配置告警模板
	configAlarmTemplates := []*AnswerTemplate{
		{
			TemplateID: "tpl_config_alarm_001",
			Name:       "告警配置确认",
			IntentType: IntentConfig,
			IntentName: "config_alarm",
			Template:   "告警规则已更新：\n\n告警类型：{{.alarmType}}\n规则详情：{{.rules}}\n\n新规则将在下一周期生效。",
			Variables:  []string{"alarmType", "rules"},
			Priority:   10,
		},
	}

	// 诊断故障模板
	diagnoseFaultTemplates := []*AnswerTemplate{
		{
			TemplateID: "tpl_diagnose_fault_001",
			Name:       "故障诊断结果",
			IntentType: IntentDiagnose,
			IntentName: "diagnose_fault",
			Template:   "{{.target}}故障诊断结果：\n\n故障类型：{{.faultType}}\n故障原因：{{.faultReason}}\n\n建议处理方案：\n{{.solutions}}",
			Variables:  []string{"target", "faultType", "faultReason", "solutions"},
			Priority:   10,
		},
		{
			TemplateID: "tpl_diagnose_fault_002",
			Name:       "故障诊断结果（无故障）",
			IntentType: IntentDiagnose,
			IntentName: "diagnose_fault",
			Template:   "{{.target}}当前运行正常，未发现故障。\n\n设备状态：{{.status}}\n最后检查时间：{{.checkTime}}",
			Variables:  []string{"target", "status", "checkTime"},
			Priority:   9,
			Conditions: []TemplateCondition{
				{Variable: "hasFault", Operator: "eq", Value: false},
			},
		},
	}

	// 诊断性能模板
	diagnosePerformanceTemplates := []*AnswerTemplate{
		{
			TemplateID: "tpl_diagnose_performance_001",
			Name:       "性能分析结果",
			IntentType: IntentDiagnose,
			IntentName: "diagnose_performance",
			Template:   "{{.target}}性能分析报告：\n\n{{.analysis}}\n\n性能评分：{{.score}}/100\n优化建议：{{.suggestions}}",
			Variables:  []string{"target", "analysis", "score", "suggestions"},
			Priority:   10,
		},
	}

	g.templates["query_realtime"] = queryRealtimeTemplates
	g.templates["query_history"] = queryHistoryTemplates
	g.templates["query_statistics"] = queryStatisticsTemplates
	g.templates["control_device"] = controlDeviceTemplates
	g.templates["control_threshold"] = controlThresholdTemplates
	g.templates["config_system"] = configSystemTemplates
	g.templates["config_alarm"] = configAlarmTemplates
	g.templates["diagnose_fault"] = diagnoseFaultTemplates
	g.templates["diagnose_performance"] = diagnosePerformanceTemplates
}

// RegisterKnowledgeProvider 注册知识提供者
func (g *AnswerGenerator) RegisterKnowledgeProvider(name string, provider KnowledgeProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.knowledgeProviders[name] = provider
	return nil
}

// Generate 生成答案
func (g *AnswerGenerator) Generate(ctx context.Context, intent *Intent, context *DialogueContext) (*Answer, error) {
	if intent == nil {
		return nil, fmt.Errorf("intent cannot be nil")
	}

	// 1. 查找匹配的模板
	template := g.findTemplate(intent, context)

	// 2. 查询知识库
	knowledgeItems := g.queryKnowledge(ctx, intent)

	// 3. 生成答案内容
	content := g.generateContent(template, intent, context, knowledgeItems)

	// 4. 生成引用
	references := g.generateReferences(knowledgeItems)

	// 5. 计算置信度
	confidence := g.calculateConfidence(intent, knowledgeItems)

	// 6. 后处理
	content = g.postProcess(content, intent)

	return &Answer{
		Content:     content,
		Confidence:  confidence,
		References:  references,
		Sources:     g.getSources(knowledgeItems),
		Metadata:    make(map[string]interface{}),
		GeneratedAt: time.Now(),
	}, nil
}

// findTemplate 查找匹配的模板
func (g *AnswerGenerator) findTemplate(intent *Intent, context *DialogueContext) *AnswerTemplate {
	key := intent.Name
	templates, exists := g.templates[key]
	if !exists {
		return nil
	}

	// 按优先级排序
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Priority > templates[j].Priority
	})

	// 查找匹配条件的模板
	for _, template := range templates {
		if g.matchTemplateConditions(template, context) {
			return template
		}
	}

	// 返回最高优先级的模板
	if len(templates) > 0 {
		return templates[0]
	}

	return nil
}

// matchTemplateConditions 匹配模板条件
func (g *AnswerGenerator) matchTemplateConditions(template *AnswerTemplate, context *DialogueContext) bool {
	if len(template.Conditions) == 0 {
		return true
	}

	for _, condition := range template.Conditions {
		value, exists := context.Variables[condition.Variable]
		if !exists {
			return false
		}

		if !g.compareValue(value, condition.Operator, condition.Value) {
			return false
		}
	}

	return true
}

// compareValue 比较值
func (g *AnswerGenerator) compareValue(value interface{}, operator string, expected interface{}) bool {
	switch operator {
	case "eq":
		return value == expected
	case "ne":
		return value != expected
	case "gt":
		if v, ok := value.(float64); ok {
			if e, ok := expected.(float64); ok {
				return v > e
			}
		}
	case "lt":
		if v, ok := value.(float64); ok {
			if e, ok := expected.(float64); ok {
				return v < e
			}
		}
	}
	return false
}

// queryKnowledge 查询知识库
func (g *AnswerGenerator) queryKnowledge(ctx context.Context, intent *Intent) []*KnowledgeItem {
	items := make([]*KnowledgeItem, 0)

	g.mu.RLock()
	providers := make([]KnowledgeProvider, 0, len(g.knowledgeProviders))
	for _, provider := range g.knowledgeProviders {
		providers = append(providers, provider)
	}
	g.mu.RUnlock()

	// 构建查询
	query := g.buildQuery(intent)

	// 从各个知识源查询
	for _, provider := range providers {
		results, err := provider.Query(ctx, query, g.config.MaxReferences)
		if err != nil {
			continue
		}
		items = append(items, results...)
	}

	// 按相关性排序
	sort.Slice(items, func(i, j int) bool {
		return items[i].Relevance > items[j].Relevance
	})

	// 限制数量
	if len(items) > g.config.MaxReferences {
		items = items[:g.config.MaxReferences]
	}

	return items
}

// buildQuery 构建查询
func (g *AnswerGenerator) buildQuery(intent *Intent) string {
	var queryParts []string

	// 添加意图类型
	queryParts = append(queryParts, string(intent.Type))

	// 添加意图名称
	if intent.Name != "" {
		queryParts = append(queryParts, intent.Name)
	}

	// 添加实体
	for _, entity := range intent.Entities {
		queryParts = append(queryParts, entity.Normalized)
	}

	// 添加槽位值
	for _, slot := range intent.Slots {
		if slot.Filled {
			queryParts = append(queryParts, fmt.Sprintf("%v", slot.Value))
		}
	}

	return strings.Join(queryParts, " ")
}

// generateContent 生成答案内容
func (g *AnswerGenerator) generateContent(template *AnswerTemplate, intent *Intent, context *DialogueContext, knowledge []*KnowledgeItem) string {
	// 如果有模板，使用模板生成
	if template != nil && g.config.TemplatePriority {
		return g.renderTemplate(template, intent, context, knowledge)
	}

	// 否则，基于知识库生成
	if len(knowledge) > 0 {
		return g.generateFromKnowledge(intent, knowledge)
	}

	// 默认答案
	return g.generateDefaultAnswer(intent)
}

// renderTemplate 渲染模板
func (g *AnswerGenerator) renderTemplate(template *AnswerTemplate, intent *Intent, context *DialogueContext, knowledge []*KnowledgeItem) string {
	// 构建变量映射
	vars := make(map[string]interface{})

	// 从槽位获取变量
	for slotName, slot := range intent.Slots {
		if slot.Filled {
			vars[slotName] = slot.Value
		}
	}

	// 从上下文获取变量
	for k, v := range context.Variables {
		vars[k] = v
	}

	// 从知识库获取变量
	if len(knowledge) > 0 {
		vars["data"] = knowledge[0].Content
		vars["source"] = knowledge[0].Title
	}

	// 简单的模板渲染
	content := template.Template
	for k, v := range vars {
		placeholder := fmt.Sprintf("{{.%s}}", k)
		content = strings.ReplaceAll(content, placeholder, fmt.Sprintf("%v", v))
	}

	return content
}

// generateFromKnowledge 基于知识库生成答案
func (g *AnswerGenerator) generateFromKnowledge(intent *Intent, knowledge []*KnowledgeItem) string {
	if len(knowledge) == 0 {
		return "抱歉，没有找到相关信息。"
	}

	var builder strings.Builder

	// 添加主要答案
	builder.WriteString(knowledge[0].Content)

	// 如果有多个知识项，添加补充信息
	if len(knowledge) > 1 {
		builder.WriteString("\n\n相关信息：\n")
		for i, item := range knowledge[1:] {
			if i >= 3 {
				break
			}
			builder.WriteString(fmt.Sprintf("- %s\n", item.Title))
		}
	}

	return builder.String()
}

// generateDefaultAnswer 生成默认答案
func (g *AnswerGenerator) generateDefaultAnswer(intent *Intent) string {
	switch intent.Type {
	case IntentQuery:
		return "正在为您查询相关信息..."
	case IntentControl:
		return "正在执行您的控制指令..."
	case IntentConfig:
		return "正在为您进行配置..."
	case IntentDiagnose:
		return "正在进行诊断分析..."
	default:
		return "我理解了您的请求，正在处理中..."
	}
}

// generateReferences 生成引用
func (g *AnswerGenerator) generateReferences(knowledge []*KnowledgeItem) []*Reference {
	references := make([]*Reference, 0)

	for i, item := range knowledge {
		if i >= g.config.MaxReferences {
			break
		}

		reference := &Reference{
			SourceID:   item.ID,
			SourceType: item.Source,
			Title:      item.Title,
			Content:    item.Content,
			Relevance:  item.Relevance,
			Position:   i,
		}

		references = append(references, reference)
	}

	return references
}

// calculateConfidence 计算置信度
func (g *AnswerGenerator) calculateConfidence(intent *Intent, knowledge []*KnowledgeItem) float64 {
	confidence := intent.Confidence

	// 根据知识库匹配度调整
	if len(knowledge) > 0 {
		avgRelevance := 0.0
		for _, item := range knowledge {
			avgRelevance += item.Relevance
		}
		avgRelevance /= float64(len(knowledge))

		confidence = confidence*0.6 + avgRelevance*0.4
	}

	// 确保在0-1之间
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0 {
		confidence = 0
	}

	return confidence
}

// getSources 获取知识源
func (g *AnswerGenerator) getSources(knowledge []*KnowledgeItem) []*KnowledgeSource {
	sources := make([]*KnowledgeSource, 0)
	sourceMap := make(map[string]bool)

	for _, item := range knowledge {
		if !sourceMap[item.Source] {
			sourceMap[item.Source] = true
			sources = append(sources, &KnowledgeSource{
				SourceID: item.Source,
				Name:     item.Source,
				Type:     item.Source,
				Weight:   item.Relevance,
				Enabled:  true,
			})
		}

		if len(sources) >= g.config.MaxSources {
			break
		}
	}

	return sources
}

// postProcess 后处理
func (g *AnswerGenerator) postProcess(content string, intent *Intent) string {
	// 去除多余空白
	content = strings.TrimSpace(content)

	// 统一换行符
	content = strings.ReplaceAll(content, "\r\n", "\n")

	// 去除连续空行
	lines := strings.Split(content, "\n")
	var result []string
	prevEmpty := false

	for _, line := range lines {
		isEmpty := strings.TrimSpace(line) == ""
		if isEmpty && prevEmpty {
			continue
		}
		result = append(result, line)
		prevEmpty = isEmpty
	}

	return strings.Join(result, "\n")
}

// AddTemplate 添加答案模板
func (g *AnswerGenerator) AddTemplate(template *AnswerTemplate) error {
	if template == nil {
		return fmt.Errorf("template cannot be nil")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	key := template.IntentName
	g.templates[key] = append(g.templates[key], template)
	return nil
}

// GetTemplates 获取模板
func (g *AnswerGenerator) GetTemplates(intentName string) []*AnswerTemplate {
	g.mu.RLock()
	defer g.mu.RUnlock()

	templates, exists := g.templates[intentName]
	if !exists {
		return nil
	}

	// 返回副本
	result := make([]*AnswerTemplate, len(templates))
	copy(result, templates)
	return result
}

// GenerateWithFallback 带降级的答案生成
func (g *AnswerGenerator) GenerateWithFallback(ctx context.Context, intent *Intent, context *DialogueContext) (*Answer, error) {
	// 尝试主要生成
	answer, err := g.Generate(ctx, intent, context)
	if err != nil {
		return nil, err
	}

	// 如果置信度过低，生成降级答案
	if answer.Confidence < g.config.MinConfidence {
		fallbackAnswer := g.generateFallbackAnswer(intent)
		fallbackAnswer.Confidence = answer.Confidence
		fallbackAnswer.References = answer.References
		return fallbackAnswer, nil
	}

	return answer, nil
}

// generateFallbackAnswer 生成降级答案
func (g *AnswerGenerator) generateFallbackAnswer(intent *Intent) *Answer {
	var content string

	switch intent.Type {
	case IntentQuery:
		content = "抱歉，我无法确定您要查询的具体内容。请尝试提供更详细的信息，例如：\n- 具体的设备名称或编号\n- 要查询的时间范围\n- 需要查看的指标"
	case IntentControl:
		content = "抱歉，我无法理解您的控制指令。请明确说明：\n- 要操作的设备\n- 要执行的操作\n- 操作的参数"
	case IntentConfig:
		content = "抱歉，我无法理解您的配置需求。请说明：\n- 要配置的功能或模块\n- 配置的具体参数"
	case IntentDiagnose:
		content = "抱歉，我无法确定要诊断的目标。请提供：\n- 要诊断的设备或系统\n- 遇到的问题或症状"
	default:
		content = "抱歉，我没有理解您的意思。请尝试换一种方式描述您的需求。"
	}

	return &Answer{
		Content:     content,
		Confidence:  0.3,
		References:  make([]*Reference, 0),
		Sources:     make([]*KnowledgeSource, 0),
		Metadata:    make(map[string]interface{}),
		GeneratedAt: time.Now(),
	}
}

// BatchGenerate 批量生成答案
func (g *AnswerGenerator) BatchGenerate(ctx context.Context, intents []*Intent, contexts []*DialogueContext) ([]*Answer, error) {
	if len(intents) != len(contexts) {
		return nil, fmt.Errorf("intents and contexts length mismatch")
	}

	answers := make([]*Answer, len(intents))
	errors := make([]error, len(intents))

	var wg sync.WaitGroup

	for i := range intents {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			answer, err := g.Generate(ctx, intents[index], contexts[index])
			answers[index] = answer
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// 检查错误
	for _, err := range errors {
		if err != nil {
			return answers, fmt.Errorf("batch generation failed: %w", err)
		}
	}

	return answers, nil
}

// ValidateAnswer 验证答案
func (g *AnswerGenerator) ValidateAnswer(answer *Answer) error {
	if answer == nil {
		return fmt.Errorf("answer cannot be nil")
	}

	if answer.Content == "" {
		return fmt.Errorf("answer content cannot be empty")
	}

	if answer.Confidence < 0 || answer.Confidence > 1 {
		return fmt.Errorf("invalid confidence: %f", answer.Confidence)
	}

	return nil
}

// EnhanceAnswer 增强答案
func (g *AnswerGenerator) EnhanceAnswer(ctx context.Context, answer *Answer, intent *Intent) (*Answer, error) {
	if answer == nil {
		return nil, fmt.Errorf("answer cannot be nil")
	}

	// 添加格式化
	enhancedContent := g.formatAnswer(answer.Content, intent)

	// 添加相关建议
	suggestions := g.generateSuggestions(intent)

	// 创建增强后的答案
	enhanced := &Answer{
		Content:     enhancedContent,
		Confidence:  answer.Confidence,
		References:  answer.References,
		Sources:     answer.Sources,
		GeneratedAt: time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// 添加建议到元数据
	if len(suggestions) > 0 {
		enhanced.Metadata["suggestions"] = suggestions
	}

	return enhanced, nil
}

// formatAnswer 格式化答案
func (g *AnswerGenerator) formatAnswer(content string, intent *Intent) string {
	// 添加标题
	var title string
	switch intent.Type {
	case IntentQuery:
		title = "📊 查询结果"
	case IntentControl:
		title = "⚙️ 执行结果"
	case IntentConfig:
		title = "🔧 配置结果"
	case IntentDiagnose:
		title = "🔍 诊断结果"
	default:
		title = "📋 处理结果"
	}

	return fmt.Sprintf("%s\n\n%s", title, content)
}

// generateSuggestions 生成建议
func (g *AnswerGenerator) generateSuggestions(intent *Intent) []string {
	suggestions := make([]string, 0)

	switch intent.Type {
	case IntentQuery:
		suggestions = []string{
			"查看历史趋势",
			"导出数据报表",
			"设置数据告警",
		}
	case IntentControl:
		suggestions = []string{
			"查看设备状态",
			"查看操作日志",
			"设置定时任务",
		}
	case IntentConfig:
		suggestions = []string{
			"查看配置历史",
			"备份当前配置",
			"应用推荐配置",
		}
	case IntentDiagnose:
		suggestions = []string{
			"查看详细报告",
			"创建工单",
			"预约维护",
		}
	}

	return suggestions
}
