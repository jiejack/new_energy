-- 005_add_alarm_rules.sql
-- 告警规则表

CREATE TABLE IF NOT EXISTS alarm_rules (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    point_id VARCHAR(36),
    device_id VARCHAR(36),
    station_id VARCHAR(36),
    type VARCHAR(20) NOT NULL,
    level INTEGER NOT NULL,
    condition TEXT NOT NULL,
    threshold DOUBLE PRECISION,
    duration INTEGER DEFAULT 0,
    notify_channels TEXT,
    notify_users TEXT,
    status INTEGER DEFAULT 1,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    created_by VARCHAR(100),
    updated_by VARCHAR(100)
);

CREATE INDEX IF NOT EXISTS idx_alarm_rules_point_id ON alarm_rules(point_id);
CREATE INDEX IF NOT EXISTS idx_alarm_rules_device_id ON alarm_rules(device_id);
CREATE INDEX IF NOT EXISTS idx_alarm_rules_station_id ON alarm_rules(station_id);
CREATE INDEX IF NOT EXISTS idx_alarm_rules_status ON alarm_rules(status);
CREATE INDEX IF NOT EXISTS idx_alarm_rules_type ON alarm_rules(type);
CREATE INDEX IF NOT EXISTS idx_alarm_rules_level ON alarm_rules(level);

-- 插入默认告警规则示例
INSERT INTO alarm_rules (id, name, description, type, level, condition, threshold, duration, status, created_at, updated_at) VALUES
('rule_001', '温度超限告警', '设备温度超过设定阈值时触发', 'limit', 3, 'temperature > threshold', 80, 60, 1, NOW(), NOW()),
('rule_002', '功率异常告警', '设备功率低于正常范围时触发', 'limit', 2, 'power < threshold', 0.8, 120, 1, NOW(), NOW()),
('rule_003', '通信中断告警', '设备通信中断超过设定时间时触发', 'trend', 4, 'offline_duration > threshold', 300, 0, 1, NOW(), NOW());
