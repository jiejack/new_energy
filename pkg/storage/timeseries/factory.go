package timeseries

import (
	"fmt"
	"time"

	"github.com/new-energy-monitoring/internal/infrastructure/config"
)

// NewTimeSeriesDBFromConfig 根据配置创建时序数据库客户端
func NewTimeSeriesDBFromConfig(cfg *config.TimeSeriesConfig) (TimeSeriesDB, error) {
	if cfg == nil {
		return nil, ErrInvalidConfig
	}

	switch cfg.Type {
	case "doris", "":
		dorisConfig := &DorisConfig{
			Hosts:        cfg.Doris.Hosts,
			Database:     cfg.Doris.Database,
			User:         cfg.Doris.User,
			Password:     cfg.Doris.Password,
			MaxOpenConns: cfg.Doris.MaxOpenConns,
			MaxIdleConns: cfg.Doris.MaxIdleConns,
			ConnTimeout:  cfg.Doris.ConnTimeout,
			WriteTimeout: cfg.Doris.WriteTimeout,
			QueryTimeout: cfg.Doris.QueryTimeout,
			BatchSize:    cfg.Doris.BatchSize,
		}

		// 设置默认值
		if len(dorisConfig.Hosts) == 0 {
			dorisConfig.Hosts = []string{"localhost:9030"}
		}
		if dorisConfig.Database == "" {
			dorisConfig.Database = "nem_ts"
		}
		if dorisConfig.MaxOpenConns == 0 {
			dorisConfig.MaxOpenConns = 100
		}
		if dorisConfig.MaxIdleConns == 0 {
			dorisConfig.MaxIdleConns = 20
		}
		if dorisConfig.ConnTimeout == 0 {
			dorisConfig.ConnTimeout = 10 * time.Second
		}
		if dorisConfig.WriteTimeout == 0 {
			dorisConfig.WriteTimeout = 30 * time.Second
		}
		if dorisConfig.QueryTimeout == 0 {
			dorisConfig.QueryTimeout = 60 * time.Second
		}
		if dorisConfig.BatchSize == 0 {
			dorisConfig.BatchSize = 10000
		}

		return NewDorisClient(dorisConfig)

	case "clickhouse":
		chConfig := &ClickHouseConfig{
			Addr:         cfg.ClickHouse.Addr,
			Database:     cfg.ClickHouse.Database,
			User:         cfg.ClickHouse.User,
			Password:     cfg.ClickHouse.Password,
			Compression:  cfg.ClickHouse.Compression,
			MaxOpenConns: cfg.ClickHouse.MaxOpenConns,
			MaxIdleConns: cfg.ClickHouse.MaxIdleConns,
			ConnTimeout:  cfg.ClickHouse.ConnTimeout,
			WriteTimeout: cfg.ClickHouse.WriteTimeout,
			QueryTimeout: cfg.ClickHouse.QueryTimeout,
			BatchSize:    cfg.ClickHouse.BatchSize,
			BlockSize:    cfg.ClickHouse.BlockSize,
			Debug:        cfg.ClickHouse.Debug,
		}

		// 设置默认值
		if len(chConfig.Addr) == 0 {
			chConfig.Addr = []string{"localhost:9000"}
		}
		if chConfig.Database == "" {
			chConfig.Database = "nem_ts"
		}
		if chConfig.Compression == "" {
			chConfig.Compression = "zstd"
		}
		if chConfig.MaxOpenConns == 0 {
			chConfig.MaxOpenConns = 100
		}
		if chConfig.MaxIdleConns == 0 {
			chConfig.MaxIdleConns = 20
		}
		if chConfig.ConnTimeout == 0 {
			chConfig.ConnTimeout = 10 * time.Second
		}
		if chConfig.WriteTimeout == 0 {
			chConfig.WriteTimeout = 30 * time.Second
		}
		if chConfig.QueryTimeout == 0 {
			chConfig.QueryTimeout = 60 * time.Second
		}
		if chConfig.BatchSize == 0 {
			chConfig.BatchSize = 10000
		}
		if chConfig.BlockSize == 0 {
			chConfig.BlockSize = 65536
		}

		return NewClickHouseClient(chConfig)

	default:
		return nil, fmt.Errorf("unsupported timeseries database type: %s", cfg.Type)
	}
}

// MustNewTimeSeriesDBFromConfig 根据配置创建时序数据库客户端（panic on error）
func MustNewTimeSeriesDBFromConfig(cfg *config.TimeSeriesConfig) TimeSeriesDB {
	client, err := NewTimeSeriesDBFromConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create timeseries client: %v", err))
	}
	return client
}
