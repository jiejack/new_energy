package operation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// OperationAPI 操作API
type OperationAPI struct {
	parser      *OperationParser
	executor    *OperationExecutor
	confirmer   *ConfirmationManager
	authChecker AuthChecker
}

// AuthChecker 授权检查器接口
type AuthChecker interface {
	CheckPermission(ctx context.Context, userID, permission string) (bool, error)
	GetAuthLevel(ctx context.Context, userID string) (int, error)
	GetUserInfo(ctx context.Context, userID string) (*UserInfo, error)
}

// UserInfo 用户信息
type UserInfo struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	AuthLevel   int      `json:"auth_level"`
}

// NewOperationAPI 创建操作API
func NewOperationAPI(
	parser *OperationParser,
	executor *OperationExecutor,
	confirmer *ConfirmationManager,
	authChecker AuthChecker,
) *OperationAPI {
	return &OperationAPI{
		parser:      parser,
		executor:    executor,
		confirmer:   confirmer,
		authChecker: authChecker,
	}
}

// OperationRequest 操作请求
type OperationRequest struct {
	Text         string                 `json:"text"`          // 自然语言指令
	UserID       string                 `json:"user_id"`       // 用户ID
	IPAddress    string                 `json:"ip_address"`    // IP地址
	UserAgent    string                 `json:"user_agent"`    // User-Agent
	DryRun       bool                   `json:"dry_run"`       // 试运行模式
	Parameters   map[string]interface{} `json:"parameters"`    // 额外参数
	Constraints  *OperationConstraints  `json:"constraints"`   // 自定义约束
}

// OperationResponse 操作响应
type OperationResponse struct {
	Success      bool               `json:"success"`
	Message      string             `json:"message"`
	Operations   []*ParsedOperation `json:"operations,omitempty"`
	Confirmations []*ConfirmInfo    `json:"confirmations,omitempty"`
	Warnings     []string           `json:"warnings,omitempty"`
	Suggestions  []string           `json:"suggestions,omitempty"`
	RequestID    string             `json:"request_id"`
	Timestamp    time.Time          `json:"timestamp"`
}

// ConfirmInfo 确认信息
type ConfirmInfo struct {
	ConfirmID         string            `json:"confirm_id"`
	OperationID       string            `json:"operation_id"`
	ConfirmCode       string            `json:"confirm_code"`
	RequiredAuthLevel int               `json:"required_auth_level"`
	ExpiresAt         time.Time         `json:"expires_at"`
	Description       string            `json:"description"`
	State             ConfirmationState `json:"state"`
}

// ConfirmRequest 确认请求
type ConfirmRequest struct {
	ConfirmID   string `json:"confirm_id"`
	ConfirmCode string `json:"confirm_code"`
	UserID      string `json:"user_id"`
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	Step        int    `json:"step"` // 1 or 2 for two-step confirmation
}

// ConfirmResponse 确认响应
type ConfirmResponse struct {
	Success     bool               `json:"success"`
	Message     string             `json:"message"`
	State       ConfirmationState  `json:"state"`
	NeedSecondStep bool            `json:"need_second_step"`
	ConfirmID   string             `json:"confirm_id,omitempty"`
	Timestamp   time.Time          `json:"timestamp"`
}

// StatusRequest 状态查询请求
type StatusRequest struct {
	OperationID string `json:"operation_id"`
}

// StatusResponse 状态查询响应
type StatusResponse struct {
	Success bool              `json:"success"`
	Status  OperationStatus   `json:"status"`
	Record  *OperationRecord  `json:"record,omitempty"`
}

// HistoryRequest 历史记录请求
type HistoryRequest struct {
	UserID    string     `json:"user_id"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	Status    string     `json:"status"`
	OperationType string `json:"operation_type"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

// HistoryResponse 历史记录响应
type HistoryResponse struct {
	Success bool               `json:"success"`
	Total   int                `json:"total"`
	Records []*OperationRecord `json:"records"`
}

// RollbackRequest 回滚请求
type RollbackRequest struct {
	OperationID string `json:"operation_id"`
	UserID      string `json:"user_id"`
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	Reason      string `json:"reason"`
}

// RollbackResponse 回滚响应
type RollbackResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	RollbackID  string    `json:"rollback_id,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// SubmitOperation 提交操作
func (api *OperationAPI) SubmitOperation(ctx context.Context, req *OperationRequest) (*OperationResponse, error) {
	response := &OperationResponse{
		Success:    false,
		RequestID:  generateRequestID(),
		Timestamp:  time.Now(),
		Warnings:   make([]string, 0),
		Suggestions: make([]string, 0),
	}

	// 1. 解析自然语言指令
	parseResult, err := api.parser.Parse(ctx, req.Text)
	if err != nil {
		response.Message = fmt.Sprintf("解析失败: %v", err)
		return response, nil
	}

	response.Operations = parseResult.Operations
	response.Warnings = append(response.Warnings, parseResult.Warnings...)
	response.Suggestions = append(response.Suggestions, parseResult.Suggestions...)

	if len(parseResult.Operations) == 0 {
		response.Message = "未识别到有效操作"
		return response, nil
	}

	// 2. 应用自定义约束和参数
	for _, op := range parseResult.Operations {
		// 合并参数
		for k, v := range req.Parameters {
			op.Parameters[k] = v
		}

		// 应用自定义约束
		if req.Constraints != nil {
			if op.Constraints == nil {
				op.Constraints = req.Constraints
			} else {
				// 合并约束
				if req.Constraints.Timeout > 0 {
					op.Constraints.Timeout = req.Constraints.Timeout
				}
				if req.Constraints.MaxRetries > 0 {
					op.Constraints.MaxRetries = req.Constraints.MaxRetries
				}
				if req.Constraints.AuthLevel > 0 {
					op.Constraints.AuthLevel = req.Constraints.AuthLevel
				}
			}
		}

		// 设置试运行模式
		if req.DryRun && op.Constraints != nil {
			op.Constraints.DryRun = true
		}
	}

	// 3. 创建确认请求
	response.Confirmations = make([]*ConfirmInfo, 0)
	for _, op := range parseResult.Operations {
		confirmRec, err := api.confirmer.CreateConfirmation(ctx, op)
		if err != nil {
			response.Warnings = append(response.Warnings, 
				fmt.Sprintf("创建确认请求失败: %v", err))
			continue
		}

		response.Confirmations = append(response.Confirmations, &ConfirmInfo{
			ConfirmID:         confirmRec.ID,
			OperationID:       confirmRec.OperationID,
			ConfirmCode:       confirmRec.ConfirmCode,
			RequiredAuthLevel: confirmRec.RequiredAuthLevel,
			ExpiresAt:         confirmRec.ExpiresAt,
			Description:       op.Description,
			State:             confirmRec.State,
		})
	}

	// 4. 如果不需要确认，直接提交执行
	for _, confirmInfo := range response.Confirmations {
		confirmRec, _ := api.confirmer.GetConfirmation(confirmInfo.ConfirmID)
		if confirmRec != nil && confirmRec.State == ConfirmationStateConfirmed {
			// 直接提交执行
			if err := api.executor.Submit(ctx, confirmRec.Operation); err != nil {
				response.Warnings = append(response.Warnings,
					fmt.Sprintf("提交执行失败: %v", err))
			}
		}
	}

	response.Success = true
	response.Message = fmt.Sprintf("成功解析 %d 个操作", len(parseResult.Operations))

	return response, nil
}

// ConfirmOperation 确认操作
func (api *OperationAPI) ConfirmOperation(ctx context.Context, req *ConfirmRequest) (*ConfirmResponse, error) {
	response := &ConfirmResponse{
		Success:   false,
		Timestamp: time.Now(),
	}

	// 获取确认记录
	confirmRec, err := api.confirmer.GetConfirmation(req.ConfirmID)
	if err != nil {
		response.Message = "确认记录不存在"
		return response, nil
	}

	// 获取用户授权级别
	authLevel := 0
	if api.authChecker != nil {
		level, err := api.authChecker.GetAuthLevel(ctx, req.UserID)
		if err == nil {
			authLevel = level
		}
	}

	// 执行确认
	var confirmErr error
	if req.Step == 1 || confirmRec.State == ConfirmationStatePending {
		confirmErr = api.confirmer.FirstStepConfirm(ctx, req.ConfirmID, req.ConfirmCode, req.UserID, authLevel)
	} else if req.Step == 2 {
		confirmErr = api.confirmer.SecondStepConfirm(ctx, req.ConfirmID, req.UserID, authLevel)
	}

	if confirmErr != nil {
		response.Message = confirmErr.Error()
		return response, nil
	}

	// 更新确认记录
	confirmRec, _ = api.confirmer.GetConfirmation(req.ConfirmID)
	response.State = confirmRec.State

	// 检查是否需要第二步确认
	if confirmRec.State == ConfirmationStateFirstStep {
		response.NeedSecondStep = true
		response.Message = "第一步确认成功，需要第二步确认"
		response.Success = true
		return response, nil
	}

	// 确认完成，提交执行
	if confirmRec.State == ConfirmationStateConfirmed {
		if err := api.executor.Submit(ctx, confirmRec.Operation); err != nil {
			response.Message = fmt.Sprintf("提交执行失败: %v", err)
			return response, nil
		}
		response.Message = "确认成功，操作已提交执行"
		response.Success = true
		response.ConfirmID = req.ConfirmID
	}

	return response, nil
}

// RejectOperation 拒绝操作
func (api *OperationAPI) RejectOperation(ctx context.Context, confirmID, userID, reason string) error {
	return api.confirmer.Reject(ctx, confirmID, userID, reason)
}

// CancelOperation 取消操作
func (api *OperationAPI) CancelOperation(ctx context.Context, confirmID, userID string) error {
	// 取消确认
	if err := api.confirmer.Cancel(ctx, confirmID, userID); err != nil {
		return err
	}

	// 获取确认记录
	confirmRec, err := api.confirmer.GetConfirmation(confirmID)
	if err != nil {
		return nil
	}

	// 取消执行器中的操作
	return api.executor.Cancel(confirmRec.OperationID)
}

// GetOperationStatus 获取操作状态
func (api *OperationAPI) GetOperationStatus(ctx context.Context, req *StatusRequest) (*StatusResponse, error) {
	record, err := api.executor.GetStatus(req.OperationID)
	if err != nil {
		return &StatusResponse{
			Success: false,
		}, nil
	}

	return &StatusResponse{
		Success: true,
		Status:  record.Status,
		Record:  record,
	}, nil
}

// GetOperationHistory 获取操作历史
func (api *OperationAPI) GetOperationHistory(ctx context.Context, req *HistoryRequest) (*HistoryResponse, error) {
	records := api.executor.GetHistory(req.Limit)

	// 过滤
	if req.Status != "" {
		filtered := make([]*OperationRecord, 0)
		for _, r := range records {
			if string(r.Status) == req.Status {
				filtered = append(filtered, r)
			}
		}
		records = filtered
	}

	if req.OperationType != "" {
		filtered := make([]*OperationRecord, 0)
		for _, r := range records {
			if string(r.Operation.Type) == req.OperationType {
				filtered = append(filtered, r)
			}
		}
		records = filtered
	}

	if req.StartTime != nil {
		filtered := make([]*OperationRecord, 0)
		for _, r := range records {
			if r.StartTime != nil && r.StartTime.After(*req.StartTime) {
				filtered = append(filtered, r)
			}
		}
		records = filtered
	}

	if req.EndTime != nil {
		filtered := make([]*OperationRecord, 0)
		for _, r := range records {
			if r.EndTime != nil && r.EndTime.Before(*req.EndTime) {
				filtered = append(filtered, r)
			}
		}
		records = filtered
	}

	// 分页
	total := len(records)
	if req.Offset > 0 && req.Offset < len(records) {
		records = records[req.Offset:]
	}
	if req.Limit > 0 && req.Limit < len(records) {
		records = records[:req.Limit]
	}

	return &HistoryResponse{
		Success: true,
		Total:   total,
		Records: records,
	}, nil
}

// RollbackOperation 回滚操作
func (api *OperationAPI) RollbackOperation(ctx context.Context, req *RollbackRequest) (*RollbackResponse, error) {
	response := &RollbackResponse{
		Success:   false,
		Timestamp: time.Now(),
	}

	err := api.confirmer.Rollback(ctx, req.OperationID, req.UserID)
	if err != nil {
		response.Message = err.Error()
		return response, nil
	}

	response.Success = true
	response.Message = "回滚成功"
	response.RollbackID = fmt.Sprintf("ROLL-%d", time.Now().UnixNano())

	return response, nil
}

// GetPendingConfirmations 获取待确认列表
func (api *OperationAPI) GetPendingConfirmations(ctx context.Context) ([]*ConfirmInfo, error) {
	records := api.confirmer.GetPendingConfirmations()

	result := make([]*ConfirmInfo, 0, len(records))
	for _, rec := range records {
		result = append(result, &ConfirmInfo{
			ConfirmID:         rec.ID,
			OperationID:       rec.OperationID,
			ConfirmCode:       rec.ConfirmCode,
			RequiredAuthLevel: rec.RequiredAuthLevel,
			ExpiresAt:         rec.ExpiresAt,
			Description:       rec.Operation.Description,
			State:             rec.State,
		})
	}

	return result, nil
}

// GetAuditLogs 获取审计日志
func (api *OperationAPI) GetAuditLogs(ctx context.Context, opID string, limit int) ([]*AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}
	return api.confirmer.GetAuditLogs(opID, limit), nil
}

// GetStats 获取统计信息
func (api *OperationAPI) GetStats(ctx context.Context) *APIStats {
	executorStats := api.executor.GetStats()
	confirmerStats := api.confirmer.GetStats()

	return &APIStats{
		Executor:  executorStats,
		Confirmer: confirmerStats,
	}
}

// APIStats API统计信息
type APIStats struct {
	Executor  *ExecutorStats     `json:"executor"`
	Confirmer *ConfirmationStats `json:"confirmer"`
}

// HTTP Handlers

// HandleSubmit HTTP处理器 - 提交操作
func (api *OperationAPI) HandleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := api.SubmitOperation(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleConfirm HTTP处理器 - 确认操作
func (api *OperationAPI) HandleConfirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := api.ConfirmOperation(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleStatus HTTP处理器 - 查询状态
func (api *OperationAPI) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	opID := r.URL.Query().Get("operation_id")
	if opID == "" {
		http.Error(w, "operation_id is required", http.StatusBadRequest)
		return
	}

	response, err := api.GetOperationStatus(r.Context(), &StatusRequest{OperationID: opID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleHistory HTTP处理器 - 查询历史
func (api *OperationAPI) HandleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req := &HistoryRequest{
		Limit: 100,
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		fmt.Sscanf(limit, "%d", &req.Limit)
	}
	if status := r.URL.Query().Get("status"); status != "" {
		req.Status = status
	}
	if opType := r.URL.Query().Get("type"); opType != "" {
		req.OperationType = opType
	}

	response, err := api.GetOperationHistory(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleRollback HTTP处理器 - 回滚操作
func (api *OperationAPI) HandleRollback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := api.RollbackOperation(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandlePending HTTP处理器 - 获取待确认列表
func (api *OperationAPI) HandlePending(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	confirmations, err := api.GetPendingConfirmations(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(confirmations)
}

// HandleAuditLogs HTTP处理器 - 获取审计日志
func (api *OperationAPI) HandleAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	opID := r.URL.Query().Get("operation_id")
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	logs, err := api.GetAuditLogs(r.Context(), opID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// HandleStats HTTP处理器 - 获取统计信息
func (api *OperationAPI) HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := api.GetStats(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ValidateRequest 验证请求
func (api *OperationAPI) ValidateRequest(req *OperationRequest) error {
	if req.Text == "" {
		return errors.New("text is required")
	}
	if req.UserID == "" {
		return errors.New("user_id is required")
	}
	return nil
}

// 辅助函数

func generateRequestID() string {
	return fmt.Sprintf("REQ-%d", time.Now().UnixNano())
}
