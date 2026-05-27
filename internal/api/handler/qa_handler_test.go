package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockQARepoForHandler struct {
	mock.Mock
}

func (m *mockQARepoForHandler) CreateSession(ctx context.Context, session *entity.QASession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}
func (m *mockQARepoForHandler) UpdateSession(ctx context.Context, session *entity.QASession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}
func (m *mockQARepoForHandler) DeleteSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockQARepoForHandler) GetSessionByID(ctx context.Context, id string) (*entity.QASession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.QASession), args.Error(1)
}
func (m *mockQARepoForHandler) GetSessionWithMessages(ctx context.Context, id string) (*entity.QASession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.QASession), args.Error(1)
}
func (m *mockQARepoForHandler) ListSessionsByUserID(ctx context.Context, userID string, status *entity.QASessionStatus, page, pageSize int) ([]*entity.QASession, int64, error) {
	args := m.Called(ctx, userID, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.QASession), args.Get(1).(int64), args.Error(2)
}
func (m *mockQARepoForHandler) CreateMessage(ctx context.Context, message *entity.QAMessage) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}
func (m *mockQARepoForHandler) GetMessagesBySessionID(ctx context.Context, sessionID string, page, pageSize int) ([]*entity.QAMessage, int64, error) {
	args := m.Called(ctx, sessionID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.QAMessage), args.Get(1).(int64), args.Error(2)
}
func (m *mockQARepoForHandler) GetRecentMessages(ctx context.Context, sessionID string, limit int) ([]*entity.QAMessage, error) {
	args := m.Called(ctx, sessionID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.QAMessage), args.Error(1)
}
func (m *mockQARepoForHandler) DeleteMessagesBySessionID(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func setupQAHandler(qaRepo *mockQARepoForHandler) (*QAHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := service.NewQAService(qaRepo, nil)
	handler := NewQAHandler(svc)
	r := gin.New()
	return handler, r
}

func TestQAHandler_CreateSession_Success(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.POST("/qa/sessions", func(c *gin.Context) {
		c.Set("user_id", "user-001")
		handler.CreateSession(c)
	})

	qaRepo.On("CreateSession", mock.Anything, mock.AnythingOfType("*entity.QASession")).Return(nil)

	body := service.CreateSessionRequest{Title: "Test Session"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/qa/sessions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestQAHandler_CreateSession_InvalidJSON(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.POST("/qa/sessions", handler.CreateSession)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/qa/sessions", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestQAHandler_CreateSession_Error(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.POST("/qa/sessions", handler.CreateSession)

	qaRepo.On("CreateSession", mock.Anything, mock.AnythingOfType("*entity.QASession")).Return(assert.AnError)

	body := service.CreateSessionRequest{Title: "Test Session"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/qa/sessions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestQAHandler_GetSession_Success(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.GET("/qa/sessions/:id", handler.GetSession)

	session := entity.NewQASession("user-001", "Test")
	qaRepo.On("GetSessionWithMessages", mock.Anything, "session-001").Return(session, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/qa/sessions/session-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQAHandler_GetSession_Error(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.GET("/qa/sessions/:id", handler.GetSession)

	qaRepo.On("GetSessionWithMessages", mock.Anything, "nonexistent").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/qa/sessions/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestQAHandler_ListSessions_Success(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.GET("/qa/sessions", func(c *gin.Context) {
		c.Set("user_id", "user-001")
		handler.ListSessions(c)
	})

	qaRepo.On("ListSessionsByUserID", mock.Anything, "user-001", (*entity.QASessionStatus)(nil), 1, 20).Return([]*entity.QASession{}, int64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/qa/sessions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQAHandler_ListSessions_Error(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.GET("/qa/sessions", func(c *gin.Context) {
		c.Set("user_id", "user-001")
		handler.ListSessions(c)
	})

	qaRepo.On("ListSessionsByUserID", mock.Anything, "user-001", (*entity.QASessionStatus)(nil), 1, 20).Return(nil, int64(0), assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/qa/sessions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestQAHandler_DeleteSession_Success(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.DELETE("/qa/sessions/:id", handler.DeleteSession)

	session := entity.NewQASession("user-001", "Test")
	qaRepo.On("GetSessionByID", mock.Anything, "session-001").Return(session, nil)
	qaRepo.On("UpdateSession", mock.Anything, mock.AnythingOfType("*entity.QASession")).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/qa/sessions/session-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQAHandler_DeleteSession_Error(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.DELETE("/qa/sessions/:id", handler.DeleteSession)

	qaRepo.On("GetSessionByID", mock.Anything, "session-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/qa/sessions/session-001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestQAHandler_Ask_Success(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.POST("/qa/ask", func(c *gin.Context) {
		c.Set("user_id", "user-001")
		handler.Ask(c)
	})

	session := entity.NewQASession("user-001", "Test")
	qaRepo.On("GetSessionByID", mock.Anything, "session-001").Return(session, nil)
	qaRepo.On("CreateMessage", mock.Anything, mock.AnythingOfType("*entity.QAMessage")).Return(nil).Twice()

	body := service.AskQuestionRequest{SessionID: "session-001", Question: "What is the alarm?"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/qa/ask", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQAHandler_Ask_InvalidJSON(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.POST("/qa/ask", handler.Ask)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/qa/ask", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestQAHandler_ArchiveSession_Success(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.PUT("/qa/sessions/:id/archive", handler.ArchiveSession)

	session := entity.NewQASession("user-001", "Test")
	qaRepo.On("GetSessionByID", mock.Anything, "session-001").Return(session, nil)
	qaRepo.On("UpdateSession", mock.Anything, mock.AnythingOfType("*entity.QASession")).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/qa/sessions/session-001/archive", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQAHandler_ArchiveSession_Error(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.PUT("/qa/sessions/:id/archive", handler.ArchiveSession)

	qaRepo.On("GetSessionByID", mock.Anything, "session-001").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/qa/sessions/session-001/archive", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestQAHandler_GetHistory_Success(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.GET("/qa/sessions/:id/history", handler.GetHistory)

	qaRepo.On("GetMessagesBySessionID", mock.Anything, "session-001", 1, 20).Return([]*entity.QAMessage{}, int64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/qa/sessions/session-001/history", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQAHandler_GetHistory_Error(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	handler, r := setupQAHandler(qaRepo)

	r.GET("/qa/sessions/:id/history", handler.GetHistory)

	qaRepo.On("GetMessagesBySessionID", mock.Anything, "session-001", 1, 20).Return(nil, int64(0), assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/qa/sessions/session-001/history", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestQAHandler_NewQAHandler(t *testing.T) {
	qaRepo := new(mockQARepoForHandler)
	svc := service.NewQAService(qaRepo, nil)
	handler := NewQAHandler(svc)
	assert.NotNil(t, handler)
}

var _ repository.QARepository = (*mockQARepoForHandler)(nil)
