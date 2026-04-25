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
    name = COALESCE($2, name),
    password = COALESCE($3, password),
    group_id = COALESCE($4, group_id),
    capacity_bytes = COALESCE($5, capacity_bytes),
    status = COALESCE($6, status),
    settings = COALESCE($7, settings),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
WHERE ($1::text IS NULL OR email ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%')
AND ($2::smallint IS NULL OR status = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE ($1::text IS NULL OR email ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%')
AND ($2::smallint IS NULL OR status = $2);

-- name: IncrementUserImageNum :exec
UPDATE users SET image_num = image_num + 1, updated_at = NOW() WHERE id = $1;

-- name: DecrementUserImageNum :exec
UPDATE users SET image_num = image_num - 1, updated_at = NOW() WHERE id = $1 AND image_num > 0;

-- name: GetUserUsedCapacity :one
SELECT COALESCE(SUM(size_bytes), 0) FROM images WHERE user_id = $1;

-- name: CreateAdminUser :one
INSERT INTO users (group_id, email, password, name, role, capacity_bytes, settings, status, email_verified, registered_ip)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, '')
RETURNING *;
