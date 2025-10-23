package main

import (
	"net/http"

	"backgammon/service"
	"backgammon/util"
)

func main() {
	mux := http.NewServeMux()

	// Authentication endpoints
	mux.HandleFunc("/api/v1/auth/login", service.LoginHandler)
	mux.HandleFunc("/api/v1/auth/logout", service.LogoutHandler)
	mux.HandleFunc("/api/v1/auth/register", service.RegisterHandler)
	mux.HandleFunc("/api/v1/auth/register/token", service.RegisterTokenHandler)
	mux.HandleFunc("/api/v1/auth/session", service.SessionHandler)

	// Lobby endpoints
	mux.HandleFunc("/api/v1/lobby/users", service.LobbyUsersHandler)
	mux.HandleFunc("/api/v1/lobby/presence", service.LobbyPresenceHandler)
	mux.HandleFunc("/api/v1/lobby/presence/heartbeat", service.LobbyPresenceHeartbeatHandler)

	// Invitation endpoints
	mux.HandleFunc("/api/v1/invitations", service.InvitationsHandler)
	mux.HandleFunc("/api/v1/invitations/{:id}", service.CancelInvitationHandler)
	mux.HandleFunc("/api/v1/invitations/{:id}/accept", service.AcceptInvitationHandler)
	mux.HandleFunc("/api/v1/invitations/{:id}/decline", service.DeclineInvitationHandler)

	// Game endpoints
	mux.HandleFunc("/api/v1/games/active", service.ActiveGamesHandler)
	mux.HandleFunc("/api/v1/games/{:id}", service.GameHandler)
	mux.HandleFunc("/api/v1/games/{:id}/state", service.GameStateHandler)
	mux.HandleFunc("/api/v1/games/{:id}/roll", service.RollDiceHandler)
	mux.HandleFunc("/api/v1/games/{:id}/moves", service.MoveHandler)
	mux.HandleFunc("/api/v1/games/{:id}/forfeit", service.ForfeitHandler)

	// Chat endpoints
	mux.HandleFunc("/api/v1/chat/rooms/{:roomId}/messages", service.ChatMessagesHandler)

	// Using var keyword to specify type explicitly
	var fs http.Handler
	fs = http.FileServer(http.Dir("./static/"))
	mux.Handle("/", fs)

	// Protect all private routes with authentication middleware
	protected := util.SessionMiddleware(mux)
	
	http.ListenAndServe("localhost:8080", protected)
}
