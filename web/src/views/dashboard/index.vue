<template>
  <div class="dashboard-container">
    <!-- 顶部标题栏 -->
    <header class="dashboard-header">
      <div class="header-left">
        <h1 class="title">新能源监控系统</h1>
      </div>
      <div class="header-center">
        <div class="datetime">
          <span class="date">{{ currentDate }}</span>
          <span class="time">{{ currentTime }}</span>
          <span class="week">{{ currentWeek }}</span>
        </div>
      </div>
      <div class="header-right">
        <el-dropdown trigger="click" @command="handleCommand">
          <div class="user-info">
            <el-avatar :size="32" :src="userStore.avatar">
              <el-icon><User /></el-icon>
            </el-avatar>
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
            <el-tag :type="stations.length > 0 ? 'success' : 'info'" size="small">
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
              <el-tag type="danger" size="small">告警</el-tag>
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
                <el-button :type="mapType === 'normal' ? 'primary' : ''" @click="mapType = 'normal'">
                  标准
                </el-button>
                <el-button :type="mapType === 'satellite' ? 'primary' : ''" @click="mapType = 'satellite'">
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
            <el-button text size="small" @click="refreshStats">
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

// 图表相关
const selectedStationIds = ref<number[]>([])
const chartData = ref<Array<{ time: string; values: Record<number, number> }>>([])
const chartLoading = ref(false)

// 地图类型
const mapType = ref<'normal' | 'satellite'>('normal')

// WebSocket连接状态
let wsConnected = false

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
  await Promise.all([
    loadAlarmStatistics(),
    loadStatistics()
  ])
  ElMessage.success('数据已刷新')
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
  if (wsConnected && wsManager.isConnected()) {
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
      wsConnected = true
      // 订阅实时数据（在连接成功后发送）
      wsManager.send('subscribe-power', { stationIds: selectedStationIds.value })
      wsManager.subscribeAlarm()
      ElMessage.success('WebSocket已连接')
    })

    wsManager.onDisconnect(() => {
      wsConnected = false
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
.dashboard-container {
  width: 100%;
  height: 100vh;
  background: linear-gradient(135deg, #1a1f2e 0%, #0d1117 100%);
  color: #e5eaf3;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

// 顶部标题栏
.dashboard-header {
  height: 60px;
  background: linear-gradient(90deg, rgba(32, 45, 65, 0.95) 0%, rgba(20, 30, 48, 0.95) 100%);
  border-bottom: 1px solid rgba(64, 158, 255, 0.3);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);

  .header-left {
    .title {
      font-size: 24px;
      font-weight: bold;
      background: linear-gradient(90deg, #409eff, #67c23a);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
      margin: 0;
      letter-spacing: 2px;
    }
  }

  .header-center {
    .datetime {
      display: flex;
      align-items: center;
      gap: 20px;
      font-size: 16px;

      .date {
        color: #909399;
      }

      .time {
        font-size: 24px;
        font-weight: bold;
        color: #409eff;
        font-family: 'Courier New', monospace;
      }

      .week {
        color: #67c23a;
      }
    }
  }

  .header-right {
    .user-info {
      display: flex;
      align-items: center;
      gap: 10px;
      cursor: pointer;
      padding: 5px 15px;
      border-radius: 20px;
      transition: background-color 0.3s;

      &:hover {
        background-color: rgba(255, 255, 255, 0.1);
      }

      .username {
        color: #e5eaf3;
        font-size: 14px;
      }

      .arrow {
        color: #909399;
        font-size: 12px;
      }
    }
  }
}

// 主内容区域
.dashboard-main {
  flex: 1;
  display: flex;
  gap: 15px;
  padding: 15px;
  overflow: hidden;
}

// 面板通用样式
.panel {
  background: linear-gradient(135deg, rgba(32, 45, 65, 0.8) 0%, rgba(20, 30, 48, 0.9) 100%);
  border: 1px solid rgba(64, 158, 255, 0.2);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);

  .panel-header {
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 15px;
    background: rgba(64, 158, 255, 0.1);
    border-bottom: 1px solid rgba(64, 158, 255, 0.2);
    flex-shrink: 0;

    .panel-title {
      font-size: 14px;
      font-weight: bold;
      color: #409eff;
      display: flex;
      align-items: center;
      gap: 8px;

      &::before {
        content: '';
        width: 3px;
        height: 14px;
        background: linear-gradient(180deg, #409eff, #67c23a);
        border-radius: 2px;
      }
    }

    .panel-actions {
      display: flex;
      align-items: center;
      gap: 10px;
    }
  }

  .panel-content {
    flex: 1;
    overflow: hidden;
    padding: 10px;
  }
}

// 左侧面板
.left-panel {
  width: 320px;
  display: flex;
  flex-direction: column;
  gap: 15px;
  flex-shrink: 0;

  .station-panel {
    flex: 1;
  }

  .alarm-panel {
    flex: 1;
  }
}

// 中间地图区域
.center-panel {
  flex: 1;
  min-width: 0;

  .map-panel {
    height: 100%;
  }
}

// 右侧面板
.right-panel {
  width: 380px;
  display: flex;
  flex-direction: column;
  gap: 15px;
  flex-shrink: 0;

  .stats-panel {
    flex-shrink: 0;
    height: auto;
  }

  .chart-panel {
    flex: 1;
    min-height: 0;
  }
}

// 响应式布局
@media (max-width: 1600px) {
  .left-panel {
    width: 280px;
  }

  .right-panel {
    width: 320px;
  }
}

@media (max-width: 1200px) {
  .dashboard-main {
    flex-wrap: wrap;
  }

  .left-panel,
  .right-panel {
    width: calc(50% - 7.5px);
  }

  .center-panel {
    width: 100%;
    order: -1;
    height: 400px;
  }
}

// 深色主题覆盖
:deep(.el-dropdown-menu) {
  background-color: #1a1f2e;
  border-color: rgba(64, 158, 255, 0.3);

  .el-dropdown-menu__item {
    color: #e5eaf3;

    &:hover {
      background-color: rgba(64, 158, 255, 0.2);
      color: #409eff;
    }
  }
}

:deep(.el-select-dropdown) {
  background-color: #1a1f2e;
  border-color: rgba(64, 158, 255, 0.3);

  .el-select-dropdown__item {
    color: #e5eaf3;

    &.selected,
    &:hover {
      background-color: rgba(64, 158, 255, 0.2);
      color: #409eff;
    }
  }
}

:deep(.el-button-group) {
  .el-button {
    background-color: rgba(32, 45, 65, 0.8);
    border-color: rgba(64, 158, 255, 0.3);
    color: #e5eaf3;

    &:hover {
      background-color: rgba(64, 158, 255, 0.3);
    }

    &.el-button--primary {
      background-color: #409eff;
      border-color: #409eff;
    }
  }
}
</style>
