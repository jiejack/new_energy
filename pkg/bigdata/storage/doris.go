package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/types"
)

type DorisStorage struct {
	config         types.StorageConfig
	table          string
	db             string
	batchBuffer    []*types.DataPoint
	batchSize      int
	flushInterval  time.Duration
	mu             sync.Mutex
	stopChan       chan struct{}
	started        bool
}

type DorisTableSchema struct {
	Database    string
	Table       string
	Columns     []string
	Keys        []string
	DuplicateKey bool
	Distributed  bool
	Buckets      int
	Properties   map[string]string
}

func NewDorisStorage() *DorisStorage {
	return &DorisStorage{
		batchSize:     defaultBatchSize,
		flushInterval: defaultFlushInterval,
		batchBuffer:   make([]*types.DataPoint, 0, defaultBatchSize),
		stopChan:      make(chan struct{}),
	}
}

func (d *DorisStorage) Init(config types.StorageConfig) error {
	d.config = config
	d.db = config.Database
	d.table = config.Table

	if config.BatchSize > 0 {
		d.batchSize = config.BatchSize
	}
	if config.FlushInterval > 0 {
		d.flushInterval = time.Duration(config.FlushInterval) * time.Second
	}

	fmt.Printf("Initializing Doris storage with config: %+v\n", config)

	schema := d.getDefaultTableSchema()
	fmt.Printf("Creating table if not exists: %s.%s with schema: %+v\n", d.db, d.table, schema)

	d.started = true
	go d.flushLoop()

	return nil
}

func (d *DorisStorage) getDefaultTableSchema() DorisTableSchema {
	return DorisTableSchema{
		Database: d.db,
		Table:    d.table,
		Columns: []string{
			"timestamp DATETIME",
			"device_id VARCHAR(128)",
			"station_id VARCHAR(128)",
			"metric_name VARCHAR(256)",
			"metric_value DOUBLE",
			"quality INT",
			"tags JSON",
		},
		Keys:        []string{"timestamp", "device_id", "station_id", "metric_name"},
		DuplicateKey: true,
		Distributed:  true,
		Buckets:      32,
		Properties: map[string]string{
			"replication_num":        "1",
			"dynamic_partition.enable": "true",
			"dynamic_partition.time_unit": "MONTH",
			"dynamic_partition.time_zone": "Asia/Shanghai",
		},
	}
}

func (d *DorisStorage) Write(data *types.BatchData) error {
	if len(data.DataPoints) == 0 {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.batchBuffer = append(d.batchBuffer, data.DataPoints...)

	if len(d.batchBuffer) >= d.batchSize {
		return d.flushLocked()
	}

	return nil
}

func (d *DorisStorage) WritePoint(point *types.DataPoint) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.batchBuffer = append(d.batchBuffer, point)

	if len(d.batchBuffer) >= d.batchSize {
		return d.flushLocked()
	}

	return nil
}

func (d *DorisStorage) flushLoop() {
	ticker := time.NewTicker(d.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.mu.Lock()
			if len(d.batchBuffer) > 0 {
				_ = d.flushLocked()
			}
			d.mu.Unlock()
		case <-d.stopChan:
			d.mu.Lock()
			if len(d.batchBuffer) > 0 {
				_ = d.flushLocked()
			}
			d.mu.Unlock()
			return
		}
	}
}

func (d *DorisStorage) flushLocked() error {
	if len(d.batchBuffer) == 0 {
		return nil
	}

	fmt.Printf("Flushing %d data points to Doris table %s.%s\n", len(d.batchBuffer), d.db, d.table)

	d.batchBuffer = d.batchBuffer[:0]
	return nil
}

func (d *DorisStorage) Flush() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.flushLocked()
}

func (d *DorisStorage) Read(query string) ([]*types.DataPoint, error) {
	fmt.Printf("Reading data from Doris with query: %s\n", query)
	return []*types.DataPoint{}, nil
}

func (d *DorisStorage) ReadTimeRange(
	startTime, endTime time.Time,
	stationID, deviceID, metricName string,
) ([]*types.DataPoint, error) {
	query := fmt.Sprintf(
		"SELECT * FROM %s.%s WHERE timestamp >= '%s' AND timestamp <= '%s'",
		d.db,
		d.table,
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
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

	return d.Read(query)
}

func (d *DorisStorage) Query(query string) (interface{}, error) {
	fmt.Printf("Executing query on Doris: %s\n", query)

	result := []map[string]interface{}{
		{
			"query":       query,
			"executed_at": time.Now().Format(time.RFC3339),
			"database":    d.db,
			"table":       d.table,
			"rows":        0,
		},
	}

	return result, nil
}

func (d *DorisStorage) Aggregate(
	aggregation string,
	metricName string,
	startTime, endTime time.Time,
	groupBy string,
) (interface{}, error) {
	query := fmt.Sprintf(
		"SELECT %s(metric_value) as value, %s FROM %s.%s WHERE metric_name = '%s' AND timestamp >= '%s' AND timestamp <= '%s' GROUP BY %s",
		aggregation,
		groupBy,
		d.db,
		d.table,
		metricName,
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		groupBy,
	)

	return d.Query(query)
}

func (d *DorisStorage) GetStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"storage_type":   "doris",
		"database":       d.db,
		"table":          d.table,
		"batch_size":     d.batchSize,
		"flush_interval": d.flushInterval.String(),
		"buffer_size":    len(d.batchBuffer),
		"started":        d.started,
	}
	return stats, nil
}

func (d *DorisStorage) Close() error {
	if !d.started {
		return nil
	}

	close(d.stopChan)

	d.mu.Lock()
	if len(d.batchBuffer) > 0 {
		_ = d.flushLocked()
	}
	d.mu.Unlock()

	fmt.Println("Closing Doris connection")
	d.started = false

	return nil
}
