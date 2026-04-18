from playwright.sync_api import sync_playwright
import time

with sync_playwright() as p:
    # 启动浏览器
    browser = p.chromium.launch(headless=False)
    page = browser.new_page()
    
    # 导航到前端应用
    page.goto('http://localhost:3000/')
    
    # 等待页面加载完成
    page.wait_for_load_state('networkidle')
    
    # 截图保存
    page.screenshot(path='frontend-screenshot.png', full_page=True)
    print('已保存首页截图: frontend-screenshot.png')
    
    # 检查页面标题
    title = page.title()
    print(f'页面标题: {title}')
    
    # 检查是否存在登录表单
    if page.locator('form').count() > 0:
        print('找到登录表单')
        
        # 尝试输入登录信息
        page.fill('input[type="username"]', 'admin')
        page.fill('input[type="password"]', 'admin123')
        
        # 截图登录页面
        page.screenshot(path='login-page.png')
        print('已保存登录页面截图: login-page.png')
        
        # 点击登录按钮
        page.click('button[type="submit"]')
        
        # 等待登录完成
        page.wait_for_load_state('networkidle')
        
        # 截图登录后的页面
        page.screenshot(path='dashboard-page.png', full_page=True)
        print('已保存仪表盘页面截图: dashboard-page.png')
        
        # 检查是否登录成功
        if '仪表盘' in page.title():
            print('登录成功，已进入仪表盘')
        else:
            print('登录可能失败，当前页面标题:', page.title())
    else:
        print('未找到登录表单，可能已经登录或不需要登录')
        
    # 关闭浏览器
    browser.close()
    print('测试完成')