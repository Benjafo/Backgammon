import {
    forfeitGame,
    getGame,
    getGameState,
    getLegalMoves,
    makeMove,
    rollDice,
} from "@/api/game";
import ChatPanel from "@/components/common/ChatPanel";
import BackgammonBoard from "@/components/game/BackgammonBoard";
import { DiceDisplay } from "@/components/game/Dice";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/contexts/AuthContext";
import { GameChatProvider, useGameChatContext } from "@/contexts/ChatContext";
import type { GameData, GameState, LegalMove } from "@/types/game";
import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

// Separate component to avoid recreation on every GamePage render
function GameChatPanel({ currentUsername }: { currentUsername: string }) {
    const { messages, connectionStatus, error, sendMessage } = useGameChatContext();
    return (
        <ChatPanel
            currentUsername={currentUsername}
            messages={messages}
            connectionStatus={connectionStatus}
            error={error}
            sendMessage={sendMessage}
            title="Game Chat"
            subtitle="Chat with your opponent"
        />
    );
}

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

    // Poll for updates every second
    useEffect(() => {
        const pollInterval = setInterval(async () => {
            await fetchGameData();
            if (gameData?.gameStatus === "in_progress") {
                await fetchGameState();
            }
        }, 1000);

        return () => clearInterval(pollInterval);
    }, [gameId, gameData?.gameStatus]);

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

        // Store previous state for rollback
        const previousState = gameState;
        const previousLegalMoves = legalMoves;

        // Optimistically apply the move to the UI
        const optimisticState = applyMoveOptimistically(gameState, move, gameData);
        setGameState(optimisticState);
        setLegalMoves([]); // Clear legal moves to prevent flash
        setDraggedPoint(null);

        // Execute the move on the server
        setActionLoading(true);
        try {
            const newState = await makeMove(parseInt(gameId), {
                fromPoint: move.fromPoint,
                toPoint: move.toPoint,
                dieUsed: move.dieUsed,
                diceIndices: move.diceIndices,
                isCombinedMove: move.isCombinedMove,
            });
            // Update with server's authoritative state
            setGameState(newState);
            await fetchGameData();
        } catch (err) {
            console.error("Failed to make move:", err);
            // Rollback to previous state on error
            setGameState(previousState);
            setLegalMoves(previousLegalMoves);
            alert(err instanceof Error ? err.message : "Failed to make move");
        } finally {
            setActionLoading(false);
        }
    };

    // Helper function to apply a move optimistically to the game state
    const applyMoveOptimistically = (
        state: GameState,
        move: LegalMove,
        game: GameData
    ): GameState => {
        const newState = { ...state };
        newState.board = [...state.board];
        newState.diceUsed = state.diceUsed ? [...state.diceUsed] : null;

        // Determine player color
        const isWhite =
            game.player1.userId === user?.id
                ? game.player1.color === "white"
                : game.player2.color === "white";
        const checkerValue = isWhite ? 1 : -1;

        // Handle moving FROM a point
        if (move.fromPoint === 0) {
            // Moving from bar
            if (isWhite) {
                newState.barWhite = Math.max(0, state.barWhite - 1);
            } else {
                newState.barBlack = Math.max(0, state.barBlack - 1);
            }
        } else if (move.fromPoint >= 1 && move.fromPoint <= 24) {
            // Moving from board point
            newState.board[move.fromPoint - 1] -= checkerValue;
        }

        // Handle moving TO a point
        if (move.toPoint === 25) {
            // Bearing off
            if (isWhite) {
                newState.bornedOffWhite = state.bornedOffWhite + 1;
            } else {
                newState.bornedOffBlack = state.bornedOffBlack + 1;
            }
        } else if (move.toPoint === 0) {
            // Hitting opponent's checker (sending to bar)
            const targetPoint = newState.board[move.toPoint];
            if (Math.abs(targetPoint) === 1) {
                // Opponent has a blot here, send it to bar
                if (isWhite) {
                    newState.barBlack += 1;
                    newState.board[move.toPoint] = checkerValue;
                } else {
                    newState.barWhite += 1;
                    newState.board[move.toPoint] = checkerValue;
                }
            } else {
                newState.board[move.toPoint] += checkerValue;
            }
        } else if (move.toPoint >= 1 && move.toPoint <= 24) {
            // Moving to board point
            const targetPoint = newState.board[move.toPoint - 1];
            const opponentValue = isWhite ? -1 : 1;

            if (targetPoint === opponentValue) {
                // Hitting opponent's blot
                if (isWhite) {
                    newState.barBlack += 1;
                } else {
                    newState.barWhite += 1;
                }
                newState.board[move.toPoint - 1] = checkerValue;
            } else {
                newState.board[move.toPoint - 1] += checkerValue;
            }
        }

        // Mark dice as used
        if (newState.diceUsed && move.diceIndices) {
            for (const idx of move.diceIndices) {
                if (idx >= 0 && idx < newState.diceUsed.length) {
                    newState.diceUsed[idx] = true;
                }
            }
        }

        return newState;
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
            <div className="min-h-screen bg-felt felt-texture flex items-center justify-center">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-4 border-gold/20 border-t-gold mx-auto mb-4" />
                    <h2 className="text-2xl font-display font-bold mb-2 text-gold-light">
                        Loading Game...
                    </h2>
                    <p className="text-muted-foreground">Game #{gameId}</p>
                </div>
            </div>
        );
    }

    if (!gameData) {
        return (
            <div className="min-h-screen bg-felt felt-texture flex items-center justify-center">
                <Card className="border-destructive shadow-gold">
                    <CardContent className="pt-6">
                        <p className="text-destructive">Failed to load game data</p>
                        <Button onClick={handleBackToLobby} variant="casino" className="mt-4">
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
        <GameChatProvider gameId={gameId ? parseInt(gameId) : null}>
            <div className="min-h-screen bg-felt felt-texture p-4 pr-[336px]">
                <div className="max-w-7xl mx-auto">
                    <div className="flex justify-between items-center mb-4">
                        <div>
                            <h1 className="text-3xl font-display font-bold text-gold-light">
                                Backgammon Game
                            </h1>
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
                        <div className="lg:col-span-1">
                            {gameState ? (
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
                                    <p className="text-muted-foreground">Loading game board...</p>
                                </div>
                            )}
                        </div>

                        {/* Sidebar */}
                        <div className="space-y-4">
                            {/* Game Status */}
                            <Card className="bg-black/60 backdrop-blur-sm border-2 border-gold">
                                <CardHeader className="pb-3">
                                    <CardTitle className="text-lg">Status</CardTitle>
                                    <CardDescription className="capitalize">
                                        {gameData.gameStatus.replace("_", " ")}
                                    </CardDescription>
                                </CardHeader>
                                <CardContent className="space-y-3">
                                    <div
                                        className={`p-3 border-2 rounded-ornate text-sm ${
                                            isMyTurn
                                                ? "bg-gold/10 border-gold shadow-gold"
                                                : "border-gold/30"
                                        }`}
                                    >
                                        <p className="font-medium">{myPlayer.username} (You)</p>
                                        <p className="text-xs text-muted-foreground">
                                            {myPlayer.color}
                                            {isMyTurn && " - Your turn!"}
                                        </p>
                                    </div>
                                    <div
                                        className={`p-3 border-2 rounded-ornate text-sm ${
                                            !isMyTurn && isGameActive
                                                ? "bg-gold/10 border-gold shadow-gold"
                                                : "border-gold/30"
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
                            <Card className="bg-black/60 backdrop-blur-sm border-2 border-gold">
                                <CardHeader className="pb-3">
                                    <CardTitle className="text-lg">Actions</CardTitle>
                                </CardHeader>
                                <CardContent className="space-y-2">
                                    {isGameActive && isMyTurn && !gameState?.diceRoll && (
                                        <Button
                                            onClick={handleRollDice}
                                            disabled={actionLoading}
                                            variant="casino"
                                            className="w-full"
                                        >
                                            Roll Dice
                                        </Button>
                                    )}

                                    {isGameActive && isMyTurn && gameState?.diceRoll && (
                                        <div className="text-sm">
                                            <div className="flex items-center gap-2 mb-1">
                                                <p className="font-medium">Dice:</p>
                                                <DiceDisplay
                                                    dice={gameState.diceRoll}
                                                    used={gameState.diceUsed || []}
                                                    size={45}
                                                />
                                            </div>
                                            {draggedPoint !== null && (
                                                <p className="text-xs text-muted-foreground mt-2">
                                                    Dragging from{" "}
                                                    {draggedPoint === 0
                                                        ? "Bar"
                                                        : `Point ${draggedPoint}`}
                                                </p>
                                            )}
                                            {legalMoves.length === 0 && !actionLoading && (
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
                                            Game has ended. You can view the results but navigating
                                            away will prevent you from returning.
                                        </p>
                                    )}
                                </CardContent>
                            </Card>

                            {/* Info */}
                            <Card className="bg-black/60 backdrop-blur-sm border-2 border-gold">
                                <CardHeader className="pb-3">
                                    <CardTitle className="text-lg">Info</CardTitle>
                                </CardHeader>
                                <CardContent className="text-xs text-muted-foreground space-y-1">
                                    <p>
                                        Game Created:{" "}
                                        {new Date(gameData.createdAt).toLocaleString()}
                                    </p>
                                    {gameData.startedAt && (
                                        <p>
                                            Game Started:{" "}
                                            {new Date(gameData.startedAt).toLocaleString()}
                                        </p>
                                    )}
                                    {gameData.endedAt && (
                                        <p>
                                            Game Ended:{" "}
                                            {new Date(gameData.endedAt).toLocaleString()}
                                        </p>
                                    )}
                                </CardContent>
                            </Card>
                        </div>
                    </div>
                </div>

                {/* Fixed Chat Panel on Right */}
                <div className="fixed top-0 right-0 bottom-0 w-80 border-l border-gold/20 bg-card/30 backdrop-blur-sm z-40">
                    <GameChatPanel currentUsername={user?.username || ""} />
                </div>
            </div>
        </GameChatProvider>
    );
}
