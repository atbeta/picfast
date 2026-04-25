-- name: GetImageByKey :one
SELECT * FROM images WHERE key = $1;

-- name: GetImageByID :one
SELECT * FROM images WHERE id = $1;

-- name: CreateImage :one
INSERT INTO images (user_id, album_id, group_id, strategy_id, key, path, name, origin_name, size_bytes, mimetype, extension, md5, sha1, width, height, permission, uploaded_ip)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
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
SELECT * FROM images
WHERE user_id = $1
AND ($2::bigint IS NULL OR album_id = $2)
AND ($3::smallint IS NULL OR permission = $3)
AND ($4::text IS NULL OR origin_name ILIKE '%' || $4 || '%')
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountImagesByUser :one
SELECT COUNT(*) FROM images
WHERE user_id = $1
AND ($2::bigint IS NULL OR album_id = $2)
AND ($3::smallint IS NULL OR permission = $3)
AND ($4::text IS NULL OR origin_name ILIKE '%' || $4 || '%');

-- name: FindDuplicateImage :one
SELECT * FROM images
WHERE strategy_id = $1 AND md5 = $2 AND sha1 = $3
LIMIT 1;

-- name: CountImagesInWindow :one
SELECT COUNT(*) FROM images
WHERE ($1::bigint IS NULL OR user_id = $1)
AND ($2::text IS NULL OR uploaded_ip = $2)
AND created_at > NOW() - ($3::text || ' seconds')::interval;

-- name: ListAllImages :many
SELECT i.*, u.email as user_email FROM images i
LEFT JOIN users u ON i.user_id = u.id
WHERE ($1::text IS NULL OR i.origin_name ILIKE '%' || $1 || '%')
AND ($2::text IS NULL OR u.email ILIKE '%' || $2 || '%')
AND ($3::text IS NULL OR i.extension = $3)
AND ($4::smallint IS NULL OR i.permission = $4)
AND ($5::boolean IS NULL OR i.is_unhealthy = $5)
ORDER BY i.created_at DESC
LIMIT $6 OFFSET $7;

-- name: CountAllImages :one
SELECT COUNT(*) FROM images i
LEFT JOIN users u ON i.user_id = u.id
WHERE ($1::text IS NULL OR i.origin_name ILIKE '%' || $1 || '%')
AND ($2::text IS NULL OR u.email ILIKE '%' || $2 || '%')
AND ($3::text IS NULL OR i.extension = $3)
AND ($4::smallint IS NULL OR i.permission = $4)
AND ($5::boolean IS NULL OR i.is_unhealthy = $5);
