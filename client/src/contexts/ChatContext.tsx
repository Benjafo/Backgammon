import { useWebSocket } from "@/hooks/useWebSocket";
import {
    type ChatMessage,
    type ChatMessageData,
    type ConnectionStatus,
    type ErrorData,
    type MessageHistoryData,
    type UserEventData,
    type WSMessage,
} from "@/types/chat";
import { useAuth } from "@/contexts/AuthContext";
import { createContext, type ReactNode, useCallback, useContext, useState } from "react";

interface ChatContextType {
    messages: ChatMessage[];
    connectionStatus: ConnectionStatus;
    error: string | null;
    sendMessage: (message: string) => boolean;
}

const ChatContext = createContext<ChatContextType | undefined>(undefined);

interface ChatProviderProps {
    children: ReactNode;
}

export function ChatProvider({ children }: ChatProviderProps) {
    const { user } = useAuth();
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [error, setError] = useState<string | null>(null);

    const handleMessage = useCallback((wsMessage: WSMessage) => {
        switch (wsMessage.type) {
            case "history": {
                const data = wsMessage.data as MessageHistoryData;
                const historyMessages: ChatMessage[] = data.messages.map(
                    (msg: ChatMessageData) => ({
                        messageId: msg.messageId,
                        userId: msg.userId,
                        username: msg.username,
                        message: msg.message,
                        timestamp: msg.timestamp,
                    })
                );
                setMessages(historyMessages);
                break;
            }

            case "chat_message": {
                const data = wsMessage.data as ChatMessageData;
                const newMessage: ChatMessage = {
                    messageId: data.messageId,
                    userId: data.userId,
                    username: data.username,
                    message: data.message,
                    timestamp: data.timestamp,
                };
                setMessages((prev) => [...prev, newMessage]);
                break;
            }

            case "user_joined": {
                const data = wsMessage.data as UserEventData;
                console.log(`User ${data.username} joined the chat`);
                // Optionally add a system message
                // const systemMessage: ChatMessage = {
                //     messageId: Date.now(),
                //     userId: 0,
                //     username: "System",
                //     message: `${data.username} joined the lobby`,
                //     timestamp: new Date().toISOString(),
                // };
                // setMessages((prev) => [...prev, systemMessage]);
                break;
            }

            case "user_left": {
                const data = wsMessage.data as UserEventData;
                console.log(`User ${data.username} left the chat`);
                // Optionally add a system message
                break;
            }

            case "error": {
                const data = wsMessage.data as ErrorData;
                setError(data.message);
                console.error("Chat error:", data.message);
                // Clear error after 5 seconds
                setTimeout(() => setError(null), 5000);
                break;
            }

            default:
                console.warn("Unknown message type:", wsMessage.type);
        }
    }, []);

    const handleOpen = useCallback(() => {
        console.log("Chat WebSocket connected");
        setError(null);
    }, []);

    const handleClose = useCallback(() => {
        console.log("Chat WebSocket disconnected");
    }, []);

    const handleError = useCallback((error: Event) => {
        console.error("Chat WebSocket error:", error);
        setError("Connection error. Reconnecting...");
    }, []);

    const { status, sendMessage: wsSendMessage } = useWebSocket({
        onMessage: handleMessage,
        onOpen: handleOpen,
        onClose: handleClose,
        onError: handleError,
        enabled: !!user, // Only connect when user is authenticated
    });

    const sendMessage = useCallback(
        (message: string): boolean => {
            if (!message.trim()) {
                return false;
            }

            const wsMessage: WSMessage = {
                type: "send_message",
                data: {
                    message: message.trim(),
                },
            };

            return wsSendMessage(wsMessage);
        },
        [wsSendMessage]
    );

    return (
        <ChatContext.Provider
            value={{
                messages,
                connectionStatus: status,
                error,
                sendMessage,
            }}
        >
            {children}
        </ChatContext.Provider>
    );
}

export function useChatContext() {
    const context = useContext(ChatContext);
    if (context === undefined) {
        throw new Error("useChatContext must be used within a ChatProvider");
    }
    return context;
}
