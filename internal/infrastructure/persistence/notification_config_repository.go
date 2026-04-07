package persistence

import (
	"context"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type notificationConfigRepository struct {
	db *Database
}

func NewNotificationConfigRepository(db *Database) repository.NotificationConfigRepository {
	return &notificationConfigRepository{db: db}
}

func (r *notificationConfigRepository) Create(ctx context.Context, config *entity.NotificationConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

func (r *notificationConfigRepository) Update(ctx context.Context, config *entity.NotificationConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *notificationConfigRepository) GetByType(ctx context.Context, notifType entity.NotificationType) (*entity.NotificationConfig, error) {
	var config entity.NotificationConfig
	err := r.db.WithContext(ctx).Where("type = ?", notifType).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *notificationConfigRepository) GetAll(ctx context.Context) ([]*entity.NotificationConfig, error) {
	var configs []*entity.NotificationConfig
	err := r.db.WithContext(ctx).Find(&configs).Error
	return configs, err
}
