package timeseries

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// BatchWriterConfig 批量写入器配置
type BatchWriterConfig struct {
	Database     string        // 数据库名
	Table        string        // 表名
	BatchSize    int           // 批次大小
	FlushTimeout time.Duration // 刷新超时
	MaxRetries   int           // 最大重试次数
	RetryDelay   time.Duration // 重试延迟
}

// DefaultBatchWriterConfig 默认批量写入器配置
func DefaultBatchWriterConfig() *BatchWriterConfig {
	return &BatchWriterConfig{
		Database:     "nem_ts",
		Table:        "data_points",
		BatchSize:    10000,
		FlushTimeout: 5 * time.Second,
		MaxRetries:   3,
		RetryDelay:   100 * time.Millisecond,
	}
}

// timeseriesBatchWriter 时序数据库批量写入器
type timeseriesBatchWriter struct {
	client   TimeSeriesDB
	config   *BatchWriterConfig
	logger   *zap.Logger

	buffer   []*DataPoint
	bufferMu sync.Mutex

	stats    WriterStats
	statsMu  sync.RWMutex

	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup

	flushCh  chan struct{}
	closeCh  chan struct{}
}

// NewBatchWriter 创建批量写入器
func NewBatchWriter(client TimeSeriesDB, config *BatchWriterConfig) BatchWriter {
	if config == nil {
		config = DefaultBatchWriterConfig()
	}

	if config.BatchSize <= 0 {
		config.BatchSize = 10000
	}

	if config.FlushTimeout <= 0 {
		config.FlushTimeout = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	w := &timeseriesBatchWriter{
		client:  client,
		config:  config,
		logger:  zap.L().Named("batch-writer"),
		buffer:  make([]*DataPoint, 0, config.BatchSize),
		flushCh: make(chan struct{}, 1),
		closeCh: make(chan struct{}),
		ctx:     ctx,
		cancel:  cancel,
	}

	// 启动后台刷新协程
	w.wg.Add(1)
	go w.flushLoop()

	return w
}

// Add 添加数据点
func (w *timeseriesBatchWriter) Add(point *DataPoint) error {
	if point == nil {
		return nil
	}

	w.bufferMu.Lock()
	defer w.bufferMu.Unlock()

	w.buffer = append(w.buffer, point)
	atomic.AddInt64(&w.stats.TotalPoints, 1)

	// 缓冲区满时触发刷新
	if len(w.buffer) >= w.config.BatchSize {
		select {
		case w.flushCh <- struct{}{}:
		default:
			// 已有待处理的刷新请求
		}
	}

	return nil
}

// Flush 刷新缓冲区
func (w *timeseriesBatchWriter) Flush() error {
	w.bufferMu.Lock()
	points := w.buffer
	w.buffer = make([]*DataPoint, 0, w.config.BatchSize)
	w.bufferMu.Unlock()

	if len(points) == 0 {
		return nil
	}

	return w.writeWithRetry(points)
}

// Close 关闭写入器
func (w *timeseriesBatchWriter) Close() error {
	// 停止后台协程
	w.cancel()
	close(w.closeCh)
	w.wg.Wait()

	// 最后一次刷新
	return w.Flush()
}

// Stats 获取统计信息
func (w *timeseriesBatchWriter) Stats() *WriterStats {
	w.statsMu.RLock()
	defer w.statsMu.RUnlock()

	stats := &WriterStats{
		TotalPoints:   atomic.LoadInt64(&w.stats.TotalPoints),
		SuccessPoints: atomic.LoadInt64(&w.stats.SuccessPoints),
		FailedPoints:  atomic.LoadInt64(&w.stats.FailedPoints),
		TotalBatches:  atomic.LoadInt64(&w.stats.TotalBatches),
		TotalBytes:    atomic.LoadInt64(&w.stats.TotalBytes),
	}
	return stats
}

// flushLoop 刷新循环
func (w *timeseriesBatchWriter) flushLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.FlushTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.flushCh:
			_ = w.Flush()
		case <-ticker.C:
			_ = w.Flush()
		case <-w.closeCh:
			return
		}
	}
}

// writeWithRetry 带重试的写入
func (w *timeseriesBatchWriter) writeWithRetry(points []*DataPoint) error {
	if len(points) == 0 {
		return nil
	}

	var lastErr error
	for i := 0; i <= w.config.MaxRetries; i++ {
		err := w.client.WriteWithTable(w.ctx, w.config.Database, w.config.Table, points)
		if err == nil {
			atomic.AddInt64(&w.stats.SuccessPoints, int64(len(points)))
			atomic.AddInt64(&w.stats.TotalBatches, 1)
			w.logger.Debug("batch write succeeded",
				zap.Int("points", len(points)),
				zap.Int("attempt", i+1))
			return nil
		}

		lastErr = err

		// 不可重试错误直接返回
		if !IsRetryableError(err) {
			break
		}

		// 等待重试
		if i < w.config.MaxRetries {
			time.Sleep(w.config.RetryDelay * time.Duration(i+1))
		}
	}

	atomic.AddInt64(&w.stats.FailedPoints, int64(len(points)))
	w.logger.Error("batch write failed after retries",
		zap.Int("points", len(points)),
		zap.Int("retries", w.config.MaxRetries),
		zap.Error(lastErr))

	return lastErr
}

// AsyncBatchWriter 异步批量写入器
type AsyncBatchWriter struct {
	writer   BatchWriter
	errCh    chan error
	closeCh  chan struct{}
	wg       sync.WaitGroup
}

// NewAsyncBatchWriter 创建异步批量写入器
func NewAsyncBatchWriter(client TimeSeriesDB, config *BatchWriterConfig) *AsyncBatchWriter {
	writer := NewBatchWriter(client, config)

	return &AsyncBatchWriter{
		writer:  writer,
		errCh:   make(chan error, 100),
		closeCh: make(chan struct{}),
	}
}

// Add 异步添加数据点
func (w *AsyncBatchWriter) Add(point *DataPoint) <-chan error {
	resultCh := make(chan error, 1)

	go func() {
		err := w.writer.Add(point)
		if err != nil {
			select {
			case w.errCh <- err:
			default:
			}
		}
		resultCh <- err
	}()

	return resultCh
}

// Errors 返回错误通道
func (w *AsyncBatchWriter) Errors() <-chan error {
	return w.errCh
}

// Close 关闭写入器
func (w *AsyncBatchWriter) Close() error {
	close(w.closeCh)
	w.wg.Wait()
	return w.writer.Close()
}

// Stats 获取统计信息
func (w *AsyncBatchWriter) Stats() *WriterStats {
	return w.writer.Stats()
}
