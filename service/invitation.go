package service

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"backgammon/repository"
	"backgammon/util"
)

type CreateInvitationRequest struct {
	ChallengedID int `json:"challengedId"`
}

// InvitationRouterHandler routes invitation requests to the appropriate handler
func InvitationRouterHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// /api/v1/invitations - GET/POST
	if path == "/api/v1/invitations" {
		InvitationsHandler(w, r)
		return
	}

	// /api/v1/invitations/{id}/accept - PUT
	if strings.HasSuffix(path, "/accept") && r.Method == http.MethodPut {
		AcceptInvitationHandler(w, r)
		return
	}

	// /api/v1/invitations/{id}/decline - PUT
	if strings.HasSuffix(path, "/decline") && r.Method == http.MethodPut {
		DeclineInvitationHandler(w, r)
		return
	}

	// /api/v1/invitations/{id} - DELETE
	if r.Method == http.MethodDelete {
		CancelInvitationHandler(w, r)
		return
	}

	util.ErrorResponse(w, http.StatusNotFound, "Endpoint not found")
}

// InvitationsHandler handles GET (list) and POST (create) for invitations
func InvitationsHandler(w http.ResponseWriter, r *http.Request) {
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
	case http.MethodGet:
		handleGetInvitations(w, r, db, userID)
	case http.MethodPost:
		handleCreateInvitation(w, r, db, userID)
	default:
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleGetInvitations retrieves all invitations for the current user
func handleGetInvitations(w http.ResponseWriter, r *http.Request, db *repository.Postgres, userID int) {
	sent, received, err := db.GetInvitationsByUser(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get invitations: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get invitations")
		return
	}

	// Format sent invitations
	sentList := []map[string]interface{}{}
	for _, inv := range sent {
		sentList = append(sentList, map[string]interface{}{
			"invitationId": inv.InvitationID,
			"challenger": map[string]interface{}{
				"userId":   inv.ChallengerID,
				"username": inv.ChallengerUsername,
			},
			"challenged": map[string]interface{}{
				"userId":   inv.ChallengedID,
				"username": inv.ChallengedUsername,
			},
			"status":    inv.Status,
			"gameId":    inv.GameID,
			"createdAt": inv.CreatedAt,
		})
	}

	// Format received invitations
	receivedList := []map[string]interface{}{}
	for _, inv := range received {
		receivedList = append(receivedList, map[string]interface{}{
			"invitationId": inv.InvitationID,
			"challenger": map[string]interface{}{
				"userId":   inv.ChallengerID,
				"username": inv.ChallengerUsername,
			},
			"challenged": map[string]interface{}{
				"userId":   inv.ChallengedID,
				"username": inv.ChallengedUsername,
			},
			"status":    inv.Status,
			"gameId":    inv.GameID,
			"createdAt": inv.CreatedAt,
		})
	}

	util.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"sent":     sentList,
		"received": receivedList,
	})
}

// handleCreateInvitation creates a new invitation
func handleCreateInvitation(w http.ResponseWriter, r *http.Request, db *repository.Postgres, userID int) {
	var req CreateInvitationRequest
	if err := util.ParseJSONBody(r, &req); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate challenged ID is provided
	if req.ChallengedID == 0 {
		util.ErrorResponse(w, http.StatusBadRequest, "challengedId is required")
		return
	}

	// Validate not challenging self
	if req.ChallengedID == userID {
		util.ErrorResponse(w, http.StatusBadRequest, "Cannot challenge yourself")
		return
	}

	// Verify challenged user exists
	challengedUser, err := db.GetUserByID(r.Context(), req.ChallengedID)
	if err != nil {
		util.ErrorResponse(w, http.StatusNotFound, "Challenged user not found")
		return
	}

	// Verify challenged user is in lobby
	inLobby, err := db.IsUserInLobby(r.Context(), req.ChallengedID)
	if err != nil {
		log.Printf("Failed to check if user in lobby: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to create invitation")
		return
	}
	if !inLobby {
		util.ErrorResponse(w, http.StatusNotFound, "Challenged user not in lobby")
		return
	}

	// Create invitation
	invitationID, err := db.CreateInvitation(r.Context(), userID, req.ChallengedID)
	if err != nil {
		if strings.Contains(err.Error(), "pending invitation already exists") {
			util.ErrorResponse(w, http.StatusConflict, "Pending invitation already exists")
			return
		}
		log.Printf("Failed to create invitation: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to create invitation")
		return
	}

	util.JSONResponse(w, http.StatusCreated, map[string]interface{}{
		"invitationId": invitationID,
		"challengedId": req.ChallengedID,
		"status":       "pending",
		"message":      "Invitation sent successfully",
	})

	_ = challengedUser // Suppress unused variable warning
}

// AcceptInvitationHandler handles accepting an invitation
func AcceptInvitationHandler(w http.ResponseWriter, r *http.Request) {
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

	// Parse invitation ID from URL path
	// Expected format: /api/v1/invitations/{id}/accept
	invitationID, err := parseInvitationIDFromPath(r.URL.Path, "/api/v1/invitations/", "/accept")
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Invalid invitation ID")
		return
	}

	// Get invitation details
	invitation, err := db.GetInvitationByID(r.Context(), invitationID)
	if err != nil {
		util.ErrorResponse(w, http.StatusNotFound, "Invitation not found")
		return
	}

	// Verify user is the challenged party
	if invitation.ChallengedID != userID {
		util.ErrorResponse(w, http.StatusBadRequest, "You are not the challenged party")
		return
	}

	// Verify invitation is pending
	if invitation.Status != "pending" {
		util.ErrorResponse(w, http.StatusBadRequest, "Invitation already processed")
		return
	}

	// Create game
	gameID, err := db.CreateGame(r.Context(), invitation.ChallengerID, invitation.ChallengedID)
	if err != nil {
		log.Printf("Failed to create game: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to create game")
		return
	}

	// Accept invitation and link to game
	err = db.AcceptInvitation(r.Context(), invitationID, gameID)
	if err != nil {
		log.Printf("Failed to accept invitation: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to accept invitation")
		return
	}

	// Remove both users from lobby (they're now in a game)
	_ = db.LeaveLobby(r.Context(), invitation.ChallengerID)
	_ = db.LeaveLobby(r.Context(), invitation.ChallengedID)

	util.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Invitation accepted",
		"gameId":  gameID,
	})
}

// DeclineInvitationHandler handles declining an invitation
func DeclineInvitationHandler(w http.ResponseWriter, r *http.Request) {
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

	// Parse invitation ID from URL path
	// Expected format: /api/v1/invitations/{id}/decline
	invitationID, err := parseInvitationIDFromPath(r.URL.Path, "/api/v1/invitations/", "/decline")
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Invalid invitation ID")
		return
	}

	// Get invitation details
	invitation, err := db.GetInvitationByID(r.Context(), invitationID)
	if err != nil {
		util.ErrorResponse(w, http.StatusNotFound, "Invitation not found")
		return
	}

	// Verify user is the challenged party
	if invitation.ChallengedID != userID {
		util.ErrorResponse(w, http.StatusBadRequest, "You are not the challenged party")
		return
	}

	// Verify invitation is pending
	if invitation.Status != "pending" {
		util.ErrorResponse(w, http.StatusBadRequest, "Invitation already processed")
		return
	}

	// Decline invitation
	err = db.DeclineInvitation(r.Context(), invitationID)
	if err != nil {
		log.Printf("Failed to decline invitation: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to decline invitation")
		return
	}

	util.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Invitation declined",
	})
}

// CancelInvitationHandler handles canceling an invitation (challenger only)
func CancelInvitationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
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

	// Parse invitation ID from URL path
	// Expected format: /api/v1/invitations/{id}
	invitationID, err := parseInvitationIDFromPath(r.URL.Path, "/api/v1/invitations/", "")
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Invalid invitation ID")
		return
	}

	// Get invitation details
	invitation, err := db.GetInvitationByID(r.Context(), invitationID)
	if err != nil {
		util.ErrorResponse(w, http.StatusNotFound, "Invitation not found")
		return
	}

	// Verify user is the challenger
	if invitation.ChallengerID != userID {
		util.ErrorResponse(w, http.StatusBadRequest, "You are not the challenger")
		return
	}

	// Verify invitation is pending
	if invitation.Status != "pending" {
		util.ErrorResponse(w, http.StatusBadRequest, "Invitation already processed")
		return
	}

	// Cancel invitation
	err = db.CancelInvitation(r.Context(), invitationID)
	if err != nil {
		log.Printf("Failed to cancel invitation: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to cancel invitation")
		return
	}

	util.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Invitation cancelled",
	})
}

// parseInvitationIDFromPath extracts the invitation ID from the URL path
// Example: /api/v1/invitations/42/accept -> returns 42
func parseInvitationIDFromPath(path, prefix, suffix string) (int, error) {
	// Remove prefix and suffix
	trimmed := strings.TrimPrefix(path, prefix)
	if suffix != "" {
		trimmed = strings.TrimSuffix(trimmed, suffix)
	}

	// Parse the ID
	id, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, err
	}

	return id, nil
}
