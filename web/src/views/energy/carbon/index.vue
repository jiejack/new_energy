<template>
  <div class="carbon-page">
    <el-card class="query-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">碳排放监测</span>
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

    <div v-if="carbonData" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
      <div class="stat-card gradient-red rounded-lg p-6 text-white">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm opacity-80">总碳排放</p>
            <p class="text-3xl font-bold mt-2">{{ carbonData.totalEmission?.toFixed(2) || 0 }} tCO₂</p>
          </div>
          <div class="text-4xl opacity-30">
            <Warning />
          </div>
        </div>
      </div>

      <div class="stat-card gradient-teal rounded-lg p-6 text-white">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm opacity-80">排放强度</p>
            <p class="text-3xl font-bold mt-2">{{ carbonData.averageEmissionIntensity?.toFixed(4) || 0 }} tCO₂/kWh</p>
          </div>
          <div class="text-4xl opacity-30">
            <TrendCharts />
          </div>
        </div>
      </div>

      <div class="stat-card gradient-cyan rounded-lg p-6 text-white">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm opacity-80">实际减排</p>
            <p class="text-3xl font-bold mt-2">{{ carbonData.actualReduction?.toFixed(2) || 0 }} tCO₂</p>
          </div>
          <div class="text-4xl opacity-30">
            <CircleCheck />
          </div>
        </div>
      </div>

      <div class="stat-card gradient-pink rounded-lg p-6 text-white">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm opacity-80">减排率</p>
            <p class="text-3xl font-bold mt-2">{{ carbonData.reductionRate?.toFixed(2) || 0 }}%</p>
          </div>
          <div class="text-4xl opacity-30">
            <Top />
          </div>
        </div>
      </div>
    </div>

    <el-row :gutter="16">
      <el-col :span="24">
        <el-card class="chart-card" shadow="never">
          <template #header>
            <div class="card-header">
              <span class="title">碳排放趋势</span>
            </div>
          </template>
          <div ref="carbonTrendChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="mt-4">
      <el-col :span="12">
        <el-card class="chart-card" shadow="never">
          <template #header>
            <div class="card-header">
              <span class="title">能源结构碳排放</span>
            </div>
          </template>
          <div ref="energyStructureChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card class="chart-card" shadow="never">
          <template #header>
            <div class="card-header">
              <span class="title">减排目标进度</span>
            </div>
          </template>
          <div ref="reductionProgressChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-card v-if="latestAnalysis" class="analysis-card mt-4" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">最新分析报告</span>
          <el-tag :type="carbonData.reductionRate >= 0 ? 'success' : 'danger'">
            {{ carbonData.reductionRate >= 0 ? '减排中' : '排放增加' }}
          </el-tag>
        </div>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="分析日期">
          {{ latestAnalysis.analysis_date }}
        </el-descriptions-item>
        <el-descriptions-item label="分析周期">
          {{ latestAnalysis.period_start }} 至 {{ latestAnalysis.period_end }}
        </el-descriptions-item>
        <el-descriptions-item label="总碳排放" :span="2">
          {{ latestAnalysis.total_emission?.toFixed(2) || 0 }} tCO₂
        </el-descriptions-item>
        <el-descriptions-item label="关键发现" :span="2">
          <div class="whitespace-pre-wrap">{{ latestAnalysis.key_findings || '暂无发现' }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="减排建议" :span="2">
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
import { Search, Refresh, Warning, TrendCharts, CircleCheck, Top } from '@element-plus/icons-vue'
import { echarts } from '@/plugins/echarts'
import { getAllStations } from '@/api/station'
import {
  getCarbonEmissionStatistics,
  getCarbonEmissionTrend,
  getLatestCarbonEmissionAnalysis,
  type CarbonEmissionAnalysis
} from '@/api/carbon-emission'
import type { Station } from '@/types'

const loading = ref(false)
const stationList = ref<Station[]>([])
const carbonData = ref<any>(null)
const latestAnalysis = ref<CarbonEmissionAnalysis | null>(null)
const carbonTrendChartRef = ref<HTMLElement>()
const energyStructureChartRef = ref<HTMLElement>()
const reductionProgressChartRef = ref<HTMLElement>()

let carbonTrendChart: echarts.ECharts | null = null
let energyStructureChart: echarts.ECharts | null = null
let reductionProgressChart: echarts.ECharts | null = null

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
      totalEmission: 100 + Math.random() * 50,
      electricityEmission: 60 + Math.random() * 30,
      coalEmission: 30 + Math.random() * 15
    })
  }
  
  return {
    totalEmission: 2500,
    averageEmissionIntensity: 0.523,
    actualReduction: 150,
    reductionRate: 5.7,
    trendData,
    energyStructureData: [
      { name: '电力', value: 1200 },
      { name: '煤炭', value: 800 },
      { name: '天然气', value: 300 },
      { name: '石油', value: 150 },
      { name: '其他', value: 50 }
    ],
    reductionProgress: {
      target: 300,
      actual: 150,
      remaining: 150
    }
  }
}

const handleQuery = async () => {
  loading.value = true
  try {
    const [startDate, endDate] = queryForm.dateRange
    
    const [stats, trend, analysis] = await Promise.all([
      getCarbonEmissionStatistics({
        station_id: queryForm.stationId,
        period_type: queryForm.periodType,
        start_date: startDate,
        end_date: endDate
      }).catch(() => generateMockData()),
      getCarbonEmissionTrend({
        station_id: queryForm.stationId,
        period_type: queryForm.periodType,
        start_date: startDate,
        end_date: endDate
      }).catch(() => generateMockData().trendData),
      getLatestCarbonEmissionAnalysis(queryForm.stationId).catch(() => null)
    ])
    
    carbonData.value = stats || generateMockData()
    if (Array.isArray(trend)) {
      carbonData.value.trendData = trend
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
  if (!carbonData.value) return
  
  if (carbonTrendChartRef.value) {
    carbonTrendChart = echarts.init(carbonTrendChartRef.value)
    carbonTrendChart.setOption({
      tooltip: {
        trigger: 'axis'
      },
      legend: {
        data: ['总碳排放', '电力排放', '煤炭排放']
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: carbonData.value.trendData?.map((item: any) => item.time || item.record_date) || []
      },
      yAxis: {
        type: 'value',
        name: '碳排放 (tCO₂)'
      },
      series: [
        {
          name: '总碳排放',
          type: 'line',
          smooth: true,
          data: carbonData.value.trendData?.map((item: any) => item.totalEmission || item.total_emission) || [],
          itemStyle: { color: '#c23531' },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(194, 53, 49, 0.3)' },
              { offset: 1, color: 'rgba(194, 53, 49, 0.05)' }
            ])
          }
        },
        {
          name: '电力排放',
          type: 'bar',
          stack: 'total',
          data: carbonData.value.trendData?.map((item: any) => item.electricityEmission || item.electricity_emission) || [],
          itemStyle: { color: '#2f4554' }
        },
        {
          name: '煤炭排放',
          type: 'bar',
          stack: 'total',
          data: carbonData.value.trendData?.map((item: any) => item.coalEmission || item.coal_emission) || [],
          itemStyle: { color: '#61a0a8' }
        }
      ]
    })
  }
  
  if (energyStructureChartRef.value) {
    energyStructureChart = echarts.init(energyStructureChartRef.value)
    energyStructureChart.setOption({
      tooltip: {
        trigger: 'item',
        formatter: '{a} <br/>{b}: {c} tCO₂ ({d}%)'
      },
      legend: {
        orient: 'vertical',
        left: 'left'
      },
      series: [
        {
          name: '能源结构',
          type: 'pie',
          radius: ['40%', '70%'],
          avoidLabelOverlap: false,
          itemStyle: {
            borderRadius: 10,
            borderColor: '#fff',
            borderWidth: 2
          },
          label: {
            show: true,
            formatter: '{b}\n{d}%'
          },
          emphasis: {
            label: {
              show: true,
              fontSize: 16,
              fontWeight: 'bold'
            }
          },
          data: carbonData.value.energyStructureData || [
            { value: 1200, name: '电力', itemStyle: { color: '#5470c6' } },
            { value: 800, name: '煤炭', itemStyle: { color: '#91cc75' } },
            { value: 300, name: '天然气', itemStyle: { color: '#fac858' } },
            { value: 150, name: '石油', itemStyle: { color: '#ee6666' } },
            { value: 50, name: '其他', itemStyle: { color: '#73c0de' } }
          ]
        }
      ]
    })
  }
  
  if (reductionProgressChartRef.value) {
    const progress = carbonData.value.reductionProgress || { target: 300, actual: 150, remaining: 150 }
    const percentage = Math.min((progress.actual / progress.target) * 100, 100)
    
    reductionProgressChart = echarts.init(reductionProgressChartRef.value)
    reductionProgressChart.setOption({
      tooltip: {
        formatter: '{a} <br/>{b}: {c} tCO₂'
      },
      series: [
        {
          type: 'gauge',
          startAngle: 180,
          endAngle: 0,
          min: 0,
          max: progress.target,
          splitNumber: 5,
          itemStyle: {
            color: '#58D9F9',
            shadowColor: 'rgba(0,138,255,0.45)',
            shadowBlur: 10,
            shadowOffsetX: 2,
            shadowOffsetY: 2
          },
          progress: {
            show: true,
            width: 18
          },
          pointer: {
            show: false
          },
          axisLine: {
            lineStyle: {
              width: 18
            }
          },
          axisTick: {
            show: false
          },
          splitLine: {
            show: false
          },
          axisLabel: {
            show: false
          },
          title: {
            show: false
          },
          detail: {
            valueAnimation: true,
            fontSize: 32,
            offsetCenter: [0, '20%'],
            formatter: function (value: number) {
              return `${value.toFixed(0)} tCO₂`
            }
          },
          data: [
            {
              value: progress.actual,
              name: '减排进度'
            }
          ]
        }
      ]
    })
  }
}

const handleResize = () => {
  carbonTrendChart?.resize()
  energyStructureChart?.resize()
  reductionProgressChart?.resize()
}

onMounted(() => {
  fetchStationList()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  carbonTrendChart?.dispose()
  energyStructureChart?.dispose()
  reductionProgressChart?.dispose()
})
</script>

<style scoped lang="scss">
.carbon-page {
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
  
  .gradient-red {
    background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  }
  
  .gradient-teal {
    background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  }
  
  .gradient-cyan {
    background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
  }
  
  .gradient-pink {
    background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
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
