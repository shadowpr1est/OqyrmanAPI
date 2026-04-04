-- ─── Users: add auth columns if missing ──────────────────────────────────────
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS email_verified    BOOLEAN     NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS google_id         TEXT        UNIQUE;

-- ─── Tokens: add session metadata columns if missing ─────────────────────────
ALTER TABLE tokens
    ADD COLUMN IF NOT EXISTS user_agent TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS ip         VARCHAR(45)  NOT NULL DEFAULT '';

-- ─── Email verification codes ─────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS email_verification_codes (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code       TEXT        NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id)
);

-- Widen existing CHAR(6) column to TEXT (bcrypt hashes are ~60 chars)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'email_verification_codes'
          AND column_name = 'code'
          AND data_type = 'character'
    ) THEN
        ALTER TABLE email_verification_codes ALTER COLUMN code TYPE TEXT;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_email_verification_codes_user_id
    ON email_verification_codes(user_id);

-- ─── Password reset codes ─────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS password_reset_codes (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code       TEXT        NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id)
);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'password_reset_codes'
          AND column_name = 'code'
          AND data_type = 'character'
    ) THEN
        ALTER TABLE password_reset_codes ALTER COLUMN code TYPE TEXT;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_password_reset_codes_user_id
    ON password_reset_codes(user_id);

-- ─── Conversations ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS conversations (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      VARCHAR(255) NOT NULL DEFAULT 'Новый чат',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON conversations(user_id);

-- ─── Chat messages ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS chat_messages (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID        NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant')),
    content         TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_conv_id     ON chat_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at  ON chat_messages(conversation_id, created_at);
