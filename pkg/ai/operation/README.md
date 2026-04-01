# 智能操作服务 (Intelligent Operation Service)

## 概述

智能操作服务提供了新能源监控系统的智能操作能力，支持自然语言指令解析、操作执行、确认机制和审计日志等完整功能。

## 核心组件

### 1. 操作解析器 (OperationParser)

负责将自然语言指令解析为结构化的操作对象。

**功能特性：**
- 自然语言指令解析
- 操作类型识别（遥控、设点、调节、查询）
- 目标设备/测点识别
- 参数提取
- 安全校验
- 置信度评估

**使用示例：**

```go
parser := operation.NewOperationParser()
ctx := context.Background()

// 解析自然语言指令
result, err := parser.Parse(ctx, "启动逆变器INV-001，功率设置为500kW")
if err != nil {
    log.Fatal(err)
}

// 查看解析结果
for _, op := range result.Operations {
    fmt.Printf("操作类型: %s\n", op.Type)
    fmt.Printf("目标: %s\n", op.TargetID)
    fmt.Printf("参数: %v\n", op.Parameters)
    fmt.Printf("置信度: %.2f\n", op.Confidence)
}
```

**支持的操作类型：**
- `remote_control`: 遥控操作（启动、停止、合闸、分闸等）
- `setpoint`: 设点操作（设置参数值）
- `adjust`: 调节操作（增加、减少、调整）
- `query`: 查询操作（读取状态、数据）

### 2. 操作执行器 (OperationExecutor)

负责管理操作队列、执行操作、跟踪状态和处理重试。

**功能特性：**
- 操作队列管理
- 并发执行控制
- 执行状态跟踪
- 超时控制
- 自动重试机制
- 操作历史记录
- 回滚支持

**使用示例：**

```go
// 创建执行器
config := &operation.ExecutorConfig{
    QueueCapacity:   1000,
    MaxWorkers:      10,
    DefaultTimeout:  30 * time.Second,
    RetryDelay:      1 * time.Second,
    MaxRetryDelay:   30 * time.Second,
    EnablePriority:  true,
    HistoryCapacity: 10000,
}
executor := operation.NewOperationExecutor(config)

// 注册操作处理器
executor.RegisterHandler(operation.OperationTypeSetPoint, &SetPointHandler{})

// 启动执行器
ctx := context.Background()
executor.Start(ctx)
defer executor.Stop()

// 提交操作
op := &operation.ParsedOperation{
    ID:       "op-001",
    Type:     operation.OperationTypeSetPoint,
    Priority: operation.PriorityNormal,
    Action:   "set_value",
    TargetID: "POINT-001",
    Parameters: map[string]interface{}{
        "value": 100.0,
        "unit":  "kW",
    },
}

if err := executor.Submit(ctx, op); err != nil {
    log.Fatal(err)
}

// 等待完成
record, err := executor.WaitForCompletion(ctx, op.ID, 30*time.Second)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("操作状态: %s\n", record.Status)
```

**自定义处理器：**

```go
type SetPointHandler struct{}

func (h *SetPointHandler) Handle(ctx context.Context, op *operation.ParsedOperation) (interface{}, error) {
    // 实现设点逻辑
    pointID := op.TargetID
    value := op.Parameters["value"]
    
    // 调用设备接口设置值
    err := deviceClient.SetPoint(pointID, value)
    if err != nil {
        return nil, err
    }
    
    return map[string]interface{}{
        "point_id": pointID,
        "value":    value,
        "success":  true,
    }, nil
}

func (h *SetPointHandler) CanHandle(op *operation.ParsedOperation) bool {
    return op.Type == operation.OperationTypeSetPoint
}

func (h *SetPointHandler) Rollback(ctx context.Context, op *operation.ParsedOperation, result interface{}) error {
    // 实现回滚逻辑
    return nil
}
```

### 3. 确认管理器 (ConfirmationManager)

提供操作确认机制，确保关键操作的安全性。

**功能特性：**
- 两步确认流程
- 操作授权验证
- 确认码生成
- 过期时间管理
- 审计日志记录
- 回滚机制

**使用示例：**

```go
// 创建确认管理器
config := &operation.ConfirmationConfig{
    DefaultTimeout:    5 * time.Minute,
    EnableTwoStep:     true,
    CodeLength:        6,
    MaxPendingConfirm: 1000,
    AuditLogCapacity:  10000,
}
confirmer := operation.NewConfirmationManager(config, executor)

// 创建确认请求
op := &operation.ParsedOperation{
    ID:       "op-002",
    Type:     operation.OperationTypeRemoteControl,
    Action:   "switch_off",
    TargetID: "DEV-001",
    Constraints: &operation.OperationConstraints{
        RequireConfirm:  true,
        RequireAuth:     true,
        AuthLevel:       2,
        AllowRollback:   true,
    },
}

confirmRec, err := confirmer.CreateConfirmation(ctx, op)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("确认ID: %s\n", confirmRec.ID)
fmt.Printf("确认码: %s\n", confirmRec.ConfirmCode)

// 第一步确认
err = confirmer.FirstStepConfirm(ctx, confirmRec.ID, confirmRec.ConfirmCode, "operator-001", 2)
if err != nil {
    log.Fatal(err)
}

// 第二步确认（需要不同用户）
err = confirmer.SecondStepConfirm(ctx, confirmRec.ID, "supervisor-001", 3)
if err != nil {
    log.Fatal(err)
}

// 查看审计日志
logs := confirmer.GetAuditLogs(op.ID, 10)
for _, log := range logs {
    fmt.Printf("[%s] %s - %s\n", log.Timestamp, log.Action, log.UserID)
}
```

### 4. 操作API (OperationAPI)

提供统一的操作接口，集成所有组件。

**功能特性：**
- 统一的操作入口
- RESTful API支持
- 操作状态查询
- 历史记录查询
- 回滚操作
- 统计信息

**使用示例：**

```go
// 创建API
parser := operation.NewOperationParser()
executor := operation.NewOperationExecutor(config)
confirmer := operation.NewConfirmationManager(confirmConfig, executor)

api := operation.NewOperationAPI(parser, executor, confirmer, authChecker)

// 启动执行器
executor.Start(ctx)
defer executor.Stop()

// 提交操作
response, err := api.SubmitOperation(ctx, &operation.OperationRequest{
    Text:      "查询设备DEV-001的状态",
    UserID:    "user-001",
    IPAddress: "192.168.1.100",
    UserAgent: "web-client",
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("成功: %v\n", response.Success)
fmt.Printf("消息: %s\n", response.Message)

// 确认操作
confirmResp, err := api.ConfirmOperation(ctx, &operation.ConfirmRequest{
    ConfirmID:   response.Confirmations[0].ConfirmID,
    ConfirmCode: response.Confirmations[0].ConfirmCode,
    UserID:      "user-001",
    Step:        1,
})

// 查询状态
statusResp, err := api.GetOperationStatus(ctx, &operation.StatusRequest{
    OperationID: response.Operations[0].ID,
})

// 查询历史
historyResp, err := api.GetOperationHistory(ctx, &operation.HistoryRequest{
    Limit:  100,
    Status: "success",
})
```

## HTTP API接口

### 提交操作
```
POST /api/operation/submit
Content-Type: application/json

{
    "text": "启动设备DEV-001",
    "user_id": "user-001",
    "ip_address": "192.168.1.100",
    "user_agent": "web-client",
    "dry_run": false
}
```

### 确认操作
```
POST /api/operation/confirm
Content-Type: application/json

{
    "confirm_id": "CONF-xxx",
    "confirm_code": "123456",
    "user_id": "user-001",
    "step": 1
}
```

### 查询状态
```
GET /api/operation/status?operation_id=OP-xxx
```

### 查询历史
```
GET /api/operation/history?limit=100&status=success
```

### 回滚操作
```
POST /api/operation/rollback
Content-Type: application/json

{
    "operation_id": "OP-xxx",
    "user_id": "user-001",
    "reason": "操作错误"
}
```

### 获取待确认列表
```
GET /api/operation/pending
```

### 获取审计日志
```
GET /api/operation/audit?operation_id=OP-xxx&limit=100
```

### 获取统计信息
```
GET /api/operation/stats
```

## 安全机制

### 1. 操作确认
- 关键操作需要确认
- 支持两步确认流程
- 确认码验证
- 过期时间控制

### 2. 授权验证
- 用户权限检查
- 授权级别验证
- 受保护目标管理

### 3. 安全校验
- 危险操作识别
- 参数范围验证
- 操作风险评估

### 4. 审计日志
- 完整操作记录
- 用户行为追踪
- 时间戳记录
- IP地址记录

## 操作约束

```go
type OperationConstraints struct {
    MinValue        float64       // 最小值
    MaxValue        float64       // 最大值
    AllowedValues   []string      // 允许的值
    RequireConfirm  bool          // 是否需要确认
    RequireAuth     bool          // 是否需要授权
    AuthLevel       int           // 授权级别 (0-3)
    Timeout         time.Duration // 超时时间
    MaxRetries      int           // 最大重试次数
    AllowRollback   bool          // 是否允许回滚
    DryRun          bool          // 试运行模式
}
```

## 操作优先级

```go
const (
    PriorityLow      OperationPriority = 1  // 低优先级
    PriorityNormal   OperationPriority = 5  // 正常优先级
    PriorityHigh     OperationPriority = 8  // 高优先级
    PriorityCritical OperationPriority = 10 // 关键优先级
)
```

## 操作状态

```go
const (
    StatusPending    OperationStatus = "pending"    // 待执行
    StatusValidating OperationStatus = "validating" // 验证中
    StatusConfirmed  OperationStatus = "confirmed"  // 已确认
    StatusExecuting  OperationStatus = "executing"  // 执行中
    StatusSuccess    OperationStatus = "success"    // 成功
    StatusFailed     OperationStatus = "failed"     // 失败
    StatusTimeout    OperationStatus = "timeout"    // 超时
    StatusCancelled  OperationStatus = "cancelled"  // 已取消
    StatusRolledBack OperationStatus = "rolledback" // 已回滚
)
```

## 最佳实践

### 1. 错误处理
```go
response, err := api.SubmitOperation(ctx, req)
if err != nil {
    // 处理系统错误
    log.Error("系统错误", err)
    return
}

if !response.Success {
    // 处理业务错误
    log.Warn("操作失败", response.Message)
    return
}

// 检查警告
if len(response.Warnings) > 0 {
    for _, warning := range response.Warnings {
        log.Warn(warning)
    }
}
```

### 2. 确认流程
```go
// 1. 提交操作
response, _ := api.SubmitOperation(ctx, req)

// 2. 检查是否需要确认
for _, confirm := range response.Confirmations {
    if confirm.State == operation.ConfirmationStateConfirmed {
        // 不需要确认，已直接提交执行
        continue
    }
    
    // 3. 显示确认信息给用户
    fmt.Printf("操作需要确认: %s\n", confirm.Description)
    fmt.Printf("确认码: %s\n", confirm.ConfirmCode)
    
    // 4. 用户确认
    confirmResp, _ := api.ConfirmOperation(ctx, &operation.ConfirmRequest{
        ConfirmID:   confirm.ConfirmID,
        ConfirmCode: confirm.ConfirmCode,
        UserID:      currentUser.ID,
        Step:        1,
    })
    
    // 5. 检查是否需要第二步确认
    if confirmResp.NeedSecondStep {
        // 通知另一个用户进行第二步确认
        notifySecondUser(confirm.ConfirmID)
    }
}
```

### 3. 状态监控
```go
// 定期检查操作状态
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        status, _ := api.GetOperationStatus(ctx, &operation.StatusRequest{
            OperationID: opID,
        })
        
        if status.Status == operation.StatusSuccess ||
           status.Status == operation.StatusFailed {
            // 操作完成
            return
        }
        
        // 显示进度
        fmt.Printf("进度: %d%%\n", status.Record.Progress)
    }
}
```

## 性能优化

### 1. 执行器配置
- 根据系统负载调整 `MaxWorkers`
- 根据操作类型设置合理的 `Timeout`
- 启用 `EnablePriority` 支持优先级队列

### 2. 队列管理
- 设置合理的 `QueueCapacity` 避免内存溢出
- 监控队列长度，及时告警

### 3. 历史记录
- 设置合理的 `HistoryCapacity`
- 定期归档历史记录

## 监控指标

```go
// 获取执行器统计
stats := executor.GetStats()
fmt.Printf("队列长度: %d\n", stats.QueueLength)
fmt.Printf("运行中: %d\n", stats.RunningCount)
fmt.Printf("总执行数: %d\n", stats.TotalExecuted)
fmt.Printf("按状态统计: %v\n", stats.ByStatus)
fmt.Printf("按类型统计: %v\n", stats.ByType)

// 获取确认统计
confirmStats := confirmer.GetStats()
fmt.Printf("按状态统计: %v\n", confirmStats.ByState)
fmt.Printf("审计日志数: %d\n", confirmStats.TotalAuditLogs)
```

## 故障排查

### 1. 操作超时
- 检查设备连接状态
- 调整超时时间
- 查看执行日志

### 2. 确认失败
- 检查确认码是否正确
- 检查用户权限
- 检查确认是否过期

### 3. 队列堵塞
- 检查执行器是否正常启动
- 检查处理器是否正确注册
- 监控系统资源使用

## 扩展开发

### 添加新的操作类型

1. 定义操作类型常量
```go
const OperationTypeCustom OperationType = "custom"
```

2. 实现处理器
```go
type CustomHandler struct{}

func (h *CustomHandler) Handle(ctx context.Context, op *operation.ParsedOperation) (interface{}, error) {
    // 实现自定义逻辑
    return nil, nil
}

func (h *CustomHandler) CanHandle(op *operation.ParsedOperation) bool {
    return op.Type == OperationTypeCustom
}

func (h *CustomHandler) Rollback(ctx context.Context, op *operation.ParsedOperation, result interface{}) error {
    return nil
}
```

3. 注册处理器
```go
executor.RegisterHandler(OperationTypeCustom, &CustomHandler{})
```

4. 更新解析器
```go
parser.keywords[OperationTypeCustom] = []string{"自定义关键词"}
```

## 许可证

MIT License
