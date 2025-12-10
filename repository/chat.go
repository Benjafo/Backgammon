package repository

import (
	"context"
	"fmt"
	"time"
)

// ChatMessage represents a message in the chat
type ChatMessage struct {
	MessageID   int       `json:"messageId"`
	RoomID      int       `json:"roomId"`
	UserID      int       `json:"userId"`
	Username    string    `json:"username"`
	MessageText string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// GetLobbyRoomID retrieves the lobby chat room ID
// Returns the room_id for the lobby room type
func (pg *Postgres) GetLobbyRoomID(ctx context.Context) (int, error) {
	query := `
		SELECT room_id
		FROM CHAT_ROOM
		WHERE room_type = 'lobby'
		LIMIT 1
	`

	var roomID int
	err := pg.db.QueryRow(ctx, query).Scan(&roomID)
	if err != nil {
		return 0, fmt.Errorf("failed to get lobby room ID: %w", err)
	}

	return roomID, nil
}

// EnsureLobbyRoomExists creates the lobby chat room if it doesn't exist
func (pg *Postgres) EnsureLobbyRoomExists(ctx context.Context) (int, error) {
	// First, try to get the existing lobby room
	roomID, err := pg.GetLobbyRoomID(ctx)
	if err == nil {
		return roomID, nil
	}

	// If not found, create it
	query := `
		INSERT INTO CHAT_ROOM (room_type, game_id)
		VALUES ('lobby', NULL)
		RETURNING room_id
	`

	err = pg.db.QueryRow(ctx, query).Scan(&roomID)
	if err != nil {
		return 0, fmt.Errorf("failed to create lobby room: %w", err)
	}

	return roomID, nil
}

// GetOrCreateGameChatRoom gets or creates a chat room for a specific game
func (pg *Postgres) GetOrCreateGameChatRoom(ctx context.Context, gameID int) (int, error) {
	// Try to get existing room
	query := `SELECT room_id FROM CHAT_ROOM WHERE room_type = 'game' AND game_id = $1`
	var roomID int
	err := pg.db.QueryRow(ctx, query, gameID).Scan(&roomID)
	if err == nil {
		return roomID, nil
	}

	// Create new room if not exists
	insertQuery := `INSERT INTO CHAT_ROOM (room_type, game_id) VALUES ('game', $1) RETURNING room_id`
	err = pg.db.QueryRow(ctx, insertQuery, gameID).Scan(&roomID)
	if err != nil {
		return 0, fmt.Errorf("failed to create game chat room: %w", err)
	}

	return roomID, nil
}

// SaveChatMessage saves a chat message to the database
func (pg *Postgres) SaveChatMessage(ctx context.Context, roomID, userID int, message string) (*ChatMessage, error) {
	query := `
		INSERT INTO CHAT_MESSAGE (room_id, user_id, message_text)
		VALUES ($1, $2, $3)
		RETURNING message_id, timestamp
	`

	var messageID int
	var timestamp time.Time
	err := pg.db.QueryRow(ctx, query, roomID, userID, message).Scan(&messageID, &timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to save chat message: %w", err)
	}

	// Get username for the response
	user, err := pg.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &ChatMessage{
		MessageID:   messageID,
		RoomID:      roomID,
		UserID:      userID,
		Username:    user.Username,
		MessageText: message,
		Timestamp:   timestamp,
	}, nil
}

// GetRecentMessages retrieves the most recent messages from a chat room
// Only returns messages from the last 30 minutes OR up to the limit, whichever is fewer
func (pg *Postgres) GetRecentMessages(ctx context.Context, roomID int, limit int) ([]*ChatMessage, error) {
	query := `
		SELECT
			cm.message_id,
			cm.room_id,
			cm.user_id,
			u.username,
			cm.message_text,
			cm.timestamp
		FROM CHAT_MESSAGE cm
		JOIN "USER" u ON cm.user_id = u.user_id
		WHERE cm.room_id = $1 AND cm.timestamp > NOW() - INTERVAL '30 minutes'
		ORDER BY cm.timestamp DESC
		LIMIT $2
	`

	rows, err := pg.db.Query(ctx, query, roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent messages: %w", err)
	}
	defer rows.Close()

	var messages []*ChatMessage
	for rows.Next() {
		var msg ChatMessage
		err := rows.Scan(
			&msg.MessageID,
			&msg.RoomID,
			&msg.UserID,
			&msg.Username,
			&msg.MessageText,
			&msg.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	// Reverse the slice to get chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetMessagesAfter retrieves messages after a specific timestamp
// Useful for syncing messages after reconnection
func (pg *Postgres) GetMessagesAfter(ctx context.Context, roomID int, after time.Time) ([]*ChatMessage, error) {
	query := `
		SELECT
			cm.message_id,
			cm.room_id,
			cm.user_id,
			u.username,
			cm.message_text,
			cm.timestamp
		FROM CHAT_MESSAGE cm
		JOIN "USER" u ON cm.user_id = u.user_id
		WHERE cm.room_id = $1 AND cm.timestamp > $2
		ORDER BY cm.timestamp ASC
	`

	rows, err := pg.db.Query(ctx, query, roomID, after)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages after timestamp: %w", err)
	}
	defer rows.Close()

	var messages []*ChatMessage
	for rows.Next() {
		var msg ChatMessage
		err := rows.Scan(
			&msg.MessageID,
			&msg.RoomID,
			&msg.UserID,
			&msg.Username,
			&msg.MessageText,
			&msg.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}
