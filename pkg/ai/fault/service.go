package fault

import (
	"context"
	"fmt"
	"time"
)

type FaultServiceImpl struct {
	detectors   map[string]FaultDetector
	classifiers map[string]FaultClassifier
	assessors   map[string]HealthAssessor
	faultEvents []*FaultEvent
	createdAt   time.Time
}

func NewFaultService() *FaultServiceImpl {
	return &FaultServiceImpl{
		detectors:   make(map[string]FaultDetector),
		classifiers: make(map[string]FaultClassifier),
		assessors:   make(map[string]HealthAssessor),
		createdAt:   time.Now(),
	}
}

func (s *FaultServiceImpl) RegisterDetector(key string, detector FaultDetector) {
	s.detectors[key] = detector
}

func (s *FaultServiceImpl) RegisterClassifier(key string, classifier FaultClassifier) {
	s.classifiers[key] = classifier
}

func (s *FaultServiceImpl) RegisterAssessor(key string, assessor HealthAssessor) {
	s.assessors[key] = assessor
}

func (s *FaultServiceImpl) DetectAnomalies(ctx context.Context, data []*TimeSeriesData) ([]*Anomaly, error) {
	var allAnomalies []*Anomaly
	
	// 使用所有注册的检测器进行异常检测
	for key, detector := range s.detectors {
		anomalies, err := detector.Detect(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("detector %s failed: %w", key, err)
		}
		allAnomalies = append(allAnomalies, anomalies...)
	}
	
	// 去重异常
	uniqueAnomalies := s.deduplicateAnomalies(allAnomalies)
	
	// 为每个异常创建故障事件
	for _, anomaly := range uniqueAnomalies {
		s.createFaultEvent(anomaly, nil, nil, nil)
	}
	
	return uniqueAnomalies, nil
}

func (s *FaultServiceImpl) ClassifyFaults(ctx context.Context, anomalies []*Anomaly) ([]*FaultClassification, error) {
	var classifications []*FaultClassification
	
	// 使用所有注册的分类器进行故障分类
	for _, anomaly := range anomalies {
		for key, classifier := range s.classifiers {
			classification, err := classifier.Classify(ctx, anomaly)
			if err != nil {
				return nil, fmt.Errorf("classifier %s failed: %w", key, err)
			}
			classifications = append(classifications, classification)
			
			// 更新对应的故障事件
			s.updateFaultEvent(anomaly.ID, classification, nil, nil)
		}
	}
	
	return classifications, nil
}

func (s *FaultServiceImpl) AssessHealth(ctx context.Context, deviceID string) (*HealthAssessment, error) {
	// 找到设备对应的评估器
	var assessor HealthAssessor
	for key, a := range s.assessors {
		if key == deviceID {
			assessor = a
			break
		}
	}
	
	if assessor == nil {
		return nil, fmt.Errorf("no assessor found for device %s", deviceID)
	}
	
	// 这里应该从数据存储中获取设备的历史数据
	// 简化实现，实际项目中需要从数据库或缓存中获取
	var data []*TimeSeriesData
	
	assessment, err := assessor.Assess(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("health assessment failed: %w", err)
	}
	
	// 更新相关的故障事件
	s.updateFaultEventsWithAssessment(deviceID, assessment)
	
	return assessment, nil
}

func (s *FaultServiceImpl) PredictRUL(ctx context.Context, deviceID string) (*RULPrediction, error) {
	// 找到设备对应的评估器
	var assessor HealthAssessor
	for key, a := range s.assessors {
		if key == deviceID {
			assessor = a
			break
		}
	}
	
	if assessor == nil {
		return nil, fmt.Errorf("no assessor found for device %s", deviceID)
	}
	
	// 这里应该从数据存储中获取设备的历史数据
	// 简化实现，实际项目中需要从数据库或缓存中获取
	var data []*TimeSeriesData
	
	rul, err := assessor.PredictRUL(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("RUL prediction failed: %w", err)
	}
	
	// 更新相关的故障事件
	s.updateFaultEventsWithRUL(deviceID, rul)
	
	return rul, nil
}

func (s *FaultServiceImpl) GetFaultEvents(ctx context.Context, deviceID string, startTime, endTime time.Time) ([]*FaultEvent, error) {
	var filteredEvents []*FaultEvent
	
	for _, event := range s.faultEvents {
		if event.DeviceID == deviceID && 
		   event.Timestamp.After(startTime) && 
		   event.Timestamp.Before(endTime) {
			filteredEvents = append(filteredEvents, event)
		}
	}
	
	return filteredEvents, nil
}

func (s *FaultServiceImpl) deduplicateAnomalies(anomalies []*Anomaly) []*Anomaly {
	// 基于时间戳、设备ID和指标去重
	seen := make(map[string]bool)
	var unique []*Anomaly
	
	for _, anomaly := range anomalies {
		key := fmt.Sprintf("%s-%s-%d", anomaly.DeviceID, anomaly.Metric, anomaly.Timestamp.Unix())
		if !seen[key] {
			seen[key] = true
			unique = append(unique, anomaly)
		}
	}
	
	return unique
}

func (s *FaultServiceImpl) createFaultEvent(anomaly *Anomaly, classification *FaultClassification, assessment *HealthAssessment, rul *RULPrediction) {
	event := &FaultEvent{
		ID:             fmt.Sprintf("event-%s", anomaly.ID),
		DeviceID:       anomaly.DeviceID,
		Anomaly:        anomaly,
		Classification: classification,
		Assessment:     assessment,
		RUL:            rul,
		Timestamp:      time.Now(),
		Status:         "active",
		Actions:        []string{"待处理"},
		AdditionalInfo: map[string]interface{}{"created_by": "FaultService"},
	}
	
	s.faultEvents = append(s.faultEvents, event)
	// 限制事件数量
	if len(s.faultEvents) > 1000 {
		s.faultEvents = s.faultEvents[len(s.faultEvents)-1000:]
	}
}

func (s *FaultServiceImpl) updateFaultEvent(anomalyID string, classification *FaultClassification, assessment *HealthAssessment, rul *RULPrediction) {
	for _, event := range s.faultEvents {
		if event.Anomaly != nil && event.Anomaly.ID == anomalyID {
			if classification != nil {
				event.Classification = classification
				event.Status = "classified"
				event.Actions = append(event.Actions, "已分类")
			}
			if assessment != nil {
				event.Assessment = assessment
				event.Status = "assessed"
				event.Actions = append(event.Actions, "已评估")
			}
			if rul != nil {
				event.RUL = rul
				event.Status = "rul_predicted"
				event.Actions = append(event.Actions, "已预测RUL")
			}
			break
		}
	}
}

func (s *FaultServiceImpl) updateFaultEventsWithAssessment(deviceID string, assessment *HealthAssessment) {
	for _, event := range s.faultEvents {
		if event.DeviceID == deviceID && event.Status == "active" {
			event.Assessment = assessment
			event.Status = "assessed"
			event.Actions = append(event.Actions, "已评估")
		}
	}
}

func (s *FaultServiceImpl) updateFaultEventsWithRUL(deviceID string, rul *RULPrediction) {
	for _, event := range s.faultEvents {
		if event.DeviceID == deviceID && event.Status == "assessed" {
			event.RUL = rul
			event.Status = "rul_predicted"
			event.Actions = append(event.Actions, "已预测RUL")
		}
	}
}

// GetServiceInfo 返回故障服务的信息
func (s *FaultServiceImpl) GetServiceInfo() map[string]interface{} {
	return map[string]interface{}{
		"service_name":     "FaultService",
		"version":          "1.0.0",
		"created_at":       s.createdAt,
		"detectors_count":  len(s.detectors),
		"classifiers_count": len(s.classifiers),
		"assessors_count":  len(s.assessors),
		"events_count":     len(s.faultEvents),
	}
}