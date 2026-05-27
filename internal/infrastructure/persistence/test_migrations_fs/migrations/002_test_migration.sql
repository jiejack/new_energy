-- Test migration 002
CREATE TABLE IF NOT EXISTS test_table_2 (
    id VARCHAR(36) PRIMARY KEY,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
