<template>
  <div class="history-data-page">
    <el-card class="query-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">查询条件</span>
        </div>
      </template>

      <el-form :model="queryForm" label-width="100px">
        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="时间范围">
              <TimeRangePicker v-model="queryForm.timeRange" @change="handleTimeChange" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="采集点">
              <PointSelector
                v-model="queryForm.pointIds"
                :show-selected="true"
                @change="handlePointChange"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="数据类型">
              <el-select v-model="queryForm.dataType" placeholder="请选择数据类型">
                <el-option label="原始数据" value="raw" />
                <el-option label="分钟均值" value="minute_avg" />
                <el-option label="小时均值" value="hour_avg" />
                <el-option label="日均值" value="day_avg" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="聚合方式">
              <el-select v-model="queryForm.aggregation" placeholder="请选择聚合方式">
                <el-option label="平均值" value="avg" />
                <el-option label="最大值" value="max" />
                <el-option label="最小值" value="min" />
                <el-option label="求和" value="sum" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="采样间隔">
              <el-select v-model="queryForm.interval" placeholder="请选择采样间隔">
                <el-option label="1分钟" :value="60" />
                <el-option label="5分钟" :value="300" />
                <el-option label="15分钟" :value="900" />
                <el-option label="30分钟" :value="1800" />
                <el-option label="1小时" :value="3600" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row>
          <el-col :span="24">
            <el-form-item>
              <el-button type="primary" :loading="loading" @click="handleQuery">
                <el-icon><Search /></el-icon>
                查询
              </el-button>
              <el-button @click="handleReset">
                <el-icon><Refresh /></el-icon>
                重置
              </el-button>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </el-card>

    <el-card v-if="queryResult.length > 0" class="result-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">查询结果</span>
          <div class="actions">
            <el-radio-group v-model="viewMode" size="small">
              <el-radio-button label="chart">图表</el-radio-button>
              <el-radio-button label="table">表格</el-radio-button>
            </el-radio-group>
            <el-dropdown @command="handleExport">
              <el-button type="primary" size="small">
                <el-icon><Download /></el-icon>
                导出
                <el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="excel">导出Excel</el-dropdown-item>
                  <el-dropdown-item command="csv">导出CSV</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </template>

      <div v-show="viewMode === 'chart'" class="chart-container">
        <DataChart
          :data="queryResult"
          height="400px"
          :show-legend="true"
          :show-data-zoom="true"
          :show-toolbox="true"
        />
      </div>

      <div v-show="viewMode === 'table'" class="table-container">
        <DataTable
          :data="queryResult"
          :show-quality="true"
          :show-pagination="true"
          :page-size="20"
        />
      </div>
    </el-card>

    <el-empty v-else description="请设置查询条件后点击查询按钮" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh, Download, ArrowDown } from '@element-plus/icons-vue'
import TimeRangePicker from '../components/TimeRangePicker.vue'
import PointSelector from '../components/PointSelector.vue'
import DataChart from '../components/DataChart.vue'
import DataTable from '../components/DataTable.vue'
import { queryHistoryData, exportData } from '@/api/data'
import type { PointData } from '@/types'

interface TimeRange {
  startTime: string
  endTime: string
}

const loading = ref(false)
const viewMode = ref<'chart' | 'table'>('chart')
const queryResult = ref<PointData[]>([])

const queryForm = reactive({
  timeRange: {
    startTime: '',
    endTime: '',
  } as TimeRange,
  pointIds: [] as number[],
  dataType: 'raw',
  aggregation: 'avg' as 'avg' | 'max' | 'min' | 'sum',
  interval: 300,
})

// 时间变化
const handleTimeChange = (range: TimeRange) => {
  queryForm.timeRange = range
}

// 采集点变化
const handlePointChange = (ids: number[]) => {
  queryForm.pointIds = ids
}

// 查询数据
const handleQuery = async () => {
  if (!queryForm.timeRange.startTime || !queryForm.timeRange.endTime) {
    ElMessage.warning('请选择时间范围')
    return
  }

  if (queryForm.pointIds.length === 0) {
    ElMessage.warning('请选择采集点')
    return
  }

  loading.value = true
  try {
    const result = await queryHistoryData({
      pointIds: queryForm.pointIds,
      startTime: queryForm.timeRange.startTime,
      endTime: queryForm.timeRange.endTime,
      interval: queryForm.interval,
      aggregation: queryForm.aggregation,
    })
    queryResult.value = result
    ElMessage.success(`查询成功，共 ${result.length} 个采集点数据`)
  } catch (error: any) {
    ElMessage.error(error.message || '查询失败')
  } finally {
    loading.value = false
  }
}

// 重置查询
const handleReset = () => {
  queryForm.timeRange = {
    startTime: '',
    endTime: '',
  }
  queryForm.pointIds = []
  queryForm.dataType = 'raw'
  queryForm.aggregation = 'avg'
  queryForm.interval = 300
  queryResult.value = []
}

// 导出数据
const handleExport = async (format: 'excel' | 'csv') => {
  if (queryResult.value.length === 0) {
    ElMessage.warning('暂无数据可导出')
    return
  }

  try {
    const blob = await exportData({
      pointIds: queryForm.pointIds,
      startTime: queryForm.timeRange.startTime,
      endTime: queryForm.timeRange.endTime,
      interval: queryForm.interval,
      aggregation: queryForm.aggregation,
      format,
    })

    // 创建下载链接
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `历史数据_${new Date().toISOString().slice(0, 10)}.${format === 'excel' ? 'xlsx' : 'csv'}`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    ElMessage.success('导出成功')
  } catch (error: any) {
    ElMessage.error(error.message || '导出失败')
  }
}
</script>

<style scoped lang="scss">
.history-data-page {
  .query-card {
    margin-bottom: 16px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .title {
      font-size: 16px;
      font-weight: 500;
    }

    .actions {
      display: flex;
      gap: 12px;
    }
  }

  .result-card {
    .chart-container,
    .table-container {
      margin-top: 16px;
    }
  }
}
</style>
