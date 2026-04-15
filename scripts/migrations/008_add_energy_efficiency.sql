-- 添加能效分析功能表
-- PostgreSQL 16+

-- 创建能效分析记录表
CREATE TABLE IF NOT EXISTS energy_efficiency_records (
    id VARCHAR(36) PRIMARY KEY,
    record_time TIMESTAMP NOT NULL,
    type VARCHAR(20) NOT NULL,
    target_id VARCHAR(36) NOT NULL,
    target_name VARCHAR(200) NOT NULL,
    input_energy DECIMAL(20,4) NOT NULL,
    output_energy DECIMAL(20,4) NOT NULL,
    efficiency DECIMAL(10,4) NOT NULL,
    efficiency_level VARCHAR(20) NOT NULL,
    benchmark_efficiency DECIMAL(10,4),
    improvement_rate DECIMAL(10,4),
    unit VARCHAR(20) DEFAULT 'kWh',
    period VARCHAR(20) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ee_record_time ON energy_efficiency_records(record_time);
CREATE INDEX idx_ee_type ON energy_efficiency_records(type);
CREATE INDEX idx_ee_target ON energy_efficiency_records(target_id);
CREATE INDEX idx_ee_efficiency ON energy_efficiency_records(efficiency);
CREATE INDEX idx_ee_period ON energy_efficiency_records(period);

-- 创建能效分析表
CREATE TABLE IF NOT EXISTS energy_efficiency_analyses (
    id VARCHAR(36) PRIMARY KEY,
    analysis_time TIMESTAMP NOT NULL,
    type VARCHAR(20) NOT NULL,
    target_id VARCHAR(36) NOT NULL,
    target_name VARCHAR(200) NOT NULL,
    time_range_start TIMESTAMP NOT NULL,
    time_range_end TIMESTAMP NOT NULL,
    avg_efficiency DECIMAL(10,4) NOT NULL,
    max_efficiency DECIMAL(10,4) NOT NULL,
    min_efficiency DECIMAL(10,4) NOT NULL,
    std_dev_efficiency DECIMAL(10,4),
    yoy_change DECIMAL(10,4),
    mom_change DECIMAL(10,4),
    trend VARCHAR(20),
    optimization_suggestions TEXT[],
    saving_potential DECIMAL(20,4),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ee_analysis_time ON energy_efficiency_analyses(analysis_time);
CREATE INDEX idx_ee_analysis_type ON energy_efficiency_analyses(type);
CREATE INDEX idx_ee_analysis_target ON energy_efficiency_analyses(target_id);

-- 为能效分析记录表创建更新时间触发器
CREATE TRIGGER update_energy_efficiency_records_updated_at 
    BEFORE UPDATE ON energy_efficiency_records 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
