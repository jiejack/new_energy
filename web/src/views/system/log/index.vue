<template>
  <div class="log-container">
    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="queryParams" inline>
        <el-form-item label="用户名">
          <el-input
            v-model="queryParams.username"
            placeholder="请输入用户名"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="操作类型">
          <el-input
            v-model="queryParams.operation"
            placeholder="请输入操作类型"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryParams.status" placeholder="请选择状态" clearable>
            <el-option label="成功" :value="1" />
            <el-option label="失败" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="操作时间">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="-"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            @change="handleDateChange"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleQuery">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 表格 -->
    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>操作日志</span>
          <div class="header-actions">
            <el-button type="primary" @click="handleExport">
              <el-icon><Download /></el-icon>
              导出
            </el-button>
            <el-button
              type="danger"
              :disabled="selectedIds.length === 0"
              @click="handleBatchDelete"
            >
              <el-icon><Delete /></el-icon>
              批量删除
            </el-button>
            <el-button type="danger" @click="handleClear">
              <el-icon><DeleteFilled /></el-icon>
              清空日志
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="logList"
        border
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column prop="username" label="操作人" width="120" />
        <el-table-column prop="operation" label="操作类型" min-width="120" />
        <el-table-column prop="method" label="请求方法" width="200" show-overflow-tooltip />
        <el-table-column prop="ip" label="IP地址" width="140" />
        <el-table-column prop="location" label="操作地点" min-width="150" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="duration" label="耗时(ms)" width="100" align="center" />
        <el-table-column prop="createdAt" label="操作时间" width="180" />
        <el-table-column label="操作" width="100" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleDetail(row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="queryParams.page"
        v-model:page-size="queryParams.pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleQuery"
        @current-change="handleQuery"
      />
    </el-card>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="日志详情" width="700px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="操作人">
          {{ currentLog?.username }}
        </el-descriptions-item>
        <el-descriptions-item label="操作类型">
          {{ currentLog?.operation }}
        </el-descriptions-item>
        <el-descriptions-item label="请求方法" :span="2">
          {{ currentLog?.method }}
        </el-descriptions-item>
        <el-descriptions-item label="IP地址">
          {{ currentLog?.ip }}
        </el-descriptions-item>
        <el-descriptions-item label="操作地点">
          {{ currentLog?.location }}
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentLog?.status === 1 ? 'success' : 'danger'" size="small">
            {{ currentLog?.status === 1 ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="耗时">
          {{ currentLog?.duration }}ms
        </el-descriptions-item>
        <el-descriptions-item label="操作时间" :span="2">
          {{ currentLog?.createdAt }}
        </el-descriptions-item>
        <el-descriptions-item label="用户代理" :span="2">
          {{ currentLog?.userAgent }}
        </el-descriptions-item>
        <el-descriptions-item label="请求参数" :span="2">
          <el-input
            v-model="currentLog?.params"
            type="textarea"
            :rows="5"
            readonly
          />
        </el-descriptions-item>
        <el-descriptions-item v-if="currentLog?.errorMsg" label="错误信息" :span="2">
          <el-input
            v-model="currentLog.errorMsg"
            type="textarea"
            :rows="3"
            readonly
          />
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getLogList, batchDeleteLogs, clearLogs, exportLogs } from '@/api/log'
import type { OperationLog, PageQuery } from '@/types'

const loading = ref(false)
const logList = ref<OperationLog[]>([])
const total = ref(0)
const selectedIds = ref<number[]>([])
const detailVisible = ref(false)
const currentLog = ref<OperationLog | null>(null)
const dateRange = ref<string[]>([])

const queryParams = reactive<PageQuery & {
  username?: string
  operation?: string
  status?: number
  startTime?: string
  endTime?: string
}>({
  page: 1,
  pageSize: 10,
  username: '',
  operation: '',
  status: undefined,
  startTime: undefined,
  endTime: undefined,
})

// 获取日志列表
async function getList() {
  loading.value = true
  try {
    const result = await getLogList(queryParams)
    logList.value = result.list
    total.value = result.total
  } catch (error) {
    console.error('获取日志列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 日期变化
function handleDateChange(val: string[] | null) {
  if (val && val.length === 2) {
    queryParams.startTime = val[0] + ' 00:00:00'
    queryParams.endTime = val[1] + ' 23:59:59'
  } else {
    queryParams.startTime = undefined
    queryParams.endTime = undefined
  }
}

// 搜索
function handleQuery() {
  queryParams.page = 1
  getList()
}

// 重置
function handleReset() {
  queryParams.username = ''
  queryParams.operation = ''
  queryParams.status = undefined
  queryParams.startTime = undefined
  queryParams.endTime = undefined
  dateRange.value = []
  handleQuery()
}

// 选择变化
function handleSelectionChange(selection: OperationLog[]) {
  selectedIds.value = selection.map((item) => item.id)
}

// 查看详情
function handleDetail(row: OperationLog) {
  currentLog.value = row
  detailVisible.value = true
}

// 导出
async function handleExport() {
  try {
    await exportLogs({
      username: queryParams.username,
      operation: queryParams.operation,
      status: queryParams.status,
      startTime: queryParams.startTime,
      endTime: queryParams.endTime,
    })
    ElMessage.success('导出成功')
  } catch (error) {
    console.error('导出失败:', error)
  }
}

// 批量删除
async function handleBatchDelete() {
  try {
    await ElMessageBox.confirm(`确认要删除选中的 ${selectedIds.value.length} 条日志吗？`, '提示', {
      type: 'warning',
    })
    await batchDeleteLogs(selectedIds.value)
    ElMessage.success('批量删除成功')
    getList()
  } catch (error) {
    console.error('批量删除失败:', error)
  }
}

// 清空日志
async function handleClear() {
  try {
    await ElMessageBox.confirm('确认要清空所有日志吗？此操作不可恢复！', '警告', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消',
    })
    await clearLogs()
    ElMessage.success('清空成功')
    getList()
  } catch (error) {
    console.error('清空失败:', error)
  }
}

onMounted(() => {
  getList()
})
</script>

<style scoped lang="scss">
.log-container {
  padding: 20px;

  .search-card {
    margin-bottom: 20px;
  }

  .table-card {
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .header-actions {
        display: flex;
        gap: 10px;
      }
    }
  }

  .el-pagination {
    margin-top: 20px;
    justify-content: flex-end;
  }
}
</style>
