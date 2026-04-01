package service

import (
	"context"
	"fmt"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

// AlarmRuleService 告警规则服务
type AlarmRuleService struct {
	ruleRepo repository.AlarmRuleRepository
}

// NewAlarmRuleService 创建告警规则服务
func NewAlarmRuleService(ruleRepo repository.AlarmRuleRepository) *AlarmRuleService {
	return &AlarmRuleService{
		ruleRepo: ruleRepo,
	}
}

// CreateAlarmRule 创建告警规则
func (s *AlarmRuleService) CreateAlarmRule(ctx context.Context, req *CreateAlarmRuleRequest) (*entity.AlarmRule, error) {
	// 创建规则实体
	rule := entity.NewAlarmRule(req.Name, req.Type, req.Level)
	rule.Description = req.Description
	rule.Condition = req.Condition
	rule.Threshold = req.Threshold
	rule.Duration = req.Duration
	rule.CreatedBy = req.CreatedBy
	rule.UpdatedBy = req.CreatedBy

	// 设置关联对象
	if req.PointID != nil {
		rule.SetPoint(*req.PointID)
	}
	if req.DeviceID != nil {
		rule.SetDevice(*req.DeviceID)
	}
	if req.StationID != nil {
		rule.SetStation(*req.StationID)
	}

	// 设置通知配置
	if len(req.NotifyChannels) > 0 {
		rule.SetNotifyChannels(req.NotifyChannels)
	}
	if len(req.NotifyUsers) > 0 {
		rule.SetNotifyUsers(req.NotifyUsers)
	}

	// 设置状态
	if req.Status != nil {
		rule.Status = *req.Status
	}

	// 保存到数据库
	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to create alarm rule: %w", err)
	}

	return rule, nil
}

// UpdateAlarmRule 更新告警规则
func (s *AlarmRuleService) UpdateAlarmRule(ctx context.Context, id string, req *UpdateAlarmRuleRequest) (*entity.AlarmRule, error) {
	// 获取现有规则
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("alarm rule not found: %w", err)
	}

	// 更新基本信息
	rule.Update(req.Name, req.Description, req.Level, req.Condition, req.Threshold, req.Duration)
	rule.UpdatedBy = req.UpdatedBy

	// 更新关联对象
	if req.PointID != nil {
		if *req.PointID == "" {
			rule.PointID = nil
		} else {
			rule.SetPoint(*req.PointID)
		}
	}
	if req.DeviceID != nil {
		if *req.DeviceID == "" {
			rule.DeviceID = nil
		} else {
			rule.SetDevice(*req.DeviceID)
		}
	}
	if req.StationID != nil {
		if *req.StationID == "" {
			rule.StationID = nil
		} else {
			rule.SetStation(*req.StationID)
		}
	}

	// 更新通知配置
	if req.NotifyChannels != nil {
		rule.SetNotifyChannels(req.NotifyChannels)
	}
	if req.NotifyUsers != nil {
		rule.SetNotifyUsers(req.NotifyUsers)
	}

	// 更新状态
	if req.Status != nil {
		rule.Status = *req.Status
	}

	// 保存到数据库
	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to update alarm rule: %w", err)
	}

	return rule, nil
}

// DeleteAlarmRule 删除告警规则
func (s *AlarmRuleService) DeleteAlarmRule(ctx context.Context, id string) error {
	// 检查规则是否存在
	_, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("alarm rule not found: %w", err)
	}

	// 删除规则
	if err := s.ruleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete alarm rule: %w", err)
	}

	return nil
}

// GetAlarmRule 获取告警规则详情
func (s *AlarmRuleService) GetAlarmRule(ctx context.Context, id string) (*entity.AlarmRule, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("alarm rule not found: %w", err)
	}
	return rule, nil
}

// ListAlarmRules 获取告警规则列表
func (s *AlarmRuleService) ListAlarmRules(ctx context.Context, query *repository.AlarmRuleQuery) ([]*entity.AlarmRule, int64, error) {
	// 设置默认分页
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 设置默认排序
	if query.OrderBy == "" {
		query.OrderBy = "created_at"
	}
	if query.Order == "" {
		query.Order = "desc"
	}

	return s.ruleRepo.List(ctx, query)
}

// EnableAlarmRule 启用告警规则
func (s *AlarmRuleService) EnableAlarmRule(ctx context.Context, id string, updatedBy string) error {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("alarm rule not found: %w", err)
	}

	rule.Enable()
	rule.UpdatedBy = updatedBy

	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return fmt.Errorf("failed to enable alarm rule: %w", err)
	}

	return nil
}

// DisableAlarmRule 禁用告警规则
func (s *AlarmRuleService) DisableAlarmRule(ctx context.Context, id string, updatedBy string) error {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("alarm rule not found: %w", err)
	}

	rule.Disable()
	rule.UpdatedBy = updatedBy

	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return fmt.Errorf("failed to disable alarm rule: %w", err)
	}

	return nil
}

// GetEnabledRules 获取启用的告警规则
func (s *AlarmRuleService) GetEnabledRules(ctx context.Context) ([]*entity.AlarmRule, error) {
	return s.ruleRepo.GetEnabledRules(ctx)
}

// GetRulesByPointID 根据采集点ID获取告警规则
func (s *AlarmRuleService) GetRulesByPointID(ctx context.Context, pointID string) ([]*entity.AlarmRule, error) {
	return s.ruleRepo.GetRulesByPointID(ctx, pointID)
}

// GetRulesByDeviceID 根据设备ID获取告警规则
func (s *AlarmRuleService) GetRulesByDeviceID(ctx context.Context, deviceID string) ([]*entity.AlarmRule, error) {
	return s.ruleRepo.GetRulesByDeviceID(ctx, deviceID)
}

// GetRulesByStationID 根据厂站ID获取告警规则
func (s *AlarmRuleService) GetRulesByStationID(ctx context.Context, stationID string) ([]*entity.AlarmRule, error) {
	return s.ruleRepo.GetRulesByStationID(ctx, stationID)
}

// CreateAlarmRuleRequest 创建告警规则请求
type CreateAlarmRuleRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	PointID     *string                 `json:"point_id"`
	DeviceID    *string                 `json:"device_id"`
	StationID   *string                 `json:"station_id"`
	Type        entity.AlarmRuleType    `json:"type" binding:"required"`
	Level       entity.AlarmLevel       `json:"level" binding:"required"`
	Condition   string                  `json:"condition"`
	Threshold   float64                 `json:"threshold"`
	Duration    int                     `json:"duration"`
	NotifyChannels []string             `json:"notify_channels"`
	NotifyUsers    []string             `json:"notify_users"`
	Status         *entity.AlarmRuleStatus `json:"status"`
	CreatedBy      string                `json:"created_by"`
}

// UpdateAlarmRuleRequest 更新告警规则请求
type UpdateAlarmRuleRequest struct {
	Name           string                  `json:"name" binding:"required"`
	Description    string                  `json:"description"`
	PointID        *string                 `json:"point_id"`
	DeviceID       *string                 `json:"device_id"`
	StationID      *string                 `json:"station_id"`
	Level          entity.AlarmLevel       `json:"level" binding:"required"`
	Condition      string                  `json:"condition"`
	Threshold      float64                 `json:"threshold"`
	Duration       int                     `json:"duration"`
	NotifyChannels []string                `json:"notify_channels"`
	NotifyUsers    []string                `json:"notify_users"`
	Status         *entity.AlarmRuleStatus `json:"status"`
	UpdatedBy      string                  `json:"updated_by"`
}
