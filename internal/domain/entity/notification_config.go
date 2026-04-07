package entity

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type NotificationType string

const (
	NotificationTypeEmail   NotificationType = "email"
	NotificationTypeSMS     NotificationType = "sms"
	NotificationTypeWebhook NotificationType = "webhook"
	NotificationTypeWeChat  NotificationType = "wechat"
)

type NotificationConfig struct {
	ID        string           `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Type      NotificationType `json:"type" gorm:"type:varchar(20);not null;uniqueIndex"`
	Name      string           `json:"name" gorm:"type:varchar(100);not null"`
	Config    JSONMap          `json:"config" gorm:"type:json"`
	Enabled   bool             `json:"enabled" gorm:"default:false"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

func (c *NotificationConfig) TableName() string {
	return "notification_configs"
}

type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

type EmailConfig struct {
	SMTPHost string `json:"smtp_host"`
	SMTPPort int    `json:"smtp_port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	UseTLS   bool   `json:"use_tls"`
}

type SMSConfig struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	SignName  string `json:"sign_name"`
	Region    string `json:"region"`
}

type WebhookConfig struct {
	URL    string            `json:"url"`
	Method string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

type WeChatConfig struct {
	CorpID  string `json:"corp_id"`
	AgentID string `json:"agent_id"`
	Secret  string `json:"secret"`
}
