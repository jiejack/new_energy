package tracing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func TestCreateSampler_Always(t *testing.T) {
	cfg := &Config{SamplerType: "always"}
	sampler := createSampler(cfg)
	require.NotNil(t, sampler)
}

func TestCreateSampler_Never(t *testing.T) {
	cfg := &Config{SamplerType: "never"}
	sampler := createSampler(cfg)
	require.NotNil(t, sampler)
}

func TestCreateSampler_Ratio(t *testing.T) {
	cfg := &Config{SamplerType: "ratio", SamplerRatio: 0.5}
	sampler := createSampler(cfg)
	require.NotNil(t, sampler)
}

func TestCreateSampler_ParentBased(t *testing.T) {
	cfg := &Config{SamplerType: "parentbased", SamplerRatio: 0.5}
	sampler := createSampler(cfg)
	require.NotNil(t, sampler)
}

func TestCreateSampler_Default(t *testing.T) {
	cfg := &Config{SamplerType: "unknown"}
	sampler := createSampler(cfg)
	require.NotNil(t, sampler)
}

func TestCreateExporter_GRPC(t *testing.T) {
	cfg := &Config{
		Endpoint: "localhost:4317",
		Protocol: "grpc",
	}
	exporter, err := createExporter(cfg)
	require.NoError(t, err)
	require.NotNil(t, exporter)
}

func TestCreateExporter_HTTP(t *testing.T) {
	cfg := &Config{
		Endpoint: "localhost:4318",
		Protocol: "http",
	}
	exporter, err := createExporter(cfg)
	require.NoError(t, err)
	require.NotNil(t, exporter)
}

func TestCreateExporter_DefaultProtocol(t *testing.T) {
	cfg := &Config{
		Endpoint: "localhost:4317",
		Protocol: "",
	}
	exporter, err := createExporter(cfg)
	require.NoError(t, err)
	require.NotNil(t, exporter)
}

func newTestTracerProvider() *TracerProvider {
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	return &TracerProvider{
		provider: provider,
		tracer:   provider.Tracer("test"),
		config:   &Config{ServiceName: "test"},
		logger:   zap.NewNop(),
	}
}

func TestTracerProvider_Shutdown_WithProvider(t *testing.T) {
	tp := newTestTracerProvider()
	err := tp.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestTracerProvider_ForceFlush_WithProvider(t *testing.T) {
	tp := newTestTracerProvider()
	defer tp.Shutdown(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := tp.ForceFlush(ctx)
	assert.NoError(t, err)
}

func TestGetTraceID_WithValidSpan(t *testing.T) {
	tp := newTestTracerProvider()
	defer tp.Shutdown(context.Background())

	ctx, span := tp.StartSpan(context.Background(), SpanConfig{Name: "test"})
	defer span.End()

	traceID := GetTraceID(ctx)
	assert.NotEmpty(t, traceID)
}

func TestGetSpanID_WithValidSpan(t *testing.T) {
	tp := newTestTracerProvider()
	defer tp.Shutdown(context.Background())

	ctx, span := tp.StartSpan(context.Background(), SpanConfig{Name: "test"})
	defer span.End()

	spanID := GetSpanID(ctx)
	assert.NotEmpty(t, spanID)
}

func TestTracerProvider_AllSpans_WithProvider(t *testing.T) {
	tp := newTestTracerProvider()
	defer tp.Shutdown(context.Background())

	ctx, span := tp.StartSpan(context.Background(), SpanConfig{
		Name: "test-span",
		Kind: trace.SpanKindServer,
	})
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()

	ctx2, span2 := tp.StartSpanFromContext(context.Background(), "ctx-span")
	assert.NotNil(t, ctx2)
	assert.NotNil(t, span2)
	span2.End()

	ctx3, span3 := tp.NewStationSpan(context.Background(), "station-op", "station-1")
	assert.NotNil(t, ctx3)
	assert.NotNil(t, span3)
	span3.End()

	ctx4, span4 := tp.NewDeviceSpan(context.Background(), "device-op", "station-1", "device-1")
	assert.NotNil(t, ctx4)
	assert.NotNil(t, span4)
	span4.End()

	ctx5, span5 := tp.NewPointSpan(context.Background(), "point-op", "station-1", "device-1", "point-1")
	assert.NotNil(t, ctx5)
	assert.NotNil(t, span5)
	span5.End()

	ctx6, span6 := tp.NewHTTPSpan(context.Background(), "GET", "/api/test")
	assert.NotNil(t, ctx6)
	assert.NotNil(t, span6)
	span6.End()
}

func TestNewTracerProvider_Enabled_ResourceError(t *testing.T) {
	_, err := NewTracerProvider(&Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Endpoint:       "localhost:4317",
		Protocol:       "grpc",
		SamplerType:    "always",
		Enabled:        true,
	}, zap.NewNop())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create resource")
}

func TestCreateAttribute_AllTypes(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"string", "k1", "v1"},
		{"int", "k2", 42},
		{"int64", "k3", int64(42)},
		{"float64", "k4", 3.14},
		{"bool", "k5", true},
		{"[]string", "k6", []string{"a", "b"}},
		{"[]int64", "k7", []int64{1, 2}},
		{"[]float64", "k8", []float64{1.1, 2.2}},
		{"[]bool", "k9", []bool{true, false}},
		{"default", "k10", struct{ X int }{X: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kv := createAttribute(tt.key, tt.value)
			assert.Equal(t, attribute.Key(tt.key), kv.Key)
		})
	}
}

func TestSpanBuilder_WithLinkAndAttrs(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
	defer tp.Shutdown(context.Background())

	_, parentSpan := tp.StartSpan(context.Background(), SpanConfig{Name: "parent"})
	parentSpan.End()

	sb := NewSpanBuilder("linked-span")
	sb.WithLink(parentSpan.SpanContext(), attribute.String("link.key", "link.val"))

	ctx, span := sb.Build(context.Background(), tp)
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestSpanContextExtractor_ExtractInject(t *testing.T) {
	tp := newTestTracerProvider()
	defer tp.Shutdown(context.Background())

	sce := NewSpanContextExtractor()
	assert.NotNil(t, sce)

	carrier := NewPropagationCarrier()

	ctx, span := tp.StartSpan(context.Background(), SpanConfig{Name: "test"})
	defer span.End()

	sce.Inject(ctx, carrier)

	extractedCtx := sce.Extract(context.Background(), carrier)
	assert.NotNil(t, extractedCtx)
}

func TestPropagationCarrier_AllMethods(t *testing.T) {
	pc := NewPropagationCarrier()
	pc.Set("key1", "value1")
	pc.Set("key2", "value2")

	assert.Equal(t, "value1", pc.Get("key1"))
	assert.Equal(t, "value2", pc.Get("key2"))
	assert.Equal(t, "", pc.Get("nonexistent"))

	keys := pc.Keys()
	assert.Equal(t, 2, len(keys))

	m := pc.ToMap()
	assert.Equal(t, "value1", m["key1"])
	assert.Equal(t, "value2", m["key2"])

	pc2 := PropagationCarrierFromMap(map[string]string{
		"a": "b",
		"c": "d",
	})
	assert.Equal(t, "b", pc2.Get("a"))
	assert.Equal(t, "d", pc2.Get("c"))
}

func TestSetSpanError_WithRealSpan(t *testing.T) {
	tp := newTestTracerProvider()
	defer tp.Shutdown(context.Background())

	_, span := tp.StartSpan(context.Background(), SpanConfig{Name: "error-span"})
	SetSpanError(span, assert.AnError)
	span.End()

	SetSpanError(span, nil)
}

func TestSetSpanErrorWithStack_WithRealSpan(t *testing.T) {
	tp := newTestTracerProvider()
	defer tp.Shutdown(context.Background())

	_, span := tp.StartSpan(context.Background(), SpanConfig{Name: "error-span"})
	SetSpanErrorWithStack(span, assert.AnError, "stack trace here")
	span.End()

	SetSpanErrorWithStack(span, nil, "")
}

func TestSpanTimer_WithRealSpan(t *testing.T) {
	tp := newTestTracerProvider()
	defer tp.Shutdown(context.Background())

	_, span := tp.StartSpan(context.Background(), SpanConfig{Name: "timed-span"})

	timer := NewSpanTimer(span, "operation", attribute.String("op", "test"))
	time.Sleep(10 * time.Millisecond)
	dur := timer.Duration()
	assert.True(t, dur >= 10*time.Millisecond)

	timer.Stop()
	span.End()
}

func TestSetHTTPStatus_WithRealSpan(t *testing.T) {
	tp := newTestTracerProvider()
	defer tp.Shutdown(context.Background())

	_, span := tp.StartSpan(context.Background(), SpanConfig{Name: "http-span"})
	SetHTTPStatus(span, 200)
	span.End()
}

func TestSpanBuilder_Full_WithRealProvider(t *testing.T) {
	tp := newTestTracerProvider()
	defer tp.Shutdown(context.Background())

	sb := NewSpanBuilder("builder-span")
	sb.WithAttribute("key1", "value1")
	sb.WithAttribute("key2", 42)
	sb.WithAttribute("key3", true)
	sb.WithAttributes(map[string]interface{}{
		"attr1": "string",
		"attr2": 123,
		"attr3": 3.14,
		"attr4": false,
		"attr5": []string{"a", "b"},
		"attr6": []int64{1, 2},
		"attr7": []float64{1.1, 2.2},
		"attr8": []bool{true, false},
		"attr9": struct{ X int }{X: 1},
	})
	sb.WithKind(trace.SpanKindServer)
	sb.WithEvent("test-event", attribute.String("event.key", "event.value"))

	ctx, span := sb.Build(context.Background(), tp)
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}
