package processor

import (
	"container/heap"
	"math"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FilterResult 滤波结果
type FilterResult struct {
	Value     float64   `json:"value"`     // 滤波后的值
	RawValue  float64   `json:"raw_value"` // 原始值
	Quality   QualityCode `json:"quality"` // 质量码
	Timestamp time.Time `json:"timestamp"` // 时间戳
	Filtered  bool      `json:"filtered"`  // 是否进行了滤波
}

// Filter 滤波器接口
type Filter interface {
	// Filter 执行滤波
	Filter(value float64) FilterResult
	// Reset 重置滤波器状态
	Reset()
	// Name 获取滤波器名称
	Name() string
}

// MovingAverageFilter 移动平均滤波器
type MovingAverageFilter struct {
	name      string
	window    int       // 窗口大小
	values    []float64 // 历史值
	sum       float64   // 值总和
	index     int       // 当前索引
	count     int       // 当前计数
	mu        sync.RWMutex
	logger    *zap.Logger
}

// MovingAverageFilterConfig 移动平均滤波器配置
type MovingAverageFilterConfig struct {
	Name   string
	Window int
	Logger *zap.Logger
}

// NewMovingAverageFilter 创建移动平均滤波器
func NewMovingAverageFilter(config MovingAverageFilterConfig) *MovingAverageFilter {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	if config.Window <= 0 {
		config.Window = 5
	}
	return &MovingAverageFilter{
		name:   config.Name,
		window: config.Window,
		values: make([]float64, config.Window),
		logger: config.Logger,
	}
}

// Filter 执行滤波
func (f *MovingAverageFilter) Filter(value float64) FilterResult {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// 减去即将被替换的值
	if f.count >= f.window {
		f.sum -= f.values[f.index]
	}
	
	// 添加新值
	f.values[f.index] = value
	f.sum += value
	f.index = (f.index + 1) % f.window
	
	if f.count < f.window {
		f.count++
	}
	
	// 计算平均值
	avg := f.sum / float64(f.count)
	
	result := FilterResult{
		Value:     avg,
		RawValue:  value,
		Quality:   QualityGood | QualityReasonFiltered,
		Timestamp: time.Now(),
		Filtered:  f.count >= f.window,
	}
	
	f.logger.Debug("moving average filter",
		zap.Float64("raw", value),
		zap.Float64("filtered", avg),
		zap.Int("count", f.count),
	)
	
	return result
}

// Reset 重置滤波器状态
func (f *MovingAverageFilter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	f.values = make([]float64, f.window)
	f.sum = 0
	f.index = 0
	f.count = 0
}

// Name 获取滤波器名称
func (f *MovingAverageFilter) Name() string {
	return f.name
}

// MedianFilter 中值滤波器
type MedianFilter struct {
	name   string
	window int            // 窗口大小
	values []float64      // 历史值
	index  int            // 当前索引
	count  int            // 当前计数
	mu     sync.RWMutex
	logger *zap.Logger
}

// MedianFilterConfig 中值滤波器配置
type MedianFilterConfig struct {
	Name   string
	Window int
	Logger *zap.Logger
}

// NewMedianFilter 创建中值滤波器
func NewMedianFilter(config MedianFilterConfig) *MedianFilter {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	if config.Window <= 0 {
		config.Window = 5
	}
	// 窗口大小必须是奇数
	if config.Window%2 == 0 {
		config.Window++
	}
	return &MedianFilter{
		name:   config.Name,
		window: config.Window,
		values: make([]float64, config.Window),
		logger: config.Logger,
	}
}

// Filter 执行滤波
func (f *MedianFilter) Filter(value float64) FilterResult {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// 添加新值
	f.values[f.index] = value
	f.index = (f.index + 1) % f.window
	
	if f.count < f.window {
		f.count++
	}
	
	// 计算中值
	median := f.calculateMedian()
	
	result := FilterResult{
		Value:     median,
		RawValue:  value,
		Quality:   QualityGood | QualityReasonFiltered,
		Timestamp: time.Now(),
		Filtered:  f.count >= f.window,
	}
	
	f.logger.Debug("median filter",
		zap.Float64("raw", value),
		zap.Float64("filtered", median),
		zap.Int("count", f.count),
	)
	
	return result
}

// calculateMedian 计算中值
func (f *MedianFilter) calculateMedian() float64 {
	// 复制值并排序
	sorted := make([]float64, f.count)
	copy(sorted, f.values[:f.count])
	sort.Float64s(sorted)
	
	// 返回中值
	mid := f.count / 2
	if f.count%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

// Reset 重置滤波器状态
func (f *MedianFilter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	f.values = make([]float64, f.window)
	f.index = 0
	f.count = 0
}

// Name 获取滤波器名称
func (f *MedianFilter) Name() string {
	return f.name
}

// KalmanFilter 卡尔曼滤波器
type KalmanFilter struct {
	name         string
	processNoise float64 // 过程噪声
	measureNoise float64 // 测量噪声
	estimate     float64 // 估计值
	errorCov     float64 // 误差协方差
	initialized  bool
	mu           sync.RWMutex
	logger       *zap.Logger
}

// KalmanFilterConfig 卡尔曼滤波器配置
type KalmanFilterConfig struct {
	Name         string
	ProcessNoise float64
	MeasureNoise float64
	InitialValue float64
	Logger       *zap.Logger
}

// NewKalmanFilter 创建卡尔曼滤波器
func NewKalmanFilter(config KalmanFilterConfig) *KalmanFilter {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	if config.ProcessNoise <= 0 {
		config.ProcessNoise = 0.01
	}
	if config.MeasureNoise <= 0 {
		config.MeasureNoise = 0.1
	}
	return &KalmanFilter{
		name:         config.Name,
		processNoise: config.ProcessNoise,
		measureNoise: config.MeasureNoise,
		estimate:     config.InitialValue,
		errorCov:     1.0,
		logger:       config.Logger,
	}
}

// Filter 执行滤波
func (f *KalmanFilter) Filter(value float64) FilterResult {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// 初始化
	if !f.initialized {
		f.estimate = value
		f.errorCov = 1.0
		f.initialized = true
		
		result := FilterResult{
			Value:     value,
			RawValue:  value,
			Quality:   QualityGood,
			Timestamp: time.Now(),
			Filtered:  false,
		}
		return result
	}
	
	// 预测步骤
	predictedCov := f.errorCov + f.processNoise
	
	// 更新步骤
	kalmanGain := predictedCov / (predictedCov + f.measureNoise)
	f.estimate = f.estimate + kalmanGain*(value-f.estimate)
	f.errorCov = (1 - kalmanGain) * predictedCov
	
	result := FilterResult{
		Value:     f.estimate,
		RawValue:  value,
		Quality:   QualityGood | QualityReasonFiltered,
		Timestamp: time.Now(),
		Filtered:  true,
	}
	
	f.logger.Debug("kalman filter",
		zap.Float64("raw", value),
		zap.Float64("filtered", f.estimate),
		zap.Float64("kalman_gain", kalmanGain),
	)
	
	return result
}

// Reset 重置滤波器状态
func (f *KalmanFilter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	f.estimate = 0
	f.errorCov = 1.0
	f.initialized = false
}

// Name 获取滤波器名称
func (f *KalmanFilter) Name() string {
	return f.name
}

// LimitFilter 限幅滤波器
type LimitFilter struct {
	name      string
	maxChange float64 // 最大变化量
	lastValue float64 // 上一次的值
	initialized bool
	mu        sync.RWMutex
	logger    *zap.Logger
}

// LimitFilterConfig 限幅滤波器配置
type LimitFilterConfig struct {
	Name      string
	MaxChange float64
	Logger    *zap.Logger
}

// NewLimitFilter 创建限幅滤波器
func NewLimitFilter(config LimitFilterConfig) *LimitFilter {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	if config.MaxChange <= 0 {
		config.MaxChange = 10.0
	}
	return &LimitFilter{
		name:      config.Name,
		maxChange: config.MaxChange,
		logger:    config.Logger,
	}
}

// Filter 执行滤波
func (f *LimitFilter) Filter(value float64) FilterResult {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// 初始化
	if !f.initialized {
		f.lastValue = value
		f.initialized = true
		
		result := FilterResult{
			Value:     value,
			RawValue:  value,
			Quality:   QualityGood,
			Timestamp: time.Now(),
			Filtered:  false,
		}
		return result
	}
	
	// 计算变化量
	change := value - f.lastValue
	
	var filteredValue float64
	filtered := false
	
	// 限幅
	if math.Abs(change) > f.maxChange {
		if change > 0 {
			filteredValue = f.lastValue + f.maxChange
		} else {
			filteredValue = f.lastValue - f.maxChange
		}
		filtered = true
	} else {
		filteredValue = value
	}
	
	f.lastValue = filteredValue
	
	result := FilterResult{
		Value:     filteredValue,
		RawValue:  value,
		Quality:   QualityGood | QualityReasonFiltered,
		Timestamp: time.Now(),
		Filtered:  filtered,
	}
	
	f.logger.Debug("limit filter",
		zap.Float64("raw", value),
		zap.Float64("filtered", filteredValue),
		zap.Float64("change", change),
		zap.Bool("limited", filtered),
	)
	
	return result
}

// Reset 重置滤波器状态
func (f *LimitFilter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	f.lastValue = 0
	f.initialized = false
}

// Name 获取滤波器名称
func (f *LimitFilter) Name() string {
	return f.name
}

// ExponentialSmoothingFilter 指数平滑滤波器
type ExponentialSmoothingFilter struct {
	name         string
	alpha        float64 // 平滑系数 (0-1)
	estimate     float64 // 估计值
	initialized  bool
	mu           sync.RWMutex
	logger       *zap.Logger
}

// ExponentialSmoothingFilterConfig 指数平滑滤波器配置
type ExponentialSmoothingFilterConfig struct {
	Name         string
	Alpha        float64
	InitialValue float64
	Logger       *zap.Logger
}

// NewExponentialSmoothingFilter 创建指数平滑滤波器
func NewExponentialSmoothingFilter(config ExponentialSmoothingFilterConfig) *ExponentialSmoothingFilter {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	if config.Alpha <= 0 || config.Alpha > 1 {
		config.Alpha = 0.3
	}
	return &ExponentialSmoothingFilter{
		name:     config.Name,
		alpha:    config.Alpha,
		estimate: config.InitialValue,
		logger:   config.Logger,
	}
}

// Filter 执行滤波
func (f *ExponentialSmoothingFilter) Filter(value float64) FilterResult {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	if !f.initialized {
		f.estimate = value
		f.initialized = true
		
		result := FilterResult{
			Value:     value,
			RawValue:  value,
			Quality:   QualityGood,
			Timestamp: time.Now(),
			Filtered:  false,
		}
		return result
	}
	
	// 指数平滑: estimate = alpha * value + (1 - alpha) * estimate
	f.estimate = f.alpha*value + (1-f.alpha)*f.estimate
	
	result := FilterResult{
		Value:     f.estimate,
		RawValue:  value,
		Quality:   QualityGood | QualityReasonFiltered,
		Timestamp: time.Now(),
		Filtered:  true,
	}
	
	f.logger.Debug("exponential smoothing filter",
		zap.Float64("raw", value),
		zap.Float64("filtered", f.estimate),
		zap.Float64("alpha", f.alpha),
	)
	
	return result
}

// Reset 重置滤波器状态
func (f *ExponentialSmoothingFilter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	f.estimate = 0
	f.initialized = false
}

// Name 获取滤波器名称
func (f *ExponentialSmoothingFilter) Name() string {
	return f.name
}

// FilterChain 滤波器链
type FilterChain struct {
	name    string
	filters []Filter
	mu      sync.RWMutex
	logger  *zap.Logger
}

// FilterChainConfig 滤波器链配置
type FilterChainConfig struct {
	Name   string
	Logger *zap.Logger
}

// NewFilterChain 创建滤波器链
func NewFilterChain(config FilterChainConfig) *FilterChain {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	return &FilterChain{
		name:    config.Name,
		filters: make([]Filter, 0),
		logger:  config.Logger,
	}
}

// AddFilter 添加滤波器
func (c *FilterChain) AddFilter(filter Filter) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.filters = append(c.filters, filter)
}

// RemoveFilter 移除滤波器
func (c *FilterChain) RemoveFilter(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for i, f := range c.filters {
		if f.Name() == name {
			c.filters = append(c.filters[:i], c.filters[i+1:]...)
			break
		}
	}
}

// Filter 执行滤波链
func (c *FilterChain) Filter(value float64) FilterResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	result := FilterResult{
		Value:     value,
		RawValue:  value,
		Quality:   QualityGood,
		Timestamp: time.Now(),
		Filtered:  false,
	}
	
	var codes []QualityCode
	
	for _, filter := range c.filters {
		result = filter.Filter(result.Value)
		if result.Filtered {
			codes = append(codes, QualityReasonFiltered)
		}
	}
	
	// 组合质量码
	if len(codes) > 0 {
		marker := NewQualityMarker(c.logger)
		result.Quality = marker.Combine(codes...)
	}
	
	c.logger.Debug("filter chain completed",
		zap.Int("filter_count", len(c.filters)),
		zap.Float64("raw", value),
		zap.Float64("filtered", result.Value),
	)
	
	return result
}

// Reset 重置所有滤波器
func (c *FilterChain) Reset() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	for _, filter := range c.filters {
		filter.Reset()
	}
}

// Name 获取滤波器名称
func (c *FilterChain) Name() string {
	return c.name
}

// GetFilters 获取所有滤波器
func (c *FilterChain) GetFilters() []Filter {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	result := make([]Filter, len(c.filters))
	copy(result, c.filters)
	return result
}

// PriorityFilter 优先级滤波器
// 用于处理多个滤波器的优先级
type PriorityFilter struct {
	name     string
	filters  []Filter
	priority []int // 每个滤波器的优先级
	mu       sync.RWMutex
	logger   *zap.Logger
}

// PriorityFilterConfig 优先级滤波器配置
type PriorityFilterConfig struct {
	Name   string
	Logger *zap.Logger
}

// NewPriorityFilter 创建优先级滤波器
func NewPriorityFilter(config PriorityFilterConfig) *PriorityFilter {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	return &PriorityFilter{
		name:     config.Name,
		filters:  make([]Filter, 0),
		priority: make([]int, 0),
		logger:   config.Logger,
	}
}

// AddFilter 添加滤波器（带优先级）
func (f *PriorityFilter) AddFilter(filter Filter, priority int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	f.filters = append(f.filters, filter)
	f.priority = append(f.priority, priority)
	
	// 按优先级排序
	for i := len(f.filters) - 1; i > 0; i-- {
		if f.priority[i] < f.priority[i-1] {
			f.filters[i], f.filters[i-1] = f.filters[i-1], f.filters[i]
			f.priority[i], f.priority[i-1] = f.priority[i-1], f.priority[i]
		} else {
			break
		}
	}
}

// Filter 执行滤波
func (f *PriorityFilter) Filter(value float64) FilterResult {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	result := FilterResult{
		Value:     value,
		RawValue:  value,
		Quality:   QualityGood,
		Timestamp: time.Now(),
		Filtered:  false,
	}
	
	for _, filter := range f.filters {
		result = filter.Filter(result.Value)
	}
	
	return result
}

// Reset 重置滤波器
func (f *PriorityFilter) Reset() {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	for _, filter := range f.filters {
		filter.Reset()
	}
}

// Name 获取滤波器名称
func (f *PriorityFilter) Name() string {
	return f.name
}

// BatchFilter 批量滤波器
type BatchFilter struct {
	filter  Filter
	workers int
	logger  *zap.Logger
}

// NewBatchFilter 创建批量滤波器
func NewBatchFilter(filter Filter, workers int, logger *zap.Logger) *BatchFilter {
	if logger == nil {
		logger = zap.NewNop()
	}
	if workers <= 0 {
		workers = 4
	}
	return &BatchFilter{
		filter:  filter,
		workers: workers,
		logger:  logger,
	}
}

// FilterBatch 批量滤波
func (f *BatchFilter) FilterBatch(values []float64) []FilterResult {
	results := make([]FilterResult, len(values))
	
	var wg sync.WaitGroup
	chunkSize := (len(values) + f.workers - 1) / f.workers
	
	for i := 0; i < f.workers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(values) {
			end = len(values)
		}
		if start >= len(values) {
			break
		}
		
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				results[j] = f.filter.Filter(values[j])
			}
		}(start, end)
	}
	
	wg.Wait()
	
	f.logger.Debug("batch filter completed",
		zap.Int("count", len(values)),
		zap.Int("workers", f.workers),
	)
	
	return results
}

// FilterFactory 滤波器工厂
type FilterFactory struct {
	logger *zap.Logger
}

// NewFilterFactory 创建滤波器工厂
func NewFilterFactory(logger *zap.Logger) *FilterFactory {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &FilterFactory{logger: logger}
}

// CreateMovingAverageFilter 创建移动平均滤波器
func (f *FilterFactory) CreateMovingAverageFilter(name string, window int) *MovingAverageFilter {
	return NewMovingAverageFilter(MovingAverageFilterConfig{
		Name:   name,
		Window: window,
		Logger: f.logger,
	})
}

// CreateMedianFilter 创建中值滤波器
func (f *FilterFactory) CreateMedianFilter(name string, window int) *MedianFilter {
	return NewMedianFilter(MedianFilterConfig{
		Name:   name,
		Window: window,
		Logger: f.logger,
	})
}

// CreateKalmanFilter 创建卡尔曼滤波器
func (f *FilterFactory) CreateKalmanFilter(name string, processNoise, measureNoise, initialValue float64) *KalmanFilter {
	return NewKalmanFilter(KalmanFilterConfig{
		Name:         name,
		ProcessNoise: processNoise,
		MeasureNoise: measureNoise,
		InitialValue: initialValue,
		Logger:       f.logger,
	})
}

// CreateLimitFilter 创建限幅滤波器
func (f *FilterFactory) CreateLimitFilter(name string, maxChange float64) *LimitFilter {
	return NewLimitFilter(LimitFilterConfig{
		Name:      name,
		MaxChange: maxChange,
		Logger:    f.logger,
	})
}

// CreateExponentialSmoothingFilter 创建指数平滑滤波器
func (f *FilterFactory) CreateExponentialSmoothingFilter(name string, alpha, initialValue float64) *ExponentialSmoothingFilter {
	return NewExponentialSmoothingFilter(ExponentialSmoothingFilterConfig{
		Name:         name,
		Alpha:        alpha,
		InitialValue: initialValue,
		Logger:       f.logger,
	})
}

// CreateFilterChain 创建滤波器链
func (f *FilterFactory) CreateFilterChain(name string) *FilterChain {
	return NewFilterChain(FilterChainConfig{
		Name:   name,
		Logger: f.logger,
	})
}

// 优先队列实现（用于中值滤波的优化版本）
type float64Heap []float64

func (h float64Heap) Len() int           { return len(h) }
func (h float64Heap) Less(i, j int) bool { return h[i] < h[j] }
func (h float64Heap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *float64Heap) Push(x interface{}) {
	*h = append(*h, x.(float64))
}

func (h *float64Heap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// MedianFilterOptimized 优化的中值滤波器（使用堆）
type MedianFilterOptimized struct {
	name      string
	window    int
	maxHeap   *float64Heap // 最大堆（存较小的一半）
	minHeap   *float64Heap // 最小堆（存较大的一半）
	values    []float64
	index     int
	count     int
	mu        sync.RWMutex
	logger    *zap.Logger
}

// NewMedianFilterOptimized 创建优化的中值滤波器
func NewMedianFilterOptimized(name string, window int, logger *zap.Logger) *MedianFilterOptimized {
	if logger == nil {
		logger = zap.NewNop()
	}
	if window <= 0 {
		window = 5
	}
	if window%2 == 0 {
		window++
	}
	
	maxH := &float64Heap{}
	minH := &float64Heap{}
	heap.Init(maxH)
	heap.Init(minH)
	
	return &MedianFilterOptimized{
		name:    name,
		window:  window,
		maxHeap: maxH,
		minHeap: minH,
		values:  make([]float64, window),
		logger:  logger,
	}
}

// Filter 执行滤波
func (f *MedianFilterOptimized) Filter(value float64) FilterResult {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// 添加新值
	oldValue := f.values[f.index]
	f.values[f.index] = value
	f.index = (f.index + 1) % f.window
	
	if f.count < f.window {
		f.count++
		// 插入到堆中
		if f.maxHeap.Len() == 0 || value <= (*f.maxHeap)[0] {
			heap.Push(f.maxHeap, -value) // 用负数模拟最大堆
		} else {
			heap.Push(f.minHeap, value)
		}
	} else {
		// 移除旧值并插入新值
		f.removeValue(oldValue)
		f.insertValue(value)
	}
	
	// 平衡堆
	f.balanceHeaps()
	
	// 获取中值
	median := f.getMedian()
	
	result := FilterResult{
		Value:     median,
		RawValue:  value,
		Quality:   QualityGood | QualityReasonFiltered,
		Timestamp: time.Now(),
		Filtered:  f.count >= f.window,
	}
	
	return result
}

func (f *MedianFilterOptimized) removeValue(value float64) {
	// 简化实现：直接重建堆
	*f.maxHeap = (*f.maxHeap)[:0]
	*f.minHeap = (*f.minHeap)[:0]
	
	for _, v := range f.values {
		if v != value {
			f.insertValue(v)
		}
	}
}

func (f *MedianFilterOptimized) insertValue(value float64) {
	if f.maxHeap.Len() == 0 || value <= -(*f.maxHeap)[0] {
		heap.Push(f.maxHeap, -value)
	} else {
		heap.Push(f.minHeap, value)
	}
}

func (f *MedianFilterOptimized) balanceHeaps() {
	// 保持两个堆的大小差不超过1
	for f.maxHeap.Len() > f.minHeap.Len()+1 {
		val := -heap.Pop(f.maxHeap).(float64)
		heap.Push(f.minHeap, val)
	}
	for f.minHeap.Len() > f.maxHeap.Len() {
		val := heap.Pop(f.minHeap).(float64)
		heap.Push(f.maxHeap, -val)
	}
}

func (f *MedianFilterOptimized) getMedian() float64 {
	if f.maxHeap.Len() > f.minHeap.Len() {
		return -(*f.maxHeap)[0]
	}
	if f.minHeap.Len() > f.maxHeap.Len() {
		return (*f.minHeap)[0]
	}
	return (-(*f.maxHeap)[0] + (*f.minHeap)[0]) / 2
}

// Reset 重置滤波器
func (f *MedianFilterOptimized) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	*f.maxHeap = (*f.maxHeap)[:0]
	*f.minHeap = (*f.minHeap)[:0]
	f.values = make([]float64, f.window)
	f.index = 0
	f.count = 0
}

// Name 获取滤波器名称
func (f *MedianFilterOptimized) Name() string {
	return f.name
}
