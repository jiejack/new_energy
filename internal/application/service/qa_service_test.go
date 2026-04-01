package service

import (
	"context"
	"errors"
	"testing"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/pkg/qa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQARepository 问答仓储Mock
type MockQARepository struct {
	mock.Mock
	sessions map[string]*entity.QASession
	messages map[string]*entity.QAMessage
}

func NewMockQARepository() *MockQARepository {
	return &MockQARepository{
		sessions: make(map[string]*entity.QASession),
		messages: make(map[string]*entity.QAMessage),
	}
}

func (m *MockQARepository) CreateSession(ctx context.Context, session *entity.QASession) error {
	args := m.Called(ctx, session)
	if args.Error(0) == nil {
		m.sessions[session.ID] = session
	}
	return args.Error(0)
}

func (m *MockQARepository) UpdateSession(ctx context.Context, session *entity.QASession) error {
	args := m.Called(ctx, session)
	if args.Error(0) == nil {
		m.sessions[session.ID] = session
	}
	return args.Error(0)
}

func (m *MockQARepository) DeleteSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if args.Error(0) == nil {
		delete(m.sessions, id)
	}
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
	if args.Error(0) == nil {
		m.messages[message.ID] = message
	}
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

// TestQAService_CreateSession_Success 测试创建会话成功
func TestQAService_CreateSession_Success(t *testing.T) {
	ctx := context.Background()

	mockRepo := NewMockQARepository()
	assistant := qa.NewAssistant(nil)
	service := NewQAService(mockRepo, assistant)

	// 设置Mock期望
	mockRepo.On("CreateSession", ctx, mock.AnythingOfType("*entity.QASession")).Return(nil)

	// 执行测试
	req := &CreateSessionRequest{
		UserID: "user-001",
		Title:  "测试会话",
	}
	resp, err := service.CreateSession(ctx, req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "user-001", resp.UserID)
	assert.Equal(t, "测试会话", resp.Title)

	mockRepo.AssertExpectations(t)
}

// TestQAService_CreateSession_DefaultTitle 测试创建会话使用默认标题
func TestQAService_CreateSession_DefaultTitle(t *testing.T) {
	ctx := context.Background()

	mockRepo := NewMockQARepository()
	assistant := qa.NewAssistant(nil)
	service := NewQAService(mockRepo, assistant)

	// 设置Mock期望
	mockRepo.On("CreateSession", ctx, mock.AnythingOfType("*entity.QASession")).Return(nil)

	// 执行测试
	req := &CreateSessionRequest{
		UserID: "user-001",
	}
	resp, err := service.CreateSession(ctx, req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "新对话", resp.Title)

	mockRepo.AssertExpectations(t)
}

// TestQAService_Ask_NewSession 测试提问创建新会话
func TestQAService_Ask_NewSession(t *testing.T) {
	ctx := context.Background()

	mockRepo := NewMockQARepository()
	assistant := qa.NewAssistant(nil)
	service := NewQAService(mockRepo, assistant)

	// 设置Mock期望
	mockRepo.On("CreateSession", ctx, mock.AnythingOfType("*entity.QASession")).Return(nil)
	mockRepo.On("CreateMessage", ctx, mock.AnythingOfType("*entity.QAMessage")).Return(nil)
	mockRepo.On("GetRecentMessages", ctx, mock.AnythingOfType("string"), 10).Return([]*entity.QAMessage{}, nil)
	mockRepo.On("UpdateSession", ctx, mock.AnythingOfType("*entity.QASession")).Return(nil)

	// 执行测试
	req := &AskRequest{
		UserID:   "user-001",
		Question: "查询1号逆变器的实时数据",
	}
	resp, err := service.Ask(ctx, req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.SessionID)
	assert.NotEmpty(t, resp.Answer)

	mockRepo.AssertExpectations(t)
}

// TestQAService_Ask_ExistingSession 测试提问使用现有会话
func TestQAService_Ask_ExistingSession(t *testing.T) {
	ctx := context.Background()

	mockRepo := NewMockQARepository()
	assistant := qa.NewAssistant(nil)
	service := NewQAService(mockRepo, assistant)

	// 准备测试数据
	session := entity.NewQASession("user-001", "测试会话")

	// 先启动一个会话
	sessionID, err := assistant.StartSession(ctx, "user-001")
	assert.NoError(t, err)
	session.ID = sessionID

	// 设置Mock期望
	mockRepo.On("GetSessionByID", ctx, sessionID).Return(session, nil)
	mockRepo.On("CreateMessage", ctx, mock.AnythingOfType("*entity.QAMessage")).Return(nil)
	mockRepo.On("GetRecentMessages", ctx, sessionID, 10).Return([]*entity.QAMessage{}, nil)
	mockRepo.On("UpdateSession", ctx, mock.AnythingOfType("*entity.QASession")).Return(nil)

	// 执行测试
	req := &AskRequest{
		SessionID: sessionID,
		UserID:    "user-001",
		Question:  "查询历史数据",
	}
	resp, err := service.Ask(ctx, req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, sessionID, resp.SessionID)

	mockRepo.AssertExpectations(t)
}

// TestQAService_Ask_EmptyQuestion 测试空问题
func TestQAService_Ask_EmptyQuestion(t *testing.T) {
	ctx := context.Background()

	mockRepo := NewMockQARepository()
	assistant := qa.NewAssistant(nil)
	service := NewQAService(mockRepo, assistant)

	// 执行测试
	req := &AskRequest{
		UserID:   "user-001",
		Question: "",
	}
	resp, err := service.Ask(ctx, req)

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidQuestion, err)
	assert.Nil(t, resp)
}

// TestQAService_Ask_SessionNotFound 测试会话不存在
func TestQAService_Ask_SessionNotFound(t *testing.T) {
	ctx := context.Background()

	mockRepo := NewMockQARepository()
	assistant := qa.NewAssistant(nil)
	service := NewQAService(mockRepo, assistant)

	// 设置Mock期望
	mockRepo.On("GetSessionByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	// 执行测试
	req := &AskRequest{
		SessionID: "nonexistent",
		UserID:    "user-001",
		Question:  "测试问题",
	}
	resp, err := service.Ask(ctx, req)

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrSessionNotFound, err)
	assert.Nil(t, resp)

	mockRepo.AssertExpectations(t)
}

// TestQAService_Ask_UnauthorizedAccess 测试未授权访问
func TestQAService_Ask_UnauthorizedAccess(t *testing.T) {
	ctx := context.Background()

	mockRepo := NewMockQARepository()
	assistant := qa.NewAssistant(nil)
	service := NewQAService(mockRepo, assistant)

	// 准备测试数据
	session := entity.NewQASession("user-002", "测试会话")

	// 设置Mock期望
	mockRepo.On("GetSessionByID", ctx, session.ID).Return(session, nil)

	// 执行测试
	req := &AskRequest{
		SessionID: session.ID,
		UserID:    "user-001",
		Question:  "测试问题",
	}
	resp, err := service.Ask(ctx, req)

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorizedAccess, err)
	assert.Nil(t, resp)

	mockRepo.AssertExpectations(t)
}

// TestQAService_GetSession_Success 测试获取会话成功
func TestQAService_GetSession_Success(t *testing.T) {
	ctx := context.Background()

	mockRepo := NewMockQARepository()
	assistant := qa.NewAssistant(nil)
	service := NewQAService(mockRepo, assistant)

	// 准备测试数据
	session := entity.NewQASession("user-001", "测试会话")
	session.Messages = []*entity.QAMessage{}

	// 设置Mock期望
	mockRepo.On("GetSessionWithMessages", ctx, session.ID).Return(session, nil)

	// 执行测试
	resp, err := service.GetSession(ctx, session.ID, "user-001")

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, session.ID, resp.ID)
	assert.Equal(t, "user-001", resp.UserID)

	mockRepo.AssertExpectations(t)
}

// TestQAService_GetSession_NotFound 测试获取会话不存在
func TestQAService_GetSession_NotFound(t *testing.T) {
	ctx := context.Background()

	mockRepo := NewMockQARepository()
	assistant := qa.NewAssistant(nil)
	service := NewQAService(mockRepo, assistant)

	// 设置Mock期望
	mockRepo.On("GetSessionWithMessages", ctx, "nonexistent").Return(nil, errors.New("not found"))

	// 执行测试
	resp, err := service.GetSession(ctx, "nonexistent", "user-001")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrSessionNotFound, err)
	assert.Nil(t, resp)

	mockRepo.AssertExpectations(t)
}

// TestQAService_ListSessions 测试获取会话列表
func TestQAService_ListSessions(t *testing.T) {
	ctx := context.Background()

	mockRepo := NewMockQARepository()
	assistant := qa.NewAssistant(nil)
	service := NewQAService(mockRepo, assistant)

	// 准备测试数据
	sessions := []*entity.QASession{
		entity.NewQASession("user-001", "会话1"),
		entity.NewQASession("user-001", "会话2"),
		entity.NewQASession("user-001", "会话3"),
	}

	// 设置Mock期望
	mockRepo.On("ListSessionsByUserID", ctx, "user-001", mock.Anything, 1, 10).
		Return(sessions, int64(3), nil)

	// 执行测试
	resp, err := service.ListSessions(ctx, "user-001", 1, 10)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Sessions, 3)
	assert.Equal(t, int64(3), resp.Total)

	mockRepo.AssertExpectations(t)
}

// TestQAService_DeleteSession_Success 测试删除会话成功
func TestQAService_DeleteSession_Success(t *testing.T) {
	ctx := context.Background()

	mockRepo := NewMockQARepository()
	assistant := qa.NewAssistant(nil)
	service := NewQAService(mockRepo, assistant)

	// 准备测试数据
	session := entity.NewQASession("user-001", "测试会话")

	// 设置Mock期望
	mockRepo.On("GetSessionByID", ctx, session.ID).Return(session, nil)
	mockRepo.On("UpdateSession", ctx, mock.AnythingOfType("*entity.QASession")).Return(nil)

	// 执行测试
	err := service.DeleteSession(ctx, session.ID, "user-001")

	// 验证结果
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// TestQAService_DeleteSession_NotFound 测试删除会话不存在
func TestQAService_DeleteSession_NotFound(t *testing.T) {
	ctx := context.Background()

	mockRepo := NewMockQARepository()
	assistant := qa.NewAssistant(nil)
	service := NewQAService(mockRepo, assistant)

	// 设置Mock期望
	mockRepo.On("GetSessionByID", ctx, "nonexistent").Return(nil, errors.New("not found"))

	// 执行测试
	err := service.DeleteSession(ctx, "nonexistent", "user-001")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrSessionNotFound, err)

	mockRepo.AssertExpectations(t)
}

// TestQAService_DeleteSession_Unauthorized 测试删除会话未授权
func TestQAService_DeleteSession_Unauthorized(t *testing.T) {
	ctx := context.Background()

	mockRepo := NewMockQARepository()
	assistant := qa.NewAssistant(nil)
	service := NewQAService(mockRepo, assistant)

	// 准备测试数据
	session := entity.NewQASession("user-002", "测试会话")

	// 设置Mock期望
	mockRepo.On("GetSessionByID", ctx, session.ID).Return(session, nil)

	// 执行测试
	err := service.DeleteSession(ctx, session.ID, "user-001")

	// 验证结果
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorizedAccess, err)

	mockRepo.AssertExpectations(t)
}

// TestQAService_ArchiveSession_Success 测试归档会话成功
func TestQAService_ArchiveSession_Success(t *testing.T) {
	ctx := context.Background()

	mockRepo := NewMockQARepository()
	assistant := qa.NewAssistant(nil)
	service := NewQAService(mockRepo, assistant)

	// 准备测试数据
	session := entity.NewQASession("user-001", "测试会话")

	// 设置Mock期望
	mockRepo.On("GetSessionByID", ctx, session.ID).Return(session, nil)
	mockRepo.On("UpdateSession", ctx, mock.AnythingOfType("*entity.QASession")).Return(nil)

	// 执行测试
	err := service.ArchiveSession(ctx, session.ID, "user-001")

	// 验证结果
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}
