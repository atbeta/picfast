ALTER TABLE site_settings
    ADD COLUMN footer_text_1 TEXT NOT NULL DEFAULT '',
    ADD COLUMN footer_link_1 TEXT NOT NULL DEFAULT '',
    ADD COLUMN footer_text_2 TEXT NOT NULL DEFAULT '',
    ADD COLUMN footer_link_2 TEXT NOT NULL DEFAULT '';

UPDATE site_settings SET
    footer_text_1 = icp_number,
    footer_link_1 = icp_link,
    footer_text_2 = psb_number,
    footer_link_2 = psb_link;

ALTER TABLE site_settings
    DROP COLUMN icp_number,
    DROP COLUMN icp_link,
    DROP COLUMN psb_number,
    DROP COLUMN psb_link;
