<template>
  <div class="stat-cards">
    <div class="stat-grid">
      <template v-if="!loading">
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
      </template>

      <template v-else>
        <!-- 骨架屏加载状态 -->
        <div v-for="i in 5" :key="i" class="stat-card skeleton-card">
          <div class="card-icon">
            <el-skeleton-item variant="circle" style="width: 50px; height: 50px;" />
          </div>
          <div class="card-content">
            <div class="card-value">
              <el-skeleton-item variant="text" style="width: 80px; height: 28px;" />
              <span class="unit-skeleton">MW</span>
            </div>
            <div class="card-label">
              <el-skeleton-item variant="text" style="width: 60px; height: 16px;" />
            </div>
          </div>
        </div>
      </template>
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
/* ============================================
   统计卡片 - 新能源监控专用样式
   ============================================ */
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
    background: linear-gradient(135deg, rgba(26, 31, 46, 0.6) 0%, rgba(13, 17, 23, 0.7) 100%);
    border: 1px solid rgba(0, 212, 170, 0.2);
    border-radius: $border-radius-base;
    overflow: hidden;
    transition: all 0.3s ease;
    backdrop-filter: blur(5px);

    /* 顶部装饰线 */
    &::before {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      height: 2px;
      background: var(--card-gradient, $gradient-primary);
      opacity: 0.8;
    }

    &:hover {
      background: linear-gradient(135deg, rgba(0, 212, 170, 0.15) 0%, rgba(9, 132, 227, 0.1) 100%);
      border-color: rgba(0, 212, 170, 0.4);
      transform: translateY(-2px);
      box-shadow: $shadow-glow;
    }

    .card-icon {
      width: 50px;
      height: 50px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 10px;
      flex-shrink: 0;
      position: relative;
      overflow: hidden;

      /* 图标发光效果 */
      &::after {
        content: '';
        position: absolute;
        top: -50%;
        left: -50%;
        width: 200%;
        height: 200%;
        background: radial-gradient(circle, rgba(255, 255, 255, 0.1) 0%, transparent 70%);
        opacity: 0;
        transition: opacity 0.3s ease;
      }

      &:hover::after {
        opacity: 1;
      }
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
          color: $text-primary;
          font-family: $font-family-number;
          background: var(--card-gradient, $gradient-primary);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
          background-clip: text;
          letter-spacing: 0.5px;
        }

        .unit {
          font-size: 12px;
          color: $text-secondary;
          font-weight: 500;
        }
      }

      .card-label {
        font-size: 12px;
        color: $text-secondary;
        font-weight: 500;
        letter-spacing: 0.5px;
      }
    }

    .card-decoration {
      position: absolute;
      top: 0;
      right: 0;
      width: 60px;
      height: 60px;
      border-radius: 50%;
      opacity: 0.08;
      transform: translate(20px, -20px);
      filter: blur(10px);
    }

    .progress-bar {
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      height: 3px;
      background: rgba(255, 255, 255, 0.05);
      overflow: hidden;

      .progress-fill {
        height: 100%;
        background: var(--card-gradient, $gradient-primary);
        transition: width 0.5s ease;
        box-shadow: 0 0 10px rgba(0, 212, 170, 0.5);
      }
    }

    // 不同卡片的样式 - 新能源主题
    &.capacity {
      --card-gradient: linear-gradient(135deg, #00d4aa 0%, #00b894 100%);

      .card-icon {
        background: linear-gradient(135deg, rgba(0, 212, 170, 0.2) 0%, rgba(0, 184, 148, 0.1) 100%);
        color: #00d4aa;
        box-shadow: 0 0 20px rgba(0, 212, 170, 0.3);
      }

      .card-decoration {
        background: radial-gradient(circle, #00d4aa 0%, transparent 70%);
      }
    }

    &.power {
      --card-gradient: linear-gradient(135deg, #00d4aa 0%, #0984e3 100%);

      .card-icon {
        background: linear-gradient(135deg, rgba(0, 212, 170, 0.2) 0%, rgba(9, 132, 227, 0.1) 100%);
        color: #00d4aa;
        box-shadow: 0 0 20px rgba(0, 212, 170, 0.3);
        animation: pulse-glow 2s ease-in-out infinite;
      }

      .card-decoration {
        background: radial-gradient(circle, #00d4aa 0%, transparent 70%);
      }
    }

    &.energy {
      --card-gradient: linear-gradient(135deg, #fdcb6e 0%, #f39c12 100%);

      .card-icon {
        background: linear-gradient(135deg, rgba(253, 203, 110, 0.2) 0%, rgba(243, 156, 18, 0.1) 100%);
        color: #fdcb6e;
        box-shadow: 0 0 20px rgba(253, 203, 110, 0.3);
      }

      .card-decoration {
        background: radial-gradient(circle, #fdcb6e 0%, transparent 70%);
      }
    }

    &.alarm {
      --card-gradient: linear-gradient(135deg, #ff7675 0%, #d63031 100%);

      .card-icon {
        background: linear-gradient(135deg, rgba(255, 118, 117, 0.2) 0%, rgba(214, 48, 49, 0.1) 100%);
        color: #ff7675;
        box-shadow: 0 0 20px rgba(255, 118, 117, 0.3);
      }

      .card-decoration {
        background: radial-gradient(circle, #ff7675 0%, transparent 70%);
      }
    }

    &.online-rate {
      grid-column: span 2;
      --card-gradient: linear-gradient(90deg, #00d4aa 0%, #0984e3 100%);

      .card-icon {
        background: linear-gradient(135deg, rgba(0, 212, 170, 0.2) 0%, rgba(9, 132, 227, 0.1) 100%);
        color: #00d4aa;
        box-shadow: 0 0 20px rgba(0, 212, 170, 0.3);
      }

      .card-decoration {
        background: radial-gradient(circle, #00d4aa 0%, transparent 70%);
      }

      .progress-bar {
        .progress-fill {
          box-shadow: 0 0 15px rgba(0, 212, 170, 0.6);
        }
      }
    }
  }
}

/* 脉冲发光动画 */
@keyframes pulse-glow {
  0%, 100% {
    box-shadow: 0 0 20px rgba(0, 212, 170, 0.3);
  }
  50% {
    box-shadow: 0 0 30px rgba(0, 212, 170, 0.5);
  }
}

// 骨架屏卡片样式
    .skeleton-card {
      &::before {
        display: none;
      }

      &:hover {
        transform: none;
        box-shadow: $shadow-light;
        border-color: rgba(0, 212, 170, 0.2);
        background: linear-gradient(135deg, rgba(26, 31, 46, 0.6) 0%, rgba(13, 17, 23, 0.7) 100%);
      }

      .card-icon {
        background: transparent;
        box-shadow: none;
      }

      .card-content {
        .card-value {
          .unit-skeleton {
            font-size: 12px;
            color: transparent;
            width: 30px;
          }
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
