package service

import (
	"context"
	"time"
)

// AIService AI服务接口定义
type AIService interface {
	// Chat 聊天对话
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	
	// Embedding 获取文本嵌入向量
	Embedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)
	
	// Completion 文本补全
	Completion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	
	// StreamChat 流式聊天对话
	StreamChat(ctx context.Context, req *ChatRequest) (<-chan ChatStreamChunk, error)
	
	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error
	
	// GetModelName 获取当前模型名称
	GetModelName() string
	
	// Close 关闭服务
	Close() error
}

// Message 聊天消息
type Message struct {
	Role    string `json:"role"`     // system, user, assistant
	Content string `json:"content"`  // 消息内容
	Name    string `json:"name,omitempty"` // 可选的名称
}

// ChatRequest 聊天请求
type ChatRequest struct {
	// 对话ID，用于关联上下文
	ConversationID string `json:"conversation_id,omitempty"`
	
	// 消息列表
	Messages []Message `json:"messages"`
	
	// 模型名称
	Model string `json:"model,omitempty"`
	
	// 温度参数，控制随机性 (0-2)
	Temperature float64 `json:"temperature,omitempty"`
	
	// Top-P采样参数
	TopP float64 `json:"top_p,omitempty"`
	
	// 最大生成token数
	MaxTokens int `json:"max_tokens,omitempty"`
	
	// 停止词列表
	Stop []string `json:"stop,omitempty"`
	
	// 频率惩罚 (-2.0 to 2.0)
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
	
	// 存在惩罚 (-2.0 to 2.0)
	PresencePenalty float64 `json:"presence_penalty,omitempty"`
	
	// 用户标识
	User string `json:"user,omitempty"`
	
	// 系统提示词
	SystemPrompt string `json:"system_prompt,omitempty"`
	
	// 是否流式输出
	Stream bool `json:"stream,omitempty"`
	
	// 额外参数
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	// 响应ID
	ID string `json:"id"`
	
	// 对话ID
	ConversationID string `json:"conversation_id,omitempty"`
	
	// 生成的消息
	Message Message `json:"message"`
	
	// 使用的模型
	Model string `json:"model"`
	
	// Token使用统计
	Usage TokenUsage `json:"usage"`
	
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
	
	// 完成原因 (stop, length, content_filter)
	FinishReason string `json:"finish_reason,omitempty"`
	
	// 额外信息
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// ChatStreamChunk 流式聊天响应块
type ChatStreamChunk struct {
	// 响应ID
	ID string `json:"id"`
	
	// 对话ID
	ConversationID string `json:"conversation_id,omitempty"`
	
	// 增量内容
	Delta Message `json:"delta"`
	
	// 使用的模型
	Model string `json:"model"`
	
	// 完成原因
	FinishReason string `json:"finish_reason,omitempty"`
	
	// 错误信息
	Error error `json:"error,omitempty"`
}

// TokenUsage Token使用统计
type TokenUsage struct {
	// 输入token数
	PromptTokens int `json:"prompt_tokens"`
	
	// 输出token数
	CompletionTokens int `json:"completion_tokens"`
	
	// 总token数
	TotalTokens int `json:"total_tokens"`
}

// EmbeddingRequest 嵌入请求
type EmbeddingRequest struct {
	// 输入文本，可以是单个字符串或字符串数组
	Input interface{} `json:"input"` // string or []string
	
	// 模型名称
	Model string `json:"model,omitempty"`
	
	// 编码格式 (float, base64)
	EncodingFormat string `json:"encoding_format,omitempty"`
	
	// 用户标识
	User string `json:"user,omitempty"`
	
	// 额外参数
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// EmbeddingResponse 嵌入响应
type EmbeddingResponse struct {
	// 响应ID
	ID string `json:"id"`
	
	// 嵌入数据列表
	Data []EmbeddingData `json:"data"`
	
	// 使用的模型
	Model string `json:"model"`
	
	// Token使用统计
	Usage TokenUsage `json:"usage"`
	
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
}

// EmbeddingData 单个嵌入数据
type EmbeddingData struct {
	// 嵌入向量
	Embedding []float64 `json:"embedding"`
	
	// 索引
	Index int `json:"index"`
	
	// 对象类型
	Object string `json:"object"`
}

// CompletionRequest 补全请求
type CompletionRequest struct {
	// 提示文本
	Prompt interface{} `json:"prompt"` // string or []string
	
	// 模型名称
	Model string `json:"model,omitempty"`
	
	// 最大生成token数
	MaxTokens int `json:"max_tokens,omitempty"`
	
	// 温度参数
	Temperature float64 `json:"temperature,omitempty"`
	
	// Top-P采样参数
	TopP float64 `json:"top_p,omitempty"`
	
	// 生成数量
	N int `json:"n,omitempty"`
	
	// 停止词列表
	Stop []string `json:"stop,omitempty"`
	
	// 回显提示
	Echo bool `json:"echo,omitempty"`
	
	// 后缀文本
	Suffix string `json:"suffix,omitempty"`
	
	// 频率惩罚
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
	
	// 存在惩罚
	PresencePenalty float64 `json:"presence_penalty,omitempty"`
	
	// 最佳选择数量
	BestOf int `json:"best_of,omitempty"`
	
	// 用户标识
	User string `json:"user,omitempty"`
	
	// 额外参数
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// CompletionResponse 补全响应
type CompletionResponse struct {
	// 响应ID
	ID string `json:"id"`
	
	// 选择列表
	Choices []CompletionChoice `json:"choices"`
	
	// 使用的模型
	Model string `json:"model"`
	
	// Token使用统计
	Usage TokenUsage `json:"usage"`
	
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
}

// CompletionChoice 补全选择
type CompletionChoice struct {
	// 生成的文本
	Text string `json:"text"`
	
	// 索引
	Index int `json:"index"`
	
	// 日志概率
	Logprobs interface{} `json:"logprobs,omitempty"`
	
	// 完成原因
	FinishReason string `json:"finish_reason"`
}

// ServiceConfig AI服务配置
type ServiceConfig struct {
	// 默认模型
	DefaultModel string `json:"default_model" yaml:"default_model"`
	
	// API密钥
	APIKey string `json:"api_key" yaml:"api_key"`
	
	// API基础URL
	BaseURL string `json:"base_url" yaml:"base_url"`
	
	// 组织ID
	Organization string `json:"organization" yaml:"organization"`
	
	// 请求超时时间
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
	
	// 最大重试次数
	MaxRetries int `json:"max_retries" yaml:"max_retries"`
	
	// 重试间隔
	RetryInterval time.Duration `json:"retry_interval" yaml:"retry_interval"`
	
	// 默认温度
	DefaultTemperature float64 `json:"default_temperature" yaml:"default_temperature"`
	
	// 默认最大token数
	DefaultMaxTokens int `json:"default_max_tokens" yaml:"default_max_tokens"`
	
	// 是否启用缓存
	EnableCache bool `json:"enable_cache" yaml:"enable_cache"`
	
	// 缓存TTL
	CacheTTL time.Duration `json:"cache_ttl" yaml:"cache_ttl"`
	
	// 代理设置
	ProxyURL string `json:"proxy_url" yaml:"proxy_url"`
	
	// 额外配置
	Extra map[string]interface{} `json:"extra" yaml:"extra"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	// 模型ID
	ID string `json:"id"`
	
	// 模型名称
	Name string `json:"name"`
	
	// 模型类型 (chat, completion, embedding)
	Type string `json:"type"`
	
	// 提供商 (openai, anthropic, local)
	Provider string `json:"provider"`
	
	// 上下文窗口大小
	ContextWindow int `json:"context_window"`
	
	// 最大输出token数
	MaxOutputTokens int `json:"max_output_tokens"`
	
	// 输入价格 (每1k tokens)
	InputPrice float64 `json:"input_price"`
	
	// 输出价格 (每1k tokens)
	OutputPrice float64 `json:"output_price"`
	
	// 是否可用
	Available bool `json:"available"`
	
	// 支持的功能
	Capabilities []string `json:"capabilities"`
}

// ModelType 模型类型枚举
type ModelType string

const (
	ModelTypeChat       ModelType = "chat"
	ModelTypeCompletion ModelType = "completion"
	ModelTypeEmbedding  ModelType = "embedding"
)

// ProviderType 提供商类型枚举
type ProviderType string

const (
	ProviderOpenAI   ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderLocal    ProviderType = "local"
	ProviderAzure    ProviderType = "azure"
)

// AIError AI服务错误
type AIError struct {
	// 错误类型
	Type string `json:"type"`
	
	// 错误消息
	Message string `json:"message"`
	
	// 错误代码
	Code string `json:"code"`
	
	// HTTP状态码
	StatusCode int `json:"status_code"`
	
	// 参数
	Param string `json:"param,omitempty"`
}

func (e *AIError) Error() string {
	return e.Message
}

// NewAIError 创建AI错误
func NewAIError(errType, message, code string, statusCode int) *AIError {
	return &AIError{
		Type:       errType,
		Message:    message,
		Code:       code,
		StatusCode: statusCode,
	}
}

// 常见错误类型
var (
	ErrInvalidRequest     = NewAIError("invalid_request_error", "invalid request", "invalid_request", 400)
	ErrAuthentication     = NewAIError("authentication_error", "authentication failed", "auth_failed", 401)
	ErrPermissionDenied   = NewAIError("permission_denied", "permission denied", "forbidden", 403)
	ErrNotFound           = NewAIError("not_found_error", "resource not found", "not_found", 404)
	ErrRateLimitExceeded  = NewAIError("rate_limit_error", "rate limit exceeded", "rate_limit", 429)
	ErrServiceUnavailable = NewAIError("service_unavailable", "service unavailable", "unavailable", 503)
	ErrModelOverloaded    = NewAIError("model_overloaded", "model is overloaded", "overloaded", 529)
	ErrContextLength      = NewAIError("context_length_exceeded", "context length exceeded", "context_too_long", 400)
	ErrContentFilter      = NewAIError("content_filter", "content filtered", "content_filter", 400)
)
