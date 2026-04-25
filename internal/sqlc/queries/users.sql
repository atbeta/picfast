-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (group_id, email, password, name, role, capacity_bytes, settings, status, email_verified, registered_ip)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdateUser :one
UPDATE users SET
    name = $2,
    password = $3,
    group_id = $4,
    capacity_bytes = $5,
    status = $6,
    settings = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAllUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: IncrementUserImageNum :exec
UPDATE users SET image_num = image_num + 1, updated_at = NOW() WHERE id = $1;

-- name: DecrementUserImageNum :exec
UPDATE users SET image_num = image_num - 1, updated_at = NOW() WHERE id = $1 AND image_num > 0;

-- name: GetUserUsedCapacity :one
SELECT COALESCE(SUM(size_bytes), 0)::bigint FROM images WHERE user_id = $1;

-- name: CreateAdminUser :one
INSERT INTO users (group_id, email, password, name, role, capacity_bytes, settings, status, email_verified, registered_ip)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, '')
RETURNING *;
