ALTER TABLE site_settings ADD COLUMN allow_oauth_registration BOOLEAN NOT NULL DEFAULT FALSE;
UPDATE site_settings SET allow_oauth_registration = allow_registration;
