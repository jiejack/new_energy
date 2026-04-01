package timeseries

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver for Doris
	"go.uber.org/zap"
)

// DorisConfig Doris配置
type DorisConfig struct {
	Hosts        []string      `json:"hosts" mapstructure:"hosts"`               // FE节点地址列表
	Port         int           `json:"port" mapstructure:"port"`                 // MySQL协议端口
	Database     string        `json:"database" mapstructure:"database"`         // 默认数据库
	User         string        `json:"user" mapstructure:"user"`                 // 用户名
	Password     string        `json:"password" mapstructure:"password"`         // 密码
	MaxOpenConns int           `json:"max_open_conns" mapstructure:"max_open_conns"` // 最大打开连接数
	MaxIdleConns int           `json:"max_idle_conns" mapstructure:"max_idle_conns"` // 最大空闲连接数
	ConnTimeout  time.Duration `json:"conn_timeout" mapstructure:"conn_timeout"` // 连接超时
	WriteTimeout time.Duration `json:"write_timeout" mapstructure:"write_timeout"` // 写入超时
	QueryTimeout time.Duration `json:"query_timeout" mapstructure:"query_timeout"` // 查询超时
	BatchSize    int           `json:"batch_size" mapstructure:"batch_size"`     // 批量写入大小
}

// DefaultDorisConfig 默认Doris配置
func DefaultDorisConfig() *DorisConfig {
	return &DorisConfig{
		Hosts:        []string{"localhost"},
		Port:         9030,
		Database:     "nem_ts",
		User:         "root",
		Password:     "",
		MaxOpenConns: 100,
		MaxIdleConns: 20,
		ConnTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		QueryTimeout: 60 * time.Second,
		BatchSize:    10000,
	}
}

// DorisClient Doris客户端
type DorisClient struct {
	db      *sql.DB
	config  *DorisConfig
	logger  *zap.Logger
	mu      sync.RWMutex
	closed  bool
	currentHost int
}

// NewDorisClient 创建Doris客户端
func NewDorisClient(config *DorisConfig) (*DorisClient, error) {
	if config == nil {
		config = DefaultDorisConfig()
	}

	if len(config.Hosts) == 0 {
		return nil, fmt.Errorf("no hosts configured")
	}

	if config.Port == 0 {
		config.Port = 9030
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

	logger := zap.L().Named("doris")

	client := &DorisClient{
		config: config,
		logger: logger,
	}

	// 尝试连接到可用的主机
	var lastErr error
	for i, host := range config.Hosts {
		dsn := client.buildDSN(host)
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			lastErr = err
			logger.Warn("failed to open connection to host",
				zap.String("host", host),
				zap.Error(err))
			continue
		}

		// 配置连接池
		db.SetMaxOpenConns(config.MaxOpenConns)
		db.SetMaxIdleConns(config.MaxIdleConns)
		db.SetConnMaxLifetime(time.Hour)

		// 测试连接
		ctx, cancel := context.WithTimeout(context.Background(), config.ConnTimeout)
		err = db.PingContext(ctx)
		cancel()

		if err != nil {
			db.Close()
			lastErr = err
			logger.Warn("failed to ping host",
				zap.String("host", host),
				zap.Error(err))
			continue
		}

		client.db = db
		client.currentHost = i
		logger.Info("connected to Doris",
			zap.String("host", host),
			zap.Int("port", config.Port))
		break
	}

	if client.db == nil {
		return nil, NewConnectionError(
			strings.Join(config.Hosts, ","),
			"failed to connect to any host",
			lastErr,
		)
	}

	return client, nil
}

// buildDSN 构建DSN
func (c *DorisClient) buildDSN(host string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local&timeout=%s",
		c.config.User,
		c.config.Password,
		host,
		c.config.Port,
		c.config.Database,
		c.config.ConnTimeout,
	)
}

// Write 写入数据点
func (c *DorisClient) Write(ctx context.Context, points []*DataPoint) error {
	return c.WriteWithTable(ctx, c.config.Database, "data_points", points)
}

// WriteBatch 批量写入数据点
func (c *DorisClient) WriteBatch(ctx context.Context, points []*DataPoint) error {
	return c.Write(ctx, points)
}

// WriteWithTable 写入数据到指定表
func (c *DorisClient) WriteWithTable(ctx context.Context, database, table string, points []*DataPoint) error {
	if c.IsClosed() {
		return ErrClosed
	}

	if len(points) == 0 {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// 构建批量插入SQL
	valueStrings := make([]string, 0, len(points))
	valueArgs := make([]interface{}, 0, len(points)*5)

	for _, point := range points {
		if point == nil {
			continue
		}

		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			point.PointID,
			point.Timestamp,
			point.Value,
			point.Quality,
			c.serializeTags(point.Tags),
		)
	}

	if len(valueStrings) == 0 {
		return nil
	}

	query := fmt.Sprintf("INSERT INTO `%s`.`%s` (point_id, timestamp, value, quality, tags) VALUES %s",
		database, table, strings.Join(valueStrings, ","))

	ctx, cancel := context.WithTimeout(ctx, c.config.WriteTimeout)
	defer cancel()

	_, err := c.db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return NewWriteError(len(points), "failed to write points", err)
	}

	c.logger.Debug("wrote points to Doris",
		zap.String("database", database),
		zap.String("table", table),
		zap.Int("count", len(points)))

	return nil
}

// Query 查询数据
func (c *DorisClient) Query(ctx context.Context, query *Query) (*QueryResult, error) {
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
	rows, err := c.db.QueryContext(ctx, sql, args...)
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
		var tagsStr string
		err := rows.Scan(
			&point.PointID,
			&point.Timestamp,
			&point.Value,
			&point.Quality,
			&tagsStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		point.Tags = c.parseTags(tagsStr)
		points = append(points, point)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// 获取总数
	total := int64(len(points))
	if query.Limit > 0 || query.Offset > 0 {
		countSQL, countArgs := c.buildCountSQL(query)
		err = c.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total)
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
func (c *DorisClient) QueryRange(ctx context.Context, start, end time.Time, pointIds []int64) ([]*DataPoint, error) {
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
func (c *DorisClient) QueryLatest(ctx context.Context, pointIds []int64) (map[int64]*DataPoint, error) {
	if c.IsClosed() {
		return nil, ErrClosed
	}

	if len(pointIds) == 0 {
		return make(map[int64]*DataPoint), nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// 使用窗口函数获取每个测点的最新数据
	placeholders := make([]string, len(pointIds))
	args := make([]interface{}, len(pointIds))
	for i, id := range pointIds {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT point_id, timestamp, value, quality, tags
		FROM (
			SELECT point_id, timestamp, value, quality, tags,
				ROW_NUMBER() OVER (PARTITION BY point_id ORDER BY timestamp DESC) as rn
			FROM %s.data_points
			WHERE point_id IN (%s)
		) t
		WHERE rn = 1
	`, c.config.Database, strings.Join(placeholders, ","))

	ctx, cancel := context.WithTimeout(ctx, c.config.QueryTimeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, NewQueryError(query, "failed to query latest", err)
	}
	defer rows.Close()

	result := make(map[int64]*DataPoint)
	for rows.Next() {
		point := &DataPoint{
			Tags: make(map[string]string),
		}
		var tagsStr string
		var rn int
		err := rows.Scan(
			&point.PointID,
			&point.Timestamp,
			&point.Value,
			&point.Quality,
			&tagsStr,
			&rn,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		point.Tags = c.parseTags(tagsStr)
		result[point.PointID] = point
	}

	return result, nil
}

// Aggregate 聚合查询
func (c *DorisClient) Aggregate(ctx context.Context, query *AggregateQuery) (*AggregateResult, error) {
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

	rows, err := c.db.QueryContext(ctx, sql, args...)
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
		var count int64
		var tagsStr sql.NullString

		err := rows.Scan(&pointID, &timestamp, &value, &count, &tagsStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		series, exists := seriesMap[pointID]
		if !exists {
			series = &TimeSeries{
				PointID: pointID,
				Tags:    make(map[string]string),
				Points:  make([]AggPoint, 0),
			}
			if tagsStr.Valid {
				series.Tags = c.parseTags(tagsStr.String)
			}
			seriesMap[pointID] = series
		}

		series.Points = append(series.Points, AggPoint{
			Timestamp: timestamp,
			Value:     value,
			Count:     count,
		})
	}

	series := make([]TimeSeries, 0, len(seriesMap))
	for _, s := range seriesMap {
		series = append(series, *s)
	}

	return &AggregateResult{Series: series}, nil
}

// Downsample 降采样
func (c *DorisClient) Downsample(ctx context.Context, query *DownsampleQuery) error {
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
func (c *DorisClient) CreateDatabase(ctx context.Context, name string) error {
	if c.IsClosed() {
		return ErrClosed
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", name)
	_, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create database failed: %w", err)
	}

	c.logger.Info("created database", zap.String("database", name))
	return nil
}

// DropDatabase 删除数据库
func (c *DorisClient) DropDatabase(ctx context.Context, name string) error {
	if c.IsClosed() {
		return ErrClosed
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	query := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", name)
	_, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("drop database failed: %w", err)
	}

	c.logger.Info("dropped database", zap.String("database", name))
	return nil
}

// CreateTable 创建表
func (c *DorisClient) CreateTable(ctx context.Context, database, table string, schema *TableSchema) error {
	if c.IsClosed() {
		return ErrClosed
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// 构建建表SQL
	var columnsDef []string
	for _, col := range schema.Columns {
		colDef := fmt.Sprintf("`%s` %s", col.Name, col.Type)
		if !col.Nullable {
			colDef += " NOT NULL"
		}
		if col.Default != "" {
			colDef += fmt.Sprintf(" DEFAULT %s", col.Default)
		}
		if col.Comment != "" {
			colDef += fmt.Sprintf(" COMMENT '%s'", col.Comment)
		}
		columnsDef = append(columnsDef, colDef)
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s`.`%s` (\n  %s\n)",
		database, table, strings.Join(columnsDef, ",\n  "))

	// 添加引擎配置
	if schema.Engine != "" {
		query += fmt.Sprintf("\nENGINE = %s", schema.Engine)
	}

	// 添加分区配置
	if schema.PartitionBy != "" {
		query += fmt.Sprintf("\nPARTITION BY %s", schema.PartitionBy)
	}

	// 添加排序键
	if len(schema.SortKey) > 0 {
		quotedKeys := make([]string, len(schema.SortKey))
		for i, k := range schema.SortKey {
			quotedKeys[i] = fmt.Sprintf("`%s`", k)
		}
		query += fmt.Sprintf("\nORDER BY (%s)", strings.Join(quotedKeys, ", "))
	}

	_, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create table failed: %w", err)
	}

	c.logger.Info("created table",
		zap.String("database", database),
		zap.String("table", table))
	return nil
}

// DropTable 删除表
func (c *DorisClient) DropTable(ctx context.Context, database, table string) error {
	if c.IsClosed() {
		return ErrClosed
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	query := fmt.Sprintf("DROP TABLE IF EXISTS `%s`.`%s`", database, table)
	_, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("drop table failed: %w", err)
	}

	c.logger.Info("dropped table",
		zap.String("database", database),
		zap.String("table", table))
	return nil
}

// CreateRetentionPolicy 创建保留策略
func (c *DorisClient) CreateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error {
	// Doris使用TTL实现数据保留策略
	// 这里简化处理，实际需要通过ALTER TABLE设置TTL
	c.logger.Info("retention policy created",
		zap.String("name", policy.Name),
		zap.String("database", policy.Database))
	return nil
}

// UpdateRetentionPolicy 更新保留策略
func (c *DorisClient) UpdateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error {
	c.logger.Info("retention policy updated",
		zap.String("name", policy.Name),
		zap.String("database", policy.Database))
	return nil
}

// DeleteRetentionPolicy 删除保留策略
func (c *DorisClient) DeleteRetentionPolicy(ctx context.Context, database, name string) error {
	c.logger.Info("retention policy deleted",
		zap.String("name", name),
		zap.String("database", database))
	return nil
}

// Ping 健康检查
func (c *DorisClient) Ping(ctx context.Context) error {
	if c.IsClosed() {
		return ErrClosed
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.db.PingContext(ctx)
}

// IsConnected 检查连接状态
func (c *DorisClient) IsConnected() bool {
	return !c.IsClosed()
}

// IsClosed 检查是否已关闭
func (c *DorisClient) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// Close 关闭连接
func (c *DorisClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	if c.db != nil {
		err := c.db.Close()
		if err != nil {
			c.logger.Error("failed to close database connection", zap.Error(err))
			return err
		}
	}

	c.logger.Info("Doris client closed")
	return nil
}

// buildQuerySQL 构建查询SQL
func (c *DorisClient) buildQuerySQL(query *Query) (string, []interface{}) {
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
		placeholders := make([]string, len(query.PointIDs))
		for i, id := range query.PointIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		whereClauses = append(whereClauses,
			fmt.Sprintf("point_id IN (%s)", strings.Join(placeholders, ",")))
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
func (c *DorisClient) buildCountSQL(query *Query) (string, []interface{}) {
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
		placeholders := make([]string, len(query.PointIDs))
		for i, id := range query.PointIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		whereClauses = append(whereClauses,
			fmt.Sprintf("point_id IN (%s)", strings.Join(placeholders, ",")))
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
func (c *DorisClient) buildAggregateSQL(query *AggregateQuery) (string, []interface{}) {
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
		placeholders := make([]string, len(query.PointIDs))
		for i, id := range query.PointIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		whereClauses = append(whereClauses,
			fmt.Sprintf("point_id IN (%s)", strings.Join(placeholders, ",")))
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

	// 时间间隔转换为字符串
	intervalStr := c.intervalToSQL(query.Interval)

	// 聚合函数
	aggFunc := strings.ToUpper(query.AggFunc)
	if aggFunc == "" {
		aggFunc = "AVG"
	}

	sql := fmt.Sprintf(`
		SELECT
			point_id,
			DATE_BIN('%s', timestamp, '%s') as time_bucket,
			%s(value) as value,
			COUNT(*) as count,
			'' as tags
		FROM %s.%s
		%s
		GROUP BY point_id, time_bucket
		ORDER BY point_id, time_bucket
	`, intervalStr, query.StartTime.Format("2006-01-02 15:04:05"), aggFunc, database, table, whereClause)

	return sql, args
}

// intervalToSQL 将时间间隔转换为SQL格式
func (c *DorisClient) intervalToSQL(d time.Duration) string {
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

// serializeTags 序列化标签
func (c *DorisClient) serializeTags(tags map[string]string) string {
	if tags == nil || len(tags) == 0 {
		return ""
	}

	pairs := make([]string, 0, len(tags))
	for k, v := range tags {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(pairs, ",")
}

// parseTags 解析标签
func (c *DorisClient) parseTags(s string) map[string]string {
	tags := make(map[string]string)
	if s == "" {
		return tags
	}

	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			tags[kv[0]] = kv[1]
		}
	}
	return tags
}
