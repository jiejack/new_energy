<template>
  <div class="point-management">
    <!-- 搜索筛选 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="queryParams" size="default">
        <el-form-item label="采集点名称">
          <el-input
            v-model="queryParams.keyword"
            placeholder="请输入采集点名称或编码"
            clearable
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="所属设备">
          <el-select
            v-model="queryParams.deviceId"
            placeholder="请选择设备"
            clearable
            filterable
            style="width: 200px"
          >
            <el-option
              v-for="device in deviceList"
              :key="device.id"
              :label="device.name"
              :value="device.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="采集点类型">
          <el-select
            v-model="queryParams.type"
            placeholder="请选择类型"
            clearable
            style="width: 150px"
          >
            <el-option label="模拟量" value="analog" />
            <el-option label="数字量" value="digital" />
            <el-option label="脉冲量" value="pulse" />
          </el-select>
        </el-form-item>
        <el-form-item label="数据类型">
          <el-select
            v-model="queryParams.dataType"
            placeholder="请选择数据类型"
            clearable
            style="width: 150px"
          >
            <el-option label="浮点数" value="float" />
            <el-option label="整数" value="int" />
            <el-option label="布尔值" value="bool" />
            <el-option label="字符串" value="string" />
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
        新增采集点
      </el-button>
      <el-button
        type="danger"
        :disabled="selectedIds.length === 0"
        @click="handleBatchDelete"
      >
        <el-icon><Delete /></el-icon>
        批量删除
      </el-button>
      <el-button @click="handleImport">
        <el-icon><Upload /></el-icon>
        批量导入
      </el-button>
      <el-button @click="handleExport">
        <el-icon><Download /></el-icon>
        配置导出
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
        <el-table-column prop="name" label="采集点名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="code" label="采集点编码" width="120" show-overflow-tooltip />
        <el-table-column prop="deviceName" label="所属设备" width="120" show-overflow-tooltip />
        <el-table-column prop="type" label="采集点类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getTypeTagType(row.type)">
              {{ getTypeName(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="dataType" label="数据类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag type="info">
              {{ getDataTypeName(row.dataType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="unit" label="单位" width="80" align="center" />
        <el-table-column prop="minValue" label="最小值" width="90" align="center" />
        <el-table-column prop="maxValue" label="最大值" width="90" align="center" />
        <el-table-column prop="createdAt" label="创建时间" width="160" align="center" />
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleView(row)">
              查看
            </el-button>
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
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
      <template #deviceId="{ form }">
        <el-select
          v-model="form.deviceId"
          placeholder="请选择所属设备"
          clearable
          filterable
          style="width: 100%"
        >
          <el-option
            v-for="device in deviceList"
            :key="device.id"
            :label="device.name"
            :value="device.id"
          />
        </el-select>
      </template>
    </FormDialog>

    <!-- 详情弹窗 -->
    <el-dialog
      v-model="detailVisible"
      title="采集点详情"
      width="800px"
      destroy-on-close
    >
      <el-descriptions :column="2" border>
        <el-descriptions-item label="采集点名称">
          {{ currentPoint?.name }}
        </el-descriptions-item>
        <el-descriptions-item label="采集点编码">
          {{ currentPoint?.code }}
        </el-descriptions-item>
        <el-descriptions-item label="所属设备">
          {{ currentPoint?.deviceName }}
        </el-descriptions-item>
        <el-descriptions-item label="采集点类型">
          <el-tag :type="getTypeTagType(currentPoint?.type)">
            {{ getTypeName(currentPoint?.type) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="数据类型">
          <el-tag type="info">
            {{ getDataTypeName(currentPoint?.dataType) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="单位">
          {{ currentPoint?.unit || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="最小值">
          {{ currentPoint?.minValue ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="最大值">
          {{ currentPoint?.maxValue ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ currentPoint?.createdAt }}
        </el-descriptions-item>
        <el-descriptions-item label="更新时间">
          {{ currentPoint?.updatedAt }}
        </el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">
          {{ currentPoint?.description || '-' }}
        </el-descriptions-item>
      </el-descriptions>

      <!-- 实时数据 -->
      <div class="point-realtime" style="margin-top: 20px">
        <h4>实时数据</h4>
        <el-empty v-if="!realtimeData" description="暂无实时数据" />
        <el-descriptions v-else :column="3" border>
          <el-descriptions-item label="当前值">
            {{ realtimeData.value }}
          </el-descriptions-item>
          <el-descriptions-item label="数据质量">
            <el-tag :type="realtimeData.quality === 1 ? 'success' : 'danger'">
              {{ realtimeData.quality === 1 ? '良好' : '异常' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="采集时间">
            {{ new Date(realtimeData.timestamp).toLocaleString() }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>

    <!-- 批量导入弹窗 -->
    <el-dialog
      v-model="importVisible"
      title="批量导入"
      width="600px"
      destroy-on-close
    >
      <el-alert
        type="info"
        :closable="false"
        style="margin-bottom: 20px"
      >
        <template #title>
          导入说明
        </template>
        <div>
          1. 请下载导入模板，按照模板格式填写数据<br>
          2. 支持 Excel (.xlsx, .xls) 格式<br>
          3. 单次导入不超过 1000 条数据
        </div>
      </el-alert>

      <el-upload
        ref="uploadRef"
        :auto-upload="false"
        :limit="1"
        accept=".xlsx,.xls"
        :on-change="handleFileChange"
        :on-exceed="handleExceed"
        drag
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">
          将文件拖到此处，或<em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            只能上传 xlsx/xls 文件
          </div>
        </template>
      </el-upload>

      <div style="margin-top: 20px">
        <el-button @click="handleDownloadTemplate">
          <el-icon><Download /></el-icon>
          下载导入模板
        </el-button>
      </div>

      <template #footer>
        <el-button @click="importVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfirmImport" :loading="importLoading">
          确定导入
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus, Delete, Download, Upload, UploadFilled } from '@element-plus/icons-vue'
import FormDialog from '@/components/FormDialog/index.vue'
import {
  getPointList,
  getPointDetail,
  createPoint,
  updatePoint,
  deletePoint,
  batchDeletePoints,
  getPointRealtimeData
} from '@/api/point'
import { getAllDevices } from '@/api/device'
import type { Point, PointType, DataType, Device } from '@/types'
import type { FormRules, UploadInstance, UploadProps, UploadUserFile } from 'element-plus'

const loading = ref(false)
const tableData = ref<Point[]>([])
const total = ref(0)
const selectedIds = ref<number[]>([])
const deviceList = ref<Device[]>([])
const dialogVisible = ref(false)
const dialogMode = ref<'add' | 'edit'>('add')
const formData = ref<Partial<Point>>({})
const detailVisible = ref(false)
const currentPoint = ref<Point | null>(null)
const realtimeData = ref<{ value: number; quality: number; timestamp: number } | null>(null)
const importVisible = ref(false)
const importLoading = ref(false)
const uploadRef = ref<UploadInstance>()
const uploadFile = ref<File | null>(null)

const queryParams = reactive({
  page: 1,
  pageSize: 20,
  keyword: '',
  deviceId: undefined as number | undefined,
  type: undefined as PointType | undefined,
  dataType: undefined as DataType | undefined
})

// 对话框标题
const dialogTitle = computed(() => {
  return dialogMode.value === 'add' ? '新增采集点' : '编辑采集点'
})

// 表单字段
const formFields = computed(() => [
  {
    prop: 'deviceId',
    label: '所属设备',
    type: 'select',
    required: true,
    span: 12
  },
  {
    prop: 'name',
    label: '采集点名称',
    type: 'input',
    required: true,
    span: 12
  },
  {
    prop: 'code',
    label: '采集点编码',
    type: 'input',
    required: true,
    span: 12
  },
  {
    prop: 'type',
    label: '采集点类型',
    type: 'select',
    required: true,
    span: 12,
    options: [
      { label: '模拟量', value: 'analog' },
      { label: '数字量', value: 'digital' },
      { label: '脉冲量', value: 'pulse' }
    ]
  },
  {
    prop: 'dataType',
    label: '数据类型',
    type: 'select',
    required: true,
    span: 12,
    options: [
      { label: '浮点数', value: 'float' },
      { label: '整数', value: 'int' },
      { label: '布尔值', value: 'bool' },
      { label: '字符串', value: 'string' }
    ]
  },
  {
    prop: 'unit',
    label: '单位',
    type: 'input',
    span: 12
  },
  {
    prop: 'minValue',
    label: '最小值',
    type: 'number',
    span: 12,
    precision: 2
  },
  {
    prop: 'maxValue',
    label: '最大值',
    type: 'number',
    span: 12,
    precision: 2
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
    { required: true, message: '请输入采集点名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入采集点编码', trigger: 'blur' },
    { pattern: /^[A-Z0-9_]+$/, message: '只能包含大写字母、数字和下划线', trigger: 'blur' }
  ],
  deviceId: [
    { required: true, message: '请选择所属设备', trigger: 'change' }
  ],
  type: [
    { required: true, message: '请选择采集点类型', trigger: 'change' }
  ],
  dataType: [
    { required: true, message: '请选择数据类型', trigger: 'change' }
  ]
}

// 获取设备列表
const fetchDeviceList = async () => {
  try {
    const data = await getAllDevices()
    deviceList.value = data
  } catch (error) {
    console.error('获取设备列表失败:', error)
  }
}

// 获取采集点列表
const fetchData = async () => {
  try {
    loading.value = true
    const result = await getPointList(queryParams)
    tableData.value = result.list
    total.value = result.total
  } catch (error) {
    ElMessage.error('获取采集点列表失败')
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
  queryParams.deviceId = undefined
  queryParams.type = undefined
  queryParams.dataType = undefined
  queryParams.page = 1
  fetchData()
}

// 选择变化
const handleSelectionChange = (rows: Point[]) => {
  selectedIds.value = rows.map(row => row.id)
}

// 新增
const handleAdd = () => {
  dialogMode.value = 'add'
  formData.value = {}
  dialogVisible.value = true
}

// 编辑
const handleEdit = async (row: Point) => {
  try {
    const data = await getPointDetail(row.id)
    dialogMode.value = 'edit'
    formData.value = { ...data }
    dialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取采集点详情失败')
  }
}

// 查看
const handleView = async (row: Point) => {
  try {
    currentPoint.value = await getPointDetail(row.id)
    realtimeData.value = await getPointRealtimeData(row.id)
    detailVisible.value = true
  } catch (error) {
    ElMessage.error('获取采集点详情失败')
  }
}

// 删除
const handleDelete = async (row: Point) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除采集点"${row.name}"吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await deletePoint(row.id)
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
      `确定要删除选中的 ${selectedIds.value.length} 个采集点吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await batchDeletePoints(selectedIds.value)
    ElMessage.success('批量删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

// 导入
const handleImport = () => {
  uploadFile.value = null
  importVisible.value = true
}

// 文件变化
const handleFileChange: UploadProps['onChange'] = (uploadFile) => {
  uploadFile.value = uploadFile.raw as File
}

// 文件超出限制
const handleExceed: UploadProps['onExceed'] = () => {
  ElMessage.warning('只能上传一个文件')
}

// 下载模板
const handleDownloadTemplate = () => {
  ElMessage.info('模板下载功能开发中...')
}

// 确认导入
const handleConfirmImport = async () => {
  if (!uploadFile.value) {
    ElMessage.warning('请选择要导入的文件')
    return
  }

  try {
    importLoading.value = true
    // 这里应该调用导入API
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success('导入成功')
    importVisible.value = false
    fetchData()
  } catch (error) {
    ElMessage.error('导入失败')
  } finally {
    importLoading.value = false
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
      await createPoint(data)
      ElMessage.success('新增成功')
    } else {
      await updatePoint(data.id, data)
      ElMessage.success('更新成功')
    }
    
    dialogVisible.value = false
    fetchData()
  } catch (error) {
    ElMessage.error(dialogMode.value === 'add' ? '新增失败' : '更新失败')
  }
}

// 获取类型名称
const getTypeName = (type?: PointType) => {
  const typeMap: Record<PointType, string> = {
    analog: '模拟量',
    digital: '数字量',
    pulse: '脉冲量'
  }
  return type ? typeMap[type] : '-'
}

// 获取类型标签类型
const getTypeTagType = (type?: PointType) => {
  const tagMap: Record<PointType, any> = {
    analog: 'primary',
    digital: 'success',
    pulse: 'warning'
  }
  return type ? tagMap[type] : ''
}

// 获取数据类型名称
const getDataTypeName = (dataType?: DataType) => {
  const typeMap: Record<DataType, string> = {
    float: '浮点数',
    int: '整数',
    bool: '布尔值',
    string: '字符串'
  }
  return dataType ? typeMap[dataType] : '-'
}

onMounted(() => {
  fetchDeviceList()
  fetchData()
})
</script>

<style scoped lang="scss">
.point-management {
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

  .point-realtime {
    h4 {
      margin-bottom: 16px;
      font-size: 16px;
      font-weight: 500;
    }
  }

  :deep(.el-upload-dragger) {
    padding: 30px;
  }
}
</style>
