package persistence

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

// MigrationManager 数据库迁移管理器
type MigrationManager struct {
	db *Database
}

// NewMigrationManager 创建迁移管理器
func NewMigrationManager(db *Database) *MigrationManager {
	return &MigrationManager{db: db}
}

// Migration 迁移记录
type Migration struct {
	Version   string    `json:"version"`
	AppliedAt time.Time `json:"applied_at"`
}

// MigrationStatus 迁移状态
type MigrationStatus struct {
	Total     int       `json:"total"`
	Applied   int       `json:"applied"`
	Pending   int       `json:"pending"`
	LastApplied *Migration `json:"last_applied,omitempty"`
}

// RunMigrations 执行数据库迁移
func (m *MigrationManager) RunMigrations(ctx context.Context, migrationsFS embed.FS) error {
	// 创建迁移记录表
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// 获取已应用的迁移
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// 读取迁移文件
	files, err := m.readMigrationFiles(migrationsFS)
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// 应用未执行的迁移
	for _, file := range files {
		version := strings.TrimSuffix(file.Name(), ".sql")
		
		// 跳过已应用的迁移
		if _, exists := applied[version]; exists {
			continue
		}

		// 读取迁移文件内容
		content, err := fs.ReadFile(migrationsFS, filepath.Join("migrations", file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		// 验证迁移脚本
		if err := m.validateMigration(string(content)); err != nil {
			return fmt.Errorf("migration validation failed for %s: %w", file.Name(), err)
		}

		// 应用迁移
		if err := m.applyMigration(ctx, version, string(content)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", file.Name(), err)
		}
	}

	return nil
}

// RunMigrationsWithLimit 执行指定数量的迁移
func (m *MigrationManager) RunMigrationsWithLimit(ctx context.Context, migrationsFS embed.FS, limit int) error {
	// 创建迁移记录表
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// 获取已应用的迁移
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// 读取迁移文件
	files, err := m.readMigrationFiles(migrationsFS)
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// 应用未执行的迁移
	appliedCount := 0
	for _, file := range files {
		if appliedCount >= limit {
			break
		}
		
		version := strings.TrimSuffix(file.Name(), ".sql")
		
		// 跳过已应用的迁移
		if _, exists := applied[version]; exists {
			continue
		}

		// 读取迁移文件内容
		content, err := fs.ReadFile(migrationsFS, filepath.Join("migrations", file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		// 验证迁移脚本
		if err := m.validateMigration(string(content)); err != nil {
			return fmt.Errorf("migration validation failed for %s: %w", file.Name(), err)
		}

		// 应用迁移
		if err := m.applyMigration(ctx, version, string(content)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", file.Name(), err)
		}
		
		appliedCount++
	}

	return nil
}

// createMigrationsTable 创建迁移记录表
func (m *MigrationManager) createMigrationsTable(ctx context.Context) error {
	sql := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT NOW()
	)
	`
	return m.db.WithContext(ctx).Exec(sql).Error
}

// getAppliedMigrations 获取已应用的迁移
func (m *MigrationManager) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	var migrations []struct {
		Version string
	}

	if err := m.db.WithContext(ctx).Table("schema_migrations").Select("version").Find(&migrations).Error; err != nil {
		return nil, err
	}

	applied := make(map[string]bool)
	for _, m := range migrations {
		applied[m.Version] = true
	}

	return applied, nil
}

// readMigrationFiles 读取迁移文件
func (m *MigrationManager) readMigrationFiles(migrationsFS embed.FS) ([]fs.DirEntry, error) {
	files, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}

	// 过滤出.sql文件
	var sqlFiles []fs.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file)
		}
	}

	// 按文件名排序
	sort.Slice(sqlFiles, func(i, j int) bool {
		return sqlFiles[i].Name() < sqlFiles[j].Name()
	})

	return sqlFiles, nil
}

// applyMigration 应用单个迁移
func (m *MigrationManager) applyMigration(ctx context.Context, version, sql string) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 执行迁移SQL
		if err := tx.Exec(sql).Error; err != nil {
			return err
		}

		// 记录迁移版本
		return tx.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)", version, time.Now()).Error
	})
}

// GetMigrationStatus 获取迁移状态
func (m *MigrationManager) GetMigrationStatus(ctx context.Context) ([]Migration, error) {
	var migrations []Migration
	
	if err := m.db.WithContext(ctx).
		Table("schema_migrations").
		Select("version, applied_at").
		Order("version ASC").
		Find(&migrations).Error; err != nil {
		return nil, err
	}

	return migrations, nil
}

// RollbackMigration 回滚迁移（谨慎使用）
func (m *MigrationManager) RollbackMigration(ctx context.Context, version string) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除迁移记录
		if err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", version).Error; err != nil {
			return err
		}

		// 注意：这里不执行回滚SQL，因为回滚需要单独的回滚脚本
		// 实际生产环境应该为每个迁移创建对应的回滚脚本
		
		return nil
	})
}

// RollbackMigrationWithScript 使用回滚脚本回滚迁移
func (m *MigrationManager) RollbackMigrationWithScript(ctx context.Context, version, rollbackSQL string) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 执行回滚SQL
		if err := tx.Exec(rollbackSQL).Error; err != nil {
			return fmt.Errorf("failed to execute rollback SQL: %w", err)
		}

		// 删除迁移记录
		if err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", version).Error; err != nil {
			return fmt.Errorf("failed to delete migration record: %w", err)
		}
		
		return nil
	})
}

// validateMigration 验证迁移脚本
func (m *MigrationManager) validateMigration(sql string) error {
	// 检查是否为空
	if strings.TrimSpace(sql) == "" {
		return fmt.Errorf("migration script is empty")
	}
	
	// 检查是否包含危险操作（可选）
	// 这里可以根据实际需求添加更多的验证规则
	
	return nil
}

// GetMigrationStatusSummary 获取迁移状态摘要
func (m *MigrationManager) GetMigrationStatusSummary(ctx context.Context, migrationsFS embed.FS) (*MigrationStatus, error) {
	// 确保迁移表存在
	if err := m.createMigrationsTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to create migrations table: %w", err)
	}
	
	// 获取已应用的迁移
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	
	// 读取迁移文件
	files, err := m.readMigrationFiles(migrationsFS)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration files: %w", err)
	}
	
	status := &MigrationStatus{
		Total:   len(files),
		Applied: len(applied),
		Pending: len(files) - len(applied),
	}
	
	// 获取最后应用的迁移
	if len(applied) > 0 {
		var lastMigration Migration
		if err := m.db.WithContext(ctx).
			Table("schema_migrations").
			Select("version, applied_at").
			Order("applied_at DESC").
			First(&lastMigration).Error; err == nil {
			status.LastApplied = &lastMigration
		}
	}
	
	return status, nil
}

// IsMigrationApplied 检查指定迁移是否已应用
func (m *MigrationManager) IsMigrationApplied(ctx context.Context, version string) (bool, error) {
	var count int64
	if err := m.db.WithContext(ctx).
		Table("schema_migrations").
		Where("version = ?", version).
		Count(&count).Error; err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// GetPendingMigrations 获取待执行的迁移列表
func (m *MigrationManager) GetPendingMigrations(ctx context.Context, migrationsFS embed.FS) ([]string, error) {
	// 确保迁移表存在
	if err := m.createMigrationsTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to create migrations table: %w", err)
	}
	
	// 获取已应用的迁移
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	
	// 读取迁移文件
	files, err := m.readMigrationFiles(migrationsFS)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration files: %w", err)
	}
	
	// 找出待执行的迁移
	var pending []string
	for _, file := range files {
		version := strings.TrimSuffix(file.Name(), ".sql")
		if _, exists := applied[version]; !exists {
			pending = append(pending, version)
		}
	}
	
	return pending, nil
}

// ApplySpecificMigration 应用指定的迁移
func (m *MigrationManager) ApplySpecificMigration(ctx context.Context, migrationsFS embed.FS, version string) error {
	// 确保迁移表存在
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	
	// 检查迁移是否已应用
	applied, err := m.IsMigrationApplied(ctx, version)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}
	
	if applied {
		return fmt.Errorf("migration %s is already applied", version)
	}
	
	// 读取迁移文件
	filePath := filepath.Join("migrations", version+".sql")
	content, err := fs.ReadFile(migrationsFS, filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}
	
	// 验证迁移脚本
	if err := m.validateMigration(string(content)); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}
	
	// 应用迁移
	if err := m.applyMigration(ctx, version, string(content)); err != nil {
		return fmt.Errorf("failed to apply migration: %w", err)
	}
	
	return nil
}
