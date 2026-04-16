<template>
  <div class="global-notification">
    <el-button
      class="notification-trigger"
      @click="showNotification = true"
    >
      <el-icon><Bell /></el-icon>
      <span v-if="unreadCount > 0" class="unread-badge">{{ unreadCount }}</span>
    </el-button>

    <el-drawer
      v-model="showNotification"
      title="通知中心"
      size="400px"
      direction="rtl"
    >
      <div class="notification-tabs">
        <el-tabs v-model="activeTab" @tab-change="handleTabChange">
          <el-tab-pane label="全部" name="all" />
          <el-tab-pane label="告警" name="alarm">
            <template #label>
              <span>告警 <el-tag type="danger" size="small">{{ alarmCount }}</el-tag></span>
            </template>
          </el-tab-pane>
          <el-tab-pane label="消息" name="message" />
          <el-tab-pane label="系统" name="system" />
        </el-tabs>
      </div>

      <div class="notification-actions">
        <el-button text size="small" @click="markAllAsRead">
          <el-icon><Check /></el-icon>
          全部已读
        </el-button>
        <el-button text size="small" type="danger" @click="clearAll">
          <el-icon><Delete /></el-icon>
          清空
        </el-button>
      </div>

      <div class="notification-list">
        <div v-if="filteredNotifications.length === 0" class="empty-state">
          <el-empty description="暂无通知" />
        </div>

        <div
          v-for="notification in filteredNotifications"
          :key="notification.id"
          class="notification-item"
          :class="{ unread: !notification.read }"
          @click="handleNotificationClick(notification)"
        >
          <div class="notification-icon">
            <el-icon :size="24" :color="getIconColor(notification.type)">
              <component :is="getIconComponent(notification.type)" />
            </el-icon>
          </div>

          <div class="notification-content">
            <div class="notification-header">
              <span class="notification-title">{{ notification.title }}</span>
              <span class="notification-time">{{ formatTime(notification.timestamp) }}</span>
            </div>
            <div class="notification-message">{{ notification.message }}</div>
            <div v-if="notification.data" class="notification-data">
              <el-tag v-for="(value, key) in notification.data" :key="key" size="small">
                {{ key }}: {{ value }}
              </el-tag>
            </div>
          </div>

          <div class="notification-status">
            <el-icon v-if="!notification.read" class="unread-dot" color="#ff7675" />
          </div>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Bell, Check, Delete, Warning, Info, Success, CircleClose } from '@element-plus/icons-vue'
import { wsManager } from '@/utils/websocket'
import dayjs from 'dayjs'

interface Notification {
  id: string
  type: 'alarm' | 'message' | 'system' | 'success' | 'warning' | 'error'
  title: string
  message: string
  read: boolean
  timestamp: Date
  data?: Record<string, any>
}

const showNotification = ref(false)
const activeTab = ref('all')
const notifications = ref<Notification[]>([])
let wsListener: (() => void) | null = null

const unreadCount = computed(() => notifications.value.filter(n => !n.read).length)
const alarmCount = computed(() => notifications.value.filter(n => n.type === 'alarm').length)

const filteredNotifications = computed(() => {
  if (activeTab.value === 'all') {
    return notifications.value
  }
  return notifications.value.filter(n => n.type === activeTab.value)
})

function getIconColor(type: string): string {
  const colorMap: Record<string, string> = {
    alarm: '#ff7675',
    error: '#ff7675',
    warning: '#fdcb6e',
    message: '#74b9ff',
    system: '#a3a7b2',
    success: '#00d4aa',
  }
  return colorMap[type] || '#a3a7b2'
}

function getIconComponent(type: string): any {
  const iconMap: Record<string, any> = {
    alarm: Warning,
    error: CircleClose,
    warning: Warning,
    message: Info,
    system: Info,
    success: Success,
  }
  return iconMap[type] || Info
}

function formatTime(timestamp: Date): string {
  const now = dayjs()
  const time = dayjs(timestamp)
  
  if (now.diff(time, 'minute') < 1) {
    return '刚刚'
  }
  if (now.diff(time, 'hour') < 1) {
    return `${now.diff(time, 'minute')}分钟前`
  }
  if (now.diff(time, 'day') < 1) {
    return `${now.diff(time, 'hour')}小时前`
  }
  return time.format('MM-DD HH:mm')
}

function handleNotificationClick(notification: Notification) {
  notification.read = true
}

function markAllAsRead() {
  notifications.value.forEach(n => n.read = true)
  ElMessage.success('已全部标记为已读')
}

async function clearAll() {
  try {
    await ElMessageBox.confirm('确定要清空所有通知吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    notifications.value = []
    ElMessage.success('通知已清空')
  } catch {
    // 用户取消
  }
}

function handleTabChange() {
  // 切换标签页时的逻辑
}

function addNotification(notification: Omit<Notification, 'id' | 'read' | 'timestamp'>) {
  const newNotification: Notification = {
    ...notification,
    id: `notif-${Date.now()}-${Math.random()}`,
    read: false,
    timestamp: new Date(),
  }
  
  notifications.value.unshift(newNotification)
  
  // 保留最近100条通知
  if (notifications.value.length > 100) {
    notifications.value = notifications.value.slice(0, 100)
  }
  
  // 如果是告警类型，显示弹窗提示
  if (notification.type === 'alarm') {
    ElMessage.warning({
      message: notification.message,
      duration: 5000,
      showClose: true,
    })
  }
}

function initWebSocket() {
  wsManager.on('alarm', (msg: any) => {
    addNotification({
      type: 'alarm',
      title: '新告警',
      message: msg.payload?.message || msg.message || '新告警通知',
      data: msg.payload,
    })
  })
  
  wsManager.on('notification', (msg: any) => {
    addNotification({
      type: msg.payload?.type || 'message',
      title: msg.payload?.title || '新消息',
      message: msg.payload?.message || '您有一条新消息',
      data: msg.payload?.data,
    })
  })
}

onMounted(() => {
  initWebSocket()
})

onUnmounted(() => {
  if (wsListener) {
    wsListener()
  }
})

// 暴露给外部使用的方法
defineExpose({
  addNotification,
})
</script>

<style scoped lang="scss">
.global-notification {
  position: relative;
  
  .notification-trigger {
    position: relative;
    
    .unread-badge {
      position: absolute;
      top: -8px;
      right: -8px;
      min-width: 18px;
      height: 18px;
      padding: 0 4px;
      font-size: 10px;
      line-height: 18px;
      text-align: center;
      background: #ff7675;
      border-radius: 9px;
      color: white;
    }
  }
}

.notification-tabs {
  margin-bottom: 16px;
}

.notification-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(0, 212, 170, 0.2);
}

.notification-list {
  max-height: calc(100vh - 200px);
  overflow-y: auto;
  
  .empty-state {
    padding: 40px 0;
  }
}

.notification-item {
  display: flex;
  gap: 12px;
  padding: 12px;
  margin-bottom: 8px;
  background: rgba(26, 31, 46, 0.6);
  border: 1px solid rgba(0, 212, 170, 0.1);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  
  &:hover {
    background: rgba(26, 31, 46, 0.9);
    border-color: rgba(0, 212, 170, 0.3);
  }
  
  &.unread {
    background: rgba(0, 212, 170, 0.05);
    border-left: 3px solid #00d4aa;
  }
}

.notification-icon {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 212, 170, 0.1);
  border-radius: 8px;
}

.notification-content {
  flex: 1;
  min-width: 0;
  
  .notification-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
    
    .notification-title {
      font-weight: 600;
      color: #e5eaf3;
    }
    
    .notification-time {
      font-size: 12px;
      color: #a3a7b2;
      flex-shrink: 0;
      margin-left: 8px;
    }
  }
  
  .notification-message {
    font-size: 14px;
    color: #cfd3dc;
    margin-bottom: 8px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  
  .notification-data {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }
}

.notification-status {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  
  .unread-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }
}
</style>
