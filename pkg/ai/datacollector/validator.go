package datacollector

import (
	"context"
	"math"
	"sort"
	"time"
)

type SimpleValidator struct {
	config ValidationConfig
}

func NewSimpleValidator(config ValidationConfig) *SimpleValidator {
	return &SimpleValidator{
		config: config,
	}
}

func (v *SimpleValidator) Validate(ctx context.Context, points []*DataPoint) ([]*DataPoint, error) {
	result := make([]*DataPoint, 0, len(points))
	
	for _, point := range points {
		validatedPoint := v.validateSinglePoint(point)
		result = append(result, validatedPoint)
	}
	
	return result, nil
}

func (v *SimpleValidator) validateSinglePoint(point *DataPoint) *DataPoint {
	if point.Quality == QualityUnknown {
		point.Quality = QualityGood
	}
	
	if v.config.CheckMissing {
		if math.IsNaN(point.Value) || math.IsInf(point.Value, 0) {
			point.Quality = QualityMissing
			return point
		}
	}
	
	if v.config.CheckRange && len(v.config.RangeConstraints) > 0 {
		if constraint, ok := v.config.RangeConstraints[point.Metric]; ok {
			if point.Value < constraint[0] || point.Value > constraint[1] {
				point.Quality = QualityBad
				return point
			}
		}
	}
	
	if v.config.CheckOutliers {
		if v.isOutlier(point, v.config.OutlierMethod) {
			if point.Quality == QualityGood {
				point.Quality = QualitySuspect
			}
		}
	}
	
	return point
}

func (v *SimpleValidator) isOutlier(point *DataPoint, method string) bool {
	switch method {
	case "3sigma":
		return v.isOutlier3Sigma(point)
	case "iqr":
		return v.isOutlierIQR(point)
	default:
		return false
	}
}

func (v *SimpleValidator) isOutlier3Sigma(point *DataPoint) bool {
	return false
}

func (v *SimpleValidator) isOutlierIQR(point *DataPoint) bool {
	return false
}

func (v *SimpleValidator) GenerateReport(ctx context.Context, stationID, deviceID, metric string, startTime, endTime time.Time) (*DataQualityReport, error) {
	report := &DataQualityReport{
		ReportDate: time.Now(),
		StationID:  stationID,
		DeviceID:   deviceID,
		Metric:     metric,
		CreatedAt:  time.Now(),
	}
	
	return report, nil
}

type BatchValidator struct {
	config ValidationConfig
}

func NewBatchValidator(config ValidationConfig) *BatchValidator {
	return &BatchValidator{config: config}
}

func (bv *BatchValidator) ValidateBatch(ctx context.Context, points []*DataPoint) ([]*DataPoint, error) {
	if len(points) == 0 {
		return points, nil
	}
	
	sortedPoints := make([]*DataPoint, len(points))
	copy(sortedPoints, points)
	sort.Slice(sortedPoints, func(i, j int) bool {
		return sortedPoints[i].Timestamp.Before(sortedPoints[j].Timestamp)
	})
	
	if bv.config.CheckOutliers {
		values := make([]float64, len(sortedPoints))
		for i, p := range sortedPoints {
			values[i] = p.Value
		}
		
		mean, stdDev := calculateMeanAndStdDev(values)
		q1, q3 := calculateQuartiles(values)
		
		for _, p := range sortedPoints {
			if bv.isOutlierBatch(p.Value, mean, stdDev, q1, q3, bv.config.OutlierMethod) {
				if p.Quality == QualityGood || p.Quality == QualityUnknown {
					p.Quality = QualitySuspect
				}
			}
		}
	}
	
	if bv.config.CheckContinuity {
		bv.checkContinuity(sortedPoints)
	}
	
	return sortedPoints, nil
}

func (bv *BatchValidator) isOutlierBatch(value, mean, stdDev, q1, q3 float64, method string) bool {
	switch method {
	case "3sigma":
		zScore := math.Abs((value - mean) / stdDev)
		return zScore > 3
	case "iqr":
		iqr := q3 - q1
		lowerBound := q1 - 1.5*iqr
		upperBound := q3 + 1.5*iqr
		return value < lowerBound || value > upperBound
	default:
		return false
	}
}

func (bv *BatchValidator) checkContinuity(points []*DataPoint) {
	if len(points) < 2 {
		return
	}
	
	for i := 1; i < len(points); i++ {
		prev := points[i-1]
		curr := points[i]
		
		timeDiff := curr.Timestamp.Sub(prev.Timestamp)
		if timeDiff > 2*time.Hour {
			if prev.Quality == QualityGood || prev.Quality == QualityUnknown {
				prev.Quality = QualitySuspect
			}
			if curr.Quality == QualityGood || curr.Quality == QualityUnknown {
				curr.Quality = QualitySuspect
			}
		}
		
		if curr.Metric == prev.Metric && curr.StationID == prev.StationID && curr.DeviceID == prev.DeviceID {
			valueDiff := math.Abs(curr.Value - prev.Value)
			avgValue := (math.Abs(curr.Value) + math.Abs(prev.Value)) / 2
			if avgValue > 0 && valueDiff/avgValue > 2.0 {
				if curr.Quality == QualityGood || curr.Quality == QualityUnknown {
					curr.Quality = QualitySuspect
				}
			}
		}
	}
}

func calculateMeanAndStdDev(values []float64) (mean, stdDev float64) {
	if len(values) == 0 {
		return 0, 0
	}
	
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))
	
	varianceSum := 0.0
	for _, v := range values {
		diff := v - mean
		varianceSum += diff * diff
	}
	stdDev = math.Sqrt(varianceSum / float64(len(values)))
	
	return mean, stdDev
}

func calculateQuartiles(values []float64) (q1, q3 float64) {
	if len(values) == 0 {
		return 0, 0
	}
	
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	
	q1Index := int(math.Floor(float64(len(sorted)) * 0.25))
	q3Index := int(math.Floor(float64(len(sorted)) * 0.75))
	
	if q1Index >= 0 && q1Index < len(sorted) {
		q1 = sorted[q1Index]
	}
	if q3Index >= 0 && q3Index < len(sorted) {
		q3 = sorted[q3Index]
	}
	
	return q1, q3
}
