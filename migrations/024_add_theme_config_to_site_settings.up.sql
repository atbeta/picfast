ALTER TABLE site_settings
    ADD COLUMN theme_config JSONB NOT NULL DEFAULT '{}';
