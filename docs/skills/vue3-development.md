# Vue 3 开发技能

## 概述

本项目使用 Vue 3 + TypeScript + Vite 技术栈进行前端开发。

## 技术栈

### 核心框架
- **Vue 3**：渐进式 JavaScript 框架
- **TypeScript**：类型安全的 JavaScript 超集
- **Vite**：下一代前端构建工具

### 配套库
- **Element Plus**：Vue 3 UI 组件库
- **Pinia**：状态管理库
- **Vue Router**：官方路由管理器
- **ECharts**：数据可视化库
- **Axios**：HTTP 客户端
- **Sass**：CSS 预处理器

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
└── main.ts          # 入口文件
```

## 组合式 API 最佳实践

### 使用 `<script setup>`
```vue
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const count = ref(0)
const doubled = computed(() => count.value * 2)

onMounted(() => {
  console.log('组件已挂载')
})
</script>
```

### 响应式数据
```typescript
// 基础类型使用 ref
const count = ref(0)

// 对象类型使用 reactive
const state = reactive({
  name: 'test',
  age: 18
})
```

### 计算属性
```typescript
const fullName = computed(() => {
  return `${firstName.value} ${lastName.value}`
})
```

## 组件设计原则

### 1. 单一职责
每个组件只负责一个功能，保持组件简洁。

### 2. 可复用性
将可复用的逻辑提取到 `composables/` 目录。

### 3. Props 验证
使用 TypeScript 进行类型验证：
```typescript
interface Props {
  title: string
  count?: number
}

const props = withDefaults(defineProps<Props>(), {
  count: 0
})
```

### 4. 事件定义
```typescript
interface Emits {
  (e: 'update:modelValue', value: string): void
  (e: 'submit'): void
}

const emit = defineEmits<Emits>()
```

## 状态管理

### Pinia Store 示例
```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const isLoggedIn = computed(() => user.value !== null)

  function setUser(newUser: User) {
    user.value = newUser
  }

  function logout() {
    user.value = null
  }

  return { user, isLoggedIn, setUser, logout }
})
```

## 路由管理

### 路由配置
```typescript
const routes = [
  {
    path: '/',
    component: MainLayout,
    children: [
      { path: '', redirect: '/dashboard' },
      { path: 'dashboard', component: Dashboard },
      { path: 'reports', component: ReportList }
    ]
  }
]
```

### 路由守卫
```typescript
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next('/login')
  } else {
    next()
  }
})
```

## 性能优化

### 1. 组件懒加载
```typescript
const Dashboard = defineAsyncComponent(() => 
  import('@/views/dashboard/index.vue')
)
```

### 2. 虚拟列表
对于大量数据，使用虚拟滚动：
```vue
<el-table-v2
  :columns="columns"
  :data="data"
  :width="700"
  :height="400"
/>
```

### 3. 计算属性缓存
使用 `computed` 缓存计算结果，避免重复计算。

## 调试技巧

### Vue DevTools
使用 Vue DevTools 浏览器扩展进行调试。

### 响应式调试
```typescript
import { watch } from 'vue'

watch(count, (newVal, oldVal) => {
  console.log(`count 从 ${oldVal} 变为 ${newVal}`)
})
```

## 相关资源

- [Vue 3 官方文档](https://cn.vuejs.org/)
- [Element Plus 文档](https://element-plus.org/)
- [Pinia 文档](https://pinia.vuejs.org/zh/)
