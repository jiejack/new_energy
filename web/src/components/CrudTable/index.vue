<template>
  <div class="crud-table">
    <!-- 工具栏 -->
    <div class="table-toolbar" v-if="showToolbar">
      <div class="toolbar-left">
        <slot name="toolbar-left"></slot>
      </div>
      <div class="toolbar-right">
        <el-input
          v-if="showSearch"
          v-model="searchKeyword"
          :placeholder="searchPlaceholder"
          clearable
          @clear="handleSearch"
          @keyup.enter="handleSearch"
          style="width: 250px; margin-right: 10px"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button v-if="showSearch" type="primary" @click="handleSearch">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
        <el-button v-if="showRefresh" @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button v-if="showExport" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出
        </el-button>
        <el-dropdown v-if="showColumnConfig" trigger="click">
          <el-button>
            <el-icon><Setting /></el-icon>
            列配置
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-checkbox-group v-model="visibleColumns">
                <el-dropdown-item
                  v-for="col in columns"
                  :key="col.prop"
                  :label="col.prop"
                >
                  <el-checkbox :label="col.prop">{{ col.label }}</el-checkbox>
                </el-dropdown-item>
              </el-checkbox-group>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <slot name="toolbar-right"></slot>
      </div>
    </div>

    <!-- 批量操作栏 -->
    <div class="batch-actions" v-if="selectedRows.length > 0">
      <el-alert
        :title="`已选择 ${selectedRows.length} 项`"
        type="info"
        show-icon
        :closable="false"
      >
        <template #default>
          <el-button
            type="danger"
            size="small"
            @click="handleBatchDelete"
            v-if="showBatchDelete"
          >
            批量删除
          </el-button>
          <slot name="batch-actions" :rows="selectedRows"></slot>
        </template>
      </el-alert>
    </div>

    <!-- 表格 -->
    <el-table
      ref="tableRef"
      :data="tableData"
      :border="border"
      :stripe="stripe"
      :size="size"
      :height="height"
      :max-height="maxHeight"
      :row-key="rowKey"
      :default-expand-all="defaultExpandAll"
      :highlight-current-row="highlightCurrentRow"
      :show-summary="showSummary"
      :summary-method="summaryMethod"
      :span-method="spanMethod"
      :lazy="lazy"
      :load="load"
      :tree-props="treeProps"
      @selection-change="handleSelectionChange"
      @sort-change="handleSortChange"
      @row-click="handleRowClick"
      @row-dblclick="handleRowDblclick"
      v-loading="loading"
    >
      <!-- 选择列 -->
      <el-table-column
        v-if="showSelection"
        type="selection"
        width="55"
        align="center"
        :reserve-selection="reserveSelection"
      />

      <!-- 序号列 -->
      <el-table-column
        v-if="showIndex"
        type="index"
        label="序号"
        width="60"
        align="center"
        :index="indexMethod"
      />

      <!-- 数据列 -->
      <template v-for="col in visibleColumnsList" :key="col.prop">
        <el-table-column
          :prop="col.prop"
          :label="col.label"
          :width="col.width"
          :min-width="col.minWidth"
          :fixed="col.fixed"
          :sortable="col.sortable"
          :align="col.align || 'left'"
          :header-align="col.headerAlign || 'center'"
          :show-overflow-tooltip="col.showOverflowTooltip !== false"
        >
          <template #default="scope">
            <slot
              v-if="$slots[col.prop]"
              :name="col.prop"
              :row="scope.row"
              :column="col"
              :$index="scope.$index"
            ></slot>
            <template v-else>
              <template v-if="col.type === 'image'">
                <el-image
                  :src="scope.row[col.prop]"
                  :preview-src-list="[scope.row[col.prop]]"
                  style="width: 50px; height: 50px"
                  fit="cover"
                />
              </template>
              <template v-else-if="col.type === 'tag'">
                <el-tag
                  :type="getTagType(scope.row[col.prop], col.tagMap)"
                  size="small"
                >
                  {{ formatValue(scope.row[col.prop], col) }}
                </el-tag>
              </template>
              <template v-else-if="col.type === 'date'">
                {{ formatDate(scope.row[col.prop], col.dateFormat) }}
              </template>
              <template v-else-if="col.type === 'link'">
                <el-link type="primary" @click="col.onClick?.(scope.row)">
                  {{ scope.row[col.prop] }}
                </el-link>
              </template>
              <template v-else>
                {{ formatValue(scope.row[col.prop], col) }}
              </template>
            </template>
          </template>
        </el-table-column>
      </template>

      <!-- 操作列 -->
      <el-table-column
        v-if="showActions"
        label="操作"
        :width="actionsWidth"
        :fixed="actionsFixed"
        align="center"
      >
        <template #default="scope">
          <slot name="actions" :row="scope.row" :$index="scope.$index">
            <el-button
              v-if="showView"
              type="primary"
              link
              size="small"
              @click="handleView(scope.row)"
            >
              查看
            </el-button>
            <el-button
              v-if="showEdit"
              type="primary"
              link
              size="small"
              @click="handleEdit(scope.row)"
            >
              编辑
            </el-button>
            <el-button
              v-if="showDelete"
              type="danger"
              link
              size="small"
              @click="handleDelete(scope.row)"
            >
              删除
            </el-button>
          </slot>
        </template>
      </el-table-column>

      <!-- 空状态 -->
      <template #empty>
        <slot name="empty">
          <el-empty description="暂无数据" />
        </slot>
      </template>
    </el-table>

    <!-- 分页 -->
    <div class="table-pagination" v-if="showPagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="pageSizes"
        :total="total"
        :layout="paginationLayout"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Download, Setting } from '@element-plus/icons-vue'
import type { TableInstance } from 'element-plus'

export interface Column {
  prop: string
  label: string
  width?: number | string
  minWidth?: number | string
  fixed?: boolean | 'left' | 'right'
  sortable?: boolean | 'custom'
  align?: 'left' | 'center' | 'right'
  headerAlign?: 'left' | 'center' | 'right'
  showOverflowTooltip?: boolean
  type?: 'text' | 'image' | 'tag' | 'date' | 'link'
  format?: (value: any, row: any) => string
  tagMap?: Record<string, string>
  dateFormat?: string
  onClick?: (row: any) => void
  hidden?: boolean
}

export interface Props {
  data?: any[]
  columns: Column[]
  loading?: boolean
  border?: boolean
  stripe?: boolean
  size?: 'large' | 'default' | 'small'
  height?: number | string
  maxHeight?: number | string
  rowKey?: string | ((row: any) => string)
  defaultExpandAll?: boolean
  highlightCurrentRow?: boolean
  showSummary?: boolean
  summaryMethod?: (data: { columns: any[]; data: any[] }) => any[]
  spanMethod?: (data: { row: any; column: any; rowIndex: number; columnIndex: number }) => any
  lazy?: boolean
  load?: (row: any, treeNode: any, resolve: (data: any[]) => void) => void
  treeProps?: { hasChildren?: string; children?: string }
  showToolbar?: boolean
  showSearch?: boolean
  searchPlaceholder?: string
  showRefresh?: boolean
  showExport?: boolean
  showColumnConfig?: boolean
  showSelection?: boolean
  reserveSelection?: boolean
  showIndex?: boolean
  showActions?: boolean
  showView?: boolean
  showEdit?: boolean
  showDelete?: boolean
  showBatchDelete?: boolean
  actionsWidth?: number | string
  actionsFixed?: boolean | 'left' | 'right'
  showPagination?: boolean
  total?: number
  page?: number
  limit?: number
  pageSizes?: number[]
  paginationLayout?: string
  requestApi?: (params: any) => Promise<any>
  deleteApi?: (id: number) => Promise<void>
  batchDeleteApi?: (ids: number[]) => Promise<void>
  exportApi?: (params: any) => Promise<Blob>
}

const props = withDefaults(defineProps<Props>(), {
  data: undefined,
  loading: false,
  border: true,
  stripe: true,
  size: 'default',
  showToolbar: true,
  showSearch: true,
  searchPlaceholder: '请输入关键词搜索',
  showRefresh: true,
  showExport: false,
  showColumnConfig: true,
  showSelection: true,
  reserveSelection: false,
  showIndex: true,
  showActions: true,
  showView: true,
  showEdit: true,
  showDelete: true,
  showBatchDelete: true,
  actionsWidth: 200,
  actionsFixed: 'right',
  showPagination: true,
  total: 0,
  page: 1,
  limit: 20,
  pageSizes: () => [10, 20, 50, 100],
  paginationLayout: 'total, sizes, prev, pager, next, jumper'
})

const emit = defineEmits<{
  'update:page': [page: number]
  'update:limit': [limit: number]
  'search': [keyword: string]
  'refresh': []
  'export': []
  'view': [row: any]
  'edit': [row: any]
  'delete': [row: any]
  'batch-delete': [rows: any[]]
  'selection-change': [rows: any[]]
  'sort-change': [sort: { prop: string; order: string }]
  'row-click': [row: any]
  'row-dblclick': [row: any]
}>()

const tableRef = ref<TableInstance>()
const tableData = ref<any[]>([])
const searchKeyword = ref('')
const selectedRows = ref<any[]>([])
const visibleColumns = ref<string[]>([])
const currentPage = ref(props.page)
const pageSize = ref(props.limit)
const sortBy = ref('')
const sortOrder = ref('')

// 计算可见列
const visibleColumnsList = computed(() => {
  return props.columns.filter(col => 
    !col.hidden && visibleColumns.value.includes(col.prop)
  )
})

// 初始化可见列
onMounted(() => {
  visibleColumns.value = props.columns
    .filter(col => !col.hidden)
    .map(col => col.prop)
  
  if (props.requestApi) {
    fetchData()
  } else if (props.data) {
    tableData.value = props.data
  }
})

// 监听data变化
watch(() => props.data, (val) => {
  if (val) {
    tableData.value = val
  }
}, { immediate: true, deep: true })

// 监听page和limit变化
watch(() => props.page, (val) => {
  currentPage.value = val
})

watch(() => props.limit, (val) => {
  pageSize.value = val
})

// 获取数据
const fetchData = async () => {
  if (!props.requestApi) return

  try {
    const params: any = {
      page: currentPage.value,
      pageSize: pageSize.value,
      keyword: searchKeyword.value
    }

    if (sortBy.value && sortOrder.value) {
      params.sortBy = sortBy.value
      params.sortOrder = sortOrder.value
    }

    const result = await props.requestApi(params)
    tableData.value = result.list || result
    
    if (result.total !== undefined) {
      emit('update:page' as any, currentPage.value)
      emit('update:limit' as any, pageSize.value)
    }
  } catch (error) {
    console.error('获取数据失败:', error)
    ElMessage.error('获取数据失败')
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  emit('search', searchKeyword.value)
  if (props.requestApi) {
    fetchData()
  }
}

// 刷新
const handleRefresh = () => {
  emit('refresh')
  if (props.requestApi) {
    fetchData()
  }
}

// 导出
const handleExport = async () => {
  emit('export')
  
  if (props.exportApi) {
    try {
      const blob = await props.exportApi({
        keyword: searchKeyword.value
      })
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `export_${Date.now()}.xlsx`
      a.click()
      window.URL.revokeObjectURL(url)
      ElMessage.success('导出成功')
    } catch (error) {
      ElMessage.error('导出失败')
    }
  }
}

// 选择变化
const handleSelectionChange = (rows: any[]) => {
  selectedRows.value = rows
  emit('selection-change', rows)
}

// 排序变化
const handleSortChange = ({ prop, order }: { prop: string; order: string | null }) => {
  sortBy.value = prop
  sortOrder.value = order === 'ascending' ? 'asc' : order === 'descending' ? 'desc' : ''
  emit('sort-change', { prop, order: sortOrder.value })
  if (props.requestApi) {
    fetchData()
  }
}

// 行点击
const handleRowClick = (row: any) => {
  emit('row-click', row)
}

// 行双击
const handleRowDblclick = (row: any) => {
  emit('row-dblclick', row)
}

// 查看
const handleView = (row: any) => {
  emit('view', row)
}

// 编辑
const handleEdit = (row: any) => {
  emit('edit', row)
}

// 删除
const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm('确定要删除该记录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    if (props.deleteApi) {
      await props.deleteApi(row.id)
      ElMessage.success('删除成功')
      fetchData()
    }
    
    emit('delete', row)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 批量删除
const handleBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedRows.value.length} 条记录吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const ids = selectedRows.value.map(row => row.id)
    
    if (props.batchDeleteApi) {
      await props.batchDeleteApi(ids)
      ElMessage.success('批量删除成功')
      selectedRows.value = []
      fetchData()
    }
    
    emit('batch-delete', selectedRows.value)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

// 分页大小变化
const handleSizeChange = (val: number) => {
  pageSize.value = val
  emit('update:limit', val)
  if (props.requestApi) {
    fetchData()
  }
}

// 当前页变化
const handleCurrentChange = (val: number) => {
  currentPage.value = val
  emit('update:page', val)
  if (props.requestApi) {
    fetchData()
  }
}

// 序号方法
const indexMethod = (index: number) => {
  return (currentPage.value - 1) * pageSize.value + index + 1
}

// 格式化值
const formatValue = (value: any, col: Column) => {
  if (col.format) {
    return col.format(value, value)
  }
  if (col.tagMap && value in col.tagMap) {
    return col.tagMap[value]
  }
  return value ?? '-'
}

// 获取标签类型
const getTagType = (value: any, tagMap?: Record<string, string>) => {
  if (!tagMap) return ''
  
  const typeMap: Record<string, any> = {
    success: 'success',
    warning: 'warning',
    danger: 'danger',
    info: 'info',
    primary: 'primary'
  }
  
  return typeMap[tagMap[value]] || ''
}

// 格式化日期
const formatDate = (value: string, format?: string) => {
  if (!value) return '-'
  
  const date = new Date(value)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  
  if (format === 'date') {
    return `${year}-${month}-${day}`
  }
  
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// 暴露方法
defineExpose({
  refresh: fetchData,
  getSelectionRows: () => selectedRows.value,
  clearSelection: () => {
    tableRef.value?.clearSelection()
    selectedRows.value = []
  }
})
</script>

<style scoped lang="scss">
.crud-table {
  .table-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    .toolbar-left {
      display: flex;
      gap: 10px;
    }

    .toolbar-right {
      display: flex;
      align-items: center;
      gap: 10px;
    }
  }

  .batch-actions {
    margin-bottom: 16px;

    :deep(.el-alert__content) {
      display: flex;
      align-items: center;
      gap: 10px;
    }
  }

  .table-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
}
</style>
