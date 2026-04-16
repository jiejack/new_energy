package processing

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"github.com/new-energy-monitoring/pkg/alarm/notifier"
	"github.com/new-energy-monitoring/pkg/ai/fault"
	"github.com/new-energy-monitoring/pkg/bigdata/types"
	"github.com/new-energy-monitoring/pkg/bigdata/visualization"
	"github.com/new-energy-monitoring/internal/infrastructure/mq"
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
	enabled       bool
	jobName       string
	operators     map[string]bool
	sinks         []map[string]interface{}
	sources       []map[string]interface{}
	checkpoint    map[string]interface{}
	metrics       map[string]interface{}
	kafkaProducer *mq.KafkaProducer
	kafkaConsumer *mq.KafkaConsumer
	kafkaConfig   mq.KafkaConfig
	ctx           context.Context
	cancel        context.CancelFunc
	anomalyDetectors map[string]fault.FaultDetector
	anomalyConfig    map[string]interface{}
	alertConfig      map[string]interface{}
	alertNotifier    notifier.Notifier
	alertChannel     chan *notifier.Notification
	visualizer       types.Visualization
	visualizationConfig types.VisualizationConfig
	visualizationEnabled bool
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
	ctx, cancel := context.WithCancel(context.Background())
	return &FlinkProcessor{
		windows:          make(map[string]*WindowState),
		stopChan:         make(chan struct{}),
		parallelism:      defaultParallelism,
		operators:        make(map[string]bool),
		sinks:            make([]map[string]interface{}, 0),
		sources:          make([]map[string]interface{}, 0),
		checkpoint:       make(map[string]interface{}),
		metrics:          make(map[string]interface{}),
		anomalyDetectors: make(map[string]fault.FaultDetector),
		anomalyConfig:    make(map[string]interface{}),
		alertConfig:      make(map[string]interface{}),
		alertChannel:     make(chan *notifier.Notification, 100),
		visualizationEnabled: false,
		ctx:              ctx,
		cancel:           cancel,
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

	// 读取Flink配置
	f.loadFlinkConfig()

	// 读取Kafka配置
	f.loadKafkaConfig()

	// 初始化Kafka
	if err := f.initKafka(); err != nil {
		return err
	}

	// 初始化异常检测
	f.initAnomalyDetection()

	// 初始化告警机制
	f.initAlerting()

	// 初始化可视化
	f.initVisualization()

	fmt.Printf("Initializing Flink processor with config: %+v\n", config)
	fmt.Printf("  Window Type: %s\n", f.windowType)
	fmt.Printf("  Window Size: %v\n", f.windowSize)
	fmt.Printf("  Slide Size: %v\n", f.slideSize)
	fmt.Printf("  Parallelism: %d\n", f.parallelism)
	fmt.Printf("  Job Name: %s\n", f.jobName)
	fmt.Printf("  Enabled: %v\n", f.enabled)
	fmt.Printf("  Kafka Brokers: %v\n", f.kafkaConfig.Brokers)
	fmt.Printf("  Anomaly Detectors: %d\n", len(f.anomalyDetectors))
	fmt.Printf("  Alerting Enabled: %v\n", f.operators["alerting"])

	f.isRunning = false
	f.started = true

	return nil
}

func (f *FlinkProcessor) initAlerting() {
	// 从Flink配置中读取告警配置
	if viper.IsSet("flink.alerting") {
		f.alertConfig = viper.Get("flink.alerting").(map[string]interface{})
	}

	// 初始化告警通道消费者
	go f.processAlerts()

	fmt.Println("Initialized alerting mechanism")
}

func (f *FlinkProcessor) initVisualization() {
	// 从Flink配置中读取可视化配置
	if viper.IsSet("flink.visualization") {
		visConfig := viper.Get("flink.visualization").(map[string]interface{})
		
		// 检查可视化是否启用
		if enabled, ok := visConfig["enabled"].(bool); ok && enabled {
			f.visualizationEnabled = true
			
			// 创建可视化配置
			f.visualizationConfig = types.VisualizationConfig{
				Type:    "basic",
				Host:    "localhost",
				Port:    8080,
				Options: visConfig,
			}
			
			// 初始化可视化器
			f.visualizer = visualization.NewBasicVisualizer()
			if err := f.visualizer.Init(f.visualizationConfig); err != nil {
				fmt.Printf("Error initializing visualizer: %v\n", err)
				f.visualizationEnabled = false
				return
			}
			
			// 创建默认仪表板
			f.createDefaultDashboard()
			fmt.Println("Initialized visualization mechanism")
		}
	}
}

func (f *FlinkProcessor) createDefaultDashboard() error {
	if !f.visualizationEnabled || f.visualizer == nil {
		return nil
	}
	
	// 创建Flink处理仪表板
	panels := []types.Panel{
		{
			ID:     "flink-metrics",
			Title:  "Flink Processing Metrics",
			Type:   "metrics",
			Data:   map[string]interface{}{},
			Options: map[string]interface{}{
				"refresh": 10,
				"unit":    "sec",
			},
		},
		{
			ID:     "anomaly-detection",
			Title:  "Anomaly Detection",
			Type:   "anomaly",
			Data:   map[string]interface{}{},
			Options: map[string]interface{}{
				"refresh": 5,
				"unit":    "sec",
			},
		},
		{
			ID:     "window-aggregations",
			Title:  "Window Aggregations",
			Type:   "aggregations",
			Data:   map[string]interface{}{},
			Options: map[string]interface{}{
				"refresh": 15,
				"unit":    "sec",
			},
		},
	}
	
	return f.visualizer.CreateDashboard("flink-processing", panels)
}

func (f *FlinkProcessor) processAlerts() {
	for {
		select {
		case <-f.ctx.Done():
			return
		case notification := <-f.alertChannel:
			f.sendAlert(notification)
		}
	}
}

func (f *FlinkProcessor) sendAlert(notification *notifier.Notification) {
	// 这里可以添加实际的告警发送逻辑
	// 目前只是打印告警信息
	fmt.Printf("Sending alert: %s - %s\n", notification.Subject, notification.Content)
	
	// 可以集成实际的告警通知器
	if f.alertNotifier != nil {
		result, err := f.alertNotifier.Send(f.ctx, notification)
		if err != nil {
			fmt.Printf("Error sending alert: %v\n", err)
		} else {
			fmt.Printf("Alert sent successfully: %s\n", result.Status)
		}
	}
}

func (f *FlinkProcessor) updateVisualization(stats map[string]interface{}, dataPoints []*types.DataPoint) {
	if !f.visualizationEnabled || f.visualizer == nil {
		return
	}
	
	// 更新Flink metrics面板
	metricsData := map[string]interface{}{
		"stats":       stats,
		"job_id":      f.jobID,
		"job_name":    f.jobName,
		"window_type": f.windowType,
		"window_size": f.windowSize.Seconds(),
		"slide_size":  f.slideSize.Seconds(),
		"timestamp":   time.Now(),
	}
	f.visualizer.UpdatePanel("flink-processing", "flink-metrics", metricsData)
	
	// 更新异常检测面板
	anomalyData := map[string]interface{}{
		"anomaly_count": stats["anomalies"],
		"detectors":     len(f.anomalyDetectors),
		"timestamp":     time.Now(),
	}
	f.visualizer.UpdatePanel("flink-processing", "anomaly-detection", anomalyData)
	
	// 更新窗口聚合面板
	windowData := map[string]interface{}{
		"window_count": len(f.windows),
		"timestamp":    time.Now(),
	}
	f.visualizer.UpdatePanel("flink-processing", "window-aggregations", windowData)
	
	// 如果有数据点，生成时间序列图表数据
	if len(dataPoints) > 0 && f.visualizer != nil {
		if basicVis, ok := f.visualizer.(*visualization.BasicVisualizer); ok {
			timeSeriesData := basicVis.GenerateTimeSeriesChart(dataPoints)
			f.visualizer.UpdatePanel("flink-processing", "flink-metrics", timeSeriesData)
		}
	}
}

func (f *FlinkProcessor) initAnomalyDetection() {
	// 从Flink配置中读取异常检测配置
	if viper.IsSet("flink.anomaly_detection") {
		f.anomalyConfig = viper.Get("flink.anomaly_detection").(map[string]interface{})
	}

	// 初始化默认的异常检测器
	// 这里可以根据配置创建不同类型的异常检测器
	// 示例：为常见的设备指标创建阈值检测器
	defaultMetrics := []string{"temperature", "humidity", "pressure", "current", "voltage", "power"}
	for _, metric := range defaultMetrics {
		detectorKey := fmt.Sprintf("default_%s", metric)
		detector := fault.NewThresholdDetector(
			"default",
			metric,
			0,      // 下界
			100,    // 上界
			5,      // 最小偏差
			60,     // 窗口大小
		)
		f.anomalyDetectors[detectorKey] = detector
	}

	fmt.Printf("Initialized %d anomaly detectors\n", len(f.anomalyDetectors))
}

func (f *FlinkProcessor) loadKafkaConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err == nil {
		f.kafkaConfig = mq.KafkaConfig{
			Brokers:     viper.GetStringSlice("kafka.brokers"),
			TopicPrefix: viper.GetString("kafka.topic_prefix"),
		}
	}
}

func (f *FlinkProcessor) initKafka() error {
	// 初始化Kafka producer
	for _, sink := range f.sinks {
		if sink["type"] == "kafka" {
			topic, ok := sink["topic"].(string)
			if ok {
				f.kafkaProducer = mq.NewKafkaProducer(f.kafkaConfig, topic)
				fmt.Printf("Kafka producer initialized for topic: %s\n", topic)
			}
		}
	}

	// 初始化Kafka consumer
	for _, source := range f.sources {
		if source["type"] == "kafka" {
			topic, ok := source["topic"].(string)
			groupID, okGroup := source["consumer_group"].(string)
			if ok && okGroup {
				f.kafkaConsumer = mq.NewKafkaConsumer(f.kafkaConfig, topic, groupID)
				fmt.Printf("Kafka consumer initialized for topic: %s, group: %s\n", topic, groupID)
			}
		}
	}

	return nil
}

func (f *FlinkProcessor) loadFlinkConfig() {
	viper.SetConfigName("flink-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err == nil {
		f.enabled = viper.GetBool("flink.enabled")
		f.jobName = viper.GetString("flink.job_name")
		
		if viper.IsSet("flink.parallelism") {
			f.parallelism = viper.GetInt("flink.parallelism")
		}
		
		if viper.IsSet("flink.window.type") {
			windowType := viper.GetString("flink.window.type")
			switch windowType {
			case "tumbling":
				f.windowType = TumblingWindow
			case "sliding":
				f.windowType = SlidingWindow
			case "session":
				f.windowType = SessionWindow
			}
		}
		
		if viper.IsSet("flink.window.size") {
			windowSize := viper.GetInt("flink.window.size")
			f.windowSize = time.Duration(windowSize) * time.Second
		}
		
		if viper.IsSet("flink.window.slide") {
			slideSize := viper.GetInt("flink.window.slide")
			f.slideSize = time.Duration(slideSize) * time.Second
		}
		
		// 加载operators
		if viper.IsSet("flink.operators") {
			operators := viper.Get("flink.operators").([]interface{})
			for _, op := range operators {
				if opMap, ok := op.(map[string]interface{}); ok {
					name, ok := opMap["name"].(string)
					enabled, okEnabled := opMap["enabled"].(bool)
					if ok && okEnabled {
						f.operators[name] = enabled
					}
				}
			}
		}
		
		// 加载sinks
		if viper.IsSet("flink.sinks") {
			f.sinks = viper.Get("flink.sinks").([]map[string]interface{})
		}
		
		// 加载sources
		if viper.IsSet("flink.sources") {
			f.sources = viper.Get("flink.sources").([]map[string]interface{})
		}
		
		// 加载metrics
		if viper.IsSet("flink.metrics") {
			f.metrics = viper.Get("flink.metrics").(map[string]interface{})
		}
		
		// 加载checkpoint
		if viper.IsSet("flink.checkpoint") {
			f.checkpoint = viper.Get("flink.checkpoint").(map[string]interface{})
		}
	}
}

func (f *FlinkProcessor) Process(data *types.BatchData) (*types.BatchData, error) {
	if !f.enabled {
		return data, nil
	}

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
	fmt.Printf("Starting Flink job: %s\n", f.jobName)
	fmt.Println("  Operators:")
	for opName, enabled := range f.operators {
		if enabled {
			fmt.Printf("    - %s\n", opName)
		}
	}

	fmt.Println("  Sources:")
	for _, source := range f.sources {
		fmt.Printf("    - %s\n", source["type"])
	}

	fmt.Println("  Sinks:")
	for _, sink := range f.sinks {
		fmt.Printf("    - %s\n", sink["type"])
	}

	f.jobID = fmt.Sprintf("job_flink_%d", time.Now().UnixNano())
	f.isRunning = true

	// 启动Kafka消费者
	if f.kafkaConsumer != nil {
		go f.startKafkaConsumer()
	}

	fmt.Printf("Flink job started with ID: %s\n", f.jobID)

	return nil
}

func (f *FlinkProcessor) startKafkaConsumer() {
	if f.kafkaConsumer == nil {
		return
	}

	fmt.Println("Starting Kafka consumer...")

	handler := func(ctx context.Context, msg kafka.Message) error {
		// 解析消息
		var batchData types.BatchData
		if err := json.Unmarshal(msg.Value, &batchData); err != nil {
			fmt.Printf("Error unmarshaling message: %v\n", err)
			return nil
		}

		// 处理数据
		processedData, err := f.Process(&batchData)
		if err != nil {
			fmt.Printf("Error processing data: %v\n", err)
			return nil
		}

		// 写入处理结果到Kafka
		if f.kafkaProducer != nil {
			if err := f.writeToKafka(processedData); err != nil {
				fmt.Printf("Error writing to Kafka: %v\n", err)
			}
		}

		return nil
	}

	if err := f.kafkaConsumer.Consume(f.ctx, handler); err != nil {
		fmt.Printf("Error consuming Kafka messages: %v\n", err)
	}
}

func (f *FlinkProcessor) writeToKafka(data *types.BatchData) error {
	if f.kafkaProducer == nil {
		return fmt.Errorf("kafka producer not initialized")
	}

	// 序列化数据
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// 发送到Kafka
	key := fmt.Sprintf("flink_%d", time.Now().UnixNano())
	if err := f.kafkaProducer.SendBytes(f.ctx, []byte(key), dataBytes); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (f *FlinkProcessor) processBatchData(data *types.BatchData) (*types.BatchData, error) {
	startTime := time.Now()
	fmt.Printf("Processing batch data with Flink, %d data points\n", len(data.DataPoints))

	processedDataPoints := make([]*types.DataPoint, 0, len(data.DataPoints))
	stats := map[string]int{
		"total":     len(data.DataPoints),
		"cleaned":   0,
		"invalid":   0,
		"anomalies": 0,
	}

	for _, dp := range data.DataPoints {
		processedDP := f.processDataPoint(dp, "batch", &stats)
		if processedDP != nil {
			processedDataPoints = append(processedDataPoints, processedDP)
		}
	}

	processingTime := time.Since(startTime)
	_ = processingTime

	fmt.Printf("Batch processing complete: cleaned=%d, invalid=%d, anomalies=%d, duration=%v\n",
		stats["cleaned"], stats["invalid"], stats["anomalies"], processingTime)

	// 转换stats为map[string]interface{}
	statsMap := make(map[string]interface{})
	for k, v := range stats {
		statsMap[k] = v
	}
	
	// 更新可视化
	f.updateVisualization(statsMap, processedDataPoints)

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
				"job_name":        f.jobName,
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
		"total":     len(data.DataPoints),
		"cleaned":   0,
		"invalid":   0,
		"windowed":  0,
		"anomalies": 0,
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

	fmt.Printf("Stream processing complete: cleaned=%d, invalid=%d, windowed=%d, anomalies=%d, duration=%v\n",
		stats["cleaned"], stats["invalid"], stats["windowed"], stats["anomalies"], processingTime)

	// 转换stats为map[string]interface{}
	statsMap := make(map[string]interface{})
	for k, v := range stats {
		statsMap[k] = v
	}
	
	// 更新可视化
	f.updateVisualization(statsMap, processedDataPoints)

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
				"job_name":        f.jobName,
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

	// 数据清洗
	if f.operators["data_cleaning"] {
		if dp.Value < 0 {
			dp.Value = 0
			(*stats)["cleaned"]++
		}

		if math.IsNaN(dp.Value) || math.IsInf(dp.Value, 0) {
			(*stats)["invalid"]++
			return nil
		}
	}

	// 质量验证
	if f.operators["quality_validation"] {
		// 这里可以添加更复杂的质量验证逻辑
	}

	// 异常检测
	if f.operators["anomaly_detection"] {
		anomalies := f.detectAnomalies(dp)
		if len(anomalies) > 0 {
			dp.Attributes["anomalies"] = anomalies
			dp.Attributes["has_anomaly"] = true
			(*stats)["anomalies"]++
			
			// 打印异常信息
			for _, anomaly := range anomalies {
				fmt.Printf("Anomaly detected: Device=%s, Metric=%s, Value=%f, Severity=%s\n",
					anomaly.DeviceID, anomaly.Metric, anomaly.Value, anomaly.Severity)
			}
		} else {
			dp.Attributes["has_anomaly"] = false
		}
	}

	dp.Attributes["processed"] = true
	dp.Attributes["processed_at"] = time.Now()
	dp.Attributes["processor"] = "flink"
	dp.Attributes["mode"] = mode
	dp.Attributes["job_id"] = f.jobID
	dp.Attributes["job_name"] = f.jobName

	if dp.Tags == nil {
		dp.Tags = make(map[string]string)
	}
	dp.Tags["processor"] = "flink"
	dp.Tags["mode"] = mode

	(*stats)["cleaned"]++
	return dp
}

func (f *FlinkProcessor) detectAnomalies(dp *types.DataPoint) []*fault.Anomaly {
	var allAnomalies []*fault.Anomaly

	// 转换数据点为异常检测器需要的格式
	timeSeriesData := &fault.TimeSeriesData{
		DeviceID:  dp.DeviceID,
		Metric:    dp.Metric,
		Value:     dp.Value,
		Timestamp: dp.Timestamp,
	}

	// 使用所有适用的检测器进行检测
	for key, detector := range f.anomalyDetectors {
		// 检查检测器是否适用于当前数据点
		if detector.GetDetectorInfo().Parameters["metric"] == dp.Metric {
			anomalies, err := detector.Detect(f.ctx, []*fault.TimeSeriesData{timeSeriesData})
			if err == nil && len(anomalies) > 0 {
				allAnomalies = append(allAnomalies, anomalies...)
				
				// 为每个异常创建告警通知
			for _, anomaly := range anomalies {
				// 转换严重程度到通知优先级
				priority := notifier.PriorityNormal
				switch anomaly.Severity {
				case "critical":
					priority = notifier.PriorityCritical
				case "high":
					priority = notifier.PriorityHigh
				case "low":
					priority = notifier.PriorityLow
				}
				
				notification := &notifier.Notification{
					ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
					AlarmID:   fmt.Sprintf("anomaly_%s_%s", anomaly.DeviceID, anomaly.Metric),
					Channel:   notifier.ChannelInternal,
					Priority:  priority,
					Status:    notifier.StatusPending,
					Subject:   fmt.Sprintf("Anomaly Detected: %s", anomaly.Metric),
					Content:   fmt.Sprintf("Device: %s, Metric: %s, Value: %f, Severity: %s, Timestamp: %s",
						anomaly.DeviceID, anomaly.Metric, anomaly.Value, anomaly.Severity, anomaly.Timestamp),
					CreatedAt: time.Now(),
					Tags: map[string]string{
						"source":    "flink_processor",
						"device_id": anomaly.DeviceID,
						"metric":    anomaly.Metric,
						"detector":  key,
						"job_id":    f.jobID,
						"job_name":  f.jobName,
					},
				}
				
				// 发送告警到通道
				select {
				case f.alertChannel <- notification:
					// 告警已发送
				default:
					// 通道已满，丢弃告警
					fmt.Printf("Alert channel full, dropping alert for %s\n", anomaly.Metric)
				}
			}
			}
		}
	}

	return allAnomalies
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
		"job_name":     f.jobName,
		"is_running":   f.isRunning,
		"started":      f.started,
		"enabled":      f.enabled,
		"window_type":  f.windowType,
		"window_size":  f.windowSize.Seconds(),
		"slide_size":   f.slideSize.Seconds(),
		"parallelism":  f.parallelism,
		"window_count": len(f.windows),
		"operators":    f.operators,
		"sources":      f.sources,
		"sinks":        f.sinks,
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

	// 取消context
	f.cancel()

	// 关闭Kafka producer
	if f.kafkaProducer != nil {
		if err := f.kafkaProducer.Close(); err != nil {
			fmt.Printf("Error closing Kafka producer: %v\n", err)
		}
	}

	// 关闭Kafka consumer
	if f.kafkaConsumer != nil {
		if err := f.kafkaConsumer.Close(); err != nil {
			fmt.Printf("Error closing Kafka consumer: %v\n", err)
		}
	}

	// 关闭可视化器
	if f.visualizer != nil {
		if err := f.visualizer.Close(); err != nil {
			fmt.Printf("Error closing visualizer: %v\n", err)
		}
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
