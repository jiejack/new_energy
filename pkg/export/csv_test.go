package export

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCSVExporter(t *testing.T) {
	exporter := NewCSVExporter()
	assert.NotNil(t, exporter)
	assert.NotNil(t, exporter.writer)
	assert.NotNil(t, exporter.buffer)
}

func TestSetDelimiter(t *testing.T) {
	exporter := NewCSVExporter()
	exporter.SetDelimiter(';')
	assert.Equal(t, rune(';'), exporter.writer.Comma)
}

func TestCSVSetHeaders(t *testing.T) {
	exporter := NewCSVExporter()
	headers := []string{"ID", "名称", "值", "创建时间"}
	fieldNames := []string{"ID", "Name", "Value", "CreatedAt"}

	err := exporter.SetHeaders(headers, fieldNames)
	assert.NoError(t, err)
	assert.Equal(t, headers, exporter.headers)
	assert.Equal(t, fieldNames, exporter.fieldNames)

	// 测试长度不匹配
	err = exporter.SetHeaders([]string{"ID", "名称"}, fieldNames)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "length mismatch")
}

func TestCSVAddRow(t *testing.T) {
	exporter := NewCSVExporter()
	headers := []string{"ID", "名称", "值", "创建时间"}
	fieldNames := []string{"ID", "Name", "Value", "CreatedAt"}

	err := exporter.SetHeaders(headers, fieldNames)
	assert.NoError(t, err)

	data := TestStruct{
		ID:        "test-001",
		Name:      "测试数据",
		Value:     123.45,
		CreatedAt: time.Now(),
	}

	err = exporter.AddRow(data, fieldNames)
	assert.NoError(t, err)

	// 测试指针类型
	err = exporter.AddRow(&data, fieldNames)
	assert.NoError(t, err)

	// 测试无效类型
	err = exporter.AddRow("invalid", fieldNames)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a struct")
}

func TestCSVAddRows(t *testing.T) {
	exporter := NewCSVExporter()
	headers := []string{"ID", "名称", "值", "创建时间"}
	fieldNames := []string{"ID", "Name", "Value", "CreatedAt"}

	err := exporter.SetHeaders(headers, fieldNames)
	assert.NoError(t, err)

	dataList := []TestStruct{
		{ID: "test-001", Name: "测试1", Value: 100.0, CreatedAt: time.Now()},
		{ID: "test-002", Name: "测试2", Value: 200.0, CreatedAt: time.Now()},
		{ID: "test-003", Name: "测试3", Value: 300.0, CreatedAt: time.Now()},
	}

	err = exporter.AddRows(dataList, fieldNames)
	assert.NoError(t, err)

	// 测试无效类型
	err = exporter.AddRows("invalid", fieldNames)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a slice")
}

func TestCSVWriteToBuffer(t *testing.T) {
	exporter := NewCSVExporter()
	headers := []string{"ID", "名称", "值"}
	fieldNames := []string{"ID", "Name", "Value"}

	err := exporter.SetHeaders(headers, fieldNames)
	assert.NoError(t, err)

	dataList := []TestStruct{
		{ID: "test-001", Name: "测试1", Value: 100.0},
		{ID: "test-002", Name: "测试2", Value: 200.0},
	}

	err = exporter.AddRows(dataList, fieldNames)
	assert.NoError(t, err)

	buf, err := exporter.WriteToBuffer()
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Greater(t, buf.Len(), 0)

	// 验证CSV格式
	content := buf.String()
	assert.Contains(t, content, "ID,名称,值")
	assert.Contains(t, content, "test-001")
	assert.Contains(t, content, "测试1")
}

func TestExportCSV(t *testing.T) {
	dataList := []TestStruct{
		{ID: "test-001", Name: "测试1", Value: 100.0, CreatedAt: time.Now()},
		{ID: "test-002", Name: "测试2", Value: 200.0, CreatedAt: time.Now()},
	}

	opt := &CSVOption{
		Headers:    []string{"ID", "名称", "值", "创建时间"},
		FieldNames: []string{"ID", "Name", "Value", "CreatedAt"},
	}

	buf, err := ExportCSV(dataList, opt)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Greater(t, buf.Len(), 0)

	// 验证CSV格式
	content := buf.String()
	assert.Contains(t, content, "ID,名称,值,创建时间")
	assert.Contains(t, content, "test-001")
	assert.Contains(t, content, "test-002")
}

func TestExportCSVWithCustomDelimiter(t *testing.T) {
	dataList := []TestStruct{
		{ID: "test-001", Name: "测试1", Value: 100.0},
		{ID: "test-002", Name: "测试2", Value: 200.0},
	}

	opt := &CSVOption{
		Headers:    []string{"ID", "名称", "值"},
		FieldNames: []string{"ID", "Name", "Value"},
		Delimiter:  ';',
	}

	buf, err := ExportCSV(dataList, opt)
	assert.NoError(t, err)
	assert.NotNil(t, buf)

	// 验证使用分号分隔
	content := buf.String()
	assert.Contains(t, content, "ID;名称;值")
}

func TestStreamCSVExport(t *testing.T) {
	opt := &CSVOption{
		Headers:    []string{"ID", "名称", "值"},
		FieldNames: []string{"ID", "Name", "Value"},
	}

	streamExport, err := NewStreamCSVExport(opt)
	assert.NoError(t, err)
	assert.NotNil(t, streamExport)

	// 写入单条数据
	data1 := TestStruct{ID: "test-001", Name: "测试1", Value: 100.0}
	err = streamExport.Write(data1)
	assert.NoError(t, err)

	// 批量写入数据
	dataList := []TestStruct{
		{ID: "test-002", Name: "测试2", Value: 200.0},
		{ID: "test-003", Name: "测试3", Value: 300.0},
	}
	err = streamExport.WriteBatch(dataList)
	assert.NoError(t, err)

	// 完成导出
	buf, err := streamExport.Finish()
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Greater(t, buf.Len(), 0)

	// 验证CSV内容
	content := buf.String()
	assert.Contains(t, content, "ID,名称,值")
	assert.Contains(t, content, "test-001")
	assert.Contains(t, content, "test-002")
	assert.Contains(t, content, "test-003")
}

func TestCSVFlush(t *testing.T) {
	exporter := NewCSVExporter()
	headers := []string{"ID", "名称", "值"}
	fieldNames := []string{"ID", "Name", "Value"}

	err := exporter.SetHeaders(headers, fieldNames)
	assert.NoError(t, err)

	data := TestStruct{ID: "test-001", Name: "测试1", Value: 100.0}
	err = exporter.AddRow(data, fieldNames)
	assert.NoError(t, err)

	exporter.Flush()

	buf, err := exporter.WriteToBuffer()
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Greater(t, buf.Len(), 0)
}

func TestExportCSVWithEmptyData(t *testing.T) {
	dataList := []TestStruct{}

	opt := &CSVOption{
		Headers:    []string{"ID", "名称", "值"},
		FieldNames: []string{"ID", "Name", "Value"},
	}

	buf, err := ExportCSV(dataList, opt)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	// 即使没有数据，也应该生成包含表头的CSV文件
	assert.Greater(t, buf.Len(), 0)

	content := buf.String()
	assert.Contains(t, content, "ID,名称,值")
}

func TestCSVAddRowWithInvalidField(t *testing.T) {
	exporter := NewCSVExporter()
	headers := []string{"ID", "名称", "值"}
	fieldNames := []string{"ID", "Name", "Value", "InvalidField"} // 包含不存在的字段

	err := exporter.SetHeaders(headers, fieldNames[:3]) // 只设置3个表头
	assert.NoError(t, err)

	data := TestStruct{
		ID:    "test-001",
		Name:  "测试数据",
		Value: 123.45,
	}

	// 使用包含无效字段名的列表
	err = exporter.AddRow(data, fieldNames)
	assert.NoError(t, err) // 应该不报错，只是写入空值

	buf, err := exporter.WriteToBuffer()
	assert.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "test-001")
}

func TestCSVWriteTo(t *testing.T) {
	exporter := NewCSVExporter()
	headers := []string{"ID", "名称", "值"}
	fieldNames := []string{"ID", "Name", "Value"}

	err := exporter.SetHeaders(headers, fieldNames)
	assert.NoError(t, err)

	dataList := []TestStruct{
		{ID: "test-001", Name: "测试1", Value: 100.0},
	}

	err = exporter.AddRows(dataList, fieldNames)
	assert.NoError(t, err)

	var buf bytes.Buffer
	n, err := exporter.WriteTo(&buf)
	assert.NoError(t, err)
	assert.Greater(t, n, int64(0))
	assert.Greater(t, buf.Len(), 0)
}
