ALTER TABLE gift_transactions
    ADD COLUMN IF NOT EXISTS idempotency_key TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS gift_transactions_idempotency_key_unique
    ON gift_transactions (idempotency_key)
    WHERE idempotency_key IS NOT NULL;

ALTER TABLE gift_transactions
    ADD COLUMN IF NOT EXISTS gift_name_snapshot TEXT;

UPDATE gift_transactions tx
   SET gift_name_snapshot = g.name
  FROM gifts g
 WHERE tx.gift_id = g.id
   AND tx.gift_name_snapshot IS NULL;

ALTER TABLE gift_transactions
    ALTER COLUMN gift_name_snapshot SET NOT NULL;
