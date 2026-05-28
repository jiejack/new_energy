package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseAdapter(t *testing.T) {
	cfg := &AdapterConfig{
		Provider:     ProviderOpenAI,
		APIKey:       "test-key",
		BaseURL:      "http://localhost:8080",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		RetryInterval: 100 * time.Millisecond,
	}
	adapter := NewBaseAdapter(cfg)
	require.NotNil(t, adapter)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.MaxRetries)
}

func TestNewBaseAdapter_DefaultValues(t *testing.T) {
	cfg := &AdapterConfig{
		Provider: ProviderOpenAI,
		BaseURL:  "http://localhost:8080",
	}
	adapter := NewBaseAdapter(cfg)
	require.NotNil(t, adapter)
	assert.Equal(t, 60*time.Second, cfg.Timeout)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 1*time.Second, cfg.RetryInterval)
}

func TestBaseAdapter_Close(t *testing.T) {
	cfg := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost:8080"}
	adapter := NewBaseAdapter(cfg)
	err := adapter.Close()
	assert.NoError(t, err)
}

func TestBaseAdapter_GetModelInfo(t *testing.T) {
	info := &ModelInfo{ID: "test-model", Name: "Test"}
	cfg := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost:8080", ModelInfo: info}
	adapter := NewBaseAdapter(cfg)
	result := adapter.GetModelInfo()
	assert.Equal(t, "test-model", result.ID)
}

func TestBaseAdapter_GetModelInfo_Nil(t *testing.T) {
	cfg := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost:8080"}
	adapter := NewBaseAdapter(cfg)
	result := adapter.GetModelInfo()
	assert.Nil(t, result)
}

func TestBaseAdapter_GenerateRequestID(t *testing.T) {
	cfg := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost:8080"}
	adapter := NewBaseAdapter(cfg)
	id1 := adapter.generateRequestID()
	id2 := adapter.generateRequestID()
	assert.NotEqual(t, id1, id2)
	assert.NotEmpty(t, id1)
}

func TestBaseAdapter_DoRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "org1", r.Header.Get("OpenAI-Organization"))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:      ProviderOpenAI,
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Organization:  "org1",
		Timeout:       5 * time.Second,
		MaxRetries:    1,
		RetryInterval: 10 * time.Millisecond,
	}
	adapter := NewBaseAdapter(cfg)

	resp, err := adapter.doRequest(context.Background(), http.MethodGet, "/test", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestBaseAdapter_DoRequest_WithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "hello", body["message"])
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderOpenAI,
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewBaseAdapter(cfg)

	resp, err := adapter.doRequest(context.Background(), http.MethodPost, "/test", map[string]interface{}{"message": "hello"})
	require.NoError(t, err)
	defer resp.Body.Close()
}

func TestBaseAdapter_DoRequest_ServerError_Retry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:      ProviderOpenAI,
		BaseURL:       server.URL,
		Timeout:       5 * time.Second,
		MaxRetries:    2,
		RetryInterval: 10 * time.Millisecond,
	}
	adapter := NewBaseAdapter(cfg)

	_, err := adapter.doRequest(context.Background(), http.MethodGet, "/test", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max retries exceeded")
}

func TestBaseAdapter_DoRequest_RateLimit_Retry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:      ProviderOpenAI,
		BaseURL:       server.URL,
		Timeout:       5 * time.Second,
		MaxRetries:    1,
		RetryInterval: 10 * time.Millisecond,
	}
	adapter := NewBaseAdapter(cfg)

	_, err := adapter.doRequest(context.Background(), http.MethodGet, "/test", nil)
	assert.Error(t, err)
}

func TestBaseAdapter_DoRequest_ContextCancelled(t *testing.T) {
	cfg := &AdapterConfig{
		Provider:      ProviderOpenAI,
		BaseURL:       "http://localhost:99999",
		Timeout:       5 * time.Second,
		MaxRetries:    3,
		RetryInterval: 10 * time.Millisecond,
	}
	adapter := NewBaseAdapter(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := adapter.doRequest(ctx, http.MethodGet, "/test", nil)
	assert.Error(t, err)
}

func TestNewOpenAIAdapter(t *testing.T) {
	cfg := &AdapterConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
	}
	adapter := NewOpenAIAdapter(cfg)
	require.NotNil(t, adapter)
	assert.Equal(t, "https://api.openai.com/v1", cfg.BaseURL)
	assert.Equal(t, "gpt-4", cfg.DefaultModel)
	require.NotNil(t, cfg.ModelInfo)
	assert.Equal(t, "gpt-4", cfg.ModelInfo.ID)
}

func TestNewOpenAIAdapter_CustomConfig(t *testing.T) {
	cfg := &AdapterConfig{
		Provider:      ProviderOpenAI,
		APIKey:        "test-key",
		BaseURL:       "http://custom-url",
		DefaultModel:  "gpt-3.5-turbo",
	}
	adapter := NewOpenAIAdapter(cfg)
	require.NotNil(t, adapter)
	assert.Equal(t, "http://custom-url", cfg.BaseURL)
	assert.Equal(t, "gpt-3.5-turbo", cfg.DefaultModel)
}

func TestOpenAIAdapter_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/chat/completions", r.URL.Path)
		resp := openAIChatResponse{
			ID:      "chat-123",
			Model:   "gpt-4",
			Created: time.Now().Unix(),
			Choices: []struct {
				Index        int                    `json:"index"`
				Message      map[string]interface{} `json:"message"`
				FinishReason string                 `json:"finish_reason"`
			}{
				{Index: 0, Message: map[string]interface{}{"role": "assistant", "content": "Hello!"}, FinishReason: "stop"},
			},
		}
		resp.Usage.PromptTokens = 10
		resp.Usage.CompletionTokens = 5
		resp.Usage.TotalTokens = 15
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderOpenAI,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewOpenAIAdapter(cfg)

	resp, err := adapter.Chat(context.Background(), &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "chat-123", resp.ID)
	assert.Equal(t, "Hello!", resp.Message.Content)
	assert.Equal(t, 15, resp.Usage.TotalTokens)
}

func TestOpenAIAdapter_Chat_DefaultModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openAIChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "gpt-4", req.Model)
		w.WriteHeader(http.StatusOK)
		resp := openAIChatResponse{ID: "chat-1", Model: "gpt-4", Created: time.Now().Unix()}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:      ProviderOpenAI,
		APIKey:        "test-key",
		BaseURL:       server.URL,
		DefaultModel:  "gpt-4",
		Timeout:       5 * time.Second,
		MaxRetries:    1,
	}
	adapter := NewOpenAIAdapter(cfg)

	_, err := adapter.Chat(context.Background(), &ChatRequest{Messages: []Message{{Role: "user", Content: "Hi"}}})
	require.NoError(t, err)
}

func TestOpenAIAdapter_Chat_SystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openAIChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, 2, len(req.Messages))
		assert.Equal(t, "system", req.Messages[0]["role"])
		w.WriteHeader(http.StatusOK)
		resp := openAIChatResponse{ID: "chat-1", Model: "gpt-4", Created: time.Now().Unix()}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderOpenAI,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewOpenAIAdapter(cfg)

	_, err := adapter.Chat(context.Background(), &ChatRequest{
		SystemPrompt: "You are helpful",
		Messages:     []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
}

func TestOpenAIAdapter_Chat_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(openAIErrorResponse{
			Error: struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Param   string `json:"param"`
				Code    string `json:"code"`
			}{Message: "bad request", Type: "invalid_request", Code: "bad"},
		})
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderOpenAI,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewOpenAIAdapter(cfg)

	_, err := adapter.Chat(context.Background(), &ChatRequest{Messages: []Message{{Role: "user", Content: "Hi"}}})
	assert.Error(t, err)
	var aiErr *AIError
	assert.True(t, errors.As(err, &aiErr))
}

func TestOpenAIAdapter_Chat_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:      ProviderOpenAI,
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       5 * time.Second,
		MaxRetries:    1,
		RetryInterval: 10 * time.Millisecond,
	}
	adapter := NewOpenAIAdapter(cfg)

	_, err := adapter.Chat(context.Background(), &ChatRequest{Messages: []Message{{Role: "user", Content: "Hi"}}})
	assert.Error(t, err)
}

func TestOpenAIAdapter_Embedding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/embeddings", r.URL.Path)
		resp := openAIEmbeddingResponse{
			Object: "list",
			Model:  "text-embedding-ada-002",
		}
		resp.Data = append(resp.Data, struct {
			Object    string    `json:"object"`
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		}{Object: "embedding", Index: 0, Embedding: []float64{0.1, 0.2, 0.3}})
		resp.Usage.PromptTokens = 5
		resp.Usage.TotalTokens = 5
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderOpenAI,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewOpenAIAdapter(cfg)

	resp, err := adapter.Embedding(context.Background(), &EmbeddingRequest{
		Input: "Hello world",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(resp.Data))
	assert.Equal(t, []float64{0.1, 0.2, 0.3}, resp.Data[0].Embedding)
}

func TestOpenAIAdapter_Embedding_DefaultModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "text-embedding-ada-002", req["model"])
		w.WriteHeader(http.StatusOK)
		resp := openAIEmbeddingResponse{Model: "text-embedding-ada-002"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderOpenAI,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewOpenAIAdapter(cfg)

	_, err := adapter.Embedding(context.Background(), &EmbeddingRequest{Input: "test"})
	require.NoError(t, err)
}

func TestOpenAIAdapter_Completion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/completions", r.URL.Path)
		resp := openAICompletionResponse{
			ID:      "comp-123",
			Model:   "gpt-4",
			Created: time.Now().Unix(),
		}
		resp.Choices = append(resp.Choices, struct {
			Text         string    `json:"text"`
			Index        int       `json:"index"`
			Logprobs     *struct{} `json:"logprobs"`
			FinishReason string    `json:"finish_reason"`
		}{Text: "Once upon a time", Index: 0, FinishReason: "stop"})
		resp.Usage.PromptTokens = 5
		resp.Usage.CompletionTokens = 10
		resp.Usage.TotalTokens = 15
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderOpenAI,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewOpenAIAdapter(cfg)

	resp, err := adapter.Completion(context.Background(), &CompletionRequest{
		Prompt: "Tell me a story",
	})
	require.NoError(t, err)
	assert.Equal(t, "comp-123", resp.ID)
	assert.Equal(t, "Once upon a time", resp.Choices[0].Text)
}

func TestOpenAIAdapter_HealthCheck_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/models", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderOpenAI,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewOpenAIAdapter(cfg)

	err := adapter.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestOpenAIAdapter_HealthCheck_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:      ProviderOpenAI,
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       5 * time.Second,
		MaxRetries:    1,
		RetryInterval: 10 * time.Millisecond,
	}
	adapter := NewOpenAIAdapter(cfg)

	err := adapter.HealthCheck(context.Background())
	assert.Error(t, err)
}

func TestOpenAIAdapter_ConvertChatResponse_EmptyChoices(t *testing.T) {
	cfg := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost"}
	adapter := NewOpenAIAdapter(cfg)

	resp := adapter.convertChatResponse(&openAIChatResponse{
		ID:      "test",
		Model:   "gpt-4",
		Created: time.Now().Unix(),
	}, "conv-1")

	assert.Equal(t, "test", resp.ID)
	assert.Equal(t, "conv-1", resp.ConversationID)
	assert.Empty(t, resp.Message.Content)
}

func TestNewClaudeAdapter(t *testing.T) {
	cfg := &AdapterConfig{
		Provider: ProviderAnthropic,
		APIKey:   "test-key",
	}
	adapter := NewClaudeAdapter(cfg)
	require.NotNil(t, adapter)
	assert.Equal(t, "https://api.anthropic.com/v1", cfg.BaseURL)
	assert.Equal(t, "claude-3-opus-20240229", cfg.DefaultModel)
	require.NotNil(t, cfg.ModelInfo)
	assert.Equal(t, string(ProviderAnthropic), cfg.ModelInfo.Provider)
}

func TestClaudeAdapter_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/messages", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

		resp := claudeResponse{
			ID:   "msg-123",
			Type: "message",
			Role: "assistant",
		}
		resp.Content = append(resp.Content, struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{Type: "text", Text: "Hello from Claude!"})
		resp.Model = "claude-3-opus-20240229"
		resp.StopReason = "end_turn"
		resp.Usage.InputTokens = 10
		resp.Usage.OutputTokens = 20
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderAnthropic,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewClaudeAdapter(cfg)

	resp, err := adapter.Chat(context.Background(), &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "msg-123", resp.ID)
	assert.Equal(t, "Hello from Claude!", resp.Message.Content)
	assert.Equal(t, 30, resp.Usage.TotalTokens)
}

func TestClaudeAdapter_Chat_DefaultModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "claude-3-opus-20240229", req["model"])
		w.WriteHeader(http.StatusOK)
		resp := claudeResponse{ID: "msg-1", Role: "assistant", Model: "claude-3-opus-20240229"}
		resp.Content = append(resp.Content, struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{Type: "text", Text: "Hi"})
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:      ProviderAnthropic,
		APIKey:        "test-key",
		BaseURL:       server.URL,
		DefaultModel:  "claude-3-opus-20240229",
		Timeout:       5 * time.Second,
		MaxRetries:    1,
	}
	adapter := NewClaudeAdapter(cfg)

	_, err := adapter.Chat(context.Background(), &ChatRequest{Messages: []Message{{Role: "user", Content: "Hi"}}})
	require.NoError(t, err)
}

func TestClaudeAdapter_Chat_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(claudeErrorResponse{
			Type: "error",
			Error: struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			}{Type: "invalid_request", Message: "bad request"},
		})
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderAnthropic,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewClaudeAdapter(cfg)

	_, err := adapter.Chat(context.Background(), &ChatRequest{Messages: []Message{{Role: "user", Content: "Hi"}}})
	assert.Error(t, err)
}

func TestClaudeAdapter_Chat_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:      ProviderAnthropic,
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       5 * time.Second,
		MaxRetries:    1,
		RetryInterval: 10 * time.Millisecond,
	}
	adapter := NewClaudeAdapter(cfg)

	_, err := adapter.Chat(context.Background(), &ChatRequest{Messages: []Message{{Role: "user", Content: "Hi"}}})
	assert.Error(t, err)
}

func TestClaudeAdapter_Embedding(t *testing.T) {
	cfg := &AdapterConfig{Provider: ProviderAnthropic, APIKey: "test-key"}
	adapter := NewClaudeAdapter(cfg)

	_, err := adapter.Embedding(context.Background(), &EmbeddingRequest{Input: "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "claude does not support embedding")
}

func TestClaudeAdapter_Completion_StringPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := claudeResponse{
			ID:   "msg-comp",
			Role: "assistant",
			Model: "claude-3-opus-20240229",
		}
		resp.Content = append(resp.Content, struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{Type: "text", Text: "Completion result"})
		resp.StopReason = "stop"
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderAnthropic,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewClaudeAdapter(cfg)

	resp, err := adapter.Completion(context.Background(), &CompletionRequest{
		Prompt: "Tell me a story",
	})
	require.NoError(t, err)
	assert.Equal(t, "Completion result", resp.Choices[0].Text)
}

func TestClaudeAdapter_Completion_StringArrayPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		messages := req["messages"].([]interface{})
		assert.Equal(t, 2, len(messages))

		resp := claudeResponse{
			ID:   "msg-comp2",
			Role: "assistant",
			Model: "claude-3-opus-20240229",
		}
		resp.Content = append(resp.Content, struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{Type: "text", Text: "Multi result"})
		resp.StopReason = "stop"
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderAnthropic,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewClaudeAdapter(cfg)

	resp, err := adapter.Completion(context.Background(), &CompletionRequest{
		Prompt: []string{"First prompt", "Second prompt"},
	})
	require.NoError(t, err)
	assert.Equal(t, "Multi result", resp.Choices[0].Text)
}

func TestClaudeAdapter_HealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := claudeResponse{ID: "msg-health", Role: "assistant", Model: "claude-3-opus-20240229"}
		resp.Content = append(resp.Content, struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{Type: "text", Text: "pong"})
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderAnthropic,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewClaudeAdapter(cfg)

	err := adapter.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestClaudeAdapter_HealthCheck_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:      ProviderAnthropic,
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       5 * time.Second,
		MaxRetries:    1,
		RetryInterval: 10 * time.Millisecond,
	}
	adapter := NewClaudeAdapter(cfg)

	err := adapter.HealthCheck(context.Background())
	assert.Error(t, err)
}

func TestClaudeAdapter_ConvertResponse_EmptyContent(t *testing.T) {
	cfg := &AdapterConfig{Provider: ProviderAnthropic, APIKey: "test-key"}
	adapter := NewClaudeAdapter(cfg)

	resp := adapter.convertClaudeResponse(&claudeResponse{
		ID:   "msg-empty",
		Role: "assistant",
	}, "conv-1")

	assert.Equal(t, "msg-empty", resp.ID)
	assert.Empty(t, resp.Message.Content)
}

func TestNewLocalModelAdapter(t *testing.T) {
	cfg := &AdapterConfig{Provider: ProviderLocal, APIKey: "test-key"}
	adapter := NewLocalModelAdapter(cfg)
	require.NotNil(t, adapter)
	assert.Equal(t, "local-llm", cfg.DefaultModel)
	require.NotNil(t, cfg.ModelInfo)
	assert.Equal(t, string(ProviderLocal), cfg.ModelInfo.Provider)
}

func TestLocalModelAdapter_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		result := struct {
			ID      string `json:"id"`
			Model   string `json:"model"`
			Choices []struct {
				Message      Message `json:"message"`
				FinishReason string  `json:"finish_reason"`
			} `json:"choices"`
			Usage TokenUsage `json:"usage"`
		}{
			ID:    "local-123",
			Model: "local-llm",
			Choices: []struct {
				Message      Message `json:"message"`
				FinishReason string  `json:"finish_reason"`
			}{{Message: Message{Role: "assistant", Content: "Local response"}, FinishReason: "stop"}},
			Usage: TokenUsage{PromptTokens: 5, CompletionTokens: 10, TotalTokens: 15},
		}
		json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderLocal,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewLocalModelAdapter(cfg)

	resp, err := adapter.Chat(context.Background(), &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "local-123", resp.ID)
	assert.Equal(t, "Local response", resp.Message.Content)
}

func TestLocalModelAdapter_Chat_NoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := struct {
			ID      string `json:"id"`
			Model   string `json:"model"`
			Choices []struct {
				Message      Message `json:"message"`
				FinishReason string  `json:"finish_reason"`
			} `json:"choices"`
		}{ID: "local-empty", Model: "local-llm", Choices: nil}
		json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderLocal,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewLocalModelAdapter(cfg)

	_, err := adapter.Chat(context.Background(), &ChatRequest{Messages: []Message{{Role: "user", Content: "Hi"}}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no response")
}

func TestLocalModelAdapter_Chat_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:      ProviderLocal,
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       5 * time.Second,
		MaxRetries:    1,
		RetryInterval: 10 * time.Millisecond,
	}
	adapter := NewLocalModelAdapter(cfg)

	_, err := adapter.Chat(context.Background(), &ChatRequest{Messages: []Message{{Role: "user", Content: "Hi"}}})
	assert.Error(t, err)
}

func TestLocalModelAdapter_Embedding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/embeddings", r.URL.Path)
		result := struct {
			Data  []EmbeddingData `json:"data"`
			Model string          `json:"model"`
			Usage TokenUsage      `json:"usage"`
		}{
			Data:  []EmbeddingData{{Embedding: []float64{0.1, 0.2}, Index: 0}},
			Model: "local-embedding",
			Usage: TokenUsage{TotalTokens: 5},
		}
		json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderLocal,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewLocalModelAdapter(cfg)

	resp, err := adapter.Embedding(context.Background(), &EmbeddingRequest{Input: "test"})
	require.NoError(t, err)
	assert.Equal(t, 1, len(resp.Data))
}

func TestLocalModelAdapter_Completion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/completions", r.URL.Path)
		result := struct {
			ID      string             `json:"id"`
			Model   string             `json:"model"`
			Choices []CompletionChoice `json:"choices"`
			Usage   TokenUsage         `json:"usage"`
		}{
			ID:      "comp-local",
			Model:   "local-llm",
			Choices: []CompletionChoice{{Text: "Local completion", Index: 0, FinishReason: "stop"}},
			Usage:   TokenUsage{TotalTokens: 10},
		}
		json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderLocal,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewLocalModelAdapter(cfg)

	resp, err := adapter.Completion(context.Background(), &CompletionRequest{Prompt: "test"})
	require.NoError(t, err)
	assert.Equal(t, "Local completion", resp.Choices[0].Text)
}

func TestLocalModelAdapter_HealthCheck_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:   ProviderLocal,
		APIKey:     "test-key",
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}
	adapter := NewLocalModelAdapter(cfg)

	err := adapter.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestLocalModelAdapter_HealthCheck_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:      ProviderLocal,
		APIKey:        "test-key",
		BaseURL:       server.URL,
		Timeout:       5 * time.Second,
		MaxRetries:    1,
		RetryInterval: 10 * time.Millisecond,
	}
	adapter := NewLocalModelAdapter(cfg)

	err := adapter.HealthCheck(context.Background())
	assert.Error(t, err)
}

func TestModelSelector_Register(t *testing.T) {
	selector := NewModelSelector("gpt-4")
	cfg := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost"}
	adapter := NewOpenAIAdapter(cfg)

	selector.Register("gpt-4", adapter)
	models := selector.List()
	assert.Equal(t, 1, len(models))
	assert.Contains(t, models, "gpt-4")
}

func TestModelSelector_Select(t *testing.T) {
	selector := NewModelSelector("gpt-4")
	cfg := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost"}
	adapter := NewOpenAIAdapter(cfg)
	selector.Register("gpt-4", adapter)

	selected, err := selector.Select("gpt-4")
	require.NoError(t, err)
	assert.NotNil(t, selected)
}

func TestModelSelector_Select_Default(t *testing.T) {
	selector := NewModelSelector("gpt-4")
	cfg := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost"}
	adapter := NewOpenAIAdapter(cfg)
	selector.Register("gpt-4", adapter)

	selected, err := selector.Select("")
	require.NoError(t, err)
	assert.NotNil(t, selected)
}

func TestModelSelector_Select_NotFound(t *testing.T) {
	selector := NewModelSelector("gpt-4")
	_, err := selector.Select("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model not found")
}

func TestModelSelector_List(t *testing.T) {
	selector := NewModelSelector("gpt-4")
	cfg1 := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost"}
	cfg2 := &AdapterConfig{Provider: ProviderAnthropic, BaseURL: "http://localhost"}
	selector.Register("gpt-4", NewOpenAIAdapter(cfg1))
	selector.Register("claude-3", NewClaudeAdapter(cfg2))

	models := selector.List()
	assert.Equal(t, 2, len(models))
}

func TestLoadBalancer_RegisterSelector(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	selector := NewModelSelector("gpt-4")
	cfg := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost"}
	selector.Register("gpt-4", NewOpenAIAdapter(cfg))

	lb.RegisterSelector("openai", selector)
	selected, err := lb.Select("openai", "gpt-4")
	require.NoError(t, err)
	assert.NotNil(t, selected)
}

func TestLoadBalancer_Select_ProviderNotFound(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	_, err := lb.Select("nonexistent", "model")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider not found")
}

func TestLoadBalancer_SelectByStrategy_NoProviders(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	_, err := lb.SelectByStrategy()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no providers available")
}

func TestLoadBalancer_SelectByStrategy_RoundRobin(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	selector := NewModelSelector("gpt-4")
	cfg := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost"}
	selector.Register("gpt-4", NewOpenAIAdapter(cfg))
	lb.RegisterSelector("openai", selector)

	selected, err := lb.SelectByStrategy()
	require.NoError(t, err)
	assert.NotNil(t, selected)
}

func TestLoadBalancer_SelectByStrategy_Random(t *testing.T) {
	lb := NewLoadBalancer(StrategyRandom)
	selector := NewModelSelector("gpt-4")
	cfg := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost"}
	selector.Register("gpt-4", NewOpenAIAdapter(cfg))
	lb.RegisterSelector("openai", selector)

	selected, err := lb.SelectByStrategy()
	require.NoError(t, err)
	assert.NotNil(t, selected)
}

func TestLoadBalancer_SelectByStrategy_LeastConn(t *testing.T) {
	lb := NewLoadBalancer(StrategyLeastConn)
	selector := NewModelSelector("gpt-4")
	cfg := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost"}
	selector.Register("gpt-4", NewOpenAIAdapter(cfg))
	lb.RegisterSelector("openai", selector)

	selected, err := lb.SelectByStrategy()
	require.NoError(t, err)
	assert.NotNil(t, selected)
}

func TestLoadBalancer_GetAllModels(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	selector1 := NewModelSelector("gpt-4")
	cfg1 := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost"}
	selector1.Register("gpt-4", NewOpenAIAdapter(cfg1))
	lb.RegisterSelector("openai", selector1)

	models := lb.GetAllModels()
	assert.Equal(t, 1, len(models))
}

func TestLoadBalancer_GetAllModels_Empty(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	models := lb.GetAllModels()
	assert.Equal(t, 0, len(models))
}

func TestAIError_Error(t *testing.T) {
	err := &AIError{Message: "test error"}
	assert.Equal(t, "test error", err.Error())
}

func TestNewAIError(t *testing.T) {
	err := NewAIError("type1", "msg1", "code1", 400)
	assert.Equal(t, "type1", err.Type)
	assert.Equal(t, "msg1", err.Message)
	assert.Equal(t, "code1", err.Code)
	assert.Equal(t, 400, err.StatusCode)
}

func TestPredefinedErrors(t *testing.T) {
	assert.Equal(t, "invalid_request_error", ErrInvalidRequest.Type)
	assert.Equal(t, "authentication_error", ErrAuthentication.Type)
	assert.Equal(t, "permission_denied", ErrPermissionDenied.Type)
	assert.Equal(t, "not_found_error", ErrNotFound.Type)
	assert.Equal(t, "rate_limit_error", ErrRateLimitExceeded.Type)
	assert.Equal(t, "service_unavailable", ErrServiceUnavailable.Type)
	assert.Equal(t, "model_overloaded", ErrModelOverloaded.Type)
	assert.Equal(t, "context_length_exceeded", ErrContextLength.Type)
	assert.Equal(t, "content_filter", ErrContentFilter.Type)
}

func TestContextManager_Create(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, err := mgr.Create(ctx, "user1", "sess1", "You are helpful")
	require.NoError(t, err)
	assert.NotEmpty(t, conv.ConversationID)
	assert.Equal(t, "user1", conv.UserID)
	assert.Equal(t, "sess1", conv.SessionID)
	assert.Equal(t, "You are helpful", conv.SystemPrompt)
	assert.Equal(t, "active", conv.Status)
}

func TestContextManager_Create_NoPersist(t *testing.T) {
	cfg := &ContextConfig{Persist: false, DefaultTTL: time.Hour}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	conv, err := mgr.Create(ctx, "user1", "sess1", "")
	require.NoError(t, err)
	assert.NotEmpty(t, conv.ConversationID)
}

func TestContextManager_Get_Cached(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	got, err := mgr.Get(ctx, conv.ConversationID)
	require.NoError(t, err)
	assert.Equal(t, conv.ConversationID, got.ConversationID)
}

func TestContextManager_Get_Expired(t *testing.T) {
	cfg := &ContextConfig{Persist: false, DefaultTTL: -1 * time.Second}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	time.Sleep(10 * time.Millisecond)
	_, err := mgr.Get(ctx, conv.ConversationID)
	assert.Error(t, err)
	assert.Equal(t, ErrContextExpired, err)
}

func TestContextManager_Get_FromStorage(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "prompt")
	mgr.mu.Lock()
	delete(mgr.cache, conv.ConversationID)
	mgr.mu.Unlock()

	got, err := mgr.Get(ctx, conv.ConversationID)
	require.NoError(t, err)
	assert.Equal(t, conv.ConversationID, got.ConversationID)
}

func TestContextManager_AddMessage(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	err := mgr.AddMessage(ctx, conv.ConversationID, &ContextMessage{
		Role:       "user",
		Content:    "Hello",
		TokenCount: 5,
	})
	require.NoError(t, err)

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.Equal(t, 1, len(got.Messages))
	assert.Equal(t, 5, got.TokenStats.TotalInputTokens)
	assert.Equal(t, 1, got.TokenStats.MessageCount)
}

func TestContextManager_AddMessage_AssistantTokenCount(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	err := mgr.AddMessage(ctx, conv.ConversationID, &ContextMessage{
		Role:       "assistant",
		Content:    "Hi there",
		TokenCount: 10,
	})
	require.NoError(t, err)

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.Equal(t, 10, got.TokenStats.TotalOutputTokens)
}

func TestContextManager_AddMessage_AutoID(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	err := mgr.AddMessage(ctx, conv.ConversationID, &ContextMessage{
		Role:    "user",
		Content: "Hello",
	})
	require.NoError(t, err)

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.NotEmpty(t, got.Messages[0].ID)
	assert.False(t, got.Messages[0].Timestamp.IsZero())
}

func TestContextManager_AddMessage_Compression(t *testing.T) {
	cfg := &ContextConfig{
		Persist:              false,
		DefaultTTL:           time.Hour,
		EnableCompression:    true,
		CompressionThreshold: 3,
		KeepRecentMessages:   2,
		MaxMessages:          100,
	}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	for i := 0; i < 5; i++ {
		mgr.AddMessage(ctx, conv.ConversationID, &ContextMessage{
			Role:    "user",
			Content: fmt.Sprintf("Message %d", i),
		})
	}

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.LessOrEqual(t, len(got.Messages), 5)
}

func TestContextManager_AddMessage_MaxMessages(t *testing.T) {
	cfg := &ContextConfig{
		Persist:           false,
		DefaultTTL:        time.Hour,
		EnableCompression: false,
		MaxMessages:       3,
	}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	for i := 0; i < 5; i++ {
		mgr.AddMessage(ctx, conv.ConversationID, &ContextMessage{
			Role:    "user",
			Content: fmt.Sprintf("Message %d", i),
		})
	}

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.LessOrEqual(t, len(got.Messages), 3)
}

func TestContextManager_GetMessages(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	for i := 0; i < 5; i++ {
		mgr.AddMessage(ctx, conv.ConversationID, &ContextMessage{
			Role:    "user",
			Content: fmt.Sprintf("Message %d", i),
		})
	}

	msgs, err := mgr.GetMessages(ctx, conv.ConversationID, 3, 0)
	require.NoError(t, err)
	assert.Equal(t, 3, len(msgs))

	msgs, err = mgr.GetMessages(ctx, conv.ConversationID, 10, 3)
	require.NoError(t, err)
	assert.Equal(t, 2, len(msgs))

	msgs, err = mgr.GetMessages(ctx, conv.ConversationID, 10, 100)
	require.NoError(t, err)
	assert.Equal(t, 0, len(msgs))
}

func TestContextManager_UpdateSystemPrompt(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "Old prompt")
	err := mgr.UpdateSystemPrompt(ctx, conv.ConversationID, "New prompt")
	require.NoError(t, err)

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.Equal(t, "New prompt", got.SystemPrompt)
}

func TestContextManager_SetModel(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	err := mgr.SetModel(ctx, conv.ConversationID, "gpt-4", map[string]interface{}{"temperature": 0.7})
	require.NoError(t, err)

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.Equal(t, "gpt-4", got.Model)
	assert.Equal(t, 0.7, got.ModelParams["temperature"])
}

func TestContextManager_SetModel_NilParams(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	err := mgr.SetModel(ctx, conv.ConversationID, "gpt-4", nil)
	require.NoError(t, err)

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.Equal(t, "gpt-4", got.Model)
}

func TestContextManager_SetMetadata(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	err := mgr.SetMetadata(ctx, conv.ConversationID, map[string]interface{}{"key1": "val1"})
	require.NoError(t, err)

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.Equal(t, "val1", got.Metadata["key1"])
}

func TestContextManager_Delete(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	err := mgr.Delete(ctx, conv.ConversationID)
	require.NoError(t, err)

	_, err = mgr.Get(ctx, conv.ConversationID)
	assert.Error(t, err)
}

func TestContextManager_Delete_NonExistent(t *testing.T) {
	cfg := &ContextConfig{Persist: true, DefaultTTL: time.Hour}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	err := mgr.Delete(ctx, "nonexistent")
	assert.NoError(t, err)
}

func TestContextManager_ExtendTTL(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	err := mgr.ExtendTTL(ctx, conv.ConversationID, 2*time.Hour)
	require.NoError(t, err)

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.True(t, got.ExpiresAt.After(time.Now().Add(time.Hour)))
}

func TestContextManager_ListByUser(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	mgr.Create(ctx, "user1", "sess1", "")
	mgr.Create(ctx, "user1", "sess2", "")
	mgr.Create(ctx, "user2", "sess3", "")

	result, err := mgr.ListByUser(ctx, "user1", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, len(result))
}

func TestContextManager_Clear(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "System prompt")
	mgr.AddMessage(ctx, conv.ConversationID, &ContextMessage{Role: "user", Content: "Hello"})
	mgr.AddMessage(ctx, conv.ConversationID, &ContextMessage{Role: "system", Content: "System msg"})

	err := mgr.Clear(ctx, conv.ConversationID)
	require.NoError(t, err)

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.Equal(t, 1, len(got.Messages))
	assert.Equal(t, "system", got.Messages[0].Role)
	assert.Equal(t, 0, got.TokenStats.MessageCount)
}

func TestContextManager_Compress(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	for i := 0; i < 15; i++ {
		mgr.AddMessage(ctx, conv.ConversationID, &ContextMessage{
			Role:    "user",
			Content: fmt.Sprintf("Message %d", i),
		})
	}

	err := mgr.Compress(ctx, conv.ConversationID)
	require.NoError(t, err)
}

func TestContextManager_CleanExpired(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	expiredConv := &ConversationContext{
		ConversationID: "expired-conv",
		UserID:         "user1",
		ExpiresAt:      time.Now().Add(-1 * time.Hour),
	}
	activeConv := &ConversationContext{
		ConversationID: "active-conv",
		UserID:         "user1",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	storage.Save(ctx, expiredConv)
	storage.Save(ctx, activeConv)

	cfg := &ContextConfig{Persist: true, DefaultTTL: time.Hour}
	mgr := NewContextManager(cfg, storage)

	count, err := mgr.CleanExpired(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(1))
}

func TestContextManager_GetStats(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	mgr.AddMessage(ctx, conv.ConversationID, &ContextMessage{Role: "user", Content: "Hello", TokenCount: 5})

	stats, err := mgr.GetStats(ctx, conv.ConversationID)
	require.NoError(t, err)
	assert.Equal(t, 5, stats.TotalInputTokens)
	assert.Equal(t, 1, stats.MessageCount)
}

func TestContextManager_BuildChatRequest(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "You are helpful")
	mgr.AddMessage(ctx, conv.ConversationID, &ContextMessage{Role: "user", Content: "Hello"})

	req, err := mgr.BuildChatRequest(ctx, conv.ConversationID, "How are you?")
	require.NoError(t, err)
	assert.Equal(t, conv.ConversationID, req.ConversationID)
	assert.Equal(t, 3, len(req.Messages))
	assert.Equal(t, "system", req.Messages[0].Role)
	assert.Equal(t, "user", req.Messages[1].Role)
	assert.Equal(t, "How are you?", req.Messages[2].Content)
}

func TestContextManager_BuildChatRequest_WithModelParams(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	mgr.SetModel(ctx, conv.ConversationID, "gpt-4", map[string]interface{}{
		"temperature": 0.5,
		"top_p":       0.9,
	})

	req, err := mgr.BuildChatRequest(ctx, conv.ConversationID, "test")
	require.NoError(t, err)
	assert.Equal(t, "gpt-4", req.Model)
	assert.Equal(t, 0.5, req.Temperature)
	assert.Equal(t, 0.9, req.TopP)
}

func TestMemoryContextStorage_Save(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	conv := &ConversationContext{
		ConversationID: "conv1",
		UserID:         "user1",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	err := storage.Save(ctx, conv)
	require.NoError(t, err)
}

func TestMemoryContextStorage_Load(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	conv := &ConversationContext{
		ConversationID: "conv1",
		UserID:         "user1",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	storage.Save(ctx, conv)

	loaded, err := storage.Load(ctx, "conv1")
	require.NoError(t, err)
	assert.Equal(t, "conv1", loaded.ConversationID)
}

func TestMemoryContextStorage_Load_NotFound(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	_, err := storage.Load(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrContextNotFound, err)
}

func TestMemoryContextStorage_Delete(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	conv := &ConversationContext{
		ConversationID: "conv1",
		UserID:         "user1",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	storage.Save(ctx, conv)
	err := storage.Delete(ctx, "conv1")
	require.NoError(t, err)

	_, err = storage.Load(ctx, "conv1")
	assert.Error(t, err)
}

func TestMemoryContextStorage_Exists(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	conv := &ConversationContext{
		ConversationID: "conv1",
		UserID:         "user1",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	storage.Save(ctx, conv)

	exists, err := storage.Exists(ctx, "conv1")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = storage.Exists(ctx, "nonexistent")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMemoryContextStorage_SetTTL(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	conv := &ConversationContext{
		ConversationID: "conv1",
		UserID:         "user1",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	storage.Save(ctx, conv)

	err := storage.SetTTL(ctx, "conv1", 2*time.Hour)
	require.NoError(t, err)

	loaded, _ := storage.Load(ctx, "conv1")
	assert.True(t, loaded.ExpiresAt.After(time.Now().Add(time.Hour)))
}

func TestMemoryContextStorage_SetTTL_NotFound(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	err := storage.SetTTL(ctx, "nonexistent", time.Hour)
	assert.Error(t, err)
}

func TestMemoryContextStorage_ListByUser(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		conv := &ConversationContext{
			ConversationID: fmt.Sprintf("conv%d", i),
			UserID:         "user1",
			ExpiresAt:      time.Now().Add(time.Hour),
		}
		storage.Save(ctx, conv)
	}

	result, err := storage.ListByUser(ctx, "user1", 3, 0)
	require.NoError(t, err)
	assert.Equal(t, 3, len(result))

	result, err = storage.ListByUser(ctx, "user1", 10, 3)
	require.NoError(t, err)
	assert.Equal(t, 2, len(result))

	result, err = storage.ListByUser(ctx, "user1", 10, 100)
	require.NoError(t, err)
	assert.Equal(t, 0, len(result))
}

func TestMemoryContextStorage_CleanExpired(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	conv1 := &ConversationContext{
		ConversationID: "conv1",
		UserID:         "user1",
		ExpiresAt:      time.Now().Add(-1 * time.Hour),
	}
	conv2 := &ConversationContext{
		ConversationID: "conv2",
		UserID:         "user1",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	storage.Save(ctx, conv1)
	storage.Save(ctx, conv2)

	count, err := storage.CleanExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRedisContextStorage_Save(t *testing.T) {
	mockRedis := &mockRedisClient{data: make(map[string]string)}
	storage := NewRedisContextStorage(mockRedis, "")
	ctx := context.Background()

	conv := &ConversationContext{
		ConversationID: "conv1",
		UserID:         "user1",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	err := storage.Save(ctx, conv)
	require.NoError(t, err)
}

func TestRedisContextStorage_Load(t *testing.T) {
	mockRedis := &mockRedisClient{data: make(map[string]string)}
	storage := NewRedisContextStorage(mockRedis, "")
	ctx := context.Background()

	conv := &ConversationContext{
		ConversationID: "conv1",
		UserID:         "user1",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	storage.Save(ctx, conv)

	loaded, err := storage.Load(ctx, "conv1")
	require.NoError(t, err)
	assert.Equal(t, "conv1", loaded.ConversationID)
}

func TestRedisContextStorage_Delete(t *testing.T) {
	mockRedis := &mockRedisClient{data: make(map[string]string)}
	storage := NewRedisContextStorage(mockRedis, "")
	ctx := context.Background()

	conv := &ConversationContext{
		ConversationID: "conv1",
		UserID:         "user1",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	storage.Save(ctx, conv)
	err := storage.Delete(ctx, "conv1")
	require.NoError(t, err)
}

func TestRedisContextStorage_Exists(t *testing.T) {
	mockRedis := &mockRedisClient{data: make(map[string]string)}
	storage := NewRedisContextStorage(mockRedis, "")
	ctx := context.Background()

	conv := &ConversationContext{
		ConversationID: "conv1",
		UserID:         "user1",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	storage.Save(ctx, conv)

	exists, err := storage.Exists(ctx, "conv1")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRedisContextStorage_SetTTL(t *testing.T) {
	mockRedis := &mockRedisClient{data: make(map[string]string)}
	storage := NewRedisContextStorage(mockRedis, "")
	ctx := context.Background()

	err := storage.SetTTL(ctx, "conv1", time.Hour)
	require.NoError(t, err)
}

func TestRedisContextStorage_ListByUser(t *testing.T) {
	mockRedis := &mockRedisClient{data: make(map[string]string)}
	storage := NewRedisContextStorage(mockRedis, "")
	ctx := context.Background()

	_, err := storage.ListByUser(ctx, "user1", 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestRedisContextStorage_CleanExpired(t *testing.T) {
	mockRedis := &mockRedisClient{data: make(map[string]string)}
	storage := NewRedisContextStorage(mockRedis, "")
	ctx := context.Background()

	count, err := storage.CleanExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRedisContextStorage_DefaultPrefix(t *testing.T) {
	mockRedis := &mockRedisClient{data: make(map[string]string)}
	storage := NewRedisContextStorage(mockRedis, "")
	assert.Equal(t, "ai:context:", storage.keyPrefix)
}

func TestRedisContextStorage_CustomPrefix(t *testing.T) {
	mockRedis := &mockRedisClient{data: make(map[string]string)}
	storage := NewRedisContextStorage(mockRedis, "custom:")
	assert.Equal(t, "custom:", storage.keyPrefix)
}

func TestTemplateManager_Create(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{
		Name:    "Test",
		Content: "Hello {{.name}}",
		Type:    TemplateTypeChat,
		Enabled: true,
	}
	err := mgr.Create(ctx, tmpl)
	require.NoError(t, err)
	assert.NotEmpty(t, tmpl.ID)
	assert.Equal(t, "1.0.0", tmpl.Version)
}

func TestTemplateManager_Create_ValidationError(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	err := mgr.Create(ctx, &PromptTemplate{Name: "", Content: "test"})
	assert.Error(t, err)

	err = mgr.Create(ctx, &PromptTemplate{Name: "Test", Content: ""})
	assert.Error(t, err)
}

func TestTemplateManager_Create_InvalidSyntax(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	err := mgr.Create(ctx, &PromptTemplate{Name: "Test", Content: "{{.invalid}"})
	assert.Error(t, err)
}

func TestTemplateManager_GetByName(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{Name: "MyTemplate", Content: "Hello", Enabled: true}
	mgr.Create(ctx, tmpl)

	found, err := mgr.GetByName(ctx, "MyTemplate")
	require.NoError(t, err)
	assert.Equal(t, "MyTemplate", found.Name)
}

func TestTemplateManager_GetByName_NotFound(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	_, err := mgr.GetByName(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestTemplateManager_Update(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{Name: "Test", Content: "Hello", Enabled: true}
	mgr.Create(ctx, tmpl)

	tmpl.Content = "Updated content"
	err := mgr.Update(ctx, tmpl)
	require.NoError(t, err)
	assert.Equal(t, "1.1.0", tmpl.Version)
}

func TestTemplateManager_Update_NotFound(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	err := mgr.Update(ctx, &PromptTemplate{ID: "nonexistent", Name: "Test", Content: "Hello"})
	assert.Error(t, err)
}

func TestTemplateManager_Delete(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{Name: "Test", Content: "Hello", Enabled: true}
	mgr.Create(ctx, tmpl)

	err := mgr.Delete(ctx, tmpl.ID)
	require.NoError(t, err)

	_, err = mgr.Get(ctx, tmpl.ID)
	assert.Error(t, err)
}

func TestTemplateManager_Delete_NotFound(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	err := mgr.Delete(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestTemplateManager_List_WithFilter(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	mgr.Create(ctx, &PromptTemplate{Name: "Chat Template", Content: "Hi", Type: TemplateTypeChat, Enabled: true})
	mgr.Create(ctx, &PromptTemplate{Name: "Completion Template", Content: "Go", Type: TemplateTypeCompletion, Enabled: true})
	mgr.Create(ctx, &PromptTemplate{Name: "Disabled", Content: "Off", Type: TemplateTypeChat, Enabled: false})

	filter := &TemplateFilter{Type: TemplateTypeChat}
	list, err := mgr.List(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 2, len(list))

	enabled := true
	filter = &TemplateFilter{Enabled: &enabled}
	list, err = mgr.List(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 2, len(list))
}

func TestTemplateManager_List_FilterByName(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	mgr.Create(ctx, &PromptTemplate{Name: "Chat Template", Content: "Hi", Enabled: true})
	mgr.Create(ctx, &PromptTemplate{Name: "Report Template", Content: "Go", Enabled: true})

	filter := &TemplateFilter{Name: "Chat"}
	list, err := mgr.List(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))
}

func TestTemplateManager_List_FilterByTags(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	mgr.Create(ctx, &PromptTemplate{Name: "T1", Content: "Hi", Tags: []string{"device"}, Enabled: true})
	mgr.Create(ctx, &PromptTemplate{Name: "T2", Content: "Go", Tags: []string{"alarm"}, Enabled: true})

	filter := &TemplateFilter{Tags: []string{"device"}}
	list, err := mgr.List(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))
}

func TestTemplateManager_List_FilterByCreatedBy(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	mgr.Create(ctx, &PromptTemplate{Name: "T1", Content: "Hi", CreatedBy: "admin", Enabled: true})
	mgr.Create(ctx, &PromptTemplate{Name: "T2", Content: "Go", CreatedBy: "user", Enabled: true})

	filter := &TemplateFilter{CreatedBy: "admin"}
	list, err := mgr.List(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))
}

func TestTemplateManager_List_FilterByCreatedTime(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	mgr.Create(ctx, &PromptTemplate{Name: "T1", Content: "Hi", Enabled: true})
	time.Sleep(10 * time.Millisecond)
	after := time.Now()
	mgr.Create(ctx, &PromptTemplate{Name: "T2", Content: "Go", Enabled: true})

	filter := &TemplateFilter{CreatedAfter: &after}
	list, err := mgr.List(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list))
}

func TestTemplateManager_Render_WithEscapeFunc(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{
		Name:    "Escape Test",
		Content: "Hello {{.name}}",
		Enabled: true,
	}
	mgr.Create(ctx, tmpl)

	opts := &RenderOptions{
		EscapeFunc: func(s string) string { return "[" + s + "]" },
	}
	rendered, err := mgr.Render(ctx, tmpl.ID, map[string]interface{}{"name": "World"}, opts)
	require.NoError(t, err)
	assert.Equal(t, "[Hello World]", rendered)
}

func TestTemplateManager_Render_StrictMode(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{
		Name:    "Strict Test",
		Content: "Hello {{.name}}",
		Variables: []TemplateVariable{
			{Name: "name", Type: "string", Required: true},
		},
		Enabled: true,
	}
	mgr.Create(ctx, tmpl)

	opts := &RenderOptions{Strict: true}
	_, err := mgr.Render(ctx, tmpl.ID, map[string]interface{}{}, opts)
	assert.Error(t, err)
}

func TestTemplateManager_Render_ValidationPattern(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{
		Name:    "Pattern Test",
		Content: "ID: {{.id}}",
		Variables: []TemplateVariable{
			{Name: "id", Type: "string", Validation: "^[a-z]+$"},
		},
		Enabled: true,
	}
	mgr.Create(ctx, tmpl)

	_, err := mgr.Render(ctx, tmpl.ID, map[string]interface{}{"id": "abc"}, nil)
	require.NoError(t, err)

	_, err = mgr.Render(ctx, tmpl.ID, map[string]interface{}{"id": "ABC"}, nil)
	assert.Error(t, err)
}

func TestTemplateManager_Render_EnumValidation(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{
		Name:    "Enum Test",
		Content: "Type: {{.type}}",
		Variables: []TemplateVariable{
			{Name: "type", Type: "string", Enum: []string{"solar", "wind"}},
		},
		Enabled: true,
	}
	mgr.Create(ctx, tmpl)

	_, err := mgr.Render(ctx, tmpl.ID, map[string]interface{}{"type": "solar"}, nil)
	require.NoError(t, err)

	_, err = mgr.Render(ctx, tmpl.ID, map[string]interface{}{"type": "nuclear"}, nil)
	assert.Error(t, err)
}

func TestTemplateManager_Render_VariableTypeValidation(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{
		Name:    "Type Test",
		Content: "Val: {{.val}}",
		Variables: []TemplateVariable{
			{Name: "val", Type: "number"},
		},
		Enabled: true,
	}
	mgr.Create(ctx, tmpl)

	_, err := mgr.Render(ctx, tmpl.ID, map[string]interface{}{"val": 42}, nil)
	require.NoError(t, err)

	_, err = mgr.Render(ctx, tmpl.ID, map[string]interface{}{"val": "not-a-number"}, nil)
	assert.Error(t, err)
}

func TestTemplateManager_Render_BooleanType(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{
		Name:    "Bool Test",
		Content: "Flag: {{.flag}}",
		Variables: []TemplateVariable{
			{Name: "flag", Type: "boolean"},
		},
		Enabled: true,
	}
	mgr.Create(ctx, tmpl)

	_, err := mgr.Render(ctx, tmpl.ID, map[string]interface{}{"flag": true}, nil)
	require.NoError(t, err)

	_, err = mgr.Render(ctx, tmpl.ID, map[string]interface{}{"flag": "yes"}, nil)
	assert.Error(t, err)
}

func TestTemplateManager_Render_ArrayType(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{
		Name:    "Array Test",
		Content: "Items: {{.items}}",
		Variables: []TemplateVariable{
			{Name: "items", Type: "array"},
		},
		Enabled: true,
	}
	mgr.Create(ctx, tmpl)

	_, err := mgr.Render(ctx, tmpl.ID, map[string]interface{}{"items": []interface{}{1, 2, 3}}, nil)
	require.NoError(t, err)

	_, err = mgr.Render(ctx, tmpl.ID, map[string]interface{}{"items": "not-array"}, nil)
	assert.Error(t, err)
}

func TestTemplateManager_Render_ObjectType(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{
		Name:    "Object Test",
		Content: "Data: {{.data}}",
		Variables: []TemplateVariable{
			{Name: "data", Type: "object"},
		},
		Enabled: true,
	}
	mgr.Create(ctx, tmpl)

	_, err := mgr.Render(ctx, tmpl.ID, map[string]interface{}{"data": map[string]interface{}{"key": "val"}}, nil)
	require.NoError(t, err)

	_, err = mgr.Render(ctx, tmpl.ID, map[string]interface{}{"data": "not-object"}, nil)
	assert.Error(t, err)
}

func TestTemplateManager_GetVersions(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{Name: "Test", Content: "Hello", Enabled: true}
	mgr.Create(ctx, tmpl)

	versions, err := mgr.GetVersions(ctx, tmpl.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(versions))
	assert.Equal(t, "1.0.0", versions[0].Version)
}

func TestTemplateManager_Rollback(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{Name: "Test", Content: "Hello", Enabled: true}
	mgr.Create(ctx, tmpl)

	tmpl.Content = "Updated"
	mgr.Update(ctx, tmpl)

	err := mgr.Rollback(ctx, tmpl.ID, "1.0.0")
	require.NoError(t, err)
}

func TestTemplateManager_Rollback_VersionNotFound(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{Name: "Test", Content: "Hello", Enabled: true}
	mgr.Create(ctx, tmpl)

	err := mgr.Rollback(ctx, tmpl.ID, "99.0.0")
	assert.Error(t, err)
}

func TestTemplateManager_Duplicate(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{Name: "Original", Content: "Hello", Type: "chat", Enabled: true}
	mgr.Create(ctx, tmpl)

	dup, err := mgr.Duplicate(ctx, tmpl.ID, "Copy")
	require.NoError(t, err)
	assert.Equal(t, "Copy", dup.Name)
	assert.NotEqual(t, tmpl.ID, dup.ID)
}

func TestTemplateManager_Export(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{Name: "Test", Content: "Hello", Enabled: true}
	mgr.Create(ctx, tmpl)

	data, err := mgr.Export(ctx, []string{tmpl.ID})
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestTemplateManager_Export_NotFound(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	_, err := mgr.Export(ctx, []string{"nonexistent"})
	assert.Error(t, err)
}

func TestTemplateManager_Import(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	templates := []*PromptTemplate{
		{Name: "Imported1", Content: "Hello", Enabled: true},
		{Name: "Imported2", Content: "World", Enabled: true},
	}
	data, _ := json.Marshal(templates)

	err := mgr.Import(ctx, data, false)
	require.NoError(t, err)

	list, _ := mgr.List(ctx, nil)
	assert.Equal(t, 2, len(list))
}

func TestTemplateManager_Import_Overwrite(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	mgr.Create(ctx, &PromptTemplate{Name: "Existing", Content: "Old", Enabled: true})

	templates := []*PromptTemplate{
		{Name: "Existing", Content: "New", Enabled: true},
	}
	data, _ := json.Marshal(templates)

	err := mgr.Import(ctx, data, true)
	require.NoError(t, err)
}

func TestTemplateManager_Import_InvalidData(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	err := mgr.Import(ctx, []byte("invalid json"), false)
	assert.Error(t, err)
}

func TestSummaryCompressor_Compress(t *testing.T) {
	compressor := NewSummaryCompressor()
	ctx := context.Background()

	messages := make([]ContextMessage, 15)
	for i := 0; i < 15; i++ {
		messages[i] = ContextMessage{
			ID:        fmt.Sprintf("msg_%d", i),
			Role:      "user",
			Content:   fmt.Sprintf("Message %d", i),
			Timestamp: time.Now(),
		}
	}

	conv := &ConversationContext{
		ConversationID: "conv1",
		Messages:       messages,
	}

	result, err := compressor.Compress(ctx, conv)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(result.Messages), 15)
}

func TestSummaryCompressor_Compress_ShortMessages(t *testing.T) {
	compressor := NewSummaryCompressor()
	ctx := context.Background()

	conv := &ConversationContext{
		ConversationID: "conv1",
		Messages: []ContextMessage{
			{ID: "m1", Role: "user", Content: "Hi"},
			{ID: "m2", Role: "assistant", Content: "Hello"},
		},
	}

	result, err := compressor.Compress(ctx, conv)
	require.NoError(t, err)
	assert.Equal(t, 2, len(result.Messages))
}

func TestSummaryCompressor_CreateSummary(t *testing.T) {
	compressor := NewSummaryCompressor()
	messages := []ContextMessage{
		{Role: "user", Content: "This is a very long message that should be truncated because it exceeds the two hundred character limit for summary purposes in the compression algorithm"},
		{Role: "assistant", Content: "Short reply"},
	}
	summary := compressor.createSummary(messages)
	assert.Contains(t, summary, "[user]")
	assert.Contains(t, summary, "[assistant]")
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

func TestIncrementVersion(t *testing.T) {
	assert.Equal(t, "1.1.0", incrementVersion("1.0.0"))
	assert.Equal(t, "2.3.0", incrementVersion("2.2.0"))
	assert.Equal(t, "1.0.1", incrementVersion("invalid"))
}

func TestBuiltInTemplates(t *testing.T) {
	assert.Equal(t, 3, len(BuiltInTemplates))
	assert.Equal(t, "builtin-chat-default", BuiltInTemplates[0].ID)
	assert.Equal(t, "builtin-alarm-analysis", BuiltInTemplates[1].ID)
	assert.Equal(t, "builtin-report-generation", BuiltInTemplates[2].ID)
}

func TestServiceConfig_Struct(t *testing.T) {
	cfg := ServiceConfig{
		DefaultModel:        "gpt-4",
		APIKey:              "key",
		BaseURL:             "http://api",
		Timeout:             30 * time.Second,
		MaxRetries:          3,
		DefaultTemperature:  0.7,
		DefaultMaxTokens:    4096,
		EnableCache:         true,
		CacheTTL:            5 * time.Minute,
	}
	assert.Equal(t, "gpt-4", cfg.DefaultModel)
	assert.True(t, cfg.EnableCache)
}

func TestConcurrentModelSelector(t *testing.T) {
	selector := NewModelSelector("gpt-4")
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cfg := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost"}
			selector.Register(fmt.Sprintf("model-%d", i), NewOpenAIAdapter(cfg))
		}(i)
	}
	wg.Wait()

	models := selector.List()
	assert.Equal(t, 10, len(models))
}

type mockRedisClient struct {
	data map[string]string
	mu   sync.RWMutex
}

func (m *mockRedisClient) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	if !ok {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return val, nil
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch v := value.(type) {
	case string:
		m.data[key] = v
	case []byte:
		m.data[key] = string(v)
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		m.data[key] = string(data)
	}
	return nil
}

func (m *mockRedisClient) Del(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

func (m *mockRedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var count int64
	for _, key := range keys {
		if _, ok := m.data[key]; ok {
			count++
		}
	}
	return count, nil
}

func (m *mockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}

func (m *mockRedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return nil, nil
}

func (m *mockRedisClient) RPush(ctx context.Context, key string, values ...interface{}) error {
	return nil
}
