CREATE TABLE IF NOT EXISTS otp_codes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    phone TEXT NOT NULL,
    code TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now(),
    CHECK (phone ~ '^\d{2}9\d{8}$')
);

CREATE INDEX IF NOT EXISTS idx_otp_codes_phone ON otp_codes (phone);
