import { type ConnectionStatus, type WSMessage } from "@/types/chat";
import { useCallback, useEffect, useRef, useState } from "react";

const INITIAL_RETRY_DELAY = 1000; // 1 second
const MAX_RETRY_DELAY = 30000; // 30 seconds
const RETRY_BACKOFF_FACTOR = 1.5;

interface UseGameWebSocketOptions {
    gameId: number | null;
    onMessage: (message: WSMessage) => void;
    onOpen?: () => void;
    onClose?: () => void;
    onError?: (error: Event) => void;
    enabled?: boolean; // If false, don't connect
}

export function useGameWebSocket({
    gameId,
    onMessage,
    onOpen,
    onClose,
    onError,
    enabled = true,
}: UseGameWebSocketOptions) {
    const [status, setStatus] = useState<ConnectionStatus>("disconnected");
    const ws = useRef<WebSocket | null>(null);
    const reconnectTimeout = useRef<ReturnType<typeof setInterval> | null>(null);
    const retryDelay = useRef(INITIAL_RETRY_DELAY);
    const shouldReconnect = useRef(true);
    const isManualClose = useRef(false);

    const connect = useCallback(() => {
        if (!gameId) {
            console.warn("Cannot connect: gameId is null");
            return;
        }

        if (
            ws.current?.readyState === WebSocket.OPEN ||
            ws.current?.readyState === WebSocket.CONNECTING
        ) {
            return;
        }

        setStatus("connecting");

        try {
            // Construct WebSocket URL for game chat
            const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
            const wsUrl = `${protocol}//${window.location.host}/api/v1/games/${gameId}/ws`;

            ws.current = new WebSocket(wsUrl);

            ws.current.onopen = () => {
                console.log(`Game chat WebSocket connected (game ${gameId})`);
                setStatus("connected");
                retryDelay.current = INITIAL_RETRY_DELAY; // Reset retry delay on successful connection
                if (onOpen) onOpen();
            };

            ws.current.onmessage = (event) => {
                try {
                    const message: WSMessage = JSON.parse(event.data);
                    onMessage(message);
                } catch (error) {
                    console.error("Error parsing WebSocket message:", error);
                }
            };

            ws.current.onclose = (event) => {
                console.log("Game chat WebSocket disconnected", event.code, event.reason);
                setStatus("disconnected");
                ws.current = null;

                if (onClose) onClose();

                // Reconnect if not a manual close and shouldReconnect is true
                if (shouldReconnect.current && !isManualClose.current) {
                    console.log(`Reconnecting in ${retryDelay.current}ms...`);
                    reconnectTimeout.current = setTimeout(() => {
                        retryDelay.current = Math.min(
                            retryDelay.current * RETRY_BACKOFF_FACTOR,
                            MAX_RETRY_DELAY
                        );
                        connect();
                    }, retryDelay.current);
                }
            };

            ws.current.onerror = (error) => {
                console.error("Game chat WebSocket error:", error);
                setStatus("error");
                if (onError) onError(error);
            };
        } catch (error) {
            console.error("Error creating game chat WebSocket:", error);
            setStatus("error");
        }
    }, [gameId, onMessage, onOpen, onClose, onError]);

    const disconnect = useCallback(() => {
        isManualClose.current = true;
        shouldReconnect.current = false;

        if (reconnectTimeout.current) {
            clearTimeout(reconnectTimeout.current);
            reconnectTimeout.current = null;
        }

        if (ws.current) {
            ws.current.close();
            ws.current = null;
        }

        setStatus("disconnected");
    }, []);

    const sendMessage = useCallback((message: WSMessage) => {
        if (ws.current?.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify(message));
            return true;
        } else {
            console.warn("Game chat WebSocket not connected, cannot send message");
            return false;
        }
    }, []);

    // Connect on mount (only if enabled and gameId is provided)
    useEffect(() => {
        if (!enabled || !gameId) {
            // If disabled or no gameId, disconnect if currently connected
            if (ws.current) {
                shouldReconnect.current = false;
                ws.current.close();
                ws.current = null;
            }
            setStatus("disconnected");
            return;
        }

        isManualClose.current = false;
        shouldReconnect.current = true;
        connect();

        // Cleanup on unmount or when gameId changes
        return () => {
            shouldReconnect.current = false;
            if (reconnectTimeout.current) {
                clearTimeout(reconnectTimeout.current);
            }
            if (ws.current) {
                ws.current.close();
            }
        };
    }, [connect, enabled, gameId]);

    return {
        status,
        sendMessage,
        disconnect,
        reconnect: connect,
    };
}
