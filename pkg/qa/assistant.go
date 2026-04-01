package qa

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	aiqa "github.com/new-energy-monitoring/pkg/ai/qa"
)

// AssistantInterface AI助手接口
type AssistantInterface interface {
	Ask(ctx context.Context, req *AskRequest) (*AskResponse, error)
	StartSession(ctx context.Context, userID string) (string, error)
	EndSession(sessionID string) error
	GenerateTitle(firstMessage string) string
}

// Assistant AI助手
type Assistant struct {
	dialogueManager *aiqa.DialogueManager
	intentRecognizer *aiqa.IntentRecognizer
	config          *AssistantConfig
	mu              sync.RWMutex
}

// 确保Assistant实现了AssistantInterface接口
var _ AssistantInterface = (*Assistant)(nil)

// AssistantConfig 助手配置
type AssistantConfig struct {
	MaxContextMessages int           // 最大上下文消息数
	ResponseTimeout    time.Duration // 响应超时时间
	EnableMultiTurn    bool          // 启用多轮对话
	SystemPrompt       string        // 系统提示词
}

// DefaultAssistantConfig 默认助手配置
func DefaultAssistantConfig() *AssistantConfig {
	return &AssistantConfig{
		MaxContextMessages: 10,
		ResponseTimeout:    30 * time.Second,
		EnableMultiTurn:    true,
		SystemPrompt: `你是一个专业的能源监控系统AI助手。你的职责是：
1. 帮助用户查询设备状态、历史数据和统计信息
2. 协助用户进行设备控制和参数配置
3. 提供故障诊断和性能分析建议
4. 解答关于能源监控系统的各种问题

请用专业、简洁、友好的语言回答用户问题。`,
	}
}

// NewAssistant 创建AI助手
func NewAssistant(config *AssistantConfig) *Assistant {
	if config == nil {
		config = DefaultAssistantConfig()
	}

	intentRecognizer := aiqa.NewIntentRecognizer(nil)
	dialogueManager := aiqa.NewDialogueManager(intentRecognizer, nil)

	return &Assistant{
		dialogueManager:  dialogueManager,
		intentRecognizer: intentRecognizer,
		config:           config,
	}
}

// AskRequest 提问请求
type AskRequest struct {
	SessionID string
	UserID    string
	Question  string
	Context   *ConversationContext
}

// AskResponse 提问响应
type AskResponse struct {
	SessionID   string
	Answer      string
	Confidence  float64
	Intent      *IntentInfo
	Suggestions []string
	RequiresMore bool
}

// IntentInfo 意图信息
type IntentInfo struct {
	Type       string
	Name       string
	Confidence float64
	Slots      map[string]interface{}
}

// ConversationContext 对话上下文
type ConversationContext struct {
	Messages []*ContextMessage
	Variables map[string]interface{}
}

// ContextMessage 上下文消息
type ContextMessage struct {
	Role      string
	Content   string
	Timestamp time.Time
}

// Ask 提问
func (a *Assistant) Ask(ctx context.Context, req *AskRequest) (*AskResponse, error) {
	if req.Question == "" {
		return nil, fmt.Errorf("question cannot be empty")
	}

	// 获取或创建会话
	sessionID := req.SessionID
	if sessionID == "" {
		session, err := a.dialogueManager.StartSession(ctx, req.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to start session: %w", err)
		}
		sessionID = session.SessionID
	}

	// 处理对话
	response, err := a.dialogueManager.Process(ctx, sessionID, req.Question)
	if err != nil {
		return nil, fmt.Errorf("failed to process dialogue: %w", err)
	}

	// 获取会话信息
	session, err := a.dialogueManager.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 构建响应
	resp := &AskResponse{
		SessionID:   sessionID,
		Answer:      response.Content,
		Confidence:  response.Confidence,
		Suggestions: response.Suggestions,
		RequiresMore: response.RequiresMore,
	}

	// 添加意图信息
	if session.CurrentIntent != nil {
		resp.Intent = &IntentInfo{
			Type:       string(session.CurrentIntent.Type),
			Name:       session.CurrentIntent.Name,
			Confidence: session.CurrentIntent.Confidence,
			Slots:      make(map[string]interface{}),
		}
		for name, slot := range session.CurrentIntent.Slots {
			if slot.Filled {
				resp.Intent.Slots[name] = slot.Value
			}
		}
	}

	return resp, nil
}

// StartSession 开始新会话
func (a *Assistant) StartSession(ctx context.Context, userID string) (string, error) {
	session, err := a.dialogueManager.StartSession(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to start session: %w", err)
	}
	return session.SessionID, nil
}

// EndSession 结束会话
func (a *Assistant) EndSession(sessionID string) error {
	return a.dialogueManager.EndSession(sessionID)
}

// GetSessionHistory 获取会话历史
func (a *Assistant) GetSessionHistory(sessionID string) ([]*ContextMessage, error) {
	turns, err := a.dialogueManager.GetDialogueHistory(sessionID)
	if err != nil {
		return nil, err
	}

	messages := make([]*ContextMessage, 0, len(turns))
	for _, turn := range turns {
		messages = append(messages, &ContextMessage{
			Role:      turn.Role,
			Content:   turn.Content,
			Timestamp: turn.Timestamp,
		})
	}

	return messages, nil
}

// SetContextVariable 设置上下文变量
func (a *Assistant) SetContextVariable(sessionID, key string, value interface{}) error {
	return a.dialogueManager.SetContextVariable(sessionID, key, value)
}

// GetContextVariable 获取上下文变量
func (a *Assistant) GetContextVariable(sessionID, key string) (interface{}, error) {
	return a.dialogueManager.GetContextVariable(sessionID, key)
}

// CleanExpiredSessions 清理过期会话
func (a *Assistant) CleanExpiredSessions() int {
	return a.dialogueManager.CleanExpiredSessions()
}

// GetActiveSessions 获取活跃会话数
func (a *Assistant) GetActiveSessions() int {
	return a.dialogueManager.GetActiveSessions()
}

// GenerateTitle 根据对话内容生成会话标题
func (a *Assistant) GenerateTitle(firstMessage string) string {
	// 简单实现：截取前50个字符作为标题
	title := strings.TrimSpace(firstMessage)
	if len(title) > 50 {
		title = title[:50] + "..."
	}
	return title
}

// BuildPrompt 构建提示词
func (a *Assistant) BuildPrompt(sessionID string, question string) (string, error) {
	session, err := a.dialogueManager.GetSession(sessionID)
	if err != nil {
		return "", err
	}

	var promptBuilder strings.Builder

	// 添加系统提示词
	promptBuilder.WriteString(a.config.SystemPrompt)
	promptBuilder.WriteString("\n\n")

	// 添加对话历史
	if len(session.Turns) > 0 {
		startIdx := 0
		if len(session.Turns) > a.config.MaxContextMessages {
			startIdx = len(session.Turns) - a.config.MaxContextMessages
		}

		for i := startIdx; i < len(session.Turns); i++ {
			turn := session.Turns[i]
			switch turn.Role {
			case "user":
				promptBuilder.WriteString(fmt.Sprintf("用户: %s\n", turn.Content))
			case "system":
				promptBuilder.WriteString(fmt.Sprintf("助手: %s\n", turn.Content))
			}
		}
	}

	// 添加当前问题
	promptBuilder.WriteString(fmt.Sprintf("用户: %s\n", question))
	promptBuilder.WriteString("助手: ")

	return promptBuilder.String(), nil
}

// RecognizeIntent 识别意图
func (a *Assistant) RecognizeIntent(ctx context.Context, text string) (*IntentInfo, error) {
	intent, err := a.intentRecognizer.Recognize(ctx, text)
	if err != nil {
		return nil, err
	}

	info := &IntentInfo{
		Type:       string(intent.Type),
		Name:       intent.Name,
		Confidence: intent.Confidence,
		Slots:      make(map[string]interface{}),
	}

	for name, slot := range intent.Slots {
		if slot.Filled {
			info.Slots[name] = slot.Value
		}
	}

	return info, nil
}

// AddIntentPattern 添加意图模式
func (a *Assistant) AddIntentPattern(pattern *aiqa.IntentPattern) error {
	return a.intentRecognizer.AddPattern(pattern)
}

// AddEntityRule 添加实体识别规则
func (a *Assistant) AddEntityRule(entityType aiqa.EntityType, rule *aiqa.EntityRule) error {
	return a.intentRecognizer.AddEntityRule(entityType, rule)
}

// AddDialoguePolicy 添加对话策略
func (a *Assistant) AddDialoguePolicy(policy *aiqa.DialoguePolicy) error {
	return a.dialogueManager.AddPolicy(policy)
}
