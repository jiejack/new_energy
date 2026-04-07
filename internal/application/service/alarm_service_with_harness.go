package service

import (
	"context"
	"fmt"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/new-energy-monitoring/pkg/harness"
)

// AlarmServiceWithHarness 带有 Harness 验证的告警服务
type AlarmServiceWithHarness struct {
	alarmRepo repository.AlarmRepository
	harness   *AlarmHarness
}

// NewAlarmServiceWithHarness 创建带有 Harness 验证的告警服务
func NewAlarmServiceWithHarness(alarmRepo repository.AlarmRepository) *AlarmServiceWithHarness {
	return &AlarmServiceWithHarness{
		alarmRepo: alarmRepo,
		harness:   NewAlarmHarness(),
	}
}

// NewAlarmServiceWithHarnessComponents 使用自定义组件创建告警服务
func NewAlarmServiceWithHarnessComponents(alarmRepo repository.AlarmRepository, h *harness.Harness) *AlarmServiceWithHarness {
	return &AlarmServiceWithHarness{
		alarmRepo: alarmRepo,
		harness:   NewAlarmHarnessWithComponents(h),
	}
}

// CreateAlarm 创建告警（带 Harness 验证）
func (s *AlarmServiceWithHarness) CreateAlarm(ctx context.Context, req *CreateAlarmRequest) (*entity.Alarm, error) {
	// 1. 使用 Harness 验证请求
	if err := s.harness.ValidateCreateAlarm(ctx, req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 2. 创建告警实体
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

	// 3. 保存到数据库
	if err := s.alarmRepo.Create(ctx, alarm); err != nil {
		return nil, fmt.Errorf("failed to create alarm: %w", err)
	}

	// 4. 创建快照用于审计
	snapshot, err := s.harness.CreateAlarmSnapshot(ctx, alarm)
	if err == nil {
		// 保存快照（可以存储到数据库或文件系统）
		_ = s.harness.GetHarness().SaveSnapshot(alarm.ID, snapshot)
	}

	return alarm, nil
}

// AcknowledgeAlarm 确认告警（带 Harness 验证）
func (s *AlarmServiceWithHarness) AcknowledgeAlarm(ctx context.Context, id, by string) error {
	// 1. 验证请求
	if err := s.harness.ValidateAcknowledgeAlarm(ctx, id, by); err != nil {
		return err
	}

	// 2. 获取告警
	alarm, err := s.alarmRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("alarm not found: %w", err)
	}

	// 3. 检查状态
	if !alarm.IsActive() {
		return fmt.Errorf("alarm is not in active state")
	}

	// 4. 确认告警
	return s.alarmRepo.Acknowledge(ctx, id, by)
}

// ClearAlarm 清除告警（带 Harness 验证）
func (s *AlarmServiceWithHarness) ClearAlarm(ctx context.Context, id string) error {
	// 1. 验证请求
	if err := s.harness.ValidateClearAlarm(ctx, id); err != nil {
		return err
	}

	// 2. 获取告警
	alarm, err := s.alarmRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("alarm not found: %w", err)
	}

	// 3. 检查状态
	if !alarm.IsActive() && !alarm.IsAcknowledged() {
		return fmt.Errorf("alarm cannot be cleared in current state")
	}

	// 4. 清除告警
	return s.alarmRepo.Clear(ctx, id)
}

// GetAlarm 获取告警
func (s *AlarmServiceWithHarness) GetAlarm(ctx context.Context, id string) (*entity.Alarm, error) {
	return s.alarmRepo.GetByID(ctx, id)
}

// GetActiveAlarms 获取活动告警
func (s *AlarmServiceWithHarness) GetActiveAlarms(ctx context.Context, stationID *string, level *entity.AlarmLevel) ([]*entity.Alarm, error) {
	return s.alarmRepo.GetActiveAlarms(ctx, stationID, level)
}

// GetHistoryAlarms 获取历史告警（带 Harness 验证）
func (s *AlarmServiceWithHarness) GetHistoryAlarms(ctx context.Context, stationID *string, startTime, endTime int64) ([]*entity.Alarm, error) {
	// 验证查询参数
	if err := s.harness.ValidateAlarmQuery(ctx, stationID, startTime, endTime); err != nil {
		return nil, err
	}

	return s.alarmRepo.GetHistoryAlarms(ctx, stationID, startTime, endTime)
}

// CountAlarmsByLevel 按级别统计告警
func (s *AlarmServiceWithHarness) CountAlarmsByLevel(ctx context.Context, stationID *string) (map[entity.AlarmLevel]int64, error) {
	return s.alarmRepo.CountByLevel(ctx, stationID)
}

// VerifyAlarmState 验证告警状态（用于测试和审计）
func (s *AlarmServiceWithHarness) VerifyAlarmState(ctx context.Context, alarmID string, expectedState entity.AlarmStatus) (bool, error) {
	alarm, err := s.alarmRepo.GetByID(ctx, alarmID)
	if err != nil {
		return false, err
	}

	return alarm.Status == expectedState, nil
}

// GetHarness 获取 Harness 实例（用于高级操作）
func (s *AlarmServiceWithHarness) GetHarness() *AlarmHarness {
	return s.harness
}
