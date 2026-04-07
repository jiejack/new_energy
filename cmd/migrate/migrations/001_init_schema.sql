-- 新能源在线监控系统数据库初始化脚本
-- PostgreSQL 16+

-- 创建区域表
CREATE TABLE IF NOT EXISTS regions (
    id VARCHAR(36) PRIMARY KEY,
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    parent_id VARCHAR(36) REFERENCES regions(id),
    level INTEGER DEFAULT 1,
    sort_order INTEGER DEFAULT 0,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_regions_parent_id ON regions(parent_id);
CREATE INDEX idx_regions_level ON regions(level);

-- 创建子区域表
CREATE TABLE IF NOT EXISTS sub_regions (
    id VARCHAR(36) PRIMARY KEY,
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    region_id VARCHAR(36) NOT NULL REFERENCES regions(id),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sub_regions_region_id ON sub_regions(region_id);

-- 创建厂站表
CREATE TABLE IF NOT EXISTS stations (
    id VARCHAR(36) PRIMARY KEY,
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(50) NOT NULL,
    sub_region_id VARCHAR(36) NOT NULL REFERENCES sub_regions(id),
    capacity DOUBLE PRECISION,
    voltage_level VARCHAR(50),
    longitude DOUBLE PRECISION,
    latitude DOUBLE PRECISION,
    address VARCHAR(500),
    status INTEGER DEFAULT 1,
    commission_date TIMESTAMP,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_stations_sub_region_id ON stations(sub_region_id);
CREATE INDEX idx_stations_type ON stations(type);
CREATE INDEX idx_stations_status ON stations(status);

-- 创建设备表
CREATE TABLE IF NOT EXISTS devices (
    id VARCHAR(36) PRIMARY KEY,
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(50) NOT NULL,
    station_id VARCHAR(36) NOT NULL REFERENCES stations(id),
    manufacturer VARCHAR(100),
    model VARCHAR(100),
    serial_number VARCHAR(100),
    rated_power DOUBLE PRECISION,
    rated_voltage DOUBLE PRECISION,
    rated_current DOUBLE PRECISION,
    protocol VARCHAR(50),
    ip_address VARCHAR(50),
    port INTEGER,
    slave_id INTEGER,
    status INTEGER DEFAULT 0,
    last_online TIMESTAMP,
    install_date TIMESTAMP,
    warranty_date TIMESTAMP,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_devices_station_id ON devices(station_id);
CREATE INDEX idx_devices_type ON devices(type);
CREATE INDEX idx_devices_status ON devices(status);
CREATE INDEX idx_devices_protocol ON devices(protocol);

-- 创建采集点表
CREATE TABLE IF NOT EXISTS points (
    id VARCHAR(36) PRIMARY KEY,
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(20) NOT NULL,
    device_id VARCHAR(36) REFERENCES devices(id),
    station_id VARCHAR(36) REFERENCES stations(id),
    unit VARCHAR(20),
    precision INTEGER,
    min_value DOUBLE PRECISION,
    max_value DOUBLE PRECISION,
    protocol VARCHAR(50),
    address INTEGER,
    data_format VARCHAR(100),
    scan_interval INTEGER,
    deadband DOUBLE PRECISION,
    is_alarm BOOLEAN DEFAULT FALSE,
    alarm_high DOUBLE PRECISION,
    alarm_low DOUBLE PRECISION,
    status INTEGER DEFAULT 1,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_points_device_id ON points(device_id);
CREATE INDEX idx_points_station_id ON points(station_id);
CREATE INDEX idx_points_type ON points(type);
CREATE INDEX idx_points_protocol ON points(protocol);
CREATE INDEX idx_points_status ON points(status);

-- 创建告警表
CREATE TABLE IF NOT EXISTS alarms (
    id VARCHAR(36) PRIMARY KEY,
    point_id VARCHAR(36) REFERENCES points(id),
    device_id VARCHAR(36) REFERENCES devices(id),
    station_id VARCHAR(36) REFERENCES stations(id),
    type VARCHAR(20) NOT NULL,
    level INTEGER NOT NULL,
    title VARCHAR(200) NOT NULL,
    message TEXT,
    value DOUBLE PRECISION,
    threshold DOUBLE PRECISION,
    status INTEGER DEFAULT 1,
    triggered_at TIMESTAMP NOT NULL,
    acknowledged_at TIMESTAMP,
    cleared_at TIMESTAMP,
    acknowledged_by VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_alarms_point_id ON alarms(point_id);
CREATE INDEX idx_alarms_device_id ON alarms(device_id);
CREATE INDEX idx_alarms_station_id ON alarms(station_id);
CREATE INDEX idx_alarms_status ON alarms(status);
CREATE INDEX idx_alarms_level ON alarms(level);
CREATE INDEX idx_alarms_triggered_at ON alarms(triggered_at);

-- 创建告警规则表
CREATE TABLE IF NOT EXISTS alarm_rules (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    point_id VARCHAR(36) REFERENCES points(id),
    rule_type VARCHAR(50) NOT NULL,
    condition_expression TEXT NOT NULL,
    level INTEGER NOT NULL,
    message_template TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_alarm_rules_point_id ON alarm_rules(point_id);
CREATE INDEX idx_alarm_rules_enabled ON alarm_rules(enabled);

-- 创建计算点表
CREATE TABLE IF NOT EXISTS compute_points (
    id VARCHAR(36) PRIMARY KEY,
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    station_id VARCHAR(36) REFERENCES stations(id),
    formula TEXT NOT NULL,
    input_points TEXT[],
    compute_type VARCHAR(50),
    schedule VARCHAR(100),
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_compute_points_station_id ON compute_points(station_id);
CREATE INDEX idx_compute_points_enabled ON compute_points(enabled);

-- 创建统计任务表
CREATE TABLE IF NOT EXISTS statistics_tasks (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    task_type VARCHAR(50) NOT NULL,
    cron_expression VARCHAR(100) NOT NULL,
    config JSONB,
    enabled BOOLEAN DEFAULT TRUE,
    last_run TIMESTAMP,
    next_run TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_statistics_tasks_enabled ON statistics_tasks(enabled);
CREATE INDEX idx_statistics_tasks_next_run ON statistics_tasks(next_run);

-- 创建统计数据表
CREATE TABLE IF NOT EXISTS statistics_data (
    id VARCHAR(36) PRIMARY KEY,
    task_id VARCHAR(36) REFERENCES statistics_tasks(id),
    dimension VARCHAR(100) NOT NULL,
    dimension_value VARCHAR(200) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DOUBLE PRECISION,
    period_type VARCHAR(20) NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_statistics_data_task_id ON statistics_data(task_id);
CREATE INDEX idx_statistics_data_dimension ON statistics_data(dimension, dimension_value);
CREATE INDEX idx_statistics_data_period ON statistics_data(period_type, period_start);

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(256) NOT NULL,
    email VARCHAR(200),
    phone VARCHAR(50),
    real_name VARCHAR(100),
    status INTEGER DEFAULT 1,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);

-- 创建角色表
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    permissions JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    user_id VARCHAR(36) REFERENCES users(id) ON DELETE CASCADE,
    role_id VARCHAR(36) REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

-- 创建操作日志表
CREATE TABLE IF NOT EXISTS operation_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(36),
    details JSONB,
    ip_address VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX idx_operation_logs_action ON operation_logs(action);
CREATE INDEX idx_operation_logs_created_at ON operation_logs(created_at);

-- 创建系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 插入默认系统配置
INSERT INTO system_configs (key, value, description) VALUES
('system.name', '新能源在线监控系统', '系统名称'),
('system.version', '1.0.0', '系统版本'),
('data.retention_days', '365', '历史数据保留天数'),
('alarm.auto_clear_hours', '72', '告警自动清除小时数');

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要的表创建更新时间触发器
CREATE TRIGGER update_regions_updated_at BEFORE UPDATE ON regions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sub_regions_updated_at BEFORE UPDATE ON sub_regions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_stations_updated_at BEFORE UPDATE ON stations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_devices_updated_at BEFORE UPDATE ON devices FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_points_updated_at BEFORE UPDATE ON points FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_system_configs_updated_at BEFORE UPDATE ON system_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
