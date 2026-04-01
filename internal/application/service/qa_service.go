package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/new-energy-monitoring/pkg/qa"
)

var (
	ErrSessionNotFound   = errors.New("session not found")
	ErrSessionDeleted    = errors.New("session has been deleted")
	ErrInvalidQuestion   = errors.New("invalid question")
	ErrUnauthorizedAccess = errors.New("unauthorized access to session")
)

// QAService 问答服务
type QAService struct {
	qaRepo    repository.QARepository
	assistant qa.AssistantInterface
}

// NewQAService 创建问答服务
func NewQAService(qaRepo repository.QARepository, assistant qa.AssistantInterface) *QAService {
	return &QAService{
		qaRepo:    qaRepo,
		assistant: assistant,
	}
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Title  string `json:"title"`
}

// CreateSessionResponse 创建会话响应
type CreateSessionResponse struct {
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Status    int       `json:"status"`
	CreatedAt string    `json:"created_at"`
}

// AskRequest 提问请求
type AskRequest struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id" binding:"required"`
	Question  string `json:"question" binding:"required"`
}

// AskResponse 提问响应
type AskResponse struct {
	SessionID   string                 `json:"session_id"`
	Answer      string                 `json:"answer"`
	Confidence  float64                `json:"confidence"`
	Intent      *IntentInfoResponse    `json:"intent,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	RequiresMore bool                  `json:"requires_more"`
}

// IntentInfoResponse 意图信息响应
type IntentInfoResponse struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Confidence float64                `json:"confidence"`
	Slots      map[string]interface{} `json:"slots,omitempty"`
}

// SessionListResponse 会话列表响应
type SessionListResponse struct {
	Sessions []*SessionInfo `json:"sessions"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// SessionInfo 会话信息
type SessionInfo struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Title     string          `json:"title"`
	Status    int             `json:"status"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

// SessionDetailResponse 会话详情响应
type SessionDetailResponse struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Title     string          `json:"title"`
	Status    int             `json:"status"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	Messages  []*MessageInfo  `json:"messages"`
}

// MessageInfo 消息信息
type MessageInfo struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// CreateSession 创建会话
func (s *QAService) CreateSession(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error) {
	// 创建会话实体
	session := entity.NewQASession(req.UserID, req.Title)

	// 如果没有标题，使用默认标题
	if session.Title == "" {
		session.Title = "新对话"
	}

	// 保存到数据库
	if err := s.qaRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &CreateSessionResponse{
		SessionID: session.ID,
		UserID:    session.UserID,
		Title:     session.Title,
		Status:    int(session.Status),
		CreatedAt: session.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// Ask 提问
func (s *QAService) Ask(ctx context.Context, req *AskRequest) (*AskResponse, error) {
	if req.Question == "" {
		return nil, ErrInvalidQuestion
	}

	var session *entity.QASession
	var err error
	var dialogueSessionID string

	// 获取或创建会话
	if req.SessionID != "" {
		session, err = s.qaRepo.GetSessionByID(ctx, req.SessionID)
		if err != nil {
			return nil, ErrSessionNotFound
		}
		// 验证用户权限
		if session.UserID != req.UserID {
			return nil, ErrUnauthorizedAccess
		}
		// 检查会话状态
		if session.IsDeleted() {
			return nil, ErrSessionDeleted
		}
		dialogueSessionID = session.ID
	} else {
		// 创建新会话
		title := s.assistant.GenerateTitle(req.Question)
		session = entity.NewQASession(req.UserID, title)
		if err := s.qaRepo.CreateSession(ctx, session); err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
		// 启动 DialogueManager 会话
		dialogueSessionID, err = s.assistant.StartSession(ctx, req.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to start dialogue session: %w", err)
		}
	}

	// 保存用户消息
	userMessage := session.AddMessage(entity.QAMessageRoleUser, req.Question)
	if err := s.qaRepo.CreateMessage(ctx, userMessage); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// 获取历史消息作为上下文
	recentMessages, err := s.qaRepo.GetRecentMessages(ctx, session.ID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent messages: %w", err)
	}

	// 构建上下文
	context := s.buildConversationContext(recentMessages)

	// 调用AI助手
	askReq := &qa.AskRequest{
		SessionID: dialogueSessionID,
		UserID:    req.UserID,
		Question:  req.Question,
		Context:   context,
	}

	askResp, err := s.assistant.Ask(ctx, askReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	// 保存助手消息
	assistantMessage := session.AddMessage(entity.QAMessageRoleAssistant, askResp.Answer)
	if err := s.qaRepo.CreateMessage(ctx, assistantMessage); err != nil {
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}

	// 更新会话
	if err := s.qaRepo.UpdateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	// 构建响应
	resp := &AskResponse{
		SessionID:    session.ID,
		Answer:       askResp.Answer,
		Confidence:   askResp.Confidence,
		Suggestions:  askResp.Suggestions,
		RequiresMore: askResp.RequiresMore,
	}

	if askResp.Intent != nil {
		resp.Intent = &IntentInfoResponse{
			Type:       askResp.Intent.Type,
			Name:       askResp.Intent.Name,
			Confidence: askResp.Intent.Confidence,
			Slots:      askResp.Intent.Slots,
		}
	}

	return resp, nil
}

// GetSession 获取会话详情
func (s *QAService) GetSession(ctx context.Context, sessionID, userID string) (*SessionDetailResponse, error) {
	session, err := s.qaRepo.GetSessionWithMessages(ctx, sessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	// 验证用户权限
	if session.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// 检查会话状态
	if session.IsDeleted() {
		return nil, ErrSessionDeleted
	}

	// 构建响应
	messages := make([]*MessageInfo, 0, len(session.Messages))
	for _, msg := range session.Messages {
		messages = append(messages, &MessageInfo{
			ID:        msg.ID,
			SessionID: msg.SessionID,
			Role:      string(msg.Role),
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &SessionDetailResponse{
		ID:        session.ID,
		UserID:    session.UserID,
		Title:     session.Title,
		Status:    int(session.Status),
		CreatedAt: session.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: session.UpdatedAt.Format("2006-01-02 15:04:05"),
		Messages:  messages,
	}, nil
}

// ListSessions 获取用户会话列表
func (s *QAService) ListSessions(ctx context.Context, userID string, page, pageSize int) (*SessionListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	sessions, total, err := s.qaRepo.ListSessionsByUserID(ctx, userID, nil, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	// 构建响应
	sessionInfos := make([]*SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		sessionInfos = append(sessionInfos, &SessionInfo{
			ID:        session.ID,
			UserID:    session.UserID,
			Title:     session.Title,
			Status:    int(session.Status),
			CreatedAt: session.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: session.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &SessionListResponse{
		Sessions: sessionInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// DeleteSession 删除会话
func (s *QAService) DeleteSession(ctx context.Context, sessionID, userID string) error {
	session, err := s.qaRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return ErrSessionNotFound
	}

	// 验证用户权限
	if session.UserID != userID {
		return ErrUnauthorizedAccess
	}

	// 软删除会话
	session.Delete()
	if err := s.qaRepo.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// ArchiveSession 归档会话
func (s *QAService) ArchiveSession(ctx context.Context, sessionID, userID string) error {
	session, err := s.qaRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return ErrSessionNotFound
	}

	// 验证用户权限
	if session.UserID != userID {
		return ErrUnauthorizedAccess
	}

	// 归档会话
	session.Archive()
	if err := s.qaRepo.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("failed to archive session: %w", err)
	}

	return nil
}

// buildConversationContext 构建对话上下文
func (s *QAService) buildConversationContext(messages []*entity.QAMessage) *qa.ConversationContext {
	context := &qa.ConversationContext{
		Messages:  make([]*qa.ContextMessage, 0, len(messages)),
		Variables: make(map[string]interface{}),
	}

	for _, msg := range messages {
		context.Messages = append(context.Messages, &qa.ContextMessage{
			Role:      string(msg.Role),
			Content:   msg.Content,
			Timestamp: msg.CreatedAt,
		})
	}

	return context
}
