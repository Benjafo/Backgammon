import {
    forfeitGame,
    getGame,
    getGameState,
    getLegalMoves,
    makeMove,
    rollDice,
    startGame,
} from "@/api/game";
import BackgammonBoard from "@/components/game/BackgammonBoard";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/contexts/AuthContext";
import type { GameData, GameState, LegalMove } from "@/types/game";
import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

export default function GamePage() {
    const { gameId } = useParams<{ gameId: string }>();
    const { user } = useAuth();
    const navigate = useNavigate();
    const [gameData, setGameData] = useState<GameData | null>(null);
    const [gameState, setGameState] = useState<GameState | null>(null);
    const [legalMoves, setLegalMoves] = useState<LegalMove[]>([]);
    const [draggedPoint, setDraggedPoint] = useState<number | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [actionLoading, setActionLoading] = useState(false);
    const hasSeenActiveGame = useRef(false);

    // Fetch game data
    const fetchGameData = async () => {
        if (!gameId) return;

        try {
            const data = await getGame(parseInt(gameId));
            setGameData(data);
            setError(null);

            // Track if we've seen the game in an active state
            if (data.gameStatus === "pending" || data.gameStatus === "in_progress") {
                hasSeenActiveGame.current = true;
            }
        } catch (err) {
            console.error("Failed to fetch game data:", err);
            setError(err instanceof Error ? err.message : "Failed to fetch game data");
        }
    };

    // Fetch game state
    const fetchGameState = async () => {
        if (!gameId) return;
        if (gameData?.gameStatus !== "in_progress") return;

        try {
            const state = await getGameState(parseInt(gameId));
            setGameState(state);
        } catch (err) {
            console.error("Failed to fetch game state:", err);
        }
    };

    // Fetch legal moves
    const fetchLegalMoves = async () => {
        if (!gameId || !gameData || !gameState) return;
        if (gameData.currentTurn !== user?.id) {
            setLegalMoves([]);
            return;
        }
        if (!gameState.diceRoll) {
            setLegalMoves([]);
            return;
        }

        try {
            const moves = await getLegalMoves(parseInt(gameId));
            setLegalMoves(moves);
        } catch (err) {
            console.error("Failed to fetch legal moves:", err);
            setLegalMoves([]);
        }
    };

    // Initial load
    useEffect(() => {
        const initialize = async () => {
            await fetchGameData();
            setLoading(false);
        };

        initialize();
    }, [gameId]);

    // Fetch game state when game data changes
    useEffect(() => {
        if (gameData?.gameStatus === "in_progress") {
            fetchGameState();
        }
    }, [gameData, gameId]);

    // Fetch legal moves when state or turn changes
    useEffect(() => {
        if (gameState && gameData) {
            fetchLegalMoves();
        }
    }, [gameState, gameData, gameId, user?.id]);

    // Redirect to lobby if navigating to an already-ended game
    useEffect(() => {
        if (
            !loading &&
            gameData &&
            (gameData.gameStatus === "completed" || gameData.gameStatus === "abandoned") &&
            !hasSeenActiveGame.current
        ) {
            navigate("/lobby");
        }
    }, [loading, gameData?.gameStatus, navigate]);

    // Poll for updates every 2 seconds
    useEffect(() => {
        const pollInterval = setInterval(async () => {
            await fetchGameData();
            if (gameData?.gameStatus === "in_progress") {
                await fetchGameState();
            }
        }, 2000);

        return () => clearInterval(pollInterval);
    }, [gameId, gameData?.gameStatus]);

    const handleStartGame = async () => {
        if (!gameId) return;
        setActionLoading(true);
        try {
            await startGame(parseInt(gameId));
            await fetchGameData();
            await fetchGameState();
        } catch (err) {
            console.error("Failed to start game:", err);
            alert(err instanceof Error ? err.message : "Failed to start game");
        } finally {
            setActionLoading(false);
        }
    };

    const handleRollDice = async () => {
        if (!gameId) return;
        setActionLoading(true);
        try {
            const newState = await rollDice(parseInt(gameId));
            setGameState(newState);
            await fetchGameData();
        } catch (err) {
            console.error("Failed to roll dice:", err);
            alert(err instanceof Error ? err.message : "Failed to roll dice");
        } finally {
            setActionLoading(false);
        }
    };

    const handleDragStart = (point: number) => {
        if (!gameId || !gameData || !gameState) return;
        if (gameData.currentTurn !== user?.id) return;
        setDraggedPoint(point);
    };

    const handleDrop = async (toPoint: number) => {
        if (!gameId || !gameData || !gameState || draggedPoint === null) return;
        if (gameData.currentTurn !== user?.id) return;

        // Find the legal move
        const move = legalMoves.find((m) => m.fromPoint === draggedPoint && m.toPoint === toPoint);
        if (!move) {
            setDraggedPoint(null);
            return;
        }

        // Execute the move
        setActionLoading(true);
        try {
            const newState = await makeMove(parseInt(gameId), {
                fromPoint: move.fromPoint,
                toPoint: move.toPoint,
                dieUsed: move.dieUsed,
                diceIndices: move.diceIndices,
                isCombinedMove: move.isCombinedMove,
            });
            setGameState(newState);
            setDraggedPoint(null);
            await fetchGameData();
        } catch (err) {
            console.error("Failed to make move:", err);
            alert(err instanceof Error ? err.message : "Failed to make move");
            setDraggedPoint(null);
        } finally {
            setActionLoading(false);
        }
    };

    const handleForfeit = async () => {
        if (!gameId) return;
        if (!confirm("Are you sure you want to forfeit this game?")) return;

        setActionLoading(true);
        try {
            await forfeitGame(parseInt(gameId));
            await fetchGameData();
        } catch (err) {
            console.error("Failed to forfeit game:", err);
            alert(err instanceof Error ? err.message : "Failed to forfeit game");
        } finally {
            setActionLoading(false);
        }
    };

    const handleBackToLobby = () => {
        navigate("/lobby");
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-background flex items-center justify-center">
                <div className="text-center">
                    <h2 className="text-2xl font-bold mb-2">Loading Game...</h2>
                    <p className="text-muted-foreground">Game ID: {gameId}</p>
                </div>
            </div>
        );
    }

    if (!gameData) {
        return (
            <div className="min-h-screen bg-background flex items-center justify-center">
                <Card className="border-destructive">
                    <CardContent className="pt-6">
                        <p className="text-destructive">Failed to load game data</p>
                        <Button onClick={handleBackToLobby} className="mt-4">
                            Back to Lobby
                        </Button>
                    </CardContent>
                </Card>
            </div>
        );
    }

    const isPlayer1 = gameData.player1.userId === user?.id;
    const myPlayer = isPlayer1 ? gameData.player1 : gameData.player2;
    const opponentPlayer = isPlayer1 ? gameData.player2 : gameData.player1;
    const isGameActive = gameData.gameStatus === "in_progress";
    const isMyTurn = gameData.currentTurn === user?.id;
    const myColor = myPlayer.color as "white" | "black";

    return (
        <div className="min-h-screen bg-background p-4">
            <div className="max-w-7xl mx-auto">
                <div className="flex justify-between items-center mb-4">
                    <div>
                        <h1 className="text-2xl font-bold">Backgammon Game</h1>
                        <p className="text-sm text-muted-foreground">Game #{gameId}</p>
                    </div>
                    <Button onClick={handleBackToLobby} variant="outline" size="sm">
                        Back to Lobby
                    </Button>
                </div>

                {error && (
                    <Card className="mb-4 border-destructive">
                        <CardContent className="pt-4">
                            <p className="text-destructive text-sm">{error}</p>
                        </CardContent>
                    </Card>
                )}

                <div className="grid gap-4 lg:grid-cols-[1fr_300px]">
                    {/* Board */}
                    <div>
                        {gameState && gameData.gameStatus !== "pending" ? (
                            <BackgammonBoard
                                gameState={gameState}
                                myColor={myColor}
                                isMyTurn={isMyTurn && isGameActive}
                                legalMoves={isGameActive ? legalMoves : []}
                                draggedPoint={isGameActive ? draggedPoint : null}
                                onDragStart={handleDragStart}
                                onDrop={handleDrop}
                            />
                        ) : (
                            <div className="border-2 border-dashed rounded-lg p-12 text-center">
                                <p className="text-muted-foreground">
                                    Game not started yet
                                </p>
                            </div>
                        )}
                    </div>

                    {/* Sidebar */}
                    <div className="space-y-4">
                        {/* Game Status */}
                        <Card>
                            <CardHeader className="pb-3">
                                <CardTitle className="text-lg">Status</CardTitle>
                                <CardDescription className="capitalize">
                                    {gameData.gameStatus.replace("_", " ")}
                                </CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-3">
                                <div
                                    className={`p-2 border rounded text-sm ${
                                        isMyTurn ? "bg-primary/10 border-primary" : ""
                                    }`}
                                >
                                    <p className="font-medium">{myPlayer.username} (You)</p>
                                    <p className="text-xs text-muted-foreground">
                                        {myPlayer.color}
                                        {isMyTurn && " - Your turn!"}
                                    </p>
                                </div>
                                <div
                                    className={`p-2 border rounded text-sm ${
                                        !isMyTurn && isGameActive
                                            ? "bg-primary/10 border-primary"
                                            : ""
                                    }`}
                                >
                                    <p className="font-medium">{opponentPlayer.username}</p>
                                    <p className="text-xs text-muted-foreground">
                                        {opponentPlayer.color}
                                        {!isMyTurn && isGameActive && " - Their turn"}
                                    </p>
                                </div>

                                {gameData.winnerId && (
                                    <div className="border-t pt-3">
                                        <p className="text-sm font-medium">
                                            {gameData.winnerId === myPlayer.userId
                                                ? "You won!"
                                                : "Better luck next time!"}
                                        </p>
                                    </div>
                                )}
                            </CardContent>
                        </Card>

                        {/* Actions */}
                        <Card>
                            <CardHeader className="pb-3">
                                <CardTitle className="text-lg">Actions</CardTitle>
                            </CardHeader>
                            <CardContent className="space-y-2">
                                {gameData.gameStatus === "pending" && (
                                    <Button
                                        onClick={handleStartGame}
                                        disabled={actionLoading}
                                        className="w-full"
                                        size="sm"
                                    >
                                        Start Game
                                    </Button>
                                )}

                                {isGameActive && isMyTurn && !gameState?.diceRoll && (
                                    <Button
                                        onClick={handleRollDice}
                                        disabled={actionLoading}
                                        className="w-full"
                                    >
                                        Roll Dice
                                    </Button>
                                )}

                                {isGameActive && isMyTurn && gameState?.diceRoll && (
                                    <div className="text-sm">
                                        <p className="font-medium mb-1">Dice:</p>
                                        <div className="flex gap-2 flex-wrap">
                                            {gameState.diceRoll.map((die, index) => (
                                                <div
                                                    key={index}
                                                    className={`border-2 rounded p-2 flex items-center justify-center w-12 h-12 text-lg font-bold ${
                                                        gameState.diceUsed?.[index]
                                                            ? "bg-gray-200 text-gray-400"
                                                            : "bg-white"
                                                    }`}
                                                >
                                                    {die}
                                                </div>
                                            ))}
                                        </div>
                                        {draggedPoint !== null && (
                                            <p className="text-xs text-muted-foreground mt-2">
                                                Dragging from{" "}
                                                {draggedPoint === 0
                                                    ? "Bar"
                                                    : `Point ${draggedPoint}`}
                                            </p>
                                        )}
                                        {legalMoves.length === 0 && (
                                            <p className="text-xs text-destructive mt-2">
                                                No legal moves available
                                            </p>
                                        )}
                                    </div>
                                )}

                                {isGameActive && (
                                    <Button
                                        onClick={handleForfeit}
                                        disabled={actionLoading}
                                        variant="destructive"
                                        className="w-full"
                                        size="sm"
                                    >
                                        Forfeit Game
                                    </Button>
                                )}

                                {!isGameActive && gameData.gameStatus !== "pending" && (
                                    <p className="text-xs text-muted-foreground text-center py-4">
                                        Game has ended. You can view the results but navigating away
                                        will prevent you from returning.
                                    </p>
                                )}
                            </CardContent>
                        </Card>

                        {/* Info */}
                        <Card>
                            <CardHeader className="pb-3">
                                <CardTitle className="text-lg">Info</CardTitle>
                            </CardHeader>
                            <CardContent className="text-xs text-muted-foreground space-y-1">
                                <p>Created: {new Date(gameData.createdAt).toLocaleString()}</p>
                                {gameData.startedAt && (
                                    <p>Started: {new Date(gameData.startedAt).toLocaleString()}</p>
                                )}
                                {gameData.endedAt && (
                                    <p>Ended: {new Date(gameData.endedAt).toLocaleString()}</p>
                                )}
                                <p className="pt-2 border-t">Polling every 2 seconds</p>
                            </CardContent>
                        </Card>
                    </div>
                </div>
            </div>
        </div>
    );
}
