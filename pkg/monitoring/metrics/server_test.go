package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDefaultServerConfig(t *testing.T) {
	cfg := DefaultServerConfig()
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "/metrics", cfg.Path)
	assert.True(t, cfg.EnableGoMetrics)
	assert.True(t, cfg.EnableProcessMetrics)
}

func TestServerConfig_ZeroPort(t *testing.T) {
	cfg := ServerConfig{Port: 0, Path: ""}
	s := NewServer(cfg, zap.NewNop())
	require.NotNil(t, s)
	assert.Equal(t, 8080, s.config.Port)
	assert.Equal(t, "/metrics", s.config.Path)
}

func TestResponseWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	rw.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, rw.statusCode)
}
