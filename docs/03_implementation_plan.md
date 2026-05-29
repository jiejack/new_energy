# 新能源监控系统（NEM）核心功能实施方案

## 文档信息

| 项目 | 内容 |
|------|------|
| 文档名称 | NEM 核心功能实施方案 |
| 版本 | V1.0 |
| 编写日期 | 2026-05-29 |
| 状态 | 评审稿 |
| 适用范围 | 新能源在线监控系统全模块 |

---

## 1. 实施概述

### 1.1 项目背景

新能源在线监控系统（New Energy Monitoring，NEM）是面向光伏、风电、储能等新能源电站的综合监控平台，实现对设备运行状态的实时采集、处理、告警与分析。系统采用 Go 语言开发，基于分层架构（API → Application → Domain → Infrastructure），融合 AI 技术实现智能化运维，需支持百万级采集点位、千万级历史数据存储。

### 1.2 实施目标

| 目标类别 | 具体目标 | 量化指标 |
|----------|----------|----------|
| 功能完整性 | 完成全部核心模块功能闭环 | 所有模块 CRUD + 业务逻辑 100% 打通 |
| 性能达标 | 满足大规模采集与查询需求 | 100 万采集点并发、日增 1000 万数据、查询 P95 ≤ 500ms |
| 质量保障 | 消除已知缺陷，提升代码质量 | 测试覆盖率 ≥ 80%、零 P0 级缺陷 |
| 安全合规 | 满足工业系统安全要求 | 通过安全审计、无高危漏洞 |
| 可运维性 | 支持云原生部署与运维 | Docker/K8s 部署、全链路可观测 |

### 1.3 实施范围

本方案覆盖以下核心模块的实施：

- **AI 模块**：故障诊断、推理服务、知识库、数据采集器
- **监控模块**：链路追踪、指标采集
- **数据采集模块**：采集器框架、IEC104 协议、Modbus 协议
- **存储模块**：数据生命周期管理、时序数据引擎
- **告警模块**：告警检测、规则引擎、通知分发
- **统计模块**：统计计算、任务调度
- **大数据模块**：数据摄取、处理、存储、查询
- **导出模块**：Excel/CSV 数据导出
- **通信模块**：WebSocket 实时推送、Nacos 服务注册

### 1.4 实施原则

1. **渐进式演进**：优先修复已知缺陷，再推进功能完善，最后进行性能优化
2. **分层解耦**：严格遵循 API → Application → Domain → Infrastructure 分层架构
3. **测试驱动**：所有新功能与修复必须先编写测试用例
4. **向后兼容**：数据库迁移采用增量脚本，API 变更保持版本兼容
5. **安全左移**：开发阶段即纳入安全审查，而非事后补救

---

## 2. 技术架构方案

### 2.1 当前架构现状

```
┌─────────────────────────────────────────────────────────────┐
│                     API 层 (Gin + Handler)                   │
│  alarm_handler / device_handler / station_handler / ...      │
├─────────────────────────────────────────────────────────────┤
│                   Application 层 (Service)                   │
│  alarm_service / device_service / fault_service / ...        │
├─────────────────────────────────────────────────────────────┤
│                     Domain 层 (Entity + Repository)          │
│  entity: alarm / device / station / point / ...              │
│  repository: 接口定义                                        │
├─────────────────────────────────────────────────────────────┤
│                  Infrastructure 层 (Persistence + MQ + Cache)│
│  GORM/PostgreSQL / Kafka / Redis / ClickHouse / Doris        │
└─────────────────────────────────────────────────────────────┘
```

**当前架构特征**：

- 采用 DDD 分层架构，层间依赖通过接口解耦
- 使用 Google Wire 进行依赖注入
- 6 个微服务入口：api-server、collector、alarm、compute、ai-service、scheduler
- 公共包（pkg/）包含 64 个包，涵盖 AI、告警、采集、协议、存储等

**已知架构问题**：

| 问题 | 影响范围 | 严重程度 |
|------|----------|----------|
| DryRun 模式存在无限循环风险 | Harness 测试框架 | 高 |
| GORM `autoCreateTime` 被覆盖导致时间戳异常 | 全部实体 | 高 |
| 部分函数过长（>200 行），圈复杂度高 | 多个 Service | 中 |
| 错误处理不一致，部分错误被静默忽略 | 全局 | 中 |
| 部分模块内存存储替代持久化存储 | FaultService、BigDataService | 中 |
| WebSocket Hub 广播存在并发写入风险 | 实时推送 | 中 |

### 2.2 目标架构

```
┌──────────────────────────────────────────────────────────────────────┐
│                         展示层 (Presentation)                         │
│  Web 前端 (Vue3) │ 移动端 APP │ 大屏展示 │ 第三方系统                 │
├──────────────────────────────────────────────────────────────────────┤
│                         网关层 (Gateway)                              │
│  API Gateway (Nginx/Kong) → 认证鉴权 → 限流熔断 → 日志记录           │
├──────────────────────────────────────────────────────────────────────┤
│                         服务层 (Services)                             │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │api-server│ │collector │ │  alarm   │ │ compute  │ │ai-service│  │
│  │  (2-10)  │ │ (3-20)   │ │  (2-5)   │ │  (2-5)   │ │  (2-5)   │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
│  ┌──────────┐ ┌──────────┐                                            │
│  │scheduler │ │ migrate  │                                            │
│  │  (1-3)   │ │  (按需)  │                                            │
│  └──────────┘ └──────────┘                                            │
├──────────────────────────────────────────────────────────────────────┤
│                         数据层 (Data)                                 │
│  PostgreSQL │ Redis │ Kafka │ ClickHouse/Doris │ Milvus               │
├──────────────────────────────────────────────────────────────────────┤
│                         设备层 (Device)                               │
│  IEC 104 │ IEC 61850 │ Modbus │ OPC UA │ 其他协议                    │
└──────────────────────────────────────────────────────────────────────┘
```

**架构改进要点**：

1. **服务通信标准化**：同步调用走 gRPC，异步通信走 Kafka，消除服务间直接数据库访问
2. **数据存储分层**：配置数据 → PostgreSQL，实时数据 → Redis，历史数据 → Doris/ClickHouse，向量数据 → Milvus
3. **可观测性全覆盖**：OpenTelemetry 链路追踪 + Prometheus 指标 + Zap 结构化日志
4. **安全加固**：JWT + RBAC + TLS + 参数化查询 + 输入校验

### 2.3 迁移路径

| 阶段 | 迁移内容 | 风险等级 |
|------|----------|----------|
| 阶段一 | 修复已知缺陷（DryRun 循环、GORM 时间戳、错误处理） | 低 |
| 阶段二 | 完善模块功能（告警规则 CRUD、统计报表、数据导出） | 低 |
| 阶段三 | 架构优化（长函数拆分、内存存储迁移持久化、并发安全） | 中 |
| 阶段四 | 性能调优（缓存策略、批量写入、连接池优化） | 中 |
| 阶段五 | 安全加固（TLS、参数化查询、密钥管理） | 高 |

---

## 3. 模块实施方案

### 3.1 AI 模块

#### 3.1.1 故障诊断（pkg/ai/fault）

**当前状态**：已实现 FaultDetector、FaultClassifier、HealthAssessor、RULPredictor 接口及 FaultService 编排层。但 FaultService 使用内存切片存储故障事件，无持久化能力。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 修复内存存储 | 将 faultEvents 从内存切片迁移至 PostgreSQL 持久化 | P0 | 3d |
| 完善检测器 | 实现基于统计阈值的异常检测器，支持动态阈值调整 | P1 | 5d |
| 完善分类器 | 实现基于规则的故障分类器，支持故障等级映射 | P1 | 3d |
| RUL 预测集成 | 对接推理服务，实现设备剩余寿命预测 | P2 | 5d |
| 故障事件流 | 通过 Kafka 发布故障事件，供告警模块消费 | P1 | 2d |

**关键接口**：

```go
type FaultDetector interface {
    Detect(ctx context.Context, data []*TimeSeriesData) ([]*Anomaly, error)
}

type FaultClassifier interface {
    Classify(ctx context.Context, anomaly *Anomaly) (*FaultClassification, error)
}

type HealthAssessor interface {
    Assess(ctx context.Context, data []*TimeSeriesData) (*HealthAssessment, error)
    PredictRUL(ctx context.Context, data []*TimeSeriesData) (*RULPrediction, error)
}
```

#### 3.1.2 推理服务（pkg/ai/inference）

**当前状态**：已实现 ModelManager、Cache、Service 层及 API 层，支持模型加载、推理缓存、批量推理。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 模型版本管理 | 对接 ModelVersion 实体，支持模型灰度发布与回滚 | P1 | 3d |
| 推理性能优化 | 实现推理结果多级缓存（内存 → Redis → 重新推理） | P1 | 3d |
| GPU 资源调度 | 支持多 GPU 环境下的推理任务调度 | P2 | 5d |
| 推理指标监控 | 集成 Prometheus 指标，监控推理延迟与吞吐 | P1 | 2d |

#### 3.1.3 知识库（pkg/ai/knowledge）

**当前状态**：已实现 Manager、Embedding、RAG、Vector 四个组件，支持文档嵌入、向量检索、RAG 问答。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 向量数据库集成 | 对接 Milvus，替代内存向量存储 | P1 | 5d |
| 知识库管理 API | 实现知识库的增删改查、文档导入导出 | P1 | 3d |
| RAG 优化 | 实现混合检索（向量 + 关键词）、重排序 | P2 | 5d |
| 知识图谱 | 构建设备-故障-处置知识图谱 | P3 | 10d |

#### 3.1.4 数据采集器（pkg/ai/datacollector）

**当前状态**：已实现 API、Cleaner、Importer、Validator、Types，支持数据导入、清洗、验证。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 数据源适配器 | 实现多数据源适配（CSV、Excel、数据库、API） | P1 | 5d |
| 清洗规则引擎 | 支持可配置的数据清洗规则（去重、填充、异常值处理） | P1 | 3d |
| 数据质量报告 | 生成数据质量评估报告，包含完整性、一致性、时效性指标 | P2 | 3d |
| 增量同步 | 支持增量数据采集，避免全量重复导入 | P2 | 3d |

### 3.2 监控模块

#### 3.2.1 链路追踪（pkg/monitoring/tracing）

**当前状态**：已基于 OpenTelemetry 实现完整的 TracerProvider，支持 gRPC/HTTP 导出、多种采样策略、SpanBuilder 模式、业务属性注入。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 全服务接入 | 为 collector、alarm、compute、ai-service、scheduler 接入链路追踪 | P1 | 3d |
| Kafka 消息传播 | 实现 Kafka 消息的 TraceContext 传播 | P1 | 2d |
| 采样策略调优 | 生产环境采用 parentbased + ratio 采样，降低开销 | P2 | 2d |
| 追踪数据存储 | 配置 Jaeger/Tempo 后端存储，支持长期查询 | P2 | 3d |

#### 3.2.2 指标采集（pkg/monitoring/metrics）

**当前状态**：已定义完整的 Prometheus 指标体系，覆盖 API、采集、计算、告警、AI、数据库、缓存、Kafka、业务九大类指标。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 指标埋点完善 | 在各 Service 层关键路径补充指标埋点 | P1 | 5d |
| Grafana 仪表盘 | 完善 API 服务、Go Runtime、PostgreSQL、质量四类仪表盘 | P1 | 3d |
| 告警规则配置 | 配置服务不可用、响应过长、错误率过高、资源不足等告警规则 | P1 | 2d |
| 自定义业务指标 | 增加电站发电量、设备在线率、告警处理率等业务指标 | P2 | 3d |

### 3.3 数据采集模块

#### 3.3.1 采集器框架（pkg/collector）

**当前状态**：已实现完整的采集器接口（Collector）、任务调度（Scheduler）、协程池（Pool）、缓冲区（Buffer），支持周期性/事件触发/一次性三种任务类型。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 采集任务持久化 | 将采集任务配置存储至 PostgreSQL，支持动态加载 | P0 | 3d |
| 采集数据流转 | 实现采集数据 → Kafka → 告警/计算/存储的完整数据流 | P0 | 5d |
| 采集器健康检查 | 完善采集器健康检查与自动恢复机制 | P1 | 2d |
| 采集指标上报 | 集成 Prometheus 采集指标（采集次数、延迟、错误率） | P1 | 2d |
| 采集器热更新 | 支持运行时动态增删采集点，无需重启服务 | P2 | 5d |

#### 3.3.2 IEC104 协议（pkg/protocol/iec104）

**当前状态**：已实现 ASDU 解析、连接管理、Master 站，支持遥信/遥测/遥控/电度四类数据。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 连接池管理 | 实现多连接池，支持大规模设备并发接入 | P1 | 5d |
| 断线重连 | 实现指数退避重连策略，确保连接可靠性 | P0 | 3d |
| ASDU 扩展 | 支持更多 ASDU 类型（时间同步、总召唤等） | P2 | 3d |
| 协议测试工具 | 开发 IEC104 模拟器，用于集成测试 | P2 | 5d |

#### 3.3.3 Modbus 协议（pkg/protocol/modbus）

**当前状态**：已实现 TCP/RTU/ASCII 三种客户端、数据类型转换器、CRC 校验，支持线圈/离散输入/保持寄存器/输入寄存器四类操作。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 整数溢出修复 | 修复 converter.go 和 master.go 中的整数溢出问题（G115） | P0 | 2d |
| 连接池优化 | 实现连接复用，减少 TCP 连接建立开销 | P1 | 3d |
| 批量读写 | 支持 Modbus 批量读写操作，提升采集效率 | P1 | 3d |
| 网关模式 | 支持 Modbus Gateway 模式，统一多从站访问 | P2 | 5d |

### 3.4 存储模块

#### 3.4.1 数据生命周期管理（pkg/storage/lifecycle）

**当前状态**：已实现归档（archive）、备份（backup）、清理（cleanup）、分层存储（tiered）四个子模块。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 冷热数据分离 | 实现基于时间的自动分层：热数据（Redis）→ 温数据（Doris）→ 冷数据（对象存储） | P1 | 5d |
| 自动归档策略 | 配置按天/周/月的自动归档规则 | P1 | 3d |
| 数据恢复 | 实现从备份中恢复数据的功能 | P2 | 3d |
| 存储容量监控 | 实现存储容量预警与自动清理 | P2 | 2d |

#### 3.4.2 时序数据引擎（pkg/storage/timeseries）

**当前状态**：已实现 TimeSeriesDB 统一接口、Doris/ClickHouse 双引擎实现、BatchWriter 批量写入器、QueryOptimizer 查询优化器、ConnectionPool 连接池。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| SQL 注入修复 | 修复 doris.go 中的 SQL 字符串格式化问题（G201），改用参数化查询 | P0 | 2d |
| 批量写入优化 | 优化 BatchWriter 的刷盘策略，支持按大小/时间/数量三重触发 | P1 | 3d |
| 查询缓存 | 实现 QueryCache，缓存热点查询结果 | P1 | 3d |
| 查询限流 | 实现 RateLimiter，防止大查询压垮数据库 | P1 | 2d |
| 降采样策略 | 实现自动降采样，长期数据自动聚合为分钟/小时/天级别 | P2 | 5d |

### 3.5 告警模块

#### 3.5.1 告警检测与规则引擎（pkg/alarm）

**当前状态**：已实现完整的告警子系统，包含规则引擎（rule/engine）、DSL 解析器（rule/dsl）、规则版本管理（rule/version）、检测器（detector）、聚合器（aggregator）、去重器（dedup）、状态机（state）、通知器（notifier）、存储（storage）。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 告警规则 CRUD | 完善告警规则的数据库持久化，实现完整的增删改查 | P0 | 3d |
| 规则热加载 | 支持运行时动态加载/更新告警规则，无需重启 | P1 | 3d |
| 通知渠道完善 | 完善 Email/SMS/Webhook 通知渠道，增加钉钉/企微 | P1 | 5d |
| 告警抑制 | 实现告警抑制与关联分析，减少告警风暴 | P2 | 5d |
| 告警升级 | 实现告警自动升级机制，超时未处理自动升级 | P2 | 3d |

**规则引擎核心流程**：

```
采集数据 → Kafka → 告警检测器 → 规则引擎评估 → 告警聚合 → 去重 → 状态机流转 → 通知分发
```

#### 3.5.2 告警服务集成（internal/application/service）

**当前状态**：已实现 AlarmService、AlarmRuleService、AlarmHarness 及对应的测试用例。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| Service 层完善 | 移除模拟数据，对接真实 Repository | P0 | 2d |
| 告警工作单联动 | 实现 fault_work_order_bridge，告警自动生成工单 | P1 | 3d |
| 告警统计 | 实现告警统计服务（按级别/类型/时间段统计） | P1 | 2d |
| AI 告警分析 | 集成 AI 模块，实现告警根因分析 | P2 | 5d |

### 3.6 统计模块

#### 3.6.1 统计计算器（pkg/statistics/calculator）

**当前状态**：已实现 station（厂站统计）、device（设备统计）、storage（储能统计）、custom（自定义统计）四类计算器。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 统计指标完善 | 补充发电量、收益、效率、碳排放等核心统计指标 | P1 | 5d |
| 同比环比计算 | 实现同比、环比、趋势分析等对比指标 | P1 | 3d |
| 统计结果缓存 | 统计结果写入 Redis 缓存，避免重复计算 | P1 | 2d |
| 自定义公式 | 支持用户自定义统计公式，对接 compute 模块 | P2 | 5d |

#### 3.6.2 任务调度器（pkg/statistics/scheduler）

**当前状态**：已实现 Cron 调度、分布式锁、执行器、监控器。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 弱随机数修复 | 修复 distributed.go 中使用 math/rand 的问题（G404），改用 crypto/rand | P0 | 1d |
| 任务持久化 | 将调度任务配置存储至 PostgreSQL | P1 | 3d |
| 分布式调度 | 完善基于 Redis 的分布式锁，支持多实例调度 | P1 | 3d |
| 任务监控 | 实现调度任务执行状态监控与告警 | P2 | 2d |

### 3.7 大数据模块

#### 3.7.1 数据处理与存储（pkg/bigdata）

**当前状态**：已实现完整的 BigDataService，包含摄取（ingestion）、处理（processing）、存储（storage）、分析（analysis）、可视化（visualization）五个子模块，支持 ClickHouse/Doris 双引擎、Flink 流处理、物化视图、预聚合、多维度分析。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 存储引擎对接 | 完成 ClickHouse/Doris 存储引擎的真实连接实现 | P1 | 5d |
| Flink 集成 | 完成 Flink 流处理任务的真实提交与管理 | P2 | 5d |
| 数据摄取管道 | 实现 Kafka → 处理 → 存储的完整数据管道 | P1 | 5d |
| 预聚合优化 | 实现自动预聚合规则，提升查询性能 | P2 | 3d |
| 数据血缘 | 实现数据血缘追踪，记录数据流转路径 | P3 | 5d |

### 3.8 导出模块

**当前状态**：已实现 Excel（excelize）和 CSV 导出，包含对应的测试用例。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 流式导出 | 实现大数据量流式导出，避免内存溢出 | P1 | 3d |
| 异步导出 | 实现异步导出任务，支持大报表后台生成 | P1 | 3d |
| 导出模板 | 支持自定义导出模板（表头、样式、公式） | P2 | 3d |
| PDF 导出 | 增加 PDF 格式导出支持 | P3 | 5d |

### 3.9 通信模块

#### 3.9.1 WebSocket（pkg/websocket）

**当前状态**：已实现 Hub/Client 模型，支持全局广播、按厂站广播、按用户广播、心跳机制。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 并发安全修复 | 修复 Hub.broadcast 中的并发 map 写入问题 | P0 | 1d |
| 连接认证 | 增加 WebSocket 连接时的 JWT 认证 | P1 | 2d |
| 消息协议 | 定义标准 WebSocket 消息协议（类型、版本、载荷） | P1 | 2d |
| 消息压缩 | 支持 WebSocket 消息压缩，降低带宽占用 | P2 | 2d |
| 集群支持 | 通过 Redis Pub/Sub 实现多实例 WebSocket 消息同步 | P2 | 3d |

#### 3.9.2 Nacos 服务注册（pkg/nacos）

**当前状态**：已实现配置客户端、服务注册、健康检查、Options 配置。

**实施计划**：

| 任务 | 描述 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 全服务接入 | 为所有微服务接入 Nacos 服务注册与发现 | P1 | 3d |
| 配置中心集成 | 实现配置中心动态配置下发与热更新 | P1 | 3d |
| 服务路由 | 基于 Nacos 实现服务间负载均衡路由 | P2 | 3d |
| 灰度发布 | 基于 Nacos 元数据实现灰度发布 | P3 | 5d |

---

## 4. 数据库方案

### 4.1 数据库选型与职责

| 数据库 | 版本 | 职责 | 数据类型 | 保留策略 |
|--------|------|------|----------|----------|
| PostgreSQL | 16+ | 关系数据存储 | 配置数据、业务数据、告警记录 | 永久 |
| Redis | 7+ | 缓存与实时数据 | 实时采集值、设备状态、会话、告警缓存 | 1天-永久 |
| ClickHouse | 24+ | 时序数据分析 | 历史数据、统计数据（分析场景） | 1年+ |
| Doris | 2.0+ | 时序数据存储 | 历史数据、统计数据（报表场景） | 3年+ |
| SQLite | 3+ | 测试数据库 | 单元测试与集成测试 | 临时 |
| Milvus | 2.3+ | 向量数据库 | AI 知识库向量索引 | 永久 |

### 4.2 PostgreSQL 方案

#### 4.2.1 表结构设计

核心表结构已在 `scripts/migrations/` 中定义，包含：

| 迁移脚本 | 内容 |
|----------|------|
| 001_init_schema.sql | 区域、子区域、厂站、设备、采集点基础表 |
| 002_add_alarm_rules.sql | 告警规则表及索引 |
| 003_add_qa_tables.sql | AI 问答会话与消息表 |
| 004_add_system_configs.sql | 系统配置表 |
| 005_add_performance_indexes.sql | 性能优化索引 |
| 006_add_table_partitions.sql | 表分区策略 |
| 007_add_operation_logs.sql | 操作日志与权限表 |
| 008_add_energy_efficiency.sql | 能效分析表 |
| 009_add_carbon_emission.sql | 碳排放表 |
| 010_add_ai_module_tables.sql | AI 模块表 |
| 011_add_alarm_rules_table.sql | 告警规则表补充 |

#### 4.2.2 索引策略

```sql
-- 高频查询索引
CREATE INDEX idx_alarms_triggered_at ON alarms(triggered_at);
CREATE INDEX idx_alarms_station_status ON alarms(station_id, status);
CREATE INDEX idx_devices_station_type ON devices(station_id, type);
CREATE INDEX idx_points_device_type ON points(device_id, type);

-- 组合索引（最左前缀原则）
CREATE INDEX idx_alarms_composite ON alarms(station_id, level, status, triggered_at);
CREATE INDEX idx_points_composite ON points(station_id, device_id, type);
```

#### 4.2.3 分区策略

```sql
-- 告警表按月分区
CREATE TABLE alarms (
    id VARCHAR(36) PRIMARY KEY,
    triggered_at TIMESTAMP NOT NULL,
    ...
) PARTITION BY RANGE (triggered_at);

-- 操作日志按月分区
CREATE TABLE operation_logs (
    id VARCHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    ...
) PARTITION BY RANGE (created_at);
```

### 4.3 ClickHouse 方案

```sql
CREATE TABLE IF NOT EXISTS nem_history_data (
    point_id String,
    station_id String,
    value Float64,
    quality Int32,
    timestamp DateTime
) ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (point_id, timestamp)
TTL timestamp + INTERVAL 1 YEAR
SETTINGS index_granularity = 8192;
```

### 4.4 Doris 方案

```sql
CREATE TABLE IF NOT EXISTS nem_statistics_data (
    station_id VARCHAR(36) NOT NULL,
    dimension VARCHAR(100) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DOUBLE,
    period_type VARCHAR(20) NOT NULL,
    period_start DATETIME NOT NULL
) ENGINE=OLAP
AGGREGATE KEY(station_id, dimension, metric_name, period_type, period_start)
PARTITION BY RANGE(period_start) ()
DISTRIBUTED BY HASH(station_id) BUCKETS 10
PROPERTIES (
    "dynamic_partition.enable" = "true",
    "dynamic_partition.time_unit" = "MONTH",
    "dynamic_partition.start" = "-36",
    "dynamic_partition.end" = "3"
);
```

### 4.5 Redis 方案

| Key 模式 | 类型 | TTL | 用途 |
|----------|------|-----|------|
| `nem:realtime:{point_id}` | Hash | 1天 | 实时采集数据 |
| `nem:device:status:{device_id}` | String | 永久 | 设备在线状态 |
| `nem:alarm:active:{alarm_id}` | Hash | 7天 | 活动告警 |
| `nem:alarm:count` | Hash | 永久 | 告警计数 |
| `nem:rule:cache:{rule_id}` | String | 5分钟 | 告警规则缓存 |
| `nem:stats:{station_id}:{date}` | Hash | 7天 | 统计结果缓存 |
| `nem:session:{token}` | String | 2小时 | 用户会话 |
| `nem:lock:scheduler:{task_id}` | String | 30秒 | 分布式锁 |

### 4.6 SQLite 测试方案

```go
// 测试中使用 SQLite 替代 PostgreSQL
func NewTestDatabase() (*Database, error) {
    cfg := DatabaseConfig{
        MaxOpenConns:    10,
        MaxIdleConns:    5,
        ConnMaxLifetime: time.Hour,
        ConnMaxIdleTime: 10 * time.Minute,
    }
    return NewDatabaseWithDialector(sqlite.Open(":memory:"), cfg)
}
```

**注意事项**：

- SQLite 不支持部分 PostgreSQL 特有语法（如 `RETURNING`、`ON CONFLICT`）
- 测试用例需避免使用 PostgreSQL 专有函数
- 迁移脚本需兼容 SQLite 语法

### 4.7 数据迁移策略

| 策略 | 描述 |
|------|------|
| 增量迁移 | 使用编号迁移脚本（001_xxx.sql、002_xxx.sql），每次变更新增脚本 |
| 向后兼容 | 新增列允许 NULL 或设置默认值，删除列采用软删除 |
| 回滚支持 | 每个迁移脚本提供对应的回滚脚本 |
| 数据备份 | PostgreSQL 每日全量备份 + WAL 增量备份；Redis RDB 快照 + AOF 日志；Doris 冷数据归档至对象存储 |

---

## 5. 部署方案

### 5.1 容器化部署

#### 5.1.1 Docker 镜像

项目已为各服务提供独立 Dockerfile：

| 服务 | Dockerfile | 基础镜像 | 暴露端口 |
|------|-----------|----------|----------|
| api-server | Dockerfile.backend | golang:1.24-alpine | 8080, 9090 |
| collector | Dockerfile.collector | golang:1.24-alpine | 8081 |
| alarm | Dockerfile.alarm | golang:1.24-alpine | 8082 |
| compute | Dockerfile.compute | golang:1.24-alpine | 8083 |
| ai-service | Dockerfile.ai-service | golang:1.24-alpine | 8084 |
| scheduler | Dockerfile.backend | golang:1.24-alpine | 8085 |
| frontend | Dockerfile.frontend | nginx:alpine | 80 |

#### 5.1.2 Docker Compose

开发环境使用 `docker-compose.yml`，完整环境使用 `docker-compose.full.yml`：

```yaml
# 核心服务编排
services:
  api-server:
    build: { context: ., dockerfile: ops/docker/Dockerfile.backend }
    ports: ["8080:8080", "9090:9090"]
    depends_on: [postgres, redis, kafka]
    environment:
      - DB_HOST=postgres
      - REDIS_ADDR=redis:6379
      - KAFKA_BROKERS=kafka:9092

  collector:
    build: { context: ., dockerfile: ops/docker/Dockerfile.collector }
    depends_on: [kafka, redis]

  # ... 其他服务
```

### 5.2 Kubernetes 部署

#### 5.2.1 资源清单

| 清单文件 | 内容 |
|----------|------|
| 01-namespace.yaml | 命名空间 new-energy |
| 02-configmap.yaml | 应用配置 |
| 03-secrets.yaml | 敏感配置（数据库密码、JWT 密钥） |
| 04-postgres.yaml | PostgreSQL StatefulSet |
| 05-redis.yaml | Redis StatefulSet |
| 06-kafka.yaml | Kafka StatefulSet |
| 07-api-server.yaml | API 服务器 Deployment + Service |
| 08-microservices.yaml | 微服务 Deployment + Service |
| 09-frontend-monitoring.yaml | 前端与监控组件 |

#### 5.2.2 Helm Chart

项目已提供 Helm Chart（`deployments/kubernetes/helm/`），支持多环境配置：

| Values 文件 | 环境 | 特点 |
|-------------|------|------|
| values.yaml | 默认 | 基础配置 |
| values-dev.yaml | 开发 | 单副本、调试模式、资源限制宽松 |
| values-prod.yaml | 生产 | 多副本、高可用、资源限制严格、HPA 自动扩缩 |

#### 5.2.3 资源配额

| 服务 | CPU 请求 | CPU 限制 | 内存请求 | 内存限制 | 副本数 |
|------|----------|----------|----------|----------|--------|
| api-server | 500m | 2000m | 512Mi | 2Gi | 2-10 |
| collector | 1000m | 4000m | 1Gi | 4Gi | 3-20 |
| alarm | 500m | 2000m | 512Mi | 2Gi | 2-5 |
| compute | 500m | 2000m | 512Mi | 2Gi | 2-5 |
| ai-service | 1000m | 4000m | 2Gi | 8Gi | 2-5 |
| scheduler | 250m | 1000m | 256Mi | 1Gi | 1-3 |
| PostgreSQL | 1000m | 4000m | 2Gi | 8Gi | 1（主从） |
| Redis | 500m | 2000m | 1Gi | 4Gi | 3（哨兵） |
| Kafka | 1000m | 2000m | 2Gi | 4Gi | 3（集群） |

### 5.3 配置管理

| 环境 | 配置文件 | 配置中心 | 说明 |
|------|----------|----------|------|
| 开发 | config-dev.yaml | 本地文件 | 调试模式、单机部署 |
| 测试 | config-test.yaml | 本地文件 | SQLite + 内存组件 |
| 独立 | config-standalone.yaml | 本地文件 | 单机全功能部署 |
| 生产 | config-prod.yaml | Nacos | 高可用、集群部署 |

**配置优先级**：环境变量 > Nacos 配置中心 > 配置文件 > 默认值

---

## 6. 安全方案

### 6.1 认证与授权

#### 6.1.1 JWT 认证

```yaml
auth:
  jwt:
    secret: ${JWT_SECRET}          # 从环境变量读取，禁止硬编码
    access_expire: 7200            # Access Token 2小时过期
    refresh_expire: 604800         # Refresh Token 7天过期
  password:
    min_length: 8
    require_uppercase: true
    require_lowercase: true
    require_digit: true
  login:
    max_attempts: 5                # 最大登录尝试次数
    lock_duration: 1800            # 锁定30分钟
```

#### 6.1.2 RBAC 权限模型

```
用户(User) → 角色(Role) → 权限(Permission)
                                ↓
                    资源(Resource) + 操作(Action)
```

**预定义角色**：

| 角色 | 权限范围 |
|------|----------|
| super_admin | 全部权限 |
| station_admin | 所属厂站全部权限 |
| operator | 设备监控、告警处理、数据查询 |
| analyst | 数据查询、报表导出、统计分析 |
| viewer | 只读权限 |

### 6.2 数据加密

| 场景 | 加密方式 | 说明 |
|------|----------|------|
| 密码存储 | bcrypt | 自适应哈希，cost factor ≥ 12 |
| 敏感字段 | AES-256-GCM | 数据库中的手机号、身份证等 |
| 传输加密 | TLS 1.3 | 所有服务间通信 |
| JWT 签名 | HS256/RS256 | 对称/非对称签名 |
| 数据库连接 | SSL | PostgreSQL/ClickHouse SSL 连接 |

### 6.3 输入验证

| 验证类型 | 实现方式 | 覆盖范围 |
|----------|----------|----------|
| 请求参数 | Gin binding tag + 自定义验证器 | 所有 Handler 入参 |
| SQL 注入 | GORM 参数化查询 + 禁止字符串拼接 | 全部数据库操作 |
| XSS | HTML 转义 + CSP 头 | 前端输入输出 |
| CSRF | SameSite Cookie + CSRF Token | 状态变更请求 |
| 文件上传 | 文件类型白名单 + 大小限制 | 导入功能 |

### 6.4 安全加固清单

基于安全审计报告，需修复以下问题：

| 优先级 | 问题 | 修复方案 | 负责模块 |
|--------|------|----------|----------|
| P0 | TLS InsecureSkipVerify（G402） | 生产环境强制启用证书验证 | pkg/ai/service |
| P0 | 弱随机数生成器（G404） | 替换 math/rand 为 crypto/rand | pkg/ai/service, pkg/statistics/scheduler |
| P0 | SQL 字符串格式化（G201） | 改用参数化查询 | pkg/storage/timeseries/doris |
| P0 | 密码硬编码 | 迁移至 Kubernetes Secrets / 环境变量 | CI/CD, Helm values |
| P1 | Slowloris 攻击风险（G112） | 配置 ReadHeaderTimeout | pkg/monitoring |
| P1 | 整数溢出（G115） | 添加边界检查 | pkg/protocol/modbus |

---

## 7. 性能优化方案

### 7.1 缓存策略

#### 7.1.1 多级缓存架构

```
请求 → 本地缓存(BigCache) → Redis 缓存 → 数据库
         ↓ 命中返回           ↓ 命中返回      ↓ 查询并回填
```

| 缓存层级 | 技术 | TTL | 适用场景 |
|----------|------|-----|----------|
| L1 本地缓存 | BigCache | 30s-5min | 配置数据、规则数据、热点查询 |
| L2 分布式缓存 | Redis | 5min-1day | 实时数据、统计结果、会话数据 |
| L3 数据库 | PostgreSQL/Doris | 持久 | 全量数据 |

#### 7.1.2 缓存策略

| 数据类型 | 缓存策略 | 失效机制 |
|----------|----------|----------|
| 配置数据 | Cache-Aside | 配置变更时主动失效 |
| 实时数据 | Write-Through | 采集写入时同步更新缓存 |
| 统计结果 | Cache-Aside | 定时任务计算后更新缓存 |
| 告警规则 | Cache-Aside + Refresh-Ahead | 规则变更时主动失效 + 定时预加载 |
| 查询结果 | Cache-Aside | TTL 过期自动失效 |

### 7.2 连接池优化

| 组件 | 参数 | 开发环境 | 生产环境 |
|------|------|----------|----------|
| PostgreSQL | MaxOpenConns | 25 | 100 |
| PostgreSQL | MaxIdleConns | 5 | 20 |
| PostgreSQL | ConnMaxLifetime | 1h | 1h |
| PostgreSQL | ConnMaxIdleTime | 10min | 10min |
| Redis | PoolSize | 10 | 100 |
| ClickHouse | MaxOpenConns | 10 | 100 |
| ClickHouse | MaxIdleConns | 5 | 20 |
| Doris | MaxOpenConns | 10 | 100 |
| Doris | MaxIdleConns | 5 | 20 |

### 7.3 批量处理

| 场景 | 批量策略 | 批量大小 | 刷盘触发 |
|------|----------|----------|----------|
| 时序数据写入 | BatchWriter | 10000 条 | 大小/时间/数量三重触发 |
| Kafka 消息生产 | 批量发送 | 100 条/批 | 10ms 或 100 条 |
| Kafka 消息消费 | 批量消费 | 500 条/批 | 100ms 或 500 条 |
| 告警规则评估 | 批量评估 | 按规则组 | 按优先级分组评估 |
| 数据导出 | 流式写入 | 1000 行/批 | excelize 流式 API |

### 7.4 查询优化

| 优化策略 | 实现方式 | 预期收益 |
|----------|----------|----------|
| 索引优化 | 高频查询字段建索引、组合索引最左前缀 | 查询耗时降低 50%+ |
| 分区裁剪 | 时序数据按天分区，查询自动裁剪 | 减少扫描数据量 |
| 预聚合 | 定时任务预计算分钟/小时/天级聚合 | 报表查询秒级响应 |
| 查询缓存 | Redis 缓存热点查询结果 | 重复查询零耗时 |
| 降采样 | 长时间范围查询自动降采样 | 减少返回数据量 |
| 分页查询 | 强制分页，限制单次返回条数 | 避免大查询压垮系统 |

---

## 8. 实施里程碑

### 8.1 里程碑总览

| 里程碑 | 名称 | 时间 | 核心交付物 | 验收标准 |
|--------|------|------|------------|----------|
| M1 | 缺陷修复与安全加固 | 第1-3周 | 已知缺陷全部修复、安全漏洞全部关闭 | 零 P0 缺陷、安全扫描无高危 |
| M2 | 核心功能完善 | 第4-8周 | 告警规则 CRUD、统计报表、数据导出 | 功能测试 100% 通过 |
| M3 | 数据采集与存储优化 | 第9-12周 | 采集数据流打通、时序引擎优化 | 100 万点位采集、查询 P95 ≤ 500ms |
| M4 | AI 模块与大数据 | 第13-17周 | 故障诊断闭环、大数据管道 | AI 推理可用、数据管道端到端 |
| M5 | 集成测试与性能调优 | 第18-20周 | 全链路集成测试、性能达标 | 覆盖率 ≥ 80%、性能指标达标 |
| M6 | 部署上线与运维 | 第21-24周 | 生产环境部署、运维体系建立 | 系统可用性 ≥ 99.9% |

### 8.2 详细里程碑计划

#### M1：缺陷修复与安全加固（第1-3周）

| 周次 | 任务 | 交付物 |
|------|------|--------|
| W1 | 修复 DryRun 无限循环、GORM autoCreateTime 覆盖问题 | 修复补丁 + 回归测试 |
| W1 | 修复 SQL 注入风险（doris.go G201） | 参数化查询改造 |
| W2 | 修复弱随机数生成器（G404）、整数溢出（G115） | crypto/rand 替换、边界检查 |
| W2 | 修复 TLS InsecureSkipVerify（G402）、Slowloris（G112） | 安全配置加固 |
| W3 | 密码硬编码清理、Kubernetes Secrets 配置 | Secrets 管理方案 |
| W3 | WebSocket 并发安全修复、错误处理统一 | 并发安全补丁、错误处理规范 |

#### M2：核心功能完善（第4-8周）

| 周次 | 任务 | 交付物 |
|------|------|--------|
| W4 | 告警规则 Repository/Service/Handler 层完善 | 告警规则 CRUD API |
| W5 | 告警规则前端页面完善、通知渠道完善 | 前后端集成完成 |
| W6 | 统计报表 Repository/Service 层实现 | 统计报表生成 API |
| W7 | 数据导出（Excel/CSV）功能完善 | 导出 API + 流式下载 |
| W8 | 统计报表前端页面完善、其他模块前后端集成 | 全功能闭环 |

#### M3：数据采集与存储优化（第9-12周）

| 周次 | 任务 | 交付物 |
|------|------|--------|
| W9 | 采集任务持久化、采集数据 Kafka 流转 | 采集数据管道 |
| W10 | IEC104/Modbus 连接池与断线重连 | 协议栈稳定性提升 |
| W11 | 时序引擎批量写入优化、查询缓存 | 写入/查询性能提升 |
| W12 | 数据生命周期管理（冷热分离、自动归档） | 存储管理自动化 |

#### M4：AI 模块与大数据（第13-17周）

| 周次 | 任务 | 交付物 |
|------|------|--------|
| W13 | 故障诊断持久化、检测器/分类器完善 | 故障诊断闭环 |
| W14 | 推理服务优化、知识库 Milvus 对接 | AI 推理可用 |
| W15 | 大数据存储引擎对接、数据摄取管道 | 大数据管道端到端 |
| W16 | 预聚合、物化视图、多维度分析 | 查询分析能力 |
| W17 | AI 模块与告警/统计模块集成 | AI 赋能业务 |

#### M5：集成测试与性能调优（第18-20周）

| 周次 | 任务 | 交付物 |
|------|------|--------|
| W18 | 全链路集成测试、端到端测试 | 测试报告 |
| W19 | 性能基准测试、瓶颈分析与优化 | 性能测试报告 |
| W20 | 安全渗透测试、合规性检查 | 安全审计报告 |

#### M6：部署上线与运维（第21-24周）

| 周次 | 任务 | 交付物 |
|------|------|--------|
| W21 | 生产环境 K8s 部署、配置管理 | 部署文档 |
| W22 | 监控告警体系完善、运维手册编写 | 运维体系 |
| W23 | 灰度发布、线上验证 | 上线报告 |
| W24 | 全量发布、运维交接 | 项目验收报告 |

---

## 9. 资源需求

### 9.1 团队资源

| 角色 | 人数 | 职责 | 技能要求 |
|------|------|------|----------|
| 项目经理 | 1 | 项目管理、进度跟踪、风险管控 | PMP、敏捷开发经验 |
| 架构师 | 1 | 技术架构设计、方案评审、技术攻关 | Go 微服务、DDD、云原生 |
| 后端开发（高级） | 3 | 核心模块开发、性能优化、代码审查 | Go、GORM、Kafka、Redis |
| 后端开发（中级） | 4 | 业务模块开发、测试用例编写 | Go、Gin、PostgreSQL |
| AI 工程师 | 2 | AI 模块开发、模型部署、知识库建设 | Python、LLM、RAG、Milvus |
| 前端开发 | 2 | 前端页面开发、前后端集成 | Vue3、TypeScript、ECharts |
| 协议开发 | 1 | IEC104/Modbus/IEC61850 协议开发 | 工业协议、嵌入式开发 |
| 测试工程师 | 2 | 测试用例设计、自动化测试、性能测试 | Go testing、k6、Playwright |
| DevOps 工程师 | 1 | CI/CD、K8s 运维、监控体系 | Docker、K8s、Prometheus |
| **合计** | **17** | | |

### 9.2 基础设施资源

#### 9.2.1 开发环境

| 资源 | 规格 | 数量 | 用途 |
|------|------|------|------|
| 开发服务器 | 8C16G | 5 | 开发调试 |
| PostgreSQL | 4C8G | 1 | 开发数据库 |
| Redis | 2C4G | 1 | 开发缓存 |
| Kafka | 4C8G | 1 | 开发消息队列 |

#### 9.2.2 测试环境

| 资源 | 规格 | 数量 | 用途 |
|------|------|------|------|
| K8s Worker | 8C32G | 3 | 服务部署 |
| PostgreSQL | 8C32G | 1 | 测试数据库 |
| Redis | 4C16G | 3 | 哨兵集群 |
| Kafka | 4C16G | 3 | 测试集群 |
| Doris FE | 4C8G | 1 | 元数据管理 |
| Doris BE | 8C32G | 3 | 数据存储 |
| ClickHouse | 8C32G | 2 | 分析引擎 |

#### 9.2.3 生产环境

| 资源 | 规格 | 数量 | 用途 |
|------|------|------|------|
| K8s Worker | 16C64G | 6 | 服务部署 |
| PostgreSQL | 16C64G | 2 | 主从数据库 |
| Redis | 8C32G | 3 | 哨兵集群 |
| Kafka | 8C32G | 3 | 消息集群 |
| Doris FE | 8C16G | 3 | 元数据高可用 |
| Doris BE | 16C64G | 5 | 数据存储 |
| ClickHouse | 16C64G | 3 | 分析集群 |
| Milvus | 8C32G | 2 | 向量数据库 |
| 监控栈 | 8C32G | 2 | Prometheus/Grafana/Jaeger |

### 9.3 工具链

| 类别 | 工具 | 用途 |
|------|------|------|
| 代码管理 | Git + GitHub | 版本控制、代码审查 |
| CI/CD | GitHub Actions | 自动化构建、测试、部署 |
| 容器化 | Docker + Kubernetes | 容器编排 |
| 包管理 | Go Modules | 依赖管理 |
| 代码质量 | golangci-lint | 静态代码分析 |
| 安全扫描 | gosec + nancy + gitleaks | 安全漏洞检测 |
| 测试 | Go testing + testify + k6 | 单元测试 + 压力测试 |
| API 文档 | Swagger (swag) | API 文档生成 |
| 监控 | Prometheus + Grafana | 指标采集与可视化 |
| 链路追踪 | OpenTelemetry + Jaeger | 分布式追踪 |
| 日志 | Zap + Loki | 结构化日志与聚合 |
| 项目管理 | GitHub Projects | 任务跟踪 |

---

## 10. 风险与应对

### 10.1 风险评估矩阵

| 编号 | 风险描述 | 可能性 | 影响 | 风险等级 | 应对策略 |
|------|----------|--------|------|----------|----------|
| R1 | DryRun 无限循环导致服务崩溃 | 高 | 高 | **极高** | 优先修复，增加循环退出条件与超时控制 |
| R2 | GORM autoCreateTime 覆盖导致数据时间戳异常 | 高 | 高 | **极高** | 统一审计所有实体时间字段，改用钩子函数 |
| R3 | SQL 注入攻击（doris.go） | 中 | 高 | **高** | 立即修复为参数化查询，代码审查强制检查 |
| R4 | 大规模采集时性能不达标 | 中 | 高 | **高** | 提前进行性能基准测试，制定降级方案 |
| R5 | AI 模型推理延迟过高 | 中 | 中 | **中** | 实现推理缓存、模型量化、异步推理 |
| R6 | Kafka 消息积压 | 中 | 高 | **高** | 实现消费者监控、自动扩缩容、死信队列 |
| R7 | 密码硬编码泄露 | 高 | 高 | **极高** | 立即迁移至 Secrets 管理，CI 扫描拦截 |
| R8 | WebSocket 并发写入导致 panic | 中 | 中 | **中** | 修复并发安全问题，增加写锁保护 |
| R9 | 数据库迁移失败 | 低 | 高 | **中** | 迁移前备份、灰度执行、回滚脚本准备 |
| R10 | 团队资源不足 | 中 | 中 | **中** | 优先级排序、核心功能先行、非核心功能延后 |
| R11 | 第三方依赖漏洞 | 中 | 中 | **中** | 定期依赖扫描、及时升级、锁定版本 |
| R12 | Doris/ClickHouse 集群故障 | 低 | 高 | **中** | 多副本部署、自动故障转移、数据备份恢复 |

### 10.2 应对措施详情

#### R1：DryRun 无限循环

- **根因**：Harness DryRun 模式缺少循环终止条件
- **修复方案**：增加最大迭代次数限制（如 1000 次）和总超时控制（如 30s）
- **验证**：编写专项测试用例，模拟 DryRun 长时间运行场景

#### R2：GORM autoCreateTime 覆盖

- **根因**：GORM `autoCreateTime` tag 在特定场景下被业务代码覆盖
- **修复方案**：
  1. 审计所有实体的 `CreatedAt`/`UpdatedAt` 字段
  2. 统一使用 GORM 钩子（`BeforeCreate`/`BeforeUpdate`）设置时间
  3. 禁止在业务代码中手动设置 `CreatedAt`
- **验证**：编写测试用例验证时间字段自动填充

#### R3：SQL 注入

- **根因**：doris.go 中使用 `fmt.Sprintf` 拼接 SQL
- **修复方案**：全部改用参数化查询，使用 `?` 占位符
- **验证**：gosec 扫描无 G201 告警

#### R7：密码硬编码

- **根因**：CI 配置和 Helm values 中硬编码数据库密码
- **修复方案**：
  1. CI 使用 GitHub Secrets
  2. K8s 使用 Secrets 资源
  3. 应用通过环境变量读取
- **验证**：gitleaks 扫描无泄露告警

### 10.3 应急预案

| 场景 | 应急措施 | 恢复时间目标 |
|------|----------|-------------|
| 服务不可用 | K8s 自动重启 + 健康检查 | ≤ 5 分钟 |
| 数据库故障 | 主从切换 + 只读模式 | ≤ 15 分钟 |
| Kafka 不可用 | 采集数据本地缓存 + 重试 | ≤ 30 分钟 |
| Redis 故障 | 本地缓存降级 + 限流 | ≤ 10 分钟 |
| 全局故障 | 灾备环境切换 | ≤ 2 小时 |

---

## 附录

### A. 术语表

| 术语 | 说明 |
|------|------|
| DDD | 领域驱动设计（Domain-Driven Design） |
| RBAC | 基于角色的访问控制（Role-Based Access Control） |
| RUL | 剩余使用寿命（Remaining Useful Life） |
| ASDU | 应用服务数据单元（IEC 104 协议） |
| DSL | 领域特定语言（Domain-Specific Language） |
| HPA | 水平 Pod 自动扩缩（Horizontal Pod Autoscaler） |
| TTL | 生存时间（Time To Live） |
| OLAP | 联机分析处理（Online Analytical Processing） |
| RAG | 检索增强生成（Retrieval-Augmented Generation） |

### B. 参考文档

| 文档 | 路径 |
|------|------|
| 系统架构设计 | docs/architecture.md |
| 数据库设计 | docs/database-design.md |
| 安全审计报告 | docs/security-audit-report.md |
| API 文档 | docs/swagger.yaml |
| 部署指南 | docs/DEPLOYMENT.md |
| 开发者指南 | docs/developer-guide.md |
| 测试计划 | docs/test-plan.md |
| 性能基准 | docs/performance-benchmark.md |
| 项目配置 | CLAUDE.md |
| 核心功能需求 | .trae/specs/nem-core-features/spec.md |
| 核心功能任务 | .trae/specs/nem-core-features/tasks.md |

### C. 变更记录

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|----------|------|
| V1.0 | 2026-05-29 | 初始版本 | NEM Team |
