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
import ChatPanel from "@/components/common/ChatPanel";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/contexts/AuthContext";
import { useChatContext } from "@/contexts/ChatContext";
import type { GameData } from "@/types/game";
import { type Invitation, type LobbyUser } from "@/types/lobby";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

export default function LobbyPage() {
    const { user, logout } = useAuth();
    const navigate = useNavigate();
    const { messages, connectionStatus, error: chatError, sendMessage } = useChatContext();

    const [onlineUsers, setOnlineUsers] = useState<LobbyUser[]>([]);
    const [sentInvitations, setSentInvitations] = useState<Invitation[]>([]);
    const [receivedInvitations, setReceivedInvitations] = useState<Invitation[]>([]);
    const [activeGames, setActiveGames] = useState<GameData[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [actionLoading, setActionLoading] = useState<number | null>(null);
    const [showingPlayers, setShowingPlayers] = useState(false);

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
            <div className="min-h-screen bg-felt felt-texture flex items-center justify-center">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-4 border-gold/20 border-t-gold mx-auto mb-4" />
                    <h2 className="text-2xl font-display font-bold mb-2 text-gold-light">
                        Joining Lobby...
                    </h2>
                    <p className="text-muted-foreground">Please wait</p>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-felt felt-texture flex flex-col">
            {/* Header */}
            <div className="border-b border-gold/20 bg-card/50 backdrop-blur-sm shadow-raised">
                <div className="max-w-full px-6 py-4 pr-[336px] flex justify-between items-center">
                    <div className="flex items-center gap-4">
                        <div>
                            <h1 className="text-3xl font-display font-bold text-gold-light tracking-wide">
                                Backgammon Lounge
                            </h1>
                            <p className="text-sm font-medium text-foreground mt-0.5">
                                Welcome back,{" "}
                                <span className="text-gold-light font-semibold">
                                    {user?.username}
                                </span>
                            </p>
                        </div>
                    </div>
                    <div className="flex items-center gap-3">
                        <Button onClick={logout} variant="destructive" size="sm">
                            Logout
                        </Button>
                    </div>
                </div>
            </div>

            {/* Error Banner */}
            {error && (
                <div className="bg-destructive/10 border-b border-destructive px-6 py-3">
                    <p className="text-sm text-destructive font-medium">{error}</p>
                </div>
            )}

            {/* Main Content - With right padding for fixed chat */}
            <div className="flex-1 overflow-hidden">
                {/* Main Lobby Content */}
                <div className="h-full p-6 pr-[336px] flex flex-col">
                    {/* Header with toggle */}
                    <div className="flex items-center justify-between mb-4">
                        <div className="flex items-center gap-3">
                            <h2 className="text-xl font-heading text-gold">
                                {showingPlayers ? "Find Players" : "Games"}
                            </h2>
                            {!showingPlayers && (
                                <span className="inline-flex items-center justify-center px-2.5 py-0.5 text-xs font-semibold rounded-full bg-gold/20 text-gold-light border border-gold/40">
                                    {activeGames.length +
                                        receivedInvitations.length +
                                        sentInvitations.length}
                                </span>
                            )}
                            {showingPlayers && onlineUsers.length > 0 && (
                                <span className="inline-flex items-center justify-center px-2.5 py-0.5 text-xs font-semibold rounded-full bg-gold/20 text-gold-light border border-gold/40">
                                    {onlineUsers.length}
                                </span>
                            )}
                        </div>
                        <Button
                            onClick={() => setShowingPlayers(!showingPlayers)}
                            variant="casino"
                            size="sm"
                        >
                            {showingPlayers ? "My Games" : "Find Players"}
                        </Button>
                    </div>

                    {/* Main Card - Fixed height with internal scroll */}
                    <div className="flex-1 bg-black/60 backdrop-blur-sm rounded-xl border-2 border-gold shadow-lg overflow-hidden flex flex-col">
                        <div
                            className="flex-1 overflow-y-auto custom-scrollbar"
                            style={{
                                scrollbarWidth: "thin",
                                scrollbarColor: "#3f3f46 transparent",
                            }}
                        >
                            {!showingPlayers ? (
                                <div className="divide-y divide-gold/20">
                                    {/* Received Invitations - At top with attention-grabbing style */}
                                    {receivedInvitations.map((invitation) => (
                                        <div
                                            key={`received-${invitation.invitationId}`}
                                            className="p-4 flex items-center justify-between transition-all duration-200 bg-gold/10 border-l-4 border-gold hover:bg-gold/15"
                                        >
                                            <div className="flex-1">
                                                <div className="flex items-center gap-2">
                                                    <span className="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold bg-gold text-mahogany">
                                                        GAME INVITATION
                                                    </span>
                                                    <p className="font-semibold text-sm text-gold-light">
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
                                                    className="bg-green-600 hover:bg-green-700 text-white"
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
                                            className="p-4 flex items-center justify-between hover:bg-gold/5 transition-all duration-200"
                                        >
                                            <div className="flex-1">
                                                {invitation.status === "pending" ? (
                                                    <div className="flex items-center gap-2">
                                                        <span className="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold bg-yellow-100 text-yellow-700 dark:bg-yellow-900/40 dark:text-yellow-400">
                                                            Sent
                                                        </span>
                                                        <p className="font-semibold text-sm">
                                                            to {invitation.challenged.username}
                                                        </p>
                                                        <span className="text-xs text-muted-foreground italic">
                                                            Waiting for response...
                                                        </span>
                                                    </div>
                                                ) : (
                                                    <div>
                                                        <div className="flex items-center gap-2">
                                                            <span className="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-400">
                                                                Accepted
                                                            </span>
                                                            <p className="font-semibold text-sm text-green-700 dark:text-green-400">
                                                                {invitation.challenged.username}{" "}
                                                                accepted!
                                                            </p>
                                                        </div>
                                                        <p className="text-xs text-muted-foreground mt-1 ml-1">
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
                                                    variant="casino"
                                                    onClick={() =>
                                                        navigate(`/game/${invitation.gameId}`)
                                                    }
                                                >
                                                    Join Game
                                                </Button>
                                            )}
                                        </div>
                                    ))}

                                    {/* Active Games */}
                                    {activeGames.map((game) => {
                                        const isPlayer1 = game.player1.userId === user?.id;
                                        const opponent = isPlayer1 ? game.player2 : game.player1;
                                        const myColor = isPlayer1
                                            ? game.player1.color
                                            : game.player2.color;

                                        return (
                                            <div
                                                key={game.gameId}
                                                className="p-4 flex items-center justify-between hover:bg-gold/5 transition-all duration-200"
                                            >
                                                <div className="flex-1">
                                                    <p className="font-semibold text-sm">
                                                        vs {opponent.username}
                                                    </p>
                                                    <div className="flex items-center gap-2 mt-1.5">
                                                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-gold/20 text-gold-light border border-gold/40">
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
                                                    variant="casino"
                                                    onClick={() => navigate(`/game/${game.gameId}`)}
                                                >
                                                    Join Game
                                                </Button>
                                            </div>
                                        );
                                    })}

                                    {/* Empty state */}
                                    {activeGames.length === 0 &&
                                        receivedInvitations.length === 0 &&
                                        sentInvitations.length === 0 && (
                                            <div className="p-8 text-center">
                                                <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-gold/10 border border-gold/30 flex items-center justify-center">
                                                    <svg
                                                        xmlns="http://www.w3.org/2000/svg"
                                                        width="32"
                                                        height="32"
                                                        viewBox="0 0 24 24"
                                                        fill="none"
                                                        stroke="currentColor"
                                                        strokeWidth="2"
                                                        strokeLinecap="round"
                                                        strokeLinejoin="round"
                                                        className="text-gold"
                                                    >
                                                        <rect
                                                            x="3"
                                                            y="3"
                                                            width="18"
                                                            height="18"
                                                            rx="2"
                                                        />
                                                        <path d="M3 9h18" />
                                                        <path d="M9 21V9" />
                                                    </svg>
                                                </div>
                                                <p className="text-sm font-medium text-muted-foreground mb-2">
                                                    No activity yet
                                                </p>
                                                <p className="text-xs text-muted-foreground">
                                                    Click "Find Players" to challenge someone!
                                                </p>
                                            </div>
                                        )}
                                </div>
                            ) : (
                                <div className="divide-y divide-gold/20">
                                    {/* Online Players */}
                                    {onlineUsers.length === 0 ? (
                                        <div className="p-8 text-center">
                                            <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-gold/10 border border-gold/30 flex items-center justify-center">
                                                <svg
                                                    xmlns="http://www.w3.org/2000/svg"
                                                    width="32"
                                                    height="32"
                                                    viewBox="0 0 24 24"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    strokeWidth="2"
                                                    strokeLinecap="round"
                                                    strokeLinejoin="round"
                                                    className="text-gold"
                                                >
                                                    <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
                                                    <circle cx="9" cy="7" r="4" />
                                                    <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
                                                    <path d="M16 3.13a4 4 0 0 1 0 7.75" />
                                                </svg>
                                            </div>
                                            <p className="text-sm font-medium text-muted-foreground mb-2">
                                                No other players in the lobby
                                            </p>
                                            <p className="text-xs text-muted-foreground">
                                                Waiting for players to join...
                                            </p>
                                        </div>
                                    ) : (
                                        onlineUsers.map((player) => {
                                            const hasPendingInvitation = sentInvitations.some(
                                                (inv) =>
                                                    inv.challenged.userId === player.userId &&
                                                    inv.status === "pending"
                                            );
                                            return (
                                                <div
                                                    key={player.userId}
                                                    className="p-4 flex items-center justify-between hover:bg-gold/5 transition-all duration-200"
                                                >
                                                    <div className="flex items-center gap-3">
                                                        <div className="w-10 h-10 rounded-full bg-gradient-to-br from-gold to-gold-dark flex items-center justify-center flex-shrink-0 shadow-poker-chip ring-2 ring-gold-light/40 hover:ring-4 hover:ring-gold/60 transition-all">
                                                            <span className="text-sm font-bold text-mahogany">
                                                                {player.username
                                                                    .slice(0, 2)
                                                                    .toUpperCase()}
                                                            </span>
                                                        </div>
                                                        <div className="flex items-center gap-2">
                                                            <div className="relative">
                                                                <p className="font-semibold text-sm">
                                                                    {player.username}
                                                                </p>
                                                                <div className="absolute -top-0.5 -right-3">
                                                                    <div className="w-2.5 h-2.5 rounded-full bg-green-500 ring-2 ring-card"></div>
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </div>
                                                    <Button
                                                        size="sm"
                                                        onClick={() =>
                                                            handleChallenge(player.userId)
                                                        }
                                                        disabled={
                                                            actionLoading === player.userId ||
                                                            hasPendingInvitation
                                                        }
                                                        variant={
                                                            hasPendingInvitation
                                                                ? "outline"
                                                                : "casino"
                                                        }
                                                    >
                                                        {hasPendingInvitation
                                                            ? "Invited"
                                                            : "Challenge"}
                                                    </Button>
                                                </div>
                                            );
                                        })
                                    )}
                                </div>
                            )}
                        </div>
                    </div>
                </div>

                {/* Fixed Chat Panel on Right */}
                <div className="fixed top-0 right-0 bottom-0 w-80 border-l border-gold/20 bg-card/30 backdrop-blur-sm z-40">
                    <ChatPanel
                        currentUsername={user?.username || "Guest"}
                        messages={messages}
                        connectionStatus={connectionStatus}
                        error={chatError}
                        sendMessage={sendMessage}
                    />
                </div>
            </div>
        </div>
    );
}
