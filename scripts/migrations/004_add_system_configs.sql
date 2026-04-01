-- 004_add_system_configs.sql
-- 更新系统配置表结构

-- 备份旧数据到临时表
CREATE TEMP TABLE system_configs_backup AS 
SELECT key, value, description FROM system_configs;

-- 删除旧的system_configs表
DROP TABLE IF EXISTS system_configs CASCADE;

-- 创建新的系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id VARCHAR(36) PRIMARY KEY,
    category VARCHAR(50) NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT,
    value_type VARCHAR(20) DEFAULT 'string',
    description TEXT,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT idx_category_key UNIQUE (category, key)
);

-- 创建索引
CREATE INDEX idx_system_configs_category ON system_configs(category);
CREATE INDEX idx_system_configs_key ON system_configs(key);

-- 插入默认配置
INSERT INTO system_configs (id, category, key, value, value_type, description, created_at, updated_at) VALUES
-- 基本设置
(uuid_generate_v4(), 'basic', 'system_name', '新能源监控系统', 'string', '系统名称', NOW(), NOW()),
(uuid_generate_v4(), 'basic', 'language', 'zh-CN', 'string', '默认语言', NOW(), NOW()),
(uuid_generate_v4(), 'basic', 'timezone', 'Asia/Shanghai', 'string', '时区', NOW(), NOW()),
(uuid_generate_v4(), 'basic', 'version', '1.0.0', 'string', '系统版本', NOW(), NOW()),

-- 告警设置
(uuid_generate_v4(), 'alarm', 'default_level', '2', 'int', '默认告警级别', NOW(), NOW()),
(uuid_generate_v4(), 'alarm', 'sound_enabled', 'true', 'bool', '告警声音', NOW(), NOW()),
(uuid_generate_v4(), 'alarm', 'email_enabled', 'false', 'bool', '邮件通知', NOW(), NOW()),
(uuid_generate_v4(), 'alarm', 'sms_enabled', 'false', 'bool', '短信通知', NOW(), NOW()),
(uuid_generate_v4(), 'alarm', 'auto_clear_hours', '72', 'int', '告警自动清除小时数', NOW(), NOW()),

-- 显示设置
(uuid_generate_v4(), 'display', 'theme', 'light', 'string', '主题', NOW(), NOW()),
(uuid_generate_v4(), 'display', 'page_size', '20', 'int', '分页大小', NOW(), NOW()),
(uuid_generate_v4(), 'display', 'refresh_interval', '30', 'int', '刷新间隔(秒)', NOW(), NOW()),

-- 数据设置
(uuid_generate_v4(), 'data', 'retention_days', '365', 'int', '历史数据保留天数', NOW(), NOW()),
(uuid_generate_v4(), 'data', 'backup_enabled', 'true', 'bool', '启用数据备份', NOW(), NOW()),
(uuid_generate_v4(), 'data', 'backup_interval', '24', 'int', '备份间隔(小时)', NOW(), NOW());

-- 添加注释
COMMENT ON TABLE system_configs IS '系统配置表';
COMMENT ON COLUMN system_configs.id IS '配置ID';
COMMENT ON COLUMN system_configs.category IS '配置分类';
COMMENT ON COLUMN system_configs.key IS '配置键';
COMMENT ON COLUMN system_configs.value IS '配置值';
COMMENT ON COLUMN system_configs.value_type IS '值类型: string, int, bool, json';
COMMENT ON COLUMN system_configs.description IS '配置描述';
COMMENT ON COLUMN system_configs.created_at IS '创建时间';
COMMENT ON COLUMN system_configs.updated_at IS '更新时间';

-- 创建更新时间触发器
CREATE TRIGGER update_system_configs_updated_at 
BEFORE UPDATE ON system_configs 
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();
