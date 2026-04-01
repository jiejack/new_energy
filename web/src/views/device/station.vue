<template>
  <div class="station-manage-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>电站管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增电站
          </el-button>
        </div>
      </template>

      <!-- 搜索栏 -->
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="电站名称">
          <el-input v-model="searchForm.name" placeholder="请输入电站名称" clearable />
        </el-form-item>
        <el-form-item label="所属区域">
          <el-select v-model="searchForm.regionId" placeholder="请选择区域" clearable>
            <el-option label="华北地区" :value="1" />
            <el-option label="华东地区" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="电站状态">
          <el-select v-model="searchForm.status" placeholder="请选择状态" clearable>
            <el-option label="在线" :value="1" />
            <el-option label="离线" :value="0" />
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
      <el-table :data="stationList" v-loading="loading" stripe>
        <el-table-column prop="name" label="电站名称" min-width="180" />
        <el-table-column prop="code" label="电站编码" width="120" />
        <el-table-column prop="regionName" label="所属区域" width="120" />
        <el-table-column prop="type" label="电站类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getTypeType(row.type)">{{ getTypeText(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="capacity" label="装机容量(kW)" width="130" align="right" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="address" label="地址" min-width="200" show-overflow-tooltip />
        <el-table-column prop="createTime" label="创建时间" width="180" />
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
            <el-form-item label="电站名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入电站名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="电站编码" prop="code">
              <el-input v-model="form.code" placeholder="请输入电站编码" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="所属区域" prop="regionId">
              <el-select v-model="form.regionId" placeholder="请选择区域" style="width: 100%">
                <el-option label="华北地区" :value="1" />
                <el-option label="华东地区" :value="2" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="电站类型" prop="type">
              <el-select v-model="form.type" placeholder="请选择类型" style="width: 100%">
                <el-option label="光伏电站" value="solar" />
                <el-option label="风电场" value="wind" />
                <el-option label="储能站" value="storage" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="装机容量" prop="capacity">
          <el-input-number v-model="form.capacity" :min="0" style="width: 200px" />
          <span class="unit">kW</span>
        </el-form-item>
        <el-form-item label="电站地址" prop="address">
          <el-input v-model="form.address" placeholder="请输入电站地址" />
        </el-form-item>
        <el-form-item label="经度" prop="longitude">
          <el-input-number v-model="form.longitude" :precision="6" style="width: 200px" />
        </el-form-item>
        <el-form-item label="纬度" prop="latitude">
          <el-input-number v-model="form.latitude" :precision="6" style="width: 200px" />
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
  regionId: undefined as number | undefined,
  status: undefined as number | undefined
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 4
})

// 电站列表数据
const stationList = ref([
  {
    id: 1,
    name: '北京朝阳光伏电站',
    code: 'BJ-CY-001',
    regionName: '华北地区',
    type: 'solar',
    capacity: 5000,
    status: 1,
    address: '北京市朝阳区XXX路XXX号',
    createTime: '2024-01-01 10:00:00'
  },
  {
    id: 2,
    name: '上海浦东风电场',
    code: 'SH-PD-001',
    regionName: '华东地区',
    type: 'wind',
    capacity: 10000,
    status: 1,
    address: '上海市浦东新区XXX路XXX号',
    createTime: '2024-01-01 10:00:00'
  },
  {
    id: 3,
    name: '广州番禺储能站',
    code: 'GZ-PY-001',
    regionName: '华东地区',
    type: 'storage',
    capacity: 2000,
    status: 0,
    address: '广州市番禺区XXX路XXX号',
    createTime: '2024-01-01 10:00:00'
  },
  {
    id: 4,
    name: '深圳南山光伏电站',
    code: 'SZ-NS-001',
    regionName: '华东地区',
    type: 'solar',
    capacity: 8000,
    status: 1,
    address: '深圳市南山区XXX路XXX号',
    createTime: '2024-01-01 10:00:00'
  }
])

// 对话框相关
const dialogVisible = ref(false)
const dialogTitle = ref('新增电站')
const formRef = ref<FormInstance>()
const isEdit = ref(false)

// 表单数据
const form = reactive({
  id: undefined as number | undefined,
  name: '',
  code: '',
  regionId: undefined as number | undefined,
  type: '',
  capacity: 0,
  address: '',
  longitude: undefined as number | undefined,
  latitude: undefined as number | undefined,
  remark: ''
})

// 表单校验规则
const rules: FormRules = {
  name: [{ required: true, message: '请输入电站名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入电站编码', trigger: 'blur' }],
  regionId: [{ required: true, message: '请选择所属区域', trigger: 'change' }],
  type: [{ required: true, message: '请选择电站类型', trigger: 'change' }],
  capacity: [{ required: true, message: '请输入装机容量', trigger: 'blur' }],
  address: [{ required: true, message: '请输入电站地址', trigger: 'blur' }]
}

// 获取类型样式
function getTypeType(type: string) {
  const types: Record<string, string> = {
    solar: 'success',
    wind: 'primary',
    storage: 'warning'
  }
  return types[type] || 'info'
}

// 获取类型文本
function getTypeText(type: string) {
  const texts: Record<string, string> = {
    solar: '光伏电站',
    wind: '风电场',
    storage: '储能站'
  }
  return texts[type] || '其他'
}

// 搜索
function handleSearch() {
  ElMessage.success('搜索完成')
}

// 重置
function handleReset() {
  searchForm.name = ''
  searchForm.regionId = undefined
  searchForm.status = undefined
}

// 分页变化
function handlePageChange() {
  // 加载数据
}

// 新增电站
function handleAdd() {
  isEdit.value = false
  dialogTitle.value = '新增电站'
  resetForm()
  dialogVisible.value = true
}

// 查看电站
function handleView(row: any) {
  ElMessage.info(`查看电站: ${row.name}`)
}

// 编辑电站
function handleEdit(row: any) {
  isEdit.value = true
  dialogTitle.value = '编辑电站'
  Object.assign(form, row)
  dialogVisible.value = true
}

// 删除电站
async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `确定要删除电站 "${row.name}" 吗？`,
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
  form.regionId = undefined
  form.type = ''
  form.capacity = 0
  form.address = ''
  form.longitude = undefined
  form.latitude = undefined
  form.remark = ''
}

onMounted(() => {
  loading.value = false
})
</script>

<style scoped lang="scss">
.station-manage-page {
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

.unit {
  margin-left: 10px;
  color: #909399;
}
</style>
