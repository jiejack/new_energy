# 前端架构说明

## 技术栈

- **框架**：Vue 3
- **语言**：TypeScript
- **构建工具**：Vite
- **UI 库**：Element Plus
- **状态管理**：Pinia
- **路由**：Vue Router
- **HTTP 客户端**：Axios
- **图表库**：ECharts
- **CSS 预处理器**：Sass

## 项目结构

```
web/src/
├── api/             # API 接口定义
├── assets/          # 静态资源
├── components/      # 公共组件
├── composables/     # 组合式函数
├── directives/      # 自定义指令
├── layouts/         # 布局组件
├── router/          # 路由配置
├── stores/          # Pinia 状态管理
├── styles/          # 全局样式
├── types/           # TypeScript 类型定义
├── utils/           # 工具函数
├── views/           # 页面组件
├── App.vue          # 根组件
├── main.ts          # 入口文件
└── vite-env.d.ts    # Vite 环境类型声明
```

## 架构设计

### 1. 组件设计

- **页面组件**：位于 `views/` 目录，对应路由页面
- **公共组件**：位于 `components/` 目录，可复用的 UI 组件
- **布局组件**：位于 `layouts/` 目录，如主布局、登录布局等

### 2. 状态管理

使用 Pinia 进行状态管理，按功能模块划分 store：

- **userStore**：用户相关状态
- **deviceStore**：设备相关状态
- **alarmStore**：告警相关状态
- **reportStore**：报表相关状态

### 3. 路由设计

```typescript
const routes = [
  {
    path: '/',
    component: MainLayout,
    children: [
      { path: '', redirect: '/dashboard' },
      { path: 'dashboard', component: Dashboard },
      { path: 'devices', component: DeviceList },
      { path: 'alarms', component: AlarmList },
      { path: 'alarm-rules', component: AlarmRuleList },
      { path: 'reports', component: ReportList },
      { path: 'settings', component: Settings },
    ]
  },
  {
    path: '/login',
    component: Login
  }
]
```

### 4. API 设计

- **统一请求封装**：在 `utils/request.ts` 中封装 Axios
- **API 接口定义**：在 `api/` 目录下按功能模块划分
- **类型定义**：在 `types/` 目录下定义请求和响应类型

### 5. 样式设计

- **全局样式**：在 `styles/` 目录下定义全局样式和变量
- **组件样式**：使用 scoped 样式或 CSS Modules
- **主题配置**：基于 Element Plus 的主题配置

## 开发规范

### 1. 代码规范

- 使用 ESLint 进行代码检查
- 使用 Prettier 进行代码格式化
- 遵循 Vue 3 组合式 API 最佳实践
- 使用 TypeScript 严格模式

### 2. 命名规范

- **组件名**：PascalCase（如 `AlarmRuleList`）
- **变量名**：camelCase（如 `deviceList`）
- **常量名**：UPPER_SNAKE_CASE（如 `API_BASE_URL`）
- **文件命名**：kebab-case（如 `alarm-rule.ts`）

### 3. 代码组织

- **组合式函数**：将相关逻辑提取到 `composables/` 目录
- **工具函数**：将通用工具函数提取到 `utils/` 目录
- **类型定义**：将 TypeScript 类型定义集中到 `types/` 目录

## 性能优化

### 1. 组件优化

- 使用 `v-memo` 缓存计算结果
- 使用 `v-for` 时添加 `key` 属性
- 避免在模板中使用复杂表达式
- 合理使用 `computed` 和 `watch`

### 2. 网络优化

- 使用 Axios 拦截器统一处理请求和响应
- 实现请求缓存机制
- 使用 WebSocket 进行实时数据推送
- 合理设置请求超时和重试机制

### 3. 构建优化

- 使用 Vite 的按需加载
- 配置合理的 chunk 分割
- 压缩静态资源
- 使用 CDN 加速静态资源

## 部署策略

### 1. 构建流程

- 执行 `npm run build` 生成生产环境代码
- 构建产物位于 `dist/` 目录

### 2. 部署方式

- 使用 Nginx 作为静态资源服务器
- 配置反向代理到后端 API
- 支持 Docker 容器化部署

### 3. 环境配置

- 开发环境：`.env.development`
- 生产环境：`.env.production`
- 测试环境：`.env.test`