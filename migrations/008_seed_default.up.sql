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

INSERT INTO strategies (name, strategy_type, configs) VALUES (
    'Default Local',
    'local',
    '{"root": "./data/uploads", "url": "/i"}'::jsonb
);

INSERT INTO group_strategies (group_id, strategy_id) VALUES (1, 1);
