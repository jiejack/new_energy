import { get, post } from '@/utils/request'

export interface CarbonEmissionRecord {
  id?: number
  station_id: number
  device_id?: number
  point_id?: number
  record_date: string
  record_time: string
  electricity_consumption?: number
  coal_consumption?: number
  natural_gas_consumption?: number
  oil_consumption?: number
  other_energy_consumption?: number
  electricity_emission?: number
  coal_emission?: number
  natural_gas_emission?: number
  oil_emission?: number
  other_energy_emission?: number
  total_emission?: number
  emission_intensity?: number
  emission_factor_electricity?: number
  emission_factor_coal?: number
  emission_factor_natural_gas?: number
  emission_factor_oil?: number
  emission_factor_other?: number
  status?: string
  remark?: string
  created_by?: number
  updated_by?: number
  created_at?: string
  updated_at?: string
}

export interface CarbonEmissionAnalysis {
  id?: number
  station_id: number
  analysis_date: string
  analysis_type: string
  period_type: 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly'
  period_start: string
  period_end: string
  total_emission?: number
  emission_change_rate?: number
  electricity_emission?: number
  coal_emission?: number
  natural_gas_emission?: number
  oil_emission?: number
  other_energy_emission?: number
  average_emission_intensity?: number
  intensity_change_rate?: number
  peak_emission?: number
  peak_emission_time?: string
  reduction_target?: number
  actual_reduction?: number
  reduction_rate?: number
  analysis_summary?: string
  key_findings?: string
  recommendations?: string
  compared_with_period?: string
  comparison_result?: Record<string, any>
  status?: string
  created_by?: number
  updated_by?: number
  created_at?: string
  updated_at?: string
}

export interface CarbonEmissionTrendParams {
  station_id?: number
  period_type?: 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly'
  start_date?: string
  end_date?: string
}

export interface CarbonEmissionStatisticsParams {
  station_id?: number
  period_type?: 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly'
  start_date?: string
  end_date?: string
}

export interface CarbonEmissionComparisonParams {
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

export function createCarbonEmissionRecord(data: CarbonEmissionRecord): Promise<CarbonEmissionRecord> {
  return post('/api/v1/carbon-emission/records', data)
}

export function batchCreateCarbonEmissionRecords(data: CarbonEmissionRecord[]): Promise<{ count: number }> {
  return post('/api/v1/carbon-emission/records/batch', { records: data })
}

export function listCarbonEmissionRecords(params: ListParams): Promise<ListResponse<CarbonEmissionRecord>> {
  return get('/api/v1/carbon-emission/records', params)
}

export function getCarbonEmissionRecord(id: number): Promise<CarbonEmissionRecord> {
  return get(`/api/v1/carbon-emission/records/${id}`)
}

export function getCarbonEmissionTrend(params: CarbonEmissionTrendParams): Promise<any> {
  return get('/api/v1/carbon-emission/trend', params)
}

export function getCarbonEmissionStatistics(params: CarbonEmissionStatisticsParams): Promise<any> {
  return get('/api/v1/carbon-emission/statistics', params)
}

export function getCarbonEmissionComparison(params: CarbonEmissionComparisonParams): Promise<any> {
  return get('/api/v1/carbon-emission/comparison', params)
}

export function createCarbonEmissionAnalysis(data: CarbonEmissionAnalysis): Promise<CarbonEmissionAnalysis> {
  return post('/api/v1/carbon-emission/analyses', data)
}

export function listCarbonEmissionAnalyses(params: ListParams): Promise<ListResponse<CarbonEmissionAnalysis>> {
  return get('/api/v1/carbon-emission/analyses', params)
}

export function getCarbonEmissionAnalysis(id: number): Promise<CarbonEmissionAnalysis> {
  return get(`/api/v1/carbon-emission/analyses/${id}`)
}

export function getLatestCarbonEmissionAnalysis(station_id?: number): Promise<CarbonEmissionAnalysis> {
  return get('/api/v1/carbon-emission/analyses/latest', { station_id })
}
