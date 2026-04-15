-- 碳排放监测相关表
-- 创建时间: 2026-04-05

-- 碳排放记录表
CREATE TABLE IF NOT EXISTS carbon_emission_records (
    id BIGSERIAL PRIMARY KEY,
    station_id BIGINT NOT NULL,
    device_id BIGINT,
    point_id BIGINT,
    record_date DATE NOT NULL,
    record_time TIMESTAMP NOT NULL,
    
    -- 能源消耗数据
    electricity_consumption DECIMAL(15,4) DEFAULT 0,
    coal_consumption DECIMAL(15,4) DEFAULT 0,
    natural_gas_consumption DECIMAL(15,4) DEFAULT 0,
    oil_consumption DECIMAL(15,4) DEFAULT 0,
    other_energy_consumption DECIMAL(15,4) DEFAULT 0,
    
    -- 碳排放数据
    electricity_emission DECIMAL(15,4) DEFAULT 0,
    coal_emission DECIMAL(15,4) DEFAULT 0,
    natural_gas_emission DECIMAL(15,4) DEFAULT 0,
    oil_emission DECIMAL(15,4) DEFAULT 0,
    other_energy_emission DECIMAL(15,4) DEFAULT 0,
    total_emission DECIMAL(15,4) DEFAULT 0,
    
    -- 排放强度
    emission_intensity DECIMAL(10,4),
    
    -- 计算参数
    emission_factor_electricity DECIMAL(10,4),
    emission_factor_coal DECIMAL(10,4),
    emission_factor_natural_gas DECIMAL(10,4),
    emission_factor_oil DECIMAL(10,4),
    emission_factor_other DECIMAL(10,4),
    
    -- 元数据
    status VARCHAR(20) DEFAULT 'normal',
    remark TEXT,
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_carbon_emission_station_id ON carbon_emission_records(station_id);
CREATE INDEX idx_carbon_emission_device_id ON carbon_emission_records(device_id);
CREATE INDEX idx_carbon_emission_point_id ON carbon_emission_records(point_id);
CREATE INDEX idx_carbon_emission_record_date ON carbon_emission_records(record_date);
CREATE INDEX idx_carbon_emission_record_time ON carbon_emission_records(record_time);
CREATE INDEX idx_carbon_emission_status ON carbon_emission_records(status);
CREATE INDEX idx_carbon_emission_deleted_at ON carbon_emission_records(deleted_at);

-- 碳排放分析表
CREATE TABLE IF NOT EXISTS carbon_emission_analyses (
    id BIGSERIAL PRIMARY KEY,
    station_id BIGINT NOT NULL,
    analysis_date DATE NOT NULL,
    analysis_type VARCHAR(50) NOT NULL,
    
    -- 分析周期
    period_type VARCHAR(20) NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    
    -- 排放总量数据
    total_emission DECIMAL(15,4) DEFAULT 0,
    emission_change_rate DECIMAL(10,4),
    
    -- 分类排放数据
    electricity_emission DECIMAL(15,4) DEFAULT 0,
    coal_emission DECIMAL(15,4) DEFAULT 0,
    natural_gas_emission DECIMAL(15,4) DEFAULT 0,
    oil_emission DECIMAL(15,4) DEFAULT 0,
    other_energy_emission DECIMAL(15,4) DEFAULT 0,
    
    -- 排放强度
    average_emission_intensity DECIMAL(10,4),
    intensity_change_rate DECIMAL(10,4),
    
    -- 峰值数据
    peak_emission DECIMAL(15,4),
    peak_emission_time TIMESTAMP,
    
    -- 减排信息
    reduction_target DECIMAL(15,4),
    actual_reduction DECIMAL(15,4),
    reduction_rate DECIMAL(10,4),
    
    -- 分析结果
    analysis_summary TEXT,
    key_findings TEXT,
    recommendations TEXT,
    
    -- 对比数据
    compared_with_period VARCHAR(50),
    comparison_result TEXT,
    
    -- 元数据
    status VARCHAR(20) DEFAULT 'draft',
    created_by BIGINT,
    updated_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_carbon_emission_analysis_station_id ON carbon_emission_analyses(station_id);
CREATE INDEX idx_carbon_emission_analysis_date ON carbon_emission_analyses(analysis_date);
CREATE INDEX idx_carbon_emission_analysis_type ON carbon_emission_analyses(analysis_type);
CREATE INDEX idx_carbon_emission_analysis_period ON carbon_emission_analyses(period_type, period_start, period_end);
CREATE INDEX idx_carbon_emission_analysis_status ON carbon_emission_analyses(status);
CREATE INDEX idx_carbon_emission_analysis_deleted_at ON carbon_emission_analyses(deleted_at);

-- 添加注释
COMMENT ON TABLE carbon_emission_records IS '碳排放记录表';
COMMENT ON TABLE carbon_emission_analyses IS '碳排放分析表';
