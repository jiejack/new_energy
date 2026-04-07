package persistence

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type alarmRuleRepository struct {
	db *Database
}

func NewAlarmRuleRepository(db *Database) repository.AlarmRuleRepository {
	return &alarmRuleRepository{db: db}
}

func (r *alarmRuleRepository) Create(ctx context.Context, rule *entity.AlarmRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *alarmRuleRepository) Update(ctx context.Context, rule *entity.AlarmRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *alarmRuleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.AlarmRule{}, "id = ?", id).Error
}

func (r *alarmRuleRepository) GetByID(ctx context.Context, id string) (*entity.AlarmRule, error) {
	var rule entity.AlarmRule
	err := r.db.WithContext(ctx).First(&rule, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *alarmRuleRepository) GetByName(ctx context.Context, name string) (*entity.AlarmRule, error) {
	var rule entity.AlarmRule
	err := r.db.WithContext(ctx).First(&rule, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *alarmRuleRepository) List(ctx context.Context, query *repository.AlarmRuleQuery) ([]*entity.AlarmRule, int64, error) {
	var rules []*entity.AlarmRule
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.AlarmRule{})

	if query.Type != nil {
		db = db.Where("type = ?", *query.Type)
	}
	if query.Level != nil {
		db = db.Where("level = ?", *query.Level)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.StationID != nil {
		db = db.Where("station_id = ?", *query.StationID)
	}
	if query.DeviceID != nil {
		db = db.Where("device_id = ?", *query.DeviceID)
	}
	if query.PointID != nil {
		db = db.Where("point_id = ?", *query.PointID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&rules).Error; err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

func (r *alarmRuleRepository) GetEnabledRules(ctx context.Context) ([]*entity.AlarmRule, error) {
	var rules []*entity.AlarmRule
	err := r.db.WithContext(ctx).Where("status = ?", entity.AlarmRuleStatusEnabled).Find(&rules).Error
	return rules, err
}
