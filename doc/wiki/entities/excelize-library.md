# Excelize 库

## 基本信息

- **名称**：Excelize
- **类型**：Go 语言 Excel 文档操作库
- **开发者**：xuri
- **首次发布**：2016年
- **当前版本**：v2.8+
- **许可证**：BSD 3-Clause
- **GitHub**：https://github.com/xuri/excelize
- **官方文档**：https://xuri.me/excelize/

## 核心特性

### 文件操作
- **创建新文件**：创建空白 Excel 工作簿
- **打开文件**：读取现有 Excel 文件
- **保存文件**：保存工作簿到文件
- **流式写入**：支持大数据量流式写入
- **内存优化**：高效的内存使用

### 工作表操作
- **创建工作表**：添加新的工作表
- **删除工作表**：删除指定工作表
- **复制工作表**：复制工作表内容
- **重命名工作表**：修改工作表名称
- **设置默认工作表**：设置默认打开的工作表

### 单元格操作
- **设置单元格值**：写入各种数据类型
- **读取单元格值**：读取单元格内容
- **合并单元格**：合并多个单元格
- **拆分单元格**：取消合并单元格
- **设置单元格样式**：字体、颜色、边框等

### 数据类型支持
- **字符串**：文本数据
- **数字**：整数、浮点数
- **日期时间**：日期和时间类型
- **布尔值**：true/false
- **公式**：Excel 公式
- **富文本**：带格式的文本

## 在本项目中的应用

### 报表导出功能
- **路径**：`internal/application/service/report_service.go`
- **功能**：生成统计报表 Excel 文件
- **特性**：表头样式、数据填充、汇总行

### 实现示例
```go
func (s *ReportService) exportExcel(ctx context.Context, reportData *ReportData) ([]byte, error) {
    f := excelize.NewFile()
    defer func() {
        if err := f.Close(); err != nil {
            log.Printf("close excel file error: %v", err)
        }
    }()

    // 创建工作表
    sheetName := "统计报表"
    index, err := f.NewSheet(sheetName)
    if err != nil {
        return nil, err
    }
    f.SetActiveSheet(index)

    // 设置表头
    headers := []string{"时间", "采集点", "数据值", "单位"}
    for i, header := range headers {
        cell, _ := excelize.CoordinatesToCellName(i+1, 1)
        f.SetCellValue(sheetName, cell, header)
    }

    // 设置表头样式
    style, err := f.NewStyle(&excelize.Style{
        Font: &excelize.Font{
            Bold: true,
            Size: 12,
        },
        Fill: excelize.Fill{
            Type:    "pattern",
            Color:   []string{"#E0EBF5"},
            Pattern: 1,
        },
    })
    if err == nil {
        f.SetRowStyle(sheetName, 1, 1, style)
    }

    // 填充数据
    for i, row := range reportData.Rows {
        rowNum := i + 2
        f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), row.Time)
        f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), row.PointName)
        f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), row.Value)
        f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), row.Unit)
    }

    // 添加汇总行
    lastRow := len(reportData.Rows) + 2
    f.SetCellValue(sheetName, fmt.Sprintf("A%d", lastRow), "合计")
    f.SetCellValue(sheetName, fmt.Sprintf("C%d", lastRow), reportData.Total)

    // 保存到内存
    buf, err := f.WriteToBuffer()
    if err != nil {
        return nil, err
    }

    return buf.Bytes(), nil
}
```

## 开发规范

### 基本操作流程
1. **创建或打开文件**：`excelize.NewFile()` 或 `excelize.OpenFile()`
2. **操作工作表**：选择或创建工作表
3. **操作单元格**：读写单元格数据
4. **设置样式**：按需设置单元格样式
5. **保存文件**：保存到文件或内存
6. **关闭文件**：使用 defer 确保文件关闭

### 错误处理
- **检查错误**：每个操作都要检查错误
- **资源清理**：使用 defer 关闭文件
- **日志记录**：记录关键操作和错误

### 性能优化
- **批量操作**：减少单元格操作次数
- **流式写入**：大数据量使用流式写入
- **样式复用**：创建样式后重复使用
- **及时关闭**：操作完成后立即关闭文件

## 最佳实践

### 文件操作
- **使用 defer 关闭**：确保文件资源释放
- **错误处理完整**：不要忽略任何错误
- **文件名规范**：使用有意义的文件名

### 样式设置
- **统一风格**：保持报表风格一致
- **避免过度样式**：只设置必要的样式
- **样式复用**：相同样式使用同一 ID

### 大数据处理
- **流式写入**：使用 StreamWriter 处理大数据
- **分批处理**：数据量过大时分批处理
- **内存监控**：注意内存使用情况

## 常用功能示例

### 创建图表
```go
f := excelize.NewFile()
// 添加数据...
// 创建图表
err := f.AddChart("Sheet1", "E1", &excelize.Chart{
    Type: excelize.Line,
    Series: []excelize.ChartSeries{
        {
            Name:       "Sheet1!$A$1",
            Categories: "Sheet1!$A$2:$A$5",
            Values:     "Sheet1!$B$2:$B$5",
        },
    },
    Title: excelize.ChartTitle{
        Name: "数据趋势图",
    },
})
```

### 数据验证
```go
dv := excelize.NewDataValidation(true)
dv.Sqref = "A1:A10"
dv.SetDropList([]string{"选项1", "选项2", "选项3"})
f.AddDataValidation("Sheet1", dv)
```

### 条件格式
```go
format, err := f.NewConditionalStyle(&excelize.Style{
    Font: &excelize.Font{Color: "#9A0511"},
    Fill: excelize.Fill{Type: "pattern", Color: []string{"#FEC7CE"}, Pattern: 1},
})
err = f.SetConditionalFormat("Sheet1", "A1:A10",
    []excelize.ConditionalFormatOptions{
        {
            Type:     "cell",
            Criteria: ">",
            Value:    "100",
            Format:   format,
        },
    })
```

## 学习资源

- **官方文档**：https://xuri.me/excelize/
- **GitHub 仓库**：https://github.com/xuri/excelize
- **GoDoc**：https://pkg.go.dev/github.com/xuri/excelize/v2
- **示例代码**：https://github.com/xuri/excelize/tree/master/examples
