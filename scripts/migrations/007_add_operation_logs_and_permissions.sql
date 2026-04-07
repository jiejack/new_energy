-- 007_add_operation_logs.sql
-- 操作日志表

CREATE TABLE IF NOT EXISTS operation_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    username VARCHAR(100),
    method VARCHAR(10),
    path VARCHAR(200),
    action VARCHAR(100),
    resource VARCHAR(100),
    resource_id VARCHAR(36),
    request_ip VARCHAR(50),
    user_agent VARCHAR(500),
    status INTEGER,
    error_msg TEXT,
    duration BIGINT,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_operation_logs_created_at ON operation_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_operation_logs_action ON operation_logs(action);
CREATE INDEX IF NOT EXISTS idx_operation_logs_resource ON operation_logs(resource);

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    code VARCHAR(100) NOT NULL UNIQUE,
    type VARCHAR(20) NOT NULL,
    parent_id VARCHAR(36),
    path VARCHAR(200),
    icon VARCHAR(50),
    sort INTEGER DEFAULT 0,
    status INTEGER DEFAULT 1,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_permissions_parent_id ON permissions(parent_id);
CREATE INDEX IF NOT EXISTS idx_permissions_type ON permissions(type);

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id VARCHAR(36) NOT NULL,
    permission_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);

-- 插入默认权限
INSERT INTO permissions (id, name, code, type, path, icon, sort, status, created_at, updated_at) VALUES
-- 菜单权限
('perm_001', '仪表盘', 'dashboard', 'menu', '/dashboard', 'Odometer', 1, 1, NOW(), NOW()),
('perm_002', '实时监控', 'monitor', 'menu', '/monitor', 'Monitor', 2, 1, NOW(), NOW()),
('perm_003', '设备管理', 'device', 'menu', '/device', 'SetUp', 3, 1, NOW(), NOW()),
('perm_004', '告警管理', 'alarm', 'menu', '/alarm', 'Bell', 4, 1, NOW(), NOW()),
('perm_005', '数据查询', 'data', 'menu', '/data', 'DataAnalysis', 5, 1, NOW(), NOW()),
('perm_006', '系统管理', 'system', 'menu', '/system', 'Tools', 6, 1, NOW(), NOW()),
-- 按钮权限
('perm_007', '新建电站', 'station:create', 'button', '', '', 1, 1, NOW(), NOW()),
('perm_008', '编辑电站', 'station:edit', 'button', '', '', 2, 1, NOW(), NOW()),
('perm_009', '删除电站', 'station:delete', 'button', '', '', 3, 1, NOW(), NOW()),
('perm_010', '新建设备', 'device:create', 'button', '', '', 4, 1, NOW(), NOW()),
('perm_011', '编辑设备', 'device:edit', 'button', '', '', 5, 1, NOW(), NOW()),
('perm_012', '删除设备', 'device:delete', 'button', '', '', 6, 1, NOW(), NOW());
