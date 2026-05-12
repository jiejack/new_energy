package forecast

import (
	"math"
	"time"
)

type PredictionWithActual struct {
	Predicted float64   `json:"predicted"`
	Actual    float64   `json:"actual"`
	Time      time.Time `json:"time"`
}

type EvaluationReport struct {
	TotalPoints   int              `json:"total_points"`
	RMSE          float64          `json:"rmse"`
	MAE           float64          `json:"mae"`
	Bias          float64          `json:"bias"`
	Accuracy      float64          `json:"accuracy"`
	ByPeriod      map[string]Stats `json:"by_period"`
	InstalledCap  float64          `json:"installed_capacity"`
}

type Stats struct {
	Count    int     `json:"count"`
	RMSE     float64 `json:"rmse"`
	MAE      float64 `json:"mae"`
	Accuracy float64 `json:"accuracy"`
}

type ForecastEvaluator interface {
	Evaluate(predictions []PredictionWithActual, installedCapacity float64) (*EvaluationReport, error)
	CalculateAccuracy(rmse, installedCapacity float64) float64
}

type forecastEvaluator struct{}

func NewForecastEvaluator() ForecastEvaluator {
	return &forecastEvaluator{}
}

func (e *forecastEvaluator) Evaluate(predictions []PredictionWithActual, installedCapacity float64) (*EvaluationReport, error) {
	if len(predictions) == 0 {
		return &EvaluationReport{}, nil
	}

	var sumSquaredErr, sumAbsErr, sumErr float64
	for _, p := range predictions {
		err := p.Predicted - p.Actual
		sumSquaredErr += err * err
		sumAbsErr += abs(err)
		sumErr += err
	}

	n := float64(len(predictions))
	rmse := math.Sqrt(sumSquaredErr / n)
	mae := sumAbsErr / n
	bias := sumErr / n

	return &EvaluationReport{
		TotalPoints:  len(predictions),
		RMSE:         rmse,
		MAE:          mae,
		Bias:         bias,
		Accuracy:     e.CalculateAccuracy(rmse, installedCapacity),
		InstalledCap: installedCapacity,
		ByPeriod:     make(map[string]Stats),
	}, nil
}

func (e *forecastEvaluator) CalculateAccuracy(rmse, installedCapacity float64) float64 {
	if installedCapacity <= 0 {
		return 0
	}
	accuracy := 1 - rmse/installedCapacity
	if accuracy < 0 {
		return 0
	}
	if accuracy > 1 {
		return 1
	}
	return accuracy
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
