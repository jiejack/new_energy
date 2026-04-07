# 项目结构

本文档介绍新能源监控系统的目录结构和各模块职责。

## 目录结构概览

```
new-energy-monitoring/
├── cmd/                          # 应用入口
│   ├── api-server/               # API服务
│   ├── collector/                # 数据采集服务
│   ├── alarm/                    # 告警服务
│   ├── compute/                  # 计算服务
│   ├── scheduler/                # 调度服务
│   ├── ai-service/               # AI服务
│   └── migrate/                  # 数据库迁移工具
├── internal/                     # 内部代码（不对外暴露）
│   ├── api/                      # API层
│   │   ├── handler/              # 请求处理器
│   │   └── dto/                  # 数据传输对象
│   ├── application/              # 应用层
│   │   └── service/              # 业务服务
│   ├── domain/                   # 领域层
│   │   ├── entity/               # 实体
│   │   ├── repository/           # 仓储接口
│   │   ├── cache/                # 缓存接口
│   │   └── logger/               # 日志接口
│   └── infrastructure/           # 基础设施层
│       ├── persistence/          # 数据持久化
│       ├── cache/                # 缓存实现
│       ├── config/               # 配置加载
│       ├── logger/               # 日志实现
│       └── mq/                   # 消息队列
├── pkg/                          # 公共包（可对外暴露）
│   ├── auth/                     # 认证授权
│   ├── cache/                    # 缓存工具
│   ├── collector/                # 采集器
│   ├── compute/                  # 计算引擎
│   ├── config/                   # 配置工具
│   ├── errors/                   # 错误处理
│   ├── export/                   # 数据导出
│   ├── harness/                  # Harness层
│   ├── monitoring/               # 监控组件
│   ├── processor/                # 数据处理器
│   ├── protocol/                 # 通信协议
│   │   ├── modbus/               # Modbus协议
│   │   ├── iec104/               # IEC104协议
│   │   └── iec61850/             # IEC61850协议
│   ├── qa/                       # QA助手
│   ├── statistics/               # 统计计算
│   └── storage/                  # 存储组件
├── web/                          # 前端项目
│   ├── src/
│   │   ├── api/                  # API调用
│   │   ├── assets/               # 静态资源
│   │   ├── components/           # 公共组件
│   │   ├── composables/          # 组合式函数
│   │   ├── directives/           # 自定义指令
│   │   ├── layouts/              # 布局组件
│   │   ├── plugins/              # 插件
│   │   ├── router/               # 路由配置
│   │   ├── stores/               # 状态管理
│   │   ├── styles/               # 样式文件
│   │   ├── types/                # 类型定义
│   │   ├── utils/                # 工具函数
│   │   └── views/                # 页面组件
│   ├── public/                   # 公共资源
│   ├── e2e/                      # E2E测试
│   └── tests/                    # 测试文件
├── deployments/                  # 部署配置
│   ├── docker/                   # Docker配置
│   └── kubernetes/               # Kubernetes配置
│       └── helm/                 # Helm Chart
├── deploy/                       # 可观测性配置
│   ├── prometheus/               # Prometheus配置
│   ├── grafana/                  # Grafana配置
│   ├── jaeger/                   # Jaeger配置
│   └── alertmanager/             # Alertmanager配置
├── docs/                         # 文档
│   ├── wiki/                     # Wiki文档
│   ├── superpowers/              # AI技能文档
│   └── swagger/                  # API文档
├── scripts/                      # 脚本
│   ├── git-hooks/                # Git钩子
│   ├── migrations/               # 数据库迁移
│   └── performance/              # 性能测试
├── tests/                        # 测试
│   ├── api/                      # API测试
│   ├── helpers/                  # 测试辅助
│   └── performance/              # 性能测试
├── configs/                      # 配置文件
├── k8s/                          # K8s配置（简化版）
├── .github/                      # GitHub配置
│   └── workflows/                # CI/CD工作流
├── go.mod                        # Go模块定义
├── go.sum                        # Go依赖锁定
├── Makefile                      # 构建脚本
├── docker-compose.yml            # Docker Compose配置
├── Dockerfile.backend            # 后端Dockerfile
├── .golangci.yml                 # Go Linter配置
├── .gitignore                    # Git忽略配置
├── CLAUDE.md                     # AI技能配置
└── README.md                     # 项目说明
```

## 模块说明

### cmd/ - 应用入口

存放各服务的启动入口，职责单一，仅负责初始化和启动服务。

| 目录 | 说明 |
|------|------|
| api-server | REST API服务，提供HTTP接口 |
| collector | 数据采集服务，负责设备数据采集 |
| alarm | 告警服务，处理告警检测和通知 |
| compute | 计算服务，执行统计计算任务 |
| scheduler | 调度服务，管理定时任务 |
| ai-service | AI服务，提供智能分析能力 |
| migrate | 数据库迁移工具 |

### internal/ - 内部代码

遵循Go项目布局规范，不对外暴露的代码。

#### api/ - API层

处理HTTP请求，负责参数验证、响应格式化。

```
api/
├── handler/          # 请求处理器
│   ├── alarm_handler.go
│   ├── device_handler.go
│   └── ...
└── dto/              # 数据传输对象
    ├── request.go
    └── response.go
```

#### application/ - 应用层

业务逻辑实现，协调领域对象完成业务功能。

```
application/
└── service/
    ├── alarm_service.go
    ├── device_service.go
    └── ...
```

#### domain/ - 领域层

核心业务模型，包含实体、值对象、仓储接口。

```
domain/
├── entity/           # 实体
├── repository/       # 仓储接口
├── cache/            # 缓存接口
└── logger/           # 日志接口
```

#### infrastructure/ - 基础设施层

技术实现，如数据库、缓存、消息队列等。

```
infrastructure/
├── persistence/      # 数据持久化
├── cache/            # 缓存实现
├── config/           # 配置加载
├── logger/           # 日志实现
└── mq/               # 消息队列
```

### pkg/ - 公共包

可被外部项目引用的公共代码。

| 目录 | 说明 |
|------|------|
| auth | JWT认证、密码加密 |
| cache | 缓存抽象和实现 |
| collector | 数据采集器框架 |
| compute | 公式计算引擎 |
| config | 配置加载工具 |
| errors | 统一错误处理 |
| export | Excel/CSV导出 |
| harness | Harness验证层 |
| monitoring | 监控指标、追踪 |
| processor | 数据处理管道 |
| protocol | 工业协议实现 |
| qa | QA智能助手 |
| statistics | 统计计算 |
| storage | 时序存储 |

### web/ - 前端项目

Vue 3 + TypeScript 单页应用。

| 目录 | 说明 |
|------|------|
| src/api | API调用封装 |
| src/components | 可复用组件 |
| src/composables | 组合式函数 |
| src/router | 路由配置 |
| src/stores | Pinia状态管理 |
| src/views | 页面组件 |
| e2e | Playwright E2E测试 |

### deployments/ - 部署配置

| 目录 | 说明 |
|------|------|
| docker | Docker和Docker Compose配置 |
| kubernetes/helm | Helm Chart部署配置 |

### deploy/ - 可观测性

| 目录 | 说明 |
|------|------|
| prometheus | Prometheus配置和告警规则 |
| grafana | Grafana仪表板 |
| jaeger | 分布式追踪配置 |
| alertmanager | 告警通知配置 |

### docs/ - 文档

| 目录 | 说明 |
|------|------|
| wiki | GitHub Wiki文档 |
| superpowers | AI技能相关文档 |
| swagger | OpenAPI文档 |

### scripts/ - 脚本

| 目录 | 说明 |
|------|------|
| git-hooks | Git钩子脚本 |
| migrations | 数据库迁移SQL |
| performance | 性能测试脚本 |

### tests/ - 测试

| 目录 | 说明 |
|------|------|
| api | API集成测试 |
| helpers | 测试辅助工具 |
| performance | 性能基准测试 |

## 架构分层

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   Web UI     │  │  REST API   │  │  WebSocket  │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │  Services    │  │   DTOs      │  │ Middleware  │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      Harness Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ Validators   │  │ Verifiers   │  │ Constraints │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                       Domain Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │  Entities    │  │ Value Obj   │  │ Repositories│          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ PostgreSQL   │  │   Redis     │  │   Kafka     │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

## 命名规范

### Go代码

| 类型 | 规范 | 示例 |
|------|------|------|
| 包名 | 小写单词 | `alarm`, `device` |
| 文件名 | 小写+下划线 | `alarm_service.go` |
| 结构体 | 大驼峰 | `AlarmService` |
| 接口 | 动词+er | `AlarmRepository` |
| 函数 | 大驼峰(导出)/小驼峰 | `CreateAlarm`, `validateInput` |
| 常量 | 大驼峰或全大写 | `MaxRetryCount`, `MAX_RETRY` |

### 前端代码

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件 | 大驼峰 | `AlarmList.vue` |
| 组合式函数 | use前缀 | `useAlarm.ts` |
| Store | 小驼峰 | `alarmStore.ts` |
| API文件 | 小驼峰 | `alarm.ts` |

---

**最后更新**: 2026-04-07
