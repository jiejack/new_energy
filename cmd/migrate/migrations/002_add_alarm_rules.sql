-- 002_add_alarm_rules.sql
-- 更新告警规则表结构

-- 删除旧的alarm_rules表（如果存在）
DROP TABLE IF EXISTS alarm_rules CASCADE;

-- 创建告警规则表
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
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_by VARCHAR(100)
);

-- 创建索引
CREATE INDEX idx_alarm_rules_point_id ON alarm_rules(point_id);
CREATE INDEX idx_alarm_rules_device_id ON alarm_rules(device_id);
CREATE INDEX idx_alarm_rules_station_id ON alarm_rules(station_id);
CREATE INDEX idx_alarm_rules_status ON alarm_rules(status);
CREATE INDEX idx_alarm_rules_type ON alarm_rules(type);
CREATE INDEX idx_alarm_rules_level ON alarm_rules(level);

-- 添加注释
COMMENT ON TABLE alarm_rules IS '告警规则表';
COMMENT ON COLUMN alarm_rules.id IS '规则ID';
COMMENT ON COLUMN alarm_rules.name IS '规则名称';
COMMENT ON COLUMN alarm_rules.description IS '规则描述';
COMMENT ON COLUMN alarm_rules.point_id IS '测点ID';
COMMENT ON COLUMN alarm_rules.device_id IS '设备ID';
COMMENT ON COLUMN alarm_rules.station_id IS '厂站ID';
COMMENT ON COLUMN alarm_rules.type IS '规则类型: limit-限值, trend-趋势, custom-自定义';
COMMENT ON COLUMN alarm_rules.level IS '告警级别: 1-提示, 2-警告, 3-严重, 4-紧急';
COMMENT ON COLUMN alarm_rules.condition IS '触发条件表达式';
COMMENT ON COLUMN alarm_rules.threshold IS '阈值';
COMMENT ON COLUMN alarm_rules.duration IS '持续时间(秒)';
COMMENT ON COLUMN alarm_rules.notify_channels IS '通知渠道(JSON数组)';
COMMENT ON COLUMN alarm_rules.notify_users IS '通知用户(JSON数组)';
COMMENT ON COLUMN alarm_rules.status IS '状态: 0-禁用, 1-启用';
COMMENT ON COLUMN alarm_rules.created_at IS '创建时间';
COMMENT ON COLUMN alarm_rules.updated_at IS '更新时间';
COMMENT ON COLUMN alarm_rules.created_by IS '创建人';
COMMENT ON COLUMN alarm_rules.updated_by IS '更新人';

-- 创建更新时间触发器
CREATE TRIGGER update_alarm_rules_updated_at 
BEFORE UPDATE ON alarm_rules 
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();
