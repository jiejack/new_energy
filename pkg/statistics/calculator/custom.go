package calculator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// DimensionType 维度类型
type DimensionType string

const (
	DimensionTypeStation    DimensionType = "station"
	DimensionTypeDeviceType DimensionType = "device_type"
	DimensionTypeRegion     DimensionType = "region"
	DimensionTypeTime       DimensionType = "time"
	DimensionTypeCustom     DimensionType = "custom"
)

// AggregationType 聚合类型
type AggregationType string

const (
	AggregationSum     AggregationType = "sum"
	AggregationAvg     AggregationType = "avg"
	AggregationMin     AggregationType = "min"
	AggregationMax     AggregationType = "max"
	AggregationCount   AggregationType = "count"
	AggregationFirst   AggregationType = "first"
	AggregationLast    AggregationType = "last"
	AggregationStdDev  AggregationType = "stddev"
	AggregationVariance AggregationType = "variance"
	AggregationMedian  AggregationType = "median"
	AggregationP95     AggregationType = "p95"
	AggregationP99     AggregationType = "p99"
)

// CustomMetric 自定义指标定义
type CustomMetric struct {
	Name         string         `json:"name"`
	DisplayName  string         `json:"display_name"`
	Unit         string         `json:"unit"`
	Description  string         `json:"description"`
	Aggregation  AggregationType `json:"aggregation"`
	Expression   string         `json:"expression"`  // 计算表达式
	PointIDs     []string       `json:"point_ids"`   // 相关采集点
	PointCodes   []string       `json:"point_codes"` // 采集点编码
	ScaleFactor  float64        `json:"scale_factor"`
	Offset       float64        `json:"offset"`
}

// CustomDimension 自定义维度定义
type CustomDimension struct {
	Name        string       `json:"name"`
	Type        DimensionType `json:"type"`
	Expression  string       `json:"expression"`  // 维度值计算表达式
	Values      []string     `json:"values"`      // 预定义值
	Mappings    map[string]string `json:"mappings"` // 值映射
}

// FilterCondition 过滤条件
type FilterCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, gte, lte, in, not_in, like
	Value    interface{} `json:"value"`
}

// FilterGroup 过滤条件组
type FilterGroup struct {
	Conditions []FilterCondition `json:"conditions"`
	Logic      string            `json:"logic"` // and, or
	Groups     []FilterGroup     `json:"groups"`
}

// GroupByField 分组字段
type GroupByField struct {
	Field      string `json:"field"`
	Alias      string `json:"alias"`
	TimeFormat string `json:"time_format"` // 时间字段格式化
}

// OrderByField 排序字段
type OrderByField struct {
	Field string `json:"field"`
	Desc  bool   `json:"desc"`
}

// CustomStatisticsConfig 自定义统计配置
type CustomStatisticsConfig struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	
	// 维度配置
	Dimensions  []CustomDimension  `json:"dimensions"`
	
	// 指标配置
	Metrics     []CustomMetric     `json:"metrics"`
	
	// 过滤条件
	Filters     FilterGroup        `json:"filters"`
	
	// 分组配置
	GroupBy     []GroupByField     `json:"group_by"`
	
	// 排序配置
	OrderBy     []OrderByField     `json:"order_by"`
	
	// 周期配置
	PeriodType  PeriodType         `json:"period_type"`
	PeriodStart time.Time          `json:"period_start"`
	PeriodEnd   time.Time          `json:"period_end"`
	
	// 分页配置
	Limit       int                `json:"limit"`
	Offset      int                `json:"offset"`
	
	// 高级配置
	Options     map[string]interface{} `json:"options"`
}

// CustomStatisticsResult 自定义统计结果
type CustomStatisticsResult struct {
	ConfigID    string                 `json:"config_id"`
	ConfigName  string                 `json:"config_name"`
	PeriodStart time.Time              `json:"period_start"`
	PeriodEnd   time.Time              `json:"period_end"`
	
	// 维度值
	Dimensions  map[string]string      `json:"dimensions"`
	
	// 指标值
	Metrics     map[string]float64     `json:"metrics"`
	
	// 元数据
	Metadata    map[string]interface{} `json:"metadata"`
	
	// 数据质量
	DataPoints  int64                  `json:"data_points"`
	Quality     float64                `json:"quality"`
}

// CustomStatisticsResults 自定义统计结果集
type CustomStatisticsResults struct {
	ConfigID    string                      `json:"config_id"`
	ConfigName  string                      `json:"config_name"`
	PeriodStart time.Time                   `json:"period_start"`
	PeriodEnd   time.Time                   `json:"period_end"`
	Results     []*CustomStatisticsResult   `json:"results"`
	Total       int                         `json:"total"`
	Summary     map[string]*AggregatedStatistics `json:"summary"`
}

// CustomCalculatorConfig 自定义统计器配置
type CustomCalculatorConfig struct {
	// 并行计算配置
	ParallelWorkers int
	BatchSize       int
	
	// 缓存配置
	CacheEnabled  bool
	CacheTTL      time.Duration
	
	// 数据源配置
	DataProvider DataProvider
}

// CustomCalculator 自定义统计器
type CustomCalculator struct {
	config  CustomCalculatorConfig
	storage StatisticsStorage
	cache   *StatisticsCache
	mu      sync.RWMutex
	
	// 已注册的统计配置
	configs map[string]*CustomStatisticsConfig
}

// NewCustomCalculator 创建自定义统计器
func NewCustomCalculator(config CustomCalculatorConfig, storage StatisticsStorage) *CustomCalculator {
	calc := &CustomCalculator{
		config:  config,
		storage: storage,
		configs: make(map[string]*CustomStatisticsConfig),
	}
	
	if config.CacheEnabled {
		calc.cache = NewStatisticsCache(config.CacheTTL)
	}
	
	return calc
}

// RegisterConfig 注册统计配置
func (c *CustomCalculator) RegisterConfig(config *CustomStatisticsConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if config.ID == "" {
		config.ID = generateUUID()
	}
	
	c.configs[config.ID] = config
	return nil
}

// GetConfig 获取统计配置
func (c *CustomCalculator) GetConfig(configID string) (*CustomStatisticsConfig, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	config, ok := c.configs[configID]
	if !ok {
		return nil, fmt.Errorf("config not found: %s", configID)
	}
	
	return config, nil
}

// ListConfigs 列出所有配置
func (c *CustomCalculator) ListConfigs() []*CustomStatisticsConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	configs := make([]*CustomStatisticsConfig, 0, len(c.configs))
	for _, config := range c.configs {
		configs = append(configs, config)
	}
	
	return configs
}

// Calculate 执行自定义统计
func (c *CustomCalculator) Calculate(ctx context.Context, configID string) (*CustomStatisticsResults, error) {
	config, err := c.GetConfig(configID)
	if err != nil {
		return nil, err
	}
	
	return c.CalculateWithConfig(ctx, config)
}

// CalculateWithConfig 使用配置执行统计
func (c *CustomCalculator) CalculateWithConfig(ctx context.Context, config *CustomStatisticsConfig) (*CustomStatisticsResults, error) {
	// 检查缓存
	cacheKey := fmt.Sprintf("custom_stats:%s:%d", config.ID, config.PeriodStart.Unix())
	if c.cache != nil {
		if cached, ok := c.cache.Get(cacheKey); ok {
			if results, ok := cached.(*CustomStatisticsResults); ok {
				return results, nil
			}
		}
	}
	
	results := &CustomStatisticsResults{
		ConfigID:    config.ID,
		ConfigName:  config.Name,
		PeriodStart: config.PeriodStart,
		PeriodEnd:   config.PeriodEnd,
		Results:     make([]*CustomStatisticsResult, 0),
		Summary:     make(map[string]*AggregatedStatistics),
	}
	
	// 获取数据源
	data, err := c.fetchSourceData(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("fetch source data failed: %w", err)
	}
	
	// 应用过滤条件
	filteredData := c.applyFilters(data, config.Filters)
	
	// 分组统计
	groupedResults := c.groupAndAggregate(filteredData, config)
	
	// 应用排序
	sortedResults := c.applySorting(groupedResults, config.OrderBy)
	
	// 应用分页
	pagedResults := c.applyPagination(sortedResults, config.Limit, config.Offset)
	
	results.Results = pagedResults
	results.Total = len(groupedResults)
	
	// 计算汇总统计
	results.Summary = c.calculateSummary(pagedResults)
	
	// 缓存结果
	if c.cache != nil {
		c.cache.Set(cacheKey, results)
	}
	
	return results, nil
}

// fetchSourceData 获取源数据
func (c *CustomCalculator) fetchSourceData(ctx context.Context, config *CustomStatisticsConfig) ([]map[string]interface{}, error) {
	var allData []map[string]interface{}
	
	// 收集所有需要的采集点
	pointIDSet := make(map[string]bool)
	for _, metric := range config.Metrics {
		for _, pointID := range metric.PointIDs {
			pointIDSet[pointID] = true
		}
	}
	
	pointIDs := make([]string, 0, len(pointIDSet))
	for pointID := range pointIDSet {
		pointIDs = append(pointIDs, pointID)
	}
	
	if len(pointIDs) == 0 {
		return allData, nil
	}
	
	// 获取时序数据
	data, err := c.config.DataProvider.GetTimeSeriesData(ctx, pointIDs, config.PeriodStart, config.PeriodEnd)
	if err != nil {
		return nil, err
	}
	
	// 转换为通用格式
	for pointID, points := range data {
		for _, p := range points {
			record := map[string]interface{}{
				"point_id":  pointID,
				"timestamp": p.Timestamp,
				"value":     p.Value,
				"quality":   p.Quality,
			}
			allData = append(allData, record)
		}
	}
	
	return allData, nil
}

// applyFilters 应用过滤条件
func (c *CustomCalculator) applyFilters(data []map[string]interface{}, filterGroup FilterGroup) []map[string]interface{} {
	if len(filterGroup.Conditions) == 0 && len(filterGroup.Groups) == 0 {
		return data
	}
	
	var result []map[string]interface{}
	
	for _, record := range data {
		if c.evaluateFilterGroup(record, filterGroup) {
			result = append(result, record)
		}
	}
	
	return result
}

// evaluateFilterGroup 评估过滤条件组
func (c *CustomCalculator) evaluateFilterGroup(record map[string]interface{}, group FilterGroup) bool {
	if len(group.Conditions) == 0 && len(group.Groups) == 0 {
		return true
	}
	
	results := make([]bool, 0)
	
	// 评估条件
	for _, cond := range group.Conditions {
		results = append(results, c.evaluateCondition(record, cond))
	}
	
	// 评估子组
	for _, subGroup := range group.Groups {
		results = append(results, c.evaluateFilterGroup(record, subGroup))
	}
	
	// 应用逻辑运算
	if group.Logic == "or" {
		for _, r := range results {
			if r {
				return true
			}
		}
		return false
	}
	
	// 默认 AND 逻辑
	for _, r := range results {
		if !r {
			return false
		}
	}
	return true
}

// evaluateCondition 评估单个条件
func (c *CustomCalculator) evaluateCondition(record map[string]interface{}, cond FilterCondition) bool {
	value, ok := record[cond.Field]
	if !ok {
		return false
	}
	
	switch cond.Operator {
	case "eq":
		return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", cond.Value)
	case "ne":
		return fmt.Sprintf("%v", value) != fmt.Sprintf("%v", cond.Value)
	case "gt":
		if v, ok := toFloat64(value); ok {
			if cv, ok := toFloat64(cond.Value); ok {
				return v > cv
			}
		}
	case "lt":
		if v, ok := toFloat64(value); ok {
			if cv, ok := toFloat64(cond.Value); ok {
				return v < cv
			}
		}
	case "gte":
		if v, ok := toFloat64(value); ok {
			if cv, ok := toFloat64(cond.Value); ok {
				return v >= cv
			}
		}
	case "lte":
		if v, ok := toFloat64(value); ok {
			if cv, ok := toFloat64(cond.Value); ok {
				return v <= cv
			}
		}
	case "in":
		if arr, ok := cond.Value.([]interface{}); ok {
			for _, item := range arr {
				if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", item) {
					return true
				}
			}
		}
		return false
	case "not_in":
		if arr, ok := cond.Value.([]interface{}); ok {
			for _, item := range arr {
				if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", item) {
					return false
				}
			}
		}
		return true
	case "like":
		return contains(fmt.Sprintf("%v", value), fmt.Sprintf("%v", cond.Value))
	}
	
	return false
}

// groupAndAggregate 分组聚合
func (c *CustomCalculator) groupAndAggregate(data []map[string]interface{}, config *CustomStatisticsConfig) []*CustomStatisticsResult {
	// 按分组字段分组
	groups := make(map[string][]map[string]interface{})
	
	for _, record := range data {
		key := c.generateGroupKey(record, config.GroupBy)
		groups[key] = append(groups[key], record)
	}
	
	// 对每个分组计算指标
	results := make([]*CustomStatisticsResult, 0, len(groups))
	
	for key, groupData := range groups {
		result := &CustomStatisticsResult{
			ConfigID:    config.ID,
			ConfigName:  config.Name,
			PeriodStart: config.PeriodStart,
			PeriodEnd:   config.PeriodEnd,
			Dimensions:  c.parseGroupKey(key, config.GroupBy),
			Metrics:     make(map[string]float64),
			Metadata:    make(map[string]interface{}),
		}
		
		// 计算每个指标
		for _, metric := range config.Metrics {
			value := c.calculateMetric(groupData, metric)
			result.Metrics[metric.Name] = value
		}
		
		// 计算数据质量
		result.DataPoints = int64(len(groupData))
		if len(groupData) > 0 {
			result.Quality = 100.0
		}
		
		results = append(results, result)
	}
	
	return results
}

// generateGroupKey 生成分组键
func (c *CustomCalculator) generateGroupKey(record map[string]interface{}, groupBy []GroupByField) string {
	key := ""
	for i, field := range groupBy {
		if i > 0 {
			key += "|"
		}
		value := record[field.Field]
		if field.TimeFormat != "" {
			if t, ok := value.(time.Time); ok {
				key += t.Format(field.TimeFormat)
				continue
			}
		}
		key += fmt.Sprintf("%v", value)
	}
	return key
}

// parseGroupKey 解析分组键
func (c *CustomCalculator) parseGroupKey(key string, groupBy []GroupByField) map[string]string {
	dimensions := make(map[string]string)
	values := splitString(key, "|")
	
	for i, field := range groupBy {
		if i < len(values) {
			fieldName := field.Alias
			if fieldName == "" {
				fieldName = field.Field
			}
			dimensions[fieldName] = values[i]
		}
	}
	
	return dimensions
}

// calculateMetric 计算指标值
func (c *CustomCalculator) calculateMetric(data []map[string]interface{}, metric CustomMetric) float64 {
	if len(data) == 0 {
		return 0
	}
	
	// 提取值
	values := make([]float64, 0)
	for _, record := range data {
		if v, ok := record["value"].(float64); ok {
			// 应用缩放因子和偏移
			v = v*metric.ScaleFactor + metric.Offset
			values = append(values, v)
		}
	}
	
	if len(values) == 0 {
		return 0
	}
	
	// 应用聚合函数
	switch metric.Aggregation {
	case AggregationSum:
		return sumValues(values)
	case AggregationAvg:
		return avgValues(values)
	case AggregationMin:
		return minValues(values)
	case AggregationMax:
		return maxValues(values)
	case AggregationCount:
		return float64(len(values))
	case AggregationFirst:
		return values[0]
	case AggregationLast:
		return values[len(values)-1]
	case AggregationStdDev:
		stats := CalculateAggregated(values)
		return stats.StdDev
	case AggregationVariance:
		stats := CalculateAggregated(values)
		return stats.Variance
	case AggregationMedian:
		return medianValues(values)
	case AggregationP95:
		return percentileValues(values, 95)
	case AggregationP99:
		return percentileValues(values, 99)
	default:
		return avgValues(values)
	}
}

// applySorting 应用排序
func (c *CustomCalculator) applySorting(results []*CustomStatisticsResult, orderBy []OrderByField) []*CustomStatisticsResult {
	if len(orderBy) == 0 {
		return results
	}
	
	// 简单排序实现
	sorted := make([]*CustomStatisticsResult, len(results))
	copy(sorted, results)
	
	// 使用冒泡排序（简化实现）
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if c.compareResults(sorted[j], sorted[j+1], orderBy) > 0 {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	return sorted
}

// compareResults 比较两个结果
func (c *CustomCalculator) compareResults(a, b *CustomStatisticsResult, orderBy []OrderByField) int {
	for _, field := range orderBy {
		var aVal, bVal float64
		var ok bool
		
		// 尝试从指标中获取值
		aVal, ok = a.Metrics[field.Field]
		if !ok {
			// 尝试从维度中获取值
			if aDim, ok := a.Dimensions[field.Field]; ok {
				aVal, _ = toFloat64(aDim)
			}
		}
		
		bVal, ok = b.Metrics[field.Field]
		if !ok {
			if bDim, ok := b.Dimensions[field.Field]; ok {
				bVal, _ = toFloat64(bDim)
			}
		}
		
		if aVal < bVal {
			if field.Desc {
				return 1
			}
			return -1
		} else if aVal > bVal {
			if field.Desc {
				return -1
			}
			return 1
		}
	}
	return 0
}

// applyPagination 应用分页
func (c *CustomCalculator) applyPagination(results []*CustomStatisticsResult, limit, offset int) []*CustomStatisticsResult {
	if offset >= len(results) {
		return []*CustomStatisticsResult{}
	}
	
	start := offset
	end := len(results)
	if limit > 0 && start+limit < end {
		end = start + limit
	}
	
	return results[start:end]
}

// calculateSummary 计算汇总统计
func (c *CustomCalculator) calculateSummary(results []*CustomStatisticsResult) map[string]*AggregatedStatistics {
	summary := make(map[string]*AggregatedStatistics)
	
	if len(results) == 0 {
		return summary
	}
	
	// 收集每个指标的所有值
	metricValues := make(map[string][]float64)
	for _, result := range results {
		for name, value := range result.Metrics {
			metricValues[name] = append(metricValues[name], value)
		}
	}
	
	// 计算每个指标的汇总统计
	for name, values := range metricValues {
		summary[name] = CalculateAggregated(values)
	}
	
	return summary
}

// CalculateMultiDimension 多维度聚合统计
func (c *CustomCalculator) CalculateMultiDimension(ctx context.Context, config *CustomStatisticsConfig) (*CustomStatisticsResults, error) {
	return c.CalculateWithConfig(ctx, config)
}

// CalculateTimeSeries 时间序列统计
func (c *CustomCalculator) CalculateTimeSeries(ctx context.Context, config *CustomStatisticsConfig, interval time.Duration) ([]*CustomStatisticsResults, error) {
	var allResults []*CustomStatisticsResults
	
	start := config.PeriodStart
	end := config.PeriodEnd
	
	for t := start; t.Before(end); t = t.Add(interval) {
		periodEnd := t.Add(interval)
		if periodEnd.After(end) {
			periodEnd = end
		}
		
		// 创建临时配置
		tempConfig := *config
		tempConfig.PeriodStart = t
		tempConfig.PeriodEnd = periodEnd
		
		results, err := c.CalculateWithConfig(ctx, &tempConfig)
		if err != nil {
			return nil, fmt.Errorf("calculate at %v failed: %w", t, err)
		}
		
		allResults = append(allResults, results)
	}
	
	return allResults, nil
}

// SaveResults 保存统计结果
func (c *CustomCalculator) SaveResults(ctx context.Context, taskID string, results *CustomStatisticsResults) error {
	var allData []*StatisticsData
	
	for _, result := range results.Results {
		// 构建维度值字符串
		dimensionValue := ""
		for _, v := range result.Dimensions {
			if dimensionValue != "" {
				dimensionValue += "_"
			}
			dimensionValue += v
		}
		
		for metricName, metricValue := range result.Metrics {
			data := &StatisticsData{
				ID:             generateUUID(),
				TaskID:         taskID,
				Dimension:      results.ConfigID,
				DimensionValue: dimensionValue,
				MetricName:     metricName,
				MetricValue:    metricValue,
				PeriodType:     PeriodTypeCustom,
				PeriodStart:    result.PeriodStart,
				PeriodEnd:      result.PeriodEnd,
				CreatedAt:      time.Now(),
			}
			
			// 添加元数据
			if len(result.Metadata) > 0 {
				if bytes, err := json.Marshal(result.Metadata); err == nil {
					data.Metadata = string(bytes)
				}
			}
			
			allData = append(allData, data)
		}
	}
	
	return c.storage.SaveBatch(ctx, allData)
}

// CreatePresetConfig 创建预设统计配置
func (c *CustomCalculator) CreatePresetConfig(presetType string, params map[string]interface{}) (*CustomStatisticsConfig, error) {
	switch presetType {
	case "generation_by_hour":
		return c.createGenerationByHourConfig(params)
	case "device_status_summary":
		return c.createDeviceStatusSummaryConfig(params)
	case "alarm_statistics":
		return c.createAlarmStatisticsConfig(params)
	case "efficiency_analysis":
		return c.createEfficiencyAnalysisConfig(params)
	default:
		return nil, fmt.Errorf("unknown preset type: %s", presetType)
	}
}

// createGenerationByHourConfig 创建按小时发电量统计配置
func (c *CustomCalculator) createGenerationByHourConfig(params map[string]interface{}) (*CustomStatisticsConfig, error) {
	config := &CustomStatisticsConfig{
		ID:          generateUUID(),
		Name:        "按小时发电量统计",
		Description: "统计每小时的发电量数据",
		PeriodType:  PeriodTypeHour,
		Dimensions: []CustomDimension{
			{
				Name: "hour",
				Type: DimensionTypeTime,
			},
		},
		Metrics: []CustomMetric{
			{
				Name:        "generation",
				DisplayName: "发电量",
				Unit:        "kWh",
				Aggregation: AggregationSum,
			},
			{
				Name:        "avg_power",
				DisplayName: "平均功率",
				Unit:        "kW",
				Aggregation: AggregationAvg,
			},
			{
				Name:        "peak_power",
				DisplayName: "峰值功率",
				Unit:        "kW",
				Aggregation: AggregationMax,
			},
		},
		GroupBy: []GroupByField{
			{
				Field:      "timestamp",
				Alias:      "hour",
				TimeFormat: "2006-01-02 15:00",
			},
		},
		OrderBy: []OrderByField{
			{Field: "hour", Desc: false},
		},
	}
	
	// 应用参数
	if stationID, ok := params["station_id"].(string); ok {
		config.Filters = FilterGroup{
			Conditions: []FilterCondition{
				{
					Field:    "station_id",
					Operator: "eq",
					Value:    stationID,
				},
			},
		}
	}
	
	return config, nil
}

// createDeviceStatusSummaryConfig 创建设备状态汇总配置
func (c *CustomCalculator) createDeviceStatusSummaryConfig(params map[string]interface{}) (*CustomStatisticsConfig, error) {
	config := &CustomStatisticsConfig{
		ID:          generateUUID(),
		Name:        "设备状态汇总",
		Description: "按设备类型统计设备状态",
		PeriodType:  PeriodTypeDay,
		Dimensions: []CustomDimension{
			{
				Name: "device_type",
				Type: DimensionTypeDeviceType,
			},
		},
		Metrics: []CustomMetric{
			{
				Name:        "total_count",
				DisplayName: "设备总数",
				Aggregation: AggregationCount,
			},
			{
				Name:        "online_rate",
				DisplayName: "在线率",
				Unit:        "%",
				Aggregation: AggregationAvg,
			},
		},
		GroupBy: []GroupByField{
			{Field: "device_type", Alias: "device_type"},
		},
	}
	
	return config, nil
}

// createAlarmStatisticsConfig 创建告警统计配置
func (c *CustomCalculator) createAlarmStatisticsConfig(params map[string]interface{}) (*CustomStatisticsConfig, error) {
	config := &CustomStatisticsConfig{
		ID:          generateUUID(),
		Name:        "告警统计",
		Description: "按级别和类型统计告警",
		PeriodType:  PeriodTypeDay,
		Dimensions: []CustomDimension{
			{
				Name: "level",
				Type: DimensionTypeCustom,
			},
			{
				Name: "type",
				Type: DimensionTypeCustom,
			},
		},
		Metrics: []CustomMetric{
			{
				Name:        "count",
				DisplayName: "告警数量",
				Aggregation: AggregationCount,
			},
			{
				Name:        "avg_duration",
				DisplayName: "平均持续时间",
				Unit:        "分钟",
				Aggregation: AggregationAvg,
			},
		},
		GroupBy: []GroupByField{
			{Field: "level", Alias: "level"},
			{Field: "type", Alias: "type"},
		},
		OrderBy: []OrderByField{
			{Field: "count", Desc: true},
		},
	}
	
	return config, nil
}

// createEfficiencyAnalysisConfig 创建效率分析配置
func (c *CustomCalculator) createEfficiencyAnalysisConfig(params map[string]interface{}) (*CustomStatisticsConfig, error) {
	config := &CustomStatisticsConfig{
		ID:          generateUUID(),
		Name:        "效率分析",
		Description: "分析设备运行效率",
		PeriodType:  PeriodTypeDay,
		Dimensions: []CustomDimension{
			{
				Name: "station",
				Type: DimensionTypeStation,
			},
		},
		Metrics: []CustomMetric{
			{
				Name:        "avg_efficiency",
				DisplayName: "平均效率",
				Unit:        "%",
				Aggregation: AggregationAvg,
			},
			{
				Name:        "max_efficiency",
				DisplayName: "最高效率",
				Unit:        "%",
				Aggregation: AggregationMax,
			},
			{
				Name:        "min_efficiency",
				DisplayName: "最低效率",
				Unit:        "%",
				Aggregation: AggregationMin,
			},
			{
				Name:        "efficiency_stddev",
				DisplayName: "效率标准差",
				Aggregation: AggregationStdDev,
			},
		},
		GroupBy: []GroupByField{
			{Field: "station_id", Alias: "station"},
		},
		OrderBy: []OrderByField{
			{Field: "avg_efficiency", Desc: true},
		},
	}
	
	return config, nil
}

// 辅助函数

// toFloat64 转换为float64
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	default:
		return 0, false
	}
}

// splitString 分割字符串
func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

// sumValues 求和
func sumValues(values []float64) float64 {
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum
}

// avgValues 平均值
func avgValues(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return sumValues(values) / float64(len(values))
}

// minValues 最小值
func minValues(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// maxValues 最大值
func maxValues(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// medianValues 中位数
func medianValues(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// 简单排序
	sorted := make([]float64, len(values))
	copy(sorted, values)
	
	// 冒泡排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// percentileValues 百分位数
func percentileValues(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// 排序
	sorted := make([]float64, len(values))
	copy(sorted, values)
	
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	index := float64(len(sorted)-1) * percentile / 100
	lower := int(index)
	upper := lower + 1
	
	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}
	
	fraction := index - float64(lower)
	return sorted[lower] + (sorted[upper]-sorted[lower])*fraction
}
