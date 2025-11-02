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

// ForfeitGame marks a game as abandoned with the opponent as winner
func (pg *Postgres) ForfeitGame(ctx context.Context, gameID int, forfeitingPlayerID int) error {
	// Get game details to determine the winner
	game, err := pg.GetGameByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Determine the winner (the other player)
	var winnerID int
	if game.Player1ID == forfeitingPlayerID {
		winnerID = game.Player2ID
	} else if game.Player2ID == forfeitingPlayerID {
		winnerID = game.Player1ID
	} else {
		return fmt.Errorf("player not in this game")
	}

	// Update game as abandoned with winner
	query := `
		UPDATE GAME
		SET game_status = 'abandoned',
		    winner_id = $2,
		    ended_at = NOW()
		WHERE game_id = $1
	`

	_, err = pg.db.Exec(ctx, query, gameID, winnerID)
	if err != nil {
		return fmt.Errorf("failed to forfeit game: %w", err)
	}

	return nil
}

// CompleteGame marks a game as completed with a winner
func (pg *Postgres) CompleteGame(ctx context.Context, gameID int, winnerID int) error {
	// Verify the winner is a player in this game
	game, err := pg.GetGameByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	if winnerID != game.Player1ID && winnerID != game.Player2ID {
		return fmt.Errorf("winner must be a player in this game")
	}

	query := `
		UPDATE GAME
		SET game_status = 'completed',
		    winner_id = $2,
		    ended_at = NOW()
		WHERE game_id = $1
	`

	_, err = pg.db.Exec(ctx, query, gameID, winnerID)
	if err != nil {
		return fmt.Errorf("failed to complete game: %w", err)
	}

	return nil
}

// StartGame marks a game as in_progress
func (pg *Postgres) StartGame(ctx context.Context, gameID int) error {
	query := `
		UPDATE GAME
		SET game_status = 'in_progress',
		    started_at = NOW()
		WHERE game_id = $1 AND game_status = 'pending'
	`

	result, err := pg.db.Exec(ctx, query, gameID)
	if err != nil {
		return fmt.Errorf("failed to start game: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("game not found or already started")
	}

	return nil
}

// GetGameWithPlayers retrieves a game with player usernames
type GameWithPlayers struct {
	GameID         int
	Player1ID      int
	Player1Username string
	Player1Color   string
	Player2ID      int
	Player2Username string
	Player2Color   string
	CurrentTurn    int
	GameStatus     string
	WinnerID       *int
	CreatedAt      time.Time
	StartedAt      *time.Time
	EndedAt        *time.Time
}

func (pg *Postgres) GetGameWithPlayers(ctx context.Context, gameID int) (*GameWithPlayers, error) {
	query := `
		SELECT
			g.game_id,
			g.player1_id,
			u1.username as player1_username,
			g.player1_color,
			g.player2_id,
			u2.username as player2_username,
			g.player2_color,
			g.current_turn,
			g.game_status,
			g.winner_id,
			g.created_at,
			g.started_at,
			g.ended_at
		FROM GAME g
		JOIN "USER" u1 ON g.player1_id = u1.user_id
		JOIN "USER" u2 ON g.player2_id = u2.user_id
		WHERE g.game_id = $1
	`

	var game GameWithPlayers
	err := pg.db.QueryRow(ctx, query, gameID).Scan(
		&game.GameID,
		&game.Player1ID,
		&game.Player1Username,
		&game.Player1Color,
		&game.Player2ID,
		&game.Player2Username,
		&game.Player2Color,
		&game.CurrentTurn,
		&game.GameStatus,
		&game.WinnerID,
		&game.CreatedAt,
		&game.StartedAt,
		&game.EndedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get game with players: %w", err)
	}

	return &game, nil
}

// GetActiveGamesForUser retrieves all active games for a user
func (pg *Postgres) GetActiveGamesForUser(ctx context.Context, userID int) ([]GameWithPlayers, error) {
	query := `
		SELECT
			g.game_id,
			g.player1_id,
			u1.username as player1_username,
			g.player1_color,
			g.player2_id,
			u2.username as player2_username,
			g.player2_color,
			g.current_turn,
			g.game_status,
			g.winner_id,
			g.created_at,
			g.started_at,
			g.ended_at
		FROM GAME g
		JOIN "USER" u1 ON g.player1_id = u1.user_id
		JOIN "USER" u2 ON g.player2_id = u2.user_id
		WHERE (g.player1_id = $1 OR g.player2_id = $1)
		  AND g.game_status IN ('pending', 'in_progress')
		ORDER BY g.created_at DESC
	`

	rows, err := pg.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active games: %w", err)
	}
	defer rows.Close()

	games := []GameWithPlayers{}
	for rows.Next() {
		var game GameWithPlayers
		err := rows.Scan(
			&game.GameID,
			&game.Player1ID,
			&game.Player1Username,
			&game.Player1Color,
			&game.Player2ID,
			&game.Player2Username,
			&game.Player2Color,
			&game.CurrentTurn,
			&game.GameStatus,
			&game.WinnerID,
			&game.CreatedAt,
			&game.StartedAt,
			&game.EndedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan active game: %w", err)
		}
		games = append(games, game)
	}

	return games, nil
}
