-- name: GetGroupByID :one
SELECT * FROM groups WHERE id = $1;

-- name: GetDefaultGroup :one
SELECT * FROM groups WHERE is_default = TRUE LIMIT 1;

-- name: GetGuestGroup :one
SELECT * FROM groups WHERE is_guest = TRUE LIMIT 1;

-- name: CreateGroup :one
INSERT INTO groups (name, is_default, is_guest, configs)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateGroup :one
UPDATE groups SET
    name = COALESCE($2, name),
    is_default = COALESCE($3, is_default),
    is_guest = COALESCE($4, is_guest),
    configs = COALESCE($5, configs),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGroup :exec
DELETE FROM groups WHERE id = $1;

-- name: ListGroups :many
SELECT * FROM groups ORDER BY created_at DESC;

-- name: UnsetOtherDefault :exec
UPDATE groups SET is_default = FALSE, updated_at = NOW() WHERE id != $1 AND is_default = TRUE;

-- name: UnsetOtherGuest :exec
UPDATE groups SET is_guest = FALSE, updated_at = NOW() WHERE id != $1 AND is_guest = TRUE;

-- name: GetGroupStrategies :many
SELECT s.* FROM strategies s
JOIN group_strategies gs ON s.id = gs.strategy_id
WHERE gs.group_id = $1;

-- name: AddGroupStrategy :exec
INSERT INTO group_strategies (group_id, strategy_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveGroupStrategy :exec
DELETE FROM group_strategies WHERE group_id = $1 AND strategy_id = $2;

-- name: ReplaceGroupStrategies :exec
DELETE FROM group_strategies WHERE group_id = $1;
