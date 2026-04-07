<template>
  <div class="alarm-list-page">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :span="4">
        <div class="stat-card total">
          <div class="stat-value">{{ statistics.total }}</div>
          <div class="stat-label">总告警</div>
        </div>
      </el-col>
      <el-col :span="4">
        <div class="stat-card critical">
          <div class="stat-value">{{ statistics.critical }}</div>
          <div class="stat-label">严重</div>
        </div>
      </el-col>
      <el-col :span="4">
        <div class="stat-card major">
          <div class="stat-value">{{ statistics.major }}</div>
          <div class="stat-label">主要</div>
        </div>
      </el-col>
      <el-col :span="4">
        <div class="stat-card minor">
          <div class="stat-value">{{ statistics.minor }}</div>
          <div class="stat-label">次要</div>
        </div>
      </el-col>
      <el-col :span="4">
        <div class="stat-card warning">
          <div class="stat-value">{{ statistics.warning }}</div>
          <div class="stat-label">警告</div>
        </div>
      </el-col>
      <el-col :span="4">
        <div class="stat-card active">
          <div class="stat-value">{{ activeCount }}</div>
          <div class="stat-label">活动告警</div>
        </div>
      </el-col>
    </el-row>

    <!-- 查询条件 -->
    <el-card class="query-card" shadow="never">
      <el-form :model="queryForm" inline>
        <el-form-item label="关键词">
          <el-input
            v-model="queryForm.keyword"
            placeholder="告警标题/内容"
            clearable
            style="width: 200px"
            @keyup.enter="handleQuery"
          />
        </el-form-item>

        <el-form-item label="告警级别">
          <el-select v-model="queryForm.level" placeholder="全部" clearable style="width: 120px">
            <el-option label="严重" value="critical" />
            <el-option label="主要" value="major" />
            <el-option label="次要" value="minor" />
            <el-option label="警告" value="warning" />
          </el-select>
        </el-form-item>

        <el-form-item label="告警状态">
          <el-select v-model="queryForm.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="活动" value="active" />
            <el-option label="已确认" value="acknowledged" />
            <el-option label="已解决" value="resolved" />
          </el-select>
        </el-form-item>

        <el-form-item label="时间范围">
          <el-date-picker
            v-model="queryForm.timeRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 360px"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleQuery">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 告警列表 -->
    <el-card class="list-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">告警列表</span>
          <div class="actions">
            <el-button
              type="success"
              :disabled="selectedIds.length === 0"
              @click="handleBatchAcknowledge"
            >
              <el-icon><Check /></el-icon>
              批量确认
            </el-button>
            <el-button
              type="primary"
              :disabled="selectedIds.length === 0"
              @click="handleBatchResolve"
            >
              <el-icon><CircleCheck /></el-icon>
              批量清除
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        ref="tableRef"
        v-loading="loading"
        :data="alarmList"
        border
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />

        <el-table-column prop="level" label="级别" width="80">
          <template #default="{ row }">
            <el-tag :type="getLevelTagType(row.level)" effect="dark" size="small">
              {{ getLevelText(row.level) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="title" label="告警标题" min-width="200" show-overflow-tooltip />

        <el-table-column prop="sourceName" label="告警源" width="150" show-overflow-tooltip />

        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="occurredAt" label="发生时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.occurredAt) }}
          </template>
        </el-table-column>

        <el-table-column prop="acknowledgedBy" label="确认人" width="100">
          <template #default="{ row }">
            {{ row.acknowledgedBy || '-' }}
          </template>
        </el-table-column>

        <el-table-column prop="resolvedBy" label="清除人" width="100">
          <template #default="{ row }">
            {{ row.resolvedBy || '-' }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewDetail(row)">
              详情
            </el-button>
            <el-button
              v-if="row.status === 'active'"
              type="success"
              link
              size="small"
              @click="handleAcknowledge(row)"
            >
              确认
            </el-button>
            <el-button
              v-if="row.status !== 'resolved'"
              type="warning"
              link
              size="small"
              @click="handleResolve(row)"
            >
              清除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          :background="true"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleQuery"
          @current-change="handleQuery"
        />
      </div>
    </el-card>

    <!-- 告警详情对话框 -->
    <el-dialog v-model="detailVisible" title="告警详情" width="600px">
      <el-descriptions v-if="currentAlarm" :column="2" border>
        <el-descriptions-item label="告警ID">{{ currentAlarm.id }}</el-descriptions-item>
        <el-descriptions-item label="告警级别">
          <el-tag :type="getLevelTagType(currentAlarm.level)" effect="dark">
            {{ getLevelText(currentAlarm.level) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="告警标题" :span="2">
          {{ currentAlarm.title }}
        </el-descriptions-item>
        <el-descriptions-item label="告警内容" :span="2">
          {{ currentAlarm.content }}
        </el-descriptions-item>
        <el-descriptions-item label="告警源">{{ currentAlarm.sourceName }}</el-descriptions-item>
        <el-descriptions-item label="告警状态">
          <el-tag :type="getStatusTagType(currentAlarm.status)">
            {{ getStatusText(currentAlarm.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="发生时间">
          {{ formatTime(currentAlarm.occurredAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="确认时间">
          {{ currentAlarm.acknowledgedAt ? formatTime(currentAlarm.acknowledgedAt) : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="确认人">{{ currentAlarm.acknowledgedBy || '-' }}</el-descriptions-item>
        <el-descriptions-item label="清除时间">
          {{ currentAlarm.resolvedAt ? formatTime(currentAlarm.resolvedAt) : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="清除人">{{ currentAlarm.resolvedBy || '-' }}</el-descriptions-item>
      </el-descriptions>

      <template #footer>
        <el-button
          v-if="currentAlarm?.status === 'active'"
          type="success"
          @click="handleAcknowledge(currentAlarm!)"
        >
          确认告警
        </el-button>
        <el-button
          v-if="currentAlarm?.status !== 'resolved'"
          type="primary"
          @click="handleResolve(currentAlarm!)"
        >
          清除告警
        </el-button>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Check, CircleCheck } from '@element-plus/icons-vue'
import {
  getAlarmList,
  getAlarmStatistics,
  acknowledgeAlarm,
  batchAcknowledgeAlarms,
  resolveAlarm,
  batchResolveAlarms,
} from '@/api/alarm'
import type { Alarm, AlarmLevel, AlarmStatus } from '@/types'
import dayjs from 'dayjs'
import { alarmStatusMapper, alarmLevelMapper } from '@/utils/enums'

const loading = ref(false)
const detailVisible = ref(false)
const currentAlarm = ref<Alarm | null>(null)
const alarmList = ref<Alarm[]>([])
const selectedIds = ref<number[]>([])

const statistics = reactive({
  total: 0,
  critical: 0,
  major: 0,
  minor: 0,
  warning: 0,
})

const queryForm = reactive({
  keyword: '',
  level: '' as AlarmLevel | '',
  status: '' as AlarmStatus | '',
  timeRange: [] as string[],
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

// 活动告警数量
const activeCount = computed(() => {
  return alarmList.value.filter((a) => a.status === 'active').length
})

// 获取级别标签类型
const getLevelTagType = (level: AlarmLevel): 'danger' | 'warning' | 'info' | '' => {
  const typeMap: Record<AlarmLevel, 'danger' | 'warning' | 'info' | ''> = {
    critical: 'danger',
    major: 'warning',
    minor: 'info',
    warning: '',
  }
  return typeMap[level]
}

const getLevelText = (level: AlarmLevel) => alarmLevelMapper.getLabel(level)
const getStatusTagType = (status: AlarmStatus) => alarmStatusMapper.getTagType(status)
const getStatusText = (status: AlarmStatus) => alarmStatusMapper.getLabel(status)

// 格式化时间
const formatTime = (time: string): string => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

// 获取告警统计
const fetchStatistics = async () => {
  try {
    const result = await getAlarmStatistics()
    Object.assign(statistics, result)
  } catch (error) {
    console.error('获取告警统计失败:', error)
  }
}

// 获取告警列表
const fetchAlarmList = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.page,
      pageSize: pagination.pageSize,
    }
    if (queryForm.keyword) params.keyword = queryForm.keyword
    if (queryForm.level) params.level = queryForm.level
    if (queryForm.status) params.status = queryForm.status
    if (queryForm.timeRange && queryForm.timeRange.length === 2) {
      params.startTime = queryForm.timeRange[0]
      params.endTime = queryForm.timeRange[1]
    }

    const result = await getAlarmList(params)
    alarmList.value = result.list
    pagination.total = result.total
  } catch (error: any) {
    ElMessage.error(error.message || '获取告警列表失败')
  } finally {
    loading.value = false
  }
}

// 查询
const handleQuery = () => {
  pagination.page = 1
  fetchAlarmList()
}

// 重置
const handleReset = () => {
  queryForm.keyword = ''
  queryForm.level = ''
  queryForm.status = ''
  queryForm.timeRange = []
  handleQuery()
}

// 选择变化
const handleSelectionChange = (selection: Alarm[]) => {
  selectedIds.value = selection.map((item) => item.id)
}

// 查看详情
const handleViewDetail = (alarm: Alarm) => {
  currentAlarm.value = alarm
  detailVisible.value = true
}

// 确认告警
const handleAcknowledge = async (alarm: Alarm) => {
  try {
    await ElMessageBox.confirm('确认该告警？', '提示', {
      type: 'warning',
    })
    await acknowledgeAlarm(alarm.id)
    ElMessage.success('确认成功')
    detailVisible.value = false
    fetchAlarmList()
    fetchStatistics()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '确认失败')
    }
  }
}

// 批量确认
const handleBatchAcknowledge = async () => {
  try {
    await ElMessageBox.confirm(`确认选中的 ${selectedIds.value.length} 条告警？`, '提示', {
      type: 'warning',
    })
    await batchAcknowledgeAlarms(selectedIds.value)
    ElMessage.success('批量确认成功')
    fetchAlarmList()
    fetchStatistics()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '批量确认失败')
    }
  }
}

// 清除告警
const handleResolve = async (alarm: Alarm) => {
  try {
    await ElMessageBox.confirm('清除该告警？', '提示', {
      type: 'warning',
    })
    await resolveAlarm(alarm.id)
    ElMessage.success('清除成功')
    detailVisible.value = false
    fetchAlarmList()
    fetchStatistics()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '清除失败')
    }
  }
}

// 批量清除
const handleBatchResolve = async () => {
  try {
    await ElMessageBox.confirm(`清除选中的 ${selectedIds.value.length} 条告警？`, '提示', {
      type: 'warning',
    })
    await batchResolveAlarms(selectedIds.value)
    ElMessage.success('批量清除成功')
    fetchAlarmList()
    fetchStatistics()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '批量清除失败')
    }
  }
}

// 初始化
onMounted(() => {
  fetchStatistics()
  fetchAlarmList()
})
</script>

<style scoped lang="scss">
.alarm-list-page {
  .stat-row {
    margin-bottom: 16px;
  }

  .stat-card {
    padding: 20px;
    border-radius: 8px;
    color: #fff;
    text-align: center;

    .stat-value {
      font-size: 32px;
      font-weight: bold;
      margin-bottom: 8px;
    }

    .stat-label {
      font-size: 14px;
      opacity: 0.9;
    }

    &.total {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    }

    &.critical {
      background: linear-gradient(135deg, #f56c6c 0%, #e64545 100%);
    }

    &.major {
      background: linear-gradient(135deg, #e6a23c 0%, #d4940c 100%);
    }

    &.minor {
      background: linear-gradient(135deg, #409eff 0%, #2d8cf0 100%);
    }

    &.warning {
      background: linear-gradient(135deg, #909399 0%, #606266 100%);
    }

    &.active {
      background: linear-gradient(135deg, #67c23a 0%, #4caf50 100%);
    }
  }

  .query-card {
    margin-bottom: 16px;
  }

  .list-card {
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .title {
        font-size: 16px;
        font-weight: 500;
      }

      .actions {
        display: flex;
        gap: 8px;
      }
    }
  }

  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
}
</style>
