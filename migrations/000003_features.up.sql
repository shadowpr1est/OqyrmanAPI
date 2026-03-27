-- total_pages for books
ALTER TABLE books ADD COLUMN total_pages INT NULL;

-- finished_at for reading_sessions
ALTER TABLE reading_sessions ADD COLUMN finished_at TIMESTAMPTZ NULL;

-- notifications table
CREATE TABLE notifications (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      VARCHAR(255) NOT NULL,
    body       TEXT         NOT NULL,
    is_read    BOOLEAN      NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    read_at    TIMESTAMPTZ  NULL
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
