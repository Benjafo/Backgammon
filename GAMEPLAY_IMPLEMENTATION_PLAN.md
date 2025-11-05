# Backgammon Gameplay Implementation Plan

## Overview
This document outlines the complete implementation plan for the backgammon gameplay system, adhering to the specifications in instructions.txt and schema.sql.

## Key Requirements from instructions.txt
- ✅ **Turn-based**: Only current player can make moves
- ✅ **Database state**: All game state stored in database
- ✅ **SVG board**: Game board drawn using SVG
- ✅ **AJAX polling**: Boards update automatically without refresh
- ✅ **Illegal move prevention**: Validate moves and provide feedback
- ✅ **Page reload resilience**: Game resumes where it left off
- ✅ **Win detection**: Both players notified when someone wins
- ✅ **SVG animation**: Animate board updates (optional enhancement)

---

## Phase 1: Backend - Game State Management

### 1.1 Repository Layer (`repository/game.go`)

**Add GAME_STATE Management to existing game.go:**
```go
type GameState struct {
    StateID         int
    GameID          int
    BoardState      string // JSONB as string
    BarWhite        int
    BarBlack        int
    BornedOffWhite  int
    BornedOffBlack  int
    DiceRoll        string // JSONB [die1, die2] or null
    DiceUsed        string // JSONB [bool, bool] or null
    LastUpdated     time.Time
}

// Methods to implement:
- InitializeGameState(gameID) - Create initial board position
- GetGameState(gameID) - Retrieve current state
- UpdateGameState(gameID, state) - Update entire state
- RollDice(gameID) - Generate new dice roll
- MarkDiceUsed(gameID, diceIndex) - Mark die as used
- ClearDice(gameID) - Clear dice at end of turn
```

**Board State Representation:**
- JSONB structure: `{"points": [count, count, ...], "colors": ["white"/"black"/null, ...]}`
- 24 points indexed 0-23 (point 1 = index 0, point 24 = index 23)
- Positive count = white checkers, negative = black checkers, 0 = empty
- Alternative: Array of 24 objects: `[{color: "white", count: 2}, ...]`

**Initial Board Setup:**
```
Point 1:  2 white    Point 13: 5 black
Point 6:  5 black    Point 17: 3 white
Point 8:  3 black    Point 19: 5 white
Point 12: 5 white    Point 24: 2 black
```

### 1.2 Repository Layer - MOVE History (`repository/game.go`)

**Add MOVE History Management to existing game.go:**
```go
type Move struct {
    MoveID       int
    GameID       int
    PlayerID     int
    MoveNumber   int
    FromPoint    int // 0-25 (0=bar, 1-24=points, 25=borne off)
    ToPoint      int
    DieUsed      int
    HitOpponent  bool
    Timestamp    time.Time
}

// Methods:
- CreateMove(move) - Record a move
- GetMoveHistory(gameID) - Get all moves for a game
- GetLastMoveNumber(gameID) - Get the latest move number
```

### 1.3 Game Rules Engine (`business/backgammon_rules.go`)

**Core Rules Functions:**
```go
// Board analysis
- IsPointOpen(boardState, point, color) bool
- GetLegalMoves(boardState, color, dice, barCount, bornedOff) []Move
- CanBearOff(boardState, color) bool
- IsInHomeBoard(point, color) bool

// Move validation
- ValidateMove(boardState, fromPoint, toPoint, die, color) error
- MustEnterFromBar(barCount) bool
- HasLegalMoves(boardState, color, dice, barCount) bool

// Move execution
- ExecuteMove(boardState, fromPoint, toPoint, color) (newState, hitOpponent)
- CalculateToPoint(fromPoint, dieValue, color) int

// Win detection
- CheckWinCondition(bornedOffCount) bool

// Dice logic
- GetUsableDice(diceRoll, diceUsed) []int
- MustUseHigherDie(boardState, dice, color) bool
- AllDiceUsed(diceUsed) bool
```

### 1.4 Service Layer (`service/game.go`)

**HTTP Handlers:**
```go
// GET /api/v1/games/{id}/state
- GetGameStateHandler - Return current game state

// POST /api/v1/games/{id}/roll
- RollDiceHandler - Roll dice for current turn

// POST /api/v1/games/{id}/move
- MoveHandler - Execute a move
- Validate: Is it player's turn?
- Validate: Is dice rolled?
- Validate: Is move legal?
- Execute move, update state
- Check if turn is complete (all dice used or no legal moves)
- If turn complete, switch to other player and clear dice

// POST /api/v1/games/{id}/end-turn
- EndTurnHandler - Manually end turn (if can't use all dice)
```

---

## Phase 2: Frontend - Game State Types & API

### 2.1 TypeScript Types (`client/src/types/game.ts`)

**Add Game State Types:**
```typescript
export interface Point {
    checkers: number;  // Positive = white, negative = black, 0 = empty
}

export interface GameState {
    stateId: number;
    gameId: number;
    board: Point[];  // 24 points
    barWhite: number;
    barBlack: number;
    bornedOffWhite: number;
    bornedOffBlack: number;
    diceRoll: [number, number] | null;
    diceUsed: [boolean, boolean] | null;
    lastUpdated: string;
}

export interface LegalMove {
    fromPoint: number;  // 0 = bar, 1-24 = board points, 25 = bear off
    toPoint: number;
    dieUsed: number;
}

export interface MoveRequest {
    fromPoint: number;
    toPoint: number;
    dieUsed: number;
}
```

### 2.2 Game API (`client/src/api/game.ts`)

**Add Game State API Functions:**
```typescript
// Get current game state
export async function getGameState(gameId: number): Promise<GameState>

// Roll dice for current turn
export async function rollDice(gameId: number): Promise<GameState>

// Execute a move
export async function makeMove(gameId: number, move: MoveRequest): Promise<GameState>

// Get legal moves for current position
export async function getLegalMoves(gameId: number): Promise<LegalMove[]>

// End turn manually
export async function endTurn(gameId: number): Promise<GameState>
```

---

## Phase 3: Frontend - SVG Board Rendering

### 3.1 Board Component (`client/src/components/game/BackgammonBoard.tsx`)

**SVG Board Structure:**
```
<svg viewBox="0 0 800 600">
  <!-- Outer border -->
  <rect class="board-border" />

  <!-- Two halves (separated by bar) -->
  <g class="board-left">  <!-- Points 1-12 -->
    <!-- 12 triangular points -->
  </g>

  <g class="board-right"> <!-- Points 13-24 -->
    <!-- 12 triangular points -->
  </g>

  <!-- Bar (middle section) -->
  <g class="bar">
    <!-- White pieces on bar -->
    <!-- Black pieces on bar -->
  </g>

  <!-- Borne off areas (outside board) -->
  <g class="borne-off-white">
    <!-- White pieces borne off -->
  </g>
  <g class="borne-off-black">
    <!-- Black pieces borne off -->
  </g>

  <!-- Checkers (circles) -->
  <g class="checkers">
    <!-- Render each checker with position -->
  </g>

  <!-- Dice display -->
  <g class="dice">
    <!-- Show current dice roll -->
  </g>
</svg>
```

**Board Layout:**
- Points arranged in traditional backgammon layout
- Points 1-6 and 19-24 on bottom (white's home board on right)
- Points 7-12 and 13-18 on top (black's home board on right)
- Bar in the middle
- Borne off areas on the sides

### 3.2 Checker Component (`client/src/components/game/Checker.tsx`)

**SVG Checker:**
- Circle element with color fill
- Drop shadow for depth
- Highlight when selected
- Animate position changes using `<animate>` or CSS transitions

### 3.3 Point Component (`client/src/components/game/Point.tsx`)

**SVG Point (Triangle):**
- Polygon element forming triangle
- Alternating colors for visibility
- Click handler to select point
- Highlight when valid move destination

### 3.4 Dice Component (`client/src/components/game/Dice.tsx`)

**SVG Dice Display:**
- Two die faces showing current roll
- Use dots or numbers
- Gray out used dice
- Roll button (only shown when it's player's turn and no dice rolled)

---

## Phase 4: Frontend - Game Interaction Logic

### 4.1 Game Page Enhancement (`client/src/pages/GamePage.tsx`)

**Add Game State Management:**
```typescript
const [gameState, setGameState] = useState<GameState | null>(null);
const [legalMoves, setLegalMoves] = useState<LegalMove[]>([]);
const [selectedPoint, setSelectedPoint] = useState<number | null>(null);
const [myColor, setMyColor] = useState<'white' | 'black'>();

// Determine player's color based on gameData
useEffect(() => {
    if (gameData && user) {
        setMyColor(gameData.player1.userId === user.id
            ? gameData.player1.color
            : gameData.player2.color);
    }
}, [gameData, user]);

// Poll game state every 2 seconds
useEffect(() => {
    const interval = setInterval(async () => {
        const state = await getGameState(gameId);
        setGameState(state);

        // If it's my turn and dice are rolled, get legal moves
        if (isMyTurn && state.diceRoll) {
            const moves = await getLegalMoves(gameId);
            setLegalMoves(moves);
        }
    }, 2000);

    return () => clearInterval(interval);
}, [gameId, isMyTurn]);
```

**Move Interaction Flow:**
1. Player clicks "Roll Dice" button (only if it's their turn and no dice rolled)
2. Server generates random dice, returns updated state
3. Client fetches legal moves
4. Player clicks a point with their checker → selected
5. Client highlights valid destination points
6. Player clicks valid destination → move executed
7. Server validates and executes move
8. State updated, dice marked as used
9. If all dice used or no legal moves remain → turn auto-ends
10. Opponent's view updates via polling

---

## Phase 5: User Testing

### 5.1 Move Validation Tests
- Can't move when not your turn
- Must enter from bar before other moves
- Can't move to blocked points
- Bearing off only when all in home board
- Must use both dice if possible
- Doubles give 4 moves

### 5.2 Win Condition Tests
- Detect when player bears off 15th checker
- Update game status to 'completed'
- Set winner_id
- Notify both players
- Prevent further moves

### 5.3 Page Reload Tests
- Refresh during opponent's turn - should show current state
- Refresh during own turn - should resume with dice/state
- Refresh after game ends - should show final state

### 5.4 Illegal Move Feedback
- Toast notifications for errors
- Highlight invalid selections
- Clear error messages

---

## API Endpoints Summary

### Game State Endpoints
- `GET /api/v1/games/{id}/state` - Get current state
- `POST /api/v1/games/{id}/roll` - Roll dice
- `POST /api/v1/games/{id}/move` - Make a move
- `GET /api/v1/games/{id}/legal-moves` - Get legal moves
- `POST /api/v1/games/{id}/end-turn` - End turn manually

---

## Success Criteria

- ✅ Two players can play a full game of backgammon
- ✅ All backgammon rules are enforced
- ✅ Board is rendered using SVG
- ✅ Game state persists across page reloads
- ✅ Only current player can make moves
- ✅ Illegal moves are prevented with feedback
- ✅ Winner is detected and both players notified
- ✅ Game updates automatically via AJAX polling
- ✅ Works in Firefox, Chrome, Safari, Edge

---

## Notes & Considerations

1. **File Organization**:
   - All GAME, GAME_STATE, and MOVE repository methods in `repository/game.go`
   - All game HTTP handlers in `service/game.go`
   - Business logic in `business/backgammon_rules.go`
2. **Board State Storage**: Using JSONB for flexibility
3. **Move Direction**: White moves 24→1, Black moves 1→24
4. **Dice Logic**: Handle doubles (4 moves), forced moves, higher die priority
5. **Animation**: CSS transitions for smooth checker movement
6. **Undo**: Not required, but could add move history review
7. **WebSockets**: Polling works fine for initial version, can upgrade later
8. **Mobile**: SVG is responsive, ensure touch-friendly hit targets

This plan provides a complete roadmap from database to UI for a fully functional backgammon game!
