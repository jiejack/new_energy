<template>
  <div class="permission-container">
    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="queryParams" inline>
        <el-form-item label="权限名称">
          <el-input
            v-model="queryParams.keyword"
            placeholder="请输入权限名称"
            clearable
            @keyup.enter="handleQuery"
          />
        </el-form-item>
        <el-form-item label="权限类型">
          <el-select v-model="queryParams.type" placeholder="请选择类型" clearable>
            <el-option label="菜单" value="menu" />
            <el-option label="按钮" value="button" />
            <el-option label="接口" value="api" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryParams.status" placeholder="请选择状态" clearable>
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
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

    <!-- 权限树 -->
    <el-card class="tree-card">
      <template #header>
        <div class="card-header">
          <span>权限树</span>
          <div class="header-actions">
            <el-button type="primary" @click="handleAdd()">
              <el-icon><Plus /></el-icon>
              新增权限
            </el-button>
            <el-button @click="toggleExpandAll">
              {{ isExpandAll ? '折叠' : '展开' }}
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        v-if="refreshTable"
        v-loading="loading"
        :data="permissionList"
        row-key="id"
        border
        :default-expand-all="isExpandAll"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
      >
        <el-table-column prop="name" label="权限名称" min-width="180" />
        <el-table-column prop="code" label="权限编码" min-width="150" />
        <el-table-column label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getPermissionTypeTag(row.type)" size="small">
              {{ getPermissionTypeLabel(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" min-width="150" />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="handleStatusChange(row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleAdd(row)">
              新增
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
    </el-card>

    <!-- 权限表单弹窗 -->
    <el-dialog
      v-model="formVisible"
      :title="formTitle"
      width="600px"
      @close="handleFormClose"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="上级权限" prop="parentId">
          <el-tree-select
            v-model="formData.parentId"
            :data="permissionOptions"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            check-strictly
            clearable
            placeholder="请选择上级权限"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="权限名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入权限名称" />
        </el-form-item>
        <el-form-item label="权限编码" prop="code">
          <el-input v-model="formData.code" placeholder="请输入权限编码" />
        </el-form-item>
        <el-form-item label="权限类型" prop="type">
          <el-select v-model="formData.type" placeholder="请选择权限类型" style="width: 100%">
            <el-option label="菜单" value="menu" />
            <el-option label="按钮" value="button" />
            <el-option label="接口" value="api" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="formData.type === 'menu'" label="路由路径" prop="path">
          <el-input v-model="formData.path" placeholder="请输入路由路径" />
        </el-form-item>
        <el-form-item v-if="formData.type === 'menu'" label="图标" prop="icon">
          <el-input v-model="formData.icon" placeholder="请输入图标名称" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="formData.sort" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :label="1">启用</el-radio>
            <el-radio :label="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="handleFormClose">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getPermissionTree, getPermissionDetail, createPermission, updatePermission, deletePermission, updatePermissionStatus } from '@/api/permission'
import type { Permission } from '@/types'

const loading = ref(false)
const permissionList = ref<Permission[]>([])
const isExpandAll = ref(true)
const refreshTable = ref(true)
const formVisible = ref(false)
const formLoading = ref(false)
const currentPermissionId = ref<number | null>(null)
const formRef = ref<FormInstance>()

const queryParams = reactive({
  keyword: '',
  type: undefined as string | undefined,
  status: undefined as number | undefined,
})

const formData = reactive<Partial<Permission>>({
  parentId: null,
  name: '',
  code: '',
  type: 'menu',
  path: '',
  icon: '',
  sort: 0,
  status: 1,
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入权限名称', trigger: 'blur' },
    { min: 2, max: 20, message: '权限名称长度为2-20个字符', trigger: 'blur' },
  ],
  code: [
    { required: true, message: '请输入权限编码', trigger: 'blur' },
    { min: 2, max: 50, message: '权限编码长度为2-50个字符', trigger: 'blur' },
  ],
  type: [{ required: true, message: '请选择权限类型', trigger: 'change' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
}

const formTitle = computed(() => (currentPermissionId.value ? '编辑权限' : '新增权限'))

// 权限选项（用于选择上级权限）
const permissionOptions = computed(() => {
  const options: Permission[] = [
    {
      id: 0,
      name: '顶级权限',
      code: 'root',
      type: 'menu',
      parentId: null,
      sort: 0,
      status: 1,
      createdAt: '',
      updatedAt: '',
      children: permissionList.value,
    },
  ]
  return options
})

// 获取权限类型标签
function getPermissionTypeLabel(type: string) {
  const map: Record<string, string> = {
    menu: '菜单',
    button: '按钮',
    api: '接口',
  }
  return map[type] || type
}

// 获取权限类型标签颜色
function getPermissionTypeTag(type: string) {
  const map: Record<string, string> = {
    menu: 'primary',
    button: 'success',
    api: 'warning',
  }
  return map[type] || 'info'
}

// 获取权限列表
async function getList() {
  loading.value = true
  try {
    permissionList.value = await getPermissionTree()
  } catch (error) {
    console.error('获取权限列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 搜索
function handleQuery() {
  getList()
}

// 重置
function handleReset() {
  queryParams.keyword = ''
  queryParams.type = undefined
  queryParams.status = undefined
  handleQuery()
}

// 展开/折叠
function toggleExpandAll() {
  refreshTable.value = false
  isExpandAll.value = !isExpandAll.value
  setTimeout(() => {
    refreshTable.value = true
  }, 100)
}

// 新增
function handleAdd(row?: Permission) {
  currentPermissionId.value = null
  formData.parentId = row?.id || null
  formData.name = ''
  formData.code = ''
  formData.type = 'menu'
  formData.path = ''
  formData.icon = ''
  formData.sort = 0
  formData.status = 1
  formVisible.value = true
}

// 编辑
async function handleEdit(row: Permission) {
  currentPermissionId.value = row.id
  try {
    const permission = await getPermissionDetail(row.id)
    Object.assign(formData, permission)
    formVisible.value = true
  } catch (error) {
    console.error('获取权限详情失败:', error)
  }
}

// 状态切换
async function handleStatusChange(row: Permission) {
  try {
    await updatePermissionStatus(row.id, row.status)
    ElMessage.success('状态更新成功')
  } catch (error) {
    row.status = row.status === 1 ? 0 : 1
  }
}

// 删除
async function handleDelete(row: Permission) {
  try {
    await ElMessageBox.confirm(`确认要删除权限"${row.name}"吗？`, '提示', {
      type: 'warning',
    })
    await deletePermission(row.id)
    ElMessage.success('删除成功')
    getList()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

// 关闭表单弹窗
function handleFormClose() {
  formVisible.value = false
  currentPermissionId.value = null
  formRef.value?.clearValidate()
}

// 提交表单
async function handleSubmit() {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    formLoading.value = true
    try {
      if (currentPermissionId.value) {
        await updatePermission(currentPermissionId.value, formData)
        ElMessage.success('更新成功')
      } else {
        await createPermission(formData)
        ElMessage.success('创建成功')
      }
      handleFormClose()
      getList()
    } catch (error) {
      console.error('提交失败:', error)
    } finally {
      formLoading.value = false
    }
  })
}

onMounted(() => {
  getList()
})
</script>

<style scoped lang="scss">
.permission-container {
  padding: 20px;

  .search-card {
    margin-bottom: 20px;
  }

  .tree-card {
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
}
</style>
