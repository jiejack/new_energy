import { get } from '@/utils/request'

export interface ReportParams {
  type?: 'daily' | 'weekly' | 'monthly' | 'yearly'
  start_time?: string
  end_time?: string
  station_id?: string | number
}

export interface StationReport {
  station_id: string
  station_name: string
  total_power: number
  yoy_change: number
  mom_change: number
  alarm_count: number
  online_rate: number
}

export interface ReportData {
  type: string
  start_time: string
  end_time: string
  stations: StationReport[]
  summary: {
    total_power: number
    total_alarms: number
    avg_online_rate: number
  }
}

export function generateReport(params: ReportParams): Promise<ReportData> {
  return get('/reports', params)
}

export function exportReport(params: ReportParams & { format?: 'excel' | 'csv' }): Promise<Blob> {
  return get('/reports/export', params, { responseType: 'blob' })
}
