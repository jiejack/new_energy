<template>
  <div
    class="pull-refresh"
    @touchstart="handleTouchStart"
    @touchmove="handleTouchMove"
    @touchend="handleTouchEnd"
  >
    <div
      class="refresh-indicator"
      :style="{ transform: `translateY(${pullDistance}px)` }"
      v-show="pullDistance > 0 || refreshing"
    >
      <el-icon
        v-if="!refreshing"
        :size="20"
        :class="{ rotate: pullDistance > threshold }"
      >
        <ArrowDown />
      </el-icon>
      <el-icon v-else :size="20" class="loading">
        <Loading />
      </el-icon>
      <span class="refresh-text">{{ refreshText }}</span>
    </div>

    <div
      class="content"
      :style="{ transform: `translateY(${pullDistance}px)` }"
    >
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ArrowDown, Loading } from '@element-plus/icons-vue'

interface Props {
  threshold?: number
}

const props = withDefaults(defineProps<Props>(), {
  threshold: 80
})

const emit = defineEmits<{
  (e: 'refresh'): Promise<void> | void
}>()

const pullDistance = ref(0)
const refreshing = ref(false)
const startY = ref(0)
const isPulling = ref(false)

const refreshText = computed(() => {
  if (refreshing.value) return '刷新中...'
  if (pullDistance.value > props.threshold) return '释放刷新'
  return '下拉刷新'
})

const handleTouchStart = (e: TouchEvent) => {
  if (refreshing.value) return
  startY.value = e.touches[0].clientY
  isPulling.value = true
}

const handleTouchMove = (e: TouchEvent) => {
  if (!isPulling.value || refreshing.value) return

  const currentY = e.touches[0].clientY
  const diff = currentY - startY.value

  if (diff > 0 && window.scrollY === 0) {
    e.preventDefault()
    pullDistance.value = Math.min(diff * 0.5, 100)
  }
}

const handleTouchEnd = async () => {
  isPulling.value = false

  if (pullDistance.value > props.threshold && !refreshing.value) {
    refreshing.value = true
    try {
      await emit('refresh')
    } finally {
      refreshing.value = false
    }
  }

  pullDistance.value = 0
}
</script>

<style scoped lang="scss">
.pull-refresh {
  position: relative;
  overflow: hidden;

  .refresh-indicator {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 50px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    color: #909399;
    font-size: 14px;
    transform: translateY(-50px);
    transition: transform 0.3s;

    .rotate {
      transform: rotate(180deg);
      transition: transform 0.3s;
    }

    .loading {
      animation: rotate 1s linear infinite;
    }

    .refresh-text {
      color: #606266;
    }
  }

  .content {
    transition: transform 0.3s;
  }
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
