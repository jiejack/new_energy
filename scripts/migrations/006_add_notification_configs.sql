-- 006_add_notification_configs.sql
-- 通知配置表

CREATE TABLE IF NOT EXISTS notification_configs (
    id VARCHAR(36) PRIMARY KEY,
    type VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    config JSONB,
    enabled BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_notification_configs_type ON notification_configs(type);
CREATE INDEX IF NOT EXISTS idx_notification_configs_enabled ON notification_configs(enabled);

-- 插入默认通知配置
INSERT INTO notification_configs (id, type, name, config, enabled, created_at, updated_at) VALUES
('notif_001', 'email', '邮件通知', '{"smtp_host": "", "smtp_port": 465, "username": "", "password": "", "from": "", "use_tls": true}', false, NOW(), NOW()),
('notif_002', 'sms', '短信通知', '{"access_key": "", "secret_key": "", "sign_name": "", "region": "cn"}', false, NOW(), NOW()),
('notif_003', 'webhook', 'Webhook通知', '{"url": "", "method": "POST", "headers": {}}', false, NOW(), NOW()),
('notif_004', 'wechat', '微信通知', '{"corp_id": "", "agent_id": "", "secret": ""}', false, NOW(), NOW());
