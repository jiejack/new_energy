package formula

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Executor 公式执行器
type Executor struct {
	functionRegistry *FunctionRegistry
	operatorRegistry *OperatorRegistry
	cache            *ResultCache
	config           *ExecutorConfig
	mu               sync.RWMutex
}

// ExecutorConfig 执行器配置
type ExecutorConfig struct {
	EnableCache       bool          // 是否启用缓存
	CacheTTL          time.Duration // 缓存过期时间
	MaxCacheSize      int           // 最大缓存数量
	Timeout           time.Duration // 执行超时时间
	MaxRecursionDepth int           // 最大递归深度
}

// DefaultExecutorConfig 默认执行器配置
func DefaultExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		EnableCache:       true,
		CacheTTL:          5 * time.Minute,
		MaxCacheSize:      10000,
		Timeout:           30 * time.Second,
		MaxRecursionDepth: 100,
	}
}

// NewExecutor 创建公式执行器
func NewExecutor(config *ExecutorConfig) *Executor {
	if config == nil {
		config = DefaultExecutorConfig()
	}

	executor := &Executor{
		functionRegistry: NewFunctionRegistry(),
		operatorRegistry: NewOperatorRegistry(),
		config:           config,
	}

	if config.EnableCache {
		executor.cache = NewResultCache(config.MaxCacheSize, config.CacheTTL)
	}

	return executor
}

// Execute 执行公式计算
func (e *Executor) Execute(formula string, variables map[string]interface{}) (interface{}, error) {
	return e.ExecuteWithContext(context.Background(), formula, variables)
}

// ExecuteWithContext 带上下文执行公式计算
func (e *Executor) ExecuteWithContext(ctx context.Context, formula string, variables map[string]interface{}) (interface{}, error) {
	// 检查缓存
	if e.config.EnableCache && e.cache != nil {
		cacheKey := e.buildCacheKey(formula, variables)
		if result, exists := e.cache.Get(cacheKey); exists {
			return result, nil
		}
	}

	// 创建求值上下文
	evalCtx := e.createEvalContext(variables)

	// 解析公式
	node, err := ParseFormula(formula)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	// 设置超时
	if e.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.config.Timeout)
		defer cancel()
	}

	// 执行求值
	result, err := e.evaluateNode(ctx, node, evalCtx, 0)
	if err != nil {
		return nil, fmt.Errorf("evaluation error: %w", err)
	}

	// 缓存结果
	if e.config.EnableCache && e.cache != nil {
		cacheKey := e.buildCacheKey(formula, variables)
		e.cache.Set(cacheKey, result)
	}

	return result, nil
}

// ExecuteBatch 批量执行公式
func (e *Executor) ExecuteBatch(formulas []string, variables map[string]interface{}) ([]interface{}, []error) {
	return e.ExecuteBatchWithContext(context.Background(), formulas, variables)
}

// ExecuteBatchWithContext 带上下文批量执行公式
func (e *Executor) ExecuteBatchWithContext(ctx context.Context, formulas []string, variables map[string]interface{}) ([]interface{}, []error) {
	results := make([]interface{}, len(formulas))
	errors := make([]error, len(formulas))

	var wg sync.WaitGroup

	for i, formula := range formulas {
		wg.Add(1)
		go func(index int, f string) {
			defer wg.Done()
			result, err := e.ExecuteWithContext(ctx, f, variables)
			results[index] = result
			errors[index] = err
		}(i, formula)
	}

	wg.Wait()
	return results, errors
}

// ExecuteParallel 并行执行公式
func (e *Executor) ExecuteParallel(formulas []string, variables map[string]interface{}, workers int) ([]interface{}, []error) {
	return e.ExecuteParallelWithContext(context.Background(), formulas, variables, workers)
}

// ExecuteParallelWithContext 带上下文并行执行公式
func (e *Executor) ExecuteParallelWithContext(ctx context.Context, formulas []string, variables map[string]interface{}, workers int) ([]interface{}, []error) {
	if workers <= 0 {
		workers = 4
	}

	results := make([]interface{}, len(formulas))
	errors := make([]error, len(formulas))

	// 创建工作通道
	formulaChan := make(chan struct {
		index   int
		formula string
	}, len(formulas))

	resultChan := make(chan struct {
		index  int
		result interface{}
		err    error
	}, len(formulas))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range formulaChan {
				result, err := e.ExecuteWithContext(ctx, item.formula, variables)
				resultChan <- struct {
					index  int
					result interface{}
					err    error
				}{item.index, result, err}
			}
		}()
	}

	// 发送任务
	go func() {
		for i, formula := range formulas {
			formulaChan <- struct {
				index   int
				formula string
			}{i, formula}
		}
		close(formulaChan)
	}()

	// 等待完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	for item := range resultChan {
		results[item.index] = item.result
		errors[item.index] = item.err
	}

	return results, errors
}

// ExecuteWithTimeout 带超时执行公式
func (e *Executor) ExecuteWithTimeout(formula string, variables map[string]interface{}, timeout time.Duration) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return e.ExecuteWithContext(ctx, formula, variables)
}

// ExecuteNode 执行AST节点
func (e *Executor) ExecuteNode(node Node, variables map[string]interface{}) (interface{}, error) {
	evalCtx := e.createEvalContext(variables)
	return node.Eval(evalCtx)
}

// RegisterFunction 注册自定义函数
func (e *Executor) RegisterFunction(name string, fn Function) {
	e.functionRegistry.Register(name, fn)
}

// GetFunction 获取已注册的函数
func (e *Executor) GetFunction(name string) (Function, bool) {
	return e.functionRegistry.Get(name)
}

// ClearCache 清除缓存
func (e *Executor) ClearCache() {
	if e.cache != nil {
		e.cache.Clear()
	}
}

// GetCacheStats 获取缓存统计信息
func (e *Executor) GetCacheStats() *CacheStats {
	if e.cache == nil {
		return nil
	}
	return e.cache.Stats()
}

// createEvalContext 创建求值上下文
func (e *Executor) createEvalContext(variables map[string]interface{}) *EvalContext {
	ctx := NewEvalContext()

	// 复制变量
	if variables != nil {
		for k, v := range variables {
			ctx.SetVariable(k, v)
		}
	}

	// 复制函数
	for name, fn := range e.functionRegistry.GetAll() {
		ctx.RegisterFunction(name, fn)
	}

	return ctx
}

// evaluateNode 求值AST节点
func (e *Executor) evaluateNode(ctx context.Context, node Node, evalCtx *EvalContext, depth int) (interface{}, error) {
	// 检查递归深度
	if depth > e.config.MaxRecursionDepth {
		return nil, fmt.Errorf("maximum recursion depth exceeded")
	}

	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return node.Eval(evalCtx)
}

// buildCacheKey 构建缓存键
func (e *Executor) buildCacheKey(formula string, variables map[string]interface{}) string {
	// 简化实现：使用公式和变量组合作为键
	key := formula
	if variables != nil {
		for k, v := range variables {
			key += fmt.Sprintf("|%s=%v", k, v)
		}
	}
	return key
}

// ==================== 结果缓存 ====================

// ResultCache 结果缓存
type ResultCache struct {
	items    map[string]*cacheItem
	maxSize  int
	ttl      time.Duration
	mu       sync.RWMutex
	stats    CacheStats
}

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Hits      int64
	Misses    int64
	Evictions int64
	Size      int
}

// NewResultCache 创建结果缓存
func NewResultCache(maxSize int, ttl time.Duration) *ResultCache {
	return &ResultCache{
		items:   make(map[string]*cacheItem),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// Get 获取缓存值
func (c *ResultCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.expiresAt) {
		c.stats.Misses++
		return nil, false
	}

	c.stats.Hits++
	return item.value, true
}

// Set 设置缓存值
func (c *ResultCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否需要淘汰
	if len(c.items) >= c.maxSize {
		c.evict()
	}

	c.items[key] = &cacheItem{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.stats.Size = len(c.items)
}

// Delete 删除缓存值
func (c *ResultCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	c.stats.Size = len(c.items)
}

// Clear 清除所有缓存
func (c *ResultCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*cacheItem)
	c.stats.Size = 0
}

// Stats 获取缓存统计信息
func (c *ResultCache) Stats() *CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = len(c.items)
	return &stats
}

// evict 淘汰缓存项
func (c *ResultCache) evict() {
	// 简单的淘汰策略：删除过期的项
	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, key)
			c.stats.Evictions++
		}
	}

	// 如果仍然超过限制，删除一半
	if len(c.items) >= c.maxSize {
		count := 0
		for key := range c.items {
			if count >= c.maxSize/2 {
				break
			}
			delete(c.items, key)
			c.stats.Evictions++
			count++
		}
	}
}

// ==================== 执行结果 ====================

// ExecutionResult 执行结果
type ExecutionResult struct {
	Formula    string
	Result     interface{}
	Error      error
	Duration   time.Duration
	CacheHit   bool
	Variables  map[string]interface{}
}

// BatchExecutionResult 批量执行结果
type BatchExecutionResult struct {
	Results    []*ExecutionResult
	TotalTime  time.Duration
	Success    int
	Failed     int
	CacheHits  int
}

// ==================== 执行器构建器 ====================

// ExecutorBuilder 执行器构建器
type ExecutorBuilder struct {
	config     *ExecutorConfig
	functions  map[string]Function
}

// NewExecutorBuilder 创建执行器构建器
func NewExecutorBuilder() *ExecutorBuilder {
	return &ExecutorBuilder{
		config:    DefaultExecutorConfig(),
		functions: make(map[string]Function),
	}
}

// WithCache 设置缓存配置
func (b *ExecutorBuilder) WithCache(enable bool, ttl time.Duration, maxSize int) *ExecutorBuilder {
	b.config.EnableCache = enable
	b.config.CacheTTL = ttl
	b.config.MaxCacheSize = maxSize
	return b
}

// WithTimeout 设置超时时间
func (b *ExecutorBuilder) WithTimeout(timeout time.Duration) *ExecutorBuilder {
	b.config.Timeout = timeout
	return b
}

// WithMaxRecursionDepth 设置最大递归深度
func (b *ExecutorBuilder) WithMaxRecursionDepth(depth int) *ExecutorBuilder {
	b.config.MaxRecursionDepth = depth
	return b
}

// WithFunction 注册自定义函数
func (b *ExecutorBuilder) WithFunction(name string, fn Function) *ExecutorBuilder {
	b.functions[name] = fn
	return b
}

// Build 构建执行器
func (b *ExecutorBuilder) Build() *Executor {
	executor := NewExecutor(b.config)

	// 注册自定义函数
	for name, fn := range b.functions {
		executor.RegisterFunction(name, fn)
	}

	return executor
}

// ==================== 变量绑定器 ====================

// VariableBinder 变量绑定器
type VariableBinder struct {
	variables map[string]interface{}
	mu        sync.RWMutex
}

// NewVariableBinder 创建变量绑定器
func NewVariableBinder() *VariableBinder {
	return &VariableBinder{
		variables: make(map[string]interface{}),
	}
}

// Bind 绑定变量
func (b *VariableBinder) Bind(name string, value interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.variables[name] = value
}

// BindMany 批量绑定变量
func (b *VariableBinder) BindMany(vars map[string]interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for k, v := range vars {
		b.variables[k] = v
	}
}

// Unbind 解绑变量
func (b *VariableBinder) Unbind(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.variables, name)
}

// Get 获取变量
func (b *VariableBinder) Get(name string) (interface{}, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	val, exists := b.variables[name]
	return val, exists
}

// GetAll 获取所有变量
func (b *VariableBinder) GetAll() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range b.variables {
		result[k] = v
	}
	return result
}

// Clear 清除所有变量
func (b *VariableBinder) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.variables = make(map[string]interface{})
}

// ==================== 执行上下文 ====================

// ExecutionContext 执行上下文
type ExecutionContext struct {
	Variables    map[string]interface{}
	Functions    map[string]Function
	Metadata     map[string]interface{}
	Parent       *ExecutionContext
	executor     *Executor
}

// NewExecutionContext 创建执行上下文
func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		Variables: make(map[string]interface{}),
		Functions: make(map[string]Function),
		Metadata:  make(map[string]interface{}),
	}
}

// SetVariable 设置变量
func (c *ExecutionContext) SetVariable(name string, value interface{}) {
	c.Variables[name] = value
}

// GetVariable 获取变量（支持向上查找父上下文）
func (c *ExecutionContext) GetVariable(name string) (interface{}, bool) {
	if val, exists := c.Variables[name]; exists {
		return val, true
	}
	if c.Parent != nil {
		return c.Parent.GetVariable(name)
	}
	return nil, false
}

// RegisterFunction 注册函数
func (c *ExecutionContext) RegisterFunction(name string, fn Function) {
	c.Functions[name] = fn
}

// GetFunction 获取函数（支持向上查找父上下文）
func (c *ExecutionContext) GetFunction(name string) (Function, bool) {
	if fn, exists := c.Functions[name]; exists {
		return fn, true
	}
	if c.Parent != nil {
		return c.Parent.GetFunction(name)
	}
	return nil, false
}

// SetMetadata 设置元数据
func (c *ExecutionContext) SetMetadata(key string, value interface{}) {
	c.Metadata[key] = value
}

// GetMetadata 获取元数据
func (c *ExecutionContext) GetMetadata(key string) (interface{}, bool) {
	if val, exists := c.Metadata[key]; exists {
		return val, true
	}
	if c.Parent != nil {
		return c.Parent.GetMetadata(key)
	}
	return nil, false
}

// CreateChild 创建子上下文
func (c *ExecutionContext) CreateChild() *ExecutionContext {
	return &ExecutionContext{
		Variables: make(map[string]interface{}),
		Functions: make(map[string]Function),
		Metadata:  make(map[string]interface{}),
		Parent:    c,
		executor:  c.executor,
	}
}

// ToEvalContext 转换为求值上下文
func (c *ExecutionContext) ToEvalContext() *EvalContext {
	ctx := NewEvalContext()

	// 收集所有变量（包括父上下文）
	current := c
	for current != nil {
		for k, v := range current.Variables {
			if _, exists := ctx.Variables[k]; !exists {
				ctx.Variables[k] = v
			}
		}
		current = current.Parent
	}

	// 收集所有函数（包括父上下文）
	current = c
	for current != nil {
		for k, v := range current.Functions {
			if _, exists := ctx.Functions[k]; !exists {
				ctx.Functions[k] = v
			}
		}
		current = current.Parent
	}

	return ctx
}

// ==================== 公式编译器 ====================

// CompiledFormula 编译后的公式
type CompiledFormula struct {
	formula   string
	ast       Node
	variables []string
	functions []string
}

// Compile 编译公式
func (e *Executor) Compile(formula string) (*CompiledFormula, error) {
	ast, err := ParseFormula(formula)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	compiled := &CompiledFormula{
		formula:   formula,
		ast:       ast,
		variables: extractVariables(ast),
		functions: extractFunctions(ast),
	}

	return compiled, nil
}

// Execute 执行编译后的公式
func (cf *CompiledFormula) Execute(executor *Executor, variables map[string]interface{}) (interface{}, error) {
	evalCtx := executor.createEvalContext(variables)
	return cf.ast.Eval(evalCtx)
}

// GetVariables 获取公式中使用的变量
func (cf *CompiledFormula) GetVariables() []string {
	return cf.variables
}

// GetFunctions 获取公式中使用的函数
func (cf *CompiledFormula) GetFunctions() []string {
	return cf.functions
}

// extractVariables 从AST中提取变量
func extractVariables(node Node) []string {
	variables := make(map[string]bool)
	collectVariables(node, variables)

	result := make([]string, 0, len(variables))
	for v := range variables {
		result = append(result, v)
	}
	return result
}

func collectVariables(node Node, variables map[string]bool) {
	switch n := node.(type) {
	case *VariableNode:
		variables[n.Name] = true
	case *BinaryOpNode:
		collectVariables(n.Left, variables)
		collectVariables(n.Right, variables)
	case *UnaryOpNode:
		collectVariables(n.Operand, variables)
	case *FunctionCallNode:
		for _, arg := range n.Arguments {
			collectVariables(arg, variables)
		}
	case *ConditionalNode:
		collectVariables(n.Condition, variables)
		collectVariables(n.ThenExpr, variables)
		collectVariables(n.ElseExpr, variables)
	case *ArrayNode:
		for _, elem := range n.Elements {
			collectVariables(elem, variables)
		}
	}
}

// extractFunctions 从AST中提取函数
func extractFunctions(node Node) []string {
	functions := make(map[string]bool)
	collectFunctions(node, functions)

	result := make([]string, 0, len(functions))
	for f := range functions {
		result = append(result, f)
	}
	return result
}

func collectFunctions(node Node, functions map[string]bool) {
	switch n := node.(type) {
	case *BinaryOpNode:
		collectFunctions(n.Left, functions)
		collectFunctions(n.Right, functions)
	case *UnaryOpNode:
		collectFunctions(n.Operand, functions)
	case *FunctionCallNode:
		functions[n.Name] = true
		for _, arg := range n.Arguments {
			collectFunctions(arg, functions)
		}
	case *ConditionalNode:
		collectFunctions(n.Condition, functions)
		collectFunctions(n.ThenExpr, functions)
		collectFunctions(n.ElseExpr, functions)
	case *ArrayNode:
		for _, elem := range n.Elements {
			collectFunctions(elem, functions)
		}
	}
}
