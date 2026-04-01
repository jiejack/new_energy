# PostgreSQL 数据库连接和迁移管理

## 概述

本项目实现了完整的PostgreSQL数据库连接池管理和迁移管理功能，包括：

- 数据库连接池管理
- 健康检查和监控
- 数据库迁移管理
- 性能优化索引和分区表

## 功能特性

### 1. 数据库连接管理

#### 连接池配置

```go
config := DatabaseConfig{
    Host:            "localhost",
    Port:            5432,
    User:            "postgres",
    Password:        "password",
    DBName:          "nem_system",
    SSLMode:         "disable",
    MaxOpenConns:    100,           // 最大打开连接数
    MaxIdleConns:    10,            // 最大空闲连接数
    ConnMaxLifetime: time.Hour,     // 连接最大生命周期
    ConnMaxIdleTime: 10 * time.Minute, // 空闲连接最大生命周期
}
```

#### 创建连接

```go
db, err := NewDatabase(config)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

#### 健康检查

```go
ctx := context.Background()
status, err := db.HealthCheck(ctx)
if err != nil {
    log.Printf("Health check failed: %v", err)
}

fmt.Printf("Status: %s\n", status.Status)
fmt.Printf("Database Version: %s\n", status.Details["database_version"])
fmt.Printf("Connection Usage: %s\n", status.Details["connection_usage"])
```

#### 连接池监控

```go
stats := db.GetStats()
fmt.Printf("Open Connections: %d\n", stats.OpenConnections)
fmt.Printf("In Use: %d\n", stats.InUse)
fmt.Printf("Idle: %d\n", stats.Idle)
fmt.Printf("Wait Count: %d\n", stats.WaitCount)
```

#### 重连机制

```go
ctx := context.Background()
if err := db.Reconnect(ctx); err != nil {
    log.Printf("Reconnect failed: %v", err)
}
```

### 2. 数据库迁移管理

#### 创建迁移管理器

```go
//go:embed migrations/*.sql
var migrationsFS embed.FS

manager := NewMigrationManager(db)
```

#### 执行所有迁移

```go
ctx := context.Background()
if err := manager.RunMigrations(ctx, migrationsFS); err != nil {
    log.Fatal(err)
}
```

#### 执行指定数量的迁移

```go
if err := manager.RunMigrationsWithLimit(ctx, migrationsFS, 1); err != nil {
    log.Fatal(err)
}
```

#### 获取迁移状态

```go
// 获取状态摘要
status, err := manager.GetMigrationStatusSummary(ctx, migrationsFS)
fmt.Printf("Total: %d, Applied: %d, Pending: %d\n", 
    status.Total, status.Applied, status.Pending)

// 获取待执行的迁移
pending, err := manager.GetPendingMigrations(ctx, migrationsFS)

// 检查指定迁移是否已应用
applied, err := manager.IsMigrationApplied(ctx, "001_init_schema")
```

#### 回滚迁移

```go
// 简单回滚（仅删除记录）
if err := manager.RollbackMigration(ctx, "001_init_schema"); err != nil {
    log.Fatal(err)
}

// 使用回滚脚本回滚
rollbackSQL := "DROP TABLE IF EXISTS test_table;"
if err := manager.RollbackMigrationWithScript(ctx, "001_init_schema", rollbackSQL); err != nil {
    log.Fatal(err)
}
```

## 迁移脚本规范

### 命名规范

迁移脚本文件应按以下格式命名：

```
{序号}_{描述}.sql
```

例如：
- `001_init_schema.sql`
- `002_add_alarm_rules.sql`
- `003_add_qa_tables.sql`

### 脚本内容规范

```sql
-- 005_add_performance_indexes.sql
-- 添加性能优化索引

-- 为告警表添加复合索引
CREATE INDEX IF NOT EXISTS idx_alarms_station_status_triggered 
ON alarms(station_id, status, triggered_at DESC);

-- 添加注释
COMMENT ON INDEX idx_alarms_station_status_triggered IS '优化按厂站查询告警的查询性能';
```

### 最佳实践

1. **幂等性**: 使用 `IF NOT EXISTS` 确保脚本可重复执行
2. **注释**: 为每个操作添加清晰的注释
3. **索引命名**: 使用有意义的索引名称，格式为 `idx_{表名}_{字段名}`
4. **分区表**: 为大表使用分区提高查询性能
5. **回滚脚本**: 为每个迁移准备对应的回滚脚本

## 性能优化

### 索引优化

项目提供了以下类型的索引：

1. **复合索引**: 优化多字段查询
   ```sql
   CREATE INDEX idx_alarms_station_status_triggered 
   ON alarms(station_id, status, triggered_at DESC);
   ```

2. **部分索引**: 只索引满足条件的数据
   ```sql
   CREATE INDEX idx_alarms_active 
   ON alarms(station_id, level, triggered_at DESC) 
   WHERE status IN (1, 2);
   ```

3. **表达式索引**: 对计算结果建索引
   ```sql
   CREATE INDEX idx_users_lower_username 
   ON users(LOWER(username));
   ```

4. **覆盖索引**: 包含查询所需的所有字段
   ```sql
   CREATE INDEX idx_alarms_stats 
   ON alarms(station_id, level, status, triggered_at) 
   INCLUDE (title, message);
   ```

### 分区表

对于大表（如告警表、操作日志表），使用按月分区：

```sql
CREATE TABLE alarms_partitioned (
    id VARCHAR(36) NOT NULL,
    -- 其他字段
    triggered_at TIMESTAMP NOT NULL,
    PRIMARY KEY (id, triggered_at)
) PARTITION BY RANGE (triggered_at);

-- 创建分区
CREATE TABLE alarms_2024_01 PARTITION OF alarms_partitioned
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

自动维护分区：

```sql
-- 每月运行一次
SELECT maintain_partitions();
```

## 监控和运维

### 连接池监控

建议定期监控以下指标：

- **Open Connections**: 当前打开的连接数
- **In Use**: 正在使用的连接数
- **Idle**: 空闲连接数
- **Wait Count**: 等待连接的请求数
- **Wait Duration**: 等待连接的总时间

### 告警规则

当以下情况发生时应触发告警：

1. 连接池使用率 > 80%
2. Wait Count 持续增长
3. Wait Duration 过长
4. 数据库健康检查失败

### 运维建议

1. **定期备份**: 每天进行数据库备份
2. **索引维护**: 定期执行 `VACUUM ANALYZE`
3. **分区管理**: 每月创建新分区，删除旧分区
4. **慢查询分析**: 启用慢查询日志，定期优化

## 测试

运行单元测试：

```bash
# 运行所有测试
go test ./internal/infrastructure/persistence/... -v

# 运行集成测试（需要数据库）
go test ./internal/infrastructure/persistence/... -v -tags=integration

# 跳过集成测试
go test ./internal/infrastructure/persistence/... -v -short
```

## 示例程序

运行迁移示例：

```bash
cd cmd/migrate
go run main.go
```

## 配置文件

配置示例（`configs/config.yaml`）：

```yaml
database:
  type: postgres
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: nem_system
  sslmode: disable
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600    # 秒
  conn_max_idle_time: 600    # 秒
```

## 故障排查

### 连接失败

1. 检查数据库服务是否运行
2. 验证连接参数是否正确
3. 检查防火墙设置
4. 查看数据库日志

### 迁移失败

1. 检查迁移脚本语法
2. 验证数据库权限
3. 查看迁移记录表
4. 手动执行失败的脚本

### 性能问题

1. 检查索引是否生效
2. 分析慢查询日志
3. 监控连接池状态
4. 考虑使用分区表

## 参考资料

- [PostgreSQL 官方文档](https://www.postgresql.org/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [PostgreSQL 索引优化](https://www.postgresql.org/docs/current/indexes.html)
- [PostgreSQL 表分区](https://www.postgresql.org/docs/current/ddl-partitioning.html)
