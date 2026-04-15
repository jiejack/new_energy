# 报表导出实现

## 功能说明

报表导出功能允许用户将统计报表数据导出为 Excel 或 CSV 格式，方便离线分析和存档。

## 技术实现

### 后端实现

#### 1. 导出服务

在 `internal/application/service/report_service.go` 中实现了导出功能：

- **exportExcel** 方法：生成 Excel 格式的报表
- **exportCSV** 方法：生成 CSV 格式的报表

#### 2. Excel 导出实现

```go
func (s *ReportService) exportExcel(report *ReportResponse, req *ReportRequest) ([]byte, string, error) {
    // 创建 Excel 文件
    f := excelize.NewFile()
    
    // 创建报表工作表
    sheetName := "报表数据"
    index, err := f.NewSheet(sheetName)
    
    // 设置表头
    headers := []string{"电站ID", "电站名称", "发电量(kWh)", "同比(%)", "环比(%)", "告警数", "在线率(%)"}
    
    // 填充数据
    for i, station := range report.Stations {
        row := i + 2
        cells := []interface{}{
            station.StationID,
            station.StationName,
            station.TotalPower,
            station.YoYChange,
            station.MoMChange,
            station.AlarmCount,
            station.OnlineRate,
        }
        // 设置单元格值
    }
    
    // 添加汇总行
    summaryRow := len(report.Stations) + 3
    f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "总计")
    f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), report.Summary.TotalPower)
    f.SetCellValue(sheetName, fmt.Sprintf("F%d", summaryRow), report.Summary.TotalAlarms)
    f.SetCellValue(sheetName, fmt.Sprintf("G%d", summaryRow), report.Summary.AvgOnlineRate)
    
    // 生成 Excel 文件字节
    buf := new(bytes.Buffer)
    f.Write(buf)
    
    return buf.Bytes(), filename, nil
}
```

#### 3. API 接口

在 `internal/api/handler/report_handler.go` 中实现了导出接口：

```go
func (h *ReportHandler) ExportReport(c *gin.Context) {
    // 解析请求参数
    // 生成报表数据
    // 根据格式导出
    // 返回文件
}
```

#### 4. 路由配置

在 `cmd/api-server/app.go` 中配置了导出路由：

```go
reports := api.Group("/reports")
{
    reports.GET("", reportHandler.GenerateReport)
    reports.GET("/export", reportHandler.ExportReport)
}
```

### 前端实现

#### 1. API 调用

在 `web/src/api/report.ts` 中定义了导出接口：

```typescript
export function exportReport(params: ReportParams & { format?: 'excel' | 'csv' }): Promise<Blob> {
  return get('/api/v1/reports/export', params, { responseType: 'blob' })
}
```

#### 2. 导出操作

在 `web/src/views/data/report/index.vue` 中实现了导出功能：

```typescript
const handleExport = async (format: 'excel' | 'csv') => {
  try {
    const params = {
      type: form.type,
      start_time: form.dateRange?.[0],
      end_time: form.dateRange?.[1],
      station_id: form.stationId,
      format
    }
    const blob = await exportReport(params)
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `report_${form.type}_${new Date().toISOString().slice(0, 10)}.${format}`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    window.URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}
```

## 导出流程

1. **用户操作**：用户在报表页面选择导出格式（Excel 或 CSV）
2. **前端请求**：前端调用 `/api/v1/reports/export` 接口，传递报表参数和导出格式
3. **后端处理**：
   - 解析请求参数
   - 生成报表数据
   - 根据格式调用相应的导出方法
   - 生成文件并返回
4. **前端处理**：
   - 接收文件 blob
   - 创建下载链接
   - 触发下载
   - 显示成功提示

## 性能优化

1. **内存优化**：使用 `bytes.Buffer` 减少内存占用
2. **处理大量数据**：实现分页处理，避免一次性加载大量数据
3. **异步处理**：对于大型报表，考虑使用异步导出
4. **缓存策略**：对相同参数的报表进行缓存，提高导出速度

## 注意事项

1. **文件大小限制**：对于大型报表，可能需要限制导出数据量
2. **导出超时**：设置合理的超时时间，避免长时间阻塞
3. **错误处理**：对导出过程中的错误进行适当处理和提示
4. **文件名规范**：使用清晰的命名规则，包含报表类型和日期