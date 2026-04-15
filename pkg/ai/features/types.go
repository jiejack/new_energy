package features

import (
	"context"
	"time"
)

type TimeSeriesFeatures struct {
	Timestamp       time.Time              `json:"timestamp"`
	TimeFeatures    TimeFeatures           `json:"time_features"`
	LagFeatures     map[int]float64        `json:"lag_features,omitempty"`
	RollingFeatures RollingFeatures       `json:"rolling_features,omitempty"`
	TrendFeatures   TrendFeatures          `json:"trend_features,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type TimeFeatures struct {
	Hour       int  `json:"hour"`
	Day        int  `json:"day"`
	Weekday    int  `json:"weekday"`
	Month      int  `json:"month"`
	Quarter    int  `json:"quarter"`
	Year       int  `json:"year"`
	DayOfYear  int  `json:"day_of_year"`
	WeekOfYear int  `json:"week_of_year"`
	IsWeekend  bool `json:"is_weekend"`
	IsHoliday  bool `json:"is_holiday"`
	Season     int  `json:"season"`
}

type RollingFeatures struct {
	SMA        map[int]float64       `json:"sma,omitempty"`
	EWMA       map[int]float64       `json:"ewma,omitempty"`
	StdDev     map[int]float64       `json:"std_dev,omitempty"`
	Variance   map[int]float64       `json:"variance,omitempty"`
	Max        map[int]float64       `json:"max,omitempty"`
	Min        map[int]float64       `json:"min,omitempty"`
	Percentile map[int]map[int]float64 `json:"percentile,omitempty"`
}

type TrendFeatures struct {
	LinearSlope     float64       `json:"linear_slope,omitempty"`
	LinearIntercept float64       `json:"linear_intercept,omitempty"`
	R2Score         float64       `json:"r2_score,omitempty"`
	RollingSlope    map[int]float64 `json:"rolling_slope,omitempty"`
	Acceleration    float64       `json:"acceleration,omitempty"`
}

type FeatureConfig struct {
	TimeFeatures    TimeFeatureConfig    `json:"time_features"`
	LagFeatures     LagFeatureConfig     `json:"lag_features"`
	RollingFeatures RollingFeatureConfig `json:"rolling_features"`
	TrendFeatures   TrendFeatureConfig   `json:"trend_features"`
}

type TimeFeatureConfig struct {
	Enabled    bool `json:"enabled"`
	Hour       bool `json:"hour"`
	Day        bool `json:"day"`
	Weekday    bool `json:"weekday"`
	Month      bool `json:"month"`
	Quarter    bool `json:"quarter"`
	Year       bool `json:"year"`
	DayOfYear  bool `json:"day_of_year"`
	WeekOfYear bool `json:"week_of_year"`
	IsWeekend  bool `json:"is_weekend"`
	Season     bool `json:"season"`
}

type LagFeatureConfig struct {
	Enabled bool  `json:"enabled"`
	Lags    []int `json:"lags"`
}

type RollingFeatureConfig struct {
	Enabled     bool   `json:"enabled"`
	Windows     []int  `json:"windows"`
	SMA         bool   `json:"sma"`
	EWMA        bool   `json:"ewma"`
	StdDev      bool   `json:"std_dev"`
	Variance    bool   `json:"variance"`
	Max         bool   `json:"max"`
	Min         bool   `json:"min"`
	Percentiles []int  `json:"percentiles"`
	Alpha       float64 `json:"alpha"`
}

type TrendFeatureConfig struct {
	Enabled       bool  `json:"enabled"`
	Linear        bool  `json:"linear"`
	RollingSlopes []int `json:"rolling_slopes"`
	Acceleration  bool  `json:"acceleration"`
}

type TimeFeatureExtractor interface {
	Extract(t time.Time) TimeFeatures
}

type LagFeatureExtractor interface {
	Extract(values []float64, lags []int) (map[int]float64, error)
}

type RollingFeatureExtractor interface {
	Extract(values []float64, config RollingFeatureConfig) (RollingFeatures, error)
}

type TrendFeatureExtractor interface {
	Extract(values []float64, config TrendFeatureConfig) (TrendFeatures, error)
}

type FeaturePipeline interface {
	ExtractFeatures(ctx context.Context, dataPoints []*DataPointWithValue) ([]*TimeSeriesFeatures, error)
}

type DataPointWithValue struct {
	Timestamp time.Time
	Value     float64
}
