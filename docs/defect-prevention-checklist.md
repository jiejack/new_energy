# 缺陷预防检查清单

## 文档信息

| 项目 | 内容 |
|------|------|
| 文档版本 | v1.0 |
| 创建日期 | 2026-04-07 |
| 目标 | 预防缺陷，提升代码质量 |
| 适用范围 | 全生命周期质量保障 |

## 使用说明

本检查清单应在以下场景使用：
- 代码编写阶段：确保代码质量
- 代码审查阶段：指导审查重点
- 测试阶段：验证测试完整性
- 发布阶段：确保发布质量

---

## 1. 代码编写阶段检查清单

### 1.1 编码规范检查

#### 1.1.1 命名规范

- [ ] 变量名使用有意义的名称，避免单字母变量（循环变量除外）
- [ ] 函数名清晰表达功能意图
- [ ] 常量使用大写字母和下划线（如 `MAX_RETRY_COUNT`）
- [ ] 接口名遵循语言规范（Go: `-er` 后缀，C#: `I` 前缀）
- [ ] 避免使用缩写和拼音命名
- [ ] 布尔变量使用 `is/has/can` 等前缀
- [ ] 集合类型使用复数形式

**示例**:
```go
// Good
var isActive bool
var userCount int
var stationList []Station
func calculateTotalPrice(items []Item) float64

// Bad
var a bool
var n int
var sl []Station
func calc(items []Item) float64
```

#### 1.1.2 代码结构

- [ ] 函数长度不超过 50 行（复杂逻辑除外）
- [ ] 文件长度不超过 500 行
- [ ] 嵌套层级不超过 4 层
- [ ] 遵循单一职责原则
- [ ] 避免重复代码（DRY原则）
- [ ] 使用早返回减少嵌套
- [ ] 复杂逻辑拆分为多个函数

**示例**:
```go
// Good: 早返回
func processUser(user *User) error {
    if user == nil {
        return errors.New("user is nil")
    }
    if !user.IsActive {
        return errors.New("user is inactive")
    }
    // 处理逻辑
    return nil
}

// Bad: 深层嵌套
func processUser(user *User) error {
    if user != nil {
        if user.IsActive {
            // 处理逻辑
        } else {
            return errors.New("user is inactive")
        }
    } else {
        return errors.New("user is nil")
    }
    return nil
}
```

#### 1.1.3 注释规范

- [ ] 复杂逻辑添加注释说明
- [ ] 公共函数添加文档注释
- [ ] 避免无意义的注释
- [ ] 注释与代码保持同步更新
- [ ] TODO 注释包含责任人和时间
- [ ] 避免注释掉的代码

**示例**:
```go
// Good
// CalculateTotalPrice 计算订单总价
// 参数:
//   - items: 订单项列表
//   - discount: 折扣率（0-1）
// 返回: 订单总价
func CalculateTotalPrice(items []Item, discount float64) float64 {
    // 使用累加器计算基础价格
    var total float64
    for _, item := range items {
        total += item.Price * float64(item.Quantity)
    }
    // 应用折扣
    return total * (1 - discount)
}

// Bad
// 计算价格
func calc(items []Item, d float64) float64 {
    var t float64
    for _, item := range items {
        t += item.Price * float64(item.Quantity) // 累加
    }
    return t * (1 - d)
}
```

#### 1.1.4 错误处理

- [ ] 所有错误都有适当的处理
- [ ] 错误信息清晰明确，包含上下文
- [ ] 避免空 `catch` 块或忽略错误
- [ ] 错误向上传递时添加上下文信息
- [ ] 使用自定义错误类型区分错误类型
- [ ] 关键操作添加错误日志

**示例**:
```go
// Good
func getUserByID(id int64) (*User, error) {
    user, err := db.FindUserByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to find user by id %d: %w", id, err)
    }
    return user, nil
}

// Bad
func getUserByID(id int64) (*User, error) {
    user, err := db.FindUserByID(id)
    if err != nil {
        return nil, err // 缺少上下文
    }
    return user, nil
}
```

#### 1.1.5 安全编码

- [ ] 所有输入参数都进行验证
- [ ] 输出数据进行适当的编码
- [ ] SQL 查询使用参数化查询
- [ ] 敏感数据加密存储
- [ ] 避免硬编码密钥和密码
- [ ] 使用安全的随机数生成器
- [ ] 避免在日志中记录敏感信息

**示例**:
```go
// Good
func getUser(username string) (*User, error) {
    // 输入验证
    if username == "" {
        return nil, errors.New("username cannot be empty")
    }
    if len(username) > 50 {
        return nil, errors.New("username too long")
    }

    // 参数化查询
    var user User
    err := db.QueryRow("SELECT * FROM users WHERE username = $1", username).Scan(&user)
    return &user, err
}

// Bad
func getUser(username string) (*User, error) {
    // SQL注入风险
    query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", username)
    var user User
    err := db.QueryRow(query).Scan(&user)
    return &user, err
}
```

### 1.2 常见缺陷预防

#### 1.2.1 空指针异常

- [ ] 访问对象属性前检查 `nil/null`
- [ ] 使用 Optional 类型（Java/TypeScript）
- [ ] 初始化所有指针和引用类型变量
- [ ] 使用空对象模式避免空指针
- [ ] 链式调用时检查中间对象

**示例**:
```go
// Good
func getUserName(user *User) string {
    if user == nil {
        return "Unknown"
    }
    return user.Name
}

// 或使用空对象模式
type User struct {
    Name string
}

var NilUser = &User{Name: "Unknown"}

func getUserName(user *User) string {
    if user == nil {
        user = NilUser
    }
    return user.Name
}
```

#### 1.2.2 数组越界

- [ ] 访问数组/切片前检查长度
- [ ] 使用安全的集合操作方法
- [ ] 循环条件正确（避免 off-by-one 错误）
- [ ] 使用 range 循环代替索引访问

**示例**:
```go
// Good
func getFirstItem(items []Item) (Item, error) {
    if len(items) == 0 {
        return Item{}, errors.New("items is empty")
    }
    return items[0], nil
}

// Good: 使用 range
for i, item := range items {
    // 安全访问
}

// Bad
func getFirstItem(items []Item) Item {
    return items[0] // 可能越界
}
```

#### 1.2.3 并发问题

- [ ] 共享资源正确加锁
- [ ] 避免死锁（锁的顺序一致）
- [ ] 使用线程安全的数据结构
- [ ] 正确处理竞态条件
- [ ] 使用 channel 进行协程通信
- [ ] 避免在锁内调用外部函数

**示例**:
```go
// Good
type SafeCounter struct {
    mu    sync.RWMutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Get() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.count
}

// Bad: 竞态条件
type Counter struct {
    count int
}

func (c *Counter) Increment() {
    c.count++ // 非原子操作
}
```

#### 1.2.4 资源泄漏

- [ ] 文件、连接等资源正确关闭
- [ ] 使用 `defer`/`finally` 确保资源释放
- [ ] 大对象及时释放引用
- [ ] 使用对象池管理资源
- [ ] 检查资源泄漏（使用工具）

**示例**:
```go
// Good
func readFile(filename string) ([]byte, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close() // 确保关闭

    return io.ReadAll(file)
}

// Bad
func readFile(filename string) ([]byte, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    // 忘记关闭文件
    return io.ReadAll(file)
}
```

#### 1.2.5 性能问题

- [ ] 避免在循环中创建对象
- [ ] 使用缓存减少重复计算
- [ ] 数据库查询优化（索引、批量操作）
- [ ] 批量操作代替循环操作
- [ ] 避免过早优化
- [ ] 使用性能分析工具定位瓶颈

**示例**:
```go
// Good
func processItems(items []Item) []Result {
    results := make([]Result, 0, len(items)) // 预分配容量
    for _, item := range items {
        results = append(results, processItem(item))
    }
    return results
}

// Bad
func processItems(items []Item) []Result {
    var results []Result
    for _, item := range items {
        results = append(results, processItem(item)) // 频繁扩容
    }
    return results
}
```

---

## 2. 代码审查阶段检查清单

### 2.1 功能正确性

- [ ] 代码实现了需求的功能
- [ ] 边界条件处理正确
- [ ] 异常情况处理完整
- [ ] 业务逻辑正确无误
- [ ] 与现有代码无冲突

### 2.2 代码质量

- [ ] 代码可读性强
- [ ] 命名清晰准确
- [ ] 结构合理清晰
- [ ] 无重复代码
- [ ] 无过度设计

### 2.3 性能影响

- [ ] 无明显的性能问题
- [ ] 数据库查询合理
- [ ] 内存使用合理
- [ ] 并发处理正确
- [ ] 无性能退化

### 2.4 安全性

- [ ] 无安全漏洞
- [ ] 输入验证完整
- [ ] 权限控制正确
- [ ] 敏感数据处理得当
- [ ] 无信息泄露风险

### 2.5 可维护性

- [ ] 代码易于理解
- [ ] 代码易于修改
- [ ] 代码易于测试
- [ ] 文档完整清晰
- [ ] 依赖关系清晰

### 2.6 测试覆盖

- [ ] 单元测试完整
- [ ] 测试用例充分
- [ ] 测试覆盖率高
- [ ] 测试可维护性好
- [ ] 边界情况测试完整

---

## 3. 测试阶段检查清单

### 3.1 单元测试检查

#### 3.1.1 测试覆盖

- [ ] 所有公共方法都有测试
- [ ] 边界条件测试完整
- [ ] 异常情况测试完整
- [ ] 正常情况测试完整
- [ ] 特殊场景测试完整

#### 3.1.2 测试质量

- [ ] 测试用例独立，无依赖
- [ ] 测试可重复执行
- [ ] 测试命名清晰
- [ ] 断言充分且准确
- [ ] 测试意图明确

#### 3.1.3 测试数据

- [ ] 使用 Mock 隔离外部依赖
- [ ] 测试数据准备充分
- [ ] 测试数据清理完整
- [ ] 避免硬编码测试数据
- [ ] 测试数据具有代表性

#### 3.1.4 测试执行

- [ ] 测试执行时间合理（< 5分钟）
- [ ] 测试结果稳定可靠
- [ ] 测试报告清晰易读
- [ ] 失败信息准确有用

**示例**:
```go
// Good
func TestCalculateTotalPrice(t *testing.T) {
    tests := []struct {
        name     string
        items    []Item
        discount float64
        want     float64
    }{
        {
            name: "normal case",
            items: []Item{
                {Price: 100, Quantity: 2},
                {Price: 50, Quantity: 1},
            },
            discount: 0.1,
            want:     225, // (100*2 + 50*1) * 0.9
        },
        {
            name:     "empty items",
            items:    []Item{},
            discount: 0.1,
            want:     0,
        },
        {
            name:     "no discount",
            items:    []Item{{Price: 100, Quantity: 1}},
            discount: 0,
            want:     100,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalculateTotalPrice(tt.items, tt.discount)
            if got != tt.want {
                t.Errorf("CalculateTotalPrice() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 3.2 集成测试检查

#### 3.2.1 接口测试

- [ ] 所有 API 接口都有测试
- [ ] 请求参数验证完整
- [ ] 响应格式验证正确
- [ ] 错误码验证准确
- [ ] 权限验证完整

#### 3.2.2 数据库测试

- [ ] 事务正确性验证
- [ ] 并发访问正确性验证
- [ ] 数据一致性验证
- [ ] 迁移脚本测试
- [ ] 性能测试

#### 3.2.3 第三方集成

- [ ] 外部服务调用测试
- [ ] 超时处理测试
- [ ] 重试机制测试
- [ ] 降级处理测试
- [ ] 错误处理测试

### 3.3 E2E 测试检查

#### 3.3.1 测试场景

- [ ] 核心业务流程覆盖
- [ ] 用户操作路径完整
- [ ] 异常场景覆盖
- [ ] 边界情况覆盖
- [ ] 跨浏览器测试

#### 3.3.2 测试数据

- [ ] 测试环境数据准备
- [ ] 测试数据隔离
- [ ] 测试数据清理
- [ ] 数据状态验证

---

## 4. 发布阶段检查清单

### 4.1 发布前检查

#### 4.1.1 代码质量

- [ ] 代码审查通过（至少 1 人 Approve）
- [ ] 所有测试通过（单元、集成、E2E）
- [ ] 代码覆盖率达标（后端 ≥70%，前端 ≥80%）
- [ ] 无高危安全漏洞
- [ ] 无 Lint 错误

#### 4.1.2 配置检查

- [ ] 配置项正确无误
- [ ] 环境变量设置正确
- [ ] 密钥配置正确
- [ ] 数据库连接配置正确
- [ ] 第三方服务配置正确

#### 4.1.3 依赖检查

- [ ] 依赖版本锁定
- [ ] 无过期依赖
- [ ] 无冲突依赖
- [ ] 依赖安全检查通过

#### 4.1.4 文档检查

- [ ] API 文档更新
- [ ] 部署文档更新
- [ ] 变更日志更新
- [ ] 用户通知准备

#### 4.1.5 回滚准备

- [ ] 回滚脚本准备
- [ ] 数据库回滚脚本准备
- [ ] 回滚流程验证
- [ ] 回滚联系人确认

### 4.2 发布中检查

#### 4.2.1 金丝雀阶段（10% 流量）

- [ ] 错误率 < 0.1%
- [ ] P95 延迟 < 200ms
- [ ] P99 延迟 < 500ms
- [ ] 健康检查 100%
- [ ] 无异常日志
- [ ] 业务指标正常

#### 4.2.2 扩大阶段（50% 流量）

- [ ] 错误率 < 0.1%
- [ ] P95 延迟 < 200ms
- [ ] P99 延迟 < 500ms
- [ ] CPU 使用率 < 70%
- [ ] 内存使用率 < 80%
- [ ] 业务指标正常

#### 4.2.3 全量阶段（100% 流量）

- [ ] 错误率 < 0.1%
- [ ] P95 延迟 < 200ms
- [ ] P99 延迟 < 500ms
- [ ] 所有服务正常
- [ ] 业务指标正常
- [ ] 用户反馈正常

### 4.3 发布后检查

#### 4.3.1 功能验证

- [ ] 核心功能正常
- [ ] 新功能可用
- [ ] 旧功能不受影响
- [ ] 用户反馈正常
- [ ] 业务数据正常

#### 4.3.2 性能验证

- [ ] 响应时间达标
- [ ] 吞吐量达标
- [ ] 资源使用正常
- [ ] 无性能退化

#### 4.3.3 监控验证

- [ ] 监控数据正常
- [ ] 告警规则生效
- [ ] 日志收集正常
- [ ] 追踪数据正常

#### 4.3.4 文档验证

- [ ] 文档已更新
- [ ] 发布说明已发布
- [ ] 用户通知已发送

---

## 5. 运维阶段检查清单

### 5.1 日常运维检查

#### 5.1.1 服务健康

- [ ] 服务运行状态正常
- [ ] 健康检查通过
- [ ] 端口监听正常
- [ ] 进程数量正常

#### 5.1.2 资源监控

- [ ] CPU 使用率 < 80%
- [ ] 内存使用率 < 85%
- [ ] 磁盘使用率 < 80%
- [ ] 网络流量正常
- [ ] 连接数正常

#### 5.1.3 日志检查

- [ ] 无错误日志
- [ ] 无警告日志
- [ ] 审计日志正常
- [ ] 访问日志正常
- [ ] 日志大小正常

#### 5.1.4 备份检查

- [ ] 数据库备份正常
- [ ] 配置备份正常
- [ ] 日志备份正常
- [ ] 备份可恢复性验证

### 5.2 故障处理检查

#### 5.2.1 故障发现

- [ ] 监控告警及时
- [ ] 用户反馈渠道畅通
- [ ] 日志分析工具可用
- [ ] 故障通知机制正常

#### 5.2.2 故障定位

- [ ] 错误日志分析
- [ ] 链路追踪可用
- [ ] 性能分析工具可用
- [ ] 资源监控数据完整

#### 5.2.3 故障处理

- [ ] 快速恢复服务
- [ ] 根因分析完成
- [ ] 修复方案确定
- [ ] 预防措施制定

#### 5.2.4 故障复盘

- [ ] 故障报告编写
- [ ] 改进措施制定
- [ ] 知识沉淀完成
- [ ] 流程优化完成

---

## 6. 安全检查清单

### 6.1 认证授权

- [ ] 用户认证机制安全
- [ ] 密码存储安全（加密）
- [ ] Session 管理安全
- [ ] Token 管理安全
- [ ] 权限控制正确

### 6.2 输入验证

- [ ] 所有输入参数验证
- [ ] SQL 注入防护
- [ ] XSS 攻击防护
- [ ] CSRF 攻击防护
- [ ] 文件上传安全

### 6.3 数据安全

- [ ] 敏感数据加密存储
- [ ] 数据传输加密（HTTPS）
- [ ] 数据备份安全
- [ ] 数据脱敏处理
- [ ] 数据访问审计

### 6.4 接口安全

- [ ] API 认证机制
- [ ] API 限流保护
- [ ] API 访问控制
- [ ] API 日志审计
- [ ] API 版本管理

### 6.5 配置安全

- [ ] 无硬编码密钥
- [ ] 配置文件权限正确
- [ ] 敏感配置加密
- [ ] 环境隔离正确
- [ ] 密钥定期轮换

---

## 7. 性能检查清单

### 7.1 代码性能

- [ ] 无明显性能瓶颈
- [ ] 算法复杂度合理
- [ ] 内存使用合理
- [ ] CPU 使用合理
- [ ] 无内存泄漏

### 7.2 数据库性能

- [ ] 索引设计合理
- [ ] 查询优化完成
- [ ] 连接池配置合理
- [ ] 事务处理正确
- [ ] 批量操作优化

### 7.3 缓存性能

- [ ] 缓存策略合理
- [ ] 缓存命中率达标
- [ ] 缓存过期策略正确
- [ ] 缓存更新机制正确
- [ ] 缓存穿透防护

### 7.4 网络性能

- [ ] 接口响应时间达标
- [ ] 并发处理能力达标
- [ ] 网络带宽充足
- [ ] 连接复用合理
- [ ] 压缩传输启用

---

## 8. 文档检查清单

### 8.1 技术文档

- [ ] 架构设计文档完整
- [ ] API 文档完整准确
- [ ] 数据库设计文档完整
- [ ] 接口文档完整准确
- [ ] 部署文档完整准确

### 8.2 运维文档

- [ ] 运维手册完整
- [ ] 故障处理手册完整
- [ ] 监控配置文档完整
- [ ] 备份恢复文档完整
- [ ] 安全配置文档完整

### 8.3 用户文档

- [ ] 用户手册完整
- [ ] 操作指南清晰
- [ ] FAQ 文档完整
- [ ] 版本更新说明清晰
- [ ] 培训材料完整

---

## 9. 使用建议

### 9.1 检查清单使用时机

| 阶段 | 使用清单 | 频率 |
|------|----------|------|
| 代码编写 | 编码规范、常见缺陷预防 | 每次提交 |
| 代码审查 | 代码审查阶段检查清单 | 每次 PR |
| 测试 | 测试阶段检查清单 | 每次测试 |
| 发布 | 发布阶段检查清单 | 每次发布 |
| 运维 | 运维阶段检查清单 | 每日/每周 |

### 9.2 检查清单维护

- 定期更新检查清单（建议每季度）
- 根据实际问题和经验补充检查项
- 删除过时或不适用的检查项
- 团队共同维护和完善

### 9.3 检查清单效果评估

- 统计使用检查清单后的缺陷减少率
- 收集团队反馈，持续改进
- 对比使用前后的质量指标
- 定期评估检查清单的有效性

---

## 10. 附录

### 10.1 检查清单模板

```markdown
## [功能名称] 检查清单

### 功能正确性
- [ ] 功能实现正确
- [ ] 边界条件处理正确
- [ ] 异常情况处理完整

### 代码质量
- [ ] 代码可读性强
- [ ] 命名清晰准确
- [ ] 结构合理清晰

### 测试覆盖
- [ ] 单元测试完整
- [ ] 测试用例充分
- [ ] 测试覆盖率高

### 性能影响
- [ ] 无性能问题
- [ ] 资源使用合理

### 安全性
- [ ] 无安全漏洞
- [ ] 数据处理安全
```

### 10.2 常用工具

| 工具类别 | 工具名称 | 用途 |
|----------|----------|------|
| 代码检查 | golangci-lint | Go 代码检查 |
| 代码检查 | ESLint | JS/TS 代码检查 |
| 安全扫描 | Trivy | 容器漏洞扫描 |
| 安全扫描 | Gosec | Go 代码安全检查 |
| 性能分析 | pprof | Go 性能分析 |
| 测试覆盖 | go test -cover | Go 测试覆盖率 |
| 测试覆盖 | Vitest coverage | 前端测试覆盖率 |

### 10.3 参考资料

- [质量门禁文档](./quality-gates.md)
- [测试计划](./test-plan.md)
- [部署指南](./deployment-guide.md)
- [运维手册](./operations-manual.md)
- [安全审计报告](./security-audit-report.md)

---

**文档维护**: 本检查清单应根据项目实际情况持续更新和完善，建议每季度进行一次全面审查。
