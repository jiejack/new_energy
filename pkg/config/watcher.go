package config

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Watcher 配置变更监听器
type Watcher struct {
	mu sync.RWMutex

	viper     *viper.Viper
	callbacks map[string][]func(key string, value interface{})
	
	running   bool
	stopChan  chan struct{}
	
	// 配置变更检测间隔
	interval  time.Duration
	
	// 上一次的配置值缓存
	lastValues map[string]interface{}
}

// NewWatcher 创建配置监听器
func NewWatcher(v *viper.Viper, callbacks map[string][]func(key string, value interface{})) *Watcher {
	return &Watcher{
		viper:      v,
		callbacks:  callbacks,
		stopChan:   make(chan struct{}),
		interval:   5 * time.Second, // 默认5秒检测一次
		lastValues: make(map[string]interface{}),
	}
}

// Start 启动监听器
func (w *Watcher) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return nil
	}

	w.running = true
	w.stopChan = make(chan struct{})

	// 初始化配置值缓存
	for key := range w.callbacks {
		w.lastValues[key] = w.viper.Get(key)
	}

	// 启动监听goroutine
	go w.watch()

	// 同时启动viper内置的配置文件监听
	w.viper.WatchConfig()
	w.viper.OnConfigChange(func(e fsnotify.Event) {
		w.handleConfigChange()
	})

	return nil
}

// Stop 停止监听器
func (w *Watcher) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return nil
	}

	w.running = false
	close(w.stopChan)

	return nil
}

// watch 监听配置变更
func (w *Watcher) watch() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.checkConfigChanges()
		}
	}
}

// checkConfigChanges 检查配置变更
func (w *Watcher) checkConfigChanges() {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for key, callbacks := range w.callbacks {
		currentValue := w.viper.Get(key)
		lastValue, exists := w.lastValues[key]

		// 如果值发生变化
		if !exists || !w.compareValues(lastValue, currentValue) {
			// 更新缓存
			w.lastValues[key] = currentValue

			// 触发回调
			for _, callback := range callbacks {
				go func(cb func(string, interface{}), k string, v interface{}) {
					defer func() {
						if r := recover(); r != nil {
							// 防止回调panic影响其他回调
						}
					}()
					cb(k, v)
				}(callback, key, currentValue)
			}
		}
	}
}

// handleConfigChange 处理配置文件变更
func (w *Watcher) handleConfigChange() {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// 重新读取所有监听的配置
	for key, callbacks := range w.callbacks {
		currentValue := w.viper.Get(key)
		w.lastValues[key] = currentValue

		// 触发所有回调
		for _, callback := range callbacks {
			go func(cb func(string, interface{}), k string, v interface{}) {
				defer func() {
					if r := recover(); r != nil {
						// 防止回调panic影响其他回调
					}
				}()
				cb(k, v)
			}(callback, key, currentValue)
		}
	}
}

// compareValues 比较两个配置值是否相等
func (w *Watcher) compareValues(v1, v2 interface{}) bool {
	// 简单类型比较
	switch v1 := v1.(type) {
	case string:
		if v2, ok := v2.(string); ok {
			return v1 == v2
		}
	case int:
		if v2, ok := v2.(int); ok {
			return v1 == v2
		}
	case int64:
		if v2, ok := v2.(int64); ok {
			return v1 == v2
		}
	case float64:
		if v2, ok := v2.(float64); ok {
			return v1 == v2
		}
	case bool:
		if v2, ok := v2.(bool); ok {
			return v1 == v2
		}
	case []string:
		if v2, ok := v2.([]string); ok {
			if len(v1) != len(v2) {
				return false
			}
			for i := range v1 {
				if v1[i] != v2[i] {
					return false
				}
			}
			return true
		}
	case map[string]interface{}:
		if v2, ok := v2.(map[string]interface{}); ok {
			return w.compareMaps(v1, v2)
		}
	case map[string]string:
		if v2, ok := v2.(map[string]string); ok {
			if len(v1) != len(v2) {
				return false
			}
			for k, val := range v1 {
				if v2[k] != val {
					return false
				}
			}
			return true
		}
	case nil:
		return v2 == nil
	}

	// 默认不相等
	return false
}

// compareMaps 比较两个map是否相等
func (w *Watcher) compareMaps(m1, m2 map[string]interface{}) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		v2, exists := m2[k]
		if !exists || !w.compareValues(v1, v2) {
			return false
		}
	}
	return true
}

// AddCallback 添加配置变更回调
func (w *Watcher) AddCallback(key string, callback func(key string, value interface{})) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.callbacks[key] = append(w.callbacks[key], callback)
	
	// 初始化缓存值
	if _, exists := w.lastValues[key]; !exists {
		w.lastValues[key] = w.viper.Get(key)
	}
}

// RemoveCallback 移除配置变更回调
func (w *Watcher) RemoveCallback(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.callbacks, key)
	delete(w.lastValues, key)
}

// SetInterval 设置检测间隔
func (w *Watcher) SetInterval(interval time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.interval = interval
}

// IsRunning 检查监听器是否运行中
func (w *Watcher) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	Key       string      // 配置键
	OldValue  interface{} // 旧值
	NewValue  interface{} // 新值
	Timestamp time.Time   // 变更时间
}

// ConfigWatcher 高级配置监听器（支持变更事件）
type ConfigWatcher struct {
	mu sync.RWMutex

	viper     *viper.Viper
	listeners map[string][]func(ConfigChangeEvent)
	
	running   bool
	stopChan  chan struct{}
	
	interval  time.Duration
	lastValues map[string]interface{}
}

// NewConfigWatcher 创建高级配置监听器
func NewConfigWatcher(v *viper.Viper) *ConfigWatcher {
	return &ConfigWatcher{
		viper:      v,
		listeners:  make(map[string][]func(ConfigChangeEvent)),
		stopChan:   make(chan struct{}),
		interval:   5 * time.Second,
		lastValues: make(map[string]interface{}),
	}
}

// OnChange 注册配置变更监听器
func (cw *ConfigWatcher) OnChange(key string, listener func(ConfigChangeEvent)) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	cw.listeners[key] = append(cw.listeners[key], listener)
	
	if _, exists := cw.lastValues[key]; !exists {
		cw.lastValues[key] = cw.viper.Get(key)
	}
}

// RemoveListener 移除配置变更监听器
func (cw *ConfigWatcher) RemoveListener(key string) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	delete(cw.listeners, key)
	delete(cw.lastValues, key)
}

// Start 启动监听器
func (cw *ConfigWatcher) Start() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if cw.running {
		return nil
	}

	cw.running = true
	cw.stopChan = make(chan struct{})

	// 初始化配置值缓存
	for key := range cw.listeners {
		cw.lastValues[key] = cw.viper.Get(key)
	}

	// 启动监听goroutine
	go cw.watch()

	return nil
}

// Stop 停止监听器
func (cw *ConfigWatcher) Stop() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if !cw.running {
		return nil
	}

	cw.running = false
	close(cw.stopChan)

	return nil
}

// watch 监听配置变更
func (cw *ConfigWatcher) watch() {
	ticker := time.NewTicker(cw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-cw.stopChan:
			return
		case <-ticker.C:
			cw.checkConfigChanges()
		}
	}
}

// checkConfigChanges 检查配置变更
func (cw *ConfigWatcher) checkConfigChanges() {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	for key, listeners := range cw.listeners {
		currentValue := cw.viper.Get(key)
		lastValue, exists := cw.lastValues[key]

		// 如果值发生变化
		if !exists || !cw.compareValues(lastValue, currentValue) {
			// 创建变更事件
			event := ConfigChangeEvent{
				Key:       key,
				OldValue:  lastValue,
				NewValue:  currentValue,
				Timestamp: time.Now(),
			}

			// 更新缓存
			cw.lastValues[key] = currentValue

			// 触发监听器
			for _, listener := range listeners {
				go func(l func(ConfigChangeEvent), e ConfigChangeEvent) {
					defer func() {
						if r := recover(); r != nil {
							// 防止监听器panic影响其他监听器
						}
					}()
					l(e)
				}(listener, event)
			}
		}
	}
}

// compareValues 比较两个配置值是否相等
func (cw *ConfigWatcher) compareValues(v1, v2 interface{}) bool {
	// 使用与Watcher相同的比较逻辑
	w := &Watcher{}
	return w.compareValues(v1, v2)
}

// IsRunning 检查监听器是否运行中
func (cw *ConfigWatcher) IsRunning() bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.running
}

// SetInterval 设置检测间隔
func (cw *ConfigWatcher) SetInterval(interval time.Duration) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.interval = interval
}
