# Jaeger 部署配置说明
# 新能源监控系统

## Jaeger 组件说明

### 1. Jaeger Agent (Sidecar 模式)
每个服务实例旁边部署一个 Agent，负责：
- 接收应用发送的追踪数据
- 批量转发给 Collector
- 支持多种协议：UDP (6831), HTTP (14268), gRPC (14250)

### 2. Jaeger Collector
集中收集追踪数据：
- 接收 Agent 发送的数据
- 数据验证和转换
- 写入存储后端

### 3. Jaeger Query
查询和展示服务：
- REST API 查询接口
- Web UI 界面 (端口 16686)

### 4. Storage Backend
存储后端选择：
- Elasticsearch (推荐生产环境)
- Cassandra
- Kafka
- Memory (仅用于测试)

## Go 应用集成

### 1. 安装依赖
```bash
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/exporters/jaeger
go get go.opentelemetry.io/otel/sdk/trace
```

### 2. 初始化 Tracer
```go
package tracing

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    tracesdk "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type Config struct {
    ServiceName    string
    Environment    string
    JaegerEndpoint string // 例如: http://jaeger:14268/api/traces
    SampleRate     float64
}

func InitTracer(cfg *Config) (func(context.Context) error, error) {
    // 创建 Jaeger exporter
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.JaegerEndpoint)))
    if err != nil {
        return nil, err
    }
    
    // 创建 TracerProvider
    tp := tracesdk.NewTracerProvider(
        tracesdk.WithBatcher(exp),
        tracesdk.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(cfg.ServiceName),
            semconv.DeploymentEnvironmentKey.String(cfg.Environment),
        )),
        tracesdk.WithSampler(tracesdk.TraceIDRatioBased(cfg.SampleRate)),
    )
    
    // 注册为全局 TracerProvider
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))
    
    return tp.Shutdown, nil
}
```

### 3. 在应用中使用
```go
package main

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func main() {
    // 初始化 Tracer
    shutdown, err := tracing.InitTracer(&tracing.Config{
        ServiceName:    "api-server",
        Environment:    "production",
        JaegerEndpoint: "http://jaeger:14268/api/traces",
        SampleRate:     0.1, // 10% 采样率
    })
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(context.Background())
    
    tracer := otel.Tracer("api-server")
    
    // 创建 Span
    ctx, span := tracer.Start(context.Background(), "operation-name")
    defer span.End()
    
    // 添加属性
    span.SetAttributes(
        attribute.String("user.id", "12345"),
        attribute.Int("request.size", 1024),
    )
    
    // 记录事件
    span.AddEvent("processing-start", trace.WithAttributes(
        attribute.String("step", "validation"),
    ))
    
    // 记录错误
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    }
}
```

### 4. HTTP 中间件
```go
package middleware

import (
    "net/http"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

func TracingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tracer := otel.Tracer("http-server")
        
        // 从请求中提取 Trace Context
        ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
        
        // 创建 Span
        spanName := r.Method + " " + r.URL.Path
        ctx, span := tracer.Start(ctx, spanName,
            trace.WithAttributes(
                attribute.String("http.method", r.Method),
                attribute.String("http.url", r.URL.String()),
                attribute.String("http.host", r.Host),
            ),
        )
        defer span.End()
        
        // 调用下一个处理器
        next.ServeHTTP(w, r.WithContext(ctx))
        
        // 记录状态码
        if statusCode, ok := ctx.Value("statusCode").(int); ok {
            span.SetAttributes(attribute.Int("http.status_code", statusCode))
            if statusCode >= 400 {
                span.SetStatus(codes.Error, http.StatusText(statusCode))
            }
        }
    })
}
```

### 5. gRPC 拦截器
```go
package interceptor

import (
    "context"
    "google.golang.org/grpc"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        tracer := otel.Tracer("grpc-server")
        
        ctx, span := tracer.Start(ctx, info.FullMethod,
            trace.WithAttributes(
                attribute.String("rpc.system", "grpc"),
                attribute.String("rpc.method", info.FullMethod),
            ),
        )
        defer span.End()
        
        resp, err := handler(ctx, req)
        if err != nil {
            span.RecordError(err)
            span.SetStatus(codes.Error, err.Error())
        }
        
        return resp, err
    }
}
```

## 环境变量配置

```bash
# Jaeger Agent 配置
export JAEGER_AGENT_HOST=jaeger-agent
export JAEGER_AGENT_PORT=6831

# Jaeger Collector 配置
export JAEGER_ENDPOINT=http://jaeger-collector:14268/api/traces

# 采样配置
export JAEGER_SAMPLER_TYPE=probabilistic
export JAEGER_SAMPLER_PARAM=0.1

# 服务名称
export JAEGER_SERVICE_NAME=api-server

# 日志级别
export JAEGER_LOG_LEVEL=info
```

## 最佳实践

1. **采样策略**
   - 开发环境：100% 采样
   - 测试环境：50% 采样
   - 生产环境：10% 采样（根据流量调整）

2. **Span 命名**
   - HTTP: `{METHOD} {PATH}` (例如: `GET /api/v1/stations`)
   - gRPC: `{PACKAGE}.{SERVICE}/{METHOD}`
   - 数据库: `{DB_OPERATION} {TABLE}`

3. **属性命名**
   - 使用标准属性：`http.method`, `http.status_code`, `db.system`
   - 自定义属性使用前缀：`custom.*`

4. **性能优化**
   - 使用 Batch 导出而非同步导出
   - 合理设置采样率
   - 避免在高频路径创建过多 Span

5. **敏感信息**
   - 不要在 Span 中记录密码、token 等敏感信息
   - 对用户数据进行脱敏处理
