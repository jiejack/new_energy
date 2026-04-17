# 项目结构

本文档介绍新能源监控系统的目录结构和各模块职责。

## 目录结构概览

```
new-energy-monitoring/
├── .github/                # GitHub Actions 工作流
├── .trae/                  # Trae 配置文件
├── api/                    # API 相关文档
│   └── docs/               # Swagger 文档
├── changelogs/             # 版本变更记录
├── cmd/                    # 应用入口
│   ├── api-server/         # API 服务
│   ├── collector/          # 数据采集服务
│   ├── alarm/              # 告警服务
│   ├── compute/            # 计算服务
│   ├── ai-service/         # AI 服务
│   ├── scheduler/          # 调度服务
│   └── migrate/            # 数据库迁移
├── configs/                # 配置文件
├── deploy/                 # 部署配置
├── deployments/            # 部署相关文件
├── docs/                   # 文档
│   ├── wiki/               # Wiki 文档
│   ├── plans/              # 计划文档
│   └── ...                 # 其他文档
├── ops/                    # 运维配置
│   ├── docker/             # Docker 配置
│   ├── k8s/                # Kubernetes 配置
│   ├── monitoring/         # 监控配置
│   └── scripts/            # 部署脚本
├── pkg/                    # 公共包
│   ├── auth/               # 认证相关
│   ├── bigdata/            # 大数据处理
│   ├── cache/              # 缓存
│   ├── collector/          # 采集相关
│   ├── config/             # 配置管理
│   ├── errors/             # 错误处理
│   ├── export/             # 导出功能
│   ├── feedback/           # 反馈机制
│   ├── harness/            # 验证框架
│   ├── monitoring/         # 监控相关
│   ├── nacos/              # Nacos 集成
│   ├── processor/          # 数据处理
│   ├── qa/                 # QA 相关
│   ├── skills/             # 技能模块
│   └── websocket/          # WebSocket 支持
├── scripts/                # 开发脚本
│   ├── git-hooks/          # Git 钩子
│   ├── migrations/         # 数据库迁移脚本
│   └── performance/        # 性能测试脚本
├── tests/                  # 测试文件
│   ├── api/                # API 测试
│   ├── helpers/            # 测试辅助工具
│   └── performance/        # 性能测试
├── web/                    # 前端项目
├── AGENT.md                # Agent 相关文档
├── CLAUDE.md               # Claude 相关文档
├── Makefile                # 构建脚本
├── README.md               # 项目说明
├── go.mod                  # Go 依赖管理
└── go.sum                  # Go 依赖校验
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

### deploy/ - 部署配置

| 目录 | 说明 |
|------|------|
| 部署相关配置文件 | 包含可观测性配置等 |

### deployments/ - 部署相关文件

| 目录 | 说明 |
|------|------|
| docker | Docker相关配置 |

### ops/ - 运维配置

| 目录 | 说明 |
|------|------|
| docker | Docker 配置文件，包括各服务的 Dockerfile 和 docker-compose 配置 |
| k8s | Kubernetes 配置文件，包括部署、服务、配置映射等 |
| monitoring | 监控配置，包括 Prometheus 和 Grafana 配置 |
| scripts | 部署和运维脚本，包括一键部署脚本等 |

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
