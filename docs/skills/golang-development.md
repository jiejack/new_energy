# Go 后端开发技能

## 概述

本项目使用 Go 1.24+ 作为后端开发语言，基于 Gin 框架构建 RESTful API。

## 技术栈

### 核心技术
- **Go 1.24+**：Go 编程语言
- **Gin**：高性能 Web 框架
- **GORM**：ORM 库（可选）
- **Wire**：依赖注入工具
- **Excelize**：Excel 文件处理库

### 基础设施
- **PostgreSQL**：关系型数据库
- **TimescaleDB**：时序数据库（可选）
- **Redis**：缓存和会话存储
- **Kafka**：消息队列

## 项目结构

```
internal/
├── api/
│   ├── dto/          # 数据传输对象
│   └── handler/      # HTTP 处理器
├── application/
│   └── service/      # 业务服务
├── domain/
│   ├── entity/       # 领域实体
│   └── repository/   # 仓储接口
└── infrastructure/
    ├── persistence/  # 数据持久化实现
    ├── cache/        # 缓存实现
    └── config/       # 配置
```

## Gin 框架使用

### 路由定义
```go
router := gin.Default()

api := router.Group("/api/v1")
{
    api.GET("/alarm-rules", alarmRuleHandler.ListAlarmRules)
    api.POST("/alarm-rules", alarmRuleHandler.CreateAlarmRule)
    api.GET("/alarm-rules/:id", alarmRuleHandler.GetAlarmRule)
}
```

### 中间件使用
```go
router.Use(gin.Logger())
router.Use(gin.Recovery())
router.Use(middleware.Auth())
```

### 请求处理
```go
func (h *Handler) List(c *gin.Context) {
    var req ListRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Code: 400,
            Message: err.Error(),
        })
        return
    }

    result, err := h.service.List(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Code: 500,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, Response{
        Code: 0,
        Message: "success",
        Data: result,
    })
}
```

## 依赖注入（Wire）

### Provider 定义
```go
func NewAlarmRuleService(repo repository.AlarmRuleRepository) *service.AlarmRuleService {
    return service.NewAlarmRuleService(repo)
}

func NewAlarmRuleHandler(service *service.AlarmRuleService) *handler.AlarmRuleHandler {
    return handler.NewAlarmRuleHandler(service)
}
```

### Wire Set
```go
var ProviderSet = wire.NewSet(
    NewAlarmRuleRepository,
    NewAlarmRuleService,
    NewAlarmRuleHandler,
)
```

### 生成代码
```bash
go run github.com/google/wire/cmd/wire
```

## 分层架构实现

### Entity 层
```go
type AlarmRule struct {
    ID          string
    Name        string
    Description string
    Status      AlarmRuleStatus
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### Repository 层
```go
type AlarmRuleRepository interface {
    Create(ctx context.Context, rule *entity.AlarmRule) error
    GetByID(ctx context.Context, id string) (*entity.AlarmRule, error)
    List(ctx context.Context, query *AlarmRuleQuery) ([]*entity.AlarmRule, int64, error)
    Update(ctx context.Context, rule *entity.AlarmRule) error
    Delete(ctx context.Context, id string) error
}
```

### Service 层
```go
func (s *AlarmRuleService) CreateRule(ctx context.Context, req *CreateAlarmRuleRequest, userID string) (*entity.AlarmRule, error) {
    rule := &entity.AlarmRule{
        Name:        req.Name,
        Description: req.Description,
        Status:      entity.AlarmRuleStatusEnabled,
    }

    if err := s.repo.Create(ctx, rule); err != nil {
        return nil, err
    }

    return rule, nil
}
```

## 错误处理

### 自定义错误
```go
var (
    ErrRuleNotFound = errors.New("alarm rule not found")
    ErrInvalidRule  = errors.New("invalid alarm rule")
)
```

### 错误包装
```go
if err != nil {
    return nil, fmt.Errorf("failed to create rule: %w", err)
}
```

## 测试策略

### 单元测试
```go
func TestAlarmRuleService_CreateRule(t *testing.T) {
    repo := NewMockAlarmRuleRepository()
    service := NewAlarmRuleService(repo)

    req := &CreateAlarmRuleRequest{
        Name: "Test Rule",
    }

    rule, err := service.CreateRule(context.Background(), req, "user1")
    assert.NoError(t, err)
    assert.NotNil(t, rule)
}
```

### 集成测试
```go
func TestAlarmRuleHandler_List(t *testing.T) {
    // 设置测试环境
    // 发起 HTTP 请求
    // 验证响应
}
```

## 性能优化

### 1. 使用连接池
```go
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(time.Hour)
```

### 2. Redis 缓存
```go
func (s *Service) GetRule(ctx context.Context, id string) (*entity.AlarmRule, error) {
    cacheKey := fmt.Sprintf("rule:%s", id)
    
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        return cached, nil
    }

    rule, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    s.cache.Set(ctx, cacheKey, rule, time.Hour)
    return rule, nil
}
```

### 3. 并发处理
```go
func (s *Service) ProcessBatch(ctx context.Context, items []*Item) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(items))

    for _, item := range items {
        wg.Add(1)
        go func(i *Item) {
            defer wg.Done()
            if err := s.processItem(ctx, i); err != nil {
                errChan <- err
            }
        }(item)
    }

    wg.Wait()
    close(errChan)

    for err := range errChan {
        if err != nil {
            return err
        }
    }

    return nil
}
```

## 相关资源

- [Go 官方文档](https://go.dev/doc/)
- [Gin 框架文档](https://gin-gonic.com/docs/)
- [Wire 文档](https://github.com/google/wire)
- [Excelize 文档](https://xuri.me/excelize/)
