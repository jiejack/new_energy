package bigdata

import (
	"fmt"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/analysis"
	"github.com/new-energy-monitoring/pkg/bigdata/ingestion"
	"github.com/new-energy-monitoring/pkg/bigdata/processing"
	"github.com/new-energy-monitoring/pkg/bigdata/storage"
	"github.com/new-energy-monitoring/pkg/bigdata/types"
	"github.com/new-energy-monitoring/pkg/bigdata/visualization"
)

// BigDataServiceImpl 实现了types.BigDataService接口
type BigDataServiceImpl struct {
	storage       types.Storage
	analysis      types.Analysis
	visualization types.Visualization
	processing    types.Processing
	ingestion     types.Ingestion
}

// NewBigDataService 创建一个新的大数据服务实例
func NewBigDataService() *BigDataServiceImpl {
	return &BigDataServiceImpl{}
}

// Init 初始化大数据服务
func (s *BigDataServiceImpl) Init(
	storageConfig types.StorageConfig,
	analysisConfig types.AnalysisConfig,
	visualizationConfig types.VisualizationConfig,
	processingConfig types.ProcessingConfig,
	ingestionConfig types.IngestionConfig,
) error {
	// 初始化存储
	switch storageConfig.Type {
	case "clickhouse":
		s.storage = storage.NewClickHouseStorage()
	case "doris":
		s.storage = storage.NewDorisStorage()
	default:
		return &types.Error{
			Code:    types.ErrCodeInvalidConfig,
			Message: fmt.Sprintf("unsupported storage type: %s", storageConfig.Type),
		}
	}

	if err := s.storage.Init(storageConfig); err != nil {
		return err
	}

	// 初始化分析
	if analysisConfig.Type == "basic" {
		s.analysis = analysis.NewBasicAnalyzer()
	} else {
		return &types.Error{
			Code:    types.ErrCodeInvalidConfig,
			Message: fmt.Sprintf("unsupported analysis type: %s", analysisConfig.Type),
		}
	}

	if err := s.analysis.Init(analysisConfig); err != nil {
		return err
	}

	// 初始化可视化
	if visualizationConfig.Type == "basic" {
		s.visualization = visualization.NewBasicVisualizer()
	} else {
		return &types.Error{
			Code:    types.ErrCodeInvalidConfig,
			Message: fmt.Sprintf("unsupported visualization type: %s", visualizationConfig.Type),
		}
	}

	if err := s.visualization.Init(visualizationConfig); err != nil {
		return err
	}

	// 初始化处理
	switch processingConfig.Type {
	case "basic":
		s.processing = processing.NewBasicProcessor()
	case "flink":
		s.processing = processing.NewFlinkProcessor()
	default:
		return &types.Error{
			Code:    types.ErrCodeInvalidConfig,
			Message: fmt.Sprintf("unsupported processing type: %s", processingConfig.Type),
		}
	}

	if err := s.processing.Init(processingConfig); err != nil {
		return err
	}

	// 初始化摄取
	if ingestionConfig.Type == "basic" {
		s.ingestion = ingestion.NewBasicIngester()
	} else {
		return &types.Error{
			Code:    types.ErrCodeInvalidConfig,
			Message: fmt.Sprintf("unsupported ingestion type: %s", ingestionConfig.Type),
		}
	}

	if err := s.ingestion.Init(ingestionConfig); err != nil {
		return err
	}

	// 注册数据处理函数
	if ingester, ok := s.ingestion.(*ingestion.BasicIngester); ok {
		ingester.RegisterHandler(func(data *types.BatchData) {
			// 处理数据
			processedData, err := s.processing.Process(data)
			if err == nil {
				// 存储数据
				_ = s.storage.Write(processedData)
			}
		})
	}

	return nil
}

// Ingest 摄取数据
func (s *BigDataServiceImpl) Ingest(data *types.BatchData) error {
	if s.ingestion == nil {
		return &types.Error{
			Code:    types.ErrCodeIngestionError,
			Message: "ingestion not initialized",
		}
	}

	// 直接摄取数据
	if ingester, ok := s.ingestion.(*ingestion.BasicIngester); ok {
		return ingester.Ingest(data)
	}

	return &types.Error{
		Code:    types.ErrCodeIngestionError,
		Message: "unsupported ingestion implementation",
	}
}

// Store 存储数据
func (s *BigDataServiceImpl) Store(data *types.BatchData) error {
	if s.storage == nil {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	return s.storage.Write(data)
}

// Analyze 分析数据
func (s *BigDataServiceImpl) Analyze(query string) (interface{}, error) {
	if s.analysis == nil {
		return nil, &types.Error{
			Code:    types.ErrCodeAnalysisError,
			Message: "analysis not initialized",
		}
	}

	return s.analysis.Execute(query)
}

// Visualize 可视化数据
func (s *BigDataServiceImpl) Visualize(dashboardID, panelID string, data interface{}) error {
	if s.visualization == nil {
		return &types.Error{
			Code:    types.ErrCodeVisualizationError,
			Message: "visualization not initialized",
		}
	}

	return s.visualization.UpdatePanel(dashboardID, panelID, data)
}

// Process 处理数据
func (s *BigDataServiceImpl) Process(data *types.BatchData) (*types.BatchData, error) {
	if s.processing == nil {
		return nil, &types.Error{
			Code:    types.ErrCodeProcessingError,
			Message: "processing not initialized",
		}
	}

	return s.processing.Process(data)
}

// StartIngestion 启动数据摄取
func (s *BigDataServiceImpl) StartIngestion() error {
	if s.ingestion == nil {
		return &types.Error{
			Code:    types.ErrCodeIngestionError,
			Message: "ingestion not initialized",
		}
	}

	return s.ingestion.Start()
}

// StopIngestion 停止数据摄取
func (s *BigDataServiceImpl) StopIngestion() error {
	if s.ingestion == nil {
		return &types.Error{
			Code:    types.ErrCodeIngestionError,
			Message: "ingestion not initialized",
		}
	}

	return s.ingestion.Stop()
}

// WritePoint 写入单个数据点
func (s *BigDataServiceImpl) WritePoint(point *types.DataPoint) error {
	if s.storage == nil {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	return s.storage.WritePoint(point)
}

// ReadTimeRange 按时间范围读取数据
func (s *BigDataServiceImpl) ReadTimeRange(
	startTime, endTime time.Time,
	stationID, deviceID, metricName string,
) ([]*types.DataPoint, error) {
	if s.storage == nil {
		return nil, &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	return s.storage.ReadTimeRange(startTime, endTime, stationID, deviceID, metricName)
}

// Aggregate 执行聚合查询
func (s *BigDataServiceImpl) Aggregate(
	aggregation string,
	metricName string,
	startTime, endTime time.Time,
	groupBy string,
) (interface{}, error) {
	if s.storage == nil {
		return nil, &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	return s.storage.Aggregate(aggregation, metricName, startTime, endTime, groupBy)
}

// Flush 手动刷新缓冲区
func (s *BigDataServiceImpl) Flush() error {
	if s.storage == nil {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	return s.storage.Flush()
}

// GetStorageStats 获取存储统计信息
func (s *BigDataServiceImpl) GetStorageStats() (map[string]interface{}, error) {
	if s.storage == nil {
		return nil, &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "storage not initialized",
		}
	}

	return s.storage.GetStats()
}

// GetProcessingStats 获取处理统计信息
func (s *BigDataServiceImpl) GetProcessingStats() (map[string]interface{}, error) {
	if flinkProc, ok := s.processing.(interface{ GetStats() map[string]interface{} }); ok {
		return flinkProc.GetStats(), nil
	}

	return map[string]interface{}{
		"processor": "unknown",
	}, nil
}

// Close 关闭服务
func (s *BigDataServiceImpl) Close() error {
	// 停止摄取
	if s.ingestion != nil {
		_ = s.ingestion.Close()
	}

	// 关闭存储
	if s.storage != nil {
		_ = s.storage.Close()
	}

	// 关闭分析
	if s.analysis != nil {
		_ = s.analysis.Close()
	}

	// 关闭可视化
	if s.visualization != nil {
		_ = s.visualization.Close()
	}

	// 关闭处理
	if s.processing != nil {
		_ = s.processing.Close()
	}

	return nil
}
