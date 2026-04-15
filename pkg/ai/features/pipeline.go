package features

import (
	"context"
	"math"
	"sort"
)

type SimpleFeaturePipeline struct {
	config           *FeatureConfig
	timeExtractor    *ConfigurableTimeExtractor
	lagExtractor     *SimpleLagExtractor
	rollingExtractor *SimpleRollingExtractor
	trendExtractor   *SimpleTrendExtractor
}

func NewFeaturePipeline(config *FeatureConfig) *SimpleFeaturePipeline {
	if config.RollingFeatures.Alpha == 0 {
		config.RollingFeatures.Alpha = 0.3
	}
	
	return &SimpleFeaturePipeline{
		config:           config,
		timeExtractor:    NewConfigurableTimeExtractor(config.TimeFeatures),
		lagExtractor:     NewSimpleLagExtractor(),
		rollingExtractor: NewSimpleRollingExtractor(),
		trendExtractor:   NewSimpleTrendExtractor(),
	}
}

func (p *SimpleFeaturePipeline) ExtractFeatures(ctx context.Context, dataPoints []*DataPointWithValue) ([]*TimeSeriesFeatures, error) {
	if len(dataPoints) == 0 {
		return []*TimeSeriesFeatures{}, nil
	}

	values := make([]float64, len(dataPoints))
	for i, dp := range dataPoints {
		values[i] = dp.Value
	}

	result := make([]*TimeSeriesFeatures, len(dataPoints))

	for i := range dataPoints {
		features := &TimeSeriesFeatures{
			Timestamp: dataPoints[i].Timestamp,
		}

		if p.config.TimeFeatures.Enabled {
			features.TimeFeatures = p.timeExtractor.Extract(dataPoints[i].Timestamp)
		}

		if p.config.LagFeatures.Enabled && len(p.config.LagFeatures.Lags) > 0 {
			lagValues := values[:i+1]
			lagFeatures, err := p.lagExtractor.Extract(lagValues, p.config.LagFeatures.Lags)
			if err == nil && len(lagFeatures) > 0 {
				features.LagFeatures = lagFeatures
			}
		}

		if p.config.RollingFeatures.Enabled && len(p.config.RollingFeatures.Windows) > 0 {
			for _, window := range p.config.RollingFeatures.Windows {
				if i+1 >= window {
					windowValues := values[i-window+1 : i+1]
					rollingFeatures, err := p.rollingExtractor.Extract(windowValues, p.config.RollingFeatures)
					if err == nil {
						if features.RollingFeatures.SMA == nil {
							features.RollingFeatures = RollingFeatures{
								SMA:        make(map[int]float64),
								EWMA:       make(map[int]float64),
								StdDev:     make(map[int]float64),
								Variance:   make(map[int]float64),
								Max:        make(map[int]float64),
								Min:        make(map[int]float64),
								Percentile: make(map[int]map[int]float64),
							}
						}
						for w, v := range rollingFeatures.SMA {
							features.RollingFeatures.SMA[w] = v
						}
						for w, v := range rollingFeatures.EWMA {
							features.RollingFeatures.EWMA[w] = v
						}
						for w, v := range rollingFeatures.StdDev {
							features.RollingFeatures.StdDev[w] = v
						}
					}
				}
			}
		}

		if p.config.TrendFeatures.Enabled {
			if p.config.TrendFeatures.Linear && i+1 >= 2 {
				trendFeatures, err := p.trendExtractor.Extract(values[:i+1], p.config.TrendFeatures)
				if err == nil {
					features.TrendFeatures = trendFeatures
				}
			}
		}

		result[i] = features
	}

	return result, nil
}

type SimpleRollingExtractor struct{}

func NewSimpleRollingExtractor() *SimpleRollingExtractor {
	return &SimpleRollingExtractor{}
}

func (e *SimpleRollingExtractor) Extract(values []float64, config RollingFeatureConfig) (RollingFeatures, error) {
	features := RollingFeatures{
		SMA:        make(map[int]float64),
		EWMA:       make(map[int]float64),
		StdDev:     make(map[int]float64),
		Variance:   make(map[int]float64),
		Max:        make(map[int]float64),
		Min:        make(map[int]float64),
		Percentile: make(map[int]map[int]float64),
	}

	n := len(values)
	if n == 0 {
		return features, nil
	}

	for _, window := range config.Windows {
		if window > n {
			window = n
		}
		windowValues := values[n-window:]

		if config.SMA {
			sum := 0.0
			for _, v := range windowValues {
				sum += v
			}
			features.SMA[window] = sum / float64(window)
		}

		if config.EWMA {
			alpha := config.Alpha
			if alpha <= 0 || alpha >= 1 {
				alpha = 0.3
			}
			ewma := windowValues[0]
			for i := 1; i < len(windowValues); i++ {
				ewma = alpha*windowValues[i] + (1-alpha)*ewma
			}
			features.EWMA[window] = ewma
		}

		if config.StdDev || config.Variance {
			mean := 0.0
			for _, v := range windowValues {
				mean += v
			}
			mean /= float64(len(windowValues))

			variance := 0.0
			for _, v := range windowValues {
				diff := v - mean
				variance += diff * diff
			}
			variance /= float64(len(windowValues))

			if config.Variance {
				features.Variance[window] = variance
			}
			if config.StdDev {
				features.StdDev[window] = math.Sqrt(variance)
			}
		}

		if config.Max {
			maxVal := windowValues[0]
			for _, v := range windowValues[1:] {
				if v > maxVal {
					maxVal = v
				}
			}
			features.Max[window] = maxVal
		}

		if config.Min {
			minVal := windowValues[0]
			for _, v := range windowValues[1:] {
				if v < minVal {
					minVal = v
				}
			}
			features.Min[window] = minVal
		}

		if len(config.Percentiles) > 0 {
			sorted := make([]float64, len(windowValues))
			copy(sorted, windowValues)
			sort.Float64s(sorted)

			features.Percentile[window] = make(map[int]float64)
			for _, p := range config.Percentiles {
				idx := int(math.Floor(float64(len(sorted)-1) * float64(p) / 100.0))
				if idx >= 0 && idx < len(sorted) {
					features.Percentile[window][p] = sorted[idx]
				}
			}
		}
	}

	return features, nil
}

type SimpleTrendExtractor struct{}

func NewSimpleTrendExtractor() *SimpleTrendExtractor {
	return &SimpleTrendExtractor{}
}

func (e *SimpleTrendExtractor) Extract(values []float64, config TrendFeatureConfig) (TrendFeatures, error) {
	features := TrendFeatures{
		RollingSlope: make(map[int]float64),
	}

	n := len(values)
	if n < 2 {
		return features, nil
	}

	if config.Linear {
		sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
		for i := 0; i < n; i++ {
			x := float64(i)
			y := values[i]
			sumX += x
			sumY += y
			sumXY += x * y
			sumX2 += x * x
		}

		denominator := float64(n)*sumX2 - sumX*sumX
		if denominator != 0 {
			features.LinearSlope = (float64(n)*sumXY - sumX*sumY) / denominator
			features.LinearIntercept = (sumY - features.LinearSlope*sumX) / float64(n)

			yMean := sumY / float64(n)
			totalSS, residualSS := 0.0, 0.0
			for i := 0; i < n; i++ {
				predicted := features.LinearIntercept + features.LinearSlope*float64(i)
				totalSS += (values[i] - yMean) * (values[i] - yMean)
				residualSS += (values[i] - predicted) * (values[i] - predicted)
			}
			if totalSS != 0 {
				features.R2Score = 1 - residualSS/totalSS
			}
		}
	}

	for _, window := range config.RollingSlopes {
		if window > 1 && n >= window {
			windowValues := values[n-window:]
			sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
			for i := 0; i < window; i++ {
				x := float64(i)
				y := windowValues[i]
				sumX += x
				sumY += y
				sumXY += x * y
				sumX2 += x * x
			}
			denominator := float64(window)*sumX2 - sumX*sumX
			if denominator != 0 {
				features.RollingSlope[window] = (float64(window)*sumXY - sumX*sumY) / denominator
			}
		}
	}

	if config.Acceleration && n >= 3 {
		slopes := make([]float64, n-1)
		for i := 1; i < n; i++ {
			slopes[i-1] = values[i] - values[i-1]
		}
		if len(slopes) >= 2 {
			features.Acceleration = slopes[len(slopes)-1] - slopes[len(slopes)-2]
		}
	}

	return features, nil
}
