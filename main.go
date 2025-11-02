package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

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
	protectedMux.HandleFunc("/api/v1/invitations", service.InvitationRouterHandler)
	protectedMux.HandleFunc("/api/v1/invitations/", service.InvitationRouterHandler)

	// Game endpoints
	protectedMux.HandleFunc("/api/v1/games/active", service.ActiveGamesHandler)
	protectedMux.HandleFunc("/api/v1/games/", service.GameRouterHandler)

	// Chat endpoint
	// protectedMux.HandleFunc("/api/v1/chat/rooms/{:roomId}/messages", service.ChatMessagesHandler)

	// Apply session middleware to protected routes
	protected := util.SessionMiddleware(protectedMux)
	mux.Handle("/api/", protected)

	// Serve static files
	fs := http.FileServer(http.Dir("./static/"))
	mux.Handle("/", fs)


	// TODO move cleanup jobs to a separate service
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		log.Println("Started stale lobby presence cleanup job (runs every 30s)")
		for range ticker.C {
			count, err := db.CleanupStaleLobbyPresence(context.Background(), 60*time.Second)
			if err != nil {
				log.Printf("Failed to cleanup stale lobby presence: %v", err)
			} else if count > 0 {
				log.Printf("Removed %d stale lobby presence records", count)
			}
		}
	}()
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		log.Println("Started expired invitation cleanup job (runs every 60s)")
		for range ticker.C {
			count, err := db.CleanupExpiredInvitations(context.Background(), 5*time.Minute)
			if err != nil {
				log.Printf("Failed to cleanup expired invitations: %v", err)
			} else if count > 0 {
				log.Printf("Marked %d invitations as expired", count)
			}
		}
	}()

	log.Println("Server starting on :8080")
	http.ListenAndServe("0.0.0.0:8080", mux)
}
