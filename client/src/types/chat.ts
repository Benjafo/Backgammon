export interface ChatMessage {
    messageId: number;
    userId: number;
    username: string;
    message: string;
    timestamp: string; // ISO 8601 format
}

export interface WSMessage {
    type: "send_message" | "chat_message" | "history" | "user_joined" | "user_left" | "error";
    data: any;
}

export interface SendMessageRequest {
    message: string;
}

export interface ChatMessageData {
    messageId: number;
    userId: number;
    username: string;
    message: string;
    timestamp: string;
}

export interface MessageHistoryData {
    messages: ChatMessageData[];
}

export interface UserEventData {
    userId: number;
    username: string;
}

export interface ErrorData {
    message: string;
}

export type ConnectionStatus = "connecting" | "connected" | "disconnected" | "error";
