# 故障排查指南

## 文档信息

| 项目 | 内容 |
|------|------|
| 项目名称 | 新能源在线监控系统 |
| 文档版本 | v1.0.0 |
| 编写日期 | 2024-03-01 |
| 文档状态 | 正式发布 |

---

## 1. 常见问题分类

### 1.1 问题分类

| 分类 | 描述 | 优先级 |
|------|------|--------|
| 服务故障 | 服务无法启动、崩溃、无响应 | P0 |
| 性能问题 | 响应慢、资源占用高 | P1 |
| 数据问题 | 数据丢失、数据不一致 | P1 |
| 连接问题 | 数据库、Redis、Kafka连接失败 | P1 |
| 配置问题 | 配置错误、配置不生效 | P2 |
| 权限问题 | 访问被拒绝、认证失败 | P2 |

### 1.2 排查流程

```
发现问题
    ↓
确认影响范围
    ↓
收集信息（日志、指标）
    ↓
定位问题原因
    ↓
制定解决方案
    ↓
实施修复
    ↓
验证恢复
    ↓
记录总结
```

---

## 2. 服务启动故障

### 2.1 服务无法启动

#### 症状
- 容器启动后立即退出
- 服务状态为CrashLoopBackOff
- 日志显示启动错误

#### 排查步骤

**步骤1：查看Pod状态**

```bash
# Kubernetes环境
kubectl get pods -n nem-system

# 查看Pod详情
kubectl describe pod <pod-name> -n nem-system

# 查看Pod事件
kubectl get events -n nem-system --field-selector involvedObject.name=<pod-name>
```

**步骤2：查看日志**

```bash
# 查看当前日志
kubectl logs <pod-name> -n nem-system

# 查看上一个容器的日志
kubectl logs <pod-name> -n nem-system --previous

# 查看所有容器的日志
kubectl logs <pod-name> -n nem-system --all-containers
```

**步骤3：检查配置**

```bash
# 查看ConfigMap
kubectl get configmap -n nem-system
kubectl describe configmap <configmap-name> -n nem-system

# 查看Secret
kubectl get secret -n nem-system
kubectl describe secret <secret-name> -n nem-system
```

**步骤4：检查资源限制**

```bash
# 查看资源使用
kubectl top pods -n nem-system
kubectl describe resourcequota -n nem-system
```

#### 常见原因及解决方案

| 原因 | 解决方案 |
|------|----------|
| 配置文件错误 | 检查config.yaml语法，验证必需配置项 |
| 环境变量缺失 | 检查Secret和环境变量配置 |
| 依赖服务未就绪 | 确保数据库、Redis等服务已启动 |
| 资源不足 | 增加资源限制或释放资源 |
| 镜像拉取失败 | 检查镜像是否存在、网络连接、镜像仓库认证 |
| 端口冲突 | 检查端口是否被占用 |

### 2.2 服务启动超时

#### 症状
- 服务启动时间过长
- 健康检查失败
- 就绪检查超时

#### 排查步骤

```bash
# 查看健康检查配置
kubectl describe deployment <deployment-name> -n nem-system | grep -A 10 "Liveness"

# 查看启动日志
kubectl logs -f <pod-name> -n nem-system

# 检查依赖服务连接
kubectl exec -it <pod-name> -n nem-system -- sh
# 在容器内测试连接
nc -zv postgres 5432
nc -zv redis 6379
```

#### 解决方案

```yaml
# 调整健康检查配置
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 60  # 增加初始延迟
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

---

## 3. 数据库故障

### 3.1 数据库连接失败

#### 症状
- 应用日志显示数据库连接错误
- API返回500错误
- 数据库相关操作失败

#### 排查步骤

**步骤1：检查数据库服务状态**

```bash
# PostgreSQL
docker-compose exec postgres pg_isready -U postgres

# Kubernetes环境
kubectl exec -it postgres-pod -n nem-system -- pg_isready -U postgres

# 检查数据库进程
docker-compose exec postgres ps aux | grep postgres
```

**步骤2：测试数据库连接**

```bash
# 手动连接测试
docker-compose exec postgres psql -U postgres -d nem_system

# Kubernetes环境
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -d nem_system

# 测试网络连接
telnet postgres 5432
nc -zv postgres 5432
```

**步骤3：检查连接配置**

```bash
# 查看环境变量
kubectl exec -it <pod-name> -n nem-system -- env | grep DB

# 查看Secret
kubectl get secret nem-secrets -n nem-system -o yaml
```

**步骤4：检查连接数**

```bash
# 查看当前连接数
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

# 查看最大连接数
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -c "SHOW max_connections;"

# 查看连接详情
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -c "SELECT pid, usename, application_name, client_addr, state, query FROM pg_stat_activity;"
```

#### 常见错误及解决方案

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

# 长期方案：优化连接池配置
# config.yaml
database:
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 3600
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

**错误3：数据库不存在**

```
FATAL: database "nem_system" does not exist
```

解决方案：

```bash
# 创建数据库
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -c "CREATE DATABASE nem_system;"

# 执行迁移脚本
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -d nem_system -f /docker-entrypoint-initdb.d/001_init_schema.sql
```

### 3.2 慢查询问题

#### 症状
- API响应缓慢
- 数据库CPU使用率高
- 查询超时

#### 排查步骤

**步骤1：开启慢查询日志**

```sql
-- 开启慢查询日志（超过1秒）
ALTER SYSTEM SET log_min_duration_statement = 1000;
SELECT pg_reload_conf();

-- 查看配置
SHOW log_min_duration_statement;
```

**步骤2：查看慢查询**

```sql
-- 安装pg_stat_statements扩展
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

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

-- 查看最耗时的SQL
SELECT
  query,
  total_time,
  calls,
  rows
FROM pg_stat_statements
ORDER BY total_time DESC
LIMIT 10;
```

**步骤3：分析查询计划**

```sql
-- 使用EXPLAIN分析
EXPLAIN ANALYZE SELECT * FROM devices WHERE station_id = 'station-001';

-- 查看详细执行计划
EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM devices WHERE station_id = 'station-001';
```

#### 解决方案

**方案1：创建索引**

```sql
-- 创建索引
CREATE INDEX idx_devices_station_id ON devices(station_id);
CREATE INDEX idx_alarms_created_at ON alarms(created_at);
CREATE INDEX idx_points_device_id ON points(device_id);

-- 创建复合索引
CREATE INDEX idx_alarms_station_time ON alarms(station_id, created_at);

-- 创建部分索引
CREATE INDEX idx_active_devices ON devices(station_id) WHERE status = 'active';
```

**方案2：优化查询**

```sql
-- 避免SELECT *
SELECT id, name, status FROM devices WHERE station_id = 'station-001';

-- 使用LIMIT
SELECT * FROM alarms WHERE station_id = 'station-001' ORDER BY created_at DESC LIMIT 100;

-- 使用JOIN代替子查询
SELECT d.*, s.name as station_name
FROM devices d
JOIN stations s ON d.station_id = s.id
WHERE d.status = 'active';

-- 分页优化
SELECT * FROM devices
WHERE id > 'last_id'
ORDER BY id
LIMIT 100;
```

**方案3：表分区**

```sql
-- 按时间分区
CREATE TABLE alarms (
  id VARCHAR(64) NOT NULL,
  station_id VARCHAR(64) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  -- 其他字段
) PARTITION BY RANGE (created_at);

-- 创建分区
CREATE TABLE alarms_2024_01 PARTITION OF alarms
  FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE alarms_2024_02 PARTITION OF alarms
  FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
```

### 3.3 数据库死锁

#### 症状
- 事务长时间等待
- 应用日志显示死锁错误
- 数据库响应缓慢

#### 排查步骤

```sql
-- 查看锁等待
SELECT
  blocked_locks.pid AS blocked_pid,
  blocked_activity.usename AS blocked_user,
  blocking_locks.pid AS blocking_pid,
  blocking_activity.usename AS blocking_user,
  blocked_activity.query AS blocked_statement,
  blocking_activity.query AS blocking_statement
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
  AND blocking_locks.database = blocked_locks.database
  AND blocking_locks.relation = blocked_locks.relation
  AND blocking_locks.page = blocked_locks.page
  AND blocking_locks.tuple = blocked_locks.tuple
  AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;

-- 查看当前锁
SELECT * FROM pg_locks WHERE NOT granted;
```

#### 解决方案

```sql
-- 终止阻塞进程
SELECT pg_terminate_backend(<blocking_pid>);

-- 设置锁超时
SET lock_timeout = '10s';

-- 优化事务顺序
-- 确保所有事务按相同顺序访问表
```

---

## 4. Redis故障

### 4.1 Redis连接失败

#### 症状
- 应用日志显示Redis连接错误
- 缓存功能失效
- 会话管理失败

#### 排查步骤

```bash
# 检查Redis服务状态
docker-compose exec redis redis-cli ping

# Kubernetes环境
kubectl exec -it redis-pod -n nem-system -- redis-cli ping

# 测试连接
redis-cli -h redis -p 6379 ping

# 查看Redis信息
redis-cli INFO

# 查看连接数
redis-cli INFO clients
```

#### 解决方案

**错误1：连接数超限**

```
ERR max number of clients reached
```

```bash
# 查看最大连接数
redis-cli CONFIG GET maxclients

# 增加最大连接数
redis-cli CONFIG SET maxclients 10000

# 查看当前连接
redis-cli CLIENT LIST
```

**错误2：内存不足**

```
OOM command not allowed when used memory > 'maxmemory'
```

```bash
# 查看内存使用
redis-cli INFO memory

# 查看最大内存配置
redis-cli CONFIG GET maxmemory

# 增加最大内存
redis-cli CONFIG SET maxmemory 4gb

# 设置内存淘汰策略
redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

### 4.2 Redis性能问题

#### 症状
- Redis响应慢
- CPU使用率高
- 内存使用持续增长

#### 排查步骤

```bash
# 查看慢查询
redis-cli SLOWLOG GET 10

# 查看内存使用详情
redis-cli MEMORY STATS

# 查看大键
redis-cli --bigkeys

# 监控实时命令
redis-cli MONITOR

# 查看统计信息
redis-cli INFO stats
```

#### 解决方案

**方案1：优化大键**

```bash
# 查找大键
redis-cli --bigkeys

# 删除大键（使用UNLINK避免阻塞）
redis-cli UNLINK big-key-name

# 拆分大键
# 将大hash拆分为多个小hash
```

**方案2：优化内存使用**

```bash
# 查看内存使用
redis-cli MEMORY USAGE key-name

# 使用更高效的数据结构
# 使用hash代替多个key
# 使用压缩列表

# 开启压缩
redis-cli CONFIG SET hash-max-ziplist-entries 512
redis-cli CONFIG SET hash-max-ziplist-value 64
```

---

## 5. Kafka故障

### 5.1 Kafka连接失败

#### 症状
- 生产者发送消息失败
- 消费者无法消费
- 连接超时

#### 排查步骤

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

#### 解决方案

**错误1：主题不存在**

```bash
# 创建主题
docker-compose exec kafka kafka-topics.sh \
  --create \
  --topic nem.data.collect \
  --partitions 12 \
  --replication-factor 1 \
  --bootstrap-server localhost:9092
```

**错误2：消费者组问题**

```bash
# 重置消费者组偏移量
docker-compose exec kafka kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --group nem-group \
  --reset-offsets \
  --to-latest \
  --topic nem.data.collect \
  --execute

# 删除消费者组
docker-compose exec kafka kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --group nem-group \
  --delete
```

### 5.2 消息积压

#### 症状
- 消费延迟持续增长
- 消费者Lag过大
- 数据处理不及时

#### 排查步骤

```bash
# 查看消费延迟
docker-compose exec kafka kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe \
  --group nem-group

# 查看主题详情
docker-compose exec kafka kafka-topics.sh \
  --describe \
  --topic nem.data.collect \
  --bootstrap-server localhost:9092
```

#### 解决方案

**方案1：增加消费者实例**

```bash
# 扩容消费者
kubectl scale deployment collector --replicas=10 -n nem-system
```

**方案2：增加分区数**

```bash
# 增加分区
docker-compose exec kafka kafka-topics.sh \
  --alter \
  --topic nem.data.collect \
  --partitions 24 \
  --bootstrap-server localhost:9092
```

**方案3：优化消费逻辑**

```go
// 增加批量消费
config := sarama.NewConfig()
config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
config.Consumer.Fetch.Default = 1024 * 1024 // 1MB
config.Consumer.MaxProcessingTime = time.Second * 5
```

---

## 6. API性能问题

### 6.1 API响应慢

#### 症状
- API响应时间超过阈值
- 用户请求超时
- 前端加载缓慢

#### 排查步骤

**步骤1：查看API指标**

```bash
# 查看Prometheus指标
curl http://localhost:9090/metrics | grep http_request_duration

# 查看API响应时间分布
curl http://localhost:9090/api/v1/query \
  --data-urlencode 'query=histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))'
```

**步骤2：分析链路追踪**

```bash
# 访问Jaeger UI
# http://localhost:16686

# 查找慢请求
# 按duration排序，找出最慢的trace
```

**步骤3：查看应用日志**

```bash
# 查看慢请求日志
kubectl logs deployment/api-server -n nem-system | grep "slow request"

# 查看错误日志
kubectl logs deployment/api-server -n nem-system | grep "ERROR"
```

**步骤4：性能分析**

```bash
# 获取CPU profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof

# 分析profile
go tool pprof cpu.prof
(pprof) top10
(pprof) list functionName

# 获取内存profile
curl http://localhost:8080/debug/pprof/heap > heap.prof

# 分析内存
go tool pprof heap.prof
(pprof) top10
```

#### 解决方案

**方案1：添加缓存**

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

**方案2：数据库查询优化**

```sql
-- 添加索引
CREATE INDEX idx_devices_station_status ON devices(station_id, status);

-- 使用覆盖索引
CREATE INDEX idx_devices_covering ON devices(station_id, status) INCLUDE (name, type);

-- 优化JOIN
-- 确保JOIN字段有索引
```

**方案3：并发处理**

```go
// 使用goroutine并发处理
func (s *Service) GetStationData(stationID string) (*StationData, error) {
    var wg sync.WaitGroup
    var devices []*Device
    var alarms []*Alarm
    var stats *Statistics

    wg.Add(3)

    go func() {
        defer wg.Done()
        devices, _ = s.getDevices(stationID)
    }()

    go func() {
        defer wg.Done()
        alarms, _ = s.getAlarms(stationID)
    }()

    go func() {
        defer wg.Done()
        stats, _ = s.getStatistics(stationID)
    }()

    wg.Wait()

    return &StationData{
        Devices: devices,
        Alarms:  alarms,
        Stats:   stats,
    }, nil
}
```

### 6.2 内存泄漏

#### 症状
- 内存使用持续增长
- 服务OOM被杀
- 性能逐渐下降

#### 排查步骤

```bash
# 监控内存使用
watch -n 1 'kubectl top pods -n nem-system'

# 获取内存profile
curl http://localhost:8080/debug/pprof/heap > heap.prof

# 分析内存
go tool pprof heap.prof
(pprof) top10
(pprof) list functionName

# 查看对象数量
curl http://localhost:8080/debug/pprof/heap?debug=1
```

#### 解决方案

**常见内存泄漏原因**

1. **未关闭的资源**

```go
// 错误示例
resp, err := http.Get(url)
// 忘记关闭resp.Body

// 正确示例
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()
```

2. **无限增长的Map**

```go
// 错误示例
var cache = make(map[string]string)
// 持续添加，从不删除

// 正确示例
var cache = make(map[string]string)
// 定期清理
for key := range cache {
    if expired(key) {
        delete(cache, key)
    }
}
```

3. **goroutine泄漏**

```go
// 错误示例
go func() {
    for {
        // 没有退出条件
        doSomething()
    }
}()

// 正确示例
ctx, cancel := context.WithCancel(context.Background())
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            doSomething()
        }
    }
}()
```

---

## 7. 网络问题

### 7.1 服务间通信失败

#### 症状
- 服务调用失败
- 连接超时
- 连接被拒绝

#### 排查步骤

```bash
# 检查Service
kubectl get svc -n nem-system

# 检查Endpoints
kubectl get endpoints -n nem-system

# 测试服务连接
kubectl exec -it <pod-name> -n nem-system -- sh
curl http://api-server:8080/health

# 检查DNS解析
kubectl exec -it <pod-name> -n nem-system -- nslookup api-server

# 检查网络策略
kubectl get networkpolicy -n nem-system
```

#### 解决方案

**问题1：Service没有Endpoints**

```bash
# 检查Pod标签
kubectl get pods -n nem-system --show-labels

# 检查Service selector
kubectl describe svc api-server -n nem-system

# 确保标签匹配
# Service selector: app=api-server
# Pod labels: app=api-server
```

**问题2：DNS解析失败**

```bash
# 检查CoreDNS
kubectl get pods -n kube-system -l k8s-app=kube-dns

# 查看CoreDNS日志
kubectl logs -n kube-system -l k8s-app=kube-dns

# 使用完整域名
# api-server.nem-system.svc.cluster.local
```

### 7.2 外部访问失败

#### 症状
- 外部无法访问服务
- Ingress不工作
- 域名解析失败

#### 排查步骤

```bash
# 检查Ingress
kubectl get ingress -n nem-system
kubectl describe ingress nem-ingress -n nem-system

# 检查Ingress Controller
kubectl get pods -n ingress-nginx
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx

# 检查Service类型
kubectl get svc -n nem-system

# 测试本地访问
kubectl port-forward svc/api-server 8080:8080 -n nem-system
```

#### 解决方案

**问题1：Ingress配置错误**

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nem-ingress
  namespace: nem-system
  annotations:
    kubernetes.io/ingress.class: nginx
spec:
  rules:
  - host: nem.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-server
            port:
              number: 8080
```

**问题2：TLS证书问题**

```bash
# 检查证书
kubectl get certificate -n nem-system
kubectl describe certificate nem-tls -n nem-system

# 查看证书详情
kubectl get secret nem-tls -n nem-system -o yaml
```

---

## 8. 存储问题

### 8.1 磁盘空间不足

#### 症状
- Pod无法启动
- 写入失败
- 日志显示磁盘空间错误

#### 排查步骤

```bash
# 查看节点磁盘使用
kubectl describe nodes | grep -A 5 "Allocated resources"

# 查看PV使用
kubectl get pv
kubectl describe pv <pv-name>

# 查看PVC
kubectl get pvc -n nem-system

# 进入Pod查看磁盘
kubectl exec -it <pod-name> -n nem-system -- df -h
```

#### 解决方案

**方案1：清理空间**

```bash
# 清理Docker镜像
docker image prune -a

# 清理未使用的卷
docker volume prune

# 清理日志
find /var/log -name "*.log" -mtime +7 -delete

# 清理容器日志
truncate -s 0 /var/lib/docker/containers/*/*-json.log
```

**方案2：扩容存储**

```bash
# 扩容PVC（需要StorageClass支持）
kubectl patch pvc postgres-data -n nem-system \
  -p '{"spec":{"resources":{"requests":{"storage":"200Gi"}}}}'
```

### 8.2 数据丢失

#### 症状
- 数据库数据丢失
- 配置丢失
- 重启后数据不存在

#### 排查步骤

```bash
# 检查PV绑定
kubectl get pv,pvc -n nem-system

# 检查存储类
kubectl get storageclass

# 检查数据卷
docker volume ls

# 查看Pod挂载
kubectl describe pod <pod-name> -n nem-system | grep -A 5 "Mounts"
```

#### 解决方案

**问题1：使用了emptyDir**

```yaml
# 错误示例
volumes:
- name: data
  emptyDir: {}  # Pod重启数据丢失

# 正确示例
volumes:
- name: data
  persistentVolumeClaim:
    claimName: postgres-data
```

**问题2：数据恢复**

```bash
# 从备份恢复
kubectl exec -it postgres-pod -n nem-system -- \
  pg_restore -U postgres -d nem_system /backup/nem_system.dump

# 从快照恢复（如果支持）
# 根据存储类提供的功能进行恢复
```

---

## 9. 安全问题

### 9.1 认证失败

#### 症状
- 登录失败
- Token无效
- 权限被拒绝

#### 排查步骤

```bash
# 查看认证日志
kubectl logs deployment/api-server -n nem-system | grep "auth"

# 检查JWT配置
kubectl get secret nem-secrets -n nem-system -o jsonpath='{.data.jwt-secret}' | base64 -d

# 测试Token
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/user/profile
```

#### 解决方案

**问题1：Token过期**

```bash
# 刷新Token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh-token>"}'
```

**问题2：JWT密钥不匹配**

```bash
# 确保所有服务使用相同的JWT密钥
kubectl get secret nem-secrets -n nem-system -o yaml

# 更新JWT密钥
kubectl create secret generic nem-secrets \
  --from-literal=jwt-secret=new-secret \
  --dry-run=client -o yaml | kubectl apply -f - -n nem-system

# 重启所有服务
kubectl rollout restart deployment -n nem-system
```

### 9.2 权限问题

#### 症状
- 访问被拒绝
- 操作无权限
- RBAC错误

#### 排查步骤

```bash
# 检查ServiceAccount
kubectl get serviceaccount -n nem-system

# 检查Role/RoleBinding
kubectl get role,rolebinding -n nem-system

# 检查ClusterRole/ClusterRoleBinding
kubectl get clusterrole,clusterrolebinding

# 查看用户权限
kubectl auth can-i list pods -n nem-system --as=system:serviceaccount:nem-system:nem-sa
```

#### 解决方案

```yaml
# 添加必要的权限
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: nem-role
  namespace: nem-system
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
```

---

## 10. 应急响应流程

### 10.1 P0级故障响应

**定义**：核心服务完全不可用，影响所有用户

**响应流程**：

```
1. 发现故障（监控告警/用户反馈）
   ↓
2. 确认故障（5分钟内）
   - 确认影响范围
   - 通知相关人员
   ↓
3. 快速恢复（15分钟内）
   - 尝试重启服务
   - 回滚到上一版本
   - 切换到备用系统
   ↓
4. 问题定位（30分钟内）
   - 收集日志
   - 分析根因
   ↓
5. 制定修复方案
   ↓
6. 实施修复
   ↓
7. 验证恢复
   ↓
8. 事后总结（24小时内）
```

### 10.2 应急操作手册

#### 服务完全不可用

```bash
# 1. 检查服务状态
kubectl get pods -n nem-system

# 2. 查看事件
kubectl get events -n nem-system --sort-by='.lastTimestamp'

# 3. 快速重启
kubectl rollout restart deployment -n nem-system

# 4. 回滚到上一版本
kubectl rollout undo deployment/api-server -n nem-system

# 5. 扩容服务
kubectl scale deployment api-server --replicas=5 -n nem-system
```

#### 数据库不可用

```bash
# 1. 检查数据库状态
kubectl exec -it postgres-pod -n nem-system -- pg_isready

# 2. 查看数据库日志
kubectl logs postgres-pod -n nem-system

# 3. 重启数据库
kubectl rollout restart statefulset/postgres -n nem-system

# 4. 从备份恢复
kubectl exec -it postgres-pod -n nem-system -- \
  pg_restore -U postgres -d nem_system /backup/latest.dump
```

#### 数据丢失

```bash
# 1. 立即停止写入
kubectl scale deployment api-server --replicas=0 -n nem-system

# 2. 评估损失
kubectl exec -it postgres-pod -n nem-system -- \
  psql -U postgres -d nem_system -c "SELECT count(*) FROM devices;"

# 3. 从备份恢复
kubectl exec -it postgres-pod -n nem-system -- \
  pg_restore -U postgres -d nem_system --clean /backup/latest.dump

# 4. 验证数据
# 5. 恢复服务
kubectl scale deployment api-server --replicas=2 -n nem-system
```

### 10.3 故障复盘模板

```markdown
# 故障复盘报告

## 基本信息
- 故障时间：YYYY-MM-DD HH:MM - HH:MM
- 故障等级：P0/P1/P2
- 影响范围：描述受影响的用户和功能
- 处理人员：列出参与处理的人员

## 故障描述
详细描述故障现象和影响

## 时间线
- HH:MM 发现故障
- HH:MM 确认故障
- HH:MM 开始处理
- HH:MM 故障恢复

## 根本原因
分析导致故障的根本原因

## 解决方案
描述采取的解决措施

## 改进措施
1. 短期改进（1周内完成）
2. 中期改进（1个月内完成）
3. 长期改进（3个月内完成）

## 经验教训
总结本次故障的经验和教训
```

---

## 11. 附录

### 11.1 常用诊断命令

```bash
# 系统资源
top                                    # CPU和内存
htop                                   # 增强版top
iostat                                 # I/O统计
vmstat                                 # 虚拟内存统计
df -h                                  # 磁盘空间
free -h                                # 内存使用

# 网络诊断
netstat -tlnp                          # 监听端口
ss -tlnp                               # socket统计
curl -v http://localhost:8080/health   # HTTP测试
telnet localhost 8080                  # 端口测试
nslookup api-server                    # DNS查询
ping api-server                        # 网络连通性

# 进程诊断
ps aux | grep api-server               # 查看进程
lsof -p <pid>                          # 打开的文件
strace -p <pid>                        # 系统调用跟踪

# 日志查看
tail -f /var/log/nem/app.log           # 实时日志
grep "ERROR" /var/log/nem/app.log      # 错误日志
journalctl -u api-server -f            # systemd日志
```

### 11.2 故障排查清单

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

### 11.3 应急联系人

| 角色 | 姓名 | 电话 | 邮箱 | 响应时间 |
|------|------|------|------|----------|
| 运维负责人 | - | - | ops@example.com | 5分钟 |
| 开发负责人 | - | - | dev@example.com | 10分钟 |
| DBA | - | - | dba@example.com | 10分钟 |
| 安全负责人 | - | - | security@example.com | 15分钟 |
| 架构师 | - | - | arch@example.com | 30分钟 |

### 11.4 变更记录

| 版本 | 日期 | 变更内容 | 变更人 |
|------|------|----------|--------|
| v1.0.0 | 2024-03-01 | 初始版本 | 运维团队 |
