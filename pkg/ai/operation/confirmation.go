package operation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrConfirmationNotFound    = errors.New("confirmation not found")
	ErrConfirmationExpired     = errors.New("confirmation expired")
	ErrConfirmationAlreadyUsed = errors.New("confirmation already used")
	ErrInvalidConfirmCode      = errors.New("invalid confirmation code")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrInsufficientAuthLevel   = errors.New("insufficient authorization level")
)

// ConfirmationState 确认状态
type ConfirmationState string

const (
	ConfirmationStatePending    ConfirmationState = "pending"
	ConfirmationStateFirstStep  ConfirmationState = "first_step_confirmed"
	ConfirmationStateConfirmed  ConfirmationState = "confirmed"
	ConfirmationStateRejected   ConfirmationState = "rejected"
	ConfirmationStateExpired    ConfirmationState = "expired"
	ConfirmationStateCancelled  ConfirmationState = "cancelled"
)

// ConfirmationRecord 确认记录
type ConfirmationRecord struct {
	ID               string            `json:"id"`
	OperationID      string            `json:"operation_id"`
	Operation        *ParsedOperation  `json:"operation"`
	State            ConfirmationState `json:"state"`
	ConfirmCode      string            `json:"confirm_code"`
	RequiredAuthLevel int              `json:"required_auth_level"`
	
	// 两步确认
	FirstConfirmBy   string     `json:"first_confirm_by,omitempty"`
	FirstConfirmAt   *time.Time `json:"first_confirm_at,omitempty"`
	SecondConfirmBy  string     `json:"second_confirm_by,omitempty"`
	SecondConfirmAt  *time.Time `json:"second_confirm_at,omitempty"`
	
	// 授权信息
	AuthorizedBy     string     `json:"authorized_by,omitempty"`
	AuthorizedAt     *time.Time `json:"authorized_at,omitempty"`
	
	// 时间信息
	CreatedAt        time.Time  `json:"created_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	ConfirmedAt      *time.Time `json:"confirmed_at,omitempty"`
	RejectedAt       *time.Time `json:"rejected_at,omitempty"`
	
	// 审计信息
	RejectReason     string     `json:"reject_reason,omitempty"`
	Remarks          string     `json:"remarks,omitempty"`
	IPAddress        string     `json:"ip_address,omitempty"`
	UserAgent        string     `json:"user_agent,omitempty"`
}

// AuditLog 审计日志
type AuditLog struct {
	ID           string                 `json:"id"`
	OperationID  string                 `json:"operation_id"`
	Action       string                 `json:"action"`
	UserID       string                 `json:"user_id"`
	Username     string                 `json:"username"`
	Timestamp    time.Time              `json:"timestamp"`
	Details      map[string]interface{} `json:"details"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
}

// RollbackRecord 回滚记录
type RollbackRecord struct {
	ID            string        `json:"id"`
	OperationID   string        `json:"operation_id"`
	OriginalOp    *ParsedOperation `json:"original_operation"`
	RollbackOp    *ParsedOperation `json:"rollback_operation"`
	Status        OperationStatus `json:"status"`
	InitiatedBy   string        `json:"initiated_by"`
	InitiatedAt   time.Time     `json:"initiated_at"`
	CompletedAt   *time.Time    `json:"completed_at,omitempty"`
	Error         string        `json:"error,omitempty"`
}

// ConfirmationConfig 确认配置
type ConfirmationConfig struct {
	DefaultTimeout      time.Duration `json:"default_timeout"`
	EnableTwoStep       bool          `json:"enable_two_step"`
	CodeLength          int           `json:"code_length"`
	MaxPendingConfirm   int           `json:"max_pending_confirm"`
	AuditLogCapacity    int           `json:"audit_log_capacity"`
}

// DefaultConfirmationConfig 默认确认配置
func DefaultConfirmationConfig() *ConfirmationConfig {
	return &ConfirmationConfig{
		DefaultTimeout:    5 * time.Minute,
		EnableTwoStep:     true,
		CodeLength:        6,
		MaxPendingConfirm: 1000,
		AuditLogCapacity:  10000,
	}
}

// ConfirmationManager 确认管理器
type ConfirmationManager struct {
	config        *ConfirmationConfig
	confirmations map[string]*ConfirmationRecord
	auditLogs     []*AuditLog
	rollbackRecs  map[string]*RollbackRecord
	mu            sync.RWMutex
	auditMu       sync.RWMutex
	executor      *OperationExecutor
}

// NewConfirmationManager 创建确认管理器
func NewConfirmationManager(config *ConfirmationConfig, executor *OperationExecutor) *ConfirmationManager {
	if config == nil {
		config = DefaultConfirmationConfig()
	}

	return &ConfirmationManager{
		config:        config,
		confirmations: make(map[string]*ConfirmationRecord),
		auditLogs:     make([]*AuditLog, 0, config.AuditLogCapacity),
		rollbackRecs:  make(map[string]*RollbackRecord),
		executor:      executor,
	}
}

// CreateConfirmation 创建确认请求
func (m *ConfirmationManager) CreateConfirmation(ctx context.Context, op *ParsedOperation) (*ConfirmationRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否需要确认
	if op.Constraints == nil || !op.Constraints.RequireConfirm {
		// 不需要确认，直接返回已确认状态
		now := time.Now()
		return &ConfirmationRecord{
			ID:               generateConfirmID(),
			OperationID:      op.ID,
			Operation:        op,
			State:            ConfirmationStateConfirmed,
			RequiredAuthLevel: 0,
			CreatedAt:        now,
			ExpiresAt:        now.Add(24 * time.Hour),
		}, nil
	}

	// 检查待确认数量
	pendingCount := 0
	for _, c := range m.confirmations {
		if c.State == ConfirmationStatePending {
			pendingCount++
		}
	}
	if pendingCount >= m.config.MaxPendingConfirm {
		return nil, errors.New("too many pending confirmations")
	}

	// 生成确认码
	confirmCode, err := generateConfirmCode(m.config.CodeLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate confirm code: %w", err)
	}

	// 创建确认记录
	now := time.Now()
	record := &ConfirmationRecord{
		ID:               generateConfirmID(),
		OperationID:      op.ID,
		Operation:        op,
		State:            ConfirmationStatePending,
		ConfirmCode:      confirmCode,
		RequiredAuthLevel: op.Constraints.AuthLevel,
		CreatedAt:        now,
		ExpiresAt:        now.Add(m.config.DefaultTimeout),
	}

	m.confirmations[record.ID] = record

	// 记录审计日志
	m.addAuditLog(&AuditLog{
		ID:          generateAuditID(),
		OperationID: op.ID,
		Action:      "confirmation_created",
		Timestamp:   now,
		Details: map[string]interface{}{
			"confirm_code":       confirmCode,
			"required_auth_level": record.RequiredAuthLevel,
			"expires_at":         record.ExpiresAt,
		},
	})

	return record, nil
}

// FirstStepConfirm 第一步确认
func (m *ConfirmationManager) FirstStepConfirm(ctx context.Context, confirmID, confirmCode, userID string, authLevel int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, exists := m.confirmations[confirmID]
	if !exists {
		return ErrConfirmationNotFound
	}

	// 检查状态
	if record.State != ConfirmationStatePending {
		return ErrConfirmationAlreadyUsed
	}

	// 检查是否过期
	if time.Now().After(record.ExpiresAt) {
		record.State = ConfirmationStateExpired
		return ErrConfirmationExpired
	}

	// 验证确认码
	if record.ConfirmCode != confirmCode {
		return ErrInvalidConfirmCode
	}

	// 检查授权级别
	if authLevel < record.RequiredAuthLevel {
		return ErrInsufficientAuthLevel
	}

	// 如果不需要两步确认，直接完成
	if !m.config.EnableTwoStep || record.RequiredAuthLevel < 2 {
		now := time.Now()
		record.State = ConfirmationStateConfirmed
		record.FirstConfirmBy = userID
		record.FirstConfirmAt = &now
		record.ConfirmedAt = &now
		
		m.addAuditLog(&AuditLog{
			ID:          generateAuditID(),
			OperationID: record.OperationID,
			Action:      "confirmed",
			UserID:      userID,
			Timestamp:   now,
			Details: map[string]interface{}{
				"confirm_type": "single_step",
			},
		})
		
		return nil
	}

	// 第一步确认
	now := time.Now()
	record.State = ConfirmationStateFirstStep
	record.FirstConfirmBy = userID
	record.FirstConfirmAt = &now

	m.addAuditLog(&AuditLog{
		ID:          generateAuditID(),
		OperationID: record.OperationID,
		Action:      "first_step_confirmed",
		UserID:      userID,
		Timestamp:   now,
	})

	return nil
}

// SecondStepConfirm 第二步确认
func (m *ConfirmationManager) SecondStepConfirm(ctx context.Context, confirmID, userID string, authLevel int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, exists := m.confirmations[confirmID]
	if !exists {
		return ErrConfirmationNotFound
	}

	// 检查状态
	if record.State != ConfirmationStateFirstStep {
		return errors.New("invalid confirmation state for second step")
	}

	// 检查是否过期
	if time.Now().After(record.ExpiresAt) {
		record.State = ConfirmationStateExpired
		return ErrConfirmationExpired
	}

	// 第二步确认不能和第一步是同一人
	if record.FirstConfirmBy == userID {
		return errors.New("second step must be confirmed by a different user")
	}

	// 检查授权级别
	if authLevel < record.RequiredAuthLevel {
		return ErrInsufficientAuthLevel
	}

	// 完成确认
	now := time.Now()
	record.State = ConfirmationStateConfirmed
	record.SecondConfirmBy = userID
	record.SecondConfirmAt = &now
	record.ConfirmedAt = &now

	m.addAuditLog(&AuditLog{
		ID:          generateAuditID(),
		OperationID: record.OperationID,
		Action:      "confirmed",
		UserID:      userID,
		Timestamp:   now,
		Details: map[string]interface{}{
			"confirm_type":     "two_step",
			"first_confirmed_by": record.FirstConfirmBy,
			"second_confirmed_by": userID,
		},
	})

	return nil
}

// Reject 拒绝操作
func (m *ConfirmationManager) Reject(ctx context.Context, confirmID, userID, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, exists := m.confirmations[confirmID]
	if !exists {
		return ErrConfirmationNotFound
	}

	// 检查状态
	if record.State != ConfirmationStatePending && record.State != ConfirmationStateFirstStep {
		return ErrConfirmationAlreadyUsed
	}

	// 拒绝操作
	now := time.Now()
	record.State = ConfirmationStateRejected
	record.RejectedAt = &now
	record.RejectReason = reason

	m.addAuditLog(&AuditLog{
		ID:          generateAuditID(),
		OperationID: record.OperationID,
		Action:      "rejected",
		UserID:      userID,
		Timestamp:   now,
		Details: map[string]interface{}{
			"reason": reason,
		},
	})

	return nil
}

// Cancel 取消确认
func (m *ConfirmationManager) Cancel(ctx context.Context, confirmID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, exists := m.confirmations[confirmID]
	if !exists {
		return ErrConfirmationNotFound
	}

	// 检查状态
	if record.State != ConfirmationStatePending && record.State != ConfirmationStateFirstStep {
		return errors.New("cannot cancel in current state")
	}

	// 取消确认
	now := time.Now()
	record.State = ConfirmationStateCancelled

	m.addAuditLog(&AuditLog{
		ID:          generateAuditID(),
		OperationID: record.OperationID,
		Action:      "cancelled",
		UserID:      userID,
		Timestamp:   now,
	})

	return nil
}

// GetConfirmation 获取确认记录
func (m *ConfirmationManager) GetConfirmation(confirmID string) (*ConfirmationRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	record, exists := m.confirmations[confirmID]
	if !exists {
		return nil, ErrConfirmationNotFound
	}

	return record, nil
}

// GetConfirmationByOperation 通过操作ID获取确认记录
func (m *ConfirmationManager) GetConfirmationByOperation(opID string) (*ConfirmationRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, record := range m.confirmations {
		if record.OperationID == opID {
			return record, nil
		}
	}

	return nil, ErrConfirmationNotFound
}

// IsConfirmed 检查是否已确认
func (m *ConfirmationManager) IsConfirmed(confirmID string) (bool, error) {
	record, err := m.GetConfirmation(confirmID)
	if err != nil {
		return false, err
	}

	return record.State == ConfirmationStateConfirmed, nil
}

// Authorize 授权操作
func (m *ConfirmationManager) Authorize(ctx context.Context, confirmID, userID string, authLevel int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, exists := m.confirmations[confirmID]
	if !exists {
		return ErrConfirmationNotFound
	}

	// 检查是否已确认
	if record.State != ConfirmationStateConfirmed {
		return errors.New("operation must be confirmed before authorization")
	}

	// 检查授权级别
	if authLevel < record.RequiredAuthLevel {
		return ErrInsufficientAuthLevel
	}

	// 授权
	now := time.Now()
	record.AuthorizedBy = userID
	record.AuthorizedAt = &now

	m.addAuditLog(&AuditLog{
		ID:          generateAuditID(),
		OperationID: record.OperationID,
		Action:      "authorized",
		UserID:      userID,
		Timestamp:   now,
	})

	return nil
}

// Rollback 回滚操作
func (m *ConfirmationManager) Rollback(ctx context.Context, opID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取确认记录
	var confirmRecord *ConfirmationRecord
	for _, record := range m.confirmations {
		if record.OperationID == opID {
			confirmRecord = record
			break
		}
	}

	if confirmRecord == nil {
		return ErrConfirmationNotFound
	}

	// 检查操作是否允许回滚
	if confirmRecord.Operation.Constraints == nil || !confirmRecord.Operation.Constraints.AllowRollback {
		return errors.New("rollback is not allowed for this operation")
	}

	// 创建回滚记录
	rollbackRec := &RollbackRecord{
		ID:          generateRollbackID(),
		OperationID: opID,
		OriginalOp:  confirmRecord.Operation,
		InitiatedBy: userID,
		InitiatedAt: time.Now(),
		Status:      StatusPending,
	}

	m.rollbackRecs[rollbackRec.ID] = rollbackRec

	// 执行回滚
	if m.executor != nil {
		if err := m.executor.Rollback(ctx, opID); err != nil {
			rollbackRec.Status = StatusFailed
			rollbackRec.Error = err.Error()
			return err
		}
	}

	// 更新回滚记录
	now := time.Now()
	rollbackRec.Status = StatusSuccess
	rollbackRec.CompletedAt = &now

	// 记录审计日志
	m.addAuditLog(&AuditLog{
		ID:          generateAuditID(),
		OperationID: opID,
		Action:      "rollback",
		UserID:      userID,
		Timestamp:   now,
		Details: map[string]interface{}{
			"rollback_id": rollbackRec.ID,
		},
	})

	return nil
}

// GetAuditLogs 获取审计日志
func (m *ConfirmationManager) GetAuditLogs(opID string, limit int) []*AuditLog {
	m.auditMu.RLock()
	defer m.auditMu.RUnlock()

	var result []*AuditLog
	for i := len(m.auditLogs) - 1; i >= 0 && len(result) < limit; i-- {
		if opID == "" || m.auditLogs[i].OperationID == opID {
			result = append(result, m.auditLogs[i])
		}
	}

	return result
}

// addAuditLog 添加审计日志
func (m *ConfirmationManager) addAuditLog(log *AuditLog) {
	m.auditMu.Lock()
	defer m.auditMu.Unlock()

	m.auditLogs = append(m.auditLogs, log)

	// 限制日志数量
	if len(m.auditLogs) > m.config.AuditLogCapacity {
		m.auditLogs = m.auditLogs[1:]
	}
}

// CleanupExpired 清理过期确认
func (m *ConfirmationManager) CleanupExpired() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	count := 0

	for _, record := range m.confirmations {
		if record.State == ConfirmationStatePending || record.State == ConfirmationStateFirstStep {
			if now.After(record.ExpiresAt) {
				record.State = ConfirmationStateExpired
				count++
			}
		}
	}

	return count
}

// GetPendingConfirmations 获取待确认列表
func (m *ConfirmationManager) GetPendingConfirmations() []*ConfirmationRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ConfirmationRecord, 0)
	for _, record := range m.confirmations {
		if record.State == ConfirmationStatePending || record.State == ConfirmationStateFirstStep {
			result = append(result, record)
		}
	}

	return result
}

// GetStats 获取统计信息
func (m *ConfirmationManager) GetStats() *ConfirmationStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &ConfirmationStats{
		ByState: make(map[ConfirmationState]int),
	}

	for _, record := range m.confirmations {
		stats.ByState[record.State]++
	}

	m.auditMu.RLock()
	stats.TotalAuditLogs = len(m.auditLogs)
	m.auditMu.RUnlock()

	return stats
}

// ConfirmationStats 确认统计信息
type ConfirmationStats struct {
	ByState        map[ConfirmationState]int `json:"by_state"`
	TotalAuditLogs int                       `json:"total_audit_logs"`
}

// 辅助函数

func generateConfirmID() string {
	return fmt.Sprintf("CONF-%d", time.Now().UnixNano())
}

func generateConfirmCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

func generateAuditID() string {
	return fmt.Sprintf("AUDIT-%d", time.Now().UnixNano())
}

func generateRollbackID() string {
	return fmt.Sprintf("ROLL-%d", time.Now().UnixNano())
}
