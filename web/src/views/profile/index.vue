<template>
  <div class="profile-container">
    <el-row :gutter="20">
      <!-- 左侧个人信息卡片 -->
      <el-col :span="8">
        <el-card class="user-card">
          <template #header>
            <div class="card-header">
              <span>个人信息</span>
            </div>
          </template>

          <div class="user-info">
            <!-- 头像 -->
            <div class="avatar-container">
              <el-avatar :size="100" :src="userInfo?.avatar">
                <el-icon :size="50"><User /></el-icon>
              </el-avatar>
              <el-upload
                :show-file-list="false"
                :before-upload="beforeAvatarUpload"
                :http-request="handleAvatarUpload"
                class="avatar-upload"
              >
                <el-button type="primary" size="small">更换头像</el-button>
              </el-upload>
            </div>

            <!-- 基本信息 -->
            <div class="info-item">
              <label>用户名：</label>
              <span>{{ userInfo?.username }}</span>
            </div>
            <div class="info-item">
              <label>昵称：</label>
              <span>{{ userInfo?.nickname }}</span>
            </div>
            <div class="info-item">
              <label>邮箱：</label>
              <span>{{ userInfo?.email }}</span>
            </div>
            <div class="info-item">
              <label>手机号：</label>
              <span>{{ userInfo?.phone }}</span>
            </div>
            <div class="info-item">
              <label>角色：</label>
              <span>
                <el-tag
                  v-for="role in userInfo?.roles"
                  :key="role"
                  type="info"
                  size="small"
                  style="margin-right: 4px"
                >
                  {{ role }}
                </el-tag>
              </span>
            </div>
            <div class="info-item">
              <label>创建时间：</label>
              <span>{{ userInfo?.createdAt }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 右侧设置面板 -->
      <el-col :span="16">
        <el-card>
          <el-tabs v-model="activeTab">
            <!-- 基本信息修改 -->
            <el-tab-pane label="基本信息" name="info">
              <el-form
                ref="infoFormRef"
                :model="infoForm"
                :rules="infoRules"
                label-width="100px"
                style="max-width: 500px"
              >
                <el-form-item label="昵称" prop="nickname">
                  <el-input v-model="infoForm.nickname" placeholder="请输入昵称" />
                </el-form-item>
                <el-form-item label="邮箱" prop="email">
                  <el-input v-model="infoForm.email" placeholder="请输入邮箱" />
                </el-form-item>
                <el-form-item label="手机号" prop="phone">
                  <el-input v-model="infoForm.phone" placeholder="请输入手机号" />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" :loading="infoLoading" @click="handleUpdateInfo">
                    保存修改
                  </el-button>
                </el-form-item>
              </el-form>
            </el-tab-pane>

            <!-- 密码修改 -->
            <el-tab-pane label="修改密码" name="password">
              <el-form
                ref="passwordFormRef"
                :model="passwordForm"
                :rules="passwordRules"
                label-width="100px"
                style="max-width: 500px"
              >
                <el-form-item label="当前密码" prop="oldPassword">
                  <el-input
                    v-model="passwordForm.oldPassword"
                    type="password"
                    placeholder="请输入当前密码"
                    show-password
                  />
                </el-form-item>
                <el-form-item label="新密码" prop="newPassword">
                  <el-input
                    v-model="passwordForm.newPassword"
                    type="password"
                    placeholder="请输入新密码"
                    show-password
                  />
                </el-form-item>
                <el-form-item label="确认密码" prop="confirmPassword">
                  <el-input
                    v-model="passwordForm.confirmPassword"
                    type="password"
                    placeholder="请再次输入新密码"
                    show-password
                  />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" :loading="passwordLoading" @click="handleChangePassword">
                    修改密码
                  </el-button>
                  <el-button @click="resetPasswordForm">重置</el-button>
                </el-form-item>
              </el-form>
            </el-tab-pane>

            <!-- 偏好设置 -->
            <el-tab-pane label="偏好设置" name="preferences">
              <el-form label-width="120px" style="max-width: 500px">
                <el-form-item label="主题">
                  <el-radio-group v-model="preferences.theme" @change="handlePreferenceChange">
                    <el-radio label="light">浅色</el-radio>
                    <el-radio label="dark">深色</el-radio>
                    <el-radio label="auto">跟随系统</el-radio>
                  </el-radio-group>
                </el-form-item>
                <el-form-item label="语言">
                  <el-select v-model="preferences.language" @change="handlePreferenceChange">
                    <el-option label="简体中文" value="zh-CN" />
                    <el-option label="English" value="en-US" />
                  </el-select>
                </el-form-item>
                <el-form-item label="消息通知">
                  <el-switch v-model="preferences.notification" @change="handlePreferenceChange" />
                </el-form-item>
                <el-form-item label="声音提醒">
                  <el-switch v-model="preferences.sound" @change="handlePreferenceChange" />
                </el-form-item>
                <el-form-item label="自动刷新">
                  <el-switch v-model="preferences.autoRefresh" @change="handlePreferenceChange" />
                </el-form-item>
                <el-form-item v-if="preferences.autoRefresh" label="刷新间隔">
                  <el-select v-model="preferences.refreshInterval" @change="handlePreferenceChange">
                    <el-option label="5秒" :value="5" />
                    <el-option label="10秒" :value="10" />
                    <el-option label="30秒" :value="30" />
                    <el-option label="1分钟" :value="60" />
                  </el-select>
                </el-form-item>
              </el-form>
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { getProfile, updateProfile, changePassword, uploadAvatar, getPreferences, updatePreferences } from '@/api/profile'
import type { UserInfo } from '@/types'

const userStore = useUserStore()
const activeTab = ref('info')
const userInfo = ref<UserInfo | null>(null)
const infoLoading = ref(false)
const passwordLoading = ref(false)

const infoFormRef = ref<FormInstance>()
const passwordFormRef = ref<FormInstance>()

const infoForm = reactive({
  nickname: '',
  email: '',
  phone: '',
})

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const preferences = reactive({
  theme: 'light',
  language: 'zh-CN',
  notification: true,
  sound: true,
  autoRefresh: true,
  refreshInterval: 10,
})

const infoRules: FormRules = {
  nickname: [
    { required: true, message: '请输入昵称', trigger: 'blur' },
    { min: 2, max: 20, message: '昵称长度为2-20个字符', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' },
  ],
}

const validateConfirmPassword = (_rule: any, value: string, callback: any) => {
  if (value !== passwordForm.newPassword) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const passwordRules: FormRules = {
  oldPassword: [
    { required: true, message: '请输入当前密码', trigger: 'blur' },
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度为6-20个字符', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
}

// 加载用户信息
async function loadUserInfo() {
  try {
    userInfo.value = await getProfile()
    infoForm.nickname = userInfo.value.nickname
    infoForm.email = userInfo.value.email
    infoForm.phone = userInfo.value.phone
  } catch (error) {
    console.error('获取用户信息失败:', error)
  }
}

// 加载偏好设置
async function loadPreferences() {
  try {
    const prefs = await getPreferences()
    Object.assign(preferences, prefs)
  } catch (error) {
    console.error('获取偏好设置失败:', error)
  }
}

// 更新个人信息
async function handleUpdateInfo() {
  if (!infoFormRef.value) return

  await infoFormRef.value.validate(async (valid) => {
    if (!valid) return

    infoLoading.value = true
    try {
      await updateProfile(infoForm)
      ElMessage.success('更新成功')
      loadUserInfo()
      // 更新store中的用户信息
      userStore.getUserInfoAction()
    } catch (error) {
      console.error('更新失败:', error)
    } finally {
      infoLoading.value = false
    }
  })
}

// 修改密码
async function handleChangePassword() {
  if (!passwordFormRef.value) return

  await passwordFormRef.value.validate(async (valid) => {
    if (!valid) return

    passwordLoading.value = true
    try {
      await changePassword(passwordForm)
      ElMessage.success('密码修改成功，请重新登录')
      // 清除登录状态
      userStore.logoutAction()
    } catch (error) {
      console.error('修改密码失败:', error)
    } finally {
      passwordLoading.value = false
    }
  })
}

// 重置密码表单
function resetPasswordForm() {
  passwordForm.oldPassword = ''
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
  passwordFormRef.value?.clearValidate()
}

// 头像上传前校验
function beforeAvatarUpload(file: File) {
  const isImage = file.type.startsWith('image/')
  const isLt2M = file.size / 1024 / 1024 < 2

  if (!isImage) {
    ElMessage.error('只能上传图片文件!')
    return false
  }
  if (!isLt2M) {
    ElMessage.error('图片大小不能超过 2MB!')
    return false
  }
  return true
}

// 上传头像
async function handleAvatarUpload(options: UploadRequestOptions) {
  try {
    const result = await uploadAvatar(options.file as File, (progress) => {
      console.log('上传进度:', progress)
    })
    ElMessage.success('头像上传成功')
    // 更新用户信息
    if (userInfo.value) {
      userInfo.value.avatar = result.url
    }
    // 更新store中的用户信息
    userStore.getUserInfoAction()
  } catch (error) {
    console.error('上传头像失败:', error)
    ElMessage.error('上传头像失败')
  }
}

// 偏好设置变化
async function handlePreferenceChange() {
  try {
    await updatePreferences(preferences)
    ElMessage.success('设置已保存')
    // 应用主题
    if (preferences.theme === 'dark') {
      document.documentElement.classList.add('dark')
    } else if (preferences.theme === 'light') {
      document.documentElement.classList.remove('dark')
    } else {
      // 跟随系统
      if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
        document.documentElement.classList.add('dark')
      } else {
        document.documentElement.classList.remove('dark')
      }
    }
  } catch (error) {
    console.error('保存设置失败:', error)
  }
}

onMounted(() => {
  loadUserInfo()
  loadPreferences()
})
</script>

<style scoped lang="scss">
.profile-container {
  padding: 20px;

  .user-card {
    .card-header {
      font-size: 16px;
      font-weight: bold;
    }

    .user-info {
      .avatar-container {
        display: flex;
        flex-direction: column;
        align-items: center;
        margin-bottom: 30px;

        .avatar-upload {
          margin-top: 15px;
        }
      }

      .info-item {
        display: flex;
        padding: 12px 0;
        border-bottom: 1px solid #ebeef5;

        &:last-child {
          border-bottom: none;
        }

        label {
          width: 80px;
          color: #606266;
          font-weight: 500;
        }

        span {
          flex: 1;
          color: #303133;
        }
      }
    }
  }
}
</style>
