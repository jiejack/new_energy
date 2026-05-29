# 新能源监控系统（NEM）核心功能 — 需求分析报告

| 项目 | 内容 |
|------|------|
| 文档名称 | 需求分析报告 |
| 版本 | V1.0 |
| 编写日期 | 2026-05-29 |
| 状态 | 正式发布 |
| 项目名称 | NEM（New Energy Monitoring）核心功能完善 |

---

## 目录

1. [项目概述](#1-项目概述)
2. [功能需求分析](#2-功能需求分析)
3. [非功能需求分析](#3-非功能需求分析)
4. [技术需求分析](#4-技术需求分析)
5. [接口需求分析](#5-接口需求分析)
6. [数据需求分析](#6-数据需求分析)
7. [约束与假设](#7-约束与假设)
8. [风险分析](#8-风险分析)
9. [需求优先级矩阵](#9-需求优先级矩阵)

---

## 1. 项目概述

### 1.1 项目背景

新能源在线监控系统（NEM）是面向光伏、风电、储能等新能源电站的综合监控平台，旨在实现对设备运行状态的实时采集、处理、告警与分析。随着国家"双碳"战略的深入推进，新能源电站规模快速增长，传统人工巡检模式已无法满足大规模电站群的运维需求，亟需一套智能化、自动化的监控平台来保障电站安全稳定运行。

当前系统已完成基础框架搭建，包括实体层、服务层、处理器层和前端页面，但存在以下突出问题：

- 大部分后端 API 仅返回模拟数据，未真正连接数据库
- 前端页面框架已搭建，但部分功能未完全实现
- AI 智能化模块（故障诊断、功率预测、知识库）尚处于原型阶段
- 23 个包缺少单元测试，整体测试覆盖不均衡
- 已知存在 DryRun 模式下生命周期清理无限循环等缺陷

### 1.2 项目目标

| 目标类别 | 具体目标 |
|----------|----------|
| **功能完善** | 完成告警规则管理、统计报表、数据导出等核心功能的完整实现 |
| **数据贯通** | 实现前后端数据完全打通，消除所有模拟数据 |
| **AI 智能化** | 完善故障诊断、功率预测、智能问答等 AI 模块 |
| **数据采集** | 实现多协议（Modbus/IEC104/IEC61850）数据采集与处理 |
| **质量保障** | 代码测试覆盖率 ≥ 80%，修复已知缺陷 |
| **运维支撑** | 完善监控、告警、日志、链路追踪等运维基础设施 |

### 1.3 目标用户

| 用户角色 | 职责描述 | 核心需求 |
|----------|----------|----------|
| 运维人员 | 电站日常运维、设备巡检 | 实时监控、告警处理、设备控制 |
| 系统管理员 | 系统配置、用户权限管理 | 配置管理、权限控制、操作审计 |
| 数据分析人员 | 数据分析、报表生成 | 统计报表、数据导出、趋势分析 |
| 决策管理层 | 运营决策、资源调配 | 综合看板、效率分析、碳排放追踪 |

### 1.4 系统性能指标

| 指标 | 目标值 |
|------|--------|
| 采集点位数 | 100 万 |
| 定时任务数 | 20 万 |
| 日数据增量 | 1000 万条 |
| 年数据存储量 | ≥ 36 亿条 |
| 查询响应时间 | 平均 ≤ 200ms，95% ≤ 500ms |
| 任务调度延迟 | ≤ 1 秒 |
| 采集延迟 | ≤ 100ms |
| 采集误差 | ≤ 0.1% |
| 系统可用性 | ≥ 99.9% |

---

## 2. 功能需求分析

### 2.1 AI 智能模块

#### 2.1.1 故障诊断（FR-AI-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-AI-001 |
| 需求名称 | 故障诊断与预警 |
| 优先级 | Must Have |

**功能描述：**

- 异常检测：基于时序数据分析，自动识别设备运行异常（温度异常、功率偏差、效率下降等）
- 故障分类：对检测到的异常进行分类（电气故障、机械故障、通信故障、环境故障等）
- 健康评估：评估设备整体健康状态，输出健康评分（0-100）
- 剩余使用寿命预测（RUL）：基于历史数据预测设备剩余可用时长
- 故障事件管理：故障事件的全生命周期管理（创建→分类→评估→预测→处理）

**关键实体：** `FaultDetectionResult`（故障检测结果）、`FaultEvent`（故障事件）

**当前状态：** `pkg/ai/fault` 包已实现 `FaultService`，包含检测器、分类器、评估器的注册与调度机制，但数据源对接和模型集成尚需完善。

#### 2.1.2 功率预测（FR-AI-002）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-AI-002 |
| 需求名称 | 功率预测分析 |
| 优先级 | Must Have |

**功能描述：**

- 光伏功率预测：基于气象数据（辐照度、温度、云量）预测光伏电站发电功率
- 风电功率预测：基于风速、风向、空气密度等数据预测风电场发电功率
- 储能预测：预测储能系统的充放电策略和荷电状态（SOC）
- 预测归因分析：分析预测偏差的主要来源（气象误差、模型误差、设备异常等）
- 预测评估：计算预测准确率、置信区间等指标
- 支持超短期（0-4h）、短期（0-72h）、中期（0-240h）三种预测时间尺度

**关键实体：** `ForecastResult`（预测结果）

**当前状态：** `pkg/ai/forecast` 包已实现 `SolarForecaster`、`WindForecaster`、`EnergyStorageForecaster` 及评估器和归因分析模块。

#### 2.1.3 智能问答（FR-AI-003）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-AI-003 |
| 需求名称 | AI 智能问答 |
| 优先级 | Should Have |

**功能描述：**

- 意图识别：识别用户提问意图（故障查询、配置建议、操作指导等）
- 对话管理：支持多轮对话，维护对话上下文
- 答案生成：基于知识库和实时数据生成准确回答
- 会话管理：支持会话的创建、归档、删除

**关键实体：** `QASession`（问答会话）、`QAMessage`（问答消息）

**当前状态：** `pkg/ai/qa` 包已实现意图识别、对话管理、答案生成模块；`pkg/ai/knowledge` 包已实现 RAG 检索增强生成、向量存储、嵌入管理。

#### 2.1.4 推理服务（FR-AI-004）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-AI-004 |
| 需求名称 | AI 模型推理服务 |
| 优先级 | Must Have |

**功能描述：**

- 单次推理：支持单条数据的实时推理请求
- 批量推理：支持批量数据的异步推理，提供任务状态查询
- 推理缓存：基于两级缓存（内存 + Redis）加速重复请求
- 模型管理：模型版本管理、模型信息查询
- 推理解释：提供特征重要性分析

**当前状态：** `pkg/ai/inference` 包已实现 `InferenceService`，包含推理缓存、批量推理、模型管理等功能。

#### 2.1.5 AI 辅助操作（FR-AI-005）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-AI-005 |
| 需求名称 | AI 辅助操作与配置 |
| 优先级 | Could Have |

**功能描述：**

- 操作解析：将自然语言指令解析为系统操作
- 操作确认：高风险操作需人工确认
- 操作执行：执行已确认的操作指令
- 配置建议：基于历史数据和运行状态提供告警阈值、采集参数等配置建议
- 数据采集器：AI 辅助的数据导入、清洗、验证

**当前状态：** `pkg/ai/operation` 包已实现操作解析、确认、执行流程；`pkg/ai/datacollector` 包已实现数据导入、清洗、验证；`pkg/ai/service` 包已实现 AI 服务适配器和上下文管理。

#### 2.1.6 边缘智能（FR-AI-006）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-AI-006 |
| 需求名称 | 边缘智能计算 |
| 优先级 | Could Have |

**功能描述：**

- 边缘数据处理：在边缘节点进行数据预处理
- 心跳监测：边缘节点心跳检测与状态管理
- 模型服务：边缘端模型部署与推理
- 数据同步：边缘节点与云端数据同步

**关键实体：** `EdgeNode`（边缘节点）

**当前状态：** `pkg/ai/edge` 包已实现数据处理器、心跳管理、模型服务、同步管理器。

### 2.2 实时监控模块

#### 2.2.1 数据采集（FR-MON-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-MON-001 |
| 需求名称 | 多协议数据采集 |
| 优先级 | Must Have |

**功能描述：**

- Modbus 协议采集：支持 Modbus TCP/RTU/ASCII 三种模式
- IEC 104 协议采集：支持电力行业标准通信协议
- IEC 61850 协议采集：支持变电站自动化通信标准（MMS 客户端、采样值、控制操作）
- 采集调度：支持可配置的采集周期和策略
- 连接池管理：管理采集连接的创建、复用和销毁
- 数据缓冲：采集数据的内存缓冲机制
- 数据质量校验：采集数据的质量标记和有效性验证

**当前状态：** `pkg/collector` 包已实现缓冲区、连接池、调度器；`pkg/protocol/modbus`、`pkg/protocol/iec104`、`pkg/protocol/iec61850` 已实现对应协议客户端。

#### 2.2.2 实时数据推送（FR-MON-002）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-MON-002 |
| 需求名称 | WebSocket 实时数据推送 |
| 优先级 | Must Have |

**功能描述：**

- WebSocket 连接管理：支持客户端连接、断开、重连
- 数据广播：实时采集数据的广播推送
- 频道订阅：支持按厂站、设备、采集点订阅数据
- 连接心跳：维持长连接的心跳机制

**当前状态：** `pkg/websocket` 包已实现 Hub、Handler、Realtime 推送功能。

#### 2.2.3 数据处理管道（FR-MON-003）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-MON-003 |
| 需求名称 | 数据处理管道 |
| 优先级 | Must Have |

**功能描述：**

- 变化检测：检测数据变化，过滤无变化数据
- 数据过滤：按条件过滤无效数据
- 数据缩放：数据单位转换和缩放
- 数据校验：数据范围和格式校验
- 质量评估：数据质量评分

**当前状态：** `pkg/processor` 包已实现变化检测、过滤器、管道、质量评估、缩放器、验证器。

#### 2.2.4 可观测性（FR-MON-004）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-MON-004 |
| 需求名称 | 系统可观测性 |
| 优先级 | Should Have |

**功能描述：**

- 指标采集：Prometheus 指标暴露和采集
- 链路追踪：OpenTelemetry 分布式链路追踪
- 健康检查：服务健康状态检测
- 告警规则：基于指标的告警规则管理
- Dashboard 集成：Grafana 仪表盘集成
- 错误监控：系统错误监控和报告

**当前状态：** `pkg/monitoring` 包已实现指标（Prometheus）、追踪（OpenTelemetry）、健康检查、告警、Dashboard（Grafana）、错误监控等子模块。

### 2.3 告警模块

#### 2.3.1 告警检测（FR-ALM-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-ALM-001 |
| 需求名称 | 实时告警检测 |
| 优先级 | Must Have |

**功能描述：**

- 限值告警：采集值超过上下限阈值触发告警
- 状态告警：设备状态变化触发告警
- 通信告警：设备通信中断触发告警
- 系统告警：系统级异常触发告警
- 设备告警：设备故障触发告警
- 告警去重：相同告警的合并和抑制
- 告警聚合：相关告警的聚合展示

**关键实体：** `Alarm`（告警记录）

**告警级别：** 信息（1）→ 警告（2）→ 重要（3）→ 紧急（4）

**告警状态机：** 活跃（Active）→ 已确认（Acknowledged）→ 已清除（Cleared）；支持抑制（Suppressed）

**当前状态：** `pkg/alarm/detector` 包已实现告警检测器；`pkg/alarm/dedup` 已实现告警去重；`pkg/alarm/aggregator` 已实现告警聚合；`pkg/alarm/state` 已实现告警状态机。

#### 2.3.2 告警规则管理（FR-ALM-002）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-ALM-002 |
| 需求名称 | 告警规则管理 |
| 优先级 | Must Have |

**功能描述：**

- 规则 CRUD：告警规则的创建、查询、更新、删除
- 规则类型：限值规则（limit）、趋势规则（trend）、自定义规则（custom）
- 触发条件：支持条件表达式配置
- 持续时间：支持配置触发持续时间
- 通知渠道：支持配置告警通知渠道（邮件、短信、Webhook、微信）
- 通知用户：支持配置告警通知人员
- 规则版本管理：规则变更的版本追踪
- DSL 解析：告警规则 DSL 解析引擎

**关键实体：** `AlarmRule`（告警规则）

**当前状态：** `pkg/alarm/rule` 包已实现 DSL 解析器、规则引擎、规则管理器、版本管理；API Handler 和 Service 层已实现完整 CRUD。

#### 2.3.3 告警通知（FR-ALM-003）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-ALM-003 |
| 需求名称 | 告警通知 |
| 优先级 | Must Have |

**功能描述：**

- 邮件通知：SMTP 邮件发送
- 短信通知：短信网关集成
- Webhook 通知：HTTP 回调通知
- 微信通知：企业微信消息推送
- 通知模板：可配置的通知消息模板
- 通知调度：通知发送的调度和重试
- 通知配置管理：通知渠道的配置和启停

**关键实体：** `NotificationConfig`（通知配置）

**当前状态：** `pkg/alarm/notifier` 包已实现邮件、短信、内部通知、模板引擎、调度器；通知配置的 CRUD 已实现。

### 2.4 统计分析模块

#### 2.4.1 统计计算（FR-STAT-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-STAT-001 |
| 需求名称 | 多维度统计计算 |
| 优先级 | Must Have |

**功能描述：**

- 厂站统计：厂站级别的发电量、功率、效率等统计
- 设备统计：设备级别的运行时长、故障率、效率等统计
- 储能统计：储能系统的充放电量、SOC 等统计
- 自定义统计：支持用户自定义统计指标和计算公式
- 统计调度：定时统计任务的调度执行

**当前状态：** `pkg/statistics/calculator` 包已实现厂站、设备、储能、自定义统计计算器；`pkg/statistics/scheduler` 包已实现 Cron 调度、分布式调度、执行器和监控。

#### 2.4.2 报表生成（FR-STAT-002）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-STAT-002 |
| 需求名称 | 统计报表生成 |
| 优先级 | Must Have |

**功能描述：**

- 报表类型：日报、周报、月报、季报、年报
- 报表维度：按时间、厂站、设备等维度
- 数据可视化：报表数据的图表展示
- 自定义时间段：支持自定义报表时间范围

**当前状态：** `ReportService` 和 `ReportHandler` 已实现，支持多类型报表生成。

### 2.5 数据导出模块（FR-EXP-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-EXP-001 |
| 需求名称 | 数据导出 |
| 优先级 | Must Have |

**功能描述：**

- Excel 导出：基于 excelize 库的 Excel 文件生成
- CSV 导出：CSV 格式数据导出
- 导出模板：支持自定义导出模板
- 大数据量导出：支持流式导出，避免内存溢出

**当前状态：** `pkg/export` 包已实现 Excel 和 CSV 导出功能；`ExportService` 和 `ExportHandler` 已实现。

### 2.6 存储管理模块

#### 2.6.1 时序数据存储（FR-STR-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-STR-001 |
| 需求名称 | 时序数据存储与查询 |
| 优先级 | Must Have |

**功能描述：**

- 双引擎支持：支持 Apache Doris 和 ClickHouse 两种时序数据库
- 批量写入：高效的批量数据写入机制
- 数据压缩：时序数据压缩存储
- 分区管理：按时间自动分区
- 查询优化：查询缓存、查询限流、查询监控
- 工厂模式：通过配置切换存储引擎

**当前状态：** `pkg/storage/timeseries` 包已实现 ClickHouse 和 Doris 适配器、批量写入器、工厂模式；`pkg/storage/compression` 已实现数据压缩；`pkg/storage/partition` 已实现分区管理；`pkg/storage/query` 已实现查询缓存、执行器、限流、监控。

#### 2.6.2 数据生命周期管理（FR-STR-002）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-STR-002 |
| 需求名称 | 数据生命周期管理 |
| 优先级 | Must Have |

**功能描述：**

- 数据清理：按策略自动清理过期数据
- 数据归档：清理前的数据归档
- 数据备份：定期数据备份
- 分层存储：热数据、温数据、冷数据的分层管理
- 清理策略管理：清理策略的 CRUD
- DryRun 模式：试运行模式，仅预览不实际删除
- 清理任务管理：任务的创建、执行、取消、状态查询

**已知缺陷：** DryRun 模式下存在无限循环 Bug（`doCleanup` 方法中，DryRun 模式下 `totalDeleted` 累加但实际未删除记录，导致循环条件永远为真）

**当前状态：** `pkg/storage/lifecycle` 包已实现清理器、归档器、备份器、分层存储管理器。

#### 2.6.3 数据分片与索引（FR-STR-003）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-STR-003 |
| 需求名称 | 数据分片与索引管理 |
| 优先级 | Should Have |

**功能描述：**

- 数据分片：按厂站或时间维度分片
- 索引管理：索引的创建、删除、优化
- 存储统计：表大小、记录数等统计信息

**当前状态：** `pkg/storage/sharding` 已实现分片管理；`pkg/storage/index` 已实现索引管理。

### 2.7 计算引擎模块（FR-CMP-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-CMP-001 |
| 需求名称 | 公式计算与规则引擎 |
| 优先级 | Must Have |

**功能描述：**

- 公式解析：数学表达式解析器
- 公式执行：公式计算执行器
- 内置函数：常用数学函数库
- 规则引擎：基于规则的数据处理引擎
- 规则触发：事件驱动的规则触发机制
- 规则调度：定时规则调度
- 计算缓存：计算结果缓存
- 并发控制：计算任务的并发锁

**当前状态：** `pkg/compute/formula` 包已实现解析器、执行器、函数库、管理器；`pkg/compute/rule` 包已实现规则引擎、触发器、调度器、缓存、并发锁。

### 2.8 大数据模块（FR-BD-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-BD-001 |
| 需求名称 | 大数据处理与分析 |
| 优先级 | Should Have |

**功能描述：**

- 数据摄入：大规模数据的批量摄入
- 数据处理：基于 Flink 的流式数据处理
- 数据分析：多维数据分析
- 数据可视化：分析结果的可视化展示
- ClickHouse/Doris 存储：大数据的列式存储

**当前状态：** `pkg/bigdata` 包已实现数据摄入、处理（Flink 集成）、分析、可视化、存储适配器。

### 2.9 配置管理模块（FR-CFG-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-CFG-001 |
| 需求名称 | 配置中心管理 |
| 优先级 | Must Have |

**功能描述：**

- 配置项 CRUD：配置项的创建、查询、更新、删除
- 多环境支持：dev/test/prod 环境隔离
- 命名空间：配置的命名空间隔离
- 配置分组：配置的分组管理
- 版本管理：配置变更的版本追踪
- 灰度发布：配置的灰度发布和全量发布
- 回滚发布：配置的回滚操作
- 审计日志：配置变更的审计追踪
- 加密存储：敏感配置的加密存储
- Nacos 集成：支持 Nacos 作为配置中心

**关键实体：** `ConfigItem`（配置项）、`ConfigVersion`（配置版本）、`ConfigRelease`（配置发布）、`ConfigAudit`（配置审计）、`SystemConfig`（系统配置）

**当前状态：** 配置管理模块已完整实现，包括多环境、版本管理、灰度发布、审计日志等。

### 2.10 设备与资产管理模块

#### 2.10.1 设备管理（FR-DEV-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-DEV-001 |
| 需求名称 | 设备全生命周期管理 |
| 优先级 | Must Have |

**功能描述：**

- 设备 CRUD：设备的创建、查询、更新、删除
- 设备类型：逆变器、电表、变压器、开关、气象站、储能系统（ESS）、PCS、BMS
- 设备状态：在线、离线、故障、维护
- 通信配置：协议、IP、端口、从站 ID 配置
- 设备生命周期：在役→维护→退役→报废
- 维护记录：设备维护历史记录
- 备件管理：设备备件信息管理
- 设备文档：设备手册、数据表、证书等文档管理

**关键实体：** `Device`、`MaintenanceRecord`、`SparePart`、`DeviceDocument`

#### 2.10.2 资产管理（FR-DEV-002）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-DEV-002 |
| 需求名称 | 资产管理 |
| 优先级 | Should Have |

**功能描述：**

- 资产 CRUD：资产信息的创建、查询、更新、删除
- 折旧计算：直线法、双倍余额递减法、年数总和法、产量法
- 资产维护：资产维护记录管理
- 资产文档：资产相关文档管理

**关键实体：** `Asset`、`AssetDepreciationRecord`、`AssetMaintenanceRecord`、`AssetDocument`

#### 2.10.3 工单管理（FR-DEV-003）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-DEV-003 |
| 需求名称 | 工单管理 |
| 优先级 | Should Have |

**功能描述：**

- 工单 CRUD：工单的创建、查询、更新、删除
- 工单类型：维护、维修、巡检
- 工单优先级：低、中、高、紧急
- 工单状态：待处理→处理中→已完成/已取消
- 故障-工单联动：故障检测结果自动创建工单

**关键实体：** `WorkOrder`

**当前状态：** `FaultWorkOrderBridge` 已实现故障到工单的桥接逻辑。

### 2.11 供应链管理模块（FR-SCM-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-SCM-001 |
| 需求名称 | 供应链与成本管理 |
| 优先级 | Could Have |

**功能描述：**

- 库存管理：物料入库、出库、调拨、盘点
- 供应商管理：供应商信息维护和评估
- 采购订单：采购订单的创建、审批、跟踪
- 收货管理：采购收货的确认和入库
- 成本管理：成本分类、录入、分配、报表
- 成本类别：直接/间接/固定/变动成本

**关键实体：** `Inventory`、`InventoryTransaction`、`Supplier`、`PurchaseOrder`、`PurchaseOrderItem`、`Receipt`、`ReceiptItem`、`CostCategory`、`CostEntry`、`CostAllocation`、`CostReport`

### 2.12 能效与碳排放模块

#### 2.12.1 能效分析（FR-EE-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-EE-001 |
| 需求名称 | 能效分析 |
| 优先级 | Should Have |

**功能描述：**

- 能效记录：设备/厂站/系统/综合能效数据记录
- 能效等级：优秀（≥95%）/良好（≥85%）/一般/较差（<70%）
- 能效分析：平均/最大/最小效率统计，同比环比分析
- 优化建议：基于能效分析结果提供优化建议
- 节能潜力：评估节能潜力

**关键实体：** `EnergyEfficiencyRecord`、`EnergyEfficiencyAnalysis`

#### 2.12.2 碳排放管理（FR-CE-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-CE-001 |
| 需求名称 | 碳排放追踪与管理 |
| 优先级 | Could Have |

**功能描述：**

- 排放因子管理：碳排放因子的维护和版本管理
- 排放记录：按 Scope 1/2/3 分类记录碳排放
- 排放汇总：按周期汇总碳排放数据
- 减排目标：设定和追踪碳减排目标
- 审批流程：碳排放数据的审批管理

**关键实体：** `CarbonEmissionFactor`、`CarbonEmissionRecord`、`CarbonEmissionSummary`、`CarbonReductionTarget`

**当前状态：** 实体和 Repository 已实现，Handler 层暂时注释（`NewCarbonEmissionHandler`）。

### 2.13 用户与权限模块（FR-AUTH-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-AUTH-001 |
| 需求名称 | 用户认证与权限管理 |
| 优先级 | Must Have |

**功能描述：**

- 用户管理：用户的创建、查询、更新、删除
- 角色管理：角色的创建、权限分配
- 权限管理：基于 RBAC 的权限控制
- JWT 认证：基于 JWT 的无状态认证
- 密码策略：密码强度要求（最小长度、大小写、数字）
- 登录安全：登录失败锁定机制
- 操作日志：用户操作的审计追踪

**关键实体：** `User`、`Role`、`Permission`、`OperationLog`

**权限资源类型：** 厂站、设备、采集点、告警、用户、角色、区域

**权限操作类型：** 创建、查看、更新、删除、控制、确认

### 2.14 Harness 测试框架（FR-HARNESS-001）

| 需求项 | 描述 |
|--------|------|
| 需求编号 | FR-HARNESS-001 |
| 需求名称 | Harness 集成测试框架 |
| 优先级 | Should Have |

**功能描述：**

- 约束验证：服务层约束条件验证
- 监控指标：测试过程指标采集
- 快照管理：测试状态快照
- 验证器：测试结果验证
- 验证器：服务行为验证

**当前状态：** `pkg/harness` 包已实现完整的测试框架，`AlarmHarness` 和 `DeviceHarness` 已集成。

---

## 3. 非功能需求分析

### 3.1 性能需求

| 编号 | 需求描述 | 指标 | 优先级 |
|------|----------|------|--------|
| NFR-PERF-001 | API 响应时间 | 平均 ≤ 200ms，95% ≤ 500ms（报表生成除外） | Must Have |
| NFR-PERF-002 | 报表生成时间 | ≤ 10s（10 万条数据以内） | Must Have |
| NFR-PERF-003 | 数据采集延迟 | ≤ 100ms | Must Have |
| NFR-PERF-004 | 数据采集误差 | ≤ 0.1% | Must Have |
| NFR-PERF-005 | 任务调度延迟 | ≤ 1s | Must Have |
| NFR-PERF-006 | 并发采集点位 | ≥ 100 万 | Must Have |
| NFR-PERF-007 | 日数据写入量 | ≥ 1000 万条 | Must Have |
| NFR-PERF-008 | WebSocket 并发连接 | ≥ 10000 | Should Have |
| NFR-PERF-009 | 批量推理吞吐量 | ≥ 1000 条/秒 | Should Have |

### 3.2 安全需求

| 编号 | 需求描述 | 优先级 |
|------|----------|--------|
| NFR-SEC-001 | 传输加密：所有 API 通信使用 HTTPS/TLS | Must Have |
| NFR-SEC-002 | 认证机制：基于 JWT 的无状态认证 | Must Have |
| NFR-SEC-003 | 权限控制：基于 RBAC 的细粒度权限管理 | Must Have |
| NFR-SEC-004 | 密码安全：密码哈希存储，强度策略校验 | Must Have |
| NFR-SEC-005 | 登录安全：连续失败锁定机制（5 次/30 分钟） | Must Have |
| NFR-SEC-006 | 敏感数据加密：配置项加密存储 | Must Have |
| NFR-SEC-007 | 操作审计：关键操作的审计日志记录 | Must Have |
| NFR-SEC-008 | SQL 注入防护：ORM 参数化查询 | Must Have |
| NFR-SEC-009 | 限流熔断：API 限流和熔断保护 | Should Have |
| NFR-SEC-010 | 密钥管理：JWT Secret 等密钥的安全管理 | Must Have |

### 3.3 可靠性需求

| 编号 | 需求描述 | 指标 | 优先级 |
|------|----------|------|--------|
| NFR-REL-001 | 系统可用性 | ≥ 99.9% | Must Have |
| NFR-REL-002 | 数据不丢失 | 采集数据零丢失 | Must Have |
| NFR-REL-003 | 故障转移 | 服务故障自动转移 | Must Have |
| NFR-REL-004 | 优雅停机 | 服务关闭时完成进行中的请求 | Must Have |
| NFR-REL-005 | 数据备份 | 定期数据备份和恢复 | Must Have |
| NFR-REL-006 | 消息持久化 | Kafka 消息持久化存储 | Must Have |
| NFR-REL-007 | 缓存降级 | Redis 不可用时降级到内存缓存 | Should Have |

### 3.4 可扩展性需求

| 编号 | 需求描述 | 优先级 |
|------|----------|--------|
| NFR-SCA-001 | 水平扩展：无状态服务支持水平扩容 | Must Have |
| NFR-SCA-002 | 协议扩展：通过接口抽象支持新协议接入 | Must Have |
| NFR-SCA-003 | 存储扩展：支持 Doris/ClickHouse 存储引擎切换 | Must Have |
| NFR-SCA-004 | 模型扩展：AI 模型的热插拔和版本管理 | Should Have |
| NFR-SCA-005 | 插件化架构：告警通知渠道的插件化扩展 | Should Have |
| NFR-SCA-006 | 多区域部署：支持按区域分库分表 | Could Have |

### 3.5 可维护性需求

| 编号 | 需求描述 | 指标 | 优先级 |
|------|----------|------|--------|
| NFR-MAINT-001 | 代码测试覆盖率 | ≥ 80% | Must Have |
| NFR-MAINT-002 | 分层架构 | 严格遵循 DDD 分层 | Must Have |
| NFR-MAINT-003 | 代码规范 | 遵循 Go 语言编码规范 | Must Have |
| NFR-MAINT-004 | 依赖注入 | 使用 Wire 进行依赖注入 | Must Have |
| NFR-MAINT-005 | 配置外部化 | 所有配置可通过 YAML 文件管理 | Must Have |
| NFR-MAINT-006 | 日志规范 | 结构化日志，统一日志格式 | Must Have |
| NFR-MAINT-007 | API 文档 | Swagger 自动生成 API 文档 | Should Have |
| NFR-MAINT-008 | 数据库迁移 | 版本化数据库迁移脚本 | Must Have |

---

## 4. 技术需求分析

### 4.1 架构需求

#### 4.1.1 微服务架构

| 服务 | 职责 | 技术栈 | 实例数 |
|------|------|--------|--------|
| api-server | RESTful API 服务 | Go + Gin + GORM | 2-10 |
| collector | 多协议数据采集 | Go + 协程池 | 3-20 |
| alarm | 告警检测与通知 | Go + 规则引擎 | 2-5 |
| compute | 公式计算与规则执行 | Go + 表达式引擎 | 2-5 |
| ai-service | AI 推理与智能服务 | Go + LLM | 2-5 |
| scheduler | 定时任务调度 | Go + Cron | 1-3 |

#### 4.1.2 领域驱动设计（DDD）

系统采用 DDD 方法划分限界上下文：

| 限界上下文 | 对应模块 | 核心实体 |
|------------|----------|----------|
| 配置上下文 | config | Region, Station, Device, Point, ConfigItem, SystemConfig |
| 采集上下文 | collector | 协议客户端、数据缓冲、连接池 |
| 告警上下文 | alarm | Alarm, AlarmRule, NotificationConfig |
| 计算上下文 | compute | Formula, Rule |
| 统计上下文 | statistics | Calculator, Scheduler |
| 存储上下文 | storage | TimeSeries, Lifecycle, Partition |
| AI 上下文 | ai | FaultDetection, Forecast, QA, Inference |

#### 4.1.3 分层架构

```
API 层（Handler）→ 应用层（Service）→ 领域层（Entity）→ 基础设施层（Repository/MQ/Cache）
```

- **API 层**：HTTP 请求处理、参数校验、响应封装
- **应用层**：业务逻辑编排、事务管理
- **领域层**：领域实体、业务规则、领域事件
- **基础设施层**：数据持久化、消息队列、缓存、外部服务

### 4.2 数据库需求

| 组件 | 用途 | 版本要求 | 高可用方案 |
|------|------|----------|------------|
| PostgreSQL | 关系数据存储 | 15+ | 主从复制 + 自动故障转移 |
| Redis | 缓存 + 实时数据 | 7+ | Sentinel 哨兵模式 |
| Apache Doris | 时序数据（OLAP） | 2.0+ | FE/BE 多节点部署 |
| ClickHouse | 时序数据（备选） | - | 多副本部署 |
| SQLite | 单元测试 | - | 单机 |

**数据库连接池配置：**

| 参数 | 值 |
|------|-----|
| 最大打开连接数 | 100 |
| 最大空闲连接数 | 10 |
| 连接最大生命周期 | 3600s |
| 连接最大空闲时间 | 600s |

### 4.3 消息队列需求

| 需求项 | 描述 |
|--------|------|
| 消息中间件 | Apache Kafka 3.7+ |
| 主题前缀 | nem |
| 消息持久化 | 启用 |
| 高可用 | 多副本 + 分区 |

**Kafka 主题设计：**

| 主题 | 生产者 | 消费者 | 说明 |
|------|--------|--------|------|
| nem.data.collect | collector | alarm, compute | 采集数据 |
| nem.alarm.event | alarm | notify | 告警事件 |
| nem.alarm.notify | alarm | sms, email, push | 告警通知 |
| nem.device.status | collector | api-server | 设备状态 |
| nem.compute.result | compute | storage | 计算结果 |

### 4.4 缓存需求

| 需求项 | 描述 |
|--------|------|
| 缓存中间件 | Redis 7+ |
| 缓存模式 | 连接池模式（Pool Size: 100） |
| 缓存场景 | 实时数据、配置缓存、会话缓存、推理缓存 |
| 缓存策略 | TTL 过期 + LRU 淘汰 |
| 降级方案 | Redis 不可用时降级到内存缓存（BigCache） |

### 4.5 监控需求

| 需求项 | 技术选型 | 描述 |
|--------|----------|------|
| 指标采集 | Prometheus | 服务指标、系统指标、业务指标 |
| 链路追踪 | OpenTelemetry + Jaeger | 分布式链路追踪 |
| 日志收集 | Loki + Promtail | 日志聚合分析 |
| 可视化 | Grafana | 指标和日志可视化 |
| 告警 | Alertmanager | 基于指标的告警管理 |

### 4.6 服务注册与配置中心

| 需求项 | 技术选型 | 描述 |
|--------|----------|------|
| 服务注册 | Nacos（可选） | 服务注册与发现 |
| 配置中心 | Nacos（可选）/ 内置 | 动态配置管理 |
| 健康检查 | Nacos / 内置 | 服务健康状态检测 |

---

## 5. 接口需求分析

### 5.1 外部 API 接口

#### 5.1.1 认证接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/auth/login | 用户登录 |
| POST | /api/v1/auth/logout | 用户登出 |
| POST | /api/v1/auth/refresh | 刷新令牌 |

#### 5.1.2 用户管理接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/users | 获取用户列表 |
| POST | /api/v1/users | 创建用户 |
| GET | /api/v1/users/:id | 获取用户详情 |
| PUT | /api/v1/users/:id | 更新用户 |
| DELETE | /api/v1/users/:id | 删除用户 |
| PUT | /api/v1/users/:id/password | 修改密码 |

#### 5.1.3 区域管理接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/regions | 获取区域树 |
| POST | /api/v1/regions | 创建区域 |
| PUT | /api/v1/regions/:id | 更新区域 |
| DELETE | /api/v1/regions/:id | 删除区域 |

#### 5.1.4 厂站管理接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/stations | 获取厂站列表 |
| POST | /api/v1/stations | 创建厂站 |
| GET | /api/v1/stations/:id | 获取厂站详情 |
| PUT | /api/v1/stations/:id | 更新厂站 |
| DELETE | /api/v1/stations/:id | 删除厂站 |
| GET | /api/v1/stations/:id/statistics | 获取厂站统计 |

#### 5.1.5 设备管理接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/devices | 获取设备列表 |
| POST | /api/v1/devices | 创建设备 |
| GET | /api/v1/devices/:id | 获取设备详情 |
| PUT | /api/v1/devices/:id | 更新设备 |
| DELETE | /api/v1/devices/:id | 删除设备 |

#### 5.1.6 采集点接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/points | 获取采集点列表 |
| POST | /api/v1/points | 创建采集点 |
| GET | /api/v1/points/:id | 获取采集点详情 |
| PUT | /api/v1/points/:id | 更新采集点 |
| DELETE | /api/v1/points/:id | 删除采集点 |

#### 5.1.7 告警接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/alarms | 获取告警列表 |
| GET | /api/v1/alarms/:id | 获取告警详情 |
| POST | /api/v1/alarms/:id/ack | 确认告警 |
| POST | /api/v1/alarms/:id/clear | 清除告警 |
| GET | /api/v1/alarms/statistics | 获取告警统计 |

#### 5.1.8 告警规则接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/alarm-rules | 获取规则列表 |
| POST | /api/v1/alarm-rules | 创建规则 |
| GET | /api/v1/alarm-rules/:id | 获取规则详情 |
| PUT | /api/v1/alarm-rules/:id | 更新规则 |
| DELETE | /api/v1/alarm-rules/:id | 删除规则 |

#### 5.1.9 实时数据接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/data/realtime | 获取实时数据 |
| GET | /api/v1/data/history | 获取历史数据 |
| GET | /api/v1/data/statistics | 获取统计数据 |
| POST | /api/v1/data/control | 遥控操作 |
| POST | /api/v1/data/setpoint | 参数设置 |

#### 5.1.10 AI 接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/ai/qa | AI 智能问答 |
| POST | /api/v1/ai/config-suggest | AI 配置建议 |
| POST | /api/v1/ai/predict | AI 推理预测 |
| POST | /api/v1/ai/batch-predict | AI 批量推理 |
| GET | /api/v1/ai/models | 获取模型列表 |
| GET | /api/v1/ai/models/:id | 获取模型详情 |
| GET | /api/v1/ai/batch-jobs/:id | 获取批量任务状态 |

#### 5.1.11 配置管理接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/configs | 获取配置列表 |
| POST | /api/v1/configs | 创建配置 |
| GET | /api/v1/configs/:id | 获取配置详情 |
| PUT | /api/v1/configs/:id | 更新配置 |
| DELETE | /api/v1/configs/:id | 删除配置 |
| GET | /api/v1/configs/:id/versions | 获取配置版本 |
| POST | /api/v1/configs/:id/release | 发布配置 |
| GET | /api/v1/configs/:id/audits | 获取审计日志 |

#### 5.1.12 导出与报表接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/export/excel | 导出 Excel |
| POST | /api/v1/export/csv | 导出 CSV |
| GET | /api/v1/reports | 获取报表列表 |
| POST | /api/v1/reports/generate | 生成报表 |

#### 5.1.13 边缘节点接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/edges | 获取边缘节点列表 |
| POST | /api/v1/edges | 注册边缘节点 |
| GET | /api/v1/edges/:id | 获取节点详情 |
| PUT | /api/v1/edges/:id | 更新节点 |
| DELETE | /api/v1/edges/:id | 删除节点 |

#### 5.1.14 资产管理接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/assets | 获取资产列表 |
| POST | /api/v1/assets | 创建资产 |
| GET | /api/v1/assets/:id | 获取资产详情 |
| PUT | /api/v1/assets/:id | 更新资产 |
| DELETE | /api/v1/assets/:id | 删除资产 |
| GET | /api/v1/assets/:id/depreciation | 获取折旧信息 |
| GET | /api/v1/assets/:id/maintenance | 获取维护记录 |
| GET | /api/v1/assets/:id/documents | 获取资产文档 |

### 5.2 内部接口

#### 5.2.1 服务间通信

| 通信方式 | 场景 | 技术选型 |
|----------|------|----------|
| 同步调用 | api-server ↔ alarm, compute, ai-service | gRPC |
| 异步消息 | collector → alarm, compute → storage | Kafka |
| 缓存共享 | 所有服务 ↔ Redis | Redis |

#### 5.2.2 采集器接口

```go
type Collector interface {
    Connect(ctx context.Context) error
    Disconnect() error
    Collect(ctx context.Context) ([]DataPoint, error)
    Control(ctx context.Context, pointID string, value interface{}) error
}
```

#### 5.2.3 时序存储接口

```go
type TimeSeriesStorage interface {
    Write(ctx context.Context, data []*TimeSeriesData) error
    Query(ctx context.Context, query *TimeSeriesQuery) ([]*TimeSeriesData, error)
    BatchWrite(ctx context.Context, batch []*TimeSeriesData) error
}
```

#### 5.2.4 AI 推理接口

```go
type InferenceService interface {
    Predict(ctx context.Context, req *PredictRequest) (*PredictResponse, error)
    BatchPredict(ctx context.Context, req *BatchPredictRequest) (*BatchPredictResponse, error)
    GetBatchJobStatus(ctx context.Context, jobID string) (*BatchJobStatus, error)
    ListModels(ctx context.Context) ([]*ModelInfo, error)
    GetModel(ctx context.Context, modelID string) (*ModelInfo, error)
}
```

---

## 6. 数据需求分析

### 6.1 核心数据模型

#### 6.1.1 组织结构模型

```
Region（区域）→ SubRegion（子区域）→ Station（厂站）→ Device（设备）→ Point（采集点）
```

| 实体 | 主要字段 | 关系 |
|------|----------|------|
| Region | code, name, parent_id, level | 1:N → SubRegion |
| SubRegion | code, name, region_id | N:1 → Region; 1:N → Station |
| Station | code, name, type, capacity, longitude, latitude, status | N:1 → SubRegion; 1:N → Device |
| Device | code, name, type, protocol, ip_address, status | N:1 → Station; 1:N → Point |
| Point | code, name, type, unit, address, alarm_high, alarm_low | N:1 → Device |

#### 6.1.2 告警数据模型

| 实体 | 主要字段 | 关系 |
|------|----------|------|
| Alarm | point_id, device_id, station_id, type, level, title, message, value, threshold, status | N:1 → Point, Device, Station |
| AlarmRule | name, type, level, condition, threshold, duration, notify_channels, status | 关联 Point/Device/Station |
| NotificationConfig | type, name, config(JSON), enabled | 独立配置 |

#### 6.1.3 AI 数据模型

| 实体 | 主要字段 | 关系 |
|------|----------|------|
| FaultDetectionResult | device_id, fault_type, severity, confidence, status, root_cause, recommendation | N:1 → Device |
| ForecastResult | station_id, forecast_type, target_time, predicted_power, accuracy, confidence | N:1 → Station |
| QASession | user_id, title, status | 1:N → QAMessage |
| QAMessage | session_id, role, content | N:1 → QASession |
| ModelVersion | model_id, version, path, status | 独立管理 |
| EdgeNode | name, station_id, ip_address, status, cpu_usage, memory_usage | N:1 → Station |

#### 6.1.4 配置数据模型

| 实体 | 主要字段 | 关系 |
|------|----------|------|
| ConfigItem | key, value, value_type, env, namespace, group, encrypted, enabled | 1:N → ConfigVersion, ConfigRelease, ConfigAudit |
| SystemConfig | category, key, value, value_type, description | 独立管理 |

#### 6.1.5 资产与供应链模型

| 实体 | 主要字段 | 关系 |
|------|----------|------|
| Asset | code, name, asset_type, cost, current_value, depreciation_method | 1:N → DepreciationRecord, MaintenanceRecord, Document |
| Inventory | code, name, type, quantity, min_quantity, unit_price | 独立管理 |
| PurchaseOrder | code, supplier_id, order_date, status, total_amount | N:1 → Supplier; 1:N → OrderItem |
| CostEntry | code, date, cost_category_id, amount, approval_status | N:1 → CostCategory |

### 6.2 数据存储策略

| 数据类型 | 存储系统 | 保留策略 | 说明 |
|----------|----------|----------|------|
| 配置数据 | PostgreSQL | 永久 | 区域、设备、采集点、用户、权限等配置 |
| 实时数据 | Redis | 1 天 | 最新采集值、设备在线状态 |
| 历史数据 | Doris/ClickHouse | 1 年+ | 时序数据，支持压缩存储 |
| 告警数据 | PostgreSQL + Redis | 1 年 | 实时告警 Redis，历史告警 PostgreSQL |
| 统计数据 | Doris/ClickHouse | 3 年 | 统计报表数据 |
| AI 结果 | PostgreSQL | 1 年 | 预测结果、故障检测结果 |
| 操作日志 | PostgreSQL | 90 天 | 用户操作审计日志 |
| 清理日志 | PostgreSQL | 30 天 | 数据清理任务日志 |

### 6.3 数据分片策略

| 策略 | 适用场景 | 描述 |
|------|----------|------|
| 时间分区 | 时序数据 | 按天分区，便于数据清理和查询 |
| 厂站分片 | 时序数据 | 同一厂站数据存储在同一分片 |
| 区域分库 | 配置数据 | 大型区域独立数据库 |
| 读写分离 | 所有数据 | 主库写入，从库读取 |

### 6.4 数据迁移需求

| 迁移脚本 | 描述 |
|----------|------|
| 001_init_schema | 初始化数据库表结构 |
| 002_add_alarm_rules | 添加告警规则表 |
| 003_add_qa_tables | 添加问答相关表 |
| 004_add_system_configs | 添加系统配置表 |
| 005_add_performance_indexes | 添加性能索引 |
| 006_add_table_partitions | 添加表分区 |
| 007_add_operation_logs | 添加操作日志和权限表 |
| 008_add_energy_efficiency | 添加能效分析表 |
| 009_add_carbon_emission | 添加碳排放管理表 |
| 010_add_ai_module_tables | 添加 AI 模块表 |
| 011_add_alarm_rules_table | 添加告警规则表（补充） |

### 6.5 数据流设计

```
设备 → 采集协议（IEC104/Modbus/61850）→ Collector → Kafka → Alarm/Compute
                                                         ↓
                                              Redis（实时数据）
                                                         ↓
                                              Doris/CK（历史数据，批量写入）
```

---

## 7. 约束与假设

### 7.1 技术约束

| 约束项 | 描述 |
|--------|------|
| 编程语言 | Go 1.24+ |
| 前端框架 | Vue 3 + Element Plus + Tailwind CSS |
| Web 框架 | Gin 1.9+ |
| ORM | GORM 1.25+ |
| 依赖注入 | Google Wire 0.7+ |
| 关系数据库 | PostgreSQL 15+ |
| 缓存 | Redis 7+ |
| 消息队列 | Kafka 3.7+ |
| 时序数据库 | Apache Doris 2.0+ 或 ClickHouse |
| 容器化 | Docker + Kubernetes |
| CI/CD | GitHub Actions |

### 7.2 业务约束

| 约束项 | 描述 |
|--------|------|
| 架构一致性 | 保持与现有系统架构的一致性，遵循 DDD 分层 |
| 向后兼容 | API 变更需保持向后兼容 |
| 数据安全 | 敏感数据加密存储，操作留痕 |
| 合规要求 | 满足电力行业安全规范 |

### 7.3 假设条件

| 假设项 | 描述 |
|--------|------|
| 数据库表结构 | 已按设计文档创建或可通过迁移脚本创建 |
| 前端页面框架 | 已完整搭建，只需补充数据绑定和交互逻辑 |
| 开发环境 | 已配置完成，工具链可用 |
| 网络环境 | 电站现场网络稳定，延迟可接受 |
| 设备协议 | 设备支持标准 Modbus/IEC104/IEC61850 协议 |
| AI 模型 | 功率预测和故障诊断模型已训练完成 |
| 第三方服务 | 邮件/短信/微信通知服务可用 |

---

## 8. 风险分析

### 8.1 技术风险

| 风险编号 | 风险描述 | 可能性 | 影响 | 风险等级 | 缓解策略 |
|----------|----------|--------|------|----------|----------|
| RSK-001 | DryRun 模式下生命周期清理无限循环 | 高 | 高 | 🔴 严重 | 修复 `doCleanup` 方法，DryRun 模式下在查询批次为空时跳出循环；增加最大迭代次数保护 |
| RSK-002 | 23 个包缺少单元测试，代码质量不可控 | 高 | 中 | 🟠 高 | 制定测试补充计划，优先覆盖核心业务逻辑包 |
| RSK-003 | 大数据量下时序查询性能下降 | 中 | 高 | 🟠 高 | 实施查询缓存、分区优化、查询限流；建立查询性能基准测试 |
| RSK-004 | Kafka 消息积压导致告警延迟 | 中 | 高 | 🟠 高 | 监控消费延迟；设置告警阈值；增加消费者实例 |
| RSK-005 | AI 模型推理延迟过高 | 中 | 中 | 🟡 中 | 实施推理缓存；优化模型；支持批量推理 |
| RSK-006 | Redis 缓存雪崩/穿透 | 中 | 高 | 🟠 高 | 实施缓存预热、互斥锁、布隆过滤器；降级到内存缓存 |
| RSK-007 | 多协议采集兼容性问题 | 中 | 中 | 🟡 中 | 充分的协议兼容性测试；建立设备接入认证流程 |

### 8.2 业务风险

| 风险编号 | 风险描述 | 可能性 | 影响 | 风险等级 | 缓解策略 |
|----------|----------|--------|------|----------|----------|
| RSK-008 | 前后端数据未完全打通，存在模拟数据残留 | 高 | 高 | 🔴 严重 | 逐模块验证前后端集成；建立 API 集成测试 |
| RSK-009 | 告警规则误报/漏报 | 中 | 高 | 🟠 高 | 支持规则试运行模式；提供告警统计分析；持续优化规则 |
| RSK-010 | 碳排放模块未启用（Handler 已注释） | 低 | 低 | 🟢 低 | 后续迭代中启用并完善 |

### 8.3 项目风险

| 风险编号 | 风险描述 | 可能性 | 影响 | 风险等级 | 缓解策略 |
|----------|----------|--------|------|----------|----------|
| RSK-011 | 模块间耦合度高，修改影响范围大 | 中 | 中 | 🟡 中 | 严格遵循 DDD 分层；使用接口抽象；建立变更影响分析流程 |
| RSK-012 | 依赖第三方服务（Nacos、邮件、短信）不可用 | 低 | 中 | 🟡 中 | 设计降级方案；关键服务支持本地配置回退 |
| RSK-013 | 数据库迁移脚本冲突 | 中 | 中 | 🟡 中 | 统一迁移脚本管理；CI 中验证迁移脚本 |

---

## 9. 需求优先级矩阵

采用 MoSCoW 方法对需求进行优先级排序：

### 9.1 Must Have（必须有）

| 编号 | 需求名称 | 模块 | 理由 |
|------|----------|------|------|
| FR-ALM-001 | 实时告警检测 | 告警 | 核心业务功能，直接影响运维安全 |
| FR-ALM-002 | 告警规则管理 | 告警 | 告警系统的基础配置能力 |
| FR-ALM-003 | 告警通知 | 告警 | 告警闭环的必要环节 |
| FR-MON-001 | 多协议数据采集 | 监控 | 系统数据来源的基础 |
| FR-MON-002 | WebSocket 实时推送 | 监控 | 实时监控的核心能力 |
| FR-MON-003 | 数据处理管道 | 监控 | 数据质量保障的基础 |
| FR-AI-001 | 故障诊断与预警 | AI | 智能化运维的核心价值 |
| FR-AI-002 | 功率预测分析 | AI | 新能源电站的核心需求 |
| FR-AI-004 | AI 模型推理服务 | AI | AI 功能的运行基础 |
| FR-STR-001 | 时序数据存储与查询 | 存储 | 数据持久化的基础 |
| FR-STR-002 | 数据生命周期管理 | 存储 | 数据治理的必要能力 |
| FR-STAT-001 | 多维度统计计算 | 统计 | 运营分析的基础 |
| FR-STAT-002 | 统计报表生成 | 统计 | 管理决策的依据 |
| FR-EXP-001 | 数据导出 | 导出 | 数据利用的必要手段 |
| FR-CMP-001 | 公式计算与规则引擎 | 计算 | 数据二次计算的核心 |
| FR-CFG-001 | 配置中心管理 | 配置 | 系统灵活配置的基础 |
| FR-DEV-001 | 设备全生命周期管理 | 设备 | 资产管理的基础 |
| FR-AUTH-001 | 用户认证与权限管理 | 认证 | 系统安全的基础 |

### 9.2 Should Have（应该有）

| 编号 | 需求名称 | 模块 | 理由 |
|------|----------|------|------|
| FR-AI-003 | AI 智能问答 | AI | 提升运维效率，非紧急 |
| FR-MON-004 | 系统可观测性 | 监控 | 运维保障能力 |
| FR-STR-003 | 数据分片与索引管理 | 存储 | 性能优化能力 |
| FR-BD-001 | 大数据处理与分析 | 大数据 | 高级分析能力 |
| FR-DEV-002 | 资产管理 | 资产 | 资产精细化管理 |
| FR-DEV-003 | 工单管理 | 工单 | 运维流程闭环 |
| FR-EE-001 | 能效分析 | 能效 | 运营优化参考 |
| FR-HARNESS-001 | Harness 测试框架 | 测试 | 质量保障工具 |

### 9.3 Could Have（可以有）

| 编号 | 需求名称 | 模块 | 理由 |
|------|----------|------|------|
| FR-AI-005 | AI 辅助操作与配置 | AI | 锦上添花功能 |
| FR-AI-006 | 边缘智能计算 | AI | 特定场景需求 |
| FR-SCM-001 | 供应链与成本管理 | 供应链 | 业务扩展功能 |
| FR-CE-001 | 碳排放追踪与管理 | 碳排放 | 合规需求，优先级可调整 |

### 9.4 Won't Have（本次不会有）

| 需求名称 | 描述 | 原因 |
|----------|------|------|
| 移动端响应式优化 | 移动端 APP 适配 | 后续迭代处理 |
| 第三方系统集成 | ERP、CRM 等系统集成 | 当前聚焦核心功能 |
| 高级 AI 功能 | 自动训练、AutoML 等 | 需要更多数据积累 |
| 多租户支持 | SaaS 化多租户 | 当前为私有化部署 |

### 9.5 优先级矩阵总览

| 优先级 | 数量 | 占比 |
|--------|------|------|
| Must Have | 18 | 56.3% |
| Should Have | 8 | 25.0% |
| Could Have | 4 | 12.5% |
| Won't Have | 4 | 6.2% |

---

## 附录 A：术语表

| 术语 | 说明 |
|------|------|
| NEM | New Energy Monitoring，新能源监控系统 |
| DDD | Domain-Driven Design，领域驱动设计 |
| RBAC | Role-Based Access Control，基于角色的访问控制 |
| RUL | Remaining Useful Life，剩余使用寿命 |
| SOC | State of Charge，荷电状态 |
| PR | Performance Ratio，性能比 |
| DSL | Domain-Specific Language，领域特定语言 |
| RAG | Retrieval-Augmented Generation，检索增强生成 |
| OLAP | Online Analytical Processing，在线分析处理 |
| 遥信 | 开关量状态采集 |
| 遥测 | 模拟量数值采集 |
| 遥控 | 远程控制操作 |
| 电度 | 电量累计值 |

## 附录 B：参考文档

| 文档 | 路径 |
|------|------|
| 系统架构设计文档 | [architecture.md](architecture.md) |
| 产品需求文档 | [.trae/specs/nem-core-features/spec.md](../.trae/specs/nem-core-features/spec.md) |
| 数据库设计文档 | [database-design.md](database-design.md) |
| API 接口文档 | [api-interface.md](api-interface.md) |
| 测试计划 | [test-plan.md](test-plan.md) |
| 部署指南 | [deployment-guide.md](deployment-guide.md) |
| 安全审计报告 | [security-audit-report.md](security-audit-report.md) |
