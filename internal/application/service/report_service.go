package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

type ReportService struct{}

func NewReportService() *ReportService {
	return &ReportService{}
}

type ReportType string

const (
	ReportTypeDaily   ReportType = "daily"
	ReportTypeWeekly  ReportType = "weekly"
	ReportTypeMonthly ReportType = "monthly"
)

type ReportRequest struct {
	Type      ReportType `json:"type"`
	StartTime time.Time  `json:"start_time"`
	EndTime   time.Time  `json:"end_time"`
	StationID string     `json:"station_id"`
}

type StationReport struct {
	StationID   string  `json:"station_id"`
	StationName string  `json:"station_name"`
	TotalPower  float64 `json:"total_power"`
	YoYChange   float64 `json:"yoy_change"`
	MoMChange   float64 `json:"mom_change"`
	AlarmCount  int     `json:"alarm_count"`
	OnlineRate  float64 `json:"online_rate"`
}

type ReportResponse struct {
	StartTime string           `json:"start_time"`
	EndTime   string           `json:"end_time"`
	Type      ReportType       `json:"type"`
	Stations  []StationReport  `json:"stations"`
	Summary   ReportSummary    `json:"summary"`
}

type ReportSummary struct {
	TotalPower    float64 `json:"total_power"`
	TotalAlarms   int     `json:"total_alarms"`
	AvgOnlineRate float64 `json:"avg_online_rate"`
}

func (s *ReportService) GenerateStationReport(ctx context.Context, req *ReportRequest) (*ReportResponse, error) {
	stations := []StationReport{
		{
			StationID:   "station_001",
			StationName: "光伏电站A",
			TotalPower:  125000,
			YoYChange:   12.5,
			MoMChange:   5.2,
			AlarmCount:  15,
			OnlineRate:  99.5,
		},
		{
			StationID:   "station_002",
			StationName: "风电场B",
			TotalPower:  89000,
			YoYChange:   8.3,
			MoMChange:   -2.1,
			AlarmCount:  8,
			OnlineRate:  98.2,
		},
		{
			StationID:   "station_003",
			StationName: "储能电站C",
			TotalPower:  45000,
			YoYChange:   15.2,
			MoMChange:   3.8,
			AlarmCount:  3,
			OnlineRate:  99.8,
		},
	}

	var totalPower float64
	var totalAlarms int
	var totalOnlineRate float64
	for _, s := range stations {
		totalPower += s.TotalPower
		totalAlarms += s.AlarmCount
		totalOnlineRate += s.OnlineRate
	}

	return &ReportResponse{
		StartTime: req.StartTime.Format("2006-01-02"),
		EndTime:   req.EndTime.Format("2006-01-02"),
		Type:      req.Type,
		Stations:  stations,
		Summary: ReportSummary{
			TotalPower:    totalPower,
			TotalAlarms:   totalAlarms,
			AvgOnlineRate: totalOnlineRate / float64(len(stations)),
		},
	}, nil
}

func (s *ReportService) ExportReport(ctx context.Context, req *ReportRequest, format string) ([]byte, string, error) {
	report, err := s.GenerateStationReport(ctx, req)
	if err != nil {
		return nil, "", err
	}

	switch format {
	case "excel":
		return s.exportExcel(report, req)
	case "csv":
		return s.exportCSV(report, req)
	default:
		return nil, "", fmt.Errorf("unsupported format: %s", format)
	}
}

func (s *ReportService) exportExcel(report *ReportResponse, req *ReportRequest) ([]byte, string, error) {
	filename := fmt.Sprintf("report_%s_%s.xlsx", req.Type, time.Now().Format("20060102150405"))
	
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()
	
	sheetName := "报表数据"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, filename, fmt.Errorf("failed to create sheet: %w", err)
	}
	
	f.SetActiveSheet(index)
	if err := f.DeleteSheet("Sheet1"); err != nil {
		return nil, filename, fmt.Errorf("failed to delete default sheet: %w", err)
	}
	
	headers := []string{"电站ID", "电站名称", "发电量(kWh)", "同比(%)", "环比(%)", "告警数", "在线率(%)"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, filename, fmt.Errorf("failed to set header: %w", err)
		}
	}
	
	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6F3FF"},
			Pattern: 1,
		},
	})
	if err != nil {
		return nil, filename, fmt.Errorf("failed to create style: %w", err)
	}
	
	if err := f.SetRowStyle(sheetName, 1, 1, style); err != nil {
		return nil, filename, fmt.Errorf("failed to apply style: %w", err)
	}
	
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
		for j, cell := range cells {
			cellName, _ := excelize.CoordinatesToCellName(j+1, row)
			if err := f.SetCellValue(sheetName, cellName, cell); err != nil {
				return nil, filename, fmt.Errorf("failed to set cell value: %w", err)
			}
		}
	}
	
	summaryRow := len(report.Stations) + 3
	if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "总计"); err != nil {
		return nil, filename, fmt.Errorf("failed to set summary title: %w", err)
	}
	if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", summaryRow), report.Summary.TotalPower); err != nil {
		return nil, filename, fmt.Errorf("failed to set total power: %w", err)
	}
	if err := f.SetCellValue(sheetName, fmt.Sprintf("F%d", summaryRow), report.Summary.TotalAlarms); err != nil {
		return nil, filename, fmt.Errorf("failed to set total alarms: %w", err)
	}
	if err := f.SetCellValue(sheetName, fmt.Sprintf("G%d", summaryRow), report.Summary.AvgOnlineRate); err != nil {
		return nil, filename, fmt.Errorf("failed to set avg online rate: %w", err)
	}
	
	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		return nil, filename, fmt.Errorf("failed to write excel: %w", err)
	}
	
	return buf.Bytes(), filename, nil
}

func (s *ReportService) exportCSV(report *ReportResponse, req *ReportRequest) ([]byte, string, error) {
	filename := fmt.Sprintf("report_%s_%s.csv", req.Type, time.Now().Format("20060102150405"))
	csvContent := "电站名称,发电量(kWh),同比,环比,告警数,在线率\n"
	for _, s := range report.Stations {
		csvContent += fmt.Sprintf("%s,%.0f,%.1f%%,%.1f%%,%d,%.1f%%\n",
			s.StationName, s.TotalPower, s.YoYChange, s.MoMChange, s.AlarmCount, s.OnlineRate)
	}
	return []byte(csvContent), filename, nil
}
