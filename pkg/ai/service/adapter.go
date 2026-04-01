package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// ModelAdapter 模型适配器接口
type ModelAdapter interface {
	// Chat 聊天对话
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	
	// Embedding 获取嵌入向量
	Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)
	
	// Completion 文本补全
	Completion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	
	// StreamChat 流式聊天
	StreamChat(ctx context.Context, req *ChatRequest) (<-chan ChatStreamChunk, error)
	
	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error
	
	// GetModelInfo 获取模型信息
	GetModelInfo() *ModelInfo
	
	// Close 关闭适配器
	Close() error
}

// AdapterConfig 适配器配置
type AdapterConfig struct {
	// 提供商类型
	Provider ProviderType `json:"provider" yaml:"provider"`
	
	// API密钥
	APIKey string `json:"api_key" yaml:"api_key"`
	
	// API基础URL
	BaseURL string `json:"base_url" yaml:"base_url"`
	
	// 默认模型
	DefaultModel string `json:"default_model" yaml:"default_model"`
	
	// 组织ID
	Organization string `json:"organization" yaml:"organization"`
	
	// 请求超时
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
	
	// 最大重试次数
	MaxRetries int `json:"max_retries" yaml:"max_retries"`
	
	// 重试间隔
	RetryInterval time.Duration `json:"retry_interval" yaml:"retry_interval"`
	
	// 代理URL
	ProxyURL string `json:"proxy_url" yaml:"proxy_url"`
	
	// TLS配置
	InsecureSkipVerify bool `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
	
	// 模型信息
	ModelInfo *ModelInfo `json:"model_info" yaml:"model_info"`
}

// BaseAdapter 基础适配器
type BaseAdapter struct {
	config     *AdapterConfig
	httpClient *http.Client
	modelInfo  *ModelInfo
	requestID  int64
}

// NewBaseAdapter 创建基础适配器
func NewBaseAdapter(config *AdapterConfig) *BaseAdapter {
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryInterval == 0 {
		config.RetryInterval = 1 * time.Second
	}
	
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	
	return &BaseAdapter{
		config: config,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   config.Timeout,
		},
		modelInfo: config.ModelInfo,
	}
}

// generateRequestID 生成请求ID
func (a *BaseAdapter) generateRequestID() string {
	id := atomic.AddInt64(&a.requestID, 1)
	return fmt.Sprintf("%s-%d", uuid.New().String()[:8], id)
}

// doRequest 执行HTTP请求
func (a *BaseAdapter) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}
	
	url := a.config.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	if a.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.APIKey)
	}
	if a.config.Organization != "" {
		req.Header.Set("OpenAI-Organization", a.config.Organization)
	}
	
	var lastErr error
	for i := 0; i <= a.config.MaxRetries; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(a.config.RetryInterval * time.Duration(i)):
			}
		}
		
		resp, err := a.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}
		
		return resp, nil
	}
	
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// Close 关闭适配器
func (a *BaseAdapter) Close() error {
	a.httpClient.CloseIdleConnections()
	return nil
}

// GetModelInfo 获取模型信息
func (a *BaseAdapter) GetModelInfo() *ModelInfo {
	return a.modelInfo
}

// OpenAIAdapter OpenAI适配器
type OpenAIAdapter struct {
	*BaseAdapter
}

// NewOpenAIAdapter 创建OpenAI适配器
func NewOpenAIAdapter(config *AdapterConfig) *OpenAIAdapter {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.DefaultModel == "" {
		config.DefaultModel = "gpt-4"
	}
	if config.ModelInfo == nil {
		config.ModelInfo = &ModelInfo{
			ID:              config.DefaultModel,
			Name:            config.DefaultModel,
			Type:            string(ModelTypeChat),
			Provider:        string(ProviderOpenAI),
			ContextWindow:   8192,
			MaxOutputTokens: 4096,
			Available:       true,
			Capabilities:    []string{"chat", "function_calling", "streaming"},
		}
	}
	
	return &OpenAIAdapter{
		BaseAdapter: NewBaseAdapter(config),
	}
}

// Chat 聊天对话
func (a *OpenAIAdapter) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = a.config.DefaultModel
	}
	
	openaiReq := a.buildOpenAIChatRequest(req)
	
	resp, err := a.doRequest(ctx, http.MethodPost, "/chat/completions", openaiReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, a.parseErrorResponse(resp)
	}
	
	var openaiResp openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	
	return a.convertChatResponse(&openaiResp, req.ConversationID), nil
}

// Embedding 获取嵌入向量
func (a *OpenAIAdapter) Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = "text-embedding-ada-002"
	}
	
	openaiReq := map[string]interface{}{
		"input":          req.Input,
		"model":          req.Model,
		"encoding_format": req.EncodingFormat,
	}
	
	resp, err := a.doRequest(ctx, http.MethodPost, "/embeddings", openaiReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, a.parseErrorResponse(resp)
	}
	
	var openaiResp openAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	
	return a.convertEmbeddingResponse(&openaiResp), nil
}

// Completion 文本补全
func (a *OpenAIAdapter) Completion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	if req.Model == "" {
		req.Model = a.config.DefaultModel
	}
	
	openaiReq := a.buildOpenAICompletionRequest(req)
	
	resp, err := a.doRequest(ctx, http.MethodPost, "/completions", openaiReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, a.parseErrorResponse(resp)
	}
	
	var openaiResp openAICompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	
	return a.convertCompletionResponse(&openaiResp), nil
}

// StreamChat 流式聊天
func (a *OpenAIAdapter) StreamChat(ctx context.Context, req *ChatRequest) (<-chan ChatStreamChunk, error) {
	if req.Model == "" {
		req.Model = a.config.DefaultModel
	}
	req.Stream = true
	
	openaiReq := a.buildOpenAIChatRequest(req)
	
	resp, err := a.doRequest(ctx, http.MethodPost, "/chat/completions", openaiReq)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, a.parseErrorResponse(resp)
	}
	
	chunkChan := make(chan ChatStreamChunk, 100)
	go a.processStreamResponse(ctx, resp, chunkChan, req.ConversationID)
	
	return chunkChan, nil
}

// HealthCheck 健康检查
func (a *OpenAIAdapter) HealthCheck(ctx context.Context) error {
	resp, err := a.doRequest(ctx, http.MethodGet, "/models", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}
	
	return nil
}

// OpenAI请求/响应结构体
type openAIChatRequest struct {
	Model            string                   `json:"model"`
	Messages         []map[string]interface{} `json:"messages"`
	Temperature      float64                  `json:"temperature,omitempty"`
	TopP             float64                  `json:"top_p,omitempty"`
	N                int                      `json:"n,omitempty"`
	Stream           bool                     `json:"stream,omitempty"`
	Stop             interface{}              `json:"stop,omitempty"`
	MaxTokens        int                      `json:"max_tokens,omitempty"`
	PresencePenalty  float64                  `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64                  `json:"frequency_penalty,omitempty"`
	User             string                   `json:"user,omitempty"`
}

type openAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int                    `json:"index"`
		Message      map[string]interface{} `json:"message"`
		FinishReason string                 `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

type openAICompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string    `json:"text"`
		Index        int       `json:"index"`
		Logprobs     *struct{} `json:"logprobs"`
		FinishReason string    `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param"`
		Code    string `json:"code"`
	} `json:"error"`
}

func (a *OpenAIAdapter) buildOpenAIChatRequest(req *ChatRequest) *openAIChatRequest {
	messages := make([]map[string]interface{}, 0, len(req.Messages)+1)
	
	if req.SystemPrompt != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": req.SystemPrompt,
		})
	}
	
	for _, msg := range req.Messages {
		messages = append(messages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	
	return &openAIChatRequest{
		Model:            req.Model,
		Messages:         messages,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		Stream:           req.Stream,
		Stop:             req.Stop,
		MaxTokens:        req.MaxTokens,
		PresencePenalty:  req.PresencePenalty,
		FrequencyPenalty: req.FrequencyPenalty,
		User:             req.User,
	}
}

func (a *OpenAIAdapter) buildOpenAICompletionRequest(req *CompletionRequest) map[string]interface{} {
	return map[string]interface{}{
		"model":             req.Model,
		"prompt":            req.Prompt,
		"max_tokens":        req.MaxTokens,
		"temperature":       req.Temperature,
		"top_p":             req.TopP,
		"n":                 req.N,
		"stream":            false,
		"stop":              req.Stop,
		"echo":              req.Echo,
		"suffix":            req.Suffix,
		"frequency_penalty": req.FrequencyPenalty,
		"presence_penalty":  req.PresencePenalty,
		"best_of":           req.BestOf,
		"user":              req.User,
	}
}

func (a *OpenAIAdapter) convertChatResponse(resp *openAIChatResponse, conversationID string) *ChatResponse {
	if len(resp.Choices) == 0 {
		return &ChatResponse{
			ID:             resp.ID,
			ConversationID: conversationID,
			Model:          resp.Model,
			CreatedAt:      time.Unix(resp.Created, 0),
		}
	}
	
	choice := resp.Choices[0]
	content, _ := choice.Message["content"].(string)
	role, _ := choice.Message["role"].(string)
	
	return &ChatResponse{
		ID:             resp.ID,
		ConversationID: conversationID,
		Message: Message{
			Role:    role,
			Content: content,
		},
		Model: resp.Model,
		Usage: TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		CreatedAt:    time.Unix(resp.Created, 0),
		FinishReason: choice.FinishReason,
	}
}

func (a *OpenAIAdapter) convertEmbeddingResponse(resp *openAIEmbeddingResponse) *EmbeddingResponse {
	data := make([]EmbeddingData, len(resp.Data))
	for i, d := range resp.Data {
		data[i] = EmbeddingData{
			Embedding: d.Embedding,
			Index:     d.Index,
			Object:    d.Object,
		}
	}
	
	return &EmbeddingResponse{
		ID:        uuid.New().String(),
		Data:      data,
		Model:     resp.Model,
		CreatedAt: time.Now(),
		Usage: TokenUsage{
			PromptTokens: resp.Usage.PromptTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
	}
}

func (a *OpenAIAdapter) convertCompletionResponse(resp *openAICompletionResponse) *CompletionResponse {
	choices := make([]CompletionChoice, len(resp.Choices))
	for i, c := range resp.Choices {
		choices[i] = CompletionChoice{
			Text:         c.Text,
			Index:        c.Index,
			FinishReason: c.FinishReason,
		}
	}
	
	return &CompletionResponse{
		ID:       resp.ID,
		Choices:  choices,
		Model:    resp.Model,
		CreatedAt: time.Unix(resp.Created, 0),
		Usage: TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}
}

func (a *OpenAIAdapter) parseErrorResponse(resp *http.Response) error {
	var errResp openAIErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("http error: %d", resp.StatusCode)
	}
	
	return &AIError{
		Type:       errResp.Error.Type,
		Message:    errResp.Error.Message,
		Code:       errResp.Error.Code,
		Param:      errResp.Error.Param,
		StatusCode: resp.StatusCode,
	}
}

func (a *OpenAIAdapter) processStreamResponse(ctx context.Context, resp *http.Response, chunkChan chan<- ChatStreamChunk, conversationID string) {
	defer close(chunkChan)
	defer resp.Body.Close()
	
	decoder := json.NewDecoder(resp.Body)
	requestID := a.generateRequestID()
	
	for {
		select {
		case <-ctx.Done():
			chunkChan <- ChatStreamChunk{Error: ctx.Err()}
			return
		default:
		}
		
		line, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return
			}
			chunkChan <- ChatStreamChunk{Error: err}
			return
		}
		
		if line == nil {
			continue
		}
		
		// 处理SSE格式的流式响应
		// 这里简化处理，实际需要解析SSE格式
		_ = line
		_ = requestID
		_ = conversationID
	}
}

// ClaudeAdapter Claude适配器
type ClaudeAdapter struct {
	*BaseAdapter
}

// NewClaudeAdapter 创建Claude适配器
func NewClaudeAdapter(config *AdapterConfig) *ClaudeAdapter {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com/v1"
	}
	if config.DefaultModel == "" {
		config.DefaultModel = "claude-3-opus-20240229"
	}
	if config.ModelInfo == nil {
		config.ModelInfo = &ModelInfo{
			ID:              config.DefaultModel,
			Name:            config.DefaultModel,
			Type:            string(ModelTypeChat),
			Provider:        string(ProviderAnthropic),
			ContextWindow:   200000,
			MaxOutputTokens: 4096,
			Available:       true,
			Capabilities:    []string{"chat", "streaming", "vision"},
		}
	}
	
	return &ClaudeAdapter{
		BaseAdapter: NewBaseAdapter(config),
	}
}

// Chat 聊天对话
func (a *ClaudeAdapter) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = a.config.DefaultModel
	}
	
	claudeReq := a.buildClaudeRequest(req)
	
	httpReq, err := a.buildClaudeHTTPRequest(ctx, claudeReq)
	if err != nil {
		return nil, err
	}
	
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, a.parseClaudeErrorResponse(resp)
	}
	
	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	
	return a.convertClaudeResponse(&claudeResp, req.ConversationID), nil
}

// Embedding 获取嵌入向量 (Claude暂不支持)
func (a *ClaudeAdapter) Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	return nil, fmt.Errorf("claude does not support embedding API")
}

// Completion 文本补全
func (a *ClaudeAdapter) Completion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	chatReq := &ChatRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		User:        req.User,
	}
	
	switch v := req.Prompt.(type) {
	case string:
		chatReq.Messages = []Message{{Role: "user", Content: v}}
	case []string:
		for _, p := range v {
			chatReq.Messages = append(chatReq.Messages, Message{Role: "user", Content: p})
		}
	}
	
	chatResp, err := a.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}
	
	return &CompletionResponse{
		ID:        chatResp.ID,
		Model:     chatResp.Model,
		CreatedAt: chatResp.CreatedAt,
		Usage:     chatResp.Usage,
		Choices: []CompletionChoice{
			{
				Text:         chatResp.Message.Content,
				Index:        0,
				FinishReason: chatResp.FinishReason,
			},
		},
	}, nil
}

// StreamChat 流式聊天
func (a *ClaudeAdapter) StreamChat(ctx context.Context, req *ChatRequest) (<-chan ChatStreamChunk, error) {
	if req.Model == "" {
		req.Model = a.config.DefaultModel
	}
	
	claudeReq := a.buildClaudeRequest(req)
	claudeReq["stream"] = true
	
	httpReq, err := a.buildClaudeHTTPRequest(ctx, claudeReq)
	if err != nil {
		return nil, err
	}
	
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, a.parseClaudeErrorResponse(resp)
	}
	
	chunkChan := make(chan ChatStreamChunk, 100)
	go a.processClaudeStream(ctx, resp, chunkChan, req.ConversationID)
	
	return chunkChan, nil
}

// HealthCheck 健康检查
func (a *ClaudeAdapter) HealthCheck(ctx context.Context) error {
	req := &ChatRequest{
		Model:    a.config.DefaultModel,
		Messages: []Message{{Role: "user", Content: "ping"}},
		MaxTokens: 5,
	}
	
	_, err := a.Chat(ctx, req)
	return err
}

// Claude请求/响应结构体
type claudeRequest struct {
	Model       string                   `json:"model"`
	Messages    []map[string]interface{} `json:"messages"`
	MaxTokens   int                      `json:"max_tokens"`
	System      string                   `json:"system,omitempty"`
	Temperature float64                  `json:"temperature,omitempty"`
	TopP        float64                  `json:"top_p,omitempty"`
	TopK        int                      `json:"top_k,omitempty"`
	Stream      bool                     `json:"stream,omitempty"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
}

type claudeResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Role         string `json:"role"`
	Content      []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type claudeErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func (a *ClaudeAdapter) buildClaudeRequest(req *ChatRequest) map[string]interface{} {
	messages := make([]map[string]interface{}, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}
	
	return map[string]interface{}{
		"model":       req.Model,
		"messages":    messages,
		"max_tokens":  req.MaxTokens,
		"system":      req.SystemPrompt,
		"temperature": req.Temperature,
		"top_p":       req.TopP,
	}
}

func (a *ClaudeAdapter) buildClaudeHTTPRequest(ctx context.Context, body interface{}) (*http.Request, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.config.BaseURL+"/messages", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	
	return req, nil
}

func (a *ClaudeAdapter) convertClaudeResponse(resp *claudeResponse, conversationID string) *ChatResponse {
	content := ""
	if len(resp.Content) > 0 {
		content = resp.Content[0].Text
	}
	
	return &ChatResponse{
		ID:             resp.ID,
		ConversationID: conversationID,
		Message: Message{
			Role:    resp.Role,
			Content: content,
		},
		Model: resp.Model,
		Usage: TokenUsage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
		CreatedAt:    time.Now(),
		FinishReason: resp.StopReason,
	}
}

func (a *ClaudeAdapter) parseClaudeErrorResponse(resp *http.Response) error {
	var errResp claudeErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("http error: %d", resp.StatusCode)
	}
	
	return &AIError{
		Type:       errResp.Error.Type,
		Message:    errResp.Error.Message,
		StatusCode: resp.StatusCode,
	}
}

func (a *ClaudeAdapter) processClaudeStream(ctx context.Context, resp *http.Response, chunkChan chan<- ChatStreamChunk, conversationID string) {
	defer close(chunkChan)
	defer resp.Body.Close()
	
	decoder := json.NewDecoder(resp.Body)
	requestID := a.generateRequestID()
	
	for {
		select {
		case <-ctx.Done():
			chunkChan <- ChatStreamChunk{Error: ctx.Err()}
			return
		default:
		}
		
		var event map[string]interface{}
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				return
			}
			chunkChan <- ChatStreamChunk{Error: err}
			return
		}
		
		eventType, _ := event["type"].(string)
		if eventType == "content_block_delta" {
			delta, _ := event["delta"].(map[string]interface{})
			text, _ := delta["text"].(string)
			
			chunkChan <- ChatStreamChunk{
				ID:             requestID,
				ConversationID: conversationID,
				Delta:          Message{Role: "assistant", Content: text},
				Model:          a.config.DefaultModel,
			}
		} else if eventType == "message_stop" {
			chunkChan <- ChatStreamChunk{
				ID:             requestID,
				ConversationID: conversationID,
				FinishReason:   "stop",
				Model:          a.config.DefaultModel,
			}
			return
		}
	}
}

// LocalModelAdapter 本地模型适配器
type LocalModelAdapter struct {
	*BaseAdapter
}

// NewLocalModelAdapter 创建本地模型适配器
func NewLocalModelAdapter(config *AdapterConfig) *LocalModelAdapter {
	if config.DefaultModel == "" {
		config.DefaultModel = "local-llm"
	}
	if config.ModelInfo == nil {
		config.ModelInfo = &ModelInfo{
			ID:              config.DefaultModel,
			Name:            config.DefaultModel,
			Type:            string(ModelTypeChat),
			Provider:        string(ProviderLocal),
			ContextWindow:   4096,
			MaxOutputTokens: 2048,
			Available:       true,
			Capabilities:    []string{"chat", "completion"},
		}
	}
	
	return &LocalModelAdapter{
		BaseAdapter: NewBaseAdapter(config),
	}
}

// Chat 聊天对话
func (a *LocalModelAdapter) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = a.config.DefaultModel
	}
	
	localReq := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
		"top_p":       req.TopP,
	}
	
	resp, err := a.doRequest(ctx, http.MethodPost, "/v1/chat/completions", localReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("local model error: %d", resp.StatusCode)
	}
	
	var result struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Choices []struct {
			Message      Message `json:"message"`
			FinishReason string  `json:"finish_reason"`
		} `json:"choices"`
		Usage TokenUsage `json:"usage"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from local model")
	}
	
	return &ChatResponse{
		ID:             result.ID,
		ConversationID: req.ConversationID,
		Message:        result.Choices[0].Message,
		Model:          result.Model,
		Usage:          result.Usage,
		CreatedAt:      time.Now(),
		FinishReason:   result.Choices[0].FinishReason,
	}, nil
}

// Embedding 获取嵌入向量
func (a *LocalModelAdapter) Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = "local-embedding"
	}
	
	localReq := map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	}
	
	resp, err := a.doRequest(ctx, http.MethodPost, "/v1/embeddings", localReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("local model error: %d", resp.StatusCode)
	}
	
	var result struct {
		Data  []EmbeddingData `json:"data"`
		Model string          `json:"model"`
		Usage TokenUsage      `json:"usage"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	
	return &EmbeddingResponse{
		ID:        uuid.New().String(),
		Data:      result.Data,
		Model:     result.Model,
		Usage:     result.Usage,
		CreatedAt: time.Now(),
	}, nil
}

// Completion 文本补全
func (a *LocalModelAdapter) Completion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	if req.Model == "" {
		req.Model = a.config.DefaultModel
	}
	
	localReq := map[string]interface{}{
		"model":       req.Model,
		"prompt":      req.Prompt,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"top_p":       req.TopP,
	}
	
	resp, err := a.doRequest(ctx, http.MethodPost, "/v1/completions", localReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("local model error: %d", resp.StatusCode)
	}
	
	var result struct {
		ID      string             `json:"id"`
		Model   string             `json:"model"`
		Choices []CompletionChoice `json:"choices"`
		Usage   TokenUsage         `json:"usage"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	
	return &CompletionResponse{
		ID:        result.ID,
		Choices:   result.Choices,
		Model:     result.Model,
		Usage:     result.Usage,
		CreatedAt: time.Now(),
	}, nil
}

// StreamChat 流式聊天
func (a *LocalModelAdapter) StreamChat(ctx context.Context, req *ChatRequest) (<-chan ChatStreamChunk, error) {
	if req.Model == "" {
		req.Model = a.config.DefaultModel
	}
	req.Stream = true
	
	localReq := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
		"top_p":       req.TopP,
		"stream":      true,
	}
	
	resp, err := a.doRequest(ctx, http.MethodPost, "/v1/chat/completions", localReq)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, fmt.Errorf("local model error: %d", resp.StatusCode)
	}
	
	chunkChan := make(chan ChatStreamChunk, 100)
	go a.processLocalStream(ctx, resp, chunkChan, req.ConversationID)
	
	return chunkChan, nil
}

// HealthCheck 健康检查
func (a *LocalModelAdapter) HealthCheck(ctx context.Context) error {
	resp, err := a.doRequest(ctx, http.MethodGet, "/health", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}
	
	return nil
}

func (a *LocalModelAdapter) processLocalStream(ctx context.Context, resp *http.Response, chunkChan chan<- ChatStreamChunk, conversationID string) {
	defer close(chunkChan)
	defer resp.Body.Close()
	
	decoder := json.NewDecoder(resp.Body)
	requestID := a.generateRequestID()
	
	for {
		select {
		case <-ctx.Done():
			chunkChan <- ChatStreamChunk{Error: ctx.Err()}
			return
		default:
		}
		
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}
		
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				return
			}
			chunkChan <- ChatStreamChunk{Error: err}
			return
		}
		
		if len(chunk.Choices) > 0 {
			chunkChan <- ChatStreamChunk{
				ID:             requestID,
				ConversationID: conversationID,
				Delta: Message{
					Role:    "assistant",
					Content: chunk.Choices[0].Delta.Content,
				},
				Model:        a.config.DefaultModel,
				FinishReason: chunk.Choices[0].FinishReason,
			}
		}
	}
}

// ModelSelector 模型选择器
type ModelSelector struct {
	adapters map[string]ModelAdapter
	defaultModel string
	mu       sync.RWMutex
}

// NewModelSelector 创建模型选择器
func NewModelSelector(defaultModel string) *ModelSelector {
	return &ModelSelector{
		adapters:     make(map[string]ModelAdapter),
		defaultModel: defaultModel,
	}
}

// Register 注册适配器
func (s *ModelSelector) Register(name string, adapter ModelAdapter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.adapters[name] = adapter
}

// Select 选择适配器
func (s *ModelSelector) Select(model string) (ModelAdapter, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if model == "" {
		model = s.defaultModel
	}
	
	adapter, ok := s.adapters[model]
	if !ok {
		return nil, fmt.Errorf("model not found: %s", model)
	}
	
	return adapter, nil
}

// List 列出所有模型
func (s *ModelSelector) List() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	models := make([]string, 0, len(s.adapters))
	for name := range s.adapters {
		models = append(models, name)
	}
	return models
}

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	selectors map[string]*ModelSelector
	strategy  LoadBalanceStrategy
	mu        sync.RWMutex
}

// LoadBalanceStrategy 负载均衡策略
type LoadBalanceStrategy string

const (
	StrategyRoundRobin LoadBalanceStrategy = "round_robin"
	StrategyRandom     LoadBalanceStrategy = "random"
	StrategyLeastConn  LoadBalanceStrategy = "least_connection"
	StrategyWeighted   LoadBalanceStrategy = "weighted"
)

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(strategy LoadBalanceStrategy) *LoadBalancer {
	return &LoadBalancer{
		selectors: make(map[string]*ModelSelector),
		strategy:  strategy,
	}
}

// RegisterSelector 注册选择器
func (lb *LoadBalancer) RegisterSelector(provider string, selector *ModelSelector) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.selectors[provider] = selector
}

// Select 选择适配器
func (lb *LoadBalancer) Select(provider, model string) (ModelAdapter, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	selector, ok := lb.selectors[provider]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", provider)
	}
	
	return selector.Select(model)
}

// SelectByStrategy 根据策略选择
func (lb *LoadBalancer) SelectByStrategy() (ModelAdapter, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	if len(lb.selectors) == 0 {
		return nil, fmt.Errorf("no providers available")
	}
	
	providers := make([]string, 0, len(lb.selectors))
	for p := range lb.selectors {
		providers = append(providers, p)
	}
	
	var selectedProvider string
	switch lb.strategy {
	case StrategyRoundRobin:
		selectedProvider = providers[0]
	case StrategyRandom:
		selectedProvider = providers[rand.Intn(len(providers))]
	default:
		selectedProvider = providers[0]
	}
	
	selector := lb.selectors[selectedProvider]
	return selector.Select("")
}

// GetAllModels 获取所有模型信息
func (lb *LoadBalancer) GetAllModels() []*ModelInfo {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	var models []*ModelInfo
	for _, selector := range lb.selectors {
		selector.mu.RLock()
		for _, adapter := range selector.adapters {
			models = append(models, adapter.GetModelInfo())
		}
		selector.mu.RUnlock()
	}
	return models
}
