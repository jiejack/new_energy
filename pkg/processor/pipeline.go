package processor

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// ProcessedData 处理后的数据
type ProcessedData struct {
	PointID      string      `json:"point_id"`      // 测点ID
	RawValue     interface{} `json:"raw_value"`     // 原始值
	Value        float64     `json:"value"`         // 处理后的值
	Quality      QualityCode `json:"quality"`       // 质量码
	Timestamp    time.Time   `json:"timestamp"`     // 时间戳
	StageResults map[string]interface{} `json:"stage_results"` // 各阶段结果
}

// StageResult 阶段处理结果
type StageResult struct {
	StageName string      `json:"stage_name"`
	Success   bool        `json:"success"`
	Value     interface{} `json:"value"`
	Quality   QualityCode `json:"quality"`
	Error     string      `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
}

// Stage 处理阶段接口
type Stage interface {
	// Name 获取阶段名称
	Name() string
	// Process 处理数据
	Process(ctx context.Context, data *ProcessedData) (*ProcessedData, error)
}

// Pipeline 数据处理管道
type Pipeline struct {
	name        string
	stages      []Stage
	parallel    bool        // 是否并行处理
	workers     int         // 并行工作数
	mu          sync.RWMutex
	logger      *zap.Logger
	
	// 统计信息
	totalProcessed int64
	totalSuccess   int64
	totalFailed    int64
	totalDuration  int64 // 纳秒
}

// PipelineConfig 管道配置
type PipelineConfig struct {
	Name     string
	Parallel bool
	Workers  int
	Logger   *zap.Logger
}

// NewPipeline 创建处理管道
func NewPipeline(config PipelineConfig) *Pipeline {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	if config.Workers <= 0 {
		config.Workers = 4
	}
	return &Pipeline{
		name:     config.Name,
		stages:   make([]Stage, 0),
		parallel: config.Parallel,
		workers:  config.Workers,
		logger:   config.Logger,
	}
}

// AddStage 添加处理阶段
func (p *Pipeline) AddStage(stage Stage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stages = append(p.stages, stage)
}

// RemoveStage 移除处理阶段
func (p *Pipeline) RemoveStage(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for i, s := range p.stages {
		if s.Name() == name {
			p.stages = append(p.stages[:i], p.stages[i+1:]...)
			break
		}
	}
}

// Process 处理数据
func (p *Pipeline) Process(ctx context.Context, data *ProcessedData) (*ProcessedData, error) {
	start := time.Now()
	
	p.mu.RLock()
	stages := make([]Stage, len(p.stages))
	copy(stages, p.stages)
	p.mu.RUnlock()
	
	if data.StageResults == nil {
		data.StageResults = make(map[string]interface{})
	}
	
	// 顺序处理各阶段
	for _, stage := range stages {
		stageStart := time.Now()
		
		result, err := stage.Process(ctx, data)
		if err != nil {
			atomic.AddInt64(&p.totalFailed, 1)
			atomic.AddInt64(&p.totalProcessed, 1)
			
			p.logger.Error("stage process failed",
				zap.String("stage", stage.Name()),
				zap.Error(err),
			)
			
			// 记录失败结果
			data.StageResults[stage.Name()] = StageResult{
				StageName: stage.Name(),
				Success:   false,
				Error:     err.Error(),
				Duration:  time.Since(stageStart),
			}
			
			return data, err
		}
		
		// 记录成功结果
		data.StageResults[stage.Name()] = StageResult{
			StageName: stage.Name(),
			Success:   true,
			Value:     result.Value,
			Duration:  time.Since(stageStart),
		}
		
		data = result
	}
	
	// 更新统计
	atomic.AddInt64(&p.totalSuccess, 1)
	atomic.AddInt64(&p.totalProcessed, 1)
	atomic.AddInt64(&p.totalDuration, int64(time.Since(start)))
	
	return data, nil
}

// ProcessBatch 批量处理
func (p *Pipeline) ProcessBatch(ctx context.Context, dataList []*ProcessedData) ([]*ProcessedData, error) {
	if p.parallel {
		return p.processBatchParallel(ctx, dataList)
	}
	return p.processBatchSequential(ctx, dataList)
}

// processBatchSequential 顺序批量处理
func (p *Pipeline) processBatchSequential(ctx context.Context, dataList []*ProcessedData) ([]*ProcessedData, error) {
	results := make([]*ProcessedData, len(dataList))
	
	for i, data := range dataList {
		result, err := p.Process(ctx, data)
		if err != nil {
			p.logger.Warn("batch process failed",
				zap.Int("index", i),
				zap.Error(err),
			)
		}
		results[i] = result
	}
	
	return results, nil
}

// processBatchParallel 并行批量处理
func (p *Pipeline) processBatchParallel(ctx context.Context, dataList []*ProcessedData) ([]*ProcessedData, error) {
	results := make([]*ProcessedData, len(dataList))
	
	var wg sync.WaitGroup
	chunkSize := (len(dataList) + p.workers - 1) / p.workers
	
	for i := 0; i < p.workers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(dataList) {
			end = len(dataList)
		}
		if start >= len(dataList) {
			break
		}
		
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				result, err := p.Process(ctx, dataList[j])
				if err != nil {
					p.logger.Warn("parallel batch process failed",
						zap.Int("index", j),
						zap.Error(err),
					)
				}
				results[j] = result
			}
		}(start, end)
	}
	
	wg.Wait()
	return results, nil
}

// GetStatistics 获取统计信息
func (p *Pipeline) GetStatistics() PipelineStatistics {
	return PipelineStatistics{
		Name:          p.name,
		TotalProcessed: atomic.LoadInt64(&p.totalProcessed),
		TotalSuccess:   atomic.LoadInt64(&p.totalSuccess),
		TotalFailed:    atomic.LoadInt64(&p.totalFailed),
		TotalDuration:  time.Duration(atomic.LoadInt64(&p.totalDuration)),
		StageCount:     len(p.stages),
	}
}

// ResetStatistics 重置统计信息
func (p *Pipeline) ResetStatistics() {
	atomic.StoreInt64(&p.totalProcessed, 0)
	atomic.StoreInt64(&p.totalSuccess, 0)
	atomic.StoreInt64(&p.totalFailed, 0)
	atomic.StoreInt64(&p.totalDuration, 0)
}

// PipelineStatistics 管道统计信息
type PipelineStatistics struct {
	Name           string        `json:"name"`
	TotalProcessed int64         `json:"total_processed"`
	TotalSuccess   int64         `json:"total_success"`
	TotalFailed    int64         `json:"total_failed"`
	TotalDuration  time.Duration `json:"total_duration"`
	StageCount     int           `json:"stage_count"`
}

// 预定义的处理阶段

// ValidationStage 校验阶段
type ValidationStage struct {
	name      string
	validator Validator
	logger    *zap.Logger
}

// NewValidationStage 创建校验阶段
func NewValidationStage(name string, validator Validator, logger *zap.Logger) *ValidationStage {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ValidationStage{
		name:      name,
		validator: validator,
		logger:    logger,
	}
}

// Name 获取阶段名称
func (s *ValidationStage) Name() string {
	return s.name
}

// Process 处理数据
func (s *ValidationStage) Process(ctx context.Context, data *ProcessedData) (*ProcessedData, error) {
	result := s.validator.Validate(data.Value)
	
	if !result.Valid {
		data.Quality = result.Quality
		return data, fmt.Errorf("validation failed: %v", result.Errors)
	}
	
	data.Quality = result.Quality
	return data, nil
}

// FilterStage 滤波阶段
type FilterStage struct {
	name   string
	filter Filter
	logger *zap.Logger
}

// NewFilterStage 创建滤波阶段
func NewFilterStage(name string, filter Filter, logger *zap.Logger) *FilterStage {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &FilterStage{
		name:   name,
		filter: filter,
		logger: logger,
	}
}

// Name 获取阶段名称
func (s *FilterStage) Name() string {
	return s.name
}

// Process 处理数据
func (s *FilterStage) Process(ctx context.Context, data *ProcessedData) (*ProcessedData, error) {
	result := s.filter.Filter(data.Value)
	
	data.Value = result.Value
	data.Quality = result.Quality
	
	return data, nil
}

// ScaleStage 量程转换阶段
type ScaleStage struct {
	name   string
	scaler Scaler
	logger *zap.Logger
}

// NewScaleStage 创建量程转换阶段
func NewScaleStage(name string, scaler Scaler, logger *zap.Logger) *ScaleStage {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ScaleStage{
		name:   name,
		scaler: scaler,
		logger: logger,
	}
}

// Name 获取阶段名称
func (s *ScaleStage) Name() string {
	return s.name
}

// Process 处理数据
func (s *ScaleStage) Process(ctx context.Context, data *ProcessedData) (*ProcessedData, error) {
	result := s.scaler.Scale(data.Value)
	
	data.Value = result.Value
	data.Quality = result.Quality
	
	return data, nil
}

// ChangeDetectionStage 变位检测阶段
type ChangeDetectionStage struct {
	name      string
	detector  ChangeDetector
	logger    *zap.Logger
}

// NewChangeDetectionStage 创建变位检测阶段
func NewChangeDetectionStage(name string, detector ChangeDetector, logger *zap.Logger) *ChangeDetectionStage {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ChangeDetectionStage{
		name:     name,
		detector: detector,
		logger:   logger,
	}
}

// Name 获取阶段名称
func (s *ChangeDetectionStage) Name() string {
	return s.name
}

// Process 处理数据
func (s *ChangeDetectionStage) Process(ctx context.Context, data *ProcessedData) (*ProcessedData, error) {
	result := s.detector.Detect(data.Value)
	
	if result.Changed && result.Event != nil {
		data.StageResults["change_event"] = result.Event
	}
	
	data.Quality = result.Quality
	
	return data, nil
}

// QualityMarkStage 质量标记阶段
type QualityMarkStage struct {
	name   string
	marker QualityMarker
	logger *zap.Logger
}

// NewQualityMarkStage 创建质量标记阶段
func NewQualityMarkStage(name string, marker QualityMarker, logger *zap.Logger) *QualityMarkStage {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &QualityMarkStage{
		name:   name,
		marker: marker,
		logger: logger,
	}
}

// Name 获取阶段名称
func (s *QualityMarkStage) Name() string {
	return s.name
}

// Process 处理数据
func (s *QualityMarkStage) Process(ctx context.Context, data *ProcessedData) (*ProcessedData, error) {
	info := s.marker.Evaluate(data.Value)
	
	data.Quality = s.marker.Combine(data.Quality, info.Code)
	
	return data, nil
}

// CustomStage 自定义处理阶段
type CustomStage struct {
	name    string
	process func(ctx context.Context, data *ProcessedData) (*ProcessedData, error)
	logger  *zap.Logger
}

// NewCustomStage 创建自定义处理阶段
func NewCustomStage(name string, process func(ctx context.Context, data *ProcessedData) (*ProcessedData, error), logger *zap.Logger) *CustomStage {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &CustomStage{
		name:    name,
		process: process,
		logger:  logger,
	}
}

// Name 获取阶段名称
func (s *CustomStage) Name() string {
	return s.name
}

// Process 处理数据
func (s *CustomStage) Process(ctx context.Context, data *ProcessedData) (*ProcessedData, error) {
	return s.process(ctx, data)
}

// PipelineBuilder 管道构建器
type PipelineBuilder struct {
	pipeline *Pipeline
	logger   *zap.Logger
}

// NewPipelineBuilder 创建管道构建器
func NewPipelineBuilder(name string, logger *zap.Logger) *PipelineBuilder {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &PipelineBuilder{
		pipeline: NewPipeline(PipelineConfig{
			Name:   name,
			Logger: logger,
		}),
		logger: logger,
	}
}

// WithParallel 设置并行处理
func (b *PipelineBuilder) WithParallel(workers int) *PipelineBuilder {
	b.pipeline.parallel = true
	b.pipeline.workers = workers
	return b
}

// AddValidationStage 添加校验阶段
func (b *PipelineBuilder) AddValidationStage(name string, validator Validator) *PipelineBuilder {
	b.pipeline.AddStage(NewValidationStage(name, validator, b.logger))
	return b
}

// AddFilterStage 添加滤波阶段
func (b *PipelineBuilder) AddFilterStage(name string, filter Filter) *PipelineBuilder {
	b.pipeline.AddStage(NewFilterStage(name, filter, b.logger))
	return b
}

// AddScaleStage 添加量程转换阶段
func (b *PipelineBuilder) AddScaleStage(name string, scaler Scaler) *PipelineBuilder {
	b.pipeline.AddStage(NewScaleStage(name, scaler, b.logger))
	return b
}

// AddChangeDetectionStage 添加变位检测阶段
func (b *PipelineBuilder) AddChangeDetectionStage(name string, detector ChangeDetector) *PipelineBuilder {
	b.pipeline.AddStage(NewChangeDetectionStage(name, detector, b.logger))
	return b
}

// AddQualityMarkStage 添加质量标记阶段
func (b *PipelineBuilder) AddQualityMarkStage(name string, marker QualityMarker) *PipelineBuilder {
	b.pipeline.AddStage(NewQualityMarkStage(name, marker, b.logger))
	return b
}

// AddCustomStage 添加自定义阶段
func (b *PipelineBuilder) AddCustomStage(name string, process func(ctx context.Context, data *ProcessedData) (*ProcessedData, error)) *PipelineBuilder {
	b.pipeline.AddStage(NewCustomStage(name, process, b.logger))
	return b
}

// Build 构建管道
func (b *PipelineBuilder) Build() *Pipeline {
	return b.pipeline
}

// DataProcessor 数据处理器
type DataProcessor struct {
	pipelines map[string]*Pipeline
	mu        sync.RWMutex
	logger    *zap.Logger
}

// NewDataProcessor 创建数据处理器
func NewDataProcessor(logger *zap.Logger) *DataProcessor {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &DataProcessor{
		pipelines: make(map[string]*Pipeline),
		logger:    logger,
	}
}

// AddPipeline 添加管道
func (p *DataProcessor) AddPipeline(name string, pipeline *Pipeline) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pipelines[name] = pipeline
}

// GetPipeline 获取管道
func (p *DataProcessor) GetPipeline(name string) (*Pipeline, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	pipeline, ok := p.pipelines[name]
	return pipeline, ok
}

// RemovePipeline 移除管道
func (p *DataProcessor) RemovePipeline(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.pipelines, name)
}

// Process 使用指定管道处理数据
func (p *DataProcessor) Process(ctx context.Context, pipelineName string, data *ProcessedData) (*ProcessedData, error) {
	p.mu.RLock()
	pipeline, ok := p.pipelines[pipelineName]
	p.mu.RUnlock()
	
	if !ok {
		return nil, fmt.Errorf("pipeline not found: %s", pipelineName)
	}
	
	return pipeline.Process(ctx, data)
}

// ProcessBatch 批量处理
func (p *DataProcessor) ProcessBatch(ctx context.Context, pipelineName string, dataList []*ProcessedData) ([]*ProcessedData, error) {
	p.mu.RLock()
	pipeline, ok := p.pipelines[pipelineName]
	p.mu.RUnlock()
	
	if !ok {
		return nil, fmt.Errorf("pipeline not found: %s", pipelineName)
	}
	
	return pipeline.ProcessBatch(ctx, dataList)
}

// GetAllStatistics 获取所有管道统计信息
func (p *DataProcessor) GetAllStatistics() map[string]PipelineStatistics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	stats := make(map[string]PipelineStatistics)
	for name, pipeline := range p.pipelines {
		stats[name] = pipeline.GetStatistics()
	}
	
	return stats
}

// ProcessorPlugin 处理器插件接口
type ProcessorPlugin interface {
	// Name 插件名称
	Name() string
	// Version 插件版本
	Version() string
	// CreateStage 创建处理阶段
	CreateStage(config map[string]interface{}) (Stage, error)
}

// PluginManager 插件管理器
type PluginManager struct {
	plugins map[string]ProcessorPlugin
	mu      sync.RWMutex
	logger  *zap.Logger
}

// NewPluginManager 创建插件管理器
func NewPluginManager(logger *zap.Logger) *PluginManager {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &PluginManager{
		plugins: make(map[string]ProcessorPlugin),
		logger:  logger,
	}
}

// RegisterPlugin 注册插件
func (m *PluginManager) RegisterPlugin(plugin ProcessorPlugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	name := plugin.Name()
	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin already registered: %s", name)
	}
	
	m.plugins[name] = plugin
	
	m.logger.Info("plugin registered",
		zap.String("name", name),
		zap.String("version", plugin.Version()),
	)
	
	return nil
}

// UnregisterPlugin 注销插件
func (m *PluginManager) UnregisterPlugin(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.plugins, name)
}

// GetPlugin 获取插件
func (m *PluginManager) GetPlugin(name string) (ProcessorPlugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	plugin, ok := m.plugins[name]
	return plugin, ok
}

// CreateStageFromPlugin 从插件创建阶段
func (m *PluginManager) CreateStageFromPlugin(pluginName string, config map[string]interface{}) (Stage, error) {
	plugin, ok := m.GetPlugin(pluginName)
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", pluginName)
	}
	
	return plugin.CreateStage(config)
}

// ProcessingContext 处理上下文
type ProcessingContext struct {
	context.Context
	PointID   string
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// NewProcessingContext 创建处理上下文
func NewProcessingContext(ctx context.Context, pointID string) *ProcessingContext {
	return &ProcessingContext{
		Context:   ctx,
		PointID:   pointID,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// SetMetadata 设置元数据
func (c *ProcessingContext) SetMetadata(key string, value interface{}) {
	c.Metadata[key] = value
}

// GetMetadata 获取元数据
func (c *ProcessingContext) GetMetadata(key string) (interface{}, bool) {
	val, ok := c.Metadata[key]
	return val, ok
}
