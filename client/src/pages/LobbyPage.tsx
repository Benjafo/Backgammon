import { getActiveGames } from "@/api/game";
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
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/contexts/AuthContext";
import type { GameData } from "@/types/game";
import { type Invitation, type LobbyUser } from "@/types/lobby";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import ChatPanel from "@/components/common/ChatPanel";

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
        <div className="min-h-screen bg-background flex flex-col">
            {/* Header */}
            <div className="border-b">
                <div className="max-w-full px-6 py-4 flex justify-between items-center">
                    <div>
                        <h1 className="text-2xl font-bold">Backgammon Lobby</h1>
                        <p className="text-sm text-muted-foreground">Welcome, {user?.username}!</p>
                    </div>
                    <Button onClick={logout} variant="outline" size="sm">
                        Logout
                    </Button>
                </div>
            </div>

            {/* Error Banner */}
            {error && (
                <div className="bg-destructive/10 border-b border-destructive px-6 py-3">
                    <p className="text-sm text-destructive">{error}</p>
                </div>
            )}

            {/* Main Content - Split Panel Layout */}
            <div className="flex-1 flex overflow-hidden">
                {/* Left Panel - Main Lobby Content */}
                <div className="flex-1 overflow-y-auto p-6 space-y-6">
                    {/* Active Games Section */}
                    <div>
                        <div className="flex items-center justify-between mb-3">
                            <div className="flex items-center gap-2">
                                <h2 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                                    My Games
                                </h2>
                                {activeGames.length > 0 && (
                                    <span className="inline-flex items-center justify-center w-5 h-5 text-xs font-medium rounded-full bg-primary text-primary-foreground">
                                        {activeGames.length}
                                    </span>
                                )}
                            </div>
                        </div>
                        <div className="bg-card rounded-lg border">
                            {activeGames.length === 0 ? (
                                <div className="p-4 text-center">
                                    <p className="text-sm text-muted-foreground">
                                        No active games
                                    </p>
                                </div>
                            ) : (
                                <div className="divide-y">
                                    {activeGames.map((game) => {
                                        const isPlayer1 = game.player1.userId === user?.id;
                                        const opponent = isPlayer1 ? game.player2 : game.player1;
                                        const myColor = isPlayer1
                                            ? game.player1.color
                                            : game.player2.color;

                                        return (
                                            <div
                                                key={game.gameId}
                                                className="p-4 flex items-center justify-between hover:bg-accent/50 transition-colors"
                                            >
                                                <div className="flex-1">
                                                    <p className="font-medium text-sm">
                                                        vs {opponent.username}
                                                    </p>
                                                    <div className="flex items-center gap-2 mt-1">
                                                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-secondary text-secondary-foreground">
                                                            {myColor}
                                                        </span>
                                                        <span className="text-xs text-muted-foreground">
                                                            {game.gameStatus === "pending"
                                                                ? "Waiting to start"
                                                                : "In progress"}
                                                        </span>
                                                    </div>
                                                </div>
                                                <Button
                                                    size="sm"
                                                    onClick={() => navigate(`/game/${game.gameId}`)}
                                                >
                                                    Join Game
                                                </Button>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Invitations Section */}
                    <div>
                        <div className="flex items-center justify-between mb-3">
                            <div className="flex items-center gap-2">
                                <h2 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                                    Invitations
                                </h2>
                                {(receivedInvitations.length + sentInvitations.length) > 0 && (
                                    <span className="inline-flex items-center justify-center w-5 h-5 text-xs font-medium rounded-full bg-primary text-primary-foreground">
                                        {receivedInvitations.length + sentInvitations.length}
                                    </span>
                                )}
                            </div>
                        </div>
                        <div className="bg-card rounded-lg border">
                            {receivedInvitations.length === 0 && sentInvitations.length === 0 ? (
                                <div className="p-4 text-center">
                                    <p className="text-sm text-muted-foreground">
                                        No pending invitations
                                    </p>
                                </div>
                            ) : (
                                <div className="divide-y">
                                    {/* Received Invitations */}
                                    {receivedInvitations.map((invitation) => (
                                        <div
                                            key={`received-${invitation.invitationId}`}
                                            className="p-4 flex items-center justify-between hover:bg-accent/50 transition-colors"
                                        >
                                            <div className="flex-1">
                                                <div className="flex items-center gap-2">
                                                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-500/10 text-green-600 dark:text-green-400">
                                                        Received
                                                    </span>
                                                    <p className="font-medium text-sm">
                                                        from {invitation.challenger.username}
                                                    </p>
                                                </div>
                                            </div>
                                            <div className="flex gap-2">
                                                <Button
                                                    size="sm"
                                                    onClick={() =>
                                                        handleAccept(invitation.invitationId)
                                                    }
                                                    disabled={
                                                        actionLoading === invitation.invitationId
                                                    }
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
                                                >
                                                    Decline
                                                </Button>
                                            </div>
                                        </div>
                                    ))}

                                    {/* Sent Invitations */}
                                    {sentInvitations.map((invitation) => (
                                        <div
                                            key={`sent-${invitation.invitationId}`}
                                            className="p-4 flex items-center justify-between hover:bg-accent/50 transition-colors"
                                        >
                                            <div className="flex-1">
                                                {invitation.status === "pending" ? (
                                                    <div className="flex items-center gap-2">
                                                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-500/10 text-yellow-600 dark:text-yellow-400">
                                                            Sent
                                                        </span>
                                                        <p className="font-medium text-sm">
                                                            to {invitation.challenged.username}
                                                        </p>
                                                        <span className="text-xs text-muted-foreground">
                                                            Waiting for response...
                                                        </span>
                                                    </div>
                                                ) : (
                                                    <div>
                                                        <div className="flex items-center gap-2">
                                                            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-500/10 text-green-600 dark:text-green-400">
                                                                Accepted
                                                            </span>
                                                            <p className="font-medium text-sm text-primary">
                                                                {invitation.challenged.username} accepted!
                                                            </p>
                                                        </div>
                                                        <p className="text-xs text-muted-foreground mt-1">
                                                            Game #{invitation.gameId} is ready
                                                        </p>
                                                    </div>
                                                )}
                                            </div>
                                            {invitation.status === "pending" ? (
                                                <Button
                                                    size="sm"
                                                    variant="outline"
                                                    onClick={() =>
                                                        handleCancel(invitation.invitationId)
                                                    }
                                                    disabled={
                                                        actionLoading === invitation.invitationId
                                                    }
                                                >
                                                    Cancel
                                                </Button>
                                            ) : (
                                                <Button
                                                    size="sm"
                                                    onClick={() =>
                                                        navigate(`/game/${invitation.gameId}`)
                                                    }
                                                >
                                                    Join Game
                                                </Button>
                                            )}
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Online Players Section */}
                    <div>
                        <div className="flex items-center justify-between mb-3">
                            <div className="flex items-center gap-2">
                                <h2 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                                    Online Players
                                </h2>
                                {onlineUsers.length > 0 && (
                                    <span className="inline-flex items-center justify-center w-5 h-5 text-xs font-medium rounded-full bg-primary text-primary-foreground">
                                        {onlineUsers.length}
                                    </span>
                                )}
                            </div>
                        </div>
                        <div className="bg-card rounded-lg border">
                            {onlineUsers.length === 0 ? (
                                <div className="p-4 text-center">
                                    <p className="text-sm text-muted-foreground">
                                        No other players online
                                    </p>
                                </div>
                            ) : (
                                <div className="divide-y">
                                    {onlineUsers.map((player) => {
                                        const hasPendingInvitation = sentInvitations.some(
                                            (inv) =>
                                                inv.challenged.userId === player.userId &&
                                                inv.status === "pending"
                                        );
                                        return (
                                            <div
                                                key={player.userId}
                                                className="p-4 flex items-center justify-between hover:bg-accent/50 transition-colors"
                                            >
                                                <div className="flex items-center gap-3">
                                                    <div className="w-2 h-2 rounded-full bg-green-500"></div>
                                                    <p className="font-medium text-sm">
                                                        {player.username}
                                                    </p>
                                                </div>
                                                <Button
                                                    size="sm"
                                                    onClick={() => handleChallenge(player.userId)}
                                                    disabled={
                                                        actionLoading === player.userId ||
                                                        hasPendingInvitation
                                                    }
                                                    variant={
                                                        hasPendingInvitation ? "outline" : "default"
                                                    }
                                                >
                                                    {hasPendingInvitation ? "Invited" : "Challenge"}
                                                </Button>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                        </div>
                    </div>
                </div>

                {/* Right Panel - Chat */}
                <div className="w-80 flex-shrink-0">
                    <ChatPanel currentUsername={user?.username || "Guest"} />
                </div>
            </div>
        </div>
    );
}
