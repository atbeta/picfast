-- name: CreateImageModeration :one
INSERT INTO image_moderations (image_id, status, provider)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateImageModeration :one
UPDATE image_moderations SET
    status = $2,
    moderator_id = $3,
    reason = $4,
    updated_at = NOW()
WHERE image_id = $1
RETURNING *;

-- name: UpdateImageModerationStatus :one
UPDATE images SET
    moderation_status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetImageModeration :one
SELECT * FROM image_moderations WHERE image_id = $1;

-- name: ListPendingImages :many
SELECT i.* FROM images i
LEFT JOIN image_moderations m ON i.id = m.image_id
WHERE i.moderation_status = 'pending'
ORDER BY i.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPendingImages :one
SELECT COUNT(*) FROM images WHERE moderation_status = 'pending';

-- name: ApproveAllPendingImages :exec
UPDATE images SET
    moderation_status = 'approved',
    updated_at = NOW()
WHERE moderation_status = 'pending';
