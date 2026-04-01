package processor

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// QualityCode 质量码定义
// 使用位域表示不同维度的质量问题
type QualityCode uint16

const (
	// 基础质量码
	QualityGood         QualityCode = 0x0000 // 质量好
	QualityBad          QualityCode = 0x8000 // 质量坏
	QualityUncertain    QualityCode = 0x4000 // 质量不确定
	QualityQuestionable QualityCode = 0x2000 // 质量可疑

	// 质量原因码 (低8位)
	QualityReasonNone           QualityCode = 0x0000 // 无原因
	QualityReasonOutOfRange     QualityCode = 0x0001 // 超出范围
	QualityReasonNull           QualityCode = 0x0002 // 空值
	QualityReasonTypeMismatch   QualityCode = 0x0004 // 类型不匹配
	QualityReasonTimeout        QualityCode = 0x0008 // 超时
	QualityReasonCommunication  QualityCode = 0x0010 // 通信故障
	QualityReasonDeviceFailure  QualityCode = 0x0020 // 设备故障
	QualityReasonSensorFailure  QualityCode = 0x0040 // 传感器故障
	QualityReasonConfiguration  QualityCode = 0x0080 // 配置错误
	QualityReasonInterpolation  QualityCode = 0x0100 // 插值数据
	QualityReasonFiltered       QualityCode = 0x0200 // 已滤波
	QualityReasonDebounced      QualityCode = 0x0400 // 已防抖
	QualityReasonScaled         QualityCode = 0x0800 // 已量程转换
	QualityReasonOverflow       QualityCode = 0x1000 // 溢出
	QualityReasonUnderflow      QualityCode = 0x2000 // 下溢
	QualityReasonCalculated     QualityCode = 0x4000 // 计算值
	QualityReasonManual         QualityCode = 0x8000 // 手动输入
)

// QualityLevel 质量等级
type QualityLevel int

const (
	QualityLevelGood    QualityLevel = iota // 好
	QualityLevelFair                        // 一般
	QualityLevelPoor                        // 差
	QualityLevelBad                         // 坏
)

// QualityInfo 质量信息
type QualityInfo struct {
	Code        QualityCode   `json:"code"`         // 质量码
	Level       QualityLevel  `json:"level"`        // 质量等级
	Reasons     []string      `json:"reasons"`      // 质量原因列表
	Timestamp   time.Time     `json:"timestamp"`    // 时间戳
	Source      string        `json:"source"`       // 数据源
	Description string        `json:"description"`  // 描述信息
}

// QualityMarker 质量标记器接口
type QualityMarker interface {
	// Mark 标记数据质量
	Mark(value interface{}, code QualityCode, reason string) QualityInfo
	// Evaluate 评估数据质量
	Evaluate(value interface{}) QualityInfo
	// Combine 组合质量码
	Combine(codes ...QualityCode) QualityCode
	// GetLevel 获取质量等级
	GetLevel(code QualityCode) QualityLevel
}

// DefaultQualityMarker 默认质量标记器
type DefaultQualityMarker struct {
	mu       sync.RWMutex
	history  map[string][]QualityInfo // 历史质量记录
	maxHistory int                    // 最大历史记录数
	logger   *zap.Logger
}

// NewQualityMarker 创建质量标记器
func NewQualityMarker(logger *zap.Logger) *DefaultQualityMarker {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &DefaultQualityMarker{
		history:    make(map[string][]QualityInfo),
		maxHistory: 100,
		logger:     logger,
	}
}

// Mark 标记数据质量
func (m *DefaultQualityMarker) Mark(value interface{}, code QualityCode, reason string) QualityInfo {
	info := QualityInfo{
		Code:      code,
		Level:     m.GetLevel(code),
		Reasons:   []string{reason},
		Timestamp: time.Now(),
	}
	
	m.logger.Debug("quality marked",
		zap.Any("value", value),
		zap.Uint16("code", uint16(code)),
		zap.String("reason", reason),
	)
	
	return info
}

// Evaluate 评估数据质量
func (m *DefaultQualityMarker) Evaluate(value interface{}) QualityInfo {
	code := QualityGood
	reasons := make([]string, 0)
	
	// 检查空值
	if value == nil {
		code |= QualityBad | QualityReasonNull
		reasons = append(reasons, "value is null")
		return QualityInfo{
			Code:      code,
			Level:     m.GetLevel(code),
			Reasons:   reasons,
			Timestamp: time.Now(),
		}
	}
	
	// 根据类型进行评估
	switch v := value.(type) {
	case float64:
		// 检查特殊值
		if v != v { // NaN
			code |= QualityBad | QualityReasonSensorFailure
			reasons = append(reasons, "value is NaN")
		}
		if v == 0 {
			code |= QualityUncertain
			reasons = append(reasons, "value is zero")
		}
	case float32:
		if v != v { // NaN
			code |= QualityBad | QualityReasonSensorFailure
			reasons = append(reasons, "value is NaN")
		}
	case int, int32, int64:
		// 整数类型检查
	case string:
		if v == "" {
			code |= QualityBad | QualityReasonNull
			reasons = append(reasons, "string is empty")
		}
	case bool:
		// 布尔类型不需要特殊检查
	default:
		code |= QualityUncertain | QualityReasonTypeMismatch
		reasons = append(reasons, fmt.Sprintf("unsupported type: %T", value))
	}
	
	return QualityInfo{
		Code:      code,
		Level:     m.GetLevel(code),
		Reasons:   reasons,
		Timestamp: time.Now(),
	}
}

// Combine 组合质量码
func (m *DefaultQualityMarker) Combine(codes ...QualityCode) QualityCode {
	if len(codes) == 0 {
		return QualityGood
	}
	
	var result QualityCode
	for _, code := range codes {
		result |= code
	}
	
	// 确定最终质量状态
	if result&QualityBad != 0 {
		result = (result & 0x0FFF) | QualityBad
	} else if result&QualityUncertain != 0 {
		result = (result & 0x0FFF) | QualityUncertain
	} else if result&QualityQuestionable != 0 {
		result = (result & 0x0FFF) | QualityQuestionable
	}
	
	return result
}

// GetLevel 获取质量等级
func (m *DefaultQualityMarker) GetLevel(code QualityCode) QualityLevel {
	if code&QualityBad != 0 {
		return QualityLevelBad
	}
	if code&QualityUncertain != 0 {
		return QualityLevelPoor
	}
	if code&QualityQuestionable != 0 {
		return QualityLevelFair
	}
	return QualityLevelGood
}

// RecordQuality 记录质量信息
func (m *DefaultQualityMarker) RecordQuality(pointID string, info QualityInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	history := m.history[pointID]
	history = append(history, info)
	
	// 限制历史记录数量
	if len(history) > m.maxHistory {
		history = history[len(history)-m.maxHistory:]
	}
	
	m.history[pointID] = history
}

// GetHistory 获取历史质量记录
func (m *DefaultQualityMarker) GetHistory(pointID string) []QualityInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	history, ok := m.history[pointID]
	if !ok {
		return []QualityInfo{}
	}
	
	result := make([]QualityInfo, len(history))
	copy(result, history)
	return result
}

// GetStatistics 获取质量统计
func (m *DefaultQualityMarker) GetStatistics(pointID string) QualityStatistics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats := QualityStatistics{
		PointID: pointID,
	}
	
	history, ok := m.history[pointID]
	if !ok {
		return stats
	}
	
	stats.TotalCount = len(history)
	
	for _, info := range history {
		switch info.Level {
		case QualityLevelGood:
			stats.GoodCount++
		case QualityLevelFair:
			stats.FairCount++
		case QualityLevelPoor:
			stats.PoorCount++
		case QualityLevelBad:
			stats.BadCount++
		}
	}
	
	if stats.TotalCount > 0 {
		stats.GoodRate = float64(stats.GoodCount) / float64(stats.TotalCount) * 100
	}
	
	return stats
}

// QualityStatistics 质量统计
type QualityStatistics struct {
	PointID    string  `json:"point_id"`
	TotalCount int     `json:"total_count"`
	GoodCount  int     `json:"good_count"`
	FairCount  int     `json:"fair_count"`
	PoorCount  int     `json:"poor_count"`
	BadCount   int     `json:"bad_count"`
	GoodRate   float64 `json:"good_rate"` // 质量好率(%)
}

// QualityChecker 质量检查器
type QualityChecker struct {
	marker QualityMarker
	rules  []QualityCheckRule
	logger *zap.Logger
}

// QualityCheckRule 质量检查规则
type QualityCheckRule struct {
	Name        string                   // 规则名称
	Description string                   // 规则描述
	Check       func(interface{}) bool   // 检查函数
	Code        QualityCode              // 失败时的质量码
	Reason      string                   // 失败原因
}

// NewQualityChecker 创建质量检查器
func NewQualityChecker(marker QualityMarker, logger *zap.Logger) *QualityChecker {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &QualityChecker{
		marker: marker,
		rules:  make([]QualityCheckRule, 0),
		logger: logger,
	}
}

// AddRule 添加检查规则
func (c *QualityChecker) AddRule(rule QualityCheckRule) {
	c.rules = append(c.rules, rule)
}

// Check 执行质量检查
func (c *QualityChecker) Check(value interface{}) QualityInfo {
	codes := make([]QualityCode, 0)
	reasons := make([]string, 0)
	
	for _, rule := range c.rules {
		if !rule.Check(value) {
			codes = append(codes, rule.Code)
			reasons = append(reasons, rule.Reason)
			
			c.logger.Debug("quality check failed",
				zap.String("rule", rule.Name),
				zap.Any("value", value),
				zap.String("reason", rule.Reason),
			)
		}
	}
	
	code := c.marker.Combine(codes...)
	return QualityInfo{
		Code:      code,
		Level:     c.marker.GetLevel(code),
		Reasons:   reasons,
		Timestamp: time.Now(),
	}
}

// QualityCodeString 质量码字符串表示
func QualityCodeString(code QualityCode) string {
	var parts []string
	
	// 质量状态
	switch {
	case code&QualityBad != 0:
		parts = append(parts, "BAD")
	case code&QualityUncertain != 0:
		parts = append(parts, "UNCERTAIN")
	case code&QualityQuestionable != 0:
		parts = append(parts, "QUESTIONABLE")
	default:
		parts = append(parts, "GOOD")
	}
	
	// 质量原因
	reasons := []struct {
		code   QualityCode
		name   string
	}{
		{QualityReasonOutOfRange, "OUT_OF_RANGE"},
		{QualityReasonNull, "NULL"},
		{QualityReasonTypeMismatch, "TYPE_MISMATCH"},
		{QualityReasonTimeout, "TIMEOUT"},
		{QualityReasonCommunication, "COMMUNICATION"},
		{QualityReasonDeviceFailure, "DEVICE_FAILURE"},
		{QualityReasonSensorFailure, "SENSOR_FAILURE"},
		{QualityReasonConfiguration, "CONFIGURATION"},
		{QualityReasonInterpolation, "INTERPOLATION"},
		{QualityReasonFiltered, "FILTERED"},
		{QualityReasonDebounced, "DEBOUNCED"},
		{QualityReasonScaled, "SCALED"},
		{QualityReasonOverflow, "OVERFLOW"},
		{QualityReasonUnderflow, "UNDERFLOW"},
		{QualityReasonCalculated, "CALCULATED"},
		{QualityReasonManual, "MANUAL"},
	}
	
	for _, r := range reasons {
		if code&r.code != 0 {
			parts = append(parts, r.name)
		}
	}
	
	return fmt.Sprintf("%v", parts)
}
