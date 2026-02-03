package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ims-erp/system/internal/events"
)

// Hub manages WebSocket connections and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// Client represents a WebSocket client
type Client struct {
	id            string
	hub           *Hub
	conn          *websocket.Conn
	send          chan []byte
	tenantID      string
	subscriptions map[string]bool
}

// Message represents a WebSocket message
type Message struct {
	Type     string          `json:"type"`
	Payload  json.RawMessage `json:"payload"`
	TenantID string          `json:"tenantId"`
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub event loop
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if client.tenantID == message.TenantID || message.TenantID == "" {
					select {
					case client.send <- mustMarshal(message):
					default:
						// Client send buffer full, close connection
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()

		case <-ctx.Done():
			return
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(msg *Message) {
	h.broadcast <- msg
}

// Upgrader configures the WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, validate origin
	},
}

// ServeWs handles WebSocket requests
func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Extract tenant ID from request context or headers
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default"
	}

	client := &Client{
		hub:           h,
		conn:          conn,
		send:          make(chan []byte, 256),
		tenantID:      tenantID,
		subscriptions: make(map[string]bool),
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024) // 512KB
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log error
			}
			break
		}

		// Parse subscription message
		var sub SubscriptionMessage
		if err := json.Unmarshal(message, &sub); err == nil {
			c.handleSubscription(sub)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SubscriptionMessage represents a subscription request
type SubscriptionMessage struct {
	Action    string                 `json:"action"`    // subscribe, unsubscribe
	EventType string                 `json:"eventType"` // client.created, invoice.paid, etc.
	Filters   map[string]interface{} `json:"filters"`
}

// handleSubscription processes subscription requests
func (c *Client) handleSubscription(sub SubscriptionMessage) {
	switch sub.Action {
	case "subscribe":
		c.subscriptions[sub.EventType] = true
	case "unsubscribe":
		delete(c.subscriptions, sub.EventType)
	}
}

// EventBridge connects NATS events to WebSocket broadcasts
type EventBridge struct {
	hub    *Hub
	events <-chan events.EventEnvelope
}

// NewEventBridge creates a new event bridge
func NewEventBridge(hub *Hub) *EventBridge {
	return &EventBridge{
		hub: hub,
	}
}

// Start begins bridging events to WebSocket
func (eb *EventBridge) Start(ctx context.Context) {
	for {
		select {
		case event := <-eb.events:
			eb.handleEvent(event)
		case <-ctx.Done():
			return
		}
	}
}

// handleEvent converts a NATS event to WebSocket message
func (eb *EventBridge) handleEvent(event events.EventEnvelope) {
	msg := &Message{
		Type:     event.Type,
		Payload:  mustMarshal(event.Data),
		TenantID: event.TenantID,
	}

	eb.hub.Broadcast(msg)
}

func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
