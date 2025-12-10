package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/time/rate"

	"backgammon/middleware"
	"backgammon/repository"
	"backgammon/service"
	"backgammon/util"
)

var db *repository.Postgres

var (
	authLimiter = middleware.NewRateLimiter(rate.Every(12*time.Second), 15) // Auth: 15 requests per minute
	gameLimiter = middleware.NewRateLimiter(rate.Every(2*time.Second), 30) // Game: 30 requests per minute
	readLimiter = middleware.NewRateLimiter(rate.Every(time.Second), 60) // Reads: 60 requests per minute
)

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

	// Ensure lobby chat room exists
	roomID, err := db.EnsureLobbyRoomExists(context.Background())
	if err != nil {
		log.Fatalf("Failed to create lobby chat room: %v", err)
	}
	log.Printf("Lobby chat room initialized (ID: %d)", roomID)

	// Initialize WebSocket hub for chat
	chatHub := service.NewHub()
	go chatHub.Run()
	log.Println("WebSocket hub initialized and running")

	// Routers
	mux := http.NewServeMux()          // Unprotected endpoints
	protectedMux := http.NewServeMux() // Protected endpoints (require authentication)

	// Public endpoints
	mux.HandleFunc("/api/v1/health", service.HealthCheckHandler)
	mux.HandleFunc("/api/v1/auth/login", authLimiter.Limit(service.LoginHandler))
	mux.HandleFunc("/api/v1/auth/register", authLimiter.Limit(service.RegisterHandler))
	mux.HandleFunc("/api/v1/auth/register/token", authLimiter.Limit(service.RegisterTokenHandler))

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
	protectedMux.HandleFunc("/api/v1/games/", func(w http.ResponseWriter, r *http.Request) {
		// Route to game chat WebSocket if path ends with /ws
		if len(r.URL.Path) > 3 && r.URL.Path[len(r.URL.Path)-3:] == "/ws" {
			service.GameChatWebSocketHandler(chatHub)(w, r)
		} else {
			service.GameRouterHandler(w, r)
		}
	})

	// Chat endpoints
	protectedMux.HandleFunc("/api/v1/lobby/ws", service.ChatWebSocketHandler(chatHub))
	// protectedMux.HandleFunc("/api/v1/chat/rooms/{:roomId}/messages", service.ChatMessagesHandler)

	// Apply session middleware to protected routes
	protected := util.SessionMiddleware(protectedMux)
	mux.Handle("/api/", protected)

	// Serve Swagger docs
	swagger := http.FileServer(http.Dir("./static/swagger/"))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", swagger))

	// Serve React app (built frontend) with SPA fallback
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve index.html for the root path
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./static/dist/index.html")
			return
		}

		// Try to serve the requested file
		filePath := "./static/dist" + r.URL.Path
		if _, err := os.Stat(filePath); err == nil {
			http.ServeFile(w, r, filePath)
			return
		}

		// If file doesn't exist, serve index.html (SPA fallback)
		http.ServeFile(w, r, "./static/dist/index.html")
	})

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
