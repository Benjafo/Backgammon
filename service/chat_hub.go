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
	roomID   int // Which chat room this client is in
}

// ClientRegistration wraps a client with its room information for registration
type ClientRegistration struct {
	client *Client
	roomID int
}

// BroadcastMessage represents a message to be broadcast to clients in a room
type BroadcastMessage struct {
	roomID int    // Which room to broadcast to
	data   []byte
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients (map[userID][]*Client to support multiple connections per user)
	clients map[int][]*Client

	// Rooms (map[roomID]map[*Client]bool for fast lookup)
	rooms map[int]map[*Client]bool

	// Inbound messages from the clients
	broadcast chan *BroadcastMessage

	// Register requests from the clients
	register chan *ClientRegistration

	// Unregister requests from clients
	unregister chan *ClientRegistration

	// Mutex for thread-safe access to clients map
	mu sync.RWMutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *BroadcastMessage, 256),
		register:   make(chan *ClientRegistration),
		unregister: make(chan *ClientRegistration),
		clients:    make(map[int][]*Client),
		rooms:      make(map[int]map[*Client]bool),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case reg := <-h.register:
			h.registerClient(reg.client, reg.roomID)

		case reg := <-h.unregister:
			h.unregisterClient(reg.client, reg.roomID)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a client to the hub and room, and broadcasts user_joined
func (h *Hub) registerClient(client *Client, roomID int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add client to the user's connection list
	h.clients[client.userID] = append(h.clients[client.userID], client)

	// Add client to the room
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true

	log.Printf("User %s (ID: %d) connected to room %d. Total clients for this user: %d",
		client.username, client.userID, roomID, len(h.clients[client.userID]))

	// Broadcast user_joined notification to other clients in the same room
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

	// Broadcast to clients in the same room
	h.mu.Unlock()
	h.broadcast <- &BroadcastMessage{
		roomID: roomID,
		data:   msgBytes,
	}
	h.mu.Lock()
}

// unregisterClient removes a client from the hub and room, and broadcasts user_left
func (h *Hub) unregisterClient(client *Client, roomID int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove client from room
	if h.rooms[roomID] != nil {
		delete(h.rooms[roomID], client)
		if len(h.rooms[roomID]) == 0 {
			delete(h.rooms, roomID)
		}
	}

	// Find and remove the specific client from the user's connection list
	connections := h.clients[client.userID]
	for i, c := range connections {
		if c == client {
			// Remove this specific connection
			h.clients[client.userID] = append(connections[:i], connections[i+1:]...)
			break
		}
	}

	// Check if user has any other connections in this room
	hasOtherConnectionsInRoom := false
	for _, conn := range h.clients[client.userID] {
		if conn.roomID == roomID {
			hasOtherConnectionsInRoom = true
			break
		}
	}

	// Broadcast user_left only if user has no other connections in this room
	if !hasOtherConnectionsInRoom {
		log.Printf("User %s (ID: %d) disconnected from room %d", client.username, client.userID, roomID)

		// Broadcast user_left notification to this room
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
		h.broadcast <- &BroadcastMessage{
			roomID: roomID,
			data:   msgBytes,
		}
		h.mu.Lock()
	}

	// If no more connections for this user at all, remove the user entry
	if len(h.clients[client.userID]) == 0 {
		delete(h.clients, client.userID)
		log.Printf("User %s (ID: %d) all connections closed", client.username, client.userID)
	}

	// Close the client's send channel
	close(client.send)
}

// broadcastMessage sends a message to all clients in the specified room
func (h *Hub) broadcastMessage(message *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Get clients in the specified room
	roomClients := h.rooms[message.roomID]
	if roomClients == nil {
		return // Room doesn't exist or has no clients
	}

	// Send message to all clients in the room
	for client := range roomClients {
		select {
		case client.send <- message.data:
		default:
			// Channel is full or closed, skip this client
			log.Printf("Failed to send message to user %d in room %d, send channel full or closed",
				client.userID, message.roomID)
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- &ClientRegistration{client: c, roomID: c.roomID}
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
