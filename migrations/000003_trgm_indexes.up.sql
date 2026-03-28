-- pg_trgm позволяет PostgreSQL использовать GIN-индексы для ILIKE '%pattern%'.
-- Без этого все поисковые запросы делали seq scan по всей таблице.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Поиск книг по названию и описанию (book_repo.Search)
CREATE INDEX IF NOT EXISTS idx_books_title_trgm       ON books       USING GIN (title       gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_books_description_trgm ON books       USING GIN (description gin_trgm_ops) WHERE deleted_at IS NULL;

-- Поиск авторов по имени (library_book_repo.SearchInLibrary)
CREATE INDEX IF NOT EXISTS idx_authors_name_trgm      ON authors     USING GIN (name        gin_trgm_ops) WHERE deleted_at IS NULL;

-- Индекс для поиска жанров по slug (genre_repo.GetBySlug) — B-tree достаточно, но добавляем явно
CREATE INDEX IF NOT EXISTS idx_genres_slug            ON genres      (slug) WHERE deleted_at IS NULL;
