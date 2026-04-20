-- name: GetUserFromRefreshToken :one
SELECT user_id from refresh_tokens
WHERE token = $1 AND expires_at > NOW() AND revoked_at IS NULL;

-- name: AddRefreshToken :exec
INSERT into refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES ($1, NOW(), NOW(), $2, (NOW() + INTERVAL '60 DAYS'), NULL);

-- name: GetRefrehToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET expires_at = NOW(), updated_at = NOW()
WHERE token = $1;