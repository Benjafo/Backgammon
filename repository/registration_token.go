package repository

import (
	"context"
	"fmt"
	"time"
)

// CreateRegistrationToken inserts a new registration token (for CSRF protection)
func (pg *Postgres) CreateRegistrationToken(ctx context.Context, tokenValue, ipAddress, userAgent string, expiresAt time.Time) error {
	query := `
		INSERT INTO REGISTRATION_TOKEN (token_value, ip_address, user_agent, expires_at, is_used)
		VALUES ($1, $2, $3, $4, false)
	`

	_, err := pg.db.Exec(ctx, query, tokenValue, ipAddress, userAgent, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create registration token: %w", err)
	}

	return nil
}

// ValidateAndUseRegistrationToken checks if token is valid and marks it as used
func (pg *Postgres) ValidateAndUseRegistrationToken(ctx context.Context, tokenValue string) error {
	// Check if token exists and is valid
	checkQuery := `
		SELECT token_id FROM REGISTRATION_TOKEN
		WHERE token_value = $1 AND is_used = false AND expires_at > NOW()
	`

	var tokenID int
	err := pg.db.QueryRow(ctx, checkQuery, tokenValue).Scan(&tokenID)
	if err != nil {
		return fmt.Errorf("invalid or expired registration token: %w", err)
	}

	// Mark token as used
	updateQuery := `
		UPDATE REGISTRATION_TOKEN
		SET is_used = true
		WHERE token_id = $1
	`

	_, err = pg.db.Exec(ctx, updateQuery, tokenID)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// CleanupExpiredRegistrationTokens removes expired tokens
func (pg *Postgres) CleanupExpiredRegistrationTokens(ctx context.Context) error {
	query := `
		DELETE FROM REGISTRATION_TOKEN
		WHERE expires_at < NOW() OR is_used = true
	`

	_, err := pg.db.Exec(ctx, query)
	return err
}
