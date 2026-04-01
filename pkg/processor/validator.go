package processor

import (
	"fmt"
	"math"
	"reflect"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ValidationResult 校验结果
type ValidationResult struct {
	Valid     bool     `json:"valid"`      // 是否有效
	Errors    []string `json:"errors"`     // 错误信息
	Warnings  []string `json:"warnings"`   // 警告信息
	Quality   QualityCode `json:"quality"` // 质量码
	Timestamp time.Time `json:"timestamp"`  // 时间戳
}

// Validator 校验器接口
type Validator interface {
	// Validate 校验数据
	Validate(value interface{}) ValidationResult
	// Name 获取校验器名称
	Name() string
}

// RangeValidator 范围校验器
type RangeValidator struct {
	name      string
	minValue  float64
	maxValue  float64
	inclusive bool // 是否包含边界
	logger    *zap.Logger
}

// RangeValidatorConfig 范围校验器配置
type RangeValidatorConfig struct {
	Name      string
	MinValue  float64
	MaxValue  float64
	Inclusive bool
	Logger    *zap.Logger
}

// NewRangeValidator 创建范围校验器
func NewRangeValidator(config RangeValidatorConfig) *RangeValidator {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	return &RangeValidator{
		name:      config.Name,
		minValue:  config.MinValue,
		maxValue:  config.MaxValue,
		inclusive: config.Inclusive,
		logger:    config.Logger,
	}
}

// Validate 校验数据
func (v *RangeValidator) Validate(value interface{}) ValidationResult {
	result := ValidationResult{
		Valid:     true,
		Errors:    make([]string, 0),
		Warnings:  make([]string, 0),
		Timestamp: time.Now(),
	}
	
	// 转换为float64
	floatVal, ok := toFloat64(value)
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("cannot convert value to float64: %v", value))
		result.Quality = QualityBad | QualityReasonTypeMismatch
		return result
	}
	
	// 检查NaN和Inf
	if math.IsNaN(floatVal) || math.IsInf(floatVal, 0) {
		result.Valid = false
		result.Errors = append(result.Errors, "value is NaN or Inf")
		result.Quality = QualityBad | QualityReasonSensorFailure
		return result
	}
	
	// 范围检查
	if v.inclusive {
		if floatVal < v.minValue || floatVal > v.maxValue {
			result.Valid = false
			result.Errors = append(result.Errors, 
				fmt.Sprintf("value %.2f is out of range [%.2f, %.2f]", floatVal, v.minValue, v.maxValue))
			result.Quality = QualityBad | QualityReasonOutOfRange
		}
	} else {
		if floatVal <= v.minValue || floatVal >= v.maxValue {
			result.Valid = false
			result.Errors = append(result.Errors, 
				fmt.Sprintf("value %.2f is out of range (%.2f, %.2f)", floatVal, v.minValue, v.maxValue))
			result.Quality = QualityBad | QualityReasonOutOfRange
		}
	}
	
	v.logger.Debug("range validation",
		zap.Float64("value", floatVal),
		zap.Float64("min", v.minValue),
		zap.Float64("max", v.maxValue),
		zap.Bool("valid", result.Valid),
	)
	
	return result
}

// Name 获取校验器名称
func (v *RangeValidator) Name() string {
	return v.name
}

// NullValidator 空值校验器
type NullValidator struct {
	name           string
	allowNull      bool
	allowEmptyStr  bool
	allowZero      bool
	logger         *zap.Logger
}

// NullValidatorConfig 空值校验器配置
type NullValidatorConfig struct {
	Name          string
	AllowNull     bool
	AllowEmptyStr bool
	AllowZero     bool
	Logger        *zap.Logger
}

// NewNullValidator 创建空值校验器
func NewNullValidator(config NullValidatorConfig) *NullValidator {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	return &NullValidator{
		name:          config.Name,
		allowNull:     config.AllowNull,
		allowEmptyStr: config.AllowEmptyStr,
		allowZero:     config.AllowZero,
		logger:        config.Logger,
	}
}

// Validate 校验数据
func (v *NullValidator) Validate(value interface{}) ValidationResult {
	result := ValidationResult{
		Valid:     true,
		Errors:    make([]string, 0),
		Warnings:  make([]string, 0),
		Timestamp: time.Now(),
	}
	
	// 检查nil
	if value == nil {
		if !v.allowNull {
			result.Valid = false
			result.Errors = append(result.Errors, "value is null")
			result.Quality = QualityBad | QualityReasonNull
		}
		return result
	}
	
	// 根据类型检查
	switch val := value.(type) {
	case string:
		if val == "" && !v.allowEmptyStr {
			result.Valid = false
			result.Errors = append(result.Errors, "string is empty")
			result.Quality = QualityBad | QualityReasonNull
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if !v.allowZero {
			floatVal, _ := toFloat64(val)
			if floatVal == 0 {
				result.Warnings = append(result.Warnings, "value is zero")
				result.Quality = QualityUncertain
			}
		}
	case float32, float64:
		if !v.allowZero {
			floatVal, _ := toFloat64(val)
			if floatVal == 0 {
				result.Warnings = append(result.Warnings, "value is zero")
				result.Quality = QualityUncertain
			}
		}
	case bool:
		// 布尔值不需要检查
	case []interface{}:
		if len(val) == 0 && !v.allowEmptyStr {
			result.Warnings = append(result.Warnings, "array is empty")
			result.Quality = QualityUncertain
		}
	case map[string]interface{}:
		if len(val) == 0 && !v.allowEmptyStr {
			result.Warnings = append(result.Warnings, "map is empty")
			result.Quality = QualityUncertain
		}
	default:
		// 使用反射检查
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			if rv.Len() == 0 && !v.allowEmptyStr {
				result.Warnings = append(result.Warnings, "collection is empty")
				result.Quality = QualityUncertain
			}
		}
	}
	
	v.logger.Debug("null validation",
		zap.Any("value", value),
		zap.Bool("valid", result.Valid),
	)
	
	return result
}

// Name 获取校验器名称
func (v *NullValidator) Name() string {
	return v.name
}

// TypeValidator 类型校验器
type TypeValidator struct {
	name       string
	expectType reflect.Type
	logger     *zap.Logger
}

// TypeValidatorConfig 类型校验器配置
type TypeValidatorConfig struct {
	Name       string
	ExpectType reflect.Type
	Logger     *zap.Logger
}

// NewTypeValidator 创建类型校验器
func NewTypeValidator(config TypeValidatorConfig) *TypeValidator {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	return &TypeValidator{
		name:       config.Name,
		expectType: config.ExpectType,
		logger:     config.Logger,
	}
}

// Validate 校验数据
func (v *TypeValidator) Validate(value interface{}) ValidationResult {
	result := ValidationResult{
		Valid:     true,
		Errors:    make([]string, 0),
		Warnings:  make([]string, 0),
		Timestamp: time.Now(),
	}
	
	if value == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "value is nil")
		result.Quality = QualityBad | QualityReasonNull
		return result
	}
	
	actualType := reflect.TypeOf(value)
	if actualType != v.expectType {
		// 尝试类型转换
		if !isConvertible(actualType, v.expectType) {
			result.Valid = false
			result.Errors = append(result.Errors, 
				fmt.Sprintf("type mismatch: expected %v, got %v", v.expectType, actualType))
			result.Quality = QualityBad | QualityReasonTypeMismatch
		} else {
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("type convertible: %v to %v", actualType, v.expectType))
			result.Quality = QualityQuestionable
		}
	}
	
	v.logger.Debug("type validation",
		zap.Any("value", value),
		zap.String("expected", v.expectType.String()),
		zap.String("actual", actualType.String()),
		zap.Bool("valid", result.Valid),
	)
	
	return result
}

// Name 获取校验器名称
func (v *TypeValidator) Name() string {
	return v.name
}

// CustomValidator 自定义校验器
type CustomValidator struct {
	name       string
	validateFn func(interface{}) (bool, string)
	logger     *zap.Logger
}

// CustomValidatorConfig 自定义校验器配置
type CustomValidatorConfig struct {
	Name       string
	ValidateFn func(interface{}) (bool, string)
	Logger     *zap.Logger
}

// NewCustomValidator 创建自定义校验器
func NewCustomValidator(config CustomValidatorConfig) *CustomValidator {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	return &CustomValidator{
		name:       config.Name,
		validateFn: config.ValidateFn,
		logger:     config.Logger,
	}
}

// Validate 校验数据
func (v *CustomValidator) Validate(value interface{}) ValidationResult {
	result := ValidationResult{
		Valid:     true,
		Errors:    make([]string, 0),
		Warnings:  make([]string, 0),
		Timestamp: time.Now(),
	}
	
	if v.validateFn == nil {
		result.Warnings = append(result.Warnings, "validate function is nil")
		return result
	}
	
	valid, msg := v.validateFn(value)
	if !valid {
		result.Valid = false
		result.Errors = append(result.Errors, msg)
		result.Quality = QualityBad
	}
	
	v.logger.Debug("custom validation",
		zap.Any("value", value),
		zap.Bool("valid", result.Valid),
		zap.String("message", msg),
	)
	
	return result
}

// Name 获取校验器名称
func (v *CustomValidator) Name() string {
	return v.name
}

// ValidatorChain 校验器链
type ValidatorChain struct {
	name       string
	validators []Validator
	stopOnFail bool // 失败时是否停止后续校验
	mu         sync.RWMutex
	logger     *zap.Logger
}

// ValidatorChainConfig 校验器链配置
type ValidatorChainConfig struct {
	Name       string
	StopOnFail bool
	Logger     *zap.Logger
}

// NewValidatorChain 创建校验器链
func NewValidatorChain(config ValidatorChainConfig) *ValidatorChain {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	return &ValidatorChain{
		name:       config.Name,
		validators: make([]Validator, 0),
		stopOnFail: config.StopOnFail,
		logger:     config.Logger,
	}
}

// AddValidator 添加校验器
func (c *ValidatorChain) AddValidator(validator Validator) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.validators = append(c.validators, validator)
}

// RemoveValidator 移除校验器
func (c *ValidatorChain) RemoveValidator(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for i, v := range c.validators {
		if v.Name() == name {
			c.validators = append(c.validators[:i], c.validators[i+1:]...)
			break
		}
	}
}

// Validate 执行校验链
func (c *ValidatorChain) Validate(value interface{}) ValidationResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	result := ValidationResult{
		Valid:     true,
		Errors:    make([]string, 0),
		Warnings:  make([]string, 0),
		Timestamp: time.Now(),
	}
	
	var codes []QualityCode
	
	for _, validator := range c.validators {
		vResult := validator.Validate(value)
		
		// 合并结果
		if !vResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, vResult.Errors...)
			codes = append(codes, vResult.Quality)
			
			if c.stopOnFail {
				break
			}
		}
		
		result.Warnings = append(result.Warnings, vResult.Warnings...)
		if vResult.Quality != QualityGood {
			codes = append(codes, vResult.Quality)
		}
	}
	
	// 组合质量码
	if len(codes) > 0 {
		marker := NewQualityMarker(c.logger)
		result.Quality = marker.Combine(codes...)
	}
	
	c.logger.Debug("validator chain completed",
		zap.Int("validator_count", len(c.validators)),
		zap.Bool("valid", result.Valid),
		zap.Int("error_count", len(result.Errors)),
	)
	
	return result
}

// Name 获取校验器名称
func (c *ValidatorChain) Name() string {
	return c.name
}

// GetValidators 获取所有校验器
func (c *ValidatorChain) GetValidators() []Validator {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	result := make([]Validator, len(c.validators))
	copy(result, c.validators)
	return result
}

// BatchValidator 批量校验器
type BatchValidator struct {
	validator Validator
	workers   int
	logger    *zap.Logger
}

// NewBatchValidator 创建批量校验器
func NewBatchValidator(validator Validator, workers int, logger *zap.Logger) *BatchValidator {
	if logger == nil {
		logger = zap.NewNop()
	}
	if workers <= 0 {
		workers = 4
	}
	return &BatchValidator{
		validator: validator,
		workers:   workers,
		logger:    logger,
	}
}

// ValidateBatch 批量校验
func (v *BatchValidator) ValidateBatch(values []interface{}) []ValidationResult {
	results := make([]ValidationResult, len(values))
	
	var wg sync.WaitGroup
	chunkSize := (len(values) + v.workers - 1) / v.workers
	
	for i := 0; i < v.workers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(values) {
			end = len(values)
		}
		if start >= len(values) {
			break
		}
		
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				results[j] = v.validator.Validate(values[j])
			}
		}(start, end)
	}
	
	wg.Wait()
	
	v.logger.Debug("batch validation completed",
		zap.Int("count", len(values)),
		zap.Int("workers", v.workers),
	)
	
	return results
}

// 辅助函数

// toFloat64 转换为float64
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

// isConvertible 检查类型是否可转换
func isConvertible(from, to reflect.Type) bool {
	if from == to {
		return true
	}
	
	// 数值类型之间可以转换
	switch from.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		switch to.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return true
		}
	}
	
	return from.ConvertibleTo(to)
}

// ValidatorFactory 校验器工厂
type ValidatorFactory struct {
	logger *zap.Logger
}

// NewValidatorFactory 创建校验器工厂
func NewValidatorFactory(logger *zap.Logger) *ValidatorFactory {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ValidatorFactory{logger: logger}
}

// CreateRangeValidator 创建范围校验器
func (f *ValidatorFactory) CreateRangeValidator(name string, min, max float64, inclusive bool) *RangeValidator {
	return NewRangeValidator(RangeValidatorConfig{
		Name:      name,
		MinValue:  min,
		MaxValue:  max,
		Inclusive: inclusive,
		Logger:    f.logger,
	})
}

// CreateNullValidator 创建空值校验器
func (f *ValidatorFactory) CreateNullValidator(name string, allowNull, allowEmptyStr, allowZero bool) *NullValidator {
	return NewNullValidator(NullValidatorConfig{
		Name:          name,
		AllowNull:     allowNull,
		AllowEmptyStr: allowEmptyStr,
		AllowZero:     allowZero,
		Logger:        f.logger,
	})
}

// CreateTypeValidator 创建类型校验器
func (f *ValidatorFactory) CreateTypeValidator(name string, expectType reflect.Type) *TypeValidator {
	return NewTypeValidator(TypeValidatorConfig{
		Name:       name,
		ExpectType: expectType,
		Logger:     f.logger,
	})
}

// CreateCustomValidator 创建自定义校验器
func (f *ValidatorFactory) CreateCustomValidator(name string, validateFn func(interface{}) (bool, string)) *CustomValidator {
	return NewCustomValidator(CustomValidatorConfig{
		Name:       name,
		ValidateFn: validateFn,
		Logger:     f.logger,
	})
}

// CreateValidatorChain 创建校验器链
func (f *ValidatorFactory) CreateValidatorChain(name string, stopOnFail bool) *ValidatorChain {
	return NewValidatorChain(ValidatorChainConfig{
		Name:       name,
		StopOnFail: stopOnFail,
		Logger:     f.logger,
	})
}
