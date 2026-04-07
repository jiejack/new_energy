package errors

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Cause   error
	Stack   []byte
	Context map[string]interface{}
	mu      sync.RWMutex
}

func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
		Stack:   captureStack(),
	}
}

func Newf(code ErrorCode, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Context: make(map[string]interface{}),
		Stack:   captureStack(),
	}
}

func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return New(code, message)
	}

	appErr, ok := err.(*AppError)
	if !ok {
		return &AppError{
			Code:    code,
			Message: message,
			Cause:   err,
			Context: make(map[string]interface{}),
			Stack:   captureStack(),
		}
	}

	return &AppError{
		Code:    code,
		Message: message,
		Cause:   appErr.Cause,
		Stack:   captureStack(),
		Context: appErr.Context,
	}
}

func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *AppError {
	return Wrap(err, code, fmt.Sprintf(format, args...))
}

func (e *AppError) Error() string {
	var sb strings.Builder
	sb.WriteString(e.Code.String())
	if e.Message != "" {
		sb.WriteString(": ")
		sb.WriteString(e.Message)
	}
	if e.Cause != nil {
		sb.WriteString(" | caused by: ")
		sb.WriteString(e.Cause.Error())
	}
	return sb.String()
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func (e *AppError) WithContext(key string, value interface{}) *AppError {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Context[key] = value
	return e
}

func (e *AppError) GetContext(key string) (interface{}, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	val, ok := e.Context[key]
	return val, ok
}

func (e *AppError) Is(target error) bool {
	if appErr, ok := target.(*AppError); ok {
		return e.Code == appErr.Code
	}
	return false
}

func (e *AppError) IsCode(code ErrorCode) bool {
	return e.Code == code
}

func (e *AppError) GetStack() []byte {
	return e.Stack
}

func (e *AppError) Format() string {
	var sb strings.Builder
	sb.WriteString("Error: ")
	sb.WriteString(e.Error())
	sb.WriteString("\nCode: ")
	sb.WriteString(fmt.Sprintf("%d", e.Code))
	if len(e.Context) > 0 {
		sb.WriteString("\nContext: ")
		for k, v := range e.Context {
			sb.WriteString(fmt.Sprintf("%s=%v; ", k, v))
		}
	}
	return sb.String()
}

func captureStack() []byte {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var buf []byte
	for i := 0; i < n; i++ {
		buf = append(buf, fmt.Sprintf("%s\n", formatFrame(pcs[i]))...)
	}
	return buf
}

func formatFrame(frame uintptr) string {
	fn := runtime.FuncForPC(frame)
	if fn == nil {
		return fmt.Sprintf("frame %d: unknown", frame)
	}
	file, line := fn.FileLine(frame)
	return fmt.Sprintf("%s %s:%d", fn.Name(), file, line)
}

func Is(err, target error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Is(target)
	}
	return err == target
}

func As(err error, target interface{}) bool {
	if appErr, ok := err.(*AppError); ok && appErr.Cause != nil {
		return As(appErr.Cause, target)
	}
	return false
}

func NewBusinessError(code ErrorCode, message string) *AppError {
	return New(code, message)
}

func NewSystemError(message string) *AppError {
	return New(ErrInternalServer, message)
}

func NewValidationError(message string) *AppError {
	return New(ErrInvalidParam, message)
}

func NewNotFoundError(resource string) *AppError {
	return Newf(ErrNotFound, "%s不存在", resource)
}
