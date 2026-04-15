# 分层架构设计

## 概述

分层架构是一种常见的软件架构设计模式，将系统分为若干个层次，每个层次负责特定的功能。本项目采用经典的分层架构设计。

## 本项目的分层架构

### Entity 层（实体层）
- **职责**：定义数据模型和业务实体
- **位置**：`internal/domain/entity/`
- **内容**：
  - [Alarm](file:///workspace/internal/domain/entity/alarm.go) - 告警实体
  - [AlarmRule](file:///workspace/internal/domain/entity/alarm_rule.go) - 告警规则实体
  - [Station](file:///workspace/internal/domain/entity/station.go) - 电站实体
  - [Device](file:///workspace/internal/domain/entity/device.go) - 设备实体

### Repository 层（仓储层）
- **职责**：数据访问抽象，负责与数据库交互
- **位置**：`internal/domain/repository/`
- **实现**：`internal/infrastructure/persistence/`
- **作用**：
  - 封装数据访问逻辑
  - 提供统一的数据访问接口
  - 支持数据持久化和查询

### Service 层（服务层）
- **职责**：业务逻辑处理，实现核心业务功能
- **位置**：`internal/application/service/`
- **内容**：
  - [ReportService](file:///workspace/internal/application/service/report_service.go) - 报表服务
  - [AlarmRuleService](file:///workspace/internal/application/service/alarm_rule_service.go) - 告警规则服务
  - 其他业务服务

### Handler 层（处理器层）
- **职责**：HTTP 请求处理，将外部请求转换为内部服务调用
- **位置**：`internal/api/handler/`
- **内容**：
  - [ReportHandler](file:///workspace/internal/api/handler/report_handler.go) - 报表处理器
  - [AlarmRuleHandler](file:///workspace/internal/api/handler/alarm_rule_handler.go) - 告警规则处理器

## 分层的优势

### 1. 关注点分离
每个层次只关注自己的职责，降低了代码复杂度。

### 2. 可维护性
修改某一层不会影响其他层，便于代码维护和升级。

### 3. 可测试性
各层可以独立进行单元测试，提高测试覆盖率。

### 4. 可扩展性
可以在不修改现有代码的情况下，添加新的功能层次。

## 数据流向

```
外部请求
    ↓
Handler 层（HTTP 请求处理）
    ↓
Service 层（业务逻辑处理）
    ↓
Repository 层（数据访问）
    ↓
数据库
```

## 最佳实践

1. **层间通信**：上层只能调用下层，不能反向调用
2. **接口抽象**：使用接口定义层间契约
3. **依赖注入**：通过依赖注入实现层间解耦
4. **错误处理**：各层负责处理自己的错误，并向上传递
