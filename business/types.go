package business

// ============================================================================
// Domain Types
// ============================================================================

// Color represents a player's color
type Color string

const (
	ColorWhite Color = "white"
	ColorBlack Color = "black"
)

// LegalMove represents a valid move option
type LegalMove struct {
	FromPoint      int   `json:"fromPoint"`      // 0=bar, 1-24=board points, 25=bear off
	ToPoint        int   `json:"toPoint"`        // Destination point
	DieUsed        int   `json:"dieUsed"`        // Sum of dice values used (e.g., 4 for a 1+3 combined move)
	DiceIndices    []int `json:"diceIndices"`    // Indices of dice being used (for combined moves)
	IsCombinedMove bool  `json:"isCombinedMove"` // True if this move uses multiple dice
}

// MoveResult contains the outcome of executing a move
type MoveResult struct {
	NewBoard    []int
	HitOpponent bool
}

// indexedDie represents a die value with its index in the dice array
type indexedDie struct {
	value int
	index int
}
