-- Add operation_logs table
CREATE TABLE IF NOT EXISTS operation_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    username VARCHAR(100),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(36),
    details JSONB,
    ip_address VARCHAR(50),
    user_agent VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for operation_logs
CREATE INDEX IF NOT EXISTS idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_operation_logs_action ON operation_logs(action);
CREATE INDEX IF NOT EXISTS idx_operation_logs_created_at ON operation_logs(created_at);

-- Add comment
COMMENT ON TABLE operation_logs IS '操作日志表';
COMMENT ON COLUMN operation_logs.id IS '日志ID';
COMMENT ON COLUMN operation_logs.user_id IS '用户ID';
COMMENT ON COLUMN operation_logs.username IS '用户名';
COMMENT ON COLUMN operation_logs.action IS '操作类型';
COMMENT ON COLUMN operation_logs.resource_type IS '资源类型';
COMMENT ON COLUMN operation_logs.resource_id IS '资源ID';
COMMENT ON COLUMN operation_logs.details IS '操作详情';
COMMENT ON COLUMN operation_logs.ip_address IS 'IP地址';
COMMENT ON COLUMN operation_logs.user_agent IS '用户代理';
COMMENT ON COLUMN operation_logs.created_at IS '创建时间';
