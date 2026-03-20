CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
                       id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       email         VARCHAR(255) NOT NULL UNIQUE,
                       phone         VARCHAR(20)  NOT NULL UNIQUE,
                       password_hash TEXT         NOT NULL,
                       full_name     VARCHAR(255) NOT NULL DEFAULT '',
                       avatar_url    TEXT         NOT NULL DEFAULT '',
                       role          VARCHAR(20)  NOT NULL DEFAULT 'User',
                       qr_code       TEXT         NOT NULL DEFAULT '',
                       created_at    TIMESTAMP    NOT NULL DEFAULT now()
);

CREATE TABLE tokens (
                        id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                        refresh_token TEXT NOT NULL UNIQUE,
                        expires_at    TIMESTAMP NOT NULL,
                        created_at    TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE authors (
                         id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                         name       VARCHAR(255) NOT NULL,
                         bio        TEXT         NOT NULL DEFAULT '',
                         birth_date DATE,
                         death_date DATE,
                         photo_url  TEXT         NOT NULL DEFAULT ''
);

CREATE TABLE genres (
                        id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        name VARCHAR(100) NOT NULL UNIQUE,
                        slug VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE books (
                       id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       author_id   UUID         NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
    -- FIX: было NOT NULL + ON DELETE SET NULL — противоречие, PostgreSQL выбрасывал ошибку
    -- при удалении жанра. Теперь RESTRICT: нельзя удалить жанр, пока к нему привязаны книги.
                       genre_id    UUID         NOT NULL REFERENCES genres(id)  ON DELETE RESTRICT,
                       title       VARCHAR(255) NOT NULL,
                       isbn        VARCHAR(20)  NOT NULL DEFAULT '',
                       cover_url   TEXT         NOT NULL DEFAULT '',
                       description TEXT         NOT NULL DEFAULT '',
                       language    VARCHAR(50)  NOT NULL DEFAULT '',
                       year        INT          NOT NULL DEFAULT 0,
                       avg_rating  FLOAT        NOT NULL DEFAULT 0
);

CREATE TABLE book_files (
                            id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                            book_id  UUID         NOT NULL REFERENCES books(id) ON DELETE CASCADE,
                            format   VARCHAR(20)  NOT NULL,
                            file_url TEXT         NOT NULL,
                            is_audio BOOLEAN      NOT NULL DEFAULT false
);

CREATE TABLE reading_sessions (
                                  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                  user_id      UUID        NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
                                  book_id      UUID        NOT NULL REFERENCES books(id)  ON DELETE CASCADE,
                                  current_page INT         NOT NULL DEFAULT 0,
                                  status       VARCHAR(20) NOT NULL DEFAULT 'reading',
                                  updated_at   TIMESTAMP   NOT NULL DEFAULT now(),
                                  UNIQUE (user_id, book_id)
);

CREATE TABLE wishlists (
                           id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           user_id  UUID      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                           book_id  UUID      NOT NULL REFERENCES books(id) ON DELETE CASCADE,
                           added_at TIMESTAMP NOT NULL DEFAULT now(),
                           UNIQUE (user_id, book_id)
);

CREATE TABLE reading_notes (
                               id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                               user_id    UUID      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                               book_id    UUID      NOT NULL REFERENCES books(id) ON DELETE CASCADE,
                               page       INT       NOT NULL DEFAULT 0,
                               content    TEXT      NOT NULL,
                               created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE libraries (
                           id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           name    VARCHAR(255) NOT NULL,
                           address TEXT         NOT NULL,
                           lat     FLOAT        NOT NULL,
                           lng     FLOAT        NOT NULL,
                           phone   VARCHAR(20)  NOT NULL DEFAULT ''
);

CREATE TABLE library_books (
                               id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                               library_id       UUID NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
                               book_id          UUID NOT NULL REFERENCES books(id)     ON DELETE CASCADE,
                               total_copies     INT  NOT NULL DEFAULT 0,
                               available_copies INT  NOT NULL DEFAULT 0,
                               UNIQUE (library_id, book_id)
);

CREATE TABLE book_machines (
                               id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                               name    VARCHAR(255) NOT NULL,
                               address TEXT         NOT NULL,
                               lat     FLOAT        NOT NULL,
                               lng     FLOAT        NOT NULL,
                               status  VARCHAR(20)  NOT NULL DEFAULT 'active'
);

CREATE TABLE book_machine_books (
                                    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                    machine_id       UUID NOT NULL REFERENCES book_machines(id) ON DELETE CASCADE,
                                    book_id          UUID NOT NULL REFERENCES books(id)         ON DELETE CASCADE,
                                    total_copies     INT  NOT NULL DEFAULT 0,
                                    available_copies INT  NOT NULL DEFAULT 0,
                                    UNIQUE (machine_id, book_id)
);

CREATE TABLE reservations (
                              id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                              user_id         UUID        NOT NULL REFERENCES users(id)             ON DELETE CASCADE,
                              library_book_id UUID                 REFERENCES library_books(id)     ON DELETE SET NULL,
                              machine_book_id UUID                 REFERENCES book_machine_books(id) ON DELETE SET NULL,
                              source_type     VARCHAR(20) NOT NULL,
                              status          VARCHAR(20) NOT NULL DEFAULT 'pending',
                              reserved_at     TIMESTAMP   NOT NULL DEFAULT now(),
                              due_date        TIMESTAMP   NOT NULL,
                              returned_at     TIMESTAMP
);

CREATE TABLE reviews (
                         id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                         user_id    UUID      NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
                         book_id    UUID      NOT NULL REFERENCES books(id)  ON DELETE CASCADE,
                         rating     INT       NOT NULL CHECK (rating >= 1 AND rating <= 5),
                         body       TEXT      NOT NULL DEFAULT '',
                         created_at TIMESTAMP NOT NULL DEFAULT now(),
                         UNIQUE (user_id, book_id)
);