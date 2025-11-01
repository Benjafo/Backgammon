package repository

import (
	"context"
	"fmt"
	"time"
)

// LobbyUser represents a user currently in the lobby
type LobbyUser struct {
	UserID        int
	Username      string
	JoinedAt      time.Time
	LastHeartbeat time.Time
}

// JoinLobby inserts a new presence record for a user joining the lobby
// If user is already in lobby, updates their heartbeat (idempotent operation)
func (pg *Postgres) JoinLobby(ctx context.Context, userID int) (int, error) {
	query := `
		INSERT INTO LOBBY_PRESENCE (user_id, joined_at, last_heartbeat)
		VALUES ($1, NOW(), NOW())
		ON CONFLICT (user_id)
		DO UPDATE SET last_heartbeat = NOW()
		RETURNING presence_id
	`

	var presenceID int
	err := pg.db.QueryRow(ctx, query, userID).Scan(&presenceID)
	if err != nil {
		return 0, fmt.Errorf("failed to join lobby: %w", err)
	}

	return presenceID, nil
}

// LeaveLobby removes a user's presence record from the lobby
func (pg *Postgres) LeaveLobby(ctx context.Context, userID int) error {
	query := `
		DELETE FROM LOBBY_PRESENCE
		WHERE user_id = $1
	`

	_, err := pg.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to leave lobby: %w", err)
	}

	return nil
}

// UpdateHeartbeat updates the last_heartbeat timestamp for a user in the lobby
func (pg *Postgres) UpdateHeartbeat(ctx context.Context, userID int) error {
	query := `
		UPDATE LOBBY_PRESENCE
		SET last_heartbeat = NOW()
		WHERE user_id = $1
	`

	result, err := pg.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not in lobby")
	}

	return nil
}

// GetLobbyUsers retrieves all users currently in the lobby with their details
func (pg *Postgres) GetLobbyUsers(ctx context.Context) ([]LobbyUser, error) {
	query := `
		SELECT lp.user_id, u.username, lp.joined_at, lp.last_heartbeat
		FROM LOBBY_PRESENCE lp
		JOIN "USER" u ON lp.user_id = u.user_id
		ORDER BY lp.joined_at DESC
	`

	rows, err := pg.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get lobby users: %w", err)
	}
	defer rows.Close()

	var users []LobbyUser
	for rows.Next() {
		var user LobbyUser
		err := rows.Scan(&user.UserID, &user.Username, &user.JoinedAt, &user.LastHeartbeat)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lobby user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lobby users: %w", err)
	}

	return users, nil
}

// CleanupStaleLobbyPresence removes presence records where last_heartbeat is older than the timeout
func (pg *Postgres) CleanupStaleLobbyPresence(ctx context.Context, timeout time.Duration) (int64, error) {
	query := `
		DELETE FROM LOBBY_PRESENCE
		WHERE last_heartbeat < NOW() - $1::interval
	`

	result, err := pg.db.Exec(ctx, query, timeout)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup stale lobby presence: %w", err)
	}

	return result.RowsAffected(), nil
}

// IsUserInLobby checks if a user is currently in the lobby
func (pg *Postgres) IsUserInLobby(ctx context.Context, userID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM LOBBY_PRESENCE WHERE user_id = $1
		)
	`

	var exists bool
	err := pg.db.QueryRow(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user in lobby: %w", err)
	}

	return exists, nil
}
