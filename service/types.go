package service

import "encoding/json"

// ============================================================================
// Auth & User Types
// ============================================================================

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"` // Registration CSRF token
}

type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// ============================================================================
// Game Types
// ============================================================================

type MoveRequest struct {
	FromPoint      int   `json:"fromPoint"`
	ToPoint        int   `json:"toPoint"`
	DieUsed        int   `json:"dieUsed"`
	DiceIndices    []int `json:"diceIndices"`    // Indices of dice being used (for combined moves)
	IsCombinedMove bool  `json:"isCombinedMove"` // True if using multiple dice
}

// ============================================================================
// Invitation Types
// ============================================================================

type CreateInvitationRequest struct {
	ChallengedID int `json:"challengedId"`
}

// ============================================================================
// WebSocket & Chat Types
// ============================================================================

type WSMessage struct {
	Type string          `json:"type"` // "send_message", "chat_message", "history", "user_joined", "user_left", "error"
	Data json.RawMessage `json:"data"`
}

type SendMessageRequest struct {
	Message string `json:"message"`
}

type ChatMessageData struct {
	MessageID int    `json:"messageId"`
	UserID    int    `json:"userId"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"` // ISO 8601 format
}

type MessageHistoryData struct {
	Messages []ChatMessageData `json:"messages"`
}

type UserEventData struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
}

type ErrorData struct {
	Message string `json:"message"`
}
