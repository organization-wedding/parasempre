CREATE TABLE IF NOT EXISTS gift_messages (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    gift_transaction_id BIGINT NOT NULL UNIQUE
        REFERENCES gift_transactions(id) ON DELETE RESTRICT,
    gift_id BIGINT NOT NULL
        REFERENCES gifts(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL
        REFERENCES users(id) ON DELETE RESTRICT,
    author_name TEXT NOT NULL,
    content TEXT NOT NULL,
    media_object_key TEXT,
    media_kind TEXT,
    media_size_bytes BIGINT,
    media_mime_type TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    deleted_by BIGINT REFERENCES users(id),

    CONSTRAINT gift_messages_content_len CHECK (char_length(content) BETWEEN 1 AND 500),
    CONSTRAINT gift_messages_author_name_len CHECK (char_length(author_name) BETWEEN 1 AND 120),
    CONSTRAINT gift_messages_media_kind_chk CHECK (
        media_kind IS NULL OR media_kind IN ('image', 'audio', 'video')
    ),
    CONSTRAINT gift_messages_media_consistency CHECK (
        (media_object_key IS NULL
            AND media_kind IS NULL
            AND media_size_bytes IS NULL
            AND media_mime_type IS NULL)
        OR
        (media_object_key IS NOT NULL
            AND media_kind IS NOT NULL
            AND media_size_bytes IS NOT NULL
            AND media_mime_type IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS gift_messages_gift_active_idx
    ON gift_messages (gift_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS gift_messages_user_id_idx
    ON gift_messages (user_id);

ALTER TABLE gift_messages ENABLE ROW LEVEL SECURITY;
