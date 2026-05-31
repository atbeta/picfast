ALTER TABLE site_settings
    DROP COLUMN IF EXISTS default_copy_format,
    DROP COLUMN IF EXISTS copy_template;
