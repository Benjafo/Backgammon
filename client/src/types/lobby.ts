export interface LobbyUser {
  userId: number;
  username: string;
  joinedAt: string;
  lastHeartbeat: string;
}

export interface Invitation {
  invitationId: number;
  challenger: {
    userId: number;
    username: string;
  };
  challenged: {
    userId: number;
    username: string;
  };
  status: 'pending' | 'accepted' | 'declined' | 'expired';
  createdAt: string;
}

export interface InvitationsResponse {
  sent: Invitation[];
  received: Invitation[];
}

export interface CreateInvitationResponse {
  invitationId: number;
  challengedId: number;
  status: string;
  message: string;
}

export interface AcceptInvitationResponse {
  message: string;
  gameId: number;
}
