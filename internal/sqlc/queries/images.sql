-- name: GetImageByKey :one
SELECT * FROM images WHERE key = $1;

-- name: GetImageByID :one
SELECT * FROM images WHERE id = $1;

-- name: CreateImage :one
INSERT INTO images (user_id, album_id, group_id, strategy_id, key, path, name, origin_name, size_bytes, mimetype, extension, md5, sha1, width, height, permission, uploaded_ip, expires_at, exif_data, phash)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
RETURNING *;

-- name: DeleteImage :exec
DELETE FROM images WHERE id = $1;

-- name: DeleteImageByKey :exec
DELETE FROM images WHERE key = $1;

-- name: UpdateImage :one
UPDATE images SET
    album_id = COALESCE($2, album_id),
    permission = COALESCE($3, permission),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListImagesByUser :many
SELECT
    images.*,
    strategies.name as strategy_name,
    strategies.strategy_type as strategy_type
FROM images
LEFT JOIN strategies ON images.strategy_id = strategies.id
WHERE images.user_id = $1
  AND (sqlc.narg('album_id')::bigint IS NULL OR images.album_id = sqlc.narg('album_id'))
  AND (sqlc.narg('keyword')::text IS NULL OR images.origin_name ILIKE '%' || sqlc.narg('keyword') || '%')
  AND (sqlc.narg('extension')::text IS NULL OR images.extension = sqlc.narg('extension'))
  AND (sqlc.narg('date_from')::timestamptz IS NULL OR images.created_at >= sqlc.narg('date_from'))
  AND (sqlc.narg('date_to')::timestamptz IS NULL OR images.created_at <= sqlc.narg('date_to'))
  AND (sqlc.narg('tag_ids')::bigint[] IS NULL OR EXISTS (SELECT 1 FROM image_tags WHERE image_tags.image_id = images.id AND tag_id = ANY(sqlc.narg('tag_ids')::bigint[])))
ORDER BY images.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountImagesByUser :one
SELECT COUNT(*) FROM images
WHERE user_id = $1
  AND (sqlc.narg('album_id')::bigint IS NULL OR album_id = sqlc.narg('album_id'))
  AND (sqlc.narg('keyword')::text IS NULL OR origin_name ILIKE '%' || sqlc.narg('keyword') || '%')
  AND (sqlc.narg('extension')::text IS NULL OR extension = sqlc.narg('extension'))
  AND (sqlc.narg('date_from')::timestamptz IS NULL OR created_at >= sqlc.narg('date_from'))
  AND (sqlc.narg('date_to')::timestamptz IS NULL OR created_at <= sqlc.narg('date_to'))
  AND (sqlc.narg('tag_ids')::bigint[] IS NULL OR EXISTS (SELECT 1 FROM image_tags WHERE image_tags.image_id = images.id AND tag_id = ANY(sqlc.narg('tag_ids')::bigint[])));

-- name: FindDuplicateImage :one
SELECT * FROM images
WHERE strategy_id = $1 AND md5 = $2 AND sha1 = $3
LIMIT 1;

-- name: CountImagesInWindow :one
SELECT COUNT(*) FROM images
WHERE user_id = $1
AND created_at > NOW() - ($2::text || ' seconds')::interval;

-- name: CountImagesInWindowByIP :one
SELECT COUNT(*) FROM images
WHERE user_id IS NULL AND uploaded_ip = $1
AND created_at > NOW() - ($2::text || ' seconds')::interval;

-- name: ListAllImages :many
SELECT images.*, users.email as user_email FROM images
LEFT JOIN users ON images.user_id = users.id
WHERE (sqlc.narg('keyword')::text IS NULL OR images.origin_name ILIKE '%' || sqlc.narg('keyword') || '%')
  AND (sqlc.narg('email')::text IS NULL OR users.email ILIKE '%' || sqlc.narg('email') || '%')
  AND (sqlc.narg('extension')::text IS NULL OR images.extension = sqlc.narg('extension'))
  AND (sqlc.narg('date_from')::timestamptz IS NULL OR images.created_at >= sqlc.narg('date_from'))
  AND (sqlc.narg('date_to')::timestamptz IS NULL OR images.created_at <= sqlc.narg('date_to'))
  AND (sqlc.narg('tag_ids')::bigint[] IS NULL OR EXISTS (SELECT 1 FROM image_tags WHERE image_tags.image_id = images.id AND tag_id = ANY(sqlc.narg('tag_ids')::bigint[])))
ORDER BY images.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAllImages :one
SELECT COUNT(*) FROM images
LEFT JOIN users ON images.user_id = users.id
WHERE (sqlc.narg('keyword')::text IS NULL OR images.origin_name ILIKE '%' || sqlc.narg('keyword') || '%')
  AND (sqlc.narg('email')::text IS NULL OR users.email ILIKE '%' || sqlc.narg('email') || '%')
  AND (sqlc.narg('extension')::text IS NULL OR images.extension = sqlc.narg('extension'))
  AND (sqlc.narg('date_from')::timestamptz IS NULL OR images.created_at >= sqlc.narg('date_from'))
  AND (sqlc.narg('date_to')::timestamptz IS NULL OR images.created_at <= sqlc.narg('date_to'))
  AND (sqlc.narg('tag_ids')::bigint[] IS NULL OR EXISTS (SELECT 1 FROM image_tags WHERE image_tags.image_id = images.id AND tag_id = ANY(sqlc.narg('tag_ids')::bigint[])));

-- name: GetGuestUsedCapacity :one
SELECT COALESCE(SUM(size_bytes), 0)::bigint FROM images WHERE user_id IS NULL;

-- name: CountImagesByGroup :one
SELECT COUNT(*) FROM images WHERE group_id = $1;

-- name: GetExpiredImages :many
SELECT * FROM images
WHERE expires_at IS NOT NULL AND expires_at <= NOW()
ORDER BY expires_at ASC
LIMIT $1;

-- name: UpdateImageExpiration :exec
UPDATE images SET expires_at = $2, updated_at = NOW() WHERE id = $1;

-- name: GetImagesByMD5 :many
SELECT * FROM images WHERE md5 = $1;

-- name: CountImagesByStrategy :one
SELECT COUNT(*) FROM images WHERE strategy_id = $1;

-- name: CountExpiredImages :one
SELECT COUNT(*) FROM images WHERE expires_at IS NOT NULL AND expires_at <= NOW();

-- name: GetImagesForPHashRecalc :many
SELECT * FROM images WHERE id > sqlc.arg('after_id') ORDER BY id LIMIT sqlc.arg('batch_size');

-- name: GetMaxImageID :one
SELECT COALESCE(MAX(id), 0)::bigint FROM images;

-- name: UpdateImagePHash :exec
UPDATE images SET phash = $2 WHERE id = $1;

-- name: UpdateImageOCR :exec
UPDATE images SET ocr_text = $2 WHERE id = $1;
