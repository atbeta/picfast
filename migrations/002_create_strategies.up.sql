CREATE TABLE strategies (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(64) NOT NULL,
    strategy_type VARCHAR(16) NOT NULL,
    configs       JSONB NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
