# Nacos集成设计文档

## 1. 概述

### 1.1 文档目的

本文档详细描述新能源在线监控系统中Nacos组件的集成设计方案，包括服务注册与发现、配置中心、高可用设计等内容。

### 1.2 适用范围

本设计适用于新能源在线监控系统的微服务架构，涵盖以下服务：
- api-server（API服务）
- collector（采集服务）
- alarm（告警服务）
- compute（计算服务）
- ai-service（AI服务）
- scheduler（调度服务）

### 1.3 参考资料

- Nacos官方文档：https://nacos.io/zh-cn/docs/what-is-nacos.html
- Nacos SDK Go：https://github.com/nacos-group/nacos-sdk-go

## 2. 架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        客户端/浏览器                              │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      负载均衡器 (Nginx/SLB)                       │
└─────────────────────────────────────────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│   API Server  │       │   Collector   │       │    Alarm      │
│   (Instance)  │       │  (Instance)   │       │  (Instance)   │
└───────────────┘       └───────────────┘       └───────────────┘
        │                       │                       │
        └───────────────────────┼───────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Nacos Server Cluster                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │   Nacos 1   │◄──►│   Nacos 2   │◄──►│   Nacos 3   │         │
│  │  (Leader)   │    │ (Follower)  │    │ (Follower)  │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
│         │                  │                  │                 │
│         └──────────────────┼──────────────────┘                 │
│                            ▼                                    │
│                   ┌─────────────────┐                          │
│                   │  MySQL Cluster  │                          │
│                   │  (数据持久化)    │                          │
│                   └─────────────────┘                          │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 组件职责

| 组件 | 职责 |
|------|------|
| Nacos Server | 服务注册中心、配置中心 |
| Nacos Client (SDK) | 服务注册、服务发现、配置获取、配置监听 |
| MySQL | Nacos数据持久化存储 |
| Nginx/SLB | 负载均衡，提供统一入口 |

## 3. 服务注册与发现

### 3.1 服务注册流程

```
┌─────────┐     1.启动服务      ┌─────────────┐
│  服务   │ ──────────────────► │  Nacos SDK  │
│ 实例   │                     └─────────────┘
└─────────┘                           │
                                      │ 2.发送注册请求
                                      ▼
                               ┌─────────────┐
                               │    Nacos    │
                               │   Server    │
                               └─────────────┘
                                      │
                                      │ 3.存储服务信息
                                      ▼
                               ┌─────────────┐
                               │   MySQL     │
                               │  (持久化)    │
                               └─────────────┘
```

### 3.2 服务注册配置

```yaml
nacos:
  enabled: true
  # 服务器地址列表
  server_configs:
    - ip_addr: "nacos1.example.com"
      port: 8848
      context_path: "/nacos"
      scheme: "http"
    - ip_addr: "nacos2.example.com"
      port: 8848
      context_path: "/nacos"
      scheme: "http"
    - ip_addr: "nacos3.example.com"
      port: 8848
      context_path: "/nacos"
      scheme: "http"
  
  # 客户端配置
  client_config:
    namespace_id: "new-energy-monitoring"
    timeout_ms: 5000
    not_load_cache_at_start: true
    update_cache_when_empty: true
    username: "nacos"
    password: "nacos"
    log_level: "info"
  
  # 服务注册配置
  service_name: "api-server"
  group: "DEFAULT_GROUP"
  cluster_name: "DEFAULT"
  weight: 1.0
  metadata:
    version: "1.0.0"
    env: "prod"
  ephemeral: true  # 临时实例，心跳检测
```

### 3.3 服务发现机制

#### 3.3.1 服务发现流程

```
┌─────────┐     1.请求服务      ┌─────────────┐
│ 消费者  │ ──────────────────► │  Nacos SDK  │
│ 服务   │                     └─────────────┘
└─────────┘                           │
      ▲                               │ 2.查询服务列表
      │                               ▼
      │                        ┌─────────────┐
      │                        │    Nacos    │
      │                        │   Server    │
      │                        └─────────────┘
      │                               │
      │ 4.返回服务实例                │ 3.返回实例列表
      │                               ▼
      └─────────────────────── ┌─────────────┐
                              │  服务实例    │
                              │  (缓存)      │
                              └─────────────┘
```

#### 3.3.2 负载均衡策略

```go
type LoadBalanceStrategy string

const (
    // WeightedRoundRobin 加权轮询（默认）
    WeightedRoundRobin LoadBalanceStrategy = "weighted_round_robin"
    // Random 随机
    Random LoadBalanceStrategy = "random"
    // ConsistentHash 一致性哈希
    ConsistentHash LoadBalanceStrategy = "consistent_hash"
)

type LoadBalancer interface {
    Select(instances []*ServiceInstance) (*ServiceInstance, error)
}
```

### 3.4 健康检查机制

#### 3.4.1 客户端心跳

```go
type HealthChecker struct {
    interval    time.Duration  // 心跳间隔，默认5秒
    timeout     time.Duration  // 超时时间，默认15秒
    maxRetry    int            // 最大重试次数，默认3次
}

func (h *HealthChecker) Start() {
    ticker := time.NewTicker(h.interval)
    for {
        select {
        case <-ticker.C:
            h.sendHeartbeat()
        case <-h.stopCh:
            return
        }
    }
}
```

#### 3.4.2 服务端健康检查

| 检查类型 | 描述 | 适用场景 |
|----------|------|----------|
| 临时实例 | 客户端心跳，超时自动剔除 | 微服务场景 |
| 持久实例 | 服务端主动探测 | 数据库、中间件等 |

### 3.5 服务订阅机制

```go
type ServiceSubscriber struct {
    serviceName string
    callback    func(instances []*ServiceInstance)
}

func (s *ServiceSubscriber) Subscribe() error {
    return s.registry.Subscribe(s.serviceName, func(instances []*ServiceInstance) {
        // 服务实例变更回调
        s.callback(instances)
        // 更新本地缓存
        s.updateLocalCache(instances)
        // 触发负载均衡器刷新
        s.loadBalancer.Refresh(instances)
    })
}
```

## 4. 配置中心

### 4.1 配置管理架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        配置管理控制台                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │  配置创建   │  │  配置编辑   │  │  配置发布   │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Nacos Config Server                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │  版本管理   │  │  灰度发布   │  │  变更审计   │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        微服务集群                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │   服务 A    │  │   服务 B    │  │   服务 C    │            │
│  │  (配置监听) │  │  (配置监听) │  │  (配置监听) │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 配置命名空间设计

```
命名空间(Namespace)
├── dev（开发环境）
│   ├── api-server.yaml
│   ├── collector.yaml
│   ├── alarm.yaml
│   └── shared-config.yaml
├── test（测试环境）
│   ├── api-server.yaml
│   ├── collector.yaml
│   ├── alarm.yaml
│   └── shared-config.yaml
├── prod（生产环境）
│   ├── api-server.yaml
│   ├── collector.yaml
│   ├── alarm.yaml
│   └── shared-config.yaml
└── standalone（单机模式）
    └── standalone.yaml
```

### 4.3 配置加载优先级

```
优先级（从高到低）：
1. 命令行参数 (--config=/path/to/config.yaml)
2. 环境变量 (NEM_DATABASE_HOST=x.x.x.x)
3. Nacos配置中心
4. 本地环境配置文件 (config-{env}.yaml)
5. 本地基础配置文件 (config.yaml)
6. 默认值
```

### 4.4 配置加载流程

```go
func (l *Loader) Load(ctx context.Context) (*Config, error) {
    // 1. 设置默认值
    l.setDefaults()
    
    // 2. 加载本地配置文件
    if err := l.loadLocalConfig(); err != nil && !l.fallback {
        return nil, err
    }
    
    // 3. 加载环境变量
    l.loadEnvVars()
    
    // 4. 加载命令行参数
    l.loadCommandLine()
    
    // 5. 从配置中心加载（覆盖本地配置）
    if l.configCenter != "" {
        if err := l.loadFromConfigCenter(ctx); err != nil {
            if !l.fallback {
                return nil, err
            }
            // 使用本地配置兜底
        }
    }
    
    // 6. 解析配置
    var cfg Config
    if err := l.viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    
    // 7. 启动配置监听
    if l.watchEnabled {
        l.startWatcher()
    }
    
    return &cfg, nil
}
```

### 4.5 配置动态更新

#### 4.5.1 配置监听机制

```go
type ConfigWatcher struct {
    client    ConfigClient
    callbacks map[string][]func(string, interface{})
}

func (w *ConfigWatcher) Watch(dataId, group string) error {
    return w.client.ListenConfig(vo.ListenConfigParam{
        DataId: dataId,
        Group:  group,
        OnChange: func(namespace, group, dataId, data string) {
            // 解析新配置
            newConfig := w.parseConfig(data)
            
            // 触发回调
            for key, callbacks := range w.callbacks {
                if newValue, ok := newConfig[key]; ok {
                    for _, cb := range callbacks {
                        cb(key, newValue)
                    }
                }
            }
        },
    })
}
```

#### 4.5.2 配置热更新示例

```go
// 数据库连接池配置热更新
loader.Watch("database.max_open_conns", func(key string, value interface{}) {
    if maxConns, ok := value.(int); ok {
        db.SetMaxOpenConns(maxConns)
        logger.Info("数据库连接池配置已更新", zap.Int("max_open_conns", maxConns))
    }
})

// Redis连接池配置热更新
loader.Watch("redis.pool_size", func(key string, value interface{}) {
    if poolSize, ok := value.(int); ok {
        redisClient.SetPoolSize(poolSize)
        logger.Info("Redis连接池配置已更新", zap.Int("pool_size", poolSize))
    }
})
```

### 4.6 灰度发布

#### 4.6.1 灰度发布流程

```
┌─────────────┐     创建灰度配置     ┌─────────────┐
│   运维人员   │ ─────────────────► │    Nacos    │
└─────────────┘                     └─────────────┘
                                          │
                                          │ 配置灰度规则
                                          ▼
                                   ┌─────────────┐
                                   │  灰度实例   │
                                   │  (Beta)     │
                                   └─────────────┘
                                          │
                                          │ 验证通过
                                          ▼
                                   ┌─────────────┐
                                   │  全量发布   │
                                   └─────────────┘
```

#### 4.6.2 灰度规则配置

```go
type GrayRule struct {
    Type    GrayRuleType `json:"type"`
    Key     string       `json:"key"`
    Values  []string     `json:"values"`
    Percent int          `json:"percent"`  // 灰度比例
}

type GrayRuleType string

const (
    GrayRuleByIP      GrayRuleType = "IP"       // 按IP灰度
    GrayRuleByTag     GrayRuleType = "TAG"      // 按标签灰度
    GrayRuleByPercent GrayRuleType = "PERCENT"  // 按比例灰度
)
```

## 5. 高可用设计

### 5.1 Nacos集群部署

#### 5.1.1 集群拓扑

```
                    ┌─────────────────┐
                    │   负载均衡器    │
                    │  (VIP: 10.0.0.100) │
                    └─────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│   Nacos 1     │   │   Nacos 2     │   │   Nacos 3     │
│  10.0.0.101   │   │  10.0.0.102   │   │  10.0.0.103   │
│   (Leader)    │◄─►│  (Follower)   │◄─►│  (Follower)   │
└───────────────┘   └───────────────┘   └───────────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
                    ┌───────┴───────┐
                    │               │
                    ▼               ▼
            ┌───────────────┐ ┌───────────────┐
            │   MySQL 1     │ │   MySQL 2     │
            │   (Master)    │ │   (Slave)     │
            └───────────────┘ └───────────────┘
```

#### 5.1.2 集群配置

```properties
# nacos/conf/cluster.conf
10.0.0.101:8848
10.0.0.102:8848
10.0.0.103:8848

# nacos/conf/application.properties
# 数据库配置
spring.datasource.platform=mysql
db.num=2
db.url.0=jdbc:mysql://10.0.0.201:3306/nacos?characterEncoding=utf8&connectTimeout=1000&socketTimeout=3000&autoReconnect=true&useUnicode=true&useSSL=false&serverTimezone=Asia/Shanghai
db.url.1=jdbc:mysql://10.0.0.202:3306/nacos?characterEncoding=utf8&connectTimeout=1000&socketTimeout=3000&autoReconnect=true&useUnicode=true&useSSL=false&serverTimezone=Asia/Shanghai
db.user.0=nacos
db.password.0=nacos_password
db.user.1=nacos
db.password.1=nacos_password

# 集群配置
nacos.member.list=10.0.0.101:8848,10.0.0.102:8848,10.0.0.103:8848
```

### 5.2 故障转移机制

#### 5.2.1 客户端故障转移

```go
type FailoverConfig struct {
    // 重试次数
    RetryCount int `yaml:"retry_count" default:"3"`
    // 重试间隔
    RetryInterval time.Duration `yaml:"retry_interval" default:"1s"`
    // 故障转移超时
    FailoverTimeout time.Duration `yaml:"failover_timeout" default:"5s"`
    // 本地缓存开关
    LocalCacheEnabled bool `yaml:"local_cache_enabled" default:"true"`
    // 本地缓存过期时间
    LocalCacheExpire time.Duration `yaml:"local_cache_expire" default:"24h"`
}

func (r *Registry) RegisterWithFailover(instance *ServiceInstance) error {
    var lastErr error
    
    for i := 0; i < r.failoverConfig.RetryCount; i++ {
        err := r.Register(instance)
        if err == nil {
            return nil
        }
        
        lastErr = err
        
        // 切换到下一个服务器
        r.switchServer()
        
        // 等待重试间隔
        time.Sleep(r.failoverConfig.RetryInterval)
    }
    
    // 所有服务器都失败，使用本地缓存
    if r.failoverConfig.LocalCacheEnabled {
        r.saveToLocalCache(instance)
    }
    
    return fmt.Errorf("register failed after %d retries: %w", r.failoverConfig.RetryCount, lastErr)
}
```

#### 5.2.2 本地缓存策略

```go
type LocalCache struct {
    cacheDir string
    expire   time.Duration
}

func (c *LocalCache) Save(serviceName string, instances []*ServiceInstance) error {
    data := CacheData{
        Instances:  instances,
        UpdateTime: time.Now(),
        ExpireTime: time.Now().Add(c.expire),
    }
    
    filePath := filepath.Join(c.cacheDir, serviceName+".cache")
    return os.WriteFile(filePath, data.Marshal(), 0644)
}

func (c *LocalCache) Load(serviceName string) ([]*ServiceInstance, error) {
    filePath := filepath.Join(c.cacheDir, serviceName+".cache")
    data, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }
    
    var cacheData CacheData
    if err := cacheData.Unmarshal(data); err != nil {
        return nil, err
    }
    
    // 检查是否过期
    if time.Now().After(cacheData.ExpireTime) {
        return nil, fmt.Errorf("cache expired")
    }
    
    return cacheData.Instances, nil
}
```

### 5.3 容灾演练

#### 5.3.1 演练场景

| 场景 | 演练内容 | 预期结果 |
|------|----------|----------|
| 单节点故障 | 停止一个Nacos节点 | 服务正常，客户端自动切换 |
| 多节点故障 | 停止两个Nacos节点 | 服务正常，剩余节点接管 |
| 网络分区 | 模拟网络分区 | 本地缓存生效，服务降级运行 |
| 数据库故障 | 停止MySQL | Nacos使用内存数据，新注册暂存 |
| 全局故障 | 所有Nacos节点不可用 | 本地缓存生效，服务降级运行 |

## 6. 监控与告警

### 6.1 监控指标

#### 6.1.1 Nacos服务端指标

| 指标名称 | 说明 | 告警阈值 |
|----------|------|----------|
| nacos_service_count | 服务数量 | - |
| nacos_instance_count | 实例数量 | - |
| nacos_config_count | 配置数量 | - |
| nacos_http_request_total | HTTP请求总数 | - |
| nacos_http_request_error | HTTP请求错误数 | > 1% |
| nacos_cpu_usage | CPU使用率 | > 80% |
| nacos_memory_usage | 内存使用率 | > 80% |
| nacos_gc_pause_seconds | GC暂停时间 | > 100ms |

#### 6.1.2 客户端指标

```go
type NacosMetrics struct {
    // 注册指标
    RegisterTotal     prometheus.Counter
    RegisterSuccess   prometheus.Counter
    RegisterFailures  prometheus.Counter
    
    // 发现指标
    DiscoverTotal     prometheus.Counter
    DiscoverSuccess   prometheus.Counter
    DiscoverFailures  prometheus.Counter
    DiscoverLatency   prometheus.Histogram
    
    // 配置指标
    ConfigGetTotal    prometheus.Counter
    ConfigGetSuccess  prometheus.Counter
    ConfigGetFailures prometheus.Counter
    ConfigUpdateTotal prometheus.Counter
    
    // 连接指标
    ConnectionStatus  prometheus.Gauge
    ReconnectTotal    prometheus.Counter
}
```

### 6.2 告警规则

```yaml
groups:
  - name: nacos_alerts
    rules:
      # Nacos服务不可用
      - alert: NacosServerDown
        expr: up{job="nacos"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Nacos服务器宕机"
          description: "Nacos服务器 {{ $labels.instance }} 已经宕机超过1分钟"
      
      # 服务注册失败率过高
      - alert: NacosRegisterFailureHigh
        expr: rate(nacos_register_failures_total[5m]) / rate(nacos_register_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "服务注册失败率过高"
          description: "服务注册失败率超过10%"
      
      # 配置获取失败率过高
      - alert: NacosConfigFailureHigh
        expr: rate(nacos_config_get_failures_total[5m]) / rate(nacos_config_get_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "配置获取失败率过高"
          description: "配置获取失败率超过10%"
      
      # Nacos内存使用率过高
      - alert: NacosMemoryHigh
        expr: nacos_memory_usage > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Nacos内存使用率过高"
          description: "Nacos内存使用率超过80%"
```

### 6.3 监控仪表盘

#### 6.3.1 Grafana仪表盘配置

```json
{
  "dashboard": {
    "title": "Nacos监控仪表盘",
    "panels": [
      {
        "title": "服务注册数量",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(nacos_service_count)"
          }
        ]
      },
      {
        "title": "实例数量",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(nacos_instance_count)"
          }
        ]
      },
      {
        "title": "请求成功率",
        "type": "gauge",
        "targets": [
          {
            "expr": "rate(nacos_http_request_total[5m]) - rate(nacos_http_request_error[5m]) / rate(nacos_http_request_total[5m])"
          }
        ]
      },
      {
        "title": "CPU使用率",
        "type": "graph",
        "targets": [
          {
            "expr": "nacos_cpu_usage"
          }
        ]
      },
      {
        "title": "内存使用率",
        "type": "graph",
        "targets": [
          {
            "expr": "nacos_memory_usage"
          }
        ]
      }
    ]
  }
}
```

## 7. 安全设计

### 7.1 认证授权

#### 7.1.1 开启认证

```properties
# nacos/conf/application.properties
nacos.core.auth.enabled=true
nacos.core.auth.server.identity.key=serverIdentity
nacos.core.auth.server.identity.value=security
nacos.core.auth.plugin.nacos.token.secret.key=SecretKey012345678901234567890123456789012345678901234567890123456789
```

#### 7.1.2 客户端认证

```go
clientConfig := constant.NewClientConfig(
    constant.WithNamespaceId("new-energy-monitoring"),
    constant.WithUsername("app_user"),
    constant.WithPassword("secure_password"),
)
```

### 7.2 网络安全

#### 7.2.1 网络隔离

```
┌─────────────────────────────────────────────────────────────────┐
│                        外部网络                                 │
└─────────────────────────────────────────────────────────────────┘
                                │
                                │ HTTPS
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      DMZ区域                                    │
│  ┌─────────────┐                                               │
│  │   Nginx     │                                               │
│  │  (反向代理)  │                                               │
│  └─────────────┘                                               │
└─────────────────────────────────────────────────────────────────┘
                                │
                                │ 内网
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      应用区域                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │  微服务集群  │  │  Nacos集群  │  │  MySQL集群  │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────┘
```

#### 7.2.2 访问控制

```yaml
# 网络策略配置
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: nacos-network-policy
spec:
  podSelector:
    matchLabels:
      app: nacos
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              access: nacos-client
      ports:
        - protocol: TCP
          port: 8848
        - protocol: TCP
          port: 9848
        - protocol: TCP
          port: 9849
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: mysql
      ports:
        - protocol: TCP
          port: 3306
```

### 7.3 数据安全

#### 7.3.1 敏感配置加密

```go
type EncryptedConfig struct {
    Key     string `json:"key"`
    Value   string `json:"value"`
    Encrypted bool  `json:"encrypted"`
}

func (c *ConfigClient) GetEncryptedConfig(dataId, group string) (string, error) {
    config, err := c.GetConfig(dataId, group)
    if err != nil {
        return "", err
    }
    
    // 解密配置
    if isEncrypted(config) {
        return decrypt(config, c.encryptionKey)
    }
    
    return config, nil
}
```

## 8. 运维指南

### 8.1 部署步骤

#### 8.1.1 准备工作

```bash
# 1. 创建数据库
mysql -u root -p -e "CREATE DATABASE nacos CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 2. 导入初始化脚本
mysql -u root -p nacos < nacos-mysql.sql

# 3. 创建应用用户
mysql -u root -p -e "CREATE USER 'nacos'@'%' IDENTIFIED BY 'secure_password';"
mysql -u root -p -e "GRANT ALL PRIVILEGES ON nacos.* TO 'nacos'@'%';"
mysql -u root -p -e "FLUSH PRIVILEGES;"
```

#### 8.1.2 Kubernetes部署

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: nacos
  namespace: middleware
spec:
  serviceName: nacos-headless
  replicas: 3
  selector:
    matchLabels:
      app: nacos
  template:
    metadata:
      labels:
        app: nacos
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            - labelSelector:
                matchLabels:
                  app: nacos
              topologyKey: kubernetes.io/hostname
      containers:
        - name: nacos
          image: nacos/nacos-server:v2.2.0
          ports:
            - containerPort: 8848
              name: http
            - containerPort: 9848
              name: client-rpc
            - containerPort: 9849
              name: raft-rpc
          env:
            - name: MODE
              value: "cluster"
            - name: SPRING_DATASOURCE_PLATFORM
              value: "mysql"
            - name: MYSQL_SERVICE_HOST
              value: "mysql-service"
            - name: MYSQL_SERVICE_PORT
              value: "3306"
            - name: MYSQL_SERVICE_DB_NAME
              value: "nacos"
            - name: MYSQL_SERVICE_USER
              value: "nacos"
            - name: MYSQL_SERVICE_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: nacos-secret
                  key: mysql-password
            - name: NACOS_SERVERS
              value: "nacos-0.nacos-headless:8848 nacos-1.nacos-headless:8848 nacos-2.nacos-headless:8848"
          resources:
            requests:
              cpu: "500m"
              memory: "1Gi"
            limits:
              cpu: "2000m"
              memory: "4Gi"
          livenessProbe:
            httpGet:
              path: /nacos/v1/console/health/liveness
              port: 8848
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /nacos/v1/console/health/readiness
              port: 8848
            initialDelaySeconds: 30
            periodSeconds: 10
```

### 8.2 日常运维

#### 8.2.1 健康检查

```bash
# 检查Nacos健康状态
curl -s http://nacos:8848/nacos/v1/console/health/readiness

# 检查服务列表
curl -s http://nacos:8848/nacos/v1/ns/service/list?pageNo=1&pageSize=100

# 检查配置列表
curl -s http://nacos:8848/nacos/v1/cs/configs?dataId=&group=&pageNo=1&pageSize=100
```

#### 8.2.2 日志管理

```yaml
# 日志轮转配置
/var/log/nacos/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0644 nacos nacos
    sharedscripts
    postrotate
        docker kill -s USR1 nacos 2>/dev/null || true
    endscript
}
```

### 8.3 故障排查

#### 8.3.1 常见问题

| 问题 | 可能原因 | 解决方案 |
|------|----------|----------|
| 服务注册失败 | 网络不通、认证失败 | 检查网络、验证用户名密码 |
| 配置获取失败 | 配置不存在、权限不足 | 检查配置、验证权限 |
| 集群选主失败 | 网络分区、节点故障 | 检查网络、重启故障节点 |
| 内存溢出 | 配置过多、内存不足 | 增加内存、清理无用配置 |

#### 8.3.2 排查命令

```bash
# 查看Nacos日志
kubectl logs -f nacos-0 -n middleware

# 查看Nacos状态
kubectl exec -it nacos-0 -n middleware -- curl -s localhost:8848/nacos/v1/console/health/readiness

# 查看集群状态
kubectl exec -it nacos-0 -n middleware -- curl -s localhost:8848/nacos/v1/core/cluster/nodes

# 查看Raft状态
kubectl exec -it nacos-0 -n middleware -- curl -s localhost:8848/nacos/v1/core/raft/peers
```

## 9. 附录

### 9.1 配置模板

#### 9.1.1 服务配置模板

```yaml
server:
  name: api-server
  port: 8080
  mode: release

database:
  type: postgres
  host: ${DATABASE_HOST:localhost}
  port: ${DATABASE_PORT:5432}
  user: ${DATABASE_USER:postgres}
  password: ${DATABASE_PASSWORD:}
  dbname: ${DATABASE_NAME:new_energy}
  sslmode: disable
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_lifetime: 300s

redis:
  addrs:
    - ${REDIS_HOST:localhost}:${REDIS_PORT:6379}
  password: ${REDIS_PASSWORD:}
  db: 0
  pool_size: 100

kafka:
  brokers:
    - ${KAFKA_BROKER:localhost:9092}
  topic_prefix: nem

nacos:
  enabled: true
  server_configs:
    - ip_addr: ${NACOS_HOST:localhost}
      port: ${NACOS_PORT:8848}
  client_config:
    namespace_id: ${NACOS_NAMESPACE:dev}
    username: ${NACOS_USERNAME:nacos}
    password: ${NACOS_PASSWORD:nacos}
  service_name: api-server
  group: DEFAULT_GROUP
```

### 9.2 API参考

#### 9.2.1 服务注册API

```go
// 注册服务实例
func (r *Registry) Register(instance *ServiceInstance) error

// 注销服务实例
func (r *Registry) Deregister(serviceName string) error

// 发现服务实例
func (r *Registry) Discover(serviceName string) ([]*ServiceInstance, error)

// 订阅服务变更
func (r *Registry) Subscribe(serviceName string, callback func([]*ServiceInstance)) error
```

#### 9.2.2 配置中心API

```go
// 获取配置
func (c *ConfigClient) GetConfig(dataId, group string) (string, error)

// 发布配置
func (c *ConfigClient) PublishConfig(dataId, group, content string) error

// 删除配置
func (c *ConfigClient) DeleteConfig(dataId, group string) error

// 监听配置
func (c *ConfigClient) ListenConfig(param ListenConfigParam) error
```

### 9.3 版本历史

| 版本 | 日期 | 作者 | 说明 |
|------|------|------|------|
| 1.0 | 2024-01-15 | 系统架构组 | 初始版本 |
| 1.1 | 2024-02-01 | 系统架构组 | 增加灰度发布设计 |
| 1.2 | 2024-03-01 | 系统架构组 | 增加监控告警设计 |
