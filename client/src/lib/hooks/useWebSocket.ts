import { useCallback, useEffect, useRef, useState } from "react";

export interface WebSocketMessage {
    type: string;
    [key: string]: unknown;
}

interface UseWebSocketOptions {
    onOpen?: () => void;
    onClose?: () => void;
    onError?: (error: Event) => void;
    onMessage?: (message: WebSocketMessage) => void;
    reconnectInterval?: number;
    maxReconnectAttempts?: number;
}

interface UseWebSocketReturn {
    isConnected: boolean;
    sendMessage: (message: WebSocketMessage) => void;
    disconnect: () => void;
    reconnect: () => void;
}

export function useWebSocket(url: string, options: UseWebSocketOptions = {}): UseWebSocketReturn {
    const {
        onOpen,
        onClose,
        onError,
        onMessage,
        reconnectInterval = 3000,
        maxReconnectAttempts = 5,
    } = options;

    const [isConnected, setIsConnected] = useState(false);
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectAttemptsRef = useRef(0);
    const reconnectTimeoutRef = useRef<number | null>(null);
    const shouldConnectRef = useRef(true);

    // Use refs for callbacks to avoid recreating connect function
    const onOpenRef = useRef(onOpen);
    const onCloseRef = useRef(onClose);
    const onErrorRef = useRef(onError);
    const onMessageRef = useRef(onMessage);

    // Update refs when callbacks change
    useEffect(() => {
        onOpenRef.current = onOpen;
    }, [onOpen]);

    useEffect(() => {
        onCloseRef.current = onClose;
    }, [onClose]);

    useEffect(() => {
        onErrorRef.current = onError;
    }, [onError]);

    useEffect(() => {
        onMessageRef.current = onMessage;
    }, [onMessage]);

    const connect = useCallback(() => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
            return;
        }

        try {
            // Determine protocol based on current page protocol
            const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
            const wsUrl =
                url.startsWith("ws://") || url.startsWith("wss://")
                    ? url
                    : `${protocol}//${window.location.host}${url}`;

            const ws = new WebSocket(wsUrl);
            wsRef.current = ws;

            ws.onopen = () => {
                setIsConnected(true);
                reconnectAttemptsRef.current = 0;
                onOpenRef.current?.();
            };

            ws.onclose = () => {
                setIsConnected(false);
                wsRef.current = null;
                onCloseRef.current?.();

                // Attempt to reconnect if we should be connected
                if (
                    shouldConnectRef.current &&
                    reconnectAttemptsRef.current < maxReconnectAttempts
                ) {
                    reconnectAttemptsRef.current++;
                    reconnectTimeoutRef.current = window.setTimeout(() => {
                        connect();
                    }, reconnectInterval);
                }
            };

            ws.onerror = (error) => {
                onErrorRef.current?.(error);
            };

            ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data) as WebSocketMessage;
                    onMessageRef.current?.(message);
                } catch (error) {
                    console.error("Failed to parse WebSocket message:", error);
                }
            };
        } catch (error) {
            console.error("Failed to create WebSocket connection:", error);
        }
    }, [url, reconnectInterval, maxReconnectAttempts]);

    const disconnect = useCallback(() => {
        shouldConnectRef.current = false;
        if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = null;
        }
        if (wsRef.current) {
            wsRef.current.close();
            wsRef.current = null;
        }
        setIsConnected(false);
    }, []);

    const reconnect = useCallback(() => {
        disconnect();
        shouldConnectRef.current = true;
        reconnectAttemptsRef.current = 0;
        connect();
    }, [connect, disconnect]);

    const sendMessage = useCallback((message: WebSocketMessage) => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify(message));
        } else {
            console.warn("WebSocket is not connected. Message not sent:", message);
        }
    }, []);

    useEffect(() => {
        connect();

        return () => {
            shouldConnectRef.current = false;
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
            if (wsRef.current) {
                wsRef.current.close();
            }
        };
    }, [connect]);

    return {
        isConnected,
        sendMessage,
        disconnect,
        reconnect,
    };
}
