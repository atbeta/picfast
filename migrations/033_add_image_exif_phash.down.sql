DROP INDEX IF EXISTS idx_images_phash;
ALTER TABLE images DROP COLUMN phash;
ALTER TABLE images DROP COLUMN exif_data;
