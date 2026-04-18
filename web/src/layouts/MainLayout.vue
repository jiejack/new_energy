<template>
  <div class="main-layout">
    <!-- 侧边栏 -->
    <div class="sidebar" :class="{ collapsed: isCollapsed }">
      <div class="logo">
        <img src="@/assets/vue.svg" alt="Logo" class="logo-img" />
        <span v-show="!isCollapsed" class="logo-text">新能源监控系统</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapsed"
        :unique-opened="true"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
        router
      >
        <template v-for="route in menuRoutes" :key="route.path">
          <!-- 没有子菜单 -->
          <el-menu-item
            v-if="!route.children || route.children.length === 1"
            :index="route.children ? route.redirect : `/${route.path}`"
            @click="handleMenuClick(route.children ? route.redirect : `/${route.path}`)"
          >
            <el-icon>
              <component :is="route.meta?.icon || 'Document'" />
            </el-icon>
            <template #title>{{ route.meta?.title || route.children?.[0]?.meta?.title }}</template>
          </el-menu-item>

          <!-- 有子菜单 -->
          <el-sub-menu v-else :index="`/${route.path}`">
            <template #title>
              <el-icon>
                <component :is="route.meta?.icon || 'Document'" />
              </el-icon>
              <span>{{ route.meta?.title }}</span>
            </template>
            <el-menu-item
              v-for="child in route.children"
              :key="child.path"
              :index="`/${route.path}/${child.path}`"
              @click="handleMenuClick(`/${route.path}/${child.path}`)"
            >
              <el-icon>
                <component :is="child.meta?.icon || 'Document'" />
              </el-icon>
              <template #title>{{ child.meta?.title }}</template>
            </el-menu-item>
          </el-sub-menu>
        </template>
      </el-menu>
    </div>

    <!-- 主内容区 -->
    <div class="main-container">
      <!-- 顶部导航栏 -->
      <div class="navbar">
        <div class="left-menu">
          <el-icon class="collapse-btn" @click="toggleSidebar">
            <Fold v-if="!isCollapsed" />
            <Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.path">
              {{ item.meta?.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="right-menu">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="32" :src="userAvatar">
                <el-icon><User /></el-icon>
              </el-avatar>
              <span class="username">{{ username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item command="settings">系统设置</el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!-- 标签页 -->
      <div class="tags-view">
        <el-tag
          v-for="tag in visitedTags"
          :key="tag.path"
          :closable="tag.path !== '/dashboard'"
          :effect="activeMenu === tag.path ? 'dark' : 'plain'"
          @click="$router.push(tag.path)"
          @close="closeTag(tag)"
        >
          {{ tag.meta?.title }}
        </el-tag>
      </div>

      <!-- 内容区 -->
      <div class="app-main">
        <router-view v-slot="{ Component }">
          <transition name="fade-transform" mode="out-in">
            <keep-alive>
              <component :is="Component" />
            </keep-alive>
          </transition>
        </router-view>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import { asyncRoutes } from '@/router'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()

// 侧边栏折叠状态
const isCollapsed = computed(() => !appStore.sidebarOpened)

// 当前激活菜单
const activeMenu = computed(() => route.path)

// 用户信息
const username = computed(() => userStore.nickname || userStore.username)
const userAvatar = computed(() => userStore.avatar)

// 菜单路由
const menuRoutes = computed(() => {
  const routes = asyncRoutes.find((r) => r.path === '/')?.children || []
  console.log('menuRoutes:', routes.map(r => ({ path: r.path, children: r.children?.length })))
  return routes
})

// 面包屑
const breadcrumbs = computed(() => {
  const matched = route.matched.filter((item) => item.meta?.title)
  return matched
})

// 访问过的标签
const visitedTags = ref<Array<any>>([
  {
    path: '/dashboard',
    meta: { title: '仪表盘' },
  },
])

// 监听路由变化，添加标签
watch(
  () => route.path,
  (path) => {
    if (route.meta?.title) {
      const exists = visitedTags.value.some((tag) => tag.path === path)
      if (!exists) {
        visitedTags.value.push({
          path,
          meta: route.meta,
        })
      }
    }
  },
  { immediate: true }
)

// 切换侧边栏
function toggleSidebar() {
  appStore.toggleSidebar()
}

// 关闭标签
function closeTag(tag: any) {
  const index = visitedTags.value.findIndex((t) => t.path === tag.path)
  visitedTags.value.splice(index, 1)

  // 如果关闭的是当前标签，跳转到上一个标签
  if (route.path === tag.path) {
    const lastTag = visitedTags.value[visitedTags.value.length - 1]
    router.push(lastTag.path)
  }
}

// 处理下拉菜单命令
async function handleCommand(command: string) {
  switch (command) {
    case 'profile':
      router.push('/profile')
      break
    case 'settings':
      router.push('/settings')
      break
    case 'logout':
      await userStore.logoutAction()
      router.push('/login')
      break
  }
}

// 处理菜单点击
function handleMenuClick(path: string) {
  console.log('Menu clicked:', path)
  router.push(path)
}
</script>

<style scoped lang="scss">
.main-layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.sidebar {
  width: $sidebar-width;
  background-color: #304156;
  transition: width $transition-duration $transition-timing-function;
  overflow: hidden;

  &.collapsed {
    width: $sidebar-collapsed-width;
  }
}

.logo {
  display: flex;
  align-items: center;
  height: $navbar-height;
  padding: 0 15px;
  background-color: #2b3649;
  overflow: hidden;

  .logo-img {
    width: 32px;
    height: 32px;
  }

  .logo-text {
    margin-left: 10px;
    font-size: 16px;
    font-weight: bold;
    color: #fff;
    white-space: nowrap;
  }
}

.el-menu {
  border: none;
  height: calc(100vh - #{$navbar-height});
  overflow-y: auto;
}

.main-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: $navbar-height;
  padding: 0 20px;
  background-color: $bg-white;
  box-shadow: $shadow-base;
  z-index: $z-index-navbar;
}

.left-menu {
  display: flex;
  align-items: center;

  .collapse-btn {
    font-size: 20px;
    cursor: pointer;
    margin-right: 15px;
  }
}

.right-menu {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;

  .username {
    margin: 0 8px;
    color: $text-primary;
  }
}

.tags-view {
  display: flex;
  align-items: center;
  height: $tagsview-height;
  padding: 0 10px;
  background-color: $bg-white;
  border-bottom: 1px solid $border-lighter;
  overflow-x: auto;

  .el-tag {
    margin-right: 5px;
    cursor: pointer;
  }
}

.app-main {
  flex: 1;
  overflow-y: auto;
  background-color: $bg-page;
}

// 过渡动画
.fade-transform-enter-active,
.fade-transform-leave-active {
  transition: all 0.3s;
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(-30px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(30px);
}
</style>
