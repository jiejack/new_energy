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

func TestNewTracerProvider_Disabled(t *testing.T) {
	tp, err := NewTracerProvider(&Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Enabled:        false,
	}, zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, tp)

	tracer := tp.GetTracer()
	assert.NotNil(t, tracer)

	ctx, span := tp.StartSpan(context.Background(), SpanConfig{
		Name: "test-span",
	})
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()

	err = tp.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestNewTracerProvider_Enabled_InvalidEndpoint(t *testing.T) {
	_, err := NewTracerProvider(&Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Endpoint:       "localhost:4317",
		Protocol:       "grpc",
		SamplerType:    "always",
		Enabled:        false,
	}, zap.NewNop())
	require.NoError(t, err)
}

func TestNewTracerProvider_NilLogger(t *testing.T) {
	tp, err := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, tp)
	tp.Shutdown(context.Background())
}

func TestTracerProvider_Shutdown_NilProvider(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
	err := tp.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestTracerProvider_ForceFlush_NilProvider(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
	err := tp.ForceFlush(context.Background())
	require.NoError(t, err)
}

func TestTracerProvider_StartSpanFromContext(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
	defer tp.Shutdown(context.Background())

	ctx, span := tp.StartSpanFromContext(context.Background(), "test-span")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestTracerProvider_NewStationSpan(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
	defer tp.Shutdown(context.Background())

	ctx, span := tp.NewStationSpan(context.Background(), "station-operation", "station-1")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestTracerProvider_NewDeviceSpan(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
	defer tp.Shutdown(context.Background())

	ctx, span := tp.NewDeviceSpan(context.Background(), "device-operation", "station-1", "device-1")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestTracerProvider_NewPointSpan(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
	defer tp.Shutdown(context.Background())

	ctx, span := tp.NewPointSpan(context.Background(), "point-operation", "station-1", "device-1", "point-1")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestTracerProvider_NewHTTPSpan(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
	defer tp.Shutdown(context.Background())

	ctx, span := tp.NewHTTPSpan(context.Background(), "GET", "/api/v1/devices")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestSpanBuilder(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
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

func TestSpanBuilder_WithLink(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
	defer tp.Shutdown(context.Background())

	_, parentSpan := tp.StartSpan(context.Background(), SpanConfig{Name: "parent"})
	parentSpan.End()

	sb := NewSpanBuilder("linked-span")
	sb.WithLink(parentSpan.SpanContext())

	ctx, span := sb.Build(context.Background(), tp)
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestSpanContextExtractor(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
	defer tp.Shutdown(context.Background())

	sce := NewSpanContextExtractor()
	assert.NotNil(t, sce)

	carrier := NewPropagationCarrier()
	ctx := sce.Extract(context.Background(), carrier)
	assert.NotNil(t, ctx)

	_, span := tp.StartSpan(ctx, SpanConfig{Name: "test"})
	defer span.End()

	sce.Inject(context.Background(), carrier)
}

func TestGetTraceID(t *testing.T) {
	traceID := GetTraceID(context.Background())
	assert.Equal(t, "", traceID)
}

func TestGetSpanID(t *testing.T) {
	spanID := GetSpanID(context.Background())
	assert.Equal(t, "", spanID)
}

func TestSetSpanError(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
	defer tp.Shutdown(context.Background())

	_, span := tp.StartSpan(context.Background(), SpanConfig{Name: "error-span"})
	SetSpanError(span, assert.AnError)
	span.End()

	SetSpanError(span, nil)
}

func TestSetSpanErrorWithStack(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
	defer tp.Shutdown(context.Background())

	_, span := tp.StartSpan(context.Background(), SpanConfig{Name: "error-span"})
	SetSpanErrorWithStack(span, assert.AnError, "stack trace here")
	span.End()

	SetSpanErrorWithStack(span, nil, "")
}

func TestSpanTimer(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
	defer tp.Shutdown(context.Background())

	_, span := tp.StartSpan(context.Background(), SpanConfig{Name: "timed-span"})

	timer := NewSpanTimer(span, "operation", attribute.String("op", "test"))
	time.Sleep(10 * time.Millisecond)
	dur := timer.Duration()
	assert.True(t, dur >= 10*time.Millisecond)

	timer.Stop()
	span.End()
}

func TestSetHTTPStatus(t *testing.T) {
	tp, _ := NewTracerProvider(&Config{
		ServiceName: "test",
		Enabled:     false,
	}, zap.NewNop())
	defer tp.Shutdown(context.Background())

	_, span := tp.StartSpan(context.Background(), SpanConfig{Name: "http-span"})
	SetHTTPStatus(span, 200)
	span.End()
}

func TestPropagationCarrier(t *testing.T) {
	pc := NewPropagationCarrier()
	assert.NotNil(t, pc)

	pc.Set("key1", "value1")
	assert.Equal(t, "value1", pc.Get("key1"))
	assert.Equal(t, "", pc.Get("nonexistent"))

	keys := pc.Keys()
	assert.Equal(t, 1, len(keys))

	m := pc.ToMap()
	assert.Equal(t, "value1", m["key1"])
}

func TestPropagationCarrierFromMap(t *testing.T) {
	m := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	pc := PropagationCarrierFromMap(m)
	assert.Equal(t, "value1", pc.Get("key1"))
	assert.Equal(t, "value2", pc.Get("key2"))
}

func TestCommonAttributes(t *testing.T) {
	assert.Equal(t, attribute.Key("service.name"), CommonAttributes.ServiceName)
	assert.Equal(t, attribute.Key("service.version"), CommonAttributes.ServiceVersion)
	assert.Equal(t, attribute.Key("environment"), CommonAttributes.Environment)
	assert.Equal(t, attribute.Key("station.id"), CommonAttributes.StationID)
	assert.Equal(t, attribute.Key("device.id"), CommonAttributes.DeviceID)
	assert.Equal(t, attribute.Key("point.id"), CommonAttributes.PointID)
	assert.Equal(t, attribute.Key("user.id"), CommonAttributes.UserID)
	assert.Equal(t, attribute.Key("request.id"), CommonAttributes.RequestID)
	assert.Equal(t, attribute.Key("http.method"), CommonAttributes.Method)
	assert.Equal(t, attribute.Key("http.endpoint"), CommonAttributes.Endpoint)
	assert.Equal(t, attribute.Key("http.status_code"), CommonAttributes.StatusCode)
	assert.Equal(t, attribute.Key("error.type"), CommonAttributes.ErrorType)
}

func TestConfig_Struct(t *testing.T) {
	cfg := Config{
		ServiceName:    "test",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Endpoint:       "localhost:4317",
		Protocol:       "grpc",
		SamplerType:    "always",
		SamplerRatio:   1.0,
		BatchTimeout:   5 * time.Second,
		ExportTimeout:  30 * time.Second,
		MaxExportBatch: 512,
		MaxQueueSize:   2048,
		Enabled:        true,
	}
	assert.Equal(t, "test", cfg.ServiceName)
	assert.Equal(t, "grpc", cfg.Protocol)
	assert.Equal(t, "always", cfg.SamplerType)
}

func TestSpanConfig_Struct(t *testing.T) {
	sc := SpanConfig{
		Name:  "test",
		Kind:  trace.SpanKindServer,
		Attributes: []attribute.KeyValue{
			attribute.String("key", "value"),
		},
	}
	assert.Equal(t, "test", sc.Name)
	assert.Equal(t, trace.SpanKindServer, sc.Kind)
}
