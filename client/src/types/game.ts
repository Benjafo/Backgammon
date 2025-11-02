export interface Player {
    userId: number;
    username: string;
    color: string;
}

export interface GameData {
    gameId: number;
    player1: Player;
    player2: Player;
    currentTurn: number;
    gameStatus: "pending" | "in_progress" | "completed" | "abandoned";
    winnerId: number | null;
    createdAt: string;
    startedAt: string | null;
    endedAt: string | null;
}

export interface ActiveGamesResponse {
    games: GameData[];
}

export interface GameState {
    stateId: number;
    gameId: number;
    board: number[]; // 24 integers: positive=white, negative=black, 0=empty
    barWhite: number;
    barBlack: number;
    bornedOffWhite: number;
    bornedOffBlack: number;
    diceRoll: number[] | null; // 2 dice for regular rolls, 4 dice for doubles
    diceUsed: boolean[] | null;
    lastUpdated: string;
}

export interface LegalMove {
    fromPoint: number; // 0=bar, 1-24=board points, 25=bear off
    toPoint: number;
    dieUsed: number; // Sum of dice values (e.g., 4 for a 1+3 combined move)
    diceIndices?: number[]; // Indices of dice being used (for combined moves)
    isCombinedMove?: boolean; // True if using multiple dice
}

export interface MoveRequest {
    fromPoint: number;
    toPoint: number;
    dieUsed: number;
    diceIndices?: number[];
    isCombinedMove?: boolean;
}

export interface LegalMovesResponse {
    moves: LegalMove[];
}
