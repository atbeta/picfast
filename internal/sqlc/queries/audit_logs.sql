-- name: CreateAuditLog :exec
INSERT INTO audit_logs (
    actor_user_id,
    action,
    resource_type,
    resource_id,
    resource_name,
    details,
    ip,
    user_agent
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListAuditLogs :many
SELECT
    al.*,
    u.email AS actor_email
FROM audit_logs al
LEFT JOIN users u ON u.id = al.actor_user_id
WHERE (sqlc.narg('action')::text IS NULL OR al.action = sqlc.narg('action'))
  AND (sqlc.narg('resource_type')::text IS NULL OR al.resource_type = sqlc.narg('resource_type'))
ORDER BY al.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAuditLogs :one
SELECT COUNT(*)
FROM audit_logs al
WHERE (sqlc.narg('action')::text IS NULL OR al.action = sqlc.narg('action'))
  AND (sqlc.narg('resource_type')::text IS NULL OR al.resource_type = sqlc.narg('resource_type'));
