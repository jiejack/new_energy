import router from './index'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permission'
import { getToken } from '@/utils/auth'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'

// 配置NProgress
NProgress.configure({ showSpinner: false })

// 白名单路由
const whiteList = ['/login', '/404']

/**
 * 路由前置守卫
 */
router.beforeEach(async (to, from, next) => {
  // 开始进度条
  NProgress.start()

  // 获取token
  const hasToken = getToken()

  if (hasToken) {
    if (to.path === '/login') {
      // 已登录，跳转到首页
      next({ path: '/' })
      NProgress.done()
    } else {
      // 检查是否已获取用户信息
      const userStore = useUserStore()
      const hasRoles = userStore.roles && userStore.roles.length > 0

      if (hasRoles) {
        // 已有用户信息，直接放行
        next()
      } else {
        try {
          // 获取用户信息
          const { roles } = await userStore.getUserInfo()

          // 根据角色生成可访问路由
          const permissionStore = usePermissionStore()
          const accessRoutes = await permissionStore.generateRoutes(roles)

          // 动态添加路由
          accessRoutes.forEach((route) => {
            router.addRoute(route)
          })

          // 确保路由已添加完成
          next({ ...to, replace: true })
        } catch (error) {
          // 获取用户信息失败，清除token并跳转到登录页
          await userStore.logout()
          next(`/login?redirect=${to.path}`)
          NProgress.done()
        }
      }
    }
  } else {
    // 未登录
    if (whiteList.includes(to.path)) {
      // 在白名单中，直接放行
      next()
    } else {
      // 不在白名单中，跳转到登录页
      next(`/login?redirect=${to.path}`)
      NProgress.done()
    }
  }
})

/**
 * 路由后置守卫
 */
router.afterEach((to) => {
  // 结束进度条
  NProgress.done()

  // 设置页面标题
  document.title = to.meta.title ? `${to.meta.title} - 新能源监控系统` : '新能源监控系统'
})

export default router
