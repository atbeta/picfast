CREATE TABLE outbox_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type            TEXT NOT NULL,
    version         TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    payload         JSONB NOT NULL,
    owner_user_id   BIGINT REFERENCES users(id) ON DELETE SET NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ,
    CONSTRAINT outbox_events_idempotency_key_unique UNIQUE (idempotency_key)
);

CREATE INDEX idx_outbox_events_status_created
    ON outbox_events(status, created_at)
    WHERE status = 'pending';

CREATE INDEX idx_outbox_events_owner_user_id
    ON outbox_events(owner_user_id, created_at DESC);

CREATE TABLE webhooks (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    url             TEXT NOT NULL,
    secret_hash     CHAR(64) NOT NULL,
    events          JSONB NOT NULL DEFAULT '[]'::jsonb,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_user_id ON webhooks(user_id);
CREATE INDEX idx_webhooks_enabled ON webhooks(enabled) WHERE enabled = TRUE;

CREATE TABLE webhook_deliveries (
    id                BIGSERIAL PRIMARY KEY,
    webhook_id        BIGINT NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    outbox_event_id   UUID NOT NULL REFERENCES outbox_events(id) ON DELETE CASCADE,
    status            TEXT NOT NULL DEFAULT 'pending',
    attempt           INT NOT NULL DEFAULT 0,
    max_attempts      INT NOT NULL DEFAULT 6,
    next_retry_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    request_headers   JSONB NOT NULL DEFAULT '{}'::jsonb,
    response_status   INT,
    response_body     TEXT NOT NULL DEFAULT '',
    error_message     TEXT NOT NULL DEFAULT '',
    duration_ms       INT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at      TIMESTAMPTZ,
    CONSTRAINT webhook_deliveries_webhook_event_unique
        UNIQUE (webhook_id, outbox_event_id)
);

CREATE INDEX idx_webhook_deliveries_pending
    ON webhook_deliveries(status, next_retry_at)
    WHERE status IN ('pending', 'retrying');

CREATE INDEX idx_webhook_deliveries_webhook_id_created
    ON webhook_deliveries(webhook_id, created_at DESC);
