package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/new-energy-monitoring/pkg/monitoring/dashboard"
	"github.com/new-energy-monitoring/pkg/monitoring/health"
	"github.com/new-energy-monitoring/pkg/monitoring/metrics"
	"github.com/new-energy-monitoring/pkg/monitoring/tracing"
	"go.uber.org/zap"
)

// Config 监控配置
type Config struct {
	// 服务信息
	ServiceName    string
	ServiceVersion string
	Environment    string

	// Prometheus配置
	Prometheus PrometheusConfig

	// Tracing配置
	Tracing TracingConfig

	// Health配置
	Health HealthConfig
}

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	Enabled     bool
	Port        int
	Endpoint    string
	Namespace   string
	ConstLabels map[string]string
}

// TracingConfig Tracing配置
type TracingConfig struct {
	Enabled      bool
	Endpoint     string
	Protocol     string // "grpc" or "http"
	SamplerType  string // "always", "never", "ratio", "parentbased"
	SamplerRatio float64
}

// HealthConfig 健康检查配置
type HealthConfig struct {
	Enabled     bool
	Port        int
	Checkers    []health.Checker
}

// Manager 监控管理器
type Manager struct {
	config         *Config
	logger         *zap.Logger
	metricsCollector *metrics.MetricsCollector
	tracerProvider  *tracing.TracerProvider
	healthChecker   *health.HealthChecker
	dashboardManager *dashboard.DashboardManager
	commonMetrics   *metrics.CommonMetrics
	mu             sync.RWMutex
	httpServer     *http.Server
}

// NewManager 创建监控管理器
func NewManager(cfg *Config, logger *zap.Logger) (*Manager, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	m := &Manager{
		config:   cfg,
		logger:   logger,
		dashboardManager: dashboard.NewDashboardManager(),
	}

	// 初始化Prometheus
	if cfg.Prometheus.Enabled {
		mc := metrics.NewMetricsCollector(&metrics.Config{
			Namespace:   cfg.Prometheus.Namespace,
			ConstLabels: cfg.Prometheus.ConstLabels,
		}, logger)

		// 注册常用指标
		cm, err := mc.RegisterCommonMetrics(cfg.Prometheus.Namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to register common metrics: %w", err)
		}

		m.metricsCollector = mc
		m.commonMetrics = cm
	}

	// 初始化Tracing
	if cfg.Tracing.Enabled {
		tp, err := tracing.NewTracerProvider(&tracing.Config{
			ServiceName:    cfg.ServiceName,
			ServiceVersion: cfg.ServiceVersion,
			Environment:    cfg.Environment,
			Endpoint:       cfg.Tracing.Endpoint,
			Protocol:       cfg.Tracing.Protocol,
			SamplerType:    cfg.Tracing.SamplerType,
			SamplerRatio:   cfg.Tracing.SamplerRatio,
			Enabled:        true,
		}, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create tracer provider: %w", err)
		}
		m.tracerProvider = tp
	}

	// 初始化健康检查
	if cfg.Health.Enabled {
		hc := health.NewHealthChecker(&health.Config{
			ServiceName:    cfg.ServiceName,
			ServiceVersion: cfg.ServiceVersion,
		}, logger)

		// 注册健康检查器
		for _, checker := range cfg.Health.Checkers {
			hc.RegisterChecker(checker)
		}

		m.healthChecker = hc
	}

	return m, nil
}

// Start 启动监控服务
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 启动Prometheus指标服务
	if m.metricsCollector != nil {
		port := m.config.Prometheus.Port
		if port == 0 {
			port = 9090
		}
		endpoint := m.config.Prometheus.Endpoint
		if endpoint == "" {
			endpoint = "/metrics"
		}
		if err := m.metricsCollector.Start(port, endpoint); err != nil {
			return fmt.Errorf("failed to start metrics server: %w", err)
		}
		m.logger.Info("prometheus metrics server started", zap.Int("port", port))
	}

	// 启动健康检查服务
	if m.healthChecker != nil {
		port := m.config.Health.Port
		if port == 0 {
			port = 8081
		}
		server, err := m.healthChecker.StartHTTPServer(port)
		if err != nil {
			return fmt.Errorf("failed to start health check server: %w", err)
		}
		m.httpServer = server
		m.logger.Info("health check server started", zap.Int("port", port))
	}

	return nil
}

// Stop 停止监控服务
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error

	// 停止Prometheus服务
	if m.metricsCollector != nil {
		if err := m.metricsCollector.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop metrics server: %w", err))
		}
	}

	// 停止Tracing
	if m.tracerProvider != nil {
		if err := m.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown tracer provider: %w", err))
		}
	}

	// 停止健康检查服务
	if m.httpServer != nil {
		if err := m.httpServer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown health check server: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}

	m.logger.Info("monitoring manager stopped")
	return nil
}

// GetMetricsCollector 获取指标采集器
func (m *Manager) GetMetricsCollector() *metrics.MetricsCollector {
	return m.metricsCollector
}

// GetTracerProvider 获取链路追踪提供者
func (m *Manager) GetTracerProvider() *tracing.TracerProvider {
	return m.tracerProvider
}

// GetHealthChecker 获取健康检查器
func (m *Manager) GetHealthChecker() *health.HealthChecker {
	return m.healthChecker
}

// GetDashboardManager 获取仪表盘管理器
func (m *Manager) GetDashboardManager() *dashboard.DashboardManager {
	return m.dashboardManager
}

// GetCommonMetrics 获取常用指标
func (m *Manager) GetCommonMetrics() *metrics.CommonMetrics {
	return m.commonMetrics
}

// RecordRequest 记录请求指标
func (m *Manager) RecordRequest(service, method, endpoint, status string, duration float64) {
	if m.commonMetrics != nil {
		m.commonMetrics.RequestsTotal.WithLabelValues(service, method, endpoint, status).Inc()
		m.commonMetrics.RequestDuration.WithLabelValues(service, method, endpoint).Observe(duration)
	}
}

// RecordError 记录错误指标
func (m *Manager) RecordError(service, component, errorType string) {
	if m.commonMetrics != nil {
		m.commonMetrics.RecordError(service, component, errorType)
	}
}

// StartSpan 开始一个新的Span
func (m *Manager) StartSpan(ctx context.Context, name string) (context.Context, interface{}) {
	if m.tracerProvider == nil {
		return ctx, nil
	}
	return m.tracerProvider.StartSpan(ctx, tracing.SpanConfig{
		Name: name,
		Kind: 1, // SpanKindInternal
	})
}

// CheckHealth 执行健康检查
func (m *Manager) CheckHealth(ctx context.Context) *health.HealthCheckResult {
	if m.healthChecker == nil {
		return &health.HealthCheckResult{
			Status: health.StatusUnknown,
		}
	}
	return m.healthChecker.Check(ctx)
}

// IsHealthy 检查是否健康
func (m *Manager) IsHealthy(ctx context.Context) bool {
	if m.healthChecker == nil {
		return true
	}
	return m.healthChecker.IsHealthy(ctx)
}

// IsReady 检查是否就绪
func (m *Manager) IsReady(ctx context.Context) bool {
	if m.healthChecker == nil {
		return true
	}
	return m.healthChecker.IsReady(ctx)
}

// RegisterHealthChecker 注册健康检查器
func (m *Manager) RegisterHealthChecker(checker health.Checker) {
	if m.healthChecker != nil {
		m.healthChecker.RegisterChecker(checker)
	}
}

// AddDashboard 添加仪表盘
func (m *Manager) AddDashboard(d *dashboard.Dashboard) {
	m.dashboardManager.AddDashboard(d)
}

// ExportDashboards 导出仪表盘
func (m *Manager) ExportDashboards(dir string) error {
	return m.dashboardManager.ExportAll(dir)
}

// Middleware 监控中间件
func (m *Manager) Middleware(service string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 跳过健康检查端点
			if r.URL.Path == "/health" || r.URL.Path == "/healthz" ||
				r.URL.Path == "/ready" || r.URL.Path == "/readyz" ||
				r.URL.Path == "/live" || r.URL.Path == "/livez" ||
				r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			// 创建Span
			ctx := r.Context()
			if m.tracerProvider != nil {
				ctx, _ = m.tracerProvider.NewHTTPSpan(ctx, r.Method, r.URL.Path)
				r = r.WithContext(ctx)
			}

			// 包装ResponseWriter以捕获状态码
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// 增加在处理请求数
			if m.commonMetrics != nil {
				m.commonMetrics.IncInFlight(service, r.Method, r.URL.Path)
				defer m.commonMetrics.DecInFlight(service, r.Method, r.URL.Path)
			}

			next.ServeHTTP(wrapped, r)

			// 记录状态码
			status := fmt.Sprintf("%d", wrapped.statusCode)
			if m.commonMetrics != nil {
				m.commonMetrics.RequestsTotal.WithLabelValues(service, r.Method, r.URL.Path, status).Inc()
			}
		})
	}
}

// responseWriter 包装http.ResponseWriter以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader 捕获状态码
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// SetupDefaultMonitoring 设置默认监控
func SetupDefaultMonitoring(cfg *Config, logger *zap.Logger) (*Manager, error) {
	return NewManager(cfg, logger)
}

// SetupDefaultDashboards 设置默认仪表盘
func (m *Manager) SetupDefaultDashboards() {
	// 添加新能源监控仪表盘
	m.AddDashboard(dashboard.NewEnergyMonitoringDashboard())

	// 添加系统概览仪表盘
	m.AddDashboard(dashboard.NewSystemOverviewDashboard())

	// 添加告警仪表盘
	m.AddDashboard(dashboard.NewAlertDashboard())
}
