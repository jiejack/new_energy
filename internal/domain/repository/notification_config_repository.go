package repository

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
)

type NotificationConfigRepository interface {
	Create(ctx context.Context, config *entity.NotificationConfig) error
	Update(ctx context.Context, config *entity.NotificationConfig) error
	GetByType(ctx context.Context, notifType entity.NotificationType) (*entity.NotificationConfig, error)
	GetAll(ctx context.Context) ([]*entity.NotificationConfig, error)
}

type NotificationConfigQuery struct {
	Type   *entity.NotificationType
	Enabled *bool
}
