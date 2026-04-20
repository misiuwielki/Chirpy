-- name: GetUserFromRefreshToken :one
SELECT * from users
WHERE ID = (
    SELECT user_id from refresh_tokens
    WHERE token = $1
);