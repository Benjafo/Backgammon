import { type ConnectionStatus, type WSMessage } from "@/types/chat";
import { useCallback, useEffect, useRef, useState } from "react";

const WS_URL = "/api/v1/lobby/ws";
const INITIAL_RETRY_DELAY = 1000; // 1 second
const MAX_RETRY_DELAY = 30000; // 30 seconds
const RETRY_BACKOFF_FACTOR = 1.5;

interface UseWebSocketOptions {
    onMessage: (message: WSMessage) => void;
    onOpen?: () => void;
    onClose?: () => void;
    onError?: (error: Event) => void;
    enabled?: boolean; // If false, don't connect
}

export function useWebSocket({ onMessage, onOpen, onClose, onError, enabled = true }: UseWebSocketOptions) {
    const [status, setStatus] = useState<ConnectionStatus>("disconnected");
    const ws = useRef<WebSocket | null>(null);
    const reconnectTimeout = useRef<ReturnType<typeof setInterval> | null>(null);
    const retryDelay = useRef(INITIAL_RETRY_DELAY);
    const shouldReconnect = useRef(true);
    const isManualClose = useRef(false);
    const isConnecting = useRef(false); // Prevent duplicate connections in StrictMode

    const connect = useCallback(() => {
        if (
            ws.current?.readyState === WebSocket.OPEN ||
            ws.current?.readyState === WebSocket.CONNECTING ||
            isConnecting.current
        ) {
            return;
        }

        isConnecting.current = true;

        setStatus("connecting");

        try {
            // Construct WebSocket URL (use wss:// for https, ws:// for http)
            const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
            const wsUrl = `${protocol}//${window.location.host}${WS_URL}`;

            ws.current = new WebSocket(wsUrl);

            ws.current.onopen = () => {
                console.log("WebSocket connected");
                setStatus("connected");
                retryDelay.current = INITIAL_RETRY_DELAY; // Reset retry delay on successful connection
                isConnecting.current = false;
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
                console.log("WebSocket disconnected", event.code, event.reason);
                setStatus("disconnected");
                ws.current = null;
                isConnecting.current = false;

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
                console.error("WebSocket error:", error);
                setStatus("error");
                if (onError) onError(error);
            };
        } catch (error) {
            console.error("Error creating WebSocket:", error);
            setStatus("error");
            isConnecting.current = false;
        }
    }, [onMessage, onOpen, onClose, onError]);

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
            console.warn("WebSocket not connected, cannot send message");
            return false;
        }
    }, []);

    // Connect on mount (only if enabled)
    useEffect(() => {
        if (!enabled) {
            // If disabled, disconnect if currently connected
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

        // Cleanup on unmount
        return () => {
            shouldReconnect.current = false;
            if (reconnectTimeout.current) {
                clearTimeout(reconnectTimeout.current);
            }
            if (ws.current) {
                ws.current.close();
            }
        };
    }, [connect, enabled]);

    return {
        status,
        sendMessage,
        disconnect,
        reconnect: connect,
    };
}
