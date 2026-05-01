CREATE TABLE site_settings (
    id                         SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    app_name                   TEXT NOT NULL,
    app_url                    TEXT NOT NULL,
    allow_guest_upload         BOOLEAN NOT NULL DEFAULT FALSE,
    allow_registration         BOOLEAN NOT NULL DEFAULT FALSE,
    require_email_verification BOOLEAN NOT NULL DEFAULT FALSE,
    user_initial_capacity      BIGINT NOT NULL DEFAULT 524288000,
    default_image_ttl          TEXT NOT NULL DEFAULT '',
    moderation_mode            TEXT NOT NULL DEFAULT 'disabled',
    created_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
