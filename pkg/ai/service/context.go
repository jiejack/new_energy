package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ConversationContext 对话上下文
type ConversationContext struct {
	// 对话ID
	ConversationID string `json:"conversation_id" redis:"conversation_id"`
	
	// 用户ID
	UserID string `json:"user_id" redis:"user_id"`
	
	// 会话ID
	SessionID string `json:"session_id" redis:"session_id"`
	
	// 消息历史
	Messages []ContextMessage `json:"messages" redis:"messages"`
	
	// 系统提示
	SystemPrompt string `json:"system_prompt" redis:"system_prompt"`
	
	// 使用的模型
	Model string `json:"model" redis:"model"`
	
	// 模型参数
	ModelParams map[string]interface{} `json:"model_params" redis:"model_params"`
	
	// 元数据
	Metadata map[string]interface{} `json:"metadata" redis:"metadata"`
	
	// 创建时间
	CreatedAt time.Time `json:"created_at" redis:"created_at"`
	
	// 更新时间
	UpdatedAt time.Time `json:"updated_at" redis:"updated_at"`
	
	// 过期时间
	ExpiresAt time.Time `json:"expires_at" redis:"expires_at"`
	
	// Token统计
	TokenStats TokenStats `json:"token_stats" redis:"token_stats"`
	
	// 状态
	Status string `json:"status" redis:"status"`
}

// ContextMessage 上下文消息
type ContextMessage struct {
	// 消息ID
	ID string `json:"id" redis:"id"`
	
	// 角色 (system, user, assistant)
	Role string `json:"role" redis:"role"`
	
	// 内容
	Content string `json:"content" redis:"content"`
	
	// Token数量
	TokenCount int `json:"token_count" redis:"token_count"`
	
	// 时间戳
	Timestamp time.Time `json:"timestamp" redis:"timestamp"`
	
	// 元数据
	Metadata map[string]interface{} `json:"metadata" redis:"metadata"`
}

// TokenStats Token统计
type TokenStats struct {
	// 总输入Token
	TotalInputTokens int `json:"total_input_tokens" redis:"total_input_tokens"`
	
	// 总输出Token
	TotalOutputTokens int `json:"total_output_tokens" redis:"total_output_tokens"`
	
	// 消息数量
	MessageCount int `json:"message_count" redis:"message_count"`
	
	// 平均消息长度
	AvgMessageLength int `json:"avg_message_length" redis:"avg_message_length"`
}

// ContextConfig 上下文配置
type ContextConfig struct {
	// 最大消息数量
	MaxMessages int `json:"max_messages" yaml:"max_messages"`
	
	// 最大Token数量
	MaxTokens int `json:"max_tokens" yaml:"max_tokens"`
	
	// 默认过期时间
	DefaultTTL time.Duration `json:"default_ttl" yaml:"default_ttl"`
	
	// 是否启用压缩
	EnableCompression bool `json:"enable_compression" yaml:"enable_compression"`
	
	// 压缩阈值（消息数量）
	CompressionThreshold int `json:"compression_threshold" yaml:"compression_threshold"`
	
	// 保留最近消息数量（压缩时）
	KeepRecentMessages int `json:"keep_recent_messages" yaml:"keep_recent_messages"`
	
	// 是否持久化
	Persist bool `json:"persist" yaml:"persist"`
}

// DefaultContextConfig 默认上下文配置
func DefaultContextConfig() *ContextConfig {
	return &ContextConfig{
		MaxMessages:          100,
		MaxTokens:            8192,
		DefaultTTL:           24 * time.Hour,
		EnableCompression:    true,
		CompressionThreshold: 50,
		KeepRecentMessages:   10,
		Persist:              true,
	}
}

// ContextManager 上下文管理器
type ContextManager struct {
	config    *ContextConfig
	storage   ContextStorage
	cache     map[string]*ConversationContext
	mu        sync.RWMutex
	compressor ContextCompressor
}

// ContextStorage 上下文存储接口
type ContextStorage interface {
	// Save 保存上下文
	Save(ctx context.Context, context *ConversationContext) error
	
	// Load 加载上下文
	Load(ctx context.Context, conversationID string) (*ConversationContext, error)
	
	// Delete 删除上下文
	Delete(ctx context.Context, conversationID string) error
	
	// Exists 检查上下文是否存在
	Exists(ctx context.Context, conversationID string) (bool, error)
	
	// SetTTL 设置过期时间
	SetTTL(ctx context.Context, conversationID string, ttl time.Duration) error
	
	// ListByUser 列出用户的对话
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*ConversationContext, error)
	
	// CleanExpired 清理过期上下文
	CleanExpired(ctx context.Context) (int64, error)
}

// ContextCompressor 上下文压缩器接口
type ContextCompressor interface {
	// Compress 压缩上下文
	Compress(ctx context.Context, context *ConversationContext) (*ConversationContext, error)
}

// NewContextManager 创建上下文管理器
func NewContextManager(config *ContextConfig, storage ContextStorage) *ContextManager {
	if config == nil {
		config = DefaultContextConfig()
	}
	
	return &ContextManager{
		config:    config,
		storage:   storage,
		cache:     make(map[string]*ConversationContext),
		compressor: NewSummaryCompressor(),
	}
}

// Create 创建新的对话上下文
func (m *ContextManager) Create(ctx context.Context, userID, sessionID, systemPrompt string) (*ConversationContext, error) {
	conversationID := generateConversationID()
	
	now := time.Now()
	context := &ConversationContext{
		ConversationID: conversationID,
		UserID:         userID,
		SessionID:      sessionID,
		Messages:       make([]ContextMessage, 0),
		SystemPrompt:   systemPrompt,
		ModelParams:    make(map[string]interface{}),
		Metadata:       make(map[string]interface{}),
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpiresAt:      now.Add(m.config.DefaultTTL),
		Status:         "active",
	}
	
	m.mu.Lock()
	m.cache[conversationID] = context
	m.mu.Unlock()
	
	if m.config.Persist {
		if err := m.storage.Save(ctx, context); err != nil {
			return nil, fmt.Errorf("save context: %w", err)
		}
	}
	
	return context, nil
}

// Get 获取对话上下文
func (m *ContextManager) Get(ctx context.Context, conversationID string) (*ConversationContext, error) {
	// 先从缓存获取
	m.mu.RLock()
	context, ok := m.cache[conversationID]
	m.mu.RUnlock()
	
	if ok {
		// 检查是否过期
		if time.Now().After(context.ExpiresAt) {
			m.mu.Lock()
			delete(m.cache, conversationID)
			m.mu.Unlock()
			return nil, ErrContextExpired
		}
		return context, nil
	}
	
	// 从存储加载
	context, err := m.storage.Load(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("load context: %w", err)
	}
	
	// 检查是否过期
	if time.Now().After(context.ExpiresAt) {
		_ = m.storage.Delete(ctx, conversationID)
		return nil, ErrContextExpired
	}
	
	// 加入缓存
	m.mu.Lock()
	m.cache[conversationID] = context
	m.mu.Unlock()
	
	return context, nil
}

// AddMessage 添加消息
func (m *ContextManager) AddMessage(ctx context.Context, conversationID string, message *ContextMessage) error {
	context, err := m.Get(ctx, conversationID)
	if err != nil {
		return err
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 设置消息ID和时间戳
	if message.ID == "" {
		message.ID = generateMessageID()
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	
	// 添加消息
	context.Messages = append(context.Messages, *message)
	context.UpdatedAt = time.Now()
	
	// 更新Token统计
	context.TokenStats.MessageCount++
	if message.TokenCount > 0 {
		if message.Role == "user" || message.Role == "system" {
			context.TokenStats.TotalInputTokens += message.TokenCount
		} else {
			context.TokenStats.TotalOutputTokens += message.TokenCount
		}
	}
	
	// 检查是否需要压缩
	if m.config.EnableCompression && len(context.Messages) >= m.config.CompressionThreshold {
		compressed, err := m.compressor.Compress(ctx, context)
		if err != nil {
			// 记录错误但继续
			fmt.Printf("warning: failed to compress context: %v\n", err)
		} else {
			context = compressed
		}
	}
	
	// 检查消息数量限制
	if len(context.Messages) > m.config.MaxMessages {
		// 移除最旧的消息，保留系统消息
		var newMessages []ContextMessage
		for _, msg := range context.Messages {
			if msg.Role == "system" {
				newMessages = append(newMessages, msg)
			}
		}
		// 添加最近的非系统消息
		nonSystemMessages := make([]ContextMessage, 0)
		for _, msg := range context.Messages {
			if msg.Role != "system" {
				nonSystemMessages = append(nonSystemMessages, msg)
			}
		}
		start := len(nonSystemMessages) - (m.config.MaxMessages - len(newMessages))
		if start > 0 {
			nonSystemMessages = nonSystemMessages[start:]
		}
		newMessages = append(newMessages, nonSystemMessages...)
		context.Messages = newMessages
	}
	
	// 更新缓存
	m.cache[conversationID] = context
	
	// 持久化
	if m.config.Persist {
		if err := m.storage.Save(ctx, context); err != nil {
			return fmt.Errorf("save context: %w", err)
		}
	}
	
	return nil
}

// GetMessages 获取消息列表
func (m *ContextManager) GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]ContextMessage, error) {
	context, err := m.Get(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	
	messages := context.Messages
	if offset >= len(messages) {
		return []ContextMessage{}, nil
	}
	
	end := offset + limit
	if end > len(messages) {
		end = len(messages)
	}
	
	return messages[offset:end], nil
}

// UpdateSystemPrompt 更新系统提示
func (m *ContextManager) UpdateSystemPrompt(ctx context.Context, conversationID, systemPrompt string) error {
	context, err := m.Get(ctx, conversationID)
	if err != nil {
		return err
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	context.SystemPrompt = systemPrompt
	context.UpdatedAt = time.Now()
	
	m.cache[conversationID] = context
	
	if m.config.Persist {
		if err := m.storage.Save(ctx, context); err != nil {
			return fmt.Errorf("save context: %w", err)
		}
	}
	
	return nil
}

// SetModel 设置模型
func (m *ContextManager) SetModel(ctx context.Context, conversationID, model string, params map[string]interface{}) error {
	context, err := m.Get(ctx, conversationID)
	if err != nil {
		return err
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	context.Model = model
	if params != nil {
		context.ModelParams = params
	}
	context.UpdatedAt = time.Now()
	
	m.cache[conversationID] = context
	
	if m.config.Persist {
		if err := m.storage.Save(ctx, context); err != nil {
			return fmt.Errorf("save context: %w", err)
		}
	}
	
	return nil
}

// SetMetadata 设置元数据
func (m *ContextManager) SetMetadata(ctx context.Context, conversationID string, metadata map[string]interface{}) error {
	context, err := m.Get(ctx, conversationID)
	if err != nil {
		return err
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for k, v := range metadata {
		context.Metadata[k] = v
	}
	context.UpdatedAt = time.Now()
	
	m.cache[conversationID] = context
	
	if m.config.Persist {
		if err := m.storage.Save(ctx, context); err != nil {
			return fmt.Errorf("save context: %w", err)
		}
	}
	
	return nil
}

// Delete 删除对话上下文
func (m *ContextManager) Delete(ctx context.Context, conversationID string) error {
	m.mu.Lock()
	delete(m.cache, conversationID)
	m.mu.Unlock()
	
	if m.config.Persist {
		if err := m.storage.Delete(ctx, conversationID); err != nil {
			return fmt.Errorf("delete context: %w", err)
		}
	}
	
	return nil
}

// ExtendTTL 延长过期时间
func (m *ContextManager) ExtendTTL(ctx context.Context, conversationID string, ttl time.Duration) error {
	context, err := m.Get(ctx, conversationID)
	if err != nil {
		return err
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	context.ExpiresAt = time.Now().Add(ttl)
	context.UpdatedAt = time.Now()
	
	m.cache[conversationID] = context
	
	if m.config.Persist {
		if err := m.storage.SetTTL(ctx, conversationID, ttl); err != nil {
			return fmt.Errorf("set ttl: %w", err)
		}
		if err := m.storage.Save(ctx, context); err != nil {
			return fmt.Errorf("save context: %w", err)
		}
	}
	
	return nil
}

// ListByUser 列出用户的对话
func (m *ContextManager) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*ConversationContext, error) {
	return m.storage.ListByUser(ctx, userID, limit, offset)
}

// Clear 清空对话消息
func (m *ContextManager) Clear(ctx context.Context, conversationID string) error {
	context, err := m.Get(ctx, conversationID)
	if err != nil {
		return err
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 保留系统消息
	var systemMessages []ContextMessage
	for _, msg := range context.Messages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, msg)
		}
	}
	
	context.Messages = systemMessages
	context.TokenStats = TokenStats{}
	context.UpdatedAt = time.Now()
	
	m.cache[conversationID] = context
	
	if m.config.Persist {
		if err := m.storage.Save(ctx, context); err != nil {
			return fmt.Errorf("save context: %w", err)
		}
	}
	
	return nil
}

// Compress 压缩上下文
func (m *ContextManager) Compress(ctx context.Context, conversationID string) error {
	context, err := m.Get(ctx, conversationID)
	if err != nil {
		return err
	}
	
	compressed, err := m.compressor.Compress(ctx, context)
	if err != nil {
		return fmt.Errorf("compress context: %w", err)
	}
	
	m.mu.Lock()
	m.cache[conversationID] = compressed
	m.mu.Unlock()
	
	if m.config.Persist {
		if err := m.storage.Save(ctx, compressed); err != nil {
			return fmt.Errorf("save context: %w", err)
		}
	}
	
	return nil
}

// CleanExpired 清理过期上下文
func (m *ContextManager) CleanExpired(ctx context.Context) (int64, error) {
	// 清理缓存
	m.mu.Lock()
	for id, context := range m.cache {
		if time.Now().After(context.ExpiresAt) {
			delete(m.cache, id)
		}
	}
	m.mu.Unlock()
	
	// 清理存储
	return m.storage.CleanExpired(ctx)
}

// GetStats 获取统计信息
func (m *ContextManager) GetStats(ctx context.Context, conversationID string) (*TokenStats, error) {
	context, err := m.Get(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	
	return &context.TokenStats, nil
}

// BuildChatRequest 构建聊天请求
func (m *ContextManager) BuildChatRequest(ctx context.Context, conversationID string, userMessage string) (*ChatRequest, error) {
	context, err := m.Get(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	
	// 构建消息列表
	messages := make([]Message, 0, len(context.Messages)+2)
	
	// 添加系统提示
	if context.SystemPrompt != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: context.SystemPrompt,
		})
	}
	
	// 添加历史消息
	for _, msg := range context.Messages {
		messages = append(messages, Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	
	// 添加用户消息
	messages = append(messages, Message{
		Role:    "user",
		Content: userMessage,
	})
	
	// 构建请求
	req := &ChatRequest{
		ConversationID: conversationID,
		Messages:       messages,
		Model:          context.Model,
	}
	
	// 设置模型参数
	if temp, ok := context.ModelParams["temperature"].(float64); ok {
		req.Temperature = temp
	}
	if maxTokens, ok := context.ModelParams["max_tokens"].(int); ok {
		req.MaxTokens = maxTokens
	}
	if topP, ok := context.ModelParams["top_p"].(float64); ok {
		req.TopP = topP
	}
	
	return req, nil
}

// 错误定义
var (
	ErrContextNotFound = errors.New("context not found")
	ErrContextExpired  = errors.New("context expired")
	ErrContextFull     = errors.New("context is full")
)

// SummaryCompressor 摘要压缩器
type SummaryCompressor struct {
	keepRecent int
}

// NewSummaryCompressor 创建摘要压缩器
func NewSummaryCompressor() *SummaryCompressor {
	return &SummaryCompressor{
		keepRecent: 10,
	}
}

// Compress 压缩上下文
func (c *SummaryCompressor) Compress(ctx context.Context, context *ConversationContext) (*ConversationContext, error) {
	if len(context.Messages) <= c.keepRecent {
		return context, nil
	}
	
	// 分离系统消息和普通消息
	var systemMessages []ContextMessage
	var otherMessages []ContextMessage
	
	for _, msg := range context.Messages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, msg)
		} else {
			otherMessages = append(otherMessages, msg)
		}
	}
	
	// 保留最近的N条消息
	recentStart := len(otherMessages) - c.keepRecent
	if recentStart < 0 {
		recentStart = 0
	}
	
	// 创建摘要消息
	var summaryContent string
	if recentStart > 0 {
		summaryContent = c.createSummary(otherMessages[:recentStart])
	}
	
	// 构建新的消息列表
	var newMessages []ContextMessage
	newMessages = append(newMessages, systemMessages...)
	
	if summaryContent != "" {
		newMessages = append(newMessages, ContextMessage{
			ID:        generateMessageID(),
			Role:      "system",
			Content:   fmt.Sprintf("[Previous conversation summary]\n%s", summaryContent),
			Timestamp: time.Now(),
		})
	}
	
	newMessages = append(newMessages, otherMessages[recentStart:]...)
	
	context.Messages = newMessages
	context.UpdatedAt = time.Now()
	
	return context, nil
}

// createSummary 创建摘要
func (c *SummaryCompressor) createSummary(messages []ContextMessage) string {
	var summary strings.Builder
	
	for _, msg := range messages {
		role := msg.Role
		content := msg.Content
		
		// 截断过长的内容
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		
		summary.WriteString(fmt.Sprintf("[%s]: %s\n", role, content))
	}
	
	return summary.String()
}

// RedisContextStorage Redis上下文存储
type RedisContextStorage struct {
	client    RedisClient
	keyPrefix string
}

// RedisClient Redis客户端接口
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	RPush(ctx context.Context, key string, values ...interface{}) error
}

// NewRedisContextStorage 创建Redis存储
func NewRedisContextStorage(client RedisClient, keyPrefix string) *RedisContextStorage {
	if keyPrefix == "" {
		keyPrefix = "ai:context:"
	}
	return &RedisContextStorage{
		client:    client,
		keyPrefix: keyPrefix,
	}
}

// Save 保存上下文
func (s *RedisContextStorage) Save(ctx context.Context, context *ConversationContext) error {
	key := s.keyPrefix + context.ConversationID
	
	data, err := json.Marshal(context)
	if err != nil {
		return fmt.Errorf("marshal context: %w", err)
	}
	
	ttl := time.Until(context.ExpiresAt)
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	
	return s.client.Set(ctx, key, data, ttl)
}

// Load 加载上下文
func (s *RedisContextStorage) Load(ctx context.Context, conversationID string) (*ConversationContext, error) {
	key := s.keyPrefix + conversationID
	
	data, err := s.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get context: %w", err)
	}
	
	var context ConversationContext
	if err := json.Unmarshal([]byte(data), &context); err != nil {
		return nil, fmt.Errorf("unmarshal context: %w", err)
	}
	
	return &context, nil
}

// Delete 删除上下文
func (s *RedisContextStorage) Delete(ctx context.Context, conversationID string) error {
	key := s.keyPrefix + conversationID
	return s.client.Del(ctx, key)
}

// Exists 检查上下文是否存在
func (s *RedisContextStorage) Exists(ctx context.Context, conversationID string) (bool, error) {
	key := s.keyPrefix + conversationID
	count, err := s.client.Exists(ctx, key)
	return count > 0, err
}

// SetTTL 设置过期时间
func (s *RedisContextStorage) SetTTL(ctx context.Context, conversationID string, ttl time.Duration) error {
	key := s.keyPrefix + conversationID
	return s.client.Expire(ctx, key, ttl)
}

// ListByUser 列出用户的对话
func (s *RedisContextStorage) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*ConversationContext, error) {
	// 需要额外的索引支持，这里简化实现
	// 实际实现中应该维护一个用户到对话列表的索引
	return nil, fmt.Errorf("not implemented")
}

// CleanExpired 清理过期上下文
func (s *RedisContextStorage) CleanExpired(ctx context.Context) (int64, error) {
	// Redis会自动清理过期键
	return 0, nil
}

// MemoryContextStorage 内存上下文存储
type MemoryContextStorage struct {
	contexts map[string]*ConversationContext
	userIndex map[string][]string // userID -> conversationIDs
	mu        sync.RWMutex
}

// NewMemoryContextStorage 创建内存存储
func NewMemoryContextStorage() *MemoryContextStorage {
	return &MemoryContextStorage{
		contexts:  make(map[string]*ConversationContext),
		userIndex: make(map[string][]string),
	}
}

// Save 保存上下文
func (s *MemoryContextStorage) Save(ctx context.Context, context *ConversationContext) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.contexts[context.ConversationID] = context
	
	// 更新用户索引
	if context.UserID != "" {
		found := false
		for _, id := range s.userIndex[context.UserID] {
			if id == context.ConversationID {
				found = true
				break
			}
		}
		if !found {
			s.userIndex[context.UserID] = append(s.userIndex[context.UserID], context.ConversationID)
		}
	}
	
	return nil
}

// Load 加载上下文
func (s *MemoryContextStorage) Load(ctx context.Context, conversationID string) (*ConversationContext, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	context, ok := s.contexts[conversationID]
	if !ok {
		return nil, ErrContextNotFound
	}
	
	return context, nil
}

// Delete 删除上下文
func (s *MemoryContextStorage) Delete(ctx context.Context, conversationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	context, ok := s.contexts[conversationID]
	if ok && context.UserID != "" {
		// 从用户索引中移除
		var newUserContexts []string
		for _, id := range s.userIndex[context.UserID] {
			if id != conversationID {
				newUserContexts = append(newUserContexts, id)
			}
		}
		s.userIndex[context.UserID] = newUserContexts
	}
	
	delete(s.contexts, conversationID)
	return nil
}

// Exists 检查上下文是否存在
func (s *MemoryContextStorage) Exists(ctx context.Context, conversationID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	_, ok := s.contexts[conversationID]
	return ok, nil
}

// SetTTL 设置过期时间
func (s *MemoryContextStorage) SetTTL(ctx context.Context, conversationID string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	context, ok := s.contexts[conversationID]
	if !ok {
		return ErrContextNotFound
	}
	
	context.ExpiresAt = time.Now().Add(ttl)
	return nil
}

// ListByUser 列出用户的对话
func (s *MemoryContextStorage) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*ConversationContext, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	ids := s.userIndex[userID]
	if offset >= len(ids) {
		return []*ConversationContext{}, nil
	}
	
	end := offset + limit
	if end > len(ids) {
		end = len(ids)
	}
	
	var result []*ConversationContext
	for i := offset; i < end; i++ {
		if context, ok := s.contexts[ids[i]]; ok {
			// 检查是否过期
			if time.Now().Before(context.ExpiresAt) {
				result = append(result, context)
			}
		}
	}
	
	return result, nil
}

// CleanExpired 清理过期上下文
func (s *MemoryContextStorage) CleanExpired(ctx context.Context) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	var count int64
	now := time.Now()
	
	for id, context := range s.contexts {
		if now.After(context.ExpiresAt) {
			delete(s.contexts, id)
			count++
		}
	}
	
	return count, nil
}

// 辅助函数
func generateConversationID() string {
	return fmt.Sprintf("conv_%d", time.Now().UnixNano())
}

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// 上下文状态
const (
	ContextStatusActive    = "active"
	ContextStatusArchived  = "archived"
	ContextStatusCompleted = "completed"
)
