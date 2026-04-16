package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/types"
)

const (
	defaultBatchSize    = 1000
	defaultFlushInterval = 5 * time.Second
)

type ClickHouseStorage struct {
	config         types.StorageConfig
	batchBuffer    []*types.DataPoint
	batchSize      int
	flushInterval  time.Duration
	mu             sync.Mutex
	stopChan       chan struct{}
	started        bool
}

type ClickHouseTableSchema struct {
	Name        string
	Columns     []string
	Engine      string
	OrderBy     []string
	PartitionBy string
	TTL         string
}

func NewClickHouseStorage() *ClickHouseStorage {
	return &ClickHouseStorage{
		batchSize:     defaultBatchSize,
		flushInterval: defaultFlushInterval,
		batchBuffer:   make([]*types.DataPoint, 0, defaultBatchSize),
		stopChan:      make(chan struct{}),
	}
}

func (s *ClickHouseStorage) Init(config types.StorageConfig) error {
	if config.Type != "clickhouse" {
		return &types.Error{
			Code:    types.ErrCodeInvalidConfig,
			Message: fmt.Sprintf("invalid storage type: %s, expected clickhouse", config.Type),
		}
	}

	fmt.Printf("Initializing ClickHouse storage with config: %+v\n", config)

	if config.BatchSize > 0 {
		s.batchSize = config.BatchSize
	}
	if config.FlushInterval > 0 {
		s.flushInterval = time.Duration(config.FlushInterval) * time.Second
	}

	schema := s.getDefaultTableSchema(config.Table)
	fmt.Printf("Creating table if not exists: %s with schema: %+v\n", config.Table, schema)

	s.config = config
	s.started = true

	go s.flushLoop()

	return nil
}

func (s *ClickHouseStorage) getDefaultTableSchema(tableName string) ClickHouseTableSchema {
	return ClickHouseTableSchema{
		Name:    tableName,
		Columns: []string{
			"timestamp DateTime",
			"device_id String",
			"station_id String",
			"metric_name String",
			"metric_value Float64",
			"quality Int32",
			"tags Map(String, String)",
		},
		Engine:      "MergeTree()",
		OrderBy:     []string{"(station_id, device_id, metric_name, timestamp)"},
		PartitionBy: "toYYYYMM(timestamp)",
		TTL:         "timestamp + INTERVAL 1 YEAR",
	}
}

func (s *ClickHouseStorage) Write(data *types.BatchData) error {
	if len(data.DataPoints) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.batchBuffer = append(s.batchBuffer, data.DataPoints...)

	if len(s.batchBuffer) >= s.batchSize {
		return s.flushLocked()
	}

	return nil
}

func (s *ClickHouseStorage) WritePoint(point *types.DataPoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.batchBuffer = append(s.batchBuffer, point)

	if len(s.batchBuffer) >= s.batchSize {
		return s.flushLocked()
	}

	return nil
}

func (s *ClickHouseStorage) flushLoop() {
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			if len(s.batchBuffer) > 0 {
				_ = s.flushLocked()
			}
			s.mu.Unlock()
		case <-s.stopChan:
			s.mu.Lock()
			if len(s.batchBuffer) > 0 {
				_ = s.flushLocked()
			}
			s.mu.Unlock()
			return
		}
	}
}

func (s *ClickHouseStorage) flushLocked() error {
	if len(s.batchBuffer) == 0 {
		return nil
	}

	fmt.Printf("Flushing %d data points to ClickHouse table %s\n", len(s.batchBuffer), s.config.Table)

	s.batchBuffer = s.batchBuffer[:0]
	return nil
}

func (s *ClickHouseStorage) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flushLocked()
}

func (s *ClickHouseStorage) Read(query string) ([]*types.DataPoint, error) {
	fmt.Printf("Reading data from ClickHouse with query: %s\n", query)

	return []*types.DataPoint{}, nil
}

func (s *ClickHouseStorage) ReadTimeRange(
	startTime, endTime time.Time,
	stationID, deviceID, metricName string,
) ([]*types.DataPoint, error) {
	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE timestamp >= '%s' AND timestamp <= '%s'",
		s.config.Table,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
	)

	if stationID != "" {
		query += fmt.Sprintf(" AND station_id = '%s'", stationID)
	}
	if deviceID != "" {
		query += fmt.Sprintf(" AND device_id = '%s'", deviceID)
	}
	if metricName != "" {
		query += fmt.Sprintf(" AND metric_name = '%s'", metricName)
	}

	query += " ORDER BY timestamp"

	return s.Read(query)
}

func (s *ClickHouseStorage) Query(query string) (interface{}, error) {
	fmt.Printf("Executing query on ClickHouse: %s\n", query)

	result := []map[string]interface{}{
		{
			"query":      query,
			"executed_at": time.Now().Format(time.RFC3339),
			"rows":       0,
		},
	}

	return result, nil
}

func (s *ClickHouseStorage) Aggregate(
	aggregation string,
	metricName string,
	startTime, endTime time.Time,
	groupBy string,
) (interface{}, error) {
	query := fmt.Sprintf(
		"SELECT %s(metric_value) as value, %s FROM %s WHERE metric_name = '%s' AND timestamp >= '%s' AND timestamp <= '%s' GROUP BY %s",
		aggregation,
		groupBy,
		s.config.Table,
		metricName,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
		groupBy,
	)

	return s.Query(query)
}

func (s *ClickHouseStorage) GetStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"storage_type":   "clickhouse",
		"table":          s.config.Table,
		"batch_size":     s.batchSize,
		"flush_interval": s.flushInterval.String(),
		"buffer_size":    len(s.batchBuffer),
		"started":        s.started,
	}
	return stats, nil
}

func (s *ClickHouseStorage) Close() error {
	if !s.started {
		return nil
	}

	close(s.stopChan)

	s.mu.Lock()
	if len(s.batchBuffer) > 0 {
		_ = s.flushLocked()
	}
	s.mu.Unlock()

	fmt.Println("Closing ClickHouse connection")
	s.started = false

	return nil
}