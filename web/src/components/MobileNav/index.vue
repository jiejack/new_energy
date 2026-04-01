<template>
  <div class="mobile-nav safe-area-bottom" v-if="isMobile">
    <div
      v-for="item in navItems"
      :key="item.path"
      class="nav-item"
      :class="{ active: isActive(item.path) }"
      @click="handleNav(item.path)"
    >
      <el-badge
        v-if="item.badge && item.badge > 0"
        :value="item.badge"
        :max="99"
        class="nav-badge"
      >
        <el-icon :size="24">
          <component :is="item.icon" />
        </el-icon>
      </el-badge>
      <el-icon v-else :size="24">
        <component :is="item.icon" />
      </el-icon>
      <span class="nav-label">{{ item.label }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDevice } from '@/utils/device'
import { HomeFilled, Bell, Monitor, User } from '@element-plus/icons-vue'

interface NavItem {
  path: string
  label: string
  icon: any
  badge?: number
}

const route = useRoute()
const router = useRouter()
const { isMobile } = useDevice()

const navItems = computed<NavItem[]>(() => [
  { path: '/dashboard', label: '首页', icon: HomeFilled },
  { path: '/alarm/list', label: '告警', icon: Bell, badge: 0 },
  { path: '/device', label: '设备', icon: Monitor },
  { path: '/profile', label: '我的', icon: User }
])

const isActive = (path: string): boolean => {
  return route.path === path || route.path.startsWith(path + '/')
}

const handleNav = (path: string) => {
  if (route.path !== path) {
    router.push(path)
  }
}
</script>

<style scoped lang="scss">
@import '@/styles/responsive.scss';

.mobile-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 60px;
  background: #fff;
  border-top: 1px solid #eee;
  display: flex;
  justify-content: space-around;
  align-items: center;
  z-index: 1000;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.1);

  .nav-item {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    position: relative;
    padding: 8px 0;
    cursor: pointer;
    transition: all 0.3s;
    color: #909399;

    &.active {
      color: #409eff;
    }

    &:active {
      transform: scale(0.95);
    }

    .nav-label {
      font-size: 12px;
      margin-top: 4px;
    }

    .nav-badge {
      :deep(.el-badge__content) {
        top: 2px;
        right: calc(50% - 20px);
      }
    }
  }
}

.safe-area-bottom {
  padding-bottom: env(safe-area-inset-bottom);
}
</style>
