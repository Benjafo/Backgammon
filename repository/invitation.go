package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Invitation represents a game invitation between two players
type Invitation struct {
	InvitationID int
	ChallengerID int
	ChallengedID int
	Status       string
	GameID       *int
	CreatedAt    time.Time
	// Extended fields for joined queries
	ChallengerUsername string
	ChallengedUsername string
}

// InvitationWithUsers contains invitation with full user details
type InvitationWithUsers struct {
	InvitationID       int
	ChallengerID       int
	ChallengerUsername string
	ChallengedID       int
	ChallengedUsername string
	Status             string
	GameID             *int
	CreatedAt          time.Time
}

// CreateInvitation creates a new game invitation
func (pg *Postgres) CreateInvitation(ctx context.Context, challengerID, challengedID int) (int, error) {
	// Check for existing pending invitation between these users
	checkQuery := `
		SELECT invitation_id FROM GAME_INVITATION
		WHERE ((challenger_id = $1 AND challenged_id = $2) OR (challenger_id = $2 AND challenged_id = $1))
		AND status = 'pending'
	`

	var existingID int
	err := pg.db.QueryRow(ctx, checkQuery, challengerID, challengedID).Scan(&existingID)
	if err == nil {
		return 0, fmt.Errorf("pending invitation already exists")
	} else if err != pgx.ErrNoRows {
		return 0, fmt.Errorf("failed to check existing invitation: %w", err)
	}

	// Create new invitation
	query := `
		INSERT INTO GAME_INVITATION (challenger_id, challenged_id, status, created_at)
		VALUES ($1, $2, 'pending', NOW())
		RETURNING invitation_id
	`

	var invitationID int
	err = pg.db.QueryRow(ctx, query, challengerID, challengedID).Scan(&invitationID)
	if err != nil {
		return 0, fmt.Errorf("failed to create invitation: %w", err)
	}

	return invitationID, nil
}

// GetInvitationsByUser retrieves all invitations for a user (both sent and received)
func (pg *Postgres) GetInvitationsByUser(ctx context.Context, userID int) (sent []InvitationWithUsers, received []InvitationWithUsers, error error) {
	// Get sent invitations
	sentQuery := `
		SELECT
			gi.invitation_id,
			gi.challenger_id,
			u1.username as challenger_username,
			gi.challenged_id,
			u2.username as challenged_username,
			gi.status,
			gi.game_id,
			gi.created_at
		FROM GAME_INVITATION gi
		JOIN "USER" u1 ON gi.challenger_id = u1.user_id
		JOIN "USER" u2 ON gi.challenged_id = u2.user_id
		WHERE gi.challenger_id = $1 AND gi.status = 'pending'
		ORDER BY gi.created_at DESC
	`

	rows, err := pg.db.Query(ctx, sentQuery, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get sent invitations: %w", err)
	}
	defer rows.Close()

	sent = []InvitationWithUsers{}
	for rows.Next() {
		var inv InvitationWithUsers
		err := rows.Scan(
			&inv.InvitationID,
			&inv.ChallengerID,
			&inv.ChallengerUsername,
			&inv.ChallengedID,
			&inv.ChallengedUsername,
			&inv.Status,
			&inv.GameID,
			&inv.CreatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan sent invitation: %w", err)
		}
		sent = append(sent, inv)
	}

	// Get received invitations
	receivedQuery := `
		SELECT
			gi.invitation_id,
			gi.challenger_id,
			u1.username as challenger_username,
			gi.challenged_id,
			u2.username as challenged_username,
			gi.status,
			gi.game_id,
			gi.created_at
		FROM GAME_INVITATION gi
		JOIN "USER" u1 ON gi.challenger_id = u1.user_id
		JOIN "USER" u2 ON gi.challenged_id = u2.user_id
		WHERE gi.challenged_id = $1 AND gi.status = 'pending'
		ORDER BY gi.created_at DESC
	`

	rows, err = pg.db.Query(ctx, receivedQuery, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get received invitations: %w", err)
	}
	defer rows.Close()

	received = []InvitationWithUsers{}
	for rows.Next() {
		var inv InvitationWithUsers
		err := rows.Scan(
			&inv.InvitationID,
			&inv.ChallengerID,
			&inv.ChallengerUsername,
			&inv.ChallengedID,
			&inv.ChallengedUsername,
			&inv.Status,
			&inv.GameID,
			&inv.CreatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan received invitation: %w", err)
		}
		received = append(received, inv)
	}

	return sent, received, nil
}

// GetInvitationByID retrieves a single invitation by ID
func (pg *Postgres) GetInvitationByID(ctx context.Context, invitationID int) (*InvitationWithUsers, error) {
	query := `
		SELECT
			gi.invitation_id,
			gi.challenger_id,
			u1.username as challenger_username,
			gi.challenged_id,
			u2.username as challenged_username,
			gi.status,
			gi.game_id,
			gi.created_at
		FROM GAME_INVITATION gi
		JOIN "USER" u1 ON gi.challenger_id = u1.user_id
		JOIN "USER" u2 ON gi.challenged_id = u2.user_id
		WHERE gi.invitation_id = $1
	`

	var inv InvitationWithUsers
	err := pg.db.QueryRow(ctx, query, invitationID).Scan(
		&inv.InvitationID,
		&inv.ChallengerID,
		&inv.ChallengerUsername,
		&inv.ChallengedID,
		&inv.ChallengedUsername,
		&inv.Status,
		&inv.GameID,
		&inv.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invitation not found")
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	return &inv, nil
}

// AcceptInvitation updates an invitation to accepted status and links it to a game
func (pg *Postgres) AcceptInvitation(ctx context.Context, invitationID, gameID int) error {
	query := `
		UPDATE GAME_INVITATION
		SET status = 'accepted', game_id = $2
		WHERE invitation_id = $1 AND status = 'pending'
	`

	result, err := pg.db.Exec(ctx, query, invitationID, gameID)
	if err != nil {
		return fmt.Errorf("failed to accept invitation: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("invitation not found or already processed")
	}

	return nil
}

// DeclineInvitation updates an invitation to declined status
func (pg *Postgres) DeclineInvitation(ctx context.Context, invitationID int) error {
	query := `
		UPDATE GAME_INVITATION
		SET status = 'declined'
		WHERE invitation_id = $1 AND status = 'pending'
	`

	result, err := pg.db.Exec(ctx, query, invitationID)
	if err != nil {
		return fmt.Errorf("failed to decline invitation: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("invitation not found or already processed")
	}

	return nil
}

// CancelInvitation deletes an invitation (challenger only)
func (pg *Postgres) CancelInvitation(ctx context.Context, invitationID int) error {
	query := `
		DELETE FROM GAME_INVITATION
		WHERE invitation_id = $1 AND status = 'pending'
	`

	result, err := pg.db.Exec(ctx, query, invitationID)
	if err != nil {
		return fmt.Errorf("failed to cancel invitation: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("invitation not found or already processed")
	}

	return nil
}

// CleanupExpiredInvitations marks old pending invitations as expired
func (pg *Postgres) CleanupExpiredInvitations(ctx context.Context, expirationTime time.Duration) (int64, error) {
	query := `
		UPDATE GAME_INVITATION
		SET status = 'expired'
		WHERE status = 'pending' AND created_at < NOW() - $1::interval
	`

	result, err := pg.db.Exec(ctx, query, expirationTime)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired invitations: %w", err)
	}

	return result.RowsAffected(), nil
}
