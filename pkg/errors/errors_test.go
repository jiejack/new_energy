package errors

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestErrorCode_IsSuccess(t *testing.T) {
	assert.True(t, ErrSuccess.IsSuccess())
	assert.False(t, ErrUnknown.IsSuccess())
	assert.False(t, ErrUserNotFound.IsSuccess())
}

func TestErrorCode_IsClientError(t *testing.T) {
	assert.True(t, ErrInvalidParam.IsClientError())
	assert.True(t, ErrUnauthorized.IsClientError())
	assert.True(t, ErrForbidden.IsClientError())
	assert.True(t, ErrNotFound.IsClientError())
	assert.True(t, ErrConflict.IsClientError())
	assert.False(t, ErrSuccess.IsClientError())
	assert.False(t, ErrInternalServer.IsClientError())
	assert.False(t, ErrUserNotFound.IsClientError())
}

func TestErrorCode_String_Unknown(t *testing.T) {
	unknown := ErrorCode(99999)
	assert.Equal(t, "未知错误", unknown.String())
}

func TestErrorCode_AllMessages(t *testing.T) {
	codes := []ErrorCode{
		ErrSuccess, ErrInvalidParam, ErrUnauthorized,
		ErrForbidden, ErrNotFound, ErrConflict, ErrInternalServer,
		ErrServiceUnavailable, ErrTimeout, ErrUserNotFound,
		ErrUserAlreadyExists, ErrInvalidPassword, ErrInvalidToken,
		ErrTokenExpired, ErrInsufficientPermission, ErrDeviceNotFound,
		ErrDeviceOffline, ErrDeviceCommunication, ErrDeviceAlreadyExists,
		ErrDeviceTypeMismatch, ErrStationNotFound, ErrStationAlreadyExists,
		ErrStationOffline, ErrPointNotFound, ErrPointAlreadyExists,
		ErrDataQuality, ErrCollectTimeout, ErrInvalidPointValue,
		ErrAlarmNotFound, ErrAlarmRuleInvalid, ErrAlarmAlreadyAcknowledged,
		ErrAlarmAlreadyCleared, ErrRuleNotFound, ErrRuleAlreadyExists,
		ErrRuleInvalidExpression, ErrConfigNotFound, ErrConfigAlreadyExists,
		ErrConfigInvalid, ErrDatabaseError, ErrCacheError, ErrQueueError,
		ErrNetworkError, ErrRateLimitExceeded, ErrResourceExhausted,
	}
	for _, code := range codes {
		msg := code.String()
		assert.NotEqual(t, "未知错误", msg, "code %d should have a message", code)
	}
}

func TestAppError_New(t *testing.T) {
	err := New(ErrUserNotFound, "用户不存在")
	assert.Equal(t, ErrUserNotFound, err.Code)
	assert.Equal(t, "用户不存在", err.Message)
	assert.Nil(t, err.Cause)
	assert.NotNil(t, err.Stack)
	assert.NotNil(t, err.Context)
}

func TestAppError_Newf(t *testing.T) {
	err := Newf(ErrNotFound, "%s不存在", "用户")
	assert.Equal(t, ErrNotFound, err.Code)
	assert.Equal(t, "用户不存在", err.Message)
}

func TestAppError_Wrap_NilError(t *testing.T) {
	err := Wrap(nil, ErrInternalServer, "服务不可用")
	assert.Equal(t, ErrInternalServer, err.Code)
	assert.Equal(t, "服务不可用", err.Message)
	assert.Nil(t, err.Cause)
}

func TestAppError_Wrap_NonAppError(t *testing.T) {
	cause := errors.New("some error")
	err := Wrap(cause, ErrInternalServer, "服务不可用")
	assert.Equal(t, ErrInternalServer, err.Code)
	assert.Equal(t, "服务不可用", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestAppError_Wrap_AppError(t *testing.T) {
	inner := New(ErrDatabaseError, "连接失败")
	wrapped := Wrap(inner, ErrInternalServer, "服务不可用")
	assert.Equal(t, ErrInternalServer, wrapped.Code)
	assert.Equal(t, "服务不可用", wrapped.Message)
	assert.Nil(t, wrapped.Cause, "Wrap(AppError) copies inner.Cause which is nil")
}

func TestAppError_Wrap_AppErrorWithCause(t *testing.T) {
	rootCause := errors.New("root cause")
	inner := New(ErrDatabaseError, "连接失败")
	inner.Cause = rootCause
	wrapped := Wrap(inner, ErrInternalServer, "服务不可用")
	assert.Equal(t, rootCause, wrapped.Cause, "Wrap(AppError) copies inner.Cause")
}

func TestAppError_Wrapf(t *testing.T) {
	cause := New(ErrDatabaseError, "连接失败")
	err := Wrapf(cause, ErrInternalServer, "处理%s失败", "请求")
	assert.Equal(t, ErrInternalServer, err.Code)
	assert.Equal(t, "处理请求失败", err.Message)
}

func TestAppError_Error_WithCause(t *testing.T) {
	cause := errors.New("root cause")
	err := New(ErrInternalServer, "服务不可用")
	err.Cause = cause
	s := err.Error()
	assert.Contains(t, s, "服务器内部错误")
	assert.Contains(t, s, "服务不可用")
	assert.Contains(t, s, "caused by: root cause")
}

func TestAppError_Error_NoCause(t *testing.T) {
	err := New(ErrUserNotFound, "用户不存在")
	s := err.Error()
	assert.Contains(t, s, "用户不存在")
	assert.NotContains(t, s, "caused by:")
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := New(ErrInternalServer, "服务不可用")
	err.Cause = cause
	assert.Equal(t, cause, err.Unwrap())
}

func TestAppError_Unwrap_Nil(t *testing.T) {
	err := New(ErrUserNotFound, "用户不存在")
	assert.Nil(t, err.Unwrap())
}

func TestAppError_WithContext(t *testing.T) {
	err := New(ErrInvalidParam, "参数错误").
		WithContext("field", "username").
		WithContext("reason", "too short")
	val, ok := err.GetContext("field")
	assert.True(t, ok)
	assert.Equal(t, "username", val)
	val, ok = err.GetContext("reason")
	assert.True(t, ok)
	assert.Equal(t, "too short", val)
	_, ok = err.GetContext("nonexistent")
	assert.False(t, ok)
}

func TestAppError_Is(t *testing.T) {
	err1 := New(ErrUserNotFound, "用户不存在")
	err2 := New(ErrUserNotFound, "其他消息")
	err3 := New(ErrDeviceNotFound, "设备不存在")
	assert.True(t, err1.Is(err2))
	assert.False(t, err1.Is(err3))
}

func TestAppError_Is_NonAppError(t *testing.T) {
	err := New(ErrUserNotFound, "用户不存在")
	other := errors.New("other error")
	assert.False(t, err.Is(other))
}

func TestAppError_IsCode(t *testing.T) {
	err := New(ErrUserNotFound, "用户不存在")
	assert.True(t, err.IsCode(ErrUserNotFound))
	assert.False(t, err.IsCode(ErrDeviceNotFound))
}

func TestAppError_GetStack(t *testing.T) {
	err := New(ErrUserNotFound, "用户不存在")
	stack := err.GetStack()
	assert.NotNil(t, stack)
	assert.NotEmpty(t, stack)
}

func TestAppError_Format(t *testing.T) {
	err := New(ErrUserNotFound, "用户不存在").
		WithContext("field", "username")
	s := err.Format()
	assert.Contains(t, s, "Error:")
	assert.Contains(t, s, "Code:")
	assert.Contains(t, s, "Context:")
	assert.Contains(t, s, "field=username")
}

func TestAppError_Format_NoContext(t *testing.T) {
	err := New(ErrUserNotFound, "用户不存在")
	s := err.Format()
	assert.Contains(t, s, "Error:")
	assert.NotContains(t, s, "Context:")
}

func TestIs_Function(t *testing.T) {
	err1 := New(ErrUserNotFound, "用户不存在")
	err2 := New(ErrUserNotFound, "其他消息")
	err3 := New(ErrDeviceNotFound, "设备不存在")
	assert.True(t, Is(err1, err2))
	assert.False(t, Is(err1, err3))
}

func TestIs_Function_NonAppError(t *testing.T) {
	err1 := errors.New("error1")
	err2 := errors.New("error1")
	assert.True(t, Is(err1, err1))
	assert.False(t, Is(err1, err2))
}

func TestAs_Function_AlwaysFalse(t *testing.T) {
	err := New(ErrUserNotFound, "用户不存在")
	var appErr *AppError
	result := As(err, &appErr)
	assert.False(t, result)
}

func TestAs_Function_WrappedError(t *testing.T) {
	inner := errors.New("root cause")
	wrapped := Wrap(inner, ErrInternalServer, "服务不可用")
	var appErr *AppError
	result := As(wrapped, &appErr)
	assert.False(t, result)
}

func TestAs_Function_NonAppError(t *testing.T) {
	err := errors.New("plain error")
	var appErr *AppError
	result := As(err, &appErr)
	assert.False(t, result)
}

func TestNewBusinessError(t *testing.T) {
	err := NewBusinessError(ErrUserNotFound, "用户不存在")
	assert.Equal(t, ErrUserNotFound, err.Code)
	assert.Equal(t, "用户不存在", err.Message)
	assert.True(t, err.Code.IsBusinessError())
}

func TestNewSystemError(t *testing.T) {
	err := NewSystemError("系统错误")
	assert.Equal(t, ErrInternalServer, err.Code)
	assert.Equal(t, "系统错误", err.Message)
	assert.True(t, err.Code.IsServerError())
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("参数无效")
	assert.Equal(t, ErrInvalidParam, err.Code)
	assert.Equal(t, "参数无效", err.Message)
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("用户")
	assert.Equal(t, ErrNotFound, err.Code)
	assert.Contains(t, err.Message, "用户")
}

func TestGetHTTPStatusCode(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected int
	}{
		{ErrSuccess, http.StatusOK},
		{ErrInvalidParam, http.StatusBadRequest},
		{ErrInvalidPassword, http.StatusBadRequest},
		{ErrInvalidPointValue, http.StatusBadRequest},
		{ErrConfigInvalid, http.StatusBadRequest},
		{ErrRuleInvalidExpression, http.StatusBadRequest},
		{ErrUnauthorized, http.StatusUnauthorized},
		{ErrTokenExpired, http.StatusUnauthorized},
		{ErrForbidden, http.StatusForbidden},
		{ErrInsufficientPermission, http.StatusForbidden},
		{ErrNotFound, http.StatusNotFound},
		{ErrUserNotFound, http.StatusNotFound},
		{ErrDeviceNotFound, http.StatusNotFound},
		{ErrStationNotFound, http.StatusNotFound},
		{ErrPointNotFound, http.StatusNotFound},
		{ErrAlarmNotFound, http.StatusNotFound},
		{ErrRuleNotFound, http.StatusNotFound},
		{ErrConfigNotFound, http.StatusNotFound},
		{ErrConflict, http.StatusConflict},
		{ErrUserAlreadyExists, http.StatusConflict},
		{ErrDeviceAlreadyExists, http.StatusConflict},
		{ErrStationAlreadyExists, http.StatusConflict},
		{ErrPointAlreadyExists, http.StatusConflict},
		{ErrRuleAlreadyExists, http.StatusConflict},
		{ErrConfigAlreadyExists, http.StatusConflict},
		{ErrInternalServer, http.StatusInternalServerError},
		{ErrDatabaseError, http.StatusInternalServerError},
		{ErrCacheError, http.StatusInternalServerError},
		{ErrQueueError, http.StatusInternalServerError},
		{ErrNetworkError, http.StatusInternalServerError},
		{ErrServiceUnavailable, http.StatusServiceUnavailable},
		{ErrResourceExhausted, http.StatusServiceUnavailable},
		{ErrTimeout, http.StatusGatewayTimeout},
		{ErrCollectTimeout, http.StatusGatewayTimeout},
		{ErrRateLimitExceeded, http.StatusTooManyRequests},
		{ErrorCode(450), http.StatusBadRequest},
		{ErrorCode(600), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		result := getHTTPStatusCode(tt.code)
		assert.Equal(t, tt.expected, result, "getHTTPStatusCode(%d)", tt.code)
	}
}

func TestWriteError_Nil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	WriteError(c, nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWriteError_AppError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	err := New(ErrUserNotFound, "用户不存在").WithContext("field", "test")
	WriteError(c, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWriteError_StandardError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	WriteError(c, errors.New("standard error"))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWriteSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	WriteSuccess(c, map[string]string{"key": "value"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWriteSuccessWithMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	WriteSuccessWithMessage(c, "操作成功", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWriteErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	WriteErrorResponse(c, ErrInvalidParam, "参数无效")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RecoveryMiddleware())
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestErrorHandlerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/error", func(c *gin.Context) {
		_ = c.Error(errors.New("test error"))
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGenerateTraceID(t *testing.T) {
	id := generateTraceID()
	assert.Equal(t, "", id)
}
