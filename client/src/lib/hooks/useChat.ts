import { useCallback, useMemo, useState } from "react";
import type { ChatMessage, IncomingMessage } from "../../types/chat";
import { useWebSocket, type WebSocketMessage } from "./useWebSocket";

interface UseChatReturn {
    messages: ChatMessage[];
    isConnected: boolean;
    sendMessage: (message: string) => void;
}

export const useChat = (): UseChatReturn => {
    const [messages, setMessages] = useState<ChatMessage[]>([]);

    const handleMessage = useCallback((message: WebSocketMessage) => {
        const incomingMsg = message as IncomingMessage;

        if (incomingMsg.type === "message") {
            // Single new message
            const chatMessage: ChatMessage = {
                messageId: incomingMsg.messageId!,
                userId: incomingMsg.userId!,
                username: incomingMsg.username!,
                message: incomingMsg.message!,
                timestamp: incomingMsg.timestamp!,
            };
            setMessages((prev) => [...prev, chatMessage]);
        } else if (incomingMsg.type === "history") {
            // Message history (received on connection)
            const historyMessages = incomingMsg.messages || [];
            setMessages(historyMessages);
        }
    }, []);

    const handleOpen = useCallback(() => {
        console.log("Chat WebSocket connected");
    }, []);

    const handleClose = useCallback(() => {
        console.log("Chat WebSocket disconnected");
    }, []);

    const handleError = useCallback((error: Event) => {
        console.error("Chat WebSocket error:", error);
    }, []);

    const websocketOptions = useMemo(
        () => ({
            onMessage: handleMessage,
            onOpen: handleOpen,
            onClose: handleClose,
            onError: handleError,
        }),
        [handleMessage, handleOpen, handleClose, handleError]
    );

    const { isConnected, sendMessage: sendWebSocketMessage } = useWebSocket(
        "/api/v1/chat/ws",
        websocketOptions
    );

    const sendMessage = useCallback(
        (message: string) => {
            if (!message.trim()) return;

            sendWebSocketMessage({
                type: "chat",
                message: message.trim(),
            });
        },
        [sendWebSocketMessage]
    );

    return {
        messages,
        isConnected,
        sendMessage,
    };
};
