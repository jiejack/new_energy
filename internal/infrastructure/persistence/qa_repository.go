package persistence

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"gorm.io/gorm"
)

// QARepository 问答会话仓储实现
type QARepository struct {
	db *Database
}

// NewQARepository 创建问答会话仓储
func NewQARepository(db *Database) repository.QARepository {
	return &QARepository{db: db}
}

// CreateSession 创建会话
func (r *QARepository) CreateSession(ctx context.Context, session *entity.QASession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// UpdateSession 更新会话
func (r *QARepository) UpdateSession(ctx context.Context, session *entity.QASession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// DeleteSession 删除会话（软删除，更新状态为已删除）
func (r *QARepository) DeleteSession(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&entity.QASession{}).
		Where("id = ?", id).
		Update("status", entity.QASessionStatusDeleted).Error
}

// GetSessionByID 根据ID获取会话
func (r *QARepository) GetSessionByID(ctx context.Context, id string) (*entity.QASession, error) {
	var session entity.QASession
	err := r.db.WithContext(ctx).
		Where("id = ? AND status != ?", id, entity.QASessionStatusDeleted).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessionWithMessages 根据ID获取会话及其消息
func (r *QARepository) GetSessionWithMessages(ctx context.Context, id string) (*entity.QASession, error) {
	var session entity.QASession
	err := r.db.WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("id = ? AND status != ?", id, entity.QASessionStatusDeleted).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// ListSessionsByUserID 获取用户的会话列表
func (r *QARepository) ListSessionsByUserID(ctx context.Context, userID string, status *entity.QASessionStatus, page, pageSize int) ([]*entity.QASession, int64, error) {
	var sessions []*entity.QASession
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.QASession{}).
		Where("user_id = ?", userID).
		Where("status != ?", entity.QASessionStatusDeleted)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("updated_at DESC").Find(&sessions).Error
	return sessions, total, err
}

// CreateMessage 创建消息
func (r *QARepository) CreateMessage(ctx context.Context, message *entity.QAMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// GetMessagesBySessionID 获取会话的消息列表
func (r *QARepository) GetMessagesBySessionID(ctx context.Context, sessionID string, page, pageSize int) ([]*entity.QAMessage, int64, error) {
	var messages []*entity.QAMessage
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.QAMessage{}).
		Where("session_id = ?", sessionID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at ASC").Find(&messages).Error
	return messages, total, err
}

// GetRecentMessages 获取会话最近的N条消息
func (r *QARepository) GetRecentMessages(ctx context.Context, sessionID string, limit int) ([]*entity.QAMessage, error) {
	var messages []*entity.QAMessage
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// 反转顺序，使消息按时间正序排列
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// DeleteMessagesBySessionID 删除会话的所有消息
func (r *QARepository) DeleteMessagesBySessionID(ctx context.Context, sessionID string) error {
	return r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Delete(&entity.QAMessage{}).Error
}
