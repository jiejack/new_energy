package service

import (
	"context"
	"fmt"
	
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type AlarmService struct {
	alarmRepo repository.AlarmRepository
}

func NewAlarmService(alarmRepo repository.AlarmRepository) *AlarmService {
	return &AlarmService{alarmRepo: alarmRepo}
}

func (s *AlarmService) CreateAlarm(ctx context.Context, req *CreateAlarmRequest) (*entity.Alarm, error) {
	alarm := entity.NewAlarm(
		req.PointID,
		req.DeviceID,
		req.StationID,
		req.Type,
		req.Level,
		req.Title,
		req.Message,
	)
	alarm.Value = req.Value
	alarm.Threshold = req.Threshold
	
	if err := s.alarmRepo.Create(ctx, alarm); err != nil {
		return nil, fmt.Errorf("failed to create alarm: %w", err)
	}
	
	return alarm, nil
}

func (s *AlarmService) GetAlarm(ctx context.Context, id string) (*entity.Alarm, error) {
	return s.alarmRepo.GetByID(ctx, id)
}

func (s *AlarmService) GetActiveAlarms(ctx context.Context, stationID *string, level *entity.AlarmLevel) ([]*entity.Alarm, error) {
	return s.alarmRepo.GetActiveAlarms(ctx, stationID, level)
}

func (s *AlarmService) GetHistoryAlarms(ctx context.Context, stationID *string, startTime, endTime int64) ([]*entity.Alarm, error) {
	return s.alarmRepo.GetHistoryAlarms(ctx, stationID, startTime, endTime)
}

func (s *AlarmService) AcknowledgeAlarm(ctx context.Context, id, by string) error {
	alarm, err := s.alarmRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("alarm not found: %w", err)
	}
	
	if !alarm.IsActive() {
		return fmt.Errorf("alarm is not in active state")
	}
	
	return s.alarmRepo.Acknowledge(ctx, id, by)
}

func (s *AlarmService) ClearAlarm(ctx context.Context, id string) error {
	alarm, err := s.alarmRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("alarm not found: %w", err)
	}
	
	if !alarm.IsActive() && !alarm.IsAcknowledged() {
		return fmt.Errorf("alarm cannot be cleared in current state")
	}
	
	return s.alarmRepo.Clear(ctx, id)
}

func (s *AlarmService) CountAlarmsByLevel(ctx context.Context, stationID *string) (map[entity.AlarmLevel]int64, error) {
	return s.alarmRepo.CountByLevel(ctx, stationID)
}

type CreateAlarmRequest struct {
	PointID   string          `json:"point_id"`
	DeviceID  string          `json:"device_id"`
	StationID string          `json:"station_id"`
	Type      entity.AlarmType `json:"type" binding:"required"`
	Level     entity.AlarmLevel `json:"level" binding:"required"`
	Title     string          `json:"title" binding:"required"`
	Message   string          `json:"message"`
	Value     float64         `json:"value"`
	Threshold float64         `json:"threshold"`
}
