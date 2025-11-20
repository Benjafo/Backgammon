package service

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"backgammon/repository"
	"backgammon/util"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 1000
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		// TODO: Restrict this in production
		return true
	},
}

// Client represents a single websocket connection
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   int
	username string
	roomID   int
}

// Hub maintains the set of active clients and broadcasts messages to clients
type Hub struct {
	// Registered clients per room
	rooms map[int]map[*Client]bool

	// Inbound messages from clients
	broadcast chan *BroadcastMessage

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// BroadcastMessage represents a message to be broadcast to a room
type BroadcastMessage struct {
	RoomID  int
	Message []byte
}

// Message types for websocket communication
type IncomingMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type OutgoingMessage struct {
	Type      string    `json:"type"`
	MessageID int       `json:"messageId"`
	UserID    int       `json:"userId"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type HistoryMessage struct {
	Type     string            `json:"type"`
	Messages []OutgoingMessage `json:"messages"`
}

// Global hub instance
var chatHub *Hub

// InitChatHub initializes the global chat hub
func InitChatHub() *Hub {
	chatHub = &Hub{
		broadcast:  make(chan *BroadcastMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rooms:      make(map[int]map[*Client]bool),
	}
	go chatHub.run()
	return chatHub
}

// GetChatHub returns the global chat hub instance
func GetChatHub() *Hub {
	return chatHub
}

// run handles hub operations
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.roomID] == nil {
				h.rooms[client.roomID] = make(map[*Client]bool)
			}
			h.rooms[client.roomID][client] = true
			h.mu.Unlock()
			log.Printf("Client registered: user=%s, room=%d, total_clients=%d",
				client.username, client.roomID, len(h.rooms[client.roomID]))

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.roomID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.send)
					log.Printf("Client unregistered: user=%s, room=%d, remaining=%d",
						client.username, client.roomID, len(clients))
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.rooms[message.RoomID]
			h.mu.RUnlock()

			for client := range clients {
				select {
				case client.send <- message.Message:
				default:
					// Client's send buffer is full, close connection
					h.mu.Lock()
					close(client.send)
					delete(h.rooms[message.RoomID], client)
					h.mu.Unlock()
				}
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
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
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Websocket error: %v", err)
			}
			break
		}

		// Parse incoming message
		var inMsg IncomingMessage
		if err := json.Unmarshal(messageBytes, &inMsg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		// Handle different message types
		switch inMsg.Type {
		case "chat":
			c.handleChatMessage(inMsg.Message)
		default:
			log.Printf("Unknown message type: %s", inMsg.Type)
		}
	}
}

// handleChatMessage processes incoming chat messages
func (c *Client) handleChatMessage(messageText string) {
	db := repository.GetDB()
	if db == nil {
		log.Printf("Database not initialized")
		return
	}

	// Save message to database
	ctx := context.Background()
	savedMsg, err := db.SaveMessage(ctx, c.roomID, c.userID, messageText)
	if err != nil {
		log.Printf("Failed to save message: %v", err)
		return
	}

	// Create outgoing message
	outMsg := OutgoingMessage{
		Type:      "message",
		MessageID: savedMsg.MessageID,
		UserID:    c.userID,
		Username:  c.username,
		Message:   messageText,
		Timestamp: savedMsg.Timestamp,
	}

	// Broadcast to all clients in the room
	msgBytes, err := json.Marshal(outMsg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	c.hub.broadcast <- &BroadcastMessage{
		RoomID:  c.roomID,
		Message: msgBytes,
	}
}

// writePump pumps messages from the hub to the websocket connection
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
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
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

// ChatWebSocketHandler handles websocket connections for chat
func ChatWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by session middleware)
	userID, ok := util.GetUserIDFromContext(r.Context())
	if !ok {
		util.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	db := repository.GetDB()
	if db == nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	// Get user details
	user, err := db.GetUserByID(r.Context(), userID)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	// Get or create lobby chat room
	roomID, err := db.GetOrCreateLobbyChatRoom(r.Context())
	if err != nil {
		log.Printf("Failed to get lobby chat room: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get chat room")
		return
	}

	// Upgrade connection to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Create client
	client := &Client{
		hub:      chatHub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		username: user.Username,
		roomID:   roomID,
	}

	// Register client with hub
	client.hub.register <- client

	// Send message history to the client
	messages, err := db.GetRecentMessages(r.Context(), roomID, 50)
	if err != nil {
		log.Printf("Failed to get message history: %v", err)
	} else {
		var outMessages []OutgoingMessage
		for _, msg := range messages {
			outMessages = append(outMessages, OutgoingMessage{
				Type:      "message",
				MessageID: msg.MessageID,
				UserID:    msg.UserID,
				Username:  msg.Username,
				Message:   msg.MessageText,
				Timestamp: msg.Timestamp,
			})
		}

		historyMsg := HistoryMessage{
			Type:     "history",
			Messages: outMessages,
		}

		historyBytes, err := json.Marshal(historyMsg)
		if err == nil {
			client.send <- historyBytes
		}
	}

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}
