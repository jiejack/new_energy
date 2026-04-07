package service

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestQAService_CreateSession_Success(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockQARepository)
	service := NewQAService(mockRepo, nil)

	mockRepo.On("CreateSession", ctx, mock.AnythingOfType("*entity.QASession")).Return(nil)

	req := &CreateSessionRequest{
		Title: "测试会话",
	}
	resp, err := service.CreateSession(ctx, "user-001", req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "user-001", resp.UserID)
	assert.Equal(t, "测试会话", resp.Title)

	mockRepo.AssertExpectations(t)
}

func TestQAService_CreateSession_DefaultTitle(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockQARepository)
	service := NewQAService(mockRepo, nil)

	mockRepo.On("CreateSession", ctx, mock.AnythingOfType("*entity.QASession")).Return(nil)

	req := &CreateSessionRequest{}
	resp, err := service.CreateSession(ctx, "user-001", req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "新对话", resp.Title)

	mockRepo.AssertExpectations(t)
}

func TestQAService_AskQuestion_NewSession(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockQARepository)
	service := NewQAService(mockRepo, nil)

	mockRepo.On("CreateSession", ctx, mock.AnythingOfType("*entity.QASession")).Return(nil)
	mockRepo.On("CreateMessage", ctx, mock.AnythingOfType("*entity.QAMessage")).Return(nil)

	req := &AskQuestionRequest{
		Question: "查询1号逆变器的实时数据",
	}
	resp, err := service.AskQuestion(ctx, req, "user-001")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.SessionID)
	assert.Contains(t, resp.Answer, "AI服务")

	mockRepo.AssertExpectations(t)
}

func TestQAService_AskQuestion_ExistingSession(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockQARepository)
	service := NewQAService(mockRepo, nil)

	session := entity.NewQASession("user-001", "测试会话")

	mockRepo.On("GetSessionByID", ctx, session.ID).Return(session, nil)
	mockRepo.On("CreateMessage", ctx, mock.AnythingOfType("*entity.QAMessage")).Return(nil)

	req := &AskQuestionRequest{
		SessionID: session.ID,
		Question:  "查询历史数据",
	}
	resp, err := service.AskQuestion(ctx, req, "user-001")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, session.ID, resp.SessionID)

	mockRepo.AssertExpectations(t)
}

func TestQAService_AskQuestion_SessionNotFound(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockQARepository)
	service := NewQAService(mockRepo, nil)

	mockRepo.On("GetSessionByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	req := &AskQuestionRequest{
		SessionID: "nonexistent",
		Question:  "测试问题",
	}
	resp, err := service.AskQuestion(ctx, req, "user-001")

	assert.Error(t, err)
	assert.Nil(t, resp)

	mockRepo.AssertExpectations(t)
}

func TestQAService_GetSession_Success(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockQARepository)
	service := NewQAService(mockRepo, nil)

	session := entity.NewQASession("user-001", "测试会话")
	session.Messages = []*entity.QAMessage{}

	mockRepo.On("GetSessionWithMessages", ctx, session.ID).Return(session, nil)

	resp, err := service.GetSession(ctx, session.ID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, session.ID, resp.ID)

	mockRepo.AssertExpectations(t)
}

func TestQAService_GetSession_NotFound(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockQARepository)
	service := NewQAService(mockRepo, nil)

	mockRepo.On("GetSessionWithMessages", ctx, "nonexistent").Return(nil, errors.New("not found"))

	resp, err := service.GetSession(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, resp)

	mockRepo.AssertExpectations(t)
}

func TestQAService_ListUserSessions(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockQARepository)
	service := NewQAService(mockRepo, nil)

	sessions := []*entity.QASession{
		entity.NewQASession("user-001", "会话1"),
		entity.NewQASession("user-001", "会话2"),
		entity.NewQASession("user-001", "会话3"),
	}

	mockRepo.On("ListSessionsByUserID", ctx, "user-001", mock.Anything, 1, 10).
		Return(sessions, int64(3), nil)

	resp, total, err := service.ListUserSessions(ctx, "user-001", 1, 10)

	assert.NoError(t, err)
	assert.Len(t, resp, 3)
	assert.Equal(t, int64(3), total)

	mockRepo.AssertExpectations(t)
}

func TestQAService_DeleteSession_Success(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockQARepository)
	service := NewQAService(mockRepo, nil)

	session := entity.NewQASession("user-001", "测试会话")

	mockRepo.On("GetSessionByID", ctx, session.ID).Return(session, nil)
	mockRepo.On("UpdateSession", ctx, mock.AnythingOfType("*entity.QASession")).Return(nil)

	err := service.DeleteSession(ctx, session.ID)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestQAService_DeleteSession_NotFound(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockQARepository)
	service := NewQAService(mockRepo, nil)

	mockRepo.On("GetSessionByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	err := service.DeleteSession(ctx, "nonexistent")

	assert.Error(t, err)

	mockRepo.AssertExpectations(t)
}

func TestQAService_ArchiveSession_Success(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockQARepository)
	service := NewQAService(mockRepo, nil)

	session := entity.NewQASession("user-001", "测试会话")

	mockRepo.On("GetSessionByID", ctx, session.ID).Return(session, nil)
	mockRepo.On("UpdateSession", ctx, mock.AnythingOfType("*entity.QASession")).Return(nil)

	err := service.ArchiveSession(ctx, session.ID)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestQAService_GetSessionHistory(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockQARepository)
	service := NewQAService(mockRepo, nil)

	session := entity.NewQASession("user-001", "测试会话")
	messages := []*entity.QAMessage{
		entity.NewQAMessage(session.ID, entity.QAMessageRoleUser, "问题1"),
		entity.NewQAMessage(session.ID, entity.QAMessageRoleAssistant, "回答1"),
	}

	mockRepo.On("GetMessagesBySessionID", ctx, session.ID, 1, 10).
		Return(messages, int64(2), nil)

	resp, total, err := service.GetSessionHistory(ctx, session.ID, 1, 10)

	assert.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, int64(2), total)

	mockRepo.AssertExpectations(t)
}
