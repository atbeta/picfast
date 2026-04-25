-- name: GetAlbumByID :one
SELECT * FROM albums WHERE id = $1;

-- name: CreateAlbum :one
INSERT INTO albums (user_id, name, intro)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateAlbum :one
UPDATE albums SET
    name = COALESCE($2, name),
    intro = COALESCE($3, intro),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteAlbum :exec
DELETE FROM albums WHERE id = $1;

-- name: ListAlbumsByUser :many
SELECT * FROM albums
WHERE user_id = $1
AND ($2::text IS NULL OR name ILIKE '%' || $2 || '%')
ORDER BY
    CASE WHEN $3 = 'newest' THEN created_at END DESC,
    CASE WHEN $3 = 'earliest' THEN created_at END ASC,
    CASE WHEN $3 = 'most' THEN image_num END DESC,
    CASE WHEN $3 = 'least' THEN image_num END ASC,
    created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountAlbumsByUser :one
SELECT COUNT(*) FROM albums
WHERE user_id = $1
AND ($2::text IS NULL OR name ILIKE '%' || $2 || '%');

-- name: IncrementAlbumImageNum :exec
UPDATE albums SET image_num = image_num + 1, updated_at = NOW() WHERE id = $1;

-- name: DecrementAlbumImageNum :exec
UPDATE albums SET image_num = image_num - 1, updated_at = NOW() WHERE id = $1 AND image_num > 0;

-- name: IncrementUserAlbumNum :exec
UPDATE users SET album_num = album_num + 1, updated_at = NOW() WHERE id = $1;

-- name: DecrementUserAlbumNum :exec
UPDATE users SET album_num = album_num - 1, updated_at = NOW() WHERE id = $1 AND album_num > 0;
