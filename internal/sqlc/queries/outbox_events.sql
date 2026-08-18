-- name: InsertOutboxEvent :one
INSERT INTO outbox_events (type, version, idempotency_key, payload, owner_user_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetOutboxEventByKey :one
SELECT * FROM outbox_events WHERE idempotency_key = $1;

-- name: MarkOutboxEventDispatched :exec
UPDATE outbox_events
SET status = 'dispatched', processed_at = NOW()
WHERE id = $1;

-- name: MarkOutboxEventFailed :exec
UPDATE outbox_events
SET status = 'failed', processed_at = NOW()
WHERE id = $1;

-- name: ListPendingOutboxEvents :many
SELECT * FROM outbox_events
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: GetOutboxEventByID :one
SELECT * FROM outbox_events WHERE id = $1;

-- name: DeleteOldOutboxEvents :execrows
DELETE FROM outbox_events
WHERE status IN ('dispatched', 'failed')
  AND created_at < NOW() - make_interval(days => $1)
  AND NOT EXISTS (
      SELECT 1 FROM webhook_deliveries d
      WHERE d.outbox_event_id = outbox_events.id
  );
