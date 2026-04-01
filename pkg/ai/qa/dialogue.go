package qa

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DialogueState 对话状态
type DialogueState string

const (
	StateInitial    DialogueState = "initial"    // 初始状态
	StateActive     DialogueState = "active"     // 活跃状态
	StateWaiting    DialogueState = "waiting"    // 等待状态（等待用户输入）
	StateCompleted  DialogueState = "completed"  // 完成状态
	StateCancelled  DialogueState = "cancelled"  // 取消状态
	StateError      DialogueState = "error"      // 错误状态
)

// DialogueTurn 对话轮次
type DialogueTurn struct {
	TurnID      string
	Role        string // "user" 或 "system"
	Content     string
	Intent      *Intent
	Timestamp   time.Time
	Metadata    map[string]interface{}
}

// DialogueContext 对话上下文
type DialogueContext struct {
	SessionID      string
	UserID         string
	CurrentState   DialogueState
	CurrentIntent  *Intent
	Turns          []*DialogueTurn
	Slots          map[string]*Slot
	Variables      map[string]interface{}
	Metadata       map[string]interface{}
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ExpiresAt      time.Time
}

// DialoguePolicy 对话策略
type DialoguePolicy struct {
	PolicyID    string
	Name        string
	Description string
	Conditions  []PolicyCondition
	Actions     []PolicyAction
	Priority    int
}

// PolicyCondition 策略条件
type PolicyCondition struct {
	Type     string // "intent", "slot", "state", "context"
	Key      string
	Operator string // "eq", "ne", "exists", "not_exists"
	Value    interface{}
}

// PolicyAction 策略动作
type PolicyAction struct {
	Type       string // "response", "query", "control", "transfer"
	Content    string
	Template   string
	Parameters map[string]interface{}
}

// DialogueResponse 对话响应
type DialogueResponse struct {
	Content      string
	Suggestions  []string
	Actions      []PolicyAction
	Confidence   float64
	RequiresMore bool // 是否需要更多信息
	Metadata     map[string]interface{}
}

// DialogueManager 对话管理器
type DialogueManager struct {
	sessions     map[string]*DialogueContext
	policies     []*DialoguePolicy
	recognizer   *IntentRecognizer
	config       *DialogueConfig
	mu           sync.RWMutex
}

// DialogueConfig 对话配置
type DialogueConfig struct {
	MaxTurns          int           // 最大对话轮次
	SessionTimeout    time.Duration // 会话超时时间
	MaxSessionAge     time.Duration // 会话最大存活时间
	EnableAutoExpire  bool          // 启用自动过期
	ContextWindowSize int           // 上下文窗口大小
}

// DefaultDialogueConfig 默认对话配置
func DefaultDialogueConfig() *DialogueConfig {
	return &DialogueConfig{
		MaxTurns:          50,
		SessionTimeout:    30 * time.Minute,
		MaxSessionAge:     24 * time.Hour,
		EnableAutoExpire:  true,
		ContextWindowSize: 10,
	}
}

// NewDialogueManager 创建对话管理器
func NewDialogueManager(recognizer *IntentRecognizer, config *DialogueConfig) *DialogueManager {
	if config == nil {
		config = DefaultDialogueConfig()
	}

	dm := &DialogueManager{
		sessions:   make(map[string]*DialogueContext),
		policies:   make([]*DialoguePolicy, 0),
		recognizer: recognizer,
		config:     config,
	}

	// 初始化默认策略
	dm.initDefaultPolicies()

	return dm
}

// initDefaultPolicies 初始化默认对话策略
func (dm *DialogueManager) initDefaultPolicies() {
	// 查询意图策略
	dm.policies = append(dm.policies, &DialoguePolicy{
		PolicyID:    "policy_query_001",
		Name:        "查询意图处理策略",
		Description: "处理用户查询意图的默认策略",
		Conditions: []PolicyCondition{
			{Type: "intent", Key: "type", Operator: "eq", Value: IntentQuery},
		},
		Actions: []PolicyAction{
			{Type: "response", Template: "query_response"},
		},
		Priority: 10,
	})

	// 控制意图策略
	dm.policies = append(dm.policies, &DialoguePolicy{
		PolicyID:    "policy_control_001",
		Name:        "控制意图处理策略",
		Description: "处理用户控制意图的默认策略",
		Conditions: []PolicyCondition{
			{Type: "intent", Key: "type", Operator: "eq", Value: IntentControl},
		},
		Actions: []PolicyAction{
			{Type: "response", Template: "control_confirm"},
		},
		Priority: 10,
	})

	// 配置意图策略
	dm.policies = append(dm.policies, &DialoguePolicy{
		PolicyID:    "policy_config_001",
		Name:        "配置意图处理策略",
		Description: "处理用户配置意图的默认策略",
		Conditions: []PolicyCondition{
			{Type: "intent", Key: "type", Operator: "eq", Value: IntentConfig},
		},
		Actions: []PolicyAction{
			{Type: "response", Template: "config_guide"},
		},
		Priority: 10,
	})

	// 诊断意图策略
	dm.policies = append(dm.policies, &DialoguePolicy{
		PolicyID:    "policy_diagnose_001",
		Name:        "诊断意图处理策略",
		Description: "处理用户诊断意图的默认策略",
		Conditions: []PolicyCondition{
			{Type: "intent", Key: "type", Operator: "eq", Value: IntentDiagnose},
		},
		Actions: []PolicyAction{
			{Type: "response", Template: "diagnose_analysis"},
		},
		Priority: 10,
	})

	// 槽位缺失策略
	dm.policies = append(dm.policies, &DialoguePolicy{
		PolicyID:    "policy_slot_missing_001",
		Name:        "槽位缺失处理策略",
		Description: "处理必填槽位缺失的情况",
		Conditions: []PolicyCondition{
			{Type: "slot", Key: "required_missing", Operator: "exists", Value: true},
		},
		Actions: []PolicyAction{
			{Type: "response", Template: "slot_prompt"},
		},
		Priority: 20, // 更高优先级
	})

	// 未知意图策略
	dm.policies = append(dm.policies, &DialoguePolicy{
		PolicyID:    "policy_unknown_001",
		Name:        "未知意图处理策略",
		Description: "处理无法识别的意图",
		Conditions: []PolicyCondition{
			{Type: "intent", Key: "type", Operator: "eq", Value: IntentUnknown},
		},
		Actions: []PolicyAction{
			{Type: "response", Template: "unknown_intent"},
		},
		Priority: 5,
	})
}

// StartSession 开始新会话
func (dm *DialogueManager) StartSession(ctx context.Context, userID string) (*DialogueContext, error) {
	sessionID := generateSessionID()

	now := time.Now()
	dialogueCtx := &DialogueContext{
		SessionID:    sessionID,
		UserID:       userID,
		CurrentState: StateInitial,
		Turns:        make([]*DialogueTurn, 0),
		Slots:        make(map[string]*Slot),
		Variables:    make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
		CreatedAt:    now,
		UpdatedAt:    now,
		ExpiresAt:    now.Add(dm.config.MaxSessionAge),
	}

	dm.mu.Lock()
	dm.sessions[sessionID] = dialogueCtx
	dm.mu.Unlock()

	return dialogueCtx, nil
}

// Process 处理用户输入
func (dm *DialogueManager) Process(ctx context.Context, sessionID string, userInput string) (*DialogueResponse, error) {
	// 获取会话
	dialogueCtx, err := dm.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// 检查会话是否过期
	if time.Now().After(dialogueCtx.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	// 识别意图
	intent, err := dm.recognizer.Recognize(ctx, userInput)
	if err != nil {
		return nil, fmt.Errorf("intent recognition failed: %w", err)
	}

	// 更新会话状态
	dm.mu.Lock()
	dialogueCtx.CurrentState = StateActive
	dialogueCtx.CurrentIntent = intent
	dialogueCtx.UpdatedAt = time.Now()

	// 合并槽位
	dm.mergeSlots(dialogueCtx, intent)

	// 添加用户轮次
	userTurn := &DialogueTurn{
		TurnID:    generateTurnID(),
		Role:      "user",
		Content:   userInput,
		Intent:    intent,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	dialogueCtx.Turns = append(dialogueCtx.Turns, userTurn)
	dm.mu.Unlock()

	// 应用对话策略
	response := dm.applyPolicy(dialogueCtx, intent)

	// 添加系统轮次
	systemTurn := &DialogueTurn{
		TurnID:    generateTurnID(),
		Role:      "system",
		Content:   response.Content,
		Timestamp: time.Now(),
		Metadata:  response.Metadata,
	}

	dm.mu.Lock()
	dialogueCtx.Turns = append(dialogueCtx.Turns, systemTurn)

	// 限制上下文窗口大小
	if len(dialogueCtx.Turns) > dm.config.ContextWindowSize {
		dialogueCtx.Turns = dialogueCtx.Turns[len(dialogueCtx.Turns)-dm.config.ContextWindowSize:]
	}

	// 更新状态
	if response.RequiresMore {
		dialogueCtx.CurrentState = StateWaiting
	} else {
		dialogueCtx.CurrentState = StateActive
	}
	dialogueCtx.UpdatedAt = time.Now()
	dm.mu.Unlock()

	return response, nil
}

// GetSession 获取会话
func (dm *DialogueManager) GetSession(sessionID string) (*DialogueContext, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	session, exists := dm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// EndSession 结束会话
func (dm *DialogueManager) EndSession(sessionID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	session, exists := dm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.CurrentState = StateCompleted
	session.UpdatedAt = time.Now()

	return nil
}

// CancelSession 取消会话
func (dm *DialogueManager) CancelSession(sessionID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	session, exists := dm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.CurrentState = StateCancelled
	session.UpdatedAt = time.Now()

	return nil
}

// DeleteSession 删除会话
func (dm *DialogueManager) DeleteSession(sessionID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if _, exists := dm.sessions[sessionID]; !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	delete(dm.sessions, sessionID)
	return nil
}

// mergeSlots 合并槽位
func (dm *DialogueManager) mergeSlots(dialogueCtx *DialogueContext, intent *Intent) {
	for slotName, slot := range intent.Slots {
		if slot.Filled {
			if existingSlot, exists := dialogueCtx.Slots[slotName]; exists {
				// 更新已存在的槽位
				existingSlot.Value = slot.Value
				existingSlot.Filled = true
				existingSlot.Entities = slot.Entities
			} else {
				// 添加新槽位
				dialogueCtx.Slots[slotName] = slot
			}
		}
	}
}

// applyPolicy 应用对话策略
func (dm *DialogueManager) applyPolicy(dialogueCtx *DialogueContext, intent *Intent) *DialogueResponse {
	// 检查槽位缺失
	missingSlots := dm.recognizer.GetMissingSlots(intent)
	if len(missingSlots) > 0 {
		return dm.handleMissingSlots(dialogueCtx, intent, missingSlots)
	}

	// 查找匹配的策略
	var matchedPolicy *DialoguePolicy
	for _, policy := range dm.policies {
		if dm.matchPolicyConditions(policy, dialogueCtx, intent) {
			if matchedPolicy == nil || policy.Priority > matchedPolicy.Priority {
				matchedPolicy = policy
			}
		}
	}

	// 执行策略动作
	if matchedPolicy != nil {
		return dm.executePolicyActions(matchedPolicy, dialogueCtx, intent)
	}

	// 默认响应
	return &DialogueResponse{
		Content:      "我理解了您的请求，正在为您处理...",
		Confidence:   intent.Confidence,
		RequiresMore: false,
		Metadata:     make(map[string]interface{}),
	}
}

// matchPolicyConditions 匹配策略条件
func (dm *DialogueManager) matchPolicyConditions(policy *DialoguePolicy, dialogueCtx *DialogueContext, intent *Intent) bool {
	for _, condition := range policy.Conditions {
		if !dm.matchCondition(condition, dialogueCtx, intent) {
			return false
		}
	}
	return true
}

// matchCondition 匹配条件
func (dm *DialogueManager) matchCondition(condition PolicyCondition, dialogueCtx *DialogueContext, intent *Intent) bool {
	var value interface{}

	switch condition.Type {
	case "intent":
		value = dm.getIntentValue(condition.Key, intent)
	case "slot":
		value = dm.getSlotValue(condition.Key, dialogueCtx, intent)
	case "state":
		value = dialogueCtx.CurrentState
	case "context":
		value = dialogueCtx.Variables[condition.Key]
	default:
		return false
	}

	return dm.compareValue(value, condition.Operator, condition.Value)
}

// getIntentValue 获取意图值
func (dm *DialogueManager) getIntentValue(key string, intent *Intent) interface{} {
	if intent == nil {
		return nil
	}

	switch key {
	case "type":
		return intent.Type
	case "name":
		return intent.Name
	case "confidence":
		return intent.Confidence
	default:
		return nil
	}
}

// getSlotValue 获取槽位值
func (dm *DialogueManager) getSlotValue(key string, dialogueCtx *DialogueContext, intent *Intent) interface{} {
	switch key {
	case "required_missing":
		missingSlots := dm.recognizer.GetMissingSlots(intent)
		return len(missingSlots) > 0
	default:
		if slot, exists := dialogueCtx.Slots[key]; exists {
			return slot.Value
		}
		if intent != nil {
			if slot, exists := intent.Slots[key]; exists {
				return slot.Value
			}
		}
		return nil
	}
}

// compareValue 比较值
func (dm *DialogueManager) compareValue(value interface{}, operator string, expected interface{}) bool {
	switch operator {
	case "eq":
		return value == expected
	case "ne":
		return value != expected
	case "exists":
		return value != nil
	case "not_exists":
		return value == nil
	default:
		return false
	}
}

// executePolicyActions 执行策略动作
func (dm *DialogueManager) executePolicyActions(policy *DialoguePolicy, dialogueCtx *DialogueContext, intent *Intent) *DialogueResponse {
	response := &DialogueResponse{
		Actions:    policy.Actions,
		Confidence: intent.Confidence,
		Metadata:   make(map[string]interface{}),
	}

	// 生成响应内容
	for _, action := range policy.Actions {
		if action.Type == "response" {
			response.Content = dm.generateResponse(action.Template, dialogueCtx, intent)
		}
	}

	return response
}

// handleMissingSlots 处理槽位缺失
func (dm *DialogueManager) handleMissingSlots(dialogueCtx *DialogueContext, intent *Intent, missingSlots []string) *DialogueResponse {
	if len(missingSlots) == 0 {
		return &DialogueResponse{
			Content:    "好的，我已经理解您的请求。",
			Confidence: intent.Confidence,
		}
	}

	// 获取第一个缺失槽位的提示
	slotName := missingSlots[0]
	prompt := dm.recognizer.GetSlotPrompt(intent, slotName)

	if prompt == "" {
		prompt = fmt.Sprintf("请提供%s信息", slotName)
	}

	// 生成建议
	suggestions := dm.generateSuggestions(intent, slotName)

	return &DialogueResponse{
		Content:      prompt,
		Suggestions:  suggestions,
		Confidence:   intent.Confidence,
		RequiresMore: true,
		Metadata: map[string]interface{}{
			"missing_slots": missingSlots,
			"current_slot":  slotName,
		},
	}
}

// generateResponse 生成响应
func (dm *DialogueManager) generateResponse(template string, dialogueCtx *DialogueContext, intent *Intent) string {
	switch template {
	case "query_response":
		return dm.generateQueryResponse(dialogueCtx, intent)
	case "control_confirm":
		return dm.generateControlResponse(dialogueCtx, intent)
	case "config_guide":
		return dm.generateConfigResponse(dialogueCtx, intent)
	case "diagnose_analysis":
		return dm.generateDiagnoseResponse(dialogueCtx, intent)
	case "slot_prompt":
		return dm.generateSlotPrompt(dialogueCtx, intent)
	case "unknown_intent":
		return "抱歉，我没有理解您的意思。您可以尝试更具体地描述您的需求，比如：\n- 查询某个设备的实时数据\n- 设置告警阈值\n- 诊断设备故障"
	default:
		return "我正在为您处理请求..."
	}
}

// generateQueryResponse 生成查询响应
func (dm *DialogueManager) generateQueryResponse(dialogueCtx *DialogueContext, intent *Intent) string {
	var target string
	if slot, exists := intent.Slots["target"]; exists && slot.Filled {
		target = fmt.Sprintf("%v", slot.Value)
	} else {
		target = "相关设备"
	}

	switch intent.Name {
	case "query_realtime":
		return fmt.Sprintf("好的，正在为您查询%s的实时数据...", target)
	case "query_history":
		return fmt.Sprintf("好的，正在为您查询%s的历史数据...", target)
	case "query_statistics":
		return fmt.Sprintf("好的，正在为您统计%s的数据...", target)
	default:
		return fmt.Sprintf("好的，正在为您查询%s的相关信息...", target)
	}
}

// generateControlResponse 生成控制响应
func (dm *DialogueManager) generateControlResponse(dialogueCtx *DialogueContext, intent *Intent) string {
	var device string
	var action string

	if slot, exists := intent.Slots["device"]; exists && slot.Filled {
		device = fmt.Sprintf("%v", slot.Value)
	}
	if slot, exists := intent.Slots["action"]; exists && slot.Filled {
		action = fmt.Sprintf("%v", slot.Value)
	}

	if device != "" && action != "" {
		return fmt.Sprintf("确认要%s设备%s吗？", action, device)
	} else if action != "" {
		return fmt.Sprintf("确认要执行%s操作吗？", action)
	}
	return "请确认您的操作意图。"
}

// generateConfigResponse 生成配置响应
func (dm *DialogueManager) generateConfigResponse(dialogueCtx *DialogueContext, intent *Intent) string {
	return "好的，我来帮您进行配置。请告诉我您想要配置的具体内容。"
}

// generateDiagnoseResponse 生成诊断响应
func (dm *DialogueManager) generateDiagnoseResponse(dialogueCtx *DialogueContext, intent *Intent) string {
	var target string
	if slot, exists := intent.Slots["target"]; exists && slot.Filled {
		target = fmt.Sprintf("%v", slot.Value)
	} else {
		target = "设备"
	}

	return fmt.Sprintf("好的，正在对%s进行诊断分析...", target)
}

// generateSlotPrompt 生成槽位提示
func (dm *DialogueManager) generateSlotPrompt(dialogueCtx *DialogueContext, intent *Intent) string {
	missingSlots := dm.recognizer.GetMissingSlots(intent)
	if len(missingSlots) > 0 {
		return dm.recognizer.GetSlotPrompt(intent, missingSlots[0])
	}
	return "请提供更多信息。"
}

// generateSuggestions 生成建议
func (dm *DialogueManager) generateSuggestions(intent *Intent, slotName string) []string {
	suggestions := make([]string, 0)

	switch slotName {
	case "target":
		suggestions = []string{
			"逆变器INV-001",
			"汇流箱CB-001",
			"1号光伏组件",
		}
	case "startTime":
		suggestions = []string{
			"今天",
			"昨天",
			"最近7天",
			"本月",
		}
	case "metric":
		suggestions = []string{
			"发电量",
			"功率",
			"电压",
			"温度",
		}
	case "threshold":
		suggestions = []string{
			"100kW",
			"500V",
			"50℃",
		}
	}

	return suggestions
}

// AddPolicy 添加对话策略
func (dm *DialogueManager) AddPolicy(policy *DialoguePolicy) error {
	if policy == nil {
		return fmt.Errorf("policy cannot be nil")
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.policies = append(dm.policies, policy)
	return nil
}

// GetDialogueHistory 获取对话历史
func (dm *DialogueManager) GetDialogueHistory(sessionID string) ([]*DialogueTurn, error) {
	session, err := dm.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// 返回副本
	history := make([]*DialogueTurn, len(session.Turns))
	copy(history, session.Turns)
	return history, nil
}

// SetContextVariable 设置上下文变量
func (dm *DialogueManager) SetContextVariable(sessionID string, key string, value interface{}) error {
	session, err := dm.GetSession(sessionID)
	if err != nil {
		return err
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	session.Variables[key] = value
	session.UpdatedAt = time.Now()
	return nil
}

// GetContextVariable 获取上下文变量
func (dm *DialogueManager) GetContextVariable(sessionID string, key string) (interface{}, error) {
	session, err := dm.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	dm.mu.RLock()
	defer dm.mu.RUnlock()

	value, exists := session.Variables[key]
	if !exists {
		return nil, fmt.Errorf("variable not found: %s", key)
	}

	return value, nil
}

// CleanExpiredSessions 清理过期会话
func (dm *DialogueManager) CleanExpiredSessions() int {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	now := time.Now()
	count := 0

	for sessionID, session := range dm.sessions {
		if now.After(session.ExpiresAt) {
			delete(dm.sessions, sessionID)
			count++
		}
	}

	return count
}

// GetActiveSessions 获取活跃会话数
func (dm *DialogueManager) GetActiveSessions() int {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	count := 0
	now := time.Now()

	for _, session := range dm.sessions {
		if now.Before(session.ExpiresAt) && session.CurrentState != StateCompleted && session.CurrentState != StateCancelled {
			count++
		}
	}

	return count
}

// generateSessionID 生成会话ID
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// generateTurnID 生成轮次ID
func generateTurnID() string {
	return fmt.Sprintf("turn_%d", time.Now().UnixNano())
}
