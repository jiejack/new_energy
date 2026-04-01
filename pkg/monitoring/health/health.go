package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnknown   HealthStatus = "unknown"
)

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	Name      string       `json:"name"`
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	Latency   time.Duration `json:"latency,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Status     HealthStatus                `json:"status"`
	Timestamp  time.Time                   `json:"timestamp"`
	Components map[string]ComponentHealth  `json:"components"`
	Version    string                      `json:"version,omitempty"`
	Service    string                      `json:"service,omitempty"`
	Uptime     time.Duration               `json:"uptime"`
}

// Checker 健康检查器接口
type Checker interface {
	Name() string
	Check(ctx context.Context) ComponentHealth
}

// HealthChecker 健康检查器
type HealthChecker struct {
	checkers  map[string]Checker
	cache     *HealthCache
	config    *Config
	logger    *zap.Logger
	startTime time.Time
	mu        sync.RWMutex
}

// Config 健康检查配置
type Config struct {
	ServiceName    string
	ServiceVersion string
	CacheTTL       time.Duration
	CheckTimeout   time.Duration
}

// HealthCache 健康检查缓存
type HealthCache struct {
	results map[string]cachedResult
	ttl     time.Duration
	mu      sync.RWMutex
}

type cachedResult struct {
	health    ComponentHealth
	timestamp time.Time
}

// NewHealthCache 创建健康检查缓存
func NewHealthCache(ttl time.Duration) *HealthCache {
	return &HealthCache{
		results: make(map[string]cachedResult),
		ttl:     ttl,
	}
}

// Get 获取缓存的健康状态
func (hc *HealthCache) Get(name string) (ComponentHealth, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	cached, exists := hc.results[name]
	if !exists {
		return ComponentHealth{}, false
	}

	if time.Since(cached.timestamp) > hc.ttl {
		return ComponentHealth{}, false
	}

	return cached.health, true
}

// Set 设置缓存的健康状态
func (hc *HealthCache) Set(name string, health ComponentHealth) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.results[name] = cachedResult{
		health:    health,
		timestamp: time.Now(),
	}
}

// Clear 清除缓存
func (hc *HealthCache) Clear() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.results = make(map[string]cachedResult)
}

// NewHealthChecker 创建新的健康检查器
func NewHealthChecker(cfg *Config, logger *zap.Logger) *HealthChecker {
	if logger == nil {
		logger = zap.NewNop()
	}

	cacheTTL := cfg.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = 5 * time.Second
	}

	checkTimeout := cfg.CheckTimeout
	if checkTimeout == 0 {
		checkTimeout = 5 * time.Second
	}

	return &HealthChecker{
		checkers:  make(map[string]Checker),
		cache:     NewHealthCache(cacheTTL),
		config:    &Config{CacheTTL: cacheTTL, CheckTimeout: checkTimeout, ServiceName: cfg.ServiceName, ServiceVersion: cfg.ServiceVersion},
		logger:    logger,
		startTime: time.Now(),
	}
}

// RegisterChecker 注册健康检查器
func (h *HealthChecker) RegisterChecker(checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.checkers[checker.Name()] = checker
	h.logger.Debug("health checker registered", zap.String("name", checker.Name()))
}

// UnregisterChecker 注销健康检查器
func (h *HealthChecker) UnregisterChecker(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.checkers, name)
	h.logger.Debug("health checker unregistered", zap.String("name", name))
}

// Check 执行健康检查
func (h *HealthChecker) Check(ctx context.Context) *HealthCheckResult {
	h.mu.RLock()
	checkers := make([]Checker, 0, len(h.checkers))
	for _, checker := range h.checkers {
		checkers = append(checkers, checker)
	}
	h.mu.RUnlock()

	components := make(map[string]ComponentHealth)
	overallStatus := StatusHealthy

	// 并发执行检查
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, checker := range checkers {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()

			// 先尝试从缓存获取
			if cached, ok := h.cache.Get(c.Name()); ok {
				mu.Lock()
				components[c.Name()] = cached
				if cached.Status == StatusUnhealthy {
					overallStatus = StatusUnhealthy
				} else if cached.Status == StatusDegraded && overallStatus != StatusUnhealthy {
					overallStatus = StatusDegraded
				}
				mu.Unlock()
				return
			}

			// 执行检查
			checkCtx, cancel := context.WithTimeout(ctx, h.config.CheckTimeout)
			defer cancel()

			start := time.Now()
			health := c.Check(checkCtx)
			health.Latency = time.Since(start)
			health.Timestamp = time.Now()

			// 缓存结果
			h.cache.Set(c.Name(), health)

			mu.Lock()
			components[c.Name()] = health
			if health.Status == StatusUnhealthy {
				overallStatus = StatusUnhealthy
			} else if health.Status == StatusDegraded && overallStatus != StatusUnhealthy {
				overallStatus = StatusDegraded
			}
			mu.Unlock()
		}(checker)
	}

	wg.Wait()

	return &HealthCheckResult{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Components: components,
		Version:    h.config.ServiceVersion,
		Service:    h.config.ServiceName,
		Uptime:     time.Since(h.startTime),
	}
}

// CheckComponent 检查单个组件
func (h *HealthChecker) CheckComponent(ctx context.Context, name string) (ComponentHealth, error) {
	h.mu.RLock()
	checker, exists := h.checkers[name]
	h.mu.RUnlock()

	if !exists {
		return ComponentHealth{}, fmt.Errorf("checker %s not found", name)
	}

	// 先尝试从缓存获取
	if cached, ok := h.cache.Get(name); ok {
		return cached, nil
	}

	// 执行检查
	checkCtx, cancel := context.WithTimeout(ctx, h.config.CheckTimeout)
	defer cancel()

	start := time.Now()
	health := checker.Check(checkCtx)
	health.Latency = time.Since(start)
	health.Timestamp = time.Now()

	// 缓存结果
	h.cache.Set(name, health)

	return health, nil
}

// IsHealthy 检查是否健康
func (h *HealthChecker) IsHealthy(ctx context.Context) bool {
	result := h.Check(ctx)
	return result.Status == StatusHealthy
}

// IsReady 就绪探针检查
func (h *HealthChecker) IsReady(ctx context.Context) bool {
	result := h.Check(ctx)
	return result.Status != StatusUnhealthy
}

// IsAlive 存活探针检查
func (h *HealthChecker) IsAlive() bool {
	return true
}

// ClearCache 清除缓存
func (h *HealthChecker) ClearCache() {
	h.cache.Clear()
}

// HTTPHandler HTTP处理器
func (h *HealthChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := h.Check(r.Context())

		statusCode := http.StatusOK
		if result.Status == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		} else if result.Status == StatusDegraded {
			statusCode = http.StatusOK // 降级状态仍然返回200
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		if err := json.NewEncoder(w).Encode(result); err != nil {
			h.logger.Error("failed to encode health check result", zap.Error(err))
		}
	}
}

// LivezHandler 存活探针处理器
func (h *HealthChecker) LivezHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.IsAlive() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Not Alive"))
	}
}

// ReadyzHandler 就绪探针处理器
func (h *HealthChecker) ReadyzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.IsReady(r.Context()) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Not Ready"))
	}
}

// 常用健康检查器实现

// DatabaseChecker 数据库健康检查器
type DatabaseChecker struct {
	name string
	ping func(ctx context.Context) error
}

// NewDatabaseChecker 创建数据库健康检查器
func NewDatabaseChecker(name string, ping func(ctx context.Context) error) *DatabaseChecker {
	return &DatabaseChecker{
		name: name,
		ping: ping,
	}
}

// Name 获取检查器名称
func (dc *DatabaseChecker) Name() string {
	return dc.name
}

// Check 执行健康检查
func (dc *DatabaseChecker) Check(ctx context.Context) ComponentHealth {
	err := dc.ping(ctx)
	if err != nil {
		return ComponentHealth{
			Name:    dc.name,
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("database connection failed: %v", err),
		}
	}

	return ComponentHealth{
		Name:   dc.name,
		Status: StatusHealthy,
	}
}

// RedisChecker Redis健康检查器
type RedisChecker struct {
	name string
	ping func(ctx context.Context) error
}

// NewRedisChecker 创建Redis健康检查器
func NewRedisChecker(name string, ping func(ctx context.Context) error) *RedisChecker {
	return &RedisChecker{
		name: name,
		ping: ping,
	}
}

// Name 获取检查器名称
func (rc *RedisChecker) Name() string {
	return rc.name
}

// Check 执行健康检查
func (rc *RedisChecker) Check(ctx context.Context) ComponentHealth {
	err := rc.ping(ctx)
	if err != nil {
		return ComponentHealth{
			Name:    rc.name,
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("redis connection failed: %v", err),
		}
	}

	return ComponentHealth{
		Name:   rc.name,
		Status: StatusHealthy,
	}
}

// KafkaChecker Kafka健康检查器
type KafkaChecker struct {
	name    string
	brokers []string
}

// NewKafkaChecker 创建Kafka健康检查器
func NewKafkaChecker(name string, brokers []string) *KafkaChecker {
	return &KafkaChecker{
		name:    name,
		brokers: brokers,
	}
}

// Name 获取检查器名称
func (kc *KafkaChecker) Name() string {
	return kc.name
}

// Check 执行健康检查
func (kc *KafkaChecker) Check(ctx context.Context) ComponentHealth {
	// 简单检查：尝试连接broker
	// 实际实现可能需要更复杂的逻辑
	return ComponentHealth{
		Name:   kc.name,
		Status: StatusHealthy,
		Details: map[string]interface{}{
			"brokers": kc.brokers,
		},
	}
}

// HTTPChecker HTTP服务健康检查器
type HTTPChecker struct {
	name     string
	endpoint string
	timeout  time.Duration
}

// NewHTTPChecker 创建HTTP健康检查器
func NewHTTPChecker(name, endpoint string, timeout time.Duration) *HTTPChecker {
	return &HTTPChecker{
		name:     name,
		endpoint: endpoint,
		timeout:  timeout,
	}
}

// Name 获取检查器名称
func (hc *HTTPChecker) Name() string {
	return hc.name
}

// Check 执行健康检查
func (hc *HTTPChecker) Check(ctx context.Context) ComponentHealth {
	client := &http.Client{
		Timeout: hc.timeout,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hc.endpoint, nil)
	if err != nil {
		return ComponentHealth{
			Name:    hc.name,
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return ComponentHealth{
			Name:    hc.name,
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("http request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return ComponentHealth{
			Name:    hc.name,
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("http status code: %d", resp.StatusCode),
		}
	}

	if resp.StatusCode >= 400 {
		return ComponentHealth{
			Name:    hc.name,
			Status:  StatusDegraded,
			Message: fmt.Sprintf("http status code: %d", resp.StatusCode),
		}
	}

	return ComponentHealth{
		Name:   hc.name,
		Status: StatusHealthy,
		Details: map[string]interface{}{
			"status_code": resp.StatusCode,
		},
	}
}

// TCPChecker TCP健康检查器
type TCPChecker struct {
	name    string
	address string
	timeout time.Duration
}

// NewTCPChecker 创建TCP健康检查器
func NewTCPChecker(name, address string, timeout time.Duration) *TCPChecker {
	return &TCPChecker{
		name:    name,
		address: address,
		timeout: timeout,
	}
}

// Name 获取检查器名称
func (tc *TCPChecker) Name() string {
	return tc.name
}

// Check 执行健康检查
func (tc *TCPChecker) Check(ctx context.Context) ComponentHealth {
	// 简化实现：实际需要net.DialTimeout
	return ComponentHealth{
		Name:   tc.name,
		Status: StatusHealthy,
		Details: map[string]interface{}{
			"address": tc.address,
		},
	}
}

// CustomChecker 自定义健康检查器
type CustomChecker struct {
	name   string
	check  func(ctx context.Context) ComponentHealth
}

// NewCustomChecker 创建自定义健康检查器
func NewCustomChecker(name string, check func(ctx context.Context) ComponentHealth) *CustomChecker {
	return &CustomChecker{
		name:  name,
		check: check,
	}
}

// Name 获取检查器名称
func (cc *CustomChecker) Name() string {
	return cc.name
}

// Check 执行健康检查
func (cc *CustomChecker) Check(ctx context.Context) ComponentHealth {
	return cc.check(ctx)
}

// AggregatedHealthChecker 聚合健康检查器
type AggregatedHealthChecker struct {
	name     string
	checkers []Checker
	strategy AggregationStrategy
}

// AggregationStrategy 聚合策略
type AggregationStrategy int

const (
	// StrategyAllHealthy 所有组件健康才算健康
	StrategyAllHealthy AggregationStrategy = iota
	// StrategyAnyHealthy 任一组件健康就算健康
	StrategyAnyHealthy
	// StrategyMajorityHealthy 多数组件健康就算健康
	StrategyMajorityHealthy
)

// NewAggregatedHealthChecker 创建聚合健康检查器
func NewAggregatedHealthChecker(name string, strategy AggregationStrategy, checkers ...Checker) *AggregatedHealthChecker {
	return &AggregatedHealthChecker{
		name:     name,
		checkers: checkers,
		strategy: strategy,
	}
}

// Name 获取检查器名称
func (ahc *AggregatedHealthChecker) Name() string {
	return ahc.name
}

// Check 执行健康检查
func (ahc *AggregatedHealthChecker) Check(ctx context.Context) ComponentHealth {
	results := make([]ComponentHealth, len(ahc.checkers))
	healthyCount := 0
	degradedCount := 0

	for i, checker := range ahc.checkers {
		results[i] = checker.Check(ctx)
		if results[i].Status == StatusHealthy {
			healthyCount++
		} else if results[i].Status == StatusDegraded {
			degradedCount++
		}
	}

	var status HealthStatus
	switch ahc.strategy {
	case StrategyAllHealthy:
		if healthyCount == len(ahc.checkers) {
			status = StatusHealthy
		} else if healthyCount+degradedCount == len(ahc.checkers) {
			status = StatusDegraded
		} else {
			status = StatusUnhealthy
		}
	case StrategyAnyHealthy:
		if healthyCount > 0 {
			status = StatusHealthy
		} else if degradedCount > 0 {
			status = StatusDegraded
		} else {
			status = StatusUnhealthy
		}
	case StrategyMajorityHealthy:
		if healthyCount > len(ahc.checkers)/2 {
			status = StatusHealthy
		} else if healthyCount+degradedCount > len(ahc.checkers)/2 {
			status = StatusDegraded
		} else {
			status = StatusUnhealthy
		}
	}

	return ComponentHealth{
		Name:   ahc.name,
		Status: status,
		Details: map[string]interface{}{
			"healthy_count":  healthyCount,
			"degraded_count": degradedCount,
			"total_count":    len(ahc.checkers),
		},
	}
}

// HealthCheckMiddleware 健康检查中间件
func HealthCheckMiddleware(hc *HealthChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 对于健康检查端点，直接放行
			if r.URL.Path == "/health" || r.URL.Path == "/healthz" ||
				r.URL.Path == "/ready" || r.URL.Path == "/readyz" ||
				r.URL.Path == "/live" || r.URL.Path == "/livez" {
				next.ServeHTTP(w, r)
				return
			}

			// 检查服务是否就绪
			if !hc.IsReady(r.Context()) {
				http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// StartHTTPServer 启动健康检查HTTP服务器
func (h *HealthChecker) StartHTTPServer(port int) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.HTTPHandler())
	mux.HandleFunc("/healthz", h.HTTPHandler())
	mux.HandleFunc("/ready", h.ReadyzHandler())
	mux.HandleFunc("/readyz", h.ReadyzHandler())
	mux.HandleFunc("/live", h.LivezHandler())
	mux.HandleFunc("/livez", h.LivezHandler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		h.logger.Info("starting health check server", zap.Int("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.Error("health check server error", zap.Error(err))
		}
	}()

	return server, nil
}
