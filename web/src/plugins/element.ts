import type { App } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

/**
 * 注册Element Plus图标
 */
function registerIcons(app: App) {
  for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
  }
}

/**
 * 配置Element Plus
 */
export function setupElementPlus(app: App) {
  // 注册图标
  registerIcons(app)

  // 使用Element Plus
  app.use(ElementPlus, {
    locale: zhCn,
    size: 'default',
    zIndex: 3000,
  })
}

export default setupElementPlus
