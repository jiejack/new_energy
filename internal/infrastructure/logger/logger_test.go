package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInit_Defaults(t *testing.T) {
	assert.NotNil(t, Log)
	assert.NotNil(t, Sugar)
}

func TestInit_JSON(t *testing.T) {
	err := Init(&Config{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	})
	require.NoError(t, err)
	assert.NotNil(t, Log)
	assert.NotNil(t, Sugar)
}

func TestInit_Console(t *testing.T) {
	err := Init(&Config{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	})
	require.NoError(t, err)
}

func TestInit_WarnLevel(t *testing.T) {
	err := Init(&Config{
		Level:  "warn",
		Format: "json",
		Output: "stdout",
	})
	require.NoError(t, err)
}

func TestInit_ErrorLevel(t *testing.T) {
	err := Init(&Config{
		Level:  "error",
		Format: "json",
		Output: "stdout",
	})
	require.NoError(t, err)
}

func TestInit_Stderr(t *testing.T) {
	err := Init(&Config{
		Level:  "info",
		Format: "json",
		Output: "stderr",
	})
	require.NoError(t, err)
}

func TestInit_InvalidFile(t *testing.T) {
	err := Init(&Config{
		Level:  "info",
		Format: "json",
		Output: "/nonexistent/path/to/log/file.log",
	})
	require.NoError(t, err)
}

func TestSync(t *testing.T) {
	Init(&Config{Level: "info", Format: "json", Output: "stdout"})
	Sync()
}

func TestDebug(t *testing.T) {
	Init(&Config{Level: "debug", Format: "json", Output: "stdout"})
	Debug("test debug message")
}

func TestInfo(t *testing.T) {
	Init(&Config{Level: "info", Format: "json", Output: "stdout"})
	Info("test info message")
}

func TestWarn(t *testing.T) {
	Init(&Config{Level: "warn", Format: "json", Output: "stdout"})
	Warn("test warn message")
}

func TestError(t *testing.T) {
	Init(&Config{Level: "error", Format: "json", Output: "stdout"})
	Error("test error message")
}

func TestWith(t *testing.T) {
	Init(&Config{Level: "info", Format: "json", Output: "stdout"})
	l := With(zap.String("key", "value"))
	assert.NotNil(t, l)
}

func TestNamed(t *testing.T) {
	Init(&Config{Level: "info", Format: "json", Output: "stdout"})
	l := Named("test-logger")
	assert.NotNil(t, l)
}

func TestConfig_Struct(t *testing.T) {
	cfg := Config{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}
	assert.Equal(t, "debug", cfg.Level)
	assert.Equal(t, "json", cfg.Format)
	assert.Equal(t, "stdout", cfg.Output)
}
