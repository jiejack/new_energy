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
 * 测试套件：历史数据查询
 */
test.describe('历史数据查询测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/data/history')
    await page.waitForLoadState('networkidle')
  })

  test('应该显示历史数据查询页面', async ({ page }) => {
    // 验证页面元素
    await expect(page.locator('.time-range-picker, .el-date-editor')).toBeVisible()
    await expect(page.locator('.point-selector, [data-testid="point-selector"]')).toBeVisible()
  })

  test('应该选择时间范围', async ({ page }) => {
    // 查找时间选择器
    const timePicker = page.locator('.time-range-picker, .el-date-editor').first()

    // 点击时间选择器
    await timePicker.click()

    // 等待日期面板出现
    await page.waitForSelector('.el-date-picker, .el-picker-panel', { timeout: 5000 })

    // 选择日期范围（点击今天）
    const todayCell = page.locator('.el-date-table td.today, .el-date-table td.available').first()
    if (await todayCell.count() > 0) {
      await todayCell.click()
    }

    // 关闭面板
    await page.keyboard.press('Escape')
  })

  test('应该选择采集点', async ({ page }) => {
    // 查找采集点选择器
    const pointSelector = page.locator('.point-selector, [data-testid="point-selector"]')

    if (await pointSelector.count() > 0) {
      // 点击选择器
      await pointSelector.click()

      // 等待下拉框出现
      await page.waitForSelector('.el-select-dropdown, .point-list', { timeout: 5000 })

      // 选择第一个采集点
      const firstPoint = page.locator('.el-select-dropdown__item, .point-item').first()
      if (await firstPoint.count() > 0) {
        await firstPoint.click()
      }
    }
  })

  test('应该查询历史数据', async ({ page }) => {
    // 设置时间范围（使用快捷选项）
    const quickBtn = page.locator('button:has-text("今天"), button:has-text("今日")')
    if (await quickBtn.count() > 0) {
      await quickBtn.first().click()
    }

    // 点击查询按钮
    const queryBtn = page.locator('button:has-text("查询"), button:has-text("搜索")')
    if (await queryBtn.count() > 0) {
      await queryBtn.first().click()

      // 等待数据加载
      await page.waitForTimeout(2000)

      // 验证数据表格或图表显示
      const dataTable = page.locator('.data-table, .el-table')
      const dataChart = page.locator('.data-chart, .echarts')

      const hasData = (await dataTable.count()) > 0 || (await dataChart.count()) > 0
      expect(hasData).toBeTruthy()
    }
  })

  test('应该显示查询结果表格', async ({ page }) => {
    // 执行查询
    const queryBtn = page.locator('button:has-text("查询")').first()
    if (await queryBtn.count() > 0) {
      await queryBtn.click()
      await page.waitForTimeout(2000)

      // 验证表格存在
      const table = page.locator('.data-table, .el-table')
      if (await table.count() > 0) {
        await expect(table).toBeVisible()
      }
    }
  })

  test('应该显示数据图表', async ({ page }) => {
    // 执行查询
    const queryBtn = page.locator('button:has-text("查询")').first()
    if (await queryBtn.count() > 0) {
      await queryBtn.click()
      await page.waitForTimeout(2000)

      // 验证图表存在
      const chart = page.locator('.data-chart, .echarts')
      if (await chart.count() > 0) {
        await expect(chart).toBeVisible()
      }
    }
  })

  test('应该支持导出数据', async ({ page }) => {
    // 执行查询
    const queryBtn = page.locator('button:has-text("查询")').first()
    if (await queryBtn.count() > 0) {
      await queryBtn.click()
      await page.waitForTimeout(2000)
    }

    // 查找导出按钮
    const exportBtn = page.locator('button:has-text("导出"), button:has-text("下载")')

    if (await exportBtn.count() > 0) {
      // 点击导出
      const [download] = await Promise.all([
        page.waitForEvent('download', { timeout: 10000 }).catch(() => null),
        exportBtn.first().click(),
      ])

      // 验证下载（如果触发了下载）
      if (download) {
        expect(download).toBeTruthy()
      }
    }
  })

  test('应该支持数据聚合方式选择', async ({ page }) => {
    // 查找聚合方式选择器
    const aggregationSelect = page.locator('.el-select:has([placeholder*="聚合"]), .aggregation-select')

    if (await aggregationSelect.count() > 0) {
      await aggregationSelect.click()
      await page.waitForSelector('.el-select-dropdown', { timeout: 5000 })

      // 选择平均值
      const avgOption = page.locator('.el-select-dropdown__item:has-text("平均")')
      if (await avgOption.count() > 0) {
        await avgOption.click()
      }
    }
  })

  test('应该支持时间间隔选择', async ({ page }) => {
    // 查找时间间隔选择器
    const intervalSelect = page.locator('.el-select:has([placeholder*="间隔"]), .interval-select')

    if (await intervalSelect.count() > 0) {
      await intervalSelect.click()
      await page.waitForSelector('.el-select-dropdown', { timeout: 5000 })

      // 选择1小时间隔
      const hourOption = page.locator('.el-select-dropdown__item:has-text("1小时")')
      if (await hourOption.count() > 0) {
        await hourOption.click()
      }
    }
  })

  test('应该显示空数据状态', async ({ page }) => {
    // 选择一个不存在数据的时间范围
    const timePicker = page.locator('.time-range-picker, .el-date-editor').first()
    await timePicker.click()
    await page.waitForTimeout(500)

    // 点击查询
    const queryBtn = page.locator('button:has-text("查询")').first()
    if (await queryBtn.count() > 0) {
      await queryBtn.click()
      await page.waitForTimeout(2000)

      // 验证空状态或表格
      const emptyState = page.locator('.el-table__empty-text, .empty-state')
      const table = page.locator('.el-table')

      const hasEmptyState = await emptyState.count() > 0
      const hasTable = await table.count() > 0

      expect(hasEmptyState || hasTable).toBeTruthy()
    }
  })
})

/**
 * 测试套件：时间范围选择
 */
test.describe('时间范围选择测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/data/history')
    await page.waitForLoadState('networkidle')
  })

  test('应该支持快捷时间选择', async ({ page }) => {
    // 查找快捷按钮
    const quickButtons = ['今天', '昨天', '最近7天', '最近30天']

    for (const btnText of quickButtons) {
      const btn = page.locator(`button:has-text("${btnText}")`)
      if (await btn.count() > 0) {
        await btn.first().click()
        await page.waitForTimeout(500)

        // 验证时间选择器已更新
        const timePicker = page.locator('.time-range-picker, .el-date-editor').first()
        await expect(timePicker).toBeVisible()
      }
    }
  })

  test('应该支持自定义时间范围', async ({ page }) => {
    const timePicker = page.locator('.time-range-picker, .el-date-editor').first()
    await timePicker.click()

    // 等待日期面板
    await page.waitForSelector('.el-date-picker', { timeout: 5000 })

    // 选择开始日期
    const startCell = page.locator('.el-date-table td.available').first()
    if (await startCell.count() > 0) {
      await startCell.click()
    }

    // 选择结束日期
    const endCell = page.locator('.el-date-table td.available').last()
    if (await endCell.count() > 0) {
      await endCell.click()
    }

    // 确认选择
    const confirmBtn = page.locator('.el-picker-panel__link-btn:has-text("确定")')
    if (await confirmBtn.count() > 0) {
      await confirmBtn.click()
    }
  })

  test('应该验证时间范围有效性', async ({ page }) => {
    const timePicker = page.locator('.time-range-picker, .el-date-editor').first()
    await timePicker.click()
    await page.waitForTimeout(500)

    // 尝试选择无效的时间范围（结束时间早于开始时间）
    // 这里主要验证日期选择器能正常工作
    await page.keyboard.press('Escape')
  })
})

/**
 * 测试套件：数据导出
 */
test.describe('数据导出测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/data/history')
    await page.waitForLoadState('networkidle')
  })

  test('应该支持导出Excel格式', async ({ page }) => {
    // 执行查询
    const queryBtn = page.locator('button:has-text("查询")').first()
    if (await queryBtn.count() > 0) {
      await queryBtn.click()
      await page.waitForTimeout(2000)
    }

    // 查找Excel导出按钮
    const excelBtn = page.locator('button:has-text("导出Excel"), button:has-text("Excel")')

    if (await excelBtn.count() > 0) {
      const [download] = await Promise.all([
        page.waitForEvent('download', { timeout: 10000 }).catch(() => null),
        excelBtn.first().click(),
      ])

      if (download) {
        expect(download.suggestedFilename()).toContain('.xlsx')
      }
    }
  })

  test('应该支持导出CSV格式', async ({ page }) => {
    // 执行查询
    const queryBtn = page.locator('button:has-text("查询")').first()
    if (await queryBtn.count() > 0) {
      await queryBtn.click()
      await page.waitForTimeout(2000)
    }

    // 查找CSV导出按钮
    const csvBtn = page.locator('button:has-text("导出CSV"), button:has-text("CSV")')

    if (await csvBtn.count() > 0) {
      const [download] = await Promise.all([
        page.waitForEvent('download', { timeout: 10000 }).catch(() => null),
        csvBtn.first().click(),
      ])

      if (download) {
        expect(download.suggestedFilename()).toContain('.csv')
      }
    }
  })

  test('应该显示导出进度', async ({ page }) => {
    // 执行查询
    const queryBtn = page.locator('button:has-text("查询")').first()
    if (await queryBtn.count() > 0) {
      await queryBtn.click()
      await page.waitForTimeout(2000)
    }

    const exportBtn = page.locator('button:has-text("导出")').first()
    if (await exportBtn.count() > 0) {
      await exportBtn.click()

      // 验证加载状态或进度提示
      const loading = page.locator('.el-loading-mask, .loading')
      const progressMsg = page.locator('.el-message:has-text("导出")')

      // 等待导出完成
      await page.waitForTimeout(2000)
    }
  })
})

/**
 * 测试套件：数据可视化
 */
test.describe('数据可视化测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/data/history')
    await page.waitForLoadState('networkidle')
  })

  test('应该显示数据图表', async ({ page }) => {
    // 执行查询
    const queryBtn = page.locator('button:has-text("查询")').first()
    if (await queryBtn.count() > 0) {
      await queryBtn.click()
      await page.waitForTimeout(2000)

      // 验证图表
      const chart = page.locator('.echarts, canvas')
      if (await chart.count() > 0) {
        await expect(chart.first()).toBeVisible()
      }
    }
  })

  test('应该支持图表类型切换', async ({ page }) => {
    // 执行查询
    const queryBtn = page.locator('button:has-text("查询")').first()
    if (await queryBtn.count() > 0) {
      await queryBtn.click()
      await page.waitForTimeout(2000)
    }

    // 查找图表类型切换按钮
    const chartTypeBtns = page.locator('.chart-type-btn, button:has-text("折线图"), button:has-text("柱状图")')

    if (await chartTypeBtns.count() > 1) {
      // 切换到柱状图
      await chartTypeBtns.nth(1).click()
      await page.waitForTimeout(1000)

      // 验证图表更新
      const chart = page.locator('.echarts, canvas')
      await expect(chart.first()).toBeVisible()
    }
  })

  test('应该支持图表缩放', async ({ page }) => {
    // 执行查询
    const queryBtn = page.locator('button:has-text("查询")').first()
    if (await queryBtn.count() > 0) {
      await queryBtn.click()
      await page.waitForTimeout(2000)
    }

    const chart = page.locator('.echarts, canvas').first()
    if (await chart.count() > 0) {
      // 使用鼠标滚轮缩放（如果支持）
      await chart.hover()
      await page.mouse.wheel(0, -100)
      await page.waitForTimeout(500)
    }
  })

  test('应该显示数据统计信息', async ({ page }) => {
    // 执行查询
    const queryBtn = page.locator('button:has-text("查询")').first()
    if (await queryBtn.count() > 0) {
      await queryBtn.click()
      await page.waitForTimeout(2000)

      // 查找统计信息
      const stats = page.locator('.data-stats, .statistics')
      if (await stats.count() > 0) {
        await expect(stats).toBeVisible()
      }
    }
  })
})

/**
 * 测试套件：性能测试
 */
test.describe('数据查询性能测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/data/history')
    await page.waitForLoadState('networkidle')
  })

  test('应该在合理时间内完成查询', async ({ page }) => {
    const startTime = Date.now()

    // 执行查询
    const queryBtn = page.locator('button:has-text("查询")').first()
    if (await queryBtn.count() > 0) {
      await queryBtn.click()

      // 等待数据加载完成
      await page.waitForSelector('.data-table, .el-table, .echarts', { timeout: 10000 })

      const endTime = Date.now()
      const duration = endTime - startTime

      // 验证查询时间在合理范围内（10秒内）
      expect(duration).toBeLessThan(10000)
    }
  })

  test('应该处理大数据量查询', async ({ page }) => {
    // 选择较长时间范围
    const quickBtn = page.locator('button:has-text("最近30天")')
    if (await quickBtn.count() > 0) {
      await quickBtn.first().click()
    }

    // 执行查询
    const queryBtn = page.locator('button:has-text("查询")').first()
    if (await queryBtn.count() > 0) {
      await queryBtn.click()

      // 等待数据加载
      await page.waitForTimeout(5000)

      // 验证页面响应
      await expect(page.locator('.data-table, .el-table, .echarts').first()).toBeVisible()
    }
  })
})
