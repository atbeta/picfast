ALTER TABLE site_settings
    ADD COLUMN show_powered_by BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN guest_upload_notice_title TEXT NOT NULL DEFAULT '游客上传',
    ADD COLUMN guest_upload_notice_subtitle TEXT NOT NULL DEFAULT '游客上传不提供后续管理能力，建议注册账号以使用完整功能。',
    ADD COLUMN show_login_link BOOLEAN NOT NULL DEFAULT TRUE;
