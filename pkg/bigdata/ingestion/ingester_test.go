package ingestion

import (
	"testing"
	"time"

	"github.com/new-energy-monitoring/pkg/bigdata/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBasicIngester(t *testing.T) {
	i := NewBasicIngester()
	assert.NotNil(t, i)
	assert.False(t, i.running)
}

func TestBasicIngester_Init_Valid(t *testing.T) {
	i := NewBasicIngester()
	err := i.Init(types.IngestionConfig{Type: "basic"})
	require.NoError(t, err)
}

func TestBasicIngester_Init_Invalid(t *testing.T) {
	i := NewBasicIngester()
	err := i.Init(types.IngestionConfig{Type: "invalid"})
	assert.Error(t, err)
}

func TestBasicIngester_StartStop(t *testing.T) {
	i := NewBasicIngester()
	i.Init(types.IngestionConfig{Type: "basic"})

	err := i.Start()
	require.NoError(t, err)
	assert.True(t, i.running)

	err = i.Stop()
	require.NoError(t, err)
	assert.False(t, i.running)
}

func TestBasicIngester_Start_AlreadyRunning(t *testing.T) {
	i := NewBasicIngester()
	i.Init(types.IngestionConfig{Type: "basic"})
	i.Start()
	defer i.Stop()

	err := i.Start()
	assert.Error(t, err)
}

func TestBasicIngester_Stop_NotRunning(t *testing.T) {
	i := NewBasicIngester()
	err := i.Stop()
	assert.Error(t, err)
}

func TestBasicIngester_Ingest(t *testing.T) {
	i := NewBasicIngester()
	i.Init(types.IngestionConfig{Type: "basic"})
	i.Start()
	defer i.Stop()

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	err := i.Ingest(bd)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
}

func TestBasicIngester_Ingest_NotRunning(t *testing.T) {
	i := NewBasicIngester()
	bd := &types.BatchData{DataPoints: []*types.DataPoint{}}
	err := i.Ingest(bd)
	assert.Error(t, err)
}

func TestBasicIngester_RegisterHandler(t *testing.T) {
	i := NewBasicIngester()
	called := false
	i.RegisterHandler(func(data *types.BatchData) {
		called = true
	})
	assert.Len(t, i.handlers, 1)

	i.Init(types.IngestionConfig{Type: "basic"})
	i.Start()

	bd := &types.BatchData{
		DataPoints: []*types.DataPoint{
			{DeviceID: "dev1", Metric: "temp", Value: 25.5, Timestamp: time.Now()},
		},
		Metadata: types.Metadata{Source: "test", BatchID: "b1", Timestamp: time.Now(), RecordCount: 1},
	}
	i.Ingest(bd)
	time.Sleep(200 * time.Millisecond)
	assert.True(t, called)

	i.Stop()
}

func TestBasicIngester_Close_WhileRunning(t *testing.T) {
	i := NewBasicIngester()
	i.Init(types.IngestionConfig{Type: "basic"})
	i.Start()
	err := i.Close()
	assert.NoError(t, err)
}

func TestBasicIngester_Close_NotRunning(t *testing.T) {
	i := NewBasicIngester()
	err := i.Close()
	assert.NoError(t, err)
}

func TestBasicIngester_GetStats(t *testing.T) {
	i := NewBasicIngester()
	i.Init(types.IngestionConfig{Type: "basic"})

	stats := i.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, false, stats["running"])
	assert.Equal(t, 0, stats["handlers_count"])
}

func TestBasicIngester_GetStats_WhileRunning(t *testing.T) {
	i := NewBasicIngester()
	i.Init(types.IngestionConfig{Type: "basic"})
	i.Start()
	defer i.Stop()

	stats := i.GetStats()
	assert.Equal(t, true, stats["running"])
}
