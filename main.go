package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"backgammon/repository"
	"backgammon/service"
	"backgammon/util"
)

var db *repository.Postgres

func main() {
	// Initialize database connection
	connString := os.Getenv("DATABASE_URL")
	db, err := repository.NewPG(context.Background(), connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ping database to check connection
	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established successfully")

	// Routers
	mux := http.NewServeMux()          // Unprotected endpoints
	protectedMux := http.NewServeMux() // Protected endpoints (require authentication)

	// Public endpoints
	mux.HandleFunc("/api/v1/health", service.HealthCheckHandler)
	mux.HandleFunc("/api/v1/auth/login", service.LoginHandler)
	mux.HandleFunc("/api/v1/auth/register", service.RegisterHandler)
	mux.HandleFunc("/api/v1/auth/register/token", service.RegisterTokenHandler)

	// Protected auth endpoints
	protectedMux.HandleFunc("/api/v1/auth/logout", service.LogoutHandler)
	protectedMux.HandleFunc("/api/v1/auth/session", service.SessionHandler)

	// Lobby endpoints
	protectedMux.HandleFunc("/api/v1/lobby/users", service.LobbyUsersHandler)
	protectedMux.HandleFunc("/api/v1/lobby/presence", service.LobbyPresenceHandler)
	protectedMux.HandleFunc("/api/v1/lobby/presence/heartbeat", service.LobbyPresenceHeartbeatHandler)

	// Invitation endpoints
	protectedMux.HandleFunc("/api/v1/invitations", service.InvitationsHandler)
	// protectedMux.HandleFunc("/api/v1/invitations/{:id}", service.CancelInvitationHandler)
	// protectedMux.HandleFunc("/api/v1/invitations/{:id}/accept", service.AcceptInvitationHandler)
	// protectedMux.HandleFunc("/api/v1/invitations/{:id}/decline", service.DeclineInvitationHandler)

	// Game endpoints
	protectedMux.HandleFunc("/api/v1/games/active", service.ActiveGamesHandler)
	// protectedMux.HandleFunc("/api/v1/games/{:id}", service.GameHandler)
	// protectedMux.HandleFunc("/api/v1/games/{:id}/state", service.GameStateHandler)
	// protectedMux.HandleFunc("/api/v1/games/{:id}/roll", service.RollDiceHandler)
	// protectedMux.HandleFunc("/api/v1/games/{:id}/moves", service.MoveHandler)
	// protectedMux.HandleFunc("/api/v1/games/{:id}/forfeit", service.ForfeitHandler)

	// Chat endpoint
	// protectedMux.HandleFunc("/api/v1/chat/rooms/{:roomId}/messages", service.ChatMessagesHandler)

	// Apply session middleware to protected routes
	protected := util.SessionMiddleware(protectedMux)
	mux.Handle("/api/", protected)

	// Serve static files
	fs := http.FileServer(http.Dir("./static/"))
	mux.Handle("/", fs)

	log.Println("Server starting on :8080")
	http.ListenAndServe("0.0.0.0:8080", mux)
}
