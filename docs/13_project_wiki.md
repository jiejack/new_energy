# NEM 项目知识 Wiki

| 字段 | 值 |
|------|------|
| 文档版本 | v1.0.0 |
| 文档名称 | NEM 项目知识 Wiki |
| 文档编号 | 13 |
| 适用项目 | 新能源监控系统（NEM） |
| 文档类型 | 项目知识库 / Wiki |
| 文档负责人 | NEM 技术负责人 |
| 关联文档 | 11_long_term_roadmap.md、12_tech_debt_register.md、14_document_version_control.md、15_skills_configuration.md |
| 最后更新 | 2026-06-17 |

---

## 目录

1. [项目概述](#1-项目概述)
2. [技术架构](#2-技术架构)
3. [功能模块清单](#3-功能模块清单)
4. [微服务架构](#4-微服务架构)
5. [数据架构](#5-数据架构)
6. [部署架构](#6-部署架构)
7. [开发规范](#7-开发规范)
8. [项目状态](#8-项目状态)
9. [关键设计决策](#9-关键设计决策)
10. [知识库索引](#10-知识库索引)

---

## 1. 项目概述

### 1.1 项目定位

新能源监控系统（New Energy Monitoring，简称 NEM）是面向光伏、风电、储能等新能源场站的工业物联网智能监控平台。系统提供设备接入、实时数据采集、时序数据存储、智能告警、AI 预测分析、报表导出等全栈能力，帮助运营商实现新能源资产的数字化、智能化运营。

### 1.2 核心目标

- **实时监控**：秒级采集设备运行数据，毫秒级异常告警响应。
- **智能分析**：基于 AI 模型实现功率预测、故障预警、健康度评估。
- **数据驱动**：时序数据高效存储与查询，支撑运营决策。
- **开放扩展**：支持多种工业协议接入，微服务架构便于水平扩展。

### 1.3 技术栈总览

| 层次 | 技术选型 |
|------|---------|
| 后端语言 | Go 1.24 |
| Web 框架 | Gin |
| ORM | GORM |
| 关系型数据库 | PostgreSQL |
| 时序数据库 | TimescaleDB / ClickHouse / Doris |
| 缓存 | Redis |
| 消息队列 | Kafka |
| 配置中心 / 服务注册 | Nacos |
| 前端框架 | Vue 3 + TypeScript + Vite |
| 前端 UI | Element Plus |
| 前端状态管理 | Pinia |
| 工业协议 | IEC 104、Modbus、IEC 61850 |

---

## 2. 技术架构

### 2.1 分层架构

NEM 后端采用领域驱动设计（DDD）分层架构，各层职责清晰、依赖方向单向（外层依赖内层）：

```
┌─────────────────────────────────────────────────┐
│  API 层 (internal/api/handler)                  │  HTTP 路由、请求校验、响应序列化
├─────────────────────────────────────────────────┤
│  Application 层 (internal/application/service)  │  业务用例编排、事务管理、DTO 转换
├─────────────────────────────────────────────────┤
│  Domain 层 (internal/domain)                    │  领域实体、仓储接口、领域服务
├─────────────────────────────────────────────────┤
│  Infrastructure 层 (internal/infrastructure)    │  数据库、缓存、消息队列、外部服务实现
└─────────────────────────────────────────────────┘
```

| 层次 | 目录 | 职责 |
|------|------|------|
| API 层 | `internal/api/handler` | 接收 HTTP 请求，参数绑定与校验，调用 Application 层，返回统一响应格式 |
| Application 层 | `internal/application/service` | 业务逻辑编排，协调领域对象，管理事务边界，调用仓储完成持久化 |
| Domain 层 | `internal/domain` | 领域实体（`entity/`）、仓储接口（`repository/`）、缓存接口（`cache/`）、日志接口（`logger/`） |
| Infrastructure 层 | `internal/infrastructure` | 仓储实现（`persistence/`）、Redis 缓存实现（`cache/`）、配置加载（`config/`）、Kafka（`mq/`）、日志实现（`logger/`） |

### 2.2 依赖注入

项目使用 `wire` 进行编译时依赖注入，依赖关系在 `cmd/api-server/wire.go` 中声明，生成代码位于 `wire_gen.go`。详见 `docs/wire-usage.md`。

### 2.3 公共包（pkg/）

`pkg/` 目录存放可被外部服务复用的公共能力包，与 `internal/`（仅本服务可用）形成隔离：

| 公共包 | 路径 | 说明 |
|-------|------|------|
| ai | `pkg/ai/` | AI 能力：预测（光伏/风电/储能）、故障检测、知识库（RAG）、推理服务、QA 对话 |
| alarm | `pkg/alarm/` | 告警引擎：规则 DSL、聚合、去重、状态机、通知器（邮件/短信） |
| auth | `pkg/auth/` | 认证授权：JWT 签发与校验、密码哈希 |
| bigdata | `pkg/bigdata/` | 大数据分析服务 |
| cache | `pkg/cache/` | 通用缓存抽象与实现 |
| collector | `pkg/collector/` | 数据采集：连接池、缓冲、调度 |
| compute | `pkg/compute/` | 实时计算：规则引擎、点位计算、触发器 |
| config | `pkg/config/` | 配置加载与热更新监听 |
| errors | `pkg/errors/` | 统一错误码与错误响应 |
| export | `pkg/export/` | 数据导出：CSV、Excel |
| feedback | `pkg/feedback/` | 反馈收集 |
| harness | `pkg/harness/` | 测试约束与验证工具 |
| monitoring | `pkg/monitoring/` | 可观测性：指标、健康检查、链路追踪、告警、Grafana 面板 |
| nacos | `pkg/nacos/` | Nacos 配置客户端与服务注册 |
| processor | `pkg/processor/` | 数据处理管道：过滤、校验、质量检查、缩放、变化检测 |
| storage | `pkg/storage/` | 存储引擎：时序（ClickHouse/Doris）、压缩、索引、分片、分区、查询、生命周期 |
| websocket | `pkg/websocket/` | WebSocket 实时推送 |

---

## 3. 功能模块清单

### 3.1 后端业务模块

| 模块 | Handler | Service | 说明 |
|------|---------|---------|------|
| 告警管理 | alarm_handler / alarm_rule_handler | alarm_service / alarm_rule_service | 告警规则配置、告警查询、AI 告警 |
| 设备管理 | device_handler / edge_handler | device_service / edge_service | 设备台账、边缘节点管理 |
| 故障管理 | fault_handler | fault_service | 故障检测、故障工单桥接 |
| 预测分析 | forecast_handler | forecast_service | 光伏/风电/储能功率预测 |
| AI 模型 | model_handler | model_service | 模型版本管理 |
| 问答系统 | qa_handler | qa_service | AI 智能问答 |
| 用户管理 | user_handler / auth_handler | user_service / auth_service | 用户、认证、权限 |
| 电站管理 | station_handler | station_service | 电站信息、区域管理 |
| 配置管理 | config_handler | config_service | 系统配置 |
| 通知配置 | notification_config_handler | notification_config_service | 通知渠道配置 |
| 操作日志 | operation_log_handler | operation_log_service | 审计日志 |
| 报表 | report_handler / export_handler | report_service / export_service | 报表生成与导出 |
| 资产管理 | asset_handler / asset_maintenance_handler / asset_depreciation_handler / asset_document_handler | asset_service 等 | 资产全生命周期 |
| 成本管理 | cost_entry_handler / cost_category_handler / cost_allocation_handler / cost_report_handler | cost_*_service | 成本录入、分摊、报表 |
| 库存采购 | inventory_handler / purchase_order_handler / receipt_handler | inventory_service 等 | 库存、采购、收货 |
| 工单 | work_order_handler | work_order_service | 工单管理 |
| 能效 | energy_efficiency_handler | energy_efficiency_service | 能效分析 |
| 碳排放 | carbon_emission_handler | carbon_emission_service | 碳排放核算 |

### 3.2 前端模块

前端位于 `web/src/`，采用 Vue 3 + TypeScript + Vite + Element Plus + Pinia 技术栈，主要模块包括：告警、设备、数据监控、配置、日志、权限、报表、电站、用户等，对应 `web/src/api/` 下的接口定义与 `web/src/stores/` 下的状态管理。

---

## 4. 微服务架构

### 4.1 微服务清单

NEM 由 7 个微服务组成，各自独立部署、独立扩展：

| 微服务 | 入口 | 职责 |
|-------|------|------|
| api-server | `cmd/api-server/main.go` | 主 API 网关，提供 RESTful 接口，处理所有前端与第三方请求 |
| collector | `cmd/collector/main.go` | 数据采集服务，对接工业协议（IEC 104/Modbus/IEC 61850），采集设备实时数据 |
| compute | `cmd/compute/main.go` | 实时计算服务，执行规则引擎、点位计算、数据加工 |
| alarm | `cmd/alarm/main.go` | 告警服务，规则匹配、告警聚合去重、通知分发 |
| scheduler | `cmd/scheduler/main.go` | 调度服务，定时任务编排与触发 |
| ai-service | `cmd/ai-service/main.go` | AI 服务，功率预测、故障预警、智能问答、模型推理 |
| migrate | `cmd/migrate/main.go` | 数据迁移服务，执行数据库 Schema 迁移脚本 |

### 4.2 服务间通信

- **同步通信**：api-server 通过 HTTP/REST 对外提供服务；内部服务间必要时通过 HTTP 调用。
- **异步通信**：基于 Kafka 消息队列实现服务间解耦，典型场景为 collector 采集数据 → Kafka → compute 计算 → Kafka → alarm 告警。
- **服务发现**：通过 Nacos 完成服务注册与发现，支持动态扩缩容。

---

## 5. 数据架构

### 5.1 数据库选型

| 数据库 | 用途 | 说明 |
|-------|------|------|
| PostgreSQL | 业务关系型数据 | 用户、设备、告警规则、配置、资产、成本等结构化数据 |
| TimescaleDB | 时序数据（PostgreSQL 扩展） | 基于 PostgreSQL 的时序扩展，支持超表（hypertable）、连续聚合 |
| ClickHouse | 大规模时序分析 | 列式存储，适合海量时序数据的高效聚合查询 |
| Doris | 实时分析 | 支持实时写入与亚秒级查询的 OLAP 引擎 |
| Redis | 缓存与会话 | 热点数据缓存、分布式锁、会话存储 |

### 5.2 数据流

```
设备 → collector（IEC 104/Modbus/IEC 61850）
            ↓
        Kafka 消息队列
            ↓
    compute（规则计算/数据加工）
       ↙        ↘
TimescaleDB    Kafka
（时序存储）      ↓
            alarm（告警匹配）
                ↓
          PostgreSQL（告警记录）
                ↓
          api-server → 前端展示 / WebSocket 实时推送
```

### 5.3 数据库迁移

迁移脚本位于 `scripts/migrations/`，按编号顺序执行（001-011），由 `migrate` 服务驱动。详见 `docs/database-migration-guide.md`。

---

## 6. 部署架构

### 6.1 容器化部署

- **Docker**：各服务均有独立 Dockerfile，位于 `ops/docker/`（Dockerfile.backend、Dockerfile.collector、Dockerfile.compute、Dockerfile.alarm、Dockerfile.scheduler、Dockerfile.ai-service、Dockerfile.frontend）。
- **Docker Compose**：`ops/docker/docker-compose.yml`（基础编排）、`docker-compose.full.yml`（全栈编排）。

### 6.2 Kubernetes 部署

K8s 部署清单位于 `ops/k8s/`，按序号组织：

| 序号 | 文件 | 说明 |
|------|------|------|
| 01 | namespace.yaml | 命名空间 |
| 02 | configmap.yaml | 配置 |
| 03 | secrets.yaml | 密钥 |
| 04 | postgres.yaml | PostgreSQL |
| 05 | redis.yaml | Redis |
| 06 | kafka.yaml | Kafka |
| 07 | api-server.yaml | API 服务 |
| 08 | microservices.yaml | 其他微服务 |
| 09 | frontend-monitoring.yaml | 前端与监控 |

### 6.3 可观测性

可观测性栈位于 `deploy/`，通过 `docker-compose.observability.yml` 一键部署：

| 组件 | 用途 |
|------|------|
| Prometheus | 指标采集（含告警规则 `rules/alert_rules.yml`、`rules/harness_rules.yml`） |
| Grafana | 可视化面板（API 服务、Go 运行时、PostgreSQL、质量面板） |
| Jaeger | 分布式链路追踪 |
| Promtail | 日志采集 |
| Alertmanager | 告警通知路由 |

---

## 7. 开发规范

### 7.1 编码规范

- **后端**：遵循 `.golangci.yml` 静态检查规则；分层架构严格单向依赖；错误使用 `pkg/errors` 统一码。
- **前端**：遵循 `.eslintrc.cjs` + `.prettierrc`；TypeScript 严格模式；组件采用 `<script setup>` 组合式 API。
- **提交信息**：遵循 Conventional Commits 规范，已配置 commitlint（`.commitlintrc.cjs`）与 git hooks（`scripts/git-hooks/`）。

### 7.2 分支管理

- 开发分支：`dev`，所有提交推送至此。
- 主分支：需经 Code Review 与 CI 流水线通过后方可合并，禁止直接推送。

### 7.3 文档先行

遵循 karpathy-guidelines 准则一，编码前必须先产出设计文档。文档统一存放 `docs/`，管理规则详见 `docs/README.md`。

---

## 8. 项目状态

### 8.1 阶段进展

| 阶段 | 状态 | 说明 |
|------|------|------|
| Phase 1 | ✅ 已完成 | 完成 6 个关键 Bug 修复，核心链路打通 |
| Phase 2 | 🔄 进行中 | 功能完善与架构优化 |

### 8.2 技术债务

技术债务登记于 `docs/12_tech_debt_register.md`，当前已修复 6 项，待偿还 30+ 项，TODO 标记 36 处。

### 8.3 长期路线图

长期规划（Phase 1 - Phase 5，6-12 个月）详见 `docs/11_long_term_roadmap.md`。

---

## 9. 关键设计决策

### 9.1 时序数据多引擎策略

NEM 同时支持 TimescaleDB、ClickHouse、Doris 三种时序存储，通过 `pkg/storage/timeseries/factory.go` 工厂模式按场景选择：
- **TimescaleDB**：与业务库同栈，适合中小规模时序场景。
- **ClickHouse**：列式存储，适合海量历史数据聚合分析。
- **Doris**：实时分析，适合亚秒级查询需求。

### 9.2 告警引擎 DSL 化

告警规则采用 DSL（`pkg/alarm/rule/dsl.go`）描述，支持灵活的条件表达式与版本管理，配合聚合器（aggregator）、去重器（dedup）、状态机（state_machine）实现工业级告警处理。

### 9.3 AI 能力分层

AI 能力按 `pkg/ai/` 下子包分层：forecast（预测）、fault（故障）、inference（推理）、knowledge（RAG 知识库）、qa（问答）、edge（边缘）、operation（操作执行）、features（特征工程），各子包独立演进。

### 9.4 工业协议抽象

通过 `pkg/collector/` 抽象采集接口，屏蔽 IEC 104、Modbus、IEC 61850 协议差异，统一接入数据管道。

---

## 10. 知识库索引

### 10.1 编号文档（01-15）

| 编号 | 文档 | 说明 |
|-----|------|------|
| 01 | `01_requirements_analysis.md` | 需求分析 |
| 02 | `02_requirements_review.md` | 需求评审 |
| 03 | `03_implementation_plan.md` | 实施计划 |
| 04 | `04_test_plan.md` | 测试计划 |
| 05 | `05_code_review.md` | 代码审查 |
| 06 | `06_iteration_plan.md` | 迭代计划 |
| 07 | `07_task_list.md` | 任务清单 |
| 08 | `08_development_plan_phase2.md` | Phase 2 开发计划 |
| 09 | `09_comprehensive_test_plan.md` | 综合测试计划 |
| 10 | `10_market_requirements_review.md` | 市场需求评审 |
| 11 | `11_long_term_roadmap.md` | 长期路线图 |
| 12 | `12_tech_debt_register.md` | 技术债务登记 |
| 13 | `13_project_wiki.md` | 项目知识 Wiki（本文档） |
| 14 | `14_document_version_control.md` | 文档版本控制机制 |
| 15 | `15_skills_configuration.md` | 开发技能与工具配置 |

### 10.2 主题文档

| 文档 | 说明 |
|------|------|
| `architecture.md` / `system-architecture.md` / `system-architecture-v2.md` | 系统架构设计 |
| `database-design.md` | 数据库设计 |
| `api-documentation.md` / `api-interface.md` / `api-reference.md` | API 接口文档 |
| `deployment-guide.md` / `DEPLOYMENT.md` | 部署指南 |
| `developer-guide.md` | 开发者指南 |
| `operations-guide.md` / `operations-manual.md` | 运维手册 |
| `security-audit-report.md` | 安全审计报告 |
| `performance-benchmark.md` | 性能基准 |
| `nacos-integration-design.md` | Nacos 集成设计 |
| `config-center-design.md` | 配置中心设计 |
| `fault-diagnosis-system.md` | 故障诊断系统 |
| `user-center-design.md` | 用户中心设计 |

### 10.3 技能文档

| 文档 | 说明 |
|------|------|
| `skills/golang-development.md` | Go 后端开发技能 |
| `skills/vue3-development.md` | Vue 3 前端开发技能 |
| `skills/testing-skills.md` | 测试技能 |
| `skills/code-review-skills.md` | 代码审查技能 |
| `skills/project-skills-500.md` | 500 轮迭代项目技能总结 |

### 10.4 Wiki 知识库

`docs/wiki/` 下维护使用指南类文档：Home、Quick-Start、Installation-Guide、Project-Structure、Configuration、API-Documentation、Feature-Guide、FAQ、SYNC-GUIDE、GitHub-Secrets-Setup。

---

## 变更记录表

| 版本 | 日期 | 变更内容 | 变更人 |
|-----|------|---------|-------|
| v1.0.0 | 2026-06-17 | 初始创建：项目概述、技术架构、功能模块、微服务架构、数据架构、部署架构、开发规范、项目状态、关键设计决策、知识库索引 | AI Agent |
