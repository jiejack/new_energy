package processor

import (
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ScaleResult 量程转换结果
type ScaleResult struct {
	Value      float64     `json:"value"`       // 转换后的值
	RawValue   float64     `json:"raw_value"`   // 原始值
	Quality    QualityCode `json:"quality"`     // 质量码
	Timestamp  time.Time   `json:"timestamp"`   // 时间戳
	Scaled     bool        `json:"scaled"`      // 是否进行了转换
	Unit       string      `json:"unit"`        // 单位
	InputUnit  string      `json:"input_unit"`  // 输入单位
	OutputUnit string      `json:"output_unit"` // 输出单位
}

// Scaler 量程转换器接口
type Scaler interface {
	// Scale 执行量程转换
	Scale(value float64) ScaleResult
	// Inverse 反向转换
	Inverse(value float64) ScaleResult
	// Name 获取转换器名称
	Name() string
}

// LinearScaler 线性转换器
// 公式: y = a * x + b
type LinearScaler struct {
	name       string
	slope      float64 // 斜率 a
	intercept  float64 // 截距 b
	inputMin   float64 // 输入最小值
	inputMax   float64 // 输入最大值
	outputMin  float64 // 输出最小值
	outputMax  float64 // 输出最大值
	inputUnit  string  // 输入单位
	outputUnit string  // 输出单位
	logger     *zap.Logger
}

// LinearScalerConfig 线性转换器配置
type LinearScalerConfig struct {
	Name       string
	Slope      float64
	Intercept  float64
	InputMin   float64
	InputMax   float64
	OutputMin  float64
	OutputMax  float64
	InputUnit  string
	OutputUnit string
	Logger     *zap.Logger
}

// NewLinearScaler 创建线性转换器
func NewLinearScaler(config LinearScalerConfig) *LinearScaler {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	
	// 如果没有指定斜率和截距，根据范围计算
	if config.Slope == 0 && config.InputMax != config.InputMin {
		config.Slope = (config.OutputMax - config.OutputMin) / (config.InputMax - config.InputMin)
		config.Intercept = config.OutputMin - config.Slope*config.InputMin
	}
	
	return &LinearScaler{
		name:       config.Name,
		slope:      config.Slope,
		intercept:  config.Intercept,
		inputMin:   config.InputMin,
		inputMax:   config.InputMax,
		outputMin:  config.OutputMin,
		outputMax:  config.OutputMax,
		inputUnit:  config.InputUnit,
		outputUnit: config.OutputUnit,
		logger:     config.Logger,
	}
}

// Scale 执行量程转换
func (s *LinearScaler) Scale(value float64) ScaleResult {
	result := ScaleResult{
		RawValue:  value,
		Timestamp: time.Now(),
		InputUnit: s.inputUnit,
		OutputUnit: s.outputUnit,
	}
	
	// 检查输入范围
	if s.inputMin != s.inputMax {
		if value < s.inputMin || value > s.inputMax {
			result.Quality = QualityUncertain | QualityReasonOutOfRange
		}
	}
	
	// 线性转换
	result.Value = s.slope*value + s.intercept
	result.Scaled = true
	result.Quality = QualityGood | QualityReasonScaled
	
	s.logger.Debug("linear scale",
		zap.Float64("raw", value),
		zap.Float64("scaled", result.Value),
		zap.Float64("slope", s.slope),
		zap.Float64("intercept", s.intercept),
	)
	
	return result
}

// Inverse 反向转换
func (s *LinearScaler) Inverse(value float64) ScaleResult {
	result := ScaleResult{
		RawValue:  value,
		Timestamp: time.Now(),
		InputUnit: s.outputUnit,
		OutputUnit: s.inputUnit,
	}
	
	// 反向转换: x = (y - b) / a
	if s.slope == 0 {
		result.Quality = QualityBad | QualityReasonConfiguration
		result.Value = 0
		return result
	}
	
	result.Value = (value - s.intercept) / s.slope
	result.Scaled = true
	result.Quality = QualityGood | QualityReasonScaled
	
	return result
}

// Name 获取转换器名称
func (s *LinearScaler) Name() string {
	return s.name
}

// PolynomialScaler 多项式转换器
// 公式: y = a0 + a1*x + a2*x^2 + ... + an*x^n
type PolynomialScaler struct {
	name       string
	coeffs     []float64 // 系数 [a0, a1, a2, ...]
	inputUnit  string
	outputUnit string
	logger     *zap.Logger
}

// PolynomialScalerConfig 多项式转换器配置
type PolynomialScalerConfig struct {
	Name       string
	Coeffs     []float64
	InputUnit  string
	OutputUnit string
	Logger     *zap.Logger
}

// NewPolynomialScaler 创建多项式转换器
func NewPolynomialScaler(config PolynomialScalerConfig) *PolynomialScaler {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	return &PolynomialScaler{
		name:       config.Name,
		coeffs:     config.Coeffs,
		inputUnit:  config.InputUnit,
		outputUnit: config.OutputUnit,
		logger:     config.Logger,
	}
}

// Scale 执行量程转换
func (s *PolynomialScaler) Scale(value float64) ScaleResult {
	result := ScaleResult{
		RawValue:  value,
		Timestamp: time.Now(),
		InputUnit: s.inputUnit,
		OutputUnit: s.outputUnit,
	}
	
	// 多项式计算
	result.Value = 0
	power := 1.0
	for _, coeff := range s.coeffs {
		result.Value += coeff * power
		power *= value
	}
	
	result.Scaled = true
	result.Quality = QualityGood | QualityReasonScaled
	
	s.logger.Debug("polynomial scale",
		zap.Float64("raw", value),
		zap.Float64("scaled", result.Value),
	)
	
	return result
}

// Inverse 反向转换（数值方法求解）
func (s *PolynomialScaler) Inverse(value float64) ScaleResult {
	result := ScaleResult{
		RawValue:  value,
		Timestamp: time.Now(),
		InputUnit: s.outputUnit,
		OutputUnit: s.inputUnit,
	}
	
	// 使用牛顿迭代法求解
	// f(x) = a0 + a1*x + a2*x^2 + ... - value = 0
	x := value // 初始猜测
	
	for i := 0; i < 100; i++ {
		// 计算f(x)和f'(x)
		f := -value
		df := 0.0
		power := 1.0
		
		for j, coeff := range s.coeffs {
			f += coeff * power
			if j > 0 {
				df += float64(j) * coeff * power / value
			}
			power *= value
		}
		
		if math.Abs(df) < 1e-10 {
			break
		}
		
		xNew := x - f/df
		if math.Abs(xNew-x) < 1e-6 {
			x = xNew
			break
		}
		x = xNew
	}
	
	result.Value = x
	result.Scaled = true
	result.Quality = QualityGood | QualityReasonScaled
	
	return result
}

// Name 获取转换器名称
func (s *PolynomialScaler) Name() string {
	return s.name
}

// LookupTableScaler 查表转换器
type LookupTableScaler struct {
	name        string
	table       []LookupEntry
	inputUnit   string
	outputUnit  string
	extrapolate bool // 是否外推
	logger      *zap.Logger
	mu          sync.RWMutex
}

// LookupEntry 查表条目
type LookupEntry struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
}

// LookupTableScalerConfig 查表转换器配置
type LookupTableScalerConfig struct {
	Name        string
	Table       []LookupEntry
	InputUnit   string
	OutputUnit  string
	Extrapolate bool
	Logger      *zap.Logger
}

// NewLookupTableScaler 创建查表转换器
func NewLookupTableScaler(config LookupTableScalerConfig) *LookupTableScaler {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	
	// 确保表按输入值排序
	table := make([]LookupEntry, len(config.Table))
	copy(table, config.Table)
	
	for i := 0; i < len(table)-1; i++ {
		for j := i + 1; j < len(table); j++ {
			if table[i].Input > table[j].Input {
				table[i], table[j] = table[j], table[i]
			}
		}
	}
	
	return &LookupTableScaler{
		name:        config.Name,
		table:       table,
		inputUnit:   config.InputUnit,
		outputUnit:  config.OutputUnit,
		extrapolate: config.Extrapolate,
		logger:      config.Logger,
	}
}

// Scale 执行量程转换
func (s *LookupTableScaler) Scale(value float64) ScaleResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := ScaleResult{
		RawValue:  value,
		Timestamp: time.Now(),
		InputUnit: s.inputUnit,
		OutputUnit: s.outputUnit,
	}
	
	if len(s.table) == 0 {
		result.Quality = QualityBad | QualityReasonConfiguration
		result.Value = value
		return result
	}
	
	// 查找对应的区间
	idx := s.findInterval(value)
	
	if idx < 0 {
		// 小于最小值
		if s.extrapolate {
			// 外推
			result.Value = s.extrapolateLeft(value)
			result.Quality = QualityQuestionable | QualityReasonScaled
		} else {
			result.Value = s.table[0].Output
			result.Quality = QualityUncertain | QualityReasonOutOfRange
		}
	} else if idx >= len(s.table)-1 {
		// 大于最大值
		if s.extrapolate {
			result.Value = s.extrapolateRight(value)
			result.Quality = QualityQuestionable | QualityReasonScaled
		} else {
			result.Value = s.table[len(s.table)-1].Output
			result.Quality = QualityUncertain | QualityReasonOutOfRange
		}
	} else {
		// 线性插值
		result.Value = s.interpolate(idx, value)
		result.Quality = QualityGood | QualityReasonScaled
	}
	
	result.Scaled = true
	
	s.logger.Debug("lookup table scale",
		zap.Float64("raw", value),
		zap.Float64("scaled", result.Value),
		zap.Int("index", idx),
	)
	
	return result
}

// findInterval 查找值所在的区间
func (s *LookupTableScaler) findInterval(value float64) int {
	for i := 0; i < len(s.table)-1; i++ {
		if value >= s.table[i].Input && value < s.table[i+1].Input {
			return i
		}
	}
	if value < s.table[0].Input {
		return -1
	}
	return len(s.table) - 1
}

// interpolate 线性插值
func (s *LookupTableScaler) interpolate(idx int, value float64) float64 {
	x0 := s.table[idx].Input
	y0 := s.table[idx].Output
	x1 := s.table[idx+1].Input
	y1 := s.table[idx+1].Output
	
	// 线性插值
	return y0 + (y1-y0)*(value-x0)/(x1-x0)
}

// extrapolateLeft 左侧外推
func (s *LookupTableScaler) extrapolateLeft(value float64) float64 {
	if len(s.table) < 2 {
		return s.table[0].Output
	}
	
	x0 := s.table[0].Input
	y0 := s.table[0].Output
	x1 := s.table[1].Input
	y1 := s.table[1].Output
	
	return y0 + (y1-y0)*(value-x0)/(x1-x0)
}

// extrapolateRight 右侧外推
func (s *LookupTableScaler) extrapolateRight(value float64) float64 {
	n := len(s.table)
	if n < 2 {
		return s.table[n-1].Output
	}
	
	x0 := s.table[n-2].Input
	y0 := s.table[n-2].Output
	x1 := s.table[n-1].Input
	y1 := s.table[n-1].Output
	
	return y0 + (y1-y0)*(value-x0)/(x1-x0)
}

// Inverse 反向转换
func (s *LookupTableScaler) Inverse(value float64) ScaleResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := ScaleResult{
		RawValue:  value,
		Timestamp: time.Now(),
		InputUnit: s.outputUnit,
		OutputUnit: s.inputUnit,
	}
	
	// 查找输出值对应的输入值
	for i := 0; i < len(s.table)-1; i++ {
		y0 := s.table[i].Output
		y1 := s.table[i+1].Output
		
		if (value >= y0 && value <= y1) || (value >= y1 && value <= y0) {
			// 线性插值
			x0 := s.table[i].Input
			x1 := s.table[i+1].Input
			
			result.Value = x0 + (x1-x0)*(value-y0)/(y1-y0)
			result.Scaled = true
			result.Quality = QualityGood | QualityReasonScaled
			return result
		}
	}
	
	result.Quality = QualityUncertain | QualityReasonOutOfRange
	result.Value = value
	return result
}

// Name 获取转换器名称
func (s *LookupTableScaler) Name() string {
	return s.name
}

// AddEntry 添加查表条目
func (s *LookupTableScaler) AddEntry(entry LookupEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.table = append(s.table, entry)
	
	// 重新排序
	for i := len(s.table) - 1; i > 0; i-- {
		if s.table[i].Input < s.table[i-1].Input {
			s.table[i], s.table[i-1] = s.table[i-1], s.table[i]
		} else {
			break
		}
	}
}

// EngineeringUnitScaler 工程量转换器
// 用于常见的工程单位转换
type EngineeringUnitScaler struct {
	name       string
	fromUnit   string
	toUnit     string
	factor     float64 // 转换因子
	offset     float64 // 偏移量
	logger     *zap.Logger
}

// EngineeringUnitScalerConfig 工程量转换器配置
type EngineeringUnitScalerConfig struct {
	Name     string
	FromUnit string
	ToUnit   string
	Factor   float64
	Offset   float64
	Logger   *zap.Logger
}

// NewEngineeringUnitScaler 创建工程量转换器
func NewEngineeringUnitScaler(config EngineeringUnitScalerConfig) *EngineeringUnitScaler {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	return &EngineeringUnitScaler{
		name:     config.Name,
		fromUnit: config.FromUnit,
		toUnit:   config.ToUnit,
		factor:   config.Factor,
		offset:   config.Offset,
		logger:   config.Logger,
	}
}

// Scale 执行量程转换
func (s *EngineeringUnitScaler) Scale(value float64) ScaleResult {
	result := ScaleResult{
		RawValue:  value,
		Timestamp: time.Now(),
		InputUnit: s.fromUnit,
		OutputUnit: s.toUnit,
	}
	
	result.Value = value*s.factor + s.offset
	result.Scaled = true
	result.Quality = QualityGood | QualityReasonScaled
	
	s.logger.Debug("engineering unit scale",
		zap.Float64("raw", value),
		zap.Float64("scaled", result.Value),
		zap.String("from", s.fromUnit),
		zap.String("to", s.toUnit),
	)
	
	return result
}

// Inverse 反向转换
func (s *EngineeringUnitScaler) Inverse(value float64) ScaleResult {
	result := ScaleResult{
		RawValue:  value,
		Timestamp: time.Now(),
		InputUnit: s.toUnit,
		OutputUnit: s.fromUnit,
	}
	
	if s.factor == 0 {
		result.Quality = QualityBad | QualityReasonConfiguration
		result.Value = 0
		return result
	}
	
	result.Value = (value - s.offset) / s.factor
	result.Scaled = true
	result.Quality = QualityGood | QualityReasonScaled
	
	return result
}

// Name 获取转换器名称
func (s *EngineeringUnitScaler) Name() string {
	return s.name
}

// ScaleChain 转换器链
type ScaleChain struct {
	name    string
	scalers []Scaler
	mu      sync.RWMutex
	logger  *zap.Logger
}

// ScaleChainConfig 转换器链配置
type ScaleChainConfig struct {
	Name   string
	Logger *zap.Logger
}

// NewScaleChain 创建转换器链
func NewScaleChain(config ScaleChainConfig) *ScaleChain {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	return &ScaleChain{
		name:    config.Name,
		scalers: make([]Scaler, 0),
		logger:  config.Logger,
	}
}

// AddScaler 添加转换器
func (c *ScaleChain) AddScaler(scaler Scaler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.scalers = append(c.scalers, scaler)
}

// RemoveScaler 移除转换器
func (c *ScaleChain) RemoveScaler(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for i, s := range c.scalers {
		if s.Name() == name {
			c.scalers = append(c.scalers[:i], c.scalers[i+1:]...)
			break
		}
	}
}

// Scale 执行转换链
func (c *ScaleChain) Scale(value float64) ScaleResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	result := ScaleResult{
		Value:     value,
		RawValue:  value,
		Quality:   QualityGood,
		Timestamp: time.Now(),
	}
	
	for _, scaler := range c.scalers {
		result = scaler.Scale(result.Value)
	}
	
	c.logger.Debug("scale chain completed",
		zap.Int("scaler_count", len(c.scalers)),
		zap.Float64("raw", value),
		zap.Float64("scaled", result.Value),
	)
	
	return result
}

// Inverse 反向转换链
func (c *ScaleChain) Inverse(value float64) ScaleResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	result := ScaleResult{
		Value:     value,
		RawValue:  value,
		Quality:   QualityGood,
		Timestamp: time.Now(),
	}
	
	// 反向执行转换器链
	for i := len(c.scalers) - 1; i >= 0; i-- {
		result = c.scalers[i].Inverse(result.Value)
	}
	
	return result
}

// Name 获取转换器名称
func (c *ScaleChain) Name() string {
	return c.name
}

// GetScalers 获取所有转换器
func (c *ScaleChain) GetScalers() []Scaler {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	result := make([]Scaler, len(c.scalers))
	copy(result, c.scalers)
	return result
}

// BatchScaler 批量转换器
type BatchScaler struct {
	scaler  Scaler
	workers int
	logger  *zap.Logger
}

// NewBatchScaler 创建批量转换器
func NewBatchScaler(scaler Scaler, workers int, logger *zap.Logger) *BatchScaler {
	if logger == nil {
		logger = zap.NewNop()
	}
	if workers <= 0 {
		workers = 4
	}
	return &BatchScaler{
		scaler:  scaler,
		workers: workers,
		logger:  logger,
	}
}

// ScaleBatch 批量转换
func (s *BatchScaler) ScaleBatch(values []float64) []ScaleResult {
	results := make([]ScaleResult, len(values))
	
	var wg sync.WaitGroup
	chunkSize := (len(values) + s.workers - 1) / s.workers
	
	for i := 0; i < s.workers; i++ {
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
				results[j] = s.scaler.Scale(values[j])
			}
		}(start, end)
	}
	
	wg.Wait()
	
	s.logger.Debug("batch scale completed",
		zap.Int("count", len(values)),
		zap.Int("workers", s.workers),
	)
	
	return results
}

// ScalerFactory 转换器工厂
type ScalerFactory struct {
	logger *zap.Logger
}

// NewScalerFactory 创建转换器工厂
func NewScalerFactory(logger *zap.Logger) *ScalerFactory {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ScalerFactory{logger: logger}
}

// CreateLinearScaler 创建线性转换器
func (f *ScalerFactory) CreateLinearScaler(name string, slope, intercept float64) *LinearScaler {
	return NewLinearScaler(LinearScalerConfig{
		Name:      name,
		Slope:     slope,
		Intercept: intercept,
		Logger:    f.logger,
	})
}

// CreateLinearScalerFromRange 根据范围创建线性转换器
func (f *ScalerFactory) CreateLinearScalerFromRange(name string, inputMin, inputMax, outputMin, outputMax float64) *LinearScaler {
	return NewLinearScaler(LinearScalerConfig{
		Name:      name,
		InputMin:  inputMin,
		InputMax:  inputMax,
		OutputMin: outputMin,
		OutputMax: outputMax,
		Logger:    f.logger,
	})
}

// CreatePolynomialScaler 创建多项式转换器
func (f *ScalerFactory) CreatePolynomialScaler(name string, coeffs []float64) *PolynomialScaler {
	return NewPolynomialScaler(PolynomialScalerConfig{
		Name:   name,
		Coeffs: coeffs,
		Logger: f.logger,
	})
}

// CreateLookupTableScaler 创建查表转换器
func (f *ScalerFactory) CreateLookupTableScaler(name string, table []LookupEntry, extrapolate bool) *LookupTableScaler {
	return NewLookupTableScaler(LookupTableScalerConfig{
		Name:        name,
		Table:       table,
		Extrapolate: extrapolate,
		Logger:      f.logger,
	})
}

// CreateEngineeringUnitScaler 创建工程量转换器
func (f *ScalerFactory) CreateEngineeringUnitScaler(name, fromUnit, toUnit string, factor, offset float64) *EngineeringUnitScaler {
	return NewEngineeringUnitScaler(EngineeringUnitScalerConfig{
		Name:     name,
		FromUnit: fromUnit,
		ToUnit:   toUnit,
		Factor:   factor,
		Offset:   offset,
		Logger:   f.logger,
	})
}

// CreateScaleChain 创建转换器链
func (f *ScalerFactory) CreateScaleChain(name string) *ScaleChain {
	return NewScaleChain(ScaleChainConfig{
		Name:   name,
		Logger: f.logger,
	})
}

// 预定义的工程量转换器

// CelsiusToFahrenheit 摄氏度转华氏度
func CelsiusToFahrenheit(logger *zap.Logger) *EngineeringUnitScaler {
	return NewEngineeringUnitScaler(EngineeringUnitScalerConfig{
		Name:     "CelsiusToFahrenheit",
		FromUnit: "°C",
		ToUnit:   "°F",
		Factor:   1.8,
		Offset:   32,
		Logger:   logger,
	})
}

// FahrenheitToCelsius 华氏度转摄氏度
func FahrenheitToCelsius(logger *zap.Logger) *EngineeringUnitScaler {
	return NewEngineeringUnitScaler(EngineeringUnitScalerConfig{
		Name:     "FahrenheitToCelsius",
		FromUnit: "°F",
		ToUnit:   "°C",
		Factor:   5.0 / 9.0,
		Offset:   -32 * 5.0 / 9.0,
		Logger:   logger,
	})
}

// BarToPsi 巴转Psi
func BarToPsi(logger *zap.Logger) *EngineeringUnitScaler {
	return NewEngineeringUnitScaler(EngineeringUnitScalerConfig{
		Name:     "BarToPsi",
		FromUnit: "bar",
		ToUnit:   "psi",
		Factor:   14.5038,
		Offset:   0,
		Logger:   logger,
	})
}

// PsiToBar Psi转巴
func PsiToBar(logger *zap.Logger) *EngineeringUnitScaler {
	return NewEngineeringUnitScaler(EngineeringUnitScalerConfig{
		Name:     "PsiToBar",
		FromUnit: "psi",
		ToUnit:   "bar",
		Factor:   1.0 / 14.5038,
		Offset:   0,
		Logger:   logger,
	})
}

// KWToMW 千瓦转兆瓦
func KWToMW(logger *zap.Logger) *EngineeringUnitScaler {
	return NewEngineeringUnitScaler(EngineeringUnitScalerConfig{
		Name:     "KWToMW",
		FromUnit: "kW",
		ToUnit:   "MW",
		Factor:   0.001,
		Offset:   0,
		Logger:   logger,
	})
}

// MWToKW 兆瓦转千瓦
func MWToKW(logger *zap.Logger) *EngineeringUnitScaler {
	return NewEngineeringUnitScaler(EngineeringUnitScalerConfig{
		Name:     "MWToKW",
		FromUnit: "MW",
		ToUnit:   "kW",
		Factor:   1000,
		Offset:   0,
		Logger:   logger,
	})
}

// PercentToDecimal 百分比转小数
func PercentToDecimal(logger *zap.Logger) *EngineeringUnitScaler {
	return NewEngineeringUnitScaler(EngineeringUnitScalerConfig{
		Name:     "PercentToDecimal",
		FromUnit: "%",
		ToUnit:   "",
		Factor:   0.01,
		Offset:   0,
		Logger:   logger,
	})
}

// DecimalToPercent 小数转百分比
func DecimalToPercent(logger *zap.Logger) *EngineeringUnitScaler {
	return NewEngineeringUnitScaler(EngineeringUnitScalerConfig{
		Name:     "DecimalToPercent",
		FromUnit: "",
		ToUnit:   "%",
		Factor:   100,
		Offset:   0,
		Logger:   logger,
	})
}

// ScaleResult 扩展字段
type ScaleResultExt struct {
	ScaleResult
	InputUnit  string `json:"input_unit"`  // 输入单位
	OutputUnit string `json:"output_unit"` // 输出单位
}

// ValidateScaleResult 验证转换结果
func ValidateScaleResult(result ScaleResult, min, max float64) error {
	if result.Value < min || result.Value > max {
		return fmt.Errorf("scaled value %.2f is out of range [%.2f, %.2f]", result.Value, min, max)
	}
	return nil
}
