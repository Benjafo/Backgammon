import { forfeitGame, getGame, startGame } from "@/api/game";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/contexts/AuthContext";
import type { GameData } from "@/types/game";
import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

export default function GamePage() {
    const { gameId } = useParams<{ gameId: string }>();
    const { user } = useAuth();
    const navigate = useNavigate();
    const [gameData, setGameData] = useState<GameData | null>(null);
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

    // Initial load
    useEffect(() => {
        const initialize = async () => {
            await fetchGameData();
            setLoading(false);
        };

        initialize();
    }, [gameId]);

    // Redirect to lobby if navigating to an already-ended game (not if it ended during play)
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

    // Poll for updates every 3 seconds
    useEffect(() => {
        const pollInterval = setInterval(async () => {
            await fetchGameData();
        }, 3000);

        return () => clearInterval(pollInterval);
    }, [gameId]);

    const handleStartGame = async () => {
        if (!gameId) return;
        setActionLoading(true);
        try {
            await startGame(parseInt(gameId));
            await fetchGameData();
        } catch (err) {
            console.error("Failed to start game:", err);
            alert(err instanceof Error ? err.message : "Failed to start game");
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
    const isGameActive = gameData.gameStatus === "in_progress" || gameData.gameStatus === "pending";

    return (
        <div className="min-h-screen bg-background p-8">
            <div className="max-w-7xl mx-auto">
                <div className="flex justify-between items-center mb-8">
                    <div>
                        <h1 className="text-3xl font-bold">Backgammon Game</h1>
                        <p className="text-muted-foreground">Game ID: {gameId}</p>
                    </div>
                    <Button onClick={handleBackToLobby} variant="outline">
                        Back to Lobby
                    </Button>
                </div>

                {error && (
                    <Card className="mb-6 border-destructive">
                        <CardContent className="pt-6">
                            <p className="text-destructive">{error}</p>
                        </CardContent>
                    </Card>
                )}

                <div className="grid gap-6 md:grid-cols-2">
                    {/* Game Status */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Game Status</CardTitle>
                            <CardDescription>
                                Status:{" "}
                                <span className="font-semibold capitalize">
                                    {gameData.gameStatus.replace("_", " ")}
                                </span>
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="space-y-4">
                                {/* Players */}
                                <div>
                                    <p className="text-sm font-medium mb-2">Players:</p>
                                    <div className="space-y-2">
                                        <div
                                            className={`p-3 border rounded ${
                                                gameData.currentTurn === myPlayer.userId
                                                    ? "bg-primary/10 border-primary"
                                                    : ""
                                            }`}
                                        >
                                            <p className="font-medium">{myPlayer.username} (You)</p>
                                            <p className="text-sm text-muted-foreground">
                                                Playing as {myPlayer.color}
                                                {gameData.currentTurn === myPlayer.userId && (
                                                    <span className="ml-2 text-primary font-semibold">
                                                        Your turn
                                                    </span>
                                                )}
                                            </p>
                                        </div>
                                        <div
                                            className={`p-3 border rounded ${
                                                gameData.currentTurn === opponentPlayer.userId
                                                    ? "bg-primary/10 border-primary"
                                                    : ""
                                            }`}
                                        >
                                            <p className="font-medium">{opponentPlayer.username}</p>
                                            <p className="text-sm text-muted-foreground">
                                                Playing as {opponentPlayer.color}
                                                {gameData.currentTurn === opponentPlayer.userId && (
                                                    <span className="ml-2 text-primary font-semibold">
                                                        Their turn
                                                    </span>
                                                )}
                                            </p>
                                        </div>
                                    </div>
                                </div>

                                {/* Winner */}
                                {gameData.winnerId && (
                                    <div className="border-t pt-4">
                                        <p className="text-sm font-medium mb-2">Winner:</p>
                                        <p className="text-lg font-bold">
                                            {gameData.winnerId === myPlayer.userId
                                                ? "You won!"
                                                : `${opponentPlayer.username} won!`}
                                        </p>
                                    </div>
                                )}

                                {/* Timestamps */}
                                <div className="border-t pt-4 text-sm space-y-1">
                                    <p>
                                        <span className="text-muted-foreground">Created:</span>{" "}
                                        {new Date(gameData.createdAt).toLocaleString()}
                                    </p>
                                    {gameData.startedAt && (
                                        <p>
                                            <span className="text-muted-foreground">Started:</span>{" "}
                                            {new Date(gameData.startedAt).toLocaleString()}
                                        </p>
                                    )}
                                    {gameData.endedAt && (
                                        <p>
                                            <span className="text-muted-foreground">Ended:</span>{" "}
                                            {new Date(gameData.endedAt).toLocaleString()}
                                        </p>
                                    )}
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    {/* Game Actions */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Game Actions</CardTitle>
                            <CardDescription>Test backend infrastructure</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="space-y-3">
                                {gameData.gameStatus === "pending" && (
                                    <Button
                                        onClick={handleStartGame}
                                        disabled={actionLoading}
                                        className="w-full"
                                    >
                                        Start Game
                                    </Button>
                                )}

                                {isGameActive && (
                                    <Button
                                        onClick={handleForfeit}
                                        disabled={actionLoading}
                                        variant="destructive"
                                        className="w-full"
                                    >
                                        Forfeit Game
                                    </Button>
                                )}

                                {!isGameActive && (
                                    <div className="text-center py-4 space-y-2">
                                        <p className="text-sm font-medium">Game has ended.</p>
                                        <p className="text-xs text-muted-foreground">
                                            You can stay to view the results, but navigating away
                                            will prevent you from returning to this page.
                                        </p>
                                    </div>
                                )}

                                <div className="border-t pt-4 mt-4">
                                    <p className="text-xs text-muted-foreground mb-2">
                                        Polling for updates every 3 seconds
                                    </p>
                                    <p className="text-xs text-muted-foreground">
                                        Game board implementation coming soon!
                                    </p>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
