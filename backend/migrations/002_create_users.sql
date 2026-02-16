CREATE TABLE IF NOT EXISTS users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    guest_id BIGINT REFERENCES guests(id),
    role TEXT NOT NULL DEFAULT 'guest',
    uracf TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT users_role_check CHECK (role IN ('guest', 'groom', 'bride')),
    CONSTRAINT users_uracf_unique UNIQUE (uracf),
    CONSTRAINT users_uracf_check CHECK (uracf ~ '^[A-Z0-9]{5}$')
);

CREATE INDEX ON users (guest_id);
