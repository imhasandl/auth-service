-- name: RefreshToken :one
INSERT INTO refresh_tokens (token, user_id, expiry_time) 
VALUES (
   $1, 
   $2, 
   $3
)
RETURNING *;

-- name: DeleteToken :exec
DELETE FROM refresh_tokens
WHERE user_id = $1;