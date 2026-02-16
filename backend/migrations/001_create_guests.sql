CREATE TABLE IF NOT EXISTS guests (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    phone TEXT,
    relationship TEXT NOT NULL,
    confirmed BOOLEAN NOT NULL DEFAULT false,
    family_group BIGINT NOT NULL,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT guests_name_unique UNIQUE (first_name, last_name),
    CONSTRAINT guests_phone_unique UNIQUE (phone),
    CONSTRAINT guests_relationship_check CHECK (relationship IN ('P', 'R')),
    CONSTRAINT guests_phone_check CHECK (phone ~ '^\d{2}9\d{8}$'),
    CONSTRAINT guests_created_by_racf CHECK (created_by ~ '^[A-Z0-9]{5}$'),
    CONSTRAINT guests_updated_by_racf CHECK (updated_by ~ '^[A-Z0-9]{5}$')
);
