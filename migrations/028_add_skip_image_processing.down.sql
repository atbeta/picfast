-- WARNING: This migration is irreversible. The skip_image_processing setting
-- will be permanently lost. Ensure you backup site_settings before downgrading.
ALTER TABLE site_settings DROP COLUMN skip_image_processing;
