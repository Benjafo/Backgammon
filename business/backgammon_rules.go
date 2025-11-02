package business

import (
	"fmt"
)

// Color represents a player's color
type Color string

const (
	ColorWhite Color = "white"
	ColorBlack Color = "black"
)

// LegalMove represents a valid move option
type LegalMove struct {
	FromPoint int // 0=bar, 1-24=board points, 25=bear off
	ToPoint   int
	DieUsed   int
}

// MoveResult contains the outcome of executing a move
type MoveResult struct {
	NewBoard    []int
	HitOpponent bool
}

// ============================================================================
// Board Analysis
// ============================================================================

// IsPointOpen checks if a point can be moved to by a specific color
// A point is open if it has 0 or 1 opponent checker, or any number of own checkers
func IsPointOpen(board []int, point int, color Color) bool {
	if point < 1 || point > 24 {
		return false // Invalid point
	}

	checkers := board[point-1] // Convert to 0-indexed

	if color == ColorWhite {
		// White can move to: empty points, white points, or black points with 1 checker
		return checkers >= -1
	} else {
		// Black can move to: empty points, black points, or white points with 1 checker
		return checkers <= 1
	}
}

// IsInHomeBoard checks if a point is in the player's home board
func IsInHomeBoard(point int, color Color) bool {
	if color == ColorWhite {
		return point >= 1 && point <= 6
	} else {
		return point >= 19 && point <= 24
	}
}

// CanBearOff checks if a player can bear off (all checkers in home board)
func CanBearOff(board []int, color Color, barCount int) bool {
	if barCount > 0 {
		return false // Can't bear off with checkers on bar
	}

	if color == ColorWhite {
		// Check if all white checkers are in points 1-6
		for i := 6; i < 24; i++ {
			if board[i] > 0 {
				return false
			}
		}
		return true
	} else {
		// Check if all black checkers are in points 19-24
		for i := 0; i < 18; i++ {
			if board[i] < 0 {
				return false
			}
		}
		return true
	}
}

// CountCheckersOnPoint counts checkers of a specific color on a point
func CountCheckersOnPoint(board []int, point int, color Color) int {
	if point < 1 || point > 24 {
		return 0
	}

	checkers := board[point-1]
	if color == ColorWhite {
		if checkers > 0 {
			return checkers
		}
	} else {
		if checkers < 0 {
			return -checkers
		}
	}
	return 0
}

// ============================================================================
// Move Calculation
// ============================================================================

// CalculateToPoint calculates the destination point for a move
// White moves from high numbers to low (24 -> 1)
// Black moves from low numbers to high (1 -> 24)
func CalculateToPoint(fromPoint int, dieValue int, color Color) int {
	if color == ColorWhite {
		return fromPoint - dieValue
	} else {
		return fromPoint + dieValue
	}
}

// ============================================================================
// Move Validation
// ============================================================================

// ValidateMove checks if a move is legal
func ValidateMove(board []int, fromPoint, toPoint, dieValue int, color Color, barCount int) error {
	// Must enter from bar first
	if barCount > 0 && fromPoint != 0 {
		return fmt.Errorf("must enter from bar first")
	}

	// Check if moving from bar
	if fromPoint == 0 {
		if barCount == 0 {
			return fmt.Errorf("no checkers on bar")
		}
		// Entering from bar: toPoint must match die value from correct end
		expectedPoint := 0
		if color == ColorWhite {
			expectedPoint = 25 - dieValue // White enters from 24 end
		} else {
			expectedPoint = dieValue // Black enters from 1 end
		}
		if toPoint != expectedPoint {
			return fmt.Errorf("invalid entry point from bar")
		}
		if !IsPointOpen(board, toPoint, color) {
			return fmt.Errorf("entry point is blocked")
		}
		return nil
	}

	// Check if bearing off
	if toPoint == 25 {
		if !CanBearOff(board, color, barCount) {
			return fmt.Errorf("cannot bear off yet")
		}
		// Must have checker on fromPoint
		if CountCheckersOnPoint(board, fromPoint, color) == 0 {
			return fmt.Errorf("no checker on source point")
		}
		// Check if exact roll or highest point
		expectedTo := CalculateToPoint(fromPoint, dieValue, color)
		if expectedTo == 0 {
			// Exact bear off
			return nil
		}
		if expectedTo < 0 || expectedTo > 24 {
			// Bearing off with higher die than needed - must be from highest occupied point
			if !isHighestOccupiedPoint(board, fromPoint, color) {
				return fmt.Errorf("must bear off from highest occupied point")
			}
			return nil
		}
		return fmt.Errorf("cannot bear off from this point")
	}

	// Regular move
	if fromPoint < 1 || fromPoint > 24 || toPoint < 1 || toPoint > 24 {
		return fmt.Errorf("invalid point numbers")
	}

	// Must have checker on fromPoint
	if CountCheckersOnPoint(board, fromPoint, color) == 0 {
		return fmt.Errorf("no checker on source point")
	}

	// Check if toPoint matches die value
	expectedTo := CalculateToPoint(fromPoint, dieValue, color)
	if expectedTo != toPoint {
		return fmt.Errorf("destination doesn't match die value")
	}

	// Check if destination is open
	if !IsPointOpen(board, toPoint, color) {
		return fmt.Errorf("destination point is blocked")
	}

	return nil
}

// isHighestOccupiedPoint checks if this is the highest occupied point for bearing off
func isHighestOccupiedPoint(board []int, point int, color Color) bool {
	if color == ColorWhite {
		// For white, check if there are any checkers on higher points (7-24 or point+1 to 6)
		for i := point; i <= 6; i++ {
			if i != point && board[i-1] > 0 {
				return false
			}
		}
		return true
	} else {
		// For black, check if there are any checkers on higher points (19-point-1)
		for i := 19; i < point; i++ {
			if board[i-1] < 0 {
				return false
			}
		}
		return true
	}
}

// ============================================================================
// Move Execution
// ============================================================================

// ExecuteMove applies a move to the board and returns the new state
func ExecuteMove(board []int, fromPoint, toPoint int, color Color) (*MoveResult, error) {
	newBoard := make([]int, 24)
	copy(newBoard, board)

	hitOpponent := false

	// Moving from bar
	if fromPoint == 0 {
		// Add checker to destination
		if color == ColorWhite {
			// Check if hitting opponent
			if newBoard[toPoint-1] == -1 {
				hitOpponent = true
				newBoard[toPoint-1] = 0
			}
			newBoard[toPoint-1]++
		} else {
			if newBoard[toPoint-1] == 1 {
				hitOpponent = true
				newBoard[toPoint-1] = 0
			}
			newBoard[toPoint-1]--
		}
		return &MoveResult{NewBoard: newBoard, HitOpponent: hitOpponent}, nil
	}

	// Bearing off
	if toPoint == 25 {
		// Remove checker from source
		if color == ColorWhite {
			newBoard[fromPoint-1]--
		} else {
			newBoard[fromPoint-1]++
		}
		return &MoveResult{NewBoard: newBoard, HitOpponent: false}, nil
	}

	// Regular move
	// Remove from source
	if color == ColorWhite {
		newBoard[fromPoint-1]--
	} else {
		newBoard[fromPoint-1]++
	}

	// Add to destination (check for hit)
	if color == ColorWhite {
		if newBoard[toPoint-1] == -1 {
			hitOpponent = true
			newBoard[toPoint-1] = 0
		}
		newBoard[toPoint-1]++
	} else {
		if newBoard[toPoint-1] == 1 {
			hitOpponent = true
			newBoard[toPoint-1] = 0
		}
		newBoard[toPoint-1]--
	}

	return &MoveResult{NewBoard: newBoard, HitOpponent: hitOpponent}, nil
}

// ============================================================================
// Legal Moves Generation
// ============================================================================

// GetLegalMoves returns all legal moves for the current position
func GetLegalMoves(board []int, color Color, dice []int, diceUsed []bool, barCount, bornedOff int) []LegalMove {
	legalMoves := []LegalMove{}

	// Get available dice
	availableDice := []int{}
	for i, used := range diceUsed {
		if !used {
			availableDice = append(availableDice, dice[i])
		}
	}

	if len(availableDice) == 0 {
		return legalMoves
	}

	// If on bar, only can enter
	if barCount > 0 {
		for _, die := range availableDice {
			var entryPoint int
			if color == ColorWhite {
				entryPoint = 25 - die
			} else {
				entryPoint = die
			}

			if IsPointOpen(board, entryPoint, color) {
				legalMoves = append(legalMoves, LegalMove{
					FromPoint: 0,
					ToPoint:   entryPoint,
					DieUsed:   die,
				})
			}
		}
		return legalMoves
	}

	// Check if can bear off
	canBear := CanBearOff(board, color, barCount)

	// Try all possible moves
	for point := 1; point <= 24; point++ {
		if CountCheckersOnPoint(board, point, color) == 0 {
			continue
		}

		for _, die := range availableDice {
			toPoint := CalculateToPoint(point, die, color)

			// Try bearing off
			if canBear {
				if toPoint <= 0 || toPoint >= 25 {
					// Bearing off
					err := ValidateMove(board, point, 25, die, color, barCount)
					if err == nil {
						legalMoves = append(legalMoves, LegalMove{
							FromPoint: point,
							ToPoint:   25,
							DieUsed:   die,
						})
					}
				}
			}

			// Regular move
			if toPoint >= 1 && toPoint <= 24 {
				err := ValidateMove(board, point, toPoint, die, color, barCount)
				if err == nil {
					legalMoves = append(legalMoves, LegalMove{
						FromPoint: point,
						ToPoint:   toPoint,
						DieUsed:   die,
					})
				}
			}
		}
	}

	return legalMoves
}

// HasLegalMoves checks if there are any legal moves available
func HasLegalMoves(board []int, color Color, dice []int, diceUsed []bool, barCount int) bool {
	moves := GetLegalMoves(board, color, dice, diceUsed, barCount, 0)
	return len(moves) > 0
}

// ============================================================================
// Dice Management
// ============================================================================

// AllDiceUsed checks if all dice have been used
func AllDiceUsed(diceUsed []bool) bool {
	for _, used := range diceUsed {
		if !used {
			return false
		}
	}
	return true
}

// ============================================================================
// Win Condition
// ============================================================================

// CheckWinCondition checks if a player has won (all 15 checkers borne off)
func CheckWinCondition(bornedOff int) bool {
	return bornedOff >= 15
}
