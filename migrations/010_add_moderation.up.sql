ALTER TABLE images ADD COLUMN moderation_status VARCHAR(16) NOT NULL DEFAULT 'approved';
CREATE INDEX idx_images_moderation_status ON images(moderation_status);
