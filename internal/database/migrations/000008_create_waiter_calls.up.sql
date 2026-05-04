CREATE TABLE IF NOT EXISTS waiter_calls (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    table_id      UUID NOT NULL REFERENCES tables(id) ON DELETE CASCADE,
    status        VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER waiter_calls_set_updated_at
    BEFORE UPDATE ON waiter_calls
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
