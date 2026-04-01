package persistence

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockQARepository 问答仓储Mock
type MockQARepository struct {
	mock.Mock
}

func (m *MockQARepository) CreateSession(ctx context.Context, session *entity.QASession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockQARepository) UpdateSession(ctx context.Context, session *entity.QASession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockQARepository) DeleteSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQARepository) GetSessionByID(ctx context.Context, id string) (*entity.QASession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.QASession), args.Error(1)
}

func (m *MockQARepository) GetSessionWithMessages(ctx context.Context, id string) (*entity.QASession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.QASession), args.Error(1)
}

func (m *MockQARepository) ListSessionsByUserID(ctx context.Context, userID string, status *entity.QASessionStatus, page, pageSize int) ([]*entity.QASession, int64, error) {
	args := m.Called(ctx, userID, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.QASession), args.Get(1).(int64), args.Error(2)
}

func (m *MockQARepository) CreateMessage(ctx context.Context, message *entity.QAMessage) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockQARepository) GetMessagesBySessionID(ctx context.Context, sessionID string, page, pageSize int) ([]*entity.QAMessage, int64, error) {
	args := m.Called(ctx, sessionID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.QAMessage), args.Get(1).(int64), args.Error(2)
}

func (m *MockQARepository) GetRecentMessages(ctx context.Context, sessionID string, limit int) ([]*entity.QAMessage, error) {
	args := m.Called(ctx, sessionID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.QAMessage), args.Error(1)
}

func (m *MockQARepository) DeleteMessagesBySessionID(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func TestQARepository_CreateSession(t *testing.T) {
	ctx := context.Background()

	t.Run("成功创建会话", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		session := entity.NewQASession("user-001", "测试会话")

		mockRepo.On("CreateSession", ctx, session).Return(nil)

		err := mockRepo.CreateSession(ctx, session)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("创建会话失败", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		session := entity.NewQASession("user-001", "测试会话")

		mockRepo.On("CreateSession", ctx, session).Return(errors.New("database error"))

		err := mockRepo.CreateSession(ctx, session)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestQARepository_UpdateSession(t *testing.T) {
	ctx := context.Background()

	t.Run("成功更新会话", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		session := entity.NewQASession("user-001", "测试会话")
		session.Title = "更新后的标题"

		mockRepo.On("UpdateSession", ctx, session).Return(nil)

		err := mockRepo.UpdateSession(ctx, session)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("更新会话失败", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		session := entity.NewQASession("user-001", "测试会话")

		mockRepo.On("UpdateSession", ctx, session).Return(errors.New("database error"))

		err := mockRepo.UpdateSession(ctx, session)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestQARepository_DeleteSession(t *testing.T) {
	ctx := context.Background()

	t.Run("成功删除会话", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		sessionID := "session-001"

		mockRepo.On("DeleteSession", ctx, sessionID).Return(nil)

		err := mockRepo.DeleteSession(ctx, sessionID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("删除不存在的会话", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		sessionID := "nonexistent"

		mockRepo.On("DeleteSession", ctx, sessionID).Return(gorm.ErrRecordNotFound)

		err := mockRepo.DeleteSession(ctx, sessionID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestQARepository_GetSessionByID(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取会话", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		expectedSession := entity.NewQASession("user-001", "测试会话")

		mockRepo.On("GetSessionByID", ctx, expectedSession.ID).Return(expectedSession, nil)

		session, err := mockRepo.GetSessionByID(ctx, expectedSession.ID)

		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, expectedSession.ID, session.ID)
		assert.Equal(t, "user-001", session.UserID)
		assert.Equal(t, "测试会话", session.Title)
		mockRepo.AssertExpectations(t)
	})

	t.Run("会话不存在", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		mockRepo.On("GetSessionByID", ctx, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

		session, err := mockRepo.GetSessionByID(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, session)
		mockRepo.AssertExpectations(t)
	})
}

func TestQARepository_GetSessionWithMessages(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取会话及消息", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		session := entity.NewQASession("user-001", "测试会话")
		session.Messages = []*entity.QAMessage{
			entity.NewQAMessage(session.ID, entity.QAMessageRoleUser, "你好"),
			entity.NewQAMessage(session.ID, entity.QAMessageRoleAssistant, "你好，有什么可以帮助你的？"),
		}

		mockRepo.On("GetSessionWithMessages", ctx, session.ID).Return(session, nil)

		result, err := mockRepo.GetSessionWithMessages(ctx, session.ID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Messages, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("会话不存在", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		mockRepo.On("GetSessionWithMessages", ctx, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

		session, err := mockRepo.GetSessionWithMessages(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, session)
		mockRepo.AssertExpectations(t)
	})
}

func TestQARepository_ListSessionsByUserID(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取用户会话列表", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		userID := "user-001"
		expectedSessions := []*entity.QASession{
			entity.NewQASession(userID, "会话1"),
			entity.NewQASession(userID, "会话2"),
			entity.NewQASession(userID, "会话3"),
		}

		mockRepo.On("ListSessionsByUserID", ctx, userID, mock.Anything, 1, 10).
			Return(expectedSessions, int64(3), nil)

		sessions, total, err := mockRepo.ListSessionsByUserID(ctx, userID, nil, 1, 10)

		assert.NoError(t, err)
		assert.Len(t, sessions, 3)
		assert.Equal(t, int64(3), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("带状态过滤", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		userID := "user-001"
		status := entity.QASessionStatusActive
		expectedSessions := []*entity.QASession{
			entity.NewQASession(userID, "活跃会话"),
		}

		mockRepo.On("ListSessionsByUserID", ctx, userID, &status, 1, 10).
			Return(expectedSessions, int64(1), nil)

		sessions, total, err := mockRepo.ListSessionsByUserID(ctx, userID, &status, 1, 10)

		assert.NoError(t, err)
		assert.Len(t, sessions, 1)
		assert.Equal(t, int64(1), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("空列表", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		userID := "user-002"

		mockRepo.On("ListSessionsByUserID", ctx, userID, mock.Anything, 1, 10).
			Return([]*entity.QASession{}, int64(0), nil)

		sessions, total, err := mockRepo.ListSessionsByUserID(ctx, userID, nil, 1, 10)

		assert.NoError(t, err)
		assert.NotNil(t, sessions)
		assert.Len(t, sessions, 0)
		assert.Equal(t, int64(0), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("查询失败", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		userID := "user-001"

		mockRepo.On("ListSessionsByUserID", ctx, userID, mock.Anything, 1, 10).
			Return(nil, int64(0), errors.New("database error"))

		sessions, total, err := mockRepo.ListSessionsByUserID(ctx, userID, nil, 1, 10)

		assert.Error(t, err)
		assert.Nil(t, sessions)
		assert.Equal(t, int64(0), total)
		mockRepo.AssertExpectations(t)
	})
}

func TestQARepository_CreateMessage(t *testing.T) {
	ctx := context.Background()

	t.Run("成功创建消息", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		session := entity.NewQASession("user-001", "测试会话")
		message := entity.NewQAMessage(session.ID, entity.QAMessageRoleUser, "你好")

		mockRepo.On("CreateMessage", ctx, message).Return(nil)

		err := mockRepo.CreateMessage(ctx, message)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("创建消息失败", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		message := entity.NewQAMessage("session-001", entity.QAMessageRoleUser, "你好")

		mockRepo.On("CreateMessage", ctx, message).Return(errors.New("database error"))

		err := mockRepo.CreateMessage(ctx, message)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestQARepository_GetMessagesBySessionID(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取会话消息列表", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		sessionID := "session-001"
		expectedMessages := []*entity.QAMessage{
			entity.NewQAMessage(sessionID, entity.QAMessageRoleUser, "问题1"),
			entity.NewQAMessage(sessionID, entity.QAMessageRoleAssistant, "回答1"),
		}

		mockRepo.On("GetMessagesBySessionID", ctx, sessionID, 1, 20).
			Return(expectedMessages, int64(2), nil)

		messages, total, err := mockRepo.GetMessagesBySessionID(ctx, sessionID, 1, 20)

		assert.NoError(t, err)
		assert.Len(t, messages, 2)
		assert.Equal(t, int64(2), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("空消息列表", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		sessionID := "session-002"

		mockRepo.On("GetMessagesBySessionID", ctx, sessionID, 1, 20).
			Return([]*entity.QAMessage{}, int64(0), nil)

		messages, total, err := mockRepo.GetMessagesBySessionID(ctx, sessionID, 1, 20)

		assert.NoError(t, err)
		assert.NotNil(t, messages)
		assert.Len(t, messages, 0)
		assert.Equal(t, int64(0), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("查询失败", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		sessionID := "session-001"

		mockRepo.On("GetMessagesBySessionID", ctx, sessionID, 1, 20).
			Return(nil, int64(0), errors.New("database error"))

		messages, total, err := mockRepo.GetMessagesBySessionID(ctx, sessionID, 1, 20)

		assert.Error(t, err)
		assert.Nil(t, messages)
		assert.Equal(t, int64(0), total)
		mockRepo.AssertExpectations(t)
	})
}

func TestQARepository_GetRecentMessages(t *testing.T) {
	ctx := context.Background()

	t.Run("成功获取最近消息", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		sessionID := "session-001"
		expectedMessages := []*entity.QAMessage{
			entity.NewQAMessage(sessionID, entity.QAMessageRoleUser, "问题1"),
			entity.NewQAMessage(sessionID, entity.QAMessageRoleAssistant, "回答1"),
			entity.NewQAMessage(sessionID, entity.QAMessageRoleUser, "问题2"),
		}

		mockRepo.On("GetRecentMessages", ctx, sessionID, 10).Return(expectedMessages, nil)

		messages, err := mockRepo.GetRecentMessages(ctx, sessionID, 10)

		assert.NoError(t, err)
		assert.Len(t, messages, 3)
		mockRepo.AssertExpectations(t)
	})

	t.Run("没有消息", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		sessionID := "session-002"

		mockRepo.On("GetRecentMessages", ctx, sessionID, 10).Return([]*entity.QAMessage{}, nil)

		messages, err := mockRepo.GetRecentMessages(ctx, sessionID, 10)

		assert.NoError(t, err)
		assert.NotNil(t, messages)
		assert.Len(t, messages, 0)
		mockRepo.AssertExpectations(t)
	})

	t.Run("查询失败", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		sessionID := "session-001"

		mockRepo.On("GetRecentMessages", ctx, sessionID, 10).Return(nil, errors.New("database error"))

		messages, err := mockRepo.GetRecentMessages(ctx, sessionID, 10)

		assert.Error(t, err)
		assert.Nil(t, messages)
		mockRepo.AssertExpectations(t)
	})
}

func TestQARepository_DeleteMessagesBySessionID(t *testing.T) {
	ctx := context.Background()

	t.Run("成功删除会话消息", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		sessionID := "session-001"

		mockRepo.On("DeleteMessagesBySessionID", ctx, sessionID).Return(nil)

		err := mockRepo.DeleteMessagesBySessionID(ctx, sessionID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("删除失败", func(t *testing.T) {
		mockRepo := new(MockQARepository)

		sessionID := "session-001"

		mockRepo.On("DeleteMessagesBySessionID", ctx, sessionID).Return(errors.New("database error"))

		err := mockRepo.DeleteMessagesBySessionID(ctx, sessionID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}
