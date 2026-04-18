CREATE TABLE IF NOT EXISTS gifts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price_cents BIGINT NOT NULL,
    image_url TEXT,
    store_url TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    dedupe_key TEXT NOT NULL,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT gifts_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT gifts_description_max_len CHECK (description IS NULL OR length(description) <= 2000),
    CONSTRAINT gifts_price_positive CHECK (price_cents > 0),
    CONSTRAINT gifts_status_check CHECK (status IN ('active', 'inactive')),
    CONSTRAINT gifts_created_by_racf CHECK (created_by ~ '^[A-Z0-9]{5}$'),
    CONSTRAINT gifts_updated_by_racf CHECK (updated_by ~ '^[A-Z0-9]{5}$'),
    CONSTRAINT gifts_image_url_https CHECK (image_url IS NULL OR image_url ~* '^https://'),
    CONSTRAINT gifts_store_url_https CHECK (store_url IS NULL OR store_url ~* '^https://')
);

CREATE UNIQUE INDEX IF NOT EXISTS gifts_dedupe_key_active_unique
    ON gifts (dedupe_key)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS gifts_status_idx ON gifts (status);
CREATE INDEX IF NOT EXISTS gifts_deleted_at_idx ON gifts (deleted_at);

CREATE TABLE IF NOT EXISTS gift_transactions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    gift_id BIGINT NOT NULL REFERENCES gifts(id),
    user_id BIGINT NOT NULL REFERENCES users(id),
    payment_method TEXT NOT NULL,
    mp_payment_id TEXT,
    mp_preference_id TEXT,
    amount_cents BIGINT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT gift_transactions_payment_method_check CHECK (payment_method IN ('credit_card', 'pix')),
    CONSTRAINT gift_transactions_amount_positive CHECK (amount_cents > 0),
    CONSTRAINT gift_transactions_status_check CHECK (status IN ('pending', 'approved', 'rejected', 'refunded', 'cancelled')),
    CONSTRAINT gift_transactions_mp_payment_id_unique UNIQUE (mp_payment_id)
);

CREATE INDEX IF NOT EXISTS gift_transactions_gift_id_idx ON gift_transactions (gift_id);
CREATE INDEX IF NOT EXISTS gift_transactions_user_id_idx ON gift_transactions (user_id);
CREATE INDEX IF NOT EXISTS gift_transactions_status_idx ON gift_transactions (status);
