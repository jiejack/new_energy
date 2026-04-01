<template>
  <div class="point-selector">
    <el-select
      v-model="selectedPoints"
      multiple
      filterable
      remote
      reserve-keyword
      placeholder="请选择采集点"
      :remote-method="handleSearch"
      :loading="loading"
      :collapse-tags="collapseTags"
      :collapse-tags-tooltip="true"
      :max-collapse-tags="3"
      style="width: 100%"
      @change="handleChange"
    >
      <el-option
        v-for="point in pointList"
        :key="point.id"
        :label="`${point.name} (${point.code})`"
        :value="point.id"
      >
        <div class="point-option">
          <span class="point-name">{{ point.name }}</span>
          <span class="point-info">
            <el-tag size="small" type="info">{{ point.deviceName }}</el-tag>
            <el-tag size="small">{{ point.unit || '无单位' }}</el-tag>
          </span>
        </div>
      </el-option>
    </el-select>

    <div v-if="showSelected && selectedPoints.length > 0" class="selected-points">
      <div class="selected-header">
        <span>已选择 {{ selectedPoints.length }} 个采集点</span>
        <el-button type="primary" link size="small" @click="clearSelection">清空</el-button>
      </div>
      <div class="selected-list">
        <el-tag
          v-for="id in selectedPoints"
          :key="id"
          closable
          size="small"
          class="selected-tag"
          @close="removePoint(id)"
        >
          {{ getPointName(id) }}
        </el-tag>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { getPointList } from '@/api/point'
import type { Point } from '@/types'

const props = defineProps<{
  modelValue?: number[]
  deviceId?: number
  showSelected?: boolean
  collapseTags?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: number[]): void
  (e: 'change', value: number[], points: Point[]): void
}>()

const loading = ref(false)
const pointList = ref<Point[]>([])
const selectedPoints = ref<number[]>([])
const allPoints = ref<Point[]>([])

// 获取采集点名称
const getPointName = (id: number): string => {
  const point = allPoints.value.find((p) => p.id === id)
  return point ? point.name : `ID: ${id}`
}

// 搜索采集点
const handleSearch = async (keyword: string) => {
  loading.value = true
  try {
    const params: any = {
      page: 1,
      pageSize: 50,
    }
    if (keyword) {
      params.keyword = keyword
    }
    if (props.deviceId) {
      params.deviceId = props.deviceId
    }
    const result = await getPointList(params)
    pointList.value = result.list

    // 更新全部采集点缓存
    result.list.forEach((point) => {
      if (!allPoints.value.find((p) => p.id === point.id)) {
        allPoints.value.push(point)
      }
    })
  } catch (error) {
    console.error('获取采集点列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 选择变化
const handleChange = (value: number[]) => {
  emit('update:modelValue', value)
  const points = allPoints.value.filter((p) => value.includes(p.id))
  emit('change', value, points)
}

// 移除采集点
const removePoint = (id: number) => {
  selectedPoints.value = selectedPoints.value.filter((p) => p !== id)
  handleChange(selectedPoints.value)
}

// 清空选择
const clearSelection = () => {
  selectedPoints.value = []
  handleChange(selectedPoints.value)
}

// 监听外部值变化
watch(
  () => props.modelValue,
  (val) => {
    if (val) {
      selectedPoints.value = [...val]
    }
  },
  { immediate: true }
)

// 监听设备ID变化
watch(
  () => props.deviceId,
  () => {
    handleSearch('')
  }
)

// 初始化
onMounted(() => {
  handleSearch('')
})
</script>

<style scoped lang="scss">
.point-selector {
  .point-option {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;

    .point-name {
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .point-info {
      display: flex;
      gap: 4px;
      margin-left: 8px;
    }
  }

  .selected-points {
    margin-top: 12px;
    padding: 12px;
    background-color: var(--el-fill-color-light);
    border-radius: 4px;

    .selected-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 8px;
      font-size: 13px;
      color: var(--el-text-color-secondary);
    }

    .selected-list {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;

      .selected-tag {
        max-width: 200px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }
  }
}
</style>
