# Swagger API 文档说明

## 概述

本项目使用 [swaggo/swag](https://github.com/swaggo/swag) 自动生成 Swagger API 文档，支持在线调试和完整的 API 文档展示。

## 访问 Swagger UI

启动 API 服务器后，访问以下地址查看 Swagger UI：

```
http://localhost:8080/swagger/index.html
```

## 生成 Swagger 文档

### 使用 Makefile

```bash
make swagger
```

### 直接使用 swag 命令

```bash
swag init -g cmd/api-server/main.go -o docs --exclude pkg/protocol/iec61850
```

## Swagger 注解说明

### 通用 API 信息

在 `main.go` 文件顶部添加以下注解：

```go
// @title 新能源监控系统 API
// @version 1.0
// @description 新能源监控系统RESTful API接口文档
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT认证令牌，格式: Bearer {token}
```

### API 操作注解

每个 API 处理函数添加以下注解：

```go
// @Summary 获取区域列表
// @Description 获取所有区域的列表，支持按父区域ID过滤
// @Tags 区域管理
// @Accept json
// @Produce json
// @Param parent_id query string false "父区域ID"
// @Success 200 {object} dto.Response{data=[]dto.RegionResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /regions [get]
func listRegions(c *gin.Context) {
    // ...
}
```

### 参数类型

- `query`: URL 查询参数
- `path`: URL 路径参数
- `body`: 请求体参数
- `header`: 请求头参数

### 响应类型

使用 `dto.Response` 和 `dto.PagedResponse` 作为通用响应结构：

```go
// @Success 200 {object} dto.Response{data=dto.RegionResponse}
// @Success 200 {object} dto.PagedResponse{data=[]dto.StationResponse}
```

## API 标签分组

系统支持以下 API 标签分组：

- **区域管理**: 区域的增删改查操作
- **厂站管理**: 厂站的增删改查操作
- **设备管理**: 设备的增删改查操作
- **采集点管理**: 采集点的增删改查操作
- **告警管理**: 告警查询和处理操作
- **数据查询**: 实时数据和历史数据查询
- **控制操作**: 遥控和参数设置操作
- **AI服务**: 智能问答和配置建议
- **用户管理**: 用户认证和管理操作

## 使用 Swagger UI 进行 API 测试

1. 访问 Swagger UI: `http://localhost:8080/swagger/index.html`
2. 选择要测试的 API 接口
3. 点击 "Try it out" 按钮
4. 填写必要的参数
5. 点击 "Execute" 执行请求
6. 查看响应结果

## JWT 认证

需要认证的 API 接口，在 Swagger UI 中：

1. 点击右上角的 "Authorize" 按钮
2. 输入 JWT Token（格式：`Bearer {token}`）
3. 点击 "Authorize" 确认
4. 之后的所有请求都会自动携带认证信息

## 导出 API 文档

### JSON 格式

```
http://localhost:8080/swagger/doc.json
```

### YAML 格式

文档生成在 `docs/swagger.yaml` 文件中。

## 其他命令

### 在 Docker 中运行 Swagger UI

```bash
make swagger-serve
```

访问 `http://localhost:8081` 查看 Swagger UI。

### 验证 Swagger 规范

```bash
make swagger-validate
```

## 注意事项

1. 修改 API 注解后，需要重新运行 `make swagger` 生成文档
2. DTO 结构体的 `example` 标签用于生成示例值
3. `interface{}` 类型不支持 `example` 标签
4. Map 类型的 `example` 标签需要特殊处理

## 相关文件

- API 注解: [cmd/api-server/main.go](../cmd/api-server/main.go)
- DTO 定义: [internal/api/dto/](../internal/api/dto/)
- 生成的文档: [docs/](../docs/)
  - `docs.go`: Go 代码
  - `swagger.json`: JSON 格式
  - `swagger.yaml`: YAML 格式
