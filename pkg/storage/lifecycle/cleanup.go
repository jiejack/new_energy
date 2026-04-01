package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CleanupStatus 清理状态
type CleanupStatus string

const (
	CleanupStatusPending   CleanupStatus = "pending"
	CleanupStatusRunning   CleanupStatus = "running"
	CleanupStatusCompleted CleanupStatus = "completed"
	CleanupStatusFailed    CleanupStatus = "failed"
	CleanupStatusCancelled CleanupStatus = "cancelled"
)

// CleanupPolicy 清理策略
type CleanupPolicy struct {
	ID             string        `json:"id" gorm:"primaryKey"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	DataType       string        `json:"data_type"`        // 数据类型/表名
	RetentionDays  int           `json:"retention_days"`   // 保留天数
	BatchSize      int           `json:"batch_size"`       // 批次大小
	Schedule       string        `json:"schedule"`         // 调度表达式 (cron)
	Enabled        bool          `json:"enabled"`
	DryRun         bool          `json:"dry_run"`          // 试运行模式
	ArchiveFirst   bool          `json:"archive_first"`    // 清理前先归档
	NotifyOnComplete bool        `json:"notify_on_complete"` // 完成时通知
	CreatedAt      time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

// CleanupTask 清理任务
type CleanupTask struct {
	ID            string        `json:"id" gorm:"primaryKey"`
	PolicyID      string        `json:"policy_id" gorm:"index"`
	Status        CleanupStatus `json:"status"`
	StartTime     *time.Time    `json:"start_time"`
	EndTime       *time.Time    `json:"end_time"`
	RecordsScanned int64        `json:"records_scanned"`  // 扫描记录数
	RecordsDeleted int64        `json:"records_deleted"`  // 删除记录数
	DataFreed     int64         `json:"data_freed"`       // 释放空间（字节）
	Error         string        `json:"error"`            // 错误信息
	DryRun        bool          `json:"dry_run"`          // 是否试运行
	CreatedAt     time.Time     `json:"created_at" gorm:"autoCreateTime"`
}

// CleanupLog 清理日志
type CleanupLog struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TaskID      string    `json:"task_id" gorm:"index"`
	Level       string    `json:"level"`        // info, warn, error
	Message     string    `json:"message"`
	RecordID    string    `json:"record_id"`    // 操作的记录ID
	Details     string    `json:"details"`      // 详细信息 (JSON)
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// CleanupConfig 清理配置
type CleanupConfig struct {
	DefaultBatchSize   int           `yaml:"default_batch_size" json:"default_batch_size"`     // 默认批次大小
	MaxBatchSize       int           `yaml:"max_batch_size" json:"max_batch_size"`             // 最大批次大小
	DefaultRetention   int           `yaml:"default_retention" json:"default_retention"`       // 默认保留天数
	ParallelWorkers    int           `yaml:"parallel_workers" json:"parallel_workers"`         // 并行工作数
	ScheduleInterval   time.Duration `yaml:"schedule_interval" json:"schedule_interval"`       // 调度检查间隔
	EnableAutoCleanup  bool          `yaml:"enable_auto_cleanup" json:"enable_auto_cleanup"`   // 启用自动清理
	LogRetentionDays   int           `yaml:"log_retention_days" json:"log_retention_days"`     // 日志保留天数
	MaxRetryCount      int           `yaml:"max_retry_count" json:"max_retry_count"`           // 最大重试次数
	RetryDelay         time.Duration `yaml:"retry_delay" json:"retry_delay"`                   // 重试延迟
}

// DefaultCleanupConfig 默认清理配置
func DefaultCleanupConfig() CleanupConfig {
	return CleanupConfig{
		DefaultBatchSize:   1000,
		MaxBatchSize:       10000,
		DefaultRetention:   90,
		ParallelWorkers:    4,
		ScheduleInterval:   1 * time.Hour,
		EnableAutoCleanup:  true,
		LogRetentionDays:   30,
		MaxRetryCount:      3,
		RetryDelay:         5 * time.Second,
	}
}

// DataCleaner 数据清理器
type DataCleaner struct {
	config    CleanupConfig
	db        *gorm.DB
	logger    *zap.Logger

	// 任务队列
	taskQueue chan *CleanupTask
	taskMu    sync.RWMutex

	// 控制通道
	stopCh    chan struct{}
	wg        sync.WaitGroup

	// 指标
	metrics   *CleanupMetrics
}

// CleanupMetrics 清理指标
type CleanupMetrics struct {
	mu                sync.RWMutex
	TotalTasks        int64
	CompletedTasks    int64
	FailedTasks       int64
	TotalRecordsDeleted int64
	TotalDataFreed    int64
	LastCleanupTime   time.Time
	AverageDuration   time.Duration
}

// NewDataCleaner 创建数据清理器
func NewDataCleaner(config CleanupConfig, db *gorm.DB, logger *zap.Logger) *DataCleaner {
	return &DataCleaner{
		config:    config,
		db:        db,
		logger:    logger,
		taskQueue: make(chan *CleanupTask, 1000),
		stopCh:    make(chan struct{}),
		metrics:   &CleanupMetrics{},
	}
}

// Start 启动清理器
func (dc *DataCleaner) Start(ctx context.Context) error {
	dc.logger.Info("Starting data cleaner")

	// 启动工作协程
	for i := 0; i < dc.config.ParallelWorkers; i++ {
		dc.wg.Add(1)
		go dc.cleanupWorker(ctx, i)
	}

	// 启动调度协程
	if dc.config.EnableAutoCleanup {
		dc.wg.Add(1)
		go dc.schedulerWorker(ctx)
	}

	// 启动日志清理协程
	dc.wg.Add(1)
	go dc.logCleanupWorker(ctx)

	// 启动指标更新协程
	dc.wg.Add(1)
	go dc.metricsWorker(ctx)

	dc.logger.Info("Data cleaner started")
	return nil
}

// Stop 停止清理器
func (dc *DataCleaner) Stop() error {
	dc.logger.Info("Stopping data cleaner")
	close(dc.stopCh)
	dc.wg.Wait()
	dc.logger.Info("Data cleaner stopped")
	return nil
}

// CreatePolicy 创建清理策略
func (dc *DataCleaner) CreatePolicy(ctx context.Context, policy *CleanupPolicy) error {
	if policy.ID == "" {
		policy.ID = uuid.New().String()
	}
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	if policy.BatchSize == 0 {
		policy.BatchSize = dc.config.DefaultBatchSize
	}

	if policy.RetentionDays == 0 {
		policy.RetentionDays = dc.config.DefaultRetention
	}

	if err := dc.db.Create(policy).Error; err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	dc.logger.Info("Cleanup policy created",
		zap.String("policy_id", policy.ID),
		zap.String("name", policy.Name),
		zap.String("data_type", policy.DataType),
	)

	return nil
}

// UpdatePolicy 更新清理策略
func (dc *DataCleaner) UpdatePolicy(ctx context.Context, policy *CleanupPolicy) error {
	policy.UpdatedAt = time.Now()

	if err := dc.db.Save(policy).Error; err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	dc.logger.Info("Cleanup policy updated",
		zap.String("policy_id", policy.ID),
	)

	return nil
}

// DeletePolicy 删除清理策略
func (dc *DataCleaner) DeletePolicy(ctx context.Context, policyID string) error {
	if err := dc.db.Delete(&CleanupPolicy{}, "id = ?", policyID).Error; err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	dc.logger.Info("Cleanup policy deleted",
		zap.String("policy_id", policyID),
	)

	return nil
}

// GetPolicy 获取清理策略
func (dc *DataCleaner) GetPolicy(ctx context.Context, policyID string) (*CleanupPolicy, error) {
	var policy CleanupPolicy
	if err := dc.db.First(&policy, "id = ?", policyID).Error; err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}
	return &policy, nil
}

// ListPolicies 列出清理策略
func (dc *DataCleaner) ListPolicies(ctx context.Context) ([]CleanupPolicy, error) {
	var policies []CleanupPolicy
	if err := dc.db.Find(&policies).Error; err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}
	return policies, nil
}

// TriggerCleanup 触发清理
func (dc *DataCleaner) TriggerCleanup(ctx context.Context, policyID string) (*CleanupTask, error) {
	// 获取策略
	policy, err := dc.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, err
	}

	if !policy.Enabled {
		return nil, fmt.Errorf("policy %s is disabled", policyID)
	}

	// 创建任务
	task := &CleanupTask{
		ID:        uuid.New().String(),
		PolicyID:  policyID,
		Status:    CleanupStatusPending,
		DryRun:    policy.DryRun,
		CreatedAt: time.Now(),
	}

	if err := dc.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// 加入队列
	dc.taskQueue <- task

	dc.logger.Info("Cleanup task triggered",
		zap.String("task_id", task.ID),
		zap.String("policy_id", policyID),
		zap.Bool("dry_run", task.DryRun),
	)

	return task, nil
}

// GetTask 获取任务状态
func (dc *DataCleaner) GetTask(ctx context.Context, taskID string) (*CleanupTask, error) {
	var task CleanupTask
	if err := dc.db.First(&task, "id = ?", taskID).Error; err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

// ListTasks 列出任务
func (dc *DataCleaner) ListTasks(ctx context.Context, policyID string, limit int) ([]CleanupTask, error) {
	var tasks []CleanupTask
	query := dc.db.Model(&CleanupTask{}).Order("created_at DESC")

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

// CancelTask 取消任务
func (dc *DataCleaner) CancelTask(ctx context.Context, taskID string) error {
	task, err := dc.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	if task.Status != CleanupStatusPending && task.Status != CleanupStatusRunning {
		return fmt.Errorf("task %s cannot be cancelled in status %s", taskID, task.Status)
	}

	task.Status = CleanupStatusCancelled
	now := time.Now()
	task.EndTime = &now

	if err := dc.db.Save(task).Error; err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	dc.logger.Info("Cleanup task cancelled",
		zap.String("task_id", taskID),
	)

	return nil
}

// GetCleanupLogs 获取清理日志
func (dc *DataCleaner) GetCleanupLogs(ctx context.Context, taskID string, limit int) ([]CleanupLog, error) {
	var logs []CleanupLog
	query := dc.db.Model(&CleanupLog{}).Where("task_id = ?", taskID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get cleanup logs: %w", err)
	}

	return logs, nil
}

// GetMetrics 获取指标
func (dc *DataCleaner) GetMetrics() CleanupMetrics {
	dc.metrics.mu.RLock()
	defer dc.metrics.mu.RUnlock()
	return *dc.metrics
}

// PreviewCleanup 预览清理
func (dc *DataCleaner) PreviewCleanup(ctx context.Context, policyID string) (int64, error) {
	policy, err := dc.GetPolicy(ctx, policyID)
	if err != nil {
		return 0, err
	}

	cutoffTime := time.Now().AddDate(0, 0, -policy.RetentionDays)

	var count int64
	if err := dc.db.Table(policy.DataType).
		Where("created_at < ?", cutoffTime).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}

	return count, nil
}

// 内部方法

func (dc *DataCleaner) cleanupWorker(ctx context.Context, workerID int) {
	defer dc.wg.Done()

	dc.logger.Debug("Cleanup worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-ctx.Done():
			return
		case <-dc.stopCh:
			return
		case task := <-dc.taskQueue:
			dc.executeCleanupTask(ctx, task)
		}
	}
}

func (dc *DataCleaner) executeCleanupTask(ctx context.Context, task *CleanupTask) {
	dc.logger.Info("Executing cleanup task",
		zap.String("task_id", task.ID),
		zap.String("policy_id", task.PolicyID),
		zap.Bool("dry_run", task.DryRun),
	)

	// 更新任务状态
	now := time.Now()
	task.Status = CleanupStatusRunning
	task.StartTime = &now
	dc.db.Save(task)

	// 获取策略
	policy, err := dc.GetPolicy(ctx, task.PolicyID)
	if err != nil {
		dc.failTask(task, fmt.Sprintf("Failed to get policy: %v", err))
		return
	}

	// 执行清理
	var lastErr error
	for i := 0; i < dc.config.MaxRetryCount; i++ {
		if i > 0 {
			time.Sleep(dc.config.RetryDelay)
		}

		lastErr = dc.doCleanup(ctx, task, policy)
		if lastErr == nil {
			break
		}

		dc.logger.Warn("Cleanup task failed, retrying",
			zap.String("task_id", task.ID),
			zap.Int("attempt", i+1),
			zap.Error(lastErr),
		)
	}

	if lastErr != nil {
		dc.failTask(task, fmt.Sprintf("Cleanup failed after %d retries: %v", dc.config.MaxRetryCount, lastErr))
		return
	}

	// 完成任务
	endTime := time.Now()
	task.Status = CleanupStatusCompleted
	task.EndTime = &endTime
	dc.db.Save(task)

	// 更新指标
	dc.metrics.mu.Lock()
	dc.metrics.CompletedTasks++
	dc.metrics.TotalRecordsDeleted += task.RecordsDeleted
	dc.metrics.TotalDataFreed += task.DataFreed
	dc.metrics.LastCleanupTime = endTime
	if dc.metrics.CompletedTasks > 0 {
		dc.metrics.AverageDuration = time.Duration(
			int64(dc.metrics.AverageDuration)*int64(dc.metrics.CompletedTasks-1)/int64(dc.metrics.CompletedTasks) +
				int64(endTime.Sub(now))/int64(dc.metrics.CompletedTasks),
		)
	}
	dc.metrics.mu.Unlock()

	dc.logger.Info("Cleanup task completed",
		zap.String("task_id", task.ID),
		zap.Int64("records_deleted", task.RecordsDeleted),
		zap.Int64("data_freed", task.DataFreed),
	)
}

func (dc *DataCleaner) doCleanup(ctx context.Context, task *CleanupTask, policy *CleanupPolicy) error {
	cutoffTime := time.Now().AddDate(0, 0, -policy.RetentionDays)

	// 记录日志
	dc.logTask(task.ID, "info", fmt.Sprintf("Starting cleanup for data before %s", cutoffTime.Format(time.RFC3339)), "", "")

	var totalDeleted int64
	var totalScanned int64
	var totalFreed int64

	for {
		// 检查是否取消
		if dc.checkTaskCancelled(task.ID) {
			dc.logTask(task.ID, "warn", "Task cancelled", "", "")
			return fmt.Errorf("task cancelled")
		}

		// 查询需要删除的记录
		var records []map[string]interface{}
		query := dc.db.Table(policy.DataType).
			Select("id, created_at").
			Where("created_at < ?", cutoffTime).
			Order("created_at ASC").
			Limit(policy.BatchSize)

		if err := query.Find(&records).Error; err != nil {
			return fmt.Errorf("failed to query records: %w", err)
		}

		if len(records) == 0 {
			break
		}

		totalScanned += int64(len(records))

		// 提取ID
		ids := make([]string, len(records))
		for i, record := range records {
			ids[i] = fmt.Sprintf("%v", record["id"])
		}

		// 删除记录
		if !task.DryRun {
			result := dc.db.Table(policy.DataType).Where("id IN ?", ids).Delete(nil)
			if result.Error != nil {
				return fmt.Errorf("failed to delete records: %w", result.Error)
			}

			deleted := result.RowsAffected
			totalDeleted += deleted

			// 估算释放空间（简化计算）
			estimatedSize := deleted * 1024 // 假设每条记录约1KB
			totalFreed += estimatedSize

			dc.logTask(task.ID, "info", fmt.Sprintf("Deleted %d records", deleted), "", fmt.Sprintf("IDs: %v", ids[:min(10, len(ids))]))
		} else {
			totalDeleted += int64(len(records))
			dc.logTask(task.ID, "info", fmt.Sprintf("[DRY RUN] Would delete %d records", len(records)), "", "")
		}

		// 更新任务进度
		task.RecordsScanned = totalScanned
		task.RecordsDeleted = totalDeleted
		task.DataFreed = totalFreed
		dc.db.Save(task)
	}

	dc.logTask(task.ID, "info", fmt.Sprintf("Cleanup completed: %d records scanned, %d deleted", totalScanned, totalDeleted), "", "")

	return nil
}

func (dc *DataCleaner) failTask(task *CleanupTask, errMsg string) {
	endTime := time.Now()
	task.Status = CleanupStatusFailed
	task.Error = errMsg
	task.EndTime = &endTime
	dc.db.Save(task)

	dc.metrics.mu.Lock()
	dc.metrics.FailedTasks++
	dc.metrics.mu.Unlock()

	dc.logTask(task.ID, "error", errMsg, "", "")

	dc.logger.Error("Cleanup task failed",
		zap.String("task_id", task.ID),
		zap.String("error", errMsg),
	)
}

func (dc *DataCleaner) checkTaskCancelled(taskID string) bool {
	var task CleanupTask
	if err := dc.db.Select("status").First(&task, "id = ?", taskID).Error; err != nil {
		return false
	}
	return task.Status == CleanupStatusCancelled
}

func (dc *DataCleaner) logTask(taskID, level, message, recordID, details string) {
	log := CleanupLog{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		Level:     level,
		Message:   message,
		RecordID:  recordID,
		Details:   details,
		CreatedAt: time.Now(),
	}

	dc.db.Create(&log)
}

func (dc *DataCleaner) schedulerWorker(ctx context.Context) {
	defer dc.wg.Done()

	ticker := time.NewTicker(dc.config.ScheduleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-dc.stopCh:
			return
		case <-ticker.C:
			dc.checkScheduledCleanups(ctx)
		}
	}
}

func (dc *DataCleaner) checkScheduledCleanups(ctx context.Context) {
	// 获取所有启用的策略
	var policies []CleanupPolicy
	if err := dc.db.Where("enabled = ?", true).Find(&policies).Error; err != nil {
		dc.logger.Error("Failed to get enabled policies", zap.Error(err))
		return
	}

	for _, policy := range policies {
		// 检查是否有正在运行的任务
		var runningCount int64
		dc.db.Model(&CleanupTask{}).
			Where("policy_id = ? AND status IN ?", policy.ID, []CleanupStatus{CleanupStatusPending, CleanupStatusRunning}).
			Count(&runningCount)

		if runningCount > 0 {
			continue
		}

		// 检查是否需要执行清理
		cutoffTime := time.Now().AddDate(0, 0, -policy.RetentionDays)

		var count int64
		dc.db.Table(policy.DataType).Where("created_at < ?", cutoffTime).Count(&count)

		if count > 0 {
			dc.logger.Info("Found data to cleanup",
				zap.String("policy_id", policy.ID),
				zap.String("data_type", policy.DataType),
				zap.Int64("count", count),
			)

			// 触发清理
			dc.TriggerCleanup(ctx, policy.ID)
		}
	}
}

func (dc *DataCleaner) logCleanupWorker(ctx context.Context) {
	defer dc.wg.Done()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-dc.stopCh:
			return
		case <-ticker.C:
			dc.cleanupOldLogs(ctx)
		}
	}
}

func (dc *DataCleaner) cleanupOldLogs(ctx context.Context) {
	cutoffTime := time.Now().AddDate(0, 0, -dc.config.LogRetentionDays)

	result := dc.db.Where("created_at < ?", cutoffTime).Delete(&CleanupLog{})
	if result.Error != nil {
		dc.logger.Error("Failed to cleanup old logs", zap.Error(result.Error))
		return
	}

	if result.RowsAffected > 0 {
		dc.logger.Info("Old cleanup logs deleted",
			zap.Int64("count", result.RowsAffected),
		)
	}
}

func (dc *DataCleaner) metricsWorker(ctx context.Context) {
	defer dc.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-dc.stopCh:
			return
		case <-ticker.C:
			dc.updateMetrics(ctx)
		}
	}
}

func (dc *DataCleaner) updateMetrics(ctx context.Context) {
	var totalTasks, completedTasks, failedTasks int64

	dc.db.Model(&CleanupTask{}).Count(&totalTasks)
	dc.db.Model(&CleanupTask{}).Where("status = ?", CleanupStatusCompleted).Count(&completedTasks)
	dc.db.Model(&CleanupTask{}).Where("status = ?", CleanupStatusFailed).Count(&failedTasks)

	dc.metrics.mu.Lock()
	dc.metrics.TotalTasks = totalTasks
	dc.metrics.CompletedTasks = completedTasks
	dc.metrics.FailedTasks = failedTasks
	dc.metrics.mu.Unlock()
}

// CleanupByQuery 按查询条件清理
func (dc *DataCleaner) CleanupByQuery(ctx context.Context, tableName string, whereClause string, args ...interface{}) (int64, error) {
	result := dc.db.Table(tableName).Where(whereClause, args...).Delete(nil)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup by query: %w", result.Error)
	}

	dc.logger.Info("Cleanup by query completed",
		zap.String("table", tableName),
		zap.Int64("deleted", result.RowsAffected),
	)

	return result.RowsAffected, nil
}

// CleanupByDate 按日期清理
func (dc *DataCleaner) CleanupByDate(ctx context.Context, tableName string, dateField string, beforeDate time.Time) (int64, error) {
	result := dc.db.Table(tableName).Where(fmt.Sprintf("%s < ?", dateField), beforeDate).Delete(nil)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup by date: %w", result.Error)
	}

	dc.logger.Info("Cleanup by date completed",
		zap.String("table", tableName),
		zap.String("date_field", dateField),
		zap.Time("before_date", beforeDate),
		zap.Int64("deleted", result.RowsAffected),
	)

	return result.RowsAffected, nil
}

// CleanupBatch 批量清理
func (dc *DataCleaner) CleanupBatch(ctx context.Context, tableName string, ids []string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	result := dc.db.Table(tableName).Where("id IN ?", ids).Delete(nil)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup batch: %w", result.Error)
	}

	dc.logger.Info("Batch cleanup completed",
		zap.String("table", tableName),
		zap.Int("requested", len(ids)),
		zap.Int64("deleted", result.RowsAffected),
	)

	return result.RowsAffected, nil
}

// GetStorageStats 获取存储统计
func (dc *DataCleaner) GetStorageStats(ctx context.Context, tableName string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalCount int64
	dc.db.Table(tableName).Count(&totalCount)

	var oldestRecord, newestRecord time.Time
	dc.db.Table(tableName).Select("MIN(created_at)").Scan(&oldestRecord)
	dc.db.Table(tableName).Select("MAX(created_at)").Scan(&newestRecord)

	stats["total_count"] = totalCount
	stats["oldest_record"] = oldestRecord
	stats["newest_record"] = newestRecord

	// 计算各时间段的数据量
	now := time.Now()
	timeRanges := map[string]time.Duration{
		"last_24h":   24 * time.Hour,
		"last_7d":    7 * 24 * time.Hour,
		"last_30d":   30 * 24 * time.Hour,
		"last_90d":   90 * 24 * time.Hour,
		"last_365d":  365 * 24 * time.Hour,
	}

	for name, duration := range timeRanges {
		var count int64
		dc.db.Table(tableName).
			Where("created_at > ?", now.Add(-duration)).
			Count(&count)
		stats[name] = count
	}

	return stats, nil
}

// EstimateCleanupSize 估算清理大小
func (dc *DataCleaner) EstimateCleanupSize(ctx context.Context, policyID string) (map[string]interface{}, error) {
	policy, err := dc.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, err
	}

	cutoffTime := time.Now().AddDate(0, 0, -policy.RetentionDays)

	var count int64
	dc.db.Table(policy.DataType).Where("created_at < ?", cutoffTime).Count(&count)

	// 估算大小（假设每条记录平均1KB）
	estimatedSize := count * 1024

	result := map[string]interface{}{
		"record_count":   count,
		"estimated_size": estimatedSize,
		"size_mb":        estimatedSize / (1024 * 1024),
		"cutoff_time":    cutoffTime,
	}

	return result, nil
}

// CleanupOrphanedRecords 清理孤立记录
func (dc *DataCleaner) CleanupOrphanedRecords(ctx context.Context, childTable, parentTable, foreignKey string) (int64, error) {
	// 删除子表中没有对应父记录的记录
	query := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE %s NOT IN (SELECT id FROM %s)
	`, childTable, foreignKey, parentTable)

	result := dc.db.Exec(query)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup orphaned records: %w", result.Error)
	}

	dc.logger.Info("Orphaned records cleanup completed",
		zap.String("child_table", childTable),
		zap.String("parent_table", parentTable),
		zap.Int64("deleted", result.RowsAffected),
	)

	return result.RowsAffected, nil
}

// CleanupDuplicates 清理重复记录
func (dc *DataCleaner) CleanupDuplicates(ctx context.Context, tableName string, uniqueFields []string, keepOldest bool) (int64, error) {
	if len(uniqueFields) == 0 {
		return 0, fmt.Errorf("unique fields cannot be empty")
	}

	// 构建去重查询
	fieldList := ""
	for i, field := range uniqueFields {
		if i > 0 {
			fieldList += ", "
		}
		fieldList += field
	}

	orderClause := "created_at ASC"
	if !keepOldest {
		orderClause = "created_at DESC"
	}

	query := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE id NOT IN (
			SELECT id FROM (
				SELECT DISTINCT ON (%s) id, created_at
				FROM %s
				ORDER BY %s, %s
			) AS keep
		)
	`, tableName, fieldList, tableName, fieldList, orderClause)

	result := dc.db.Exec(query)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup duplicates: %w", result.Error)
	}

	dc.logger.Info("Duplicates cleanup completed",
		zap.String("table", tableName),
		zap.Strings("unique_fields", uniqueFields),
		zap.Bool("keep_oldest", keepOldest),
		zap.Int64("deleted", result.RowsAffected),
	)

	return result.RowsAffected, nil
}

// VacuumTable 清理表空间
func (dc *DataCleaner) VacuumTable(ctx context.Context, tableName string) error {
	// PostgreSQL VACUUM
	result := dc.db.Exec(fmt.Sprintf("VACUUM ANALYZE %s", tableName))
	if result.Error != nil {
		return fmt.Errorf("failed to vacuum table: %w", result.Error)
	}

	dc.logger.Info("Table vacuumed",
		zap.String("table", tableName),
	)

	return nil
}

// ReindexTable 重建索引
func (dc *DataCleaner) ReindexTable(ctx context.Context, tableName string) error {
	result := dc.db.Exec(fmt.Sprintf("REINDEX TABLE %s", tableName))
	if result.Error != nil {
		return fmt.Errorf("failed to reindex table: %w", result.Error)
	}

	dc.logger.Info("Table reindexed",
		zap.String("table", tableName),
	)

	return nil
}

// GetTableSize 获取表大小
func (dc *DataCleaner) GetTableSize(ctx context.Context, tableName string) (map[string]interface{}, error) {
	var result struct {
		TotalSize string
		TableSize string
		IndexSize string
	}

	err := dc.db.Raw(`
		SELECT 
			pg_size_pretty(pg_total_relation_size(?)) as total_size,
			pg_size_pretty(pg_relation_size(?)) as table_size,
			pg_size_pretty(pg_indexes_size(?)) as index_size
	`, tableName, tableName, tableName).Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get table size: %w", err)
	}

	return map[string]interface{}{
		"total_size": result.TotalSize,
		"table_size": result.TableSize,
		"index_size": result.IndexSize,
	}, nil
}

// AnalyzeTable 分析表统计信息
func (dc *DataCleaner) AnalyzeTable(ctx context.Context, tableName string) error {
	result := dc.db.Exec(fmt.Sprintf("ANALYZE %s", tableName))
	if result.Error != nil {
		return fmt.Errorf("failed to analyze table: %w", result.Error)
	}

	dc.logger.Info("Table analyzed",
		zap.String("table", tableName),
	)

	return nil
}

// 辅助函数

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
