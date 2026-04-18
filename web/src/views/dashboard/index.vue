<template>
  <div class="dashboard-container">
    <!-- 顶部标题栏 -->
    <header class="dashboard-header">
      <div class="header-left">
        <h1 class="title">
          <span class="title-icon">⚡</span>
          <span class="title-text">新能源监控系统</span>
        </h1>
      </div>
      <div class="header-center">
        <div class="datetime">
          <span class="date">{{ currentDate }}</span>
          <span class="time">{{ currentTime }}</span>
          <span class="week">{{ currentWeek }}</span>
        </div>
      </div>
      <div class="header-right">
        <div class="status-indicator">
          <span class="status-dot" :class="{ 'online': wsConnected, 'offline': !wsConnected }"></span>
          <span class="status-text">{{ wsConnected ? '实时连接' : '连接中...' }}</span>
        </div>
        <el-dropdown trigger="click" @command="handleCommand">
          <div class="user-info">
            <div class="avatar-container">
              <el-avatar :size="32" :src="userStore.avatar">
                <el-icon><User /></el-icon>
              </el-avatar>
              <div class="avatar-glow"></div>
            </div>
            <span class="username">{{ userStore.nickname || userStore.username }}</span>
            <el-icon class="arrow"><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人中心</el-dropdown-item>
              <el-dropdown-item command="settings">系统设置</el-dropdown-item>
              <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <!-- 主内容区域 -->
    <main class="dashboard-main">
      <!-- 左侧面板 -->
      <aside class="left-panel">
        <div class="panel station-panel">
          <div class="panel-header">
            <span class="panel-title">电站列表</span>
            <el-tag :type="stations.length > 0 ? 'success' : 'info'" size="small" effect="dark">
              {{ stations.length }} 个
            </el-tag>
          </div>
          <div class="panel-content">
            <StationList
              :stations="stations"
              :loading="stationsLoading"
              @select="handleStationSelect"
            />
          </div>
        </div>
        <div class="panel alarm-panel">
          <div class="panel-header">
            <span class="panel-title">实时告警</span>
            <el-badge :value="alarmCount" :max="99" :hidden="alarmCount === 0">
              <el-tag type="danger" size="small" effect="dark">告警</el-tag>
            </el-badge>
          </div>
          <div class="panel-content">
            <AlarmList
              :alarms="alarms"
              :loading="alarmsLoading"
              @acknowledge="handleAcknowledgeAlarm"
            />
          </div>
        </div>
      </aside>

      <!-- 中间地图区域 -->
      <section class="center-panel">
        <div class="panel map-panel">
          <div class="panel-header">
            <span class="panel-title">电站分布</span>
            <div class="panel-actions">
              <el-button-group size="small">
                <el-button :type="mapType === 'normal' ? 'primary' : ''" @click="mapType = 'normal'" effect="dark">
                  标准
                </el-button>
                <el-button :type="mapType === 'satellite' ? 'primary' : ''" @click="mapType = 'satellite'" effect="dark">
                  卫星
                </el-button>
              </el-button-group>
            </div>
          </div>
          <div class="panel-content">
            <StationMap
              :stations="stations"
              :selected-station="selectedStation"
              :map-type="mapType"
              @select="handleStationSelect"
            />
          </div>
        </div>
      </section>

      <!-- 右侧面板 -->
      <aside class="right-panel">
        <div class="panel stats-panel">
          <div class="panel-header">
            <span class="panel-title">数据统计</span>
            <el-button text size="small" @click="refreshStats" :loading="refreshingStats">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
          <div class="panel-content">
            <StatCards :stats="stats" :loading="statsLoading" />
          </div>
        </div>
        <div class="panel chart-panel">
          <div class="panel-header">
            <span class="panel-title">实时功率曲线</span>
            <div class="panel-actions">
              <el-select v-model="selectedStationIds" multiple collapse-tags collapse-tags-tooltip
                placeholder="选择电站" size="small" style="width: 200px" @change="handleStationChange">
                <el-option v-for="station in stations" :key="station.id" :label="station.name" :value="station.id" />
              </el-select>
            </div>
          </div>
          <div class="panel-content">
            <RealtimeChart
              :data="chartData"
              :stations="selectedStations"
              :loading="chartLoading"
            />
          </div>
        </div>
      </aside>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { User, ArrowDown, Refresh } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useLoadingStore } from '@/stores/loading'
import { getStationList, getStationStatistics } from '@/api/station'
import { getAlarmList, acknowledgeAlarm, getAlarmStatistics } from '@/api/alarm'
import { wsManager } from '@/utils/websocket'
import type { Station, Alarm } from '@/types'
import StationList from './components/StationList.vue'
import AlarmList from './components/AlarmList.vue'
import StationMap from './components/StationMap.vue'
import StatCards from './components/StatCards.vue'
import RealtimeChart from './components/RealtimeChart.vue'

const router = useRouter()
const userStore = useUserStore()
const loadingStore = useLoadingStore()

// 时间相关
const currentDate = ref('')
const currentTime = ref('')
const currentWeek = ref('')
let timeTimer: number | null = null

// 电站相关
const stations = ref<Station[]>([])
const stationsLoading = ref(false)
const selectedStation = ref<Station | null>(null)

// 告警相关
const alarms = ref<Alarm[]>([])
const alarmsLoading = ref(false)
const alarmCount = ref(0)

// 统计数据
const stats = ref({
  totalCapacity: 0,
  currentPower: 0,
  todayEnergy: 0,
  alarmCount: 0,
  onlineRate: 0
})
const statsLoading = ref(false)
const refreshingStats = ref(false)

// 图表相关
const selectedStationIds = ref<number[]>([])
const chartData = ref<Array<{ time: string; values: Record<number, number> }>>([])
const chartLoading = ref(false)

// 地图类型
const mapType = ref<'normal' | 'satellite'>('normal')

// WebSocket连接状态
const wsConnected = ref(false)

// 计算选中的电站
const selectedStations = computed(() => {
  return stations.value.filter(s => selectedStationIds.value.includes(s.id))
})

/**
 * 更新时间
 */
function updateTime() {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')
  const seconds = String(now.getSeconds()).padStart(2, '0')
  const weeks = ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六']

  currentDate.value = `${year}-${month}-${day}`
  currentTime.value = `${hours}:${minutes}:${seconds}`
  currentWeek.value = weeks[now.getDay()]
}

/**
 * 加载电站列表
 */
async function loadStations() {
  stationsLoading.value = true
  try {
    const result = await getStationList({ page: 1, pageSize: 100 })
    stations.value = result.list
    // 默认选中前3个电站用于图表展示
    if (stations.value.length > 0 && selectedStationIds.value.length === 0) {
      selectedStationIds.value = stations.value.slice(0, 3).map(s => s.id)
    }
  } catch (error) {
    console.error('加载电站列表失败:', error)
    ElMessage.error('加载电站列表失败')
  } finally {
    stationsLoading.value = false
  }
}

/**
 * 加载告警列表
 */
async function loadAlarms() {
  alarmsLoading.value = true
  try {
    const result = await getAlarmList({
      page: 1,
      pageSize: 20,
      status: 'active'
    })
    alarms.value = result.list
  } catch (error) {
    console.error('加载告警列表失败:', error)
  } finally {
    alarmsLoading.value = false
  }
}

/**
 * 加载告警统计
 */
async function loadAlarmStatistics() {
  try {
    const result = await getAlarmStatistics()
    alarmCount.value = result.total
    stats.value.alarmCount = result.total
  } catch (error) {
    console.error('加载告警统计失败:', error)
  }
}

/**
 * 加载统计数据
 */
async function loadStatistics() {
  statsLoading.value = true
  try {
    // 计算总装机容量
    const totalCapacity = stations.value.reduce((sum, s) => sum + (s.capacity || 0), 0)

    // 获取各电站统计信息
    let totalPower = 0
    let totalEnergy = 0
    let totalDevices = 0
    let onlineDevices = 0

    for (const station of stations.value) {
      try {
        const stationStats = await getStationStatistics(station.id)
        totalPower += stationStats.power || 0
        totalEnergy += stationStats.energy || 0
        totalDevices += stationStats.deviceCount || 0
        onlineDevices += stationStats.onlineDeviceCount || 0
      } catch (error) {
        console.error(`获取电站 ${station.id} 统计失败:`, error)
      }
    }

    stats.value = {
      totalCapacity,
      currentPower: totalPower,
      todayEnergy: totalEnergy,
      alarmCount: alarmCount.value,
      onlineRate: totalDevices > 0 ? (onlineDevices / totalDevices) * 100 : 0
    }
  } catch (error) {
    console.error('加载统计数据失败:', error)
  } finally {
    statsLoading.value = false
  }
}

/**
 * 刷新统计数据
 */
async function refreshStats() {
  const actionKey = 'refresh-stats'
  loadingStore.setActionLoading(actionKey, true)
  refreshingStats.value = true
  
  try {
    await Promise.all([
      loadAlarmStatistics(),
      loadStatistics()
    ])
    ElMessage.success('数据已刷新')
  } finally {
    loadingStore.setActionLoading(actionKey, false)
    refreshingStats.value = false
  }
}

/**
 * 处理电站选择
 */
function handleStationSelect(station: Station) {
  selectedStation.value = station
}

/**
 * 处理告警确认
 */
async function handleAcknowledgeAlarm(alarm: Alarm) {
  try {
    await acknowledgeAlarm(alarm.id)
    ElMessage.success('告警已确认')
    await loadAlarms()
    await loadAlarmStatistics()
  } catch (error) {
    console.error('确认告警失败:', error)
    ElMessage.error('确认告警失败')
  }
}

/**
 * 处理电站变化
 */
function handleStationChange() {
  // 重新订阅实时数据（只有在连接时才发送）
  if (wsConnected.value && wsManager.isConnected()) {
    wsManager.send('subscribe-power', { stationIds: selectedStationIds.value })
  }
}

/**
 * 处理下拉菜单命令
 */
function handleCommand(command: string) {
  switch (command) {
    case 'profile':
      router.push('/profile')
      break
    case 'settings':
      router.push('/settings')
      break
    case 'logout':
      handleLogout()
      break
  }
}

/**
 * 处理登出
 */
async function handleLogout() {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await userStore.logoutAction()
    router.push('/login')
    ElMessage.success('已退出登录')
  } catch (error) {
    // 用户取消
  }
}

/**
 * 初始化WebSocket
 */
async function initWebSocket() {
  try {
    // 先设置连接成功回调，在回调中发送订阅消息
    wsManager.onConnect(() => {
      wsConnected.value = true
      // 订阅实时数据（在连接成功后发送）
      wsManager.send('subscribe-power', { stationIds: selectedStationIds.value })
      wsManager.subscribeAlarm()
      ElMessage.success('WebSocket已连接')
    })

    wsManager.onDisconnect(() => {
      wsConnected.value = false
      ElMessage.warning('WebSocket连接断开，正在重连...')
    })

    // 连接WebSocket
    await wsManager.connect()

    // 监听实时数据
    wsManager.on('realtime-power', (msg: any) => {
      const data = msg.payload || msg
      // 更新图表数据
      const now = new Date()
      const time = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`

      chartData.value.push({
        time,
        values: data
      })

      // 保留最近100条数据
      if (chartData.value.length > 100) {
        chartData.value.shift()
      }
    })

    // 监听告警
    wsManager.on('alarm', (msg: any) => {
      const alarm = msg.payload || msg
      alarms.value.unshift(alarm)
      alarmCount.value++
      stats.value.alarmCount++

      // 保持最多20条
      if (alarms.value.length > 20) {
        alarms.value.pop()
      }

      ElMessage.warning(`新告警: ${alarm.title || alarm.message || '新告警'}`)
    })
  } catch (error) {
    console.error('WebSocket连接失败:', error)
    ElMessage.error('实时数据连接失败')
  }
}

/**
 * 初始化
 */
async function init() {
  loadingStore.setPageLoading(true, '加载数据中...')
  
  try {
    // 更新时间
    updateTime()
    timeTimer = window.setInterval(updateTime, 1000)

    // 加载数据
    await loadStations()
    await Promise.all([
      loadAlarms(),
      loadAlarmStatistics(),
      loadStatistics()
    ])

    // 初始化WebSocket
    await initWebSocket()
  } finally {
    loadingStore.setPageLoading(false)
  }
}

// 生命周期
onMounted(() => {
  init()
})

onUnmounted(() => {
  if (timeTimer) {
    clearInterval(timeTimer)
  }
  wsManager.disconnect()
})
</script>

<style scoped lang="scss">
/* ============================================
   新能源监控系统 - Dashboard 样式
   ============================================ */
@import url('https://fonts.googleapis.com/css2?family=Orbitron:wght@400;500;600;700&family=Rajdhani:wght@300;400;500;600;700&display=swap');

.dashboard-container {
  width: 100%;
  height: 100vh;
  background: linear-gradient(135deg, #0d1117 0%, #080b0f 100%);
  color: #e5eaf3;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
  font-family: 'Rajdhani', sans-serif;

  /* 背景装饰 - 电网纹理 */
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-image: 
      linear-gradient(rgba(0, 212, 170, 0.03) 1px, transparent 1px),
      linear-gradient(90deg, rgba(0, 212, 170, 0.03) 1px, transparent 1px);
    background-size: 50px 50px;
    pointer-events: none;
    z-index: 0;
  }

  /* 背景装饰 - 发光球体 */
  &::after {
    content: '';
    position: absolute;
    top: -50%;
    right: -20%;
    width: 800px;
    height: 800px;
    background: radial-gradient(circle, rgba(124, 58, 237, 0.05) 0%, transparent 70%);
    border-radius: 50%;
    filter: blur(80px);
    pointer-events: none;
    z-index: 0;
  }
}

// 顶部标题栏
.dashboard-header {
  height: 70px;
  background: linear-gradient(90deg, rgba(26, 31, 46, 0.95) 0%, rgba(13, 17, 23, 0.95) 100%);
  border-bottom: 1px solid rgba(0, 212, 170, 0.3);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 30px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
  position: relative;
  z-index: 10;
  backdrop-filter: blur(10px);

  /* 顶部装饰线 */
  &::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 3px;
    background: linear-gradient(90deg, #00d4aa, #7c3aed);
    box-shadow: 0 0 15px rgba(0, 212, 170, 0.6);
  }

  .header-left {
    .title {
      font-size: 28px;
      font-weight: 700;
      display: flex;
      align-items: center;
      gap: 15px;
      margin: 0;
      font-family: 'Orbitron', sans-serif;

      .title-icon {
        font-size: 32px;
        animation: pulse 2s infinite;
      }

      .title-text {
        background: linear-gradient(90deg, #00d4aa, #7c3aed);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
        letter-spacing: 2px;
        text-shadow: 0 0 30px rgba(0, 212, 170, 0.4);
      }
    }
  }

  .header-center {
    .datetime {
      display: flex;
      align-items: center;
      gap: 25px;
      font-size: 18px;
      font-family: 'Orbitron', sans-serif;

      .date {
        color: #94a3b8;
        font-weight: 500;
      }

      .time {
        font-size: 32px;
        font-weight: 700;
        color: #00d4aa;
        text-shadow: 0 0 20px rgba(0, 212, 170, 0.6);
        animation: glow 3s ease-in-out infinite alternate;
      }

      .week {
        color: #4ade80;
        font-weight: 500;
      }
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 20px;

    .status-indicator {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 6px 12px;
      background: rgba(26, 31, 46, 0.8);
      border: 1px solid rgba(0, 212, 170, 0.2);
      border-radius: 20px;

      .status-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        transition: all 0.3s ease;

        &.online {
          background: #4ade80;
          box-shadow: 0 0 10px rgba(74, 222, 128, 0.8);
        }

        &.offline {
          background: #f87171;
          box-shadow: 0 0 10px rgba(248, 113, 113, 0.8);
          animation: pulse 1.5s infinite;
        }
      }

      .status-text {
        font-size: 14px;
        color: #94a3b8;
      }
    }

    .user-info {
      display: flex;
      align-items: center;
      gap: 12px;
      cursor: pointer;
      padding: 8px 20px;
      border-radius: 25px;
      transition: all 0.3s ease;
      border: 1px solid transparent;
      background: rgba(26, 31, 46, 0.8);

      &:hover {
        background: rgba(0, 212, 170, 0.1);
        border-color: rgba(0, 212, 170, 0.4);
        box-shadow: 0 0 20px rgba(0, 212, 170, 0.2);
        transform: translateY(-2px);
      }

      .avatar-container {
        position: relative;

        .avatar-glow {
          position: absolute;
          top: -2px;
          left: -2px;
          right: -2px;
          bottom: -2px;
          background: linear-gradient(45deg, #00d4aa, #7c3aed);
          border-radius: 50%;
          z-index: -1;
          animation: rotate 3s linear infinite;
          opacity: 0.7;
        }
      }

      .username {
        color: #e5eaf3;
        font-size: 16px;
        font-weight: 500;
      }

      .arrow {
        color: #94a3b8;
        font-size: 14px;
        transition: transform 0.3s ease;
      }

      &:hover .arrow {
        transform: rotate(180deg);
      }
    }
  }
}

// 主内容区域
.dashboard-main {
  flex: 1;
  display: flex;
  gap: 20px;
  padding: 20px;
  overflow: hidden;
  position: relative;
  z-index: 1;
}

// 面板通用样式 - 新能源监控专用
.panel {
  background: linear-gradient(135deg, rgba(26, 31, 46, 0.85) 0%, rgba(13, 17, 23, 0.9) 100%);
  border: 1px solid rgba(0, 212, 170, 0.2);
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(15px);
  transition: all 0.3s ease;
  position: relative;
  animation: fadeInUp 0.6s ease-out;
  animation-fill-mode: both;

  /* 面板发光效果 */
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 3px;
    background: linear-gradient(90deg, #00d4aa, #7c3aed);
    opacity: 0.7;
  }

  &:hover {
    border-color: rgba(0, 212, 170, 0.4);
    box-shadow: 0 12px 40px rgba(0, 212, 170, 0.2);
    transform: translateY(-5px);
  }

  .panel-header {
    height: 50px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 20px;
    background: rgba(0, 212, 170, 0.05);
    border-bottom: 1px solid rgba(0, 212, 170, 0.15);
    flex-shrink: 0;

    .panel-title {
      font-size: 16px;
      font-weight: 600;
      color: #00d4aa;
      display: flex;
      align-items: center;
      gap: 10px;
      letter-spacing: 0.5px;
      font-family: 'Orbitron', sans-serif;

      &::before {
        content: '';
        width: 4px;
        height: 18px;
        background: linear-gradient(180deg, #00d4aa, #7c3aed);
        border-radius: 2px;
        box-shadow: 0 0 10px rgba(0, 212, 170, 0.6);
      }
    }

    .panel-actions {
      display: flex;
      align-items: center;
      gap: 12px;
    }
  }

  .panel-content {
    flex: 1;
    overflow: hidden;
    padding: 15px;
  }
}

// 左侧面板
.left-panel {
  width: 340px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  flex-shrink: 0;

  .station-panel {
    flex: 1;
    animation-delay: 0.1s;
  }

  .alarm-panel {
    flex: 1;
    animation-delay: 0.2s;
  }
}

// 中间地图区域
.center-panel {
  flex: 1;
  min-width: 0;

  .map-panel {
    height: 100%;
    animation-delay: 0.3s;
  }
}

// 右侧面板
.right-panel {
  width: 400px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  flex-shrink: 0;

  .stats-panel {
    flex-shrink: 0;
    height: auto;
    animation-delay: 0.4s;
  }

  .chart-panel {
    flex: 1;
    min-height: 0;
    animation-delay: 0.5s;
  }
}

// 响应式布局
@media (max-width: 1600px) {
  .left-panel {
    width: 300px;
  }

  .right-panel {
    width: 360px;
  }
}

@media (max-width: 1200px) {
  .dashboard-main {
    flex-wrap: wrap;
  }

  .left-panel,
  .right-panel {
    width: calc(50% - 10px);
  }

  .center-panel {
    width: 100%;
    order: -1;
    height: 450px;
  }
}

// 深色主题覆盖
:deep(.el-dropdown-menu) {
  background: linear-gradient(135deg, rgba(26, 31, 46, 0.95) 0%, rgba(13, 17, 23, 0.95) 100%);
  border: 1px solid rgba(0, 212, 170, 0.3);
  border-radius: 12px;
  backdrop-filter: blur(10px);

  .el-dropdown-menu__item {
    color: #e5eaf3;
    padding: 10px 20px;
    transition: all 0.3s ease;

    &:hover {
      background: rgba(0, 212, 170, 0.15);
      color: #00d4aa;
      transform: translateX(5px);
    }
  }
}

:deep(.el-select-dropdown) {
  background: linear-gradient(135deg, rgba(26, 31, 46, 0.95) 0%, rgba(13, 17, 23, 0.95) 100%);
  border: 1px solid rgba(0, 212, 170, 0.3);
  border-radius: 12px;
  backdrop-filter: blur(10px);

  .el-select-dropdown__item {
    color: #e5eaf3;
    padding: 10px 15px;
    transition: all 0.3s ease;

    &.selected,
    &:hover {
      background: rgba(0, 212, 170, 0.15);
      color: #00d4aa;
    }
  }
}

:deep(.el-button-group) {
  .el-button {
    background: rgba(26, 31, 46, 0.8);
    border: 1px solid rgba(0, 212, 170, 0.3);
    color: #e5eaf3;
    transition: all 0.3s ease;

    &:hover {
      background: rgba(0, 212, 170, 0.2);
      border-color: #00d4aa;
    }

    &.el-button--primary {
      background: linear-gradient(90deg, #00d4aa, #7c3aed);
      border: none;
      box-shadow: 0 4px 15px rgba(0, 212, 170, 0.4);

      &:hover {
        box-shadow: 0 6px 20px rgba(0, 212, 170, 0.6);
        transform: translateY(-2px);
      }
    }
  }
}

:deep(.el-tag) {
  &.el-tag--success {
    background: rgba(74, 222, 128, 0.2);
    border: 1px solid rgba(74, 222, 128, 0.4);
    color: #4ade80;
  }

  &.el-tag--danger {
    background: rgba(248, 113, 113, 0.2);
    border: 1px solid rgba(248, 113, 113, 0.4);
    color: #f87171;
  }
}

// 动画效果
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

@keyframes glow {
  from {
    text-shadow: 0 0 20px rgba(0, 212, 170, 0.6);
  }
  to {
    text-shadow: 0 0 30px rgba(0, 212, 170, 0.9), 0 0 40px rgba(124, 58, 237, 0.6);
  }
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>