package operation

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// OperationType 操作类型
type OperationType string

const (
	OperationTypeRemoteControl OperationType = "remote_control" // 遥控操作
	OperationTypeSetPoint      OperationType = "setpoint"       // 设点操作
	OperationTypeAdjust        OperationType = "adjust"         // 调节操作
	OperationTypeQuery         OperationType = "query"          // 查询操作
	OperationTypeBatch         OperationType = "batch"          // 批量操作
)

// OperationPriority 操作优先级
type OperationPriority int

const (
	PriorityLow      OperationPriority = 1
	PriorityNormal   OperationPriority = 5
	PriorityHigh     OperationPriority = 8
	PriorityCritical OperationPriority = 10
)

// OperationStatus 操作状态
type OperationStatus string

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

// ParsedOperation 解析后的操作
type ParsedOperation struct {
	ID           string                 `json:"id"`
	Type         OperationType          `json:"type"`
	Priority     OperationPriority      `json:"priority"`
	TargetType   string                 `json:"target_type"`   // device, point, station
	TargetID     string                 `json:"target_id"`     // 目标ID
	TargetName   string                 `json:"target_name"`   // 目标名称
	Action       string                 `json:"action"`        // 具体动作
	Parameters   map[string]interface{} `json:"parameters"`    // 操作参数
	Constraints  *OperationConstraints  `json:"constraints"`   // 操作约束
	Description  string                 `json:"description"`   // 操作描述
	OriginalText string                 `json:"original_text"` // 原始文本
	Confidence   float64                `json:"confidence"`    // 解析置信度
	CreatedAt    time.Time              `json:"created_at"`
}

// OperationConstraints 操作约束
type OperationConstraints struct {
	MinValue        float64 `json:"min_value,omitempty"`
	MaxValue        float64 `json:"max_value,omitempty"`
	AllowedValues   []string `json:"allowed_values,omitempty"`
	RequireConfirm  bool     `json:"require_confirm"`
	RequireAuth     bool     `json:"require_auth"`
	AuthLevel       int      `json:"auth_level"`
	Timeout         time.Duration `json:"timeout"`
	MaxRetries      int      `json:"max_retries"`
	AllowRollback   bool     `json:"allow_rollback"`
	DryRun          bool     `json:"dry_run"` // 试运行模式
}

// ParseResult 解析结果
type ParseResult struct {
	Operations []*ParsedOperation `json:"operations"`
	Warnings   []string           `json:"warnings"`
	Suggestions []string          `json:"suggestions"`
}

// OperationParser 操作解析器
type OperationParser struct {
	keywords      map[OperationType][]string
	actionPatterns map[string]*regexp.Regexp
	safetyChecker *SafetyChecker
}

// SafetyChecker 安全校验器
type SafetyChecker struct {
	dangerousPatterns []string
	protectedTargets  map[string]bool
}

// NewSafetyChecker 创建安全校验器
func NewSafetyChecker() *SafetyChecker {
	return &SafetyChecker{
		dangerousPatterns: []string{
			"shutdown", "关闭", "停机", "断电",
			"delete", "删除", "清除",
			"reset", "重置", "复位",
			"emergency", "紧急", "急停",
		},
		protectedTargets: make(map[string]bool),
	}
}

// NewOperationParser 创建操作解析器
func NewOperationParser() *OperationParser {
	return &OperationParser{
		keywords: map[OperationType][]string{
			OperationTypeRemoteControl: {
				"遥控", "控制", "开关", "启停", "启动", "停止",
				"合闸", "分闸", "remote", "control", "switch",
			},
			OperationTypeSetPoint: {
				"设点", "设定", "设置", "配置", "参数",
				"setpoint", "set", "configure", "parameter",
			},
			OperationTypeAdjust: {
				"调节", "调整", "修改", "改变", "增加", "减少",
				"adjust", "modify", "change", "increase", "decrease",
			},
			OperationTypeQuery: {
				"查询", "读取", "获取", "查看", "显示",
				"query", "read", "get", "show", "display",
			},
		},
		actionPatterns: map[string]*regexp.Regexp{
			"switch_on":  regexp.MustCompile(`(?i)(启动|开启|合闸|start|open|on|switch\s*on)`),
			"switch_off": regexp.MustCompile(`(?i)(停止|关闭|分闸|stop|close|off|switch\s*off)`),
			"set_value":  regexp.MustCompile(`(?i)(设置|设定|设为|set\s*to?|=\s*(\d+\.?\d*))`),
			"adjust":     regexp.MustCompile(`(?i)(调整|调节|增加|减少|adjust|increase|decrease)\s*(\d+\.?\d*)`),
			"query":      regexp.MustCompile(`(?i)(查询|读取|获取|query|read|get)`),
		},
		safetyChecker: NewSafetyChecker(),
	}
}

// Parse 解析自然语言指令
func (p *OperationParser) Parse(ctx context.Context, text string) (*ParseResult, error) {
	result := &ParseResult{
		Operations: make([]*ParsedOperation, 0),
		Warnings:   make([]string, 0),
		Suggestions: make([]string, 0),
	}

	// 预处理文本
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("empty operation text")
	}

	// 分割多个操作指令
	commands := p.splitCommands(text)

	for _, cmd := range commands {
		op, warnings := p.parseSingleCommand(ctx, cmd)
		if op != nil {
			result.Operations = append(result.Operations, op)
		}
		result.Warnings = append(result.Warnings, warnings...)
	}

	// 安全检查
	for _, op := range result.Operations {
		if warnings := p.safetyChecker.Check(op); len(warnings) > 0 {
			result.Warnings = append(result.Warnings, warnings...)
		}
	}

	// 生成建议
	result.Suggestions = p.generateSuggestions(result.Operations)

	return result, nil
}

// parseSingleCommand 解析单个命令
func (p *OperationParser) parseSingleCommand(ctx context.Context, text string) (*ParsedOperation, []string) {
	warnings := make([]string, 0)

	op := &ParsedOperation{
		ID:           generateOperationID(),
		Parameters:   make(map[string]interface{}),
		OriginalText: text,
		CreatedAt:    time.Now(),
		Confidence:   0.0,
	}

	// 识别操作类型
	opType, typeConfidence := p.identifyOperationType(text)
	op.Type = opType
	op.Confidence = typeConfidence

	// 识别目标
	targetID, targetType, targetName, targetConfidence := p.identifyTarget(text)
	op.TargetID = targetID
	op.TargetType = targetType
	op.TargetName = targetName
	op.Confidence = (op.Confidence + targetConfidence) / 2

	// 识别动作
	action, actionConfidence := p.identifyAction(text, opType)
	op.Action = action
	op.Confidence = (op.Confidence + actionConfidence) / 2

	// 提取参数
	params, paramConfidence := p.extractParameters(text, opType, action)
	for k, v := range params {
		op.Parameters[k] = v
	}
	if paramConfidence > 0 {
		op.Confidence = (op.Confidence + paramConfidence) / 2
	}

	// 设置默认约束
	op.Constraints = p.getDefaultConstraints(opType, action)

	// 生成描述
	op.Description = p.generateDescription(op)

	// 设置优先级
	op.Priority = p.determinePriority(op)

	// 低置信度警告
	if op.Confidence < 0.6 {
		warnings = append(warnings, fmt.Sprintf("解析置信度较低 (%.2f)，请确认操作是否正确", op.Confidence))
	}

	return op, warnings
}

// identifyOperationType 识别操作类型
func (p *OperationParser) identifyOperationType(text string) (OperationType, float64) {
	scores := make(map[OperationType]float64)

	for opType, keywords := range p.keywords {
		score := 0.0
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(text), strings.ToLower(keyword)) {
				score += 1.0
			}
		}
		scores[opType] = score / float64(len(keywords))
	}

	maxScore := 0.0
	var result OperationType = OperationTypeQuery // 默认查询

	for opType, score := range scores {
		if score > maxScore {
			maxScore = score
			result = opType
		}
	}

	return result, min(maxScore+0.3, 1.0)
}

// identifyTarget 识别操作目标
func (p *OperationParser) identifyTarget(text string) (string, string, string, float64) {
	// 设备编号模式: DEV-001, 设备001
	devicePattern := regexp.MustCompile(`(?i)(设备|DEV|device)[\s\-]*(\d+|[a-zA-Z0-9\-]+)`)
	if matches := devicePattern.FindStringSubmatch(text); len(matches) > 0 {
		return matches[2], "device", matches[0], 0.9
	}

	// 测点编号模式: POINT-001, 测点001
	pointPattern := regexp.MustCompile(`(?i)(测点|点|POINT|point)[\s\-]*(\d+|[a-zA-Z0-9\-]+)`)
	if matches := pointPattern.FindStringSubmatch(text); len(matches) > 0 {
		return matches[2], "point", matches[0], 0.9
	}

	// 电站模式: 电站001
	stationPattern := regexp.MustCompile(`(?i)(电站|站|STATION|station)[\s\-]*(\d+|[a-zA-Z0-9\-]+)`)
	if matches := stationPattern.FindStringSubmatch(text); len(matches) > 0 {
		return matches[2], "station", matches[0], 0.9
	}

	// 逆变器模式
	inverterPattern := regexp.MustCompile(`(?i)(逆变器|INV|inverter)[\s\-]*(\d+|[a-zA-Z0-9\-]+)`)
	if matches := inverterPattern.FindStringSubmatch(text); len(matches) > 0 {
		return matches[2], "device", matches[0], 0.85
	}

	// 默认返回
	return "", "", "", 0.3
}

// identifyAction 识别操作动作
func (p *OperationParser) identifyAction(text string, opType OperationType) (string, float64) {
	for action, pattern := range p.actionPatterns {
		if pattern.MatchString(text) {
			return action, 0.9
		}
	}

	// 根据操作类型返回默认动作
	switch opType {
	case OperationTypeRemoteControl:
		return "control", 0.5
	case OperationTypeSetPoint:
		return "set", 0.5
	case OperationTypeAdjust:
		return "adjust", 0.5
	case OperationTypeQuery:
		return "query", 0.5
	default:
		return "unknown", 0.3
	}
}

// extractParameters 提取参数
func (p *OperationParser) extractParameters(text string, opType OperationType, action string) (map[string]interface{}, float64) {
	params := make(map[string]interface{})
	confidence := 0.0

	// 提取数值参数
	numberPattern := regexp.MustCompile(`(\d+\.?\d*)`)
	if matches := numberPattern.FindAllString(text, -1); len(matches) > 0 {
		if len(matches) == 1 {
			if value, err := strconv.ParseFloat(matches[0], 64); err == nil {
				params["value"] = value
				confidence = 0.8
			}
		} else {
			// 多个数值，尝试识别
			values := make([]float64, 0)
			for _, m := range matches {
				if v, err := strconv.ParseFloat(m, 64); err == nil {
					values = append(values, v)
				}
			}
			if len(values) > 0 {
				params["values"] = values
				confidence = 0.6
			}
		}
	}

	// 提取单位
	unitPattern := regexp.MustCompile(`(kW|MW|kWh|V|A|Hz|℃|%|度)`)
	if matches := unitPattern.FindStringSubmatch(text); len(matches) > 0 {
		params["unit"] = matches[1]
		confidence = max(confidence, 0.7)
	}

	// 提取时间参数
	timePattern := regexp.MustCompile(`(\d{1,2}):(\d{2})`)
	if matches := timePattern.FindStringSubmatch(text); len(matches) > 0 {
		params["time"] = matches[0]
		confidence = max(confidence, 0.8)
	}

	// 提取延迟参数
	delayPattern := regexp.MustCompile(`(?i)(延迟|延时|delay)\s*(\d+)\s*(秒|分钟|小时|s|m|h)`)
	if matches := delayPattern.FindStringSubmatch(text); len(matches) > 0 {
		params["delay"] = matches[2]
		params["delay_unit"] = matches[3]
		confidence = max(confidence, 0.85)
	}

	return params, confidence
}

// getDefaultConstraints 获取默认约束
func (p *OperationParser) getDefaultConstraints(opType OperationType, action string) *OperationConstraints {
	constraints := &OperationConstraints{
		RequireConfirm: true,
		RequireAuth:    true,
		AuthLevel:      1,
		Timeout:        30 * time.Second,
		MaxRetries:     3,
		AllowRollback:  true,
		DryRun:         false,
	}

	switch opType {
	case OperationTypeRemoteControl:
		constraints.Timeout = 10 * time.Second
		constraints.AuthLevel = 2
		constraints.RequireConfirm = true

	case OperationTypeSetPoint:
		constraints.Timeout = 15 * time.Second
		constraints.AuthLevel = 2
		constraints.RequireConfirm = true

	case OperationTypeAdjust:
		constraints.Timeout = 20 * time.Second
		constraints.AuthLevel = 1
		constraints.RequireConfirm = true

	case OperationTypeQuery:
		constraints.Timeout = 5 * time.Second
		constraints.AuthLevel = 0
		constraints.RequireConfirm = false
		constraints.AllowRollback = false
	}

	// 危险操作需要更高级别授权
	if action == "switch_off" || action == "shutdown" {
		constraints.AuthLevel = 3
		constraints.RequireConfirm = true
	}

	return constraints
}

// generateDescription 生成操作描述
func (p *OperationParser) generateDescription(op *ParsedOperation) string {
	var desc strings.Builder

	switch op.Type {
	case OperationTypeRemoteControl:
		desc.WriteString("遥控操作: ")
	case OperationTypeSetPoint:
		desc.WriteString("设点操作: ")
	case OperationTypeAdjust:
		desc.WriteString("调节操作: ")
	case OperationTypeQuery:
		desc.WriteString("查询操作: ")
	}

	if op.TargetName != "" {
		desc.WriteString(op.TargetName)
	} else if op.TargetID != "" {
		desc.WriteString(op.TargetID)
	}

	desc.WriteString(" - ")
	desc.WriteString(op.Action)

	if value, ok := op.Parameters["value"]; ok {
		desc.WriteString(fmt.Sprintf(" (值: %v", value))
		if unit, ok := op.Parameters["unit"]; ok {
			desc.WriteString(fmt.Sprintf(" %v", unit))
		}
		desc.WriteString(")")
	}

	return desc.String()
}

// determinePriority 确定优先级
func (p *OperationParser) determinePriority(op *ParsedOperation) OperationPriority {
	// 根据操作类型和动作确定优先级
	switch op.Action {
	case "switch_off", "shutdown", "emergency":
		return PriorityCritical
	case "switch_on":
		return PriorityHigh
	}

	switch op.Type {
	case OperationTypeRemoteControl:
		return PriorityHigh
	case OperationTypeSetPoint:
		return PriorityNormal
	case OperationTypeAdjust:
		return PriorityNormal
	case OperationTypeQuery:
		return PriorityLow
	default:
		return PriorityNormal
	}
}

// splitCommands 分割多个命令
func (p *OperationParser) splitCommands(text string) []string {
	separators := []string{"然后", "接着", "并且", "同时", ";", "，", ",", "and", "then"}
	
	result := []string{text}
	for _, sep := range separators {
		var newResult []string
		for _, cmd := range result {
			parts := strings.Split(cmd, sep)
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					newResult = append(newResult, part)
				}
			}
		}
		result = newResult
	}

	return result
}

// generateSuggestions 生成建议
func (p *OperationParser) generateSuggestions(operations []*ParsedOperation) []string {
	suggestions := make([]string, 0)

	for _, op := range operations {
		// 低置信度建议确认
		if op.Confidence < 0.7 {
			suggestions = append(suggestions, 
				fmt.Sprintf("建议确认操作 '%s' 的准确性", op.Description))
		}

		// 缺少目标建议
		if op.TargetID == "" {
			suggestions = append(suggestions,
				fmt.Sprintf("操作 '%s' 未指定明确目标，建议补充目标信息", op.Description))
		}

		// 危险操作建议
		if op.Constraints.AuthLevel >= 3 {
			suggestions = append(suggestions,
				fmt.Sprintf("操作 '%s' 需要高级别授权，请确保有相应权限", op.Description))
		}
	}

	return suggestions
}

// Check 安全校验
func (s *SafetyChecker) Check(op *ParsedOperation) []string {
	warnings := make([]string, 0)

	// 检查危险操作
	text := strings.ToLower(op.OriginalText)
	for _, pattern := range s.dangerousPatterns {
		if strings.Contains(text, strings.ToLower(pattern)) {
			warnings = append(warnings, 
				fmt.Sprintf("检测到潜在危险操作关键词 '%s'，需要额外确认", pattern))
		}
	}

	// 检查受保护目标
	if s.protectedTargets[op.TargetID] {
		warnings = append(warnings,
			fmt.Sprintf("目标 '%s' 是受保护对象，需要特殊授权", op.TargetID))
	}

	// 检查参数范围
	if op.Constraints != nil {
		if value, ok := op.Parameters["value"].(float64); ok {
			if op.Constraints.MinValue != 0 || op.Constraints.MaxValue != 0 {
				if value < op.Constraints.MinValue || value > op.Constraints.MaxValue {
					warnings = append(warnings,
						fmt.Sprintf("参数值 %.2f 超出允许范围 [%.2f, %.2f]",
							value, op.Constraints.MinValue, op.Constraints.MaxValue))
				}
			}
		}
	}

	return warnings
}

// AddProtectedTarget 添加受保护目标
func (s *SafetyChecker) AddProtectedTarget(targetID string) {
	s.protectedTargets[targetID] = true
}

// RemoveProtectedTarget 移除受保护目标
func (s *SafetyChecker) RemoveProtectedTarget(targetID string) {
	delete(s.protectedTargets, targetID)
}

// generateOperationID 生成操作ID
func generateOperationID() string {
	return fmt.Sprintf("OP-%d", time.Now().UnixNano())
}

// ValidateOperation 验证操作
func (p *OperationParser) ValidateOperation(op *ParsedOperation) error {
	if op.Type == "" {
		return errors.New("operation type is required")
	}

	if op.TargetID == "" && op.Type != OperationTypeQuery {
		return errors.New("target ID is required for non-query operations")
	}

	if op.Action == "" {
		return errors.New("action is required")
	}

	// 验证参数
	if value, ok := op.Parameters["value"]; ok {
		if _, ok := value.(float64); !ok {
			return errors.New("value parameter must be a number")
		}
	}

	return nil
}
