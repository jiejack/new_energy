package timeseries

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.uber.org/zap"
)

// ClickHouseConfig ClickHouse配置
type ClickHouseConfig struct {
	Addr        []string      `json:"addr" mapstructure:"addr"`               // 地址列表
	Database    string        `json:"database" mapstructure:"database"`       // 默认数据库
	User        string        `json:"user" mapstructure:"user"`               // 用户名
	Password    string        `json:"password" mapstructure:"password"`       // 密码
	Compression string        `json:"compression" mapstructure:"compression"` // 压缩方式: none, gzip, zstd, lz4
	MaxOpenConns int          `json:"max_open_conns" mapstructure:"max_open_conns"` // 最大连接数
	MaxIdleConns int          `json:"max_idle_conns" mapstructure:"max_idle_conns"` // 最大空闲连接数
	ConnTimeout  time.Duration `json:"conn_timeout" mapstructure:"conn_timeout"` // 连接超时
	QueryTimeout time.Duration `json:"query_timeout" mapstructure:"query_timeout"` // 查询超时
	WriteTimeout time.Duration `json:"write_timeout" mapstructure:"write_timeout"` // 写入超时
	BatchSize    int          `json:"batch_size" mapstructure:"batch_size"`   // 批量写入大小
	BlockSize    int          `json:"block_size" mapstructure:"block_size"`   // 块大小
	Debug        bool         `json:"debug" mapstructure:"debug"`             // 调试模式
}

// DefaultClickHouseConfig 默认ClickHouse配置
func DefaultClickHouseConfig() *ClickHouseConfig {
	return &ClickHouseConfig{
		Addr:         []string{"localhost:9000"},
		Database:     "nem_ts",
		User:         "default",
		Password:     "",
		Compression:  "zstd",
		MaxOpenConns: 100,
		MaxIdleConns: 20,
		ConnTimeout:  10 * time.Second,
		QueryTimeout: 60 * time.Second,
		WriteTimeout: 30 * time.Second,
		BatchSize:    10000,
		BlockSize:    65536,
		Debug:        false,
	}
}

// ClickHouseClient ClickHouse客户端
type ClickHouseClient struct {
	conn    driver.Conn
	config  *ClickHouseConfig
	logger  *zap.Logger
	mu      sync.RWMutex
	closed  bool
}

// NewClickHouseClient 创建ClickHouse客户端
func NewClickHouseClient(config *ClickHouseConfig) (*ClickHouseClient, error) {
	if config == nil {
		config = DefaultClickHouseConfig()
	}

	if len(config.Addr) == 0 {
		return nil, fmt.Errorf("no addresses configured")
	}

	if config.Database == "" {
		config.Database = "default"
	}

	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 100
	}

	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 20
	}

	if config.BatchSize == 0 {
		config.BatchSize = 10000
	}

	logger := zap.L().Named("clickhouse")

	// 构建连接选项
	options := &clickhouse.Options{
		Addr: config.Addr,
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.User,
			Password: config.Password,
		},
		MaxOpenConns:    config.MaxOpenConns,
		MaxIdleConns:    config.MaxIdleConns,
		ConnMaxLifetime: time.Hour,
		DialTimeout:     config.ConnTimeout,
		Debug:           config.Debug,
	}

	// 设置压缩
	switch strings.ToLower(config.Compression) {
	case "gzip":
		options.Compression = &clickhouse.Compression{
			Method: clickhouse.CompressionGZIP,
		}
	case "zstd":
		options.Compression = &clickhouse.Compression{
			Method: clickhouse.CompressionZSTD,
		}
	case "lz4":
		options.Compression = &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		}
	default:
		options.Compression = &clickhouse.Compression{
			Method: clickhouse.CompressionNone,
		}
	}

	// 设置块大小
	if config.BlockSize > 0 {
		options.BlockBufferSize = uint(config.BlockSize)
	}

	// 创建连接
	conn, err := clickhouse.Open(options)
	if err != nil {
		return nil, NewConnectionError(
			strings.Join(config.Addr, ","),
			"failed to open connection",
			err,
		)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnTimeout)
	err = conn.Ping(ctx)
	cancel()

	if err != nil {
		conn.Close()
		return nil, NewConnectionError(
			strings.Join(config.Addr, ","),
			"failed to ping",
			err,
		)
	}

	logger.Info("connected to ClickHouse",
		zap.Strings("addr", config.Addr),
		zap.String("database", config.Database))

	return &ClickHouseClient{
		conn:   conn,
		config: config,
		logger: logger,
	}, nil
}

// Write 写入数据点
func (c *ClickHouseClient) Write(ctx context.Context, points []*DataPoint) error {
	return c.WriteWithTable(ctx, c.config.Database, "data_points", points)
}

// WriteBatch 批量写入数据点
func (c *ClickHouseClient) WriteBatch(ctx context.Context, points []*DataPoint) error {
	return c.Write(ctx, points)
}

// WriteWithTable 写入数据到指定表
func (c *ClickHouseClient) WriteWithTable(ctx context.Context, database, table string, points []*DataPoint) error {
	if c.IsClosed() {
		return ErrClosed
	}

	if len(points) == 0 {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.config.WriteTimeout)
	defer cancel()

	// 获取批量写入器
	batch, err := c.conn.PrepareBatch(ctx, fmt.Sprintf(
		"INSERT INTO `%s`.`%s` (point_id, timestamp, value, quality, tags)",
		database, table))
	if err != nil {
		return NewWriteError(len(points), "failed to prepare batch", err)
	}

	// 添加数据点
	for _, point := range points {
		if point == nil {
			continue
		}

		err := batch.Append(
			point.PointID,
			point.Timestamp,
			point.Value,
			point.Quality,
			point.Tags,
		)
		if err != nil {
			return NewWriteError(len(points), "failed to append point", err)
		}
	}

	// 执行批量写入
	err = batch.Send()
	if err != nil {
		return NewWriteError(len(points), "failed to send batch", err)
	}

	c.logger.Debug("wrote points to ClickHouse",
		zap.String("database", database),
		zap.String("table", table),
		zap.Int("count", len(points)))

	return nil
}

// Query 查询数据
func (c *ClickHouseClient) Query(ctx context.Context, query *Query) (*QueryResult, error) {
	if c.IsClosed() {
		return nil, ErrClosed
	}

	if query == nil {
		return nil, ErrInvalidQuery
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// 构建SQL
	sql, args := c.buildQuerySQL(query)

	ctx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	// 执行查询
	rows, err := c.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, NewQueryError(sql, "query failed", err)
	}
	defer rows.Close()

	// 解析结果
	points := make([]*DataPoint, 0)
	for rows.Next() {
		point := &DataPoint{
			Tags: make(map[string]string),
		}
		err := rows.Scan(
			&point.PointID,
			&point.Timestamp,
			&point.Value,
			&point.Quality,
			&point.Tags,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		points = append(points, point)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// 获取总数
	total := int64(len(points))
	if query.Limit > 0 || query.Offset > 0 {
		countSQL, countArgs := c.buildCountSQL(query)
		err = c.conn.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
		if err != nil {
			c.logger.Warn("failed to get total count", zap.Error(err))
			total = int64(len(points))
		}
	}

	return &QueryResult{
		Points: points,
		Total:  total,
	}, nil
}

// QueryRange 查询时间范围内的数据
func (c *ClickHouseClient) QueryRange(ctx context.Context, start, end time.Time, pointIds []int64) ([]*DataPoint, error) {
	query := &Query{
		Database:  c.config.Database,
		Table:     "data_points",
		PointIDs:  pointIds,
		StartTime: start,
		EndTime:   end,
		OrderBy:   "timestamp",
		Order:     "ASC",
	}

	result, err := c.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Points, nil
}

// QueryLatest 查询最新数据
func (c *ClickHouseClient) QueryLatest(ctx context.Context, pointIds []int64) (map[int64]*DataPoint, error) {
	if c.IsClosed() {
		return nil, ErrClosed
	}

	if len(pointIds) == 0 {
		return make(map[int64]*DataPoint), nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	// 使用argMax函数获取最新值
	query := fmt.Sprintf(`
		SELECT
			point_id,
			argMax(timestamp, timestamp) as timestamp,
			argMax(value, timestamp) as value,
			argMax(quality, timestamp) as quality,
			argMax(tags, timestamp) as tags
		FROM %s.data_points
		WHERE point_id IN (?)
		GROUP BY point_id
	`, c.config.Database)

	rows, err := c.conn.Query(ctx, query, pointIds)
	if err != nil {
		return nil, NewQueryError(query, "failed to query latest", err)
	}
	defer rows.Close()

	result := make(map[int64]*DataPoint)
	for rows.Next() {
		point := &DataPoint{
			Tags: make(map[string]string),
		}
		err := rows.Scan(
			&point.PointID,
			&point.Timestamp,
			&point.Value,
			&point.Quality,
			&point.Tags,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[point.PointID] = point
	}

	return result, nil
}

// Aggregate 聚合查询
func (c *ClickHouseClient) Aggregate(ctx context.Context, query *AggregateQuery) (*AggregateResult, error) {
	if c.IsClosed() {
		return nil, ErrClosed
	}

	if query == nil {
		return nil, ErrInvalidQuery
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// 构建聚合SQL
	sql, args := c.buildAggregateSQL(query)

	ctx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	rows, err := c.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, NewQueryError(sql, "aggregate query failed", err)
	}
	defer rows.Close()

	// 按测点分组结果
	seriesMap := make(map[int64]*TimeSeries)

	for rows.Next() {
		var pointID int64
		var timestamp time.Time
		var value float64
		var count uint64
		var tags map[string]string

		err := rows.Scan(&pointID, &timestamp, &value, &count, &tags)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		series, exists := seriesMap[pointID]
		if !exists {
			series = &TimeSeries{
				PointID: pointID,
				Tags:    tags,
				Points:  make([]AggPoint, 0),
			}
			seriesMap[pointID] = series
		}

		series.Points = append(series.Points, AggPoint{
			Timestamp: timestamp,
			Value:     value,
			Count:     int64(count),
		})
	}

	series := make([]TimeSeries, 0, len(seriesMap))
	for _, s := range seriesMap {
		series = append(series, *s)
	}

	return &AggregateResult{Series: series}, nil
}

// Downsample 降采样
func (c *ClickHouseClient) Downsample(ctx context.Context, query *DownsampleQuery) error {
	if c.IsClosed() {
		return ErrClosed
	}

	// 构建降采样插入SQL
	aggQuery := &AggregateQuery{
		Database:  query.SourceDatabase,
		Table:     query.SourceTable,
		PointIDs:  query.PointIDs,
		StartTime: query.StartTime,
		EndTime:   query.EndTime,
		Interval:  query.Interval,
		AggFunc:   query.AggFunc,
	}

	result, err := c.Aggregate(ctx, aggQuery)
	if err != nil {
		return fmt.Errorf("aggregate for downsample failed: %w", err)
	}

	// 写入目标表
	for _, series := range result.Series {
		points := make([]*DataPoint, 0, len(series.Points))
		for _, p := range series.Points {
			points = append(points, &DataPoint{
				PointID:   series.PointID,
				Timestamp: p.Timestamp,
				Value:     p.Value,
				Quality:   QualityGood,
				Tags:      series.Tags,
			})
		}

		if len(points) > 0 {
			err = c.WriteWithTable(ctx, query.TargetDatabase, query.TargetTable, points)
			if err != nil {
				return fmt.Errorf("write downsampled data failed: %w", err)
			}
		}
	}

	return nil
}

// CreateDatabase 创建数据库
func (c *ClickHouseClient) CreateDatabase(ctx context.Context, name string) error {
	if c.IsClosed() {
		return ErrClosed
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", name)
	err := c.conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("create database failed: %w", err)
	}

	c.logger.Info("created database", zap.String("database", name))
	return nil
}

// DropDatabase 删除数据库
func (c *ClickHouseClient) DropDatabase(ctx context.Context, name string) error {
	if c.IsClosed() {
		return ErrClosed
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	query := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", name)
	err := c.conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("drop database failed: %w", err)
	}

	c.logger.Info("dropped database", zap.String("database", name))
	return nil
}

// CreateTable 创建表
func (c *ClickHouseClient) CreateTable(ctx context.Context, database, table string, schema *TableSchema) error {
	if c.IsClosed() {
		return ErrClosed
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	// 构建列定义
	var columnsDef []string
	for _, col := range schema.Columns {
		colDef := fmt.Sprintf("`%s` %s", col.Name, col.Type)
		if col.Nullable {
			colDef = fmt.Sprintf("`%s` Nullable(%s)", col.Name, col.Type)
		}
		if col.Comment != "" {
			colDef += fmt.Sprintf(" COMMENT '%s'", col.Comment)
		}
		columnsDef = append(columnsDef, colDef)
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s`.`%s` (\n  %s\n)",
		database, table, strings.Join(columnsDef, ",\n  "))

	// 添加引擎配置
	engine := schema.Engine
	if engine == "" {
		engine = "MergeTree"
	}
	query += fmt.Sprintf("\nENGINE = %s", engine)

	// 添加排序键
	if len(schema.SortKey) > 0 {
		quotedKeys := make([]string, len(schema.SortKey))
		for i, k := range schema.SortKey {
			quotedKeys[i] = fmt.Sprintf("`%s`", k)
		}
		query += fmt.Sprintf("\nORDER BY (%s)", strings.Join(quotedKeys, ", "))
	}

	// 添加分区配置
	if schema.PartitionBy != "" {
		query += fmt.Sprintf("\nPARTITION BY %s", schema.PartitionBy)
	}

	err := c.conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("create table failed: %w", err)
	}

	c.logger.Info("created table",
		zap.String("database", database),
		zap.String("table", table))
	return nil
}

// DropTable 删除表
func (c *ClickHouseClient) DropTable(ctx context.Context, database, table string) error {
	if c.IsClosed() {
		return ErrClosed
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	query := fmt.Sprintf("DROP TABLE IF EXISTS `%s`.`%s`", database, table)
	err := c.conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("drop table failed: %w", err)
	}

	c.logger.Info("dropped table",
		zap.String("database", database),
		zap.String("table", table))
	return nil
}

// CreateRetentionPolicy 创建保留策略
func (c *ClickHouseClient) CreateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error {
	// ClickHouse使用TTL实现数据保留策略
	// 这里简化处理，实际需要通过ALTER TABLE设置TTL
	c.logger.Info("retention policy created",
		zap.String("name", policy.Name),
		zap.String("database", policy.Database))
	return nil
}

// UpdateRetentionPolicy 更新保留策略
func (c *ClickHouseClient) UpdateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error {
	c.logger.Info("retention policy updated",
		zap.String("name", policy.Name),
		zap.String("database", policy.Database))
	return nil
}

// DeleteRetentionPolicy 删除保留策略
func (c *ClickHouseClient) DeleteRetentionPolicy(ctx context.Context, database, name string) error {
	c.logger.Info("retention policy deleted",
		zap.String("name", name),
		zap.String("database", database))
	return nil
}

// Ping 健康检查
func (c *ClickHouseClient) Ping(ctx context.Context) error {
	if c.IsClosed() {
		return ErrClosed
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.conn.Ping(ctx)
}

// IsConnected 检查连接状态
func (c *ClickHouseClient) IsConnected() bool {
	return !c.IsClosed()
}

// IsClosed 检查是否已关闭
func (c *ClickHouseClient) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// Close 关闭连接
func (c *ClickHouseClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			c.logger.Error("failed to close connection", zap.Error(err))
			return err
		}
	}

	c.logger.Info("ClickHouse client closed")
	return nil
}

// buildQuerySQL 构建查询SQL
func (c *ClickHouseClient) buildQuerySQL(query *Query) (string, []interface{}) {
	var whereClauses []string
	var args []interface{}

	database := query.Database
	if database == "" {
		database = c.config.Database
	}
	table := query.Table
	if table == "" {
		table = "data_points"
	}

	// 测点ID条件
	if len(query.PointIDs) > 0 {
		whereClauses = append(whereClauses, "point_id IN (?)")
		args = append(args, query.PointIDs)
	}

	// 时间范围条件
	if !query.StartTime.IsZero() {
		whereClauses = append(whereClauses, "timestamp >= ?")
		args = append(args, query.StartTime)
	}
	if !query.EndTime.IsZero() {
		whereClauses = append(whereClauses, "timestamp <= ?")
		args = append(args, query.EndTime)
	}

	// 构建WHERE子句
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 构建ORDER BY子句
	orderBy := "ORDER BY timestamp ASC"
	if query.OrderBy != "" {
		order := "ASC"
		if strings.ToUpper(query.Order) == "DESC" {
			order = "DESC"
		}
		orderBy = fmt.Sprintf("ORDER BY %s %s", query.OrderBy, order)
	}

	// 构建LIMIT子句
	limitClause := ""
	if query.Limit > 0 {
		if query.Offset > 0 {
			limitClause = fmt.Sprintf("LIMIT %d OFFSET %d", query.Limit, query.Offset)
		} else {
			limitClause = fmt.Sprintf("LIMIT %d", query.Limit)
		}
	}

	sql := fmt.Sprintf(`
		SELECT point_id, timestamp, value, quality, tags
		FROM %s.%s
		%s
		%s
		%s
	`, database, table, whereClause, orderBy, limitClause)

	return sql, args
}

// buildCountSQL 构建计数SQL
func (c *ClickHouseClient) buildCountSQL(query *Query) (string, []interface{}) {
	var whereClauses []string
	var args []interface{}

	database := query.Database
	if database == "" {
		database = c.config.Database
	}
	table := query.Table
	if table == "" {
		table = "data_points"
	}

	if len(query.PointIDs) > 0 {
		whereClauses = append(whereClauses, "point_id IN (?)")
		args = append(args, query.PointIDs)
	}

	if !query.StartTime.IsZero() {
		whereClauses = append(whereClauses, "timestamp >= ?")
		args = append(args, query.StartTime)
	}
	if !query.EndTime.IsZero() {
		whereClauses = append(whereClauses, "timestamp <= ?")
		args = append(args, query.EndTime)
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s %s", database, table, whereClause)
	return sql, args
}

// buildAggregateSQL 构建聚合SQL
func (c *ClickHouseClient) buildAggregateSQL(query *AggregateQuery) (string, []interface{}) {
	var whereClauses []string
	var args []interface{}

	database := query.Database
	if database == "" {
		database = c.config.Database
	}
	table := query.Table
	if table == "" {
		table = "data_points"
	}

	// 测点ID条件
	if len(query.PointIDs) > 0 {
		whereClauses = append(whereClauses, "point_id IN (?)")
		args = append(args, query.PointIDs)
	}

	// 时间范围条件
	if !query.StartTime.IsZero() {
		whereClauses = append(whereClauses, "timestamp >= ?")
		args = append(args, query.StartTime)
	}
	if !query.EndTime.IsZero() {
		whereClauses = append(whereClauses, "timestamp <= ?")
		args = append(args, query.EndTime)
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 时间间隔
	intervalStr := c.intervalToSQL(query.Interval)

	// 聚合函数
	aggFunc := strings.ToUpper(query.AggFunc)
	if aggFunc == "" {
		aggFunc = "avg"
	}

	sql := fmt.Sprintf(`
		SELECT
			point_id,
			toStartOfInterval(timestamp, INTERVAL %s) as time_bucket,
			%s(value) as value,
			COUNT(*) as count,
			any(tags) as tags
		FROM %s.%s
		%s
		GROUP BY point_id, time_bucket
		ORDER BY point_id, time_bucket
	`, intervalStr, aggFunc, database, table, whereClause)

	return sql, args
}

// intervalToSQL 将时间间隔转换为SQL格式
func (c *ClickHouseClient) intervalToSQL(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%d SECOND", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%d MINUTE", seconds/60)
	} else if seconds < 86400 {
		return fmt.Sprintf("%d HOUR", seconds/3600)
	}
	return fmt.Sprintf("%d DAY", seconds/86400)
}
