/**
 * 全局类型声明文件
 */

// 声明 global 对象
declare var global: typeof globalThis

// 扩展 Window 接口
interface Window {
  localStorage: Storage
  WebSocket: typeof WebSocket
}

// 声明 nprogress 模块
declare module 'nprogress' {
  interface NProgress {
    start(): void
    done(): void
    set(n: number): void
    inc(n?: number): void
    configure(options: NProgressOptions): void
  }

  interface NProgressOptions {
    minimum?: number
    template?: string
    easing?: string
    positionUsing?: string
    speed?: number
    trickle?: boolean
    trickleSpeed?: number
    showSpinner?: boolean
    barSelector?: string
    spinnerSelector?: string
    parent?: string
  }

  const nprogress: NProgress
  export default nprogress
}
