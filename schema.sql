-- ============================================================================
-- Backgammon Schema - PostgreSQL
-- ============================================================================
-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS CHAT_MESSAGE CASCADE;
DROP TABLE IF EXISTS CHAT_ROOM CASCADE;
DROP TABLE IF EXISTS MOVE CASCADE;
DROP TABLE IF EXISTS GAME_STATE CASCADE;
DROP TABLE IF EXISTS GAME_INVITATION CASCADE;
DROP TABLE IF EXISTS LOBBY_PRESENCE CASCADE;
DROP TABLE IF EXISTS GAME CASCADE;
DROP TABLE IF EXISTS SESSIONS CASCADE;
DROP TABLE IF EXISTS REGISTRATION_TOKEN CASCADE;
DROP TABLE IF EXISTS "USER" CASCADE;

-- ============================================================================
-- USER table
-- Store player account information and credentials
-- ============================================================================
CREATE TABLE "USER" (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    email VARCHAR(255) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP NULL,
    -- Constraints
    CONSTRAINT chk_username_length CHECK (LENGTH(username) >= 3),
    CONSTRAINT chk_email_format CHECK (
        email IS NULL
        OR email LIKE '%_@_%._%'
    )
);

-- ============================================================================
-- REGISTRATION_TOKEN table
-- Validate user registration to prevent CSRF attacks
-- ============================================================================
CREATE TABLE REGISTRATION_TOKEN (
    token_id SERIAL PRIMARY KEY,
    token_value VARCHAR(500) NOT NULL UNIQUE,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_used BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_registration_token_expires_at ON REGISTRATION_TOKEN(expires_at);
CREATE INDEX idx_registration_token_is_used ON REGISTRATION_TOKEN(is_used);

-- ============================================================================
-- SESSION table
-- Manage user authentication tokens and active user sessions
-- ============================================================================
CREATE TABLE SESSIONS (
    session_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    session_token VARCHAR(500) NOT NULL UNIQUE,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    -- Foreign key
    CONSTRAINT fk_session_user FOREIGN KEY (user_id) REFERENCES "USER" (user_id) ON DELETE CASCADE
);

CREATE INDEX idx_session_user_id ON SESSIONS(user_id);
CREATE INDEX idx_session_expires_at ON SESSIONS(expires_at);
CREATE INDEX idx_session_is_active ON SESSIONS(is_active);

-- ============================================================================
-- GAME table
-- Represent backgammon matches between two players
-- ============================================================================
CREATE TYPE game_status_enum AS ENUM ('pending', 'in_progress', 'completed', 'abandoned');
CREATE TYPE color_enum AS ENUM ('white', 'black');

CREATE TABLE GAME (
    game_id SERIAL PRIMARY KEY,
    player1_id INT NOT NULL,
    player2_id INT NOT NULL,
    current_turn INT NOT NULL,
    game_status game_status_enum NOT NULL DEFAULT 'pending',
    winner_id INT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP NULL,
    ended_at TIMESTAMP NULL,
    player1_color color_enum NOT NULL,
    player2_color color_enum NOT NULL,
    -- Foreign keys
    CONSTRAINT fk_game_player1 FOREIGN KEY (player1_id) REFERENCES "USER" (user_id) ON DELETE CASCADE,
    CONSTRAINT fk_game_player2 FOREIGN KEY (player2_id) REFERENCES "USER" (user_id) ON DELETE CASCADE,
    CONSTRAINT fk_game_current_turn FOREIGN KEY (current_turn) REFERENCES "USER" (user_id) ON DELETE CASCADE,
    CONSTRAINT fk_game_winner FOREIGN KEY (winner_id) REFERENCES "USER" (user_id) ON DELETE SET NULL,
    -- Constraints
    CONSTRAINT chk_different_players CHECK (player1_id != player2_id),
    CONSTRAINT chk_different_colors CHECK (player1_color != player2_color),
    CONSTRAINT chk_valid_turn CHECK (current_turn IN (player1_id, player2_id))
);

CREATE INDEX idx_game_player1_id ON GAME(player1_id);
CREATE INDEX idx_game_player2_id ON GAME(player2_id);
CREATE INDEX idx_game_status ON GAME(game_status);
CREATE INDEX idx_game_created_at ON GAME(created_at);

-- ============================================================================
-- GAME_STATE table
-- Store current board configuration and game state
-- ============================================================================
CREATE TABLE GAME_STATE (
    state_id SERIAL PRIMARY KEY,
    game_id INT NOT NULL UNIQUE,
    board_state JSONB NOT NULL,
    bar_white INT NOT NULL DEFAULT 0,
    bar_black INT NOT NULL DEFAULT 0,
    borne_off_white INT NOT NULL DEFAULT 0,
    borne_off_black INT NOT NULL DEFAULT 0,
    dice_roll JSONB NULL,
    dice_used JSONB NULL,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- Foreign keys
    CONSTRAINT fk_gamestate_game FOREIGN KEY (game_id) REFERENCES GAME (game_id) ON DELETE CASCADE,
    -- Constraints
    CONSTRAINT chk_bar_white_range CHECK (
        bar_white >= 0
        AND bar_white <= 15
    ),
    CONSTRAINT chk_bar_black_range CHECK (
        bar_black >= 0
        AND bar_black <= 15
    ),
    CONSTRAINT chk_borne_white_range CHECK (
        borne_off_white >= 0
        AND borne_off_white <= 15
    ),
    CONSTRAINT chk_borne_black_range CHECK (
        borne_off_black >= 0
        AND borne_off_black <= 15
    )
);

CREATE INDEX idx_gamestate_last_updated ON GAME_STATE(last_updated);

-- ============================================================================
-- MOVE table
-- Store moves history
-- ============================================================================
CREATE TABLE MOVE (
    move_id SERIAL PRIMARY KEY,
    game_id INT NOT NULL,
    player_id INT NOT NULL,
    move_number INT NOT NULL,
    from_point INT NOT NULL,
    to_point INT NOT NULL,
    die_used INT NOT NULL,
    hit_opponent BOOLEAN DEFAULT FALSE,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- Foreign keys
    CONSTRAINT fk_move_game FOREIGN KEY (game_id) REFERENCES GAME (game_id) ON DELETE CASCADE,
    CONSTRAINT fk_move_player FOREIGN KEY (player_id) REFERENCES "USER" (user_id) ON DELETE CASCADE,
    -- Constraints
    CONSTRAINT chk_from_point_range CHECK (
        from_point >= 0
        AND from_point <= 25
    ),
    CONSTRAINT chk_to_point_range CHECK (
        to_point >= 0
        AND to_point <= 25
    ),
    CONSTRAINT chk_die_value CHECK (
        die_used >= 1
        AND die_used <= 6
    ),
    CONSTRAINT chk_move_number_positive CHECK (move_number > 0)
);

CREATE INDEX idx_move_game_move ON MOVE(game_id, move_number);
CREATE INDEX idx_move_player_id ON MOVE(player_id);
CREATE INDEX idx_move_timestamp ON MOVE(timestamp);

-- ============================================================================
-- CHAT_ROOM table
-- Separate chat contexts for lobby and individual game rooms
-- ============================================================================
CREATE TYPE room_type_enum AS ENUM ('lobby', 'game');

CREATE TABLE CHAT_ROOM (
    room_id SERIAL PRIMARY KEY,
    room_type room_type_enum NOT NULL,
    game_id INT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- Foreign key
    CONSTRAINT fk_chatroom_game FOREIGN KEY (game_id) REFERENCES GAME (game_id) ON DELETE CASCADE,
    -- Constraints
    CONSTRAINT chk_lobby_no_game CHECK (
        (
            room_type = 'lobby'
            AND game_id IS NULL
        )
        OR (
            room_type = 'game'
            AND game_id IS NOT NULL
        )
    )
);

CREATE INDEX idx_chatroom_room_type ON CHAT_ROOM(room_type);
CREATE INDEX idx_chatroom_game_id ON CHAT_ROOM(game_id);

-- ============================================================================
-- CHAT_MESSAGE table
-- Store user chat messages
-- ============================================================================
CREATE TABLE CHAT_MESSAGE (
    message_id SERIAL PRIMARY KEY,
    room_id INT NOT NULL,
    user_id INT NOT NULL,
    message_text TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- Foreign keys
    CONSTRAINT fk_chatmessage_room FOREIGN KEY (room_id) REFERENCES CHAT_ROOM (room_id) ON DELETE CASCADE,
    CONSTRAINT fk_chatmessage_user FOREIGN KEY (user_id) REFERENCES "USER" (user_id) ON DELETE CASCADE,
    -- Constraints
    CONSTRAINT chk_message_length CHECK (
        LENGTH(message_text) > 0
        AND LENGTH(message_text) <= 1000
    )
);

CREATE INDEX idx_chatmessage_room_timestamp ON CHAT_MESSAGE(room_id, timestamp);
CREATE INDEX idx_chatmessage_user_id ON CHAT_MESSAGE(user_id);

-- ============================================================================
-- LOBBY_PRESENCE table
-- Track which users are currently active in the lobby
-- ============================================================================
CREATE TABLE LOBBY_PRESENCE (
    presence_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_heartbeat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- Foreign key
    CONSTRAINT fk_lobbypresence_user FOREIGN KEY (user_id) REFERENCES "USER" (user_id) ON DELETE CASCADE
);

CREATE INDEX idx_lobbypresence_last_heartbeat ON LOBBY_PRESENCE(last_heartbeat);

-- ============================================================================
-- GAME_INVITATION table
-- Manage game requests between players in the lobby
-- ============================================================================
CREATE TYPE invitation_status_enum AS ENUM ('pending', 'accepted', 'declined', 'expired');

CREATE TABLE GAME_INVITATION (
    invitation_id SERIAL PRIMARY KEY,
    challenger_id INT NOT NULL,
    challenged_id INT NOT NULL,
    status invitation_status_enum NOT NULL DEFAULT 'pending',
    game_id INT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- Foreign keys
    CONSTRAINT fk_invitation_challenger FOREIGN KEY (challenger_id) REFERENCES "USER" (user_id) ON DELETE CASCADE,
    CONSTRAINT fk_invitation_challenged FOREIGN KEY (challenged_id) REFERENCES "USER" (user_id) ON DELETE CASCADE,
    CONSTRAINT fk_invitation_game FOREIGN KEY (game_id) REFERENCES GAME (game_id) ON DELETE SET NULL,
    -- Constraints
    CONSTRAINT chk_different_users CHECK (challenger_id != challenged_id),
    CONSTRAINT chk_game_only_when_accepted CHECK (
        (
            status = 'accepted'
            AND game_id IS NOT NULL
        )
        OR (
            status != 'accepted'
            AND game_id IS NULL
        )
    )
);

CREATE INDEX idx_invitation_challenged_status ON GAME_INVITATION(challenged_id, status);
CREATE INDEX idx_invitation_challenger_id ON GAME_INVITATION(challenger_id);

-- ============================================================================
-- Create the lobby chat room
-- ============================================================================
INSERT INTO CHAT_ROOM (room_type, game_id) VALUES ('lobby', NULL);
