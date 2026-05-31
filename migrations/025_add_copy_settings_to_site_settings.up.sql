ALTER TABLE site_settings
    ADD COLUMN IF NOT EXISTS default_copy_format TEXT NOT NULL DEFAULT 'markdown',
    ADD COLUMN IF NOT EXISTS copy_template TEXT NOT NULL DEFAULT '';
