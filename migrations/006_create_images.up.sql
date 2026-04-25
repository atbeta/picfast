CREATE TABLE images (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT REFERENCES users(id) ON DELETE SET NULL,
    album_id      BIGINT REFERENCES albums(id) ON DELETE SET NULL,
    group_id      BIGINT REFERENCES groups(id) ON DELETE SET NULL,
    strategy_id   BIGINT REFERENCES strategies(id) ON DELETE SET NULL,
    key           VARCHAR(32) NOT NULL UNIQUE,
    path          TEXT NOT NULL DEFAULT '',
    name          VARCHAR(255) NOT NULL,
    origin_name   VARCHAR(255) NOT NULL DEFAULT '',
    size_bytes    BIGINT NOT NULL DEFAULT 0,
    mimetype      VARCHAR(64) NOT NULL,
    extension     VARCHAR(16) NOT NULL,
    md5           CHAR(32) NOT NULL,
    sha1          CHAR(40) NOT NULL,
    width         INTEGER NOT NULL DEFAULT 0,
    height        INTEGER NOT NULL DEFAULT 0,
    permission    SMALLINT NOT NULL DEFAULT 0,
    is_unhealthy  BOOLEAN NOT NULL DEFAULT FALSE,
    uploaded_ip   VARCHAR(45) NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_images_user_id ON images(user_id);
CREATE INDEX idx_images_album_id ON images(album_id);
CREATE INDEX idx_images_key ON images(key);
CREATE INDEX idx_images_dedup ON images(md5, sha1, strategy_id);
CREATE INDEX idx_images_created_at ON images(created_at);
