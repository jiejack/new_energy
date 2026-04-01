<template>
  <div class="region-manage-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>区域管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增区域
          </el-button>
        </div>
      </template>

      <el-table :data="regionList" v-loading="loading" row-key="id" default-expand-all>
        <el-table-column prop="name" label="区域名称" min-width="200">
          <template #default="{ row }">
            <el-icon><Location /></el-icon>
            <span style="margin-left: 8px">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="code" label="区域编码" width="150" />
        <el-table-column prop="level" label="层级" width="100">
          <template #default="{ row }">
            <el-tag :type="getLevelType(row.level)">{{ getLevelText(row.level) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="stationCount" label="电站数量" width="100" align="center" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="createTime" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleAddChild(row)">添加子区域</el-button>
            <el-button type="primary" link @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="上级区域" v-if="form.parentId">
          <el-input v-model="form.parentName" disabled />
        </el-form-item>
        <el-form-item label="区域名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入区域名称" />
        </el-form-item>
        <el-form-item label="区域编码" prop="code">
          <el-input v-model="form.code" placeholder="请输入区域编码" />
        </el-form-item>
        <el-form-item label="区域描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入区域描述"
          />
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
import { Plus, Location } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

// 加载状态
const loading = ref(false)

// 区域列表数据
const regionList = ref([
  {
    id: 1,
    name: '华北地区',
    code: 'HB',
    level: 1,
    stationCount: 5,
    description: '包括北京、天津、河北、山西、内蒙古',
    createTime: '2024-01-01 10:00:00',
    children: [
      {
        id: 11,
        name: '北京市',
        code: 'BJ',
        level: 2,
        stationCount: 2,
        description: '北京市区域',
        createTime: '2024-01-01 10:00:00',
        parentId: 1
      },
      {
        id: 12,
        name: '河北省',
        code: 'HE',
        level: 2,
        stationCount: 3,
        description: '河北省区域',
        createTime: '2024-01-01 10:00:00',
        parentId: 1
      }
    ]
  },
  {
    id: 2,
    name: '华东地区',
    code: 'HD',
    level: 1,
    stationCount: 8,
    description: '包括上海、江苏、浙江、安徽、福建、江西、山东',
    createTime: '2024-01-01 10:00:00',
    children: [
      {
        id: 21,
        name: '上海市',
        code: 'SH',
        level: 2,
        stationCount: 3,
        description: '上海市区域',
        createTime: '2024-01-01 10:00:00',
        parentId: 2
      },
      {
        id: 22,
        name: '江苏省',
        code: 'JS',
        level: 2,
        stationCount: 5,
        description: '江苏省区域',
        createTime: '2024-01-01 10:00:00',
        parentId: 2
      }
    ]
  }
])

// 对话框相关
const dialogVisible = ref(false)
const dialogTitle = ref('新增区域')
const formRef = ref<FormInstance>()
const isEdit = ref(false)

// 表单数据
const form = reactive({
  id: undefined as number | undefined,
  parentId: undefined as number | undefined,
  parentName: '',
  name: '',
  code: '',
  description: ''
})

// 表单校验规则
const rules: FormRules = {
  name: [{ required: true, message: '请输入区域名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入区域编码', trigger: 'blur' }]
}

// 获取层级类型
function getLevelType(level: number) {
  const types = ['', 'success', 'warning', 'info']
  return types[level] || 'info'
}

// 获取层级文本
function getLevelText(level: number) {
  const texts = ['', '一级', '二级', '三级']
  return texts[level] || '其他'
}

// 新增区域
function handleAdd() {
  isEdit.value = false
  dialogTitle.value = '新增区域'
  resetForm()
  dialogVisible.value = true
}

// 添加子区域
function handleAddChild(row: any) {
  isEdit.value = false
  dialogTitle.value = '添加子区域'
  resetForm()
  form.parentId = row.id
  form.parentName = row.name
  dialogVisible.value = true
}

// 编辑区域
function handleEdit(row: any) {
  isEdit.value = true
  dialogTitle.value = '编辑区域'
  Object.assign(form, row)
  dialogVisible.value = true
}

// 删除区域
async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `确定要删除区域 "${row.name}" 吗？删除后无法恢复。`,
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
  form.parentId = undefined
  form.parentName = ''
  form.name = ''
  form.code = ''
  form.description = ''
}

onMounted(() => {
  loading.value = false
})
</script>

<style scoped lang="scss">
.region-manage-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
