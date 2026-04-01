package repository

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
)

// AlarmRuleRepository 告警规则仓储接口
type AlarmRuleRepository interface {
	// Create 创建告警规则
	Create(ctx context.Context, rule *entity.AlarmRule) error

	// Update 更新告警规则
	Update(ctx context.Context, rule *entity.AlarmRule) error

	// Delete 删除告警规则
	Delete(ctx context.Context, id string) error

	// GetByID 根据ID获取告警规则
	GetByID(ctx context.Context, id string) (*entity.AlarmRule, error)

	// List 获取告警规则列表
	List(ctx context.Context, query *AlarmRuleQuery) ([]*entity.AlarmRule, int64, error)

	// GetEnabledRules 获取启用的告警规则
	GetEnabledRules(ctx context.Context) ([]*entity.AlarmRule, error)

	// GetRulesByPointID 根据采集点ID获取告警规则
	GetRulesByPointID(ctx context.Context, pointID string) ([]*entity.AlarmRule, error)

	// GetRulesByDeviceID 根据设备ID获取告警规则
	GetRulesByDeviceID(ctx context.Context, deviceID string) ([]*entity.AlarmRule, error)

	// GetRulesByStationID 根据厂站ID获取告警规则
	GetRulesByStationID(ctx context.Context, stationID string) ([]*entity.AlarmRule, error)
}

// AlarmRuleQuery 告警规则查询条件
type AlarmRuleQuery struct {
	// 分页
	Page     int `json:"page"`
	PageSize int `json:"page_size"`

	// 过滤条件
	Name      string                `json:"name"`
	Type      *entity.AlarmRuleType `json:"type"`
	Level     *entity.AlarmLevel    `json:"level"`
	Status    *entity.AlarmRuleStatus `json:"status"`
	PointID   string                `json:"point_id"`
	DeviceID  string                `json:"device_id"`
	StationID string                `json:"station_id"`

	// 排序
	OrderBy string `json:"order_by"`
	Order   string `json:"order"` // asc or desc
}
