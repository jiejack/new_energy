package websocket

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHubCov(t *testing.T) {
	hub := NewHub()
	require.NotNil(t, hub)
	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.broadcast)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
	assert.Equal(t, 0, hub.GetClientCount())
}

func TestHub_RegisterUnregisterCov(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer func() {
		hub.unregister <- &Client{ID: "stop", Send: make(chan []byte, 1)}
	}()

	client := NewClient("c1", "user1", "station1", hub)
	hub.register <- client

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, hub.GetClientCount())

	hub.unregister <- client
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, hub.GetClientCount())
}

func TestHub_UnregisterNonExistent(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("c-nonexistent", "user1", "station1", hub)
	hub.unregister <- client
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, hub.GetClientCount())
}

func TestHub_BroadcastToAllCov(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := NewClient("c1", "user1", "station1", hub)
	client2 := NewClient("c2", "user2", "station2", hub)
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	msg := &Message{Type: "test", Data: "hello"}
	hub.BroadcastToAll(msg)

	received1 := false
	select {
	case <-client1.Send:
		received1 = true
	case <-time.After(100 * time.Millisecond):
	}

	received2 := false
	select {
	case <-client2.Send:
		received2 = true
	case <-time.After(100 * time.Millisecond):
	}

	assert.True(t, received1)
	assert.True(t, received2)
}

func TestHub_BroadcastToStationCov(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := NewClient("c1", "user1", "station1", hub)
	client2 := NewClient("c2", "user2", "station2", hub)
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	msg := &Message{Type: "test", Data: "station hello"}
	hub.BroadcastToStation("station1", msg)

	select {
	case <-client1.Send:
	case <-time.After(100 * time.Millisecond):
		t.Error("client1 should have received message")
	}

	select {
	case <-client2.Send:
		t.Error("client2 should not have received message")
	default:
	}
}

func TestHub_BroadcastToStation_EmptyStationID(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := NewClient("c1", "user1", "", hub)
	client2 := NewClient("c2", "user2", "station2", hub)
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	msg := &Message{Type: "test", Data: "broadcast to all with empty station"}
	hub.BroadcastToStation("station1", msg)

	select {
	case <-client1.Send:
	case <-time.After(100 * time.Millisecond):
		t.Error("client1 with empty station should receive message")
	}
}

func TestHub_BroadcastToUserCov(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := NewClient("c1", "user1", "station1", hub)
	client2 := NewClient("c2", "user2", "station1", hub)
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	msg := &Message{Type: "test", Data: "user hello"}
	hub.BroadcastToUser("user1", msg)

	select {
	case <-client1.Send:
	case <-time.After(100 * time.Millisecond):
		t.Error("client1 should have received message")
	}

	select {
	case <-client2.Send:
		t.Error("client2 should not have received message for user1")
	default:
	}
}

func TestHub_GetStationClientCountCov(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := NewClient("c1", "user1", "station1", hub)
	client2 := NewClient("c2", "user2", "station1", hub)
	client3 := NewClient("c3", "user3", "station2", hub)
	hub.register <- client1
	hub.register <- client2
	hub.register <- client3
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 2, hub.GetStationClientCount("station1"))
	assert.Equal(t, 1, hub.GetStationClientCount("station2"))
	assert.Equal(t, 0, hub.GetStationClientCount("station3"))
}

func TestHub_BroadcastToAll_FullSendChannel(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("c-full", "user1", "station1", hub)
	client.Send = make(chan []byte, 1)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	for i := 0; i < 5; i++ {
		hub.BroadcastToAll(&Message{Type: "test", Data: i})
		time.Sleep(10 * time.Millisecond)
	}
}

func TestNewClientCov(t *testing.T) {
	hub := NewHub()
	client := NewClient("c1", "user1", "station1", hub)
	require.NotNil(t, client)
	assert.Equal(t, "c1", client.ID)
	assert.Equal(t, "user1", client.UserID)
	assert.Equal(t, "station1", client.StationID)
	assert.NotNil(t, client.Send)
	assert.Equal(t, hub, client.Hub)
}

func TestClient_ReadPumpCov(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("c1", "user1", "station1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	received := make([]byte, 0, 10)
	done := make(chan struct{})

	go func() {
		client.ReadPump(func(msg []byte) {
			received = append(received, msg...)
		})
		close(done)
	}()

	client.Send <- []byte("test message")
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, "test message", string(received))

	hub.unregister <- client
	select {
	case <-done:
	case <-time.After(1 * time.Second):
	}
}

func TestClient_WritePump(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("c1", "user1", "station1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		client.WritePump(func() ([]byte, bool) {
			return nil, false
		})
		close(done)
	}()

	hub.unregister <- client
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("WritePump should have stopped")
	}
}

func TestMessage_JSONCov(t *testing.T) {
	msg := &Message{
		Type:      "test",
		StationID: "station1",
		Data:      map[string]interface{}{"key": "value"},
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test")
	assert.Contains(t, string(data), "station1")

	var parsed Message
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "test", parsed.Type)
}

func TestNewWSHandlerCov(t *testing.T) {
	hub := NewHub()
	handler := NewWSHandler(hub)
	require.NotNil(t, handler)
	assert.Equal(t, hub, handler.hub)
}

func TestRealtimeDataService_New(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	service := NewRealtimeDataService(hub)
	require.NotNil(t, service)
	assert.NotNil(t, service.subscriptions)
	assert.NotNil(t, service.dataChannels)
}

func TestRealtimeDataService_BroadcastStationDataCov(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("c1", "user1", "station1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	service := NewRealtimeDataService(hub)

	data := &StationData{
		StationID: "station1",
		Timestamp: time.Now().Unix(),
		Power:     850.5,
		Energy:    5000.2,
		DeviceStatus: map[string]string{
			"device_001": "online",
		},
	}
	service.BroadcastStationData(data)

	select {
	case msg := <-client.Send:
		assert.Contains(t, string(msg), "station_data")
	case <-time.After(100 * time.Millisecond):
		t.Error("client should have received station data")
	}
}

func TestRealtimeDataService_BroadcastAlarmCov(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("c1", "user1", "station1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	service := NewRealtimeDataService(hub)

	alarm := &AlarmInfo{
		ID:        "alarm1",
		DeviceID:  "device_001",
		Level:     2,
		Message:   "High temperature",
		Value:     95.5,
		Threshold: 90.0,
	}
	service.BroadcastAlarm("station1", alarm)

	select {
	case msg := <-client.Send:
		assert.Contains(t, string(msg), "alarm")
	case <-time.After(100 * time.Millisecond):
		t.Error("client should have received alarm")
	}
}

func TestRealtimeDataService_BroadcastDeviceStatusCov(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("c1", "user1", "station1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	service := NewRealtimeDataService(hub)
	service.BroadcastDeviceStatus("station1", "device_001", "offline")

	select {
	case msg := <-client.Send:
		assert.Contains(t, string(msg), "device_status")
	case <-time.After(100 * time.Millisecond):
		t.Error("client should have received device status")
	}
}

func TestRealtimeDataService_PushDataCov(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	service := NewRealtimeDataService(hub)

	data := &StationData{
		StationID: "station1",
		Timestamp: time.Now().Unix(),
		Power:     100.0,
		Energy:    200.0,
	}

	err := service.PushData(context.Background(), data)
	require.NoError(t, err)
}

func TestRealtimeDataService_GetConnectedClientsCov(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("c1", "user1", "station1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	service := NewRealtimeDataService(hub)
	count := service.GetConnectedClients("station1")
	assert.Equal(t, 1, count)
}

func TestRealtimeDataService_StartStopMockDataStreamCov(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	service := NewRealtimeDataService(hub)

	service.StartMockDataStream("station1")
	assert.NotNil(t, service.dataChannels["station1"])

	service.StartMockDataStream("station1")

	service.StopMockDataStream("station1")
	_, exists := service.dataChannels["station1"]
	assert.False(t, exists)

	service.StopMockDataStream("nonexistent")
}

func TestStationData_StructCov(t *testing.T) {
	data := &StationData{
		StationID: "st1",
		Timestamp: time.Now().Unix(),
		Power:     850.5,
		Energy:    5000.2,
		DeviceStatus: map[string]string{
			"dev1": "online",
			"dev2": "offline",
		},
		Alarms: []AlarmInfo{
			{ID: "a1", DeviceID: "dev1", Level: 3, Message: "Critical", Value: 100.0, Threshold: 90.0},
		},
	}
	assert.Equal(t, "st1", data.StationID)
	assert.Len(t, data.Alarms, 1)
}

func TestAlarmInfo_StructCov(t *testing.T) {
	alarm := &AlarmInfo{
		ID:        "alarm1",
		DeviceID:  "dev1",
		Level:     2,
		Message:   "Warning",
		Value:     85.5,
		Threshold: 80.0,
	}
	assert.Equal(t, "alarm1", alarm.ID)
	assert.Equal(t, 2, alarm.Level)
}

func TestHub_Heartbeat(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("c1", "user1", "station1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	hub.sendHeartbeat()

	select {
	case msg := <-client.Send:
		assert.Contains(t, string(msg), "heartbeat")
	case <-time.After(100 * time.Millisecond):
		t.Error("client should have received heartbeat")
	}
}

func TestHub_SendHeartbeat_FullChannel(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("c1", "user1", "station1", hub)
	client.Send = make(chan []byte, 1)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	client.Send <- []byte("fill")
	hub.sendHeartbeat()

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, hub.GetClientCount())
}
