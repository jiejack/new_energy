package export

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/xuri/excelize/v2"
)

// ExcelExporter Excel导出器
type ExcelExporter struct {
	file      *excelize.File
	sheetName string
	rowIndex  int
}

// ExcelOption Excel导出选项
type ExcelOption struct {
	SheetName    string
	Headers      []string
	FieldNames   []string
	ColumnWidths map[string]float64
	Styles       *ExcelStyles
}

// ExcelStyles Excel样式配置
type ExcelStyles struct {
	HeaderStyle int
	DataStyle   int
}

// NewExcelExporter 创建Excel导出器
func NewExcelExporter() *ExcelExporter {
	return &ExcelExporter{
		file:      excelize.NewFile(),
		sheetName: "Sheet1",
		rowIndex:  1,
	}
}

// SetSheetName 设置工作表名称
func (e *ExcelExporter) SetSheetName(name string) error {
	e.sheetName = name
	return e.file.SetSheetName("Sheet1", name)
}

// SetHeaders 设置表头
func (e *ExcelExporter) SetHeaders(headers []string, fieldNames []string) error {
	if len(headers) != len(fieldNames) {
		return fmt.Errorf("headers and fieldNames length mismatch")
	}

	for i, header := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, e.rowIndex)
		if err != nil {
			return fmt.Errorf("failed to get cell name: %w", err)
		}
		if err := e.file.SetCellValue(e.sheetName, cell, header); err != nil {
			return fmt.Errorf("failed to set header: %w", err)
		}
	}

	// 设置表头样式
	style, err := e.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 11,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#CCCCCC"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err == nil {
		endCol, _ := excelize.CoordinatesToCellName(len(headers), e.rowIndex)
		startCol, _ := excelize.CoordinatesToCellName(1, e.rowIndex)
		e.file.SetCellStyle(e.sheetName, startCol, endCol, style)
	}

	e.rowIndex++
	return nil
}

// AddRow 添加一行数据
func (e *ExcelExporter) AddRow(data interface{}, fieldNames []string) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a struct or pointer to struct")
	}

	for i, fieldName := range fieldNames {
		field := val.FieldByName(fieldName)
		if !field.IsValid() {
			// 如果字段不存在，写入空值
			cell, _ := excelize.CoordinatesToCellName(i+1, e.rowIndex)
			e.file.SetCellValue(e.sheetName, cell, "")
			continue
		}

		cell, err := excelize.CoordinatesToCellName(i+1, e.rowIndex)
		if err != nil {
			return fmt.Errorf("failed to get cell name: %w", err)
		}

		value := field.Interface()
		if err := e.file.SetCellValue(e.sheetName, cell, value); err != nil {
			return fmt.Errorf("failed to set cell value: %w", err)
		}
	}

	e.rowIndex++
	return nil
}

// AddRows 批量添加数据行
func (e *ExcelExporter) AddRows(dataList interface{}, fieldNames []string) error {
	val := reflect.ValueOf(dataList)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Slice {
		return fmt.Errorf("dataList must be a slice or pointer to slice")
	}

	for i := 0; i < val.Len(); i++ {
		if err := e.AddRow(val.Index(i).Interface(), fieldNames); err != nil {
			return err
		}
	}

	return nil
}

// SetColumnWidth 设置列宽
func (e *ExcelExporter) SetColumnWidth(col string, width float64) error {
	return e.file.SetColWidth(e.sheetName, col, col, width)
}

// SetColumnWidths 批量设置列宽
func (e *ExcelExporter) SetColumnWidths(widths map[string]float64) error {
	for col, width := range widths {
		if err := e.SetColumnWidth(col, width); err != nil {
			return err
		}
	}
	return nil
}

// SetCellStyle 设置单元格样式
func (e *ExcelExporter) SetCellStyle(startCol, startRow, endCol, endRow string, styleID int) error {
	return e.file.SetCellStyle(e.sheetName, startCol, endCol, styleID)
}

// NewStyle 创建新样式
func (e *ExcelExporter) NewStyle(style *excelize.Style) (int, error) {
	return e.file.NewStyle(style)
}

// WriteTo 写入到Writer
func (e *ExcelExporter) WriteTo(w io.Writer) (int64, error) {
	buf, err := e.file.WriteToBuffer()
	if err != nil {
		return 0, fmt.Errorf("failed to write to buffer: %w", err)
	}
	return buf.WriteTo(w)
}

// WriteToBuffer 写入到buffer
func (e *ExcelExporter) WriteToBuffer() (*bytes.Buffer, error) {
	return e.file.WriteToBuffer()
}

// Close 关闭文件
func (e *ExcelExporter) Close() error {
	return e.file.Close()
}

// Export 导出数据到Excel
func Export(dataList interface{}, opt *ExcelOption) (*bytes.Buffer, error) {
	exporter := NewExcelExporter()

	if opt.SheetName != "" {
		if err := exporter.SetSheetName(opt.SheetName); err != nil {
			exporter.Close()
			return nil, err
		}
	}

	if err := exporter.SetHeaders(opt.Headers, opt.FieldNames); err != nil {
		exporter.Close()
		return nil, err
	}

	if err := exporter.AddRows(dataList, opt.FieldNames); err != nil {
		exporter.Close()
		return nil, err
	}

	// 设置列宽
	if len(opt.ColumnWidths) > 0 {
		exporter.SetColumnWidths(opt.ColumnWidths)
	}

	buf, err := exporter.WriteToBuffer()
	if err != nil {
		exporter.Close()
		return nil, err
	}

	exporter.Close()
	return buf, nil
}

// StreamExport 流式导出大数据量到Excel
type StreamExport struct {
	exporter   *ExcelExporter
	fieldNames []string
}

// NewStreamExport 创建流式导出器
func NewStreamExport(opt *ExcelOption) (*StreamExport, error) {
	exporter := NewExcelExporter()

	if opt.SheetName != "" {
		if err := exporter.SetSheetName(opt.SheetName); err != nil {
			exporter.Close()
			return nil, err
		}
	}

	if err := exporter.SetHeaders(opt.Headers, opt.FieldNames); err != nil {
		exporter.Close()
		return nil, err
	}

	// 设置列宽
	if len(opt.ColumnWidths) > 0 {
		exporter.SetColumnWidths(opt.ColumnWidths)
	}

	return &StreamExport{
		exporter:   exporter,
		fieldNames: opt.FieldNames,
	}, nil
}

// Write 写入数据
func (s *StreamExport) Write(data interface{}) error {
	return s.exporter.AddRow(data, s.fieldNames)
}

// WriteBatch 批量写入数据
func (s *StreamExport) WriteBatch(dataList interface{}) error {
	return s.exporter.AddRows(dataList, s.fieldNames)
}

// Finish 完成导出
func (s *StreamExport) Finish() (*bytes.Buffer, error) {
	buf, err := s.exporter.WriteToBuffer()
	if err != nil {
		s.exporter.Close()
		return nil, err
	}
	s.exporter.Close()
	return buf, nil
}

// Close 关闭导出器
func (s *StreamExport) Close() error {
	return s.exporter.Close()
}
