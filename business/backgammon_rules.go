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
	FromPoint      int   // 0=bar, 1-24=board points, 25=bear off
	ToPoint        int
	DieUsed        int   // Sum of dice values used (e.g., 4 for a 1+3 combined move)
	DiceIndices    []int // Indices of dice being used (for combined moves)
	IsCombinedMove bool  // True if this move uses multiple dice
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
		if expectedTo == 0 || expectedTo == 25 {
			// Exact bear off (white: expectedTo=0, black: expectedTo=25)
			return nil
		}
		if expectedTo < 0 || expectedTo > 25 {
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
		// White moves 24→1, so higher points are those with larger numbers (point+1 to 6)
		for i := point + 1; i <= 6; i++ {
			if board[i-1] > 0 {
				return false
			}
		}
		return true
	} else {
		// Black moves 1→24, so higher points (furthest from start) are those with smaller numbers in home board (19 to point-1)
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

	// Get available dice with their indices
	availableDice := []indexedDie{}
	for i, used := range diceUsed {
		if !used {
			availableDice = append(availableDice, indexedDie{value: dice[i], index: i})
		}
	}

	if len(availableDice) == 0 {
		return legalMoves
	}

	// If on bar, only can enter (no combined moves from bar)
	if barCount > 0 {
		for _, die := range availableDice {
			var entryPoint int
			if color == ColorWhite {
				entryPoint = 25 - die.value
			} else {
				entryPoint = die.value
			}

			if IsPointOpen(board, entryPoint, color) {
				legalMoves = append(legalMoves, LegalMove{
					FromPoint:      0,
					ToPoint:        entryPoint,
					DieUsed:        die.value,
					DiceIndices:    []int{die.index},
					IsCombinedMove: false,
				})
			}
		}
		return legalMoves
	}

	// Check if can bear off
	canBear := CanBearOff(board, color, barCount)

	// Try moves for each point with checkers
	for point := 1; point <= 24; point++ {
		if CountCheckersOnPoint(board, point, color) == 0 {
			continue
		}

		// Try all possible combinations of available dice (1 die, 2 dice, 3 dice, 4 dice)
		for numDice := 1; numDice <= len(availableDice); numDice++ {
			// Generate all combinations of numDice from availableDice
			combinations := generateCombinations(availableDice, numDice)

			for _, combo := range combinations {
				// Calculate total value and indices
				totalValue := 0
				indices := []int{}
				for _, die := range combo {
					totalValue += die.value
					indices = append(indices, die.index)
				}

				// For single die moves, use existing logic
				if numDice == 1 {
					toPoint := CalculateToPoint(point, totalValue, color)

					// Try bearing off
					if canBear && (toPoint <= 0 || toPoint >= 25) {
						err := ValidateMove(board, point, 25, totalValue, color, barCount)
						if err == nil {
							legalMoves = append(legalMoves, LegalMove{
								FromPoint:      point,
								ToPoint:        25,
								DieUsed:        totalValue,
								DiceIndices:    indices,
								IsCombinedMove: false,
							})
						}
					}

					// Regular move
					if toPoint >= 1 && toPoint <= 24 {
						err := ValidateMove(board, point, toPoint, totalValue, color, barCount)
						if err == nil {
							legalMoves = append(legalMoves, LegalMove{
								FromPoint:      point,
								ToPoint:        toPoint,
								DieUsed:        totalValue,
								DiceIndices:    indices,
								IsCombinedMove: false,
							})
						}
					}
				} else {
					// Combined move: validate sequence of moves
					if trySequentialMove(board, point, combo, color, barCount, canBear) {
						finalPoint := CalculateToPoint(point, totalValue, color)

						// Determine final destination
						if canBear && (finalPoint <= 0 || finalPoint >= 25) {
							legalMoves = append(legalMoves, LegalMove{
								FromPoint:      point,
								ToPoint:        25,
								DieUsed:        totalValue,
								DiceIndices:    indices,
								IsCombinedMove: true,
							})
						} else if finalPoint >= 1 && finalPoint <= 24 {
							legalMoves = append(legalMoves, LegalMove{
								FromPoint:      point,
								ToPoint:        finalPoint,
								DieUsed:        totalValue,
								DiceIndices:    indices,
								IsCombinedMove: true,
							})
						}
					}
				}
			}
		}
	}

	return legalMoves
}

// generateCombinations generates all combinations of n dice from the available dice
func generateCombinations(dice []indexedDie, n int) [][]indexedDie {
	result := [][]indexedDie{}
	if n == 0 {
		return result
	}
	if n > len(dice) {
		return result
	}

	var generate func(start int, current []indexedDie)
	generate = func(start int, current []indexedDie) {
		if len(current) == n {
			combo := make([]indexedDie, n)
			copy(combo, current)
			result = append(result, combo)
			return
		}

		for i := start; i < len(dice); i++ {
			generate(i+1, append(current, dice[i]))
		}
	}

	generate(0, []indexedDie{})
	return result
}

// trySequentialMove validates a sequence of moves using multiple dice
func trySequentialMove(board []int, fromPoint int, dice []indexedDie, color Color, barCount int, canBear bool) bool {
	currentBoard := make([]int, len(board))
	copy(currentBoard, board)
	currentPoint := fromPoint

	// Try each die in sequence
	for i, die := range dice {
		toPoint := CalculateToPoint(currentPoint, die.value, color)

		// Last die can bear off
		if i == len(dice)-1 && canBear && (toPoint <= 0 || toPoint >= 25) {
			err := ValidateMove(currentBoard, currentPoint, 25, die.value, color, barCount)
			if err != nil {
				return false
			}
			return true
		}

		// Regular move must land on valid point
		if toPoint < 1 || toPoint > 24 {
			return false
		}

		err := ValidateMove(currentBoard, currentPoint, toPoint, die.value, color, barCount)
		if err != nil {
			return false
		}

		// Execute the move to update board state for next iteration
		result, err := ExecuteMove(currentBoard, currentPoint, toPoint, color)
		if err != nil {
			return false
		}

		currentBoard = result.NewBoard
		currentPoint = toPoint
	}

	return true
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
