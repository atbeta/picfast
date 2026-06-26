-- WARNING: This migration is destructive. Restoring the columns will
-- require restoring from backup.
ALTER TABLE site_settings
    ADD COLUMN default_copy_format TEXT NOT NULL DEFAULT 'markdown',
    ADD COLUMN copy_template TEXT NOT NULL DEFAULT '';
