# 运维指南

## 文档信息

| 项目 | 内容 |
|------|------|
| 项目名称 | 新能源在线监控系统 |
| 文档版本 | v1.0.0 |
| 编写日期 | 2024-03-01 |
| 文档状态 | 正式发布 |

---

## 1. 日常运维操作

### 1.1 服务管理

#### Docker环境

```bash
# 查看所有服务状态
docker-compose ps

# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 重启单个服务
docker-compose restart api-server

# 重启所有服务
docker-compose restart

# 查看服务日志
docker-compose logs -f api-server

# 查看最近100行日志
docker-compose logs --tail=100 api-server

# 进入容器
docker-compose exec api-server sh

# 查看容器资源使用
docker stats
```

#### Kubernetes环境

```bash
# 查看所有Pod状态
kubectl get pods -n nem-system

# 查看Pod详细信息
kubectl describe pod <pod-name> -n nem-system

# 查看服务状态
kubectl get svc -n nem-system

# 查看部署状态
kubectl get deployments -n nem-system

# 重启服务
kubectl rollout restart deployment/api-server -n nem-system

# 扩缩容
kubectl scale deployment api-server --replicas=5 -n nem-system

# 查看日志
kubectl logs -f <pod-name> -n nem-system

# 查看多容器Pod的日志
kubectl logs -f <pod-name> -c <container-name> -n nem-system

# 进入容器
kubectl exec -it <pod-name> -n nem-system -- sh

# 查看资源使用
kubectl top pods -n nem-system
kubectl top nodes
```

### 1.2 配置管理

#### 更新配置文件

**Docker环境**

```bash
# 1. 修改配置文件
vim configs/config.yaml

# 2. 重启服务使配置生效
docker-compose restart api-server
```

**Kubernetes环境**

```bash
# 1. 更新ConfigMap
kubectl create configmap nem-config \
  --from-file=config.yaml=configs/config.yaml \
  --dry-run=client -o yaml | kubectl apply -f - -n nem-system

# 2. 重启Pod使配置生效
kubectl rollout restart deployment/api-server -n nem-system

# 3. 查看更新状态
kubectl rollout status deployment/api-server -n nem-system
```

#### 环境变量管理

```bash
# Kubernetes更新Secret
kubectl create secret generic nem-secrets \
  --from-literal=db-password=new-password \
  --dry-run=client -o yaml | kubectl apply -f - -n nem-system

# 重启服务
kubectl rollout restart deployment/api-server -n nem-system
```

### 1.3 日志管理

#### 日志查看

```bash
# Docker环境
docker-compose logs -f --tail=100 api-server

# Kubernetes环境
kubectl logs -f deployment/api-server -n nem-system --tail=100

# 查看前一个容器的日志（容器重启后）
kubectl logs deployment/api-server -n nem-system --previous

# 导出日志到文件
kubectl logs deployment/api-server -n nem-system > api-server.log
```

#### 日志级别动态调整

```bash
# 通过API调整日志级别
curl -X PUT http://localhost:8080/api/v1/admin/log/level \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"level":"debug"}'

# 查看当前日志级别
curl http://localhost:8080/api/v1/admin/log/level \
  -H "Authorization: Bearer <token>"
```

#### 日志轮转配置

```yaml
# logging配置
logging:
  level: info
  format: json
  output: file
  file:
    path: /var/log/nem/app.log
    max_size: 100        # 单文件最大100MB
    max_backups: 10      # 保留10个备份
    max_age: 30          # 保留30天
    compress: true       # 压缩旧日志
```

### 1.4 定时任务管理

#### 查看定时任务

```bash
# 查看Kubernetes CronJob
kubectl get cronjobs -n nem-system

# 查看CronJob详情
kubectl describe cronjob <cronjob-name> -n nem-system

# 查看任务执行历史
kubectl get jobs -n nem-system

# 手动触发任务
kubectl create job --from=cronjob/<cronjob-name> manual-run-$(date +%s) -n nem-system
```

#### 定时任务配置

```yaml
# 统计计算任务
apiVersion: batch/v1
kind: CronJob
metadata:
  name: statistics-compute
  namespace: nem-system
spec:
  schedule: "0 * * * *"  # 每小时执行
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: compute
            image: new-energy-monitoring/compute:1.0.0
            command: ["./compute", "-task", "hourly-stats"]
          restartPolicy: OnFailure
```

---

## 2. 监控和告警

### 2.1 Prometheus监控

#### 访问Prometheus

```bash
# 端口转发
kubectl port-forward svc/prometheus 9090:9090 -n monitoring

# 访问UI
# http://localhost:9090
```

#### 常用监控查询

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

#### 服务健康检查

```bash
# API健康检查
curl http://localhost:8080/health

# 就绪检查
curl http://localhost:8080/ready

# 详细健康状态
curl http://localhost:8080/health/detailed
```

### 2.2 Grafana可视化

#### 访问Grafana

```bash
# 端口转发
kubectl port-forward svc/grafana 3000:80 -n monitoring

# 访问UI
# http://localhost:3000
# 默认账号: admin/admin
```

#### 关键Dashboard

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

### 2.3 告警管理

#### 查看告警规则

```bash
# 查看Prometheus告警规则
kubectl get prometheusrules -n nem-system

# 查看告警状态
curl http://prometheus:9090/api/v1/alerts
```

#### 告警规则示例

```yaml
groups:
- name: nem-alerts
  rules:
  # 服务宕机告警
  - alert: ServiceDown
    expr: up == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "服务 {{ $labels.job }} 已下线"
      description: "实例 {{ $labels.instance }} 已下线超过1分钟"

  # 高CPU告警
  - alert: HighCPU
    expr: rate(process_cpu_seconds_total[5m]) > 0.8
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "CPU使用率过高"
      description: "服务 {{ $labels.job }} CPU使用率 {{ $value }}"

  # 内存告警
  - alert: HighMemory
    expr: process_resident_memory_bytes / 1024 / 1024 / 1024 > 1.5
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "内存使用过高"
      description: "服务 {{ $labels.job }} 内存使用 {{ $value }}GB"

  # 磁盘空间告警
  - alert: DiskSpaceLow
    expr: (node_filesystem_avail_bytes / node_filesystem_size_bytes) < 0.1
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "磁盘空间不足"
      description: "磁盘 {{ $labels.mountpoint }} 剩余空间不足10%"

  # API错误率告警
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.01
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "API错误率过高"
      description: "HTTP 5xx错误率 {{ $value }}"

  # 数据库连接数告警
  - alert: HighDBConnections
    expr: pg_stat_activity_count / pg_settings_max_connections > 0.8
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "数据库连接数过高"
      description: "连接数使用率 {{ $value }}"
```

#### 告警通知配置

```yaml
# Alertmanager配置
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
  slack_configs:
  - api_url: 'https://hooks.slack.com/services/xxx'
    channel: '#alerts'
```

### 2.4 链路追踪

#### Jaeger访问

```bash
# 端口转发
kubectl port-forward svc/jaeger 16686:16686 -n monitoring

# 访问UI
# http://localhost:16686
```

#### 追踪配置

```yaml
tracing:
  enabled: true
  endpoint: jaeger-collector:4317
  sampler_ratio: 0.1  # 采样率10%
  service_name: nem-api-server
```

---

## 3. 备份与恢复

### 3.1 数据备份策略

| 数据类型 | 备份方式 | 频率 | 保留周期 | 存储位置 |
|----------|----------|------|----------|----------|
| PostgreSQL | 全量备份 | 每日 | 30天 | /backup/postgres |
| PostgreSQL | WAL归档 | 实时 | 7天 | /backup/wal |
| Redis | RDB快照 | 每小时 | 7天 | /backup/redis |
| Doris | 快照备份 | 每日 | 30天 | /backup/doris |
| 配置文件 | Git版本 | 实时 | 永久 | Git仓库 |

### 3.2 PostgreSQL备份

#### 手动备份

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

#### 自动备份（Kubernetes）

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

#### 数据恢复

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

### 3.3 Redis备份

#### 手动备份

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

#### 数据恢复

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

### 3.4 配置文件备份

```bash
#!/bin/bash
# backup_configs.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/configs"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份配置文件
tar czf $BACKUP_DIR/configs_${DATE}.tar.gz \
  configs/ \
  deployments/kubernetes/helm/values*.yaml \
  k8s/*.yaml

# 删除30天前的备份
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete

echo "Config backup completed: configs_${DATE}.tar.gz"
```

---

## 4. 扩容与缩容

### 4.1 水平扩容

#### Kubernetes手动扩容

```bash
# 扩容API Server
kubectl scale deployment api-server --replicas=5 -n nem-system

# 扩容Collector
kubectl scale deployment collector --replicas=10 -n nem-system

# 验证扩容状态
kubectl get pods -n nem-system -l app=api-server
```

#### 自动扩容（HPA）

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-server-hpa
  namespace: nem-system
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-server
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

```bash
# 创建HPA
kubectl apply -f hpa.yaml -n nem-system

# 查看HPA状态
kubectl get hpa -n nem-system

# 查看HPA详情
kubectl describe hpa api-server-hpa -n nem-system
```

### 4.2 垂直扩容

#### 调整资源限制

```bash
# 编辑Deployment
kubectl edit deployment api-server -n nem-system

# 修改resources部分
resources:
  limits:
    cpu: 4000m
    memory: 8Gi
  requests:
    cpu: 2000m
    memory: 4Gi
```

#### 使用VPA（Vertical Pod Autoscaler）

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: api-server-vpa
  namespace: nem-system
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-server
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: api-server
      minAllowed:
        cpu: 500m
        memory: 1Gi
      maxAllowed:
        cpu: 8000m
        memory: 16Gi
```

### 4.3 数据库扩容

#### PostgreSQL扩容

```bash
# 增加存储
kubectl patch pvc postgres-data -n nem-system \
  -p '{"spec":{"resources":{"requests":{"storage":"200Gi"}}}}'

# 增加连接数（需要重启）
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -c "ALTER SYSTEM SET max_connections = 300;"
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -c "SELECT pg_reload_conf();"
```

#### Redis扩容

```bash
# 增加内存限制
kubectl edit deployment redis -n nem-system

# 修改maxmemory配置
# --maxmemory 8gb
```

### 4.4 缩容操作

```bash
# 手动缩容
kubectl scale deployment api-server --replicas=2 -n nem-system

# 设置HPA最小副本数
kubectl patch hpa api-server-hpa -n nem-system \
  -p '{"spec":{"minReplicas":2}}'
```

---

## 5. 日志管理

### 5.1 日志收集架构

```
应用Pod -> Fluentd/Fluent Bit -> Elasticsearch -> Kibana
                                    ↓
                                日志归档存储
```

### 5.2 Fluentd配置

```yaml
# fluentd-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
  namespace: logging
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/containers/*.log
      pos_file /var/log/fluentd-containers.log.pos
      tag kubernetes.*
      format json
      time_key time
      time_format %Y-%m-%dT%H:%M:%S.%NZ
    </source>

    <filter kubernetes.**>
      @type kubernetes_metadata
    </filter>

    <match kubernetes.var.log.containers.**>
      @type elasticsearch
      host elasticsearch
      port 9200
      logstash_format true
      logstash_prefix nem-logs
      <buffer>
        @type file
        path /var/log/fluentd-buffer
        flush_interval 5s
      </buffer>
    </match>
```

### 5.3 日志查询

#### Kibana查询

```
# 查询API Server错误日志
kubernetes.pod_name: "api-server*" AND level: "error"

# 查询特定时间范围的日志
@timestamp: ["2024-03-01T00:00:00" TO "2024-03-01T23:59:59"]

# 查询包含特定关键字的日志
message: "database connection failed"
```

#### kubectl日志查询

```bash
# 查询特定时间段的日志
kubectl logs deployment/api-server -n nem-system \
  --since=1h

# 查询特定标签的Pod日志
kubectl logs -l app=api-server -n nem-system

# 实时查看日志
kubectl logs -f deployment/api-server -n nem-system

# 查看所有容器的日志
kubectl logs deployment/api-server -n nem-system --all-containers
```

### 5.4 日志归档

```bash
#!/bin/bash
# archive_logs.sh

DATE=$(date +%Y%m%d)
ARCHIVE_DIR="/archive/logs"

# 创建归档目录
mkdir -p $ARCHIVE_DIR

# 归档7天前的日志
find /var/log/nem -name "*.log" -mtime +7 -exec gzip {} \;

# 移动到归档目录
find /var/log/nem -name "*.gz" -exec mv {} $ARCHIVE_DIR/ \;

# 删除30天前的归档
find $ARCHIVE_DIR -name "*.gz" -mtime +30 -delete

echo "Log archive completed for $DATE"
```

---

## 6. 安全运维

### 6.1 访问控制

#### RBAC配置

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

### 6.2 网络策略

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

### 6.3 密钥轮换

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

### 6.4 安全审计

```bash
# 查看审计日志
kubectl logs -n kube-system kube-apiserver-master | grep audit

# 查看用户操作
kubectl get events -n nem-system --sort-by='.lastTimestamp'

# 检查Pod安全策略
kubectl get psp

# 检查NetworkPolicy
kubectl get networkpolicy -n nem-system
```

---

## 7. 性能优化

### 7.1 数据库优化

```sql
-- 查看慢查询
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- 创建索引
CREATE INDEX idx_device_station_id ON devices(station_id);
CREATE INDEX idx_alarm_created_at ON alarms(created_at);

-- 分析查询计划
EXPLAIN ANALYZE SELECT * FROM devices WHERE station_id = 'station-001';

-- 更新统计信息
ANALYZE devices;

-- 清理空间
VACUUM FULL devices;
```

### 7.2 Redis优化

```bash
# 查看Redis信息
redis-cli INFO

# 查看内存使用
redis-cli INFO memory

# 查看慢查询
redis-cli SLOWLOG GET 10

# 清理过期键
redis-cli DEBUG SLEEP 0.1

# 监控命令
redis-cli MONITOR
```

### 7.3 应用优化

```bash
# 查看应用性能指标
curl http://localhost:9090/metrics | grep http_request_duration

# 生成CPU profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof

# 分析CPU profile
go tool pprof cpu.prof

# 生成内存profile
curl http://localhost:8080/debug/pprof/heap > heap.prof

# 分析内存profile
go tool pprof heap.prof
```

---

## 8. 运维工具脚本

### 8.1 健康检查脚本

```bash
#!/bin/bash
# health_check.sh

echo "=== 系统健康检查 ==="
echo ""

# 检查API服务
echo "1. API服务检查"
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "   ✓ API服务正常"
else
    echo "   ✗ API服务异常"
fi

# 检查数据库
echo "2. 数据库检查"
if docker-compose exec -T postgres pg_isready > /dev/null 2>&1; then
    echo "   ✓ 数据库连接正常"
else
    echo "   ✗ 数据库连接异常"
fi

# 检查Redis
echo "3. Redis检查"
if docker-compose exec -T redis redis-cli ping | grep -q PONG; then
    echo "   ✓ Redis连接正常"
else
    echo "   ✗ Redis连接异常"
fi

# 检查Kafka
echo "4. Kafka检查"
if docker-compose exec -T kafka kafka-broker-api-versions.sh -bootstrap-server localhost:9092 > /dev/null 2>&1; then
    echo "   ✓ Kafka连接正常"
else
    echo "   ✗ Kafka连接异常"
fi

# 检查磁盘空间
echo "5. 磁盘空间检查"
DISK_USAGE=$(df -h / | tail -1 | awk '{print $5}' | sed 's/%//')
if [ $DISK_USAGE -lt 80 ]; then
    echo "   ✓ 磁盘空间充足 (使用率: ${DISK_USAGE}%)"
else
    echo "   ✗ 磁盘空间不足 (使用率: ${DISK_USAGE}%)"
fi

# 检查内存
echo "6. 内存检查"
MEM_USAGE=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100}')
if [ $(echo "$MEM_USAGE < 80" | bc) -eq 1 ]; then
    echo "   ✓ 内存充足 (使用率: ${MEM_USAGE}%)"
else
    echo "   ✗ 内存不足 (使用率: ${MEM_USAGE}%)"
fi

echo ""
echo "健康检查完成"
```

### 8.2 服务重启脚本

```bash
#!/bin/bash
# restart_services.sh

SERVICE=$1

if [ -z "$SERVICE" ]; then
    echo "用法: $0 <service-name>"
    echo "可用服务: api-server, collector, alarm, compute, ai-service, scheduler"
    exit 1
fi

echo "重启服务: $SERVICE"

# 健康检查
health_check() {
    for i in {1..30}; do
        if curl -f http://localhost:8080/health > /dev/null 2>&1; then
            return 0
        fi
        sleep 1
    done
    return 1
}

# 重启服务
docker-compose restart $SERVICE

# 等待服务就绪
echo "等待服务启动..."
sleep 5

# 健康检查
if health_check; then
    echo "✓ 服务重启成功"
else
    echo "✗ 服务重启失败"
    exit 1
fi
```

### 8.3 日志清理脚本

```bash
#!/bin/bash
# clean_logs.sh

LOG_DIR="/var/log/nem"
MAX_DAYS=7

echo "清理日志文件..."

# 删除旧日志
find $LOG_DIR -name "*.log" -mtime +$MAX_DAYS -delete
find $LOG_DIR -name "*.gz" -mtime +$MAX_DAYS -delete

# 清理Docker日志
docker-compose logs --no-color > /dev/null 2>&1

# 清理Kubernetes日志（如果使用k8s）
if command -v kubectl &> /dev/null; then
    kubectl logs --previous --all-containers=true -n nem-system > /dev/null 2>&1
fi

echo "日志清理完成"
```

---

## 9. 应急响应

### 9.1 服务故障响应流程

```
1. 发现故障
   ↓
2. 确认影响范围
   ↓
3. 初步诊断
   ↓
4. 尝试快速恢复
   ↓
5. 问题定位
   ↓
6. 制定解决方案
   ↓
7. 实施修复
   ↓
8. 验证恢复
   ↓
9. 事后总结
```

### 9.2 应急联系人

| 角色 | 姓名 | 电话 | 邮箱 |
|------|------|------|------|
| 运维负责人 | - | - | ops@example.com |
| 开发负责人 | - | - | dev@example.com |
| DBA | - | - | dba@example.com |
| 安全负责人 | - | - | security@example.com |

### 9.3 应急操作手册

#### 服务宕机

```bash
# 1. 查看Pod状态
kubectl get pods -n nem-system

# 2. 查看Pod事件
kubectl describe pod <pod-name> -n nem-system

# 3. 查看日志
kubectl logs <pod-name> -n nem-system --previous

# 4. 重启服务
kubectl rollout restart deployment/<deployment-name> -n nem-system

# 5. 回滚到上一版本
kubectl rollout undo deployment/<deployment-name> -n nem-system
```

#### 数据库故障

```bash
# 1. 检查数据库状态
kubectl exec -it postgres-pod -n nem-system -- pg_isready

# 2. 查看数据库日志
kubectl logs postgres-pod -n nem-system

# 3. 检查连接数
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

# 4. 终止空闲连接
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'idle';"
```

---

## 10. 附录

### 10.1 常用命令速查表

```bash
# Docker Compose
docker-compose ps                    # 查看服务状态
docker-compose logs -f api-server    # 查看日志
docker-compose restart api-server    # 重启服务
docker-compose exec api-server sh    # 进入容器

# Kubernetes
kubectl get pods -n nem-system       # 查看Pod
kubectl logs -f <pod> -n nem-system  # 查看日志
kubectl exec -it <pod> -n nem-system -- sh  # 进入容器
kubectl describe pod <pod> -n nem-system    # Pod详情

# 数据库
psql -U postgres -d nem_system       # 连接数据库
pg_dump -U postgres nem_system > backup.sql  # 备份

# Redis
redis-cli                            # 连接Redis
redis-cli INFO                       # 查看信息
redis-cli MONITOR                    # 监控命令
```

### 10.2 运维检查清单

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

### 10.3 变更记录

| 版本 | 日期 | 变更内容 | 变更人 |
|------|------|----------|--------|
| v1.0.0 | 2024-03-01 | 初始版本 | 运维团队 |
