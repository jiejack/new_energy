-- 003_add_qa_tables.sql
-- 创建问答相关表

-- 创建问答会话表
CREATE TABLE IF NOT EXISTS qa_sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    title VARCHAR(200),
    status INTEGER DEFAULT 1,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建问答消息表
CREATE TABLE IF NOT EXISTS qa_messages (
    id VARCHAR(36) PRIMARY KEY,
    session_id VARCHAR(36) NOT NULL,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_qa_messages_session FOREIGN KEY (session_id) 
        REFERENCES qa_sessions(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX idx_qa_sessions_user_id ON qa_sessions(user_id);
CREATE INDEX idx_qa_sessions_status ON qa_sessions(status);
CREATE INDEX idx_qa_sessions_created_at ON qa_sessions(created_at);
CREATE INDEX idx_qa_messages_session_id ON qa_messages(session_id);
CREATE INDEX idx_qa_messages_role ON qa_messages(role);
CREATE INDEX idx_qa_messages_created_at ON qa_messages(created_at);

-- 添加注释
COMMENT ON TABLE qa_sessions IS '问答会话表';
COMMENT ON COLUMN qa_sessions.id IS '会话ID';
COMMENT ON COLUMN qa_sessions.user_id IS '用户ID';
COMMENT ON COLUMN qa_sessions.title IS '会话标题';
COMMENT ON COLUMN qa_sessions.status IS '状态: 1-活跃, 2-已归档, 3-已删除';
COMMENT ON COLUMN qa_sessions.created_at IS '创建时间';
COMMENT ON COLUMN qa_sessions.updated_at IS '更新时间';

COMMENT ON TABLE qa_messages IS '问答消息表';
COMMENT ON COLUMN qa_messages.id IS '消息ID';
COMMENT ON COLUMN qa_messages.session_id IS '会话ID';
COMMENT ON COLUMN qa_messages.role IS '角色: user-用户, assistant-助手, system-系统';
COMMENT ON COLUMN qa_messages.content IS '消息内容';
COMMENT ON COLUMN qa_messages.created_at IS '创建时间';

-- 创建更新时间触发器
CREATE TRIGGER update_qa_sessions_updated_at 
BEFORE UPDATE ON qa_sessions 
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();
