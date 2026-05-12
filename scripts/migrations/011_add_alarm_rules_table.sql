-- 告警规则表
-- 创建时间: 2026-05-12

CREATE TABLE IF NOT EXISTS alarm_rules (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    point_id VARCHAR(36),
    device_id VARCHAR(36),
    station_id VARCHAR(36),
    type VARCHAR(20) NOT NULL,
    level INT NOT NULL,
    condition_text TEXT NOT NULL,
    threshold DOUBLE PRECISION DEFAULT 0,
    duration INT DEFAULT 0,
    notify_channels TEXT,
    notify_users TEXT,
    status INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_by VARCHAR(100)
);

CREATE UNIQUE INDEX idx_alarm_rules_name ON alarm_rules(name);
CREATE INDEX idx_alarm_rules_point_id ON alarm_rules(point_id);
CREATE INDEX idx_alarm_rules_device_id ON alarm_rules(device_id);
CREATE INDEX idx_alarm_rules_station_id ON alarm_rules(station_id);
CREATE INDEX idx_alarm_rules_type ON alarm_rules(type);
CREATE INDEX idx_alarm_rules_status ON alarm_rules(status);
