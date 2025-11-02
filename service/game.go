package service

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"backgammon/repository"
	"backgammon/util"
)

// GameRouterHandler routes game requests to the appropriate handler
func GameRouterHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// /api/v1/games/{id} - GET
	if r.Method == http.MethodGet && !strings.Contains(path, "/forfeit") && !strings.Contains(path, "/start") {
		GameHandler(w, r)
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

// ActiveGamesHandler returns active games for the current user (stub)
func ActiveGamesHandler(w http.ResponseWriter, r *http.Request) {
	util.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"games": []interface{}{},
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

// Placeholder handlers for future implementation
func GameStateHandler(w http.ResponseWriter, r *http.Request) {
	util.ErrorResponse(w, http.StatusNotImplemented, "Not implemented yet")
}

func RollDiceHandler(w http.ResponseWriter, r *http.Request) {
	util.ErrorResponse(w, http.StatusNotImplemented, "Not implemented yet")
}

func MoveHandler(w http.ResponseWriter, r *http.Request) {
	util.ErrorResponse(w, http.StatusNotImplemented, "Not implemented yet")
}
