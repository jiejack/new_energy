<template>
  <div v-loading="loading" class="permission-tree">
    <el-tree
      ref="treeRef"
      :data="permissionTree"
      :props="treeProps"
      :default-checked-keys="checkedKeys"
      show-checkbox
      node-key="id"
      :check-strictly="false"
      :expand-on-click-node="false"
      default-expand-all
    >
      <template #default="{ node, data }">
        <span class="custom-tree-node">
          <el-icon v-if="data.icon" style="margin-right: 4px">
            <component :is="data.icon" />
          </el-icon>
          <span>{{ node.label }}</span>
          <el-tag
            :type="getPermissionTypeTag(data.type)"
            size="small"
            style="margin-left: 8px"
          >
            {{ getPermissionTypeLabel(data.type) }}
          </el-tag>
        </span>
      </template>
    </el-tree>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { getPermissionTree } from '@/api/permission'
import { getRolePermissions } from '@/api/role'
import type { Permission } from '@/types'
import type { ElTree } from 'element-plus'

const props = defineProps<{
  roleId: number | null
}>()

const loading = ref(false)
const treeRef = ref<InstanceType<typeof ElTree>>()
const permissionTree = ref<Permission[]>([])
const checkedKeys = ref<number[]>([])

const treeProps = {
  label: 'name',
  children: 'children',
}

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

// 加载权限树
async function loadPermissionTree() {
  loading.value = true
  try {
    permissionTree.value = await getPermissionTree()
  } catch (error) {
    console.error('获取权限树失败:', error)
  } finally {
    loading.value = false
  }
}

// 加载角色权限
async function loadRolePermissions() {
  if (!props.roleId) return
  try {
    checkedKeys.value = await getRolePermissions(props.roleId)
  } catch (error) {
    console.error('获取角色权限失败:', error)
  }
}

// 获取选中的权限ID
function getCheckedKeys(): number[] {
  return treeRef.value?.getCheckedKeys(false) as number[]
}

// 监听roleId变化
watch(
  () => props.roleId,
  (val) => {
    if (val) {
      loadRolePermissions()
    } else {
      checkedKeys.value = []
    }
  }
)

onMounted(() => {
  loadPermissionTree()
})

defineExpose({
  getCheckedKeys,
})
</script>

<style scoped lang="scss">
.permission-tree {
  max-height: 400px;
  overflow-y: auto;

  .custom-tree-node {
    display: flex;
    align-items: center;
    flex: 1;
    font-size: 14px;
  }
}
</style>
