-- name: CreateWebhook :one
INSERT INTO webhooks (user_id, name, url, secret_hash, secret_ciphertext, events)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetWebhookByID :one
SELECT * FROM webhooks WHERE id = $1;

-- name: ListWebhooksByUser :many
SELECT * FROM webhooks
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CountWebhooksByUser :one
SELECT COUNT(*) FROM webhooks WHERE user_id = $1;

-- name: UpdateWebhook :one
UPDATE webhooks SET
    name = COALESCE($2, name),
    url = COALESCE($3, url),
    events = COALESCE($4, events),
    enabled = COALESCE($5, enabled),
    updated_at = NOW()
WHERE id = $1 AND user_id = $6
RETURNING *;

-- name: UpdateWebhookSecret :exec
UPDATE webhooks SET
    secret_hash = $2,
    secret_ciphertext = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteWebhook :exec
DELETE FROM webhooks WHERE id = $1 AND user_id = $2;

-- name: ListEnabledWebhooks :many
SELECT * FROM webhooks
WHERE enabled = TRUE
ORDER BY id;