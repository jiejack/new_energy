package ingestion

import (
	"fmt"
	"sync"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata"
)

// BasicIngester 实现了Ingestion接口，提供基本的数据摄取功能
type BasicIngester struct {
	config     bigdata.IngestionConfig
	running    bool
	dataChan   chan *bigdata.BatchData
	stopChan   chan struct{}
	wg         sync.WaitGroup
	handlers   []DataHandler
}

// DataHandler 数据处理函数类型
type DataHandler func(data *bigdata.BatchData)

// NewBasicIngester 创建一个新的基本摄取器实例
func NewBasicIngester() *BasicIngester {
	return &BasicIngester{
		dataChan: make(chan *bigdata.BatchData, 100),
		stopChan: make(chan struct{}),
		handlers: make([]DataHandler, 0),
	}
}

// Init 初始化摄取器
func (i *BasicIngester) Init(config bigdata.IngestionConfig) error {
	if config.Type != "basic" {
		return &bigdata.Error{
			Code:    bigdata.ErrCodeInvalidConfig,
			Message: fmt.Sprintf("invalid ingestion type: %s, expected basic", config.Type),
		}
	}

	i.config = config
	return nil
}

// Start 启动摄取器
func (i *BasicIngester) Start() error {
	if i.running {
		return &bigdata.Error{
			Code:    bigdata.ErrCodeIngestionError,
			Message: "ingester already running",
		}
	}

	i.running = true
	i.wg.Add(1)

	// 启动处理协程
	go i.processData()

	return nil
}

// Stop 停止摄取器
func (i *BasicIngester) Stop() error {
	if !i.running {
		return &bigdata.Error{
			Code:    bigdata.ErrCodeIngestionError,
			Message: "ingester not running",
		}
	}

	i.running = false
	close(i.stopChan)
	i.wg.Wait()

	return nil
}

// Close 关闭摄取器
func (i *BasicIngester) Close() error {
	if i.running {
		if err := i.Stop(); err != nil {
			return err
		}
	}

	close(i.dataChan)
	return nil
}

// RegisterHandler 注册数据处理函数
func (i *BasicIngester) RegisterHandler(handler DataHandler) {
	i.handlers = append(i.handlers, handler)
}

// Ingest 摄取数据
func (i *BasicIngester) Ingest(data *bigdata.BatchData) error {
	if !i.running {
		return &bigdata.Error{
			Code:    bigdata.ErrCodeIngestionError,
			Message: "ingester not running",
		}
	}

	select {
	case i.dataChan <- data:
		return nil
	case <-i.stopChan:
		return &bigdata.Error{
			Code:    bigdata.ErrCodeIngestionError,
			Message: "ingester stopped",
		}
	default:
		return &bigdata.Error{
			Code:    bigdata.ErrCodeIngestionError,
			Message: "data channel full",
		}
	}
}

// processData 处理数据
func (i *BasicIngester) processData() {
	defer i.wg.Done()

	for {
		select {
		case data, ok := <-i.dataChan:
			if !ok {
				return
			}
			
			// 处理数据
			i.handleData(data)

		case <-i.stopChan:
			// 处理剩余数据
			for {
				select {
				case data, ok := <-i.dataChan:
					if !ok {
						return
					}
					i.handleData(data)
				default:
					return
				}
			}
		}
	}
}

// handleData 处理单个数据批次
func (i *BasicIngester) handleData(data *bigdata.BatchData) {
	// 记录处理开始时间
	startTime := time.Now()

	// 调用所有注册的处理函数
	for _, handler := range i.handlers {
		handler(data)
	}

	// 记录处理时间
	processingTime := time.Since(startTime)
	
	// 可以在这里添加日志记录
	// fmt.Printf("Processed batch %s with %d data points in %v\n", data.Metadata.BatchID, data.Metadata.RecordCount, processingTime)
}

// GetStats 获取摄取器统计信息
func (i *BasicIngester) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"running":     i.running,
		"queue_size":  len(i.dataChan),
		"queue_capacity": cap(i.dataChan),
		"handlers_count": len(i.handlers),
		"config":      i.config,
	}
}
