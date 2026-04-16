package processing

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/types"
)

const (
	defaultWindowSize  = 60
	defaultSlideSize   = 30
	defaultParallelism = 4
)

type WindowType string

const (
	TumblingWindow WindowType = "tumbling"
	SlidingWindow  WindowType = "sliding"
	SessionWindow  WindowType = "session"
)

type AggregationType string

const (
	AggSum   AggregationType = "sum"
	AggAvg   AggregationType = "avg"
	AggMin   AggregationType = "min"
	AggMax   AggregationType = "max"
	AggCount AggregationType = "count"
)

type FlinkProcessor struct {
	config        types.ProcessingConfig
	jobID         string
	isRunning     bool
	windowType    WindowType
	windowSize    time.Duration
	slideSize     time.Duration
	parallelism   int
	windows       map[string]*WindowState
	mu            sync.Mutex
	stopChan      chan struct{}
	started       bool
}

type WindowState struct {
	DataPoints  []*types.DataPoint
	StartTime   time.Time
	EndTime     time.Time
	Aggregations map[string]map[AggregationType]float64
}

type StreamOperator interface {
	Process(dp *types.DataPoint) (*types.DataPoint, error)
}

type FilterOperator struct {
	Condition func(*types.DataPoint) bool
}

type MapOperator struct {
	Transform func(*types.DataPoint) *types.DataPoint
}

type AggregationOperator struct {
	GroupBy    string
	AggType    AggregationType
	MetricName string
}

func NewFlinkProcessor() *FlinkProcessor {
	return &FlinkProcessor{
		windows:     make(map[string]*WindowState),
		stopChan:    make(chan struct{}),
		parallelism: defaultParallelism,
	}
}

func (f *FlinkProcessor) Init(config types.ProcessingConfig) error {
	f.config = config

	if config.Parallelism > 0 {
		f.parallelism = config.Parallelism
	}

	windowSizeSec := defaultWindowSize
	if config.WindowSize != "" {
		fmt.Sscanf(config.WindowSize, "%d", &windowSizeSec)
	}
	f.windowSize = time.Duration(windowSizeSec) * time.Second

	slideSizeSec := defaultSlideSize
	if config.SlideSize != "" {
		fmt.Sscanf(config.SlideSize, "%d", &slideSizeSec)
	}
	f.slideSize = time.Duration(slideSizeSec) * time.Second

	f.windowType = TumblingWindow

	fmt.Printf("Initializing Flink processor with config: %+v\n", config)
	fmt.Printf("  Window Type: %s\n", f.windowType)
	fmt.Printf("  Window Size: %v\n", f.windowSize)
	fmt.Printf("  Slide Size: %v\n", f.slideSize)
	fmt.Printf("  Parallelism: %d\n", f.parallelism)

	f.isRunning = false
	f.started = true

	return nil
}

func (f *FlinkProcessor) Process(data *types.BatchData) (*types.BatchData, error) {
	if !f.isRunning {
		if err := f.startFlinkJob(); err != nil {
			return nil, err
		}
	}

	if f.config.Type == "batch" {
		return f.processBatchData(data)
	}

	return f.processStreamData(data)
}

func (f *FlinkProcessor) startFlinkJob() error {
	fmt.Println("Starting Flink job: New Energy Monitoring Stream Processing")
	fmt.Println("  Operators:")
	fmt.Println("    - Data Cleaning")
	fmt.Println("    - Quality Validation")
	fmt.Println("    - Window Aggregation")
	fmt.Println("    - Real-time Alerting")

	f.jobID = fmt.Sprintf("job_flink_%d", time.Now().UnixNano())
	f.isRunning = true

	fmt.Printf("Flink job started with ID: %s\n", f.jobID)

	return nil
}

func (f *FlinkProcessor) processBatchData(data *types.BatchData) (*types.BatchData, error) {
	startTime := time.Now()
	fmt.Printf("Processing batch data with Flink, %d data points\n", len(data.DataPoints))

	processedDataPoints := make([]*types.DataPoint, 0, len(data.DataPoints))
	stats := map[string]int{
		"total":   len(data.DataPoints),
		"cleaned": 0,
		"invalid": 0,
	}

	for _, dp := range data.DataPoints {
		processedDP := f.processDataPoint(dp, "batch", &stats)
		if processedDP != nil {
			processedDataPoints = append(processedDataPoints, processedDP)
		}
	}

	processingTime := time.Since(startTime)
	_ = processingTime

	fmt.Printf("Batch processing complete: cleaned=%d, invalid=%d, duration=%v\n",
		stats["cleaned"], stats["invalid"], processingTime)

	return &types.BatchData{
		DataPoints: processedDataPoints,
		Metadata: types.Metadata{
			Source:      "flink_batch",
			BatchID:     fmt.Sprintf("batch_flink_%d", time.Now().UnixNano()),
			Timestamp:   time.Now(),
			RecordCount: len(processedDataPoints),
			Properties: map[string]interface{}{
				"processor":       "flink",
				"mode":            "batch",
				"job_id":          f.jobID,
				"processing_time": processingTime.Seconds(),
				"stats":           stats,
			},
		},
	}, nil
}

func (f *FlinkProcessor) processStreamData(data *types.BatchData) (*types.BatchData, error) {
	startTime := time.Now()
	fmt.Printf("Processing stream data with Flink, %d data points\n", len(data.DataPoints))

	processedDataPoints := make([]*types.DataPoint, 0, len(data.DataPoints))
	stats := map[string]int{
		"total":   len(data.DataPoints),
		"cleaned": 0,
		"invalid": 0,
		"windowed": 0,
	}

	for _, dp := range data.DataPoints {
		processedDP := f.processDataPoint(dp, "stream", &stats)
		if processedDP != nil {
			processedDataPoints = append(processedDataPoints, processedDP)
			f.updateWindows(processedDP, &stats)
		}
	}

	processingTime := time.Since(startTime)
	_ = processingTime

	fmt.Printf("Stream processing complete: cleaned=%d, invalid=%d, windowed=%d, duration=%v\n",
		stats["cleaned"], stats["invalid"], stats["windowed"], processingTime)

	return &types.BatchData{
		DataPoints: processedDataPoints,
		Metadata: types.Metadata{
			Source:      "flink_stream",
			BatchID:     fmt.Sprintf("stream_flink_%d", time.Now().UnixNano()),
			Timestamp:   time.Now(),
			RecordCount: len(processedDataPoints),
			Properties: map[string]interface{}{
				"processor":       "flink",
				"mode":            "stream",
				"job_id":          f.jobID,
				"window_type":     f.windowType,
				"window_size":     f.windowSize.Seconds(),
				"processing_time": processingTime.Seconds(),
				"stats":           stats,
			},
		},
	}, nil
}

func (f *FlinkProcessor) processDataPoint(dp *types.DataPoint, mode string, stats *map[string]int) *types.DataPoint {
	if dp.Attributes == nil {
		dp.Attributes = make(map[string]interface{})
	}

	if dp.Value < 0 {
		dp.Value = 0
		(*stats)["cleaned"]++
	}

	if math.IsNaN(dp.Value) || math.IsInf(dp.Value, 0) {
		(*stats)["invalid"]++
		return nil
	}

	dp.Attributes["processed"] = true
	dp.Attributes["processed_at"] = time.Now()
	dp.Attributes["processor"] = "flink"
	dp.Attributes["mode"] = mode
	dp.Attributes["job_id"] = f.jobID

	if dp.Tags == nil {
		dp.Tags = make(map[string]string)
	}
	dp.Tags["processor"] = "flink"
	dp.Tags["mode"] = mode

	(*stats)["cleaned"]++
	return dp
}

func (f *FlinkProcessor) updateWindows(dp *types.DataPoint, stats *map[string]int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	windowKey := fmt.Sprintf("%s_%s", dp.DeviceID, dp.Metric)
	windowStart := dp.Timestamp.Truncate(f.windowSize)
	windowEnd := windowStart.Add(f.windowSize)

	if _, exists := f.windows[windowKey]; !exists {
		f.windows[windowKey] = &WindowState{
			DataPoints:   make([]*types.DataPoint, 0),
			StartTime:    windowStart,
			EndTime:      windowEnd,
			Aggregations: make(map[string]map[AggregationType]float64),
		}
	}

	window := f.windows[windowKey]
	window.DataPoints = append(window.DataPoints, dp)

	f.calculateAggregations(window, dp)

	(*stats)["windowed"]++
}

func (f *FlinkProcessor) calculateAggregations(window *WindowState, dp *types.DataPoint) {
	metricKey := dp.Metric

	if _, exists := window.Aggregations[metricKey]; !exists {
		window.Aggregations[metricKey] = make(map[AggregationType]float64)
		window.Aggregations[metricKey][AggMin] = math.MaxFloat64
		window.Aggregations[metricKey][AggMax] = -math.MaxFloat64
	}

	agg := window.Aggregations[metricKey]
	count := agg[AggCount] + 1
	sum := agg[AggSum] + dp.Value

	agg[AggCount] = count
	agg[AggSum] = sum
	agg[AggAvg] = sum / count

	if dp.Value < agg[AggMin] {
		agg[AggMin] = dp.Value
	}
	if dp.Value > agg[AggMax] {
		agg[AggMax] = dp.Value
	}
}

func (f *FlinkProcessor) GetWindowAggregations(deviceID, metricName string) map[AggregationType]float64 {
	f.mu.Lock()
	defer f.mu.Unlock()

	windowKey := fmt.Sprintf("%s_%s", deviceID, metricName)
	if window, exists := f.windows[windowKey]; exists {
		if agg, exists := window.Aggregations[metricName]; exists {
			return agg
		}
	}

	return nil
}

func (f *FlinkProcessor) GetStats() map[string]interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	stats := map[string]interface{}{
		"processor":    "flink",
		"job_id":       f.jobID,
		"is_running":   f.isRunning,
		"started":      f.started,
		"window_type":  f.windowType,
		"window_size":  f.windowSize.Seconds(),
		"slide_size":   f.slideSize.Seconds(),
		"parallelism":  f.parallelism,
		"window_count": len(f.windows),
	}

	return stats
}

func (f *FlinkProcessor) StopJob() error {
	if !f.isRunning || f.jobID == "" {
		return nil
	}

	fmt.Printf("Stopping Flink job: %s\n", f.jobID)
	f.isRunning = false

	return nil
}

func (f *FlinkProcessor) Close() error {
	if !f.started {
		return nil
	}

	if f.isRunning && f.jobID != "" {
		fmt.Printf("Cancelling Flink job: %s\n", f.jobID)
	}

	close(f.stopChan)

	f.mu.Lock()
	f.windows = make(map[string]*WindowState)
	f.mu.Unlock()

	f.isRunning = false
	f.started = false
	fmt.Println("Flink processor closed")

	return nil
}
