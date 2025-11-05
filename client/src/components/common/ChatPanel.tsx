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
        <div className="flex flex-col h-full border-l">
            {/* Header */}
            <div className="border-b px-4 py-3">
                <h2 className="font-semibold text-sm">LOBBY CHAT</h2>
                <p className="text-xs text-muted-foreground mt-0.5">
                    Talk with players online
                </p>
            </div>

            {/* Messages Area */}
            <div className="flex-1 overflow-y-auto p-4 space-y-3">
                {messages.length === 0 ? (
                    <div className="flex items-center justify-center h-full">
                        <p className="text-sm text-muted-foreground text-center">
                            No messages yet.
                            <br />
                            Start a conversation!
                        </p>
                    </div>
                ) : (
                    messages.map((message) => (
                        <div key={message.id} className="flex gap-2">
                            {/* Avatar */}
                            <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                                <span className="text-xs font-semibold text-primary">
                                    {getUserInitials(message.username)}
                                </span>
                            </div>
                            {/* Message Content */}
                            <div className="flex-1 min-w-0">
                                <div className="flex items-baseline gap-2">
                                    <span className="text-sm font-semibold">
                                        {message.username}
                                    </span>
                                    <span className="text-xs text-muted-foreground">
                                        {formatTime(message.timestamp)}
                                    </span>
                                </div>
                                <p className="text-sm mt-0.5 break-words">
                                    {message.content}
                                </p>
                            </div>
                        </div>
                    ))
                )}
            </div>

            {/* Input Area */}
            <div className="border-t p-4">
                <form onSubmit={handleSendMessage} className="flex gap-2">
                    <Input
                        type="text"
                        placeholder="Type a message..."
                        value={newMessage}
                        onChange={(e) => setNewMessage(e.target.value)}
                        className="flex-1"
                    />
                    <Button type="submit" size="sm">
                        Send
                    </Button>
                </form>
            </div>
        </div>
    );
}
