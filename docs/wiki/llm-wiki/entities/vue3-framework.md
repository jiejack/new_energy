# Vue 3 前端框架

## 基本信息

- **名称**：Vue.js
- **版本**：Vue 3+
- **开发者**：Evan You 及社区
- **首次发布**：2014年2月（Vue 3 于 2020年9月发布）
- **许可证**：MIT
- **官方网站**：https://vuejs.org/

## 核心特性

### Composition API
- **setup 函数**：组件逻辑入口
- **响应式 API**：ref、reactive、computed、watch
- **生命周期钩子**：onMounted、onUpdated、onUnmounted 等
- **自定义 Hooks**：逻辑复用的最佳方式

### 性能优化
- **虚拟 DOM 优化**：编译时优化
- **Tree-shaking 支持**：按需引入
- **更小的包体积**：相比 Vue 2 减少约 40%
- **更快的渲染**：渲染速度提升约 1.3-2 倍

### TypeScript 支持
- **完整类型定义**：内置 TypeScript 支持
- **类型推断**：更好的类型推断
- **类型安全**：编译时类型检查

## 在本项目中的应用

### 技术栈
- **构建工具**：Vite
- **语言**：TypeScript
- **UI 组件库**：Element Plus
- **状态管理**：Pinia
- **路由**：Vue Router
- **HTTP 客户端**：Axios

### 项目结构
```
web/src/
├── assets/          # 静态资源
├── components/      # 公共组件
├── router/          # 路由配置
├── stores/          # Pinia 状态管理
├── utils/           # 工具函数
├── views/           # 页面组件
│   ├── alarm/       # 告警相关页面
│   ├── data/        # 数据相关页面
│   └── ...
├── App.vue          # 根组件
└── main.ts          # 入口文件
```

### 页面示例

#### 告警规则页面
- **路径**：`web/src/views/alarm/rule/index.vue`
- **功能**：告警规则列表、创建、编辑、删除、启用/禁用
- **特性**：表格展示、分页、搜索、表单验证

#### 统计报表页面
- **路径**：`web/src/views/data/report/index.vue`
- **功能**：数据查询、图表展示、Excel 导出
- **特性**：日期范围选择、多维度统计、数据可视化

## 开发规范

### 组件命名
- **单文件组件**：PascalCase，如 `AlarmRuleList.vue`
- **组件注册**：全局组件使用 PascalCase
- **模板中使用**：kebab-case 或 PascalCase

### Composition API 使用
- **优先使用 setup script**：`<script setup lang="ts">`
- **响应式数据**：基本类型用 ref，对象用 reactive
- **计算属性**：使用 computed
- **侦听器**：合理使用 watch 和 watchEffect

### 代码组织
- **逻辑复用**：使用自定义 Hooks
- **组件拆分**：保持组件单一职责
- **类型定义**：完善的 TypeScript 类型
- **注释规范**：必要的代码注释

## 最佳实践

### 性能优化
- **合理使用 v-show 和 v-if**
- **列表渲染使用 key**
- **避免不必要的响应式数据**
- **使用 defineAsyncComponent 懒加载组件**
- **合理使用 computed 缓存计算结果**

### 代码质量
- **类型安全**：充分利用 TypeScript
- **组件通信**：使用 props 和 emits
- **状态管理**：合理使用 Pinia
- **错误处理**：完善的错误捕获和处理
- **代码格式化**：使用 Prettier 和 ESLint

### 开发体验
- **热模块替换**：利用 Vite 的 HMR
- **开发工具**：使用 Vue DevTools
- **调试技巧**：合理使用 console 和 debugger

## 学习资源

- **官方文档**：https://vuejs.org/
- **中文文档**：https://cn.vuejs.org/
- **Vue Router**：https://router.vuejs.org/
- **Pinia**：https://pinia.vuejs.org/
- **TypeScript 文档**：https://www.typescriptlang.org/
