-- name: CreateUserIdentity :one
INSERT INTO user_identities (user_id, provider, provider_subject, email)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserIdentityByProviderSubject :one
SELECT * FROM user_identities
WHERE provider = $1 AND provider_subject = $2;

-- name: ListUserIdentitiesByUser :many
SELECT * FROM user_identities
WHERE user_id = $1
ORDER BY linked_at DESC;

-- name: DeleteUserIdentity :exec
DELETE FROM user_identities
WHERE user_id = $1 AND provider = $2;
