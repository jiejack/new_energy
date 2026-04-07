# gstack/Superpowers 工作流集成指南

## 概述

gstack 是一套完整的 AI 驱动开发工作流工具集,包含 37 个专业技能模块,覆盖从需求分析到生产部署的完整软件开发生命周期。本指南将帮助团队在新能源监控项目中有效集成和使用这些技能。

## 安装位置

gstack 已安装在全局目录:
```
C:\Users\jieja\.claude\skills\gstack
```

## 核心工作流

### 1. 需求澄清阶段

#### /office-hours - YC 办公时间模式
**用途**: 项目启动时的需求澄清和产品定位

**触发场景**:
- 新功能开发前的需求分析
- 产品方向不明确时
- 需要重新思考产品定位时

**工作流程**:
1. 六个强制性问题重新框架产品思考
2. 前提假设验证
3. 生成 2-3 个实施方案
4. 输出设计文档到 `~/.gstack/projects/`

**使用示例**:
```
用户: 我们需要开发一个新的告警通知系统
Claude: [自动调用 /office-hours]
        让我们先理解真正的问题。谁需要这个通知系统?
        他们现在如何处理告警?具体痛点是什么?
```

#### /plan-ceo-review - CEO/创始人模式
**用途**: 战略层面审查和产品愿景优化

**触发场景**:
- 需要重新思考产品方向
- 评估功能优先级
- 扩展或缩减项目范围

**四种模式**:
- **SCOPE EXPANSION**: 扩大范围,探索更大可能性
- **SELECTIVE EXPANSION**: 选择性扩展,逐个评估机会
- **HOLD SCOPE**: 保持现有范围,最大化执行质量
- **SCOPE REDUCTION**: 缩减范围,聚焦核心价值

**使用示例**:
```
用户: 这个功能是不是太简单了?
Claude: [自动调用 /plan-ceo-review]
        让我们从用户角度重新思考。这个功能的真正价值是什么?
        是否有更好的实现方式?
```

### 2. 架构设计阶段

#### /plan-eng-review - 工程经理模式
**用途**: 技术架构审查和实现细节确认

**触发场景**:
- 确定技术方案
- 设计系统架构
- 评估技术风险

**审查内容**:
- 架构设计
- 系统边界
- 数据流
- 状态转换
- 失败模式
- 边界情况
- 信任边界
- 测试覆盖

**输出**:
- 架构图
- 数据流图
- 测试矩阵
- 技术决策文档

**使用示例**:
```
用户: 我们需要实现实时数据采集系统
Claude: [自动调用 /plan-eng-review]
        让我分析架构需求:
        - 数据采集频率和延迟要求
        - 协议支持(Modbus/IEC104/IEC61850)
        - 数据存储方案
        - 失败重试机制
```

#### /plan-design-review - 设计师模式
**用途**: 设计层面的计划审查

**触发场景**:
- UI/UX 相关功能开发
- 用户界面设计
- 交互流程设计

**七个维度审查**:
1. 信息架构
2. 交互状态覆盖
3. 用户旅程
4. AI 生成内容风险
5. 设计系统一致性
6. 响应式/可访问性
7. 未解决的设计决策

**使用示例**:
```
用户: 设计一个新的仪表盘页面
Claude: [自动调用 /plan-design-review]
        初始设计评分: 4/10
        问题: 缺少空状态、加载状态、错误状态定义
        正在补充交互状态表...
```

### 3. 开发实现阶段

#### /autoplan - 自动审查流水线
**用途**: 一键运行完整审查流程

**触发场景**:
- 需要快速完成所有审查
- 不想逐个回答中间问题

**工作流程**:
```
CEO Review → Design Review → Eng Review → DX Review
```

**自动决策原则**:
1. 优先完整性
2. 匹配现有模式
3. 选择可逆选项
4. 参考历史决策
5. 延迟模糊项
6. 升级安全问题

**使用示例**:
```
用户: /autoplan
Claude: 运行 CEO 审查... [4 个范围决策自动解决]
        运行设计审查... [3 个设计维度自动评分]
        运行工程审查... [2 个架构决策自动解决]

        品味决策(需要您的输入):
        1. 范围: Codex 建议添加搜索功能 - 边界扩展。是否添加?
        2. 设计: 两种方案评分相差 1 分。哪个更合适?
```

#### TDD - 测试驱动开发
**集成方式**: 通过 /ship 技能自动集成

**测试框架**:
- 后端: Go testing + testify
- 前端: Vitest + Playwright

**测试类型**:
- 单元测试
- 集成测试
- E2E 测试
- 性能测试

#### /review - 代码审查
**用途**: 合并前的代码质量审查

**触发场景**:
- 准备提交 PR
- 代码审查请求
- 检查代码变更

**审查内容**:
- SQL 注入风险
- LLM 信任边界
- 条件副作用
- N+1 查询
- 竞态条件
- 安全漏洞
- 性能问题

**自动修复**:
- 明显的机械问题自动修复
- 复杂问题提交给用户决策

**使用示例**:
```
用户: /review
Claude: 分析变更...

        发现问题:
        [HIGH] N+1 查询: internal/application/service/alarm_service.go:47
               告警列表查询未预加载关联数据
        [AUTO-FIXED] 已添加 Preload() 预加载

        [MEDIUM] 潜在竞态条件: pkg/collector/buffer.go:89
                 并发写入未加锁
        需要决策: 使用互斥锁还是通道?
```

### 4. 测试验证阶段

#### /qa - QA 测试
**用途**: 系统化测试和 Bug 修复

**触发场景**:
- 功能开发完成
- 准备发布前
- 回归测试

**三种模式**:
- **Quick**: 关键/高优先级问题(30 秒)
- **Standard**: + 中等优先级问题(5-15 分钟)
- **Exhaustive**: + 外观问题(完整测试)

**工作流程**:
1. 读取 git diff 识别变更页面
2. 启动浏览器测试
3. 发现 Bug
4. 原子化修复
5. 重新验证
6. 生成回归测试

**使用示例**:
```
用户: /qa https://staging.new-energy.local
Claude: [探索 12 个页面,填写 3 个表单,测试 2 个流程]

        QA 报告: staging.new-energy.local - 健康评分: 72/100

        前 3 个问题:
        1. CRITICAL: 告警表单提交时必填字段为空仍可提交
        2. HIGH: 移动端导航菜单选择后未关闭
        3. MEDIUM: 仪表盘图表在 1024px 以下与侧边栏重叠

        [完整报告已保存到 .gstack/qa-reports/]
```

#### /cso - 安全审计
**用途**: OWASP Top 10 + STRIDE 安全审计

**触发场景**:
- 安全审计需求
- 处理敏感数据
- API 安全检查

**审查内容**:
- 注入漏洞
- 认证问题
- 敏感数据暴露
- XML 外部实体
- 访问控制
- 安全配置
- XSS
- 不安全的反序列化
- 已知漏洞组件
- 日志不足

**使用示例**:
```
用户: /cso
Claude: 运行 OWASP Top 10 + STRIDE 安全审计...

        CRITICAL: SQL 注入风险 (internal/api/handler/device_handler.go:47)
        HIGH: 会话令牌存储在 localStorage (web/src/utils/auth.ts:12)
        MEDIUM: /api/login 端点缺少速率限制

        4 个发现,扫描 12 个文件。1 个严重,1 个高危。
```

### 5. 发布部署阶段

#### /ship - 发布工程
**用途**: 自动化发布流程

**触发场景**:
- 准备发布
- 创建 PR
- 更新版本

**工作流程**:
1. 同步主分支
2. 运行测试
3. 审查覆盖率
4. 更新 CHANGELOG
5. 推送代码
6. 创建/更新 PR

**测试引导**:
- 自动检测测试框架
- 安装最佳实践框架
- 生成 3-5 个真实测试
- 设置 CI/CD
- 创建 TESTING.md

**使用示例**:
```
用户: /ship
Claude: 同步 main 分支...
        运行测试: 47/47 通过
        覆盖率审计: 78% (目标: 80%)
        更新 CHANGELOG...
        推送到 origin/feature/alarm-notification...

        PR 已创建: https://github.com/your-org/new-energy-monitoring/pull/42
        测试: 42 → 47 (+5 新增)
```

#### /canary - 金丝雀监控
**用途**: 发布后监控

**触发场景**:
- 新版本发布后
- 监控生产环境
- 检测性能回归

**监控内容**:
- 控制台错误
- 性能回归
- 页面失败
- 视觉异常

**使用示例**:
```
用户: /canary https://new-energy.local
Claude: 监控 8 个页面,每 2 分钟检查一次...

        周期 1: ✓ 所有页面健康。p95: 340ms。0 个控制台错误。
        周期 2: ✓ 所有页面健康。p95: 380ms。0 个控制台错误。
        周期 3: ⚠ /dashboard — 新控制台错误:
                 "TypeError: Cannot read property 'map' of undefined"
                 at dashboard.js:142
                 截图已保存。

        警报: 3 个监控周期后发现 1 个新控制台错误。
```

## 项目特定配置

### 测试命令配置

在 `CLAUDE.md` 中配置:

```markdown
## Test commands

Backend tests:
- Unit: `go test ./...`
- Integration: `go test -tags=integration ./...`
- Coverage: `go test -coverprofile=coverage.out ./...`

Frontend tests:
- Unit: `cd web && npm run test`
- E2E: `cd web && npm run test:e2e`
- Coverage: `cd web && npm run test:coverage`

Performance tests:
- Load test: `cd scripts/performance && k6 run load_test.js`
- Benchmark: `cd tests/performance && ./run_benchmarks.sh`
```

### 部署配置

```markdown
## Deploy commands

Development:
- Deploy: `docker-compose up -d`
- Logs: `docker-compose logs -f`

Production:
- Deploy: `kubectl apply -f k8s/`
- Status: `kubectl get pods -n new-energy`
- Rollback: `kubectl rollout undo deployment/api-server -n new-energy`

Health checks:
- API: `curl http://localhost:8080/health`
- Frontend: `curl http://localhost:3000/health`
```

### 技能路由规则

```markdown
## Skill routing

When the user's request matches an available skill, ALWAYS invoke it using the Skill
tool as your FIRST action. Do NOT answer directly, do NOT use other tools first.

Key routing rules:
- Product ideas, "is this worth building", brainstorming → invoke office-hours
- Bugs, errors, "why is this broken", 500 errors → invoke investigate
- Ship, deploy, push, create PR → invoke ship
- QA, test the site, find bugs → invoke qa
- Code review, check my diff → invoke review
- Security audit, OWASP, vulnerability scan → invoke cso
- Performance issues, benchmark → invoke benchmark
- Architecture review → invoke plan-eng-review
```

## 最佳实践

### 1. 工作流顺序

推荐的开发流程:

```
需求澄清 → 架构设计 → 开发实现 → 测试验证 → 发布部署
    ↓           ↓           ↓           ↓           ↓
office-hours → plan-eng-review → TDD → qa → ship
    ↓           ↓           ↓           ↓       ↓
plan-ceo-review → plan-design-review → review → cso → canary
```

### 2. 技能组合使用

**新功能开发**:
```
/office-hours → /plan-eng-review → 开发 → /review → /qa → /ship
```

**Bug 修复**:
```
/investigate → 修复 → /review → /qa → /ship
```

**安全审计**:
```
/cso → 修复 → /review → /qa → /ship
```

**性能优化**:
```
/benchmark → 优化 → /review → /benchmark → /qa → /ship
```

### 3. 自动化程度

**完全自动化**:
- /autoplan: 自动审查流程
- /ship: 自动发布流程
- /canary: 自动监控

**半自动化**:
- /review: 自动修复明显问题,复杂问题需确认
- /qa: 自动测试,修复需确认
- /cso: 自动扫描,修复需确认

**交互式**:
- /office-hours: 需要用户输入
- /plan-ceo-review: 需要战略决策
- /plan-eng-review: 需要技术决策

### 4. 质量门禁

**必须通过**:
- /review: 无严重问题
- /qa: 健康评分 ≥ 80
- /cso: 无严重安全漏洞

**建议通过**:
- /plan-eng-review: 架构审查
- /benchmark: 性能基准
- /canary: 发布后监控

## 常见问题

### Q: 如何启用/禁用主动建议?

A: 使用配置命令:
```bash
# 启用主动建议
~/.claude/skills/gstack/bin/gstack-config set proactive true

# 禁用主动建议
~/.claude/skills/gstack/bin/gstack-config set proactive false
```

### Q: 如何查看学习记录?

A: 使用 /learn 技能:
```
用户: /learn
Claude: 23 条学习记录 (14 高置信度, 6 中等, 3 低)

        顶部模式:
        - [9/10] API 响应总是包装在 { data, error } 信封中
        - [8/10] 测试使用 test/support/factories.ts 中的工厂助手
        - [8/10] 所有 DB 查询通过仓储模式,从不直接查询
```

### Q: 如何升级 gstack?

A: 使用升级技能:
```
用户: /gstack-upgrade
Claude: 当前版本: 0.7.4
        最新版本: 0.8.2

        新功能:
        - 浏览器移交用于 CAPTCHA 和认证墙
        - /codex 多 AI 第二意见
        - /qa 现在总是使用浏览器
        - 安全技能: /careful, /freeze, /guard

        已升级到 0.8.2。
```

### Q: 如何处理认证页面测试?

A: 使用 cookie 导入:
```
用户: /setup-browser-cookies
Claude: Cookie 选择器已打开 — 在浏览器中选择要导入的域名,
        然后告诉我完成。

        [你在浏览器中选择 github.com, myapp.com]

用户: done
Claude: 已导入 2 个域名(47 个 cookie)。会话已准备就绪。
```

## 技能完整列表

### 规划类
- /office-hours: YC 办公时间模式
- /plan-ceo-review: CEO 审查
- /plan-eng-review: 工程审查
- /plan-design-review: 设计审查
- /autoplan: 自动审查流水线

### 开发类
- /review: 代码审查
- /investigate: 问题调查
- /codex: 多 AI 第二意见

### 测试类
- /qa: QA 测试
- /qa-only: 仅报告模式
- /benchmark: 性能基准
- /cso: 安全审计

### 部署类
- /ship: 发布工程
- /land-and-deploy: 合并部署
- /canary: 金丝雀监控
- /document-release: 文档更新

### 设计类
- /design-consultation: 设计咨询
- /design-review: 设计审查
- /design-shotgun: 设计探索
- /design-html: 设计转代码

### 工具类
- /browse: 浏览器控制
- /setup-browser-cookies: Cookie 导入
- /learn: 学习管理
- /retro: 回顾总结

### 安全类
- /careful: 谨慎模式
- /freeze: 编辑锁定
- /guard: 完全安全模式
- /unfreeze: 解除锁定

### 其他
- /checkpoint: 检查点
- /health: 健康检查
- /gstack-upgrade: 升级工具

## 参考资源

- [gstack 官方文档](https://github.com/gstack/gstack)
- [技能详细说明](C:\Users\jieja\.claude\skills\gstack\docs\skills.md)
- [浏览器使用指南](C:\Users\jieja\.claude\skills\gstack\BROWSER.md)
- [架构设计文档](C:\Users\jieja\.claude\skills\gstack\ARCHITECTURE.md)

## 更新记录

- 2026-04-07: 创建初始版本
