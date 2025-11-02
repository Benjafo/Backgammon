import type { ActiveGamesResponse, GameData } from "@/types/game";

const API_BASE = '/api/v1';

// Get game details
export async function getGame(gameId: number): Promise<GameData> {
  const response = await fetch(`${API_BASE}/games/${gameId}`, {
    method: 'GET',
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to get game');
  }

  return response.json();
}

// Forfeit a game
export async function forfeitGame(gameId: number): Promise<void> {
  const response = await fetch(`${API_BASE}/games/${gameId}/forfeit`, {
    method: 'POST',
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to forfeit game');
  }
}

// Start a game (change status from pending to in_progress)
export async function startGame(gameId: number): Promise<void> {
  const response = await fetch(`${API_BASE}/games/${gameId}/start`, {
    method: 'POST',
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to start game');
  }
}

// Get all active games for the current user
export async function getActiveGames(): Promise<GameData[]> {
  const response = await fetch(`${API_BASE}/games/active`, {
    method: 'GET',
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to get active games');
  }

  const data: ActiveGamesResponse = await response.json();
  return data.games;
}
