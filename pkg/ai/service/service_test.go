package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderType_Constants(t *testing.T) {
	assert.Equal(t, ProviderType("openai"), ProviderOpenAI)
	assert.Equal(t, ProviderType("azure"), ProviderAzure)
	assert.Equal(t, ProviderType("anthropic"), ProviderAnthropic)
	assert.Equal(t, ProviderType("local"), ProviderLocal)
}

func TestMessage_Struct(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Hello, how are you?",
		Name:    "test_user",
	}
	assert.Equal(t, "user", msg.Role)
	assert.Equal(t, "Hello, how are you?", msg.Content)
}

func TestChatRequest_Struct(t *testing.T) {
	req := &ChatRequest{
		ConversationID:   "conv1",
		Messages:         []Message{{Role: "user", Content: "Hello"}},
		Model:            "gpt-4",
		Temperature:      0.7,
		TopP:             1.0,
		MaxTokens:        1000,
		FrequencyPenalty: 0.0,
		PresencePenalty:  0.0,
		User:             "user1",
		SystemPrompt:     "You are a helpful assistant",
		Stream:           false,
	}
	assert.Equal(t, "conv1", req.ConversationID)
	assert.Equal(t, 1, len(req.Messages))
	assert.Equal(t, 0.7, req.Temperature)
}

func TestChatResponse_Struct(t *testing.T) {
	resp := &ChatResponse{
		ID:             "resp1",
		ConversationID: "conv1",
		Message:        Message{Role: "assistant", Content: "Hello! How can I help you?"},
		Model:          "gpt-4",
		FinishReason:   "stop",
		Usage:          TokenUsage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30},
		CreatedAt:      time.Now(),
	}
	assert.Equal(t, "resp1", resp.ID)
	assert.Equal(t, 30, resp.Usage.TotalTokens)
}

func TestChatStreamChunk_Struct(t *testing.T) {
	chunk := ChatStreamChunk{
		ID:           "chunk1",
		ConversationID: "conv1",
		Delta:        Message{Role: "assistant", Content: "Hello"},
		Model:        "gpt-4",
		FinishReason: "",
	}
	assert.Equal(t, "chunk1", chunk.ID)
	assert.Equal(t, "Hello", chunk.Delta.Content)
}

func TestEmbeddingRequest_Struct(t *testing.T) {
	req := &EmbeddingRequest{
		Input:          []string{"Hello world"},
		Model:          "text-embedding-ada-002",
		EncodingFormat: "float",
		User:           "user1",
	}
	assert.Equal(t, 1, len(req.Input.([]string)))
}

func TestEmbeddingResponse_Struct(t *testing.T) {
	resp := &EmbeddingResponse{
		ID:    "emb1",
		Model: "text-embedding-ada-002",
		Data:  []EmbeddingData{{Index: 0, Embedding: []float64{0.1, 0.2}}},
		Usage: TokenUsage{PromptTokens: 2, TotalTokens: 2},
	}
	assert.Equal(t, 1, len(resp.Data))
}

func TestEmbeddingData_Struct(t *testing.T) {
	data := EmbeddingData{
		Index:     0,
		Embedding: []float64{0.1, 0.2, 0.3},
	}
	assert.Equal(t, 3, len(data.Embedding))
}

func TestTokenUsage_Struct(t *testing.T) {
	usage := TokenUsage{
		PromptTokens:     10,
		CompletionTokens: 10,
		TotalTokens:      20,
	}
	assert.Equal(t, 10, usage.PromptTokens)
}

func TestCompletionRequest_Struct(t *testing.T) {
	req := &CompletionRequest{
		Prompt:           "Once upon a time",
		Model:            "gpt-4",
		MaxTokens:        500,
		Temperature:      0.8,
		TopP:             1.0,
		FrequencyPenalty: 0.0,
		PresencePenalty:  0.0,
	}
	assert.Equal(t, "Once upon a time", req.Prompt)
	assert.Equal(t, 500, req.MaxTokens)
}

func TestCompletionResponse_Struct(t *testing.T) {
	resp := &CompletionResponse{
		ID:    "comp1",
		Model: "gpt-4",
		Choices: []CompletionChoice{
			{Text: "there was a kingdom", Index: 0, FinishReason: "stop"},
		},
		Usage:     TokenUsage{PromptTokens: 5, CompletionTokens: 10, TotalTokens: 15},
		CreatedAt: time.Now(),
	}
	assert.Equal(t, "comp1", resp.ID)
	assert.Equal(t, 15, resp.Usage.TotalTokens)
}

func TestModelInfo_Struct(t *testing.T) {
	info := &ModelInfo{
		Name:      "gpt-4",
		Provider:  "openai",
		Type:      "chat",
		Available: true,
		Capabilities: []string{"chat", "completion", "streaming"},
	}
	assert.Equal(t, "gpt-4", info.Name)
	assert.True(t, info.Available)
}

func TestAdapterConfig_Struct(t *testing.T) {
	cfg := AdapterConfig{
		Provider:          ProviderOpenAI,
		APIKey:            "test-key",
		BaseURL:           "https://api.openai.com/v1",
		DefaultModel:      "gpt-4",
		Organization:      "org1",
		Timeout:           30 * time.Second,
		MaxRetries:        3,
		RetryInterval:     5 * time.Second,
		InsecureSkipVerify: false,
	}
	assert.Equal(t, ProviderOpenAI, cfg.Provider)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
}

func TestConversationContext_Struct(t *testing.T) {
	ctx := &ConversationContext{
		ConversationID: "conv1",
		UserID:         "user1",
		SessionID:      "sess1",
		Messages:       []ContextMessage{{ID: "m1", Role: "user", Content: "Hello"}},
		SystemPrompt:   "You are helpful",
		Model:          "gpt-4",
		ModelParams:    map[string]interface{}{"temperature": 0.7},
		Metadata:       map[string]interface{}{"source": "web"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Status:         "active",
	}
	assert.Equal(t, "conv1", ctx.ConversationID)
	assert.Equal(t, 1, len(ctx.Messages))
}

func TestContextMessage_Struct(t *testing.T) {
	msg := ContextMessage{
		ID:         "m1",
		Role:       "user",
		Content:    "Hello",
		TokenCount: 5,
		Timestamp:  time.Now(),
	}
	assert.Equal(t, "m1", msg.ID)
	assert.Equal(t, 5, msg.TokenCount)
}

func TestTokenStats_Struct(t *testing.T) {
	stats := TokenStats{
		TotalInputTokens:  100,
		TotalOutputTokens: 200,
		MessageCount:      5,
	}
	assert.Equal(t, 100, stats.TotalInputTokens)
}

func TestPromptTemplate_Struct(t *testing.T) {
	tmpl := &PromptTemplate{
		ID:          "pt1",
		Name:        "Device Analysis",
		Description: "Analyze device data",
		Content:     "Analyze the following device data: {{.data}}",
		Type:        "analysis",
		Variables:   []TemplateVariable{{Name: "data", Type: "string", Required: true}},
		Tags:        []string{"device", "analysis"},
		Version:     "1.0",
		Enabled:     true,
	}
	assert.Equal(t, "pt1", tmpl.ID)
	assert.Equal(t, 1, len(tmpl.Variables))
	assert.True(t, tmpl.Enabled)
}

func TestTemplateVariable_Struct(t *testing.T) {
	v := TemplateVariable{
		Name:        "device_id",
		Description: "The device identifier",
		Type:        "string",
		Required:    true,
		Default:     "default_device",
		Validation:  "^[a-zA-Z0-9]+$",
	}
	assert.Equal(t, "device_id", v.Name)
	assert.True(t, v.Required)
}

func TestTemplateExample_Struct(t *testing.T) {
	ex := TemplateExample{
		Name:   "Example 1",
		Input:  map[string]interface{}{"device": "device001"},
		Output: "Device001 is currently online.",
	}
	assert.Equal(t, "Example 1", ex.Name)
}

func TestPromptManager(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	require.NotNil(t, mgr)

	tmpl := &PromptTemplate{
		Name:    "Test Template",
		Content: "Hello {{.name}}, your device {{.device}} is {{.status}}",
		Type:    "test",
		Variables: []TemplateVariable{
			{Name: "name", Type: "string", Required: true},
			{Name: "device", Type: "string", Required: true},
			{Name: "status", Type: "string", Required: false, Default: "online"},
		},
		Enabled: true,
	}

	err := mgr.Create(context.Background(), tmpl)
	require.NoError(t, err)

	tmplID := tmpl.ID
	require.NotEmpty(t, tmplID)

	got, err := mgr.Get(context.Background(), tmplID)
	require.NoError(t, err)
	assert.Equal(t, "Test Template", got.Name)

	_, err = mgr.Get(context.Background(), "nonexistent")
	assert.Error(t, err)

	list, err := mgr.List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))

	rendered, err := mgr.Render(context.Background(), tmplID, map[string]interface{}{
		"name":   "User",
		"device": "dev001",
	}, nil)
	require.NoError(t, err)
	assert.Contains(t, rendered, "User")
	assert.Contains(t, rendered, "dev001")
}

func TestPromptManager_Render_DefaultValue(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	tmpl := &PromptTemplate{
		Name:    "Default Test",
		Content: "Device {{.device}} is {{.status}}",
		Variables: []TemplateVariable{
			{Name: "device", Type: "string", Required: true},
			{Name: "status", Type: "string", Required: false, Default: "unknown"},
		},
		Enabled: true,
	}
	mgr.Create(context.Background(), tmpl)

	rendered, err := mgr.Render(context.Background(), tmpl.ID, map[string]interface{}{
		"device": "dev001",
	}, nil)
	require.NoError(t, err)
	assert.Contains(t, rendered, "dev001")
}

func TestPromptManager_Unregister(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	tmpl := &PromptTemplate{Name: "Test", Content: "Hello", Enabled: true}
	mgr.Create(context.Background(), tmpl)

	err := mgr.Delete(context.Background(), tmpl.ID)
	require.NoError(t, err)

	_, err = mgr.Get(context.Background(), tmpl.ID)
	assert.Error(t, err)
}

func TestContextManager(t *testing.T) {
	memStorage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, memStorage)
	require.NotNil(t, mgr)

	ctx := context.Background()
	conv, err := mgr.Create(ctx, "user1", "sess1", "You are helpful")
	require.NoError(t, err)
	require.NotNil(t, conv)

	got, err := mgr.Get(ctx, conv.ConversationID)
	require.NoError(t, err)
	assert.Equal(t, conv.ConversationID, got.ConversationID)

	_, err = mgr.Get(ctx, "nonexistent")
	assert.Error(t, err)

	err = mgr.AddMessage(ctx, conv.ConversationID, &ContextMessage{
		ID:      "m1",
		Role:    "user",
		Content: "Hello",
	})
	require.NoError(t, err)

	got, _ = mgr.Get(ctx, conv.ConversationID)
	assert.Equal(t, 1, len(got.Messages))

	err = mgr.Delete(ctx, conv.ConversationID)
	require.NoError(t, err)

	_, err = mgr.Get(ctx, conv.ConversationID)
	assert.Error(t, err)
}

func TestContextManager_AddMessage_NotFound(t *testing.T) {
	memStorage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, memStorage)
	ctx := context.Background()
	err := mgr.AddMessage(ctx, "nonexistent", &ContextMessage{ID: "m1", Role: "user", Content: "Hello"})
	assert.Error(t, err)
}

func TestNewModelSelector(t *testing.T) {
	selector := NewModelSelector("gpt-4")
	require.NotNil(t, selector)
}

func TestNewLoadBalancer(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	require.NotNil(t, lb)
}
