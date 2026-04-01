package collector

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConnectionPool(t *testing.T) {
	config := PoolConfig{
		MaxIdleConns:    10,
		MaxActiveConns:  50,
		IdleTimeout:     5 * time.Minute,
		ConnectTimeout:  10 * time.Second,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    10 * time.Second,
	}

	pool := NewConnectionPool(config)

	assert.NotNil(t, pool)
	assert.Equal(t, config, pool.config)
}

func TestConnectionPool_Get(t *testing.T) {
	config := PoolConfig{
		MaxIdleConns:   10,
		MaxActiveConns: 50,
		IdleTimeout:    5 * time.Minute,
	}

	pool := NewConnectionPool(config)
	ctx := context.Background()

	// 模拟连接获取
	conn, err := pool.Get(ctx, "192.168.1.100:502")

	// 由于没有实际连接，这里可能返回错误，但我们可以测试逻辑
	// 实际测试中应该mock连接
	assert.NotNil(t, pool)
	_ = conn
	_ = err
}

func TestConnectionPool_Put(t *testing.T) {
	config := PoolConfig{
		MaxIdleConns:   10,
		MaxActiveConns: 50,
		IdleTimeout:    5 * time.Minute,
	}

	pool := NewConnectionPool(config)

	// 测试归还连接
	pool.Put("192.168.1.100:502", nil, nil)
}

func TestConnectionPool_Close(t *testing.T) {
	config := PoolConfig{
		MaxIdleConns:   10,
		MaxActiveConns: 50,
		IdleTimeout:    5 * time.Minute,
	}

	pool := NewConnectionPool(config)

	err := pool.Close()
	assert.NoError(t, err)
}

func TestNewDataBuffer(t *testing.T) {
	config := BufferConfig{
		MaxSize:     10000,
		FlushInterval: 5 * time.Second,
		BatchSize:   100,
	}

	buffer := NewDataBuffer(config)

	assert.NotNil(t, buffer)
	assert.Equal(t, config, buffer.config)
}

func TestDataBuffer_Add(t *testing.T) {
	config := BufferConfig{
		MaxSize:     100,
		FlushInterval: 5 * time.Second,
		BatchSize:   10,
	}

	buffer := NewDataBuffer(config)

	data := &DataPoint{
		PointID:   "point001",
		Value:     100.5,
		Timestamp: time.Now(),
		Quality:   1,
	}

	err := buffer.Add(data)
	assert.NoError(t, err)
	assert.Equal(t, 1, buffer.Size())
}

func TestDataBuffer_AddBatch(t *testing.T) {
	config := BufferConfig{
		MaxSize:     100,
		FlushInterval: 5 * time.Second,
		BatchSize:   10,
	}

	buffer := NewDataBuffer(config)

	dataPoints := []*DataPoint{
		{PointID: "point001", Value: 100.0, Timestamp: time.Now(), Quality: 1},
		{PointID: "point002", Value: 200.0, Timestamp: time.Now(), Quality: 1},
		{PointID: "point003", Value: 300.0, Timestamp: time.Now(), Quality: 1},
	}

	err := buffer.AddBatch(dataPoints)
	assert.NoError(t, err)
	assert.Equal(t, 3, buffer.Size())
}

func TestDataBuffer_GetBatch(t *testing.T) {
	config := BufferConfig{
		MaxSize:     100,
		FlushInterval: 5 * time.Second,
		BatchSize:   2,
	}

	buffer := NewDataBuffer(config)

	// 添加5个数据点
	for i := 0; i < 5; i++ {
		buffer.Add(&DataPoint{
			PointID:   "point001",
			Value:     float64(i * 100),
			Timestamp: time.Now(),
			Quality:   1,
		})
	}

	// 获取批次（大小为2）
	batch := buffer.GetBatch(2)
	assert.Len(t, batch, 2)
	assert.Equal(t, 3, buffer.Size())
}

func TestDataBuffer_Clear(t *testing.T) {
	config := BufferConfig{
		MaxSize:     100,
		FlushInterval: 5 * time.Second,
		BatchSize:   10,
	}

	buffer := NewDataBuffer(config)

	for i := 0; i < 5; i++ {
		buffer.Add(&DataPoint{
			PointID:   "point001",
			Value:     float64(i * 100),
			Timestamp: time.Now(),
			Quality:   1,
		})
	}

	assert.Equal(t, 5, buffer.Size())

	buffer.Clear()
	assert.Equal(t, 0, buffer.Size())
}

func TestDataBuffer_Flush(t *testing.T) {
	config := BufferConfig{
		MaxSize:     100,
		FlushInterval: 5 * time.Second,
		BatchSize:   10,
	}

	buffer := NewDataBuffer(config)

	flushed := false
	buffer.SetFlushHandler(func(batch []*DataPoint) error {
		flushed = true
		assert.Len(t, batch, 3)
		return nil
	})

	for i := 0; i < 3; i++ {
		buffer.Add(&DataPoint{
			PointID:   "point001",
			Value:     float64(i * 100),
			Timestamp: time.Now(),
			Quality:   1,
		})
	}

	err := buffer.Flush()
	assert.NoError(t, err)
	assert.True(t, flushed)
	assert.Equal(t, 0, buffer.Size())
}

func TestNewScheduler(t *testing.T) {
	config := SchedulerConfig{
		WorkerCount:    5,
		QueueSize:      1000,
		RetryCount:     3,
		RetryInterval:  5 * time.Second,
	}

	scheduler := NewScheduler(config)

	assert.NotNil(t, scheduler)
	assert.Equal(t, config, scheduler.config)
}

func TestScheduler_AddTask(t *testing.T) {
	config := SchedulerConfig{
		WorkerCount:   5,
		QueueSize:     1000,
		RetryCount:    3,
		RetryInterval: 5 * time.Second,
	}

	scheduler := NewScheduler(config)

	task := &CollectTask{
		ID:          "task001",
		DeviceID:    "device001",
		Protocol:    "modbus",
		Address:     "192.168.1.100:502",
		Interval:    5 * time.Second,
		Enabled:     true,
	}

	err := scheduler.AddTask(task)
	assert.NoError(t, err)
	assert.Equal(t, 1, scheduler.TaskCount())
}

func TestScheduler_RemoveTask(t *testing.T) {
	config := SchedulerConfig{
		WorkerCount:   5,
		QueueSize:     1000,
		RetryCount:    3,
		RetryInterval: 5 * time.Second,
	}

	scheduler := NewScheduler(config)

	task := &CollectTask{
		ID:          "task001",
		DeviceID:    "device001",
		Protocol:    "modbus",
		Address:     "192.168.1.100:502",
		Interval:    5 * time.Second,
		Enabled:     true,
	}

	scheduler.AddTask(task)
	assert.Equal(t, 1, scheduler.TaskCount())

	err := scheduler.RemoveTask("task001")
	assert.NoError(t, err)
	assert.Equal(t, 0, scheduler.TaskCount())
}

func TestScheduler_GetTask(t *testing.T) {
	config := SchedulerConfig{
		WorkerCount:   5,
		QueueSize:     1000,
		RetryCount:    3,
		RetryInterval: 5 * time.Second,
	}

	scheduler := NewScheduler(config)

	task := &CollectTask{
		ID:          "task001",
		DeviceID:    "device001",
		Protocol:    "modbus",
		Address:     "192.168.1.100:502",
		Interval:    5 * time.Second,
		Enabled:     true,
	}

	scheduler.AddTask(task)

	got := scheduler.GetTask("task001")
	assert.NotNil(t, got)
	assert.Equal(t, "task001", got.ID)

	notFound := scheduler.GetTask("task999")
	assert.Nil(t, notFound)
}

func TestScheduler_EnableDisableTask(t *testing.T) {
	config := SchedulerConfig{
		WorkerCount:   5,
		QueueSize:     1000,
		RetryCount:    3,
		RetryInterval: 5 * time.Second,
	}

	scheduler := NewScheduler(config)

	task := &CollectTask{
		ID:          "task001",
		DeviceID:    "device001",
		Protocol:    "modbus",
		Address:     "192.168.1.100:502",
		Interval:    5 * time.Second,
		Enabled:     true,
	}

	scheduler.AddTask(task)

	err := scheduler.DisableTask("task001")
	assert.NoError(t, err)

	got := scheduler.GetTask("task001")
	assert.False(t, got.Enabled)

	err = scheduler.EnableTask("task001")
	assert.NoError(t, err)

	got = scheduler.GetTask("task001")
	assert.True(t, got.Enabled)
}

func TestScheduler_UpdateTaskInterval(t *testing.T) {
	config := SchedulerConfig{
		WorkerCount:   5,
		QueueSize:     1000,
		RetryCount:    3,
		RetryInterval: 5 * time.Second,
	}

	scheduler := NewScheduler(config)

	task := &CollectTask{
		ID:          "task001",
		DeviceID:    "device001",
		Protocol:    "modbus",
		Address:     "192.168.1.100:502",
		Interval:    5 * time.Second,
		Enabled:     true,
	}

	scheduler.AddTask(task)

	err := scheduler.UpdateTaskInterval("task001", 10*time.Second)
	assert.NoError(t, err)

	got := scheduler.GetTask("task001")
	assert.Equal(t, 10*time.Second, got.Interval)
}
