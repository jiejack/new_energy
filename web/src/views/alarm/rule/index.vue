<template>
  <div class="alarm-rule-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">告警规则配置</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增规则
          </el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="ruleList" border stripe>
        <el-table-column prop="name" label="规则名称" min-width="150" />

        <el-table-column prop="pointName" label="关联采集点" width="150" />

        <el-table-column prop="condition" label="触发条件" width="200">
          <template #default="{ row }">
            <span>{{ getConditionText(row) }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="level" label="告警级别" width="100">
          <template #default="{ row }">
            <el-tag :type="getLevelTagType(row.level)" effect="dark" size="small">
              {{ getLevelText(row.level) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              @change="handleStatusChange(row)"
            />
          </template>
        </el-table-column>

        <el-table-column prop="createdAt" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleTest(row)">
              测试
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

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          :background="true"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchRuleList"
          @current-change="fetchRuleList"
        />
      </div>
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑规则' : '新增规则'" width="600px">
      <el-form ref="formRef" :model="ruleForm" :rules="rules" label-width="100px">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="ruleForm.name" placeholder="请输入规则名称" />
        </el-form-item>

        <el-form-item label="关联采集点" prop="pointId">
          <el-select v-model="ruleForm.pointId" placeholder="请选择采集点" style="width: 100%">
            <el-option
              v-for="point in pointList"
              :key="point.id"
              :label="`${point.name} (${point.code})`"
              :value="point.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="触发条件" prop="condition">
          <el-row :gutter="10">
            <el-col :span="8">
              <el-select v-model="ruleForm.operator" placeholder="运算符">
                <el-option label="大于" value=">" />
                <el-option label="大于等于" value=">=" />
                <el-option label="小于" value="<" />
                <el-option label="小于等于" value="<=" />
                <el-option label="等于" value="==" />
                <el-option label="不等于" value="!=" />
              </el-select>
            </el-col>
            <el-col :span="16">
              <el-input-number
                v-model="ruleForm.threshold"
                placeholder="阈值"
                style="width: 100%"
              />
            </el-col>
          </el-row>
        </el-form-item>

        <el-form-item label="持续时间" prop="duration">
          <el-input-number
            v-model="ruleForm.duration"
            :min="0"
            placeholder="持续时间(秒)"
            style="width: 200px"
          />
          <span style="margin-left: 10px; color: #909399">秒 (0表示立即触发)</span>
        </el-form-item>

        <el-form-item label="告警级别" prop="level">
          <el-select v-model="ruleForm.level" placeholder="请选择告警级别">
            <el-option label="严重" value="critical" />
            <el-option label="主要" value="major" />
            <el-option label="次要" value="minor" />
            <el-option label="警告" value="warning" />
          </el-select>
        </el-form-item>

        <el-form-item label="告警标题" prop="title">
          <el-input v-model="ruleForm.title" placeholder="请输入告警标题" />
        </el-form-item>

        <el-form-item label="告警内容" prop="content">
          <el-input
            v-model="ruleForm.content"
            type="textarea"
            :rows="3"
            placeholder="请输入告警内容，支持变量: {pointName}, {value}, {threshold}"
          />
        </el-form-item>

        <el-form-item label="通知方式" prop="notifyChannels">
          <el-checkbox-group v-model="ruleForm.notifyChannels">
            <el-checkbox label="sms">短信</el-checkbox>
            <el-checkbox label="email">邮件</el-checkbox>
            <el-checkbox label="dingtalk">钉钉</el-checkbox>
          </el-checkbox-group>
        </el-form-item>

        <el-form-item label="启用状态">
          <el-switch v-model="ruleForm.enabled" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 测试结果对话框 -->
    <el-dialog v-model="testVisible" title="规则测试结果" width="500px">
      <el-result
        :icon="testResult.success ? 'success' : 'warning'"
        :title="testResult.success ? '测试通过' : '测试未通过'"
      >
        <template #sub-title>
          <div v-if="testResult.success">
            <p>当前值: {{ testResult.currentValue }}</p>
            <p>触发条件: {{ testResult.condition }}</p>
          </div>
          <div v-else>
            <p>{{ testResult.message }}</p>
          </div>
        </template>
      </el-result>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getAlarmRuleList,
  createAlarmRule,
  updateAlarmRule,
  deleteAlarmRule,
  type AlarmRule as ApiAlarmRule,
} from '@/api/alarm-rule'
import { getPointList } from '@/api/point'
import type { Point, AlarmLevel } from '@/types'
import dayjs from 'dayjs'
import { alarmLevelMapper } from '@/utils/enums'

interface AlarmRule {
  id: string
  name: string
  pointId: number | undefined
  pointName: string
  operator: string
  threshold: number
  duration: number
  level: AlarmLevel
  title: string
  content: string
  notifyChannels: string[]
  enabled: boolean
  createdAt: string
}

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const testVisible = ref(false)
const isEdit = ref(false)
const ruleList = ref<AlarmRule[]>([])
const pointList = ref<Point[]>([])
const formRef = ref<FormInstance>()

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const ruleForm = reactive({
  id: '',
  name: '',
  pointId: undefined as number | undefined,
  operator: '>' as string,
  threshold: 0,
  duration: 0,
  level: 'warning' as AlarmLevel,
  title: '',
  content: '',
  notifyChannels: [] as string[],
  enabled: true,
})

const testResult = reactive({
  success: false,
  currentValue: 0,
  condition: '',
  message: '',
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  pointId: [{ required: true, message: '请选择采集点', trigger: 'change' }],
  operator: [{ required: true, message: '请选择运算符', trigger: 'change' }],
  threshold: [{ required: true, message: '请输入阈值', trigger: 'blur' }],
  level: [{ required: true, message: '请选择告警级别', trigger: 'change' }],
  title: [{ required: true, message: '请输入告警标题', trigger: 'blur' }],
}

const getLevelTagType = (level: AlarmLevel) => alarmLevelMapper.getTagType(level)
const getLevelText = (level: AlarmLevel) => alarmLevelMapper.getLabel(level)

// 获取条件文本
const getConditionText = (rule: AlarmRule): string => {
  const operatorMap: Record<string, string> = {
    '>': '大于',
    '>=': '大于等于',
    '<': '小于',
    '<=': '小于等于',
    '==': '等于',
    '!=': '不等于',
  }
  return `${operatorMap[rule.operator] || rule.operator} ${rule.threshold}`
}

// 格式化时间
const formatTime = (time: string): string => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

// 获取采集点列表
const fetchPointList = async () => {
  try {
    const result = await getPointList({ page: 1, pageSize: 1000 })
    pointList.value = result.list
  } catch (error) {
    console.error('获取采集点列表失败:', error)
  }
}

// 获取规则列表
const fetchRuleList = async () => {
  loading.value = true
  try {
    const result = await getAlarmRuleList({
      page: pagination.page,
      pageSize: pagination.pageSize,
    })
    ruleList.value = result.list.map((rule: ApiAlarmRule) => ({
      id: rule.id,
      name: rule.name,
      pointId: rule.point_id ? Number(rule.point_id) : undefined,
      pointName: rule.point_id || '-',
      operator: rule.condition.split(' ')[0] || '>',
      threshold: rule.threshold,
      duration: rule.duration,
      level: ['critical', 'major', 'minor', 'warning'][rule.level - 1] as AlarmLevel || 'warning',
      title: rule.name,
      content: rule.description,
      notifyChannels: rule.notify_channels || [],
      enabled: rule.status === 1,
      createdAt: rule.created_at,
    }))
    pagination.total = result.total
  } catch (error: any) {
    ElMessage.error(error.message || '获取规则列表失败')
  } finally {
    loading.value = false
  }
}

// 新增
const handleAdd = () => {
  isEdit.value = false
  Object.assign(ruleForm, {
    id: '',
    name: '',
    pointId: undefined,
    operator: '>',
    threshold: 0,
    duration: 0,
    level: 'warning',
    title: '',
    content: '',
    notifyChannels: [],
    enabled: true,
  })
  dialogVisible.value = true
}

// 编辑
const handleEdit = (row: AlarmRule) => {
  isEdit.value = true
  Object.assign(ruleForm, {
    id: row.id,
    name: row.name,
    pointId: row.pointId,
    operator: row.operator,
    threshold: row.threshold,
    duration: row.duration,
    level: row.level,
    title: row.title,
    content: row.content,
    notifyChannels: row.notifyChannels,
    enabled: row.enabled,
  })
  dialogVisible.value = true
}

// 删除
const handleDelete = async (row: AlarmRule) => {
  try {
    await ElMessageBox.confirm('确定删除该规则？', '提示', {
      type: 'warning',
    })
    await deleteAlarmRule(row.id)
    ElMessage.success('删除成功')
    fetchRuleList()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

// 状态变更
const handleStatusChange = async (row: AlarmRule) => {
  try {
    await updateAlarmRule(row.id, {
      id: row.id,
      name: row.name,
      description: row.content,
      type: 'threshold',
      level: ['critical', 'major', 'minor', 'warning'].indexOf(row.level) + 1,
      condition: row.operator,
      threshold: row.threshold,
      duration: row.duration,
      point_id: row.pointId?.toString(),
      notify_channels: row.notifyChannels,
      notify_users: [],
    })
    ElMessage.success(row.enabled ? '已启用' : '已禁用')
  } catch (error: any) {
    row.enabled = !row.enabled
    ElMessage.error(error.message || '操作失败')
  }
}

// 测试规则
const handleTest = (row: AlarmRule) => {
  // 模拟测试
  const currentValue = Math.random() * 100
  let success = false

  switch (row.operator) {
    case '>':
      success = currentValue > row.threshold
      break
    case '>=':
      success = currentValue >= row.threshold
      break
    case '<':
      success = currentValue < row.threshold
      break
    case '<=':
      success = currentValue <= row.threshold
      break
    case '==':
      success = currentValue === row.threshold
      break
    case '!=':
      success = currentValue !== row.threshold
      break
  }

  testResult.success = success
  testResult.currentValue = Number(currentValue.toFixed(2))
  testResult.condition = getConditionText(row)
  testResult.message = success ? '条件满足，会触发告警' : '条件不满足，不会触发告警'
  testVisible.value = true
}

// 提交
const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitLoading.value = true
    try {
      const apiData = {
        name: ruleForm.name,
        description: ruleForm.content,
        type: 'threshold',
        level: ['critical', 'major', 'minor', 'warning'].indexOf(ruleForm.level) + 1,
        condition: ruleForm.operator,
        threshold: ruleForm.threshold,
        duration: ruleForm.duration,
        point_id: ruleForm.pointId?.toString(),
        notify_channels: ruleForm.notifyChannels,
        notify_users: [],
      }

      if (isEdit.value) {
        await updateAlarmRule(ruleForm.id, { ...apiData, id: ruleForm.id })
        ElMessage.success('更新成功')
      } else {
        await createAlarmRule(apiData)
        ElMessage.success('创建成功')
      }

      dialogVisible.value = false
      fetchRuleList()
    } catch (error: any) {
      ElMessage.error(error.message || '操作失败')
    } finally {
      submitLoading.value = false
    }
  })
}

// 初始化
onMounted(() => {
  fetchPointList()
  fetchRuleList()
})
</script>

<style scoped lang="scss">
.alarm-rule-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .title {
      font-size: 16px;
      font-weight: 500;
    }
  }

  .pagination-container {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
}
</style>
