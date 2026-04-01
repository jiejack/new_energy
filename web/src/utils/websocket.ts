import { getToken } from './auth'
import type { WsMessage, RealtimeData } from '@/types'

type MessageHandler = (data: any) => void
type ConnectionHandler = () => void
type ErrorHandler = (error: Event) => void

interface WebSocketOptions {
  url?: string
  heartbeatInterval?: number
  reconnectInterval?: number
  maxReconnectAttempts?: number
}

class WebSocketManager {
  private ws: WebSocket | null = null
  private url: string
  private heartbeatInterval: number
  private reconnectInterval: number
  private maxReconnectAttempts: number
  private reconnectAttempts = 0
  private heartbeatTimer: number | null = null
  private reconnectTimer: number | null = null
  private messageHandlers: Map<string, Set<MessageHandler>> = new Map()
  private connectionHandlers: Set<ConnectionHandler> = new Set()
  private disconnectionHandlers: Set<ConnectionHandler> = new Set()
  private errorHandlers: Set<ErrorHandler> = new Set()
  private isManualClose = false

  constructor(options: WebSocketOptions = {}) {
    this.url = options.url || this.getDefaultUrl()
    this.heartbeatInterval = options.heartbeatInterval || 30000
    this.reconnectInterval = options.reconnectInterval || 5000
    this.maxReconnectAttempts = options.maxReconnectAttempts || 5
  }

  /**
   * 获取默认WebSocket URL
   */
  private getDefaultUrl(): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    return `${protocol}//${host}/ws`
  }

  /**
   * 连接WebSocket
   */
  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        resolve()
        return
      }

      this.isManualClose = false
      const token = getToken()
      const wsUrl = token ? `${this.url}?token=${token}` : this.url

      try {
        this.ws = new WebSocket(wsUrl)

        this.ws.onopen = () => {
          console.log('WebSocket connected')
          this.reconnectAttempts = 0
          this.startHeartbeat()
          this.connectionHandlers.forEach((handler) => handler())
          resolve()
        }

        this.ws.onmessage = (event) => {
          this.handleMessage(event.data)
        }

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error)
          this.errorHandlers.forEach((handler) => handler(error))
          reject(error)
        }

        this.ws.onclose = () => {
          console.log('WebSocket closed')
          this.stopHeartbeat()
          this.disconnectionHandlers.forEach((handler) => handler())

          if (!this.isManualClose) {
            this.reconnect()
          }
        }
      } catch (error) {
        reject(error)
      }
    })
  }

  /**
   * 断开连接
   */
  disconnect(): void {
    this.isManualClose = true
    this.stopHeartbeat()
    this.stopReconnect()

    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  /**
   * 重连
   */
  private reconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnect attempts reached')
      return
    }

    this.reconnectAttempts++
    console.log(`Reconnecting... Attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts}`)

    this.reconnectTimer = window.setTimeout(() => {
      this.connect().catch((error) => {
        console.error('Reconnect failed:', error)
      })
    }, this.reconnectInterval)
  }

  /**
   * 停止重连
   */
  private stopReconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
  }

  /**
   * 开始心跳
   */
  private startHeartbeat(): void {
    this.heartbeatTimer = window.setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.send('heartbeat', { timestamp: Date.now() })
      }
    }, this.heartbeatInterval)
  }

  /**
   * 停止心跳
   */
  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  /**
   * 处理消息
   */
  private handleMessage(data: string): void {
    try {
      const message: WsMessage = JSON.parse(data)
      const handlers = this.messageHandlers.get(message.type)

      if (handlers) {
        handlers.forEach((handler) => handler(message.payload))
      }

      // 处理心跳响应
      if (message.type === 'heartbeat') {
        // 心跳响应，无需处理
      }
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error)
    }
  }

  /**
   * 发送消息
   */
  send(type: string, payload: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message: WsMessage = {
        type,
        payload,
        timestamp: Date.now(),
      }
      this.ws.send(JSON.stringify(message))
    } else {
      console.error('WebSocket is not connected')
    }
  }

  /**
   * 订阅实时数据
   */
  subscribeRealtimeData(pointIds: number[]): void {
    this.send('subscribe', { pointIds })
  }

  /**
   * 取消订阅实时数据
   */
  unsubscribeRealtimeData(pointIds: number[]): void {
    this.send('unsubscribe', { pointIds })
  }

  /**
   * 订阅告警
   */
  subscribeAlarm(): void {
    this.send('subscribe-alarm', {})
  }

  /**
   * 取消订阅告警
   */
  unsubscribeAlarm(): void {
    this.send('unsubscribe-alarm', {})
  }

  /**
   * 监听消息
   */
  on(type: string, handler: MessageHandler): () => void {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, new Set())
    }
    this.messageHandlers.get(type)!.add(handler)

    // 返回取消监听函数
    return () => {
      this.off(type, handler)
    }
  }

  /**
   * 取消监听消息
   */
  off(type: string, handler: MessageHandler): void {
    const handlers = this.messageHandlers.get(type)
    if (handlers) {
      handlers.delete(handler)
      if (handlers.size === 0) {
        this.messageHandlers.delete(type)
      }
    }
  }

  /**
   * 监听连接
   */
  onConnect(handler: ConnectionHandler): () => void {
    this.connectionHandlers.add(handler)
    return () => {
      this.connectionHandlers.delete(handler)
    }
  }

  /**
   * 监听断开连接
   */
  onDisconnect(handler: ConnectionHandler): () => void {
    this.disconnectionHandlers.add(handler)
    return () => {
      this.disconnectionHandlers.delete(handler)
    }
  }

  /**
   * 监听错误
   */
  onError(handler: ErrorHandler): () => void {
    this.errorHandlers.add(handler)
    return () => {
      this.errorHandlers.delete(handler)
    }
  }

  /**
   * 获取连接状态
   */
  getReadyState(): number {
    return this.ws ? this.ws.readyState : WebSocket.CLOSED
  }

  /**
   * 是否已连接
   */
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }
}

// 创建全局WebSocket实例
export const wsManager = new WebSocketManager()

export default WebSocketManager
