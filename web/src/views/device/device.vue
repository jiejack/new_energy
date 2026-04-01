<template>
  <div class="device-manage-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>设备管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增设备
          </el-button>
        </div>
      </template>

      <!-- 搜索栏 -->
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="设备名称">
          <el-input v-model="searchForm.name" placeholder="请输入设备名称" clearable />
        </el-form-item>
        <el-form-item label="所属电站">
          <el-select v-model="searchForm.stationId" placeholder="请选择电站" clearable style="width: 180px">
            <el-option label="北京朝阳光伏电站" :value="1" />
            <el-option label="上海浦东风电场" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="设备类型">
          <el-select v-model="searchForm.type" placeholder="请选择类型" clearable style="width: 150px">
            <el-option label="逆变器" value="inverter" />
            <el-option label="汇流箱" value="combiner" />
            <el-option label="风机" value="turbine" />
            <el-option label="储能单元" value="battery" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 数据表格 -->
      <el-table :data="deviceList" v-loading="loading" stripe>
        <el-table-column prop="name" label="设备名称" min-width="160" />
        <el-table-column prop="code" label="设备编码" width="130" />
        <el-table-column prop="stationName" label="所属电站" min-width="150" />
        <el-table-column prop="type" label="设备类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getTypeType(row.type)">{{ getTypeText(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="model" label="设备型号" width="130" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="onlineTime" label="最近上线" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleView(row)">查看</el-button>
            <el-button type="primary" link @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @change="handlePageChange"
        class="pagination"
      />
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="设备名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入设备名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="设备编码" prop="code">
              <el-input v-model="form.code" placeholder="请输入设备编码" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="所属电站" prop="stationId">
              <el-select v-model="form.stationId" placeholder="请选择电站" style="width: 100%">
                <el-option label="北京朝阳光伏电站" :value="1" />
                <el-option label="上海浦东风电场" :value="2" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="设备类型" prop="type">
              <el-select v-model="form.type" placeholder="请选择类型" style="width: 100%">
                <el-option label="逆变器" value="inverter" />
                <el-option label="汇流箱" value="combiner" />
                <el-option label="风机" value="turbine" />
                <el-option label="储能单元" value="battery" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="设备型号" prop="model">
          <el-input v-model="form.model" placeholder="请输入设备型号" />
        </el-form-item>
        <el-form-item label="生产厂家" prop="manufacturer">
          <el-input v-model="form.manufacturer" placeholder="请输入生产厂家" />
        </el-form-item>
        <el-form-item label="安装日期" prop="installDate">
          <el-date-picker v-model="form.installDate" type="date" placeholder="选择日期" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="form.remark" type="textarea" :rows="3" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Plus, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

// 加载状态
const loading = ref(false)

// 搜索表单
const searchForm = reactive({
  name: '',
  stationId: undefined as number | undefined,
  type: ''
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 5
})

// 设备列表数据
const deviceList = ref([
  {
    id: 1,
    name: '逆变器 #01',
    code: 'INV-BJ-001',
    stationName: '北京朝阳光伏电站',
    type: 'inverter',
    model: 'SUN-50K-G03',
    status: 'online',
    onlineTime: '2024-01-15 14:30:00'
  },
  {
    id: 2,
    name: '逆变器 #02',
    code: 'INV-BJ-002',
    stationName: '北京朝阳光伏电站',
    type: 'inverter',
    model: 'SUN-50K-G03',
    status: 'online',
    onlineTime: '2024-01-15 14:30:00'
  },
  {
    id: 3,
    name: '汇流箱 #01',
    code: 'CB-BJ-001',
    stationName: '北京朝阳光伏电站',
    type: 'combiner',
    model: 'CB-16',
    status: 'online',
    onlineTime: '2024-01-15 14:30:00'
  },
  {
    id: 4,
    name: '风机 #01',
    code: 'WT-SH-001',
    stationName: '上海浦东风电场',
    type: 'turbine',
    model: 'GW-1500',
    status: 'offline',
    onlineTime: '2024-01-14 08:00:00'
  },
  {
    id: 5,
    name: '储能单元 #01',
    code: 'BAT-GZ-001',
    stationName: '广州番禺储能站',
    type: 'battery',
    model: 'BYD-B10',
    status: 'online',
    onlineTime: '2024-01-15 14:30:00'
  }
])

// 对话框相关
const dialogVisible = ref(false)
const dialogTitle = ref('新增设备')
const formRef = ref<FormInstance>()
const isEdit = ref(false)

// 表单数据
const form = reactive({
  id: undefined as number | undefined,
  name: '',
  code: '',
  stationId: undefined as number | undefined,
  type: '',
  model: '',
  manufacturer: '',
  installDate: undefined as Date | undefined,
  remark: ''
})

// 表单校验规则
const rules: FormRules = {
  name: [{ required: true, message: '请输入设备名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入设备编码', trigger: 'blur' }],
  stationId: [{ required: true, message: '请选择所属电站', trigger: 'change' }],
  type: [{ required: true, message: '请选择设备类型', trigger: 'change' }],
  model: [{ required: true, message: '请输入设备型号', trigger: 'blur' }]
}

// 获取类型样式
function getTypeType(type: string) {
  const types: Record<string, string> = {
    inverter: 'success',
    combiner: 'primary',
    turbine: 'warning',
    battery: 'info'
  }
  return types[type] || 'info'
}

// 获取类型文本
function getTypeText(type: string) {
  const texts: Record<string, string> = {
    inverter: '逆变器',
    combiner: '汇流箱',
    turbine: '风机',
    battery: '储能单元'
  }
  return texts[type] || '其他'
}

// 获取状态样式
function getStatusType(status: string) {
  const types: Record<string, string> = {
    online: 'success',
    offline: 'info',
    fault: 'danger',
    maintenance: 'warning'
  }
  return types[status] || 'info'
}

// 获取状态文本
function getStatusText(status: string) {
  const texts: Record<string, string> = {
    online: '在线',
    offline: '离线',
    fault: '故障',
    maintenance: '维护'
  }
  return texts[status] || '未知'
}

// 搜索
function handleSearch() {
  ElMessage.success('搜索完成')
}

// 重置
function handleReset() {
  searchForm.name = ''
  searchForm.stationId = undefined
  searchForm.type = ''
}

// 分页变化
function handlePageChange() {
  // 加载数据
}

// 新增设备
function handleAdd() {
  isEdit.value = false
  dialogTitle.value = '新增设备'
  resetForm()
  dialogVisible.value = true
}

// 查看设备
function handleView(row: any) {
  ElMessage.info(`查看设备: ${row.name}`)
}

// 编辑设备
function handleEdit(row: any) {
  isEdit.value = true
  dialogTitle.value = '编辑设备'
  Object.assign(form, row)
  dialogVisible.value = true
}

// 删除设备
async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `确定要删除设备 "${row.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    ElMessage.success('删除成功')
  } catch {
    // 用户取消
  }
}

// 提交表单
async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate((valid) => {
    if (valid) {
      ElMessage.success(isEdit.value ? '编辑成功' : '新增成功')
      dialogVisible.value = false
    }
  })
}

// 重置表单
function resetForm() {
  form.id = undefined
  form.name = ''
  form.code = ''
  form.stationId = undefined
  form.type = ''
  form.model = ''
  form.manufacturer = ''
  form.installDate = undefined
  form.remark = ''
}

onMounted(() => {
  loading.value = false
})
</script>

<style scoped lang="scss">
.device-manage-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-form {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  justify-content: flex-end;
}
</style>
