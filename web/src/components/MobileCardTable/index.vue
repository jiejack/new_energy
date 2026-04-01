<template>
  <el-table v-if="!isMobile" :data="data" v-bind="$attrs">
    <slot />
  </el-table>

  <div v-else class="mobile-card-list">
    <div
      v-for="item in data"
      :key="item.id || item"
      class="mobile-card"
      @click="handleClick(item)"
    >
      <div class="card-header" v-if="$slots.header || title">
        <span class="card-title">{{ getTitle(item) }}</span>
        <el-tag
          v-if="statusField"
          :type="getStatusType(item[statusField])"
          size="small"
        >
          {{ item[statusField] }}
        </el-tag>
      </div>

      <div class="card-body">
        <div
          v-for="field in cardFields"
          :key="field.prop"
          class="card-field"
        >
          <span class="field-label">{{ field.label }}:</span>
          <span class="field-value">{{ item[field.prop] || '-' }}</span>
        </div>
      </div>

      <div class="card-footer" v-if="actions && actions.length > 0">
        <el-button
          v-for="action in actions"
          :key="action.label"
          :type="action.type || 'default'"
          size="small"
          @click.stop="action.handler(item)"
        >
          {{ action.label }}
        </el-button>
      </div>
    </div>

    <el-empty v-if="data.length === 0" description="暂无数据" />
  </div>
</template>

<script setup lang="ts">
import { useDevice } from '@/utils/device'

interface CardField {
  prop: string
  label: string
}

interface Action {
  label: string
  type?: 'primary' | 'success' | 'warning' | 'danger' | 'info'
  handler: (row: any) => void
}

interface Props {
  data: any[]
  cardFields: CardField[]
  title?: string
  statusField?: string
  actions?: Action[]
}

const props = withDefaults(defineProps<Props>(), {
  data: () => [],
  cardFields: () => [],
  title: '',
  statusField: '',
  actions: () => []
})

const emit = defineEmits<{
  (e: 'click', row: any): void
}>()

const { isMobile } = useDevice()

const getTitle = (item: any): string => {
  if (props.title) {
    return props.title
  }
  return item.name || item.title || item.id || '未命名'
}

const getStatusType = (status: string): 'success' | 'warning' | 'danger' | 'info' => {
  const typeMap: Record<string, 'success' | 'warning' | 'danger' | 'info'> = {
    online: 'success',
    offline: 'info',
    fault: 'danger',
    maintenance: 'warning',
    active: 'danger',
    acknowledged: 'warning',
    resolved: 'success'
  }
  return typeMap[status] || 'info'
}

const handleClick = (item: any) => {
  emit('click', item)
}
</script>

<style scoped lang="scss">
@import '@/styles/responsive.scss';

.mobile-card-list {
  .mobile-card {
    background: #fff;
    border-radius: 8px;
    padding: 15px;
    margin-bottom: 10px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    transition: all 0.3s;

    &:active {
      transform: scale(0.98);
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 10px;
      padding-bottom: 10px;
      border-bottom: 1px solid #f0f0f0;

      .card-title {
        font-size: 16px;
        font-weight: 500;
        color: #303133;
      }
    }

    .card-body {
      .card-field {
        display: flex;
        justify-content: space-between;
        padding: 8px 0;
        font-size: 14px;

        .field-label {
          color: #909399;
          min-width: 80px;
        }

        .field-value {
          color: #606266;
          text-align: right;
          flex: 1;
        }
      }
    }

    .card-footer {
      display: flex;
      gap: 10px;
      justify-content: flex-end;
      margin-top: 10px;
      padding-top: 10px;
      border-top: 1px solid #f0f0f0;
    }
  }
}
</style>
