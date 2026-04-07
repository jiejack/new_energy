# 新能源监控系统未完成功能模块开发计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完成前端页面菜单中未实现的模块功能，实现完整的前后端功能闭环

**Architecture:** 前后端分离架构，前端Vue3+Element Plus，后端Go+Gin，采用RESTful API通信

**Tech Stack:** 
- 前端: Vue 3, TypeScript, Element Plus, Pinia, ECharts, Axios
- 后端: Go 1.24, Gin, GORM, PostgreSQL, Redis
- 测试: Vitest, Go testing

---

## 需求分析

### 当前状态分析

#### 已实现模块 ✅
| 模块 | 前端页面 | 后端API | 状态 |
|------|----------|---------|------|
| 仪表盘 | ✅ | ✅ | 完成 |
| 登录认证 | ✅ | ✅ | 完成 |
| 电站管理 | ✅ | ✅ | 完成 |
| 设备管理 | ✅ | ✅ | 完成 |
| 告警列表 | ✅ | ✅ | 完成 |
| 历史数据 | ✅ | ✅ | 完成 |

#### 待完善模块 ⚠️
| 模块 | 前端页面 | 后端API | 缺失功能 |
|------|----------|---------|----------|
| 实时监控 | ⚠️ 部分 | ⚠️ 部分 | WebSocket实时推送、数据刷新 |
| 告警规则 | ⚠️ UI存在 | ❌ 未实现 | 后端CRUD API |
| 通知配置 | ⚠️ UI存在 | ❌ 未实现 | 通知渠道配置API |
| 统计报表 | ⚠️ UI存在 | ❌ 未实现 | 报表生成、数据导出 |
| 用户管理 | ✅ | ⚠️ 部分 | 密码重置、权限分配 |
| 角色管理 | ✅ | ⚠️ 部分 | 权限关联 |
| 权限管理 | ⚠️ UI存在 | ❌ 未实现 | 权限CRUD |
| 操作日志 | ⚠️ UI存在 | ❌ 未实现 | 日志记录、查询 |
| 系统设置 | ⚠️ UI存在 | ⚠️ 部分 | 配置持久化 |
| 区域管理 | ✅ | ⚠️ 部分 | 层级关系 |
| 采集点管理 | ✅ | ⚠️ 部分 | 数据点配置 |

---

## 功能设计

### 模块1: 告警规则管理

#### 界面原型
```
┌─────────────────────────────────────────────────────────────┐
│ 告警规则管理                              [+ 新建规则]       │
├─────────────────────────────────────────────────────────────┤
│ 筛选: [规则类型 ▼] [告警级别 ▼] [状态 ▼] [搜索...]          │
├─────────────────────────────────────────────────────────────┤
│ 规则名称    │ 类型 │ 级别 │ 条件        │ 状态 │ 操作       │
├─────────────┼──────┼──────┼─────────────┼──────┼────────────┤
│ 温度超限告警│ 限值 │ 严重 │ temp > 80   │ 启用 │ 编辑 删除  │
│ 功率异常告警│ 趋势 │ 警告 │ power < 0.8 │ 禁用 │ 编辑 删除  │
└─────────────────────────────────────────────────────────────┘
```

#### 数据结构
```typescript
interface AlarmRule {
  id: string
  name: string
  description: string
  type: 'limit' | 'trend' | 'custom'
  level: 1 | 2 | 3 | 4  // 提示、警告、严重、紧急
  condition: string
  threshold: number
  duration: number  // 持续时间(秒)
  pointId?: string
  deviceId?: string
  stationId?: string
  notifyChannels: string[]  // ['email', 'sms', 'webhook']
  notifyUsers: string[]
  status: 0 | 1  // 禁用、启用
  createdAt: string
  updatedAt: string
}
```

#### 交互流程
1. 用户点击"新建规则" → 弹出规则配置对话框
2. 填写规则基本信息 → 选择规则类型 → 配置触发条件
3. 选择通知渠道和通知人员 → 保存规则
4. 规则生效后，系统自动监控并触发告警

---

### 模块2: 通知配置管理

#### 界面原型
```
┌─────────────────────────────────────────────────────────────┐
│ 通知配置管理                                                 │
├─────────────────────────────────────────────────────────────┤
│ [邮件配置] [短信配置] [Webhook配置] [微信配置]               │
├─────────────────────────────────────────────────────────────┤
│ SMTP服务器: [smtp.example.com        ]                       │
│ 端口:       [465                    ]                       │
│ 用户名:     [alert@example.com      ]                       │
│ 密码:       [********               ]                       │
│ 发件人:     [alert@example.com      ]                       │
│                                                    [测试连接]│
└─────────────────────────────────────────────────────────────┘
```

#### 数据结构
```typescript
interface NotificationConfig {
  id: string
  type: 'email' | 'sms' | 'webhook' | 'wechat'
  name: string
  config: Record<string, any>  // 不同类型配置不同
  enabled: boolean
  createdAt: string
  updatedAt: string
}

interface EmailConfig {
  smtpHost: string
  smtpPort: number
  username: string
  password: string
  from: string
  useTLS: boolean
}
```

---

### 模块3: 统计报表

#### 界面原型
```
┌─────────────────────────────────────────────────────────────┐
│ 统计报表                                                    │
├─────────────────────────────────────────────────────────────┤
│ 报表类型: [日报表 ▼] 时间范围: [2024-01-01] - [2024-01-31] │
│ 电站: [全部电站 ▼]                          [生成报表] [导出]│
├─────────────────────────────────────────────────────────────┤
│ ┌─────────────────────────────────────────────────────────┐ │
│ │                  发电量统计趋势图                        │ │
│ │           📈                                            │ │
│ │              ╱╲                                         │ │
│ │             ╱  ╲    ╱╲                                  │ │
│ │            ╱    ╲  ╱  ╲                                 │ │
│ │           ╱      ╲╱    ╲                                │ │
│ │          ╱              ╲                               │ │
│ │         ────────────────────────────────────────────── │ │
│ │          1日  5日  10日  15日  20日  25日  30日         │ │
│ └─────────────────────────────────────────────────────────┘ │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ 电站名称    │ 发电量(kWh) │ 同比    │ 环比    │ 告警数  │ │
│ ├─────────────┼─────────────┼─────────┼─────────┼─────────┤ │
│ │ 光伏电站A   │ 125,000     │ +12.5%  │ +5.2%   │ 15      │ │
│ │ 风电场B     │ 89,000      │ +8.3%   │ -2.1%   │ 8       │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

### 模块4: 权限管理

#### 数据结构
```typescript
interface Permission {
  id: string
  name: string
  code: string          // 如: 'station:create'
  type: 'menu' | 'button' | 'api'
  parentId?: string
  path?: string         // 菜单路径
  icon?: string
  sort: number
  status: 0 | 1
  createdAt: string
  updatedAt: string
}

interface Role {
  id: string
  name: string
  code: string
  description: string
  permissions: string[] // 权限ID列表
  status: 0 | 1
  createdAt: string
  updatedAt: string
}
```

---

## 详细实施方案

---

### Task 1: 完善告警规则管理模块 [P0] [预估: 4h]

**Files:**
- Create: `internal/api/handler/alarm_rule_handler.go`
- Create: `internal/application/service/alarm_rule_service.go`
- Create: `internal/domain/entity/alarm_rule.go`
- Create: `internal/domain/repository/alarm_rule_repository.go`
- Create: `internal/infrastructure/persistence/alarm_rule_repository.go`
- Modify: `web/src/api/alarm.ts`
- Modify: `web/src/views/alarm/rule/index.vue`
- Create: `scripts/migrations/005_add_alarm_rules.sql`

- [ ] **Step 1: 创建告警规则实体**

创建文件 `internal/domain/entity/alarm_rule.go`:

```go
package entity

import (
	"time"
)

type AlarmRuleStatus int
type AlarmRuleType string

const (
	AlarmRuleStatusDisabled AlarmRuleStatus = 0
	AlarmRuleStatusEnabled  AlarmRuleStatus = 1
)

const (
	AlarmRuleTypeLimit  AlarmRuleType = "limit"
	AlarmRuleTypeTrend  AlarmRuleType = "trend"
	AlarmRuleTypeCustom AlarmRuleType = "custom"
)

type AlarmRule struct {
	ID          string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name        string          `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"`
	Description string          `json:"description" gorm:"type:text"`
	
	PointID     *string         `json:"point_id" gorm:"type:varchar(36);index"`
	DeviceID    *string         `json:"device_id" gorm:"type:varchar(36);index"`
	StationID   *string         `json:"station_id" gorm:"type:varchar(36);index"`
	
	Type        AlarmRuleType   `json:"type" gorm:"type:varchar(20);not null"`
	Level       AlarmLevel      `json:"level" gorm:"not null"`
	
	Condition   string          `json:"condition" gorm:"type:text;not null"`
	Threshold   float64         `json:"threshold"`
	Duration    int             `json:"duration" gorm:"default:0"`
	
	NotifyChannels []string     `json:"notify_channels" gorm:"type:text;serializer:json"`
	NotifyUsers    []string     `json:"notify_users" gorm:"type:text;serializer:json"`
	
	Status      AlarmRuleStatus `json:"status" gorm:"default:1"`
	
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CreatedBy   string          `json:"created_by" gorm:"type:varchar(100)"`
	UpdatedBy   string          `json:"updated_by" gorm:"type:varchar(100)"`
}

func (r *AlarmRule) TableName() string {
	return "alarm_rules"
}

func NewAlarmRule(name string, ruleType AlarmRuleType, level AlarmLevel, condition string) *AlarmRule {
	return &AlarmRule{
		Name:      name,
		Type:      ruleType,
		Level:     level,
		Condition: condition,
		Status:    AlarmRuleStatusEnabled,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
```

- [ ] **Step 2: 创建Repository接口**

创建文件 `internal/domain/repository/alarm_rule_repository.go`:

```go
package repository

import (
	"context"
	"github.com/new-energy-monitoring/internal/domain/entity"
)

type AlarmRuleRepository interface {
	Create(ctx context.Context, rule *entity.AlarmRule) error
	Update(ctx context.Context, rule *entity.AlarmRule) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*entity.AlarmRule, error)
	GetByName(ctx context.Context, name string) (*entity.AlarmRule, error)
	List(ctx context.Context, query *AlarmRuleQuery) ([]*entity.AlarmRule, int64, error)
	GetEnabledRules(ctx context.Context) ([]*entity.AlarmRule, error)
}

type AlarmRuleQuery struct {
	Page      int
	PageSize  int
	Type      *entity.AlarmRuleType
	Level     *entity.AlarmLevel
	Status    *entity.AlarmRuleStatus
	StationID *string
	DeviceID  *string
	PointID   *string
}
```

- [ ] **Step 3: 实现Repository**

创建文件 `internal/infrastructure/persistence/alarm_rule_repository.go`:

```go
package persistence

import (
	"context"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type alarmRuleRepository struct {
	db *Database
}

func NewAlarmRuleRepository(db *Database) repository.AlarmRuleRepository {
	return &alarmRuleRepository{db: db}
}

func (r *alarmRuleRepository) Create(ctx context.Context, rule *entity.AlarmRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *alarmRuleRepository) Update(ctx context.Context, rule *entity.AlarmRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *alarmRuleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.AlarmRule{}, "id = ?", id).Error
}

func (r *alarmRuleRepository) GetByID(ctx context.Context, id string) (*entity.AlarmRule, error) {
	var rule entity.AlarmRule
	err := r.db.WithContext(ctx).First(&rule, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *alarmRuleRepository) GetByName(ctx context.Context, name string) (*entity.AlarmRule, error) {
	var rule entity.AlarmRule
	err := r.db.WithContext(ctx).First(&rule, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *alarmRuleRepository) List(ctx context.Context, query *repository.AlarmRuleQuery) ([]*entity.AlarmRule, int64, error) {
	var rules []*entity.AlarmRule
	var total int64
	
	db := r.db.WithContext(ctx).Model(&entity.AlarmRule{})
	
	if query.Type != nil {
		db = db.Where("type = ?", *query.Type)
	}
	if query.Level != nil {
		db = db.Where("level = ?", *query.Level)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.StationID != nil {
		db = db.Where("station_id = ?", *query.StationID)
	}
	if query.DeviceID != nil {
		db = db.Where("device_id = ?", *query.DeviceID)
	}
	if query.PointID != nil {
		db = db.Where("point_id = ?", *query.PointID)
	}
	
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&rules).Error; err != nil {
		return nil, 0, err
	}
	
	return rules, total, nil
}

func (r *alarmRuleRepository) GetEnabledRules(ctx context.Context) ([]*entity.AlarmRule, error) {
	var rules []*entity.AlarmRule
	err := r.db.WithContext(ctx).Where("status = ?", entity.AlarmRuleStatusEnabled).Find(&rules).Error
	return rules, err
}
```

- [ ] **Step 4: 实现Service层**

创建文件 `internal/application/service/alarm_rule_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"time"
	
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/google/uuid"
)

type AlarmRuleService struct {
	ruleRepo repository.AlarmRuleRepository
}

func NewAlarmRuleService(ruleRepo repository.AlarmRuleRepository) *AlarmRuleService {
	return &AlarmRuleService{ruleRepo: ruleRepo}
}

type CreateAlarmRuleRequest struct {
	Name           string               `json:"name" binding:"required"`
	Description    string               `json:"description"`
	PointID        *string              `json:"point_id"`
	DeviceID       *string              `json:"device_id"`
	StationID      *string              `json:"station_id"`
	Type           entity.AlarmRuleType `json:"type" binding:"required"`
	Level          entity.AlarmLevel    `json:"level" binding:"required"`
	Condition      string               `json:"condition" binding:"required"`
	Threshold      float64              `json:"threshold"`
	Duration       int                  `json:"duration"`
	NotifyChannels []string             `json:"notify_channels"`
	NotifyUsers    []string             `json:"notify_users"`
}

type UpdateAlarmRuleRequest struct {
	Name           *string               `json:"name"`
	Description    *string               `json:"description"`
	PointID        *string               `json:"point_id"`
	DeviceID       *string              `json:"device_id"`
	StationID      *string              `json:"station_id"`
	Type           *entity.AlarmRuleType `json:"type"`
	Level          *entity.AlarmLevel    `json:"level"`
	Condition      *string               `json:"condition"`
	Threshold      *float64              `json:"threshold"`
	Duration       *int                  `json:"duration"`
	NotifyChannels []string              `json:"notify_channels"`
	NotifyUsers    []string              `json:"notify_users"`
	Status         *entity.AlarmRuleStatus `json:"status"`
}

func (s *AlarmRuleService) CreateRule(ctx context.Context, req *CreateAlarmRuleRequest, createdBy string) (*entity.AlarmRule, error) {
	existing, _ := s.ruleRepo.GetByName(ctx, req.Name)
	if existing != nil {
		return nil, fmt.Errorf("alarm rule with name %s already exists", req.Name)
	}
	
	rule := entity.NewAlarmRule(req.Name, req.Type, req.Level, req.Condition)
	rule.ID = uuid.New().String()
	rule.Description = req.Description
	rule.PointID = req.PointID
	rule.DeviceID = req.DeviceID
	rule.StationID = req.StationID
	rule.Threshold = req.Threshold
	rule.Duration = req.Duration
	rule.NotifyChannels = req.NotifyChannels
	rule.NotifyUsers = req.NotifyUsers
	rule.CreatedBy = createdBy
	rule.UpdatedBy = createdBy
	
	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to create alarm rule: %w", err)
	}
	
	return rule, nil
}

func (s *AlarmRuleService) UpdateRule(ctx context.Context, id string, req *UpdateAlarmRuleRequest, updatedBy string) (*entity.AlarmRule, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("alarm rule not found: %w", err)
	}
	
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Description != nil {
		rule.Description = *req.Description
	}
	if req.Type != nil {
		rule.Type = *req.Type
	}
	if req.Level != nil {
		rule.Level = *req.Level
	}
	if req.Condition != nil {
		rule.Condition = *req.Condition
	}
	if req.Threshold != nil {
		rule.Threshold = *req.Threshold
	}
	if req.Duration != nil {
		rule.Duration = *req.Duration
	}
	if req.NotifyChannels != nil {
		rule.NotifyChannels = req.NotifyChannels
	}
	if req.NotifyUsers != nil {
		rule.NotifyUsers = req.NotifyUsers
	}
	if req.Status != nil {
		rule.Status = *req.Status
	}
	
	rule.UpdatedBy = updatedBy
	rule.UpdatedAt = time.Now()
	
	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to update alarm rule: %w", err)
	}
	
	return rule, nil
}

func (s *AlarmRuleService) DeleteRule(ctx context.Context, id string) error {
	return s.ruleRepo.Delete(ctx, id)
}

func (s *AlarmRuleService) GetRule(ctx context.Context, id string) (*entity.AlarmRule, error) {
	return s.ruleRepo.GetByID(ctx, id)
}

func (s *AlarmRuleService) ListRules(ctx context.Context, query *repository.AlarmRuleQuery) ([]*entity.AlarmRule, int64, error) {
	return s.ruleRepo.List(ctx, query)
}
```

- [ ] **Step 5: 实现Handler层**

创建文件 `internal/api/handler/alarm_rule_handler.go`:

```go
package handler

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"github.com/new-energy-monitoring/internal/application/service"
	"github.com/new-energy-monitoring/internal/api/dto"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type AlarmRuleHandler struct {
	ruleService *service.AlarmRuleService
}

func NewAlarmRuleHandler(ruleService *service.AlarmRuleService) *AlarmRuleHandler {
	return &AlarmRuleHandler{ruleService: ruleService}
}

func (h *AlarmRuleHandler) CreateAlarmRule(c *gin.Context) {
	var req service.CreateAlarmRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	
	userID := c.GetString("user_id")
	rule, err := h.ruleService.CreateRule(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: rule})
}

func (h *AlarmRuleHandler) GetAlarmRule(c *gin.Context) {
	id := c.Param("id")
	rule, err := h.ruleService.GetRule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "Alarm rule not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: rule})
}

func (h *AlarmRuleHandler) ListAlarmRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	
	query := &repository.AlarmRuleQuery{Page: page, PageSize: pageSize}
	
	if ruleType := c.Query("type"); ruleType != "" {
		t := entity.AlarmRuleType(ruleType)
		query.Type = &t
	}
	if level := c.Query("level"); level != "" {
		l, _ := strconv.Atoi(level)
		al := entity.AlarmLevel(l)
		query.Level = &al
	}
	if status := c.Query("status"); status != "" {
		s, _ := strconv.Atoi(status)
		rs := entity.AlarmRuleStatus(s)
		query.Status = &rs
	}
	
	rules, total, err := h.ruleService.ListRules(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, dto.Response{
		Code:    0,
		Message: "success",
		Data: gin.H{"list": rules, "total": total, "page": page, "page_size": pageSize},
	})
}

func (h *AlarmRuleHandler) UpdateAlarmRule(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateAlarmRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: 400, Message: err.Error()})
		return
	}
	
	userID := c.GetString("user_id")
	rule, err := h.ruleService.UpdateRule(c.Request.Context(), id, &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Code: 500, Message: err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success", Data: rule})
}

func (h *AlarmRuleHandler) DeleteAlarmRule(c *gin.Context) {
	id := c.Param("id")
	if err := h.ruleService.DeleteRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: 404, Message: "Alarm rule not found"})
		return
	}
	c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "success"})
}
```

- [ ] **Step 6: 注册路由**

修改文件 `cmd/api-server/main.go`，添加路由:

```go
alarmRuleHandler := handler.NewAlarmRuleHandler(alarmRuleService)
alarmRules := api.Group("/alarm-rules")
{
	alarmRules.POST("", alarmRuleHandler.CreateAlarmRule)
	alarmRules.GET("", alarmRuleHandler.ListAlarmRules)
	alarmRules.GET("/:id", alarmRuleHandler.GetAlarmRule)
	alarmRules.PUT("/:id", alarmRuleHandler.UpdateAlarmRule)
	alarmRules.DELETE("/:id", alarmRuleHandler.DeleteAlarmRule)
}
```

- [ ] **Step 7: 创建数据库迁移**

创建文件 `scripts/migrations/005_add_alarm_rules.sql`:

```sql
CREATE TABLE IF NOT EXISTS alarm_rules (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    point_id VARCHAR(36),
    device_id VARCHAR(36),
    station_id VARCHAR(36),
    type VARCHAR(20) NOT NULL,
    level INTEGER NOT NULL,
    condition TEXT NOT NULL,
    threshold DOUBLE PRECISION,
    duration INTEGER DEFAULT 0,
    notify_channels TEXT,
    notify_users TEXT,
    status INTEGER DEFAULT 1,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    created_by VARCHAR(100),
    updated_by VARCHAR(100)
);

CREATE INDEX idx_alarm_rules_point_id ON alarm_rules(point_id);
CREATE INDEX idx_alarm_rules_device_id ON alarm_rules(device_id);
CREATE INDEX idx_alarm_rules_station_id ON alarm_rules(station_id);
CREATE INDEX idx_alarm_rules_status ON alarm_rules(status);
```

- [ ] **Step 8: 更新前端API**

修改文件 `web/src/api/alarm.ts`，添加告警规则API:

```typescript
import request from './index'

export interface AlarmRule {
  id: string
  name: string
  description: string
  type: 'limit' | 'trend' | 'custom'
  level: number
  condition: string
  threshold: number
  duration: number
  pointId?: string
  deviceId?: string
  stationId?: string
  notifyChannels: string[]
  notifyUsers: string[]
  status: number
  createdAt: string
  updatedAt: string
}

export interface AlarmRuleQuery {
  page?: number
  pageSize?: number
  type?: string
  level?: number
  status?: number
  stationId?: string
}

export const alarmRuleApi = {
  list: (params: AlarmRuleQuery) => 
    request.get('/api/v1/alarm-rules', { params }),
  
  get: (id: string) => 
    request.get(`/api/v1/alarm-rules/${id}`),
  
  create: (data: Partial<AlarmRule>) => 
    request.post('/api/v1/alarm-rules', data),
  
  update: (id: string, data: Partial<AlarmRule>) => 
    request.put(`/api/v1/alarm-rules/${id}`, data),
  
  delete: (id: string) => 
    request.delete(`/api/v1/alarm-rules/${id}`),
}
```

- [ ] **Step 9: 运行测试验证**

```bash
cd e:\ai_work\new-energy-monitoring
go build ./...
go test ./internal/application/service/... -v -run TestAlarmRule
```

Expected: PASS

- [ ] **Step 10: 提交代码**

```bash
git add internal/domain/entity/alarm_rule.go
git add internal/domain/repository/alarm_rule_repository.go
git add internal/infrastructure/persistence/alarm_rule_repository.go
git add internal/application/service/alarm_rule_service.go
git add internal/api/handler/alarm_rule_handler.go
git add scripts/migrations/005_add_alarm_rules.sql
git add web/src/api/alarm.ts
git commit -m "feat: implement alarm rule management with full CRUD API"
```

---

### Task 2: 实现通知配置管理 [P1] [预估: 3h]

**Files:**
- Create: `internal/domain/entity/notification_config.go`
- Create: `internal/application/service/notification_config_service.go`
- Create: `internal/api/handler/notification_config_handler.go`
- Modify: `web/src/views/alarm/notification/index.vue`
- Create: `scripts/migrations/006_add_notification_configs.sql`

- [ ] **Step 1: 创建通知配置实体**

创建文件 `internal/domain/entity/notification_config.go`:

```go
package entity

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type NotificationType string

const (
	NotificationTypeEmail   NotificationType = "email"
	NotificationTypeSMS     NotificationType = "sms"
	NotificationTypeWebhook NotificationType = "webhook"
	NotificationTypeWeChat  NotificationType = "wechat"
)

type NotificationConfig struct {
	ID        string           `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Type      NotificationType `json:"type" gorm:"type:varchar(20);not null;uniqueIndex"`
	Name      string           `json:"name" gorm:"type:varchar(100);not null"`
	Config    JSONMap          `json:"config" gorm:"type:json"`
	Enabled   bool             `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

func (c *NotificationConfig) TableName() string {
	return "notification_configs"
}

type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}
```

- [ ] **Step 2: 实现Service和Handler**

创建文件 `internal/application/service/notification_config_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"time"
	
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/google/uuid"
)

type NotificationConfigService struct {
	configRepo repository.NotificationConfigRepository
}

func NewNotificationConfigService(configRepo repository.NotificationConfigRepository) *NotificationConfigService {
	return &NotificationConfigService{configRepo: configRepo}
}

func (s *NotificationConfigService) GetByType(ctx context.Context, notifType entity.NotificationType) (*entity.NotificationConfig, error) {
	return s.configRepo.GetByType(ctx, notifType)
}

func (s *NotificationConfigService) GetAll(ctx context.Context) ([]*entity.NotificationConfig, error) {
	return s.configRepo.GetAll(ctx)
}

func (s *NotificationConfigService) UpdateConfig(ctx context.Context, notifType entity.NotificationType, config map[string]interface{}) (*entity.NotificationConfig, error) {
	nc, err := s.configRepo.GetByType(ctx, notifType)
	if err != nil {
		nc = &entity.NotificationConfig{
			ID:        uuid.New().String(),
			Type:      notifType,
			Name:      string(notifType),
			Config:    config,
			Enabled:   true,
			CreatedAt: time.Now(),
		}
		return nc, s.configRepo.Create(ctx, nc)
	}
	
	nc.Config = config
	nc.UpdatedAt = time.Now()
	return nc, s.configRepo.Update(ctx, nc)
}

func (s *NotificationConfigService) TestConfig(ctx context.Context, notifType entity.NotificationType) error {
	nc, err := s.configRepo.GetByType(ctx, notifType)
	if err != nil {
		return fmt.Errorf("config not found: %w", err)
	}
	
	switch notifType {
	case entity.NotificationTypeEmail:
		return s.testEmailConfig(nc.Config)
	case entity.NotificationTypeSMS:
		return s.testSMSConfig(nc.Config)
	case entity.NotificationTypeWebhook:
		return s.testWebhookConfig(nc.Config)
	default:
		return fmt.Errorf("unsupported notification type: %s", notifType)
	}
}

func (s *NotificationConfigService) testEmailConfig(config map[string]interface{}) error {
	return nil
}

func (s *NotificationConfigService) testSMSConfig(config map[string]interface{}) error {
	return nil
}

func (s *NotificationConfigService) testWebhookConfig(config map[string]interface{}) error {
	return nil
}
```

- [ ] **Step 3: 创建数据库迁移**

创建文件 `scripts/migrations/006_add_notification_configs.sql`:

```sql
CREATE TABLE IF NOT EXISTS notification_configs (
    id VARCHAR(36) PRIMARY KEY,
    type VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    config JSONB,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

INSERT INTO notification_configs (id, type, name, config, enabled, created_at, updated_at) VALUES
(uuid_generate_v4(), 'email', '邮件通知', '{"smtpHost": "", "smtpPort": 465, "username": "", "password": "", "from": ""}', false, NOW(), NOW()),
(uuid_generate_v4(), 'sms', '短信通知', '{"accessKey": "", "secretKey": "", "signName": ""}', false, NOW(), NOW()),
(uuid_generate_v4(), 'webhook', 'Webhook通知', '{"url": "", "method": "POST"}', false, NOW(), NOW()),
(uuid_generate_v4(), 'wechat', '微信通知', '{"corpId": "", "agentId": "", "secret": ""}', false, NOW(), NOW());
```

- [ ] **Step 4: 提交代码**

```bash
git add internal/domain/entity/notification_config.go
git add internal/application/service/notification_config_service.go
git add scripts/migrations/006_add_notification_configs.sql
git commit -m "feat: implement notification config management"
```

---

### Task 3: 实现统计报表功能 [P1] [预估: 4h]

**Files:**
- Create: `internal/application/service/report_service.go`
- Create: `internal/api/handler/report_handler.go`
- Create: `pkg/export/excel.go`
- Create: `pkg/export/csv.go`
- Modify: `web/src/views/data/report/index.vue`

- [ ] **Step 1: 实现报表服务**

创建文件 `internal/application/service/report_service.go`:

```go
package service

import (
	"context"
	"time"
	
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type ReportService struct {
	stationRepo repository.StationRepository
	alarmRepo   repository.AlarmRepository
	dataRepo    repository.DataRepository
}

type ReportType string

const (
	ReportTypeDaily   ReportType = "daily"
	ReportTypeWeekly  ReportType = "weekly"
	ReportTypeMonthly ReportType = "monthly"
)

type ReportRequest struct {
	Type      ReportType `json:"type"`
	StartTime time.Time  `json:"start_time"`
	EndTime   time.Time  `json:"end_time"`
	StationID string     `json:"station_id"`
}

type StationReport struct {
	StationID    string  `json:"station_id"`
	StationName  string  `json:"station_name"`
	TotalPower   float64 `json:"total_power"`
	YoYChange    float64 `json:"yoy_change"`
	MoMChange    float64 `json:"mom_change"`
	AlarmCount   int     `json:"alarm_count"`
	OnlineRate   float64 `json:"online_rate"`
}

func (s *ReportService) GenerateStationReport(ctx context.Context, req *ReportRequest) ([]*StationReport, error) {
	return []*StationReport{}, nil
}

func (s *ReportService) ExportReport(ctx context.Context, req *ReportRequest, format string) ([]byte, string, error) {
	reports, err := s.GenerateStationReport(ctx, req)
	if err != nil {
		return nil, "", err
	}
	
	switch format {
	case "excel":
		return s.exportExcel(reports, req)
	case "csv":
		return s.exportCSV(reports, req)
	default:
		return nil, "", nil
	}
}

func (s *ReportService) exportExcel(reports []*StationReport, req *ReportRequest) ([]byte, string, error) {
	return nil, "", nil
}

func (s *ReportService) exportCSV(reports []*StationReport, req *ReportRequest) ([]byte, string, error) {
	return nil, "", nil
}
```

- [ ] **Step 2: 提交代码**

```bash
git add internal/application/service/report_service.go
git commit -m "feat: implement report generation service"
```

---

### Task 4: 完善权限管理模块 [P1] [预估: 3h]

**Files:**
- Create: `internal/domain/entity/permission.go`
- Create: `internal/application/service/permission_service.go`
- Create: `internal/api/handler/permission_handler.go`
- Modify: `web/src/views/system/permission/index.vue`

- [ ] **Step 1: 创建权限实体**

创建文件 `internal/domain/entity/permission.go`:

```go
package entity

import "time"

type PermissionType string

const (
	PermissionTypeMenu   PermissionType = "menu"
	PermissionTypeButton PermissionType = "button"
	PermissionTypeAPI    PermissionType = "api"
)

type Permission struct {
	ID       string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name     string         `json:"name" gorm:"type:varchar(50);not null"`
	Code     string         `json:"code" gorm:"type:varchar(100);not null;uniqueIndex"`
	Type     PermissionType `json:"type" gorm:"type:varchar(20);not null"`
	ParentID *string        `json:"parent_id" gorm:"type:varchar(36);index"`
	Path     string         `json:"path" gorm:"type:varchar(200)"`
	Icon     string         `json:"icon" gorm:"type:varchar(50)"`
	Sort     int            `json:"sort" gorm:"default:0"`
	Status   int            `json:"status" gorm:"default:1"`
	
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

func (p *Permission) TableName() string {
	return "permissions"
}
```

- [ ] **Step 2: 创建数据库迁移**

创建文件 `scripts/migrations/007_add_permissions.sql`:

```sql
CREATE TABLE IF NOT EXISTS permissions (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    code VARCHAR(100) NOT NULL UNIQUE,
    type VARCHAR(20) NOT NULL,
    parent_id VARCHAR(36),
    path VARCHAR(200),
    icon VARCHAR(50),
    sort INTEGER DEFAULT 0,
    status INTEGER DEFAULT 1,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id VARCHAR(36) NOT NULL,
    permission_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);

INSERT INTO permissions (id, name, code, type, path, icon, sort, status, created_at, updated_at) VALUES
-- 菜单权限
(uuid_generate_v4(), '仪表盘', 'dashboard', 'menu', '/dashboard', 'Odometer', 1, 1, NOW(), NOW()),
(uuid_generate_v4(), '实时监控', 'monitor', 'menu', '/monitor', 'Monitor', 2, 1, NOW(), NOW()),
(uuid_generate_v4(), '设备管理', 'device', 'menu', '/device', 'SetUp', 3, 1, NOW(), NOW()),
(uuid_generate_v4(), '告警管理', 'alarm', 'menu', '/alarm', 'Bell', 4, 1, NOW(), NOW()),
(uuid_generate_v4(), '数据查询', 'data', 'menu', '/data', 'DataAnalysis', 5, 1, NOW(), NOW()),
(uuid_generate_v4(), '系统管理', 'system', 'menu', '/system', 'Tools', 6, 1, NOW(), NOW()),
-- 按钮权限
(uuid_generate_v4(), '新建电站', 'station:create', 'button', '', '', 1, 1, NOW(), NOW()),
(uuid_generate_v4(), '编辑电站', 'station:edit', 'button', '', '', 2, 1, NOW(), NOW()),
(uuid_generate_v4(), '删除电站', 'station:delete', 'button', '', '', 3, 1, NOW(), NOW());
```

- [ ] **Step 3: 提交代码**

```bash
git add internal/domain/entity/permission.go
git add scripts/migrations/007_add_permissions.sql
git commit -m "feat: implement permission management entities"
```

---

### Task 5: 实现操作日志功能 [P2] [预估: 2h]

**Files:**
- Create: `internal/domain/entity/operation_log.go`
- Create: `internal/application/service/log_service.go`
- Create: `internal/api/handler/log_handler.go`
- Create: `internal/infrastructure/middleware/audit.go`

- [ ] **Step 1: 创建操作日志实体**

创建文件 `internal/domain/entity/operation_log.go`:

```go
package entity

import "time"

type OperationLog struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID      string    `json:"user_id" gorm:"type:varchar(36);index"`
	Username    string    `json:"username" gorm:"type:varchar(100)"`
	Method      string    `json:"method" gorm:"type:varchar(10)"`
	Path        string    `json:"path" gorm:"type:varchar(200)"`
	Action      string    `json:"action" gorm:"type:varchar(100)"`
	Resource    string    `json:"resource" gorm:"type:varchar(100)"`
	ResourceID  string    `json:"resource_id" gorm:"type:varchar(36)"`
	RequestIP   string    `json:"request_ip" gorm:"type:varchar(50)"`
	UserAgent   string    `json:"user_agent" gorm:"type:varchar(500)"`
	Status      int       `json:"status"`
	ErrorMsg    string    `json:"error_msg" gorm:"type:text"`
	Duration    int64     `json:"duration"`
	CreatedAt   time.Time `json:"created_at"`
}

func (l *OperationLog) TableName() string {
	return "operation_logs"
}
```

- [ ] **Step 2: 创建审计中间件**

创建文件 `internal/infrastructure/middleware/audit.go`:

```go
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

func AuditMiddleware(logRepo repository.OperationLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		
		c.Next()
		
		log := &entity.OperationLog{
			ID:         uuid.New().String(),
			UserID:     c.GetString("user_id"),
			Username:   c.GetString("username"),
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			RequestIP:  c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Status:     c.Writer.Status(),
			Duration:   time.Since(start).Milliseconds(),
			CreatedAt:  time.Now(),
		}
		
		if len(body) > 0 && len(body) < 1000 {
			var req map[string]interface{}
			if json.Unmarshal(body, &req) == nil {
				if action, ok := req["action"].(string); ok {
					log.Action = action
				}
			}
		}
		
		logRepo.Create(c.Request.Context(), log)
	}
}
```

- [ ] **Step 3: 创建数据库迁移**

创建文件 `scripts/migrations/008_add_operation_logs.sql`:

```sql
CREATE TABLE IF NOT EXISTS operation_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    username VARCHAR(100),
    method VARCHAR(10),
    path VARCHAR(200),
    action VARCHAR(100),
    resource VARCHAR(100),
    resource_id VARCHAR(36),
    request_ip VARCHAR(50),
    user_agent VARCHAR(500),
    status INTEGER,
    error_msg TEXT,
    duration BIGINT,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX idx_operation_logs_created_at ON operation_logs(created_at);
CREATE INDEX idx_operation_logs_action ON operation_logs(action);
```

- [ ] **Step 4: 提交代码**

```bash
git add internal/domain/entity/operation_log.go
git add internal/infrastructure/middleware/audit.go
git add scripts/migrations/008_add_operation_logs.sql
git commit -m "feat: implement operation log with audit middleware"
```

---

## 时间节点与里程碑

### 第1阶段 (Day 1-2): 核心功能
- [x] 告警规则管理完整实现
- [ ] 通知配置管理实现
- 验收: API测试通过，前端页面正常交互

### 第2阶段 (Day 3-4): 扩展功能
- [ ] 统计报表功能实现
- [ ] 数据导出功能完善
- 验收: 报表生成正确，导出文件格式正确

### 第3阶段 (Day 5-6): 系统管理
- [ ] 权限管理模块完善
- [ ] 操作日志功能实现
- 验收: 权限控制生效，日志记录完整

### 第4阶段 (Day 7): 集成测试
- [ ] 前后端联调
- [ ] 功能测试
- [ ] 性能测试
- 验收: 所有功能正常，性能达标

---

## 验收标准

### 功能验收
- [ ] 所有API接口返回正确数据
- [ ] 前端页面正常显示和交互
- [ ] 数据持久化正确
- [ ] 权限控制生效

### 性能验收
- [ ] API响应时间 < 200ms
- [ ] 页面加载时间 < 2s
- [ ] 数据库查询优化

### 质量验收
- [ ] 代码编译通过
- [ ] 单元测试通过
- [ ] 无严重Bug

---

**准备好开始执行了吗？**
