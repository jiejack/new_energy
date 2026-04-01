<template>
  <div class="station-list">
    <!-- 搜索框 -->
    <el-input
      v-model="searchKeyword"
      placeholder="搜索电站名称"
      clearable
      size="small"
      class="search-input"
    >
      <template #prefix>
        <el-icon><Search /></el-icon>
      </template>
    </el-input>

    <!-- 电站列表 -->
    <el-scrollbar class="list-scrollbar">
      <div v-if="loading" class="loading-container">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>加载中...</span>
      </div>

      <div v-else-if="filteredStations.length === 0" class="empty-container">
        <el-icon><DocumentRemove /></el-icon>
        <span>暂无电站数据</span>
      </div>

      <div v-else class="station-items">
        <div
          v-for="station in filteredStations"
          :key="station.id"
          class="station-item"
          :class="{ active: selectedStation?.id === station.id }"
          @click="handleSelect(station)"
        >
          <div class="station-icon">
            <el-icon :size="24">
              <component :is="getStationIcon(station.type)" />
            </el-icon>
          </div>
          <div class="station-info">
            <div class="station-name">{{ station.name }}</div>
            <div class="station-meta">
              <span class="station-type">{{ getStationTypeName(station.type) }}</span>
              <span class="station-capacity">{{ station.capacity }} MW</span>
            </div>
          </div>
          <div class="station-status">
            <el-tag
              :type="getStatusType(station.status)"
              size="small"
              effect="dark"
            >
              {{ getStatusName(station.status) }}
            </el-tag>
          </div>
        </div>
      </div>
    </el-scrollbar>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Search, Loading, DocumentRemove, Sunny, WindPower, Opportunity, Coin } from '@element-plus/icons-vue'
import type { Station, StationType, StationStatus } from '@/types'

interface Props {
  stations: Station[]
  loading?: boolean
  selectedStation?: Station | null
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  selectedStation: null
})

const emit = defineEmits<{
  select: [station: Station]
}>()

const searchKeyword = ref('')

// 过滤后的电站列表
const filteredStations = computed(() => {
  if (!searchKeyword.value) {
    return props.stations
  }
  const keyword = searchKeyword.value.toLowerCase()
  return props.stations.filter(station =>
    station.name.toLowerCase().includes(keyword) ||
    station.code.toLowerCase().includes(keyword)
  )
})

/**
 * 获取电站图标
 */
function getStationIcon(type: StationType) {
  const iconMap: Record<StationType, any> = {
    solar: Sunny,
    wind: WindPower,
    hydro: Opportunity,
    storage: Coin
  }
  return iconMap[type] || Sunny
}

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
 * 处理选择
 */
function handleSelect(station: Station) {
  emit('select', station)
}
</script>

<style scoped lang="scss">
.station-list {
  height: 100%;
  display: flex;
  flex-direction: column;

  .search-input {
    margin-bottom: 10px;
    flex-shrink: 0;

    :deep(.el-input__wrapper) {
      background-color: rgba(32, 45, 65, 0.6);
      border-color: rgba(64, 158, 255, 0.3);
      box-shadow: none;

      &:hover {
        border-color: rgba(64, 158, 255, 0.5);
      }

      &.is-focus {
        border-color: #409eff;
      }
    }

    :deep(.el-input__inner) {
      color: #e5eaf3;

      &::placeholder {
        color: #909399;
      }
    }

    :deep(.el-input__prefix) {
      color: #909399;
    }
  }

  .list-scrollbar {
    flex: 1;

    :deep(.el-scrollbar__bar) {
      &.is-vertical {
        width: 6px;
        right: 2px;
      }

      .el-scrollbar__thumb {
        background-color: rgba(64, 158, 255, 0.3);

        &:hover {
          background-color: rgba(64, 158, 255, 0.5);
        }
      }
    }
  }

  .loading-container,
  .empty-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 200px;
    color: #909399;
    gap: 10px;

    .el-icon {
      font-size: 32px;
    }
  }

  .station-items {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .station-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: rgba(32, 45, 65, 0.4);
    border: 1px solid rgba(64, 158, 255, 0.2);
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.3s;

    &:hover {
      background: rgba(64, 158, 255, 0.15);
      border-color: rgba(64, 158, 255, 0.4);
      transform: translateX(4px);
    }

    &.active {
      background: rgba(64, 158, 255, 0.25);
      border-color: #409eff;

      .station-icon {
        background: linear-gradient(135deg, #409eff, #67c23a);
      }
    }

    .station-icon {
      width: 40px;
      height: 40px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: rgba(64, 158, 255, 0.2);
      border-radius: 8px;
      color: #409eff;
      flex-shrink: 0;
    }

    .station-info {
      flex: 1;
      min-width: 0;

      .station-name {
        font-size: 14px;
        font-weight: 500;
        color: #e5eaf3;
        margin-bottom: 4px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .station-meta {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 12px;
        color: #909399;

        .station-type {
          padding: 2px 6px;
          background: rgba(103, 194, 58, 0.2);
          border-radius: 3px;
          color: #67c23a;
        }

        .station-capacity {
          color: #e6a23c;
        }
      }
    }

    .station-status {
      flex-shrink: 0;
    }
  }
}
</style>
