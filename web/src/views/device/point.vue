<template>
  <div class="point-manage-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>采集点管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增采集点
          </el-button>
        </div>
      </template>

      <!-- 搜索栏 -->
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="采集点名称">
          <el-input v-model="searchForm.name" placeholder="请输入采集点名称" clearable />
        </el-form-item>
        <el-form-item label="所属设备">
          <el-select v-model="searchForm.deviceId" placeholder="请选择设备" clearable style="width: 180px">
            <el-option label="逆变器 #01" :value="1" />
            <el-option label="逆变器 #02" :value="2" />
            <el-option label="汇流箱 #01" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="数据类型">
          <el-select v-model="searchForm.dataType" placeholder="请选择类型" clearable style="width: 150px">
            <el-option label="遥测" value="telemetry" />
            <el-option label="遥信" value="telesignal" />
            <el-option label="遥脉" value="telepulse" />
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
      <el-table :data="pointList" v-loading="loading" stripe>
        <el-table-column prop="name" label="采集点名称" min-width="160" />
        <el-table-column prop="code" label="采集点编码" width="130" />
        <el-table-column prop="deviceName" label="所属设备" min-width="150" />
        <el-table-column prop="dataType" label="数据类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getDataTypeType(row.dataType)">{{ getDataTypeText(row.dataType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="dataFormat" label="数据格式" width="100">
          <template #default="{ row }">
            {{ getDataFormatText(row.dataFormat) }}
          </template>
        </el-table-column>
        <el-table-column prop="unit" label="单位" width="80" align="center" />
        <el-table-column prop="currentValue" label="当前值" width="100" align="right">
          <template #default="{ row }">
            <span :class="getValueClass(row)">{{ row.currentValue }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="updateTime" label="更新时间" width="180" />
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
            <el-form-item label="采集点名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入采集点名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="采集点编码" prop="code">
              <el-input v-model="form.code" placeholder="请输入采集点编码" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="所属设备" prop="deviceId">
              <el-select v-model="form.deviceId" placeholder="请选择设备" style="width: 100%">
                <el-option label="逆变器 #01" :value="1" />
                <el-option label="逆变器 #02" :value="2" />
                <el-option label="汇流箱 #01" :value="3" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="数据类型" prop="dataType">
              <el-select v-model="form.dataType" placeholder="请选择类型" style="width: 100%">
                <el-option label="遥测" value="telemetry" />
                <el-option label="遥信" value="telesignal" />
                <el-option label="遥脉" value="telepulse" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="数据格式" prop="dataFormat">
              <el-select v-model="form.dataFormat" placeholder="请选择格式" style="width: 100%">
                <el-option label="浮点数" value="float" />
                <el-option label="整数" value="int" />
                <el-option label="布尔值" value="bool" />
                <el-option label="字符串" value="string" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="单位" prop="unit">
              <el-input v-model="form.unit" placeholder="如: kW, V, A" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="量程上限" prop="maxValue">
              <el-input-number v-model="form.maxValue" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="量程下限" prop="minValue">
              <el-input-number v-model="form.minValue" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
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
  deviceId: undefined as number | undefined,
  dataType: ''
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 6
})

// 采集点列表数据
const pointList = ref([
  {
    id: 1,
    name: '直流电压',
    code: 'DC_VOLTAGE',
    deviceName: '逆变器 #01',
    dataType: 'telemetry',
    dataFormat: 'float',
    unit: 'V',
    currentValue: '650.5',
    updateTime: '2024-01-15 14:30:00'
  },
  {
    id: 2,
    name: '直流电流',
    code: 'DC_CURRENT',
    deviceName: '逆变器 #01',
    dataType: 'telemetry',
    dataFormat: 'float',
    unit: 'A',
    currentValue: '125.8',
    updateTime: '2024-01-15 14:30:00'
  },
  {
    id: 3,
    name: '交流功率',
    code: 'AC_POWER',
    deviceName: '逆变器 #01',
    dataType: 'telemetry',
    dataFormat: 'float',
    unit: 'kW',
    currentValue: '78.5',
    updateTime: '2024-01-15 14:30:00'
  },
  {
    id: 4,
    name: '运行状态',
    code: 'RUN_STATUS',
    deviceName: '逆变器 #01',
    dataType: 'telesignal',
    dataFormat: 'int',
    unit: '-',
    currentValue: '1',
    updateTime: '2024-01-15 14:30:00'
  },
  {
    id: 5,
    name: '今日发电量',
    code: 'DAILY_ENERGY',
    deviceName: '逆变器 #01',
    dataType: 'telepulse',
    dataFormat: 'float',
    unit: 'kWh',
    currentValue: '156.8',
    updateTime: '2024-01-15 14:30:00'
  },
  {
    id: 6,
    name: '总发电量',
    code: 'TOTAL_ENERGY',
    deviceName: '逆变器 #01',
    dataType: 'telepulse',
    dataFormat: 'float',
    unit: 'MWh',
    currentValue: '1256.5',
    updateTime: '2024-01-15 14:30:00'
  }
])

// 对话框相关
const dialogVisible = ref(false)
const dialogTitle = ref('新增采集点')
const formRef = ref<FormInstance>()
const isEdit = ref(false)

// 表单数据
const form = reactive({
  id: undefined as number | undefined,
  name: '',
  code: '',
  deviceId: undefined as number | undefined,
  dataType: '',
  dataFormat: '',
  unit: '',
  maxValue: undefined as number | undefined,
  minValue: undefined as number | undefined,
  remark: ''
})

// 表单校验规则
const rules: FormRules = {
  name: [{ required: true, message: '请输入采集点名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入采集点编码', trigger: 'blur' }],
  deviceId: [{ required: true, message: '请选择所属设备', trigger: 'change' }],
  dataType: [{ required: true, message: '请选择数据类型', trigger: 'change' }],
  dataFormat: [{ required: true, message: '请选择数据格式', trigger: 'change' }]
}

// 获取数据类型样式
function getDataTypeType(type: string) {
  const types: Record<string, string> = {
    telemetry: 'success',
    telesignal: 'primary',
    telepulse: 'warning'
  }
  return types[type] || 'info'
}

// 获取数据类型文本
function getDataTypeText(type: string) {
  const texts: Record<string, string> = {
    telemetry: '遥测',
    telesignal: '遥信',
    telepulse: '遥脉'
  }
  return texts[type] || '其他'
}

// 获取数据格式文本
function getDataFormatText(format: string) {
  const texts: Record<string, string> = {
    float: '浮点数',
    int: '整数',
    bool: '布尔值',
    string: '字符串'
  }
  return texts[format] || format
}

// 获取值样式
function getValueClass(row: any) {
  if (row.dataType === 'telesignal') {
    return row.currentValue === '1' ? 'value-on' : 'value-off'
  }
  return ''
}

// 搜索
function handleSearch() {
  ElMessage.success('搜索完成')
}

// 重置
function handleReset() {
  searchForm.name = ''
  searchForm.deviceId = undefined
  searchForm.dataType = ''
}

// 分页变化
function handlePageChange() {
  // 加载数据
}

// 新增采集点
function handleAdd() {
  isEdit.value = false
  dialogTitle.value = '新增采集点'
  resetForm()
  dialogVisible.value = true
}

// 查看采集点
function handleView(row: any) {
  ElMessage.info(`查看采集点: ${row.name}`)
}

// 编辑采集点
function handleEdit(row: any) {
  isEdit.value = true
  dialogTitle.value = '编辑采集点'
  Object.assign(form, row)
  dialogVisible.value = true
}

// 删除采集点
async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `确定要删除采集点 "${row.name}" 吗？`,
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
  form.deviceId = undefined
  form.dataType = ''
  form.dataFormat = ''
  form.unit = ''
  form.maxValue = undefined
  form.minValue = undefined
  form.remark = ''
}

onMounted(() => {
  loading.value = false
})
</script>

<style scoped lang="scss">
.point-manage-page {
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

.value-on {
  color: #67c23a;
  font-weight: bold;
}

.value-off {
  color: #909399;
}
</style>
