-- name: CreateTag :one
INSERT INTO tags (user_id, name, type)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTagsByUserID :many
SELECT * FROM tags
WHERE user_id = $1
ORDER BY name ASC;

-- name: AddImageTag :exec
INSERT INTO image_tags (image_id, tag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveImageTag :exec
DELETE FROM image_tags
WHERE image_id = $1 AND tag_id = $2;

-- name: GetTagByID :one
SELECT * FROM tags WHERE id = $1;

-- name: GetImageTags :many
SELECT t.*
FROM tags t
JOIN image_tags it ON t.id = it.tag_id
WHERE it.image_id = $1
ORDER BY t.name ASC;
