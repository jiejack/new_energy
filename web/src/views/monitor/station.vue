<template>
  <div class="station-monitor-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>电站监控</span>
          <div class="header-actions">
            <el-input
              v-model="searchQuery"
              placeholder="搜索电站名称"
              style="width: 200px"
              clearable
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-button type="primary" @click="refreshData">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <el-row :gutter="20">
        <el-col
          v-for="station in filteredStations"
          :key="station.id"
          :xs="24"
          :sm="12"
          :md="8"
          :lg="6"
        >
          <el-card class="station-card" :class="{ 'offline': !station.online }">
            <div class="station-header">
              <div class="station-status">
                <el-tag :type="station.online ? 'success' : 'info'" size="small">
                  {{ station.online ? '在线' : '离线' }}
                </el-tag>
              </div>
              <el-dropdown trigger="click" @command="(cmd: string) => handleCommand(cmd, station)">
                <el-icon class="more-icon"><MoreFilled /></el-icon>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="view">查看详情</el-dropdown-item>
                    <el-dropdown-item command="data">实时数据</el-dropdown-item>
                    <el-dropdown-item command="alarm">告警记录</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>

            <div class="station-info">
              <h3 class="station-name">{{ station.name }}</h3>
              <p class="station-address">{{ station.address }}</p>
            </div>

            <div class="station-stats">
              <div class="stat-item">
                <div class="stat-value">{{ station.capacity }} kW</div>
                <div class="stat-label">装机容量</div>
              </div>
              <div class="stat-item">
                <div class="stat-value" :class="{ 'text-success': station.online }">
                  {{ station.currentPower }} kW
                </div>
                <div class="stat-label">当前功率</div>
              </div>
              <div class="stat-item">
                <div class="stat-value">{{ station.todayEnergy }} kWh</div>
                <div class="stat-label">今日发电</div>
              </div>
            </div>

            <div class="station-devices">
              <div class="device-stat">
                <span class="device-label">设备:</span>
                <span class="device-value">
                  <el-tag size="small" type="success">{{ station.onlineDevices }} 在线</el-tag>
                  <el-tag size="small" type="info">{{ station.totalDevices - station.onlineDevices }} 离线</el-tag>
                </span>
              </div>
            </div>

            <div class="station-footer">
              <span class="update-time">更新: {{ station.updateTime }}</span>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-empty v-if="filteredStations.length === 0" description="暂无电站数据" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Refresh, Search, MoreFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const router = useRouter()
const searchQuery = ref('')

// 模拟电站数据
const stations = ref([
  {
    id: 1,
    name: '北京朝阳光伏电站',
    address: '北京市朝阳区XXX路XXX号',
    capacity: 5000,
    currentPower: 3200,
    todayEnergy: 15600,
    online: true,
    totalDevices: 25,
    onlineDevices: 24,
    updateTime: '2024-01-15 14:30:00'
  },
  {
    id: 2,
    name: '上海浦东风电场',
    address: '上海市浦东新区XXX路XXX号',
    capacity: 10000,
    currentPower: 6800,
    todayEnergy: 42300,
    online: true,
    totalDevices: 18,
    onlineDevices: 18,
    updateTime: '2024-01-15 14:30:00'
  },
  {
    id: 3,
    name: '广州番禺储能站',
    address: '广州市番禺区XXX路XXX号',
    capacity: 2000,
    currentPower: 0,
    todayEnergy: 5200,
    online: false,
    totalDevices: 12,
    onlineDevices: 0,
    updateTime: '2024-01-15 14:30:00'
  },
  {
    id: 4,
    name: '深圳南山光伏电站',
    address: '深圳市南山区XXX路XXX号',
    capacity: 8000,
    currentPower: 5600,
    todayEnergy: 28900,
    online: true,
    totalDevices: 32,
    onlineDevices: 31,
    updateTime: '2024-01-15 14:30:00'
  }
])

// 过滤后的电站列表
const filteredStations = computed(() => {
  if (!searchQuery.value) return stations.value
  return stations.value.filter(s =>
    s.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

// 刷新数据
function refreshData() {
  ElMessage.success('数据已刷新')
}

// 处理下拉菜单命令
function handleCommand(command: string, station: any) {
  switch (command) {
    case 'view':
      router.push(`/system/station?id=${station.id}`)
      break
    case 'data':
      router.push(`/data/history?stationId=${station.id}`)
      break
    case 'alarm':
      router.push(`/alarm/list?stationId=${station.id}`)
      break
  }
}
</script>

<style scoped lang="scss">
.station-monitor-page {
  padding: 20px;
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

.station-card {
  margin-bottom: 20px;
  transition: all 0.3s;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  &.offline {
    opacity: 0.7;
    background-color: #f5f7fa;
  }
}

.station-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.more-icon {
  cursor: pointer;
  color: #909399;

  &:hover {
    color: #409eff;
  }
}

.station-info {
  margin-bottom: 15px;

  .station-name {
    font-size: 16px;
    font-weight: bold;
    margin: 0 0 8px 0;
    color: #303133;
  }

  .station-address {
    font-size: 12px;
    color: #909399;
    margin: 0;
  }
}

.station-stats {
  display: flex;
  justify-content: space-between;
  padding: 15px 0;
  border-top: 1px solid #ebeef5;
  border-bottom: 1px solid #ebeef5;
  margin-bottom: 15px;

  .stat-item {
    text-align: center;

    .stat-value {
      font-size: 18px;
      font-weight: bold;
      color: #303133;
      margin-bottom: 4px;

      &.text-success {
        color: #67c23a;
      }
    }

    .stat-label {
      font-size: 12px;
      color: #909399;
    }
  }
}

.station-devices {
  margin-bottom: 15px;

  .device-stat {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .device-label {
      font-size: 12px;
      color: #606266;
    }

    .device-value {
      display: flex;
      gap: 5px;
    }
  }
}

.station-footer {
  text-align: right;

  .update-time {
    font-size: 11px;
    color: #c0c4cc;
  }
}
</style>
