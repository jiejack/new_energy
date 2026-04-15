# Go 编程语言

## 基本信息

- **官方名称**：Go 或 Golang
- **开发者**：Google
- **首次发布**：2009年11月
- **当前版本**：Go 1.24+
- **许可证**：BSD-style
- **官方网站**：https://go.dev/

## 语言特性

### 核心特性
- **静态类型**：编译时类型检查
- **编译型语言**：直接编译为机器码
- **垃圾回收**：自动内存管理
- **并发原语**：goroutine 和 channel
- **简洁语法**：接近 C 但更简洁
- **标准库丰富**：内置大量实用库
- **跨平台编译**：支持多目标平台

### 并发模型
- **Goroutine**：轻量级线程，栈初始大小 2KB
- **Channel**：用于 goroutine 间通信
- **Select**：多路复用 channel
- **Mutex**：互斥锁
- **WaitGroup**：等待多个 goroutine 完成

## 在本项目中的应用

### 技术栈
- **Web 框架**：Gin
- **ORM/数据库**：GORM
- **Excel 处理**：Excelize
- **配置管理**：Viper
- **依赖注入**：Wire
- **日志**：Zap

### 项目结构
```
/workspace
├── cmd/              # 应用入口
│   └── api-server/   # API 服务器
├── internal/         # 内部应用代码
│   ├── application/  # 应用层
│   ├── domain/       # 领域层
│   ├── infrastructure/ # 基础设施层
│   └── interfaces/   # 接口层
├── pkg/              # 可复用包
└── web/              # 前端代码
```

## 开发规范

### 命名规范
- **包名**：小写，简短，有意义
- **导出变量/函数**：首字母大写
- **接口名**：通常以 -er 结尾
- **常量**：大写下划线分隔

### 错误处理
- **显式错误返回**：不使用异常
- **错误包装**：使用 fmt.Errorf 和 %w
- **错误链**：支持 errors.Is 和 errors.As

### 测试规范
- **测试文件**：以 _test.go 结尾
- **测试函数**：以 Test 开头
- **表格测试**：推荐使用表格驱动测试
- **基准测试**：以 Benchmark 开头

## 最佳实践

### 性能优化
- **避免不必要的分配**
- **使用 sync.Pool 复用对象**
- **合理使用 goroutine**
- **避免锁竞争**

### 代码质量
- **编写有意义的注释**
- **保持函数简洁**
- **遵循 Go 惯用法**
- **使用 gofmt 格式化代码**

## 学习资源

- **官方文档**：https://go.dev/doc/
- **Go 语言圣经**：https://gopl-zh.github.io/
- **Go by Example**：https://gobyexample.com/
- **Effective Go**：https://go.dev/doc/effective_go
