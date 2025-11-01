package repository

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// Game represents a backgammon game
type Game struct {
	GameID       int
	Player1ID    int
	Player2ID    int
	CurrentTurn  int
	GameStatus   string
	WinnerID     *int
	CreatedAt    time.Time
	StartedAt    *time.Time
	EndedAt      *time.Time
	Player1Color string
	Player2Color string
}

// CreateGame creates a new game between two players with random color and turn assignment
func (pg *Postgres) CreateGame(ctx context.Context, player1ID, player2ID int) (int, error) {
	// Validate that players are different
	if player1ID == player2ID {
		return 0, fmt.Errorf("cannot create game with same player")
	}

	// Randomly assign colors (0 = player1 is white, 1 = player1 is black)
	colorRand, err := rand.Int(rand.Reader, big.NewInt(2))
	if err != nil {
		return 0, fmt.Errorf("failed to generate random color: %w", err)
	}

	var player1Color, player2Color string
	if colorRand.Int64() == 0 {
		player1Color = "white"
		player2Color = "black"
	} else {
		player1Color = "black"
		player2Color = "white"
	}

	// Randomly select starting player (0 = player1, 1 = player2)
	turnRand, err := rand.Int(rand.Reader, big.NewInt(2))
	if err != nil {
		return 0, fmt.Errorf("failed to generate random turn: %w", err)
	}

	var currentTurn int
	if turnRand.Int64() == 0 {
		currentTurn = player1ID
	} else {
		currentTurn = player2ID
	}

	// Create game record
	query := `
		INSERT INTO GAME (
			player1_id,
			player2_id,
			current_turn,
			game_status,
			player1_color,
			player2_color,
			created_at
		)
		VALUES ($1, $2, $3, 'pending', $4, $5, NOW())
		RETURNING game_id
	`

	var gameID int
	err = pg.db.QueryRow(ctx, query, player1ID, player2ID, currentTurn, player1Color, player2Color).Scan(&gameID)
	if err != nil {
		return 0, fmt.Errorf("failed to create game: %w", err)
	}

	return gameID, nil
}

// GetGameByID retrieves a game by its ID
func (pg *Postgres) GetGameByID(ctx context.Context, gameID int) (*Game, error) {
	query := `
		SELECT
			game_id,
			player1_id,
			player2_id,
			current_turn,
			game_status,
			winner_id,
			created_at,
			started_at,
			ended_at,
			player1_color,
			player2_color
		FROM GAME
		WHERE game_id = $1
	`

	var game Game
	err := pg.db.QueryRow(ctx, query, gameID).Scan(
		&game.GameID,
		&game.Player1ID,
		&game.Player2ID,
		&game.CurrentTurn,
		&game.GameStatus,
		&game.WinnerID,
		&game.CreatedAt,
		&game.StartedAt,
		&game.EndedAt,
		&game.Player1Color,
		&game.Player2Color,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return &game, nil
}

// UpdateGameStatus updates the status of a game
func (pg *Postgres) UpdateGameStatus(ctx context.Context, gameID int, status string) error {
	query := `
		UPDATE GAME
		SET game_status = $2
		WHERE game_id = $1
	`

	_, err := pg.db.Exec(ctx, query, gameID, status)
	if err != nil {
		return fmt.Errorf("failed to update game status: %w", err)
	}

	return nil
}
