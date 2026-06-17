# Gin Web 框架

## 基本信息

- **名称**：Gin
- **类型**：Go Web 框架
- **开发者**：Gin Community
- **首次发布**：2014年
- **当前版本**：v1.9+
- **许可证**：MIT
- **GitHub**：https://github.com/gin-gonic/gin

## 核心特性

### 性能特性
- **高性能**：基于 httprouter，性能优异
- **低内存占用**：优化的路由匹配
- **零分配路由**：路由匹配无内存分配

### 功能特性
- **RESTful API**：完整的 REST 支持
- **路由分组**：支持路由分组和嵌套
- **中间件**：强大的中间件系统
- **参数解析**：路径参数、查询参数、表单参数
- **请求验证**：内置参数验证
- **响应渲染**：JSON、XML、HTML、Protobuf 等
- **错误处理**：统一的错误处理机制

## 在本项目中的应用

### API 路径设计
- **统一前缀**：`/api/v1`
- **版本管理**：通过路径前缀管理版本
- **资源命名**：使用复数形式，如 `/alarm-rules`

### 路由结构
```go
// cmd/api-server/app.go
api := r.Group("/api/v1")
{
    // 告警规则路由
    alarmRules := api.Group("/alarm-rules")
    {
        alarmRules.GET("", alarmRuleHandler.ListAlarmRules)
        alarmRules.POST("", alarmRuleHandler.CreateAlarmRule)
        alarmRules.GET("/:id", alarmRuleHandler.GetAlarmRule)
        alarmRules.PUT("/:id", alarmRuleHandler.UpdateAlarmRule)
        alarmRules.DELETE("/:id", alarmRuleHandler.DeleteAlarmRule)
        alarmRules.POST("/:id/enable", alarmRuleHandler.EnableAlarmRule)
        alarmRules.POST("/:id/disable", alarmRuleHandler.DisableAlarmRule)
    }
}
```

### 中间件使用
- **日志中间件**：记录请求日志
- **恢复中间件**：panic 恢复
- **CORS 中间件**：跨域支持
- **认证中间件**：JWT 认证
- **权限中间件**：权限验证

## 开发规范

### Handler 编写
- **函数签名**：`func(c *gin.Context)`
- **参数获取**：使用 `c.Param()`, `c.Query()`, `c.ShouldBind()`
- **响应返回**：使用 `c.JSON()`, `c.XML()` 等
- **错误处理**：使用 `c.AbortWithStatusJSON()`

### 参数验证
- **结构体标签**：使用 `binding:"required"` 等标签
- **自定义验证**：支持自定义验证器
- **验证错误**：统一的错误响应格式

## 最佳实践

### 路由设计
- **RESTful 风格**：遵循 REST 设计原则
- **版本控制**：API 版本通过路径管理
- **资源命名**：使用名词复数形式
- **HTTP 方法**：正确使用 GET、POST、PUT、DELETE

### 错误处理
- **统一错误格式**：标准化错误响应
- **HTTP 状态码**：正确使用状态码
- **错误日志**：记录详细错误信息

### 性能优化
- **路由分组**：合理组织路由
- **中间件顺序**：注意中间件执行顺序
- **响应压缩**：启用 Gzip 压缩
- **静态资源**：合理配置静态文件服务

## 学习资源

- **官方文档**：https://gin-gonic.com/docs/
- **GitHub 仓库**：https://github.com/gin-gonic/gin
- **示例代码**：https://github.com/gin-gonic/examples
