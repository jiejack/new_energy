# 新能源监控系统 - Claude Code 项目配置

## 项目概述

新能源监控系统是一个完整的工业物联网监控平台,包含数据采集、实时计算、告警管理、智能问答等核心功能。

### 技术栈

**后端**:
- 语言: Go 1.21+
- 框架: Gin
- 数据库: PostgreSQL + TimescaleDB
- 缓存: Redis
- 消息队列: Kafka
- 容器化: Docker + Kubernetes

**前端**:
- 框架: Vue 3 + TypeScript
- UI 库: Element Plus
- 状态管理: Pinia
- 构建工具: Vite
- 测试: Vitest + Playwright

### 项目结构

```
new-energy-monitoring/
├── cmd/                    # 应用入口
│   ├── api-server/        # API 服务器
│   ├── collector/         # 数据采集服务
│   ├── compute/           # 计算服务
│   ├── alarm/             # 告警服务
│   ├── scheduler/         # 调度服务
│   └── ai-service/        # AI 服务
├── internal/              # 内部代码
│   ├── api/              # API 层
│   ├── application/      # 应用层
│   ├── domain/           # 领域层
│   └── infrastructure/   # 基础设施层
├── pkg/                   # 公共包
│   ├── ai/               # AI 相关
│   ├── alarm/            # 告警系统
│   ├── auth/             # 认证授权
│   ├── collector/        # 数据采集
│   ├── compute/          # 计算引擎
│   ├── harness/          # 测试工具
│   └── protocol/         # 工业协议
├── web/                   # 前端代码
│   ├── src/              # 源代码
│   ├── public/           # 静态资源
│   └── tests/            # 测试
├── deployments/           # 部署配置
│   ├── docker/           # Docker 配置
│   └── kubernetes/       # K8s 配置
├── configs/               # 配置文件
├── scripts/               # 脚本工具
└── docs/                  # 文档
```

## Test commands

### 后端测试

```bash
# 单元测试
go test ./...

# 集成测试
go test -tags=integration ./...

# 覆盖率测试
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 性能测试
cd tests/performance
./run_benchmarks.sh

# 压力测试
cd scripts/performance
k6 run load_test.js
```

### 前端测试

```bash
cd web

# 单元测试
npm run test

# E2E 测试
npm run test:e2e

# 覆盖率测试
npm run test:coverage

# 性能测试
npm run test:performance
```

### 测试覆盖率目标

- 后端: ≥ 80%
- 前端: ≥ 75%
- 关键路径: 100%

## Deploy commands

### 开发环境

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 重启服务
docker-compose restart

# 停止服务
docker-compose down
```

### 生产环境

```bash
# 部署到 Kubernetes
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/

# 查看部署状态
kubectl get pods -n new-energy
kubectl get services -n new-energy

# 查看日志
kubectl logs -f deployment/api-server -n new-energy

# 回滚
kubectl rollout undo deployment/api-server -n new-energy
```

### Helm 部署

```bash
# 开发环境
helm install new-energy deployments/kubernetes/helm \
  -f deployments/kubernetes/helm/values-dev.yaml \
  -n new-energy

# 生产环境
helm upgrade new-energy deployments/kubernetes/helm \
  -f deployments/kubernetes/helm/values-prod.yaml \
  -n new-energy
```

### 健康检查

```bash
# API 健康检查
curl http://localhost:8080/health

# 前端健康检查
curl http://localhost:3000/health

# 数据库连接检查
curl http://localhost:8080/health/db

# Redis 连接检查
curl http://localhost:8080/health/redis

# Kafka 连接检查
curl http://localhost:8080/health/kafka
```

## Skill routing

当用户的请求匹配可用技能时,始终首先使用 Skill 工具调用它。不要直接回答,不要先使用其他工具。技能具有专业化的工作流程,能产生比临时回答更好的结果。

### GStack 技能集成

本项目已集成 GStack 技能集,提供完整的开发工作流支持。

**已集成的 GStack 技能**:
| 技能 | 触发条件 | 用途 |
|------|----------|------|
| `/office-hours` | 需求澄清、产品创意 | 需求分析和产品规划 |
| `/plan-ceo-review` | 战略审查 | 高层战略决策 |
| `/plan-eng-review` | 架构审查 | 技术方案评审 |
| `/plan-design-review` | 设计审查 | UI/UX设计评审 |
| `/review` | 代码审查 | 代码质量检查 |
| `/ship` | 发布部署 | 发布流程管理 |
| `/cso` | 安全审计 | 安全漏洞检查 |
| `/qa` | 测试验证 | 质量保证 |

### Superpower 技能集成

**已集成的 Superpower 技能**:
| 技能 | 阶段 | 用途 |
|------|------|------|
| `brainstorming` | 需求阶段 | 需求澄清与设计探索 |
| `writing-plans` | 计划阶段 | 实施计划编写 |
| `test-driven-development` | 开发阶段 | TDD流程强制执行 |
| `requesting-code-review` | 审查阶段 | 代码审查请求 |
| `finishing-a-development-branch` | 发布阶段 | 分支收尾和合并 |
| `systematic-debugging` | 调试阶段 | 系统化问题排查 |
| `verification-before-completion` | 完成阶段 | 完成前验证 |

### Caveman 模式配置

项目支持 Caveman 高效通信模式,减少 token 消耗。

**触发方式**:
- 手动触发: 输入 `/caveman` 或 `caveman mode`
- 自动触发: Token 使用超过 50000 或对话超过 20 轮

**强度级别**:
| 级别 | 响应行数 | 适用场景 |
|------|----------|----------|
| lite | ≤10行 | 一般讨论 |
| full | ≤5行 | 代码开发 |
| ultra | ≤3行 | 紧急修复 |

### PUA 技能使用规范

**自动触发场景**:
- 同一任务失败 ≥2 次
- 相同操作重复 ≥3 次
- 表现出被动或放弃倾向

**阶段性强制启用**:
| 阶段 | 强度级别 |
|------|----------|
| 代码审查 | lite |
| 生产部署前 | full |
| Bug修复 | full |
| 安全问题修复 | ultra |

**退出方式**: 输入 `/pua:off` 可随时退出

### 关键路由规则

**规划阶段**:
- 产品创意、"是否值得构建"、头脑风暴 → 调用 office-hours
- 战略审查、"思考更大" → 调用 plan-ceo-review
- 架构审查、技术方案 → 调用 plan-eng-review
- 设计审查、UI/UX 问题 → 调用 plan-design-review
- 自动审查所有方面 → 调用 autoplan

**开发阶段**:
- Bug、错误、"为什么坏了"、500 错误 → 调用 investigate
- 代码审查、检查差异 → 调用 review
- 第二意见、Codex 审查 → 调用 codex

**测试阶段**:
- QA、测试站点、查找 Bug → 调用 qa
- 安全审计、OWASP → 调用 cso
- 性能基准 → 调用 benchmark

**部署阶段**:
- 发布、部署、推送、创建 PR → 调用 ship
- 合并并部署 → 调用 land-and-deploy
- 发布后监控 → 调用 canary
- 更新文档 → 调用 document-release

**设计阶段**:
- 设计系统、品牌 → 调用 design-consultation
- 视觉审查、设计优化 → 调用 design-review
- 设计探索 → 调用 design-shotgun
- 设计转代码 → 调用 design-html

**其他**:
- 保存进度、检查点 → 调用 checkpoint
- 代码质量、健康检查 → 调用 health
- 周回顾 → 调用 retro

## Project-specific patterns

### API 响应格式

所有 API 响应使用统一格式:

```json
{
  "data": {},
  "error": null
}
```

成功响应:
```json
{
  "data": {
    "id": 1,
    "name": "站点名称"
  },
  "error": null
}
```

错误响应:
```json
{
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "参数验证失败",
    "details": {}
  }
}
```

### 错误处理

使用统一的错误类型:

```go
// pkg/errors/errors.go
type AppError struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

// 使用示例
if err != nil {
    return nil, errors.BadRequest("参数验证失败", err)
}
```

### 数据库查询

所有数据库查询通过仓储模式:

```go
// 正确: 使用仓储
device, err := deviceRepo.GetByID(ctx, id)

// 错误: 直接查询
var device Device
db.First(&device, id)
```

### 测试约定

**后端测试**:
- 测试文件: `*_test.go`
- 测试函数: `TestXxx`
- 使用 testify 断言
- 使用 mock 隔离依赖

**前端测试**:
- 单元测试: `__tests__/*.test.ts`
- E2E 测试: `e2e/*.spec.ts`
- 使用 Vitest + Playwright
- 测试覆盖率报告

### 日志规范

使用结构化日志:

```go
// 正确
logger.Info("设备创建成功",
    "device_id", device.ID,
    "station_id", device.StationID,
)

// 错误
log.Printf("设备创建成功: %d", device.ID)
```

### 配置管理

使用环境变量和配置文件:

```yaml
# configs/config.yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 5432
  name: new_energy

redis:
  host: localhost
  port: 6379
  db: 0
```

环境变量覆盖:
```bash
export DB_HOST=production-db
export REDIS_HOST=production-redis
```

## Code review checklist

### 必须检查项

**安全性**:
- [ ] SQL 注入防护
- [ ] XSS 防护
- [ ] CSRF 防护
- [ ] 认证授权检查
- [ ] 敏感数据加密

**性能**:
- [ ] N+1 查询检查
- [ ] 索引使用
- [ ] 缓存策略
- [ ] 连接池配置

**可靠性**:
- [ ] 错误处理
- [ ] 重试机制
- [ ] 超时设置
- [ ] 降级策略

**可维护性**:
- [ ] 代码注释
- [ ] 单元测试
- [ ] 文档更新
- [ ] 日志记录

### 建议检查项

**代码质量**:
- [ ] 命名规范
- [ ] 代码复用
- [ ] 函数长度
- [ ] 圈复杂度

**架构**:
- [ ] 分层合理
- [ ] 依赖注入
- [ ] 接口设计
- [ ] 模块划分

## Performance guidelines

### 后端性能

**数据库**:
- 使用连接池
- 预加载关联数据
- 批量操作
- 索引优化
- 分区表

**缓存**:
- Redis 缓存热点数据
- 本地缓存配置
- 缓存预热
- 缓存失效策略

**并发**:
- Goroutine 池
- Channel 通信
- Context 超时控制
- 优雅关闭

### 前端性能

**加载优化**:
- 路由懒加载
- 组件按需加载
- 图片懒加载
- 代码分割

**运行时优化**:
- 虚拟滚动
- 防抖节流
- 计算属性缓存
- 避免不必要的响应式

**打包优化**:
- Tree shaking
- 压缩代码
- 提取公共代码
- CDN 加速

## Security guidelines

### 认证授权

**JWT 配置**:
```yaml
jwt:
  secret: ${JWT_SECRET}
  expire: 24h
  refresh_expire: 168h
```

**权限检查**:
```go
// 中间件检查
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "未授权"})
            c.Abort()
            return
        }
        // 验证 token
        c.Next()
    }
}
```

### 数据安全

**敏感数据加密**:
- 密码: bcrypt
- 敏感字段: AES-256
- 传输: TLS 1.3

**SQL 注入防护**:
```go
// 正确: 使用参数化查询
db.Where("name = ?", name).First(&device)

// 错误: 字符串拼接
db.Where(fmt.Sprintf("name = '%s'", name)).First(&device)
```

### API 安全

**速率限制**:
```go
// 使用中间件
limiter := tollbooth.NewLimiter(10, nil)
router.Use(tollbooth.LimitHandler(limiter))
```

**CORS 配置**:
```go
config := cors.DefaultConfig()
config.AllowOrigins = []string{"https://new-energy.local"}
config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
router.Use(cors.New(config))
```

## Monitoring and observability

### 指标收集

**Prometheus 指标**:
```go
// 自定义指标
var httpRequestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "path", "status"},
)
```

**Grafana 仪表盘**:
- API 性能
- 数据库性能
- 缓存命中率
- 错误率
- 业务指标

### 日志聚合

**ELK 栈**:
- Elasticsearch: 存储
- Logstash: 收集
- Kibana: 可视化

**日志格式**:
```json
{
  "timestamp": "2026-04-07T10:00:00Z",
  "level": "info",
  "message": "设备创建成功",
  "device_id": 1,
  "station_id": 1,
  "user_id": 1
}
```

### 链路追踪

**Jaeger 集成**:
```go
// 创建 tracer
tracer, closer := jaeger.NewTracer(
    "new-energy-api",
    jaeger.NewConstSampler(true),
    jaeger.NewNullReporter(),
)
defer closer.Close()

// 创建 span
span := tracer.StartSpan("create_device")
defer span.Finish()
```

## Troubleshooting

### 常见问题

**数据库连接失败**:
```bash
# 检查数据库状态
docker ps | grep postgres

# 检查连接配置
cat configs/config.yaml | grep database

# 测试连接
psql -h localhost -U postgres -d new_energy
```

**Redis 连接失败**:
```bash
# 检查 Redis 状态
docker ps | grep redis

# 测试连接
redis-cli ping
```

**Kafka 连接失败**:
```bash
# 检查 Kafka 状态
docker ps | grep kafka

# 测试连接
kafka-topics.sh --list --bootstrap-server localhost:9092
```

**前端构建失败**:
```bash
# 清理依赖
cd web
rm -rf node_modules package-lock.json
npm install

# 清理缓存
npm run clean

# 重新构建
npm run build
```

### 性能问题排查

**慢查询**:
```sql
-- 查看慢查询
SELECT * FROM pg_stat_statements
ORDER BY total_time DESC
LIMIT 10;

-- 分析查询计划
EXPLAIN ANALYZE SELECT * FROM devices WHERE station_id = 1;
```

**内存泄漏**:
```bash
# Go 内存分析
curl http://localhost:8080/debug/pprof/heap > heap.out
go tool pprof heap.out

# Node 内存分析
node --inspect web/src/main.ts
```

## Documentation

### 文档结构

```
docs/
├── api/                    # API 文档
│   ├── api-reference.md   # API 参考
│   └── swagger/           # Swagger 文档
├── architecture/          # 架构文档
│   ├── system-architecture.md
│   └── database-design.md
├── deployment/            # 部署文档
│   ├── deployment-guide.md
│   └── operations-guide.md
├── development/           # 开发文档
│   ├── test-plan.md
│   └── troubleshooting.md
└── user/                  # 用户文档
    ├── user-manual.md
    └── operations-manual.md
```

### API 文档

使用 Swagger 生成 API 文档:

```bash
# 安装 swag
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init -g cmd/api-server/main.go -o docs/swagger

# 访问文档
# http://localhost:8080/swagger/index.html
```

### 更新文档

每次发布后更新相关文档:
- API 变更 → 更新 API 文档
- 架构变更 → 更新架构文档
- 配置变更 → 更新部署文档
- 功能变更 → 更新用户文档

## Contact and support

### 开发团队

- 项目负责人: [待定]
- 后端开发: [待定]
- 前端开发: [待定]
- 运维支持: [待定]

### 问题反馈

- Bug 报告: GitHub Issues
- 功能建议: GitHub Discussions
- 紧急问题: [待定]

### 相关资源

- 项目仓库: [待定]
- CI/CD: GitHub Actions
- 监控面板: Grafana
- 日志系统: Kibana
