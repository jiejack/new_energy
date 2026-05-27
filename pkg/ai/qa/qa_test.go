package qa

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntentType_Constants(t *testing.T) {
	assert.Equal(t, IntentType("query"), IntentQuery)
	assert.Equal(t, IntentType("control"), IntentControl)
	assert.Equal(t, IntentType("config"), IntentConfig)
	assert.Equal(t, IntentType("diagnose"), IntentDiagnose)
	assert.Equal(t, IntentType("unknown"), IntentUnknown)
}

func TestEntityType_Constants(t *testing.T) {
	assert.Equal(t, EntityType("device"), EntityDevice)
	assert.Equal(t, EntityType("point"), EntityPoint)
	assert.Equal(t, EntityType("station"), EntityStation)
	assert.Equal(t, EntityType("time"), EntityTime)
	assert.Equal(t, EntityType("metric"), EntityMetric)
	assert.Equal(t, EntityType("threshold"), EntityThreshold)
	assert.Equal(t, EntityType("status"), EntityStatus)
}

func TestEntity_Struct(t *testing.T) {
	e := Entity{
		Type:       EntityDevice,
		Value:      "inverter-001",
		Normalized: "INV001",
		Position:   Position{Start: 0, End: 12},
	}
	assert.Equal(t, EntityDevice, e.Type)
	assert.Equal(t, "inverter-001", e.Value)
}

func TestPosition_Struct(t *testing.T) {
	p := Position{Start: 5, End: 10}
	assert.Equal(t, 5, p.Start)
	assert.Equal(t, 10, p.End)
}

func TestSlot_Struct(t *testing.T) {
	s := &Slot{
		Name:     "device_id",
		Value:    "dev1",
		Required: true,
		Filled:   true,
	}
	assert.Equal(t, "device_id", s.Name)
	assert.True(t, s.Required)
	assert.True(t, s.Filled)
}

func TestIntent_Struct(t *testing.T) {
	intent := &Intent{
		Type:       IntentQuery,
		Name:       "query_device_status",
		Confidence: 0.95,
		Entities: []Entity{{Type: EntityDevice, Value: "dev1"}},
		Slots:      map[string]*Slot{"device": {Name: "device", Filled: true}},
		RawText:    "What is the status of device dev1?",
		Timestamp:  time.Now(),
	}
	assert.Equal(t, IntentQuery, intent.Type)
	assert.Equal(t, 0.95, intent.Confidence)
	assert.Equal(t, 1, len(intent.Entities))
}

func TestIntentPattern_Struct(t *testing.T) {
	pattern := IntentPattern{
		IntentType:  IntentQuery,
		IntentName:  "query_status",
		Keywords:    []string{"status", "query"},
		EntityTypes: []EntityType{EntityDevice},
		Priority:    10,
	}
	assert.Equal(t, IntentQuery, pattern.IntentType)
}

func TestIntentRecognizer(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	require.NotNil(t, recognizer)

	intent, err := recognizer.Recognize(context.Background(), "查询设备状态")
	require.NoError(t, err)
	require.NotNil(t, intent)
}

func TestIntentRecognizer_Unknown(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	intent, err := recognizer.Recognize(context.Background(), "随机文本")
	require.NoError(t, err)
	assert.Equal(t, IntentUnknown, intent.Type)
}

func TestDialogueState_Constants(t *testing.T) {
	assert.Equal(t, DialogueState("initial"), StateInitial)
	assert.Equal(t, DialogueState("active"), StateActive)
	assert.Equal(t, DialogueState("waiting"), StateWaiting)
	assert.Equal(t, DialogueState("completed"), StateCompleted)
	assert.Equal(t, DialogueState("cancelled"), StateCancelled)
	assert.Equal(t, DialogueState("error"), StateError)
}

func TestDialogueTurn_Struct(t *testing.T) {
	turn := &DialogueTurn{
		TurnID:    "t1",
		Role:      "user",
		Content:   "What is the status?",
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"source": "web"},
	}
	assert.Equal(t, "t1", turn.TurnID)
	assert.Equal(t, "user", turn.Role)
}

func TestDialogueContext_Struct(t *testing.T) {
	ctx := &DialogueContext{
		SessionID:    "s1",
		UserID:       "u1",
		CurrentState: StateActive,
		Turns:        []*DialogueTurn{{TurnID: "t1", Role: "user"}},
		Slots:        map[string]*Slot{"device": {Name: "device", Filled: true}},
		Variables:    map[string]interface{}{"station": "st1"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	assert.Equal(t, "s1", ctx.SessionID)
	assert.Equal(t, StateActive, ctx.CurrentState)
}

func TestDialoguePolicy_Struct(t *testing.T) {
	policy := &DialoguePolicy{
		PolicyID:    "p1",
		Name:        "Default Query Policy",
		Description: "Handle query intents",
		Conditions:  []PolicyCondition{{Type: "intent", Key: "type", Operator: "eq", Value: "query"}},
		Actions:     []PolicyAction{{Type: "response", Content: "Here is the information"}},
		Priority:    10,
	}
	assert.Equal(t, "p1", policy.PolicyID)
	assert.Equal(t, 1, len(policy.Conditions))
}

func TestDialogueResponse_Struct(t *testing.T) {
	resp := &DialogueResponse{
		Content:      "The device is online",
		Suggestions:  []string{"Check history", "View details"},
		Confidence:   0.95,
		RequiresMore: false,
	}
	assert.Equal(t, "The device is online", resp.Content)
	assert.Equal(t, 2, len(resp.Suggestions))
}

func TestDialogueManager(t *testing.T) {
	recognizer := NewIntentRecognizer(nil)
	dm := NewDialogueManager(recognizer, nil)
	require.NotNil(t, dm)

	dctx, err := dm.StartSession(context.Background(), "user1")
	require.NoError(t, err)
	require.NotNil(t, dctx)
	assert.Equal(t, StateInitial, dctx.CurrentState)

	got, err := dm.GetSession(dctx.SessionID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "user1", got.UserID)

	_, err = dm.Process(context.Background(), dctx.SessionID, "你好")
	require.NoError(t, err)
}

func TestAnswer_Struct(t *testing.T) {
	answer := &Answer{
		Content:    "The device is operating normally",
		Confidence: 0.92,
		References: []*Reference{{SourceID: "s1", Title: "Device Manual"}},
		Sources:    []*KnowledgeSource{{SourceID: "s1", Name: "Manual", Type: "document"}},
		GeneratedAt: time.Now(),
	}
	assert.Equal(t, "The device is operating normally", answer.Content)
	assert.Equal(t, 0.92, answer.Confidence)
}

func TestReference_Struct(t *testing.T) {
	ref := &Reference{
		SourceID:   "s1",
		SourceType: "document",
		Title:      "Device Manual",
		Content:    "Section 3.2",
		Relevance:  0.9,
	}
	assert.Equal(t, "s1", ref.SourceID)
	assert.Equal(t, 0.9, ref.Relevance)
}

func TestKnowledgeSource_Struct(t *testing.T) {
	ks := &KnowledgeSource{
		SourceID:   "ks1",
		Name:       "Device Manual",
		Type:       "document",
		Weight:     0.8,
		Enabled:    true,
		LastUpdate: time.Now(),
	}
	assert.Equal(t, "ks1", ks.SourceID)
	assert.True(t, ks.Enabled)
}

func TestAnswerTemplate_Struct(t *testing.T) {
	tmpl := &AnswerTemplate{
		TemplateID: "at1",
		Name:       "Device Status",
		IntentType: IntentQuery,
		Template:   "The device {{.device}} is {{.status}}",
		Variables:  []string{"device", "status"},
		Priority:   10,
	}
	assert.Equal(t, "at1", tmpl.TemplateID)
	assert.Equal(t, IntentQuery, tmpl.IntentType)
}

func TestKnowledgeItem_Struct(t *testing.T) {
	item := &KnowledgeItem{
		ID:        "ki1",
		Title:     "Device Troubleshooting",
		Content:   "Check power supply first",
		Category:  "troubleshooting",
		Tags:      []string{"device", "power"},
		Relevance: 0.85,
		Source:    "manual",
	}
	assert.Equal(t, "ki1", item.ID)
	assert.Equal(t, 0.85, item.Relevance)
}

func TestAnswerGenerator(t *testing.T) {
	generator := NewAnswerGenerator(nil)
	require.NotNil(t, generator)

	answer, err := generator.Generate(context.Background(), &Intent{
		Type:       IntentQuery,
		Name:       "test",
		Confidence: 0.9,
		RawText:    "What is the status?",
	}, &DialogueContext{
		SessionID:     "sess1",
		CurrentIntent: &Intent{Type: IntentQuery, Name: "test"},
		Slots:         map[string]*Slot{},
		Variables:     map[string]interface{}{},
	})
	require.NoError(t, err)
	require.NotNil(t, answer)
}
