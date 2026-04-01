package persistence

import (
	"context"
	"fmt"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// AlarmRuleRepository 告警规则仓储实现
type AlarmRuleRepository struct {
	db *Database
}

// NewAlarmRuleRepository 创建告警规则仓储
func NewAlarmRuleRepository(db *Database) *AlarmRuleRepository {
	return &AlarmRuleRepository{db: db}
}

// Create 创建告警规则
func (r *AlarmRuleRepository) Create(ctx context.Context, rule *entity.AlarmRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// Update 更新告警规则
func (r *AlarmRuleRepository) Update(ctx context.Context, rule *entity.AlarmRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

// Delete 删除告警规则
func (r *AlarmRuleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.AlarmRule{}, "id = ?", id).Error
}

// GetByID 根据ID获取告警规则
func (r *AlarmRuleRepository) GetByID(ctx context.Context, id string) (*entity.AlarmRule, error) {
	var rule entity.AlarmRule
	err := r.db.WithContext(ctx).First(&rule, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// List 获取告警规则列表
func (r *AlarmRuleRepository) List(ctx context.Context, query *repository.AlarmRuleQuery) ([]*entity.AlarmRule, int64, error) {
	var rules []*entity.AlarmRule
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.AlarmRule{})

	// 应用过滤条件
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.Type != nil {
		db = db.Where("type = ?", *query.Type)
	}
	if query.Level != nil {
		db = db.Where("level = ?", *query.Level)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.PointID != "" {
		db = db.Where("point_id = ?", query.PointID)
	}
	if query.DeviceID != "" {
		db = db.Where("device_id = ?", query.DeviceID)
	}
	if query.StationID != "" {
		db = db.Where("station_id = ?", query.StationID)
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count alarm rules: %w", err)
	}

	// 应用排序
	orderClause := query.OrderBy
	if query.Order == "desc" {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}
	db = db.Order(orderClause)

	// 应用分页
	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Find(&rules).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list alarm rules: %w", err)
	}

	return rules, total, nil
}

// GetEnabledRules 获取启用的告警规则
func (r *AlarmRuleRepository) GetEnabledRules(ctx context.Context) ([]*entity.AlarmRule, error) {
	var rules []*entity.AlarmRule
	err := r.db.WithContext(ctx).
		Where("status = ?", entity.AlarmRuleStatusEnabled).
		Order("created_at DESC").
		Find(&rules).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled alarm rules: %w", err)
	}
	return rules, nil
}

// GetRulesByPointID 根据采集点ID获取告警规则
func (r *AlarmRuleRepository) GetRulesByPointID(ctx context.Context, pointID string) ([]*entity.AlarmRule, error) {
	var rules []*entity.AlarmRule
	err := r.db.WithContext(ctx).
		Where("point_id = ? AND status = ?", pointID, entity.AlarmRuleStatusEnabled).
		Order("created_at DESC").
		Find(&rules).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get alarm rules by point ID: %w", err)
	}
	return rules, nil
}

// GetRulesByDeviceID 根据设备ID获取告警规则
func (r *AlarmRuleRepository) GetRulesByDeviceID(ctx context.Context, deviceID string) ([]*entity.AlarmRule, error) {
	var rules []*entity.AlarmRule
	err := r.db.WithContext(ctx).
		Where("device_id = ? AND status = ?", deviceID, entity.AlarmRuleStatusEnabled).
		Order("created_at DESC").
		Find(&rules).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get alarm rules by device ID: %w", err)
	}
	return rules, nil
}

// GetRulesByStationID 根据厂站ID获取告警规则
func (r *AlarmRuleRepository) GetRulesByStationID(ctx context.Context, stationID string) ([]*entity.AlarmRule, error) {
	var rules []*entity.AlarmRule
	err := r.db.WithContext(ctx).
		Where("station_id = ? AND status = ?", stationID, entity.AlarmRuleStatusEnabled).
		Order("created_at DESC").
		Find(&rules).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get alarm rules by station ID: %w", err)
	}
	return rules, nil
}
