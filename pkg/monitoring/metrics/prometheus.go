package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MetricsCollector 指标采集器
type MetricsCollector struct {
	registry    *prometheus.Registry
	counters    map[string]*prometheus.CounterVec
	gauges      map[string]*prometheus.GaugeVec
	histograms  map[string]*prometheus.HistogramVec
	summaries   map[string]*prometheus.SummaryVec
	constLabels prometheus.Labels
	mu          sync.RWMutex
	logger      *zap.Logger
	server      *http.Server
}

// Config 指标采集器配置
type Config struct {
	Namespace   string
	Subsystem   string
	ConstLabels map[string]string
	Port        int
	Endpoint    string
}

// CounterConfig Counter指标配置
type CounterConfig struct {
	Name   string
	Help   string
	Labels []string
}

// GaugeConfig Gauge指标配置
type GaugeConfig struct {
	Name   string
	Help   string
	Labels []string
}

// HistogramConfig Histogram指标配置
type HistogramConfig struct {
	Name    string
	Help    string
	Labels  []string
	Buckets []float64
}

// SummaryConfig Summary指标配置
type SummaryConfig struct {
	Name       string
	Help       string
	Labels     []string
	Objectives map[float64]float64
	MaxAge     time.Duration
	AgeBuckets uint32
	BufCap     uint32
}

// NewMetricsCollector 创建新的指标采集器
func NewMetricsCollector(cfg *Config, logger *zap.Logger) *MetricsCollector {
	if logger == nil {
		logger = zap.NewNop()
	}

	constLabels := make(prometheus.Labels)
	for k, v := range cfg.ConstLabels {
		constLabels[k] = v
	}

	return &MetricsCollector{
		registry:    prometheus.NewRegistry(),
		counters:    make(map[string]*prometheus.CounterVec),
		gauges:      make(map[string]*prometheus.GaugeVec),
		histograms:  make(map[string]*prometheus.HistogramVec),
		summaries:   make(map[string]*prometheus.SummaryVec),
		constLabels: constLabels,
		logger:      logger,
	}
}

// RegisterCounter 注册Counter指标
func (mc *MetricsCollector) RegisterCounter(cfg CounterConfig) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.counters[cfg.Name]; exists {
		return fmt.Errorf("counter %s already registered", cfg.Name)
	}

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        cfg.Name,
			Help:        cfg.Help,
			ConstLabels: mc.constLabels,
		},
		cfg.Labels,
	)

	if err := mc.registry.Register(counter); err != nil {
		return fmt.Errorf("failed to register counter %s: %w", cfg.Name, err)
	}

	mc.counters[cfg.Name] = counter
	mc.logger.Debug("counter registered", zap.String("name", cfg.Name))
	return nil
}

// RegisterGauge 注册Gauge指标
func (mc *MetricsCollector) RegisterGauge(cfg GaugeConfig) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.gauges[cfg.Name]; exists {
		return fmt.Errorf("gauge %s already registered", cfg.Name)
	}

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        cfg.Name,
			Help:        cfg.Help,
			ConstLabels: mc.constLabels,
		},
		cfg.Labels,
	)

	if err := mc.registry.Register(gauge); err != nil {
		return fmt.Errorf("failed to register gauge %s: %w", cfg.Name, err)
	}

	mc.gauges[cfg.Name] = gauge
	mc.logger.Debug("gauge registered", zap.String("name", cfg.Name))
	return nil
}

// RegisterHistogram 注册Histogram指标
func (mc *MetricsCollector) RegisterHistogram(cfg HistogramConfig) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.histograms[cfg.Name]; exists {
		return fmt.Errorf("histogram %s already registered", cfg.Name)
	}

	buckets := cfg.Buckets
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        cfg.Name,
			Help:        cfg.Help,
			Buckets:     buckets,
			ConstLabels: mc.constLabels,
		},
		cfg.Labels,
	)

	if err := mc.registry.Register(histogram); err != nil {
		return fmt.Errorf("failed to register histogram %s: %w", cfg.Name, err)
	}

	mc.histograms[cfg.Name] = histogram
	mc.logger.Debug("histogram registered", zap.String("name", cfg.Name))
	return nil
}

// RegisterSummary 注册Summary指标
func (mc *MetricsCollector) RegisterSummary(cfg SummaryConfig) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.summaries[cfg.Name]; exists {
		return fmt.Errorf("summary %s already registered", cfg.Name)
	}

	objectives := cfg.Objectives
	if len(objectives) == 0 {
		objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	}

	maxAge := cfg.MaxAge
	if maxAge == 0 {
		maxAge = 10 * time.Minute
	}

	ageBuckets := cfg.AgeBuckets
	if ageBuckets == 0 {
		ageBuckets = 5
	}

	bufCap := cfg.BufCap
	if bufCap == 0 {
		bufCap = 500
	}

	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:        cfg.Name,
			Help:        cfg.Help,
			Objectives:  objectives,
			MaxAge:      maxAge,
			AgeBuckets:  ageBuckets,
			BufCap:      bufCap,
			ConstLabels: mc.constLabels,
		},
		cfg.Labels,
	)

	if err := mc.registry.Register(summary); err != nil {
		return fmt.Errorf("failed to register summary %s: %w", cfg.Name, err)
	}

	mc.summaries[cfg.Name] = summary
	mc.logger.Debug("summary registered", zap.String("name", cfg.Name))
	return nil
}

// IncCounter 增加Counter指标
func (mc *MetricsCollector) IncCounter(name string, labels prometheus.Labels) error {
	mc.mu.RLock()
	counter, exists := mc.counters[name]
	mc.mu.RUnlock()

	if !exists {
		return fmt.Errorf("counter %s not found", name)
	}

	counter.With(labels).Inc()
	return nil
}

// AddCounter 增加Counter指标指定值
func (mc *MetricsCollector) AddCounter(name string, value float64, labels prometheus.Labels) error {
	mc.mu.RLock()
	counter, exists := mc.counters[name]
	mc.mu.RUnlock()

	if !exists {
		return fmt.Errorf("counter %s not found", name)
	}

	counter.With(labels).Add(value)
	return nil
}

// SetGauge 设置Gauge指标
func (mc *MetricsCollector) SetGauge(name string, value float64, labels prometheus.Labels) error {
	mc.mu.RLock()
	gauge, exists := mc.gauges[name]
	mc.mu.RUnlock()

	if !exists {
		return fmt.Errorf("gauge %s not found", name)
	}

	gauge.With(labels).Set(value)
	return nil
}

// IncGauge 增加Gauge指标
func (mc *MetricsCollector) IncGauge(name string, labels prometheus.Labels) error {
	mc.mu.RLock()
	gauge, exists := mc.gauges[name]
	mc.mu.RUnlock()

	if !exists {
		return fmt.Errorf("gauge %s not found", name)
	}

	gauge.With(labels).Inc()
	return nil
}

// DecGauge 减少Gauge指标
func (mc *MetricsCollector) DecGauge(name string, labels prometheus.Labels) error {
	mc.mu.RLock()
	gauge, exists := mc.gauges[name]
	mc.mu.RUnlock()

	if !exists {
		return fmt.Errorf("gauge %s not found", name)
	}

	gauge.With(labels).Dec()
	return nil
}

// AddGauge 增加Gauge指标指定值
func (mc *MetricsCollector) AddGauge(name string, value float64, labels prometheus.Labels) error {
	mc.mu.RLock()
	gauge, exists := mc.gauges[name]
	mc.mu.RUnlock()

	if !exists {
		return fmt.Errorf("gauge %s not found", name)
	}

	gauge.With(labels).Add(value)
	return nil
}

// ObserveHistogram 观察Histogram指标
func (mc *MetricsCollector) ObserveHistogram(name string, value float64, labels prometheus.Labels) error {
	mc.mu.RLock()
	histogram, exists := mc.histograms[name]
	mc.mu.RUnlock()

	if !exists {
		return fmt.Errorf("histogram %s not found", name)
	}

	histogram.With(labels).Observe(value)
	return nil
}

// ObserveSummary 观察Summary指标
func (mc *MetricsCollector) ObserveSummary(name string, value float64, labels prometheus.Labels) error {
	mc.mu.RLock()
	summary, exists := mc.summaries[name]
	mc.mu.RUnlock()

	if !exists {
		return fmt.Errorf("summary %s not found", name)
	}

	summary.With(labels).Observe(value)
	return nil
}

// Timer 计时器，用于测量时间间隔
type Timer struct {
	histogram *prometheus.HistogramVec
	summary   *prometheus.SummaryVec
	labels    prometheus.Labels
	start     time.Time
}

// NewTimer 创建新的计时器
func (mc *MetricsCollector) NewTimer(name string, labels prometheus.Labels) (*Timer, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	timer := &Timer{
		labels: labels,
		start:  time.Now(),
	}

	if histogram, exists := mc.histograms[name]; exists {
		timer.histogram = histogram
		return timer, nil
	}

	if summary, exists := mc.summaries[name]; exists {
		timer.summary = summary
		return timer, nil
	}

	return nil, fmt.Errorf("histogram or summary %s not found", name)
}

// Record 记录时间间隔
func (t *Timer) Record() {
	duration := time.Since(t.start).Seconds()
	if t.histogram != nil {
		t.histogram.With(t.labels).Observe(duration)
	} else if t.summary != nil {
		t.summary.With(t.labels).Observe(duration)
	}
}

// RecordDuration 记录指定时间间隔
func (t *Timer) RecordDuration(duration time.Duration) {
	seconds := duration.Seconds()
	if t.histogram != nil {
		t.histogram.With(t.labels).Observe(seconds)
	} else if t.summary != nil {
		t.summary.With(t.labels).Observe(seconds)
	}
}

// Start 启动指标导出HTTP服务
func (mc *MetricsCollector) Start(port int, endpoint string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.server != nil {
		return fmt.Errorf("metrics server already running")
	}

	mux := http.NewServeMux()
	mux.Handle(endpoint, promhttp.HandlerFor(mc.registry, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	}))

	mc.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		mc.logger.Info("starting metrics server", zap.Int("port", port), zap.String("endpoint", endpoint))
		if err := mc.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mc.logger.Error("metrics server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop 停止指标导出HTTP服务
func (mc *MetricsCollector) Stop(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.server == nil {
		return nil
	}

	if err := mc.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown metrics server: %w", err)
	}

	mc.server = nil
	mc.logger.Info("metrics server stopped")
	return nil
}

// GetRegistry 获取Prometheus注册表
func (mc *MetricsCollector) GetRegistry() *prometheus.Registry {
	return mc.registry
}

// RegisterCustomCollector 注册自定义采集器
func (mc *MetricsCollector) RegisterCustomCollector(collector prometheus.Collector) error {
	return mc.registry.Register(collector)
}

// UnregisterCustomCollector 注销自定义采集器
func (mc *MetricsCollector) UnregisterCustomCollector(collector prometheus.Collector) bool {
	return mc.registry.Unregister(collector)
}

// DefaultBuckets 默认Histogram buckets
var DefaultBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// DefaultObjectives 默认Summary objectives
var DefaultObjectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.01, 0.99: 0.001}

// NewDefaultHistogramConfig 创建默认Histogram配置
func NewDefaultHistogramConfig(name, help string, labels []string) HistogramConfig {
	return HistogramConfig{
		Name:    name,
		Help:    help,
		Labels:  labels,
		Buckets: DefaultBuckets,
	}
}

// NewDefaultSummaryConfig 创建默认Summary配置
func NewDefaultSummaryConfig(name, help string, labels []string) SummaryConfig {
	return SummaryConfig{
		Name:       name,
		Help:       help,
		Labels:     labels,
		Objectives: DefaultObjectives,
		MaxAge:     10 * time.Minute,
		AgeBuckets: 5,
		BufCap:     500,
	}
}

// MetricLabels 常用指标标签
type MetricLabels struct {
	Service   string
	Method    string
	Endpoint  string
	Status    string
	Component string
	Station   string
	Device    string
	Point     string
}

// ToPrometheusLabels 转换为Prometheus标签
func (ml *MetricLabels) ToPrometheusLabels() prometheus.Labels {
	labels := make(prometheus.Labels)
	if ml.Service != "" {
		labels["service"] = ml.Service
	}
	if ml.Method != "" {
		labels["method"] = ml.Method
	}
	if ml.Endpoint != "" {
		labels["endpoint"] = ml.Endpoint
	}
	if ml.Status != "" {
		labels["status"] = ml.Status
	}
	if ml.Component != "" {
		labels["component"] = ml.Component
	}
	if ml.Station != "" {
		labels["station"] = ml.Station
	}
	if ml.Device != "" {
		labels["device"] = ml.Device
	}
	if ml.Point != "" {
		labels["point"] = ml.Point
	}
	return labels
}

// CommonMetrics 常用指标
type CommonMetrics struct {
	RequestsTotal      *prometheus.CounterVec
	RequestDuration    *prometheus.HistogramVec
	RequestsInFlight   *prometheus.GaugeVec
	ErrorsTotal        *prometheus.CounterVec
	DataProcessedBytes *prometheus.CounterVec
}

// RegisterCommonMetrics 注册常用指标
func (mc *MetricsCollector) RegisterCommonMetrics(namespace string) (*CommonMetrics, error) {
	cm := &CommonMetrics{}

	// 请求总数
	cm.RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Name:        "requests_total",
			Help:        "Total number of requests",
			ConstLabels: mc.constLabels,
		},
		[]string{"service", "method", "endpoint", "status"},
	)
	if err := mc.registry.Register(cm.RequestsTotal); err != nil {
		return nil, fmt.Errorf("failed to register requests_total: %w", err)
	}

	// 请求延迟
	cm.RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   namespace,
			Name:        "request_duration_seconds",
			Help:        "Request duration in seconds",
			Buckets:     DefaultBuckets,
			ConstLabels: mc.constLabels,
		},
		[]string{"service", "method", "endpoint"},
	)
	if err := mc.registry.Register(cm.RequestDuration); err != nil {
		return nil, fmt.Errorf("failed to register request_duration_seconds: %w", err)
	}

	// 在处理请求数
	cm.RequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "requests_in_flight",
			Help:        "Number of requests currently being processed",
			ConstLabels: mc.constLabels,
		},
		[]string{"service", "method", "endpoint"},
	)
	if err := mc.registry.Register(cm.RequestsInFlight); err != nil {
		return nil, fmt.Errorf("failed to register requests_in_flight: %w", err)
	}

	// 错误总数
	cm.ErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Name:        "errors_total",
			Help:        "Total number of errors",
			ConstLabels: mc.constLabels,
		},
		[]string{"service", "component", "error_type"},
	)
	if err := mc.registry.Register(cm.ErrorsTotal); err != nil {
		return nil, fmt.Errorf("failed to register errors_total: %w", err)
	}

	// 处理数据量
	cm.DataProcessedBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Name:        "data_processed_bytes_total",
			Help:        "Total bytes of data processed",
			ConstLabels: mc.constLabels,
		},
		[]string{"service", "component"},
	)
	if err := mc.registry.Register(cm.DataProcessedBytes); err != nil {
		return nil, fmt.Errorf("failed to register data_processed_bytes_total: %w", err)
	}

	return cm, nil
}

// RecordRequest 记录请求指标
func (cm *CommonMetrics) RecordRequest(service, method, endpoint, status string, duration time.Duration) {
	cm.RequestsTotal.WithLabelValues(service, method, endpoint, status).Inc()
	cm.RequestDuration.WithLabelValues(service, method, endpoint).Observe(duration.Seconds())
}

// IncInFlight 增加在处理请求数
func (cm *CommonMetrics) IncInFlight(service, method, endpoint string) {
	cm.RequestsInFlight.WithLabelValues(service, method, endpoint).Inc()
}

// DecInFlight 减少在处理请求数
func (cm *CommonMetrics) DecInFlight(service, method, endpoint string) {
	cm.RequestsInFlight.WithLabelValues(service, method, endpoint).Dec()
}

// RecordError 记录错误
func (cm *CommonMetrics) RecordError(service, component, errorType string) {
	cm.ErrorsTotal.WithLabelValues(service, component, errorType).Inc()
}

// RecordDataProcessed 记录处理数据量
func (cm *CommonMetrics) RecordDataProcessed(service, component string, bytes float64) {
	cm.DataProcessedBytes.WithLabelValues(service, component).Add(bytes)
}
