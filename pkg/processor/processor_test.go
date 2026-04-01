package processor

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

// 测试质量标记器
func TestQualityMarker(t *testing.T) {
	logger := zap.NewNop()
	marker := NewQualityMarker(logger)

	// 测试标记
	info := marker.Mark(100.0, QualityGood, "test")
	if info.Code != QualityGood {
		t.Errorf("Expected QualityGood, got %v", info.Code)
	}

	// 测试评估
	info = marker.Evaluate(nil)
	if info.Code&QualityBad == 0 {
		t.Errorf("Expected QualityBad for nil value")
	}

	// 测试组合
	code := marker.Combine(QualityBad, QualityReasonOutOfRange)
	if code&QualityBad == 0 {
		t.Errorf("Expected QualityBad in combined code")
	}

	// 测试质量等级
	level := marker.GetLevel(QualityBad)
	if level != QualityLevelBad {
		t.Errorf("Expected QualityLevelBad, got %v", level)
	}
}

// 测试范围校验器
func TestRangeValidator(t *testing.T) {
	logger := zap.NewNop()
	validator := NewRangeValidator(RangeValidatorConfig{
		Name:      "test_range",
		MinValue:  0,
		MaxValue:  100,
		Inclusive: true,
		Logger:    logger,
	})

	// 测试有效值
	result := validator.Validate(50.0)
	if !result.Valid {
		t.Errorf("Expected valid for value 50.0")
	}

	// 测试超出范围
	result = validator.Validate(150.0)
	if result.Valid {
		t.Errorf("Expected invalid for value 150.0")
	}
	if result.Quality&QualityReasonOutOfRange == 0 {
		t.Errorf("Expected QualityReasonOutOfRange")
	}

	// 测试边界值
	result = validator.Validate(0.0)
	if !result.Valid {
		t.Errorf("Expected valid for value 0.0 (inclusive)")
	}

	result = validator.Validate(100.0)
	if !result.Valid {
		t.Errorf("Expected valid for value 100.0 (inclusive)")
	}
}

// 测试空值校验器
func TestNullValidator(t *testing.T) {
	logger := zap.NewNop()
	validator := NewNullValidator(NullValidatorConfig{
		Name:          "test_null",
		AllowNull:     false,
		AllowEmptyStr: false,
		AllowZero:     true,
		Logger:        logger,
	})

	// 测试nil值
	result := validator.Validate(nil)
	if result.Valid {
		t.Errorf("Expected invalid for nil value")
	}

	// 测试空字符串
	result = validator.Validate("")
	if result.Valid {
		t.Errorf("Expected invalid for empty string")
	}

	// 测试零值
	result = validator.Validate(0.0)
	if !result.Valid {
		t.Errorf("Expected valid for zero value (allowZero=true)")
	}
}

// 测试校验器链
func TestValidatorChain(t *testing.T) {
	logger := zap.NewNop()
	chain := NewValidatorChain(ValidatorChainConfig{
		Name:       "test_chain",
		StopOnFail: true,
		Logger:     logger,
	})

	// 添加校验器
	chain.AddValidator(NewRangeValidator(RangeValidatorConfig{
		Name:      "range",
		MinValue:  0,
		MaxValue:  100,
		Inclusive: true,
		Logger:    logger,
	}))

	chain.AddValidator(NewNullValidator(NullValidatorConfig{
		Name:      "null",
		AllowNull: false,
		Logger:    logger,
	}))

	// 测试有效值
	result := chain.Validate(50.0)
	if !result.Valid {
		t.Errorf("Expected valid for value 50.0")
	}

	// 测试无效值
	result = chain.Validate(nil)
	if result.Valid {
		t.Errorf("Expected invalid for nil value")
	}
}

// 测试移动平均滤波器
func TestMovingAverageFilter(t *testing.T) {
	logger := zap.NewNop()
	filter := NewMovingAverageFilter(MovingAverageFilterConfig{
		Name:   "test_ma",
		Window: 3,
		Logger: logger,
	})

	// 测试滤波
	values := []float64{10.0, 20.0, 30.0, 40.0}
	expectedAvg := []float64{10.0, 15.0, 20.0, 30.0}

	for i, v := range values {
		result := filter.Filter(v)
		expected := expectedAvg[i]
		if result.Value != expected {
			t.Errorf("At index %d, expected %.2f, got %.2f", i, expected, result.Value)
		}
	}
}

// 测试中值滤波器
func TestMedianFilter(t *testing.T) {
	logger := zap.NewNop()
	filter := NewMedianFilter(MedianFilterConfig{
		Name:   "test_median",
		Window: 3,
		Logger: logger,
	})

	// 测试滤波
	values := []float64{10.0, 100.0, 20.0, 30.0, 40.0}

	for _, v := range values {
		result := filter.Filter(v)
		t.Logf("Median filter: raw=%.2f, filtered=%.2f", v, result.Value)
	}
}

// 测试卡尔曼滤波器
func TestKalmanFilter(t *testing.T) {
	logger := zap.NewNop()
	filter := NewKalmanFilter(KalmanFilterConfig{
		Name:         "test_kalman",
		ProcessNoise: 0.01,
		MeasureNoise: 0.1,
		Logger:       logger,
	})

	// 测试滤波
	values := []float64{10.0, 10.5, 9.8, 10.2, 10.1}

	for _, v := range values {
		result := filter.Filter(v)
		t.Logf("Kalman filter: raw=%.2f, filtered=%.2f", v, result.Value)
	}
}

// 测试限幅滤波器
func TestLimitFilter(t *testing.T) {
	logger := zap.NewNop()
	filter := NewLimitFilter(LimitFilterConfig{
		Name:      "test_limit",
		MaxChange: 10.0,
		Logger:    logger,
	})

	// 测试滤波
	values := []float64{10.0, 25.0, 35.0, 30.0}

	for _, v := range values {
		result := filter.Filter(v)
		t.Logf("Limit filter: raw=%.2f, filtered=%.2f, filtered=%v", v, result.Value, result.Filtered)
	}
}

// 测试线性转换器
func TestLinearScaler(t *testing.T) {
	logger := zap.NewNop()
	scaler := NewLinearScaler(LinearScalerConfig{
		Name:      "test_linear",
		Slope:     2.0,
		Intercept: 10.0,
		Logger:    logger,
	})

	// 测试转换
	result := scaler.Scale(5.0)
	expected := 20.0 // 2*5 + 10
	if result.Value != expected {
		t.Errorf("Expected %.2f, got %.2f", expected, result.Value)
	}

	// 测试反向转换
	result = scaler.Inverse(20.0)
	if result.Value != 5.0 {
		t.Errorf("Expected 5.0, got %.2f", result.Value)
	}
}

// 测试查表转换器
func TestLookupTableScaler(t *testing.T) {
	logger := zap.NewNop()
	table := []LookupEntry{
		{Input: 0, Output: 0},
		{Input: 50, Output: 100},
		{Input: 100, Output: 200},
	}

	scaler := NewLookupTableScaler(LookupTableScalerConfig{
		Name:        "test_lookup",
		Table:       table,
		Extrapolate: false,
		Logger:      logger,
	})

	// 测试插值
	result := scaler.Scale(25.0)
	expected := 50.0 // 线性插值
	if result.Value != expected {
		t.Errorf("Expected %.2f, got %.2f", expected, result.Value)
	}

	// 测试精确值
	result = scaler.Scale(50.0)
	if result.Value != 100.0 {
		t.Errorf("Expected 100.0, got %.2f", result.Value)
	}
}

// 测试变位检测器
func TestChangeDetector(t *testing.T) {
	logger := zap.NewNop()
	detector := NewBasicChangeDetector(BasicChangeDetectorConfig{
		Name:   "test_change",
		Logger: logger,
	})

	// 测试变位检测
	result := detector.Detect(10.0)
	if result.Changed {
		t.Errorf("First value should not be a change")
	}

	result = detector.Detect(20.0)
	if !result.Changed {
		t.Errorf("Value change should be detected")
	}
	if result.Event == nil {
		t.Errorf("Change event should not be nil")
	} else {
		if result.Event.Type != ChangeTypeRising {
			t.Errorf("Expected rising change, got %v", result.Event.Type)
		}
	}
}

// 测试死区变位检测器
func TestDeadbandChangeDetector(t *testing.T) {
	logger := zap.NewNop()
	detector := NewDeadbandChangeDetector(DeadbandChangeDetectorConfig{
		Name:     "test_deadband",
		Deadband: 5.0,
		Logger:   logger,
	})

	// 测试死区
	result := detector.Detect(10.0)
	result = detector.Detect(12.0) // 变化小于死区
	if result.Changed {
		t.Errorf("Change within deadband should not be detected")
	}

	result = detector.Detect(20.0) // 变化大于死区
	if !result.Changed {
		t.Errorf("Change beyond deadband should be detected")
	}
}

// 测试防抖处理器
func TestDebouncer(t *testing.T) {
	logger := zap.NewNop()
	debouncer := NewDebouncer(DebouncerConfig{
		Name:         "test_debounce",
		DebounceTime: 100 * time.Millisecond,
		Logger:       logger,
	})

	// 测试防抖
	result := debouncer.Process(10.0)
	time.Sleep(50 * time.Millisecond)
	result = debouncer.Process(20.0)
	
	// 等待防抖时间
	time.Sleep(150 * time.Millisecond)
	result = debouncer.Process(20.0)
	
	if !result.Changed {
		t.Logf("Debounced value: %.2f", result.Value)
	}
}

// 测试处理管道
func TestPipeline(t *testing.T) {
	logger := zap.NewNop()

	// 创建管道
	pipeline := NewPipeline(PipelineConfig{
		Name:   "test_pipeline",
		Logger: logger,
	})

	// 添加阶段
	pipeline.AddStage(NewValidationStage("validation", NewRangeValidator(RangeValidatorConfig{
		Name:      "range",
		MinValue:  0,
		MaxValue:  100,
		Inclusive: true,
		Logger:    logger,
	}), logger))

	pipeline.AddStage(NewFilterStage("filter", NewMovingAverageFilter(MovingAverageFilterConfig{
		Name:   "ma",
		Window: 3,
		Logger: logger,
	}), logger))

	pipeline.AddStage(NewScaleStage("scale", NewLinearScaler(LinearScalerConfig{
		Name:      "linear",
		Slope:     1.0,
		Intercept: 0.0,
		Logger:    logger,
	}), logger))

	// 处理数据
	ctx := context.Background()
	data := &ProcessedData{
		PointID:   "test_point",
		Value:     50.0,
		Timestamp: time.Now(),
	}

	result, err := pipeline.Process(ctx, data)
	if err != nil {
		t.Errorf("Pipeline process failed: %v", err)
	}

	t.Logf("Processed value: %.2f, quality: %d", result.Value, result.Quality)

	// 获取统计信息
	stats := pipeline.GetStatistics()
	t.Logf("Pipeline stats: processed=%d, success=%d, failed=%d",
		stats.TotalProcessed, stats.TotalSuccess, stats.TotalFailed)
}

// 测试管道构建器
func TestPipelineBuilder(t *testing.T) {
	logger := zap.NewNop()

	pipeline := NewPipelineBuilder("test_builder", logger).
		AddValidationStage("validation", NewRangeValidator(RangeValidatorConfig{
			Name:      "range",
			MinValue:  0,
			MaxValue:  100,
			Inclusive: true,
			Logger:    logger,
		})).
		AddFilterStage("filter", NewMovingAverageFilter(MovingAverageFilterConfig{
			Name:   "ma",
			Window: 3,
			Logger: logger,
		})).
		Build()

	if pipeline == nil {
		t.Errorf("Pipeline builder returned nil")
	}

	stats := pipeline.GetStatistics()
	t.Logf("Built pipeline with %d stages", stats.StageCount)
}

// 测试批量处理
func TestBatchProcessing(t *testing.T) {
	logger := zap.NewNop()

	pipeline := NewPipelineBuilder("test_batch", logger).
		WithParallel(4).
		AddValidationStage("validation", NewRangeValidator(RangeValidatorConfig{
			Name:      "range",
			MinValue:  0,
			MaxValue:  100,
			Inclusive: true,
			Logger:    logger,
		})).
		Build()

	// 创建批量数据
	dataList := make([]*ProcessedData, 10)
	for i := 0; i < 10; i++ {
		dataList[i] = &ProcessedData{
			PointID:   "test_point",
			Value:     float64(i * 10),
			Timestamp: time.Now(),
		}
	}

	// 批量处理
	ctx := context.Background()
	results, err := pipeline.ProcessBatch(ctx, dataList)
	if err != nil {
		t.Errorf("Batch process failed: %v", err)
	}

	if len(results) != 10 {
		t.Errorf("Expected 10 results, got %d", len(results))
	}

	stats := pipeline.GetStatistics()
	t.Logf("Batch stats: processed=%d, success=%d", stats.TotalProcessed, stats.TotalSuccess)
}

// 测试SOE记录器
func TestSOERecorder(t *testing.T) {
	logger := zap.NewNop()
	recorder := NewSOERecorder(SOERecorderConfig{
		Name:      "test_soe",
		MaxEvents: 100,
		Logger:    logger,
	})
	defer recorder.Stop()

	// 记录事件
	event := ChangeEvent{
		Type:      ChangeTypeRising,
		OldValue:  0.0,
		NewValue:  1.0,
		Timestamp: time.Now(),
		Quality:   QualityGood,
	}

	recorder.Record(event, "point_001", "Test Point", 1)

	// 等待事件处理
	time.Sleep(100 * time.Millisecond)

	// 获取事件
	events := recorder.GetEvents(10)
	if len(events) == 0 {
		t.Errorf("Expected at least 1 event")
	}

	t.Logf("Recorded %d events", len(events))
}

// 测试数据处理器
func TestDataProcessor(t *testing.T) {
	logger := zap.NewNop()
	processor := NewDataProcessor(logger)

	// 创建管道
	pipeline := NewPipelineBuilder("test_pipeline", logger).
		AddValidationStage("validation", NewRangeValidator(RangeValidatorConfig{
			Name:      "range",
			MinValue:  0,
			MaxValue:  100,
			Inclusive: true,
			Logger:    logger,
		})).
		Build()

	// 添加管道
	processor.AddPipeline("default", pipeline)

	// 处理数据
	ctx := context.Background()
	data := &ProcessedData{
		PointID:   "test_point",
		Value:     50.0,
		Timestamp: time.Now(),
	}

	result, err := processor.Process(ctx, "default", data)
	if err != nil {
		t.Errorf("Processor failed: %v", err)
	}

	t.Logf("Processed value: %.2f", result.Value)

	// 获取统计
	stats := processor.GetAllStatistics()
	t.Logf("Processor has %d pipelines", len(stats))
}
