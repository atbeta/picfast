-- name: GetSiteSettings :one
SELECT * FROM site_settings WHERE id = 1;

-- name: UpsertSiteSettings :one
INSERT INTO site_settings (
    id,
    app_name,
    app_url,
    allow_guest_upload,
    guest_capacity_bytes,
    allow_registration,
    require_email_verification,
    user_initial_capacity,
    default_image_ttl,
    moderation_mode,
    site_description,
    favicon_url,
    icp_number,
    icp_link,
    psb_number,
    psb_link,
    analytics_provider,
    analytics_config
)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
ON CONFLICT (id) DO UPDATE SET
    app_name = EXCLUDED.app_name,
    app_url = EXCLUDED.app_url,
    allow_guest_upload = EXCLUDED.allow_guest_upload,
    guest_capacity_bytes = EXCLUDED.guest_capacity_bytes,
    allow_registration = EXCLUDED.allow_registration,
    require_email_verification = EXCLUDED.require_email_verification,
    user_initial_capacity = EXCLUDED.user_initial_capacity,
    default_image_ttl = EXCLUDED.default_image_ttl,
    moderation_mode = EXCLUDED.moderation_mode,
    site_description = EXCLUDED.site_description,
    favicon_url = EXCLUDED.favicon_url,
    icp_number = EXCLUDED.icp_number,
    icp_link = EXCLUDED.icp_link,
    psb_number = EXCLUDED.psb_number,
    psb_link = EXCLUDED.psb_link,
    analytics_provider = EXCLUDED.analytics_provider,
    analytics_config = EXCLUDED.analytics_config,
    updated_at = NOW()
RETURNING *;
