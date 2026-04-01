package formula

import (
	"fmt"
	"math"
	"sort"
)

// FunctionRegistry 函数注册表
type FunctionRegistry struct {
	functions map[string]Function
}

// NewFunctionRegistry 创建函数注册表
func NewFunctionRegistry() *FunctionRegistry {
	registry := &FunctionRegistry{
		functions: make(map[string]Function),
	}

	// 注册默认函数
	registry.registerDefaultFunctions()

	return registry
}

// Register 注册函数
func (r *FunctionRegistry) Register(name string, fn Function) {
	r.functions[name] = fn
}

// Get 获取函数
func (r *FunctionRegistry) Get(name string) (Function, bool) {
	fn, exists := r.functions[name]
	return fn, exists
}

// GetAll 获取所有函数
func (r *FunctionRegistry) GetAll() map[string]Function {
	result := make(map[string]Function)
	for k, v := range r.functions {
		result[k] = v
	}
	return result
}

// registerDefaultFunctions 注册默认函数
func (r *FunctionRegistry) registerDefaultFunctions() {
	// ==================== 基础运算函数 ====================
	r.Register("abs", funcAbs)
	r.Register("sign", funcSign)
	r.Register("floor", funcFloor)
	r.Register("ceil", funcCeil)
	r.Register("round", funcRound)
	r.Register("trunc", funcTrunc)

	// ==================== 幂运算函数 ====================
	r.Register("pow", funcPow)
	r.Register("sqrt", funcSqrt)
	r.Register("cbrt", funcCbrt)
	r.Register("exp", funcExp)
	r.Register("exp2", funcExp2)
	r.Register("exp10", funcExp10)

	// ==================== 对数函数 ====================
	r.Register("log", funcLog)
	r.Register("log10", funcLog10)
	r.Register("log2", funcLog2)
	r.Register("log1p", funcLog1p)

	// ==================== 三角函数 ====================
	r.Register("sin", funcSin)
	r.Register("cos", funcCos)
	r.Register("tan", funcTan)
	r.Register("asin", funcAsin)
	r.Register("acos", funcAcos)
	r.Register("atan", funcAtan)
	r.Register("atan2", funcAtan2)
	r.Register("sinh", funcSinh)
	r.Register("cosh", funcCosh)
	r.Register("tanh", funcTanh)
	r.Register("asinh", funcAsinh)
	r.Register("acosh", funcAcosh)
	r.Register("atanh", funcAtanh)

	// ==================== 角度转换函数 ====================
	r.Register("rad", funcRad)     // 角度转弧度
	r.Register("deg", funcDeg)     // 弧度转角度
	r.Register("radians", funcRad) // 别名
	r.Register("degrees", funcDeg) // 别名

	// ==================== 统计函数 ====================
	r.Register("avg", funcAvg)
	r.Register("sum", funcSum)
	r.Register("max", funcMax)
	r.Register("min", funcMin)
	r.Register("count", funcCount)
	r.Register("mean", funcAvg) // 别名
	r.Register("product", funcProduct)
	r.Register("variance", funcVariance)
	r.Register("stddev", funcStdDev)
	r.Register("median", funcMedian)

	// ==================== 条件函数 ====================
	r.Register("if", funcIf)
	r.Register("switch", funcSwitch)
	r.Register("coalesce", funcCoalesce)
	r.Register("iif", funcIf) // 别名

	// ==================== 类型转换函数 ====================
	r.Register("int", funcInt)
	r.Register("float", funcFloat)
	r.Register("string", funcString)
	r.Register("bool", funcBool)

	// ==================== 字符串函数 ====================
	r.Register("len", funcLen)
	r.Register("upper", funcUpper)
	r.Register("lower", funcLower)
	r.Register("trim", funcTrim)
	r.Register("substr", funcSubstr)
	r.Register("concat", funcConcat)
	r.Register("contains", funcContains)
	r.Register("startsWith", funcStartsWith)
	r.Register("endsWith", funcEndsWith)
	r.Register("replace", funcReplace)
	r.Register("split", funcSplit)
	r.Register("join", funcJoin)

	// ==================== 数组函数 ====================
	r.Register("array", funcArray)
	r.Register("first", funcFirst)
	r.Register("last", funcLast)
	r.Register("nth", funcNth)
	r.Register("slice", funcSlice)
	r.Register("push", funcPush)
	r.Register("pop", funcPop)
	r.Register("reverse", funcReverse)
	r.Register("sort", funcSort)
	r.Register("filter", funcFilter)
	r.Register("map", funcMap)
	r.Register("reduce", funcReduce)

	// ==================== 数学常量 ====================
	r.Register("pi", funcPi)
	r.Register("e", funcE)
	r.Register("phi", funcPhi) // 黄金比例

	// ==================== 其他实用函数 ====================
	r.Register("clamp", funcClamp)
	r.Register("lerp", funcLerp)
	r.Register("step", funcStep)
	r.Register("smoothstep", funcSmoothstep)
	r.Register("mod", funcMod)
	r.Register("gcd", funcGcd)
	r.Register("lcm", funcLcm)
	r.Register("isFinite", funcIsFinite)
	r.Register("isInf", funcIsInf)
	r.Register("isNaN", funcIsNaN)
	r.Register("random", funcRandom)
}

// ==================== 基础运算函数实现 ====================

func funcAbs(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("abs", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Abs(x), nil
}

func funcSign(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("sign", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	if x > 0 {
		return 1.0, nil
	} else if x < 0 {
		return -1.0, nil
	}
	return 0.0, nil
}

func funcFloor(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("floor", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Floor(x), nil
}

func funcCeil(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("ceil", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Ceil(x), nil
}

func funcRound(args ...interface{}) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("round requires 1 or 2 arguments, got %d", len(args))
	}

	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}

	if len(args) == 1 {
		return math.Round(x), nil
	}

	// 带精度的四舍五入
	precision, err := toInt64(args[1])
	if err != nil {
		return nil, err
	}

	multiplier := math.Pow(10, float64(precision))
	return math.Round(x*multiplier) / multiplier, nil
}

func funcTrunc(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("trunc", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Trunc(x), nil
}

// ==================== 幂运算函数实现 ====================

func funcPow(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("pow", args, 2); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	y, err := toFloat64(args[1])
	if err != nil {
		return nil, err
	}
	return math.Pow(x, y), nil
}

func funcSqrt(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("sqrt", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	if x < 0 {
		return nil, fmt.Errorf("sqrt: cannot calculate square root of negative number")
	}
	return math.Sqrt(x), nil
}

func funcCbrt(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("cbrt", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Cbrt(x), nil
}

func funcExp(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("exp", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Exp(x), nil
}

func funcExp2(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("exp2", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Exp2(x), nil
}

func funcExp10(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("exp10", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Pow(10, x), nil
}

// ==================== 对数函数实现 ====================

func funcLog(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("log", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	if x <= 0 {
		return nil, fmt.Errorf("log: cannot calculate logarithm of non-positive number")
	}
	return math.Log(x), nil
}

func funcLog10(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("log10", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	if x <= 0 {
		return nil, fmt.Errorf("log10: cannot calculate logarithm of non-positive number")
	}
	return math.Log10(x), nil
}

func funcLog2(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("log2", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	if x <= 0 {
		return nil, fmt.Errorf("log2: cannot calculate logarithm of non-positive number")
	}
	return math.Log2(x), nil
}

func funcLog1p(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("log1p", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	if x <= -1 {
		return nil, fmt.Errorf("log1p: argument must be greater than -1")
	}
	return math.Log1p(x), nil
}

// ==================== 三角函数实现 ====================

func funcSin(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("sin", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Sin(x), nil
}

func funcCos(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("cos", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Cos(x), nil
}

func funcTan(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("tan", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Tan(x), nil
}

func funcAsin(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("asin", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	if x < -1 || x > 1 {
		return nil, fmt.Errorf("asin: argument must be between -1 and 1")
	}
	return math.Asin(x), nil
}

func funcAcos(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("acos", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	if x < -1 || x > 1 {
		return nil, fmt.Errorf("acos: argument must be between -1 and 1")
	}
	return math.Acos(x), nil
}

func funcAtan(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("atan", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Atan(x), nil
}

func funcAtan2(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("atan2", args, 2); err != nil {
		return nil, err
	}
	y, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	x, err := toFloat64(args[1])
	if err != nil {
		return nil, err
	}
	return math.Atan2(y, x), nil
}

func funcSinh(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("sinh", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Sinh(x), nil
}

func funcCosh(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("cosh", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Cosh(x), nil
}

func funcTanh(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("tanh", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Tanh(x), nil
}

func funcAsinh(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("asinh", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return math.Asinh(x), nil
}

func funcAcosh(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("acosh", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	if x < 1 {
		return nil, fmt.Errorf("acosh: argument must be >= 1")
	}
	return math.Acosh(x), nil
}

func funcAtanh(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("atanh", args, 1); err != nil {
		return nil, err
	}
	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	if x <= -1 || x >= 1 {
		return nil, fmt.Errorf("atanh: argument must be between -1 and 1 (exclusive)")
	}
	return math.Atanh(x), nil
}

// ==================== 角度转换函数实现 ====================

func funcRad(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("rad", args, 1); err != nil {
		return nil, err
	}
	degrees, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return degrees * math.Pi / 180, nil
}

func funcDeg(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("deg", args, 1); err != nil {
		return nil, err
	}
	radians, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}
	return radians * 180 / math.Pi, nil
}

// ==================== 统计函数实现 ====================

func funcAvg(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("avg requires at least 1 argument")
	}

	numbers, err := extractNumbers(args)
	if err != nil {
		return nil, err
	}

	if len(numbers) == 0 {
		return 0.0, nil
	}

	sum := 0.0
	for _, n := range numbers {
		sum += n
	}
	return sum / float64(len(numbers)), nil
}

func funcSum(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("sum requires at least 1 argument")
	}

	numbers, err := extractNumbers(args)
	if err != nil {
		return nil, err
	}

	sum := 0.0
	for _, n := range numbers {
		sum += n
	}
	return sum, nil
}

func funcMax(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("max requires at least 1 argument")
	}

	numbers, err := extractNumbers(args)
	if err != nil {
		return nil, err
	}

	if len(numbers) == 0 {
		return nil, fmt.Errorf("max: no valid numbers")
	}

	max := numbers[0]
	for _, n := range numbers[1:] {
		if n > max {
			max = n
		}
	}
	return max, nil
}

func funcMin(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("min requires at least 1 argument")
	}

	numbers, err := extractNumbers(args)
	if err != nil {
		return nil, err
	}

	if len(numbers) == 0 {
		return nil, fmt.Errorf("min: no valid numbers")
	}

	min := numbers[0]
	for _, n := range numbers[1:] {
		if n < min {
			min = n
		}
	}
	return min, nil
}

func funcCount(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return 0.0, nil
	}

	// 如果第一个参数是数组，返回数组长度
	if arr, ok := args[0].([]interface{}); ok {
		return float64(len(arr)), nil
	}

	return float64(len(args)), nil
}

func funcProduct(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("product requires at least 1 argument")
	}

	numbers, err := extractNumbers(args)
	if err != nil {
		return nil, err
	}

	product := 1.0
	for _, n := range numbers {
		product *= n
	}
	return product, nil
}

func funcVariance(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("variance requires at least 1 argument")
	}

	numbers, err := extractNumbers(args)
	if err != nil {
		return nil, err
	}

	if len(numbers) < 2 {
		return 0.0, nil
	}

	// 计算平均值
	sum := 0.0
	for _, n := range numbers {
		sum += n
	}
	mean := sum / float64(len(numbers))

	// 计算方差
	variance := 0.0
	for _, n := range numbers {
		diff := n - mean
		variance += diff * diff
	}
	variance /= float64(len(numbers))

	return variance, nil
}

func funcStdDev(args ...interface{}) (interface{}, error) {
	variance, err := funcVariance(args...)
	if err != nil {
		return nil, err
	}
	return math.Sqrt(variance.(float64)), nil
}

func funcMedian(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("median requires at least 1 argument")
	}

	numbers, err := extractNumbers(args)
	if err != nil {
		return nil, err
	}

	if len(numbers) == 0 {
		return nil, fmt.Errorf("median: no valid numbers")
	}

	sort.Float64s(numbers)

	n := len(numbers)
	if n%2 == 1 {
		return numbers[n/2], nil
	}
	return (numbers[n/2-1] + numbers[n/2]) / 2, nil
}

// ==================== 条件函数实现 ====================

func funcIf(args ...interface{}) (interface{}, error) {
	if len(args) < 3 || len(args) > 4 {
		return nil, fmt.Errorf("if requires 3 or 4 arguments, got %d", len(args))
	}

	condition, err := toBool(args[0])
	if err != nil {
		return nil, fmt.Errorf("if condition: %w", err)
	}

	if condition {
		return args[1], nil
	}
	return args[2], nil
}

func funcSwitch(args ...interface{}) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("switch requires at least 3 arguments")
	}

	value := args[0]

	// 检查 case-value pairs
	i := 1
	for i < len(args)-1 {
		caseValue := args[i]
		if value == caseValue {
			return args[i+1], nil
		}
		i += 2
	}

	// 检查是否有默认值
	if i < len(args) {
		return args[i], nil
	}

	return nil, fmt.Errorf("switch: no matching case found")
}

func funcCoalesce(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("coalesce requires at least 1 argument")
	}

	for _, arg := range args {
		if arg != nil {
			switch v := arg.(type) {
			case float64:
				if !math.IsNaN(v) {
					return v, nil
				}
			case string:
				if v != "" {
					return v, nil
				}
			default:
				return v, nil
			}
		}
	}

	return nil, nil
}

// ==================== 类型转换函数实现 ====================

func funcInt(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("int", args, 1); err != nil {
		return nil, err
	}
	val, err := toInt64(args[0])
	if err != nil {
		return nil, err
	}
	return float64(val), nil
}

func funcFloat(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("float", args, 1); err != nil {
		return nil, err
	}
	return toFloat64(args[0])
}

func funcString(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("string", args, 1); err != nil {
		return nil, err
	}
	return fmt.Sprintf("%v", args[0]), nil
}

func funcBool(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("bool", args, 1); err != nil {
		return nil, err
	}
	return toBool(args[0])
}

// ==================== 字符串函数实现 ====================

func funcLen(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("len", args, 1); err != nil {
		return nil, err
	}

	switch v := args[0].(type) {
	case string:
		return float64(len(v)), nil
	case []interface{}:
		return float64(len(v)), nil
	default:
		return nil, fmt.Errorf("len: unsupported type %T", args[0])
	}
}

func funcUpper(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("upper", args, 1); err != nil {
		return nil, err
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("upper: argument must be string")
	}
	return stringsToUpper(s), nil
}

func stringsToUpper(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			result[i] = c - 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func funcLower(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("lower", args, 1); err != nil {
		return nil, err
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("lower: argument must be string")
	}
	return stringsToLower(s), nil
}

func stringsToLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func funcTrim(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("trim", args, 1); err != nil {
		return nil, err
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("trim: argument must be string")
	}
	return stringsTrim(s), nil
}

func stringsTrim(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

func funcSubstr(args ...interface{}) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("substr requires 2 or 3 arguments")
	}

	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("substr: first argument must be string")
	}

	start, err := toInt64(args[1])
	if err != nil {
		return nil, err
	}

	if start < 0 {
		start = int64(len(s)) + start
	}
	if start < 0 {
		start = 0
	}
	if start > int64(len(s)) {
		return "", nil
	}

	if len(args) == 2 {
		return s[start:], nil
	}

	length, err := toInt64(args[2])
	if err != nil {
		return nil, err
	}

	end := start + length
	if end > int64(len(s)) {
		end = int64(len(s))
	}

	return s[start:end], nil
}

func funcConcat(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}

	result := ""
	for _, arg := range args {
		result += fmt.Sprintf("%v", arg)
	}
	return result, nil
}

func funcContains(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("contains", args, 2); err != nil {
		return nil, err
	}

	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("contains: first argument must be string")
	}

	substr, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("contains: second argument must be string")
	}

	return stringsContains(s, substr), nil
}

func stringsContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func funcStartsWith(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("startsWith", args, 2); err != nil {
		return nil, err
	}

	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("startsWith: first argument must be string")
	}

	prefix, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("startsWith: second argument must be string")
	}

	if len(prefix) > len(s) {
		return false, nil
	}

	for i := 0; i < len(prefix); i++ {
		if s[i] != prefix[i] {
			return false, nil
		}
	}
	return true, nil
}

func funcEndsWith(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("endsWith", args, 2); err != nil {
		return nil, err
	}

	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("endsWith: first argument must be string")
	}

	suffix, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("endsWith: second argument must be string")
	}

	if len(suffix) > len(s) {
		return false, nil
	}

	offset := len(s) - len(suffix)
	for i := 0; i < len(suffix); i++ {
		if s[offset+i] != suffix[i] {
			return false, nil
		}
	}
	return true, nil
}

func funcReplace(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("replace", args, 3); err != nil {
		return nil, err
	}

	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("replace: first argument must be string")
	}

	old, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("replace: second argument must be string")
	}

	newStr, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("replace: third argument must be string")
	}

	return stringsReplace(s, old, newStr), nil
}

func stringsReplace(s, old, newStr string) string {
	if len(old) == 0 {
		return s
	}

	result := ""
	i := 0
	for i < len(s) {
		if i <= len(s)-len(old) {
			match := true
			for j := 0; j < len(old); j++ {
				if s[i+j] != old[j] {
					match = false
					break
				}
			}
			if match {
				result += newStr
				i += len(old)
				continue
			}
		}
		result += string(s[i])
		i++
	}
	return result
}

func funcSplit(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("split", args, 2); err != nil {
		return nil, err
	}

	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("split: first argument must be string")
	}

	sep, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("split: second argument must be string")
	}

	if sep == "" {
		result := make([]interface{}, len(s))
		for i, c := range s {
			result[i] = string(c)
		}
		return result, nil
	}

	parts := stringsSplit(s, sep)
	result := make([]interface{}, len(parts))
	for i, p := range parts {
		result[i] = p
	}
	return result, nil
}

func stringsSplit(s, sep string) []string {
	if sep == "" {
		result := make([]string, len(s))
		for i, c := range s {
			result[i] = string(c)
		}
		return result
	}

	var result []string
	start := 0

	for i := 0; i <= len(s)-len(sep); {
		match := true
		for j := 0; j < len(sep); j++ {
			if s[i+j] != sep[j] {
				match = false
				break
			}
		}
		if match {
			result = append(result, s[start:i])
			start = i + len(sep)
			i = start
		} else {
			i++
		}
	}
	result = append(result, s[start:])
	return result
}

func funcJoin(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("join", args, 2); err != nil {
		return nil, err
	}

	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("join: first argument must be array")
	}

	sep, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("join: second argument must be string")
	}

	strs := make([]string, len(arr))
	for i, v := range arr {
		strs[i] = fmt.Sprintf("%v", v)
	}

	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result, nil
}

// ==================== 数组函数实现 ====================

func funcArray(args ...interface{}) (interface{}, error) {
	return args, nil
}

func funcFirst(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("first", args, 1); err != nil {
		return nil, err
	}

	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("first: argument must be array")
	}

	if len(arr) == 0 {
		return nil, fmt.Errorf("first: array is empty")
	}

	return arr[0], nil
}

func funcLast(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("last", args, 1); err != nil {
		return nil, err
	}

	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("last: argument must be array")
	}

	if len(arr) == 0 {
		return nil, fmt.Errorf("last: array is empty")
	}

	return arr[len(arr)-1], nil
}

func funcNth(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("nth", args, 2); err != nil {
		return nil, err
	}

	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("nth: first argument must be array")
	}

	index, err := toInt64(args[1])
	if err != nil {
		return nil, err
	}

	if index < 0 {
		index = int64(len(arr)) + index
	}

	if index < 0 || index >= int64(len(arr)) {
		return nil, fmt.Errorf("nth: index out of range")
	}

	return arr[index], nil
}

func funcSlice(args ...interface{}) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("slice requires 2 or 3 arguments")
	}

	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("slice: first argument must be array")
	}

	start, err := toInt64(args[1])
	if err != nil {
		return nil, err
	}

	if start < 0 {
		start = int64(len(arr)) + start
	}
	if start < 0 {
		start = 0
	}
	if start > int64(len(arr)) {
		return []interface{}{}, nil
	}

	if len(args) == 2 {
		return arr[start:], nil
	}

	end, err := toInt64(args[2])
	if err != nil {
		return nil, err
	}

	if end < 0 {
		end = int64(len(arr)) + end
	}
	if end > int64(len(arr)) {
		end = int64(len(arr))
	}

	if end <= start {
		return []interface{}{}, nil
	}

	return arr[start:end], nil
}

func funcPush(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("push requires at least 2 arguments")
	}

	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("push: first argument must be array")
	}

	result := make([]interface{}, len(arr)+len(args)-1)
	copy(result, arr)
	for i, v := range args[1:] {
		result[len(arr)+i] = v
	}
	return result, nil
}

func funcPop(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("pop", args, 1); err != nil {
		return nil, err
	}

	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("pop: argument must be array")
	}

	if len(arr) == 0 {
		return nil, fmt.Errorf("pop: array is empty")
	}

	return arr[:len(arr)-1], nil
}

func funcReverse(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("reverse", args, 1); err != nil {
		return nil, err
	}

	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("reverse: argument must be array")
	}

	result := make([]interface{}, len(arr))
	for i, v := range arr {
		result[len(arr)-1-i] = v
	}
	return result, nil
}

func funcSort(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("sort", args, 1); err != nil {
		return nil, err
	}

	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("sort: argument must be array")
	}

	// 尝试作为数字排序
	numbers := make([]float64, 0, len(arr))
	allNumbers := true
	for _, v := range arr {
		if n, ok := v.(float64); ok {
			numbers = append(numbers, n)
		} else {
			allNumbers = false
			break
		}
	}

	if allNumbers {
		sort.Float64s(numbers)
		result := make([]interface{}, len(numbers))
		for i, n := range numbers {
			result[i] = n
		}
		return result, nil
	}

	// 作为字符串排序
	strs := make([]string, len(arr))
	for i, v := range arr {
		strs[i] = fmt.Sprintf("%v", v)
	}
	sort.Strings(strs)

	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result, nil
}

func funcFilter(args ...interface{}) (interface{}, error) {
	// filter(array, predicate) - 简化实现，predicate 为条件表达式
	if err := checkArgCount("filter", args, 2); err != nil {
		return nil, err
	}

	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("filter: first argument must be array")
	}

	// predicate 暂时只支持函数名
	predicateName, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("filter: second argument must be function name string")
	}

	result := make([]interface{}, 0)
	for _, v := range arr {
		// 简化：如果值不为 nil 且不为零值，则保留
		if predicateName == "truthy" {
			if isTruthy(v) {
				result = append(result, v)
			}
		}
	}
	return result, nil
}

func funcMap(args ...interface{}) (interface{}, error) {
	// map(array, func) - 简化实现
	if err := checkArgCount("map", args, 2); err != nil {
		return nil, err
	}

	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("map: first argument must be array")
	}

	// func 暂时只支持函数名
	funcName, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("map: second argument must be function name string")
	}

	result := make([]interface{}, len(arr))
	for i, v := range arr {
		// 简化：应用简单转换
		switch funcName {
		case "double":
			if n, ok := v.(float64); ok {
				result[i] = n * 2
			} else {
				result[i] = v
			}
		case "string":
			result[i] = fmt.Sprintf("%v", v)
		default:
			result[i] = v
		}
	}
	return result, nil
}

func funcReduce(args ...interface{}) (interface{}, error) {
	// reduce(array, func, initial) - 简化实现
	if len(args) < 3 {
		return nil, fmt.Errorf("reduce requires 3 arguments")
	}

	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("reduce: first argument must be array")
	}

	// func 暂时只支持函数名
	funcName, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("reduce: second argument must be function name string")
	}

	initial := args[2]

	accumulator := initial
	for _, v := range arr {
		switch funcName {
		case "sum":
			if n, ok := v.(float64); ok {
				if acc, ok := accumulator.(float64); ok {
					accumulator = acc + n
				}
			}
		case "product":
			if n, ok := v.(float64); ok {
				if acc, ok := accumulator.(float64); ok {
					accumulator = acc * n
				}
			}
		}
	}
	return accumulator, nil
}

// ==================== 数学常量函数实现 ====================

func funcPi(args ...interface{}) (interface{}, error) {
	return math.Pi, nil
}

func funcE(args ...interface{}) (interface{}, error) {
	return math.E, nil
}

func funcPhi(args ...interface{}) (interface{}, error) {
	return (1 + math.Sqrt(5)) / 2, nil
}

// ==================== 其他实用函数实现 ====================

func funcClamp(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("clamp", args, 3); err != nil {
		return nil, err
	}

	value, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}

	min, err := toFloat64(args[1])
	if err != nil {
		return nil, err
	}

	max, err := toFloat64(args[2])
	if err != nil {
		return nil, err
	}

	if value < min {
		return min, nil
	}
	if value > max {
		return max, nil
	}
	return value, nil
}

func funcLerp(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("lerp", args, 3); err != nil {
		return nil, err
	}

	a, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}

	b, err := toFloat64(args[1])
	if err != nil {
		return nil, err
	}

	t, err := toFloat64(args[2])
	if err != nil {
		return nil, err
	}

	return a + (b-a)*t, nil
}

func funcStep(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("step", args, 2); err != nil {
		return nil, err
	}

	edge, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}

	x, err := toFloat64(args[1])
	if err != nil {
		return nil, err
	}

	if x < edge {
		return 0.0, nil
	}
	return 1.0, nil
}

func funcSmoothstep(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("smoothstep", args, 3); err != nil {
		return nil, err
	}

	edge0, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}

	edge1, err := toFloat64(args[1])
	if err != nil {
		return nil, err
	}

	x, err := toFloat64(args[2])
	if err != nil {
		return nil, err
	}

	// Clamp x to [0, 1]
	t := (x - edge0) / (edge1 - edge0)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	// Smoothstep formula
	return t * t * (3 - 2*t), nil
}

func funcMod(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("mod", args, 2); err != nil {
		return nil, err
	}

	x, err := toFloat64(args[0])
	if err != nil {
		return nil, err
	}

	y, err := toFloat64(args[1])
	if err != nil {
		return nil, err
	}

	if y == 0 {
		return nil, fmt.Errorf("mod: division by zero")
	}

	return math.Mod(x, y), nil
}

func funcGcd(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("gcd", args, 2); err != nil {
		return nil, err
	}

	a, err := toInt64(args[0])
	if err != nil {
		return nil, err
	}

	b, err := toInt64(args[1])
	if err != nil {
		return nil, err
	}

	// 确保 a 和 b 为正数
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}

	for b != 0 {
		a, b = b, a%b
	}

	return float64(a), nil
}

func funcLcm(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("lcm", args, 2); err != nil {
		return nil, err
	}

	a, err := toInt64(args[0])
	if err != nil {
		return nil, err
	}

	b, err := toInt64(args[1])
	if err != nil {
		return nil, err
	}

	if a == 0 || b == 0 {
		return 0.0, nil
	}

	// 确保 a 和 b 为正数
	absA := a
	absB := b
	if a < 0 {
		absA = -a
	}
	if b < 0 {
		absB = -b
	}

	gcd, _ := funcGcd(float64(absA), float64(absB))
	return float64(absA*absB) / gcd.(float64), nil
}

func funcIsFinite(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("isFinite", args, 1); err != nil {
		return nil, err
	}

	x, err := toFloat64(args[0])
	if err != nil {
		return false, nil
	}

	return !math.IsInf(x, 0) && !math.IsNaN(x), nil
}

func funcIsInf(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("isInf", args, 1); err != nil {
		return nil, err
	}

	x, err := toFloat64(args[0])
	if err != nil {
		return false, nil
	}

	return math.IsInf(x, 0), nil
}

func funcIsNaN(args ...interface{}) (interface{}, error) {
	if err := checkArgCount("isNaN", args, 1); err != nil {
		return nil, err
	}

	x, err := toFloat64(args[0])
	if err != nil {
		return false, nil
	}

	return math.IsNaN(x), nil
}

func funcRandom(args ...interface{}) (interface{}, error) {
	// 简化的随机数生成器
	// 使用简单的线性同余生成器
	seed := int64(12345)
	if len(args) > 0 {
		if s, err := toInt64(args[0]); err == nil {
			seed = s
		}
	}

	// LCG parameters
	const a = 1103515245
	const c = 12345
	const m = 1 << 31

	seed = (a*seed + c) % m
	return float64(seed) / float64(m), nil
}

// ==================== 辅助函数 ====================

// checkArgCount 检查参数数量
func checkArgCount(name string, args []interface{}, expected int) error {
	if len(args) != expected {
		return fmt.Errorf("%s requires %d arguments, got %d", name, expected, len(args))
	}
	return nil
}

// extractNumbers 从参数中提取数字
func extractNumbers(args []interface{}) ([]float64, error) {
	numbers := make([]float64, 0)

	for _, arg := range args {
		switch v := arg.(type) {
		case float64:
			numbers = append(numbers, v)
		case []interface{}:
			for _, elem := range v {
				if n, ok := elem.(float64); ok {
					numbers = append(numbers, n)
				}
			}
		default:
			n, err := toFloat64(v)
			if err != nil {
				return nil, fmt.Errorf("cannot convert %T to number", v)
			}
			numbers = append(numbers, n)
		}
	}

	return numbers, nil
}

// isTruthy 判断值是否为真
func isTruthy(v interface{}) bool {
	if v == nil {
		return false
	}

	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case string:
		return val != ""
	case []interface{}:
		return len(val) > 0
	default:
		return true
	}
}
