package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type SkillType string

const (
	SkillTypeGStack     SkillType = "gstack"
	SkillTypeSuperpower SkillType = "superpower"
)

type SkillCall struct {
	Type     SkillType     `json:"type"`
	Name     string        `json:"name"`
	Input    interface{}   `json:"input,omitempty"`
	Timeout  time.Duration `json:"timeout,omitempty"`
	Priority int           `json:"priority,omitempty"`
}

type SkillResult struct {
	Success bool        `json:"success"`
	Output  interface{} `json:"output,omitempty"`
	Error   string      `json:"error,omitempty"`
	Metrics SkillMetrics `json:"metrics"`
}

type SkillMetrics struct {
	Duration   time.Duration `json:"duration"`
	TokenUsage int           `json:"token_usage"`
	RetryCount int           `json:"retry_count"`
}

type SkillBridge interface {
	InvokeGStack(ctx context.Context, skill string, input interface{}) (*SkillResult, error)
	InvokeSuperpower(ctx context.Context, skill string, input interface{}) (*SkillResult, error)
	Chain(ctx context.Context, calls []SkillCall) ([]SkillResult, error)
	InvokeWorkflow(ctx context.Context, workflowName string, input interface{}) ([]SkillResult, error)
}

type BridgeConfig struct {
	DefaultTimeout   time.Duration `json:"default_timeout"`
	MaxRetries       int           `json:"max_retries"`
	EnableMetrics    bool          `json:"enable_metrics"`
	EnableCache      bool          `json:"enable_cache"`
	CacheTTL         time.Duration `json:"cache_ttl"`
}

type DefaultSkillBridge struct {
	config    BridgeConfig
	cache     map[string]*cachedResult
	cacheMu   sync.RWMutex
	metrics   map[string]*SkillMetrics
	metricsMu sync.RWMutex
}

type cachedResult struct {
	result    *SkillResult
	expiresAt time.Time
}

func NewSkillBridge(config BridgeConfig) *DefaultSkillBridge {
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = 5 * time.Minute
	}
	
	return &DefaultSkillBridge{
		config:  config,
		cache:   make(map[string]*cachedResult),
		metrics: make(map[string]*SkillMetrics),
	}
}

func (b *DefaultSkillBridge) InvokeGStack(ctx context.Context, skill string, input interface{}) (*SkillResult, error) {
	call := SkillCall{
		Type:    SkillTypeGStack,
		Name:    skill,
		Input:   input,
		Timeout: b.config.DefaultTimeout,
	}
	return b.invoke(ctx, call)
}

func (b *DefaultSkillBridge) InvokeSuperpower(ctx context.Context, skill string, input interface{}) (*SkillResult, error) {
	call := SkillCall{
		Type:    SkillTypeSuperpower,
		Name:    skill,
		Input:   input,
		Timeout: b.config.DefaultTimeout,
	}
	return b.invoke(ctx, call)
}

func (b *DefaultSkillBridge) Chain(ctx context.Context, calls []SkillCall) ([]SkillResult, error) {
	results := make([]SkillResult, len(calls))
	
	for i, call := range calls {
		result, err := b.invoke(ctx, call)
		if err != nil {
			return results, fmt.Errorf("skill %s failed: %w", call.Name, err)
		}
		results[i] = *result
	}
	
	return results, nil
}

func (b *DefaultSkillBridge) InvokeWorkflow(ctx context.Context, workflowName string, input interface{}) ([]SkillResult, error) {
	workflow, err := b.loadWorkflow(workflowName)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow: %w", err)
	}
	
	return b.executeWorkflow(ctx, workflow, input)
}

func (b *DefaultSkillBridge) invoke(ctx context.Context, call SkillCall) (*SkillResult, error) {
	cacheKey := b.getCacheKey(call)
	
	if b.config.EnableCache {
		if cached := b.getFromCache(cacheKey); cached != nil {
			return cached, nil
		}
	}
	
	var result *SkillResult
	var err error
	
	for i := 0; i <= b.config.MaxRetries; i++ {
		result, err = b.executeSkill(ctx, call)
		if err == nil && result.Success {
			break
		}
		if i < b.config.MaxRetries {
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}
	}
	
	if err != nil {
		return nil, err
	}
	
	if b.config.EnableCache && result.Success {
		b.setToCache(cacheKey, result)
	}
	
	if b.config.EnableMetrics {
		b.recordMetrics(call.Name, result.Metrics)
	}
	
	return result, nil
}

func (b *DefaultSkillBridge) executeSkill(ctx context.Context, call SkillCall) (*SkillResult, error) {
	start := time.Now()
	
	if call.Timeout == 0 {
		call.Timeout = b.config.DefaultTimeout
	}
	
	ctx, cancel := context.WithTimeout(ctx, call.Timeout)
	defer cancel()
	
	var result *SkillResult
	var err error
	
	switch call.Type {
	case SkillTypeGStack:
		result, err = b.executeGStackSkill(ctx, call)
	case SkillTypeSuperpower:
		result, err = b.executeSuperpowerSkill(ctx, call)
	default:
		return nil, fmt.Errorf("unknown skill type: %s", call.Type)
	}
	
	if result != nil {
		result.Metrics.Duration = time.Since(start)
	}
	
	return result, err
}

func (b *DefaultSkillBridge) executeGStackSkill(ctx context.Context, call SkillCall) (*SkillResult, error) {
	inputJSON, err := json.Marshal(call.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}
	
	_ = inputJSON
	
	return &SkillResult{
		Success: true,
		Output: map[string]interface{}{
			"skill":  call.Name,
			"status": "executed",
		},
		Metrics: SkillMetrics{},
	}, nil
}

func (b *DefaultSkillBridge) executeSuperpowerSkill(ctx context.Context, call SkillCall) (*SkillResult, error) {
	inputJSON, err := json.Marshal(call.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}
	
	_ = inputJSON
	
	return &SkillResult{
		Success: true,
		Output: map[string]interface{}{
			"skill":  call.Name,
			"status": "executed",
		},
		Metrics: SkillMetrics{},
	}, nil
}

func (b *DefaultSkillBridge) loadWorkflow(name string) (*Workflow, error) {
	return &Workflow{
		Name: name,
		Steps: []WorkflowStep{
			{ID: "step1", Skill: "brainstorming"},
			{ID: "step2", Skill: "writing-plans"},
		},
	}, nil
}

func (b *DefaultSkillBridge) executeWorkflow(ctx context.Context, workflow *Workflow, input interface{}) ([]SkillResult, error) {
	results := make([]SkillResult, len(workflow.Steps))
	
	for i, step := range workflow.Steps {
		call := SkillCall{
			Type: SkillTypeSuperpower,
			Name: step.Skill,
			Input: map[string]interface{}{
				"workflow": workflow.Name,
				"step":     step.ID,
				"data":     input,
			},
		}
		
		result, err := b.invoke(ctx, call)
		if err != nil {
			return results, fmt.Errorf("workflow step %s failed: %w", step.ID, err)
		}
		
		results[i] = *result
		input = result.Output
	}
	
	return results, nil
}

func (b *DefaultSkillBridge) getCacheKey(call SkillCall) string {
	inputJSON, _ := json.Marshal(call.Input)
	return fmt.Sprintf("%s:%s:%s", call.Type, call.Name, string(inputJSON))
}

func (b *DefaultSkillBridge) getFromCache(key string) *SkillResult {
	b.cacheMu.RLock()
	defer b.cacheMu.RUnlock()
	
	if cached, ok := b.cache[key]; ok {
		if time.Now().Before(cached.expiresAt) {
			return cached.result
		}
	}
	return nil
}

func (b *DefaultSkillBridge) setToCache(key string, result *SkillResult) {
	b.cacheMu.Lock()
	defer b.cacheMu.Unlock()
	
	b.cache[key] = &cachedResult{
		result:    result,
		expiresAt: time.Now().Add(b.config.CacheTTL),
	}
}

func (b *DefaultSkillBridge) recordMetrics(skill string, metrics SkillMetrics) {
	b.metricsMu.Lock()
	defer b.metricsMu.Unlock()
	
	b.metrics[skill] = &metrics
}

type Workflow struct {
	Name  string         `json:"name"`
	Steps []WorkflowStep `json:"steps"`
}

type WorkflowStep struct {
	ID        string `json:"id"`
	Skill     string `json:"skill"`
	DependsOn string `json:"depends_on,omitempty"`
}
