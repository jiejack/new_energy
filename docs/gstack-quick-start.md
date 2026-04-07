# gstack 快速上手指南

## 目录

1. [环境准备](#环境准备)
2. [基础使用](#基础使用)
3. [常见场景](#常见场景)
4. [最佳实践](#最佳实践)
5. [故障排查](#故障排查)

## 环境准备

### 1. 确认安装

gstack 已安装在全局目录:
```bash
ls C:\Users\jieja\.claude\skills\gstack
```

### 2. 初始配置

首次使用时,gstack 会自动引导配置:

```bash
# 启动 Claude Code
cd e:\ai_work\new-energy-monitoring
claude

# gstack 会自动检测并提示:
# - 遥测设置
# - 主动建议设置
# - 技能路由规则
```

### 3. 验证配置

检查配置文件:
```bash
# 查看项目配置
cat CLAUDE.md

# 查看全局配置
cat ~/.gstack/config.yaml
```

## 基础使用

### 1. 需求澄清

**场景**: 开发新功能前的需求分析

```
用户: 我们需要开发一个新的告警通知系统

Claude: [自动调用 /office-hours]

让我先理解真正的问题:

1. 谁需要这个通知系统?
   - 运维人员?设备管理员?还是两者都需要?

2. 他们现在如何处理告警?
   - 邮件?短信?还是人工巡检?

3. 具体痛点是什么?
   - 响应不及时?通知遗漏?还是噪音太多?

请告诉我具体的使用场景...
```

### 2. 架构设计

**场景**: 确定技术方案

```
用户: 我们需要实现实时数据采集系统

Claude: [自动调用 /plan-eng-review]

让我分析架构需求:

## 数据采集架构

### 1. 协议支持
- Modbus TCP/RTU
- IEC 104
- IEC 61850

### 2. 数据流设计
[采集器] → [缓冲区] → [处理器] → [存储]
   ↓           ↓           ↓         ↓
 协议解析    数据验证    业务计算   时序数据库

### 3. 性能要求
- 采集频率: 1秒/次
- 延迟要求: < 100ms
- 并发连接: 1000+

### 4. 失败处理
- 重试机制: 指数退避
- 断线重连: 自动重连
- 数据缓存: 本地缓存

需要我详细设计某个部分吗?
```

### 3. 代码审查

**场景**: 提交代码前的质量检查

```
用户: /review

Claude: 分析变更...

## 审查结果

### 发现的问题

**[HIGH] N+1 查询**
位置: internal/application/service/alarm_service.go:47
问题: 告警列表查询未预加载关联数据
影响: 每个告警都会额外查询一次数据库
修复: 已自动添加 Preload() 预加载

**[MEDIUM] 潜在竞态条件**
位置: pkg/collector/buffer.go:89
问题: 并发写入未加锁
建议: 使用 sync.Mutex 或 channel

### 统计
- 扫描文件: 12
- 发现问题: 2
- 自动修复: 1
- 需要决策: 1

是否修复竞态条件问题?
```

### 4. QA 测试

**场景**: 功能开发完成后的测试

```
用户: /qa https://localhost:3000

Claude: [启动浏览器,开始测试]

## QA 测试进度

✓ 首页加载
✓ 登录流程
✓ 仪表盘显示
✓ 设备列表
✓ 告警管理
✓ 数据查询

## 发现的问题

**[CRITICAL] 告警表单验证**
位置: web/src/views/alarm/list/index.vue
问题: 必填字段为空时仍可提交
修复: 添加表单验证规则

**[HIGH] 移动端导航**
位置: web/src/components/MobileNav/index.vue
问题: 菜单选择后未自动关闭
修复: 添加菜单关闭逻辑

## 测试报告
- 测试页面: 12
- 发现问题: 2
- 健康评分: 72/100

是否自动修复这些问题?
```

### 5. 发布部署

**场景**: 准备发布新版本

```
用户: /ship

Claude: 开始发布流程...

## 发布步骤

### 1. 同步主分支
✓ git fetch origin
✓ git merge origin/main

### 2. 运行测试
✓ 后端测试: 47/47 通过
✓ 前端测试: 32/32 通过
✓ E2E 测试: 15/15 通过

### 3. 覆盖率检查
当前: 78%
目标: 80%
状态: 接近目标

### 4. 更新文档
✓ CHANGELOG.md
✓ README.md

### 5. 创建 PR
✓ 推送到 origin/feature/alarm-notification
✓ 创建 Pull Request #42

## 发布摘要
- 分支: feature/alarm-notification
- 提交: 12
- 变更: +1,234 -567
- 测试: +5 新增

PR 地址: https://github.com/your-org/new-energy-monitoring/pull/42
```

## 常见场景

### 场景 1: 新功能开发

完整流程:

```
1. 需求澄清
用户: 我们需要开发一个新的数据导出功能
Claude: [调用 /office-hours]
        让我理解需求...

2. 架构设计
用户: 确认技术方案
Claude: [调用 /plan-eng-review]
        设计导出架构...

3. 开发实现
用户: 开始实现
Claude: [使用 TDD 模式]
        先写测试,再实现...

4. 代码审查
用户: /review
Claude: 审查代码...

5. QA 测试
用户: /qa
Claude: 测试功能...

6. 发布部署
用户: /ship
Claude: 创建 PR...
```

### 场景 2: Bug 修复

快速修复流程:

```
1. 问题调查
用户: 告警通知没有发送
Claude: [调用 /investigate]
        系统调查中...

        可能原因:
        1. 邮件服务配置错误
        2. 通知规则未触发
        3. 队列阻塞

        让我检查邮件服务配置...

2. 修复问题
Claude: 发现问题: SMTP 配置缺少端口
        修复: 添加端口配置

3. 验证修复
用户: /qa
Claude: 测试修复...

4. 发布修复
用户: /ship
Claude: 创建修复 PR...
```

### 场景 3: 性能优化

性能优化流程:

```
1. 性能基准
用户: /benchmark
Claude: 建立性能基准...

        当前性能:
        - 首页加载: 1.2s
        - API 响应: 340ms
        - 数据库查询: 120ms

2. 识别瓶颈
Claude: 发现性能瓶颈:
        1. N+1 查询
        2. 缺少索引
        3. 未使用缓存

3. 实施优化
Claude: 应用优化:
        - 添加预加载
        - 创建索引
        - 启用缓存

4. 验证优化
用户: /benchmark
Claude: 对比性能...

        优化后:
        - 首页加载: 0.8s (-33%)
        - API 响应: 180ms (-47%)
        - 数据库查询: 45ms (-62%)

5. 发布优化
用户: /ship
Claude: 创建优化 PR...
```

### 场景 4: 安全审计

安全检查流程:

```
1. 安全扫描
用户: /cso
Claude: 运行安全审计...

        发现安全问题:
        [CRITICAL] SQL 注入风险
        [HIGH] 会话令牌存储不安全
        [MEDIUM] 缺少速率限制

2. 修复问题
Claude: 修复安全问题:
        ✓ 参数化查询
        ✓ 使用 HttpOnly Cookie
        ✓ 添加速率限制

3. 验证修复
用户: /cso
Claude: 重新扫描...

        安全评分: A
        所有严重问题已修复

4. 发布修复
用户: /ship
Claude: 创建安全修复 PR...
```

## 最佳实践

### 1. 工作流顺序

推荐的开发流程:

```
需求 → 设计 → 开发 → 测试 → 发布
  ↓      ↓      ↓      ↓      ↓
office  plan   TDD    qa    ship
-hours  -eng          ↓
        -review     review
```

### 2. 技能组合

**新功能**:
```
/office-hours → /plan-eng-review → 开发 → /review → /qa → /ship
```

**Bug 修复**:
```
/investigate → 修复 → /review → /qa → /ship
```

**性能优化**:
```
/benchmark → 优化 → /review → /benchmark → /qa → /ship
```

**安全审计**:
```
/cso → 修复 → /review → /qa → /ship
```

### 3. 质量门禁

**必须通过**:
- /review: 无严重问题
- /qa: 健康评分 ≥ 80
- /cso: 无严重安全漏洞

**建议通过**:
- /plan-eng-review: 架构审查
- /benchmark: 性能基准
- /canary: 发布后监控

### 4. 自动化程度

**完全自动化**:
- /autoplan: 自动审查
- /ship: 自动发布
- /canary: 自动监控

**半自动化**:
- /review: 自动修复明显问题
- /qa: 自动测试
- /cso: 自动扫描

**交互式**:
- /office-hours: 需要用户输入
- /plan-ceo-review: 需要战略决策
- /plan-eng-review: 需要技术决策

## 故障排查

### 问题 1: 技能未触发

**症状**: 输入命令后技能未执行

**解决方案**:
```bash
# 检查配置
cat ~/.gstack/config.yaml

# 确认主动模式已启用
~/.claude/skills/gstack/bin/gstack-config get proactive

# 如果未启用,启用它
~/.claude/skills/gstack/bin/gstack-config set proactive true
```

### 问题 2: 浏览器测试失败

**症状**: /qa 或 /browse 失败

**解决方案**:
```bash
# 检查浏览器是否安装
which chromium

# 检查浏览器服务
curl http://localhost:9222

# 重启浏览器服务
pkill -f chromium
```

### 问题 3: 测试覆盖率不足

**症状**: /ship 时覆盖率检查失败

**解决方案**:
```bash
# 查看覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 添加缺失的测试
# 参考 docs/test-plan.md
```

### 问题 4: 安全扫描误报

**症状**: /cso 报告误报

**解决方案**:
```bash
# 查看历史记录
cat ~/.gstack/greptile-history.md

# 标记为误报
# 在 Claude 中说明为什么是误报
# Claude 会自动记录并跳过
```

### 问题 5: 性能测试失败

**症状**: /benchmark 失败

**解决方案**:
```bash
# 检查服务状态
docker ps

# 检查资源使用
docker stats

# 检查网络连接
curl http://localhost:8080/health
```

## 高级技巧

### 1. 批量操作

使用 chain 命令批量执行:

```bash
# 批量浏览器操作
echo '[
  ["goto","https://localhost:3000"],
  ["snapshot","-i"],
  ["fill","@e3","admin"],
  ["fill","@e4","password"],
  ["click","@e5"],
  ["snapshot","-D"],
  ["screenshot","/tmp/login.png"]
]' | $B chain
```

### 2. 自定义审查

创建自定义审查配置:

```markdown
## Custom review rules

在 CLAUDE.md 中添加:

### 必须检查
- [ ] 所有 API 都有单元测试
- [ ] 所有数据库查询都有索引
- [ ] 所有外部调用都有超时设置
- [ ] 所有错误都有日志记录
```

### 3. 学习记录管理

查看和管理学习记录:

```bash
# 查看学习记录
用户: /learn
Claude: 23 条学习记录...

# 搜索特定模式
用户: /learn search "API 响应格式"

# 清理过期记录
用户: /learn prune
```

### 4. 回顾总结

定期回顾工作:

```bash
# 周回顾
用户: /retro
Claude: 本周工作总结...

        ## 你的工作
        - 提交: 32
        - 代码: +2.4k LOC
        - 测试: 41%
        - 最大贡献: 告警通知系统

        ## 团队工作
        - Alice: 12 提交,专注告警模块
        - Bob: 3 提交,性能优化

        ## 改进建议
        1. 测试覆盖率需要提升
        2. 代码审查频率可以增加
        3. 文档更新需要及时
```

## 参考资源

- [完整工作流指南](./ai-workflow-guide.md)
- [项目配置](../CLAUDE.md)
- [gstack 官方文档](https://github.com/gstack/gstack)
- [技能详细说明](C:\Users\jieja\.claude\skills\gstack\docs\skills.md)

## 更新记录

- 2026-04-07: 创建初始版本
