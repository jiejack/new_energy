# 可观测性配置使用说明

## 概述

本配置为新能源监控系统提供完整的可观测性解决方案，包括：
- **Prometheus**: 指标采集和存储
- **Grafana**: 可视化监控面板
- **Jaeger**: 分布式追踪
- **Alertmanager**: 告警管理
- **Loki**: 日志聚合

## 目录结构

```
deploy/
├── prometheus/
│   ├── prometheus.yml          # Prometheus 主配置
│   └── rules/
│       └── alert_rules.yml     # 告警规则
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/
│   │   │   └── datasources.yml # 数据源配置
│   │   └── dashboards/
│   │       └── dashboards.yml  # Dashboard 自动加载配置
│   └── dashboards/
│       ├── api-service-dashboard.json      # API 服务监控面板
│       ├── postgresql-dashboard.json       # PostgreSQL 数据库监控面板
│       └── go-runtime-dashboard.json       # Go 应用运行时监控面板
├── jaeger/
│   ├── jaeger-config.yml       # Jaeger 配置
│   └── README.md               # Jaeger 集成说明
├── alertmanager/
│   └── alertmanager.yml        # 告警管理器配置
├── promtail/
│   └── config.yml              # 日志采集配置
└── docker-compose.observability.yml  # Docker Compose 编排文件
```

## 快速开始

### 1. 启动可观测性组件

```bash
cd deploy
docker-compose -f docker-compose.observability.yml up -d
```

### 2. 访问服务

| 服务 | 地址 | 默认账号 |
|------|------|----------|
| Prometheus | http://localhost:9090 | - |
| Grafana | http://localhost:3000 | admin / admin123 |
| Jaeger UI | http://localhost:16686 | - |
| Alertmanager | http://localhost:9093 | - |

### 3. 查看监控面板

1. 打开 Grafana: http://localhost:3000
2. 使用默认账号登录
3. 导航到 Dashboards 页面
4. 选择预置的 Dashboard:
   - API 服务监控
   - PostgreSQL 数据库监控
   - Go 应用运行时监控

## Prometheus 配置

### 采集目标

Prometheus 配置了以下采集目标：

| Job 名称 | 目标 | 描述 |
|---------|------|------|
| prometheus | localhost:9090 | Prometheus 自身监控 |
| api-server | api-server:8080 | API 服务 |
| collector | collector:8081 | 数据采集服务 |
| compute-service | compute-service:8082 | 计算服务 |
| alarm-service | alarm-service:8083 | 告警服务 |
| ai-service | ai-service:8084 | AI 服务 |
| scheduler | scheduler:8085 | 调度服务 |
| postgresql | postgres-exporter:9187 | PostgreSQL 数据库 |
| redis | redis-exporter:9121 | Redis 缓存 |
| kafka | kafka-exporter:9308 | Kafka 消息队列 |
| node-exporter | node-exporter:9100 | 系统指标 |
| cadvisor | cadvisor:8080 | 容器指标 |

### 告警规则

告警规则文件位于 `deploy/prometheus/rules/alert_rules.yml`，包含：

- API 服务告警（高延迟、高错误率、服务宕机）
- 数据采集服务告警（错误率、队列积压）
- 计算服务告警（高延迟、服务宕机）
- 告警服务告警（通知失败）
- AI 服务告警（高延迟）
- 数据库告警（连接数、复制延迟）
- Redis 告警（内存使用、连接数）
- Kafka 告警（消费者延迟）
- 系统资源告警（CPU、内存、磁盘）

## Grafana 配置

### 数据源

Grafana 预配置了以下数据源：

1. **Prometheus**: 指标数据源
2. **Jaeger**: 分布式追踪数据源
3. **PostgreSQL**: 业务数据查询
4. **Loki**: 日志数据源

### Dashboard 说明

#### API 服务监控面板
- API 响应时间 (P95)
- 错误率
- 请求速率趋势
- 响应时间分布
- HTTP 状态码分布

#### PostgreSQL 数据库监控面板
- 活跃连接数
- 连接使用率
- 事务速率
- 缓存命中率
- 数据库连接数趋势
- 事务速率趋势
- 数据库 I/O
- 数据库大小

#### Go 应用运行时监控面板
- Goroutines 数量
- 堆内存使用
- GC 停顿时间
- GC 频率
- 内存分配速率
- GC 停顿时间分布

## Jaeger 分布式追踪

### 配置说明

Jaeger 使用 Elasticsearch 作为存储后端，配置文件位于 `deploy/jaeger/jaeger-config.yml`。

### Go 应用集成

详细集成说明请参考 `deploy/jaeger/README.md`，包括：

1. 安装依赖
2. 初始化 Tracer
3. HTTP 中间件
4. gRPC 拦截器
5. 环境变量配置

### 采样策略建议

| 环境 | 采样率 |
|------|--------|
| 开发环境 | 100% |
| 测试环境 | 50% |
| 生产环境 | 10% |

## Alertmanager 配置

### 告警路由

Alertmanager 配置了以下告警路由：

1. **严重告警**: 立即通知，10秒等待，1小时重复
2. **API 服务告警**: 发送给 API 团队
3. **数据库告警**: 发送给数据库团队
4. **系统告警**: 发送给运维团队

### 通知渠道

配置文件中预留了以下通知渠道：

- 邮件通知 (SMTP)
- Slack 通知（可选）
- Webhook 通知（可选）

### 配置邮件通知

编辑 `deploy/alertmanager/alertmanager.yml`，修改以下配置：

```yaml
global:
  smtp_smarthost: 'smtp.example.com:587'
  smtp_from: 'alertmanager@example.com'
  smtp_auth_username: 'alertmanager@example.com'
  smtp_auth_password: 'your-password'
```

## 应用集成

### 1. 添加指标端点

在应用中添加 `/metrics` 端点：

```go
import (
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // 创建 HTTP 路由
    mux := http.NewServeMux()
    
    // 添加指标端点
    mux.Handle("/metrics", promhttp.Handler())
    
    // 启动服务
    http.ListenAndServe(":8080", mux)
}
```

### 2. 使用自定义指标

```go
import "new-energy-monitoring/pkg/monitoring/metrics"

// 记录 API 请求
metrics.APIRequestsTotal.WithLabelValues("api-server", "GET", "/api/v1/stations", "200").Inc()

// 记录数据库查询
metrics.DBConnectionsOpen.Set(10)

// 记录缓存命中
metrics.CacheOperationsTotal.WithLabelValues("get", "hit").Inc()
```

### 3. 添加追踪

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
)

func ProcessData(ctx context.Context) {
    tracer := otel.Tracer("api-server")
    ctx, span := tracer.Start(ctx, "ProcessData")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("user.id", "12345"),
        attribute.Int("data.size", 1024),
    )
    
    // 处理数据...
}
```

## 运维操作

### 查看服务状态

```bash
# 查看所有服务状态
docker-compose -f docker-compose.observability.yml ps

# 查看服务日志
docker-compose -f docker-compose.observability.yml logs -f prometheus
docker-compose -f docker-compose.observability.yml logs -f grafana
docker-compose -f docker-compose.observability.yml logs -f jaeger
```

### 重启服务

```bash
# 重启所有服务
docker-compose -f docker-compose.observability.yml restart

# 重启单个服务
docker-compose -f docker-compose.observability.yml restart prometheus
```

### 停止服务

```bash
# 停止所有服务
docker-compose -f docker-compose.observability.yml down

# 停止并删除数据卷
docker-compose -f docker-compose.observability.yml down -v
```

### 更新配置

```bash
# 修改配置后重新加载 Prometheus
curl -X POST http://localhost:9090/-/reload

# 重启 Grafana 以加载新 Dashboard
docker-compose -f docker-compose.observability.yml restart grafana
```

### 备份数据

```bash
# 备份 Prometheus 数据
docker cp prometheus:/prometheus ./backup/prometheus

# 备份 Grafana 数据
docker cp grafana:/var/lib/grafana ./backup/grafana
```

## 性能优化

### Prometheus 优化

1. **存储优化**
   - 调整 `storage.tsdb.retention.time` 控制数据保留时间
   - 调整 `storage.tsdb.retention.size` 控制存储大小

2. **采集优化**
   - 根据服务重要性调整 `scrape_interval`
   - 使用 `scrape_timeout` 避免超时

### Grafana 优化

1. **Dashboard 优化**
   - 减少面板数量
   - 使用变量减少查询次数
   - 合理设置刷新间隔

2. **查询优化**
   - 使用 `rate()` 而非 `increase()`
   - 避免使用 `avg_over_time()` 等高开销函数

### Jaeger 优化

1. **采样优化**
   - 生产环境使用 10% 采样率
   - 关键路径使用 100% 采样

2. **存储优化**
   - Elasticsearch 定期清理旧数据
   - 使用索引生命周期管理

## 故障排查

### Prometheus 无法采集指标

1. 检查目标服务是否运行
2. 检查 `/metrics` 端点是否可访问
3. 查看 Prometheus 日志

### Grafana 无法显示数据

1. 检查数据源配置
2. 检查 Prometheus 是否有数据
3. 检查 Dashboard 查询语句

### Jaeger 无法接收追踪数据

1. 检查 Jaeger 服务状态
2. 检查应用是否正确配置 Jaeger Agent 地址
3. 查看 Jaeger 日志

### 告警未发送

1. 检查 Alertmanager 配置
2. 检查邮件/Slack 配置
3. 查看 Alertmanager 日志

## 监控最佳实践

1. **指标命名规范**
   - 使用标准前缀：`http_`, `db_`, `cache_`
   - 包含单位：`_seconds`, `_bytes`, `_total`

2. **告警阈值设置**
   - 根据历史数据设置合理阈值
   - 避免告警疲劳
   - 分级告警（warning, critical）

3. **Dashboard 设计**
   - 分层次展示（概览 -> 详情）
   - 使用变量提高复用性
   - 添加文档说明

4. **追踪使用**
   - 关键路径必须追踪
   - 合理设置采样率
   - 避免追踪敏感信息

## 相关文档

- [Prometheus 官方文档](https://prometheus.io/docs/)
- [Grafana 官方文档](https://grafana.com/docs/)
- [Jaeger 官方文档](https://www.jaegertracing.io/docs/)
- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/instrumentation/go/)
