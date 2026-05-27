package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	require.NotNil(t, hub)
	assert.Equal(t, 0, hub.GetClientCount())
}

func TestHub_Run_RegisterUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer func() {
		for c := range hub.clients {
			hub.unregister <- c
		}
	}()

	client := NewClient("c1", "user1", "st1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, hub.GetClientCount())

	hub.unregister <- client
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, hub.GetClientCount())
}

func TestHub_BroadcastToAll(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := NewClient("c1", "user1", "st1", hub)
	client2 := NewClient("c2", "user2", "st2", hub)
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	hub.BroadcastToAll(&Message{Type: "test", Data: "hello"})
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 2, hub.GetClientCount())

	hub.unregister <- client1
	hub.unregister <- client2
}

func TestHub_BroadcastToStation(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := NewClient("c1", "user1", "st1", hub)
	client2 := NewClient("c2", "user2", "st2", hub)
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	hub.BroadcastToStation("st1", &Message{Type: "station_data", Data: "data"})
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 2, hub.GetClientCount())

	hub.unregister <- client1
	hub.unregister <- client2
}

func TestHub_BroadcastToUser(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := NewClient("c1", "user1", "st1", hub)
	client2 := NewClient("c2", "user2", "st2", hub)
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	hub.BroadcastToUser("user1", &Message{Type: "private", Data: "secret"})
	time.Sleep(50 * time.Millisecond)

	hub.unregister <- client1
	hub.unregister <- client2
}

func TestHub_GetClientCount(t *testing.T) {
	hub := NewHub()
	assert.Equal(t, 0, hub.GetClientCount())

	go hub.Run()
	client := NewClient("c1", "user1", "st1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, hub.GetClientCount())

	hub.unregister <- client
}

func TestHub_GetStationClientCount(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := NewClient("c1", "user1", "st1", hub)
	client2 := NewClient("c2", "user2", "st1", hub)
	client3 := NewClient("c3", "user3", "st2", hub)
	hub.register <- client1
	hub.register <- client2
	hub.register <- client3
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 2, hub.GetStationClientCount("st1"))
	assert.Equal(t, 1, hub.GetStationClientCount("st2"))
	assert.Equal(t, 0, hub.GetStationClientCount("st3"))

	hub.unregister <- client1
	hub.unregister <- client2
	hub.unregister <- client3
}

func TestNewClient(t *testing.T) {
	hub := NewHub()
	client := NewClient("c1", "user1", "st1", hub)
	assert.Equal(t, "c1", client.ID)
	assert.Equal(t, "user1", client.UserID)
	assert.Equal(t, "st1", client.StationID)
	assert.NotNil(t, client.Send)
	assert.Equal(t, hub, client.Hub)
}

func TestClient_ReadPump(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("c1", "user1", "st1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	received := make([]byte, 0)
	go client.ReadPump(func(msg []byte) {
		received = append(received, msg...)
	})

	client.Send <- []byte("test message")
	time.Sleep(50 * time.Millisecond)

	hub.unregister <- client
}

func TestMessage_Struct(t *testing.T) {
	msg := &Message{
		Type:      "alarm",
		StationID: "st1",
		Data:      map[string]interface{}{"key": "value"},
		Timestamp: time.Now().Unix(),
	}
	assert.Equal(t, "alarm", msg.Type)
	assert.Equal(t, "st1", msg.StationID)
}

func TestNewWSHandler(t *testing.T) {
	hub := NewHub()
	handler := NewWSHandler(hub)
	require.NotNil(t, handler)
}

func TestNewRealtimeDataService(t *testing.T) {
	hub := NewHub()
	service := NewRealtimeDataService(hub)
	require.NotNil(t, service)
}

func TestRealtimeDataService_BroadcastStationData(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	service := NewRealtimeDataService(hub)

	data := &StationData{
		StationID: "st1",
		Timestamp: time.Now().Unix(),
		Power:     850.5,
		Energy:    5000.2,
		DeviceStatus: map[string]string{
			"device_001": "online",
		},
	}
	service.BroadcastStationData(data)
}

func TestRealtimeDataService_BroadcastAlarm(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	service := NewRealtimeDataService(hub)

	alarm := &AlarmInfo{
		ID:        "alarm1",
		DeviceID:  "dev1",
		Level:     3,
		Message:   "High temperature",
		Value:     95.5,
		Threshold: 90.0,
	}
	service.BroadcastAlarm("st1", alarm)
}

func TestRealtimeDataService_BroadcastDeviceStatus(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	service := NewRealtimeDataService(hub)

	service.BroadcastDeviceStatus("st1", "dev1", "offline")
}

func TestRealtimeDataService_PushData(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	service := NewRealtimeDataService(hub)

	data := &StationData{
		StationID: "st1",
		Timestamp: time.Now().Unix(),
		Power:     100.0,
	}
	err := service.PushData(context.Background(), data)
	require.NoError(t, err)
}

func TestRealtimeDataService_GetConnectedClients(t *testing.T) {
	hub := NewHub()
	service := NewRealtimeDataService(hub)

	count := service.GetConnectedClients("st1")
	assert.Equal(t, 0, count)
}

func TestRealtimeDataService_StartStopMockDataStream(t *testing.T) {
	hub := NewHub()
	service := NewRealtimeDataService(hub)

	service.StartMockDataStream("st1")
	service.StopMockDataStream("st1")
}

func TestStationData_Struct(t *testing.T) {
	data := &StationData{
		StationID: "st1",
		Timestamp: time.Now().Unix(),
		Power:     850.5,
		Energy:    5000.2,
		DeviceStatus: map[string]string{"dev1": "online"},
		Alarms: []AlarmInfo{{ID: "a1", Level: 3}},
	}
	assert.Equal(t, "st1", data.StationID)
	assert.Equal(t, 850.5, data.Power)
}

func TestAlarmInfo_Struct(t *testing.T) {
	info := &AlarmInfo{
		ID:        "a1",
		DeviceID:  "dev1",
		Level:     3,
		Message:   "test",
		Value:     95.5,
		Threshold: 90.0,
	}
	assert.Equal(t, "a1", info.ID)
	assert.Equal(t, 95.5, info.Value)
}
