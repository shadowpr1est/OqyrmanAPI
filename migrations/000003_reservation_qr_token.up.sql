ALTER TABLE reservations ADD COLUMN qr_token TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX idx_reservations_qr_token ON reservations(qr_token) WHERE qr_token != '';
