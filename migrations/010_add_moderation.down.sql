DROP INDEX IF EXISTS idx_images_moderation_status;
ALTER TABLE images DROP COLUMN IF EXISTS moderation_status;
