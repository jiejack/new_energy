from playwright.sync_api import sync_playwright
import time

# 测试前端应用
def test_frontend():
    with sync_playwright() as p:
        # 启动浏览器
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        
        print("=== 开始测试前端应用 ===")
        
        # 访问前端应用
        print("1. 访问前端应用: http://localhost:3001/")
        page.goto('http://localhost:3001/')
        
        # 等待页面加载完成
        print("2. 等待页面加载完成...")
        page.wait_for_load_state('networkidle')
        time.sleep(2)  # 额外等待确保所有内容加载
        
        # 捕获页面标题
        title = page.title()
        print(f"3. 页面标题: {title}")
        
        # 捕获页面截图
        screenshot_path = '/workspace/frontend-screenshot.png'
        print(f"4. 捕获页面截图: {screenshot_path}")
        page.screenshot(path=screenshot_path, full_page=True)
        
        # 检查页面元素
        print("5. 检查页面元素...")
        
        # 尝试找到导航栏
        nav_elements = page.locator('nav').all()
        if nav_elements:
            print("   ✓ 找到导航栏")
        else:
            print("   ✗ 未找到导航栏")
        
        # 尝试找到按钮
        buttons = page.locator('button').all()
        print(f"   ✓ 找到 {len(buttons)} 个按钮")
        
        # 尝试找到链接
        links = page.locator('a').all()
        print(f"   ✓ 找到 {len(links)} 个链接")
        
        # 捕获控制台日志
        print("6. 检查控制台日志...")
        console_logs = []
        page.on('console', lambda msg: console_logs.append(msg.text))
        
        # 等待几秒钟捕获日志
        time.sleep(3)
        
        if console_logs:
            print(f"   ✓ 捕获到 {len(console_logs)} 条控制台日志")
            print("   最近的5条日志:")
            for log in console_logs[-5:]:
                print(f"     - {log}")
        else:
            print("   ✓ 没有控制台日志")
        
        # 检查是否有错误
        errors = [log for log in console_logs if 'error' in log.lower() or 'exception' in log.lower()]
        if errors:
            print(f"   ✗ 发现 {len(errors)} 个错误:")
            for error in errors:
                print(f"     - {error}")
        else:
            print("   ✓ 没有发现错误")
        
        # 测试页面导航
        print("7. 测试页面导航...")
        
        # 尝试点击第一个链接
        if links:
            try:
                first_link = links[0]
                link_text = first_link.text_content() or "[无文本]"
                print(f"   点击第一个链接: {link_text}")
                first_link.click()
                page.wait_for_load_state('networkidle')
                time.sleep(1)
                print("   ✓ 导航成功")
                
                # 捕获导航后的截图
                nav_screenshot_path = '/workspace/frontend-nav-screenshot.png'
                page.screenshot(path=nav_screenshot_path, full_page=True)
                print(f"   捕获导航后截图: {nav_screenshot_path}")
                
                # 返回到首页
                page.go_back()
                page.wait_for_load_state('networkidle')
            except Exception as e:
                print(f"   ✗ 导航测试失败: {e}")
        
        # 测试响应式布局
        print("8. 测试响应式布局...")
        
        # 模拟不同屏幕尺寸
        viewport_sizes = [
            (1920, 1080),  # 桌面
            (1366, 768),   # 笔记本
            (768, 1024),   # 平板
            (375, 667)     # 手机
        ]
        
        for width, height in viewport_sizes:
            try:
                page.set_viewport_size({'width': width, 'height': height})
                time.sleep(1)
                print(f"   ✓ 调整到 {width}x{height}")
                
                # 捕获不同尺寸的截图
                responsive_screenshot_path = f'/workspace/frontend-{width}x{height}.png'
                page.screenshot(path=responsive_screenshot_path, full_page=True)
                print(f"   捕获 {width}x{height} 截图: {responsive_screenshot_path}")
            except Exception as e:
                print(f"   ✗ 响应式测试失败 {width}x{height}: {e}")
        
        # 恢复默认视口
        page.set_viewport_size({'width': 1920, 'height': 1080})
        
        print("=== 测试完成 ===")
        
        # 关闭浏览器
        browser.close()

if __name__ == "__main__":
    test_frontend()
