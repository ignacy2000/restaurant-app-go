CREATE TABLE IF NOT EXISTS allowed_origins (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    origin     VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
