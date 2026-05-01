ALTER TABLE site_settings
    ADD COLUMN site_description TEXT NOT NULL DEFAULT '',
    ADD COLUMN icp_number TEXT NOT NULL DEFAULT '',
    ADD COLUMN icp_link TEXT NOT NULL DEFAULT '',
    ADD COLUMN psb_number TEXT NOT NULL DEFAULT '',
    ADD COLUMN psb_link TEXT NOT NULL DEFAULT '',
    ADD COLUMN analytics_provider TEXT NOT NULL DEFAULT '',
    ADD COLUMN analytics_config JSONB NOT NULL DEFAULT '{}'::jsonb;
