package dto

import "time"

// Response 通用响应结构
type Response struct {
	Code      int         `json:"code" example:"0"`
	Message   string      `json:"message" example:"success"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp" example:"1709500800000"`
}

// PagedResponse 分页响应结构
type PagedResponse struct {
	Code      int         `json:"code" example:"0"`
	Message   string      `json:"message" example:"success"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp" example:"1709500800000"`
	Total     int64       `json:"total" example:"100"`
	Page      int         `json:"page" example:"1"`
	PageSize  int         `json:"page_size" example:"20"`
}

// RegionResponse 区域响应
type RegionResponse struct {
	ID          string            `json:"id" example:"region-001"`
	Code        string            `json:"code" example:"EAST"`
	Name        string            `json:"name" example:"华东区域"`
	ParentID    *string           `json:"parent_id" example:"region-000"`
	Level       int               `json:"level" example:"1"`
	SortOrder   int               `json:"sort_order" example:"1"`
	Description string            `json:"description" example:"华东区域"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	SubRegions  []*RegionResponse `json:"sub_regions,omitempty"`
}

// StationResponse 厂站响应
type StationResponse struct {
	ID             string        `json:"id" example:"station-001"`
	Code           string        `json:"code" example:"PV_SH_001"`
	Name           string        `json:"name" example:"上海光伏电站1号"`
	Type           string        `json:"type" example:"pv"`
	SubRegionID    string        `json:"sub_region_id" example:"region-002"`
	Capacity       float64       `json:"capacity" example:"50.0"`
	VoltageLevel   string        `json:"voltage_level" example:"35kV"`
	Longitude      float64       `json:"longitude" example:"121.4737"`
	Latitude       float64       `json:"latitude" example:"31.2304"`
	Address        string        `json:"address" example:"上海市浦东新区"`
	Status         int           `json:"status" example:"1"`
	CommissionDate *time.Time    `json:"commission_date"`
	Description    string        `json:"description" example:"光伏电站"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Devices        []DeviceBrief `json:"devices,omitempty"`
}

// DeviceBrief 设备简要信息
type DeviceBrief struct {
	ID     string `json:"id" example:"device-001"`
	Code   string `json:"code" example:"INV_001"`
	Name   string `json:"name" example:"1号逆变器"`
	Type   string `json:"type" example:"inverter"`
	Status int    `json:"status" example:"1"`
}

// DeviceResponse 设备响应
type DeviceResponse struct {
	ID            string     `json:"id" example:"device-001"`
	Code          string     `json:"code" example:"INV_001"`
	Name          string     `json:"name" example:"1号逆变器"`
	Type          string     `json:"type" example:"inverter"`
	StationID     string     `json:"station_id" example:"station-001"`
	Manufacturer  string     `json:"manufacturer" example:"华为"`
	Model         string     `json:"model" example:"SUN2000-100KTL"`
	SerialNumber  string     `json:"serial_number" example:"SN123456"`
	RatedPower    float64    `json:"rated_power" example:"100.0"`
	RatedVoltage  float64    `json:"rated_voltage" example:"380.0"`
	RatedCurrent  float64    `json:"rated_current" example:"150.0"`
	Protocol      string     `json:"protocol" example:"modbus"`
	IPAddress     string     `json:"ip_address" example:"192.168.1.101"`
	Port          int        `json:"port" example:"502"`
	SlaveID       int        `json:"slave_id" example:"1"`
	Status        int        `json:"status" example:"1"`
	LastOnline    *time.Time `json:"last_online"`
	InstallDate   *time.Time `json:"install_date"`
	WarrantyDate  *time.Time `json:"warranty_date"`
	Description   string     `json:"description" example:"逆变器设备"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Points        []PointBrief `json:"points,omitempty"`
}

// PointBrief 采集点简要信息
type PointBrief struct {
	ID   string `json:"id" example:"point-001"`
	Code string `json:"code" example:"INV_001_P"`
	Name string `json:"name" example:"有功功率"`
	Type string `json:"type" example:"yaoc"`
}

// PointResponse 采集点响应
type PointResponse struct {
	ID           string    `json:"id" example:"point-001"`
	Code         string    `json:"code" example:"INV_001_P"`
	Name         string    `json:"name" example:"有功功率"`
	Type         string    `json:"type" example:"yaoc"`
	DeviceID     string    `json:"device_id" example:"device-001"`
	StationID    string    `json:"station_id" example:"station-001"`
	Unit         string    `json:"unit" example:"kW"`
	Precision    int       `json:"precision" example:"2"`
	MinValue     float64   `json:"min_value" example:"0.0"`
	MaxValue     float64   `json:"max_value" example:"100.0"`
	Protocol     string    `json:"protocol" example:"modbus"`
	Address      int       `json:"address" example:"40001"`
	DataFormat   string    `json:"data_format" example:"float32"`
	ScanInterval int       `json:"scan_interval" example:"1000"`
	Deadband     float64   `json:"deadband" example:"0.1"`
	IsAlarm      bool      `json:"is_alarm" example:"true"`
	AlarmHigh    float64   `json:"alarm_high" example:"95.0"`
	AlarmLow     float64   `json:"alarm_low" example:"0.0"`
	Status       int       `json:"status" example:"1"`
	Description  string    `json:"description" example:"有功功率采集点"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AlarmResponse 告警响应
type AlarmResponse struct {
	ID             string     `json:"id" example:"alarm-001"`
	PointID        string     `json:"point_id" example:"point-001"`
	DeviceID       string     `json:"device_id" example:"device-001"`
	StationID      string     `json:"station_id" example:"station-001"`
	Type           string     `json:"type" example:"limit"`
	Level          int        `json:"level" example:"3"`
	Title          string     `json:"title" example:"逆变器温度过高"`
	Message        string     `json:"message" example:"逆变器温度为87.5°C，已超过85°C阈值"`
	Value          float64    `json:"value" example:"87.5"`
	Threshold      float64    `json:"threshold" example:"85.0"`
	Status         int        `json:"status" example:"1"`
	TriggeredAt    time.Time  `json:"triggered_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	ClearedAt      *time.Time `json:"cleared_at"`
	AcknowledgedBy string     `json:"acknowledged_by" example:"admin"`
	CreatedAt      time.Time  `json:"created_at"`
}

// RealtimeDataResponse 实时数据响应
type RealtimeDataResponse struct {
	PointID   string  `json:"point_id" example:"point-001"`
	Value     float64 `json:"value" example:"85.6"`
	Quality   int     `json:"quality" example:"192"`
	Timestamp int64   `json:"timestamp" example:"1709500800000"`
}

// HistoryDataResponse 历史数据响应
type HistoryDataResponse struct {
	PointID   string  `json:"point_id" example:"point-001"`
	Value     float64 `json:"value" example:"85.6"`
	Quality   int     `json:"quality" example:"192"`
	Timestamp int64   `json:"timestamp" example:"1709500800000"`
}

// StatisticsResponse 统计数据响应
type StatisticsResponse struct {
	StationID    string  `json:"station_id" example:"station-001"`
	TotalPower   float64 `json:"total_power" example:"1250.5"`
	DailyEnergy  float64 `json:"daily_energy" example:"850.3"`
	MonthlyEnergy float64 `json:"monthly_energy" example:"25600.8"`
	PR           float64 `json:"pr" example:"0.85"`
	Availability float64 `json:"availability" example:"0.98"`
	Date         string  `json:"date" example:"2024-03-01"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID         string     `json:"id" example:"user-001"`
	Username   string     `json:"username" example:"admin"`
	Email      string     `json:"email" example:"admin@example.com"`
	Phone      string     `json:"phone" example:"13800138000"`
	RealName   string     `json:"real_name" example:"管理员"`
	Avatar     string     `json:"avatar" example:"https://example.com/avatar.jpg"`
	Status     int        `json:"status" example:"1"`
	LastLogin  *time.Time `json:"last_login"`
	LoginCount int        `json:"login_count" example:"10"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Roles      []string   `json:"roles" example:"admin,operator"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt int64        `json:"expires_at" example:"1709587200000"`
	User      UserResponse `json:"user"`
}

// AlarmStatisticsResponse 告警统计响应
type AlarmStatisticsResponse struct {
	Total       int64            `json:"total" example:"100"`
	Active      int64            `json:"active" example:"15"`
	Acknowledged int64           `json:"acknowledged" example:"20"`
	Cleared     int64            `json:"cleared" example:"65"`
	ByLevel     map[int]int64    `json:"by_level"`
	ByType      map[string]int64 `json:"by_type"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code      int    `json:"code" example:"400"`
	Message   string `json:"message" example:"请求参数错误"`
	Timestamp int64  `json:"timestamp" example:"1709500800000"`
}

// ProfileResponse 用户资料响应
type ProfileResponse struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Nickname   string    `json:"nickname"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Avatar     string    `json:"avatar"`
	Role       string    `json:"role"`
	Status     int       `json:"status"`
	CreateTime string    `json:"create_time"`
}

// PreferencesResponse 偏好设置响应
type PreferencesResponse struct {
	Theme            string   `json:"theme"`
	Language         string   `json:"language"`
	Timezone         string   `json:"timezone"`
	NotifyEnabled    bool     `json:"notify_enabled"`
	NotifyTypes      []string `json:"notify_types"`
	DashboardLayout  string   `json:"dashboard_layout"`
}

// UploadAvatarResponse 上传头像响应
type UploadAvatarResponse struct {
	Avatar string `json:"avatar"`
}

// AvatarResponse 头像响应
type AvatarResponse struct {
	Avatar string `json:"avatar"`
}

// StationStatisticsResponse 厂站统计响应
type StationStatisticsResponse struct {
	DeviceCount       int `json:"device_count" example:"25"`
	OnlineDeviceCount int `json:"online_device_count" example:"24"`
	OfflineDeviceCount int `json:"offline_device_count" example:"1"`
	AlarmCount        int     `json:"alarm_count" example:"5"`
	Power             float64 `json:"power" example:"4500.0"`
	Energy            float64 `json:"energy" example:"28500.0"`
}

// AlarmRuleResponse 告警规则响应
type AlarmRuleResponse struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Type           string   `json:"type"`
	Level          int      `json:"level"`
	Condition      string   `json:"condition"`
	Threshold      float64  `json:"threshold"`
	Duration       int      `json:"duration"`
	PointID        *string  `json:"point_id,omitempty"`
	DeviceID       *string  `json:"device_id,omitempty"`
	StationID      *string  `json:"station_id,omitempty"`
	NotifyChannels []string `json:"notify_channels"`
	NotifyUsers    []string `json:"notify_users"`
	Status         int      `json:"status"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

// NotificationConfigResponse 通知配置响应
type NotificationConfigResponse struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Name    string                 `json:"name"`
	Config  map[string]interface{} `json:"config"`
	Enabled bool                   `json:"enabled"`
}

// ReportResponse 报表响应
type ReportResponse struct {
	Type       string                   `json:"type"`
	StartTime  string                   `json:"start_time"`
	EndTime    string                   `json:"end_time"`
	Stations   []map[string]interface{} `json:"stations"`
	Summary    map[string]interface{}   `json:"summary"`
}

// OperationLogResponse 操作日志响应
type OperationLogResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Action     string `json:"action"`
	Resource   string `json:"resource"`
	ResourceID string `json:"resource_id"`
	RequestIP  string `json:"request_ip"`
	Status     int    `json:"status"`
	Duration   int64  `json:"duration"`
	CreatedAt  string `json:"created_at"`
}
