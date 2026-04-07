# 新能源监控系统 - 模块设计文档

## 1. 系统架构概述

### 1.1 整体架构

系统采用领域驱动设计(DDD)和微服务架构，分为以下层次：

```
┌─────────────────────────────────────────────────────────────────┐
│                        前端应用层 (Vue 3 + TypeScript)           │
├─────────────────────────────────────────────────────────────────┤
│                        API网关层 (Gin Router)                    │
├─────────────────────────────────────────────────────────────────┤
│                        应用服务层 (Application Services)         │
├─────────────────────────────────────────────────────────────────┤
│                        领域层 (Domain Entities & Services)       │
├─────────────────────────────────────────────────────────────────┤
│                        基础设施层 (Infrastructure)               │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 技术栈

**后端**:
- Go 1.24
- Gin Web Framework
- GORM ORM
- PostgreSQL 18.3
- Redis 8.6.2
- Kafka 4.2.0
- NATS JetStream
- EMQX MQTT Broker

**前端**:
- Vue 3
- TypeScript
- Element Plus
- ECharts
- Vite

---

## 2. 核心模块设计

### 2.1 区域管理模块 (Region Module)

**职责**: 管理电站的地理区域层级结构

**实体**:
```go
type Region struct {
    ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Name        string    `json:"name" gorm:"type:varchar(100);not null"`
    Code        string    `json:"code" gorm:"type:varchar(20);uniqueIndex"`
    ParentID    *string   `json:"parent_id" gorm:"type:varchar(36)"`
    Level       int       `json:"level" gorm:"default:1"`
    Description string    `json:"description" gorm:"type:text"`
    Status      int       `json:"status" gorm:"default:1"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

**服务接口**:
```go
type RegionService interface {
    Create(ctx context.Context, req *CreateRegionRequest) (*Region, error)
    Update(ctx context.Context, id string, req *UpdateRegionRequest) (*Region, error)
    Delete(ctx context.Context, id string) error
    GetByID(ctx context.Context, id string) (*Region, error)
    List(ctx context.Context, query *RegionQuery) ([]*Region, int64, error)
    GetTree(ctx context.Context) ([]*RegionNode, error)
}
```

### 2.2 厂站管理模块 (Station Module)

**职责**: 管理光伏电站、风电场、储能站等

**实体**:
```go
type Station struct {
    ID            string      `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Code          string      `json:"code" gorm:"type:varchar(20);uniqueIndex"`
    Name          string      `json:"name" gorm:"type:varchar(100);not null"`
    Type          StationType `json:"type" gorm:"type:varchar(20);not null"`
    SubRegionID   string      `json:"sub_region_id" gorm:"type:varchar(36)"`
    Capacity      float64     `json:"capacity" gorm:"type:decimal(10,2)"`
    VoltageLevel  string      `json:"voltage_level" gorm:"type:varchar(20)"`
    Longitude     float64     `json:"longitude" gorm:"type:decimal(10,6)"`
    Latitude      float64     `json:"latitude" gorm:"type:decimal(10,6)"`
    Address       string      `json:"address" gorm:"type:varchar(200)"`
    Status        int         `json:"status" gorm:"default:1"`
    CommissionDate *time.Time `json:"commission_date"`
    Description   string      `json:"description" gorm:"type:text"`
    CreatedAt     time.Time   `json:"created_at"`
    UpdatedAt     time.Time   `json:"updated_at"`
}
```

**厂站类型**:
- `solar`: 光伏电站
- `wind`: 风电场
- `storage`: 储能站

### 2.3 设备管理模块 (Device Module)

**职责**: 管理电站内的设备信息

**实体**:
```go
type Device struct {
    ID            string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Code          string    `json:"code" gorm:"type:varchar(20);uniqueIndex"`
    Name          string    `json:"name" gorm:"type:varchar(100);not null"`
    Type          string    `json:"type" gorm:"type:varchar(20);not null"`
    StationID     string    `json:"station_id" gorm:"type:varchar(36);index"`
    Manufacturer  string    `json:"manufacturer" gorm:"type:varchar(50)"`
    Model         string    `json:"model" gorm:"type:varchar(50)"`
    SerialNumber  string    `json:"serial_number" gorm:"type:varchar(50)"`
    RatedPower    float64   `json:"rated_power" gorm:"type:decimal(10,2)"`
    RatedVoltage  float64   `json:"rated_voltage" gorm:"type:decimal(10,2)"`
    RatedCurrent  float64   `json:"rated_current" gorm:"type:decimal(10,2)"`
    Protocol      string    `json:"protocol" gorm:"type:varchar(20)"`
    IPAddress     string    `json:"ip_address" gorm:"type:varchar(15)"`
    Port          int       `json:"port"`
    SlaveID       int       `json:"slave_id"`
    Status        int       `json:"status" gorm:"default:1"`
    LastOnline    *time.Time `json:"last_online"`
    InstallDate   *time.Time `json:"install_date"`
    WarrantyDate  *time.Time `json:"warranty_date"`
    Description   string    `json:"description" gorm:"type:text"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}
```

**设备类型**:
- `inverter`: 逆变器
- `meter`: 电表
- `combiner`: 汇流箱
- `transformer`: 变压器
- `weather_station`: 气象站
- `wind_turbine`: 风机
- `bms`: 电池管理系统

### 2.4 采集点管理模块 (Point Module)

**职责**: 管理设备上的数据采集点

**实体**:
```go
type Point struct {
    ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Code         string    `json:"code" gorm:"type:varchar(20);uniqueIndex"`
    Name         string    `json:"name" gorm:"type:varchar(100);not null"`
    Type         PointType `json:"type" gorm:"type:varchar(20);not null"`
    DeviceID     string    `json:"device_id" gorm:"type:varchar(36);index"`
    DataType     string    `json:"data_type" gorm:"type:varchar(20)"`
    Unit         string    `json:"unit" gorm:"type:varchar(20)"`
    Precision    int       `json:"precision" gorm:"default:2"`
    MinValue     float64   `json:"min_value" gorm:"type:decimal(10,4)"`
    MaxValue     float64   `json:"max_value" gorm:"type:decimal(10,4)"`
    RegisterAddr int       `json:"register_addr"`
    RegisterLen  int       `json:"register_len" gorm:"default:1"`
    Status       int       `json:"status" gorm:"default:1"`
    Description  string    `json:"description" gorm:"type:text"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

**采集点类型**:
- `yx`: 遥信（数字量）
- `yc`: 遥测（模拟量）
- `yk`: 遥控（控制量）
- `yt`: 遥调（参数设置）

### 2.5 告警管理模块 (Alarm Module)

**职责**: 管理告警规则、告警检测、告警通知

**告警规则实体**:
```go
type AlarmRule struct {
    ID            string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Name          string          `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"`
    Type          AlarmRuleType   `json:"type" gorm:"type:varchar(20);not null"`
    Level         AlarmLevel      `json:"level" gorm:"not null"`
    Condition     string          `json:"condition" gorm:"type:text;not null"`
    Threshold     float64         `json:"threshold"`
    Duration      int             `json:"duration" gorm:"default:0"`
    PointID       *string         `json:"point_id" gorm:"type:varchar(36)"`
    DeviceID      *string         `json:"device_id" gorm:"type:varchar(36)"`
    StationID     *string         `json:"station_id" gorm:"type:varchar(36)"`
    NotifyChannels []string       `json:"notify_channels" gorm:"type:text;serializer:json"`
    NotifyUsers    []string       `json:"notify_users" gorm:"type:text;serializer:json"`
    Status        AlarmRuleStatus `json:"status" gorm:"default:1"`
    CreatedAt     time.Time       `json:"created_at"`
    UpdatedAt     time.Time       `json:"updated_at"`
}
```

**告警级别**:
- `1`: 提示
- `2`: 警告
- `3`: 严重
- `4`: 紧急

**告警规则类型**:
- `limit`: 越限告警
- `trend`: 趋势告警
- `quality`: 质量告警
- `communication`: 通信告警

**告警实体**:
```go
type Alarm struct {
    ID          string      `json:"id" gorm:"primaryKey;type:varchar(36)"`
    PointID     string      `json:"point_id" gorm:"type:varchar(36);index"`
    DeviceID    string      `json:"device_id" gorm:"type:varchar(36);index"`
    StationID   string      `json:"station_id" gorm:"type:varchar(36);index"`
    RuleID      string      `json:"rule_id" gorm:"type:varchar(36)"`
    Type        string      `json:"type" gorm:"type:varchar(20)"`
    Level       AlarmLevel  `json:"level" gorm:"not null"`
    Title       string      `json:"title" gorm:"type:varchar(200)"`
    Message     string      `json:"message" gorm:"type:text"`
    Value       float64     `json:"value" gorm:"type:decimal(10,4)"`
    Threshold   float64     `json:"threshold" gorm:"type:decimal(10,4)"`
    Status      AlarmStatus `json:"status" gorm:"default:1"`
    TriggeredAt time.Time   `json:"triggered_at" gorm:"index"`
    AcknowledgedAt *time.Time `json:"acknowledged_at"`
    AcknowledgedBy *string   `json:"acknowledged_by"`
    ClearedAt   *time.Time  `json:"cleared_at"`
    ClearedBy   *string     `json:"cleared_by"`
    CreatedAt   time.Time   `json:"created_at"`
}
```

### 2.6 通知配置模块 (Notification Module)

**职责**: 管理告警通知渠道配置

**实体**:
```go
type NotificationConfig struct {
    ID      string                 `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Type    NotificationType       `json:"type" gorm:"type:varchar(20);uniqueIndex"`
    Name    string                 `json:"name" gorm:"type:varchar(50)"`
    Config  JSONMap                `json:"config" gorm:"type:jsonb"`
    Enabled bool                   `json:"enabled" gorm:"default:false"`
    CreatedAt time.Time            `json:"created_at"`
    UpdatedAt time.Time            `json:"updated_at"`
}
```

**通知类型**:
- `email`: 邮件通知
- `sms`: 短信通知
- `webhook`: Webhook通知
- `wechat`: 企业微信通知

### 2.7 用户管理模块 (User Module)

**职责**: 管理用户认证、授权、权限

**用户实体**:
```go
type User struct {
    ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Username  string    `json:"username" gorm:"type:varchar(50);uniqueIndex"`
    Password  string    `json:"-" gorm:"type:varchar(100)"`
    Nickname  string    `json:"nickname" gorm:"type:varchar(50)"`
    Email     string    `json:"email" gorm:"type:varchar(100)"`
    Phone     string    `json:"phone" gorm:"type:varchar(20)"`
    Avatar    string    `json:"avatar" gorm:"type:varchar(200)"`
    Role      string    `json:"role" gorm:"type:varchar(20)"`
    Status    int       `json:"status" gorm:"default:1"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**角色实体**:
```go
type Role struct {
    ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Name        string    `json:"name" gorm:"type:varchar(50);uniqueIndex"`
    Code        string    `json:"code" gorm:"type:varchar(20);uniqueIndex"`
    Description string    `json:"description" gorm:"type:text"`
    Permissions []string  `json:"permissions" gorm:"type:text;serializer:json"`
    Status      int       `json:"status" gorm:"default:1"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

**权限实体**:
```go
type Permission struct {
    ID       string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Name     string    `json:"name" gorm:"type:varchar(50);not null"`
    Code     string    `json:"code" gorm:"type:varchar(50);uniqueIndex"`
    Type     string    `json:"type" gorm:"type:varchar(20)"`
    ParentID *string   `json:"parent_id" gorm:"type:varchar(36)"`
    Path     string    `json:"path" gorm:"type:varchar(100)"`
    Method   string    `json:"method" gorm:"type:varchar(10)"`
    Status   int       `json:"status" gorm:"default:1"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### 2.8 操作日志模块 (OperationLog Module)

**职责**: 记录系统操作日志

**实体**:
```go
type OperationLog struct {
    ID         string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
    UserID     string    `json:"user_id" gorm:"type:varchar(36);index"`
    Username   string    `json:"username" gorm:"type:varchar(50)"`
    Method     string    `json:"method" gorm:"type:varchar(10)"`
    Path       string    `json:"path" gorm:"type:varchar(200);index"`
    Action     string    `json:"action" gorm:"type:varchar(50)"`
    Resource   string    `json:"resource" gorm:"type:varchar(50);index"`
    ResourceID string    `json:"resource_id" gorm:"type:varchar(36)"`
    RequestIP  string    `json:"request_ip" gorm:"type:varchar(50)"`
    UserAgent  string    `json:"user_agent" gorm:"type:varchar(200)"`
    Request    string    `json:"request" gorm:"type:text"`
    Response   string    `json:"response" gorm:"type:text"`
    Status     int       `json:"status"`
    Duration   int64     `json:"duration"`
    Error      string    `json:"error" gorm:"type:text"`
    CreatedAt  time.Time `json:"created_at" gorm:"index"`
}
```

---

## 3. 数据流设计

### 3.1 数据采集流程

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   设备层    │───▶│   采集器    │───▶│   消息队列  │───▶│   处理器    │
│  (Device)   │    │ (Collector) │    │   (Kafka)   │    │ (Processor) │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                                                                │
                                                                ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   前端展示  │◀───│   API服务   │◀───│   缓存层    │◀───│   存储层    │
│   (Web)     │    │  (API)      │    │   (Redis)   │    │ (PostgreSQL)│
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

### 3.2 告警处理流程

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   数据检测  │───▶│   告警引擎  │───▶│   去重处理  │───▶│   告警存储  │
│ (Detector)  │    │   (Engine)  │    │   (Dedup)   │    │  (Storage)  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                                                                │
                                                                ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   前端展示  │◀───│  WebSocket  │◀───│   聚合器    │◀───│   通知器    │
│   (Web)     │    │   (WS)      │    │ (Aggregator)│    │ (Notifier)  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

---

## 4. 安全设计

### 4.1 认证机制

- JWT Token认证
- Token有效期: 2小时
- 刷新Token机制

### 4.2 权限控制

- 基于角色的访问控制(RBAC)
- 菜单权限
- 操作权限
- 数据权限

### 4.3 数据安全

- 密码加密存储(bcrypt)
- 敏感数据脱敏
- SQL注入防护
- XSS防护
- CSRF防护

---

## 5. 性能优化

### 5.1 缓存策略

- Redis缓存热点数据
- 本地缓存配置数据
- 缓存预热
- 缓存穿透防护

### 5.2 数据库优化

- 索引优化
- 分页查询
- 连接池管理
- 慢查询监控

### 5.3 前端优化

- 路由懒加载
- 组件按需加载
- 图片懒加载
- 虚拟滚动

---

## 6. 部署架构

### 6.1 容器化部署

```yaml
services:
  postgres:
    image: postgres:18.3-alpine
  redis:
    image: redis:8.6.2-alpine
  kafka:
    image: apache/kafka-native:4.2.0
  nacos:
    image: qingpan/rnacos:v0.8.2-alpine
  emqx:
    image: emqx/emqx:6.2.0
  backend:
    build: .
    ports:
      - "8080:8080"
  frontend:
    build: ./web
    ports:
      - "80:80"
```

### 6.2 Kubernetes部署

- Deployment: 无状态应用部署
- StatefulSet: 有状态应用部署
- Service: 服务发现
- Ingress: 入口路由
- ConfigMap: 配置管理
- Secret: 密钥管理

---

## 7. 监控告警

### 7.1 应用监控

- Prometheus指标采集
- Grafana可视化
- 健康检查接口

### 7.2 日志管理

- 结构化日志
- 日志级别管理
- 日志聚合分析

---

## 8. 扩展性设计

### 8.1 水平扩展

- 无状态服务设计
- 负载均衡
- 数据分片

### 8.2 插件化设计

- 协议插件
- 告警规则插件
- 通知渠道插件

---

## 版本历史

| 版本 | 日期 | 说明 |
|-----|------|------|
| 1.0.0 | 2024-04-05 | 初始版本 |
