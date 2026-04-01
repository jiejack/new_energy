package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	TraceID string                 `json:"trace_id,omitempty"`
}

type SuccessResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

func WriteError(c *gin.Context, err error) {
	if err == nil {
		WriteSuccess(c, nil)
		return
	}

	var appErr *AppError
	if As(err, &appErr) {
		appErr = err.(*AppError)
	} else {
		appErr = Wrap(err, ErrInternalServer, "服务器内部错误")
	}

	traceID := c.GetString("trace_id")
	if traceID == "" {
		traceID = generateTraceID()
	}

	response := ErrorResponse{
		Code:    int(appErr.Code),
		Message: appErr.Message,
		TraceID: traceID,
	}

	if len(appErr.Context) > 0 {
		response.Details = appErr.Context
	}

	statusCode := getHTTPStatusCode(appErr.Code)
	c.JSON(statusCode, response)
}

func WriteSuccess(c *gin.Context, data interface{}) {
	traceID := c.GetString("trace_id")
	if traceID == "" {
		traceID = generateTraceID()
	}

	response := SuccessResponse{
		Code:    0,
		Message: "操作成功",
		Data:    data,
		TraceID: traceID,
	}

	c.JSON(http.StatusOK, response)
}

func WriteSuccessWithMessage(c *gin.Context, message string, data interface{}) {
	traceID := c.GetString("trace_id")
	if traceID == "" {
		traceID = generateTraceID()
	}

	response := SuccessResponse{
		Code:    0,
		Message: message,
		Data:    data,
		TraceID: traceID,
	}

	c.JSON(http.StatusOK, response)
}

func WriteErrorResponse(c *gin.Context, code ErrorCode, message string) {
	traceID := c.GetString("trace_id")
	if traceID == "" {
		traceID = generateTraceID()
	}

	response := ErrorResponse{
		Code:    int(code),
		Message: message,
		TraceID: traceID,
	}

	statusCode := getHTTPStatusCode(code)
	c.JSON(statusCode, response)
}

func getHTTPStatusCode(code ErrorCode) int {
	switch code {
	case ErrSuccess:
		return http.StatusOK
	case ErrInvalidParam, ErrInvalidPassword, ErrInvalidToken,
		ErrInvalidPointValue, ErrConfigInvalid, ErrRuleInvalidExpression:
		return http.StatusBadRequest
	case ErrUnauthorized, ErrInvalidToken, ErrTokenExpired:
		return http.StatusUnauthorized
	case ErrForbidden, ErrInsufficientPermission:
		return http.StatusForbidden
	case ErrNotFound, ErrUserNotFound, ErrDeviceNotFound, ErrStationNotFound,
		ErrPointNotFound, ErrAlarmNotFound, ErrRuleNotFound, ErrConfigNotFound:
		return http.StatusNotFound
	case ErrConflict, ErrUserAlreadyExists, ErrDeviceAlreadyExists,
		ErrStationAlreadyExists, ErrPointAlreadyExists, ErrRuleAlreadyExists,
		ErrConfigAlreadyExists:
		return http.StatusConflict
	case ErrInternalServer, ErrDatabaseError, ErrCacheError,
		ErrQueueError, ErrNetworkError:
		return http.StatusInternalServerError
	case ErrServiceUnavailable, ErrResourceExhausted:
		return http.StatusServiceUnavailable
	case ErrTimeout, ErrCollectTimeout:
		return http.StatusGatewayTimeout
	case ErrRateLimitExceeded:
		return http.StatusTooManyRequests
	default:
		if code < 500 {
			return http.StatusBadRequest
		}
		return http.StatusInternalServerError
	}
}

func generateTraceID() string {
	return ""
}

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				appErr := New(ErrInternalServer, "服务器内部错误")
				appErr = appErr.WithContext("panic", err)
				c.Set("error", appErr)
				WriteError(c, appErr)
				c.Abort()
			}
		}()
		c.Next()
	}
}

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			if err != nil {
				WriteError(c, err.Err)
			}
		}
	}
}
