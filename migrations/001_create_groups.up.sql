CREATE TABLE groups (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(64) NOT NULL,
    is_default    BOOLEAN NOT NULL DEFAULT FALSE,
    is_guest      BOOLEAN NOT NULL DEFAULT FALSE,
    configs       JSONB NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
