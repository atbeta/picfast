-- Seeds the default group only. The default storage strategy is deliberately
-- NOT seeded here: bootstrapCoreData creates it on first start and honours
-- PICFAST_STORAGE_DEFAULT_TYPE / PICFAST_STORAGE_DEFAULT_CONFIG. Seeding a
-- local strategy here made every fresh database look like it "already had
-- strategies", so the environment default was silently ignored.
INSERT INTO groups (name, is_default, is_guest, configs) VALUES (
    'Default',
    TRUE,
    TRUE,
    '{
        "maximum_file_size": 5242880,
        "accepted_extensions": ["jpeg","jpg","png","gif","webp","bmp","svg","ico"],
        "limit_per_minute": 20,
        "limit_per_hour": 100,
        "limit_per_day": 300,
        "limit_per_month": 999,
        "path_naming_rule": "{Y}/{m}/{d}",
        "file_naming_rule": "{uniqid}",
        "image_save_quality": 75,
        "image_save_format": "",
        "is_enable_watermark": false,
        "watermark_configs": {},
        "is_enable_original_protection": false
    }'::jsonb
);
