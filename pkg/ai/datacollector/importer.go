package datacollector

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

type CSVImporter struct {
	validator  *SimpleValidator
	cleaner    *SimpleCleaner
	jobs       map[string]*ImportJob
	jobsMutex  sync.RWMutex
}

func NewCSVImporter() *CSVImporter {
	return &CSVImporter{
		validator: NewSimpleValidator(ValidationConfig{
			CheckMissing:  true,
			CheckOutliers: true,
			CheckRange:    true,
		}),
		cleaner: NewSimpleCleaner(),
		jobs:    make(map[string]*ImportJob),
	}
}

func (ci *CSVImporter) Import(ctx context.Context, req *BatchImportRequest) (*ImportJob, error) {
	jobID := uuid.New().String()
	
	job := &ImportJob{
		ID:         jobID,
		FilePath:   req.FilePath,
		Format:     req.Format,
		Status:     JobStatusPending,
		CreatedAt:  time.Now(),
	}
	
	ci.jobsMutex.Lock()
	ci.jobs[jobID] = job
	ci.jobsMutex.Unlock()
	
	go ci.processImport(ctx, job, req)
	
	return job, nil
}

func (ci *CSVImporter) processImport(ctx context.Context, job *ImportJob, req *BatchImportRequest) {
	ci.updateJobStatus(job.ID, JobStatusRunning, nil)
	startedAt := time.Now()
	job.StartedAt = &startedAt
	
	file, err := os.Open(req.FilePath)
	if err != nil {
		ci.updateJobStatus(job.ID, JobStatusFailed, fmt.Errorf("failed to open file: %w", err))
		return
	}
	defer file.Close()
	
	var points []*DataPoint
	var parseErr error
	
	switch req.Format {
	case FormatCSV:
		points, parseErr = ci.parseCSV(file, req)
	case FormatJSON:
		points, parseErr = ci.parseJSON(file, req)
	default:
		parseErr = fmt.Errorf("unsupported format: %s", req.Format)
	}
	
	if parseErr != nil {
		ci.updateJobStatus(job.ID, JobStatusFailed, parseErr)
		return
	}
	
	job.TotalRecords = len(points)
	
	if req.Options.Deduplicate {
		deduplicated, err := ci.cleaner.Deduplicate(ctx, points)
		if err != nil {
			ci.updateJobStatus(job.ID, JobStatusFailed, fmt.Errorf("deduplication failed: %w", err))
			return
		}
		points = deduplicated
	}
	
	validated, err := ci.validator.Validate(ctx, points)
	if err != nil {
		ci.updateJobStatus(job.ID, JobStatusFailed, fmt.Errorf("validation failed: %w", err))
		return
	}
	
	if req.Options.FillMissing {
		filled, err := ci.cleaner.FillMissing(ctx, validated, req.Options.FillMethod)
		if err != nil {
			ci.updateJobStatus(job.ID, JobStatusFailed, fmt.Errorf("fill missing failed: %w", err))
			return
		}
		validated = filled
	}
	
	successCount := 0
	failedCount := 0
	
	for _, p := range validated {
		if p.Quality == QualityGood || p.Quality == QualitySuspect {
			successCount++
		} else {
			failedCount++
		}
	}
	
	job.SuccessRecords = successCount
	job.FailedRecords = failedCount
	
	completedAt := time.Now()
	job.CompletedAt = &completedAt
	
	ci.updateJobStatus(job.ID, JobStatusCompleted, nil)
}

func (ci *CSVImporter) parseCSV(file *os.File, req *BatchImportRequest) ([]*DataPoint, error) {
	reader := csv.NewReader(file)
	
	if req.Options.Delimiter != "" {
		reader.Comma = rune(req.Options.Delimiter[0])
	}
	
	if req.Options.SkipHeader {
		_, err := reader.Read()
		if err != nil {
			return nil, err
		}
	}
	
	var points []*DataPoint
	
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		
		point, err := ci.recordToDataPoint(record, req)
		if err != nil {
			continue
		}
		
		points = append(points, point)
	}
	
	return points, nil
}

func (ci *CSVImporter) recordToDataPoint(record []string, req *BatchImportRequest) (*DataPoint, error) {
	point := &DataPoint{
		Quality: QualityUnknown,
	}
	
	for field, indexStr := range req.Mapping {
		index, err := strconv.Atoi(indexStr)
		if err != nil || index < 0 || index >= len(record) {
			continue
		}
		
		value := record[index]
		
		switch field {
		case "timestamp":
			t, err := ci.parseTime(value, req.Options.TimeFormat, req.Options.TimeZone)
			if err == nil {
				point.Timestamp = t
			}
		case "station_id":
			point.StationID = value
		case "device_id":
			point.DeviceID = value
		case "metric":
			point.Metric = value
		case "value":
			v, err := strconv.ParseFloat(value, 64)
			if err == nil {
				point.Value = v
			}
		case "unit":
			point.Unit = value
		}
	}
	
	if point.Timestamp.IsZero() || point.Metric == "" || math.IsNaN(point.Value) {
		return nil, fmt.Errorf("missing required fields")
	}
	
	return point, nil
}

func (ci *CSVImporter) parseTime(value, format, timezone string) (time.Time, error) {
	if format == "" {
		format = time.RFC3339
	}
	
	t, err := time.Parse(format, value)
	if err != nil {
		return time.Time{}, err
	}
	
	if timezone != "" {
		loc, err := time.LoadLocation(timezone)
		if err == nil {
			t = t.In(loc)
		}
	}
	
	return t, nil
}

func (ci *CSVImporter) parseJSON(file *os.File, req *BatchImportRequest) ([]*DataPoint, error) {
	var data []map[string]interface{}
	
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&data)
	if err != nil {
		return nil, err
	}
	
	var points []*DataPoint
	
	for _, item := range data {
		point := &DataPoint{
			Quality: QualityUnknown,
		}
		
		if ts, ok := item["timestamp"].(string); ok {
			t, err := ci.parseTime(ts, req.Options.TimeFormat, req.Options.TimeZone)
			if err == nil {
				point.Timestamp = t
			}
		}
		
		if stationID, ok := item["station_id"].(string); ok {
			point.StationID = stationID
		}
		
		if deviceID, ok := item["device_id"].(string); ok {
			point.DeviceID = deviceID
		}
		
		if metric, ok := item["metric"].(string); ok {
			point.Metric = metric
		}
		
		if value, ok := item["value"].(float64); ok {
			point.Value = value
		}
		
		if unit, ok := item["unit"].(string); ok {
			point.Unit = unit
		}
		
		if !point.Timestamp.IsZero() && point.Metric != "" {
			points = append(points, point)
		}
	}
	
	return points, nil
}

func (ci *CSVImporter) GetJobStatus(ctx context.Context, jobID string) (*ImportJob, error) {
	ci.jobsMutex.RLock()
	defer ci.jobsMutex.RUnlock()
	
	job, exists := ci.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	
	return job, nil
}

func (ci *CSVImporter) CancelJob(ctx context.Context, jobID string) error {
	ci.jobsMutex.Lock()
	defer ci.jobsMutex.Unlock()
	
	job, exists := ci.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}
	
	if job.Status == JobStatusRunning {
		job.Status = JobStatusPaused
	}
	
	return nil
}

func (ci *CSVImporter) updateJobStatus(jobID string, status JobStatus, err error) {
	ci.jobsMutex.Lock()
	defer ci.jobsMutex.Unlock()
	
	job, exists := ci.jobs[jobID]
	if !exists {
		return
	}
	
	job.Status = status
	if err != nil {
		job.ErrorMessage = err.Error()
	}
}

type BatchProcessor struct {
	importer    *CSVImporter
	validator   *BatchValidator
	cleaner     *AdvancedCleaner
	batchSize   int
}

func NewBatchProcessor(batchSize int) *BatchProcessor {
	return &BatchProcessor{
		importer:  NewCSVImporter(),
		validator: NewBatchValidator(ValidationConfig{}),
		cleaner:   NewAdvancedCleaner(),
		batchSize: batchSize,
	}
}

func (bp *BatchProcessor) ProcessLargeFile(ctx context.Context, req *BatchImportRequest) (*ImportJob, error) {
	job, err := bp.importer.Import(ctx, req)
	if err != nil {
		return nil, err
	}
	
	return job, nil
}
