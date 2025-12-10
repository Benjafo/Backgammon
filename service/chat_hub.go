package service

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = 50 * time.Second

	// Maximum message size allowed from peer
	maxMessageSize = 1024

	// Size of the send channel buffer
	sendBufferSize = 256
)

// Client represents a single WebSocket connection
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   int
	username string
}

// BroadcastMessage represents a message to be broadcast to clients
type BroadcastMessage struct {
	data    []byte
	exclude *Client // Optional: exclude this client from broadcast
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients (map[userID][]*Client to support multiple connections per user)
	clients map[int][]*Client

	// Inbound messages from the clients
	broadcast chan *BroadcastMessage

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe access to clients map
	mu sync.RWMutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *BroadcastMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[int][]*Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a client to the hub and broadcasts user_joined
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add client to the user's connection list
	h.clients[client.userID] = append(h.clients[client.userID], client)

	log.Printf("User %s (ID: %d) connected. Total clients for this user: %d",
		client.username, client.userID, len(h.clients[client.userID]))

	// Broadcast user_joined notification to all other clients
	userData := UserEventData{
		UserID:   client.userID,
		Username: client.username,
	}

	userDataJSON, err := json.Marshal(userData)
	if err != nil {
		log.Printf("Error marshaling user data: %v", err)
		return
	}

	joinMsg := WSMessage{
		Type: "user_joined",
		Data: json.RawMessage(userDataJSON),
	}

	msgBytes, err := json.Marshal(joinMsg)
	if err != nil {
		log.Printf("Error marshaling user_joined message: %v", err)
		return
	}

	// Broadcast to all clients except the newly connected one
	h.mu.Unlock()
	h.broadcast <- &BroadcastMessage{
		data:    msgBytes,
		exclude: client,
	}
	h.mu.Lock()
}

// unregisterClient removes a client from the hub and broadcasts user_left
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Find and remove the specific client from the user's connection list
	connections := h.clients[client.userID]
	for i, c := range connections {
		if c == client {
			// Remove this specific connection
			h.clients[client.userID] = append(connections[:i], connections[i+1:]...)
			break
		}
	}

	// If no more connections for this user, remove the user entry and broadcast user_left
	if len(h.clients[client.userID]) == 0 {
		delete(h.clients, client.userID)

		log.Printf("User %s (ID: %d) disconnected. All connections closed.", client.username, client.userID)

		// Broadcast user_left notification
		userData := UserEventData{
			UserID:   client.userID,
			Username: client.username,
		}

		userDataJSON, err := json.Marshal(userData)
		if err != nil {
			log.Printf("Error marshaling user data: %v", err)
			return
		}

		leaveMsg := WSMessage{
			Type: "user_left",
			Data: json.RawMessage(userDataJSON),
		}

		msgBytes, err := json.Marshal(leaveMsg)
		if err != nil {
			log.Printf("Error marshaling user_left message: %v", err)
			return
		}

		h.mu.Unlock()
		h.broadcast <- &BroadcastMessage{data: msgBytes}
		h.mu.Lock()
	} else {
		log.Printf("User %s (ID: %d) connection closed. Remaining connections: %d",
			client.username, client.userID, len(h.clients[client.userID]))
	}

	// Close the client's send channel
	close(client.send)
}

// broadcastMessage sends a message to all connected clients
func (h *Hub) broadcastMessage(message *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, connections := range h.clients {
		for _, client := range connections {
			// Skip if this is the excluded client
			if message.exclude != nil && client == message.exclude {
				continue
			}

			select {
			case client.send <- message.data:
			default:
				// Channel is full or closed, skip this client
				log.Printf("Failed to send message to user %d, send channel full or closed", client.userID)
			}
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	c.conn.SetReadLimit(maxMessageSize)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %d: %v", c.userID, err)
			}
			break
		}
		handleClientMessage(c, message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
