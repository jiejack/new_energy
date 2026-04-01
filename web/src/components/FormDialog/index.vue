<template>
  <el-dialog
    v-model="visible"
    :title="dialogTitle"
    :width="width"
    :top="top"
    :modal="modal"
    :modal-class="modalClass"
    :close-on-click-modal="closeOnClickModal"
    :close-on-press-escape="closeOnPressEscape"
    :show-close="showClose"
    :draggable="draggable"
    :destroy-on-close="destroyOnClose"
    @close="handleClose"
    @open="handleOpen"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      :label-width="labelWidth"
      :label-position="labelPosition"
      :size="size"
      :disabled="disabled"
    >
      <el-row :gutter="gutter">
        <template v-for="field in fields" :key="field.prop">
          <el-col
            :span="field.span || 24"
            v-if="!field.hidden && shouldShowField(field)"
          >
            <el-form-item
              :label="field.label"
              :prop="field.prop"
              :required="field.required"
            >
              <!-- 自定义插槽 -->
              <template v-if="$slots[field.prop]">
                <slot
                  :name="field.prop"
                  :form="formData"
                  :field="field"
                ></slot>
              </template>

              <!-- 输入框 -->
              <template v-else-if="field.type === 'input'">
                <el-input
                  v-model="formData[field.prop]"
                  :placeholder="field.placeholder || `请输入${field.label}`"
                  :maxlength="field.maxlength"
                  :show-word-limit="field.showWordLimit"
                  :clearable="field.clearable !== false"
                  :disabled="field.disabled"
                  :readonly="field.readonly"
                  :prefix-icon="field.prefixIcon"
                  :suffix-icon="field.suffixIcon"
                  :type="field.inputType || 'text'"
                  :rows="field.rows"
                  :autosize="field.autosize"
                />
              </template>

              <!-- 数字输入框 -->
              <template v-else-if="field.type === 'number'">
                <el-input-number
                  v-model="formData[field.prop]"
                  :placeholder="field.placeholder || `请输入${field.label}`"
                  :min="field.min"
                  :max="field.max"
                  :step="field.step || 1"
                  :precision="field.precision"
                  :disabled="field.disabled"
                  :controls="field.controls !== false"
                  :controls-position="field.controlsPosition"
                  style="width: 100%"
                />
              </template>

              <!-- 选择器 -->
              <template v-else-if="field.type === 'select'">
                <el-select
                  v-model="formData[field.prop]"
                  :placeholder="field.placeholder || `请选择${field.label}`"
                  :clearable="field.clearable !== false"
                  :disabled="field.disabled"
                  :multiple="field.multiple"
                  :filterable="field.filterable"
                  :remote="field.remote"
                  :remote-method="field.remoteMethod"
                  :loading="field.loading"
                  style="width: 100%"
                >
                  <el-option
                    v-for="option in field.options || []"
                    :key="option.value"
                    :label="option.label"
                    :value="option.value"
                    :disabled="option.disabled"
                  />
                </el-select>
              </template>

              <!-- 级联选择器 -->
              <template v-else-if="field.type === 'cascader'">
                <el-cascader
                  v-model="formData[field.prop]"
                  :options="field.options || []"
                  :placeholder="field.placeholder || `请选择${field.label}`"
                  :clearable="field.clearable !== false"
                  :disabled="field.disabled"
                  :props="field.cascaderProps"
                  :show-all-levels="field.showAllLevels !== false"
                  :collapse-tags="field.collapseTags"
                  :filterable="field.filterable"
                  style="width: 100%"
                />
              </template>

              <!-- 日期选择器 -->
              <template v-else-if="field.type === 'date'">
                <el-date-picker
                  v-model="formData[field.prop]"
                  :type="field.dateType || 'date'"
                  :placeholder="field.placeholder || `请选择${field.label}`"
                  :clearable="field.clearable !== false"
                  :disabled="field.disabled"
                  :format="field.format"
                  :value-format="field.valueFormat"
                  :start-placeholder="field.startPlaceholder"
                  :end-placeholder="field.endPlaceholder"
                  :picker-options="field.pickerOptions"
                  style="width: 100%"
                />
              </template>

              <!-- 时间选择器 -->
              <template v-else-if="field.type === 'time'">
                <el-time-picker
                  v-model="formData[field.prop]"
                  :placeholder="field.placeholder || `请选择${field.label}`"
                  :clearable="field.clearable !== false"
                  :disabled="field.disabled"
                  :format="field.format"
                  :value-format="field.valueFormat"
                  :is-range="field.isRange"
                  :start-placeholder="field.startPlaceholder"
                  :end-placeholder="field.endPlaceholder"
                  style="width: 100%"
                />
              </template>

              <!-- 开关 -->
              <template v-else-if="field.type === 'switch'">
                <el-switch
                  v-model="formData[field.prop]"
                  :disabled="field.disabled"
                  :active-text="field.activeText"
                  :inactive-text="field.inactiveText"
                  :active-value="field.activeValue ?? true"
                  :inactive-value="field.inactiveValue ?? false"
                />
              </template>

              <!-- 单选框组 -->
              <template v-else-if="field.type === 'radio'">
                <el-radio-group
                  v-model="formData[field.prop]"
                  :disabled="field.disabled"
                >
                  <component
                    :is="field.radioButton ? 'el-radio-button' : 'el-radio'"
                    v-for="option in field.options || []"
                    :key="option.value"
                    :label="option.value"
                    :disabled="option.disabled"
                  >
                    {{ option.label }}
                  </component>
                </el-radio-group>
              </template>

              <!-- 复选框组 -->
              <template v-else-if="field.type === 'checkbox'">
                <el-checkbox-group
                  v-model="formData[field.prop]"
                  :disabled="field.disabled"
                >
                  <component
                    :is="field.checkboxButton ? 'el-checkbox-button' : 'el-checkbox'"
                    v-for="option in field.options || []"
                    :key="option.value"
                    :label="option.value"
                    :disabled="option.disabled"
                  >
                    {{ option.label }}
                  </component>
                </el-checkbox-group>
              </template>

              <!-- 文本域 -->
              <template v-else-if="field.type === 'textarea'">
                <el-input
                  v-model="formData[field.prop]"
                  type="textarea"
                  :placeholder="field.placeholder || `请输入${field.label}`"
                  :maxlength="field.maxlength"
                  :show-word-limit="field.showWordLimit"
                  :disabled="field.disabled"
                  :readonly="field.readonly"
                  :rows="field.rows || 4"
                  :autosize="field.autosize"
                />
              </template>

              <!-- 上传 -->
              <template v-else-if="field.type === 'upload'">
                <el-upload
                  :action="field.action"
                  :headers="field.headers"
                  :multiple="field.multiple"
                  :data="field.data"
                  :name="field.name || 'file'"
                  :accept="field.accept"
                  :limit="field.limit"
                  :disabled="field.disabled"
                  :show-file-list="field.showFileList !== false"
                  :drag="field.drag"
                  :list-type="field.listType"
                  :auto-upload="field.autoUpload !== false"
                  :file-list="formData[field.prop + 'List'] || []"
                  :on-preview="field.onPreview"
                  :on-remove="(file: any, fileList: any[]) => handleUploadRemove(field.prop, file, fileList)"
                  :on-success="(response: any, file: any, fileList: any[]) => handleUploadSuccess(field.prop, response, file, fileList)"
                  :on-error="field.onError"
                  :on-progress="field.onProgress"
                  :on-change="field.onChange"
                  :before-upload="field.beforeUpload"
                  :before-remove="field.beforeRemove"
                  :http-request="field.httpRequest"
                >
                  <template v-if="field.listType === 'picture-card'">
                    <el-icon><Plus /></el-icon>
                  </template>
                  <template v-else>
                    <el-button type="primary">
                      {{ field.buttonText || '点击上传' }}
                    </el-button>
                  </template>
                  <template #tip v-if="field.tip">
                    <div class="el-upload__tip">{{ field.tip }}</div>
                  </template>
                </el-upload>
              </template>

              <!-- 富文本编辑器 -->
              <template v-else-if="field.type === 'editor'">
                <el-input
                  v-model="formData[field.prop]"
                  type="textarea"
                  :placeholder="field.placeholder || `请输入${field.label}`"
                  :rows="field.rows || 6"
                />
              </template>

              <!-- 滑块 -->
              <template v-else-if="field.type === 'slider'">
                <el-slider
                  v-model="formData[field.prop]"
                  :min="field.min || 0"
                  :max="field.max || 100"
                  :step="field.step || 1"
                  :disabled="field.disabled"
                  :show-input="field.showInput"
                  :show-stops="field.showStops"
                  :range="field.range"
                  :marks="field.marks"
                />
              </template>

              <!-- 评分 -->
              <template v-else-if="field.type === 'rate'">
                <el-rate
                  v-model="formData[field.prop]"
                  :max="field.max || 5"
                  :disabled="field.disabled"
                  :allow-half="field.allowHalf"
                  :show-text="field.showText"
                  :show-score="field.showScore"
                  :texts="field.texts"
                  :colors="field.colors"
                />
              </template>

              <!-- 颜色选择器 -->
              <template v-else-if="field.type === 'color'">
                <el-color-picker
                  v-model="formData[field.prop]"
                  :disabled="field.disabled"
                  :show-alpha="field.showAlpha"
                  :color-format="field.colorFormat"
                  :predefine="field.predefine"
                />
              </template>

              <!-- 默认文本显示 -->
              <template v-else-if="field.type === 'text'">
                <span>{{ formData[field.prop] || field.defaultValue || '-' }}</span>
              </template>

              <!-- 默认输入框 -->
              <template v-else>
                <el-input
                  v-model="formData[field.prop]"
                  :placeholder="field.placeholder || `请输入${field.label}`"
                  :maxlength="field.maxlength"
                  :show-word-limit="field.showWordLimit"
                  :clearable="field.clearable !== false"
                  :disabled="field.disabled"
                  :readonly="field.readonly"
                />
              </template>
            </el-form-item>
          </el-col>
        </template>
      </el-row>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleCancel">{{ cancelText }}</el-button>
        <el-button
          type="primary"
          @click="handleSubmit"
          :loading="submitLoading"
        >
          {{ confirmText }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'

export interface Field {
  prop: string
  label: string
  type?: string
  span?: number
  required?: boolean
  hidden?: boolean
  disabled?: boolean
  readonly?: boolean
  placeholder?: string
  defaultValue?: any
  options?: Array<{ label: string; value: any; disabled?: boolean }>
  rules?: any[]
  // 条件显示
  show?: (form: any) => boolean
  // 各种类型的属性
  [key: string]: any
}

export interface Props {
  modelValue: boolean
  title?: string
  mode?: 'add' | 'edit' | 'view'
  fields: Field[]
  data?: Record<string, any>
  rules?: FormRules
  width?: string | number
  top?: string
  modal?: boolean
  modalClass?: string
  closeOnClickModal?: boolean
  closeOnPressEscape?: boolean
  showClose?: boolean
  draggable?: boolean
  destroyOnClose?: boolean
  labelWidth?: string | number
  labelPosition?: 'left' | 'right' | 'top'
  size?: 'large' | 'default' | 'small'
  disabled?: boolean
  gutter?: number
  confirmText?: string
  cancelText?: string
  submitLoading?: boolean
  beforeSubmit?: (data: any) => boolean | Promise<boolean>
}

const props = withDefaults(defineProps<Props>(), {
  title: '',
  mode: 'add',
  data: () => ({}),
  rules: () => ({}),
  width: '600px',
  top: '15vh',
  modal: true,
  closeOnClickModal: false,
  closeOnPressEscape: true,
  showClose: true,
  draggable: true,
  destroyOnClose: true,
  labelWidth: '100px',
  labelPosition: 'right',
  size: 'default',
  disabled: false,
  gutter: 20,
  confirmText: '确定',
  cancelText: '取消',
  submitLoading: false
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submit': [data: any]
  'cancel': []
  'close': []
  'open': []
}>()

const formRef = ref<FormInstance>()
const formData = ref<Record<string, any>>({})
const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const dialogTitle = computed(() => {
  if (props.title) return props.title
  
  const titleMap = {
    add: '新增',
    edit: '编辑',
    view: '查看'
  }
  
  return titleMap[props.mode]
})

// 初始化表单数据
const initFormData = () => {
  const data: Record<string, any> = {}
  
  props.fields.forEach(field => {
    if (field.type === 'checkbox' || field.multiple) {
      data[field.prop] = props.data[field.prop] || field.defaultValue || []
    } else if (field.type === 'switch') {
      data[field.prop] = props.data[field.prop] ?? field.defaultValue ?? false
    } else if (field.type === 'number' || field.type === 'slider') {
      data[field.prop] = props.data[field.prop] ?? field.defaultValue ?? 0
    } else {
      data[field.prop] = props.data[field.prop] ?? field.defaultValue ?? ''
    }
  })
  
  formData.value = data
}

// 监听数据变化
watch(() => props.data, () => {
  if (visible.value) {
    initFormData()
  }
}, { deep: true })

// 监听弹窗显示
watch(visible, (val) => {
  if (val) {
    initFormData()
  }
})

// 判断字段是否显示
const shouldShowField = (field: Field) => {
  if (typeof field.show === 'function') {
    return field.show(formData.value)
  }
  return true
}

// 上传成功
const handleUploadSuccess = (prop: string, response: any, file: any, fileList: any[]) => {
  const field = props.fields.find(f => f.prop === prop)
  if (field?.onSuccess) {
    field.onSuccess(response, file, fileList)
  } else {
    formData.value[prop] = fileList
    formData.value[prop + 'List'] = fileList
  }
}

// 上传移除
const handleUploadRemove = (prop: string, file: any, fileList: any[]) => {
  const field = props.fields.find(f => f.prop === prop)
  if (field?.onRemove) {
    field.onRemove(file, fileList)
  } else {
    formData.value[prop] = fileList
    formData.value[prop + 'List'] = fileList
  }
}

// 提交
const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    
    // 提交前处理
    if (props.beforeSubmit) {
      const canSubmit = await props.beforeSubmit(formData.value)
      if (!canSubmit) return
    }
    
    emit('submit', formData.value)
  } catch (error) {
    console.error('表单验证失败:', error)
  }
}

// 取消
const handleCancel = () => {
  emit('cancel')
  visible.value = false
}

// 关闭
const handleClose = () => {
  formRef.value?.resetFields()
  emit('close')
}

// 打开
const handleOpen = () => {
  emit('open')
}

// 重置表单
const resetForm = () => {
  formRef.value?.resetFields()
  initFormData()
}

// 验证表单
const validate = () => {
  return formRef.value?.validate()
}

// 清除验证
const clearValidate = (props?: string | string[]) => {
  formRef.value?.clearValidate(props)
}

// 暴露方法
defineExpose({
  resetForm,
  validate,
  clearValidate,
  getFormData: () => formData.value,
  setFormData: (data: Record<string, any>) => {
    formData.value = { ...formData.value, ...data }
  }
})
</script>

<style scoped lang="scss">
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.el-dialog__body) {
  max-height: 60vh;
  overflow-y: auto;
}

:deep(.el-upload__tip) {
  color: #999;
  font-size: 12px;
  margin-top: 5px;
}
</style>
