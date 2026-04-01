package entity

import (
	"time"

	"github.com/google/uuid"
)

// QASessionStatus 问答会话状态
type QASessionStatus int

const (
	QASessionStatusActive    QASessionStatus = 1 // 活跃
	QASessionStatusArchived  QASessionStatus = 2 // 已归档
	QASessionStatusDeleted   QASessionStatus = 3 // 已删除
)

// QAMessageRole 问答消息角色
type QAMessageRole string

const (
	QAMessageRoleUser      QAMessageRole = "user"      // 用户
	QAMessageRoleAssistant QAMessageRole = "assistant" // 助手
	QAMessageRoleSystem    QAMessageRole = "system"    // 系统
)

// QASession 问答会话实体
type QASession struct {
	ID        string           `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID    string           `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Title     string           `json:"title" gorm:"type:varchar(200)"`
	Status    QASessionStatus  `json:"status" gorm:"default:1"`
	
	CreatedAt time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
	
	// 关联
	Messages  []*QAMessage     `json:"messages,omitempty" gorm:"foreignKey:SessionID"`
}

// TableName 返回表名
func (s *QASession) TableName() string {
	return "qa_sessions"
}

// NewQASession 创建问答会话
func NewQASession(userID string, title string) *QASession {
	return &QASession{
		ID:        uuid.New().String(),
		UserID:    userID,
		Title:     title,
		Status:    QASessionStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// AddMessage 添加消息
func (s *QASession) AddMessage(role QAMessageRole, content string) *QAMessage {
	message := &QAMessage{
		ID:        uuid.New().String(),
		SessionID: s.ID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}
	s.UpdatedAt = time.Now()
	return message
}

// Archive 归档会话
func (s *QASession) Archive() {
	s.Status = QASessionStatusArchived
	s.UpdatedAt = time.Now()
}

// Delete 删除会话
func (s *QASession) Delete() {
	s.Status = QASessionStatusDeleted
	s.UpdatedAt = time.Now()
}

// IsActive 是否活跃
func (s *QASession) IsActive() bool {
	return s.Status == QASessionStatusActive
}

// IsArchived 是否已归档
func (s *QASession) IsArchived() bool {
	return s.Status == QASessionStatusArchived
}

// IsDeleted 是否已删除
func (s *QASession) IsDeleted() bool {
	return s.Status == QASessionStatusDeleted
}

// QAMessage 问答消息实体
type QAMessage struct {
	ID        string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	SessionID string          `json:"session_id" gorm:"type:varchar(36);not null;index"`
	Role      QAMessageRole   `json:"role" gorm:"type:varchar(20);not null"`
	Content   string          `json:"content" gorm:"type:text;not null"`
	
	CreatedAt time.Time       `json:"created_at" gorm:"autoCreateTime"`
	
	// 关联
	Session   *QASession      `json:"session,omitempty" gorm:"foreignKey:SessionID"`
}

// TableName 返回表名
func (m *QAMessage) TableName() string {
	return "qa_messages"
}

// NewQAMessage 创建问答消息
func NewQAMessage(sessionID string, role QAMessageRole, content string) *QAMessage {
	return &QAMessage{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}
}

// IsUserMessage 是否用户消息
func (m *QAMessage) IsUserMessage() bool {
	return m.Role == QAMessageRoleUser
}

// IsAssistantMessage 是否助手消息
func (m *QAMessage) IsAssistantMessage() bool {
	return m.Role == QAMessageRoleAssistant
}

// IsSystemMessage 是否系统消息
func (m *QAMessage) IsSystemMessage() bool {
	return m.Role == QAMessageRoleSystem
}

// QASessionFilter 问答会话查询过滤器
type QASessionFilter struct {
	UserID   *string
	Status   *QASessionStatus
	Page     int
	PageSize int
}

// QAMessageFilter 问答消息查询过滤器
type QAMessageFilter struct {
	SessionID *string
	Role      *QAMessageRole
	Page      int
	PageSize  int
}
