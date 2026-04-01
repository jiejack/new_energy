# 新能源监控系统 - 性能测试

本目录包含新能源监控系统的完整性能测试套件，用于评估系统的采集、查询和并发处理能力。

## 目录结构

```
tests/performance/
├── collector_bench_test.go    # 采集性能测试
├── query_bench_test.go        # 查询性能测试
├── stress_test.go             # 并发压力测试
├── api_perf_test.go           # API性能测试 (新增)
├── db_perf_test.go            # 数据库性能测试 (新增)
├── memory_leak_test.go        # 内存泄漏测试 (新增)
├── mock_helpers.go            # 模拟辅助函数 (新增)
├── report.go                  # 性能测试报告生成
├── run_benchmarks.sh          # Linux/Mac运行脚本
├── run_benchmarks.ps1         # Windows PowerShell运行脚本
└── README.md                  # 本文档

web/tests/performance/
├── component-perf.test.ts     # 组件渲染性能测试 (新增)
├── state-perf.test.ts         # 状态更新性能测试 (新增)

scripts/performance/
├── run_perf_tests.sh          # Linux/Mac性能测试脚本 (新增)
├── run_perf_tests.ps1         # Windows性能测试脚本 (新增)
└── load_test.js               # k6负载测试脚本 (新增)
```

## 测试内容

### 1. 采集性能测试 (collector_bench_test.go)

测试采集器的各项性能指标：

- **100万点位模拟采集** (`BenchmarkCollectorMillionPoints`)
  - 测试大规模数据采集能力
  - 并行采集性能评估

- **并发连接测试** (`BenchmarkCollectorConcurrentConnections`)
  - 10/50/100/500/1000个并发连接
  - 连接建立和维护性能

- **数据吞吐量测试** (`BenchmarkCollectorThroughput`)
  - 1K/10K/100K/500K/1M点位吞吐量
  - 每秒处理点位数评估

- **内存占用测试** (`BenchmarkCollectorMemoryUsage`)
  - 不同规模数据的内存占用
  - 内存分配和GC性能

- **CPU使用率测试** (`BenchmarkCollectorCPUUsage`)
  - CPU使用率分析
  - 生成CPU性能分析文件

- **协程池性能测试** (`BenchmarkCollectorWorkerPool`)
  - 不同工作协程数量的性能
  - 任务调度效率

- **缓冲区性能测试** (`BenchmarkCollectorBuffer`)
  - 不同缓冲区大小的性能
  - 数据缓冲效率

### 2. 查询性能测试 (query_bench_test.go)

测试查询系统的各项性能指标：

- **500万记录查询响应测试** (`BenchmarkQueryMillionRecords`)
  - 100K/500K/1M/2M/5M记录查询
  - 查询响应时间评估

- **复杂查询性能测试** (`BenchmarkQueryComplexQuery`)
  - 简单查询
  - 时间范围查询
  - 聚合查询
  - JOIN查询
  - 复杂组合查询

- **并发查询测试** (`BenchmarkQueryConcurrent`)
  - 10/50/100/200/500并发查询
  - 并发处理能力

- **缓存命中率测试** (`BenchmarkQueryCacheHitRate`)
  - 缓存效率评估
  - 命中率统计

- **流式查询测试** (`BenchmarkQueryStream`)
  - 大数据量流式处理
  - 内存效率评估

- **查询计划优化测试** (`BenchmarkQueryPlanOptimization`)
  - 查询优化器性能
  - 执行计划生成效率

### 3. 并发压力测试 (stress_test.go)

测试系统的并发处理能力：

- **API接口压力测试** (`BenchmarkAPIPressure`)
  - 各API端点压力测试
  - GET/POST请求性能

- **API并发压力测试** (`BenchmarkAPIConcurrentPressure`)
  - 10/50/100/200/500并发请求
  - 并发处理能力

- **WebSocket连接压力测试** (`BenchmarkWebSocketPressure`)
  - 10/50/100/200/500并发连接
  - 消息收发性能

- **数据库连接池压力测试** (`BenchmarkDatabasePoolPressure`)
  - 不同连接池大小性能
  - 连接获取效率

- **消息队列压力测试** (`BenchmarkMessageQueuePressure`)
  - 生产者/消费者性能
  - 队列吞吐量

- **内存压力测试** (`BenchmarkMemoryPressure`)
  - 不同数据大小的内存使用
  - GC性能评估

- **CPU密集型任务压力测试** (`BenchmarkCPUIntensiveTask`)
  - CPU使用率分析
  - 并发计算性能

### 4. API性能测试 (api_perf_test.go) - 新增

测试API的各项性能指标：

- **API响应时间测试** (`BenchmarkAPIResponseTime`)
  - 各API端点响应时间
  - GET/POST请求性能

- **API并发处理能力测试** (`BenchmarkAPIConcurrentHandling`)
  - 10/50/100/200/500/1000并发请求
  - 并发处理能力评估

- **API吞吐量测试** (`BenchmarkAPIThroughput`)
  - 请求吞吐量评估
  - 每秒请求数统计

- **混合负载测试** (`BenchmarkAPIMixedLoad`)
  - 70%读、20%写、10%删除
  - 真实场景模拟

- **数据库查询性能测试** (`BenchmarkAPIDatabaseQueryPerformance`)
  - 简单查询、过滤查询、排序查询
  - 分页查询、复杂查询

- **JSON编解码性能测试** (`BenchmarkAPIJSONEncoding/Decoding`)
  - 不同数据大小的编解码性能
  - 序列化/反序列化效率

- **中间件性能测试** (`BenchmarkAPIMiddlewarePerformance`)
  - 认证中间件
  - 限流中间件
  - 日志中间件
  - CORS中间件

### 5. 数据库性能测试 (db_perf_test.go) - 新增

测试数据库的各项性能指标：

- **批量插入性能测试** (`BenchmarkDatabaseBatchInsert`)
  - 100/500/1K/5K/10K批量插入
  - 插入吞吐量评估

- **复杂查询性能测试** (`BenchmarkDatabaseComplexQuery`)
  - 简单查询、范围查询、时间范围查询
  - 多条件查询、聚合查询、JOIN查询

- **索引效率测试** (`BenchmarkDatabaseIndexEfficiency`)
  - 有索引 vs 无索引查询对比
  - 索引命中效率

- **连接池性能测试** (`BenchmarkDatabaseConnectionPool`)
  - 不同连接池大小性能
  - 连接获取效率

- **事务性能测试** (`BenchmarkDatabaseTransaction`)
  - 单条事务 vs 批量事务
  - 事务提交效率

- **读写混合测试** (`BenchmarkDatabaseReadWriteMix`)
  - 90/10、80/20、50/50、20/80、10/90读写比例
  - 真实场景模拟

- **并发查询测试** (`BenchmarkDatabaseConcurrentQuery`)
  - 10/50/100/200/500并发查询
  - 并发处理能力

### 6. 内存泄漏测试 (memory_leak_test.go) - 新增

测试内存管理和泄漏检测：

- **长时间运行测试** (`BenchmarkMemoryLeakLongRunning`)
  - 10000次迭代内存监控
  - 内存增长趋势分析

- **对象生命周期测试** (`BenchmarkObjectLifecycle`)
  - 简单结构体、复杂结构体
  - 切片、Map对象生命周期

- **GC压力测试** (`BenchmarkGCPressure`)
  - 不同分配速率的GC表现
  - GC触发频率分析

- **内存碎片测试** (`BenchmarkMemoryFragmentation`)
  - 碎片化内存分配
  - 碎片率计算

- **协程泄漏测试** (`BenchmarkGoroutineLeak`)
  - 协程数量监控
  - 协程泄漏检测

- **Channel泄漏测试** (`BenchmarkChannelLeak`)
  - Channel资源管理
  - 生产者/消费者模式

- **资源清理测试** (`BenchmarkResourceCleanup`)
  - defer vs 手动清理对比
  - 资源释放效率

- **sync.Pool内存优化测试** (`BenchmarkSyncPoolMemory`)
  - 使用Pool vs 不使用Pool对比
  - 内存复用效率

### 7. 前端性能测试 (web/tests/performance/) - 新增

#### 组件渲染性能测试 (component-perf.test.ts)

- **大列表渲染测试**
  - 10000条数据表格渲染
  - 虚拟滚动性能
  - 列表更新效率

- **复杂组件渲染测试**
  - 图表组件渲染
  - 表单对话框渲染
  - 嵌套组件渲染
  - 告警列表渲染

- **组件更新性能测试**
  - Props更新效率
  - 条件渲染效率
  - 图表数据更新

- **组件销毁性能测试**
  - 大型组件树销毁
  - 事件监听器清理

- **响应式性能测试**
  - 计算属性效率
  - Watcher效率
  - 深度响应式

- **插槽性能测试**
  - 普通插槽渲染
  - 作用域插槽渲染

#### 状态更新性能测试 (state-perf.test.ts)

- **Store更新性能测试**
  - 用户Store更新
  - 权限Store更新
  - 应用Store更新

- **大量数据更新测试**
  - 大数组更新
  - 批量添加/删除/更新
  - 过滤和排序

- **计算属性性能测试**
  - 过滤数据计算
  - 聚合数据计算
  - 嵌套数据计算

- **批量更新性能测试**
  - 批量更新/删除/插入

- **状态订阅性能测试**
  - 多订阅者处理
  - 深度监听

- **状态持久化性能测试**
  - localStorage读写
  - 状态序列化

- **并发更新性能测试**
  - 并发更新处理
  - 快速状态变化

- **Undo/Redo性能测试**
  - 状态历史跟踪
  - 撤销/重做效率

### 8. k6负载测试 (scripts/performance/load_test.js) - 新增

使用k6进行专业负载测试：

- **阶段性负载测试**
  - 预热阶段：30秒增加到20用户
  - 增长阶段：1分钟增加到50用户
  - 峰值阶段：2分钟增加到100用户
  - 稳定阶段：保持100用户1分钟
  - 降温阶段：30秒降到0

- **测试场景**
  - 站点列表查询 (30%)
  - 设备列表查询 (20%)
  - 测点数据查询 (20%)
  - 实时数据查询 (10%)
  - 历史数据查询 (10%)
  - 告警列表查询 (10%)

- **性能阈值**
  - P95延迟 < 500ms
  - P99延迟 < 1000ms
  - 错误率 < 5%
  - 失败率 < 1%

### 9. 性能测试报告 (report.go)

自动生成性能测试报告：

- **性能指标收集**
  - 吞吐量、延迟、内存、CPU等指标
  - 成功率、错误统计

- **基准线对比**
  - 与历史基准对比
  - 性能趋势分析

- **瓶颈分析**
  - CPU瓶颈识别
  - 内存瓶颈识别
  - IO瓶颈识别
  - 并发瓶颈识别

- **优化建议**
  - 自动生成优化建议
  - 预期影响评估

- **报告格式**
  - JSON格式报告
  - HTML格式报告（带图表）

## 运行测试

### Windows系统

使用PowerShell脚本：

```powershell
# 运行所有测试
.\tests\performance\run_benchmarks.ps1 -TestType all

# 运行采集性能测试
.\tests\performance\run_benchmarks.ps1 -TestType collector

# 运行查询性能测试
.\tests\performance\run_benchmarks.ps1 -TestType query

# 运行压力测试
.\tests\performance\run_benchmarks.ps1 -TestType stress

# 生成CPU性能分析
.\tests\performance\run_benchmarks.ps1 -TestType cpu-profile

# 生成内存性能分析
.\tests\performance\run_benchmarks.ps1 -TestType mem-profile

# 生成性能测试报告
.\tests\performance\run_benchmarks.ps1 -TestType report
```

### Linux/Mac系统

使用Bash脚本：

```bash
# 添加执行权限
chmod +x tests/performance/run_benchmarks.sh

# 运行脚本
./tests/performance/run_benchmarks.sh
```

### 直接使用go test

```bash
# 运行所有性能测试
go test -bench=. -benchmem ./tests/performance/...

# 运行特定测试
go test -bench=BenchmarkCollector -benchmem ./tests/performance/...

# 生成CPU性能分析
go test -bench=BenchmarkCollectorMillionPoints -cpuprofile=cpu.prof ./tests/performance/...

# 生成内存性能分析
go test -bench=BenchmarkCollectorMemoryUsage -memprofile=mem.prof ./tests/performance/...

# 查看性能分析
go tool pprof -http=:8080 cpu.prof
```

## 性能指标说明

### 关键指标

| 指标 | 说明 | 单位 |
|------|------|------|
| ops/s | 每秒操作数 | operations/second |
| avg_latency | 平均延迟 | milliseconds |
| p95_latency | 95%延迟 | milliseconds |
| p99_latency | 99%延迟 | milliseconds |
| memory_alloc | 内存分配 | MB |
| cpu_usage | CPU使用率 | % |
| success_rate | 成功率 | % |

### 性能基准

#### 采集性能基准

| 测试场景 | 目标性能 | 最低要求 |
|---------|---------|---------|
| 100万点位采集 | > 50,000 ops/s | > 30,000 ops/s |
| 平均延迟 | < 20ms | < 50ms |
| 内存使用 | < 500MB | < 1GB |
| CPU使用率 | < 70% | < 90% |

#### 查询性能基准

| 测试场景 | 目标性能 | 最低要求 |
|---------|---------|---------|
| 100万记录查询 | > 5,000 ops/s | > 1,000 ops/s |
| 平均延迟 | < 100ms | < 500ms |
| 缓存命中率 | > 80% | > 60% |

#### 并发性能基准

| 测试场景 | 目标性能 | 最低要求 |
|---------|---------|---------|
| 100并发API请求 | > 1,000 ops/s | > 500 ops/s |
| 100并发WebSocket | > 5,000 msg/s | > 1,000 msg/s |
| 成功率 | > 99.9% | > 99% |

## 性能分析工具

### pprof使用

```bash
# 查看CPU性能分析
go tool pprof -http=:8080 reports/performance/cpu.prof

# 查看内存性能分析
go tool pprof -http=:8080 reports/performance/mem.prof

# 命令行模式
go tool pprof reports/performance/cpu.prof
(pprof) top10    # 查看CPU占用最高的10个函数
(pprof) list 函数名  # 查看具体函数代码
(pprof) web      # 生成调用图（需要graphviz）
```

### trace使用

```bash
# 生成trace文件
go test -bench=. -trace=trace.out ./tests/performance/...

# 查看trace
go tool trace trace.out
```

## 性能优化建议

### 采集优化

1. **使用批量采集**
   - 减少网络往返次数
   - 提高吞吐量

2. **优化协程池**
   - 合理设置工作协程数量
   - 避免过度创建协程

3. **使用缓冲区**
   - 减少内存分配
   - 提高数据处理效率

### 查询优化

1. **添加索引**
   - 为常用查询字段添加索引
   - 提高查询速度

2. **使用缓存**
   - 缓存热点查询结果
   - 减少数据库访问

3. **优化查询计划**
   - 避免全表扫描
   - 使用合适的JOIN策略

### 并发优化

1. **减少锁竞争**
   - 使用读写锁
   - 使用无锁数据结构

2. **优化连接池**
   - 合理设置连接池大小
   - 复用连接

3. **异步处理**
   - 使用消息队列
   - 异步IO操作

## 持续性能监控

建议在生产环境中实施持续性能监控：

1. **集成Prometheus**
   - 收集性能指标
   - 配置告警规则

2. **定期性能测试**
   - 每周运行完整性能测试
   - 对比历史基准

3. **性能回归检测**
   - 代码提交前运行性能测试
   - 防止性能回归

## 报告示例

运行测试后，会在 `reports/performance/` 目录生成以下文件：

```
reports/performance/
├── collector_bench.txt       # 采集性能测试结果
├── query_bench.txt           # 查询性能测试结果
├── stress_test.txt           # 压力测试结果
├── cpu.prof                  # CPU性能分析文件
├── mem.prof                  # 内存性能分析文件
├── performance_report.json   # JSON格式报告
└── performance_report.html   # HTML格式报告
```

HTML报告包含：
- 性能等级评分（A-F）
- 性能总结卡片
- 系统信息
- 测试结果详情表格
- 瓶颈分析
- 优化建议
- 性能图表

## 注意事项

1. **测试环境**
   - 在独立环境运行测试
   - 避免其他程序干扰
   - 确保足够系统资源

2. **测试时间**
   - 完整测试套件需要30-60分钟
   - 可根据需要选择部分测试

3. **资源消耗**
   - 性能测试会消耗大量CPU和内存
   - 建议在测试环境运行

4. **结果解读**
   - 关注趋势而非单次结果
   - 多次运行取平均值
   - 对比基准线评估性能变化

## 故障排查

### 测试失败

1. 检查系统资源是否充足
2. 检查依赖服务是否正常
3. 查看错误日志

### 性能异常

1. 运行pprof分析
2. 检查是否有资源泄漏
3. 检查是否有死锁

### 报告生成失败

1. 检查报告目录权限
2. 检查磁盘空间
3. 查看错误信息

## 联系方式

如有问题或建议，请联系开发团队。

---

**新能源监控系统** - 性能测试套件 v1.0
