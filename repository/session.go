package repository

import (
	"context"
	"fmt"
	"time"
)

// CreateSession inserts a new session into the database
func (pg *Postgres) CreateSession(ctx context.Context, userID int, sessionToken, ipAddress, userAgent string, expiresAt time.Time) error {
	query := `
		INSERT INTO SESSIONS (user_id, session_token, ip_address, user_agent, expires_at, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`

	_, err := pg.db.Exec(ctx, query, userID, sessionToken, ipAddress, userAgent, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSessionByToken retrieves a session by token and validates it
func (pg *Postgres) GetSessionByToken(ctx context.Context, sessionToken string) (*Session, error) {
	query := `
		SELECT session_id, user_id, session_token, ip_address, user_agent, created_at, expires_at, is_active
		FROM SESSIONS
		WHERE session_token = $1 AND is_active = true AND expires_at > NOW()
	`

	var session Session
	err := pg.db.QueryRow(ctx, query, sessionToken).Scan(
		&session.SessionID,
		&session.UserID,
		&session.SessionToken,
		&session.IPAddress,
		&session.UserAgent,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired session: %w", err)
	}

	return &session, nil
}

// DeleteSession invalidates a session (logout)
func (pg *Postgres) DeleteSession(ctx context.Context, sessionToken string) error {
	query := `
		UPDATE SESSIONS
		SET is_active = false
		WHERE session_token = $1
	`

	_, err := pg.db.Exec(ctx, query, sessionToken)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions (can be called periodically)
func (pg *Postgres) CleanupExpiredSessions(ctx context.Context) error {
	query := `
		DELETE FROM SESSIONS
		WHERE expires_at < NOW() OR is_active = false
	`

	_, err := pg.db.Exec(ctx, query)
	return err
}
