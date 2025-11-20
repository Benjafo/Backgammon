export interface ChatMessage {
  messageId: number;
  userId: number;
  username: string;
  message: string;
  timestamp: string;
}

export interface IncomingMessage {
  type: 'message' | 'history';
  messageId?: number;
  userId?: number;
  username?: string;
  message?: string;
  timestamp?: string;
  messages?: ChatMessage[];
}

export interface OutgoingMessage {
  type: 'chat';
  message: string;
}
