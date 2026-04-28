DROP INDEX IF EXISTS idx_images_expires_at;
ALTER TABLE images DROP COLUMN IF EXISTS expires_at;
