package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type NotificationConfigService struct {
	configRepo repository.NotificationConfigRepository
}

func NewNotificationConfigService(configRepo repository.NotificationConfigRepository) *NotificationConfigService {
	return &NotificationConfigService{configRepo: configRepo}
}

type UpdateNotificationConfigRequest struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

func (s *NotificationConfigService) GetByType(ctx context.Context, notifType entity.NotificationType) (*entity.NotificationConfig, error) {
	return s.configRepo.GetByType(ctx, notifType)
}

func (s *NotificationConfigService) GetAll(ctx context.Context) ([]*entity.NotificationConfig, error) {
	return s.configRepo.GetAll(ctx)
}

func (s *NotificationConfigService) UpdateConfig(ctx context.Context, notifType entity.NotificationType, req *UpdateNotificationConfigRequest) (*entity.NotificationConfig, error) {
	nc, err := s.configRepo.GetByType(ctx, notifType)
	if err != nil {
		nc = &entity.NotificationConfig{
			ID:        uuid.New().String(),
			Type:      notifType,
			Name:      req.Name,
			Config:    req.Config,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		return nc, s.configRepo.Create(ctx, nc)
	}

	nc.Name = req.Name
	nc.Config = req.Config
	nc.UpdatedAt = time.Now()
	return nc, s.configRepo.Update(ctx, nc)
}

func (s *NotificationConfigService) EnableConfig(ctx context.Context, notifType entity.NotificationType) error {
	nc, err := s.configRepo.GetByType(ctx, notifType)
	if err != nil {
		return fmt.Errorf("config not found: %w", err)
	}
	nc.Enabled = true
	nc.UpdatedAt = time.Now()
	return s.configRepo.Update(ctx, nc)
}

func (s *NotificationConfigService) DisableConfig(ctx context.Context, notifType entity.NotificationType) error {
	nc, err := s.configRepo.GetByType(ctx, notifType)
	if err != nil {
		return fmt.Errorf("config not found: %w", err)
	}
	nc.Enabled = false
	nc.UpdatedAt = time.Now()
	return s.configRepo.Update(ctx, nc)
}

func (s *NotificationConfigService) TestConfig(ctx context.Context, notifType entity.NotificationType) error {
	nc, err := s.configRepo.GetByType(ctx, notifType)
	if err != nil {
		return fmt.Errorf("config not found: %w", err)
	}

	switch notifType {
	case entity.NotificationTypeEmail:
		return s.testEmailConfig(nc.Config)
	case entity.NotificationTypeSMS:
		return s.testSMSConfig(nc.Config)
	case entity.NotificationTypeWebhook:
		return s.testWebhookConfig(nc.Config)
	case entity.NotificationTypeWeChat:
		return s.testWeChatConfig(nc.Config)
	default:
		return fmt.Errorf("unsupported notification type: %s", notifType)
	}
}

func (s *NotificationConfigService) testEmailConfig(config map[string]interface{}) error {
	smtpHost, ok := config["smtp_host"].(string)
	if !ok || smtpHost == "" {
		return fmt.Errorf("smtp_host is required")
	}
	return nil
}

func (s *NotificationConfigService) testSMSConfig(config map[string]interface{}) error {
	accessKey, ok := config["access_key"].(string)
	if !ok || accessKey == "" {
		return fmt.Errorf("access_key is required")
	}
	return nil
}

func (s *NotificationConfigService) testWebhookConfig(config map[string]interface{}) error {
	url, ok := config["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("url is required")
	}
	return nil
}

func (s *NotificationConfigService) testWeChatConfig(config map[string]interface{}) error {
	corpID, ok := config["corp_id"].(string)
	if !ok || corpID == "" {
		return fmt.Errorf("corp_id is required")
	}
	return nil
}
