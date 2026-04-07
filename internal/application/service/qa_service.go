package service

import (
	"context"
	"fmt"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/new-energy-monitoring/pkg/ai/qa"
)

type QAService struct {
	qaRepo      repository.QARepository
	dialogueMgr *qa.DialogueManager
}

func NewQAService(
	qaRepo repository.QARepository,
	dialogueMgr *qa.DialogueManager,
) *QAService {
	return &QAService{
		qaRepo:      qaRepo,
		dialogueMgr: dialogueMgr,
	}
}

type AskQuestionRequest struct {
	SessionID string `json:"session_id"`
	Question  string `json:"question" binding:"required"`
	UserID    string `json:"user_id,omitempty"`
}

type AskQuestionResponse struct {
	SessionID string `json:"session_id"`
	Answer    string `json:"answer"`
}

type CreateSessionRequest struct {
	Title string `json:"title"`
}

func (s *QAService) CreateSession(ctx context.Context, userID string, req *CreateSessionRequest) (*entity.QASession, error) {
	title := req.Title
	if title == "" {
		title = "新对话"
	}

	session := entity.NewQASession(userID, title)

	if err := s.qaRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (s *QAService) AskQuestion(ctx context.Context, req *AskQuestionRequest, userID string) (*AskQuestionResponse, error) {
	var session *entity.QASession
	var err error

	if req.SessionID == "" {
		session = entity.NewQASession(userID, "新对话")
		if err := s.qaRepo.CreateSession(ctx, session); err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
	} else {
		session, err = s.qaRepo.GetSessionByID(ctx, req.SessionID)
		if err != nil {
			return nil, fmt.Errorf("session not found: %w", err)
		}
	}

	userMsg := entity.NewQAMessage(session.ID, entity.QAMessageRoleUser, req.Question)
	if err := s.qaRepo.CreateMessage(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	answer, err := s.getAIResponse(ctx, session.ID, req.Question)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	assistantMsg := entity.NewQAMessage(session.ID, entity.QAMessageRoleAssistant, answer)
	if err := s.qaRepo.CreateMessage(ctx, assistantMsg); err != nil {
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}

	return &AskQuestionResponse{
		SessionID: session.ID,
		Answer:    answer,
	}, nil
}

func (s *QAService) getAIResponse(ctx context.Context, sessionID, question string) (string, error) {
	if s.dialogueMgr == nil {
		return "AI服务暂未配置，请检查系统设置。", nil
	}

	response, err := s.dialogueMgr.Process(ctx, sessionID, question)
	if err != nil {
		return "", fmt.Errorf("AI dialogue failed: %w", err)
	}

	return response.Content, nil
}

func (s *QAService) GetSession(ctx context.Context, sessionID string) (*entity.QASession, error) {
	session, err := s.qaRepo.GetSessionWithMessages(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	return session, nil
}

func (s *QAService) ListUserSessions(ctx context.Context, userID string, page, pageSize int) ([]*entity.QASession, int64, error) {
	sessions, total, err := s.qaRepo.ListSessionsByUserID(ctx, userID, nil, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list sessions: %w", err)
	}
	return sessions, total, nil
}

func (s *QAService) DeleteSession(ctx context.Context, sessionID string) error {
	session, err := s.qaRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	session.Delete()
	return s.qaRepo.UpdateSession(ctx, session)
}

func (s *QAService) ArchiveSession(ctx context.Context, sessionID string) error {
	session, err := s.qaRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	session.Archive()
	return s.qaRepo.UpdateSession(ctx, session)
}

func (s *QAService) GetSessionHistory(ctx context.Context, sessionID string, page, pageSize int) ([]*entity.QAMessage, int64, error) {
	return s.qaRepo.GetMessagesBySessionID(ctx, sessionID, page, pageSize)
}
