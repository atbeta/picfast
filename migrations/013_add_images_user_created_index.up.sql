CREATE INDEX IF NOT EXISTS idx_images_user_created ON images(user_id, created_at DESC);
