package datacollector

import (
	"context"
	"time"
)

type DataQuality int

const (
	QualityUnknown DataQuality = iota
	QualityGood
	QualitySuspect
	QualityBad
	QualityMissing
)

type DataFormat string

const (
	FormatCSV     DataFormat = "csv"
	FormatJSON    DataFormat = "json"
	FormatParquet DataFormat = "parquet"
	FormatExcel   DataFormat = "excel"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusPaused    JobStatus = "paused"
)

type DataPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	StationID string                 `json:"station_id"`
	DeviceID  string                 `json:"device_id"`
	Metric    string                 `json:"metric"`
	Value     float64                `json:"value"`
	Unit      string                 `json:"unit"`
	Quality   DataQuality            `json:"quality"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type BatchImportRequest struct {
	FilePath   string                 `json:"file_path"`
	Format     DataFormat             `json:"format"`
	Mapping    map[string]string      `json:"mapping"`
	Validation ValidationConfig       `json:"validation"`
	Options    ImportOptions          `json:"options"`
}

type ValidationConfig struct {
	CheckMissing     bool              `json:"check_missing"`
	CheckOutliers    bool              `json:"check_outliers"`
	CheckRange       bool              `json:"check_range"`
	CheckContinuity  bool              `json:"check_continuity"`
	RangeConstraints map[string][2]float64 `json:"range_constraints"`
	OutlierMethod    string            `json:"outlier_method"`
}

type ImportOptions struct {
	BatchSize      int           `json:"batch_size"`
	SkipHeader     bool          `json:"skip_header"`
	Delimiter      string        `json:"delimiter"`
	TimeFormat     string        `json:"time_format"`
	TimeZone       string        `json:"time_zone"`
	Deduplicate    bool          `json:"deduplicate"`
	FillMissing    bool          `json:"fill_missing"`
	FillMethod     string        `json:"fill_method"`
}

type ImportJob struct {
	ID              string    `json:"id"`
	FilePath        string    `json:"file_path"`
	Format          DataFormat `json:"format"`
	Status          JobStatus `json:"status"`
	TotalRecords    int       `json:"total_records"`
	SuccessRecords  int       `json:"success_records"`
	FailedRecords   int       `json:"failed_records"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type DataQualityReport struct {
	ID               int64     `json:"id"`
	ReportDate       time.Time `json:"report_date"`
	StationID        string    `json:"station_id,omitempty"`
	DeviceID         string    `json:"device_id,omitempty"`
	Metric           string    `json:"metric,omitempty"`
	TotalPoints      int       `json:"total_points"`
	GoodPoints       int       `json:"good_points"`
	SuspectPoints    int       `json:"suspect_points"`
	BadPoints        int       `json:"bad_points"`
	MissingPoints    int       `json:"missing_points"`
	CompletenessRate float64   `json:"completeness_rate"`
	AccuracyRate     float64   `json:"accuracy_rate"`
	CreatedAt        time.Time `json:"created_at"`
}

type DataCollector interface {
	Collect(ctx context.Context) (<-chan *DataPoint, error)
	Close() error
}

type DataImporter interface {
	Import(ctx context.Context, req *BatchImportRequest) (*ImportJob, error)
	GetJobStatus(ctx context.Context, jobID string) (*ImportJob, error)
	CancelJob(ctx context.Context, jobID string) error
}

type DataValidator interface {
	Validate(ctx context.Context, points []*DataPoint) ([]*DataPoint, error)
	GenerateReport(ctx context.Context, stationID, deviceID, metric string, startTime, endTime time.Time) (*DataQualityReport, error)
}

type DataCleaner interface {
	Deduplicate(ctx context.Context, points []*DataPoint) ([]*DataPoint, error)
	FillMissing(ctx context.Context, points []*DataPoint, method string) ([]*DataPoint, error)
	RemoveOutliers(ctx context.Context, points []*DataPoint, method string) ([]*DataPoint, error)
}

type DataStore interface {
	Save(ctx context.Context, points []*DataPoint) error
	Query(ctx context.Context, stationID, deviceID, metric string, startTime, endTime time.Time) ([]*DataPoint, error)
	Delete(ctx context.Context, stationID, deviceID, metric string, startTime, endTime time.Time) error
}
