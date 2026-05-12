package edge

import (
	"context"
	"fmt"
	"time"
)

type InferenceRequest struct {
	ModelName string                 `json:"model_name"`
	Input     map[string]interface{} `json:"input"`
	DeviceID  string                 `json:"device_id"`
}

type InferenceResult struct {
	ModelName   string                 `json:"model_name"`
	Output      map[string]interface{} `json:"output"`
	InferenceMs int64                  `json:"inference_ms"`
	DeviceID    string                 `json:"device_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Confidence  float64                `json:"confidence"`
}

type ModelServer interface {
	LoadModel(ctx context.Context, modelName, version string) error
	UnloadModel(ctx context.Context, modelName string) error
	Infer(ctx context.Context, req InferenceRequest) (*InferenceResult, error)
	GetLoadedModels(ctx context.Context) []string
	IsHealthy(ctx context.Context) bool
}

type modelServer struct {
	loadedModels map[string]string
}

func NewModelServer() ModelServer {
	return &modelServer{
		loadedModels: make(map[string]string),
	}
}

func (s *modelServer) LoadModel(ctx context.Context, modelName, version string) error {
	s.loadedModels[modelName] = version
	return nil
}

func (s *modelServer) UnloadModel(ctx context.Context, modelName string) error {
	delete(s.loadedModels, modelName)
	return nil
}

func (s *modelServer) Infer(ctx context.Context, req InferenceRequest) (*InferenceResult, error) {
	if _, ok := s.loadedModels[req.ModelName]; !ok {
		return nil, fmt.Errorf("model %s not loaded", req.ModelName)
	}
	start := time.Now()
	output := make(map[string]interface{})
	for k, v := range req.Input {
		output[k] = v
	}
	output["predicted"] = 0.0
	elapsed := time.Since(start).Milliseconds()
	return &InferenceResult{
		ModelName:   req.ModelName,
		Output:      output,
		InferenceMs: elapsed,
		DeviceID:    req.DeviceID,
		Timestamp:   time.Now(),
		Confidence:  0.85,
	}, nil
}

func (s *modelServer) GetLoadedModels(ctx context.Context) []string {
	models := make([]string, 0, len(s.loadedModels))
	for name := range s.loadedModels {
		models = append(models, name)
	}
	return models
}

func (s *modelServer) IsHealthy(ctx context.Context) bool {
	return true
}
