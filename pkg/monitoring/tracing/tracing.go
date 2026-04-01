package tracing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TracerProvider 链路追踪提供者
type TracerProvider struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
	config   *Config
	logger   *zap.Logger
	mu       sync.RWMutex
}

// Config 链路追踪配置
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string
	Protocol       string // "grpc" or "http"
	SamplerType    string // "always", "never", "ratio", "parentbased"
	SamplerRatio   float64
	BatchTimeout   time.Duration
	ExportTimeout  time.Duration
	MaxExportBatch int
	MaxQueueSize   int
	Enabled        bool
}

// SpanConfig Span配置
type SpanConfig struct {
	Name       string
	Attributes []attribute.KeyValue
	Kind       trace.SpanKind
}

// NewTracerProvider 创建新的链路追踪提供者
func NewTracerProvider(cfg *Config, logger *zap.Logger) (*TracerProvider, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	if !cfg.Enabled {
		logger.Info("tracing is disabled, using no-op tracer")
		return &TracerProvider{
			provider: nil,
			tracer:   trace.NoopTracerProvider{}.Tracer(cfg.ServiceName),
			config:   cfg,
			logger:   logger,
		}, nil
	}

	// 创建资源
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 创建导出器
	exporter, err := createExporter(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// 创建采样器
	sampler := createSampler(cfg)

	// 创建TracerProvider
	batchTimeout := cfg.BatchTimeout
	if batchTimeout == 0 {
		batchTimeout = 5 * time.Second
	}

	exportTimeout := cfg.ExportTimeout
	if exportTimeout == 0 {
		exportTimeout = 30 * time.Second
	}

	maxExportBatch := cfg.MaxExportBatch
	if maxExportBatch == 0 {
		maxExportBatch = 512
	}

	maxQueueSize := cfg.MaxQueueSize
	if maxQueueSize == 0 {
		maxQueueSize = 2048
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(batchTimeout),
			sdktrace.WithExportTimeout(exportTimeout),
			sdktrace.WithMaxExportBatchSize(maxExportBatch),
			sdktrace.WithMaxQueueSize(maxQueueSize),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// 设置全局TracerProvider
	otel.SetTracerProvider(provider)

	// 设置全局传播器
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := provider.Tracer(cfg.ServiceName, trace.WithInstrumentationVersion(cfg.ServiceVersion))

	logger.Info("tracer provider created",
		zap.String("service", cfg.ServiceName),
		zap.String("endpoint", cfg.Endpoint),
		zap.String("sampler", cfg.SamplerType),
	)

	return &TracerProvider{
		provider: provider,
		tracer:   tracer,
		config:   cfg,
		logger:   logger,
	}, nil
}

// createExporter 创建导出器
func createExporter(cfg *Config) (sdktrace.SpanExporter, error) {
	var client otlptrace.Client

	switch cfg.Protocol {
	case "grpc":
		client = otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
			otlptracegrpc.WithInsecure(),
		)
	case "http":
		client = otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(cfg.Endpoint),
			otlptracehttp.WithInsecure(),
		)
	default:
		// 默认使用gRPC
		client = otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
			otlptracegrpc.WithInsecure(),
		)
	}

	return otlptrace.New(context.Background(), client)
}

// createSampler 创建采样器
func createSampler(cfg *Config) sdktrace.Sampler {
	switch cfg.SamplerType {
	case "always":
		return sdktrace.AlwaysSample()
	case "never":
		return sdktrace.NeverSample()
	case "ratio":
		return sdktrace.TraceIDRatioBased(cfg.SamplerRatio)
	case "parentbased":
		return sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(cfg.SamplerRatio),
			sdktrace.WithLocalParentSampled(sdktrace.AlwaysSample()),
			sdktrace.WithLocalParentNotSampled(sdktrace.NeverSample()),
			sdktrace.WithRemoteParentSampled(sdktrace.AlwaysSample()),
			sdktrace.WithRemoteParentNotSampled(sdktrace.NeverSample()),
		)
	default:
		// 默认使用parentbased
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	}
}

// StartSpan 开始一个新的Span
func (tp *TracerProvider) StartSpan(ctx context.Context, cfg SpanConfig) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{
		trace.WithAttributes(cfg.Attributes...),
		trace.WithSpanKind(cfg.Kind),
	}
	return tp.tracer.Start(ctx, cfg.Name, opts...)
}

// StartSpanFromContext 从上下文开始Span
func (tp *TracerProvider) StartSpanFromContext(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tp.tracer.Start(ctx, name, opts...)
}

// GetTracer 获取Tracer
func (tp *TracerProvider) GetTracer() trace.Tracer {
	return tp.tracer
}

// Shutdown 关闭TracerProvider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if tp.provider == nil {
		return nil
	}

	if err := tp.provider.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown tracer provider: %w", err)
	}

	tp.logger.Info("tracer provider shutdown completed")
	return nil
}

// ForceFlush 强制刷新
func (tp *TracerProvider) ForceFlush(ctx context.Context) error {
	if tp.provider == nil {
		return nil
	}
	return tp.provider.ForceFlush(ctx)
}

// SpanBuilder Span构建器
type SpanBuilder struct {
	name       string
	attributes []attribute.KeyValue
	events     []Event
	kind       trace.SpanKind
	links      []Link
}

// Event Span事件
type Event struct {
	Name       string
	Attributes []attribute.KeyValue
	Timestamp  time.Time
}

// Link Span链接
type Link struct {
	SpanContext trace.SpanContext
	Attributes  []attribute.KeyValue
}

// NewSpanBuilder 创建Span构建器
func NewSpanBuilder(name string) *SpanBuilder {
	return &SpanBuilder{
		name:       name,
		attributes: make([]attribute.KeyValue, 0),
		events:     make([]Event, 0),
		links:      make([]Link, 0),
		kind:       trace.SpanKindInternal,
	}
}

// WithAttribute 添加属性
func (sb *SpanBuilder) WithAttribute(key string, value interface{}) *SpanBuilder {
	sb.attributes = append(sb.attributes, createAttribute(key, value))
	return sb
}

// WithAttributes 添加多个属性
func (sb *SpanBuilder) WithAttributes(attrs map[string]interface{}) *SpanBuilder {
	for k, v := range attrs {
		sb.attributes = append(sb.attributes, createAttribute(k, v))
	}
	return sb
}

// WithKind 设置Span类型
func (sb *SpanBuilder) WithKind(kind trace.SpanKind) *SpanBuilder {
	sb.kind = kind
	return sb
}

// WithEvent 添加事件
func (sb *SpanBuilder) WithEvent(name string, attrs ...attribute.KeyValue) *SpanBuilder {
	sb.events = append(sb.events, Event{
		Name:       name,
		Attributes: attrs,
		Timestamp:  time.Now(),
	})
	return sb
}

// WithLink 添加链接
func (sb *SpanBuilder) WithLink(sc trace.SpanContext, attrs ...attribute.KeyValue) *SpanBuilder {
	sb.links = append(sb.links, Link{
		SpanContext: sc,
		Attributes:  attrs,
	})
	return sb
}

// Build 构建并启动Span
func (sb *SpanBuilder) Build(ctx context.Context, tp *TracerProvider) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{
		trace.WithAttributes(sb.attributes...),
		trace.WithSpanKind(sb.kind),
	}

	for _, link := range sb.links {
		opts = append(opts, trace.WithLink(link.SpanContext, link.Attributes...))
	}

	ctx, span := tp.tracer.Start(ctx, sb.name, opts...)

	// 添加事件
	for _, event := range sb.events {
		span.AddEvent(event.Name, trace.WithAttributes(event.Attributes...), trace.WithTimestamp(event.Timestamp))
	}

	return ctx, span
}

// createAttribute 创建属性
func createAttribute(key string, value interface{}) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int64(key, int64(v))
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	case []string:
		return attribute.StringSlice(key, v)
	case []int64:
		return attribute.Int64Slice(key, v)
	case []float64:
		return attribute.Float64Slice(key, v)
	case []bool:
		return attribute.BoolSlice(key, v)
	default:
		return attribute.String(key, fmt.Sprintf("%v", v))
	}
}

// SpanContextExtractor Span上下文提取器
type SpanContextExtractor struct {
	propagator propagation.TextMapPropagator
}

// NewSpanContextExtractor 创建Span上下文提取器
func NewSpanContextExtractor() *SpanContextExtractor {
	return &SpanContextExtractor{
		propagator: otel.GetTextMapPropagator(),
	}
}

// Extract 从carrier提取Span上下文
func (sce *SpanContextExtractor) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return sce.propagator.Extract(ctx, carrier)
}

// Inject 将Span上下文注入到carrier
func (sce *SpanContextExtractor) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sce.propagator.Inject(ctx, carrier)
}

// GetTraceID 获取TraceID
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID 获取SpanID
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// SetSpanError 设置Span错误
func SetSpanError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetAttributes(
		attribute.Bool("error", true),
		attribute.String("error.message", err.Error()),
	)
}

// SetSpanErrorWithStack 设置Span错误（带堆栈）
func SetSpanErrorWithStack(span trace.Span, err error, stack string) {
	if err == nil {
		return
	}
	span.RecordError(err, trace.WithAttributes(
		attribute.String("stack", stack),
	))
	span.SetAttributes(
		attribute.Bool("error", true),
		attribute.String("error.message", err.Error()),
		attribute.String("error.stack", stack),
	)
}

// SpanTimer Span计时器
type SpanTimer struct {
	span     trace.Span
	start    time.Time
	name     string
	attrs    []attribute.KeyValue
}

// NewSpanTimer 创建Span计时器
func NewSpanTimer(span trace.Span, name string, attrs ...attribute.KeyValue) *SpanTimer {
	return &SpanTimer{
		span:  span,
		start: time.Now(),
		name:  name,
		attrs: attrs,
	}
}

// Stop 停止计时并记录事件
func (st *SpanTimer) Stop() {
	duration := time.Since(st.start)
	attrs := append(st.attrs, attribute.Int64("duration_ms", duration.Milliseconds()))
	st.span.AddEvent(st.name, trace.WithAttributes(attrs...))
}

// Duration 获取持续时间
func (st *SpanTimer) Duration() time.Duration {
	return time.Since(st.start)
}

// CommonAttributes 常用属性
var CommonAttributes = struct {
	ServiceName    attribute.Key
	ServiceVersion attribute.Key
	Environment    attribute.Key
	StationID      attribute.Key
	DeviceID       attribute.Key
	PointID        attribute.Key
	UserID         attribute.Key
	RequestID      attribute.Key
	Method         attribute.Key
	Endpoint       attribute.Key
	StatusCode     attribute.Key
	ErrorType      attribute.Key
}{
	ServiceName:    attribute.Key("service.name"),
	ServiceVersion: attribute.Key("service.version"),
	Environment:    attribute.Key("environment"),
	StationID:      attribute.Key("station.id"),
	DeviceID:       attribute.Key("device.id"),
	PointID:        attribute.Key("point.id"),
	UserID:         attribute.Key("user.id"),
	RequestID:      attribute.Key("request.id"),
	Method:         attribute.Key("http.method"),
	Endpoint:       attribute.Key("http.endpoint"),
	StatusCode:     attribute.Key("http.status_code"),
	ErrorType:      attribute.Key("error.type"),
}

// NewStationSpan 创建站点相关Span
func (tp *TracerProvider) NewStationSpan(ctx context.Context, name, stationID string) (context.Context, trace.Span) {
	return tp.StartSpan(ctx, SpanConfig{
		Name: name,
		Attributes: []attribute.KeyValue{
			CommonAttributes.StationID.String(stationID),
		},
		Kind: trace.SpanKindInternal,
	})
}

// NewDeviceSpan 创建设备相关Span
func (tp *TracerProvider) NewDeviceSpan(ctx context.Context, name, stationID, deviceID string) (context.Context, trace.Span) {
	return tp.StartSpan(ctx, SpanConfig{
		Name: name,
		Attributes: []attribute.KeyValue{
			CommonAttributes.StationID.String(stationID),
			CommonAttributes.DeviceID.String(deviceID),
		},
		Kind: trace.SpanKindInternal,
	})
}

// NewPointSpan 创建测点相关Span
func (tp *TracerProvider) NewPointSpan(ctx context.Context, name, stationID, deviceID, pointID string) (context.Context, trace.Span) {
	return tp.StartSpan(ctx, SpanConfig{
		Name: name,
		Attributes: []attribute.KeyValue{
			CommonAttributes.StationID.String(stationID),
			CommonAttributes.DeviceID.String(deviceID),
			CommonAttributes.PointID.String(pointID),
		},
		Kind: trace.SpanKindInternal,
	})
}

// NewHTTPSpan 创建HTTP请求Span
func (tp *TracerProvider) NewHTTPSpan(ctx context.Context, method, endpoint string) (context.Context, trace.Span) {
	return tp.StartSpan(ctx, SpanConfig{
		Name: fmt.Sprintf("%s %s", method, endpoint),
		Attributes: []attribute.KeyValue{
			CommonAttributes.Method.String(method),
			CommonAttributes.Endpoint.String(endpoint),
		},
		Kind: trace.SpanKindServer,
	})
}

// SetHTTPStatus 设置HTTP状态
func SetHTTPStatus(span trace.Span, statusCode int) {
	span.SetAttributes(CommonAttributes.StatusCode.Int(statusCode))
}

// PropagationCarrier 传播载体（用于HTTP Headers等）
type PropagationCarrier struct {
	headers map[string]string
}

// NewPropagationCarrier 创建传播载体
func NewPropagationCarrier() *PropagationCarrier {
	return &PropagationCarrier{
		headers: make(map[string]string),
	}
}

// Get 获取值
func (pc *PropagationCarrier) Get(key string) string {
	return pc.headers[key]
}

// Set 设置值
func (pc *PropagationCarrier) Set(key, value string) {
	pc.headers[key] = value
}

// Keys 获取所有键
func (pc *PropagationCarrier) Keys() []string {
	keys := make([]string, 0, len(pc.headers))
	for k := range pc.headers {
		keys = append(keys, k)
	}
	return keys
}

// ToMap 转换为map
func (pc *PropagationCarrier) ToMap() map[string]string {
	result := make(map[string]string)
	for k, v := range pc.headers {
		result[k] = v
	}
	return result
}

// FromMap 从map创建
func PropagationCarrierFromMap(m map[string]string) *PropagationCarrier {
	pc := NewPropagationCarrier()
	for k, v := range m {
		pc.Set(k, v)
	}
	return pc
}
