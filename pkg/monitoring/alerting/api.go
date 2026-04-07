package alerting

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// AlertAPI 告警API
type AlertAPI struct {
	ruleManager     *RuleManager
	aggregator      *AlertAggregator
	notifier        *AlertNotifier
	alertStore      AlertStore
}

// AlertStore 告警存储接口
type AlertStore interface {
	// Save 保存告警
	Save(ctx context.Context, alert *AlertInstance) error

	// Update 更新告警
	Update(ctx context.Context, alert *AlertInstance) error

	// Get 获取告警
	Get(ctx context.Context, id string) (*AlertInstance, error)

	// Query 查询告警
	Query(ctx context.Context, query *AlertQuery) ([]*AlertInstance, int64, error)

	// Acknowledge 确认告警
	Acknowledge(ctx context.Context, id string, by string) error

	// Clear 清除告警
	Clear(ctx context.Context, id string) error

	// GetStatistics 获取统计信息
	GetStatistics(ctx context.Context, query *StatisticsQuery) (*AlertStatistics, error)

	// GetHistory 获取历史记录
	GetHistory(ctx context.Context, query *HistoryQuery) ([]*AlertHistory, int64, error)
}

// AlertQuery 告警查询
type AlertQuery struct {
	// 过滤条件
	IDs         []string       `json:"ids,omitempty"`
	RuleIDs     []string       `json:"rule_ids,omitempty"`
	Categories  []AlertCategory `json:"categories,omitempty"`
	Severities  []AlertSeverity `json:"severities,omitempty"`
	Sources     []string       `json:"sources,omitempty"`
	Status      []string       `json:"status,omitempty"`

	// 时间范围
	StartTime   *time.Time     `json:"start_time,omitempty"`
	EndTime     *time.Time     `json:"end_time,omitempty"`

	// 标签过滤
	Labels      map[string]string `json:"labels,omitempty"`

	// 分页
	Page        int            `json:"page"`
	PageSize    int            `json:"page_size"`

	// 排序
	SortBy      string         `json:"sort_by,omitempty"`
	SortOrder   string         `json:"sort_order,omitempty"` // asc, desc
}

// AlertStatistics 告警统计
type AlertStatistics struct {
	// 总数
	TotalCount    int64 `json:"total_count"`

	// 按严重程度统计
	BySeverity    map[AlertSeverity]int64 `json:"by_severity"`

	// 按类别统计
	ByCategory    map[AlertCategory]int64 `json:"by_category"`

	// 按状态统计
	ByStatus      map[string]int64 `json:"by_status"`

	// 按来源统计
	BySource      map[string]int64 `json:"by_source"`

	// 时间分布
	TimeDistribution []*TimeDistribution `json:"time_distribution"`

	// 趋势
	Trend         *AlertTrend `json:"trend,omitempty"`
}

// TimeDistribution 时间分布
type TimeDistribution struct {
	Time  time.Time `json:"time"`
	Count int64     `json:"count"`
}

// AlertTrend 告警趋势
type AlertTrend struct {
	CurrentPeriod int64   `json:"current_period"`
	LastPeriod    int64   `json:"last_period"`
	Change        float64 `json:"change"` // 变化百分比
}

// StatisticsQuery 统计查询
type StatisticsQuery struct {
	// 时间范围
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`

	// 分组维度
	GroupBy   []string `json:"group_by,omitempty"` // severity, category, source, status

	// 过滤条件
	RuleIDs    []string        `json:"rule_ids,omitempty"`
	Categories []AlertCategory `json:"categories,omitempty"`
	Sources    []string        `json:"sources,omitempty"`
}

// HistoryQuery 历史查询
type HistoryQuery struct {
	AlertID    string     `json:"alert_id,omitempty"`
	StartTime  *time.Time `json:"start_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
}

// AlertHistory 告警历史
type AlertHistory struct {
	ID          string        `json:"id"`
	AlertID     string        `json:"alert_id"`
	Action      string        `json:"action"` // created, acknowledged, cleared, escalated
	Operator    string        `json:"operator"`
	Timestamp   time.Time     `json:"timestamp"`
	Details     string        `json:"details,omitempty"`
	OldValue    string        `json:"old_value,omitempty"`
	NewValue    string        `json:"new_value,omitempty"`
}

// APIResponse API响应
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	PageSize int        `json:"page_size"`
}

// NewAlertAPI 创建告警API
func NewAlertAPI(ruleManager *RuleManager, aggregator *AlertAggregator, notifier *AlertNotifier, store AlertStore) *AlertAPI {
	return &AlertAPI{
		ruleManager: ruleManager,
		aggregator:  aggregator,
		notifier:    notifier,
		alertStore:  store,
	}
}

// RegisterRoutes 注册路由
func (api *AlertAPI) RegisterRoutes(mux *http.ServeMux) {
	// 告警查询
	mux.HandleFunc("/api/v1/alerts", api.handleAlerts)
	mux.HandleFunc("/api/v1/alerts/", api.handleAlertDetail)

	// 告警操作
	mux.HandleFunc("/api/v1/alerts/acknowledge", api.handleAcknowledge)
	mux.HandleFunc("/api/v1/alerts/clear", api.handleClear)

	// 告警统计
	mux.HandleFunc("/api/v1/alerts/statistics", api.handleStatistics)

	// 告警历史
	mux.HandleFunc("/api/v1/alerts/history", api.handleHistory)

	// 规则管理
	mux.HandleFunc("/api/v1/alerts/rules", api.handleRules)
	mux.HandleFunc("/api/v1/alerts/rules/", api.handleRuleDetail)

	// 聚合器管理
	mux.HandleFunc("/api/v1/alerts/aggregator/stats", api.handleAggregatorStats)
	mux.HandleFunc("/api/v1/alerts/aggregator/flush", api.handleAggregatorFlush)

	// 静默管理
	mux.HandleFunc("/api/v1/alerts/silences", api.handleSilences)
	mux.HandleFunc("/api/v1/alerts/silences/", api.handleSilenceDetail)

	// 抑制规则管理
	mux.HandleFunc("/api/v1/alerts/suppressions", api.handleSuppressions)
	mux.HandleFunc("/api/v1/alerts/suppressions/", api.handleSuppressionDetail)

	// 通知模板管理
	mux.HandleFunc("/api/v1/alerts/templates", api.handleTemplates)
	mux.HandleFunc("/api/v1/alerts/templates/", api.handleTemplateDetail)

	// 升级规则管理
	mux.HandleFunc("/api/v1/alerts/escalations", api.handleEscalations)
	mux.HandleFunc("/api/v1/alerts/escalations/", api.handleEscalationDetail)
}

// handleAlerts 处理告警列表
func (api *AlertAPI) handleAlerts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.queryAlerts(w, r)
	case http.MethodPost:
		api.createAlert(w, r)
	default:
		api.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// queryAlerts 查询告警
func (api *AlertAPI) queryAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := &AlertQuery{
		Page:     1,
		PageSize: 20,
	}

	// 解析查询参数
	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			query.Page = p
		}
	}

	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			query.PageSize = ps
		}
	}

	if ruleIDs := r.URL.Query().Get("rule_ids"); ruleIDs != "" {
		query.RuleIDs = strings.Split(ruleIDs, ",")
	}

	if categories := r.URL.Query().Get("categories"); categories != "" {
		for _, cat := range strings.Split(categories, ",") {
			query.Categories = append(query.Categories, AlertCategory(cat))
		}
	}

	if severities := r.URL.Query().Get("severities"); severities != "" {
		for _, sev := range strings.Split(severities, ",") {
			query.Severities = append(query.Severities, AlertSeverity(sev))
		}
	}

	if sources := r.URL.Query().Get("sources"); sources != "" {
		query.Sources = strings.Split(sources, ",")
	}

	if startTime := r.URL.Query().Get("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			query.StartTime = &t
		}
	}

	if endTime := r.URL.Query().Get("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			query.EndTime = &t
		}
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		query.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		query.SortOrder = sortOrder
	}

	// 查询告警
	alerts, total, err := api.alertStore.Query(ctx, query)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to query alerts: %v", err))
		return
	}

	api.writePaginatedResponse(w, alerts, total, query.Page, query.PageSize)
}

// createAlert 创建告警
func (api *AlertAPI) createAlert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var alert AlertInstance
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		api.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// 设置默认值
	if alert.ID == "" {
		alert.ID = generateAlertID()
	}
	if alert.TriggeredAt.IsZero() {
		alert.TriggeredAt = time.Now()
	}

	// 保存告警
	if err := api.alertStore.Save(ctx, &alert); err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save alert: %v", err))
		return
	}

	// 聚合告警
	if api.aggregator != nil {
		_, _, _ = api.aggregator.Aggregate(ctx, &alert)
	}

	api.writeSuccess(w, alert)
}

// handleAlertDetail 处理告警详情
func (api *AlertAPI) handleAlertDetail(w http.ResponseWriter, r *http.Request) {
	// 提取告警ID
	alertID := strings.TrimPrefix(r.URL.Path, "/api/v1/alerts/")
	if alertID == "" {
		api.writeError(w, http.StatusBadRequest, "alert id is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		api.getAlert(w, r, alertID)
	case http.MethodDelete:
		api.deleteAlert(w, r, alertID)
	default:
		api.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// getAlert 获取告警详情
func (api *AlertAPI) getAlert(w http.ResponseWriter, r *http.Request, alertID string) {
	ctx := r.Context()

	alert, err := api.alertStore.Get(ctx, alertID)
	if err != nil {
		api.writeError(w, http.StatusNotFound, fmt.Sprintf("alert not found: %v", err))
		return
	}

	api.writeSuccess(w, alert)
}

// deleteAlert 删除告警
func (api *AlertAPI) deleteAlert(w http.ResponseWriter, r *http.Request, alertID string) {
	// 这里应该实现删除逻辑
	// 为了示例，我们返回成功
	api.writeSuccess(w, map[string]string{
		"message": "alert deleted",
		"id":      alertID,
	})
}

// handleAcknowledge 处理告警确认
func (api *AlertAPI) handleAcknowledge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx := r.Context()

	var req struct {
		AlertID string `json:"alert_id"`
		By      string `json:"by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if req.AlertID == "" {
		api.writeError(w, http.StatusBadRequest, "alert_id is required")
		return
	}

	if err := api.alertStore.Acknowledge(ctx, req.AlertID, req.By); err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to acknowledge alert: %v", err))
		return
	}

	api.writeSuccess(w, map[string]string{
		"message": "alert acknowledged",
		"id":      req.AlertID,
	})
}

// handleClear 处理告警清除
func (api *AlertAPI) handleClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx := r.Context()

	var req struct {
		AlertID string `json:"alert_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if req.AlertID == "" {
		api.writeError(w, http.StatusBadRequest, "alert_id is required")
		return
	}

	if err := api.alertStore.Clear(ctx, req.AlertID); err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to clear alert: %v", err))
		return
	}

	api.writeSuccess(w, map[string]string{
		"message": "alert cleared",
		"id":      req.AlertID,
	})
}

// handleStatistics 处理告警统计
func (api *AlertAPI) handleStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx := r.Context()

	query := &StatisticsQuery{}

	// 解析查询参数
	if startTime := r.URL.Query().Get("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			query.StartTime = &t
		}
	}

	if endTime := r.URL.Query().Get("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			query.EndTime = &t
		}
	}

	if groupBy := r.URL.Query().Get("group_by"); groupBy != "" {
		query.GroupBy = strings.Split(groupBy, ",")
	}

	stats, err := api.alertStore.GetStatistics(ctx, query)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get statistics: %v", err))
		return
	}

	api.writeSuccess(w, stats)
}

// handleHistory 处理告警历史
func (api *AlertAPI) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx := r.Context()

	query := &HistoryQuery{
		Page:     1,
		PageSize: 20,
	}

	// 解析查询参数
	if alertID := r.URL.Query().Get("alert_id"); alertID != "" {
		query.AlertID = alertID
	}

	if startTime := r.URL.Query().Get("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			query.StartTime = &t
		}
	}

	if endTime := r.URL.Query().Get("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			query.EndTime = &t
		}
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			query.Page = p
		}
	}

	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			query.PageSize = ps
		}
	}

	history, total, err := api.alertStore.GetHistory(ctx, query)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get history: %v", err))
		return
	}

	api.writePaginatedResponse(w, history, total, query.Page, query.PageSize)
}

// handleRules 处理规则列表
func (api *AlertAPI) handleRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.listRules(w, r)
	case http.MethodPost:
		api.createRule(w, r)
	default:
		api.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// listRules 列出规则
func (api *AlertAPI) listRules(w http.ResponseWriter, r *http.Request) {
	rules := api.ruleManager.GetAllRules()
	api.writeSuccess(w, rules)
}

// createRule 创建规则
func (api *AlertAPI) createRule(w http.ResponseWriter, r *http.Request) {
	var rule AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		api.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if err := api.ruleManager.AddRule(&rule); err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create rule: %v", err))
		return
	}

	api.writeSuccess(w, rule)
}

// handleRuleDetail 处理规则详情
func (api *AlertAPI) handleRuleDetail(w http.ResponseWriter, r *http.Request) {
	ruleID := strings.TrimPrefix(r.URL.Path, "/api/v1/alerts/rules/")
	if ruleID == "" {
		api.writeError(w, http.StatusBadRequest, "rule id is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		api.getRule(w, r, ruleID)
	case http.MethodPut:
		api.updateRule(w, r, ruleID)
	case http.MethodDelete:
		api.deleteRule(w, r, ruleID)
	default:
		api.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// getRule 获取规则
func (api *AlertAPI) getRule(w http.ResponseWriter, r *http.Request, ruleID string) {
	rule, exists := api.ruleManager.GetRule(ruleID)
	if !exists {
		api.writeError(w, http.StatusNotFound, "rule not found")
		return
	}

	api.writeSuccess(w, rule)
}

// updateRule 更新规则
func (api *AlertAPI) updateRule(w http.ResponseWriter, r *http.Request, ruleID string) {
	var rule AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		api.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	rule.ID = ruleID
	if err := api.ruleManager.AddRule(&rule); err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to update rule: %v", err))
		return
	}

	api.writeSuccess(w, rule)
}

// deleteRule 删除规则
func (api *AlertAPI) deleteRule(w http.ResponseWriter, r *http.Request, ruleID string) {
	api.ruleManager.RemoveRule(ruleID)
	api.writeSuccess(w, map[string]string{
		"message": "rule deleted",
		"id":      ruleID,
	})
}

// handleAggregatorStats 处理聚合器统计
func (api *AlertAPI) handleAggregatorStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	stats := api.aggregator.GetStats()
	api.writeSuccess(w, stats)
}

// handleAggregatorFlush 处理聚合器刷新
func (api *AlertAPI) handleAggregatorFlush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx := r.Context()
	if err := api.aggregator.Flush(ctx); err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to flush: %v", err))
		return
	}

	api.writeSuccess(w, map[string]string{
		"message": "aggregator flushed",
	})
}

// handleSilences 处理静默列表
func (api *AlertAPI) handleSilences(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.listSilences(w, r)
	case http.MethodPost:
		api.createSilence(w, r)
	default:
		api.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// listSilences 列出静默规则
func (api *AlertAPI) listSilences(w http.ResponseWriter, r *http.Request) {
	silences := api.aggregator.silenceManager.ListSilences()
	api.writeSuccess(w, silences)
}

// createSilence 创建静默规则
func (api *AlertAPI) createSilence(w http.ResponseWriter, r *http.Request) {
	var silence Silence
	if err := json.NewDecoder(r.Body).Decode(&silence); err != nil {
		api.writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if err := api.aggregator.silenceManager.AddSilence(&silence); err != nil {
		api.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create silence: %v", err))
		return
	}

	api.writeSuccess(w, silence)
}

// handleSilenceDetail 处理静默详情
func (api *AlertAPI) handleSilenceDetail(w http.ResponseWriter, r *http.Request) {
	silenceID := strings.TrimPrefix(r.URL.Path, "/api/v1/alerts/silences/")
	if silenceID == "" {
		api.writeError(w, http.StatusBadRequest, "silence id is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		api.getSilence(w, r, silenceID)
	case http.MethodDelete:
		api.deleteSilence(w, r, silenceID)
	default:
		api.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// getSilence 获取静默规则
func (api *AlertAPI) getSilence(w http.ResponseWriter, r *http.Request, silenceID string) {
	silence, exists := api.aggregator.silenceManager.GetSilence(silenceID)
	if !exists {
		api.writeError(w, http.StatusNotFound, "silence not found")
		return
	}

	api.writeSuccess(w, silence)
}

// deleteSilence 删除静默规则
func (api *AlertAPI) deleteSilence(w http.ResponseWriter, r *http.Request, silenceID string) {
	api.aggregator.silenceManager.RemoveSilence(silenceID)
	api.writeSuccess(w, map[string]string{
		"message": "silence deleted",
		"id":      silenceID,
	})
}

// handleSuppressions 处理抑制规则列表
func (api *AlertAPI) handleSuppressions(w http.ResponseWriter, r *http.Request) {
	// 简化实现
	api.writeSuccess(w, []interface{}{})
}

// handleSuppressionDetail 处理抑制规则详情
func (api *AlertAPI) handleSuppressionDetail(w http.ResponseWriter, r *http.Request) {
	// 简化实现
	api.writeError(w, http.StatusNotImplemented, "not implemented")
}

// handleTemplates 处理模板列表
func (api *AlertAPI) handleTemplates(w http.ResponseWriter, r *http.Request) {
	// 简化实现
	api.writeSuccess(w, []interface{}{})
}

// handleTemplateDetail 处理模板详情
func (api *AlertAPI) handleTemplateDetail(w http.ResponseWriter, r *http.Request) {
	// 简化实现
	api.writeError(w, http.StatusNotImplemented, "not implemented")
}

// handleEscalations 处理升级规则列表
func (api *AlertAPI) handleEscalations(w http.ResponseWriter, r *http.Request) {
	// 简化实现
	api.writeSuccess(w, []interface{}{})
}

// handleEscalationDetail 处理升级规则详情
func (api *AlertAPI) handleEscalationDetail(w http.ResponseWriter, r *http.Request) {
	// 简化实现
	api.writeError(w, http.StatusNotImplemented, "not implemented")
}

// writeSuccess 写入成功响应
func (api *AlertAPI) writeSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

// writeError 写入错误响应
func (api *AlertAPI) writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   message,
	})
}

// writePaginatedResponse 写入分页响应
func (api *AlertAPI) writePaginatedResponse(w http.ResponseWriter, data interface{}, total int64, page, pageSize int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(PaginatedResponse{
		Success:  true,
		Data:     data,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// generateAlertID 生成告警ID
func generateAlertID() string {
	return fmt.Sprintf("alert-%d", time.Now().UnixNano())
}
