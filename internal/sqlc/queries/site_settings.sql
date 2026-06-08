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
    allow_oauth_registration,
    require_email_verification,
    user_initial_capacity,
    default_image_ttl,
    guest_image_ttl,
    moderation_mode,
    site_description,
    favicon_url,
    footer_text_1,
    footer_link_1,
    footer_text_2,
    footer_link_2,
    analytics_provider,
    analytics_config,
    allow_user_image_processing,
    skip_image_processing,
    theme_config,
    default_copy_format,
    copy_template
)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
ON CONFLICT (id) DO UPDATE SET
    app_name = EXCLUDED.app_name,
    app_url = EXCLUDED.app_url,
    allow_guest_upload = EXCLUDED.allow_guest_upload,
    guest_capacity_bytes = EXCLUDED.guest_capacity_bytes,
    allow_registration = EXCLUDED.allow_registration,
    allow_oauth_registration = EXCLUDED.allow_oauth_registration,
    require_email_verification = EXCLUDED.require_email_verification,
    user_initial_capacity = EXCLUDED.user_initial_capacity,
    default_image_ttl = EXCLUDED.default_image_ttl,
    guest_image_ttl = EXCLUDED.guest_image_ttl,
    moderation_mode = EXCLUDED.moderation_mode,
    site_description = EXCLUDED.site_description,
    favicon_url = EXCLUDED.favicon_url,
    footer_text_1 = EXCLUDED.footer_text_1,
    footer_link_1 = EXCLUDED.footer_link_1,
    footer_text_2 = EXCLUDED.footer_text_2,
    footer_link_2 = EXCLUDED.footer_link_2,
    analytics_provider = EXCLUDED.analytics_provider,
    analytics_config = EXCLUDED.analytics_config,
    allow_user_image_processing = EXCLUDED.allow_user_image_processing,
    skip_image_processing = EXCLUDED.skip_image_processing,
    theme_config = EXCLUDED.theme_config,
    default_copy_format = EXCLUDED.default_copy_format,
    copy_template = EXCLUDED.copy_template,
    updated_at = NOW()
RETURNING *;
