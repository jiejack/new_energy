package lifecycle

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ArchiveFormat 归档格式
type ArchiveFormat string

const (
	ArchiveFormatParquet ArchiveFormat = "parquet"
	ArchiveFormatArrow   ArchiveFormat = "arrow"
	ArchiveFormatJSON    ArchiveFormat = "json"
	ArchiveFormatCSV     ArchiveFormat = "csv"
)

// ArchiveStatus 归档状态
type ArchiveStatus string

const (
	ArchiveStatusPending    ArchiveStatus = "pending"
	ArchiveStatusRunning    ArchiveStatus = "running"
	ArchiveStatusCompleted  ArchiveStatus = "completed"
	ArchiveStatusFailed     ArchiveStatus = "failed"
	ArchiveStatusValidating ArchiveStatus = "validating"
)

// ArchivePolicy 归档策略
type ArchivePolicy struct {
	ID              string        `json:"id" gorm:"primaryKey"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	DataType        string        `json:"data_type"`       // 数据类型
	RetentionDays   int           `json:"retention_days"`  // 保留天数
	ArchiveAfter    time.Duration `json:"archive_after"`   // 归档时间阈值
	Format          ArchiveFormat `json:"format"`          // 归档格式
	Compression     bool          `json:"compression"`     // 是否压缩
	BatchSize       int           `json:"batch_size"`      // 批次大小
	Schedule        string        `json:"schedule"`        // 调度表达式 (cron)
	Destination     string        `json:"destination"`     // 目标路径
	Enabled         bool          `json:"enabled"`
	CreatedAt       time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

// ArchiveTask 归档任务
type ArchiveTask struct {
	ID           string        `json:"id" gorm:"primaryKey"`
	PolicyID     string        `json:"policy_id" gorm:"index"`
	Status       ArchiveStatus `json:"status"`
	StartTime    *time.Time    `json:"start_time"`
	EndTime      *time.Time    `json:"end_time"`
	RecordsCount int64         `json:"records_count"`
	DataSize     int64         `json:"data_size"`      // 字节
	FilePath     string        `json:"file_path"`      // 归档文件路径
	Checksum     string        `json:"checksum"`       // 校验和
	Error        string        `json:"error"`          // 错误信息
	CreatedAt    time.Time     `json:"created_at" gorm:"autoCreateTime"`
}

// ArchiveRecord 归档记录
type ArchiveRecord struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TaskID      string    `json:"task_id" gorm:"index"`
	OriginalID  string    `json:"original_id"`  // 原始数据ID
	DataType    string    `json:"data_type"`    // 数据类型
	ArchivedAt  time.Time `json:"archived_at"`
	FilePath    string    `json:"file_path"`
	Offset      int64     `json:"offset"`       // 文件偏移量
	Size        int64     `json:"size"`         // 记录大小
	Checksum    string    `json:"checksum"`     // 记录校验和
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ArchiveConfig 归档配置
type ArchiveConfig struct {
	StoragePath      string        `yaml:"storage_path" json:"storage_path"`           // 归档存储路径
	TempPath         string        `yaml:"temp_path" json:"temp_path"`                 // 临时文件路径
	MaxFileSize      int64         `yaml:"max_file_size" json:"max_file_size"`         // 单文件最大大小
	DefaultFormat    ArchiveFormat `yaml:"default_format" json:"default_format"`       // 默认格式
	CompressionLevel int           `yaml:"compression_level" json:"compression_level"` // 压缩级别
	EnableVerify     bool          `yaml:"enable_verify" json:"enable_verify"`         // 启用验证
	ParallelWorkers  int           `yaml:"parallel_workers" json:"parallel_workers"`   // 并行工作数
	RetryCount       int           `yaml:"retry_count" json:"retry_count"`             // 重试次数
	RetryDelay       time.Duration `yaml:"retry_delay" json:"retry_delay"`             // 重试延迟
}

// DefaultArchiveConfig 默认归档配置
func DefaultArchiveConfig() ArchiveConfig {
	return ArchiveConfig{
		StoragePath:      "/data/archive",
		TempPath:         "/tmp/archive",
		MaxFileSize:      1024 * 1024 * 1024, // 1GB
		DefaultFormat:    ArchiveFormatJSON,
		CompressionLevel: gzip.DefaultCompression,
		EnableVerify:     true,
		ParallelWorkers:  4,
		RetryCount:       3,
		RetryDelay:       5 * time.Second,
	}
}

// DataArchiver 数据归档器
type DataArchiver struct {
	config    ArchiveConfig
	db        *gorm.DB
	logger    *zap.Logger

	// 任务队列
	taskQueue chan *ArchiveTask
	taskMu    sync.RWMutex

	// 控制通道
	stopCh    chan struct{}
	wg        sync.WaitGroup

	// 指标
	metrics   *ArchiveMetrics
}

// ArchiveMetrics 归档指标
type ArchiveMetrics struct {
	mu                sync.RWMutex
	TotalTasks        int64
	CompletedTasks    int64
	FailedTasks       int64
	TotalRecords      int64
	TotalDataSize     int64
	LastArchiveTime   time.Time
	AverageDuration   time.Duration
	CompressionRatio  float64
}

// NewDataArchiver 创建数据归档器
func NewDataArchiver(config ArchiveConfig, db *gorm.DB, logger *zap.Logger) *DataArchiver {
	return &DataArchiver{
		config:    config,
		db:        db,
		logger:    logger,
		taskQueue: make(chan *ArchiveTask, 1000),
		stopCh:    make(chan struct{}),
		metrics:   &ArchiveMetrics{},
	}
}

// Start 启动归档器
func (da *DataArchiver) Start(ctx context.Context) error {
	da.logger.Info("Starting data archiver")

	// 确保目录存在
	if err := os.MkdirAll(da.config.StoragePath, 0755); err != nil {
		return fmt.Errorf("failed to create storage path: %w", err)
	}
	if err := os.MkdirAll(da.config.TempPath, 0755); err != nil {
		return fmt.Errorf("failed to create temp path: %w", err)
	}

	// 启动工作协程
	for i := 0; i < da.config.ParallelWorkers; i++ {
		da.wg.Add(1)
		go da.archiveWorker(ctx, i)
	}

	// 启动调度协程
	da.wg.Add(1)
	go da.schedulerWorker(ctx)

	// 启动指标更新协程
	da.wg.Add(1)
	go da.metricsWorker(ctx)

	da.logger.Info("Data archiver started")
	return nil
}

// Stop 停止归档器
func (da *DataArchiver) Stop() error {
	da.logger.Info("Stopping data archiver")
	close(da.stopCh)
	da.wg.Wait()
	da.logger.Info("Data archiver stopped")
	return nil
}

// CreatePolicy 创建归档策略
func (da *DataArchiver) CreatePolicy(ctx context.Context, policy *ArchivePolicy) error {
	if policy.ID == "" {
		policy.ID = uuid.New().String()
	}
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	if err := da.db.Create(policy).Error; err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	da.logger.Info("Archive policy created",
		zap.String("policy_id", policy.ID),
		zap.String("name", policy.Name),
	)

	return nil
}

// UpdatePolicy 更新归档策略
func (da *DataArchiver) UpdatePolicy(ctx context.Context, policy *ArchivePolicy) error {
	policy.UpdatedAt = time.Now()

	if err := da.db.Save(policy).Error; err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	da.logger.Info("Archive policy updated",
		zap.String("policy_id", policy.ID),
	)

	return nil
}

// DeletePolicy 删除归档策略
func (da *DataArchiver) DeletePolicy(ctx context.Context, policyID string) error {
	if err := da.db.Delete(&ArchivePolicy{}, "id = ?", policyID).Error; err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	da.logger.Info("Archive policy deleted",
		zap.String("policy_id", policyID),
	)

	return nil
}

// GetPolicy 获取归档策略
func (da *DataArchiver) GetPolicy(ctx context.Context, policyID string) (*ArchivePolicy, error) {
	var policy ArchivePolicy
	if err := da.db.First(&policy, "id = ?", policyID).Error; err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}
	return &policy, nil
}

// ListPolicies 列出归档策略
func (da *DataArchiver) ListPolicies(ctx context.Context) ([]ArchivePolicy, error) {
	var policies []ArchivePolicy
	if err := da.db.Find(&policies).Error; err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}
	return policies, nil
}

// TriggerArchive 触发归档
func (da *DataArchiver) TriggerArchive(ctx context.Context, policyID string) (*ArchiveTask, error) {
	// 获取策略
	policy, err := da.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, err
	}

	if !policy.Enabled {
		return nil, fmt.Errorf("policy %s is disabled", policyID)
	}

	// 创建任务
	task := &ArchiveTask{
		ID:        uuid.New().String(),
		PolicyID:  policyID,
		Status:    ArchiveStatusPending,
		CreatedAt: time.Now(),
	}

	if err := da.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// 加入队列
	da.taskQueue <- task

	da.logger.Info("Archive task triggered",
		zap.String("task_id", task.ID),
		zap.String("policy_id", policyID),
	)

	return task, nil
}

// GetTask 获取任务状态
func (da *DataArchiver) GetTask(ctx context.Context, taskID string) (*ArchiveTask, error) {
	var task ArchiveTask
	if err := da.db.First(&task, "id = ?", taskID).Error; err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

// ListTasks 列出任务
func (da *DataArchiver) ListTasks(ctx context.Context, policyID string, limit int) ([]ArchiveTask, error) {
	var tasks []ArchiveTask
	query := da.db.Model(&ArchiveTask{}).Order("created_at DESC")

	if policyID != "" {
		query = query.Where("policy_id = ?", policyID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, nil
}

// RestoreArchive 恢复归档数据
func (da *DataArchiver) RestoreArchive(ctx context.Context, taskID string, targetTable string) error {
	task, err := da.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	if task.Status != ArchiveStatusCompleted {
		return fmt.Errorf("task %s is not completed", taskID)
	}

	da.logger.Info("Restoring archive",
		zap.String("task_id", taskID),
		zap.String("target_table", targetTable),
	)

	// 打开归档文件
	file, err := os.Open(task.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open archive file: %w", err)
	}
	defer file.Close()

	// 解压
	var reader io.Reader = file
	if filepath.Ext(task.FilePath) == ".gz" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// 读取并恢复数据
	decoder := json.NewDecoder(reader)
	batch := make([]map[string]interface{}, 0, 1000)
	count := 0

	for decoder.More() {
		var record map[string]interface{}
		if err := decoder.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode record: %w", err)
		}

		batch = append(batch, record)
		count++

		// 批量插入
		if len(batch) >= 1000 {
			if err := da.db.Table(targetTable).Create(&batch).Error; err != nil {
				return fmt.Errorf("failed to insert batch: %w", err)
			}
			batch = batch[:0]
		}
	}

	// 插入剩余数据
	if len(batch) > 0 {
		if err := da.db.Table(targetTable).Create(&batch).Error; err != nil {
			return fmt.Errorf("failed to insert remaining batch: %w", err)
		}
	}

	da.logger.Info("Archive restored",
		zap.String("task_id", taskID),
		zap.Int("records_restored", count),
	)

	return nil
}

// GetMetrics 获取指标
func (da *DataArchiver) GetMetrics() ArchiveMetrics {
	da.metrics.mu.RLock()
	defer da.metrics.mu.RUnlock()
	return *da.metrics
}

// 内部方法

func (da *DataArchiver) archiveWorker(ctx context.Context, workerID int) {
	defer da.wg.Done()

	da.logger.Debug("Archive worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-ctx.Done():
			return
		case <-da.stopCh:
			return
		case task := <-da.taskQueue:
			da.executeArchiveTask(ctx, task)
		}
	}
}

func (da *DataArchiver) executeArchiveTask(ctx context.Context, task *ArchiveTask) {
	da.logger.Info("Executing archive task",
		zap.String("task_id", task.ID),
		zap.String("policy_id", task.PolicyID),
	)

	// 更新任务状态
	now := time.Now()
	task.Status = ArchiveStatusRunning
	task.StartTime = &now
	da.db.Save(task)

	// 获取策略
	policy, err := da.GetPolicy(ctx, task.PolicyID)
	if err != nil {
		da.failTask(task, fmt.Sprintf("Failed to get policy: %v", err))
		return
	}

	// 执行归档
	var lastErr error
	for i := 0; i < da.config.RetryCount; i++ {
		if i > 0 {
			time.Sleep(da.config.RetryDelay)
		}

		lastErr = da.doArchive(ctx, task, policy)
		if lastErr == nil {
			break
		}

		da.logger.Warn("Archive task failed, retrying",
			zap.String("task_id", task.ID),
			zap.Int("attempt", i+1),
			zap.Error(lastErr),
		)
	}

	if lastErr != nil {
		da.failTask(task, fmt.Sprintf("Archive failed after %d retries: %v", da.config.RetryCount, lastErr))
		return
	}

	// 验证归档
	if da.config.EnableVerify {
		task.Status = ArchiveStatusValidating
		da.db.Save(task)

		if err := da.verifyArchive(task); err != nil {
			da.failTask(task, fmt.Sprintf("Archive verification failed: %v", err))
			return
		}
	}

	// 完成任务
	endTime := time.Now()
	task.Status = ArchiveStatusCompleted
	task.EndTime = &endTime
	da.db.Save(task)

	// 更新指标
	da.metrics.mu.Lock()
	da.metrics.CompletedTasks++
	da.metrics.TotalRecords += task.RecordsCount
	da.metrics.TotalDataSize += task.DataSize
	da.metrics.LastArchiveTime = endTime
	if da.metrics.CompletedTasks > 0 {
		da.metrics.AverageDuration = time.Duration(
			int64(da.metrics.AverageDuration)*int64(da.metrics.CompletedTasks-1)/int64(da.metrics.CompletedTasks) +
				int64(endTime.Sub(now))/int64(da.metrics.CompletedTasks),
		)
	}
	da.metrics.mu.Unlock()

	da.logger.Info("Archive task completed",
		zap.String("task_id", task.ID),
		zap.Int64("records", task.RecordsCount),
		zap.Int64("size", task.DataSize),
	)
}

func (da *DataArchiver) doArchive(ctx context.Context, task *ArchiveTask, policy *ArchivePolicy) error {
	// 查询需要归档的数据
	cutoffTime := time.Now().Add(-policy.ArchiveAfter)

	var records []map[string]interface{}
	query := da.db.Table(policy.DataType).
		Where("created_at < ?", cutoffTime).
		Order("created_at ASC").
		Limit(policy.BatchSize)

	if err := query.Find(&records).Error; err != nil {
		return fmt.Errorf("failed to query data: %w", err)
	}

	if len(records) == 0 {
		da.logger.Info("No data to archive",
			zap.String("task_id", task.ID),
			zap.String("data_type", policy.DataType),
		)
		return nil
	}

	// 创建归档文件
	fileName := fmt.Sprintf("%s_%s.json", policy.DataType, time.Now().Format("20060102_150405"))
	if policy.Compression {
		fileName += ".gz"
	}
	filePath := filepath.Join(da.config.StoragePath, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer file.Close()

	var writer io.Writer = file
	var hashWriter = sha256.New()

	if policy.Compression {
		gzWriter := gzip.NewWriter(file)
		gzWriter.Level = da.config.CompressionLevel
		defer gzWriter.Close()
		writer = io.MultiWriter(gzWriter, hashWriter)
	} else {
		writer = io.MultiWriter(file, hashWriter)
	}

	// 写入数据
	encoder := json.NewEncoder(writer)
	var totalSize int64

	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			return fmt.Errorf("failed to encode record: %w", err)
		}

		// 记录归档信息
		recordBytes, _ := json.Marshal(record)
		totalSize += int64(len(recordBytes))
	}

	// 计算校验和
	checksum := fmt.Sprintf("%x", hashWriter.Sum(nil))

	// 更新任务信息
	task.RecordsCount = int64(len(records))
	task.DataSize = totalSize
	task.FilePath = filePath
	task.Checksum = checksum

	// 删除已归档的数据
	ids := make([]string, len(records))
	for i, record := range records {
		if id, ok := record["id"].(string); ok {
			ids[i] = id
		}
	}

	if err := da.db.Table(policy.DataType).Where("id IN ?", ids).Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to delete archived data: %w", err)
	}

	// 记录归档记录
	for _, record := range records {
		archiveRecord := ArchiveRecord{
			ID:         uuid.New().String(),
			TaskID:     task.ID,
			OriginalID: fmt.Sprintf("%v", record["id"]),
			DataType:   policy.DataType,
			ArchivedAt: time.Now(),
			FilePath:   filePath,
			CreatedAt:  time.Now(),
		}
		da.db.Create(&archiveRecord)
	}

	return nil
}

func (da *DataArchiver) verifyArchive(task *ArchiveTask) error {
	// 打开文件
	file, err := os.Open(task.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open archive file: %w", err)
	}
	defer file.Close()

	// 计算校验和
	hash := sha256.New()
	var reader io.Reader = file

	if filepath.Ext(task.FilePath) == ".gz" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// 统计记录数
	decoder := json.NewDecoder(io.TeeReader(reader, hash))
	count := int64(0)

	for decoder.More() {
		var record map[string]interface{}
		if err := decoder.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode record: %w", err)
		}
		count++
	}

	// 验证记录数
	if count != task.RecordsCount {
		return fmt.Errorf("record count mismatch: expected %d, got %d", task.RecordsCount, count)
	}

	// 验证校验和
	checksum := fmt.Sprintf("%x", hash.Sum(nil))
	if checksum != task.Checksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", task.Checksum, checksum)
	}

	return nil
}

func (da *DataArchiver) failTask(task *ArchiveTask, errMsg string) {
	endTime := time.Now()
	task.Status = ArchiveStatusFailed
	task.Error = errMsg
	task.EndTime = &endTime
	da.db.Save(task)

	da.metrics.mu.Lock()
	da.metrics.FailedTasks++
	da.metrics.mu.Unlock()

	da.logger.Error("Archive task failed",
		zap.String("task_id", task.ID),
		zap.String("error", errMsg),
	)
}

func (da *DataArchiver) schedulerWorker(ctx context.Context) {
	defer da.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-da.stopCh:
			return
		case <-ticker.C:
			da.checkScheduledArchives(ctx)
		}
	}
}

func (da *DataArchiver) checkScheduledArchives(ctx context.Context) {
	// 获取所有启用的策略
	var policies []ArchivePolicy
	if err := da.db.Where("enabled = ?", true).Find(&policies).Error; err != nil {
		da.logger.Error("Failed to get enabled policies", zap.Error(err))
		return
	}

	for _, policy := range policies {
		// 检查是否需要执行归档
		// 这里简化处理，实际应该解析 cron 表达式
		cutoffTime := time.Now().Add(-policy.ArchiveAfter)

		var count int64
		da.db.Table(policy.DataType).Where("created_at < ?", cutoffTime).Count(&count)

		if count > 0 {
			da.logger.Info("Found data to archive",
				zap.String("policy_id", policy.ID),
				zap.String("data_type", policy.DataType),
				zap.Int64("count", count),
			)

			// 触发归档
			da.TriggerArchive(ctx, policy.ID)
		}
	}
}

func (da *DataArchiver) metricsWorker(ctx context.Context) {
	defer da.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-da.stopCh:
			return
		case <-ticker.C:
			da.updateMetrics(ctx)
		}
	}
}

func (da *DataArchiver) updateMetrics(ctx context.Context) {
	var totalTasks, completedTasks, failedTasks int64

	da.db.Model(&ArchiveTask{}).Count(&totalTasks)
	da.db.Model(&ArchiveTask{}).Where("status = ?", ArchiveStatusCompleted).Count(&completedTasks)
	da.db.Model(&ArchiveTask{}).Where("status = ?", ArchiveStatusFailed).Count(&failedTasks)

	da.metrics.mu.Lock()
	da.metrics.TotalTasks = totalTasks
	da.metrics.CompletedTasks = completedTasks
	da.metrics.FailedTasks = failedTasks
	da.metrics.mu.Unlock()
}

// ValidateArchiveFile 验证归档文件
func (da *DataArchiver) ValidateArchiveFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file
	if filepath.Ext(filePath) == ".gz" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	decoder := json.NewDecoder(reader)
	count := 0

	for decoder.More() {
		var record map[string]interface{}
		if err := decoder.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("invalid JSON at record %d: %w", count+1, err)
		}
		count++
	}

	da.logger.Info("Archive file validated",
		zap.String("file", filePath),
		zap.Int("records", count),
	)

	return nil
}

// GetArchiveStats 获取归档统计
func (da *DataArchiver) GetArchiveStats(ctx context.Context, policyID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalRecords, totalSize int64
	query := da.db.Model(&ArchiveTask{}).Where("status = ?", ArchiveStatusCompleted)

	if policyID != "" {
		query = query.Where("policy_id = ?", policyID)
	}

	if err := query.Select("SUM(records_count)").Scan(&totalRecords).Error; err != nil {
		return nil, err
	}

	if err := query.Select("SUM(data_size)").Scan(&totalSize).Error; err != nil {
		return nil, err
	}

	stats["total_records"] = totalRecords
	stats["total_size"] = totalSize
	stats["total_size_mb"] = totalSize / (1024 * 1024)

	return stats, nil
}

// CompactArchives 压缩归档文件
func (da *DataArchiver) CompactArchives(ctx context.Context, policyID string) error {
	policy, err := da.GetPolicy(ctx, policyID)
	if err != nil {
		return err
	}

	// 获取该策略的所有归档文件
	var tasks []ArchiveTask
	if err := da.db.Where("policy_id = ? AND status = ?", policyID, ArchiveStatusCompleted).
		Order("created_at ASC").Find(&tasks).Error; err != nil {
		return err
	}

	if len(tasks) < 2 {
		da.logger.Info("Not enough archives to compact",
			zap.String("policy_id", policyID),
		)
		return nil
	}

	da.logger.Info("Compacting archives",
		zap.String("policy_id", policyID),
		zap.Int("files", len(tasks)),
	)

	// 创建合并文件
	mergedFileName := fmt.Sprintf("%s_merged_%s.json", policy.DataType, time.Now().Format("20060102_150405"))
	if policy.Compression {
		mergedFileName += ".gz"
	}
	mergedFilePath := filepath.Join(da.config.StoragePath, mergedFileName)

	mergedFile, err := os.Create(mergedFilePath)
	if err != nil {
		return fmt.Errorf("failed to create merged file: %w", err)
	}
	defer mergedFile.Close()

	var mergedWriter io.Writer = mergedFile
	if policy.Compression {
		gzWriter := gzip.NewWriter(mergedFile)
		gzWriter.Level = da.config.CompressionLevel
		defer gzWriter.Close()
		mergedWriter = gzWriter
	}

	encoder := json.NewEncoder(mergedWriter)
	var totalRecords int64
	var totalSize int64

	// 合并所有文件
	for _, task := range tasks {
		file, err := os.Open(task.FilePath)
		if err != nil {
			da.logger.Warn("Failed to open archive file, skipping",
				zap.String("file", task.FilePath),
				zap.Error(err),
			)
			continue
		}

		var reader io.Reader = file
		if filepath.Ext(task.FilePath) == ".gz" {
			gzReader, err := gzip.NewReader(file)
			if err != nil {
				file.Close()
				continue
			}
			reader = gzReader
			defer gzReader.Close()
		}

		decoder := json.NewDecoder(reader)
		for decoder.More() {
			var record map[string]interface{}
			if err := decoder.Decode(&record); err != nil {
				if err == io.EOF {
					break
				}
				continue
			}

			if err := encoder.Encode(record); err != nil {
				da.logger.Warn("Failed to encode record",
					zap.Error(err),
				)
				continue
			}

			totalRecords++
		}

		file.Close()
		totalSize += task.DataSize

		// 删除原文件
		os.Remove(task.FilePath)
	}

	// 更新任务记录
	// 创建新的合并任务
	mergedTask := &ArchiveTask{
		ID:           uuid.New().String(),
		PolicyID:     policyID,
		Status:       ArchiveStatusCompleted,
		RecordsCount: totalRecords,
		DataSize:     totalSize,
		FilePath:     mergedFilePath,
		CreatedAt:    time.Now(),
	}
	now := time.Now()
	mergedTask.StartTime = &now
	mergedTask.EndTime = &now

	da.db.Create(mergedTask)

	// 删除旧任务记录
	da.db.Where("policy_id = ? AND id != ?", policyID, mergedTask.ID).Delete(&ArchiveTask{})

	da.logger.Info("Archives compacted",
		zap.String("policy_id", policyID),
		zap.Int64("total_records", totalRecords),
		zap.Int64("total_size", totalSize),
	)

	return nil
}

// ArchiveData 直接归档数据
func (da *DataArchiver) ArchiveData(ctx context.Context, dataType string, records []map[string]interface{}) (string, error) {
	if len(records) == 0 {
		return "", fmt.Errorf("no records to archive")
	}

	// 创建归档文件
	fileName := fmt.Sprintf("%s_manual_%s.json", dataType, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(da.config.StoragePath, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create archive file: %w", err)
	}
	defer file.Close()

	// 写入数据
	encoder := json.NewEncoder(file)
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			return "", fmt.Errorf("failed to encode record: %w", err)
		}
	}

	da.logger.Info("Data archived manually",
		zap.String("file", filePath),
		zap.Int("records", len(records)),
	)

	return filePath, nil
}

// ReadArchive 读取归档数据
func (da *DataArchiver) ReadArchive(ctx context.Context, filePath string) ([]map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive file: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file
	if filepath.Ext(filePath) == ".gz" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	var records []map[string]interface{}
	decoder := json.NewDecoder(reader)

	for decoder.More() {
		var record map[string]interface{}
		if err := decoder.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode record: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

// StreamArchive 流式读取归档数据
func (da *DataArchiver) StreamArchive(ctx context.Context, filePath string, handler func(record map[string]interface{}) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open archive file: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file
	if filepath.Ext(filePath) == ".gz" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	decoder := json.NewDecoder(reader)
	count := 0

	for decoder.More() {
		var record map[string]interface{}
		if err := decoder.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode record at %d: %w", count+1, err)
		}

		if err := handler(record); err != nil {
			return fmt.Errorf("handler error at record %d: %w", count+1, err)
		}

		count++
	}

	return nil
}

// CalculateChecksum 计算文件校验和
func (da *DataArchiver) CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file
	if filepath.Ext(filePath) == ".gz" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return "", fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// CompareChecksum 比较校验和
func (da *DataArchiver) CompareChecksum(filePath, expectedChecksum string) (bool, error) {
	actualChecksum, err := da.CalculateChecksum(filePath)
	if err != nil {
		return false, err
	}

	return actualChecksum == expectedChecksum, nil
}

// ArchiveBuffer 归档缓冲区
type ArchiveBuffer struct {
	buffer *bytes.Buffer
	mu     sync.Mutex
}

// NewArchiveBuffer 创建归档缓冲区
func NewArchiveBuffer() *ArchiveBuffer {
	return &ArchiveBuffer{
		buffer: bytes.NewBuffer(nil),
	}
}

// Write 写入数据
func (ab *ArchiveBuffer) Write(data []byte) (int, error) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	return ab.buffer.Write(data)
}

// Bytes 获取数据
func (ab *ArchiveBuffer) Bytes() []byte {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	return ab.buffer.Bytes()
}

// Reset 重置缓冲区
func (ab *ArchiveBuffer) Reset() {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	ab.buffer.Reset()
}

// Len 获取长度
func (ab *ArchiveBuffer) Len() int {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	return ab.buffer.Len()
}
