# TypeScript 编程语言

## 基本信息

- **名称**：TypeScript
- **开发者**：Microsoft
- **首次发布**：2012年10月
- **当前版本**：TypeScript 5.0+
- **许可证**：Apache 2.0
- **官方网站**：https://www.typescriptlang.org/
- **GitHub**：https://github.com/microsoft/TypeScript

## 核心特性

### 类型系统
- **静态类型检查**：编译时类型检查
- **类型推断**：智能类型推断
- **接口**：定义对象结构
- **类型别名**：自定义类型名称
- **泛型**：类型参数化
- **联合类型**：多种类型之一
- **交叉类型**：多种类型组合
- **类型守卫**：运行时类型检查

### ES6+ 支持
- **类**：面向对象编程
- **模块**：模块化开发
- **箭头函数**：简洁的函数语法
- **解构赋值**：数组和对象解构
- **展开运算符**：数组和对象展开
- **Promise/async-await**：异步编程
- **装饰器**：注解和元编程

### 工具类型
- **Partial**：所有属性可选
- **Required**：所有属性必填
- **Readonly**：所有属性只读
- **Pick**：选取部分属性
- **Omit**：排除部分属性
- **Record**：键值对类型
- **Exclude**：排除类型
- **Extract**：提取类型

## 在本项目中的应用

### 技术栈
- **Vue 3**：使用 `<script setup lang="ts">`
- **Vite**：构建工具，原生支持 TypeScript
- **Vue Router**：类型安全的路由
- **Pinia**：类型安全的状态管理
- **Axios**：类型安全的 HTTP 请求

### 类型定义示例

#### API 响应类型
```typescript
// web/src/api/types.ts
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export interface PageResponse<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}
```

#### 告警规则类型
```typescript
// web/src/views/alarm/rule/types.ts
export interface AlarmRule {
  id: number
  name: string
  pointId: number
  pointName?: string
  condition: string
  threshold: number
  level: 'info' | 'warning' | 'error' | 'critical'
  enabled: boolean
  createdAt: string
  updatedAt: string
}

export interface AlarmRuleQuery {
  name?: string
  level?: string
  enabled?: boolean
  page?: number
  pageSize?: number
}

export interface AlarmRuleForm {
  name: string
  pointId: number
  condition: '>' | '>=' | '<' | '<=' | '==' | '!='
  threshold: number
  level: 'info' | 'warning' | 'error' | 'critical'
}
```

#### 报表数据类型
```typescript
// web/src/views/data/report/types.ts
export interface ReportData {
  time: string
  pointName: string
  value: number
  unit: string
}

export interface ReportQuery {
  startTime: string
  endTime: string
  pointIds?: number[]
  interval?: 'hour' | 'day' | 'week' | 'month'
}

export interface ReportSummary {
  total: number
  avg: number
  max: number
  min: number
  count: number
}
```

### Composition API 类型使用
```vue
<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import type { AlarmRule, AlarmRuleQuery, AlarmRuleForm } from './types'

// 响应式数据类型
const loading = ref(false)
const list = ref<AlarmRule[]>([])
const query = reactive<AlarmRuleQuery>({
  page: 1,
  pageSize: 10
})

// 计算属性类型
const pointMap = computed<Map<number | string, string>>(() => {
  const map = new Map<number | string, string>()
  // ...
  return map
})

// 函数参数和返回值类型
async function fetchList(): Promise<void> {
  loading.value = true
  try {
    // ...
  } finally {
    loading.value = false
  }
}

// 生命周期钩子
onMounted(() => {
  fetchList()
})
</script>
```

## 开发规范

### 类型定义
- **接口优先**：优先使用 interface 定义对象类型
- **类型别名**：使用 type 定义联合类型、工具类型等
- **命名规范**：接口和类型使用 PascalCase
- **导出类型**：需要复用的类型要导出

### 类型注解
- **函数参数**：明确标注参数类型
- **函数返回值**：明确标注返回值类型
- **变量声明**：合理使用类型推断，必要时显式注解
- **泛型使用**：充分利用泛型提高复用性

### 严格模式
- **启用严格模式**：`strict: true`
- **noImplicitAny**：禁止隐式 any
- **strictNullChecks**：严格空值检查
- **strictFunctionTypes**：严格函数类型检查

## 最佳实践

### 类型安全
- **避免 any**：尽量使用 unknown 代替 any
- **类型守卫**：使用类型守卫进行运行时检查
- **类型断言**：谨慎使用类型断言
- **非空断言**：避免使用 ! 非空断言

### 代码组织
- **类型文件**：类型定义集中管理
- **模块导出**：合理组织类型导出
- **类型复用**：充分利用工具类型
- **声明合并**：合理使用声明合并

### 性能优化
- **类型推断**：合理利用类型推断，避免过度注解
- **条件类型**：避免复杂的条件类型
- **类型缓存**：注意类型计算的性能影响

## 常用工具类型

### 内置工具类型
```typescript
// Partial - 所有属性可选
type PartialUser = Partial<User>

// Required - 所有属性必填
type RequiredUser = Required<User>

// Readonly - 所有属性只读
type ReadonlyUser = Readonly<User>

// Pick - 选取部分属性
type UserName = Pick<User, 'name' | 'email'>

// Omit - 排除部分属性
type UserWithoutPassword = Omit<User, 'password'>

// Record - 键值对类型
type UserMap = Record<number, User>

// Exclude - 排除类型
type NonNullable = Exclude<string | null | undefined, null | undefined>

// Extract - 提取类型
type StringOrNumber = Extract<string | number | boolean, string | number>
```

### 自定义工具类型
```typescript
//  Maybe - 可能为 null 或 undefined
type Maybe<T> = T | null | undefined

//  AsyncResult - 异步操作结果
type AsyncResult<T> = Promise<{ data: T; error: null } | { data: null; error: Error }>

//  DeepPartial - 深度可选
type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P]
}
```

## 学习资源

- **官方文档**：https://www.typescriptlang.org/docs/
- **中文文档**：https://www.tslang.cn/docs/home.html
- **TypeScript 手册**：https://www.typescriptlang.org/docs/handbook/intro.html
- **TypeScript Playground**：https://www.typescriptlang.org/play
- **GitHub 仓库**：https://github.com/microsoft/TypeScript
