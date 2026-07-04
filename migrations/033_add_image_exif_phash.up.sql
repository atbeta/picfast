ALTER TABLE images ADD COLUMN exif_data JSONB;
ALTER TABLE images ADD COLUMN phash BIGINT;

CREATE INDEX idx_images_phash ON images(phash);
