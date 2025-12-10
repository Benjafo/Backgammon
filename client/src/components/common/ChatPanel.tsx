import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { ChatMessage, ConnectionStatus } from "@/types/chat";
import { useEffect, useRef, useState } from "react";

interface ChatPanelProps {
    currentUsername: string;
    messages: ChatMessage[];
    connectionStatus: ConnectionStatus;
    error: string | null;
    sendMessage: (message: string) => boolean;
    title?: string;
    subtitle?: string;
}

export default function ChatPanel({
    currentUsername,
    messages,
    connectionStatus,
    error,
    sendMessage,
    title = "Lounge Chat",
    subtitle = "Chat with other players",
}: ChatPanelProps) {
    const [newMessage, setNewMessage] = useState("");
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const messagesContainerRef = useRef<HTMLDivElement>(null);

    // Auto-scroll to bottom when new messages arrive (use auto to prevent infinite loop)
    useEffect(() => {
        if (messagesEndRef.current) {
            messagesEndRef.current.scrollIntoView({ behavior: "auto" });
        }
    }, [messages]);

    const handleSendMessage = (e: React.FormEvent) => {
        e.preventDefault();
        if (!newMessage.trim() || connectionStatus !== "connected") return;

        const success = sendMessage(newMessage);
        if (success) {
            setNewMessage("");
        }
    };

    const formatTime = (timestamp: string) => {
        const date = new Date(timestamp);
        return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    };

    const getUserInitials = (username: string) => {
        return username.slice(0, 2).toUpperCase();
    };

    const getConnectionStatusColor = () => {
        switch (connectionStatus) {
            case "connected":
                return "bg-green-500";
            case "connecting":
                return "bg-yellow-500 animate-pulse";
            case "disconnected":
                return "bg-red-500";
            case "error":
                return "bg-red-600";
            default:
                return "bg-gray-500";
        }
    };

    const getConnectionStatusText = () => {
        switch (connectionStatus) {
            case "connected":
                return "Connected";
            case "connecting":
                return "Connecting...";
            case "disconnected":
                return "Disconnected";
            case "error":
                return "Error";
            default:
                return "Unknown";
        }
    };

    return (
        <div className="flex flex-col h-full bg-black/50 backdrop-blur-sm overflow-hidden">
            {/* Header */}
            <div className="border-b border-gold/20 px-4 py-4 bg-black/40">
                <div className="flex items-center justify-between">
                    <div>
                        <h2 className="font-heading font-bold text-lg text-gold-light">
                            {title}
                        </h2>
                        <p className="text-xs text-muted-foreground mt-1">
                            {subtitle}
                        </p>
                    </div>
                    {/* Connection Status Indicator */}
                    <div className="flex items-center gap-2">
                        <div className={`w-2 h-2 rounded-full ${getConnectionStatusColor()}`} />
                        <span className="text-xs text-muted-foreground">
                            {getConnectionStatusText()}
                        </span>
                    </div>
                </div>
                {/* Error Message */}
                {error && (
                    <div className="mt-2 px-2 py-1 bg-red-500/20 border border-red-500/50 rounded text-xs text-red-200">
                        {error}
                    </div>
                )}
            </div>

            {/* Messages Area */}
            <div
                ref={messagesContainerRef}
                className="flex-1 overflow-y-auto overflow-x-hidden p-4 space-y-4 custom-scrollbar"
                style={{
                    scrollbarWidth: "thin",
                    scrollbarColor: "#3f3f46 transparent",
                }}
            >
                {messages.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full">
                        <div className="w-16 h-16 mb-4 rounded-full bg-gold/10 border border-gold/30 flex items-center justify-center">
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
                                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
                                <path d="M9 10h6" />
                                <path d="M9 14h3" />
                            </svg>
                        </div>
                        <p className="text-sm font-medium text-muted-foreground text-center mb-1">
                            {connectionStatus === "connected"
                                ? "No messages yet"
                                : "Connecting to chat..."}
                        </p>
                        <p className="text-xs text-muted-foreground text-center px-8">
                            {connectionStatus === "connected"
                                ? "Be the first to start a conversation!"
                                : "Please wait while we connect you to the chat."}
                        </p>
                    </div>
                ) : (
                    <>
                        {messages.map((message) => (
                            <div key={message.messageId} className="flex gap-3">
                                {/* Avatar */}
                                <div className="w-8 h-8 rounded-full bg-gradient-to-br from-gold to-gold-dark flex items-center justify-center flex-shrink-0 shadow-poker-chip ring-2 ring-gold-light/30">
                                    <span className="text-xs font-bold text-mahogany">
                                        {getUserInitials(message.username)}
                                    </span>
                                </div>
                                {/* Message Content */}
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-baseline gap-2 mb-1">
                                        <span className="text-sm font-semibold text-foreground">
                                            {message.username}
                                        </span>
                                        {message.username === currentUsername && (
                                            <span className="text-xs text-gold/70">(You)</span>
                                        )}
                                        <span className="text-xs text-muted-foreground">
                                            {formatTime(message.timestamp)}
                                        </span>
                                    </div>
                                    <p className="text-sm text-foreground/90 break-words leading-relaxed">
                                        {message.message}
                                    </p>
                                </div>
                            </div>
                        ))}
                        <div ref={messagesEndRef} />
                    </>
                )}
            </div>

            {/* Input Area */}
            <div className="border-t border-gold/20 p-4 bg-black/40">
                {connectionStatus !== "connected" && (
                    <div className="mb-2 text-xs text-center text-yellow-500">
                        {connectionStatus === "connecting"
                            ? "Connecting..."
                            : "Reconnecting to chat..."}
                    </div>
                )}
                <form onSubmit={handleSendMessage} className="flex gap-2">
                    <Input
                        type="text"
                        placeholder={
                            connectionStatus === "connected"
                                ? "Type a message..."
                                : "Waiting for connection..."
                        }
                        value={newMessage}
                        onChange={(e) => setNewMessage(e.target.value)}
                        disabled={connectionStatus !== "connected"}
                        className="flex-1"
                        maxLength={1000}
                    />
                    <Button
                        type="submit"
                        size="sm"
                        variant="casino"
                        className="px-3"
                        disabled={connectionStatus !== "connected" || !newMessage.trim()}
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            strokeWidth="2"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                        >
                            <path d="M22 2 11 13" />
                            <path d="M22 2 15 22 11 13 2 9z" />
                        </svg>
                    </Button>
                </form>
            </div>
        </div>
    );
}
