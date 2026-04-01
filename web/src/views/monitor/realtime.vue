<template>
  <div class="realtime-monitor-page">
    <el-card class="page-header">
      <template #header>
        <div class="card-header">
          <span>实时数据监控</span>
          <div class="header-actions">
            <el-button type="primary" @click="refreshData">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
            <el-button @click="toggleAutoRefresh">
              <el-icon><Timer /></el-icon>
              {{ autoRefresh ? '停止自动刷新' : '自动刷新' }}
            </el-button>
          </div>
        </div>
      </template>

      <el-row :gutter="20">
        <el-col :span="6">
          <div class="stat-card">
            <div class="stat-value">{{ stats.totalStations }}</div>
            <div class="stat-label">电站总数</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card">
            <div class="stat-value">{{ stats.onlineStations }}</div>
            <div class="stat-label">在线电站</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card">
            <div class="stat-value">{{ stats.totalPower.toFixed(2) }}</div>
            <div class="stat-label">总功率 (kW)</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card">
            <div class="stat-value">{{ stats.todayEnergy.toFixed(2) }}</div>
            <div class="stat-label">今日发电量 (kWh)</div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <el-row :gutter="20" class="data-section">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>功率曲线</span>
          </template>
          <div ref="powerChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>发电量对比</span>
          </template>
          <div ref="energyChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="data-table-section">
      <template #header>
        <span>实时数据列表</span>
      </template>
      <el-table :data="realtimeData" v-loading="loading" stripe>
        <el-table-column prop="stationName" label="电站名称" />
        <el-table-column prop="deviceName" label="设备名称" />
        <el-table-column prop="pointName" label="采集点" />
        <el-table-column prop="value" label="当前值">
          <template #default="{ row }">
            <span :class="getValueClass(row)">{{ row.value }} {{ row.unit }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="updateTime" label="更新时间" />
        <el-table-column prop="status" label="状态">
          <template #default="{ row }">
            <el-tag :type="row.status === 'normal' ? 'success' : 'danger'">
              {{ row.status === 'normal' ? '正常' : '异常' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next"
        @change="handlePageChange"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import { Refresh, Timer } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { getRealtimeData } from '@/api/data'

// 统计数据
const stats = reactive({
  totalStations: 12,
  onlineStations: 10,
  totalPower: 1256.8,
  todayEnergy: 15680.5
})

// 实时数据
const realtimeData = ref<any[]>([])
const loading = ref(false)
const autoRefresh = ref(false)
let refreshTimer: number | null = null

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 图表引用
const powerChartRef = ref<HTMLElement>()
const energyChartRef = ref<HTMLElement>()
let powerChart: echarts.ECharts | null = null
let energyChart: echarts.ECharts | null = null

// 获取实时数据
async function loadRealtimeData() {
  loading.value = true
  try {
    const result = await getRealtimeData({
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    realtimeData.value = result.list
    pagination.total = result.total
  } catch (error) {
    console.error('加载实时数据失败:', error)
    ElMessage.error('加载实时数据失败')
  } finally {
    loading.value = false
  }
}

// 刷新数据
function refreshData() {
  loadRealtimeData()
  updateCharts()
  ElMessage.success('数据已刷新')
}

// 切换自动刷新
function toggleAutoRefresh() {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) {
    refreshTimer = window.setInterval(() => {
      loadRealtimeData()
      updateCharts()
    }, 5000)
    ElMessage.success('已开启自动刷新（5秒）')
  } else {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
    ElMessage.info('已停止自动刷新')
  }
}

// 分页变化
function handlePageChange() {
  loadRealtimeData()
}

// 获取值样式
function getValueClass(row: any) {
  if (row.status === 'abnormal') return 'value-warning'
  if (row.alarmLevel === 'high') return 'value-danger'
  return 'value-normal'
}

// 初始化图表
function initCharts() {
  if (powerChartRef.value) {
    powerChart = echarts.init(powerChartRef.value)
    powerChart.setOption({
      tooltip: { trigger: 'axis' },
      xAxis: {
        type: 'category',
        data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00', '24:00']
      },
      yAxis: { type: 'value', name: '功率(kW)' },
      series: [{
        name: '功率',
        type: 'line',
        smooth: true,
        data: [120, 132, 101, 134, 90, 230, 210],
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
            { offset: 1, color: 'rgba(64, 158, 255, 0.05)' }
          ])
        }
      }]
    })
  }

  if (energyChartRef.value) {
    energyChart = echarts.init(energyChartRef.value)
    energyChart.setOption({
      tooltip: { trigger: 'axis' },
      xAxis: {
        type: 'category',
        data: ['电站A', '电站B', '电站C', '电站D', '电站E']
      },
      yAxis: { type: 'value', name: '发电量(kWh)' },
      series: [{
        name: '今日发电',
        type: 'bar',
        data: [320, 302, 301, 334, 390],
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#67c23a' },
            { offset: 1, color: '#95d475' }
          ])
        }
      }]
    })
  }
}

// 更新图表
function updateCharts() {
  // 模拟数据更新
  if (powerChart) {
    const newData = Array.from({ length: 7 }, () => Math.floor(Math.random() * 200) + 50)
    powerChart.setOption({
      series: [{ data: newData }]
    })
  }
}

// 生命周期
onMounted(() => {
  loadRealtimeData()
  nextTick(() => {
    initCharts()
  })
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  window.removeEventListener('resize', handleResize)
  powerChart?.dispose()
  energyChart?.dispose()
})

function handleResize() {
  powerChart?.resize()
  energyChart?.resize()
}
</script>

<style scoped lang="scss">
.realtime-monitor-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.stat-card {
  text-align: center;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  color: white;

  .stat-value {
    font-size: 32px;
    font-weight: bold;
    margin-bottom: 8px;
  }

  .stat-label {
    font-size: 14px;
    opacity: 0.9;
  }
}

.data-section {
  margin-bottom: 20px;
}

.chart-container {
  height: 300px;
}

.data-table-section {
  .el-pagination {
    margin-top: 20px;
    justify-content: flex-end;
  }
}

.value-normal {
  color: #67c23a;
}

.value-warning {
  color: #e6a23c;
}

.value-danger {
  color: #f56c6c;
  font-weight: bold;
}
</style>
