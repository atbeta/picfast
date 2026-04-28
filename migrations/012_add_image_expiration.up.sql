ALTER TABLE images ADD COLUMN expires_at TIMESTAMPTZ;
CREATE INDEX idx_images_expires_at ON images(expires_at) WHERE expires_at IS NOT NULL;
