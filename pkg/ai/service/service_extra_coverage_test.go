package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseAdapter(t *testing.T) {
	config := &AdapterConfig{
		BaseURL:    "https://api.test.com/v1",
		APIKey:     "test-key",
		Timeout:    30 * time.Second,
		MaxRetries: 5,
	}

	adapter := NewBaseAdapter(config)
	require.NotNil(t, adapter)
	assert.Equal(t, config, adapter.config)
	assert.NotNil(t, adapter.httpClient)
	assert.Nil(t, adapter.modelInfo)
}

func TestNewBaseAdapterDefaults(t *testing.T) {
	config := &AdapterConfig{}
	adapter := NewBaseAdapter(config)
	require.NotNil(t, adapter)
	assert.Equal(t, 60*time.Second, config.Timeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.RetryInterval)
}

func TestBaseAdapter_GenerateRequestID(t *testing.T) {
	adapter := NewBaseAdapter(&AdapterConfig{})
	id1 := adapter.generateRequestID()
	id2 := adapter.generateRequestID()
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

func TestBaseAdapter_GetModelInfo(t *testing.T) {
	info := &ModelInfo{ID: "test-model", Name: "Test Model"}
	adapter := NewBaseAdapter(&AdapterConfig{ModelInfo: info})
	got := adapter.GetModelInfo()
	assert.Same(t, info, got)
}

func TestBaseAdapter_Close(t *testing.T) {
	adapter := NewBaseAdapter(&AdapterConfig{})
	err := adapter.Close()
	assert.NoError(t, err)
}

func TestNewOpenAIAdapter(t *testing.T) {
	adapter := NewOpenAIAdapter(&AdapterConfig{})
	require.NotNil(t, adapter)
	assert.Equal(t, "https://api.openai.com/v1", adapter.config.BaseURL)
	assert.Equal(t, "gpt-4", adapter.config.DefaultModel)
	require.NotNil(t, adapter.modelInfo)
	assert.Equal(t, "gpt-4", adapter.modelInfo.ID)
	assert.Equal(t, string(ProviderOpenAI), adapter.modelInfo.Provider)
}

func TestNewOpenAIAdapter_CustomConfig(t *testing.T) {
	info := &ModelInfo{ID: "custom", Name: "Custom"}
	adapter := NewOpenAIAdapter(&AdapterConfig{
		BaseURL:       "https://custom.com/v1",
		DefaultModel:  "gpt-3.5-turbo",
		ModelInfo:     info,
	})
	assert.Equal(t, "https://custom.com/v1", adapter.config.BaseURL)
	assert.Equal(t, "gpt-3.5-turbo", adapter.config.DefaultModel)
	assert.Same(t, info, adapter.modelInfo)
}

func TestOpenAIAdapter_BuildOpenAIChatRequest(t *testing.T) {
	adapter := NewOpenAIAdapter(&AdapterConfig{})

	req := &ChatRequest{
		Model:         "gpt-4",
		SystemPrompt:   "You are helpful.",
		Messages:       []Message{{Role: "user", Content: "Hello"}},
		Temperature:   0.7,
		TopP:          0.9,
		MaxTokens:     100,
		Stop:          []string{"\n"},
		User:          "test-user",
		Stream:        true,
		PresencePenalty: 0.1,
		FrequencyPenalty: 0.2,
	}

	openaiReq := adapter.buildOpenAIChatRequest(req)
	assert.Equal(t, "gpt-4", openaiReq.Model)
	assert.Len(t, openaiReq.Messages, 2)
	assert.Equal(t, float64(0.7), openaiReq.Temperature)
	assert.Equal(t, float64(0.9), openaiReq.TopP)
	assert.True(t, openaiReq.Stream)
	assert.Equal(t, 100, openaiReq.MaxTokens)
	assert.Equal(t, "test-user", openaiReq.User)
}

func TestOpenAIAdapter_BuildOpenAIChatRequest_NoSystemPrompt(t *testing.T) {
	adapter := NewOpenAIAdapter(&AdapterConfig{})
	req := &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	}
	openaiReq := adapter.buildOpenAIChatRequest(req)
	assert.Len(t, openaiReq.Messages, 1)
}

func TestOpenAIAdapter_BuildOpenAICompletionRequest(t *testing.T) {
	adapter := NewOpenAIAdapter(&AdapterConfig{})
	req := &CompletionRequest{
		Model:            "gpt-4",
		Prompt:           "Complete this",
		MaxTokens:        50,
		Temperature:      0.8,
		TopP:             0.95,
		N:                2,
		Stop:             []string{"."},
		Echo:             true,
		Suffix:           " suffix",
		FrequencyPenalty: 0.3,
		PresencePenalty:  0.4,
		BestOf:           3,
		User:             "user1",
	}
	result := adapter.buildOpenAICompletionRequest(req)
	data, _ := json.Marshal(result)
	var m map[string]interface{}
	json.Unmarshal(data, &m)
	assert.Equal(t, "gpt-4", m["model"])
	assert.Equal(t, float64(50), m["max_tokens"])
	assert.Equal(t, false, m["stream"])
}

func TestOpenAIAdapter_ConvertChatResponse(t *testing.T) {
	adapter := NewOpenAIAdapter(&AdapterConfig{})

	resp := &openAIChatResponse{
		ID:      "chatcmpl-123",
		Created: 1700000000,
		Model:   "gpt-4",
		Choices: []struct {
			Index        int                    `json:"index"`
			Message      map[string]interface{} `json:"message"`
			FinishReason string                 `json:"finish_reason"`
		}{
			{
				Index:        0,
				Message:      map[string]interface{}{"role": "assistant", "content": "Hello!"},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
	}

	result := adapter.convertChatResponse(resp, "conv-1")
	assert.Equal(t, "chatcmpl-123", result.ID)
	assert.Equal(t, "conv-1", result.ConversationID)
	assert.Equal(t, "assistant", result.Message.Role)
	assert.Equal(t, "Hello!", result.Message.Content)
	assert.Equal(t, "gpt-4", result.Model)
	assert.Equal(t, 10, result.Usage.PromptTokens)
	assert.Equal(t, "stop", result.FinishReason)
}

func TestOpenAIAdapter_ConvertChatResponse_EmptyChoices(t *testing.T) {
	adapter := NewOpenAIAdapter(&AdapterConfig{})
	resp := &openAIChatResponse{
		ID:      "chatcmpl-empty",
		Created: 1700000000,
		Model:   "gpt-4",
		Choices: []struct {
			Index        int                    `json:"index"`
			Message      map[string]interface{} `json:"message"`
			FinishReason string                 `json:"finish_reason"`
		}{},
	}
	result := adapter.convertChatResponse(resp, "conv-2")
	assert.Equal(t, "chatcmpl-empty", result.ID)
	assert.Empty(t, result.Message.Content)
}

func TestOpenAIAdapter_ConvertEmbeddingResponse(t *testing.T) {
	adapter := NewOpenAIAdapter(&AdapterConfig{})
	resp := &openAIEmbeddingResponse{
		Object: "list",
		Data: []struct {
			Object    string    `json:"object"`
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		}{
			{Object: "embedding", Index: 0, Embedding: []float64{0.1, 0.2, 0.3}},
			{Object: "embedding", Index: 1, Embedding: []float64{0.4, 0.5, 0.6}},
		},
		Model: "text-embedding-ada-002",
		Usage: struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		}{PromptTokens: 8, TotalTokens: 8},
	}
	result := adapter.convertEmbeddingResponse(resp)
	assert.NotEmpty(t, result.ID)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, []float64{0.1, 0.2, 0.3}, result.Data[0].Embedding)
	assert.Equal(t, "text-embedding-ada-002", result.Model)
	assert.Equal(t, 8, result.Usage.PromptTokens)
}

func TestOpenAIAdapter_ConvertCompletionResponse(t *testing.T) {
	adapter := NewOpenAIAdapter(&AdapterConfig{})
	resp := &openAICompletionResponse{
		ID:      "comp-123",
		Created: 1700000000,
		Model:   "gpt-4",
		Choices: []struct {
			Text         string    `json:"text"`
			Index        int       `json:"index"`
			Logprobs     *struct{} `json:"logprobs"`
			FinishReason string    `json:"finish_reason"`
		}{
			{Text: "Completed text", Index: 0, FinishReason: "stop"},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{PromptTokens: 5, CompletionTokens: 10, TotalTokens: 15},
	}
	result := adapter.convertCompletionResponse(resp)
	assert.Equal(t, "comp-123", result.ID)
	assert.Len(t, result.Choices, 1)
	assert.Equal(t, "Completed text", result.Choices[0].Text)
	assert.Equal(t, "stop", result.Choices[0].FinishReason)
	assert.Equal(t, 15, result.Usage.TotalTokens)
}

func TestOpenAIAdapter_ParseErrorResponse(t *testing.T) {
	adapter := NewOpenAIAdapter(&AdapterConfig{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"type":"invalid_request_error","message":"Bad request","param":"model","code":"invalid_request"}}`))
	}))
	defer server.Close()

	adapter.config.BaseURL = server.URL
	resp, _ := http.Get(server.URL + "/test")
	err := adapter.parseErrorResponse(resp)
	require.Error(t, err)

	aiErr, ok := err.(*AIError)
	require.True(t, ok)
	assert.Equal(t, "invalid_request_error", aiErr.Type)
	assert.Equal(t, "Bad request", aiErr.Message)
	assert.Equal(t, "invalid_request", aiErr.Code)
	assert.Equal(t, "model", aiErr.Param)
	assert.Equal(t, http.StatusBadRequest, aiErr.StatusCode)
}

func TestNewClaudeAdapter(t *testing.T) {
	adapter := NewClaudeAdapter(&AdapterConfig{})
	require.NotNil(t, adapter)
	assert.Equal(t, "https://api.anthropic.com/v1", adapter.config.BaseURL)
	assert.Equal(t, "claude-3-opus-20240229", adapter.config.DefaultModel)
	require.NotNil(t, adapter.modelInfo)
	assert.Equal(t, string(ProviderAnthropic), adapter.modelInfo.Provider)
}

func TestClaudeAdapter_Embedding(t *testing.T) {
	adapter := NewClaudeAdapter(&AdapterConfig{})
	result, err := adapter.Embedding(context.Background(), &EmbeddingRequest{Input: "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support embedding")
	assert.Nil(t, result)
}

func TestClaudeAdapter_BuildClaudeRequest(t *testing.T) {
	adapter := NewClaudeAdapter(&AdapterConfig{})
	req := &ChatRequest{
		Model:       "claude-3-opus-20240229",
		Messages:    []Message{{Role: "user", Content: "Hi"}, {Role: "assistant", Content: "Hello!"}},
		MaxTokens:   100,
		SystemPrompt: "Be helpful",
		Temperature: 0.5,
		TopP:        0.9,
	}
	result := adapter.buildClaudeRequest(req)
	assert.Equal(t, "claude-3-opus-20240229", result["model"])
	assert.Len(t, result["messages"], 2)
	assert.Equal(t, 100, result["max_tokens"])
	assert.Equal(t, "Be helpful", result["system"])
	assert.Equal(t, float64(0.5), result["temperature"])
}

func TestClaudeAdapter_ConvertClaudeResponse(t *testing.T) {
	adapter := NewClaudeAdapter(&AdapterConfig{})
	resp := &claudeResponse{
		ID:      "msg-123",
		Type:    "message",
		Role:    "assistant",
		Content: []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{{Type: "text", Text: "Response from Claude"}},
		Model:      "claude-3-opus-20240229",
		StopReason: "end_turn",
		Usage: struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		}{InputTokens: 20, OutputTokens: 10},
	}
	result := adapter.convertClaudeResponse(resp, "conv-claude")
	assert.Equal(t, "msg-123", result.ID)
	assert.Equal(t, "conv-claude", result.ConversationID)
	assert.Equal(t, "Response from Claude", result.Message.Content)
	assert.Equal(t, "end_turn", result.FinishReason)
	assert.Equal(t, 30, result.Usage.TotalTokens)
}

func TestClaudeAdapter_ConvertClaudeResponse_EmptyContent(t *testing.T) {
	adapter := NewClaudeAdapter(&AdapterConfig{})
	resp := &claudeResponse{
		ID:      "msg-empty",
		Content: []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{},
		Role: "assistant",
		Usage: struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		}{},
	}
	result := adapter.convertClaudeResponse(resp, "conv-x")
	assert.Empty(t, result.Message.Content)
}

func TestClaudeAdapter_Completion_StringPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(claudeResponse{
			ID:   "msg-comp",
			Role: "assistant",
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{{Type: "text", Text: "Completed"}},
			StopReason: "end_turn",
			Usage:      struct{ InputTokens int "json:\"input_tokens\""; OutputTokens int "json:\"output_tokens\"" }{InputTokens: 5, OutputTokens: 3},
	})
	}))
	defer server.Close()

	adapter := NewClaudeAdapter(&AdapterConfig{APIKey: "test-key", BaseURL: server.URL})
	req := &CompletionRequest{Prompt: "Complete this text"}
	result, err := adapter.Completion(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "Completed", result.Choices[0].Text)
}

func TestClaudeAdapter_Completion_SlicePrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(claudeResponse{
			ID:   "msg-slice",
			Role: "assistant",
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{{Type: "text", Text: "Done both"}},
			StopReason: "end_turn",
			Usage:      struct{ InputTokens int "json:\"input_tokens\""; OutputTokens int "json:\"output_tokens\"" }{InputTokens: 6, OutputTokens: 2},
		})
	}))
	defer server.Close()

	adapter := NewClaudeAdapter(&AdapterConfig{APIKey: "test-key", BaseURL: server.URL})
	req := &CompletionRequest{Prompt: []string{"First", "Second"}}
	result, err := adapter.Completion(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "Done both", result.Choices[0].Text)
}

func TestNewLocalModelAdapter(t *testing.T) {
	adapter := NewLocalModelAdapter(&AdapterConfig{})
	require.NotNil(t, adapter)
	assert.Equal(t, "local-llm", adapter.config.DefaultModel)
	require.NotNil(t, adapter.modelInfo)
	assert.Equal(t, string(ProviderLocal), adapter.modelInfo.Provider)
	assert.Contains(t, adapter.modelInfo.Capabilities, "chat")
}

func TestModelSelector_RegisterAndSelect(t *testing.T) {
	selector := NewModelSelector("default")

	mockAdapter := &mockModelAdapterForSelector{name: "test"}
	selector.Register("test", mockAdapter)

	selected, err := selector.Select("test")
	require.NoError(t, err)
	assert.Same(t, mockAdapter, selected)
}

func TestModelSelector_SelectDefault(t *testing.T) {
	selector := NewModelSelector("my-default")
	mockAdapter := &mockModelAdapterForSelector{name: "my-default"}
	selector.Register("my-default", mockAdapter)

	selected, err := selector.Select("")
	require.NoError(t, err)
	assert.Same(t, mockAdapter, selected)
}

func TestModelSelector_SelectNotFound(t *testing.T) {
	selector := NewModelSelector("default")
	_, err := selector.Select("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestModelSelector_List(t *testing.T) {
	selector := NewModelSelector("")
	selector.Register("a", &mockModelAdapterForSelector{name: "a"})
	selector.Register("b", &mockModelAdapterForSelector{name: "b"})

	models := selector.List()
	assert.Len(t, models, 2)
	assert.ElementsMatch(t, []string{"a", "b"}, models)
}

type mockModelAdapterForSelector struct {
	name string
}

func (m *mockModelAdapterForSelector) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	return nil, nil
}
func (m *mockModelAdapterForSelector) Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	return nil, nil
}
func (m *mockModelAdapterForSelector) Completion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	return nil, nil
}
func (m *mockModelAdapterForSelector) StreamChat(ctx context.Context, req *ChatRequest) (<-chan ChatStreamChunk, error) {
	return nil, nil
}
func (m *mockModelAdapterForSelector) HealthCheck(ctx context.Context) error { return nil }
func (m *mockModelAdapterForSelector) GetModelInfo() *ModelInfo {
	return &ModelInfo{ID: m.name, Name: m.name}
}
func (m *mockModelAdapterForSelector) Close() error { return nil }

func TestLoadBalancer_RegisterAndSelect(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	selector := NewModelSelector("default")
	mockAdapter := &mockModelAdapterForSelector{name: "lb-test"}
	selector.Register("default", mockAdapter)

	lb.RegisterSelector("openai", selector)

	selected, err := lb.Select("openai", "")
	require.NoError(t, err)
	assert.Same(t, mockAdapter, selected)
}

func TestLoadBalancer_SelectProviderNotFound(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	_, err := lb.Select("nonexistent", "any")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider not found")
}

func TestLoadBalancer_SelectByStrategy_RoundRobin(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	s1 := NewModelSelector("m1")
	s2 := NewModelSelector("m2")
	s1.Register("m1", &mockModelAdapterForSelector{name: "m1"})
	s2.Register("m2", &mockModelAdapterForSelector{name: "m2"})
	lb.RegisterSelector("p1", s1)
	lb.RegisterSelector("p2", s2)

	adapter, err := lb.SelectByStrategy()
	require.NoError(t, err)
	require.NotNil(t, adapter)
}

func TestLoadBalancer_SelectByStrategy_NoProviders(t *testing.T) {
	lb := NewLoadBalancer(StrategyRandom)
	_, err := lb.SelectByStrategy()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no providers")
}

func TestLoadBalancer_GetAllModels(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	s1 := NewModelSelector("")
	s1.Register("a", &mockModelAdapterForSelector{name: "a"})
	lb.RegisterSelector("p1", s1)

	models := lb.GetAllModels()
	assert.Len(t, models, 1)
	assert.Equal(t, "a", models[0].ID)
}

func TestLoadBalanceStrategy_Constants(t *testing.T) {
	assert.Equal(t, LoadBalanceStrategy("round_robin"), StrategyRoundRobin)
	assert.Equal(t, LoadBalanceStrategy("random"), StrategyRandom)
	assert.Equal(t, LoadBalanceStrategy("least_connection"), StrategyLeastConn)
	assert.Equal(t, LoadBalanceStrategy("weighted"), StrategyWeighted)
}

func TestContextManager_CreateAndGet(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)

	ctx, err := mgr.Create(context.Background(), "user1", "session1", "You are helpful.")
	require.NoError(t, err)
	require.NotNil(t, ctx)
	assert.Equal(t, "user1", ctx.UserID)
	assert.Equal(t, "session1", ctx.SessionID)
	assert.Equal(t, "You are helpful.", ctx.SystemPrompt)
	assert.Equal(t, "active", ctx.Status)
	assert.NotEmpty(t, ctx.ConversationID)

	got, err := mgr.Get(context.Background(), ctx.ConversationID)
	require.NoError(t, err)
	assert.Equal(t, ctx.ConversationID, got.ConversationID)
}

func TestContextManager_AddMessage(t *testing.T) {
	storage := NewMemoryContextStorage()
	cfg := &ContextConfig{Persist: true, DefaultTTL: 24 * time.Hour, EnableCompression: false, MaxMessages: 100}
	mgr := NewContextManager(cfg, storage)

	ctx, _ := mgr.Create(context.Background(), "u1", "s1", "")

	err := mgr.AddMessage(context.Background(), ctx.ConversationID, &ContextMessage{
		Role:    "user",
		Content: "Hello",
	})
	require.NoError(t, err)

	got, _ := mgr.Get(context.Background(), ctx.ConversationID)
	require.Len(t, got.Messages, 1)
	assert.Equal(t, "user", got.Messages[0].Role)
	assert.Equal(t, "Hello", got.Messages[0].Content)
	assert.True(t, !got.UpdatedAt.IsZero())
}

func TestContextManager_AddMessage_AutoGenerateID(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(&ContextConfig{Persist: true, DefaultTTL: 24 * time.Hour, EnableCompression: false, MaxMessages: 100}, storage)

	ctx, _ := mgr.Create(context.Background(), "u1", "s1", "")

	err := mgr.AddMessage(context.Background(), ctx.ConversationID, &ContextMessage{
		Role:    "user",
		Content: "Test",
	})
	require.NoError(t, err)

	got, _ := mgr.Get(context.Background(), ctx.ConversationID)
	assert.NotEmpty(t, got.Messages[0].ID)
}

func TestContextManager_GetMessages(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(&ContextConfig{Persist: true, DefaultTTL: 24 * time.Hour, EnableCompression: false, MaxMessages: 100}, storage)

	ctx, _ := mgr.Create(context.Background(), "u1", "s1", "")
	mgr.AddMessage(context.Background(), ctx.ConversationID, &ContextMessage{Role: "user", Content: "M1"})
	mgr.AddMessage(context.Background(), ctx.ConversationID, &ContextMessage{Role: "assistant", Content: "M2"})

	msgs, err := mgr.GetMessages(context.Background(), ctx.ConversationID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, msgs, 2)

	msgsOffset, err := mgr.GetMessages(context.Background(), ctx.ConversationID, 1, 1)
	require.NoError(t, err)
	assert.Len(t, msgsOffset, 1)
	assert.Equal(t, "M2", msgsOffset[0].Content)
}

func TestContextManager_UpdateSystemPrompt(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(&ContextConfig{Persist: true, DefaultTTL: 24 * time.Hour, EnableCompression: false, MaxMessages: 100}, storage)

	ctx, _ := mgr.Create(context.Background(), "u1", "s1", "Old prompt")
	err := mgr.UpdateSystemPrompt(context.Background(), ctx.ConversationID, "New prompt")
	require.NoError(t, err)

	got, _ := mgr.Get(context.Background(), ctx.ConversationID)
	assert.Equal(t, "New prompt", got.SystemPrompt)
}

func TestContextManager_SetModel(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(&ContextConfig{Persist: true, DefaultTTL: 24 * time.Hour, EnableCompression: false, MaxMessages: 100}, storage)

	ctx, _ := mgr.Create(context.Background(), "u1", "s1", "")
	params := map[string]interface{}{"temperature": 0.8, "max_tokens": 200}
	err := mgr.SetModel(context.Background(), ctx.ConversationID, "gpt-4", params)
	require.NoError(t, err)

	got, _ := mgr.Get(context.Background(), ctx.ConversationID)
	assert.Equal(t, "gpt-4", got.Model)
	assert.Equal(t, 0.8, got.ModelParams["temperature"])
}

func TestContextManager_SetMetadata(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(&ContextConfig{Persist: true, DefaultTTL: 24 * time.Hour, EnableCompression: false, MaxMessages: 100}, storage)

	ctx, _ := mgr.Create(context.Background(), "u1", "s1", "")
	meta := map[string]interface{}{"key1": "value1", "key2": 42}
	err := mgr.SetMetadata(context.Background(), ctx.ConversationID, meta)
	require.NoError(t, err)

	got, _ := mgr.Get(context.Background(), ctx.ConversationID)
	assert.Equal(t, "value1", got.Metadata["key1"])
	assert.Equal(t, 42, got.Metadata["key2"])
}

func TestContextManager_Delete(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(&ContextConfig{Persist: true, DefaultTTL: 24 * time.Hour, EnableCompression: false, MaxMessages: 100}, storage)

	ctx, _ := mgr.Create(context.Background(), "u1", "s1", "")
	err := mgr.Delete(context.Background(), ctx.ConversationID)
	require.NoError(t, err)

	_, err = mgr.Get(context.Background(), ctx.ConversationID)
	assert.Error(t, err)
}

func TestContextManager_ExtendTTL(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(&ContextConfig{Persist: false, DefaultTTL: 1 * time.Hour}, storage)

	ctx, _ := mgr.Create(context.Background(), "u1", "s1", "")
	originalExpiry := ctx.ExpiresAt

	err := mgr.ExtendTTL(context.Background(), ctx.ConversationID, 24*time.Hour)
	require.NoError(t, err)

	got, _ := mgr.Get(context.Background(), ctx.ConversationID)
	assert.True(t, got.ExpiresAt.After(originalExpiry))
}

func TestContextManager_Clear(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(&ContextConfig{Persist: true, DefaultTTL: 24 * time.Hour, EnableCompression: false, MaxMessages: 100}, storage)

	ctx, _ := mgr.Create(context.Background(), "u1", "s1", "")
	mgr.AddMessage(context.Background(), ctx.ConversationID, &ContextMessage{Role: "system", Content: "Sys msg"})
	mgr.AddMessage(context.Background(), ctx.ConversationID, &ContextMessage{Role: "user", Content: "User msg"})

	err := mgr.Clear(context.Background(), ctx.ConversationID)
	require.NoError(t, err)

	got, _ := mgr.Get(context.Background(), ctx.ConversationID)
	assert.Len(t, got.Messages, 1)
	assert.Equal(t, "system", got.Messages[0].Role)
}

func TestContextManager_Compress(t *testing.T) {
	storage := NewMemoryContextStorage()
	cfg := DefaultContextConfig()
	cfg.CompressionThreshold = 5
	cfg.EnableCompression = true
	mgr := NewContextManager(cfg, storage)

	ctx, _ := mgr.Create(context.Background(), "u1", "s1", "")
	for i := 0; i < 15; i++ {
		mgr.AddMessage(context.Background(), ctx.ConversationID, &ContextMessage{
			Role:    "user",
			Content: strings.Repeat("Message content for testing ", 5),
		})
	}

	err := mgr.Compress(context.Background(), ctx.ConversationID)
	require.NoError(t, err)

	got, _ := mgr.Get(context.Background(), ctx.ConversationID)
	assert.LessOrEqual(t, len(got.Messages), 15)
}

func TestContextManager_CleanExpired(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(&ContextConfig{DefaultTTL: -1 * time.Second, Persist: false}, storage)

	ctx, _ := mgr.Create(context.Background(), "u1", "s1", "")
	_ = ctx

	count, err := mgr.CleanExpired(context.Background())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(0))
}

func TestContextManager_GetStats(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(&ContextConfig{Persist: true, DefaultTTL: 24 * time.Hour, EnableCompression: false, MaxMessages: 100}, storage)

	ctx, _ := mgr.Create(context.Background(), "u1", "s1", "")
	mgr.AddMessage(context.Background(), ctx.ConversationID, &ContextMessage{Role: "user", Content: "Hi", TokenCount: 10})

	stats, err := mgr.GetStats(context.Background(), ctx.ConversationID)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.MessageCount)
	assert.Equal(t, 10, stats.TotalInputTokens)
}

func TestContextManager_BuildChatRequest(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(&ContextConfig{Persist: true, DefaultTTL: 24 * time.Hour, EnableCompression: false, MaxMessages: 100}, storage)

	ctx, _ := mgr.Create(context.Background(), "u1", "s1", "System prompt here")
	ctx.Model = "gpt-4"
	ctx.ModelParams = map[string]interface{}{"temperature": 0.7}
	storage.Save(context.Background(), ctx)

	req, err := mgr.BuildChatRequest(context.Background(), ctx.ConversationID, "User question")
	require.NoError(t, err)
	assert.Equal(t, ctx.ConversationID, req.ConversationID)
	assert.Equal(t, "gpt-4", req.Model)
	assert.Equal(t, 0.7, req.Temperature)
	assert.Len(t, req.Messages, 2)
	assert.Equal(t, "system", req.Messages[0].Role)
	assert.Equal(t, "User question", req.Messages[1].Content)
}

func TestContextManager_ContextExpired(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(&ContextConfig{DefaultTTL: 1 * time.Nanosecond, Persist: false}, storage)

	ctx, _ := mgr.Create(context.Background(), "u1", "s1", "")
	time.Sleep(2 * time.Nanosecond)

	_, err := mgr.Get(context.Background(), ctx.ConversationID)
	assert.ErrorIs(t, err, ErrContextExpired)
}

func TestMemoryContextStorage_AllOperations(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	c := &ConversationContext{
		ConversationID: "conv-mem-1",
		UserID:         "user1",
		Status:         "active",
		ExpiresAt:      time.Now().Add(1 * time.Hour),
	}

	err := storage.Save(ctx, c)
	require.NoError(t, err)

	got, err := storage.Load(ctx, "conv-mem-1")
	require.NoError(t, err)
	assert.Equal(t, "conv-mem-1", got.ConversationID)

	exists, err := storage.Exists(ctx, "conv-mem-1")
	require.NoError(t, err)
	assert.True(t, exists)

	err = storage.SetTTL(ctx, "conv-mem-1", 2*time.Hour)
	require.NoError(t, err)

	list, err := storage.ListByUser(ctx, "user1", 10, 0)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	err = storage.Delete(ctx, "conv-mem-1")
	require.NoError(t, err)

	_, err = storage.Load(ctx, "conv-mem-1")
	assert.ErrorIs(t, err, ErrContextNotFound)
}

func TestMemoryContextStorage_CleanExpired(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	storage.Save(ctx, &ConversationContext{
		ConversationID: "exp-1",
		UserID:         "u1",
		ExpiresAt:      time.Now().Add(-1 * time.Hour),
	})
	storage.Save(ctx, &ConversationContext{
		ConversationID: "exp-2",
		UserID:         "u1",
		ExpiresAt:      time.Now().Add(1 * time.Hour),
	})

	count, err := storage.CleanExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestSummaryCompressor_Compress(t *testing.T) {
	compressor := NewSummaryCompressor()
	compressor.keepRecent = 3

	cc := &ConversationContext{
		Messages: make([]ContextMessage, 0),
	}
	for i := 0; i < 10; i++ {
		cc.Messages = append(cc.Messages, ContextMessage{
			Role:    "user",
			Content: strings.Repeat("Test message content ", 10),
		})
	}

	result, err := compressor.Compress(context.Background(), cc)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(result.Messages), 5)
}

func TestSummaryCompressor_NoCompressNeeded(t *testing.T) {
	compressor := NewSummaryCompressor()
	compressor.keepRecent = 10

	cc := &ConversationContext{
		Messages: []ContextMessage{
			{Role: "user", Content: "Short"},
		},
	}

	result, err := compressor.Compress(context.Background(), cc)
	require.NoError(t, err)
	assert.Same(t, cc, result)
}

func TestSummaryCompressor_CreateSummary(t *testing.T) {
	compressor := NewSummaryCompressor()
	messages := []ContextMessage{
		{Role: "user", Content: "Question about X"},
		{Role: "assistant", Content: "Answer about Y"},
	}
	summary := compressor.createSummary(messages)
	assert.Contains(t, summary, "user")
	assert.Contains(t, summary, "assistant")
}

func TestTemplateManager_CRUD(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, false)

	tmpl := &PromptTemplate{
		Name:    "test-template",
		Content: "Hello {{.name}}!",
		Type:    TemplateTypeChat,
		Variables: []TemplateVariable{
			{Name: "name", Type: "string", Required: true},
		},
	}

	err := mgr.Create(context.Background(), tmpl)
	require.NoError(t, err)
	assert.NotEmpty(t, tmpl.ID)
	assert.Equal(t, "1.0.0", tmpl.Version)

	got, err := mgr.Get(context.Background(), tmpl.ID)
	require.NoError(t, err)
	assert.Equal(t, tmpl.Name, got.Name)

	byName, err := mgr.GetByName(context.Background(), "test-template")
	require.NoError(t, err)
	assert.Equal(t, tmpl.ID, byName.ID)

	got.Description = "Updated description"
	err = mgr.Update(context.Background(), got)
	require.NoError(t, err)
	assert.NotEqual(t, "1.0.0", got.Version)

	err = mgr.Delete(context.Background(), tmpl.ID)
	require.NoError(t, err)

	_, err = mgr.Get(context.Background(), tmpl.ID)
	assert.Error(t, err)
}

func TestTemplateManager_Render(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, false)

	tmpl := &PromptTemplate{
		Name:    "render-test",
		Content: "Hello {{.name}}, you are {{.age}} years old!",
		Variables: []TemplateVariable{
			{Name: "name", Type: "string", Required: true},
			{Name: "age", Type: "number", Required: true},
		},
	}
	mgr.Create(context.Background(), tmpl)

	result, err := mgr.Render(context.Background(), tmpl.ID, map[string]interface{}{
		"name": "Alice",
		"age":   30,
	}, nil)
	require.NoError(t, err)
	assert.Contains(t, result, "Alice")
	assert.Contains(t, result, "30")
}

func TestTemplateManager_RenderWithDefaults(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, false)

	tmpl := &PromptTemplate{
		Name:    "defaults-test",
		Content: "Value is {{.optional}}",
		Variables: []TemplateVariable{
			{Name: "optional", Type: "string", Required: false, Default: "fallback"},
		},
	}
	mgr.Create(context.Background(), tmpl)

	result, err := mgr.Render(context.Background(), tmpl.ID, map[string]interface{}{}, nil)
	require.NoError(t, err)
	assert.Contains(t, result, "fallback")
}

func TestTemplateManager_RenderStrictMode(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, false)

	tmpl := &PromptTemplate{
		Name:    "strict-test",
		Content: "{{.required_field}}",
		Variables: []TemplateVariable{
			{Name: "required_field", Type: "string", Required: true},
		},
	}
	mgr.Create(context.Background(), tmpl)

	_, err := mgr.Render(context.Background(), tmpl.ID, map[string]interface{}{}, &RenderOptions{Strict: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required variable missing")
}

func TestTemplateManager_ListWithFilter(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, false)

	mgr.Create(context.Background(), &PromptTemplate{Name: "alpha-chat", Type: TemplateTypeChat, Content: "A", Tags: []string{"tag1"}})
	mgr.Create(context.Background(), &PromptTemplate{Name: "beta-system", Type: TemplateTypeSystem, Content: "B", Tags: []string{"tag2"}})
	mgr.Create(context.Background(), &PromptTemplate{Name: "gamma-chat", Type: TemplateTypeChat, Content: "C", Enabled: false})

	filtered, err := mgr.List(context.Background(), &TemplateFilter{
		Type:   TemplateTypeChat,
		Tags:   []string{"tag1"},
		Enabled: boolPtr(true),
	})
	require.NoError(t, err)
	for _, tmpl := range filtered {
		assert.Equal(t, TemplateTypeChat, tmpl.Type)
	}
}

func TestTemplateManager_Duplicate(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, false)

	original := &PromptTemplate{
		Name:    "original",
		Content: "Original content",
		Type:    TemplateTypeChat,
	}
	mgr.Create(context.Background(), original)

	dup, err := mgr.Duplicate(context.Background(), original.ID, "copy-of-original")
	require.NoError(t, err)
	assert.Equal(t, "copy-of-original", dup.Name)
	assert.Equal(t, "Original content", dup.Content)
	assert.NotEqual(t, original.ID, dup.ID)
}

func TestTemplateManager_ExportImport(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, false)

	tmpl := &PromptTemplate{Name: "export-me", Content: "Data to export", Type: TemplateTypeChat}
	mgr.Create(context.Background(), tmpl)

	data, err := mgr.Export(context.Background(), []string{tmpl.ID})
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	storage2 := NewMemoryTemplateStorage()
	mgr2 := NewTemplateManager(storage2, false)
	err = mgr2.Import(context.Background(), data, false)
	require.NoError(t, err)
}

func TestTemplateManager_Rollback(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, false)

	tmpl := &PromptTemplate{Name: "rollback-test", Content: "v1 content", Type: TemplateTypeChat}
	mgr.Create(context.Background(), tmpl)
	v1Version := tmpl.Version

	tmpl.Content = "v2 content"
	mgr.Update(context.Background(), tmpl)

	versions, _ := mgr.GetVersions(context.Background(), tmpl.ID)
	assert.Len(t, versions, 2)

	err := mgr.Rollback(context.Background(), tmpl.ID, v1Version)
	require.NoError(t, err)

	current, _ := mgr.Get(context.Background(), tmpl.ID)
	assert.Equal(t, "v1 content", current.Content)
}

func TestTemplateManager_Rollback_VersionNotFound(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, false)

	tmpl := &PromptTemplate{Name: "rb-notfound", Content: "X", Type: TemplateTypeChat}
	mgr.Create(context.Background(), tmpl)

	err := mgr.Rollback(context.Background(), tmpl.ID, "99.99.99")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version not found")
}

func TestTemplateManager_ValidateErrors(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, false)

	err := mgr.Create(context.Background(), &PromptTemplate{Content: "no name"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")

	err = mgr.Create(context.Background(), &PromptTemplate{Name: "no-content"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "content is required")

	err = mgr.Create(context.Background(), &PromptTemplate{Name: "bad-syntax", Content: "{{unclosed}"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template syntax")
}

func TestTemplateManager_ValidateVariableTypes(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, false)

	tmpl := &PromptTemplate{
		Name:    "vtype-test",
		Content: "{{.val}}",
		Variables: []TemplateVariable{
			{Name: "val", Type: "string", Required: true},
		},
	}
	mgr.Create(context.Background(), tmpl)

	_, err := mgr.Render(context.Background(), tmpl.ID, map[string]interface{}{"val": 123}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be string")
}

func TestTemplateManager_ValidateEnum(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, false)

	tmpl := &PromptTemplate{
		Name:    "enum-test",
		Content: "{{.color}}",
		Variables: []TemplateVariable{
			{Name: "color", Type: "string", Enum: []string{"red", "green", "blue"}},
		},
	}
	mgr.Create(context.Background(), tmpl)

	_, err := mgr.Render(context.Background(), tmpl.ID, map[string]interface{}{"color": "yellow"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be one of")
}

func TestIncrementVersion(t *testing.T) {
	assert.Equal(t, "1.1.0", incrementVersion("1.0.0"))
	assert.Equal(t, "2.1.0", incrementVersion("2.0.0"))
	assert.Equal(t, "1.0.1", incrementVersion("bad-format"))
}

func TestDefaultContextConfig(t *testing.T) {
	cfg := DefaultContextConfig()
	assert.Equal(t, 100, cfg.MaxMessages)
	assert.Equal(t, 8192, cfg.MaxTokens)
	assert.Equal(t, 24*time.Hour, cfg.DefaultTTL)
	assert.True(t, cfg.EnableCompression)
	assert.Equal(t, 50, cfg.CompressionThreshold)
	assert.Equal(t, 10, cfg.KeepRecentMessages)
	assert.True(t, cfg.Persist)
}

func TestContextStatusConstants(t *testing.T) {
	assert.Equal(t, "active", ContextStatusActive)
	assert.Equal(t, "archived", ContextStatusArchived)
	assert.Equal(t, "completed", ContextStatusCompleted)
}

func TestContextErrors(t *testing.T) {
	assert.Error(t, ErrContextNotFound)
	assert.Error(t, ErrContextExpired)
	assert.Error(t, ErrContextFull)
}

func TestAIError_ErrorMethod(t *testing.T) {
	err := NewAIError("test_type", "test message", "TEST_CODE", 500)
	assert.Equal(t, "test message", err.Error())
	assert.Equal(t, "test_type", err.Type)
	assert.Equal(t, "TEST_CODE", err.Code)
	assert.Equal(t, 500, err.StatusCode)
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *AIError
		typeStr  string
		code     string
		status   int
	}{
		{"InvalidRequest", ErrInvalidRequest, "invalid_request_error", "invalid_request", 400},
		{"Authentication", ErrAuthentication, "authentication_error", "auth_failed", 401},
		{"PermissionDenied", ErrPermissionDenied, "permission_denied", "forbidden", 403},
		{"NotFound", ErrNotFound, "not_found_error", "not_found", 404},
		{"RateLimit", ErrRateLimitExceeded, "rate_limit_error", "rate_limit", 429},
		{"ServiceUnavailable", ErrServiceUnavailable, "service_unavailable", "unavailable", 503},
		{"ModelOverloaded", ErrModelOverloaded, "model_overloaded", "overloaded", 529},
		{"ContextLength", ErrContextLength, "context_length_exceeded", "context_too_long", 400},
		{"ContentFilter", ErrContentFilter, "content_filter", "content_filter", 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.typeStr, tt.err.Type)
			assert.Equal(t, tt.code, tt.err.Code)
			assert.Equal(t, tt.status, tt.err.StatusCode)
		})
	}
}

func TestServiceConfig_Struct(t *testing.T) {
	cfg := &ServiceConfig{
		DefaultModel:       "gpt-4",
		APIKey:             "sk-key",
		BaseURL:            "https://api.openai.com/v1",
		Timeout:            60 * time.Second,
		DefaultTemperature: 0.7,
		DefaultMaxTokens:   2048,
		EnableCache:        true,
		CacheTTL:           1 * time.Hour,
		Extra:              map[string]interface{}{"env": "prod"},
	}
	assert.Equal(t, "gpt-4", cfg.DefaultModel)
	assert.True(t, cfg.EnableCache)
}

func TestModelInfo_FullStruct(t *testing.T) {
	info := &ModelInfo{
		ID:              "model-id",
		Name:            "GPT-4",
		Type:            string(ModelTypeChat),
		Provider:        string(ProviderOpenAI),
		ContextWindow:   8192,
		MaxOutputTokens: 4096,
		InputPrice:      0.03,
		OutputPrice:     0.06,
		Available:       true,
		Capabilities:    []string{"chat", "streaming"},
	}
	assert.Equal(t, "model-id", info.ID)
	assert.Equal(t, ProviderOpenAI, ProviderType(info.Provider))
}

func TestProviderType_AllConstants(t *testing.T) {
	assert.Equal(t, ProviderType("openai"), ProviderOpenAI)
	assert.Equal(t, ProviderType("anthropic"), ProviderAnthropic)
	assert.Equal(t, ProviderType("local"), ProviderLocal)
	assert.Equal(t, ProviderType("azure"), ProviderAzure)
}

func TestModelTypeConstants(t *testing.T) {
	assert.Equal(t, ModelType("chat"), ModelTypeChat)
	assert.Equal(t, ModelType("completion"), ModelTypeCompletion)
	assert.Equal(t, ModelType("embedding"), ModelTypeEmbedding)
}

func TestBuiltInTemplates(t *testing.T) {
	assert.Len(t, BuiltInTemplates, 3)
	names := make(map[string]bool)
	for _, bt := range BuiltInTemplates {
		names[bt.Name] = true
		assert.NotEmpty(t, bt.ID)
		assert.True(t, bt.Enabled)
	}
	assert.True(t, names["default-chat"])
	assert.True(t, names["alarm-analysis"])
	assert.True(t, names["report-generation"])
}

func TestTemplateTypeConstants(t *testing.T) {
	assert.Equal(t, "chat", TemplateTypeChat)
	assert.Equal(t, "completion", TemplateTypeCompletion)
	assert.Equal(t, "system", TemplateTypeSystem)
	assert.Equal(t, "function", TemplateTypeFunction)
}

func TestRenderOptions_Struct(t *testing.T) {
	opts := &RenderOptions{
		Strict:                   true,
		MissingVariableHandling: "error",
		Delimiters:               []string{"{{", "}}"},
		EscapeFunc:               strings.ToUpper,
		Context:                  map[string]interface{}{"extra": "data"},
	}
	assert.True(t, opts.Strict)
	assert.NotNil(t, opts.EscapeFunc)
}

func TestPromptTemplate_FullStruct(t *testing.T) {
	now := time.Now()
	pt := &PromptTemplate{
		ID:          "tmpl-full",
		Name:        "full template",
		Description: "A full template",
		Content:     "{{.var}} output",
		Type:        TemplateTypeChat,
		Variables: []TemplateVariable{
			{
				Name:        "var",
				Description: "A variable",
				Type:        "string",
				Required:    true,
				Default:     "default-val",
				Validation:  "^[a-z]+$",
				Enum:        []string{"a", "b"},
				Min:         floatPtr(0),
				Max:         floatPtr(100),
				MinLength:   intPtr(1),
				MaxLength:   intPtr(255),
			},
		},
		Examples: []TemplateExample{
			{Name: "ex1", Input: map[string]interface{}{"var": "hello"}, Output: "hello output"},
		},
		Tags:       []string{"tag-a", "tag-b"},
		Version:    "2.0.0",
		CreatedAt:  now,
		UpdatedAt:  now,
		CreatedBy:  "admin",
		Enabled:    true,
		Metadata:   map[string]interface{}{"cat": "test"},
	}
	assert.Equal(t, "tmpl-full", pt.ID)
	assert.Len(t, pt.Variables, 1)
	assert.Len(t, pt.Examples, 1)
}

func TestTemplateFilter_Struct(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	future := time.Now().Add(24 * time.Hour)
	enabled := true
	f := &TemplateFilter{
		Name:          "search",
		Type:          TemplateTypeChat,
		Tags:          []string{"important"},
		Enabled:       &enabled,
		CreatedBy:     "admin",
		CreatedAfter:  &past,
		CreatedBefore: &future,
	}
	assert.Equal(t, "search", f.Name)
	assert.True(t, *f.Enabled)
}

func TestTemplateVersion_Struct(t *testing.T) {
	tv := &TemplateVersion{
		TemplateID: "tmpl-1",
		Version:    "1.2.3",
		Content:    "old version content",
		ChangeLog:  "Bug fix",
		CreatedAt:  time.Now(),
		CreatedBy:  "dev",
	}
	assert.Equal(t, "tmpl-1", tv.TemplateID)
	assert.Equal(t, "1.2.3", tv.Version)
}

func TestTemplateExample_FullStruct(t *testing.T) {
	te := &TemplateExample{
		Name:  "example 1",
		Input: map[string]interface{}{"query": "test"},
		Output: "result output",
		Note:   "This is an example",
	}
	assert.Equal(t, "example 1", te.Name)
	assert.Equal(t, "result output", te.Output)
}

func TestMemoryTemplateStorage_AllMethods(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	ctx := context.Background()

	tmpl := &PromptTemplate{ID: "mem-store-1", Name: "Mem Store Test", Content: "Test {{.x}}", Type: TemplateTypeChat}
	err := storage.Save(ctx, tmpl)
	require.NoError(t, err)

	got, err := storage.Load(ctx, "mem-store-1")
	require.NoError(t, err)
	assert.Equal(t, "Mem Store Test", got.Name)

	all, err := storage.LoadAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	version := &TemplateVersion{TemplateID: "mem-store-1", Version: "1.0.0", Content: "v1 content"}
	err = storage.SaveVersion(ctx, version)
	require.NoError(t, err)

	versions, err := storage.LoadVersions(ctx, "mem-store-1")
	require.NoError(t, err)
	assert.Len(t, versions, 1)

	err = storage.Delete(ctx, "mem-store-1")
	require.NoError(t, err)

	_, err = storage.Load(ctx, "mem-store-1")
	assert.Error(t, err)
}

func TestAdapterConfig_FullStruct(t *testing.T) {
	cfg := &AdapterConfig{
		Provider:            ProviderOpenAI,
		APIKey:              "sk-xxx",
		BaseURL:             "https://api.example.com/v1",
		DefaultModel:        "gpt-4",
		Organization:        "org-123",
		Timeout:             120 * time.Second,
		MaxRetries:          5,
		RetryInterval:       5 * time.Second,
		ProxyURL:            "http://proxy:8080",
		InsecureSkipVerify: true,
		ModelInfo:          &ModelInfo{ID: "mi-1"},
	}
	assert.Equal(t, ProviderOpenAI, cfg.Provider)
	assert.True(t, cfg.InsecureSkipVerify)
}

func TestChatRequest_FullStruct(t *testing.T) {
	cr := &ChatRequest{
		ConversationID:   "chat-req-1",
		Messages:          []Message{{Role: "user", Content: "Hi"}},
		Model:             "gpt-4",
		Temperature:       0.8,
		TopP:              0.95,
		MaxTokens:         300,
		Stop:              []string{"\n"},
		FrequencyPenalty:  0.3,
		PresencePenalty:   0.4,
		User:              "u1",
		SystemPrompt:      "You are helpful.",
		Stream:            true,
		Extra:             map[string]interface{}{"key": "val"},
	}
	assert.Equal(t, "chat-req-1", cr.ConversationID)
	assert.True(t, cr.Stream)
}

func TestChatResponse_FullStruct(t *testing.T) {
	resp := &ChatResponse{
		ID:             "resp-1",
		ConversationID: "conv-1",
		Message:        Message{Role: "assistant", Content: "Response"},
		Model:          "gpt-4",
		Usage:          TokenUsage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		CreatedAt:      time.Now(),
		FinishReason:   "stop",
		Extra:          map[string]interface{}{"latency": 150},
	}
	assert.Equal(t, "resp-1", resp.ID)
	assert.Equal(t, 15, resp.Usage.TotalTokens)
}

func TestChatStreamChunk_FullStruct(t *testing.T) {
	chunk := &ChatStreamChunk{
		ID:             "chunk-1",
		ConversationID: "conv-1",
		Delta:          Message{Role: "assistant", Content: "Hello"},
		Model:          "gpt-4",
		FinishReason:   "",
		Error:          nil,
	}
	assert.Equal(t, "chunk-1", chunk.ID)
	assert.Nil(t, chunk.Error)
}

func TestEmbeddingRequest_FullStruct(t *testing.T) {
	er := &EmbeddingRequest{
		Input:          "embed this text",
		Model:          "text-embedding-ada-002",
		EncodingFormat: "float",
		User:           "u1",
		Extra:          map[string]interface{}{},
	}
	assert.Equal(t, "embed this text", er.Input)
}

func TestEmbeddingResponse_FullStruct(t *testing.T) {
	er := &EmbeddingResponse{
		ID:    "emb-1",
		Data:  []EmbeddingData{{Embedding: []float64{0.1, 0.2}, Index: 0, Object: "embedding"}},
		Model: "ada-002",
		Usage: TokenUsage{PromptTokens: 4, TotalTokens: 4},
	}
	assert.Len(t, er.Data, 1)
}

func TestEmbeddingData_FullStruct(t *testing.T) {
	ed := &EmbeddingData{
		Embedding: []float64{0.1, 0.2, 0.3},
		Index:     0,
		Object:    "embedding",
	}
	assert.Len(t, ed.Embedding, 3)
}

func TestCompletionRequest_FullStruct(t *testing.T) {
	cr := &CompletionRequest{
		Model:            "gpt-4",
		Prompt:           "Complete this sentence",
		MaxTokens:        100,
		Temperature:      0.7,
		TopP:             0.9,
		N:                2,
		Stop:             []string{"."},
		Echo:             true,
		Suffix:           " end",
		FrequencyPenalty: 0.2,
		PresencePenalty:  0.3,
		BestOf:           2,
		User:             "u1",
	}
	assert.Equal(t, "Complete this sentence", cr.Prompt)
	assert.True(t, cr.Echo)
}

func TestCompletionResponse_FullStruct(t *testing.T) {
	cr := &CompletionResponse{
		ID:      "comp-1",
		Choices: []CompletionChoice{{Text: "Result", Index: 0, FinishReason: "stop"}},
		Model:   "gpt-4",
		Usage:   TokenUsage{PromptTokens: 5, CompletionTokens: 3, TotalTokens: 8},
	}
	assert.Len(t, cr.Choices, 1)
}

func TestCompletionChoice_Struct(t *testing.T) {
	cc := &CompletionChoice{
		Text:         "Generated text",
		Index:        0,
		Logprobs:     nil,
		FinishReason: "length",
	}
	assert.Equal(t, "Generated text", cc.Text)
}

func TestTokenUsage_FullStruct(t *testing.T) {
	tu := TokenUsage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150}
	assert.Equal(t, 150, tu.TotalTokens)
}

func TestModelAdapter_Interface(t *testing.T) {
	var _ ModelAdapter = (*mockModelAdapterForSelector)(nil)
}

func boolPtr(b bool) *bool       { return &b }
func intPtr(i int) *int         { return &i }
func floatPtr(f float64) *float64 { return &f }
