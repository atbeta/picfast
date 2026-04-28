CREATE TABLE api_tokens (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,
    token_hash   CHAR(64) NOT NULL UNIQUE,
    scopes       JSONB NOT NULL DEFAULT '["read", "write"]',
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_tokens_user_id ON api_tokens(user_id);
CREATE INDEX idx_api_tokens_hash ON api_tokens(token_hash);
