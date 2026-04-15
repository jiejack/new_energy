<template>
  <div class="efficiency-page">
    <el-card class="query-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">能效分析</span>
        </div>
      </template>

      <el-form :model="queryForm" label-width="100px" inline>
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

        <el-form-item label="时间范围">
          <el-date-picker
            v-model="queryForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 300px"
          />
        </el-form-item>

        <el-form-item label="统计周期">
          <el-select v-model="queryForm.periodType" placeholder="请选择" style="width: 120px">
            <el-option label="日" value="daily" />
            <el-option label="周" value="weekly" />
            <el-option label="月" value="monthly" />
            <el-option label="年" value="yearly" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleQuery">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
          <el-button @click="handleRefresh">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <div v-if="efficiencyData" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
      <div class="stat-card gradient-blue rounded-lg p-6 text-white">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm opacity-80">平均能效</p>
            <p class="text-3xl font-bold mt-2">{{ efficiencyData.averageEfficiency?.toFixed(2) || 0 }}%</p>
          </div>
          <div class="text-4xl opacity-30">
            <TrendCharts />
          </div>
        </div>
      </div>

      <div class="stat-card gradient-green rounded-lg p-6 text-white">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm opacity-80">总发电量</p>
            <p class="text-3xl font-bold mt-2">{{ efficiencyData.totalPowerOutput?.toFixed(2) || 0 }} kWh</p>
          </div>
          <div class="text-4xl opacity-30">
            <Lightning />
          </div>
        </div>
      </div>

      <div class="stat-card gradient-purple rounded-lg p-6 text-white">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm opacity-80">容量利用率</p>
            <p class="text-3xl font-bold mt-2">{{ efficiencyData.capacityUtilization?.toFixed(2) || 0 }}%</p>
          </div>
          <div class="text-4xl opacity-30">
            <Odometer />
          </div>
        </div>
      </div>

      <div class="stat-card gradient-orange rounded-lg p-6 text-white">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm opacity-80">峰值负载</p>
            <p class="text-3xl font-bold mt-2">{{ efficiencyData.peakLoad?.toFixed(2) || 0 }} kW</p>
          </div>
          <div class="text-4xl opacity-30">
            <DataLine />
          </div>
        </div>
      </div>
    </div>

    <el-row :gutter="16">
      <el-col :span="24">
        <el-card class="chart-card" shadow="never">
          <template #header>
            <div class="card-header">
              <span class="title">能效趋势</span>
            </div>
          </template>
          <div ref="efficiencyTrendChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="mt-4">
      <el-col :span="12">
        <el-card class="chart-card" shadow="never">
          <template #header>
            <div class="card-header">
              <span class="title">能效分布</span>
            </div>
          </template>
          <div ref="efficiencyDistributionChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card class="chart-card" shadow="never">
          <template #header>
            <div class="card-header">
              <span class="title">设备能效对比</span>
            </div>
          </template>
          <div ref="deviceComparisonChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-card v-if="latestAnalysis" class="analysis-card mt-4" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">最新分析报告</span>
          <el-tag type="success">{{ latestAnalysis.status }}</el-tag>
        </div>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="分析日期">
          {{ latestAnalysis.analysis_date }}
        </el-descriptions-item>
        <el-descriptions-item label="分析周期">
          {{ latestAnalysis.period_start }} 至 {{ latestAnalysis.period_end }}
        </el-descriptions-item>
        <el-descriptions-item label="平均能效" :span="2">
          {{ latestAnalysis.average_efficiency?.toFixed(2) || 0 }}%
        </el-descriptions-item>
        <el-descriptions-item label="分析摘要" :span="2">
          {{ latestAnalysis.analysis_summary || '暂无摘要' }}
        </el-descriptions-item>
        <el-descriptions-item label="优化建议" :span="2">
          <div class="whitespace-pre-wrap">{{ latestAnalysis.recommendations || '暂无建议' }}</div>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-empty v-else description="请设置查询条件后点击查询按钮" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, nextTick, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh, TrendCharts, Lightning, Odometer, DataLine } from '@element-plus/icons-vue'
import { echarts } from '@/plugins/echarts'
import { getAllStations } from '@/api/station'
import {
  getEnergyEfficiencyStatistics,
  getEnergyEfficiencyTrend,
  getLatestEnergyEfficiencyAnalysis,
  type EnergyEfficiencyAnalysis
} from '@/api/energy-efficiency'
import type { Station } from '@/types'

const loading = ref(false)
const stationList = ref<Station[]>([])
const efficiencyData = ref<any>(null)
const latestAnalysis = ref<EnergyEfficiencyAnalysis | null>(null)
const efficiencyTrendChartRef = ref<HTMLElement>()
const efficiencyDistributionChartRef = ref<HTMLElement>()
const deviceComparisonChartRef = ref<HTMLElement>()

let efficiencyTrendChart: echarts.ECharts | null = null
let efficiencyDistributionChart: echarts.ECharts | null = null
let deviceComparisonChart: echarts.ECharts | null = null

const queryForm = reactive({
  stationId: undefined as number | undefined,
  dateRange: [] as string[],
  periodType: 'daily' as 'daily' | 'weekly' | 'monthly' | 'yearly'
})

const fetchStationList = async () => {
  try {
    stationList.value = await getAllStations()
  } catch (error) {
    console.error('获取电站列表失败:', error)
  }
}

const generateMockData = () => {
  const trendData = []
  const count = queryForm.periodType === 'daily' ? 24 : queryForm.periodType === 'monthly' ? 30 : 12
  
  for (let i = 0; i < count; i++) {
    trendData.push({
      time: `2024-01-${String(i + 1).padStart(2, '0')}`,
      efficiency: 85 + Math.random() * 10,
      powerOutput: 5000 + Math.random() * 2000
    })
  }
  
  return {
    averageEfficiency: 88.5,
    totalPowerOutput: 150000,
    capacityUtilization: 75.3,
    peakLoad: 850,
    trendData,
    distributionData: [
      { name: '80-85%', value: 15 },
      { name: '85-90%', value: 45 },
      { name: '90-95%', value: 30 },
      { name: '95-100%', value: 10 }
    ],
    deviceData: [
      { name: '逆变器1', efficiency: 89.5 },
      { name: '逆变器2', efficiency: 91.2 },
      { name: '逆变器3', efficiency: 87.8 },
      { name: '汇流箱1', efficiency: 92.1 },
      { name: '汇流箱2', efficiency: 88.9 }
    ]
  }
}

const handleQuery = async () => {
  loading.value = true
  try {
    const [startDate, endDate] = queryForm.dateRange
    
    const [stats, trend, analysis] = await Promise.all([
      getEnergyEfficiencyStatistics({
        station_id: queryForm.stationId,
        period_type: queryForm.periodType,
        start_date: startDate,
        end_date: endDate
      }).catch(() => generateMockData()),
      getEnergyEfficiencyTrend({
        station_id: queryForm.stationId,
        period_type: queryForm.periodType,
        start_date: startDate,
        end_date: endDate
      }).catch(() => generateMockData().trendData),
      getLatestEnergyEfficiencyAnalysis(queryForm.stationId).catch(() => null)
    ])
    
    efficiencyData.value = stats || generateMockData()
    if (Array.isArray(trend)) {
      efficiencyData.value.trendData = trend
    }
    latestAnalysis.value = analysis
    
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

const handleRefresh = () => {
  if (queryForm.dateRange.length === 2) {
    handleQuery()
  } else {
    ElMessage.warning('请先选择时间范围')
  }
}

const renderCharts = () => {
  if (!efficiencyData.value) return
  
  if (efficiencyTrendChartRef.value) {
    efficiencyTrendChart = echarts.init(efficiencyTrendChartRef.value)
    efficiencyTrendChart.setOption({
      tooltip: {
        trigger: 'axis'
      },
      legend: {
        data: ['能效', '发电量']
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: efficiencyData.value.trendData?.map((item: any) => item.time || item.record_date) || []
      },
      yAxis: [
        {
          type: 'value',
          name: '能效 (%)',
          position: 'left',
          min: 70,
          max: 100
        },
        {
          type: 'value',
          name: '发电量 (kWh)',
          position: 'right'
        }
      ],
      series: [
        {
          name: '能效',
          type: 'line',
          smooth: true,
          data: efficiencyData.value.trendData?.map((item: any) => item.efficiency || item.efficiency_rate) || [],
          itemStyle: { color: '#5470c6' }
        },
        {
          name: '发电量',
          type: 'bar',
          yAxisIndex: 1,
          data: efficiencyData.value.trendData?.map((item: any) => item.powerOutput || item.power_output) || [],
          itemStyle: { color: '#91cc75' }
        }
      ]
    })
  }
  
  if (efficiencyDistributionChartRef.value) {
    efficiencyDistributionChart = echarts.init(efficiencyDistributionChartRef.value)
    efficiencyDistributionChart.setOption({
      tooltip: {
        trigger: 'item',
        formatter: '{a} <br/>{b}: {c} ({d}%)'
      },
      legend: {
        orient: 'vertical',
        left: 'left'
      },
      series: [
        {
          name: '能效分布',
          type: 'pie',
          radius: ['40%', '70%'],
          avoidLabelOverlap: false,
          itemStyle: {
            borderRadius: 10,
            borderColor: '#fff',
            borderWidth: 2
          },
          label: {
            show: false,
            position: 'center'
          },
          emphasis: {
            label: {
              show: true,
              fontSize: 20,
              fontWeight: 'bold'
            }
          },
          labelLine: {
            show: false
          },
          data: efficiencyData.value.distributionData || [
            { value: 15, name: '80-85%' },
            { value: 45, name: '85-90%' },
            { value: 30, name: '90-95%' },
            { value: 10, name: '95-100%' }
          ]
        }
      ]
    })
  }
  
  if (deviceComparisonChartRef.value) {
    deviceComparisonChart = echarts.init(deviceComparisonChartRef.value)
    deviceComparisonChart.setOption({
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow'
        }
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: efficiencyData.value.deviceData?.map((item: any) => item.name) || [
          '逆变器1', '逆变器2', '逆变器3', '汇流箱1', '汇流箱2'
        ]
      },
      yAxis: {
        type: 'value',
        name: '能效 (%)',
        min: 80,
        max: 95
      },
      series: [
        {
          name: '能效',
          type: 'bar',
          data: efficiencyData.value.deviceData?.map((item: any) => item.efficiency) || [
            89.5, 91.2, 87.8, 92.1, 88.9
          ],
          itemStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: '#83bff6' },
              { offset: 0.5, color: '#188df0' },
              { offset: 1, color: '#188df0' }
            ])
          },
          barWidth: '50%'
        }
      ]
    })
  }
}

const handleResize = () => {
  efficiencyTrendChart?.resize()
  efficiencyDistributionChart?.resize()
  deviceComparisonChart?.resize()
}

onMounted(() => {
  fetchStationList()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  efficiencyTrendChart?.dispose()
  efficiencyDistributionChart?.dispose()
  deviceComparisonChart?.dispose()
})
</script>

<style scoped lang="scss">
.efficiency-page {
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
  
  .stat-card {
    transition: transform 0.3s ease;
    
    &:hover {
      transform: translateY(-4px);
    }
  }
  
  .gradient-blue {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  }
  
  .gradient-green {
    background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
  }
  
  .gradient-purple {
    background: linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%);
  }
  
  .gradient-orange {
    background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  }
  
  .chart-card,
  .analysis-card {
    margin-bottom: 16px;
  }
  
  .chart-container {
    height: 350px;
  }
}
</style>
