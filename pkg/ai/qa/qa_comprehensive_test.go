package qa

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIntentRecognizer_NilConfig(t *testing.T) {
	r := NewIntentRecognizer(nil)
	require.NotNil(t, r)
	require.NotNil(t, r.config)
	assert.Equal(t, 0.6, r.config.MinConfidence)
	assert.Equal(t, 20, r.config.MaxEntities)
	assert.True(t, r.config.EnableSubIntents)
}

func TestNewIntentRecognizer_CustomConfig(t *testing.T) {
	cfg := &RecognizerConfig{
		MinConfidence:    0.8,
		MaxEntities:      10,
		EnableSubIntents: false,
		CacheEnabled:     false,
		CacheTTL:         time.Minute,
	}
	r := NewIntentRecognizer(cfg)
	require.NotNil(t, r)
	assert.Equal(t, 0.8, r.config.MinConfidence)
	assert.Equal(t, 10, r.config.MaxEntities)
}

func newLowConfRecognizer() *IntentRecognizer {
	return NewIntentRecognizer(&RecognizerConfig{
		MinConfidence:    0.0,
		MaxEntities:      20,
		EnableSubIntents: true,
		CacheEnabled:     true,
		CacheTTL:         5 * time.Minute,
	})
}

func TestIntentRecognizer_Recognize_EmptyText(t *testing.T) {
	r := NewIntentRecognizer(nil)
	_, err := r.Recognize(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "text cannot be empty")
}

func TestIntentRecognizer_Recognize_QueryRealtime(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "查询逆变器的实时数据")
	require.NoError(t, err)
	require.NotNil(t, intent)
	assert.Equal(t, IntentQuery, intent.Type)
	assert.Equal(t, "query_realtime", intent.Name)
	assert.Greater(t, intent.Confidence, 0.0)
	assert.Equal(t, "查询逆变器的实时数据", intent.RawText)
}

func TestIntentRecognizer_Recognize_QueryHistory(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "查询昨天的历史数据")
	require.NoError(t, err)
	require.NotNil(t, intent)
	assert.Equal(t, IntentQuery, intent.Type)
}

func TestIntentRecognizer_Recognize_QueryStatistics(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "统计发电量数据的平均值")
	require.NoError(t, err)
	require.NotNil(t, intent)
	assert.Equal(t, IntentQuery, intent.Type)
	assert.Equal(t, "query_statistics", intent.Name)
}

func TestIntentRecognizer_Recognize_ControlDevice(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "启动逆变器设备")
	require.NoError(t, err)
	require.NotNil(t, intent)
	assert.Equal(t, IntentControl, intent.Type)
	assert.Equal(t, "control_device", intent.Name)
}

func TestIntentRecognizer_Recognize_ControlThreshold(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "设置告警阈值")
	require.NoError(t, err)
	require.NotNil(t, intent)
	assert.Equal(t, IntentControl, intent.Type)
}

func TestIntentRecognizer_Recognize_ConfigSystem(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "配置系统参数")
	require.NoError(t, err)
	require.NotNil(t, intent)
	assert.Equal(t, IntentConfig, intent.Type)
	assert.Equal(t, "config_system", intent.Name)
}

func TestIntentRecognizer_Recognize_ConfigAlarm(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "配置告警规则")
	require.NoError(t, err)
	require.NotNil(t, intent)
	assert.Equal(t, IntentConfig, intent.Type)
	assert.Equal(t, "config_alarm", intent.Name)
}

func TestIntentRecognizer_Recognize_DiagnoseFault(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "诊断逆变器故障")
	require.NoError(t, err)
	require.NotNil(t, intent)
	assert.Equal(t, IntentDiagnose, intent.Type)
	assert.Equal(t, "diagnose_fault", intent.Name)
}

func TestIntentRecognizer_Recognize_DiagnosePerformance(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "分析逆变器性能")
	require.NoError(t, err)
	require.NotNil(t, intent)
	assert.Equal(t, IntentDiagnose, intent.Type)
	assert.Equal(t, "diagnose_performance", intent.Name)
}

func TestIntentRecognizer_Recognize_UnknownIntent(t *testing.T) {
	r := NewIntentRecognizer(nil)
	intent, err := r.Recognize(context.Background(), "随便说说而已")
	require.NoError(t, err)
	require.NotNil(t, intent)
	assert.Equal(t, IntentUnknown, intent.Type)
}

func TestIntentRecognizer_Recognize_EntityExtraction(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "查询逆变器的电压实时数据")
	require.NoError(t, err)
	require.NotNil(t, intent)
	foundDevice := false
	foundPoint := false
	for _, e := range intent.Entities {
		if e.Type == EntityDevice {
			foundDevice = true
		}
		if e.Type == EntityPoint {
			foundPoint = true
		}
	}
	assert.True(t, foundDevice)
	assert.True(t, foundPoint)
}

func TestIntentRecognizer_Recognize_TimeEntity(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "查询昨天的历史数据记录")
	require.NoError(t, err)
	require.NotNil(t, intent)
	foundTime := false
	for _, e := range intent.Entities {
		if e.Type == EntityTime {
			foundTime = true
			assert.NotEmpty(t, e.Normalized)
		}
	}
	assert.True(t, foundTime)
}

func TestIntentRecognizer_Recognize_StationEntity(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "查询光伏电站的数据")
	require.NoError(t, err)
	require.NotNil(t, intent)
	foundStation := false
	for _, e := range intent.Entities {
		if e.Type == EntityStation {
			foundStation = true
		}
	}
	assert.True(t, foundStation)
}

func TestIntentRecognizer_Recognize_StatusEntity(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "查看设备运行状态")
	require.NoError(t, err)
	require.NotNil(t, intent)
	foundStatus := false
	for _, e := range intent.Entities {
		if e.Type == EntityStatus {
			foundStatus = true
		}
	}
	assert.True(t, foundStatus)
}

func TestIntentRecognizer_Recognize_ThresholdEntity(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "设置阈值100kW")
	require.NoError(t, err)
	require.NotNil(t, intent)
	foundThreshold := false
	for _, e := range intent.Entities {
		if e.Type == EntityThreshold {
			foundThreshold = true
		}
	}
	assert.True(t, foundThreshold)
}

func TestIntentRecognizer_Recognize_MetricEntity(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "查询发电量指标")
	require.NoError(t, err)
	require.NotNil(t, intent)
	foundMetric := false
	for _, e := range intent.Entities {
		if e.Type == EntityMetric {
			foundMetric = true
		}
	}
	assert.True(t, foundMetric)
}

func TestIntentRecognizer_Recognize_SlotFilling(t *testing.T) {
	r := newLowConfRecognizer()
	intent, err := r.Recognize(context.Background(), "查询逆变器的实时数据")
	require.NoError(t, err)
	require.NotNil(t, intent)
	if intent.Type != IntentUnknown {
		assert.NotNil(t, intent.Slots)
	}
}

func TestIntentRecognizer_Recognize_Timestamp(t *testing.T) {
	r := NewIntentRecognizer(nil)
	before := time.Now()
	intent, err := r.Recognize(context.Background(), "查询实时数据")
	require.NoError(t, err)
	after := time.Now()
	assert.True(t, intent.Timestamp.After(before) || intent.Timestamp.Equal(before))
	assert.True(t, intent.Timestamp.Before(after) || intent.Timestamp.Equal(after))
}

func TestIntentRecognizer_BatchRecognize(t *testing.T) {
	r := NewIntentRecognizer(nil)
	texts := []string{
		"查询实时数据",
		"启动逆变器设备",
		"配置系统参数",
	}
	results, err := r.BatchRecognize(context.Background(), texts)
	require.NoError(t, err)
	assert.Len(t, results, 3)
	for _, intent := range results {
		require.NotNil(t, intent)
	}
}

func TestIntentRecognizer_BatchRecognize_WithEmpty(t *testing.T) {
	r := NewIntentRecognizer(nil)
	texts := []string{
		"查询实时数据",
		"",
	}
	_, err := r.BatchRecognize(context.Background(), texts)
	assert.Error(t, err)
}

func TestIntentRecognizer_AddPattern_Nil(t *testing.T) {
	r := NewIntentRecognizer(nil)
	err := r.AddPattern(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pattern cannot be nil")
}

func TestIntentRecognizer_AddPattern_Valid(t *testing.T) {
	r := NewIntentRecognizer(nil)
	pattern := &IntentPattern{
		IntentType: IntentQuery,
		IntentName: "custom_query",
		Patterns:   []*regexp.Regexp{regexp.MustCompile(`(?i)custom.*query`)},
		Keywords:   []string{"custom", "query"},
		Priority:   5,
	}
	err := r.AddPattern(pattern)
	assert.NoError(t, err)
	patterns := r.GetIntentPatterns(IntentQuery)
	found := false
	for _, p := range patterns {
		if p.IntentName == "custom_query" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestIntentRecognizer_AddEntityRule_Nil(t *testing.T) {
	r := NewIntentRecognizer(nil)
	err := r.AddEntityRule(EntityDevice, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule cannot be nil")
}

func TestIntentRecognizer_AddEntityRule_Valid(t *testing.T) {
	r := NewIntentRecognizer(nil)
	rule := &EntityRule{
		Type:       EntityDevice,
		Pattern:    regexp.MustCompile(`(?i)custom_device_\w+`),
		Dictionary: []string{"custom_device_a"},
	}
	err := r.AddEntityRule(EntityDevice, rule)
	assert.NoError(t, err)
	rules := r.GetEntityRules(EntityDevice)
	found := false
	for _, r := range rules {
		if r.Pattern != nil && r.Pattern.String() == `(?i)custom_device_\w+` {
			found = true
		}
	}
	assert.True(t, found)
}

func TestIntentRecognizer_GetSlotPrompt_NilIntent(t *testing.T) {
	r := NewIntentRecognizer(nil)
	prompt := r.GetSlotPrompt(nil, "target")
	assert.Empty(t, prompt)
}

func TestIntentRecognizer_GetSlotPrompt_UnknownIntent(t *testing.T) {
	r := NewIntentRecognizer(nil)
	intent := &Intent{Type: IntentUnknown, Name: "unknown"}
	prompt := r.GetSlotPrompt(intent, "target")
	assert.Empty(t, prompt)
}

func TestIntentRecognizer_GetSlotPrompt_ExistingSlot(t *testing.T) {
	r := NewIntentRecognizer(nil)
	intent := &Intent{Type: IntentQuery, Name: "query_realtime"}
	prompt := r.GetSlotPrompt(intent, "target")
	assert.NotEmpty(t, prompt)
}

func TestIntentRecognizer_GetSlotPrompt_NonExistingSlot(t *testing.T) {
	r := NewIntentRecognizer(nil)
	intent := &Intent{Type: IntentQuery, Name: "query_realtime"}
	prompt := r.GetSlotPrompt(intent, "nonexistent")
	assert.Empty(t, prompt)
}

func TestIntentRecognizer_GetSlotPrompt_SlotWithPrompts(t *testing.T) {
	r := NewIntentRecognizer(nil)
	pattern := &IntentPattern{
		IntentType: IntentQuery,
		IntentName: "test_prompt",
		SlotDefs: map[string]*SlotDefinition{
			"field1": {
				Name:     "field1",
				Type:     "string",
				Required: true,
				Prompts:  []string{"请输入field1的值"},
			},
		},
	}
	r.AddPattern(pattern)
	intent := &Intent{Type: IntentQuery, Name: "test_prompt"}
	prompt := r.GetSlotPrompt(intent, "field1")
	assert.Equal(t, "请输入field1的值", prompt)
}

func TestIntentRecognizer_GetSlotPrompt_SlotWithoutPrompts(t *testing.T) {
	r := NewIntentRecognizer(nil)
	pattern := &IntentPattern{
		IntentType: IntentQuery,
		IntentName: "test_noprompt",
		SlotDefs: map[string]*SlotDefinition{
			"field2": {
				Name:     "field2",
				Type:     "string",
				Required: true,
			},
		},
	}
	r.AddPattern(pattern)
	intent := &Intent{Type: IntentQuery, Name: "test_noprompt"}
	prompt := r.GetSlotPrompt(intent, "field2")
	assert.Contains(t, prompt, "field2")
}

func TestIntentRecognizer_ValidateIntent_NilIntent(t *testing.T) {
	r := NewIntentRecognizer(nil)
	err := r.ValidateIntent(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "intent cannot be nil")
}

func TestIntentRecognizer_ValidateIntent_AllSlotsFilled(t *testing.T) {
	r := NewIntentRecognizer(nil)
	intent := &Intent{
		Type: IntentQuery,
		Name: "test",
		Slots: map[string]*Slot{
			"target": {Name: "target", Required: true, Filled: true},
		},
	}
	err := r.ValidateIntent(intent)
	assert.NoError(t, err)
}

func TestIntentRecognizer_ValidateIntent_MissingRequiredSlot(t *testing.T) {
	r := NewIntentRecognizer(nil)
	intent := &Intent{
		Type: IntentQuery,
		Name: "test",
		Slots: map[string]*Slot{
			"target": {Name: "target", Required: true, Filled: false},
		},
	}
	err := r.ValidateIntent(intent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required slot")
}

func TestIntentRecognizer_ValidateIntent_OptionalSlotUnfilled(t *testing.T) {
	r := NewIntentRecognizer(nil)
	intent := &Intent{
		Type: IntentQuery,
		Name: "test",
		Slots: map[string]*Slot{
			"target": {Name: "target", Required: true, Filled: true},
			"metric": {Name: "metric", Required: false, Filled: false},
		},
	}
	err := r.ValidateIntent(intent)
	assert.NoError(t, err)
}

func TestIntentRecognizer_GetMissingSlots_NilIntent(t *testing.T) {
	r := NewIntentRecognizer(nil)
	missing := r.GetMissingSlots(nil)
	assert.Empty(t, missing)
}

func TestIntentRecognizer_GetMissingSlots_AllFilled(t *testing.T) {
	r := NewIntentRecognizer(nil)
	intent := &Intent{
		Type: IntentQuery,
		Slots: map[string]*Slot{
			"target": {Name: "target", Required: true, Filled: true},
		},
	}
	missing := r.GetMissingSlots(intent)
	assert.Empty(t, missing)
}

func TestIntentRecognizer_GetMissingSlots_SomeMissing(t *testing.T) {
	r := NewIntentRecognizer(nil)
	intent := &Intent{
		Type: IntentQuery,
		Slots: map[string]*Slot{
			"target":    {Name: "target", Required: true, Filled: true},
			"startTime": {Name: "startTime", Required: true, Filled: false},
			"metric":    {Name: "metric", Required: false, Filled: false},
		},
	}
	missing := r.GetMissingSlots(intent)
	assert.Len(t, missing, 1)
	assert.Contains(t, missing, "startTime")
}

func TestIntentRecognizer_GetIntentPatterns_Existing(t *testing.T) {
	r := NewIntentRecognizer(nil)
	patterns := r.GetIntentPatterns(IntentQuery)
	assert.NotEmpty(t, patterns)
}

func TestIntentRecognizer_GetIntentPatterns_NonExisting(t *testing.T) {
	r := NewIntentRecognizer(nil)
	patterns := r.GetIntentPatterns(IntentType("nonexistent"))
	assert.Nil(t, patterns)
}

func TestIntentRecognizer_GetEntityRules_Existing(t *testing.T) {
	r := NewIntentRecognizer(nil)
	rules := r.GetEntityRules(EntityDevice)
	assert.NotEmpty(t, rules)
}

func TestIntentRecognizer_GetEntityRules_NonExisting(t *testing.T) {
	r := NewIntentRecognizer(nil)
	rules := r.GetEntityRules(EntityType("nonexistent"))
	assert.Nil(t, rules)
}

func TestIntentRecognizer_MaxEntities(t *testing.T) {
	cfg := &RecognizerConfig{
		MinConfidence: 0.0,
		MaxEntities:   2,
	}
	r := NewIntentRecognizer(cfg)
	intent, err := r.Recognize(context.Background(), "查询逆变器电压电流温度辐照风速发电量频率数据")
	require.NoError(t, err)
	require.NotNil(t, intent)
	assert.LessOrEqual(t, len(intent.Entities), 2)
}

func TestDeduplicateEntities(t *testing.T) {
	entities := []Entity{
		{Type: EntityDevice, Value: "逆变器", Position: Position{Start: 0, End: 3}},
		{Type: EntityDevice, Value: "逆变器", Position: Position{Start: 0, End: 3}},
		{Type: EntityDevice, Value: "变压器", Position: Position{Start: 5, End: 8}},
	}
	result := deduplicateEntities(entities)
	assert.Len(t, result, 2)
}

func TestNormalizeTime(t *testing.T) {
	now := time.Now()
	assert.Equal(t, now.Format("2006-01-02"), normalizeTime("今天"))
	assert.Equal(t, now.AddDate(0, 0, -1).Format("2006-01-02"), normalizeTime("昨天"))
	assert.Equal(t, now.AddDate(0, 0, -2).Format("2006-01-02"), normalizeTime("前天"))
	assert.Equal(t, now.Format("2006-01"), normalizeTime("本月"))
	assert.Equal(t, now.AddDate(0, -1, 0).Format("2006-01"), normalizeTime("上月"))
	assert.Equal(t, "2024-01-15", normalizeTime("2024-01-15"))
	assert.Equal(t, "最近7天", normalizeTime("最近7天"))
	assert.Equal(t, "random", normalizeTime("random"))
}

func TestNormalizeTime_Weeks(t *testing.T) {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	expectedStart := now.AddDate(0, 0, -weekday+1).Format("2006-01-02")
	assert.Equal(t, expectedStart, normalizeTime("本周"))

	expectedLastWeek := now.AddDate(0, 0, -weekday-6).Format("2006-01-02")
	assert.Equal(t, expectedLastWeek, normalizeTime("上周"))
}

func TestNewAnswerGenerator_NilConfig(t *testing.T) {
	g := NewAnswerGenerator(nil)
	require.NotNil(t, g)
	require.NotNil(t, g.config)
	assert.Equal(t, 0.5, g.config.MinConfidence)
	assert.Equal(t, 5, g.config.MaxReferences)
	assert.Equal(t, 3, g.config.MaxSources)
	assert.True(t, g.config.TemplatePriority)
}

func TestNewAnswerGenerator_CustomConfig(t *testing.T) {
	cfg := &GeneratorConfig{
		MinConfidence:    0.7,
		MaxReferences:    10,
		MaxSources:       5,
		EnableCache:      false,
		TemplatePriority: false,
	}
	g := NewAnswerGenerator(cfg)
	require.NotNil(t, g)
	assert.Equal(t, 0.7, g.config.MinConfidence)
	assert.False(t, g.config.TemplatePriority)
}

func TestAnswerGenerator_Generate_NilIntent(t *testing.T) {
	g := NewAnswerGenerator(nil)
	_, err := g.Generate(context.Background(), nil, &DialogueContext{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "intent cannot be nil")
}

func TestAnswerGenerator_Generate_QueryIntent(t *testing.T) {
	g := NewAnswerGenerator(nil)
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "query_realtime",
		Confidence: 0.9,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{
		Variables: map[string]interface{}{},
	}
	answer, err := g.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
	assert.NotEmpty(t, answer.Content)
	assert.GreaterOrEqual(t, answer.Confidence, 0.0)
	assert.LessOrEqual(t, answer.Confidence, 1.0)
	assert.NotNil(t, answer.Metadata)
	assert.False(t, answer.GeneratedAt.IsZero())
}

func TestAnswerGenerator_Generate_ControlIntent(t *testing.T) {
	g := NewAnswerGenerator(nil)
	intent := &Intent{
		Type:       IntentControl,
		Name:       "control_device",
		Confidence: 0.8,
		Slots: map[string]*Slot{
			"device": {Name: "device", Filled: true, Value: "逆变器"},
			"action": {Name: "action", Filled: true, Value: "启动"},
		},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := g.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
	assert.NotEmpty(t, answer.Content)
}

func TestAnswerGenerator_Generate_ConfigIntent(t *testing.T) {
	g := NewAnswerGenerator(nil)
	intent := &Intent{
		Type:       IntentConfig,
		Name:       "config_system",
		Confidence: 0.85,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := g.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
	assert.NotEmpty(t, answer.Content)
}

func TestAnswerGenerator_Generate_DiagnoseIntent(t *testing.T) {
	g := NewAnswerGenerator(nil)
	intent := &Intent{
		Type:       IntentDiagnose,
		Name:       "diagnose_fault",
		Confidence: 0.75,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := g.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
	assert.NotEmpty(t, answer.Content)
}

func TestAnswerGenerator_Generate_UnknownIntent(t *testing.T) {
	g := NewAnswerGenerator(nil)
	intent := &Intent{
		Type:       IntentUnknown,
		Name:       "unknown",
		Confidence: 0.1,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := g.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
}

func TestAnswerGenerator_Generate_WithKnowledgeProvider(t *testing.T) {
	g := NewAnswerGenerator(&GeneratorConfig{
		MinConfidence:    0.0,
		MaxReferences:    5,
		MaxSources:       3,
		TemplatePriority: false,
	})
	provider := &mockKnowledgeProvider{
		items: []*KnowledgeItem{
			{ID: "1", Title: "Test", Content: "Test content", Relevance: 0.9, Source: "test"},
		},
	}
	err := g.RegisterKnowledgeProvider("test", provider)
	require.NoError(t, err)
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "query_realtime",
		Confidence: 0.8,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := g.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
	assert.NotEmpty(t, answer.Content)
}

func TestAnswerGenerator_GenerateWithFallback_HighConfidence(t *testing.T) {
	g := NewAnswerGenerator(nil)
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "query_realtime",
		Confidence: 0.9,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := g.GenerateWithFallback(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
}

func TestAnswerGenerator_GenerateWithFallback_LowConfidence(t *testing.T) {
	g := NewAnswerGenerator(&GeneratorConfig{
		MinConfidence:    0.5,
		MaxReferences:    5,
		MaxSources:       3,
		TemplatePriority: true,
	})
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "query_realtime",
		Confidence: 0.1,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := g.GenerateWithFallback(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
}

func TestAnswerGenerator_GenerateWithFallback_LowConfidenceControl(t *testing.T) {
	g := NewAnswerGenerator(&GeneratorConfig{
		MinConfidence:    0.5,
		MaxReferences:    5,
		MaxSources:       3,
		TemplatePriority: true,
	})
	intent := &Intent{
		Type:       IntentControl,
		Name:       "control_device",
		Confidence: 0.1,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := g.GenerateWithFallback(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
}

func TestAnswerGenerator_GenerateWithFallback_LowConfidenceConfig(t *testing.T) {
	g := NewAnswerGenerator(&GeneratorConfig{
		MinConfidence:    0.5,
		MaxReferences:    5,
		MaxSources:       3,
		TemplatePriority: true,
	})
	intent := &Intent{
		Type:       IntentConfig,
		Name:       "config_system",
		Confidence: 0.1,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := g.GenerateWithFallback(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
}

func TestAnswerGenerator_GenerateWithFallback_LowConfidenceDiagnose(t *testing.T) {
	g := NewAnswerGenerator(&GeneratorConfig{
		MinConfidence:    0.5,
		MaxReferences:    5,
		MaxSources:       3,
		TemplatePriority: true,
	})
	intent := &Intent{
		Type:       IntentDiagnose,
		Name:       "diagnose_fault",
		Confidence: 0.1,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := g.GenerateWithFallback(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
}

func TestAnswerGenerator_GenerateWithFallback_LowConfidenceUnknown(t *testing.T) {
	g := NewAnswerGenerator(&GeneratorConfig{
		MinConfidence:    0.5,
		MaxReferences:    5,
		MaxSources:       3,
		TemplatePriority: true,
	})
	intent := &Intent{
		Type:       IntentUnknown,
		Name:       "unknown",
		Confidence: 0.1,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := g.GenerateWithFallback(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
}

func TestAnswerGenerator_BatchGenerate(t *testing.T) {
	g := NewAnswerGenerator(nil)
	intents := []*Intent{
		{Type: IntentQuery, Name: "query_realtime", Confidence: 0.9, Slots: map[string]*Slot{}},
		{Type: IntentControl, Name: "control_device", Confidence: 0.8, Slots: map[string]*Slot{}},
	}
	contexts := []*DialogueContext{
		{Variables: map[string]interface{}{}},
		{Variables: map[string]interface{}{}},
	}
	answers, err := g.BatchGenerate(context.Background(), intents, contexts)
	require.NoError(t, err)
	assert.Len(t, answers, 2)
}

func TestAnswerGenerator_BatchGenerate_LengthMismatch(t *testing.T) {
	g := NewAnswerGenerator(nil)
	intents := []*Intent{
		{Type: IntentQuery, Name: "query_realtime", Confidence: 0.9, Slots: map[string]*Slot{}},
	}
	contexts := []*DialogueContext{
		{Variables: map[string]interface{}{}},
		{Variables: map[string]interface{}{}},
	}
	_, err := g.BatchGenerate(context.Background(), intents, contexts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mismatch")
}

func TestAnswerGenerator_BatchGenerate_WithNilIntent(t *testing.T) {
	g := NewAnswerGenerator(nil)
	intents := []*Intent{
		{Type: IntentQuery, Name: "query_realtime", Confidence: 0.9, Slots: map[string]*Slot{}},
		nil,
	}
	contexts := []*DialogueContext{
		{Variables: map[string]interface{}{}},
		{Variables: map[string]interface{}{}},
	}
	_, err := g.BatchGenerate(context.Background(), intents, contexts)
	assert.Error(t, err)
}

func TestAnswerGenerator_ValidateAnswer_Nil(t *testing.T) {
	g := NewAnswerGenerator(nil)
	err := g.ValidateAnswer(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "answer cannot be nil")
}

func TestAnswerGenerator_ValidateAnswer_EmptyContent(t *testing.T) {
	g := NewAnswerGenerator(nil)
	err := g.ValidateAnswer(&Answer{Content: "", Confidence: 0.5})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "content cannot be empty")
}

func TestAnswerGenerator_ValidateAnswer_InvalidConfidence_Negative(t *testing.T) {
	g := NewAnswerGenerator(nil)
	err := g.ValidateAnswer(&Answer{Content: "test", Confidence: -0.1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid confidence")
}

func TestAnswerGenerator_ValidateAnswer_InvalidConfidence_OverOne(t *testing.T) {
	g := NewAnswerGenerator(nil)
	err := g.ValidateAnswer(&Answer{Content: "test", Confidence: 1.5})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid confidence")
}

func TestAnswerGenerator_ValidateAnswer_Valid(t *testing.T) {
	g := NewAnswerGenerator(nil)
	err := g.ValidateAnswer(&Answer{Content: "test answer", Confidence: 0.8})
	assert.NoError(t, err)
}

func TestAnswerGenerator_EnhanceAnswer_NilAnswer(t *testing.T) {
	g := NewAnswerGenerator(nil)
	_, err := g.EnhanceAnswer(context.Background(), nil, &Intent{Type: IntentQuery})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "answer cannot be nil")
}

func TestAnswerGenerator_EnhanceAnswer_QueryIntent(t *testing.T) {
	g := NewAnswerGenerator(nil)
	answer := &Answer{
		Content:    "Test answer",
		Confidence: 0.8,
		References: []*Reference{},
		Sources:    []*KnowledgeSource{},
	}
	intent := &Intent{Type: IntentQuery, Name: "query_realtime", Confidence: 0.8}
	enhanced, err := g.EnhanceAnswer(context.Background(), answer, intent)
	require.NoError(t, err)
	require.NotNil(t, enhanced)
	assert.NotEmpty(t, enhanced.Content)
	assert.NotNil(t, enhanced.Metadata)
}

func TestAnswerGenerator_EnhanceAnswer_ControlIntent(t *testing.T) {
	g := NewAnswerGenerator(nil)
	answer := &Answer{
		Content:    "Test answer",
		Confidence: 0.8,
		References: []*Reference{},
		Sources:    []*KnowledgeSource{},
	}
	intent := &Intent{Type: IntentControl, Name: "control_device", Confidence: 0.8}
	enhanced, err := g.EnhanceAnswer(context.Background(), answer, intent)
	require.NoError(t, err)
	require.NotNil(t, enhanced)
}

func TestAnswerGenerator_EnhanceAnswer_ConfigIntent(t *testing.T) {
	g := NewAnswerGenerator(nil)
	answer := &Answer{
		Content:    "Test answer",
		Confidence: 0.8,
		References: []*Reference{},
		Sources:    []*KnowledgeSource{},
	}
	intent := &Intent{Type: IntentConfig, Name: "config_system", Confidence: 0.8}
	enhanced, err := g.EnhanceAnswer(context.Background(), answer, intent)
	require.NoError(t, err)
	require.NotNil(t, enhanced)
}

func TestAnswerGenerator_EnhanceAnswer_DiagnoseIntent(t *testing.T) {
	g := NewAnswerGenerator(nil)
	answer := &Answer{
		Content:    "Test answer",
		Confidence: 0.8,
		References: []*Reference{},
		Sources:    []*KnowledgeSource{},
	}
	intent := &Intent{Type: IntentDiagnose, Name: "diagnose_fault", Confidence: 0.8}
	enhanced, err := g.EnhanceAnswer(context.Background(), answer, intent)
	require.NoError(t, err)
	require.NotNil(t, enhanced)
}

func TestAnswerGenerator_EnhanceAnswer_UnknownIntent(t *testing.T) {
	g := NewAnswerGenerator(nil)
	answer := &Answer{
		Content:    "Test answer",
		Confidence: 0.8,
		References: []*Reference{},
		Sources:    []*KnowledgeSource{},
	}
	intent := &Intent{Type: IntentUnknown, Name: "unknown", Confidence: 0.8}
	enhanced, err := g.EnhanceAnswer(context.Background(), answer, intent)
	require.NoError(t, err)
	require.NotNil(t, enhanced)
}

func TestAnswerGenerator_RegisterKnowledgeProvider_Nil(t *testing.T) {
	g := NewAnswerGenerator(nil)
	err := g.RegisterKnowledgeProvider("test", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider cannot be nil")
}

func TestAnswerGenerator_RegisterKnowledgeProvider_Valid(t *testing.T) {
	g := NewAnswerGenerator(nil)
	provider := &mockKnowledgeProvider{items: []*KnowledgeItem{}}
	err := g.RegisterKnowledgeProvider("test", provider)
	assert.NoError(t, err)
}

func TestAnswerGenerator_AddTemplate_Nil(t *testing.T) {
	g := NewAnswerGenerator(nil)
	err := g.AddTemplate(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template cannot be nil")
}

func TestAnswerGenerator_AddTemplate_Valid(t *testing.T) {
	g := NewAnswerGenerator(nil)
	tmpl := &AnswerTemplate{
		TemplateID: "custom_001",
		Name:       "Custom Template",
		IntentName: "custom_intent",
		Template:   "Result: {{.value}}",
		Variables:  []string{"value"},
		Priority:   5,
	}
	err := g.AddTemplate(tmpl)
	assert.NoError(t, err)
	templates := g.GetTemplates("custom_intent")
	assert.NotEmpty(t, templates)
}

func TestAnswerGenerator_GetTemplates_Existing(t *testing.T) {
	g := NewAnswerGenerator(nil)
	templates := g.GetTemplates("query_realtime")
	assert.NotEmpty(t, templates)
}

func TestAnswerGenerator_GetTemplates_NonExisting(t *testing.T) {
	g := NewAnswerGenerator(nil)
	templates := g.GetTemplates("nonexistent")
	assert.Nil(t, templates)
}

func TestAnswerGenerator_QueryKnowledge_WithError(t *testing.T) {
	g := NewAnswerGenerator(nil)
	provider := &mockKnowledgeProvider{err: fmt.Errorf("query failed")}
	g.RegisterKnowledgeProvider("fail_provider", provider)
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "query_realtime",
		Confidence: 0.8,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := g.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
}

func TestAnswerGenerator_QueryKnowledge_MultipleProviders(t *testing.T) {
	g := NewAnswerGenerator(&GeneratorConfig{
		MinConfidence:    0.0,
		MaxReferences:    5,
		MaxSources:       3,
		TemplatePriority: false,
	})
	provider1 := &mockKnowledgeProvider{
		items: []*KnowledgeItem{
			{ID: "1", Title: "Item 1", Content: "Content 1", Relevance: 0.9, Source: "source1"},
		},
	}
	provider2 := &mockKnowledgeProvider{
		items: []*KnowledgeItem{
			{ID: "2", Title: "Item 2", Content: "Content 2", Relevance: 0.8, Source: "source2"},
		},
	}
	g.RegisterKnowledgeProvider("p1", provider1)
	g.RegisterKnowledgeProvider("p2", provider2)
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "query_realtime",
		Confidence: 0.8,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := g.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
}

func TestAnswerGenerator_TemplateConditions(t *testing.T) {
	g := NewAnswerGenerator(nil)
	tmpl := &AnswerTemplate{
		TemplateID: "cond_001",
		Name:       "Conditional Template",
		IntentName: "test_cond",
		Template:   "Conditional result",
		Variables:  []string{},
		Priority:   15,
		Conditions: []TemplateCondition{
			{Variable: "success", Operator: "eq", Value: false},
		},
	}
	g.AddTemplate(tmpl)
	intent := &Intent{
		Type:       IntentControl,
		Name:       "test_cond",
		Confidence: 0.8,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{"success": false}}
	answer, err := g.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	require.NotNil(t, answer)
}

func TestAnswerGenerator_CompareValue_Eq(t *testing.T) {
	g := NewAnswerGenerator(nil)
	assert.True(t, g.compareValue("hello", "eq", "hello"))
	assert.False(t, g.compareValue("hello", "eq", "world"))
	assert.True(t, g.compareValue(42, "eq", 42))
}

func TestAnswerGenerator_CompareValue_Ne(t *testing.T) {
	g := NewAnswerGenerator(nil)
	assert.True(t, g.compareValue("hello", "ne", "world"))
	assert.False(t, g.compareValue("hello", "ne", "hello"))
}

func TestAnswerGenerator_CompareValue_Gt(t *testing.T) {
	g := NewAnswerGenerator(nil)
	assert.True(t, g.compareValue(10.0, "gt", 5.0))
	assert.False(t, g.compareValue(3.0, "gt", 5.0))
	assert.False(t, g.compareValue("hello", "gt", 5.0))
}

func TestAnswerGenerator_CompareValue_Lt(t *testing.T) {
	g := NewAnswerGenerator(nil)
	assert.True(t, g.compareValue(3.0, "lt", 5.0))
	assert.False(t, g.compareValue(10.0, "lt", 5.0))
	assert.False(t, g.compareValue("hello", "lt", 5.0))
}

func TestAnswerGenerator_CompareValue_UnknownOperator(t *testing.T) {
	g := NewAnswerGenerator(nil)
	assert.False(t, g.compareValue(1.0, "unknown", 2.0))
}

func TestAnswerGenerator_PostProcess(t *testing.T) {
	g := NewAnswerGenerator(nil)
	result := g.postProcess("  hello  \r\n\r\n\r\nworld  ", &Intent{})
	assert.NotContains(t, result, "\r\n")
}

func TestAnswerGenerator_GenerateDefaultAnswer_Query(t *testing.T) {
	g := NewAnswerGenerator(nil)
	result := g.generateDefaultAnswer(&Intent{Type: IntentQuery})
	assert.Contains(t, result, "查询")
}

func TestAnswerGenerator_GenerateDefaultAnswer_Control(t *testing.T) {
	g := NewAnswerGenerator(nil)
	result := g.generateDefaultAnswer(&Intent{Type: IntentControl})
	assert.Contains(t, result, "控制")
}

func TestAnswerGenerator_GenerateDefaultAnswer_Config(t *testing.T) {
	g := NewAnswerGenerator(nil)
	result := g.generateDefaultAnswer(&Intent{Type: IntentConfig})
	assert.Contains(t, result, "配置")
}

func TestAnswerGenerator_GenerateDefaultAnswer_Diagnose(t *testing.T) {
	g := NewAnswerGenerator(nil)
	result := g.generateDefaultAnswer(&Intent{Type: IntentDiagnose})
	assert.Contains(t, result, "诊断")
}

func TestAnswerGenerator_GenerateDefaultAnswer_Unknown(t *testing.T) {
	g := NewAnswerGenerator(nil)
	result := g.generateDefaultAnswer(&Intent{Type: IntentUnknown})
	assert.Contains(t, result, "处理中")
}

func TestAnswerGenerator_GenerateFromKnowledge(t *testing.T) {
	g := NewAnswerGenerator(nil)
	items := []*KnowledgeItem{
		{ID: "1", Title: "Item 1", Content: "Main content", Relevance: 0.9},
		{ID: "2", Title: "Item 2", Content: "Secondary content", Relevance: 0.8},
	}
	result := g.generateFromKnowledge(&Intent{Type: IntentQuery}, items)
	assert.Contains(t, result, "Main content")
}

func TestAnswerGenerator_GenerateFromKnowledge_Empty(t *testing.T) {
	g := NewAnswerGenerator(nil)
	result := g.generateFromKnowledge(&Intent{Type: IntentQuery}, nil)
	assert.Contains(t, result, "没有找到")
}

func TestAnswerGenerator_GenerateReferences(t *testing.T) {
	g := NewAnswerGenerator(nil)
	items := []*KnowledgeItem{
		{ID: "1", Title: "Item 1", Content: "Content 1", Source: "kb1", Relevance: 0.9},
		{ID: "2", Title: "Item 2", Content: "Content 2", Source: "kb2", Relevance: 0.8},
	}
	refs := g.generateReferences(items)
	assert.Len(t, refs, 2)
	assert.Equal(t, "1", refs[0].SourceID)
	assert.Equal(t, 0, refs[0].Position)
}

func TestAnswerGenerator_GenerateReferences_MaxLimit(t *testing.T) {
	g := NewAnswerGenerator(&GeneratorConfig{MaxReferences: 1})
	items := []*KnowledgeItem{
		{ID: "1", Title: "Item 1", Content: "Content 1", Source: "kb1", Relevance: 0.9},
		{ID: "2", Title: "Item 2", Content: "Content 2", Source: "kb2", Relevance: 0.8},
	}
	refs := g.generateReferences(items)
	assert.Len(t, refs, 1)
}

func TestAnswerGenerator_GetSources(t *testing.T) {
	g := NewAnswerGenerator(nil)
	items := []*KnowledgeItem{
		{ID: "1", Source: "kb1", Relevance: 0.9},
		{ID: "2", Source: "kb1", Relevance: 0.8},
		{ID: "3", Source: "kb2", Relevance: 0.7},
	}
	sources := g.getSources(items)
	assert.Len(t, sources, 2)
}

func TestAnswerGenerator_GetSources_MaxLimit(t *testing.T) {
	g := NewAnswerGenerator(&GeneratorConfig{MaxSources: 1})
	items := []*KnowledgeItem{
		{ID: "1", Source: "kb1", Relevance: 0.9},
		{ID: "2", Source: "kb2", Relevance: 0.8},
	}
	sources := g.getSources(items)
	assert.Len(t, sources, 1)
}

func TestAnswerGenerator_CalculateConfidence(t *testing.T) {
	g := NewAnswerGenerator(nil)
	intent := &Intent{Confidence: 0.8}
	items := []*KnowledgeItem{
		{Relevance: 0.9},
		{Relevance: 0.7},
	}
	confidence := g.calculateConfidence(intent, items)
	assert.GreaterOrEqual(t, confidence, 0.0)
	assert.LessOrEqual(t, confidence, 1.0)
}

func TestAnswerGenerator_CalculateConfidence_NoKnowledge(t *testing.T) {
	g := NewAnswerGenerator(nil)
	intent := &Intent{Confidence: 0.8}
	confidence := g.calculateConfidence(intent, nil)
	assert.Equal(t, 0.8, confidence)
}

func TestAnswerGenerator_CalculateConfidence_OverOne(t *testing.T) {
	g := NewAnswerGenerator(nil)
	intent := &Intent{Confidence: 1.5}
	confidence := g.calculateConfidence(intent, nil)
	assert.Equal(t, 1.0, confidence)
}

func TestAnswerGenerator_CalculateConfidence_Negative(t *testing.T) {
	g := NewAnswerGenerator(nil)
	intent := &Intent{Confidence: -0.5}
	confidence := g.calculateConfidence(intent, nil)
	assert.Equal(t, 0.0, confidence)
}

func TestAnswerGenerator_BuildQuery(t *testing.T) {
	g := NewAnswerGenerator(nil)
	intent := &Intent{
		Type: IntentQuery,
		Name: "query_realtime",
		Entities: []Entity{
			{Normalized: "逆变器"},
		},
		Slots: map[string]*Slot{
			"target": {Filled: true, Value: "INV-001"},
		},
	}
	query := g.buildQuery(intent)
	assert.Contains(t, query, "query")
	assert.Contains(t, query, "query_realtime")
	assert.Contains(t, query, "逆变器")
	assert.Contains(t, query, "INV-001")
}

func TestAnswerGenerator_RenderTemplate(t *testing.T) {
	g := NewAnswerGenerator(nil)
	tmpl := &AnswerTemplate{
		Template: "设备{{.target}}状态为{{.status}}",
	}
	intent := &Intent{
		Slots: map[string]*Slot{
			"target": {Filled: true, Value: "逆变器"},
		},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{"status": "正常"}}
	result := g.renderTemplate(tmpl, intent, ctx, nil)
	assert.Contains(t, result, "逆变器")
	assert.Contains(t, result, "正常")
}

func TestAnswerGenerator_RenderTemplate_WithKnowledge(t *testing.T) {
	g := NewAnswerGenerator(nil)
	tmpl := &AnswerTemplate{
		Template: "数据：{{.data}}",
	}
	intent := &Intent{Slots: map[string]*Slot{}}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	knowledge := []*KnowledgeItem{
		{Content: "实时数据100kW", Title: "数据源"},
	}
	result := g.renderTemplate(tmpl, intent, ctx, knowledge)
	assert.Contains(t, result, "实时数据100kW")
}

func TestNewDialogueManager_NilConfig(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	require.NotNil(t, dm)
	require.NotNil(t, dm.config)
	assert.Equal(t, 50, dm.config.MaxTurns)
	assert.Equal(t, 30*time.Minute, dm.config.SessionTimeout)
	assert.Equal(t, 24*time.Hour, dm.config.MaxSessionAge)
	assert.True(t, dm.config.EnableAutoExpire)
	assert.Equal(t, 10, dm.config.ContextWindowSize)
}

func TestNewDialogueManager_CustomConfig(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	cfg := &DialogueConfig{
		MaxTurns:          20,
		SessionTimeout:    10 * time.Minute,
		MaxSessionAge:     12 * time.Hour,
		EnableAutoExpire:  false,
		ContextWindowSize: 5,
	}
	dm := NewDialogueManager(recognizer, cfg)
	require.NotNil(t, dm)
	assert.Equal(t, 20, dm.config.MaxTurns)
}

func TestDialogueManager_StartSession(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, err := dm.StartSession(context.Background(), "user1")
	require.NoError(t, err)
	require.NotNil(t, ctx)
	assert.NotEmpty(t, ctx.SessionID)
	assert.Equal(t, "user1", ctx.UserID)
	assert.Equal(t, StateInitial, ctx.CurrentState)
	assert.NotNil(t, ctx.Turns)
	assert.NotNil(t, ctx.Slots)
	assert.NotNil(t, ctx.Variables)
	assert.False(t, ctx.CreatedAt.IsZero())
	assert.False(t, ctx.UpdatedAt.IsZero())
	assert.False(t, ctx.ExpiresAt.IsZero())
}

func TestDialogueManager_GetSession_Existing(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	got, err := dm.GetSession(ctx.SessionID)
	require.NoError(t, err)
	assert.Equal(t, ctx.SessionID, got.SessionID)
}

func TestDialogueManager_GetSession_NonExisting(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	_, err := dm.GetSession("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestDialogueManager_Process_QueryIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	resp, err := dm.Process(context.Background(), ctx.SessionID, "查询逆变器的实时数据")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Content)
}

func TestDialogueManager_Process_ControlIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	resp, err := dm.Process(context.Background(), ctx.SessionID, "启动逆变器设备")
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestDialogueManager_Process_ConfigIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	resp, err := dm.Process(context.Background(), ctx.SessionID, "配置系统参数")
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestDialogueManager_Process_DiagnoseIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	resp, err := dm.Process(context.Background(), ctx.SessionID, "诊断逆变器故障")
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestDialogueManager_Process_UnknownIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	resp, err := dm.Process(context.Background(), ctx.SessionID, "随便说说")
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestDialogueManager_Process_InvalidSession(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	_, err := dm.Process(context.Background(), "nonexistent", "hello")
	assert.Error(t, err)
}

func TestDialogueManager_Process_ExpiredSession(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	dm.mu.Lock()
	ctx.ExpiresAt = time.Now().Add(-1 * time.Hour)
	dm.mu.Unlock()
	_, err := dm.Process(context.Background(), ctx.SessionID, "hello")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session expired")
}

func TestDialogueManager_Process_UpdatesState(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	_, _ = dm.Process(context.Background(), ctx.SessionID, "查询实时数据")
	got, _ := dm.GetSession(ctx.SessionID)
	assert.NotEqual(t, StateInitial, got.CurrentState)
}

func TestDialogueManager_Process_AddsTurns(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	_, _ = dm.Process(context.Background(), ctx.SessionID, "查询实时数据")
	got, _ := dm.GetSession(ctx.SessionID)
	assert.GreaterOrEqual(t, len(got.Turns), 2)
}

func TestDialogueManager_Process_ContextWindowLimit(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	cfg := &DialogueConfig{
		MaxTurns:          50,
		SessionTimeout:    30 * time.Minute,
		MaxSessionAge:     24 * time.Hour,
		EnableAutoExpire:  true,
		ContextWindowSize: 4,
	}
	dm := NewDialogueManager(recognizer, cfg)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	for i := 0; i < 5; i++ {
		_, _ = dm.Process(context.Background(), ctx.SessionID, "查询实时数据")
	}
	got, _ := dm.GetSession(ctx.SessionID)
	assert.LessOrEqual(t, len(got.Turns), 4)
}

func TestDialogueManager_EndSession(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	err := dm.EndSession(ctx.SessionID)
	assert.NoError(t, err)
	got, _ := dm.GetSession(ctx.SessionID)
	assert.Equal(t, StateCompleted, got.CurrentState)
}

func TestDialogueManager_EndSession_NonExisting(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	err := dm.EndSession("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestDialogueManager_CancelSession(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	err := dm.CancelSession(ctx.SessionID)
	assert.NoError(t, err)
	got, _ := dm.GetSession(ctx.SessionID)
	assert.Equal(t, StateCancelled, got.CurrentState)
}

func TestDialogueManager_CancelSession_NonExisting(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	err := dm.CancelSession("nonexistent")
	assert.Error(t, err)
}

func TestDialogueManager_DeleteSession(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	err := dm.DeleteSession(ctx.SessionID)
	assert.NoError(t, err)
	_, err = dm.GetSession(ctx.SessionID)
	assert.Error(t, err)
}

func TestDialogueManager_DeleteSession_NonExisting(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	err := dm.DeleteSession("nonexistent")
	assert.Error(t, err)
}

func TestDialogueManager_AddPolicy_Nil(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	err := dm.AddPolicy(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy cannot be nil")
}

func TestDialogueManager_AddPolicy_Valid(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	policy := &DialoguePolicy{
		PolicyID: "custom_policy",
		Name:     "Custom Policy",
		Conditions: []PolicyCondition{
			{Type: "intent", Key: "type", Operator: "eq", Value: IntentQuery},
		},
		Actions: []PolicyAction{
			{Type: "response", Content: "Custom response"},
		},
		Priority: 30,
	}
	err := dm.AddPolicy(policy)
	assert.NoError(t, err)
}

func TestDialogueManager_GetDialogueHistory(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	_, _ = dm.Process(context.Background(), ctx.SessionID, "查询实时数据")
	history, err := dm.GetDialogueHistory(ctx.SessionID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(history), 2)
}

func TestDialogueManager_GetDialogueHistory_NonExisting(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	_, err := dm.GetDialogueHistory("nonexistent")
	assert.Error(t, err)
}

func TestDialogueManager_SetContextVariable(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	err := dm.SetContextVariable(ctx.SessionID, "station", "station1")
	assert.NoError(t, err)
	val, err := dm.GetContextVariable(ctx.SessionID, "station")
	require.NoError(t, err)
	assert.Equal(t, "station1", val)
}

func TestDialogueManager_SetContextVariable_NonExistingSession(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	err := dm.SetContextVariable("nonexistent", "key", "value")
	assert.Error(t, err)
}

func TestDialogueManager_GetContextVariable_NonExistingSession(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	_, err := dm.GetContextVariable("nonexistent", "key")
	assert.Error(t, err)
}

func TestDialogueManager_GetContextVariable_NonExistingVariable(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	_, err := dm.GetContextVariable(ctx.SessionID, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable not found")
}

func TestDialogueManager_CleanExpiredSessions(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx1, _ := dm.StartSession(context.Background(), "user1")
	ctx2, _ := dm.StartSession(context.Background(), "user2")
	dm.mu.Lock()
	ctx1.ExpiresAt = time.Now().Add(-1 * time.Hour)
	ctx2.ExpiresAt = time.Now().Add(1 * time.Hour)
	dm.mu.Unlock()
	count := dm.CleanExpiredSessions()
	assert.Equal(t, 1, count)
	_, err := dm.GetSession(ctx1.SessionID)
	assert.Error(t, err)
	_, err = dm.GetSession(ctx2.SessionID)
	assert.NoError(t, err)
}

func TestDialogueManager_CleanExpiredSessions_NoneExpired(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dm.StartSession(context.Background(), "user1")
	count := dm.CleanExpiredSessions()
	assert.Equal(t, 0, count)
}

func TestDialogueManager_GetActiveSessions(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dm.StartSession(context.Background(), "user1")
	dm.StartSession(context.Background(), "user2")
	count := dm.GetActiveSessions()
	assert.Equal(t, 2, count)
}

func TestDialogueManager_GetActiveSessions_ExcludeCompleted(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx1, _ := dm.StartSession(context.Background(), "user1")
	dm.StartSession(context.Background(), "user2")
	dm.EndSession(ctx1.SessionID)
	count := dm.GetActiveSessions()
	assert.Equal(t, 1, count)
}

func TestDialogueManager_GetActiveSessions_ExcludeCancelled(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx1, _ := dm.StartSession(context.Background(), "user1")
	dm.StartSession(context.Background(), "user2")
	dm.CancelSession(ctx1.SessionID)
	count := dm.GetActiveSessions()
	assert.Equal(t, 1, count)
}

func TestDialogueManager_GetActiveSessions_ExcludeExpired(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx1, _ := dm.StartSession(context.Background(), "user1")
	dm.StartSession(context.Background(), "user2")
	dm.mu.Lock()
	ctx1.ExpiresAt = time.Now().Add(-1 * time.Hour)
	dm.mu.Unlock()
	count := dm.GetActiveSessions()
	assert.Equal(t, 1, count)
}

func TestDialogueManager_MergeSlots(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{
		Slots: map[string]*Slot{
			"target": {Name: "target", Value: "old", Filled: true},
		},
	}
	intent := &Intent{
		Slots: map[string]*Slot{
			"target": {Name: "target", Value: "new", Filled: true, Entities: []Entity{{Type: EntityDevice, Value: "逆变器"}}},
			"metric": {Name: "metric", Value: "power", Filled: true},
		},
	}
	dm.mergeSlots(dialogueCtx, intent)
	assert.Equal(t, "new", dialogueCtx.Slots["target"].Value)
	assert.Equal(t, "power", dialogueCtx.Slots["metric"].Value)
}

func TestDialogueManager_MatchPolicyConditions_IntentType(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	policy := &DialoguePolicy{
		Conditions: []PolicyCondition{
			{Type: "intent", Key: "type", Operator: "eq", Value: IntentQuery},
		},
	}
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{Type: IntentQuery, Name: "query_realtime", Confidence: 0.9, Slots: map[string]*Slot{}}
	assert.True(t, dm.matchPolicyConditions(policy, dialogueCtx, intent))
}

func TestDialogueManager_MatchPolicyConditions_State(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	policy := &DialoguePolicy{
		Conditions: []PolicyCondition{
			{Type: "state", Key: "state", Operator: "eq", Value: StateActive},
		},
	}
	dialogueCtx := &DialogueContext{CurrentState: StateActive, Variables: map[string]interface{}{}}
	intent := &Intent{Type: IntentQuery, Slots: map[string]*Slot{}}
	assert.True(t, dm.matchPolicyConditions(policy, dialogueCtx, intent))
}

func TestDialogueManager_MatchPolicyConditions_Context(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	policy := &DialoguePolicy{
		Conditions: []PolicyCondition{
			{Type: "context", Key: "role", Operator: "eq", Value: "admin"},
		},
	}
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{"role": "admin"}}
	intent := &Intent{Type: IntentQuery, Slots: map[string]*Slot{}}
	assert.True(t, dm.matchPolicyConditions(policy, dialogueCtx, intent))
}

func TestDialogueManager_MatchPolicyConditions_ContextNotExists(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	policy := &DialoguePolicy{
		Conditions: []PolicyCondition{
			{Type: "context", Key: "role", Operator: "eq", Value: "admin"},
		},
	}
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{Type: IntentQuery, Slots: map[string]*Slot{}}
	assert.False(t, dm.matchPolicyConditions(policy, dialogueCtx, intent))
}

func TestDialogueManager_MatchPolicyConditions_UnknownType(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	policy := &DialoguePolicy{
		Conditions: []PolicyCondition{
			{Type: "unknown_type", Key: "key", Operator: "eq", Value: "val"},
		},
	}
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{Type: IntentQuery, Slots: map[string]*Slot{}}
	assert.False(t, dm.matchPolicyConditions(policy, dialogueCtx, intent))
}

func TestDialogueManager_CompareValue_Eq(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	assert.True(t, dm.compareValue("hello", "eq", "hello"))
	assert.False(t, dm.compareValue("hello", "eq", "world"))
}

func TestDialogueManager_CompareValue_Ne(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	assert.True(t, dm.compareValue("hello", "ne", "world"))
	assert.False(t, dm.compareValue("hello", "ne", "hello"))
}

func TestDialogueManager_CompareValue_Exists(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	assert.True(t, dm.compareValue("something", "exists", true))
	assert.False(t, dm.compareValue(nil, "exists", true))
}

func TestDialogueManager_CompareValue_NotExists(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	assert.True(t, dm.compareValue(nil, "not_exists", true))
	assert.False(t, dm.compareValue("something", "not_exists", true))
}

func TestDialogueManager_CompareValue_UnknownOperator(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	assert.False(t, dm.compareValue("val", "unknown", "val"))
}

func TestDialogueManager_GetIntentValue(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	intent := &Intent{Type: IntentQuery, Name: "query_realtime", Confidence: 0.9}
	assert.Equal(t, IntentQuery, dm.getIntentValue("type", intent))
	assert.Equal(t, "query_realtime", dm.getIntentValue("name", intent))
	assert.Equal(t, 0.9, dm.getIntentValue("confidence", intent))
	assert.Nil(t, dm.getIntentValue("other", intent))
}

func TestDialogueManager_GetIntentValue_NilIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	assert.Nil(t, dm.getIntentValue("type", nil))
}

func TestDialogueManager_GetSlotValue_RequiredMissing(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Slots: map[string]*Slot{}}
	intent := &Intent{
		Type: IntentQuery,
		Slots: map[string]*Slot{
			"target": {Name: "target", Required: true, Filled: false},
		},
	}
	val := dm.getSlotValue("required_missing", dialogueCtx, intent)
	assert.True(t, val.(bool))
}

func TestDialogueManager_GetSlotValue_FromDialogueCtx(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{
		Slots: map[string]*Slot{
			"target": {Name: "target", Value: "逆变器"},
		},
	}
	intent := &Intent{Slots: map[string]*Slot{}}
	val := dm.getSlotValue("target", dialogueCtx, intent)
	assert.Equal(t, "逆变器", val)
}

func TestDialogueManager_GetSlotValue_FromIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Slots: map[string]*Slot{}}
	intent := &Intent{
		Slots: map[string]*Slot{
			"target": {Name: "target", Value: "变压器"},
		},
	}
	val := dm.getSlotValue("target", dialogueCtx, intent)
	assert.Equal(t, "变压器", val)
}

func TestDialogueManager_GenerateResponse_QueryResponse(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{
		Name: "query_realtime",
		Slots: map[string]*Slot{
			"target": {Filled: true, Value: "逆变器"},
		},
	}
	result := dm.generateResponse("query_response", dialogueCtx, intent)
	assert.Contains(t, result, "逆变器")
}

func TestDialogueManager_GenerateResponse_QueryHistory(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{
		Name: "query_history",
		Slots: map[string]*Slot{
			"target": {Filled: true, Value: "变压器"},
		},
	}
	result := dm.generateResponse("query_response", dialogueCtx, intent)
	assert.Contains(t, result, "变压器")
}

func TestDialogueManager_GenerateResponse_QueryStatistics(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{
		Name: "query_statistics",
		Slots: map[string]*Slot{
			"target": {Filled: true, Value: "储能"},
		},
	}
	result := dm.generateResponse("query_response", dialogueCtx, intent)
	assert.Contains(t, result, "储能")
}

func TestDialogueManager_GenerateResponse_QueryDefault(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{
		Name:   "query_other",
		Slots:  map[string]*Slot{},
	}
	result := dm.generateResponse("query_response", dialogueCtx, intent)
	assert.NotEmpty(t, result)
}

func TestDialogueManager_GenerateResponse_ControlConfirm(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{
		Slots: map[string]*Slot{
			"device": {Filled: true, Value: "逆变器"},
			"action": {Filled: true, Value: "启动"},
		},
	}
	result := dm.generateResponse("control_confirm", dialogueCtx, intent)
	assert.Contains(t, result, "启动")
	assert.Contains(t, result, "逆变器")
}

func TestDialogueManager_GenerateResponse_ControlActionOnly(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{
		Slots: map[string]*Slot{
			"action": {Filled: true, Value: "启动"},
		},
	}
	result := dm.generateResponse("control_confirm", dialogueCtx, intent)
	assert.Contains(t, result, "启动")
}

func TestDialogueManager_GenerateResponse_ControlNoSlots(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{Slots: map[string]*Slot{}}
	result := dm.generateResponse("control_confirm", dialogueCtx, intent)
	assert.NotEmpty(t, result)
}

func TestDialogueManager_GenerateResponse_ConfigGuide(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{Slots: map[string]*Slot{}}
	result := dm.generateResponse("config_guide", dialogueCtx, intent)
	assert.NotEmpty(t, result)
}

func TestDialogueManager_GenerateResponse_DiagnoseAnalysis(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{
		Slots: map[string]*Slot{
			"target": {Filled: true, Value: "逆变器"},
		},
	}
	result := dm.generateResponse("diagnose_analysis", dialogueCtx, intent)
	assert.Contains(t, result, "逆变器")
}

func TestDialogueManager_GenerateResponse_DiagnoseNoTarget(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{Slots: map[string]*Slot{}}
	result := dm.generateResponse("diagnose_analysis", dialogueCtx, intent)
	assert.Contains(t, result, "设备")
}

func TestDialogueManager_GenerateResponse_SlotPrompt(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{
		Type: IntentQuery,
		Name: "query_realtime",
		Slots: map[string]*Slot{
			"target": {Name: "target", Required: true, Filled: false},
		},
	}
	result := dm.generateResponse("slot_prompt", dialogueCtx, intent)
	assert.NotEmpty(t, result)
}

func TestDialogueManager_GenerateResponse_UnknownIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{Slots: map[string]*Slot{}}
	result := dm.generateResponse("unknown_intent", dialogueCtx, intent)
	assert.NotEmpty(t, result)
}

func TestDialogueManager_GenerateResponse_Default(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{Slots: map[string]*Slot{}}
	result := dm.generateResponse("nonexistent_template", dialogueCtx, intent)
	assert.NotEmpty(t, result)
}

func TestDialogueManager_GenerateSuggestions_Target(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	suggestions := dm.generateSuggestions(&Intent{}, "target")
	assert.NotEmpty(t, suggestions)
}

func TestDialogueManager_GenerateSuggestions_StartTime(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	suggestions := dm.generateSuggestions(&Intent{}, "startTime")
	assert.NotEmpty(t, suggestions)
}

func TestDialogueManager_GenerateSuggestions_Metric(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	suggestions := dm.generateSuggestions(&Intent{}, "metric")
	assert.NotEmpty(t, suggestions)
}

func TestDialogueManager_GenerateSuggestions_Threshold(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	suggestions := dm.generateSuggestions(&Intent{}, "threshold")
	assert.NotEmpty(t, suggestions)
}

func TestDialogueManager_GenerateSuggestions_UnknownSlot(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	suggestions := dm.generateSuggestions(&Intent{}, "unknown")
	assert.Empty(t, suggestions)
}

func TestDialogueManager_HandleMissingSlots_Empty(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{Confidence: 0.8, Slots: map[string]*Slot{}}
	resp := dm.handleMissingSlots(dialogueCtx, intent, []string{})
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Content)
}

func TestDialogueManager_HandleMissingSlots_WithMissing(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "query_realtime",
		Confidence: 0.8,
		Slots: map[string]*Slot{
			"target": {Name: "target", Required: true, Filled: false},
		},
	}
	resp := dm.handleMissingSlots(dialogueCtx, intent, []string{"target"})
	assert.NotNil(t, resp)
	assert.True(t, resp.RequiresMore)
	assert.NotNil(t, resp.Metadata)
}

func TestDialogueManager_HandleMissingSlots_NoPrompt(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dialogueCtx := &DialogueContext{Variables: map[string]interface{}{}}
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "custom_intent_no_pattern",
		Confidence: 0.8,
		Slots: map[string]*Slot{
			"custom_slot": {Name: "custom_slot", Required: true, Filled: false},
		},
	}
	resp := dm.handleMissingSlots(dialogueCtx, intent, []string{"custom_slot"})
	assert.NotNil(t, resp)
	assert.Contains(t, resp.Content, "custom_slot")
}

func TestDialogueManager_MultipleSessions(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx1, _ := dm.StartSession(context.Background(), "user1")
	ctx2, _ := dm.StartSession(context.Background(), "user2")
	assert.NotEqual(t, ctx1.SessionID, ctx2.SessionID)
	assert.Equal(t, 2, dm.GetActiveSessions())
}

func TestDialogueManager_Process_MultipleTurns(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	_, _ = dm.Process(context.Background(), ctx.SessionID, "查询实时数据")
	_, _ = dm.Process(context.Background(), ctx.SessionID, "启动逆变器设备")
	history, _ := dm.GetDialogueHistory(ctx.SessionID)
	assert.GreaterOrEqual(t, len(history), 4)
}

func TestGenerateSessionID(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "session_")
}

func TestGenerateTurnID(t *testing.T) {
	id := generateTurnID()
	assert.Contains(t, id, "turn_")
}

func TestSlotDefinition_Struct(t *testing.T) {
	sd := &SlotDefinition{
		Name:       "target",
		Type:       "string",
		Required:   true,
		Prompts:    []string{"请输入目标"},
		Default:    "default_value",
		EntityType: EntityDevice,
	}
	assert.Equal(t, "target", sd.Name)
	assert.True(t, sd.Required)
	assert.Equal(t, EntityDevice, sd.EntityType)
}

func TestEntityRule_Struct(t *testing.T) {
	rule := &EntityRule{
		Type:       EntityDevice,
		Pattern:    regexp.MustCompile(`test`),
		Dictionary: []string{"test"},
		Normalizer: func(s string) string { return s },
	}
	assert.Equal(t, EntityDevice, rule.Type)
	assert.NotNil(t, rule.Pattern)
}

func TestRecognizerConfig_Default(t *testing.T) {
	cfg := DefaultRecognizerConfig()
	assert.Equal(t, 0.6, cfg.MinConfidence)
	assert.Equal(t, 20, cfg.MaxEntities)
	assert.True(t, cfg.EnableSubIntents)
	assert.True(t, cfg.CacheEnabled)
	assert.Equal(t, 5*time.Minute, cfg.CacheTTL)
}

func TestGeneratorConfig_Default(t *testing.T) {
	cfg := DefaultGeneratorConfig()
	assert.Equal(t, 0.5, cfg.MinConfidence)
	assert.Equal(t, 5, cfg.MaxReferences)
	assert.Equal(t, 3, cfg.MaxSources)
	assert.True(t, cfg.EnableCache)
	assert.Equal(t, 10*time.Minute, cfg.CacheTTL)
	assert.True(t, cfg.TemplatePriority)
}

func TestDialogueConfig_Default(t *testing.T) {
	cfg := DefaultDialogueConfig()
	assert.Equal(t, 50, cfg.MaxTurns)
	assert.Equal(t, 30*time.Minute, cfg.SessionTimeout)
	assert.Equal(t, 24*time.Hour, cfg.MaxSessionAge)
	assert.True(t, cfg.EnableAutoExpire)
	assert.Equal(t, 10, cfg.ContextWindowSize)
}

type mockKnowledgeProvider struct {
	items []*KnowledgeItem
	err   error
}

func (m *mockKnowledgeProvider) Query(ctx context.Context, query string, limit int) ([]*KnowledgeItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	if limit > len(m.items) {
		limit = len(m.items)
	}
	return m.items[:limit], nil
}

func (m *mockKnowledgeProvider) GetByID(ctx context.Context, id string) (*KnowledgeItem, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockKnowledgeProvider) GetRelated(ctx context.Context, id string, limit int) ([]*KnowledgeItem, error) {
	return m.items, nil
}
