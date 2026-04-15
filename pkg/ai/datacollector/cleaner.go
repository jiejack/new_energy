package datacollector

import (
	"context"
	"math"
	"sort"
	"time"
)

type SimpleCleaner struct {
}

func NewSimpleCleaner() *SimpleCleaner {
	return &SimpleCleaner{}
}

func (c *SimpleCleaner) Deduplicate(ctx context.Context, points []*DataPoint) ([]*DataPoint, error) {
	if len(points) <= 1 {
		return points, nil
	}

	type key struct {
		timestamp time.Time
		stationID string
		deviceID  string
		metric    string
	}

	seen := make(map[key]bool)
	result := make([]*DataPoint, 0, len(points))

	for _, point := range points {
		k := key{
			timestamp: point.Timestamp.Truncate(time.Second),
			stationID: point.StationID,
			deviceID:  point.DeviceID,
			metric:    point.Metric,
		}

		if !seen[k] {
			seen[k] = true
			result = append(result, point)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	return result, nil
}

func (c *SimpleCleaner) FillMissing(ctx context.Context, points []*DataPoint, method string) ([]*DataPoint, error) {
	if len(points) <= 1 {
		return points, nil
	}

	sortedPoints := make([]*DataPoint, len(points))
	copy(sortedPoints, points)
	sort.Slice(sortedPoints, func(i, j int) bool {
		return sortedPoints[i].Timestamp.Before(sortedPoints[j].Timestamp)
	})

	result := make([]*DataPoint, 0, len(sortedPoints))

	for i := 0; i < len(sortedPoints); i++ {
		point := sortedPoints[i]
		
		if point.Quality == QualityMissing {
			switch method {
			case "forward":
				if i > 0 {
					point.Value = sortedPoints[i-1].Value
					point.Quality = QualityGood
					point.Metadata = map[string]interface{}{
						"filled_by": "forward",
						"source_index": i - 1,
					}
				}
			case "backward":
				if i < len(sortedPoints)-1 {
					point.Value = sortedPoints[i+1].Value
					point.Quality = QualityGood
					point.Metadata = map[string]interface{}{
						"filled_by": "backward",
						"source_index": i + 1,
					}
				}
			case "linear":
				if i > 0 && i < len(sortedPoints)-1 {
					prev := sortedPoints[i-1]
					next := sortedPoints[i+1]
					
					if prev.Quality != QualityMissing && next.Quality != QualityMissing {
						timeDiff := next.Timestamp.Sub(prev.Timestamp)
						pointDiff := point.Timestamp.Sub(prev.Timestamp)
						ratio := float64(pointDiff) / float64(timeDiff)
						
						point.Value = prev.Value + (next.Value-prev.Value)*ratio
						point.Quality = QualityGood
						point.Metadata = map[string]interface{}{
							"filled_by": "linear",
							"prev_index": i - 1,
							"next_index": i + 1,
						}
					}
				}
			}
		}
		
		result = append(result, point)
	}

	return result, nil
}

func (c *SimpleCleaner) RemoveOutliers(ctx context.Context, points []*DataPoint, method string) ([]*DataPoint, error) {
	if len(points) <= 3 {
		return points, nil
	}

	values := make([]float64, 0, len(points))
	validPoints := make([]*DataPoint, 0, len(points))
	
	for _, point := range points {
		if point.Quality != QualityMissing && !math.IsNaN(point.Value) && !math.IsInf(point.Value, 0) {
			values = append(values, point.Value)
			validPoints = append(validPoints, point)
		}
	}

	if len(values) <= 3 {
		return points, nil
	}

	var lowerBound, upperBound float64

	switch method {
	case "3sigma":
		mean, stdDev := calculateMeanAndStdDev(values)
		lowerBound = mean - 3*stdDev
		upperBound = mean + 3*stdDev
	case "iqr":
		q1, q3 := calculateQuartiles(values)
		iqr := q3 - q1
		lowerBound = q1 - 1.5*iqr
		upperBound = q3 + 1.5*iqr
	default:
		return points, nil
	}

	result := make([]*DataPoint, 0, len(points))
	for _, point := range points {
		if point.Quality == QualityMissing || math.IsNaN(point.Value) || math.IsInf(point.Value, 0) {
			result = append(result, point)
			continue
		}
		
		if point.Value >= lowerBound && point.Value <= upperBound {
			result = append(result, point)
		} else {
			point.Quality = QualityBad
			point.Metadata = map[string]interface{}{
				"removed_as": "outlier",
				"method": method,
				"lower_bound": lowerBound,
				"upper_bound": upperBound,
			}
			result = append(result, point)
		}
	}

	return result, nil
}

type SlidingWindowCleaner struct {
	windowSize time.Duration
}

func NewSlidingWindowCleaner(windowSize time.Duration) *SlidingWindowCleaner {
	return &SlidingWindowCleaner{
		windowSize: windowSize,
	}
}

func (swc *SlidingWindowCleaner) DeduplicateWithWindow(ctx context.Context, points []*DataPoint) ([]*DataPoint, error) {
	if len(points) <= 1 {
		return points, nil
	}

	sortedPoints := make([]*DataPoint, len(points))
	copy(sortedPoints, points)
	sort.Slice(sortedPoints, func(i, j int) bool {
		return sortedPoints[i].Timestamp.Before(sortedPoints[j].Timestamp)
	})

	result := make([]*DataPoint, 0, len(sortedPoints))
	type key struct {
		stationID string
		deviceID  string
		metric    string
	}
	windowMap := make(map[key]*DataPoint)

	for i, point := range sortedPoints {
		k := key{
			stationID: point.StationID,
			deviceID:  point.DeviceID,
			metric:    point.Metric,
		}

		for j := i - 1; j >= 0; j-- {
			if sortedPoints[i].Timestamp.Sub(sortedPoints[j].Timestamp) > swc.windowSize {
				break
			}
			if sortedPoints[j].StationID == point.StationID && 
			   sortedPoints[j].DeviceID == point.DeviceID && 
			   sortedPoints[j].Metric == point.Metric {
				delete(windowMap, k)
			}
		}

		if existing, exists := windowMap[k]; exists {
			if point.Timestamp.Sub(existing.Timestamp) <= swc.windowSize {
				continue
			}
		}

		windowMap[k] = point
		result = append(result, point)
	}

	return result, nil
}

type AdvancedCleaner struct {
	simpleCleaner *SimpleCleaner
}

func NewAdvancedCleaner() *AdvancedCleaner {
	return &AdvancedCleaner{
		simpleCleaner: NewSimpleCleaner(),
	}
}

func (ac *AdvancedCleaner) CleanPipeline(ctx context.Context, points []*DataPoint, opts CleanOptions) ([]*DataPoint, error) {
	var err error
	result := make([]*DataPoint, len(points))
	copy(result, points)

	if opts.Deduplicate {
		result, err = ac.simpleCleaner.Deduplicate(ctx, result)
		if err != nil {
			return nil, err
		}
	}

	if opts.RemoveOutliers {
		result, err = ac.simpleCleaner.RemoveOutliers(ctx, result, opts.OutlierMethod)
		if err != nil {
			return nil, err
		}
	}

	if opts.FillMissing {
		result, err = ac.simpleCleaner.FillMissing(ctx, result, opts.FillMethod)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

type CleanOptions struct {
	Deduplicate    bool   `json:"deduplicate"`
	RemoveOutliers bool   `json:"remove_outliers"`
	FillMissing    bool   `json:"fill_missing"`
	OutlierMethod  string `json:"outlier_method"`
	FillMethod     string `json:"fill_method"`
}
