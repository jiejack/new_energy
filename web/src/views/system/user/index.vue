<template>
  <div class="user-container">
    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-form :model="queryParams" inline>
        <el-form-item label="用户名">
          <el-input
            v-model="queryParams.keyword"
            placeholder="请输入用户名/昵称/手机号"
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
          <span>用户列表</span>
          <div class="header-actions">
            <el-button type="primary" @click="handleAdd">
              <el-icon><Plus /></el-icon>
              新增用户
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
        :data="userList"
        border
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="nickname" label="昵称" min-width="120" />
        <el-table-column prop="email" label="邮箱" min-width="180" />
        <el-table-column prop="phone" label="手机号" min-width="130" />
        <el-table-column label="角色" min-width="150">
          <template #default="{ row }">
            <el-tag
              v-for="role in row.roles"
              :key="role"
              type="info"
              size="small"
              style="margin-right: 4px"
            >
              {{ role }}
            </el-tag>
          </template>
        </el-table-column>
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
        <el-table-column label="操作" width="240" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button type="primary" link size="small" @click="handleAssignRole(row)">
              分配角色
            </el-button>
            <el-button type="warning" link size="small" @click="handleResetPassword(row)">
              重置密码
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

    <!-- 用户表单弹窗 -->
    <UserForm
      v-model:visible="formVisible"
      :user-id="currentUserId"
      @success="handleQuery"
    />

    <!-- 角色分配弹窗 -->
    <AssignRole
      v-model:visible="assignRoleVisible"
      :user-id="currentUserId"
      @success="handleQuery"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUserList, deleteUser, batchDeleteUsers, updateUserStatus, resetPassword } from '@/api/user'
import type { UserInfo, PageQuery } from '@/types'
import UserForm from './components/UserForm.vue'
import AssignRole from './components/AssignRole.vue'

const loading = ref(false)
const userList = ref<UserInfo[]>([])
const total = ref(0)
const selectedIds = ref<number[]>([])
const formVisible = ref(false)
const assignRoleVisible = ref(false)
const currentUserId = ref<number | null>(null)

const queryParams = reactive<PageQuery & { keyword?: string; status?: number }>({
  page: 1,
  pageSize: 10,
  keyword: '',
  status: undefined,
})

// 获取用户列表
async function getList() {
  loading.value = true
  try {
    const result = await getUserList(queryParams)
    userList.value = result.list
    total.value = result.total
  } catch (error) {
    console.error('获取用户列表失败:', error)
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
function handleSelectionChange(selection: UserInfo[]) {
  selectedIds.value = selection.map((item) => item.id)
}

// 新增
function handleAdd() {
  currentUserId.value = null
  formVisible.value = true
}

// 编辑
function handleEdit(row: UserInfo) {
  currentUserId.value = row.id
  formVisible.value = true
}

// 分配角色
function handleAssignRole(row: UserInfo) {
  currentUserId.value = row.id
  assignRoleVisible.value = true
}

// 状态切换
async function handleStatusChange(row: UserInfo) {
  try {
    await updateUserStatus(row.id, row.status)
    ElMessage.success('状态更新成功')
  } catch (error) {
    row.status = row.status === 1 ? 0 : 1
  }
}

// 重置密码
async function handleResetPassword(row: UserInfo) {
  try {
    await ElMessageBox.confirm(`确认要重置用户"${row.username}"的密码吗？`, '提示', {
      type: 'warning',
    })
    await resetPassword(row.id)
    ElMessage.success('密码重置成功，默认密码为：123456')
  } catch (error) {
    console.error('重置密码失败:', error)
  }
}

// 删除
async function handleDelete(row: UserInfo) {
  try {
    await ElMessageBox.confirm(`确认要删除用户"${row.username}"吗？`, '提示', {
      type: 'warning',
    })
    await deleteUser(row.id)
    ElMessage.success('删除成功')
    getList()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

// 批量删除
async function handleBatchDelete() {
  try {
    await ElMessageBox.confirm(`确认要删除选中的 ${selectedIds.value.length} 个用户吗？`, '提示', {
      type: 'warning',
    })
    await batchDeleteUsers(selectedIds.value)
    ElMessage.success('批量删除成功')
    getList()
  } catch (error) {
    console.error('批量删除失败:', error)
  }
}

onMounted(() => {
  getList()
})
</script>

<style scoped lang="scss">
.user-container {
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
