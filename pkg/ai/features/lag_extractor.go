package features

import (
	"errors"
	"fmt"
)

type SimpleLagExtractor struct{}

func NewSimpleLagExtractor() *SimpleLagExtractor {
	return &SimpleLagExtractor{}
}

func (e *SimpleLagExtractor) Extract(values []float64, lags []int) (map[int]float64, error) {
	if len(values) == 0 {
		return nil, errors.New("values slice is empty")
	}

	result := make(map[int]float64)
	maxLag := 0
	for _, lag := range lags {
		if lag > maxLag {
			maxLag = lag
		}
	}

	if maxLag >= len(values) {
		return nil, fmt.Errorf("max lag %d exceeds data length %d", maxLag, len(values))
	}

	currentIndex := len(values) - 1
	for _, lag := range lags {
		if lag < 1 {
			return nil, fmt.Errorf("lag must be positive, got %d", lag)
		}
		if lag > len(values) {
			continue
		}
		index := currentIndex - lag + 1
		if index >= 0 && index < len(values) {
			result[lag] = values[index]
		}
	}

	return result, nil
}

func (e *SimpleLagExtractor) ExtractBatch(values []float64, lags []int) ([]map[int]float64, error) {
	if len(values) == 0 {
		return nil, errors.New("values slice is empty")
	}

	result := make([]map[int]float64, len(values))
	
	for i := range values {
		lagFeatures := make(map[int]float64)
		for _, lag := range lags {
			if lag > 0 && i-lag >= 0 {
				lagFeatures[lag] = values[i-lag]
			}
		}
		result[i] = lagFeatures
	}

	return result, nil
}
