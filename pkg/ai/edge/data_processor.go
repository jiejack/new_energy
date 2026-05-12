package edge

import (
	"context"
	"fmt"
	"time"
)

type DataProcessor interface {
	Process(ctx context.Context, rawData map[string]interface{}) (map[string]interface{}, error)
	Validate(ctx context.Context, data map[string]interface{}) bool
	Transform(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error)
}

type dataProcessor struct{}

func NewDataProcessor() DataProcessor {
	return &dataProcessor{}
}

func (p *dataProcessor) Process(ctx context.Context, rawData map[string]interface{}) (map[string]interface{}, error) {
	if !p.Validate(ctx, rawData) {
		return nil, fmt.Errorf("data validation failed")
	}
	return p.Transform(ctx, rawData)
}

func (p *dataProcessor) Validate(ctx context.Context, data map[string]interface{}) bool {
	return data != nil
}

func (p *dataProcessor) Transform(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for k, v := range data {
		result[k] = v
	}
	result["processed_at"] = time.Now().Unix()
	return result, nil
}
