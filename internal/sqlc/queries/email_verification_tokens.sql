-- name: CreateEmailVerificationToken :one
INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetEmailVerificationTokenByHash :one
SELECT * FROM email_verification_tokens WHERE token_hash = $1;

-- name: GetLatestUnusedEmailVerificationTokenByUser :one
SELECT *
FROM email_verification_tokens
WHERE user_id = $1 AND used_at IS NULL
ORDER BY created_at DESC
LIMIT 1;

-- name: MarkEmailVerificationTokenUsed :one
UPDATE email_verification_tokens
SET used_at = NOW()
WHERE id = $1 AND used_at IS NULL
RETURNING *;

-- name: DeleteUnusedEmailVerificationTokensByUser :exec
DELETE FROM email_verification_tokens
WHERE user_id = $1 AND used_at IS NULL;

-- name: DeleteEmailVerificationTokenByHash :exec
DELETE FROM email_verification_tokens
WHERE token_hash = $1;
