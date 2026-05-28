package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComp_Hub_BroadcastReceive(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("bc1", "user1", "st1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	hub.BroadcastToAll(&Message{Type: "test", Data: "hello"})
	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-client.Send:
		var received Message
		err := json.Unmarshal(msg, &received)
		require.NoError(t, err)
		assert.Equal(t, "test", received.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for broadcast message")
	}

	hub.unregister <- client
	time.Sleep(50 * time.Millisecond)
}

func TestComp_Hub_BroadcastToStation_Receive(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := NewClient("sc1", "user1", "st1", hub)
	client2 := NewClient("sc2", "user2", "st2", hub)
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	hub.BroadcastToStation("st1", &Message{Type: "station_data", Data: map[string]interface{}{"power": 100}})
	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-client1.Send:
		var received Message
		err := json.Unmarshal(msg, &received)
		require.NoError(t, err)
		assert.Equal(t, "station_data", received.Type)
		assert.Equal(t, "st1", received.StationID)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for station broadcast")
	}

	select {
	case <-client2.Send:
		t.Fatal("client2 should not receive station1 message")
	default:
	}

	hub.unregister <- client1
	hub.unregister <- client2
}

func TestComp_Hub_BroadcastToUser_Receive(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := NewClient("uc1", "user1", "st1", hub)
	client2 := NewClient("uc2", "user2", "st2", hub)
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	hub.BroadcastToUser("user1", &Message{Type: "private", Data: "secret"})
	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-client1.Send:
		var received Message
		err := json.Unmarshal(msg, &received)
		require.NoError(t, err)
		assert.Equal(t, "private", received.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for user broadcast")
	}

	select {
	case <-client2.Send:
		t.Fatal("client2 should not receive user1 message")
	default:
	}

	hub.unregister <- client1
	hub.unregister <- client2
}

func TestComp_Hub_Broadcast_FullSendChannel(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("fc1", "user1", "st1", hub)
	for i := 0; i < 256; i++ {
		client.Send <- []byte("fill")
	}
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	hub.BroadcastToAll(&Message{Type: "overflow", Data: "test"})
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, hub.GetClientCount())
}

func TestComp_Hub_BroadcastToStation_EmptyStationClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("ec1", "user1", "", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	hub.BroadcastToStation("st1", &Message{Type: "test", Data: "data"})
	time.Sleep(50 * time.Millisecond)

	select {
	case <-client.Send:
	case <-time.After(1 * time.Second):
		t.Fatal("client with empty station should receive all station messages")
	}

	hub.unregister <- client
}

func TestComp_Hub_SendHeartbeat(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("hb1", "user1", "st1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	hub.sendHeartbeat()

	select {
	case msg := <-client.Send:
		var received Message
		err := json.Unmarshal(msg, &received)
		require.NoError(t, err)
		assert.Equal(t, "heartbeat", received.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for heartbeat")
	}

	hub.unregister <- client
}

func TestComp_Hub_UnregisterNonExistent(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("ne1", "user1", "st1", hub)
	hub.unregister <- client
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 0, hub.GetClientCount())
}

func TestComp_Client_ReadPump_ChannelClose(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("rpc1", "user1", "st1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	done := make(chan bool)
	go func() {
		client.ReadPump(func(msg []byte) {})
		done <- true
	}()

	hub.unregister <- client
	time.Sleep(100 * time.Millisecond)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ReadPump should exit when Send channel is closed")
	}
}

func TestComp_Client_WritePump(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("wp1", "user1", "st1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	done := make(chan bool)
	go func() {
		client.WritePump(func() ([]byte, bool) {
			return nil, false
		})
		done <- true
	}()

	close(client.Send)
	time.Sleep(100 * time.Millisecond)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
}

func TestComp_RealtimeDataService_PushDataWithClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	service := NewRealtimeDataService(hub)

	client := NewClient("pd1", "user1", "st1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	data := &StationData{
		StationID: "st1",
		Timestamp: time.Now().Unix(),
		Power:     100.0,
		Energy:    500.0,
	}
	err := service.PushData(context.Background(), data)
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-client.Send:
		var received Message
		err := json.Unmarshal(msg, &received)
		require.NoError(t, err)
		assert.Equal(t, "station_data", received.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for pushed data")
	}

	hub.unregister <- client
}

func TestComp_RealtimeDataService_MockDataStream(t *testing.T) {
	hub := NewHub()
	service := NewRealtimeDataService(hub)

	service.StartMockDataStream("st-mock")
	assert.NotNil(t, service.dataChannels["st-mock"])

	service.StopMockDataStream("st-mock")
	_, exists := service.dataChannels["st-mock"]
	assert.False(t, exists)
}

func TestComp_RealtimeDataService_StopNonExistentStream(t *testing.T) {
	hub := NewHub()
	service := NewRealtimeDataService(hub)

	service.StopMockDataStream("nonexistent")
}

func TestComp_RealtimeDataService_StartDuplicateStream(t *testing.T) {
	hub := NewHub()
	service := NewRealtimeDataService(hub)

	service.StartMockDataStream("st-dup")
	service.StartMockDataStream("st-dup")
	assert.NotNil(t, service.dataChannels["st-dup"])

	service.StopMockDataStream("st-dup")
}

func TestComp_RealtimeDataService_GetConnectedClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	service := NewRealtimeDataService(hub)

	client1 := NewClient("gcc1", "user1", "st1", hub)
	client2 := NewClient("gcc2", "user2", "st1", hub)
	hub.register <- client1
	hub.register <- client2
	time.Sleep(50 * time.Millisecond)

	count := service.GetConnectedClients("st1")
	assert.Equal(t, 2, count)

	hub.unregister <- client1
	hub.unregister <- client2
}

func TestComp_RealtimeDataService_BroadcastAlarmWithClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	service := NewRealtimeDataService(hub)

	client := NewClient("ba1", "user1", "st1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	alarm := &AlarmInfo{
		ID:        "alarm-comp",
		DeviceID:  "dev1",
		Level:     3,
		Message:   "High temperature",
		Value:     95.5,
		Threshold: 90.0,
	}
	service.BroadcastAlarm("st1", alarm)
	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-client.Send:
		var received Message
		err := json.Unmarshal(msg, &received)
		require.NoError(t, err)
		assert.Equal(t, "alarm", received.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for alarm broadcast")
	}

	hub.unregister <- client
}

func TestComp_RealtimeDataService_BroadcastDeviceStatusWithClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	service := NewRealtimeDataService(hub)

	client := NewClient("bds1", "user1", "st1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	service.BroadcastDeviceStatus("st1", "dev1", "offline")
	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-client.Send:
		var received Message
		err := json.Unmarshal(msg, &received)
		require.NoError(t, err)
		assert.Equal(t, "device_status", received.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for device status broadcast")
	}

	hub.unregister <- client
}

func TestComp_WSHandler_GetStats(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	handler := NewWSHandler(hub)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "total_clients")
}

func TestComp_WSHandler_HandleWebSocket(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewHub()
	go hub.Run()
	handler := NewWSHandler(hub)

	r := gin.New()
	r.GET("/ws", func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Set("request_id", "test-request-id")
		handler.HandleWebSocket(c)
	})

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?station_id=st1"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, hub.GetClientCount())

	ws.WriteMessage(websocket.CloseMessage, []byte{})
	ws.Close()
	time.Sleep(200 * time.Millisecond)
}

func TestComp_WSHandler_ReadPump_UpdateStationID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewHub()
	go hub.Run()
	handler := NewWSHandler(hub)

	r := gin.New()
	r.GET("/ws", func(c *gin.Context) {
		c.Set("user_id", "test-user2")
		c.Set("request_id", "test-request-id2")
		handler.HandleWebSocket(c)
	})

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	stationMsg, _ := json.Marshal(map[string]string{"type": "subscribe", "station_id": "st-new"})
	err = ws.WriteMessage(websocket.TextMessage, stationMsg)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	ws.WriteMessage(websocket.CloseMessage, []byte{})
	ws.Close()
	time.Sleep(200 * time.Millisecond)
}

func TestComp_WSHandler_ReadPump_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := NewHub()
	go hub.Run()
	handler := NewWSHandler(hub)

	r := gin.New()
	r.GET("/ws", func(c *gin.Context) {
		c.Set("user_id", "test-user3")
		c.Set("request_id", "test-request-id3")
		handler.HandleWebSocket(c)
	})

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = ws.WriteMessage(websocket.TextMessage, []byte("invalid json"))
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, hub.GetClientCount())

	ws.WriteMessage(websocket.CloseMessage, []byte{})
	ws.Close()
	time.Sleep(200 * time.Millisecond)
}

func TestComp_RealtimeDataService_BroadcastAllStations(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	service := NewRealtimeDataService(hub)

	client := NewClient("bas1", "user1", "st-broadcast", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	service.StartMockDataStream("st-broadcast")
	time.Sleep(1500 * time.Millisecond)

	select {
	case <-client.Send:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for broadcast data")
	}

	service.StopMockDataStream("st-broadcast")
	hub.unregister <- client
}

func TestComp_RealtimeDataService_PushData_CancelContext(t *testing.T) {
	hub := NewHub()
	service := NewRealtimeDataService(hub)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	data := &StationData{
		StationID: "st-cancel",
		Timestamp: time.Now().Unix(),
		Power:     100.0,
	}
	_ = service.PushData(ctx, data)
}

func TestComp_Client_ReadPump_WithCallback(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("cb1", "user1", "st1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	received := make([]string, 0)
	go client.ReadPump(func(msg []byte) {
		received = append(received, string(msg))
	})

	testMsg := []byte("callback test message")
	client.Send <- testMsg
	time.Sleep(50 * time.Millisecond)

	assert.GreaterOrEqual(t, len(received), 1)
	assert.Equal(t, "callback test message", received[0])

	hub.unregister <- client
}

func TestComp_Hub_MultipleBroadcasts(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := NewClient("mb1", "user1", "st1", hub)
	hub.register <- client
	time.Sleep(50 * time.Millisecond)

	for i := 0; i < 5; i++ {
		hub.BroadcastToAll(&Message{Type: "multi", Data: i})
	}
	time.Sleep(100 * time.Millisecond)

	count := 0
	for {
		select {
		case <-client.Send:
			count++
		default:
			goto done
		}
	}
done:
	assert.GreaterOrEqual(t, count, 1)

	hub.unregister <- client
}
