CREATE TABLE image_moderations (
    id            BIGSERIAL PRIMARY KEY,
    image_id      BIGINT NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    status        VARCHAR(16) NOT NULL DEFAULT 'pending',
    provider      VARCHAR(64) NOT NULL DEFAULT '',
    score         NUMERIC(5,4) NOT NULL DEFAULT 0,
    labels        JSONB NOT NULL DEFAULT '[]',
    reason        TEXT NOT NULL DEFAULT '',
    moderator_id  BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_image_moderations_image_id ON image_moderations(image_id);
CREATE INDEX idx_image_moderations_status ON image_moderations(status);
