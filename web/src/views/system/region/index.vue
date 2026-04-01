<template>
  <div class="region-management">
    <el-row :gutter="20">
      <!-- 左侧树形结构 -->
      <el-col :span="8">
        <el-card class="tree-card">
          <template #header>
            <div class="card-header">
              <span>区域结构</span>
              <el-button type="primary" size="small" @click="handleAdd()">
                <el-icon><Plus /></el-icon>
                新增区域
              </el-button>
            </div>
          </template>

          <el-input
            v-model="filterText"
            placeholder="输入关键字筛选"
            clearable
            style="margin-bottom: 16px"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>

          <el-tree
            ref="treeRef"
            :data="regionTree"
            :props="treeProps"
            :filter-node-method="filterNode"
            :expand-on-click-node="false"
            :highlight-current="true"
            :default-expand-all="true"
            :draggable="true"
            :allow-drop="allowDrop"
            :allow-drag="allowDrag"
            node-key="id"
            @node-click="handleNodeClick"
            @node-drop="handleDrop"
          >
            <template #default="{ node, data }">
              <div class="tree-node">
                <span class="node-label">
                  <el-icon v-if="data.children && data.children.length > 0">
                    <Folder />
                  </el-icon>
                  <el-icon v-else>
                    <Document />
                  </el-icon>
                  {{ node.label }}
                </span>
                <span class="node-actions">
                  <el-button
                    type="primary"
                    link
                    size="small"
                    @click.stop="handleAdd(data)"
                  >
                    新增
                  </el-button>
                  <el-button
                    type="primary"
                    link
                    size="small"
                    @click.stop="handleEdit(data)"
                  >
                    编辑
                  </el-button>
                  <el-button
                    type="danger"
                    link
                    size="small"
                    @click.stop="handleDelete(data)"
                  >
                    删除
                  </el-button>
                </span>
              </div>
            </template>
          </el-tree>
        </el-card>
      </el-col>

      <!-- 右侧详情 -->
      <el-col :span="16">
        <el-card class="detail-card">
          <template #header>
            <div class="card-header">
              <span>区域详情</span>
            </div>
          </template>

          <template v-if="currentRegion">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="区域名称">
                {{ currentRegion.name }}
              </el-descriptions-item>
              <el-descriptions-item label="区域编码">
                {{ currentRegion.code }}
              </el-descriptions-item>
              <el-descriptions-item label="上级区域">
                {{ getParentName(currentRegion.parentId) }}
              </el-descriptions-item>
              <el-descriptions-item label="区域层级">
                {{ getLevelName(currentRegion.level) }}
              </el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="currentRegion.status === 1 ? 'success' : 'danger'">
                  {{ currentRegion.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="创建时间">
                {{ currentRegion.createdAt }}
              </el-descriptions-item>
              <el-descriptions-item label="更新时间">
                {{ currentRegion.updatedAt }}
              </el-descriptions-item>
              <el-descriptions-item label="描述" :span="2">
                {{ currentRegion.description || '-' }}
              </el-descriptions-item>
            </el-descriptions>

            <!-- 子区域统计 -->
            <div class="region-stats" style="margin-top: 20px">
              <h4>子区域统计</h4>
              <el-row :gutter="16">
                <el-col :span="6">
                  <el-statistic title="子区域数量" :value="childStats.total" />
                </el-col>
                <el-col :span="6">
                  <el-statistic title="启用数量" :value="childStats.enabled" />
                </el-col>
                <el-col :span="6">
                  <el-statistic title="禁用数量" :value="childStats.disabled" />
                </el-col>
                <el-col :span="6">
                  <el-statistic title="电站数量" :value="childStats.stationCount" />
                </el-col>
              </el-row>
            </div>
          </template>

          <el-empty v-else description="请选择区域查看详情" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 新增/编辑弹窗 -->
    <FormDialog
      v-model="dialogVisible"
      :mode="dialogMode"
      :title="dialogTitle"
      :fields="formFields"
      :data="formData"
      :rules="formRules"
      width="600px"
      @submit="handleSubmit"
    >
      <template #parentId="{ form }">
        <el-tree-select
          v-model="form.parentId"
          :data="regionTree"
          :props="{ label: 'name', value: 'id' }"
          placeholder="请选择上级区域"
          clearable
          check-strictly
          style="width: 100%"
        />
      </template>
    </FormDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Folder, Document } from '@element-plus/icons-vue'
import FormDialog from '@/components/FormDialog/index.vue'
import {
  getRegionTree,
  getRegionDetail,
  createRegion,
  updateRegion,
  deleteRegion,
  updateRegionStatus
} from '@/api/region'
import type { Region } from '@/types'
import type { FormRules } from 'element-plus'

const treeRef = ref()
const filterText = ref('')
const regionTree = ref<Region[]>([])
const currentRegion = ref<Region | null>(null)
const dialogVisible = ref(false)
const dialogMode = ref<'add' | 'edit'>('add')
const formData = ref<Partial<Region>>({})
const loading = ref(false)

const treeProps = {
  children: 'children',
  label: 'name'
}

// 对话框标题
const dialogTitle = computed(() => {
  return dialogMode.value === 'add' ? '新增区域' : '编辑区域'
})

// 表单字段
const formFields = computed(() => [
  {
    prop: 'parentId',
    label: '上级区域',
    type: 'select',
    span: 24
  },
  {
    prop: 'name',
    label: '区域名称',
    type: 'input',
    required: true,
    span: 12
  },
  {
    prop: 'code',
    label: '区域编码',
    type: 'input',
    required: true,
    span: 12
  },
  {
    prop: 'level',
    label: '区域层级',
    type: 'select',
    required: true,
    span: 12,
    options: [
      { label: '省级', value: 1 },
      { label: '市级', value: 2 },
      { label: '区/县级', value: 3 },
      { label: '乡镇/街道', value: 4 }
    ]
  },
  {
    prop: 'status',
    label: '状态',
    type: 'switch',
    span: 12,
    activeText: '启用',
    inactiveText: '禁用'
  },
  {
    prop: 'description',
    label: '描述',
    type: 'textarea',
    span: 24,
    rows: 3
  }
])

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入区域名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入区域编码', trigger: 'blur' },
    { pattern: /^[A-Z0-9_]+$/, message: '只能包含大写字母、数字和下划线', trigger: 'blur' }
  ],
  level: [
    { required: true, message: '请选择区域层级', trigger: 'change' }
  ]
}

// 子区域统计
const childStats = computed(() => {
  if (!currentRegion.value || !currentRegion.value.children) {
    return {
      total: 0,
      enabled: 0,
      disabled: 0,
      stationCount: 0
    }
  }

  const children = currentRegion.value.children
  return {
    total: children.length,
    enabled: children.filter(c => c.status === 1).length,
    disabled: children.filter(c => c.status === 0).length,
    stationCount: 0 // 需要从后端获取
  }
})

// 筛选节点
const filterNode = (value: string, data: Region) => {
  if (!value) return true
  return data.name.includes(value) || data.code.includes(value)
}

// 监听筛选文本
watch(filterText, (val) => {
  treeRef.value?.filter(val)
})

// 获取区域树
const fetchRegionTree = async () => {
  try {
    loading.value = true
    const data = await getRegionTree()
    regionTree.value = data
  } catch (error) {
    ElMessage.error('获取区域树失败')
  } finally {
    loading.value = false
  }
}

// 节点点击
const handleNodeClick = (data: Region) => {
  currentRegion.value = data
}

// 新增
const handleAdd = (parent?: Region) => {
  dialogMode.value = 'add'
  formData.value = {
    parentId: parent?.id || null,
    level: parent ? parent.level + 1 : 1,
    status: 1
  }
  dialogVisible.value = true
}

// 编辑
const handleEdit = (data: Region) => {
  dialogMode.value = 'edit'
  formData.value = { ...data }
  dialogVisible.value = true
}

// 删除
const handleDelete = async (data: Region) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除区域"${data.name}"吗？删除后将同时删除其子区域！`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await deleteRegion(data.id)
    ElMessage.success('删除成功')
    fetchRegionTree()
    
    if (currentRegion.value?.id === data.id) {
      currentRegion.value = null
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 提交
const handleSubmit = async (data: any) => {
  try {
    if (dialogMode.value === 'add') {
      await createRegion(data)
      ElMessage.success('新增成功')
    } else {
      await updateRegion(data.id, data)
      ElMessage.success('更新成功')
    }
    
    dialogVisible.value = false
    fetchRegionTree()
  } catch (error) {
    ElMessage.error(dialogMode.value === 'add' ? '新增失败' : '更新失败')
  }
}

// 拖拽判断
const allowDrop = (draggingNode: any, dropNode: any, type: string) => {
  // 不允许拖拽到子节点内部
  if (type === 'inner') {
    return false
  }
  return true
}

const allowDrag = (draggingNode: any) => {
  // 根节点不允许拖拽
  return draggingNode.data.parentId !== null
}

// 拖拽完成
const handleDrop = async (draggingNode: any, dropNode: any, dropType: string) => {
  try {
    // 更新父节点ID
    const dragData = draggingNode.data
    const newData = {
      ...dragData,
      parentId: dropType === 'inner' ? dropNode.data.id : dropNode.data.parentId
    }
    
    await updateRegion(dragData.id, newData)
    ElMessage.success('移动成功')
  } catch (error) {
    ElMessage.error('移动失败')
    fetchRegionTree()
  }
}

// 获取父区域名称
const getParentName = (parentId: number | null) => {
  if (!parentId) return '无'
  
  const findParent = (nodes: Region[]): string => {
    for (const node of nodes) {
      if (node.id === parentId) {
        return node.name
      }
      if (node.children) {
        const result = findParent(node.children)
        if (result) return result
      }
    }
    return ''
  }
  
  return findParent(regionTree.value) || '未知'
}

// 获取层级名称
const getLevelName = (level: number) => {
  const levelMap: Record<number, string> = {
    1: '省级',
    2: '市级',
    3: '区/县级',
    4: '乡镇/街道'
  }
  return levelMap[level] || '未知'
}

onMounted(() => {
  fetchRegionTree()
})
</script>

<style scoped lang="scss">
.region-management {
  height: 100%;

  .el-row {
    height: 100%;
  }

  .tree-card,
  .detail-card {
    height: 100%;
    overflow: hidden;

    :deep(.el-card__body) {
      height: calc(100% - 60px);
      overflow-y: auto;
    }
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .tree-node {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 14px;
    padding-right: 8px;

    .node-label {
      display: flex;
      align-items: center;
      gap: 5px;
    }

    .node-actions {
      display: none;
    }

    &:hover .node-actions {
      display: inline-flex;
      gap: 5px;
    }
  }

  .region-stats {
    h4 {
      margin-bottom: 16px;
      font-size: 16px;
      font-weight: 500;
    }
  }
}
</style>
