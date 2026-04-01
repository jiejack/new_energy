import { Page } from '@playwright/test'

/**
 * 测试辅助函数
 */

/**
 * 登录函数
 */
export async function login(page: Page, username = 'admin', password = '123456') {
  await page.goto('/login')
  await page.fill('input[placeholder="请输入用户名"]', username)
  await page.fill('input[placeholder="请输入密码"]', password)
  await page.click('button:has-text("登录")')
  await page.waitForURL('**/dashboard', { timeout: 10000 })
}

/**
 * 登出函数
 */
export async function logout(page: Page) {
  await page.click('.el-dropdown-link, .user-info')
  await page.click('text=退出登录')
  await page.waitForURL('**/login', { timeout: 10000 })
}

/**
 * 等待表格加载
 */
export async function waitForTable(page: Page, timeout = 10000) {
  await page.waitForSelector('.el-table, .crud-table', { timeout })
}

/**
 * 等待对话框
 */
export async function waitForDialog(page: Page, timeout = 5000) {
  await page.waitForSelector('.el-dialog', { timeout })
}

/**
 * 关闭对话框
 */
export async function closeDialog(page: Page) {
  await page.click('.el-dialog button:has-text("取消")')
  await page.waitForSelector('.el-dialog', { state: 'hidden' })
}

/**
 * 确认对话框
 */
export async function confirmDialog(page: Page) {
  await page.click('.el-dialog button:has-text("确定")')
  await page.waitForSelector('.el-dialog', { state: 'hidden' })
}

/**
 * 确认消息框
 */
export async function confirmMessageBox(page: Page) {
  await page.click('.el-message-box button:has-text("确定")')
  await page.waitForSelector('.el-message-box', { state: 'hidden' })
}

/**
 * 取消消息框
 */
export async function cancelMessageBox(page: Page) {
  await page.click('.el-message-box button:has-text("取消")')
  await page.waitForSelector('.el-message-box', { state: 'hidden' })
}

/**
 * 等待成功消息
 */
export async function waitForSuccess(page: Page, timeout = 5000) {
  await page.waitForSelector('.el-message--success', { timeout })
}

/**
 * 等待错误消息
 */
export async function waitForError(page: Page, timeout = 5000) {
  await page.waitForSelector('.el-message--error', { timeout })
}

/**
 * 填写表单字段
 */
export async function fillForm(page: Page, fields: Record<string, string>) {
  for (const [placeholder, value] of Object.entries(fields)) {
    await page.fill(`input[placeholder*="${placeholder}"]`, value)
  }
}

/**
 * 选择下拉选项
 */
export async function selectOption(page: Page, placeholder: string, optionText: string) {
  const select = page.locator(`.el-select:has([placeholder*="${placeholder}"])`)
  await select.click()
  await page.waitForSelector('.el-select-dropdown', { timeout: 5000 })
  await page.click(`.el-select-dropdown__item:has-text("${optionText}")`)
}

/**
 * 检查元素是否可见
 */
export async function isVisible(page: Page, selector: string): Promise<boolean> {
  const element = page.locator(selector)
  return await element.isVisible()
}

/**
 * 获取表格行数
 */
export async function getTableRowCount(page: Page): Promise<number> {
  return await page.locator('.el-table__row').count()
}

/**
 * 点击表格行按钮
 */
export async function clickTableRowButton(page: Page, rowIndex: number, buttonText: string) {
  const row = page.locator('.el-table__row').nth(rowIndex)
  await row.locator(`button:has-text("${buttonText}")`).click()
}

/**
 * 等待加载完成
 */
export async function waitForLoading(page: Page, timeout = 10000) {
  await page.waitForSelector('.el-loading-mask', { state: 'hidden', timeout })
}

/**
 * 截图对比
 */
export async function takeScreenshot(page: Page, name: string) {
  await page.screenshot({ path: `test-results/${name}.png`, fullPage: true })
}

/**
 * 模拟网络延迟
 */
export async function simulateSlowNetwork(page: Page, delay = 1000) {
  await page.route('**/api/**', async (route) => {
    await new Promise((resolve) => setTimeout(resolve, delay))
    route.continue()
  })
}

/**
 * 模拟网络错误
 */
export async function simulateNetworkError(page: Page) {
  await page.route('**/api/**', (route) => route.abort())
}

/**
 * 模拟API响应
 */
export async function mockApiResponse(page: Page, path: string, response: any) {
  await page.route(`**/api/${path}`, (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(response),
    })
  })
}

/**
 * 清除localStorage
 */
export async function clearLocalStorage(page: Page) {
  await page.evaluate(() => localStorage.clear())
}

/**
 * 获取localStorage值
 */
export async function getLocalStorage(page: Page, key: string): Promise<string | null> {
  return await page.evaluate((k) => localStorage.getItem(k), key)
}

/**
 * 设置localStorage值
 */
export async function setLocalStorage(page: Page, key: string, value: string) {
  await page.evaluate(({ k, v }) => localStorage.setItem(k, v), { k: key, v: value })
}

/**
 * 等待WebSocket连接
 */
export async function waitForWebSocket(page: Page, timeout = 10000) {
  await page.waitForFunction(
    () => {
      return (window as any).__wsConnected === true
    },
    { timeout }
  )
}

/**
 * 检查响应式布局
 */
export async function checkResponsiveLayout(page: Page, width: number, height: number) {
  await page.setViewportSize({ width, height })
  await page.waitForTimeout(500)
}

/**
 * 测试数据生成器
 */
export const testData = {
  region: {
    name: `测试区域_${Date.now()}`,
    code: `TEST_REGION_${Date.now()}`,
    description: '测试区域描述',
  },
  station: {
    name: `测试电站_${Date.now()}`,
    code: `TEST_STATION_${Date.now()}`,
    type: 'solar',
    capacity: 100,
    address: '测试地址',
  },
  device: {
    name: `测试设备_${Date.now()}`,
    code: `TEST_DEVICE_${Date.now()}`,
    type: 'inverter',
    model: 'TEST-001',
    manufacturer: '测试厂商',
  },
  point: {
    name: `测试采集点_${Date.now()}`,
    code: `TEST_POINT_${Date.now()}`,
    type: 'analog',
    unit: 'kW',
  },
  user: {
    username: `testuser_${Date.now()}`,
    password: 'Test@123456',
    nickname: '测试用户',
    email: `test_${Date.now()}@example.com`,
    phone: '13800138000',
  },
}
