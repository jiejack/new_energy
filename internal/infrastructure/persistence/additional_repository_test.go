package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelVersionRepository_RealDB_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewModelVersionRepository(db)
	ctx := context.Background()

	accuracy := 0.95
	model := &entity.ModelVersion{
		ID:           uuid.New().String(),
		ModelName:    "solar_v2",
		Version:      "1.0.0",
		ArtifactPath: "/path/to/model",
		Accuracy:     &accuracy,
		Status:       entity.ModelStatusStaging,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := repo.Create(ctx, model)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, model.ID)
	assert.NoError(t, err)
	assert.Equal(t, "solar_v2", found.ModelName)
	assert.Equal(t, entity.ModelStatusStaging, found.Status)
}

func TestModelVersionRepository_RealDB_GetProductionModel(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewModelVersionRepository(db)
	ctx := context.Background()

	accuracy := 0.95
	model := &entity.ModelVersion{
		ID:           uuid.New().String(),
		ModelName:    "solar_prod",
		Version:      "1.0.0",
		ArtifactPath: "/path/to/model",
		Accuracy:     &accuracy,
		Status:       entity.ModelStatusProd,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := repo.Create(ctx, model)
	require.NoError(t, err)

	found, err := repo.GetProductionModel(ctx, "solar_prod")
	assert.NoError(t, err)
	assert.Equal(t, model.ID, found.ID)
	assert.Equal(t, entity.ModelStatusProd, found.Status)
}

func TestModelVersionRepository_RealDB_ListByModel(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewModelVersionRepository(db)
	ctx := context.Background()

	accuracy := 0.95
	m1 := &entity.ModelVersion{ID: uuid.New().String(), ModelName: "solar_list", Version: "1.0.0", Accuracy: &accuracy, Status: entity.ModelStatusStaging, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	m2 := &entity.ModelVersion{ID: uuid.New().String(), ModelName: "solar_list", Version: "2.0.0", Accuracy: &accuracy, Status: entity.ModelStatusProd, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	err := repo.Create(ctx, m1)
	require.NoError(t, err)
	err = repo.Create(ctx, m2)
	require.NoError(t, err)

	models, err := repo.ListByModel(ctx, "solar_list")
	assert.NoError(t, err)
	assert.Len(t, models, 2)
}

func TestModelVersionRepository_RealDB_UpdateStatus(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewModelVersionRepository(db)
	ctx := context.Background()

	accuracy := 0.95
	model := &entity.ModelVersion{ID: uuid.New().String(), ModelName: "solar_up", Version: "1.0.0", Accuracy: &accuracy, Status: entity.ModelStatusStaging, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	err := repo.Create(ctx, model)
	require.NoError(t, err)

	err = repo.UpdateStatus(ctx, model.ID, entity.ModelStatusProd)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, model.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity.ModelStatusProd, found.Status)
}

func TestForecastResultRepository_RealDB_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewForecastResultRepository(db)
	ctx := context.Background()

	result := &entity.ForecastResult{
		ID:             uuid.New().String(),
		StationID:      "station-001",
		ForecastType:   entity.ForecastTypeShortTerm,
		TargetTime:     time.Now().Add(1 * time.Hour),
		PredictedPower: 500.0,
		CreatedAt:      time.Now(),
	}
	err := repo.Create(ctx, result)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, result.ID)
	assert.NoError(t, err)
	assert.Equal(t, "station-001", found.StationID)
	assert.Equal(t, 500.0, found.PredictedPower)
}

func TestForecastResultRepository_RealDB_ListByStation(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewForecastResultRepository(db)
	ctx := context.Background()

	r1 := &entity.ForecastResult{ID: uuid.New().String(), StationID: "station-f", ForecastType: entity.ForecastTypeShortTerm, TargetTime: time.Now().Add(1 * time.Hour), PredictedPower: 500.0, CreatedAt: time.Now()}
	r2 := &entity.ForecastResult{ID: uuid.New().String(), StationID: "station-f", ForecastType: entity.ForecastTypeUltraShortTerm, TargetTime: time.Now().Add(2 * time.Hour), PredictedPower: 300.0, CreatedAt: time.Now()}
	err := repo.Create(ctx, r1)
	require.NoError(t, err)
	err = repo.Create(ctx, r2)
	require.NoError(t, err)

	results, total, err := repo.ListByStation(ctx, "station-f", nil, nil, nil, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)

	ft := entity.ForecastTypeShortTerm
	results, total, err = repo.ListByStation(ctx, "station-f", &ft, nil, nil, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
}

func TestForecastResultRepository_RealDB_UpdateActualPower(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewForecastResultRepository(db)
	ctx := context.Background()

	result := &entity.ForecastResult{ID: uuid.New().String(), StationID: "station-u", ForecastType: entity.ForecastTypeShortTerm, TargetTime: time.Now().Add(1 * time.Hour), PredictedPower: 500.0, CreatedAt: time.Now()}
	err := repo.Create(ctx, result)
	require.NoError(t, err)

	err = repo.UpdateActualPower(ctx, result.ID, 480.0)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, result.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found.ActualPower)
	assert.Equal(t, 480.0, *found.ActualPower)
}

func TestForecastResultRepository_RealDB_GetAccuracyStats(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewForecastResultRepository(db)
	ctx := context.Background()

	actualPower1 := 480.0
	actualPower2 := 290.0
	accuracy1 := 96.0
	accuracy2 := 96.7

	r1 := &entity.ForecastResult{ID: uuid.New().String(), StationID: "station-a", ForecastType: entity.ForecastTypeShortTerm, TargetTime: time.Now().Add(-1 * time.Hour), PredictedPower: 500.0, ActualPower: &actualPower1, Accuracy: &accuracy1, CreatedAt: time.Now()}
	r2 := &entity.ForecastResult{ID: uuid.New().String(), StationID: "station-a", ForecastType: entity.ForecastTypeShortTerm, TargetTime: time.Now().Add(-30 * time.Minute), PredictedPower: 300.0, ActualPower: &actualPower2, Accuracy: &accuracy2, CreatedAt: time.Now()}
	err := repo.Create(ctx, r1)
	require.NoError(t, err)
	err = repo.Create(ctx, r2)
	require.NoError(t, err)

	stats, err := repo.GetAccuracyStats(ctx, "station-a", nil, time.Now().Add(-2*time.Hour), time.Now())
	assert.NoError(t, err)
	assert.Equal(t, "station-a", stats.StationID)
	assert.Equal(t, int64(2), stats.TotalPoints)
	assert.Greater(t, stats.RMSE, 0.0)
}

func TestForecastResultRepository_RealDB_GetAccuracyStats_NoData(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewForecastResultRepository(db)
	ctx := context.Background()

	stats, err := repo.GetAccuracyStats(ctx, "nonexistent", nil, time.Now().Add(-2*time.Hour), time.Now())
	assert.NoError(t, err)
	assert.Equal(t, "nonexistent", stats.StationID)
	assert.Equal(t, int64(0), stats.TotalPoints)
}

func TestFaultDetectionResultRepository_RealDB_CRUD(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewFaultDetectionResultRepository(db)
	ctx := context.Background()

	result := entity.NewFaultDetectionResult("device-001", "temperature", entity.FaultSeverityWarning, 0.9, "test", "1.0")
	result.ID = uuid.New().String()
	err := repo.Create(ctx, result)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, result.ID)
	assert.NoError(t, err)
	assert.Equal(t, "device-001", found.DeviceID)
	assert.Equal(t, entity.FaultSeverityWarning, found.Severity)
}

func TestFaultDetectionResultRepository_RealDB_ListByDevice(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewFaultDetectionResultRepository(db)
	ctx := context.Background()

	r1 := entity.NewFaultDetectionResult("device-l", "temp", entity.FaultSeverityWarning, 0.9, "test", "1.0")
	r1.ID = uuid.New().String()
	r2 := entity.NewFaultDetectionResult("device-l", "voltage", entity.FaultSeverityCritical, 0.95, "test", "1.0")
	r2.ID = uuid.New().String()
	err := repo.Create(ctx, r1)
	require.NoError(t, err)
	err = repo.Create(ctx, r2)
	require.NoError(t, err)

	results, total, err := repo.ListByDevice(ctx, "device-l", nil, nil, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)

	warning := entity.FaultSeverityWarning
	results, total, err = repo.ListByDevice(ctx, "device-l", &warning, nil, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
}

func TestFaultDetectionResultRepository_RealDB_UpdateStatus(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewFaultDetectionResultRepository(db)
	ctx := context.Background()

	result := entity.NewFaultDetectionResult("device-u", "temp", entity.FaultSeverityWarning, 0.9, "test", "1.0")
	result.ID = uuid.New().String()
	err := repo.Create(ctx, result)
	require.NoError(t, err)

	err = repo.UpdateStatus(ctx, result.ID, entity.FaultDetectionStatus(2))
	assert.NoError(t, err)
}

func TestFaultDetectionResultRepository_RealDB_CountBySeverity(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewFaultDetectionResultRepository(db)
	ctx := context.Background()

	r1 := entity.NewFaultDetectionResult("device-c", "temp", entity.FaultSeverityWarning, 0.9, "test", "1.0")
	r1.ID = uuid.New().String()
	err := repo.Create(ctx, r1)
	require.NoError(t, err)

	counts, err := repo.CountBySeverity(ctx, nil)
	assert.NoError(t, err)
	assert.NotNil(t, counts)
}

func TestFaultDetectionResultRepository_RealDB_LinkWorkOrder(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewFaultDetectionResultRepository(db)
	ctx := context.Background()

	result := entity.NewFaultDetectionResult("device-w", "temp", entity.FaultSeverityCritical, 0.9, "test", "1.0")
	result.ID = uuid.New().String()
	err := repo.Create(ctx, result)
	require.NoError(t, err)

	err = repo.LinkWorkOrder(ctx, result.ID, "wo-001")
	assert.NoError(t, err)
}

func TestFaultDetectionResultRepository_RealDB_ListByStation(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewFaultDetectionResultRepository(db)
	ctx := context.Background()

	r1 := entity.NewFaultDetectionResult("device-s", "temp", entity.FaultSeverityWarning, 0.9, "test", "1.0")
	r1.ID = uuid.New().String()
	err := repo.Create(ctx, r1)
	require.NoError(t, err)

	results, total, err := repo.ListByStation(ctx, "station-s", nil, 0, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(0))
	assert.NotNil(t, results)
}

func TestFaultDetectionResultRepository_RealDB_UpdateRootCause(t *testing.T) {
	db := setupTestDBWithMigrate(t)
	repo := NewFaultDetectionResultRepository(db)
	ctx := context.Background()

	result := entity.NewFaultDetectionResult("device-rc", "temp", entity.FaultSeverityWarning, 0.9, "test", "1.0")
	result.ID = uuid.New().String()
	err := repo.Create(ctx, result)
	require.NoError(t, err)

	err = repo.UpdateRootCause(ctx, result.ID, "overheating due to fan failure")
	assert.NoError(t, err)
}
