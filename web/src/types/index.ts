/**
 * 用户相关类型定义
 */
export interface UserInfo {
  id: number
  username: string
  nickname: string
  email: string
  phone: string
  avatar: string
  status: number
  roles: string[]
  permissions: string[]
  createdAt: string
  updatedAt: string
}

export interface LoginForm {
  username: string
  password: string
  captcha?: string
  uuid?: string
}

export interface LoginResponse {
  token: string
  refreshToken: string
  expiresIn: number
  user: UserInfo
}

/**
 * 区域相关类型定义
 */
export interface Region {
  id: number
  name: string
  code: string
  parentId: number | null
  level: number
  status: number
  description: string
  createdAt: string
  updatedAt: string
  children?: Region[]
}

/**
 * 电站相关类型定义
 */
export interface Station {
  id: number
  name: string
  code: string
  regionId: number
  regionName: string
  type: StationType
  capacity: number
  status: StationStatus
  address: string
  longitude: number
  latitude: number
  description: string
  createdAt: string
  updatedAt: string
}

export type StationType = 'wind' | 'solar' | 'hydro' | 'storage'
export type StationStatus = 'online' | 'offline' | 'maintenance' | 'fault'

/**
 * 设备相关类型定义
 */
export interface Device {
  id: number
  name: string
  code: string
  stationId: number
  stationName: string
  type: DeviceType
  model: string
  manufacturer: string
  status: DeviceStatus
  installDate: string
  lastMaintenanceDate: string
  description: string
  createdAt: string
  updatedAt: string
}

export type DeviceType = 'inverter' | 'meter' | 'sensor' | 'controller' | 'combiner'
export type DeviceStatus = 'online' | 'offline' | 'maintenance' | 'fault'

/**
 * 采集点相关类型定义
 */
export interface Point {
  id: number
  name: string
  code: string
  deviceId: number
  deviceName: string
  type: PointType
  unit: string
  dataType: DataType
  minValue: number
  maxValue: number
  description: string
  createdAt: string
  updatedAt: string
}

export type PointType = 'analog' | 'digital' | 'pulse'
export type DataType = 'float' | 'int' | 'bool' | 'string'

/**
 * 告警相关类型定义
 */
export interface Alarm {
  id: number
  title: string
  content: string
  level: AlarmLevel
  status: AlarmStatus
  sourceType: string
  sourceId: number
  sourceName: string
  occurredAt: string
  acknowledgedAt: string | null
  acknowledgedBy: string | null
  resolvedAt: string | null
  resolvedBy: string | null
  createdAt: string
  updatedAt: string
}

export type AlarmLevel = 'critical' | 'major' | 'minor' | 'warning'
export type AlarmStatus = 'active' | 'acknowledged' | 'resolved'

/**
 * 数据查询相关类型定义
 */
export interface DataQuery {
  pointIds: number[]
  startTime: string
  endTime: string
  interval?: number
  aggregation?: 'avg' | 'max' | 'min' | 'sum'
}

export interface DataPoint {
  timestamp: string
  value: number
  quality: number
}

export interface PointData {
  pointId: number
  pointName: string
  unit: string
  data: DataPoint[]
}

/**
 * 分页相关类型定义
 */
export interface PageQuery {
  page: number
  pageSize: number
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

/**
 * API响应通用类型
 */
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

/**
 * 路由相关类型定义
 */
export interface RouteMeta {
  title: string
  icon?: string
  hidden?: boolean
  keepAlive?: boolean
  requiresAuth?: boolean
  permissions?: string[]
}

export interface AppRoute {
  path: string
  name?: string
  component?: any
  redirect?: string
  meta?: RouteMeta
  children?: AppRoute[]
}

/**
 * WebSocket消息类型
 */
export interface WsMessage<T = any> {
  type: string
  payload: T
  timestamp: number
}

export interface RealtimeData {
  pointId: number
  value: number
  quality: number
  timestamp: number
}

/**
 * 角色相关类型定义
 */
export interface Role {
  id: number
  name: string
  code: string
  description: string
  status: number
  permissions: number[]
  createdAt: string
  updatedAt: string
}

/**
 * 权限相关类型定义
 */
export interface Permission {
  id: number
  name: string
  code: string
  type: PermissionType
  parentId: number | null
  path?: string
  icon?: string
  sort: number
  status: number
  children?: Permission[]
  createdAt: string
  updatedAt: string
}

export type PermissionType = 'menu' | 'button' | 'api'

/**
 * 操作日志相关类型定义
 */
export interface OperationLog {
  id: number
  userId: number
  username: string
  operation: string
  method: string
  params: string
  ip: string
  location: string
  userAgent: string
  status: number
  errorMsg?: string
  duration: number
  createdAt: string
}

/**
 * 用户扩展信息
 */
export interface UserDetail extends UserInfo {
  deptId?: number
  deptName?: string
  postIds?: number[]
  postNames?: string[]
  roleIds?: number[]
  roleNames?: string[]
}
