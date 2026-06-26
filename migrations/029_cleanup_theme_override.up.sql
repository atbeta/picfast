-- Drop the legacy users.settings.theme_override key for users who set
-- a personal theme override before the theme system was collapsed to a
-- single built-in default. The key is no longer read or written by the
-- application; leaving it in place would just be dead weight in JSONB.
UPDATE users
SET settings = settings - 'theme_override'
WHERE settings ? 'theme_override';
