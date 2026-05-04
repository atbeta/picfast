-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPasswordResetTokenByHash :one
SELECT * FROM password_reset_tokens WHERE token_hash = $1;

-- name: MarkPasswordResetTokenUsed :one
UPDATE password_reset_tokens
SET used_at = NOW()
WHERE id = $1 AND used_at IS NULL
RETURNING *;

-- name: DeleteUnusedPasswordResetTokensByUser :exec
DELETE FROM password_reset_tokens
WHERE user_id = $1 AND used_at IS NULL;

-- name: DeletePasswordResetTokenByHash :exec
DELETE FROM password_reset_tokens
WHERE token_hash = $1;
