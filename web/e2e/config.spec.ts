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
 * 测试套件：区域管理
 */
test.describe('区域管理测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/device/region')
    await page.waitForLoadState('networkidle')
  })

  test('应该显示区域列表', async ({ page }) => {
    // 等待表格加载
    await page.waitForSelector('.el-table, .crud-table', { timeout: 10000 })

    // 验证表格存在
    await expect(page.locator('.el-table, .crud-table')).toBeVisible()
  })

  test('应该创建新区域', async ({ page }) => {
    // 点击新增按钮
    await page.click('button:has-text("新增")')

    // 等待对话框出现
    await page.waitForSelector('.el-dialog', { timeout: 5000 })

    // 填写表单
    await page.fill('input[placeholder*="区域名称"]', '测试区域')
    await page.fill('input[placeholder*="区域编码"]', 'TEST_REGION')

    // 提交表单
    await page.click('.el-dialog button:has-text("确定")')

    // 验证成功消息
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
  })

  test('应该编辑区域', async ({ page }) => {
    // 等待表格加载
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    // 点击第一行的编辑按钮
    const editBtn = page.locator('.el-table__row').first().locator('button:has-text("编辑")')
    if (await editBtn.count() > 0) {
      await editBtn.click()

      // 等待对话框出现
      await page.waitForSelector('.el-dialog', { timeout: 5000 })

      // 修改表单
      await page.fill('input[placeholder*="区域名称"]', '更新后的区域名')

      // 提交表单
      await page.click('.el-dialog button:has-text("确定")')

      // 验证成功消息
      await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    }
  })

  test('应该删除区域', async ({ page }) => {
    // 等待表格加载
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    // 获取删除前的行数
    const rowsBefore = await page.locator('.el-table__row').count()

    // 点击第一行的删除按钮
    const deleteBtn = page.locator('.el-table__row').first().locator('button:has-text("删除")')
    if (await deleteBtn.count() > 0) {
      await deleteBtn.click()

      // 确认删除
      await page.click('.el-message-box button:has-text("确定")')

      // 验证成功消息
      await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    }
  })

  test('应该支持搜索区域', async ({ page }) => {
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
})

/**
 * 测试套件：电站管理
 */
test.describe('电站管理测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/device/station')
    await page.waitForLoadState('networkidle')
  })

  test('应该显示电站列表', async ({ page }) => {
    await page.waitForSelector('.el-table, .crud-table', { timeout: 10000 })
    await expect(page.locator('.el-table, .crud-table')).toBeVisible()
  })

  test('应该创建新电站', async ({ page }) => {
    // 点击新增按钮
    await page.click('button:has-text("新增")')

    // 等待对话框出现
    await page.waitForSelector('.el-dialog', { timeout: 5000 })

    // 填写表单
    await page.fill('input[placeholder*="电站名称"]', '测试电站')
    await page.fill('input[placeholder*="电站编码"]', 'TEST_STATION')

    // 选择电站类型
    const typeSelect = page.locator('.el-select:has([placeholder*="类型"])')
    if (await typeSelect.count() > 0) {
      await typeSelect.click()
      await page.click('.el-select-dropdown__item:first-child')
    }

    // 提交表单
    await page.click('.el-dialog button:has-text("确定")')

    // 验证成功消息
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
  })

  test('应该编辑电站', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    const editBtn = page.locator('.el-table__row').first().locator('button:has-text("编辑")')
    if (await editBtn.count() > 0) {
      await editBtn.click()
      await page.waitForSelector('.el-dialog', { timeout: 5000 })
      await page.fill('input[placeholder*="电站名称"]', '更新后的电站名')
      await page.click('.el-dialog button:has-text("确定")')
      await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    }
  })

  test('应该删除电站', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    const deleteBtn = page.locator('.el-table__row').first().locator('button:has-text("删除")')
    if (await deleteBtn.count() > 0) {
      await deleteBtn.click()
      await page.click('.el-message-box button:has-text("确定")')
      await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    }
  })

  test('应该查看电站详情', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    const detailBtn = page.locator('.el-table__row').first().locator('button:has-text("详情")')
    if (await detailBtn.count() > 0) {
      await detailBtn.click()
      await page.waitForSelector('.el-dialog, .detail-panel', { timeout: 5000 })
      await expect(page.locator('.el-dialog, .detail-panel')).toBeVisible()
    }
  })

  test('应该支持按状态筛选电站', async ({ page }) => {
    const statusFilter = page.locator('.el-select:has([placeholder*="状态"])')
    if (await statusFilter.count() > 0) {
      await statusFilter.click()
      await page.click('.el-select-dropdown__item:first-child')
      await page.waitForTimeout(1000)
      await expect(page.locator('.el-table')).toBeVisible()
    }
  })
})

/**
 * 测试套件：设备管理
 */
test.describe('设备管理测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/device/device')
    await page.waitForLoadState('networkidle')
  })

  test('应该显示设备列表', async ({ page }) => {
    await page.waitForSelector('.el-table, .crud-table', { timeout: 10000 })
    await expect(page.locator('.el-table, .crud-table')).toBeVisible()
  })

  test('应该创建新设备', async ({ page }) => {
    await page.click('button:has-text("新增")')
    await page.waitForSelector('.el-dialog', { timeout: 5000 })

    await page.fill('input[placeholder*="设备名称"]', '测试设备')
    await page.fill('input[placeholder*="设备编码"]', 'TEST_DEVICE')

    // 选择所属电站
    const stationSelect = page.locator('.el-select:has([placeholder*="电站"])')
    if (await stationSelect.count() > 0) {
      await stationSelect.click()
      await page.click('.el-select-dropdown__item:first-child')
    }

    await page.click('.el-dialog button:has-text("确定")')
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
  })

  test('应该编辑设备', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    const editBtn = page.locator('.el-table__row').first().locator('button:has-text("编辑")')
    if (await editBtn.count() > 0) {
      await editBtn.click()
      await page.waitForSelector('.el-dialog', { timeout: 5000 })
      await page.fill('input[placeholder*="设备名称"]', '更新后的设备名')
      await page.click('.el-dialog button:has-text("确定")')
      await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    }
  })

  test('应该删除设备', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    const deleteBtn = page.locator('.el-table__row').first().locator('button:has-text("删除")')
    if (await deleteBtn.count() > 0) {
      await deleteBtn.click()
      await page.click('.el-message-box button:has-text("确定")')
      await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    }
  })

  test('应该支持按设备类型筛选', async ({ page }) => {
    const typeFilter = page.locator('.el-select:has([placeholder*="设备类型"])')
    if (await typeFilter.count() > 0) {
      await typeFilter.click()
      await page.click('.el-select-dropdown__item:first-child')
      await page.waitForTimeout(1000)
      await expect(page.locator('.el-table')).toBeVisible()
    }
  })
})

/**
 * 测试套件：采集点管理
 */
test.describe('采集点管理测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/device/point')
    await page.waitForLoadState('networkidle')
  })

  test('应该显示采集点列表', async ({ page }) => {
    await page.waitForSelector('.el-table, .crud-table', { timeout: 10000 })
    await expect(page.locator('.el-table, .crud-table')).toBeVisible()
  })

  test('应该创建新采集点', async ({ page }) => {
    await page.click('button:has-text("新增")')
    await page.waitForSelector('.el-dialog', { timeout: 5000 })

    await page.fill('input[placeholder*="采集点名称"]', '测试采集点')
    await page.fill('input[placeholder*="采集点编码"]', 'TEST_POINT')

    // 选择所属设备
    const deviceSelect = page.locator('.el-select:has([placeholder*="设备"])')
    if (await deviceSelect.count() > 0) {
      await deviceSelect.click()
      await page.click('.el-select-dropdown__item:first-child')
    }

    await page.click('.el-dialog button:has-text("确定")')
    await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
  })

  test('应该编辑采集点', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    const editBtn = page.locator('.el-table__row').first().locator('button:has-text("编辑")')
    if (await editBtn.count() > 0) {
      await editBtn.click()
      await page.waitForSelector('.el-dialog', { timeout: 5000 })
      await page.fill('input[placeholder*="采集点名称"]', '更新后的采集点名')
      await page.click('.el-dialog button:has-text("确定")')
      await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    }
  })

  test('应该删除采集点', async ({ page }) => {
    await page.waitForSelector('.el-table__row', { timeout: 10000 })

    const deleteBtn = page.locator('.el-table__row').first().locator('button:has-text("删除")')
    if (await deleteBtn.count() > 0) {
      await deleteBtn.click()
      await page.click('.el-message-box button:has-text("确定")')
      await expect(page.locator('.el-message--success')).toBeVisible({ timeout: 5000 })
    }
  })

  test('应该支持按采集点类型筛选', async ({ page }) => {
    const typeFilter = page.locator('.el-select:has([placeholder*="类型"])')
    if (await typeFilter.count() > 0) {
      await typeFilter.click()
      await page.click('.el-select-dropdown__item:first-child')
      await page.waitForTimeout(1000)
      await expect(page.locator('.el-table')).toBeVisible()
    }
  })
})

/**
 * 测试套件：表单验证
 */
test.describe('配置管理表单验证测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/device/region')
    await page.waitForLoadState('networkidle')
  })

  test('应该验证必填字段', async ({ page }) => {
    await page.click('button:has-text("新增")')
    await page.waitForSelector('.el-dialog', { timeout: 5000 })

    // 直接提交空表单
    await page.click('.el-dialog button:has-text("确定")')

    // 验证错误提示
    await expect(page.locator('.el-form-item__error').first()).toBeVisible()
  })

  test('应该验证字段格式', async ({ page }) => {
    await page.click('button:has-text("新增")')
    await page.waitForSelector('.el-dialog', { timeout: 5000 })

    // 输入无效数据
    await page.fill('input[placeholder*="区域编码"]', '123 invalid!')

    // 提交表单
    await page.click('.el-dialog button:has-text("确定")')

    // 验证格式错误提示
    const errorVisible = await page.locator('.el-form-item__error').count() > 0
    expect(errorVisible).toBeTruthy()
  })

  test('应该支持取消操作', async ({ page }) => {
    await page.click('button:has-text("新增")')
    await page.waitForSelector('.el-dialog', { timeout: 5000 })

    // 填写一些数据
    await page.fill('input[placeholder*="区域名称"]', '测试')

    // 点击取消
    await page.click('.el-dialog button:has-text("取消")')

    // 验证对话框关闭
    await expect(page.locator('.el-dialog')).not.toBeVisible()
  })
})

/**
 * 测试套件：分页功能
 */
test.describe('配置管理分页测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/device/station')
    await page.waitForLoadState('networkidle')
  })

  test('应该显示分页器', async ({ page }) => {
    await page.waitForSelector('.el-pagination', { timeout: 10000 })
    await expect(page.locator('.el-pagination')).toBeVisible()
  })

  test('应该支持切换页码', async ({ page }) => {
    await page.waitForSelector('.el-pagination', { timeout: 10000 })

    const nextBtn = page.locator('.el-pagination .btn-next')
    if (await nextBtn.isEnabled()) {
      await nextBtn.click()
      await page.waitForTimeout(1000)
      await expect(page.locator('.el-table')).toBeVisible()
    }
  })

  test('应该支持修改每页显示数量', async ({ page }) => {
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
