package service

import (
	"context"
	"fmt"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/pkg/harness"
)

// AlarmHarness 告警 Harness 验证器
type AlarmHarness struct {
	harness *harness.Harness
}

// NewAlarmHarness 创建新的告警 Harness 实例
func NewAlarmHarness() *AlarmHarness {
	return &AlarmHarness{
		harness: harness.NewHarness(),
	}
}

// NewAlarmHarnessWithComponents 使用自定义组件创建告警 Harness 实例
func NewAlarmHarnessWithComponents(h *harness.Harness) *AlarmHarness {
	return &AlarmHarness{
		harness: h,
	}
}

// ValidateCreateAlarm 验证创建告警请求
func (ah *AlarmHarness) ValidateCreateAlarm(ctx context.Context, req *CreateAlarmRequest) error {
	// 基础验证
	if err := ah.harness.Validate(ctx, req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 业务规则验证
	if err := ah.validateAlarmBusinessRules(req); err != nil {
		return err
	}

	return nil
}

// validateAlarmBusinessRules 验证告警业务规则
func (ah *AlarmHarness) validateAlarmBusinessRules(req *CreateAlarmRequest) error {
	// 验证告警级别
	if !isValidAlarmLevel(req.Level) {
		return fmt.Errorf("invalid alarm level: %d", req.Level)
	}

	// 验证告警类型
	if !isValidAlarmType(req.Type) {
		return fmt.Errorf("invalid alarm type: %s", req.Type)
	}

	// 验证标题不能为空
	if req.Title == "" {
		return fmt.Errorf("alarm title cannot be empty")
	}

	// 验证阈值关系（如果设置了值和阈值）
	if req.Value != 0 && req.Threshold != 0 {
		// 根据告警类型验证值和阈值的关系
		switch req.Type {
		case entity.AlarmTypeLimit:
			// 限值告警：值应该超过阈值
			if req.Level >= entity.AlarmLevelWarning {
				// 警告及以上级别的限值告警，值应该显著偏离阈值
				deviation := req.Value - req.Threshold
				if deviation == 0 {
					return fmt.Errorf("limit alarm value should deviate from threshold")
				}
			}
		}
	}

	return nil
}

// ValidateAcknowledgeAlarm 验证确认告警请求
func (ah *AlarmHarness) ValidateAcknowledgeAlarm(ctx context.Context, alarmID, operator string) error {
	// 基础验证
	if alarmID == "" {
		return fmt.Errorf("alarm ID cannot be empty")
	}
	if operator == "" {
		return fmt.Errorf("operator cannot be empty")
	}

	return nil
}

// ValidateClearAlarm 验证清除告警请求
func (ah *AlarmHarness) ValidateClearAlarm(ctx context.Context, alarmID string) error {
	// 基础验证
	if alarmID == "" {
		return fmt.Errorf("alarm ID cannot be empty")
	}

	return nil
}

// ValidateAlarmQuery 验证告警查询请求
func (ah *AlarmHarness) ValidateAlarmQuery(ctx context.Context, stationID *string, startTime, endTime int64) error {
	// 验证时间范围
	if startTime > 0 && endTime > 0 {
		if startTime > endTime {
			return fmt.Errorf("start time cannot be greater than end time")
		}
	}

	return nil
}

// VerifyAlarmOutput 验证告警输出
func (ah *AlarmHarness) VerifyAlarmOutput(ctx context.Context, expected, actual *entity.Alarm) (bool, error) {
	return ah.harness.Verify(ctx, expected, actual)
}

// CreateAlarmSnapshot 创建告警快照
func (ah *AlarmHarness) CreateAlarmSnapshot(ctx context.Context, alarm *entity.Alarm) ([]byte, error) {
	return ah.harness.Snapshot(ctx, alarm)
}

// GetHarness 获取底层 Harness 实例
func (ah *AlarmHarness) GetHarness() *harness.Harness {
	return ah.harness
}

// isValidAlarmLevel 验证告警级别是否有效
func isValidAlarmLevel(level entity.AlarmLevel) bool {
	return level == entity.AlarmLevelInfo ||
		level == entity.AlarmLevelWarning ||
		level == entity.AlarmLevelMajor ||
		level == entity.AlarmLevelCritical
}

// isValidAlarmType 验证告警类型是否有效
func isValidAlarmType(alarmType entity.AlarmType) bool {
	return alarmType == entity.AlarmTypeLimit ||
		alarmType == entity.AlarmTypeStatus ||
		alarmType == entity.AlarmTypeComm ||
		alarmType == entity.AlarmTypeSystem ||
		alarmType == entity.AlarmTypeDevice
}
