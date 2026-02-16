CREATE TABLE IF NOT EXISTS guests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nome TEXT NOT NULL,
    sobrenome TEXT NOT NULL,
    telefone TEXT NOT NULL,
    relacionamento TEXT NOT NULL CHECK (relacionamento IN ('noivo', 'noiva')),
    confirmacao BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
