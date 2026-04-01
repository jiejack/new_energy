<template>
  <div class="settings-container">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>系统设置</span>
        </div>
      </template>

      <el-tabs v-model="activeTab" tab-position="left" class="settings-tabs">
        <!-- 基本设置 -->
        <el-tab-pane label="基本设置" name="basic">
          <div class="settings-content">
            <h3 class="section-title">基本设置</h3>
            <el-form
              ref="basicFormRef"
              :model="basicForm"
              :rules="basicRules"
              label-width="120px"
              class="settings-form"
            >
              <el-form-item label="系统名称" prop="systemName">
                <el-input
                  v-model="basicForm.systemName"
                  placeholder="请输入系统名称"
                  maxlength="50"
                  show-word-limit
                />
              </el-form-item>

              <el-form-item label="系统Logo" prop="logo">
                <div class="logo-upload">
                  <el-upload
                    class="logo-uploader"
                    :show-file-list="false"
                    :before-upload="beforeLogoUpload"
                    :http-request="handleLogoUpload"
                  >
                    <img v-if="basicForm.logo" :src="basicForm.logo" class="logo-preview" />
                    <el-icon v-else class="logo-uploader-icon"><Plus /></el-icon>
                  </el-upload>
                  <div class="logo-tips">
                    <p>建议尺寸：200x60像素</p>
                    <p>支持格式：JPG、PNG、SVG</p>
                    <p>文件大小：不超过2MB</p>
                  </div>
                </div>
              </el-form-item>

              <el-form-item label="系统语言" prop="language">
                <el-select v-model="basicForm.language" placeholder="请选择系统语言">
                  <el-option label="简体中文" value="zh-CN" />
                  <el-option label="English" value="en-US" />
                </el-select>
              </el-form-item>

              <el-form-item label="系统时区" prop="timezone">
                <el-select v-model="basicForm.timezone" placeholder="请选择系统时区" filterable>
                  <el-option
                    v-for="tz in timezoneOptions"
                    :key="tz.value"
                    :label="tz.label"
                    :value="tz.value"
                  />
                </el-select>
              </el-form-item>

              <el-form-item>
                <el-button type="primary" :loading="saving.basic" @click="handleSaveBasic">
                  保存设置
                </el-button>
                <el-button @click="handleResetBasic">重置</el-button>
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>

        <!-- 告警设置 -->
        <el-tab-pane label="告警设置" name="alarm">
          <div class="settings-content">
            <h3 class="section-title">告警设置</h3>
            <el-form
              ref="alarmFormRef"
              :model="alarmForm"
              :rules="alarmRules"
              label-width="120px"
              class="settings-form"
            >
              <el-form-item label="默认告警级别" prop="defaultLevel">
                <el-select v-model="alarmForm.defaultLevel" placeholder="请选择默认告警级别">
                  <el-option label="严重" value="critical" />
                  <el-option label="重要" value="major" />
                  <el-option label="次要" value="minor" />
                  <el-option label="警告" value="warning" />
                </el-select>
              </el-form-item>

              <el-divider content-position="left">通知方式</el-divider>

              <el-form-item label="声音提醒">
                <el-switch v-model="alarmForm.soundEnabled" />
                <span class="form-item-tip">启用后将播放告警提示音</span>
              </el-form-item>

              <el-form-item label="邮件通知">
                <el-switch v-model="alarmForm.emailEnabled" />
                <span class="form-item-tip">启用后将发送邮件通知</span>
              </el-form-item>

              <el-form-item
                v-if="alarmForm.emailEnabled"
                label="邮件接收人"
                prop="emailRecipients"
              >
                <el-input
                  v-model="alarmForm.emailRecipients"
                  type="textarea"
                  :rows="3"
                  placeholder="请输入邮件接收人，多个邮箱用逗号分隔"
                />
              </el-form-item>

              <el-form-item label="短信通知">
                <el-switch v-model="alarmForm.smsEnabled" />
                <span class="form-item-tip">启用后将发送短信通知</span>
              </el-form-item>

              <el-form-item
                v-if="alarmForm.smsEnabled"
                label="短信接收人"
                prop="smsRecipients"
              >
                <el-input
                  v-model="alarmForm.smsRecipients"
                  type="textarea"
                  :rows="3"
                  placeholder="请输入手机号，多个手机号用逗号分隔"
                />
              </el-form-item>

              <el-form-item>
                <el-button type="primary" :loading="saving.alarm" @click="handleSaveAlarm">
                  保存设置
                </el-button>
                <el-button @click="handleResetAlarm">重置</el-button>
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>

        <!-- 显示设置 -->
        <el-tab-pane label="显示设置" name="display">
          <div class="settings-content">
            <h3 class="section-title">显示设置</h3>
            <el-form
              ref="displayFormRef"
              :model="displayForm"
              :rules="displayRules"
              label-width="120px"
              class="settings-form"
            >
              <el-form-item label="系统主题" prop="theme">
                <el-radio-group v-model="displayForm.theme" @change="handleThemeChange">
                  <el-radio label="light">
                    <el-icon><Sunny /></el-icon>
                    浅色
                  </el-radio>
                  <el-radio label="dark">
                    <el-icon><Moon /></el-icon>
                    深色
                  </el-radio>
                </el-radio-group>
              </el-form-item>

              <el-form-item label="默认分页大小" prop="pageSize">
                <el-select v-model="displayForm.pageSize" placeholder="请选择默认分页大小">
                  <el-option label="10条/页" :value="10" />
                  <el-option label="20条/页" :value="20" />
                  <el-option label="50条/页" :value="50" />
                  <el-option label="100条/页" :value="100" />
                </el-select>
              </el-form-item>

              <el-form-item label="数据刷新间隔" prop="refreshInterval">
                <el-select v-model="displayForm.refreshInterval" placeholder="请选择数据刷新间隔">
                  <el-option label="5秒" :value="5" />
                  <el-option label="10秒" :value="10" />
                  <el-option label="30秒" :value="30" />
                  <el-option label="1分钟" :value="60" />
                  <el-option label="5分钟" :value="300" />
                </el-select>
              </el-form-item>

              <el-form-item label="日期格式" prop="dateFormat">
                <el-select v-model="displayForm.dateFormat" placeholder="请选择日期格式">
                  <el-option label="YYYY-MM-DD HH:mm:ss" value="YYYY-MM-DD HH:mm:ss" />
                  <el-option label="YYYY/MM/DD HH:mm:ss" value="YYYY/MM/DD HH:mm:ss" />
                  <el-option label="MM/DD/YYYY HH:mm:ss" value="MM/DD/YYYY HH:mm:ss" />
                  <el-option label="DD-MM-YYYY HH:mm:ss" value="DD-MM-YYYY HH:mm:ss" />
                </el-select>
              </el-form-item>

              <el-form-item>
                <el-button type="primary" :loading="saving.display" @click="handleSaveDisplay">
                  保存设置
                </el-button>
                <el-button @click="handleResetDisplay">重置</el-button>
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
import { useConfigStore } from '@/stores/config'
import type { BasicConfig, AlarmConfig, DisplayConfig } from '@/api/config'

const configStore = useConfigStore()

const loading = ref(false)
const activeTab = ref('basic')

const basicFormRef = ref<FormInstance>()
const alarmFormRef = ref<FormInstance>()
const displayFormRef = ref<FormInstance>()

const saving = reactive({
  basic: false,
  alarm: false,
  display: false,
})

// 基本设置表单
const basicForm = reactive<BasicConfig>({
  systemName: '',
  logo: '',
  language: 'zh-CN',
  timezone: 'Asia/Shanghai',
})

// 告警设置表单
const alarmForm = reactive<AlarmConfig>({
  defaultLevel: 'warning',
  soundEnabled: true,
  emailEnabled: false,
  smsEnabled: false,
  emailRecipients: '',
  smsRecipients: '',
})

// 显示设置表单
const displayForm = reactive<DisplayConfig>({
  theme: 'light',
  pageSize: 10,
  refreshInterval: 10,
  dateFormat: 'YYYY-MM-DD HH:mm:ss',
})

// 时区选项
const timezoneOptions = [
  { label: '(UTC+08:00) 北京，上海', value: 'Asia/Shanghai' },
  { label: '(UTC+08:00) 香港', value: 'Asia/Hong_Kong' },
  { label: '(UTC+08:00) 台北', value: 'Asia/Taipei' },
  { label: '(UTC+09:00) 东京', value: 'Asia/Tokyo' },
  { label: '(UTC+09:00) 首尔', value: 'Asia/Seoul' },
  { label: '(UTC+00:00) 伦敦', value: 'Europe/London' },
  { label: '(UTC+01:00) 巴黎', value: 'Europe/Paris' },
  { label: '(UTC-05:00) 纽约', value: 'America/New_York' },
  { label: '(UTC-08:00) 洛杉矶', value: 'America/Los_Angeles' },
]

// 基本设置验证规则
const basicRules: FormRules = {
  systemName: [
    { required: true, message: '请输入系统名称', trigger: 'blur' },
    { min: 2, max: 50, message: '系统名称长度为2-50个字符', trigger: 'blur' },
  ],
  language: [{ required: true, message: '请选择系统语言', trigger: 'change' }],
  timezone: [{ required: true, message: '请选择系统时区', trigger: 'change' }],
}

// 告警设置验证规则
const alarmRules: FormRules = {
  defaultLevel: [{ required: true, message: '请选择默认告警级别', trigger: 'change' }],
  emailRecipients: [
    {
      validator: (_rule, value, callback) => {
        if (alarmForm.emailEnabled && !value) {
          callback(new Error('请输入邮件接收人'))
        } else if (value) {
          const emails = value.split(',').map((e: string) => e.trim())
          const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
          const invalidEmail = emails.find((e: string) => !emailRegex.test(e))
          if (invalidEmail) {
            callback(new Error(`邮箱格式不正确: ${invalidEmail}`))
          } else {
            callback()
          }
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
  smsRecipients: [
    {
      validator: (_rule, value, callback) => {
        if (alarmForm.smsEnabled && !value) {
          callback(new Error('请输入短信接收人'))
        } else if (value) {
          const phones = value.split(',').map((p: string) => p.trim())
          const phoneRegex = /^1[3-9]\d{9}$/
          const invalidPhone = phones.find((p: string) => !phoneRegex.test(p))
          if (invalidPhone) {
            callback(new Error(`手机号格式不正确: ${invalidPhone}`))
          } else {
            callback()
          }
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
}

// 显示设置验证规则
const displayRules: FormRules = {
  theme: [{ required: true, message: '请选择系统主题', trigger: 'change' }],
  pageSize: [{ required: true, message: '请选择默认分页大小', trigger: 'change' }],
  refreshInterval: [{ required: true, message: '请选择数据刷新间隔', trigger: 'change' }],
  dateFormat: [{ required: true, message: '请选择日期格式', trigger: 'change' }],
}

// 加载配置
async function loadConfigs() {
  loading.value = true
  try {
    await configStore.loadConfigs()
    // 填充表单
    Object.assign(basicForm, configStore.basicConfig)
    Object.assign(alarmForm, configStore.alarmConfig)
    Object.assign(displayForm, configStore.displayConfig)
  } catch (error) {
    ElMessage.error('加载配置失败')
  } finally {
    loading.value = false
  }
}

// 保存基本设置
async function handleSaveBasic() {
  if (!basicFormRef.value) return

  await basicFormRef.value.validate(async (valid) => {
    if (!valid) return

    saving.basic = true
    try {
      await configStore.updateBasicConfig(basicForm)
      ElMessage.success('基本设置保存成功')
    } catch (error) {
      ElMessage.error('保存失败')
    } finally {
      saving.basic = false
    }
  })
}

// 重置基本设置
function handleResetBasic() {
  Object.assign(basicForm, configStore.basicConfig)
  basicFormRef.value?.clearValidate()
}

// 保存告警设置
async function handleSaveAlarm() {
  if (!alarmFormRef.value) return

  await alarmFormRef.value.validate(async (valid) => {
    if (!valid) return

    saving.alarm = true
    try {
      await configStore.updateAlarmConfig(alarmForm)
      ElMessage.success('告警设置保存成功')
    } catch (error) {
      ElMessage.error('保存失败')
    } finally {
      saving.alarm = false
    }
  })
}

// 重置告警设置
function handleResetAlarm() {
  Object.assign(alarmForm, configStore.alarmConfig)
  alarmFormRef.value?.clearValidate()
}

// 保存显示设置
async function handleSaveDisplay() {
  if (!displayFormRef.value) return

  await displayFormRef.value.validate(async (valid) => {
    if (!valid) return

    saving.display = true
    try {
      await configStore.updateDisplayConfig(displayForm)
      ElMessage.success('显示设置保存成功')
    } catch (error) {
      ElMessage.error('保存失败')
    } finally {
      saving.display = false
    }
  })
}

// 重置显示设置
function handleResetDisplay() {
  Object.assign(displayForm, configStore.displayConfig)
  displayFormRef.value?.clearValidate()
}

// Logo上传前校验
function beforeLogoUpload(file: File) {
  const isImage = ['image/jpeg', 'image/png', 'image/svg+xml'].includes(file.type)
  const isLt2M = file.size / 1024 / 1024 < 2

  if (!isImage) {
    ElMessage.error('只能上传 JPG/PNG/SVG 格式的图片!')
    return false
  }
  if (!isLt2M) {
    ElMessage.error('图片大小不能超过 2MB!')
    return false
  }
  return true
}

// 上传Logo
async function handleLogoUpload(options: UploadRequestOptions) {
  try {
    // 这里模拟上传，实际项目中应该调用上传API
    const file = options.file as File
    const reader = new FileReader()
    reader.onload = (e) => {
      basicForm.logo = e.target?.result as string
      ElMessage.success('Logo上传成功')
    }
    reader.readAsDataURL(file)
  } catch (error) {
    ElMessage.error('上传Logo失败')
  }
}

// 主题变化
function handleThemeChange(theme: string) {
  if (theme === 'dark') {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}

onMounted(() => {
  loadConfigs()
})
</script>

<style scoped lang="scss">
.settings-container {
  padding: 20px;

  .card-header {
    font-size: 16px;
    font-weight: bold;
  }

  .settings-tabs {
    min-height: 600px;

    :deep(.el-tabs__content) {
      padding: 0 20px;
    }
  }

  .settings-content {
    .section-title {
      margin: 0 0 20px 0;
      padding-bottom: 10px;
      border-bottom: 1px solid #ebeef5;
      font-size: 16px;
      font-weight: 500;
      color: #303133;
    }
  }

  .settings-form {
    max-width: 600px;

    .form-item-tip {
      margin-left: 10px;
      color: #909399;
      font-size: 12px;
    }
  }

  .logo-upload {
    display: flex;
    align-items: flex-start;

    .logo-uploader {
      width: 200px;
      height: 60px;
      border: 1px dashed #d9d9d9;
      border-radius: 6px;
      cursor: pointer;
      overflow: hidden;
      display: flex;
      align-items: center;
      justify-content: center;
      transition: border-color 0.3s;

      &:hover {
        border-color: #409eff;
      }

      .logo-preview {
        width: 100%;
        height: 100%;
        object-fit: contain;
      }

      .logo-uploader-icon {
        font-size: 28px;
        color: #8c939d;
      }
    }

    .logo-tips {
      margin-left: 20px;
      color: #909399;
      font-size: 12px;
      line-height: 1.8;

      p {
        margin: 0;
      }
    }
  }

  :deep(.el-divider__text) {
    font-size: 14px;
    color: #606266;
  }
}

// 响应式布局
@media (max-width: 768px) {
  .settings-container {
    padding: 10px;

    .settings-tabs {
      :deep(.el-tabs__nav-wrap) {
        &::after {
          display: none;
        }
      }

      :deep(.el-tabs__nav-scroll) {
        overflow-x: auto;
      }

      :deep(.el-tabs__content) {
        padding: 0 10px;
      }
    }

    .settings-form {
      max-width: 100%;
    }

    .logo-upload {
      flex-direction: column;

      .logo-tips {
        margin-left: 0;
        margin-top: 10px;
      }
    }
  }
}
</style>
