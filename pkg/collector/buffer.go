package collector

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/new-energy-monitoring/internal/infrastructure/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	ErrBufferClosed     = errors.New("buffer is closed")
	ErrBufferFull       = errors.New("buffer is full")
	ErrInvalidConfig    = errors.New("invalid configuration")
	ErrWriterNotSet     = errors.New("writer not set")
)

// Prometheus指标
var (
	bufferDataTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "collector_buffer_data_total",
		Help: "Total number of data points processed by buffer",
	}, []string{"status"})

	bufferSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "collector_buffer_size",
		Help: "Current buffer size",
	})

	bufferFlushDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "collector_buffer_flush_duration_seconds",
		Help:    "Buffer flush duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"status"})

	bufferRetryTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "collector_buffer_retry_total",
		Help: "Total number of buffer write retries",
	})
)

// BufferConfig 缓冲区配置
type BufferConfig struct {
	MaxSize          int           // 最大缓冲区大小
	FlushInterval    time.Duration // 刷新间隔
	FlushThreshold   int           // 刷新阈值
	MaxRetryCount    int           // 最大重试次数
	RetryDelay       time.Duration // 重试延迟
	EnableMetrics    bool          // 是否启用指标
	EnableCompression bool         // 是否启用压缩
}

// BufferOption 缓冲区配置选项
type BufferOption func(*DataBuffer)

// WithMaxSize 设置最大缓冲区大小
func WithMaxSize(size int) BufferOption {
	return func(b *DataBuffer) {
		if size > 0 {
			b.config.MaxSize = size
		}
	}
}

// WithFlushInterval 设置刷新间隔
func WithFlushInterval(interval time.Duration) BufferOption {
	return func(b *DataBuffer) {
		if interval > 0 {
			b.config.FlushInterval = interval
		}
	}
}

// WithFlushThreshold 设置刷新阈值
func WithFlushThreshold(threshold int) BufferOption {
	return func(b *DataBuffer) {
		if threshold > 0 {
			b.config.FlushThreshold = threshold
		}
	}
}

// WithMaxRetryCount 设置最大重试次数
func WithMaxRetryCount(count int) BufferOption {
	return func(b *DataBuffer) {
		if count >= 0 {
			b.config.MaxRetryCount = count
		}
	}
}

// WithRetryDelay 设置重试延迟
func WithRetryDelay(delay time.Duration) BufferOption {
	return func(b *DataBuffer) {
		if delay > 0 {
			b.config.RetryDelay = delay
		}
	}
}

// WithEnableMetrics 设置是否启用指标
func WithEnableMetrics(enable bool) BufferOption {
	return func(b *DataBuffer) {
		b.config.EnableMetrics = enable
	}
}

// WithEnableCompression 设置是否启用压缩
func WithEnableCompression(enable bool) BufferOption {
	return func(b *DataBuffer) {
		b.config.EnableCompression = enable
	}
}

// BufferMetrics 缓冲区指标
type BufferMetrics struct {
	TotalData       int64 // 总数据量
	CurrentSize     int64 // 当前大小
	FlushCount      int64 // 刷新次数
	SuccessCount    int64 // 成功次数
	FailureCount    int64 // 失败次数
	RetryCount      int64 // 重试次数
	AverageLatency  int64 // 平均延迟(纳秒)
	LastFlushTime   time.Time
}

// DataBuffer 数据缓冲区
type DataBuffer struct {
	config     BufferConfig
	running    int32
	closed     int32

	// 数据缓冲
	data       []PointData
	dataMutex  sync.RWMutex

	// 数据通道
	dataChan   chan PointData
	flushChan  chan struct{}

	// 写入器
	writer     DataWriter

	// 指标
	metrics      BufferMetrics
	metricsMutex sync.RWMutex

	// 控制
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup

	// 日志
	logger *zap.Logger
}

// NewDataBuffer 创建数据缓冲区
func NewDataBuffer(opts ...BufferOption) *DataBuffer {
	// 默认配置
	b := &DataBuffer{
		config: BufferConfig{
			MaxSize:          1000000, // 100万数据点
			FlushInterval:    5 * time.Second,
			FlushThreshold:   10000,
			MaxRetryCount:    3,
			RetryDelay:       1 * time.Second,
			EnableMetrics:    true,
			EnableCompression: false,
		},
		data:     make([]PointData, 0),
		dataChan: make(chan PointData, 100000),
		flushChan: make(chan struct{}, 1),
		logger:   logger.Named("data-buffer"),
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(b)
	}

	// 创建上下文
	b.ctx, b.cancelFunc = context.WithCancel(context.Background())

	return b
}

// SetWriter 设置写入器
func (b *DataBuffer) SetWriter(writer DataWriter) {
	b.writer = writer
}

// Start 启动缓冲区
func (b *DataBuffer) Start() error {
	if atomic.LoadInt32(&b.running) == 1 {
		return errors.New("buffer is already running")
	}

	if b.writer == nil {
		return ErrWriterNotSet
	}

	atomic.StoreInt32(&b.running, 1)
	atomic.StoreInt32(&b.closed, 0)

	b.logger.Info("Starting data buffer",
		zap.Int("maxSize", b.config.MaxSize),
		zap.Duration("flushInterval", b.config.FlushInterval),
		zap.Int("flushThreshold", b.config.FlushThreshold))

	// 启动数据接收器
	b.wg.Add(1)
	go b.receiveData()

	// 启动定时刷新器
	b.wg.Add(1)
	go b.periodicFlush()

	// 启动指标收集器
	if b.config.EnableMetrics {
		b.wg.Add(1)
		go b.collectMetrics()
	}

	return nil
}

// Stop 停止缓冲区
func (b *DataBuffer) Stop() error {
	if atomic.LoadInt32(&b.running) == 0 {
		return errors.New("buffer is not running")
	}

	b.logger.Info("Stopping data buffer")

	atomic.StoreInt32(&b.running, 0)
	atomic.StoreInt32(&b.closed, 1)

	// 取消上下文
	b.cancelFunc()

	// 等待所有协程完成
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 最后一次刷新
		b.flush()
		b.logger.Info("Data buffer stopped successfully")
		return nil
	case <-time.After(30 * time.Second):
		b.logger.Warn("Data buffer stop timeout")
		return errors.New("stop timeout")
	}
}

// Write 写入数据
func (b *DataBuffer) Write(data PointData) error {
	if atomic.LoadInt32(&b.closed) == 1 {
		return ErrBufferClosed
	}

	select {
	case b.dataChan <- data:
		atomic.AddInt64(&b.metrics.TotalData, 1)
		return nil
	case <-b.ctx.Done():
		return b.ctx.Err()
	default:
		return ErrBufferFull
	}
}

// WriteBatch 批量写入数据
func (b *DataBuffer) WriteBatch(data []PointData) error {
	if atomic.LoadInt32(&b.closed) == 1 {
		return ErrBufferClosed
	}

	for _, d := range data {
		select {
		case b.dataChan <- d:
			atomic.AddInt64(&b.metrics.TotalData, 1)
		case <-b.ctx.Done():
			return b.ctx.Err()
		default:
			return ErrBufferFull
		}
	}

	return nil
}

// receiveData 接收数据
func (b *DataBuffer) receiveData() {
	defer b.wg.Done()

	for {
		select {
		case data := <-b.dataChan:
			b.appendData(data)

		case <-b.ctx.Done():
			return
		}
	}
}

// appendData 添加数据到缓冲区
func (b *DataBuffer) appendData(data PointData) {
	b.dataMutex.Lock()
	defer b.dataMutex.Unlock()

	// 检查缓冲区大小
	if len(b.data) >= b.config.MaxSize {
		// 触发刷新
		b.triggerFlush()
	}

	b.data = append(b.data, data)

	// 检查是否达到刷新阈值
	if len(b.data) >= b.config.FlushThreshold {
		b.triggerFlush()
	}
}

// triggerFlush 触发刷新
func (b *DataBuffer) triggerFlush() {
	select {
	case b.flushChan <- struct{}{}:
	default:
		// 已经有刷新请求在等待
	}
}

// periodicFlush 定期刷新
func (b *DataBuffer) periodicFlush() {
	defer b.wg.Done()

	ticker := time.NewTicker(b.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.flush()

		case <-b.flushChan:
			b.flush()

		case <-b.ctx.Done():
			return
		}
	}
}

// flush 刷新缓冲区
func (b *DataBuffer) flush() {
	b.dataMutex.Lock()
	if len(b.data) == 0 {
		b.dataMutex.Unlock()
		return
	}

	// 取出数据
	data := b.data
	b.data = make([]PointData, 0)
	b.dataMutex.Unlock()

	startTime := time.Now()

	// 执行写入
	err := b.writeWithRetry(data)

	duration := time.Since(startTime)

	// 更新指标
	atomic.AddInt64(&b.metrics.FlushCount, 1)
	b.metricsMutex.Lock()
	b.metrics.LastFlushTime = time.Now()
	b.metricsMutex.Unlock()

	if err != nil {
		atomic.AddInt64(&b.metrics.FailureCount, 1)
		b.logger.Error("Failed to flush buffer",
			zap.Int("dataCount", len(data)),
			zap.Error(err))

		if b.config.EnableMetrics {
			bufferDataTotal.WithLabelValues("failed").Add(float64(len(data)))
			bufferFlushDuration.WithLabelValues("failed").Observe(duration.Seconds())
		}
	} else {
		atomic.AddInt64(&b.metrics.SuccessCount, 1)
		b.logger.Info("Buffer flushed successfully",
			zap.Int("dataCount", len(data)),
			zap.Duration("duration", duration))

		if b.config.EnableMetrics {
			bufferDataTotal.WithLabelValues("success").Add(float64(len(data)))
			bufferFlushDuration.WithLabelValues("success").Observe(duration.Seconds())
		}
	}
}

// writeWithRetry 带重试的写入
func (b *DataBuffer) writeWithRetry(data []PointData) error {
	var lastError error

	for i := 0; i <= b.config.MaxRetryCount; i++ {
		if i > 0 {
			atomic.AddInt64(&b.metrics.RetryCount, 1)
			bufferRetryTotal.Inc()

			b.logger.Info("Retrying write",
				zap.Int("attempt", i),
				zap.Int("maxRetry", b.config.MaxRetryCount))

			time.Sleep(b.config.RetryDelay * time.Duration(i))
		}

		ctx, cancel := context.WithTimeout(b.ctx, 10*time.Second)
		err := b.writer.Write(ctx, data)
		cancel()

		if err == nil {
			return nil
		}

		lastError = err
		b.logger.Warn("Write attempt failed",
			zap.Int("attempt", i),
			zap.Error(err))
	}

	return lastError
}

// collectMetrics 收集指标
func (b *DataBuffer) collectMetrics() {
	defer b.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.dataMutex.RLock()
			size := int64(len(b.data))
			b.dataMutex.RUnlock()

			bufferSize.Set(float64(size))

			b.metricsMutex.Lock()
			b.metrics.CurrentSize = size
			b.metricsMutex.Unlock()

		case <-b.ctx.Done():
			return
		}
	}
}

// GetMetrics 获取指标
func (b *DataBuffer) GetMetrics() BufferMetrics {
	b.metricsMutex.RLock()
	defer b.metricsMutex.RUnlock()

	return BufferMetrics{
		TotalData:      atomic.LoadInt64(&b.metrics.TotalData),
		CurrentSize:    atomic.LoadInt64(&b.metrics.CurrentSize),
		FlushCount:     atomic.LoadInt64(&b.metrics.FlushCount),
		SuccessCount:   atomic.LoadInt64(&b.metrics.SuccessCount),
		FailureCount:   atomic.LoadInt64(&b.metrics.FailureCount),
		RetryCount:     atomic.LoadInt64(&b.metrics.RetryCount),
		AverageLatency: b.metrics.AverageLatency,
		LastFlushTime:  b.metrics.LastFlushTime,
	}
}

// IsRunning 检查是否运行中
func (b *DataBuffer) IsRunning() bool {
	return atomic.LoadInt32(&b.running) == 1
}

// IsClosed 检查是否已关闭
func (b *DataBuffer) IsClosed() bool {
	return atomic.LoadInt32(&b.closed) == 1
}

// GetCurrentSize 获取当前缓冲区大小
func (b *DataBuffer) GetCurrentSize() int {
	b.dataMutex.RLock()
	defer b.dataMutex.RUnlock()
	return len(b.data)
}

// ForceFlush 强制刷新
func (b *DataBuffer) ForceFlush() error {
	if atomic.LoadInt32(&b.closed) == 1 {
		return ErrBufferClosed
	}

	b.triggerFlush()
	return nil
}

// Clear 清空缓冲区
func (b *DataBuffer) Clear() {
	b.dataMutex.Lock()
	defer b.dataMutex.Unlock()
	b.data = make([]PointData, 0)
}

// BatchWriter 批量写入器
type BatchWriter struct {
	writer       DataWriter
	batchSize    int
	parallelism  int
	timeout      time.Duration

	// 控制
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup

	// 日志
	logger *zap.Logger
}

// BatchWriterConfig 批量写入器配置
type BatchWriterConfig struct {
	BatchSize   int           // 批量大小
	Parallelism int           // 并行度
	Timeout     time.Duration // 超时时间
}

// BatchWriterOption 批量写入器配置选项
type BatchWriterOption func(*BatchWriter)

// WithBatchSize 设置批量大小
func WithBatchSize(size int) BatchWriterOption {
	return func(bw *BatchWriter) {
		if size > 0 {
			bw.batchSize = size
		}
	}
}

// WithParallelism 设置并行度
func WithParallelism(p int) BatchWriterOption {
	return func(bw *BatchWriter) {
		if p > 0 {
			bw.parallelism = p
		}
	}
}

// WithTimeout 设置超时时间
func WithBatchTimeout(timeout time.Duration) BatchWriterOption {
	return func(bw *BatchWriter) {
		if timeout > 0 {
			bw.timeout = timeout
		}
	}
}

// NewBatchWriter 创建批量写入器
func NewBatchWriter(writer DataWriter, opts ...BatchWriterOption) *BatchWriter {
	bw := &BatchWriter{
		writer:      writer,
		batchSize:   1000,
		parallelism: 10,
		timeout:     30 * time.Second,
		logger:      logger.Named("batch-writer"),
	}

	for _, opt := range opts {
		opt(bw)
	}

	bw.ctx, bw.cancelFunc = context.WithCancel(context.Background())

	return bw
}

// Write 写入数据
func (bw *BatchWriter) Write(ctx context.Context, data []PointData) error {
	if len(data) == 0 {
		return nil
	}

	// 分批处理
	batches := bw.splitIntoBatches(data)

	// 并行写入
	errorChan := make(chan error, len(batches))
	var wg sync.WaitGroup

	for i, batch := range batches {
		wg.Add(1)
		go func(idx int, b []PointData) {
			defer wg.Done()

			writeCtx, cancel := context.WithTimeout(ctx, bw.timeout)
			defer cancel()

			if err := bw.writer.Write(writeCtx, b); err != nil {
				bw.logger.Error("Batch write failed",
					zap.Int("batchIndex", idx),
					zap.Int("batchSize", len(b)),
					zap.Error(err))
				errorChan <- err
			}
		}(i, batch)
	}

	// 等待所有批次完成
	go func() {
		wg.Wait()
		close(errorChan)
	}()

	// 收集错误
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// WriteBatch 批量写入
func (bw *BatchWriter) WriteBatch(ctx context.Context, batch [][]PointData) error {
	errorChan := make(chan error, len(batch))
	var wg sync.WaitGroup

	for i, b := range batch {
		wg.Add(1)
		go func(idx int, data []PointData) {
			defer wg.Done()

			writeCtx, cancel := context.WithTimeout(ctx, bw.timeout)
			defer cancel()

			if err := bw.writer.Write(writeCtx, data); err != nil {
				bw.logger.Error("Batch write failed",
					zap.Int("batchIndex", idx),
					zap.Int("batchSize", len(data)),
					zap.Error(err))
				errorChan <- err
			}
		}(i, b)
	}

	go func() {
		wg.Wait()
		close(errorChan)
	}()

	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// splitIntoBatches 将数据分割成批次
func (bw *BatchWriter) splitIntoBatches(data []PointData) [][]PointData {
	var batches [][]PointData

	for i := 0; i < len(data); i += bw.batchSize {
		end := i + bw.batchSize
		if end > len(data) {
			end = len(data)
		}

		batches = append(batches, data[i:end])
	}

	return batches
}

// Close 关闭写入器
func (bw *BatchWriter) Close() error {
	bw.cancelFunc()
	bw.wg.Wait()

	if bw.writer != nil {
		return bw.writer.Close()
	}

	return nil
}

// RetryWriter 带重试机制的写入器
type RetryWriter struct {
	writer       DataWriter
	maxRetry     int
	retryDelay   time.Duration
	exponentialBackoff bool

	logger *zap.Logger
}

// RetryWriterOption 重试写入器配置选项
type RetryWriterOption func(*RetryWriter)

// WithMaxRetry 设置最大重试次数
func WithWriterMaxRetry(count int) RetryWriterOption {
	return func(rw *RetryWriter) {
		if count >= 0 {
			rw.maxRetry = count
		}
	}
}

// WithWriterRetryDelay 设置重试延迟
func WithWriterRetryDelay(delay time.Duration) RetryWriterOption {
	return func(rw *RetryWriter) {
		if delay > 0 {
			rw.retryDelay = delay
		}
	}
}

// WithExponentialBackoff 设置是否启用指数退避
func WithExponentialBackoff(enable bool) RetryWriterOption {
	return func(rw *RetryWriter) {
		rw.exponentialBackoff = enable
	}
}

// NewRetryWriter 创建带重试机制的写入器
func NewRetryWriter(writer DataWriter, opts ...RetryWriterOption) *RetryWriter {
	rw := &RetryWriter{
		writer:             writer,
		maxRetry:           3,
		retryDelay:         1 * time.Second,
		exponentialBackoff: true,
		logger:             logger.Named("retry-writer"),
	}

	for _, opt := range opts {
		opt(rw)
	}

	return rw
}

// Write 写入数据
func (rw *RetryWriter) Write(ctx context.Context, data []PointData) error {
	var lastError error

	for i := 0; i <= rw.maxRetry; i++ {
		if i > 0 {
			delay := rw.retryDelay
			if rw.exponentialBackoff {
				delay = rw.retryDelay * time.Duration(1<<uint(i-1))
			}

			rw.logger.Info("Retrying write",
				zap.Int("attempt", i),
				zap.Duration("delay", delay))

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		if err := rw.writer.Write(ctx, data); err != nil {
			lastError = err
			rw.logger.Warn("Write attempt failed",
				zap.Int("attempt", i),
				zap.Error(err))
			continue
		}

		return nil
	}

	return lastError
}

// WriteBatch 批量写入数据
func (rw *RetryWriter) WriteBatch(ctx context.Context, batch [][]PointData) error {
	var lastError error

	for i := 0; i <= rw.maxRetry; i++ {
		if i > 0 {
			delay := rw.retryDelay
			if rw.exponentialBackoff {
				delay = rw.retryDelay * time.Duration(1<<uint(i-1))
			}

			rw.logger.Info("Retrying batch write",
				zap.Int("attempt", i),
				zap.Duration("delay", delay))

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		if err := rw.writer.WriteBatch(ctx, batch); err != nil {
			lastError = err
			rw.logger.Warn("Batch write attempt failed",
				zap.Int("attempt", i),
				zap.Error(err))
			continue
		}

		return nil
	}

	return lastError
}

// Close 关闭写入器
func (rw *RetryWriter) Close() error {
	if rw.writer != nil {
		return rw.writer.Close()
	}
	return nil
}
