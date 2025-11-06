package service

// ============================================================================
// Auth & User Types
// ============================================================================

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"` // Registration CSRF token
}

type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// ============================================================================
// Game Types
// ============================================================================

type MoveRequest struct {
	FromPoint      int   `json:"fromPoint"`
	ToPoint        int   `json:"toPoint"`
	DieUsed        int   `json:"dieUsed"`
	DiceIndices    []int `json:"diceIndices"`    // Indices of dice being used (for combined moves)
	IsCombinedMove bool  `json:"isCombinedMove"` // True if using multiple dice
}

// ============================================================================
// Invitation Types
// ============================================================================

type CreateInvitationRequest struct {
	ChallengedID int `json:"challengedId"`
}
