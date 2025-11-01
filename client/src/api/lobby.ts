import {
    type AcceptInvitationResponse,
    type CreateInvitationResponse,
    type InvitationsResponse,
    type LobbyUser,
} from "../types/lobby";

const API_BASE = "/api/v1";

// Lobby presence functions

export async function joinLobby(): Promise<void> {
    const response = await fetch(`${API_BASE}/lobby/presence`, {
        method: "POST",
        credentials: "include",
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to join lobby");
    }
}

export async function leaveLobby(): Promise<void> {
    const response = await fetch(`${API_BASE}/lobby/presence`, {
        method: "DELETE",
        credentials: "include",
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to leave lobby");
    }
}

export async function sendHeartbeat(): Promise<void> {
    const response = await fetch(`${API_BASE}/lobby/presence/heartbeat`, {
        method: "PUT",
        credentials: "include",
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to send heartbeat");
    }
}

export async function getLobbyUsers(): Promise<LobbyUser[]> {
    const response = await fetch(`${API_BASE}/lobby/users`, {
        method: "GET",
        credentials: "include",
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to get lobby users");
    }

    const data = await response.json();
    return data.users || [];
}

// Invitation functions

export async function sendInvitation(challengedId: number): Promise<CreateInvitationResponse> {
    const response = await fetch(`${API_BASE}/invitations`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({ challengedId }),
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to send invitation");
    }

    return response.json();
}

export async function getInvitations(): Promise<InvitationsResponse> {
    const response = await fetch(`${API_BASE}/invitations`, {
        method: "GET",
        credentials: "include",
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to get invitations");
    }

    return response.json();
}

export async function acceptInvitation(invitationId: number): Promise<AcceptInvitationResponse> {
    const response = await fetch(`${API_BASE}/invitations/${invitationId}/accept`, {
        method: "PUT",
        credentials: "include",
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to accept invitation");
    }

    return response.json();
}

export async function declineInvitation(invitationId: number): Promise<void> {
    const response = await fetch(`${API_BASE}/invitations/${invitationId}/decline`, {
        method: "PUT",
        credentials: "include",
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to decline invitation");
    }
}

export async function cancelInvitation(invitationId: number): Promise<void> {
    const response = await fetch(`${API_BASE}/invitations/${invitationId}`, {
        method: "DELETE",
        credentials: "include",
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Failed to cancel invitation");
    }
}
