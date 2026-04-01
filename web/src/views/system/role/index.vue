<template>
  <div class="role-container">
    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="queryParams" inline>
        <el-form-item label="角色名称">
          <el-input
            v-model="queryParams.keyword"
            placeholder="请输入角色名称"
            clearable
            @keyup.enter="handleQuery"
          />
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

    <!-- 表格 -->
    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>角色列表</span>
          <div class="header-actions">
            <el-button type="primary" @click="handleAdd">
              <el-icon><Plus /></el-icon>
              新增角色
            </el-button>
            <el-button
              type="danger"
              :disabled="selectedIds.length === 0"
              @click="handleBatchDelete"
            >
              <el-icon><Delete /></el-icon>
              批量删除
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="roleList"
        border
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column prop="name" label="角色名称" min-width="120" />
        <el-table-column prop="code" label="角色编码" min-width="120" />
        <el-table-column prop="description" label="描述" min-width="200" />
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
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button type="primary" link size="small" @click="handleAssignPermission(row)">
              分配权限
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">
              删除
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

    <!-- 角色表单弹窗 -->
    <RoleForm
      v-model:visible="formVisible"
      :role-id="currentRoleId"
      @success="handleQuery"
    />

    <!-- 权限分配弹窗 -->
    <el-dialog
      v-model="permissionVisible"
      title="分配权限"
      width="500px"
      @close="handlePermissionClose"
    >
      <PermissionTree ref="permissionTreeRef" :role-id="currentRoleId" />
      <template #footer>
        <el-button @click="handlePermissionClose">取消</el-button>
        <el-button type="primary" :loading="permissionLoading" @click="handlePermissionSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRoleList, deleteRole, batchDeleteRoles, updateRoleStatus, assignPermissions } from '@/api/role'
import type { Role, PageQuery } from '@/types'
import RoleForm from './components/RoleForm.vue'
import PermissionTree from './components/PermissionTree.vue'

const loading = ref(false)
const roleList = ref<Role[]>([])
const total = ref(0)
const selectedIds = ref<number[]>([])
const formVisible = ref(false)
const permissionVisible = ref(false)
const currentRoleId = ref<number | null>(null)
const permissionLoading = ref(false)
const permissionTreeRef = ref<InstanceType<typeof PermissionTree>>()

const queryParams = reactive<PageQuery & { keyword?: string; status?: number }>({
  page: 1,
  pageSize: 10,
  keyword: '',
  status: undefined,
})

// 获取角色列表
async function getList() {
  loading.value = true
  try {
    const result = await getRoleList(queryParams)
    roleList.value = result.list
    total.value = result.total
  } catch (error) {
    console.error('获取角色列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 搜索
function handleQuery() {
  queryParams.page = 1
  getList()
}

// 重置
function handleReset() {
  queryParams.keyword = ''
  queryParams.status = undefined
  handleQuery()
}

// 选择变化
function handleSelectionChange(selection: Role[]) {
  selectedIds.value = selection.map((item) => item.id)
}

// 新增
function handleAdd() {
  currentRoleId.value = null
  formVisible.value = true
}

// 编辑
function handleEdit(row: Role) {
  currentRoleId.value = row.id
  formVisible.value = true
}

// 分配权限
function handleAssignPermission(row: Role) {
  currentRoleId.value = row.id
  permissionVisible.value = true
}

// 状态切换
async function handleStatusChange(row: Role) {
  try {
    await updateRoleStatus(row.id, row.status)
    ElMessage.success('状态更新成功')
  } catch (error) {
    row.status = row.status === 1 ? 0 : 1
  }
}

// 删除
async function handleDelete(row: Role) {
  try {
    await ElMessageBox.confirm(`确认要删除角色"${row.name}"吗？`, '提示', {
      type: 'warning',
    })
    await deleteRole(row.id)
    ElMessage.success('删除成功')
    getList()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

// 批量删除
async function handleBatchDelete() {
  try {
    await ElMessageBox.confirm(`确认要删除选中的 ${selectedIds.value.length} 个角色吗？`, '提示', {
      type: 'warning',
    })
    await batchDeleteRoles(selectedIds.value)
    ElMessage.success('批量删除成功')
    getList()
  } catch (error) {
    console.error('批量删除失败:', error)
  }
}

// 关闭权限弹窗
function handlePermissionClose() {
  permissionVisible.value = false
  currentRoleId.value = null
}

// 提交权限分配
async function handlePermissionSubmit() {
  if (!currentRoleId.value || !permissionTreeRef.value) return

  const permissionIds = permissionTreeRef.value.getCheckedKeys()
  permissionLoading.value = true
  try {
    await assignPermissions(currentRoleId.value, permissionIds)
    ElMessage.success('分配权限成功')
    handlePermissionClose()
  } catch (error) {
    console.error('分配权限失败:', error)
  } finally {
    permissionLoading.value = false
  }
}

onMounted(() => {
  getList()
})
</script>

<style scoped lang="scss">
.role-container {
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
