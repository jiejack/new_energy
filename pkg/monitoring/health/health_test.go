package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHealthStatus_Constants(t *testing.T) {
	assert.Equal(t, HealthStatus("healthy"), StatusHealthy)
	assert.Equal(t, HealthStatus("unhealthy"), StatusUnhealthy)
	assert.Equal(t, HealthStatus("degraded"), StatusDegraded)
	assert.Equal(t, HealthStatus("unknown"), StatusUnknown)
}

func TestNewHealthCache(t *testing.T) {
	hc := NewHealthCache(5 * time.Second)
	require.NotNil(t, hc)
}

func TestHealthCache_Set_Get(t *testing.T) {
	hc := NewHealthCache(5 * time.Second)
	health := ComponentHealth{
		Name:   "test",
		Status: StatusHealthy,
	}
	hc.Set("test", health)

	got, ok := hc.Get("test")
	assert.True(t, ok)
	assert.Equal(t, StatusHealthy, got.Status)
}

func TestHealthCache_Get_Expired(t *testing.T) {
	hc := NewHealthCache(1 * time.Nanosecond)
	health := ComponentHealth{
		Name:   "test",
		Status: StatusHealthy,
	}
	hc.Set("test", health)
	time.Sleep(10 * time.Millisecond)

	_, ok := hc.Get("test")
	assert.False(t, ok)
}

func TestHealthCache_Get_NotFound(t *testing.T) {
	hc := NewHealthCache(5 * time.Second)
	_, ok := hc.Get("nonexistent")
	assert.False(t, ok)
}

func TestHealthCache_Clear(t *testing.T) {
	hc := NewHealthCache(5 * time.Second)
	hc.Set("test", ComponentHealth{Name: "test", Status: StatusHealthy})
	hc.Clear()
	_, ok := hc.Get("test")
	assert.False(t, ok)
}

func TestNewHealthChecker(t *testing.T) {
	hc := NewHealthChecker(&Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}, zap.NewNop())
	require.NotNil(t, hc)
}

func TestNewHealthChecker_NilLogger(t *testing.T) {
	hc := NewHealthChecker(&Config{
		ServiceName: "test",
	}, nil)
	require.NotNil(t, hc)
}

func TestNewHealthChecker_DefaultConfig(t *testing.T) {
	hc := NewHealthChecker(&Config{}, zap.NewNop())
	require.NotNil(t, hc)
}

func TestHealthChecker_RegisterChecker(t *testing.T) {
	hc := NewHealthChecker(&Config{}, zap.NewNop())
	checker := NewCustomChecker("custom", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "custom", Status: StatusHealthy}
	})
	hc.RegisterChecker(checker)
}

func TestHealthChecker_UnregisterChecker(t *testing.T) {
	hc := NewHealthChecker(&Config{}, zap.NewNop())
	checker := NewCustomChecker("custom", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "custom", Status: StatusHealthy}
	})
	hc.RegisterChecker(checker)
	hc.UnregisterChecker("custom")
}

func TestHealthChecker_Check(t *testing.T) {
	hc := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	hc.RegisterChecker(NewCustomChecker("healthy", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "healthy", Status: StatusHealthy}
	}))

	result := hc.Check(context.Background())
	assert.Equal(t, StatusHealthy, result.Status)
	assert.Equal(t, 1, len(result.Components))
}

func TestHealthChecker_Check_Unhealthy(t *testing.T) {
	hc := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	hc.RegisterChecker(NewCustomChecker("unhealthy", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "unhealthy", Status: StatusUnhealthy}
	}))

	result := hc.Check(context.Background())
	assert.Equal(t, StatusUnhealthy, result.Status)
}

func TestHealthChecker_Check_Degraded(t *testing.T) {
	hc := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	hc.RegisterChecker(NewCustomChecker("degraded", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "degraded", Status: StatusDegraded}
	}))

	result := hc.Check(context.Background())
	assert.Equal(t, StatusDegraded, result.Status)
}

func TestHealthChecker_Check_Multiple(t *testing.T) {
	hc := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	hc.RegisterChecker(NewCustomChecker("healthy1", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "healthy1", Status: StatusHealthy}
	}))
	hc.RegisterChecker(NewCustomChecker("degraded1", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "degraded1", Status: StatusDegraded}
	}))

	result := hc.Check(context.Background())
	assert.Equal(t, StatusDegraded, result.Status)
}

func TestHealthChecker_CheckComponent(t *testing.T) {
	hc := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	hc.RegisterChecker(NewCustomChecker("comp1", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "comp1", Status: StatusHealthy}
	}))

	health, err := hc.CheckComponent(context.Background(), "comp1")
	require.NoError(t, err)
	assert.Equal(t, StatusHealthy, health.Status)
}

func TestHealthChecker_CheckComponent_NotFound(t *testing.T) {
	hc := NewHealthChecker(&Config{}, zap.NewNop())
	_, err := hc.CheckComponent(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestHealthChecker_IsHealthy(t *testing.T) {
	hc := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	hc.RegisterChecker(NewCustomChecker("healthy", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "healthy", Status: StatusHealthy}
	}))
	assert.True(t, hc.IsHealthy(context.Background()))

	hc2 := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	hc2.RegisterChecker(NewCustomChecker("unhealthy", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "unhealthy", Status: StatusUnhealthy}
	}))
	assert.False(t, hc2.IsHealthy(context.Background()))
}

func TestHealthChecker_IsReady(t *testing.T) {
	hc := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	hc.RegisterChecker(NewCustomChecker("healthy", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "healthy", Status: StatusHealthy}
	}))
	assert.True(t, hc.IsReady(context.Background()))

	hc2 := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	hc2.RegisterChecker(NewCustomChecker("unhealthy", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "unhealthy", Status: StatusUnhealthy}
	}))
	assert.False(t, hc2.IsReady(context.Background()))
}

func TestHealthChecker_IsAlive(t *testing.T) {
	hc := NewHealthChecker(&Config{}, zap.NewNop())
	assert.True(t, hc.IsAlive())
}

func TestHealthChecker_ClearCache(t *testing.T) {
	hc := NewHealthChecker(&Config{}, zap.NewNop())
	hc.ClearCache()
}

func TestHealthChecker_HTTPHandler(t *testing.T) {
	hc := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	hc.RegisterChecker(NewCustomChecker("healthy", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "healthy", Status: StatusHealthy}
	}))

	handler := hc.HTTPHandler()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result HealthCheckResult
	err := json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, StatusHealthy, result.Status)
}

func TestHealthChecker_HTTPHandler_Unhealthy(t *testing.T) {
	hc := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	hc.RegisterChecker(NewCustomChecker("unhealthy", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "unhealthy", Status: StatusUnhealthy}
	}))

	handler := hc.HTTPHandler()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestHealthChecker_LivezHandler(t *testing.T) {
	hc := NewHealthChecker(&Config{}, zap.NewNop())
	handler := hc.LivezHandler()

	req := httptest.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestHealthChecker_ReadyzHandler(t *testing.T) {
	hc := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	hc.RegisterChecker(NewCustomChecker("healthy", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "healthy", Status: StatusHealthy}
	}))

	handler := hc.ReadyzHandler()
	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDatabaseChecker_Healthy(t *testing.T) {
	checker := NewDatabaseChecker("db", func(ctx context.Context) error {
		return nil
	})
	assert.Equal(t, "db", checker.Name())

	health := checker.Check(context.Background())
	assert.Equal(t, StatusHealthy, health.Status)
}

func TestDatabaseChecker_Unhealthy(t *testing.T) {
	checker := NewDatabaseChecker("db", func(ctx context.Context) error {
		return errors.New("connection failed")
	})
	health := checker.Check(context.Background())
	assert.Equal(t, StatusUnhealthy, health.Status)
	assert.Contains(t, health.Message, "connection failed")
}

func TestRedisChecker_Healthy(t *testing.T) {
	checker := NewRedisChecker("redis", func(ctx context.Context) error {
		return nil
	})
	assert.Equal(t, "redis", checker.Name())

	health := checker.Check(context.Background())
	assert.Equal(t, StatusHealthy, health.Status)
}

func TestRedisChecker_Unhealthy(t *testing.T) {
	checker := NewRedisChecker("redis", func(ctx context.Context) error {
		return errors.New("connection refused")
	})
	health := checker.Check(context.Background())
	assert.Equal(t, StatusUnhealthy, health.Status)
}

func TestKafkaChecker(t *testing.T) {
	checker := NewKafkaChecker("kafka", []string{"localhost:9092"})
	assert.Equal(t, "kafka", checker.Name())

	health := checker.Check(context.Background())
	assert.Equal(t, StatusHealthy, health.Status)
	assert.Equal(t, []string{"localhost:9092"}, health.Details["brokers"])
}

func TestHTTPChecker_Unreachable(t *testing.T) {
	checker := NewHTTPChecker("http", "http://localhost:19999/health", 1*time.Second)
	assert.Equal(t, "http", checker.Name())

	health := checker.Check(context.Background())
	assert.Equal(t, StatusUnhealthy, health.Status)
}

func TestHTTPChecker_InvalidURL(t *testing.T) {
	checker := NewHTTPChecker("http", "://invalid", 1*time.Second)
	health := checker.Check(context.Background())
	assert.Equal(t, StatusUnhealthy, health.Status)
}

func TestHTTPChecker_Healthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := NewHTTPChecker("http", server.URL, 5*time.Second)
	health := checker.Check(context.Background())
	assert.Equal(t, StatusHealthy, health.Status)
}

func TestHTTPChecker_Degraded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	checker := NewHTTPChecker("http", server.URL, 5*time.Second)
	health := checker.Check(context.Background())
	assert.Equal(t, StatusDegraded, health.Status)
}

func TestHTTPChecker_Unhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	checker := NewHTTPChecker("http", server.URL, 5*time.Second)
	health := checker.Check(context.Background())
	assert.Equal(t, StatusUnhealthy, health.Status)
}

func TestTCPChecker(t *testing.T) {
	checker := NewTCPChecker("tcp", "localhost:8080", 5*time.Second)
	assert.Equal(t, "tcp", checker.Name())

	health := checker.Check(context.Background())
	assert.Equal(t, StatusHealthy, health.Status)
}

func TestCustomChecker(t *testing.T) {
	checker := NewCustomChecker("custom", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "custom", Status: StatusHealthy, Message: "all good"}
	})
	assert.Equal(t, "custom", checker.Name())

	health := checker.Check(context.Background())
	assert.Equal(t, StatusHealthy, health.Status)
	assert.Equal(t, "all good", health.Message)
}

func TestAggregatedHealthChecker_AllHealthy(t *testing.T) {
	checker := NewAggregatedHealthChecker("agg", StrategyAllHealthy,
		NewCustomChecker("c1", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c1", Status: StatusHealthy}
		}),
		NewCustomChecker("c2", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c2", Status: StatusHealthy}
		}),
	)
	assert.Equal(t, "agg", checker.Name())

	health := checker.Check(context.Background())
	assert.Equal(t, StatusHealthy, health.Status)
}

func TestAggregatedHealthChecker_AllHealthy_OneUnhealthy(t *testing.T) {
	checker := NewAggregatedHealthChecker("agg", StrategyAllHealthy,
		NewCustomChecker("c1", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c1", Status: StatusHealthy}
		}),
		NewCustomChecker("c2", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c2", Status: StatusUnhealthy}
		}),
	)

	health := checker.Check(context.Background())
	assert.Equal(t, StatusUnhealthy, health.Status)
}

func TestAggregatedHealthChecker_AllHealthy_Degraded(t *testing.T) {
	checker := NewAggregatedHealthChecker("agg", StrategyAllHealthy,
		NewCustomChecker("c1", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c1", Status: StatusHealthy}
		}),
		NewCustomChecker("c2", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c2", Status: StatusDegraded}
		}),
	)

	health := checker.Check(context.Background())
	assert.Equal(t, StatusDegraded, health.Status)
}

func TestAggregatedHealthChecker_AnyHealthy(t *testing.T) {
	checker := NewAggregatedHealthChecker("agg", StrategyAnyHealthy,
		NewCustomChecker("c1", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c1", Status: StatusUnhealthy}
		}),
		NewCustomChecker("c2", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c2", Status: StatusHealthy}
		}),
	)

	health := checker.Check(context.Background())
	assert.Equal(t, StatusHealthy, health.Status)
}

func TestAggregatedHealthChecker_AnyHealthy_None(t *testing.T) {
	checker := NewAggregatedHealthChecker("agg", StrategyAnyHealthy,
		NewCustomChecker("c1", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c1", Status: StatusUnhealthy}
		}),
		NewCustomChecker("c2", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c2", Status: StatusUnhealthy}
		}),
	)

	health := checker.Check(context.Background())
	assert.Equal(t, StatusUnhealthy, health.Status)
}

func TestAggregatedHealthChecker_MajorityHealthy(t *testing.T) {
	checker := NewAggregatedHealthChecker("agg", StrategyMajorityHealthy,
		NewCustomChecker("c1", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c1", Status: StatusHealthy}
		}),
		NewCustomChecker("c2", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c2", Status: StatusHealthy}
		}),
		NewCustomChecker("c3", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c3", Status: StatusUnhealthy}
		}),
	)

	health := checker.Check(context.Background())
	assert.Equal(t, StatusHealthy, health.Status)
}

func TestAggregatedHealthChecker_MajorityHealthy_Fails(t *testing.T) {
	checker := NewAggregatedHealthChecker("agg", StrategyMajorityHealthy,
		NewCustomChecker("c1", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c1", Status: StatusUnhealthy}
		}),
		NewCustomChecker("c2", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c2", Status: StatusUnhealthy}
		}),
		NewCustomChecker("c3", func(ctx context.Context) ComponentHealth {
			return ComponentHealth{Name: "c3", Status: StatusHealthy}
		}),
	)

	health := checker.Check(context.Background())
	assert.Equal(t, StatusUnhealthy, health.Status)
}

func TestHealthCheckMiddleware(t *testing.T) {
	hc := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	hc.RegisterChecker(NewCustomChecker("healthy", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "healthy", Status: StatusHealthy}
	}))

	middleware := HealthCheckMiddleware(hc)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(next)

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthCheckMiddleware_HealthEndpoint(t *testing.T) {
	hc := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	middleware := HealthCheckMiddleware(hc)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(next)

	paths := []string{"/health", "/healthz", "/ready", "/readyz", "/live", "/livez"}
	for _, path := range paths {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestHealthCheckMiddleware_NotReady(t *testing.T) {
	hc := NewHealthChecker(&Config{CacheTTL: 0}, zap.NewNop())
	hc.RegisterChecker(NewCustomChecker("unhealthy", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{Name: "unhealthy", Status: StatusUnhealthy}
	}))

	middleware := HealthCheckMiddleware(hc)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(next)

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestHealthChecker_StartHTTPServer(t *testing.T) {
	hc := NewHealthChecker(&Config{}, zap.NewNop())
	server, err := hc.StartHTTPServer(0)
	require.NoError(t, err)
	require.NotNil(t, server)
	server.Close()
}

func TestComponentHealth_Struct(t *testing.T) {
	ch := ComponentHealth{
		Name:      "test",
		Status:    StatusHealthy,
		Message:   "all good",
		Timestamp: time.Now(),
		Latency:   10 * time.Millisecond,
		Details:   map[string]interface{}{"key": "value"},
	}
	assert.Equal(t, "test", ch.Name)
	assert.Equal(t, StatusHealthy, ch.Status)
}

func TestHealthCheckResult_Struct(t *testing.T) {
	result := HealthCheckResult{
		Status:     StatusHealthy,
		Timestamp:  time.Now(),
		Components: map[string]ComponentHealth{"db": {Name: "db", Status: StatusHealthy}},
		Version:    "1.0.0",
		Service:    "test-service",
		Uptime:     5 * time.Minute,
	}
	assert.Equal(t, StatusHealthy, result.Status)
	assert.Equal(t, "1.0.0", result.Version)
}
