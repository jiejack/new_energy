<template>
  <div class="skeleton-loader">
    <div v-if="type === 'card'" class="skeleton-card">
      <div class="skeleton-header">
        <el-skeleton :rows="1" animated />
      </div>
      <div class="skeleton-content">
        <el-skeleton :rows="3" animated />
      </div>
    </div>

    <div v-else-if="type === 'table'" class="skeleton-table">
      <div class="skeleton-table-header">
        <el-skeleton :rows="1" animated />
      </div>
      <div class="skeleton-table-body">
        <el-skeleton v-for="i in 5" :key="i" :rows="1" animated />
      </div>
    </div>

    <div v-else-if="type === 'chart'" class="skeleton-chart">
      <div class="skeleton-chart-header">
        <el-skeleton :rows="1" animated />
      </div>
      <div class="skeleton-chart-body">
        <div class="skeleton-chart-placeholder"></div>
      </div>
    </div>

    <div v-else-if="type === 'list'" class="skeleton-list">
      <div v-for="i in count" :key="i" class="skeleton-list-item">
        <el-skeleton :rows="2" animated />
      </div>
    </div>

    <div v-else class="skeleton-default">
      <el-skeleton :rows="rows" :animated="animated" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ElSkeleton } from 'element-plus'

interface Props {
  type?: 'card' | 'table' | 'chart' | 'list' | 'default'
  rows?: number
  animated?: boolean
  count?: number
}

withDefaults(defineProps<Props>(), {
  type: 'default',
  rows: 3,
  animated: true,
  count: 5,
})
</script>

<style scoped lang="scss">
.skeleton-loader {
  width: 100%;
}

.skeleton-card {
  background: rgba(26, 31, 46, 0.85);
  border: 1px solid rgba(0, 212, 170, 0.2);
  border-radius: 8px;
  padding: 16px;

  .skeleton-header {
    margin-bottom: 12px;
  }
}

.skeleton-table {
  background: rgba(26, 31, 46, 0.85);
  border: 1px solid rgba(0, 212, 170, 0.2);
  border-radius: 8px;
  padding: 16px;

  .skeleton-table-header {
    margin-bottom: 12px;
  }

  .skeleton-table-body > div {
    margin-bottom: 8px;

    &:last-child {
      margin-bottom: 0;
    }
  }
}

.skeleton-chart {
  background: rgba(26, 31, 46, 0.85);
  border: 1px solid rgba(0, 212, 170, 0.2);
  border-radius: 8px;
  padding: 16px;

  .skeleton-chart-header {
    margin-bottom: 12px;
  }

  .skeleton-chart-placeholder {
    height: 200px;
    background: linear-gradient(
      90deg,
      rgba(0, 212, 170, 0.1) 25%,
      rgba(0, 212, 170, 0.2) 37%,
      rgba(0, 212, 170, 0.1) 63%
    );
    background-size: 400% 100%;
    animation: skeleton-loading 1.4s ease infinite;
    border-radius: 4px;
  }
}

.skeleton-list {
  .skeleton-list-item {
    background: rgba(26, 31, 46, 0.85);
    border: 1px solid rgba(0, 212, 170, 0.2);
    border-radius: 8px;
    padding: 12px 16px;
    margin-bottom: 12px;

    &:last-child {
      margin-bottom: 0;
    }
  }
}

.skeleton-default {
  background: rgba(26, 31, 46, 0.85);
  border: 1px solid rgba(0, 212, 170, 0.2);
  border-radius: 8px;
  padding: 16px;
}

@keyframes skeleton-loading {
  0% {
    background-position: 100% 50%;
  }
  100% {
    background-position: 0 50%;
  }
}
</style>
