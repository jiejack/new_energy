package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHandler struct {
	hub *Hub
}

func NewWSHandler(hub *Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	userID := c.GetString("user_id")
	stationID := c.Query("station_id")
	clientID := c.GetString("request_id")

	client := NewClient(clientID, userID, stationID, h.hub)
	h.hub.register <- client

	go h.writePump(conn, client)
	go h.readPump(conn, client)
}

func (h *WSHandler) writePump(conn *websocket.Conn, client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WSHandler) readPump(conn *websocket.Conn, client *Client) {
	defer func() {
		h.hub.unregister <- client
		conn.Close()
	}()

	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		var msg struct {
			Type      string `json:"type"`
			StationID string `json:"station_id,omitempty"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		if msg.StationID != "" {
			client.StationID = msg.StationID
		}
	}
}

func (h *WSHandler) GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_clients": h.hub.GetClientCount(),
	})
}
