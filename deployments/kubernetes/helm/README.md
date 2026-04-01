# 新能源监控系统 Helm Chart 部署说明

## 概述

本Helm Chart用于在Kubernetes集群中部署新能源在线监控系统的完整微服务架构。

## 前置条件

- Kubernetes 1.20+
- Helm 3.0+
- PV provisioner支持（用于持久化存储）
- Ingress Controller（如nginx-ingress）
- Prometheus Operator（可选，用于监控）

## 架构组件

### 微服务
- **API Server**: 核心API服务，提供RESTful接口
- **Collector**: 数据采集服务，支持IEC104、Modbus、IEC61850协议
- **Alarm Service**: 告警服务，处理告警规则和通知
- **Compute Service**: 计算服务，执行公式计算和规则引擎
- **AI Service**: AI服务，提供智能分析和推理能力
- **Scheduler**: 调度服务，管理定时任务和统计计算

### 基础设施
- **PostgreSQL**: 主数据库
- **Redis**: 缓存和消息队列
- **Kafka**: 消息中间件

## 安装

### 1. 添加依赖仓库

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

### 2. 更新依赖

```bash
cd deployments/kubernetes/helm
helm dependency update
```

### 3. 开发环境部署

```bash
# 使用开发环境配置部署
helm install nem . -f values-dev.yaml -n new-energy-monitoring-dev --create-namespace
```

### 4. 生产环境部署

```bash
# 使用生产环境配置部署
helm install nem . -f values-prod.yaml -n new-energy-monitoring --create-namespace
```

### 5. 自定义部署

```bash
# 创建自定义values文件
helm install nem . -f my-values.yaml -n new-energy-monitoring --create-namespace
```

## 配置说明

### 全局配置

```yaml
global:
  imageRegistry: "your-registry.com"  # 镜像仓库地址
  imagePullSecrets:                   # 镜像拉取密钥
    - name: registry-secret
  storageClass: "standard"            # 存储类
  namespace: "new-energy-monitoring"  # 命名空间
```

### 服务配置

每个服务都支持以下配置项：

```yaml
serviceName:
  replicaCount: 3                     # 副本数
  image:
    repository: nem/service-name      # 镜像仓库
    tag: "1.0.0"                      # 镜像标签
    pullPolicy: IfNotPresent          # 拉取策略
  
  service:
    type: ClusterIP                   # 服务类型
    port: 8080                        # 服务端口
  
  resources:                          # 资源限制
    limits:
      cpu: 2000m
      memory: 4Gi
    requests:
      cpu: 500m
      memory: 1Gi
  
  autoscaling:                        # 自动扩缩容
    enabled: true
    minReplicas: 2
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
  
  healthCheck:                        # 健康检查
    liveness:
      enabled: true
      path: /health/live
      port: 8080
    readiness:
      enabled: true
      path: /health/ready
      port: 8080
```

### 数据库配置

```yaml
postgresql:
  enabled: true                       # 是否启用内置PostgreSQL
  auth:
    postgresPassword: "postgres"      # postgres用户密码
    username: "nem"                   # 应用用户名
    password: "nem123456"             # 应用用户密码
    database: "nem_system"            # 数据库名
  primary:
    persistence:
      enabled: true
      size: 100Gi                     # 存储大小
```

### Redis配置

```yaml
redis:
  enabled: true                       # 是否启用内置Redis
  auth:
    enabled: true
    password: "redis123456"
  master:
    persistence:
      enabled: true
      size: 20Gi
  replica:
    replicaCount: 2                   # 从节点数量
```

### Kafka配置

```yaml
kafka:
  enabled: true                       # 是否启用内置Kafka
  replicaCount: 3                     # Kafka节点数
  persistence:
    enabled: true
    size: 50Gi
  zookeeper:
    enabled: true
    replicaCount: 3                   # Zookeeper节点数
```

### Ingress配置

```yaml
ingress:
  enabled: true
  className: nginx
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: nem.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: nem-tls
      hosts:
        - nem.example.com
```

### 监控配置

```yaml
monitoring:
  enabled: true
  serviceMonitor:
    enabled: true                     # 启用Prometheus ServiceMonitor
    interval: 30s
  prometheusRule:
    enabled: true                     # 启用Prometheus告警规则
```

## 升级

```bash
# 升级到新版本
helm upgrade nem . -f values-prod.yaml -n new-energy-monitoring

# 强制重启Pod
helm upgrade nem . -f values-prod.yaml -n new-energy-monitoring --force
```

## 回滚

```bash
# 查看历史版本
helm history nem -n new-energy-monitoring

# 回滚到指定版本
helm rollback nem 1 -n new-energy-monitoring
```

## 卸载

```bash
# 卸载Release
helm uninstall nem -n new-energy-monitoring

# 删除命名空间（可选）
kubectl delete namespace new-energy-monitoring
```

## 验证部署

### 检查Pod状态

```bash
kubectl get pods -n new-energy-monitoring
```

### 检查服务状态

```bash
kubectl get svc -n new-energy-monitoring
```

### 检查Ingress

```bash
kubectl get ingress -n new-energy-monitoring
```

### 查看日志

```bash
# 查看API Server日志
kubectl logs -f deployment/nem-api-server -n new-energy-monitoring

# 查看Collector日志
kubectl logs -f deployment/nem-collector -n new-energy-monitoring
```

### 访问服务

```bash
# 端口转发访问API Server
kubectl port-forward svc/nem-api-server 8080:8080 -n new-energy-monitoring

# 访问健康检查端点
curl http://localhost:8080/health/ready
```

## 故障排查

### Pod无法启动

1. 检查镜像是否存在
2. 检查资源是否足够
3. 检查配置是否正确
4. 查看Pod事件

```bash
kubectl describe pod <pod-name> -n new-energy-monitoring
```

### 服务无法访问

1. 检查Service是否正确创建
2. 检查Endpoints是否有Pod
3. 检查Ingress配置
4. 检查网络策略

### 数据库连接失败

1. 检查PostgreSQL是否正常运行
2. 检查Secret配置是否正确
3. 检查网络连接
4. 查看数据库日志

## 生产环境建议

### 高可用配置

1. 至少3个API Server副本
2. 至少3个Collector副本
3. PostgreSQL主从复制
4. Redis哨兵或集群模式
5. Kafka集群模式（至少3个节点）

### 资源规划

| 服务 | CPU请求 | CPU限制 | 内存请求 | 内存限制 |
|------|---------|---------|----------|----------|
| API Server | 1000m | 4000m | 2Gi | 8Gi |
| Collector | 2000m | 4000m | 2Gi | 4Gi |
| Alarm Service | 500m | 2000m | 512Mi | 2Gi |
| Compute Service | 1000m | 4000m | 1Gi | 4Gi |
| AI Service | 2000m | 8000m | 4Gi | 16Gi |
| Scheduler | 500m | 2000m | 512Mi | 2Gi |
| PostgreSQL | 4000m | 8000m | 8Gi | 16Gi |
| Redis | 2000m | 4000m | 4Gi | 8Gi |
| Kafka | 2000m | 4000m | 4Gi | 8Gi |

### 安全配置

1. 启用RBAC
2. 使用Network Policy限制网络访问
3. 启用Pod Security Policy
4. 定期更新镜像和依赖
5. 使用Secret管理敏感信息
6. 启用TLS加密通信

### 监控告警

1. 配置Prometheus监控
2. 设置关键指标告警
3. 配置日志收集
4. 设置性能基线
5. 定期审查告警规则

## 环境差异

### 开发环境 (values-dev.yaml)

- 单副本部署
- 较低的资源限制
- NodePort服务类型
- 禁用TLS
- 调试日志级别

### 生产环境 (values-prod.yaml)

- 多副本高可用部署
- 较高的资源限制
- ClusterIP服务类型 + Ingress
- 启用TLS
- 信息日志级别
- Pod反亲和性配置
- 资源配额和限制

## 维护操作

### 扩缩容

```bash
# 手动扩容API Server
kubectl scale deployment nem-api-server --replicas=5 -n new-energy-monitoring

# 手动缩容
kubectl scale deployment nem-api-server --replicas=2 -n new-energy-monitoring
```

### 配置更新

```bash
# 更新ConfigMap
kubectl create configmap nem-config --from-file=config.yaml --dry-run=client -o yaml | kubectl apply -f - -n new-energy-monitoring

# 重启Pod使配置生效
kubectl rollout restart deployment/nem-api-server -n new-energy-monitoring
```

### 数据备份

```bash
# 备份PostgreSQL
kubectl exec -it nem-postgresql-0 -n new-energy-monitoring -- pg_dump -U nem nem_system > backup.sql

# 备份Redis
kubectl exec -it nem-redis-master-0 -n new-energy-monitoring -- redis-cli SAVE
kubectl cp nem-redis-master-0:/data/dump.rdb ./redis-backup.rdb -n new-energy-monitoring
```

## 技术支持

如有问题，请联系：
- 邮箱: nem@example.com
- 文档: https://github.com/new-energy-monitoring
