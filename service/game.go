package service

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"backgammon/business"
	"backgammon/repository"
	"backgammon/util"
)

// GameRouterHandler routes game requests to the appropriate handler
func GameRouterHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// /api/v1/games/{id}/state - GET
	if strings.HasSuffix(path, "/state") && r.Method == http.MethodGet {
		GetGameStateHandler(w, r)
		return
	}

	// /api/v1/games/{id}/roll - POST
	if strings.HasSuffix(path, "/roll") && r.Method == http.MethodPost {
		RollDiceHandler(w, r)
		return
	}

	// /api/v1/games/{id}/move - POST
	if strings.HasSuffix(path, "/move") && r.Method == http.MethodPost {
		MoveHandler(w, r)
		return
	}

	// /api/v1/games/{id}/legal-moves - GET
	if strings.HasSuffix(path, "/legal-moves") && r.Method == http.MethodGet {
		GetLegalMovesHandler(w, r)
		return
	}

	// /api/v1/games/{id}/forfeit - POST
	if strings.HasSuffix(path, "/forfeit") && r.Method == http.MethodPost {
		ForfeitHandler(w, r)
		return
	}

	// /api/v1/games/{id}/start - POST
	if strings.HasSuffix(path, "/start") && r.Method == http.MethodPost {
		StartGameHandler(w, r)
		return
	}

	// /api/v1/games/{id} - GET
	if r.Method == http.MethodGet {
		GameHandler(w, r)
		return
	}

	util.ErrorResponse(w, http.StatusNotFound, "Endpoint not found")
}

// GameHandler retrieves game details
func GameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := repository.GetDB()
	if db == nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	// Get current user ID from context
	userID, ok := util.GetUserIDFromContext(r.Context())
	if !ok {
		util.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse game ID from URL path: /api/v1/games/{id}
	gameID, err := parseGameIDFromPath(r.URL.Path)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Invalid game ID")
		return
	}

	// Get game with player details
	game, err := db.GetGameWithPlayers(r.Context(), gameID)
	if err != nil {
		log.Printf("Failed to get game: %v", err)
		util.ErrorResponse(w, http.StatusNotFound, "Game not found")
		return
	}

	// Verify user is a player in this game
	if game.Player1ID != userID && game.Player2ID != userID {
		util.ErrorResponse(w, http.StatusForbidden, "You are not a player in this game")
		return
	}

	util.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"gameId": game.GameID,
		"player1": map[string]interface{}{
			"userId":   game.Player1ID,
			"username": game.Player1Username,
			"color":    game.Player1Color,
		},
		"player2": map[string]interface{}{
			"userId":   game.Player2ID,
			"username": game.Player2Username,
			"color":    game.Player2Color,
		},
		"currentTurn": game.CurrentTurn,
		"gameStatus":  game.GameStatus,
		"winnerId":    game.WinnerID,
		"createdAt":   game.CreatedAt,
		"startedAt":   game.StartedAt,
		"endedAt":     game.EndedAt,
	})
}

// ForfeitHandler allows a player to forfeit the game
func ForfeitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := repository.GetDB()
	if db == nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	// Get current user ID from context
	userID, ok := util.GetUserIDFromContext(r.Context())
	if !ok {
		util.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse game ID from URL path
	gameID, err := parseGameIDFromPath(strings.TrimSuffix(r.URL.Path, "/forfeit"))
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Invalid game ID")
		return
	}

	// Get game details
	game, err := db.GetGameByID(r.Context(), gameID)
	if err != nil {
		util.ErrorResponse(w, http.StatusNotFound, "Game not found")
		return
	}

	// Verify user is a player in this game
	if game.Player1ID != userID && game.Player2ID != userID {
		util.ErrorResponse(w, http.StatusForbidden, "You are not a player in this game")
		return
	}

	// Verify game is not already finished
	if game.GameStatus == "completed" || game.GameStatus == "abandoned" {
		util.ErrorResponse(w, http.StatusBadRequest, "Game already finished")
		return
	}

	// Forfeit the game
	err = db.ForfeitGame(r.Context(), gameID, userID)
	if err != nil {
		log.Printf("Failed to forfeit game: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to forfeit game")
		return
	}

	util.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Game forfeited successfully",
	})
}

// StartGameHandler starts a pending game
func StartGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := repository.GetDB()
	if db == nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	// Get current user ID from context
	userID, ok := util.GetUserIDFromContext(r.Context())
	if !ok {
		util.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse game ID from URL path
	gameID, err := parseGameIDFromPath(strings.TrimSuffix(r.URL.Path, "/start"))
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Invalid game ID")
		return
	}

	// Get game details
	game, err := db.GetGameByID(r.Context(), gameID)
	if err != nil {
		util.ErrorResponse(w, http.StatusNotFound, "Game not found")
		return
	}

	// Verify user is a player in this game
	if game.Player1ID != userID && game.Player2ID != userID {
		util.ErrorResponse(w, http.StatusForbidden, "You are not a player in this game")
		return
	}

	// Start the game
	err = db.StartGame(r.Context(), gameID)
	if err != nil {
		log.Printf("Failed to start game: %v", err)
		util.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	util.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Game started successfully",
	})
}

// ActiveGamesHandler returns active games for the current user
func ActiveGamesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := repository.GetDB()
	if db == nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	// Get current user ID from context
	userID, ok := util.GetUserIDFromContext(r.Context())
	if !ok {
		util.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get active games for user
	games, err := db.GetActiveGamesForUser(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get active games: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get active games")
		return
	}

	// Format game list
	gamesList := []map[string]interface{}{}
	for _, game := range games {
		gamesList = append(gamesList, map[string]interface{}{
			"gameId": game.GameID,
			"player1": map[string]interface{}{
				"userId":   game.Player1ID,
				"username": game.Player1Username,
				"color":    game.Player1Color,
			},
			"player2": map[string]interface{}{
				"userId":   game.Player2ID,
				"username": game.Player2Username,
				"color":    game.Player2Color,
			},
			"currentTurn": game.CurrentTurn,
			"gameStatus":  game.GameStatus,
			"winnerId":    game.WinnerID,
			"createdAt":   game.CreatedAt,
			"startedAt":   game.StartedAt,
			"endedAt":     game.EndedAt,
		})
	}

	util.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"games": gamesList,
	})
}

// parseGameIDFromPath extracts the game ID from the URL path
// Example: /api/v1/games/42 -> returns 42
func parseGameIDFromPath(path string) (int, error) {
	// Remove prefix
	trimmed := strings.TrimPrefix(path, "/api/v1/games/")

	// Parse the ID
	id, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// ============================================================================
// Game State Handlers
// ============================================================================

// GetGameStateHandler returns the current game state
func GetGameStateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := repository.GetDB()
	if db == nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	// Get current user ID from context
	userID, ok := util.GetUserIDFromContext(r.Context())
	if !ok {
		util.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse game ID from URL path
	gameID, err := parseGameIDFromPath(strings.TrimSuffix(r.URL.Path, "/state"))
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Invalid game ID")
		return
	}

	// Get game details to verify user is a player
	game, err := db.GetGameByID(r.Context(), gameID)
	if err != nil {
		util.ErrorResponse(w, http.StatusNotFound, "Game not found")
		return
	}

	// Verify user is a player in this game
	if game.Player1ID != userID && game.Player2ID != userID {
		util.ErrorResponse(w, http.StatusForbidden, "You are not a player in this game")
		return
	}

	// Get game state
	state, err := db.GetGameState(r.Context(), gameID)
	if err != nil {
		log.Printf("Failed to get game state: %v", err)
		util.ErrorResponse(w, http.StatusNotFound, "Game state not found")
		return
	}

	// Format response
	util.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"stateId":        state.StateID,
		"gameId":         state.GameID,
		"board":          state.BoardState,
		"barWhite":       state.BarWhite,
		"barBlack":       state.BarBlack,
		"bornedOffWhite": state.BornedOffWhite,
		"bornedOffBlack": state.BornedOffBlack,
		"diceRoll":       state.DiceRoll,
		"diceUsed":       state.DiceUsed,
		"lastUpdated":    state.LastUpdated,
	})
}

// RollDiceHandler rolls dice for the current turn
func RollDiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := repository.GetDB()
	if db == nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	// Get current user ID from context
	userID, ok := util.GetUserIDFromContext(r.Context())
	if !ok {
		util.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse game ID from URL path
	gameID, err := parseGameIDFromPath(strings.TrimSuffix(r.URL.Path, "/roll"))
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Invalid game ID")
		return
	}

	// Get game details
	game, err := db.GetGameByID(r.Context(), gameID)
	if err != nil {
		util.ErrorResponse(w, http.StatusNotFound, "Game not found")
		return
	}

	// Verify user is a player in this game
	if game.Player1ID != userID && game.Player2ID != userID {
		util.ErrorResponse(w, http.StatusForbidden, "You are not a player in this game")
		return
	}

	// Verify it's the user's turn
	if game.CurrentTurn != userID {
		util.ErrorResponse(w, http.StatusBadRequest, "Not your turn")
		return
	}

	// Verify game is in progress
	if game.GameStatus != "in_progress" {
		util.ErrorResponse(w, http.StatusBadRequest, "Game is not in progress")
		return
	}

	// Get game state to check if dice already rolled
	state, err := db.GetGameState(r.Context(), gameID)
	if err != nil {
		log.Printf("Failed to get game state: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get game state")
		return
	}

	// Check if dice already rolled
	if state.DiceRoll != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Dice already rolled for this turn")
		return
	}

	// Roll dice
	dice, err := db.RollDice(r.Context(), gameID)
	if err != nil {
		log.Printf("Failed to roll dice: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to roll dice")
		return
	}

	// Get updated state
	state, err = db.GetGameState(r.Context(), gameID)
	if err != nil {
		log.Printf("Failed to get updated state: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get state")
		return
	}

	// Format response
	util.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"stateId":        state.StateID,
		"gameId":         state.GameID,
		"board":          state.BoardState,
		"barWhite":       state.BarWhite,
		"barBlack":       state.BarBlack,
		"bornedOffWhite": state.BornedOffWhite,
		"bornedOffBlack": state.BornedOffBlack,
		"diceRoll":       dice,
		"diceUsed":       state.DiceUsed,
		"lastUpdated":    state.LastUpdated,
	})
}

// MoveRequest represents a move request
type MoveRequest struct {
	FromPoint int `json:"fromPoint"`
	ToPoint   int `json:"toPoint"`
	DieUsed   int `json:"dieUsed"`
}

// MoveHandler executes a checker move
func MoveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := repository.GetDB()
	if db == nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	// Get current user ID from context
	userID, ok := util.GetUserIDFromContext(r.Context())
	if !ok {
		util.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse game ID from URL path
	gameID, err := parseGameIDFromPath(strings.TrimSuffix(r.URL.Path, "/move"))
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Invalid game ID")
		return
	}

	// Parse request body
	var req MoveRequest
	if err := util.ParseJSONBody(r, &req); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get game details
	game, err := db.GetGameByID(r.Context(), gameID)
	if err != nil {
		util.ErrorResponse(w, http.StatusNotFound, "Game not found")
		return
	}

	// Verify user is a player in this game
	if game.Player1ID != userID && game.Player2ID != userID {
		util.ErrorResponse(w, http.StatusForbidden, "You are not a player in this game")
		return
	}

	// Verify it's the user's turn
	if game.CurrentTurn != userID {
		util.ErrorResponse(w, http.StatusBadRequest, "Not your turn")
		return
	}

	// Verify game is in progress
	if game.GameStatus != "in_progress" {
		util.ErrorResponse(w, http.StatusBadRequest, "Game is not in progress")
		return
	}

	// Get game state
	state, err := db.GetGameState(r.Context(), gameID)
	if err != nil {
		log.Printf("Failed to get game state: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get game state")
		return
	}

	// Check if dice have been rolled
	if state.DiceRoll == nil || len(state.DiceRoll) != 2 {
		util.ErrorResponse(w, http.StatusBadRequest, "Dice not rolled yet")
		return
	}

	// Determine player color
	var color business.Color
	if game.Player1ID == userID {
		color = business.Color(game.Player1Color)
	} else {
		color = business.Color(game.Player2Color)
	}

	// Determine bar count
	var barCount int
	if color == business.ColorWhite {
		barCount = state.BarWhite
	} else {
		barCount = state.BarBlack
	}

	// Validate the move
	err = business.ValidateMove(state.BoardState, req.FromPoint, req.ToPoint, req.DieUsed, color, barCount)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Find which die was used
	dieIndex := -1
	for i, die := range state.DiceRoll {
		if die == req.DieUsed && !state.DiceUsed[i] {
			dieIndex = i
			break
		}
	}
	if dieIndex == -1 {
		util.ErrorResponse(w, http.StatusBadRequest, "Die not available or already used")
		return
	}

	// Execute the move
	result, err := business.ExecuteMove(state.BoardState, req.FromPoint, req.ToPoint, color)
	if err != nil {
		log.Printf("Failed to execute move: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to execute move")
		return
	}

	// Update state
	state.BoardState = result.NewBoard
	state.DiceUsed[dieIndex] = true

	// Update bar/borne-off counts
	if req.FromPoint == 0 {
		// Moving from bar
		if color == business.ColorWhite {
			state.BarWhite--
		} else {
			state.BarBlack--
		}
	}

	if req.ToPoint == 25 {
		// Bearing off
		if color == business.ColorWhite {
			state.BornedOffWhite++
		} else {
			state.BornedOffBlack++
		}
	}

	if result.HitOpponent {
		// Opponent checker sent to bar
		if color == business.ColorWhite {
			state.BarBlack++
		} else {
			state.BarWhite++
		}
	}

	// Save updated state
	err = db.UpdateGameState(r.Context(), state)
	if err != nil {
		log.Printf("Failed to update game state: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to update state")
		return
	}

	// Record the move
	moveNumber, err := db.GetLastMoveNumber(r.Context(), gameID)
	if err != nil {
		log.Printf("Failed to get last move number: %v", err)
	} else {
		move := &repository.Move{
			GameID:      gameID,
			PlayerID:    userID,
			MoveNumber:  moveNumber + 1,
			FromPoint:   req.FromPoint,
			ToPoint:     req.ToPoint,
			DieUsed:     req.DieUsed,
			HitOpponent: result.HitOpponent,
		}
		_, err = db.CreateMove(r.Context(), move)
		if err != nil {
			log.Printf("Failed to record move: %v", err)
		}
	}

	// Check for win condition
	var bornedOff int
	if color == business.ColorWhite {
		bornedOff = state.BornedOffWhite
	} else {
		bornedOff = state.BornedOffBlack
	}

	if business.CheckWinCondition(bornedOff) {
		// Player won!
		err = db.CompleteGame(r.Context(), gameID, userID)
		if err != nil {
			log.Printf("Failed to complete game: %v", err)
		}
	} else {
		// Check if turn should end (all dice used or no legal moves)
		if business.AllDiceUsed(state.DiceUsed) || !business.HasLegalMoves(state.BoardState, color, state.DiceRoll, state.DiceUsed, barCount) {
			// End turn: switch to other player and clear dice
			var nextPlayer int
			if game.CurrentTurn == game.Player1ID {
				nextPlayer = game.Player2ID
			} else {
				nextPlayer = game.Player1ID
			}

			err = db.UpdateGameTurn(r.Context(), gameID, nextPlayer)
			if err != nil {
				log.Printf("Failed to update turn: %v", err)
			}

			err = db.ClearDice(r.Context(), gameID)
			if err != nil {
				log.Printf("Failed to clear dice: %v", err)
			}
		}
	}

	// Get updated state
	state, err = db.GetGameState(r.Context(), gameID)
	if err != nil {
		log.Printf("Failed to get updated state: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get state")
		return
	}

	// Format response
	util.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"stateId":        state.StateID,
		"gameId":         state.GameID,
		"board":          state.BoardState,
		"barWhite":       state.BarWhite,
		"barBlack":       state.BarBlack,
		"bornedOffWhite": state.BornedOffWhite,
		"bornedOffBlack": state.BornedOffBlack,
		"diceRoll":       state.DiceRoll,
		"diceUsed":       state.DiceUsed,
		"lastUpdated":    state.LastUpdated,
	})
}

// GetLegalMovesHandler returns all legal moves for the current position
func GetLegalMovesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := repository.GetDB()
	if db == nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	// Get current user ID from context
	userID, ok := util.GetUserIDFromContext(r.Context())
	if !ok {
		util.ErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse game ID from URL path
	gameID, err := parseGameIDFromPath(strings.TrimSuffix(r.URL.Path, "/legal-moves"))
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Invalid game ID")
		return
	}

	// Get game details
	game, err := db.GetGameByID(r.Context(), gameID)
	if err != nil {
		util.ErrorResponse(w, http.StatusNotFound, "Game not found")
		return
	}

	// Verify user is a player in this game
	if game.Player1ID != userID && game.Player2ID != userID {
		util.ErrorResponse(w, http.StatusForbidden, "You are not a player in this game")
		return
	}

	// Get game state
	state, err := db.GetGameState(r.Context(), gameID)
	if err != nil {
		log.Printf("Failed to get game state: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get game state")
		return
	}

	// Check if dice have been rolled
	if state.DiceRoll == nil {
		util.JSONResponse(w, http.StatusOK, map[string]interface{}{
			"moves": []interface{}{},
		})
		return
	}

	// Determine player color
	var color business.Color
	if game.Player1ID == userID {
		color = business.Color(game.Player1Color)
	} else {
		color = business.Color(game.Player2Color)
	}

	// Determine bar and borne-off counts
	var barCount, bornedOff int
	if color == business.ColorWhite {
		barCount = state.BarWhite
		bornedOff = state.BornedOffWhite
	} else {
		barCount = state.BarBlack
		bornedOff = state.BornedOffBlack
	}

	// Get legal moves
	legalMoves := business.GetLegalMoves(state.BoardState, color, state.DiceRoll, state.DiceUsed, barCount, bornedOff)

	// Format response
	movesList := []map[string]interface{}{}
	for _, move := range legalMoves {
		movesList = append(movesList, map[string]interface{}{
			"fromPoint": move.FromPoint,
			"toPoint":   move.ToPoint,
			"dieUsed":   move.DieUsed,
		})
	}

	util.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"moves": movesList,
	})
}
