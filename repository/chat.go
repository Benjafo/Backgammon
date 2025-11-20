package repository

import (
	"context"
	"fmt"
	"time"
)

// GetOrCreateLobbyChatRoom ensures the lobby chat room exists and returns its ID
func (pg *Postgres) GetOrCreateLobbyChatRoom(ctx context.Context) (int, error) {
	// First, try to get existing lobby chat room
	queryExisting := `
		SELECT room_id FROM CHAT_ROOM WHERE room_type = 'lobby' LIMIT 1
	`

	var roomID int
	err := pg.db.QueryRow(ctx, queryExisting).Scan(&roomID)
	if err == nil {
		// Found existing lobby room
		return roomID, nil
	}

	// No existing lobby room, create one
	queryInsert := `
		INSERT INTO CHAT_ROOM (room_type, game_id)
		VALUES ('lobby', NULL)
		RETURNING room_id
	`

	err = pg.db.QueryRow(ctx, queryInsert).Scan(&roomID)
	if err != nil {
		return 0, fmt.Errorf("failed to create lobby chat room: %w", err)
	}

	return roomID, nil
}

// SaveMessage saves a chat message to the database
func (pg *Postgres) SaveMessage(ctx context.Context, roomID, userID int, messageText string) (ChatMessage, error) {
	query := `
		INSERT INTO CHAT_MESSAGE (room_id, user_id, message_text)
		VALUES ($1, $2, $3)
		RETURNING message_id, timestamp
	`

	var message ChatMessage
	message.RoomID = roomID
	message.UserID = userID
	message.MessageText = messageText

	err := pg.db.QueryRow(ctx, query, roomID, userID, messageText).Scan(&message.MessageID, &message.Timestamp)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("failed to save message: %w", err)
	}

	return message, nil
}

// GetRecentMessages retrieves the most recent messages from a chat room
func (pg *Postgres) GetRecentMessages(ctx context.Context, roomID int, limit int) ([]ChatMessageWithUser, error) {
	query := `
		SELECT cm.message_id, cm.room_id, cm.user_id, u.username, cm.message_text, cm.timestamp
		FROM CHAT_MESSAGE cm
		JOIN "USER" u ON cm.user_id = u.user_id
		WHERE cm.room_id = $1
		ORDER BY cm.timestamp DESC
		LIMIT $2
	`

	rows, err := pg.db.Query(ctx, query, roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent messages: %w", err)
	}
	defer rows.Close()

	var messages []ChatMessageWithUser
	for rows.Next() {
		var msg ChatMessageWithUser
		err := rows.Scan(&msg.MessageID, &msg.RoomID, &msg.UserID, &msg.Username, &msg.MessageText, &msg.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	// Reverse to get chronological order (oldest to newest)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// CleanupOldMessages removes messages older than the specified duration
func (pg *Postgres) CleanupOldMessages(ctx context.Context, roomID int, maxAge time.Duration) (int64, error) {
	query := `
		DELETE FROM CHAT_MESSAGE
		WHERE room_id = $1 AND timestamp < NOW() - $2::interval
	`

	result, err := pg.db.Exec(ctx, query, roomID, maxAge)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old messages: %w", err)
	}

	return result.RowsAffected(), nil
}
