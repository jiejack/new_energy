-- AI模块相关表
-- 创建时间: 2026-05-12

-- 功率预测结果表
CREATE TABLE IF NOT EXISTS forecast_results (
    id VARCHAR(36) PRIMARY KEY,
    station_id VARCHAR(36) NOT NULL,
    forecast_type VARCHAR(20) NOT NULL,
    target_time TIMESTAMP NOT NULL,
    predicted_power DECIMAL(12,4) NOT NULL,
    actual_power DECIMAL(12,4),
    accuracy DECIMAL(6,4),
    confidence DECIMAL(6,4),
    confidence_lower DECIMAL(12,4),
    confidence_upper DECIMAL(12,4),
    model_version VARCHAR(50) NOT NULL,
    attribution_type VARCHAR(30),
    attribution_detail TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_forecast_results_station ON forecast_results(station_id);
CREATE INDEX idx_forecast_results_target_time ON forecast_results(target_time);
CREATE INDEX idx_forecast_results_station_type ON forecast_results(station_id, forecast_type);

-- 故障检测结果表
CREATE TABLE IF NOT EXISTS fault_detection_results (
    id VARCHAR(36) PRIMARY KEY,
    device_id VARCHAR(36) NOT NULL,
    fault_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    confidence DECIMAL(6,4) NOT NULL,
    description TEXT,
    status INT DEFAULT 1,
    root_cause TEXT,
    recommendation TEXT,
    remaining_useful_life_hrs INT,
    health_score INT,
    work_order_id VARCHAR(36),
    model_version VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_fault_detection_device ON fault_detection_results(device_id);
CREATE INDEX idx_fault_detection_severity ON fault_detection_results(severity);

-- 边缘节点表
CREATE TABLE IF NOT EXISTS edge_nodes (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    station_id VARCHAR(36) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    status VARCHAR(20) DEFAULT 'offline',
    cpu_usage DECIMAL(5,2),
    memory_usage DECIMAL(5,2),
    model_versions JSONB,
    last_heartbeat TIMESTAMP,
    config_version VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_edge_nodes_station ON edge_nodes(station_id);

-- 模型版本表
CREATE TABLE IF NOT EXISTS model_versions (
    id VARCHAR(36) PRIMARY KEY,
    model_name VARCHAR(100) NOT NULL,
    version VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'training',
    accuracy DECIMAL(6,4),
    training_config TEXT,
    artifact_path VARCHAR(500),
    deployed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_model_versions_name ON model_versions(model_name);
CREATE INDEX idx_model_versions_status ON model_versions(status);
CREATE UNIQUE INDEX idx_model_versions_name_version ON model_versions(model_name, version);
