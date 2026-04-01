<template>
  <div class="stat-cards">
    <div class="stat-grid">
      <!-- 总装机容量 -->
      <div class="stat-card capacity">
        <div class="card-icon">
          <el-icon :size="28"><Coin /></el-icon>
        </div>
        <div class="card-content">
          <div class="card-value">
            <span class="value">{{ formatNumber(stats.totalCapacity) }}</span>
            <span class="unit">MW</span>
          </div>
          <div class="card-label">总装机容量</div>
        </div>
        <div class="card-decoration"></div>
      </div>

      <!-- 实时发电功率 -->
      <div class="stat-card power">
        <div class="card-icon">
          <el-icon :size="28"><Promotion /></el-icon>
        </div>
        <div class="card-content">
          <div class="card-value">
            <span class="value">{{ formatNumber(stats.currentPower) }}</span>
            <span class="unit">MW</span>
          </div>
          <div class="card-label">实时发电功率</div>
        </div>
        <div class="card-decoration"></div>
      </div>

      <!-- 今日发电量 -->
      <div class="stat-card energy">
        <div class="card-icon">
          <el-icon :size="28"><Sunny /></el-icon>
        </div>
        <div class="card-content">
          <div class="card-value">
            <span class="value">{{ formatNumber(stats.todayEnergy) }}</span>
            <span class="unit">MWh</span>
          </div>
          <div class="card-label">今日发电量</div>
        </div>
        <div class="card-decoration"></div>
      </div>

      <!-- 告警数量 -->
      <div class="stat-card alarm">
        <div class="card-icon">
          <el-icon :size="28"><Bell /></el-icon>
        </div>
        <div class="card-content">
          <div class="card-value">
            <span class="value">{{ stats.alarmCount }}</span>
            <span class="unit">个</span>
          </div>
          <div class="card-label">告警数量</div>
        </div>
        <div class="card-decoration"></div>
      </div>

      <!-- 设备在线率 -->
      <div class="stat-card online-rate">
        <div class="card-icon">
          <el-icon :size="28"><Connection /></el-icon>
        </div>
        <div class="card-content">
          <div class="card-value">
            <span class="value">{{ stats.onlineRate.toFixed(1) }}</span>
            <span class="unit">%</span>
          </div>
          <div class="card-label">设备在线率</div>
        </div>
        <div class="card-decoration"></div>
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: `${stats.onlineRate}%` }"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Coin, Promotion, Sunny, Bell, Connection } from '@element-plus/icons-vue'

interface Stats {
  totalCapacity: number
  currentPower: number
  todayEnergy: number
  alarmCount: number
  onlineRate: number
}

interface Props {
  stats: Stats
  loading?: boolean
}

withDefaults(defineProps<Props>(), {
  loading: false
})

/**
 * 格式化数字
 */
function formatNumber(num: number) {
  if (num >= 10000) {
    return (num / 10000).toFixed(2) + '万'
  }
  return num.toFixed(2)
}
</script>

<style scoped lang="scss">
.stat-cards {
  .stat-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }

  .stat-card {
    position: relative;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 15px;
    background: rgba(32, 45, 65, 0.4);
    border: 1px solid rgba(64, 158, 255, 0.2);
    border-radius: 8px;
    overflow: hidden;
    transition: all 0.3s;

    &:hover {
      background: rgba(64, 158, 255, 0.15);
      border-color: rgba(64, 158, 255, 0.4);
      transform: translateY(-2px);
    }

    .card-icon {
      width: 50px;
      height: 50px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 10px;
      flex-shrink: 0;
    }

    .card-content {
      flex: 1;
      min-width: 0;

      .card-value {
        display: flex;
        align-items: baseline;
        gap: 4px;
        margin-bottom: 4px;

        .value {
          font-size: 22px;
          font-weight: bold;
          color: #e5eaf3;
          font-family: 'DIN', 'Courier New', monospace;
        }

        .unit {
          font-size: 12px;
          color: #909399;
        }
      }

      .card-label {
        font-size: 12px;
        color: #909399;
      }
    }

    .card-decoration {
      position: absolute;
      top: 0;
      right: 0;
      width: 60px;
      height: 60px;
      border-radius: 50%;
      opacity: 0.1;
      transform: translate(20px, -20px);
    }

    .progress-bar {
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      height: 3px;
      background: rgba(255, 255, 255, 0.1);

      .progress-fill {
        height: 100%;
        transition: width 0.5s ease;
      }
    }

    // 不同卡片的样式
    &.capacity {
      .card-icon {
        background: linear-gradient(135deg, #409eff33, #409eff11);
        color: #409eff;
      }

      .card-decoration {
        background: #409eff;
      }
    }

    &.power {
      .card-icon {
        background: linear-gradient(135deg, #67c23a33, #67c23a11);
        color: #67c23a;
      }

      .card-decoration {
        background: #67c23a;
      }
    }

    &.energy {
      .card-icon {
        background: linear-gradient(135deg, #e6a23c33, #e6a23c11);
        color: #e6a23c;
      }

      .card-decoration {
        background: #e6a23c;
      }
    }

    &.alarm {
      .card-icon {
        background: linear-gradient(135deg, #f56c6c33, #f56c6c11);
        color: #f56c6c;
      }

      .card-decoration {
        background: #f56c6c;
      }
    }

    &.online-rate {
      grid-column: span 2;

      .card-icon {
        background: linear-gradient(135deg, #00d4ff33, #00d4ff11);
        color: #00d4ff;
      }

      .card-decoration {
        background: #00d4ff;
      }

      .progress-bar {
        .progress-fill {
          background: linear-gradient(90deg, #409eff, #67c23a);
        }
      }
    }
  }
}

// 响应式
@media (max-width: 1400px) {
  .stat-cards {
    .stat-grid {
      grid-template-columns: 1fr;

      .stat-card {
        &.online-rate {
          grid-column: span 1;
        }
      }
    }
  }
}
</style>
