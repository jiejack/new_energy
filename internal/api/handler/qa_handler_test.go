package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/pkg/qa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

// MockAssistant AI助手Mock
type MockAssistant struct {
	mock.Mock
}

func (m *MockAssistant) Ask(ctx context.Context, req *qa.AskRequest) (*qa.AskResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*qa.AskResponse), args.Error(1)
}

func (m *MockAssistant) StartSession(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockAssistant) EndSession(sessionID string) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockAssistant) GenerateTitle(firstMessage string) string {
	args := m.Called(firstMessage)
	return args.String(0)
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestQAHandler_CreateSession(t *testing.T) {
	t.Run("成功创建会话", func(t *testing.T) {
		mockRepo := new(MockQARepository)
		mockAssistant := new(MockAssistant)
		qaService := service.NewQAService(mockRepo, mockAssistant)
		handler := NewQAHandler(qaService)

		req := map[string]interface{}{
			"user_id": "user-001",
			"title":   "测试会话",
		}

		mockRepo.On("CreateSession", mock.Anything, mock.AnythingOfType("*entity.QASession")).Return(nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/qa/sessions", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateSession(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("无效的请求参数", func(t *testing.T) {
		mockRepo := new(MockQARepository)
		mockAssistant := new(MockAssistant)
		qaService := service.NewQAService(mockRepo, mockAssistant)
		handler := NewQAHandler(qaService)

		body := []byte(`{"invalid": "data"}`)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/qa/sessions", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateSession(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestQAHandler_GetSession(t *testing.T) {
	t.Run("成功获取会话", func(t *testing.T) {
		mockRepo := new(MockQARepository)
		mockAssistant := new(MockAssistant)
		qaService := service.NewQAService(mockRepo, mockAssistant)
		handler := NewQAHandler(qaService)

		sessionID := "session-001"
		userID := "user-001"

		expectedSession := entity.NewQASession(userID, "测试会话")
		expectedSession.ID = sessionID

		mockRepo.On("GetSessionWithMessages", mock.Anything, sessionID).Return(expectedSession, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/qa/sessions/"+sessionID, nil)
		c.Params = gin.Params{{Key: "id", Value: sessionID}}
		c.Set("user_id", userID)

		handler.GetSession(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("会话不存在", func(t *testing.T) {
		mockRepo := new(MockQARepository)
		mockAssistant := new(MockAssistant)
		qaService := service.NewQAService(mockRepo, mockAssistant)
		handler := NewQAHandler(qaService)

		sessionID := "nonexistent"
		userID := "user-001"

		mockRepo.On("GetSessionWithMessages", mock.Anything, sessionID).Return(nil, errors.New("not found"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/qa/sessions/"+sessionID, nil)
		c.Params = gin.Params{{Key: "id", Value: sessionID}}
		c.Set("user_id", userID)

		handler.GetSession(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestQAHandler_ListSessions(t *testing.T) {
	t.Run("成功获取会话列表", func(t *testing.T) {
		mockRepo := new(MockQARepository)
		mockAssistant := new(MockAssistant)
		qaService := service.NewQAService(mockRepo, mockAssistant)
		handler := NewQAHandler(qaService)

		userID := "user-001"

		expectedSessions := []*entity.QASession{
			entity.NewQASession(userID, "会话1"),
			entity.NewQASession(userID, "会话2"),
		}

		mockRepo.On("ListSessionsByUserID", mock.Anything, userID, mock.Anything, 1, 20).
			Return(expectedSessions, int64(2), nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/qa/sessions?page=1&page_size=20", nil)
		c.Set("user_id", userID)

		handler.ListSessions(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("服务返回错误", func(t *testing.T) {
		mockRepo := new(MockQARepository)
		mockAssistant := new(MockAssistant)
		qaService := service.NewQAService(mockRepo, mockAssistant)
		handler := NewQAHandler(qaService)

		userID := "user-001"

		mockRepo.On("ListSessionsByUserID", mock.Anything, userID, mock.Anything, 1, 20).
			Return(nil, int64(0), errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/qa/sessions", nil)
		c.Set("user_id", userID)

		handler.ListSessions(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestQAHandler_DeleteSession(t *testing.T) {
	t.Run("成功删除会话", func(t *testing.T) {
		mockRepo := new(MockQARepository)
		mockAssistant := new(MockAssistant)
		qaService := service.NewQAService(mockRepo, mockAssistant)
		handler := NewQAHandler(qaService)

		sessionID := "session-001"
		userID := "user-001"

		expectedSession := entity.NewQASession(userID, "测试会话")
		expectedSession.ID = sessionID

		mockRepo.On("GetSessionByID", mock.Anything, sessionID).Return(expectedSession, nil)
		mockRepo.On("UpdateSession", mock.Anything, mock.AnythingOfType("*entity.QASession")).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/qa/sessions/"+sessionID, nil)
		c.Params = gin.Params{{Key: "id", Value: sessionID}}
		c.Set("user_id", userID)

		handler.DeleteSession(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("会话不存在", func(t *testing.T) {
		mockRepo := new(MockQARepository)
		mockAssistant := new(MockAssistant)
		qaService := service.NewQAService(mockRepo, mockAssistant)
		handler := NewQAHandler(qaService)

		sessionID := "nonexistent"
		userID := "user-001"

		mockRepo.On("GetSessionByID", mock.Anything, sessionID).Return(nil, errors.New("not found"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/qa/sessions/"+sessionID, nil)
		c.Params = gin.Params{{Key: "id", Value: sessionID}}
		c.Set("user_id", userID)

		handler.DeleteSession(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
