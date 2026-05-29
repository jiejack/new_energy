package tracing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func TestCov_NewTracerProvider_Disabled(t *testing.T) {
	cfg := &Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Enabled:        false,
	}
	tp, err := NewTracerProvider(cfg, nil)
	require.NoError(t, err)
	require.NotNil(t, tp)
	assert.Nil(t, tp.provider)
	assert.NotNil(t, tp.tracer)
}

func TestCov_NewTracerProvider_WithLogger(t *testing.T) {
	logger := zap.NewNop()
	cfg := &Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Enabled:        false,
	}
	tp, err := NewTracerProvider(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, tp)
}

func TestCov_TracerProvider_Shutdown_NilProvider(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	err := tp.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestCov_TracerProvider_ForceFlush_NilProvider(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	err := tp.ForceFlush(context.Background())
	assert.NoError(t, err)
}

func TestCov_TracerProvider_StartSpan(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	ctx, span := tp.StartSpan(context.Background(), SpanConfig{
		Name: "test-span",
		Attributes: []attribute.KeyValue{
			attribute.String("key", "value"),
		},
		Kind: trace.SpanKindInternal,
	})
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestCov_TracerProvider_StartSpanFromContext(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	ctx, span := tp.StartSpanFromContext(context.Background(), "child-span")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestCov_TracerProvider_GetTracer(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	tracer := tp.GetTracer()
	assert.NotNil(t, tracer)
}

func TestCov_TracerProvider_NewStationSpan(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	ctx, span := tp.NewStationSpan(context.Background(), "station-op", "station-001")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestCov_TracerProvider_NewDeviceSpan(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	ctx, span := tp.NewDeviceSpan(context.Background(), "device-op", "station-001", "device-001")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestCov_TracerProvider_NewPointSpan(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	ctx, span := tp.NewPointSpan(context.Background(), "point-op", "station-001", "device-001", "point-001")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestCov_TracerProvider_NewHTTPSpan(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	ctx, span := tp.NewHTTPSpan(context.Background(), "GET", "/api/v1/stations")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestCov_SpanBuilder_WithAttribute(t *testing.T) {
	sb := NewSpanBuilder("test-span")
	result := sb.WithAttribute("key1", "value1")
	assert.Equal(t, sb, result)
	assert.Len(t, sb.attributes, 1)
}

func TestCov_SpanBuilder_WithAttributes(t *testing.T) {
	sb := NewSpanBuilder("test-span")
	result := sb.WithAttributes(map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})
	assert.Equal(t, sb, result)
	assert.Len(t, sb.attributes, 2)
}

func TestCov_SpanBuilder_WithKind(t *testing.T) {
	sb := NewSpanBuilder("test-span")
	result := sb.WithKind(trace.SpanKindClient)
	assert.Equal(t, sb, result)
	assert.Equal(t, trace.SpanKindClient, sb.kind)
}

func TestCov_SpanBuilder_WithEvent(t *testing.T) {
	sb := NewSpanBuilder("test-span")
	result := sb.WithEvent("event1", attribute.String("ekey", "eval"))
	assert.Equal(t, sb, result)
	assert.Len(t, sb.events, 1)
	assert.Equal(t, "event1", sb.events[0].Name)
}

func TestCov_SpanBuilder_WithLink(t *testing.T) {
	sb := NewSpanBuilder("test-span")
	sc := trace.SpanContext{}
	result := sb.WithLink(sc, attribute.String("lkey", "lval"))
	assert.Equal(t, sb, result)
	assert.Len(t, sb.links, 1)
}

func TestCov_SpanBuilder_Build(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	sb := NewSpanBuilder("built-span").
		WithAttribute("key1", "value1").
		WithKind(trace.SpanKindInternal).
		WithEvent("my-event")
	ctx, span := sb.Build(context.Background(), tp)
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestCov_SpanBuilder_BuildWithLink(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	parentCtx, parentSpan := tp.StartSpan(context.Background(), SpanConfig{Name: "parent"})
	parentSpan.End()
	sc := parentSpan.SpanContext()
	sb := NewSpanBuilder("child-with-link").WithLink(sc)
	ctx, span := sb.Build(parentCtx, tp)
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestCov_CreateAttribute_String(t *testing.T) {
	kv := createAttribute("k", "v")
	assert.Equal(t, attribute.STRING, kv.Value.Type())
}

func TestCov_CreateAttribute_Int(t *testing.T) {
	kv := createAttribute("k", 42)
	assert.Equal(t, attribute.INT64, kv.Value.Type())
}

func TestCov_CreateAttribute_Int64(t *testing.T) {
	kv := createAttribute("k", int64(42))
	assert.Equal(t, attribute.INT64, kv.Value.Type())
}

func TestCov_CreateAttribute_Float64(t *testing.T) {
	kv := createAttribute("k", 3.14)
	assert.Equal(t, attribute.FLOAT64, kv.Value.Type())
}

func TestCov_CreateAttribute_Bool(t *testing.T) {
	kv := createAttribute("k", true)
	assert.Equal(t, attribute.BOOL, kv.Value.Type())
}

func TestCov_CreateAttribute_StringSlice(t *testing.T) {
	kv := createAttribute("k", []string{"a", "b"})
	assert.Equal(t, attribute.STRINGSLICE, kv.Value.Type())
}

func TestCov_CreateAttribute_Int64Slice(t *testing.T) {
	kv := createAttribute("k", []int64{1, 2})
	assert.Equal(t, attribute.INT64SLICE, kv.Value.Type())
}

func TestCov_CreateAttribute_Float64Slice(t *testing.T) {
	kv := createAttribute("k", []float64{1.1, 2.2})
	assert.Equal(t, attribute.FLOAT64SLICE, kv.Value.Type())
}

func TestCov_CreateAttribute_BoolSlice(t *testing.T) {
	kv := createAttribute("k", []bool{true, false})
	assert.Equal(t, attribute.BOOLSLICE, kv.Value.Type())
}

func TestCov_CreateAttribute_Default(t *testing.T) {
	kv := createAttribute("k", struct{}{})
	assert.Equal(t, attribute.STRING, kv.Value.Type())
}

func TestCov_SpanContextExtractor_ExtractAndInject(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	sce := NewSpanContextExtractor()
	assert.NotNil(t, sce)

	ctx, span := tp.StartSpan(context.Background(), SpanConfig{Name: "extract-test"})
	defer span.End()

	carrier := NewPropagationCarrier()
	sce.Inject(ctx, carrier)

	extractedCtx := sce.Extract(context.Background(), carrier)
	assert.NotNil(t, extractedCtx)
}

func TestCov_GetTraceID_NoSpan(t *testing.T) {
	id := GetTraceID(context.Background())
	assert.Equal(t, "", id)
}

func TestCov_GetSpanID_NoSpan(t *testing.T) {
	id := GetSpanID(context.Background())
	assert.Equal(t, "", id)
}

func TestCov_GetTraceID_WithSpan(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	ctx, span := tp.StartSpan(context.Background(), SpanConfig{Name: "trace-id-test"})
	defer span.End()
	traceID := GetTraceID(ctx)
	_ = traceID
}

func TestCov_GetSpanID_WithSpan(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	ctx, span := tp.StartSpan(context.Background(), SpanConfig{Name: "span-id-test"})
	defer span.End()
	spanID := GetSpanID(ctx)
	_ = spanID
}

func TestCov_SetSpanError_Nil(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	_, span := tp.StartSpan(context.Background(), SpanConfig{Name: "err-test"})
	defer span.End()
	SetSpanError(span, nil)
}

func TestCov_SetSpanError_WithError(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	_, span := tp.StartSpan(context.Background(), SpanConfig{Name: "err-test"})
	defer span.End()
	SetSpanError(span, assert.AnError)
}

func TestCov_SetSpanErrorWithStack_Nil(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	_, span := tp.StartSpan(context.Background(), SpanConfig{Name: "stack-test"})
	defer span.End()
	SetSpanErrorWithStack(span, nil, "stack-trace")
}

func TestCov_SetSpanErrorWithStack_WithError(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	_, span := tp.StartSpan(context.Background(), SpanConfig{Name: "stack-test"})
	defer span.End()
	SetSpanErrorWithStack(span, assert.AnError, "goroutine 1:\n\tmain.go:10")
}

func TestCov_SpanTimer(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	_, span := tp.StartSpan(context.Background(), SpanConfig{Name: "timer-test"})
	defer span.End()
	timer := NewSpanTimer(span, "operation", attribute.String("op", "test"))
	assert.NotNil(t, timer)
	time.Sleep(10 * time.Millisecond)
	d := timer.Duration()
	assert.True(t, d >= 10*time.Millisecond)
	timer.Stop()
}

func TestCov_PropagationCarrier_MultiKey(t *testing.T) {
	pc := NewPropagationCarrier()
	assert.NotNil(t, pc)
	assert.Empty(t, pc.Keys())

	pc.Set("key1", "value1")
	pc.Set("key2", "value2")
	assert.Equal(t, "value1", pc.Get("key1"))
	assert.Equal(t, "value2", pc.Get("key2"))
	assert.Equal(t, "", pc.Get("nonexistent"))

	keys := pc.Keys()
	assert.Len(t, keys, 2)

	m := pc.ToMap()
	assert.Len(t, m, 2)
	assert.Equal(t, "value1", m["key1"])
}

func TestCov_PropagationCarrierFromMap(t *testing.T) {
	m := map[string]string{
		"traceparent": "00-abc123-def456-01",
		"baggage":     "key=value",
	}
	pc := PropagationCarrierFromMap(m)
	assert.NotNil(t, pc)
	assert.Equal(t, "00-abc123-def456-01", pc.Get("traceparent"))
	assert.Equal(t, "key=value", pc.Get("baggage"))
}

func TestCov_CreateSampler_Always(t *testing.T) {
	cfg := &Config{SamplerType: "always"}
	s := createSampler(cfg)
	assert.NotNil(t, s)
}

func TestCov_CreateSampler_Never(t *testing.T) {
	cfg := &Config{SamplerType: "never"}
	s := createSampler(cfg)
	assert.NotNil(t, s)
}

func TestCov_CreateSampler_Ratio(t *testing.T) {
	cfg := &Config{SamplerType: "ratio", SamplerRatio: 0.5}
	s := createSampler(cfg)
	assert.NotNil(t, s)
}

func TestCov_CreateSampler_ParentBased(t *testing.T) {
	cfg := &Config{SamplerType: "parentbased", SamplerRatio: 0.5}
	s := createSampler(cfg)
	assert.NotNil(t, s)
}

func TestCov_CreateSampler_Default(t *testing.T) {
	cfg := &Config{SamplerType: "unknown"}
	s := createSampler(cfg)
	assert.NotNil(t, s)
}

func TestCov_SpanBuilder_ChainedCalls(t *testing.T) {
	cfg := &Config{ServiceName: "test", Enabled: false}
	tp, _ := NewTracerProvider(cfg, nil)
	sb := NewSpanBuilder("chained").
		WithAttribute("a1", "v1").
		WithAttributes(map[string]interface{}{"a2": 10, "a3": true}).
		WithKind(trace.SpanKindClient).
		WithEvent("e1").
		WithEvent("e2", attribute.String("ek", "ev"))
	ctx, span := sb.Build(context.Background(), tp)
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}
