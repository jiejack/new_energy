<template>
  <div class="station-map">
    <div ref="mapContainerRef" class="map-container"></div>

    <!-- 电站详情弹窗 -->
    <transition name="fade">
      <div v-if="showPopup && popupStation" class="station-popup" :style="popupStyle">
        <div class="popup-header">
          <span class="popup-title">{{ popupStation.name }}</span>
          <el-icon class="popup-close" @click="closePopup"><Close /></el-icon>
        </div>
        <div class="popup-content">
          <div class="popup-row">
            <span class="popup-label">类型:</span>
            <span class="popup-value">{{ getStationTypeName(popupStation.type) }}</span>
          </div>
          <div class="popup-row">
            <span class="popup-label">容量:</span>
            <span class="popup-value">{{ popupStation.capacity }} MW</span>
          </div>
          <div class="popup-row">
            <span class="popup-label">状态:</span>
            <el-tag :type="getStatusType(popupStation.status)" size="small">
              {{ getStatusName(popupStation.status) }}
            </el-tag>
          </div>
          <div class="popup-row">
            <span class="popup-label">地址:</span>
            <span class="popup-value">{{ popupStation.address }}</span>
          </div>
        </div>
        <div class="popup-footer">
          <el-button type="primary" size="small" @click="viewDetail(popupStation)">
            查看详情
          </el-button>
        </div>
      </div>
    </transition>

    <!-- 地图控制 -->
    <div class="map-controls">
      <el-button-group>
        <el-button size="small" @click="zoomIn">
          <el-icon><Plus /></el-icon>
        </el-button>
        <el-button size="small" @click="zoomOut">
          <el-icon><Minus /></el-icon>
        </el-button>
        <el-button size="small" @click="resetView">
          <el-icon><Aim /></el-icon>
        </el-button>
      </el-button-group>
    </div>

    <!-- 图例 -->
    <div class="map-legend">
      <div class="legend-title">图例</div>
      <div class="legend-items">
        <div class="legend-item">
          <span class="legend-marker online"></span>
          <span class="legend-label">在线</span>
        </div>
        <div class="legend-item">
          <span class="legend-marker offline"></span>
          <span class="legend-label">离线</span>
        </div>
        <div class="legend-item">
          <span class="legend-marker maintenance"></span>
          <span class="legend-label">维护</span>
        </div>
        <div class="legend-item">
          <span class="legend-marker fault"></span>
          <span class="legend-label">故障</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { Close, Plus, Minus, Aim } from '@element-plus/icons-vue'
import type { Station, StationType, StationStatus } from '@/types'

interface Props {
  stations: Station[]
  selectedStation?: Station | null
  mapType?: 'normal' | 'satellite'
}

const props = withDefaults(defineProps<Props>(), {
  selectedStation: null,
  mapType: 'normal'
})

const emit = defineEmits<{
  select: [station: Station]
}>()

const mapContainerRef = ref<HTMLDivElement>()
const showPopup = ref(false)
const popupStation = ref<Station | null>(null)
const popupStyle = ref({ left: '0px', top: '0px' })

// 地图状态
let map: any = null
let markers: Map<number, any> = new Map()
let centerLng = 116.404
let centerLat = 39.915
let zoom = 5

/**
 * 获取电站类型名称
 */
function getStationTypeName(type: StationType) {
  const nameMap: Record<StationType, string> = {
    solar: '光伏电站',
    wind: '风电场',
    hydro: '水电站',
    storage: '储能站'
  }
  return nameMap[type] || '未知'
}

/**
 * 获取状态类型
 */
function getStatusType(status: StationStatus) {
  const typeMap: Record<StationStatus, 'success' | 'warning' | 'danger' | 'info'> = {
    online: 'success',
    offline: 'danger',
    maintenance: 'warning',
    fault: 'danger'
  }
  return typeMap[status] || 'info'
}

/**
 * 获取状态名称
 */
function getStatusName(status: StationStatus) {
  const nameMap: Record<StationStatus, string> = {
    online: '在线',
    offline: '离线',
    maintenance: '维护',
    fault: '故障'
  }
  return nameMap[status] || '未知'
}

/**
 * 获取标记颜色
 */
function getMarkerColor(status: StationStatus) {
  const colorMap: Record<StationStatus, string> = {
    online: '#67c23a',
    offline: '#f56c6c',
    maintenance: '#e6a23c',
    fault: '#f56c6c'
  }
  return colorMap[status] || '#909399'
}

/**
 * 初始化地图
 */
function initMap() {
  if (!mapContainerRef.value) return

  // 创建canvas地图
  const canvas = document.createElement('canvas')
  canvas.style.width = '100%'
  canvas.style.height = '100%'
  mapContainerRef.value.appendChild(canvas)

  map = {
    canvas,
    ctx: canvas.getContext('2d'),
    width: 0,
    height: 0
  }

  // 设置canvas尺寸
  resizeCanvas()

  // 绑定事件
  canvas.addEventListener('click', handleMapClick)
  canvas.addEventListener('mousemove', handleMapMouseMove)
  window.addEventListener('resize', resizeCanvas)

  // 绘制地图
  drawMap()
}

/**
 * 调整canvas尺寸
 */
function resizeCanvas() {
  if (!map || !mapContainerRef.value) return

  const rect = mapContainerRef.value.getBoundingClientRect()
  map.width = rect.width
  map.height = rect.height
  map.canvas.width = rect.width * window.devicePixelRatio
  map.canvas.height = rect.height * window.devicePixelRatio
  map.ctx.scale(window.devicePixelRatio, window.devicePixelRatio)

  drawMap()
}

/**
 * 绘制地图
 */
function drawMap() {
  if (!map) return

  const { ctx, width, height } = map

  // 清空画布
  ctx.clearRect(0, 0, width, height)

  // 绘制背景
  const gradient = ctx.createLinearGradient(0, 0, width, height)
  if (props.mapType === 'satellite') {
    gradient.addColorStop(0, '#1a2a3a')
    gradient.addColorStop(1, '#0d1520')
  } else {
    gradient.addColorStop(0, '#1e3a5f')
    gradient.addColorStop(1, '#0d1f35')
  }
  ctx.fillStyle = gradient
  ctx.fillRect(0, 0, width, height)

  // 绘制网格
  drawGrid(ctx, width, height)

  // 绘制电站标记
  drawMarkers(ctx, width, height)
}

/**
 * 绘制网格
 */
function drawGrid(ctx: CanvasRenderingContext2D, width: number, height: number) {
  ctx.strokeStyle = 'rgba(64, 158, 255, 0.1)'
  ctx.lineWidth = 1

  // 绘制经线
  for (let i = 0; i <= 18; i++) {
    const x = (width / 18) * i
    ctx.beginPath()
    ctx.moveTo(x, 0)
    ctx.lineTo(x, height)
    ctx.stroke()
  }

  // 绘制纬线
  for (let i = 0; i <= 9; i++) {
    const y = (height / 9) * i
    ctx.beginPath()
    ctx.moveTo(0, y)
    ctx.lineTo(width, y)
    ctx.stroke()
  }
}

/**
 * 绘制电站标记
 */
function drawMarkers(ctx: CanvasRenderingContext2D, width: number, height: number) {
  markers.clear()

  props.stations.forEach(station => {
    // 将经纬度转换为画布坐标
    const x = ((station.longitude - (centerLng - 180 / zoom)) / (360 / zoom)) * width
    const y = ((centerLat + 90 / zoom - station.latitude) / (180 / zoom)) * height

    // 检查是否在可视范围内
    if (x < 0 || x > width || y < 0 || y > height) return

    const color = getMarkerColor(station.status)
    const isSelected = props.selectedStation?.id === station.id
    const radius = isSelected ? 12 : 8

    // 绘制外圈光晕
    ctx.beginPath()
    ctx.arc(x, y, radius + 4, 0, Math.PI * 2)
    ctx.fillStyle = `${color}33`
    ctx.fill()

    // 绘制标记点
    ctx.beginPath()
    ctx.arc(x, y, radius, 0, Math.PI * 2)
    ctx.fillStyle = color
    ctx.fill()

    // 绘制边框
    if (isSelected) {
      ctx.strokeStyle = '#fff'
      ctx.lineWidth = 2
      ctx.stroke()
    }

    // 绘制电站名称
    ctx.font = '12px sans-serif'
    ctx.fillStyle = '#e5eaf3'
    ctx.textAlign = 'center'
    ctx.fillText(station.name, x, y + radius + 16)

    // 保存标记位置信息
    markers.set(station.id, { x, y, radius, station })
  })
}

/**
 * 处理地图点击
 */
function handleMapClick(event: MouseEvent) {
  if (!map) return

  const rect = map.canvas.getBoundingClientRect()
  const x = event.clientX - rect.left
  const y = event.clientY - rect.top

  // 检查是否点击了标记
  for (const [, marker] of markers) {
    const distance = Math.sqrt(Math.pow(x - marker.x, 2) + Math.pow(y - marker.y, 2))
    if (distance <= marker.radius + 4) {
      showStationPopup(marker.station, event)
      emit('select', marker.station)
      return
    }
  }

  // 点击空白处关闭弹窗
  closePopup()
}

/**
 * 处理鼠标移动
 */
function handleMapMouseMove(event: MouseEvent) {
  if (!map) return

  const rect = map.canvas.getBoundingClientRect()
  const x = event.clientX - rect.left
  const y = event.clientY - rect.top

  let isOverMarker = false
  for (const [, marker] of markers) {
    const distance = Math.sqrt(Math.pow(x - marker.x, 2) + Math.pow(y - marker.y, 2))
    if (distance <= marker.radius + 4) {
      isOverMarker = true
      break
    }
  }

  map.canvas.style.cursor = isOverMarker ? 'pointer' : 'default'
}

/**
 * 显示电站弹窗
 */
function showStationPopup(station: Station, event: MouseEvent) {
  popupStation.value = station
  showPopup.value = true

  nextTick(() => {
    const rect = mapContainerRef.value?.getBoundingClientRect()
    if (!rect) return

    let left = event.clientX - rect.left + 10
    let top = event.clientY - rect.top + 10

    // 防止超出边界
    const popupWidth = 200
    const popupHeight = 180

    if (left + popupWidth > rect.width) {
      left = left - popupWidth - 20
    }
    if (top + popupHeight > rect.height) {
      top = top - popupHeight - 20
    }

    popupStyle.value = {
      left: `${left}px`,
      top: `${top}px`
    }
  })
}

/**
 * 关闭弹窗
 */
function closePopup() {
  showPopup.value = false
  popupStation.value = null
}

/**
 * 查看详情
 */
function viewDetail(station: Station) {
  emit('select', station)
  closePopup()
}

/**
 * 放大
 */
function zoomIn() {
  zoom = Math.min(zoom + 1, 15)
  drawMap()
}

/**
 * 缩小
 */
function zoomOut() {
  zoom = Math.max(zoom - 1, 3)
  drawMap()
}

/**
 * 重置视图
 */
function resetView() {
  zoom = 5
  centerLng = 116.404
  centerLat = 39.915
  drawMap()
}

// 监听电站列表变化
watch(() => props.stations, () => {
  drawMap()
}, { deep: true })

// 监听选中电站变化
watch(() => props.selectedStation, () => {
  drawMap()
})

// 监听地图类型变化
watch(() => props.mapType, () => {
  drawMap()
})

// 生命周期
onMounted(() => {
  initMap()
})

onUnmounted(() => {
  if (map) {
    map.canvas.removeEventListener('click', handleMapClick)
    map.canvas.removeEventListener('mousemove', handleMapMouseMove)
    window.removeEventListener('resize', resizeCanvas)
  }
})
</script>

<style scoped lang="scss">
.station-map {
  position: relative;
  width: 100%;
  height: 100%;
  border-radius: 4px;
  overflow: hidden;

  .map-container {
    width: 100%;
    height: 100%;
  }

  .station-popup {
    position: absolute;
    width: 200px;
    background: rgba(26, 31, 46, 0.95);
    border: 1px solid rgba(64, 158, 255, 0.3);
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    z-index: 100;

    .popup-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 10px 12px;
      border-bottom: 1px solid rgba(64, 158, 255, 0.2);

      .popup-title {
        font-size: 14px;
        font-weight: bold;
        color: #409eff;
      }

      .popup-close {
        cursor: pointer;
        color: #909399;
        transition: color 0.3s;

        &:hover {
          color: #f56c6c;
        }
      }
    }

    .popup-content {
      padding: 12px;

      .popup-row {
        display: flex;
        align-items: center;
        margin-bottom: 8px;

        &:last-child {
          margin-bottom: 0;
        }

        .popup-label {
          width: 50px;
          font-size: 12px;
          color: #909399;
        }

        .popup-value {
          flex: 1;
          font-size: 12px;
          color: #e5eaf3;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
      }
    }

    .popup-footer {
      padding: 8px 12px;
      border-top: 1px solid rgba(64, 158, 255, 0.2);
      text-align: right;
    }
  }

  .map-controls {
    position: absolute;
    top: 10px;
    right: 10px;
    z-index: 10;

    :deep(.el-button-group) {
      .el-button {
        background-color: rgba(26, 31, 46, 0.9);
        border-color: rgba(64, 158, 255, 0.3);
        color: #e5eaf3;

        &:hover {
          background-color: rgba(64, 158, 255, 0.3);
        }
      }
    }
  }

  .map-legend {
    position: absolute;
    bottom: 10px;
    left: 10px;
    background: rgba(26, 31, 46, 0.9);
    border: 1px solid rgba(64, 158, 255, 0.3);
    border-radius: 6px;
    padding: 10px;
    z-index: 10;

    .legend-title {
      font-size: 12px;
      font-weight: bold;
      color: #409eff;
      margin-bottom: 8px;
    }

    .legend-items {
      display: flex;
      flex-direction: column;
      gap: 6px;

      .legend-item {
        display: flex;
        align-items: center;
        gap: 8px;

        .legend-marker {
          width: 12px;
          height: 12px;
          border-radius: 50%;

          &.online {
            background-color: #67c23a;
          }

          &.offline {
            background-color: #f56c6c;
          }

          &.maintenance {
            background-color: #e6a23c;
          }

          &.fault {
            background-color: #f56c6c;
          }
        }

        .legend-label {
          font-size: 11px;
          color: #909399;
        }
      }
    }
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s, transform 0.3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
