<template>
  <el-dialog
    :model-value="visible"
    title="分配角色"
    width="500px"
    @update:model-value="emit('update:visible', $event)"
    @close="handleClose"
  >
    <el-form ref="formRef" label-width="80px">
      <el-form-item label="用户角色">
        <el-checkbox-group v-model="selectedRoleIds">
          <el-checkbox
            v-for="role in roleList"
            :key="role.id"
            :label="role.id"
            :disabled="role.status === 0"
          >
            {{ role.name }}
          </el-checkbox>
        </el-checkbox-group>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">
        确定
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getAllRoles } from '@/api/role'
import { getUserRoles, assignRoles } from '@/api/user'
import type { Role } from '@/types'

const props = defineProps<{
  visible: boolean
  userId: number | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  success: []
}>()

const loading = ref(false)
const roleList = ref<Role[]>([])
const selectedRoleIds = ref<number[]>([])

// 监听visible变化
watch(
  () => props.visible,
  (val) => {
    if (val && props.userId) {
      loadRoles()
      loadUserRoles()
    }
  }
)

// 加载角色列表
async function loadRoles() {
  try {
    roleList.value = await getAllRoles()
  } catch (error) {
    console.error('获取角色列表失败:', error)
  }
}

// 加载用户角色
async function loadUserRoles() {
  if (!props.userId) return
  try {
    selectedRoleIds.value = await getUserRoles(props.userId)
  } catch (error) {
    console.error('获取用户角色失败:', error)
  }
}

// 关闭弹窗
function handleClose() {
  emit('update:visible', false)
  selectedRoleIds.value = []
}

// 提交
async function handleSubmit() {
  if (!props.userId) return

  loading.value = true
  try {
    await assignRoles(props.userId, selectedRoleIds.value)
    ElMessage.success('分配角色成功')
    emit('success')
    handleClose()
  } catch (error) {
    console.error('分配角色失败:', error)
  } finally {
    loading.value = false
  }
}
</script>
