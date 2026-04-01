package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
)

// CSVExporter CSV导出器
type CSVExporter struct {
	writer     *csv.Writer
	buffer     *bytes.Buffer
	headers    []string
	fieldNames []string
}

// CSVOption CSV导出选项
type CSVOption struct {
	Headers    []string
	FieldNames []string
	Delimiter  rune // 分隔符，默认为逗号
}

// NewCSVExporter 创建CSV导出器
func NewCSVExporter() *CSVExporter {
	buf := &bytes.Buffer{}
	return &CSVExporter{
		writer: csv.NewWriter(buf),
		buffer: buf,
	}
}

// SetDelimiter 设置分隔符
func (e *CSVExporter) SetDelimiter(delimiter rune) {
	e.writer.Comma = delimiter
}

// SetHeaders 设置表头
func (e *CSVExporter) SetHeaders(headers []string, fieldNames []string) error {
	if len(headers) != len(fieldNames) {
		return fmt.Errorf("headers and fieldNames length mismatch")
	}

	e.headers = headers
	e.fieldNames = fieldNames

	if err := e.writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	return nil
}

// AddRow 添加一行数据
func (e *CSVExporter) AddRow(data interface{}, fieldNames []string) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a struct or pointer to struct")
	}

	row := make([]string, len(fieldNames))
	for i, fieldName := range fieldNames {
		field := val.FieldByName(fieldName)
		if !field.IsValid() {
			row[i] = ""
			continue
		}

		row[i] = fmt.Sprintf("%v", field.Interface())
	}

	if err := e.writer.Write(row); err != nil {
		return fmt.Errorf("failed to write row: %w", err)
	}

	return nil
}

// AddRows 批量添加数据行
func (e *CSVExporter) AddRows(dataList interface{}, fieldNames []string) error {
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

// Flush 刷新缓冲区
func (e *CSVExporter) Flush() {
	e.writer.Flush()
}

// WriteTo 写入到Writer
func (e *CSVExporter) WriteTo(w io.Writer) (int64, error) {
	e.Flush()
	return e.buffer.WriteTo(w)
}

// WriteToBuffer 写入到buffer
func (e *CSVExporter) WriteToBuffer() (*bytes.Buffer, error) {
	e.Flush()
	return e.buffer, nil
}

// Export 导出数据到CSV
func ExportCSV(dataList interface{}, opt *CSVOption) (*bytes.Buffer, error) {
	exporter := NewCSVExporter()

	if opt.Delimiter != 0 {
		exporter.SetDelimiter(opt.Delimiter)
	}

	if err := exporter.SetHeaders(opt.Headers, opt.FieldNames); err != nil {
		return nil, err
	}

	if err := exporter.AddRows(dataList, opt.FieldNames); err != nil {
		return nil, err
	}

	return exporter.WriteToBuffer()
}

// StreamCSVExport 流式导出大数据量到CSV
type StreamCSVExport struct {
	exporter   *CSVExporter
	fieldNames []string
}

// NewStreamCSVExport 创建流式CSV导出器
func NewStreamCSVExport(opt *CSVOption) (*StreamCSVExport, error) {
	exporter := NewCSVExporter()

	if opt.Delimiter != 0 {
		exporter.SetDelimiter(opt.Delimiter)
	}

	if err := exporter.SetHeaders(opt.Headers, opt.FieldNames); err != nil {
		return nil, err
	}

	return &StreamCSVExport{
		exporter:   exporter,
		fieldNames: opt.FieldNames,
	}, nil
}

// Write 写入数据
func (s *StreamCSVExport) Write(data interface{}) error {
	return s.exporter.AddRow(data, s.fieldNames)
}

// WriteBatch 批量写入数据
func (s *StreamCSVExport) WriteBatch(dataList interface{}) error {
	return s.exporter.AddRows(dataList, s.fieldNames)
}

// Finish 完成导出
func (s *StreamCSVExport) Finish() (*bytes.Buffer, error) {
	return s.exporter.WriteToBuffer()
}

// Flush 刷新缓冲区
func (s *StreamCSVExport) Flush() {
	s.exporter.Flush()
}
