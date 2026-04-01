<template>
  <div ref="chartRef" class="data-chart" :style="{ height: height }"></div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { echarts } from '@/plugins/echarts'
import type { EChartsOption } from 'echarts'
import type { PointData } from '@/types'

const props = defineProps<{
  data: PointData[]
  height?: string
  showLegend?: boolean
  showDataZoom?: boolean
  showToolbox?: boolean
  title?: string
}>()

const chartRef = ref<HTMLElement>()
let chartInstance: echarts.ECharts | null = null

// 颜色配置
const colors = [
  '#5470c6',
  '#91cc75',
  '#fac858',
  '#ee6666',
  '#73c0de',
  '#3ba272',
  '#fc8452',
  '#9a60b4',
  '#ea7ccc',
]

// 获取图表配置
const getChartOption = (): EChartsOption => {
  const series: any[] = []
  const legendData: string[] = []

  props.data.forEach((pointData) => {
    legendData.push(pointData.pointName)
    series.push({
      name: pointData.pointName,
      type: 'line',
      smooth: true,
      symbol: 'none',
      sampling: 'lttb',
      lineStyle: {
        width: 2,
      },
      areaStyle: {
        opacity: 0.1,
      },
      emphasis: {
        focus: 'series',
      },
      data: pointData.data.map((item) => [item.timestamp, item.value]),
    })
  })

  const option: EChartsOption = {
    title: props.title
      ? {
          text: props.title,
          left: 'center',
        }
      : undefined,
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross',
        label: {
          backgroundColor: '#6a7985',
        },
      },
      formatter: (params: any) => {
        if (!Array.isArray(params) || params.length === 0) return ''
        const time = params[0].data[0]
        let html = `<div style="font-weight: bold; margin-bottom: 4px;">${time}</div>`
        params.forEach((param: any) => {
          const pointData = props.data.find((p) => p.pointName === param.seriesName)
          const unit = pointData?.unit || ''
          html += `
            <div style="display: flex; justify-content: space-between; gap: 20px;">
              <span>${param.marker} ${param.seriesName}</span>
              <span style="font-weight: bold;">${param.data[1]?.toFixed(2) ?? '-'} ${unit}</span>
            </div>
          `
        })
        return html
      },
    },
    legend: props.showLegend
      ? {
          data: legendData,
          top: 30,
          type: 'scroll',
        }
      : undefined,
    grid: {
      left: '3%',
      right: '4%',
      bottom: props.showDataZoom ? 80 : 20,
      top: props.showLegend ? 60 : 20,
      containLabel: true,
    },
    toolbox: props.showToolbox
      ? {
          feature: {
            dataZoom: {
              yAxisIndex: 'none',
            },
            restore: {},
            saveAsImage: {},
          },
          right: 20,
        }
      : undefined,
    dataZoom: props.showDataZoom
      ? [
          {
            type: 'inside',
            start: 0,
            end: 100,
          },
          {
            start: 0,
            end: 100,
          },
        ]
      : undefined,
    xAxis: {
      type: 'time',
      axisLine: {
        lineStyle: {
          color: '#999',
        },
      },
      axisLabel: {
        formatter: (value: number) => {
          const date = new Date(value)
          const hours = date.getHours().toString().padStart(2, '0')
          const minutes = date.getMinutes().toString().padStart(2, '0')
          const month = (date.getMonth() + 1).toString().padStart(2, '0')
          const day = date.getDate().toString().padStart(2, '0')
          return `${month}-${day} ${hours}:${minutes}`
        },
      },
    },
    yAxis: {
      type: 'value',
      axisLine: {
        show: false,
      },
      axisTick: {
        show: false,
      },
      splitLine: {
        lineStyle: {
          type: 'dashed',
        },
      },
    },
    series,
    color: colors,
  }

  return option
}

// 初始化图表
const initChart = () => {
  if (!chartRef.value) return

  if (chartInstance) {
    chartInstance.dispose()
  }

  chartInstance = echarts.init(chartRef.value)
  chartInstance.setOption(getChartOption())
}

// 更新图表
const updateChart = () => {
  if (!chartInstance) {
    initChart()
    return
  }

  chartInstance.setOption(getChartOption(), true)
}

// 处理窗口大小变化
const handleResize = () => {
  chartInstance?.resize()
}

// 监听数据变化
watch(
  () => props.data,
  () => {
    nextTick(() => {
      updateChart()
    })
  },
  { deep: true }
)

// 组件挂载
onMounted(() => {
  nextTick(() => {
    initChart()
  })
  window.addEventListener('resize', handleResize)
})

// 组件卸载
onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  chartInstance?.dispose()
})
</script>

<style scoped lang="scss">
.data-chart {
  width: 100%;
  min-height: 300px;
}
</style>
