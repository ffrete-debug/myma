package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"ark-server-commander/database"
	"ark-server-commander/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Timings for the update WebSocket pumps. hubPingPeriod must stay below
// hubPongWait so a responsive peer always refreshes its read deadline in time.
const (
	hubWriteTimeout   = 10 * time.Second
	hubPongWait       = 60 * time.Second
	hubPingPeriod     = (hubPongWait * 9) / 10
	hubMaxMessageSize = 512
)

type Client struct {
	ServerID uint
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
}

type Hub struct {
	// clients is keyed by server ID and then by connection: several viewers may
	// watch the same server, so a set of clients is kept per server instead of a
	// single slot that later connections would evict.
	clients    map[uint]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uint]map[*Client]bool),
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
			if h.clients[client.ServerID] == nil {
				h.clients[client.ServerID] = make(map[*Client]bool)
			}
			h.clients[client.ServerID][client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if peers, ok := h.clients[client.ServerID]; ok {
				if _, ok := peers[client]; ok {
					delete(peers, client)
					close(client.Send)
					if len(peers) == 0 {
						delete(h.clients, client.ServerID)
					}
				}
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			for _, peers := range h.clients {
				for client := range peers {
					select {
					case client.Send <- message:
					default:
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToServer sends a message only to clients subscribed to the given serverID.
// Client.ServerID is the server they registered with via HandleWebSocket.
func (h *Hub) BroadcastToServer(serverID uint, data interface{}) {
	msg := map[string]interface{}{
		"type":      "update_status",
		"server_id": serverID,
		"data":      data,
	}
	bytes, _ := json.Marshal(msg)

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients[serverID] {
		select {
		case client.Send <- bytes:
		default:
		}
	}
}

// BroadcastToAll sends a message to every connected WebSocket client.
func (h *Hub) BroadcastToAll(data interface{}) {
	bytes, _ := json.Marshal(data)

	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, peers := range h.clients {
		for client := range peers {
			select {
			case client.Send <- bytes:
			default:
			}
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
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	id, err := strconv.ParseUint(serverID, 10, 32)
	if err != nil {
		safeWriteText(conn, "error: invalid server id")
		conn.Close()
		return
	}

	// Ownership check: the update stream must only be readable by the user the
	// server belongs to, same as the logs and RCON handlers.
	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		safeWriteText(conn, "error: server not found")
		conn.Close()
		return
	}

	client := &Client{
		ServerID: server.ID,
		Hub:      h,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}
	h.register <- client

	// The write pump drains Send (without it nothing ever reaches the browser),
	// the read pump blocks here until the peer goes away.
	go client.writePump()
	client.readPump()
}

// readPump keeps the read side alive so control frames are processed and a
// disconnect is noticed promptly. The update stream is one-way, so inbound
// payloads are discarded. On exit the client is unregistered and the connection
// closed, which also stops the write pump.
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(hubMaxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(hubPongWait))
	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(hubPongWait))
	})

	for {
		if _, _, err := c.Conn.ReadMessage(); err != nil {
			return
		}
	}
}

// writePump is the only goroutine writing to the connection: it forwards queued
// broadcasts and sends periodic pings so dead peers are detected.
func (c *Client) writePump() {
	ticker := time.NewTicker(hubPingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(hubWriteTimeout))
			if !ok {
				// The hub closed the channel — say goodbye and stop.
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(hubWriteTimeout))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
