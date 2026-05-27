package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewMetricsCollector(t *testing.T) {
	mc := NewMetricsCollector(&Config{
		Namespace:   "test",
		Subsystem:   "sub",
		ConstLabels: map[string]string{"env": "test"},
	}, zap.NewNop())
	require.NotNil(t, mc)
}

func TestNewMetricsCollector_NilLogger(t *testing.T) {
	mc := NewMetricsCollector(&Config{
		Namespace: "test",
	}, nil)
	require.NotNil(t, mc)
}

func TestMetricsCollector_RegisterCounter(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	err := mc.RegisterCounter(CounterConfig{
		Name:   "test_counter",
		Help:   "A test counter",
		Labels: []string{"method"},
	})
	require.NoError(t, err)

	err = mc.RegisterCounter(CounterConfig{
		Name:   "test_counter",
		Help:   "Duplicate counter",
		Labels: []string{"method"},
	})
	assert.Error(t, err)
}

func TestMetricsCounter_IncCounter(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	mc.RegisterCounter(CounterConfig{
		Name:   "test_counter",
		Help:   "A test counter",
		Labels: []string{"method"},
	})

	err := mc.IncCounter("test_counter", prometheus.Labels{"method": "GET"})
	require.NoError(t, err)

	err = mc.IncCounter("nonexistent", prometheus.Labels{})
	assert.Error(t, err)
}

func TestMetricsCollector_AddCounter(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	mc.RegisterCounter(CounterConfig{
		Name:   "test_counter",
		Help:   "A test counter",
		Labels: []string{"method"},
	})

	err := mc.AddCounter("test_counter", 5.0, prometheus.Labels{"method": "POST"})
	require.NoError(t, err)

	err = mc.AddCounter("nonexistent", 1.0, prometheus.Labels{})
	assert.Error(t, err)
}

func TestMetricsCollector_RegisterGauge(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	err := mc.RegisterGauge(GaugeConfig{
		Name:   "test_gauge",
		Help:   "A test gauge",
		Labels: []string{"component"},
	})
	require.NoError(t, err)

	err = mc.RegisterGauge(GaugeConfig{
		Name:   "test_gauge",
		Help:   "Duplicate gauge",
		Labels: []string{"component"},
	})
	assert.Error(t, err)
}

func TestMetricsCollector_SetGauge(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	mc.RegisterGauge(GaugeConfig{
		Name:   "test_gauge",
		Help:   "A test gauge",
		Labels: []string{"component"},
	})

	err := mc.SetGauge("test_gauge", 42.0, prometheus.Labels{"component": "db"})
	require.NoError(t, err)

	err = mc.SetGauge("nonexistent", 1.0, prometheus.Labels{})
	assert.Error(t, err)
}

func TestMetricsCollector_IncGauge(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	mc.RegisterGauge(GaugeConfig{
		Name:   "test_gauge",
		Help:   "A test gauge",
		Labels: []string{"component"},
	})

	err := mc.IncGauge("test_gauge", prometheus.Labels{"component": "db"})
	require.NoError(t, err)

	err = mc.IncGauge("nonexistent", prometheus.Labels{})
	assert.Error(t, err)
}

func TestMetricsCollector_DecGauge(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	mc.RegisterGauge(GaugeConfig{
		Name:   "test_gauge",
		Help:   "A test gauge",
		Labels: []string{"component"},
	})

	mc.SetGauge("test_gauge", 10.0, prometheus.Labels{"component": "db"})
	err := mc.DecGauge("test_gauge", prometheus.Labels{"component": "db"})
	require.NoError(t, err)

	err = mc.DecGauge("nonexistent", prometheus.Labels{})
	assert.Error(t, err)
}

func TestMetricsCollector_AddGauge(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	mc.RegisterGauge(GaugeConfig{
		Name:   "test_gauge",
		Help:   "A test gauge",
		Labels: []string{"component"},
	})

	mc.SetGauge("test_gauge", 10.0, prometheus.Labels{"component": "db"})
	err := mc.AddGauge("test_gauge", 5.0, prometheus.Labels{"component": "db"})
	require.NoError(t, err)

	err = mc.AddGauge("nonexistent", 1.0, prometheus.Labels{})
	assert.Error(t, err)
}

func TestMetricsCollector_RegisterHistogram(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	err := mc.RegisterHistogram(HistogramConfig{
		Name:    "test_histogram",
		Help:    "A test histogram",
		Labels:  []string{"method"},
		Buckets: []float64{0.1, 0.5, 1.0},
	})
	require.NoError(t, err)

	err = mc.RegisterHistogram(HistogramConfig{
		Name:   "test_histogram",
		Help:   "Duplicate histogram",
		Labels: []string{"method"},
	})
	assert.Error(t, err)
}

func TestMetricsCollector_RegisterHistogram_DefaultBuckets(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	err := mc.RegisterHistogram(HistogramConfig{
		Name:   "test_hist_default",
		Help:   "Histogram with default buckets",
		Labels: []string{},
	})
	require.NoError(t, err)
}

func TestMetricsCollector_ObserveHistogram(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	mc.RegisterHistogram(HistogramConfig{
		Name:    "test_histogram",
		Help:    "A test histogram",
		Labels:  []string{"method"},
		Buckets: []float64{0.1, 0.5, 1.0},
	})

	err := mc.ObserveHistogram("test_histogram", 0.25, prometheus.Labels{"method": "GET"})
	require.NoError(t, err)

	err = mc.ObserveHistogram("nonexistent", 0.1, prometheus.Labels{})
	assert.Error(t, err)
}

func TestMetricsCollector_RegisterSummary(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	err := mc.RegisterSummary(SummaryConfig{
		Name:       "test_summary",
		Help:       "A test summary",
		Labels:     []string{"method"},
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01},
	})
	require.NoError(t, err)

	err = mc.RegisterSummary(SummaryConfig{
		Name:   "test_summary",
		Help:   "Duplicate summary",
		Labels: []string{"method"},
	})
	assert.Error(t, err)
}

func TestMetricsCollector_RegisterSummary_Defaults(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	err := mc.RegisterSummary(SummaryConfig{
		Name:   "test_summary_default",
		Help:   "Summary with defaults",
		Labels: []string{},
	})
	require.NoError(t, err)
}

func TestMetricsCollector_ObserveSummary(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	mc.RegisterSummary(SummaryConfig{
		Name:       "test_summary",
		Help:       "A test summary",
		Labels:     []string{"method"},
		Objectives: map[float64]float64{0.5: 0.05},
	})

	err := mc.ObserveSummary("test_summary", 0.5, prometheus.Labels{"method": "GET"})
	require.NoError(t, err)

	err = mc.ObserveSummary("nonexistent", 0.1, prometheus.Labels{})
	assert.Error(t, err)
}

func TestMetricsCollector_NewTimer_Histogram(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	mc.RegisterHistogram(HistogramConfig{
		Name:    "test_timer_hist",
		Help:    "Timer histogram",
		Labels:  []string{"op"},
		Buckets: DefaultBuckets,
	})

	timer, err := mc.NewTimer("test_timer_hist", prometheus.Labels{"op": "query"})
	require.NoError(t, err)
	require.NotNil(t, timer)

	time.Sleep(10 * time.Millisecond)
	timer.Record()
}

func TestMetricsCollector_NewTimer_Summary(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	mc.RegisterSummary(SummaryConfig{
		Name:       "test_timer_summary",
		Help:       "Timer summary",
		Labels:     []string{"op"},
		Objectives: DefaultObjectives,
	})

	timer, err := mc.NewTimer("test_timer_summary", prometheus.Labels{"op": "query"})
	require.NoError(t, err)
	require.NotNil(t, timer)

	timer.RecordDuration(100 * time.Millisecond)
}

func TestMetricsCollector_NewTimer_NotFound(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	_, err := mc.NewTimer("nonexistent", prometheus.Labels{})
	assert.Error(t, err)
}

func TestMetricsCollector_StartStop(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	err := mc.Start(0, "/metrics")
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = mc.Stop(ctx)
	require.NoError(t, err)
}

func TestMetricsCollector_Stop_NotRunning(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	err := mc.Stop(context.Background())
	require.NoError(t, err)
}

func TestMetricsCollector_Start_AlreadyRunning(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	err := mc.Start(0, "/metrics")
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = mc.Start(0, "/metrics")
	assert.Error(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mc.Stop(ctx)
}

func TestMetricsCollector_GetRegistry(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	reg := mc.GetRegistry()
	require.NotNil(t, reg)
}

func TestMetricsCollector_RegisterCustomCollector(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	custom := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "custom_metric",
		Help: "A custom metric",
	})
	err := mc.RegisterCustomCollector(custom)
	require.NoError(t, err)

	unregistered := mc.UnregisterCustomCollector(custom)
	assert.True(t, unregistered)
}

func TestNewDefaultHistogramConfig(t *testing.T) {
	cfg := NewDefaultHistogramConfig("test_hist", "help text", []string{"method"})
	assert.Equal(t, "test_hist", cfg.Name)
	assert.Equal(t, "help text", cfg.Help)
	assert.Equal(t, []string{"method"}, cfg.Labels)
	assert.Equal(t, DefaultBuckets, cfg.Buckets)
}

func TestNewDefaultSummaryConfig(t *testing.T) {
	cfg := NewDefaultSummaryConfig("test_summary", "help text", []string{"method"})
	assert.Equal(t, "test_summary", cfg.Name)
	assert.Equal(t, "help text", cfg.Help)
	assert.Equal(t, []string{"method"}, cfg.Labels)
	assert.Equal(t, DefaultObjectives, cfg.Objectives)
	assert.Equal(t, 10*time.Minute, cfg.MaxAge)
	assert.Equal(t, uint32(5), cfg.AgeBuckets)
	assert.Equal(t, uint32(500), cfg.BufCap)
}

func TestMetricLabels_ToPrometheusLabels(t *testing.T) {
	ml := &MetricLabels{
		Service:   "api",
		Method:    "GET",
		Endpoint:  "/test",
		Status:    "200",
		Component: "db",
		Station:   "st1",
		Device:    "dev1",
		Point:     "pt1",
	}
	labels := ml.ToPrometheusLabels()
	assert.Equal(t, "api", labels["service"])
	assert.Equal(t, "GET", labels["method"])
	assert.Equal(t, "/test", labels["endpoint"])
	assert.Equal(t, "200", labels["status"])
	assert.Equal(t, "db", labels["component"])
	assert.Equal(t, "st1", labels["station"])
	assert.Equal(t, "dev1", labels["device"])
	assert.Equal(t, "pt1", labels["point"])
}

func TestMetricLabels_ToPrometheusLabels_Empty(t *testing.T) {
	ml := &MetricLabels{}
	labels := ml.ToPrometheusLabels()
	assert.Equal(t, 0, len(labels))
}

func TestMetricsCollector_RegisterCommonMetrics(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	cm, err := mc.RegisterCommonMetrics("test")
	require.NoError(t, err)
	require.NotNil(t, cm)
}

func TestCommonMetrics_RecordRequest(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	cm, _ := mc.RegisterCommonMetrics("test")
	cm.RecordRequest("api", "GET", "/test", "200", 100*time.Millisecond)
}

func TestCommonMetrics_IncDecInFlight(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	cm, _ := mc.RegisterCommonMetrics("test")
	cm.IncInFlight("api", "GET", "/test")
	cm.DecInFlight("api", "GET", "/test")
}

func TestCommonMetrics_RecordError(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	cm, _ := mc.RegisterCommonMetrics("test")
	cm.RecordError("api", "db", "timeout")
}

func TestCommonMetrics_RecordDataProcessed(t *testing.T) {
	mc := NewMetricsCollector(&Config{Namespace: "test"}, zap.NewNop())
	cm, _ := mc.RegisterCommonMetrics("test")
	cm.RecordDataProcessed("api", "collector", 1024.0)
}

func TestGlobalMetrics(t *testing.T) {
	APIRequestsTotal.WithLabelValues("api", "GET", "/test", "200").Inc()
	APIRequestDuration.WithLabelValues("api", "GET", "/test").Observe(0.1)
	CollectorDataPointsTotal.WithLabelValues("st1", "dev1").Inc()
	CollectorErrorsTotal.WithLabelValues("st1", "dev1", "timeout").Inc()
	CollectorQueueSize.Set(100)
	CollectorBufferUsage.WithLabelValues("st1").Set(50.0)
	ComputeTasksTotal.WithLabelValues("aggregate", "success").Inc()
	ComputeTaskDuration.WithLabelValues("aggregate").Observe(0.5)
	ComputeActiveTasks.Set(5)
	AlarmTotal.WithLabelValues("st1", "critical", "threshold").Inc()
	AlarmNotificationTotal.WithLabelValues("email", "sent").Inc()
	AlarmNotificationFailedTotal.Inc()
	AlarmActiveCount.WithLabelValues("st1", "critical").Set(3)
	AIRequestsTotal.WithLabelValues("qa", "gpt-4", "success").Inc()
	AIRequestDuration.WithLabelValues("qa", "gpt-4").Observe(1.0)
	AITokensTotal.WithLabelValues("qa", "gpt-4", "input").Add(100)
	DBConnectionsOpen.Set(10)
	DBConnectionsInUse.Set(5)
	DBConnectionsIdle.Set(5)
	DBWaitCount.Inc()
	DBWaitDuration.Add(0.01)
	CacheOperationsTotal.WithLabelValues("get", "hit").Inc()
	CacheHitRate.Set(0.85)
	CacheMemoryUsage.Set(1024 * 1024)
	KafkaMessagesProduced.WithLabelValues("topic1").Inc()
	KafkaMessagesConsumed.WithLabelValues("topic1", "group1").Inc()
	KafkaConsumerLag.WithLabelValues("topic1", "group1", "0").Set(10)
	StationsTotal.Set(50)
	DevicesTotal.WithLabelValues("st1", "inverter").Set(10)
	DevicesOnline.WithLabelValues("st1", "inverter").Set(8)
	DataPointsStored.Inc()
	DataPointsQueried.Inc()
}
