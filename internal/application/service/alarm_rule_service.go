package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type AlarmRuleService struct {
	ruleRepo repository.AlarmRuleRepository
}

func NewAlarmRuleService(ruleRepo repository.AlarmRuleRepository) *AlarmRuleService {
	return &AlarmRuleService{ruleRepo: ruleRepo}
}

type CreateAlarmRuleRequest struct {
	Name           string                `json:"name" binding:"required"`
	Description    string                `json:"description"`
	PointID        *string               `json:"point_id"`
	DeviceID       *string               `json:"device_id"`
	StationID      *string               `json:"station_id"`
	Type           entity.AlarmRuleType  `json:"type" binding:"required"`
	Level          entity.AlarmLevel     `json:"level" binding:"required"`
	Condition      string                `json:"condition" binding:"required"`
	Threshold      float64               `json:"threshold"`
	Duration       int                   `json:"duration"`
	NotifyChannels []string              `json:"notify_channels"`
	NotifyUsers    []string              `json:"notify_users"`
}

type UpdateAlarmRuleRequest struct {
	Name           *string                `json:"name"`
	Description    *string                `json:"description"`
	PointID        *string                `json:"point_id"`
	DeviceID       *string                `json:"device_id"`
	StationID      *string                `json:"station_id"`
	Type           *entity.AlarmRuleType  `json:"type"`
	Level          *entity.AlarmLevel     `json:"level"`
	Condition      *string                `json:"condition"`
	Threshold      *float64               `json:"threshold"`
	Duration       *int                   `json:"duration"`
	NotifyChannels []string               `json:"notify_channels"`
	NotifyUsers    []string               `json:"notify_users"`
	Status         *entity.AlarmRuleStatus `json:"status"`
}

func (s *AlarmRuleService) CreateRule(ctx context.Context, req *CreateAlarmRuleRequest, createdBy string) (*entity.AlarmRule, error) {
	existing, _ := s.ruleRepo.GetByName(ctx, req.Name)
	if existing != nil {
		return nil, fmt.Errorf("alarm rule with name %s already exists", req.Name)
	}

	if req.Type != entity.AlarmRuleTypeLimit && req.Type != entity.AlarmRuleTypeTrend && req.Type != entity.AlarmRuleTypeCustom {
		return nil, fmt.Errorf("invalid alarm rule type: %s", req.Type)
	}
	if req.Condition == "" {
		return nil, fmt.Errorf("condition is required")
	}
	if req.Threshold == 0 && req.Type == entity.AlarmRuleTypeLimit {
		return nil, fmt.Errorf("threshold is required for limit type rules")
	}

	rule := entity.NewAlarmRule(req.Name, req.Type, req.Level, req.Condition)
	rule.ID = uuid.New().String()
	rule.Description = req.Description
	rule.PointID = req.PointID
	rule.DeviceID = req.DeviceID
	rule.StationID = req.StationID
	rule.Threshold = req.Threshold
	rule.Duration = req.Duration
	rule.NotifyChannels = req.NotifyChannels
	rule.NotifyUsers = req.NotifyUsers
	rule.CreatedBy = createdBy
	rule.UpdatedBy = createdBy

	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to create alarm rule: %w", err)
	}

	return rule, nil
}

func (s *AlarmRuleService) UpdateRule(ctx context.Context, id string, req *UpdateAlarmRuleRequest, updatedBy string) (*entity.AlarmRule, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("alarm rule not found: %w", err)
	}

	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Description != nil {
		rule.Description = *req.Description
	}
	if req.Type != nil {
		rule.Type = *req.Type
	}
	if req.Level != nil {
		rule.Level = *req.Level
	}
	if req.Condition != nil {
		rule.Condition = *req.Condition
	}
	if req.Threshold != nil {
		rule.Threshold = *req.Threshold
	}
	if req.Duration != nil {
		rule.Duration = *req.Duration
	}
	if req.NotifyChannels != nil {
		rule.NotifyChannels = req.NotifyChannels
	}
	if req.NotifyUsers != nil {
		rule.NotifyUsers = req.NotifyUsers
	}
	if req.Status != nil {
		rule.Status = *req.Status
	}

	rule.UpdatedBy = updatedBy
	rule.UpdatedAt = time.Now()

	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to update alarm rule: %w", err)
	}

	return rule, nil
}

func (s *AlarmRuleService) DeleteRule(ctx context.Context, id string) error {
	return s.ruleRepo.Delete(ctx, id)
}

func (s *AlarmRuleService) GetRule(ctx context.Context, id string) (*entity.AlarmRule, error) {
	return s.ruleRepo.GetByID(ctx, id)
}

func (s *AlarmRuleService) ListRules(ctx context.Context, query *repository.AlarmRuleQuery) ([]*entity.AlarmRule, int64, error) {
	return s.ruleRepo.List(ctx, query)
}

func (s *AlarmRuleService) GetEnabledRules(ctx context.Context) ([]*entity.AlarmRule, error) {
	return s.ruleRepo.GetEnabledRules(ctx)
}

func (s *AlarmRuleService) EnableRule(ctx context.Context, id string, updatedBy string) (*entity.AlarmRule, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("alarm rule not found: %w", err)
	}
	if rule.Status == entity.AlarmRuleStatusEnabled {
		return rule, nil
	}
	rule.Status = entity.AlarmRuleStatusEnabled
	rule.UpdatedBy = updatedBy
	rule.UpdatedAt = time.Now()
	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to enable alarm rule: %w", err)
	}
	return rule, nil
}

func (s *AlarmRuleService) DisableRule(ctx context.Context, id string, updatedBy string) (*entity.AlarmRule, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("alarm rule not found: %w", err)
	}
	if rule.Status == entity.AlarmRuleStatusDisabled {
		return rule, nil
	}
	rule.Status = entity.AlarmRuleStatusDisabled
	rule.UpdatedBy = updatedBy
	rule.UpdatedAt = time.Now()
	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to disable alarm rule: %w", err)
	}
	return rule, nil
}

func (s *AlarmRuleService) GetRulesByPointID(ctx context.Context, pointID string) ([]*entity.AlarmRule, error) {
	return s.ruleRepo.GetRulesByPointID(ctx, pointID)
}

func (s *AlarmRuleService) GetRulesByDeviceID(ctx context.Context, deviceID string) ([]*entity.AlarmRule, error) {
	return s.ruleRepo.GetRulesByDeviceID(ctx, deviceID)
}

func (s *AlarmRuleService) GetRulesByStationID(ctx context.Context, stationID string) ([]*entity.AlarmRule, error) {
	return s.ruleRepo.GetRulesByStationID(ctx, stationID)
}
