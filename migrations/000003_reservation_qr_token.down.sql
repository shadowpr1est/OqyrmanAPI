DROP INDEX IF EXISTS idx_reservations_qr_token;
ALTER TABLE reservations DROP COLUMN IF EXISTS qr_token;
