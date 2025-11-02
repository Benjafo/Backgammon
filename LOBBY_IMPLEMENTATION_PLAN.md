aaaaaaaaaaa# Lobby System Implementation Plan

## Overview
Complete implementation of the lobby and invitation system for the Backgammon game, allowing users to see other online players and challenge them to games.

---

## Phase 1: Repository Layer

### 1.1 Lobby Presence Repository (`repository/lobby.go`)

#### Requirements:
- **JoinLobby(ctx, userID)**
  - Insert new presence record with `joined_at` and `last_heartbeat` timestamps
  - Handle duplicate entries (user already in lobby) - return conflict error
  - **Acceptance Criteria:**
    - Successfully inserts presence for new user
    - Returns error if user already has active presence
    - Sets `joined_at` and `last_heartbeat` to current timestamp

- **LeaveLobby(ctx, userID)**
  - Delete presence record for user
  - **Acceptance Criteria:**
    - Successfully removes user from lobby
    - Returns no error if user not in lobby (idempotent)

- **UpdateHeartbeat(ctx, userID)**
  - Update `last_heartbeat` timestamp to NOW()
  - **Acceptance Criteria:**
    - Updates timestamp for existing presence
    - Returns error if user not in lobby

- **GetLobbyUsers(ctx)**
  - Query all users with active presence
  - Join with USER table to get username, email
  - Return list with user_id, username, joined_at, last_heartbeat
  - **Acceptance Criteria:**
    - Returns all users in LOBBY_PRESENCE
    - Includes user details from USER table
    - Ordered by joined_at (newest first)

- **CleanupStaleLobbyPresence(ctx, timeoutDuration)**
  - Delete presence records where `last_heartbeat < NOW() - timeout`
  - Default timeout: 60 seconds
  - **Acceptance Criteria:**
    - Removes users who haven't sent heartbeat in >60s
    - Returns number of rows deleted
    - Does not affect active users

### 1.2 Game Invitation Repository (`repository/invitation.go`)

#### Requirements:
- **CreateInvitation(ctx, challengerID, challengedID)**
  - Insert GAME_INVITATION with status='pending'
  - Validate challenger != challenged (business logic)
  - Check for existing pending invitation between same users
  - **Acceptance Criteria:**
    - Creates invitation with status='pending'
    - Returns invitation_id
    - Prevents duplicate pending invitations
    - Sets created_at timestamp

- **GetInvitationsByUser(ctx, userID)**
  - Query invitations where user is challenger OR challenged
  - Separate into sent vs received
  - Join with USER table for opponent details
  - Filter status='pending' only
  - **Acceptance Criteria:**
    - Returns two lists: sent and received
    - Includes opponent user details
    - Only returns pending invitations
    - Ordered by created_at DESC

- **GetInvitationByID(ctx, invitationID)**
  - Retrieve single invitation with full details
  - **Acceptance Criteria:**
    - Returns invitation if exists
    - Returns error if not found
    - Includes challenger and challenged user details

- **AcceptInvitation(ctx, invitationID, gameID)**
  - Begin transaction
  - Update invitation: status='accepted', game_id=gameID
  - Verify invitation is still pending
  - Commit transaction
  - **Acceptance Criteria:**
    - Updates status to 'accepted'
    - Links invitation to game_id
    - Returns error if already processed
    - Atomic operation (transaction)

- **DeclineInvitation(ctx, invitationID)**
  - Update invitation: status='declined'
  - Verify invitation is pending
  - **Acceptance Criteria:**
    - Updates status to 'declined'
    - Returns error if already processed
    - Does not create game

- **CancelInvitation(ctx, invitationID)**
  - Delete invitation (or mark as cancelled)
  - Only challenger can cancel
  - **Acceptance Criteria:**
    - Removes/cancels invitation
    - Only works for pending invitations
    - Returns error if not challenger

- **CleanupExpiredInvitations(ctx, expirationTime)**
  - Update invitations to 'expired' if created_at > expirationTime
  - Default: 5 minutes
  - **Acceptance Criteria:**
    - Marks old pending invitations as expired
    - Does not affect accepted/declined invitations
    - Returns count of expired invitations

---

## Phase 2: Service Layer (HTTP Handlers)

### 2.1 Lobby Handlers (`service/lobby.go`)

#### **GET /api/v1/lobby/users** - `LobbyUsersHandler`
**Requirements:**
- Must be authenticated (session middleware)
- Call repository.GetLobbyUsers()
- Return JSON list of online users

**Response Format:**
```json
{
  "users": [
    {
      "userId": 42,
      "username": "player123",
      "joinedAt": "2025-11-01T12:00:00Z",
      "lastHeartbeat": "2025-11-01T12:05:00Z"
    }
  ],
  "count": 1
}
```

**Acceptance Criteria:**
- ✅ Returns 200 OK with user list
- ✅ Returns 401 if not authenticated
- ✅ Excludes current user's own presence (optional - decide)
- ✅ Returns empty array if no users

---

#### **POST /api/v1/lobby/presence** - `LobbyPresenceHandler`
**Requirements:**
- Must be authenticated
- Get userID from session context
- Call repository.JoinLobby(userID)
- Handle duplicate entry (409 Conflict)

**Response Format:**
```json
{
  "message": "Joined lobby successfully",
  "presenceId": 156,
  "joinedAt": "2025-11-01T12:00:00Z"
}
```

**Acceptance Criteria:**
- ✅ Returns 201 Created on success
- ✅ Returns 409 Conflict if already in lobby
- ✅ Returns 401 if not authenticated

---

#### **DELETE /api/v1/lobby/presence** - `LobbyPresenceHandler` (same handler, check method)
**Requirements:**
- Must be authenticated
- Get userID from session context
- Call repository.LeaveLobby(userID)
- Return success even if not in lobby (idempotent)

**Response Format:**
```json
{
  "message": "Left lobby successfully"
}
```

**Acceptance Criteria:**
- ✅ Returns 200 OK on success
- ✅ Returns 401 if not authenticated
- ✅ Idempotent - no error if already left

---

#### **PUT /api/v1/lobby/presence/heartbeat** - `LobbyPresenceHeartbeatHandler`
**Requirements:**
- Must be authenticated
- Get userID from session context
- Call repository.UpdateHeartbeat(userID)
- Return 404 if user not in lobby

**Response Format:**
```json
{
  "message": "Heartbeat updated",
  "lastHeartbeat": "2025-11-01T12:05:00Z"
}
```

**Acceptance Criteria:**
- ✅ Returns 200 OK on success
- ✅ Returns 404 if user not in lobby
- ✅ Returns 401 if not authenticated
- ✅ Updates last_heartbeat timestamp

---

### 2.2 Invitation Handlers (`service/invitation.go`)

#### **POST /api/v1/invitations** - `InvitationsHandler` (Create)
**Requirements:**
- Must be authenticated
- Parse JSON body: `{"challengedId": 73}`
- Validate challengerID != challengedId (can't invite yourself)
- Verify challenged user exists
- Verify challenged user is in lobby
- Call repository.CreateInvitation(challengerID, challengedID)
- Handle duplicate pending invitation (409 Conflict)

**Request Body:**
```json
{
  "challengedId": 73
}
```

**Response Format:**
```json
{
  "invitationId": 42,
  "challengedId": 73,
  "status": "pending",
  "message": "Invitation sent successfully"
}
```

**Acceptance Criteria:**
- ✅ Returns 201 Created on success
- ✅ Returns 400 if challenging self
- ✅ Returns 400 if challengedId missing
- ✅ Returns 404 if challenged user not found
- ✅ Returns 404 if challenged user not in lobby
- ✅ Returns 409 if pending invitation already exists
- ✅ Returns 401 if not authenticated

---

#### **GET /api/v1/invitations** - `InvitationsHandler` (List)
**Requirements:**
- Must be authenticated
- Get userID from session context
- Call repository.GetInvitationsByUser(userID)
- Return separate lists: sent and received

**Response Format:**
```json
{
  "sent": [
    {
      "invitationId": 42,
      "challenger": {
        "userId": 10,
        "username": "player1"
      },
      "challenged": {
        "userId": 73,
        "username": "player2"
      },
      "status": "pending",
      "createdAt": "2025-11-01T12:00:00Z"
    }
  ],
  "received": [
    {
      "invitationId": 43,
      "challenger": {
        "userId": 20,
        "username": "player3"
      },
      "challenged": {
        "userId": 10,
        "username": "player1"
      },
      "status": "pending",
      "createdAt": "2025-11-01T12:01:00Z"
    }
  ]
}
```

**Acceptance Criteria:**
- ✅ Returns 200 OK with invitation lists
- ✅ Separates sent vs received invitations
- ✅ Only returns pending invitations
- ✅ Includes opponent user details
- ✅ Returns 401 if not authenticated

---

#### **PUT /api/v1/invitations/{id}/accept** - `AcceptInvitationHandler`
**Requirements:**
- Must be authenticated
- Parse invitation ID from URL path
- Get userID from session context
- Call repository.GetInvitationByID(id)
- Verify user is the challenged party (not challenger)
- Verify invitation is pending
- **Create new game:**
  - Insert into GAME table (status='pending')
  - Assign random colors to players
  - Set current_turn to a player (random)
- Call repository.AcceptInvitation(id, gameID)
- Return game ID

**Response Format:**
```json
{
  "message": "Invitation accepted",
  "gameId": 123
}
```

**Acceptance Criteria:**
- ✅ Returns 200 OK with game ID
- ✅ Returns 404 if invitation not found
- ✅ Returns 400 if user is not challenged party
- ✅ Returns 400 if invitation already processed
- ✅ Creates game in GAME table
- ✅ Updates invitation status to 'accepted'
- ✅ Links invitation to game_id
- ✅ Returns 401 if not authenticated
- ✅ Uses transaction for game creation + invitation update

---

#### **PUT /api/v1/invitations/{id}/decline** - `DeclineInvitationHandler`
**Requirements:**
- Must be authenticated
- Parse invitation ID from URL path
- Get userID from session context
- Call repository.GetInvitationByID(id)
- Verify user is the challenged party
- Verify invitation is pending
- Call repository.DeclineInvitation(id)

**Response Format:**
```json
{
  "message": "Invitation declined"
}
```

**Acceptance Criteria:**
- ✅ Returns 200 OK
- ✅ Returns 404 if invitation not found
- ✅ Returns 400 if user is not challenged party
- ✅ Returns 400 if invitation already processed
- ✅ Updates invitation status to 'declined'
- ✅ Returns 401 if not authenticated

---

#### **DELETE /api/v1/invitations/{id}** - `CancelInvitationHandler`
**Requirements:**
- Must be authenticated
- Parse invitation ID from URL path
- Get userID from session context
- Call repository.GetInvitationByID(id)
- Verify user is the challenger (not challenged)
- Verify invitation is pending
- Call repository.CancelInvitation(id)

**Response Format:**
```json
{
  "message": "Invitation cancelled"
}
```

**Acceptance Criteria:**
- ✅ Returns 200 OK
- ✅ Returns 404 if invitation not found
- ✅ Returns 400 if user is not challenger
- ✅ Returns 400 if invitation already processed
- ✅ Deletes invitation record
- ✅ Returns 401 if not authenticated

---

## Phase 3: Frontend Implementation

### 3.1 Update LobbyPage (`client/src/pages/LobbyPage.tsx`)

#### Requirements:

**On Component Mount:**
- Call `/api/v1/lobby/presence` (POST) to join lobby
- Start heartbeat interval (every 30 seconds)
- Start polling for online users (every 5 seconds)
- Start polling for invitations (every 5 seconds)

**On Component Unmount:**
- Call `/api/v1/lobby/presence` (DELETE) to leave lobby
- Clear all intervals

**Display:**
- List of online users (excluding self)
- Each user has "Challenge" button
- List of received invitations with Accept/Decline buttons
- List of sent invitations with Cancel button

**Acceptance Criteria:**
- ✅ Auto-joins lobby on mount
- ✅ Leaves lobby on unmount
- ✅ Sends heartbeat every 30s while on page
- ✅ Polls for users every 5s
- ✅ Polls for invitations every 5s
- ✅ Displays online users with challenge button
- ✅ Shows received invitations
- ✅ Shows sent invitations
- ✅ Handle "Challenge" click → sends invitation
- ✅ Handle "Accept" click → accepts invitation, navigates to game
- ✅ Handle "Decline" click → declines invitation
- ✅ Handle "Cancel" click → cancels sent invitation
- ✅ Shows loading states
- ✅ Shows error messages

---

### 3.2 Create API Client Functions (`client/src/api/lobby.ts`)

#### Requirements:
```typescript
// Lobby presence
export async function joinLobby(): Promise<void>
export async function leaveLobby(): Promise<void>
export async function sendHeartbeat(): Promise<void>
export async function getLobbyUsers(): Promise<LobbyUser[]>

// Invitations
export async function sendInvitation(challengedId: number): Promise<Invitation>
export async function getInvitations(): Promise<{sent: Invitation[], received: Invitation[]}>
export async function acceptInvitation(invitationId: number): Promise<{gameId: number}>
export async function declineInvitation(invitationId: number): Promise<void>
export async function cancelInvitation(invitationId: number): Promise<void>
```

**Acceptance Criteria:**
- ✅ All functions use `fetch` with `credentials: 'include'`
- ✅ Proper error handling (throw on non-ok response)
- ✅ TypeScript types for all responses
- ✅ JSON parsing for responses

---

### 3.3 Create Types (`client/src/types/lobby.ts`)

#### Requirements:
```typescript
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
```

**Acceptance Criteria:**
- ✅ Matches API response structure
- ✅ Exported from types file

---

### 3.4 UI Components

#### **OnlineUsersList Component**
- Display list of online users
- Each user has username and "Challenge" button
- Exclude current user from list
- Show count of online users

#### **InvitationsPanel Component**
- Two sections: Received and Sent
- **Received:** Show Accept/Decline buttons
- **Sent:** Show Cancel button + "Waiting for response..." status
- Show timestamp for each invitation
- Auto-remove from list when processed

**Acceptance Criteria:**
- ✅ Clean, shadcn-styled UI
- ✅ Buttons have loading states
- ✅ Error messages display properly
- ✅ Real-time updates via polling

---

## Phase 4: Background Jobs & Cleanup

### 4.1 Stale Presence Cleanup

#### Requirements:
- Run periodic job (every 30 seconds) in Go
- Call `repository.CleanupStaleLobbyPresence(ctx, 60*time.Second)`
- Log number of users removed

**Implementation:**
```go
// In main.go
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    for range ticker.C {
        db := repository.GetDB()
        db.CleanupStaleLobbyPresence(context.Background(), 60*time.Second)
    }
}()
```

**Acceptance Criteria:**
- ✅ Runs automatically in background
- ✅ Removes users with last_heartbeat > 60s old
- ✅ Does not interfere with active users

---

### 4.2 Expired Invitation Cleanup

#### Requirements:
- Run periodic job (every 60 seconds) in Go
- Call `repository.CleanupExpiredInvitations(ctx, 5*time.Minute)`
- Mark old pending invitations as expired

**Acceptance Criteria:**
- ✅ Runs automatically in background
- ✅ Expires invitations older than 5 minutes
- ✅ Only affects pending invitations

---

## Phase 5: Game Creation (Part of Accept Invitation)

### 5.1 Create Game Repository Function

#### **CreateGame(ctx, player1ID, player2ID)** in `repository/game.go`
**Requirements:**
- Insert into GAME table
- Randomly assign colors (white/black)
- Randomly select starting player (current_turn)
- Set status='pending' or 'in_progress'
- Return game_id

**Acceptance Criteria:**
- ✅ Creates game record
- ✅ Random color assignment
- ✅ Random turn assignment
- ✅ Returns game ID
- ✅ Enforces player1_id != player2_id

---

## Testing Checklist

### User Tests:
- [ ] Full flow: User A challenges User B
- [ ] User B accepts → Game created
- [ ] User A and B removed from lobby
- [ ] Stale users cleaned up automatically
- [ ] Old invitations expire

---

## Success Criteria

**The lobby system is complete when:**
1. ✅ Users can see other online players
2. ✅ Users can challenge other players
3. ✅ Users can accept/decline challenges
4. ✅ Accepting creates a game
5. ✅ Stale users are automatically removed
6. ✅ Old invitations expire automatically
7. ✅ All API endpoints work as specified
8. ✅ Frontend updates in real-time via polling
9. ✅ Error handling is robust
10. ✅ UI is clean and intuitive

---

## Implementation Order

1. Repository layer (lobby + invitations)
2. Service handlers (lobby endpoints)
3. Service handlers (invitation endpoints)
4. Game creation logic
5. Background cleanup jobs
6. Frontend API client functions
7. Frontend types
8. Update LobbyPage UI
9. Component styling (shadcn)
10. Testing & bug fixes
11. Polish & error handling

---

## Notes

- **Invitation Expiration:** 5 minutes default, configurable.
- **Heartbeat Timeout:** 60 seconds default, configurable.
- **Lobby Cleanup:** Runs every 30 seconds, can be adjusted.
- **Colors:** Random assignment on game creation. Can add preference later.
- **Turn Order:** Random selection of starting player. Can add dice roll later.

---

## API Endpoints Summary

| Method | Endpoint | Handler | Auth Required |
|--------|----------|---------|---------------|
| GET | /api/v1/lobby/users | LobbyUsersHandler | ✅ |
| POST | /api/v1/lobby/presence | LobbyPresenceHandler | ✅ |
| DELETE | /api/v1/lobby/presence | LobbyPresenceHandler | ✅ |
| PUT | /api/v1/lobby/presence/heartbeat | LobbyPresenceHeartbeatHandler | ✅ |
| POST | /api/v1/invitations | InvitationsHandler | ✅ |
| GET | /api/v1/invitations | InvitationsHandler | ✅ |
| PUT | /api/v1/invitations/{id}/accept | AcceptInvitationHandler | ✅ |
| PUT | /api/v1/invitations/{id}/decline | DeclineInvitationHandler | ✅ |
| DELETE | /api/v1/invitations/{id} | CancelInvitationHandler | ✅ |

---

## Database Schema Reference

**LOBBY_PRESENCE:**
- presence_id (PK)
- user_id (FK → USER, UNIQUE)
- joined_at
- last_heartbeat

**GAME_INVITATION:**
- invitation_id (PK)
- challenger_id (FK → USER)
- challenged_id (FK → USER)
- status (pending/accepted/declined/expired)
- game_id (FK → GAME, nullable)
- created_at

**GAME:**
- game_id (PK)
- player1_id (FK → USER)
- player2_id (FK → USER)
- current_turn (FK → USER)
- game_status (pending/in_progress/completed/abandoned)
- winner_id (FK → USER, nullable)
- player1_color (white/black)
- player2_color (white/black)
- created_at, started_at, ended_at
