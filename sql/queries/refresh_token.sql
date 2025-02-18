-- name: RefreshToken :one
INSERT INTO refresh_tokens (token, user_id, expiry_time) 
VALUES (
   $1, 
   $2, 
   $3
)
RETURNING *;