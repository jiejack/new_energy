package qa

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntentRecognizer_QueryRealtime(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	result, err := recognizer.Recognize(context.Background(), "查询逆变器的实时数据")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestIntentRecognizer_QueryHistory(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	result, err := recognizer.Recognize(context.Background(), "查询昨天的历史数据")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestIntentRecognizer_QueryStatistics(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	result, err := recognizer.Recognize(context.Background(), "统计平均值数据")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestIntentRecognizer_ControlDevice(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	result, err := recognizer.Recognize(context.Background(), "启动设备逆变器")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestIntentRecognizer_ControlThreshold(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	result, err := recognizer.Recognize(context.Background(), "设置告警阈值")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestIntentRecognizer_ConfigSystem(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	result, err := recognizer.Recognize(context.Background(), "配置系统参数")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestIntentRecognizer_ConfigAlarm(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	result, err := recognizer.Recognize(context.Background(), "配置告警规则")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestIntentRecognizer_DiagnoseFault(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	result, err := recognizer.Recognize(context.Background(), "诊断设备故障")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestIntentRecognizer_DiagnosePerformance(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	result, err := recognizer.Recognize(context.Background(), "分析设备性能")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestIntentRecognizer_EmptyText(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	_, err := recognizer.Recognize(context.Background(), "")
	assert.Error(t, err)
}

func TestIntentRecognizer_AddPattern(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	err := recognizer.AddPattern(&IntentPattern{
		IntentType: IntentQuery,
		IntentName: "custom_query",
		Keywords:   []string{"自定义"},
		Priority:   5,
	})
	require.NoError(t, err)
	patterns := recognizer.GetIntentPatterns(IntentQuery)
	found := false
	for _, p := range patterns {
		if p.IntentName == "custom_query" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestIntentRecognizer_AddPattern_Nil(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	err := recognizer.AddPattern(nil)
	assert.Error(t, err)
}

func TestIntentRecognizer_AddEntityRule(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	err := recognizer.AddEntityRule(EntityDevice, &EntityRule{
		Type:       EntityDevice,
		Dictionary: []string{"自定义设备"},
	})
	require.NoError(t, err)
	rules := recognizer.GetEntityRules(EntityDevice)
	assert.NotEmpty(t, rules)
}

func TestIntentRecognizer_AddEntityRule_Nil(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	err := recognizer.AddEntityRule(EntityDevice, nil)
	assert.Error(t, err)
}

func TestIntentRecognizer_GetSlotPrompt(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	intent, _ := recognizer.Recognize(context.Background(), "查询逆变器实时数据")
	if intent != nil {
		prompt := recognizer.GetSlotPrompt(intent, "target")
		_ = prompt
	}
}

func TestIntentRecognizer_GetSlotPrompt_NilIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	prompt := recognizer.GetSlotPrompt(nil, "target")
	assert.Empty(t, prompt)
}

func TestIntentRecognizer_GetSlotPrompt_UnknownSlot(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	intent, _ := recognizer.Recognize(context.Background(), "查询实时数据")
	prompt := recognizer.GetSlotPrompt(intent, "nonexistent_slot")
	assert.Empty(t, prompt)
}

func TestIntentRecognizer_ValidateIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "test",
		Slots:      map[string]*Slot{"device": {Name: "device", Required: true, Filled: true}},
	}
	err := recognizer.ValidateIntent(intent)
	assert.NoError(t, err)
}

func TestIntentRecognizer_ValidateIntent_MissingSlot(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "test",
		Slots:      map[string]*Slot{"device": {Name: "device", Required: true, Filled: false}},
	}
	err := recognizer.ValidateIntent(intent)
	assert.Error(t, err)
}

func TestIntentRecognizer_ValidateIntent_Nil(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	err := recognizer.ValidateIntent(nil)
	assert.Error(t, err)
}

func TestIntentRecognizer_GetMissingSlots(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	intent := &Intent{
		Type:  IntentQuery,
		Name:  "test",
		Slots: map[string]*Slot{
			"device":  {Name: "device", Required: true, Filled: false},
			"metric":  {Name: "metric", Required: false, Filled: false},
		},
	}
	missing := recognizer.GetMissingSlots(intent)
	assert.Len(t, missing, 1)
	assert.Contains(t, missing, "device")
}

func TestIntentRecognizer_GetMissingSlots_Nil(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	missing := recognizer.GetMissingSlots(nil)
	assert.Empty(t, missing)
}

func TestIntentRecognizer_BatchRecognize(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	texts := []string{"查询设备状态", "启动逆变器", "随机文本"}
	results, err := recognizer.BatchRecognize(context.Background(), texts)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestIntentRecognizer_BatchRecognize_WithEmpty(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	texts := []string{"查询设备状态", ""}
	_, err := recognizer.BatchRecognize(context.Background(), texts)
	assert.Error(t, err)
}

func TestIntentRecognizer_GetIntentPatterns_NotFound(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	patterns := recognizer.GetIntentPatterns("nonexistent")
	assert.Nil(t, patterns)
}

func TestIntentRecognizer_GetEntityRules_NotFound(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	rules := recognizer.GetEntityRules("nonexistent")
	assert.Nil(t, rules)
}

func TestDefaultRecognizerConfig(t *testing.T) {
	cfg := DefaultRecognizerConfig()
	assert.Equal(t, 0.6, cfg.MinConfidence)
	assert.Equal(t, 20, cfg.MaxEntities)
	assert.True(t, cfg.EnableSubIntents)
	assert.True(t, cfg.CacheEnabled)
	assert.Equal(t, 5*time.Minute, cfg.CacheTTL)
}

func TestDeduplicateEntities(t *testing.T) {
	entities := []Entity{
		{Type: EntityDevice, Value: "inv", Position: Position{Start: 0, End: 3}},
		{Type: EntityDevice, Value: "inv", Position: Position{Start: 0, End: 3}},
		{Type: EntityPoint, Value: "voltage", Position: Position{Start: 5, End: 12}},
	}
	deduped := deduplicateEntities(entities)
	assert.Len(t, deduped, 2)
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
	assert.Equal(t, "othertext", normalizeTime("othertext"))
}

func TestDialogueManager_StartSession(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, err := dm.StartSession(context.Background(), "user1")
	require.NoError(t, err)
	assert.NotEmpty(t, ctx.SessionID)
	assert.Equal(t, "user1", ctx.UserID)
	assert.Equal(t, StateInitial, ctx.CurrentState)
}

func TestDialogueManager_Process_QueryIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	resp, err := dm.Process(context.Background(), ctx.SessionID, "查询逆变器实时数据")
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Content)
}

func TestDialogueManager_Process_ControlIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	resp, err := dm.Process(context.Background(), ctx.SessionID, "启动设备逆变器")
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDialogueManager_Process_ConfigIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	resp, err := dm.Process(context.Background(), ctx.SessionID, "配置系统参数")
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDialogueManager_Process_DiagnoseIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	resp, err := dm.Process(context.Background(), ctx.SessionID, "诊断逆变器故障")
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDialogueManager_Process_UnknownIntent(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	resp, err := dm.Process(context.Background(), ctx.SessionID, "随机文本")
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDialogueManager_Process_SessionNotFound(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	_, err := dm.Process(context.Background(), "nonexistent", "test")
	assert.Error(t, err)
}

func TestDialogueManager_EndSession(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	err := dm.EndSession(ctx.SessionID)
	require.NoError(t, err)
	got, _ := dm.GetSession(ctx.SessionID)
	assert.Equal(t, StateCompleted, got.CurrentState)
}

func TestDialogueManager_EndSession_NotFound(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	err := dm.EndSession("nonexistent")
	assert.Error(t, err)
}

func TestDialogueManager_CancelSession(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	err := dm.CancelSession(ctx.SessionID)
	require.NoError(t, err)
	got, _ := dm.GetSession(ctx.SessionID)
	assert.Equal(t, StateCancelled, got.CurrentState)
}

func TestDialogueManager_CancelSession_NotFound(t *testing.T) {
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
	require.NoError(t, err)
	_, err = dm.GetSession(ctx.SessionID)
	assert.Error(t, err)
}

func TestDialogueManager_DeleteSession_NotFound(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	err := dm.DeleteSession("nonexistent")
	assert.Error(t, err)
}

func TestDialogueManager_GetDialogueHistory(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	dm.Process(context.Background(), ctx.SessionID, "查询设备状态")
	history, err := dm.GetDialogueHistory(ctx.SessionID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(history), 2)
}

func TestDialogueManager_GetDialogueHistory_NotFound(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	_, err := dm.GetDialogueHistory("nonexistent")
	assert.Error(t, err)
}

func TestDialogueManager_SetContextVariable(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	err := dm.SetContextVariable(ctx.SessionID, "station", "ST001")
	require.NoError(t, err)
}

func TestDialogueManager_SetContextVariable_NotFound(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	err := dm.SetContextVariable("nonexistent", "key", "val")
	assert.Error(t, err)
}

func TestDialogueManager_GetContextVariable(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	dm.SetContextVariable(ctx.SessionID, "station", "ST001")
	val, err := dm.GetContextVariable(ctx.SessionID, "station")
	require.NoError(t, err)
	assert.Equal(t, "ST001", val)
}

func TestDialogueManager_GetContextVariable_NotFound(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	_, err := dm.GetContextVariable("nonexistent", "key")
	assert.Error(t, err)
}

func TestDialogueManager_GetContextVariable_KeyNotFound(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	_, err := dm.GetContextVariable(ctx.SessionID, "nonexistent_key")
	assert.Error(t, err)
}

func TestDialogueManager_CleanExpiredSessions(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	cfg := DefaultDialogueConfig()
	cfg.MaxSessionAge = 1 * time.Nanosecond
	dm := NewDialogueManager(recognizer, cfg)
	dm.StartSession(context.Background(), "user1")
	time.Sleep(10 * time.Millisecond)
	count := dm.CleanExpiredSessions()
	assert.Equal(t, 1, count)
}

func TestDialogueManager_GetActiveSessions(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dm.StartSession(context.Background(), "user1")
	dm.StartSession(context.Background(), "user2")
	assert.Equal(t, 2, dm.GetActiveSessions())
}

func TestDialogueManager_AddPolicy(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	err := dm.AddPolicy(&DialoguePolicy{
		PolicyID: "custom-policy",
		Name:     "Custom",
		Conditions: []PolicyCondition{
			{Type: "intent", Key: "type", Operator: "eq", Value: IntentQuery},
		},
		Actions:  []PolicyAction{{Type: "response", Content: "custom response"}},
		Priority: 100,
	})
	require.NoError(t, err)
}

func TestDialogueManager_AddPolicy_Nil(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	err := dm.AddPolicy(nil)
	assert.Error(t, err)
}

func TestDialogueManager_MatchPolicyConditions_State(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dm.AddPolicy(&DialoguePolicy{
		PolicyID: "state-policy",
		Name:     "State Policy",
		Conditions: []PolicyCondition{
			{Type: "state", Key: "state", Operator: "eq", Value: StateActive},
		},
		Actions:  []PolicyAction{{Type: "response", Content: "state response"}},
		Priority: 100,
	})
	ctx, _ := dm.StartSession(context.Background(), "user1")
	dm.Process(context.Background(), ctx.SessionID, "查询实时数据")
}

func TestDialogueManager_MatchPolicyConditions_Context(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	dm.AddPolicy(&DialoguePolicy{
		PolicyID: "ctx-policy",
		Name:     "Context Policy",
		Conditions: []PolicyCondition{
			{Type: "context", Key: "env", Operator: "eq", Value: "production"},
		},
		Actions:  []PolicyAction{{Type: "response", Content: "ctx response"}},
		Priority: 100,
	})
	sessCtx, _ := dm.StartSession(context.Background(), "user1")
	dm.SetContextVariable(sessCtx.SessionID, "env", "production")
	dm.Process(context.Background(), sessCtx.SessionID, "查询实时数据")
}

func TestDefaultDialogueConfig(t *testing.T) {
	cfg := DefaultDialogueConfig()
	assert.Equal(t, 50, cfg.MaxTurns)
	assert.Equal(t, 30*time.Minute, cfg.SessionTimeout)
	assert.Equal(t, 24*time.Hour, cfg.MaxSessionAge)
	assert.True(t, cfg.EnableAutoExpire)
	assert.Equal(t, 10, cfg.ContextWindowSize)
}

func TestAnswerGenerator_Generate_NilIntent(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	_, err := generator.Generate(context.Background(), nil, &DialogueContext{})
	assert.Error(t, err)
}

func TestAnswerGenerator_Generate_WithTemplate(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "query_realtime",
		Confidence: 0.9,
		Slots: map[string]*Slot{
			"target": {Name: "target", Value: "逆变器INV-001", Filled: true},
		},
	}
	ctx := &DialogueContext{
		Variables: map[string]interface{}{
			"updateTime": time.Now().Format("15:04:05"),
		},
	}
	answer, err := generator.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, answer.Content)
	assert.Greater(t, answer.Confidence, 0.0)
}

func TestAnswerGenerator_Generate_NoTemplate(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "custom_unknown",
		Confidence: 0.8,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := generator.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, answer.Content)
}

func TestAnswerGenerator_GenerateWithFallback(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "test",
		Confidence: 0.1,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := generator.GenerateWithFallback(context.Background(), intent, ctx)
	require.NoError(t, err)
	assert.NotNil(t, answer)
}

func TestAnswerGenerator_BatchGenerate(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	intents := []*Intent{
		{Type: IntentQuery, Name: "test1", Confidence: 0.9, Slots: map[string]*Slot{}},
		{Type: IntentControl, Name: "test2", Confidence: 0.8, Slots: map[string]*Slot{}},
	}
	contexts := []*DialogueContext{
		{Variables: map[string]interface{}{}},
		{Variables: map[string]interface{}{}},
	}
	answers, err := generator.BatchGenerate(context.Background(), intents, contexts)
	require.NoError(t, err)
	assert.Len(t, answers, 2)
}

func TestAnswerGenerator_BatchGenerate_Mismatch(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	intents := []*Intent{{Type: IntentQuery, Name: "test", Confidence: 0.9, Slots: map[string]*Slot{}}}
	contexts := []*DialogueContext{{}, {}}
	_, err := generator.BatchGenerate(context.Background(), intents, contexts)
	assert.Error(t, err)
}

func TestAnswerGenerator_ValidateAnswer_Nil(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	err := generator.ValidateAnswer(nil)
	assert.Error(t, err)
}

func TestAnswerGenerator_ValidateAnswer_EmptyContent(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	err := generator.ValidateAnswer(&Answer{Content: "", Confidence: 0.9})
	assert.Error(t, err)
}

func TestAnswerGenerator_ValidateAnswer_InvalidConfidence(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	err := generator.ValidateAnswer(&Answer{Content: "test", Confidence: 1.5})
	assert.Error(t, err)
}

func TestAnswerGenerator_ValidateAnswer_Valid(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	err := generator.ValidateAnswer(&Answer{Content: "test", Confidence: 0.9})
	assert.NoError(t, err)
}

func TestAnswerGenerator_EnhanceAnswer(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	answer := &Answer{Content: "test content", Confidence: 0.9}
	intent := &Intent{Type: IntentQuery, Name: "test", Confidence: 0.9, Slots: map[string]*Slot{}}
	enhanced, err := generator.EnhanceAnswer(context.Background(), answer, intent)
	require.NoError(t, err)
	assert.Contains(t, enhanced.Content, "test content")
}

func TestAnswerGenerator_EnhanceAnswer_Nil(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	intent := &Intent{Type: IntentQuery, Name: "test", Confidence: 0.9, Slots: map[string]*Slot{}}
	_, err := generator.EnhanceAnswer(context.Background(), nil, intent)
	assert.Error(t, err)
}

func TestAnswerGenerator_AddTemplate(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	err := generator.AddTemplate(&AnswerTemplate{
		TemplateID: "custom-tpl",
		Name:       "Custom",
		IntentName: "custom_intent",
		Template:   "Custom: {{.value}}",
		Variables:  []string{"value"},
		Priority:   10,
	})
	require.NoError(t, err)
	templates := generator.GetTemplates("custom_intent")
	assert.Len(t, templates, 1)
}

func TestAnswerGenerator_AddTemplate_Nil(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	err := generator.AddTemplate(nil)
	assert.Error(t, err)
}

func TestAnswerGenerator_GetTemplates_NotFound(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	templates := generator.GetTemplates("nonexistent")
	assert.Nil(t, templates)
}

func TestAnswerGenerator_RegisterKnowledgeProvider(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	err := generator.RegisterKnowledgeProvider("test", &MockKnowledgeProvider{})
	require.NoError(t, err)
}

func TestAnswerGenerator_RegisterKnowledgeProvider_Nil(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	err := generator.RegisterKnowledgeProvider("test", nil)
	assert.Error(t, err)
}

func TestAnswerGenerator_Generate_WithKnowledge(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	generator.RegisterKnowledgeProvider("test", &MockKnowledgeProvider{})
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "custom_unknown",
		Confidence: 0.9,
		Slots:      map[string]*Slot{},
		Entities:   []Entity{{Type: EntityDevice, Normalized: "INV-001"}},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := generator.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, answer.Content)
}

func TestAnswerGenerator_Generate_ControlIntent(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	intent := &Intent{
		Type:       IntentControl,
		Name:       "control_device",
		Confidence: 0.9,
		Slots: map[string]*Slot{
			"device": {Name: "device", Value: "INV-001", Filled: true},
			"action": {Name: "action", Value: "启动", Filled: true},
		},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := generator.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, answer.Content)
}

func TestAnswerGenerator_Generate_ConfigIntent(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	intent := &Intent{
		Type:       IntentConfig,
		Name:       "config_system",
		Confidence: 0.9,
		Slots:      map[string]*Slot{},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := generator.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, answer.Content)
}

func TestAnswerGenerator_Generate_DiagnoseIntent(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	intent := &Intent{
		Type:       IntentDiagnose,
		Name:       "diagnose_fault",
		Confidence: 0.9,
		Slots: map[string]*Slot{
			"target": {Name: "target", Value: "INV-001", Filled: true},
		},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	answer, err := generator.Generate(context.Background(), intent, ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, answer.Content)
}

func TestAnswerGenerator_DefaultGeneratorConfig(t *testing.T) {
	cfg := DefaultGeneratorConfig()
	assert.Equal(t, 0.5, cfg.MinConfidence)
	assert.Equal(t, 5, cfg.MaxReferences)
	assert.Equal(t, 3, cfg.MaxSources)
	assert.True(t, cfg.EnableCache)
	assert.Equal(t, 10*time.Minute, cfg.CacheTTL)
	assert.True(t, cfg.TemplatePriority)
}

func TestAnswerGenerator_CompareValue(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	assert.True(t, generator.compareValue("a", "eq", "a"))
	assert.False(t, generator.compareValue("a", "eq", "b"))
	assert.True(t, generator.compareValue("a", "ne", "b"))
	assert.False(t, generator.compareValue(1.0, "gt", 2.0))
	assert.True(t, generator.compareValue(2.0, "gt", 1.0))
	assert.True(t, generator.compareValue(1.0, "lt", 2.0))
	assert.False(t, generator.compareValue(2.0, "lt", 1.0))
	assert.False(t, generator.compareValue("a", "unknown", "b"))
}

func TestAnswerGenerator_MatchTemplateConditions(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	tmpl := &AnswerTemplate{
		Conditions: []TemplateCondition{
			{Variable: "success", Operator: "eq", Value: false},
		},
	}
	ctx := &DialogueContext{Variables: map[string]interface{}{"success": false}}
	assert.True(t, generator.matchTemplateConditions(tmpl, ctx))
	ctx2 := &DialogueContext{Variables: map[string]interface{}{"success": true}}
	assert.False(t, generator.matchTemplateConditions(tmpl, ctx2))
	ctx3 := &DialogueContext{Variables: map[string]interface{}{}}
	assert.False(t, generator.matchTemplateConditions(tmpl, ctx3))
}

func TestAnswerGenerator_MatchTemplateConditions_NoConditions(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	tmpl := &AnswerTemplate{}
	ctx := &DialogueContext{Variables: map[string]interface{}{}}
	assert.True(t, generator.matchTemplateConditions(tmpl, ctx))
}

func TestAnswerGenerator_EnhanceAnswer_AllTypes(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	answer := &Answer{Content: "test", Confidence: 0.9}
	types := []IntentType{IntentQuery, IntentControl, IntentConfig, IntentDiagnose, IntentUnknown}
	for _, it := range types {
		intent := &Intent{Type: it, Name: "test", Confidence: 0.9, Slots: map[string]*Slot{}}
		enhanced, err := generator.EnhanceAnswer(context.Background(), answer, intent)
		require.NoError(t, err)
		assert.NotEmpty(t, enhanced.Content)
	}
}

type MockKnowledgeProvider struct{}

func (m *MockKnowledgeProvider) Query(ctx context.Context, query string, limit int) ([]*KnowledgeItem, error) {
	return []*KnowledgeItem{
		{
			ID:        "ki1",
			Title:     "Test Knowledge",
			Content:   "This is test knowledge content",
			Category:  "test",
			Relevance: 0.9,
			Source:    "mock",
		},
	}, nil
}

func (m *MockKnowledgeProvider) GetByID(ctx context.Context, id string) (*KnowledgeItem, error) {
	return &KnowledgeItem{ID: id, Title: "Test", Content: "Content"}, nil
}

func (m *MockKnowledgeProvider) GetRelated(ctx context.Context, id string, limit int) ([]*KnowledgeItem, error) {
	return []*KnowledgeItem{}, nil
}

func TestDialogueManager_Process_ExpiredSession(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	cfg := DefaultDialogueConfig()
	cfg.MaxSessionAge = 1 * time.Nanosecond
	dm := NewDialogueManager(recognizer, cfg)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	time.Sleep(10 * time.Millisecond)
	_, err := dm.Process(context.Background(), ctx.SessionID, "查询设备状态")
	assert.Error(t, err)
}

func TestDialogueManager_Process_ContextWindowLimit(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	cfg := DefaultDialogueConfig()
	cfg.ContextWindowSize = 2
	dm := NewDialogueManager(recognizer, cfg)
	ctx, _ := dm.StartSession(context.Background(), "user1")
	for i := 0; i < 5; i++ {
		dm.Process(context.Background(), ctx.SessionID, fmt.Sprintf("查询设备状态%d", i))
	}
	history, _ := dm.GetDialogueHistory(ctx.SessionID)
	assert.LessOrEqual(t, len(history), 2)
}

func TestSlotDefinition_Struct(t *testing.T) {
	sd := &SlotDefinition{
		Name:       "device",
		Type:       "string",
		Required:   true,
		Prompts:    []string{"请输入设备名称"},
		Default:    "default_device",
		EntityType: EntityDevice,
	}
	assert.Equal(t, "device", sd.Name)
	assert.True(t, sd.Required)
}

func TestEntityRule_Struct(t *testing.T) {
	rule := &EntityRule{
		Type:       EntityDevice,
		Dictionary: []string{"逆变器"},
	}
	assert.Equal(t, EntityDevice, rule.Type)
}

func TestRecognizerConfig_Defaults(t *testing.T) {
	cfg := DefaultRecognizerConfig()
	assert.Equal(t, 0.6, cfg.MinConfidence)
}

func TestGeneratorConfig_Defaults(t *testing.T) {
	cfg := DefaultGeneratorConfig()
	assert.Equal(t, 0.5, cfg.MinConfidence)
}
