package errors

import (
	"testing"
)

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		code    ErrorCode
		message string
		isBiz   bool
		isSrv   bool
	}{
		{ErrSuccess, "操作成功", false, false},
		{ErrUnknown, "未知错误", false, false},
		{ErrInvalidParam, "无效参数", true, false},
		{ErrUnauthorized, "未授权", true, false},
		{ErrInternalServer, "服务器内部错误", false, true},
		{ErrUserNotFound, "用户不存在", true, false},
		{ErrDeviceNotFound, "设备不存在", true, false},
		{ErrDatabaseError, "数据库错误", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.code.String(), func(t *testing.T) {
			if tt.code.String() != tt.message {
				t.Errorf("expected %s, got %s", tt.message, tt.code.String())
			}
			if tt.code.IsBusinessError() != tt.isBiz {
				t.Errorf("IsBusinessError() for %d = %v, expected %v", tt.code, tt.code.IsBusinessError(), tt.isBiz)
			}
			if tt.code.IsServerError() != tt.isSrv {
				t.Errorf("IsServerError() for %d = %v, expected %v", tt.code, tt.code.IsServerError(), tt.isSrv)
			}
		})
	}
}

func TestAppError(t *testing.T) {
	t.Run("New error", func(t *testing.T) {
		err := New(ErrUserNotFound, "用户不存在")
		if err.Code != ErrUserNotFound {
			t.Errorf("expected code %d, got %d", ErrUserNotFound, err.Code)
		}
		if err.Message != "用户不存在" {
			t.Errorf("expected message '用户不存在', got '%s'", err.Message)
		}
		if err.Cause != nil {
			t.Error("expected nil cause")
		}
		if err.Stack == nil {
			t.Error("expected non-nil stack")
		}
	})

	t.Run("Wrap error", func(t *testing.T) {
		cause := New(ErrDatabaseError, "连接失败")
		err := Wrap(cause, ErrInternalServer, "服务不可用")
		if err.Code != ErrInternalServer {
			t.Errorf("expected code %d, got %d", ErrInternalServer, err.Code)
		}
		if err.Cause != cause {
			t.Error("expected cause to be wrapped error")
		}
	})

	t.Run("WithContext", func(t *testing.T) {
		err := New(ErrInvalidParam, "参数错误").
			WithContext("field", "username").
			WithContext("reason", "too short")
		if val, ok := err.GetContext("field"); !ok || val != "username" {
			t.Errorf("expected field=username, got %v,%v", val, ok)
		}
		if val, ok := err.GetContext("reason"); !ok || val != "too short" {
			t.Errorf("expected reason=too short, got %v,%v", val, ok)
		}
	})

	t.Run("Error string", func(t *testing.T) {
		err := New(ErrUserNotFound, "用户不存在")
		expected := "用户不存在"
		if err.Error() != expected {
			t.Errorf("expected '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("Is error code", func(t *testing.T) {
		err := New(ErrUserNotFound, "用户不存在")
		if !err.IsCode(ErrUserNotFound) {
			t.Error("expected IsCode(ErrUserNotFound) to be true")
		}
		if err.IsCode(ErrDeviceNotFound) {
			t.Error("expected IsCode(ErrDeviceNotFound) to be false")
		}
	})
}

func TestErrorEquality(t *testing.T) {
	err1 := New(ErrUserNotFound, "用户不存在")
	err2 := New(ErrUserNotFound, "用户不存在")
	err3 := New(ErrDeviceNotFound, "设备不存在")

	if !err1.Is(err2) {
		t.Error("expected same error codes to be equal")
	}
	if err1.Is(err3) {
		t.Error("expected different error codes to be not equal")
	}
}

func TestErrorFormats(t *testing.T) {
	t.Run("Newf format", func(t *testing.T) {
		err := Newf(ErrNotFound, "%s不存在", "用户")
		if err.Message != "用户不存在" {
			t.Errorf("expected '用户不存在', got '%s'", err.Message)
		}
	})

	t.Run("Wrapf format", func(t *testing.T) {
		cause := New(ErrDatabaseError, "连接失败")
		err := Wrapf(cause, ErrInternalServer, "处理%s失败", "请求")
		if err.Message != "处理请求失败" {
			t.Errorf("expected '处理请求失败', got '%s'", err.Message)
		}
	})
}

func BenchmarkErrorCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(ErrUserNotFound, "用户不存在")
	}
}

func BenchmarkErrorWrap(b *testing.B) {
	cause := New(ErrDatabaseError, "连接失败")
	for i := 0; i < b.N; i++ {
		Wrap(cause, ErrInternalServer, "服务不可用")
	}
}
