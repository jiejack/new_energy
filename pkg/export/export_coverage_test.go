package export

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestCSVFlush_Explicit(t *testing.T) {
	exporter := NewCSVExporter()
	headers := []string{"ID", "名称"}
	fieldNames := []string{"ID", "Name"}

	err := exporter.SetHeaders(headers, fieldNames)
	require.NoError(t, err)

	data := TestStruct{ID: "test-001", Name: "测试1"}
	err = exporter.AddRow(data, fieldNames)
	require.NoError(t, err)

	exporter.Flush()

	assert.False(t, exporter.writer.Error() != nil)
}

func TestStreamCSVExport_Flush(t *testing.T) {
	opt := &CSVOption{
		Headers:    []string{"ID", "名称"},
		FieldNames: []string{"ID", "Name"},
	}

	streamExport, err := NewStreamCSVExport(opt)
	require.NoError(t, err)

	data := TestStruct{ID: "test-001", Name: "测试1"}
	err = streamExport.Write(data)
	require.NoError(t, err)

	streamExport.Flush()

	buf, err := streamExport.Finish()
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Greater(t, buf.Len(), 0)
}

func TestNewStreamCSVExport_WithDelimiter(t *testing.T) {
	opt := &CSVOption{
		Headers:    []string{"ID", "名称"},
		FieldNames: []string{"ID", "Name"},
		Delimiter:  ';',
	}

	streamExport, err := NewStreamCSVExport(opt)
	require.NoError(t, err)

	data := TestStruct{ID: "test-001", Name: "测试1"}
	err = streamExport.Write(data)
	require.NoError(t, err)

	buf, err := streamExport.Finish()
	assert.NoError(t, err)
	content := buf.String()
	assert.Contains(t, content, "ID;名称")
}

func TestNewStreamCSVExport_HeaderMismatch(t *testing.T) {
	opt := &CSVOption{
		Headers:    []string{"ID"},
		FieldNames: []string{"ID", "Name"},
	}

	_, err := NewStreamCSVExport(opt)
	assert.Error(t, err)
}

func TestExportCSV_HeaderMismatch(t *testing.T) {
	dataList := []TestStruct{{ID: "1", Name: "test"}}

	opt := &CSVOption{
		Headers:    []string{"ID"},
		FieldNames: []string{"ID", "Name"},
	}

	_, err := ExportCSV(dataList, opt)
	assert.Error(t, err)
}

func TestCSVAddRows_PointerToSlice(t *testing.T) {
	exporter := NewCSVExporter()
	headers := []string{"ID", "名称"}
	fieldNames := []string{"ID", "Name"}

	err := exporter.SetHeaders(headers, fieldNames)
	require.NoError(t, err)

	dataList := []TestStruct{
		{ID: "test-001", Name: "测试1"},
		{ID: "test-002", Name: "测试2"},
	}

	err = exporter.AddRows(&dataList, fieldNames)
	assert.NoError(t, err)

	buf, err := exporter.WriteToBuffer()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "test-001")
	assert.Contains(t, buf.String(), "test-002")
}

func TestExcelSetCellStyle(t *testing.T) {
	exporter := NewExcelExporter()
	headers := []string{"ID", "名称"}
	fieldNames := []string{"ID", "Name"}

	err := exporter.SetHeaders(headers, fieldNames)
	require.NoError(t, err)

	styleID, err := exporter.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	require.NoError(t, err)

	err = exporter.SetCellStyle("A1", "1", "B1", "1", styleID)
	assert.NoError(t, err)
}

func TestExcelNewStyle(t *testing.T) {
	exporter := NewExcelExporter()

	styleID, err := exporter.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FF0000"},
			Pattern: 1,
		},
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, styleID, 0)
}

func TestExcelWriteTo(t *testing.T) {
	exporter := NewExcelExporter()
	headers := []string{"ID", "名称"}
	fieldNames := []string{"ID", "Name"}

	err := exporter.SetHeaders(headers, fieldNames)
	require.NoError(t, err)

	data := TestStruct{ID: "test-001", Name: "测试1"}
	err = exporter.AddRow(data, fieldNames)
	require.NoError(t, err)

	var buf bytes.Buffer
	n, err := exporter.WriteTo(&buf)
	assert.NoError(t, err)
	assert.Greater(t, n, int64(0))
	assert.Greater(t, buf.Len(), 0)
}

func TestExcelExport_WithStyles(t *testing.T) {
	dataList := []TestStruct{
		{ID: "test-001", Name: "测试1", Value: 100.0},
	}

	opt := &ExcelOption{
		SheetName:  "样式测试",
		Headers:    []string{"ID", "名称", "值"},
		FieldNames: []string{"ID", "Name", "Value"},
		Styles:     &ExcelStyles{HeaderStyle: 0, DataStyle: 0},
	}

	buf, err := Export(dataList, opt)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestExcelExport_NoSheetName(t *testing.T) {
	dataList := []TestStruct{
		{ID: "test-001", Name: "测试1", Value: 100.0},
	}

	opt := &ExcelOption{
		Headers:    []string{"ID", "名称", "值"},
		FieldNames: []string{"ID", "Name", "Value"},
	}

	buf, err := Export(dataList, opt)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestExcelExport_NoColumnWidths(t *testing.T) {
	dataList := []TestStruct{
		{ID: "test-001", Name: "测试1", Value: 100.0},
	}

	opt := &ExcelOption{
		SheetName:    "测试",
		Headers:      []string{"ID", "名称", "值"},
		FieldNames:   []string{"ID", "Name", "Value"},
		ColumnWidths: map[string]float64{},
	}

	buf, err := Export(dataList, opt)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestNewStreamExport_WithColumnWidths(t *testing.T) {
	opt := &ExcelOption{
		SheetName:  "流式测试",
		Headers:    []string{"ID", "名称", "值"},
		FieldNames: []string{"ID", "Name", "Value"},
		ColumnWidths: map[string]float64{
			"A": 20,
			"B": 15,
		},
	}

	streamExport, err := NewStreamExport(opt)
	require.NoError(t, err)
	require.NotNil(t, streamExport)
	defer streamExport.Close()

	data := TestStruct{ID: "test-001", Name: "测试1", Value: 100.0}
	err = streamExport.Write(data)
	assert.NoError(t, err)

	buf, err := streamExport.Finish()
	assert.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestNewStreamExport_NoSheetName(t *testing.T) {
	opt := &ExcelOption{
		Headers:    []string{"ID", "名称"},
		FieldNames: []string{"ID", "Name"},
	}

	streamExport, err := NewStreamExport(opt)
	require.NoError(t, err)
	require.NotNil(t, streamExport)
	defer streamExport.Close()

	data := TestStruct{ID: "test-001", Name: "测试1"}
	err = streamExport.Write(data)
	assert.NoError(t, err)

	buf, err := streamExport.Finish()
	assert.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestNewStreamExport_HeaderMismatch(t *testing.T) {
	opt := &ExcelOption{
		Headers:    []string{"ID"},
		FieldNames: []string{"ID", "Name"},
	}

	_, err := NewStreamExport(opt)
	assert.Error(t, err)
}

func TestExcelAddRows_PointerToSlice(t *testing.T) {
	exporter := NewExcelExporter()
	headers := []string{"ID", "名称"}
	fieldNames := []string{"ID", "Name"}

	err := exporter.SetHeaders(headers, fieldNames)
	require.NoError(t, err)

	dataList := []TestStruct{
		{ID: "test-001", Name: "测试1"},
		{ID: "test-002", Name: "测试2"},
	}

	err = exporter.AddRows(&dataList, fieldNames)
	assert.NoError(t, err)

	buf, err := exporter.WriteToBuffer()
	assert.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestExcelAddRow_InvalidField(t *testing.T) {
	exporter := NewExcelExporter()
	headers := []string{"ID", "名称", "值"}
	fieldNames := []string{"ID", "Name", "Value", "InvalidField"}

	err := exporter.SetHeaders(headers, fieldNames[:3])
	require.NoError(t, err)

	data := TestStruct{ID: "test-001", Name: "测试1", Value: 100.0}
	err = exporter.AddRow(data, fieldNames)
	assert.NoError(t, err)
}

func TestExcelSetColumnWidths_Multiple(t *testing.T) {
	exporter := NewExcelExporter()

	widths := map[string]float64{
		"A": 20.0,
		"B": 15.0,
		"C": 25.0,
	}
	err := exporter.SetColumnWidths(widths)
	assert.NoError(t, err)
}

func TestExcelClose_AfterWrite(t *testing.T) {
	exporter := NewExcelExporter()
	headers := []string{"ID"}
	fieldNames := []string{"ID"}

	err := exporter.SetHeaders(headers, fieldNames)
	require.NoError(t, err)

	_, err = exporter.WriteToBuffer()
	require.NoError(t, err)

	err = exporter.Close()
	assert.NoError(t, err)
}

func TestStreamExport_WriteBatch(t *testing.T) {
	opt := &ExcelOption{
		Headers:    []string{"ID", "名称", "值"},
		FieldNames: []string{"ID", "Name", "Value"},
	}

	streamExport, err := NewStreamExport(opt)
	require.NoError(t, err)
	defer streamExport.Close()

	dataList := []TestStruct{
		{ID: "test-001", Name: "测试1", Value: 100.0},
		{ID: "test-002", Name: "测试2", Value: 200.0},
	}
	err = streamExport.WriteBatch(dataList)
	assert.NoError(t, err)

	buf, err := streamExport.Finish()
	assert.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestStreamExport_Close(t *testing.T) {
	opt := &ExcelOption{
		Headers:    []string{"ID"},
		FieldNames: []string{"ID"},
	}

	streamExport, err := NewStreamExport(opt)
	require.NoError(t, err)

	err = streamExport.Close()
	assert.NoError(t, err)
}

func TestExport_WithCreatedAt(t *testing.T) {
	dataList := []TestStruct{
		{ID: "test-001", Name: "测试1", Value: 100.0, CreatedAt: time.Now()},
	}

	opt := &ExcelOption{
		SheetName:    "测试",
		Headers:      []string{"ID", "名称", "值", "创建时间"},
		FieldNames:   []string{"ID", "Name", "Value", "CreatedAt"},
		ColumnWidths: map[string]float64{"A": 20, "B": 20, "C": 15, "D": 25},
	}

	buf, err := Export(dataList, opt)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestCSVWriteTo_AfterFlush(t *testing.T) {
	exporter := NewCSVExporter()
	headers := []string{"ID", "名称"}
	fieldNames := []string{"ID", "Name"}

	err := exporter.SetHeaders(headers, fieldNames)
	require.NoError(t, err)

	data := TestStruct{ID: "test-001", Name: "测试1"}
	err = exporter.AddRow(data, fieldNames)
	require.NoError(t, err)

	exporter.Flush()

	var buf bytes.Buffer
	n, err := exporter.WriteTo(&buf)
	assert.NoError(t, err)
	assert.Greater(t, n, int64(0))
}
