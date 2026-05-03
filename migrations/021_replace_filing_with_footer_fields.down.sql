ALTER TABLE site_settings
    ADD COLUMN icp_number TEXT NOT NULL DEFAULT '',
    ADD COLUMN icp_link TEXT NOT NULL DEFAULT '',
    ADD COLUMN psb_number TEXT NOT NULL DEFAULT '',
    ADD COLUMN psb_link TEXT NOT NULL DEFAULT '';

UPDATE site_settings SET
    icp_number = footer_text_1,
    icp_link = footer_link_1,
    psb_number = footer_text_2,
    psb_link = footer_link_2;

ALTER TABLE site_settings
    DROP COLUMN footer_text_1,
    DROP COLUMN footer_link_1,
    DROP COLUMN footer_text_2,
    DROP COLUMN footer_link_2;
