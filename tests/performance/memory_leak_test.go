package performance

import (
	"context"
	"fmt"
	"runtime"
	"runtime/pprof"
	"sync"
	"testing"
	"time"
)

// BenchmarkMemoryLeakLongRunning 长时间运行内存泄漏测试
func BenchmarkMemoryLeakLongRunning(b *testing.B) {
	// 创建需要测试的对象
	collector := NewMockCollector("leak-test", 10000)
	ctx := context.Background()
	config := &CollectorConfig{
		ID:         "leak-test",
		BufferSize: 10000,
		BatchSize:  1000,
	}
	
	if err := collector.Initialize(ctx, config); err != nil {
		b.Fatalf("Failed to initialize: %v", err)
	}
	
	if err := collector.Start(ctx); err != nil {
		b.Fatalf("Failed to start: %v", err)
	}
	defer collector.Stop(ctx)
	
	// 记录初始内存状态
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	
	// 运行较长时间
	iterations := 10000
	
	b.ResetTimer()
	
	for i := 0; i < iterations; i++ {
		_, err := collector.Collect(ctx)
		if err != nil {
			b.Errorf("Collect failed: %v", err)
		}
		
		// 每1000次迭代检查一次内存
		if i%1000 == 0 {
			runtime.GC()
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			
			// 如果内存增长超过100MB，可能存在泄漏
			if m.Alloc > m1.Alloc+100*1024*1024 {
				b.Logf("Potential memory leak detected at iteration %d: Alloc=%d MB", 
					i, m.Alloc/1024/1024)
			}
		}
	}
	
	// 最终内存检查
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	
	b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MB_growth")
	b.ReportMetric(float64(m2.HeapAlloc)/1024/1024, "MB_final_heap")
	b.ReportMetric(float64(m2.NumGC), "gc_cycles")
	
	// 检查是否存在明显的内存泄漏
	if m2.Alloc > m1.Alloc+50*1024*1024 {
		b.Errorf("Potential memory leak: memory grew by %d MB", (m2.Alloc-m1.Alloc)/1024/1024)
	}
}

// BenchmarkObjectLifecycle 对象生命周期测试
func BenchmarkObjectLifecycle(b *testing.B) {
	objectTypes := []struct {
		name   string
		create func() interface{}
	}{
		{
			name: "SimpleStruct",
			create: func() interface{} {
				return &SimpleStruct{
					ID:      1,
					Name:    "test",
					Value:   123.45,
					Created: time.Now(),
				}
			},
		},
		{
			name: "ComplexStruct",
			create: func() interface{} {
				return &ComplexStruct{
					ID:       1,
					Name:     "test",
					Metadata: make(map[string]string),
					Items:    make([]*SimpleStruct, 100),
					Data:     make([]byte, 1024),
				}
			},
		},
		{
			name: "SliceStruct",
			create: func() interface{} {
				return make([]int, 10000)
			},
		},
		{
			name: "MapStruct",
			create: func() interface{} {
				m := make(map[string]interface{})
				for i := 0; i < 1000; i++ {
					m[fmt.Sprintf("key-%d", i)] = i
				}
				return m
			},
		},
	}
	
	for _, ot := range objectTypes {
		b.Run(ot.name, func(b *testing.B) {
			runtime.GC()
			var m1 runtime.MemStats
			runtime.ReadMemStats(&m1)
			
			b.ResetTimer()
			
			// 创建和销毁对象
			for i := 0; i < b.N; i++ {
				obj := ot.create()
				_ = obj // 使用对象防止被优化掉
			}
			
			runtime.GC()
			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)
			
			b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MB_allocated")
			b.ReportMetric(float64(m2.NumGC-m1.NumGC), "gc_triggered")
		})
	}
}

// BenchmarkGCPressure GC压力测试
func BenchmarkGCPressure(b *testing.B) {
	allocRates := []int{
		100 * 1024,      // 100KB/iteration
		1024 * 1024,     // 1MB/iteration
		10 * 1024 * 1024, // 10MB/iteration
	}
	
	for _, rate := range allocRates {
		b.Run(fmt.Sprintf("AllocRate_%dKB", rate/1024), func(b *testing.B) {
			var totalAllocated int64
			var gcCount uint32
			
			// 获取初始GC次数
			var m1 runtime.MemStats
			runtime.ReadMemStats(&m1)
			initialGC := m1.NumGC
			
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				// 分配内存
				data := make([]byte, rate)
				for j := 0; j < len(data); j++ {
					data[j] = byte(j % 256)
				}
				
				totalAllocated += int64(rate)
				
				// 让对象可以被回收
				_ = data
			}
			
			// 获取最终GC次数
			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)
			gcCount = m2.NumGC - initialGC
			
			b.ReportMetric(float64(totalAllocated)/1024/1024, "MB_total_allocated")
			b.ReportMetric(float64(gcCount), "gc_count")
			b.ReportMetric(float64(totalAllocated)/float64(gcCount)/1024/1024, "MB_per_gc")
		})
	}
}

// BenchmarkMemoryFragmentation 内存碎片测试
func BenchmarkMemoryFragmentation(b *testing.B) {
	b.Run("FragmentedAllocation", func(b *testing.B) {
		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)
		
		// 创建大量小对象，导致内存碎片
		slices := make([][]byte, 0)
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			// 分配不同大小的对象
			size := (i%100 + 1) * 100 // 100字节到10KB
			data := make([]byte, size)
			slices = append(slices, data)
			
			// 随机释放一些对象
			if i%10 == 0 && len(slices) > 100 {
				// 释放一半的对象
				for j := 0; j < len(slices)/2; j++ {
					slices[j] = nil
				}
				slices = slices[len(slices)/2:]
			}
		}
		
		runtime.GC()
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)
		
		b.ReportMetric(float64(m2.HeapInuse)/1024/1024, "MB_heap_inuse")
		b.ReportMetric(float64(m2.HeapAlloc)/1024/1024, "MB_heap_alloc")
		b.ReportMetric(float64(m2.HeapSys)/1024/1024, "MB_heap_sys")
		
		// 碎片率 = (HeapInuse - HeapAlloc) / HeapInuse
		fragmentation := float64(m2.HeapInuse-m2.HeapAlloc) / float64(m2.HeapInuse) * 100
		b.ReportMetric(fragmentation, "fragmentation_percent")
	})
}

// BenchmarkGoroutineLeak 协程泄漏测试
func BenchmarkGoroutineLeak(b *testing.B) {
	// 记录初始协程数
	initialGoroutines := runtime.NumGoroutine()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// 启动协程
		go func() {
			time.Sleep(time.Millisecond * 10)
		}()
		
		// 每100次检查一次协程数
		if i%100 == 0 {
			currentGoroutines := runtime.NumGoroutine()
			if currentGoroutines > initialGoroutines+100 {
				b.Logf("Potential goroutine leak: %d goroutines (initial: %d)", 
					currentGoroutines, initialGoroutines)
			}
		}
	}
	
	// 等待所有协程完成
	time.Sleep(time.Second)
	
	finalGoroutines := runtime.NumGoroutine()
	b.ReportMetric(float64(finalGoroutines), "final_goroutines")
	b.ReportMetric(float64(initialGoroutines), "initial_goroutines")
	
	// 检查是否存在协程泄漏
	if finalGoroutines > initialGoroutines+10 {
		b.Errorf("Potential goroutine leak: %d goroutines leaked", finalGoroutines-initialGoroutines)
	}
}

// BenchmarkChannelLeak Channel泄漏测试
func BenchmarkChannelLeak(b *testing.B) {
	b.ResetTimer()
	
	var channels []chan int
	
	for i := 0; i < b.N; i++ {
		// 创建channel
		ch := make(chan int, 100)
		channels = append(channels, ch)
		
		// 启动生产者和消费者
		go func(c chan int) {
			for j := 0; j < 10; j++ {
				c <- j
			}
			close(c)
		}(ch)
		
		go func(c chan int) {
			for range c {
				// 消费数据
			}
		}(ch)
		
		// 每1000次清理一次
		if i%1000 == 0 && len(channels) > 100 {
			channels = channels[100:]
		}
	}
	
	// 等待所有协程完成
	time.Sleep(time.Millisecond * 100)
	
	b.ReportMetric(float64(len(channels)), "channels_remaining")
}

// BenchmarkResourceCleanup 资源清理测试
func BenchmarkResourceCleanup(b *testing.B) {
	b.Run("WithDefer", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			func() {
				resource := NewMockResource()
				defer resource.Close()
				
				// 使用资源
				resource.DoWork()
			}()
		}
	})
	
	b.Run("WithoutDefer", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			resource := NewMockResource()
			resource.DoWork()
			resource.Close()
		}
	})
}

// BenchmarkMemoryProfile 内存性能分析
func BenchmarkMemoryProfile(b *testing.B) {
	// 创建内存profile文件
	memProfile, err := CreateMemProfile("mem_leak.prof")
	if err != nil {
		b.Fatalf("Failed to create memory profile: %v", err)
	}
	defer memProfile.Close()
	
	// 执行测试
	collector := NewMockCollector("profile-test", 10000)
	ctx := context.Background()
	config := &CollectorConfig{
		ID:         "profile-test",
		BufferSize: 10000,
		BatchSize:  1000,
	}
	
	collector.Initialize(ctx, config)
	collector.Start(ctx)
	defer collector.Stop(ctx)
	
	runtime.GC()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := collector.Collect(ctx)
		if err != nil {
			b.Errorf("Collect failed: %v", err)
		}
	}
	
	// 写入内存profile
	runtime.GC()
	pprof.WriteHeapProfile(memProfile)
}

// BenchmarkFinalizer Finalizer测试
func BenchmarkFinalizer(b *testing.B) {
	b.Run("WithFinalizer", func(b *testing.B) {
		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			obj := &ObjectWithFinalizer{ID: i}
			runtime.SetFinalizer(obj, func(o *ObjectWithFinalizer) {
				// 清理资源
				o.Cleanup()
			})
		}
		
		runtime.GC()
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)
		
		b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MB_allocated")
	})
	
	b.Run("WithoutFinalizer", func(b *testing.B) {
		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			obj := &ObjectWithFinalizer{ID: i}
			_ = obj
		}
		
		runtime.GC()
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)
		
		b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MB_allocated")
	})
}

// BenchmarkSyncPoolMemory sync.Pool内存优化测试
func BenchmarkSyncPoolMemory(b *testing.B) {
	b.Run("WithSyncPool", func(b *testing.B) {
		pool := &sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024)
			},
		}
		
		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			data := pool.Get().([]byte)
			// 使用数据
			for j := 0; j < len(data); j++ {
				data[j] = byte(j)
			}
			pool.Put(data)
		}
		
		runtime.GC()
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)
		
		b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MB_allocated")
	})
	
	b.Run("WithoutSyncPool", func(b *testing.B) {
		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			data := make([]byte, 1024)
			// 使用数据
			for j := 0; j < len(data); j++ {
				data[j] = byte(j)
			}
		}
		
		runtime.GC()
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)
		
		b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1024/1024, "MB_allocated")
	})
}

// 辅助类型

type SimpleStruct struct {
	ID      int
	Name    string
	Value   float64
	Created time.Time
}

type ComplexStruct struct {
	ID       int
	Name     string
	Metadata map[string]string
	Items    []*SimpleStruct
	Data     []byte
}

type CollectorConfig struct {
	ID         string
	BufferSize int
	BatchSize  int
}

type MockResource struct {
	closed bool
}

func NewMockResource() *MockResource {
	return &MockResource{}
}

func (r *MockResource) DoWork() {
	time.Sleep(time.Microsecond)
}

func (r *MockResource) Close() {
	r.closed = true
}

type ObjectWithFinalizer struct {
	ID int
}

func (o *ObjectWithFinalizer) Cleanup() {
	// 清理资源
}

func CreateMemProfile(filename string) (file interface{ Close() error }, err error) {
	// 返回一个简单的文件接口
	return &mockFile{}, nil
}

type mockFile struct{}

func (f *mockFile) Close() error { return nil }
