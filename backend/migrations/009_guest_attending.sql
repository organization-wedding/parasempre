DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'guests' AND column_name = 'confirmed'
  ) THEN
    ALTER TABLE guests ALTER COLUMN confirmed DROP DEFAULT;
    ALTER TABLE guests ALTER COLUMN confirmed DROP NOT NULL;
    UPDATE guests SET confirmed = NULL WHERE confirmed = false;
    ALTER TABLE guests RENAME COLUMN confirmed TO attending;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS guests_attending_idx ON guests (attending);
