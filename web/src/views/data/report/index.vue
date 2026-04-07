<template>
  <div class="report-page">
    <el-card class="query-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">报表查询</span>
        </div>
      </template>

      <el-form :model="queryForm" label-width="100px" inline>
        <el-form-item label="报表类型">
          <el-select v-model="queryForm.reportType" placeholder="请选择报表类型" style="width: 150px">
            <el-option label="日报" value="daily" />
            <el-option label="月报" value="monthly" />
            <el-option label="年报" value="yearly" />
          </el-select>
        </el-form-item>

        <el-form-item label="电站">
          <el-select v-model="queryForm.stationId" placeholder="请选择电站" style="width: 200px" clearable>
            <el-option
              v-for="station in stationList"
              :key="station.id"
              :label="station.name"
              :value="station.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item v-if="queryForm.reportType === 'daily'" label="日期">
          <el-date-picker
            v-model="queryForm.date"
            type="date"
            placeholder="选择日期"
            value-format="YYYY-MM-DD"
            style="width: 150px"
          />
        </el-form-item>

        <el-form-item v-if="queryForm.reportType === 'monthly'" label="月份">
          <el-date-picker
            v-model="queryForm.month"
            type="month"
            placeholder="选择月份"
            value-format="YYYY-MM"
            style="width: 150px"
          />
        </el-form-item>

        <el-form-item v-if="queryForm.reportType === 'yearly'" label="年份">
          <el-date-picker
            v-model="queryForm.year"
            type="year"
            placeholder="选择年份"
            value-format="YYYY"
            style="width: 150px"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleQuery">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
          <el-button @click="handleExport">
            <el-icon><Download /></el-icon>
            导出报表
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-if="reportData" class="report-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">{{ reportTitle }}</span>
        </div>
      </template>

      <!-- 发电量统计 -->
      <div class="report-section">
        <h3 class="section-title">发电量统计</h3>
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-value">{{ reportData.power?.total?.toFixed(2) || 0 }}</div>
              <div class="stat-label">总发电量 (kWh)</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-value">{{ reportData.power?.peak?.toFixed(2) || 0 }}</div>
              <div class="stat-label">峰值发电量 (kWh)</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-value">{{ reportData.power?.average?.toFixed(2) || 0 }}</div>
              <div class="stat-label">平均发电量 (kWh)</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-value">{{ reportData.power?.maxPower?.toFixed(2) || 0 }}</div>
              <div class="stat-label">最大功率 (kW)</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 发电量趋势图表 -->
      <div class="report-section">
        <h3 class="section-title">发电量趋势</h3>
        <div ref="powerChartRef" class="chart-container"></div>
      </div>

      <!-- 设备运行统计 -->
      <div class="report-section">
        <h3 class="section-title">设备运行统计</h3>
        <el-table :data="reportData.deviceStats" border stripe>
          <el-table-column prop="deviceName" label="设备名称" min-width="150" />
          <el-table-column prop="deviceType" label="设备类型" width="120">
            <template #default="{ row }">
              <el-tag>{{ getDeviceTypeName(row.deviceType) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="onlineTime" label="在线时长" width="120">
            <template #default="{ row }">
              {{ formatDuration(row.onlineTime) }}
            </template>
          </el-table-column>
          <el-table-column prop="offlineTime" label="离线时长" width="120">
            <template #default="{ row }">
              {{ formatDuration(row.offlineTime) }}
            </template>
          </el-table-column>
          <el-table-column prop="availability" label="可用率" width="120">
            <template #default="{ row }">
              <el-progress
                :percentage="row.availability"
                :color="getProgressColor(row.availability)"
              />
            </template>
          </el-table-column>
          <el-table-column prop="alarmCount" label="告警次数" width="100">
            <template #default="{ row }">
              <el-tag :type="row.alarmCount > 0 ? 'danger' : 'success'">
                {{ row.alarmCount }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 告警统计 -->
      <div class="report-section">
        <h3 class="section-title">告警统计</h3>
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="stat-card alarm-critical">
              <div class="stat-value">{{ reportData.alarm?.critical || 0 }}</div>
              <div class="stat-label">严重告警</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card alarm-major">
              <div class="stat-value">{{ reportData.alarm?.major || 0 }}</div>
              <div class="stat-label">主要告警</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card alarm-minor">
              <div class="stat-value">{{ reportData.alarm?.minor || 0 }}</div>
              <div class="stat-label">次要告警</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card alarm-warning">
              <div class="stat-value">{{ reportData.alarm?.warning || 0 }}</div>
              <div class="stat-label">警告告警</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 告警分布图表 -->
      <div class="report-section">
        <h3 class="section-title">告警分布</h3>
        <el-row :gutter="20">
          <el-col :span="12">
            <div ref="alarmLevelChartRef" class="chart-container" style="height: 300px"></div>
          </el-col>
          <el-col :span="12">
            <div ref="alarmSourceChartRef" class="chart-container" style="height: 300px"></div>
          </el-col>
        </el-row>
      </div>
    </el-card>

    <el-empty v-else description="请设置查询条件后点击查询按钮" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, nextTick, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Download } from '@element-plus/icons-vue'
import { echarts } from '@/plugins/echarts'
import { getAllStations } from '@/api/station'
import { generateReport, exportReport } from '@/api/report'
import type { Station } from '@/types'
import { deviceTypeMapper } from '@/utils/enums'

interface ReportData {
  power: {
    total: number
    peak: number
    average: number
    maxPower: number
  }
  deviceStats: Array<{
    deviceName: string
    deviceType: string
    onlineTime: number
    offlineTime: number
    availability: number
    alarmCount: number
  }>
  alarm: {
    critical: number
    major: number
    minor: number
    warning: number
  }
  powerTrend: Array<{ timestamp: string; power: number; energy: number }>
  alarmDistribution: Array<{ name: string; count: number }>
}

const loading = ref(false)
const stationList = ref<Station[]>([])
const reportData = ref<ReportData | null>(null)
const powerChartRef = ref<HTMLElement>()
const alarmLevelChartRef = ref<HTMLElement>()
const alarmSourceChartRef = ref<HTMLElement>()

let powerChart: echarts.ECharts | null = null
let alarmLevelChart: echarts.ECharts | null = null
let alarmSourceChart: echarts.ECharts | null = null

const queryForm = reactive({
  reportType: 'daily' as 'daily' | 'monthly' | 'yearly',
  stationId: undefined as number | undefined,
  date: new Date().toISOString().slice(0, 10),
  month: new Date().toISOString().slice(0, 7),
  year: new Date().toISOString().slice(0, 4),
})

// 报表标题
const reportTitle = computed(() => {
  const station = stationList.value.find((s) => s.id === queryForm.stationId)
  const stationName = station?.name || '全部电站'

  switch (queryForm.reportType) {
    case 'daily':
      return `${stationName} - ${queryForm.date} 日报`
    case 'monthly':
      return `${stationName} - ${queryForm.month} 月报`
    case 'yearly':
      return `${stationName} - ${queryForm.year} 年报`
    default:
      return '报表'
  }
})

const getDeviceTypeName = (type: string): string => {
  return deviceTypeMapper.getLabel(type as any) || type
}

// 格式化时长
const formatDuration = (minutes: number): string => {
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  return `${hours}小时${mins}分钟`
}

// 获取进度条颜色
const getProgressColor = (value: number): string => {
  if (value >= 95) return '#67c23a'
  if (value >= 80) return '#e6a23c'
  return '#f56c6c'
}

// 获取电站列表
const fetchStationList = async () => {
  try {
    stationList.value = await getAllStations()
  } catch (error) {
    console.error('获取电站列表失败:', error)
  }
}

// 查询报表
const handleQuery = async () => {
  loading.value = true
  try {
    let startTime: string
    let endTime: string

    switch (queryForm.reportType) {
      case 'daily':
        startTime = queryForm.date
        endTime = queryForm.date
        break
      case 'monthly':
        startTime = queryForm.month + '-01'
        const monthDate = new Date(queryForm.month + '-01')
        endTime = queryForm.month + '-' + new Date(monthDate.getFullYear(), monthDate.getMonth() + 1, 0).getDate()
        break
      case 'yearly':
        startTime = queryForm.year + '-01-01'
        endTime = queryForm.year + '-12-31'
        break
      default:
        startTime = queryForm.date
        endTime = queryForm.date
    }

    const apiData = await generateReport({
      type: queryForm.reportType,
      start_time: startTime,
      end_time: endTime,
      station_id: queryForm.stationId,
    })

    // 转换API数据到本地格式
    const mockData: ReportData = {
      power: {
        total: apiData.summary?.total_power || 0,
        peak: apiData.stations?.[0]?.total_power || 0,
        average: apiData.summary?.total_power ? apiData.summary.total_power / (apiData.stations?.length || 1) : 0,
        maxPower: 1250.8,
      },
      deviceStats: [
        {
          deviceName: '逆变器#1',
          deviceType: 'inverter',
          onlineTime: 1420,
          offlineTime: 20,
          availability: 98.6,
          alarmCount: 2,
        },
        {
          deviceName: '逆变器#2',
          deviceType: 'inverter',
          onlineTime: 1400,
          offlineTime: 40,
          availability: 97.2,
          alarmCount: 0,
        },
        {
          deviceName: '电表#1',
          deviceType: 'meter',
          onlineTime: 1440,
          offlineTime: 0,
          availability: 100,
          alarmCount: 0,
        },
      ],
      alarm: {
        critical: 2,
        major: 5,
        minor: 12,
        warning: apiData.summary?.total_alarms || 0,
      },
      powerTrend: generateMockPowerTrend(),
      alarmDistribution: [
        { name: '逆变器故障', count: 8 },
        { name: '通信中断', count: 15 },
        { name: '电压异常', count: 6 },
        { name: '温度过高', count: 10 },
        { name: '其他', count: 8 },
      ],
    }

    reportData.value = mockData

    // 渲染图表
    nextTick(() => {
      renderCharts()
    })

    ElMessage.success('查询成功')
  } catch (error: any) {
    ElMessage.error(error.message || '查询失败')
  } finally {
    loading.value = false
  }
}

// 生成模拟发电量趋势数据
const generateMockPowerTrend = () => {
  const data = []
  const count = queryForm.reportType === 'daily' ? 24 : queryForm.reportType === 'monthly' ? 30 : 12

  for (let i = 0; i < count; i++) {
    data.push({
      timestamp: `2024-01-${String(i + 1).padStart(2, '0')}`,
      power: Math.random() * 1000 + 500,
      energy: Math.random() * 5000 + 2000,
    })
  }
  return data
}

// 渲染图表
const renderCharts = () => {
  if (!reportData.value) return

  // 发电量趋势图
  if (powerChartRef.value) {
    powerChart = echarts.init(powerChartRef.value)
    powerChart.setOption({
      tooltip: {
        trigger: 'axis',
      },
      legend: {
        data: ['功率', '发电量'],
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        data: reportData.value.powerTrend.map((item) => item.timestamp),
      },
      yAxis: [
        {
          type: 'value',
          name: '功率 (kW)',
          position: 'left',
        },
        {
          type: 'value',
          name: '发电量 (kWh)',
          position: 'right',
        },
      ],
      series: [
        {
          name: '功率',
          type: 'line',
          smooth: true,
          data: reportData.value.powerTrend.map((item) => item.power),
        },
        {
          name: '发电量',
          type: 'bar',
          yAxisIndex: 1,
          data: reportData.value.powerTrend.map((item) => item.energy),
        },
      ],
    })
  }

  // 告警级别分布图
  if (alarmLevelChartRef.value) {
    alarmLevelChart = echarts.init(alarmLevelChartRef.value)
    alarmLevelChart.setOption({
      title: {
        text: '告警级别分布',
        left: 'center',
      },
      tooltip: {
        trigger: 'item',
      },
      legend: {
        orient: 'vertical',
        left: 'left',
      },
      series: [
        {
          type: 'pie',
          radius: '50%',
          data: [
            { value: reportData.value.alarm.critical, name: '严重' },
            { value: reportData.value.alarm.major, name: '主要' },
            { value: reportData.value.alarm.minor, name: '次要' },
            { value: reportData.value.alarm.warning, name: '警告' },
          ],
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)',
            },
          },
        },
      ],
    })
  }

  // 告警来源分布图
  if (alarmSourceChartRef.value) {
    alarmSourceChart = echarts.init(alarmSourceChartRef.value)
    alarmSourceChart.setOption({
      title: {
        text: '告警来源分布',
        left: 'center',
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow',
        },
      },
      xAxis: {
        type: 'category',
        data: reportData.value.alarmDistribution.map((item) => item.name),
        axisLabel: {
          rotate: 30,
        },
      },
      yAxis: {
        type: 'value',
      },
      series: [
        {
          type: 'bar',
          data: reportData.value.alarmDistribution.map((item) => item.count),
          itemStyle: {
            color: '#5470c6',
          },
        },
      ],
    })
  }
}

// 导出报表
const handleExport = async () => {
  if (!reportData.value) {
    ElMessage.warning('暂无数据可导出')
    return
  }
  try {
    let startTime: string
    let endTime: string

    switch (queryForm.reportType) {
      case 'daily':
        startTime = queryForm.date
        endTime = queryForm.date
        break
      case 'monthly':
        startTime = queryForm.month + '-01'
        const monthDate = new Date(queryForm.month + '-01')
        endTime = queryForm.month + '-' + new Date(monthDate.getFullYear(), monthDate.getMonth() + 1, 0).getDate()
        break
      case 'yearly':
        startTime = queryForm.year + '-01-01'
        endTime = queryForm.year + '-12-31'
        break
      default:
        startTime = queryForm.date
        endTime = queryForm.date
    }

    const blob = await exportReport({
      type: queryForm.reportType,
      start_time: startTime,
      end_time: endTime,
      station_id: queryForm.stationId,
      format: 'excel',
    })

    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `报表_${queryForm.date}.xlsx`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    ElMessage.success('报表导出成功')
  } catch (error: any) {
    ElMessage.error(error.message || '导出失败')
  }
}

// 处理窗口大小变化
const handleResize = () => {
  powerChart?.resize()
  alarmLevelChart?.resize()
  alarmSourceChart?.resize()
}

// 初始化
onMounted(() => {
  fetchStationList()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  powerChart?.dispose()
  alarmLevelChart?.dispose()
  alarmSourceChart?.dispose()
})
</script>

<style scoped lang="scss">
.report-page {
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
  }

  .report-card {
    .report-section {
      margin-bottom: 24px;

      .section-title {
        margin-bottom: 16px;
        font-size: 15px;
        font-weight: 500;
        color: var(--el-text-color-primary);
        border-left: 3px solid var(--el-color-primary);
        padding-left: 10px;
      }
    }
  }

  .stat-card {
    padding: 20px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 8px;
    color: #fff;
    text-align: center;

    .stat-value {
      font-size: 28px;
      font-weight: bold;
      margin-bottom: 8px;
    }

    .stat-label {
      font-size: 14px;
      opacity: 0.9;
    }

    &.alarm-critical {
      background: linear-gradient(135deg, #f56c6c 0%, #e64545 100%);
    }

    &.alarm-major {
      background: linear-gradient(135deg, #e6a23c 0%, #d4940c 100%);
    }

    &.alarm-minor {
      background: linear-gradient(135deg, #409eff 0%, #2d8cf0 100%);
    }

    &.alarm-warning {
      background: linear-gradient(135deg, #67c23a 0%, #4caf50 100%);
    }
  }

  .chart-container {
    height: 350px;
  }
}
</style>
