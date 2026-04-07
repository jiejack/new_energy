# 新能源在线监控系统 - 运维手册

## 文档信息

| 项目 | 内容 |
|------|------|
| 项目名称 | 新能源在线监控系统 |
| 文档版本 | v1.0.0 |
| 编写日期 | 2026-04-07 |
| 文档状态 | 正式发布 |
| 维护团队 | 运维团队 |

---

## 目录

1. [系统部署指南](#1-系统部署指南)
2. [配置说明](#2-配置说明)
3. [监控告警配置](#3-监控告警配置)
4. [故障排查指南](#4-故障排查指南)
5. [备份恢复流程](#5-备份恢复流程)
6. [安全加固指南](#6-安全加固指南)
7. [Harness 验证框架使用说明](#7-harness-验证框架使用说明)
8. [附录](#8-附录)

---

## 1. 系统部署指南

### 1.1 环境准备

#### 1.1.1 系统要求

**操作系统**

| 操作系统 | 版本要求 | 架构 |
|----------|----------|------|
| CentOS | 7.6+ | x86_64 |
| Ubuntu | 20.04 LTS+ | x86_64 |
| Debian | 11+ | x86_64 |

**硬件要求**

| 组件 | 最低配置 | 推荐配置 |
|------|----------|----------|
| CPU | 4核 | 8核+ |
| 内存 | 16GB | 32GB+ |
| 磁盘 | 500GB SSD | 1TB+ SSD |
| 网络 | 1Gbps | 10Gbps |

**生产环境资源规划**

| 服务 | CPU | 内存 | 磁盘 | 实例数 |
|------|-----|------|------|--------|
| api-server | 2核 | 4GB | - | 2-10 |
| collector | 4核 | 8GB | - | 3-20 |
| alarm | 2核 | 4GB | - | 2-5 |
| compute | 2核 | 4GB | - | 2-5 |
| ai-service | 4核 | 16GB | - | 2-5 |
| scheduler | 2核 | 4GB | - | 1-3 |
| PostgreSQL | 4核 | 16GB | 500GB | 2 |
| Redis | 2核 | 8GB | 50GB | 3 |
| Kafka | 2核 | 8GB | 200GB | 3 |
| Doris FE | 4核 | 16GB | 100GB | 3 |
| Doris BE | 8核 | 32GB | 1TB | 3 |

#### 1.1.2 软件依赖

**必需软件**

| 软件 | 版本要求 | 说明 |
|------|----------|------|
| Docker | 24.0+ | 容器运行时 |
| Docker Compose | 2.20+ | 容器编排工具 |
| Kubernetes | 1.28+ | 容器编排平台（生产环境） |
| Helm | 3.12+ | Kubernetes包管理器 |
| Git | 2.40+ | 版本控制 |

**可选软件**

| 软件 | 版本要求 | 说明 |
|------|----------|------|
| Nacos | 2.2+ | 服务注册与配置中心 |
| Prometheus | 2.45+ | 监控系统 |
| Grafana | 10.0+ | 可视化平台 |

#### 1.1.3 网络要求

**端口规划**

| 服务 | 端口 | 协议 | 说明 |
|------|------|------|------|
| api-server | 8080 | HTTP | API服务端口 |
| api-server | 9090 | HTTP | Prometheus指标端口 |
| PostgreSQL | 5432 | TCP | 数据库端口 |
| Redis | 6379 | TCP | 缓存端口 |
| Kafka | 9092 | TCP | 消息队列端口 |
| Doris FE | 9030 | HTTP | Doris FE HTTP端口 |
| Prometheus | 9091 | HTTP | 监控端口 |
| Grafana | 3000 | HTTP | 可视化端口 |

**防火墙配置**

```bash
# 开放必要端口
firewall-cmd --permanent --add-port=8080/tcp
firewall-cmd --permanent --add-port=5432/tcp
firewall-cmd --permanent --add-port=6379/tcp
firewall-cmd --permanent --add-port=9092/tcp
firewall-cmd --permanent --add-port=9030/tcp
firewall-cmd --reload
```

### 1.2 Docker 部署

#### 1.2.1 快速开始

```bash
# 克隆代码
git clone https://github.com/new-energy-monitoring/new-energy-monitoring.git
cd new-energy-monitoring

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f api-server
```

#### 1.2.2 镜像构建

```bash
# 使用Makefile构建
make docker-build

# 或手动构建
docker build -t new-energy-monitoring/api-server:1.0.0 \
  -f deployments/docker/Dockerfile \
  --build-arg SERVICE=api-server .
```

#### 1.2.3 常用命令

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 重启服务
docker-compose restart api-server

# 查看日志
docker-compose logs -f api-server

# 进入容器
docker-compose exec api-server sh

# 查看资源使用
docker stats

# 清理无用镜像
docker image prune -a
```

### 1.3 Kubernetes 部署

#### 1.3.1 前置条件

- Kubernetes版本：1.28+
- 节点数量：≥3（生产环境）
- 存储类：支持动态供给
- Ingress控制器：nginx-ingress

#### 1.3.2 使用 Helm 部署

```bash
# 创建命名空间
kubectl create namespace nem-system

# 安装Chart
helm install nem-system ./deployments/kubernetes/helm \
  -n nem-system \
  -f values-prod.yaml

# 升级Chart
helm upgrade nem-system ./deployments/kubernetes/helm \
  -n nem-system \
  -f values-prod.yaml

# 卸载Chart
helm uninstall nem-system -n nem-system
```

#### 1.3.3 常用 kubectl 命令

```bash
# 查看Pod状态
kubectl get pods -n nem-system

# 查看Pod详情
kubectl describe pod <pod-name> -n nem-system

# 查看日志
kubectl logs -f <pod-name> -n nem-system

# 进入容器
kubectl exec -it <pod-name> -n nem-system -- sh

# 查看服务
kubectl get svc -n nem-system

# 查看Ingress
kubectl get ingress -n nem-system

# 扩缩容
kubectl scale deployment api-server --replicas=5 -n nem-system

# 重启Pod
kubectl rollout restart deployment api-server -n nem-system

# 查看资源使用
kubectl top pods -n nem-system
```

---

## 2. 配置说明

### 2.1 配置文件结构

```
configs/
├── config.yaml              # 默认配置
├── config-dev.yaml          # 开发环境配置
├── config-test.yaml         # 测试环境配置
├── config-prod.yaml         # 生产环境配置
└── config-standalone.yaml   # 单机部署配置
```

### 2.2 核心配置项

#### 2.2.1 服务配置

```yaml
server:
  name: api-server           # 服务名称
  port: 8080                 # 服务端口
  mode: release              # 运行模式: debug/release
  graceful_shutdown: 30      # 优雅关闭超时时间(秒)
```

#### 2.2.2 数据库配置

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
```

#### 2.2.3 Redis 配置

```yaml
redis:
  addrs:                     # 集群地址列表
    - localhost:6379
  password: ""               # 密码
  db: 0                      # 数据库索引
  pool_size: 100             # 连接池大小
```

#### 2.2.4 Kafka 配置

```yaml
kafka:
  brokers:                   # Broker地址列表
    - localhost:9092
  topic_prefix: nem          # 主题前缀
```

#### 2.2.5 时序数据库配置

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

#### 2.2.6 日志配置

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
```

#### 2.2.7 认证配置

```yaml
auth:
  jwt:
    secret: your-jwt-secret-key-change-in-production
    access_expire: 7200      # Access Token过期时间(秒)
    refresh_expire: 604800   # Refresh Token过期时间(秒)
  password:
    min_length: 8            # 最小长度
    require_uppercase: true  # 需要大写字母
    require_lowercase: true  # 需要小写字母
    require_digit: true      # 需要数字
  login:
    max_attempts: 5          # 最大尝试次数
    lock_duration: 1800      # 锁定时长(秒)
```

#### 2.2.8 监控配置

```yaml
tracing:
  enabled: true              # 是否启用链路追踪
  endpoint: localhost:4317   # OTLP端点
  sampler_ratio: 0.1         # 采样率

metrics:
  enabled: true              # 是否启用指标采集
  port: 9090                 # Prometheus指标端口
```

### 2.3 环境变量覆盖

支持通过环境变量覆盖配置文件中的值：

```bash
# 数据库配置
export DB_HOST=postgres.example.com
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=secure-password

# Redis配置
export REDIS_ADDR=redis.example.com:6379
export REDIS_PASSWORD=redis-password

# Kafka配置
export KAFKA_BROKERS=kafka1:9092,kafka2:9092,kafka3:9092
```

### 2.4 配置中心集成

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

config_center:
  enabled: true
  provider: nacos
  server_addr: nacos.example.com:8848
  namespace: production
  group: DEFAULT_GROUP
  data_id: nem-api-server.yaml
  refresh_interval: 30
```

---

## 3. 监控告警配置

### 3.1 Prometheus 配置

#### 3.1.1 prometheus.yml

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

alerting:
  alertmanagers:
  - static_configs:
    - targets:
      - alertmanager:9093

rule_files:
  - /etc/prometheus/rules/*.yml

scrape_configs:
  # API服务
  - job_name: 'api-server'
    kubernetes_sd_configs:
    - role: pod
      namespaces:
        names:
        - nem-system
    relabel_configs:
    - source_labels: [__meta_kubernetes_pod_label_app]
      action: keep
      regex: api-server
    - source_labels: [__meta_kubernetes_pod_ip]
      target_label: __address__
      replacement: ${1}:9090

  # PostgreSQL
  - job_name: 'postgres'
    static_configs:
    - targets: ['postgres-exporter:9187']

  # Redis
  - job_name: 'redis'
    static_configs:
    - targets: ['redis-exporter:9121']

  # Kafka
  - job_name: 'kafka'
    static_configs:
    - targets: ['kafka-exporter:9308']
```

### 3.2 告警规则

#### 3.2.1 rules/alerts.yml

```yaml
groups:
- name: nem-alerts
  rules:
  # 服务可用性告警
  - alert: ServiceDown
    expr: up == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "服务 {{ $labels.job }} 实例 {{ $labels.instance }} 已下线"
      description: "服务已下线超过1分钟"

  # 高CPU使用率
  - alert: HighCPUUsage
    expr: rate(process_cpu_seconds_total[5m]) > 0.8
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "服务 {{ $labels.job }} CPU使用率过高"
      description: "CPU使用率超过80%，当前值: {{ $value }}"

  # 高内存使用率
  - alert: HighMemoryUsage
    expr: process_resident_memory_bytes / 1024 / 1024 / 1024 > 1.5
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "服务 {{ $labels.job }} 内存使用过高"
      description: "内存使用超过1.5GB，当前值: {{ $value }}GB"

  # HTTP错误率过高
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.01
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "API错误率过高"
      description: "HTTP 5xx错误率超过1%，当前值: {{ $value }}"

  # 响应时间过长
  - alert: SlowResponse
    expr: histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "API响应时间过长"
      description: "P99响应时间超过1秒，当前值: {{ $value }}秒"

  # 数据库连接数过高
  - alert: HighDBConnections
    expr: pg_stat_activity_count / pg_settings_max_connections > 0.8
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "数据库连接数过高"
      description: "数据库连接数超过80%，当前值: {{ $value }}"

  # Kafka消费延迟
  - alert: KafkaConsumerLag
    expr: kafka_consumer_lag > 10000
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Kafka消费延迟过高"
      description: "消费延迟超过10000条，当前值: {{ $value }}"
```

### 3.3 Alertmanager 配置

#### 3.3.1 alertmanager.yml

```yaml
global:
  resolve_timeout: 5m
  smtp_smarthost: 'smtp.example.com:587'
  smtp_from: 'alertmanager@example.com'
  smtp_auth_username: 'alertmanager@example.com'
  smtp_auth_password: 'password'

route:
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  receiver: 'default-receiver'
  routes:
  - match:
      severity: critical
    receiver: 'critical-receiver'
  - match:
      severity: warning
    receiver: 'warning-receiver'

receivers:
- name: 'default-receiver'
  email_configs:
  - to: 'team@example.com'

- name: 'critical-receiver'
  email_configs:
  - to: 'oncall@example.com'
  webhook_configs:
  - url: 'http://webhook.example.com/alert'
    send_resolved: true

- name: 'warning-receiver'
  email_configs:
  - to: 'team@example.com'

inhibit_rules:
- source_match:
    severity: 'critical'
  target_match:
    severity: 'warning'
  equal: ['alertname', 'instance']
```

### 3.4 Grafana Dashboard

#### 3.4.1 关键监控面板

1. **系统概览Dashboard**
   - 服务状态
   - 请求量统计
   - 错误率趋势
   - 响应时间分布

2. **资源监控Dashboard**
   - CPU使用率
   - 内存使用率
   - 磁盘I/O
   - 网络流量

3. **业务监控Dashboard**
   - 设备在线率
   - 数据采集量
   - 告警数量
   - 站点状态

4. **数据库监控Dashboard**
   - 连接数
   - 查询性能
   - 慢查询统计
   - 锁等待

### 3.5 常用监控查询

```promql
# API请求速率
rate(http_requests_total[5m])

# API响应时间P99
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

# 错误率
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])

# CPU使用率
rate(process_cpu_seconds_total[5m])

# 内存使用
process_resident_memory_bytes / 1024 / 1024 / 1024

# 数据库连接数
pg_stat_activity_count

# Redis连接数
redis_connected_clients

# Kafka消费延迟
kafka_consumer_lag
```

---

## 4. 故障排查指南

### 4.1 服务启动故障

#### 4.1.1 服务无法启动

**症状**:
- 容器启动后立即退出
- 服务状态为CrashLoopBackOff
- 日志显示启动错误

**排查步骤**:

```bash
# 1. 查看Pod状态
kubectl get pods -n nem-system

# 2. 查看Pod详情
kubectl describe pod <pod-name> -n nem-system

# 3. 查看日志
kubectl logs <pod-name> -n nem-system --previous

# 4. 检查配置
kubectl get configmap -n nem-system
kubectl describe configmap <configmap-name> -n nem-system

# 5. 检查资源限制
kubectl top pods -n nem-system
kubectl describe resourcequota -n nem-system
```

**常见原因及解决方案**:

| 原因 | 解决方案 |
|------|----------|
| 配置文件错误 | 检查config.yaml语法，验证必需配置项 |
| 环境变量缺失 | 检查Secret和环境变量配置 |
| 依赖服务未就绪 | 确保数据库、Redis等服务已启动 |
| 资源不足 | 增加资源限制或释放资源 |
| 镜像拉取失败 | 检查镜像是否存在、网络连接、镜像仓库认证 |
| 端口冲突 | 检查端口是否被占用 |

### 4.2 数据库故障

#### 4.2.1 数据库连接失败

**症状**:
- 应用日志显示数据库连接错误
- API返回500错误
- 数据库相关操作失败

**排查步骤**:

```bash
# 1. 检查数据库服务状态
docker-compose exec postgres pg_isready -U postgres

# 2. 测试数据库连接
docker-compose exec postgres psql -U postgres -d nem_system

# 3. 检查连接数
docker-compose exec postgres psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

# 4. 检查最大连接数
docker-compose exec postgres psql -U postgres -c "SHOW max_connections;"
```

**常见错误及解决方案**:

**错误1：连接数超限**

```
FATAL: sorry, too many clients already
```

解决方案：

```bash
# 临时增加连接数
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -c "ALTER SYSTEM SET max_connections = 300;"
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -c "SELECT pg_reload_conf();"

# 终止空闲连接
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'idle' AND pid <> pg_backend_pid();"
```

**错误2：认证失败**

```
FATAL: password authentication failed for user "postgres"
```

解决方案：

```bash
# 检查密码配置
kubectl get secret nem-secrets -n nem-system -o jsonpath='{.data.db-password}' | base64 -d

# 重置密码
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -c "ALTER USER postgres WITH PASSWORD 'new-password';"

# 更新Secret
kubectl create secret generic nem-secrets \
  --from-literal=db-password=new-password \
  --dry-run=client -o yaml | kubectl apply -f - -n nem-system

# 重启应用
kubectl rollout restart deployment/api-server -n nem-system
```

#### 4.2.2 慢查询问题

**排查步骤**:

```sql
-- 开启慢查询日志（超过1秒）
ALTER SYSTEM SET log_min_duration_statement = 1000;
SELECT pg_reload_conf();

-- 查看最慢的10条SQL
SELECT
  query,
  calls,
  total_time,
  mean_time,
  max_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- 分析查询计划
EXPLAIN ANALYZE SELECT * FROM devices WHERE station_id = 'station-001';
```

**解决方案**:

```sql
-- 创建索引
CREATE INDEX idx_devices_station_id ON devices(station_id);
CREATE INDEX idx_alarms_created_at ON alarms(created_at);

-- 创建复合索引
CREATE INDEX idx_alarms_station_time ON alarms(station_id, created_at);
```

### 4.3 Redis 故障

#### 4.3.1 Redis连接失败

**排查步骤**:

```bash
# 检查Redis服务状态
docker-compose exec redis redis-cli ping

# 测试连接
redis-cli -h redis -p 6379 ping

# 查看Redis信息
redis-cli INFO

# 查看连接数
redis-cli INFO clients
```

**常见错误及解决方案**:

**错误1：连接数超限**

```
ERR max number of clients reached
```

```bash
# 查看最大连接数
redis-cli CONFIG GET maxclients

# 增加最大连接数
redis-cli CONFIG SET maxclients 10000
```

**错误2：内存不足**

```
OOM command not allowed when used memory > 'maxmemory'
```

```bash
# 查看内存使用
redis-cli INFO memory

# 增加最大内存
redis-cli CONFIG SET maxmemory 4gb

# 设置内存淘汰策略
redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

### 4.4 Kafka 故障

#### 4.4.1 Kafka连接失败

**排查步骤**:

```bash
# 检查Kafka状态
docker-compose exec kafka kafka-broker-api-versions.sh -bootstrap-server localhost:9092

# 查看主题列表
docker-compose exec kafka kafka-topics.sh --list --bootstrap-server localhost:9092

# 查看消费者组
docker-compose exec kafka kafka-consumer-groups.sh --bootstrap-server localhost:9092 --list

# 查看消费者组详情
docker-compose exec kafka kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe \
  --group nem-group
```

#### 4.4.2 消息积压

**解决方案**:

```bash
# 方案1：增加消费者实例
kubectl scale deployment collector --replicas=10 -n nem-system

# 方案2：增加分区数
docker-compose exec kafka kafka-topics.sh \
  --alter \
  --topic nem.data.collect \
  --partitions 24 \
  --bootstrap-server localhost:9092
```

### 4.5 API 性能问题

#### 4.5.1 API响应慢

**排查步骤**:

```bash
# 1. 查看API指标
curl http://localhost:9090/metrics | grep http_request_duration

# 2. 分析链路追踪
# 访问Jaeger UI: http://localhost:16686

# 3. 查看应用日志
kubectl logs deployment/api-server -n nem-system | grep "slow request"

# 4. 性能分析
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
```

**解决方案**:

1. **添加缓存**

```go
// 使用Redis缓存
func (s *Service) GetDevice(id string) (*Device, error) {
    // 先从缓存获取
    cached, err := s.redis.Get(ctx, "device:"+id).Result()
    if err == nil {
        var device Device
        json.Unmarshal([]byte(cached), &device)
        return &device, nil
    }

    // 从数据库获取
    device, err := s.repo.GetDevice(id)
    if err != nil {
        return nil, err
    }

    // 写入缓存
    data, _ := json.Marshal(device)
    s.redis.Set(ctx, "device:"+id, data, time.Hour)

    return device, nil
}
```

2. **数据库查询优化**

```sql
-- 添加索引
CREATE INDEX idx_devices_station_status ON devices(station_id, status);

-- 使用覆盖索引
CREATE INDEX idx_devices_covering ON devices(station_id, status) INCLUDE (name, type);
```

#### 4.5.2 内存泄漏

**排查步骤**:

```bash
# 监控内存使用
watch -n 1 'kubectl top pods -n nem-system'

# 获取内存profile
curl http://localhost:8080/debug/pprof/heap > heap.prof

# 分析内存
go tool pprof heap.prof
(pprof) top10
(pprof) list functionName
```

---

## 5. 备份恢复流程

### 5.1 数据备份策略

| 数据类型 | 备份方式 | 频率 | 保留周期 | 存储位置 |
|----------|----------|------|----------|----------|
| PostgreSQL | 全量备份 | 每日 | 30天 | /backup/postgres |
| PostgreSQL | WAL归档 | 实时 | 7天 | /backup/wal |
| Redis | RDB快照 | 每小时 | 7天 | /backup/redis |
| Doris | 快照备份 | 每日 | 30天 | /backup/doris |
| 配置文件 | Git版本 | 实时 | 永久 | Git仓库 |

### 5.2 PostgreSQL 备份

#### 5.2.1 手动备份

```bash
#!/bin/bash
# backup_postgres.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/postgres"
DB_NAME="nem_system"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 全量备份
docker-compose exec -T postgres pg_dump \
  -U postgres \
  -F c \
  -f $BACKUP_DIR/${DB_NAME}_${DATE}.dump \
  $DB_NAME

# 或SQL格式
docker-compose exec -T postgres pg_dump \
  -U postgres \
  $DB_NAME | gzip > $BACKUP_DIR/${DB_NAME}_${DATE}.sql.gz

# 删除30天前的备份
find $BACKUP_DIR -name "*.dump" -mtime +30 -delete
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete

echo "Backup completed: ${DB_NAME}_${DATE}"
```

#### 5.2.2 自动备份（Kubernetes）

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
  namespace: nem-system
spec:
  schedule: "0 2 * * *"  # 每天凌晨2点
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:16-alpine
            command:
            - /bin/sh
            - -c
            - |
              pg_dump -U postgres -F c -f /backup/nem_system_$(date +%Y%m%d).dump nem_system
              find /backup -name "*.dump" -mtime +30 -delete
            env:
            - name: PGHOST
              value: postgres
            - name: PGUSER
              value: postgres
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: nem-secrets
                  key: db-password
            volumeMounts:
            - name: backup
              mountPath: /backup
          volumes:
          - name: backup
            persistentVolumeClaim:
              claimName: backup-pvc
          restartPolicy: OnFailure
```

#### 5.2.3 数据恢复

```bash
# 从dump文件恢复
docker-compose exec -T postgres pg_restore \
  -U postgres \
  -d nem_system \
  -c \
  /backup/nem_system_20240301.dump

# 从SQL文件恢复
gunzip < /backup/nem_system_20240301.sql.gz | \
  docker-compose exec -T postgres psql -U postgres nem_system

# Kubernetes环境
kubectl exec -it postgres-pod -n nem-system -- \
  pg_restore -U postgres -d nem_system /backup/nem_system_20240301.dump
```

### 5.3 Redis 备份

#### 5.3.1 手动备份

```bash
#!/bin/bash
# backup_redis.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/redis"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 触发RDB快照
docker-compose exec redis redis-cli BGSAVE

# 等待快照完成
sleep 5

# 复制RDB文件
docker cp nem-redis:/data/dump.rdb $BACKUP_DIR/dump_${DATE}.rdb

# 删除7天前的备份
find $BACKUP_DIR -name "*.rdb" -mtime +7 -delete

echo "Redis backup completed: dump_${DATE}.rdb"
```

#### 5.3.2 数据恢复

```bash
# 1. 停止Redis服务
docker-compose stop redis

# 2. 复制备份文件
docker cp $BACKUP_DIR/dump_20240301.rdb nem-redis:/data/dump.rdb

# 3. 启动Redis服务
docker-compose start redis

# 4. 验证数据
docker-compose exec redis redis-cli DBSIZE
```

### 5.4 定时备份配置

```bash
# crontab -e

# PostgreSQL每日备份 (凌晨2点)
0 2 * * * /opt/scripts/backup_postgres.sh >> /var/log/backup.log 2>&1

# Redis每小时备份
0 * * * * /opt/scripts/backup_redis.sh >> /var/log/backup.log 2>&1

# 配置文件备份 (每天)
0 3 * * * /opt/scripts/backup_configs.sh >> /var/log/backup.log 2>&1
```

---

## 6. 安全加固指南

### 6.1 网络安全

#### 6.1.1 防火墙配置

```bash
# 配置防火墙规则
firewall-cmd --permanent --add-rich-rule='rule family="ipv4" source address="10.0.0.0/8" port protocol="tcp" port="5432" accept'
firewall-cmd --permanent --add-rich-rule='rule family="ipv4" source address="10.0.0.0/8" port protocol="tcp" port="6379" accept'
firewall-cmd --reload
```

#### 6.1.2 网络策略（Kubernetes）

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: nem-network-policy
  namespace: nem-system
spec:
  podSelector:
    matchLabels:
      app: api-server
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: frontend
    - podSelector:
        matchLabels:
          app: collector
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
```

### 6.2 数据库安全

#### 6.2.1 创建应用专用用户

```sql
-- 创建应用专用用户
CREATE USER nem_app WITH PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE nem_system TO nem_app;
GRANT USAGE ON SCHEMA public TO nem_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO nem_app;

-- 禁用超级用户远程登录
-- 编辑 pg_hba.conf
```

### 6.3 密钥管理

#### 6.3.1 使用 Kubernetes Secret

```bash
# 创建Secret
kubectl create secret generic nem-secrets \
  --from-literal=db-password=secure-password \
  --from-literal=redis-password=secure-password \
  --from-literal=jwt-secret=secure-jwt-secret \
  -n nem-system

# 或使用外部密钥管理系统
# HashiCorp Vault, AWS Secrets Manager等
```

#### 6.3.2 密钥轮换

```bash
#!/bin/bash
# rotate_secrets.sh

# 生成新的JWT密钥
NEW_JWT_SECRET=$(openssl rand -base64 32)

# 更新Kubernetes Secret
kubectl create secret generic nem-secrets \
  --from-literal=jwt-secret=$NEW_JWT_SECRET \
  --dry-run=client -o yaml | kubectl apply -f - -n nem-system

# 重启服务
kubectl rollout restart deployment/api-server -n nem-system

echo "JWT secret rotated successfully"
```

### 6.4 安全扫描

#### 6.4.1 已配置的安全扫描工具

1. **go vet** - Go 静态代码分析
2. **golangci-lint (gosec)** - Go 安全检查
3. **nancy** - 依赖漏洞扫描
4. **gitleaks** - 敏感信息泄露检测

#### 6.4.2 运行安全扫描

```bash
# Linux/macOS
chmod +x scripts/security-audit.sh
./scripts/security-audit.sh

# Windows (Git Bash)
bash scripts/security-audit.sh
```

#### 6.4.3 高优先级安全问题修复

**问题1：TLS InsecureSkipVerify 可能为 true**

```go
// 生产环境必须启用证书验证
InsecureSkipVerify: false
```

**问题2：使用弱随机数生成器**

```go
import "crypto/rand"
// 替换 math/rand 为 crypto/rand
```

**问题3：潜在的 Slowloris 攻击**

```go
server := &http.Server{
    Addr:              ":8080",
    Handler:           mux,
    ReadHeaderTimeout: 10 * time.Second,
}
```

**问题4：SQL 字符串格式化**

```go
// 使用参数化查询替代字符串拼接
query := "SELECT * FROM table WHERE id = ?"
rows, err := db.Query(query, id)
```

### 6.5 RBAC 配置

```yaml
# 创建ServiceAccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: nem-service-account
  namespace: nem-system

---
# 创建Role
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: nem-role
  namespace: nem-system
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["pods/log"]
  verbs: ["get"]

---
# 创建RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: nem-role-binding
  namespace: nem-system
subjects:
- kind: ServiceAccount
  name: nem-service-account
  namespace: nem-system
roleRef:
  kind: Role
  name: nem-role
  apiGroup: rbac.authorization.k8s.io
```

---

## 7. Harness 验证框架使用说明

### 7.1 概述

Harness 层是新能源监控系统的验证框架，提供输入验证、输出验证、约束检查和监控功能。它已成功集成到告警模块，确保数据的完整性和一致性。

### 7.2 核心组件

#### 7.2.1 Harness 主入口

```go
type Harness struct {
    validator  Validator      // 输入验证器
    verifier   Verifier       // 输出验证器
    constraint Constraint     // 约束检查器
    monitor    Monitor        // 监控器
    snapshot   *SnapshotManager // 快照管理器
}
```

#### 7.2.2 组件说明

1. **Validator（验证器）**: 验证输入数据的格式和业务规则
2. **Verifier（验证器）**: 验证输出数据的正确性
3. **Constraint（约束器）**: 检查业务约束条件
4. **Monitor（监控器）**: 记录验证指标和事件
5. **SnapshotManager（快照管理器）**: 创建和管理测试快照

### 7.3 基本使用

#### 7.3.1 创建 Harness 实例

```go
// 创建默认 Harness 实例
harness := harness.NewHarness()

// 或使用自定义组件创建
customHarness := harness.NewHarnessWithComponents(
    harness.NewDefaultValidator(),
    harness.NewDefaultVerifier(),
    harness.NewDefaultConstraint(),
    harness.NewDefaultMonitor(),
)
```

#### 7.3.2 执行验证

```go
ctx := context.Background()

// 执行输入验证
if err := harness.Validate(ctx, input); err != nil {
    log.Printf("验证失败: %v", err)
    return
}

// 执行输出验证
match, err := harness.Verify(ctx, expected, actual)
if err != nil {
    log.Printf("验证失败: %v", err)
    return
}
if !match {
    log.Println("输出不匹配")
}
```

#### 7.3.3 快照管理

```go
// 创建快照
snapshot, err := harness.Snapshot(ctx, target)
if err != nil {
    log.Printf("创建快照失败: %v", err)
    return
}

// 保存快照
if err := harness.SaveSnapshot("snapshot-001", snapshot); err != nil {
    log.Printf("保存快照失败: %v", err)
    return
}

// 加载快照
loaded, err := harness.LoadSnapshot("snapshot-001")
if err != nil {
    log.Printf("加载快照失败: %v", err)
    return
}

// 比较快照
match, err := harness.CompareSnapshot("snapshot-001", newData)
if err != nil {
    log.Printf("比较快照失败: %v", err)
    return
}
```

### 7.4 告警模块集成示例

#### 7.4.1 创建告警验证

```go
// 创建 AlarmHarness 实例
alarmHarness := NewAlarmHarness()
ctx := context.Background()

// 验证创建告警请求
req := &CreateAlarmRequest{
    PointID:   "point001",
    DeviceID:  "device001",
    StationID: "station001",
    Type:      entity.AlarmTypeLimit,
    Level:     entity.AlarmLevelWarning,
    Title:     "电压高限告警",
    Value:     450.0,
    Threshold: 400.0,
}

if err := alarmHarness.ValidateCreateAlarm(ctx, req); err != nil {
    log.Printf("验证失败: %v", err)
    return
}
```

#### 7.4.2 集成服务使用

```go
// 创建带有 Harness 验证的告警服务
service := NewAlarmServiceWithHarness(alarmRepo)

// 创建告警（会自动进行验证）
alarm, err := service.CreateAlarm(ctx, req)
if err != nil {
    log.Printf("创建告警失败: %v", err)
    return
}

// 验证告警状态
if err := service.VerifyAlarmState(ctx, alarm.ID, entity.AlarmStateActive); err != nil {
    log.Printf("状态验证失败: %v", err)
    return
}
```

### 7.5 性能基准

#### 7.5.1 快照管理器性能

| 测试项 | 执行时间 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) |
|--------|------------------|-----------------|---------------------|
| SnapshotManager_Save | 559.0 | 192 | 3 |
| SnapshotManager_Load | 17.38 | 0 | 0 |
| SnapshotManager_Compare | 700.1 | 128 | 2 |
| CalculateChecksum | 453.1 | 128 | 2 |

#### 7.5.2 性能优化建议

1. **短期优化**:
   - 实现 Snapshot 对象池减少内存分配
   - 优化 Checksum 计算

2. **中期优化**:
   - 使用更快的哈希算法（如 xxHash）
   - 实现批量操作优化

3. **长期优化**:
   - 使用持久化存储（如 Redis）
   - 实现增量快照

### 7.6 测试覆盖

#### 7.6.1 单元测试

- 创建告警验证（10个测试用例）
- 确认告警验证（3个测试用例）
- 清除告警验证（2个测试用例）
- 查询告警验证（3个测试用例）
- 输出验证（2个测试用例）
- 快照创建（1个测试用例）
- 辅助函数验证（12个测试用例）

#### 7.6.2 集成测试

- 创建告警集成测试（3个测试用例）
- 确认告警集成测试（3个测试用例）
- 清除告警集成测试（2个测试用例）
- 历史告警查询集成测试（2个测试用例）
- 状态验证测试（2个测试用例）

#### 7.6.3 测试覆盖率

- 整体覆盖率: 47.3%
- 新增代码覆盖率: >90%

### 7.7 最佳实践

#### 7.7.1 验证规则设计

1. **输入验证**: 验证数据格式、范围、必填字段
2. **业务规则验证**: 验证业务逻辑、状态转换
3. **约束检查**: 检查业务约束条件
4. **输出验证**: 验证输出数据的正确性

#### 7.7.2 错误处理

```go
// 提供详细的错误信息
type ValidationError struct {
    Field   string
    Message string
    Value   interface{}
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s - %s (value: %v)", e.Field, e.Message, e.Value)
}
```

#### 7.7.3 监控集成

```go
// 记录验证指标
metrics.Counter("validation_total").Inc()
metrics.Counter("validation_success").Inc()
metrics.Histogram("validation_duration").Record(duration)
```

---

## 8. 附录

### 8.1 常用命令速查表

#### 8.1.1 Docker Compose

```bash
docker-compose ps                    # 查看服务状态
docker-compose logs -f api-server    # 查看日志
docker-compose restart api-server    # 重启服务
docker-compose exec api-server sh    # 进入容器
docker stats                         # 资源使用
```

#### 8.1.2 Kubernetes

```bash
kubectl get pods -n nem-system       # 查看Pod
kubectl logs -f <pod> -n nem-system  # 查看日志
kubectl exec -it <pod> -n nem-system -- sh  # 进入容器
kubectl describe pod <pod> -n nem-system    # Pod详情
kubectl get svc -n nem-system        # 查看服务
kubectl get ingress -n nem-system    # 查看Ingress
```

#### 8.1.3 Helm

```bash
helm install nem-system ./helm -n nem-system    # 安装
helm upgrade nem-system ./helm -n nem-system    # 升级
helm rollback nem-system 1 -n nem-system        # 回滚
helm uninstall nem-system -n nem-system         # 卸载
```

#### 8.1.4 数据库

```bash
psql -U postgres -d nem_system       # 连接数据库
pg_dump -U postgres nem_system > backup.sql  # 备份
pg_restore -U postgres -d nem_system backup.dump  # 恢复
```

#### 8.1.5 Redis

```bash
redis-cli                            # 连接Redis
redis-cli INFO                       # 查看信息
redis-cli MONITOR                    # 监控命令
```

### 8.2 故障排查清单

**服务故障**
- [ ] 检查Pod状态
- [ ] 查看Pod事件
- [ ] 查看容器日志
- [ ] 检查资源使用
- [ ] 检查配置文件
- [ ] 检查依赖服务

**数据库故障**
- [ ] 检查数据库进程
- [ ] 测试数据库连接
- [ ] 查看连接数
- [ ] 查看慢查询
- [ ] 检查锁等待
- [ ] 查看磁盘空间

**网络故障**
- [ ] 检查Service配置
- [ ] 检查Endpoints
- [ ] 测试DNS解析
- [ ] 检查网络策略
- [ ] 检查防火墙规则
- [ ] 查看Ingress配置

**性能问题**
- [ ] 查看CPU使用率
- [ ] 查看内存使用
- [ ] 查看磁盘I/O
- [ ] 查看网络流量
- [ ] 分析慢请求
- [ ] 检查数据库查询

### 8.3 运维检查清单

**每日检查**
- [ ] 服务运行状态
- [ ] 错误日志检查
- [ ] 磁盘空间检查
- [ ] 备份任务执行情况

**每周检查**
- [ ] 系统资源使用趋势
- [ ] 安全审计日志
- [ ] 性能指标分析
- [ ] 告警规则审查

**每月检查**
- [ ] 容量规划评估
- [ ] 安全漏洞扫描
- [ ] 备份恢复演练
- [ ] 文档更新

### 8.4 应急联系人

| 角色 | 姓名 | 电话 | 邮箱 | 响应时间 |
|------|------|------|------|----------|
| 运维负责人 | - | - | ops@example.com | 5分钟 |
| 开发负责人 | - | - | dev@example.com | 10分钟 |
| DBA | - | - | dba@example.com | 10分钟 |
| 安全负责人 | - | - | security@example.com | 15分钟 |
| 架构师 | - | - | arch@example.com | 30分钟 |

### 8.5 变更记录

| 版本 | 日期 | 变更内容 | 变更人 |
|------|------|----------|--------|
| v1.0.0 | 2026-04-07 | 初始版本，整合部署、配置、监控、故障排查、备份恢复、安全加固和Harness验证框架 | 运维团队 |

---

## 文档维护

本文档由运维团队维护，如有问题或建议，请联系：

- 邮箱：ops@example.com
- 文档仓库：https://github.com/new-energy-monitoring/docs

**注意**: 本文档包含敏感信息，请妥善保管，不要泄露给未授权人员。
