package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/new-energy-monitoring/pkg/bigdata/types"
)

// ClickHouseStorage 实现了types.Storage接口
type ClickHouseStorage struct {
	client driver.Conn
	config types.StorageConfig
}

// NewClickHouseStorage 创建一个新的ClickHouse存储实例
func NewClickHouseStorage() *ClickHouseStorage {
	return &ClickHouseStorage{}
}

// Init 初始化ClickHouse连接
func (s *ClickHouseStorage) Init(config types.StorageConfig) error {
	if config.Type != "clickhouse" {
		return &types.Error{
			Code:    types.ErrCodeInvalidConfig,
			Message: fmt.Sprintf("invalid storage type: %s, expected clickhouse", config.Type),
		}
	}

	opts := []clickhouse.Option{
		clickhouse.WithAddr(fmt.Sprintf("%s:%d", config.Host, config.Port)),
		clickhouse.WithDatabase(config.Database),
		clickhouse.WithAuth(config.Username, config.Password),
		clickhouse.WithCompression(&clickhouse.Compression{Method: clickhouse.CompressionLZ4}),
		clickhouse.WithDialTimeout(10 * time.Second),
	}

	// 添加额外选项
	if config.Options != nil {
		if secure, ok := config.Options["secure"].(bool); ok {
			opts = append(opts, clickhouse.WithSecure(secure))
		}
	}

	client, err := clickhouse.Open(opts)
	if err != nil {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: fmt.Sprintf("failed to connect to clickhouse: %v", err),
		}
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: fmt.Sprintf("failed to ping clickhouse: %v", err),
		}
	}

	// 创建表（如果不存在）
	if err := s.createTableIfNotExists(client, config.Table); err != nil {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: fmt.Sprintf("failed to create table: %v", err),
		}
	}

	s.client = client
	s.config = config
	return nil
}

// createTableIfNotExists 创建表如果不存在
func (s *ClickHouseStorage) createTableIfNotExists(client driver.Conn, table string) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			timestamp DateTime,
			device_id String,
			metric String,
			value Float64,
			tags Map(String, String),
			attributes Map(String, String),
			created_at DateTime DEFAULT now()
		) ENGINE = MergeTree()
		PARTITION BY toDate(timestamp)
		ORDER BY (device_id, metric, timestamp)
		TTL timestamp + INTERVAL 30 DAY
	`, table)

	ctx := context.Background()
	return client.Exec(ctx, query)
}

// Write 写入批量数据
func (s *ClickHouseStorage) Write(data *types.BatchData) error {
	if s.client == nil {
		return &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "clickhouse client not initialized",
		}
	}

	if len(data.DataPoints) == 0 {
		return nil
	}

	ctx := context.Background()
	batch, err := s.client.PrepareBatch(ctx, fmt.Sprintf(
		"INSERT INTO %s (timestamp, device_id, metric, value, tags, attributes)",
		s.config.Table,
	))
	if err != nil {
		return &types.Error{
			Code:    bigdata.ErrCodeStorageError,
			Message: fmt.Sprintf("failed to prepare batch: %v", err),
		}
	}

	for _, point := range data.DataPoints {
		// 转换attributes为Map(String, String)
		attributes := make(map[string]string)
		if point.Attributes != nil {
			for k, v := range point.Attributes {
				attributes[k] = fmt.Sprintf("%v", v)
			}
		}

		if err := batch.Append(
			point.Timestamp,
			point.DeviceID,
			point.Metric,
			point.Value,
			point.Tags,
			attributes,
		); err != nil {
			return &types.Error{
				Code:    bigdata.ErrCodeStorageError,
				Message: fmt.Sprintf("failed to append data: %v", err),
			}
		}
	}

	if err := batch.Send(); err != nil {
		return &types.Error{
			Code:    bigdata.ErrCodeStorageError,
			Message: fmt.Sprintf("failed to send batch: %v", err),
		}
	}

	return nil
}

// Read 读取数据
func (s *ClickHouseStorage) Read(query string) ([]*types.DataPoint, error) {
	if s.client == nil {
		return nil, &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: "clickhouse client not initialized",
		}
	}

	ctx := context.Background()
	rows, err := s.client.Query(ctx, query)
	if err != nil {
		return nil, &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: fmt.Sprintf("failed to execute query: %v", err),
		}
	}
	defer rows.Close()

	var dataPoints []*types.DataPoint
	for rows.Next() {
		var point types.DataPoint
		var timestamp time.Time
		var deviceID, metric string
		var value float64
		var tags map[string]string
		var attributes map[string]string

		if err := rows.Scan(&timestamp, &deviceID, &metric, &value, &tags, &attributes); err != nil {
		return nil, &types.Error{
			Code:    types.ErrCodeStorageError,
			Message: fmt.Sprintf("failed to scan row: %v", err),
		}
	}

		// 转换attributes为map[string]interface{}
		attrs := make(map[string]interface{})
		for k, v := range attributes {
			attrs[k] = v
		}

		point = bigdata.DataPoint{
			Timestamp:  timestamp,
			DeviceID:   deviceID,
			Metric:     metric,
			Value:      value,
			Tags:       tags,
			Attributes: attrs,
		}

		dataPoints = append(dataPoints, &point)
	}

	if err := rows.Err(); err != nil {
		return nil, &types.Error{
			Code:    bigdata.ErrCodeStorageError,
			Message: fmt.Sprintf("error during rows iteration: %v", err),
		}
	}

	return dataPoints, nil
}

// Query 执行查询
func (s *ClickHouseStorage) Query(query string) (interface{}, error) {
	if s.client == nil {
		return nil, &types.Error{
			Code:    bigdata.ErrCodeStorageError,
			Message: "clickhouse client not initialized",
		}
	}

	ctx := context.Background()
	rows, err := s.client.Query(ctx, query)
	if err != nil {
		return nil, &types.Error{
			Code:    bigdata.ErrCodeStorageError,
			Message: fmt.Sprintf("failed to execute query: %v", err),
		}
	}
	defer rows.Close()

	// 获取列信息
	columns := rows.ColumnTypes()
	columnNames := make([]string, len(columns))
	for i, col := range columns {
		columnNames[i] = col.Name()
	}

	// 存储结果
	var results []map[string]interface{}

	for rows.Next() {
		// 创建一个切片来存储行数据
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, &types.Error{
				Code:    bigdata.ErrCodeStorageError,
				Message: fmt.Sprintf("failed to scan row: %v", err),
			}
		}

		// 构建结果映射
		row := make(map[string]interface{})
		for i, colName := range columnNames {
			row[colName] = values[i]
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, &types.Error{
			Code:    bigdata.ErrCodeStorageError,
			Message: fmt.Sprintf("error during rows iteration: %v", err),
		}
	}

	return results, nil
}

// Close 关闭连接
func (s *ClickHouseStorage) Close() error {
	if s.client != nil {
		if err := s.client.Close(); err != nil {
			return &types.Error{
				Code:    bigdata.ErrCodeStorageError,
				Message: fmt.Sprintf("failed to close clickhouse connection: %v", err),
			}
		}
	}
	return nil
}
