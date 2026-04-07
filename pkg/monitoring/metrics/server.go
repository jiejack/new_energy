// Package metrics 提供应用监控指标暴露功能
package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Config 指标配置
type Config struct {
	// Port 指标暴露端口
	Port int `yaml:"port" json:"port"`
	// Path 指标路径
	Path string `yaml:"path" json:"path"`
	// EnableGoMetrics 是否启用 Go 运行时指标
	EnableGoMetrics bool `yaml:"enable_go_metrics" json:"enable_go_metrics"`
	// EnableProcessMetrics 是否启用进程指标
	EnableProcessMetrics bool `yaml:"enable_process_metrics" json:"enable_process_metrics"`
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		Port:                 8080,
		Path:                 "/metrics",
		EnableGoMetrics:      true,
		EnableProcessMetrics: true,
	}
}

// Server 指标服务器
type Server struct {
	config Config
	server *http.Server
	logger *zap.Logger

	// 自定义指标
	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	httpRequestsInFlight  *prometheus.GaugeVec
	databaseQueriesTotal  *prometheus.CounterVec
	databaseQueryDuration *prometheus.HistogramVec
	cacheHitsTotal        *prometheus.CounterVec
	cacheMissesTotal      *prometheus.CounterVec
}

// NewServer 创建指标服务器
func NewServer(config Config, logger *zap.Logger) *Server {
	if config.Port == 0 {
		config.Port = 8080
	}
	if config.Path == "" {
		config.Path = "/metrics"
	}

	s := &Server{
		config: config,
		logger: logger,
	}

	// 注册默认指标
	if config.EnableGoMetrics {
		prometheus.MustRegister(prometheus.NewGoCollector())
	}
	if config.EnableProcessMetrics {
		prometheus.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	}

	// 注册自定义指标
	s.registerCustomMetrics()

	return s
}

// registerCustomMetrics 注册自定义指标
func (s *Server) registerCustomMetrics() {
	// HTTP 请求总数
	s.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP 请求总数",
		},
		[]string{"service", "method", "path", "status"},
	)
	prometheus.MustRegister(s.httpRequestsTotal)

	// HTTP 请求持续时间
	s.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP 请求持续时间",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method", "path"},
	)
	prometheus.MustRegister(s.httpRequestDuration)

	// 正在处理的 HTTP 请求
	s.httpRequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "正在处理的 HTTP 请求数",
		},
		[]string{"service", "method"},
	)
	prometheus.MustRegister(s.httpRequestsInFlight)

	// 数据库查询总数
	s.databaseQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "数据库查询总数",
		},
		[]string{"service", "operation", "table"},
	)
	prometheus.MustRegister(s.databaseQueriesTotal)

	// 数据库查询持续时间
	s.databaseQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "数据库查询持续时间",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"service", "operation", "table"},
	)
	prometheus.MustRegister(s.databaseQueryDuration)

	// 缓存命中
	s.cacheHitsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "缓存命中总数",
		},
		[]string{"service", "cache_name"},
	)
	prometheus.MustRegister(s.cacheHitsTotal)

	// 缓存未命中
	s.cacheMissesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "缓存未命中总数",
		},
		[]string{"service", "cache_name"},
	)
	prometheus.MustRegister(s.cacheMissesTotal)
}

// Start 启动指标服务器
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle(s.config.Path, promhttp.Handler())

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 就绪检查端点
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})

	addr := fmt.Sprintf(":%d", s.config.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info("启动指标服务器",
		zap.Int("port", s.config.Port),
		zap.String("path", s.config.Path),
	)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("指标服务器启动失败", zap.Error(err))
		}
	}()

	return nil
}

// Stop 停止指标服务器
func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		s.logger.Info("停止指标服务器")
		return s.server.Shutdown(ctx)
	}
	return nil
}

// RecordHTTPRequest 记录 HTTP 请求
func (s *Server) RecordHTTPRequest(service, method, path, status string, duration time.Duration) {
	s.httpRequestsTotal.WithLabelValues(service, method, path, status).Inc()
	s.httpRequestDuration.WithLabelValues(service, method, path).Observe(duration.Seconds())
}

// IncHTTPRequestInFlight 增加正在处理的请求数
func (s *Server) IncHTTPRequestInFlight(service, method string) {
	s.httpRequestsInFlight.WithLabelValues(service, method).Inc()
}

// DecHTTPRequestInFlight 减少正在处理的请求数
func (s *Server) DecHTTPRequestInFlight(service, method string) {
	s.httpRequestsInFlight.WithLabelValues(service, method).Dec()
}

// RecordDatabaseQuery 记录数据库查询
func (s *Server) RecordDatabaseQuery(service, operation, table string, duration time.Duration) {
	s.databaseQueriesTotal.WithLabelValues(service, operation, table).Inc()
	s.databaseQueryDuration.WithLabelValues(service, operation, table).Observe(duration.Seconds())
}

// RecordCacheHit 记录缓存命中
func (s *Server) RecordCacheHit(service, cacheName string) {
	s.cacheHitsTotal.WithLabelValues(service, cacheName).Inc()
}

// RecordCacheMiss 记录缓存未命中
func (s *Server) RecordCacheMiss(service, cacheName string) {
	s.cacheMissesTotal.WithLabelValues(service, cacheName).Inc()
}

// HTTPMiddleware HTTP 中间件
func (s *Server) HTTPMiddleware(service string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		method := r.Method

		// 增加正在处理的请求数
		s.IncHTTPRequestInFlight(service, method)
		defer s.DecHTTPRequestInFlight(service, method)

		// 包装 ResponseWriter 以获取状态码
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// 调用下一个处理器
		next.ServeHTTP(wrapped, r)

		// 记录指标
		duration := time.Since(start)
		status := fmt.Sprintf("%d", wrapped.statusCode)
		s.RecordHTTPRequest(service, method, path, status, duration)
	})
}

// responseWriter 包装 http.ResponseWriter 以获取状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
