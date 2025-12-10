package service

import (
	"backgammon/repository"
	"backgammon/util"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, you should validate the origin properly
		// For now, allow all origins (since frontend is served from same server)
		return true
	},
}

// ChatWebSocketHandler upgrades HTTP connections to WebSocket for realtime chat
func ChatWebSocketHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// User is already authenticated by SessionMiddleware
		userID, ok := util.GetUserIDFromContext(r.Context())
		if !ok {
			util.ErrorResponse(w, http.StatusUnauthorized, "Not authenticated")
			return
		}

		// Get database instance
		db := repository.GetDB()

		// Get user info
		user, err := db.GetUserByID(r.Context(), userID)
		if err != nil {
			log.Printf("Error getting user: %v", err)
			util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get user")
			return
		}

		// Upgrade connection to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Error upgrading connection: %v", err)
			return
		}

		// Get or create lobby room
		roomID, err := db.EnsureLobbyRoomExists(r.Context())
		if err != nil {
			log.Printf("Error getting lobby room: %v", err)
			util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get lobby room")
			return
		}

		// Create client
		client := &Client{
			hub:      hub,
			conn:     conn,
			send:     make(chan []byte, sendBufferSize),
			userID:   userID,
			username: user.Username,
			roomID:   roomID,
		}

		// Register client with hub and room
		hub.register <- &ClientRegistration{client: client, roomID: roomID}

		// Send message history
		go func() {
			if err := sendMessageHistory(client, db, roomID); err != nil {
				log.Printf("Error sending message history: %v", err)
			}
		}()

		// Start client's goroutines
		go client.writePump()
		go client.readPump()
	}
}

// sendMessageHistory sends the recent chat history to a newly connected client
func sendMessageHistory(client *Client, db *repository.Postgres, roomID int) error {
	ctx := context.Background()

	// Get recent messages (last 50)
	messages, err := db.GetRecentMessages(ctx, roomID, 50)
	if err != nil {
		log.Printf("Error getting recent messages: %v", err)
		return err
	}

	// Convert to ChatMessageData format
	chatMessages := make([]ChatMessageData, len(messages))
	for i, msg := range messages {
		chatMessages[i] = ChatMessageData{
			MessageID: msg.MessageID,
			UserID:    msg.UserID,
			Username:  msg.Username,
			Message:   msg.MessageText,
			Timestamp: msg.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// Create history message
	historyData := MessageHistoryData{
		Messages: chatMessages,
	}

	historyJSON, err := json.Marshal(historyData)
	if err != nil {
		log.Printf("Error marshaling history data: %v", err)
		return err
	}

	historyMsg := WSMessage{
		Type: "history",
		Data: json.RawMessage(historyJSON),
	}

	msgBytes, err := json.Marshal(historyMsg)
	if err != nil {
		log.Printf("Error marshaling history message: %v", err)
		return err
	}

	// Send to client
	client.send <- msgBytes

	return nil
}

// handleClientMessage processes incoming messages from a client
func handleClientMessage(client *Client, message []byte) {
	var wsMsg WSMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		sendErrorToClient(client, "Invalid message format")
		return
	}

	switch wsMsg.Type {
	case "send_message":
		handleSendMessage(client, wsMsg.Data)
	default:
		log.Printf("Unknown message type: %s", wsMsg.Type)
		sendErrorToClient(client, "Unknown message type")
	}
}

// handleSendMessage processes a send_message request
func handleSendMessage(client *Client, data json.RawMessage) {
	var req SendMessageRequest
	if err := json.Unmarshal(data, &req); err != nil {
		log.Printf("Error unmarshaling send_message request: %v", err)
		sendErrorToClient(client, "Invalid message data")
		return
	}

	// Validate message
	message := strings.TrimSpace(req.Message)
	if len(message) == 0 {
		sendErrorToClient(client, "Message cannot be empty")
		return
	}
	if len(message) > 1000 {
		sendErrorToClient(client, "Message too long (max 1000 characters)")
		return
	}

	// Get database instance
	db := repository.GetDB()
	ctx := context.Background()

	// Save message to database (using client's room)
	savedMsg, err := db.SaveChatMessage(ctx, client.roomID, client.userID, message)
	if err != nil {
		log.Printf("Error saving message: %v", err)
		sendErrorToClient(client, "Failed to send message")
		return
	}

	// Create chat message to broadcast
	chatMsgData := ChatMessageData{
		MessageID: savedMsg.MessageID,
		UserID:    savedMsg.UserID,
		Username:  savedMsg.Username,
		Message:   savedMsg.MessageText,
		Timestamp: savedMsg.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}

	chatMsgJSON, err := json.Marshal(chatMsgData)
	if err != nil {
		log.Printf("Error marshaling chat message data: %v", err)
		return
	}

	broadcastMsg := WSMessage{
		Type: "chat_message",
		Data: json.RawMessage(chatMsgJSON),
	}

	msgBytes, err := json.Marshal(broadcastMsg)
	if err != nil {
		log.Printf("Error marshaling broadcast message: %v", err)
		return
	}

	// Broadcast to all clients in this room
	client.hub.broadcast <- &BroadcastMessage{
		roomID: client.roomID,
		data:   msgBytes,
	}
}

// sendErrorToClient sends an error message to a specific client
func sendErrorToClient(client *Client, errorMsg string) {
	errorData := ErrorData{
		Message: errorMsg,
	}

	errorJSON, err := json.Marshal(errorData)
	if err != nil {
		log.Printf("Error marshaling error data: %v", err)
		return
	}

	wsMsg := WSMessage{
		Type: "error",
		Data: json.RawMessage(errorJSON),
	}

	msgBytes, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Error marshaling error message: %v", err)
		return
	}

	select {
	case client.send <- msgBytes:
	default:
		log.Printf("Failed to send error message to client %d, channel full", client.userID)
	}
}

// GameChatWebSocketHandler upgrades HTTP connections to WebSocket for game-specific chat
func GameChatWebSocketHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// User is already authenticated by SessionMiddleware
		userID, ok := util.GetUserIDFromContext(r.Context())
		if !ok {
			util.ErrorResponse(w, http.StatusUnauthorized, "Not authenticated")
			return
		}

		// Extract gameID from URL path (/api/v1/games/{gameId}/ws)
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 4 {
			util.ErrorResponse(w, http.StatusBadRequest, "Invalid game ID")
			return
		}

		gameID, err := strconv.Atoi(pathParts[3])
		if err != nil {
			util.ErrorResponse(w, http.StatusBadRequest, "Invalid game ID")
			return
		}

		// Get database instance
		db := repository.GetDB()

		// Get user info
		user, err := db.GetUserByID(r.Context(), userID)
		if err != nil {
			log.Printf("Error getting user: %v", err)
			util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get user")
			return
		}

		// Get game and verify user is a player
		game, err := db.GetGameByID(r.Context(), gameID)
		if err != nil {
			log.Printf("Error getting game: %v", err)
			util.ErrorResponse(w, http.StatusNotFound, "Game not found")
			return
		}

		// Verify user is a player in this game
		if game.Player1ID != userID && game.Player2ID != userID {
			util.ErrorResponse(w, http.StatusForbidden, "Not a player in this game")
			return
		}

		// Get or create game chat room
		roomID, err := db.GetOrCreateGameChatRoom(r.Context(), gameID)
		if err != nil {
			log.Printf("Error getting game chat room: %v", err)
			util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get game chat room")
			return
		}

		// Upgrade connection to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Error upgrading connection: %v", err)
			return
		}

		// Create client
		client := &Client{
			hub:      hub,
			conn:     conn,
			send:     make(chan []byte, sendBufferSize),
			userID:   userID,
			username: user.Username,
			roomID:   roomID,
		}

		// Register client with hub and room
		hub.register <- &ClientRegistration{client: client, roomID: roomID}

		// Send message history
		go func() {
			if err := sendMessageHistory(client, db, roomID); err != nil {
				log.Printf("Error sending message history: %v", err)
			}
		}()

		// Start client's goroutines
		go client.writePump()
		go client.readPump()
	}
}

// ChatMessagesHandler is a placeholder for potential REST API endpoint (currently unused)
func ChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	// This endpoint is not currently used
	// All chat functionality is handled via WebSockets
	util.ErrorResponse(w, http.StatusNotImplemented, "Use WebSocket endpoint at /api/v1/lobby/ws")
}
