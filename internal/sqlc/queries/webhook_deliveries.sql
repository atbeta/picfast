-- name: CreateWebhookDelivery :one
INSERT INTO webhook_deliveries (webhook_id, outbox_event_id, request_headers)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetWebhookDeliveryByID :one
SELECT * FROM webhook_deliveries WHERE id = $1;

-- name: ListWebhookDeliveriesByWebhook :many
SELECT * FROM webhook_deliveries
WHERE webhook_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountWebhookDeliveriesByWebhook :one
SELECT COUNT(*) FROM webhook_deliveries WHERE webhook_id = $1;

-- name: ListPendingWebhookDeliveries :many
SELECT * FROM webhook_deliveries
WHERE status IN ('pending', 'retrying')
  AND next_retry_at <= NOW()
ORDER BY next_retry_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: UpdateWebhookDeliveryStatus :exec
UPDATE webhook_deliveries SET
    status = $2,
    attempt = $3,
    next_retry_at = $4,
    response_status = $5,
    response_body = $6,
    error_message = $7,
    duration_ms = $8,
    completed_at = CASE WHEN $9::boolean THEN NOW() ELSE completed_at END
WHERE id = $1;

-- name: ResetWebhookDeliveryForReplay :exec
UPDATE webhook_deliveries SET
    status = 'pending',
    attempt = 0,
    next_retry_at = NOW(),
    error_message = '',
    completed_at = NULL
WHERE id = $1;
