# 配置中心模块设计文档

## 1. 模块概述

### 1.1 功能定位

配置中心是新能源在线监控系统的核心配置管理模块，支持配置项的动态下发、版本控制、灰度发布及配置变更审计，实现多环境、多集群的统一配置管理。

### 1.2 核心功能

- 集中式配置管理
- 配置动态下发
- 版本控制与回滚
- 灰度发布
- 配置变更审计
- 多环境支持

---

## 2. 架构设计

### 2.1 配置加载优先级

```
优先级从高到低：

1. 命令行参数 (--config=/path/to/config.yaml)
2. 环境变量 (NEM_CONFIG_*)
3. 配置中心 (Nacos/Apollo/Consul)
4. 本地配置文件 (按环境区分)
   - config-{env}.yaml
   - config.yaml
5. 默认值

冲突解决机制：高优先级配置覆盖低优先级配置
```

### 2.2 配置加载流程

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              配置加载流程                                    │
│                                                                              │
│  ┌─────────────┐                                                            │
│  │  应用启动   │                                                            │
│  └──────┬──────┘                                                            │
│         │                                                                    │
│         ▼                                                                    │
│  ┌─────────────┐     ┌─────────────┐                                        │
│  │ 检查命令行  │ ──► │ 加载指定配置 │                                        │
│  │   参数      │     │   文件      │                                        │
│  └──────┬──────┘     └─────────────┘                                        │
│         │ 无参数                                                              │
│         ▼                                                                    │
│  ┌─────────────┐     ┌─────────────┐                                        │
│  │ 检查环境变量│ ──► │ NEM_ENV     │                                        │
│  │ NEM_ENV     │     │ 确定环境    │                                        │
│  └──────┬──────┘     └─────────────┘                                        │
│         │                                                                    │
│         ▼                                                                    │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                    │
│  │ 尝试连接    │ ──► │ 连接成功    │ ──► │ 加载远程配置│                    │
│  │ 配置中心    │     │             │     │             │                    │
│  └──────┬──────┘     └─────────────┘     └──────┬──────┘                    │
│         │ 连接失败                          │                                │
│         ▼                                    │                                │
│  ┌─────────────┐                            │                                │
│  │ 加载本地配置│ ◄──────────────────────────┘                                │
│  │ config-{env}│                                                            │
│  └──────┬──────┘                                                            │
│         │                                                                    │
│         ▼                                                                    │
│  ┌─────────────┐                                                            │
│  │ 合并配置    │                                                            │
│  │ 启动应用    │                                                            │
│  └─────────────┘                                                            │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.3 配置中心架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              配置中心架构                                    │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         配置管理控制台                                │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐                   │   │
│  │  │配置编辑 │ │版本管理 │ │灰度发布 │ │审计日志 │                   │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘                   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                      │                                       │
│                                      ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         配置存储层                                    │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                  │   │
│  │  │ PostgreSQL  │  │    Redis    │  │  对象存储   │                  │   │
│  │  │ (配置元数据) │  │ (配置缓存)  │  │ (配置备份)  │                  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘                  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                      │                                       │
│                                      ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         配置推送层                                    │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                  │   │
│  │  │ WebSocket   │  │   Kafka    │  │  HTTP轮询   │                  │   │
│  │  │ (实时推送)  │  │ (批量通知) │  │ (兜底方案)  │                  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘                  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. 数据模型

### 3.1 配置项表 (config_items)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | UUID主键 |
| key | VARCHAR(200) | UNIQUE, NOT NULL | 配置键 |
| value | TEXT | | 配置值 |
| value_type | VARCHAR(20) | DEFAULT 'string' | 值类型 |
| env | VARCHAR(20) | NOT NULL | 环境 |
| namespace | VARCHAR(100) | DEFAULT 'default' | 命名空间 |
| group | VARCHAR(100) | DEFAULT 'default' | 配置分组 |
| description | TEXT | | 描述 |
| encrypted | BOOLEAN | DEFAULT FALSE | 是否加密 |
| enabled | BOOLEAN | DEFAULT TRUE | 是否启用 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW() | 更新时间 |

### 3.2 配置版本表 (config_versions)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | UUID主键 |
| config_id | VARCHAR(36) | FOREIGN KEY | 配置项ID |
| version | INTEGER | NOT NULL | 版本号 |
| value | TEXT | | 配置值 |
| change_reason | TEXT | | 变更原因 |
| changed_by | VARCHAR(100) | | 变更人 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |

### 3.3 配置发布表 (config_releases)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | UUID主键 |
| config_id | VARCHAR(36) | FOREIGN KEY | 配置项ID |
| version | INTEGER | NOT NULL | 发布版本 |
| env | VARCHAR(20) | NOT NULL | 目标环境 |
| release_type | VARCHAR(20) | NOT NULL | 发布类型 |
| target_instances | TEXT[] | | 目标实例 |
| status | VARCHAR(20) | DEFAULT 'pending' | 发布状态 |
| released_by | VARCHAR(100) | | 发布人 |
| released_at | TIMESTAMP | | 发布时间 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |

### 3.4 配置审计表 (config_audits)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | UUID主键 |
| config_id | VARCHAR(36) | FOREIGN KEY | 配置项ID |
| action | VARCHAR(50) | NOT NULL | 操作类型 |
| old_value | TEXT | | 旧值 |
| new_value | TEXT | | 新值 |
| operator | VARCHAR(100) | | 操作人 |
| ip_address | VARCHAR(50) | | IP地址 |
| created_at | TIMESTAMP | DEFAULT NOW() | 创建时间 |

---

## 4. 多环境配置策略

### 4.1 环境类型

| 环境 | 标识 | 说明 |
|------|------|------|
| 开发环境 | dev | 开发人员本地开发使用 |
| 测试环境 | test | 集成测试、系统测试使用 |
| 生产环境 | prod | 正式生产环境 |
| 单机模式 | standalone | 无外部依赖的独立运行模式 |

### 4.2 配置文件命名规范

```
configs/
├── config.yaml              # 基础配置（所有环境共享）
├── config-dev.yaml          # 开发环境配置
├── config-test.yaml         # 测试环境配置
├── config-prod.yaml         # 生产环境配置
└── config-standalone.yaml   # 单机模式配置
```

### 4.3 配置继承关系

```
config.yaml (基础配置)
    │
    ├── config-dev.yaml (开发环境，覆盖基础配置)
    │
    ├── config-test.yaml (测试环境，覆盖基础配置)
    │
    ├── config-prod.yaml (生产环境，覆盖基础配置)
    │
    └── config-standalone.yaml (单机模式，覆盖基础配置)
```

---

## 5. 灰度发布设计

### 5.1 发布策略

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  创建配置   │ ──► │  选择实例   │ ──► │  灰度发布   │ ──► │  全量发布   │
│   新版本    │     │  灰度比例   │     │  观察效果   │     │  正式生效   │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
```

### 5.2 灰度规则

- 按实例比例：指定百分比实例
- 按实例标签：指定特定标签实例
- 按区域：指定特定区域实例

---

## 6. API接口设计

### 6.1 配置管理接口

#### 6.1.1 获取配置列表

**GET** `/api/v1/configs`

#### 6.1.2 创建配置

**POST** `/api/v1/configs`

#### 6.1.3 更新配置

**PUT** `/api/v1/configs/{id}`

#### 6.1.4 删除配置

**DELETE** `/api/v1/configs/{id}`

### 6.2 版本管理接口

#### 6.2.1 获取配置版本列表

**GET** `/api/v1/configs/{id}/versions`

#### 6.2.2 回滚到指定版本

**POST** `/api/v1/configs/{id}/rollback`

### 6.3 发布管理接口

#### 6.3.1 发布配置

**POST** `/api/v1/configs/{id}/release`

#### 6.3.2 获取发布状态

**GET** `/api/v1/configs/{id}/release/status`

---

## 7. 客户端SDK设计

### 7.1 配置加载示例

```go
package main

import (
    "context"
    "github.com/new-energy-monitoring/pkg/config"
)

func main() {
    // 创建配置加载器
    loader := config.NewLoader(
        config.WithEnv("prod"),                    // 指定环境
        config.WithConfigCenter("nacos:8848"),     // 配置中心地址
        config.WithLocalConfig("./configs"),       // 本地配置目录
        config.WithFallback(true),                 // 启用本地配置兜底
    )
    
    // 加载配置
    cfg, err := loader.Load(context.Background())
    if err != nil {
        panic(err)
    }
    
    // 获取配置值
    dbHost := cfg.GetString("database.host")
    dbPort := cfg.GetInt("database.port")
}
```

### 7.2 配置变更监听

```go
// 监听配置变更
loader.Watch("database.host", func(key string, value interface{}) {
    log.Printf("配置 %s 已变更: %v", key, value)
    // 执行配置变更回调
})
```

---

## 8. 部署说明

### 8.1 单机模式部署

```yaml
# 启动参数
java -jar app.jar --spring.profiles.active=standalone

# 或环境变量
export NEM_ENV=standalone
java -jar app.jar
```

### 8.2 集群模式部署

```yaml
# 启动参数
java -jar app.jar --spring.profiles.active=prod --config.center=nacos:8848

# 或环境变量
export NEM_ENV=prod
export NEM_CONFIG_CENTER=nacos:8848
java -jar app.jar
```

### 8.3 配置中心依赖

| 配置中心 | 支持状态 | 说明 |
|----------|----------|------|
| Nacos | 支持 | 推荐，功能完善 |
| Apollo | 支持 | 支持灰度发布 |
| Consul | 支持 | 支持服务发现 |
| Etcd | 支持 | 轻量级方案 |
