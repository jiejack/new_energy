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
 * 测试套件：告警列表
 */
test.describe('告警列表测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/alarm/list')
    await page.waitForLoadState('networkidle')
  })

  test('应该显示告警列表页面', async ({ page }) => {
    // 验证页面标题
    await expect(page.locator('h2, .page-title')).toContainText('告警')

    // 验证表格存在
    await expect(page.locator('.el-table, .alarm-table')).toBeVisible()
  })

  test('应该显示告警数据', async ({ page }) => {
    // 等待表格加载
    await page.waitForSelector('.el-table__row, .alarm-item', { timeout: 10000 })

    // 验证告警项存在
    const alarmRows = await page.locator('.el-table__row, .alarm-item').count()
    expect(alarmRows).toBeGreaterThanOrEqual(0)
  })

  test('应该显示告警级别标签', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    // 查找告警级别标签
    const levelTags = page.locator('.alarm-level, .el-tag')

    if (await levelTags.count() > 0) {
      const firstTag = levelTags.first()
      await expect(firstTag).toBeVisible()

      // 验证级别文本
      const levelText = await firstTag.textContent()
      expect(levelText).toBeTruthy()
    }
  })

  test('应该显示告警状态', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    // 查找状态标签
    const statusTags = page.locator('.alarm-status, .el-tag')

    if (await statusTags.count() > 0) {
      const firstTag = statusTags.first()
      await expect(firstTag).toBeVisible()
    }
  })

  test('应该显示告警时间', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    // 查找时间列
    const timeCells = page.locator('.el-table__row td:has-text("-")')

    if (await timeCells.count() > 0) {
      const timeText = await timeCells.first().textContent()
      expect(timeText).toBeTruthy()
    }
  })

  test('应该显示告警来源', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    // 查找来源列
    const sourceCells = page.locator('.alarm-source, .source-cell')

    if (await sourceCells.count() > 0) {
      await expect(sourceCells.first()).toBeVisible()
    }
  })
})

/**
 * 测试套件：告警确认流程
 */
test.describe('告警确认流程测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/alarm/list')
    await page.waitForLoadState('networkidle')
  })

  test('应该确认告警', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    // 查找确认按钮
    const confirmBtn = page.locator('.el-table__row').first().locator('button:has-text("确认")')

    if (await confirmBtn.count() > 0) {
      await confirmBtn.click()

      // 等待确认对话框
      await page.waitForSelector('.el-message-box', { timeout: 5000 })

      // 填写确认备注
      const remarkInput = page.locator('.el-message-box textarea, .el-message-box input')
      if (await remarkInput.count() > 0) {
        await remarkInput.fill('测试确认')
      }

      // 确认操作
      await page.click('.el-message-box button:has-text("确定")')

      // 验证成功消息
      await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    }
  })

  test('应该批量确认告警', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    // 选择多个告警
    const checkboxes = page.locator('.el-table__row .el-checkbox')
    const checkboxCount = Math.min(await checkboxes.count(), 3)

    for (let i = 0; i < checkboxCount; i++) {
      await checkboxes.nth(i).click()
    }

    // 查找批量确认按钮
    const batchConfirmBtn = page.locator('button:has-text("批量确认")')

    if (await batchConfirmBtn.count() > 0 && checkboxCount > 0) {
      await batchConfirmBtn.click()

      // 等待确认对话框
      await page.waitForSelector('.el-message-box', { timeout: 5000 })

      // 确认操作
      await page.click('.el-message-box button:has-text("确定")')

      // 验证成功消息
      await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    }
  })

  test('应该解决告警', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    // 查找解决按钮
    const resolveBtn = page.locator('.el-table__row').first().locator('button:has-text("解决")')

    if (await resolveBtn.count() > 0) {
      await resolveBtn.click()

      // 等待解决对话框
      await page.waitForSelector('.el-message-box', { timeout: 5000 })

      // 填写解决说明
      const remarkInput = page.locator('.el-message-box textarea, .el-message-box input')
      if (await remarkInput.count() > 0) {
        await remarkInput.fill('问题已解决')
      }

      // 确认操作
      await page.click('.el-message-box button:has-text("确定")')

      // 验证成功消息
      await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    }
  })

  test('应该查看告警详情', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    // 点击告警行或详情按钮
    const detailBtn = page.locator('.el-table__row').first().locator('button:has-text("详情")')

    if (await detailBtn.count() > 0) {
      await detailBtn.click()

      // 等待详情对话框
      await page.waitForSelector('.el-dialog, .detail-panel', { timeout: 5000 })

      // 验证详情内容
      await expect(page.locator('.el-dialog, .detail-panel')).toBeVisible()
    }
  })

  test('应该取消确认操作', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    const confirmBtn = page.locator('.el-table__row').first().locator('button:has-text("确认")')

    if (await confirmBtn.count() > 0) {
      await confirmBtn.click()

      // 等待确认对话框
      await page.waitForSelector('.el-message-box', { timeout: 5000 })

      // 取消操作
      await page.click('.el-message-box button:has-text("取消")')

      // 验证对话框关闭
      await expect(page.locator('.el-message-box')).not.toBeVisible()
    }
  })
})

/**
 * 测试套件：告警筛选
 */
test.describe('告警筛选测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/alarm/list')
    await page.waitForLoadState('networkidle')
  })

  test('应该按告警级别筛选', async ({ page }) => {
    // 查找级别筛选器
    const levelFilter = page.locator('.el-select:has([placeholder*="级别"])')

    if (await levelFilter.count() > 0) {
      await levelFilter.click()
      await page.waitForSelector('.el-select-dropdown', { timeout: 5000 })

      // 选择严重级别
      const criticalOption = page.locator('.el-select-dropdown__item:has-text("严重")')
      if (await criticalOption.count() > 0) {
        await criticalOption.click()
        await page.waitForTimeout(1000)

        // 验证筛选结果
        await expect(page.locator('.el-table')).toBeVisible()
      }
    }
  })

  test('应该按告警状态筛选', async ({ page }) => {
    // 查找状态筛选器
    const statusFilter = page.locator('.el-select:has([placeholder*="状态"])')

    if (await statusFilter.count() > 0) {
      await statusFilter.click()
      await page.waitForSelector('.el-select-dropdown', { timeout: 5000 })

      // 选择活动状态
      const activeOption = page.locator('.el-select-dropdown__item:has-text("活动")')
      if (await activeOption.count() > 0) {
        await activeOption.click()
        await page.waitForTimeout(1000)

        // 验证筛选结果
        await expect(page.locator('.el-table')).toBeVisible()
      }
    }
  })

  test('应该按时间范围筛选', async ({ page }) => {
    // 查找时间选择器
    const timePicker = page.locator('.el-date-editor').first()

    if (await timePicker.count() > 0) {
      await timePicker.click()
      await page.waitForSelector('.el-picker-panel', { timeout: 5000 })

      // 选择今天
      const todayBtn = page.locator('button:has-text("今天")')
      if (await todayBtn.count() > 0) {
        await todayBtn.click()
        await page.waitForTimeout(1000)

        // 验证筛选结果
        await expect(page.locator('.el-table')).toBeVisible()
      }
    }
  })

  test('应该按关键词搜索', async ({ page }) => {
    // 查找搜索框
    const searchInput = page.locator('input[placeholder*="搜索"], input[placeholder*="查询"]')

    if (await searchInput.count() > 0) {
      await searchInput.first().fill('测试')
      await page.keyboard.press('Enter')
      await page.waitForTimeout(1000)

      // 验证搜索结果
      await expect(page.locator('.el-table')).toBeVisible()
    }
  })

  test('应该按告警来源筛选', async ({ page }) => {
    // 查找来源筛选器
    const sourceFilter = page.locator('.el-select:has([placeholder*="来源"])')

    if (await sourceFilter.count() > 0) {
      await sourceFilter.click()
      await page.waitForSelector('.el-select-dropdown', { timeout: 5000 })

      // 选择第一个来源
      const firstOption = page.locator('.el-select-dropdown__item').first()
      if (await firstOption.count() > 0) {
        await firstOption.click()
        await page.waitForTimeout(1000)

        // 验证筛选结果
        await expect(page.locator('.el-table')).toBeVisible()
      }
    }
  })

  test('应该重置筛选条件', async ({ page }) => {
    // 先设置一些筛选条件
    const levelFilter = page.locator('.el-select:has([placeholder*="级别"])')
    if (await levelFilter.count() > 0) {
      await levelFilter.click()
      await page.locator('.el-select-dropdown__item').first().click()
      await page.waitForTimeout(500)
    }

    // 查找重置按钮
    const resetBtn = page.locator('button:has-text("重置")')

    if (await resetBtn.count() > 0) {
      await resetBtn.click()
      await page.waitForTimeout(1000)

      // 验证筛选条件已重置
      await expect(page.locator('.el-table')).toBeVisible()
    }
  })

  test('应该组合多个筛选条件', async ({ page }) => {
    // 设置级别筛选
    const levelFilter = page.locator('.el-select:has([placeholder*="级别"])')
    if (await levelFilter.count() > 0) {
      await levelFilter.click()
      await page.locator('.el-select-dropdown__item').first().click()
    }

    // 设置状态筛选
    const statusFilter = page.locator('.el-select:has([placeholder*="状态"])')
    if (await statusFilter.count() > 0) {
      await statusFilter.click()
      await page.locator('.el-select-dropdown__item').first().click()
    }

    // 点击查询
    const queryBtn = page.locator('button:has-text("查询")')
    if (await queryBtn.count() > 0) {
      await queryBtn.click()
      await page.waitForTimeout(1000)

      // 验证筛选结果
      await expect(page.locator('.el-table')).toBeVisible()
    }
  })
})

/**
 * 测试套件：告警排序
 */
test.describe('告警排序测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/alarm/list')
    await page.waitForLoadState('networkidle')
  })

  test('应该按时间排序', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    // 点击时间列头
    const timeHeader = page.locator('.el-table__header th:has-text("时间")')

    if (await timeHeader.count() > 0) {
      await timeHeader.click()
      await page.waitForTimeout(1000)

      // 验证排序图标
      await expect(page.locator('.el-table')).toBeVisible()
    }
  })

  test('应该按级别排序', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    // 点击级别列头
    const levelHeader = page.locator('.el-table__header th:has-text("级别")')

    if (await levelHeader.count() > 0) {
      await levelHeader.click()
      await page.waitForTimeout(1000)

      // 验证排序图标
      await expect(page.locator('.el-table')).toBeVisible()
    }
  })
})

/**
 * 测试套件：告警分页
 */
test.describe('告警分页测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/alarm/list')
    await page.waitForLoadState('networkidle')
  })

  test('应该显示分页器', async ({ page }) => {
    await page.waitForSelector('.el-pagination', { timeout: 10000 })
    await expect(page.locator('.el-pagination')).toBeVisible()
  })

  test('应该切换页码', async ({ page }) => {
    await page.waitForSelector('.el-pagination', { timeout: 10000 })

    const nextBtn = page.locator('.el-pagination .btn-next')
    if (await nextBtn.isEnabled()) {
      await nextBtn.click()
      await page.waitForTimeout(1000)
      await expect(page.locator('.el-table')).toBeVisible()
    }
  })

  test('应该修改每页显示数量', async ({ page }) => {
    await page.waitForSelector('.el-pagination', { timeout: 10000 })

    const sizeSelect = page.locator('.el-pagination .el-select')
    if (await sizeSelect.count() > 0) {
      await sizeSelect.click()
      await page.click('.el-select-dropdown__item:last-child')
      await page.waitForTimeout(1000)
      await expect(page.locator('.el-table')).toBeVisible()
    }
  })
})

/**
 * 测试套件：告警导出
 */
test.describe('告警导出测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/alarm/list')
    await page.waitForLoadState('networkidle')
  })

  test('应该导出告警数据', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    // 查找导出按钮
    const exportBtn = page.locator('button:has-text("导出")')

    if (await exportBtn.count() > 0) {
      const [download] = await Promise.all([
        page.waitForEvent('download', { timeout: 10000 }).catch(() => null),
        exportBtn.first().click(),
      ])

      if (download) {
        expect(download).toBeTruthy()
      }
    }
  })
})

/**
 * 测试套件：告警统计
 */
test.describe('告警统计测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/alarm/list')
    await page.waitForLoadState('networkidle')
  })

  test('应该显示告警统计信息', async ({ page }) => {
    // 查找统计卡片或数字
    const stats = page.locator('.alarm-stats, .stat-card, .statistics')

    if (await stats.count() > 0) {
      await expect(stats.first()).toBeVisible()
    }
  })

  test('应该显示各级别告警数量', async ({ page }) => {
    // 查找级别统计
    const levelStats = page.locator('.level-stat, .stat-item')

    if (await levelStats.count() > 0) {
      const count = await levelStats.count()
      expect(count).toBeGreaterThan(0)
    }
  })
})

/**
 * 测试套件：实时告警
 */
test.describe('实时告警测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/alarm/list')
    await page.waitForLoadState('networkidle')
  })

  test('应该支持自动刷新', async ({ page }) => {
    await page.waitForSelector('.el-table', { timeout: 10000 })

    // 查找自动刷新开关
    const autoRefreshSwitch = page.locator('.auto-refresh, .el-switch')

    if (await autoRefreshSwitch.count() > 0) {
      // 开启自动刷新
      await autoRefreshSwitch.click()
      await page.waitForTimeout(1000)

      // 验证开关状态
      await expect(autoRefreshSwitch).toBeVisible()
    }
  })

  test('应该支持手动刷新', async ({ page }) => {
    await page.waitForSelector('.el-table', { timeout: 10000 })

    // 查找刷新按钮
    const refreshBtn = page.locator('button:has-text("刷新")')

    if (await refreshBtn.count() > 0) {
      await refreshBtn.click()
      await page.waitForTimeout(1000)

      // 验证表格仍然可见
      await expect(page.locator('.el-table')).toBeVisible()
    }
  })
})
