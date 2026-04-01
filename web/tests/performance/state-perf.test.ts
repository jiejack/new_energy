import { describe, it, expect, benchmark, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useUserStore } from '@/stores/user'
import { usePermissionStore } from '@/stores/permission'
import { useAppStore } from '@/stores/app'
import { ref, computed, nextTick } from 'vue'

describe('State Update Performance', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  // Store更新性能测试
  describe('Store Update Performance', () => {
    it('should efficiently update user store', async () => {
      const userStore = useUserStore()
      
      await benchmark(async () => {
        userStore.setUser({
          id: '1',
          username: 'testuser',
          email: 'test@example.com',
          roles: ['admin'],
        })
        await nextTick()
      }, { iterations: 100 })
    })

    it('should efficiently update permission store', async () => {
      const permissionStore = usePermissionStore()
      const permissions = generatePermissions(100)
      
      await benchmark(async () => {
        permissionStore.setPermissions(permissions)
        await nextTick()
      }, { iterations: 50 })
    })

    it('should efficiently update app store', async () => {
      const appStore = useAppStore()
      
      await benchmark(async () => {
        appStore.setLoading(true)
        await nextTick()
        appStore.setLoading(false)
        await nextTick()
      }, { iterations: 200 })
    })

    it('should efficiently handle multiple store updates', async () => {
      const userStore = useUserStore()
      const permissionStore = usePermissionStore()
      const appStore = useAppStore()
      
      await benchmark(async () => {
        userStore.setUser({
          id: '1',
          username: 'testuser',
          email: 'test@example.com',
          roles: ['admin'],
        })
        permissionStore.setPermissions(generatePermissions(50))
        appStore.setLoading(true)
        await nextTick()
      }, { iterations: 30 })
    })
  })

  // 大量数据更新性能测试
  describe('Large Data Update Performance', () => {
    it('should efficiently update large array in store', async () => {
      const store = createTestStore()
      const largeArray = generateLargeArray(10000)
      
      await benchmark(async () => {
        store.setItems(largeArray)
        await nextTick()
      }, { iterations: 10 })
    })

    it('should efficiently add items to large array', async () => {
      const store = createTestStore()
      store.setItems(generateLargeArray(5000))
      
      await benchmark(async () => {
        const newItem = { id: Date.now(), name: 'New Item' }
        store.addItem(newItem)
        await nextTick()
      }, { iterations: 100 })
    })

    it('should efficiently remove items from large array', async () => {
      const store = createTestStore()
      const items = generateLargeArray(5000)
      store.setItems(items)
      
      await benchmark(async () => {
        const index = Math.floor(Math.random() * store.items.length)
        store.removeItem(index)
        await nextTick()
      }, { iterations: 100 })
    })

    it('should efficiently update item in large array', async () => {
      const store = createTestStore()
      store.setItems(generateLargeArray(5000))
      
      await benchmark(async () => {
        const index = Math.floor(Math.random() * store.items.length)
        store.updateItem(index, { name: 'Updated' })
        await nextTick()
      }, { iterations: 100 })
    })

    it('should efficiently filter large array', async () => {
      const store = createTestStore()
      store.setItems(generateLargeArray(10000))
      
      await benchmark(async () => {
        const filtered = store.items.filter(item => item.status === 'active')
        await nextTick()
      }, { iterations: 20 })
    })

    it('should efficiently sort large array', async () => {
      const store = createTestStore()
      store.setItems(generateLargeArray(10000))
      
      await benchmark(async () => {
        const sorted = [...store.items].sort((a, b) => a.id - b.id)
        await nextTick()
      }, { iterations: 10 })
    })
  })

  // 计算属性性能测试
  describe('Computed Property Performance', () => {
    it('should efficiently compute filtered data', async () => {
      const store = createTestStore()
      store.setItems(generateLargeArray(5000))
      
      const filtered = computed(() => {
        return store.items.filter(item => item.status === 'active')
      })
      
      await benchmark(async () => {
        // 触发重新计算
        store.setItems(generateLargeArray(5000))
        await nextTick()
        const _ = filtered.value
      }, { iterations: 20 })
    })

    it('should efficiently compute aggregated data', async () => {
      const store = createTestStore()
      store.setItems(generateLargeArray(5000))
      
      const aggregated = computed(() => {
        const result = {
          total: store.items.length,
          active: store.items.filter(i => i.status === 'active').length,
          inactive: store.items.filter(i => i.status === 'inactive').length,
        }
        return result
      })
      
      await benchmark(async () => {
        store.setItems(generateLargeArray(5000))
        await nextTick()
        const _ = aggregated.value
      }, { iterations: 15 })
    })

    it('should efficiently compute nested data', async () => {
      const store = createNestedStore()
      store.setNestedData(generateNestedData(100, 10))
      
      const flattened = computed(() => {
        const result: any[] = []
        store.nestedData.forEach(parent => {
          parent.children.forEach(child => {
            result.push({ ...child, parentId: parent.id })
          })
        })
        return result
      })
      
      await benchmark(async () => {
        store.setNestedData(generateNestedData(100, 10))
        await nextTick()
        const _ = flattened.value
      }, { iterations: 10 })
    })
  })

  // 批量更新性能测试
  describe('Batch Update Performance', () => {
    it('should efficiently handle batch updates', async () => {
      const store = createTestStore()
      const updates = generateUpdates(1000)
      
      await benchmark(async () => {
        store.batchUpdate(updates)
        await nextTick()
      }, { iterations: 10 })
    })

    it('should efficiently handle batch delete', async () => {
      const store = createTestStore()
      store.setItems(generateLargeArray(10000))
      const ids = store.items.slice(0, 1000).map(i => i.id)
      
      await benchmark(async () => {
        store.batchDelete(ids)
        await nextTick()
      }, { iterations: 10 })
    })

    it('should efficiently handle batch insert', async () => {
      const store = createTestStore()
      const newItems = generateLargeArray(1000)
      
      await benchmark(async () => {
        store.batchInsert(newItems)
        await nextTick()
      }, { iterations: 10 })
    })
  })

  // 状态订阅性能测试
  describe('State Subscription Performance', () => {
    it('should efficiently handle multiple subscribers', async () => {
      const store = createTestStore()
      const callbacks: Array<() => void> = []
      
      // 添加多个订阅者
      for (let i = 0; i < 100; i++) {
        callbacks.push(() => {
          const _ = store.items.length
        })
      }
      
      await benchmark(async () => {
        store.setItems(generateLargeArray(100))
        await nextTick()
        callbacks.forEach(cb => cb())
      }, { iterations: 20 })
    })

    it('should efficiently handle deep watch', async () => {
      const store = createTestStore()
      store.setItems(generateLargeArray(1000))
      
      await benchmark(async () => {
        // 深度更新
        if (store.items[0]) {
          store.items[0].name = `Updated ${Date.now()}`
        }
        await nextTick()
      }, { iterations: 50 })
    })
  })

  // 状态持久化性能测试
  describe('State Persistence Performance', () => {
    it('should efficiently save state to localStorage', async () => {
      const store = createTestStore()
      store.setItems(generateLargeArray(1000))
      
      await benchmark(async () => {
        const state = JSON.stringify(store.$state)
        localStorage.setItem('test-store', state)
      }, { iterations: 10 })
      
      localStorage.removeItem('test-store')
    })

    it('should efficiently load state from localStorage', async () => {
      const store = createTestStore()
      const state = JSON.stringify(generateLargeArray(1000))
      localStorage.setItem('test-store', state)
      
      await benchmark(async () => {
        const loaded = localStorage.getItem('test-store')
        if (loaded) {
          store.$patch(JSON.parse(loaded))
        }
      }, { iterations: 10 })
      
      localStorage.removeItem('test-store')
    })

    it('should efficiently serialize large state', async () => {
      const store = createTestStore()
      store.setItems(generateLargeArray(10000))
      
      await benchmark(async () => {
        const serialized = JSON.stringify(store.$state)
        const _ = serialized.length
      }, { iterations: 5 })
    })
  })

  // 并发更新性能测试
  describe('Concurrent Update Performance', () => {
    it('should efficiently handle concurrent updates', async () => {
      const store = createTestStore()
      store.setItems(generateLargeArray(1000))
      
      await benchmark(async () => {
        // 模拟并发更新
        const updates = Array.from({ length: 10 }, (_, i) => ({
          index: i * 100,
          data: { name: `Updated ${i}` },
        }))
        
        updates.forEach(update => {
          store.updateItem(update.index, update.data)
        })
        
        await nextTick()
      }, { iterations: 20 })
    })

    it('should efficiently handle rapid state changes', async () => {
      const store = createTestStore()
      
      await benchmark(async () => {
        for (let i = 0; i < 100; i++) {
          store.addItem({ id: i, name: `Item ${i}`, status: 'active' })
        }
        await nextTick()
      }, { iterations: 10 })
    })
  })

  // 状态快照性能测试
  describe('State Snapshot Performance', () => {
    it('should efficiently create state snapshot', async () => {
      const store = createTestStore()
      store.setItems(generateLargeArray(5000))
      
      await benchmark(async () => {
        const snapshot = JSON.parse(JSON.stringify(store.$state))
        const _ = snapshot
      }, { iterations: 10 })
    })

    it('should efficiently restore from snapshot', async () => {
      const store = createTestStore()
      store.setItems(generateLargeArray(5000))
      const snapshot = JSON.parse(JSON.stringify(store.$state))
      
      await benchmark(async () => {
        store.$patch(JSON.parse(JSON.stringify(snapshot)))
        await nextTick()
      }, { iterations: 10 })
    })
  })

  // Undo/Redo性能测试
  describe('Undo/Redo Performance', () => {
    it('should efficiently track state history', async () => {
      const store = createHistoryStore()
      
      await benchmark(async () => {
        store.addItem({ id: Date.now(), name: 'New Item' })
        await nextTick()
      }, { iterations: 100 })
    })

    it('should efficiently undo state changes', async () => {
      const store = createHistoryStore()
      
      // 创建历史记录
      for (let i = 0; i < 50; i++) {
        store.addItem({ id: i, name: `Item ${i}` })
      }
      
      await benchmark(async () => {
        store.undo()
        await nextTick()
      }, { iterations: 20 })
    })

    it('should efficiently redo state changes', async () => {
      const store = createHistoryStore()
      
      // 创建历史记录
      for (let i = 0; i < 50; i++) {
        store.addItem({ id: i, name: `Item ${i}` })
      }
      
      // 撤销
      store.undo()
      
      await benchmark(async () => {
        store.redo()
        await nextTick()
      }, { iterations: 20 })
    })
  })
})

// 辅助函数和类型

function generatePermissions(count: number): string[] {
  return Array.from({ length: count }, (_, i) => `permission_${i}`)
}

function generateLargeArray(count: number): any[] {
  return Array.from({ length: count }, (_, i) => ({
    id: i,
    name: `Item ${i}`,
    status: ['active', 'inactive'][i % 2],
    value: Math.random() * 100,
    timestamp: Date.now() - i * 1000,
  }))
}

function generateUpdates(count: number): any[] {
  return Array.from({ length: count }, (_, i) => ({
    index: i,
    data: { name: `Updated ${i}` },
  }))
}

function generateNestedData(parentCount: number, childCount: number): any[] {
  return Array.from({ length: parentCount }, (_, i) => ({
    id: `parent_${i}`,
    name: `Parent ${i}`,
    children: Array.from({ length: childCount }, (_, j) => ({
      id: `child_${i}_${j}`,
      name: `Child ${i}-${j}`,
      value: Math.random() * 100,
    })),
  }))
}

// 创建测试Store
function createTestStore() {
  const pinia = createPinia()
  setActivePinia(pinia)
  
  return defineStore('test', {
    state: () => ({
      items: [] as any[],
    }),
    actions: {
      setItems(items: any[]) {
        this.items = items
      },
      addItem(item: any) {
        this.items.push(item)
      },
      removeItem(index: number) {
        this.items.splice(index, 1)
      },
      updateItem(index: number, data: any) {
        if (this.items[index]) {
          Object.assign(this.items[index], data)
        }
      },
      batchUpdate(updates: any[]) {
        updates.forEach(update => {
          this.updateItem(update.index, update.data)
        })
      },
      batchDelete(ids: number[]) {
        this.items = this.items.filter(item => !ids.includes(item.id))
      },
      batchInsert(items: any[]) {
        this.items.push(...items)
      },
    },
  })()
}

function createNestedStore() {
  const pinia = createPinia()
  setActivePinia(pinia)
  
  return defineStore('nested', {
    state: () => ({
      nestedData: [] as any[],
    }),
    actions: {
      setNestedData(data: any[]) {
        this.nestedData = data
      },
    },
  })()
}

function createHistoryStore() {
  const pinia = createPinia()
  setActivePinia(pinia)
  
  return defineStore('history', {
    state: () => ({
      items: [] as any[],
      history: [] as any[],
      historyIndex: -1,
    }),
    actions: {
      addItem(item: any) {
        // 保存当前状态到历史
        this.history = this.history.slice(0, this.historyIndex + 1)
        this.history.push(JSON.stringify(this.items))
        this.historyIndex++
        
        this.items.push(item)
      },
      undo() {
        if (this.historyIndex >= 0) {
          this.items = JSON.parse(this.history[this.historyIndex])
          this.historyIndex--
        }
      },
      redo() {
        if (this.historyIndex < this.history.length - 1) {
          this.historyIndex++
          this.items = JSON.parse(this.history[this.historyIndex])
        }
      },
    },
  })()
}

// 导入defineStore
import { defineStore } from 'pinia'
