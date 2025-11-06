package repository

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/jackc/pgx/v5"
)

// Create a new game between two players with random color and turn assignment
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

// Retrieve a game by its ID
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

// Update the status of a game
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

// Mark a game as abandoned with the opponent as winner
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

// Mark a game as completed with a winner
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

// Mark a game as in_progress
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

// Retrieve a game with player usernames
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

// Retrieve all active games for a user
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

// ============================================================================
// GAME_STATE Management
// ============================================================================

// Create the initial board state for a new game
func (pg *Postgres) InitializeGameState(ctx context.Context, gameID int) error {
	// Standard backgammon setup:
	// White moves from 24->1 (counterclockwise), Black moves from 1->24 (clockwise)
	// Point 1: 2 black, Point 6: 5 white, Point 8: 3 white, Point 12: 5 black
	// Point 13: 5 white, Point 17: 3 black, Point 19: 5 black, Point 24: 2 white
	// Using array indices 0-23 for points 1-24
	initialBoard := make([]int, 24)
	initialBoard[0] = -2   // Point 1: 2 black
	initialBoard[5] = 5    // Point 6: 5 white
	initialBoard[7] = 3    // Point 8: 3 white
	initialBoard[11] = -5  // Point 12: 5 black
	initialBoard[12] = 5   // Point 13: 5 white
	initialBoard[16] = -3  // Point 17: 3 black
	initialBoard[18] = -5  // Point 19: 5 black
	initialBoard[23] = 2   // Point 24: 2 white

	// TESTING SETUP (commented out - for testing bear-off):
	// Both players have checkers in home board for testing bear-off
	// White home: points 1-6, Black home: points 19-24
	// initialBoard := make([]int, 24)
	// initialBoard[3] = 5    // Point 4: 5 white
	// initialBoard[4] = 5    // Point 5: 5 white
	// initialBoard[5] = 5    // Point 6: 5 white
	// initialBoard[18] = -5  // Point 19: 5 black
	// initialBoard[19] = -5  // Point 20: 5 black
	// initialBoard[20] = -5  // Point 21: 5 black

	boardJSON, err := json.Marshal(initialBoard)
	if err != nil {
		return fmt.Errorf("failed to marshal board state: %w", err)
	}

	query := `
		INSERT INTO GAME_STATE (
			game_id, board_state, bar_white, bar_black,
			borne_off_white, borne_off_black, dice_roll, dice_used, last_updated
		)
		VALUES ($1, $2, 0, 0, 0, 0, NULL, NULL, NOW())
	`

	_, err = pg.db.Exec(ctx, query, gameID, boardJSON)
	if err != nil {
		return fmt.Errorf("failed to initialize game state: %w", err)
	}

	return nil
}

// Retrieve the current game state
func (pg *Postgres) GetGameState(ctx context.Context, gameID int) (*GameState, error) {
	query := `
		SELECT
			state_id, game_id, board_state, bar_white, bar_black,
			borne_off_white, borne_off_black, dice_roll, dice_used, last_updated
		FROM GAME_STATE
		WHERE game_id = $1
	`

	var state GameState
	var boardJSON []byte
	var diceRollJSON []byte
	var diceUsedJSON []byte

	err := pg.db.QueryRow(ctx, query, gameID).Scan(
		&state.StateID,
		&state.GameID,
		&boardJSON,
		&state.BarWhite,
		&state.BarBlack,
		&state.BornedOffWhite,
		&state.BornedOffBlack,
		&diceRollJSON,
		&diceUsedJSON,
		&state.LastUpdated,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("game state not found")
		}
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	// Unmarshal board state
	if err := json.Unmarshal(boardJSON, &state.BoardState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal board state: %w", err)
	}

	// Unmarshal dice roll if present
	if diceRollJSON != nil {
		if err := json.Unmarshal(diceRollJSON, &state.DiceRoll); err != nil {
			return nil, fmt.Errorf("failed to unmarshal dice roll: %w", err)
		}
	}

	// Unmarshal dice used if present
	if diceUsedJSON != nil {
		if err := json.Unmarshal(diceUsedJSON, &state.DiceUsed); err != nil {
			return nil, fmt.Errorf("failed to unmarshal dice used: %w", err)
		}
	}

	return &state, nil
}

// Update the game state
func (pg *Postgres) UpdateGameState(ctx context.Context, state *GameState) error {
	boardJSON, err := json.Marshal(state.BoardState)
	if err != nil {
		return fmt.Errorf("failed to marshal board state: %w", err)
	}

	var diceRollJSON []byte
	var diceUsedJSON []byte

	if state.DiceRoll != nil {
		diceRollJSON, err = json.Marshal(state.DiceRoll)
		if err != nil {
			return fmt.Errorf("failed to marshal dice roll: %w", err)
		}
	}

	if state.DiceUsed != nil {
		diceUsedJSON, err = json.Marshal(state.DiceUsed)
		if err != nil {
			return fmt.Errorf("failed to marshal dice used: %w", err)
		}
	}

	query := `
		UPDATE GAME_STATE
		SET board_state = $2,
		    bar_white = $3,
		    bar_black = $4,
		    borne_off_white = $5,
		    borne_off_black = $6,
		    dice_roll = $7,
		    dice_used = $8,
		    last_updated = NOW()
		WHERE game_id = $1
	`

	result, err := pg.db.Exec(ctx, query,
		state.GameID,
		boardJSON,
		state.BarWhite,
		state.BarBlack,
		state.BornedOffWhite,
		state.BornedOffBlack,
		diceRollJSON,
		diceUsedJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to update game state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("game state not found")
	}

	return nil
}

// Generate a new dice roll for the current turn
func (pg *Postgres) RollDice(ctx context.Context, gameID int) ([]int, error) {
	// Generate two random dice (1-6)
	die1, err := rand.Int(rand.Reader, big.NewInt(6))
	if err != nil {
		return nil, fmt.Errorf("failed to generate die 1: %w", err)
	}

	die2, err := rand.Int(rand.Reader, big.NewInt(6))
	if err != nil {
		return nil, fmt.Errorf("failed to generate die 2: %w", err)
	}

	val1 := int(die1.Int64()) + 1
	val2 := int(die2.Int64()) + 1

	// For doubles, player gets 4 moves of the same value
	var dice []int
	var diceUsed []bool
	if val1 == val2 {
		dice = []int{val1, val1, val1, val1}
		diceUsed = []bool{false, false, false, false}
	} else {
		dice = []int{val1, val2}
		diceUsed = []bool{false, false}
	}

	diceJSON, err := json.Marshal(dice)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dice: %w", err)
	}

	diceUsedJSON, err := json.Marshal(diceUsed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dice used: %w", err)
	}

	query := `
		UPDATE GAME_STATE
		SET dice_roll = $2, dice_used = $3, last_updated = NOW()
		WHERE game_id = $1
	`

	result, err := pg.db.Exec(ctx, query, gameID, diceJSON, diceUsedJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to roll dice: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil, fmt.Errorf("game state not found")
	}

	return dice, nil
}

// Clear the dice roll at the end of a turn
func (pg *Postgres) ClearDice(ctx context.Context, gameID int) error {
	query := `
		UPDATE GAME_STATE
		SET dice_roll = NULL, dice_used = NULL, last_updated = NOW()
		WHERE game_id = $1
	`

	_, err := pg.db.Exec(ctx, query, gameID)
	if err != nil {
		return fmt.Errorf("failed to clear dice: %w", err)
	}

	return nil
}

// ============================================================================
// MOVE History Management
// ============================================================================

// Record a move in the database
func (pg *Postgres) CreateMove(ctx context.Context, move *Move) (int, error) {
	query := `
		INSERT INTO MOVE (
			game_id, player_id, move_number, from_point, to_point,
			die_used, hit_opponent, timestamp
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING move_id
	`

	var moveID int
	err := pg.db.QueryRow(ctx, query,
		move.GameID,
		move.PlayerID,
		move.MoveNumber,
		move.FromPoint,
		move.ToPoint,
		move.DieUsed,
		move.HitOpponent,
	).Scan(&moveID)

	if err != nil {
		return 0, fmt.Errorf("failed to create move: %w", err)
	}

	return moveID, nil
}

// Retrieve all moves for a game
func (pg *Postgres) GetMoveHistory(ctx context.Context, gameID int) ([]Move, error) {
	query := `
		SELECT
			move_id, game_id, player_id, move_number, from_point,
			to_point, die_used, hit_opponent, timestamp
		FROM MOVE
		WHERE game_id = $1
		ORDER BY move_number ASC, timestamp ASC
	`

	rows, err := pg.db.Query(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get move history: %w", err)
	}
	defer rows.Close()

	moves := []Move{}
	for rows.Next() {
		var move Move
		err := rows.Scan(
			&move.MoveID,
			&move.GameID,
			&move.PlayerID,
			&move.MoveNumber,
			&move.FromPoint,
			&move.ToPoint,
			&move.DieUsed,
			&move.HitOpponent,
			&move.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan move: %w", err)
		}
		moves = append(moves, move)
	}

	return moves, nil
}

// Get the latest move number for a game
func (pg *Postgres) GetLastMoveNumber(ctx context.Context, gameID int) (int, error) {
	query := `
		SELECT COALESCE(MAX(move_number), 0)
		FROM MOVE
		WHERE game_id = $1
	`

	var moveNumber int
	err := pg.db.QueryRow(ctx, query, gameID).Scan(&moveNumber)
	if err != nil {
		return 0, fmt.Errorf("failed to get last move number: %w", err)
	}

	return moveNumber, nil
}

// Update whose turn it is
func (pg *Postgres) UpdateGameTurn(ctx context.Context, gameID int, playerID int) error {
	query := `
		UPDATE GAME
		SET current_turn = $2
		WHERE game_id = $1
	`

	_, err := pg.db.Exec(ctx, query, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to update game turn: %w", err)
	}

	return nil
}
