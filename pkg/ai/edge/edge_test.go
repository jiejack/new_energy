package edge

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModelServer(t *testing.T) {
	ms := NewModelServer()
	require.NotNil(t, ms)
}

func TestModelServer_LoadModel(t *testing.T) {
	ms := NewModelServer()
	ctx := context.Background()

	err := ms.LoadModel(ctx, "anomaly-detect", "v1.0")
	require.NoError(t, err)

	models := ms.GetLoadedModels(ctx)
	assert.Equal(t, 1, len(models))
	assert.Contains(t, models, "anomaly-detect")
}

func TestModelServer_UnloadModel(t *testing.T) {
	ms := NewModelServer()
	ctx := context.Background()

	ms.LoadModel(ctx, "anomaly-detect", "v1.0")
	err := ms.UnloadModel(ctx, "anomaly-detect")
	require.NoError(t, err)

	models := ms.GetLoadedModels(ctx)
	assert.Equal(t, 0, len(models))
}

func TestModelServer_Infer(t *testing.T) {
	ms := NewModelServer()
	ctx := context.Background()

	ms.LoadModel(ctx, "anomaly-detect", "v1.0")

	result, err := ms.Infer(ctx, InferenceRequest{
		ModelName: "anomaly-detect",
		Input:     map[string]interface{}{"temperature": 85.5},
		DeviceID:  "dev1",
	})
	require.NoError(t, err)
	assert.Equal(t, "anomaly-detect", result.ModelName)
	assert.Equal(t, "dev1", result.DeviceID)
	assert.True(t, result.Confidence > 0)
	assert.NotNil(t, result.Output)
}

func TestModelServer_Infer_NotLoaded(t *testing.T) {
	ms := NewModelServer()
	ctx := context.Background()

	_, err := ms.Infer(ctx, InferenceRequest{
		ModelName: "nonexistent",
		Input:     map[string]interface{}{},
	})
	assert.Error(t, err)
}

func TestModelServer_IsHealthy(t *testing.T) {
	ms := NewModelServer()
	ctx := context.Background()
	assert.True(t, ms.IsHealthy(ctx))
}

func TestNewHeartbeatMonitor(t *testing.T) {
	hm := NewHeartbeatMonitor()
	require.NotNil(t, hm)
}

func TestHeartbeatMonitor_Receive(t *testing.T) {
	hm := NewHeartbeatMonitor()
	ctx := context.Background()

	err := hm.Receive(ctx, HeartbeatInfo{
		NodeID:      "node1",
		CPUUsage:    45.5,
		MemoryUsage: 60.0,
		Status:      "healthy",
	})
	require.NoError(t, err)

	info, err := hm.GetNodeStatus(ctx, "node1")
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "node1", info.NodeID)
	assert.Equal(t, 45.5, info.CPUUsage)
	assert.Equal(t, "healthy", info.Status)
}

func TestHeartbeatMonitor_GetNodeStatus_NotFound(t *testing.T) {
	hm := NewHeartbeatMonitor()
	ctx := context.Background()

	info, err := hm.GetNodeStatus(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, info)
}

func TestHeartbeatMonitor_CheckTimeout(t *testing.T) {
	hm := NewHeartbeatMonitor()
	ctx := context.Background()

	hm.Receive(ctx, HeartbeatInfo{NodeID: "node1", Status: "healthy"})
	hm.Receive(ctx, HeartbeatInfo{NodeID: "node2", Status: "healthy"})

	timedOut := hm.CheckTimeout(ctx, 1*time.Hour)
	assert.Equal(t, 0, len(timedOut))
}

func TestNewSyncManager(t *testing.T) {
	ms := NewModelServer()
	sm := NewSyncManager(ms)
	require.NotNil(t, sm)
}

func TestSyncManager_SyncToEdge(t *testing.T) {
	ms := NewModelServer()
	sm := NewSyncManager(ms)
	ctx := context.Background()

	ms.LoadModel(ctx, "model1", "v1")
	record, err := sm.SyncToEdge(ctx, "node1")
	require.NoError(t, err)
	assert.Equal(t, "node1", record.NodeID)
	assert.Equal(t, "cloud_to_edge", record.Direction)
	assert.Equal(t, "completed", record.Status)
	assert.Equal(t, 1, record.Items)
}

func TestSyncManager_SyncFromEdge(t *testing.T) {
	ms := NewModelServer()
	sm := NewSyncManager(ms)
	ctx := context.Background()

	record, err := sm.SyncFromEdge(ctx, "node1")
	require.NoError(t, err)
	assert.Equal(t, "edge_to_cloud", record.Direction)
	assert.Equal(t, "completed", record.Status)
}

func TestSyncManager_GetSyncStatus(t *testing.T) {
	ms := NewModelServer()
	sm := NewSyncManager(ms)
	ctx := context.Background()

	_, err := sm.GetSyncStatus(ctx, "node1")
	assert.Error(t, err)
}

func TestNewDataProcessor(t *testing.T) {
	dp := NewDataProcessor()
	require.NotNil(t, dp)
}

func TestDataProcessor_Process(t *testing.T) {
	dp := NewDataProcessor()
	ctx := context.Background()

	result, err := dp.Process(ctx, map[string]interface{}{
		"temperature": 85.5,
		"humidity":    60.0,
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 85.5, result["temperature"])
	assert.NotNil(t, result["processed_at"])
}

func TestDataProcessor_Process_NilData(t *testing.T) {
	dp := NewDataProcessor()
	ctx := context.Background()

	_, err := dp.Process(ctx, nil)
	assert.Error(t, err)
}

func TestDataProcessor_Validate(t *testing.T) {
	dp := NewDataProcessor()
	ctx := context.Background()

	assert.True(t, dp.Validate(ctx, map[string]interface{}{"key": "value"}))
	assert.False(t, dp.Validate(ctx, nil))
}

func TestDataProcessor_Transform(t *testing.T) {
	dp := NewDataProcessor()
	ctx := context.Background()

	result, err := dp.Transform(ctx, map[string]interface{}{"key": "value"})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "value", result["key"])
	assert.NotNil(t, result["processed_at"])
}

func TestInferenceRequest_Struct(t *testing.T) {
	req := InferenceRequest{
		ModelName: "test",
		Input:     map[string]interface{}{"x": 1},
		DeviceID:  "dev1",
	}
	assert.Equal(t, "test", req.ModelName)
}

func TestInferenceResult_Struct(t *testing.T) {
	result := InferenceResult{
		ModelName:   "test",
		Output:      map[string]interface{}{"y": 1},
		InferenceMs: 50,
		DeviceID:    "dev1",
		Confidence:  0.95,
	}
	assert.Equal(t, "test", result.ModelName)
	assert.Equal(t, int64(50), result.InferenceMs)
}

func TestHeartbeatInfo_Struct(t *testing.T) {
	info := HeartbeatInfo{
		NodeID:      "node1",
		CPUUsage:    45.5,
		MemoryUsage: 60.0,
		Status:      "healthy",
	}
	assert.Equal(t, "node1", info.NodeID)
}

func TestSyncRecord_Struct(t *testing.T) {
	record := SyncRecord{
		NodeID:    "node1",
		Direction: "cloud_to_edge",
		Status:    "completed",
		Items:     5,
	}
	assert.Equal(t, "node1", record.NodeID)
	assert.Equal(t, 5, record.Items)
}
