package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
	"github.com/new-energy-monitoring/pkg/export"
)

// ExportServiceInterface 导出服务接口
type ExportServiceInterface interface {
	Export(ctx context.Context, req *ExportRequest) (*ExportResult, error)
}

// ExportService 导出服务
type ExportService struct {
	alarmRepo   repository.AlarmRepository
	deviceRepo  repository.DeviceRepository
	stationRepo repository.StationRepository
}

// 确保ExportService实现了ExportServiceInterface接口
var _ ExportServiceInterface = (*ExportService)(nil)

// NewExportService 创建导出服务
func NewExportService(
	alarmRepo repository.AlarmRepository,
	deviceRepo repository.DeviceRepository,
	stationRepo repository.StationRepository,
) *ExportService {
	return &ExportService{
		alarmRepo:   alarmRepo,
		deviceRepo:  deviceRepo,
		stationRepo: stationRepo,
	}
}

// ExportType 导出类型
type ExportType string

const (
	ExportTypeAlarm   ExportType = "alarm"
	ExportTypeDevice  ExportType = "device"
	ExportTypeStation ExportType = "station"
)

// ExportFormat 导出格式
type ExportFormat string

const (
	ExportFormatExcel ExportFormat = "excel"
	ExportFormatCSV   ExportFormat = "csv"
)

// ExportRequest 导出请求
type ExportRequest struct {
	Type      ExportType      `json:"type" binding:"required"`
	Format    ExportFormat    `json:"format" binding:"required"`
	StartTime int64           `json:"start_time"`
	EndTime   int64           `json:"end_time"`
	Filters   map[string]interface{} `json:"filters"`
}

// ExportResult 导出结果
type ExportResult struct {
	Buffer      *bytes.Buffer
	Filename    string
	ContentType string
}

// Export 导出数据
func (s *ExportService) Export(ctx context.Context, req *ExportRequest) (*ExportResult, error) {
	switch req.Type {
	case ExportTypeAlarm:
		return s.exportAlarms(ctx, req)
	case ExportTypeDevice:
		return s.exportDevices(ctx, req)
	case ExportTypeStation:
		return s.exportStations(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported export type: %s", req.Type)
	}
}

// exportAlarms 导出告警数据
func (s *ExportService) exportAlarms(ctx context.Context, req *ExportRequest) (*ExportResult, error) {
	// 获取告警数据
	var alarms []*entity.Alarm
	var err error

	if req.StartTime > 0 && req.EndTime > 0 {
		// 导出历史告警
		var stationID *string
		if v, ok := req.Filters["station_id"].(string); ok {
			stationID = &v
		}
		alarms, err = s.alarmRepo.GetHistoryAlarms(ctx, stationID, req.StartTime, req.EndTime)
	} else {
		// 导出活跃告警
		var stationID *string
		var level *entity.AlarmLevel
		if v, ok := req.Filters["station_id"].(string); ok {
			stationID = &v
		}
		if v, ok := req.Filters["level"].(int); ok {
			l := entity.AlarmLevel(v)
			level = &l
		}
		alarms, err = s.alarmRepo.GetActiveAlarms(ctx, stationID, level)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get alarms: %w", err)
	}

	// 准备导出选项
	headers := []string{"ID", "设备ID", "厂站ID", "类型", "级别", "标题", "消息", "值", "阈值", "状态", "触发时间", "确认时间", "清除时间", "确认人"}
	fieldNames := []string{"ID", "DeviceID", "StationID", "Type", "Level", "Title", "Message", "Value", "Threshold", "Status", "TriggeredAt", "AcknowledgedAt", "ClearedAt", "AcknowledgedBy"}

	filename := fmt.Sprintf("alarms_%s.xlsx", time.Now().Format("20060102150405"))
	contentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

	var buf *bytes.Buffer

	switch req.Format {
	case ExportFormatExcel:
		opt := &export.ExcelOption{
			SheetName:  "告警数据",
			Headers:    headers,
			FieldNames: fieldNames,
			ColumnWidths: map[string]float64{
				"A": 36, // ID
				"B": 36, // DeviceID
				"C": 36, // StationID
				"D": 15, // Type
				"E": 10, // Level
				"F": 30, // Title
				"G": 50, // Message
				"H": 15, // Value
				"I": 15, // Threshold
				"J": 10, // Status
				"K": 20, // TriggeredAt
				"L": 20, // AcknowledgedAt
				"M": 20, // ClearedAt
				"N": 20, // AcknowledgedBy
			},
		}
		buf, err = export.Export(alarms, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to export excel: %w", err)
		}

	case ExportFormatCSV:
		filename = fmt.Sprintf("alarms_%s.csv", time.Now().Format("20060102150405"))
		contentType = "text/csv"
		opt := &export.CSVOption{
			Headers:    headers,
			FieldNames: fieldNames,
		}
		buf, err = export.ExportCSV(alarms, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to export csv: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported export format: %s", req.Format)
	}

	return &ExportResult{
		Buffer:      buf,
		Filename:    filename,
		ContentType: contentType,
	}, nil
}

// exportDevices 导出设备数据
func (s *ExportService) exportDevices(ctx context.Context, req *ExportRequest) (*ExportResult, error) {
	// 获取设备数据
	var devices []*entity.Device
	var err error

	var stationID *string
	var deviceType *entity.DeviceType
	if v, ok := req.Filters["station_id"].(string); ok {
		stationID = &v
	}
	if v, ok := req.Filters["type"].(string); ok {
		dt := entity.DeviceType(v)
		deviceType = &dt
	}

	devices, err = s.deviceRepo.List(ctx, stationID, deviceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	// 准备导出选项
	headers := []string{"ID", "编码", "名称", "类型", "厂站ID", "厂商", "型号", "序列号", "额定功率", "额定电压", "额定电流", "协议", "IP地址", "端口", "从站ID", "状态", "最后在线时间", "安装日期", "保修日期", "描述"}
	fieldNames := []string{"ID", "Code", "Name", "Type", "StationID", "Manufacturer", "Model", "SerialNumber", "RatedPower", "RatedVoltage", "RatedCurrent", "Protocol", "IPAddress", "Port", "SlaveID", "Status", "LastOnline", "InstallDate", "WarrantyDate", "Description"}

	filename := fmt.Sprintf("devices_%s.xlsx", time.Now().Format("20060102150405"))
	contentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

	var buf *bytes.Buffer

	switch req.Format {
	case ExportFormatExcel:
		opt := &export.ExcelOption{
			SheetName:  "设备数据",
			Headers:    headers,
			FieldNames: fieldNames,
			ColumnWidths: map[string]float64{
				"A": 36, // ID
				"B": 20, // Code
				"C": 30, // Name
				"D": 15, // Type
				"E": 36, // StationID
				"F": 20, // Manufacturer
				"G": 25, // Model
				"H": 20, // SerialNumber
				"I": 15, // RatedPower
				"J": 15, // RatedVoltage
				"K": 15, // RatedCurrent
				"L": 15, // Protocol
				"M": 20, // IPAddress
				"N": 10, // Port
				"O": 10, // SlaveID
				"P": 10, // Status
				"Q": 20, // LastOnline
				"R": 15, // InstallDate
				"S": 15, // WarrantyDate
				"T": 30, // Description
			},
		}
		buf, err = export.Export(devices, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to export excel: %w", err)
		}

	case ExportFormatCSV:
		filename = fmt.Sprintf("devices_%s.csv", time.Now().Format("20060102150405"))
		contentType = "text/csv"
		opt := &export.CSVOption{
			Headers:    headers,
			FieldNames: fieldNames,
		}
		buf, err = export.ExportCSV(devices, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to export csv: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported export format: %s", req.Format)
	}

	return &ExportResult{
		Buffer:      buf,
		Filename:    filename,
		ContentType: contentType,
	}, nil
}

// exportStations 导出厂站数据
func (s *ExportService) exportStations(ctx context.Context, req *ExportRequest) (*ExportResult, error) {
	// 获取厂站数据
	var stations []*entity.Station
	var err error

	var subRegionID *string
	var stationType *entity.StationType
	if v, ok := req.Filters["sub_region_id"].(string); ok {
		subRegionID = &v
	}
	if v, ok := req.Filters["type"].(string); ok {
		st := entity.StationType(v)
		stationType = &st
	}

	stations, err = s.stationRepo.List(ctx, subRegionID, stationType)
	if err != nil {
		return nil, fmt.Errorf("failed to get stations: %w", err)
	}

	// 准备导出选项
	headers := []string{"ID", "编码", "名称", "类型", "子区域ID", "容量", "电压等级", "经度", "纬度", "地址", "状态", "投运日期", "描述"}
	fieldNames := []string{"ID", "Code", "Name", "Type", "SubRegionID", "Capacity", "VoltageLevel", "Longitude", "Latitude", "Address", "Status", "CommissionDate", "Description"}

	filename := fmt.Sprintf("stations_%s.xlsx", time.Now().Format("20060102150405"))
	contentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

	var buf *bytes.Buffer

	switch req.Format {
	case ExportFormatExcel:
		opt := &export.ExcelOption{
			SheetName:  "厂站数据",
			Headers:    headers,
			FieldNames: fieldNames,
			ColumnWidths: map[string]float64{
				"A": 36, // ID
				"B": 20, // Code
				"C": 30, // Name
				"D": 15, // Type
				"E": 36, // SubRegionID
				"F": 15, // Capacity
				"G": 15, // VoltageLevel
				"H": 15, // Longitude
				"I": 15, // Latitude
				"J": 50, // Address
				"K": 10, // Status
				"L": 15, // CommissionDate
				"M": 30, // Description
			},
		}
		buf, err = export.Export(stations, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to export excel: %w", err)
		}

	case ExportFormatCSV:
		filename = fmt.Sprintf("stations_%s.csv", time.Now().Format("20060102150405"))
		contentType = "text/csv"
		opt := &export.CSVOption{
			Headers:    headers,
			FieldNames: fieldNames,
		}
		buf, err = export.ExportCSV(stations, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to export csv: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported export format: %s", req.Format)
	}

	return &ExportResult{
		Buffer:      buf,
		Filename:    filename,
		ContentType: contentType,
	}, nil
}

// StreamExportAlarms 流式导出告警数据（大数据量）
func (s *ExportService) StreamExportAlarms(ctx context.Context, req *ExportRequest, batchSize int) (*ExportResult, error) {
	headers := []string{"ID", "设备ID", "厂站ID", "类型", "级别", "标题", "消息", "值", "阈值", "状态", "触发时间", "确认时间", "清除时间", "确认人"}
	fieldNames := []string{"ID", "DeviceID", "StationID", "Type", "Level", "Title", "Message", "Value", "Threshold", "Status", "TriggeredAt", "AcknowledgedAt", "ClearedAt", "AcknowledgedBy"}

	filename := fmt.Sprintf("alarms_%s.xlsx", time.Now().Format("20060102150405"))
	contentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

	var buf *bytes.Buffer
	var err error

	switch req.Format {
	case ExportFormatExcel:
		opt := &export.ExcelOption{
			SheetName:  "告警数据",
			Headers:    headers,
			FieldNames: fieldNames,
			ColumnWidths: map[string]float64{
				"A": 36, "B": 36, "C": 36, "D": 15, "E": 10, "F": 30, "G": 50,
				"H": 15, "I": 15, "J": 10, "K": 20, "L": 20, "M": 20, "N": 20,
			},
		}

		streamExport, createErr := export.NewStreamExport(opt)
		if createErr != nil {
			return nil, fmt.Errorf("failed to create stream export: %w", createErr)
		}
		defer streamExport.Close()

		// 分批获取数据并导出
		// 这里简化处理，实际应该实现分页查询
		var stationID *string
		if v, ok := req.Filters["station_id"].(string); ok {
			stationID = &v
		}

		alarms, getErr := s.alarmRepo.GetHistoryAlarms(ctx, stationID, req.StartTime, req.EndTime)
		if getErr != nil {
			return nil, fmt.Errorf("failed to get alarms: %w", getErr)
		}

		if writeErr := streamExport.WriteBatch(alarms); writeErr != nil {
			return nil, fmt.Errorf("failed to write batch: %w", writeErr)
		}

		buf, err = streamExport.Finish()
		if err != nil {
			return nil, fmt.Errorf("failed to finish export: %w", err)
		}

	case ExportFormatCSV:
		filename = fmt.Sprintf("alarms_%s.csv", time.Now().Format("20060102150405"))
		contentType = "text/csv"
		opt := &export.CSVOption{
			Headers:    headers,
			FieldNames: fieldNames,
		}

		streamExport, createErr := export.NewStreamCSVExport(opt)
		if createErr != nil {
			return nil, fmt.Errorf("failed to create stream export: %w", createErr)
		}

		var stationID *string
		if v, ok := req.Filters["station_id"].(string); ok {
			stationID = &v
		}

		alarms, getErr := s.alarmRepo.GetHistoryAlarms(ctx, stationID, req.StartTime, req.EndTime)
		if getErr != nil {
			return nil, fmt.Errorf("failed to get alarms: %w", getErr)
		}

		if writeErr := streamExport.WriteBatch(alarms); writeErr != nil {
			return nil, fmt.Errorf("failed to write batch: %w", writeErr)
		}

		buf, err = streamExport.Finish()
		if err != nil {
			return nil, fmt.Errorf("failed to finish export: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported export format: %s", req.Format)
	}

	return &ExportResult{
		Buffer:      buf,
		Filename:    filename,
		ContentType: contentType,
	}, nil
}
