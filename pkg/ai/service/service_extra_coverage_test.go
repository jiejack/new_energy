package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIAdapter_StreamChat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/chat/completions", r.URL.Path)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher, _ := w.(http.Flusher)
		fmt.Fprintf(w, "data: {\"id\":\"chat-1\",\"model\":\"gpt-4\",\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"Hello\"},\"finish_reason\":\"\"}]}\n\n")
		flusher.Flush()
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
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

	chunkChan, err := adapter.StreamChat(context.Background(), &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
	assert.NotNil(t, chunkChan)
}

func TestOpenAIAdapter_StreamChat_ErrorResponse(t *testing.T) {
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

	_, err := adapter.StreamChat(context.Background(), &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	assert.Error(t, err)
}

func TestOpenAIAdapter_StreamChat_ServerError(t *testing.T) {
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

	_, err := adapter.StreamChat(context.Background(), &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	assert.Error(t, err)
}

func TestOpenAIAdapter_parseErrorResponse_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("not json"))
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
	assert.Contains(t, err.Error(), "http error")
}

func TestOpenAIAdapter_Embedding_WithModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "text-embedding-custom", req["model"])
		w.WriteHeader(http.StatusOK)
		resp := openAIEmbeddingResponse{Model: "text-embedding-custom"}
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

	_, err := adapter.Embedding(context.Background(), &EmbeddingRequest{
		Input: "test",
		Model: "text-embedding-custom",
	})
	require.NoError(t, err)
}

func TestOpenAIAdapter_Completion_WithModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "gpt-3.5-turbo-instruct", req["model"])
		w.WriteHeader(http.StatusOK)
		resp := openAICompletionResponse{ID: "comp-1", Model: "gpt-3.5-turbo-instruct", Created: time.Now().Unix()}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:      ProviderOpenAI,
		APIKey:        "test-key",
		BaseURL:       server.URL,
		DefaultModel:  "gpt-3.5-turbo-instruct",
		Timeout:       5 * time.Second,
		MaxRetries:    1,
	}
	adapter := NewOpenAIAdapter(cfg)

	_, err := adapter.Completion(context.Background(), &CompletionRequest{
		Prompt: "test",
	})
	require.NoError(t, err)
}

func TestOpenAIAdapter_GetModelName(t *testing.T) {
	cfg := &AdapterConfig{
		Provider:     ProviderOpenAI,
		APIKey:       "test-key",
		BaseURL:      "http://localhost",
		DefaultModel: "gpt-4-turbo",
	}
	adapter := NewOpenAIAdapter(cfg)
	assert.Equal(t, "gpt-4-turbo", adapter.config.DefaultModel)
}

func TestClaudeAdapter_StreamChat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/messages", r.URL.Path)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher, _ := w.(http.Flusher)
		fmt.Fprintf(w, `{"type":"content_block_delta","delta":{"type":"text_delta","text":"Hi"}}`+"\n")
		flusher.Flush()
		fmt.Fprintf(w, `{"type":"message_stop"}`+"\n")
		flusher.Flush()
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

	chunkChan, err := adapter.StreamChat(context.Background(), &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
	assert.NotNil(t, chunkChan)

	time.Sleep(100 * time.Millisecond)
}

func TestClaudeAdapter_StreamChat_ErrorResponse(t *testing.T) {
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

	_, err := adapter.StreamChat(context.Background(), &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	assert.Error(t, err)
}

func TestClaudeAdapter_StreamChat_ServerError(t *testing.T) {
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

	_, err := adapter.StreamChat(context.Background(), &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	assert.Error(t, err)
}

func TestClaudeAdapter_GetModelName(t *testing.T) {
	cfg := &AdapterConfig{
		Provider:     ProviderAnthropic,
		APIKey:       "test-key",
		BaseURL:      "http://localhost",
		DefaultModel: "claude-3-sonnet",
	}
	adapter := NewClaudeAdapter(cfg)
	assert.Equal(t, "claude-3-sonnet", adapter.config.DefaultModel)
}

func TestClaudeAdapter_Completion_WithModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := claudeResponse{
			ID:   "msg-comp",
			Role: "assistant",
			Model: "claude-3-opus-20240229",
		}
		resp.Content = append(resp.Content, struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{Type: "text", Text: "Result"})
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

	_, err := adapter.Completion(context.Background(), &CompletionRequest{
		Prompt: "test",
		Model:  "claude-3-opus-20240229",
	})
	require.NoError(t, err)
}

func TestLocalModelAdapter_StreamChat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		result := struct {
			ID      string `json:"id"`
			Model   string `json:"model"`
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}{
			ID:    "local-stream-1",
			Model: "local-llm",
			Choices: []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			}{{Delta: struct {
				Content string `json:"content"`
			}{Content: "Hello"}, FinishReason: "stop"}},
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

	chunkChan, err := adapter.StreamChat(context.Background(), &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
	assert.NotNil(t, chunkChan)

	time.Sleep(100 * time.Millisecond)
}

func TestLocalModelAdapter_StreamChat_ErrorResponse(t *testing.T) {
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

	_, err := adapter.StreamChat(context.Background(), &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	assert.Error(t, err)
}

func TestLocalModelAdapter_GetModelName(t *testing.T) {
	cfg := &AdapterConfig{
		Provider:     ProviderLocal,
		APIKey:       "test-key",
		BaseURL:      "http://localhost",
		DefaultModel: "my-llm",
	}
	adapter := NewLocalModelAdapter(cfg)
	assert.Equal(t, "my-llm", adapter.config.DefaultModel)
}

func TestLocalModelAdapter_Embedding_WithModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := struct {
			Data  []EmbeddingData `json:"data"`
			Model string          `json:"model"`
			Usage TokenUsage      `json:"usage"`
		}{
			Data:  []EmbeddingData{{Embedding: []float64{0.1}, Index: 0}},
			Model: "local-emb",
			Usage: TokenUsage{TotalTokens: 3},
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

	_, err := adapter.Embedding(context.Background(), &EmbeddingRequest{
		Input: "test",
		Model: "local-emb",
	})
	require.NoError(t, err)
}

func TestLocalModelAdapter_Completion_WithModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := struct {
			ID      string             `json:"id"`
			Model   string             `json:"model"`
			Choices []CompletionChoice `json:"choices"`
			Usage   TokenUsage         `json:"usage"`
		}{
			ID:      "comp-1",
			Model:   "local-llm",
			Choices: []CompletionChoice{{Text: "Result", Index: 0, FinishReason: "stop"}},
			Usage:   TokenUsage{TotalTokens: 5},
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

	_, err := adapter.Completion(context.Background(), &CompletionRequest{
		Prompt: "test",
		Model:  "local-llm",
	})
	require.NoError(t, err)
}

func TestContextManager_ExtendTTL_WithPersist(t *testing.T) {
	cfg := &ContextConfig{Persist: true, DefaultTTL: time.Hour}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	err := mgr.ExtendTTL(ctx, conv.ConversationID, 2*time.Hour)
	require.NoError(t, err)

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.True(t, got.ExpiresAt.After(time.Now().Add(time.Hour)))
}

func TestContextManager_ExtendTTL_NotFound(t *testing.T) {
	cfg := &ContextConfig{Persist: false, DefaultTTL: time.Hour}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	err := mgr.ExtendTTL(ctx, "nonexistent", time.Hour)
	assert.Error(t, err)
}

func TestContextManager_CleanExpired_NoExpired(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	activeConv := &ConversationContext{
		ConversationID: "active-conv",
		UserID:         "user1",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	storage.Save(ctx, activeConv)

	cfg := &ContextConfig{Persist: true, DefaultTTL: time.Hour}
	mgr := NewContextManager(cfg, storage)

	count, err := mgr.CleanExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestContextManager_GetStats_NotFound(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	_, err := mgr.GetStats(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestContextManager_Compress_NotFound(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	err := mgr.Compress(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestContextManager_BuildChatRequest_NoSystemPrompt(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	req, err := mgr.BuildChatRequest(ctx, conv.ConversationID, "test")
	require.NoError(t, err)
	assert.Equal(t, 1, len(req.Messages))
	assert.Equal(t, "user", req.Messages[0].Role)
}

func TestContextManager_BuildChatRequest_NotFound(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	_, err := mgr.BuildChatRequest(ctx, "nonexistent", "test")
	assert.Error(t, err)
}

func TestContextManager_Clear_WithPersist(t *testing.T) {
	cfg := &ContextConfig{Persist: true, DefaultTTL: time.Hour}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "System prompt")
	mgr.AddMessage(ctx, conv.ConversationID, &ContextMessage{Role: "user", Content: "Hello"})

	err := mgr.Clear(ctx, conv.ConversationID)
	require.NoError(t, err)

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.Equal(t, 0, got.TokenStats.MessageCount)
}

func TestContextManager_Clear_NotFound(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	err := mgr.Clear(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestContextManager_Delete_WithPersist(t *testing.T) {
	cfg := &ContextConfig{Persist: true, DefaultTTL: time.Hour}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	err := mgr.Delete(ctx, conv.ConversationID)
	require.NoError(t, err)

	_, err = mgr.Get(ctx, conv.ConversationID)
	assert.Error(t, err)
}

func TestContextManager_UpdateSystemPrompt_WithPersist(t *testing.T) {
	cfg := &ContextConfig{Persist: true, DefaultTTL: time.Hour}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "Old")
	err := mgr.UpdateSystemPrompt(ctx, conv.ConversationID, "New")
	require.NoError(t, err)

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.Equal(t, "New", got.SystemPrompt)
}

func TestContextManager_SetModel_WithPersist(t *testing.T) {
	cfg := &ContextConfig{Persist: true, DefaultTTL: time.Hour}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	err := mgr.SetModel(ctx, conv.ConversationID, "gpt-4", nil)
	require.NoError(t, err)

	got, _ := mgr.Get(ctx, conv.ConversationID)
	assert.Equal(t, "gpt-4", got.Model)
}

func TestContextManager_SetMetadata_WithPersist(t *testing.T) {
	cfg := &ContextConfig{Persist: true, DefaultTTL: time.Hour}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	err := mgr.SetMetadata(ctx, conv.ConversationID, map[string]interface{}{"key": "val"})
	require.NoError(t, err)
}

func TestContextManager_AddMessage_WithPersist(t *testing.T) {
	cfg := &ContextConfig{Persist: true, DefaultTTL: time.Hour}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "")
	err := mgr.AddMessage(ctx, conv.ConversationID, &ContextMessage{
		Role:       "user",
		Content:    "Hello",
		TokenCount: 5,
	})
	require.NoError(t, err)
}

func TestContextManager_AddMessage_NotFound_WithPersist(t *testing.T) {
	cfg := &ContextConfig{Persist: true, DefaultTTL: time.Hour}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	err := mgr.AddMessage(ctx, "nonexistent", &ContextMessage{Role: "user", Content: "Hello"})
	assert.Error(t, err)
}

func TestMemoryContextStorage_Save_WithTTL(t *testing.T) {
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

func TestMemoryContextStorage_Delete_NotFound(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	err := storage.Delete(ctx, "nonexistent")
	assert.NoError(t, err)
}

func TestMemoryContextStorage_CleanExpired_None(t *testing.T) {
	storage := NewMemoryContextStorage()
	ctx := context.Background()

	count, err := storage.CleanExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestTemplateManager_GetVersions_FromStorage(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{Name: "Test", Content: "Hello", Enabled: true}
	mgr.Create(ctx, tmpl)

	versions, err := mgr.GetVersions(ctx, tmpl.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(versions))
}

func TestTemplateManager_GetVersions_NotInCache(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{Name: "Test", Content: "Hello", Enabled: true}
	mgr.Create(ctx, tmpl)

	mgr.mu.Lock()
	delete(mgr.versions, tmpl.ID)
	mgr.mu.Unlock()

	versions, err := mgr.GetVersions(ctx, tmpl.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(versions))
}

func TestTemplateManager_GetVersions_NoVersions(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	_, err := mgr.GetVersions(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestTemplateManager_Duplicate_NotFound(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	_, err := mgr.Duplicate(ctx, "nonexistent", "Copy")
	assert.Error(t, err)
}

func TestTemplateManager_Render_NotFound(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	_, err := mgr.Render(ctx, "nonexistent", map[string]interface{}{}, nil)
	assert.Error(t, err)
}

func TestTemplateManager_Render_DisabledTemplate(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{Name: "Disabled", Content: "Hello {{.name}}", Enabled: false}
	mgr.Create(ctx, tmpl)

	_, err := mgr.Render(ctx, tmpl.ID, map[string]interface{}{"name": "World"}, nil)
	assert.NoError(t, err)
}

func TestTemplateManager_Render_WithPrefixSuffix(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	mgr := NewTemplateManager(storage, true)
	ctx := context.Background()

	tmpl := &PromptTemplate{Name: "Wrap", Content: "Hello {{.name}}", Enabled: true}
	mgr.Create(ctx, tmpl)

	opts := &RenderOptions{
		EscapeFunc: func(s string) string { return "[START]" + s + "[END]" },
	}
	rendered, err := mgr.Render(ctx, tmpl.ID, map[string]interface{}{"name": "World"}, opts)
	require.NoError(t, err)
	assert.Contains(t, rendered, "Hello World")
}

func TestMemoryTemplateStorage_LoadAll(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	ctx := context.Background()

	storage.Save(ctx, &PromptTemplate{ID: "t1", Name: "T1", Content: "Hello"})
	storage.Save(ctx, &PromptTemplate{ID: "t2", Name: "T2", Content: "World"})

	templates, err := storage.LoadAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(templates))
}

func TestMemoryTemplateStorage_LoadVersions(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	ctx := context.Background()

	storage.SaveVersion(ctx, &TemplateVersion{TemplateID: "t1", Version: "1.0.0"})

	versions, err := storage.LoadVersions(ctx, "t1")
	require.NoError(t, err)
	assert.Equal(t, 1, len(versions))
}

func TestMemoryTemplateStorage_LoadVersions_NotFound(t *testing.T) {
	storage := NewMemoryTemplateStorage()
	ctx := context.Background()

	_, err := storage.LoadVersions(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestBaseAdapter_DoRequest_ConnectionError(t *testing.T) {
	cfg := &AdapterConfig{
		Provider:      ProviderOpenAI,
		BaseURL:       "http://localhost:99999",
		Timeout:       1 * time.Second,
		MaxRetries:    1,
		RetryInterval: 10 * time.Millisecond,
	}
	adapter := NewBaseAdapter(cfg)

	_, err := adapter.doRequest(context.Background(), http.MethodGet, "/test", nil)
	assert.Error(t, err)
}

func TestBaseAdapter_DoRequest_429Retry(t *testing.T) {
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
	assert.Equal(t, 2, callCount)
}

func TestBaseAdapter_DoRequest_503Retry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusServiceUnavailable)
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
	assert.Equal(t, 2, callCount)
}

func TestBaseAdapter_DoRequest_400NoRetry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	cfg := &AdapterConfig{
		Provider:      ProviderOpenAI,
		BaseURL:       server.URL,
		Timeout:       5 * time.Second,
		MaxRetries:    3,
		RetryInterval: 10 * time.Millisecond,
	}
	adapter := NewBaseAdapter(cfg)

	resp, err := adapter.doRequest(context.Background(), http.MethodGet, "/test", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, 1, callCount)
	resp.Body.Close()
}

func TestOpenAIAdapter_Chat_WithConversationID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIChatResponse{
			ID:      "chat-1",
			Model:   "gpt-4",
			Created: time.Now().Unix(),
		}
		resp.Choices = append(resp.Choices, struct {
			Index        int                    `json:"index"`
			Message      map[string]interface{} `json:"message"`
			FinishReason string                 `json:"finish_reason"`
		}{Index: 0, Message: map[string]interface{}{"role": "assistant", "content": "Hi"}, FinishReason: "stop"})
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
		ConversationID: "conv-1",
		Messages:       []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "conv-1", resp.ConversationID)
}

func TestOpenAIAdapter_Chat_WithExtra(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openAIChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, 0.8, req.Temperature)
		assert.Equal(t, 500, req.MaxTokens)

		resp := openAIChatResponse{
			ID:      "chat-1",
			Model:   "gpt-4",
			Created: time.Now().Unix(),
		}
		resp.Choices = append(resp.Choices, struct {
			Index        int                    `json:"index"`
			Message      map[string]interface{} `json:"message"`
			FinishReason string                 `json:"finish_reason"`
		}{Index: 0, Message: map[string]interface{}{"role": "assistant", "content": "Hi"}, FinishReason: "stop"})
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
		Messages:    []Message{{Role: "user", Content: "Hi"}},
		Temperature: 0.8,
		MaxTokens:   500,
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestClaudeAdapter_Chat_WithSystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "You are helpful", req["system"])

		resp := claudeResponse{
			ID:   "msg-1",
			Role: "assistant",
		}
		resp.Content = append(resp.Content, struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{Type: "text", Text: "Hi"})
		resp.Model = "claude-3-opus-20240229"
		resp.StopReason = "end_turn"
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

	_, err := adapter.Chat(context.Background(), &ChatRequest{
		SystemPrompt: "You are helpful",
		Messages:     []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
}

func TestClaudeAdapter_Chat_WithModelParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "claude-3-opus-20240229", req["model"])

		resp := claudeResponse{
			ID:   "msg-1",
			Role: "assistant",
		}
		resp.Content = append(resp.Content, struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{Type: "text", Text: "Hi"})
		resp.Model = "claude-3-opus-20240229"
		resp.StopReason = "end_turn"
		json.NewEncoder(w).Encode(resp)
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

	_, err := adapter.Chat(context.Background(), &ChatRequest{
		Model:    "claude-3-opus-20240229",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
}

func TestProcessStreamResponse_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		flusher.Flush()
		time.Sleep(5 * time.Second)
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

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := adapter.StreamChat(ctx, &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	assert.Error(t, err)
}

func TestProcessClaudeStream_WithEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher, _ := w.(http.Flusher)
		fmt.Fprintf(w, `{"type":"content_block_delta","delta":{"text":"Hello"}}`+"\n")
		flusher.Flush()
		fmt.Fprintf(w, `{"type":"message_stop"}`+"\n")
		flusher.Flush()
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

	chunkChan, err := adapter.StreamChat(context.Background(), &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)

	var chunks []ChatStreamChunk
	timeout := time.After(2 * time.Second)
	for {
		select {
		case chunk, ok := <-chunkChan:
			if !ok {
				goto done
			}
			chunks = append(chunks, chunk)
			if chunk.FinishReason == "stop" {
				goto done
			}
		case <-timeout:
			goto done
		}
	}
done:
	assert.GreaterOrEqual(t, len(chunks), 1)
}

func TestProcessLocalStream_WithChunks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		flusher, _ := w.(http.Flusher)
		result := struct {
			ID      string `json:"id"`
			Model   string `json:"model"`
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}{
			ID:    "local-1",
			Model: "local-llm",
			Choices: []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			}{{Delta: struct {
				Content string `json:"content"`
			}{Content: "Hello"}, FinishReason: "stop"}},
		}
		json.NewEncoder(w).Encode(result)
		flusher.Flush()
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

	chunkChan, err := adapter.StreamChat(context.Background(), &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)

	var chunks []ChatStreamChunk
	timeout := time.After(2 * time.Second)
	for {
		select {
		case chunk, ok := <-chunkChan:
			if !ok {
				goto done2
			}
			chunks = append(chunks, chunk)
			if chunk.FinishReason == "stop" {
				goto done2
			}
		case <-timeout:
			goto done2
		}
	}
done2:
	assert.GreaterOrEqual(t, len(chunks), 1)
}

func TestLocalModelAdapter_Completion_NoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := struct {
			ID      string             `json:"id"`
			Model   string             `json:"model"`
			Choices []CompletionChoice `json:"choices"`
			Usage   TokenUsage         `json:"usage"`
		}{
			ID:    "comp-empty",
			Model: "local-llm",
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
	assert.Equal(t, 0, len(resp.Choices))
}

func TestLoadBalancer_SelectByStrategy_InvalidStrategy(t *testing.T) {
	lb := NewLoadBalancer("invalid")
	selector := NewModelSelector("gpt-4")
	cfg := &AdapterConfig{Provider: ProviderOpenAI, BaseURL: "http://localhost"}
	selector.Register("gpt-4", NewOpenAIAdapter(cfg))
	lb.RegisterSelector("openai", selector)

	selected, err := lb.SelectByStrategy()
	require.NoError(t, err)
	assert.NotNil(t, selected)
}

func TestOpenAIAdapter_Chat_WithStop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openAIChatRequest
		json.NewDecoder(r.Body).Decode(&req)

		resp := openAIChatResponse{
			ID:      "chat-1",
			Model:   "gpt-4",
			Created: time.Now().Unix(),
		}
		resp.Choices = append(resp.Choices, struct {
			Index        int                    `json:"index"`
			Message      map[string]interface{} `json:"message"`
			FinishReason string                 `json:"finish_reason"`
		}{Index: 0, Message: map[string]interface{}{"role": "assistant", "content": "Hi"}, FinishReason: "stop"})
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
		Stop:     []string{"\n"},
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestOpenAIAdapter_Chat_WithExtraFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIChatResponse{
			ID:      "chat-1",
			Model:   "gpt-4",
			Created: time.Now().Unix(),
		}
		resp.Choices = append(resp.Choices, struct {
			Index        int                    `json:"index"`
			Message      map[string]interface{} `json:"message"`
			FinishReason string                 `json:"finish_reason"`
		}{Index: 0, Message: map[string]interface{}{"role": "assistant", "content": "Hi"}, FinishReason: "length"})
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
		Messages:         []Message{{Role: "user", Content: "Hi"}},
		FrequencyPenalty: 0.5,
		PresencePenalty:  0.3,
		User:             "test-user",
	})
	require.NoError(t, err)
	assert.Equal(t, "length", resp.FinishReason)
}

func TestProcessStreamResponse_WithSSEData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher, _ := w.(http.Flusher)
		writer := bufio.NewWriter(w)
		writer.WriteString("data: {\"id\":\"chat-1\",\"model\":\"gpt-4\",\"choices\":[{\"delta\":{\"content\":\"Hi\"},\"finish_reason\":\"\"}]}\n\n")
		writer.Flush()
		flusher.Flush()
		time.Sleep(50 * time.Millisecond)
		writer.WriteString("data: [DONE]\n\n")
		writer.Flush()
		flusher.Flush()
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

	chunkChan, err := adapter.StreamChat(context.Background(), &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)

	timeout := time.After(3 * time.Second)
	for {
		select {
		case _, ok := <-chunkChan:
			if !ok {
				return
			}
		case <-timeout:
			return
		}
	}
}

func TestProcessLocalStream_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		flusher.Flush()
		time.Sleep(5 * time.Second)
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

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := adapter.StreamChat(ctx, &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	assert.Error(t, err)
}

func TestProcessClaudeStream_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		flusher.Flush()
		time.Sleep(5 * time.Second)
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

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := adapter.StreamChat(ctx, &ChatRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	assert.Error(t, err)
}

func TestOpenAIAdapter_Chat_WithModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openAIChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "gpt-4-turbo", req.Model)

		resp := openAIChatResponse{
			ID:      "chat-1",
			Model:   "gpt-4-turbo",
			Created: time.Now().Unix(),
		}
		resp.Choices = append(resp.Choices, struct {
			Index        int                    `json:"index"`
			Message      map[string]interface{} `json:"message"`
			FinishReason string                 `json:"finish_reason"`
		}{Index: 0, Message: map[string]interface{}{"role": "assistant", "content": "Hi"}, FinishReason: "stop"})
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

	resp, err := adapter.Chat(context.Background(), &ChatRequest{
		Model:    "gpt-4-turbo",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "gpt-4-turbo", resp.Model)
}

func TestClaudeAdapter_Chat_WithTemperature(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		assert.InDelta(t, 0.5, req["temperature"], 0.01)

		resp := claudeResponse{
			ID:   "msg-1",
			Role: "assistant",
		}
		resp.Content = append(resp.Content, struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{Type: "text", Text: "Hi"})
		resp.Model = "claude-3-opus-20240229"
		resp.StopReason = "end_turn"
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

	_, err := adapter.Chat(context.Background(), &ChatRequest{
		Messages:    []Message{{Role: "user", Content: "Hi"}},
		Temperature: 0.5,
	})
	require.NoError(t, err)
}

func TestOpenAIAdapter_Embedding_ArrayInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		inputs, ok := req["input"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 2, len(inputs))

		resp := openAIEmbeddingResponse{Object: "list", Model: "text-embedding-ada-002"}
		resp.Data = append(resp.Data, struct {
			Object    string    `json:"object"`
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		}{Object: "embedding", Index: 0, Embedding: []float64{0.1}})
		resp.Data = append(resp.Data, struct {
			Object    string    `json:"object"`
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		}{Object: "embedding", Index: 1, Embedding: []float64{0.2}})
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
		Input: []string{"hello", "world"},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, len(resp.Data))
}

func TestOpenAIAdapter_Completion_ArrayPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAICompletionResponse{
			ID:      "comp-1",
			Model:   "gpt-4",
			Created: time.Now().Unix(),
		}
		resp.Choices = append(resp.Choices, struct {
			Text         string    `json:"text"`
			Index        int       `json:"index"`
			Logprobs     *struct{} `json:"logprobs"`
			FinishReason string    `json:"finish_reason"`
		}{Text: "Result", Index: 0, FinishReason: "stop"})
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
		Prompt: []string{"First", "Second"},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(resp.Choices))
}

func TestModelAdapter_Interface(t *testing.T) {
	var _ ModelAdapter = NewOpenAIAdapter(&AdapterConfig{Provider: ProviderOpenAI})
	var _ ModelAdapter = NewClaudeAdapter(&AdapterConfig{Provider: ProviderAnthropic})
	var _ ModelAdapter = NewLocalModelAdapter(&AdapterConfig{Provider: ProviderLocal})
}

func TestOpenAIAdapter_Chat_NilMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIChatResponse{
			ID:      "chat-1",
			Model:   "gpt-4",
			Created: time.Now().Unix(),
		}
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
		Messages: []Message{},
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestClaudeAdapter_Chat_NilMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := claudeResponse{
			ID:   "msg-1",
			Role: "assistant",
		}
		resp.Content = append(resp.Content, struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{Type: "text", Text: "Hi"})
		resp.Model = "claude-3-opus-20240229"
		resp.StopReason = "end_turn"
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
		Messages: []Message{},
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestLocalModelAdapter_Chat_WithModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "custom-llm", req["model"])

		result := struct {
			ID      string `json:"id"`
			Model   string `json:"model"`
			Choices []struct {
				Message      Message `json:"message"`
				FinishReason string  `json:"finish_reason"`
			} `json:"choices"`
			Usage TokenUsage `json:"usage"`
		}{
			ID:    "local-1",
			Model: "custom-llm",
			Choices: []struct {
				Message      Message `json:"message"`
				FinishReason string  `json:"finish_reason"`
			}{{Message: Message{Role: "assistant", Content: "Hi"}, FinishReason: "stop"}},
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

	resp, err := adapter.Chat(context.Background(), &ChatRequest{
		Model:    "custom-llm",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "custom-llm", resp.Model)
}

func TestOpenAIAdapter_Chat_WithTopP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openAIChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.InDelta(t, 0.9, req.TopP, 0.01)

		resp := openAIChatResponse{ID: "chat-1", Model: "gpt-4", Created: time.Now().Unix()}
		resp.Choices = append(resp.Choices, struct {
			Index        int                    `json:"index"`
			Message      map[string]interface{} `json:"message"`
			FinishReason string                 `json:"finish_reason"`
		}{Index: 0, Message: map[string]interface{}{"role": "assistant", "content": "Hi"}, FinishReason: "stop"})
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
		Messages: []Message{{Role: "user", Content: "Hi"}},
		TopP:     0.9,
	})
	require.NoError(t, err)
}

func TestContextManager_Create_WithConfig(t *testing.T) {
	cfg := &ContextConfig{
		Persist:           true,
		DefaultTTL:        2 * time.Hour,
		MaxMessages:       50,
		MaxTokens:         4096,
		EnableCompression: false,
	}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	conv, err := mgr.Create(ctx, "user1", "sess1", "System")
	require.NoError(t, err)
	assert.Equal(t, "user1", conv.UserID)
	assert.Equal(t, "System", conv.SystemPrompt)
}

func TestContextManager_Get_NotFound(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	_, err := mgr.Get(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestContextManager_GetMessages_NotFound(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	_, err := mgr.GetMessages(ctx, "nonexistent", 10, 0)
	assert.Error(t, err)
}

func TestContextManager_UpdateSystemPrompt_NotFound(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	err := mgr.UpdateSystemPrompt(ctx, "nonexistent", "new prompt")
	assert.Error(t, err)
}

func TestContextManager_SetModel_NotFound(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	err := mgr.SetModel(ctx, "nonexistent", "gpt-4", nil)
	assert.Error(t, err)
}

func TestContextManager_SetMetadata_NotFound(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	err := mgr.SetMetadata(ctx, "nonexistent", map[string]interface{}{})
	assert.Error(t, err)
}

func TestContextManager_ExtendTTL_NotFound_WithPersist(t *testing.T) {
	cfg := &ContextConfig{Persist: true, DefaultTTL: time.Hour}
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(cfg, storage)
	ctx := context.Background()

	err := mgr.ExtendTTL(ctx, "nonexistent", time.Hour)
	assert.Error(t, err)
}

func TestContextManager_ListByUser_Empty(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	result, err := mgr.ListByUser(ctx, "user1", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, len(result))
}

func TestContextManager_BuildChatRequest_WithMaxTokens(t *testing.T) {
	storage := NewMemoryContextStorage()
	mgr := NewContextManager(nil, storage)
	ctx := context.Background()

	conv, _ := mgr.Create(ctx, "user1", "sess1", "You are helpful")
	mgr.SetModel(ctx, conv.ConversationID, "gpt-4", map[string]interface{}{
		"max_tokens": 1000,
	})

	req, err := mgr.BuildChatRequest(ctx, conv.ConversationID, "test")
	require.NoError(t, err)
	assert.Equal(t, 1000, req.MaxTokens)
}

func TestSummaryCompressor_CreateSummary_LongContent(t *testing.T) {
	compressor := NewSummaryCompressor()
	longContent := strings.Repeat("a", 300)
	messages := []ContextMessage{
		{Role: "user", Content: longContent},
	}
	summary := compressor.createSummary(messages)
	assert.True(t, len(summary) < 300+len("[user] "))
}

func TestModelType_Constants(t *testing.T) {
	assert.Equal(t, ModelType("chat"), ModelTypeChat)
	assert.Equal(t, ModelType("completion"), ModelTypeCompletion)
	assert.Equal(t, ModelType("embedding"), ModelTypeEmbedding)
}

func TestLoadBalancerStrategies(t *testing.T) {
	assert.Equal(t, LoadBalanceStrategy("round_robin"), StrategyRoundRobin)
	assert.Equal(t, LoadBalanceStrategy("random"), StrategyRandom)
	assert.Equal(t, LoadBalanceStrategy("least_connection"), StrategyLeastConn)
}
