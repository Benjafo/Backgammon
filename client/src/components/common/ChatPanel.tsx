import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useState } from "react";

interface Message {
    id: number;
    username: string;
    content: string;
    timestamp: Date;
}

interface ChatPanelProps {
    currentUsername: string;
}

export default function ChatPanel({ currentUsername }: ChatPanelProps) {
    const [messages, setMessages] = useState<Message[]>([]);
    const [newMessage, setNewMessage] = useState("");

    const handleSendMessage = (e: React.FormEvent) => {
        e.preventDefault();
        if (!newMessage.trim()) return;

        // TODO: Implement actual chat API call
        const message: Message = {
            id: Date.now(),
            username: currentUsername,
            content: newMessage,
            timestamp: new Date(),
        };

        setMessages([...messages, message]);
        setNewMessage("");
    };

    const formatTime = (date: Date) => {
        return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    };

    const getUserInitials = (username: string) => {
        return username.slice(0, 2).toUpperCase();
    };

    return (
        <div className="flex flex-col h-full">
            {/* Header */}
            <div className="border-b px-4 py-4 bg-card">
                <h2 className="font-bold text-base">Lobby Chat</h2>
                <p className="text-xs text-muted-foreground mt-1">
                    Chat with other players
                </p>
            </div>

            {/* Messages Area */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-background/50">
                {messages.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full">
                        <div className="w-16 h-16 mb-4 rounded-full bg-amber-100 dark:bg-amber-900/20 flex items-center justify-center">
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
                                className="text-amber-600 dark:text-amber-400"
                            >
                                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
                                <path d="M9 10h6" />
                                <path d="M9 14h3" />
                            </svg>
                        </div>
                        <p className="text-sm font-medium text-muted-foreground text-center mb-1">
                            No messages yet
                        </p>
                        <p className="text-xs text-muted-foreground text-center px-8">
                            Be the first to start a conversation!
                        </p>
                    </div>
                ) : (
                    messages.map((message) => (
                        <div key={message.id} className="flex gap-3">
                            {/* Avatar */}
                            <div className="w-8 h-8 rounded-full bg-gradient-to-br from-amber-400 to-amber-600 flex items-center justify-center flex-shrink-0 shadow-sm">
                                <span className="text-xs font-bold text-white">
                                    {getUserInitials(message.username)}
                                </span>
                            </div>
                            {/* Message Content */}
                            <div className="flex-1 min-w-0">
                                <div className="flex items-baseline gap-2 mb-1">
                                    <span className="text-sm font-semibold text-foreground">
                                        {message.username}
                                    </span>
                                    <span className="text-xs text-muted-foreground">
                                        {formatTime(message.timestamp)}
                                    </span>
                                </div>
                                <p className="text-sm text-foreground/90 break-words leading-relaxed">
                                    {message.content}
                                </p>
                            </div>
                        </div>
                    ))
                )}
            </div>

            {/* Input Area */}
            <div className="border-t p-4 bg-card">
                <form onSubmit={handleSendMessage} className="flex gap-2">
                    <Input
                        type="text"
                        placeholder="Type a message..."
                        value={newMessage}
                        onChange={(e) => setNewMessage(e.target.value)}
                        className="flex-1"
                    />
                    <Button
                        type="submit"
                        size="sm"
                        className="bg-amber-600 hover:bg-amber-700 text-white px-3"
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
