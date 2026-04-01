<template>
  <div class="device-management">
    <!-- 搜索筛选 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="queryParams" size="default">
        <el-form-item label="设备名称">
          <el-input
            v-model="queryParams.keyword"
            placeholder="请输入设备名称或编码"
            clearable
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="所属电站">
          <el-select
            v-model="queryParams.stationId"
            placeholder="请选择电站"
            clearable
            filterable
            style="width: 200px"
          >
            <el-option
              v-for="station in stationList"
              :key="station.id"
              :label="station.name"
              :value="station.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="设备类型">
          <el-select
            v-model="queryParams.type"
            placeholder="请选择类型"
            clearable
            style="width: 150px"
          >
            <el-option label="逆变器" value="inverter" />
            <el-option label="电表" value="meter" />
            <el-option label="传感器" value="sensor" />
            <el-option label="控制器" value="controller" />
            <el-option label="汇流箱" value="combiner" />
          </el-select>
        </el-form-item>
        <el-form-item label="设备状态">
          <el-select
            v-model="queryParams.status"
            placeholder="请选择状态"
            clearable
            style="width: 150px"
          >
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="维护" value="maintenance" />
            <el-option label="故障" value="fault" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
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

    <!-- 操作栏 -->
    <el-card class="toolbar-card">
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        新增设备
      </el-button>
      <el-button
        type="danger"
        :disabled="selectedIds.length === 0"
        @click="handleBatchDelete"
      >
        <el-icon><Delete /></el-icon>
        批量删除
      </el-button>
      <el-button @click="handleExport">
        <el-icon><Download /></el-icon>
        导出
      </el-button>
    </el-card>

    <!-- 表格 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="tableData"
        border
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column type="index" label="序号" width="60" align="center" />
        <el-table-column prop="name" label="设备名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="code" label="设备编码" width="120" show-overflow-tooltip />
        <el-table-column prop="stationName" label="所属电站" width="120" show-overflow-tooltip />
        <el-table-column prop="type" label="设备类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getTypeTagType(row.type)">
              {{ getTypeName(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="model" label="设备型号" width="120" show-overflow-tooltip />
        <el-table-column prop="manufacturer" label="制造商" width="120" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)">
              {{ getStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="installDate" label="安装日期" width="110" align="center" />
        <el-table-column prop="lastMaintenanceDate" label="最后维护日期" width="120" align="center" />
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleView(row)">
              查看
            </el-button>
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button
              :type="row.status === 'online' ? 'warning' : 'success'"
              link
              size="small"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 'online' ? '停用' : '启用' }}
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchData"
          @current-change="fetchData"
        />
      </div>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <FormDialog
      v-model="dialogVisible"
      :mode="dialogMode"
      :title="dialogTitle"
      :fields="formFields"
      :data="formData"
      :rules="formRules"
      width="800px"
      @submit="handleSubmit"
    >
      <template #stationId="{ form }">
        <el-select
          v-model="form.stationId"
          placeholder="请选择所属电站"
          clearable
          filterable
          style="width: 100%"
        >
          <el-option
            v-for="station in stationList"
            :key="station.id"
            :label="station.name"
            :value="station.id"
          />
        </el-select>
      </template>
    </FormDialog>

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="设备详情"
      width="800px"
      destroy-on-close
    >
      <el-descriptions :column="2" border>
        <el-descriptions-item label="设备名称">
          {{ currentDevice?.name }}
        </el-descriptions-item>
        <el-descriptions-item label="设备编码">
          {{ currentDevice?.code }}
        </el-descriptions-item>
        <el-descriptions-item label="所属电站">
          {{ currentDevice?.stationName }}
        </el-descriptions-item>
        <el-descriptions-item label="设备类型">
          <el-tag :type="getTypeTagType(currentDevice?.type)">
            {{ getTypeName(currentDevice?.type) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="设备型号">
          {{ currentDevice?.model }}
        </el-descriptions-item>
        <el-descriptions-item label="制造商">
          {{ currentDevice?.manufacturer }}
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusTagType(currentDevice?.status)">
            {{ getStatusName(currentDevice?.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="安装日期">
          {{ currentDevice?.installDate }}
        </el-descriptions-item>
        <el-descriptions-item label="最后维护日期">
          {{ currentDevice?.lastMaintenanceDate }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ currentDevice?.createdAt }}
        </el-descriptions-item>
        <el-descriptions-item label="更新时间">
          {{ currentDevice?.updatedAt }}
        </el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">
          {{ currentDevice?.description || '-' }}
        </el-descriptions-item>
      </el-descriptions>

      <!-- 实时数据 -->
      <div class="device-realtime" style="margin-top: 20px">
        <h4>实时数据</h4>
        <el-empty v-if="!realtimeData" description="暂无实时数据" />
        <el-row v-else :gutter="16">
          <el-col :span="6" v-for="(value, key) in realtimeData" :key="key">
            <el-statistic :title="key" :value="value" />
          </el-col>
        </el-row>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus, Delete, Download } from '@element-plus/icons-vue'
import FormDialog from '@/components/FormDialog/index.vue'
import {
  getDeviceList,
  getDeviceDetail,
  createDevice,
  updateDevice,
  deleteDevice,
  batchDeleteDevices,
  updateDeviceStatus,
  getDeviceRealtimeData,
  getAllDevices
} from '@/api/device'
import { getAllStations } from '@/api/station'
import type { Device, DeviceType, DeviceStatus, Station } from '@/types'
import type { FormRules } from 'element-plus'

const loading = ref(false)
const tableData = ref<Device[]>([])
const total = ref(0)
const selectedIds = ref<number[]>([])
const stationList = ref<Station[]>([])
const dialogVisible = ref(false)
const dialogMode = ref<'add' | 'edit'>('add')
const formData = ref<Partial<Device>>({})
const detailVisible = ref(false)
const currentDevice = ref<Device | null>(null)
const realtimeData = ref<Record<string, any> | null>(null)

const queryParams = reactive({
  page: 1,
  pageSize: 20,
  keyword: '',
  stationId: undefined as number | undefined,
  type: undefined as DeviceType | undefined,
  status: undefined as DeviceStatus | undefined
})

// 对话框标题
const dialogTitle = computed(() => {
  return dialogMode.value === 'add' ? '新增设备' : '编辑设备'
})

// 表单字段
const formFields = computed(() => [
  {
    prop: 'stationId',
    label: '所属电站',
    type: 'select',
    required: true,
    span: 12
  },
  {
    prop: 'name',
    label: '设备名称',
    type: 'input',
    required: true,
    span: 12
  },
  {
    prop: 'code',
    label: '设备编码',
    type: 'input',
    required: true,
    span: 12
  },
  {
    prop: 'type',
    label: '设备类型',
    type: 'select',
    required: true,
    span: 12,
    options: [
      { label: '逆变器', value: 'inverter' },
      { label: '电表', value: 'meter' },
      { label: '传感器', value: 'sensor' },
      { label: '控制器', value: 'controller' },
      { label: '汇流箱', value: 'combiner' }
    ]
  },
  {
    prop: 'model',
    label: '设备型号',
    type: 'input',
    span: 12
  },
  {
    prop: 'manufacturer',
    label: '制造商',
    type: 'input',
    span: 12
  },
  {
    prop: 'status',
    label: '状态',
    type: 'select',
    span: 12,
    options: [
      { label: '在线', value: 'online' },
      { label: '离线', value: 'offline' },
      { label: '维护', value: 'maintenance' },
      { label: '故障', value: 'fault' }
    ],
    defaultValue: 'online'
  },
  {
    prop: 'installDate',
    label: '安装日期',
    type: 'date',
    span: 12,
    valueFormat: 'YYYY-MM-DD'
  },
  {
    prop: 'lastMaintenanceDate',
    label: '最后维护日期',
    type: 'date',
    span: 12,
    valueFormat: 'YYYY-MM-DD'
  },
  {
    prop: 'description',
    label: '描述',
    type: 'textarea',
    span: 24,
    rows: 3
  }
])

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入设备名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入设备编码', trigger: 'blur' },
    { pattern: /^[A-Z0-9_]+$/, message: '只能包含大写字母、数字和下划线', trigger: 'blur' }
  ],
  stationId: [
    { required: true, message: '请选择所属电站', trigger: 'change' }
  ],
  type: [
    { required: true, message: '请选择设备类型', trigger: 'change' }
  ]
}

// 获取电站列表
const fetchStationList = async () => {
  try {
    const data = await getAllStations()
    stationList.value = data
  } catch (error) {
    console.error('获取电站列表失败:', error)
  }
}

// 获取设备列表
const fetchData = async () => {
  try {
    loading.value = true
    const result = await getDeviceList(queryParams)
    tableData.value = result.list
    total.value = result.total
  } catch (error) {
    ElMessage.error('获取设备列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  queryParams.page = 1
  fetchData()
}

// 重置
const handleReset = () => {
  queryParams.keyword = ''
  queryParams.stationId = undefined
  queryParams.type = undefined
  queryParams.status = undefined
  queryParams.page = 1
  fetchData()
}

// 选择变化
const handleSelectionChange = (rows: Device[]) => {
  selectedIds.value = rows.map(row => row.id)
}

// 新增
const handleAdd = () => {
  dialogMode.value = 'add'
  formData.value = {
    status: 'online'
  }
  dialogVisible.value = true
}

// 编辑
const handleEdit = async (row: Device) => {
  try {
    const data = await getDeviceDetail(row.id)
    dialogMode.value = 'edit'
    formData.value = { ...data }
    dialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取设备详情失败')
  }
}

// 查看
const handleView = async (row: Device) => {
  try {
    currentDevice.value = await getDeviceDetail(row.id)
    realtimeData.value = await getDeviceRealtimeData(row.id)
    detailVisible.value = true
  } catch (error) {
    ElMessage.error('获取设备详情失败')
  }
}

// 删除
const handleDelete = async (row: Device) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除设备"${row.name}"吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await deleteDevice(row.id)
    ElMessage.success('删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 批量删除
const handleBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedIds.value.length} 个设备吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await batchDeleteDevices(selectedIds.value)
    ElMessage.success('批量删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

// 切换状态
const handleToggleStatus = async (row: Device) => {
  try {
    const newStatus = row.status === 'online' ? 'offline' : 'online'
    await updateDeviceStatus(row.id, newStatus)
    ElMessage.success('状态更新成功')
    fetchData()
  } catch (error) {
    ElMessage.error('状态更新失败')
  }
}

// 导出
const handleExport = () => {
  ElMessage.info('导出功能开发中...')
}

// 提交
const handleSubmit = async (data: any) => {
  try {
    if (dialogMode.value === 'add') {
      await createDevice(data)
      ElMessage.success('新增成功')
    } else {
      await updateDevice(data.id, data)
      ElMessage.success('更新成功')
    }
    
    dialogVisible.value = false
    fetchData()
  } catch (error) {
    ElMessage.error(dialogMode.value === 'add' ? '新增失败' : '更新失败')
  }
}

// 获取类型名称
const getTypeName = (type?: DeviceType) => {
  const typeMap: Record<DeviceType, string> = {
    inverter: '逆变器',
    meter: '电表',
    sensor: '传感器',
    controller: '控制器',
    combiner: '汇流箱'
  }
  return type ? typeMap[type] : '-'
}

// 获取类型标签类型
const getTypeTagType = (type?: DeviceType) => {
  const tagMap: Record<DeviceType, any> = {
    inverter: 'primary',
    meter: 'success',
    sensor: 'warning',
    controller: 'danger',
    combiner: 'info'
  }
  return type ? tagMap[type] : ''
}

// 获取状态名称
const getStatusName = (status?: DeviceStatus) => {
  const statusMap: Record<DeviceStatus, string> = {
    online: '在线',
    offline: '离线',
    maintenance: '维护',
    fault: '故障'
  }
  return status ? statusMap[status] : '-'
}

// 获取状态标签类型
const getStatusTagType = (status?: DeviceStatus) => {
  const tagMap: Record<DeviceStatus, any> = {
    online: 'success',
    offline: 'info',
    maintenance: 'warning',
    fault: 'danger'
  }
  return status ? tagMap[status] : ''
}

onMounted(() => {
  fetchStationList()
  fetchData()
})
</script>

<style scoped lang="scss">
.device-management {
  .filter-card,
  .toolbar-card {
    margin-bottom: 16px;
  }

  .table-card {
    .pagination {
      display: flex;
      justify-content: flex-end;
      margin-top: 16px;
    }
  }

  .device-realtime {
    h4 {
      margin-bottom: 16px;
      font-size: 16px;
      font-weight: 500;
    }
  }
}
</style>
