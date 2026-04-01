<template>
  <div class="time-range-picker">
    <el-radio-group v-model="quickRange" class="quick-options" @change="handleQuickChange">
      <el-radio-button label="today">今天</el-radio-button>
      <el-radio-button label="yesterday">昨天</el-radio-button>
      <el-radio-button label="week">近7天</el-radio-button>
      <el-radio-button label="month">近30天</el-radio-button>
      <el-radio-button label="custom">自定义</el-radio-button>
    </el-radio-group>

    <el-date-picker
      v-if="quickRange === 'custom'"
      v-model="customRange"
      type="datetimerange"
      range-separator="至"
      start-placeholder="开始时间"
      end-placeholder="结束时间"
      :shortcuts="shortcuts"
      :disabled-date="disabledDate"
      value-format="YYYY-MM-DD HH:mm:ss"
      @change="handleCustomChange"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import dayjs from 'dayjs'

interface TimeRange {
  startTime: string
  endTime: string
}

const props = defineProps<{
  modelValue?: TimeRange
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: TimeRange): void
  (e: 'change', value: TimeRange): void
}>()

const quickRange = ref<string>('today')
const customRange = ref<[string, string] | null>(null)

// 快捷选项
const shortcuts = [
  {
    text: '最近1小时',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 3600 * 1000)
      return [start, end]
    },
  },
  {
    text: '最近6小时',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 3600 * 1000 * 6)
      return [start, end]
    },
  },
  {
    text: '最近12小时',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 3600 * 1000 * 12)
      return [start, end]
    },
  },
  {
    text: '最近24小时',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 3600 * 1000 * 24)
      return [start, end]
    },
  },
  {
    text: '最近一周',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 7)
      return [start, end]
    },
  },
  {
    text: '最近一个月',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 30)
      return [start, end]
    },
  },
]

// 禁用未来日期
const disabledDate = (time: Date) => {
  return time.getTime() > Date.now()
}

// 获取时间范围
const getTimeRange = (range: string): TimeRange => {
  const now = dayjs()
  let startTime = ''
  let endTime = now.format('YYYY-MM-DD HH:mm:ss')

  switch (range) {
    case 'today':
      startTime = now.startOf('day').format('YYYY-MM-DD HH:mm:ss')
      break
    case 'yesterday':
      startTime = now.subtract(1, 'day').startOf('day').format('YYYY-MM-DD HH:mm:ss')
      endTime = now.subtract(1, 'day').endOf('day').format('YYYY-MM-DD HH:mm:ss')
      break
    case 'week':
      startTime = now.subtract(6, 'day').startOf('day').format('YYYY-MM-DD HH:mm:ss')
      break
    case 'month':
      startTime = now.subtract(29, 'day').startOf('day').format('YYYY-MM-DD HH:mm:ss')
      break
    default:
      break
  }

  return { startTime, endTime }
}

// 快捷选择变化
const handleQuickChange = (value: string) => {
  if (value !== 'custom') {
    const range = getTimeRange(value)
    emit('update:modelValue', range)
    emit('change', range)
  }
}

// 自定义时间变化
const handleCustomChange = (value: [string, string] | null) => {
  if (value && value.length === 2) {
    const range: TimeRange = {
      startTime: value[0],
      endTime: value[1],
    }
    emit('update:modelValue', range)
    emit('change', range)
  }
}

// 监听外部值变化
watch(
  () => props.modelValue,
  (val) => {
    if (val) {
      // 检查是否匹配快捷选项
      const todayStart = dayjs().startOf('day').format('YYYY-MM-DD HH:mm:ss')
      const yesterdayStart = dayjs().subtract(1, 'day').startOf('day').format('YYYY-MM-DD HH:mm:ss')
      const yesterdayEnd = dayjs().subtract(1, 'day').endOf('day').format('YYYY-MM-DD HH:mm:ss')
      const weekStart = dayjs().subtract(6, 'day').startOf('day').format('YYYY-MM-DD HH:mm:ss')
      const monthStart = dayjs().subtract(29, 'day').startOf('day').format('YYYY-MM-DD HH:mm:ss')

      if (val.startTime === todayStart) {
        quickRange.value = 'today'
      } else if (val.startTime === yesterdayStart && val.endTime === yesterdayEnd) {
        quickRange.value = 'yesterday'
      } else if (val.startTime === weekStart) {
        quickRange.value = 'week'
      } else if (val.startTime === monthStart) {
        quickRange.value = 'month'
      } else {
        quickRange.value = 'custom'
        customRange.value = [val.startTime, val.endTime]
      }
    }
  },
  { immediate: true }
)

// 初始化
if (!props.modelValue) {
  const range = getTimeRange('today')
  emit('update:modelValue', range)
}
</script>

<style scoped lang="scss">
.time-range-picker {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;

  .quick-options {
    flex-wrap: wrap;
  }
}
</style>
