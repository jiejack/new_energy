package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type InferenceServiceImpl struct {
	cache        CacheStrategy
	modelManager ModelManager
	jobs         map[string]*BatchJobStatus
	jobsLock     sync.RWMutex
}

func NewInferenceService(cache CacheStrategy, modelManager ModelManager) *InferenceServiceImpl {
	return &InferenceServiceImpl{
		cache:        cache,
		modelManager: modelManager,
		jobs:         make(map[string]*BatchJobStatus),
	}
}

func (s *InferenceServiceImpl) Predict(ctx context.Context, req *PredictRequest) (*PredictResponse, error) {
	startTime := time.Now()
	requestID := uuid.New().String()

	var cacheKey string
	if twoLevelCache, ok := s.cache.(*TwoLevelCache); ok {
		cacheKey = twoLevelCache.GenerateCacheKey(req.ModelID, req.Version, req.Inputs)
	} else {
		cacheKey = fmt.Sprintf("ai:predict:%s:%s:%s", req.ModelID, req.Version, hashInputs(req.Inputs))
	}

	if req.Options.CacheTTLSeconds > 0 {
		cachedResp, err := s.cache.Get(ctx, cacheKey)
		if err == nil {
			cachedResp.RequestID = requestID
			cachedResp.Metadata.InferenceTimeMs = time.Since(startTime).Milliseconds()
			cachedResp.Metadata.Cached = true
			cachedResp.Metadata.Timestamp = time.Now()
			return cachedResp, nil
		}
	}

	prediction, err := s.modelManager.Predict(ctx, req.ModelID, req.Inputs)
	if err != nil {
		return nil, err
	}

	var explanation *Explanation
	if req.Options.IncludeExplanation {
		explanation = &Explanation{
			FeatureImportance: generateFeatureImportance(req.Inputs),
		}
	}

	resp := &PredictResponse{
		RequestID:  requestID,
		ModelID:    req.ModelID,
		Version:    req.Version,
		Prediction: *prediction,
		Explanation: explanation,
		Metadata: InferenceMetadata{
			InferenceTimeMs: time.Since(startTime).Milliseconds(),
			Cached:          false,
			Timestamp:       time.Now(),
		},
	}

	if req.Options.CacheTTLSeconds > 0 {
		ttl := time.Duration(req.Options.CacheTTLSeconds) * time.Second
		s.cache.Set(ctx, cacheKey, resp, ttl)
	}

	return resp, nil
}

func (s *InferenceServiceImpl) BatchPredict(ctx context.Context, req *BatchPredictRequest) (*BatchPredictResponse, error) {
	jobID := uuid.New().String()

	jobStatus := &BatchJobStatus{
		JobID:        jobID,
		Status:       JobStatusQueued,
		Progress:     0,
		ItemCount:    len(req.Inputs),
		ErrorMessage: "",
	}

	s.jobsLock.Lock()
	s.jobs[jobID] = jobStatus
	s.jobsLock.Unlock()

	go s.processBatchJob(ctx, jobID, req)

	return &BatchPredictResponse{
		JobID:               jobID,
		Status:              JobStatusQueued,
		EstimatedTimeSeconds: len(req.Inputs) * 2,
		QueuePosition:        len(s.jobs),
	}, nil
}

func (s *InferenceServiceImpl) processBatchJob(ctx context.Context, jobID string, req *BatchPredictRequest) {
	s.jobsLock.Lock()
	job := s.jobs[jobID]
	job.Status = JobStatusRunning
	now := time.Now()
	job.StartedAt = &now
	s.jobsLock.Unlock()

	defer func() {
		s.jobsLock.Lock()
		job := s.jobs[jobID]
		job.Status = JobStatusCompleted
		completedAt := time.Now()
		job.CompletedAt = &completedAt
		job.Progress = 100
		job.ResultURL = fmt.Sprintf("/api/results/%s", jobID)
		s.jobsLock.Unlock()
	}()

	for i := range req.Inputs {
		select {
		case <-ctx.Done():
			s.jobsLock.Lock()
			job := s.jobs[jobID]
			job.Status = JobStatusCancelled
			s.jobsLock.Unlock()
			return
		default:
			time.Sleep(10 * time.Millisecond)
		}

		progress := (i + 1) * 100 / len(req.Inputs)
		s.jobsLock.Lock()
		job := s.jobs[jobID]
		job.Progress = progress
		s.jobsLock.Unlock()
	}
}

func (s *InferenceServiceImpl) GetBatchJobStatus(ctx context.Context, jobID string) (*BatchJobStatus, error) {
	s.jobsLock.RLock()
	defer s.jobsLock.RUnlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

func (s *InferenceServiceImpl) ListModels(ctx context.Context) ([]*ModelInfo, error) {
	return s.modelManager.ListModels(ctx)
}

func (s *InferenceServiceImpl) GetModel(ctx context.Context, modelID string) (*ModelInfo, error) {
	return s.modelManager.GetModelInfo(ctx, modelID)
}

func hashInputs(inputs map[string]interface{}) string {
	data, _ := json.Marshal(sortMapSimple(inputs))
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func sortMapSimple(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		result[k] = m[k]
	}
	return result
}

func generateFeatureImportance(inputs map[string]interface{}) map[string]float64 {
	importance := make(map[string]float64)
	total := 0.0
	for key := range inputs {
		val := 0.1 + float64(len(key))*0.05
		importance[key] = val
		total += val
	}
	for key := range importance {
		importance[key] /= total
	}
	return importance
}
