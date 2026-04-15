import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

/**
 * 公共路由配置
 */
export const constantRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: {
      title: '登录',
      hidden: true,
    },
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: {
      title: '404',
      hidden: true,
    },
  },
]

/**
 * 需要权限的路由配置
 */
export const asyncRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/layouts/MainLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: {
          title: '仪表盘',
          icon: 'Odometer',
          requiresAuth: true,
        },
      },
      {
        path: 'monitor',
        name: 'Monitor',
        redirect: '/monitor/realtime',
        meta: {
          title: '实时监控',
          icon: 'Monitor',
          requiresAuth: true,
        },
        children: [
          {
            path: 'realtime',
            name: 'RealtimeMonitor',
            component: () => import('@/views/monitor/realtime.vue'),
            meta: {
              title: '实时数据',
              icon: 'DataLine',
              requiresAuth: true,
            },
          },
          {
            path: 'station',
            name: 'StationMonitor',
            component: () => import('@/views/monitor/station.vue'),
            meta: {
              title: '电站监控',
              icon: 'OfficeBuilding',
              requiresAuth: true,
            },
          },
        ],
      },
      {
        path: 'device',
        name: 'Device',
        redirect: '/device/station',
        meta: {
          title: '设备管理',
          icon: 'SetUp',
          requiresAuth: true,
        },
        children: [
          {
            path: 'region',
            name: 'RegionManage',
            component: () => import('@/views/device/region.vue'),
            meta: {
              title: '区域管理',
              icon: 'Location',
              requiresAuth: true,
            },
          },
          {
            path: 'station',
            name: 'StationManage',
            component: () => import('@/views/device/station.vue'),
            meta: {
              title: '电站管理',
              icon: 'OfficeBuilding',
              requiresAuth: true,
            },
          },
          {
            path: 'device',
            name: 'DeviceManage',
            component: () => import('@/views/device/device.vue'),
            meta: {
              title: '设备管理',
              icon: 'SetUp',
              requiresAuth: true,
            },
          },
          {
            path: 'point',
            name: 'PointManage',
            component: () => import('@/views/device/point.vue'),
            meta: {
              title: '采集点管理',
              icon: 'Connection',
              requiresAuth: true,
            },
          },
        ],
      },
      {
        path: 'alarm',
        name: 'Alarm',
        redirect: '/alarm/list',
        meta: {
          title: '告警管理',
          icon: 'Bell',
          requiresAuth: true,
        },
        children: [
          {
            path: 'list',
            name: 'AlarmList',
            component: () => import('@/views/alarm/list/index.vue'),
            meta: {
              title: '告警列表',
              icon: 'Warning',
              requiresAuth: true,
            },
          },
          {
            path: 'rule',
            name: 'AlarmRule',
            component: () => import('@/views/alarm/rule/index.vue'),
            meta: {
              title: '告警规则',
              icon: 'Setting',
              requiresAuth: true,
            },
          },
          {
            path: 'notification',
            name: 'AlarmNotification',
            component: () => import('@/views/alarm/notification/index.vue'),
            meta: {
              title: '通知配置',
              icon: 'BellFilled',
              requiresAuth: true,
            },
          },
        ],
      },
      {
        path: 'data',
        name: 'Data',
        redirect: '/data/history',
        meta: {
          title: '数据查询',
          icon: 'DataAnalysis',
          requiresAuth: true,
        },
        children: [
          {
            path: 'history',
            name: 'DataHistory',
            component: () => import('@/views/data/history/index.vue'),
            meta: {
              title: '历史数据',
              icon: 'Search',
              requiresAuth: true,
            },
          },
          {
            path: 'report',
            name: 'DataReport',
            component: () => import('@/views/data/report/index.vue'),
            meta: {
              title: '统计报表',
              icon: 'Document',
              requiresAuth: true,
            },
          },
        ],
      },
      {
        path: 'energy',
        name: 'Energy',
        redirect: '/energy/efficiency',
        meta: {
          title: '能源分析',
          icon: 'Lightning',
          requiresAuth: true,
        },
        children: [
          {
            path: 'efficiency',
            name: 'EnergyEfficiency',
            component: () => import('@/views/energy/efficiency/index.vue'),
            meta: {
              title: '能效分析',
              icon: 'TrendCharts',
              requiresAuth: true,
            },
          },
          {
            path: 'carbon',
            name: 'CarbonEmission',
            component: () => import('@/views/energy/carbon/index.vue'),
            meta: {
              title: '碳排放监测',
              icon: 'Opportunity',
              requiresAuth: true,
            },
          },
        ],
      },
      {
        path: 'system',
        name: 'System',
        redirect: '/system/user',
        meta: {
          title: '系统管理',
          icon: 'Tools',
          requiresAuth: true,
        },
        children: [
          {
            path: 'user',
            name: 'UserManage',
            component: () => import('@/views/system/user/index.vue'),
            meta: {
              title: '用户管理',
              icon: 'User',
              requiresAuth: true,
            },
          },
          {
            path: 'role',
            name: 'RoleManage',
            component: () => import('@/views/system/role/index.vue'),
            meta: {
              title: '角色管理',
              icon: 'UserFilled',
              requiresAuth: true,
            },
          },
          {
            path: 'permission',
            name: 'PermissionManage',
            component: () => import('@/views/system/permission/index.vue'),
            meta: {
              title: '权限管理',
              icon: 'Lock',
              requiresAuth: true,
            },
          },
          {
            path: 'log',
            name: 'OperationLog',
            component: () => import('@/views/system/log/index.vue'),
            meta: {
              title: '操作日志',
              icon: 'Tickets',
              requiresAuth: true,
            },
          },
          {
            path: 'settings',
            name: 'SystemSettings',
            component: () => import('@/views/system/settings/index.vue'),
            meta: {
              title: '系统设置',
              icon: 'Setting',
              requiresAuth: true,
            },
          },
        ],
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/profile/index.vue'),
        meta: {
          title: '个人设置',
          icon: 'UserFilled',
          requiresAuth: true,
          hidden: true,
        },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/404',
    meta: {
      hidden: true,
    },
  },
]

/**
 * 创建路由实例
 */
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [...constantRoutes, ...asyncRoutes],
  scrollBehavior: () => ({ top: 0 }),
})

/**
 * 重置路由
 */
export function resetRouter() {
  const newRouter = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [...constantRoutes, ...asyncRoutes],
  })
  // @ts-ignore - matcher is internal API
  router.matcher = newRouter.matcher
}

export default router
