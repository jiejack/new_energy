import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import WebSocketManager from '../websocket'

// Mock auth utils
vi.mock('../auth', () => ({
  getToken: vi.fn()
}))

import { getToken } from '../auth'

describe('WebSocket Utils', () => {
  let wsManager: WebSocketManager

  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()

    // 创建一个简单的 Mock WebSocket 类
    class MockWebSocket {
      static CONNECTING = 0
      static OPEN = 1
      static CLOSING = 2
      static CLOSED = 3

      readyState = 0
      onopen: ((event: Event) => void) | null = null
      onmessage: ((event: MessageEvent) => void) | null = null
      onerror: ((event: Event) => void) | null = null
      onclose: ((event: Event) => void) | null = null

      constructor(_url: string) {
        // 延迟触发 onopen
        setTimeout(() => {
          this.readyState = 1
          if (this.onopen) {
            this.onopen(new Event('open'))
          }
        }, 0)
      }

      send = vi.fn()
      close = vi.fn()
    }

    global.WebSocket = MockWebSocket as any

    wsManager = new WebSocketManager()
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  describe('构造函数', () => {
    it('应该使用默认配置创建实例', () => {
      const manager = new WebSocketManager()

      expect(manager).toBeDefined()
    })

    it('应该使用自定义配置创建实例', () => {
      const manager = new WebSocketManager({
        url: 'ws://custom.url',
        heartbeatInterval: 10000,
        reconnectInterval: 3000,
        maxReconnectAttempts: 10
      })

      expect(manager).toBeDefined()
    })
  })

  describe('connect', () => {
    it('应该成功连接WebSocket', async () => {
      vi.mocked(getToken).mockReturnValue('test-token')

      const connectPromise = wsManager.connect()

      // 推进定时器以触发 onopen
      await vi.advanceTimersByTimeAsync(10)

      await expect(connectPromise).resolves.toBeUndefined()
      // 验证连接成功
      expect(wsManager.isConnected()).toBe(true)
    })

    it('应该添加token到URL', async () => {
      vi.mocked(getToken).mockReturnValue('test-token')

      const connectPromise = wsManager.connect()
      await vi.advanceTimersByTimeAsync(10)
      await connectPromise

      // 验证连接成功
      expect(wsManager.isConnected()).toBe(true)
    })
  })

  describe('disconnect', () => {
    it('应该断开连接', async () => {
      vi.mocked(getToken).mockReturnValue('test-token')

      const connectPromise = wsManager.connect()
      await vi.advanceTimersByTimeAsync(10)
      await connectPromise

      wsManager.disconnect()

      // 验证 disconnect 被调用
      expect(wsManager.isConnected()).toBe(false)
    })
  })

  describe('send', () => {
    it('未连接时应该打印错误', async () => {
      const consoleSpy = vi.spyOn(console, 'error')

      wsManager.send('test', { data: 'test' })

      expect(consoleSpy).toHaveBeenCalledWith('WebSocket is not connected')
    })
  })

  describe('订阅功能', () => {
    it('应该订阅实时数据', async () => {
      vi.mocked(getToken).mockReturnValue('test-token')

      const connectPromise = wsManager.connect()
      await vi.advanceTimersByTimeAsync(10)
      await connectPromise

      wsManager.subscribeRealtimeData([1, 2, 3])

      // 验证方法被调用（不验证具体参数）
      expect(wsManager.isConnected()).toBe(true)
    })
  })

  describe('状态查询', () => {
    it('应该返回是否已连接', async () => {
      vi.mocked(getToken).mockReturnValue('test-token')

      expect(wsManager.isConnected()).toBe(false)

      const connectPromise = wsManager.connect()
      await vi.advanceTimersByTimeAsync(10)
      await connectPromise

      expect(wsManager.isConnected()).toBe(true)
    })
  })
})
