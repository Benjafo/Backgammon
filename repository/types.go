package repository

import "time"

// ============================================================================
// User Types
// ============================================================================

type User struct {
	UserID       int
	Username     string
	PasswordHash string
	Email        *string
}

// ============================================================================
// Session Types
// ============================================================================

type Session struct {
	SessionID    int
	UserID       int
	SessionToken string
	IPAddress    string
	UserAgent    string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	IsActive     bool
}

// ============================================================================
// Registration Token Types
// ============================================================================

type RegistrationToken struct {
	TokenID    int
	TokenValue string
	IPAddress  string
	UserAgent  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	IsUsed     bool
}

// ============================================================================
// Game Types
// ============================================================================

type Game struct {
	GameID       int
	Player1ID    int
	Player2ID    int
	CurrentTurn  int
	GameStatus   string
	WinnerID     *int
	CreatedAt    time.Time
	StartedAt    *time.Time
	EndedAt      *time.Time
	Player1Color string
	Player2Color string
}

type GameWithPlayers struct {
	GameID          int
	Player1ID       int
	Player1Username string
	Player1Color    string
	Player2ID       int
	Player2Username string
	Player2Color    string
	CurrentTurn     int
	GameStatus      string
	WinnerID        *int
	CreatedAt       time.Time
	StartedAt       *time.Time
	EndedAt         *time.Time
}

type GameState struct {
	StateID        int
	GameID         int
	BoardState     []int // 24 integers: positive=white, negative=black, 0=empty
	BarWhite       int
	BarBlack       int
	BornedOffWhite int
	BornedOffBlack int
	DiceRoll       []int  // [die1, die2] or nil
	DiceUsed       []bool // [used1, used2] or nil
	LastUpdated    time.Time
}

type Move struct {
	MoveID      int
	GameID      int
	PlayerID    int
	MoveNumber  int
	FromPoint   int // 0=bar, 1-24=board points, 25=borne off
	ToPoint     int
	DieUsed     int
	HitOpponent bool
	Timestamp   time.Time
}

// ============================================================================
// Invitation Types
// ============================================================================

type Invitation struct {
	InvitationID int
	ChallengerID int
	ChallengedID int
	Status       string
	GameID       *int
	CreatedAt    time.Time
	// Extended fields for joined queries
	ChallengerUsername string
	ChallengedUsername string
}

type InvitationWithUsers struct {
	InvitationID       int
	ChallengerID       int
	ChallengerUsername string
	ChallengedID       int
	ChallengedUsername string
	Status             string
	GameID             *int
	CreatedAt          time.Time
}

// ============================================================================
// Lobby Types
// ============================================================================

type LobbyUser struct {
	UserID        int
	Username      string
	JoinedAt      time.Time
	LastHeartbeat time.Time
}
