package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Hub  *Hub
	Send chan []byte
}

type Hub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToServer sends a message only to clients subscribed to the given serverID.
// Client.ID is the server ID they registered with via HandleWebSocket.
func (h *Hub) BroadcastToServer(serverID uint, data interface{}) {
	msg := map[string]interface{}{
		"type":      "update_status",
		"server_id": serverID,
		"data":      data,
	}
	bytes, _ := json.Marshal(msg)

	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, client := range h.clients {
		if client.ID == fmt.Sprintf("%d", serverID) {
			select {
			case client.Send <- bytes:
			default:
			}
		}
	}
}

// BroadcastToAll sends a message to every connected WebSocket client.
func (h *Hub) BroadcastToAll(data interface{}) {
	bytes, _ := json.Marshal(data)

	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, client := range h.clients {
		select {
		case client.Send <- bytes:
		default:
		}
	}
}

var (
	globalHub *Hub
)

// SetGlobalHub sets the global hub instance used for broadcasting
// server status changes across WebSocket connections.
func SetGlobalHub(h *Hub) {
	globalHub = h
}

// GetGlobalHub returns the global hub, or nil if not yet initialized.
func GetGlobalHub() *Hub {
	return globalHub
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Hub) HandleWebSocket(c *gin.Context) {
	serverID := c.Param("id")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	client := &Client{
		ID:   serverID,
		Hub:  h,
		Send: make(chan []byte, 256),
	}
	h.register <- client

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
	h.unregister <- client
}
