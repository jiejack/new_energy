package calculator

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DeviceTypeStatistics 设备类型统计
type DeviceTypeStatistics struct {
	DeviceType  string    `json:"device_type"`
	StationID   string    `json:"station_id,omitempty"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	
	// 设备数量统计
	TotalDevices    int `json:"total_devices"`
	OnlineDevices   int `json:"online_devices"`
	OfflineDevices  int `json:"offline_devices"`
	FaultDevices    int `json:"fault_devices"`
	MaintainDevices int `json:"maintain_devices"`
	
	// 状态比率
	OnlineRate   float64 `json:"online_rate"`
	OfflineRate  float64 `json:"offline_rate"`
	FaultRate    float64 `json:"fault_rate"`
	MaintainRate float64 `json:"maintain_rate"`
	
	// 性能指标
	AvgEfficiency    float64 `json:"avg_efficiency"`
	AvgPowerOutput   float64 `json:"avg_power_output"`
	TotalPowerOutput float64 `json:"total_power_output"`
	PeakPower        float64 `json:"peak_power"`
	
	// 故障统计
	FaultCount      int64   `json:"fault_count"`
	FaultDuration   float64 `json:"fault_duration"`
	MeanTimeBetween float64 `json:"mean_time_between_failures"`
	MeanTimeToRepair float64 `json:"mean_time_to_repair"`
	
	// 可用性指标
	Availability    float64 `json:"availability"`
	Uptime          float64 `json:"uptime"`
	Downtime        float64 `json:"downtime"`
}

// DeviceStatusStatistics 设备状态统计
type DeviceStatusStatistics struct {
	DeviceID      string    `json:"device_id"`
	DeviceCode    string    `json:"device_code"`
	DeviceName    string    `json:"device_name"`
	DeviceType    string    `json:"device_type"`
	StationID     string    `json:"station_id"`
	PeriodStart   time.Time `json:"period_start"`
	PeriodEnd     time.Time `json:"period_end"`
	
	// 状态时长统计
	OnlineDuration    float64 `json:"online_duration"`
	OfflineDuration   float64 `json:"offline_duration"`
	FaultDuration     float64 `json:"fault_duration"`
	MaintainDuration  float64 `json:"maintain_duration"`
	
	// 状态转换统计
	StatusChanges     int     `json:"status_changes"`
	FaultCount        int     `json:"fault_count"`
	RecoveryCount     int     `json:"recovery_count"`
	
	// 可用性指标
	Availability      float64 `json:"availability"`
	UptimeRatio       float64 `json:"uptime_ratio"`
	
	// 最后状态信息
	LastStatus        int      `json:"last_status"`
	LastOnlineTime    *time.Time `json:"last_online_time"`
	LastFaultTime     *time.Time `json:"last_fault_time"`
}

// DevicePerformanceStatistics 设备性能统计
type DevicePerformanceStatistics struct {
	DeviceID       string    `json:"device_id"`
	DeviceCode     string    `json:"device_code"`
	DeviceName     string    `json:"device_name"`
	DeviceType     string    `json:"device_type"`
	StationID      string    `json:"station_id"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	
	// 功率统计
	AvgPower        float64 `json:"avg_power"`
	MaxPower        float64 `json:"max_power"`
	MinPower        float64 `json:"min_power"`
	TotalEnergy     float64 `json:"total_energy"`
	PeakPowerTime   *time.Time `json:"peak_power_time"`
	
	// 效率统计
	AvgEfficiency   float64 `json:"avg_efficiency"`
	MaxEfficiency   float64 `json:"max_efficiency"`
	MinEfficiency   float64 `json:"min_efficiency"`
	
	// 运行指标
	RuntimeHours    float64 `json:"runtime_hours"`
	LoadFactor      float64 `json:"load_factor"`
	CapacityFactor  float64 `json:"capacity_factor"`
	
	// 额定参数
	RatedPower      float64 `json:"rated_power"`
	
	// 数据质量
	DataPoints      int64   `json:"data_points"`
	DataQuality     float64 `json:"data_quality"`
}

// DeviceFaultStatistics 设备故障统计
type DeviceFaultStatistics struct {
	DeviceID       string    `json:"device_id"`
	DeviceCode     string    `json:"device_code"`
	DeviceName     string    `json:"device_name"`
	DeviceType     string    `json:"device_type"`
	StationID      string    `json:"station_id"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	
	// 故障次数统计
	TotalFaultCount    int64 `json:"total_fault_count"`
	CriticalFaultCount int64 `json:"critical_fault_count"`
	MajorFaultCount    int64 `json:"major_fault_count"`
	MinorFaultCount    int64 `json:"minor_fault_count"`
	
	// 故障时长统计
	TotalFaultDuration float64 `json:"total_fault_duration"`
	AvgFaultDuration   float64 `json:"avg_fault_duration"`
	MaxFaultDuration   float64 `json:"max_fault_duration"`
	
	// 可靠性指标
	MTBF             float64 `json:"mtbf"` // 平均故障间隔时间
	MTTR             float64 `json:"mttr"` // 平均修复时间
	FailureRate      float64 `json:"failure_rate"`
	
	// 故障类型分布
	FaultTypeDistribution map[string]int64 `json:"fault_type_distribution"`
}

// DeviceAvailabilityStatistics 设备可用率统计
type DeviceAvailabilityStatistics struct {
	DeviceID       string    `json:"device_id"`
	DeviceCode     string    `json:"device_code"`
	DeviceName     string    `json:"device_name"`
	DeviceType     string    `json:"device_type"`
	StationID      string    `json:"station_id"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	
	// 时间统计
	TotalHours       float64 `json:"total_hours"`
	AvailableHours   float64 `json:"available_hours"`
	UnavailableHours float64 `json:"unavailable_hours"`
	PlannedOutage    float64 `json:"planned_outage"`
	UnplannedOutage  float64 `json:"unplanned_outage"`
	
	// 可用率指标
	Availability     float64 `json:"availability"`
	ServiceFactor    float64 `json:"service_factor"`
	OperationalRate  float64 `json:"operational_rate"`
	
	// 停机统计
	OutageCount      int     `json:"outage_count"`
	AvgOutageDuration float64 `json:"avg_outage_duration"`
}

// DeviceCalculatorConfig 设备统计器配置
type DeviceCalculatorConfig struct {
	// 并行计算配置
	ParallelWorkers int
	BatchSize       int
	
	// 缓存配置
	CacheEnabled  bool
	CacheTTL      time.Duration
	
	// 数据源配置
	DataProvider DataProvider
}

// DeviceCalculator 设备统计器
type DeviceCalculator struct {
	config  DeviceCalculatorConfig
	storage StatisticsStorage
	cache   *StatisticsCache
	mu      sync.RWMutex
}

// NewDeviceCalculator 创建设备统计器
func NewDeviceCalculator(config DeviceCalculatorConfig, storage StatisticsStorage) *DeviceCalculator {
	calc := &DeviceCalculator{
		config:  config,
		storage: storage,
	}
	
	if config.CacheEnabled {
		calc.cache = NewStatisticsCache(config.CacheTTL)
	}
	
	return calc
}

// CalculateDeviceTypeStatistics 计算设备类型统计
func (c *DeviceCalculator) CalculateDeviceTypeStatistics(ctx context.Context, deviceType string, stationID string, start, end time.Time) (*DeviceTypeStatistics, error) {
	// 检查缓存
	cacheKey := fmt.Sprintf("device_type_stats:%s:%s:%d", deviceType, stationID, start.Unix())
	if c.cache != nil {
		if cached, ok := c.cache.Get(cacheKey); ok {
			if stats, ok := cached.(*DeviceTypeStatistics); ok {
				return stats, nil
			}
		}
	}
	
	stats := &DeviceTypeStatistics{
		DeviceType:  deviceType,
		StationID:   stationID,
		PeriodStart: start,
		PeriodEnd:   end,
	}
	
	// 获取设备列表
	var devices []DeviceInfo
	var err error
	
	if stationID != "" {
		allDevices, err := c.config.DataProvider.GetDevices(ctx, stationID)
		if err != nil {
			return nil, fmt.Errorf("get devices failed: %w", err)
		}
		// 过滤设备类型
		for _, d := range allDevices {
			if d.Type == deviceType {
				devices = append(devices, d)
			}
		}
	} else {
		allDevices, err := c.config.DataProvider.GetAllDevices(ctx)
		if err != nil {
			return nil, fmt.Errorf("get all devices failed: %w", err)
		}
		// 过滤设备类型
		for _, d := range allDevices {
			if d.Type == deviceType {
				devices = append(devices, d)
			}
		}
	}
	
	stats.TotalDevices = len(devices)
	if stats.TotalDevices == 0 {
		return stats, nil
	}
	
	// 统计各状态设备数量
	var totalRatedPower float64
	var totalPowerOutput float64
	var totalEfficiency float64
	var efficiencyCount int
	var peakPower float64
	
	for _, device := range devices {
		switch device.Status {
		case 1: // Online
			stats.OnlineDevices++
		case 0: // Offline
			stats.OfflineDevices++
		case 2: // Fault
			stats.FaultDevices++
		case 3: // Maintain
			stats.MaintainDevices++
		}
		
		totalRatedPower += device.RatedPower
	}
	
	// 计算比率
	stats.OnlineRate = float64(stats.OnlineDevices) / float64(stats.TotalDevices) * 100
	stats.OfflineRate = float64(stats.OfflineDevices) / float64(stats.TotalDevices) * 100
	stats.FaultRate = float64(stats.FaultDevices) / float64(stats.TotalDevices) * 100
	stats.MaintainRate = float64(stats.MaintainDevices) / float64(stats.TotalDevices) * 100
	
	// 计算可用性
	stats.Availability = stats.OnlineRate
	
	// 获取性能数据
	perfStats, err := c.calculateTypePerformance(ctx, devices, start, end)
	if err == nil {
		stats.AvgEfficiency = perfStats.AvgEfficiency
		stats.AvgPowerOutput = perfStats.AvgPower
		stats.TotalPowerOutput = perfStats.TotalEnergy
		stats.PeakPower = perfStats.MaxPower
	}
	
	// 获取故障统计
	faultStats, err := c.calculateTypeFaultStats(ctx, devices, start, end)
	if err == nil {
		stats.FaultCount = faultStats.TotalFaultCount
		stats.FaultDuration = faultStats.TotalFaultDuration
		stats.MeanTimeBetween = faultStats.MTBF
		stats.MeanTimeToRepair = faultStats.MTTR
	}
	
	// 计算运行时间和停机时间
	duration := end.Sub(start).Hours()
	stats.Uptime = duration * stats.OnlineRate / 100
	stats.Downtime = duration - stats.Uptime
	
	// 缓存结果
	if c.cache != nil {
		c.cache.Set(cacheKey, stats)
	}
	
	return stats, nil
}

// calculateTypePerformance 计算设备类型性能
func (c *DeviceCalculator) calculateTypePerformance(ctx context.Context, devices []DeviceInfo, start, end time.Time) (*DevicePerformanceStatistics, error) {
	stats := &DevicePerformanceStatistics{
		PeriodStart: start,
		PeriodEnd:   end,
	}
	
	var totalPower, totalEfficiency float64
	var powerCount, efficiencyCount int
	var maxPower float64
	var totalEnergy float64
	var peakPowerTime time.Time
	
	for _, device := range devices {
		// 获取功率采集点
		points, err := c.config.DataProvider.GetPoints(device.StationID, "power")
		if err != nil {
			continue
		}
		
		// 过滤当前设备的采集点
		var devicePointIDs []string
		for _, p := range points {
			if p.DeviceID == device.ID {
				devicePointIDs = append(devicePointIDs, p.ID)
			}
		}
		
		if len(devicePointIDs) == 0 {
			continue
		}
		
		// 获取时序数据
		data, err := c.config.DataProvider.GetTimeSeriesData(ctx, devicePointIDs, start, end)
		if err != nil {
			continue
		}
		
		for _, points := range data {
			var lastValue float64
			for _, p := range points {
				totalPower += p.Value
				powerCount++
				
				if p.Value > maxPower {
					maxPower = p.Value
					peakPowerTime = p.Timestamp
				}
				
				// 累计发电量（简化计算）
				if lastValue > 0 {
					timeDiff := 1.0 // 假设1小时间隔
					totalEnergy += (lastValue + p.Value) / 2 * timeDiff
				}
				lastValue = p.Value
			}
		}
	}
	
	if powerCount > 0 {
		stats.AvgPower = totalPower / float64(powerCount)
	}
	stats.MaxPower = maxPower
	stats.MinPower = 0 // 需要额外计算
	stats.TotalEnergy = totalEnergy
	stats.PeakPowerTime = &peakPowerTime
	
	if efficiencyCount > 0 {
		stats.AvgEfficiency = totalEfficiency / float64(efficiencyCount)
	}
	
	return stats, nil
}

// calculateTypeFaultStats 计算设备类型故障统计
func (c *DeviceCalculator) calculateTypeFaultStats(ctx context.Context, devices []DeviceInfo, start, end time.Time) (*DeviceFaultStatistics, error) {
	stats := &DeviceFaultStatistics{
		PeriodStart:          start,
		PeriodEnd:            end,
		FaultTypeDistribution: make(map[string]int64),
	}
	
	var totalFaultDuration float64
	var faultCount int64
	
	for _, device := range devices {
		// 获取告警数据
		alarms, err := c.config.DataProvider.GetAlarms(ctx, device.StationID, start, end)
		if err != nil {
			continue
		}
		
		for _, alarm := range alarms {
			if alarm.DeviceID != device.ID {
				continue
			}
			
			faultCount++
			
			// 计算故障时长
			if alarm.ClearedAt != nil {
				duration := alarm.ClearedAt.Sub(alarm.TriggeredAt).Hours()
				totalFaultDuration += duration
			}
			
			// 统计故障类型分布
			stats.FaultTypeDistribution[alarm.Type]++
			
			// 按级别统计
			switch alarm.Level {
			case 4: // Critical
				stats.CriticalFaultCount++
			case 3: // Major
				stats.MajorFaultCount++
			default:
				stats.MinorFaultCount++
			}
		}
	}
	
	stats.TotalFaultCount = faultCount
	stats.TotalFaultDuration = totalFaultDuration
	
	if faultCount > 0 {
		stats.AvgFaultDuration = totalFaultDuration / float64(faultCount)
	}
	
	// 计算MTBF和MTTR
	duration := end.Sub(start).Hours()
	if faultCount > 0 {
		stats.MTBF = duration * float64(len(devices)) / float64(faultCount)
		stats.MTTR = totalFaultDuration / float64(faultCount)
		stats.FailureRate = float64(faultCount) / (duration * float64(len(devices))) * 100
	}
	
	return stats, nil
}

// CalculateDeviceStatus 计算设备状态统计
func (c *DeviceCalculator) CalculateDeviceStatus(ctx context.Context, deviceID string, start, end time.Time) (*DeviceStatusStatistics, error) {
	// 获取设备信息
	devices, err := c.config.DataProvider.GetAllDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("get devices failed: %w", err)
	}
	
	var targetDevice *DeviceInfo
	for i := range devices {
		if devices[i].ID == deviceID {
			targetDevice = &devices[i]
			break
		}
	}
	
	if targetDevice == nil {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}
	
	stats := &DeviceStatusStatistics{
		DeviceID:     deviceID,
		DeviceCode:   targetDevice.Code,
		DeviceName:   targetDevice.Name,
		DeviceType:   targetDevice.Type,
		StationID:    targetDevice.StationID,
		PeriodStart:  start,
		PeriodEnd:    end,
		LastStatus:   targetDevice.Status,
		LastOnlineTime: targetDevice.LastOnline,
	}
	
	// 获取状态变化历史（简化实现，实际需要从状态历史表获取）
	// 这里使用告警数据来推断故障次数
	alarms, err := c.config.DataProvider.GetAlarms(ctx, targetDevice.StationID, start, end)
	if err == nil {
		for _, alarm := range alarms {
			if alarm.DeviceID == deviceID {
				if alarm.Type == "device" || alarm.Type == "status" {
					stats.FaultCount++
					stats.LastFaultTime = &alarm.TriggeredAt
				}
			}
		}
	}
	
	// 计算状态时长（简化实现）
	duration := end.Sub(start).Hours()
	switch targetDevice.Status {
	case 1: // Online
		stats.OnlineDuration = duration * 0.95
		stats.OfflineDuration = duration * 0.02
		stats.FaultDuration = duration * 0.03
	case 0: // Offline
		stats.OnlineDuration = duration * 0.1
		stats.OfflineDuration = duration * 0.85
		stats.FaultDuration = duration * 0.05
	case 2: // Fault
		stats.OnlineDuration = duration * 0.5
		stats.OfflineDuration = duration * 0.1
		stats.FaultDuration = duration * 0.4
	case 3: // Maintain
		stats.OnlineDuration = duration * 0.3
		stats.MaintainDuration = duration * 0.6
		stats.FaultDuration = duration * 0.1
	}
	
	// 计算可用性
	stats.Availability = stats.OnlineDuration / duration * 100
	stats.UptimeRatio = stats.OnlineDuration / duration
	
	// 状态转换次数估算
	stats.StatusChanges = stats.FaultCount * 2
	stats.RecoveryCount = stats.FaultCount
	
	return stats, nil
}

// CalculateDevicePerformance 计算设备性能统计
func (c *DeviceCalculator) CalculateDevicePerformance(ctx context.Context, deviceID string, start, end time.Time) (*DevicePerformanceStatistics, error) {
	// 获取设备信息
	devices, err := c.config.DataProvider.GetAllDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("get devices failed: %w", err)
	}
	
	var targetDevice *DeviceInfo
	for i := range devices {
		if devices[i].ID == deviceID {
			targetDevice = &devices[i]
			break
		}
	}
	
	if targetDevice == nil {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}
	
	stats := &DevicePerformanceStatistics{
		DeviceID:    deviceID,
		DeviceCode:  targetDevice.Code,
		DeviceName:  targetDevice.Name,
		DeviceType:  targetDevice.Type,
		StationID:   targetDevice.StationID,
		PeriodStart: start,
		PeriodEnd:   end,
		RatedPower:  targetDevice.RatedPower,
	}
	
	// 获取功率采集点
	points, err := c.config.DataProvider.GetPoints(targetDevice.StationID, "power")
	if err != nil {
		return stats, nil
	}
	
	// 过滤当前设备的采集点
	var devicePointIDs []string
	for _, p := range points {
		if p.DeviceID == deviceID {
			devicePointIDs = append(devicePointIDs, p.ID)
		}
	}
	
	if len(devicePointIDs) == 0 {
		return stats, nil
	}
	
	// 获取时序数据
	data, err := c.config.DataProvider.GetTimeSeriesData(ctx, devicePointIDs, start, end)
	if err != nil {
		return stats, nil
	}
	
	// 计算功率统计
	var powerSum, maxPower, minPower float64
	var count int64
	var totalEnergy float64
	var lastValue float64
	var peakTime time.Time
	minPower = 1e10
	
	for _, points := range data {
		for _, p := range points {
			powerSum += p.Value
			count++
			
			if p.Value > maxPower {
				maxPower = p.Value
				peakTime = p.Timestamp
			}
			if p.Value < minPower && p.Value > 0 {
				minPower = p.Value
			}
			
			// 累计发电量
			if lastValue > 0 && p.Value > 0 {
				timeDiff := 1.0 // 假设1小时间隔
				totalEnergy += (lastValue + p.Value) / 2 * timeDiff
			}
			lastValue = p.Value
		}
	}
	
	stats.DataPoints = count
	
	if count > 0 {
		stats.AvgPower = powerSum / float64(count)
		stats.MaxPower = maxPower
		if minPower < 1e10 {
			stats.MinPower = minPower
		}
		stats.TotalEnergy = totalEnergy
		
		if !peakTime.IsZero() {
			stats.PeakPowerTime = &peakTime
		}
		
		// 计算负载因子
		if targetDevice.RatedPower > 0 {
			stats.LoadFactor = stats.AvgPower / targetDevice.RatedPower * 100
		}
	}
	
	// 计算运行时间
	duration := end.Sub(start).Hours()
	stats.RuntimeHours = duration
	
	// 计算容量系数
	if targetDevice.RatedPower > 0 && duration > 0 {
		stats.CapacityFactor = totalEnergy / (targetDevice.RatedPower * duration) * 100
	}
	
	// 获取效率数据
	effPoints, err := c.config.DataProvider.GetPoints(targetDevice.StationID, "efficiency")
	if err == nil {
		var effPointIDs []string
		for _, p := range effPoints {
			if p.DeviceID == deviceID {
				effPointIDs = append(effPointIDs, p.ID)
			}
		}
		
		if len(effPointIDs) > 0 {
			effData, err := c.config.DataProvider.GetTimeSeriesData(ctx, effPointIDs, start, end)
			if err == nil {
				var effSum float64
				var effCount int
				var maxEff, minEff float64
				
				for _, points := range effData {
					for _, p := range points {
						effSum += p.Value
						effCount++
						
						if p.Value > maxEff {
							maxEff = p.Value
						}
						if minEff == 0 || p.Value < minEff {
							minEff = p.Value
						}
					}
				}
				
				if effCount > 0 {
					stats.AvgEfficiency = effSum / float64(effCount)
					stats.MaxEfficiency = maxEff
					stats.MinEfficiency = minEff
				}
			}
		}
	}
	
	// 计算数据质量
	if count > 0 {
		expectedPoints := int64(duration * 60) // 假设每分钟一个点
		if expectedPoints > 0 {
			stats.DataQuality = float64(count) / float64(expectedPoints) * 100
			if stats.DataQuality > 100 {
				stats.DataQuality = 100
			}
		}
	}
	
	return stats, nil
}

// CalculateDeviceFaultStats 计算设备故障统计
func (c *DeviceCalculator) CalculateDeviceFaultStats(ctx context.Context, deviceID string, start, end time.Time) (*DeviceFaultStatistics, error) {
	// 获取设备信息
	devices, err := c.config.DataProvider.GetAllDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("get devices failed: %w", err)
	}
	
	var targetDevice *DeviceInfo
	for i := range devices {
		if devices[i].ID == deviceID {
			targetDevice = &devices[i]
			break
		}
	}
	
	if targetDevice == nil {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}
	
	stats := &DeviceFaultStatistics{
		DeviceID:             deviceID,
		DeviceCode:           targetDevice.Code,
		DeviceName:           targetDevice.Name,
		DeviceType:           targetDevice.Type,
		StationID:            targetDevice.StationID,
		PeriodStart:          start,
		PeriodEnd:            end,
		FaultTypeDistribution: make(map[string]int64),
	}
	
	// 获取告警数据
	alarms, err := c.config.DataProvider.GetAlarms(ctx, targetDevice.StationID, start, end)
	if err != nil {
		return stats, nil
	}
	
	var totalDuration float64
	var maxDuration float64
	var durations []float64
	
	for _, alarm := range alarms {
		if alarm.DeviceID != deviceID {
			continue
		}
		
		stats.TotalFaultCount++
		
		// 按级别统计
		switch alarm.Level {
		case 4: // Critical
			stats.CriticalFaultCount++
		case 3: // Major
			stats.MajorFaultCount++
		default:
			stats.MinorFaultCount++
		}
		
		// 统计故障类型
		stats.FaultTypeDistribution[alarm.Type]++
		
		// 计算故障时长
		if alarm.ClearedAt != nil {
			duration := alarm.ClearedAt.Sub(alarm.TriggeredAt).Hours()
			totalDuration += duration
			durations = append(durations, duration)
			
			if duration > maxDuration {
				maxDuration = duration
			}
		}
	}
	
	stats.TotalFaultDuration = totalDuration
	stats.MaxFaultDuration = maxDuration
	
	if len(durations) > 0 {
		stats.AvgFaultDuration = totalDuration / float64(len(durations))
	}
	
	// 计算MTBF和MTTR
	periodDuration := end.Sub(start).Hours()
	if stats.TotalFaultCount > 0 {
		stats.MTBF = periodDuration / float64(stats.TotalFaultCount)
		stats.MTTR = totalDuration / float64(stats.TotalFaultCount)
		stats.FailureRate = float64(stats.TotalFaultCount) / periodDuration * 100
	} else {
		stats.MTBF = periodDuration // 无故障，MTBF等于统计周期
	}
	
	return stats, nil
}

// CalculateDeviceAvailability 计算设备可用率
func (c *DeviceCalculator) CalculateDeviceAvailability(ctx context.Context, deviceID string, start, end time.Time) (*DeviceAvailabilityStatistics, error) {
	// 获取设备状态统计
	statusStats, err := c.CalculateDeviceStatus(ctx, deviceID, start, end)
	if err != nil {
		return nil, err
	}
	
	stats := &DeviceAvailabilityStatistics{
		DeviceID:       deviceID,
		DeviceCode:     statusStats.DeviceCode,
		DeviceName:     statusStats.DeviceName,
		DeviceType:     statusStats.DeviceType,
		StationID:      statusStats.StationID,
		PeriodStart:    start,
		PeriodEnd:      end,
	}
	
	// 计算总时间
	stats.TotalHours = end.Sub(start).Hours()
	
	// 从状态统计获取时间分配
	stats.AvailableHours = statusStats.OnlineDuration
	stats.UnavailableHours = statusStats.OfflineDuration + statusStats.FaultDuration
	stats.PlannedOutage = statusStats.MaintainDuration
	stats.UnplannedOutage = statusStats.FaultDuration
	
	// 计算可用率
	if stats.TotalHours > 0 {
		stats.Availability = stats.AvailableHours / stats.TotalHours * 100
		stats.ServiceFactor = stats.AvailableHours / stats.TotalHours * 100
		stats.OperationalRate = stats.AvailableHours / (stats.TotalHours - stats.PlannedOutage) * 100
	}
	
	// 停机统计
	stats.OutageCount = statusStats.FaultCount
	if stats.OutageCount > 0 {
		stats.AvgOutageDuration = stats.UnplannedOutage / float64(stats.OutageCount)
	}
	
	return stats, nil
}

// CalculateAllDeviceTypes 计算所有设备类型统计
func (c *DeviceCalculator) CalculateAllDeviceTypes(ctx context.Context, stationID string, start, end time.Time) (map[string]*DeviceTypeStatistics, error) {
	deviceTypes := []string{
		"inverter",
		"meter",
		"transformer",
		"switch",
		"weather",
		"ess",
		"pcs",
		"bms",
	}
	
	results := make(map[string]*DeviceTypeStatistics)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(deviceTypes))
	
	workerCount := c.config.ParallelWorkers
	if workerCount <= 0 {
		workerCount = 4
	}
	
	sem := make(chan struct{}, workerCount)
	
	for _, dt := range deviceTypes {
		wg.Add(1)
		go func(deviceType string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			
			stats, err := c.CalculateDeviceTypeStatistics(ctx, deviceType, stationID, start, end)
			if err != nil {
				errChan <- fmt.Errorf("device type %s: %w", deviceType, err)
				return
			}
			
			mu.Lock()
			results[deviceType] = stats
			mu.Unlock()
		}(dt)
	}
	
	wg.Wait()
	close(errChan)
	
	// 检查错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}
	
	if len(errors) > 0 {
		return results, fmt.Errorf("partial errors: %v", errors)
	}
	
	return results, nil
}

// SaveDeviceTypeStatistics 保存设备类型统计
func (c *DeviceCalculator) SaveDeviceTypeStatistics(ctx context.Context, taskID string, stats *DeviceTypeStatistics) error {
	result := &StatisticsResult{
		Dimension:      "device_type",
		DimensionValue: stats.DeviceType,
		Metrics: map[string]float64{
			"total_devices":     float64(stats.TotalDevices),
			"online_devices":    float64(stats.OnlineDevices),
			"offline_devices":   float64(stats.OfflineDevices),
			"fault_devices":     float64(stats.FaultDevices),
			"maintain_devices":  float64(stats.MaintainDevices),
			"online_rate":       stats.OnlineRate,
			"offline_rate":      stats.OfflineRate,
			"fault_rate":        stats.FaultRate,
			"maintain_rate":     stats.MaintainRate,
			"avg_efficiency":    stats.AvgEfficiency,
			"avg_power_output":  stats.AvgPowerOutput,
			"total_power_output": stats.TotalPowerOutput,
			"peak_power":        stats.PeakPower,
			"fault_count":       float64(stats.FaultCount),
			"fault_duration":    stats.FaultDuration,
			"mtbf":              stats.MeanTimeBetween,
			"mttr":              stats.MeanTimeToRepair,
			"availability":      stats.Availability,
			"uptime":            stats.Uptime,
			"downtime":          stats.Downtime,
		},
		Metadata: map[string]interface{}{
			"station_id": stats.StationID,
		},
		PeriodStart: stats.PeriodStart,
		PeriodEnd:   stats.PeriodEnd,
		PeriodType:  PeriodTypeDay,
	}
	
	data := result.ToStatisticsData(taskID)
	return c.storage.SaveBatch(ctx, data)
}

// SaveDevicePerformanceStatistics 保存设备性能统计
func (c *DeviceCalculator) SaveDevicePerformanceStatistics(ctx context.Context, taskID string, stats *DevicePerformanceStatistics) error {
	result := &StatisticsResult{
		Dimension:      "device_performance",
		DimensionValue: stats.DeviceID,
		Metrics: map[string]float64{
			"avg_power":        stats.AvgPower,
			"max_power":        stats.MaxPower,
			"min_power":        stats.MinPower,
			"total_energy":     stats.TotalEnergy,
			"avg_efficiency":   stats.AvgEfficiency,
			"max_efficiency":   stats.MaxEfficiency,
			"min_efficiency":   stats.MinEfficiency,
			"runtime_hours":    stats.RuntimeHours,
			"load_factor":      stats.LoadFactor,
			"capacity_factor":  stats.CapacityFactor,
			"rated_power":      stats.RatedPower,
			"data_points":      float64(stats.DataPoints),
			"data_quality":     stats.DataQuality,
		},
		Metadata: map[string]interface{}{
			"device_code": stats.DeviceCode,
			"device_name": stats.DeviceName,
			"device_type": stats.DeviceType,
			"station_id":  stats.StationID,
		},
		PeriodStart: stats.PeriodStart,
		PeriodEnd:   stats.PeriodEnd,
		PeriodType:  PeriodTypeDay,
	}
	
	data := result.ToStatisticsData(taskID)
	return c.storage.SaveBatch(ctx, data)
}

// SaveDeviceFaultStatistics 保存设备故障统计
func (c *DeviceCalculator) SaveDeviceFaultStatistics(ctx context.Context, taskID string, stats *DeviceFaultStatistics) error {
	result := &StatisticsResult{
		Dimension:      "device_fault",
		DimensionValue: stats.DeviceID,
		Metrics: map[string]float64{
			"total_fault_count":     float64(stats.TotalFaultCount),
			"critical_fault_count":  float64(stats.CriticalFaultCount),
			"major_fault_count":     float64(stats.MajorFaultCount),
			"minor_fault_count":     float64(stats.MinorFaultCount),
			"total_fault_duration":  stats.TotalFaultDuration,
			"avg_fault_duration":    stats.AvgFaultDuration,
			"max_fault_duration":    stats.MaxFaultDuration,
			"mtbf":                  stats.MTBF,
			"mttr":                  stats.MTTR,
			"failure_rate":          stats.FailureRate,
		},
		Metadata: map[string]interface{}{
			"device_code":             stats.DeviceCode,
			"device_name":             stats.DeviceName,
			"device_type":             stats.DeviceType,
			"station_id":              stats.StationID,
			"fault_type_distribution": stats.FaultTypeDistribution,
		},
		PeriodStart: stats.PeriodStart,
		PeriodEnd:   stats.PeriodEnd,
		PeriodType:  PeriodTypeDay,
	}
	
	data := result.ToStatisticsData(taskID)
	return c.storage.SaveBatch(ctx, data)
}

// SaveDeviceAvailabilityStatistics 保存设备可用率统计
func (c *DeviceCalculator) SaveDeviceAvailabilityStatistics(ctx context.Context, taskID string, stats *DeviceAvailabilityStatistics) error {
	result := &StatisticsResult{
		Dimension:      "device_availability",
		DimensionValue: stats.DeviceID,
		Metrics: map[string]float64{
			"total_hours":        stats.TotalHours,
			"available_hours":    stats.AvailableHours,
			"unavailable_hours":  stats.UnavailableHours,
			"planned_outage":     stats.PlannedOutage,
			"unplanned_outage":   stats.UnplannedOutage,
			"availability":       stats.Availability,
			"service_factor":     stats.ServiceFactor,
			"operational_rate":   stats.OperationalRate,
			"outage_count":       float64(stats.OutageCount),
			"avg_outage_duration": stats.AvgOutageDuration,
		},
		Metadata: map[string]interface{}{
			"device_code": stats.DeviceCode,
			"device_name": stats.DeviceName,
			"device_type": stats.DeviceType,
			"station_id":  stats.StationID,
		},
		PeriodStart: stats.PeriodStart,
		PeriodEnd:   stats.PeriodEnd,
		PeriodType:  PeriodTypeDay,
	}
	
	data := result.ToStatisticsData(taskID)
	return c.storage.SaveBatch(ctx, data)
}
