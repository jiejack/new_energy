# 配置说明

本文档详细介绍新能源监控系统的配置项和配置方法。

## 目录

1. [配置文件结构](#1-配置文件结构)
2. [核心配置项](#2-核心配置项)
3. [环境变量覆盖](#3-环境变量覆盖)
4. [配置中心集成](#4-配置中心集成)
5. [多环境配置](#5-多环境配置)
6. [配置最佳实践](#6-配置最佳实践)

---

## 1. 配置文件结构

系统采用YAML格式的配置文件，位于 `configs/` 目录：

```
configs/
├── config.yaml              # 默认配置
├── config-dev.yaml          # 开发环境配置
├── config-test.yaml         # 测试环境配置
├── config-prod.yaml         # 生产环境配置
└── config-standalone.yaml   # 单机部署配置
```

### 配置文件加载顺序

1. 首先加载默认配置 `config.yaml`
2. 然后加载环境特定配置（如 `config-dev.yaml`）
3. 最后加载环境变量覆盖

## 2. 核心配置项

### 2.1 服务配置

```yaml
server:
  name: api-server           # 服务名称
  port: 8080                 # 服务端口
  mode: release              # 运行模式: debug/release
  graceful_shutdown: 30      # 优雅关闭超时时间(秒)
  cors:
    enabled: true            # 是否启用CORS
    allow_origins: ["*"]      # 允许的来源
  timeout:
    read: 10                 # 读超时(秒)
    write: 10                # 写超时(秒)
    idle: 60                 # 空闲超时(秒)
```

### 2.2 数据库配置

```yaml
database:
  type: postgres             # 数据库类型
  host: localhost            # 主机地址
  port: 5432                 # 端口
  user: postgres             # 用户名
  password: postgres         # 密码
  dbname: nem_system         # 数据库名
  sslmode: disable           # SSL模式
  max_open_conns: 100        # 最大连接数
  max_idle_conns: 10         # 最大空闲连接数
  conn_max_lifetime: 3600    # 连接最大生命周期(秒)
  conn_max_idle_time: 600    # 空闲连接最大生命周期(秒)
  migrations:
    auto: true               # 是否自动执行迁移
    path: ./scripts/migrations  # 迁移文件路径
```

### 2.3 Redis 配置

```yaml
redis:
  addrs:                     # 集群地址列表
    - localhost:6379
  password: ""               # 密码
  db: 0                      # 数据库索引
  pool_size: 100             # 连接池大小
  dial_timeout: 5            # 连接超时(秒)
  read_timeout: 3            # 读取超时(秒)
  write_timeout: 3           # 写入超时(秒)
  idle_timeout: 60           # 空闲超时(秒)
```

### 2.4 Kafka 配置

```yaml
kafka:
  brokers:                   # Broker地址列表
    - localhost:9092
  topic_prefix: nem          # 主题前缀
  consumer:
    group_id: nem-consumer   # 消费者组ID
    auto_offset_reset: latest  # 偏移量重置策略
    session_timeout: 30      # 会话超时(秒)
  producer:
    acks: all                # 确认级别
    retries: 3               # 重试次数
    batch_size: 16384        # 批处理大小
    linger_ms: 1             # linger时间(毫秒)
```

### 2.5 时序数据库配置

```yaml
timeseries:
  type: doris                # 类型: doris/clickhouse
  doris:
    hosts:
      - localhost:9030
    database: nem_ts
    user: root
    password: ""
    max_open_conns: 100
    max_idle_conns: 20
    conn_timeout: 10s
    write_timeout: 30s
    query_timeout: 60s
    batch_size: 10000
```

### 2.6 日志配置

```yaml
logging:
  level: info                # 日志级别: debug/info/warn/error
  format: json               # 格式: console/json
  output: stdout             # 输出: stdout/file
  file:
    path: /var/log/nem/app.log
    max_size: 100            # 单文件最大大小(MB)
    max_backups: 10          # 最大备份数
    max_age: 30              # 最大保留天数
    compress: true           # 是否压缩
  zap:
    development: false       # 是否开发模式
    sampling:
      initial: 100
      thereafter: 100
```

### 2.7 认证配置

```yaml
auth:
  jwt:
    secret: your-jwt-secret-key-change-in-production
    access_expire: 7200      # Access Token过期时间(秒)
    refresh_expire: 604800   # Refresh Token过期时间(秒)
    issuer: nem-system       # 令牌发行者
  password:
    min_length: 8            # 最小长度
    require_uppercase: true  # 需要大写字母
    require_lowercase: true  # 需要小写字母
    require_digit: true      # 需要数字
    require_special: false   # 需要特殊字符
  login:
    max_attempts: 5          # 最大尝试次数
    lock_duration: 1800      # 锁定时长(秒)
  session:
    cookie_name: nem_session
    cookie_secure: false      # 是否仅HTTPS
    cookie_http_only: true    # 是否仅HTTP
    cookie_same_site: lax     # SameSite策略
```

### 2.8 监控配置

```yaml
tracing:
  enabled: true              # 是否启用链路追踪
  endpoint: localhost:4317   # OTLP端点
  sampler_ratio: 0.1         # 采样率
  service_name: api-server

metrics:
  enabled: true              # 是否启用指标采集
  port: 9090                 # Prometheus指标端口
  path: /metrics             # 指标路径
  namespace: nem             # 指标命名空间

health:
  enabled: true              # 是否启用健康检查
  path: /health              # 健康检查路径
  interval: 30               # 检查间隔(秒)
```

### 2.9 采集配置

```yaml
collector:
  enabled: true              # 是否启用采集
  interval: 5                # 采集间隔(秒)
  batch_size: 1000           # 批处理大小
  buffer_size: 10000          # 缓冲区大小
  retry_count: 3              # 重试次数
  protocols:
    modbus:
      timeout: 3              # 超时(秒)
      retries: 2              # 重试次数
    iec104:
      timeout: 5              # 超时(秒)
      retry_interval: 10      # 重试间隔(秒)
```

### 2.10 告警配置

```yaml
alarm:
  enabled: true              # 是否启用告警
  check_interval: 10          # 检查间隔(秒)
  notification:
    enabled: true            # 是否启用通知
    channels:
      email: true            # 邮件通知
      sms: false             # 短信通知
      webhook: false         # Webhook通知
  rules:
    max_rules: 1000          # 最大规则数
    cache_ttl: 3600          # 缓存TTL(秒)
```

## 3. 环境变量覆盖

系统支持通过环境变量覆盖配置文件中的值。环境变量采用大写字母，下划线分隔的格式：

### 数据库配置

```bash
# 数据库配置
export DB_HOST=postgres.example.com
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=secure-password
export DB_NAME=nem_system
export DB_SSLMODE=disable
```

### Redis配置

```bash
# Redis配置
export REDIS_ADDR=redis.example.com:6379
export REDIS_PASSWORD=redis-password
export REDIS_DB=0
```

### Kafka配置

```bash
# Kafka配置
export KAFKA_BROKERS=kafka1:9092,kafka2:9092,kafka3:9092
export KAFKA_TOPIC_PREFIX=nem
```

### 服务配置

```bash
# 服务配置
export SERVER_PORT=8080
export SERVER_MODE=release
export SERVER_NAME=api-server
```

### 认证配置

```bash
# 认证配置
export JWT_SECRET=your-secret-key
export JWT_ACCESS_EXPIRE=7200
export JWT_REFRESH_EXPIRE=604800
```

## 4. 配置中心集成

系统支持集成外部配置中心，如Nacos：

### 4.1 Nacos 配置

```yaml
nacos:
  enabled: true
  server_addr: nacos.example.com:8848
  namespace: production
  group: DEFAULT_GROUP
  service_name: nem-api-server
  weight: 1
  metadata:
    version: 1.0.0
    env: production
  registry:
    enabled: true            # 是否注册服务
    interval: 30             # 注册间隔(秒)
  config:
    enabled: true            # 是否从配置中心获取配置
    data_id: nem-api-server.yaml
    refresh_interval: 30     # 刷新间隔(秒)
```

### 4.2 配置中心优先级

1. 环境变量
2. 配置中心配置
3. 本地配置文件
4. 默认配置

## 5. 多环境配置

### 5.1 开发环境 (config-dev.yaml)

```yaml
server:
  mode: debug
  port: 8080

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: nem_dev

logging:
  level: debug
  format: console

tracing:
  enabled: false

metrics:
  enabled: true
```

### 5.2 测试环境 (config-test.yaml)

```yaml
server:
  mode: release
  port: 8080

database:
  host: postgres
  port: 5432
  user: postgres
  password: postgres
  dbname: nem_test

logging:
  level: info
  format: json

tracing:
  enabled: true

metrics:
  enabled: true
```

### 5.3 生产环境 (config-prod.yaml)

```yaml
server:
  mode: release
  port: 8080
  cors:
    enabled: true
    allow_origins: ["https://your-domain.com"]

database:
  host: postgres
  port: 5432
  user: nem_app
  password: ${DB_PASSWORD}
  dbname: nem_system
  sslmode: require

redis:
  addrs:
    - redis:6379
  password: ${REDIS_PASSWORD}

logging:
  level: warn
  format: json
  output: file
  file:
    path: /var/log/nem/app.log

tracing:
  enabled: true
  endpoint: jaeger:4317

metrics:
  enabled: true
  port: 9090
```

### 5.4 单机部署 (config-standalone.yaml)

```yaml
server:
  mode: release
  port: 8080

database:
  type: sqlite
  path: ./data/nem.db

redis:
  enabled: false

kafka:
  enabled: false

collector:
  enabled: true
  interval: 10

alarm:
  enabled: true
  check_interval: 30
```

## 6. 配置最佳实践

### 6.1 安全最佳实践

1. **敏感信息处理**
   - 不要在配置文件中硬编码密码和密钥
   - 使用环境变量或配置中心存储敏感信息
   - 生产环境使用Kubernetes Secret管理敏感信息

2. **权限控制**
   - 配置文件权限设置为 644
   - 敏感配置文件权限设置为 600
   - 定期轮换密钥和密码

3. **配置加密**
   - 对敏感配置项进行加密
   - 使用KMS或其他加密服务管理密钥

### 6.2 性能最佳实践

1. **连接池配置**
   - 根据服务负载调整数据库连接池大小
   - 设置合理的连接超时和最大生命周期

2. **缓存配置**
   - 合理设置Redis缓存大小和淘汰策略
   - 对热点数据使用缓存

3. **日志配置**
   - 生产环境使用json格式日志
   - 合理设置日志级别，避免过多日志影响性能

### 6.3 运维最佳实践

1. **配置版本管理**
   - 使用Git管理配置文件
   - 为不同环境创建配置分支
   - 配置变更需要代码审查

2. **配置备份**
   - 定期备份配置文件
   - 保存配置变更历史

3. **配置验证**
   - 部署前验证配置文件格式
   - 使用配置校验工具检查配置有效性

4. **监控配置**
   - 监控配置文件变更
   - 对关键配置项设置监控告警

### 6.4 常见配置问题

**问题1：数据库连接失败**
- 检查数据库地址和端口
- 验证用户名和密码
- 确认网络连接
- 检查数据库服务状态

**问题2：Redis连接失败**
- 检查Redis地址和端口
- 验证密码
- 确认Redis服务状态
- 检查连接池配置

**问题3：Kafka消息积压**
- 检查消费者组配置
- 增加消费者实例
- 调整批处理大小
- 检查网络连接

**问题4：配置不生效**
- 检查配置文件路径
- 验证配置文件格式
- 检查环境变量覆盖
- 查看应用启动日志

## 7. 配置示例

### 7.1 完整配置示例

```yaml
# config.yaml
server:
  name: api-server
  port: 8080
  mode: release
  graceful_shutdown: 30
  cors:
    enabled: true
    allow_origins: ["*"]

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
  conn_max_lifetime: 3600
  conn_max_idle_time: 600

redis:
  addrs:
    - localhost:6379
  password: ""
  db: 0
  pool_size: 100

kafka:
  brokers:
    - localhost:9092
  topic_prefix: nem

logging:
  level: info
  format: json
  output: stdout

auth:
  jwt:
    secret: your-jwt-secret-key
    access_expire: 7200
    refresh_expire: 604800

metrics:
  enabled: true
  port: 9090

tracing:
  enabled: false

collector:
  enabled: true
  interval: 5

alarm:
  enabled: true
  check_interval: 10
```

### 7.2 生产环境配置示例

```yaml
# config-prod.yaml
server:
  name: api-server
  port: 8080
  mode: release

database:
  host: ${DB_HOST}
  port: ${DB_PORT}
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  dbname: ${DB_NAME}
  sslmode: require

redis:
  addrs:
    - ${REDIS_ADDR}
  password: ${REDIS_PASSWORD}
  db: 0

kafka:
  brokers: ${KAFKA_BROKERS}
  topic_prefix: nem

logging:
  level: warn
  format: json
  output: file
  file:
    path: /var/log/nem/app.log

auth:
  jwt:
    secret: ${JWT_SECRET}
    access_expire: 7200
    refresh_expire: 604800

metrics:
  enabled: true
  port: 9090

tracing:
  enabled: true
  endpoint: jaeger:4317

collector:
  enabled: true
  interval: 5

alarm:
  enabled: true
  check_interval: 10
  notification:
    enabled: true
    channels:
      email: true
```

## 8. 故障排查

### 8.1 配置相关问题

| 问题 | 可能原因 | 解决方案 |
|------|----------|----------|
| 配置文件解析错误 | YAML格式错误 | 检查配置文件语法，使用YAML验证工具 |
| 环境变量未生效 | 变量名错误或未导出 | 检查环境变量名称，确保已正确导出 |
| 配置中心连接失败 | 网络问题或配置错误 | 检查网络连接，验证配置中心地址和认证信息 |
| 敏感信息泄露 | 配置文件权限不当 | 设置正确的文件权限，使用环境变量或Secret |

### 8.2 调试配置

```bash
# 启用调试模式
export SERVER_MODE=debug

# 查看配置加载过程
export LOG_LEVEL=debug

# 验证配置文件
./scripts/verify-config.sh

# 查看当前生效的配置
curl http://localhost:8080/debug/config
```

## 9. 总结

合理的配置管理是系统稳定运行的关键。本文档提供了新能源监控系统的完整配置指南，包括：

- 详细的配置文件结构和配置项说明
- 环境变量覆盖机制
- 配置中心集成方案
- 多环境配置策略
- 配置最佳实践
- 常见问题排查

通过遵循这些配置指南，可以确保系统在不同环境下的稳定运行，并为后续的维护和扩展提供良好的基础。