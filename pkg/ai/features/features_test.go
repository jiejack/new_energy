package features

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeFeatureExtractor(t *testing.T) {
	extractor := NewSimpleTimeExtractor()
	
	testTime, _ := time.Parse(time.RFC3339, "2024-06-15T14:30:00Z")
	features := extractor.Extract(testTime)
	
	assert.Equal(t, 14, features.Hour)
	assert.Equal(t, 15, features.Day)
	assert.Equal(t, 6, features.Weekday)
	assert.Equal(t, 6, features.Month)
	assert.Equal(t, 2024, features.Year)
	assert.Equal(t, 2, features.Quarter)
	assert.Equal(t, true, features.IsWeekend)
	assert.Equal(t, 3, features.Season)
}

func TestTimeFeatureExtractorWithHoliday(t *testing.T) {
	extractor := NewSimpleTimeExtractor()
	holidayDate := "2024-12-25"
	extractor.AddHoliday(holidayDate)
	
	testTime, _ := time.Parse("2006-01-02", holidayDate)
	features := extractor.Extract(testTime)
	
	assert.Equal(t, true, features.IsHoliday)
}

func TestConfigurableTimeExtractor(t *testing.T) {
	config := TimeFeatureConfig{
		Enabled: true,
		Hour:    true,
		Day:     true,
		Month:   true,
		Season:  true,
	}
	
	extractor := NewConfigurableTimeExtractor(config)
	testTime, _ := time.Parse(time.RFC3339, "2024-03-15T10:00:00Z")
	features := extractor.Extract(testTime)
	
	assert.Equal(t, 10, features.Hour)
	assert.Equal(t, 15, features.Day)
	assert.Equal(t, 3, features.Month)
	assert.Equal(t, 2, features.Season)
	assert.Equal(t, 0, features.Year)
}

func TestLagFeatureExtractor(t *testing.T) {
	extractor := NewSimpleLagExtractor()
	
	values := []float64{10, 20, 30, 40, 50}
	lags := []int{1, 2, 3}
	
	features, err := extractor.Extract(values, lags)
	
	assert.NoError(t, err)
	assert.Equal(t, 50.0, features[1])
	assert.Equal(t, 40.0, features[2])
	assert.Equal(t, 30.0, features[3])
}

func TestLagFeatureExtractorBatch(t *testing.T) {
	extractor := NewSimpleLagExtractor()
	
	values := []float64{10, 20, 30, 40}
	lags := []int{1, 2}
	
	features, err := extractor.ExtractBatch(values, lags)
	
	assert.NoError(t, err)
	assert.Len(t, features, 4)
	assert.Empty(t, features[0])
	assert.Equal(t, 10.0, features[1][1])
	assert.Equal(t, 20.0, features[2][1])
	assert.Equal(t, 10.0, features[2][2])
}

func TestRollingFeatureExtractor(t *testing.T) {
	extractor := NewSimpleRollingExtractor()
	
	config := RollingFeatureConfig{
		Enabled: true,
		Windows: []int{3},
		SMA:     true,
		EWMA:    true,
		StdDev:  true,
		Max:     true,
		Min:     true,
		Alpha:   0.3,
	}
	
	values := []float64{10, 20, 30}
	features, err := extractor.Extract(values, config)
	
	assert.NoError(t, err)
	assert.InDelta(t, 20.0, features.SMA[3], 0.001)
	assert.InDelta(t, 30.0, features.Max[3], 0.001)
	assert.InDelta(t, 10.0, features.Min[3], 0.001)
}

func TestTrendFeatureExtractor(t *testing.T) {
	extractor := NewSimpleTrendExtractor()
	
	config := TrendFeatureConfig{
		Enabled:       true,
		Linear:        true,
		RollingSlopes: []int{3},
		Acceleration:  true,
	}
	
	values := []float64{10, 20, 30, 40}
	features, err := extractor.Extract(values, config)
	
	assert.NoError(t, err)
	assert.InDelta(t, 10.0, features.LinearSlope, 0.001)
	assert.InDelta(t, 1.0, features.R2Score, 0.1)
}

func TestFeaturePipeline(t *testing.T) {
	config := &FeatureConfig{
		TimeFeatures: TimeFeatureConfig{
			Enabled: true,
			Hour:    true,
			Day:     true,
			Month:   true,
		},
		LagFeatures: LagFeatureConfig{
			Enabled: true,
			Lags:    []int{1, 2},
		},
		RollingFeatures: RollingFeatureConfig{
			Enabled: true,
			Windows: []int{2},
			SMA:     true,
		},
		TrendFeatures: TrendFeatureConfig{
			Enabled: true,
			Linear:  true,
		},
	}
	
	pipeline := NewFeaturePipeline(config)
	
	baseTime := time.Now()
	dataPoints := []*DataPointWithValue{
		{Timestamp: baseTime, Value: 10},
		{Timestamp: baseTime.Add(time.Hour), Value: 20},
		{Timestamp: baseTime.Add(2 * time.Hour), Value: 30},
	}
	
	features, err := pipeline.ExtractFeatures(context.Background(), dataPoints)
	
	assert.NoError(t, err)
	assert.Len(t, features, 3)
	assert.NotNil(t, features[2].TimeFeatures)
	assert.NotNil(t, features[2].LagFeatures)
	assert.NotNil(t, features[2].RollingFeatures)
	assert.NotNil(t, features[2].TrendFeatures)
}

func TestGetSeason(t *testing.T) {
	assert.Equal(t, 1, getSeason(time.December))
	assert.Equal(t, 1, getSeason(time.January))
	assert.Equal(t, 1, getSeason(time.February))
	assert.Equal(t, 2, getSeason(time.March))
	assert.Equal(t, 2, getSeason(time.April))
	assert.Equal(t, 2, getSeason(time.May))
	assert.Equal(t, 3, getSeason(time.June))
	assert.Equal(t, 3, getSeason(time.July))
	assert.Equal(t, 3, getSeason(time.August))
	assert.Equal(t, 4, getSeason(time.September))
	assert.Equal(t, 4, getSeason(time.October))
	assert.Equal(t, 4, getSeason(time.November))
}

func TestRollingPercentile(t *testing.T) {
	extractor := NewSimpleRollingExtractor()
	
	config := RollingFeatureConfig{
		Enabled:     true,
		Windows:     []int{5},
		Percentiles: []int{25, 50, 75},
	}
	
	values := []float64{10, 20, 30, 40, 50}
	features, err := extractor.Extract(values, config)
	
	assert.NoError(t, err)
	assert.NotNil(t, features.Percentile[5])
	assert.InDelta(t, 20.0, features.Percentile[5][25], 0.001)
	assert.InDelta(t, 30.0, features.Percentile[5][50], 0.001)
	assert.InDelta(t, 40.0, features.Percentile[5][75], 0.001)
}
