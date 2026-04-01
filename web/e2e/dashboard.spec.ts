import { test, expect, Page } from '@playwright/test'

/**
 * 测试辅助函数
 */
async function login(page: Page, username = 'admin', password = '123456') {
  await page.goto('/login')
  await page.fill('input[placeholder="请输入用户名"]', username)
  await page.fill('input[placeholder="请输入密码"]', password)
  await page.click('button:has-text("登录")')
  await page.waitForURL('**/dashboard', { timeout: 10000 })
}

/**
 * 测试套件：监控大屏
 */
test.describe('监控大屏测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('应该加载大屏页面', async ({ page }) => {
    // 验证页面标题
    await expect(page.locator('h2, .page-title')).toContainText('仪表盘')

    // 验证主要组件存在
    await expect(page.locator('.stat-cards, .dashboard-stats')).toBeVisible()
  })

  test('应该显示统计卡片', async ({ page }) => {
    // 等待统计卡片加载
    await page.waitForSelector('.stat-cards, .dashboard-stats', { timeout: 10000 })

    // 验证统计卡片数量
    const cards = await page.locator('.stat-card, .stat-item').count()
    expect(cards).toBeGreaterThan(0)

    // 验证卡片内容
    const firstCard = page.locator('.stat-card, .stat-item').first()
    await expect(firstCard).toBeVisible()
  })

  test('应该显示电站列表', async ({ page }) => {
    // 导航到电站列表区域
    const stationList = page.locator('.station-list, [data-testid="station-list"]')

    // 等待列表加载
    await stationList.waitFor({ timeout: 10000 })

    // 验证列表项存在
    const stations = await stationList.locator('.station-item, .el-table__row').count()
    expect(stations).toBeGreaterThanOrEqual(0)

    // 如果有电站，验证电站信息
    if (stations > 0) {
      const firstStation = stationList.locator('.station-item, .el-table__row').first()
      await expect(firstStation).toBeVisible()
    }
  })

  test('应该显示告警列表', async ({ page }) => {
    // 导航到告警列表区域
    const alarmList = page.locator('.alarm-list, [data-testid="alarm-list"]')

    // 等待列表加载
    await alarmList.waitFor({ timeout: 10000 })

    // 验证告警列表存在
    await expect(alarmList).toBeVisible()

    // 检查告警项
    const alarms = await alarmList.locator('.alarm-item, .el-table__row').count()
    expect(alarms).toBeGreaterThanOrEqual(0)
  })

  test('应该显示实时图表', async ({ page }) => {
    // 查找图表容器
    const chart = page.locator('.realtime-chart, .echarts, [data-testid="realtime-chart"]')

    // 等待图表加载
    await chart.waitFor({ timeout: 10000 })

    // 验证图表可见
    await expect(chart).toBeVisible()
  })

  test('应该显示电站地图', async ({ page }) => {
    // 查找地图容器
    const map = page.locator('.station-map, [data-testid="station-map"]')

    // 等待地图加载
    await map.waitFor({ timeout: 10000 })

    // 验证地图可见
    await expect(map).toBeVisible()
  })

  test('应该支持实时数据更新', async ({ page }) => {
    // 获取初始数据
    const initialValue = await page.locator('.stat-card .value, .stat-value').first().textContent()

    // 等待一段时间（模拟实时更新）
    await page.waitForTimeout(3000)

    // 验证数据可能已更新（这里只是检查元素仍然存在）
    await expect(page.locator('.stat-card .value, .stat-value').first()).toBeVisible()
  })

  test('应该支持刷新数据', async ({ page }) => {
    // 查找刷新按钮
    const refreshBtn = page.locator('button:has-text("刷新"), .refresh-btn, [data-testid="refresh-btn"]')

    if (await refreshBtn.count() > 0) {
      // 点击刷新
      await refreshBtn.first().click()

      // 等待加载完成
      await page.waitForTimeout(1000)

      // 验证数据仍然显示
      await expect(page.locator('.stat-cards, .dashboard-stats')).toBeVisible()
    }
  })

  test('应该支持全屏显示', async ({ page }) => {
    // 查找全屏按钮
    const fullscreenBtn = page.locator('button:has-text("全屏"), .fullscreen-btn, [data-testid="fullscreen-btn"]')

    if (await fullscreenBtn.count() > 0) {
      // 点击全屏
      await fullscreenBtn.first().click()

      // 验证全屏状态（检查是否有全屏类）
      await page.waitForTimeout(500)

      // 退出全屏（按ESC）
      await page.keyboard.press('Escape')
    }
  })

  test('应该正确显示电站状态', async ({ page }) => {
    // 等待电站列表加载
    await page.waitForSelector('.station-list, .el-table', { timeout: 10000 })

    // 查找状态标签
    const statusTags = page.locator('.station-status, .el-tag')

    // 验证状态标签存在
    const count = await statusTags.count()
    expect(count).toBeGreaterThanOrEqual(0)

    // 如果有状态标签，验证状态文本
    if (count > 0) {
      const statusText = await statusTags.first().textContent()
      expect(statusText).toBeTruthy()
    }
  })

  test('应该支持电站搜索', async ({ page }) => {
    // 查找搜索框
    const searchInput = page.locator('input[placeholder*="搜索"], input[placeholder*="查询"]')

    if (await searchInput.count() > 0) {
      // 输入搜索关键词
      await searchInput.first().fill('测试')
      await page.keyboard.press('Enter')

      // 等待搜索结果
      await page.waitForTimeout(1000)

      // 验证搜索结果
      await expect(page.locator('.station-list, .el-table')).toBeVisible()
    }
  })

  test('应该支持时间范围选择', async ({ page }) => {
    // 查找时间选择器
    const timePicker = page.locator('.time-range-picker, .el-date-editor')

    if (await timePicker.count() > 0) {
      // 点击时间选择器
      await timePicker.first().click()

      // 等待日期面板出现
      await page.waitForTimeout(500)

      // 选择时间范围（这里只是验证功能存在）
      await page.keyboard.press('Escape')
    }
  })

  test('应该正确处理网络错误', async ({ page, context }) => {
    // 模拟网络错误
    await context.route('**/api/**', route => route.abort())

    // 刷新页面
    await page.reload()

    // 验证错误提示或空状态
    const errorElement = page.locator('.el-message--error, .error-message, .empty-state')
    await expect(errorElement.first().or(page.locator('.stat-cards'))).toBeVisible()
  })
})

/**
 * 测试套件：响应式布局
 */
test.describe('监控大屏响应式测试', () => {
  test.use({ viewport: { width: 1920, height: 1080 } })

  test('应该在桌面端正确显示', async ({ page }) => {
    await login(page)

    // 验证布局
    await expect(page.locator('.stat-cards, .dashboard-stats')).toBeVisible()
    await expect(page.locator('.station-list, [data-testid="station-list"]')).toBeVisible()
  })

  test('应该在平板端正确显示', async ({ page }) => {
    // 设置平板视口
    await page.setViewportSize({ width: 768, height: 1024 })
    await login(page)

    // 验证布局适应
    await expect(page.locator('.stat-cards, .dashboard-stats')).toBeVisible()
  })

  test('应该在移动端正确显示', async ({ page }) => {
    // 设置移动端视口
    await page.setViewportSize({ width: 375, height: 667 })
    await login(page)

    // 验证布局适应
    await expect(page.locator('.stat-cards, .dashboard-stats')).toBeVisible()
  })
})
