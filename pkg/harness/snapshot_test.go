package harness

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNewSnapshotManager(t *testing.T) {
	sm := NewSnapshotManager()
	if sm == nil {
		t.Fatal("NewSnapshotManager() returned nil")
	}
	if sm.storage == nil {
		t.Error("SnapshotManager.storage is nil")
	}
}

func TestSnapshotManager_Save(t *testing.T) {
	sm := NewSnapshotManager()

	tests := []struct {
		name    string
		id      string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid snapshot",
			id:      "test-snapshot-1",
			data:    []byte("test data"),
			wantErr: false,
		},
		{
			name:    "empty id",
			id:      "",
			data:    []byte("test data"),
			wantErr: true,
		},
		{
			name:    "nil data",
			id:      "test-snapshot-2",
			data:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sm.Save(tt.id, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("SnapshotManager.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSnapshotManager_Load(t *testing.T) {
	sm := NewSnapshotManager()

	// 先保存一个快照
	testID := "test-load-snapshot"
	testData := []byte("test data for load")
	err := sm.Save(testID, testData)
	if err != nil {
		t.Fatalf("Failed to save snapshot: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "load existing snapshot",
			id:      testID,
			wantErr: false,
		},
		{
			name:    "load non-existing snapshot",
			id:      "non-existing-id",
			wantErr: true,
		},
		{
			name:    "empty id",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot, err := sm.Load(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SnapshotManager.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if snapshot == nil {
					t.Error("SnapshotManager.Load() returned nil snapshot")
					return
				}
				if snapshot.ID != tt.id {
					t.Errorf("Snapshot.ID = %v, want %v", snapshot.ID, tt.id)
				}
				if string(snapshot.Data) != string(testData) {
					t.Errorf("Snapshot.Data = %v, want %v", string(snapshot.Data), string(testData))
				}
				if snapshot.Checksum == "" {
					t.Error("Snapshot.Checksum is empty")
				}
				if snapshot.CreatedAt == 0 {
					t.Error("Snapshot.CreatedAt is zero")
				}
			}
		})
	}
}

func TestSnapshotManager_Delete(t *testing.T) {
	sm := NewSnapshotManager()

	// 先保存一个快照
	testID := "test-delete-snapshot"
	err := sm.Save(testID, []byte("test data"))
	if err != nil {
		t.Fatalf("Failed to save snapshot: %v", err)
	}

	// 删除存在的快照
	err = sm.Delete(testID)
	if err != nil {
		t.Errorf("SnapshotManager.Delete() error = %v", err)
	}

	// 验证快照已被删除
	_, err = sm.Load(testID)
	if err != ErrSnapshotNotFound {
		t.Errorf("Expected ErrSnapshotNotFound, got %v", err)
	}

	// 删除不存在的快照
	err = sm.Delete("non-existing-id")
	if err != ErrSnapshotNotFound {
		t.Errorf("Expected ErrSnapshotNotFound, got %v", err)
	}

	// 删除空ID
	err = sm.Delete("")
	if err == nil {
		t.Error("SnapshotManager.Delete() should return error for empty id")
	}
}

func TestSnapshotManager_List(t *testing.T) {
	sm := NewSnapshotManager()

	// 空管理器
	ids := sm.List()
	if len(ids) != 0 {
		t.Errorf("Empty SnapshotManager.List() = %v, want empty", ids)
	}

	// 添加多个快照
	testIDs := []string{"snapshot-1", "snapshot-2", "snapshot-3"}
	for _, id := range testIDs {
		err := sm.Save(id, []byte("data"))
		if err != nil {
			t.Fatalf("Failed to save snapshot %s: %v", id, err)
		}
	}

	// 验证列表
	ids = sm.List()
	if len(ids) != len(testIDs) {
		t.Errorf("SnapshotManager.List() length = %d, want %d", len(ids), len(testIDs))
	}

	// 验证所有ID都在列表中
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}
	for _, testID := range testIDs {
		if !idMap[testID] {
			t.Errorf("Snapshot ID %s not found in list", testID)
		}
	}
}

func TestSnapshotManager_Compare(t *testing.T) {
	sm := NewSnapshotManager()

	// 保存原始快照
	testID := "test-compare-snapshot"
	originalData := []byte("original data")
	err := sm.Save(testID, originalData)
	if err != nil {
		t.Fatalf("Failed to save snapshot: %v", err)
	}

	// 比较相同数据
	match, err := sm.Compare(testID, originalData)
	if err != nil {
		t.Errorf("SnapshotManager.Compare() error = %v", err)
	}
	if !match {
		t.Error("SnapshotManager.Compare() = false for same data, want true")
	}

	// 比较不同数据
	differentData := []byte("different data")
	match, err = sm.Compare(testID, differentData)
	if err != nil {
		t.Errorf("SnapshotManager.Compare() error = %v", err)
	}
	if match {
		t.Error("SnapshotManager.Compare() = true for different data, want false")
	}

	// 比较不存在的快照
	_, err = sm.Compare("non-existing-id", originalData)
	if err != ErrSnapshotNotFound {
		t.Errorf("Expected ErrSnapshotNotFound, got %v", err)
	}
}

func TestSnapshotManager_calculateChecksum(t *testing.T) {
	sm := NewSnapshotManager()

	// 相同数据应该产生相同的校验和
	data1 := []byte("test data")
	data2 := []byte("test data")
	checksum1 := sm.calculateChecksum(data1)
	checksum2 := sm.calculateChecksum(data2)

	if checksum1 != checksum2 {
		t.Errorf("Checksums for same data differ: %s vs %s", checksum1, checksum2)
	}

	// 不同数据应该产生不同的校验和
	data3 := []byte("different data")
	checksum3 := sm.calculateChecksum(data3)

	if checksum1 == checksum3 {
		t.Error("Checksums for different data are the same")
	}

	// 校验和应该是64个字符（SHA256的十六进制表示）
	if len(checksum1) != 64 {
		t.Errorf("Checksum length = %d, want 64", len(checksum1))
	}
}

func TestSnapshot_ToJSON(t *testing.T) {
	snapshot := &Snapshot{
		ID:        "test-json-snapshot",
		Data:      []byte("test data"),
		CreatedAt: 1234567890,
		Checksum:  "abc123",
	}

	jsonData, err := snapshot.ToJSON()
	if err != nil {
		t.Errorf("Snapshot.ToJSON() error = %v", err)
	}

	// 验证JSON是否包含所有字段
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	if err != nil {
		t.Errorf("Failed to parse JSON: %v", err)
	}

	if parsed["id"] != snapshot.ID {
		t.Errorf("JSON id = %v, want %v", parsed["id"], snapshot.ID)
	}
	if parsed["created_at"].(float64) != float64(snapshot.CreatedAt) {
		t.Errorf("JSON created_at = %v, want %v", parsed["created_at"], snapshot.CreatedAt)
	}
	if parsed["checksum"] != snapshot.Checksum {
		t.Errorf("JSON checksum = %v, want %v", parsed["checksum"], snapshot.Checksum)
	}
}

func TestSnapshot_FromJSON(t *testing.T) {
	jsonStr := `{"id":"test-json-snapshot","data":"dGVzdCBkYXRh","created_at":1234567890,"checksum":"abc123"}`

	snapshot := &Snapshot{}
	err := snapshot.FromJSON([]byte(jsonStr))
	if err != nil {
		t.Errorf("Snapshot.FromJSON() error = %v", err)
	}

	if snapshot.ID != "test-json-snapshot" {
		t.Errorf("Snapshot.ID = %v, want test-json-snapshot", snapshot.ID)
	}
	if string(snapshot.Data) != "test data" {
		t.Errorf("Snapshot.Data = %v, want test data", string(snapshot.Data))
	}
	if snapshot.CreatedAt != 1234567890 {
		t.Errorf("Snapshot.CreatedAt = %v, want 1234567890", snapshot.CreatedAt)
	}
	if snapshot.Checksum != "abc123" {
		t.Errorf("Snapshot.Checksum = %v, want abc123", snapshot.Checksum)
	}
}

func TestSnapshot_String(t *testing.T) {
	snapshot := &Snapshot{
		ID:        "test-string-snapshot",
		Data:      []byte("test data"),
		CreatedAt: 1234567890,
		Checksum:  "abc123",
	}

	str := snapshot.String()
	if str == "" {
		t.Error("Snapshot.String() returned empty string")
	}

	// 验证字符串包含关键字段
	if !contains(str, "test-string-snapshot") {
		t.Error("Snapshot.String() does not contain ID")
	}
	if !contains(str, "1234567890") {
		t.Error("Snapshot.String() does not contain CreatedAt")
	}
	if !contains(str, "abc123") {
		t.Error("Snapshot.String() does not contain Checksum")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// 基准测试
func BenchmarkSnapshotManager_Save(b *testing.B) {
	sm := NewSnapshotManager()
	data := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sm.Save("benchmark-snapshot", data)
	}
}

func BenchmarkSnapshotManager_SaveWithPool(b *testing.B) {
	sm := NewSnapshotManager()
	data := []byte("benchmark test data")

	// 预热对象池
	for i := 0; i < 100; i++ {
		_ = sm.Save("warmup-snapshot", data)
		_ = sm.Delete("warmup-snapshot")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sm.Save("benchmark-snapshot", data)
	}
}

func BenchmarkSnapshotManager_SaveParallel(b *testing.B) {
	sm := NewSnapshotManager()
	data := []byte("benchmark test data")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = sm.Save(fmt.Sprintf("benchmark-snapshot-%d", i), data)
			i++
		}
	})
}

func BenchmarkSnapshotManager_Load(b *testing.B) {
	sm := NewSnapshotManager()
	data := []byte("benchmark test data")
	_ = sm.Save("benchmark-snapshot", data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sm.Load("benchmark-snapshot")
	}
}

func BenchmarkSnapshotManager_LoadParallel(b *testing.B) {
	sm := NewSnapshotManager()
	data := []byte("benchmark test data")
	_ = sm.Save("benchmark-snapshot", data)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = sm.Load("benchmark-snapshot")
		}
	})
}

func BenchmarkSnapshotManager_SaveDelete(b *testing.B) {
	sm := NewSnapshotManager()
	data := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sm.Save("benchmark-snapshot", data)
		_ = sm.Delete("benchmark-snapshot")
	}
}

func BenchmarkSnapshotManager_SaveDeleteParallel(b *testing.B) {
	sm := NewSnapshotManager()
	data := []byte("benchmark test data")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			id := fmt.Sprintf("benchmark-snapshot-%d", i)
			_ = sm.Save(id, data)
			_ = sm.Delete(id)
			i++
		}
	})
}

func BenchmarkSnapshotManager_Compare(b *testing.B) {
	sm := NewSnapshotManager()
	data := []byte("benchmark test data")
	_ = sm.Save("benchmark-snapshot", data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sm.Compare("benchmark-snapshot", data)
	}
}

func BenchmarkCalculateChecksum(b *testing.B) {
	sm := NewSnapshotManager()
	data := []byte("benchmark test data for checksum calculation")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sm.calculateChecksum(data)
	}
}
