package export

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestStruct 测试结构体
type TestStruct struct {
	ID        string
	Name      string
	Value     float64
	CreatedAt time.Time
}

func TestNewExcelExporter(t *testing.T) {
	exporter := NewExcelExporter()
	assert.NotNil(t, exporter)
	assert.NotNil(t, exporter.file)
	assert.Equal(t, "Sheet1", exporter.sheetName)
	assert.Equal(t, 1, exporter.rowIndex)
}

func TestSetSheetName(t *testing.T) {
	exporter := NewExcelExporter()
	err := exporter.SetSheetName("TestData")
	assert.NoError(t, err)
	assert.Equal(t, "TestData", exporter.sheetName)
}

func TestSetHeaders(t *testing.T) {
	exporter := NewExcelExporter()
	headers := []string{"ID", "名称", "值", "创建时间"}
	fieldNames := []string{"ID", "Name", "Value", "CreatedAt"}

	err := exporter.SetHeaders(headers, fieldNames)
	assert.NoError(t, err)
	assert.Equal(t, 2, exporter.rowIndex)

	// 测试长度不匹配
	err = exporter.SetHeaders([]string{"ID", "名称"}, fieldNames)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "length mismatch")
}

func TestAddRow(t *testing.T) {
	exporter := NewExcelExporter()
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
	assert.Equal(t, 3, exporter.rowIndex)

	// 测试指针类型
	err = exporter.AddRow(&data, fieldNames)
	assert.NoError(t, err)
	assert.Equal(t, 4, exporter.rowIndex)

	// 测试无效类型
	err = exporter.AddRow("invalid", fieldNames)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a struct")
}

func TestAddRows(t *testing.T) {
	exporter := NewExcelExporter()
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
	assert.Equal(t, 5, exporter.rowIndex)

	// 测试无效类型
	err = exporter.AddRows("invalid", fieldNames)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a slice")
}

func TestWriteToBuffer(t *testing.T) {
	exporter := NewExcelExporter()
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

	// 验证是否为有效的Excel文件
	assert.True(t, bytes.Contains(buf.Bytes()[:4], []byte("PK")))
}

func TestExport(t *testing.T) {
	dataList := []TestStruct{
		{ID: "test-001", Name: "测试1", Value: 100.0, CreatedAt: time.Now()},
		{ID: "test-002", Name: "测试2", Value: 200.0, CreatedAt: time.Now()},
	}

	opt := &ExcelOption{
		SheetName:  "测试数据",
		Headers:    []string{"ID", "名称", "值", "创建时间"},
		FieldNames: []string{"ID", "Name", "Value", "CreatedAt"},
		ColumnWidths: map[string]float64{
			"A": 20,
			"B": 20,
			"C": 15,
			"D": 25,
		},
	}

	buf, err := Export(dataList, opt)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Greater(t, buf.Len(), 0)
}

func TestStreamExport(t *testing.T) {
	opt := &ExcelOption{
		SheetName:  "流式导出测试",
		Headers:    []string{"ID", "名称", "值"},
		FieldNames: []string{"ID", "Name", "Value"},
	}

	streamExport, err := NewStreamExport(opt)
	assert.NoError(t, err)
	assert.NotNil(t, streamExport)
	defer streamExport.Close()

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
}

func TestSetColumnWidth(t *testing.T) {
	exporter := NewExcelExporter()
	err := exporter.SetColumnWidth("A", 20.0)
	assert.NoError(t, err)

	widths := map[string]float64{
		"B": 15.0,
		"C": 25.0,
		"D": 30.0,
	}
	err = exporter.SetColumnWidths(widths)
	assert.NoError(t, err)
}

func TestClose(t *testing.T) {
	exporter := NewExcelExporter()
	err := exporter.Close()
	assert.NoError(t, err)
}

func TestExportWithEmptyData(t *testing.T) {
	dataList := []TestStruct{}

	opt := &ExcelOption{
		SheetName:  "空数据测试",
		Headers:    []string{"ID", "名称", "值"},
		FieldNames: []string{"ID", "Name", "Value"},
	}

	buf, err := Export(dataList, opt)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	// 即使没有数据，也应该生成有效的Excel文件
	assert.Greater(t, buf.Len(), 0)
}

func TestAddRowWithInvalidField(t *testing.T) {
	exporter := NewExcelExporter()
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
}
