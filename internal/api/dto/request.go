package dto

// CreateRegionRequest 创建区域请求
type CreateRegionRequest struct {
	Code        string  `json:"code" binding:"required" example:"EAST_SH"`
	Name        string  `json:"name" binding:"required" example:"上海子区域"`
	ParentID    *string `json:"parent_id" example:"region-001"`
	Level       int     `json:"level" example:"2"`
	SortOrder   int     `json:"sort_order" example:"1"`
	Description string  `json:"description" example:"上海子区域"`
}

// UpdateRegionRequest 更新区域请求
type UpdateRegionRequest struct {
	Name        string `json:"name" binding:"required" example:"上海子区域"`
	SortOrder   int    `json:"sort_order" example:"1"`
	Description string `json:"description" example:"上海子区域"`
}

// CreateStationRequest 创建厂站请求
type CreateStationRequest struct {
	Code           string   `json:"code" binding:"required" example:"PV_SH_002"`
	Name           string   `json:"name" binding:"required" example:"上海光伏电站2号"`
	Type           string   `json:"type" binding:"required" example:"pv"`
	SubRegionID    string   `json:"sub_region_id" binding:"required" example:"region-002"`
	Capacity       float64  `json:"capacity" example:"100.0"`
	VoltageLevel   string   `json:"voltage_level" example:"35kV"`
	Longitude      float64  `json:"longitude" example:"121.4737"`
	Latitude       float64  `json:"latitude" example:"31.2304"`
	Address        string   `json:"address" example:"上海市浦东新区"`
	CommissionDate *string  `json:"commission_date" example:"2024-01-01"`
	Description    string   `json:"description" example:"光伏电站"`
}

// UpdateStationRequest 更新厂站请求
type UpdateStationRequest struct {
	Name           string  `json:"name" binding:"required" example:"上海光伏电站2号"`
	Capacity       float64 `json:"capacity" example:"100.0"`
	VoltageLevel   string  `json:"voltage_level" example:"35kV"`
	Longitude      float64 `json:"longitude" example:"121.4737"`
	Latitude       float64 `json:"latitude" example:"31.2304"`
	Address        string  `json:"address" example:"上海市浦东新区"`
	Description    string  `json:"description" example:"光伏电站"`
}

// CreateDeviceRequest 创建设备请求
type CreateDeviceRequest struct {
	Code          string  `json:"code" binding:"required" example:"INV_001"`
	Name          string  `json:"name" binding:"required" example:"1号逆变器"`
	Type          string  `json:"type" binding:"required" example:"inverter"`
	StationID     string  `json:"station_id" binding:"required" example:"station-001"`
	Manufacturer  string  `json:"manufacturer" example:"华为"`
	Model         string  `json:"model" example:"SUN2000-100KTL"`
	SerialNumber  string  `json:"serial_number" example:"SN123456"`
	RatedPower    float64 `json:"rated_power" example:"100.0"`
	RatedVoltage  float64 `json:"rated_voltage" example:"380.0"`
	RatedCurrent  float64 `json:"rated_current" example:"150.0"`
	Protocol      string  `json:"protocol" example:"modbus"`
	IPAddress     string  `json:"ip_address" example:"192.168.1.101"`
	Port          int     `json:"port" example:"502"`
	SlaveID       int     `json:"slave_id" example:"1"`
	InstallDate   *string `json:"install_date" example:"2024-01-01"`
	WarrantyDate  *string `json:"warranty_date" example:"2025-01-01"`
	Description   string  `json:"description" example:"逆变器设备"`
}

// UpdateDeviceRequest 更新设备请求
type UpdateDeviceRequest struct {
	Name          string  `json:"name" binding:"required" example:"1号逆变器"`
	Manufacturer  string  `json:"manufacturer" example:"华为"`
	Model         string  `json:"model" example:"SUN2000-100KTL"`
	SerialNumber  string  `json:"serial_number" example:"SN123456"`
	RatedPower    float64 `json:"rated_power" example:"100.0"`
	RatedVoltage  float64 `json:"rated_voltage" example:"380.0"`
	RatedCurrent  float64 `json:"rated_current" example:"150.0"`
	Protocol      string  `json:"protocol" example:"modbus"`
	IPAddress     string  `json:"ip_address" example:"192.168.1.101"`
	Port          int     `json:"port" example:"502"`
	SlaveID       int     `json:"slave_id" example:"1"`
	Description   string  `json:"description" example:"逆变器设备"`
}

// CreatePointRequest 创建采集点请求
type CreatePointRequest struct {
	Code         string  `json:"code" binding:"required" example:"INV_001_P"`
	Name         string  `json:"name" binding:"required" example:"有功功率"`
	Type         string  `json:"type" binding:"required" example:"yaoc"`
	DeviceID     string  `json:"device_id" example:"device-001"`
	StationID    string  `json:"station_id" example:"station-001"`
	Unit         string  `json:"unit" example:"kW"`
	Precision    int     `json:"precision" example:"2"`
	MinValue     float64 `json:"min_value" example:"0.0"`
	MaxValue     float64 `json:"max_value" example:"100.0"`
	Protocol     string  `json:"protocol" example:"modbus"`
	Address      int     `json:"address" example:"40001"`
	DataFormat   string  `json:"data_format" example:"float32"`
	ScanInterval int     `json:"scan_interval" example:"1000"`
	Deadband     float64 `json:"deadband" example:"0.1"`
	IsAlarm      bool    `json:"is_alarm" example:"true"`
	AlarmHigh    float64 `json:"alarm_high" example:"95.0"`
	AlarmLow     float64 `json:"alarm_low" example:"0.0"`
	Description  string  `json:"description" example:"有功功率采集点"`
}

// UpdatePointRequest 更新采集点请求
type UpdatePointRequest struct {
	Name         string  `json:"name" binding:"required" example:"有功功率"`
	Unit         string  `json:"unit" example:"kW"`
	Precision    int     `json:"precision" example:"2"`
	MinValue     float64 `json:"min_value" example:"0.0"`
	MaxValue     float64 `json:"max_value" example:"100.0"`
	ScanInterval int     `json:"scan_interval" example:"1000"`
	Deadband     float64 `json:"deadband" example:"0.1"`
	IsAlarm      bool    `json:"is_alarm" example:"true"`
	AlarmHigh    float64 `json:"alarm_high" example:"95.0"`
	AlarmLow     float64 `json:"alarm_low" example:"0.0"`
	Description  string  `json:"description" example:"有功功率采集点"`
}

// AckAlarmRequest 确认告警请求
type AckAlarmRequest struct {
	Operator string `json:"operator" binding:"required" example:"admin"`
	Comment  string `json:"comment" example:"已确认，正在处理"`
}

// ClearAlarmRequest 清除告警请求
type ClearAlarmRequest struct {
	Operator string `json:"operator" binding:"required" example:"admin"`
	Comment  string `json:"comment" example:"问题已解决"`
}

// GetRealtimeDataRequest 获取实时数据请求
type GetRealtimeDataRequest struct {
	PointIDs string `form:"point_ids" binding:"required" example:"point-001,point-002"`
}

// GetHistoryDataRequest 获取历史数据请求
type GetHistoryDataRequest struct {
	PointID   string `form:"point_id" binding:"required" example:"point-001"`
	StartTime int64  `form:"start_time" binding:"required" example:"1709414400000"`
	EndTime   int64  `form:"end_time" binding:"required" example:"1709500800000"`
	Interval  int    `form:"interval" example:"3600"`
}

// GetStatisticsRequest 获取统计数据请求
type GetStatisticsRequest struct {
	StationID string `form:"station_id" example:"station-001"`
	Type      string `form:"type" example:"daily"`
	StartTime int64  `form:"start_time" example:"1709414400000"`
	EndTime   int64  `form:"end_time" example:"1709500800000"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required" example:"user001"`
	Password string `json:"password" binding:"required" example:"password123"`
	Email    string `json:"email" example:"user@example.com"`
	Phone    string `json:"phone" example:"13800138000"`
	RealName string `json:"real_name" example:"张三"`
	Avatar   string `json:"avatar" example:"https://example.com/avatar.jpg"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Phone    string `json:"phone" example:"13800138000"`
	RealName string `json:"real_name" example:"张三"`
	Avatar   string `json:"avatar" example:"https://example.com/avatar.jpg"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpass123"`
	NewPassword string `json:"new_password" binding:"required" example:"newpass123"`
}

// ListRegionsRequest 获取区域列表请求
type ListRegionsRequest struct {
	ParentID string `form:"parent_id" example:"region-001"`
}

// ListStationsRequest 获取厂站列表请求
type ListStationsRequest struct {
	SubRegionID string `form:"sub_region_id" example:"region-002"`
	Type        string `form:"type" example:"pv"`
	Status      int    `form:"status" example:"1"`
	Page        int    `form:"page" example:"1"`
	PageSize    int    `form:"page_size" example:"20"`
}

// ListDevicesRequest 获取设备列表请求
type ListDevicesRequest struct {
	StationID string `form:"station_id" example:"station-001"`
	Type      string `form:"type" example:"inverter"`
	Status    int    `form:"status" example:"1"`
	Page      int    `form:"page" example:"1"`
	PageSize  int    `form:"page_size" example:"20"`
}

// ListPointsRequest 获取采集点列表请求
type ListPointsRequest struct {
	DeviceID string `form:"device_id" example:"device-001"`
	Type     string `form:"type" example:"yaoc"`
	Status   int    `form:"status" example:"1"`
	Page     int    `form:"page" example:"1"`
	PageSize int    `form:"page_size" example:"20"`
}

// ListAlarmsRequest 获取告警列表请求
type ListAlarmsRequest struct {
	StationID string `form:"station_id" example:"station-001"`
	DeviceID  string `form:"device_id" example:"device-001"`
	Level     int    `form:"level" example:"3"`
	Status    int    `form:"status" example:"1"`
	Type      string `form:"type" example:"limit"`
	StartTime int64  `form:"start_time" example:"1709414400000"`
	EndTime   int64  `form:"end_time" example:"1709500800000"`
	Page      int    `form:"page" example:"1"`
	PageSize  int    `form:"page_size" example:"20"`
}

// ControlOperateRequest 遥控操作请求
type ControlOperateRequest struct {
	PointID  string      `json:"point_id" binding:"required" example:"point-001"`
	Value    interface{} `json:"value" binding:"required"`
	Operator string      `json:"operator" binding:"required" example:"admin"`
	Reason   string      `json:"reason" example:"设备检修"`
}

// SetPointRequest 参数设置请求
type SetPointRequest struct {
	PointID  string  `json:"point_id" binding:"required" example:"point-001"`
	Value    float64 `json:"value" binding:"required" example:"50.0"`
	Operator string  `json:"operator" binding:"required" example:"admin"`
	Reason   string  `json:"reason" example:"调整参数"`
}

// AIQARequest AI问答请求
type AIQARequest struct {
	Question  string `json:"question" binding:"required" example:"1号逆变器为什么报警？"`
	SessionID string `json:"session_id" example:"session-001"`
	Context   string `json:"context" example:"逆变器温度过高"`
}

// AIQAResponse AI问答响应
type AIQAResponse struct {
	Answer     string  `json:"answer" example:"1号逆变器温度为87.5°C，已超过85°C阈值。"`
	Confidence float64 `json:"confidence" example:"0.95"`
	SessionID  string  `json:"session_id" example:"session-001"`
}

// AIConfigSuggestRequest AI配置建议请求
type AIConfigSuggestRequest struct {
	DeviceID string `json:"device_id" binding:"required" example:"device-001"`
	Type     string `json:"type" example:"alarm"`
	Context  string `json:"context" example:"设备温度过高"`
}

// AIConfigSuggestResponse AI配置建议响应
type AIConfigSuggestResponse struct {
	Suggestions []ConfigSuggestion `json:"suggestions"`
}

// ConfigSuggestion 配置建议
type ConfigSuggestion struct {
	Type          string      `json:"type" example:"alarm_threshold"`
	Name          string      `json:"name" example:"温度告警阈值"`
	CurrentValue  interface{} `json:"current_value"`
	SuggestedValue interface{} `json:"suggested_value"`
	Reason        string      `json:"reason" example:"根据历史数据分析，建议提高阈值以减少误报"`
	Priority      int         `json:"priority" example:"1"`
}

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" example:"管理员"`
	Email    string `json:"email" example:"admin@example.com"`
	Phone    string `json:"phone" example:"13800138000"`
}

// UpdatePreferencesRequest 更新偏好设置请求
type UpdatePreferencesRequest struct {
	Theme            string   `json:"theme" example:"light"`
	Language         string   `json:"language" example:"zh-CN"`
	Timezone         string   `json:"timezone" example:"Asia/Shanghai"`
	NotifyEnabled    bool     `json:"notify_enabled"`
	NotifyTypes      []string `json:"notify_types"`
	DashboardLayout  string   `json:"dashboard_layout" example:"default"`
}

// UploadAvatarRequest 上传头像请求
type UploadAvatarRequest struct {
	Avatar string `json:"avatar" binding:"required" example:"data:image/png;base64,..."`
}
