-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, password, username, is_premium)
VALUES (
   $1,
   NOW(),
   NOW(),
   $2,
   $3,
   $4,
   $5
)
RETURNING *;