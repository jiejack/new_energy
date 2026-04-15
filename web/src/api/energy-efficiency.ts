import { get, post } from '@/utils/request'

export interface EnergyEfficiencyRecord {
  id?: number
  station_id: number
  device_id?: number
  point_id?: number
  record_date: string
  record_time: string
  power_output?: number
  energy_consumption?: number
  efficiency_rate?: number
  capacity_factor?: number
  availability_rate?: number
  performance_ratio?: number
  comparison_data?: Record<string, any>
  weather_data?: Record<string, any>
  status?: string
  remark?: string
  created_by?: number
  updated_by?: number
  created_at?: string
  updated_at?: string
}

export interface EnergyEfficiencyAnalysis {
  id?: number
  station_id: number
  analysis_date: string
  analysis_type: string
  period_type: 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly'
  period_start: string
  period_end: string
  average_efficiency?: number
  max_efficiency?: number
  min_efficiency?: number
  efficiency_trend?: string
  total_power_output?: number
  total_energy_consumption?: number
  capacity_utilization?: number
  peak_load?: number
  peak_load_time?: string
  analysis_summary?: string
  recommendations?: string
  compared_with_period?: string
  comparison_result?: Record<string, any>
  status?: string
  created_by?: number
  updated_by?: number
  created_at?: string
  updated_at?: string
}

export interface EnergyEfficiencyTrendParams {
  station_id?: number
  period_type?: 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly'
  start_date?: string
  end_date?: string
}

export interface EnergyEfficiencyStatisticsParams {
  station_id?: number
  period_type?: 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly'
  start_date?: string
  end_date?: string
}

export interface EnergyEfficiencyComparisonParams {
  station_id?: number
  period_type?: 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly'
  current_period_start?: string
  current_period_end?: string
  compare_period_start?: string
  compare_period_end?: string
}

export interface ListParams {
  page?: number
  page_size?: number
  station_id?: number
  start_date?: string
  end_date?: string
  status?: string
}

export interface ListResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export function createEnergyEfficiencyRecord(data: EnergyEfficiencyRecord): Promise<EnergyEfficiencyRecord> {
  return post('/api/v1/energy-efficiency/records', data)
}

export function batchCreateEnergyEfficiencyRecords(data: EnergyEfficiencyRecord[]): Promise<{ count: number }> {
  return post('/api/v1/energy-efficiency/records/batch', { records: data })
}

export function listEnergyEfficiencyRecords(params: ListParams): Promise<ListResponse<EnergyEfficiencyRecord>> {
  return get('/api/v1/energy-efficiency/records', params)
}

export function getEnergyEfficiencyRecord(id: number): Promise<EnergyEfficiencyRecord> {
  return get(`/api/v1/energy-efficiency/records/${id}`)
}

export function getEnergyEfficiencyTrend(params: EnergyEfficiencyTrendParams): Promise<any> {
  return get('/api/v1/energy-efficiency/trend', params)
}

export function getEnergyEfficiencyStatistics(params: EnergyEfficiencyStatisticsParams): Promise<any> {
  return get('/api/v1/energy-efficiency/statistics', params)
}

export function getEnergyEfficiencyComparison(params: EnergyEfficiencyComparisonParams): Promise<any> {
  return get('/api/v1/energy-efficiency/comparison', params)
}

export function createEnergyEfficiencyAnalysis(data: EnergyEfficiencyAnalysis): Promise<EnergyEfficiencyAnalysis> {
  return post('/api/v1/energy-efficiency/analyses', data)
}

export function listEnergyEfficiencyAnalyses(params: ListParams): Promise<ListResponse<EnergyEfficiencyAnalysis>> {
  return get('/api/v1/energy-efficiency/analyses', params)
}

export function getEnergyEfficiencyAnalysis(id: number): Promise<EnergyEfficiencyAnalysis> {
  return get(`/api/v1/energy-efficiency/analyses/${id}`)
}

export function getLatestEnergyEfficiencyAnalysis(station_id?: number): Promise<EnergyEfficiencyAnalysis> {
  return get('/api/v1/energy-efficiency/analyses/latest', { station_id })
}
