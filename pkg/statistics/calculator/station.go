package calculator

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// StationStatistics 厂站统计数据
type StationStatistics struct {
	StationID   string    `json:"station_id"`
	StationCode string    `json:"station_code"`
	StationName string    `json:"station_name"`
	StationType string    `json:"station_type"`
	Capacity    float64   `json:"capacity"`
	
	// 发电统计
	DailyGeneration   float64 `json:"daily_generation"`
	MonthlyGeneration float64 `json:"monthly_generation"`
	YearlyGeneration  float64 `json:"yearly_generation"`
	
	// 运行统计
	DeviceCount      int     `json:"device_count"`
	OnlineDeviceCount int    `json:"online_device_count"`
	DeviceRunRate    float64 `json:"device_run_rate"`
	
	// 告警统计
	TotalAlarmCount    int64 `json:"total_alarm_count"`
	ActiveAlarmCount   int64 `json:"active_alarm_count"`
	CriticalAlarmCount int64 `json:"critical_alarm_count"`
	
	// 效率指标
	Efficiency       float64 `json:"efficiency"`
	PerformanceRatio float64 `json:"performance_ratio"`
	
	// 等效利用小时
	EquivalentHours float64 `json:"equivalent_hours"`
	
	// 时间范围
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
}

// GenerationStatistics 发电量统计
type GenerationStatistics struct {
	StationID        string    `json:"station_id"`
	PeriodType       PeriodType `json:"period_type"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	TotalGeneration  float64   `json:"total_generation"`
	PeakPower        float64   `json:"peak_power"`
	AveragePower     float64   `json:"average_power"`
	GenerationHours  float64   `json:"generation_hours"`
	CapacityFactor   float64   `json:"capacity_factor"`
	DataPoints       int64     `json:"data_points"`
}

// DeviceRunRateStatistics 设备运行率统计
type DeviceRunRateStatistics struct {
	StationID         string    `json:"station_id"`
	PeriodStart       time.Time `json:"period_start"`
	PeriodEnd         time.Time `json:"period_end"`
	TotalDevices      int       `json:"total_devices"`
	OnlineDevices     int       `json:"online_devices"`
	FaultDevices      int       `json:"fault_devices"`
	MaintainDevices   int       `json:"maintain_devices"`
	RunRate           float64   `json:"run_rate"`
	OnlineRate        float64   `json:"online_rate"`
	FaultRate         float64   `json:"fault_rate"`
	MaintainRate      float64   `json:"maintain_rate"`
	AvgOnlineDuration float64   `json:"avg_online_duration"`
}

// AlarmStatistics 告警统计
type AlarmStatistics struct {
	StationID          string    `json:"station_id"`
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
	TotalCount         int64     `json:"total_count"`
	ActiveCount        int64     `json:"active_count"`
	AcknowledgedCount  int64     `json:"acknowledged_count"`
	ClearedCount       int64     `json:"cleared_count"`
	InfoCount          int64     `json:"info_count"`
	WarningCount       int64     `json:"warning_count"`
	MajorCount         int64     `json:"major_count"`
	CriticalCount      int64     `json:"critical_count"`
	AvgResponseTime    float64   `json:"avg_response_time"`
	AvgClearTime       float64   `json:"avg_clear_time"`
}

// EfficiencyStatistics 效率指标统计
type EfficiencyStatistics struct {
	StationID           string    `json:"station_id"`
	PeriodStart         time.Time `json:"period_start"`
	PeriodEnd           time.Time `json:"period_end"`
	SystemEfficiency    float64   `json:"system_efficiency"`
	InverterEfficiency  float64   `json:"inverter_efficiency"`
	TransformerEfficiency float64 `json:"transformer_efficiency"`
	LineLoss            float64   `json:"line_loss"`
	PerformanceRatio    float64   `json:"performance_ratio"`
	PR                  float64   `json:"pr"`
}

// EquivalentHoursStatistics 等效利用小时统计
type EquivalentHoursStatistics struct {
	StationID        string    `json:"station_id"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	TotalHours       float64   `json:"total_hours"`
	EquivalentHours  float64   `json:"equivalent_hours"`
	FullLoadHours    float64   `json:"full_load_hours"`
	PartialLoadHours float64   `json:"partial_load_hours"`
	CapacityFactor   float64   `json:"capacity_factor"`
	Availability     float64   `json:"availability"`
}

// StationCalculatorConfig 厂站统计器配置
type StationCalculatorConfig struct {
	// 并行计算配置
	ParallelWorkers int
	BatchSize       int
	
	// 缓存配置
	CacheEnabled  bool
	CacheTTL      time.Duration
	
	// 数据源配置
	DataProvider DataProvider
}

// DataProvider 数据提供者接口
type DataProvider interface {
	// 获取时序数据
	GetTimeSeriesData(ctx context.Context, pointIDs []string, start, end time.Time) (map[string][]TimeSeriesPoint, error)
	
	// 获取设备数据
	GetDevices(ctx context.Context, stationID string) ([]DeviceInfo, error)
	GetAllDevices(ctx context.Context) ([]DeviceInfo, error)
	
	// 获取厂站数据
	GetStation(ctx context.Context, stationID string) (*StationInfo, error)
	GetAllStations(ctx context.Context) ([]StationInfo, error)
	
	// 获取告警数据
	GetAlarms(ctx context.Context, stationID string, start, end time.Time) ([]AlarmInfo, error)
	
	// 获取采集点数据
	GetPoints(ctx context.Context, stationID string, pointType string) ([]PointInfo, error)
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	StationID    string `json:"station_id"`
	RatedPower   float64 `json:"rated_power"`
	Status       int    `json:"status"`
	LastOnline   *time.Time `json:"last_online"`
}

// StationInfo 厂站信息
type StationInfo struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Capacity  float64   `json:"capacity"`
	Status    int       `json:"status"`
}

// AlarmInfo 告警信息
type AlarmInfo struct {
	ID           string     `json:"id"`
	StationID    string     `json:"station_id"`
	DeviceID     string     `json:"device_id"`
	Type         string     `json:"type"`
	Level        int        `json:"level"`
	Status       int        `json:"status"`
	TriggeredAt  time.Time  `json:"triggered_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	ClearedAt    *time.Time `json:"cleared_at"`
}

// PointInfo 采集点信息
type PointInfo struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	DeviceID string `json:"device_id"`
	Unit     string `json:"unit"`
}

// StationCalculator 厂站统计器
type StationCalculator struct {
	config  StationCalculatorConfig
	storage StatisticsStorage
	cache   *StatisticsCache
	mu      sync.RWMutex
}

// NewStationCalculator 创建厂站统计器
func NewStationCalculator(config StationCalculatorConfig, storage StatisticsStorage) *StationCalculator {
	calc := &StationCalculator{
		config:  config,
		storage: storage,
	}
	
	if config.CacheEnabled {
		calc.cache = NewStatisticsCache(config.CacheTTL)
	}
	
	return calc
}

// CalculateStationStatistics 计算厂站综合统计
func (c *StationCalculator) CalculateStationStatistics(ctx context.Context, stationID string, periodType PeriodType, start, end time.Time) (*StationStatistics, error) {
	// 检查缓存
	cacheKey := fmt.Sprintf("station_stats:%s:%s:%d", stationID, periodType, start.Unix())
	if c.cache != nil {
		if cached, ok := c.cache.Get(cacheKey); ok {
			if stats, ok := cached.(*StationStatistics); ok {
				return stats, nil
			}
		}
	}
	
	// 获取厂站信息
	station, err := c.config.DataProvider.GetStation(ctx, stationID)
	if err != nil {
		return nil, fmt.Errorf("get station failed: %w", err)
	}
	
	stats := &StationStatistics{
		StationID:   stationID,
		StationCode: station.Code,
		StationName: station.Name,
		StationType: station.Type,
		Capacity:    station.Capacity,
		PeriodStart: start,
		PeriodEnd:   end,
	}
	
	// 并行计算各项统计
	var wg sync.WaitGroup
	errChan := make(chan error, 5)
	
	// 计算发电量统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		genStats, err := c.CalculateGeneration(ctx, stationID, periodType, start, end)
		if err != nil {
			errChan <- fmt.Errorf("calculate generation failed: %w", err)
			return
		}
		stats.DailyGeneration = genStats.TotalGeneration
		if periodType == PeriodTypeMonth {
			stats.MonthlyGeneration = genStats.TotalGeneration
		} else if periodType == PeriodTypeYear {
			stats.YearlyGeneration = genStats.TotalGeneration
		}
		stats.EquivalentHours = genStats.GenerationHours
	}()
	
	// 计算设备运行率
	wg.Add(1)
	go func() {
		defer wg.Done()
		runStats, err := c.CalculateDeviceRunRate(ctx, stationID, start, end)
		if err != nil {
			errChan <- fmt.Errorf("calculate device run rate failed: %w", err)
			return
		}
		stats.DeviceCount = runStats.TotalDevices
		stats.OnlineDeviceCount = runStats.OnlineDevices
		stats.DeviceRunRate = runStats.RunRate
	}()
	
	// 计算告警统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		alarmStats, err := c.CalculateAlarmStats(ctx, stationID, start, end)
		if err != nil {
			errChan <- fmt.Errorf("calculate alarm stats failed: %w", err)
			return
		}
		stats.TotalAlarmCount = alarmStats.TotalCount
		stats.ActiveAlarmCount = alarmStats.ActiveCount
		stats.CriticalAlarmCount = alarmStats.CriticalCount
	}()
	
	// 计算效率指标
	wg.Add(1)
	go func() {
		defer wg.Done()
		effStats, err := c.CalculateEfficiency(ctx, stationID, start, end)
		if err != nil {
			errChan <- fmt.Errorf("calculate efficiency failed: %w", err)
			return
		}
		stats.Efficiency = effStats.SystemEfficiency
		stats.PerformanceRatio = effStats.PerformanceRatio
	}()
	
	// 计算等效利用小时
	wg.Add(1)
	go func() {
		defer wg.Done()
		hoursStats, err := c.CalculateEquivalentHours(ctx, stationID, start, end)
		if err != nil {
			errChan <- fmt.Errorf("calculate equivalent hours failed: %w", err)
			return
		}
		stats.EquivalentHours = hoursStats.EquivalentHours
	}()
	
	wg.Wait()
	close(errChan)
	
	// 检查错误
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}
	
	// 缓存结果
	if c.cache != nil {
		c.cache.Set(cacheKey, stats)
	}
	
	return stats, nil
}

// CalculateGeneration 计算发电量统计
func (c *StationCalculator) CalculateGeneration(ctx context.Context, stationID string, periodType PeriodType, start, end time.Time) (*GenerationStatistics, error) {
	stats := &GenerationStatistics{
		StationID:   stationID,
		PeriodType:  periodType,
		PeriodStart: start,
		PeriodEnd:   end,
	}
	
	// 获取发电量采集点
	points, err := c.config.DataProvider.GetPoints(ctx, stationID, "generation")
	if err != nil {
		return nil, fmt.Errorf("get generation points failed: %w", err)
	}
	
	if len(points) == 0 {
		return stats, nil
	}
	
	pointIDs := make([]string, len(points))
	for i, p := range points {
		pointIDs[i] = p.ID
	}
	
	// 获取时序数据
	data, err := c.config.DataProvider.GetTimeSeriesData(ctx, pointIDs, start, end)
	if err != nil {
		return nil, fmt.Errorf("get time series data failed: %w", err)
	}
	
	// 计算发电量
	var totalGeneration float64
	var peakPower float64
	var powerSum float64
	var dataPoints int64
	var lastValues = make(map[string]float64)
	
	for pointID, points := range data {
		for _, p := range points {
			dataPoints++
			powerSum += p.Value
			
			if p.Value > peakPower {
				peakPower = p.Value
			}
			
			// 累计发电量计算（假设数据是功率，需要积分）
			if lastVal, ok := lastValues[pointID]; ok {
				// 简化计算：使用梯形法则积分
				timeDiff := 1.0 // 假设采样间隔为1小时
				totalGeneration += (lastVal + p.Value) / 2 * timeDiff
			}
			lastValues[pointID] = p.Value
		}
	}
	
	stats.TotalGeneration = totalGeneration
	stats.PeakPower = peakPower
	stats.DataPoints = dataPoints
	
	if dataPoints > 0 {
		stats.AveragePower = powerSum / float64(dataPoints)
	}
	
	// 获取厂站容量
	station, err := c.config.DataProvider.GetStation(ctx, stationID)
	if err == nil && station.Capacity > 0 {
		// 计算发电小时数
		duration := end.Sub(start).Hours()
		stats.GenerationHours = totalGeneration / station.Capacity
		
		// 计算容量系数
		if duration > 0 {
			stats.CapacityFactor = totalGeneration / (station.Capacity * duration)
		}
	}
	
	return stats, nil
}

// CalculateDeviceRunRate 计算设备运行率
func (c *StationCalculator) CalculateDeviceRunRate(ctx context.Context, stationID string, start, end time.Time) (*DeviceRunRateStatistics, error) {
	stats := &DeviceRunRateStatistics{
		StationID:   stationID,
		PeriodStart: start,
		PeriodEnd:   end,
	}
	
	// 获取设备列表
	devices, err := c.config.DataProvider.GetDevices(ctx, stationID)
	if err != nil {
		return nil, fmt.Errorf("get devices failed: %w", err)
	}
	
	stats.TotalDevices = len(devices)
	if stats.TotalDevices == 0 {
		return stats, nil
	}
	
	// 统计各状态设备数量
	var onlineDuration float64
	var onlineCount int
	
	for _, device := range devices {
		switch device.Status {
		case 1: // Online
			stats.OnlineDevices++
			onlineCount++
			if device.LastOnline != nil {
				onlineDuration += time.Since(*device.LastOnline).Hours()
			}
		case 2: // Fault
			stats.FaultDevices++
		case 3: // Maintain
			stats.MaintainDevices++
		}
	}
	
	// 计算比率
	stats.RunRate = float64(stats.OnlineDevices) / float64(stats.TotalDevices) * 100
	stats.OnlineRate = float64(stats.OnlineDevices) / float64(stats.TotalDevices) * 100
	stats.FaultRate = float64(stats.FaultDevices) / float64(stats.TotalDevices) * 100
	stats.MaintainRate = float64(stats.MaintainDevices) / float64(stats.TotalDevices) * 100
	
	// 计算平均在线时长
	if onlineCount > 0 {
		stats.AvgOnlineDuration = onlineDuration / float64(onlineCount)
	}
	
	return stats, nil
}

// CalculateAlarmStats 计算告警统计
func (c *StationCalculator) CalculateAlarmStats(ctx context.Context, stationID string, start, end time.Time) (*AlarmStatistics, error) {
	stats := &AlarmStatistics{
		StationID:   stationID,
		PeriodStart: start,
		PeriodEnd:   end,
	}
	
	// 获取告警数据
	alarms, err := c.config.DataProvider.GetAlarms(ctx, stationID, start, end)
	if err != nil {
		return nil, fmt.Errorf("get alarms failed: %w", err)
	}
	
	stats.TotalCount = int64(len(alarms))
	if stats.TotalCount == 0 {
		return stats, nil
	}
	
	var responseTimeSum float64
	var clearTimeSum float64
	var responseCount, clearCount int
	
	for _, alarm := range alarms {
		// 按状态统计
		switch alarm.Status {
		case 1: // Active
			stats.ActiveCount++
		case 2: // Acknowledged
			stats.AcknowledgedCount++
		case 3: // Cleared
			stats.ClearedCount++
		}
		
		// 按级别统计
		switch alarm.Level {
		case 1: // Info
			stats.InfoCount++
		case 2: // Warning
			stats.WarningCount++
		case 3: // Major
			stats.MajorCount++
		case 4: // Critical
			stats.CriticalCount++
		}
		
		// 计算响应时间
		if alarm.AcknowledgedAt != nil {
			responseTime := alarm.AcknowledgedAt.Sub(alarm.TriggeredAt).Minutes()
			responseTimeSum += responseTime
			responseCount++
		}
		
		// 计算清除时间
		if alarm.ClearedAt != nil {
			clearTime := alarm.ClearedAt.Sub(alarm.TriggeredAt).Minutes()
			clearTimeSum += clearTime
			clearCount++
		}
	}
	
	// 计算平均响应时间和清除时间
	if responseCount > 0 {
		stats.AvgResponseTime = responseTimeSum / float64(responseCount)
	}
	if clearCount > 0 {
		stats.AvgClearTime = clearTimeSum / float64(clearCount)
	}
	
	return stats, nil
}

// CalculateEfficiency 计算效率指标
func (c *StationCalculator) CalculateEfficiency(ctx context.Context, stationID string, start, end time.Time) (*EfficiencyStatistics, error) {
	stats := &EfficiencyStatistics{
		StationID:   stationID,
		PeriodStart: start,
		PeriodEnd:   end,
	}
	
	// 获取效率相关采集点
	effPoints, err := c.config.DataProvider.GetPoints(ctx, stationID, "efficiency")
	if err != nil {
		return nil, fmt.Errorf("get efficiency points failed: %w", err)
	}
	
	if len(effPoints) == 0 {
		// 如果没有效率采集点，使用默认计算方法
		return c.calculateDefaultEfficiency(ctx, stationID, start, end)
	}
	
	// 获取时序数据
	pointIDs := make([]string, len(effPoints))
	pointMap := make(map[string]string)
	for i, p := range effPoints {
		pointIDs[i] = p.ID
		pointMap[p.ID] = p.Code
	}
	
	data, err := c.config.DataProvider.GetTimeSeriesData(ctx, pointIDs, start, end)
	if err != nil {
		return nil, fmt.Errorf("get time series data failed: %w", err)
	}
	
	// 计算各项效率指标
	var systemEffSum, inverterEffSum, transformerEffSum float64
	var systemCount, inverterCount, transformerCount int
	
	for pointID, points := range data {
		code := pointMap[pointID]
		var sum float64
		var count int
		
		for _, p := range points {
			sum += p.Value
			count++
		}
		
		avg := 0.0
		if count > 0 {
			avg = sum / float64(count)
		}
		
		// 根据采集点编码分类
		switch {
		case contains(code, "system_eff"):
			systemEffSum += avg
			systemCount++
		case contains(code, "inverter_eff"):
			inverterEffSum += avg
			inverterCount++
		case contains(code, "transformer_eff"):
			transformerEffSum += avg
			transformerCount++
		}
	}
	
	if systemCount > 0 {
		stats.SystemEfficiency = systemEffSum / float64(systemCount)
	}
	if inverterCount > 0 {
		stats.InverterEfficiency = inverterEffSum / float64(inverterCount)
	}
	if transformerCount > 0 {
		stats.TransformerEfficiency = transformerEffSum / float64(transformerCount)
	}
	
	// 计算性能比
	stats.PerformanceRatio = stats.SystemEfficiency
	stats.PR = stats.SystemEfficiency
	
	return stats, nil
}

// calculateDefaultEfficiency 默认效率计算方法
func (c *StationCalculator) calculateDefaultEfficiency(ctx context.Context, stationID string, start, end time.Time) (*EfficiencyStatistics, error) {
	stats := &EfficiencyStatistics{
		StationID:   stationID,
		PeriodStart: start,
		PeriodEnd:   end,
	}
	
	// 获取输入输出功率采集点
	inputPoints, err := c.config.DataProvider.GetPoints(ctx, stationID, "input_power")
	if err != nil {
		return stats, nil
	}
	
	outputPoints, err := c.config.DataProvider.GetPoints(ctx, stationID, "output_power")
	if err != nil {
		return stats, nil
	}
	
	if len(inputPoints) == 0 || len(outputPoints) == 0 {
		return stats, nil
	}
	
	// 获取时序数据
	inputIDs := make([]string, len(inputPoints))
	for i, p := range inputPoints {
		inputIDs[i] = p.ID
	}
	
	outputIDs := make([]string, len(outputPoints))
	for i, p := range outputPoints {
		outputIDs[i] = p.ID
	}
	
	inputData, err := c.config.DataProvider.GetTimeSeriesData(ctx, inputIDs, start, end)
	if err != nil {
		return stats, nil
	}
	
	outputData, err := c.config.DataProvider.GetTimeSeriesData(ctx, outputIDs, start, end)
	if err != nil {
		return stats, nil
	}
	
	// 计算平均效率
	var inputSum, outputSum float64
	var count int
	
	for _, points := range inputData {
		for _, p := range points {
			inputSum += p.Value
			count++
		}
	}
	
	for _, points := range outputData {
		for _, p := range points {
			outputSum += p.Value
		}
	}
	
	if inputSum > 0 {
		stats.SystemEfficiency = outputSum / inputSum * 100
		stats.PerformanceRatio = stats.SystemEfficiency
		stats.PR = stats.SystemEfficiency
	}
	
	return stats, nil
}

// CalculateEquivalentHours 计算等效利用小时
func (c *StationCalculator) CalculateEquivalentHours(ctx context.Context, stationID string, start, end time.Time) (*EquivalentHoursStatistics, error) {
	stats := &EquivalentHoursStatistics{
		StationID:   stationID,
		PeriodStart: start,
		PeriodEnd:   end,
	}
	
	// 获取厂站信息
	station, err := c.config.DataProvider.GetStation(ctx, stationID)
	if err != nil {
		return nil, fmt.Errorf("get station failed: %w", err)
	}
	
	// 计算总小时数
	duration := end.Sub(start)
	stats.TotalHours = duration.Hours()
	
	// 获取发电量统计
	genStats, err := c.CalculateGeneration(ctx, stationID, PeriodTypeCustom, start, end)
	if err != nil {
		return nil, fmt.Errorf("calculate generation failed: %w", err)
	}
	
	// 计算等效利用小时
	if station.Capacity > 0 {
		stats.EquivalentHours = genStats.TotalGeneration / station.Capacity
		stats.FullLoadHours = stats.EquivalentHours
		
		// 计算容量系数
		if stats.TotalHours > 0 {
			stats.CapacityFactor = stats.EquivalentHours / stats.TotalHours
		}
	}
	
	// 计算可用率
	deviceStats, err := c.CalculateDeviceRunRate(ctx, stationID, start, end)
	if err == nil {
		stats.Availability = deviceStats.RunRate
	}
	
	return stats, nil
}

// CalculateAllStations 计算所有厂站统计
func (c *StationCalculator) CalculateAllStations(ctx context.Context, periodType PeriodType, start, end time.Time) ([]*StationStatistics, error) {
	// 获取所有厂站
	stations, err := c.config.DataProvider.GetAllStations(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all stations failed: %w", err)
	}
	
	results := make([]*StationStatistics, 0, len(stations))
	resultChan := make(chan *StationStatistics, len(stations))
	errChan := make(chan error, len(stations))
	
	// 并行计算
	workerCount := c.config.ParallelWorkers
	if workerCount <= 0 {
		workerCount = 5
	}
	
	sem := make(chan struct{}, workerCount)
	var wg sync.WaitGroup
	
	for _, station := range stations {
		wg.Add(1)
		go func(s StationInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			
			stats, err := c.CalculateStationStatistics(ctx, s.ID, periodType, start, end)
			if err != nil {
				errChan <- fmt.Errorf("station %s: %w", s.ID, err)
				return
			}
			resultChan <- stats
		}(station)
	}
	
	wg.Wait()
	close(resultChan)
	close(errChan)
	
	// 收集结果
	for stats := range resultChan {
		results = append(results, stats)
	}
	
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

// SaveStatistics 保存统计结果
func (c *StationCalculator) SaveStatistics(ctx context.Context, taskID string, stats *StationStatistics) error {
	result := &StatisticsResult{
		Dimension:      "station",
		DimensionValue: stats.StationID,
		Metrics: map[string]float64{
			"daily_generation":     stats.DailyGeneration,
			"monthly_generation":   stats.MonthlyGeneration,
			"yearly_generation":    stats.YearlyGeneration,
			"device_run_rate":      stats.DeviceRunRate,
			"total_alarm_count":    float64(stats.TotalAlarmCount),
			"active_alarm_count":   float64(stats.ActiveAlarmCount),
			"critical_alarm_count": float64(stats.CriticalAlarmCount),
			"efficiency":           stats.Efficiency,
			"performance_ratio":    stats.PerformanceRatio,
			"equivalent_hours":     stats.EquivalentHours,
		},
		Metadata: map[string]interface{}{
			"station_code":     stats.StationCode,
			"station_name":     stats.StationName,
			"station_type":     stats.StationType,
			"capacity":         stats.Capacity,
			"device_count":     stats.DeviceCount,
			"online_device_count": stats.OnlineDeviceCount,
		},
		PeriodStart: stats.PeriodStart,
		PeriodEnd:   stats.PeriodEnd,
		PeriodType:  PeriodTypeDay,
	}
	
	data := result.ToStatisticsData(taskID)
	return c.storage.SaveBatch(ctx, data)
}

// StatisticsCache 统计缓存
type StatisticsCache struct {
	ttl   time.Duration
	data  map[string]cacheItem
	mu    sync.RWMutex
}

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

// NewStatisticsCache 创建统计缓存
func NewStatisticsCache(ttl time.Duration) *StatisticsCache {
	return &StatisticsCache{
		ttl:  ttl,
		data: make(map[string]cacheItem),
	}
}

// Get 获取缓存
func (c *StatisticsCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, ok := c.data[key]
	if !ok {
		return nil, false
	}
	
	if time.Now().After(item.expiresAt) {
		return nil, false
	}
	
	return item.value, true
}

// Set 设置缓存
func (c *StatisticsCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.data[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// contains 字符串包含检查
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Math helper functions
func mathAbs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// mathPow 计算幂
func mathPow(x, y float64) float64 {
	return math.Pow(x, y)
}

// mathSqrt 计算平方根
func mathSqrt(x float64) float64 {
	return math.Sqrt(x)
}
