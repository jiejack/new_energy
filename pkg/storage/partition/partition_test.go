package partition

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPartitioner(t *testing.T) {
	config := PartitionConfig{
		Strategy:       "time",
		PartitionSize:  24 * time.Hour,
		MaxPartitions:  30,
	}

	partitioner, err := NewPartitioner(config)
	assert.NoError(t, err)
	assert.NotNil(t, partitioner)
}

func TestPartitioner_GetPartition(t *testing.T) {
	config := PartitionConfig{
		Strategy:      "time",
		PartitionSize: 24 * time.Hour,
		MaxPartitions: 30,
	}

	partitioner, err := NewPartitioner(config)
	assert.NoError(t, err)

	now := time.Now()
	partition := partitioner.GetPartition(now)

	assert.NotEmpty(t, partition)
}

func TestPartitioner_GetPartitionForPoint(t *testing.T) {
	config := PartitionConfig{
		Strategy:      "time",
		PartitionSize: 24 * time.Hour,
		MaxPartitions: 30,
	}

	partitioner, err := NewPartitioner(config)
	assert.NoError(t, err)

	point := &PartitionPoint{
		PointID:   "point001",
		Timestamp: time.Now(),
		Value:     100.5,
	}

	partition := partitioner.GetPartitionForPoint(point)
	assert.NotEmpty(t, partition)
}

func TestPartitioner_GetPartitionsInRange(t *testing.T) {
	config := PartitionConfig{
		Strategy:      "time",
		PartitionSize: 24 * time.Hour,
		MaxPartitions: 30,
	}

	partitioner, err := NewPartitioner(config)
	assert.NoError(t, err)

	now := time.Now()
	start := now.Add(-5 * 24 * time.Hour)
	end := now

	partitions := partitioner.GetPartitionsInRange(start, end)
	assert.GreaterOrEqual(t, len(partitions), 5)
	assert.LessOrEqual(t, len(partitions), 7) // 考虑边界
}

func TestPartitioner_CreatePartition(t *testing.T) {
	config := PartitionConfig{
		Strategy:      "time",
		PartitionSize: 24 * time.Hour,
		MaxPartitions: 30,
	}

	partitioner, err := NewPartitioner(config)
	assert.NoError(t, err)

	name := "partition_20240101"
	err = partitioner.CreatePartition(name)
	assert.NoError(t, err)
}

func TestPartitioner_DropPartition(t *testing.T) {
	config := PartitionConfig{
		Strategy:      "time",
		PartitionSize: 24 * time.Hour,
		MaxPartitions: 30,
	}

	partitioner, err := NewPartitioner(config)
	assert.NoError(t, err)

	name := "partition_20240101"
	partitioner.CreatePartition(name)

	err = partitioner.DropPartition(name)
	assert.NoError(t, err)
}

func TestPartitioner_ListPartitions(t *testing.T) {
	config := PartitionConfig{
		Strategy:      "time",
		PartitionSize: 24 * time.Hour,
		MaxPartitions: 30,
	}

	partitioner, err := NewPartitioner(config)
	assert.NoError(t, err)

	// 创建几个分区
	for i := 0; i < 3; i++ {
		name := "partition_" + string(rune('a'+i))
		partitioner.CreatePartition(name)
	}

	partitions := partitioner.ListPartitions()
	assert.GreaterOrEqual(t, len(partitions), 3)
}

func TestTimeBasedPartitioner(t *testing.T) {
	partitioner := NewTimeBasedPartitioner(24 * time.Hour)

	now := time.Now()
	tests := []struct {
		name      string
		timestamp time.Time
		expected  string
	}{
		{
			name:      "当前时间",
			timestamp: now,
			expected:  now.Format("20060102"),
		},
		{
			name:      "昨天",
			timestamp: now.Add(-24 * time.Hour),
			expected:  now.Add(-24 * time.Hour).Format("20060102"),
		},
		{
			name:      "明天",
			timestamp: now.Add(24 * time.Hour),
			expected:  now.Add(24 * time.Hour).Format("20060102"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partition := partitioner.GetPartition(tt.timestamp)
			assert.Contains(t, partition, tt.expected)
		})
	}
}

func TestHashBasedPartitioner(t *testing.T) {
	partitioner := NewHashBasedPartitioner(10)

	tests := []struct {
		key      string
		expected int
	}{
		{"point001", partitioner.GetPartitionIndex("point001")},
		{"point002", partitioner.GetPartitionIndex("point002")},
		{"point003", partitioner.GetPartitionIndex("point003")},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			index := partitioner.GetPartitionIndex(tt.key)
			assert.GreaterOrEqual(t, index, 0)
			assert.Less(t, index, 10)
			assert.Equal(t, tt.expected, index)
		})
	}
}

func TestPartitionManager(t *testing.T) {
	config := PartitionConfig{
		Strategy:      "time",
		PartitionSize: 24 * time.Hour,
		MaxPartitions: 30,
	}

	manager := NewPartitionManager(config)

	assert.NotNil(t, manager)
}

func TestPartitionManager_Rotate(t *testing.T) {
	config := PartitionConfig{
		Strategy:      "time",
		PartitionSize: 24 * time.Hour,
		MaxPartitions: 5, // 最多保留5个分区
	}

	manager := NewPartitionManager(config)

	// 创建超过最大数量的分区
	for i := 0; i < 10; i++ {
		name := "partition_" + string(rune('0'+i))
		manager.CreatePartition(name)
	}

	// 执行轮转
	err := manager.Rotate()
	assert.NoError(t, err)

	// 应该只保留最新的分区
	partitions := manager.ListPartitions()
	assert.LessOrEqual(t, len(partitions), 5)
}

func TestPartitionManager_GetActivePartition(t *testing.T) {
	config := PartitionConfig{
		Strategy:      "time",
		PartitionSize: 24 * time.Hour,
		MaxPartitions: 30,
	}

	manager := NewPartitionManager(config)

	active := manager.GetActivePartition()
	assert.NotEmpty(t, active)
}

func TestPartitionManager_GetPartitionStats(t *testing.T) {
	config := PartitionConfig{
		Strategy:      "time",
		PartitionSize: 24 * time.Hour,
		MaxPartitions: 30,
	}

	manager := NewPartitionManager(config)

	name := "partition_test"
	manager.CreatePartition(name)

	stats := manager.GetPartitionStats(name)
	assert.NotNil(t, stats)
}
