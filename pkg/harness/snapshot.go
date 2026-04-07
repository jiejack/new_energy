package harness

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrSnapshotNotFound 快照未找到错误
var ErrSnapshotNotFound = errors.New("snapshot not found")

// Snapshot 快照测试结果
type Snapshot struct {
	ID        string `json:"id"`
	Data      []byte `json:"data"`
	CreatedAt int64  `json:"created_at"`
	Checksum  string `json:"checksum"`
}

// Reset 重置 Snapshot 对象以便重用
func (s *Snapshot) Reset() {
	s.ID = ""
	s.Data = nil
	s.CreatedAt = 0
	s.Checksum = ""
}

// SnapshotManager 快照管理器
type SnapshotManager struct {
	storage map[string]*Snapshot
	mu      sync.RWMutex

	// 对象池
	snapshotPool sync.Pool

	// checksum 计算缓冲区池
	checksumBufPool sync.Pool
}

// NewSnapshotManager 创建新的快照管理器
func NewSnapshotManager() *SnapshotManager {
	return &SnapshotManager{
		storage: make(map[string]*Snapshot),
		snapshotPool: sync.Pool{
			New: func() interface{} {
				return &Snapshot{}
			},
		},
		checksumBufPool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, sha256.Size)
				return buf
			},
		},
	}
}

// Save 保存快照
func (sm *SnapshotManager) Save(id string, data []byte) error {
	if id == "" {
		return errors.New("snapshot id cannot be empty")
	}

	if data == nil {
		return errors.New("snapshot data cannot be nil")
	}

	checksum := sm.calculateChecksum(data)

	// 从对象池获取 Snapshot
	snapshot := sm.snapshotPool.Get().(*Snapshot)
	snapshot.ID = id
	snapshot.Data = data
	snapshot.CreatedAt = time.Now().Unix()
	snapshot.Checksum = checksum

	sm.mu.Lock()
	sm.storage[id] = snapshot
	sm.mu.Unlock()
	return nil
}

// Load 加载快照
func (sm *SnapshotManager) Load(id string) (*Snapshot, error) {
	if id == "" {
		return nil, errors.New("snapshot id cannot be empty")
	}

	sm.mu.RLock()
	snapshot, exists := sm.storage[id]
	sm.mu.RUnlock()

	if !exists {
		return nil, ErrSnapshotNotFound
	}

	return snapshot, nil
}

// Delete 删除快照
func (sm *SnapshotManager) Delete(id string) error {
	if id == "" {
		return errors.New("snapshot id cannot be empty")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	snapshot, exists := sm.storage[id]
	if !exists {
		return ErrSnapshotNotFound
	}

	// 将 Snapshot 放回对象池
	snapshot.Reset()
	sm.snapshotPool.Put(snapshot)

	delete(sm.storage, id)
	return nil
}

// List 列出所有快照ID
func (sm *SnapshotManager) List() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ids := make([]string, 0, len(sm.storage))
	for id := range sm.storage {
		ids = append(ids, id)
	}
	return ids
}

// Compare 比较快照数据
func (sm *SnapshotManager) Compare(id string, data []byte) (bool, error) {
	snapshot, err := sm.Load(id)
	if err != nil {
		return false, err
	}

	newChecksum := sm.calculateChecksum(data)
	return snapshot.Checksum == newChecksum, nil
}

// calculateChecksum 计算数据校验和
func (sm *SnapshotManager) calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// ToJSON 将快照序列化为JSON
func (s *Snapshot) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// FromJSON 从JSON反序列化快照
func (s *Snapshot) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}

// String 返回快照的字符串表示
func (s *Snapshot) String() string {
	return fmt.Sprintf("Snapshot{ID: %s, CreatedAt: %d, Checksum: %s, DataSize: %d}",
		s.ID, s.CreatedAt, s.Checksum, len(s.Data))
}
