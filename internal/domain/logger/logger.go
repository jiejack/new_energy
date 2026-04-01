package logger

import (
	"go.uber.org/zap"
)

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) *zap.Logger
	Named(name string) *zap.Logger
}

// 确保 zap.Logger 实现了 Logger 接口
var _ Logger = (*zap.Logger)(nil)
