package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQASession(t *testing.T) {
	session := NewQASession("user-001", "Test Session")
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "user-001", session.UserID)
	assert.Equal(t, "Test Session", session.Title)
	assert.Equal(t, QASessionStatusActive, session.Status)
	assert.NotZero(t, session.CreatedAt)
	assert.NotZero(t, session.UpdatedAt)
}

func TestQASession_AddMessage(t *testing.T) {
	session := NewQASession("user-001", "Test Session")
	msg := session.AddMessage(QAMessageRoleUser, "Hello")
	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, session.ID, msg.SessionID)
	assert.Equal(t, QAMessageRoleUser, msg.Role)
	assert.Equal(t, "Hello", msg.Content)
	assert.NotZero(t, msg.CreatedAt)
}

func TestQASession_Archive(t *testing.T) {
	session := NewQASession("user-001", "Test Session")
	session.Archive()
	assert.Equal(t, QASessionStatusArchived, session.Status)
	assert.True(t, session.IsArchived())
	assert.False(t, session.IsActive())
}

func TestQASession_Delete(t *testing.T) {
	session := NewQASession("user-001", "Test Session")
	session.Delete()
	assert.Equal(t, QASessionStatusDeleted, session.Status)
	assert.True(t, session.IsDeleted())
}

func TestQASession_IsActive(t *testing.T) {
	session := NewQASession("user-001", "Test Session")
	assert.True(t, session.IsActive())

	session.Archive()
	assert.False(t, session.IsActive())
}

func TestQASession_IsArchived(t *testing.T) {
	session := NewQASession("user-001", "Test Session")
	assert.False(t, session.IsArchived())

	session.Archive()
	assert.True(t, session.IsArchived())
}

func TestQASession_IsDeleted(t *testing.T) {
	session := NewQASession("user-001", "Test Session")
	assert.False(t, session.IsDeleted())

	session.Delete()
	assert.True(t, session.IsDeleted())
}

func TestQASession_TableName(t *testing.T) {
	session := QASession{}
	assert.Equal(t, "qa_sessions", session.TableName())
}

func TestNewQAMessage(t *testing.T) {
	msg := NewQAMessage("session-001", QAMessageRoleUser, "Hello")
	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, "session-001", msg.SessionID)
	assert.Equal(t, QAMessageRoleUser, msg.Role)
	assert.Equal(t, "Hello", msg.Content)
	assert.NotZero(t, msg.CreatedAt)
}

func TestQAMessage_IsUserMessage(t *testing.T) {
	msg := NewQAMessage("session-001", QAMessageRoleUser, "Hello")
	assert.True(t, msg.IsUserMessage())
	assert.False(t, msg.IsAssistantMessage())
	assert.False(t, msg.IsSystemMessage())
}

func TestQAMessage_IsAssistantMessage(t *testing.T) {
	msg := NewQAMessage("session-001", QAMessageRoleAssistant, "Hi there!")
	assert.True(t, msg.IsAssistantMessage())
	assert.False(t, msg.IsUserMessage())
}

func TestQAMessage_IsSystemMessage(t *testing.T) {
	msg := NewQAMessage("session-001", QAMessageRoleSystem, "System message")
	assert.True(t, msg.IsSystemMessage())
	assert.False(t, msg.IsUserMessage())
}

func TestQAMessage_TableName(t *testing.T) {
	msg := QAMessage{}
	assert.Equal(t, "qa_messages", msg.TableName())
}

func TestQASessionStatus_Constants(t *testing.T) {
	assert.Equal(t, QASessionStatus(1), QASessionStatusActive)
	assert.Equal(t, QASessionStatus(2), QASessionStatusArchived)
	assert.Equal(t, QASessionStatus(3), QASessionStatusDeleted)
}

func TestQAMessageRole_Constants(t *testing.T) {
	assert.Equal(t, QAMessageRole("user"), QAMessageRoleUser)
	assert.Equal(t, QAMessageRole("assistant"), QAMessageRoleAssistant)
	assert.Equal(t, QAMessageRole("system"), QAMessageRoleSystem)
}
