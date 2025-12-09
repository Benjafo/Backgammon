package service

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"backgammon/repository"
	"backgammon/util"
)

// Generate a CSRF token for registration
func RegisterTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := repository.GetDB()
	if db == nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	// Generate custom registration token with embedded IP, User-Agent, and timestamp
	clientIP := util.GetClientIP(r)
	token, err := util.GenerateRegistrationToken(clientIP, r.UserAgent(), time.Now())
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Token expires in 15 minutes
	expiresAt := time.Now().Add(15 * time.Minute)

	// Store token in database
	err = db.CreateRegistrationToken(r.Context(), token, clientIP, r.UserAgent(), expiresAt)
	if err != nil {
		log.Printf("Failed to create registration token: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to create token")
		return
	}

	util.JSONResponse(w, http.StatusOK, map[string]string{
		"token": token,
	})
}

// Create a new user account
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := repository.GetDB()
	if db == nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	var req RegisterRequest
	if err := util.ParseJSONBody(r, &req); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		util.ErrorResponse(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	if len(req.Username) < 3 {
		util.ErrorResponse(w, http.StatusBadRequest, "Username must be at least 3 characters")
		return
	}

	if len(req.Password) < 6 {
		util.ErrorResponse(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	// Validate registration token
	if req.Token == "" {
		util.ErrorResponse(w, http.StatusBadRequest, "Registration token is required")
		return
	}

	// Validate token structure and embedded data
	clientIP := util.GetClientIP(r)
	_, err := util.ValidateRegistrationTokenStructure(req.Token, clientIP, r.UserAgent())
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		util.ErrorResponse(w, http.StatusUnauthorized, fmt.Sprintf("Invalid registration token: %v", err))
		return
	}

	// Check token in database
	if err := db.ValidateAndUseRegistrationToken(r.Context(), req.Token); err != nil {
		log.Printf("Token database validation failed: %v", err)
		util.ErrorResponse(w, http.StatusUnauthorized, "Invalid or expired registration token")
		return
	}

	// Check if username already exists
	existingUser, _ := db.GetUserByUsername(r.Context(), req.Username)
	if existingUser != nil {
		util.ErrorResponse(w, http.StatusConflict, "Username already exists")
		return
	}

	// Hash password
	passwordHash, err := util.HashPassword(req.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	// Create user
	userID, err := db.CreateUser(r.Context(), req.Username, passwordHash)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	// Create session automatically after registration
	sessionToken, err := util.GenerateSecureToken(32)
	if err != nil {
		log.Printf("Failed to generate session token: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Registration successful but login failed")
		return
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days

	err = db.CreateSession(r.Context(), userID, sessionToken, r.RemoteAddr, r.UserAgent(), expiresAt)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Registration successful but login failed")
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})

	util.JSONResponse(w, http.StatusCreated, map[string]interface{}{
		"message": "Registration successful",
		"user": UserResponse{
			ID:       userID,
			Username: req.Username,
		},
	})
}

// Authenticate a user and creates a session
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := repository.GetDB()
	if db == nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	var req LoginRequest
	if err := util.ParseJSONBody(r, &req); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		util.ErrorResponse(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Get user from database
	user, err := db.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		util.ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check password
	if err := util.CheckPassword(user.PasswordHash, req.Password); err != nil {
		util.ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate session token
	sessionToken, err := util.GenerateSecureToken(32)
	if err != nil {
		log.Printf("Failed to generate session token: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Login failed")
		return
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days

	// Create session
	err = db.CreateSession(r.Context(), user.UserID, sessionToken, r.RemoteAddr, r.UserAgent(), expiresAt)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Login failed")
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})

	util.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"user": UserResponse{
			ID:       user.UserID,
			Username: user.Username,
		},
	})
}

// Invalidate the current session
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := repository.GetDB()
	if db == nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	// Get session cookie
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		util.ErrorResponse(w, http.StatusUnauthorized, "No active session")
		return
	}

	// Delete session from database
	if err := db.DeleteSession(r.Context(), cookie.Value); err != nil {
		log.Printf("Failed to delete session: %v", err)
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	util.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Logout successful",
	})
}

// Validate the current session and returns user info
func SessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		util.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	db := repository.GetDB()
	if db == nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "Database not initialized")
		return
	}

	// Get session cookie
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		util.ErrorResponse(w, http.StatusUnauthorized, "No active session")
		return
	}

	// Validate session
	session, err := db.GetSessionByToken(r.Context(), cookie.Value)
	if err != nil {
		util.ErrorResponse(w, http.StatusUnauthorized, "Invalid or expired session")
		return
	}

	// Get user details
	user, err := db.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		util.ErrorResponse(w, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	util.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"user": UserResponse{
			ID:       user.UserID,
			Username: user.Username,
		},
	})
}
