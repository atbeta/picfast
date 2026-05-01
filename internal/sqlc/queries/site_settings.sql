-- name: GetSiteSettings :one
SELECT * FROM site_settings WHERE id = 1;

-- name: UpsertSiteSettings :one
INSERT INTO site_settings (
    id,
    app_name,
    app_url,
    allow_guest_upload,
    allow_registration,
    require_email_verification,
    user_initial_capacity,
    default_image_ttl,
    moderation_mode
)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE SET
    app_name = EXCLUDED.app_name,
    app_url = EXCLUDED.app_url,
    allow_guest_upload = EXCLUDED.allow_guest_upload,
    allow_registration = EXCLUDED.allow_registration,
    require_email_verification = EXCLUDED.require_email_verification,
    user_initial_capacity = EXCLUDED.user_initial_capacity,
    default_image_ttl = EXCLUDED.default_image_ttl,
    moderation_mode = EXCLUDED.moderation_mode,
    updated_at = NOW()
RETURNING *;
