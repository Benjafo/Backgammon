-- ============================================================================
-- Backgammon Schema
-- ============================================================================
-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS CHAT_MESSAGE;

DROP TABLE IF EXISTS CHAT_ROOM;

DROP TABLE IF EXISTS MOVE;

DROP TABLE IF EXISTS GAME_STATE;

DROP TABLE IF EXISTS GAME_INVITATION;

DROP TABLE IF EXISTS LOBBY_PRESENCE;

DROP TABLE IF EXISTS GAME;

DROP TABLE IF EXISTS SESSIONS;

DROP TABLE IF EXISTS REGISTRATION_TOKEN;

DROP TABLE IF EXISTS USER;

-- ============================================================================
-- USER table
-- Store player account information and credentials
-- ============================================================================
CREATE TABLE
    USER (
        user_id INT AUTO_INCREMENT PRIMARY KEY,
        username VARCHAR(50) NOT NULL UNIQUE,
        password_hash VARCHAR(255) NOT NULL,
        email VARCHAR(100) UNIQUE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        last_login TIMESTAMP NULL,
        -- Constraints
        CONSTRAINT chk_username_length CHECK (CHAR_LENGTH(username) >= 3),
        CONSTRAINT chk_email_format CHECK (
            email IS NULL
            OR email LIKE '%_@_%._%'
        )
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ============================================================================
-- REGISTRATION_TOKEN table
-- Validate user registration to prevent CSRF attacks
-- ============================================================================
CREATE TABLE
    REGISTRATION_TOKEN (
        token_id INT AUTO_INCREMENT PRIMARY KEY,
        token_value VARCHAR(64) NOT NULL UNIQUE,
        ip_address VARCHAR(45) NOT NULL,
        user_agent VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        expires_at TIMESTAMP NOT NULL,
        is_used BOOLEAN DEFAULT FALSE,
        -- Indexes
        INDEX idx_expires_at (expires_at),
        INDEX idx_is_used (is_used)
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ============================================================================
-- SESSION table
-- Manage user authentication tokens and active user sessions
-- ============================================================================
CREATE TABLE
    SESSIONS (
        session_id INT AUTO_INCREMENT PRIMARY KEY,
        user_id INT NOT NULL,
        session_token VARCHAR(64) NOT NULL UNIQUE,
        ip_address VARCHAR(45) NOT NULL,
        user_agent VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        expires_at TIMESTAMP NOT NULL,
        is_active BOOLEAN DEFAULT TRUE,
        -- Foreign key
        CONSTRAINT fk_session_user FOREIGN KEY (user_id) REFERENCES USER (user_id) ON DELETE CASCADE,
        -- Indexes
        INDEX idx_user_id (user_id),
        INDEX idx_expires_at (expires_at),
        INDEX idx_is_active (is_active)
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ============================================================================
-- GAME table
-- Represent backgammon matches between two players
-- ============================================================================
CREATE TABLE
    GAME (
        game_id INT AUTO_INCREMENT PRIMARY KEY,
        player1_id INT NOT NULL,
        player2_id INT NOT NULL,
        current_turn INT NOT NULL,
        game_status ENUM (
            'pending',
            'in_progress',
            'completed',
            'abandoned'
        ) NOT NULL DEFAULT 'pending',
        winner_id INT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        started_at TIMESTAMP NULL,
        ended_at TIMESTAMP NULL,
        player1_color ENUM ('white', 'black') NOT NULL,
        player2_color ENUM ('white', 'black') NOT NULL,
        -- Foreign keys
        CONSTRAINT fk_game_player1 FOREIGN KEY (player1_id) REFERENCES USER (user_id) ON DELETE CASCADE,
        CONSTRAINT fk_game_player2 FOREIGN KEY (player2_id) REFERENCES USER (user_id) ON DELETE CASCADE,
        CONSTRAINT fk_game_current_turn FOREIGN KEY (current_turn) REFERENCES USER (user_id) ON DELETE CASCADE,
        CONSTRAINT fk_game_winner FOREIGN KEY (winner_id) REFERENCES USER (user_id) ON DELETE SET NULL,
        -- Constraints
        CONSTRAINT chk_different_players CHECK (player1_id != player2_id),
        CONSTRAINT chk_different_colors CHECK (player1_color != player2_color),
        CONSTRAINT chk_valid_turn CHECK (current_turn IN (player1_id, player2_id)),
        -- Indexes
        INDEX idx_player1_id (player1_id),
        INDEX idx_player2_id (player2_id),
        INDEX idx_game_status (game_status),
        INDEX idx_created_at (created_at)
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ============================================================================
-- GAME_STATE table
-- Store current board configuration and game state
-- ============================================================================
CREATE TABLE
    GAME_STATE (
        state_id INT AUTO_INCREMENT PRIMARY KEY,
        game_id INT NOT NULL UNIQUE,
        board_state JSON NOT NULL,
        bar_white INT NOT NULL DEFAULT 0,
        bar_black INT NOT NULL DEFAULT 0,
        borne_off_white INT NOT NULL DEFAULT 0,
        borne_off_black INT NOT NULL DEFAULT 0,
        dice_roll JSON NULL,
        dice_used JSON NULL,
        last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
        ),
        -- Indexes
        INDEX idx_last_updated (last_updated)
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ============================================================================
-- MOVE table
-- Store moves history
-- ============================================================================
CREATE TABLE
    MOVE (
        move_id INT AUTO_INCREMENT PRIMARY KEY,
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
        CONSTRAINT fk_move_player FOREIGN KEY (player_id) REFERENCES USER (user_id) ON DELETE CASCADE,
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
        CONSTRAINT chk_move_number_positive CHECK (move_number > 0),
        -- Indexes
        INDEX idx_game_move (game_id, move_number),
        INDEX idx_player_id (player_id),
        INDEX idx_timestamp (timestamp)
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ============================================================================
-- CHAT_ROOM table
-- Separate chat contexts for lobby and individual game rooms
-- ============================================================================
CREATE TABLE
    CHAT_ROOM (
        room_id INT AUTO_INCREMENT PRIMARY KEY,
        room_type ENUM ('lobby', 'game') NOT NULL,
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
        ),
        -- Indexes
        INDEX idx_room_type (room_type),
        INDEX idx_game_id (game_id)
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ============================================================================
-- CHAT_MESSAGE table
-- Store user chat messages
-- ============================================================================
CREATE TABLE
    CHAT_MESSAGE (
        message_id INT AUTO_INCREMENT PRIMARY KEY,
        room_id INT NOT NULL,
        user_id INT NOT NULL,
        message_text TEXT NOT NULL,
        timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        -- Foreign keys
        CONSTRAINT fk_chatmessage_room FOREIGN KEY (room_id) REFERENCES CHAT_ROOM (room_id) ON DELETE CASCADE,
        CONSTRAINT fk_chatmessage_user FOREIGN KEY (user_id) REFERENCES USER (user_id) ON DELETE CASCADE,
        -- Constraints
        CONSTRAINT chk_message_length CHECK (
            CHAR_LENGTH(message_text) > 0
            AND CHAR_LENGTH(message_text) <= 1000
        ),
        -- Indexes
        INDEX idx_room_timestamp (room_id, timestamp),
        INDEX idx_user_id (user_id)
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ============================================================================
-- LOBBY_PRESENCE table
-- Track which users are currently active in the lobby
-- ============================================================================
CREATE TABLE
    LOBBY_PRESENCE (
        presence_id INT AUTO_INCREMENT PRIMARY KEY,
        user_id INT NOT NULL UNIQUE,
        joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        last_heartbeat TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        -- Foreign key
        CONSTRAINT fk_lobbypresence_user FOREIGN KEY (user_id) REFERENCES USER (user_id) ON DELETE CASCADE,
        -- Index
        INDEX idx_last_heartbeat (last_heartbeat)
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ============================================================================
-- GAME_INVITATION table
-- Manage game requests between players in the lobby
-- ============================================================================
CREATE TABLE
    GAME_INVITATION (
        invitation_id INT AUTO_INCREMENT PRIMARY KEY,
        challenger_id INT NOT NULL,
        challenged_id INT NOT NULL,
        status ENUM ('pending', 'accepted', 'declined', 'expired') NOT NULL DEFAULT 'pending',
        game_id INT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        -- Foreign keys
        CONSTRAINT fk_invitation_challenger FOREIGN KEY (challenger_id) REFERENCES USER (user_id) ON DELETE CASCADE,
        CONSTRAINT fk_invitation_challenged FOREIGN KEY (challenged_id) REFERENCES USER (user_id) ON DELETE CASCADE,
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
        ),
        -- Indexes
        INDEX idx_challenged_status (challenged_id, status),
        INDEX idx_challenger_id (challenger_id)
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- ============================================================================
-- Create the lobby chat room
-- ============================================================================
INSERT INTO
    CHAT_ROOM (room_type, game_id)
VALUES
    ('lobby', NULL);