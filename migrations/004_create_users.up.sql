CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    group_id        BIGINT REFERENCES groups(id) ON DELETE SET NULL,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password        VARCHAR(255) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    role            VARCHAR(16) NOT NULL DEFAULT 'user',
    capacity_bytes  BIGINT NOT NULL DEFAULT 0,
    image_num       BIGINT NOT NULL DEFAULT 0,
    album_num       BIGINT NOT NULL DEFAULT 0,
    settings        JSONB NOT NULL DEFAULT '{}',
    status          SMALLINT NOT NULL DEFAULT 1,
    email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    registered_ip   VARCHAR(45) NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_users_email ON users(email);
