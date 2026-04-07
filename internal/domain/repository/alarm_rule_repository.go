package repository

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
)

type AlarmRuleRepository interface {
	Create(ctx context.Context, rule *entity.AlarmRule) error
	Update(ctx context.Context, rule *entity.AlarmRule) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.AlarmRule, error)
	GetByName(ctx context.Context, name string) (*entity.AlarmRule, error)
	List(ctx context.Context, query *AlarmRuleQuery) ([]*entity.AlarmRule, int64, error)
	GetEnabledRules(ctx context.Context) ([]*entity.AlarmRule, error)
}

type AlarmRuleQuery struct {
	Page      int
	PageSize  int
	Type      *entity.AlarmRuleType
	Level     *entity.AlarmLevel
	Status    *entity.AlarmRuleStatus
	StationID *string
	DeviceID  *string
	PointID   *string
}
