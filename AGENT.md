# 新能源监控系统知识库

## 项目核心要点

- **技术栈**：Go 1.24+ 后端，Vue 3+ 前端，Gin 框架，Excelize 库
- **架构**：前后端分离，API 路径前缀 `/api/v1`
- **分支管理**：开发分支 `dev`，主分支需审核后合并
- **分层架构**：Entity → Repository → Service → Handler
- **依赖注入**：使用 wire 进行依赖注入
- **LLM-wiki 架构**：采用 Andrej Karpathy 的 LLM-wiki 三层架构思想

## 开发规则

- 遵循 "文档先行" 原则，编码前完成专业文档编写
- 代码提交需推送到 `dev` 分支，禁止直接推送到主分支
- 前端使用 TypeScript，后端使用 Go 标准库和指定框架
- 所有 API 接口需保持前后端一致性
- 代码质量要求：无 TypeScript 错误，前端构建成功
- 所有项目文档统一存放在 `docs/` 目录下

## LLM-wiki 知识库索引

### 知识库总览
- [知识库索引](file:///workspace/docs/wiki/llm-wiki/index.md) - 知识库完整索引和分类

### 来源摘要
- [Andrej Karpathy LLM-wiki 概念](file:///workspace/docs/wiki/llm-wiki/sources/karpathy-llm-wiki.md) - LLM-wiki 核心概念和三层架构

### 实体文档
- [Go 编程语言](file:///workspace/docs/wiki/llm-wiki/entities/go-language.md) - Go 1.24+ 编程语言
- [Gin Web 框架](file:///workspace/docs/wiki/llm-wiki/entities/gin-framework.md) - Go Web 框架
- [Vue 3 前端框架](file:///workspace/docs/wiki/llm-wiki/entities/vue3-framework.md) - Vue.js 3.x 前端框架
- [Excelize 库](file:///workspace/docs/wiki/llm-wiki/entities/excelize-library.md) - Go Excel 文档操作库
- [TypeScript 编程语言](file:///workspace/docs/wiki/llm-wiki/entities/typescript-language.md) - TypeScript 5.0+ 编程语言

### 概念文档
- [分层架构设计](file:///workspace/docs/wiki/llm-wiki/concepts/layered-architecture.md) - 本项目的分层架构设计说明

### 业务需求
- [能效分析功能](file:///workspace/docs/wiki/llm-wiki/requirements/energy-efficiency-analysis.md) - 能效分析功能需求文档
- [碳排放监测功能](file:///workspace/docs/wiki/llm-wiki/requirements/carbon-emission-monitoring.md) - 碳排放监测功能需求文档

### 领域知识
- [新能源行业知识](file:///workspace/docs/wiki/llm-wiki/domains/new-energy-industry.md) - 新能源行业概述、技术、标准等

## 技能文档索引

### 前端开发
- [Vue 3 开发技能](file:///workspace/docs/skills/vue3-development.md) - Vue 3 + TypeScript + Vite 开发指南

### 后端开发
- [Go 后端开发技能](file:///workspace/docs/skills/golang-development.md) - Go 1.24+ + Gin 开发指南

## 功能文档索引

### 统计报表功能
- [报表功能说明](file:///workspace/docs/report/overview.md)
- [报表导出实现](file:///workspace/docs/report/export.md)

### 前端设计
- [前端架构说明](file:///workspace/docs/frontend/architecture.md)
- [前端页面设计规范](file:///workspace/docs/frontend/design.md)
- [前端优化与设计提升需求分析](file:///workspace/docs/frontend/optimization.md)

### 错误处理
- [错误处理流程](file:///workspace/docs/error-handling/process.md)
- [常见错误解决方案](file:///workspace/docs/error-handling/solutions.md)

## 项目难点解决方案
- [项目难点解决方案](file:///workspace/docs/architecture/challenges.md)
