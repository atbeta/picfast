-- name: CreateAPIToken :one
INSERT INTO api_tokens (user_id, name, token_hash, scopes, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAPITokenByHash :one
SELECT t.*, u.id as user_id, u.role, u.group_id, u.status
FROM api_tokens t
JOIN users u ON t.user_id = u.id
WHERE t.token_hash = $1;

-- name: ListAPITokensByUser :many
SELECT * FROM api_tokens WHERE user_id = $1 ORDER BY created_at DESC;

-- name: DeleteAPIToken :exec
DELETE FROM api_tokens WHERE id = $1 AND user_id = $2;

-- name: UpdateAPITokenLastUsed :exec
UPDATE api_tokens SET last_used_at = NOW() WHERE id = $1;
