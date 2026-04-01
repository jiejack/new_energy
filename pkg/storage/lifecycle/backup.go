package lifecycle

import (
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

// BackupType 备份类型
type BackupType string

const (
	BackupTypeFull      BackupType = "full"      // 全量备份
	BackupTypeIncrement BackupType = "increment" // 增量备份
	BackupTypeDifferential BackupType = "differential" // 差异备份
)

// BackupStatus 备份状态
type BackupStatus string

const (
	BackupStatusPending    BackupStatus = "pending"
	BackupStatusRunning    BackupStatus = "running"
	BackupStatusCompleted  BackupStatus = "completed"
	BackupStatusFailed     BackupStatus = "failed"
	BackupStatusValidating BackupStatus = "validating"
	BackupStatusRestoring  BackupStatus = "restoring"
)

// BackupPolicy 备份策略
type BackupPolicy struct {
	ID              string        `json:"id" gorm:"primaryKey"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Type            BackupType    `json:"type"`
	Schedule        string        `json:"schedule"`         // 调度表达式 (cron)
	RetentionDays   int           `json:"retention_days"`   // 保留天数
	MaxBackups      int           `json:"max_backups"`      // 最大备份数
	Compression     bool          `json:"compression"`      // 是否压缩
	Encryption      bool          `json:"encryption"`       // 是否加密
	EncryptionKey   string        `json:"encryption_key"`   // 加密密钥
	Destination     string        `json:"destination"`      // 目标路径
	Tables          []string      `json:"tables" gorm:"type:text;serializer:json"` // 备份的表
	Enabled         bool          `json:"enabled"`
	CreatedAt       time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

// BackupRecord 备份记录
type BackupRecord struct {
	ID              string        `json:"id" gorm:"primaryKey"`
	PolicyID        string        `json:"policy_id" gorm:"index"`
	Type            BackupType    `json:"type"`
	Status          BackupStatus  `json:"status"`
	StartTime       *time.Time    `json:"start_time"`
	EndTime         *time.Time    `json:"end_time"`
	Size            int64         `json:"size"`             // 字节
	CompressedSize  int64         `json:"compressed_size"`  // 压缩后大小
	FilePath        string        `json:"file_path"`        // 备份文件路径
	Checksum        string        `json:"checksum"`         // 校验和
	TablesCount     int           `json:"tables_count"`     // 表数量
	RecordsCount    int64         `json:"records_count"`    // 记录数
	BaseBackupID    string        `json:"base_backup_id"`   // 基础备份ID（增量备份）
	Error           string        `json:"error"`            // 错误信息
	Metadata        string        `json:"metadata"`         // 元数据 (JSON)
	CreatedAt       time.Time     `json:"created_at" gorm:"autoCreateTime"`
}

// BackupTableRecord 备份表记录
type BackupTableRecord struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	BackupID     string    `json:"backup_id" gorm:"index"`
	TableName    string    `json:"table_name"`
	RecordsCount int64     `json:"records_count"`
	Size         int64     `json:"size"`
	Checksum     string    `json:"checksum"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// RestoreRecord 恢复记录
type RestoreRecord struct {
	ID           string        `json:"id" gorm:"primaryKey"`
	BackupID     string        `json:"backup_id" gorm:"index"`
	Status       BackupStatus  `json:"status"`
	StartTime    *time.Time    `json:"start_time"`
	EndTime      *time.Time    `json:"end_time"`
	TablesRestored int         `json:"tables_restored"`
	RecordsRestored int64      `json:"records_restored"`
	Error        string        `json:"error"`
	CreatedAt    time.Time     `json:"created_at" gorm:"autoCreateTime"`
}

// BackupConfig 备份配置
type BackupConfig struct {
	StoragePath       string        `yaml:"storage_path" json:"storage_path"`           // 备份存储路径
	TempPath          string        `yaml:"temp_path" json:"temp_path"`                 // 临时文件路径
	DefaultRetention  int           `yaml:"default_retention" json:"default_retention"` // 默认保留天数
	MaxBackups        int           `yaml:"max_backups" json:"max_backups"`             // 最大备份数
	CompressionLevel  int           `yaml:"compression_level" json:"compression_level"` // 压缩级别
	EnableEncryption  bool          `yaml:"enable_encryption" json:"enable_encryption"` // 启用加密
	EncryptionKey     string        `yaml:"encryption_key" json:"encryption_key"`       // 加密密钥
	ParallelWorkers   int           `yaml:"parallel_workers" json:"parallel_workers"`   // 并行工作数
	BatchSize         int           `yaml:"batch_size" json:"batch_size"`               // 批次大小
	EnableVerify      bool          `yaml:"enable_verify" json:"enable_verify"`         // 启用验证
	RetryCount        int           `yaml:"retry_count" json:"retry_count"`             // 重试次数
	RetryDelay        time.Duration `yaml:"retry_delay" json:"retry_delay"`             // 重试延迟
	ScheduleInterval  time.Duration `yaml:"schedule_interval" json:"schedule_interval"` // 调度检查间隔
}

// DefaultBackupConfig 默认备份配置
func DefaultBackupConfig() BackupConfig {
	return BackupConfig{
		StoragePath:      "/data/backup",
		TempPath:         "/tmp/backup",
		DefaultRetention: 30,
		MaxBackups:       10,
		CompressionLevel: gzip.DefaultCompression,
		EnableEncryption: false,
		ParallelWorkers:  4,
		BatchSize:        10000,
		EnableVerify:     true,
		RetryCount:       3,
		RetryDelay:       5 * time.Second,
		ScheduleInterval: 1 * time.Hour,
	}
}

// BackupManager 备份管理器
type BackupManager struct {
	config    BackupConfig
	db        *gorm.DB
	logger    *zap.Logger

	// 任务队列
	backupQueue chan *BackupRecord
	restoreQueue chan *RestoreRecord
	taskMu      sync.RWMutex

	// 控制通道
	stopCh      chan struct{}
	wg          sync.WaitGroup

	// 指标
	metrics     *BackupMetrics
}

// BackupMetrics 备份指标
type BackupMetrics struct {
	mu                sync.RWMutex
	TotalBackups      int64
	CompletedBackups  int64
	FailedBackups     int64
	TotalSize         int64
	CompressedSize    int64
	TotalRecords      int64
	LastBackupTime    time.Time
	AverageDuration   time.Duration
	CompressionRatio  float64
}

// NewBackupManager 创建备份管理器
func NewBackupManager(config BackupConfig, db *gorm.DB, logger *zap.Logger) *BackupManager {
	return &BackupManager{
		config:      config,
		db:          db,
		logger:      logger,
		backupQueue: make(chan *BackupRecord, 1000),
		restoreQueue: make(chan *RestoreRecord, 1000),
		stopCh:      make(chan struct{}),
		metrics:     &BackupMetrics{},
	}
}

// Start 启动备份管理器
func (bm *BackupManager) Start(ctx context.Context) error {
	bm.logger.Info("Starting backup manager")

	// 确保目录存在
	if err := os.MkdirAll(bm.config.StoragePath, 0755); err != nil {
		return fmt.Errorf("failed to create storage path: %w", err)
	}
	if err := os.MkdirAll(bm.config.TempPath, 0755); err != nil {
		return fmt.Errorf("failed to create temp path: %w", err)
	}

	// 启动备份工作协程
	for i := 0; i < bm.config.ParallelWorkers; i++ {
		bm.wg.Add(1)
		go bm.backupWorker(ctx, i)
	}

	// 启动恢复工作协程
	bm.wg.Add(1)
	go bm.restoreWorker(ctx)

	// 启动调度协程
	bm.wg.Add(1)
	go bm.schedulerWorker(ctx)

	// 启动清理协程
	bm.wg.Add(1)
	go bm.cleanupWorker(ctx)

	// 启动指标更新协程
	bm.wg.Add(1)
	go bm.metricsWorker(ctx)

	bm.logger.Info("Backup manager started")
	return nil
}

// Stop 停止备份管理器
func (bm *BackupManager) Stop() error {
	bm.logger.Info("Stopping backup manager")
	close(bm.stopCh)
	bm.wg.Wait()
	bm.logger.Info("Backup manager stopped")
	return nil
}

// CreatePolicy 创建备份策略
func (bm *BackupManager) CreatePolicy(ctx context.Context, policy *BackupPolicy) error {
	if policy.ID == "" {
		policy.ID = uuid.New().String()
	}
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	if policy.RetentionDays == 0 {
		policy.RetentionDays = bm.config.DefaultRetention
	}

	if policy.MaxBackups == 0 {
		policy.MaxBackups = bm.config.MaxBackups
	}

	if err := bm.db.Create(policy).Error; err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	bm.logger.Info("Backup policy created",
		zap.String("policy_id", policy.ID),
		zap.String("name", policy.Name),
		zap.String("type", string(policy.Type)),
	)

	return nil
}

// UpdatePolicy 更新备份策略
func (bm *BackupManager) UpdatePolicy(ctx context.Context, policy *BackupPolicy) error {
	policy.UpdatedAt = time.Now()

	if err := bm.db.Save(policy).Error; err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	bm.logger.Info("Backup policy updated",
		zap.String("policy_id", policy.ID),
	)

	return nil
}

// DeletePolicy 删除备份策略
func (bm *BackupManager) DeletePolicy(ctx context.Context, policyID string) error {
	if err := bm.db.Delete(&BackupPolicy{}, "id = ?", policyID).Error; err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	bm.logger.Info("Backup policy deleted",
		zap.String("policy_id", policyID),
	)

	return nil
}

// GetPolicy 获取备份策略
func (bm *BackupManager) GetPolicy(ctx context.Context, policyID string) (*BackupPolicy, error) {
	var policy BackupPolicy
	if err := bm.db.First(&policy, "id = ?", policyID).Error; err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}
	return &policy, nil
}

// ListPolicies 列出备份策略
func (bm *BackupManager) ListPolicies(ctx context.Context) ([]BackupPolicy, error) {
	var policies []BackupPolicy
	if err := bm.db.Find(&policies).Error; err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}
	return policies, nil
}

// TriggerBackup 触发备份
func (bm *BackupManager) TriggerBackup(ctx context.Context, policyID string) (*BackupRecord, error) {
	// 获取策略
	policy, err := bm.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, err
	}

	if !policy.Enabled {
		return nil, fmt.Errorf("policy %s is disabled", policyID)
	}

	// 创建备份记录
	record := &BackupRecord{
		ID:        uuid.New().String(),
		PolicyID:  policyID,
		Type:      policy.Type,
		Status:    BackupStatusPending,
		CreatedAt: time.Now(),
	}

	// 如果是增量备份，查找基础备份
	if policy.Type == BackupTypeIncrement {
		var lastBackup BackupRecord
		if err := bm.db.Where("policy_id = ? AND status = ?", policyID, BackupStatusCompleted).
			Order("created_at DESC").
			First(&lastBackup).Error; err == nil {
			record.BaseBackupID = lastBackup.ID
		} else {
			// 没有基础备份，转为全量备份
			record.Type = BackupTypeFull
		}
	}

	if err := bm.db.Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}

	// 加入队列
	bm.backupQueue <- record

	bm.logger.Info("Backup triggered",
		zap.String("backup_id", record.ID),
		zap.String("policy_id", policyID),
		zap.String("type", string(record.Type)),
	)

	return record, nil
}

// CreateFullBackup 创建全量备份
func (bm *BackupManager) CreateFullBackup(ctx context.Context, tables []string, description string) (*BackupRecord, error) {
	policy := &BackupPolicy{
		ID:          uuid.New().String(),
		Name:        "Manual Full Backup",
		Description: description,
		Type:        BackupTypeFull,
		Tables:      tables,
		Compression: true,
		Enabled:     true,
	}

	if err := bm.CreatePolicy(ctx, policy); err != nil {
		return nil, err
	}

	return bm.TriggerBackup(ctx, policy.ID)
}

// CreateIncrementBackup 创建增量备份
func (bm *BackupManager) CreateIncrementBackup(ctx context.Context, policyID string) (*BackupRecord, error) {
	policy, err := bm.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, err
	}

	if policy.Type != BackupTypeIncrement {
		return nil, fmt.Errorf("policy %s is not an incremental backup policy", policyID)
	}

	return bm.TriggerBackup(ctx, policyID)
}

// GetBackup 获取备份记录
func (bm *BackupManager) GetBackup(ctx context.Context, backupID string) (*BackupRecord, error) {
	var record BackupRecord
	if err := bm.db.First(&record, "id = ?", backupID).Error; err != nil {
		return nil, fmt.Errorf("failed to get backup: %w", err)
	}
	return &record, nil
}

// ListBackups 列出备份记录
func (bm *BackupManager) ListBackups(ctx context.Context, policyID string, limit int) ([]BackupRecord, error) {
	var records []BackupRecord
	query := bm.db.Model(&BackupRecord{}).Order("created_at DESC")

	if policyID != "" {
		query = query.Where("policy_id = ?", policyID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	return records, nil
}

// RestoreBackup 恢复备份
func (bm *BackupManager) RestoreBackup(ctx context.Context, backupID string, tables []string) (*RestoreRecord, error) {
	backup, err := bm.GetBackup(ctx, backupID)
	if err != nil {
		return nil, err
	}

	if backup.Status != BackupStatusCompleted {
		return nil, fmt.Errorf("backup %s is not completed", backupID)
	}

	// 创建恢复记录
	record := &RestoreRecord{
		ID:        uuid.New().String(),
		BackupID:  backupID,
		Status:    BackupStatusPending,
		CreatedAt: time.Now(),
	}

	if err := bm.db.Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to create restore record: %w", err)
	}

	// 加入队列
	bm.restoreQueue <- record

	bm.logger.Info("Restore triggered",
		zap.String("restore_id", record.ID),
		zap.String("backup_id", backupID),
	)

	return record, nil
}

// GetRestore 获取恢复记录
func (bm *BackupManager) GetRestore(ctx context.Context, restoreID string) (*RestoreRecord, error) {
	var record RestoreRecord
	if err := bm.db.First(&record, "id = ?", restoreID).Error; err != nil {
		return nil, fmt.Errorf("failed to get restore: %w", err)
	}
	return &record, nil
}

// ListRestores 列出恢复记录
func (bm *BackupManager) ListRestores(ctx context.Context, backupID string, limit int) ([]RestoreRecord, error) {
	var records []RestoreRecord
	query := bm.db.Model(&RestoreRecord{}).Order("created_at DESC")

	if backupID != "" {
		query = query.Where("backup_id = ?", backupID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to list restores: %w", err)
	}

	return records, nil
}

// DeleteBackup 删除备份
func (bm *BackupManager) DeleteBackup(ctx context.Context, backupID string) error {
	backup, err := bm.GetBackup(ctx, backupID)
	if err != nil {
		return err
	}

	// 删除文件
	if backup.FilePath != "" {
		if err := os.Remove(backup.FilePath); err != nil && !os.IsNotExist(err) {
			bm.logger.Warn("Failed to delete backup file",
				zap.String("file", backup.FilePath),
				zap.Error(err),
			)
		}
	}

	// 删除记录
	if err := bm.db.Delete(&BackupRecord{}, "id = ?", backupID).Error; err != nil {
		return fmt.Errorf("failed to delete backup record: %w", err)
	}

	// 删除表记录
	bm.db.Delete(&BackupTableRecord{}, "backup_id = ?", backupID)

	bm.logger.Info("Backup deleted",
		zap.String("backup_id", backupID),
	)

	return nil
}

// VerifyBackup 验证备份
func (bm *BackupManager) VerifyBackup(ctx context.Context, backupID string) error {
	backup, err := bm.GetBackup(ctx, backupID)
	if err != nil {
		return err
	}

	if backup.Status != BackupStatusCompleted {
		return fmt.Errorf("backup %s is not completed", backupID)
	}

	bm.logger.Info("Verifying backup",
		zap.String("backup_id", backupID),
	)

	// 检查文件是否存在
	if _, err := os.Stat(backup.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backup.FilePath)
	}

	// 验证校验和
	file, err := os.Open(backup.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file
	if filepath.Ext(backup.FilePath) == ".gz" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	checksum := fmt.Sprintf("%x", hash.Sum(nil))
	if checksum != backup.Checksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", backup.Checksum, checksum)
	}

	bm.logger.Info("Backup verified",
		zap.String("backup_id", backupID),
		zap.String("checksum", checksum),
	)

	return nil
}

// GetMetrics 获取指标
func (bm *BackupManager) GetMetrics() BackupMetrics {
	bm.metrics.mu.RLock()
	defer bm.metrics.mu.RUnlock()
	return *bm.metrics
}

// 内部方法

func (bm *BackupManager) backupWorker(ctx context.Context, workerID int) {
	defer bm.wg.Done()

	bm.logger.Debug("Backup worker started", zap.Int("worker_id", workerID))

	for {
		select {
		case <-ctx.Done():
			return
		case <-bm.stopCh:
			return
		case record := <-bm.backupQueue:
			bm.executeBackup(ctx, record)
		}
	}
}

func (bm *BackupManager) executeBackup(ctx context.Context, record *BackupRecord) {
	bm.logger.Info("Executing backup",
		zap.String("backup_id", record.ID),
		zap.String("type", string(record.Type)),
	)

	// 更新状态
	now := time.Now()
	record.Status = BackupStatusRunning
	record.StartTime = &now
	bm.db.Save(record)

	// 获取策略
	policy, err := bm.GetPolicy(ctx, record.PolicyID)
	if err != nil {
		bm.failBackup(record, fmt.Sprintf("Failed to get policy: %v", err))
		return
	}

	// 执行备份
	var lastErr error
	for i := 0; i < bm.config.RetryCount; i++ {
		if i > 0 {
			time.Sleep(bm.config.RetryDelay)
		}

		lastErr = bm.doBackup(ctx, record, policy)
		if lastErr == nil {
			break
		}

		bm.logger.Warn("Backup failed, retrying",
			zap.String("backup_id", record.ID),
			zap.Int("attempt", i+1),
			zap.Error(lastErr),
		)
	}

	if lastErr != nil {
		bm.failBackup(record, fmt.Sprintf("Backup failed after %d retries: %v", bm.config.RetryCount, lastErr))
		return
	}

	// 验证备份
	if bm.config.EnableVerify {
		record.Status = BackupStatusValidating
		bm.db.Save(record)

		if err := bm.VerifyBackup(ctx, record.ID); err != nil {
			bm.failBackup(record, fmt.Sprintf("Backup verification failed: %v", err))
			return
		}
	}

	// 完成
	endTime := time.Now()
	record.Status = BackupStatusCompleted
	record.EndTime = &endTime
	bm.db.Save(record)

	// 更新指标
	bm.metrics.mu.Lock()
	bm.metrics.CompletedBackups++
	bm.metrics.TotalSize += record.Size
	bm.metrics.CompressedSize += record.CompressedSize
	bm.metrics.TotalRecords += record.RecordsCount
	bm.metrics.LastBackupTime = endTime
	if record.Size > 0 {
		bm.metrics.CompressionRatio = float64(record.CompressedSize) / float64(record.Size)
	}
	bm.metrics.mu.Unlock()

	bm.logger.Info("Backup completed",
		zap.String("backup_id", record.ID),
		zap.Int64("size", record.Size),
		zap.Int64("compressed_size", record.CompressedSize),
		zap.Int64("records", record.RecordsCount),
	)
}

func (bm *BackupManager) doBackup(ctx context.Context, record *BackupRecord, policy *BackupPolicy) error {
	// 确定备份的表
	tables := policy.Tables
	if len(tables) == 0 {
		// 获取所有表
		var tableNames []string
		if err := bm.db.Raw(`
			SELECT table_name FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		`).Scan(&tableNames).Error; err != nil {
			return fmt.Errorf("failed to get tables: %w", err)
		}
		tables = tableNames
	}

	record.TablesCount = len(tables)

	// 创建备份文件
	fileName := fmt.Sprintf("backup_%s_%s.json", record.Type, time.Now().Format("20060102_150405"))
	if policy.Compression {
		fileName += ".gz"
	}
	filePath := filepath.Join(bm.config.StoragePath, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	var writer io.Writer = file
	hashWriter := sha256.New()

	if policy.Compression {
		gzWriter := gzip.NewWriter(file)
		gzWriter.Level = bm.config.CompressionLevel
		defer gzWriter.Close()
		writer = io.MultiWriter(gzWriter, hashWriter)
	} else {
		writer = io.MultiWriter(file, hashWriter)
	}

	// 写入备份元数据
	metadata := map[string]interface{}{
		"backup_id":   record.ID,
		"type":        record.Type,
		"created_at":  time.Now(),
		"tables":      tables,
		"base_backup": record.BaseBackupID,
	}

	encoder := json.NewEncoder(writer)
	if err := encoder.Encode(map[string]interface{}{"metadata": metadata}); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// 备份每个表
	var totalRecords int64
	var totalSize int64

	for _, table := range tables {
		records, size, err := bm.backupTable(ctx, table, writer, record)
		if err != nil {
			return fmt.Errorf("failed to backup table %s: %w", table, err)
		}

		totalRecords += records
		totalSize += size

		// 记录表备份信息
		tableRecord := BackupTableRecord{
			ID:           uuid.New().String(),
			BackupID:     record.ID,
			TableName:    table,
			RecordsCount: records,
			Size:         size,
			CreatedAt:    time.Now(),
		}
		bm.db.Create(&tableRecord)
	}

	// 更新记录
	record.FilePath = filePath
	record.Checksum = fmt.Sprintf("%x", hashWriter.Sum(nil))
	record.RecordsCount = totalRecords
	record.Size = totalSize

	// 获取压缩后文件大小
	if info, err := os.Stat(filePath); err == nil {
		record.CompressedSize = info.Size()
	}

	return nil
}

func (bm *BackupManager) backupTable(ctx context.Context, tableName string, writer io.Writer, record *BackupRecord) (int64, int64, error) {
	// 写入表开始标记
	encoder := json.NewEncoder(writer)
	if err := encoder.Encode(map[string]interface{}{"table_start": tableName}); err != nil {
		return 0, 0, err
	}

	var totalRecords int64
	var totalSize int64

	// 分批读取数据
	offset := 0
	batchSize := bm.config.BatchSize

	for {
		var rows []map[string]interface{}
		if err := bm.db.Table(tableName).
			Offset(offset).
			Limit(batchSize).
			Find(&rows).Error; err != nil {
			return 0, 0, err
		}

		if len(rows) == 0 {
			break
		}

		// 写入行数据
		for _, row := range rows {
			if err := encoder.Encode(map[string]interface{}{"row": row}); err != nil {
				return 0, 0, err
			}

			rowBytes, _ := json.Marshal(row)
			totalSize += int64(len(rowBytes))
			totalRecords++
		}

		offset += len(rows)
	}

	// 写入表结束标记
	if err := encoder.Encode(map[string]interface{}{"table_end": tableName}); err != nil {
		return 0, 0, err
	}

	return totalRecords, totalSize, nil
}

func (bm *BackupManager) failBackup(record *BackupRecord, errMsg string) {
	endTime := time.Now()
	record.Status = BackupStatusFailed
	record.Error = errMsg
	record.EndTime = &endTime
	bm.db.Save(record)

	bm.metrics.mu.Lock()
	bm.metrics.FailedBackups++
	bm.metrics.mu.Unlock()

	bm.logger.Error("Backup failed",
		zap.String("backup_id", record.ID),
		zap.String("error", errMsg),
	)
}

func (bm *BackupManager) restoreWorker(ctx context.Context) {
	defer bm.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-bm.stopCh:
			return
		case record := <-bm.restoreQueue:
			bm.executeRestore(ctx, record)
		}
	}
}

func (bm *BackupManager) executeRestore(ctx context.Context, record *RestoreRecord) {
	bm.logger.Info("Executing restore",
		zap.String("restore_id", record.ID),
		zap.String("backup_id", record.BackupID),
	)

	// 更新状态
	now := time.Now()
	record.Status = BackupStatusRestoring
	record.StartTime = &now
	bm.db.Save(record)

	// 获取备份记录
	backup, err := bm.GetBackup(ctx, record.BackupID)
	if err != nil {
		bm.failRestore(record, fmt.Sprintf("Failed to get backup: %v", err))
		return
	}

	// 执行恢复
	if err := bm.doRestore(ctx, backup, record); err != nil {
		bm.failRestore(record, fmt.Sprintf("Restore failed: %v", err))
		return
	}

	// 完成
	endTime := time.Now()
	record.Status = BackupStatusCompleted
	record.EndTime = &endTime
	bm.db.Save(record)

	bm.logger.Info("Restore completed",
		zap.String("restore_id", record.ID),
		zap.Int("tables_restored", record.TablesRestored),
		zap.Int64("records_restored", record.RecordsRestored),
	)
}

func (bm *BackupManager) doRestore(ctx context.Context, backup *BackupRecord, record *RestoreRecord) error {
	// 打开备份文件
	file, err := os.Open(backup.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file
	if filepath.Ext(backup.FilePath) == ".gz" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	decoder := json.NewDecoder(reader)

	// 读取元数据
	var metadataWrapper map[string]interface{}
	if err := decoder.Decode(&metadataWrapper); err != nil {
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	// 恢复表数据
	var currentTable string
	var batch []map[string]interface{}

	for decoder.More() {
		var item map[string]interface{}
		if err := decoder.Decode(&item); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode item: %w", err)
		}

		// 处理不同类型的标记
		if tableName, ok := item["table_start"].(string); ok {
			currentTable = tableName
			batch = make([]map[string]interface{}, 0, bm.config.BatchSize)
			continue
		}

		if _, ok := item["table_end"]; ok {
			// 插入剩余数据
			if len(batch) > 0 {
				if err := bm.db.Table(currentTable).Create(&batch).Error; err != nil {
					bm.logger.Warn("Failed to insert batch",
						zap.String("table", currentTable),
						zap.Error(err),
					)
				}
				record.RecordsRestored += int64(len(batch))
			}

			record.TablesRestored++
			currentTable = ""
			continue
		}

		if row, ok := item["row"].(map[string]interface{}); ok {
			batch = append(batch, row)

			// 批量插入
			if len(batch) >= bm.config.BatchSize {
				if err := bm.db.Table(currentTable).Create(&batch).Error; err != nil {
					bm.logger.Warn("Failed to insert batch",
						zap.String("table", currentTable),
						zap.Error(err),
					)
				}
				record.RecordsRestored += int64(len(batch))
				batch = batch[:0]
			}
		}
	}

	return nil
}

func (bm *BackupManager) failRestore(record *RestoreRecord, errMsg string) {
	endTime := time.Now()
	record.Status = BackupStatusFailed
	record.Error = errMsg
	record.EndTime = &endTime
	bm.db.Save(record)

	bm.logger.Error("Restore failed",
		zap.String("restore_id", record.ID),
		zap.String("error", errMsg),
	)
}

func (bm *BackupManager) schedulerWorker(ctx context.Context) {
	defer bm.wg.Done()

	ticker := time.NewTicker(bm.config.ScheduleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-bm.stopCh:
			return
		case <-ticker.C:
			bm.checkScheduledBackups(ctx)
		}
	}
}

func (bm *BackupManager) checkScheduledBackups(ctx context.Context) {
	// 获取所有启用的策略
	var policies []BackupPolicy
	if err := bm.db.Where("enabled = ?", true).Find(&policies).Error; err != nil {
		bm.logger.Error("Failed to get enabled policies", zap.Error(err))
		return
	}

	for _, policy := range policies {
		// 检查是否需要执行备份
		// 这里简化处理，实际应该解析 cron 表达式
		var lastBackup BackupRecord
		err := bm.db.Where("policy_id = ? AND status = ?", policy.ID, BackupStatusCompleted).
			Order("created_at DESC").
			First(&lastBackup).Error

		shouldBackup := false
		if err == gorm.ErrRecordNotFound {
			shouldBackup = true
		} else if err == nil {
			// 检查上次备份时间
			// 简化：每天检查一次
			if time.Since(lastBackup.CreatedAt) > 24*time.Hour {
				shouldBackup = true
			}
		}

		if shouldBackup {
			bm.logger.Info("Triggering scheduled backup",
				zap.String("policy_id", policy.ID),
				zap.String("name", policy.Name),
			)

			bm.TriggerBackup(ctx, policy.ID)
		}
	}
}

func (bm *BackupManager) cleanupWorker(ctx context.Context) {
	defer bm.wg.Done()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-bm.stopCh:
			return
		case <-ticker.C:
			bm.cleanupOldBackups(ctx)
		}
	}
}

func (bm *BackupManager) cleanupOldBackups(ctx context.Context) {
	// 获取所有策略
	var policies []BackupPolicy
	if err := bm.db.Find(&policies).Error; err != nil {
		bm.logger.Error("Failed to get policies for cleanup", zap.Error(err))
		return
	}

	for _, policy := range policies {
		// 按保留天数清理
		cutoffTime := time.Now().AddDate(0, 0, -policy.RetentionDays)

		var oldBackups []BackupRecord
		if err := bm.db.Where("policy_id = ? AND created_at < ?", policy.ID, cutoffTime).
			Find(&oldBackups).Error; err != nil {
			continue
		}

		for _, backup := range oldBackups {
			bm.DeleteBackup(ctx, backup.ID)
		}

		// 按最大备份数清理
		var allBackups []BackupRecord
		if err := bm.db.Where("policy_id = ?", policy.ID).
			Order("created_at DESC").
			Find(&allBackups).Error; err != nil {
			continue
		}

		if len(allBackups) > policy.MaxBackups {
			for i := policy.MaxBackups; i < len(allBackups); i++ {
				bm.DeleteBackup(ctx, allBackups[i].ID)
			}
		}
	}
}

func (bm *BackupManager) metricsWorker(ctx context.Context) {
	defer bm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-bm.stopCh:
			return
		case <-ticker.C:
			bm.updateMetrics(ctx)
		}
	}
}

func (bm *BackupManager) updateMetrics(ctx context.Context) {
	var totalBackups, completedBackups, failedBackups int64

	bm.db.Model(&BackupRecord{}).Count(&totalBackups)
	bm.db.Model(&BackupRecord{}).Where("status = ?", BackupStatusCompleted).Count(&completedBackups)
	bm.db.Model(&BackupRecord{}).Where("status = ?", BackupStatusFailed).Count(&failedBackups)

	bm.metrics.mu.Lock()
	bm.metrics.TotalBackups = totalBackups
	bm.metrics.CompletedBackups = completedBackups
	bm.metrics.FailedBackups = failedBackups
	bm.metrics.mu.Unlock()
}

// GetBackupStats 获取备份统计
func (bm *BackupManager) GetBackupStats(ctx context.Context, policyID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalSize, compressedSize int64
	var count int64

	query := bm.db.Model(&BackupRecord{}).Where("status = ?", BackupStatusCompleted)
	if policyID != "" {
		query = query.Where("policy_id = ?", policyID)
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, err
	}

	if err := query.Select("SUM(size)").Scan(&totalSize).Error; err != nil {
		return nil, err
	}

	if err := query.Select("SUM(compressed_size)").Scan(&compressedSize).Error; err != nil {
		return nil, err
	}

	stats["total_backups"] = count
	stats["total_size"] = totalSize
	stats["total_size_mb"] = totalSize / (1024 * 1024)
	stats["compressed_size"] = compressedSize
	stats["compressed_size_mb"] = compressedSize / (1024 * 1024)

	if totalSize > 0 {
		stats["compression_ratio"] = float64(compressedSize) / float64(totalSize)
	}

	return stats, nil
}

// ExportBackup 导出备份到指定路径
func (bm *BackupManager) ExportBackup(ctx context.Context, backupID string, destPath string) error {
	backup, err := bm.GetBackup(ctx, backupID)
	if err != nil {
		return err
	}

	if backup.Status != BackupStatusCompleted {
		return fmt.Errorf("backup %s is not completed", backupID)
	}

	// 复制文件
	srcFile, err := os.Open(backup.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	bm.logger.Info("Backup exported",
		zap.String("backup_id", backupID),
		zap.String("dest", destPath),
	)

	return nil
}

// ImportBackup 导入备份
func (bm *BackupManager) ImportBackup(ctx context.Context, srcPath string, policyID string) (*BackupRecord, error) {
	// 打开源文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// 创建备份记录
	record := &BackupRecord{
		ID:        uuid.New().String(),
		PolicyID:  policyID,
		Type:      BackupTypeFull,
		Status:    BackupStatusCompleted,
		CreatedAt: time.Now(),
	}

	// 复制到存储路径
	fileName := filepath.Base(srcPath)
	destPath := filepath.Join(bm.config.StoragePath, fileName)

	destFile, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	hash := sha256.New()
	writer := io.MultiWriter(destFile, hash)

	if _, err := io.Copy(writer, srcFile); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// 获取文件信息
	info, _ := os.Stat(destPath)

	record.FilePath = destPath
	record.Checksum = fmt.Sprintf("%x", hash.Sum(nil))
	record.CompressedSize = info.Size()

	if err := bm.db.Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}

	bm.logger.Info("Backup imported",
		zap.String("backup_id", record.ID),
		zap.String("source", srcPath),
	)

	return record, nil
}

// GetBackupTables 获取备份包含的表
func (bm *BackupManager) GetBackupTables(ctx context.Context, backupID string) ([]BackupTableRecord, error) {
	var tables []BackupTableRecord
	if err := bm.db.Where("backup_id = ?", backupID).Find(&tables).Error; err != nil {
		return nil, fmt.Errorf("failed to get backup tables: %w", err)
	}
	return tables, nil
}

// CompareBackups 比较两个备份
func (bm *BackupManager) CompareBackups(ctx context.Context, backupID1, backupID2 string) (map[string]interface{}, error) {
	backup1, err := bm.GetBackup(ctx, backupID1)
	if err != nil {
		return nil, err
	}

	backup2, err := bm.GetBackup(ctx, backupID2)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"backup1": map[string]interface{}{
			"id":            backup1.ID,
			"created_at":    backup1.CreatedAt,
			"size":          backup1.Size,
			"records_count": backup1.RecordsCount,
		},
		"backup2": map[string]interface{}{
			"id":            backup2.ID,
			"created_at":    backup2.CreatedAt,
			"size":          backup2.Size,
			"records_count": backup2.RecordsCount,
		},
		"size_diff":      backup2.Size - backup1.Size,
		"records_diff":   backup2.RecordsCount - backup1.RecordsCount,
		"time_diff":      backup2.CreatedAt.Sub(backup1.CreatedAt),
	}

	return result, nil
}

// ScheduleBackup 调度备份
func (bm *BackupManager) ScheduleBackup(ctx context.Context, policyID string, scheduledTime time.Time) error {
	policy, err := bm.GetPolicy(ctx, policyID)
	if err != nil {
		return err
	}

	// 创建一个待执行的备份记录
	record := &BackupRecord{
		ID:        uuid.New().String(),
		PolicyID:  policyID,
		Type:      policy.Type,
		Status:    BackupStatusPending,
		CreatedAt: scheduledTime,
	}

	if err := bm.db.Create(record).Error; err != nil {
		return fmt.Errorf("failed to create scheduled backup: %w", err)
	}

	bm.logger.Info("Backup scheduled",
		zap.String("backup_id", record.ID),
		zap.String("policy_id", policyID),
		zap.Time("scheduled_time", scheduledTime),
	)

	return nil
}

// CancelScheduledBackup 取消调度的备份
func (bm *BackupManager) CancelScheduledBackup(ctx context.Context, backupID string) error {
	record, err := bm.GetBackup(ctx, backupID)
	if err != nil {
		return err
	}

	if record.Status != BackupStatusPending {
		return fmt.Errorf("backup %s is not pending", backupID)
	}

	if err := bm.db.Delete(&BackupRecord{}, "id = ?", backupID).Error; err != nil {
		return fmt.Errorf("failed to cancel scheduled backup: %w", err)
	}

	bm.logger.Info("Scheduled backup cancelled",
		zap.String("backup_id", backupID),
	)

	return nil
}
