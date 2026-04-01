<template>
  <div class="data-table">
    <el-table
      ref="tableRef"
      :data="tableData"
      :height="height"
      :max-height="maxHeight"
      :stripe="stripe"
      :border="border"
      :row-key="rowKey"
      :default-expand-all="defaultExpandAll"
      :highlight-current-row="highlightCurrentRow"
      @selection-change="handleSelectionChange"
      @sort-change="handleSortChange"
    >
      <el-table-column v-if="showSelection" type="selection" width="55" />

      <el-table-column v-if="showIndex" type="index" label="序号" width="60" />

      <el-table-column prop="timestamp" label="时间" width="180" sortable>
        <template #default="{ row }">
          {{ formatTime(row.timestamp) }}
        </template>
      </el-table-column>

      <el-table-column
        v-for="point in pointColumns"
        :key="point.pointId"
        :prop="`point_${point.pointId}`"
        :label="point.pointName"
        min-width="150"
        sortable
      >
        <template #default="{ row }">
          <span :class="getValueClass(row[`point_${point.pointId}`]?.quality)">
            {{ formatValue(row[`point_${point.pointId}`]?.value) }}
            <span v-if="point.unit" class="unit">{{ point.unit }}</span>
          </span>
        </template>
      </el-table-column>

      <el-table-column v-if="showQuality" label="数据质量" width="100">
        <template #default="{ row }">
          <el-tag :type="getQualityTagType(row.quality)" size="small">
            {{ getQualityText(row.quality) }}
          </el-tag>
        </template>
      </el-table-column>

      <slot name="operation"></slot>
    </el-table>

    <div v-if="showPagination" class="pagination-container">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100, 200]"
        :total="total"
        :background="true"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import dayjs from 'dayjs'
import type { PointData, DataPoint } from '@/types'

interface TableRow {
  timestamp: string
  quality: number
  [key: string]: any
}

const props = withDefaults(
  defineProps<{
    data: PointData[]
    height?: string | number
    maxHeight?: string | number
    stripe?: boolean
    border?: boolean
    rowKey?: string
    defaultExpandAll?: boolean
    highlightCurrentRow?: boolean
    showSelection?: boolean
    showIndex?: boolean
    showQuality?: boolean
    showPagination?: boolean
    pageSize?: number
  }>(),
  {
    stripe: true,
    border: false,
    rowKey: 'timestamp',
    defaultExpandAll: false,
    highlightCurrentRow: false,
    showSelection: false,
    showIndex: true,
    showQuality: false,
    showPagination: true,
    pageSize: 20,
  }
)

const emit = defineEmits<{
  (e: 'selection-change', selection: TableRow[]): void
  (e: 'sort-change', sort: { prop: string; order: string }): void
  (e: 'page-change', page: number, pageSize: number): void
}>()

const tableRef = ref()
const currentPage = ref(1)
const pageSize = ref(props.pageSize)

// 采集点列配置
const pointColumns = computed(() => {
  return props.data.map((point) => ({
    pointId: point.pointId,
    pointName: point.pointName,
    unit: point.unit,
  }))
})

// 转换数据为表格格式
const tableData = computed(() => {
  // 获取所有时间点
  const timeMap = new Map<string, TableRow>()

  props.data.forEach((pointData) => {
    pointData.data.forEach((dataPoint: DataPoint) => {
      const timestamp = dataPoint.timestamp
      if (!timeMap.has(timestamp)) {
        timeMap.set(timestamp, {
          timestamp,
          quality: dataPoint.quality,
        })
      }
      const row = timeMap.get(timestamp)!
      row[`point_${pointData.pointId}`] = {
        value: dataPoint.value,
        quality: dataPoint.quality,
      }
    })
  })

  // 转换为数组并排序
  const rows = Array.from(timeMap.values())
  rows.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())

  // 分页
  if (props.showPagination) {
    const start = (currentPage.value - 1) * pageSize.value
    const end = start + pageSize.value
    return rows.slice(start, end)
  }

  return rows
})

// 总数
const total = computed(() => {
  const timeSet = new Set<string>()
  props.data.forEach((pointData) => {
    pointData.data.forEach((dataPoint: DataPoint) => {
      timeSet.add(dataPoint.timestamp)
    })
  })
  return timeSet.size
})

// 格式化时间
const formatTime = (timestamp: string): string => {
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

// 格式化值
const formatValue = (value: number | undefined): string => {
  if (value === undefined || value === null) return '-'
  return value.toFixed(2)
}

// 获取值样式类
const getValueClass = (quality: number | undefined): string => {
  if (quality === undefined || quality >= 200) return 'value-bad'
  if (quality >= 100) return 'value-good'
  return 'value-normal'
}

// 获取数据质量标签类型
const getQualityTagType = (quality: number): 'success' | 'warning' | 'danger' | 'info' => {
  if (quality >= 200) return 'danger'
  if (quality >= 100) return 'success'
  return 'warning'
}

// 获取数据质量文本
const getQualityText = (quality: number): string => {
  if (quality >= 200) return '异常'
  if (quality >= 100) return '良好'
  return '一般'
}

// 选择变化
const handleSelectionChange = (selection: TableRow[]) => {
  emit('selection-change', selection)
}

// 排序变化
const handleSortChange = ({ prop, order }: { prop: string; order: string | null }) => {
  emit('sort-change', { prop, order: order || '' })
}

// 每页数量变化
const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  emit('page-change', currentPage.value, pageSize.value)
}

// 当前页变化
const handleCurrentChange = (page: number) => {
  currentPage.value = page
  emit('page-change', currentPage.value, pageSize.value)
}

// 监听数据变化重置分页
watch(
  () => props.data,
  () => {
    currentPage.value = 1
  }
)

// 暴露方法
defineExpose({
  clearSelection: () => tableRef.value?.clearSelection(),
  toggleRowSelection: (row: TableRow, selected: boolean) =>
    tableRef.value?.toggleRowSelection(row, selected),
})
</script>

<style scoped lang="scss">
.data-table {
  .value-good {
    color: var(--el-color-success);
  }

  .value-normal {
    color: var(--el-color-warning);
  }

  .value-bad {
    color: var(--el-color-danger);
  }

  .unit {
    margin-left: 4px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
}
</style>
