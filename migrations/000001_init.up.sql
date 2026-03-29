CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ─── Users ────────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    phone         VARCHAR(20)  NOT NULL UNIQUE,
    password_hash TEXT         NOT NULL,
    name          VARCHAR(100) NOT NULL DEFAULT '',
    surname       VARCHAR(100) NOT NULL DEFAULT '',
    full_name     VARCHAR(255) NOT NULL DEFAULT '',
    avatar_url    TEXT         NOT NULL DEFAULT '',
    role          VARCHAR(20)  NOT NULL DEFAULT 'User',
    library_id    UUID,        -- NULL для admin/user, NOT NULL для staff (constraint ниже)
    qr_code       TEXT         NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

-- ─── Tokens ───────────────────────────────────────────────────────────────────
CREATE TABLE tokens (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token TEXT        NOT NULL UNIQUE,
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ─── Authors ──────────────────────────────────────────────────────────────────
CREATE TABLE authors (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    bio        TEXT         NOT NULL DEFAULT '',
    birth_date DATE,
    death_date DATE,
    photo_url  TEXT         NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- ─── Genres ───────────────────────────────────────────────────────────────────
CREATE TABLE genres (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) NOT NULL UNIQUE,
    slug       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- ─── Books ────────────────────────────────────────────────────────────────────
CREATE TABLE books (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id   UUID         NOT NULL REFERENCES authors(id) ON DELETE RESTRICT,
    genre_id    UUID         NOT NULL REFERENCES genres(id)  ON DELETE RESTRICT,
    title       VARCHAR(255) NOT NULL,
    isbn        VARCHAR(20)  NOT NULL DEFAULT '',
    cover_url   TEXT         NOT NULL DEFAULT '',
    description TEXT         NOT NULL DEFAULT '',
    language    VARCHAR(50)  NOT NULL DEFAULT '',
    year        INT          NOT NULL DEFAULT 0,
    avg_rating  FLOAT        NOT NULL DEFAULT 0,
    total_pages INT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

-- ─── Book files ───────────────────────────────────────────────────────────────
-- Constraint uq_book_file_type enforces max 1 audio + 1 document file per book.
-- Application-level check (magic bytes, size) happens in the usecase before insert.
CREATE TABLE book_files (
    id       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id  UUID        NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    format   VARCHAR(20) NOT NULL,
    file_url TEXT        NOT NULL,
    is_audio BOOLEAN     NOT NULL DEFAULT false,
    CONSTRAINT uq_book_file_type UNIQUE (book_id, is_audio)
);

-- ─── Libraries ────────────────────────────────────────────────────────────────
CREATE TABLE libraries (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    address    TEXT         NOT NULL,
    lat        FLOAT        NOT NULL,
    lng        FLOAT        NOT NULL,
    phone      VARCHAR(20)  NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- FK на library_id в users — добавляем после создания libraries
ALTER TABLE users
    ADD CONSTRAINT fk_users_library
        FOREIGN KEY (library_id) REFERENCES libraries(id) ON DELETE SET NULL;

-- staff обязан иметь library_id, остальные не должны
ALTER TABLE users
    ADD CONSTRAINT chk_staff_library
        CHECK (
            (role = 'Staff' AND library_id IS NOT NULL) OR
            (role != 'Staff' AND library_id IS NULL)
        );

-- ─── Library books ────────────────────────────────────────────────────────────
CREATE TABLE library_books (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    library_id       UUID NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
    book_id          UUID NOT NULL REFERENCES books(id)     ON DELETE CASCADE,
    total_copies     INT  NOT NULL DEFAULT 0,
    available_copies INT  NOT NULL DEFAULT 0,
    UNIQUE (library_id, book_id),
    CONSTRAINT chk_copies_non_negative
        CHECK (available_copies >= 0 AND total_copies >= 0),
    CONSTRAINT chk_available_lte_total
        CHECK (available_copies <= total_copies)
);

-- ─── Reservations ─────────────────────────────────────────────────────────────
CREATE TABLE reservations (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id)         ON DELETE CASCADE,
    library_book_id UUID        NOT NULL REFERENCES library_books(id) ON DELETE RESTRICT,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    reserved_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    due_date        TIMESTAMPTZ NOT NULL,
    returned_at     TIMESTAMPTZ,
    CONSTRAINT chk_reservation_status
        CHECK (status IN ('pending', 'active', 'completed', 'cancelled'))
);

-- ─── Reviews ──────────────────────────────────────────────────────────────────
CREATE TABLE reviews (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id    UUID        NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    rating     INT         NOT NULL CHECK (rating >= 1 AND rating <= 5),
    body       TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (user_id, book_id)
);

-- ─── Reading sessions ─────────────────────────────────────────────────────────
CREATE TABLE reading_sessions (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id      UUID        NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    current_page INT         NOT NULL DEFAULT 0,
    status       VARCHAR(20) NOT NULL DEFAULT 'reading',
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at  TIMESTAMPTZ,
    UNIQUE (user_id, book_id)
);

-- ─── Wishlists ────────────────────────────────────────────────────────────────
CREATE TABLE wishlists (
    id       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id  UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id  UUID        NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, book_id)
);

-- ─── Reading notes ────────────────────────────────────────────────────────────
CREATE TABLE reading_notes (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id    UUID        NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    page       INT         NOT NULL DEFAULT 0,
    content    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ─── Notifications ────────────────────────────────────────────────────────────
CREATE TABLE notifications (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      VARCHAR(255) NOT NULL,
    body       TEXT         NOT NULL,
    is_read    BOOLEAN      NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    read_at    TIMESTAMPTZ
);

-- ─── Events ───────────────────────────────────────────────────────────────────
CREATE TABLE events (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    cover_url   VARCHAR(500),
    location    VARCHAR(255),
    starts_at   TIMESTAMPTZ  NOT NULL,
    ends_at     TIMESTAMPTZ  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

-- ─── Indexes ──────────────────────────────────────────────────────────────────

-- users
CREATE INDEX idx_users_email      ON users(email)      WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role       ON users(role)        WHERE deleted_at IS NULL;
CREATE INDEX idx_users_library_id ON users(library_id)  WHERE deleted_at IS NULL;

-- books
CREATE INDEX idx_books_author_id  ON books(author_id)  WHERE deleted_at IS NULL;
CREATE INDEX idx_books_genre_id   ON books(genre_id)   WHERE deleted_at IS NULL;
CREATE INDEX idx_books_title      ON books(title)       WHERE deleted_at IS NULL;

-- libraries
CREATE INDEX idx_libraries_location ON libraries USING gist (point(lng, lat))
    WHERE deleted_at IS NULL;

-- library_books
CREATE INDEX idx_library_books_library_id ON library_books(library_id);
CREATE INDEX idx_library_books_book_id    ON library_books(book_id);

-- reservations
CREATE INDEX idx_reservations_user_id         ON reservations(user_id);
CREATE INDEX idx_reservations_library_book_id ON reservations(library_book_id);
CREATE INDEX idx_reservations_status          ON reservations(status);
CREATE INDEX idx_reservations_due_date        ON reservations(due_date)
    WHERE status IN ('pending', 'active');

-- reviews
CREATE INDEX idx_reviews_book_id ON reviews(book_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_reviews_user_id ON reviews(user_id) WHERE deleted_at IS NULL;

-- notifications
CREATE INDEX idx_notifications_user_id ON notifications(user_id);

-- events
CREATE INDEX idx_events_starts_at ON events(starts_at) WHERE deleted_at IS NULL;

-- full-text / trigram search (pg_trgm)
CREATE INDEX idx_books_title_trgm       ON books   USING GIN (title       gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_books_description_trgm ON books   USING GIN (description gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_authors_name_trgm      ON authors USING GIN (name        gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX idx_genres_slug            ON genres  (slug) WHERE deleted_at IS NULL;
