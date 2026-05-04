CREATE TABLE IF NOT EXISTS menu_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    menu_id     UUID NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    price       NUMERIC(10,2) NOT NULL DEFAULT 0,
    position    INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
