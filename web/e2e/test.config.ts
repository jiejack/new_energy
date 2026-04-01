import { test as base } from '@playwright/test'

/**
 * 扩展测试配置
 */
export const test = base.extend({
  // 自动登录
  autoLogin: [true, { option: true }],

  // 测试前自动登录
  page: async ({ page, autoLogin }, use) => {
    if (autoLogin) {
      // 模拟登录状态
      await page.goto('/login')
      await page.evaluate(() => {
        localStorage.setItem('nem_token', 'test-token')
        localStorage.setItem('nem_refresh_token', 'test-refresh-token')
      })
    }
    await use(page)
  },
})

/**
 * 测试数据Mock
 */
export const mockData = {
  // 用户信息
  user: {
    id: 1,
    username: 'admin',
    nickname: '管理员',
    email: 'admin@example.com',
    phone: '13800138000',
    avatar: '',
    status: 1,
    roles: ['admin'],
    permissions: ['*'],
    createdAt: '2024-01-01 00:00:00',
    updatedAt: '2024-01-01 00:00:00',
  },

  // 登录响应
  loginResponse: {
    token: 'test-token',
    refreshToken: 'test-refresh-token',
    expiresIn: 7200,
    user: {
      id: 1,
      username: 'admin',
      nickname: '管理员',
      email: 'admin@example.com',
      phone: '13800138000',
      avatar: '',
      status: 1,
      roles: ['admin'],
      permissions: ['*'],
      createdAt: '2024-01-01 00:00:00',
      updatedAt: '2024-01-01 00:00:00',
    },
  },

  // 区域列表
  regions: [
    {
      id: 1,
      name: '华东区域',
      code: 'HD',
      parentId: null,
      level: 1,
      status: 1,
      description: '华东区域',
      createdAt: '2024-01-01 00:00:00',
      updatedAt: '2024-01-01 00:00:00',
    },
    {
      id: 2,
      name: '华北区域',
      code: 'HB',
      parentId: null,
      level: 1,
      status: 1,
      description: '华北区域',
      createdAt: '2024-01-01 00:00:00',
      updatedAt: '2024-01-01 00:00:00',
    },
  ],

  // 电站列表
  stations: [
    {
      id: 1,
      name: '测试光伏电站1',
      code: 'GF001',
      regionId: 1,
      regionName: '华东区域',
      type: 'solar',
      capacity: 100,
      status: 'online',
      address: '江苏省南京市',
      longitude: 118.78,
      latitude: 32.07,
      description: '测试光伏电站',
      createdAt: '2024-01-01 00:00:00',
      updatedAt: '2024-01-01 00:00:00',
    },
    {
      id: 2,
      name: '测试风电场1',
      code: 'WF001',
      regionId: 2,
      regionName: '华北区域',
      type: 'wind',
      capacity: 200,
      status: 'online',
      address: '河北省张家口市',
      longitude: 114.89,
      latitude: 40.77,
      description: '测试风电场',
      createdAt: '2024-01-01 00:00:00',
      updatedAt: '2024-01-01 00:00:00',
    },
  ],

  // 设备列表
  devices: [
    {
      id: 1,
      name: '逆变器1',
      code: 'INV001',
      stationId: 1,
      stationName: '测试光伏电站1',
      type: 'inverter',
      model: 'SUN2000-100KTL',
      manufacturer: '华为',
      status: 'online',
      installDate: '2024-01-01',
      lastMaintenanceDate: '2024-01-01',
      description: '测试逆变器',
      createdAt: '2024-01-01 00:00:00',
      updatedAt: '2024-01-01 00:00:00',
    },
  ],

  // 采集点列表
  points: [
    {
      id: 1,
      name: '有功功率',
      code: 'P',
      deviceId: 1,
      deviceName: '逆变器1',
      type: 'analog',
      unit: 'kW',
      dataType: 'float',
      minValue: 0,
      maxValue: 100,
      description: '有功功率',
      createdAt: '2024-01-01 00:00:00',
      updatedAt: '2024-01-01 00:00:00',
    },
  ],

  // 告警列表
  alarms: [
    {
      id: 1,
      title: '逆变器温度过高',
      content: '逆变器温度超过设定阈值',
      level: 'major',
      status: 'active',
      sourceType: 'device',
      sourceId: 1,
      sourceName: '逆变器1',
      occurredAt: '2024-01-01 10:00:00',
      acknowledgedAt: null,
      acknowledgedBy: null,
      resolvedAt: null,
      resolvedBy: null,
      createdAt: '2024-01-01 10:00:00',
      updatedAt: '2024-01-01 10:00:00',
    },
    {
      id: 2,
      title: '通信中断',
      content: '设备通信中断',
      level: 'critical',
      status: 'active',
      sourceType: 'device',
      sourceId: 2,
      sourceName: '逆变器2',
      occurredAt: '2024-01-01 11:00:00',
      acknowledgedAt: null,
      acknowledgedBy: null,
      resolvedAt: null,
      resolvedBy: null,
      createdAt: '2024-01-01 11:00:00',
      updatedAt: '2024-01-01 11:00:00',
    },
  ],

  // 历史数据
  historyData: {
    pointId: 1,
    pointName: '有功功率',
    unit: 'kW',
    data: Array.from({ length: 100 }, (_, i) => ({
      timestamp: new Date(Date.now() - i * 3600000).toISOString(),
      value: Math.random() * 100,
      quality: 1,
    })),
  },

  // 统计数据
  statistics: {
    totalStations: 10,
    onlineStations: 8,
    offlineStations: 2,
    totalCapacity: 1000,
    currentPower: 800,
    totalAlarms: 5,
    activeAlarms: 3,
    todayEnergy: 5000,
  },
}

/**
 * API Mock响应
 */
export const mockApiResponses = {
  // 登录
  '/auth/login': mockData.loginResponse,

  // 用户信息
  '/auth/user-info': mockData.user,

  // 区域列表
  '/region/list': {
    list: mockData.regions,
    total: mockData.regions.length,
    page: 1,
    pageSize: 10,
    totalPages: 1,
  },

  // 电站列表
  '/station/list': {
    list: mockData.stations,
    total: mockData.stations.length,
    page: 1,
    pageSize: 10,
    totalPages: 1,
  },

  // 设备列表
  '/device/list': {
    list: mockData.devices,
    total: mockData.devices.length,
    page: 1,
    pageSize: 10,
    totalPages: 1,
  },

  // 采集点列表
  '/point/list': {
    list: mockData.points,
    total: mockData.points.length,
    page: 1,
    pageSize: 10,
    totalPages: 1,
  },

  // 告警列表
  '/alarm/list': {
    list: mockData.alarms,
    total: mockData.alarms.length,
    page: 1,
    pageSize: 10,
    totalPages: 1,
  },

  // 历史数据
  '/data/history': mockData.historyData,

  // 统计数据
  '/dashboard/statistics': mockData.statistics,
}

/**
 * 设置API Mock
 */
export async function setupApiMocks(page: any) {
  // Mock所有API请求
  for (const [path, response] of Object.entries(mockApiResponses)) {
    await page.route(`**/api${path}`, (route: any) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          code: 200,
          message: 'success',
          data: response,
        }),
      })
    })
  }
}

/**
 * 清除API Mock
 */
export async function clearApiMocks(page: any) {
  await page.unrouteAll()
}

export { expect } from '@playwright/test'
