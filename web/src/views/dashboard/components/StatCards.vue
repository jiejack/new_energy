<template>
  <div class="stat-cards">
    <div class="stat-grid">
      <template v-if="!loading">
        <!-- 总装机容量 -->
        <div class="stat-card capacity">
          <div class="card-icon">
            <el-icon :size="32"><Coin /></el-icon>
          </div>
          <div class="card-content">
            <div class="card-value">
              <span class="value">{{ formatNumber(stats.totalCapacity) }}</span>
              <span class="unit">MW</span>
            </div>
            <div class="card-label">总装机容量</div>
          </div>
          <div class="card-decoration"></div>
          <div class="card-glow"></div>
        </div>

        <!-- 实时发电功率 -->
        <div class="stat-card power">
          <div class="card-icon">
            <el-icon :size="32"><Promotion /></el-icon>
          </div>
          <div class="card-content">
            <div class="card-value">
              <span class="value">{{ formatNumber(stats.currentPower) }}</span>
              <span class="unit">MW</span>
            </div>
            <div class="card-label">实时发电功率</div>
          </div>
          <div class="card-decoration"></div>
          <div class="card-glow"></div>
        </div>

        <!-- 今日发电量 -->
        <div class="stat-card energy">
          <div class="card-icon">
            <el-icon :size="32"><Sunny /></el-icon>
          </div>
          <div class="card-content">
            <div class="card-value">
              <span class="value">{{ formatNumber(stats.todayEnergy) }}</span>
              <span class="unit">MWh</span>
            </div>
            <div class="card-label">今日发电量</div>
          </div>
          <div class="card-decoration"></div>
          <div class="card-glow"></div>
        </div>

        <!-- 告警数量 -->
        <div class="stat-card alarm">
          <div class="card-icon">
            <el-icon :size="32"><Bell /></el-icon>
          </div>
          <div class="card-content">
            <div class="card-value">
              <span class="value">{{ stats.alarmCount }}</span>
              <span class="unit">个</span>
            </div>
            <div class="card-label">告警数量</div>
          </div>
          <div class="card-decoration"></div>
          <div class="card-glow"></div>
        </div>

        <!-- 设备在线率 -->
        <div class="stat-card online-rate">
          <div class="card-icon">
            <el-icon :size="32"><Connection /></el-icon>
          </div>
          <div class="card-content">
            <div class="card-value">
              <span class="value">{{ stats.onlineRate.toFixed(1) }}</span>
              <span class="unit">%</span>
            </div>
            <div class="card-label">设备在线率</div>
          </div>
          <div class="card-decoration"></div>
          <div class="card-glow"></div>
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: `${stats.onlineRate}%` }"></div>
            <div class="progress-glow"></div>
          </div>
        </div>
      </template>

      <template v-else>
        <!-- 骨架屏加载状态 -->
        <div v-for="i in 5" :key="i" class="stat-card skeleton-card">
          <div class="card-icon">
            <el-skeleton-item variant="circle" style="width: 60px; height: 60px;" />
          </div>
          <div class="card-content">
            <div class="card-value">
              <el-skeleton-item variant="text" style="width: 100px; height: 32px;" />
              <span class="unit-skeleton">MW</span>
            </div>
            <div class="card-label">
              <el-skeleton-item variant="text" style="width: 80px; height: 18px;" />
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
@import url('https://fonts.googleapis.com/css2?family=Orbitron:wght@400;500;600;700&family=Rajdhani:wght@300;400;500;600;700&display=swap');

.stat-cards {
  .stat-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 15px;
  }

  .stat-card {
    position: relative;
    display: flex;
    align-items: center;
    gap: 15px;
    padding: 20px;
    background: linear-gradient(135deg, rgba(26, 31, 46, 0.8) 0%, rgba(13, 17, 23, 0.9) 100%);
    border: 1px solid rgba(0, 212, 170, 0.2);
    border-radius: 16px;
    overflow: hidden;
    transition: all 0.3s ease;
    backdrop-filter: blur(10px);
    animation: fadeInUp 0.6s ease-out;
    animation-fill-mode: both;

    /* 顶部装饰线 */
    &::before {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      height: 3px;
      background: var(--card-gradient, linear-gradient(90deg, #00d4aa, #7c3aed));
      opacity: 0.8;
      box-shadow: 0 0 10px rgba(0, 212, 170, 0.6);
    }

    &:hover {
      background: linear-gradient(135deg, rgba(0, 212, 170, 0.15) 0%, rgba(124, 58, 237, 0.1) 100%);
      border-color: rgba(0, 212, 170, 0.4);
      transform: translateY(-5px);
      box-shadow: 0 12px 40px rgba(0, 212, 170, 0.2);
    }

    .card-icon {
      width: 60px;
      height: 60px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 12px;
      flex-shrink: 0;
      position: relative;
      overflow: hidden;
      font-weight: bold;

      /* 图标发光效果 */
      &::after {
        content: '';
        position: absolute;
        top: -50%;
        left: -50%;
        width: 200%;
        height: 200%;
        background: radial-gradient(circle, rgba(255, 255, 255, 0.2) 0%, transparent 70%);
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
        gap: 6px;
        margin-bottom: 6px;

        .value {
          font-size: 26px;
          font-weight: 700;
          color: #e5eaf3;
          font-family: 'Orbitron', sans-serif;
          background: var(--card-gradient, linear-gradient(90deg, #00d4aa, #7c3aed));
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
          background-clip: text;
          letter-spacing: 1px;
          text-shadow: 0 0 20px rgba(0, 212, 170, 0.4);
        }

        .unit {
          font-size: 14px;
          color: #94a3b8;
          font-weight: 500;
          font-family: 'Rajdhani', sans-serif;
        }
      }

      .card-label {
        font-size: 14px;
        color: #94a3b8;
        font-weight: 500;
        letter-spacing: 0.5px;
        font-family: 'Rajdhani', sans-serif;
      }
    }

    .card-decoration {
      position: absolute;
      top: 0;
      right: 0;
      width: 80px;
      height: 80px;
      border-radius: 50%;
      opacity: 0.1;
      transform: translate(30px, -30px);
      filter: blur(15px);
    }

    .card-glow {
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: linear-gradient(135deg, transparent 0%, var(--card-glow, rgba(0, 212, 170, 0.05)) 100%);
      pointer-events: none;
    }

    .progress-bar {
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      height: 4px;
      background: rgba(255, 255, 255, 0.05);
      overflow: hidden;
      border-radius: 0 0 16px 16px;

      .progress-fill {
        height: 100%;
        background: var(--card-gradient, linear-gradient(90deg, #00d4aa, #7c3aed));
        transition: width 0.8s ease;
        box-shadow: 0 0 15px rgba(0, 212, 170, 0.6);
        position: relative;
      }

      .progress-glow {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: linear-gradient(90deg, transparent, rgba(0, 212, 170, 0.3), transparent);
        animation: progressFlow 2s ease-in-out infinite;
      }
    }

    // 不同卡片的样式 - 新能源主题
    &.capacity {
      --card-gradient: linear-gradient(135deg, #00d4aa 0%, #00b894 100%);
      --card-glow: rgba(0, 212, 170, 0.1);
      animation-delay: 0.1s;

      .card-icon {
        background: linear-gradient(135deg, rgba(0, 212, 170, 0.2) 0%, rgba(0, 184, 148, 0.1) 100%);
        color: #00d4aa;
        box-shadow: 0 0 25px rgba(0, 212, 170, 0.3);
      }

      .card-decoration {
        background: radial-gradient(circle, #00d4aa 0%, transparent 70%);
      }
    }

    &.power {
      --card-gradient: linear-gradient(135deg, #00d4aa 0%, #7c3aed 100%);
      --card-glow: rgba(124, 58, 237, 0.1);
      animation-delay: 0.2s;

      .card-icon {
        background: linear-gradient(135deg, rgba(0, 212, 170, 0.2) 0%, rgba(124, 58, 237, 0.1) 100%);
        color: #00d4aa;
        box-shadow: 0 0 25px rgba(0, 212, 170, 0.3);
        animation: pulse-glow 2s ease-in-out infinite;
      }

      .card-decoration {
        background: radial-gradient(circle, #7c3aed 0%, transparent 70%);
      }
    }

    &.energy {
      --card-gradient: linear-gradient(135deg, #fdcb6e 0%, #f39c12 100%);
      --card-glow: rgba(253, 203, 110, 0.1);
      animation-delay: 0.3s;

      .card-icon {
        background: linear-gradient(135deg, rgba(253, 203, 110, 0.2) 0%, rgba(243, 156, 18, 0.1) 100%);
        color: #fdcb6e;
        box-shadow: 0 0 25px rgba(253, 203, 110, 0.3);
      }

      .card-decoration {
        background: radial-gradient(circle, #fdcb6e 0%, transparent 70%);
      }
    }

    &.alarm {
      --card-gradient: linear-gradient(135deg, #ff7675 0%, #d63031 100%);
      --card-glow: rgba(255, 118, 117, 0.1);
      animation-delay: 0.4s;

      .card-icon {
        background: linear-gradient(135deg, rgba(255, 118, 117, 0.2) 0%, rgba(214, 48, 49, 0.1) 100%);
        color: #ff7675;
        box-shadow: 0 0 25px rgba(255, 118, 117, 0.3);
      }

      .card-decoration {
        background: radial-gradient(circle, #ff7675 0%, transparent 70%);
      }
    }

    &.online-rate {
      grid-column: span 2;
      --card-gradient: linear-gradient(90deg, #00d4aa 0%, #7c3aed 100%);
      --card-glow: rgba(0, 212, 170, 0.1);
      animation-delay: 0.5s;

      .card-icon {
        background: linear-gradient(135deg, rgba(0, 212, 170, 0.2) 0%, rgba(124, 58, 237, 0.1) 100%);
        color: #00d4aa;
        box-shadow: 0 0 25px rgba(0, 212, 170, 0.3);
      }

      .card-decoration {
        background: radial-gradient(circle, #00d4aa 0%, transparent 70%);
      }

      .progress-bar {
        .progress-fill {
          box-shadow: 0 0 20px rgba(0, 212, 170, 0.6);
        }
      }
    }
  }

  // 骨架屏卡片样式
  .skeleton-card {
    &::before {
      display: none;
    }

    &:hover {
      transform: none;
      box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
      border-color: rgba(0, 212, 170, 0.2);
      background: linear-gradient(135deg, rgba(26, 31, 46, 0.8) 0%, rgba(13, 17, 23, 0.9) 100%);
    }

    .card-icon {
      background: transparent;
      box-shadow: none;
    }

    .card-content {
      .card-value {
        .unit-skeleton {
          font-size: 14px;
          color: transparent;
          width: 35px;
        }
      }
    }
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

@keyframes pulse-glow {
  0%, 100% {
    box-shadow: 0 0 25px rgba(0, 212, 170, 0.3);
  }
  50% {
    box-shadow: 0 0 40px rgba(0, 212, 170, 0.5);
  }
}

@keyframes progressFlow {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
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