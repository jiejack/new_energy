<template>
  <div class="realtime-chart">
    <div ref="chartRef" class="chart-container"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import * as echarts from 'echarts'
import type { Station } from '@/types'

interface ChartDataPoint {
  time: string
  values: Record<number, number>
}

interface Props {
  data: ChartDataPoint[]
  stations: Station[]
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  data: () => [],
  stations: () => [],
  loading: false
})

const chartRef = ref<HTMLDivElement>()
let chart: echarts.ECharts | null = null

// 颜色配置
const colors = [
  '#409eff',
  '#67c23a',
  '#e6a23c',
  '#f56c6c',
  '#909399',
  '#00d4ff',
  '#ff6b9d',
  '#c792ea'
]

/**
 * 初始化图表
 */
function initChart() {
  if (!chartRef.value) return

  chart = echarts.init(chartRef.value, undefined, {
    renderer: 'canvas'
  })

  updateChart()

  // 监听窗口大小变化
  window.addEventListener('resize', handleResize)
}

/**
 * 更新图表
 */
function updateChart() {
  if (!chart) return

  const series = props.stations.map((station, index) => ({
    name: station.name,
    type: 'line' as const,
    smooth: true,
    symbol: 'none',
    lineStyle: {
      width: 2,
      color: colors[index % colors.length]
    },
    areaStyle: {
      color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
        { offset: 0, color: `${colors[index % colors.length]}40` },
        { offset: 1, color: `${colors[index % colors.length]}05` }
      ])
    },
    data: props.data.map(d => d.values[station.id] ?? null)
  }))

  const option: echarts.EChartsOption = {
    backgroundColor: 'transparent',
    grid: {
      top: 40,
      right: 20,
      bottom: 30,
      left: 50,
      containLabel: false
    },
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(26, 31, 46, 0.95)',
      borderColor: 'rgba(64, 158, 255, 0.3)',
      textStyle: {
        color: '#e5eaf3'
      },
      axisPointer: {
        type: 'cross',
        lineStyle: {
          color: 'rgba(64, 158, 255, 0.5)'
        },
        crossStyle: {
          color: 'rgba(64, 158, 255, 0.5)'
        }
      },
      formatter: (params: any) => {
        if (!Array.isArray(params) || params.length === 0) return ''
        
        let html = `<div style="font-weight: bold; margin-bottom: 8px;">${params[0].axisValue}</div>`
        params.forEach((item: any) => {
          if (item.value !== null && item.value !== undefined) {
            html += `
              <div style="display: flex; align-items: center; gap: 8px; margin: 4px 0;">
                <span style="display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: ${item.color};"></span>
                <span style="flex: 1;">${item.seriesName}</span>
                <span style="font-weight: bold;">${item.value.toFixed(2)} MW</span>
              </div>
            `
          }
        })
        return html
      }
    },
    legend: {
      show: true,
      top: 5,
      right: 10,
      textStyle: {
        color: '#909399',
        fontSize: 11
      },
      itemWidth: 15,
      itemHeight: 8,
      itemGap: 10
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: props.data.map(d => d.time),
      axisLine: {
        lineStyle: {
          color: 'rgba(64, 158, 255, 0.3)'
        }
      },
      axisTick: {
        show: false
      },
      axisLabel: {
        color: '#909399',
        fontSize: 11,
        interval: 'auto',
        rotate: 0
      },
      splitLine: {
        show: false
      }
    },
    yAxis: {
      type: 'value',
      name: '功率(MW)',
      nameTextStyle: {
        color: '#909399',
        fontSize: 11,
        padding: [0, 0, 0, -40]
      },
      axisLine: {
        show: false
      },
      axisTick: {
        show: false
      },
      axisLabel: {
        color: '#909399',
        fontSize: 11,
        formatter: (value: number) => {
          if (value >= 1000) {
            return (value / 1000).toFixed(1) + 'k'
          }
          return value.toFixed(0)
        }
      },
      splitLine: {
        lineStyle: {
          color: 'rgba(64, 158, 255, 0.1)'
        }
      }
    },
    series
  }

  chart.setOption(option, true)
}

/**
 * 处理窗口大小变化
 */
function handleResize() {
  chart?.resize()
}

/**
 * 清空图表
 */
function clearChart() {
  chart?.clear()
}

// 监听数据变化
watch(() => props.data, () => {
  nextTick(() => {
    updateChart()
  })
}, { deep: true })

// 监听电站变化
watch(() => props.stations, () => {
  nextTick(() => {
    updateChart()
  })
}, { deep: true })

// 生命周期
onMounted(() => {
  nextTick(() => {
    initChart()
  })
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  chart?.dispose()
  chart = null
})

// 暴露方法
defineExpose({
  resize: handleResize,
  clear: clearChart
})
</script>

<style scoped lang="scss">
.realtime-chart {
  width: 100%;
  height: 100%;

  .chart-container {
    width: 100%;
    height: 100%;
    min-height: 200px;
  }
}
</style>
