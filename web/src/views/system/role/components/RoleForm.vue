<template>
  <el-dialog
    v-model="visible"
    :title="roleId ? '编辑角色' : '新增角色'"
    width="600px"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="100px"
    >
      <el-form-item label="角色名称" prop="name">
        <el-input v-model="formData.name" placeholder="请输入角色名称" />
      </el-form-item>
      <el-form-item label="角色编码" prop="code">
        <el-input
          v-model="formData.code"
          placeholder="请输入角色编码"
          :disabled="!!roleId"
        />
      </el-form-item>
      <el-form-item label="描述" prop="description">
        <el-input
          v-model="formData.description"
          type="textarea"
          :rows="3"
          placeholder="请输入描述"
        />
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-radio-group v-model="formData.status">
          <el-radio :label="1">启用</el-radio>
          <el-radio :label="0">禁用</el-radio>
        </el-radio-group>
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
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getRoleDetail, createRole, updateRole } from '@/api/role'
import type { Role } from '@/types'

const props = defineProps<{
  visible: boolean
  roleId: number | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  success: []
}>()

const formRef = ref<FormInstance>()
const loading = ref(false)

const formData = reactive<Partial<Role>>({
  name: '',
  code: '',
  description: '',
  status: 1,
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入角色名称', trigger: 'blur' },
    { min: 2, max: 20, message: '角色名称长度为2-20个字符', trigger: 'blur' },
  ],
  code: [
    { required: true, message: '请输入角色编码', trigger: 'blur' },
    { min: 2, max: 20, message: '角色编码长度为2-20个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z_]+$/, message: '角色编码只能包含字母和下划线', trigger: 'blur' },
  ],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
}

// 监听visible变化
watch(
  () => props.visible,
  (val) => {
    if (val) {
      if (props.roleId) {
        loadRoleDetail()
      } else {
        resetForm()
      }
    }
  }
)

// 加载角色详情
async function loadRoleDetail() {
  if (!props.roleId) return
  try {
    const role = await getRoleDetail(props.roleId)
    Object.assign(formData, role)
  } catch (error) {
    console.error('获取角色详情失败:', error)
  }
}

// 重置表单
function resetForm() {
  formData.name = ''
  formData.code = ''
  formData.description = ''
  formData.status = 1
  formRef.value?.clearValidate()
}

// 关闭弹窗
function handleClose() {
  emit('update:visible', false)
  resetForm()
}

// 提交
async function handleSubmit() {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      if (props.roleId) {
        await updateRole(props.roleId, formData)
        ElMessage.success('更新成功')
      } else {
        await createRole(formData)
        ElMessage.success('创建成功')
      }
      emit('success')
      handleClose()
    } catch (error) {
      console.error('提交失败:', error)
    } finally {
      loading.value = false
    }
  })
}
</script>
