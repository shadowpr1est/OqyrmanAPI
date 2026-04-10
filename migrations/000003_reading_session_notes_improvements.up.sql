-- #5: Add cfi_position to reading_sessions (for EPUB CFI bookmark persistence)
ALTER TABLE reading_sessions
    ADD COLUMN IF NOT EXISTS cfi_position TEXT;

-- #6: Rename reading_notes.page → position (TEXT) to support both page numbers and EPUB CFI
ALTER TABLE reading_notes
    ADD COLUMN IF NOT EXISTS position TEXT NOT NULL DEFAULT '';

-- Migrate existing page data to position column
UPDATE reading_notes SET position = page::TEXT WHERE position = '' AND page > 0;

ALTER TABLE reading_notes
    DROP COLUMN IF EXISTS page;

-- #7: Add updated_at to reading_notes
ALTER TABLE reading_notes
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Backfill updated_at from created_at for existing rows
UPDATE reading_notes SET updated_at = created_at WHERE updated_at = now();
