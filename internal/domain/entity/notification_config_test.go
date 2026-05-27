package entity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotificationConfig_TableName(t *testing.T) {
	config := NotificationConfig{}
	assert.Equal(t, "notification_configs", config.TableName())
}

func TestNotificationType_Constants(t *testing.T) {
	assert.Equal(t, NotificationType("email"), NotificationTypeEmail)
	assert.Equal(t, NotificationType("sms"), NotificationTypeSMS)
	assert.Equal(t, NotificationType("webhook"), NotificationTypeWebhook)
	assert.Equal(t, NotificationType("wechat"), NotificationTypeWeChat)
}

func TestJSONMap_Value(t *testing.T) {
	jm := JSONMap{"smtp_host": "smtp.example.com", "smtp_port": 587}
	v, err := jm.Value()
	assert.NoError(t, err)
	assert.NotNil(t, v)

	var parsed map[string]interface{}
	err = json.Unmarshal(v.([]byte), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, "smtp.example.com", parsed["smtp_host"])
}

func TestJSONMap_Value_Nil(t *testing.T) {
	var jm JSONMap
	v, err := jm.Value()
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestJSONMap_Scan(t *testing.T) {
	jsonData := `{"smtp_host":"smtp.example.com","smtp_port":587}`
	jm := JSONMap{}
	err := jm.Scan([]byte(jsonData))
	assert.NoError(t, err)
	assert.Equal(t, "smtp.example.com", jm["smtp_host"])
}

func TestJSONMap_Scan_Nil(t *testing.T) {
	jm := JSONMap{"existing": "data"}
	err := jm.Scan(nil)
	assert.NoError(t, err)
	assert.Nil(t, jm)
}

func TestJSONMap_Scan_InvalidType(t *testing.T) {
	jm := JSONMap{}
	err := jm.Scan(12345)
	assert.NoError(t, err)
}

func TestEmailConfig(t *testing.T) {
	config := EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		From:     "noreply@example.com",
		UseTLS:   true,
	}
	assert.Equal(t, "smtp.example.com", config.SMTPHost)
	assert.Equal(t, 587, config.SMTPPort)
	assert.True(t, config.UseTLS)
}

func TestSMSConfig(t *testing.T) {
	config := SMSConfig{
		AccessKey: "test-key",
		SecretKey: "test-secret",
		SignName:  "NEM",
		Region:    "cn-hangzhou",
	}
	assert.Equal(t, "test-key", config.AccessKey)
	assert.Equal(t, "cn-hangzhou", config.Region)
}

func TestWebhookConfig(t *testing.T) {
	config := WebhookConfig{
		URL:     "https://example.com/webhook",
		Method:  "POST",
		Headers: map[string]string{"Authorization": "Bearer token"},
	}
	assert.Equal(t, "https://example.com/webhook", config.URL)
	assert.Equal(t, "POST", config.Method)
}

func TestWeChatConfig(t *testing.T) {
	config := WeChatConfig{
		CorpID:  "corp-123",
		AgentID: "agent-456",
		Secret:  "secret-789",
	}
	assert.Equal(t, "corp-123", config.CorpID)
	assert.Equal(t, "agent-456", config.AgentID)
}
