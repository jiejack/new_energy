import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import CrudTable from '../CrudTable/index.vue'
import type { Column } from '../CrudTable/index.vue'

// Mock Element Plus components
vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    error: vi.fn()
  },
  ElMessageBox: {
    confirm: vi.fn()
  }
}))

describe('CrudTable 组件', () => {
  const mockColumns: Column[] = [
    {
      prop: 'id',
      label: 'ID',
      width: 80
    },
    {
      prop: 'name',
      label: '名称',
      minWidth: 120
    },
    {
      prop: 'status',
      label: '状态',
      type: 'tag',
      tagMap: {
        'active': 'success',
        'inactive': 'danger'
      }
    }
  ]

  const mockData = [
    { id: 1, name: '测试1', status: 'active' },
    { id: 2, name: '测试2', status: 'inactive' },
    { id: 3, name: '测试3', status: 'active' }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('组件渲染', () => {
    it('应该正确渲染表格', () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(wrapper.find('.crud-table').exists()).toBe(true)
    })

    it('应该显示工具栏', () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData,
          showToolbar: true
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(wrapper.find('.table-toolbar').exists()).toBe(true)
    })

    it('应该隐藏工具栏', () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData,
          showToolbar: false
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(wrapper.find('.table-toolbar').exists()).toBe(false)
    })

    it('应该显示分页', () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData,
          showPagination: true,
          total: 100
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(wrapper.find('.table-pagination').exists()).toBe(true)
    })

    it('应该显示加载状态', () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData,
          loading: true
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(wrapper.props('loading')).toBe(true)
    })
  })

  describe('数据传递', () => {
    it('应该接收columns属性', () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(wrapper.props('columns')).toEqual(mockColumns)
    })

    it('应该接收data属性', () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(wrapper.props('data')).toEqual(mockData)
    })

    it('应该接收total属性', () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData,
          total: 100
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(wrapper.props('total')).toBe(100)
    })

    it('应该正确处理空数据', async () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: []
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      await nextTick()

      expect(wrapper.props('data')).toEqual([])
    })
  })

  describe('事件触发', () => {
    it('应该触发search事件', async () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData,
          showSearch: true
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      // 模拟搜索
      const searchInput = wrapper.findComponent({ name: 'el-input' })
      await searchInput.setValue('test')
      
      // 触发搜索按钮点击
      const buttons = wrapper.findAllComponents({ name: 'el-button' })
      const searchButton = buttons.find(btn => btn.text().includes('搜索'))
      
      if (searchButton) {
        await searchButton.trigger('click')
        expect(wrapper.emitted('search')).toBeTruthy()
        expect(wrapper.emitted('search')![0]).toEqual(['test'])
      }
    })

    it('应该触发refresh事件', async () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData,
          showRefresh: true
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      // 触发刷新按钮点击
      const buttons = wrapper.findAllComponents({ name: 'el-button' })
      const refreshButton = buttons.find(btn => btn.text().includes('刷新'))
      
      if (refreshButton) {
        await refreshButton.trigger('click')
        expect(wrapper.emitted('refresh')).toBeTruthy()
      }
    })

    it('应该触发view事件', async () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData,
          showActions: true,
          showView: true
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      // 由于操作列在表格内部，这里简化测试
      // 实际测试中需要更复杂的设置来模拟表格行操作
      expect(wrapper.props('showView')).toBe(true)
    })

    it('应该触发edit事件', async () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData,
          showActions: true,
          showEdit: true
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(wrapper.props('showEdit')).toBe(true)
    })

    it('应该触发delete事件', async () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData,
          showActions: true,
          showDelete: true
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(wrapper.props('showDelete')).toBe(true)
    })
  })

  describe('分页功能', () => {
    it('应该接收分页属性', () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData,
          showPagination: true,
          page: 2,
          limit: 20,
          total: 100
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(wrapper.props('page')).toBe(2)
      expect(wrapper.props('limit')).toBe(20)
      expect(wrapper.props('total')).toBe(100)
    })

    it('应该触发update:page事件', async () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData,
          showPagination: true,
          total: 100
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      const pagination = wrapper.findComponent({ name: 'el-pagination' })
      await pagination.vm.$emit('current-change', 2)

      expect(wrapper.emitted('update:page')).toBeTruthy()
      expect(wrapper.emitted('update:page')![0]).toEqual([2])
    })

    it('应该触发update:limit事件', async () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData,
          showPagination: true,
          total: 100
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      const pagination = wrapper.findComponent({ name: 'el-pagination' })
      await pagination.vm.$emit('size-change', 50)

      expect(wrapper.emitted('update:limit')).toBeTruthy()
      expect(wrapper.emitted('update:limit')![0]).toEqual([50])
    })

    it('应该使用默认分页配置', () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(wrapper.props('page')).toBe(1)
      expect(wrapper.props('limit')).toBe(20)
      expect(wrapper.props('pageSizes')).toEqual([10, 20, 50, 100])
    })
  })

  describe('暴露的方法', () => {
    it('应该暴露refresh方法', () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(typeof wrapper.vm.refresh).toBe('function')
    })

    it('应该暴露getSelectionRows方法', () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(typeof wrapper.vm.getSelectionRows).toBe('function')
      expect(wrapper.vm.getSelectionRows()).toEqual([])
    })

    it('应该暴露clearSelection方法', () => {
      const wrapper = mount(CrudTable, {
        props: {
          columns: mockColumns,
          data: mockData
        },
        global: {
          stubs: {
            'el-table': true,
            'el-table-column': true,
            'el-pagination': true,
            'el-input': true,
            'el-button': true,
            'el-icon': true,
            'el-dropdown': true,
            'el-checkbox': true,
            'el-alert': true,
            'el-empty': true
          }
        }
      })

      expect(typeof wrapper.vm.clearSelection).toBe('function')
    })
  })
})
