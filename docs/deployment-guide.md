# 部署运维文档

## 文档信息

| 项目 | 内容 |
|------|------|
| 项目名称 | 新能源在线监控系统 |
| 文档版本 | v1.1.0 |
| 编写日期 | 2026-04-02 |
| 文档状态 | 正式发布 |

---

## 1. 环境准备

### 1.1 系统要求

#### 操作系统

| 操作系统 | 版本要求 | 架构 |
|----------|----------|------|
| CentOS | 7.6+ | x86_64 |
| Ubuntu | 20.04 LTS+ | x86_64 |
| Debian | 11+ | x86_64 |

#### 硬件要求

| 组件 | 最低配置 | 推荐配置 |
|------|----------|----------|
| CPU | 4核 | 8核+ |
| 内存 | 16GB | 32GB+ |
| 磁盘 | 500GB SSD | 1TB+ SSD |
| 网络 | 1Gbps | 10Gbps |

#### 生产环境资源规划

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

### 1.2 软件依赖

#### 必需软件

| 软件 | 版本要求 | 说明 |
|------|----------|------|
| Docker | 24.0+ | 容器运行时 |
| Docker Compose | 2.20+ | 容器编排工具 |
| Kubernetes | 1.28+ | 容器编排平台 |
| Helm | 3.12+ | Kubernetes包管理器 |
| Git | 2.40+ | 版本控制 |

#### 可选软件

| 软件 | 版本要求 | 说明 |
|------|----------|------|
| Nacos | 2.2+ | 服务注册与配置中心 |
| Prometheus | 2.45+ | 监控系统 |
| Grafana | 10.0+ | 可视化平台 |

### 1.3 网络要求

#### 端口规划

| 服务 | 端口 | 协议 | 说明 |
|------|------|------|------|
| api-server | 8080 | HTTP | API服务端口 |
| api-server | 9090 | HTTP | Prometheus指标端口 |
| PostgreSQL | 5432 | TCP | 数据库端口 |
| Redis | 6379 | TCP | 缓存端口 |
| Kafka | 9092 | TCP | 消息队列端口 |
| Doris FE | 9030 | HTTP | Doris FE HTTP端口 |
| Doris FE | 8030 | HTTP | Doris FE Web端口 |
| Doris BE | 8040 | HTTP | Doris BE Web端口 |
| Prometheus | 9091 | HTTP | 监控端口 |
| Grafana | 3000 | HTTP | 可视化端口 |

#### 防火墙配置

```bash
# 开放必要端口
firewall-cmd --permanent --add-port=8080/tcp
firewall-cmd --permanent --add-port=5432/tcp
firewall-cmd --permanent --add-port=6379/tcp
firewall-cmd --permanent --add-port=9092/tcp
firewall-cmd --permanent --add-port=9030/tcp
firewall-cmd --reload
```

### 1.4 环境检查脚本

```bash
#!/bin/bash
# check_environment.sh

echo "=== 环境检查 ==="

# 检查操作系统
echo "操作系统: $(cat /etc/os-release | grep PRETTY_NAME)"

# 检查内核版本
echo "内核版本: $(uname -r)"

# 检查Docker
if command -v docker &> /dev/null; then
    echo "Docker版本: $(docker --version)"
else
    echo "错误: Docker未安装"
    exit 1
fi

# 检查Docker Compose
if command -v docker-compose &> /dev/null; then
    echo "Docker Compose版本: $(docker-compose --version)"
else
    echo "错误: Docker Compose未安装"
    exit 1
fi

# 检查kubectl
if command -v kubectl &> /dev/null; then
    echo "kubectl版本: $(kubectl version --client --short)"
fi

# 检查Helm
if command -v helm &> /dev/null; then
    echo "Helm版本: $(helm version --short)"
fi

# 检查系统资源
echo ""
echo "=== 系统资源 ==="
echo "CPU核心数: $(nproc)"
echo "内存总量: $(free -h | grep Mem | awk '{print $2}')"
echo "磁盘空间: $(df -h / | tail -1 | awk '{print $4}')"

echo ""
echo "环境检查完成"
```

---

## 2. Docker部署指南

### 2.1 快速开始

#### 克隆代码

```bash
git clone https://github.com/new-energy-monitoring/new-energy-monitoring.git
cd new-energy-monitoring
```

#### 配置文件准备

```bash
# 复制配置文件模板
cp configs/config.yaml configs/config-prod.yaml

# 编辑配置文件
vim configs/config-prod.yaml
```

#### 启动服务

```bash
# 启动所有服务
docker-compose -f deployments/docker/docker-compose.yml up -d

# 查看服务状态
docker-compose -f deployments/docker/docker-compose.yml ps

# 查看日志
docker-compose -f deployments/docker/docker-compose.yml logs -f api-server
```

### 2.2 镜像构建

#### 构建所有服务镜像

```bash
# 使用Makefile构建
make docker-build

# 或手动构建
docker build -t new-energy-monitoring/api-server:1.0.0 \
  -f deployments/docker/Dockerfile \
  --build-arg SERVICE=api-server .
```

#### Dockerfile示例

```dockerfile
# deployments/docker/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git make

# 复制go.mod和go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api-server ./cmd/api-server

# 运行阶段
FROM alpine:3.19

WORKDIR /app

# 安装ca证书
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 复制二进制文件
COPY --from=builder /app/bin/api-server /app/
COPY --from=builder /app/configs /app/configs

# 暴露端口
EXPOSE 8080 9090

# 运行
ENTRYPOINT ["/app/api-server"]
CMD ["-config", "/app/configs/config.yaml"]
```

### 2.3 Docker Compose配置详解

```yaml
# deployments/docker/docker-compose.yml
version: '3.8'

services:
  # PostgreSQL数据库
  postgres:
    image: postgres:16-alpine
    container_name: nem-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD:-postgres}
      POSTGRES_DB: nem_system
      PGDATA: /var/lib/postgresql/data/pgdata
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ../../scripts/migrations:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - nem-network
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G
        reservations:
          cpus: '2'
          memory: 4G

  # Redis缓存
  redis:
    image: redis:7-alpine
    container_name: nem-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: >
      redis-server
      --appendonly yes
      --maxmemory 4gb
      --maxmemory-policy allkeys-lru
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - nem-network

  # Kafka消息队列
  kafka:
    image: bitnami/kafka:3.7
    container_name: nem-kafka
    ports:
      - "9092:9092"
      - "9093:9093"
    environment:
      KAFKA_CFG_NODE_ID: 1
      KAFKA_CFG_PROCESS_ROLES: broker,controller
      KAFKA_CFG_CONTROLLER_QUORUM_VOTER: 1@kafka:9093
      KAFKA_CFG_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093
      KAFKA_CFG_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092
      KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP: CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
      KAFKA_CFG_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_CFG_INTER_BROKER_LISTENER_NAME: PLAINTEXT
      KAFKA_CFG_AUTO_CREATE_TOPICS_ENABLE: "true"
      KAFKA_CFG_NUM_PARTITIONS: 12
      KAFKA_CFG_DEFAULT_REPLICATION_FACTOR: 1
    volumes:
      - kafka_data:/bitnami/kafka
    networks:
      - nem-network

  # Doris时序数据库
  doris-fe:
    image: apache/doris:2.0-fe
    container_name: nem-doris-fe
    hostname: doris-fe
    ports:
      - "9030:9030"
      - "8030:8030"
    environment:
      FE_SERVERS: fe1:doris-fe:9010
      FE_ID: 1
    networks:
      - nem-network

  doris-be:
    image: apache/doris:2.0-be
    container_name: nem-doris-be
    hostname: doris-be
    ports:
      - "8040:8040"
    environment:
      FE_SERVERS: fe1:doris-fe:9010
      BE_ADDR: doris-be:9050
    depends_on:
      - doris-fe
    networks:
      - nem-network

  # API服务
  api-server:
    build:
      context: ../..
      dockerfile: deployments/docker/Dockerfile
      args:
        SERVICE: api-server
    image: new-energy-monitoring/api-server:${VERSION:-1.0.0}
    container_name: nem-api-server
    ports:
      - "8080:8080"
      - "9090:9090"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      - GIN_MODE=release
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD:-postgres}
      - DB_NAME=nem_system
      - REDIS_ADDR=redis:6379
      - KAFKA_BROKERS=kafka:9092
    volumes:
      - ../../configs:/app/configs:ro
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - nem-network
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3

  # 采集服务
  collector:
    build:
      context: ../..
      dockerfile: deployments/docker/Dockerfile
      args:
        SERVICE: collector
    image: new-energy-monitoring/collector:${VERSION:-1.0.0}
    container_name: nem-collector
    depends_on:
      - kafka
      - redis
    environment:
      - REDIS_ADDR=redis:6379
      - KAFKA_BROKERS=kafka:9092
    volumes:
      - ../../configs:/app/configs:ro
    networks:
      - nem-network
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 4G

  # 告警服务
  alarm:
    build:
      context: ../..
      dockerfile: deployments/docker/Dockerfile
      args:
        SERVICE: alarm
    image: new-energy-monitoring/alarm:${VERSION:-1.0.0}
    container_name: nem-alarm
    depends_on:
      - kafka
      - redis
      - postgres
    environment:
      - DB_HOST=postgres
      - REDIS_ADDR=redis:6379
      - KAFKA_BROKERS=kafka:9092
    volumes:
      - ../../configs:/app/configs:ro
    networks:
      - nem-network

  # 计算服务
  compute:
    build:
      context: ../..
      dockerfile: deployments/docker/Dockerfile
      args:
        SERVICE: compute
    image: new-energy-monitoring/compute:${VERSION:-1.0.0}
    container_name: nem-compute
    depends_on:
      - redis
      - postgres
    environment:
      - DB_HOST=postgres
      - REDIS_ADDR=redis:6379
    volumes:
      - ../../configs:/app/configs:ro
    networks:
      - nem-network

  # AI服务
  ai-service:
    build:
      context: ../..
      dockerfile: deployments/docker/Dockerfile
      args:
        SERVICE: ai-service
    image: new-energy-monitoring/ai-service:${VERSION:-1.0.0}
    container_name: nem-ai-service
    depends_on:
      - redis
      - postgres
    environment:
      - DB_HOST=postgres
      - REDIS_ADDR=redis:6379
    volumes:
      - ../../configs:/app/configs:ro
    networks:
      - nem-network
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G

  # 调度服务
  scheduler:
    build:
      context: ../..
      dockerfile: deployments/docker/Dockerfile
      args:
        SERVICE: scheduler
    image: new-energy-monitoring/scheduler:${VERSION:-1.0.0}
    container_name: nem-scheduler
    depends_on:
      - redis
      - postgres
    environment:
      - DB_HOST=postgres
      - REDIS_ADDR=redis:6379
    volumes:
      - ../../configs:/app/configs:ro
    networks:
      - nem-network

volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local
  kafka_data:
    driver: local

networks:
  nem-network:
    driver: bridge
```

### 2.4 常用命令

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

# 备份数据卷
docker run --rm -v nem-postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres_backup.tar.gz /data
```

---

## 3. Kubernetes部署指南

### 3.1 前置条件

#### Kubernetes集群要求

- Kubernetes版本：1.28+
- 节点数量：≥3（生产环境）
- 存储类：支持动态供给
- Ingress控制器：nginx-ingress

#### 安装Helm

```bash
# macOS
brew install helm

# Linux
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

### 3.2 使用Helm部署

#### 添加依赖仓库

```bash
# 添加Bitnami仓库
helm repo add bitnami https://charts.bitnami.com/bitnami

# 更新仓库
helm repo update
```

#### 创建命名空间

```bash
kubectl create namespace nem-system
```

#### 配置values.yaml

```bash
# 复制默认配置
cp deployments/kubernetes/helm/values.yaml values-prod.yaml

# 编辑配置
vim values-prod.yaml
```

#### 安装Chart

```bash
# 安装
helm install nem-system ./deployments/kubernetes/helm \
  -n nem-system \
  -f values-prod.yaml

# 或使用--set参数覆盖
helm install nem-system ./deployments/kubernetes/helm \
  -n nem-system \
  --set apiServer.replicaCount=3 \
  --set postgresql.auth.postgresPassword=your-password
```

#### 升级Chart

```bash
helm upgrade nem-system ./deployments/kubernetes/helm \
  -n nem-system \
  -f values-prod.yaml
```

#### 卸载Chart

```bash
helm uninstall nem-system -n nem-system
```

### 3.3 Helm Chart结构

```
deployments/kubernetes/helm/
├── Chart.yaml              # Chart元数据
├── values.yaml             # 默认配置值
├── templates/
│   ├── _helpers.tpl        # 模板助手函数
│   ├── api-server-deployment.yaml
│   ├── api-server-service.yaml
│   ├── ingress.yaml
│   └── ...
└── charts/                 # 依赖的子Chart
```

### 3.4 Kubernetes资源示例

#### Deployment

```yaml
# api-server-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
  namespace: nem-system
  labels:
    app: api-server
spec:
  replicas: 2
  selector:
    matchLabels:
      app: api-server
  template:
    metadata:
      labels:
        app: api-server
    spec:
      containers:
      - name: api-server
        image: new-energy-monitoring/api-server:1.0.0
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: GIN_MODE
          value: release
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: nem-secrets
              key: db-host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: nem-secrets
              key: db-password
        resources:
          limits:
            cpu: 1000m
            memory: 1Gi
          requests:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: config
          mountPath: /app/configs
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: nem-config
```

#### Service

```yaml
# api-server-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: api-server
  namespace: nem-system
spec:
  type: ClusterIP
  ports:
  - port: 8080
    targetPort: 8080
    name: http
  - port: 9090
    targetPort: 9090
    name: metrics
  selector:
    app: api-server
```

#### Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nem-ingress
  namespace: nem-system
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - nem.example.com
    secretName: nem-tls
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

#### ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: nem-config
  namespace: nem-system
data:
  config.yaml: |
    server:
      name: api-server
      port: 8080
      mode: release
    
    database:
      type: postgres
      host: postgres.nem-system.svc.cluster.local
      port: 5432
      user: postgres
      dbname: nem_system
      max_open_conns: 100
      max_idle_conns: 10
    
    redis:
      addrs:
        - redis-master.nem-system.svc.cluster.local:6379
      db: 0
      pool_size: 100
    
    logging:
      level: info
      format: json
```

#### Secret

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: nem-secrets
  namespace: nem-system
type: Opaque
stringData:
  db-host: postgres.nem-system.svc.cluster.local
  db-password: your-secure-password
  redis-password: your-redis-password
  jwt-secret: your-jwt-secret-key
```

### 3.5 自动扩缩容配置

```yaml
# hpa.yaml
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

### 3.6 常用kubectl命令

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

# 查看事件
kubectl get events -n nem-system --sort-by='.lastTimestamp'

# 扩缩容
kubectl scale deployment api-server --replicas=5 -n nem-system

# 重启Pod
kubectl rollout restart deployment api-server -n nem-system

# 查看资源使用
kubectl top pods -n nem-system
```

---

## 4. 配置说明

### 4.1 配置文件结构

```
configs/
├── config.yaml              # 默认配置
├── config-dev.yaml          # 开发环境配置
├── config-test.yaml         # 测试环境配置
├── config-prod.yaml         # 生产环境配置
└── config-standalone.yaml   # 单机部署配置
```

### 4.2 配置项详解

#### 服务配置

```yaml
server:
  name: api-server           # 服务名称
  port: 8080                 # 服务端口
  mode: release              # 运行模式: debug/release
  graceful_shutdown: 30      # 优雅关闭超时时间(秒)
```

#### 数据库配置

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
  log_level: error           # 日志级别
```

#### Redis配置

```yaml
redis:
  addrs:                     # 集群地址列表
    - localhost:6379
  password: ""               # 密码
  db: 0                      # 数据库索引
  pool_size: 100             # 连接池大小
  min_idle_conns: 10         # 最小空闲连接数
  max_retries: 3             # 最大重试次数
  dial_timeout: 5            # 连接超时(秒)
  read_timeout: 3            # 读超时(秒)
  write_timeout: 3           # 写超时(秒)
```

#### Kafka配置

```yaml
kafka:
  brokers:                   # Broker地址列表
    - localhost:9092
  topic_prefix: nem          # 主题前缀
  consumer_group: nem-group  # 消费者组
  auto_offset_reset: latest  # 偏移量重置策略
  enable_auto_commit: true   # 自动提交偏移量
```

#### 时序数据库配置

```yaml
timeseries:
  type: doris                # 类型: doris/clickhouse
  host: localhost            # 主机地址
  port: 9030                 # 端口
  user: root                 # 用户名
  password: ""               # 密码
  database: nem_ts           # 数据库名
  max_open_conns: 50         # 最大连接数
```

#### 日志配置

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

#### 认证配置

```yaml
auth:
  jwt:
    secret: your-jwt-secret-key-change-in-production
    access_expire: 7200      # Access Token过期时间(秒)
    refresh_expire: 604800   # Refresh Token过期时间(秒)
    issuer: nem-system       # 签发者
  password:
    min_length: 8            # 最小长度
    require_uppercase: true  # 需要大写字母
    require_lowercase: true  # 需要小写字母
    require_digit: true      # 需要数字
    require_special: false   # 需要特殊字符
  login:
    max_attempts: 5          # 最大尝试次数
    lock_duration: 1800      # 锁定时长(秒)
```

#### 监控配置

```yaml
tracing:
  enabled: true              # 是否启用链路追踪
  endpoint: localhost:4317   # OTLP端点
  sampler_ratio: 0.1         # 采样率

metrics:
  enabled: true              # 是否启用指标采集
  port: 9090                 # Prometheus指标端口
  path: /metrics             # 指标路径
```

### 4.3 环境变量覆盖

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

### 4.4 配置中心集成

#### Nacos配置

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

## 5. 故障排查指南

### 5.1 常见问题

#### 服务无法启动

**症状**: 服务启动失败，日志显示连接错误

**排查步骤**:

```bash
# 1. 检查服务日志
docker-compose logs api-server

# 2. 检查依赖服务状态
docker-compose ps

# 3. 检查网络连接
docker-compose exec api-server ping postgres

# 4. 检查配置文件
docker-compose exec api-server cat /app/configs/config.yaml

# 5. 检查端口占用
netstat -tlnp | grep 8080
```

**解决方案**:

- 确保依赖服务已启动
- 检查配置文件中的连接地址
- 检查网络连通性
- 检查防火墙规则

#### 数据库连接失败

**症状**: 日志显示数据库连接错误

**排查步骤**:

```bash
# 1. 检查PostgreSQL状态
docker-compose exec postgres pg_isready

# 2. 测试数据库连接
docker-compose exec postgres psql -U postgres -d nem_system

# 3. 检查连接数
docker-compose exec postgres psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

# 4. 检查最大连接数
docker-compose exec postgres psql -U postgres -c "SHOW max_connections;"
```

**解决方案**:

- 检查数据库用户名密码
- 增加最大连接数配置
- 检查连接池配置

#### 内存不足

**症状**: 服务OOM被杀

**排查步骤**:

```bash
# 1. 查看容器资源使用
docker stats

# 2. 查看内存限制
docker inspect <container-id> | grep -i memory

# 3. 查看系统内存
free -h

# 4. 查看进程内存
docker-compose exec api-server top
```

**解决方案**:

- 增加容器内存限制
- 优化应用内存使用
- 检查内存泄漏

#### 告警未触发

**症状**: 数据超过阈值但未触发告警

**排查步骤**:

```bash
# 1. 检查告警服务日志
docker-compose logs alarm

# 2. 检查Kafka消息
docker-compose exec kafka kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic nem.data.collect \
  --from-beginning

# 3. 检查告警规则
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/alarm/rules

# 4. 检查Redis中的实时数据
docker-compose exec redis redis-cli GET "nem:realtime:point-001"
```

### 5.2 日志分析

#### 日志位置

```bash
# Docker日志
docker-compose logs -f api-server

# Kubernetes日志
kubectl logs -f <pod-name> -n nem-system

# 文件日志（如果配置了文件输出）
tail -f /var/log/nem/app.log
```

#### 日志级别调整

```bash
# 临时调整日志级别
curl -X PUT http://localhost:8080/api/v1/admin/log/level \
  -H "Authorization: Bearer <token>" \
  -d '{"level":"debug"}'
```

#### 日志格式示例

```json
{
  "level": "error",
  "ts": "2024-03-01T10:30:00.000Z",
  "caller": "service/device_service.go:45",
  "msg": "Failed to get device",
  "error": "device not found",
  "device_id": "device-001",
  "trace_id": "abc123",
  "span_id": "def456"
}
```

### 5.3 性能问题排查

#### 慢查询分析

```sql
-- PostgreSQL慢查询
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- 开启慢查询日志
ALTER SYSTEM SET log_min_duration_statement = 1000;
SELECT pg_reload_conf();
```

#### 接口响应慢

```bash
# 1. 检查API响应时间
curl -w "Time: %{time_total}s\n" http://localhost:8080/api/v1/stations

# 2. 检查Prometheus指标
curl http://localhost:9090/metrics | grep http_request_duration

# 3. 检查链路追踪
# 访问Jaeger UI查看trace
```

#### 内存泄漏排查

```bash
# 1. 获取内存profile
curl http://localhost:8080/debug/pprof/heap > heap.out

# 2. 分析profile
go tool pprof heap.out

# 3. 查看top内存使用
(pprof) top10

# 4. 查看具体函数
(pprof) list functionName
```

### 5.4 数据恢复

#### PostgreSQL数据恢复

```bash
# 从备份恢复
docker-compose exec postgres pg_restore \
  -U postgres \
  -d nem_system \
  /backup/nem_system_backup.dump

# 从SQL文件恢复
docker-compose exec -T postgres psql -U postgres nem_system < backup.sql
```

#### Redis数据恢复

```bash
# 从RDB文件恢复
docker cp dump.rdb nem-redis:/data/
docker-compose restart redis
```

---

## 6. 监控告警配置

### 6.1 Prometheus配置

#### prometheus.yml

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

### 6.2 告警规则

#### rules/alerts.yml

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

### 6.3 Alertmanager配置

#### alertmanager.yml

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

### 6.4 Grafana Dashboard

#### 导入Dashboard

```bash
# 导入预置Dashboard
# Dashboard ID: 12345 (示例)
# 或导入JSON文件
```

#### 关键监控面板

- **服务概览**: 请求量、响应时间、错误率
- **资源使用**: CPU、内存、磁盘、网络
- **业务指标**: 采集点数、告警数、设备在线率
- **数据库**: 连接数、查询性能、慢查询
- **消息队列**: 生产/消费速率、延迟

---

## 7. 备份与恢复

### 7.1 数据备份策略

| 数据类型 | 备份方式 | 频率 | 保留周期 |
|----------|----------|------|----------|
| PostgreSQL | 全量+增量 | 每日全量+实时WAL | 30天 |
| Redis | RDB+AOF | 每小时RDB | 7天 |
| Doris | 快照 | 每日 | 30天 |
| 配置文件 | Git版本控制 | 实时 | 永久 |

### 7.2 备份脚本

#### PostgreSQL备份

```bash
#!/bin/bash
# backup_postgres.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/postgres"
DB_NAME="nem_system"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 全量备份
docker-compose exec -T postgres pg_dump -U postgres $DB_NAME | gzip > $BACKUP_DIR/${DB_NAME}_${DATE}.sql.gz

# 删除30天前的备份
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete

echo "Backup completed: ${DB_NAME}_${DATE}.sql.gz"
```

#### Redis备份

```bash
#!/bin/bash
# backup_redis.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/redis"

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

### 7.3 定时备份配置

```bash
# crontab -e

# PostgreSQL每日备份 (凌晨2点)
0 2 * * * /opt/scripts/backup_postgres.sh >> /var/log/backup.log 2>&1

# Redis每小时备份
0 * * * * /opt/scripts/backup_redis.sh >> /var/log/backup.log 2>&1
```

---

## 8. 安全加固

### 8.1 网络安全

```bash
# 配置防火墙规则
firewall-cmd --permanent --add-rich-rule='rule family="ipv4" source address="10.0.0.0/8" port protocol="tcp" port="5432" accept'
firewall-cmd --permanent --add-rich-rule='rule family="ipv4" source address="10.0.0.0/8" port protocol="tcp" port="6379" accept'
firewall-cmd --reload
```

### 8.2 数据库安全

```sql
-- 创建应用专用用户
CREATE USER nem_app WITH PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE nem_system TO nem_app;
GRANT USAGE ON SCHEMA public TO nem_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO nem_app;

-- 禁用超级用户远程登录
-- 编辑 pg_hba.conf
```

### 8.3 密钥管理

```bash
# 使用Kubernetes Secret管理敏感信息
kubectl create secret generic nem-secrets \
  --from-literal=db-password=secure-password \
  --from-literal=redis-password=secure-password \
  --from-literal=jwt-secret=secure-jwt-secret \
  -n nem-system

# 或使用外部密钥管理系统
# HashiCorp Vault, AWS Secrets Manager等
```

---

## 9. 发布流程

### 9.1 版本管理

#### 版本号规范

采用语义化版本号：`MAJOR.MINOR.PATCH`

- **MAJOR**：不兼容的API变更
- **MINOR**：向后兼容的功能新增
- **PATCH**：向后兼容的问题修复

示例：
- `v1.0.0`：初始版本
- `v1.1.0`：新增功能
- `v1.1.1`：Bug修复

#### 分支管理

```
main (生产分支)
  ├── develop (开发分支)
  │   ├── feature/xxx (功能分支)
  │   ├── bugfix/xxx (修复分支)
  │   └── release/v1.1.0 (发布分支)
  └── hotfix/xxx (紧急修复分支)
```

### 9.2 发布流程

#### 标准发布流程

```
1. 代码开发
   ↓
2. 代码审查
   ↓
3. 自动化测试
   ↓
4. 构建镜像
   ↓
5. 部署到测试环境
   ↓
6. 测试验证
   ↓
7. 部署到预发布环境
   ↓
8. 生产环境发布
   ↓
9. 发布后验证
   ↓
10. 监控观察
```

#### CI/CD流程

**GitHub Actions配置示例**

```yaml
# .github/workflows/cd.yml
name: CD Pipeline

on:
  push:
    branches:
      - main
      - release/*

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Run tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...

    - name: Build Docker image
      run: |
        docker build -t ${{ secrets.REGISTRY }}/nem-api-server:${{ github.sha }} \
          -f deployments/docker/Dockerfile \
          --build-arg SERVICE=api-server .

    - name: Push to registry
      run: |
        echo ${{ secrets.REGISTRY_PASSWORD }} | docker login -u ${{ secrets.REGISTRY_USER }} --password-stdin ${{ secrets.REGISTRY }}
        docker push ${{ secrets.REGISTRY }}/nem-api-server:${{ github.sha }}

    - name: Deploy to staging
      if: github.ref == 'refs/heads/develop'
      run: |
        kubectl set image deployment/api-server \
          api-server=${{ secrets.REGISTRY }}/nem-api-server:${{ github.sha }} \
          -n nem-staging

    - name: Deploy to production
      if: github.ref == 'refs/heads/main'
      run: |
        kubectl set image deployment/api-server \
          api-server=${{ secrets.REGISTRY }}/nem-api-server:${{ github.sha }} \
          -n nem-system
```

### 9.3 发布策略

#### 滚动发布

```bash
# Kubernetes默认使用滚动更新
kubectl set image deployment/api-server \
  api-server=nem-api-server:v1.1.0 \
  -n nem-system

# 查看更新状态
kubectl rollout status deployment/api-server -n nem-system

# 查看更新历史
kubectl rollout history deployment/api-server -n nem-system
```

**滚动更新配置**

```yaml
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1          # 最多可以超出期望副本数的数量
      maxUnavailable: 0    # 最多不可用的副本数
```

#### 蓝绿发布

```bash
# 1. 部署新版本（绿色环境）
kubectl apply -f deployment-api-server-green.yaml -n nem-system

# 2. 验证新版本
kubectl exec -it api-server-green-pod -n nem-system -- curl http://localhost:8080/health

# 3. 切换流量到新版本
kubectl patch service api-server -n nem-system -p '{"spec":{"selector":{"version":"green"}}}'

# 4. 观察一段时间后删除旧版本
kubectl delete deployment api-server-blue -n nem-system
```

#### 金丝雀发布

```bash
# 1. 部署金丝雀版本（10%流量）
kubectl apply -f deployment-api-server-canary.yaml -n nem-system

# 2. 逐步增加流量
# 10% -> 25% -> 50% -> 100%
kubectl scale deployment api-server-canary --replicas=1 -n nem-system
kubectl scale deployment api-server --replicas=9 -n nem-system

# 3. 监控指标
# 观察错误率、响应时间等指标

# 4. 完成发布
kubectl scale deployment api-server --replicas=10 -n nem-system
kubectl delete deployment api-server-canary -n nem-system
```

### 9.4 发布检查清单

#### 发布前检查

- [ ] 代码已合并到发布分支
- [ ] 所有测试通过
- [ ] 代码审查完成
- [ ] 变更日志已更新
- [ ] 数据库迁移脚本已准备
- [ ] 配置文件已更新
- [ ] 回滚方案已准备
- [ ] 相关人员已通知

#### 发布中监控

- [ ] 服务健康检查
- [ ] 错误率监控
- [ ] 响应时间监控
- [ ] 资源使用监控
- [ ] 业务指标监控

#### 发布后验证

- [ ] 功能验证测试
- [ ] 性能测试
- [ ] 日志检查
- [ ] 告警检查
- [ ] 用户反馈

### 9.5 回滚操作

#### 快速回滚

```bash
# 回滚到上一版本
kubectl rollout undo deployment/api-server -n nem-system

# 回滚到指定版本
kubectl rollout undo deployment/api-server --to-revision=2 -n nem-system

# 查看回滚状态
kubectl rollout status deployment/api-server -n nem-system
```

#### Helm回滚

```bash
# 查看发布历史
helm history nem-system -n nem-system

# 回滚到指定版本
helm rollback nem-system 2 -n nem-system

# 查看回滚状态
helm status nem-system -n nem-system
```

#### 数据库回滚

```bash
# 执行数据库回滚脚本
kubectl exec -it migrate-pod -n nem-system -- \
  ./migrate -database "postgres://user:pass@postgres:5432/nem_system?sslmode=disable" \
  -path /migrations \
  down 1

# 恢复数据备份（如果需要）
kubectl exec -it postgres-pod -n nem-system -- \
  pg_restore -U postgres -d nem_system --clean /backup/pre-release.dump
```

### 9.6 发布窗口

#### 常规发布

- **时间**：每周二、周四 10:00-18:00
- **要求**：
  - 避开业务高峰期
  - 确保相关人员在线
  - 提前1天通知

#### 紧急发布

- **条件**：
  - 修复P0级故障
  - 安全漏洞修复
  - 数据丢失风险
- **流程**：
  - 立即通知相关人员
  - 快速评审
  - 发布并验证

#### 发布冻结期

- **时间**：每月最后3个工作日
- **原因**：月度结算、报表生成
- **例外**：紧急修复需CTO审批

### 9.7 发布通知模板

#### 发布前通知

```
主题：【发布通知】新能源监控系统 v1.1.0 发布计划

各位同事：

我们计划于 YYYY-MM-DD HH:MM 进行系统发布，详情如下：

发布版本：v1.1.0
发布时间：YYYY-MM-DD HH:MM - HH:MM
发布内容：
1. 新增功能A
2. 优化功能B
3. 修复问题C

影响范围：
- API服务将短暂不可用（预计5分钟）
- 数据查询功能将受影响

注意事项：
- 请提前保存工作内容
- 发布期间请勿进行重要操作

联系人：XXX
联系电话：XXX

运维团队
YYYY-MM-DD
```

#### 发布完成通知

```
主题：【发布完成】新能源监控系统 v1.1.0 发布成功

各位同事：

系统发布已完成，详情如下：

发布版本：v1.1.0
发布时间：YYYY-MM-DD HH:MM - HH:MM
发布结果：成功

验证情况：
- ✓ 服务健康检查通过
- ✓ 功能验证测试通过
- ✓ 性能指标正常

如遇问题，请联系：XXX

运维团队
YYYY-MM-DD
```

---

## 10. 附录

### 10.1 常用命令速查

```bash
# Docker
docker-compose up -d                    # 启动服务
docker-compose down                     # 停止服务
docker-compose logs -f api-server       # 查看日志
docker-compose restart api-server       # 重启服务
docker stats                            # 资源使用

# Kubernetes
kubectl get pods -n nem-system          # 查看Pod
kubectl logs -f <pod> -n nem-system     # 查看日志
kubectl exec -it <pod> -n nem-system -- sh  # 进入容器
kubectl describe pod <pod> -n nem-system    # Pod详情
kubectl get svc -n nem-system           # 查看服务
kubectl get ingress -n nem-system       # 查看Ingress

# Helm
helm install nem-system ./helm -n nem-system    # 安装
helm upgrade nem-system ./helm -n nem-system    # 升级
helm rollback nem-system 1 -n nem-system        # 回滚
helm uninstall nem-system -n nem-system         # 卸载

# 数据库
psql -U postgres -d nem_system          # 连接数据库
pg_dump -U postgres nem_system > backup.sql  # 备份
pg_restore -U postgres -d nem_system backup.dump  # 恢复
```

### 10.2 故障排查清单

- [ ] 检查服务状态
- [ ] 检查日志输出
- [ ] 检查网络连通性
- [ ] 检查配置文件
- [ ] 检查资源使用
- [ ] 检查依赖服务
- [ ] 检查防火墙规则
- [ ] 检查证书有效期

### 10.3 变更记录

| 版本 | 日期 | 变更内容 | 变更人 |
|------|------|----------|--------|
| v1.0.0 | 2024-03-01 | 初始版本 | 运维团队 |
| v1.1.0 | 2024-03-15 | 添加发布流程章节 | 运维团队 |
