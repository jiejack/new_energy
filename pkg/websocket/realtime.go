package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
)

type RealtimeDataService struct {
	hub            *Hub
	subscriptions  map[string][]string
	mu             sync.RWMutex
	dataChannels   map[string]chan *StationData
}

type StationData struct {
	StationID    string             `json:"station_id"`
	Timestamp    int64              `json:"timestamp"`
	Power        float64            `json:"power"`
	Energy       float64            `json:"energy"`
	DeviceStatus map[string]string  `json:"device_status"`
	Alarms       []AlarmInfo        `json:"alarms,omitempty"`
}

type AlarmInfo struct {
	ID        string  `json:"id"`
	DeviceID  string  `json:"device_id"`
	Level     int     `json:"level"`
	Message   string  `json:"message"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
}

type DataSubscriber interface {
	Subscribe(stationID string) <-chan *StationData
	Unsubscribe(stationID string, ch <-chan *StationData)
}

func NewRealtimeDataService(hub *Hub) *RealtimeDataService {
	s := &RealtimeDataService{
		hub:           hub,
		subscriptions: make(map[string][]string),
		dataChannels:  make(map[string]chan *StationData),
	}
	go s.startBroadcast()
	return s
}

func (s *RealtimeDataService) startBroadcast() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.broadcastAllStations()
	}
}

func (s *RealtimeDataService) broadcastAllStations() {
	s.mu.RLock()
	stationIDs := make([]string, 0, len(s.dataChannels))
	for id := range s.dataChannels {
		stationIDs = append(stationIDs, id)
	}
	s.mu.RUnlock()

	for _, stationID := range stationIDs {
		data := s.getMockStationData(stationID)
		s.BroadcastStationData(data)
	}
}

func (s *RealtimeDataService) getMockStationData(stationID string) *StationData {
	return &StationData{
		StationID: stationID,
		Timestamp: time.Now().Unix(),
		Power:     850.5 + float64(time.Now().Second()%10)*10,
		Energy:    5000.2 + float64(time.Now().Minute())*100,
		DeviceStatus: map[string]string{
			"device_001": "online",
			"device_002": "online",
			"device_003": "warning",
		},
	}
}

func (s *RealtimeDataService) BroadcastStationData(data *StationData) {
	msg := &Message{
		Type:      "station_data",
		StationID: data.StationID,
		Data:      data,
	}
	s.hub.BroadcastToStation(data.StationID, msg)
}

func (s *RealtimeDataService) BroadcastAlarm(stationID string, alarm *AlarmInfo) {
	msg := &Message{
		Type:      "alarm",
		StationID: stationID,
		Data:      alarm,
	}
	s.hub.BroadcastToStation(stationID, msg)
}

func (s *RealtimeDataService) BroadcastDeviceStatus(stationID, deviceID, status string) {
	msg := &Message{
		Type:      "device_status",
		StationID: stationID,
		Data: map[string]interface{}{
			"device_id": deviceID,
			"status":    status,
		},
	}
	s.hub.BroadcastToStation(stationID, msg)
}

func (s *RealtimeDataService) PushData(ctx context.Context, data *StationData) error {
	msg := &Message{
		Type:      "station_data",
		StationID: data.StationID,
		Data:      data,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	s.hub.broadcast <- msgBytes
	return nil
}

func (s *RealtimeDataService) GetConnectedClients(stationID string) int {
	return s.hub.GetStationClientCount(stationID)
}

func (s *RealtimeDataService) StartMockDataStream(stationID string) {
	s.mu.Lock()
	if _, exists := s.dataChannels[stationID]; !exists {
		s.dataChannels[stationID] = make(chan *StationData, 100)
	}
	s.mu.Unlock()

	log.Printf("Started mock data stream for station: %s", stationID)
}

func (s *RealtimeDataService) StopMockDataStream(stationID string) {
	s.mu.Lock()
	if ch, exists := s.dataChannels[stationID]; exists {
		close(ch)
		delete(s.dataChannels, stationID)
	}
	s.mu.Unlock()

	log.Printf("Stopped mock data stream for station: %s", stationID)
}
