-- 006_add_table_partitions.sql
-- 为大表添加分区支持

-- 为告警表创建分区（按月分区）
-- 注意：PostgreSQL不支持直接修改现有表为分区表，需要重建表

-- 创建新的分区告警表
CREATE TABLE IF NOT EXISTS alarms_partitioned (
    id VARCHAR(36) NOT NULL,
    point_id VARCHAR(36),
    device_id VARCHAR(36),
    station_id VARCHAR(36),
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
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, triggered_at)
) PARTITION BY RANGE (triggered_at);

-- 创建默认分区
CREATE TABLE IF NOT EXISTS alarms_default PARTITION OF alarms_partitioned DEFAULT;

-- 创建最近3个月的分区
CREATE TABLE IF NOT EXISTS alarms_2024_01 PARTITION OF alarms_partitioned
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE IF NOT EXISTS alarms_2024_02 PARTITION OF alarms_partitioned
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

CREATE TABLE IF NOT EXISTS alarms_2024_03 PARTITION OF alarms_partitioned
    FOR VALUES FROM ('2024-03-01') TO ('2024-04-01');

-- 为分区表创建索引
CREATE INDEX IF NOT EXISTS idx_alarms_part_station ON alarms_partitioned(station_id);
CREATE INDEX IF NOT EXISTS idx_alarms_part_device ON alarms_partitioned(device_id);
CREATE INDEX IF NOT EXISTS idx_alarms_part_status ON alarms_partitioned(status);

-- 为操作日志表创建分区（按月分区）
CREATE TABLE IF NOT EXISTS operation_logs_partitioned (
    id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(36),
    details JSONB,
    ip_address VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- 创建默认分区
CREATE TABLE IF NOT EXISTS operation_logs_default PARTITION OF operation_logs_partitioned DEFAULT;

-- 创建最近3个月的分区
CREATE TABLE IF NOT EXISTS operation_logs_2024_01 PARTITION OF operation_logs_partitioned
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE IF NOT EXISTS operation_logs_2024_02 PARTITION OF operation_logs_partitioned
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

CREATE TABLE IF NOT EXISTS operation_logs_2024_03 PARTITION OF operation_logs_partitioned
    FOR VALUES FROM ('2024-03-01') TO ('2024-04-01');

-- 为分区表创建索引
CREATE INDEX IF NOT EXISTS idx_operation_logs_part_user ON operation_logs_partitioned(user_id);
CREATE INDEX IF NOT EXISTS idx_operation_logs_part_action ON operation_logs_partitioned(action);

-- 创建分区管理函数
CREATE OR REPLACE FUNCTION create_monthly_partition(
    table_name TEXT,
    partition_date DATE
) RETURNS VOID AS $$
DECLARE
    partition_name TEXT;
    start_date TEXT;
    end_date TEXT;
BEGIN
    -- 生成分区名称
    partition_name := table_name || '_' || TO_CHAR(partition_date, 'YYYY_MM');
    
    -- 计算时间范围
    start_date := TO_CHAR(partition_date, 'YYYY-MM-DD');
    end_date := TO_CHAR(partition_date + INTERVAL '1 month', 'YYYY-MM-DD');
    
    -- 创建分区
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        table_name,
        start_date,
        end_date
    );
    
    RAISE NOTICE 'Created partition % for table %', partition_name, table_name;
END;
$$ LANGUAGE plpgsql;

-- 创建自动分区维护函数（每月运行一次）
CREATE OR REPLACE FUNCTION maintain_partitions() RETURNS VOID AS $$
DECLARE
    next_month DATE;
BEGIN
    -- 为下个月创建分区
    next_month := DATE_TRUNC('month', CURRENT_DATE + INTERVAL '1 month');
    
    -- 为告警表创建分区
    PERFORM create_monthly_partition('alarms_partitioned', next_month);
    
    -- 为操作日志表创建分区
    PERFORM create_monthly_partition('operation_logs_partitioned', next_month);
    
    RAISE NOTICE 'Partition maintenance completed for %', next_month;
END;
$$ LANGUAGE plpgsql;

-- 添加注释
COMMENT ON TABLE alarms_partitioned IS '告警分区表（按月分区）';
COMMENT ON TABLE operation_logs_partitioned IS '操作日志分区表（按月分区）';
COMMENT ON FUNCTION create_monthly_partition IS '创建月度分区';
COMMENT ON FUNCTION maintain_partitions IS '自动维护分区（每月运行）';

-- 注意：实际迁移数据时需要：
-- 1. 在低峰期执行
-- 2. 使用 INSERT INTO ... SELECT FROM 迁移数据
-- 3. 重命名表切换应用
-- 示例：
-- INSERT INTO alarms_partitioned SELECT * FROM alarms;
-- ALTER TABLE alarms RENAME TO alarms_old;
-- ALTER TABLE alarms_partitioned RENAME TO alarms;
