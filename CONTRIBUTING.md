# Contributing Guide

感谢您有兴趣为新能源监控系统做出贡献！

## 开发流程

### 1. Fork 并克隆仓库

```bash
git clone https://github.com/YOUR_USERNAME/new_energy.git
cd new_energy
```

### 2. 创建开发分支

```bash
git checkout -b feature/your-feature-name
```

分支命名规范：
- `feat/xxx` - 新功能
- `fix/xxx` - Bug修复
- `docs/xxx` - 文档更新
- `refactor/xxx` - 代码重构
- `test/xxx` - 测试相关

### 3. 开发与测试

```bash
# 后端开发
go mod download
go run cmd/api-server/main.go

# 前端开发
cd web
npm install
npm run dev

# 运行测试
go test ./...
cd web && npm run test
```

### 4. 代码规范

#### Go 代码规范

- 遵循 [Effective Go](https://golang.org/doc/effective_go)
- 使用 `golangci-lint` 进行代码检查
- 单元测试覆盖率 ≥ 80%

```bash
# 代码检查
golangci-lint run

# 测试覆盖率
go test -coverprofile=coverage.out ./...
```

#### 前端代码规范

- 遵循 Vue 3 + TypeScript 最佳实践
- 使用 ESLint + Prettier 格式化代码
- 组件命名使用 PascalCase

```bash
# 代码检查
npm run lint

# 代码格式化
npm run format

# 类型检查
npm run typecheck
```

### 5. 提交代码

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型说明**：
| 类型 | 说明 |
|------|------|
| feat | 新功能 |
| fix | Bug修复 |
| docs | 文档更新 |
| style | 代码格式 |
| refactor | 重构 |
| perf | 性能优化 |
| test | 测试 |
| chore | 构建/工具 |

**示例**：
```
feat(alarm): add alarm rule management API

- Add CRUD operations for alarm rules
- Add validation for rule conditions
- Add unit tests

Closes #123
```

### 6. 创建 Pull Request

1. 推送到您的 Fork
2. 在 GitHub 上创建 Pull Request
3. 填写 PR 模板
4. 等待代码审查

## 代码审查标准

### 必须通过

- [ ] 所有测试通过
- [ ] 代码覆盖率不降低
- [ ] 无 ESLint/golangci-lint 错误
- [ ] 提交信息符合规范

### 建议满足

- [ ] 有适当的单元测试
- [ ] 有必要的文档更新
- [ ] 代码逻辑清晰易读
- [ ] 无重复代码

## 问题反馈

### Bug 报告

请使用 [GitHub Issues](https://github.com/jiejack/new_energy/issues)，包含：

1. 问题描述
2. 复现步骤
3. 期望行为
4. 实际行为
5. 环境信息

### 功能建议

欢迎提出新功能建议，请描述：

1. 功能需求
2. 使用场景
3. 预期效果

## 开发环境

### 必需工具

| 工具 | 版本 |
|------|------|
| Go | ≥1.24 |
| Node.js | ≥20.x |
| Docker | ≥24.0 |
| PostgreSQL | ≥15 |
| Redis | ≥7 |

### 推荐工具

| 工具 | 用途 |
|------|------|
| golangci-lint | Go代码检查 |
| ESLint | 前端代码检查 |
| Prettier | 代码格式化 |
| Git Hooks | 提交检查 |

## 许可证

提交代码即表示您同意将代码以 MIT 许可证授权给项目。

---

感谢您的贡献！
