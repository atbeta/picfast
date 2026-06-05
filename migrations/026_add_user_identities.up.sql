ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

CREATE TABLE user_identities (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider         VARCHAR(32)  NOT NULL,
    provider_subject VARCHAR(255) NOT NULL,
    email            VARCHAR(255) NOT NULL,
    linked_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_subject)
);
CREATE INDEX idx_user_identities_user_id ON user_identities(user_id);
