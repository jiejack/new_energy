package repository

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
)

// QARepository 问答会话仓储接口
type QARepository interface {
	// CreateSession 创建会话
	CreateSession(ctx context.Context, session *entity.QASession) error

	// UpdateSession 更新会话
	UpdateSession(ctx context.Context, session *entity.QASession) error

	// DeleteSession 删除会话
	DeleteSession(ctx context.Context, id string) error

	// GetSessionByID 根据ID获取会话
	GetSessionByID(ctx context.Context, id string) (*entity.QASession, error)

	// GetSessionWithMessages 根据ID获取会话及其消息
	GetSessionWithMessages(ctx context.Context, id string) (*entity.QASession, error)

	// ListSessionsByUserID 获取用户的会话列表
	ListSessionsByUserID(ctx context.Context, userID string, status *entity.QASessionStatus, page, pageSize int) ([]*entity.QASession, int64, error)

	// CreateMessage 创建消息
	CreateMessage(ctx context.Context, message *entity.QAMessage) error

	// GetMessagesBySessionID 获取会话的消息列表
	GetMessagesBySessionID(ctx context.Context, sessionID string, page, pageSize int) ([]*entity.QAMessage, int64, error)

	// GetRecentMessages 获取会话最近的N条消息
	GetRecentMessages(ctx context.Context, sessionID string, limit int) ([]*entity.QAMessage, error)

	// DeleteMessagesBySessionID 删除会话的所有消息
	DeleteMessagesBySessionID(ctx context.Context, sessionID string) error
}
