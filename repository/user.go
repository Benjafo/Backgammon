package repository

import (
	"context"
	"fmt"
)

type User struct {
	UserID       int
	Username     string
	PasswordHash string
	Email        *string
}

// CreateUser inserts a new user into the database
func (pg *Postgres) CreateUser(ctx context.Context, username, passwordHash string, email *string) (int, error) {
	query := `
		INSERT INTO "USER" (username, password_hash, email)
		VALUES ($1, $2, $3)
		RETURNING user_id
	`

	var userID int
	err := pg.db.QueryRow(ctx, query, username, passwordHash, email).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

// GetUserByUsername retrieves a user by username
func (pg *Postgres) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT user_id, username, password_hash, email
		FROM "USER"
		WHERE username = $1
	`

	var user User
	err := pg.db.QueryRow(ctx, query, username).Scan(
		&user.UserID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (pg *Postgres) GetUserByID(ctx context.Context, userID int) (*User, error) {
	query := `
		SELECT user_id, username, password_hash, email
		FROM "USER"
		WHERE user_id = $1
	`

	var user User
	err := pg.db.QueryRow(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}
