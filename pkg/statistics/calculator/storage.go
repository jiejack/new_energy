package calculator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PeriodType 统计周期类型
type PeriodType string

const (
	PeriodTypeMinute PeriodType = "minute"
	PeriodTypeHour   PeriodType = "hour"
	PeriodTypeDay    PeriodType = "day"
	PeriodTypeMonth  PeriodType = "month"
	PeriodTypeYear   PeriodType = "year"
	PeriodTypeCustom PeriodType = "custom"
)

// StatisticsData 统计数据实体
type StatisticsData struct {
	ID            string      `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TaskID        string      `json:"task_id" gorm:"type:varchar(36);index"`
	Dimension     string      `json:"dimension" gorm:"type:varchar(100);not null;index"`
	DimensionValue string     `json:"dimension_value" gorm:"type:varchar(200);not null;index"`
	MetricName    string      `json:"metric_name" gorm:"type:varchar(100);not null"`
	MetricValue   float64     `json:"metric_value"`
	Metadata      string      `json:"metadata" gorm:"type:text"`
	PeriodType    PeriodType  `json:"period_type" gorm:"type:varchar(20);not null;index"`
	PeriodStart   time.Time   `json:"period_start" gorm:"not null;index"`
	PeriodEnd     time.Time   `json:"period_end" gorm:"not null"`
	CreatedAt     time.Time   `json:"created_at"`
}

func (s *StatisticsData) TableName() string {
	return "statistics_data"
}

// StatisticsTask 统计任务
type StatisticsTask struct {
	ID            string                 `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name          string                 `json:"name" gorm:"type:varchar(200);not null"`
	TaskType      string                 `json:"task_type" gorm:"type:varchar(50);not null"`
	CronExpression string                `json:"cron_expression" gorm:"type:varchar(100);not null"`
	Config        string                 `json:"config" gorm:"type:text"`
	Enabled       bool                   `json:"enabled" gorm:"default:true"`
	LastRun       *time.Time             `json:"last_run"`
	NextRun       *time.Time             `json:"next_run"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

func (s *StatisticsTask) TableName() string {
	return "statistics_tasks"
}

// TimeSeriesPoint 时序数据点
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Quality   int       `json:"quality"`
}

// TimeSeriesData 时序数据
type TimeSeriesData struct {
	PointID   string            `json:"point_id"`
	PointCode string            `json:"point_code"`
	Unit      string            `json:"unit"`
	Data      []TimeSeriesPoint `json:"data"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	
	// 时序数据库配置
	TSDBEnabled     bool
	TSDBHost        string
	TSDBPort        int
	TSDBName        string
	
	// 数据压缩配置
	CompressionEnabled bool
	CompressionDays    int
	
	// 归档配置
	ArchiveEnabled  bool
	ArchiveDays     int
}

// StatisticsStorage 统计存储接口
type StatisticsStorage interface {
	// 保存统计数据
	Save(ctx context.Context, data *StatisticsData) error
	SaveBatch(ctx context.Context, data []*StatisticsData) error
	
	// 查询统计数据
	Query(ctx context.Context, query *StatisticsQuery) ([]*StatisticsData, error)
	QueryLatest(ctx context.Context, dimension, dimensionValue, metricName string) (*StatisticsData, error)
	
	// 时序数据操作
	SaveTimeSeries(ctx context.Context, data *TimeSeriesData) error
	QueryTimeSeries(ctx context.Context, pointID string, start, end time.Time) (*TimeSeriesData, error)
	
	// 任务管理
	SaveTask(ctx context.Context, task *StatisticsTask) error
	GetTask(ctx context.Context, taskID string) (*StatisticsTask, error)
	ListTasks(ctx context.Context, enabled *bool) ([]*StatisticsTask, error)
	UpdateTaskRunTime(ctx context.Context, taskID string, lastRun, nextRun time.Time) error
	
	// 数据压缩和归档
	CompressData(ctx context.Context, before time.Time) error
	ArchiveData(ctx context.Context, before time.Time) error
	
	// 健康检查
	Ping(ctx context.Context) error
	Close() error
}

// StatisticsQuery 统计查询条件
type StatisticsQuery struct {
	TaskID         string
	Dimension      string
	DimensionValue string
	MetricName     string
	PeriodType     PeriodType
	PeriodStart    time.Time
	PeriodEnd      time.Time
	Limit          int
	Offset         int
	OrderBy        string
	OrderDesc      bool
}

// PostgreSQLStorage PostgreSQL存储实现
type PostgreSQLStorage struct {
	db       *gorm.DB
	config   StorageConfig
	compress *DataCompressor
	archive  *DataArchiver
	mu       sync.RWMutex
}

// NewPostgreSQLStorage 创建PostgreSQL存储
func NewPostgreSQLStorage(config StorageConfig) (*PostgreSQLStorage, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)
	
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}
	
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	
	storage := &PostgreSQLStorage{
		db:     db,
		config: config,
	}
	
	// 初始化压缩器
	if config.CompressionEnabled {
		storage.compress = NewDataCompressor(config.CompressionDays)
	}
	
	// 初始化归档器
	if config.ArchiveEnabled {
		storage.archive = NewDataArchiver(config.ArchiveDays)
	}
	
	return storage, nil
}

// Save 保存单条统计数据
func (s *PostgreSQLStorage) Save(ctx context.Context, data *StatisticsData) error {
	if data.ID == "" {
		data.ID = generateUUID()
	}
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}
	return s.db.WithContext(ctx).Create(data).Error
}

// SaveBatch 批量保存统计数据
func (s *PostgreSQLStorage) SaveBatch(ctx context.Context, data []*StatisticsData) error {
	if len(data) == 0 {
		return nil
	}
	
	now := time.Now()
	for _, d := range data {
		if d.ID == "" {
			d.ID = generateUUID()
		}
		if d.CreatedAt.IsZero() {
			d.CreatedAt = now
		}
	}
	
	// 分批插入，每批1000条
	batchSize := 1000
	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}
		batch := data[i:end]
		if err := s.db.WithContext(ctx).CreateInBatches(batch, len(batch)).Error; err != nil {
			return fmt.Errorf("batch insert failed at index %d: %w", i, err)
		}
	}
	
	return nil
}

// Query 查询统计数据
func (s *PostgreSQLStorage) Query(ctx context.Context, query *StatisticsQuery) ([]*StatisticsData, error) {
	db := s.db.WithContext(ctx).Model(&StatisticsData{})
	
	if query.TaskID != "" {
		db = db.Where("task_id = ?", query.TaskID)
	}
	if query.Dimension != "" {
		db = db.Where("dimension = ?", query.Dimension)
	}
	if query.DimensionValue != "" {
		db = db.Where("dimension_value = ?", query.DimensionValue)
	}
	if query.MetricName != "" {
		db = db.Where("metric_name = ?", query.MetricName)
	}
	if query.PeriodType != "" {
		db = db.Where("period_type = ?", query.PeriodType)
	}
	if !query.PeriodStart.IsZero() {
		db = db.Where("period_start >= ?", query.PeriodStart)
	}
	if !query.PeriodEnd.IsZero() {
		db = db.Where("period_end <= ?", query.PeriodEnd)
	}
	
	// 排序
	orderBy := "period_start"
	if query.OrderBy != "" {
		orderBy = query.OrderBy
	}
	orderDir := "ASC"
	if query.OrderDesc {
		orderDir = "DESC"
	}
	db = db.Order(fmt.Sprintf("%s %s", orderBy, orderDir))
	
	// 分页
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}
	if query.Offset > 0 {
		db = db.Offset(query.Offset)
	}
	
	var results []*StatisticsData
	if err := db.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	
	return results, nil
}

// QueryLatest 查询最新统计数据
func (s *PostgreSQLStorage) QueryLatest(ctx context.Context, dimension, dimensionValue, metricName string) (*StatisticsData, error) {
	var result StatisticsData
	err := s.db.WithContext(ctx).
		Where("dimension = ? AND dimension_value = ? AND metric_name = ?", dimension, dimensionValue, metricName).
		Order("period_start DESC").
		First(&result).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("query latest failed: %w", err)
	}
	
	return &result, nil
}

// SaveTimeSeries 保存时序数据
func (s *PostgreSQLStorage) SaveTimeSeries(ctx context.Context, data *TimeSeriesData) error {
	if len(data.Data) == 0 {
		return nil
	}
	
	// 使用批量插入优化性能
	records := make([]map[string]interface{}, 0, len(data.Data))
	now := time.Now()
	
	for _, point := range data.Data {
		records = append(records, map[string]interface{}{
			"id":         generateUUID(),
			"point_id":   data.PointID,
			"point_code": data.PointCode,
			"timestamp":  point.Timestamp,
			"value":      point.Value,
			"quality":    point.Quality,
			"created_at": now,
		})
	}
	
	// 分批插入
	batchSize := 1000
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]
		if err := s.db.WithContext(ctx).Table("time_series_data").Create(batch).Error; err != nil {
			return fmt.Errorf("save time series failed: %w", err)
		}
	}
	
	return nil
}

// QueryTimeSeries 查询时序数据
func (s *PostgreSQLStorage) QueryTimeSeries(ctx context.Context, pointID string, start, end time.Time) (*TimeSeriesData, error) {
	var results []struct {
		Timestamp time.Time
		Value     float64
		Quality   int
	}
	
	err := s.db.WithContext(ctx).
		Table("time_series_data").
		Select("timestamp, value, quality").
		Where("point_id = ? AND timestamp >= ? AND timestamp <= ?", pointID, start, end).
		Order("timestamp ASC").
		Find(&results).Error
	
	if err != nil {
		return nil, fmt.Errorf("query time series failed: %w", err)
	}
	
	data := &TimeSeriesData{
		PointID: pointID,
		Data:    make([]TimeSeriesPoint, 0, len(results)),
	}
	
	for _, r := range results {
		data.Data = append(data.Data, TimeSeriesPoint{
			Timestamp: r.Timestamp,
			Value:     r.Value,
			Quality:   r.Quality,
		})
	}
	
	return data, nil
}

// SaveTask 保存统计任务
func (s *PostgreSQLStorage) SaveTask(ctx context.Context, task *StatisticsTask) error {
	if task.ID == "" {
		task.ID = generateUUID()
	}
	now := time.Now()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	task.UpdatedAt = now
	
	return s.db.WithContext(ctx).Save(task).Error
}

// GetTask 获取统计任务
func (s *PostgreSQLStorage) GetTask(ctx context.Context, taskID string) (*StatisticsTask, error) {
	var task StatisticsTask
	err := s.db.WithContext(ctx).First(&task, "id = ?", taskID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get task failed: %w", err)
	}
	return &task, nil
}

// ListTasks 列出统计任务
func (s *PostgreSQLStorage) ListTasks(ctx context.Context, enabled *bool) ([]*StatisticsTask, error) {
	db := s.db.WithContext(ctx).Model(&StatisticsTask{})
	if enabled != nil {
		db = db.Where("enabled = ?", *enabled)
	}
	
	var tasks []*StatisticsTask
	if err := db.Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("list tasks failed: %w", err)
	}
	
	return tasks, nil
}

// UpdateTaskRunTime 更新任务运行时间
func (s *PostgreSQLStorage) UpdateTaskRunTime(ctx context.Context, taskID string, lastRun, nextRun time.Time) error {
	return s.db.WithContext(ctx).
		Model(&StatisticsTask{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"last_run":  lastRun,
			"next_run":  nextRun,
			"updated_at": time.Now(),
		}).Error
}

// CompressData 压缩数据
func (s *PostgreSQLStorage) CompressData(ctx context.Context, before time.Time) error {
	if s.compress == nil {
		return fmt.Errorf("compression not enabled")
	}
	return s.compress.Compress(ctx, s.db, before)
}

// ArchiveData 归档数据
func (s *PostgreSQLStorage) ArchiveData(ctx context.Context, before time.Time) error {
	if s.archive == nil {
		return fmt.Errorf("archive not enabled")
	}
	return s.archive.Archive(ctx, s.db, before)
}

// Ping 健康检查
func (s *PostgreSQLStorage) Ping(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Close 关闭连接
func (s *PostgreSQLStorage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// DataCompressor 数据压缩器
type DataCompressor struct {
	compressionDays int
}

// NewDataCompressor 创建数据压缩器
func NewDataCompressor(compressionDays int) *DataCompressor {
	return &DataCompressor{
		compressionDays: compressionDays,
	}
}

// Compress 压缩历史数据
func (c *DataCompressor) Compress(ctx context.Context, db *gorm.DB, before time.Time) error {
	// 按维度和指标分组压缩
	var groups []struct {
		Dimension      string
		DimensionValue string
		MetricName     string
		PeriodType     PeriodType
	}
	
	err := db.WithContext(ctx).
		Model(&StatisticsData{}).
		Select("DISTINCT dimension, dimension_value, metric_name, period_type").
		Where("period_start < ?", before).
		Find(&groups).Error
	
	if err != nil {
		return fmt.Errorf("query groups failed: %w", err)
	}
	
	for _, group := range groups {
		// 聚合压缩数据
		var compressed struct {
			MinValue float64
			MaxValue float64
			AvgValue float64
			Count    int64
		}
		
		err := db.WithContext(ctx).
			Model(&StatisticsData{}).
			Select("MIN(metric_value) as min_value, MAX(metric_value) as max_value, AVG(metric_value) as avg_value, COUNT(*) as count").
			Where("dimension = ? AND dimension_value = ? AND metric_name = ? AND period_type = ? AND period_start < ?",
				group.Dimension, group.DimensionValue, group.MetricName, group.PeriodType, before).
			Scan(&compressed).Error
		
		if err != nil {
			return fmt.Errorf("compress group failed: %w", err)
		}
		
		if compressed.Count == 0 {
			continue
		}
		
		// 创建压缩记录
		metadata := map[string]interface{}{
			"min_value": compressed.MinValue,
			"max_value": compressed.MaxValue,
			"count":     compressed.Count,
		}
		metadataJSON, _ := json.Marshal(metadata)
		
		compressedData := &StatisticsData{
			ID:             generateUUID(),
			Dimension:      group.Dimension,
			DimensionValue: group.DimensionValue,
			MetricName:     group.MetricName + "_compressed",
			MetricValue:    compressed.AvgValue,
			Metadata:       string(metadataJSON),
			PeriodType:     PeriodTypeCustom,
			PeriodStart:    before.AddDate(0, 0, -c.compressionDays),
			PeriodEnd:      before,
			CreatedAt:      time.Now(),
		}
		
		// 事务：删除原数据，插入压缩数据
		err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// 删除原数据
			if err := tx.Where("dimension = ? AND dimension_value = ? AND metric_name = ? AND period_type = ? AND period_start < ?",
				group.Dimension, group.DimensionValue, group.MetricName, group.PeriodType, before).
				Delete(&StatisticsData{}).Error; err != nil {
				return err
			}
			
			// 插入压缩数据
			return tx.Create(compressedData).Error
		})
		
		if err != nil {
			return fmt.Errorf("save compressed data failed: %w", err)
		}
	}
	
	return nil
}

// DataArchiver 数据归档器
type DataArchiver struct {
	archiveDays int
}

// NewDataArchiver 创建数据归档器
func NewDataArchiver(archiveDays int) *DataArchiver {
	return &DataArchiver{
		archiveDays: archiveDays,
	}
}

// Archive 归档历史数据
func (a *DataArchiver) Archive(ctx context.Context, db *gorm.DB, before time.Time) error {
	// 创建归档表
	archiveTable := fmt.Sprintf("statistics_data_archive_%s", before.Format("200601"))
	
	err := db.WithContext(ctx).Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s AS 
		SELECT * FROM statistics_data WHERE 1=0
	`, archiveTable)).Error
	if err != nil {
		return fmt.Errorf("create archive table failed: %w", err)
	}
	
	// 迁移数据到归档表
	err = db.WithContext(ctx).Exec(fmt.Sprintf(`
		INSERT INTO %s 
		SELECT * FROM statistics_data 
		WHERE period_start < ?
	`, archiveTable), before).Error
	if err != nil {
		return fmt.Errorf("archive data failed: %w", err)
	}
	
	// 删除已归档的数据
	err = db.WithContext(ctx).
		Where("period_start < ?", before).
		Delete(&StatisticsData{}).Error
	if err != nil {
		return fmt.Errorf("delete archived data failed: %w", err)
	}
	
	return nil
}

// TimeSeriesDBStorage 时序数据库存储实现
type TimeSeriesDBStorage struct {
	pgStorage *PostgreSQLStorage
	config    StorageConfig
}

// NewTimeSeriesDBStorage 创建时序数据库存储
func NewTimeSeriesDBStorage(pgStorage *PostgreSQLStorage, config StorageConfig) (*TimeSeriesDBStorage, error) {
	return &TimeSeriesDBStorage{
		pgStorage: pgStorage,
		config:    config,
	}, nil
}

// Save 保存统计数据
func (s *TimeSeriesDBStorage) Save(ctx context.Context, data *StatisticsData) error {
	return s.pgStorage.Save(ctx, data)
}

// SaveBatch 批量保存统计数据
func (s *TimeSeriesDBStorage) SaveBatch(ctx context.Context, data []*StatisticsData) error {
	return s.pgStorage.SaveBatch(ctx, data)
}

// Query 查询统计数据
func (s *TimeSeriesDBStorage) Query(ctx context.Context, query *StatisticsQuery) ([]*StatisticsData, error) {
	return s.pgStorage.Query(ctx, query)
}

// QueryLatest 查询最新统计数据
func (s *TimeSeriesDBStorage) QueryLatest(ctx context.Context, dimension, dimensionValue, metricName string) (*StatisticsData, error) {
	return s.pgStorage.QueryLatest(ctx, dimension, dimensionValue, metricName)
}

// SaveTimeSeries 保存时序数据（优化实现）
func (s *TimeSeriesDBStorage) SaveTimeSeries(ctx context.Context, data *TimeSeriesData) error {
	// 时序数据库优化：使用COPY协议批量写入
	// 这里简化实现，实际可使用TimescaleDB或其他时序数据库
	return s.pgStorage.SaveTimeSeries(ctx, data)
}

// QueryTimeSeries 查询时序数据（优化实现）
func (s *TimeSeriesDBStorage) QueryTimeSeries(ctx context.Context, pointID string, start, end time.Time) (*TimeSeriesData, error) {
	// 时序数据库优化：使用时序索引查询
	return s.pgStorage.QueryTimeSeries(ctx, pointID, start, end)
}

// SaveTask 保存统计任务
func (s *TimeSeriesDBStorage) SaveTask(ctx context.Context, task *StatisticsTask) error {
	return s.pgStorage.SaveTask(ctx, task)
}

// GetTask 获取统计任务
func (s *TimeSeriesDBStorage) GetTask(ctx context.Context, taskID string) (*StatisticsTask, error) {
	return s.pgStorage.GetTask(ctx, taskID)
}

// ListTasks 列出统计任务
func (s *TimeSeriesDBStorage) ListTasks(ctx context.Context, enabled *bool) ([]*StatisticsTask, error) {
	return s.pgStorage.ListTasks(ctx, enabled)
}

// UpdateTaskRunTime 更新任务运行时间
func (s *TimeSeriesDBStorage) UpdateTaskRunTime(ctx context.Context, taskID string, lastRun, nextRun time.Time) error {
	return s.pgStorage.UpdateTaskRunTime(ctx, taskID, lastRun, nextRun)
}

// CompressData 压缩数据
func (s *TimeSeriesDBStorage) CompressData(ctx context.Context, before time.Time) error {
	return s.pgStorage.CompressData(ctx, before)
}

// ArchiveData 归档数据
func (s *TimeSeriesDBStorage) ArchiveData(ctx context.Context, before time.Time) error {
	return s.pgStorage.ArchiveData(ctx, before)
}

// Ping 健康检查
func (s *TimeSeriesDBStorage) Ping(ctx context.Context) error {
	return s.pgStorage.Ping(ctx)
}

// Close 关闭连接
func (s *TimeSeriesDBStorage) Close() error {
	return s.pgStorage.Close()
}

// generateUUID 生成UUID
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// StatisticsResult 统计结果
type StatisticsResult struct {
	Dimension      string                 `json:"dimension"`
	DimensionValue string                 `json:"dimension_value"`
	Metrics        map[string]float64     `json:"metrics"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	PeriodStart    time.Time              `json:"period_start"`
	PeriodEnd      time.Time              `json:"period_end"`
	PeriodType     PeriodType             `json:"period_type"`
}

// ToStatisticsData 转换为存储格式
func (r *StatisticsResult) ToStatisticsData(taskID string) []*StatisticsData {
	var results []*StatisticsData
	now := time.Now()
	
	for name, value := range r.Metrics {
		var metadata string
		if r.Metadata != nil {
			if bytes, err := json.Marshal(r.Metadata); err == nil {
				metadata = string(bytes)
			}
		}
		
		results = append(results, &StatisticsData{
			ID:             generateUUID(),
			TaskID:         taskID,
			Dimension:      r.Dimension,
			DimensionValue: r.DimensionValue,
			MetricName:     name,
			MetricValue:    value,
			Metadata:       metadata,
			PeriodType:     r.PeriodType,
			PeriodStart:    r.PeriodStart,
			PeriodEnd:      r.PeriodEnd,
			CreatedAt:      now,
		})
	}
	
	return results
}

// AggregatedStatistics 聚合统计结果
type AggregatedStatistics struct {
	Sum     float64 `json:"sum"`
	Avg     float64 `json:"avg"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Count   int64   `json:"count"`
	StdDev  float64 `json:"std_dev"`
	Variance float64 `json:"variance"`
}

// CalculateAggregated 计算聚合统计
func CalculateAggregated(values []float64) *AggregatedStatistics {
	if len(values) == 0 {
		return &AggregatedStatistics{}
	}
	
	stats := &AggregatedStatistics{
		Count: int64(len(values)),
		Min:   values[0],
		Max:   values[0],
	}
	
	var sum float64
	for _, v := range values {
		sum += v
		if v < stats.Min {
			stats.Min = v
		}
		if v > stats.Max {
			stats.Max = v
		}
	}
	
	stats.Sum = sum
	stats.Avg = sum / float64(len(values))
	
	// 计算方差和标准差
	var varianceSum float64
	for _, v := range values {
		diff := v - stats.Avg
		varianceSum += diff * diff
	}
	stats.Variance = varianceSum / float64(len(values))
	stats.StdDev = sqrt(stats.Variance)
	
	return stats
}

// sqrt 简单平方根实现
func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x == 0 {
		return 0
	}
	
	// 牛顿迭代法
	z := x
	for i := 0; i < 100; i++ {
		z = (z + x/z) / 2
		if z*z-x < 1e-10 && z*z-x > -1e-10 {
			break
		}
	}
	return z
}
