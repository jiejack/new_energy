<template>
  <div class="station-management">
    <!-- 搜索筛选 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="queryParams" size="default">
        <el-form-item label="电站名称">
          <el-input
            v-model="queryParams.keyword"
            placeholder="请输入电站名称或编码"
            clearable
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="所属区域">
          <el-tree-select
            v-model="queryParams.regionId"
            :data="regionTree"
            :props="{ label: 'name', value: 'id' }"
            placeholder="请选择区域"
            clearable
            check-strictly
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item label="电站类型">
          <el-select
            v-model="queryParams.type"
            placeholder="请选择类型"
            clearable
            style="width: 150px"
          >
            <el-option label="光伏电站" value="solar" />
            <el-option label="风力电站" value="wind" />
            <el-option label="水力电站" value="hydro" />
            <el-option label="储能电站" value="storage" />
          </el-select>
        </el-form-item>
        <el-form-item label="电站状态">
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
        新增电站
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
        <el-table-column prop="name" label="电站名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="code" label="电站编码" width="120" show-overflow-tooltip />
        <el-table-column prop="regionName" label="所属区域" width="120" show-overflow-tooltip />
        <el-table-column prop="type" label="电站类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getTypeTagType(row.type)">
              {{ getTypeName(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="capacity" label="装机容量(MW)" width="120" align="center" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)">
              {{ getStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="address" label="地址" min-width="200" show-overflow-tooltip />
        <el-table-column prop="createdAt" label="创建时间" width="160" align="center" />
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
      <template #regionId="{ form }">
        <el-tree-select
          v-model="form.regionId"
          :data="regionTree"
          :props="{ label: 'name', value: 'id' }"
          placeholder="请选择所属区域"
          clearable
          check-strictly
          style="width: 100%"
        />
      </template>
      <template #location="{ form }">
        <el-row :gutter="10">
          <el-col :span="12">
            <el-input
              v-model.number="form.longitude"
              placeholder="经度"
              type="number"
            >
              <template #prepend>经度</template>
            </el-input>
          </el-col>
          <el-col :span="12">
            <el-input
              v-model.number="form.latitude"
              placeholder="纬度"
              type="number"
            >
              <template #prepend>纬度</template>
            </el-input>
          </el-col>
        </el-row>
        <div style="margin-top: 10px">
          <el-button size="small" @click="handleSelectLocation(form)">
            <el-icon><Location /></el-icon>
            地图选点
          </el-button>
        </div>
      </template>
    </FormDialog>

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="电站详情"
      width="800px"
      destroy-on-close
    >
      <el-descriptions :column="2" border>
        <el-descriptions-item label="电站名称">
          {{ currentStation?.name }}
        </el-descriptions-item>
        <el-descriptions-item label="电站编码">
          {{ currentStation?.code }}
        </el-descriptions-item>
        <el-descriptions-item label="所属区域">
          {{ currentStation?.regionName }}
        </el-descriptions-item>
        <el-descriptions-item label="电站类型">
          <el-tag :type="getTypeTagType(currentStation?.type)">
            {{ getTypeName(currentStation?.type) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="装机容量">
          {{ currentStation?.capacity }} MW
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusTagType(currentStation?.status)">
            {{ getStatusName(currentStation?.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="地址" :span="2">
          {{ currentStation?.address }}
        </el-descriptions-item>
        <el-descriptions-item label="经度">
          {{ currentStation?.longitude }}
        </el-descriptions-item>
        <el-descriptions-item label="纬度">
          {{ currentStation?.latitude }}
        </el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">
          {{ currentStation?.description || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ currentStation?.createdAt }}
        </el-descriptions-item>
        <el-descriptions-item label="更新时间">
          {{ currentStation?.updatedAt }}
        </el-descriptions-item>
      </el-descriptions>

      <!-- 统计信息 -->
      <div class="station-stats" style="margin-top: 20px">
        <h4>统计信息</h4>
        <el-row :gutter="16">
          <el-col :span="6">
            <el-statistic title="设备总数" :value="stationStats.deviceCount" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="在线设备" :value="stationStats.onlineDeviceCount" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="离线设备" :value="stationStats.offlineDeviceCount" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="告警数量" :value="stationStats.alarmCount" />
          </el-col>
        </el-row>
      </div>
    </el-dialog>

    <!-- 地图选点弹窗 -->
    <el-dialog
      v-model="mapVisible"
      title="地图选点"
      width="800px"
      destroy-on-close
    >
      <div class="map-container">
        <el-alert
          type="info"
          :closable="false"
          style="margin-bottom: 10px"
        >
          请在下方输入框中输入经纬度坐标，或点击地图选择位置
        </el-alert>
        <el-row :gutter="10" style="margin-bottom: 10px">
          <el-col :span="12">
            <el-input
              v-model.number="tempLocation.longitude"
              placeholder="经度"
              type="number"
            >
              <template #prepend>经度</template>
            </el-input>
          </el-col>
          <el-col :span="12">
            <el-input
              v-model.number="tempLocation.latitude"
              placeholder="纬度"
              type="number"
            >
              <template #prepend>纬度</template>
            </el-input>
          </el-col>
        </el-row>
        <div class="map-placeholder">
          <el-empty description="地图功能待集成（可接入百度地图/高德地图）" />
        </div>
      </div>
      <template #footer>
        <el-button @click="mapVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfirmLocation">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus, Delete, Download, Location } from '@element-plus/icons-vue'
import FormDialog from '@/components/FormDialog/index.vue'
import {
  getStationList,
  getStationDetail,
  createStation,
  updateStation,
  deleteStation,
  batchDeleteStations,
  updateStationStatus,
  getStationStatistics,
  getAllStations
} from '@/api/station'
import { getRegionTree } from '@/api/region'
import type { Station, StationType, StationStatus, Region } from '@/types'
import type { FormRules } from 'element-plus'

const loading = ref(false)
const tableData = ref<Station[]>([])
const total = ref(0)
const selectedIds = ref<number[]>([])
const regionTree = ref<Region[]>([])
const dialogVisible = ref(false)
const dialogMode = ref<'add' | 'edit'>('add')
const formData = ref<Partial<Station>>({})
const detailVisible = ref(false)
const currentStation = ref<Station | null>(null)
const mapVisible = ref(false)
const tempLocation = ref({ longitude: 0, latitude: 0 })
const currentFormRef = ref<any>(null)

const queryParams = reactive({
  page: 1,
  pageSize: 20,
  keyword: '',
  regionId: undefined as number | undefined,
  type: undefined as StationType | undefined,
  status: undefined as StationStatus | undefined
})

const stationStats = ref({
  deviceCount: 0,
  onlineDeviceCount: 0,
  offlineDeviceCount: 0,
  alarmCount: 0
})

// 对话框标题
const dialogTitle = computed(() => {
  return dialogMode.value === 'add' ? '新增电站' : '编辑电站'
})

// 表单字段
const formFields = computed(() => [
  {
    prop: 'regionId',
    label: '所属区域',
    type: 'select',
    required: true,
    span: 12
  },
  {
    prop: 'name',
    label: '电站名称',
    type: 'input',
    required: true,
    span: 12
  },
  {
    prop: 'code',
    label: '电站编码',
    type: 'input',
    required: true,
    span: 12
  },
  {
    prop: 'type',
    label: '电站类型',
    type: 'select',
    required: true,
    span: 12,
    options: [
      { label: '光伏电站', value: 'solar' },
      { label: '风力电站', value: 'wind' },
      { label: '水力电站', value: 'hydro' },
      { label: '储能电站', value: 'storage' }
    ]
  },
  {
    prop: 'capacity',
    label: '装机容量(MW)',
    type: 'number',
    required: true,
    span: 12,
    min: 0,
    precision: 2
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
    prop: 'address',
    label: '地址',
    type: 'input',
    span: 24
  },
  {
    prop: 'location',
    label: '地理位置',
    span: 24
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
    { required: true, message: '请输入电站名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入电站编码', trigger: 'blur' },
    { pattern: /^[A-Z0-9_]+$/, message: '只能包含大写字母、数字和下划线', trigger: 'blur' }
  ],
  regionId: [
    { required: true, message: '请选择所属区域', trigger: 'change' }
  ],
  type: [
    { required: true, message: '请选择电站类型', trigger: 'change' }
  ],
  capacity: [
    { required: true, message: '请输入装机容量', trigger: 'blur' },
    { type: 'number', min: 0, message: '装机容量必须大于等于0', trigger: 'blur' }
  ]
}

// 获取区域树
const fetchRegionTree = async () => {
  try {
    const data = await getRegionTree()
    regionTree.value = data
  } catch (error) {
    console.error('获取区域树失败:', error)
  }
}

// 获取电站列表
const fetchData = async () => {
  try {
    loading.value = true
    const result = await getStationList(queryParams)
    tableData.value = result.list
    total.value = result.total
  } catch (error) {
    ElMessage.error('获取电站列表失败')
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
  queryParams.regionId = undefined
  queryParams.type = undefined
  queryParams.status = undefined
  queryParams.page = 1
  fetchData()
}

// 选择变化
const handleSelectionChange = (rows: Station[]) => {
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
const handleEdit = async (row: Station) => {
  try {
    const data = await getStationDetail(row.id)
    dialogMode.value = 'edit'
    formData.value = { ...data }
    dialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取电站详情失败')
  }
}

// 查看
const handleView = async (row: Station) => {
  try {
    currentStation.value = await getStationDetail(row.id)
    const stats = await getStationStatistics(row.id)
    stationStats.value = stats
    detailVisible.value = true
  } catch (error) {
    ElMessage.error('获取电站详情失败')
  }
}

// 删除
const handleDelete = async (row: Station) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除电站"${row.name}"吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await deleteStation(row.id)
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
      `确定要删除选中的 ${selectedIds.value.length} 个电站吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await batchDeleteStations(selectedIds.value)
    ElMessage.success('批量删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

// 切换状态
const handleToggleStatus = async (row: Station) => {
  try {
    const newStatus = row.status === 'online' ? 'offline' : 'online'
    await updateStationStatus(row.id, newStatus)
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

// 地图选点
const handleSelectLocation = (form: any) => {
  currentFormRef.value = form
  tempLocation.value = {
    longitude: form.longitude || 0,
    latitude: form.latitude || 0
  }
  mapVisible.value = true
}

// 确认位置
const handleConfirmLocation = () => {
  if (currentFormRef.value) {
    currentFormRef.value.longitude = tempLocation.value.longitude
    currentFormRef.value.latitude = tempLocation.value.latitude
  }
  mapVisible.value = false
}

// 提交
const handleSubmit = async (data: any) => {
  try {
    if (dialogMode.value === 'add') {
      await createStation(data)
      ElMessage.success('新增成功')
    } else {
      await updateStation(data.id, data)
      ElMessage.success('更新成功')
    }
    
    dialogVisible.value = false
    fetchData()
  } catch (error) {
    ElMessage.error(dialogMode.value === 'add' ? '新增失败' : '更新失败')
  }
}

// 获取类型名称
const getTypeName = (type?: StationType) => {
  const typeMap: Record<StationType, string> = {
    solar: '光伏',
    wind: '风电',
    hydro: '水电',
    storage: '储能'
  }
  return type ? typeMap[type] : '-'
}

// 获取类型标签类型
const getTypeTagType = (type?: StationType) => {
  const tagMap: Record<StationType, any> = {
    solar: 'warning',
    wind: 'success',
    hydro: 'primary',
    storage: 'info'
  }
  return type ? tagMap[type] : ''
}

// 获取状态名称
const getStatusName = (status?: StationStatus) => {
  const statusMap: Record<StationStatus, string> = {
    online: '在线',
    offline: '离线',
    maintenance: '维护',
    fault: '故障'
  }
  return status ? statusMap[status] : '-'
}

// 获取状态标签类型
const getStatusTagType = (status?: StationStatus) => {
  const tagMap: Record<StationStatus, any> = {
    online: 'success',
    offline: 'info',
    maintenance: 'warning',
    fault: 'danger'
  }
  return status ? tagMap[status] : ''
}

onMounted(() => {
  fetchRegionTree()
  fetchData()
})
</script>

<style scoped lang="scss">
.station-management {
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

  .station-stats {
    h4 {
      margin-bottom: 16px;
      font-size: 16px;
      font-weight: 500;
    }
  }

  .map-container {
    .map-placeholder {
      height: 400px;
      border: 1px solid #dcdfe6;
      border-radius: 4px;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }
}
</style>
