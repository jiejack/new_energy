# Playwright E2E 测试

## 测试文件结构

```
e2e/
├── auth.spec.ts          # 登录流程测试
├── dashboard.spec.ts     # 监控大屏测试
├── config.spec.ts        # 配置管理测试
├── data.spec.ts          # 数据查询测试
├── alarm.spec.ts         # 告警管理测试
└── utils.ts              # 测试辅助函数
```

## 运行测试

### 安装依赖

```bash
npm install
npx playwright install
```

### 运行所有测试

```bash
npm run test:e2e
```

### 运行特定浏览器测试

```bash
npx playwright test --project=chromium
npx playwright test --project=firefox
```

### 运行特定测试文件

```bash
npx playwright test e2e/auth.spec.ts
```

### 运行UI模式

```bash
npm run test:e2e:ui
```

### 调试模式

```bash
npm run test:e2e:debug
```

### 查看测试报告

```bash
npm run test:e2e:report
```

## 测试覆盖范围

### 1. 登录流程测试 (auth.spec.ts)

- 登录页面显示
- 登录成功流程
- 登录失败提示
- 表单验证
- Token持久化
- 记住我功能
- 回车键登录
- 重定向功能
- 登出流程

### 2. 监控大屏测试 (dashboard.spec.ts)

- 大屏加载
- 统计卡片显示
- 电站列表显示
- 告警列表显示
- 实时图表
- 电站地图
- 实时数据更新
- 刷新功能
- 全屏显示
- 响应式布局

### 3. 配置管理测试 (config.spec.ts)

- 区域管理CRUD
- 电站管理CRUD
- 设备管理CRUD
- 采集点管理CRUD
- 表单验证
- 分页功能

### 4. 数据查询测试 (data.spec.ts)

- 历史数据查询
- 时间范围选择
- 采集点选择
- 数据导出
- 数据聚合
- 数据可视化
- 性能测试

### 5. 告警管理测试 (alarm.spec.ts)

- 告警列表显示
- 告警确认流程
- 批量确认
- 告警解决
- 告警详情
- 告警筛选
- 告警排序
- 告警分页
- 告警导出
- 实时告警

## 测试配置

### playwright.config.ts

- 测试目录: `./e2e`
- 浏览器: Chromium, Firefox
- 基础URL: `http://localhost:3000`
- 失败时截图
- 失败时录制视频
- 失败时追踪

### 测试隔离

每个测试文件使用独立的浏览器上下文，确保测试之间相互隔离。

### 测试数据

使用 `testData` 对象生成唯一的测试数据，避免数据冲突。

## 最佳实践

1. **使用辅助函数**: 使用 `e2e/utils.ts` 中的辅助函数简化测试代码
2. **等待策略**: 使用合适的等待策略，避免硬编码延迟
3. **选择器**: 使用语义化的选择器，避免依赖DOM结构
4. **断言**: 使用明确的断言，提高测试可读性
5. **测试隔离**: 确保每个测试独立运行，不依赖其他测试
6. **清理数据**: 测试后清理创建的测试数据

## 故障排查

### 测试超时

- 检查网络连接
- 增加超时时间
- 检查元素是否存在

### 元素未找到

- 检查选择器是否正确
- 等待元素加载完成
- 检查页面是否正确渲染

### 测试失败

- 查看截图和视频
- 查看测试报告
- 使用调试模式

## 持续集成

在CI环境中运行测试:

```bash
CI=true npm run test:e2e
```

CI配置特点:
- 失败重试2次
- 单线程运行
- 生成HTML报告
