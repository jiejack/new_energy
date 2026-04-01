-- 005_add_performance_indexes.sql
-- 添加性能优化索引

-- 为告警表添加复合索引
CREATE INDEX IF NOT EXISTS idx_alarms_station_status_triggered ON alarms(station_id, status, triggered_at DESC);
CREATE INDEX IF NOT EXISTS idx_alarms_device_status_triggered ON alarms(device_id, status, triggered_at DESC);
CREATE INDEX IF NOT EXISTS idx_alarms_level_triggered ON alarms(level, triggered_at DESC) WHERE status = 1;

-- 为设备表添加复合索引
CREATE INDEX IF NOT EXISTS idx_devices_station_type_status ON devices(station_id, type, status);
CREATE INDEX IF NOT EXISTS idx_devices_protocol_status ON devices(protocol, status) WHERE status = 1;

-- 为采集点表添加复合索引
CREATE INDEX IF NOT EXISTS idx_points_station_type_status ON points(station_id, type, status);
CREATE INDEX IF NOT EXISTS idx_points_device_type ON points(device_id, type) WHERE status = 1;

-- 为厂站表添加复合索引
CREATE INDEX IF NOT EXISTS idx_stations_region_type_status ON stations(sub_region_id, type, status);
CREATE INDEX IF NOT EXISTS idx_stations_type_status ON stations(type, status);

-- 为操作日志表添加复合索引
CREATE INDEX IF NOT EXISTS idx_operation_logs_user_action_time ON operation_logs(user_id, action, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_operation_logs_resource ON operation_logs(resource_type, resource_id, created_at DESC);

-- 为问答消息表添加复合索引
CREATE INDEX IF NOT EXISTS idx_qa_messages_session_created ON qa_messages(session_id, created_at DESC);

-- 为告警规则表添加复合索引
CREATE INDEX IF NOT EXISTS idx_alarm_rules_station_status ON alarm_rules(station_id, status) WHERE status = 1;
CREATE INDEX IF NOT EXISTS idx_alarm_rules_device_status ON alarm_rules(device_id, status) WHERE status = 1;
CREATE INDEX IF NOT EXISTS idx_alarm_rules_type_level ON alarm_rules(type, level) WHERE status = 1;

-- 添加部分索引（Partial Index）优化查询
CREATE INDEX IF NOT EXISTS idx_alarms_active ON alarms(station_id, level, triggered_at DESC) WHERE status IN (1, 2);
CREATE INDEX IF NOT EXISTS idx_devices_online ON devices(station_id, last_online DESC) WHERE status = 1;
CREATE INDEX IF NOT EXISTS idx_points_active ON points(station_id, type) WHERE status = 1 AND is_alarm = TRUE;

-- 添加表达式索引
CREATE INDEX IF NOT EXISTS idx_users_lower_username ON users(LOWER(username));
CREATE INDEX IF NOT EXISTS idx_users_lower_email ON users(LOWER(email)) WHERE email IS NOT NULL;

-- 为统计查询添加覆盖索引
CREATE INDEX IF NOT EXISTS idx_alarms_stats ON alarms(station_id, level, status, triggered_at) 
INCLUDE (title, message);

-- 为时间范围查询优化
CREATE INDEX IF NOT EXISTS idx_alarms_triggered_range ON alarms(triggered_at) 
WHERE triggered_at > CURRENT_TIMESTAMP - INTERVAL '30 days';

-- 添加注释
COMMENT ON INDEX idx_alarms_station_status_triggered IS '优化按厂站查询告警的查询性能';
COMMENT ON INDEX idx_alarms_level_triggered IS '优化按级别查询活动告警的查询性能';
COMMENT ON INDEX idx_devices_station_type_status IS '优化按厂站和类型查询设备的查询性能';
COMMENT ON INDEX idx_points_station_type_status IS '优化按厂站和类型查询采集点的查询性能';
COMMENT ON INDEX idx_alarms_active IS '优化查询活动告警的性能';
COMMENT ON INDEX idx_devices_online IS '优化查询在线设备的性能';
