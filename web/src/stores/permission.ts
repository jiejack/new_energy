import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RouteRecordRaw } from 'vue-router'
import { asyncRoutes, constantRoutes } from '@/router'

/**
 * 判断用户是否有权限访问该路由
 */
function hasPermission(roles: string[], route: RouteRecordRaw): boolean {
  if (route.meta && route.meta.permissions) {
    return roles.some((role) => (route.meta?.permissions as string[])?.includes(role))
  }
  return true
}

/**
 * 递归过滤异步路由
 */
function filterAsyncRoutes(routes: RouteRecordRaw[], roles: string[]): RouteRecordRaw[] {
  const res: RouteRecordRaw[] = []
  routes.forEach((route) => {
    const tmp = { ...route }
    if (hasPermission(roles, tmp)) {
      if (tmp.children) {
        tmp.children = filterAsyncRoutes(tmp.children, roles)
      }
      res.push(tmp)
    }
  })
  return res
}

export const usePermissionStore = defineStore('permission', () => {
  // 状态
  const routes = ref<RouteRecordRaw[]>([])
  const addRoutes = ref<RouteRecordRaw[]>([])

  /**
   * 生成可访问路由
   */
  function generateRoutes(roles: string[]): Promise<RouteRecordRaw[]> {
    return new Promise((resolve) => {
      let accessedRoutes: RouteRecordRaw[]

      if (roles.includes('admin')) {
        // 管理员拥有所有权限
        accessedRoutes = asyncRoutes
      } else {
        // 根据角色过滤路由
        accessedRoutes = filterAsyncRoutes(asyncRoutes, roles)
      }

      addRoutes.value = accessedRoutes
      routes.value = constantRoutes.concat(accessedRoutes)
      resolve(accessedRoutes)
    })
  }

  /**
   * 重置路由
   */
  function resetRoutes() {
    routes.value = []
    addRoutes.value = []
  }

  return {
    routes,
    addRoutes,
    generateRoutes,
    resetRoutes,
  }
})
