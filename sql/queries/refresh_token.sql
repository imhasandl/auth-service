-- name: RefreshToken :one
INSERT INTO refresh_tokens (token, user_id, expiry_time) 
VALUES (
   $1, 
   $2, 
   $3
)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: DeleteTokenByUserID :exec
DELETE FROM refresh_tokens
WHERE user_id = $1;

-- name: DeleteRefreshTokenByToken :exec
DELETE FROM refresh_tokens
WHERE token = $1;

