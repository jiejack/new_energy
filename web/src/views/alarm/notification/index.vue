<template>
  <div class="notification-page">
    <el-tabs v-model="activeTab">
      <!-- 通知渠道配置 -->
      <el-tab-pane label="通知渠道" name="channel">
        <el-card shadow="never">
          <el-table :data="channelList" border stripe>
            <el-table-column prop="name" label="渠道名称" width="150" />
            <el-table-column prop="type" label="渠道类型" width="120">
              <template #default="{ row }">
                <el-tag :type="getChannelTagType(row.type)">
                  {{ getChannelTypeName(row.type) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="config" label="配置信息" min-width="250">
              <template #default="{ row }">
                <span v-if="row.type === 'sms'">签名: {{ row.config.signature }}</span>
                <span v-else-if="row.type === 'email'">
                  SMTP: {{ row.config.smtpServer }}:{{ row.config.smtpPort }}
                </span>
                <span v-else-if="row.type === 'dingtalk'">
                  Webhook: {{ row.config.webhook?.slice(0, 50) }}...
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.enabled ? 'success' : 'info'">
                  {{ row.enabled ? '已启用' : '已禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleEditChannel(row)">
                  配置
                </el-button>
                <el-button type="success" link size="small" @click="handleTestChannel(row)">
                  测试
                </el-button>
                <el-button
                  :type="row.enabled ? 'warning' : 'success'"
                  link
                  size="small"
                  @click="handleToggleChannel(row)"
                >
                  {{ row.enabled ? '禁用' : '启用' }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 通知模板管理 -->
      <el-tab-pane label="通知模板" name="template">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">
              <span>通知模板列表</span>
              <el-button type="primary" size="small" @click="handleAddTemplate">
                <el-icon><Plus /></el-icon>
                新增模板
              </el-button>
            </div>
          </template>

          <el-table :data="templateList" border stripe>
            <el-table-column prop="name" label="模板名称" width="150" />
            <el-table-column prop="type" label="模板类型" width="120">
              <template #default="{ row }">
                <el-tag>{{ getChannelTypeName(row.type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="title" label="模板标题" min-width="200" />
            <el-table-column prop="content" label="模板内容" min-width="300">
              <template #default="{ row }">
                <el-text line-clamp="2">{{ row.content }}</el-text>
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="创建时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.createdAt) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleEditTemplate(row)">
                  编辑
                </el-button>
                <el-button type="danger" link size="small" @click="handleDeleteTemplate(row)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 通知规则配置 -->
      <el-tab-pane label="通知规则" name="rule">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">
              <span>通知规则列表</span>
              <el-button type="primary" size="small" @click="handleAddNotifyRule">
                <el-icon><Plus /></el-icon>
                新增规则
              </el-button>
            </div>
          </template>

          <el-table :data="notifyRuleList" border stripe>
            <el-table-column prop="name" label="规则名称" width="150" />
            <el-table-column prop="alarmLevel" label="告警级别" width="150">
              <template #default="{ row }">
                <el-tag
                  v-for="level in row.alarmLevels"
                  :key="level"
                  :type="getLevelTagType(level)"
                  size="small"
                  style="margin-right: 4px"
                >
                  {{ getLevelText(level) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="channels" label="通知渠道" width="200">
              <template #default="{ row }">
                <el-tag
                  v-for="channel in row.channels"
                  :key="channel"
                  size="small"
                  style="margin-right: 4px"
                >
                  {{ getChannelTypeName(channel) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="receivers" label="接收人" min-width="200">
              <template #default="{ row }">
                {{ row.receivers.join(', ') }}
              </template>
            </el-table-column>
            <el-table-column prop="enabled" label="状态" width="100">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" @change="handleToggleNotifyRule(row)" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleEditNotifyRule(row)">
                  编辑
                </el-button>
                <el-button type="danger" link size="small" @click="handleDeleteNotifyRule(row)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 渠道配置对话框 -->
    <el-dialog v-model="channelDialogVisible" title="渠道配置" width="500px">
      <el-form :model="channelForm" label-width="100px">
        <el-form-item label="渠道名称">
          <el-input v-model="channelForm.name" disabled />
        </el-form-item>

        <!-- 短信配置 -->
        <template v-if="channelForm.type === 'sms'">
          <el-form-item label="AccessKey">
            <el-input v-model="channelForm.config.accessKey" placeholder="请输入AccessKey" />
          </el-form-item>
          <el-form-item label="AccessSecret">
            <el-input
              v-model="channelForm.config.accessSecret"
              type="password"
              placeholder="请输入AccessSecret"
              show-password
            />
          </el-form-item>
          <el-form-item label="签名">
            <el-input v-model="channelForm.config.signature" placeholder="请输入短信签名" />
          </el-form-item>
          <el-form-item label="模板Code">
            <el-input v-model="channelForm.config.templateCode" placeholder="请输入短信模板Code" />
          </el-form-item>
        </template>

        <!-- 邮件配置 -->
        <template v-if="channelForm.type === 'email'">
          <el-form-item label="SMTP服务器">
            <el-input v-model="channelForm.config.smtpServer" placeholder="smtp.example.com" />
          </el-form-item>
          <el-form-item label="SMTP端口">
            <el-input-number v-model="channelForm.config.smtpPort" :min="1" :max="65535" />
          </el-form-item>
          <el-form-item label="发件人邮箱">
            <el-input v-model="channelForm.config.sender" placeholder="noreply@example.com" />
          </el-form-item>
          <el-form-item label="邮箱密码">
            <el-input
              v-model="channelForm.config.password"
              type="password"
              placeholder="请输入邮箱密码或授权码"
              show-password
            />
          </el-form-item>
          <el-form-item label="使用SSL">
            <el-switch v-model="channelForm.config.useSSL" />
          </el-form-item>
        </template>

        <!-- 钉钉配置 -->
        <template v-if="channelForm.type === 'dingtalk'">
          <el-form-item label="Webhook地址">
            <el-input
              v-model="channelForm.config.webhook"
              type="textarea"
              :rows="2"
              placeholder="请输入钉钉机器人Webhook地址"
            />
          </el-form-item>
          <el-form-item label="加签密钥">
            <el-input
              v-model="channelForm.config.secret"
              type="password"
              placeholder="选填，安全设置中的加签密钥"
              show-password
            />
          </el-form-item>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="channelDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveChannel">保存</el-button>
      </template>
    </el-dialog>

    <!-- 模板编辑对话框 -->
    <el-dialog v-model="templateDialogVisible" :title="isEditTemplate ? '编辑模板' : '新增模板'" width="600px">
      <el-form :model="templateForm" label-width="100px">
        <el-form-item label="模板名称">
          <el-input v-model="templateForm.name" placeholder="请输入模板名称" />
        </el-form-item>
        <el-form-item label="模板类型">
          <el-select v-model="templateForm.type" placeholder="请选择模板类型">
            <el-option label="短信" value="sms" />
            <el-option label="邮件" value="email" />
            <el-option label="钉钉" value="dingtalk" />
          </el-select>
        </el-form-item>
        <el-form-item label="模板标题">
          <el-input v-model="templateForm.title" placeholder="请输入模板标题" />
        </el-form-item>
        <el-form-item label="模板内容">
          <el-input
            v-model="templateForm.content"
            type="textarea"
            :rows="5"
            placeholder="请输入模板内容，支持变量: {alarmTitle}, {alarmLevel}, {alarmContent}, {sourceName}, {occurredAt}"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="templateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveTemplate">保存</el-button>
      </template>
    </el-dialog>

    <!-- 通知规则编辑对话框 -->
    <el-dialog v-model="notifyRuleDialogVisible" :title="isEditNotifyRule ? '编辑规则' : '新增规则'" width="600px">
      <el-form :model="notifyRuleForm" label-width="100px">
        <el-form-item label="规则名称">
          <el-input v-model="notifyRuleForm.name" placeholder="请输入规则名称" />
        </el-form-item>
        <el-form-item label="告警级别">
          <el-checkbox-group v-model="notifyRuleForm.alarmLevels">
            <el-checkbox label="critical">严重</el-checkbox>
            <el-checkbox label="major">主要</el-checkbox>
            <el-checkbox label="minor">次要</el-checkbox>
            <el-checkbox label="warning">警告</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="通知渠道">
          <el-checkbox-group v-model="notifyRuleForm.channels">
            <el-checkbox label="sms">短信</el-checkbox>
            <el-checkbox label="email">邮件</el-checkbox>
            <el-checkbox label="dingtalk">钉钉</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="接收人">
          <el-select v-model="notifyRuleForm.receivers" multiple placeholder="请选择接收人" style="width: 100%">
            <el-option label="张三 (zhangsan@example.com)" value="张三" />
            <el-option label="李四 (lisi@example.com)" value="李四" />
            <el-option label="王五 (wangwu@example.com)" value="王五" />
          </el-select>
        </el-form-item>
        <el-form-item label="静默时间">
          <el-time-picker
            v-model="notifyRuleForm.silentPeriod"
            is-range
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            format="HH:mm"
          />
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch v-model="notifyRuleForm.enabled" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="notifyRuleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveNotifyRule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import type { AlarmLevel } from '@/types'
import dayjs from 'dayjs'
import {
  getNotificationConfigs,
  updateNotificationConfig,
  enableNotificationConfig,
  disableNotificationConfig,
  testNotificationConfig,
  type NotificationConfig,
  type NotificationType,
} from '@/api/notification'

type ChannelType = 'sms' | 'email' | 'dingtalk' | 'webhook' | 'wechat'

interface Channel {
  id: string
  name: string
  type: ChannelType
  config: Record<string, any>
  enabled: boolean
}

interface Template {
  id: number
  name: string
  type: ChannelType
  title: string
  content: string
  createdAt: string
}

interface NotifyRule {
  id: number
  name: string
  alarmLevels: AlarmLevel[]
  channels: ChannelType[]
  receivers: string[]
  silentPeriod: [Date, Date] | null
  enabled: boolean
}

const activeTab = ref('channel')
const channelDialogVisible = ref(false)
const templateDialogVisible = ref(false)
const notifyRuleDialogVisible = ref(false)
const isEditTemplate = ref(false)
const isEditNotifyRule = ref(false)
const loading = ref(false)

// 渠道列表
const channelList = ref<Channel[]>([])

// 模板列表
const templateList = ref<Template[]>([
  {
    id: 1,
    name: '严重告警通知',
    type: 'sms',
    title: '',
    content: '【新能源监控】严重告警：{alarmTitle}，来源：{sourceName}，时间：{occurredAt}',
    createdAt: '2024-01-01 10:00:00',
  },
  {
    id: 2,
    name: '告警邮件通知',
    type: 'email',
    title: '【新能源监控】{alarmTitle}',
    content: '告警级别：{alarmLevel}\n告警内容：{alarmContent}\n告警来源：{sourceName}\n发生时间：{occurredAt}',
    createdAt: '2024-01-01 10:00:00',
  },
])

// 通知规则列表
const notifyRuleList = ref<NotifyRule[]>([
  {
    id: 1,
    name: '严重告警通知',
    alarmLevels: ['critical'],
    channels: ['sms', 'email', 'dingtalk'],
    receivers: ['张三', '李四'],
    silentPeriod: null,
    enabled: true,
  },
  {
    id: 2,
    name: '主要告警通知',
    alarmLevels: ['major'],
    channels: ['email', 'dingtalk'],
    receivers: ['张三'],
    silentPeriod: null,
    enabled: true,
  },
])

// 表单数据
const channelForm = reactive<Channel>({
  id: '',
  name: '',
  type: 'sms',
  config: {},
  enabled: false,
})

const templateForm = reactive({
  id: 0,
  name: '',
  type: 'sms' as ChannelType,
  title: '',
  content: '',
})

const notifyRuleForm = reactive({
  id: 0,
  name: '',
  alarmLevels: [] as AlarmLevel[],
  channels: [] as ChannelType[],
  receivers: [] as string[],
  silentPeriod: null as [Date, Date] | null,
  enabled: true,
})

// 获取渠道类型名称
const getChannelTypeName = (type: ChannelType): string => {
  const nameMap: Record<ChannelType, string> = {
    sms: '短信',
    email: '邮件',
    dingtalk: '钉钉',
    webhook: 'Webhook',
    wechat: '微信',
  }
  return nameMap[type]
}

// 获取渠道标签类型
const getChannelTagType = (type: ChannelType): 'primary' | 'success' | 'warning' | 'info' => {
  const typeMap: Record<ChannelType, 'primary' | 'success' | 'warning' | 'info'> = {
    sms: 'primary',
    email: 'success',
    dingtalk: 'warning',
    webhook: 'info',
    wechat: 'success',
  }
  return typeMap[type]
}

// 获取级别标签类型
const getLevelTagType = (level: AlarmLevel): 'danger' | 'warning' | 'info' | '' => {
  const typeMap: Record<AlarmLevel, 'danger' | 'warning' | 'info' | ''> = {
    critical: 'danger',
    major: 'warning',
    minor: 'info',
    warning: '',
  }
  return typeMap[level]
}

// 获取级别文本
const getLevelText = (level: AlarmLevel): string => {
  const textMap: Record<AlarmLevel, string> = {
    critical: '严重',
    major: '主要',
    minor: '次要',
    warning: '警告',
  }
  return textMap[level]
}

// 格式化时间
const formatTime = (time: string): string => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

// 获取通知配置列表
const fetchNotificationConfigs = async () => {
  loading.value = true
  try {
    const configs = await getNotificationConfigs()
    channelList.value = configs.map((config: NotificationConfig) => ({
      id: config.id,
      name: config.name,
      type: config.type as ChannelType,
      config: config.config,
      enabled: config.enabled,
    }))
  } catch (error: any) {
    ElMessage.error(error.message || '获取通知配置失败')
  } finally {
    loading.value = false
  }
}

// 编辑渠道
const handleEditChannel = (row: Channel) => {
  Object.assign(channelForm, row)
  channelDialogVisible.value = true
}

// 测试渠道
const handleTestChannel = async (row: Channel) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入测试接收地址', '测试通知', {
      confirmButtonText: '发送',
      cancelButtonText: '取消',
      inputPattern: row.type === 'email' ? /^[^\s@]+@[^\s@]+\.[^\s@]+$/ : /^1[3-9]\d{9}$/,
      inputErrorMessage: row.type === 'email' ? '请输入正确的邮箱地址' : '请输入正确的手机号',
    })
    await testNotificationConfig(row.type as NotificationType, value)
    ElMessage.success('测试通知已发送')
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '测试发送失败')
    }
  }
}

// 切换渠道状态
const handleToggleChannel = async (row: Channel) => {
  try {
    if (row.enabled) {
      await disableNotificationConfig(row.type as NotificationType)
      row.enabled = false
      ElMessage.success('已禁用')
    } else {
      await enableNotificationConfig(row.type as NotificationType)
      row.enabled = true
      ElMessage.success('已启用')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  }
}

// 保存渠道
const handleSaveChannel = async () => {
  try {
    await updateNotificationConfig(channelForm.type as NotificationType, channelForm.config)
    const index = channelList.value.findIndex((c) => c.id === channelForm.id)
    if (index > -1) {
      channelList.value[index] = { ...channelForm }
    }
    channelDialogVisible.value = false
    ElMessage.success('保存成功')
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  }
}

// 新增模板
const handleAddTemplate = () => {
  isEditTemplate.value = false
  Object.assign(templateForm, {
    id: 0,
    name: '',
    type: 'sms',
    title: '',
    content: '',
  })
  templateDialogVisible.value = true
}

// 编辑模板
const handleEditTemplate = (row: Template) => {
  isEditTemplate.value = true
  Object.assign(templateForm, row)
  templateDialogVisible.value = true
}

// 删除模板
const handleDeleteTemplate = async (row: Template) => {
  try {
    await ElMessageBox.confirm('确定删除该模板？', '提示', { type: 'warning' })
    const index = templateList.value.findIndex((t) => t.id === row.id)
    if (index > -1) {
      templateList.value.splice(index, 1)
    }
    ElMessage.success('删除成功')
  } catch (error) {
    // 用户取消
  }
}

// 保存模板
const handleSaveTemplate = () => {
  if (isEditTemplate.value) {
    const index = templateList.value.findIndex((t) => t.id === templateForm.id)
    if (index > -1) {
      templateList.value[index] = { ...templateForm, createdAt: templateList.value[index].createdAt }
    }
  } else {
    templateList.value.push({
      ...templateForm,
      id: Date.now(),
      createdAt: new Date().toISOString(),
    })
  }
  templateDialogVisible.value = false
  ElMessage.success('保存成功')
}

// 新增通知规则
const handleAddNotifyRule = () => {
  isEditNotifyRule.value = false
  Object.assign(notifyRuleForm, {
    id: 0,
    name: '',
    alarmLevels: [],
    channels: [],
    receivers: [],
    silentPeriod: null,
    enabled: true,
  })
  notifyRuleDialogVisible.value = true
}

// 编辑通知规则
const handleEditNotifyRule = (row: NotifyRule) => {
  isEditNotifyRule.value = true
  Object.assign(notifyRuleForm, row)
  notifyRuleDialogVisible.value = true
}

// 删除通知规则
const handleDeleteNotifyRule = async (row: NotifyRule) => {
  try {
    await ElMessageBox.confirm('确定删除该规则？', '提示', { type: 'warning' })
    const index = notifyRuleList.value.findIndex((r) => r.id === row.id)
    if (index > -1) {
      notifyRuleList.value.splice(index, 1)
    }
    ElMessage.success('删除成功')
  } catch (error) {
    // 用户取消
  }
}

// 切换通知规则状态
const handleToggleNotifyRule = (row: NotifyRule) => {
  ElMessage.success(row.enabled ? '已启用' : '已禁用')
}

// 保存通知规则
const handleSaveNotifyRule = () => {
  if (isEditNotifyRule.value) {
    const index = notifyRuleList.value.findIndex((r) => r.id === notifyRuleForm.id)
    if (index > -1) {
      notifyRuleList.value[index] = { ...notifyRuleForm }
    }
  } else {
    notifyRuleList.value.push({
      ...notifyRuleForm,
      id: Date.now(),
    })
  }
  notifyRuleDialogVisible.value = false
  ElMessage.success('保存成功')
}

// 初始化
onMounted(() => {
  fetchNotificationConfigs()
})
</script>

<style scoped lang="scss">
.notification-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>
