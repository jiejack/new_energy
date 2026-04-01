<template>
  <div class="alarm-list">
    <!-- 告警列表 -->
    <el-scrollbar ref="scrollbarRef" class="list-scrollbar">
      <div v-if="loading" class="loading-container">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>加载中...</span>
      </div>

      <div v-else-if="alarms.length === 0" class="empty-container">
        <el-icon><CircleCheck /></el-icon>
        <span>暂无告警</span>
      </div>

      <div v-else class="alarm-items" ref="alarmItemsRef">
        <div
          v-for="alarm in alarms"
          :key="alarm.id"
          class="alarm-item"
          :class="`level-${alarm.level}`"
        >
          <div class="alarm-level">
            <el-icon :size="16">
              <component :is="getLevelIcon(alarm.level)" />
            </el-icon>
          </div>
          <div class="alarm-content">
            <div class="alarm-title">{{ alarm.title }}</div>
            <div class="alarm-desc">{{ alarm.content }}</div>
            <div class="alarm-meta">
              <span class="alarm-source">{{ alarm.sourceName }}</span>
              <span class="alarm-time">{{ formatTime(alarm.occurredAt) }}</span>
            </div>
          </div>
          <div class="alarm-actions">
            <el-button
              type="primary"
              size="small"
              text
              @click="handleAcknowledge(alarm)"
            >
              确认
            </el-button>
          </div>
        </div>
      </div>
    </el-scrollbar>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { Loading, CircleCheck, WarningFilled, CircleCloseFilled, InfoFilled, BellFilled } from '@element-plus/icons-vue'
import type { Alarm, AlarmLevel } from '@/types'

interface Props {
  alarms: Alarm[]
  loading?: boolean
  autoScroll?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  autoScroll: true
})

const emit = defineEmits<{
  acknowledge: [alarm: Alarm]
}>()

const scrollbarRef = ref()
const alarmItemsRef = ref()
let scrollTimer: number | null = null
let scrollDirection = 1
let isHovering = false

/**
 * 获取告警级别图标
 */
function getLevelIcon(level: AlarmLevel) {
  const iconMap: Record<AlarmLevel, any> = {
    critical: CircleCloseFilled,
    major: WarningFilled,
    minor: BellFilled,
    warning: InfoFilled
  }
  return iconMap[level] || InfoFilled
}

/**
 * 格式化时间
 */
function formatTime(time: string) {
  const date = new Date(time)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  // 小于1分钟
  if (diff < 60000) {
    return '刚刚'
  }

  // 小于1小时
  if (diff < 3600000) {
    return `${Math.floor(diff / 60000)}分钟前`
  }

  // 小于24小时
  if (diff < 86400000) {
    return `${Math.floor(diff / 3600000)}小时前`
  }

  // 显示日期时间
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${date.getMonth() + 1}/${date.getDate()} ${hours}:${minutes}`
}

/**
 * 处理确认
 */
function handleAcknowledge(alarm: Alarm) {
  emit('acknowledge', alarm)
}

/**
 * 自动滚动
 */
function startAutoScroll() {
  if (!props.autoScroll || props.alarms.length <= 3) return

  scrollTimer = window.setInterval(() => {
    if (isHovering || !scrollbarRef.value) return

    const scrollbar = scrollbarRef.value
    const scrollElement = scrollbar.wrapRef
    if (!scrollElement) return

    const { scrollTop, scrollHeight, clientHeight } = scrollElement
    const maxScroll = scrollHeight - clientHeight

    // 到达底部，改变方向
    if (scrollTop >= maxScroll - 5) {
      scrollDirection = -1
    }
    // 到达顶部，改变方向
    else if (scrollTop <= 5) {
      scrollDirection = 1
    }

    scrollElement.scrollTop += scrollDirection * 0.5
  }, 50)
}

/**
 * 停止自动滚动
 */
function stopAutoScroll() {
  if (scrollTimer) {
    clearInterval(scrollTimer)
    scrollTimer = null
  }
}

// 监听告警列表变化，自动滚动到底部
watch(() => props.alarms, (newAlarms, oldAlarms) => {
  if (newAlarms.length > (oldAlarms?.length || 0)) {
    nextTick(() => {
      if (scrollbarRef.value) {
        const scrollElement = scrollbarRef.value.wrapRef
        if (scrollElement) {
          scrollElement.scrollTop = 0
        }
      }
    })
  }
}, { deep: true })

// 鼠标悬停事件
onMounted(() => {
  startAutoScroll()

  if (alarmItemsRef.value) {
    alarmItemsRef.value.addEventListener('mouseenter', () => {
      isHovering = true
    })
    alarmItemsRef.value.addEventListener('mouseleave', () => {
      isHovering = false
    })
  }
})

onUnmounted(() => {
  stopAutoScroll()
})
</script>

<style scoped lang="scss">
.alarm-list {
  height: 100%;
  display: flex;
  flex-direction: column;

  .list-scrollbar {
    flex: 1;

    :deep(.el-scrollbar__bar) {
      &.is-vertical {
        width: 6px;
        right: 2px;
      }

      .el-scrollbar__thumb {
        background-color: rgba(64, 158, 255, 0.3);

        &:hover {
          background-color: rgba(64, 158, 255, 0.5);
        }
      }
    }
  }

  .loading-container,
  .empty-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 200px;
    color: #909399;
    gap: 10px;

    .el-icon {
      font-size: 32px;
    }
  }

  .empty-container {
    .el-icon {
      color: #67c23a;
    }
  }

  .alarm-items {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .alarm-item {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 10px;
    background: rgba(32, 45, 65, 0.4);
    border: 1px solid rgba(64, 158, 255, 0.2);
    border-radius: 6px;
    transition: all 0.3s;

    &:hover {
      background: rgba(64, 158, 255, 0.1);
      border-color: rgba(64, 158, 255, 0.4);
    }

    // 不同级别的样式
    &.level-critical {
      border-left: 3px solid #f56c6c;
      background: rgba(245, 108, 108, 0.1);

      .alarm-level {
        color: #f56c6c;
      }
    }

    &.level-major {
      border-left: 3px solid #e6a23c;
      background: rgba(230, 162, 60, 0.1);

      .alarm-level {
        color: #e6a23c;
      }
    }

    &.level-minor {
      border-left: 3px solid #409eff;
      background: rgba(64, 158, 255, 0.1);

      .alarm-level {
        color: #409eff;
      }
    }

    &.level-warning {
      border-left: 3px solid #909399;
      background: rgba(144, 147, 153, 0.1);

      .alarm-level {
        color: #909399;
      }
    }

    .alarm-level {
      flex-shrink: 0;
      width: 24px;
      height: 24px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: rgba(0, 0, 0, 0.2);
      border-radius: 50%;
    }

    .alarm-content {
      flex: 1;
      min-width: 0;

      .alarm-title {
        font-size: 13px;
        font-weight: 500;
        color: #e5eaf3;
        margin-bottom: 4px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .alarm-desc {
        font-size: 12px;
        color: #909399;
        margin-bottom: 6px;
        overflow: hidden;
        text-overflow: ellipsis;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
      }

      .alarm-meta {
        display: flex;
        align-items: center;
        gap: 10px;
        font-size: 11px;
        color: #606266;

        .alarm-source {
          color: #409eff;
        }

        .alarm-time {
          color: #909399;
        }
      }
    }

    .alarm-actions {
      flex-shrink: 0;

      .el-button {
        padding: 4px 8px;
        font-size: 12px;
      }
    }
  }
}
</style>
