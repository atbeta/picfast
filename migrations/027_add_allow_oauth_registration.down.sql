-- WARNING: This migration is irreversible. The allow_oauth_registration setting
-- will be permanently lost. Ensure you backup site_settings before downgrading.
ALTER TABLE site_settings DROP COLUMN allow_oauth_registration;
