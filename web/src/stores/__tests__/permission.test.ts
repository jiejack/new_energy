import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { usePermissionStore } from '../permission'

// Mock router
vi.mock('@/router', () => ({
  asyncRoutes: [
    {
      path: '/dashboard',
      name: 'Dashboard',
      meta: { permissions: ['dashboard:view'] },
      children: [
        {
          path: 'detail',
          name: 'DashboardDetail',
          meta: { permissions: ['dashboard:detail'] }
        }
      ]
    },
    {
      path: '/system',
      name: 'System',
      meta: { permissions: ['system:view'] },
      children: [
        {
          path: 'user',
          name: 'User',
          meta: { permissions: ['user:view'] }
        }
      ]
    }
  ],
  constantRoutes: [
    {
      path: '/login',
      name: 'Login'
    }
  ]
}))

describe('Permission Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('状态初始化', () => {
    it('应该初始化为空路由', () => {
      const store = usePermissionStore()

      expect(store.routes).toEqual([])
      expect(store.addRoutes).toEqual([])
    })
  })

  describe('generateRoutes', () => {
    it('管理员应该拥有所有路由', async () => {
      const store = usePermissionStore()
      const roles = ['admin']

      const result = await store.generateRoutes(roles)

      expect(result).toHaveLength(2)
      expect(store.addRoutes).toHaveLength(2)
      expect(store.routes).toHaveLength(3) // constantRoutes + asyncRoutes
    })

    it('普通用户应该根据权限过滤路由', async () => {
      const store = usePermissionStore()
      const roles = ['dashboard:view']

      const result = await store.generateRoutes(roles)

      expect(result).toHaveLength(1)
      expect(result[0].path).toBe('/dashboard')
      expect(store.addRoutes).toHaveLength(1)
    })

    it('没有权限的用户应该没有异步路由', async () => {
      const store = usePermissionStore()
      const roles = ['other:permission']

      const result = await store.generateRoutes(roles)

      expect(result).toHaveLength(0)
      expect(store.addRoutes).toHaveLength(0)
      expect(store.routes).toHaveLength(1) // 只有 constantRoutes
    })

    it('应该递归过滤子路由', async () => {
      const store = usePermissionStore()
      const roles = ['dashboard:view'] // 没有 dashboard:detail 权限

      const result = await store.generateRoutes(roles)

      expect(result).toHaveLength(1)
      expect(result[0].children).toHaveLength(0) // 子路由被过滤掉
    })

    it('应该保留有权限的子路由', async () => {
      const store = usePermissionStore()
      const roles = ['dashboard:view', 'dashboard:detail']

      const result = await store.generateRoutes(roles)

      expect(result).toHaveLength(1)
      expect(result[0].children).toHaveLength(1)
      expect(result[0].children![0].name).toBe('DashboardDetail')
    })
  })

  describe('resetRoutes', () => {
    it('应该重置路由状态', async () => {
      const store = usePermissionStore()

      // 先生成路由
      await store.generateRoutes(['admin'])
      expect(store.routes.length).toBeGreaterThan(0)
      expect(store.addRoutes.length).toBeGreaterThan(0)

      // 重置
      store.resetRoutes()

      expect(store.routes).toEqual([])
      expect(store.addRoutes).toEqual([])
    })
  })

  describe('路由权限判断', () => {
    it('没有meta的路由应该允许访问', async () => {
      const store = usePermissionStore()
      const roles = ['any:role']

      await store.generateRoutes(roles)

      // constantRoutes 中的 /login 没有 meta，应该始终可访问
      expect(store.routes).toHaveLength(1)
      expect(store.routes[0].path).toBe('/login')
    })

    it('多个角色应该合并权限', async () => {
      const store = usePermissionStore()
      const roles = ['dashboard:view', 'system:view']

      const result = await store.generateRoutes(roles)

      expect(result).toHaveLength(2)
    })
  })
})
