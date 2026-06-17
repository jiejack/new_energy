# 开发技能与工具配置文档

| 字段 | 值 |
|------|------|
| 文档版本 | v1.0.0 |
| 文档名称 | NEM 开发技能与工具配置文档 |
| 文档编号 | 15 |
| 适用项目 | 新能源监控系统（NEM） |
| 文档类型 | 技能配置规范 |
| 文档负责人 | NEM 技术负责人 |
| 关联文档 | AGENTS.md、13_project_wiki.md、skills/golang-development.md、skills/vue3-development.md、skills/testing-skills.md、skills/code-review-skills.md |
| 最后更新 | 2026-06-17 |

---

## 目录

1. [技能配置概述](#1-技能配置概述)
2. [行为准则技能](#2-行为准则技能)
3. [Go 语言开发技能](#3-go-语言开发技能)
4. [测试技能](#4-测试技能)
5. [代码审查技能](#5-代码审查技能)
6. [前端开发技能](#6-前端开发技能)
7. [GStack 技能集配置](#7-gstack-技能集配置)
8. [Superpower 技能集配置](#8-superpower-技能集配置)
9. [Caveman 模式配置](#9-caveman-模式配置)
10. [PUA 技能配置](#10-pua-技能配置)
11. [技能使用规范](#11-技能使用规范)

---

## 1. 技能配置概述

### 1.1 配置目的

本文档统一登记与说明 NEM 项目已配置的全部开发技能与工具，为 AI 代理与人类开发者提供技能索引、配置说明与使用规范。所有技能配置以 `AGENTS.md` 为统一入口，本文档为其详细展开。

### 1.2 已配置技能清单

| 技能类别 | 技能名称 | 配置位置 | 优先级 |
|---------|---------|---------|-------|
| 行为准则 | karpathy-guidelines | `AGENTS.md` 第一章 | 最高（必须最先配置） |
| 开发技能 | Go 后端开发技能 | `docs/skills/golang-development.md` | 高 |
| 开发技能 | Vue 3 前端开发技能 | `docs/skills/vue3-development.md` | 高 |
| 开发技能 | 测试技能 | `docs/skills/testing-skills.md` | 高 |
| 开发技能 | 代码审查技能 | `docs/skills/code-review-skills.md` | 高 |
| 技能集 | GStack 技能集 | `AGENTS.md` 3.3 节 | 中 |
| 技能集 | Superpower 技能集 | `AGENTS.md` 3.4 节 | 中 |
| 通信模式 | Caveman 模式 | `.trae/caveman-config.json` | 按需触发 |
| 执行模式 | PUA 技能 | `.trae/pua-config.json` | 按需触发 |

### 1.3 加载顺序

技能按以下顺序加载与生效，发生冲突时以上方优先级为准：

1. `karpathy-guidelines`（行为基准，最高优先级）
2. 项目开发技能（Go / 前端 / 测试 / 代码审查）
3. GStack 技能集 / Superpower 技能集
4. Caveman 模式 / PUA 技能（按需触发）

---

## 2. 行为准则技能

### 2.1 karpathy-guidelines 详解

本项目采用 Andrej Karpathy 四条核心编码准则（karpathy-guidelines）作为 AI 代理与开发者的行为基准。四条准则按优先级从上到下执行，冲突时以上方准则为准。

#### 准则一：先思考再编码（Think Before You Code）

- **核心思想**：编码前必须完成设计文档编写与方案评审，杜绝"边想边写"。
- **执行要求**：
  - 接到任务后先输出设计方案（目标、范围、接口契约、数据流、风险点）。
  - 涉及多模块改动时，必须先在 `docs/plans/` 下产出实施计划文档。
  - 设计方案需明确"做什么 / 不做什么"的边界，避免范围蔓延。
  - 复杂决策需对照 `docs/architecture/challenges.md` 中的既有难点方案。
- **自检清单**：
  - [ ] 是否已明确本次变更的目标与验收标准？
  - [ ] 是否已产出设计文档或实施计划？
  - [ ] 是否已评估对现有模块的影响范围？
  - [ ] 是否已识别潜在风险并给出应对方案？

#### 准则二：简单优先（Simplicity First）

- **核心思想**：在所有可行方案中，选择最简单且能解决问题的方案。
- **执行要求**：
  - 优先使用标准库与项目既有依赖，避免引入新依赖。
  - 禁止为一次性操作创建抽象层、工厂或配置体系。
  - 禁止"以防万一"式的防御性代码与错误处理。
  - 复杂逻辑优先拆分为可独立验证的小函数。
- **自检清单**：
  - [ ] 是否存在更简单的实现方式？
  - [ ] 是否引入了非必要的新依赖？
  - [ ] 是否为未来"可能"的需求预留了过度设计？
  - [ ] 代码是否可被一名新成员在 5 分钟内读懂？

#### 准则三：精准修改（Precise Modifications）

- **核心思想**：只做必要的修改，保持变更最小化，不夹带"顺手改进"。
- **执行要求**：
  - 每次变更严格限定在任务范围内，不重构未涉及的代码。
  - 不为未修改的代码添加注释、文档字符串或类型标注。
  - 不调整与本次任务无关的代码格式与导入顺序。
  - 变更需可被一次 Code Review 完整审阅。
- **自检清单**：
  - [ ] 本次 diff 是否仅包含任务所需改动？
  - [ ] 是否夹带了格式化、重命名等无关变更？
  - [ ] 是否修改了与任务无关的文件？

#### 准则四：目标驱动（Goal-Driven）

- **核心思想**：始终以最终目标为导向，避免陷入工具或过程的细节泥潭。
- **执行要求**：
  - 每一步操作都需能回答"这如何推进最终目标"。
  - 遇到阻塞时，优先评估是否可绕过非关键障碍推进主线。
  - 完成任务后需对照目标做交付验证，而非仅以"代码写完"为终点。
  - 遇到模糊需求时主动澄清，而非自行假设后大规模实施。
- **自检清单**：
  - [ ] 当前操作是否直接服务于最终目标？
  - [ ] 是否存在偏离目标的"镀金"行为？
  - [ ] 完成后是否对照目标做了交付验证？

---

## 3. Go 语言开发技能

### 3.1 技术栈

- **语言**：Go 1.24+
- **Web 框架**：Gin
- **ORM**：GORM
- **依赖注入**：Wire
- **数据库**：PostgreSQL + TimescaleDB
- **缓存**：Redis
- **消息队列**：Kafka
- **配置中心**：Nacos

### 3.2 分层架构

```
internal/
├── api/handler/         # HTTP 处理器（API 层）
├── application/service/ # 业务服务（Application 层）
├── domain/              # 领域层（entity / repository / cache / logger）
└── infrastructure/      # 基础设施层（persistence / cache / config / mq / logger）
```

### 3.3 关键规范

- 分层依赖单向：API → Application → Domain ← Infrastructure。
- 错误统一使用 `pkg/errors` 定义的错误码。
- 依赖注入通过 `wire` 编译时生成，禁止运行时反射注入。
- 详见 `docs/skills/golang-development.md`。

---

## 4. 测试技能

### 4.1 测试技术栈

| 类别 | 工具 |
|------|------|
| Go 单元测试 | testify（assert / mock / require）、httptest、gin.TestMode |
| Go 测试替身 | miniredis、SQLite 内存数据库 |
| 前端单元测试 | Vitest |
| 前端 E2E 测试 | Playwright |
| 性能测试 | k6 |

### 4.2 测试规范

- 遵循测试金字塔：单元测试为主，集成测试为辅，E2E 测试覆盖关键路径。
- 采用 AAA 模式（Arrange-Act-Assert）组织测试。
- 保证测试隔离，禁止测试间状态依赖。
- 详见 `docs/skills/testing-skills.md`。

---

## 5. 代码审查技能

### 5.1 审查流程

```
提交前自审 → 同伴审查 → 高级审查 → 合并
```

### 5.2 审查清单维度

| 维度 | 要点 |
|------|------|
| 安全性 | SQL 注入、XSS、CSRF、认证授权、敏感数据 |
| 性能 | N+1 查询、内存泄漏、并发瓶颈 |
| 可靠性 | 错误处理、资源释放、边界条件 |
| 可维护性 | 命名、注释、复杂度、重复代码 |
| 架构 | 分层依赖、接口契约、模块边界 |

### 5.3 工具配置

- 后端：`golangci-lint`（配置 `.golangci.yml`）
- 前端：`ESLint`（配置 `.eslintrc.cjs`）+ `Prettier`
- 安全：`gitleaks`（配置 `.gitleaks.toml`）
- 详见 `docs/skills/code-review-skills.md`。

---

## 6. 前端开发技能

### 6.1 技术栈

- **框架**：Vue 3 + TypeScript + Vite
- **UI 库**：Element Plus
- **状态管理**：Pinia
- **路由**：Vue Router
- **可视化**：ECharts
- **HTTP 客户端**：Axios
- **样式**：Sass + Tailwind CSS

### 6.2 关键规范

- 组件采用 `<script setup>` 组合式 API。
- TypeScript 严格模式，禁止 `any` 滥用。
- API 接口定义统一存放 `web/src/api/`，与后端保持一致。
- 详见 `docs/skills/vue3-development.md`。

---

## 7. GStack 技能集配置

GStack 技能集覆盖从规划到交付的完整研发流程，按研发阶段组织：

| 技能名称 | 阶段 | 说明 |
|---------|------|------|
| `office-hours` | 规划 | 办公时间式需求澄清与方案讨论 |
| `plan-ceo-review` | 规划 | 计划的 CEO 视角评审（业务价值） |
| `plan-eng-review` | 规划 | 计划的工程视角评审（技术可行性） |
| `plan-design-review` | 规划 | 计划的设计视角评审（架构与体验） |
| `review` | 执行 | 代码审查执行 |
| `ship` | 交付 | 发布与交付流程 |
| `cso` | 治理 | 首席安全官视角的安全审查 |
| `qa` | 验证 | 质量保证与验收测试 |

### 7.1 使用要点

- 规划阶段优先使用 `office-hours` 澄清需求，再依次进行 CEO / 工程 / 设计三视角评审。
- 执行阶段使用 `review` 进行代码审查。
- 交付前使用 `cso` 做安全审查、`qa` 做质量验收，最后 `ship` 发布。
- 快速上手参见 `docs/gstack-quick-start.md`。

---

## 8. Superpower 技能集配置

Superpower 技能集提供端到端的高质量研发工作流：

| 技能名称 | 说明 | 触发时机 |
|---------|------|---------|
| `brainstorming` | 创意构思与需求探索 | 任何创造性工作前必须使用 |
| `writing-plans` | 编写实施计划 | 多步骤任务动代码前使用 |
| `test-driven-development` | 测试驱动开发 | 实现功能或修复前使用 |
| `requesting-code-review` | 请求代码审查 | 代码完成后使用 |
| `finishing-a-development-branch` | 完成开发分支 | 合并前收尾 |
| `systematic-debugging` | 系统化调试 | 排查复杂问题 |
| `verification-before-completion` | 完成前验证 | 交付前必须执行 |

### 8.1 使用要点

- `brainstorming` 与 `writing-plans` 对应 karpathy-guidelines 准则一（先思考后编码）。
- `test-driven-development` 保证新功能配套测试。
- `verification-before-completion` 是交付前的强制验证关卡。

---

## 9. Caveman 模式配置

Caveman 模式为高 Token 效率通信模式，通过压缩响应内容减少 Token 消耗，同时保持技术准确性。配置文件位于 `.trae/caveman-config.json`。

### 9.1 强度等级

| 强度等级 | 最大响应行数 | 最大代码行数 | 仅代码 | 无解释 | 适用场景 |
|---------|------------|------------|-------|-------|---------|
| `lite` | 10 | 50 | 否 | 否 | 日常轻量交互 |
| `full` | 5 | 30 | 是 | 否 | 默认强度，常规开发 |
| `ultra` | 3 | 15 | 是 | 是 | 极端 Token 受限场景 |

### 9.2 自动触发条件

- Token 用量 > 50000
- 对话轮次 > 20
- 用户输入触发词：`/caveman`、`caveman mode`、`use caveman`、`less tokens`、`be brief`

### 9.3 退出命令

`/normal`、`/standard`、`normal mode`、`standard mode`

### 9.4 上下文保留

启用上下文保留（`context_preservation`），优先保留字段：`error`、`file`、`function`、`line`，最多保留 5 项上下文。

---

## 10. PUA 技能配置

PUA 技能为 AI 代理的高强度执行与自驱模式，配置文件位于 `.trae/pua-config.json`。

### 10.1 强度等级

| 强度等级 | 风格 | 特性 | 最大时长 | 适用场景 |
|---------|------|------|---------|---------|
| `lite` | 鼓励式 | 问题引导、方案提示 | 15 分钟 | 日常开发辅助 |
| `full` | 施压式 | 问题追问、方案要求、进度追踪 | 30 分钟 | 独立承担模块开发 |
| `ultra` | 高压式 | 穷尽搜索、绝不放弃、结果导向 | 60 分钟 | 无人值守的批量任务 |

### 10.2 自动触发条件

| 触发类型 | 阈值 | 强度 | 说明 |
|---------|------|------|------|
| 连续失败 | 2 次 | lite | 同一任务连续失败 2 次 |
| 循环检测 | 3 次 | full | 相同操作重复 3 次 |
| 被动行为 | 检测到 | full | 表现出被动行为（不搜索、不读源码、等待输入、表达限制） |
| 放弃倾向 | 检测到 | ultra | 表达"无法完成""需要手动"等放弃倾向 |

### 10.3 阶段强制启用

| 阶段 | 触发时机 | 强度 |
|------|---------|------|
| 代码审查 | PR 提交前 | lite |
| 部署前验证 | 生产部署前 | full |
| 生产 Bug 修复 | 生产 Bug 修复时 | full |
| 安全漏洞修复 | 安全漏洞修复时 | ultra |

### 10.4 安全控制

- 退出命令：`/pua:off`
- 最大会话时长：60 分钟
- 效果检查间隔：10 分钟
- 退化阈值：3 次（效果持续不达标则降级或退出）
- 启用人工介入提示

---

## 11. 技能使用规范

### 11.1 先思考后开发原则

- 遵循 karpathy-guidelines 准则一，编码前必须完成设计文档与方案评审。
- 遵循"文档先行"原则，所有功能开发先在 `docs/` 下产出对应文档。
- 复杂任务必须先调用 `brainstorming` 与 `writing-plans` 技能产出计划。
- 创造性工作（新功能、新组件、新模块）前必须使用 `brainstorming` 技能。

### 11.2 失败处理规则：3 次失败终止

为保证 AI 代理行为可控、可追溯，执行以下失败处理规则：

1. **三次失败终止原则**：任何单一操作（如编译、测试、修复某个 Bug、调用某个工具）连续尝试 **3 次** 仍未成功时，**立即终止**该操作链路，禁止无限重试。
2. **问题记录**：终止后，将问题详情记录至问题清单（`docs/error-handling/iteration-errors-500.md` 或对应迭代错误文档），记录内容包括：
   - 操作描述与目标
   - 三次尝试的具体做法与各自失败原因
   - 最终错误信息（日志、堆栈、退出码）
   - 已排除的可能性与待排查方向
3. **上报与等待**：记录完成后，将问题摘要上报，等待人工介入或新指令，禁止在未获新指令前自行尝试新方案。
4. **禁止行为**：
   - 禁止通过"重试相同命令"碰运气。
   - 禁止隐瞒失败、伪造成功结果。
   - 禁止绕过失败以"部分完成"伪装交付。

### 11.3 代码推送规则

- 所有代码提交必须推送到 `dev` 分支，禁止直接推送到主分支。
- 主分支合并需经过 Code Review（`review` 技能）与 CI 流水线通过。
- 提交信息遵循 Conventional Commits 规范（项目已配置 commitlint）。

### 11.4 代码质量规则

- 前端：无 TypeScript 错误，构建成功，遵循 ESLint + Prettier 规范。
- 后端：遵循 `.golangci.yml` 规范，所有 API 接口保持前后端一致性。
- 测试：新功能必须配套测试，遵循 `test-driven-development` 技能。

---

## 变更记录表

| 版本 | 日期 | 变更内容 | 变更人 |
|-----|------|---------|-------|
| v1.0.0 | 2026-06-17 | 初始创建：技能配置概述、karpathy-guidelines 详解、Go/测试/代码审查/前端开发技能、GStack 与 Superpower 技能集、Caveman 模式与 PUA 技能配置、技能使用规范（含 3 次失败终止规则） | AI Agent |
