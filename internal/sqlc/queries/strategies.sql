-- name: GetStrategyByID :one
SELECT * FROM strategies WHERE id = $1;

-- name: CreateStrategy :one
INSERT INTO strategies (name, strategy_type, configs)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateStrategy :one
UPDATE strategies SET
    name = $2,
    strategy_type = $3,
    configs = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteStrategy :exec
DELETE FROM strategies WHERE id = $1;

-- name: ListStrategies :many
SELECT * FROM strategies ORDER BY created_at DESC;
