package service

import (
	"log"
	"net/http"

	"backgammon/repository"
	"backgammon/util"
)

// Return the list of users currently in the lobby
func LobbyUsersHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get all lobby users
	users, err := db.GetLobbyUsers(r.Context())
	if err != nil {
		log.Printf("Failed to get lobby users: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get lobby users")
		return
	}

	// Filter out current user from the list
	var filteredUsers []map[string]interface{}
	for _, user := range users {
		if user.UserID != userID {
			filteredUsers = append(filteredUsers, map[string]interface{}{
				"userId":        user.UserID,
				"username":      user.Username,
				"joinedAt":      user.JoinedAt,
				"lastHeartbeat": user.LastHeartbeat,
			})
		}
	}

	// Handle nil slice
	if filteredUsers == nil {
		filteredUsers = []map[string]interface{}{}
	}

	util.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"users": filteredUsers,
		"count": len(filteredUsers),
	})
}

// Handle joining and leaving the lobby
func LobbyPresenceHandler(w http.ResponseWriter, r *http.Request) {
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

	switch r.Method {
	case http.MethodPost:
		// Join lobby (idempotent - updates heartbeat if already in lobby)
		presenceID, err := db.JoinLobby(r.Context(), userID)
		if err != nil {
			log.Printf("Failed to join lobby: %v", err)
			util.ErrorResponse(w, http.StatusInternalServerError, "Failed to join lobby")
			return
		}

		util.JSONResponse(w, http.StatusCreated, map[string]interface{}{
			"message":    "Joined lobby successfully",
			"presenceId": presenceID,
		})

	case http.MethodDelete:
		// Leave lobby
		err := db.LeaveLobby(r.Context(), userID)
		if err != nil {
			log.Printf("Failed to leave lobby: %v", err)
			util.ErrorResponse(w, http.StatusInternalServerError, "Failed to leave lobby")
			return
		}

		util.JSONResponse(w, http.StatusOK, map[string]string{
			"message": "Left lobby successfully",
		})

	default:
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Update the user's last heartbeat timestamp
func LobbyPresenceHeartbeatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
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

	// Update heartbeat
	err := db.UpdateHeartbeat(r.Context(), userID)
	if err != nil {
		if err.Error() == "user not in lobby" {
			util.ErrorResponse(w, http.StatusNotFound, "User not in lobby")
			return
		}
		log.Printf("Failed to update heartbeat: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to update heartbeat")
		return
	}

	util.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Heartbeat updated",
	})
}
