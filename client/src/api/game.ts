import type { ActiveGamesResponse, GameData, GameState, LegalMove, LegalMovesResponse, MoveRequest } from "@/types/game";

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

// Get current game state
export async function getGameState(gameId: number): Promise<GameState> {
  const response = await fetch(`${API_BASE}/games/${gameId}/state`, {
    method: 'GET',
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to get game state');
  }

  return response.json();
}

// Roll dice for current turn
export async function rollDice(gameId: number): Promise<GameState> {
  const response = await fetch(`${API_BASE}/games/${gameId}/roll`, {
    method: 'POST',
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to roll dice');
  }

  return response.json();
}

// Execute a move
export async function makeMove(gameId: number, move: MoveRequest): Promise<GameState> {
  const response = await fetch(`${API_BASE}/games/${gameId}/move`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(move),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to make move');
  }

  return response.json();
}

// Get legal moves for current position
export async function getLegalMoves(gameId: number): Promise<LegalMove[]> {
  const response = await fetch(`${API_BASE}/games/${gameId}/legal-moves`, {
    method: 'GET',
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to get legal moves');
  }

  const data: LegalMovesResponse = await response.json();
  return data.moves;
}
