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
