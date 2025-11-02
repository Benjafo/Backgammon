import {
    acceptInvitation,
    cancelInvitation,
    declineInvitation,
    getInvitations,
    getLobbyUsers,
    joinLobby,
    leaveLobby,
    sendHeartbeat,
    sendInvitation,
} from "@/api/lobby";
import { getActiveGames } from "@/api/game";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/contexts/AuthContext";
import { type Invitation, type LobbyUser } from "@/types/lobby";
import type { GameData } from "@/types/game";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

export default function LobbyPage() {
    const { user, logout } = useAuth();
    const navigate = useNavigate();

    const [onlineUsers, setOnlineUsers] = useState<LobbyUser[]>([]);
    const [sentInvitations, setSentInvitations] = useState<Invitation[]>([]);
    const [receivedInvitations, setReceivedInvitations] = useState<Invitation[]>([]);
    const [activeGames, setActiveGames] = useState<GameData[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [actionLoading, setActionLoading] = useState<number | null>(null);

    // Join lobby on mount, leave on unmount
    useEffect(() => {
        let mounted = true;

        const initialize = async () => {
            try {
                await joinLobby();
                if (mounted) {
                    setLoading(false);
                    // Initial data fetch
                    await fetchLobbyData();
                }
            } catch (err) {
                console.error("Failed to join lobby:", err);
                if (mounted) {
                    setError(err instanceof Error ? err.message : "Failed to join lobby");
                    setLoading(false);
                }
            }
        };

        initialize();

        return () => {
            mounted = false;
            leaveLobby().catch((err) => console.error("Failed to leave lobby:", err));
        };
    }, []);

    // Fetch lobby data (users, invitations, and active games)
    const fetchLobbyData = async () => {
        try {
            const [users, invitations, games] = await Promise.all([
                getLobbyUsers(),
                getInvitations(),
                getActiveGames(),
            ]);
            setOnlineUsers(users);
            setSentInvitations(invitations.sent);
            setReceivedInvitations(invitations.received);
            setActiveGames(games);
            setError(null);
        } catch (err) {
            console.error("Failed to fetch lobby data:", err);
            setError(err instanceof Error ? err.message : "Failed to fetch lobby data");
        }
    };

    // Send heartbeat every 30 seconds
    useEffect(() => {
        const heartbeatInterval = setInterval(async () => {
            try {
                await sendHeartbeat();
            } catch (err) {
                console.error("Failed to send heartbeat:", err);
            }
        }, 30000);

        return () => clearInterval(heartbeatInterval);
    }, []);

    // Poll for updates every 5 seconds
    useEffect(() => {
        const pollInterval = setInterval(async () => {
            await fetchLobbyData();
        }, 5000);

        return () => clearInterval(pollInterval);
    }, []);

    // Handle challenge user
    const handleChallenge = async (userId: number) => {
        setActionLoading(userId);
        try {
            await sendInvitation(userId);
            await fetchLobbyData();
        } catch (err) {
            console.error("Failed to send invitation:", err);
            alert(err instanceof Error ? err.message : "Failed to send invitation");
        } finally {
            setActionLoading(null);
        }
    };

    // Handle accept invitation
    const handleAccept = async (invitationId: number) => {
        setActionLoading(invitationId);
        try {
            const response = await acceptInvitation(invitationId);
            // Navigate to game page
            navigate(`/game/${response.gameId}`);
        } catch (err) {
            console.error("Failed to accept invitation:", err);
            alert(err instanceof Error ? err.message : "Failed to accept invitation");
            setActionLoading(null);
        }
    };

    // Handle decline invitation
    const handleDecline = async (invitationId: number) => {
        setActionLoading(invitationId);
        try {
            await declineInvitation(invitationId);
            await fetchLobbyData();
        } catch (err) {
            console.error("Failed to decline invitation:", err);
            alert(err instanceof Error ? err.message : "Failed to decline invitation");
        } finally {
            setActionLoading(null);
        }
    };

    // Handle cancel invitation
    const handleCancel = async (invitationId: number) => {
        setActionLoading(invitationId);
        try {
            await cancelInvitation(invitationId);
            await fetchLobbyData();
        } catch (err) {
            console.error("Failed to cancel invitation:", err);
            alert(err instanceof Error ? err.message : "Failed to cancel invitation");
        } finally {
            setActionLoading(null);
        }
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-background flex items-center justify-center">
                <div className="text-center">
                    <h2 className="text-2xl font-bold mb-2">Joining Lobby...</h2>
                    <p className="text-muted-foreground">Please wait</p>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-background p-8">
            <div className="max-w-7xl mx-auto">
                <div className="flex justify-between items-center mb-8">
                    <div>
                        <h1 className="text-3xl font-bold">Backgammon Lobby</h1>
                        <p className="text-muted-foreground">Welcome, {user?.username}!</p>
                    </div>
                    <Button onClick={logout} variant="outline">
                        Logout
                    </Button>
                </div>

                {error && (
                    <Card className="mb-6 border-destructive">
                        <CardContent className="pt-6">
                            <p className="text-destructive">{error}</p>
                        </CardContent>
                    </Card>
                )}

                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                    {/* Active Games */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Active Games</CardTitle>
                            <CardDescription>
                                {activeGames.length} active game{activeGames.length !== 1 ? "s" : ""}
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {activeGames.length === 0 ? (
                                <p className="text-sm text-muted-foreground">
                                    No active games
                                </p>
                            ) : (
                                <div className="space-y-3">
                                    {activeGames.map((game) => {
                                        const isPlayer1 = game.player1.userId === user?.id;
                                        const opponent = isPlayer1 ? game.player2 : game.player1;
                                        const myColor = isPlayer1 ? game.player1.color : game.player2.color;

                                        return (
                                            <div
                                                key={game.gameId}
                                                className="p-3 rounded-md border space-y-2"
                                            >
                                                <p className="font-medium">
                                                    vs {opponent.username}
                                                </p>
                                                <p className="text-xs text-muted-foreground">
                                                    Playing as {myColor} • {game.gameStatus === "pending" ? "Waiting to start" : "In progress"}
                                                </p>
                                                <Button
                                                    size="sm"
                                                    onClick={() => navigate(`/game/${game.gameId}`)}
                                                    className="w-full"
                                                >
                                                    Join Game
                                                </Button>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    {/* Online Users */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Online Players</CardTitle>
                            <CardDescription>
                                {onlineUsers.length} player{onlineUsers.length !== 1 ? "s" : ""}{" "}
                                online
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {onlineUsers.length === 0 ? (
                                <p className="text-sm text-muted-foreground">
                                    No other players online
                                </p>
                            ) : (
                                <div className="space-y-3">
                                    {onlineUsers.map((player) => {
                                        const hasPendingInvitation = sentInvitations.some(
                                            (inv) => inv.challenged.userId === player.userId
                                        );
                                        return (
                                            <div
                                                key={player.userId}
                                                className="flex items-center justify-between p-2 rounded-md border"
                                            >
                                                <div>
                                                    <p className="font-medium">{player.username}</p>
                                                </div>
                                                <Button
                                                    size="sm"
                                                    onClick={() => handleChallenge(player.userId)}
                                                    disabled={
                                                        actionLoading === player.userId ||
                                                        hasPendingInvitation
                                                    }
                                                >
                                                    {hasPendingInvitation ? "Invited" : "Challenge"}
                                                </Button>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    {/* Received Invitations */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Received Invitations</CardTitle>
                            <CardDescription>
                                {receivedInvitations.length} pending invitation
                                {receivedInvitations.length !== 1 ? "s" : ""}
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {receivedInvitations.length === 0 ? (
                                <p className="text-sm text-muted-foreground">
                                    No pending invitations
                                </p>
                            ) : (
                                <div className="space-y-3">
                                    {receivedInvitations.map((invitation) => (
                                        <div
                                            key={invitation.invitationId}
                                            className="p-3 rounded-md border space-y-2"
                                        >
                                            <p className="font-medium">
                                                Challenge from {invitation.challenger.username}
                                            </p>
                                            <div className="flex gap-2">
                                                <Button
                                                    size="sm"
                                                    onClick={() =>
                                                        handleAccept(invitation.invitationId)
                                                    }
                                                    disabled={
                                                        actionLoading === invitation.invitationId
                                                    }
                                                    className="flex-1"
                                                >
                                                    Accept
                                                </Button>
                                                <Button
                                                    size="sm"
                                                    variant="outline"
                                                    onClick={() =>
                                                        handleDecline(invitation.invitationId)
                                                    }
                                                    disabled={
                                                        actionLoading === invitation.invitationId
                                                    }
                                                    className="flex-1"
                                                >
                                                    Decline
                                                </Button>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    {/* Sent Invitations */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Sent Invitations</CardTitle>
                            <CardDescription>
                                {sentInvitations.length} invitation
                                {sentInvitations.length !== 1 ? "s" : ""}
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {sentInvitations.length === 0 ? (
                                <p className="text-sm text-muted-foreground">No sent invitations</p>
                            ) : (
                                <div className="space-y-3">
                                    {sentInvitations.map((invitation) => (
                                        <div
                                            key={invitation.invitationId}
                                            className="p-3 rounded-md border space-y-2"
                                        >
                                            {invitation.status === "pending" ? (
                                                <>
                                                    <p className="font-medium">
                                                        Waiting for {invitation.challenged.username}
                                                    </p>
                                                    <p className="text-xs text-muted-foreground">
                                                        Waiting for response...
                                                    </p>
                                                    <Button
                                                        size="sm"
                                                        variant="outline"
                                                        onClick={() =>
                                                            handleCancel(invitation.invitationId)
                                                        }
                                                        disabled={
                                                            actionLoading === invitation.invitationId
                                                        }
                                                        className="w-full"
                                                    >
                                                        Cancel
                                                    </Button>
                                                </>
                                            ) : (
                                                <>
                                                    <p className="font-medium text-primary">
                                                        {invitation.challenged.username} accepted your
                                                        challenge!
                                                    </p>
                                                    <p className="text-xs text-muted-foreground">
                                                        Game #{invitation.gameId} is ready
                                                    </p>
                                                    <Button
                                                        size="sm"
                                                        onClick={() =>
                                                            navigate(`/game/${invitation.gameId}`)
                                                        }
                                                        className="w-full"
                                                    >
                                                        Join Game
                                                    </Button>
                                                </>
                                            )}
                                        </div>
                                    ))}
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
