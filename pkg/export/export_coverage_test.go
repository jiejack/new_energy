package export

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRow struct {
	Name   string
	Age    int
	Score  float64
	Active bool
}

func TestCSVExporter_NewCoverage(t *testing.T) {
	exporter := NewCSVExporter()
	require.NotNil(t, exporter)
	assert.NotNil(t, exporter.writer)
	assert.NotNil(t, exporter.buffer)
}

func TestCSVExporter_SetDelimiterCoverage(t *testing.T) {
	exporter := NewCSVExporter()
	exporter.SetDelimiter('\t')
	assert.Equal(t, '\t', exporter.writer.Comma)
}

func TestCSVExporter_SetHeadersCoverage(t *testing.T) {
	exporter := NewCSVExporter()
	err := exporter.SetHeaders([]string{"Name", "Age"}, []string{"Name", "Age"})
	require.NoError(t, err)
	assert.Equal(t, []string{"Name", "Age"}, exporter.headers)
}

func TestCSVExporter_SetHeaders_MismatchCoverage(t *testing.T) {
	exporter := NewCSVExporter()
	err := exporter.SetHeaders([]string{"Name"}, []string{"Name", "Age"})
	assert.Error(t, err)
}

func TestCSVExporter_AddRowCoverage(t *testing.T) {
	exporter := NewCSVExporter()
	err := exporter.SetHeaders([]string{"Name", "Age"}, []string{"Name", "Age"})
	require.NoError(t, err)

	err = exporter.AddRow(&testRow{Name: "Alice", Age: 30, Score: 95.5, Active: true}, []string{"Name", "Age"})
	require.NoError(t, err)
}

func TestCSVExporter_AddRow_NonStructCoverage(t *testing.T) {
	exporter := NewCSVExporter()
	err := exporter.AddRow("not a struct", []string{"Name"})
	assert.Error(t, err)
}

func TestCSVExporter_AddRow_PtrToStructCoverage(t *testing.T) {
	exporter := NewCSVExporter()
	row := &testRow{Name: "Bob", Age: 25}
	err := exporter.AddRow(row, []string{"Name"})
	require.NoError(t, err)
}

func TestCSVExporter_AddRowsCoverage(t *testing.T) {
	exporter := NewCSVExporter()
	data := []testRow{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	}
	err := exporter.AddRows(data, []string{"Name", "Age"})
	require.NoError(t, err)
}

func TestCSVExporter_WriteToBufferCoverage(t *testing.T) {
	exporter := NewCSVExporter()
	exporter.SetHeaders([]string{"Name"}, []string{"Name"})

	buf, err := exporter.WriteToBuffer()
	require.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestExportCSVCoverage(t *testing.T) {
	data := []testRow{
		{Name: "Alice", Age: 30, Score: 95.5, Active: true},
		{Name: "Bob", Age: 25, Score: 88.0, Active: false},
	}
	opt := &CSVOption{
		Headers:    []string{"Name", "Age", "Score", "Active"},
		FieldNames: []string{"Name", "Age", "Score", "Active"},
		Delimiter:  ',',
	}

	buf, err := ExportCSV(data, opt)
	require.NoError(t, err)
	assert.NotNil(t, buf)
	output := buf.String()
	assert.Contains(t, output, "Alice")
	assert.Contains(t, output, "Bob")
}

func TestStreamCSVExport_OperationsCoverage(t *testing.T) {
	stream, err := NewStreamCSVExport(&CSVOption{
		Headers:    []string{"Name"},
		FieldNames: []string{"Name"},
	})
	require.NoError(t, err)

	err = stream.Write(testRow{Name: "StreamData"})
	require.NoError(t, err)

	data := []testRow{{Name: "A"}, {Name: "B"}}
	err = stream.WriteBatch(data)
	require.NoError(t, err)

	buf, err := stream.Finish()
	require.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestExcelExporter_FullOperationsCoverage(t *testing.T) {
	exporter := NewExcelExporter()
	defer exporter.Close()

	err := exporter.SetSheetName("MySheet")
	require.NoError(t, err)

	err = exporter.SetHeaders([]string{"Name", "Age"}, []string{"Name", "Age"})
	require.NoError(t, err)

	err = exporter.AddRow(testRow{Name: "ExcelRow"}, []string{"Name"})
	require.NoError(t, err)

	data := []testRow{{Name: "A"}, {Name: "B"}}
	err = exporter.AddRows(data, []string{"Name"})
	require.NoError(t, err)

	err = exporter.SetColumnWidth("A", 20.0)
	require.NoError(t, err)

	widths := map[string]float64{"A": 15.0, "B": 30.0}
	err = exporter.SetColumnWidths(widths)
	require.NoError(t, err)

	styleID, err := exporter.NewStyle(nil)
	require.NoError(t, err)

	err = exporter.SetCellStyle("A1", "1", "B1", "1", styleID)
	require.NoError(t, err)

	buf, err := exporter.WriteToBuffer()
	require.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Greater(t, buf.Len(), 0)
}

func TestExport_ExcelCoverage(t *testing.T) {
	data := []testRow{
		{Name: "Alice", Age: 30, Score: 95.5, Active: true},
	}
	opt := &ExcelOption{
		SheetName:     "TestData",
		Headers:       []string{"Name", "Age", "Score", "Active"},
		FieldNames:    []string{"Name", "Age", "Score", "Active"},
		ColumnWidths:  map[string]float64{"A": 15.0},
	}

	buf, err := Export(data, opt)
	require.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Greater(t, buf.Len(), 0)
}

func TestStreamExport_OperationsCoverage(t *testing.T) {
	stream, _ := NewStreamExport(&ExcelOption{
		Headers:      []string{"Name"},
		FieldNames:   []string{"Name"},
		ColumnWidths: map[string]float64{"A": 20.0},
	})
	defer stream.Close()

	err := stream.Write(testRow{Name: "StreamRow"})
	require.NoError(t, err)

	data := []testRow{{Name: "A"}, {Name: "B"}}
	err = stream.WriteBatch(data)
	require.NoError(t, err)

	buf, err := stream.Finish()
	require.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Greater(t, buf.Len(), 0)
}
