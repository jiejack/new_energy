# docs/ 目录说明文档

| 字段 | 值 |
|------|------|
| 文档版本 | v2.0.0 |
| 文档名称 | NEM docs 目录说明文档 |
| 适用项目 | 新能源监控系统（NEM） |
| 文档类型 | 目录说明 / 管理规范 |
| 文档负责人 | NEM 技术负责人 |
| 关联文档 | 13_project_wiki.md、14_document_version_control.md、15_skills_configuration.md、AGENTS.md |
| 最后更新 | 2026-06-17 |

> 本文件是项目文档目录（`docs/`）的统一说明，定义文档管理规则、目录结构与分类规范。
> 所有协作者（含 AI 代理）在新增或修改文档前，必须先阅读本文件。

---

## 目录

1. [文档管理规则](#一文档管理规则)
2. [目录结构](#二目录结构)
3. [文档分类说明表](#三文档分类说明表)
4. [变更记录表](#四变更记录表)

---

## 一、文档管理规则

本项目遵循以下 5 条文档管理规则，确保文档体系清晰、可追溯、可维护：

### 规则一：统一存放

- 所有项目文档**统一存放在 `docs/` 目录下**，禁止散落在仓库根目录、代码目录或其他位置。
- 临时笔记、草稿应放入 `docs/plans/` 或对应子目录，不得在仓库根目录创建游离的 `.md` 文件。
- 代码内联注释不替代正式文档，关键设计必须沉淀为 `docs/` 下的独立文档。

### 规则二：分类清晰

- 文档按主题归入对应子目录（如 `architecture/`、`frontend/`、`error-handling/` 等）。
- 跨研发流程的核心文档使用 `01-15` 编号前缀，按流程顺序组织。
- 无法归入既有子目录的新主题，应先在本文件"目录结构"中新增子目录说明，再创建文档。

### 规则三：版本可追溯

- 每个文档子目录及核心文档需维护"变更记录表"，记录版本、日期、变更内容、变更人。
- 重要文档变更应在 `docs/CHANGELOG.md` 中同步登记。
- 废弃文档不得直接删除，应移动至 `docs/wiki/archive/` 或在文档顶部标注"已废弃"及替代文档链接。
- 版本号采用语义化版本（Major.Minor.Patch），详见 `14_document_version_control.md`。

### 规则四：链接规范

- 文档间引用使用相对路径或 `file:///workspace/docs/...` 绝对路径，确保在本地与 IDE 中可点击跳转。
- 引用编号文档时使用统一格式：`[13 项目知识 Wiki](13_project_wiki.md)`。
- 外部链接需注明来源与访问日期，避免链接腐烂导致信息丢失。

### 规则五：文档先行

- 遵循"文档先行"原则，**编码前必须先产出对应设计文档**（对应 karpathy-guidelines 准则一）。
- 新功能开发流程：需求文档 → 设计文档 → 实施计划 → 编码 → 测试文档 → 验收文档。
- 文档与代码需同步更新，代码变更导致接口或行为变化时，对应文档必须同步修订。

---

## 二、目录结构

`docs/` 目录的完整结构如下：

```
docs/
├── README.md                          # 本文件，文档目录说明
├── CHANGELOG.md                       # 文档变更日志
│
├── 01_requirements_analysis.md        # 01 需求分析
├── 02_requirements_review.md          # 02 需求评审
├── 03_implementation_plan.md          # 03 实施计划
├── 04_test_plan.md                    # 04 测试计划
├── 05_code_review.md                  # 05 代码审查
├── 06_iteration_plan.md               # 06 迭代计划
├── 07_task_list.md                    # 07 任务清单
├── 08_development_plan_phase2.md      # 08 Phase 2 开发计划
├── 09_comprehensive_test_plan.md      # 09 综合测试计划
├── 10_market_requirements_review.md   # 10 市场需求评审
├── 11_long_term_roadmap.md            # 11 长期路线图
├── 12_tech_debt_register.md           # 12 技术债务登记
├── 13_project_wiki.md                 # 13 项目知识 Wiki
├── 14_document_version_control.md     # 14 文档版本控制机制
├── 15_skills_configuration.md         # 15 开发技能与工具配置
│
├── architecture/                      # 架构设计文档
│   └── challenges.md                  # 项目难点解决方案
│
├── error-handling/                    # 错误处理文档
│   ├── process.md                     # 错误处理流程
│   ├── solutions.md                   # 常见错误解决方案
│   └── iteration-errors-500.md        # 迭代错误记录（500 轮）
│
├── frontend/                          # 前端文档
│   ├── architecture.md                # 前端架构说明
│   ├── design.md                      # 前端页面设计规范
│   └── optimization.md                # 前端优化与设计提升需求
│
├── plans/                             # 计划文档
│   ├── 1000-iterations-master-plan.md # 1000 轮迭代总计划
│   └── 2025-04-05-incomplete-modules-development.md # 未完成模块开发计划
│
├── report/                            # 报表与报告文档
│   ├── overview.md                    # 报表功能说明
│   └── export.md                      # 报表导出实现
│
├── skills/                            # 技能文档
│   ├── golang-development.md          # Go 后端开发技能
│   ├── vue3-development.md            # Vue 3 前端开发技能
│   ├── testing-skills.md              # 测试技能
│   ├── code-review-skills.md          # 代码审查技能
│   └── project-skills-500.md          # 项目技能（500 轮）
│
├── superpowers/                       # Superpower 技能规格
│   └── specs/                         # 技能规格文档
│
├── swagger/                           # Swagger 接口文档
│   └── README.md                      # Swagger 使用说明
│
└── wiki/                              # 知识库 Wiki
    ├── Home.md                        # 知识库首页
    ├── Quick-Start.md                 # 快速开始
    ├── Installation-Guide.md          # 安装指南
    ├── Project-Structure.md           # 项目结构
    ├── Configuration.md               # 配置说明
    ├── API-Documentation.md           # API 文档
    ├── Feature-Guide.md               # 功能指南
    ├── FAQ.md                         # 常见问题
    ├── SYNC-GUIDE.md                  # 同步指南
    ├── GitHub-Secrets-Setup.md        # GitHub Secrets 配置
    └── llm-wiki/                      # LLM-wiki 知识库
        └── index.md                   # LLM-wiki 索引
```

> 注：上述结构反映当前已存在的目录与文件。`docs/` 下还存在若干历史主题文档（如 `architecture.md`、`database-design.md`、`api-documentation.md`、`deployment-guide.md`、`developer-guide.md`、`operations-guide.md`、`security-audit-report.md`、`performance-benchmark.md` 等），将按编号文档体系逐步整合归并。

---

## 三、文档分类说明表

| 分类 | 目录 | 说明 | 典型文档 |
|-----|------|------|---------|
| 流程编号文档 | `docs/01-15_*.md` | 按研发流程顺序组织的核心文档，覆盖需求→交付全生命周期 | 需求分析、实施计划、测试计划、项目 Wiki、技能配置等 |
| 架构设计 | `docs/architecture/` | 系统架构设计、技术选型、难点解决方案 | challenges.md |
| 错误处理 | `docs/error-handling/` | 错误处理流程、常见错误解决方案、迭代错误记录 | process.md、solutions.md、iteration-errors-500.md |
| 前端文档 | `docs/frontend/` | 前端架构、页面设计规范、优化方案 | architecture.md、design.md、optimization.md |
| 计划文档 | `docs/plans/` | 实施计划、迭代计划、模块开发计划 | 1000-iterations-master-plan.md |
| 报表报告 | `docs/report/` | 报表功能说明、导出实现、项目报告 | overview.md、export.md |
| 技能文档 | `docs/skills/` | AI 代理与开发者技能规范 | golang-development.md、vue3-development.md、testing-skills.md、code-review-skills.md |
| Superpower 规格 | `docs/superpowers/` | Superpower 技能集规格文档 | specs/ |
| 接口文档 | `docs/swagger/` | Swagger 自动生成的接口文档 | README.md |
| 知识库 | `docs/wiki/` | 项目知识库 Wiki，含使用指南与常见问题 | Home.md、Quick-Start.md、FAQ.md |

### 编号文档（01-15）速查

| 编号 | 文档 | 说明 | 状态 |
|-----|------|------|------|
| 01 | `01_requirements_analysis.md` | 需求分析 | ✅ 已存在 |
| 02 | `02_requirements_review.md` | 需求评审 | ✅ 已存在 |
| 03 | `03_implementation_plan.md` | 实施计划 | ✅ 已存在 |
| 04 | `04_test_plan.md` | 测试计划 | ✅ 已存在 |
| 05 | `05_code_review.md` | 代码审查 | ✅ 已存在 |
| 06 | `06_iteration_plan.md` | 迭代计划 | ✅ 已存在 |
| 07 | `07_task_list.md` | 任务清单 | ✅ 已存在 |
| 08 | `08_development_plan_phase2.md` | Phase 2 开发计划 | ✅ 已存在 |
| 09 | `09_comprehensive_test_plan.md` | 综合测试计划 | ✅ 已存在 |
| 10 | `10_market_requirements_review.md` | 市场需求评审 | ✅ 已存在 |
| 11 | `11_long_term_roadmap.md` | 长期路线图 | ✅ 已存在 |
| 12 | `12_tech_debt_register.md` | 技术债务登记 | ✅ 已存在 |
| 13 | `13_project_wiki.md` | 项目知识 Wiki | ✅ 已存在 |
| 14 | `14_document_version_control.md` | 文档版本控制机制 | ✅ 已存在 |
| 15 | `15_skills_configuration.md` | 开发技能与工具配置 | ✅ 已存在 |

---

## 四、变更记录表

| 版本 | 日期 | 变更内容 | 变更人 |
|-----|------|---------|-------|
| v1.0 | 2026-06-17 | 初始创建：建立文档管理规则、目录结构、文档分类说明 | AI Agent |
| v2.0.0 | 2026-06-17 | 重大更新：新增 13-15 编号文档（项目 Wiki、文档版本控制、技能配置）；新增 skills/testing-skills.md、skills/code-review-skills.md；更新目录结构反映当前实际文件；补充编号文档速查表；版本号规范对齐 14_document_version_control.md | AI Agent |
