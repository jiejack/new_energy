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
 * 测试套件：登录流程
 */
test.describe('登录流程测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
  })

  test('应该显示登录页面', async ({ page }) => {
    // 验证页面标题
    await expect(page.locator('h2.title')).toHaveText('新能源监控系统')
    await expect(page.locator('p.subtitle')).toHaveText('New Energy Monitoring System')

    // 验证表单元素存在
    await expect(page.locator('input[placeholder="请输入用户名"]')).toBeVisible()
    await expect(page.locator('input[placeholder="请输入密码"]')).toBeVisible()
    await expect(page.locator('button:has-text("登录")')).toBeVisible()

    // 验证默认账号提示
    await expect(page.locator('.login-footer p')).toContainText('默认账号: admin / 123456')
  })

  test('应该成功登录', async ({ page }) => {
    // 填写登录表单
    await page.fill('input[placeholder="请输入用户名"]', 'admin')
    await page.fill('input[placeholder="请输入密码"]', '123456')

    // 点击登录按钮
    await page.click('button:has-text("登录")')

    // 验证跳转到仪表盘
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 })

    // 验证成功消息
    await expect(page.locator('.el-message--success')).toBeVisible()
  })

  test('应该显示登录失败提示', async ({ page }) => {
    // 填写错误的登录信息
    await page.fill('input[placeholder="请输入用户名"]', 'wronguser')
    await page.fill('input[placeholder="请输入密码"]', 'wrongpass')

    // 点击登录按钮
    await page.click('button:has-text("登录")')

    // 验证错误消息
    await expect(page.locator('.el-message--error')).toBeVisible()
  })

  test('应该验证必填字段', async ({ page }) => {
    // 直接点击登录按钮
    await page.click('button:has-text("登录")')

    // 验证表单验证错误
    await expect(page.locator('.el-form-item__error').first()).toBeVisible()
  })

  test('应该验证密码长度', async ({ page }) => {
    // 填写短密码
    await page.fill('input[placeholder="请输入用户名"]', 'admin')
    await page.fill('input[placeholder="请输入密码"]', '12345')
    await page.click('button:has-text("登录")')

    // 验证密码长度错误提示
    await expect(page.locator('.el-form-item__error')).toContainText('密码长度不能少于6位')
  })

  test('应该持久化Token', async ({ page, context }) => {
    // 登录
    await page.fill('input[placeholder="请输入用户名"]', 'admin')
    await page.fill('input[placeholder="请输入密码"]', '123456')
    await page.click('button:has-text("登录")')
    await page.waitForURL('**/dashboard', { timeout: 10000 })

    // 验证localStorage中有token
    const token = await page.evaluate(() => localStorage.getItem('nem_token'))
    expect(token).toBeTruthy()

    // 刷新页面
    await page.reload()

    // 验证仍然在登录状态
    await expect(page).toHaveURL(/.*dashboard/)
  })

  test('应该支持记住我功能', async ({ page }) => {
    // 勾选记住我
    await page.check('input[type="checkbox"]')

    // 登录
    await page.fill('input[placeholder="请输入用户名"]', 'admin')
    await page.fill('input[placeholder="请输入密码"]', '123456')
    await page.click('button:has-text("登录")')
    await page.waitForURL('**/dashboard', { timeout: 10000 })

    // 验证登录成功
    await expect(page).toHaveURL(/.*dashboard/)
  })

  test('应该支持回车键登录', async ({ page }) => {
    // 填写登录表单
    await page.fill('input[placeholder="请输入用户名"]', 'admin')
    await page.fill('input[placeholder="请输入密码"]', '123456')

    // 按回车键
    await page.press('input[placeholder="请输入密码"]', 'Enter')

    // 验证跳转到仪表盘
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 })
  })

  test('应该在登录后跳转到重定向地址', async ({ page }) => {
    // 访问需要权限的页面
    await page.goto('/alarm/list')

    // 验证跳转到登录页并带有重定向参数
    await expect(page).toHaveURL(/.*login.*redirect=%2Falarm%2Flist/)

    // 登录
    await page.fill('input[placeholder="请输入用户名"]', 'admin')
    await page.fill('input[placeholder="请输入密码"]', '123456')
    await page.click('button:has-text("登录")')

    // 验证跳转到原始页面
    await expect(page).toHaveURL(/.*alarm\/list/, { timeout: 10000 })
  })
})

/**
 * 测试套件：登出流程
 */
test.describe('登出流程测试', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('应该成功登出', async ({ page }) => {
    // 点击用户头像或菜单
    await page.click('.el-dropdown-link, .user-info')

    // 点击登出按钮
    await page.click('text=退出登录')

    // 验证跳转到登录页
    await expect(page).toHaveURL(/.*login/)

    // 验证token被清除
    const token = await page.evaluate(() => localStorage.getItem('nem_token'))
    expect(token).toBeNull()
  })
})
