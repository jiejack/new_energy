package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

type Client struct {
	ID       string
	UserID   string
	StationID string
	Send     chan []byte
	Hub      *Hub
}

type Message struct {
	Type      string      `json:"type"`
	StationID string      `json:"station_id,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected: %s, total: %d", client.ID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Client disconnected: %s, total: %d", client.ID, len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			h.sendHeartbeat()
		}
	}
}

func (h *Hub) sendHeartbeat() {
	msg := Message{
		Type:      "heartbeat",
		Timestamp: time.Now().Unix(),
	}
	data, _ := json.Marshal(msg)
	h.mu.RLock()
	for client := range h.clients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(h.clients, client)
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) BroadcastToAll(msg *Message) {
	msg.Timestamp = time.Now().Unix()
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}
	h.broadcast <- data
}

func (h *Hub) BroadcastToStation(stationID string, msg *Message) {
	msg.StationID = stationID
	msg.Timestamp = time.Now().Unix()
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	h.mu.RLock()
	for client := range h.clients {
		if client.StationID == "" || client.StationID == stationID {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) BroadcastToUser(userID string, msg *Message) {
	msg.Timestamp = time.Now().Unix()
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	h.mu.RLock()
	for client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) GetStationClientCount(stationID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for client := range h.clients {
		if client.StationID == stationID {
			count++
		}
	}
	return count
}

func NewClient(id, userID, stationID string, hub *Hub) *Client {
	return &Client{
		ID:        id,
		UserID:    userID,
		StationID: stationID,
		Send:      make(chan []byte, 256),
		Hub:       hub,
	}
}

func (c *Client) ReadPump(onMessage func([]byte)) {
	defer func() {
		c.Hub.unregister <- c
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				return
			}
			if onMessage != nil {
				onMessage(message)
			}
		}
	}
}

func (c *Client) WritePump(send func() ([]byte, bool)) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				return
			}
			_ = message
		case <-ticker.C:
		}
	}
}
