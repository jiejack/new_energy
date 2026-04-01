package helpers

import (
	"context"
	"sync"
	"time"
)

// MockRedis 模拟Redis客户端
type MockRedis struct {
	mu      sync.RWMutex
	data    map[string]string
	expiry  map[string]time.Time
	lists   map[string][]string
	sets    map[string]map[string]struct{}
	hashes  map[string]map[string]string
}

// NewMockRedis 创建模拟Redis客户端
func NewMockRedis() *MockRedis {
	return &MockRedis{
		data:   make(map[string]string),
		expiry: make(map[string]time.Time),
		lists:  make(map[string][]string),
		sets:   make(map[string]map[string]struct{}),
		hashes: make(map[string]map[string]string),
	}
}

// Set 设置键值
func (m *MockRedis) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	if expiration > 0 {
		m.expiry[key] = time.Now().Add(expiration)
	} else {
		delete(m.expiry, key)
	}
	return nil
}

// Get 获取键值
func (m *MockRedis) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// 检查是否过期
	if exp, ok := m.expiry[key]; ok && time.Now().After(exp) {
		return "", ErrKeyNotFound
	}
	
	val, ok := m.data[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	return val, nil
}

// Del 删除键
func (m *MockRedis) Del(ctx context.Context, keys ...string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var count int64
	for _, key := range keys {
		if _, ok := m.data[key]; ok {
			delete(m.data, key)
			delete(m.expiry, key)
			delete(m.lists, key)
			delete(m.sets, key)
			delete(m.hashes, key)
			count++
		}
	}
	return count, nil
}

// Exists 检查键是否存在
func (m *MockRedis) Exists(ctx context.Context, keys ...string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var count int64
	for _, key := range keys {
		if _, ok := m.data[key]; ok {
			// 检查是否过期
			if exp, expOk := m.expiry[key]; expOk && time.Now().After(exp) {
				continue
			}
			count++
		}
	}
	return count, nil
}

// Expire 设置过期时间
func (m *MockRedis) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, ok := m.data[key]; !ok {
		return false, nil
	}
	m.expiry[key] = time.Now().Add(expiration)
	return true, nil
}

// TTL 获取剩余过期时间
func (m *MockRedis) TTL(ctx context.Context, key string) (time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if _, ok := m.data[key]; !ok {
		return -2 * time.Second, nil
	}
	
	if exp, ok := m.expiry[key]; ok {
		ttl := time.Until(exp)
		if ttl < 0 {
			return -2 * time.Second, nil
		}
		return ttl, nil
	}
	
	return -1 * time.Second, nil
}

// Incr 自增
func (m *MockRedis) Incr(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	val, ok := m.data[key]
	if !ok {
		m.data[key] = "1"
		return 1, nil
	}
	
	var intVal int64 = 1
	if val != "" {
		_, err := time.ParseDuration(val)
		if err == nil {
			intVal = 1
		}
	}
	
	m.data[key] = val + string(rune(intVal))
	return intVal, nil
}

// LPush 列表左侧插入
func (m *MockRedis) LPush(ctx context.Context, key string, values ...string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.lists[key] == nil {
		m.lists[key] = make([]string, 0)
	}
	
	m.lists[key] = append(values, m.lists[key]...)
	return int64(len(m.lists[key])), nil
}

// RPush 列表右侧插入
func (m *MockRedis) RPush(ctx context.Context, key string, values ...string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.lists[key] == nil {
		m.lists[key] = make([]string, 0)
	}
	
	m.lists[key] = append(m.lists[key], values...)
	return int64(len(m.lists[key])), nil
}

// LRange 获取列表范围
func (m *MockRedis) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	list, ok := m.lists[key]
	if !ok {
		return []string{}, nil
	}
	
	length := int64(len(list))
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}
	
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	
	if start > stop || start >= length {
		return []string{}, nil
	}
	
	return list[start : stop+1], nil
}

// SAdd 集合添加
func (m *MockRedis) SAdd(ctx context.Context, key string, members ...string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.sets[key] == nil {
		m.sets[key] = make(map[string]struct{})
	}
	
	var count int64
	for _, member := range members {
		if _, ok := m.sets[key][member]; !ok {
			m.sets[key][member] = struct{}{}
			count++
		}
	}
	return count, nil
}

// SMembers 获取集合所有成员
func (m *MockRedis) SMembers(ctx context.Context, key string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	set, ok := m.sets[key]
	if !ok {
		return []string{}, nil
	}
	
	members := make([]string, 0, len(set))
	for member := range set {
		members = append(members, member)
	}
	return members, nil
}

// SIsMember 检查是否是集合成员
func (m *MockRedis) SIsMember(ctx context.Context, key, member string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	set, ok := m.sets[key]
	if !ok {
		return false, nil
	}
	
	_, exists := set[member]
	return exists, nil
}

// HSet 哈希设置
func (m *MockRedis) HSet(ctx context.Context, key string, fieldValues ...string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.hashes[key] == nil {
		m.hashes[key] = make(map[string]string)
	}
	
	var count int64
	for i := 0; i < len(fieldValues); i += 2 {
		if i+1 < len(fieldValues) {
			field := fieldValues[i]
			value := fieldValues[i+1]
			if _, ok := m.hashes[key][field]; !ok {
				count++
			}
			m.hashes[key][field] = value
		}
	}
	return count, nil
}

// HGet 哈希获取
func (m *MockRedis) HGet(ctx context.Context, key, field string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	hash, ok := m.hashes[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	
	val, ok := hash[field]
	if !ok {
		return "", ErrKeyNotFound
	}
	return val, nil
}

// HGetAll 哈希获取所有
func (m *MockRedis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	hash, ok := m.hashes[key]
	if !ok {
		return map[string]string{}, nil
	}
	
	result := make(map[string]string)
	for k, v := range hash {
		result[k] = v
	}
	return result, nil
}

// Clear 清空所有数据
func (m *MockRedis) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]string)
	m.expiry = make(map[string]time.Time)
	m.lists = make(map[string][]string)
	m.sets = make(map[string]map[string]struct{})
	m.hashes = make(map[string]map[string]string)
}

// 错误定义
var ErrKeyNotFound = &KeyNotFoundError{}

// KeyNotFoundError 键不存在错误
type KeyNotFoundError struct{}

func (e *KeyNotFoundError) Error() string {
	return "key not found"
}
