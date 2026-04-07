# Harness 层实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立 Harness 层基础设施，实现 AI Agent 行为约束和验证机制

**Architecture:** 在 DDD 分层架构中新增 Harness 层， Validator/Verifier/Constraint/Monitor 组件，约束 AI 辅助开发过程

**Tech Stack:** Go 1.24, testify, Prometheus, Grafana

---

## 文件结构

### 新增文件
- `pkg/harness/validator.go` - 输入验证器
- `pkg/harness/verifier.go` - 输出验证器
- `pkg/harness/constraint.go` - 约束条件
- `pkg/harness/monitor.go` - 运行监控
- `pkg/harness/snapshot.go` - 快照测试
- `pkg/harness/result.go` - 结果类型

- `pkg/harness/harness.go` - 主入口文件

### 测试文件
- `pkg/harness/validator_test.go`
- `pkg/harness/verifier_test.go`
- `pkg/harness/constraint_test.go`
- `pkg/harness/monitor_test.go`

---

## Task 1: 实现 Validator 组件

**Files:**
- Create: `pkg/harness/validator.go`
- Create: `pkg/harness/validator_test.go`

- [ ] **Step 1.1: 编写 Validator 接口定义**

```go
package harness

// Validator 输入验证器接口
type Validator interface {
    Validate(ctx context.Context, input interface{}) error
    ValidateAsync(ctx context.Context, input interface{}) (<-chan ValidationResult, error)
}

// ValidationResult 验证结果
type ValidationResult struct {
    Valid    bool
    Errors   []error
    Warnings []string
}
```

- [ ] **Step 1.2: 编写 Validator 单元测试**

```go
package harness_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestValidator_Interface(t *testing.T) {
    // 测试接口定义是否存在
    var validator Validator = &MockValidator{}
    t.Implement(&validator, nil)
    
    assert.NotNil(t, validator)
}
```

- [ ] **Step 1.3: 运行测试验证接口定义**

Run: `cd e:\ai_work\new-energy-monitoring`
 go test ./pkg/harness/... -v`
Expected: PASS (接口定义测试通过)

- [ ] **Step 1.4: 提交代码**

```bash
git add pkg/harness/validator.go pkg/harness/validator_test.go
git commit -m "feat: add Validator interface and unit test"
```

---

## Task 2: 实现 Verifier 组件
**Files:**
- Create: `pkg/harness/verifier.go`
- Create: `pkg/harness/verifier_test.go`

- [ ] **Step 2.1: 编写 Verifier 接口定义**

```go
package harness

// Verifier 输出验证器接口
type Verifier interface {
    Verify(ctx context.Context, expected, actual interface{}) (bool, error)
    Snapshot(ctx context.Context, target interface{}) ([]byte, error)
}
```

- [ ] **Step 2.2: 编写 Verifier 单元测试**

```go
package harness_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestVerifier_Interface(t *testing.T) {
    var verifier Verifier = &MockVerifier{}
    t.Implement(&verifier, nil)
    
    assert.NotNil(t, verifier)
}
```

- [ ] **Step 2.3: 运行测试验证接口定义**

Run: `cd e:\ai_work\new-energy-monitoring && go test ./pkg/harness/... -v`
Expected: PASS

- [ ] **Step 2.4: 提交代码**

```bash
git add pkg/harness/verifier.go pkg/harness/verifier_test.go
git commit -m "feat: add Verifier interface and unit test"
```

---

## Task 3: 实现 Constraint 组件
**Files:**
- Create: `pkg/harness/constraint.go`
- Create: `pkg/harness/constraint_test.go`

- [ ] **Step 3.1: 编写 Constraint 接口定义**

```go
package harness

// Constraint 约束条件接口
type Constraint interface {
    Check(ctx context.Context, target interface{}) (bool, error)
    Apply(ctx context.Context, target interface{}) error
}
```

- [ ] **Step 3.2: 编写 Constraint 单元测试**

```go
package harness_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestConstraint_Interface(t *testing.T) {
    var constraint Constraint = &MockConstraint{}
    t.Implement(&constraint, nil)
    
    assert.NotNil(t, constraint)
}
```

- [ ] **Step 3.3: 运行测试验证接口定义**

Run: `cd e:\ai_work\new-energy-monitoring && go test ./pkg/harness/... -v`
Expected: PASS

- [ ] **Step 3.4: 提交代码**

```bash
git add pkg/harness/constraint.go pkg/harness/constraint_test.go
git commit -m "feat: add Constraint interface and unit test"
```

---

## Task 4: 实现 Monitor 组件
**Files:**
- Create: `pkg/harness/monitor.go`
- Create: `pkg/harness/monitor_test.go`

- [ ] **Step 4.1: 编写 Monitor 接口定义**

```go
package harness

import "context"

// Monitor 运行监控器接口
type Monitor interface {
    Record(ctx context.Context, metric string, value float) error
    GetMetrics(ctx context.Context, pattern string) ([]Metric, error)
}

// Metric 指标数据
type Metric struct {
    Name      string
    Value     float
    Timestamp int64
    Labels    map[string]string
}
```

- [ ] **Step 4.2: 编写 Monitor 单元测试**

```go
package harness_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMonitor_Interface(t *testing.T) {
    var monitor Monitor = &MockMonitor{}
    t.Implement(&monitor, nil)
    
    assert.NotNil(t, monitor)
}
```

- [ ] **Step 4.3: 运行测试验证接口定义**

Run: `cd e:\ai_work\new-energy-monitoring && go test ./pkg/harness/... -v`
Expected: PASS

- [ ] **Step 4.4: 提交代码**

```bash
git add pkg/harness/monitor.go pkg/harness/monitor_test.go
git commit -m "feat: add Monitor interface and unit test"
```

---

## Task 5: 实现 Snapshot 组件
**Files:**
- Create: `pkg/harness/snapshot.go`
- Create: `pkg/harness/snapshot_test.go`

- [ ] **Step 5.1: 编写 Snapshot 功能**

```go
package harness

import "encoding/json"

// Snapshot 快照测试结果
type Snapshot struct {
    ID        string
    Data      []byte
    CreatedAt int64
    Checksum  string
}

// SnapshotManager 快照管理器
type SnapshotManager struct {
    storage map[string]*Snapshot
}

func NewSnapshotManager() *SnapshotManager {
    return &SnapshotManager{
        storage: make(map[string]*Snapshot),
    }
}

func (sm *SnapshotManager) Save(id string, data []byte) error {
    snapshot := &Snapshot{
        ID:        id,
        Data:      data,
        CreatedAt: time.Now().Unix(),
        Checksum: sm.calculateChecksum(data),
    }
    sm.storage[id] = snapshot
    return nil
}

func (sm *SnapshotManager) Load(id string) (*Snapshot, error) {
    snapshot, exists := sm.storage[id]
    if !exists {
        return nil, ErrSnapshotNotFound
    }
    return snapshot, nil
}

func (sm *SnapshotManager) calculateChecksum(data []byte) string {
    // 简化的校验和计算
    sum := 0
    for _, b := range data {
        sum += int(b)
    }
    return fmt.Sprintf("%x", sum)
}

var ErrSnapshotNotFound = errors.New("snapshot not found")
```

- [ ] **Step 5.2: 编写 Snapshot 单元测试**

```go
package harness_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/new-energy-monitoring/pkg/harness"
)

func TestSnapshotManager_Save(t *testing.T) {
    sm := harness.NewSnapshotManager()
    data := []byte("test data")
    
    err := sm.Save("test-id", data)
    assert.NoError(t, err)
    
    snapshot, err := sm.Load("test-id")
    assert.NoError(t, err)
    assert.Equal(t, data, snapshot.Data)
}
```

- [ ] **Step 5.3: 运行测试**

Run: `cd e:\ai_work\new-energy-monitoring && go test ./pkg/harness/... -v`
Expected: PASS

- [ ] **Step 5.4: 提交代码**

```bash
git add pkg/harness/snapshot.go pkg/harness/snapshot_test.go
git commit -m "feat: add Snapshot component and unit tests"
```

---

## Task 6: 实现 Harness 主入口
**Files:**
- Create: `pkg/harness/harness.go`

- [ ] **Step 6.1: 编写 Harness 主入口**

```go
package harness

import "context"

// Harness 主入口
type Harness struct {
    validator  Validator
    verifier   Verifier
    constraint Constraint
    monitor    Monitor
    snapshot   *SnapshotManager
}

func NewHarness() *Harness {
    return &Harness{
        validator:  NewDefaultValidator(),
        verifier:   NewDefaultVerifier(),
        constraint: NewDefaultConstraint(),
        monitor:   NewDefaultMonitor(),
        snapshot:  NewSnapshotManager(),
    }
}

// Validate 执行验证
func (h *Harness) Validate(ctx context.Context, input interface{}) error {
    if err := h.validator.Validate(ctx, input); err != nil {
        return err
    }
    if err := h.constraint.Check(ctx, input); err != nil {
        return err
    }
    return nil
}

// Verify 执行验证
func (h *Harness) Verify(ctx context.Context, expected, actual interface{}) (bool, error) {
    return h.verifier.Verify(ctx, expected, actual)
}
```

- [ ] **Step 6.2: 运行测试**

Run: `cd e:\ai_work\new-energy-monitoring && go test ./pkg/harness/... -v`
Expected: PASS

- [ ] **Step 6.3: 提交代码**

```bash
git add pkg/harness/harness.go
git commit -m "feat: add Harness main entry point"
```

---

## 任务依赖关系

- Task 2 依赖 Task 1
- Task 3 依赖 Task 1
- Task 4 依赖 Task 1
- Task 5 依赖 Task 1
- Task 6 依赖 Task 2, Task 3, Task 4, Task 5

---

## 验收标准

1. 所有单元测试通过
2. 代码覆盖率 ≥ 80%
3. 所有接口定义清晰
4. 所有组件可独立使用

---

**计划完成时间**: 2026-04-07
